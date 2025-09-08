package main

import (
	"hellish/Database"
	"hellish/Discord"
	"hellish/crypto"
	"log"
)

func main() {

	if err := Database.ConnectDB(); err != nil {
		log.Fatalf("Fatal error: Failed to connect to the database: %v", err)
	}
	if err := crypto.Init(); err != nil {
		log.Fatalf("Fatal error: Failed to initialize encryption: %v", err)
	}
	defer Database.DisconnectDB()

	Discord.Dc()
}
