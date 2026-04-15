package service

import (
	"context"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/repo"
)

type AdminService struct {
	repository *repo.Repository
}

func NewAdminService(repository *repo.Repository) *AdminService {
	return &AdminService{repository: repository}
}

func (s *AdminService) SaveRoleMapping(ctx context.Context, mapping *model.RoleMapping) error {
	return s.repository.SaveRoleMapping(ctx, mapping)
}

func (s *AdminService) SaveRoleOwner(ctx context.Context, owner *model.RoleOwner) error {
	return s.repository.SaveRoleOwner(ctx, owner)
}

func (s *AdminService) ListRoleOwners(ctx context.Context) ([]model.RoleOwner, error) {
	return s.repository.ListRoleOwners(ctx)
}

func (s *AdminService) SyncKnowledgeSources(ctx context.Context, items []model.KnowledgeSource) error {
	return s.repository.SaveKnowledgeSources(ctx, items)
}
