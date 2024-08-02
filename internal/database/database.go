package database

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./zoombot.db")
    if err != nil {
        return nil, err
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS meetings (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            link TEXT NOT NULL,
            password TEXT,
            start_time DATETIME,
            status TEXT DEFAULT 'scheduled'
        )
    `)
    if err != nil {
        return nil, err
    }

    return db, nil
}