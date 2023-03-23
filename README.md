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

## ⬇️ Installation

`go get github.com/ing-bank/ginerr`

## 📋 Usage

```go
package main

import (
	"github.com/ing-bank/ginerr"
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
	handler := func(myError *MyError) (int, Response) {
		return http.StatusInternalServerError, Response{
			Errors: map[string]any{
				"error": myError.Error(),
			},
		}
	}

	ginerr.RegisterErrorHandler(handler)
}
```

## 🚀 Development

1. Clone the repository
2. Run `make t` to run unit tests
3. Run `make fmt` to format code

You can run `make` to see a list of useful commands.

## 🔭 Future Plans

Nothing here yet!
