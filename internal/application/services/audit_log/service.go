package audit_log

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	exportservice "skyclust/internal/application/services/export"
	"skyclust/internal/domain"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service implements the audit log business logic
type Service struct {
	auditLogRepo domain.AuditLogRepository
}

// NewService creates a new audit log service
func NewService(auditLogRepo domain.AuditLogRepository) domain.AuditLogService {
	return &Service{
		auditLogRepo: auditLogRepo,
	}
}

// LogAction: 사용자 액션을 로깅합니다
func (s *Service) LogAction(userID uuid.UUID, action, resource string, details map[string]interface{}) error {
	log := &domain.AuditLog{
		UserID:   userID,
		Action:   action,
		Resource: resource,
		Details:  domain.JSONBMap(details),
	}
	return s.auditLogRepo.Create(log)
}

// GetUserLogs: 사용자의 감사 로그 항목을 페이지네이션과 함께 조회합니다
func (s *Service) GetUserLogs(userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditLogRepo.GetByUserID(userID, limit, offset)
}

// GetLogsByAction: 액션별 감사 로그 항목을 페이지네이션과 함께 조회합니다
func (s *Service) GetLogsByAction(action string, limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditLogRepo.GetByAction(action, limit, offset)
}

// GetLogsByDateRange: 날짜 범위별 감사 로그 항목을 페이지네이션과 함께 조회합니다
func (s *Service) GetLogsByDateRange(start, end time.Time, limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditLogRepo.GetByDateRange(start, end, limit, offset)
}

// CleanupOldLogs: 오래된 감사 로그 항목을 제거합니다
func (s *Service) CleanupOldLogs(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	_, err := s.auditLogRepo.DeleteOldLogs(cutoffDate)
	return err
}

// GetAuditLogs: 필터를 사용하여 감사 로그를 조회합니다 (관리자 메서드)
func (s *Service) GetAuditLogs(filters domain.AuditLogFilters) ([]*domain.AuditLog, int64, error) {
	// Use existing repository methods based on filters
	var logs []*domain.AuditLog
	var total int64
	var err error

	// Calculate offset from page
	offset := 0
	if filters.Page > 0 && filters.Limit > 0 {
		offset = (filters.Page - 1) * filters.Limit
	}

	// Set default limit if not provided
	limit := filters.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	// Apply filters based on what's available
	if filters.UserID != nil && filters.Action != "" {
		// Get by user ID first, then filter by action in memory (could be optimized)
		logs, err = s.auditLogRepo.GetByUserID(*filters.UserID, limit*10, offset) // Get more to filter
		if err != nil {
			return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get audit logs: %v", err), 500)
		}
		// Filter by action
		filteredLogs := make([]*domain.AuditLog, 0)
		for _, log := range logs {
			if log.Action == filters.Action {
				filteredLogs = append(filteredLogs, log)
				if len(filteredLogs) >= limit {
					break
				}
			}
		}
		logs = filteredLogs
		total, _ = s.auditLogRepo.CountByUserID(*filters.UserID)
	} else if filters.UserID != nil {
		logs, err = s.auditLogRepo.GetByUserID(*filters.UserID, limit, offset)
		if err != nil {
			return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get audit logs: %v", err), 500)
		}
		total, err = s.auditLogRepo.CountByUserID(*filters.UserID)
		if err != nil {
			return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to count audit logs: %v", err), 500)
		}
	} else if filters.Action != "" {
		logs, err = s.auditLogRepo.GetByAction(filters.Action, limit, offset)
		if err != nil {
			return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get audit logs: %v", err), 500)
		}
		total, err = s.auditLogRepo.CountByAction(filters.Action)
		if err != nil {
			return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to count audit logs: %v", err), 500)
		}
	} else if filters.StartTime != nil && filters.EndTime != nil {
		logs, err = s.auditLogRepo.GetByDateRange(*filters.StartTime, *filters.EndTime, limit, offset)
		if err != nil {
			return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get audit logs: %v", err), 500)
		}
		// Count logs in date range (approximation - would need repository method)
		total = int64(len(logs)) // Placeholder, would need CountByDateRange
	} else {
		// No filters - return empty (should require at least one filter for security)
		return []*domain.AuditLog{}, 0, domain.NewDomainError(domain.ErrCodeBadRequest, "at least one filter is required", 400)
	}

	return logs, total, nil
}

// GetAuditLogByID: ID로 특정 감사 로그를 조회합니다 (관리자 메서드)
func (s *Service) GetAuditLogByID(id uuid.UUID) (*domain.AuditLog, error) {
	log, err := s.auditLogRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return log, nil
}

// GetAuditStats: 감사 로그 통계를 조회합니다 (관리자 메서드)
func (s *Service) GetAuditStats(filters domain.AuditStatsFilters) (*domain.AuditStats, error) {
	// Get total events count
	totalEvents, err := s.auditLogRepo.GetTotalCount(filters)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get total events count: %v", err), 500)
	}

	// Get unique users count
	uniqueUsers, err := s.auditLogRepo.GetUniqueUsersCount(filters)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get unique users count: %v", err), 500)
	}

	// Get top actions (limit to top 10)
	topActions, err := s.auditLogRepo.GetTopActions(filters, 10)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get top actions: %v", err), 500)
	}

	// Get top resources (limit to top 10)
	topResources, err := s.auditLogRepo.GetTopResources(filters, 10)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get top resources: %v", err), 500)
	}

	// Get events by day
	eventsByDay, err := s.auditLogRepo.GetEventsByDay(filters)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get events by day: %v", err), 500)
	}

	return &domain.AuditStats{
		TotalEvents:  totalEvents,
		UniqueUsers:  uniqueUsers,
		TopActions:   topActions,
		TopResources: topResources,
		EventsByDay:  eventsByDay,
	}, nil
}

// ExportAuditLogs: 다양한 형식으로 감사 로그를 내보냅니다 (관리자 메서드)
func (s *Service) ExportAuditLogs(filters domain.AuditLogFilters, format string) ([]byte, error) {
	// Validate date range if both are provided
	if filters.StartTime != nil && filters.EndTime != nil {
		if filters.StartTime.After(*filters.EndTime) {
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "start_time must be before or equal to end_time", 400)
		}
	}

	// Set reasonable limit for export
	if filters.Limit <= 0 || filters.Limit > exportservice.MaxExportRecords {
		filters.Limit = exportservice.MaxExportRecords
	}
	filters.Page = 1 // Start from first page

	// Get filtered audit logs
	logs, _, err := s.GetAuditLogs(filters)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get audit logs: %v", err), 500)
	}

	// Handle empty results
	if len(logs) == 0 {
		switch strings.ToLower(format) {
		case "csv":
			// Return CSV with header only
			var buf strings.Builder
			writer := csv.NewWriter(&buf)
			header := []string{"ID", "User ID", "Action", "Resource", "IP Address", "User Agent", "Details", "Created At"}
			if err := writer.Write(header); err != nil {
				return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to write CSV header: %v", err), 500)
			}
			writer.Flush()
			return []byte(buf.String()), nil
		case "json":
			// Return empty JSON array
			return []byte("[]"), nil
		default:
			return []byte{}, nil
		}
	}

	// Convert to requested format
	switch strings.ToLower(format) {
	case "csv":
		return s.exportAuditLogsToCSV(logs)
	case "json":
		return s.exportAuditLogsToJSON(logs)
	case "xlsx":
		// XLSX not implemented yet
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "XLSX export not implemented", 501)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported export format: %s. Supported formats: json, csv", format), 400)
	}
}

// exportAuditLogsToCSV: 감사 로그를 CSV 형식으로 내보냅니다
func (s *Service) exportAuditLogsToCSV(logs []*domain.AuditLog) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "User ID", "Action", "Resource", "IP Address", "User Agent", "Details", "Created At"}
	if err := writer.Write(header); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to write CSV header: %v", err), 500)
	}

	// Write data
	for _, log := range logs {
		// Convert details map to JSON string
		detailsStr := ""
		if len(log.Details) > 0 {
			if detailsJSON, err := json.Marshal(log.Details); err == nil {
				detailsStr = string(detailsJSON)
			}
		}

		record := []string{
			log.ID.String(),
			log.UserID.String(),
			log.Action,
			log.Resource,
			log.IPAddress,
			log.UserAgent,
			detailsStr,
			log.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to write CSV record: %v", err), 500)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to flush CSV: %v", err), 500)
	}

	return []byte(buf.String()), nil
}

// exportAuditLogsToJSON: 감사 로그를 JSON 형식으로 내보냅니다
func (s *Service) exportAuditLogsToJSON(logs []*domain.AuditLog) ([]byte, error) {
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal JSON: %v", err), 500)
	}
	return data, nil
}

// CleanupAuditLogs: 보존 정책에 따라 오래된 감사 로그를 제거합니다 (관리자 메서드)
func (s *Service) CleanupAuditLogs(retentionDays int) (int64, error) {
	if retentionDays < 0 {
		return 0, domain.NewDomainError(domain.ErrCodeValidationFailed, "retention days must be non-negative", 400)
	}

	if retentionDays == 0 {
		return 0, domain.NewDomainError(domain.ErrCodeValidationFailed, "retention days must be greater than 0 to prevent accidental deletion of all logs", 400)
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Delete old logs and get actual count of deleted logs
	deletedCount, err := s.auditLogRepo.DeleteOldLogs(cutoffDate)
	if err != nil {
		return 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to cleanup audit logs: %v", err), 500)
	}

	return deletedCount, nil
}

// GetAuditLogSummary: 감사 로그 요약을 조회합니다 (관리자 메서드)
func (s *Service) GetAuditLogSummary(startTime, endTime time.Time) (*domain.AuditLogSummary, error) {
	filters := domain.AuditStatsFilters{
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	// Get total events count
	totalEvents, err := s.auditLogRepo.GetTotalCount(filters)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get total events count: %v", err), 500)
	}

	// Get unique users count
	uniqueUsers, err := s.auditLogRepo.GetUniqueUsersCount(filters)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get unique users count: %v", err), 500)
	}

	// Get most active user
	mostActiveUserID, _, err := s.auditLogRepo.GetMostActiveUser(startTime, endTime)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get most active user: %v", err), 500)
	}

	mostActiveUserStr := ""
	if mostActiveUserID != uuid.Nil {
		mostActiveUserStr = mostActiveUserID.String()
	}

	// Get top actions (limit to top 5 for summary)
	topActions, err := s.auditLogRepo.GetTopActions(filters, 5)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get top actions: %v", err), 500)
	}

	// Get security events count
	securityEvents, err := s.auditLogRepo.GetSecurityEventsCount(startTime, endTime)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get security events count: %v", err), 500)
	}

	// Get error events count
	errorEvents, err := s.auditLogRepo.GetErrorEventsCount(startTime, endTime)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get error events count: %v", err), 500)
	}

	return &domain.AuditLogSummary{
		TotalEvents:    totalEvents,
		UniqueUsers:    uniqueUsers,
		MostActiveUser: mostActiveUserStr,
		TopActions:     topActions,
		SecurityEvents: securityEvents,
		ErrorEvents:    errorEvents,
	}, nil
}
