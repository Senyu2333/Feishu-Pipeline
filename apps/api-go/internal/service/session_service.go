package service

import (
	"context"
	"strings"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/repo"
	"feishu-pipeline/apps/api-go/internal/utils"
)

type SessionPublisher interface {
	PublishSession(ctx context.Context, userID string, sessionID string) error
}

type SessionService struct {
	repository  *repo.Repository
	authService *AuthService
	publisher   SessionPublisher
}

func NewSessionService(repository *repo.Repository, authService *AuthService) *SessionService {
	return &SessionService{
		repository:  repository,
		authService: authService,
	}
}

func (s *SessionService) SetPublisher(publisher SessionPublisher) {
	s.publisher = publisher
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
	user, err := s.authService.CurrentUser(ctx, userID)
	if err != nil {
		return err
	}

	if _, err := s.repository.AddMessage(ctx, sessionID, model.MessageUser, content); err != nil {
		return err
	}

	aggregate, err := s.repository.GetSessionAggregate(ctx, sessionID)
	if err != nil {
		return err
	}

	if isPublishIntent(content) {
		if aggregate.Session.Status == model.SessionDraft {
			if user.Role == model.RoleProduct || user.Role == model.RoleAdmin {
				if s.publisher != nil {
					if err := s.publisher.PublishSession(ctx, userID, sessionID); err != nil {
						return err
					} else {
						_, err = s.repository.AddMessage(ctx, sessionID, model.MessageAssistant, autoPublishAcceptedReply(content))
						return err
					}
				}
			} else {
				_, err = s.repository.AddMessage(ctx, sessionID, model.MessageAssistant, publishPermissionDeniedReply(content))
				return err
			}
		}
	}

	reply := draftAssistantReply(content)
	if aggregate.Requirement != nil {
		reply = publishedAssistantReply(content, aggregate.Requirement.Summary)
	} else if aggregate.Session.Status != model.SessionDraft {
		reply = publishInProgressReply(content)
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

func isPublishIntent(content string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(content), " ", ""))
	if normalized == "" {
		return false
	}

	keywords := []string{
		"发布需求",
		"发布这个需求",
		"请发布",
		"帮我发布",
		"需求发布",
		"发版需求",
		"正式发布",
		"submitrequirement",
		"publishrequirement",
		"publishthisrequirement",
	}
	for _, keyword := range keywords {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return false
}

func autoPublishAcceptedReply(question string) string {
	return "已识别到“发布需求”意图，系统已自动创建需求编号并触发交付工作流（任务拆解、负责人匹配、飞书通知、文档分发）。你可以继续补充信息：" + utils.Summarize(question, 120)
}

func publishPermissionDeniedReply(question string) string {
	return "已识别到发布意图，但当前账号不是产品角色，不能发布需求。你可以继续咨询项目信息或让产品同学执行发布。当前消息：" + utils.Summarize(question, 120)
}

func publishInProgressReply(question string) string {
	return "当前需求已进入发布流程，正在生成任务与分发结果。你可继续提问，系统将基于已发布上下文回答：" + utils.Summarize(question, 120)
}
