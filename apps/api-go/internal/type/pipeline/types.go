package pipelinetype

import (
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
)

type CreatePipelineRunRequest struct {
	TemplateID      string `json:"templateId"`
	Title           string `json:"title" binding:"required"`
	RequirementText string `json:"requirementText" binding:"required"`
	TargetRepo      string `json:"targetRepo"`
	TargetBranch    string `json:"targetBranch"`
}

type CreatePipelineRunFromSessionRequest struct {
	SessionID    string `json:"sessionId" binding:"required"`
	TemplateID   string `json:"templateId"`
	TargetRepo   string `json:"targetRepo"`
	TargetBranch string `json:"targetBranch"`
}

type UpdateCheckpointDecisionRequest struct {
	Comment string `json:"comment"`
}

type UpdateRunStatusRequest struct {
	Status string `json:"status"`
}

type RunStageListResponse struct {
	Stages []StageRunResponse `json:"stages"`
}

type RunArtifactListResponse struct {
	Artifacts []ArtifactResponse `json:"artifacts"`
}

type RunCheckpointListResponse struct {
	Checkpoints []CheckpointResponse `json:"checkpoints"`
}

type RunStatusResponse struct {
	ID              string                  `json:"id"`
	Status          model.PipelineRunStatus `json:"status"`
	CurrentStageKey string                  `json:"currentStageKey"`
	StartedAt       *time.Time              `json:"startedAt,omitempty"`
	FinishedAt      *time.Time              `json:"finishedAt,omitempty"`
	UpdatedAt       time.Time               `json:"updatedAt"`
}

type PipelineTemplateResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PipelineRunResponse struct {
	ID              string                  `json:"id"`
	TemplateID      string                  `json:"templateId"`
	Title           string                  `json:"title"`
	RequirementText string                  `json:"requirementText"`
	SourceSessionID string                  `json:"sourceSessionId,omitempty"`
	TargetRepo      string                  `json:"targetRepo"`
	TargetBranch    string                  `json:"targetBranch"`
	WorkBranch      string                  `json:"workBranch"`
	Status          model.PipelineRunStatus `json:"status"`
	CurrentStageKey string                  `json:"currentStageKey"`
	CreatedBy       string                  `json:"createdBy"`
	StartedAt       *time.Time              `json:"startedAt,omitempty"`
	FinishedAt      *time.Time              `json:"finishedAt,omitempty"`
	CreatedAt       time.Time               `json:"createdAt"`
	UpdatedAt       time.Time               `json:"updatedAt"`
}

type StageRunResponse struct {
	ID            string               `json:"id"`
	PipelineRunID string               `json:"pipelineRunId"`
	StageKey      string               `json:"stageKey"`
	StageType     model.StageType      `json:"stageType"`
	Status        model.StageRunStatus `json:"status"`
	Attempt       int                  `json:"attempt"`
	InputJSON     string               `json:"inputJson,omitempty"`
	OutputJSON    string               `json:"outputJson,omitempty"`
	ErrorMessage  string               `json:"errorMessage,omitempty"`
	StartedAt     *time.Time           `json:"startedAt,omitempty"`
	FinishedAt    *time.Time           `json:"finishedAt,omitempty"`
	CreatedAt     time.Time            `json:"createdAt"`
	UpdatedAt     time.Time            `json:"updatedAt"`
}

type ArtifactResponse struct {
	ID            string             `json:"id"`
	PipelineRunID string             `json:"pipelineRunId"`
	StageRunID    string             `json:"stageRunId,omitempty"`
	ArtifactType  model.ArtifactType `json:"artifactType"`
	Title         string             `json:"title"`
	ContentText   string             `json:"contentText,omitempty"`
	ContentJSON   string             `json:"contentJson,omitempty"`
	FilePath      string             `json:"filePath,omitempty"`
	MetaJSON      string             `json:"metaJson,omitempty"`
	CreatedAt     time.Time          `json:"createdAt"`
}

type CheckpointResponse struct {
	ID             string                 `json:"id"`
	PipelineRunID  string                 `json:"pipelineRunId"`
	StageRunID     string                 `json:"stageRunId,omitempty"`
	CheckpointType model.CheckpointType   `json:"checkpointType"`
	Status         model.CheckpointStatus `json:"status"`
	ApproverID     string                 `json:"approverId,omitempty"`
	Decision       string                 `json:"decision,omitempty"`
	Comment        string                 `json:"comment,omitempty"`
	DecidedAt      *time.Time             `json:"decidedAt,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

type PipelineRunDetailResponse struct {
	Run         PipelineRunResponse  `json:"run"`
	Stages      []StageRunResponse   `json:"stages"`
	Artifacts   []ArtifactResponse   `json:"artifacts"`
	Checkpoints []CheckpointResponse `json:"checkpoints"`
}

func NewPipelineTemplateResponse(item model.PipelineTemplate) PipelineTemplateResponse {
	return PipelineTemplateResponse{ID: item.ID, Name: item.Name, Description: item.Description, Version: item.Version, IsActive: item.IsActive, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func NewPipelineRunResponse(item model.PipelineRun) PipelineRunResponse {
	return PipelineRunResponse{ID: item.ID, TemplateID: item.TemplateID, Title: item.Title, RequirementText: item.RequirementText, SourceSessionID: item.SourceSessionID, TargetRepo: item.TargetRepo, TargetBranch: item.TargetBranch, WorkBranch: item.WorkBranch, Status: item.Status, CurrentStageKey: item.CurrentStageKey, CreatedBy: item.CreatedBy, StartedAt: item.StartedAt, FinishedAt: item.FinishedAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func NewStageRunResponse(item model.StageRun) StageRunResponse {
	return StageRunResponse{ID: item.ID, PipelineRunID: item.PipelineRunID, StageKey: item.StageKey, StageType: item.StageType, Status: item.Status, Attempt: item.Attempt, InputJSON: item.InputJSON, OutputJSON: item.OutputJSON, ErrorMessage: item.ErrorMessage, StartedAt: item.StartedAt, FinishedAt: item.FinishedAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func NewArtifactResponse(item model.Artifact) ArtifactResponse {
	return ArtifactResponse{ID: item.ID, PipelineRunID: item.PipelineRunID, StageRunID: item.StageRunID, ArtifactType: item.ArtifactType, Title: item.Title, ContentText: item.ContentText, ContentJSON: item.ContentJSON, FilePath: item.FilePath, MetaJSON: item.MetaJSON, CreatedAt: item.CreatedAt}
}

func NewCheckpointResponse(item model.Checkpoint) CheckpointResponse {
	return CheckpointResponse{ID: item.ID, PipelineRunID: item.PipelineRunID, StageRunID: item.StageRunID, CheckpointType: item.CheckpointType, Status: item.Status, ApproverID: item.ApproverID, Decision: item.Decision, Comment: item.Comment, DecidedAt: item.DecidedAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
