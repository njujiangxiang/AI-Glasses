package monitoring

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGinRequestLoggerCapturesSafeSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewHub()
	r := gin.New()
	r.Use(GinRequestLogger(hub, GinLoggerConfig{}))
	r.GET("/users/:id", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/users/123?token=secret", strings.NewReader(`{"password":"secret"}`))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Cookie", "session=secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	result := hub.Recent(10, 0)
	if len(result.Entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(result.Entries))
	}
	message := result.Entries[0].Message
	if !strings.Contains(message, "method=GET") || !strings.Contains(message, "path=/users/:id") || !strings.Contains(message, "status=200") {
		t.Fatalf("unexpected summary %q", message)
	}
	if strings.Contains(message, "secret") || strings.Contains(message, "token=") || strings.Contains(message, "Authorization") || strings.Contains(message, "Cookie") {
		t.Fatalf("summary leaked raw sensitive input: %q", message)
	}
}

func TestGinRequestLoggerSkipsMonitorEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewHub()
	r := gin.New()
	r.Use(GinRequestLogger(hub, GinLoggerConfig{SkipPaths: map[string]bool{"/api/admin/monitoring/logs/recent": true}}))
	r.GET("/api/admin/monitoring/logs/recent", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/admin/monitoring/logs/recent", nil))

	if result := hub.Recent(10, 0); len(result.Entries) != 0 {
		t.Fatalf("expected monitor endpoint to be skipped, got %d entries", len(result.Entries))
	}
}
