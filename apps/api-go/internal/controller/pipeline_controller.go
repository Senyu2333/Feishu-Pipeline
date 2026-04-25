package controller

import (
	"net/http"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/service"
	pipelinetype "feishu-pipeline/apps/api-go/internal/type/pipeline"

	"github.com/gin-gonic/gin"
)

type PipelineController struct {
	pipelineService *service.PipelineService
}

func NewPipelineController(pipelineService *service.PipelineService) *PipelineController {
	return &PipelineController{pipelineService: pipelineService}
}

func (c *PipelineController) ListTemplates(ctx *gin.Context) {
	items, err := c.pipelineService.ListPipelineTemplates(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	response := make([]pipelinetype.PipelineTemplateResponse, 0, len(items))
	for _, item := range items {
		response = append(response, pipelinetype.NewPipelineTemplateResponse(item))
	}
	writeSuccess(ctx, http.StatusOK, response)
}

func (c *PipelineController) ListRuns(ctx *gin.Context) {
	items, err := c.pipelineService.ListPipelineRuns(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	response := make([]pipelinetype.PipelineRunResponse, 0, len(items))
	for _, item := range items {
		response = append(response, pipelinetype.NewPipelineRunResponse(item))
	}
	writeSuccess(ctx, http.StatusOK, response)
}

func (c *PipelineController) CreateRun(ctx *gin.Context) {
	var req pipelinetype.CreatePipelineRunRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	item, err := c.pipelineService.CreatePipelineRun(ctx.Request.Context(), service.CreatePipelineRunInput{TemplateID: req.TemplateID, Title: req.Title, RequirementText: req.RequirementText, TargetRepo: req.TargetRepo, TargetBranch: req.TargetBranch, CreatedBy: currentUserID(ctx)})
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusCreated, mapPipelineRunDetail(item))
}

func (c *PipelineController) CreateRunFromSession(ctx *gin.Context) {
	var req pipelinetype.CreatePipelineRunFromSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	item, err := c.pipelineService.CreatePipelineRunFromSession(ctx.Request.Context(), req.SessionID, req.TemplateID, req.TargetRepo, req.TargetBranch, currentUserID(ctx))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusCreated, mapPipelineRunDetail(item))
}

func (c *PipelineController) GetRun(ctx *gin.Context) {
	item, err := c.pipelineService.GetPipelineRunDetail(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusNotFound, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapPipelineRunDetail(item))
}

func (c *PipelineController) ListStages(ctx *gin.Context) {
	items, err := c.pipelineService.ListStageRuns(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	response := make([]pipelinetype.StageRunResponse, 0, len(items))
	for _, item := range items {
		response = append(response, pipelinetype.NewStageRunResponse(item))
	}
	writeSuccess(ctx, http.StatusOK, pipelinetype.RunStageListResponse{Stages: response})
}

func (c *PipelineController) ListArtifacts(ctx *gin.Context) {
	items, err := c.pipelineService.ListArtifacts(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	response := make([]pipelinetype.ArtifactResponse, 0, len(items))
	for _, item := range items {
		response = append(response, pipelinetype.NewArtifactResponse(item))
	}
	writeSuccess(ctx, http.StatusOK, pipelinetype.RunArtifactListResponse{Artifacts: response})
}

func (c *PipelineController) ListCheckpoints(ctx *gin.Context) {
	items, err := c.pipelineService.ListCheckpoints(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	response := make([]pipelinetype.CheckpointResponse, 0, len(items))
	for _, item := range items {
		response = append(response, pipelinetype.NewCheckpointResponse(item))
	}
	writeSuccess(ctx, http.StatusOK, pipelinetype.RunCheckpointListResponse{Checkpoints: response})
}

func (c *PipelineController) StartRun(ctx *gin.Context) {
	item, err := c.pipelineService.StartPipelineRun(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapRunStatus(item))
}

func (c *PipelineController) PauseRun(ctx *gin.Context) {
	item, err := c.pipelineService.PausePipelineRun(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapRunStatus(item))
}

func (c *PipelineController) ResumeRun(ctx *gin.Context) {
	item, err := c.pipelineService.ResumePipelineRun(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapRunStatus(item))
}

func (c *PipelineController) TerminateRun(ctx *gin.Context) {
	item, err := c.pipelineService.TerminatePipelineRun(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapRunStatus(item))
}

func (c *PipelineController) ApproveCheckpoint(ctx *gin.Context) {
	var req pipelinetype.UpdateCheckpointDecisionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	item, err := c.pipelineService.ApproveCheckpoint(ctx.Request.Context(), ctx.Param("checkpointID"), req.Comment, currentUserID(ctx))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, pipelinetype.NewCheckpointResponse(item))
}

func (c *PipelineController) RejectCheckpoint(ctx *gin.Context) {
	var req pipelinetype.UpdateCheckpointDecisionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	item, err := c.pipelineService.RejectCheckpoint(ctx.Request.Context(), ctx.Param("checkpointID"), req.Comment, currentUserID(ctx))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, pipelinetype.NewCheckpointResponse(item))
}

func mapRunStatus(item model.PipelineRun) pipelinetype.RunStatusResponse {
	return pipelinetype.RunStatusResponse{ID: item.ID, Status: item.Status, CurrentStageKey: item.CurrentStageKey, StartedAt: item.StartedAt, FinishedAt: item.FinishedAt, UpdatedAt: item.UpdatedAt}
}

func mapPipelineRunDetail(item *service.PipelineRunDetail) pipelinetype.PipelineRunDetailResponse {
	stages := make([]pipelinetype.StageRunResponse, 0, len(item.Stages))
	for _, stage := range item.Stages {
		stages = append(stages, pipelinetype.NewStageRunResponse(stage))
	}
	artifacts := make([]pipelinetype.ArtifactResponse, 0, len(item.Artifacts))
	for _, artifact := range item.Artifacts {
		artifacts = append(artifacts, pipelinetype.NewArtifactResponse(artifact))
	}
	checkpoints := make([]pipelinetype.CheckpointResponse, 0, len(item.Checkpoints))
	for _, checkpoint := range item.Checkpoints {
		checkpoints = append(checkpoints, pipelinetype.NewCheckpointResponse(checkpoint))
	}
	return pipelinetype.PipelineRunDetailResponse{Run: pipelinetype.NewPipelineRunResponse(item.Run), Stages: stages, Artifacts: artifacts, Checkpoints: checkpoints}
}
