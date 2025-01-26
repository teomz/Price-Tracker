package main

import (
	"log"
	"os"
)

func main() {

	r := initialization()
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
