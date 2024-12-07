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

// Input struct for creating and updating ExpenseCategorys
type ExpenseCategoryInput struct {
	Name 				string `json:"name"`
	Description 		string `json:"description"`
	validator.Validator
}

// // Response struct for returning ExpenseCategory data
type ExpenseCategoryResponse struct {
	ExpenseCategoryId    string `json:"expenseCategoryId"`
	Name  				 string `json:"name"`
	Description 		 string `json:"description"`
	Flash 				 string `json:"flash"`
}

// read all user categories
func (app *application) categoriesView(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	cats, err := app.expenseCategory.All(userId)
	if err != nil {
		app.serverError(w, err)
		return
	}

    response := make([]ExpenseCategoryResponse, len(cats))
    for i, cat := range cats {
        response[i] = ExpenseCategoryResponse{
            ExpenseCategoryId: cat.ExpenseCategoryId,
            Name:              cat.Name,
            Description:       cat.Description,
        }
    }

	encodeJSON(w, http.StatusOK, response)
}

// read all expenses of a specific category
func (app *application) specificCategoryExpensesView(w http.ResponseWriter, r *http.Request) {
	id := app.GetIdFromParams(r, "categoryId")
	if id == "" {
		app.notFound(w)
		return
	}

	exps, err := app.expenseCategory.AllExpensesPerCategory(id)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	var total int64
	if len(exps) > 1 {
		total = app.GetExpensesTotal(exps)
	}
	
    response := map[string]interface{}{
		"totalSpent": total,
        "expenses": exps,
    }

    encodeJSON(w, http.StatusOK, response)
}

// create
func (app *application) categoryCreate(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.Get(r.Context(), "authenticatedUserID").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	var input ExpenseCategoryInput
	err := decodeJSON(w, r, &input)
	if err != nil {
		return
	}

	// validate input
	input.Validate()
	if !input.Valid() {
		encodeJSON(w, http.StatusBadRequest, input.FieldErrors)
		return
	}

	newId := uuid.New().String()

	// Insert the new ExpenseCategory using the ID and body
	newExpenseCategoryId, err := app.expenseCategory.Insert(newId, userId, input.Name, input.Description)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.setFlash(r.Context(), "Expense category has been created.")

	// Create a response that includes both ID and body
	response := ExpenseCategoryResponse{
		ExpenseCategoryId: newExpenseCategoryId,
		Name: input.Name,
		Description: input.Description,
		Flash: app.getFlash(r.Context()),
	}

	// Write the response struct to the response as JSON
	err = encodeJSON(w, http.StatusCreated, response)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// delete
func (app *application) categoryDelete(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting deletion...")

	userId := app.sessionManager.Get(r.Context(), "userId").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}
	id := app.GetIdFromParams(r, "expenseCategoryId")
	if id == "" {
		app.notFound(w)
		log.Printf("Exiting due to invalid id")
		return
	}

	// Delete the ExpenseCategory using the ID
	err := app.expenseCategory.Delete(id, userId)
	if err != nil {
		app.serverError(w, err)
		return
	} 
	
	encodeJSON(w, http.StatusOK, "Deleted successfully!")
}
