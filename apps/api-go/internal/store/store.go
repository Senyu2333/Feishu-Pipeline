package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/domain"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(ctx context.Context, databasePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create database dir: %w", err)
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		return nil, err
	}
	if err := s.seed(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			feishu_open_id TEXT,
			name TEXT NOT NULL,
			email TEXT,
			role TEXT NOT NULL,
			departments_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			summary TEXT NOT NULL,
			status TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS requirements (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			summary TEXT NOT NULL,
			status TEXT NOT NULL,
			delivery_summary TEXT NOT NULL,
			referenced_knowledge_json TEXT NOT NULL,
			published_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL,
			assignee_name TEXT NOT NULL,
			assignee_role TEXT NOT NULL,
			acceptance_criteria_json TEXT NOT NULL,
			risks_json TEXT NOT NULL,
			doc_url TEXT NOT NULL,
			bitable_record_url TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS role_mappings (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			keyword TEXT NOT NULL,
			role TEXT NOT NULL,
			departments_json TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS knowledge_sources (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS message_deliveries (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			channel TEXT NOT NULL,
			receiver TEXT NOT NULL,
			status TEXT NOT NULL,
			remote_id TEXT NOT NULL,
			raw_payload TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
	}

	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate statement failed: %w", err)
		}
	}

	return nil
}

func (s *Store) seed(ctx context.Context) error {
	count := 0
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		now := nowString()
		users := []domain.User{
			{ID: "u_product_demo", Name: "产品经理小明", Email: "product@example.com", Role: domain.RoleProduct, Departments: []string{"产品部"}},
			{ID: "u_frontend_demo", Name: "前端负责人小红", Email: "frontend@example.com", Role: domain.RoleFrontend, Departments: []string{"前端部"}},
			{ID: "u_backend_demo", Name: "后端负责人小李", Email: "backend@example.com", Role: domain.RoleBackend, Departments: []string{"后端部"}},
			{ID: "u_admin_demo", Name: "管理员小周", Email: "admin@example.com", Role: domain.RoleAdmin, Departments: []string{"平台治理组"}},
		}
		for _, user := range users {
			if _, err := s.db.ExecContext(
				ctx,
				`INSERT INTO users (id, feishu_open_id, name, email, role, departments_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				user.ID,
				user.FeishuOpenID,
				user.Name,
				user.Email,
				string(user.Role),
				mustJSON(user.Departments),
				now,
				now,
			); err != nil {
				return err
			}
		}
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM role_mappings`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		mappings := []domain.RoleMapping{
			{ID: "rm_product", Name: "产品角色", Keyword: "产品", Role: domain.RoleProduct, Departments: []string{"产品部"}},
			{ID: "rm_frontend", Name: "前端角色", Keyword: "前端", Role: domain.RoleFrontend, Departments: []string{"前端部"}},
			{ID: "rm_backend", Name: "后端角色", Keyword: "后端", Role: domain.RoleBackend, Departments: []string{"后端部"}},
		}
		for _, mapping := range mappings {
			if err := s.SaveRoleMapping(ctx, mapping); err != nil {
				return err
			}
		}
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM knowledge_sources`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		now := time.Now()
		sources := []domain.KnowledgeSource{
			{ID: "ks_api", Title: "接口规范", Content: "接口返回统一 JSON 包装，错误不暴露内部堆栈，状态码与业务状态分离。", UpdatedAt: now},
			{ID: "ks_ui", Title: "UI 规范", Content: "涉及前端任务时，优先保证列表检索、表单反馈、状态标签与空态说明。", UpdatedAt: now},
			{ID: "ks_delivery", Title: "提测流程", Content: "开发完成后进入已提测状态，并同步测试负责人和需求会话状态。", UpdatedAt: now},
		}
		if err := s.SaveKnowledgeSources(ctx, sources); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) FindUserByID(ctx context.Context, userID string) (domain.User, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, feishu_open_id, name, email, role, departments_json FROM users WHERE id = ?`, userID)
	return scanUser(row)
}

func (s *Store) UpsertUser(ctx context.Context, user domain.User) error {
	now := nowString()
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO users (id, feishu_open_id, name, email, role, departments_json, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		 feishu_open_id = excluded.feishu_open_id,
		 name = excluded.name,
		 email = excluded.email,
		 role = excluded.role,
		 departments_json = excluded.departments_json,
		 updated_at = excluded.updated_at`,
		user.ID,
		user.FeishuOpenID,
		user.Name,
		user.Email,
		string(user.Role),
		mustJSON(user.Departments),
		now,
		now,
	)
	return err
}

func (s *Store) ListSessions(ctx context.Context) ([]domain.Session, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.id, s.title, s.summary, s.status, s.owner_id, u.name, s.created_at, s.updated_at,
		       (SELECT COUNT(1) FROM messages m WHERE m.session_id = s.id)
		FROM sessions s
		JOIN users u ON u.id = s.owner_id
		ORDER BY s.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (s *Store) CreateSession(ctx context.Context, owner domain.User, title, prompt string) (domain.SessionDetail, error) {
	now := time.Now().UTC()
	session := domain.Session{
		ID:        newID("sess"),
		Title:     title,
		Summary:   prompt,
		Status:    domain.SessionDraft,
		OwnerID:   owner.ID,
		OwnerName: owner.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := s.db.ExecContext(ctx, `INSERT INTO sessions (id, title, summary, status, owner_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.Title, session.Summary, string(session.Status), session.OwnerID, formatTime(now), formatTime(now)); err != nil {
		return domain.SessionDetail{}, err
	}

	createdMessage, err := s.AddMessage(ctx, session.ID, domain.MessageUser, prompt)
	if err != nil {
		return domain.SessionDetail{}, err
	}

	return domain.SessionDetail{
		Session:  session,
		Messages: []domain.Message{createdMessage},
	}, nil
}

func (s *Store) AddMessage(ctx context.Context, sessionID string, role domain.MessageRole, content string) (domain.Message, error) {
	msg := domain.Message{
		ID:        newID("msg"),
		SessionID: sessionID,
		Role:      role,
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now().UTC(),
	}

	if _, err := s.db.ExecContext(ctx, `INSERT INTO messages (id, session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		msg.ID, msg.SessionID, string(msg.Role), msg.Content, formatTime(msg.CreatedAt)); err != nil {
		return domain.Message{}, err
	}

	if _, err := s.db.ExecContext(ctx, `UPDATE sessions SET updated_at = ?, summary = ? WHERE id = ?`,
		formatTime(msg.CreatedAt), summarize(msg.Content), sessionID); err != nil {
		return domain.Message{}, err
	}

	return msg, nil
}

func (s *Store) GetSessionDetail(ctx context.Context, sessionID string) (domain.SessionDetail, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT s.id, s.title, s.summary, s.status, s.owner_id, u.name, s.created_at, s.updated_at,
		       (SELECT COUNT(1) FROM messages m WHERE m.session_id = s.id)
		FROM sessions s
		JOIN users u ON u.id = s.owner_id
		WHERE s.id = ?`, sessionID)
	session, err := scanSession(row)
	if err != nil {
		return domain.SessionDetail{}, err
	}

	messages, err := s.listMessages(ctx, sessionID)
	if err != nil {
		return domain.SessionDetail{}, err
	}

	requirement, reqErr := s.getRequirement(ctx, sessionID)
	if reqErr != nil && reqErr != sql.ErrNoRows {
		return domain.SessionDetail{}, reqErr
	}

	tasks, err := s.listTasksBySession(ctx, sessionID)
	if err != nil {
		return domain.SessionDetail{}, err
	}

	var reqPtr *domain.Requirement
	if reqErr == nil {
		reqPtr = &requirement
	}

	return domain.SessionDetail{
		Session:     session,
		Messages:    messages,
		Requirement: reqPtr,
		Tasks:       tasks,
	}, nil
}

func (s *Store) GetTask(ctx context.Context, taskID string) (domain.Task, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, session_id, title, description, type, status, assignee_name, assignee_role,
		       acceptance_criteria_json, risks_json, doc_url, bitable_record_url, created_at, updated_at
		FROM tasks WHERE id = ?`, taskID)
	return scanTask(row)
}

func (s *Store) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) (domain.Task, error) {
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?`,
		string(status), formatTime(now), taskID); err != nil {
		return domain.Task{}, err
	}

	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}

	if err := s.refreshSessionStatus(ctx, task.SessionID); err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

func (s *Store) MarkSessionPublished(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET status = ?, updated_at = ? WHERE id = ?`,
		string(domain.SessionPublished), nowString(), sessionID)
	return err
}

func (s *Store) SavePublishResult(ctx context.Context, requirement domain.Requirement, tasks []domain.Task, deliveries []domain.DeliveryRecord) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `INSERT INTO requirements (id, session_id, title, summary, status, delivery_summary, referenced_knowledge_json, published_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET
			title = excluded.title,
			summary = excluded.summary,
			status = excluded.status,
			delivery_summary = excluded.delivery_summary,
			referenced_knowledge_json = excluded.referenced_knowledge_json,
			published_at = excluded.published_at`,
		requirement.ID,
		requirement.SessionID,
		requirement.Title,
		requirement.Summary,
		string(requirement.Status),
		requirement.DeliverySummary,
		mustJSON(requirement.ReferencedKnowledge),
		formatTime(requirement.PublishedAt),
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM tasks WHERE session_id = ?`, requirement.SessionID); err != nil {
		return err
	}
	for _, task := range tasks {
		if _, err := tx.ExecContext(ctx, `INSERT INTO tasks (id, session_id, title, description, type, status, assignee_name, assignee_role, acceptance_criteria_json, risks_json, doc_url, bitable_record_url, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			task.ID,
			task.SessionID,
			task.Title,
			task.Description,
			string(task.Type),
			string(task.Status),
			task.AssigneeName,
			string(task.AssigneeRole),
			mustJSON(task.AcceptanceCriteria),
			mustJSON(task.Risks),
			task.DocURL,
			task.BitableRecordURL,
			formatTime(task.CreatedAt),
			formatTime(task.UpdatedAt),
		); err != nil {
			return err
		}
	}

	for _, record := range deliveries {
		if _, err := tx.ExecContext(ctx, `INSERT INTO message_deliveries (id, task_id, channel, receiver, status, remote_id, raw_payload, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			record.ID,
			record.TaskID,
			record.Channel,
			record.Receiver,
			record.Status,
			record.RemoteID,
			record.RawPayload,
			formatTime(record.CreatedAt),
		); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET status = ?, summary = ?, updated_at = ? WHERE id = ?`,
		string(domain.SessionInDelivery), summarize(requirement.Summary), nowString(), requirement.SessionID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) SaveRoleMapping(ctx context.Context, mapping domain.RoleMapping) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO role_mappings (id, name, keyword, role, departments_json) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, keyword = excluded.keyword, role = excluded.role, departments_json = excluded.departments_json`,
		coalesce(mapping.ID, newID("rm")),
		mapping.Name,
		mapping.Keyword,
		string(mapping.Role),
		mustJSON(mapping.Departments),
	)
	return err
}

func (s *Store) ListRoleMappings(ctx context.Context) ([]domain.RoleMapping, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, keyword, role, departments_json FROM role_mappings ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.RoleMapping
	for rows.Next() {
		var mapping domain.RoleMapping
		var departmentsJSON string
		if err := rows.Scan(&mapping.ID, &mapping.Name, &mapping.Keyword, &mapping.Role, &departmentsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(departmentsJSON), &mapping.Departments); err != nil {
			return nil, err
		}
		items = append(items, mapping)
	}
	return items, rows.Err()
}

func (s *Store) SaveKnowledgeSources(ctx context.Context, items []domain.KnowledgeSource) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `INSERT INTO knowledge_sources (id, title, content, updated_at) VALUES (?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET title = excluded.title, content = excluded.content, updated_at = excluded.updated_at`,
			coalesce(item.ID, newID("ks")),
			item.Title,
			item.Content,
			formatTime(item.UpdatedAt),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) SearchKnowledgeSources(ctx context.Context, query string, limit int) ([]domain.KnowledgeSource, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, content, updated_at FROM knowledge_sources WHERE title LIKE ? OR content LIKE ? ORDER BY updated_at DESC LIMIT ?`,
		"%"+query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.KnowledgeSource
	for rows.Next() {
		var item domain.KnowledgeSource
		var updatedAt string
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &updatedAt); err != nil {
			return nil, err
		}
		item.UpdatedAt = parseTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) listMessages(ctx context.Context, sessionID string) ([]domain.Message, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, role, content, created_at FROM messages WHERE session_id = ? ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Message
	for rows.Next() {
		var msg domain.Message
		var createdAt string
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &createdAt); err != nil {
			return nil, err
		}
		msg.CreatedAt = parseTime(createdAt)
		items = append(items, msg)
	}
	return items, rows.Err()
}

func (s *Store) getRequirement(ctx context.Context, sessionID string) (domain.Requirement, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, session_id, title, summary, status, delivery_summary, referenced_knowledge_json, published_at FROM requirements WHERE session_id = ?`, sessionID)
	var req domain.Requirement
	var referenced string
	var publishedAt string
	if err := row.Scan(&req.ID, &req.SessionID, &req.Title, &req.Summary, &req.Status, &req.DeliverySummary, &referenced, &publishedAt); err != nil {
		return domain.Requirement{}, err
	}
	if err := json.Unmarshal([]byte(referenced), &req.ReferencedKnowledge); err != nil {
		return domain.Requirement{}, err
	}
	req.PublishedAt = parseTime(publishedAt)
	return req, nil
}

func (s *Store) listTasksBySession(ctx context.Context, sessionID string) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, session_id, title, description, type, status, assignee_name, assignee_role,
		       acceptance_criteria_json, risks_json, doc_url, bitable_record_url, created_at, updated_at
		FROM tasks WHERE session_id = ? ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, task)
	}
	return items, rows.Err()
}

func (s *Store) refreshSessionStatus(ctx context.Context, sessionID string) error {
	rows, err := s.db.QueryContext(ctx, `SELECT status FROM tasks WHERE session_id = ?`, sessionID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var statuses []domain.TaskStatus
	for rows.Next() {
		var status domain.TaskStatus
		if err := rows.Scan(&status); err != nil {
			return err
		}
		statuses = append(statuses, status)
	}
	if len(statuses) == 0 {
		return nil
	}

	nextStatus := domain.SessionInDelivery
	allDone := true
	allTestingOrDone := true
	for _, status := range statuses {
		if status != domain.TaskDone {
			allDone = false
		}
		if status != domain.TaskDone && status != domain.TaskTesting {
			allTestingOrDone = false
		}
	}
	if allDone {
		nextStatus = domain.SessionDone
	} else if allTestingOrDone {
		nextStatus = domain.SessionTesting
	}

	_, err = s.db.ExecContext(ctx, `UPDATE sessions SET status = ?, updated_at = ? WHERE id = ?`, string(nextStatus), nowString(), sessionID)
	return err
}

func scanUser(scanner interface{ Scan(dest ...any) error }) (domain.User, error) {
	var user domain.User
	var departmentsJSON string
	if err := scanner.Scan(&user.ID, &user.FeishuOpenID, &user.Name, &user.Email, &user.Role, &departmentsJSON); err != nil {
		return domain.User{}, err
	}
	if err := json.Unmarshal([]byte(departmentsJSON), &user.Departments); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func scanSession(scanner interface{ Scan(dest ...any) error }) (domain.Session, error) {
	var session domain.Session
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(&session.ID, &session.Title, &session.Summary, &session.Status, &session.OwnerID, &session.OwnerName, &createdAt, &updatedAt, &session.MessageCount); err != nil {
		return domain.Session{}, err
	}
	session.CreatedAt = parseTime(createdAt)
	session.UpdatedAt = parseTime(updatedAt)
	return session, nil
}

func scanTask(scanner interface{ Scan(dest ...any) error }) (domain.Task, error) {
	var task domain.Task
	var acceptanceJSON string
	var risksJSON string
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&task.ID,
		&task.SessionID,
		&task.Title,
		&task.Description,
		&task.Type,
		&task.Status,
		&task.AssigneeName,
		&task.AssigneeRole,
		&acceptanceJSON,
		&risksJSON,
		&task.DocURL,
		&task.BitableRecordURL,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.Task{}, err
	}
	if err := json.Unmarshal([]byte(acceptanceJSON), &task.AcceptanceCriteria); err != nil {
		return domain.Task{}, err
	}
	if err := json.Unmarshal([]byte(risksJSON), &task.Risks); err != nil {
		return domain.Task{}, err
	}
	task.CreatedAt = parseTime(createdAt)
	task.UpdatedAt = parseTime(updatedAt)
	return task, nil
}

func mustJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func nowString() string {
	return formatTime(time.Now())
}

func parseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed
}

func summarize(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 120 {
		return value
	}
	return value[:120] + "..."
}

func coalesce(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
