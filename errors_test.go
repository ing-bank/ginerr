package ginerr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Register functions are tested through NewErrorResponse[E error]

type AError struct {
	message string
}

func (e *AError) Error() string {
	return e.message
}

type BError struct {
	message string
}

func (e *BError) Error() string {
	return e.message
}

type dummyContextKey string

// The top ones are not parallel because it uses the DefaultErrorRegistry, which is a global

//nolint:paralleltest // Can't be used, we use global variables
func TestErrorResponse_UsesDefaultErrorRegistry(t *testing.T) {
	// Arrange
	expectedResponse := "user-friendly error"

	var calledWithError *AError
	callback := func(_ context.Context, err *AError) (int, any) {
		calledWithError = err
		return 634, expectedResponse
	}

	err := &AError{message: "It was the man with one hand!"}

	RegisterErrorHandler(&AError{}, callback)

	// Act
	code, response := NewErrorResponse(context.Background(), err)

	// Assert
	assert.Equal(t, err, calledWithError)
	assert.Equal(t, 634, code)
	assert.Equal(t, expectedResponse, response)
}

// These are parallel because it uses the 'from' variant

func TestErrorResponseFrom_ReturnsNullOnNoDefaultErrorDefined(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()

	// Act
	code, response := NewErrorResponseFrom(context.Background(), registry, assert.AnError)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Nil(t, response)
}

func TestErrorResponseFrom_UsesDefaultCallbackOnNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()

	var calledWithErr error
	var calledWithCtx context.Context
	callback := func(ctx context.Context, err error) (int, any) {
		calledWithErr = err
		calledWithCtx = ctx
		return http.StatusPaymentRequired, "abc"
	}

	registry.RegisterDefaultHandler(callback)

	// Dummy value to check whether it was passed correctly
	ctx := context.WithValue(context.Background(), dummyContextKey("abc"), "good")

	// Act
	code, response := NewErrorResponseFrom(ctx, registry, assert.AnError)

	// Assert
	assert.Equal(t, "abc", response)
	assert.Equal(t, http.StatusPaymentRequired, code)

	assert.Equal(t, ctx, calledWithCtx)
	assert.Equal(t, assert.AnError, calledWithErr)
}

func TestErrorResponseFrom_ReturnsExpectedResponsesOnErrorTypes(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()

	var calledWithErrA *AError
	var calledWithCtxA context.Context
	callbackA := func(ctx context.Context, err *AError) (int, any) {
		calledWithErrA = err
		calledWithCtxA = ctx
		return 123, "abc"
	}

	var calledWithErrB *BError
	var calledWithCtxB context.Context
	callbackB := func(ctx context.Context, err *BError) (int, any) {
		calledWithErrB = err
		calledWithCtxB = ctx
		return 456, "def"
	}

	errA := &AError{message: "It was the man with one hand!"}
	errB := &BError{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, &AError{}, callbackA)
	RegisterErrorHandlerOn(registry, &BError{}, callbackB)

	ctx := context.WithValue(context.Background(), dummyContextKey("abc"), "cool")

	// Act
	codeA, responseA := NewErrorResponseFrom(ctx, registry, errA)
	codeB, responseB := NewErrorResponseFrom(ctx, registry, errB)

	// Assert
	assert.Equal(t, 123, codeA)
	assert.Equal(t, "abc", responseA)
	assert.Equal(t, errA, calledWithErrA)
	assert.Equal(t, ctx, calledWithCtxA)

	assert.Equal(t, 456, codeB)
	assert.Equal(t, "def", responseB)
	assert.Equal(t, errB, calledWithErrB)
	assert.Equal(t, ctx, calledWithCtxB)
}

func TestErrorResponseFrom_ReturnsExpectedResponsesOnErrorStrings(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()

	var calledWithErrA error
	var calledWithCtxA context.Context
	callbackA := func(ctx context.Context, err error) (int, any) {
		calledWithErrA = err
		calledWithCtxA = ctx
		return 123, "abc"
	}

	var calledWithErrB error
	var calledWithCtxB context.Context
	callbackB := func(ctx context.Context, err error) (int, any) {
		calledWithErrB = err
		calledWithCtxB = ctx
		return 456, "def"
	}

	errA := errors.New("error A")
	errB := errors.New("error B")

	RegisterErrorHandlerOn(registry, errA, callbackA)
	RegisterErrorHandlerOn(registry, errB, callbackB)

	ctx := context.WithValue(context.Background(), dummyContextKey("abc"), "cool")

	// Act
	codeA, responseA := NewErrorResponseFrom(ctx, registry, fmt.Errorf("error A: %w", errA))
	codeB, responseB := NewErrorResponseFrom(ctx, registry, fmt.Errorf("error B: %w", errB))

	// Assert
	assert.Equal(t, 123, codeA)
	assert.Equal(t, "abc", responseA)
	assert.Equal(t, errA, calledWithErrA)
	assert.Equal(t, ctx, calledWithCtxA)

	assert.Equal(t, 456, codeB)
	assert.Equal(t, "def", responseB)
	assert.Equal(t, errB, calledWithErrB)
	assert.Equal(t, ctx, calledWithCtxB)
}

func TestErrorResponseFrom_HandlesWrappedErrorsProperly(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := "user-friendly error"

	var calledWithErr *AError
	callback := func(_ context.Context, err *AError) (int, any) {
		calledWithErr = err
		return http.StatusInternalServerError, expectedResponse
	}

	err := fmt.Errorf("this error happened: %w", &AError{message: "abc"})

	RegisterErrorHandlerOn(registry, &AError{}, callback)

	// Act
	code, response := NewErrorResponseFrom(context.Background(), registry, err)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)

	assert.Equal(t, &AError{message: "abc"}, calledWithErr)
}

func TestErrorResponseFrom_ReturnsErrorBInInterface(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()
	expectedResponse := "user-friendly error"

	var calledWithErr error
	callback := func(_ context.Context, err *BError) (int, any) {
		calledWithErr = err
		return http.StatusInternalServerError, expectedResponse
	}

	var err error = &BError{message: "It was the man with one hand!"}

	RegisterErrorHandlerOn(registry, &BError{}, callback)

	// Act
	code, response := NewErrorResponseFrom(context.Background(), registry, err)

	// Assert
	assert.Equal(t, calledWithErr, err)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, expectedResponse, response)
}

func TestErrorResponseFrom_ReturnsGenericErrorOnTypeNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewErrorRegistry()

	// Act
	code, response := NewErrorResponseFrom(context.Background(), registry, &BError{})

	// Assert
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Nil(t, response)
}
