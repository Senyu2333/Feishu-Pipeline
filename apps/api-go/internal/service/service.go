package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/agent"
	"feishu-pipeline/apps/api-go/internal/domain"
	"feishu-pipeline/apps/api-go/internal/feishu"
	"feishu-pipeline/apps/api-go/internal/job"
	"feishu-pipeline/apps/api-go/internal/store"
)

type PublishQueue interface {
	Enqueue(job.PublishJob)
}

type Service struct {
	store   *store.Store
	agent   *agent.Engine
	feishu  *feishu.Client
	queue   PublishQueue
	version string
}

func New(store *store.Store, agentEngine *agent.Engine, feishuClient *feishu.Client, version string) *Service {
	return &Service{
		store:   store,
		agent:   agentEngine,
		feishu:  feishuClient,
		version: version,
	}
}

func (s *Service) SetQueue(queue PublishQueue) {
	s.queue = queue
}

func (s *Service) Health() map[string]string {
	return map[string]string{
		"status":  "ok",
		"service": "requirement-delivery-api",
		"version": s.version,
		"now":     time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *Service) LoginURL(state string) string {
	return s.feishu.AuthLoginURL(state)
}

func (s *Service) LoginByCode(ctx context.Context, code string) (domain.User, error) {
	user, err := s.feishu.ExchangeCode(ctx, code)
	if err != nil {
		return domain.User{}, err
	}
	if err := s.store.UpsertUser(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *Service) CurrentUser(ctx context.Context, userID string) (domain.User, error) {
	if strings.TrimSpace(userID) == "" {
		userID = "u_product_demo"
	}
	return s.store.FindUserByID(ctx, userID)
}

func (s *Service) ListSessions(ctx context.Context) ([]domain.Session, error) {
	return s.store.ListSessions(ctx)
}

func (s *Service) CreateSession(ctx context.Context, userID, title, prompt string) (domain.SessionDetail, error) {
	user, err := s.CurrentUser(ctx, userID)
	if err != nil {
		return domain.SessionDetail{}, err
	}
	detail, err := s.store.CreateSession(ctx, user, title, prompt)
	if err != nil {
		return domain.SessionDetail{}, err
	}

	assistantReply := draftAssistantReply(prompt)
	reply, err := s.store.AddMessage(ctx, detail.Session.ID, domain.MessageAssistant, assistantReply)
	if err != nil {
		return domain.SessionDetail{}, err
	}
	detail.Messages = append(detail.Messages, reply)
	return s.store.GetSessionDetail(ctx, detail.Session.ID)
}

func (s *Service) GetSessionDetail(ctx context.Context, sessionID string) (domain.SessionDetail, error) {
	return s.store.GetSessionDetail(ctx, sessionID)
}

func (s *Service) AddSessionMessage(ctx context.Context, userID, sessionID, content string) (domain.SessionDetail, error) {
	if _, err := s.CurrentUser(ctx, userID); err != nil {
		return domain.SessionDetail{}, err
	}
	if _, err := s.store.AddMessage(ctx, sessionID, domain.MessageUser, content); err != nil {
		return domain.SessionDetail{}, err
	}

	detail, err := s.store.GetSessionDetail(ctx, sessionID)
	if err != nil {
		return domain.SessionDetail{}, err
	}

	replyText := draftAssistantReply(content)
	if detail.Requirement != nil {
		replyText = publishedAssistantReply(content, detail.Requirement.Summary)
	}
	if _, err := s.store.AddMessage(ctx, sessionID, domain.MessageAssistant, replyText); err != nil {
		return domain.SessionDetail{}, err
	}
	return s.store.GetSessionDetail(ctx, sessionID)
}

func (s *Service) PublishSession(ctx context.Context, userID, sessionID string) error {
	user, err := s.CurrentUser(ctx, userID)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleProduct && user.Role != domain.RoleAdmin {
		return errors.New("only product or admin can publish requirement")
	}

	detail, err := s.store.GetSessionDetail(ctx, sessionID)
	if err != nil {
		return err
	}
	if detail.Session.Status != domain.SessionDraft {
		return fmt.Errorf("session %s is not in draft status", sessionID)
	}
	if err := s.store.MarkSessionPublished(ctx, sessionID); err != nil {
		return err
	}
	if s.queue == nil {
		return errors.New("publish queue not configured")
	}
	s.queue.Enqueue(job.PublishJob{SessionID: sessionID})
	return nil
}

func (s *Service) HandlePublish(ctx context.Context, payload job.PublishJob) error {
	detail, err := s.store.GetSessionDetail(ctx, payload.SessionID)
	if err != nil {
		return err
	}

	mappings, err := s.store.ListRoleMappings(ctx)
	if err != nil {
		return err
	}

	knowledge, err := s.store.SearchKnowledgeSources(ctx, detail.Session.Summary, 5)
	if err != nil {
		return err
	}

	output, err := s.agent.Execute(ctx, agent.PublishInput{
		Session:      detail,
		RoleMappings: mappings,
		Knowledge:    knowledge,
	})
	if err != nil {
		return err
	}

	deliveries := make([]domain.DeliveryRecord, 0, len(output.Tasks))
	for idx := range output.Tasks {
		docURL, err := s.feishu.CreateTaskDoc(ctx, detail.Session.Title, output.Tasks[idx])
		if err != nil {
			return err
		}
		recordURL, err := s.feishu.UpsertTaskRecord(ctx, output.Tasks[idx])
		if err != nil {
			return err
		}
		output.Tasks[idx].DocURL = docURL
		output.Tasks[idx].BitableRecordURL = recordURL

		sendResult, err := s.feishu.SendTaskMessage(ctx, output.Tasks[idx])
		if err != nil {
			return err
		}
		deliveries = append(deliveries, domain.DeliveryRecord{
			ID:         fmt.Sprintf("delivery_%d", time.Now().UnixNano()),
			TaskID:     output.Tasks[idx].ID,
			Channel:    sendResult.Channel,
			Receiver:   sendResult.Receiver,
			Status:     sendResult.Status,
			RemoteID:   sendResult.RemoteID,
			RawPayload: sendResult.RawPayload,
			CreatedAt:  time.Now().UTC(),
		})
	}

	return s.store.SavePublishResult(ctx, output.Requirement, output.Tasks, deliveries)
}

func (s *Service) GetTask(ctx context.Context, taskID string) (domain.Task, error) {
	return s.store.GetTask(ctx, taskID)
}

func (s *Service) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) (domain.Task, error) {
	task, err := s.store.UpdateTaskStatus(ctx, taskID, status)
	if err != nil {
		return domain.Task{}, err
	}
	recordURL, err := s.feishu.UpsertTaskRecord(ctx, task)
	if err == nil && recordURL != "" {
		task.BitableRecordURL = recordURL
	}
	return task, nil
}

func (s *Service) SaveRoleMapping(ctx context.Context, mapping domain.RoleMapping) error {
	return s.store.SaveRoleMapping(ctx, mapping)
}

func (s *Service) SyncKnowledgeSources(ctx context.Context, items []domain.KnowledgeSource) error {
	return s.store.SaveKnowledgeSources(ctx, items)
}

func (s *Service) EnsureUser(ctx context.Context, userID string) domain.User {
	user, err := s.CurrentUser(ctx, userID)
	if err == nil {
		return user
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{
			ID:          "u_product_demo",
			Name:        "产品经理小明",
			Email:       "product@example.com",
			Role:        domain.RoleProduct,
			Departments: []string{"产品部"},
		}
	}
	return domain.User{
		ID:          "u_product_demo",
		Name:        "产品经理小明",
		Email:       "product@example.com",
		Role:        domain.RoleProduct,
		Departments: []string{"产品部"},
	}
}

func draftAssistantReply(prompt string) string {
	return fmt.Sprintf("已收到草稿需求，我建议先确认目标、核心流程、验收标准和前后端边界。当前整理结果：%s", summarize(prompt))
}

func publishedAssistantReply(question string, summary string) string {
	return fmt.Sprintf("当前需求已发布，以下回答基于正式摘要，不会直接改写任务拆解结果。正式摘要：%s\n\n本次答复：%s", summarize(summary), summarize(question))
}

func summarize(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 140 {
		return value
	}
	return value[:140] + "..."
}
