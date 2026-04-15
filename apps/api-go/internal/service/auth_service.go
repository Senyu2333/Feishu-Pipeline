package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/external/feishu"
	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/repo"
	"feishu-pipeline/apps/api-go/internal/utils"

	"gorm.io/gorm"
)

var ErrAuthenticationRequired = errors.New("authentication required")

type AuthService struct {
	repository   *repo.Repository
	feishuClient *feishu.Client
	sessionTTL   time.Duration
}

func NewAuthService(repository *repo.Repository, feishuClient *feishu.Client, sessionTTL time.Duration) *AuthService {
	return &AuthService{
		repository:   repository,
		feishuClient: feishuClient,
		sessionTTL:   sessionTTL,
	}
}

func (s *AuthService) FeishuAppID() string {
	return s.feishuClient.AppID()
}

func (s *AuthService) FeishuEnabled() bool {
	return s.feishuClient.Enabled()
}

func (s *AuthService) LoginByCode(ctx context.Context, code string) (model.User, model.LoginSession, error) {
	token, err := s.feishuClient.ExchangeCodeForUserToken(ctx, code)
	if err != nil {
		return model.User{}, model.LoginSession{}, err
	}

	profile, err := s.feishuClient.GetUserInfo(ctx, token.AccessToken)
	if err != nil {
		return model.User{}, model.LoginSession{}, err
	}

	user := mapProfileToUser(profile)
	if existing, err := s.repository.FindUserByID(ctx, user.ID); err == nil {
		user.Role = existing.Role
		if len(existing.Departments) > 0 {
			user.Departments = existing.Departments
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, model.LoginSession{}, err
	}

	now := time.Now().UTC()
	credential := model.FeishuCredential{
		ID:                    "cred_" + user.ID,
		UserID:                user.ID,
		OpenID:                profile.OpenID,
		UnionID:               profile.UnionID,
		FeishuUserID:          profile.FeishuUserID,
		AccessToken:           token.AccessToken,
		RefreshToken:          token.RefreshToken,
		AccessTokenExpiresAt:  token.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: token.RefreshTokenExpiresAt,
		LastLoginAt:           now,
		LastRefreshAt:         now,
	}
	loginSession := model.LoginSession{
		ID:        utils.NewID("login"),
		UserID:    user.ID,
		ExpiresAt: now.Add(s.sessionTTL),
	}

	if err := s.repository.SaveFeishuLoginState(ctx, &repo.FeishuLoginState{
		User:       user,
		Credential: credential,
		Session:    loginSession,
	}); err != nil {
		return model.User{}, model.LoginSession{}, err
	}
	return user, loginSession, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID string) (model.User, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return model.User{}, ErrAuthenticationRequired
	}

	user, err := s.repository.FindUserByID(ctx, userID)
	if err == nil {
		if _, refreshErr := s.EnsureFreshCredential(ctx, userID); refreshErr != nil && !errors.Is(refreshErr, gorm.ErrRecordNotFound) {
			return model.User{}, refreshErr
		}
		return user, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, ErrAuthenticationRequired
	}
	return user, err
}

func (s *AuthService) ResolveSessionUserID(ctx context.Context, sessionID string) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", ErrAuthenticationRequired
	}

	if err := s.repository.DeleteExpiredLoginSessions(ctx, time.Now().UTC()); err != nil {
		return "", err
	}

	session, err := s.repository.FindLoginSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrAuthenticationRequired
		}
		return "", err
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		_ = s.repository.DeleteLoginSessionByID(ctx, sessionID)
		return "", ErrAuthenticationRequired
	}
	return session.UserID, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	return s.repository.DeleteLoginSessionByID(ctx, sessionID)
}

func (s *AuthService) EnsureFreshCredential(ctx context.Context, userID string) (model.FeishuCredential, error) {
	credential, err := s.repository.FindCredentialByUserID(ctx, userID)
	if err != nil {
		return model.FeishuCredential{}, err
	}

	now := time.Now().UTC()
	if credential.AccessTokenExpiresAt.After(now.Add(1 * time.Minute)) {
		return credential, nil
	}
	if credential.RefreshTokenExpiresAt.Before(now) {
		return model.FeishuCredential{}, ErrAuthenticationRequired
	}

	token, err := s.feishuClient.RefreshUserToken(ctx, credential.RefreshToken)
	if err != nil {
		return model.FeishuCredential{}, err
	}

	credential.AccessToken = token.AccessToken
	credential.RefreshToken = token.RefreshToken
	credential.AccessTokenExpiresAt = token.AccessTokenExpiresAt
	credential.RefreshTokenExpiresAt = token.RefreshTokenExpiresAt
	credential.LastRefreshAt = now
	if err := s.repository.SaveCredential(ctx, &credential); err != nil {
		return model.FeishuCredential{}, err
	}
	return credential, nil
}

func mapProfileToUser(profile feishu.UserProfile) model.User {
	return model.User{
		ID:           "fs_" + profile.OpenID,
		FeishuOpenID: profile.OpenID,
		Name:         utils.Coalesce(profile.Name, profile.EnName, "飞书用户"),
		Email:        utils.Coalesce(profile.EnterpriseEmail, profile.Email),
		Role:         model.RoleProduct,
		Departments:  []string{"飞书组织"},
	}
}
