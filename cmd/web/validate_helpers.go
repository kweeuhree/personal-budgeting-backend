package main

import (
	"strconv"

	"kweeuhree.personal-budgeting-backend/internal/validator"
)

func (input *ExpenseInput) Validate() {
	amountInCentsStr := strconv.FormatInt(input.AmountInCents, 10)
	input.CheckField(validator.NotBlank(amountInCentsStr), "amountInCents", "This field cannot be blank")
}

func (input *ExpenseCategoryInput) Validate() {
	input.CheckField(validator.NotBlank(input.Name), "amountInCents", "This field cannot be blank")
}

func (form *UserSignUpInput) Validate() {
	form.CheckField(validator.NotBlank(form.DisplayName), "displayName", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
}

// checks that email and password are provided
// and also check the format of the email address as
// a UX-nicety (in case the user makes a typo).
func (form *UserLoginInput) Validate() {
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
}

func (input *BudgetInput) Validate() {
	checkingBalanceStr := strconv.FormatInt(input.CheckingBalance, 10)
	input.CheckField(validator.NotBlank(checkingBalanceStr), "checkingBalance", "This field cannot be blank")
}

func (input *BudgetUpdate) Validate() {
	sumInCentsStr := strconv.FormatInt(input.UpdateSumInCents, 10)
	input.CheckField(validator.NotBlank(sumInCentsStr), "updateSumInCents", "This field cannot be blank")
	input.CheckField(validator.NotBlank(input.UpdateType), "updateType", "This field cannot be blank")
}
