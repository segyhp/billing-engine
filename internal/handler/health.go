package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/segyhp/billing-engine/pkg/response"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db    *sqlx.DB
	redis *redis.Client
}

func NewHealthHandler(db *sqlx.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// Health performs a basic health check
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now(),
		Checks:    make(map[string]string),
	}

	response.Success(w, status)
}

// Ready performs readiness check including database and redis connectivity
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now(),
		Checks:    make(map[string]string),
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		status.Status = "error"
		status.Checks["database"] = "failed: " + err.Error()
	} else {
		status.Checks["database"] = "ok"
	}

	// Check Redis connectivity
	redisCtx, redisCancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer redisCancel()

	if err := h.redis.Ping(redisCtx).Err(); err != nil {
		status.Status = "error"
		status.Checks["redis"] = "failed: " + err.Error()
	} else {
		status.Checks["redis"] = "ok"
	}

	if status.Status == "error" {
		response.Error(w, http.StatusServiceUnavailable, "Service not ready", nil)
		return
	}

	response.Success(w, status)
}
