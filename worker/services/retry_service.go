package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RetryService handles retry logic with exponential backoff and dead letter queue
type RetryService struct {
	db              *gorm.DB
	maxRetries      int
	baseDelay       time.Duration
	maxDelay        time.Duration
	deadLetterQueue chan RetryJob
	workers         int
	mu              sync.RWMutex
	activeJobs      map[string]*RetryJob
	logger          *zap.Logger
}

// RetryJob represents a job that needs to be retried
type RetryJob struct {
	ID          string                 `json:"id" gorm:"column:id;primaryKey"`
	JobType     string                 `json:"job_type" gorm:"column:job_type;not null"`
	CompanyID   string                 `json:"company_id" gorm:"column:company_id;not null"`
	Data        map[string]interface{} `json:"data" gorm:"column:data;type:jsonb"`
	RetryCount  int                    `json:"retry_count" gorm:"column:retry_count;default:0"`
	MaxRetries  int                    `json:"max_retries" gorm:"column:max_retries;default:3"`
	Status      string                 `json:"status" gorm:"column:status;default:'pending'"`
	Error       string                 `json:"error,omitempty" gorm:"column:error"`
	NextRetryAt *time.Time             `json:"next_retry_at,omitempty" gorm:"column:next_retry_at"`
	CreatedAt   time.Time              `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" gorm:"column:completed_at"`
}

// TableName specifies the table name for GORM
func (RetryJob) TableName() string {
	return "retry_jobs"
}

// RetryJobResult represents the result of a retry job
type RetryJobResult struct {
	Success bool
	Error   error
	Data    interface{}
}

// NewRetryService creates a new retry service
func NewRetryService(db *gorm.DB, maxRetries int, baseDelay, maxDelay time.Duration, workers int, logger *zap.Logger) *RetryService {
	// Safety check: ensure we have a valid database connection
	if db == nil {
		panic("RetryService requires a valid database connection")
	}

	// Add service context to logger
	serviceLogger := logger.With(zap.String("service", "retry-service"), zap.String("version", "1.0.0"))

	service := &RetryService{
		db:              db,
		maxRetries:      maxRetries,
		baseDelay:       baseDelay,
		maxDelay:        maxDelay,
		deadLetterQueue: make(chan RetryJob, 1000),
		workers:         workers,
		activeJobs:      make(map[string]*RetryJob),
		logger:          serviceLogger,
	}

	serviceLogger.Info("Retry service initialized",
		zap.Int("max_retries", maxRetries),
		zap.Duration("base_delay", baseDelay),
		zap.Duration("max_delay", maxDelay),
		zap.Int("workers", workers),
		zap.Int("dead_letter_queue_size", 1000))

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		go service.worker()
	}

	// Start dead letter queue processor
	go service.processDeadLetterQueue()

	return service
}

// SubmitJob submits a new job for retry processing
func (r *RetryService) SubmitJob(ctx context.Context, jobType, companyID string, data map[string]interface{}) (*RetryJob, error) {
	// Safety check: ensure we have a valid database connection
	if r.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Check if the retry_jobs table exists first
	var tableExists bool
	if err := r.db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'retry_jobs')").Scan(&tableExists).Error; err != nil {
		return nil, fmt.Errorf("failed to check if retry_jobs table exists: %w", err)
	}

	if !tableExists {
		// Table doesn't exist yet, return a mock job for now
		job := &RetryJob{
			ID:         uuid.New().String(),
			JobType:    jobType,
			CompanyID:  companyID,
			Data:       data,
			RetryCount: 0,
			MaxRetries: r.maxRetries,
			Status:     "pending",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// Add to active jobs without database persistence
		r.mu.Lock()
		r.activeJobs[job.ID] = job
		r.mu.Unlock()

		// Process immediately
		go r.processJob(job)

		return job, nil
	}

	job := &RetryJob{
		ID:         uuid.New().String(),
		JobType:    jobType,
		CompanyID:  companyID,
		Data:       data,
		RetryCount: 0,
		MaxRetries: r.maxRetries,
		Status:     "pending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save to database
	if err := r.db.Create(job).Error; err != nil {
		return nil, fmt.Errorf("failed to save retry job: %w", err)
	}

	// Add to active jobs
	r.mu.Lock()
	r.activeJobs[job.ID] = job
	r.mu.Unlock()

	// Process immediately
	go r.processJob(job)

	return job, nil
}

// processJob processes a single retry job
func (r *RetryService) processJob(job *RetryJob) {
	ctx := context.Background()

	// Execute the job
	result := r.executeJob(ctx, job)

	if result.Success {
		// Job completed successfully
		r.completeJob(job, nil)
		return
	}

	// Job failed, check if we should retry
	if job.RetryCount < job.MaxRetries {
		r.scheduleRetry(job, result.Error)
	} else {
		// Max retries exceeded, send to dead letter queue
		r.sendToDeadLetterQueue(job, result.Error)
	}
}

// executeJob executes the actual job logic
func (r *RetryService) executeJob(ctx context.Context, job *RetryJob) *RetryJobResult {
	// This is a placeholder - in real implementation, this would call the actual job handler
	switch job.JobType {
	case "webhook_processing":
		return r.executeWebhookJob(ctx, job)
	case "payment_retry":
		return r.executePaymentRetryJob(ctx, job)
	case "alert_notification":
		return r.executeAlertNotificationJob(ctx, job)
	default:
		return &RetryJobResult{
			Success: false,
			Error:   fmt.Errorf("unknown job type: %s", job.JobType),
		}
	}
}

// executeWebhookJob executes a webhook processing job
func (r *RetryService) executeWebhookJob(ctx context.Context, job *RetryJob) *RetryJobResult {
	// Placeholder implementation
	// In real implementation, this would call the webhook processing logic
	return &RetryJobResult{
		Success: true,
		Data:    "webhook processed successfully",
	}
}

// executePaymentRetryJob executes a payment retry job
func (r *RetryService) executePaymentRetryJob(ctx context.Context, job *RetryJob) *RetryJobResult {
	// Placeholder implementation
	// In real implementation, this would call the payment retry logic
	return &RetryJobResult{
		Success: true,
		Data:    "payment retry completed",
	}
}

// executeAlertNotificationJob executes an alert notification job
func (r *RetryService) executeAlertNotificationJob(ctx context.Context, job *RetryJob) *RetryJobResult {
	// Placeholder implementation
	// In real implementation, this would call the alert notification logic
	return &RetryJobResult{
		Success: true,
		Data:    "alert notification sent",
	}
}

// scheduleRetry schedules a job for retry with exponential backoff
func (r *RetryService) scheduleRetry(job *RetryJob, err error) {
	job.RetryCount++
	job.Status = "scheduled"
	job.Error = err.Error()

	// Calculate delay with exponential backoff
	delay := r.calculateDelay(job.RetryCount)
	nextRetry := time.Now().Add(delay)
	job.NextRetryAt = &nextRetry

	// Update database
	if updateErr := r.db.Save(job).Error; updateErr != nil {
		r.logger.Error("Failed to update retry job",
			zap.String("job_id", job.ID),
			zap.String("job_type", job.JobType),
			zap.String("company_id", job.CompanyID),
			zap.Error(updateErr))
		return
	}

	// Schedule retry
	time.AfterFunc(delay, func() {
		r.processJob(job)
	})
}

// calculateDelay calculates the delay for the next retry using exponential backoff
func (r *RetryService) calculateDelay(retryCount int) time.Duration {
	delay := r.baseDelay * time.Duration(1<<retryCount)
	if delay > r.maxDelay {
		delay = r.maxDelay
	}
	return delay
}

// completeJob marks a job as completed
func (r *RetryService) completeJob(job *RetryJob, result interface{}) {
	job.Status = "completed"
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	// Update database
	if err := r.db.Save(job).Error; err != nil {
		r.logger.Error("Failed to update completed job",
			zap.String("job_id", job.ID),
			zap.String("job_type", job.JobType),
			zap.String("company_id", job.CompanyID),
			zap.Error(err))
	}

	// Remove from active jobs
	r.mu.Lock()
	delete(r.activeJobs, job.ID)
	r.mu.Unlock()
}

// sendToDeadLetterQueue sends a job to the dead letter queue
func (r *RetryService) sendToDeadLetterQueue(job *RetryJob, err error) {
	job.Status = "dead_letter"
	job.Error = err.Error()

	// Update database
	if updateErr := r.db.Save(job).Error; updateErr != nil {
		r.logger.Error("Failed to update dead letter job",
			zap.String("job_id", job.ID),
			zap.String("job_type", job.JobType),
			zap.String("company_id", job.CompanyID),
			zap.Error(updateErr))
	}

	// Send to dead letter queue
	select {
	case r.deadLetterQueue <- *job:
		// Successfully queued
	default:
		// Queue is full, log error
		r.logger.Error("Dead letter queue full, dropping job",
			zap.String("job_id", job.ID),
			zap.String("job_type", job.JobType),
			zap.String("company_id", job.CompanyID),
			zap.Int("queue_size", len(r.deadLetterQueue)))
	}

	// Remove from active jobs
	r.mu.Lock()
	delete(r.activeJobs, job.ID)
	r.mu.Unlock()
}

// worker processes jobs from the active jobs map
func (r *RetryService) worker() {
	// This worker processes jobs that are scheduled for retry
	// In a real implementation, this would poll the database for scheduled jobs
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		r.processScheduledJobs()
	}
}

// processScheduledJobs processes jobs that are scheduled for retry
func (r *RetryService) processScheduledJobs() {
	// Safety check: ensure we have a valid database connection
	if r.db == nil {
		return
	}

	// Check if the retry_jobs table exists first
	var tableExists bool
	if err := r.db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'retry_jobs')").Scan(&tableExists).Error; err != nil {
		r.logger.Error("Failed to check if retry_jobs table exists",
			zap.Error(err))
		return
	}

	if !tableExists {
		// Table doesn't exist yet, skip processing
		return
	}

	var scheduledJobs []RetryJob
	if err := r.db.Where("status = ? AND next_retry_at <= ?", "scheduled", time.Now()).Find(&scheduledJobs).Error; err != nil {
		r.logger.Error("Failed to get scheduled jobs",
			zap.Error(err))
		return
	}

	for _, job := range scheduledJobs {
		// Check if job is still active
		r.mu.RLock()
		_, exists := r.activeJobs[job.ID]
		r.mu.RUnlock()

		if !exists {
			// Job was removed, skip
			continue
		}

		// Process the job
		go r.processJob(&job)
	}
}

// processDeadLetterQueue processes jobs in the dead letter queue
func (r *RetryService) processDeadLetterQueue() {
	for job := range r.deadLetterQueue {
		// Log dead letter job
		r.logger.Error("Dead letter job detected",
			zap.String("job_id", job.ID),
			zap.String("job_type", job.JobType),
			zap.String("company_id", job.CompanyID),
			zap.String("error", job.Error),
			zap.String("retry_count", fmt.Sprintf("%d", job.RetryCount)))

		// In a real implementation, this would:
		// 1. Send alerts to administrators
		// 2. Store in persistent dead letter storage
		// 3. Provide manual intervention capabilities
	}
}

// GetJobStatus gets the status of a specific job
func (r *RetryService) GetJobStatus(jobID string) (*RetryJob, error) {
	// Check if the retry_jobs table exists first
	var tableExists bool
	if err := r.db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'retry_jobs')").Scan(&tableExists).Error; err != nil {
		return nil, fmt.Errorf("failed to check if retry_jobs table exists: %w", err)
	}

	if !tableExists {
		return nil, fmt.Errorf("retry_jobs table does not exist yet")
	}

	var job RetryJob
	if err := r.db.Where("id = ?", jobID).First(&job).Error; err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	return &job, nil
}

// GetCompanyJobs gets all jobs for a specific company
func (r *RetryService) GetCompanyJobs(companyID string, limit int) ([]RetryJob, error) {
	// Check if the retry_jobs table exists first
	var tableExists bool
	if err := r.db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'retry_jobs')").Scan(&tableExists).Error; err != nil {
		return nil, fmt.Errorf("failed to check if retry_jobs table exists: %w", err)
	}

	if !tableExists {
		return []RetryJob{}, nil // Return empty list if table doesn't exist
	}

	var jobs []RetryJob
	query := r.db.Where("company_id = ?", companyID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to get company jobs: %w", err)
	}

	return jobs, nil
}

// RetryFailedJob manually retries a failed job
func (r *RetryService) RetryFailedJob(jobID string) error {
	var job RetryJob
	if err := r.db.Where("id = ?", jobID).First(&job).Error; err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	if job.Status != "failed" && job.Status != "dead_letter" {
		return fmt.Errorf("job is not in a retryable state: %s", job.Status)
	}

	// Reset job for retry
	job.Status = "pending"
	job.RetryCount = 0
	job.Error = ""
	job.NextRetryAt = nil

	// Update database
	if err := r.db.Save(&job).Error; err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Add back to active jobs
	r.mu.Lock()
	r.activeJobs[job.ID] = &job
	r.mu.Unlock()

	// Process immediately
	go r.processJob(&job)

	return nil
}

// GetStats returns retry service statistics
func (r *RetryService) GetStats() map[string]interface{} {
	// Check if the retry_jobs table exists first
	var tableExists bool
	if err := r.db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'retry_jobs')").Scan(&tableExists).Error; err != nil {
		fmt.Printf("Failed to check if retry_jobs table exists: %v\n", err)
		// Return basic stats without database
		r.mu.RLock()
		activeJobsCount := len(r.activeJobs)
		r.mu.RUnlock()

		return map[string]interface{}{
			"total_jobs":       0,
			"completed_jobs":   0,
			"failed_jobs":      0,
			"pending_jobs":     0,
			"scheduled_jobs":   0,
			"active_jobs":      activeJobsCount,
			"dead_letter_size": len(r.deadLetterQueue),
			"table_exists":     false,
		}
	}

	var totalJobs, completedJobs, failedJobs, pendingJobs, scheduledJobs int64

	r.db.Model(&RetryJob{}).Count(&totalJobs)
	r.db.Model(&RetryJob{}).Where("status = ?", "completed").Count(&completedJobs)
	r.db.Model(&RetryJob{}).Where("status = ?", "failed").Count(&failedJobs)
	r.db.Model(&RetryJob{}).Where("status = ?", "pending").Count(&pendingJobs)
	r.db.Model(&RetryJob{}).Where("status = ?", "scheduled").Count(&scheduledJobs)

	r.mu.RLock()
	activeJobsCount := len(r.activeJobs)
	r.mu.RUnlock()

	return map[string]interface{}{
		"total_jobs":       totalJobs,
		"completed_jobs":   completedJobs,
		"failed_jobs":      failedJobs,
		"pending_jobs":     pendingJobs,
		"scheduled_jobs":   scheduledJobs,
		"active_jobs":      activeJobsCount,
		"dead_letter_size": len(r.deadLetterQueue),
		"table_exists":     true,
	}
}
