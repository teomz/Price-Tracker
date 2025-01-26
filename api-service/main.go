package main

import (
	"log"
	"os"

	"github.com/teomz/Price-Tracker/api-service/minio"
)

func main() {
	r := minio.Initialize()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	r.Run(":" + port)

	// Initialize the database connection
	// fmt.Println("Initializing database...")

	// Calling the initialize method from the database package
	// err := database.Initialize()
	// if err != nil {
	// 	fmt.Println("Failed to create database instance: %v", err)
	// }

	// You can add more logic here if necessary after initialization
}
