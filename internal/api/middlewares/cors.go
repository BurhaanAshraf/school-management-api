package middlewares

import (
	"net/http"
)

// api is hosted at www.myapi.com
// frontend is also hosted on server
// so frontend server is at www.myfrontend.com

// Allowed origins

var allowedOrigins = []string{
	"https://my-origin-url.com",
	"https://www.myfrontend.com",
	"https://localhost:3000",
}

func Cors(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" {
			if !isOriginAllowed(origin) {
				http.Error(w, "Not allowed by CORS", http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		// w.Header().Set()
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)

	})
}

func isOriginAllowed(origin string) bool {
	for _, allowedorigin := range allowedOrigins {
		if origin == allowedorigin {
			return true
		}
	}
	return false
}
