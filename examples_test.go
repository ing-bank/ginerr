package ginerr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

var ErrDatabaseOverloaded = errors.New("database overloaded")

type InputValidationError struct {
	// ...
}

func (m *InputValidationError) Error() string {
	return "..."
}

func ExampleRegisterErrorHandler() {
	// Write your error handlers
	validationHandler := func(_ context.Context, err *InputValidationError) (int, any) {
		return http.StatusBadRequest, "Your input was invalid: " + err.Error()
	}
	databaseOverloadedHandler := func(context.Context, error) (int, any) {
		return http.StatusBadGateway, "Please try again later"
	}

	// Register the error handlers and supply an empty version of the type for type reference
	RegisterErrorHandler(&InputValidationError{}, validationHandler)
	RegisterErrorHandler(ErrDatabaseOverloaded, databaseOverloadedHandler)

	// Return errors somewhere deep in your code
	errA := fmt.Errorf("validation error: %w", &InputValidationError{})
	errB := fmt.Errorf("could not connect to database: %w", ErrDatabaseOverloaded)

	// In your HTTP handlers, instantiate responses and return those to the users
	codeA, responseA := NewErrorResponse(context.Background(), errA)
	codeB, responseB := NewErrorResponse(context.Background(), errB)

	// Check the output
	fmt.Printf("%d: %s\n", codeA, responseA)
	fmt.Printf("%d: %s\n", codeB, responseB)

	// Output:
	// 400: Your input was invalid: ...
	// 502: Please try again later
}

func ExampleRegisterErrorHandlerOn() {
	registry := NewErrorRegistry()

	// Write your error handlers
	validationHandler := func(_ context.Context, err *InputValidationError) (int, any) {
		return http.StatusBadRequest, "Your input was invalid: " + err.Error()
	}
	databaseOverloadedHandler := func(context.Context, error) (int, any) {
		return http.StatusBadGateway, "please try again later"
	}

	// Register the error handlers and supply an empty version of the type for type reference
	RegisterErrorHandlerOn(registry, &InputValidationError{}, validationHandler)
	RegisterErrorHandlerOn(registry, ErrDatabaseOverloaded, databaseOverloadedHandler)

	// Return errors somewhere deep in your code
	errA := fmt.Errorf("validation error: %w", &InputValidationError{})
	errB := fmt.Errorf("could not connect to database: %w", ErrDatabaseOverloaded)

	// In your HTTP handlers, instantiate responses and return those to the users
	codeA, responseA := NewErrorResponseFrom(context.Background(), registry, errA)
	codeB, responseB := NewErrorResponseFrom(context.Background(), registry, errB)

	// Check the output
	fmt.Printf("%d: %s\n", codeA, responseA)
	fmt.Printf("%d: %s\n", codeB, responseB)

	// Output:
	// 400: Your input was invalid: ...
	// 502: please try again later
}
