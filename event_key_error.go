package gecore

import "errors"

var (
	ErrIsNotEventKeyError        = errors.New("Not event key error")
	ErrCodeKeyBound              = errors.New("Key bound")
	ErrCodeExtendedFunction      = errors.New("Extended function")
	ErrCodeFunc                  = errors.New("VCommand")
	ErrCodeAction                = errors.New("Action")
	ErrCodeKeyBindingNotFount    = errors.New("Key binding not found")
	ErrCodeUnknownKeyBindingType = errors.New("Unknown key binding type")
)
