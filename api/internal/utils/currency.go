package utils

import (
	"errors"
	"fmt"
	"math"
)

const (
	// CentsPerDollar represents the number of cents in one dollar
	CentsPerDollar = 100
	// MaxAmount represents the maximum allowed amount in dollars to prevent overflow
	MaxAmount = 92233720368547760.00 // math.MaxInt64 / 100
)

// CurrencyConversionError represents errors during currency conversion
type CurrencyConversionError struct {
	Amount    float64
	Operation string
	Err       error
}

func (e *CurrencyConversionError) Error() string {
	return fmt.Sprintf("currency conversion error during %s with amount %.2f: %v", e.Operation, e.Amount, e.Err)
}

// DollarsToCents converts a dollar amount to cents safely
func DollarsToCents(dollars float64) (int64, error) {
	// Check for NaN or infinity first
	if math.IsNaN(dollars) || math.IsInf(dollars, 1) || math.IsInf(dollars, -1) {
		return 0, &CurrencyConversionError{
			Amount:    dollars,
			Operation: "dollars_to_cents",
			Err:       errors.New("amount must be a finite number"),
		}
	}

	if dollars < 0 {
		return 0, &CurrencyConversionError{
			Amount:    dollars,
			Operation: "dollars_to_cents",
			Err:       errors.New("amount cannot be negative"),
		}
	}

	if dollars > MaxAmount {
		return 0, &CurrencyConversionError{
			Amount:    dollars,
			Operation: "dollars_to_cents",
			Err:       fmt.Errorf("amount %.2f exceeds maximum allowed %.2f", dollars, MaxAmount),
		}
	}

	// Convert to cents with proper rounding
	cents := int64(math.Round(dollars * CentsPerDollar))

	// Verify the conversion is accurate
	if cents < 0 {
		return 0, &CurrencyConversionError{
			Amount:    dollars,
			Operation: "dollars_to_cents",
			Err:       errors.New("conversion resulted in negative cents"),
		}
	}

	return cents, nil
}

// CentsToDollars converts cents to dollars safely
func CentsToDollars(cents int64) (float64, error) {
	if cents < 0 {
		return 0, &CurrencyConversionError{
			Amount:    float64(cents),
			Operation: "cents_to_dollars",
			Err:       errors.New("cents cannot be negative"),
		}
	}

	dollars := float64(cents) / CentsPerDollar
	return dollars, nil
}

// ValidateAmount validates a dollar amount for common constraints
func ValidateAmount(amount *float64) error {
	if amount == nil {
		return nil // nil amount is valid (optional field)
	}

	// Check for NaN or infinity first
	if math.IsNaN(*amount) || math.IsInf(*amount, 1) || math.IsInf(*amount, -1) {
		return fmt.Errorf("amount must be a finite number, got: %v", *amount)
	}

	if *amount < 0 {
		return fmt.Errorf("amount cannot be negative: %.2f", *amount)
	}

	if *amount > MaxAmount {
		return fmt.Errorf("amount %.2f exceeds maximum allowed %.2f", *amount, MaxAmount)
	}

	return nil
}

// ValidateCents validates a cents amount for common constraints
func ValidateCents(cents int64) error {
	if cents < 0 {
		return fmt.Errorf("cents cannot be negative: %d", cents)
	}

	return nil
}

// SafeDollarsToCents converts dollars to cents with validation
func SafeDollarsToCents(dollars *float64) (int64, error) {
	if err := ValidateAmount(dollars); err != nil {
		return 0, err
	}

	if dollars == nil {
		return 0, nil // nil amount results in 0 cents
	}

	return DollarsToCents(*dollars)
}

// FormatCents formats cents as a currency string
func FormatCents(cents int64) string {
	dollars := float64(cents) / CentsPerDollar
	return fmt.Sprintf("%.2f", dollars)
}

// RoundToNearestCent rounds a dollar amount to the nearest cent
func RoundToNearestCent(dollars float64) float64 {
	return math.Round(dollars*CentsPerDollar) / CentsPerDollar
}
