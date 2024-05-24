package main

import (
	"backend/backup"
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
	if os.Getenv("TOKEN") == "" {
		log.Fatal("missing token")
	}

	logger := logger.NewLogger()
	e, err := email.NewEmailService(logger)

	if err == nil {

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
	}

	dir, err := os.Getwd()
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting current directory: %s", err))
	} else {
		b, err := backup.New(logger)
		if err != nil {
			logger.Error(fmt.Sprintf("error instantiating backup: %s ", err))
		} else {
			ticker := time.NewTicker(24 * time.Hour)
			quit := make(chan struct{})
			go func() {
				for {
					select {
					case <-ticker.C:
						b.Upload(fmt.Sprintf("%s/database.db", dir))
					case <-quit:
						ticker.Stop()
						return
					}
				}
			}()
			// TODO as a separat script
			// b.Download("/Users/ms/coding/card-transactions/backend/database.db")
		}
	}

	h := handler.NewHandler(logger)
	r := mux.NewRouter()

	r.HandleFunc("/transactions", h.GetTransactions(logger)).Methods("GET")
	r.HandleFunc("/transaction", h.PostTransaction(logger)).Methods("POST")
	r.HandleFunc("/transaction", h.DeleteTransaction(logger)).Methods("DELETE")
	r.Use(loggingMiddleware(logger))
	r.Use(AuthMiddleware)

	isLocalhost := func(origin string) bool {
		return strings.HasPrefix(origin, "http://localhost") ||
			strings.HasPrefix(origin, "https://card-transactions-frontend.fly.dev")
	}

	cr := handlers.CORS(
		handlers.AllowedOriginValidator(isLocalhost),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
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
