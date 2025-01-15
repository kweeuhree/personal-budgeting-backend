package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-sql-driver/mysql"
)

type Config struct {
	Addr           string
	DSN            string
	TLSConfig      *tls.Config
	DebugPprof     bool
	SessionManager *scs.SessionManager
	ErrorLog       *log.Logger
}

const (
	Production  = "production"
	Development = "development"
)

// if you were to make TLS 1.3 the minimum supported
// version in the TLS config for your server, then all browsers able to
// use your application will support SameSite cookies
var tlsConfig = &tls.Config{
	MinVersion:       tls.VersionTLS13,
	CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
}

// Load() loads the configuration based on the provided environment.
func Load(env string, errorLog *log.Logger) (*Config, error) {

	switch env {
	case Development:
		return devConfig(errorLog), nil
	case Production:
		return prodConfig(errorLog), nil
	default:
		return nil, errors.New("unknown environment: " + env)
	}
}

func devConfig(errorLog *log.Logger) *Config {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("PORT")

	envVars := []any{dbUser, dbPassword, dbName}
	if len(envVars) < 3 {
		errorLog.Fatalf("failed to load environment variables")
	}

	// Local MySQL instance
	// Define new command-line flag for the mysql dsn string
	DSNstring := fmt.Sprintf("%s:%s@/%s?parseTime=true", envVars...)
	dsn := flag.String("dsn", DSNstring, "MySQL data source name")
	addr := flag.String("addr", port, "HTTP network address")
	flag.Parse()
	// Session manager configuration
	sessionManager := scs.New()
	// Use the MySQL session store with the session manager.
	sessionManager.Lifetime = 12 * time.Hour

	return &Config{
		Addr:           *addr,
		DSN:            *dsn,
		TLSConfig:      tlsConfig,
		DebugPprof:     true,
		SessionManager: sessionManager,
	}
}

func prodConfig(errorLog *log.Logger) *Config {

	// Load production environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	caAivenCert := os.Getenv("CA_AIVEN_CERT")

	envVars := []any{dbUser, dbPassword, dbName, dbPort, dbName, caAivenCert}
	if len(envVars) < 6 {
		errorLog.Fatalf("failed to load environment variables")
	}

	// Load Aiven CA certificate
	rootCertPool := x509.NewCertPool()
	pem, err := os.ReadFile(caAivenCert)
	if err != nil {
		errorLog.Fatalf("failed to read CA certificate: %v", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		errorLog.Fatalf("failed to append CA certificate PEM")
	}

	// Register TLS config with MySQL driver
	err = mysql.RegisterTLSConfig("aiven", &tls.Config{
		RootCAs: rootCertPool,
	})
	if err != nil {
		errorLog.Fatalf("failed to register TLS config: %v", err)
	}

	// Production DSN
	DSNstring := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=aiven&parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	dsn := flag.String("dsn", DSNstring, "Aiven MySQL data source name")

	// Session manager configuration
	sessionManager := scs.New()

	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Secure = true
	sessionManager.Cookie.SameSite = http.SameSiteNoneMode
	sessionManager.Cookie.Path = "/"

	return &Config{
		Addr:           fmt.Sprintf(":%s", dbPort),
		DSN:            *dsn,
		DebugPprof:     false,
		SessionManager: sessionManager,
		TLSConfig:      tlsConfig,
	}
}
