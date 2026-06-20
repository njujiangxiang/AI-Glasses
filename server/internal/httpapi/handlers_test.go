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
	"aiglasses/server/internal/datascope"
	"aiglasses/server/internal/monitoring"
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/rbac"
	userssvc "aiglasses/server/internal/users"
)

type testEnv struct {
	router     *gin.Engine
	token      string
	db         *gorm.DB
	bcSvc      *businesscodes.Service
	monitorHub *monitoring.Hub
	minired    *miniredis.Miniredis
}

func setupHandlerTest(t *testing.T) *testEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.Role{}, &database.Permission{}, &database.RolePermission{}, &database.Organization{}, &database.BusinessCode{}); err != nil {
		t.Fatal(err)
	}

	org := database.Organization{Code: "ROOT", Name: "默认单位", Status: "active"}
	if err := db.Create(&org).Error; err != nil {
		t.Fatal(err)
	}

	adminRole := database.Role{Name: "系统管理员", Code: "admin", DataScope: database.DataScopeAll, Status: "active"}
	if err := db.Create(&adminRole).Error; err != nil {
		t.Fatal(err)
	}
	monitorPerm := database.Permission{Name: "实时监控查看", Code: "monitor:view", Perms: rbac.MonitorViewPerm, Type: "A", Visible: true, Status: "active"}
	if err := db.Create(&monitorPerm).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&database.RolePermission{RoleID: adminRole.ID, PermissionID: monitorPerm.ID}).Error; err != nil {
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
		Name:         "系统管理员",
		DisplayName:  "系统管理员",
		OrgCode:      org.Code,
		RoleID:       adminRole.ID,
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

	datascopeSvc := datascope.NewService(db)
	monitorHub := monitoring.NewHub(monitoring.WithMaxEntries(3))
	handler := NewHandler(
		authSvc,
		nil, // attachments
		bcSvc,
		datascopeSvc,
		nil, // defects
		nil, // devices
		nil, // menus
		nil, // organizations
		nil, // plans
		nil, // roles
		nil, // tasks
		nil, // templates
		userssvc.NewService(db),
		nil, // workflows
		nil, // scheduler
		rbac.NewService(db),
		monitorHub,
	)

	router := gin.New()
	handler.Register(router)

	return &testEnv{
		router:     router,
		token:      token,
		db:         db,
		bcSvc:      bcSvc,
		monitorHub: monitorHub,
		minired:    mr,
	}
}

func TestCreateBusinessCode(t *testing.T) {
	env := setupHandlerTest(t)

	body := map[string]any{
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

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if data, ok := resp["data"].(map[string]any); ok {
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
	body := map[string]any{
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

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if data, ok := resp["data"].([]any); ok {
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

	body := map[string]any{
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

	body := map[string]any{
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

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if data, ok := resp["data"].(map[string]any); ok {
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

	body := map[string]any{
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

func TestUpdateCurrentUserPreservesAdminFields(t *testing.T) {
	env := setupHandlerTest(t)

	body := map[string]any{
		"name":        "新姓名",
		"gender":      "female",
		"birth_year":  1990,
		"birth_month": 6,
		"id_card_no":  "11010119900601002X",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/admin/users/me/update", bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "Bearer "+env.token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var stored database.User
	if err := env.db.Where("username = ?", "admin").First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Name != "新姓名" || stored.Gender != "female" {
		t.Fatalf("profile fields not updated: name=%s gender=%s", stored.Name, stored.Gender)
	}
	if stored.OrgCode != "ROOT" {
		t.Fatalf("expected org_code preserved as ROOT, got %q", stored.OrgCode)
	}
	if stored.RoleID == 0 {
		t.Fatal("expected role_id preserved, got 0")
	}
	if stored.Status != "active" {
		t.Fatalf("expected status preserved as active, got %q", stored.Status)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	data := resp["data"].(map[string]any)
	if data["org_name"] != "默认单位" || data["company_name"] != "默认单位" {
		t.Fatalf("expected org names in response, got %#v", data)
	}
}

func TestCurrentUserReturnsOrganizationName(t *testing.T) {
	env := setupHandlerTest(t)

	req := httptest.NewRequest("GET", "/api/admin/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	data := resp["data"].(map[string]any)
	if data["org_name"] != "默认单位" {
		t.Fatalf("expected org_name 默认单位, got %#v", data["org_name"])
	}
}

func TestCurrentUserDoesNotRequireDataScopeRole(t *testing.T) {
	env := setupHandlerTest(t)
	if err := env.db.Model(&database.User{}).Where("username = ?", "admin").Update("role_id", 0).Error; err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/api/admin/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
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

func TestRecentMonitorLogsAllowsAdminWithPermission(t *testing.T) {
	env := setupHandlerTest(t)
	env.monitorHub.Append("LOG", "test", "hello")

	req := httptest.NewRequest("GET", "/api/admin/monitoring/logs/recent?limit=200", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	data := resp["data"].(map[string]any)
	if data["stream_id"] == "" || data["entries"] == nil {
		t.Fatalf("expected stream_id and entries, got %#v", data)
	}
}

func TestRecentMonitorLogsRequiresAuth(t *testing.T) {
	env := setupHandlerTest(t)
	req := httptest.NewRequest("GET", "/api/admin/monitoring/logs/recent", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRecentMonitorLogsRejectsNonSuperAdminWithPermission(t *testing.T) {
	env := setupHandlerTest(t)
	role := database.Role{Name: "普通管理员", Code: "normal", DataScope: database.DataScopeAll, Status: "active"}
	if err := env.db.Create(&role).Error; err != nil {
		t.Fatal(err)
	}
	var perm database.Permission
	if err := env.db.Where("perms = ?", rbac.MonitorViewPerm).First(&perm).Error; err != nil {
		t.Fatal(err)
	}
	if err := env.db.Create(&database.RolePermission{RoleID: role.ID, PermissionID: perm.ID}).Error; err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("normal"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := database.User{Username: "normal", PasswordHash: string(hash), Name: "普通管理员", RoleID: role.ID, Status: "active"}
	if err := env.db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	authSvc := auth.NewService(env.db, "test-secret", time.Hour)
	token, err := authSvc.IssueAccessToken(user.ID, nil, auth.ScopeAdmin)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/api/admin/monitoring/logs/recent", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRecentMonitorLogsAfterIDLimitGapAndInvalidAfterID(t *testing.T) {
	env := setupHandlerTest(t)
	for i := 0; i < 5; i++ {
		env.monitorHub.Append("LOG", "test", "line")
	}

	req := httptest.NewRequest("GET", "/api/admin/monitoring/logs/recent?limit=999999&after_id=1", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	data := resp["data"].(map[string]any)
	if data["gap"] != true {
		t.Fatalf("expected gap=true, got %#v", data)
	}
	entries := data["entries"].([]any)
	if len(entries) != 3 {
		t.Fatalf("expected retained 3 entries, got %d", len(entries))
	}

	req = httptest.NewRequest("GET", "/api/admin/monitoring/logs/recent?after_id=bad", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d: %s", w.Code, w.Body.String())
	}
}
