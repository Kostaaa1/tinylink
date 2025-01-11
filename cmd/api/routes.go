package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *app) Routes() *mux.Router {
	r := mux.NewRouter()

	r.MethodNotAllowedHandler = http.HandlerFunc(a.methodNotAllowedResponse)
	r.NotFoundHandler = http.HandlerFunc(a.notFoundResponse)

	r.Use(a.persistSessionMW)
	r.HandleFunc("/", a.Index).Methods("GET")
	r.HandleFunc("/getAll", a.GetAll).Methods("GET")
	r.HandleFunc("/create", a.Create).Methods("POST")
	r.HandleFunc("/{tinylink}", a.Redirect).Methods("GET")
	r.HandleFunc("/{tinylink}", a.DeleteTinylink).Methods("DELETE")

	return r
}
