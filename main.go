package main

import (
	"hellish/Database"
	"hellish/Discord"
	"hellish/crypto"
	"log"

	"github.com/joho/godotenv" // Add this import
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, relying on variables from environment.")
	}

	if err := Database.ConnectDB(); err != nil {
		log.Fatalf("Fatal error: Failed to connect to the database: %v", err)
	}
	if err := crypto.Init(); err != nil {
		log.Fatalf("Fatal error: Failed to initialize encryption: %v", err)
	}
	defer Database.DisconnectDB()

	Discord.Dc()
}
