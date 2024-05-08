package main

import (
	"backend/email"
	"backend/handler"
	"backend/logger"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func loggingMiddleware(logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info(r.RequestURI)
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	logger := logger.NewLogger()
	h := handler.NewHandler(logger)
	r := mux.NewRouter()

	e := email.NewEmailService(logger)
	err := e.GetEmails()
	if err != nil {
		logger.Error(err.Error())
	}

	r.HandleFunc("/transactions", h.GetTransactions()).Methods("GET")
	r.HandleFunc("/transactions", h.PostTransaction()).Methods("POST")
	r.Use(loggingMiddleware(logger))
	log.Fatal(http.ListenAndServe(":8080", r))
}
