package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/joho/godotenv"
	"kweeuhree.personal-budgeting-backend/internal/models"

	// models

	// environment variables

	// we need the driver’s init() function to run so that it can register itself with the
	// database/sql package. The trick to getting around this is to alias the package name
	// to the blank identifier. This is standard practice for most of Go’s SQL drivers

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
)

// with underscore

// Define an application struct to hold the application-wide dependencies for
// the web application.
type application struct {
	errorLog        *log.Logger
	infoLog         *log.Logger
	user            *models.UserModel
	budget          *models.BudgetModel
	expenses        *models.ExpenseModel
	expenseCategory *models.ExpenseCategoryModel
	sessionManager  *scs.SessionManager
}

func main() {
	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// DSN string with loaded env variables
	DSNstring := fmt.Sprintf("%s:%s@/%s?parseTime=true", dbUser, dbPassword, dbName)

	// define  new command-line flag for the mysql dsn string
	dsn := flag.String("dsn", DSNstring, "MySQL data source name")
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "4000" // default port
	}
	// addr := flag.String("addr", ":4000", "HTTP network address")

	// parse flags
	flag.Parse()

	// error and info logs
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	log.Println("Starting database connection...")
	// create connection pool, pass openDB() the dsn from the command-line flag
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	log.Println("Database connected successfully")
	// defer a call to db.Close() so that the connection pool is closed before
	// the main() function exits
	defer db.Close()

	// Create a new MySQL session store using the connection pool.
	store := mysqlstore.New(db)
	// Initialize a new session manager.
	sessionManager := scs.New()
	// Use the MySQL session store with the session manager.
	sessionManager.Store = store
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		errorLog:        errorLog,
		infoLog:         infoLog,
		user:            &models.UserModel{DB: db},
		budget:          &models.BudgetModel{DB: db},
		expenses:        &models.ExpenseModel{DB: db},
		expenseCategory: &models.ExpenseCategoryModel{DB: db},
		sessionManager:  sessionManager,
	}

	// Initialize a tls.Config struct to hold the non-default TLS settings we
	// want the server to use. In this case the only thing that we're
	// changing is the curve preferences value, so that only elliptic curves with
	// assembly implementations are used.
	tlsConfig := &tls.Config{
		// if you were to make TLS 1.3 the minimum supported
		// version in the TLS config for your server, then all browsers able to
		// use your application will support SameSite cookies
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	log.Println("Starting server...")
	srv := &http.Server{
		Addr:      ":" + addr,
		ErrorLog:  errorLog,
		Handler:   app.routes(),
		TLSConfig: tlsConfig,
		// connection timeouts
		// -- all keep-alive connections will be automatically closed
		// -- after 1 minute of inactivity
		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		// -- prevent the data that the handler returns
		// -- from taking too long to write
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", addr)

	// ListenAndServeTLS() starts HTTPS server
	// err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	err = srv.ListenAndServe()
	// in case of errors log and exit
	errorLog.Fatal(err)
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
