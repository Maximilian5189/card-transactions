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
	db     *db.TransactionsDB
}

func NewHandler(logger logger.Logger, database *db.TransactionsDB) Handler {
	return Handler{logger, database}
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

		from = from.UTC()

		hour, min := from.Hour(), from.Minute()

		to := from.AddDate(0, 0, 7)
		to = time.Date(to.Year(), to.Month(), to.Day(), hour, min, 0, 0, to.Location())
		transactions, err := handler.db.Select(from.Unix(), to.Unix())
		if handler.handleErr(err, w) != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	}

}

func (handler *Handler) PostTransaction(logger logger.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		err = handler.db.Insert(transaction)
		if handler.handleErr(err, w) != nil {
			return
		}
	}
}

func (handler *Handler) DeleteTransaction(logger logger.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "DELETE")

		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusInternalServerError)
		}

		i, err := strconv.Atoi(id)
		if handler.handleErr(err, w) != nil {
			return
		}

		err = handler.db.DeleteByID(i)
		handler.handleErr(err, w)
	}
}
