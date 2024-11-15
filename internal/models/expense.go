package models

import (
	"database/sql"
	"errors"
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
	// use placeholder parameters instead of interpolating data in the SQL query
	// as this is untrusted user input from a form
	stmt := `INSERT INTO Expenses (expenseId, userId, categoryId, description, expenseType, amountInCents, createdAt) 	
			VALUES(?, ?, ?, ?, ?, UTC_TIMESTAMP())`

	_, err := m.DB.Exec(stmt, expenseId, userId, categoryId, description, expenseType, amountInCents)
	if err != nil {
		return "", err
	}
	
	// use the LastInserId() method on the result to get the ID of
	// the newly created record in the snippets table
	// id, err := result.LastInsertId()
	// if err != nil {
	// 	return 0, err
	// }

	return expenseId, nil
}

// return a specific expense based on its id
func (m *ExpenseModel) Get(expenseId string) (*Expense, error) {
	// Write the SQL statement we want to execute.
	stmt := `SELECT expenseId, userId, categoryId, description, expenseType, amountInCents 
			FROM Expenses WHERE expenseId = ?`

	// Use the QueryRow() method on the connection pool to execute our
	// SQL statement, passing in the untrusted id variable as the value for
	// the placeholder parameter. This returns a pointer to a sql.Row object
	// which holds the result from the database.
	row := m.DB.QueryRow(stmt, expenseId)

	// Initialize a pointer to a new zeroed Snippet struct.
	exp := &Expense{}

	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the Snippet struct. Notice that the arguments
	// to row.Scan are *pointers* to the place you want to copy the data
	// into, and the number of arguments must be exactly the same as the number of
	// columns returned by your statement.
	err := row.Scan(&exp.ExpenseId, &exp.UserId, &exp.CategoryId, &exp.Description, &exp.ExpenseType, &exp.AmountInCents)
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
	// If everything went OK then return the Expense object.
	return exp, nil
}

// return the all created Expenses
func (m *ExpenseModel) All() ([]*Expense, error) {
	// SQL statement we want to execute
	stmt := `SELECT * FROM Expenses
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
	Expenses := []*Expense{}

	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by
	// the rows.Scan() method. If iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed Snippet struct.
		exp := &Expense{}
		// Use rows.Scan() to copy the values from each field in the row to
		// the new Snippet object that we created. Again, the arguments to
		// row.Scan() must be pointers to the place you want to copy the data into, and
		// the number of arguments must be exactly the same as the number of
		// columns returned by your statement.
		err = rows.Scan(&exp.ExpenseId, &exp.UserId, &exp.CategoryId, &exp.Description, &exp.ExpenseType, &exp.AmountInCents)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of snippets.
		Expenses = append(Expenses, exp)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve
	// any error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the Snippets slice.
	return Expenses, nil
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
func (m *ExpenseModel) Delete(expenseId string) error {
	// Execute the statement with the provided id
	stmt := `DELETE FROM Expenses WHERE expenseId = ?`

	result, err := m.DB.Exec(stmt, expenseId)
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


// toggle status
// func (m *ExpenseModel) Toggle(id string) error {
// 	// SQL statement we want to execute
// 	stmt := `UPDATE Expenses SET status = !status WHERE id = ?`

// 	// Execute the statement with the provided id and body
// 	_, err := m.DB.Exec(stmt, id)
// 	if err != nil {
// 		log.Printf("Error while attempting Expense status toggle %s", err)
// 		return err
// 	}

// 	log.Printf("Status toggled successfully")
// 	return nil
// }