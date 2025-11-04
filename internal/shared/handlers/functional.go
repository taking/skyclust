package handlers

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FunctionalHandler provides functional programming patterns for handlers
type FunctionalHandler struct {
	*BaseHandler
}

// NewFunctionalHandler creates a new functional handler
func NewFunctionalHandler() *FunctionalHandler {
	return &FunctionalHandler{
		BaseHandler: NewBaseHandler("functional"),
	}
}

// Result represents a result that can be either success or error
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful result
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, err: nil}
}

// Err creates an error result
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// IsOk checks if the result is successful
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr checks if the result is an error
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Value returns the value if successful
func (r Result[T]) Value() T {
	return r.value
}

// Error returns the error if failed
func (r Result[T]) Error() error {
	return r.err
}

// Map applies a function to the value if successful
func (r Result[T]) Map(fn func(T) T) Result[T] {
	if r.IsErr() {
		return r
	}
	return Ok(fn(r.value))
}

// MapErr applies a function to the error if failed
func (r Result[T]) MapErr(fn func(error) error) Result[T] {
	if r.IsOk() {
		return r
	}
	return Err[T](fn(r.err))
}

// AndThen applies a function that returns a Result if successful
func (r Result[T]) AndThen(fn func(T) Result[T]) Result[T] {
	if r.IsErr() {
		return r
	}
	return fn(r.value)
}

// OrElse returns the value if successful, otherwise returns the default value
func (r Result[T]) OrElse(defaultValue T) T {
	if r.IsOk() {
		return r.value
	}
	return defaultValue
}

// HandleResult handles a Result and sends appropriate response
func HandleResult[T any](h *BaseHandler, c *gin.Context, result Result[T], successMessage string) {
	if result.IsErr() {
		h.HandleError(c, result.Error(), "operation")
		return
	}
	h.OK(c, result.Value(), successMessage)
}

// HandleResultWithStatus handles a Result and sends response with custom status
func HandleResultWithStatus[T any](h *BaseHandler, c *gin.Context, result Result[T], statusCode int, successMessage string) {
	if result.IsErr() {
		h.HandleError(c, result.Error(), "operation")
		return
	}
	h.Success(c, statusCode, result.Value(), successMessage)
}

// ChainResult chains multiple operations that return Results
func ChainResult[T any](operations ...func() Result[T]) Result[T] {
	for _, op := range operations {
		result := op()
		if result.IsErr() {
			return result
		}
	}
	// Return the last successful result
	if len(operations) > 0 {
		return operations[len(operations)-1]()
	}
	return Err[T](domain.NewDomainError(domain.ErrCodeInternalError, "No operations provided", 500))
}

// ValidateAndExtract validates request and extracts user ID
func ValidateAndExtract[T any](h *BaseHandler, c *gin.Context, req T) Result[uuid.UUID] {
	// Validate request
	if err := h.ValidateRequest(c, &req); err != nil {
		return Err[uuid.UUID](err)
	}

	// Extract user ID
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		return Err[uuid.UUID](err)
	}

	return Ok(userID)
}

// ValidateAndExtractAdmin validates request and extracts admin user ID
func ValidateAndExtractAdmin[T any](h *BaseHandler, c *gin.Context, req T) Result[uuid.UUID] {
	// Validate request
	if err := h.ValidateRequest(c, &req); err != nil {
		return Err[uuid.UUID](err)
	}

	// Extract user ID
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		return Err[uuid.UUID](err)
	}

	// Check admin role
	userRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		return Err[uuid.UUID](err)
	}

	if userRole != domain.AdminRoleType {
		return Err[uuid.UUID](domain.NewDomainError(domain.ErrCodeForbidden, "Admin access required", 403))
	}

	return Ok(userID)
}

// SafeExecute safely executes a function and returns a Result
func SafeExecute[T any](fn func() (T, error)) Result[T] {
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error
			panicErr := domain.NewDomainError(domain.ErrCodeInternalError, "Panic occurred", 500)
			panicErr.Details = map[string]interface{}{
				"panic": r,
			}
		}
	}()

	value, err := fn()
	if err != nil {
		return Err[T](err)
	}
	return Ok(value)
}

// Compose composes multiple functions into a single function
func Compose[T any, U any, V any](f func(T) U, g func(U) V) func(T) V {
	return func(x T) V {
		return g(f(x))
	}
}

// Pipe pipes a value through multiple functions
func Pipe[T any](value T, functions ...func(T) T) T {
	result := value
	for _, fn := range functions {
		result = fn(result)
	}
	return result
}

// Maybe represents an optional value
type Maybe[T any] struct {
	value    T
	hasValue bool
}

// Some creates a Maybe with a value
func Some[T any](value T) Maybe[T] {
	return Maybe[T]{value: value, hasValue: true}
}

// None creates a Maybe without a value
func None[T any]() Maybe[T] {
	return Maybe[T]{hasValue: false}
}

// IsSome checks if Maybe has a value
func (m Maybe[T]) IsSome() bool {
	return m.hasValue
}

// IsNone checks if Maybe has no value
func (m Maybe[T]) IsNone() bool {
	return !m.hasValue
}

// Value returns the value if present
func (m Maybe[T]) Value() T {
	return m.value
}

// ValueOr returns the value if present, otherwise returns the default
func (m Maybe[T]) ValueOr(defaultValue T) T {
	if m.IsSome() {
		return m.value
	}
	return defaultValue
}

// Map applies a function to the value if present
func (m Maybe[T]) Map(fn func(T) T) Maybe[T] {
	if m.IsNone() {
		return m
	}
	return Some(fn(m.value))
}

// FlatMap applies a function that returns a Maybe if present
func (m Maybe[T]) FlatMap(fn func(T) Maybe[T]) Maybe[T] {
	if m.IsNone() {
		return m
	}
	return fn(m.value)
}
