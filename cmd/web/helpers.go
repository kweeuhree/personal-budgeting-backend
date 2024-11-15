package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"

	"kweeuhree.personal-budgeting-backend/internal/validator"
)

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
// -- use the debug.Stack() function to get a stack trace for the current goroutine and append it to the
// -- log message. Being able to see the execution path of the
// -- application via the stack trace can be helpful when you’re trying to debug errors.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// report the file name and line number one step back in the stack trace
	// to have a clearer idea of where the error actually originated from
	// set frame depth to 2
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400
// "Bad Request" when there's a problem with the request that the user sent.
// -- use the http.StatusText() function to automatically generate a human-friendly text
// representation of a given HTTP status code. For example,
// http.StatusText(400) will return the string "Bad Request".
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found
// response to the user.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return err
	}
	return nil
}

func encodeJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// Helper method to set a flash message in the session
func (app *application) setFlash(ctx context.Context, message string) {
	app.sessionManager.Put(ctx, "flash", message)
}

// Helper method to get and clear the flash message from the session
func (app *application) getFlash(ctx context.Context) string {
	return app.sessionManager.PopString(ctx, "flash")
}

func (input *ExpenseInput) Validate() {
	amountInCentsStr := strconv.FormatInt(input.AmountInCents, 10)
	input.CheckField(validator.NotBlank(amountInCentsStr), "amountInCents", "This field cannot be blank")
}

func (input *ExpenseCategoryInput) Validate() {
	input.CheckField(validator.NotBlank(input.Name), "amountInCents", "This field cannot be blank")
}

func (form *UserSignUpInput) Validate() {
	form.CheckField(validator.NotBlank(form.DisplayName), "displayName", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
}

// checks that email and password are provided
// and also check the format of the email address as
// a UX-nicety (in case the user makes a typo).
func (form *UserLoginInput) Validate() {
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
}

func (input *BudgetInput) Validate() {
	checkingBalanceStr := strconv.FormatInt(input.CheckingBalance, 10)
	input.CheckField(validator.NotBlank(checkingBalanceStr), "checkingBalance", "This field cannot be blank")
}

// Return true if the current request is from an authenticated user, otherwise return false
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	fmt.Println("Is user authenticated:", isAuthenticated)

	if !ok {
		return false
	}
	return isAuthenticated
}



