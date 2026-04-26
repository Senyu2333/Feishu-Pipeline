package pipeline

import (
	"context"
	"fmt"
	"strings"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/utils"
)

type SolutionDesignHandler struct{}

func (SolutionDesignHandler) Execute(_ context.Context, ctx StageContext) (StageExecutionResult, error) {
	requirementSummary := nestedString(ctx.Input, "latestArtifacts", model.ArtifactStructuredRequirement, SchemaFieldSummary)
	if requirementSummary == "" {
		requirementSummary = utils.Summarize(ctx.Run.RequirementText, 180)
	}
	repoContext := BuildRepositoryContext(ctx.Run.RequirementText + " " + requirementSummary)
	impactFiles := impactFilesFromContext(repoContext)
	if len(impactFiles) == 0 {
		impactFiles = []string{"apps/api-go/internal/pipeline/executor.go", "apps/api-go/internal/pipeline/engine.go", "apps/api-go/internal/service/pipeline_service.go"}
	}
	implementationPlan := []string{"复用现有 PipelineRun、StageRun、Artifact、Checkpoint 模型承载状态与产物", "在阶段执行器中补充结构化输出契约，保证下游阶段只消费明确字段", "通过只读仓库上下文识别影响文件，先生成变更计划而不直接写入文件", "在测试阶段执行受控命令并沉淀测试报告，支撑评审与交付摘要"}
	payload := baseStagePayload(ctx)
	payload[SchemaFieldSummary] = requirementSummary
	payload[SchemaFieldImpactFiles] = impactFiles
	payload[SchemaFieldAPIChanges] = []string{"保持现有 Pipeline API，新增 AgentRun 可观测查询接口"}
	payload[SchemaFieldDataModelChanges] = []string{"继续复用 PipelineRun、StageRun、Artifact、Checkpoint、AgentRun"}
	payload[SchemaFieldImplementationPlan] = implementationPlan
	payload["risks"] = []string{"当前阶段只做静态上下文召回，尚未接语义索引", "代码生成阶段将输出计划而不是直接改文件"}
	payload["inputs"] = []string{"结构化需求", "仓库上下文", "历史 checkpoint 意见"}
	payload["outputs"] = []string{"技术方案", "影响文件", "实现步骤", "风险说明"}
	payload["nextActions"] = []string{"等待方案审批", "审批通过后生成代码变更计划"}
	payload["repositoryContext"] = repoContext
	return newStageResult(model.ArtifactSolutionDesign, "技术方案", payload, fmt.Sprintf("方案摘要：%s\n影响文件：\n- %s\n实现步骤：\n- %s", requirementSummary, strings.Join(impactFiles, "\n- "), strings.Join(implementationPlan, "\n- "))), nil
}
