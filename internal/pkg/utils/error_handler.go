package utils

import (
	"fmt"
	"log"
	"os"
)

func ErrorHandler(err error, msg string) error {
	errorLogger := log.New(os.Stderr, "ERROR:", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger.Println(msg, err)
	return fmt.Errorf("%s", msg)
}
