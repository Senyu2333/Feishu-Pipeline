package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/utils"
)

type EngineRepository interface {
	GetPipelineRunByID(context.Context, string) (model.PipelineRun, error)
	ListStageRunsByPipelineRunID(context.Context, string) ([]model.StageRun, error)
	ListArtifactsByPipelineRunID(context.Context, string) ([]model.Artifact, error)
	ListCheckpointsByPipelineRunID(context.Context, string) ([]model.Checkpoint, error)
	SaveStageRunInput(context.Context, string, string) error
	UpdateStageRunStatus(context.Context, string, model.StageRunStatus) error
	UpdatePipelineRunCurrentStage(context.Context, string, string) error
	UpdatePipelineRunStatus(context.Context, string, model.PipelineRunStatus) error
	CreateAgentRun(context.Context, *model.AgentRun) error
	UpdateAgentRunStatus(context.Context, string, model.AgentRunStatus, string, string) error
	SaveStageRunOutput(context.Context, string, string, string) error
	CreateArtifact(context.Context, *model.Artifact) error
	ResetCheckpointByStageRunID(context.Context, string) error
}

type Engine struct {
	repository EngineRepository
	executor   Executor
}

func NewEngine(repository EngineRepository, executor Executor) *Engine {
	return &Engine{repository: repository, executor: executor}
}

func (e *Engine) Run(ctx context.Context, runID string) error {
	run, err := e.repository.GetPipelineRunByID(ctx, runID)
	if err != nil {
		return err
	}
	stages, err := e.repository.ListStageRunsByPipelineRunID(ctx, runID)
	if err != nil {
		return err
	}
	artifacts, err := e.repository.ListArtifactsByPipelineRunID(ctx, runID)
	if err != nil {
		return err
	}
	checkpoints, err := e.repository.ListCheckpointsByPipelineRunID(ctx, runID)
	if err != nil {
		return err
	}

	for _, stage := range stages {
		if run.Status == model.PipelineRunPaused || run.Status == model.PipelineRunTerminated || run.Status == model.PipelineRunCompleted {
			return nil
		}
		if stage.Status == model.StageRunSucceeded || stage.Status == model.StageRunWaitingApproval {
			continue
		}
		if !IsRunnableStageStatus(stage.Status) {
			continue
		}
		if stage.StageType == model.StageTypeCheckpoint || ShouldStageWaitForApproval(stage.StageKey) {
			if err := e.repository.UpdateStageRunStatus(ctx, stage.ID, model.StageRunWaitingApproval); err != nil {
				return err
			}
			if err := e.repository.ResetCheckpointByStageRunID(ctx, stage.ID); err != nil {
				return err
			}
			if err := e.repository.UpdatePipelineRunCurrentStage(ctx, runID, stage.StageKey); err != nil {
				return err
			}
			if err := e.repository.UpdatePipelineRunStatus(ctx, runID, model.PipelineRunWaitingApproval); err != nil {
				return err
			}
			return nil
		}

		input := buildStageInput(run, stage, artifacts, checkpoints)
		inputJSON, _ := json.Marshal(input)
		if err := e.repository.SaveStageRunInput(ctx, stage.ID, string(inputJSON)); err != nil {
			return err
		}
		if err := e.repository.UpdateStageRunStatus(ctx, stage.ID, model.StageRunRunning); err != nil {
			return err
		}
		if err := e.repository.UpdatePipelineRunCurrentStage(ctx, runID, stage.StageKey); err != nil {
			return err
		}
		if err := e.repository.UpdatePipelineRunStatus(ctx, runID, model.PipelineRunRunning); err != nil {
			return err
		}

		now := time.Now().UTC()
		agentRun := model.AgentRun{
			ID:            utils.NewID("agentrun"),
			PipelineRunID: run.ID,
			StageRunID:    stage.ID,
			AgentKey:      stage.StageKey,
			Provider:      "internal",
			Model:         "deterministic",
			InputJSON:     string(inputJSON),
			Status:        model.AgentRunRunning,
			BaseModel:     model.BaseModel{CreatedAt: now, UpdatedAt: now},
		}
		if err := e.repository.CreateAgentRun(ctx, &agentRun); err != nil {
			return err
		}

		result, execErr := e.executor.Execute(ctx, StageContext{Run: run, Stage: stage, Artifacts: artifacts, Checkpoints: checkpoints, Input: input})
		if execErr != nil {
			_ = e.repository.UpdateStageRunStatus(ctx, stage.ID, model.StageRunFailed)
			_ = e.repository.SaveStageRunOutput(ctx, stage.ID, "", execErr.Error())
			_ = e.repository.UpdateAgentRunStatus(ctx, agentRun.ID, model.AgentRunFailed, "", execErr.Error())
			_ = e.repository.UpdatePipelineRunStatus(ctx, runID, model.PipelineRunFailed)
			return execErr
		}

		if err := e.repository.SaveStageRunOutput(ctx, stage.ID, result.OutputJSON, ""); err != nil {
			return err
		}
		if err := e.repository.UpdateStageRunStatus(ctx, stage.ID, model.StageRunSucceeded); err != nil {
			return err
		}
		if err := e.repository.UpdateAgentRunStatus(ctx, agentRun.ID, model.AgentRunSucceeded, result.OutputJSON, ""); err != nil {
			return err
		}

		artifact := model.Artifact{
			ID:            utils.NewID("artifact"),
			PipelineRunID: run.ID,
			StageRunID:    stage.ID,
			ArtifactType:  result.ArtifactType,
			Title:         result.Title,
			ContentText:   result.ContentText,
			ContentJSON:   result.ContentJSON,
			BaseModel:     model.BaseModel{CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		}
		if err := e.repository.CreateArtifact(ctx, &artifact); err != nil {
			return err
		}
		artifacts = append(artifacts, artifact)
	}

	if err := e.repository.UpdatePipelineRunStatus(ctx, runID, model.PipelineRunCompleted); err != nil {
		return err
	}
	if err := e.repository.UpdatePipelineRunCurrentStage(ctx, runID, StageDelivery); err != nil {
		return err
	}
	return nil
}

func buildStageInput(run model.PipelineRun, stage model.StageRun, artifacts []model.Artifact, checkpoints []model.Checkpoint) map[string]any {
	input := map[string]any{
		"runId":           run.ID,
		"stageKey":        stage.StageKey,
		"attempt":         stage.Attempt,
		"startedAt":       time.Now().UTC().Format(time.RFC3339),
		"requirement":     buildRunRequirement(run),
		"latestArtifacts": latestArtifactsByType(artifacts),
		"checkpoints":     checkpointInputs(checkpoints),
	}
	mergeJSONMap(input, stage.InputJSON)
	return input
}

func buildRunRequirement(run model.PipelineRun) map[string]any {
	return map[string]any{
		"title":           run.Title,
		"requirementText": run.RequirementText,
		"targetRepo":      run.TargetRepo,
		"targetBranch":    run.TargetBranch,
		"workBranch":      run.WorkBranch,
	}
}

func latestArtifactsByType(artifacts []model.Artifact) map[string]any {
	result := map[string]any{}
	for _, artifact := range artifacts {
		entry := map[string]any{
			"id":          artifact.ID,
			"title":       artifact.Title,
			"contentText": artifact.ContentText,
			"stageRunId":  artifact.StageRunID,
		}
		mergeJSONMap(entry, artifact.ContentJSON)
		result[string(artifact.ArtifactType)] = entry
	}
	return result
}

func checkpointInputs(checkpoints []model.Checkpoint) []map[string]any {
	items := make([]map[string]any, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		items = append(items, map[string]any{
			"id":         checkpoint.ID,
			"stageRunId": checkpoint.StageRunID,
			"status":     checkpoint.Status,
			"decision":   checkpoint.Decision,
			"comment":    checkpoint.Comment,
		})
	}
	return items
}

func mergeJSONMap(target map[string]any, raw string) {
	if raw == "" {
		return
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return
	}
	for key, value := range payload {
		target[key] = value
	}
}

func RequireRunID(runID string) error {
	if runID == "" {
		return fmt.Errorf("run id is required")
	}
	return nil
}
