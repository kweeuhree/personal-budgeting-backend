package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid" // router
	"kweeuhree.personal-budgeting-backend/internal/models"
	"kweeuhree.personal-budgeting-backend/internal/validator"
)

// Input struct for creating and updating expenses
type ExpenseInput struct {
	Description   string  `json:"description"`
	AmountInCents int64   `json:"amountInCents"`
	CategoryId	  string  `json:"categoryId"`
	ExpenseType   string  `json:"expenseType"`
	validator.Validator
}

// // Response struct for returning expense data
type ExpenseResponse struct {
	ExpenseId    	string 		`json:"expenseId"`
	CategoryId	    *string  	`json:"categoryId"`
	AmountInCents	int64   	`json:"amountInCents"`
	Description   	string  	`json:"description"`
	ExpenseType   	string  	`json:"expenseType"`
	CreatedAt		time.Time 	`json:"createdAt"`
	Flash 			string		`json:"flash"`
}

// read all user expenses
func (app *application) expensesView(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}
	
	exps, err := app.expenses.All(userId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	response := make([]ExpenseResponse, len(exps))
	for i, exp := range exps {
		response[i] = ExpenseResponse{
			ExpenseId: exp.ExpenseId,
			CategoryId: &exp.CategoryId,
			AmountInCents: exp.AmountInCents,
			Description: exp.Description,
			ExpenseType: exp.ExpenseType,
			CreatedAt: exp.CreatedAt,
		}
	}

	encodeJSON(w, http.StatusOK, response)
}

// read a specific user expense
func (app *application) specificExpenseView(w http.ResponseWriter, r *http.Request) {
	id := app.GetIdFromParams(r, "expenseId")

	if id == "" {
		app.notFound(w)
		return
	}

	

	exp, err := app.expenses.Get(id)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// write the Expense data as a plain-text HTTP response body
	encodeJSON(w, http.StatusOK, exp)
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

	// validate the input against existing balance
	err = app.CurrentBudgetIsValid(userId, input.ExpenseType, input.AmountInCents)
		
	if err != nil {
		app.serverError(w, fmt.Errorf("failed to update current budget: %v", err))
		return
	}

	newId := uuid.New().String()

	id, err := app.expenses.Insert(newId, userId, input.CategoryId, input.Description, input.ExpenseType, input.AmountInCents)
	if err != nil {
		app.serverError(w, fmt.Errorf("unable to add an expense %d; %s", input.AmountInCents, err))
		return
	}
	// Update the budget in the database
	err = app.CalculateAndUpdateBudget(userId, UpdateTypeSubtract, input.ExpenseType, input.AmountInCents, true)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.UpdateCategoryExpenses(userId, input.CategoryId, Increment, input.AmountInCents)
	if err != nil {
		app.serverError(w, fmt.Errorf("unable to increment category expenses %d; %s", input.AmountInCents, err))
		return
	}

    app.setFlash(r.Context(), "Expense has been created.")

	response := ExpenseResponse{
		ExpenseId:  id,
		CategoryId: &input.CategoryId,
		AmountInCents: input.AmountInCents,
		Description: input.Description,
		ExpenseType: input.ExpenseType,
		Flash: app.getFlash(r.Context()),
	}

	err = encodeJSON(w, http.StatusCreated, response)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// update
func (app *application) expenseUpdate(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	expenseId := app.GetIdFromParams(r, "expenseId")
	if expenseId == "" {
		app.notFound(w)
		return
	}
	log.Printf("Current Expense id: %s", expenseId)

	// Decode the JSON body into the input struct
	var input ExpenseInput
	err := decodeJSON(w, r, &input)
	if err != nil {
		log.Printf("Exiting after decoding attempt: %s", err)
		return
	}

	log.Printf("Received input.AmountInCents: %d", input.AmountInCents)

	// validate input
	input.Validate()
	if !input.Valid() {
		encodeJSON(w, http.StatusBadRequest, input.FieldErrors)
		return
	}

	app.setFlash(r.Context(), "Expense has been updated.")

	if input.AmountInCents != 0 {
		_, _, _, _, _, _, err := app.CalculateBudgetUpdates(userId, UpdateTypeSubtract, input.ExpenseType, input.AmountInCents, true)
		if err != nil {
			app.serverError(w, err)
			return
		}

		response := ExpenseResponse{
			ExpenseId:    expenseId,
			AmountInCents: input.AmountInCents,
			Flash: app.getFlash(r.Context()),
		}

		log.Printf("Budget successfully updated")
		err = encodeJSON(w, http.StatusOK, response)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	// Update the new Expense 
	err = app.expenses.Put(expenseId, userId, input.CategoryId, input.Description, input.ExpenseType, input.AmountInCents)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Create a response that includes both ID and body
	response := ExpenseResponse{
		ExpenseId:  expenseId,
		CategoryId: &input.CategoryId,
		AmountInCents: input.AmountInCents,
		Description: input.Description,
		ExpenseType: input.ExpenseType,
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
	log.Printf("deleting an expense")
	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	expenseId := app.GetIdFromParams(r, "expenseId")
	if expenseId == "" {
		app.notFound(w)
		log.Printf("Exiting due to invalid id")
		return
	}
	log.Printf("Current Expense id: %s", expenseId)

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
	} 
	// add the expense amount back to the budget
	err = app.CalculateAndUpdateBudget(userId, UpdateTypeAdd, deletedExpense.ExpenseType, deletedExpense.AmountInCents, true)
	if err != nil {
		app.serverError(w, fmt.Errorf("unable to increment category expenses %d; %s", deletedExpense.AmountInCents, err))
		return
	}
	
	// Update relevant category expenses
	err = app.UpdateCategoryExpenses(userId, deletedExpense.CategoryId, Decrement, deletedExpense.AmountInCents)
	if err != nil {
		app.serverError(w, fmt.Errorf("unable to increment category expenses %d; %s", deletedExpense.AmountInCents, err))
		return
	}
	
	encodeJSON(w, http.StatusOK, "Deleted successfully!")
}
