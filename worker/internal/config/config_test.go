package config

import (
	"testing"
)

func TestIsSovereignCompliant(t *testing.T) {
	tests := []struct {
		name          string
		sovereignMode bool
		dbHost        string
		expected      bool
	}{
		{
			name:          "Not in sovereign mode",
			sovereignMode: false,
			dbHost:        "database.us-east-1.amazonaws.com",
			expected:      true,
		},
		{
			name:          "Sovereign mode with AU host",
			sovereignMode: true,
			dbHost:        "database.ap-southeast-2.amazonaws.com",
			expected:      true,
		},
		{
			name:          "Sovereign mode with localhost",
			sovereignMode: true,
			dbHost:        "localhost",
			expected:      true,
		},
		{
			name:          "Sovereign mode with 127.0.0.1",
			sovereignMode: true,
			dbHost:        "127.0.0.1",
			expected:      true,
		},
		{
			name:          "Sovereign mode with internal cluster DNS",
			sovereignMode: true,
			dbHost:        "lexure-mvp-postgres",
			expected:      true,
		},
		{
			name:          "Sovereign mode with svc.cluster.local",
			sovereignMode: true,
			dbHost:        "lexure-mvp-postgres.lexure.svc.cluster.local",
			expected:      true,
		},
		{
			name:          "Sovereign mode with US host (Non-compliant)",
			sovereignMode: true,
			dbHost:        "database.us-east-1.amazonaws.com",
			expected:      false,
		},
		{
			name:          "Sovereign mode with generic external host (Non-compliant)",
			sovereignMode: true,
			dbHost:        "external-database.com",
			expected:      false,
		},
		{
			name:          "Sovereign mode with Azure host",
			sovereignMode: true,
			dbHost:        "database.australiaeast.database.azure.com",
			expected:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				SovereignMode: tc.sovereignMode,
				Database: DatabaseConfig{
					Host: tc.dbHost,
				},
			}
			result := cfg.IsSovereignCompliant()
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for dbHost: %s with SovereignMode: %v", tc.expected, result, tc.dbHost, tc.sovereignMode)
			}
		})
	}
}
