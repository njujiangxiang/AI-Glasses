package glass_api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aiglasses/server/internal/auth"
	"github.com/gin-gonic/gin"
)

func setupAuthOnlyRouter(t *testing.T) (*gin.Engine, *auth.Service) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	authSvc := auth.NewService(nil, "test-secret", time.Hour)
	router := gin.New()
	NewHandler(authSvc, nil, nil).Register(router)
	return router, authSvc
}

func TestProtectedRoutesRequireBearerToken(t *testing.T) {
	router, _ := setupAuthOnlyRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d: %s", w.Code, w.Body.String())
	}
}

func TestProtectedRoutesRejectAdminScopeToken(t *testing.T) {
	router, authSvc := setupAuthOnlyRouter(t)
	token, err := authSvc.IssueAccessToken(1, nil, auth.ScopeAdmin)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/my", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for admin token, got %d: %s", w.Code, w.Body.String())
	}
}

func TestParseAttachmentIDs(t *testing.T) {
	ids, err := parseAttachmentIDs("1,2, 3")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Fatalf("unexpected ids: %#v", ids)
	}
}
