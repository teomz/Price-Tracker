package minio

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/teomz/Price-Tracker/api-service/models"
)

// @BasePath /api/v1

// deleteImage godoc
// @Summary Get a file from MinIO
// @Schemes
// @Description Get a file from MinIO storage
// @Tags minio
// @Accept json
// @Produce json
// @Param TaskUser  query string true "User calling api"
// @Param BucketNameKey query string true "The name of the minio bucket"
// @Param ObjectNameKey query string true "The name of the object with extension in minio eg image.png"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /minio/deleteImage [delete]
func deleteImage(g *gin.Context) {

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

	userID := g.DefaultQuery("TaskUser", "default_user")
	if userID != os.Getenv("AIRFLOW_USER") {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "DeleteImage",
			Error:  "Wrong User",
		})
		return
	}

	minioClient, err := createMinioInstance()

	if err != nil {
		g.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Action: "DeleteImage",
			Error:  err.Error(),
		})
		return
	}

	bucketname := g.DefaultQuery("BucketNameKey", "defaultBucket")
	objectname := g.DefaultQuery("ObjectNameKey", "defaultObject")

	err = minioClient.RemoveObject(context.Background(), bucketname, objectname, minio.RemoveObjectOptions{})
	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "DeleteImage",
			Error:  err.Error(),
		})
		return
	}

	g.JSON(http.StatusOK, models.SuccessResponse{
		Action:     "DeleteImage Successful",
		BucketName: bucketname,
		ObjectName: objectname,
	})
}
