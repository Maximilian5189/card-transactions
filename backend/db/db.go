package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Transaction struct {
	Name      string `json:"name"`
	MessageID string `json:"messageID"`
	Date      string `json:"date"`
}

type TransactionsDB struct {
	db *sql.DB
}

func NewTransactionsDB() (*TransactionsDB, error) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS transactions (id INTEGER PRIMARY KEY, name TEXT, messageid TEXT UNIQUE, date TEXT)")
	if err != nil {
		return nil, err
	}
	return &TransactionsDB{
		db: db,
	}, nil
}

func (t *TransactionsDB) Insert(transaction Transaction) error {
	_, err := t.db.Exec(
		"INSERT INTO transactions (name, messageid, date) VALUES(?,?,?);",
		transaction.Name, transaction.MessageID, transaction.Date)
	return err
}

func (t *TransactionsDB) Select() ([]Transaction, error) {
	rows, err := t.db.Query("SELECT name, messageid, date FROM transactions")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		transaction := Transaction{}

		err = rows.Scan(&transaction.Name, &transaction.MessageID, &transaction.Date)
		if err != nil {
			// TODO return with error? Or what?
			return nil, err
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
