package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/utils"
)

type StageExecutionResult struct {
	ArtifactType model.ArtifactType
	Title        string
	ContentText  string
	ContentJSON  string
	OutputJSON   string
}

type Executor interface {
	Execute(context.Context, StageContext) (StageExecutionResult, error)
}

type StageContext struct {
	Run         model.PipelineRun
	Stage       model.StageRun
	Artifacts   []model.Artifact
	Checkpoints []model.Checkpoint
	Input       map[string]any
}

type SequentialExecutor struct{}

func NewSequentialExecutor() *SequentialExecutor {
	return &SequentialExecutor{}
}

func (e *SequentialExecutor) Execute(ctx context.Context, stageContext StageContext) (StageExecutionResult, error) {
	switch stageContext.Stage.StageKey {
	case StageRequirementAnalysis:
		return e.executeRequirementAnalysis(stageContext), nil
	case StageSolutionDesign:
		return e.executeSolutionDesign(stageContext), nil
	case StageCodeGeneration:
		return e.executeCodeGeneration(stageContext), nil
	case StageTestGeneration:
		return e.executeTestGeneration(ctx, stageContext)
	case StageCodeReview:
		return e.executeCodeReview(stageContext), nil
	case StageDelivery:
		return e.executeDelivery(stageContext), nil
	default:
		return e.executeDefault(stageContext), nil
	}
}

func (e *SequentialExecutor) executeRequirementAnalysis(ctx StageContext) StageExecutionResult {
	summary := utils.Summarize(ctx.Run.RequirementText, 240)
	acceptanceCriteria := []string{
		"需求目标已被结构化描述",
		"影响范围已被识别",
		"方案设计可直接消费该结果",
	}
	payload := baseStagePayload(ctx)
	payload["summary"] = summary
	payload["inputs"] = []string{"自然语言需求", "目标仓库", "目标分支"}
	payload["outputs"] = []string{"结构化需求", "验收标准", "风险提示"}
	payload["risks"] = []string{"需求仍可能存在业务边界不清，需要在方案审批时确认"}
	payload["nextActions"] = []string{"进入方案设计阶段", "结合仓库上下文识别影响范围"}
	payload["requirement"] = map[string]any{
		"title":              ctx.Run.Title,
		"targetRepo":         ctx.Run.TargetRepo,
		"targetBranch":       ctx.Run.TargetBranch,
		"workBranch":         ctx.Run.WorkBranch,
		"requirementExcerpt": summary,
		"acceptanceCriteria": acceptanceCriteria,
	}
	return newStageResult(model.ArtifactStructuredRequirement, "结构化需求", payload, fmt.Sprintf("需求摘要：%s\n验收标准：\n- %s", summary, strings.Join(acceptanceCriteria, "\n- ")))
}

func (e *SequentialExecutor) executeSolutionDesign(ctx StageContext) StageExecutionResult {
	requirementSummary := nestedString(ctx.Input, "latestArtifacts", model.ArtifactStructuredRequirement, "summary")
	if requirementSummary == "" {
		requirementSummary = utils.Summarize(ctx.Run.RequirementText, 180)
	}
	repoContext := BuildRepositoryContext(ctx.Run.RequirementText + " " + requirementSummary)
	impactFiles := impactFilesFromContext(repoContext)
	if len(impactFiles) == 0 {
		impactFiles = []string{
			"apps/api-go/internal/pipeline/executor.go",
			"apps/api-go/internal/pipeline/engine.go",
			"apps/api-go/internal/service/pipeline_service.go",
		}
	}
	implementationPlan := []string{
		"复用现有 PipelineRun、StageRun、Artifact、Checkpoint 模型承载状态与产物",
		"在阶段执行器中补充结构化输出契约，保证下游阶段只消费明确字段",
		"通过只读仓库上下文识别影响文件，先生成变更计划而不直接写入文件",
		"在测试阶段执行受控命令并沉淀测试报告，支撑评审与交付摘要",
	}
	payload := baseStagePayload(ctx)
	payload["summary"] = requirementSummary
	payload["inputs"] = []string{"结构化需求", "仓库上下文", "历史 checkpoint 意见"}
	payload["outputs"] = []string{"技术方案", "影响文件", "实现步骤", "风险说明"}
	payload["risks"] = []string{"当前阶段只做静态上下文召回，尚未接语义索引", "代码生成阶段将输出计划而不是直接改文件"}
	payload["nextActions"] = []string{"等待方案审批", "审批通过后生成代码变更计划"}
	payload["repositoryContext"] = repoContext
	payload["impactFiles"] = impactFiles
	payload["apiChanges"] = []string{"保持现有 Pipeline API，新增 AgentRun 可观测查询接口"}
	payload["implementationPlan"] = implementationPlan
	payload["riskNotes"] = payload["risks"]
	return newStageResult(model.ArtifactSolutionDesign, "技术方案", payload, fmt.Sprintf("方案摘要：%s\n影响文件：\n- %s\n实现步骤：\n- %s", requirementSummary, strings.Join(impactFiles, "\n- "), strings.Join(implementationPlan, "\n- ")))
}

func (e *SequentialExecutor) executeCodeGeneration(ctx StageContext) StageExecutionResult {
	impactFiles := nestedStringSlice(ctx.Input, "latestArtifacts", model.ArtifactSolutionDesign, "impactFiles")
	if len(impactFiles) == 0 {
		impactFiles = []string{"apps/api-go/internal/pipeline/executor.go", "apps/api-go/internal/service/pipeline_service.go"}
	}
	changeSet := make([]map[string]any, 0, len(impactFiles))
	for _, file := range impactFiles {
		changeSet = append(changeSet, map[string]any{
			"filePath":             file,
			"changeType":           changeTypeForPath(file),
			"reason":               "支撑 DevFlow Pipeline 阶段产物真实化与可演示闭环",
			"proposedPatchSummary": proposedPatchSummary(file),
		})
	}
	payload := baseStagePayload(ctx)
	payload["summary"] = "生成受控代码变更计划，不直接写入工作区。"
	payload["inputs"] = []string{"技术方案", "影响文件列表", "仓库上下文"}
	payload["outputs"] = []string{"changeSet", "diff 摘要", "待验证文件列表"}
	payload["risks"] = []string{"当前阶段尚未应用真实 patch，需要后续人工或安全执行器落地"}
	payload["nextActions"] = []string{"进入测试生成与执行阶段", "用测试报告验证变更计划风险"}
	payload["workBranch"] = ctx.Run.WorkBranch
	payload["changedFiles"] = impactFiles
	payload["changeSet"] = changeSet
	payload["diffSummary"] = []string{"拆分阶段执行处理器", "补充仓库上下文召回", "输出结构化测试、评审和交付产物"}
	return newStageResult(model.ArtifactCodeDiff, "代码变更计划", payload, formatChangeSet(ctx.Run.WorkBranch, changeSet))
}

func (e *SequentialExecutor) executeTestGeneration(ctx context.Context, stageContext StageContext) (StageExecutionResult, error) {
	changedFiles := nestedStringSlice(stageContext.Input, "latestArtifacts", model.ArtifactCodeDiff, "changedFiles")
	tests := []string{
		"pipeline engine 顺序执行与阶段输入组装",
		"checkpoint approve/reject 状态回流",
		"pipeline service 创建 run 与模板装载",
	}
	report := map[string]any{"status": "skipped", "reason": "target repo is not self"}
	if stageContext.Run.TargetRepo == "self" {
		report = runGoPipelineTests(ctx)
	}
	payload := baseStagePayload(stageContext)
	payload["summary"] = "生成测试计划并执行受控后端测试命令。"
	payload["inputs"] = []string{"代码变更计划", "结构化需求"}
	payload["outputs"] = []string{"测试建议", "命令执行结果", "失败摘要"}
	payload["risks"] = []string{"测试范围暂聚焦 pipeline/service/repo，尚未覆盖前端和完整集成链路"}
	payload["nextActions"] = []string{"进入代码评审阶段", "由评审阶段结合测试结果判断风险"}
	payload["changedFiles"] = changedFiles
	payload["tests"] = tests
	payload["commandReport"] = report
	return newStageResult(model.ArtifactTestReport, "测试报告", payload, formatTestReport(changedFiles, tests, report)), nil
}

func (e *SequentialExecutor) executeCodeReview(ctx StageContext) StageExecutionResult {
	changeSet := nestedMapSlice(ctx.Input, "latestArtifacts", model.ArtifactCodeDiff, "changeSet")
	testStatus := nestedString(ctx.Input, "latestArtifacts", model.ArtifactTestReport, "summary")
	issues := []map[string]any{
		{"severity": "medium", "filePath": "apps/api-go/internal/pipeline/executor.go", "message": "当前代码生成阶段仍为计划产物，尚未应用真实 patch。", "suggestion": "下一阶段接入受控文件执行器前继续保留人工确认。"},
	}
	payload := baseStagePayload(ctx)
	payload["summary"] = "从正确性、安全性、可维护性维度完成 AI 预审。"
	payload["inputs"] = []string{"代码变更计划", "测试报告", "技术方案"}
	payload["outputs"] = []string{"评审结论", "问题列表", "安全说明", "维护性说明"}
	payload["risks"] = []string{"真实写文件与 Git 交付尚未启用，需要人工确认后进入下一阶段"}
	payload["nextActions"] = []string{"等待评审确认 checkpoint", "确认通过后生成交付摘要"}
	payload["conclusion"] = "needs_human_confirmation"
	payload["issues"] = issues
	payload["securityNotes"] = []string{"本阶段未执行写文件、提交、推送等高风险操作", "测试命令限制在 Go 后端 pipeline 相关包"}
	payload["maintainabilityNotes"] = []string{"阶段执行逻辑已按处理器拆分，便于后续替换为真实 Agent Provider"}
	payload["changeSet"] = changeSet
	payload["testStatus"] = testStatus
	payload["checkpoints"] = checkpointSummaries(ctx.Checkpoints)
	return newStageResult(model.ArtifactReviewReport, "评审报告", payload, "评审结论：待人工确认\n问题列表：\n- 当前代码生成阶段仍为计划产物，尚未应用真实 patch。\n安全说明：未执行写文件、提交、推送等高风险操作。")
}

func (e *SequentialExecutor) executeDelivery(ctx StageContext) StageExecutionResult {
	artifacts := artifactTitles(ctx.Artifacts)
	changeSet := nestedMapSlice(ctx.Input, "latestArtifacts", model.ArtifactCodeDiff, "changeSet")
	payload := baseStagePayload(ctx)
	payload["summary"] = "Pipeline 已完成需求、方案、变更计划、测试、评审和交付摘要产物。"
	payload["inputs"] = []string{"评审报告", "测试报告", "代码变更计划"}
	payload["outputs"] = []string{"交付摘要", "手动 PR/MR 建议", "演示检查清单"}
	payload["risks"] = []string{"尚未自动创建 PR/MR，交付动作需要下一阶段接入 Git Provider"}
	payload["nextActions"] = []string{"人工确认是否接入真实文件修改", "下一阶段实现 GitDelivery API"}
	payload["deliverySummary"] = "Pipeline 已输出结构化需求、技术方案、代码变更计划、测试报告、评审报告和交付摘要。"
	payload["workBranch"] = ctx.Run.WorkBranch
	payload["artifacts"] = artifacts
	payload["changeSet"] = changeSet
	payload["manualPRSuggestion"] = map[string]any{"title": ctx.Run.Title, "branch": ctx.Run.WorkBranch, "includeArtifacts": artifacts}
	return newStageResult(model.ArtifactDeliverySummary, "交付摘要", payload, fmt.Sprintf("交付分支：%s\n已生成产物：%s\n下一步：人工确认后可接入 GitDelivery 创建 PR/MR。", ctx.Run.WorkBranch, strings.Join(artifacts, ", ")))
}

func (e *SequentialExecutor) executeDefault(ctx StageContext) StageExecutionResult {
	payload := baseStagePayload(ctx)
	payload["summary"] = "阶段执行完成。"
	return newStageResult(model.ArtifactDeliverySummary, ctx.Stage.StageKey, payload, "阶段执行完成。")
}

func baseStagePayload(ctx StageContext) map[string]any {
	return map[string]any{
		"runId":       ctx.Run.ID,
		"stageKey":    ctx.Stage.StageKey,
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"attempt":     ctx.Stage.Attempt,
		"input":       ctx.Input,
	}
}

func newStageResult(artifactType model.ArtifactType, title string, payload map[string]any, contentText string) StageExecutionResult {
	contentJSON, _ := json.Marshal(payload)
	return StageExecutionResult{ArtifactType: artifactType, Title: title, ContentText: contentText, ContentJSON: string(contentJSON), OutputJSON: string(contentJSON)}
}

func impactFilesFromContext(context RepositoryContext) []string {
	result := make([]string, 0, len(context.CandidateFiles))
	seen := map[string]bool{}
	for _, item := range context.CandidateFiles {
		if seen[item.Path] {
			continue
		}
		seen[item.Path] = true
		result = append(result, item.Path)
	}
	return result
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

func runGoPipelineTests(ctx context.Context) map[string]any {
	commandContext, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(commandContext, "go", "test", "./internal/pipeline", "./internal/service")
	cmd.Dir = findAPIModuleDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	status := "passed"
	exitCode := 0
	if err != nil {
		status = "failed"
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	if commandContext.Err() == context.DeadlineExceeded {
		status = "timeout"
		exitCode = -1
	}
	return map[string]any{
		"status":   status,
		"command":  "go test ./internal/pipeline ./internal/service",
		"exitCode": exitCode,
		"stdout":   utils.Summarize(stdout.String(), 1200),
		"stderr":   utils.Summarize(stderr.String(), 1200),
	}
}

func findAPIModuleDir() string {
	root := findWorkspaceRoot()
	return root + "/apps/api-go"
}

func formatTestReport(changedFiles []string, tests []string, report map[string]any) string {
	return fmt.Sprintf("关联文件：%s\n测试项：\n- %s\n命令：%s\n状态：%s\n输出：%s%s", strings.Join(changedFiles, ", "), strings.Join(tests, "\n- "), report["command"], report["status"], report["stdout"], stderrLine(report))
}

func stderrLine(report map[string]any) string {
	stderr, _ := report["stderr"].(string)
	if strings.TrimSpace(stderr) == "" {
		return ""
	}
	return "\n错误输出：" + stderr
}

func nestedString(input map[string]any, key string, artifactType model.ArtifactType, field string) string {
	latestArtifacts, ok := input[key].(map[string]any)
	if !ok {
		return ""
	}
	artifact, ok := latestArtifacts[string(artifactType)].(map[string]any)
	if !ok {
		return ""
	}
	value, _ := artifact[field].(string)
	return value
}

func nestedStringSlice(input map[string]any, key string, artifactType model.ArtifactType, field string) []string {
	latestArtifacts, ok := input[key].(map[string]any)
	if !ok {
		return nil
	}
	artifact, ok := latestArtifacts[string(artifactType)].(map[string]any)
	if !ok {
		return nil
	}
	items, ok := artifact[field].([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if ok && text != "" {
			result = append(result, text)
		}
	}
	return result
}

func nestedMapSlice(input map[string]any, key string, artifactType model.ArtifactType, field string) []map[string]any {
	latestArtifacts, ok := input[key].(map[string]any)
	if !ok {
		return nil
	}
	artifact, ok := latestArtifacts[string(artifactType)].(map[string]any)
	if !ok {
		return nil
	}
	items, ok := artifact[field].([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if ok {
			result = append(result, entry)
		}
	}
	return result
}

func checkpointSummaries(items []model.Checkpoint) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"id":       item.ID,
			"status":   item.Status,
			"decision": item.Decision,
			"comment":  item.Comment,
		})
	}
	return result
}

func artifactTitles(items []model.Artifact) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item.Title != "" {
			result = append(result, item.Title)
		}
	}
	return result
}
