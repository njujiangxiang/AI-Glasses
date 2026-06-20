package database

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAutoMigrateAddsDataScopeToExistingRolesTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`CREATE TABLE roles (
		id integer PRIMARY KEY AUTOINCREMENT,
		name text NOT NULL UNIQUE,
		created_at datetime,
		updated_at datetime
	)`).Error; err != nil {
		t.Fatalf("create old roles table: %v", err)
	}

	if db.Migrator().HasColumn(&Role{}, "data_scope") {
		t.Fatal("test setup invalid: old roles table already has data_scope")
	}

	if err := AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	if !db.Migrator().HasColumn(&Role{}, "data_scope") {
		t.Fatal("expected AutoMigrate to add roles.data_scope")
	}
}
