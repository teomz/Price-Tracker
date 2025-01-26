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

func Setup() error {

	// Ensure the bucket 'InfoImage' exists, create it if necessary

	ctx := context.Background()
	minioClient, err := createMinioInstance()
	if err != nil {
		return err
	}
	bucketName := "infoimage"
	location := "local"

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			return err
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	return nil
}
