package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/external/ai"
	"feishu-pipeline/apps/api-go/internal/model"
	agenttype "feishu-pipeline/apps/api-go/internal/type/agent"
	"feishu-pipeline/apps/api-go/internal/utils"

	"github.com/cloudwego/eino/compose"
)

type PublishInput struct {
	Session      *agenttype.SessionAggregate
	RoleMappings []model.RoleMapping
	RoleOwners   []model.RoleOwner
	Knowledge    []model.KnowledgeSource
}

type PublishOutput struct {
	Requirement model.Requirement
	Tasks       []model.Task
}

type Engine struct {
	runnable compose.Runnable[*pipelineState, PublishOutput]
}

type pipelineState struct {
	Input               PublishInput
	Normalized          normalizedRequirement
	ReferencedKnowledge []string
	Tasks               []model.Task
}

func NewEngine(client ai.Client) (*Engine, error) {
	graph := compose.NewGraph[*pipelineState, PublishOutput]()

	_ = graph.AddLambdaNode("normalize", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return normalizeState(ctx, client, state)
	}))
	_ = graph.AddLambdaNode("rag", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return enrichKnowledge(state), nil
	}))
	_ = graph.AddLambdaNode("split", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return splitTasks(ctx, client, state)
	}))
	_ = graph.AddLambdaNode("assign", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return assignOwners(state), nil
	}))
	_ = graph.AddLambdaNode("schedule", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return scheduleTasks(ctx, client, state)
	}))
	_ = graph.AddLambdaNode("docs", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return writeTaskDocs(state), nil
	}))
	_ = graph.AddLambdaNode("summary", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (PublishOutput, error) {
		return composeOutput(state), nil
	}))

	_ = graph.AddEdge(compose.START, "normalize")
	_ = graph.AddEdge("normalize", "rag")
	_ = graph.AddEdge("rag", "split")
	_ = graph.AddEdge("split", "assign")
	_ = graph.AddEdge("assign", "schedule")
	_ = graph.AddEdge("schedule", "docs")
	_ = graph.AddEdge("docs", "summary")
	_ = graph.AddEdge("summary", compose.END)

	runnable, err := graph.Compile(context.Background())
	if err != nil {
		return nil, fmt.Errorf("compile eino graph: %w", err)
	}

	return &Engine{runnable: runnable}, nil
}

func (e *Engine) Execute(ctx context.Context, input PublishInput) (PublishOutput, error) {
	return e.runnable.Invoke(ctx, &pipelineState{Input: input})
}

func normalizeState(ctx context.Context, client ai.Client, state *pipelineState) (*pipelineState, error) {
	if client != nil {
		plan, err := normalizeWithLLM(ctx, client, state)
		if err == nil {
			state.Normalized = plan
			return state, nil
		}
	}

	state.Normalized = deterministicRequirement(state)
	return state, nil
}

func enrichKnowledge(state *pipelineState) *pipelineState {
	content := strings.ToLower(state.Normalized.Title + " " + state.Normalized.Summary)
	referencedTitles := make(map[string]struct{}, len(state.Normalized.ReferencedKnowledgeTitles))
	for _, title := range state.Normalized.ReferencedKnowledgeTitles {
		referencedTitles[strings.ToLower(strings.TrimSpace(title))] = struct{}{}
	}

	seen := map[string]struct{}{}
	for _, source := range state.Input.Knowledge {
		title := strings.ToLower(source.Title)
		body := strings.ToLower(source.Content)
		_, referenced := referencedTitles[title]
		if referenced || strings.Contains(content, title) || strings.Contains(body, "规范") || strings.Contains(body, "流程") {
			if _, ok := seen[source.Title]; !ok {
				state.ReferencedKnowledge = append(state.ReferencedKnowledge, fmt.Sprintf("%s：%s", source.Title, utils.Summarize(source.Content, 120)))
				seen[source.Title] = struct{}{}
			}
		}
	}
	if len(state.ReferencedKnowledge) == 0 {
		for idx, source := range state.Input.Knowledge {
			if idx >= 3 {
				break
			}
			state.ReferencedKnowledge = append(state.ReferencedKnowledge, fmt.Sprintf("%s：%s", source.Title, utils.Summarize(source.Content, 120)))
		}
	}
	return state
}

func splitTasks(ctx context.Context, client ai.Client, state *pipelineState) (*pipelineState, error) {
	if client != nil {
		tasks, err := splitTasksWithLLM(ctx, client, state)
		if err == nil && len(tasks) > 0 {
			state.Tasks = tasks
			return state, nil
		}
	}

	state.Tasks = deterministicTasks(state)
	return state, nil
}

func assignOwners(state *pipelineState) *pipelineState {
	for idx := range state.Tasks {
		role := state.Tasks[idx].AssigneeRole
		owner, ok := resolveRoleOwner(role, state.Input.RoleOwners)
		if ok {
			state.Tasks[idx].AssigneeName = owner.OwnerName
			state.Tasks[idx].AssigneeID = owner.FeishuID
			state.Tasks[idx].AssigneeIDType = owner.FeishuIDType
			continue
		}
		state.Tasks[idx].AssigneeName = defaultAssigneeName(role)
	}
	return state
}

func scheduleTasks(ctx context.Context, client ai.Client, state *pipelineState) (*pipelineState, error) {
	assignScheduleWindows(state.Tasks)

	if client != nil {
		if err := fillNotifyContentWithLLM(ctx, client, state); err == nil {
			return state, nil
		}
	}

	for idx := range state.Tasks {
		state.Tasks[idx].NotifyContent = defaultNotifyContent(state.Normalized.Title, state.Tasks[idx])
	}
	return state, nil
}

func writeTaskDocs(state *pipelineState) *pipelineState {
	for idx := range state.Tasks {
		task := &state.Tasks[idx]
		task.Description = fmt.Sprintf(
			"# %s\n\n## 需求摘要\n%s\n\n## 任务说明\n%s\n\n## 验收标准\n- %s\n\n## 风险提示\n- %s\n\n## 排期信息\n- 优先级：%s\n- 预计工期：%d 天\n- 计划开始：%s\n- 计划结束：%s\n\n## 通知文案\n%s\n\n## 参考知识\n- %s\n",
			task.Title,
			state.Normalized.Summary,
			task.Description,
			strings.Join(task.AcceptanceCriteria, "\n- "),
			strings.Join(task.Risks, "\n- "),
			task.Priority,
			task.EstimateDays,
			formatPromptDate(task.PlannedStartAt),
			formatPromptDate(task.PlannedEndAt),
			task.NotifyContent,
			strings.Join(state.ReferencedKnowledge, "\n- "),
		)
	}
	return state
}

func composeOutput(state *pipelineState) PublishOutput {
	deliverySummary := state.Normalized.DeliverySummary
	if deliverySummary == "" {
		deliverySummary = fmt.Sprintf("已拆解 %d 个任务，并完成通知文案和排期。", len(state.Tasks))
	}

	requirement := model.Requirement{
		ID:                  utils.NewID("req"),
		SessionID:           state.Input.Session.Session.ID,
		Title:               state.Normalized.Title,
		Summary:             state.Normalized.Summary,
		Status:              model.SessionInDelivery,
		DeliverySummary:     deliverySummary,
		ReferencedKnowledge: state.ReferencedKnowledge,
		PublishedAt:         time.Now().UTC(),
	}

	return PublishOutput{
		Requirement: requirement,
		Tasks:       state.Tasks,
	}
}

func normalizeWithLLM(ctx context.Context, client ai.Client, state *pipelineState) (normalizedRequirement, error) {
	raw, _, err := client.Generate(
		ctx,
		buildNormalizeSystemPrompt(),
		buildNormalizeUserPrompt(sessionTitle(state), conversationFromMessages(state.Input.Session.Messages), state.Input.Knowledge),
	)
	if err != nil {
		return normalizedRequirement{}, err
	}

	plan, err := parseStructuredJSON[normalizedRequirement](raw)
	if err != nil {
		return normalizedRequirement{}, err
	}
	return validateRequirementPlan(plan)
}

func splitTasksWithLLM(ctx context.Context, client ai.Client, state *pipelineState) ([]model.Task, error) {
	raw, _, err := client.Generate(
		ctx,
		buildSplitSystemPrompt(),
		buildSplitUserPrompt(state.Normalized, state.ReferencedKnowledge, state.Input.RoleMappings),
	)
	if err != nil {
		return nil, err
	}

	output, err := parseStructuredJSON[taskSplitOutput](raw)
	if err != nil {
		return nil, err
	}

	plans, err := validateTaskSplit(output)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	tasks := make([]model.Task, 0, len(plans))
	for _, item := range plans {
		taskType := normalizeTaskType(item.Type)
		tasks = append(tasks, model.Task{
			ID:                 utils.NewID(string(taskType)),
			SessionID:          state.Input.Session.Session.ID,
			Title:              item.Title,
			Description:        item.Description,
			Type:               taskType,
			Status:             model.TaskTodo,
			AssigneeRole:       normalizeRole(item.AssigneeRole, taskType),
			Priority:           normalizeTaskPriority(item.Priority),
			EstimateDays:       item.EstimateDays,
			AcceptanceCriteria: item.AcceptanceCriteria,
			Risks:              item.Risks,
			BaseModel:          model.BaseModel{CreatedAt: now, UpdatedAt: now},
		})
	}
	return tasks, nil
}

func fillNotifyContentWithLLM(ctx context.Context, client ai.Client, state *pipelineState) error {
	raw, _, err := client.Generate(
		ctx,
		buildNotifySystemPrompt(),
		buildNotifyUserPrompt(state.Normalized.Title, state.Normalized.Summary, state.Tasks),
	)
	if err != nil {
		return err
	}

	output, err := parseStructuredJSON[notificationOutput](raw)
	if err != nil {
		return err
	}

	byTitle := make(map[string]string, len(output.Items))
	for _, item := range output.Items {
		title := strings.TrimSpace(item.TaskTitle)
		content := strings.TrimSpace(item.NotifyContent)
		if title != "" && content != "" {
			byTitle[title] = content
		}
	}

	for idx := range state.Tasks {
		if content := byTitle[state.Tasks[idx].Title]; content != "" {
			state.Tasks[idx].NotifyContent = content
		} else {
			state.Tasks[idx].NotifyContent = defaultNotifyContent(state.Normalized.Title, state.Tasks[idx])
		}
	}
	return nil
}

func deterministicRequirement(state *pipelineState) normalizedRequirement {
	title := sessionTitle(state)
	summary := utils.Summarize(conversationFromMessages(state.Input.Session.Messages), 300)
	if summary == "" {
		summary = utils.Summarize(state.Input.Session.Session.Summary, 300)
	}

	return normalizedRequirement{
		Title:           title,
		Summary:         summary,
		DeliverySummary: fmt.Sprintf("已基于需求会话生成交付摘要，待拆解 %d 类任务。", 2),
	}
}

func deterministicTasks(state *pipelineState) []model.Task {
	now := time.Now().UTC()
	lowerSummary := strings.ToLower(state.Normalized.Summary)

	tasks := []model.Task{
		{
			ID:                 utils.NewID("frontend"),
			SessionID:          state.Input.Session.Session.ID,
			Title:              "前端交互与页面实现",
			Description:        "实现会话列表、聊天主界面、需求详情侧栏，并联调任务状态与发布结果展示。",
			Type:               model.TaskFrontend,
			Status:             model.TaskTodo,
			AssigneeRole:       model.RoleFrontend,
			Priority:           model.TaskPriorityHigh,
			EstimateDays:       3,
			AcceptanceCriteria: []string{"支持登录后查看需求会话列表", "支持发送消息和查看 AI 回复", "支持查看需求详情和任务状态"},
			Risks:              []string{"页面状态同步复杂", "接口加载与空态处理需要稳定"},
			BaseModel:          model.BaseModel{CreatedAt: now, UpdatedAt: now},
		},
		{
			ID:                 utils.NewID("backend"),
			SessionID:          state.Input.Session.Session.ID,
			Title:              "后端接口与交付流程实现",
			Description:        "实现会话、任务、发布、知识同步和飞书分发接口，保证需求发布后可进入后台工作流。",
			Type:               model.TaskBackend,
			Status:             model.TaskTodo,
			AssigneeRole:       model.RoleBackend,
			Priority:           model.TaskPriorityHigh,
			EstimateDays:       4,
			AcceptanceCriteria: []string{"提供会话与任务 REST API", "发布需求后异步生成任务并持久化", "支持任务状态更新与飞书同步"},
			Risks:              []string{"飞书接口配置缺失", "发布流程需要保证幂等"},
			BaseModel:          model.BaseModel{CreatedAt: now, UpdatedAt: now},
		},
	}

	if strings.Contains(lowerSummary, "联调") || strings.Contains(lowerSummary, "接口") || strings.Contains(lowerSummary, "验收") {
		tasks = append(tasks, model.Task{
			ID:                 utils.NewID("shared"),
			SessionID:          state.Input.Session.Session.ID,
			Title:              "公共联调与验收准备",
			Description:        "整理验收标准、联调依赖和提测说明，确保前后端对齐交付口径。",
			Type:               model.TaskShared,
			Status:             model.TaskTodo,
			AssigneeRole:       model.RoleProduct,
			Priority:           model.TaskPriorityMedium,
			EstimateDays:       2,
			AcceptanceCriteria: []string{"明确接口字段与状态流转", "明确提测前置条件"},
			Risks:              []string{"需求变更导致验收口径漂移"},
			BaseModel:          model.BaseModel{CreatedAt: now, UpdatedAt: now},
		})
	}

	return tasks
}

func assignScheduleWindows(tasks []model.Task) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	cursors := map[model.Role]time.Time{}

	for idx := range tasks {
		role := tasks[idx].AssigneeRole
		start := today
		if cursor, ok := cursors[role]; ok {
			start = cursor
		}

		end := start.AddDate(0, 0, max(tasks[idx].EstimateDays-1, 0))
		startCopy := start
		endCopy := end
		tasks[idx].PlannedStartAt = &startCopy
		tasks[idx].PlannedEndAt = &endCopy
		cursors[role] = end.AddDate(0, 0, 1)
	}
}

func resolveRoleOwner(role model.Role, owners []model.RoleOwner) (model.RoleOwner, bool) {
	for _, owner := range owners {
		if owner.Role == role && owner.Enabled {
			return owner, true
		}
	}
	return model.RoleOwner{}, false
}

func defaultAssigneeName(role model.Role) string {
	switch role {
	case model.RoleFrontend:
		return "前端负责人"
	case model.RoleBackend:
		return "后端负责人"
	case model.RoleAdmin:
		return "管理员"
	default:
		return "产品负责人"
	}
}

func defaultNotifyContent(requirementTitle string, task model.Task) string {
	return fmt.Sprintf("你有新的需求任务《%s》：%s，优先级 %s，预计 %d 天，请按排期推进并及时同步状态。", requirementTitle, task.Title, task.Priority, task.EstimateDays)
}

func conversationFromMessages(messages []model.Message) string {
	lines := make([]string, 0, len(messages))
	for _, message := range messages {
		if message.Role == model.MessageUser {
			lines = append(lines, strings.TrimSpace(message.Content))
		}
	}
	return strings.Join(lines, "\n")
}

func sessionTitle(state *pipelineState) string {
	title := strings.TrimSpace(state.Input.Session.Session.Title)
	if title == "" {
		return "未命名需求"
	}
	return title
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
