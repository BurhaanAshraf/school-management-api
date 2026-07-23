package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"schoolmanagementapi/internal/api/models"
	"schoolmanagementapi/internal/api/repositories/sqlconnect"
	"schoolmanagementapi/internal/pkg/utils"
	"strconv"
)

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	page, limit := GetPaginationParams(r)
	teachers, err, totalTeachersCount := sqlconnect.GetTeachersDbHandler(r, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := struct {
		Status   string           `json:"status"`
		Count    int              `json:"count"`
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
		Data     []models.Teacher `json:"data"`
	}{
		Status:   "success",
		Count:    totalTeachersCount,
		Page:     page,
		PageSize: limit,
		Data:     teachers,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Cannot Encode", http.StatusInternalServerError)
		return
	}

}

func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	ID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}
	teacher, err := sqlconnect.GetTeacherByID(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)

}

func AddTeachersHandler(w http.ResponseWriter, r *http.Request) {

	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var newTeachers []models.Teacher

	var rawTeachers []map[string]any

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body...", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &rawTeachers)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	fields := GetFieldNames(models.Teacher{})

	allowedFields := make(map[string]struct{})

	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, teacher := range rawTeachers {
		for key := range teacher {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable field found in request...", http.StatusBadRequest)
				return
			}
		}
	}
	err = json.Unmarshal(body, &newTeachers)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	for _, teacher := range newTeachers {
		err := CheckBlankFields(teacher)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	addedTeachers, err := sqlconnect.AddTeachersDBHandler(newTeachers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)
}

func UpdateTeacherHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}

	var updatedTeacher models.Teacher

	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	updatedTeacherFromDB, err := sqlconnect.UpdateTeacher(id, updatedTeacher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacherFromDB)

}

func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updates []map[string]any

	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err = sqlconnect.PatchTeachers(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func PatchOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}
	var updates map[string]any

	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}
	updatedTeacher, err := sqlconnect.PatchOneTeacher(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)

}

func DeleteOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	err = sqlconnect.DeleteOneTeacher(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Teacher successfully deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)

}

func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(utils.ContextKey("role")).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := utils.AuthorizeUser(role, "super_admin", "student_affairs", "principal", "registrar", "secretary", "vice_principal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var ids []int
	err = json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	deleteIDs, err := sqlconnect.DeleteTeachers(ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int  `json:"deleted_ids"`
	}{
		Status:     "Teachers successfully deleted",
		DeletedIDs: deleteIDs,
	}
	json.NewEncoder(w).Encode(response)

}
func GetStudentsByTeacherId(w http.ResponseWriter, r *http.Request) {
	teacherID := r.PathValue("id")

	var students []models.Student

	students, err := sqlconnect.GetStudentsByTeacherIdDBHandler(teacherID, students)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(students),
		Data:   students,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
func GetStudentCountByTeacherId(w http.ResponseWriter, r *http.Request) {

	teacherID := r.PathValue("id")
	var count int

	count, err := sqlconnect.GetStudentCountByTeacherIdDbHandler(teacherID, count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}{
		Status: "success",
		Count:  count,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
