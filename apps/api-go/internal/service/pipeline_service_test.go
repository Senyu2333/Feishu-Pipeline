package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/pipeline"
	"feishu-pipeline/apps/api-go/internal/repo"
)

func TestRejectCheckpointResetsPreviousStage(t *testing.T) {
	repository := newPipelineTestRepository(t)
	service := NewPipelineService(repository)
	ctx := context.Background()
	now := time.Now().UTC()

	run := model.PipelineRun{
		ID:              "run_test_reject",
		TemplateID:      pipeline.DefaultTemplateID,
		Title:           "Reject checkpoint test",
		RequirementText: "需要在 reject 后回退到方案阶段",
		TargetRepo:      "self",
		TargetBranch:    "main",
		WorkBranch:      "devflow/test-reject",
		Status:          model.PipelineRunWaitingApproval,
		CurrentStageKey: pipeline.StageCheckpointDesign,
		CreatedBy:       "tester",
		BaseModel:       model.BaseModel{CreatedAt: now, UpdatedAt: now},
	}
	if err := repository.CreatePipelineRun(ctx, &run); err != nil {
		t.Fatalf("create run: %v", err)
	}
	stages := []model.StageRun{
		{ID: "stage_req_2", PipelineRunID: run.ID, StageKey: pipeline.StageRequirementAnalysis, StageType: model.StageTypeAnalysis, Status: model.StageRunSucceeded, Attempt: 1, BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now}},
		{ID: "stage_design_2", PipelineRunID: run.ID, StageKey: pipeline.StageSolutionDesign, StageType: model.StageTypeDesign, Status: model.StageRunSucceeded, Attempt: 1, OutputJSON: `{"summary":"old design"}`, BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now}},
		{ID: "stage_checkpoint_2", PipelineRunID: run.ID, StageKey: pipeline.StageCheckpointDesign, StageType: model.StageTypeCheckpoint, Status: model.StageRunWaitingApproval, Attempt: 1, BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now}},
		{ID: "stage_codegen_2", PipelineRunID: run.ID, StageKey: pipeline.StageCodeGeneration, StageType: model.StageTypeCodegen, Status: model.StageRunPending, Attempt: 1, BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now}},
	}
	if err := repository.CreateStageRuns(ctx, stages); err != nil {
		t.Fatalf("create stages: %v", err)
	}
	checkpoint := model.Checkpoint{ID: "cp_design_2", PipelineRunID: run.ID, StageRunID: "stage_checkpoint_2", CheckpointType: model.CheckpointDesignReview, Status: model.CheckpointPending, BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now}}
	if err := repository.CreateCheckpoint(ctx, &checkpoint); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
	artifact := model.Artifact{ID: "artifact_design_2", PipelineRunID: run.ID, StageRunID: "stage_design_2", ArtifactType: model.ArtifactSolutionDesign, Title: "技术方案", ContentJSON: `{"summary":"old design"}`, MetaJSON: `{}`, BaseModel: model.BaseModel{CreatedAt: now, UpdatedAt: now}}
	if err := repository.CreateArtifact(ctx, &artifact); err != nil {
		t.Fatalf("create artifact: %v", err)
	}

	updatedCheckpoint, err := service.RejectCheckpoint(ctx, checkpoint.ID, "方案需要补充回退信息", "reviewer")
	if err != nil {
		t.Fatalf("reject checkpoint: %v", err)
	}
	if updatedCheckpoint.Status != model.CheckpointRejected {
		t.Fatalf("expected rejected checkpoint, got %s", updatedCheckpoint.Status)
	}

	designStage, err := repository.GetStageRunByKey(ctx, run.ID, pipeline.StageSolutionDesign)
	if err != nil {
		t.Fatalf("get design stage: %v", err)
	}
	if designStage.Status != model.StageRunQueued {
		t.Fatalf("expected design stage queued, got %s", designStage.Status)
	}
	if designStage.Attempt != 2 {
		t.Fatalf("expected design stage attempt 2, got %d", designStage.Attempt)
	}
	if designStage.InputJSON == "" {
		t.Fatalf("expected reject context on design stage input")
	}
	if !strings.Contains(designStage.InputJSON, "方案需要补充回退信息") {
		t.Fatalf("expected reject reason in design stage input: %s", designStage.InputJSON)
	}

	checkpointStage, err := repository.GetStageRunByKey(ctx, run.ID, pipeline.StageCheckpointDesign)
	if err != nil {
		t.Fatalf("get checkpoint stage: %v", err)
	}
	if checkpointStage.Status != model.StageRunPending {
		t.Fatalf("expected checkpoint stage pending, got %s", checkpointStage.Status)
	}

	updatedRun, err := repository.GetPipelineRunByID(ctx, run.ID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if updatedRun.Status != model.PipelineRunQueued {
		t.Fatalf("expected queued run, got %s", updatedRun.Status)
	}
	if updatedRun.CurrentStageKey != pipeline.StageSolutionDesign {
		t.Fatalf("expected current stage %s, got %s", pipeline.StageSolutionDesign, updatedRun.CurrentStageKey)
	}
}

func newPipelineTestRepository(t *testing.T) *repo.Repository {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), "pipeline-service-test.db")
	repository, err := repo.NewSQLiteRepository(databasePath)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	t.Cleanup(func() {
		_ = repository.Close()
	})
	return repository
}
