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

func (c *TaskController) GetTask(ctx *gin.Context) {
	task, err := c.taskService.GetTask(ctx.Request.Context(), ctx.Param("taskID"))
	if err != nil {
		writeError(ctx, http.StatusNotFound, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, tasktype.NewTaskResponse(task))
}

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
