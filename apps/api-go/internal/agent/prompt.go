package agent

import (
	"fmt"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
)

const (
	jsonOnlyInstruction = "你必须只输出合法 JSON，不要输出 Markdown 代码块、解释、前后缀说明。"
)

// BuildChatSystemPrompt 需求助手对话系统 prompt（供 session_service 调用）
func BuildChatSystemPrompt() string {
	return strings.Join([]string{
		"你是需求交付引擎的 AI 助手，专注于帮助产品经理和研发团队整理、分析、完善产品需求。",
		"你可以帮助用户澄清需求目标、识别边界、梳理核心流程、明确验收标准、提示风险点。",
		"当用户描述需求时，请积极追问关键细节，帮助他们完善需求内容。",
		"回答应简洁专业，使用中文，避免过度发散。",
		"如果用户说「发布需求」或类似意图，告知他们系统会自动触发交付工作流。",
	}, "\n")
}

// BuildChatUserPrompt 构建带历史上下文的用户 prompt（供 session_service 调用）
func BuildChatUserPrompt(history []model.Message, userMsg string) string {
	if len(history) == 0 {
		return userMsg
	}

	// 最多带最近 10 条历史，避免 token 过长
	start := 0
	if len(history) > 10 {
		start = len(history) - 10
	}

	lines := make([]string, 0, len(history)-start+2)
	lines = append(lines, "以下是之前的对话记录：")
	for _, msg := range history[start:] {
		role := "用户"
		if msg.Role == model.MessageAssistant {
			role = "助手"
		}
		lines = append(lines, fmt.Sprintf("[%s]: %s", role, strings.TrimSpace(msg.Content)))
	}
	lines = append(lines, "\n当前用户消息："+userMsg)
	return strings.Join(lines, "\n")
}

func buildNormalizeSystemPrompt() string {
	return strings.Join([]string{
		"你是需求交付引擎中的需求分析助手。",
		"你的任务是把产品需求会话整理为正式需求摘要，为后续任务拆解提供稳定输入。",
		jsonOnlyInstruction,
		`输出 JSON 结构：{"title":"","summary":"","delivery_summary":"","referenced_knowledge_titles":[]}`,
		"title 要简洁准确，summary 要覆盖目标、流程、范围、验收重点，delivery_summary 要概述交付结果。",
	}, "\n")
}

func buildNormalizeUserPrompt(title string, conversation string, knowledge []model.KnowledgeSource) string {
	return fmt.Sprintf("会话标题：%s\n\n会话内容：\n%s\n\n可参考知识标题：\n%s",
		title,
		conversation,
		joinKnowledgeTitles(knowledge),
	)
}

func buildSplitSystemPrompt() string {
	return strings.Join([]string{
		"你是需求拆解助手，需要把需求拆成可执行任务。",
		"任务类型只能是 frontend、backend、shared，负责人角色只能是 frontend、backend、product。",
		"每个任务都必须包含 title、description、acceptance_criteria、risks、priority、estimate_days、assignee_role。",
		"priority 只能是 high、medium、low；estimate_days 是 1 到 30 的整数。",
		jsonOnlyInstruction,
		`输出 JSON 结构：{"tasks":[{"type":"","title":"","description":"","acceptance_criteria":[],"risks":[],"priority":"","estimate_days":1,"assignee_role":""}]}`,
	}, "\n")
}

func buildSplitUserPrompt(requirement normalizedRequirement, knowledge []string, mappings []model.RoleMapping) string {
	return fmt.Sprintf("正式需求标题：%s\n\n正式需求摘要：\n%s\n\n交付总结：\n%s\n\n知识片段：\n%s\n\n角色映射：\n%s",
		requirement.Title,
		requirement.Summary,
		requirement.DeliverySummary,
		strings.Join(knowledge, "\n"),
		joinRoleMappings(mappings),
	)
}

func buildNotifySystemPrompt() string {
	return strings.Join([]string{
		"你是研发任务通知助手，需要为每个任务生成适合飞书发送的通知文案。",
		"文案要清晰说明任务目标、截止预期和协作提醒，每条文案控制在 120 字以内。",
		jsonOnlyInstruction,
		`输出 JSON 结构：{"items":[{"task_title":"","notify_content":""}]}`,
	}, "\n")
}

func buildNotifyUserPrompt(requirementTitle string, requirementSummary string, tasks []model.Task) string {
	parts := make([]string, 0, len(tasks))
	for _, task := range tasks {
		parts = append(parts, fmt.Sprintf(
			"任务标题：%s\n任务类型：%s\n负责人角色：%s\n优先级：%s\n计划开始：%s\n计划结束：%s\n任务说明：%s",
			task.Title,
			task.Type,
			task.AssigneeRole,
			task.Priority,
			formatPromptDate(task.PlannedStartAt),
			formatPromptDate(task.PlannedEndAt),
			task.Description,
		))
	}

	return fmt.Sprintf("需求标题：%s\n\n需求摘要：%s\n\n任务列表：\n%s",
		requirementTitle,
		requirementSummary,
		strings.Join(parts, "\n\n"),
	)
}

func joinKnowledgeTitles(knowledge []model.KnowledgeSource) string {
	if len(knowledge) == 0 {
		return "暂无"
	}

	titles := make([]string, 0, len(knowledge))
	for _, item := range knowledge {
		titles = append(titles, "- "+item.Title)
	}
	return strings.Join(titles, "\n")
}

func joinRoleMappings(mappings []model.RoleMapping) string {
	if len(mappings) == 0 {
		return "暂无"
	}

	lines := make([]string, 0, len(mappings))
	for _, item := range mappings {
		lines = append(lines, fmt.Sprintf("- role=%s keyword=%s departments=%s", item.Role, item.Keyword, strings.Join(item.Departments, ",")))
	}
	return strings.Join(lines, "\n")
}

func formatPromptDate(value *time.Time) string {
	if value == nil {
		return "待排期"
	}
	return value.Format("2006-01-02")
}
