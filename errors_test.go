package ginerr

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultErrorGenerator_IsSet(t *testing.T) {
	t.Parallel()
	// Act
	result := DefaultErrorRegistry

	// Assert
	assert.NotNil(t, result)
}

// Register functions are tested through NewErrorResponse[E error]

type aError struct {
	message string
}

func (e aError) Error() string {
	return e.message
}

type bError struct {
	message string
}

func (e bError) Error() string {
	return e.message
}

// The top ones are not parallel because it uses the DefaultErrorRegistry, which is a global

//nolint:paralleltest // Because of global state
func TestErrorResponse_UsesDefaultErrorRegistry(t *testing.T) {
	// Arrange
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	callback := func(err *aError) (int, Response) {
		return 634, Response{
			Errors: map[string]any{"error": err.Error()},
		}
	}

	err := &aError{message: "It was the man with one hand!"}

	RegisterErrorHandler(callback)

	// Act
	code, response := NewErrorResponse(err)

	// Assert
	assert.Equal(t, 634, code)
	assert.Equal(t, expectedResponse, response)
}

//nolint:paralleltest // Because of global state
func TestErrorResponse_UsesDefaultErrorRegistryForStrings(t *testing.T) {
	// Arrange
	expectedResponse := Response{
		Errors: map[string]any{"error": "my error"},
	}

	callback := func(err string) (int, Response) {
		return 123, Response{
			Errors: map[string]any{"error": err},
		}
	}

	err := errors.New("my error")

	RegisterStringErrorHandler("my error", callback)

	// Act
	code, response := NewErrorResponse(err)

	// Assert
	assert.Equal(t, 123, code)
	assert.Equal(t, expectedResponse, response)
}

//nolint:paralleltest // Because of global state
func TestErrorResponse_UsesDefaultErrorRegistryForCustomTypes(t *testing.T) {
	// Arrange
	expectedResponse := Response{
		Errors: map[string]any{"error": "my error"},
	}

	callback := func(err error) (int, Response) {
		return 123, Response{
			Errors: map[string]any{"error": err.Error()},
		}
	}

	err := errors.New("my error")

	RegisterCustomErrorTypeHandler("*errors.errorString", callback)

	// Act
	code, response := NewErrorResponse(err)

	// Assert
	assert.Equal(t, 123, code)
	assert.Equal(t, expectedResponse, response)
}

//nolint:paralleltest // Because of global state
func TestErrorResponseFrom_ReturnsGenericErrorOnNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	registry.SetDefaultResponse(123, "test")
	err := errors.New("test error")

	// Act
	code, response := NewErrorResponseFrom(registry, err)

	// Assert
	assert.Equal(t, 123, code)
	assert.Equal(t, "test", response)
}

//nolint:paralleltest // Because of global state
func TestErrorResponseFrom_ReturnsErrorA(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	callback := func(err *aError) (int, Response) {
		return 500, expectedResponse
	}

	err := &aError{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, callback)

	// Act
	code, response := NewErrorResponseFrom(registry, err)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)
}

//nolint:paralleltest // Because of global state
func TestErrorResponseFrom_ReturnsErrorB(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	callback := func(err *bError) (int, Response) {
		return 500, expectedResponse
	}

	err := &bError{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, callback)

	// Act
	code, response := NewErrorResponseFrom(registry, err)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)
}

//nolint:paralleltest // Because of global state
func TestErrorResponseFrom_ReturnsErrorBInInterface(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	callback := func(err *bError) (int, Response) {
		return 500, expectedResponse
	}

	var err error = &bError{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, callback)

	// Act
	code, response := NewErrorResponseFrom(registry, err)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)
}

func TestErrorResponseFrom_ReturnsErrorStrings(t *testing.T) {
	t.Parallel()
	tests := []string{
		"Something went completely wrong!",
		"Record not found",
	}

	for _, errorString := range tests {
		errorString := errorString
		t.Run(errorString, func(t *testing.T) {
			t.Parallel()
			// Arrange
			registry := NewErrorRegistry()
			expectedResponse := Response{
				Errors: map[string]any{"error": errorString},
			}

			callback := func(err string) (int, Response) {
				return 234, Response{
					Errors: map[string]any{"error": err},
				}
			}

			err := errors.New(errorString)

			RegisterStringErrorHandlerOn(registry, errorString, callback)

			// Act
			code, response := NewErrorResponseFrom(registry, err)

			// Assert
			assert.Equal(t, 234, code)
			assert.Equal(t, expectedResponse, response)
		})
	}
}

func TestErrorResponseFrom_CanConfigureMultipleErrorStrings(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()

	callback1 := func(err string) (int, Response) {
		return 456, Response{}
	}

	callback2 := func(err string) (int, Response) {
		return 123, Response{}
	}

	RegisterStringErrorHandlerOn(registry, "callback1", callback1)
	RegisterStringErrorHandlerOn(registry, "callback2", callback2)

	err1 := errors.New("callback1")
	err2 := errors.New("callback2")

	// Act
	code1, _ := NewErrorResponseFrom(registry, err1)
	code2, _ := NewErrorResponseFrom(registry, err2)

	// Assert
	assert.Equal(t, 456, code1)
	assert.Equal(t, 123, code2)
}

func TestErrorResponseFrom_ReturnsCustomErrorHandlers(t *testing.T) {
	t.Parallel()
	tests := []string{
		"Something went completely wrong!",
		"Record not found",
	}

	for _, errorString := range tests {
		errorString := errorString
		t.Run(errorString, func(t *testing.T) {
			t.Parallel()
			// Arrange
			registry := NewErrorRegistry()
			expectedResponse := Response{
				Errors: map[string]any{"error": errorString},
			}

			callback := func(err error) (int, Response) {
				return 234, Response{
					Errors: map[string]any{"error": err.Error()},
				}
			}

			err := errors.New(errorString)

			RegisterCustomErrorTypeHandlerOn(registry, "*errors.errorString", callback)

			// Act
			code, response := NewErrorResponseFrom(registry, err)

			// Assert
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
	code, response := NewErrorResponseFrom(registry, &bError{})

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
