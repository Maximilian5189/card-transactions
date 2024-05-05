package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Transaction struct {
	Name string `json:"name"`
}

type TransactionsDB struct {
	db *sql.DB
}

func NewTransactions() (*TransactionsDB, error) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS transactions (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		return nil, err
	}
	return &TransactionsDB{
		db: db,
	}, nil
}

func (t *TransactionsDB) Insert(transaction Transaction) error {
	_, err := t.db.Exec("INSERT INTO transactions (name) VALUES(?);", transaction.Name)
	return err
}

func (t *TransactionsDB) Select() ([]Transaction, error) {
	rows, err := t.db.Query("SELECT name FROM transactions")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		transaction := Transaction{}

		err = rows.Scan(&transaction.Name)
		if err != nil {
			// TODO return with error? Or what?
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	err = rows.Err()

	return transactions, err
}
