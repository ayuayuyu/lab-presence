package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// リトライ付きで接続確認（DBコンテナ起動待ち）
	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			log.Println("database connected")
			return db, nil
		}
		log.Printf("waiting for database... (%d/30)", i+1)
		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("database not ready after 30s: %w", err)
}
