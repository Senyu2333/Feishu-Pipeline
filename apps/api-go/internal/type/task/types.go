package tasktype

import (
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
)

type UpdateTaskStatusRequest struct {
	Status model.TaskStatus `json:"status" binding:"required"`
}

type TaskResponse struct {
	ID                 string           `json:"id"`
	SessionID          string           `json:"sessionId"`
	Title              string           `json:"title"`
	Description        string           `json:"description"`
	Type               model.TaskType   `json:"type"`
	Status             model.TaskStatus `json:"status"`
	AssigneeName       string           `json:"assigneeName"`
	AssigneeRole       model.Role       `json:"assigneeRole"`
	DocURL             string           `json:"docURL,omitempty"`
	BitableRecordURL   string           `json:"bitableRecordURL,omitempty"`
	AcceptanceCriteria []string         `json:"acceptanceCriteria"`
	Risks              []string         `json:"risks"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
}

func NewTaskResponse(task model.Task) TaskResponse {
	return TaskResponse{
		ID:                 task.ID,
		SessionID:          task.SessionID,
		Title:              task.Title,
		Description:        task.Description,
		Type:               task.Type,
		Status:             task.Status,
		AssigneeName:       task.AssigneeName,
		AssigneeRole:       task.AssigneeRole,
		DocURL:             task.DocURL,
		BitableRecordURL:   task.BitableRecordURL,
		AcceptanceCriteria: task.AcceptanceCriteria,
		Risks:              task.Risks,
		CreatedAt:          task.CreatedAt,
		UpdatedAt:          task.UpdatedAt,
	}
}
