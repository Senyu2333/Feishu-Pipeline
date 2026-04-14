package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/domain"
	"feishu-pipeline/apps/api-go/internal/service"
)

type Router struct {
	logger     *log.Logger
	service    *service.Service
	cookieName string
}

func NewRouter(logger *log.Logger, svc *service.Service, cookieName string) http.Handler {
	r := &Router{
		logger:     logger,
		service:    svc,
		cookieName: cookieName,
	}
	return r.routes()
}

func (r *Router) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", r.handleHealth)
	mux.HandleFunc("GET /api/auth/feishu/login", r.handleFeishuLogin)
	mux.HandleFunc("GET /api/auth/feishu/callback", r.handleFeishuCallback)
	mux.HandleFunc("GET /api/me", r.withLogging(r.handleMe))
	mux.HandleFunc("GET /api/sessions", r.withLogging(r.handleListSessions))
	mux.HandleFunc("POST /api/sessions", r.withLogging(r.handleCreateSession))
	mux.HandleFunc("GET /api/sessions/{sessionID}", r.withLogging(r.handleGetSession))
	mux.HandleFunc("POST /api/sessions/{sessionID}/messages", r.withLogging(r.handleAddMessage))
	mux.HandleFunc("POST /api/sessions/{sessionID}/publish", r.withLogging(r.handlePublishSession))
	mux.HandleFunc("GET /api/tasks/{taskID}", r.withLogging(r.handleGetTask))
	mux.HandleFunc("PATCH /api/tasks/{taskID}/status", r.withLogging(r.handleUpdateTaskStatus))
	mux.HandleFunc("POST /api/admin/role-mappings", r.withLogging(r.handleCreateRoleMapping))
	mux.HandleFunc("POST /api/admin/knowledge/sync", r.withLogging(r.handleSyncKnowledge))

	return withCORS(mux)
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": r.service.Health()})
}

func (r *Router) handleFeishuLogin(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, r.service.LoginURL("feishu-pipeline-state"), http.StatusFound)
}

func (r *Router) handleFeishuCallback(w http.ResponseWriter, req *http.Request) {
	user, err := r.service.LoginByCode(req.Context(), req.URL.Query().Get("code"))
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     r.cookieName,
		Value:    user.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, req, "/", http.StatusFound)
}

func (r *Router) handleMe(w http.ResponseWriter, req *http.Request) {
	user, err := r.service.CurrentUser(req.Context(), r.currentUserID(req))
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": user})
}

func (r *Router) handleListSessions(w http.ResponseWriter, req *http.Request) {
	sessions, err := r.service.ListSessions(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": sessions})
}

func (r *Router) handleCreateSession(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Title  string `json:"title"`
		Prompt string `json:"prompt"`
	}
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	detail, err := r.service.CreateSession(req.Context(), r.currentUserID(req), strings.TrimSpace(body.Title), strings.TrimSpace(body.Prompt))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": detail})
}

func (r *Router) handleGetSession(w http.ResponseWriter, req *http.Request) {
	detail, err := r.service.GetSessionDetail(req.Context(), req.PathValue("sessionID"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": detail})
}

func (r *Router) handleAddMessage(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	detail, err := r.service.AddSessionMessage(req.Context(), r.currentUserID(req), req.PathValue("sessionID"), body.Content)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": detail})
}

func (r *Router) handlePublishSession(w http.ResponseWriter, req *http.Request) {
	if err := r.service.PublishSession(req.Context(), r.currentUserID(req), req.PathValue("sessionID")); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"data": map[string]string{
			"status":  "accepted",
			"message": "需求已受理，后台正在生成任务和飞书分发结果。",
		},
	})
}

func (r *Router) handleGetTask(w http.ResponseWriter, req *http.Request) {
	task, err := r.service.GetTask(req.Context(), req.PathValue("taskID"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": task})
}

func (r *Router) handleUpdateTaskStatus(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Status domain.TaskStatus `json:"status"`
	}
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	task, err := r.service.UpdateTaskStatus(req.Context(), req.PathValue("taskID"), body.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": task})
}

func (r *Router) handleCreateRoleMapping(w http.ResponseWriter, req *http.Request) {
	if !r.ensureAdmin(w, req) {
		return
	}
	var body struct {
		Name        string      `json:"name"`
		Keyword     string      `json:"keyword"`
		Role        domain.Role `json:"role"`
		Departments []string    `json:"departments"`
	}
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	err := r.service.SaveRoleMapping(req.Context(), domain.RoleMapping{
		Name:        body.Name,
		Keyword:     body.Keyword,
		Role:        body.Role,
		Departments: body.Departments,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]string{"status": "saved"}})
}

func (r *Router) handleSyncKnowledge(w http.ResponseWriter, req *http.Request) {
	if !r.ensureAdmin(w, req) {
		return
	}
	var body struct {
		Sources []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"sources"`
	}
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	items := make([]domain.KnowledgeSource, 0, len(body.Sources))
	for _, source := range body.Sources {
		items = append(items, domain.KnowledgeSource{
			ID:        "",
			Title:     source.Title,
			Content:   source.Content,
			UpdatedAt: time.Now().UTC(),
		})
	}
	if err := r.service.SyncKnowledgeSources(req.Context(), items); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{"count": len(items)}})
}

func (r *Router) currentUserID(req *http.Request) string {
	if cookie, err := req.Cookie(r.cookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	if header := req.Header.Get("X-Demo-User"); header != "" {
		return header
	}
	return "u_product_demo"
}

func (r *Router) ensureAdmin(w http.ResponseWriter, req *http.Request) bool {
	user := r.service.EnsureUser(req.Context(), r.currentUserID(req))
	if user.Role != domain.RoleAdmin {
		writeError(w, http.StatusForbidden, errors.New("admin permission required"))
		return false
	}
	return true
}

func (r *Router) withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.logger.Printf("%s %s", req.Method, req.URL.Path)
		next(w, req)
	}
}

func decodeJSON(req *http.Request, target any) error {
	defer req.Body.Close()
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Demo-User")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, req)
	})
}
