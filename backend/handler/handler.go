package handler

import (
	"backend/db"
	"backend/logger"
	"encoding/json"
	"net/http"
	"time"
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

		// TODO maybe move from/to logic to frontend?
		now := time.Now()
		offset := int(time.Monday - now.Weekday())
		if offset > 0 {
			offset -= 7 // Adjust for the week starting on Monday
		}
		from := now.AddDate(0, 0, offset)
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

		to := from.AddDate(0, 0, 6)
		to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, to.Location())
		transactions, err := d.Select(from.Unix(), to.Unix())
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
