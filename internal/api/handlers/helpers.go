package handlers

import (
	"errors"
	"net/http"
	"reflect"
	"schoolmanagementapi/internal/pkg/utils"
	"strconv"
	"strings"
)

func CheckBlankFields(val any) error {
	reflectVal := reflect.ValueOf(val)
	for i := 0; i < reflectVal.NumField(); i++ {
		field := reflectVal.Field(i)
		if field.Kind() == reflect.String && field.String() == "" {
			// http.Error(w, "All fields are required...", http.StatusBadRequest)
			return utils.ErrorHandler(errors.New("All fields are required..."), "All fields are required...")
		}
	}
	return nil
}

func GetFieldNames(model any) []string {
	val := reflect.TypeOf(model)
	fields := []string{}
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldToAdd := strings.Split(string(field.Tag.Get("json")), ",")[0]
		fields = append(fields, fieldToAdd)
	}
	return fields
}

func GetPaginationParams(r *http.Request) (int, int) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {

		page = 1
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {

		limit = 10
	}

	return page, limit

}
