package models

import (
	"database/sql"
	"errors"
	"log"
)

// define a ExpenseCategory type
type ExpenseCategory struct {
	ExpenseCategoryId string
	UserId            string
	Name              string
	Description       string
	TotalSum          int64
}

// define a ExpenseCategory model type which wraps a sql.DB connection pool
type ExpenseCategoryModel struct {
	DB *sql.DB
}

// insert a new ExpenseCategory into the database
func (m *ExpenseCategoryModel) Insert(expenseCategoryId, userId, name, description string, totalSum int64) (string, error) {
	// use placeholder parameters instead of interpolating data in the SQL query
	// as this is untrusted user input from a form
	stmt := `INSERT INTO ExpenseCategory (expenseCategoryId, userId, name, description, totalSum) 	
			VALUES(?, ?, ?, ?, ?)`

	_, err := m.DB.Exec(stmt, expenseCategoryId, userId, name, description, totalSum)
	if err != nil {
		return "", err
	}

	return expenseCategoryId, nil
}

// return a specific expenseCategory based on its id
func (m *ExpenseCategoryModel) Get(expenseCategoryId string) (*ExpenseCategory, error) {
	stmt := `SELECT expenseCategoryId, userId, name, description, totalSum, 
			FROM ExpenseCategory WHERE expenseCategoryId = ?`

	// This returns a pointer to a sql.Row object
	// which holds the result from the database
	row := m.DB.QueryRow(stmt, expenseCategoryId)

	// Initialize a pointer to a new zeroed ExpenseCategory struct
	exp := &ExpenseCategory{}

	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the ExpenseCategory struct.
	// The arguments to row.Scan are *pointers* to the place to copy the data into
	err := row.Scan(&exp.ExpenseCategoryId, &exp.UserId, &exp.Name, &exp.Description, &exp.TotalSum)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a
		// sql.ErrNoRows error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	// If everything went OK then return the ExpenseCategory object
	return exp, nil
}

// return all created ExpenseCategories
func (m *ExpenseCategoryModel) All(userId string) ([]*ExpenseCategory, error) {
	stmt := `SELECT * FROM ExpenseCategory
			WHERE userId = ?`

	// Use the Query() method on the connection pool to execute the stmt
	// this returns a sql.Rows resultset containing the result of query
	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}

	// We defer rows.Close() to ensure the sql.Rows resultset is always
	// properly closed before the Latest() method returns
	// Defer stmt should come *after* you check for an error from the Query()
	// method. Otherwise, if Query() returns an error, you'll get a panic trying to close a nil resultset
	defer rows.Close()

	// Initialize an empty slice to hold the ExpenseCategory structs
	expenseCategories := []*ExpenseCategory{}

	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by
	// the rows.Scan() method. If iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed ExpenseCategory struct.
		exp := &ExpenseCategory{}
		// Use rows.Scan() to copy the values from each field in the row to
		// the new ExpenseCategory object
		err = rows.Scan(&exp.ExpenseCategoryId, &exp.UserId, &exp.Name, &exp.Description, &exp.TotalSum)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of ExpenseCategories
		expenseCategories = append(expenseCategories, exp)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve
	// any error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the ExpenseCategories slice
	return expenseCategories, nil
}

// return all created ExpenseCategories
func (m *ExpenseCategoryModel) AllExpensesPerCategory(categoryId string) ([]*Expense, error) {
	stmt := `SELECT expenseId, userId, categoryId, description, expenseType, amountInCents, createdAt 
	 		FROM Expenses WHERE categoryId = ?
			ORDER BY createdAt DESC`

	// this returns a sql.Rows resultset containing the result of query
	rows, err := m.DB.Query(stmt, categoryId)
	if err != nil {
		return nil, err
	}

	// We defer rows.Close() to ensure the sql.Rows resultset is always
	// properly closed before the Latest() method returns
	// Defer stmt should come *after* you check for an error from the Query()
	// method. Otherwise, if Query() returns an error, you'll get a panic trying to close a nil resultset
	defer rows.Close()

	// Initialize an empty slice to hold the ExpenseCategory structs
	exps := []*Expense{}

	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by
	// the rows.Scan() method. If iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed ExpenseCategory struct.
		exp := &Expense{}
		// Use rows.Scan() to copy the values from each field in the row to
		// the new ExpenseCategory object
		err = rows.Scan(&exp.ExpenseId, &exp.UserId, &exp.CategoryId, &exp.Description, &exp.ExpenseType, &exp.AmountInCents, &exp.CreatedAt)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of ExpenseCategories
		exps = append(exps, exp)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve
	// any error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the ExpenseCategories slice
	return exps, nil
}

// update an expense category
func (m *ExpenseCategoryModel) Put(userId, expenseCategoryId, name, description string) error {
	stmt := `UPDATE ExpenseCategory 
			SET name = ?,
			description = ? 
			WHERE expenseCategoryId = ? and
			userId = ?`

	// Execute the statement
	_, err := m.DB.Exec(stmt, name, description, expenseCategoryId, userId)
	if err != nil {
		log.Printf("Error while attempting ExpenseCategory update %s", err)
		return err
	}

	log.Printf("Updated successfully")
	return nil
}

// delete an expense category
func (m *ExpenseCategoryModel) Delete(expenseCategoryId, userId string) error {
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

func (m *ExpenseCategoryModel) PutTotalSum(userId, categoryId string, amount int64) error {
	stmt := `UPDATE ExpenseCategory 
			SET totalSum = ? 
			WHERE expenseCategoryId = ? and
			userId = ?`

	// Execute the statement
	_, err := m.DB.Exec(stmt, amount, categoryId, userId)
	if err != nil {
		log.Printf("Error while attempting updating ExpenseCategory total sum %s", err)
		return err
	}

	log.Printf("Updated successfully")
	return nil
}

// return a specific expenseCategory based on its id
func (m *ExpenseCategoryModel) GetCategoryTotalSum(userId, expenseCategoryId string) (int64, error) {
	stmt := `SELECT totalSum 
			FROM ExpenseCategory 
			WHERE userId = ? and expenseCategoryId = ?`

	// This returns a pointer to a sql.Row object
	// which holds the result from the database
	row := m.DB.QueryRow(stmt, userId, expenseCategoryId)

	var totalSum int64

	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the ExpenseCategory struct.
	// The arguments to row.Scan are *pointers* to the place to copy the data into
	err := row.Scan(&totalSum)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a
		// sql.ErrNoRows error
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNoRecord
		} else {
			return 0, err
		}
	}
	// If everything went OK then return the totalSum
	return totalSum, nil
}

// Void all expense categories totalSums upon resetting user budget
func (m *ExpenseCategoryModel) VoidAllTotalSums(userId string) error {
	stmt := `UPDATE ExpenseCategory 
			SET totalSum = 0 
			WHERE userId = ?`

	// Execute the statement
	_, err := m.DB.Exec(stmt, userId)
	if err != nil {
		log.Printf("Error while attempting voiding ExpenseCategory totalSums %s", err)
		return err
	}

	log.Printf("Voided successfully")
	return nil
}
