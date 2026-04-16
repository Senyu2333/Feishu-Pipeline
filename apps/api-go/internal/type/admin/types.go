package admintype

import (
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
)

type CreateRoleMappingRequest struct {
	Name        string     `json:"name" binding:"required"`
	Keyword     string     `json:"keyword" binding:"required"`
	Role        model.Role `json:"role" binding:"required"`
	Departments []string   `json:"departments"`
}

type SaveRoleOwnerRequest struct {
	Role         model.Role `json:"role" binding:"required"`
	OwnerName    string     `json:"ownerName" binding:"required"`
	FeishuID     string     `json:"feishuId"`
	FeishuIDType string     `json:"feishuIdType"`
	Enabled      bool       `json:"enabled"`
}

type SyncKnowledgeRequest struct {
	Sources []KnowledgeSourceInput `json:"sources" binding:"required"`
}

type KnowledgeSourceInput struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type CreateRoleMappingResponse struct {
	Status string `json:"status"`
}

type RoleOwnerResponse struct {
	ID           string     `json:"id"`
	Role         model.Role `json:"role"`
	OwnerName    string     `json:"ownerName"`
	FeishuID     string     `json:"feishuId,omitempty"`
	FeishuIDType string     `json:"feishuIdType,omitempty"`
	Enabled      bool       `json:"enabled"`
}

type SyncKnowledgeResponse struct {
	Count int `json:"count"`
}

func NewRoleOwnerResponse(owner model.RoleOwner) RoleOwnerResponse {
	return RoleOwnerResponse{
		ID:           owner.ID,
		Role:         owner.Role,
		OwnerName:    owner.OwnerName,
		FeishuID:     owner.FeishuID,
		FeishuIDType: owner.FeishuIDType,
		Enabled:      owner.Enabled,
	}
}

func NewKnowledgeSourceModels(inputs []KnowledgeSourceInput) []model.KnowledgeSource {
	items := make([]model.KnowledgeSource, 0, len(inputs))
	for _, input := range inputs {
		items = append(items, model.KnowledgeSource{
			Title:     input.Title,
			Content:   input.Content,
			UpdatedAt: time.Now().UTC(),
		})
	}
	return items
}
