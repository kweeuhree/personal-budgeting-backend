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
	// Get the value of the "id" named parameter
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("expenseCategoryId")

	// return a 404 Not Found in case of invalid id or error
	if id == "" {
		app.notFound(w)
		return
	}

	ExpenseCategory, err := app.expenseCategory.Get(id)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "ExpenseCategory successfully added!")

	// write the ExpenseCategory data as a plain-text HTTP response body
	fmt.Fprintf(w, "%+v", ExpenseCategory)
}

// read all expenses of a specific category
func (app *application) specificCategoryExpensesView(w http.ResponseWriter, r *http.Request) {
	// Get the value of the "id" named parameter
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("expenseCategoryId")

	// return a 404 Not Found in case of invalid id or error
	if id == "" {
		app.notFound(w)
		return
	}

	ExpenseCategory, err := app.expenseCategory.Get(id)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "ExpenseCategory successfully added!")

	// write the ExpenseCategory data as a plain-text HTTP response body
	fmt.Fprintf(w, "%+v", ExpenseCategory)
}

// create
func (app *application) categoryCreate(w http.ResponseWriter, r *http.Request) {

	userId := app.sessionManager.Get(r.Context(), "userId").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	// Decode the JSON body into the input struct
	var input ExpenseCategoryInput
	err := decodeJSON(w, r, &input)
	if err != nil {
		return
	}

	log.Printf("Received input.Body: %s", input.Name)

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

	app.setFlash(r.Context(), "ExpenseCategory has been created.")

	// Create a response that includes both ID and body
	response := ExpenseCategoryResponse{
		ExpenseCategoryId: newExpenseCategoryId,
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
func (app *application) categoryDelete(w http.ResponseWriter, r *http.Request) {
	log.Printf("Attempting deletion...")

	userId := app.sessionManager.Get(r.Context(), "userId").(string)
	if userId == "" {
		app.serverError(w, fmt.Errorf("userId not found in session"))
		return
	}

	// Get the value of the "id" named parameter
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("expenseCategoryId")
	log.Printf("Current ExpenseCategory id: %s", id)

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
	} else {
		json.NewEncoder(w).Encode("Deleted successfully!")
		return
	}
}
