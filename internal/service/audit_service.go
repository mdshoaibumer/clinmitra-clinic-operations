package service

import (
	"encoding/json"
	"time"

	"clinmitra/internal/models"
	"clinmitra/internal/repository"

	"github.com/google/uuid"
)

// AuditService provides audit trail functionality for tracking user actions
// on entities. All logging is best-effort to avoid blocking parent operations.
type AuditService struct {
	auditRepo    repository.AuditRepository
	maxValueSize int // max bytes for old/new JSON values
}

// NewAuditService creates an AuditService with an 8KB cap on serialized
// old/new values to prevent excessively large audit records.
func NewAuditService(auditRepo repository.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo:    auditRepo,
		maxValueSize: 8192, // 8KB cap per value — sufficient for any single entity
	}
}

// LogAction records an audit entry for a user action. Old and new values are
// JSON-serialized and truncated to maxValueSize. Errors are silently ignored
// so the parent operation is never affected.
func (s *AuditService) LogAction(userID string, action models.AuditAction, entityType, entityID string, oldValue, newValue interface{}) {
	log := &models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		CreatedAt:  time.Now(),
	}

	if oldValue != nil {
		if data, err := json.Marshal(oldValue); err == nil {
			log.OldValue = truncateJSON(string(data), s.maxValueSize)
		}
	}
	if newValue != nil {
		if data, err := json.Marshal(newValue); err == nil {
			log.NewValue = truncateJSON(string(data), s.maxValueSize)
		}
	}

	// Best effort - run in background to avoid deadlocking if caller is in a transaction
	// with MaxOpenConns=1.
	go func() {
		_ = s.auditRepo.Create(log)
	}()
}

// truncateJSON safely truncates a JSON string to maxLen bytes.
// If truncated, appends a marker so reviewers know data was clipped.
func truncateJSON(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-15] + "...[TRUNCATED]"
}

// GetEntityHistory returns all audit log entries for a specific entity.
func (s *AuditService) GetEntityHistory(entityType, entityID string) ([]models.AuditLog, error) {
	return s.auditRepo.ListByEntity(entityType, entityID)
}

// GetRecentActivity returns the most recent audit log entries up to the given limit.
func (s *AuditService) GetRecentActivity(limit int) ([]models.AuditLog, error) {
	return s.auditRepo.ListRecent(limit)
}
