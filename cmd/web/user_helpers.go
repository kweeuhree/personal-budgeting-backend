package main

import "github.com/google/uuid"

func (app *application) CreateAndStoreUser(email, displayName, password string) (string, error) {
	newId := uuid.New().String()

	// Try to create a new user record in the database. If the email already
	// exists then add an error message to the form and re-display it.
	err := app.user.Insert(newId, email, displayName, password)
	if err != nil {
		return "", err
	}

	return newId, nil
}
