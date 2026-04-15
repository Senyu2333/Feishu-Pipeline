package agent

import (
	"context"
	"fmt"
	"strings"

	"feishu-pipeline/apps/api-go/internal/model"

	"github.com/cloudwego/eino/schema"
)

func parseStructuredJSON[T any](raw string) (T, error) {
	parser := schema.NewMessageJSONParser[T](&schema.MessageJSONParseConfig{
		ParseFrom: schema.MessageParseFromContent,
	})

	message := schema.AssistantMessage(cleanStructuredJSON(raw), nil)
	return parser.Parse(context.Background(), message)
}

func cleanStructuredJSON(raw string) string {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	for _, pair := range [][2]byte{{'{', '}'}, {'[', ']'}} {
		start := strings.IndexByte(cleaned, pair[0])
		end := strings.LastIndexByte(cleaned, pair[1])
		if start >= 0 && end > start {
			return strings.TrimSpace(cleaned[start : end+1])
		}
	}

	return cleaned
}

func validateRequirementPlan(plan normalizedRequirement) (normalizedRequirement, error) {
	plan.Title = strings.TrimSpace(plan.Title)
	plan.Summary = strings.TrimSpace(plan.Summary)
	plan.DeliverySummary = strings.TrimSpace(plan.DeliverySummary)
	if plan.Title == "" || plan.Summary == "" {
		return normalizedRequirement{}, fmt.Errorf("invalid normalized requirement")
	}
	if plan.DeliverySummary == "" {
		plan.DeliverySummary = "已完成需求解析与任务拆解，待进入交付。"
	}
	return plan, nil
}

func validateTaskSplit(output taskSplitOutput) ([]taskPlan, error) {
	if len(output.Tasks) == 0 {
		return nil, fmt.Errorf("empty task split output")
	}

	result := make([]taskPlan, 0, len(output.Tasks))
	for _, item := range output.Tasks {
		item.Title = strings.TrimSpace(item.Title)
		item.Description = strings.TrimSpace(item.Description)
		if item.Title == "" || item.Description == "" {
			continue
		}
		if item.EstimateDays <= 0 {
			item.EstimateDays = 1
		}
		if item.EstimateDays > 30 {
			item.EstimateDays = 30
		}
		if len(item.AcceptanceCriteria) == 0 {
			item.AcceptanceCriteria = []string{"实现需求描述中的核心流程并完成自测"}
		}
		if len(item.Risks) == 0 {
			item.Risks = []string{"需求细节可能在联调阶段调整"}
		}
		result = append(result, item)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid task split output")
	}
	return result, nil
}

func normalizeTaskType(value string) model.TaskType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(model.TaskFrontend):
		return model.TaskFrontend
	case string(model.TaskShared):
		return model.TaskShared
	default:
		return model.TaskBackend
	}
}

func normalizeTaskPriority(value string) model.TaskPriority {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(model.TaskPriorityHigh):
		return model.TaskPriorityHigh
	case string(model.TaskPriorityLow):
		return model.TaskPriorityLow
	default:
		return model.TaskPriorityMedium
	}
}

func normalizeRole(value string, taskType model.TaskType) model.Role {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(model.RoleFrontend):
		return model.RoleFrontend
	case string(model.RoleProduct):
		return model.RoleProduct
	case string(model.RoleAdmin):
		return model.RoleAdmin
	case string(model.RoleBackend):
		return model.RoleBackend
	}

	switch taskType {
	case model.TaskFrontend:
		return model.RoleFrontend
	case model.TaskShared:
		return model.RoleProduct
	default:
		return model.RoleBackend
	}
}
