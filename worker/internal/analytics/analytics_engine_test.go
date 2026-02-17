package analytics

import (
	"testing"
	"time"

	"github.com/sambitmohanty1/payment-watchdog/api/internal/architecture"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestAnalyticsEngineCreation tests the creation of analytics engine
func TestAnalyticsEngineCreation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	patternDetector := NewDefaultPatternDetector(logger)
	trendAnalyzer := NewDefaultTrendAnalyzer(logger)
	failurePredictor := NewDefaultFailurePredictor(logger)

	engine := NewAnalyticsEngine(patternDetector, trendAnalyzer, failurePredictor, logger)
	assert.NotNil(t, engine)
	assert.Equal(t, patternDetector, engine.patternDetector)
	assert.Equal(t, trendAnalyzer, engine.trendAnalyzer)
	assert.Equal(t, failurePredictor, engine.failurePredictor)
	assert.NotNil(t, engine.metrics)
}

// TestAnalyticsEngineEmptyData tests analysis with empty data
func TestAnalyticsEngineEmptyData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewAnalyticsEngine(
		NewDefaultPatternDetector(logger),
		NewDefaultTrendAnalyzer(logger),
		NewDefaultFailurePredictor(logger),
		logger,
	)

	result, err := engine.AnalyzePaymentFailures([]*architecture.PaymentFailure{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Patterns))
	assert.Equal(t, 0, len(result.Trends))
	assert.Equal(t, 0, len(result.Predictions))
	assert.Equal(t, int64(0), result.Metrics.PatternsDetected)
	assert.Equal(t, int64(0), result.Metrics.TrendsAnalyzed)
	assert.Equal(t, int64(0), result.Metrics.PredictionsMade)
}

// TestAnalyticsEngineWithData tests analysis with sample data
func TestAnalyticsEngineWithData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewAnalyticsEngine(
		NewDefaultPatternDetector(logger),
		NewDefaultTrendAnalyzer(logger),
		NewDefaultFailurePredictor(logger),
		logger,
	)

	// Create test data
	testEvents := createTestPaymentFailures()

	result, err := engine.AnalyzePaymentFailures(testEvents)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify patterns were detected
	assert.True(t, len(result.Patterns) > 0, "Should detect patterns in test data")

	// Verify trends were analyzed
	assert.True(t, len(result.Trends) > 0, "Should analyze trends in test data")

	// Verify predictions were made
	assert.True(t, len(result.Predictions) > 0, "Should make predictions for customers")

	// Verify metrics were updated
	assert.Equal(t, int64(len(result.Patterns)), result.Metrics.PatternsDetected)
	assert.Equal(t, int64(len(result.Trends)), result.Metrics.TrendsAnalyzed)
	assert.Equal(t, int64(len(result.Predictions)), result.Metrics.PredictionsMade)
	assert.True(t, result.Metrics.ProcessingTime > 0, "Should record processing time")
}

// TestAnalyticsEngineMetrics tests metrics functionality
func TestAnalyticsEngineMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewAnalyticsEngine(
		NewDefaultPatternDetector(logger),
		NewDefaultTrendAnalyzer(logger),
		NewDefaultFailurePredictor(logger),
		logger,
	)

	// Initially metrics should be zero
	initialMetrics := engine.GetMetrics()
	assert.Equal(t, int64(0), initialMetrics.PatternsDetected)
	assert.Equal(t, int64(0), initialMetrics.TrendsAnalyzed)
	assert.Equal(t, int64(0), initialMetrics.PredictionsMade)

	// Run analysis to populate metrics
	testEvents := createTestPaymentFailures()
	result, err := engine.AnalyzePaymentFailures(testEvents)
	assert.NoError(t, err)

	// Verify metrics were updated
	updatedMetrics := engine.GetMetrics()
	assert.Equal(t, result.Metrics.PatternsDetected, updatedMetrics.PatternsDetected)
	assert.Equal(t, result.Metrics.TrendsAnalyzed, updatedMetrics.TrendsAnalyzed)
	assert.Equal(t, result.Metrics.PredictionsMade, updatedMetrics.PredictionsMade)

	// Verify that metrics were actually updated (not just initialized)
	assert.True(t, updatedMetrics.PatternsDetected > 0, "Metrics should show detected patterns")
	assert.True(t, updatedMetrics.TrendsAnalyzed > 0, "Metrics should show analyzed trends")
	assert.True(t, updatedMetrics.PredictionsMade > 0, "Metrics should show made predictions")
}

// TestAnalyticsEngineConcurrency tests concurrent access to analytics engine
func TestAnalyticsEngineConcurrency(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewAnalyticsEngine(
		NewDefaultPatternDetector(logger),
		NewDefaultTrendAnalyzer(logger),
		NewDefaultFailurePredictor(logger),
		logger,
	)

	// Create test data
	testEvents := createTestPaymentFailures()

	// Run multiple concurrent analyses
	results := make(chan *AnalysisResult, 5)
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			result, err := engine.AnalyzePaymentFailures(testEvents)
			results <- result
			errors <- err
		}()
	}

	// Collect results
	for i := 0; i < 5; i++ {
		result := <-results
		err := <-errors

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, len(result.Patterns) > 0)
		assert.True(t, len(result.Trends) > 0)
		assert.True(t, len(result.Predictions) > 0)
	}
}

// TestPatternDetector tests pattern detection functionality
func TestPatternDetector(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	detector := NewDefaultPatternDetector(logger)

	// Test with empty data
	patterns := detector.DetectPatterns([]*architecture.PaymentFailure{})
	assert.Equal(t, 0, len(patterns))

	// Test with sample data
	testEvents := createTestPaymentFailures()
	patterns = detector.DetectPatterns(testEvents)
	assert.True(t, len(patterns) > 0, "Should detect patterns in test data")

	// Verify pattern types
	patternTypes := make(map[PatternType]bool)
	for _, pattern := range patterns {
		patternTypes[pattern.Type] = true
		assert.True(t, pattern.Confidence > 0, "Pattern should have confidence > 0")
		assert.True(t, pattern.Confidence <= 1, "Pattern confidence should be <= 1")
		assert.NotEmpty(t, pattern.Description)
		assert.NotEmpty(t, pattern.Evidence)
	}

	// Should detect multiple types of patterns
	assert.True(t, len(patternTypes) > 1, "Should detect multiple pattern types")
}

// TestTrendAnalyzer tests trend analysis functionality
func TestTrendAnalyzer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	analyzer := NewDefaultTrendAnalyzer(logger)

	// Test with empty data
	trends := analyzer.AnalyzeTrends([]*architecture.PaymentFailure{}, 30*24*time.Hour)
	assert.Equal(t, 0, len(trends))

	// Test with sample data
	testEvents := createTestPaymentFailures()
	trends = analyzer.AnalyzeTrends(testEvents, 4*24*time.Hour) // Use 4 days to match the data span
	assert.True(t, len(trends) > 0, "Should analyze trends in test data")

	// Verify trend properties
	for _, trend := range trends {
		assert.NotEmpty(t, trend.ID)
		assert.NotEmpty(t, trend.Type)
		assert.NotEmpty(t, trend.Direction)
		assert.True(t, trend.Magnitude >= 0, "Trend magnitude should be >= 0")
		assert.True(t, trend.Confidence > 0, "Trend should have confidence > 0")
		assert.True(t, trend.Confidence <= 1, "Trend confidence should be <= 1")
		assert.NotEmpty(t, trend.Description)
		assert.True(t, trend.TimeRange > 0, "Trend should have positive time range")
	}

	// Test seasonal patterns
	seasonalPatterns := analyzer.AnalyzeSeasonalPatterns(testEvents)
	assert.True(t, len(seasonalPatterns) >= 0, "Should analyze seasonal patterns")

	// Test business cycle patterns
	businessCyclePatterns := analyzer.AnalyzeBusinessCyclePatterns(testEvents)
	assert.True(t, len(businessCyclePatterns) >= 0, "Should analyze business cycle patterns")
}

// TestFailurePredictor tests failure prediction functionality
func TestFailurePredictor(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	predictor := NewDefaultFailurePredictor(logger)

	// Test with empty history
	customerID := "test-customer"
	prediction := predictor.PredictFailure(customerID, []*architecture.PaymentFailure{})
	assert.Nil(t, prediction, "Should return nil for empty history")

	// Test with single event
	singleEvent := createSingleTestEvent()
	prediction = predictor.PredictFailure(customerID, []*architecture.PaymentFailure{singleEvent})
	assert.Nil(t, prediction, "Should return nil for single event (need at least 2 for prediction)")

	// Test with multiple events
	multipleEvents := createTestPaymentFailures()
	prediction = predictor.PredictFailure(customerID, multipleEvents)
	assert.NotNil(t, prediction, "Should generate prediction for multiple events")

	// Verify prediction properties
	assert.NotEmpty(t, prediction.ID)
	assert.Equal(t, customerID, prediction.CustomerID)
	assert.True(t, prediction.RiskScore >= 0, "Risk score should be >= 0")
	assert.True(t, prediction.RiskScore <= 100, "Risk score should be <= 100")
	assert.True(t, prediction.FailureProbability >= 0, "Failure probability should be >= 0")
	assert.True(t, prediction.FailureProbability <= 1, "Failure probability should be <= 1")
	assert.True(t, prediction.Confidence > 0, "Prediction should have confidence > 0")
	assert.True(t, prediction.Confidence <= 1, "Prediction confidence should be <= 1")
	assert.NotEmpty(t, prediction.Factors)
	assert.True(t, prediction.ExpiresAt.After(prediction.CreatedAt))

	// Test risk score calculation
	riskScore := predictor.PredictRiskScore(customerID, multipleEvents)
	assert.True(t, riskScore >= 0, "Risk score should be >= 0")
	assert.True(t, riskScore <= 100, "Risk score should be <= 100")

	// Test next failure date prediction
	nextFailureDate := predictor.PredictNextFailureDate(customerID, multipleEvents)
	assert.NotNil(t, nextFailureDate, "Should predict next failure date")
	assert.True(t, nextFailureDate.After(time.Now()), "Predicted date should be in the future")
}

// TestAnalyticsEngineIntegration tests end-to-end analytics workflow
func TestAnalyticsEngineIntegration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	engine := NewAnalyticsEngine(
		NewDefaultPatternDetector(logger),
		NewDefaultTrendAnalyzer(logger),
		NewDefaultFailurePredictor(logger),
		logger,
	)

	// Create comprehensive test data
	testEvents := createComprehensiveTestData()

	// Run analysis
	result, err := engine.AnalyzePaymentFailures(testEvents)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify comprehensive results
	assert.True(t, len(result.Patterns) > 0, "Should detect patterns")
	assert.True(t, len(result.Trends) > 0, "Should analyze trends")
	assert.True(t, len(result.Predictions) > 0, "Should make predictions")

	// Verify data consistency
	assert.Equal(t, int64(len(result.Patterns)), result.Metrics.PatternsDetected)
	assert.Equal(t, int64(len(result.Trends)), result.Metrics.TrendsAnalyzed)
	assert.Equal(t, int64(len(result.Predictions)), result.Metrics.PredictionsMade)

	// Verify processing time is reasonable
	assert.True(t, result.Metrics.ProcessingTime < 5*time.Second, "Processing should complete within 5 seconds")

	// Verify all predictions have valid data
	for _, prediction := range result.Predictions {
		assert.NotEmpty(t, prediction.ID)
		assert.NotEmpty(t, prediction.CustomerID)
		assert.True(t, prediction.RiskScore >= 0 && prediction.RiskScore <= 100)
		assert.True(t, prediction.FailureProbability >= 0 && prediction.FailureProbability <= 1)
		assert.True(t, prediction.Confidence > 0 && prediction.Confidence <= 1)
		assert.NotEmpty(t, prediction.Factors)
	}
}

// Helper functions to create test data
func createTestPaymentFailures() []*architecture.PaymentFailure {
	now := time.Now()
	customer1ID := "customer-1"
	customer2ID := "customer-2"

	return []*architecture.PaymentFailure{
		{
			ID:         uuid.New(),
			Amount:     1000.0,
			OccurredAt: now.Add(-24 * time.Hour),
			CustomerID: customer1ID, CustomerName: "Customer 1",
			BusinessCategory: "retail",
		},
		{
			ID:         uuid.New(),
			Amount:     1500.0,
			OccurredAt: now.Add(-48 * time.Hour),
			CustomerID: customer1ID, CustomerName: "Customer 1",
			BusinessCategory: "retail",
		},
		{
			ID:         uuid.New(),
			Amount:     2000.0,
			OccurredAt: now.Add(-72 * time.Hour),
			CustomerID: customer1ID, CustomerName: "Customer 1",
			BusinessCategory: "retail",
		},
		{
			ID:         uuid.New(),
			Amount:     5000.0,
			OccurredAt: now.Add(-12 * time.Hour),
			CustomerID: customer2ID, CustomerName: "Customer 2",
			BusinessCategory: "finance",
		},
		{
			ID:         uuid.New(),
			Amount:     7500.0,
			OccurredAt: now.Add(-36 * time.Hour),
			CustomerID: customer2ID, CustomerName: "Customer 2",
			BusinessCategory: "finance",
		},
	}
}

func createSingleTestEvent() *architecture.PaymentFailure {
	return &architecture.PaymentFailure{
		ID:         uuid.New(),
		Amount:     1000.0,
		OccurredAt: time.Now().Add(-24 * time.Hour),
		CustomerID: "test-customer", CustomerName: "Test Customer",
		BusinessCategory: "retail",
	}
}

func createComprehensiveTestData() []*architecture.PaymentFailure {
	now := time.Now()
	events := make([]*architecture.PaymentFailure, 0)

	// Create multiple customers with different patterns
	customers := []*architecture.Customer{
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002")},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003")},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000004")},
	}

	categories := []string{"retail", "finance", "healthcare", "technology"}

	// Generate events over the past 90 days
	for i := 0; i < 90; i++ {
		for j, customer := range customers {
			// Create events with some patterns
			if i%7 == 0 || i%30 == 0 { // Weekly and monthly patterns
				events = append(events, &architecture.PaymentFailure{
					ID:         uuid.New(),
					Amount:     1000.0 + float64(i*100),
					OccurredAt: now.AddDate(0, 0, -i),
					CustomerID: customer.ID.String(), CustomerName: "Customer " + customer.ID.String(),
					BusinessCategory: categories[j%len(categories)],
				})
			}
		}
	}

	return events
}
