package middlewares

import (
	"net/http"
	"strings"
)

type HPPOptions struct {
	CheckQuery                  bool
	CheckBody                   bool
	CheckBodyOnlyForContentType string
	Whitelist                   []string
}

func HppMiddleware(options HPPOptions) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if options.CheckBody && r.Method == http.MethodPost && isCorrectContentType(r, options.CheckBodyOnlyForContentType) {
				// filtering the body params
				filteringBodyParams(r, options.Whitelist)
			}
			if options.CheckQuery && r.URL.Query() != nil {
				// filtering the query params
				filteringQueryParams(r, options.Whitelist)
			}

			next.ServeHTTP(w, r)

		})
	}
}

func isCorrectContentType(r *http.Request, contentType string) bool {

	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func filteringBodyParams(r *http.Request, Whitelist []string) {
	err := r.ParseForm()
	if err != nil {
		return
	}
	for k, v := range r.Form {
		if len(v) > 1 {
			r.Form.Set(k, v[0])
			// r.Form.Set(k , v[len(v) - 1]) --> Last Value
		}
		if !isWhitelisted(k, Whitelist) {
			delete(r.Form, k)
		}
	}
}
func filteringQueryParams(r *http.Request, Whitelist []string) {
	query := r.URL.Query()

	for k, v := range query {
		if len(v) > 1 {
			query.Set(k, v[0])
			//query.Set(k , v[len(v) - 1]) --> Last Value
		}
		if !isWhitelisted(k, Whitelist) {
			query.Del(k)
		}
	}
	r.URL.RawQuery = query.Encode()
}
func isWhitelisted(param string, Whitelist []string) bool {
	for _, v := range Whitelist {
		if param == v {
			return true
		}
	}
	return false
}
