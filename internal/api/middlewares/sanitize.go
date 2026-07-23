package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"schoolmanagementapi/internal/pkg/utils"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func XSSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// -----------------------------
		// Sanitize URL Path
		// -----------------------------

		sanitizedPath, err := clean(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// -----------------------------
		// Sanitize Query Parameters
		// -----------------------------

		query := r.URL.Query()
		sanitizedQuery := make(url.Values)

		for key, values := range query {

			sanitizedKey, err := clean(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var sanitizedValues []string

			for _, value := range values {

				sanitizedValue, err := clean(value)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				sanitizedStr := sanitizedValue.(string)

				sanitizedValues = append(sanitizedValues, sanitizedStr)

				sanitizedQuery.Add(
					sanitizedKey.(string),
					sanitizedStr,
				)
			}
		}

		r.URL.Path = sanitizedPath.(string)
		r.URL.RawQuery = sanitizedQuery.Encode()

		// -----------------------------
		// Skip body processing if empty
		// -----------------------------

		if r.Body == nil || r.ContentLength == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// -----------------------------
		// Validate Content-Type
		// -----------------------------

		ct := r.Header.Get("Content-Type")

		if ct != "" && !strings.HasPrefix(ct, "application/json") {
			http.Error(
				w,
				"Unsupported content-type. Please use application/json",
				http.StatusUnsupportedMediaType,
			)
			return
		}

		// -----------------------------
		// Read Body
		// -----------------------------

		bodyBytes, err := io.ReadAll(r.Body)
		r.Body.Close()

		if err != nil {
			http.Error(
				w,
				utils.ErrorHandler(err, "error reading request body").Error(),
				http.StatusBadRequest,
			)
			return
		}

		bodyBytes = bytes.TrimSpace(bodyBytes)

		if len(bodyBytes) == 0 {
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			next.ServeHTTP(w, r)
			return
		}

		// -----------------------------
		// Decode JSON
		// -----------------------------

		var input any

		if err := json.Unmarshal(bodyBytes, &input); err != nil {
			http.Error(
				w,
				utils.ErrorHandler(err, "Invalid JSON Body").Error(),
				http.StatusBadRequest,
			)
			return
		}

		// -----------------------------
		// Sanitize JSON
		// -----------------------------

		sanitizedData, err := clean(input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// -----------------------------
		// Encode back into request body
		// -----------------------------

		sanitizedBody, err := json.Marshal(sanitizedData)
		if err != nil {
			http.Error(
				w,
				utils.ErrorHandler(err, "Error sanitizing body").Error(),
				http.StatusBadRequest,
			)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(sanitizedBody))
		r.ContentLength = int64(len(sanitizedBody))

		next.ServeHTTP(w, r)
	})
}

// Clean sanitizes input data to prevent XSS attacks
func clean(data any) (any, error) {
	switch d := data.(type) {
	case map[string]any:
		for k, v := range d {
			d[k] = sanitizeValue(v)
		}
		return d, nil
	case []any:
		for i, v := range d {
			d[i] = sanitizeValue(v)
		}
		return d, nil
	case string:
		return sanitizeString(d), nil
	default:
		return nil, utils.ErrorHandler(fmt.Errorf("Unsupported Type %T", data), fmt.Sprintf("Unsupported Type: %T", data))
	}
}

func sanitizeValue(data any) any {

	switch d := data.(type) {
	case string:
		return sanitizeString(d)

	case map[string]any:
		for k, v := range d {
			d[k] = sanitizeValue(v)
		}
		return d

	case []any:
		for i, v := range d {
			d[i] = sanitizeValue(v)
		}
		return d
	default:
		return d
	}
}

func sanitizeString(value string) string {
	return bluemonday.UGCPolicy().Sanitize(value)
}
