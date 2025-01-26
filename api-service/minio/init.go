package minio

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// createMinioInstance initializes and returns a MinIO client.
func createMinioInstance() (*minio.Client, error) {

	ctx := context.Background()
	envFile := "../.env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		log.Printf("No .env file found at: %s\n", envFile)
	} else {
		// Load the .env file
		err := godotenv.Load(envFile)
		if err != nil {
			log.Printf("Error loading .env file: %v", err)
		}
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

	// Ensure the bucket 'InfoImage' exists, create it if necessary
	bucketName := "infoimage"
	location := "local"

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
			return nil, err
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
	// Return the MinIO client instance
	return minioClient, nil
}
