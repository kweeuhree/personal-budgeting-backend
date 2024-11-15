package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// define User type
type User struct {
	Uuid           string
	Email          string
	DisplayName    string
	HashedPassword []byte
	CreatedAt      time.Time
}

// define UserModel type which wraps a database connection pool
type UserModel struct {
	DB *sql.DB
}

// add a new record to the users table
func (m *UserModel) Insert(newId, email, displayName, password string) error {
	fmt.Println("Attempting to insert new user into database...")

	// create a bcrypt hash of the plain-text password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (uuid, email, displayName, hashedPassword, createdAt)
VALUES(?, ?, ?, ?, UTC_TIMESTAMP())`

	// insert with Exec()
	_, err = m.DB.Exec(stmt, newId, email, displayName, string(hashedPassword))
	if err != nil {
		// If this returns an error, we use the errors.As() function to check
		// whether the error has the type *mysql.MySQLError. If it does, the
		// error will be assigned to the mySQLError variable. We can then check
		// whether or not the error relates to our users_uc_email key by
		// checking if the error code equals 1062 and the contents of the error
		// message string. If it does, we return an ErrDuplicateEmail error.
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 &&
				strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

// Authenticate method verifies whether a user exists with the provided email
// and password. Returns relevant user ID
func (m *UserModel) Authenticate(email, password string) (string, error) {
	// Retrieve the id and hashed password associated with the given email.

	// If  no matching email exists we return the ErrInvalidCredentials error.
	var uuid string
	var hashedPassword []byte
	stmt := "SELECT uuid, hashed_password FROM users WHERE email = ?"
	err := m.DB.QueryRow(stmt, email).Scan(&uuid, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "Invalid credentials", ErrInvalidCredentials
		} else {
			return "", err
		}
	}
	// Check whether the hashed password and plain-text password provided match.
	// If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "Invalid credentials", ErrInvalidCredentials
		} else {
			return "", err
		}
	}
	// Otherwise, the password is correct. Return the user ID.
	return uuid, nil

}

// Exists method checks if a user exists with a specific ID.
func (m *UserModel) Exists(uuid string) (bool, error) {
	var exists bool
	stmt := "SELECT EXISTS(SELECT true FROM users WHERE uuid = ?)"

	err := m.DB.QueryRow(stmt, uuid).Scan(&exists)

	return exists, err
}
