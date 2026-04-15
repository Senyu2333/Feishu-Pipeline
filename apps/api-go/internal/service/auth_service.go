package service

import (
	"context"
	"errors"
	"strings"

	"feishu-pipeline/apps/api-go/internal/external/feishu"
	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/repo"

	"gorm.io/gorm"
)

type AuthService struct {
	repository   *repo.Repository
	feishuClient *feishu.Client
}

func NewAuthService(repository *repo.Repository, feishuClient *feishu.Client) *AuthService {
	return &AuthService{
		repository:   repository,
		feishuClient: feishuClient,
	}
}

func (s *AuthService) LoginURL(state string) string {
	return s.feishuClient.BuildLoginURL(state)
}

func (s *AuthService) LoginByCode(ctx context.Context, code string) (model.User, error) {
	user, err := s.feishuClient.ExchangeCode(ctx, code)
	if err != nil {
		return model.User{}, err
	}
	if err := s.repository.UpsertUser(ctx, &user); err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID string) (model.User, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		userID = "u_product_demo"
	}

	user, err := s.repository.FindUserByID(ctx, userID)
	if err == nil {
		return user, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) && userID != "u_product_demo" {
		return s.repository.FindUserByID(ctx, "u_product_demo")
	}
	return user, err
}

func (s *AuthService) EnsureUser(ctx context.Context, userID string) model.User {
	user, err := s.CurrentUser(ctx, userID)
	if err == nil {
		return user
	}
	return model.User{
		ID:          "u_product_demo",
		Name:        "产品经理小明",
		Email:       "product@example.com",
		Role:        model.RoleProduct,
		Departments: []string{"产品部"},
	}
}
