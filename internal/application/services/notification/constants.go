package notification

// Notification Service Constants
// These constants are specific to notification operations

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeBudget  NotificationType = "budget"
	NotificationTypeVM      NotificationType = "vm"
	NotificationTypeSystem  NotificationType = "system"
)

// NotificationPriority represents notification priority levels
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityMedium   NotificationPriority = "medium"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// NotificationChannel represents delivery channels
type NotificationChannel string

const (
	ChannelInApp   NotificationChannel = "in_app"
	ChannelEmail   NotificationChannel = "email"
	ChannelBrowser NotificationChannel = "browser"
	ChannelSMS     NotificationChannel = "sms"
	ChannelWebhook NotificationChannel = "webhook"
)

// Notification status constants
const (
	NotificationStatusUnread   = "unread"
	NotificationStatusRead     = "read"
	NotificationStatusArchived = "archived"
)
