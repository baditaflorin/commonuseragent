package api

import (
	"context"
	"encoding/json"

	"html"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/baditaflorin/commonuseragent"
	"github.com/baditaflorin/commonuseragent/internal/database"
)

// DB interface for database operations
type DB interface {
	LogRequest(ctx context.Context, log database.RequestLog) (int64, error)
	GetRecentRequests(ctx context.Context, limit int) ([]database.RequestLog, error)
	GetRequestsByType(ctx context.Context, agentType string, limit int) ([]database.RequestLog, error)
	GetStats(ctx context.Context) (*database.Stats, error)
	DeleteOldRequests(ctx context.Context, olderThan time.Duration) (int64, error)
}

// Handler handles HTTP requests
type Handler struct {
	db DB
}

// NewHandler creates a new API handler
func NewHandler(db DB) *Handler {
	return &Handler{db: db}
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// GetRandomDesktop returns a random desktop user agent
func (h *Handler) GetRandomDesktop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ua, err := commonuseragent.GetRandomDesktopUA()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get user agent")
		return
	}

	// Log the request
	if h.db != nil {
		ip := sanitizeIP(getClientIP(r))
		_ = h.logRequest(r.Context(), ua, "desktop", ip, r.URL.Path)
	}

	h.sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]string{
			"userAgent": ua,
			"type":      "desktop",
		},
	})
}

// GetRandomMobile returns a random mobile user agent
func (h *Handler) GetRandomMobile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ua, err := commonuseragent.GetRandomMobileUA()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get user agent")
		return
	}

	// Log the request
	if h.db != nil {
		ip := sanitizeIP(getClientIP(r))
		_ = h.logRequest(r.Context(), ua, "mobile", ip, r.URL.Path)
	}

	h.sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]string{
			"userAgent": ua,
			"type":      "mobile",
		},
	})
}

// GetRandom returns a random user agent (desktop or mobile)
func (h *Handler) GetRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ua, err := commonuseragent.GetRandomUA()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get user agent")
		return
	}

	// Log the request
	if h.db != nil {
		ip := sanitizeIP(getClientIP(r))
		_ = h.logRequest(r.Context(), ua, "random", ip, r.URL.Path)
	}

	h.sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]string{
			"userAgent": ua,
			"type":      "random",
		},
	})
}

// GetAllDesktop returns all desktop user agents
func (h *Handler) GetAllDesktop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	agents, err := commonuseragent.GetAllDesktop()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get user agents")
		return
	}

	h.sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"agents": agents,
			"count":  len(agents),
			"type":   "desktop",
		},
	})
}

// GetAllMobile returns all mobile user agents
func (h *Handler) GetAllMobile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	agents, err := commonuseragent.GetAllMobile()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get user agents")
		return
	}

	h.sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"agents": agents,
			"count":  len(agents),
			"type":   "mobile",
		},
	})
}

// GetRecentRequests returns recent request logs
func (h *Handler) GetRecentRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse and validate limit parameter
	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 1000 {
			h.sendError(w, http.StatusBadRequest, "invalid limit parameter (must be between 1 and 1000)")
			return
		}
		limit = parsedLimit
	}

	logs, err := h.db.GetRecentRequests(r.Context(), limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get recent requests")
		return
	}

	h.sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"logs":  logs,
			"count": len(logs),
		},
	})
}

// GetStats returns aggregated statistics
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats, err := h.db.GetStats(r.Context())
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get statistics")
		return
	}

	h.sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// Health returns the health status of the service
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	h.sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		},
	})
}

// Helper functions

func (h *Handler) logRequest(ctx context.Context, userAgent, agentType, ip, endpoint string) error {
	if h.db == nil {
		return nil
	}

	log := database.RequestLog{
		UserAgent:   userAgent,
		AgentType:   agentType,
		RequestedAt: time.Now(),
		IPAddress:   ip,
		Endpoint:    endpoint,
	}

	_, err := h.db.LogRequest(ctx, log)
	return err
}

func (h *Handler) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but don't expose internal details
		http.Error(w, `{"success":false,"error":"encoding error"}`, http.StatusInternalServerError)
	}
}

func (h *Handler) sendError(w http.ResponseWriter, statusCode int, message string) {
	// Sanitize error message to prevent information leakage
	sanitizedMessage := sanitizeErrorMessage(message)

	h.sendJSON(w, statusCode, ErrorResponse{
		Success: false,
		Error:   sanitizedMessage,
	})
}

// Input validation and sanitization functions

// sanitizeIP validates and sanitizes IP addresses
func sanitizeIP(ip string) string {
	// Parse and validate IP
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "unknown"
	}
	return parsed.String()
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (validate it first)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// sanitizeErrorMessage ensures error messages don't leak sensitive information
func sanitizeErrorMessage(message string) string {
	// HTML escape to prevent XSS
	message = html.EscapeString(message)

	// Remove common sensitive patterns
	sensitivePatterns := []string{
		"/home/",
		"/usr/",
		"/var/",
		"password",
		"token",
		"secret",
		"key",
	}

	lowerMessage := strings.ToLower(message)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerMessage, pattern) {
			return "an error occurred"
		}
	}

	// Limit message length
	if len(message) > 200 {
		return "an error occurred"
	}

	return message
}

// RateLimitMiddleware provides basic rate limiting
func RateLimitMiddleware(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	type client struct {
		requests  int
		lastReset time.Time
	}

	clients := make(map[string]*client)
	var mu sync.Mutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			mu.Lock()
			c, exists := clients[ip]
			now := time.Now()

			if !exists || now.Sub(c.lastReset) > window {
				clients[ip] = &client{
					requests:  1,
					lastReset: now,
				}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if c.requests >= maxRequests {
				mu.Unlock()
				http.Error(w, `{"success":false,"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			c.requests++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
