package models

import (
	"errors"
)

// use ErrNoRecord instead of sql.ErrNoRows to  encapsulate the model
// completely, so that our application isn’t concerned with the
// underlying datastore or reliant on datastore-specific errors for its
// behavior.
var (
	ErrNoRecord = errors.New("models: no matching record found")

	// ErrInvalidCredentials error will be used if a user tries to login
	// with an incorrect email address or password
	ErrInvalidCredentials = errors.New("models: invalid credentials")

	// ErrDuplicateEmail error will be used if a user tries to
	// signup with an email address that's already in use
	ErrDuplicateEmail = errors.New("models: duplicate email")
)
