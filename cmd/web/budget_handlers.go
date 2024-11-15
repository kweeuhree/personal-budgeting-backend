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

// Input struct for creating and updating budgets
type BudgetInput struct {
	CheckingBalance 	int64 `json:"checkingBalance"`
	SavingsBalance 		int64 `json:"savingsBalance"`
	validator.Validator
}

// // Response struct for returning budget data
type BudgetResponse struct {
	BudgetId    		string 	`json:"budgetId"`
	CheckingBalance  	int64 	`json:"checkingBalance"`
	SavingsBalance 		int64 	`json:"savingsBalance"`
	BudgetTotal 		int64 	`json:"budgetTotal"`
	BudgetRemaining 	int64 	`json:"budgetRemaining"`
	TotalSpent 			int64 	`json:"totalSpent"`
	UpdatedAt			string	`json:"updatedAt"`
	Flash 				string 	`json:"flash"`
}

// read
func (app *application) budgetView(w http.ResponseWriter, r *http.Request) {
	// Get the value of the "id" named parameter
	params := httprouter.ParamsFromContext(r.Context())
	budgetId := params.ByName("budgetId")

	// return a 404 Not Found in case of invalid id or error
	if budgetId == "" {
		app.notFound(w)
		return
	}

	budget, err := app.budget.Get(budgetId)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// write the budget data as a plain-text HTTP response body
	fmt.Fprintf(w, "%+v", budget)
}

func (app *application) budgetSummary(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting to get budgetSummary...")
}

// create
func (app *application) budgetCreate(w http.ResponseWriter, r *http.Request) {

	userId := app.sessionManager.Get(r.Context(), "userId").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	// Decode the JSON body into the input struct
	var input BudgetInput
	err := decodeJSON(w, r, &input)
	if err != nil {
		return
	}

	log.Printf("Received input.Checking: %s, Savings: %s", input.CheckingBalance, input.SavingsBalance)

	// validate input
	input.Validate()
	if !input.Valid() {
		encodeJSON(w, http.StatusBadRequest, input.FieldErrors)
		return
	}

	newId := uuid.New().String()
	budgetTotal := input.CheckingBalance + input.SavingsBalance

	// Insert the new budget using the ID and body
	id, err := app.budget.Insert(newId, userId, input.CheckingBalance, input.SavingsBalance, budgetTotal)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.setFlash(r.Context(), "Budget has been created.")

	// Create a response that includes both ID and body
	response := BudgetResponse{
		BudgetId:    id,
		CheckingBalance:  input.CheckingBalance,
		SavingsBalance: input.SavingsBalance,
		BudgetTotal: budgetTotal,
		Flash: app.getFlash(r.Context()),
	}

	// Write the response struct to the response as JSON
	err = encodeJSON(w, http.StatusOK, response)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *application) handleBudgetUpdate(userId, expenseType string, expenseAmount int64) (*models.Budget, error) {
	// Fetch the current budget for the user
	currentBudget, err := app.budget.GetBudgetByUserId(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current budget: %v", err)
	}

	// Calculate updated budget values
	checkingBalance, savingsBalance, budgetTotal, budgetRemaining, err := app.calculateBudgetUpdates(currentBudget, expenseType, expenseAmount)

	if err != nil {
		return nil, err // Return the error if the expenseType was invalid
	}

	// Update the budget
	app.budgetUpdate(
		currentBudget.BudgetId, 
		userId, 
		checkingBalance, 
		savingsBalance, 
		budgetTotal, 
		budgetRemaining, 
		currentBudget.TotalSpent+expenseAmount,
	)

	updatedBudget := &models.Budget{
		BudgetId:        currentBudget.BudgetId,
		CheckingBalance: checkingBalance,
		SavingsBalance:  savingsBalance,
		BudgetTotal:     budgetTotal,
		BudgetRemaining: budgetRemaining,
		TotalSpent:      currentBudget.TotalSpent + expenseAmount,
	}

	// Return the updated budget
	return updatedBudget, nil
}

func (app *application) calculateBudgetUpdates(currentBudget *models.Budget, expenseType string, expenseAmount int64) (int64, int64, int64, int64, error) {
		// Adjust checking or savings balance based on expense type
	switch expenseType {
	case "checking":
		currentBudget.CheckingBalance -= expenseAmount
	case "savings":
		currentBudget.SavingsBalance -= expenseAmount
	default:
		// Handle invalid expense type
		err := fmt.Errorf("invalid expense type: %s", expenseType)
		return 0, 0, 0, 0, err
	}

	// Update total spent
	newTotalSpent := currentBudget.TotalSpent + expenseAmount

	// Calculate remaining budget
	newBudgetRemaining := currentBudget.BudgetTotal - newTotalSpent

	// Return updated fields
	return currentBudget.CheckingBalance, currentBudget.SavingsBalance, currentBudget.BudgetTotal, newBudgetRemaining, nil
}

// update
func (app *application) budgetUpdate(budgetId, userId string, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent int64) {

	err := app.budget.Put(budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent)
	if err != nil {
		log.Printf("Failed to update budget with ID %s for user %s: %s", budgetId, userId, err)
		return
	}

	log.Printf("Budget updated successfully")
	return
}

// delete
func (app *application) budgetDelete(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting deletion...")

	userId := app.sessionManager.Get(r.Context(), "userId").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	// Get the value of the "id" named parameter
	params := httprouter.ParamsFromContext(r.Context())
	budgetId := params.ByName("budgetId")

	if budgetId == "" {
		log.Printf("Exiting due to invalid id")
		app.notFound(w)
		return
	}
	

	// Delete the budget using the ID
	err := app.budget.Delete(budgetId, userId)
	if err != nil {
		app.serverError(w, err)
		return
	} else {
		json.NewEncoder(w).Encode("Deleted successfully!")
		return
	}
}
