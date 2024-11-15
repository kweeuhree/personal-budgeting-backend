package models

import (
	"database/sql"
	"errors"
	"log"
)

// define a ExpenseCategory type
type ExpenseCategory struct {
	ExpenseCategoryId      string
	UserId    	   		   string
	Name	     		   string
	Description    		   string
}

// define a ExpenseCategory model type which wraps a sql.DB connection pool
type ExpenseCategoryModel struct {
	DB *sql.DB
}

// insert a new ExpenseCategory into the database
func (m *ExpenseCategoryModel) Insert(expenseCategoryId, userId, name, description string) (string, error) {
	// use placeholder parameters instead of interpolating data in the SQL query
	// as this is untrusted user input from a form
	stmt := `INSERT INTO ExpenseCategory (expenseCategoryId, userId, name, description) 	
			VALUES(?, ?, ?, ?)`

	_, err := m.DB.Exec(stmt, expenseCategoryId, userId, name, description)
	if err != nil {
		return "", err
	}

	return expenseCategoryId, nil
}


// return a specific expenseCategory based on its id
func (m *ExpenseCategoryModel) Get(expenseCategoryId string) (*ExpenseCategory, error) {
	// Write the SQL statement we want to execute.
	stmt := `SELECT expenseCategoryId, userId, name, description 
			FROM ExpenseCategory WHERE expenseCategoryId = ?`

	// Use the QueryRow() method on the connection pool to execute our
	// SQL statement, passing in the untrusted id variable as the value for
	// the placeholder parameter. This returns a pointer to a sql.Row object
	// which holds the result from the database.
	row := m.DB.QueryRow(stmt, expenseCategoryId)

	// Initialize a pointer to a new zeroed Snippet struct.
	exp := &ExpenseCategory{}

	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the Snippet struct. Notice that the arguments
	// to row.Scan are *pointers* to the place you want to copy the data
	// into, and the number of arguments must be exactly the same as the number of
	// columns returned by your statement.
	err := row.Scan(&exp.ExpenseCategoryId, &exp.UserId, &exp.Name, &exp.Description)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a
		// sql.ErrNoRows error. We use the errors.Is() function check for
		// that error specifically, and return our own ErrNoRecord error
		// instead (we'll create this in a moment).
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	// If everything went OK then return the ExpenseCategory object.
	return exp, nil
}

// return the all created ExpenseCategorys
func (m *ExpenseCategoryModel) All() ([]*ExpenseCategory, error) {
	// SQL statement we want to execute
	stmt := `SELECT * FROM ExpenseCategory
			ORDER BY created DESC`

	// Use the Query() method on the connection pool to execute the stmt
	// this returns a sql.Rows resultset containing the result of our query
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// We defer rows.Close() to ensure the sql.Rows resultset is always
	// properly closed before the Latest() method returns
	// Defer stmt should come *after* you check for an error from the Query()
	// method. Otherwise, if Query() returns an error, you'll get a panic trying to close a nil resultset
	defer rows.Close()

	// Initialize an empty slice to hold the Snippet structs
	expenseCategories := []*ExpenseCategory{}

	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by
	// the rows.Scan() method. If iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed Snippet struct.
		exp := &ExpenseCategory{}
		// Use rows.Scan() to copy the values from each field in the row to
		// the new Snippet object that we created. Again, the arguments to
		// row.Scan() must be pointers to the place you want to copy the data into, and
		// the number of arguments must be exactly the same as the number of
		// columns returned by your statement.
		err = rows.Scan(&exp.ExpenseCategoryId, &exp.UserId, &exp.Name, &exp.Description)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of snippets.
		expenseCategories = append(expenseCategories, exp)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve
	// any error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the Snippets slice.
	return expenseCategories, nil
}

// update
func (m *ExpenseCategoryModel) Put(userId, expenseCategoryId, name, description string) error {
	// SQL statement we want to execute
	stmt := `UPDATE ExpenseCategory 
			SET name = ?,
			description = ? 
			WHERE expenseCategoryId = ? and
			userId = ?`

	// Execute the statement with the provided id and body
	_, err := m.DB.Exec(stmt, name, description, expenseCategoryId, userId)
	if err != nil {
		log.Printf("Error while attempting ExpenseCategory update %s", err)
		return err
	}

	log.Printf("Updated successfully")
	return nil
}

// delete
func (m *ExpenseCategoryModel) Delete(expenseCategoryId, userId string) error {
	// Execute the statement with the provided id
	stmt := `DELETE FROM ExpenseCategory WHERE expenseCategoryId = ? and userId = ?`

	result, err := m.DB.Exec(stmt, expenseCategoryId, userId)
	if err != nil {
		log.Printf("Error while deleting a ExpenseCategory: %s", err)
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
		log.Printf("No rows affected, possible non-existent ID: %s", expenseCategoryId)
		return nil
	}

	log.Printf("Deleted successfully")
	return nil
}
