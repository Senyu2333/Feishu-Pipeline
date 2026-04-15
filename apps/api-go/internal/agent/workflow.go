package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	agenttype "feishu-pipeline/apps/api-go/internal/type/agent"
	"feishu-pipeline/apps/api-go/internal/utils"

	"github.com/cloudwego/eino/compose"
)

type PublishInput struct {
	Session      *agenttype.SessionAggregate
	RoleMappings []model.RoleMapping
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
	Title               string
	Summary             string
	ReferencedKnowledge []string
	Tasks               []model.Task
}

func NewEngine() (*Engine, error) {
	graph := compose.NewGraph[*pipelineState, PublishOutput]()

	_ = graph.AddLambdaNode("normalize", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return normalizeState(state), nil
	}))
	_ = graph.AddLambdaNode("rag", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return enrichKnowledge(state), nil
	}))
	_ = graph.AddLambdaNode("split", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return splitTasks(state), nil
	}))
	_ = graph.AddLambdaNode("assign", compose.InvokableLambda(func(ctx context.Context, state *pipelineState) (*pipelineState, error) {
		return assignOwners(state), nil
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
	_ = graph.AddEdge("assign", "docs")
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

func normalizeState(state *pipelineState) *pipelineState {
	var productMessages []string
	for _, message := range state.Input.Session.Messages {
		if message.Role == model.MessageUser {
			productMessages = append(productMessages, strings.TrimSpace(message.Content))
		}
	}
	title := strings.TrimSpace(state.Input.Session.Session.Title)
	if title == "" {
		title = "未命名需求"
	}

	summary := strings.Join(productMessages, "\n")
	if summary == "" {
		summary = state.Input.Session.Session.Summary
	}

	state.Title = title
	state.Summary = utils.Summarize(summary, 240)
	return state
}

func enrichKnowledge(state *pipelineState) *pipelineState {
	content := strings.ToLower(state.Title + " " + state.Summary)
	seen := map[string]struct{}{}
	for _, source := range state.Input.Knowledge {
		title := strings.ToLower(source.Title)
		body := strings.ToLower(source.Content)
		if strings.Contains(content, title) || strings.Contains(body, "规范") || strings.Contains(body, "流程") {
			if _, ok := seen[source.Title]; !ok {
				state.ReferencedKnowledge = append(state.ReferencedKnowledge, fmt.Sprintf("%s：%s", source.Title, utils.Summarize(source.Content, 120)))
				seen[source.Title] = struct{}{}
			}
		}
	}
	if len(state.ReferencedKnowledge) == 0 {
		for idx, source := range state.Input.Knowledge {
			if idx >= 2 {
				break
			}
			state.ReferencedKnowledge = append(state.ReferencedKnowledge, fmt.Sprintf("%s：%s", source.Title, utils.Summarize(source.Content, 120)))
		}
	}
	return state
}

func splitTasks(state *pipelineState) *pipelineState {
	now := time.Now().UTC()
	lowerSummary := strings.ToLower(state.Summary)

	state.Tasks = []model.Task{
		{
			ID:          utils.NewID("frontend"),
			SessionID:   state.Input.Session.Session.ID,
			Title:       "前端交互与页面实现",
			Description: "实现会话列表、聊天主界面、需求详情侧栏，并联调任务状态与发布结果展示。",
			Type:        model.TaskFrontend,
			Status:      model.TaskTodo,
			AcceptanceCriteria: []string{
				"支持登录后查看需求会话列表",
				"支持发送消息和查看 AI 回复",
				"支持查看需求详情和任务状态",
			},
			Risks:     []string{"页面状态同步复杂", "接口加载与空态处理需要稳定"},
			BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now},
		},
		{
			ID:          utils.NewID("backend"),
			SessionID:   state.Input.Session.Session.ID,
			Title:       "后端接口与交付流程实现",
			Description: "实现会话、任务、发布、知识同步和飞书分发接口，保证需求发布后可进入后台工作流。",
			Type:        model.TaskBackend,
			Status:      model.TaskTodo,
			AcceptanceCriteria: []string{
				"提供会话与任务 REST API",
				"发布需求后异步生成任务并持久化",
				"支持任务状态更新与飞书同步",
			},
			Risks:     []string{"飞书接口配置缺失", "发布流程需要保证幂等"},
			BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now},
		},
	}

	if strings.Contains(lowerSummary, "联调") || strings.Contains(lowerSummary, "接口") || strings.Contains(lowerSummary, "验收") {
		state.Tasks = append(state.Tasks, model.Task{
			ID:          utils.NewID("shared"),
			SessionID:   state.Input.Session.Session.ID,
			Title:       "公共联调与验收准备",
			Description: "整理验收标准、联调依赖和提测说明，确保前后端对齐交付口径。",
			Type:        model.TaskShared,
			Status:      model.TaskTodo,
			AcceptanceCriteria: []string{
				"明确接口字段与状态流转",
				"明确提测前置条件",
			},
			Risks:     []string{"需求变更导致验收口径漂移"},
			BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now},
		})
	}

	return state
}

func assignOwners(state *pipelineState) *pipelineState {
	for idx := range state.Tasks {
		switch state.Tasks[idx].Type {
		case model.TaskFrontend:
			state.Tasks[idx].AssigneeRole = model.RoleFrontend
			state.Tasks[idx].AssigneeName = resolveAssigneeName(model.RoleFrontend, state.Input.RoleMappings)
		case model.TaskBackend:
			state.Tasks[idx].AssigneeRole = model.RoleBackend
			state.Tasks[idx].AssigneeName = resolveAssigneeName(model.RoleBackend, state.Input.RoleMappings)
		default:
			state.Tasks[idx].AssigneeRole = model.RoleProduct
			state.Tasks[idx].AssigneeName = resolveAssigneeName(model.RoleProduct, state.Input.RoleMappings)
		}
	}
	return state
}

func writeTaskDocs(state *pipelineState) *pipelineState {
	for idx := range state.Tasks {
		task := &state.Tasks[idx]
		task.Description = fmt.Sprintf("# %s\n\n## 需求摘要\n%s\n\n## 任务说明\n%s\n\n## 参考知识\n- %s\n",
			task.Title,
			state.Summary,
			task.Description,
			strings.Join(state.ReferencedKnowledge, "\n- "),
		)
	}
	return state
}

func composeOutput(state *pipelineState) PublishOutput {
	requirement := model.Requirement{
		ID:                  utils.NewID("req"),
		SessionID:           state.Input.Session.Session.ID,
		Title:               state.Title,
		Summary:             state.Summary,
		Status:              model.SessionInDelivery,
		DeliverySummary:     fmt.Sprintf("已拆解 %d 个任务，并为前后端负责人生成交付说明。", len(state.Tasks)),
		ReferencedKnowledge: state.ReferencedKnowledge,
		PublishedAt:         time.Now().UTC(),
	}

	return PublishOutput{
		Requirement: requirement,
		Tasks:       state.Tasks,
	}
}

func resolveAssigneeName(role model.Role, mappings []model.RoleMapping) string {
	for _, mapping := range mappings {
		if mapping.Role == role {
			switch role {
			case model.RoleFrontend:
				return "前端负责人小红"
			case model.RoleBackend:
				return "后端负责人小李"
			case model.RoleProduct:
				return "产品经理小明"
			}
		}
	}

	switch role {
	case model.RoleFrontend:
		return "前端负责人小红"
	case model.RoleBackend:
		return "后端负责人小李"
	default:
		return "产品经理小明"
	}
}
