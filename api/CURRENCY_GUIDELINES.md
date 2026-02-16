# Currency Field Usage Guidelines

## Overview

This document provides guidelines for consistent usage of currency fields throughout the payment-watchdog codebase to avoid precision errors and confusion.

## Field Types

### AmountCents (Preferred)
- **Type**: `int64`
- **Usage**: **PRIMARY** field for storing monetary values
- **Format**: Amount in cents (e.g., $10.50 = 1050 cents)
- **Benefits**: 
  - No floating-point precision issues
  - Exact arithmetic operations
  - Database-friendly
  - Industry standard for financial applications

### Amount (Legacy)
- **Type**: `float64`
- **Usage**: **LEGACY** field, being phased out
- **Format**: Amount in dollars (e.g., $10.50)
- **Issues**:
  - Floating-point precision errors
  - Inexact arithmetic
  - Not suitable for financial calculations

## Conversion Rules

### When to Use AmountCents
- **Database storage** (all new models)
- **Financial calculations**
- **API responses** (convert to dollars for display)
- **Payment processing**

### When to Use Amount
- **Display purposes** (converted from AmountCents)
- **External APIs that require dollars**
- **Legacy system compatibility**

## Conversion Functions

Use the utility functions in `internal/utils/currency.go`:

```go
// Convert dollars to cents safely
cents, err := utils.SafeDollarsToCents(dollarAmount)

// Convert cents to dollars
dollars, err := utils.CentsToDollars(cents)

// Validate amounts
err := utils.ValidateAmount(dollarAmount)
err := utils.ValidateCents(cents)

// Format for display
formatted := utils.FormatCents(cents) // "10.50"
```

## Database Schema Guidelines

### New Tables
```sql
-- CORRECT
CREATE TABLE payment_failures (
    id UUID PRIMARY KEY,
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    -- other fields
);

-- AVOID
CREATE TABLE payment_failures (
    id UUID PRIMARY KEY,
    amount DECIMAL(10,2) NOT NULL, -- Can have precision issues
    -- other fields
);
```

### Existing Tables
- Migrate `amount` fields to `amount_cents`
- Add migration scripts to convert existing data
- Update application code to use AmountCents

## API Response Format

### JSON Responses
Always use dollars for API responses (converted from AmountCents):

```json
{
  "amount": 10.50,
  "currency": "USD"
}
```

### JSON Requests
Accept dollars but validate and convert to cents:

```go
// Input validation
amountCents, err := utils.SafeDollarsToCents(request.Amount)
if err != nil {
    return ValidationError("Invalid amount")
}
```

## Code Examples

### Correct Usage
```go
// Model
type PaymentFailure struct {
    ID          uuid.UUID `json:"id"`
    AmountCents int64     `json:"-"` // Hidden in JSON
    Currency    string    `json:"currency"`
}

// API handler
func (h *Handler) GetPayment(c *gin.Context) {
    payment := getPayment()
    
    response := gin.H{
        "amount":    float64(payment.AmountCents) / 100,
        "currency":  payment.Currency,
    }
    c.JSON(http.StatusOK, response)
}

// Calculation
total := payment1.AmountCents + payment2.AmountCents // Safe
```

### Incorrect Usage
```go
// AVOID: Direct float64 arithmetic
total := payment1.Amount + payment2.Amount // Precision issues

// AVOID: Storing dollars in database
amount := 10.50 // float64
```

## Migration Strategy

### Phase 1: Add AmountCents Fields
1. Add `amount_cents` columns to existing tables
2. Keep existing `amount` columns for compatibility
3. Add triggers to keep both fields in sync

### Phase 2: Update Application Code
1. Use utility functions for all conversions
2. Update business logic to use AmountCents
3. Update API handlers to convert for display

### Phase 3: Remove Legacy Fields
1. Remove `amount` columns from database
2. Update models to remove Amount fields
3. Update documentation

## Testing

### Unit Tests
- Test all currency conversions
- Test edge cases (zero, negative, large amounts)
- Test validation functions

### Integration Tests
- Test API request/response conversion
- Test database storage and retrieval
- Test calculation accuracy

## Common Pitfalls

### 1. Floating-Point Precision
```go
// WRONG
amount := 10.10 + 10.20 // Result: 20.300000000000004

// RIGHT
cents1 := 1010
cents2 := 1020
total := cents1 + cents2 // Result: 2030 cents ($20.30)
```

### 2. Mixing Field Types
```go
// WRONG
if payment.AmountCents > request.Amount { // Comparing cents to dollars

// RIGHT
requestCents, _ := utils.SafeDollarsToCents(request.Amount)
if payment.AmountCents > requestCents { // Comparing cents to cents
```

### 3. Missing Validation
```go
// WRONG
cents := int64(dollars * 100) // Can overflow

// RIGHT
cents, err := utils.SafeDollarsToCents(dollars)
if err != nil {
    return err
}
```

## References

- [IEEE 754 Floating Point Issues](https://en.wikipedia.org/wiki/IEEE_754)
- [Financial Calculations Best Practices](https://martin.kleppmann.com/2012/12/05/youre-certain-about-money-youre-wrong.html)
- [Go Currency Handling](https://blog.golang.org/strings)
