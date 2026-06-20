package organizations

import (
	"testing"

	"aiglasses/server/internal/platform/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupOrgTest(t *testing.T) (*gorm.DB, *Service) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&database.Organization{}); err != nil {
		t.Fatal(err)
	}
	return db, NewService(db)
}

func TestTreeIncludesDeepChildren(t *testing.T) {
	db, svc := setupOrgTest(t)
	orgs := []database.Organization{
		{Code: "ROOT", Name: "根单位", Status: StatusActive},
		{Code: "A", Name: "一级单位", ParentCode: "ROOT", Status: StatusActive},
		{Code: "A-1", Name: "二级单位", ParentCode: "A", Status: StatusActive},
		{Code: "A-1-1", Name: "三级单位", ParentCode: "A-1", Status: StatusActive},
	}
	if err := db.Create(&orgs).Error; err != nil {
		t.Fatal(err)
	}

	tree, err := svc.Tree()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 1 || tree[0].Code != "ROOT" {
		t.Fatalf("expected ROOT as only root, got %#v", tree)
	}
	if len(tree[0].Children) != 1 || tree[0].Children[0].Code != "A" {
		t.Fatalf("expected ROOT -> A, got %#v", tree[0].Children)
	}
	level2 := tree[0].Children[0].Children
	if len(level2) != 1 || level2[0].Code != "A-1" {
		t.Fatalf("expected ROOT -> A -> A-1, got %#v", level2)
	}
	level3 := level2[0].Children
	if len(level3) != 1 || level3[0].Code != "A-1-1" {
		t.Fatalf("expected deep child A-1-1, got %#v", level3)
	}
}

func TestTreeDoesNotDropOrphanOrganizations(t *testing.T) {
	db, svc := setupOrgTest(t)
	orgs := []database.Organization{
		{Code: "ROOT", Name: "根单位", Status: StatusActive},
		{Code: "ORPHAN", Name: "孤儿单位", ParentCode: "MISSING", Status: StatusActive},
	}
	if err := db.Create(&orgs).Error; err != nil {
		t.Fatal(err)
	}

	tree, err := svc.Tree()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, org := range tree {
		seen[org.Code] = true
	}
	if !seen["ROOT"] || !seen["ORPHAN"] {
		t.Fatalf("expected root and orphan organizations to be visible, got %#v", tree)
	}
}
