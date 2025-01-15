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
	Addr string
	DSN  string
	// TLSConfig      *tls.Config
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

	// Define new command-line flag for the mysql dsn string
	DSNstring := fmt.Sprintf("%s:%s@/%s?parseTime=true", dbUser, dbPassword, dbName)
	dsn := flag.String("dsn", DSNstring, "MySQL data source name")

	// Session manager configuration
	sessionManager := scs.New()
	// Use the MySQL session store with the session manager.
	sessionManager.Lifetime = 12 * time.Hour

	return &Config{
		Addr:           *addr, // Development default port
		DSN:            *dsn,
		DebugPprof:     true,
		SessionManager: sessionManager,
	}
}

func prodConfig() (*Config, error) {
	// Load production environment variables
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	caAivenCert := os.Getenv("CA_AIVEN_CERT")

	// if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" || dbPort == "" || caAivenCert == "" {
	// 	return nil, errors.New("missing required production environment variables")
	// }

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

	dsn := flag.String("dsn", DSNstring, "MySQL data source name")

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
		// TLSConfig:      tlsConfig,
	}, nil
}
