package sqlconnect

import (
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

func AddSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

func AddFilters(r *http.Request, query string, args []any) (string, []any) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
		"subject":    "subject",
	}
	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}
	return query, args
}

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}
	return validFields[field]
}

func GenerateSelectQuery(tableName string, model any, where string) string {
	var columns string
	reflectType := reflect.TypeOf(model)

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		dbTag := field.Tag.Get("db")
		tag := strings.Split(dbTag, ",")[0]
		if tag != "" {
			if columns != "" {
				columns += ", "
			}
			columns += tag
		}
	}
	query := fmt.Sprintf("SELECT %s FROM %s", columns, tableName)

	if where != "" {
		query = query + " " + where
	}
	return query
}

func GenerateCountQuery(tableName, where string) string {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)

	if where != "" {
		query += " " + where
	}

	return query
}

func GenerateInsertQuery(tableName string, model any) string {
	modelType := reflect.TypeOf(model)
	var columns, placeholders string
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		tag := strings.Split(dbTag, ",")[0]
		if tag != "" && tag != "id" {

			if columns != "" {
				columns += ", "
				placeholders += ", "
			}
			columns += tag
			placeholders += "?"
		}

	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columns, placeholders)
}

func GetStructValues(model any) []any {
	modelVal := reflect.ValueOf(model)
	modelType := modelVal.Type()

	values := []interface{}{}
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		dbTag := field.Tag.Get("db")
		tag := strings.Split(dbTag, ",")[0]
		if tag != "" && tag != "id" {
			values = append(values, modelVal.Field(i).Interface())
		}
	}
	return values
}

func GenerateUpdateQuery(tableName string, model any) string {
	var columns string
	reflectType := reflect.TypeOf(model)
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		dbTag := field.Tag.Get("db")
		tag := strings.Split(dbTag, ",")[0]
		if tag != "" && tag != "id" {
			if columns != "" {
				columns += ", "
			}
			columns += tag
			columns += " = ?"
		}
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE id = ? ", tableName, columns)
}

func GetUpdateValues(model any) []any {
	modelVal := reflect.ValueOf(model)
	modelType := modelVal.Type()

	values := []interface{}{}
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		dbTag := field.Tag.Get("db")
		tag := strings.Split(dbTag, ",")[0]
		if tag != "" && tag != "id" {
			values = append(values, modelVal.Field(i).Interface())
		}
	}
	return values
}

func ScanStruct(row *sql.Row, model any) error {
	modelVal := reflect.ValueOf(model).Elem()
	modelType := modelVal.Type()

	values := []interface{}{}
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		tag := strings.Split(dbTag, ",")[0]
		if tag != "" {
			values = append(values, modelVal.Field(i).Addr().Interface())
		}
	}
	return row.Scan(values...)
}

func ScanRows(rows *sql.Rows, model any) error {
	modelVal := reflect.ValueOf(model).Elem()
	modelType := modelVal.Type()

	values := []interface{}{}
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		tag := strings.Split(dbTag, ",")[0]
		if tag != "" {
			values = append(values, modelVal.Field(i).Addr().Interface())
		}
	}
	return rows.Scan(values...)
}
