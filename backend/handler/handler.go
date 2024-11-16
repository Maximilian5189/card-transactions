package handler

import (
	"backend/db"
	"backend/logger"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
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

		var from time.Time
		if r.URL.Query().Get("from") != "" {
			f := r.URL.Query().Get("from")
			fromInt64, err := strconv.ParseInt(f, 10, 64)
			if handler.handleErr(err, w) != nil {
				return
			}
			from = time.Unix(fromInt64/1000, 0)
		} else {
			now := time.Now()
			offset := int(time.Monday - now.Weekday())
			if offset > 0 {
				offset -= 7 // Adjust for the week starting on Monday
			}
			from = now.AddDate(0, 0, offset)
			from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		}

		to := from.AddDate(0, 0, 6)
		to = time.Date(to.Year(), to.Month(), to.Day(), 24, 0, 0, 0, to.Location())
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

		body, err := io.ReadAll(r.Body)
		if handler.handleErr(err, w) != nil {
			return
		}
		defer r.Body.Close()

		var transaction db.Transaction
		err = json.Unmarshal(body, &transaction)
		if handler.handleErr(err, w) != nil {
			return
		}
		transaction.Date /= 1000

		err = d.Insert(transaction)
		if handler.handleErr(err, w) != nil {
			return
		}
	}
}

func (handler *Handler) DeleteTransaction(logger logger.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "DELETE")
		d, err := db.NewTransactionsDB(logger)
		if handler.handleErr(err, w) != nil {
			return
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusInternalServerError)
		}

		i, err := strconv.Atoi(id)
		if handler.handleErr(err, w) != nil {
			return
		}

		err = d.DeleteByID(i)
		handler.handleErr(err, w)
	}
}
