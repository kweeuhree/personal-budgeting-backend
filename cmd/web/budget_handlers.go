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

// Input struct for creating budgets
type BudgetInput struct {
	CheckingBalance 	int64 `json:"checkingBalance"`
	SavingsBalance 		int64 `json:"savingsBalance"`
	validator.Validator
}

// Input struct for updating budgets
type BudgetUpdate struct {
	UpdateSumInCents 	int64  `json:"updateSumInCents"`
	BalanceType 		string `json:"balanceType"`
	UpdateType			string `json:"updateType"`
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

type BudgetUpdateResponse struct {
	Balance 	int64 `json:"checkingBalance"`
	SavingsBalance 		int64 `json:"savingsBalance"`
	Flash				string 	`json:"flash"`
}

const (
    BalanceTypeChecking = "CheckingBalance"
    BalanceTypeSavings  = "SavingsBalance"
)

const (
    UpdateTypeAdd = "add"
    UpdateTypeSubtract  = "subtract"
)

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

// update
func (app *application) budgetUpdate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting update...")

	// Get the value of the "id" named parameter
	params := httprouter.ParamsFromContext(r.Context())
	budgetId := params.ByName("budgetId")
	log.Printf("Current budget id: %s", budgetId)

	if budgetId == "" {
		app.notFound(w)
		log.Printf("Exiting due to invalid id")
		return
	}

	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
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
	app.handleBalanceUpdate(w, userId, input.BalanceType, input.UpdateType, input.UpdateSumInCents)
}

func (app *application) handleBudgetUpdate(
    w http.ResponseWriter, 
    userId, balanceType, updateType string, 
    sumInCents int64, 
    calculateFunc func(*models.Budget, string, string, int64) (int64, int64, int64, int64, int64, error),
) (*models.Budget, error) {
    // Fetch the current budget for the user
    currentBudget, err := app.budget.GetBudgetByUserId(userId)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch current budget: %v", err)
    }

    // Validate if update type is subtract
    if updateType == UpdateTypeSubtract {
        err = app.CurrentBudgetIsValid(currentBudget, balanceType, sumInCents)
        if err != nil {
			http.Error(w, "Insufficient funds to add this expense", http.StatusUnprocessableEntity)
            return nil, fmt.Errorf("unable to update budget: %v", err)
        }
    }

    // Calculate updated budget values using the provided function
    checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent, err := calculateFunc(currentBudget, updateType, balanceType, sumInCents)
    if err != nil {
        return nil, err
    }

    // Update the budget in the database
    app.updateBudgetInDB(currentBudget.BudgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent)

    // Create the updated budget model
    updatedBudget := &models.Budget{
        BudgetId:        currentBudget.BudgetId,
        CheckingBalance: checkingBalance,
        SavingsBalance:  savingsBalance,
        BudgetTotal:     budgetTotal,
        BudgetRemaining: budgetRemaining,
        TotalSpent:      totalSpent,
    }

    // Encode the response
    encodeJSON(w, http.StatusOK, updatedBudget)

    return updatedBudget, nil
}

// update the budget in the database
func (app *application) updateBudgetInDB(budgetId, userId string, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent int64) {
	err := app.budget.Put(budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent)
	if err != nil {
		log.Printf("Failed to update budget with ID %s for user %s: %s", budgetId, userId, err)
		return
	}

	log.Printf("Budget updated successfully")
}

// delete
func (app *application) budgetDelete(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting deletion...")

	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
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

func (app *application) handleBalanceUpdate(w http.ResponseWriter, userId, balanceType, updateType string, sumInCents int64) (*models.Budget, error) {
    return app.handleBudgetUpdate(w, userId, balanceType, updateType, sumInCents, func(currentBudget *models.Budget, updateType, balanceType string, sumInCents int64) (int64, int64, int64, int64, int64, error) {
        return app.CalculateUpdates(currentBudget, updateType, balanceType, sumInCents, false)
    })
}

