package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	agenttype "feishu-pipeline/apps/api-go/internal/type/agent"
	"feishu-pipeline/apps/api-go/internal/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SessionSummary struct {
	Session      model.Session
	OwnerName    string
	MessageCount int
}

type SessionAggregate = agenttype.SessionAggregate

type PublishResult struct {
	Requirement model.Requirement
	Tasks       []model.Task
	Deliveries  []model.MessageDelivery
}

type Repository struct {
	db *gorm.DB
}

func NewSQLiteRepository(databasePath string) (*Repository, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create database dir: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite with gorm: %w", err)
	}

	repository := &Repository{db: db}
	if err := repository.AutoMigrate(context.Background()); err != nil {
		return nil, err
	}
	if err := repository.Seed(context.Background()); err != nil {
		return nil, err
	}
	return repository, nil
}

func (r *Repository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (r *Repository) AutoMigrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(
		&model.User{},
		&model.Session{},
		&model.Message{},
		&model.Requirement{},
		&model.Task{},
		&model.RoleMapping{},
		&model.KnowledgeSource{},
		&model.MessageDelivery{},
	)
}

func (r *Repository) Seed(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.User{}).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			users := []model.User{
				{ID: "u_product_demo", Name: "产品经理小明", Email: "product@example.com", Role: model.RoleProduct, Departments: []string{"产品部"}},
				{ID: "u_frontend_demo", Name: "前端负责人小红", Email: "frontend@example.com", Role: model.RoleFrontend, Departments: []string{"前端部"}},
				{ID: "u_backend_demo", Name: "后端负责人小李", Email: "backend@example.com", Role: model.RoleBackend, Departments: []string{"后端部"}},
				{ID: "u_admin_demo", Name: "管理员小周", Email: "admin@example.com", Role: model.RoleAdmin, Departments: []string{"平台治理组"}},
			}
			if err := tx.Create(&users).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&model.RoleMapping{}).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			mappings := []model.RoleMapping{
				{ID: "rm_product", Name: "产品角色", Keyword: "产品", Role: model.RoleProduct, Departments: []string{"产品部"}},
				{ID: "rm_frontend", Name: "前端角色", Keyword: "前端", Role: model.RoleFrontend, Departments: []string{"前端部"}},
				{ID: "rm_backend", Name: "后端角色", Keyword: "后端", Role: model.RoleBackend, Departments: []string{"后端部"}},
			}
			if err := tx.Create(&mappings).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&model.KnowledgeSource{}).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			now := time.Now().UTC()
			items := []model.KnowledgeSource{
				{ID: "ks_api", Title: "接口规范", Content: "接口返回统一 JSON 包装，错误不暴露内部堆栈，状态码与业务状态分离。", UpdatedAt: now},
				{ID: "ks_ui", Title: "UI 规范", Content: "涉及前端任务时，优先保证列表检索、表单反馈、状态标签与空态说明。", UpdatedAt: now},
				{ID: "ks_delivery", Title: "提测流程", Content: "开发完成后进入已提测状态，并同步测试负责人和需求会话状态。", UpdatedAt: now},
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Repository) FindUserByID(ctx context.Context, userID string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error
	return user, err
}

func (r *Repository) UpsertUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *Repository) ListSessions(ctx context.Context) ([]SessionSummary, error) {
	var sessions []model.Session
	if err := r.db.WithContext(ctx).Preload("Owner").Order("updated_at DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}

	items := make([]SessionSummary, 0, len(sessions))
	for _, session := range sessions {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.Message{}).Where("session_id = ?", session.ID).Count(&count).Error; err != nil {
			return nil, err
		}
		items = append(items, SessionSummary{
			Session:      session,
			OwnerName:    session.Owner.Name,
			MessageCount: int(count),
		})
	}
	return items, nil
}

func (r *Repository) CreateSession(ctx context.Context, owner model.User, title string, prompt string) (model.Session, error) {
	session := model.Session{
		ID:      utils.NewID("sess"),
		Title:   title,
		Summary: prompt,
		Status:  model.SessionDraft,
		OwnerID: owner.ID,
	}
	return session, r.db.WithContext(ctx).Create(&session).Error
}

func (r *Repository) AddMessage(ctx context.Context, sessionID string, role model.MessageRole, content string) (model.Message, error) {
	message := model.Message{
		ID:        utils.NewID("msg"),
		SessionID: sessionID,
		Role:      role,
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now().UTC(),
	}
	if err := r.db.WithContext(ctx).Create(&message).Error; err != nil {
		return model.Message{}, err
	}
	if err := r.db.WithContext(ctx).Model(&model.Session{}).Where("id = ?", sessionID).Updates(map[string]any{
		"summary":    utils.Summarize(message.Content, 120),
		"updated_at": message.CreatedAt,
	}).Error; err != nil {
		return model.Message{}, err
	}
	return message, nil
}

func (r *Repository) GetSessionAggregate(ctx context.Context, sessionID string) (*agenttype.SessionAggregate, error) {
	var session model.Session
	if err := r.db.WithContext(ctx).Preload("Owner").First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}

	var messages []model.Message
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&messages, "session_id = ?", sessionID).Error; err != nil {
		return nil, err
	}

	var requirement model.Requirement
	reqErr := r.db.WithContext(ctx).First(&requirement, "session_id = ?", sessionID).Error
	var reqPtr *model.Requirement
	if reqErr == nil {
		reqPtr = &requirement
	} else if reqErr != gorm.ErrRecordNotFound {
		return nil, reqErr
	}

	var tasks []model.Task
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&tasks, "session_id = ?", sessionID).Error; err != nil {
		return nil, err
	}

	var messageCount int64
	if err := r.db.WithContext(ctx).Model(&model.Message{}).Where("session_id = ?", sessionID).Count(&messageCount).Error; err != nil {
		return nil, err
	}

	return &agenttype.SessionAggregate{
		Session:      session,
		Owner:        session.Owner,
		MessageCount: int(messageCount),
		Messages:     messages,
		Requirement:  reqPtr,
		Tasks:        tasks,
	}, nil
}

func (r *Repository) MarkSessionPublished(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).Model(&model.Session{}).Where("id = ?", sessionID).Updates(map[string]any{
		"status":     model.SessionPublished,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (r *Repository) SavePublishResult(ctx context.Context, result PublishResult) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result.Requirement.Status = model.SessionInDelivery
		if err := tx.Where("session_id = ?", result.Requirement.SessionID).Delete(&model.Task{}).Error; err != nil {
			return err
		}

		if err := tx.Where("session_id = ?", result.Requirement.SessionID).
			Assign(result.Requirement).
			FirstOrCreate(&model.Requirement{SessionID: result.Requirement.SessionID}).Error; err != nil {
			return err
		}

		if len(result.Tasks) > 0 {
			if err := tx.Create(&result.Tasks).Error; err != nil {
				return err
			}
		}
		if len(result.Deliveries) > 0 {
			if err := tx.Create(&result.Deliveries).Error; err != nil {
				return err
			}
		}

		return tx.Model(&model.Session{}).Where("id = ?", result.Requirement.SessionID).Updates(map[string]any{
			"status":     model.SessionInDelivery,
			"summary":    utils.Summarize(result.Requirement.Summary, 120),
			"updated_at": time.Now().UTC(),
		}).Error
	})
}

func (r *Repository) GetTaskByID(ctx context.Context, taskID string) (model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).First(&task, "id = ?", taskID).Error
	return task, err
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, taskID string, status model.TaskStatus) (model.Task, error) {
	if err := r.db.WithContext(ctx).Model(&model.Task{}).Where("id = ?", taskID).Updates(map[string]any{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}).Error; err != nil {
		return model.Task{}, err
	}

	task, err := r.GetTaskByID(ctx, taskID)
	if err != nil {
		return model.Task{}, err
	}
	if err := r.RefreshSessionStatus(ctx, task.SessionID); err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func (r *Repository) UpdateTaskLinks(ctx context.Context, taskID string, docURL string, bitableURL string) error {
	return r.db.WithContext(ctx).Model(&model.Task{}).Where("id = ?", taskID).Updates(map[string]any{
		"doc_url":            docURL,
		"bitable_record_url": bitableURL,
		"updated_at":         time.Now().UTC(),
	}).Error
}

func (r *Repository) RefreshSessionStatus(ctx context.Context, sessionID string) error {
	var tasks []model.Task
	if err := r.db.WithContext(ctx).Find(&tasks, "session_id = ?", sessionID).Error; err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}

	nextStatus := model.SessionInDelivery
	allDone := true
	allTestingOrDone := true
	for _, task := range tasks {
		if task.Status != model.TaskDone {
			allDone = false
		}
		if task.Status != model.TaskDone && task.Status != model.TaskTesting {
			allTestingOrDone = false
		}
	}
	if allDone {
		nextStatus = model.SessionDone
	} else if allTestingOrDone {
		nextStatus = model.SessionTesting
	}

	return r.db.WithContext(ctx).Model(&model.Session{}).Where("id = ?", sessionID).Updates(map[string]any{
		"status":     nextStatus,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (r *Repository) SaveRoleMapping(ctx context.Context, mapping *model.RoleMapping) error {
	if mapping.ID == "" {
		mapping.ID = utils.NewID("rm")
	}
	return r.db.WithContext(ctx).Save(mapping).Error
}

func (r *Repository) ListRoleMappings(ctx context.Context) ([]model.RoleMapping, error) {
	var items []model.RoleMapping
	err := r.db.WithContext(ctx).Order("name ASC").Find(&items).Error
	return items, err
}

func (r *Repository) SaveKnowledgeSources(ctx context.Context, items []model.KnowledgeSource) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for idx := range items {
			if items[idx].ID == "" {
				items[idx].ID = utils.NewID("ks")
			}
			if items[idx].UpdatedAt.IsZero() {
				items[idx].UpdatedAt = time.Now().UTC()
			}
			if err := tx.Save(&items[idx]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) SearchKnowledgeSources(ctx context.Context, query string, limit int) ([]model.KnowledgeSource, error) {
	var items []model.KnowledgeSource
	if limit <= 0 {
		limit = 5
	}
	err := r.db.WithContext(ctx).
		Where("title LIKE ? OR content LIKE ?", "%"+query+"%", "%"+query+"%").
		Order("updated_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}
