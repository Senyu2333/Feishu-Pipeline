package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/pipeline"
	"feishu-pipeline/apps/api-go/internal/repo"
	"feishu-pipeline/apps/api-go/internal/utils"
)

type PipelineService struct {
	repository *repo.Repository
}

type PipelineRunDetail struct {
	Run         model.PipelineRun
	Stages      []model.StageRun
	Artifacts   []model.Artifact
	Checkpoints []model.Checkpoint
}

type CreatePipelineRunInput struct {
	TemplateID      string
	Title           string
	RequirementText string
	TargetRepo      string
	TargetBranch    string
	SourceSessionID string
	CreatedBy       string
}

func NewPipelineService(repository *repo.Repository) *PipelineService {
	return &PipelineService{repository: repository}
}

func (s *PipelineService) ListPipelineTemplates(ctx context.Context) ([]model.PipelineTemplate, error) {
	return s.repository.ListPipelineTemplates(ctx)
}

func (s *PipelineService) ListPipelineRuns(ctx context.Context) ([]model.PipelineRun, error) {
	return s.repository.ListPipelineRuns(ctx)
}

func (s *PipelineService) GetPipelineRunDetail(ctx context.Context, runID string) (*PipelineRunDetail, error) {
	run, err := s.repository.GetPipelineRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	stages, err := s.repository.ListStageRunsByPipelineRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	artifacts, err := s.repository.ListArtifactsByPipelineRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	checkpoints, err := s.repository.ListCheckpointsByPipelineRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	return &PipelineRunDetail{Run: run, Stages: stages, Artifacts: artifacts, Checkpoints: checkpoints}, nil
}

func (s *PipelineService) CreatePipelineRun(ctx context.Context, input CreatePipelineRunInput) (*PipelineRunDetail, error) {
	templateID := strings.TrimSpace(input.TemplateID)
	if templateID == "" {
		templateID = pipeline.DefaultTemplateID
	}
	if _, err := s.repository.GetPipelineTemplateByID(ctx, templateID); err != nil {
		return nil, err
	}

	targetRepo := strings.TrimSpace(input.TargetRepo)
	if targetRepo == "" {
		targetRepo = "self"
	}
	targetBranch := strings.TrimSpace(input.TargetBranch)
	if targetBranch == "" {
		targetBranch = "main"
	}
	now := time.Now().UTC()
	run := model.PipelineRun{
		ID:              utils.NewID("run"),
		TemplateID:      templateID,
		Title:           strings.TrimSpace(input.Title),
		RequirementText: strings.TrimSpace(input.RequirementText),
		SourceSessionID: strings.TrimSpace(input.SourceSessionID),
		TargetRepo:      targetRepo,
		TargetBranch:    targetBranch,
		WorkBranch:      fmt.Sprintf("devflow/%s", strings.TrimPrefix(utils.NewID("branch"), "branch_")),
		Status:          model.PipelineRunDraft,
		CurrentStageKey: pipeline.DefaultStageDefinitions[0].Key,
		CreatedBy:       input.CreatedBy,
		BaseModel:       model.BaseModel{CreatedAt: now, UpdatedAt: now},
	}
	if err := s.repository.CreatePipelineRun(ctx, &run); err != nil {
		return nil, err
	}

	stageRuns := make([]model.StageRun, 0, len(pipeline.DefaultStageDefinitions))
	checkpoints := make([]model.Checkpoint, 0, 2)
	for idx, stage := range pipeline.DefaultStageDefinitions {
		status := model.StageRunPending
		if idx == 0 {
			status = model.StageRunQueued
		}
		stageRun := model.StageRun{
			ID:            utils.NewID("stage"),
			PipelineRunID: run.ID,
			StageKey:      stage.Key,
			StageType:     stage.Type,
			Status:        status,
			Attempt:       1,
			BaseModel:     model.BaseModel{CreatedAt: now, UpdatedAt: now},
		}
		stageRuns = append(stageRuns, stageRun)
		if stage.IsCheckpoint {
			checkpoints = append(checkpoints, model.Checkpoint{
				ID:             utils.NewID("checkpoint"),
				PipelineRunID:  run.ID,
				StageRunID:     stageRun.ID,
				CheckpointType: checkpointTypeForStage(stage.Key),
				Status:         model.CheckpointPending,
				BaseModel:      model.BaseModel{CreatedAt: now, UpdatedAt: now},
			})
		}
	}
	if err := s.repository.CreateStageRuns(ctx, stageRuns); err != nil {
		return nil, err
	}
	for idx := range checkpoints {
		if err := s.repository.CreateCheckpoint(ctx, &checkpoints[idx]); err != nil {
			return nil, err
		}
	}

	artifact := model.Artifact{
		ID:            utils.NewID("artifact"),
		PipelineRunID: run.ID,
		ArtifactType:  model.ArtifactStructuredRequirement,
		Title:         "结构化需求输入",
		ContentText:   run.RequirementText,
		BaseModel:     model.BaseModel{CreatedAt: now, UpdatedAt: now},
	}
	if err := s.repository.CreateArtifact(ctx, &artifact); err != nil {
		return nil, err
	}

	return s.GetPipelineRunDetail(ctx, run.ID)
}

func (s *PipelineService) CreatePipelineRunFromSession(ctx context.Context, sessionID string, templateID string, targetRepo string, targetBranch string, createdBy string) (*PipelineRunDetail, error) {
	session, err := s.repository.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	messages, err := s.repository.ListMessagesBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	requirementText := session.Summary
	if len(messages) > 0 {
		parts := make([]string, 0, len(messages))
		for _, message := range messages {
			parts = append(parts, fmt.Sprintf("[%s] %s", message.Role, message.Content))
		}
		requirementText = strings.Join(parts, "\n")
	}
	if strings.TrimSpace(requirementText) == "" {
		requirementText = session.Title
	}
	return s.CreatePipelineRun(ctx, CreatePipelineRunInput{TemplateID: templateID, Title: session.Title, RequirementText: requirementText, TargetRepo: targetRepo, TargetBranch: targetBranch, SourceSessionID: session.ID, CreatedBy: createdBy})
}

func (s *PipelineService) ListStageRuns(ctx context.Context, runID string) ([]model.StageRun, error) {
	return s.repository.ListStageRunsByPipelineRunID(ctx, runID)
}

func (s *PipelineService) ListArtifacts(ctx context.Context, runID string) ([]model.Artifact, error) {
	return s.repository.ListArtifactsByPipelineRunID(ctx, runID)
}

func (s *PipelineService) ListCheckpoints(ctx context.Context, runID string) ([]model.Checkpoint, error) {
	return s.repository.ListCheckpointsByPipelineRunID(ctx, runID)
}

func (s *PipelineService) StartPipelineRun(ctx context.Context, runID string) (model.PipelineRun, error) {
	if err := s.repository.UpdatePipelineRunStatus(ctx, runID, model.PipelineRunRunning); err != nil {
		return model.PipelineRun{}, err
	}
	stages, err := s.repository.ListStageRunsByPipelineRunID(ctx, runID)
	if err != nil {
		return model.PipelineRun{}, err
	}
	if len(stages) > 0 {
		firstStage := stages[0]
		if firstStage.StageType == model.StageTypeCheckpoint {
			if err := s.repository.UpdateStageRunStatus(ctx, firstStage.ID, model.StageRunWaitingApproval); err != nil {
				return model.PipelineRun{}, err
			}
			if err := s.repository.UpdatePipelineRunStatus(ctx, runID, model.PipelineRunWaitingApproval); err != nil {
				return model.PipelineRun{}, err
			}
		} else {
			if err := s.repository.UpdateStageRunStatus(ctx, firstStage.ID, model.StageRunRunning); err != nil {
				return model.PipelineRun{}, err
			}
		}
		if err := s.repository.UpdatePipelineRunCurrentStage(ctx, runID, firstStage.StageKey); err != nil {
			return model.PipelineRun{}, err
		}
	}
	return s.repository.GetPipelineRunByID(ctx, runID)
}

func (s *PipelineService) PausePipelineRun(ctx context.Context, runID string) (model.PipelineRun, error) {
	if err := s.repository.UpdatePipelineRunStatus(ctx, runID, model.PipelineRunPaused); err != nil {
		return model.PipelineRun{}, err
	}
	return s.repository.GetPipelineRunByID(ctx, runID)
}

func (s *PipelineService) ResumePipelineRun(ctx context.Context, runID string) (model.PipelineRun, error) {
	if err := s.repository.UpdatePipelineRunStatus(ctx, runID, model.PipelineRunRunning); err != nil {
		return model.PipelineRun{}, err
	}
	return s.repository.GetPipelineRunByID(ctx, runID)
}

func (s *PipelineService) TerminatePipelineRun(ctx context.Context, runID string) (model.PipelineRun, error) {
	if err := s.repository.UpdatePipelineRunStatus(ctx, runID, model.PipelineRunTerminated); err != nil {
		return model.PipelineRun{}, err
	}
	return s.repository.GetPipelineRunByID(ctx, runID)
}

func (s *PipelineService) ApproveCheckpoint(ctx context.Context, checkpointID string, comment string, approverID string) (model.Checkpoint, error) {
	if err := s.repository.UpdateCheckpointDecision(ctx, checkpointID, model.CheckpointApproved, "approve", strings.TrimSpace(comment), approverID); err != nil {
		return model.Checkpoint{}, err
	}
	checkpoint, err := s.repository.GetCheckpointByID(ctx, checkpointID)
	if err != nil {
		return model.Checkpoint{}, err
	}
	if checkpoint.PipelineRunID != "" {
		_ = s.repository.UpdatePipelineRunStatus(ctx, checkpoint.PipelineRunID, model.PipelineRunRunning)
	}
	return checkpoint, nil
}

func (s *PipelineService) RejectCheckpoint(ctx context.Context, checkpointID string, comment string, approverID string) (model.Checkpoint, error) {
	if err := s.repository.UpdateCheckpointDecision(ctx, checkpointID, model.CheckpointRejected, "reject", strings.TrimSpace(comment), approverID); err != nil {
		return model.Checkpoint{}, err
	}
	checkpoint, err := s.repository.GetCheckpointByID(ctx, checkpointID)
	if err != nil {
		return model.Checkpoint{}, err
	}
	if checkpoint.PipelineRunID != "" {
		_ = s.repository.UpdatePipelineRunStatus(ctx, checkpoint.PipelineRunID, model.PipelineRunWaitingApproval)
	}
	return checkpoint, nil
}

func checkpointTypeForStage(stageKey string) model.CheckpointType {
	switch stageKey {
	case pipeline.StageCheckpointDesign:
		return model.CheckpointDesignReview
	case pipeline.StageCheckpointReview:
		return model.CheckpointCodeReview
	default:
		return model.CheckpointCodeReview
	}
}
