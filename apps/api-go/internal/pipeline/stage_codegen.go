package pipeline

import (
	"context"
	"fmt"
	"strings"

	"feishu-pipeline/apps/api-go/internal/model"
)

type CodeGenerationHandler struct{}

func (CodeGenerationHandler) Execute(_ context.Context, ctx StageContext) (StageExecutionResult, error) {
	impactFiles := nestedStringSlice(ctx.Input, "latestArtifacts", model.ArtifactSolutionDesign, SchemaFieldImpactFiles)
	if len(impactFiles) == 0 {
		impactFiles = []string{"apps/api-go/internal/pipeline/executor.go", "apps/api-go/internal/service/pipeline_service.go"}
	}
	changeSet := make([]map[string]any, 0, len(impactFiles))
	patches := make([]map[string]any, 0, len(impactFiles))
	for _, file := range impactFiles {
		changeSet = append(changeSet, map[string]any{"filePath": file, "changeType": changeTypeForPath(file), "reason": "支撑 DevFlow Pipeline 阶段产物真实化与可演示闭环", "proposedPatchSummary": proposedPatchSummary(file)})
		patches = append(patches, map[string]any{"filePath": file, "changeType": changeTypeForPath(file), "patch": "", "reason": proposedPatchSummary(file)})
	}
	payload := baseStagePayload(ctx)
	payload[SchemaFieldSummary] = "生成受控代码变更计划，不直接写入工作区。"
	payload[SchemaFieldChangedFiles] = impactFiles
	payload[SchemaFieldPatches] = patches
	payload[SchemaFieldDiffSummary] = []string{"拆分阶段执行处理器", "补充仓库上下文召回", "输出结构化测试、评审和交付产物"}
	payload[SchemaFieldManualSteps] = []string{"后续阶段接入受控 patch apply 后再写入工作区"}
	payload["inputs"] = []string{"技术方案", "影响文件列表", "仓库上下文"}
	payload["outputs"] = []string{"changeSet", "diff 摘要", "待验证文件列表"}
	payload["risks"] = []string{"当前阶段尚未应用真实 patch，需要后续人工或安全执行器落地"}
	payload["nextActions"] = []string{"进入测试生成与执行阶段", "用测试报告验证变更计划风险"}
	payload["workBranch"] = ctx.Run.WorkBranch
	payload["changeSet"] = changeSet
	return newStageResult(model.ArtifactCodeDiff, "代码变更计划", payload, formatChangeSet(ctx.Run.WorkBranch, changeSet)), nil
}

func changeTypeForPath(path string) string {
	if strings.HasSuffix(path, "context.go") || strings.HasSuffix(path, "executor_test.go") {
		return "create"
	}
	return "modify"
}

func proposedPatchSummary(path string) string {
	switch {
	case strings.Contains(path, "executor"):
		return "拆分阶段处理逻辑并补齐结构化阶段产物。"
	case strings.Contains(path, "pipeline_service"):
		return "复用现有生命周期与 checkpoint 回退能力，补充可观测数据入口。"
	case strings.Contains(path, "controller") || strings.Contains(path, "types"):
		return "为前端工作台暴露阶段执行记录与产物字段。"
	case strings.Contains(path, "docs"):
		return "保持实现与产品开发设计、赛题要求一致。"
	default:
		return "按技术方案补充 DevFlow Pipeline 闭环能力。"
	}
}

func formatChangeSet(workBranch string, changeSet []map[string]any) string {
	lines := []string{fmt.Sprintf("工作分支：%s", workBranch), "变更计划："}
	for _, item := range changeSet {
		lines = append(lines, fmt.Sprintf("- %s [%s]：%s", item["filePath"], item["changeType"], item["proposedPatchSummary"]))
	}
	return strings.Join(lines, "\n")
}
