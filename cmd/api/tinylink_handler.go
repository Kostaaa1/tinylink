package main

import "github.com/gorilla/mux"

type TinylinkController struct {
	service TinylinkService
}

func NewTinylinkController(r *mux.Router, tinylinkService TinylinkService) {
	r.HandleFunc("/getAll", tinylinkService.List).Methods("GET")
	r.HandleFunc("/create", tinylinkService.Create).Methods("POST")
	r.HandleFunc("/{alias}", tinylinkService.Get).Methods("GET")
	r.HandleFunc("/{alias}", tinylinkService.Delete).Methods("DELETE")
}
