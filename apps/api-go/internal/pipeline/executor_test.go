package pipeline_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/pipeline"
)

func TestSequentialExecutorBuildsSolutionDesignWithImpactFiles(t *testing.T) {
	executor := pipeline.NewSequentialExecutor()
	result, err := executor.Execute(context.Background(), stageContext(pipeline.StageSolutionDesign, model.StageTypeDesign, map[string]any{}))
	if err != nil {
		t.Fatalf("execute solution design: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result.OutputJSON), &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	impactFiles, ok := payload["impactFiles"].([]any)
	if !ok || len(impactFiles) == 0 {
		t.Fatalf("expected impactFiles in solution design output")
	}
	if payload["repositoryContext"] == nil {
		t.Fatalf("expected repositoryContext in solution design output")
	}
}

func TestSequentialExecutorBuildsCodeGenerationChangeSet(t *testing.T) {
	executor := pipeline.NewSequentialExecutor()
	input := map[string]any{"latestArtifacts": map[string]any{string(model.ArtifactSolutionDesign): map[string]any{"impactFiles": []any{"apps/api-go/internal/pipeline/executor.go"}}}}
	result, err := executor.Execute(context.Background(), stageContext(pipeline.StageCodeGeneration, model.StageTypeCodegen, input))
	if err != nil {
		t.Fatalf("execute code generation: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result.OutputJSON), &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	changeSet, ok := payload["changeSet"].([]any)
	if !ok || len(changeSet) != 1 {
		t.Fatalf("expected one changeSet item, got %v", payload["changeSet"])
	}
}

func TestSequentialExecutorBuildsReviewFromTestReport(t *testing.T) {
	executor := pipeline.NewSequentialExecutor()
	input := map[string]any{"latestArtifacts": map[string]any{
		string(model.ArtifactCodeDiff):   map[string]any{"changeSet": []any{map[string]any{"filePath": "apps/api-go/internal/pipeline/executor.go"}}},
		string(model.ArtifactTestReport): map[string]any{"summary": "生成测试计划并执行受控后端测试命令。"},
	}}
	result, err := executor.Execute(context.Background(), stageContext(pipeline.StageCodeReview, model.StageTypeReview, input))
	if err != nil {
		t.Fatalf("execute code review: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result.OutputJSON), &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if payload["conclusion"] != "needs_fix" {
		t.Fatalf("unexpected review conclusion: %v", payload["conclusion"])
	}
	issues, ok := payload["issues"].([]any)
	if !ok || len(issues) == 0 {
		t.Fatalf("expected review issues")
	}
}

func stageContext(stageKey string, stageType model.StageType, input map[string]any) pipeline.StageContext {
	now := time.Now().UTC()
	return pipeline.StageContext{
		Run: model.PipelineRun{
			ID:              "run_executor_test",
			TemplateID:      pipeline.DefaultTemplateID,
			Title:           "Executor test",
			RequirementText: "根据赛题要求补齐 pipeline stage artifact checkpoint agent delivery 能力",
			TargetRepo:      "self",
			TargetBranch:    "main",
			WorkBranch:      "devflow/executor-test",
			Status:          model.PipelineRunQueued,
			CurrentStageKey: stageKey,
			CreatedBy:       "tester",
			BaseModel:       model.BaseModel{CreatedAt: now, UpdatedAt: now},
		},
		Stage: model.StageRun{ID: "stage_executor_test", PipelineRunID: "run_executor_test", StageKey: stageKey, StageType: stageType, Status: model.StageRunQueued, Attempt: 1, BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now}},
		Input: input,
	}
}
