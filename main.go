package main

import (
	"os"

	"git.sr.ht/~anecdotal/kitab/db"
	"git.sr.ht/~anecdotal/kitab/server"
)

func main() {
	// Initialize database connection
	db_url := os.Getenv("DATABASE_URL")
	if db_url == "" {
		db_url = "kitab.db"
	}
	db.Init(db_url)

	// Setup server object
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Init(port)
}
