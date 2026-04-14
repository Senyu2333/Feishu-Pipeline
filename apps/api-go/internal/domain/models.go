package domain

import "time"

type Role string

const (
	RoleProduct  Role = "product"
	RoleFrontend Role = "frontend"
	RoleBackend  Role = "backend"
	RoleAdmin    Role = "admin"
)

type SessionStatus string

const (
	SessionDraft      SessionStatus = "draft"
	SessionPublished  SessionStatus = "published"
	SessionInDelivery SessionStatus = "in_delivery"
	SessionTesting    SessionStatus = "testing"
	SessionDone       SessionStatus = "done"
	SessionArchived   SessionStatus = "archived"
)

type MessageRole string

const (
	MessageUser      MessageRole = "user"
	MessageAssistant MessageRole = "assistant"
	MessageSystem    MessageRole = "system"
)

type TaskType string

const (
	TaskFrontend TaskType = "frontend"
	TaskBackend  TaskType = "backend"
	TaskShared   TaskType = "shared"
)

type TaskStatus string

const (
	TaskTodo       TaskStatus = "todo"
	TaskInProgress TaskStatus = "in_progress"
	TaskTesting    TaskStatus = "testing"
	TaskDone       TaskStatus = "done"
)

type User struct {
	ID           string   `json:"id"`
	FeishuOpenID string   `json:"feishuOpenID,omitempty"`
	Name         string   `json:"name"`
	Email        string   `json:"email,omitempty"`
	Role         Role     `json:"role"`
	Departments  []string `json:"departments"`
}

type Session struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	Summary      string        `json:"summary"`
	Status       SessionStatus `json:"status"`
	OwnerID      string        `json:"ownerID"`
	OwnerName    string        `json:"ownerName"`
	MessageCount int           `json:"messageCount"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

type Message struct {
	ID        string      `json:"id"`
	SessionID string      `json:"sessionId"`
	Role      MessageRole `json:"role"`
	Content   string      `json:"content"`
	CreatedAt time.Time   `json:"createdAt"`
}

type Requirement struct {
	ID                  string        `json:"id"`
	SessionID           string        `json:"sessionId"`
	Title               string        `json:"title"`
	Summary             string        `json:"summary"`
	Status              SessionStatus `json:"status"`
	DeliverySummary     string        `json:"deliverySummary"`
	ReferencedKnowledge []string      `json:"referencedKnowledge"`
	PublishedAt         time.Time     `json:"publishedAt"`
}

type Task struct {
	ID                 string     `json:"id"`
	SessionID          string     `json:"sessionId"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Type               TaskType   `json:"type"`
	Status             TaskStatus `json:"status"`
	AssigneeName       string     `json:"assigneeName"`
	AssigneeRole       Role       `json:"assigneeRole"`
	AcceptanceCriteria []string   `json:"acceptanceCriteria"`
	Risks              []string   `json:"risks"`
	DocURL             string     `json:"docURL,omitempty"`
	BitableRecordURL   string     `json:"bitableRecordURL,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type SessionDetail struct {
	Session     Session      `json:"session"`
	Messages    []Message    `json:"messages"`
	Requirement *Requirement `json:"requirement,omitempty"`
	Tasks       []Task       `json:"tasks,omitempty"`
}

type RoleMapping struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Keyword     string   `json:"keyword"`
	Role        Role     `json:"role"`
	Departments []string `json:"departments"`
}

type KnowledgeSource struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DeliveryRecord struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"taskId"`
	Channel    string    `json:"channel"`
	Receiver   string    `json:"receiver"`
	Status     string    `json:"status"`
	RemoteID   string    `json:"remoteId"`
	RawPayload string    `json:"rawPayload"`
	CreatedAt  time.Time `json:"createdAt"`
}
