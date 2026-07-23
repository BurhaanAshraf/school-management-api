package sqlconnect

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
)

func registerTLS() error {
	rootCertPool, err := x509.SystemCertPool()
	if err != nil {
		return err
	}
	if rootCertPool == nil {
		rootCertPool = x509.NewCertPool()
	}

	return mysql.RegisterTLSConfig("tidb", &tls.Config{
		RootCAs:    rootCertPool,
		MinVersion: tls.VersionTLS12,
	})
}

func ConnectDB() (*sql.DB, error) {

	if err := registerTLS(); err != nil {
		return nil, err
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=tidb&parseTime=true", user, password, host, dbPort, dbName)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil

}
