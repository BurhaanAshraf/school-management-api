package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerifyPassword(password, encodedHash string) error {
	parts := strings.Split(encodedHash, ".")
	if len(parts) != 2 {
		return ErrorHandler(errors.New("Invalid hash format"), "Internal Server Error")
	}
	saltBase64 := parts[0]
	hashBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)

	if err != nil {
		return ErrorHandler(errors.New("decoding salt failed..."), "Internal Server Error")

	}

	hashPassword, err := base64.StdEncoding.DecodeString(hashBase64)

	if err != nil {
		return ErrorHandler(err, "failed to decode hash")

	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	if len(hashPassword) != len(hash) {
		return ErrorHandler(errors.New("hash length mismatch"), "Wrong password")

	}
	if subtle.ConstantTimeCompare(hash, hashPassword) != 1 {
		return ErrorHandler(errors.New("User entered wrong password"), "Wrong password")

	}
	return nil
}

func HashingPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(errors.New("Failed to generate salt"), "error adding data")
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
	var hashedPassword string = encodedHash
	return hashedPassword, nil
}
