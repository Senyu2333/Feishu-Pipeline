package sessiontype

import (
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	tasktype "feishu-pipeline/apps/api-go/internal/type/task"
)

type CreateSessionRequest struct {
	Title  string `json:"title" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
}

type CreateMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

type AutoPublishCheckRequest struct {
	Content string `json:"content" binding:"required"`
}

type AutoPublishCheckResponse struct {
	Triggered bool   `json:"triggered"`
	Reason    string `json:"reason"`
}

type SessionSummaryResponse struct {
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	Summary      string              `json:"summary"`
	Status       model.SessionStatus `json:"status"`
	OwnerName    string              `json:"ownerName"`
	MessageCount int                 `json:"messageCount"`
	UpdatedAt    time.Time           `json:"updatedAt"`
}

type MessageResponse struct {
	ID        string            `json:"id"`
	SessionID string            `json:"sessionId"`
	Role      model.MessageRole `json:"role"`
	Content   string            `json:"content"`
	CreatedAt time.Time         `json:"createdAt"`
}

type RequirementResponse struct {
	SessionID           string              `json:"sessionId"`
	RequirementID       string              `json:"requirementId,omitempty"`
	Title               string              `json:"title"`
	Summary             string              `json:"summary"`
	Status              model.SessionStatus `json:"status"`
	PublishedAt         *time.Time          `json:"publishedAt,omitempty"`
	DeliverySummary     string              `json:"deliverySummary,omitempty"`
	ReferencedKnowledge []string            `json:"referencedKnowledge"`
}

type SessionDetailResponse struct {
	Session     SessionSummaryResponse  `json:"session"`
	Messages    []MessageResponse       `json:"messages"`
	Requirement *RequirementResponse    `json:"requirement,omitempty"`
	Tasks       []tasktype.TaskResponse `json:"tasks"`
}

func NewSessionSummaryResponse(session model.Session, ownerName string, messageCount int) SessionSummaryResponse {
	return SessionSummaryResponse{
		ID:           session.ID,
		Title:        session.Title,
		Summary:      session.Summary,
		Status:       session.Status,
		OwnerName:    ownerName,
		MessageCount: messageCount,
		UpdatedAt:    session.UpdatedAt,
	}
}

func NewMessageResponse(message model.Message) MessageResponse {
	return MessageResponse{
		ID:        message.ID,
		SessionID: message.SessionID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}
}

func NewRequirementResponse(requirement *model.Requirement) *RequirementResponse {
	if requirement == nil {
		return nil
	}
	publishedAt := requirement.PublishedAt
	return &RequirementResponse{
		SessionID:           requirement.SessionID,
		RequirementID:       requirement.ID,
		Title:               requirement.Title,
		Summary:             requirement.Summary,
		Status:              requirement.Status,
		PublishedAt:         &publishedAt,
		DeliverySummary:     requirement.DeliverySummary,
		ReferencedKnowledge: requirement.ReferencedKnowledge,
	}
}
