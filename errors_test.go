package ginerr

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestDefaultErrorGenerator_IsSet(t *testing.T) {
	t.Parallel()
	// Act
	result := DefaultErrorRegistry

	// Assert
	assert.NotNil(t, result)
}

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

	callback := func(err *ErrorA) (int, Response) {
		return 634, Response{
			Errors: map[string]any{"error": err.Error()},
		}
	}

	err := &ErrorA{message: "It was the man with one hand!"}

	RegisterErrorHandler(callback)

	// Act
	code, response := NewErrorResponse(err)

	// Assert
	assert.Equal(t, 634, code)
	assert.Equal(t, expectedResponse, response)
}

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

func TestErrorResponseFrom_ReturnsErrorA(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	callback := func(err *ErrorA) (int, Response) {
		return 500, expectedResponse
	}

	err := &ErrorA{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, callback)

	// Act
	code, response := NewErrorResponseFrom(registry, err)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)
}

func TestErrorResponseFrom_ReturnsErrorB(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := Response{
		Errors: map[string]any{"error": "It was the man with one hand!"},
	}

	callback := func(err *ErrorB) (int, Response) {
		return 500, expectedResponse
	}

	err := &ErrorB{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, callback)

	// Act
	code, response := NewErrorResponseFrom(registry, err)

	// Assert
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

	callback := func(err *ErrorB) (int, Response) {
		return 500, expectedResponse
	}

	var err error = &ErrorB{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, callback)

	// Act
	code, response := NewErrorResponseFrom(registry, err)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)
}

func TestErrorResponseFrom_ReturnsErrorStrings(t *testing.T) {
	tests := map[string]struct {
		input string
	}{
		"first": {
			input: "Something went completely wrong!",
		},
		"second": {
			input: "Record not found",
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			// Arrange
			registry := NewErrorRegistry()
			expectedResponse := Response{
				Errors: map[string]any{"error": testData.input},
			}

			callback := func(err string) (int, Response) {
				return 234, Response{
					Errors: map[string]any{"error": err},
				}
			}

			err := errors.New(testData.input)

			RegisterStringErrorHandlerOn(registry, testData.input, callback)

			// Act
			code, response := NewErrorResponseFrom(registry, err)

			// Assert
			assert.Equal(t, 234, code)
			assert.Equal(t, expectedResponse, response)
		})
	}
}

func TestErrorResponseFrom_CanConfigureMultipleErrorStrings(t *testing.T) {
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
	tests := map[string]struct {
		input string
	}{
		"first": {
			input: "Something went completely wrong!",
		},
		"second": {
			input: "Record not found",
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			// Arrange
			registry := NewErrorRegistry()
			expectedResponse := Response{
				Errors: map[string]any{"error": testData.input},
			}

			callback := func(err error) (int, Response) {
				return 234, Response{
					Errors: map[string]any{"error": err.Error()},
				}
			}

			err := errors.New(testData.input)

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
	code, response := NewErrorResponseFrom(registry, &ErrorB{})

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
