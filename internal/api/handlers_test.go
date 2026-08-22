package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/baditaflorin/commonuseragent/internal/database"
)

type fakeDB struct {
	logs   []database.RequestLog
	recent []database.RequestLog
	stats  *database.Stats
	err    error
}

func (db *fakeDB) LogRequest(_ context.Context, log database.RequestLog) (int64, error) {
	db.logs = append(db.logs, log)
	return int64(len(db.logs)), db.err
}

func (db *fakeDB) GetRecentRequests(_ context.Context, _ int) ([]database.RequestLog, error) {
	return db.recent, db.err
}

func (db *fakeDB) GetRequestsByType(_ context.Context, _ string, _ int) ([]database.RequestLog, error) {
	return nil, db.err
}

func (db *fakeDB) GetStats(_ context.Context) (*database.Stats, error) {
	return db.stats, db.err
}

func (db *fakeDB) DeleteOldRequests(_ context.Context, _ time.Duration) (int64, error) {
	return 0, db.err
}

func TestGetRandomDesktopReturnsAndLogsUserAgent(t *testing.T) {
	db := &fakeDB{}
	handler := NewHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/api/desktop", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	recorder := httptest.NewRecorder()

	handler.GetRandomDesktop(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			UserAgent string `json:"userAgent"`
			Type      string `json:"type"`
		} `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.Success || response.Data.Type != "desktop" || response.Data.UserAgent == "" {
		t.Errorf("unexpected response: %+v", response)
	}
	if len(db.logs) != 1 {
		t.Fatalf("logged requests = %d, want 1", len(db.logs))
	}
	if got := db.logs[0]; got.AgentType != "desktop" || got.IPAddress != "203.0.113.10" || got.Endpoint != "/api/desktop" {
		t.Errorf("unexpected request log: %+v", got)
	}
}

func TestMonitoringHandlersRequireDatabase(t *testing.T) {
	handler := NewHandler(nil)

	for name, endpoint := range map[string]http.HandlerFunc{
		"logs":  handler.GetRecentRequests,
		"stats": handler.GetStats,
	} {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			endpoint(recorder, httptest.NewRequest(http.MethodGet, "/api/"+name, nil))

			if recorder.Code != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
			}
			if !strings.Contains(recorder.Body.String(), "request logging is unavailable") {
				t.Errorf("response = %q, want availability error", recorder.Body.String())
			}
		})
	}
}

func TestGetRecentRequestsRejectsInvalidLimit(t *testing.T) {
	handler := NewHandler(&fakeDB{})
	recorder := httptest.NewRecorder()

	handler.GetRecentRequests(recorder, httptest.NewRequest(http.MethodGet, "/api/logs?limit=1001", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestRateLimitMiddlewareReturnsJSON(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := RateLimitMiddleware(2, time.Minute)(next)

	for i := 0; i < 3; i++ {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

		if i < 2 && recorder.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d", i+1, recorder.Code, http.StatusNoContent)
		}
		if i == 2 {
			if recorder.Code != http.StatusTooManyRequests {
				t.Fatalf("request 3 status = %d, want %d", recorder.Code, http.StatusTooManyRequests)
			}
			if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
				t.Errorf("Content-Type = %q, want application/json", contentType)
			}
		}
	}
}
