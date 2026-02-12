package database

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// RunMigrations runs all database migrations in order
func RunMigrations(db *gorm.DB) error {
	// Get list of migration files
	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("failed to glob migration files: %w", err)
	}

	// Sort files to ensure correct order
	sort.Strings(files)

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Run each migration
	for _, file := range files {
		if err := runMigration(db, file); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func createMigrationsTable(db *gorm.DB) error {
	sql := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id SERIAL PRIMARY KEY,
		version VARCHAR(255) NOT NULL UNIQUE,
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	`
	return db.Exec(sql).Error
}

// runMigration runs a single migration file
func runMigration(db *gorm.DB, filePath string) error {
	// Extract version from filename
	version := strings.TrimSuffix(filepath.Base(filePath), ".up.sql")

	// Check if migration already applied
	var count int64
	db.Table("schema_migrations").Where("version = ?", version).Count(&count)
	if count > 0 {
		return nil // Already applied
	}

	// Read migration file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Parse SQL statements more intelligently to handle PostgreSQL functions
	statements := parseSQLStatements(string(content))

	// Execute each statement
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		// Try to execute the statement
		if err := db.Exec(statement).Error; err != nil {
			// Check if it's a "relation already exists" error
			if strings.Contains(err.Error(), "already exists") {
				// Log but continue - table/index already exists
				continue
			}
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	// Record migration as applied
	if err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version).Error; err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(db *gorm.DB) ([]MigrationStatus, error) {
	var migrations []MigrationStatus

	err := db.Table("schema_migrations").
		Select("version, applied_at").
		Order("applied_at ASC").
		Find(&migrations).Error

	return migrations, err
}

// MigrationStatus represents a migration status
type MigrationStatus struct {
	Version   string `json:"version"`
	AppliedAt string `json:"applied_at"`
}

// SeedData seeds the database with initial data
func SeedData(db *gorm.DB) error {
	// Create a default company for testing
	company := Company{
		Name:   "Demo Company",
		Domain: "demo.com",
		Status: "active",
		AlertSettings: map[string]interface{}{
			"email_enabled":   true,
			"sms_enabled":     false,
			"alert_threshold": 30, // seconds
		},
		RetrySettings: map[string]interface{}{
			"immediate_retry": true,
			"max_attempts":    3,
			"retry_delay":     300, // seconds
		},
	}

	// Check if company already exists
	var existing Company
	if err := db.Where("name = ?", company.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
		if err := db.Create(&company).Error; err != nil {
			return fmt.Errorf("failed to create demo company: %w", err)
		}
	}

	return nil
}

// Company represents a company for seeding
type Company struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Domain        string                 `json:"domain"`
	Status        string                 `json:"status"`
	AlertSettings map[string]interface{} `json:"alert_settings"`
	RetrySettings map[string]interface{} `json:"retry_settings"`
}

// parseSQLStatements parses SQL content into individual statements, handling PostgreSQL functions properly
func parseSQLStatements(content string) []string {
	var statements []string
	var currentStatement strings.Builder
	var inFunction bool
	var dollarQuoteCount int

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if we're entering a function definition
		if strings.Contains(strings.ToUpper(trimmedLine), "CREATE OR REPLACE FUNCTION") ||
			strings.Contains(strings.ToUpper(trimmedLine), "CREATE FUNCTION") {
			inFunction = true
			dollarQuoteCount = 0
		}

		// Count dollar quotes to track function body
		if inFunction {
			dollarQuoteCount += strings.Count(trimmedLine, "$$")
		}

		// Add line to current statement
		currentStatement.WriteString(line)
		currentStatement.WriteString("\n")

		// Check if statement is complete
		if !inFunction && strings.TrimSpace(line) != "" && strings.HasSuffix(strings.TrimSpace(line), ";") {
			// Regular statement ending with semicolon
			statements = append(statements, currentStatement.String())
			currentStatement.Reset()
		} else if inFunction && dollarQuoteCount%2 == 0 && dollarQuoteCount > 0 && strings.TrimSpace(line) != "" {
			// Function body complete (even number of dollar quotes means we've closed the function)
			statements = append(statements, currentStatement.String())
			currentStatement.Reset()
			inFunction = false
		}
	}

	// Add any remaining statement
	if currentStatement.Len() > 0 {
		statements = append(statements, currentStatement.String())
	}

	return statements
}
