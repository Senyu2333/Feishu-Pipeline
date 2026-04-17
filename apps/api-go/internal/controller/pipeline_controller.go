package controller

import (
	"net/http"

	"feishu-pipeline/apps/api-go/internal/service"

	"github.com/gin-gonic/gin"
)

type PipelineController struct {
	pipelineService *service.PipelineService
}

func NewPipelineController(pipelineService *service.PipelineService) *PipelineController {
	return &PipelineController{pipelineService: pipelineService}
}

func (c *PipelineController) Create(ctx *gin.Context) {
	var req service.PipelineResult
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := c.pipelineService.CreatePipeline(ctx.Request.Context(), req)
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	writeSuccess(ctx, http.StatusOK, result)
}
