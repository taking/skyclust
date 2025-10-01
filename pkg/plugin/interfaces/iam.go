package interfaces

import "context"

// IAMProvider defines the interface for Identity and Access Management
type IAMProvider interface {
	// GetName returns the name of the IAM provider
	GetName() string

	// GetVersion returns the version of the IAM provider
	GetVersion() string

	// Initialize initializes the IAM provider with configuration
	Initialize(config map[string]interface{}) error

	// User Management
	CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	ListUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, userID string) error

	// Group Management
	CreateGroup(ctx context.Context, req CreateGroupRequest) (*Group, error)
	GetGroup(ctx context.Context, groupID string) (*Group, error)
	ListGroups(ctx context.Context) ([]Group, error)
	UpdateGroup(ctx context.Context, groupID string, req UpdateGroupRequest) (*Group, error)
	DeleteGroup(ctx context.Context, groupID string) error

	// Role Management
	CreateRole(ctx context.Context, req CreateRoleRequest) (*Role, error)
	GetRole(ctx context.Context, roleID string) (*Role, error)
	ListRoles(ctx context.Context) ([]Role, error)
	UpdateRole(ctx context.Context, roleID string, req UpdateRoleRequest) (*Role, error)
	DeleteRole(ctx context.Context, roleID string) error

	// Policy Management
	CreatePolicy(ctx context.Context, req CreatePolicyRequest) (*Policy, error)
	GetPolicy(ctx context.Context, policyID string) (*Policy, error)
	ListPolicies(ctx context.Context) ([]Policy, error)
	UpdatePolicy(ctx context.Context, policyID string, req UpdatePolicyRequest) (*Policy, error)
	DeletePolicy(ctx context.Context, policyID string) error

	// Permission Management
	AttachPolicyToUser(ctx context.Context, userID, policyID string) error
	DetachPolicyFromUser(ctx context.Context, userID, policyID string) error
	AttachPolicyToGroup(ctx context.Context, groupID, policyID string) error
	DetachPolicyFromGroup(ctx context.Context, groupID, policyID string) error
	AttachPolicyToRole(ctx context.Context, roleID, policyID string) error
	DetachPolicyFromRole(ctx context.Context, roleID, policyID string) error

	// Access Key Management
	CreateAccessKey(ctx context.Context, userID string) (*AccessKey, error)
	ListAccessKeys(ctx context.Context, userID string) ([]AccessKey, error)
	DeleteAccessKey(ctx context.Context, userID, keyID string) error
}

// User represents a user in the IAM system
type User struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	DisplayName string            `json:"display_name"`
	Status      string            `json:"status"`
	CreatedAt   string            `json:"created_at"`
	LastLogin   string            `json:"last_login,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// Group represents a group in the IAM system
type Group struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedAt   string            `json:"created_at"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// Role represents a role in the IAM system
type Role struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"` // service, user, etc.
	CreatedAt   string            `json:"created_at"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// Policy represents a policy in the IAM system
type Policy struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Document    string            `json:"document"` // JSON policy document
	Version     string            `json:"version"`
	CreatedAt   string            `json:"created_at"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// AccessKey represents an access key for a user
type AccessKey struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key,omitempty"` // Only returned on creation
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// Request/Response types
type CreateUserRequest struct {
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	DisplayName string            `json:"display_name"`
	Password    string            `json:"password,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type UpdateUserRequest struct {
	Email       string            `json:"email,omitempty"`
	DisplayName string            `json:"display_name,omitempty"`
	Status      string            `json:"status,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type CreateGroupRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type UpdateGroupRequest struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type CreateRoleRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type UpdateRoleRequest struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type CreatePolicyRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Document    string            `json:"document"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type UpdatePolicyRequest struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Document    string            `json:"document,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

