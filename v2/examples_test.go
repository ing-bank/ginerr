package ginerr

import "net/http"

type Response struct {
	Errors map[string]any `json:"errors,omitempty"`
}

type MyError struct {
}

func (m MyError) Error() string {
	return "Something went wrong!"
}

func ExampleRegisterErrorHandler() {
	handler := func(myError *MyError) (int, any) {
		return http.StatusInternalServerError, Response{
			Errors: map[string]any{
				"error": myError.Error(),
			},
		}
	}

	RegisterErrorHandler(&MyError{}, handler)
}

func ExampleRegisterErrorHandlerOn() {
	registry := NewErrorRegistry()

	handler := func(myError *MyError) (int, any) {
		return http.StatusInternalServerError, Response{
			Errors: map[string]any{
				"error": myError.Error(),
			},
		}
	}

	RegisterErrorHandlerOn(registry, &MyError{}, handler)
}

func ExampleRegisterStringErrorHandler() {
	handler := func(myError string) (int, any) {
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

	handler := func(myError string) (int, any) {
		return http.StatusInternalServerError, Response{
			Errors: map[string]any{
				"error": myError,
			},
		}
	}

	RegisterStringErrorHandlerOn(registry, "some string error", handler)
}
