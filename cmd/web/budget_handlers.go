package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid" // router
	"kweeuhree.personal-budgeting-backend/internal/models"
	"kweeuhree.personal-budgeting-backend/internal/validator"
)

// Input struct for creating budgets
type BudgetInput struct {
	CheckingBalance int64 `json:"checkingBalance"`
	SavingsBalance  int64 `json:"savingsBalance"`
	validator.Validator
}

// Input struct for updating budgets
type BudgetUpdate struct {
	UpdateSumInCents int64  `json:"updateSumInCents"`
	BalanceType      string `json:"balanceType"`
	UpdateType       string `json:"updateType"`
	validator.Validator
}

// // Response struct for returning budget data
type BudgetResponse struct {
	BudgetId        string `json:"budgetId"`
	CheckingBalance int64  `json:"checkingBalance"`
	SavingsBalance  int64  `json:"savingsBalance"`
	BudgetTotal     int64  `json:"budgetTotal"`
	BudgetRemaining int64  `json:"budgetRemaining"`
	TotalSpent      int64  `json:"totalSpent"`
	UpdatedAt       string `json:"updatedAt"`
	Flash           string `json:"flash"`
}

type BudgetUpdateResponse struct {
	Balance        int64  `json:"checkingBalance"`
	SavingsBalance int64  `json:"savingsBalance"`
	Flash          string `json:"flash"`
}

const (
	BalanceTypeChecking = "checkingBalance"
	BalanceTypeSavings  = "savingsBalance"
)

const (
	UpdateTypeAdd      = "add"
	UpdateTypeSubtract = "subtract"
)

// read
func (app *application) budgetView(w http.ResponseWriter, r *http.Request) {
	budgetId := app.GetIdFromParams(r, "budgetId")

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

	response := &BudgetResponse{
		BudgetId:        budgetId,
		CheckingBalance: budget.CheckingBalance,
		SavingsBalance:  budget.SavingsBalance,
		BudgetTotal:     budget.BudgetTotal,
		BudgetRemaining: budget.BudgetRemaining,
		TotalSpent:      budget.TotalSpent,
		UpdatedAt:       budget.UpdatedAt.GoString(),
	}

	encodeJSON(w, http.StatusOK, response)
}

func (app *application) budgetSummary(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting to get budgetSummary...")
}

// create
func (app *application) budgetCreate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting to create budget...")
	// Decode the JSON body into the input struct
	var input BudgetInput
	err := decodeJSON(w, r, &input)
	if err != nil {
		return
	}

	log.Printf("Received input.Checking: %d, Savings: %d", input.CheckingBalance, input.SavingsBalance)

	// validate input
	input.Validate()
	if !input.Valid() {
		encodeJSON(w, http.StatusBadRequest, input.FieldErrors)
		return
	}

	newId := uuid.New().String()
	budgetTotal := input.CheckingBalance + input.SavingsBalance

	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}
	log.Printf("Authenticated user id: %s", userId)
	// Insert the new budget using the ID and body
	id, err := app.budget.Insert(newId, userId, input.CheckingBalance, input.SavingsBalance, budgetTotal)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.setFlash(r.Context(), "Budget has been created.")

	response := BudgetResponse{
		BudgetId:        id,
		CheckingBalance: input.CheckingBalance,
		SavingsBalance:  input.SavingsBalance,
		BudgetTotal:     budgetTotal,
		Flash:           app.getFlash(r.Context()),
	}

	err = encodeJSON(w, http.StatusCreated, response)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// update
func (app *application) budgetUpdate(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	budgetId := app.GetIdFromParams(r, "budgetId")
	if budgetId == "" {
		app.notFound(w)
		log.Printf("Exiting due to invalid id")
		return
	}

	// Decode the JSON body into the input struct
	var input BudgetUpdate
	err := decodeJSON(w, r, &input)
	if err != nil {
		log.Printf("Exiting after decoding attempt...")
		log.Printf("Error message %s", err)
		return
	}

	log.Printf("Received input. UpdateType: %s, BalanceType: %s, Sum: %d", input.UpdateType, input.BalanceType, input.UpdateSumInCents)

	// validate input
	input.Validate()
	if !input.Valid() {
		encodeJSON(w, http.StatusBadRequest, input.FieldErrors)
		return
	}

	// if updateType is Subtract, validate the input, otherwise just update the budget
	if input.UpdateType == UpdateTypeSubtract {
		// validate the input against existing balance
		err = app.CurrentBudgetIsValid(userId, input.BalanceType, input.UpdateSumInCents)

		if err != nil {
			app.serverError(w, fmt.Errorf("failed to update current budget: %v", err))
			return
		}
	}

	updatedBudget, err := app.handleBudgetUpdate(userId, input.BalanceType, input.UpdateType, input.UpdateSumInCents)
	if err != nil {
		app.serverError(w, fmt.Errorf("failed to update current budget: %v", err))
		return
	}

	app.setFlash(r.Context(), "Budget has been updated.")

	response := &BudgetResponse{
		BudgetId:        budgetId,
		CheckingBalance: updatedBudget.CheckingBalance,
		SavingsBalance:  updatedBudget.SavingsBalance,
		BudgetTotal:     updatedBudget.BudgetTotal,
		BudgetRemaining: updatedBudget.BudgetRemaining,
		TotalSpent:      updatedBudget.TotalSpent,
		UpdatedAt:       updatedBudget.UpdatedAt.GoString(),
		Flash:           app.getFlash(r.Context()),
	}

	err = encodeJSON(w, http.StatusCreated, response)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *application) handleBudgetUpdate(
	userId, balanceType, updateType string,
	sumInCents int64,
) (*models.Budget, error) {

	budgetId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent, err := app.CalculateBudgetUpdates(userId, updateType, balanceType, sumInCents, false)
	if err != nil {
		return nil, err
	}

	// Update the budget in the database
	app.UpdateBudgetInDB(budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent)

	// Create the updated budget model
	updatedBudget := &models.Budget{
		BudgetId:        budgetId,
		CheckingBalance: checkingBalance,
		SavingsBalance:  savingsBalance,
		BudgetTotal:     budgetTotal,
		BudgetRemaining: budgetRemaining,
		TotalSpent:      totalSpent,
	}

	return updatedBudget, nil
}

// update the budget in the database
func (app *application) UpdateBudgetInDB(budgetId, userId string, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent int64) {
	err := app.budget.Put(budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent)
	if err != nil {
		log.Printf("Failed to update budget with ID %s for user %s: %s", budgetId, userId, err)
		return
	}

	log.Printf("Budget updated successfully")
}

// delete
func (app *application) budgetDelete(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	budgetId := app.GetIdFromParams(r, "budgetId")
	if budgetId == "" {
		app.notFound(w)
		return
	}

	// Delete the budget using the ID
	err := app.budget.Delete(budgetId, userId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Delete all expenses associated with that user
	err = app.expenses.DeleteAll(userId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Void all expense categories totalSums
	err = app.expenseCategory.VoidAllTotalSums(userId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	encodeJSON(w, http.StatusOK, "Deleted successfully!")
}
