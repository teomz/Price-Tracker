package minio

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/teomz/Price-Tracker/api-service/models"
	"github.com/teomz/Price-Tracker/api-service/utilities"
)

// @BasePath /api/v1

// uploadFile godoc
// @Summary Upload a file to MinIO
// @Schemes
// @Description Upload a file to MinIO storage
// @Tags minio
// @Accept png,jpeg
// @Produce json
// @Param TaskUser  query string true "User calling api"
// @Param file formData file true "File to be uploaded"
// @Param BucketNameKey query string true "The name of the minio bucket"
// @Param extension formData string true "File extension (e.g., png, jpeg)"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /minio/uploadImage [post]
func uploadImage(g *gin.Context) {

	err := g.Request.ParseMultipartForm(10 << 20) // Max 10 MB

	if err != nil {
		fmt.Println("Error parsing multipart form:", err)
		return
	}

	envFile := "../.env"
	err = utilities.LoadEnvFile(envFile)
	if err != nil {
		// Handle error if necessary
		log.Println("Error occurred while loading .env file.")
	}

	if err := utilities.CheckUser(g, os.Getenv("AIRFLOW_USER")); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadImage",
			Error:  "Wrong User",
		})
		return
	}

	extList := []string{"png", "jpeg"}
	mimeList := []string{"image/png", "image/jpeg"}
	if _, err := utilities.Validate_File(g, extList, mimeList); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "UploadImage",
			Error:  err.Error(),
		})
		return
	}

	minioClient, err := createMinioInstance()

	if err != nil {
		g.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Action: "UploadImage",
			Error:  err.Error(),
		})
		return
	}

	fileHeader, fileContent, err := utilities.GetFile(g)
	defer fileContent.Close()
	if err != nil {
		g.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Action: "UploadImage",
			Error:  err.Error(),
		})
		return
	}
	bucketName := g.DefaultQuery("BucketNameKey", "defaultBucket")
	bucketname := bucketName
	objectname := fileHeader.Filename

	_, err = minioClient.PutObject(context.Background(), bucketname, objectname, fileContent, fileHeader.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "UploadImage",
			Error:  err.Error(),
		})
		return
	}

	g.JSON(http.StatusOK, models.SuccessResponse{
		Action:     "UploadImage Successful",
		BucketName: bucketname,
		ObjectName: objectname,
	})
}
