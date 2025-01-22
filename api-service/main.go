package main

import (
	"fmt"
	"github.com/teomz/Price-Tracker/api-service/database"
)

func main() {
	// Initialize the database connection
	fmt.Println("Initializing database...")

	// Calling the initialize method from the database package
	err := database.Initialize()
	if err != null {
		fmt.Println("Failed to create database instance: %v", err)
	}


	// You can add more logic here if necessary after initialization
}

