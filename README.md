# 🦁 Gin Error Registry

[![Go package](https://github.com/ing-bank/ginerr/actions/workflows/test.yaml/badge.svg)](https://github.com/ing-bank/ginerr/actions/workflows/test.yaml)
![GitHub](https://img.shields.io/github/license/ing-bank/ginerr)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ing-bank/ginerr)

Sending any error back to the user can pose a [big security risk](https://owasp.org/www-community/Improper_Error_Handling).
For this reason we developed an error registry that allows you to register specific error handlers
for your application. This way you can control what information is sent back to the user.

You can register errors in 3 ways:
- By error type
- By value of string errors
- By defining the error name yourself

## 👷 V2 migration guide

V2 of this library changes the interface of all the methods to allow contexts to be passed to handlers. This
allows you to add additional data to the final response.

The interface changes are as follows.

- `RegisterErrorHandler` and all its variants take a context as a first parameter in the handler, allowing you to pass more data to the response
- `RegisterErrorHandler` and all its variants require the callback function to return `(int, any)` instead of `(int, R)`, removing the unnecessary generic
- Both `NewErrorResponse` and `NewErrorResponseFrom` take a context as a first parameter, this could be the request context but that's up to you

## ⬇️ Installation

`go get github.com/ing-bank/ginerr/v2`

## 📋 Usage

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ing-bank/ginerr/v2"
	"net/http"
)

type MyError struct {
}

func (m *MyError) Error() string {
	return "Something went wrong!"
}

// Response is an example response object, you can return anything you like
type Response struct {
	Errors map[string]any `json:"errors,omitempty"`
}

func main() {
	handler := func(ctx context.Context, myError *MyError) (int, any) {
		return http.StatusInternalServerError, Response{
			Errors: map[string]any{
				"error": myError.Error(),
			},
		}
	}

	ginerr.RegisterErrorHandler(handler)
	
	// [...]
}

func handleGet(c *gin.Context) {
	err := &MyError{}
	c.JSON(ginerr.NewErrorResponse(c.Request.Context(), err))
}
```

## 🚀 Development

1. Clone the repository
2. Run `make tools` to install necessary tools
3. Run `make t` to run unit tests
4. Run `make fmt` to format code
4. Run `make lint` to lint your code

You can run `make` to see a list of useful commands.

## 🔭 Future Plans

Nothing here yet!
