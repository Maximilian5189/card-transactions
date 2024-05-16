package main

import (
	"backend/email"
	"backend/handler"
	"backend/logger"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

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

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the username and password from the URL parameters
		t := r.URL.Query().Get("t")

		if t != os.Getenv("TOKEN") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := logger.NewLogger()
	e := email.NewEmailService(logger)

	go e.GetEmails()

	ticker := time.NewTicker(15 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				e.GetEmails()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

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
	r.Use(AuthMiddleware)

	// // todo add fly (or maybe not needed?)
	// corsOptions := handlers.AllowedOrigins([]string{"http://localhost:"})
	// corsHandler := handlers.CORS(corsOptions)
	// cr := corsHandler(r)

	isLocalhost := func(origin string) bool {
		return strings.HasPrefix(origin, "http://localhost")
	}

	cr := handlers.CORS(
		handlers.AllowedOriginValidator(isLocalhost),
	)(r)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: cr,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	go func() {
		<-stopChan
		logger.Info("Received stop signal, gracefully shutting down...")

		// Create a context with a timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Shut down the server gracefully
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("Error during shutdown: %v", err))
		}
	}()

	logger.Info("Starting server on :8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe failed: %v", err)
	}

	logger.Info("Server stopped gracefully")
}
