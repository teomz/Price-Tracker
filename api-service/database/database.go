package database

import (
	"context"
	"fmt"
	"os"
	"io/ioutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type DatabaseController struct {
	pool *pgxpool.Pool
}

func createDBInstance() (*DatabaseController, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("Error loading .env file: %v", err)
	}

	fmt.Println("Connecting to Postgres ... ")

	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")
	hostDB := os.Getenv("HOST_DB")
	postgresPort := os.Getenv("POSTGRES_PORT")

	fmt.Println(postgresDB)

	conString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", postgresUser, postgresPassword, hostDB, postgresPort, postgresDB)



	pool, err := pgxpool.New(context.Background(), conString)
	if err != nil {
		return nil, err
	}

	db := &DatabaseController{
		pool: pool,
	}

	fmt.Println(db.Health())

	return db, nil

}

func (db *DatabaseController) Health() string {
	err := db.pool.Ping(context.Background())
	if err != nil {
		return err.Error()
	}

	message := "It's healthy"
	return message
}

func (db *DatabaseController) Close() {
	db.pool.Close()
}

func Initialize() error {
	db, err := createDBInstance()
	if err != nil {
		return err

	}
	defer db.Close()

	fmt.Println("Running sql query ...")

	sqlFile := "./database/sql/up_v1.sql"
	query, err := ioutil.ReadFile(sqlFile)

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