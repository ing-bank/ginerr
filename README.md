# 🦁 Gin Error Registry

[![Go package](https://github.com/ing-bank/ginerr/actions/workflows/test.yaml/badge.svg)](https://github.com/ing-bank/ginerr/actions/workflows/test.yaml)
![GitHub](https://img.shields.io/github/license/ing-bank/ginerr)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ing-bank/ginerr)

Sending any error back to the user can pose a [big security risk](https://owasp.org/www-community/Improper_Error_Handling).
For this reason we developed an error registry that allows you to register specific error handlers
for your application. This way you can control what information is sent back to the user.

## 👷 V3 migration guide

V3 completely revamps the `ErrorRegistry` and now utilises the `errors` package to match errors.
The following changes have been made:

- `RegisterErrorHandler` now requires a concrete instance of the error as its first argument
- `RegisterErrorHandlerOn` now requires a concrete instance of the error as its second argument
- `RegisterStringErrorHandler` has been removed, use static `errors.New` in `RegisterErrorHandler` to get this to work
- `RegisterStringErrorHandlerOn` has been removed, use static `errors.New` in `RegisterErrorHandlerOn` to get this to work
- `RegisterCustomErrorTypeHandler` has been removed, wrap unexported errors from libraries to create handlers for these
- `RegisterCustomErrorTypeHandlerOn` has been removed, wrap unexported errors from libraries to create handlers for these
- `ErrorRegistry` changes:
  - `DefaultCode` has been removed, use `RegisterDefaultHandler` instead
  - `DefaultResponse` has been removed, use `RegisterDefaultHandler` instead
  - `SetDefaultResponse` has been removed, use `RegisterDefaultHandler` instead

## ⬇️ Installation

`go get github.com/ing-bank/ginerr/v3`

## 📋 Usage

Check out [the examples here](./examples_test.go).

## 🚀 Development

1. Clone the repository
2. Run `make tools` to install necessary tools
3. Run `make fmt` to format code
4. Run `make lint` to lint your code

You can run `make` to see a list of useful commands.

## 🔭 Future Plans

Nothing here yet!
