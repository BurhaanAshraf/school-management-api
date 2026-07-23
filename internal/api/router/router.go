package router

import "net/http"

func MainRouter() *http.ServeMux {
	root := http.NewServeMux()

	tRouter := TeachersRouter()
	sRouter := StudentsRouter()

	sRouter.Handle("/", ExecsRouter())
	tRouter.Handle("/", sRouter)

	root.Handle("/", tRouter)
	root.HandleFunc("GET /{$}", HomeHandler)

	return root
}
