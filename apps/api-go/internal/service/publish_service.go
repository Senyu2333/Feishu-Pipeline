package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/agent"
	"feishu-pipeline/apps/api-go/internal/external/feishu"
	"feishu-pipeline/apps/api-go/internal/job"
	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/repo"
	"feishu-pipeline/apps/api-go/internal/utils"
)

type PublishQueue interface {
	Enqueue(job.PublishJob)
}

type PublishService struct {
	repository   *repo.Repository
	authService  *AuthService
	agentEngine  *agent.Engine
	feishuClient *feishu.Client
	queue        PublishQueue
}

func NewPublishService(repository *repo.Repository, authService *AuthService, agentEngine *agent.Engine, feishuClient *feishu.Client) *PublishService {
	return &PublishService{
		repository:   repository,
		authService:  authService,
		agentEngine:  agentEngine,
		feishuClient: feishuClient,
	}
}

func (s *PublishService) SetQueue(queue PublishQueue) {
	s.queue = queue
}

func (s *PublishService) PublishSession(ctx context.Context, userID string, sessionID string) error {
	user, err := s.authService.CurrentUser(ctx, userID)
	if err != nil {
		return err
	}
	if user.Role != model.RoleProduct && user.Role != model.RoleAdmin {
		return errors.New("only product or admin can publish requirement")
	}

	aggregate, err := s.repository.GetSessionAggregate(ctx, sessionID)
	if err != nil {
		return err
	}
	if aggregate.Session.Status != model.SessionDraft {
		return errors.New("session is not in draft status")
	}

	if err := s.repository.MarkSessionPublished(ctx, sessionID); err != nil {
		return err
	}
	if s.queue == nil {
		return errors.New("publish queue not configured")
	}

	s.queue.Enqueue(job.PublishJob{SessionID: sessionID})
	return nil
}

func (s *PublishService) TryAutoPublishByMessage(ctx context.Context, userID string, sessionID string, content string) (bool, string, error) {
	user, err := s.authService.CurrentUser(ctx, userID)
	if err != nil {
		return false, "authentication required", err
	}
	if user.Role != model.RoleProduct && user.Role != model.RoleAdmin {
		return false, "user role is not product/admin", nil
	}
	if !containsScheduleSignal(content) {
		return false, "message does not include schedule signal", nil
	}

	aggregate, err := s.repository.GetSessionAggregate(ctx, sessionID)
	if err != nil {
		return false, "session not found", err
	}
	if aggregate.Session.Status != model.SessionDraft {
		return false, "session is not in draft status", nil
	}

	if err := s.PublishSession(ctx, userID, sessionID); err != nil {
		return false, "auto publish failed", err
	}
	return true, "auto publish accepted", nil
}

func containsScheduleSignal(content string) bool {
	text := strings.ToLower(strings.TrimSpace(content))
	if text == "" {
		return false
	}

	keywords := []string{
		"排期", "上线", "截止", "ddl", "deadline", "milestone", "里程碑", "交付时间", "本周", "下周",
	}
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func (s *PublishService) HandlePublish(ctx context.Context, payload job.PublishJob) error {
	aggregate, err := s.repository.GetSessionAggregate(ctx, payload.SessionID)
	if err != nil {
		return err
	}

	mappings, err := s.repository.ListRoleMappings(ctx)
	if err != nil {
		return err
	}
	knowledge, err := s.repository.SearchKnowledgeSources(ctx, aggregate.Session.Summary, 5)
	if err != nil {
		return err
	}

	output, err := s.agentEngine.Execute(ctx, agent.PublishInput{
		Session:      aggregate,
		RoleMappings: mappings,
		Knowledge:    knowledge,
	})
	if err != nil {
		return err
	}

	deliveries := make([]model.MessageDelivery, 0, len(output.Tasks))
	for idx := range output.Tasks {
		docURL, err := s.feishuClient.CreateTaskDoc(ctx, aggregate.Session.Title, output.Tasks[idx])
		if err != nil {
			return err
		}
		recordURL, err := s.feishuClient.UpsertTaskRecord(ctx, output.Tasks[idx])
		if err != nil {
			return err
		}
		output.Tasks[idx].DocURL = docURL
		output.Tasks[idx].BitableRecordURL = recordURL

		sendResult, err := s.feishuClient.SendTaskMessage(ctx, output.Tasks[idx])
		if err != nil {
			return err
		}
		deliveries = append(deliveries, model.MessageDelivery{
			ID:         utils.NewID("delivery"),
			TaskID:     output.Tasks[idx].ID,
			Channel:    sendResult.Channel,
			Receiver:   sendResult.Receiver,
			Status:     sendResult.Status,
			RemoteID:   sendResult.RemoteID,
			RawPayload: sendResult.RawPayload,
			CreatedAt:  time.Now().UTC(),
		})
	}

	return s.repository.SavePublishResult(ctx, repo.PublishResult{
		Requirement: output.Requirement,
		Tasks:       output.Tasks,
		Deliveries:  deliveries,
	})
}
