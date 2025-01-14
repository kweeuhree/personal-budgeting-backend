package models

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

// define a Budget type
type Budget struct {
	BudgetId        string
	UserId          string
	CheckingBalance int64
	SavingsBalance  int64
	BudgetTotal     int64
	BudgetRemaining int64
	TotalSpent      int64
	UpdatedAt       time.Time
	CreatedAt       time.Time
}

// define a Budget model type which wraps a sql.DB connection pool
type BudgetModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func NewBudgetModel(db *sql.DB, infoLog *log.Logger, errorLog *log.Logger) *BudgetModel {
	return &BudgetModel{
		DB:       db,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}
}

// insert a new Budget into the database
func (m *BudgetModel) Insert(budgetId, userId string, checkingBalance, savingsBalance, budgetTotal int64) (string, error) {
	stmt := `INSERT INTO budget (
		budgetId, userId, checkingBalance, savingsBalance, budgetTotal, 
		budgetRemaining, totalSpent, createdAt, updatedAt
	) 
		VALUES (?, ?, ?, ?, ?, ?, ?, UTC_TIMESTAMP(), UTC_TIMESTAMP())`

	budgetRemaining := budgetTotal
	totalSpent := int64(0)

	_, err := m.DB.Exec(stmt, budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent)
	if err != nil {
		return "", err
	}

	return budgetId, nil
}

// return a specific Budget based on its id
func (m *BudgetModel) Get(budgetId string) (*Budget, error) {
	stmt := `SELECT budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent, updatedAt, createdAt
			FROM budget WHERE budgetId = ?`

	// This returns a pointer to a sql.Row object
	// which holds the result from the database
	row := m.DB.QueryRow(stmt, budgetId)

	// Initialize a pointer to a new zeroed Budget struct
	bud := &Budget{}

	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the Snippet struct.
	// The arguments to row.Scan are *pointers* to the place to copy the data into
	err := row.Scan(
		&bud.BudgetId,
		&bud.UserId,
		&bud.CheckingBalance,
		&bud.SavingsBalance,
		&bud.BudgetTotal,
		&bud.BudgetRemaining,
		&bud.TotalSpent,
		&bud.UpdatedAt,
		&bud.CreatedAt,
	)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a
		// sql.ErrNoRows error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	// If everything went OK then return the Budget object
	return bud, nil
}

// return all created Budgets
func (m *BudgetModel) All() ([]*Budget, error) {
	stmt := `SELECT * FROM budget
			ORDER BY created DESC`

	// Use the Query() method on the connection pool to execute the stmt
	// this returns a sql.Rows resultset containing the result of query
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// We defer rows.Close() to ensure the sql.Rows resultset is always
	// properly closed.
	// Defer stmt should come *after* you check for an error from the Query()
	// method. Otherwise, if Query() returns an error, you'll get a panic trying to close a nil resultset
	defer rows.Close()

	// Initialize an empty slice to hold the Budget structs
	budgets := []*Budget{}

	// Use rows.Next to iterate through the rows in the resultset.
	// If iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed Budget struct.
		bud := &Budget{}
		// Use rows.Scan() to copy the values from each field in the row to
		// the new budget object.
		err = rows.Scan(
			&bud.BudgetId,
			&bud.UserId,
			&bud.CheckingBalance,
			&bud.SavingsBalance,
			&bud.BudgetTotal,
			&bud.BudgetRemaining,
			&bud.TotalSpent,
			&bud.CreatedAt,
			&bud.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of Budgets.
		budgets = append(budgets, bud)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve
	// any error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the Budgets slice
	return budgets, nil
}

// update a Budget
func (m *BudgetModel) Put(budgetId, userId string, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent int64) error {
	stmt := `UPDATE budget 
			SET checkingBalance = ?, 
			savingsBalance = ?, 
			budgetTotal = ?,
			budgetRemaining = ?,
			totalSpent = ?,
			updatedAt = UTC_TIMESTAMP()
			WHERE budgetId = ?
			and userId = ?`

	// Execute the statement with the provided id and body
	_, err := m.DB.Exec(stmt, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent, budgetId, userId)
	if err != nil {
		m.ErrorLog.Printf("Error while attempting Budget update for budgetId: %s, userId: %s - %v", budgetId, userId, err)
		return err
	}

	log.Printf("Updated successfully")
	return nil
}

// delete a Budget
func (m *BudgetModel) Delete(budgetId, userId string) error {
	// Execute the statement with the provided id
	stmt := `DELETE FROM budget WHERE budgetId = ? and userId = ?`

	result, err := m.DB.Exec(stmt, budgetId, userId)
	if err != nil {
		m.ErrorLog.Printf("Error while deleting a Budget: %s", err)
		return err
	}

	// Check if the record was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		m.ErrorLog.Printf("Error while checking rows affected: %s", err)
		return err
	}

	if rowsAffected == 0 {
		// No rows were affected, meaning the ID might not exist
		m.ErrorLog.Printf("No rows affected, possible non-existent ID: %s", budgetId)
		return nil
	}

	log.Printf("Deleted successfully")
	return nil
}

// Find a budget based on UserId
func (m *BudgetModel) GetBudgetByUserId(userId string) (*Budget, error) {
	stmt := `SELECT budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent 
			FROM budget WHERE userId = ?`
	row := m.DB.QueryRow(stmt, userId)

	budget := &Budget{}
	err := row.Scan(&budget.BudgetId, &budget.UserId, &budget.CheckingBalance, &budget.SavingsBalance, &budget.BudgetTotal, &budget.BudgetRemaining, &budget.TotalSpent)
	if err != nil {
		return nil, err
	}

	return budget, nil
}
