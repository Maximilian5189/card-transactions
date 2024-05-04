import sqlite3 from "sqlite3"
import path from "path"

// const databasePath = path.join(process.cwd(), "database.db")
const databasePath =
  "/Users/ms/coding/card-transactions/backend/node/dist/src/database.db"

export const db = async () => {
  console.log(databasePath)
  const db = new sqlite3.Database(
    databasePath,
    sqlite3.OPEN_READWRITE,
    (err) => {
      if (err) {
        console.error(err.message)
      } else {
        console.log("Connected to the database.")
      }
    }
  )
  db.exec(`CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    email TEXT
  )`)

  // const stmt = db.prepare("INSERT INTO users (name, email) VALUES (?, ?)")
  // stmt.run("John Doe", "john@example.com")
  // stmt.run("Jane Smith", "jane@example.com")
  // stmt.finalize()

  db.all("SELECT * FROM users", (err, rows) => console.log(rows))
}
