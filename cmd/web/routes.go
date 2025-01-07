package main

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter" // router
	"github.com/justinas/alice"           // middleware
)

func (app *application) routes() http.Handler {
	log.Println("Routing...")
	// Initialize the router.
	router := httprouter.New()

	// Serve static files
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
	router.Handler(http.MethodGet, "/api/budget/:budgetId/view", protected.ThenFunc(app.budgetView))
	router.Handler(http.MethodGet, "/api/budget/:budgetId/summary", protected.ThenFunc(app.budgetSummary))
	router.Handler(http.MethodPost, "/api/budget/create", protected.ThenFunc(app.budgetCreate))
	router.Handler(http.MethodPut, "/api/budget/update/:budgetId", protected.ThenFunc(app.budgetUpdate))
	router.Handler(http.MethodDelete, "/api/budget/delete/:budgetId", protected.ThenFunc(app.budgetDelete))

	// expense routes
	router.Handler(http.MethodGet, "/api/expenses/view", protected.ThenFunc(app.expensesView))
	router.Handler(http.MethodGet, "/api/expenses/view/:expenseId", protected.ThenFunc(app.specificExpenseView))
	router.Handler(http.MethodPost, "/api/expenses/create", protected.ThenFunc(app.expenseCreate))
	router.Handler(http.MethodPut, "/api/expenses/update/:expenseId", protected.ThenFunc(app.expenseUpdate))
	router.Handler(http.MethodDelete, "/api/expenses/delete/:expenseId", protected.ThenFunc(app.expenseDelete))

	// expense category routes
	router.Handler(http.MethodGet, "/api/categories/view", protected.ThenFunc(app.categoriesView))
	router.Handler(http.MethodGet, "/api/categories/expenses/:categoryId", protected.ThenFunc(app.specificCategoryExpensesView))
	router.Handler(http.MethodPost, "/api/categories/create", protected.ThenFunc(app.categoryCreate))
	router.Handler(http.MethodDelete, "/api/categories/delete/:categoryId", protected.ThenFunc(app.categoryDelete))

	// Catch-all route to serve index.html for all other routes
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("./ui/static/index.html"); os.IsNotExist(err) {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, "./ui/static/index.html")
	})

	router.GET("/check-index", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if _, err := os.Stat("./ui/static/index.html"); os.IsNotExist(err) {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			log.Println("Error: index.html not found in ./ui/static/")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("index.html is available"))
		log.Println("index.html is registered and available.")
	})

	// Create a middleware chain containing our 'standard' middleware
	// which will be used for every request our application receives.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	// Return the 'standard' middleware chain followed by the servemux.
	return standard.Then(router)
}
