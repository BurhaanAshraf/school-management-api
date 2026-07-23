package sqlconnect

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"schoolmanagementapi/internal/api/models"
	"schoolmanagementapi/internal/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-mail/mail/v2"
	"golang.org/x/crypto/argon2"
)

func GetExecsDbHandler(r *http.Request, page, limit int) ([]models.ExecResponse, error, int) {

	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data"), 0
	}
	defer db.Close()

	query := "SELECT id , first_name, last_name, username, email FROM execs WHERE 1=1"

	var args []any

	query, args = AddFilters(r, query, args)
	query = AddSorting(r, query)
	offset := (page - 1) * limit
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data"), 0
	}
	defer rows.Close()

	var execs []models.ExecResponse

	for rows.Next() {
		var exec models.ExecResponse
		err := rows.Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Username, &exec.Email)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error retrieving data"), 0
		}
		execs = append(execs, exec)
	}

	if err := rows.Err(); err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data"), 0
	}
	var totalExecsCount int
	query = GenerateCountQuery("execs", "WHERE 1=1")
	err = db.QueryRow(query).Scan(&totalExecsCount)
	if err != nil {
		utils.ErrorHandler(err, "")
		totalExecsCount = 0
	}
	return execs, nil, totalExecsCount
}

func GetExecByID(ID int) (models.ExecResponse, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.ExecResponse{}, utils.ErrorHandler(err, "error retreiving data")
	}
	defer db.Close()

	var exec models.ExecResponse
	query := "SELECT id , first_name , last_name , username , email FROM execs WHERE id = ?"
	err = db.QueryRow(query, ID).Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Username, &exec.Email)
	if err == sql.ErrNoRows {
		return models.ExecResponse{}, utils.ErrorHandler(err, "error retreiving data")
	} else if err != nil {
		return models.ExecResponse{}, utils.ErrorHandler(err, "error retreiving data")
	}
	return exec, nil
}

func AddExecsDBHandler(newExecs []models.Exec) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer db.Close()

	stmt, err := db.Prepare(GenerateInsertQuery("execs", models.Exec{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer stmt.Close()

	addedexecs := make([]models.Exec, len(newExecs))

	for i, newExec := range newExecs {
		if newExec.Password == "" {
			return nil, utils.ErrorHandler(errors.New("Password is blank"), "please enter password")
		}
		salt := make([]byte, 16)
		_, err := rand.Read(salt)
		if err != nil {
			return nil, utils.ErrorHandler(errors.New("Failed to generate salt"), "error adding data")
		}
		hash := argon2.IDKey([]byte(newExec.Password), salt, 1, 64*1024, 4, 32)
		saltBase64 := base64.StdEncoding.EncodeToString(salt)
		hashBase64 := base64.StdEncoding.EncodeToString(hash)
		encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
		newExec.Password = encodedHash

		values := GetStructValues(newExec)
		res, err := stmt.Exec(values...)
		if err != nil {

			return nil, utils.ErrorHandler(err, "error adding data")
		}
		lastId, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "error adding data")
		}
		newExec.ID = int(lastId)
		addedexecs[i] = newExec
	}
	return addedexecs, nil
}

func PatchExecs(updates []map[string]any) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}

	selectQuery := GenerateSelectQuery("execs", models.Exec{}, "WHERE id = ?")

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			return utils.ErrorHandler(err, "Invalid ID")
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Invalid ID")
		}

		var execFromDB models.Exec
		row := tx.QueryRow(selectQuery, id)
		if err := ScanStruct(row, &execFromDB); err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return utils.ErrorHandler(err, "exec not found")
			}
			return utils.ErrorHandler(err, "error updating data")
		}

		// Apply updates using reflection
		execVal := reflect.ValueOf(&execFromDB).Elem()
		execType := execVal.Type()

		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < execVal.NumField(); i++ {
				field := execType.Field(i)
				if strings.Split(field.Tag.Get("json"), ",")[0] == k {
					fieldVal := execVal.Field(i)
					if fieldVal.CanSet() {
						value := reflect.ValueOf(v)
						if value.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(value.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							log.Printf("Cannot convert %v to %v", value.Type(), fieldVal.Type())
							return utils.ErrorHandler(err, "error updating data")
						}
					}
					break
				}
			}
		}

		updateQuery := GenerateUpdateQuery("execs", execFromDB)
		values := GetUpdateValues(execFromDB)
		values = append(values, execFromDB.ID)

		if _, err = tx.Exec(updateQuery, values...); err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "error updating data")
		}
	}

	if err = tx.Commit(); err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	return nil
}

func PatchOneExec(id int, updates map[string]any) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	var existingexec models.Exec
	query := GenerateSelectQuery("execs", models.Exec{}, "WHERE id = ?")
	row := db.QueryRow(query, id)

	if err := ScanStruct(row, &existingexec); err != nil {
		if err == sql.ErrNoRows {
			return models.Exec{}, utils.ErrorHandler(err, "exec not found")
		}
		return models.Exec{}, utils.ErrorHandler(err, "error updating data")
	}

	execVal := reflect.ValueOf(&existingexec).Elem()
	execType := execVal.Type()

	for k, v := range updates {
		for i := 0; i < execVal.NumField(); i++ {
			field := execType.Field(i)
			tag := strings.Split(field.Tag.Get("json"), ",")[0]
			if tag == k && execVal.Field(i).CanSet() {
				fieldVal := execVal.Field(i)
				newVal := reflect.ValueOf(v)
				if newVal.Type().ConvertibleTo(fieldVal.Type()) {
					fieldVal.Set(newVal.Convert(fieldVal.Type()))
				}
			}
		}
	}

	updateQuery := GenerateUpdateQuery("execs", existingexec)
	values := GetUpdateValues(existingexec)
	values = append(values, id)

	if _, err = db.Exec(updateQuery, values...); err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "error updating data")
	}
	return existingexec, nil
}

func DeleteOneExec(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM execs WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "exec not found")
	}
	return nil
}
func GetUserByUsername(username string) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(errors.New("Cannot connect to db"), "error updating data...")

	}
	defer db.Close()

	user := models.Exec{}

	query := "SELECT id,first_name,last_name,username,email,password,role,inactive_status FROM execs WHERE username = ?"

	err = db.QueryRow(query, username).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &user.Password, &user.Role, &user.InactiveStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Exec{}, utils.ErrorHandler(err, "user not found")
		}
		return models.Exec{}, utils.ErrorHandler(errors.New("cannot scan struct values..."), "error retreiving data...")
	}
	return user, nil
}
func UpdatePasswordInDB(userID int, currentPassword string, NewPassword string) (string, error) {
	db, err := ConnectDB()
	if err != nil {
		return "", utils.ErrorHandler(err, "Failed to connect to DB...")

	}
	defer db.Close()

	var username string
	var userPassword string
	var role string

	err = db.QueryRow("SELECT username , password , role FROM execs WHERE id = ?", userID).Scan(&username, &userPassword, &role)
	if err != nil {
		return "", utils.ErrorHandler(err, "User not found...")
	}

	err = utils.VerifyPassword(currentPassword, userPassword)
	if err != nil {
		log.Println(err)
		return "", utils.ErrorHandler(err, "The password entered does not match current password...")

	}
	salt := make([]byte, 16)
	_, err = rand.Read(salt)
	if err != nil {
		return "", utils.ErrorHandler(err, "Cannot generate salt...")

	}
	hash := argon2.IDKey([]byte(NewPassword), salt, 1, 64*1024, 4, 32)
	passwordHash := base64.StdEncoding.EncodeToString(hash)
	saltHash := base64.StdEncoding.EncodeToString(salt)
	encodedHash := fmt.Sprintf("%s.%s", saltHash, passwordHash)

	currTime := time.Now().Format(time.RFC3339)

	_, err = db.Exec("UPDATE execs SET password = ? , password_changed_at = ? WHERE id = ?", encodedHash, currTime, userID)
	if err != nil {
		return "", utils.ErrorHandler(err, "Failed to update the password...")

	}
	tokenString, err := utils.SignJWT(userID, username, role)
	if err != nil {
		return "", utils.ErrorHandler(err, "Password updated... Couldn't create token")

	}
	return tokenString, nil
}
func ForgotPasswordDbHandler(emailID string) error {

	if emailID == "" {
		return utils.ErrorHandler(errors.New("User did not enter email ID"), "email ID required")
	}

	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Internal Error...")

	}
	defer db.Close()

	var exec models.Exec
	err = db.QueryRow("SELECT id from execs WHERE email = ? ", emailID).Scan(&exec.ID)
	if err != nil {
		return utils.ErrorHandler(err, "User not found...")

	}

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email...")

	}
	mins := time.Duration(duration)

	expiry := time.Now().Add(mins * time.Minute).Format(time.RFC3339)

	tokenBytes := make([]byte, 32)

	_, err = rand.Read(tokenBytes)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email...")
	}

	token := hex.EncodeToString(tokenBytes)

	hashedToken := sha256.Sum256(tokenBytes)

	hashedTokenString := hex.EncodeToString(hashedToken[:])

	_, err = db.Exec("UPDATE execs SET password_reset_token = ?  , password_token_expires = ? WHERE id = ?", hashedTokenString, expiry, exec.ID)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email...")

	}

	// Send the reset email

	resetURL := fmt.Sprintf("https://localhost:3000/execs/reset/resetpassword/%s", token)
	message := fmt.Sprintf("Forgot your password? Reset your password using the following link:\n%s\nIf you did not request password reset, please ignore this email. This link is only valid for %d minutes", resetURL, int(mins))

	m := mail.NewMessage()
	m.SetHeader("From", "schooladmin@school.com") // Sender Email
	m.SetHeader("To", emailID)                    // receiver email
	m.SetHeader("Subject", "Your password reset link")
	m.SetBody("text/plain", message)
	d := mail.NewDialer("localhost", 1025, "", "")
	// dialer returns SMTP dialer and parameters and used to connect to SMTP server
	err = d.DialAndSend(m)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email...")

	}
	return nil
}
func ResetPasswordDBHandler(token, NewPassword string) error {

	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return utils.ErrorHandler(err, "Internal error...")

	}
	hashedToken := sha256.Sum256(tokenBytes)
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "connection failed...")

	}
	defer db.Close()

	var user models.Exec

	query := "SELECT id , email FROM execs WHERE password_reset_token = ? AND password_token_expires > ? "
	err = db.QueryRow(query, hashedTokenString, time.Now().Format(time.RFC3339)).Scan(&user.ID, &user.Email)
	if err != nil {
		return utils.ErrorHandler(err, "Invalid or expired reset code")

	}
	// Hash the new password

	hashedPassword, err := utils.HashingPassword(NewPassword)
	if err != nil {
		return utils.ErrorHandler(err, "Internal error...")

	}
	updateQuery := "UPDATE execs SET password = ? , password_reset_token = NULL , password_token_expires = NULL , password_changed_at = ? WHERE id = ? "

	_, err = db.Exec(updateQuery, hashedPassword, time.Now().Format(time.RFC3339), &user.ID)
	if err != nil {
		utils.ErrorHandler(err, "Internal error...")
		return err
	}
	return nil
}
