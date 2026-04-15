package controller

import (
	"net/http"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/service"
	admintype "feishu-pipeline/apps/api-go/internal/type/admin"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	adminService *service.AdminService
}

func NewAdminController(adminService *service.AdminService) *AdminController {
	return &AdminController{adminService: adminService}
}

func (c *AdminController) CreateRoleMapping(ctx *gin.Context) {
	var request admintype.CreateRoleMappingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := c.adminService.SaveRoleMapping(ctx.Request.Context(), &model.RoleMapping{
		Name:        request.Name,
		Keyword:     request.Keyword,
		Role:        request.Role,
		Departments: request.Departments,
	})
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusCreated, admintype.CreateRoleMappingResponse{Status: "saved"})
}

func (c *AdminController) SyncKnowledge(ctx *gin.Context) {
	var request admintype.SyncKnowledgeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	items := admintype.NewKnowledgeSourceModels(request.Sources)
	if err := c.adminService.SyncKnowledgeSources(ctx.Request.Context(), items); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusCreated, admintype.SyncKnowledgeResponse{Count: len(items)})
}
