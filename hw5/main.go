package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Example: postgres://user:pass@localhost:5432/hw5?sslmode=disable
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	repo := NewRepository(db)
	srv := NewServer(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/users", srv.handleGetUsers)
	mux.HandleFunc("/users/common-friends", srv.handleGetCommonFriends)

	addr := ":8081"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

