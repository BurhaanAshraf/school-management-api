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

func GetStudentsDbHandler(r *http.Request, limit, page int) ([]models.Student, error, int) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retreiving data"), 0
	}
	defer db.Close()
	var students []models.Student
	query := GenerateSelectQuery("students", models.Student{}, "WHERE 1=1")
	var args []any
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
		var student models.Student
		if err := ScanRows(rows, &student); err != nil {
			return nil, utils.ErrorHandler(err, "error retreiving data"), 0
		}
		students = append(students, student)
	}
	// Getting total count
	var totalStudentsCount int
	query = GenerateCountQuery("students", "WHERE 1=1")
	err = db.QueryRow(query).Scan(&totalStudentsCount)
	if err != nil {
		utils.ErrorHandler(err, "")
		totalStudentsCount = 0
	}
	return students, nil, totalStudentsCount
}

func GetStudentByID(ID int) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "error retreiving data")
	}
	defer db.Close()

	var student models.Student
	query := GenerateSelectQuery("students", models.Student{}, "WHERE id = ?")
	row := db.QueryRow(query, ID)

	err = ScanStruct(row, &student)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, "error retreiving data")
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "error retreiving data")
	}
	return student, nil
}

func AddStudentsDBHandler(newStudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer db.Close()

	stmt, err := db.Prepare(GenerateInsertQuery("students", models.Student{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer stmt.Close()

	addedstudents := make([]models.Student, len(newStudents))

	for i, newStudent := range newStudents {
		values := GetStructValues(newStudent)
		res, err := stmt.Exec(values...)
		if err != nil {
			if strings.Contains(err.Error(), "a foreign key constraint fails (`school`.`students`, CONSTRAINT `students_ibfk_1` FOREIGN KEY (`class`) REFERENCES `teachers` (`class`))") {
				return nil, utils.ErrorHandler(err, "class / class-teacher does not exist")
			}
			return nil, utils.ErrorHandler(err, "error adding data")
		}
		lastId, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "error adding data")
		}
		newStudent.ID = int(lastId)
		addedstudents[i] = newStudent
	}
	return addedstudents, nil
}

func UpdateStudent(id int, updatedstudent models.Student) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	var existingstudent models.Student
	query := GenerateSelectQuery("students", models.Student{}, "WHERE id = ?")
	row := db.QueryRow(query, id)

	if err := ScanStruct(row, &existingstudent); err != nil {
		if err == sql.ErrNoRows {
			log.Println(err)
			return models.Student{}, utils.ErrorHandler(err, "error updating data")
		}
		log.Println(err)
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}

	updatedstudent.ID = existingstudent.ID

	updateQuery := GenerateUpdateQuery("students", updatedstudent)
	values := GetUpdateValues(updatedstudent)
	values = append(values, id)

	_, err = db.Exec(updateQuery, values...)
	if err != nil {
		log.Println(err)
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}
	return updatedstudent, nil
}

func PatchStudents(updates []map[string]any) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}

	selectQuery := GenerateSelectQuery("students", models.Student{}, "WHERE id = ?")

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

		var studentFromDB models.Student
		row := tx.QueryRow(selectQuery, id)
		if err := ScanStruct(row, &studentFromDB); err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return utils.ErrorHandler(err, "student not found")
			}
			return utils.ErrorHandler(err, "error updating data")
		}

		// Apply updates using reflection
		studentVal := reflect.ValueOf(&studentFromDB).Elem()
		studentType := studentVal.Type()

		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < studentVal.NumField(); i++ {
				field := studentType.Field(i)
				if strings.Split(field.Tag.Get("json"), ",")[0] == k {
					fieldVal := studentVal.Field(i)
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

		updateQuery := GenerateUpdateQuery("students", studentFromDB)
		values := GetUpdateValues(studentFromDB)
		values = append(values, studentFromDB.ID)

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

func PatchOneStudent(id int, updates map[string]any) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	var existingstudent models.Student
	query := GenerateSelectQuery("students", models.Student{}, "WHERE id = ?")
	row := db.QueryRow(query, id)

	if err := ScanStruct(row, &existingstudent); err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, utils.ErrorHandler(err, "student not found")
		}
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}

	studentVal := reflect.ValueOf(&existingstudent).Elem()
	studentType := studentVal.Type()

	for k, v := range updates {
		for i := 0; i < studentVal.NumField(); i++ {
			field := studentType.Field(i)
			tag := strings.Split(field.Tag.Get("json"), ",")[0]
			if tag == k && studentVal.Field(i).CanSet() {
				fieldVal := studentVal.Field(i)
				newVal := reflect.ValueOf(v)
				if newVal.Type().ConvertibleTo(fieldVal.Type()) {
					fieldVal.Set(newVal.Convert(fieldVal.Type()))
				}
			}
		}
	}

	updateQuery := GenerateUpdateQuery("students", existingstudent)
	values := GetUpdateValues(existingstudent)
	values = append(values, id)

	if _, err = db.Exec(updateQuery, values...); err != nil {
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}
	return existingstudent, nil
}

func DeleteOneStudent(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "student not found")
	}
	return nil
}

func DeleteStudents(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}

	stmt, err := tx.Prepare("DELETE FROM students where id = ?")
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
