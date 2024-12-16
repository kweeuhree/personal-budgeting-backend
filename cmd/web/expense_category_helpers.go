package main

import "log"

const (
	Increment = "increment"
	Decrement = "decrement"
)

func (app *application) UpdateCategoryExpenses(userId, categoryId, updateType string, amount int64) error {

	currentTotalSum, err := app.expenseCategory.GetCategoryTotalSum(userId, categoryId)

	if err != nil {
		return err
	}
	newTotalSum := app.calculateTotalSum(currentTotalSum, amount, updateType)

	err = app.expenseCategory.PutTotalSum(userId, categoryId, newTotalSum)

	if err != nil {
		return err
	}

	log.Printf("Updated successfully")
	return nil
}

func (app *application) calculateTotalSum(currentTotalSum, amount int64, updateType string) int64 {
	var total int64
	switch updateType {
	case Increment:
		total = currentTotalSum + amount
	case Decrement:
		total = currentTotalSum - amount
		if total < 0 {
			total = 0
		}
	}

	return total
}
