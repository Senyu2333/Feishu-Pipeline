package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/utils"
)

type TestGenerationHandler struct{}

func (TestGenerationHandler) Execute(ctx context.Context, stageContext StageContext) (StageExecutionResult, error) {
	changedFiles := nestedStringSlice(stageContext.Input, "latestArtifacts", model.ArtifactCodeDiff, SchemaFieldChangedFiles)
	tests := []string{"pipeline engine 顺序执行与阶段输入组装", "checkpoint approve/reject 状态回流", "pipeline service 创建 run 与模板装载"}
	report := map[string]any{"status": "skipped", "reason": "target repo is not self", "command": "", "stdout": "", "stderr": "", "exitCode": 0}
	if stageContext.Run.TargetRepo == "self" {
		report = runGoPipelineTests(ctx)
	}
	payload := baseStagePayload(stageContext)
	payload[SchemaFieldSummary] = "生成测试计划并执行受控后端测试命令。"
	payload[SchemaFieldTestPlan] = tests
	payload[SchemaFieldCommands] = []string{fmt.Sprintf("%v", report["command"])}
	payload[SchemaFieldCommandResults] = []map[string]any{report}
	payload[SchemaFieldStatus] = report["status"]
	payload[SchemaFieldChangedFiles] = changedFiles
	payload["inputs"] = []string{"代码变更计划", "结构化需求"}
	payload["outputs"] = []string{"测试建议", "命令执行结果", "失败摘要"}
	payload["risks"] = []string{"测试范围暂聚焦 pipeline/service/repo，尚未覆盖前端和完整集成链路"}
	payload["nextActions"] = []string{"进入代码评审阶段", "由评审阶段结合测试结果判断风险"}
	return newStageResult(model.ArtifactTestReport, "测试报告", payload, formatTestReport(changedFiles, tests, report)), nil
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
	startedAt := time.Now()
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
	return map[string]any{"status": status, "command": "go test ./internal/pipeline ./internal/service", "exitCode": exitCode, "stdout": utils.Summarize(stdout.String(), 1200), "stderr": utils.Summarize(stderr.String(), 1200), "durationMs": time.Since(startedAt).Milliseconds()}
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
