package db

import (
	"backend/logger"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Transaction struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	MessageID string `json:"messageID"`
	Date      int64  `json:"date"`
	Amount    string `json:"amount"`
}

type TransactionsDB struct {
	logger logger.Logger
	db     *sql.DB
}

func NewTransactionsDB(logger logger.Logger) (*TransactionsDB, error) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS transactions " +
			"(id INTEGER PRIMARY KEY, name TEXT, messageid TEXT UNIQUE, date INTEGER, amount TEXT, owner TEXT)")
	if err != nil {
		return nil, err
	}
	return &TransactionsDB{logger, db}, nil
}

func (t *TransactionsDB) Insert(transaction Transaction) error {
	_, err := t.db.Exec(
		"INSERT INTO transactions (name, messageid, date, amount) VALUES(?,?,?,?);",
		transaction.Name, transaction.MessageID, transaction.Date, transaction.Amount)
	return err
}

func (t *TransactionsDB) Select(from int64, to int64) ([]Transaction, error) {
	rows, err := t.db.Query(
		"SELECT id, name, messageid, date, amount FROM transactions where date > ? and date < ?", from, to)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		transaction := Transaction{}

		err = rows.Scan(
			&transaction.Id, &transaction.Name, &transaction.MessageID, &transaction.Date, &transaction.Amount)
		if err != nil {
			t.logger.Error(err.Error())
		}
		transactions = append(transactions, transaction)
	}

	err = rows.Err()

	return transactions, err
}

func (t *TransactionsDB) SelectByMessageID(messageID string) (Transaction, error) {
	transaction := Transaction{}

	row, err := t.db.Query("SELECT name FROM transactions where messageid = ?", messageID)
	if err != nil {
		return transaction, err
	}
	defer row.Close()

	err = row.Scan(&transaction.Name, &transaction.MessageID)

	return transaction, err
}
