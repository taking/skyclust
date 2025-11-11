package vm

import (
	"context"
	"fmt"
	"time"

	"skyclust/internal/domain"
)

// Cache key helpers
const (
	cachePrefixList = "list"
	cachePrefixItem = "item"
	cacheResourceVM = "vm"
)

// Default TTL for VM cache
const defaultVMTTL = 5 * time.Minute

// buildVMItemKey builds a cache key for individual VMs
// Format: item:vm:{vm_id}
func buildVMItemKey(vmID string) string {
	return fmt.Sprintf("%s:%s:%s", cachePrefixItem, cacheResourceVM, vmID)
}

// buildVMListKey builds a cache key for VM lists
// Format: list:vm:{workspace_id}
func buildVMListKey(workspaceID string) string {
	return fmt.Sprintf("%s:%s:%s", cachePrefixList, cacheResourceVM, workspaceID)
}

// invalidateVMCache invalidates both list and item cache for a VM
func (s *Service) invalidateVMCache(ctx context.Context, vmID, workspaceID string) {
	if s.cacheService == nil {
		return
	}

	// Invalidate list cache
	listKey := buildVMListKey(workspaceID)
	if err := s.cacheService.Delete(ctx, listKey); err != nil {
		s.logger.Warn(ctx, "Failed to invalidate VM list cache",
			domain.NewLogField("workspace_id", workspaceID),
			domain.NewLogField("error", err))
	}

	// Invalidate item cache
	itemKey := buildVMItemKey(vmID)
	if err := s.cacheService.Delete(ctx, itemKey); err != nil {
		s.logger.Warn(ctx, "Failed to invalidate VM item cache",
			domain.NewLogField("vm_id", vmID),
			domain.NewLogField("error", err))
	}
}

