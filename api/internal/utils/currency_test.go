package utils

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDollarsToCents(t *testing.T) {
	tests := []struct {
		name        string
		dollars     float64
		expected    int64
		expectError bool
		errorMsg    string
	}{
		{
			name:     "basic conversion",
			dollars:  10.50,
			expected: 1050,
		},
		{
			name:     "zero amount",
			dollars:  0.00,
			expected: 0,
		},
		{
			name:     "rounding up",
			dollars:  10.995,
			expected: 1100,
		},
		{
			name:     "rounding down",
			dollars:  10.994,
			expected: 1099,
		},
		{
			name:     "large amount",
			dollars:  1000000.00,
			expected: 100000000,
		},
		{
			name:        "negative amount",
			dollars:     -10.50,
			expectError: true,
			errorMsg:    "amount cannot be negative",
		},
		{
			name:        "NaN amount",
			dollars:     math.NaN(),
			expectError: true,
			errorMsg:    "amount must be a finite number",
		},
		{
			name:        "infinity amount",
			dollars:     math.Inf(1),
			expectError: true,
			errorMsg:    "amount must be a finite number",
		},
		{
			name:        "negative infinity",
			dollars:     math.Inf(-1),
			expectError: true,
			errorMsg:    "amount must be a finite number",
		},
		{
			name:        "amount exceeding max",
			dollars:     MaxAmount * 2,
			expectError: true,
			errorMsg:    "exceeds maximum allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DollarsToCents(tt.dollars)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, int64(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCentsToDollars(t *testing.T) {
	tests := []struct {
		name        string
		cents       int64
		expected    float64
		expectError bool
		errorMsg    string
	}{
		{
			name:     "basic conversion",
			cents:    1050,
			expected: 10.50,
		},
		{
			name:     "zero cents",
			cents:    0,
			expected: 0.00,
		},
		{
			name:     "large amount",
			cents:    100000000,
			expected: 1000000.00,
		},
		{
			name:        "negative cents",
			cents:       -1050,
			expectError: true,
			errorMsg:    "cents cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CentsToDollars(tt.cents)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, float64(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name        string
		amount      *float64
		expectError bool
		errorMsg    string
	}{
		{
			name:   "valid amount",
			amount: func() *float64 { v := 10.50; return &v }(),
		},
		{
			name:   "zero amount",
			amount: func() *float64 { v := 0.00; return &v }(),
		},
		{
			name:   "nil amount",
			amount: nil,
		},
		{
			name:        "negative amount",
			amount:      func() *float64 { v := -10.50; return &v }(),
			expectError: true,
			errorMsg:    "amount cannot be negative",
		},
		{
			name:        "NaN amount",
			amount:      func() *float64 { v := math.NaN(); return &v }(),
			expectError: true,
			errorMsg:    "amount must be a finite number",
		},
		{
			name:        "infinity amount",
			amount:      func() *float64 { v := math.Inf(1); return &v }(),
			expectError: true,
			errorMsg:    "amount must be a finite number",
		},
		{
			name:        "amount exceeding max",
			amount:      func() *float64 { v := MaxAmount * 2; return &v }(),
			expectError: true,
			errorMsg:    "exceeds maximum allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCents(t *testing.T) {
	tests := []struct {
		name        string
		cents       int64
		expectError bool
		errorMsg    string
	}{
		{
			name:  "valid cents",
			cents: 1050,
		},
		{
			name:  "zero cents",
			cents: 0,
		},
		{
			name:        "negative cents",
			cents:       -1050,
			expectError: true,
			errorMsg:    "cents cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCents(tt.cents)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSafeDollarsToCents(t *testing.T) {
	tests := []struct {
		name        string
		dollars     *float64
		expected    int64
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid amount",
			dollars:  func() *float64 { v := 10.50; return &v }(),
			expected: 1050,
		},
		{
			name:     "nil amount",
			dollars:  nil,
			expected: 0,
		},
		{
			name:        "negative amount",
			dollars:     func() *float64 { v := -10.50; return &v }(),
			expectError: true,
			errorMsg:    "amount cannot be negative",
		},
		{
			name:        "NaN amount",
			dollars:     func() *float64 { v := math.NaN(); return &v }(),
			expectError: true,
			errorMsg:    "amount must be a finite number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafeDollarsToCents(tt.dollars)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, int64(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormatCents(t *testing.T) {
	tests := []struct {
		name     string
		cents    int64
		expected string
	}{
		{
			name:     "basic formatting",
			cents:    1050,
			expected: "10.50",
		},
		{
			name:     "zero cents",
			cents:    0,
			expected: "0.00",
		},
		{
			name:     "single digit cents",
			cents:    1005,
			expected: "10.05",
		},
		{
			name:     "large amount",
			cents:    123456789,
			expected: "1234567.89",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCents(tt.cents)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundToNearestCent(t *testing.T) {
	tests := []struct {
		name     string
		dollars  float64
		expected float64
	}{
		{
			name:     "already at cent precision",
			dollars:  10.50,
			expected: 10.50,
		},
		{
			name:     "rounding up",
			dollars:  10.995,
			expected: 11.00,
		},
		{
			name:     "rounding down",
			dollars:  10.994,
			expected: 10.99,
		},
		{
			name:     "exactly halfway",
			dollars:  10.995,
			expected: 11.00,
		},
		{
			name:     "very small amount",
			dollars:  0.001,
			expected: 0.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundToNearestCent(tt.dollars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCurrencyConversionError(t *testing.T) {
	err := &CurrencyConversionError{
		Amount:    -10.50,
		Operation: "test_operation",
		Err:       assert.AnError,
	}

	expectedMsg := "currency conversion error during test_operation with amount -10.50: assert.AnError general error for testing"
	assert.Equal(t, expectedMsg, err.Error())
}

// Integration tests for common conversion scenarios
func TestIntegrationScenarios(t *testing.T) {
	t.Run("round trip conversion", func(t *testing.T) {
		originalDollars := 123.456

		// Convert to cents and back
		cents, err := DollarsToCents(originalDollars)
		require.NoError(t, err)

		backToDollars, err := CentsToDollars(cents)
		require.NoError(t, err)

		// Should be rounded to nearest cent
		assert.Equal(t, 123.46, backToDollars)
	})

	t.Run("edge case amounts", func(t *testing.T) {
		testCases := []float64{
			0.001,  // Less than a cent
			0.004,  // Should round down
			0.005,  // Should round up
			0.009,  // Should round up
			99.99,  // Just under 100
			100.00, // Exactly 100
			100.01, // Just over 100
		}

		for _, amount := range testCases {
			t.Run(fmt.Sprintf("amount_%.3f", amount), func(t *testing.T) {
				cents, err := DollarsToCents(amount)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, cents, int64(0))

				// Convert back and verify it's properly rounded
				backToDollars, err := CentsToDollars(cents)
				assert.NoError(t, err)
				assert.Equal(t, RoundToNearestCent(amount), backToDollars)
			})
		}
	})

	t.Run("maximum safe conversion", func(t *testing.T) {
		// Test the maximum safe amount
		cents, err := DollarsToCents(MaxAmount)
		assert.NoError(t, err)
		assert.Equal(t, int64(9223372036854775807), cents)

		// One cent more should fail
		_, err = DollarsToCents(MaxAmount * 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum allowed")
	})
}
