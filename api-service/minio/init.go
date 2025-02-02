package minio

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/teomz/Price-Tracker/api-service/utilities"
)

// createMinioInstance initializes and returns a MinIO client.
func createMinioInstance() (*minio.Client, error) {

	envFile := "../.env"
	err := utilities.LoadEnvFile(envFile)
	if err != nil {
		// Handle error if necessary
		log.Println("Error occurred while loading .env file.")
	}

	log.Println("Connecting to Minio ... ")

	// Retrieve MinIO configurations from environment variables
	endpoint := os.Getenv("MINIO_ENDPOINT") // Example: "localhost:9000"
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false // For local usage, use false (non-SSL)

	// Validate if the environment variables are set
	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("MinIO environment variables not set properly")
	}

	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Printf("Error initializing MinIO client: %v\n", err)
		return nil, err
	}

	// Return the MinIO client instance
	return minioClient, nil
}

func setup() {
	// List of bucket names to ensure their existence
	bucketNames := []string{"infoimage", "rawjson"}
	location := "local"

	// Create Minio client
	ctx := context.Background()
	minioClient, err := createMinioInstance()
	if err != nil {
		log.Printf(err.Error())
		return
	}

	// Loop through the list of bucket names and ensure each exists
	for _, bucketName := range bucketNames {
		// Try to create the bucket
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
		if err != nil {
			// Check if the bucket already exists
			exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
			if errBucketExists == nil && exists {
				log.Printf("We already own %s\n", bucketName)
			} else {
				log.Printf("Error creating bucket %s: %s\n", bucketName, err.Error())
			}
		} else {
			log.Printf("Successfully created %s\n", bucketName)
		}
	}
}
