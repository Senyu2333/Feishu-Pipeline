package service

import (
	"context"
	"strings"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/repo"
	"feishu-pipeline/apps/api-go/internal/utils"
)

type SessionService struct {
	repository  *repo.Repository
	authService *AuthService
}

func NewSessionService(repository *repo.Repository, authService *AuthService) *SessionService {
	return &SessionService{
		repository:  repository,
		authService: authService,
	}
}

func (s *SessionService) ListSessions(ctx context.Context) ([]repo.SessionSummary, error) {
	return s.repository.ListSessions(ctx)
}

func (s *SessionService) CreateSession(ctx context.Context, userID string, title string, prompt string) (*repo.SessionAggregate, error) {
	user, err := s.authService.CurrentUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	session, err := s.repository.CreateSession(ctx, user, strings.TrimSpace(title), strings.TrimSpace(prompt))
	if err != nil {
		return nil, err
	}

	if _, err := s.repository.AddMessage(ctx, session.ID, model.MessageUser, prompt); err != nil {
		return nil, err
	}
	if _, err := s.repository.AddMessage(ctx, session.ID, model.MessageAssistant, draftAssistantReply(prompt)); err != nil {
		return nil, err
	}
	return s.repository.GetSessionAggregate(ctx, session.ID)
}

func (s *SessionService) GetSessionDetail(ctx context.Context, sessionID string) (*repo.SessionAggregate, error) {
	aggregate, err := s.repository.GetSessionAggregate(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return &repo.SessionAggregate{
		Session:      aggregate.Session,
		Owner:        aggregate.Owner,
		MessageCount: aggregate.MessageCount,
		Messages:     aggregate.Messages,
		Requirement:  aggregate.Requirement,
		Tasks:        aggregate.Tasks,
	}, nil
}

func (s *SessionService) AddMessage(ctx context.Context, userID string, sessionID string, content string) error {
	if _, err := s.authService.CurrentUser(ctx, userID); err != nil {
		return err
	}

	if _, err := s.repository.AddMessage(ctx, sessionID, model.MessageUser, content); err != nil {
		return err
	}

	aggregate, err := s.repository.GetSessionAggregate(ctx, sessionID)
	if err != nil {
		return err
	}

	reply := draftAssistantReply(content)
	if aggregate.Requirement != nil {
		reply = publishedAssistantReply(content, aggregate.Requirement.Summary)
	}
	_, err = s.repository.AddMessage(ctx, sessionID, model.MessageAssistant, reply)
	return err
}

func draftAssistantReply(prompt string) string {
	return "已收到草稿需求，我建议先确认目标、核心流程、验收标准和前后端边界。当前整理结果：" + utils.Summarize(prompt, 140)
}

func publishedAssistantReply(question string, summary string) string {
	return "当前需求已发布，以下回答基于正式摘要，不会直接改写任务拆解结果。正式摘要：" + utils.Summarize(summary, 140) + "\n\n本次答复：" + utils.Summarize(question, 140)
}
