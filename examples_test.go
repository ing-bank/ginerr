package ginerr

import (
	"context"
	"net/http"
)

type Response struct {
	Errors map[string]any `json:"errors,omitempty"`
}

type MyError struct{}

func (m MyError) Error() string {
	return "Something went wrong!"
}

func ExampleRegisterErrorHandler() {
	handler := func(ctx context.Context, myError *MyError) (int, any) {
		return http.StatusInternalServerError, Response{
			Errors: map[string]any{
				"error": myError.Error(),
			},
		}
	}

	RegisterErrorHandler(handler)
}

func ExampleRegisterErrorHandlerOn() {
	registry := NewErrorRegistry()

	handler := func(ctx context.Context, myError *MyError) (int, any) {
		return http.StatusInternalServerError, Response{
			Errors: map[string]any{
				"error": myError.Error(),
			},
		}
	}

	RegisterErrorHandlerOn(registry, handler)
}

func ExampleRegisterStringErrorHandler() {
	handler := func(ctx context.Context, myError string) (int, any) {
		return http.StatusInternalServerError, Response{
			Errors: map[string]any{
				"error": myError,
			},
		}
	}

	RegisterStringErrorHandler("some string error", handler)
}

func ExampleRegisterStringErrorHandlerOn() {
	registry := NewErrorRegistry()

	handler := func(ctx context.Context, myError string) (int, any) {
		return http.StatusInternalServerError, Response{
			Errors: map[string]any{
				"error": myError,
			},
		}
	}

	RegisterStringErrorHandlerOn(registry, "some string error", handler)
}
