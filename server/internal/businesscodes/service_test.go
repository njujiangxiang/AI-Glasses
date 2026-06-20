package businesscodes

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "time/tzdata"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
)

type testEnv struct {
	db    *gorm.DB
	redis *miniredis.Miniredis
	svc   *Service
}

func setupTest(t *testing.T) *testEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&database.BusinessCode{}, &database.Organization{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&database.Organization{Code: "ROOT", Name: "默认单位", Status: "active"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&database.Organization{Code: "DISABLED", Name: "停用单位", Status: "disabled"}).Error; err != nil {
		t.Fatal(err)
	}
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc, err := NewService(db, client)
	if err != nil {
		t.Fatal(err)
	}
	fixedNow := time.Date(2026, 6, 16, 10, 0, 0, 0, svc.location)
	svc.SetNowForTest(func() time.Time { return fixedNow })
	return &testEnv{db: db, redis: mr, svc: svc}
}

func TestCreateValidation(t *testing.T) {
	env := setupTest(t)

	// missing name
	_, err := env.svc.Create(Input{Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if !isValidationError(err) {
		t.Errorf("expected validation error for missing name, got %v", err)
	}

	// invalid code
	_, err = env.svc.Create(Input{Name: "Test", Code: "tk:1", DateFormat: DateFormatDaily, SeqPadding: 4})
	if !isValidationError(err) {
		t.Errorf("expected validation error for invalid code, got %v", err)
	}

	// invalid date format
	_, err = env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: "yyyy", SeqPadding: 4})
	if !isValidationError(err) {
		t.Errorf("expected validation error for invalid date format, got %v", err)
	}

	// seq padding out of range
	_, err = env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 0})
	if !isValidationError(err) {
		t.Errorf("expected validation error for zero seq padding, got %v", err)
	}

	_, err = env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 13})
	if !isValidationError(err) {
		t.Errorf("expected validation error for seq padding 13, got %v", err)
	}

	// use separator but empty separator
	_, err = env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4, UseSeparator: true, Separator: ""})
	if !isValidationError(err) {
		t.Errorf("expected validation error for use_separator with empty separator, got %v", err)
	}

	// code normalization
	rule, err := env.svc.Create(Input{Name: "Test", Code: " tk ", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}
	if rule.Code != "TK" {
		t.Errorf("expected code TK, got %s", rule.Code)
	}

	// duplicate code
	_, err = env.svc.Create(Input{Name: "Test2", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if !isValidationError(err) {
		t.Errorf("expected validation error for duplicate code, got %v", err)
	}
}

func TestUpdateValidation(t *testing.T) {
	env := setupTest(t)

	rule, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	// update with empty name should fail validation
	_, err = env.svc.Update(rule.ID, Input{Name: "", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if !isValidationError(err) {
		t.Errorf("expected validation error for empty name, got %v", err)
	}

	// update with invalid seq padding should fail validation
	_, err = env.svc.Update(rule.ID, Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 0})
	if !isValidationError(err) {
		t.Errorf("expected validation error for zero seq padding, got %v", err)
	}

	// Code field in input is ignored — even invalid codes are accepted without error
	_, err = env.svc.Update(rule.ID, Input{Name: "Test", Code: "tk:2", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Errorf("expected no error when Code is ignored, got %v", err)
	}

	// successful update
	updated, err := env.svc.Update(rule.ID, Input{Name: "Updated", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 5})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Updated" || updated.SeqPadding != 5 {
		t.Errorf("update did not apply: got name=%s, seq_padding=%d", updated.Name, updated.SeqPadding)
	}
}

func TestEnableDisableDelete(t *testing.T) {
	env := setupTest(t)

	rule, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	// disable
	if err := env.svc.Disable(rule.ID); err != nil {
		t.Fatal(err)
	}
	var disabled database.BusinessCode
	if err := env.db.First(&disabled, rule.ID).Error; err != nil {
		t.Fatal(err)
	}
	if disabled.Status != StatusDisabled {
		t.Errorf("expected disabled status, got %s", disabled.Status)
	}

	// enable
	if err := env.svc.Enable(rule.ID); err != nil {
		t.Fatal(err)
	}
	var enabled database.BusinessCode
	if err := env.db.First(&enabled, rule.ID).Error; err != nil {
		t.Fatal(err)
	}
	if enabled.Status != StatusActive {
		t.Errorf("expected active status, got %s", enabled.Status)
	}

	// delete
	if err := env.svc.Delete(rule.ID); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := env.db.Model(&database.BusinessCode{}).Where("id = ?", rule.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 records after delete, got %d", count)
	}
}

func TestGenerateDailyHappyPath(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	// Fix the test date to 2026-06-16
	testDate := time.Date(2026, 6, 16, 10, 0, 0, 0, env.svc.location)
	env.svc.SetNowForTest(func() time.Time { return testDate })

	ctx := context.Background()
	code1, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code1 != "TK202606160001" {
		t.Errorf("expected TK202606160001, got %s", code1)
	}

	code2, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code2 != "TK202606160002" {
		t.Errorf("expected TK202606160002, got %s", code2)
	}

	// verify Redis key
	key := "BNO:TK:20260616"
	if !env.redis.Exists(key) {
		t.Error("expected Redis key to exist")
	}
	ttl := env.redis.TTL(key)
	if ttl < 47*time.Hour || ttl > 48*time.Hour {
		t.Errorf("expected TTL around 48h, got %v", ttl)
	}
}

func TestGenerateDailyWithSeparator(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4, UseSeparator: true, Separator: "-"})
	if err != nil {
		t.Fatal(err)
	}

	// Fix the test date to 2026-06-16
	testDate := time.Date(2026, 6, 16, 10, 0, 0, 0, env.svc.location)
	env.svc.SetNowForTest(func() time.Time { return testDate })

	ctx := context.Background()
	code, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code != "TK-20260616-0001" {
		t.Errorf("expected TK-20260616-0001, got %s", code)
	}
}

func TestGenerateDailyWithoutDateUsesGlobalSequence(t *testing.T) {
	env := setupTest(t)
	useDate := false

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", UseDate: &useDate, DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	code1, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code1 != "TK0001" {
		t.Errorf("expected TK0001, got %s", code1)
	}

	env.svc.SetNowForTest(func() time.Time { return time.Date(2026, 6, 17, 10, 0, 0, 0, env.svc.location) })
	code2, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code2 != "TK0002" {
		t.Errorf("expected TK0002 after date rollover, got %s", code2)
	}
	if !env.redis.Exists("BNO:TK:global") {
		t.Error("expected global Redis key to exist")
	}
	if ttl := env.redis.TTL("BNO:TK:global"); ttl != 0 {
		t.Errorf("expected global key without TTL, got %v", ttl)
	}
}

func TestGenerateDailyWithoutDateWithOrgAndSeparator(t *testing.T) {
	env := setupTest(t)
	useDate := false

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", UseDate: &useDate, DateFormat: DateFormatDaily, SeqPadding: 4, UseSeparator: true, Separator: "-", UseOrgCode: true, OrgCode: "ROOT"})
	if err != nil {
		t.Fatal(err)
	}

	code, err := env.svc.GenerateDaily(context.Background(), "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code != "TK-ROOT-0001" {
		t.Errorf("expected TK-ROOT-0001, got %s", code)
	}
	if !env.redis.Exists("BNO:TK:ROOT:global") {
		t.Error("expected org global Redis key to exist")
	}
	if ttl := env.redis.TTL("BNO:TK:ROOT:global"); ttl != 0 {
		t.Errorf("expected org global key without TTL, got %v", ttl)
	}
}

func TestGenerateDailyWithOrgCode(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4, UseOrgCode: true, OrgCode: "ROOT"})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	code, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code != "TKROOT202606160001" {
		t.Errorf("expected TKROOT202606160001, got %s", code)
	}
	if !env.redis.Exists("BNO:TK:20260616:ROOT") {
		t.Error("expected org daily Redis key to exist")
	}
}

func TestGenerateDailyWithOrgCodeAndSeparator(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4, UseSeparator: true, Separator: "-", UseOrgCode: true, OrgCode: "ROOT"})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	code, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code != "TK-ROOT-20260616-0001" {
		t.Errorf("expected TK-ROOT-20260616-0001, got %s", code)
	}
}

func TestBusinessCodeOrgValidation(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4, UseOrgCode: true})
	if !isValidationError(err) {
		t.Errorf("expected validation error for missing org code, got %v", err)
	}

	_, err = env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4, UseOrgCode: true, OrgCode: "DISABLED"})
	if !isValidationError(err) {
		t.Errorf("expected validation error for disabled org code, got %v", err)
	}
}

func TestGenerateDailyRejectsDisabledOrgWithoutConsumingSequence(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4, UseOrgCode: true, OrgCode: "ROOT"})
	if err != nil {
		t.Fatal(err)
	}
	if err := env.db.Model(&database.Organization{}).Where("code = ?", "ROOT").Update("status", "disabled").Error; err != nil {
		t.Fatal(err)
	}

	_, err = env.svc.GenerateDaily(context.Background(), "TK")
	if !isValidationError(err) {
		t.Fatalf("expected validation error for disabled org during generation, got %v", err)
	}
	if env.redis.Exists("BNO:TK:20260616") {
		t.Fatal("expected no Redis sequence consumption when org is disabled")
	}
}

func TestUpdateDisablesOrgCode(t *testing.T) {
	env := setupTest(t)

	rule, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4, UseOrgCode: true, OrgCode: "ROOT"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := env.svc.Update(rule.ID, Input{Name: "Test", DateFormat: DateFormatDaily, SeqPadding: 4, UseOrgCode: false, OrgCode: "ROOT"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.UseOrgCode || updated.OrgCode != "" {
		t.Fatalf("expected org code disabled and cleared, got use=%v org=%q", updated.UseOrgCode, updated.OrgCode)
	}
	code, err := env.svc.GenerateDaily(context.Background(), "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code != "TK202606160001" {
		t.Errorf("expected TK202606160001, got %s", code)
	}
}

func TestGenerateDailyShortDateFormat(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatShort, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	// Fix the test date to 2026-06-16
	testDate := time.Date(2026, 6, 16, 10, 0, 0, 0, env.svc.location)
	env.svc.SetNowForTest(func() time.Time { return testDate })

	ctx := context.Background()
	code, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code != "TK2606160001" {
		t.Errorf("expected TK2606160001, got %s", code)
	}

	// with separator
	_, err = env.svc.Create(Input{Name: "Test2", Code: "DF", DateFormat: DateFormatShort, SeqPadding: 3, UseSeparator: true, Separator: "-"})
	if err != nil {
		t.Fatal(err)
	}
	code2, err := env.svc.GenerateDaily(ctx, "DF")
	if err != nil {
		t.Fatal(err)
	}
	if code2 != "DF-260616-001" {
		t.Errorf("expected DF-260616-001, got %s", code2)
	}
}

func TestGenerateDailyDayRollover(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 6, 16, 10, 0, 0, 0, env.svc.location)
	env.svc.SetNowForTest(func() time.Time { return now })

	ctx := context.Background()
	code1, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code1 != "TK202606160001" {
		t.Errorf("expected TK202606160001, got %s", code1)
	}

	// advance to next day
	now = time.Date(2026, 6, 17, 10, 0, 0, 0, env.svc.location)
	code2, err := env.svc.GenerateDaily(ctx, "TK")
	if err != nil {
		t.Fatal(err)
	}
	if code2 != "TK202606170001" {
		t.Errorf("expected TK202606170001, got %s", code2)
	}

	// verify both keys exist
	if !env.redis.Exists("BNO:TK:20260616") {
		t.Error("expected day 1 key to exist")
	}
	if !env.redis.Exists("BNO:TK:20260617") {
		t.Error("expected day 2 key to exist")
	}
}

func TestGenerateDailyMissingConfig(t *testing.T) {
	env := setupTest(t)

	ctx := context.Background()
	_, err := env.svc.GenerateDaily(ctx, "MISSING")
	if !isValidationError(err) {
		t.Errorf("expected validation error for missing config, got %v", err)
	}
}

func TestGenerateDailyDisabled(t *testing.T) {
	env := setupTest(t)

	rule, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	if err := env.svc.Disable(rule.ID); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = env.svc.GenerateDaily(ctx, "TK")
	if !isValidationError(err) {
		t.Errorf("expected validation error for disabled status, got %v", err)
	}
}

func TestGenerateDailyOverflow(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 2})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	// generate up to max (99)
	for range 99 {
		_, err := env.svc.GenerateDaily(ctx, "TK")
		if err != nil {
			t.Fatal(err)
		}
	}

	// next should overflow
	_, err = env.svc.GenerateDaily(ctx, "TK")
	if !isValidationError(err) {
		t.Errorf("expected validation error for overflow, got %v", err)
	}
}

func TestGenerateDailyRedisUnavailable(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	env.redis.Close()

	ctx := context.Background()
	_, err = env.svc.GenerateDaily(ctx, "TK")
	if !isInternalError(err) {
		t.Errorf("expected internal error for Redis unavailable, got %v", err)
	}
}

func TestListKeywordFilter(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Task", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.svc.Create(Input{Name: "Defect", Code: "DF", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	all, err := env.svc.List("")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 items, got %d", len(all))
	}

	filtered, err := env.svc.List("Task")
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 {
		t.Errorf("expected 1 item for keyword Task, got %d", len(filtered))
	}
	if filtered[0].Code != "TK" {
		t.Errorf("expected TK, got %s", filtered[0].Code)
	}
}

func TestListLikeWildcardEscaping(t *testing.T) {
	env := setupTest(t)

	_, err := env.svc.Create(Input{Name: "Task100%", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.svc.Create(Input{Name: "Defect", Code: "DF", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	// Searching for literal '%' should only match the record containing '%'.
	results, err := env.svc.List("%")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 item for literal %% keyword, got %d", len(results))
	}

	// Searching for '_' should not match any record (no name contains underscore).
	results, err = env.svc.List("_")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 items for literal _ keyword, got %d", len(results))
	}
}

func TestUpdateCodeImmutable(t *testing.T) {
	env := setupTest(t)

	rule, err := env.svc.Create(Input{Name: "Test", Code: "TK", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}

	// Attempt to change Code via Update should be ignored.
	updated, err := env.svc.Update(rule.ID, Input{Name: "Updated", Code: "NEWCODE", DateFormat: DateFormatDaily, SeqPadding: 4})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Code != "TK" {
		t.Errorf("expected Code to remain TK, got %s", updated.Code)
	}
	if updated.Name != "Updated" {
		t.Errorf("expected Name to be Updated, got %s", updated.Name)
	}
}

func isValidationError(err error) bool {
	apiErr, ok := err.(httperr.APIError)
	return ok && apiErr.Code == httperr.ValidationFailed
}

func isInternalError(err error) bool {
	apiErr, ok := err.(httperr.APIError)
	return ok && apiErr.Code == httperr.InternalError
}
