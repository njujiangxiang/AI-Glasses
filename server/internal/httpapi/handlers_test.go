package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "time/tzdata"

	"aiglasses/server/internal/auth"
	"aiglasses/server/internal/businesscodes"
	"aiglasses/server/internal/platform/database"
)

type testEnv struct {
	router  *gin.Engine
	token   string
	bcSvc   *businesscodes.Service
	minired *miniredis.Miniredis
}

func setupHandlerTest(t *testing.T) *testEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.BusinessCode{}); err != nil {
		t.Fatal(err)
	}

	// Create test admin user with bcrypt password
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	adminUser := database.User{
		Username:     "admin",
		PasswordHash: string(hash),
		Status:       "active",
	}
	if err := db.Create(&adminUser).Error; err != nil {
		t.Fatal(err)
	}

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	authSvc := auth.NewService(db, "test-secret", time.Hour)
	bcSvc, err := businesscodes.NewService(db, redisClient)
	if err != nil {
		t.Fatal(err)
	}

	token, err := authSvc.IssueAccessToken(adminUser.ID, nil, auth.ScopeAdmin)
	if err != nil {
		t.Fatal(err)
	}

	handler := NewHandler(
		authSvc,
		nil, // attachments
		bcSvc,
		nil, // defects
		nil, // devices
		nil, // organizations
		nil, // plans
		nil, // tasks
		nil, // templates
		nil, // users
		nil, // scheduler
	)

	router := gin.New()
	handler.Register(router)

	return &testEnv{
		router:  router,
		token:   token,
		bcSvc:   bcSvc,
		minired: mr,
	}
}

func TestCreateBusinessCode(t *testing.T) {
	env := setupHandlerTest(t)

	body := map[string]interface{}{
		"name":          "Task",
		"code":          "TK",
		"date_format":   "yyyyMMdd",
		"seq_padding":   4,
		"separator":     "",
		"use_separator": false,
		"status":        "active",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/admin/business-codes", bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "Bearer "+env.token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if data["code"] != "TK" {
			t.Errorf("expected code TK, got %v", data["code"])
		}
	} else {
		t.Error("expected data in response")
	}
}

func TestCreateBusinessCodeValidationError(t *testing.T) {
	env := setupHandlerTest(t)

	// Invalid code format
	body := map[string]interface{}{
		"name":        "Task",
		"code":        "tk:1",
		"date_format": "yyyyMMdd",
		"seq_padding": 4,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/admin/business-codes", bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "Bearer "+env.token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListBusinessCodes(t *testing.T) {
	env := setupHandlerTest(t)

	// Create a business code first
	env.bcSvc.Create(businesscodes.Input{
		Name:       "Task",
		Code:       "TK",
		DateFormat: "yyyyMMdd",
		SeqPadding: 4,
		Status:     "active",
	})

	req := httptest.NewRequest("GET", "/api/admin/business-codes", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if data, ok := resp["data"].([]interface{}); ok {
		if len(data) != 1 {
			t.Errorf("expected 1 item, got %d", len(data))
		}
	} else {
		t.Error("expected data array in response")
	}
}

func TestUpdateBusinessCode(t *testing.T) {
	env := setupHandlerTest(t)

	// Create a business code first
	created, err := env.bcSvc.Create(businesscodes.Input{
		Name:       "Task",
		Code:       "TK",
		DateFormat: "yyyyMMdd",
		SeqPadding: 4,
		Status:     "active",
	})
	if err != nil {
		t.Fatal(err)
	}

	body := map[string]interface{}{
		"name":          "Task Updated",
		"code":          "TK",
		"date_format":   "yyyyMMdd",
		"seq_padding":   5,
		"separator":     "",
		"use_separator": false,
		"status":        "active",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/admin/business-codes/1/update", bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "Bearer "+env.token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify update
	updated, err := env.bcSvc.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Task Updated" || updated.SeqPadding != 5 {
		t.Errorf("update did not apply: got name=%s, seq_padding=%d", updated.Name, updated.SeqPadding)
	}
}

func TestEnableDisableBusinessCode(t *testing.T) {
	env := setupHandlerTest(t)

	created, err := env.bcSvc.Create(businesscodes.Input{
		Name:       "Task",
		Code:       "TK",
		DateFormat: "yyyyMMdd",
		SeqPadding: 4,
		Status:     "active",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Disable
	req := httptest.NewRequest("POST", "/api/admin/business-codes/1/disable", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for disable, got %d: %s", w.Code, w.Body.String())
	}

	disabled, _ := env.bcSvc.Get(created.ID)
	if disabled.Status != "disabled" {
		t.Errorf("expected disabled status, got %s", disabled.Status)
	}

	// Enable
	req = httptest.NewRequest("POST", "/api/admin/business-codes/1/enable", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)

	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for enable, got %d: %s", w.Code, w.Body.String())
	}

	enabled, _ := env.bcSvc.Get(created.ID)
	if enabled.Status != "active" {
		t.Errorf("expected active status, got %s", enabled.Status)
	}
}

func TestDeleteBusinessCode(t *testing.T) {
	env := setupHandlerTest(t)

	_, err := env.bcSvc.Create(businesscodes.Input{
		Name:       "Task",
		Code:       "TK",
		DateFormat: "yyyyMMdd",
		SeqPadding: 4,
		Status:     "active",
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/admin/business-codes/1/delete", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for delete, got %d: %s", w.Code, w.Body.String())
	}

	// Verify deletion
	_, err = env.bcSvc.Get(1)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestGenerateBusinessCode(t *testing.T) {
	env := setupHandlerTest(t)

	_, err := env.bcSvc.Create(businesscodes.Input{
		Name:       "Task",
		Code:       "TK",
		DateFormat: "yyyyMMdd",
		SeqPadding: 4,
		Status:     "active",
	})
	if err != nil {
		t.Fatal(err)
	}

	body := map[string]interface{}{
		"code": "TK",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/admin/business-codes/generate", bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "Bearer "+env.token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if data, ok := resp["data"].(map[string]interface{}); ok {
		code, ok := data["code"].(string)
		if !ok || code == "" {
			t.Error("expected code in response")
		}
		// Check format: TK202606160001 (2 + 8 + 4 = 14 characters)
		if len(code) != 14 {
			t.Errorf("expected code length 14, got %d: %s", len(code), code)
		}
	} else {
		t.Error("expected data in response")
	}
}

func TestGenerateBusinessCodeMissing(t *testing.T) {
	env := setupHandlerTest(t)

	body := map[string]interface{}{
		"code": "MISSING",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/admin/business-codes/generate", bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "Bearer "+env.token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422 for missing code, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouteRegistration(t *testing.T) {
	env := setupHandlerTest(t)

	// Test that business-code routes are registered under admin group
	req := httptest.NewRequest("GET", "/api/admin/business-codes", nil)
	// Without auth should get 401
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 without auth, got %d", w.Code)
	}
}
