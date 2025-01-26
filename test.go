package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	envFile := "../.env"
	err := LoadEnvFile(envFile)
	if err != nil {
		// Handle error if necessary
		log.Println("Error occurred while loading .env file.")
	}

	fmt.Println("Connecting to Postgres ... ")

	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")
	hostDB := "localhost"
	postgresPort := os.Getenv("POSTGRES_PORT")

	conString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", postgresUser, postgresPassword, hostDB, postgresPort, postgresDB)

	pool, err := pgxpool.New(context.Background(), conString)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = pool.Ping(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Healthy")
	}

}

func LoadEnvFile(envFilePath string) error {
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		log.Printf("No .env file found at: %s\n", envFilePath)
		return nil // No error, just log and return
	} else {
		// Load the .env file
		err := godotenv.Load(envFilePath)
		if err != nil {
			log.Printf("Error loading .env file: %v", err)
			return err // Return the error if loading fails
		}
	}
	return nil
}
