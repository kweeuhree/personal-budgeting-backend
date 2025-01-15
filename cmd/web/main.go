package main

import (
	// The driver’s init() function must be run so that it can register itself with the
	// database/sql package. To ensure this, we use the blank identifier to import
	// the package. This is a common pattern in Go for initializing SQL drivers.
	"database/sql"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"kweeuhree.personal-budgeting-backend/internal/config"
	"kweeuhree.personal-budgeting-backend/internal/models"

	// Load environment variables for development
	"github.com/joho/godotenv"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
)

// Define an application struct to hold the application-wide dependencies for
// the web application
type application struct {
	errorLog        *log.Logger
	infoLog         *log.Logger
	user            *models.UserModel
	budget          *models.BudgetModel
	expenses        *models.ExpenseModel
	expenseCategory *models.ExpenseCategoryModel
	sessionManager  *scs.SessionManager
}

const (
	Production  = "production"
	Development = "development"
)

func main() {
	// This will be populated by Render.com
	env := os.Getenv("ENV")
	// If env was not populated, set it to development
	if env != Production {
		err := godotenv.Load()
		if err != nil {
			log.Printf("Error loading .env file: %v", err)
		}
		env = Development
	}

	// Initialize ppof, error and info logs
	ppofLog := log.New(os.Stdout, "PPOF\t", log.Ldate|log.Ltime)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Load the configuration based on the environment, pass errorLog
	cfg, err := config.Load(env, errorLog)
	if err != nil {
		errorLog.Fatalf("Error loading configuration: %v", err)
	}

	// Enable pprof in development
	if cfg.DebugPprof {
		go func() {
			ppofLog.Println("debugging enabled on /debug/pprof")
			ppofLog.Println(http.ListenAndServe(fmt.Sprintf("localhost%s", cfg.Addr), nil))
		}()
	}

	// Initialize connection with relevant database connection string
	db, err := openDB(cfg.DSN)
	if err != nil {
		errorLog.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// Create a new MySQL session store using the connection pool
	store := mysqlstore.New(db)
	// Use the MySQL session store with the session manager
	cfg.SessionManager.Store = store

	budgetModel := models.NewBudgetModel(db, infoLog, errorLog)

	// Initialize application with its dependencies
	app := &application{
		errorLog:        errorLog,
		infoLog:         infoLog,
		user:            &models.UserModel{DB: db},
		budget:          budgetModel,
		expenses:        &models.ExpenseModel{DB: db},
		expenseCategory: &models.ExpenseCategoryModel{DB: db},
		sessionManager:  cfg.SessionManager,
	}

	infoLog.Printf("Configuring server for %s...", env)
	srv := &http.Server{
		Addr:      cfg.Addr,
		ErrorLog:  errorLog,
		Handler:   app.routes(),
		TLSConfig: cfg.TLSConfig,
		// connection timeouts
		// -- all keep-alive connections will be automatically closed
		// -- after 1 minute of inactivity
		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		// -- prevent the data that the handler returns
		// -- from taking too long to write
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", cfg.Addr)

	switch env {
	case Development:
		// Get self-signed certificates
		cert := os.Getenv("CERT_PEM")
		key := os.Getenv("KEY_PEM")
		// ListenAndServeTLS() starts HTTPS server
		err = srv.ListenAndServeTLS(cert, key)
	case Production:
		// ListenAndServe() starts HTTP server, security is handled by Render.com
		err = srv.ListenAndServe()
	}

	// In case of errors log and exit
	if err != nil {
		errorLog.Fatal(err)
	}
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool for a given dsn
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
