package pipeline

import (
	"context"
	"fmt"
	"strings"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/utils"
)

type RequirementAnalysisHandler struct{}

func (RequirementAnalysisHandler) Execute(_ context.Context, ctx StageContext) (StageExecutionResult, error) {
	summary := utils.Summarize(ctx.Run.RequirementText, 240)
	acceptanceCriteria := []string{"需求目标已被结构化描述", "影响范围已被识别", "方案设计可直接消费该结果"}
	payload := baseStagePayload(ctx)
	payload[SchemaFieldSummary] = summary
	payload[SchemaFieldGoals] = []string{ctx.Run.Title}
	payload[SchemaFieldNonGoals] = []string{"不在当前阶段直接执行代码写入或远程交付"}
	payload[SchemaFieldAcceptanceCriteria] = acceptanceCriteria
	payload["risks"] = []string{"需求仍可能存在业务边界不清，需要在方案审批时确认"}
	payload[SchemaFieldQuestions] = []string{}
	payload["inputs"] = []string{"自然语言需求", "目标仓库", "目标分支"}
	payload["outputs"] = []string{"结构化需求", "验收标准", "风险提示"}
	payload["nextActions"] = []string{"进入方案设计阶段", "结合仓库上下文识别影响范围"}
	payload["requirement"] = map[string]any{"title": ctx.Run.Title, "targetRepo": ctx.Run.TargetRepo, "targetBranch": ctx.Run.TargetBranch, "workBranch": ctx.Run.WorkBranch, "requirementExcerpt": summary, "acceptanceCriteria": acceptanceCriteria}
	return newStageResult(model.ArtifactStructuredRequirement, "结构化需求", payload, fmt.Sprintf("需求摘要：%s\n验收标准：\n- %s", summary, strings.Join(acceptanceCriteria, "\n- "))), nil
}
