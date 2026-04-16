package tasktype

import (
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
)

type UpdateTaskStatusRequest struct {
	Status model.TaskStatus `json:"status" binding:"required"`
}

type TaskResponse struct {
	ID                 string             `json:"id"`
	SessionID          string             `json:"sessionId"`
	Title              string             `json:"title"`
	Description        string             `json:"description"`
	Type               model.TaskType     `json:"type"`
	Status             model.TaskStatus   `json:"status"`
	Priority           model.TaskPriority `json:"priority"`
	EstimateDays       int                `json:"estimateDays"`
	AssigneeName       string             `json:"assigneeName"`
	AssigneeRole       model.Role         `json:"assigneeRole"`
	AssigneeID         string             `json:"assigneeId,omitempty"`
	AssigneeIDType     string             `json:"assigneeIdType,omitempty"`
	PlannedStartAt     *time.Time         `json:"plannedStartAt,omitempty"`
	PlannedEndAt       *time.Time         `json:"plannedEndAt,omitempty"`
	NotifyContent      string             `json:"notifyContent,omitempty"`
	DocURL             string             `json:"docURL,omitempty"`
	BitableRecordURL   string             `json:"bitableRecordURL,omitempty"`
	AcceptanceCriteria []string           `json:"acceptanceCriteria"`
	Risks              []string           `json:"risks"`
	CreatedAt          time.Time          `json:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt"`
}

func NewTaskResponse(task model.Task) TaskResponse {
	return TaskResponse{
		ID:                 task.ID,
		SessionID:          task.SessionID,
		Title:              task.Title,
		Description:        task.Description,
		Type:               task.Type,
		Status:             task.Status,
		Priority:           task.Priority,
		EstimateDays:       task.EstimateDays,
		AssigneeName:       task.AssigneeName,
		AssigneeRole:       task.AssigneeRole,
		AssigneeID:         task.AssigneeID,
		AssigneeIDType:     task.AssigneeIDType,
		PlannedStartAt:     task.PlannedStartAt,
		PlannedEndAt:       task.PlannedEndAt,
		NotifyContent:      task.NotifyContent,
		DocURL:             task.DocURL,
		BitableRecordURL:   task.BitableRecordURL,
		AcceptanceCriteria: task.AcceptanceCriteria,
		Risks:              task.Risks,
		CreatedAt:          task.CreatedAt,
		UpdatedAt:          task.UpdatedAt,
	}
}
