package ginerr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// DefaultErrorRegistry is a global singleton empty ErrorRegistry for convenience.
var DefaultErrorRegistry = NewErrorRegistry()

// errorHandler encompasses the methods necessary to validate an error and calculate a response.
type errorHandler struct {
	// isStringError is used to catch cases where errors.New() is being used as error types,
	// as we need to use errors.Is for those cases, errors.As is not enough
	isStringError bool

	// isType is a wrapped around an `errors.As` from RegisterERrorHandler with the type information
	// of the target error still intact.
	isType func(err error) bool

	// handle will calculate the response. It's a wrapper around the user-provided handler
	// which ensures that the type of the error is properly asserted using `errors.As`.
	handle func(ctx context.Context, err error) (int, any)
}

// NewErrorRegistry instantiates a new ErrorRegistry. If you're looking for the 'default' error
// registry, check out DefaultErrorRegistry.
func NewErrorRegistry() *ErrorRegistry {
	registry := &ErrorRegistry{
		handlers: make(map[error]*errorHandler),
		defaultHandler: func(context.Context, error) (int, any) {
			return http.StatusInternalServerError, nil
		},
	}

	return registry
}

// ErrorRegistry is the place where errors and callbacks are stored.
type ErrorRegistry struct {
	// handlers maps error types with their handlers
	handlers map[error]*errorHandler

	// defaultHandler is called if no matching error was registered
	defaultHandler func(ctx context.Context, err error) (int, any)
}

func (e *ErrorRegistry) RegisterDefaultHandler(callback func(ctx context.Context, err error) (int, any)) {
	e.defaultHandler = callback
}

// NewErrorResponse Returns an error response using the DefaultErrorRegistry. If no specific handler could be found,
// it will return the defaults.
func NewErrorResponse(ctx context.Context, err error) (int, any) {
	return NewErrorResponseFrom(ctx, DefaultErrorRegistry, err)
}

// NewErrorResponseFrom Returns an error response using the given registry. If no specific handler could be found,
// it will return the defaults.
func NewErrorResponseFrom[E error](ctx context.Context, registry *ErrorRegistry, err E) (int, any) {
	for errConcrete, handler := range registry.handlers {
		// We can't use `errors.As` here directly, as we don't have a concrete version of the type here
		if !handler.isType(err) {
			continue
		}

		// If it's a string error, it must match the given error exactly, otherwise it might mix up if we only
		// check on type
		if handler.isStringError {
			if errors.Is(err, errConcrete) {
				// It might be wrapped, so we pass the concrete type
				return handler.handle(ctx, errConcrete)
			}

			continue
		}

		return handler.handle(ctx, err)
	}

	return registry.defaultHandler(ctx, err)
}

// RegisterErrorHandler registers an error handler in DefaultErrorRegistry.
func RegisterErrorHandler[E error](instance E, handler func(context.Context, E) (int, any)) {
	RegisterErrorHandlerOn(DefaultErrorRegistry, instance, handler)
}

// errorStringType is used to check if an error was created by errors.New or fmt.Errorf
//
//nolint:err113 // We need it here for the type name
var errorStringType = fmt.Sprintf("%T", errors.New(""))

// RegisterErrorHandlerOn registers an error handler in the given registry.
func RegisterErrorHandlerOn[E error](registry *ErrorRegistry, instance E, handler func(context.Context, E) (int, any)) {
	// Wrap it in a closure, we can't save it directly because err E is not available in NewErrorResponseFrom. It will
	// be available in the closure when it is called. Check out TestErrorResponseFrom_ReturnsErrorBInInterface for an example.
	registry.handlers[instance] = &errorHandler{
		// Necessary to make sure we match error strings using `errors.Is`
		isStringError: fmt.Sprintf("%T", instance) == errorStringType,

		// Handler that uses errors.As to cast to an error
		handle: func(ctx context.Context, err error) (int, any) {
			var errorOfType E

			// This function should only be called if errors.Is succeeded, so this should never fail
			_ = errors.As(err, &errorOfType)

			return handler(ctx, errorOfType)
		},

		// Type check, as we need `instance` from this function
		isType: func(err error) bool {
			return errors.As(err, &instance)
		},
	}
}
