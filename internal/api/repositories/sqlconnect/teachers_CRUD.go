package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"schoolmanagementapi/internal/api/models"
	"schoolmanagementapi/internal/pkg/utils"
	"strconv"
	"strings"
)

func GetTeachersDbHandler(r *http.Request, page, limit int) ([]models.Teacher, error, int) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retreiving data"), 0
	}
	defer db.Close()

	query := GenerateSelectQuery("teachers", models.Teacher{}, "WHERE 1=1")

	var args []any
	var teachers []models.Teacher

	query, args = AddFilters(r, query, args)
	query = AddSorting(r, query)
	offset := (page - 1) * limit
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retreiving data"), 0
	}
	defer rows.Close()

	for rows.Next() {
		var teacher models.Teacher
		if err := ScanRows(rows, &teacher); err != nil {
			return nil, utils.ErrorHandler(err, "error retreiving data"), 0
		}
		teachers = append(teachers, teacher)
	}
	var totalTeachersCount int
	query = GenerateCountQuery("teachers", "WHERE 1=1")
	err = db.QueryRow(query).Scan(&totalTeachersCount)
	if err != nil {
		return []models.Teacher{}, utils.ErrorHandler(err, "error retreiving data"), 0
	}

	return teachers, nil, totalTeachersCount
}

func GetTeacherByID(ID int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "error retreiving data")
	}
	defer db.Close()

	var teacher models.Teacher
	query := GenerateSelectQuery("teachers", models.Teacher{}, "WHERE id = ?")
	row := db.QueryRow(query, ID)

	err = ScanStruct(row, &teacher)
	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.ErrorHandler(err, "error retreiving data")
	} else if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "error retreiving data")
	}
	return teacher, nil
}

func AddTeachersDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer db.Close()

	stmt, err := db.Prepare(GenerateInsertQuery("teachers", models.Teacher{}))

	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))

	for i, newTeacher := range newTeachers {
		values := GetStructValues(newTeacher)
		res, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error adding data")
		}
		lastId, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "error adding data")
		}
		newTeacher.ID = int(lastId)
		addedTeachers[i] = newTeacher
	}
	return addedTeachers, nil
}

func UpdateTeacher(id int, updatedTeacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	var existingTeacher models.Teacher
	query := GenerateSelectQuery("teachers", models.Teacher{}, "WHERE id = ?")
	row := db.QueryRow(query, id)

	if err := ScanStruct(row, &existingTeacher); err != nil {
		if err == sql.ErrNoRows {
			return models.Teacher{}, utils.ErrorHandler(err, "error updating data")
		}
		return models.Teacher{}, utils.ErrorHandler(err, "error updating data")
	}

	updatedTeacher.ID = existingTeacher.ID

	updateQuery := GenerateUpdateQuery("teachers", updatedTeacher)
	values := GetUpdateValues(updatedTeacher)
	values = append(values, id)

	_, err = db.Exec(updateQuery, values...)
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "error updating data")
	}
	return updatedTeacher, nil
}

func PatchTeachers(updates []map[string]any) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}

	selectQuery := GenerateSelectQuery("teachers", models.Teacher{}, "WHERE id = ?")

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

		var teacherFromDB models.Teacher
		row := tx.QueryRow(selectQuery, id)
		if err := ScanStruct(row, &teacherFromDB); err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return utils.ErrorHandler(err, "Teacher not found")
			}
			return utils.ErrorHandler(err, "error updating data")
		}

		// Apply updates using reflection
		teacherVal := reflect.ValueOf(&teacherFromDB).Elem()
		teacherType := teacherVal.Type()

		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				if strings.Split(field.Tag.Get("json"), ",")[0] == k {
					fieldVal := teacherVal.Field(i)
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

		updateQuery := GenerateUpdateQuery("teachers", teacherFromDB)
		values := GetUpdateValues(teacherFromDB)
		values = append(values, teacherFromDB.ID)

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

func PatchOneTeacher(id int, updates map[string]any) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	var existingTeacher models.Teacher
	query := GenerateSelectQuery("teachers", models.Teacher{}, "WHERE id = ?")
	row := db.QueryRow(query, id)

	if err := ScanStruct(row, &existingTeacher); err != nil {
		if err == sql.ErrNoRows {
			return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")
		}
		return models.Teacher{}, utils.ErrorHandler(err, "error updating data")
	}

	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()

	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherType.Field(i)
			tag := strings.Split(field.Tag.Get("json"), ",")[0]
			if tag == k && teacherVal.Field(i).CanSet() {
				fieldVal := teacherVal.Field(i)
				newVal := reflect.ValueOf(v)
				if newVal.Type().ConvertibleTo(fieldVal.Type()) {
					fieldVal.Set(newVal.Convert(fieldVal.Type()))
				}
			}
		}
	}

	updateQuery := GenerateUpdateQuery("teachers", existingTeacher)
	values := GetUpdateValues(existingTeacher)
	values = append(values, id)

	if _, err = db.Exec(updateQuery, values...); err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "error updating data")
	}
	return existingTeacher, nil
}

func DeleteOneTeacher(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "Teacher not found")
	}
	return nil
}

func DeleteTeachers(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}

	stmt, err := tx.Prepare("DELETE FROM teachers where id = ?")
	if err != nil {
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "error deleting data")
	}
	defer stmt.Close()

	deleteIDs := []int{}

	for _, id := range ids {
		res, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "error deleting data")
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "error deleting data")
		}
		if rowsAffected == 0 {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, fmt.Sprintf("ID %v does not exist", id))
		}
		deleteIDs = append(deleteIDs, id)
	}

	if err = tx.Commit(); err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}

	if len(deleteIDs) < 1 {
		return nil, utils.ErrorHandler(err, "IDs do not exist")
	}
	return deleteIDs, nil
}
func GetStudentsByTeacherIdDBHandler(teacherID string, students []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {

		return nil, utils.ErrorHandler(err, "error retreiving data")
	}
	defer db.Close()

	query := GenerateSelectQuery("students", models.Student{}, "WHERE class = (SELECT class from teachers where id = ?)")
	rows, err := db.Query(query, teacherID)
	if err != nil {

		return nil, utils.ErrorHandler(err, "error retreiving data")
	}

	defer rows.Close()

	for rows.Next() {
		var student models.Student
		err := ScanRows(rows, &student)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error retreiving data")
		}
		students = append(students, student)
	}
	err = rows.Err()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retreiving data")
	}
	return students, nil
}
func GetStudentCountByTeacherIdDbHandler(teacherID string, count int) (int, error) {
	db, err := ConnectDB()
	if err != nil {
		return 0, utils.ErrorHandler(err, "error retreiving data")

	}
	defer db.Close()

	query := GenerateCountQuery("students", "WHERE class = (SELECT class FROM teachers WHERE id = ?)")

	err = db.QueryRow(query, teacherID).Scan(&count)

	if err != nil {
		return 0, utils.ErrorHandler(err, "error retreiving data")
	}
	return count, nil
}
