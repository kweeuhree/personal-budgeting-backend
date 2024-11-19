package main

import (
	"kweeuhree.personal-budgeting-backend/internal/models"
)

func (app *application) GetExpensesTotal(exps []*models.Expense) (int64) {
	var total int64
    for _, exp := range exps {
        total += exp.AmountInCents
    }

	return total
}