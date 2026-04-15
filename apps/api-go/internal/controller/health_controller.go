package controller

import (
	"net/http"

	"feishu-pipeline/apps/api-go/internal/service"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
	healthService *service.HealthService
}

func NewHealthController(healthService *service.HealthService) *HealthController {
	return &HealthController{healthService: healthService}
}

func (c *HealthController) Health(ctx *gin.Context) {
	writeSuccess(ctx, http.StatusOK, c.healthService.Health())
}
