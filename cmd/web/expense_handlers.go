package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter" // router
	"kweeuhree.personal-budgeting-backend/internal/models"
	"kweeuhree.personal-budgeting-backend/internal/validator"
)

// Input struct for creating and updating expenses
type ExpenseInput struct {
	AmountInCents int64   `json:"amountInCents"`
	CategoryId	  string  `json:"categoryId"`
	Description   string  `json:"description"`
	ExpenseType   string  `json:"expenseType"`
	validator.Validator
}

// // Response struct for returning expense data
type ExpenseResponse struct {
	ExpenseId    	string 			`json:"expenseId"`
	AmountInCents	int64   		`json:"amountInCents"`
	Flash 			string			`json:"flash"`
}

// read all user expenses
func (app *application) expensesView(w http.ResponseWriter, r *http.Request) {
	expenses, err := app.expenses.All()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Set the Content-Type header to application/json if you are sending JSON
	w.Header().Set("Content-Type", "application/json")

	// Write the todos to the response as JSON
	err = json.NewEncoder(w).Encode(expenses)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// read a specific user expense
func (app *application) specificExpenseView(w http.ResponseWriter, r *http.Request) {
	// Get the value of the "id" named parameter
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("expenseId")

	// return a 404 Not Found in case of invalid id or error
	if id == "" {
		app.notFound(w)
		return
	}

	Expense, err := app.expenses.Get(id)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Expense successfully added!")

	// write the Expense data as a plain-text HTTP response body
	fmt.Fprintf(w, "%+v", Expense)
}

// create
func (app *application) expenseCreate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting to create an expense")
	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	// Decode the JSON body into the input struct
	var input ExpenseInput
	err := decodeJSON(w, r, &input)
	if err != nil {
		return
	}

	log.Printf("Received input.amountInCents: %d", input.AmountInCents)
	log.Printf("Received input.categoryId: %s", input.CategoryId)
	// validate input
	input.Validate()
	if !input.Valid() {
		encodeJSON(w, http.StatusBadRequest, input.FieldErrors)
		return
	}

	newId := uuid.New().String()

	
	// Insert the new Expense using the ID and body
	id, err := app.expenses.Insert(newId, userId, input.CategoryId, input.Description, input.ExpenseType, input.AmountInCents)
	if err != nil {
		app.serverError(w, fmt.Errorf("unable to add an expense %d; %s", input.AmountInCents, err))
		return
	}

	app.setFlash(r.Context(), "Expense has been created.")

	_, err = app.handleExpenseUpdate(w, userId, input.ExpenseType, UpdateTypeSubtract, input.AmountInCents)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Create a response that includes both ID and body
	response := ExpenseResponse{
		ExpenseId:  id,
		AmountInCents: input.AmountInCents,
		Flash: app.getFlash(r.Context()),
	}

	// Write the response struct to the response as JSON
	err = encodeJSON(w, http.StatusOK, response)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// update
func (app *application) expenseUpdate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting update...")

	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	// Get the value of the "expenseId" named parameter
	params := httprouter.ParamsFromContext(r.Context())
	expenseId := params.ByName("expenseId")
	log.Printf("Current Expense id: %s", expenseId)

	if expenseId == "" {
		app.notFound(w)
		log.Printf("Exiting due to invalid id")
		return
	}

	// Decode the JSON body into the input struct
	var input ExpenseInput
	err := decodeJSON(w, r, &input)
	if err != nil {
		log.Printf("Exiting after decoding attempt...")
		log.Printf("Error message %s", err)
		return
	}

	log.Printf("Received input.AmountInCents: %d", input.AmountInCents)

	// validate input
	input.Validate()
	if !input.Valid() {
		encodeJSON(w, http.StatusBadRequest, input.FieldErrors)
		return
	}

	// Update the new Expense 
	err = app.expenses.Put(expenseId, userId, input.CategoryId, input.Description, input.ExpenseType, input.AmountInCents)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.setFlash(r.Context(), "Expense has been updated.")

	if input.AmountInCents != 0 {
		updatedBudget, err := app.handleExpenseUpdate(w, userId, input.ExpenseType, UpdateTypeSubtract, input.AmountInCents)
		if err != nil {
			app.serverError(w, err)
			return
		}

		response := ExpenseResponse{
			ExpenseId:    expenseId,
			AmountInCents: input.AmountInCents,
			Flash: app.getFlash(r.Context()),
		}

		log.Printf("Budget successfully updated: %+v", updatedBudget)
		err = encodeJSON(w, http.StatusOK, response)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	// Create a response that includes both ID and body
	response := ExpenseResponse{
		ExpenseId:    expenseId,
		Flash: app.getFlash(r.Context()),
	}

	// Write the response struct to the response as JSON
	err = encodeJSON(w, http.StatusOK, response)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// delete
func (app *application) expenseDelete(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting deletion...")

	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	// Get the value of the "id" named parameter
	params := httprouter.ParamsFromContext(r.Context())
	expenseId := params.ByName("expenseId")
	log.Printf("Current Expense id: %s", expenseId)

	if expenseId == "" {
		app.notFound(w)
		log.Printf("Exiting due to invalid id")
		return
	}

	deletedExpense, err := app.expenses.Get(expenseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	// Delete the Expense using the ID
	err = app.expenses.Delete(expenseId, userId)
	if err != nil {
		app.serverError(w, err)
		return
	} else {
		json.NewEncoder(w).Encode("Deleted successfully!")
		// add the expense amount back to the budget
		app.handleExpenseUpdate(w, userId, deletedExpense.ExpenseType, UpdateTypeAdd, deletedExpense.AmountInCents)
		return
	}
}
