package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter" // router
	"github.com/justinas/alice"           // middleware
)

func (app *application) routes() http.Handler {
	// Initialize the router.
	router := httprouter.New()

	// Create a handler function which wraps our notFound() helper
	// Assign it as the custom handler for 404 Not Found responses
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// uprotected application routes using the "dynamic" middleware chain, use nosurf middleware
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// csrf token route
	router.Handler(http.MethodGet, "/api/csrf-token", dynamic.ThenFunc(app.CSRFToken))
	// unprotected user routes
	router.Handler(http.MethodPost, "/api/users/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/api/users/login", dynamic.ThenFunc(app.userLogin))

	// protected application routes, which uses requireAuthentication middleware
	protected := dynamic.Append(app.requireAuthentication)
	log.Println("Setting up protected routes...")

	// protected user routes
	router.Handler(http.MethodGet, "/api/users/view/:userId", protected.ThenFunc(app.viewSpecificUser))
	router.Handler(http.MethodPost, "/api/users/logout", protected.ThenFunc(app.userLogout))
	
	// budget routes
	router.Handler(http.MethodGet, "/api/users/budget/:budgetId/view", protected.ThenFunc(app.budgetView))
	router.Handler(http.MethodGet, "/api/users/budget/:budgetId/summary", protected.ThenFunc(app.budgetSummary))
	router.Handler(http.MethodPost, "/api/users/budget/create", protected.ThenFunc(app.budgetCreate))
	router.Handler(http.MethodPut, "/api/users/budget/update/:budgetId", protected.ThenFunc(app.budgetUpdate))
	router.Handler(http.MethodDelete, "/api/users/budget/delete/:budgetId", protected.ThenFunc(app.budgetDelete))
	
	// expense routes
	router.Handler(http.MethodGet, "/api/users/expenses/view", protected.ThenFunc(app.expensesView))
	router.Handler(http.MethodGet, "/api/users/expenses/view/:expenseId", protected.ThenFunc(app.specificExpenseView))
	router.Handler(http.MethodPost, "/api/users/expenses/create", protected.ThenFunc(app.expenseCreate))
	router.Handler(http.MethodPut, "/api/users/expenses/update/:expenseId", protected.ThenFunc(app.expenseUpdate))
	router.Handler(http.MethodDelete, "/api/users/expenses/delete/:expenseId", protected.ThenFunc(app.expenseDelete))

	// expense category routes
	router.Handler(http.MethodGet, "/api/users/categories/view", protected.ThenFunc(app.categoriesView))
	// should also return total expenses calculation per selected category
	router.Handler(http.MethodGet, "/api/users/categories/expenses/:categoryId", protected.ThenFunc(app.specificCategoryExpensesView))
	router.Handler(http.MethodPost, "/api/users/categories/create", protected.ThenFunc(app.categoryCreate))
	router.Handler(http.MethodDelete, "/api/users/categories/delete/:categoryId", protected.ThenFunc(app.categoryDelete))
	
	// Create a middleware chain containing our 'standard' middleware
	// which will be used for every request our application receives.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	// Return the 'standard' middleware chain followed by the servemux.
	return standard.Then(router)
}
