package main

import (
	"backend/email"
	"backend/handler"
	"backend/logger"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
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
	e := email.NewEmailService(logger)
	go e.GetEmails()

	// b, err := backup.New(logger)
	// if err != nil {
	// 	logger.Error(fmt.Sprintf("error instantiating backup: %s ", err))
	// } else {
	// 	b.Upload("/Users/ms/coding/card-transactions/backend/database.db") // TODO
	// 	// TODO maybe as a separat script?
	// 	b.Download("/Users/ms/coding/card-transactions/backend/database.db") // TODO
	// }

	h := handler.NewHandler(logger)
	r := mux.NewRouter()

	r.HandleFunc("/transactions", h.GetTransactions(logger)).Methods("GET")
	r.HandleFunc("/transactions", h.PostTransaction(logger)).Methods("POST")
	r.Use(loggingMiddleware(logger))

	// // todo temporary solution
	corsOptions := handlers.AllowedOrigins([]string{"*"})
	corsHandler := handlers.CORS(corsOptions)
	cr := corsHandler(r)

	log.Fatal(http.ListenAndServe(":8080", cr))
}
