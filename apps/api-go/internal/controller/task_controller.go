package controller

import (
	"net/http"

	"feishu-pipeline/apps/api-go/internal/service"
	tasktype "feishu-pipeline/apps/api-go/internal/type/task"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	taskService *service.TaskService
}

func NewTaskController(taskService *service.TaskService) *TaskController {
	return &TaskController{taskService: taskService}
}

// GetTask
// @tags 任务
// @summary 任务详情
// @router /api/tasks/{taskID} [GET]
// @param taskID path string true "任务ID"
// @produce application/json
func (c *TaskController) GetTask(ctx *gin.Context) {
	task, err := c.taskService.GetTask(ctx.Request.Context(), ctx.Param("taskID"))
	if err != nil {
		writeError(ctx, http.StatusNotFound, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, tasktype.NewTaskResponse(task))
}

// UpdateTaskStatus
// @tags 任务
// @summary 更新任务状态
// @router /api/tasks/{taskID}/status [PATCH]
// @accept application/json
// @produce application/json
// @param taskID path string true "任务ID"
// @param req body tasktype.UpdateTaskStatusRequest true "json入参"
func (c *TaskController) UpdateTaskStatus(ctx *gin.Context) {
	var request tasktype.UpdateTaskStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	task, err := c.taskService.UpdateTaskStatus(ctx.Request.Context(), ctx.Param("taskID"), request.Status)
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, tasktype.NewTaskResponse(task))
}
