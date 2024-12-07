package main

import (
	"errors"
	"fmt"
	"log"
)

func (app *application) CurrentBudgetIsValid(userId, balanceType string, sumInCents int64) (error) {
    // Fetch the current budget for the user
    currentBudget, err := app.budget.GetBudgetByUserId(userId)

    if err != nil {
        return fmt.Errorf("failed to fetch current budget: %v", err)
    }

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

func (app *application) CalculateBudgetUpdates(
   userId, updateType, balanceType string, sumInCents int64, isExpense bool,
) (string, int64, int64, int64, int64, int64, error) {
    currentBudget, err := app.budget.GetBudgetByUserId(userId)
    log.Printf("Current Budget: %+v", currentBudget)
    if err != nil {
        return "", 0, 0, 0, 0, 0, fmt.Errorf("failed to fetch current budget: %v", err)
    }

    updatedBudget := *currentBudget

    // Adjust checking or savings balance based on balance type
    if balanceType == BalanceTypeChecking {
        newCheckingBalance := app.updateBalance(updatedBudget.CheckingBalance, sumInCents, updateType)
        updatedBudget.CheckingBalance = newCheckingBalance
    } else if balanceType == BalanceTypeSavings {
        newSavingsBalance := app.updateBalance(updatedBudget.SavingsBalance, sumInCents, updateType)
        updatedBudget.SavingsBalance = newSavingsBalance
    }

    newBudgetRemaining := updatedBudget.CheckingBalance + updatedBudget.SavingsBalance
    newTotalSpent := updatedBudget.TotalSpent

    var newBudgetTotal int64    

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

    return updatedBudget.BudgetId, updatedBudget.CheckingBalance, updatedBudget.SavingsBalance, newBudgetTotal, newBudgetRemaining, newTotalSpent, nil
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