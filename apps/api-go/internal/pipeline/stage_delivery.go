package pipeline

import (
	"context"
	"fmt"
	"strings"

	"feishu-pipeline/apps/api-go/internal/model"
)

type DeliveryHandler struct{}

func (DeliveryHandler) Execute(_ context.Context, ctx StageContext) (StageExecutionResult, error) {
	artifacts := artifactTitles(ctx.Artifacts)
	changeSet := nestedMapSlice(ctx.Input, "latestArtifacts", model.ArtifactCodeDiff, "changeSet")
	changedFiles := nestedStringSlice(ctx.Input, "latestArtifacts", model.ArtifactCodeDiff, SchemaFieldChangedFiles)
	payload := baseStagePayload(ctx)
	payload[SchemaFieldSummary] = "Pipeline 已完成需求、方案、变更计划、测试、评审和交付摘要产物。"
	payload[SchemaFieldChangedFiles] = changedFiles
	payload[SchemaFieldValidation] = []string{"已生成结构化需求", "已生成技术方案", "已生成代码变更计划", "已生成测试报告", "已生成评审报告"}
	payload[SchemaFieldPRTitle] = ctx.Run.Title
	payload[SchemaFieldPRBody] = "包含 Pipeline 阶段产物、测试报告、评审报告和交付摘要。"
	payload[SchemaFieldManualReleaseNotes] = []string{"当前阶段不自动 push，不自动创建远程 PR/MR"}
	payload["inputs"] = []string{"评审报告", "测试报告", "代码变更计划"}
	payload["outputs"] = []string{"交付摘要", "手动 PR/MR 建议", "演示检查清单"}
	payload["risks"] = []string{"尚未自动创建 PR/MR，交付动作需要下一阶段接入 Git Provider"}
	payload["nextActions"] = []string{"人工确认是否接入真实文件修改", "下一阶段实现 GitDelivery API"}
	payload["deliverySummary"] = payload[SchemaFieldSummary]
	payload["workBranch"] = ctx.Run.WorkBranch
	payload["artifacts"] = artifacts
	payload["changeSet"] = changeSet
	payload["manualPRSuggestion"] = map[string]any{"title": ctx.Run.Title, "branch": ctx.Run.WorkBranch, "includeArtifacts": artifacts}
	return newStageResult(model.ArtifactDeliverySummary, "交付摘要", payload, fmt.Sprintf("交付分支：%s\n已生成产物：%s\n下一步：人工确认后可接入 GitDelivery 创建 PR/MR。", ctx.Run.WorkBranch, strings.Join(artifacts, ", "))), nil
}
