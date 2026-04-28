package controller

import (
	"errors"
	"net/http"
	"os"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/repo"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OpenAPIController struct {
	repo *repo.Repository
}

func NewOpenAPIController(repo *repo.Repository) *OpenAPIController {
	return &OpenAPIController{repo: repo}
}

// SaveSpec
// @tags OpenAPI
// @summary 保存 OpenAPI 规范
// @router /api/openapi/specs [POST]
// @produce application/json
// @accept application/json
// @param body body SaveSpecRequest true "OpenAPI 规范请求"
func (c *OpenAPIController) SaveSpec(ctx *gin.Context) {
	var req SaveSpecRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	specID := "spec_" + uuid.New().String()[:8]
	
	// 获取前端 URL
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	swaggerURL := frontendURL + "/swagger?specId=" + specID

	spec := &model.OpenAPISpec{
		ID:         specID,
		Title:      req.Title,
		SpecJSON:   req.SpecJSON,
		SwaggerURL: swaggerURL,
		BaseModel: model.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if err := c.repo.CreateOpenAPISpec(ctx.Request.Context(), spec); err != nil {
		writeError(ctx, http.StatusInternalServerError, errors.New("failed to save spec"))
		return
	}

	writeSuccess(ctx, http.StatusOK, gin.H{
		"specId":     specID,
		"swaggerUrl": swaggerURL,
	})
}

// GetSpec
// @tags OpenAPI
// @summary 获取 OpenAPI 规范
// @router /api/openapi/specs/:specId [GET]
// @produce json
// @param specId path string true "Spec ID"
func (c *OpenAPIController) GetSpec(ctx *gin.Context) {
	specID := ctx.Param("specId")

	spec, err := c.repo.GetOpenAPISpec(ctx.Request.Context(), specID)
	if err != nil {
		writeError(ctx, http.StatusNotFound, errors.New("spec not found"))
		return
	}

	writeSuccess(ctx, http.StatusOK, spec.SpecJSON)
}

// SaveSpecPublic 不需要认证的保存接口（供 AI 工具调用）
func (c *OpenAPIController) SaveSpecPublic(ctx *gin.Context) {
	var req SaveSpecRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	specID := "spec_" + uuid.New().String()[:8]

	// 获取前端 URL
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	swaggerURL := frontendURL + "/swagger?specId=" + specID

	spec := &model.OpenAPISpec{
		ID:         specID,
		Title:      req.Title,
		SpecJSON:   req.SpecJSON,
		SwaggerURL: swaggerURL,
		BaseModel: model.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if err := c.repo.CreateOpenAPISpec(ctx.Request.Context(), spec); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to save spec"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"specId":     specID,
			"swaggerUrl": swaggerURL,
		},
	})
}

// GetSpecPublic 不需要认证的获取接口（供 Swagger UI 调用）
func (c *OpenAPIController) GetSpecPublic(ctx *gin.Context) {
	specID := ctx.Param("specId")

	spec, err := c.repo.GetOpenAPISpec(ctx.Request.Context(), specID)
	if err != nil {
		writeError(ctx, http.StatusNotFound, errors.New("spec not found"))
		return
	}

	// 直接返回 JSON，不包装
	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, spec.SpecJSON)
}

type SaveSpecRequest struct {
	Title    string `json:"title"`
	SpecJSON string `json:"spec_json" binding:"required"`
}
