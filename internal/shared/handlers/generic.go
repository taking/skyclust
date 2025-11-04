package handlers

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/responses"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Service interface for generic operations
type Service[T any, R any] interface {
	Create(ctx *gin.Context, userID uuid.UUID, req T) (R, error)
	Get(ctx *gin.Context, userID uuid.UUID, id string) (R, error)
	Update(ctx *gin.Context, userID uuid.UUID, id string, req T) (R, error)
	Delete(ctx *gin.Context, userID uuid.UUID, id string) error
	List(ctx *gin.Context, userID uuid.UUID, limit, offset int) ([]R, error)
	ListWithPagination(ctx *gin.Context, userID uuid.UUID, limit, offset int) ([]R, int64, error)
}

// AdminService interface for admin operations
type AdminService[T any, R any] interface {
	Service[T, R]
	GetAll(ctx *gin.Context, limit, offset int) ([]R, error)
	UpdateByAdmin(ctx *gin.Context, id string, req T) (R, error)
	DeleteByAdmin(ctx *gin.Context, id string) error
}

// PublicService interface for public operations
type PublicService[T any, R any] interface {
	Create(ctx *gin.Context, req T) (R, error)
	Get(ctx *gin.Context, id string) (R, error)
	List(ctx *gin.Context, limit, offset int) ([]R, error)
}

// BatchService interface for batch operations
type BatchService[T any, R any] interface {
	Service[T, R]
	CreateBatch(ctx *gin.Context, userID uuid.UUID, reqs []T) ([]R, error)
	UpdateBatch(ctx *gin.Context, userID uuid.UUID, updates map[string]T) ([]R, error)
	DeleteBatch(ctx *gin.Context, userID uuid.UUID, ids []string) error
}

// GenericHandler provides generic CRUD operations
type GenericHandler[T any, R any] struct {
	*BaseHandler
	service Service[T, R]
}

// NewGenericHandler creates a new generic handler
func NewGenericHandler[T any, R any](service Service[T, R]) *GenericHandler[T, R] {
	return &GenericHandler[T, R]{
		BaseHandler: NewBaseHandler("generic"),
		service:     service,
	}
}

// Create handles resource creation
func (h *GenericHandler[T, R]) Create(c *gin.Context) {
	HandleCreate(h.BaseHandler, "create", h.service.Create)(c)
}

// Get handles resource retrieval
func (h *GenericHandler[T, R]) Get(c *gin.Context) {
	HandleGet(h.BaseHandler, "get", h.service.Get)(c)
}

// Update handles resource updates
func (h *GenericHandler[T, R]) Update(c *gin.Context) {
	HandleUpdate(h.BaseHandler, "update", h.service.Update)(c)
}

// Delete handles resource deletion
func (h *GenericHandler[T, R]) Delete(c *gin.Context) {
	HandleDelete(h.BaseHandler, "delete", h.service.Delete)(c)
}

// List handles resource listing
func (h *GenericHandler[T, R]) List(c *gin.Context) {
	HandleList(h.BaseHandler, "list", h.service.List)(c)
}

// ListWithPagination handles paginated resource listing
func (h *GenericHandler[T, R]) ListWithPagination(c *gin.Context) {
	HandlePaginatedList(h.BaseHandler, "list_paginated", h.service.ListWithPagination)(c)
}

// SetupRoutes sets up standard CRUD routes
func (h *GenericHandler[T, R]) SetupRoutes(router *gin.RouterGroup) {
	router.POST("", h.Create)
	router.GET("/:id", h.Get)
	router.PUT("/:id", h.Update)
	router.DELETE("/:id", h.Delete)
	router.GET("", h.ListWithPagination)
}

// HandleCreate is a generic create handler
func HandleCreate[T any, R any](baseHandler *BaseHandler, operation string, serviceFunc func(*gin.Context, uuid.UUID, T) (R, error)) HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			baseHandler.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), operation)
			return
		}

		// Parse request
		var req T
		if err := baseHandler.ValidateRequest(c, &req); err != nil {
			baseHandler.HandleError(c, err, operation)
			return
		}

		// Call service
		result, err := serviceFunc(c, userID.(uuid.UUID), req)
		if err != nil {
			baseHandler.HandleError(c, err, operation)
			return
		}

		// Send response
		baseHandler.Created(c, result, "Resource created successfully")
	}
}

// HandleGet is a generic get handler
func HandleGet[T any](baseHandler *BaseHandler, operation string, serviceFunc func(*gin.Context, uuid.UUID, string) (T, error)) HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			baseHandler.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), operation)
			return
		}

		// Get ID from path
		id := c.Param("id")
		if id == "" {
			baseHandler.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "ID is required", 400), operation)
			return
		}

		// Call service
		result, err := serviceFunc(c, userID.(uuid.UUID), id)
		if err != nil {
			baseHandler.HandleError(c, err, operation)
			return
		}

		// Send response
		baseHandler.OK(c, result, "Resource retrieved successfully")
	}
}

// HandleUpdate is a generic update handler
func HandleUpdate[T any, R any](baseHandler *BaseHandler, operation string, serviceFunc func(*gin.Context, uuid.UUID, string, T) (R, error)) HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			baseHandler.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), operation)
			return
		}

		// Get ID from path
		id := c.Param("id")
		if id == "" {
			baseHandler.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "ID is required", 400), operation)
			return
		}

		// Parse request
		var req T
		if err := baseHandler.ValidateRequest(c, &req); err != nil {
			baseHandler.HandleError(c, err, operation)
			return
		}

		// Call service
		result, err := serviceFunc(c, userID.(uuid.UUID), id, req)
		if err != nil {
			baseHandler.HandleError(c, err, operation)
			return
		}

		// Send response
		baseHandler.OK(c, result, "Resource updated successfully")
	}
}

// HandleDelete is a generic delete handler
func HandleDelete(baseHandler *BaseHandler, operation string, serviceFunc func(*gin.Context, uuid.UUID, string) error) HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			baseHandler.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), operation)
			return
		}

		// Get ID from path
		id := c.Param("id")
		if id == "" {
			baseHandler.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "ID is required", 400), operation)
			return
		}

		// Call service
		err := serviceFunc(c, userID.(uuid.UUID), id)
		if err != nil {
			baseHandler.HandleError(c, err, operation)
			return
		}

		// Send response
		baseHandler.OK(c, nil, "Resource deleted successfully")
	}
}

// HandleList is a generic list handler
func HandleList[T any](baseHandler *BaseHandler, operation string, serviceFunc func(*gin.Context, uuid.UUID, int, int) ([]T, error)) HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			baseHandler.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), operation)
			return
		}

		// Parse pagination parameters
		limit, offset := baseHandler.ParsePaginationParams(c)

		// Call service
		result, err := serviceFunc(c, userID.(uuid.UUID), limit, offset)
		if err != nil {
			baseHandler.HandleError(c, err, operation)
			return
		}

		// Send response
		baseHandler.OK(c, result, "Resources retrieved successfully")
	}
}

// HandlePaginatedList is a generic paginated list handler
func HandlePaginatedList[T any](baseHandler *BaseHandler, operation string, serviceFunc func(*gin.Context, uuid.UUID, int, int) ([]T, int64, error)) HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			baseHandler.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), operation)
			return
		}

		// Parse pagination parameters
		limit, offset := baseHandler.ParsePaginationParams(c)

		// Call service
		result, total, err := serviceFunc(c, userID.(uuid.UUID), limit, offset)
		if err != nil {
			baseHandler.HandleError(c, err, operation)
			return
		}

		// Send response with pagination
		page := (offset / limit) + 1
		responses.NewResponseBuilder(c).
			WithData(result).
			WithMessage("Resources retrieved successfully").
			WithPagination(page, limit, total).
			SendOK()
	}
}
