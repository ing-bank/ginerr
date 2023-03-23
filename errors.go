package ginerr

import (
	"fmt"
	"net/http"
)

const defaultCode = http.StatusInternalServerError

// DefaultErrorRegistry is the default ErrorRegistry for the application, can be overridden for rare use-cases.
var DefaultErrorRegistry = NewErrorRegistry()

type internalHandler func(err error) (int, any)
type internalStringHandler func(err string) (int, any)

// CustomErrorHandler is the template for custom errors that have unexported errors.
type CustomErrorHandler[R any] func(err error) (int, R)

// ErrorStringHandler is the template for string errors that don't have their own object available. R is the
// type of the response body.
type ErrorStringHandler[R any] func(err string) (int, R)

// ErrorHandler is the template of an error handler in the ErrorRegistry. The E type is the error type that
// the handler is registered for. The R type is the type of the response body.
type ErrorHandler[E error, R any] func(E) (int, R)

// NewErrorRegistry instantiates a new ErrorRegistry, ideal for testing or overriding the default one.
func NewErrorRegistry() *ErrorRegistry {
	return &ErrorRegistry{
		handlers:       map[string]internalHandler{},
		stringHandlers: map[string]internalStringHandler{},
		DefaultCode:    defaultCode,
	}
}

// ErrorRegistry contains a map of errors and ErrorHandler-s.
type ErrorRegistry struct {
	// handlers are used when we know the type of the error
	handlers map[string]internalHandler
	// stringHandlers are used when the error is only a string
	stringHandlers map[string]internalStringHandler

	// DefaultCode is the default code to return when no handler is found
	DefaultCode int

	// defaultBody is the default body to return when no handler is found
	DefaultResponse any
}

// SetDefaultResponse sets the default response on no handlers found
func (e *ErrorRegistry) SetDefaultResponse(code int, response any) {
	e.DefaultCode = code
	e.DefaultResponse = response
}

// NewErrorResponse Returns an error response using the DefaultErrorRegistry. If no specific handler could be found,
// it will return the defaults.
func NewErrorResponse[E error](err E) (int, any) {
	return NewErrorResponseFrom(DefaultErrorRegistry, err)
}

// NewErrorResponseFrom Returns an error response using the given registry. If no specific handler could be found,
// it will return the defaults.
func NewErrorResponseFrom[E error](registry *ErrorRegistry, err E) (int, any) {
	errorType := fmt.Sprintf("%T", err)

	// If a handler is registered for the error type, use it.
	if entry, ok := registry.handlers[errorType]; ok {
		return entry(err)
	}

	// In production, we should return a generic error message. If you want to know why, read this:
	// https://owasp.org/www-community/Improper_Error_Handling
	return registry.DefaultCode, registry.DefaultResponse
}

// RegisterErrorHandler registers an error handler in DefaultErrorRegistry. The R type is the type of the response body.
func RegisterErrorHandler[E error, R any](handler ErrorHandler[E, R]) {
	RegisterErrorHandlerOn[E](DefaultErrorRegistry, handler)
}

// RegisterErrorHandlerOn registers an error handler in the given registry. The R type is the type of the response body.
func RegisterErrorHandlerOn[E error, R any](registry *ErrorRegistry, handler ErrorHandler[E, R]) {
	// Name of the type
	errorType := fmt.Sprintf("%R", *new(E))

	// Wrap it in a closure, we can't save it directly because err E is not available in NewErrorResponseFrom. It will
	// be available in the closure when it is called. Check out TestErrorResponseFrom_ReturnsErrorBInInterface for an example.
	registry.handlers[errorType] = func(err error) (int, any) {
		// We can safely cast it here, because we know it's the right type.
		return handler(err.(E))
	}
}

// RegisterCustomErrorTypeHandler registers an error handler in DefaultErrorRegistry. Same as RegisterErrorHandler,
// but you can set the fmt.Sprint("%R", err) error yourself. Allows you to register error types that aren't exported
// from their respective packages such as the uuid error or *errors.errorString. The R type is the type of the response body.
func RegisterCustomErrorTypeHandler[R any](errorType string, handler CustomErrorHandler[R]) {
	RegisterCustomErrorTypeHandlerOn(DefaultErrorRegistry, errorType, handler)
}

// RegisterCustomErrorTypeHandlerOn registers an error handler in the given registry. Same as RegisterErrorHandlerOn,
// but you can set the fmt.Sprint("%R", err) error yourself. Allows you to register error types that aren't exported
// from their respective packages such as the uuid error or *errors.errorString. The R type is the type of the response body.
func RegisterCustomErrorTypeHandlerOn[R any](registry *ErrorRegistry, errorType string, handler CustomErrorHandler[R]) {
	// Wrap it in a closure, we can't save it directly
	registry.handlers[errorType] = func(err error) (int, any) {
		return handler(err)
	}
}

// RegisterStringErrorHandler allows you to register an error handler for a simple errorString created with
// errors.New() or fmt.Errorf(). Can be used in case you are dealing with libraries that don't have exported
// error objects. Uses the DefaultErrorRegistry. The R type is the type of the response body.
func RegisterStringErrorHandler[R any](errorString string, handler ErrorStringHandler[R]) {
	RegisterStringErrorHandlerOn(DefaultErrorRegistry, errorString, handler)
}

// RegisterStringErrorHandlerOn allows you to register an error handler for a simple errorString created with
// errors.New() or fmt.Errorf(). Can be used in case you are dealing with libraries that don't have exported
// error objects. The R type is the type of the response body.
func RegisterStringErrorHandlerOn[R any](registry *ErrorRegistry, errorString string, handler ErrorStringHandler[R]) {
	registry.stringHandlers[errorString] = func(err string) (int, any) {
		return handler(err)
	}

	// Ensure that the string register is in the normal list of handlers
	if _, ok := registry.handlers["*errors.errorString"]; !ok {

		// Define the default string handler
		registry.handlers["*errors.errorString"] = func(err error) (int, any) {
			// Check if the error string exists
			if handler, ok := registry.stringHandlers[err.Error()]; ok {
				return handler(err.Error())
			}

			return registry.DefaultCode, registry.DefaultResponse
		}
	}
}
