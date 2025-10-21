package iac

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"skyclust/internal/infrastructure/database"
	"skyclust/internal/infrastructure/messaging"
)

// Execution represents an OpenTofu execution
type Execution = database.Execution

// Service defines the IaC service interface
type Service interface {
	// OpenTofu operations
	Plan(ctx context.Context, workspaceID, config string) (*Execution, error)
	Apply(ctx context.Context, workspaceID, config string) (*Execution, error)
	Destroy(ctx context.Context, workspaceID, config string) (*Execution, error)

	// Execution management
	GetExecution(ctx context.Context, workspaceID, executionID string) (*Execution, error)
	ListExecutions(ctx context.Context, workspaceID string) ([]*Execution, error)
	CancelExecution(ctx context.Context, workspaceID, executionID string) error

	// State management
	GetState(ctx context.Context, workspaceID string) (map[string]interface{}, error)
	SaveState(ctx context.Context, workspaceID string, state map[string]interface{}) error
}

// NewService creates a new IaC service
func NewService(db database.Service, eventBus messaging.Bus) Service {
	return &service{
		db:       db,
		eventBus: eventBus,
	}
}

type service struct {
	db       database.Service
	eventBus messaging.Bus
}

// Plan plans OpenTofu execution
func (s *service) Plan(ctx context.Context, workspaceID, config string) (*Execution, error) {
	execution := &database.Execution{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Command:     "plan",
		Status:      "running",
		StartedAt:   time.Now(),
	}

	// Save execution to database
	if err := s.db.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to save execution: %w", err)
	}

	// Publish event
	_ = s.eventBus.PublishToWorkspace(ctx, workspaceID, &messaging.Event{
		Type:        "tofu.execution.started",
		WorkspaceID: workspaceID,
		Data: map[string]interface{}{
			"execution_id": execution.ID,
			"command":      execution.Command,
		},
	})

	// Simulate OpenTofu execution (in production, run actual OpenTofu)
	go s.simulateExecution(ctx, execution, config)

	return (*Execution)(execution), nil
}

// Apply applies OpenTofu configuration
func (s *service) Apply(ctx context.Context, workspaceID, config string) (*Execution, error) {
	execution := &database.Execution{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Command:     "apply",
		Status:      "running",
		StartedAt:   time.Now(),
	}

	// Save execution to database
	if err := s.db.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to save execution: %w", err)
	}

	// Publish event
	_ = s.eventBus.PublishToWorkspace(ctx, workspaceID, &messaging.Event{
		Type:        "tofu.execution.started",
		WorkspaceID: workspaceID,
		Data: map[string]interface{}{
			"execution_id": execution.ID,
			"command":      execution.Command,
		},
	})

	// Simulate OpenTofu execution
	go s.simulateExecution(ctx, execution, config)

	return (*Execution)(execution), nil
}

// Destroy destroys OpenTofu resources
func (s *service) Destroy(ctx context.Context, workspaceID, config string) (*Execution, error) {
	execution := &database.Execution{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Command:     "destroy",
		Status:      "running",
		StartedAt:   time.Now(),
	}

	// Save execution to database
	if err := s.db.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to save execution: %w", err)
	}

	// Publish event
	_ = s.eventBus.PublishToWorkspace(ctx, workspaceID, &messaging.Event{
		Type:        "tofu.execution.started",
		WorkspaceID: workspaceID,
		Data: map[string]interface{}{
			"execution_id": execution.ID,
			"command":      execution.Command,
		},
	})

	// Simulate OpenTofu execution
	go s.simulateExecution(ctx, execution, config)

	return (*Execution)(execution), nil
}

// GetExecution gets a specific execution
func (s *service) GetExecution(ctx context.Context, workspaceID, executionID string) (*Execution, error) {
	execution, err := s.db.GetExecution(ctx, workspaceID, executionID)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %w", err)
	}
	return (*Execution)(execution), nil
}

// ListExecutions lists all executions for a workspace
func (s *service) ListExecutions(ctx context.Context, workspaceID string) ([]*Execution, error) {
	executions, err := s.db.ListExecutions(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list executions: %w", err)
	}
	return executions, nil
}

// CancelExecution cancels a running execution
func (s *service) CancelExecution(ctx context.Context, workspaceID, executionID string) error {
	// In production, this would cancel the actual OpenTofu process
	return s.db.UpdateExecutionStatus(ctx, workspaceID, executionID, "cancelled")
}

// GetState gets the current state
func (s *service) GetState(ctx context.Context, workspaceID string) (map[string]interface{}, error) {
	state, err := s.db.GetState(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}
	return state, nil
}

// SaveState saves the current state
func (s *service) SaveState(ctx context.Context, workspaceID string, state map[string]interface{}) error {
	if err := s.db.SaveState(ctx, workspaceID, state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	return nil
}

// simulateExecution simulates OpenTofu execution
func (s *service) simulateExecution(ctx context.Context, execution *Execution, config string) {
	// Simulate execution time
	time.Sleep(5 * time.Second)

	// Update execution status
	execution.Status = "completed"
	now := time.Now()
	execution.CompletedAt = &now
	execution.Output = fmt.Sprintf("OpenTofu %s completed successfully", execution.Command)

	// Save to database
	_ = s.db.UpdateExecution(ctx, execution)

	// Publish completion event
	_ = s.eventBus.PublishToWorkspace(ctx, execution.WorkspaceID, &messaging.Event{
		Type:        "tofu.execution.completed",
		WorkspaceID: execution.WorkspaceID,
		Data: map[string]interface{}{
			"execution_id": execution.ID,
			"command":      execution.Command,
			"status":       execution.Status,
		},
	})
}

// generateID generates a random ID
