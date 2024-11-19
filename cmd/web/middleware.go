package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	// environment variables
	"github.com/joho/godotenv"
	// double submit cookies
	"github.com/justinas/nosurf"
)

func secureHeaders(next http.Handler) http.Handler {

	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// reactAddress := os.Getenv("REACT_ADDRESS")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Specify origin
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Allow specific HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Allow specific headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("requireAuthentication middleware triggered for", r.URL.Path)
		// if the user is not authenticated, redirect them to the login page and
		// return from the middleware chain so that no subsequent handlers in
		// the chain are executed.
		if !app.isAuthenticated(r) {
			log.Println("Authenticated request blocked.")
			response := map[string]string{
				"status":  "401 Unauthorized",
				"message": "You must be logged in to access this resource",
			}
			encodeJSON(w, http.StatusUnauthorized, response)
			return
		}
		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache
		// (or other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")
		log.Println("Authenticated request proceeding.")
		// And call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("authenticate middleware triggered for", r.URL.Path)
		// Retrieve the authenticatedUserId value from the session
		id := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

		// When we don’t have a valid authenticated user, we pass the
		// original and unchanged *http.Request to the next handler in the chain.
		if id == "" {
			next.ServeHTTP(w, r)
			return
		}
		// Otherwise, we check to see if a user with that ID exists in our
		// database.
		exists, err := app.user.Exists(id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		// If a matching user is found, we know that the request is
		// coming from an authenticated user who exists in our database. We
		// create a new copy of the request (with an isAuthenticatedContextKey

		// value of true in the request context) and assign it to r.
		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}
		fmt.Println("User ID from session:", id)
		fmt.Println("User exists:", exists)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})

}

// Create a NoSurf middleware function which uses a customized CSRF cookie
// with the Secure, Path and HttpOnly attributes set.
func noSurf(next http.Handler) http.Handler {
	fmt.Println("inside nosurf handler")
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}

// Returns the CSRF token as a JSON response
func (app *application) CSRFToken(w http.ResponseWriter, r *http.Request) {
	token := nosurf.Token(r)
	if token == "" {
		app.errorLog.Println("CSRF token is empty")
	} else {
		app.infoLog.Println("CSRF token generated:", token)
	}

	err := encodeJSON(w, http.StatusOK, map[string]string{"csrf_token": token})
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method,
			r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

// There are two details about this which are worth explaining:
// Setting the Connection: Close header on the response acts as a
// trigger to make Go’s HTTP server automatically close the current
// connection after a response has been sent. It also informs the
// user that the connection will be closed. Note: If the protocol being
// used is HTTP/2, Go will automatically strip the Connection: Close
// header from the response (so it is not malformed) and send a
// GOAWAY frame.
// The value returned by the builtin recover() function has the type
// any, and its underlying type could be string, error, or something
// else — whatever the parameter passed to panic() was. In our
// case, it’s the string "oops! something went wrong". In the code
// above, we normalize this into an error by using the fmt.Errorf()
// function to create a new error object containing the default
// textual representation of the any value, and then pass this error
// to the app.serverError() helper method.

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal Server response.
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
