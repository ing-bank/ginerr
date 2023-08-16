package ginerr

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

// Register functions are tested through NewErrorResponse[E error]

type ErrorA struct {
	message string
}

func (e ErrorA) Error() string {
	return e.message
}

type ErrorB struct {
	message string
}

func (e ErrorB) Error() string {
	return e.message
}

// The top ones are not parallel because it uses the DefaultErrorRegistry, which is a global

func TestErrorResponse_UsesDefaultErrorRegistry(t *testing.T) {
	// Arrange
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	var calledWithError *ErrorA
	callback := func(ctx context.Context, err *ErrorA) (int, any) {
		calledWithError = err
		return 634, Response{
			Errors: map[string]any{"error": err.Error()},
		}
	}

	err := &ErrorA{message: "It was the man with one hand!"}

	RegisterErrorHandler(callback)

	// Act
	code, response := NewErrorResponse(context.Background(), err)

	// Assert
	assert.Equal(t, err, calledWithError)
	assert.Equal(t, 634, code)
	assert.Equal(t, expectedResponse, response)
}

func TestErrorResponse_UsesDefaultErrorRegistryForStrings(t *testing.T) {
	// Arrange
	expectedResponse := Response{
		Errors: map[string]any{"error": "my error"},
	}

	var calledWithError string
	var calledWithContext context.Context
	callback := func(ctx context.Context, err string) (int, any) {
		calledWithError = err
		calledWithContext = ctx
		return 123, Response{
			Errors: map[string]any{"error": err},
		}
	}

	err := errors.New("my error")

	RegisterStringErrorHandler("my error", callback)

	ctx := context.WithValue(context.Background(), ErrorA{}, "anything")

	// Act
	code, response := NewErrorResponse(ctx, err)

	// Assert
	assert.Equal(t, ctx, calledWithContext)
	assert.Equal(t, err.Error(), calledWithError)
	assert.Equal(t, 123, code)
	assert.Equal(t, expectedResponse, response)
}

func TestErrorResponse_UsesDefaultErrorRegistryForCustomTypes(t *testing.T) {
	// Arrange
	expectedResponse := Response{
		Errors: map[string]any{"error": "assert.AnError general error for testing"},
	}

	var calledWithError error
	callback := func(ctx context.Context, err error) (int, any) {
		calledWithError = err
		return 123, Response{
			Errors: map[string]any{"error": err.Error()},
		}
	}

	RegisterCustomErrorTypeHandler("*errors.errorString", callback)

	// Act
	code, response := NewErrorResponse(context.Background(), assert.AnError)

	// Assert
	assert.Equal(t, assert.AnError, calledWithError)
	assert.Equal(t, 123, code)
	assert.Equal(t, expectedResponse, response)
}

// These are parallel because it uses the 'from' variant

func TestErrorResponseFrom_ReturnsGenericErrorOnNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	registry.SetDefaultResponse(123, "test")

	// Act
	code, response := NewErrorResponseFrom(registry, context.Background(), assert.AnError)

	// Assert
	assert.Equal(t, 123, code)
	assert.Equal(t, "test", response)
}

func TestErrorResponseFrom_UsesDefaultCallbackOnNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()

	expectedResponse := Response{
		Errors: map[string]any{"error": "internal server error"},
	}

	var calledWithErr error
	var calledWithCtx context.Context
	callback := func(ctx context.Context, err error) (int, any) {
		calledWithErr = err
		calledWithCtx = ctx
		return http.StatusInternalServerError, expectedResponse
	}

	registry.RegisterDefaultHandler(callback)

	ctx := context.WithValue(context.Background(), ErrorA{}, "good")

	// Act
	code, response := NewErrorResponseFrom(registry, ctx, assert.AnError)

	// Assert
	assert.Equal(t, expectedResponse, response)
	assert.Equal(t, code, http.StatusInternalServerError)

	assert.Equal(t, ctx, calledWithCtx)
	assert.Equal(t, assert.AnError, calledWithErr)
}

func TestErrorResponseFrom_ReturnsErrorAWithContext(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	var calledWithErr *ErrorA
	var calledWithCtx context.Context
	callback := func(ctx context.Context, err *ErrorA) (int, any) {
		calledWithErr = err
		calledWithCtx = ctx
		return http.StatusInternalServerError, expectedResponse
	}

	err := &ErrorA{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, callback)

	ctx := context.WithValue(context.Background(), ErrorA{}, "cool")

	// Act
	code, response := NewErrorResponseFrom(registry, ctx, err)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)

	assert.Equal(t, err, calledWithErr)
	assert.Equal(t, ctx, calledWithCtx)
}

func TestErrorResponseFrom_ReturnsErrorB(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	var calledWithErr *ErrorB
	callback := func(ctx context.Context, err *ErrorB) (int, any) {
		calledWithErr = err
		return http.StatusInternalServerError, expectedResponse
	}

	err := &ErrorB{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, callback)

	// Act
	code, response := NewErrorResponseFrom(registry, context.Background(), err)

	// Assert
	assert.Equal(t, calledWithErr, err)

	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)
}

func TestErrorResponseFrom_ReturnsErrorBInInterface(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	var calledWithErr error
	callback := func(ctx context.Context, err *ErrorB) (int, any) {
		calledWithErr = err
		return http.StatusInternalServerError, expectedResponse
	}

	var err error = &ErrorB{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, callback)

	// Act
	code, response := NewErrorResponseFrom(registry, context.Background(), err)

	// Assert
	assert.Equal(t, calledWithErr, err)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)
}

func TestErrorResponseFrom_ReturnsErrorStrings(t *testing.T) {
	tests := []string{
		"Something went completely wrong!",
		"Record not found",
	}

	for _, errorString := range tests {
		errorString := errorString
		t.Run(errorString, func(t *testing.T) {
			// Arrange
			registry := NewErrorRegistry()
			expectedResponse := Response{
				Errors: map[string]any{"error": errorString},
			}

			var calledWithContext context.Context
			var calledWithErr string
			callback := func(ctx context.Context, err string) (int, any) {
				calledWithErr = err
				calledWithContext = ctx
				return 234, Response{
					Errors: map[string]any{"error": err},
				}
			}

			err := errors.New(errorString)

			RegisterStringErrorHandlerOn(registry, errorString, callback)

			ctx := context.WithValue(context.Background(), ErrorA{}, "good")

			// Act
			code, response := NewErrorResponseFrom(registry, ctx, err)

			// Assert
			assert.Equal(t, ctx, calledWithContext)
			assert.Equal(t, err.Error(), calledWithErr)
			assert.Equal(t, 234, code)
			assert.Equal(t, expectedResponse, response)
		})
	}
}

func TestErrorResponseFrom_CanConfigureMultipleErrorStrings(t *testing.T) {
	// Arrange
	registry := NewErrorRegistry()

	var callback1CalledWithString string
	var callback1CalledWithContext context.Context
	callback1 := func(ctx context.Context, err string) (int, any) {
		callback1CalledWithContext = ctx
		callback1CalledWithString = err
		return 456, Response{}
	}

	var callback2CalledWithString string
	var callback2CalledWithContext context.Context
	callback2 := func(ctx context.Context, err string) (int, any) {
		callback2CalledWithContext = ctx
		callback2CalledWithString = err
		return 123, Response{}
	}

	RegisterStringErrorHandlerOn(registry, "callback1", callback1)
	RegisterStringErrorHandlerOn(registry, "callback2", callback2)

	err1 := errors.New("callback1")
	err2 := errors.New("callback2")

	ctx := context.WithValue(context.Background(), ErrorA{}, "anything")

	// Act
	code1, _ := NewErrorResponseFrom(registry, ctx, err1)
	code2, _ := NewErrorResponseFrom(registry, ctx, err2)

	// Assert
	assert.Equal(t, err1.Error(), callback1CalledWithString)
	assert.Equal(t, err2.Error(), callback2CalledWithString)

	assert.Equal(t, ctx, callback1CalledWithContext)
	assert.Equal(t, ctx, callback2CalledWithContext)

	assert.Equal(t, 456, code1)
	assert.Equal(t, 123, code2)
}

func TestErrorResponseFrom_ReturnsCustomErrorHandlers(t *testing.T) {
	tests := []string{
		"Something went completely wrong!",
		"Record not found",
	}

	for _, errorString := range tests {
		errorString := errorString
		t.Run(errorString, func(t *testing.T) {
			// Arrange
			registry := NewErrorRegistry()
			expectedResponse := Response{
				Errors: map[string]any{"error": errorString},
			}

			var calledWithContext context.Context
			var calledWithErr error
			callback := func(ctx context.Context, err error) (int, any) {
				calledWithErr = err
				calledWithContext = ctx
				return 234, Response{
					Errors: map[string]any{"error": err.Error()},
				}
			}

			err := errors.New(errorString)

			RegisterCustomErrorTypeHandlerOn(registry, "*errors.errorString", callback)

			ctx := context.WithValue(context.Background(), ErrorA{}, "good")

			// Act
			code, response := NewErrorResponseFrom(registry, ctx, err)

			// Assert
			assert.Equal(t, ctx, calledWithContext)
			assert.Equal(t, err, calledWithErr)
			assert.Equal(t, 234, code)
			assert.Equal(t, expectedResponse, response)
		})
	}
}

func TestErrorResponseFrom_ReturnsGenericErrorOnTypeNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()

	// Act
	code, response := NewErrorResponseFrom(registry, context.Background(), &ErrorB{})

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Nil(t, response)
}

func TestErrorRegistry_SetDefaultResponse_SetsProperties(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()

	code := 123
	response := Response{Errors: map[string]any{"error": "Something went wrong"}}

	// Act
	registry.SetDefaultResponse(123, response)

	// Assert
	assert.Equal(t, code, registry.DefaultCode)
	assert.Equal(t, response, registry.DefaultResponse)
}
