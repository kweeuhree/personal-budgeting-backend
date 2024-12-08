package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

// define a Expense type
type Expense struct {
	ExpenseId      string
	UserId    	   string
	CategoryId     string
	Description    string
	ExpenseType	   string
	AmountInCents  int64
	CreatedAt      time.Time
}

// define a Expense model type which wraps a sql.DB connection pool
type ExpenseModel struct {
	DB *sql.DB
}

// insert a new Expense into the database
func (m *ExpenseModel) Insert(expenseId, userId, categoryId, description, expenseType string, amountInCents int64) (string, error) {
	stmt := `INSERT INTO Expenses (expenseId, userId, categoryId, description, expenseType, amountInCents, createdAt) 	
			VALUES(?, ?, ?, ?, ?, ?, UTC_TIMESTAMP())`

	_, err := m.DB.Exec(stmt, expenseId, userId, categoryId, description, expenseType, amountInCents)
	if err != nil {
		return "", err
	}

	return expenseId, nil
}

// return a specific expense based on its id
func (m *ExpenseModel) Get(expenseId string) (*Expense, error) {
	// Write the SQL statement we want to execute
	stmt := `SELECT expenseId, userId, categoryId, description, expenseType, amountInCents, createdAt 
			FROM Expenses WHERE expenseId = ?`

	// This returns a pointer to a sql.Row object
	// which holds the result from the database
	row := m.DB.QueryRow(stmt, expenseId)

	// Initialize a pointer to a new zeroed Expense struct
	exp := &Expense{}

	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the expense struct. 
	// The arguments to row.Scan are *pointers* to the place to copy the data into
	err := row.Scan(&exp.ExpenseId, &exp.UserId, &exp.CategoryId, &exp.Description, &exp.ExpenseType, &exp.AmountInCents, &exp.CreatedAt)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a
		// sql.ErrNoRows error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	// If everything went OK then return the Expense object
	return exp, nil
}

// return the all created Expenses
func (m *ExpenseModel) All(userId string) ([]*Expense, error) {
	stmt := `SELECT * FROM Expenses
			WHERE userId = ?
			ORDER BY createdAt DESC`

	// This returns a sql.Rows resultset containing the result of query
	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}

	// Defer rows.Close() to ensure the sql.Rows resultset is always
	// properly closed before the Latest() method returns
	// Defer stmt should come *after* you check for an error from the Query()
	// method. Otherwise, if Query() returns an error, you'll get a panic trying to close a nil resultset
	defer rows.Close()

	// Initialize an empty slice to hold the Expenses structs
	exps := []*Expense{}

	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by
	// the rows.Scan() method. If iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed Expense struct
		exp := &Expense{}
		// Use rows.Scan() to copy the values from each field in the row, 
		// the arguments to row.Scan() must be pointers 
		err = rows.Scan(&exp.ExpenseId, &exp.UserId, &exp.CategoryId, &exp.Description, &exp.ExpenseType, &exp.AmountInCents, &exp.CreatedAt)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of expenses
		exps = append(exps, exp)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve
	// any error that was encountered during the iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the expenses slice.
	return exps, nil
}

// update
func (m *ExpenseModel) Put(expenseId, userId, categoryId, description, expenseType string, amountInCents int64) error {
	// SQL statement we want to execute
	stmt := `UPDATE Expenses 
			SET categoryId = ?, 
			description = ?, 
			expenseType = ?,
			amountInCents = ?  
			WHERE expenseId = ?
			and userId = ?`

	// Execute the statement with the provided id and body
	_, err := m.DB.Exec(stmt, categoryId, description, expenseType, amountInCents, expenseId, userId)
	if err != nil {
		log.Printf("Error while attempting Expense update %s", err)
		return err
	}

	log.Printf("Updated successfully")
	return nil
}

// delete
func (m *ExpenseModel) Delete(expenseId, userId string) error {
	// Execute the statement with the provided id
	stmt := `DELETE FROM Expenses WHERE expenseId = ? and userId = ?`

	result, err := m.DB.Exec(stmt, expenseId, userId)
	if err != nil {
		log.Printf("Error while deleting a Expense: %s", err)
		return err
	}

	// Check if the record was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error while checking rows affected: %s", err)
		return err
	}

	if rowsAffected == 0 {
		// No rows were affected, meaning the ID might not exist
		log.Printf("No rows affected, possible non-existent ID: %s", expenseId)
		return nil
	}

	log.Printf("Deleted successfully")
	return nil
}

func (m *ExpenseModel) DeleteAll(userId string) error {
	stmt := `DELETE FROM Expenses
			WHERE userId = ?`
	
	result, err := m.DB.Exec(stmt, userId);
	if err != nil {
		log.Printf("Failed to delete expenses for user %s: %v", userId, err)
		return fmt.Errorf("could not delete expenses: %w", err)
	}

	// Check if the record was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error while checking rows affected: %s", err)
		return err
	}
	
	if rowsAffected == 0 {
		// No rows were affected, meaning the ID might not exist
		log.Printf("No rows affected, possible non-existent ID: %s", userId)
		return nil
	}
	
	log.Printf("Successfully deleted %d expenses for user %s", rowsAffected, userId)
	return nil
}