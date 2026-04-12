package database

import (
	"regexp"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTenantIDValidation(t *testing.T) {
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	
	testCases := []struct {
		id    string
		valid bool
	}{
		{"tenant123", true},
		{"tenant_456", true},
		{"tenant-789", true},
		{"tenant; DROP TABLE users;", false},
		{"tenant space", false},
		{"tenant$#%", false},
		{"", false},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.valid, re.MatchString(tc.id), "Tenant ID: %s", tc.id)
	}
}

func TestSchemaExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql: %s", err)
	}
	defer db.Close()

	gormDB, _ := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = ?)")).
		WithArgs("tenant_test").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := SchemaExists(gormDB, "tenant_test")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCreateSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql: %s", err)
	}
	defer db.Close()

	gormDB, _ := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})

	mock.ExpectExec("CREATE SCHEMA IF NOT EXISTS tenant_test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = CreateSchema(gormDB, "tenant_test")
	assert.NoError(t, err)
}
