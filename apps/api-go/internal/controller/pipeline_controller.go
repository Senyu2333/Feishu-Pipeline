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
	var req struct {
		SessionID string `json:"sessionId" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	writeSuccess(ctx, statusOK, map[string]string{"message": "use publish workflow to create bitable automatically", "sessionId": req.SessionID})
}
