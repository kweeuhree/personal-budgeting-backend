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
}

// Load() loads the configuration based on the provided environment.
func Load(env string) (*Config, error) {

	switch env {
	case "development":
		return devConfig(), nil
	case "production":
		return prodConfig()
	default:
		return nil, errors.New("unknown environment: " + env)
	}
}

func devConfig() *Config {
	// Local MySQL instance
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	DSNstringVars := []any{dbUser, dbPassword, dbName}
	for indx, value := range DSNstringVars {
		if value == nil {
			fmt.Printf("%s at index %d is nil", value, indx)
		}
	}

	// Define new command-line flag for the mysql dsn string
	DSNstring := fmt.Sprintf("%s:%s@/%s?parseTime=true", DSNstringVars...)
	dsn := flag.String("dsn", DSNstring, "MySQL data source name")

	// Session manager configuration
	sessionManager := scs.New()
	// Use the MySQL session store with the session manager.
	sessionManager.Lifetime = 12 * time.Hour

	// assembly implementations are used.
	tlsConfig := &tls.Config{
		// if you were to make TLS 1.3 the minimum supported
		// version in the TLS config for your server, then all browsers able to
		// use your application will support SameSite cookies
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	return &Config{
		Addr:           *addr, // Development default port
		DSN:            *dsn,
		TLSConfig:      tlsConfig,
		DebugPprof:     true,
		SessionManager: sessionManager,
	}
}

func prodConfig() (*Config, error) {

	// Load production environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	caAivenCert := os.Getenv("CA_AIVEN_CERT")

	prodVars := []any{dbUser, dbPassword, dbName, dbPort, dbName, caAivenCert}
	for _, value := range prodVars {
		if value == nil {
			fmt.Printf("%s is nil", value)
		}
	}

	log.Println("Loading CA certificate...")
	// Load Aiven CA certificate
	rootCertPool := x509.NewCertPool()
	pem, err := os.ReadFile(caAivenCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return nil, errors.New("failed to append CA certificate PEM")
	}

	// Register TLS config with MySQL driver
	err = mysql.RegisterTLSConfig("aiven", &tls.Config{
		RootCAs: rootCertPool,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register TLS config: %v", err)
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

	// assembly implementations are used.
	tlsConfig := &tls.Config{
		// if you were to make TLS 1.3 the minimum supported
		// version in the TLS config for your server, then all browsers able to
		// use your application will support SameSite cookies
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	return &Config{
		Addr:           fmt.Sprintf(":%s", dbPort),
		DSN:            *dsn,
		DebugPprof:     false,
		SessionManager: sessionManager,
		TLSConfig:      tlsConfig,
	}, nil
}
