package application

import "errors"

// sentinel errors ให้ handler map เป็น HTTP status ด้วย errors.Is
var (
	ErrEmailExists        = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrNoFieldsToUpdate   = errors.New("no fields to update")
)
