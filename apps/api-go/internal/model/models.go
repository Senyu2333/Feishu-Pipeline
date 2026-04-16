package model

import "time"

type Role string

const (
	RoleProduct  Role = "product"
	RoleFrontend Role = "frontend"
	RoleBackend  Role = "backend"
	RoleOther    Role = "other"
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

type TaskPriority string

const (
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityLow    TaskPriority = "low"
)

type TaskStatus string

const (
	TaskTodo       TaskStatus = "todo"
	TaskInProgress TaskStatus = "in_progress"
	TaskTesting    TaskStatus = "testing"
	TaskDone       TaskStatus = "done"
)

type BaseModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID           string   `gorm:"primaryKey;size:64"`
	FeishuOpenID string   `gorm:"size:128;uniqueIndex"`
	Name         string   `gorm:"size:128;not null"`
	Email        string   `gorm:"size:128"`
	Role         Role     `gorm:"size:32;not null"`
	Departments  []string `gorm:"serializer:json"`
	BaseModel
}

type FeishuCredential struct {
	ID                    string    `gorm:"primaryKey;size:64"`
	UserID                string    `gorm:"size:64;not null;uniqueIndex"`
	User                  User      `gorm:"foreignKey:UserID;references:ID"`
	OpenID                string    `gorm:"size:128;not null;uniqueIndex"`
	UnionID               string    `gorm:"size:128"`
	FeishuUserID          string    `gorm:"size:128"`
	AccessToken           string    `gorm:"type:text;not null"`
	RefreshToken          string    `gorm:"type:text;not null"`
	AccessTokenExpiresAt  time.Time `gorm:"not null"`
	RefreshTokenExpiresAt time.Time `gorm:"not null"`
	LastLoginAt           time.Time `gorm:"not null"`
	LastRefreshAt         time.Time
	BaseModel
}

type LoginSession struct {
	ID        string    `gorm:"primaryKey;size:128"`
	UserID    string    `gorm:"size:64;not null;index"`
	User      User      `gorm:"foreignKey:UserID;references:ID"`
	ExpiresAt time.Time `gorm:"not null;index"`
	BaseModel
}

type Session struct {
	ID      string        `gorm:"primaryKey;size:64"`
	Title   string        `gorm:"size:255;not null"`
	Summary string        `gorm:"type:text;not null"`
	Status  SessionStatus `gorm:"size:32;not null;index"`
	OwnerID string        `gorm:"size:64;not null;index"`
	Owner   User          `gorm:"foreignKey:OwnerID;references:ID"`
	BaseModel
}

type Message struct {
	ID        string      `gorm:"primaryKey;size:64"`
	SessionID string      `gorm:"size:64;not null;index"`
	Role      MessageRole `gorm:"size:32;not null"`
	Content   string      `gorm:"type:text;not null"`
	CreatedAt time.Time   `gorm:"autoCreateTime"`
}

type Requirement struct {
	ID                  string        `gorm:"primaryKey;size:64"`
	SessionID           string        `gorm:"size:64;not null;uniqueIndex"`
	Title               string        `gorm:"size:255;not null"`
	Summary             string        `gorm:"type:text;not null"`
	Status              SessionStatus `gorm:"size:32;not null"`
	DeliverySummary     string        `gorm:"type:text;not null"`
	ReferencedKnowledge []string      `gorm:"serializer:json"`
	PublishedAt         time.Time
	BaseModel
}

type Task struct {
	ID                 string       `gorm:"primaryKey;size:64"`
	SessionID          string       `gorm:"size:64;not null;index"`
	Title              string       `gorm:"size:255;not null"`
	Description        string       `gorm:"type:text;not null"`
	Type               TaskType     `gorm:"size:32;not null"`
	Status             TaskStatus   `gorm:"size:32;not null"`
	AssigneeName       string       `gorm:"size:128;not null"`
	AssigneeRole       Role         `gorm:"size:32;not null"`
	AssigneeID         string       `gorm:"size:128"`
	AssigneeIDType     string       `gorm:"size:32"`
	Priority           TaskPriority `gorm:"size:32;not null;default:medium"`
	EstimateDays       int          `gorm:"not null;default:1"`
	PlannedStartAt     *time.Time
	PlannedEndAt       *time.Time
	NotifyContent      string   `gorm:"type:text"`
	AcceptanceCriteria []string `gorm:"serializer:json"`
	Risks              []string `gorm:"serializer:json"`
	DocURL             string   `gorm:"size:512"`
	BitableAppToken    string   `gorm:"size:128"`
	BitableTableID     string   `gorm:"size:128"`
	BitableRecordID    string   `gorm:"size:128"`
	BitableRecordURL   string   `gorm:"size:512"`
	BaseModel
}

type RoleMapping struct {
	ID          string   `gorm:"primaryKey;size:64"`
	Name        string   `gorm:"size:128;not null"`
	Keyword     string   `gorm:"size:128;not null"`
	Role        Role     `gorm:"size:32;not null"`
	Departments []string `gorm:"serializer:json"`
}

type RoleOwner struct {
	ID           string `gorm:"primaryKey;size:64"`
	Role         Role   `gorm:"size:32;not null;uniqueIndex"`
	OwnerName    string `gorm:"size:128;not null"`
	FeishuID     string `gorm:"size:128"`
	FeishuIDType string `gorm:"size:32"`
	Enabled      bool   `gorm:"not null;default:false"`
	BaseModel
}

type KnowledgeSource struct {
	ID        string `gorm:"primaryKey;size:64"`
	Title     string `gorm:"size:255;not null"`
	Content   string `gorm:"type:text;not null"`
	UpdatedAt time.Time
}

type MessageDelivery struct {
	ID         string    `gorm:"primaryKey;size:64"`
	TaskID     string    `gorm:"size:64;not null;index"`
	Channel    string    `gorm:"size:64;not null"`
	Receiver   string    `gorm:"size:128;not null"`
	Status     string    `gorm:"size:64;not null"`
	RemoteID   string    `gorm:"size:128"`
	RawPayload string    `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}
