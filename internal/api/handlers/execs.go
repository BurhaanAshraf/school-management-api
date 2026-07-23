package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"schoolmanagementapi/internal/api/models"
	"schoolmanagementapi/internal/api/repositories/sqlconnect"
	"schoolmanagementapi/internal/pkg/utils"
	"strconv"
	"time"
)

func GetExecsHandler(w http.ResponseWriter, r *http.Request) {

	page, limit := GetPaginationParams(r)
	execs, err, totalExecsCount := sqlconnect.GetExecsDbHandler(r, page, limit)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := struct {
		Status   string                `json:"status"`
		Count    int                   `json:"count"`
		Page     int                   `json:"page"`
		PageSize int                   `json:"page_size"`
		Data     []models.ExecResponse `json:"data"`
	}{
		Status:   "success",
		Count:    totalExecsCount,
		Page:     page,
		PageSize: limit,
		Data:     execs,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Cannot Encode", http.StatusInternalServerError)
		return
	}
}

func GetOneExecsHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	ID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Exec ID", http.StatusBadRequest)
		return
	}
	Exec, err := sqlconnect.GetExecByID(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Exec)

}

func AddExecsHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	var newExecs []models.Exec

	var rawExecs []map[string]any

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body...", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &rawExecs)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	fields := GetFieldNames(models.Exec{})

	allowedFields := make(map[string]struct{})

	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, Exec := range rawExecs {
		for key := range Exec {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable field found in request...", http.StatusBadRequest)
				return
			}
		}
	}
	err = json.Unmarshal(body, &newExecs)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	for _, Exec := range newExecs {
		err := CheckBlankFields(Exec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	addedExecs, err := sqlconnect.AddExecsDBHandler(newExecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}
	json.NewEncoder(w).Encode(response)
}

func PatchExecsHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	var updates []map[string]any

	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err = sqlconnect.PatchExecs(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func PatchOneExecsHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Exec ID", http.StatusBadRequest)
		return
	}
	var updates map[string]any

	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}
	updatedExec, err := sqlconnect.PatchOneExec(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedExec)

}

func DeleteOneExecsHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	err = sqlconnect.DeleteOneExec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Exec successfully deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	// Data Validation

	var req models.Exec

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// Search for user if user actually exists

	user, err := sqlconnect.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "invalid username or password...", http.StatusBadRequest)
		return
	}

	// is user active

	if user.InactiveStatus {
		http.Error(w, "account is inactive", http.StatusForbidden)
		return
	}

	// verify password

	err = utils.VerifyPassword(req.Password, user.Password)
	if err != nil {
		http.Error(w, "invalid username or password...", http.StatusForbidden)
		return
	}

	// generate token

	tokenString, err := utils.SignJWT(user.ID, user.Username, user.Role)

	if err != nil {
		http.Error(w, "Could not create login token", http.StatusInternalServerError)
		return
	}
	// send token as a response or set as a cookie

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})
	response := struct {
		Status string `json:"status"`
	}{
		Status: "logged out successfully...",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Exec ID", http.StatusBadRequest)
		return
	}
	request := models.UpdatePasswordRequest{}

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	if request.CurrentPassword == "" || request.NewPassword == "" {
		http.Error(w, "Please enter password", http.StatusBadRequest)
		return
	}

	tokenString, err := sqlconnect.UpdatePasswordInDB(userID, request.CurrentPassword, request.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	response := struct {
		Status string `json:"status"`
		Token  string `json:"token"`
	}{
		Token:  tokenString,
		Status: "password has been updated successfully",
	}
	json.NewEncoder(w).Encode(response)
	w.Header().Set("Content-Type", "application/json")

}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email" db:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body...", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = sqlconnect.ForgotPasswordDbHandler(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Password reset link send to %s", req.Email)

}

type ResetPasswordRequest struct {
	NewPassword     string `json:"new_password" db:"new_password"`
	ConfirmPassword string `json:"confirm_password" db:"confirm_password"`
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("resetcode")

	var req ResetPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if req.NewPassword == "" || req.ConfirmPassword == "" {
		http.Error(w, "Password field cannot be empty...", http.StatusUnauthorized)
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		http.Error(w, "Passwords do not match...", http.StatusBadRequest)
		return
	}

	err = sqlconnect.ResetPasswordDBHandler(token, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}

	fmt.Fprintln(w, "Password reset successfully...")

}
