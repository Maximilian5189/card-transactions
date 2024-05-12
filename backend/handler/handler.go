package handler

import (
	"backend/db"
	"backend/logger"
	"encoding/json"
	"net/http"
)

type Handler struct {
	logger logger.Logger
}

func NewHandler(logger logger.Logger) Handler {
	return Handler{logger}
}

func (handler *Handler) handleErr(err error, w http.ResponseWriter) error {
	if err != nil {
		handler.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	return err
}

func (handler *Handler) GetTransactions(logger logger.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		d, err := db.NewTransactionsDB(logger)
		if handler.handleErr(err, w) != nil {
			return
		}

		transactions, err := d.Select()
		if handler.handleErr(err, w) != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	}

}

func (handler *Handler) PostTransaction(logger logger.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		d, err := db.NewTransactionsDB(logger)
		if handler.handleErr(err, w) != nil {
			return
		}

		transaction := db.Transaction{Name: "Jörg"} // TODO lol
		err = d.Insert(transaction)
		if handler.handleErr(err, w) != nil {
			return
		}
	}
}
