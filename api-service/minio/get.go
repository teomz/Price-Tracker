package minio

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/teomz/Price-Tracker/api-service/models"
)

// @BasePath /api/v1

// getImage godoc
// @Summary Get a file from MinIO
// @Schemes
// @Description Get a file from MinIO storage
// @Tags minio
// @Accept json
// @Produce png,jpeg
// @Param TaskUser  query string true "User calling api"
// @Param BucketNameKey query string true "The name of the minio bucket"
// @Param ObjectNameKey query string true "The name of the object with extension in minio eg image.png"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /minio/getImage [get]
func getImage(g *gin.Context) {

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
			Action: "GetImage",
			Error:  "Wrong User",
		})
		return
	}

	minioClient, err := createMinioInstance()

	if err != nil {
		g.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Action: "GetImage",
			Error:  err.Error(),
		})
		return
	}

	bucketname := g.DefaultQuery("BucketNameKey", "defaultBucket")
	objectname := g.DefaultQuery("ObjectNameKey", "defaultObject")

	object, err := minioClient.GetObject(context.Background(), bucketname, objectname, minio.GetObjectOptions{})
	defer object.Close()
	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "GetImage",
			Error:  err.Error(),
		})
		return
	}

	buffer := make([]byte, 512)
	_, err = object.Read(buffer)
	if err != nil && err != io.EOF {
		g.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Action: "GetImage",
			Error:  err.Error(),
		})
		return
	}

	mimeType := http.DetectContentType(buffer)
	g.Header("Content-Type", mimeType)

	//object.Seek(0, io.SeekStart)

	// Stream the object content directly to the HTTP response
	_, err = io.Copy(g.Writer, object)
	if err != nil {
		g.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Action: "GetImage",
			Error:  err.Error(),
		})
		return
	}

	g.JSON(http.StatusOK, models.SuccessResponse{
		Action:     "GetImage Successful",
		BucketName: bucketname,
		ObjectName: objectname,
	})

}
