package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"schoolmanagementapi/internal/api/models"
	"schoolmanagementapi/internal/api/repositories/sqlconnect"
	"schoolmanagementapi/internal/pkg/utils"
	"strconv"
)

func GetStudentsHandler(w http.ResponseWriter, r *http.Request) {

	page, limit := GetPaginationParams(r)

	students, err, totalStudentsCount := sqlconnect.GetStudentsDbHandler(r, limit, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// url?limit=50&page=1
	// database will not show calculated entries from the beginning , page - 1 * limit for page 1 it is 0
	// for page 2 it will be 50 which means it will ignore first 50 entries

	response := struct {
		Status   string           `json:"status"`
		Count    int              `json:"count"`
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
		Data     []models.Student `json:"data"`
	}{
		Status:   "success",
		Count:    totalStudentsCount,
		Page:     page,
		PageSize: limit,
		Data:     students,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Cannot Encode", http.StatusInternalServerError)
		return
	}

}

func GetOneStudentHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	ID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		return
	}
	student, err := sqlconnect.GetStudentByID(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)

}

func AddStudentsHandler(w http.ResponseWriter, r *http.Request) {
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

	var newStudents []models.Student

	var rawStudents []map[string]any

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body...", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &rawStudents)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	fields := GetFieldNames(models.Student{})

	allowedFields := make(map[string]struct{})

	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, student := range rawStudents {
		for key := range student {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable field found in request...", http.StatusBadRequest)
				return
			}
		}
	}
	err = json.Unmarshal(body, &newStudents)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	for _, student := range newStudents {
		err := CheckBlankFields(student)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	addedStudents, err := sqlconnect.AddStudentsDBHandler(newStudents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	json.NewEncoder(w).Encode(response)
}

func UpdateStudentHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		return
	}

	var updatedStudent models.Student

	err = json.NewDecoder(r.Body).Decode(&updatedStudent)
	if err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		log.Println(err)
		return
	}

	updatedStudentFromDB, err := sqlconnect.UpdateStudent(id, updatedStudent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStudentFromDB)

}

func PatchStudentsHandler(w http.ResponseWriter, r *http.Request) {
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
	err = sqlconnect.PatchStudents(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func PatchOneStudentHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		return
	}
	var updates map[string]any

	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}
	updatedStudent, err := sqlconnect.PatchOneStudent(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStudent)

}

func DeleteOneStudentHandler(w http.ResponseWriter, r *http.Request) {
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
	err = sqlconnect.DeleteOneStudent(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Student successfully deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)

}

func DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
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

	var ids []int
	err = json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	deleteIDs, err := sqlconnect.DeleteStudents(ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int  `json:"deleted_ids"`
	}{
		Status:     "Students successfully deleted",
		DeletedIDs: deleteIDs,
	}
	json.NewEncoder(w).Encode(response)

}
