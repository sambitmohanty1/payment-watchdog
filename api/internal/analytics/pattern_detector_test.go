package analytics

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sambitmohanty1/payment-watchdog/api/internal/architecture"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDetectPatterns_Empty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pd := NewDefaultPatternDetector(logger)

	patterns := pd.DetectPatterns([]*architecture.PaymentFailure{})
	assert.Empty(t, patterns)
}

func TestDetectPatterns_Recurring(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pd := NewDefaultPatternDetector(logger)

	customerID := "cust-1"
	now := time.Now()

	events := []*architecture.PaymentFailure{
		{ID: uuid.New(), CustomerID: customerID, OccurredAt: now.Add(-3 * time.Hour), AmountCents: 10000},
		{ID: uuid.New(), CustomerID: customerID, OccurredAt: now.Add(-2 * time.Hour), AmountCents: 10000},
		{ID: uuid.New(), CustomerID: customerID, OccurredAt: now.Add(-1 * time.Hour), AmountCents: 10000},
	}

	patterns := pd.DetectPatterns(events)

	var found bool
	for _, p := range patterns {
		if p.Type == PatternTypeRecurring {
			found = true
			assert.Contains(t, p.Description, "Recurring")
			assert.Equal(t, customerID, p.Metadata["customer_id"])
			assert.Equal(t, 3, p.Metadata["event_count"])
		}
	}
	assert.True(t, found, "Recurring pattern should be detected")
}

func TestDetectPatterns_Amount(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pd := NewDefaultPatternDetector(logger)

	now := time.Now()

	// 10 events, 90th percentile threshold will be the 9th event's amount
	// Index = int(10 * 0.9) = 9. amounts[9] = 1000.
	events := make([]*architecture.PaymentFailure, 10)
	for i := 0; i < 10; i++ {
		events[i] = &architecture.PaymentFailure{
			ID:          uuid.New(),
			AmountCents: int64(i+1) * 10000,
			OccurredAt:  now.Add(time.Duration(-i) * time.Hour),
		}
	}

	patterns := pd.DetectPatterns(events)

	var found bool
	for _, p := range patterns {
		if p.Type == PatternTypeAmount {
			found = true
			assert.Contains(t, p.Description, "High-value")
			assert.Equal(t, int64(100000), p.Metadata["threshold"])
		}
	}
	assert.True(t, found, "Amount pattern should be detected")
}

func TestDetectPatterns_Business(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pd := NewDefaultPatternDetector(logger)

	category := "retail"
	now := time.Now()

	events := []*architecture.PaymentFailure{
		{ID: uuid.New(), BusinessCategory: category, OccurredAt: now.Add(-2 * time.Hour), AmountCents: 10000},
		{ID: uuid.New(), BusinessCategory: category, OccurredAt: now.Add(-1 * time.Hour), AmountCents: 10000},
	}

	patterns := pd.DetectPatterns(events)

	var found bool
	for _, p := range patterns {
		if p.Type == PatternTypeBusiness {
			found = true
			assert.Contains(t, p.Description, "Business category")
			assert.Equal(t, category, p.Metadata["business_category"])
			assert.Equal(t, 2, p.Metadata["event_count"])
		}
	}
	assert.True(t, found, "Business pattern should be detected")
}

func TestDetectPatterns_DayOfWeek(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pd := NewDefaultPatternDetector(logger)

	// Threshold: len(events) >= 7 and maxCount/len(events) >= 0.3
	// Use 10 events, 3 on Monday (Weekday 1)
	monday := time.Date(2023, 10, 23, 10, 0, 0, 0, time.UTC) // 2023-10-23 is a Monday
	tue := monday.AddDate(0, 0, 1)
	wed := monday.AddDate(0, 0, 2)
	thu := monday.AddDate(0, 0, 3)
	fri := monday.AddDate(0, 0, 4)

	events := []*architecture.PaymentFailure{
		{ID: uuid.New(), OccurredAt: monday, AmountCents: 10000},
		{ID: uuid.New(), OccurredAt: monday, AmountCents: 10000},
		{ID: uuid.New(), OccurredAt: monday, AmountCents: 10000}, // 3/10 = 0.3
		{ID: uuid.New(), OccurredAt: tue, AmountCents: 10000},
		{ID: uuid.New(), OccurredAt: tue, AmountCents: 10000},
		{ID: uuid.New(), OccurredAt: wed, AmountCents: 10000},
		{ID: uuid.New(), OccurredAt: wed, AmountCents: 10000},
		{ID: uuid.New(), OccurredAt: thu, AmountCents: 10000},
		{ID: uuid.New(), OccurredAt: thu, AmountCents: 10000},
		{ID: uuid.New(), OccurredAt: fri, AmountCents: 10000},
	}

	patterns := pd.DetectPatterns(events)

	var found bool
	for _, p := range patterns {
		if p.Type == PatternTypeDayOfWeek {
			found = true
			assert.Equal(t, "Monday", p.Metadata["peak_day"])
			assert.Equal(t, 3, p.Metadata["peak_count"])
		}
	}
	assert.True(t, found, "Day of week pattern should be detected")
}

func TestDetectPatterns_TimeOfDay(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pd := NewDefaultPatternDetector(logger)

	// Threshold: len(events) >= 24 and maxCount/len(events) >= 0.15
	// 24 events, 0.15 * 24 = 3.6 -> need 4 events in one hour
	now := time.Date(2023, 10, 23, 10, 0, 0, 0, time.UTC)

	events := make([]*architecture.PaymentFailure, 24)
	for i := 0; i < 24; i++ {
		// Distribute so we avoid hour 10 for i >= 4
		hour := (i % 23)
		if hour >= 10 {
			hour++ // skip 10
		}

		if i < 4 {
			hour = 10 // first 4 events at 10:00
		}
		events[i] = &architecture.PaymentFailure{
			ID:          uuid.New(),
			OccurredAt:  now.Add(time.Duration(hour-10) * time.Hour),
			AmountCents: 10000,
		}
	}

	patterns := pd.DetectPatterns(events)

	var found bool
	for _, p := range patterns {
		if p.Type == PatternTypeTimeOfDay {
			found = true
			assert.Equal(t, 10, p.Metadata["peak_hour"])
			assert.Equal(t, 4, p.Metadata["peak_count"])
		}
	}
	assert.True(t, found, "Time of day pattern should be detected")
}

func TestDetectCustomerPatterns(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pd := NewDefaultPatternDetector(logger)

	customerID := "cust-1"
	now := time.Now()

	events := []*architecture.PaymentFailure{
		{ID: uuid.New(), CustomerID: customerID, OccurredAt: now.Add(-3 * time.Hour), AmountCents: 100000},
		{ID: uuid.New(), CustomerID: customerID, OccurredAt: now.Add(-2 * time.Hour), AmountCents: 200000},
		{ID: uuid.New(), CustomerID: customerID, OccurredAt: now.Add(-1 * time.Hour), AmountCents: 300000},
		{ID: uuid.New(), CustomerID: "other", OccurredAt: now, AmountCents: 50000},
	}

	patterns := pd.DetectCustomerPatterns(customerID, events)

	assert.Len(t, patterns, 1)
	assert.Equal(t, customerID, patterns[0].CustomerID)
	assert.Equal(t, PatternTypeRecurring, patterns[0].Pattern.Type)
	assert.Equal(t, float64(3), patterns[0].Frequency)
	assert.Equal(t, int64(600000), patterns[0].TotalAmountCents)
	assert.Equal(t, "medium", patterns[0].RiskLevel) // 3 events and 600000 cents -> medium
}

func TestDetectTemporalPatterns(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pd := NewDefaultPatternDetector(logger)

	// Need enough events for both day of week (7) and time of day (24)
	// We'll just test DayOfWeek to keep it simple, similar logic for TimeOfDay
	monday := time.Date(2023, 10, 23, 10, 0, 0, 0, time.UTC)
	events := make([]*architecture.PaymentFailure, 10)
	for i := 0; i < 10; i++ {
		day := i % 5
		if i < 4 {
			day = 0 // Monday
		}
		events[i] = &architecture.PaymentFailure{
			ID:          uuid.New(),
			OccurredAt:  monday.AddDate(0, 0, day),
			AmountCents: 10000,
		}
	}

	patterns := pd.DetectTemporalPatterns(events, 7*24*time.Hour)

	assert.NotEmpty(t, patterns)
	var foundWeekly bool
	for _, p := range patterns {
		if p.Seasonality == "weekly" {
			foundWeekly = true
			assert.Equal(t, PatternTypeDayOfWeek, p.Pattern.Type)
			assert.NotEmpty(t, p.PeakTimes)
		}
	}
	assert.True(t, foundWeekly)
}

func TestDetectBusinessPatterns(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pd := NewDefaultPatternDetector(logger)

	category := "retail"
	events := []*architecture.PaymentFailure{
		{ID: uuid.New(), BusinessCategory: category, OccurredAt: time.Now(), AmountCents: 10000},
		{ID: uuid.New(), BusinessCategory: category, OccurredAt: time.Now(), AmountCents: 10000},
	}

	patterns := pd.DetectBusinessPatterns(events)

	assert.Len(t, patterns, 1)
	assert.Equal(t, PatternTypeBusiness, patterns[0].Type)
	assert.Equal(t, category, patterns[0].Metadata["business_category"])
}
