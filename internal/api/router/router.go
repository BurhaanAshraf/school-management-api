package router

import (
	"net/http"
)

func MainRouter() *http.ServeMux {

	tRouter := TeachersRouter()
	sRouter := StudentsRouter()
	eRouter := ExecsRouter()

	sRouter.Handle("/", eRouter)
	tRouter.Handle("/", sRouter)
	return tRouter

}
