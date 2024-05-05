package main

import (
	"backend/handler"
	"backend/logger"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	logger := logger.NewLogger()
	h := handler.NewHandler(logger)
	r := mux.NewRouter()
	r.HandleFunc("/transactions", h.GetTransactions()).Methods("GET")
	r.HandleFunc("/transactions", h.PostTransaction()).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", r))
}
