package models

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

// define a Budget type
type Budget struct {
	BudgetId       	 string
	UserId			 string
	CheckingBalance  int64
	SavingsBalance   int64
	BudgetTotal      int64
	BudgetRemaining  int64
	TotalSpent		 int64
	UpdatedAt		 time.Time
	CreatedAt        time.Time
}

// define a Budget model type which wraps a sql.DB connection pool
type BudgetModel struct {
	DB *sql.DB
}

// insert a new Budget into the database
func (m *BudgetModel) Insert(budgetId, userId string, checkingBalance, savingsBalance, budgetTotal int64) (string, error) {
	stmt := `INSERT INTO Budget (
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
	// Write the SQL statement we want to execute.
	stmt := `SELECT budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent, updatedAt, createdAt
			FROM Budget WHERE budgetId = ?`

	// Use the QueryRow() method on the connection pool to execute our
	// SQL statement, passing in the untrusted id variable as the value for
	// the placeholder parameter. This returns a pointer to a sql.Row object
	// which holds the result from the database.
	row := m.DB.QueryRow(stmt, budgetId)

	// Initialize a pointer to a new zeroed Snippet struct.
	bud := &Budget{}

	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the Snippet struct. Notice that the arguments
	// to row.Scan are *pointers* to the place you want to copy the data
	// into, and the number of arguments must be exactly the same as the number of
	// columns returned by your statement.
	err := row.Scan(
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
	// If everything went OK then return the Budget object.
	return bud, nil
}

// return the all created Budgets
func (m *BudgetModel) All() ([]*Budget, error) {
	// SQL statement we want to execute
	stmt := `SELECT * FROM Budget
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

	// Initialize an empty slice to hold the budget structs
	budget := []*Budget{}

	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by
	// the rows.Scan() method. If iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed budget struct.
		bud := &Budget{}
		// Use rows.Scan() to copy the values from each field in the row to
		// the new budget object that we created. Again, the arguments to
		// row.Scan() must be pointers to the place you want to copy the data into, and
		// the number of arguments must be exactly the same as the number of
		// columns returned by your statement.
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
		// Append it to the slice of snippets.
		budget = append(budget, bud)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve
	// any error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the budget slice.
	return budget, nil
}

// update
func (m *BudgetModel) Put(budgetId, userId string, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent int64) error {
	// SQL statement we want to execute
	stmt := `UPDATE Budget 
			SET checkingBalance = ?, 
			savingsBalance = ?, 
			budgetTotal = ?,
			budgetRemaining = ?,
			totalSpent = ?,
			updatedAt = UTC_TIMESTAMP()
			WHERE budgetId = ?
			and userId = ?`
	
	// Execute the statement with the provided id and body
	_, err := m.DB.Exec(stmt, budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent)
	if err != nil {
		log.Printf("Error while attempting Budget update for budgetId: %s, userId: %s - %v", budgetId, userId, err)
		return err
	}

	log.Printf("Updated successfully")
	return nil
}

// delete
func (m *BudgetModel) Delete(budgetId, userId string) error {
	// Execute the statement with the provided id
	stmt := `DELETE FROM Budget WHERE budgetId = ? and userId = ?`

	result, err := m.DB.Exec(stmt, budgetId, userId)
	if err != nil {
		log.Printf("Error while deleting a Budget: %s", err)
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
		log.Printf("No rows affected, possible non-existent ID: %s", budgetId)
		return nil
	}

	log.Printf("Deleted successfully")
	return nil
}


func (m *BudgetModel) GetBudgetByUserId(userId string) (*Budget, error) {
	stmt := `SELECT budgetId, userId, checkingBalance, savingsBalance, budgetTotal, budgetRemaining, totalSpent 
			FROM Budget WHERE userId = ?`
	row := m.DB.QueryRow(stmt, userId)

	budget := &Budget{}
	err := row.Scan(&budget.BudgetId, &budget.UserId, &budget.CheckingBalance, &budget.SavingsBalance, &budget.BudgetTotal, &budget.BudgetRemaining, &budget.TotalSpent)
	if err != nil {
		return nil, err
	}

	return budget, nil
}