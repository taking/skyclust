package sse

import "time"

// SSE event types
const (
	EventTypeSystemNotification = "system-notification"
	EventTypeSystemAlert        = "system-alert"
	EventTypeVMStatus           = "vm-status"
	EventTypeVMResource         = "vm-resource"
	EventTypeVMError            = "vm-error"
	EventTypeProviderStatus     = "provider-status"
	EventTypeProviderInstance   = "provider-instance"
)

// SSE timing constants
const (
	CleanupInterval = 30 * time.Second
	ClientTimeout   = 5 * time.Minute
)

// SSE data field names
const (
	FieldVMID     = "vmId"
	FieldProvider = "provider"
)

// Error messages
const (
	ErrClientNotFound = "client not found"
)
