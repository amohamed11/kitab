package main

import (
	"git.sr.ht/~anecdotal/kitab/db"
	"git.sr.ht/~anecdotal/kitab/server"
)

func main() {
	// Initialize database connection
	db.Init()

	// Setup server object
	server.Init("8080")
}
