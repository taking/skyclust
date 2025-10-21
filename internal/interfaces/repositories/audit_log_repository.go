package repositories

import (
	"context"
	"skyclust/internal/domain"

	"github.com/google/uuid"
)

// AuditLogRepository defines the interface for audit log data operations
type AuditLogRepository interface {
	// Basic CRUD operations
	Create(auditLog *domain.AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error)
	Update(auditLog *domain.AuditLog) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List operations
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, error)
	GetByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*domain.AuditLog, error)
	List(limit, offset int) ([]*domain.AuditLog, error)
	Count() (int64, error)

	// Search operations
	Search(query string, limit, offset int) ([]*domain.AuditLog, error)
	GetByAction(action string, limit, offset int) ([]*domain.AuditLog, error)
	GetByDateRange(startDate, endDate string, limit, offset int) ([]*domain.AuditLog, error)
}
