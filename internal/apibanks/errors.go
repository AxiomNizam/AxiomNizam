package apibanks

import "errors"

var (
	ErrBankNotFound      = errors.New("bank not found")
	ErrBankAlreadyExists = errors.New("bank already exists")
	ErrAPINotFound       = errors.New("api not found in bank")
	ErrAPIAlreadyInBank  = errors.New("api already in bank")
	ErrNameRequired      = errors.New("bank name is required")
	ErrTypeMismatch      = errors.New("apibanks: reconciler received non-APIBankResource")
)
