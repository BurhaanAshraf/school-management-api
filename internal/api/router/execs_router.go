package router

import (
	"net/http"
	"schoolmanagementapi/internal/api/handlers"
)

func ExecsRouter() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /execs", handlers.GetExecsHandler)
	mux.HandleFunc("POST /execs", handlers.AddExecsHandler)
	mux.HandleFunc("PATCH /execs", handlers.PatchExecsHandler)

	mux.HandleFunc("GET /execs/{id}", handlers.GetOneExecsHandler)
	mux.HandleFunc("PATCH /execs/{id}", handlers.PatchOneExecsHandler)
	mux.HandleFunc("DELETE /execs/{id}", handlers.DeleteOneExecsHandler)
	mux.HandleFunc("POST /execs/{id}/updatepassword", handlers.UpdatePasswordHandler)

	mux.HandleFunc("POST /execs/login", handlers.LoginHandler)
	mux.HandleFunc("POST /execs/logout", handlers.LogoutHandler)
	mux.HandleFunc("POST /execs/forgotpassword", handlers.ForgotPasswordHandler)
	mux.HandleFunc("POST /execs/reset/resetpassword/{resetcode}", handlers.ResetPasswordHandler)

	return mux
}
