package postgresql

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teomz/Price-Tracker/api-service/utilities"
)

type DatabaseController struct {
	pool *pgxpool.Pool
}

func createDBInstance() (*DatabaseController, error) {

	envFile := "../.env"
	err := utilities.LoadEnvFile(envFile)
	if err != nil {
		// Handle error if necessary
		log.Println("Error occurred while loading .env file.")
	}

	fmt.Println("Connecting to Postgres ... ")

	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")
	hostDB := os.Getenv("HOST_DB")
	postgresPort := os.Getenv("POSTGRES_PORT")

	conString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", postgresUser, postgresPassword, hostDB, postgresPort, postgresDB)

	pool, err := pgxpool.New(context.Background(), conString)
	if err != nil {
		return nil, err
	}

	db := &DatabaseController{
		pool: pool,
	}

	fmt.Println(db.health())

	return db, nil

}

func (db *DatabaseController) health() string {
	err := db.pool.Ping(context.Background())
	if err != nil {
		return err.Error()
	}

	message := "It's healthy"
	return message
}

func (db *DatabaseController) close() {
	db.pool.Close()
}

func Setup() error {
	db, err := createDBInstance()
	if err != nil {
		return err

	}
	defer db.close()

	fmt.Println("Running sql query ...")

	sqlFile := "./postgresql/sql/up_v1.sql"
	query, err := os.ReadFile(sqlFile)

	if err != nil {
		return err
	}

	_, err = db.pool.Exec(context.Background(), string(query))

	if err != nil {
		return err
	}

	fmt.Println("Done sql query ...")

	return nil
}
