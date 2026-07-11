package internal

import (
	"database/sql"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func InitGuestbook() error {
	path := os.Getenv("GUESTBOOK_DB_PATH")
	if path == "" {
		path = "guestbook.db"
	}

	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	// sqlite locks the whole file per write; one conn avoids "database is locked"
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS guestbook (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`)
	return err
}

type GuestbookEntry struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func AddGuestbookEntry(name, message string) (GuestbookEntry, error) {
	now := time.Now().Format(time.RFC3339)
	res, err := db.Exec(`INSERT INTO guestbook (name, message, created_at) VALUES (?, ?, ?)`, name, message, now)
	if err != nil {
		return GuestbookEntry{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return GuestbookEntry{}, err
	}
	return GuestbookEntry{ID: id, Name: name, Message: message, CreatedAt: now}, nil
}

func ListGuestbookEntries() ([]GuestbookEntry, error) {
	rows, err := db.Query(`SELECT id, name, message, created_at FROM guestbook ORDER BY id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []GuestbookEntry{}
	for rows.Next() {
		var e GuestbookEntry
		if err := rows.Scan(&e.ID, &e.Name, &e.Message, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
