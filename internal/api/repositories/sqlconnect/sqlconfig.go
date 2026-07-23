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
	rootCertPool := x509.NewCertPool()

	pem, err := os.ReadFile("./cert/isrgrootx1.pem")
	if err != nil {
		return err
	}

	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return fmt.Errorf("failed to append CA certificate")
	}

	return mysql.RegisterTLSConfig("tidb", &tls.Config{
		RootCAs: rootCertPool,
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
	// connectionString := "root:Burhan@6965@tcp(127.0.0.1:3306)/" + dbName //Connection String

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
