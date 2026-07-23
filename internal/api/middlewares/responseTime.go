package middlewares

import (
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	start  time.Time
	wrote  bool
}

func (rw *responseWriter) WriteHeader(code int) {

	if !rw.wrote {
		duration := time.Since(rw.start)
		rw.Header().Set("X-Response-Time", duration.String())

		rw.status = code
		rw.wrote = true
	}

	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {

	if !rw.wrote {
		rw.WriteHeader(http.StatusOK)
	}

	return rw.ResponseWriter.Write(b)
}

func ResponseTimeMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
			start:          time.Now(),
		}

		next.ServeHTTP(rw, r)

		_ = time.Since(rw.start)

		// fmt.Printf(
		// 	"%s %s | Status: %d | Duration: %v\n",
		// 	r.Method,
		// 	r.URL.Path,
		// 	rw.status,
		// 	duration,
		// )
	})
}
