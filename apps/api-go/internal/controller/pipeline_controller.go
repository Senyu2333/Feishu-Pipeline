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

// ListTemplates
// @tags Pipeline
// @summary 获取流水线模板列表
// @router /api/pipeline-templates [GET]
// @produce application/json
// @success 200 {object} pipelinetype.PipelineTemplateListEnvelope
// @failure 500 {object} pipelinetype.ErrorEnvelope
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

// ListRuns
// @tags Pipeline
// @summary 获取流水线运行列表
// @router /api/pipeline-runs [GET]
// @produce application/json
// @success 200 {object} pipelinetype.PipelineRunListEnvelope
// @failure 500 {object} pipelinetype.ErrorEnvelope
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

// CreateRun
// @tags Pipeline
// @summary 创建流水线运行
// @router /api/pipeline-runs [POST]
// @accept application/json
// @produce application/json
// @param req body pipelinetype.CreatePipelineRunRequest true "json入参"
// @success 201 {object} pipelinetype.PipelineRunDetailEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
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

// CreateRunFromSession
// @tags Pipeline
// @summary 从会话创建流水线运行
// @router /api/pipeline-runs/from-session [POST]
// @accept application/json
// @produce application/json
// @param req body pipelinetype.CreatePipelineRunFromSessionRequest true "json入参"
// @success 201 {object} pipelinetype.PipelineRunDetailEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
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

// GetRun
// @tags Pipeline
// @summary 获取流水线运行详情
// @router /api/pipeline-runs/{id} [GET]
// @produce application/json
// @param id path string true "流水线运行ID"
// @success 200 {object} pipelinetype.PipelineRunDetailEnvelope
// @failure 404 {object} pipelinetype.ErrorEnvelope
func (c *PipelineController) GetRun(ctx *gin.Context) {
	item, err := c.pipelineService.GetPipelineRunDetail(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusNotFound, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapPipelineRunDetail(item))
}

// ListStages
// @tags Pipeline
// @summary 获取流水线阶段列表
// @router /api/pipeline-runs/{id}/stages [GET]
// @produce application/json
// @param id path string true "流水线运行ID"
// @success 200 {object} pipelinetype.RunStageListEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
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

// ListArtifacts
// @tags Pipeline
// @summary 获取流水线产物列表
// @router /api/pipeline-runs/{id}/artifacts [GET]
// @produce application/json
// @param id path string true "流水线运行ID"
// @success 200 {object} pipelinetype.RunArtifactListEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
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

// ListCheckpoints
// @tags Pipeline
// @summary 获取流水线检查点列表
// @router /api/pipeline-runs/{id}/checkpoints [GET]
// @produce application/json
// @param id path string true "流水线运行ID"
// @success 200 {object} pipelinetype.RunCheckpointListEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
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

// ListAgentRuns
// @tags Pipeline
// @summary 获取流水线 Agent 执行记录
// @router /api/pipeline-runs/{id}/agent-runs [GET]
// @produce application/json
// @param id path string true "流水线运行ID"
// @success 200 {object} pipelinetype.RunAgentRunListEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
func (c *PipelineController) ListAgentRuns(ctx *gin.Context) {
	items, err := c.pipelineService.ListAgentRuns(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	response := make([]pipelinetype.AgentRunResponse, 0, len(items))
	for _, item := range items {
		response = append(response, pipelinetype.NewAgentRunResponse(item))
	}
	writeSuccess(ctx, http.StatusOK, pipelinetype.RunAgentRunListResponse{AgentRuns: response})
}

// StartRun
// @tags Pipeline
// @summary 启动流水线运行
// @router /api/pipeline-runs/{id}/start [POST]
// @produce application/json
// @param id path string true "流水线运行ID"
// @success 200 {object} pipelinetype.RunStatusEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
func (c *PipelineController) StartRun(ctx *gin.Context) {
	item, err := c.pipelineService.StartPipelineRun(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapRunStatus(item))
}

// PauseRun
// @tags Pipeline
// @summary 暂停流水线运行
// @router /api/pipeline-runs/{id}/pause [POST]
// @produce application/json
// @param id path string true "流水线运行ID"
// @success 200 {object} pipelinetype.RunStatusEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
func (c *PipelineController) PauseRun(ctx *gin.Context) {
	item, err := c.pipelineService.PausePipelineRun(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapRunStatus(item))
}

// ResumeRun
// @tags Pipeline
// @summary 恢复流水线运行
// @router /api/pipeline-runs/{id}/resume [POST]
// @produce application/json
// @param id path string true "流水线运行ID"
// @success 200 {object} pipelinetype.RunStatusEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
func (c *PipelineController) ResumeRun(ctx *gin.Context) {
	item, err := c.pipelineService.ResumePipelineRun(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapRunStatus(item))
}

// TerminateRun
// @tags Pipeline
// @summary 终止流水线运行
// @router /api/pipeline-runs/{id}/terminate [POST]
// @produce application/json
// @param id path string true "流水线运行ID"
// @success 200 {object} pipelinetype.RunStatusEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
func (c *PipelineController) TerminateRun(ctx *gin.Context) {
	item, err := c.pipelineService.TerminatePipelineRun(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapRunStatus(item))
}

// ApproveCheckpoint
// @tags Checkpoint
// @summary 审批通过检查点
// @router /api/checkpoints/{checkpointID}/approve [POST]
// @accept application/json
// @produce application/json
// @param checkpointID path string true "检查点ID"
// @param req body pipelinetype.UpdateCheckpointDecisionRequest true "json入参"
// @success 200 {object} pipelinetype.CheckpointEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
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

// RejectCheckpoint
// @tags Checkpoint
// @summary 驳回检查点
// @router /api/checkpoints/{checkpointID}/reject [POST]
// @accept application/json
// @produce application/json
// @param checkpointID path string true "检查点ID"
// @param req body pipelinetype.UpdateCheckpointDecisionRequest true "json入参"
// @success 200 {object} pipelinetype.CheckpointEnvelope
// @failure 400 {object} pipelinetype.ErrorEnvelope
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
