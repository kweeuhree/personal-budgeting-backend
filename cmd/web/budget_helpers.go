package main

import (
	"errors"
	"fmt"

	"kweeuhree.personal-budgeting-backend/internal/models"
)

func (app *application) CurrentBudgetIsValid(currentBudget *models.Budget, balanceType string, sumInCents int64) (error) {
    if currentBudget == nil {
        return errors.New("current budget cannot be null")
    }
    if sumInCents <= 0 {
        return errors.New("balance amount must be positive")
    }
    if balanceType != BalanceTypeChecking && balanceType != BalanceTypeSavings {
        return fmt.Errorf("invalid balance type: %s", balanceType)
    }
    if (balanceType == BalanceTypeChecking && currentBudget.CheckingBalance < sumInCents) ||
       (balanceType == BalanceTypeSavings && currentBudget.SavingsBalance < sumInCents) {
        return fmt.Errorf("insufficient funds in %s account", balanceType)
    }
    return nil
}

func (app *application) CalculateUpdates(
    currentBudget *models.Budget, updateType, balanceType string, sumInCents int64, isExpense bool,
) (int64, int64, int64, int64, int64, error) {
    updatedBudget := *currentBudget

    // Adjust checking or savings balance based on balance type
    if balanceType == BalanceTypeChecking {
        updatedBudget.CheckingBalance = app.updateBalance(updatedBudget.CheckingBalance, sumInCents, updateType)
    } else if balanceType == BalanceTypeSavings {
        updatedBudget.SavingsBalance = app.updateBalance(updatedBudget.SavingsBalance, sumInCents, updateType)
    }

    // Calculate new budget remaining
    newBudgetRemaining := updatedBudget.CheckingBalance + updatedBudget.SavingsBalance

    // Calculate new budget total and total spent
    var newBudgetTotal, newTotalSpent int64
    switch updateType {
    case UpdateTypeAdd:
        newBudgetTotal = updatedBudget.BudgetTotal + sumInCents
        if isExpense {
            newTotalSpent = updatedBudget.TotalSpent - sumInCents
        }
    case UpdateTypeSubtract:
        newBudgetTotal = updatedBudget.BudgetTotal - sumInCents
        if isExpense {
            newTotalSpent = updatedBudget.TotalSpent + sumInCents
        }
    }

    if newTotalSpent < 0 {
        newTotalSpent = 0
    }

    return updatedBudget.CheckingBalance, updatedBudget.SavingsBalance, newBudgetTotal, newBudgetRemaining, newTotalSpent, nil
}


func (app *application) updateBalance(currentBalance, sumInCents int64, updateType string) int64 {
	var updatedBalance int64
	switch updateType {
	case UpdateTypeAdd:
		updatedBalance = currentBalance + sumInCents
	case UpdateTypeSubtract:
		updatedBalance = currentBalance - sumInCents
	}
	return updatedBalance
}