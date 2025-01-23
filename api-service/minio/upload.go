package minio

import (
    // "context"
    // "github.com/minio/minio-go/v7"
    // "github.com/minio/minio-go/v7/pkg/credentials"
    "github.com/gin-gonic/gin"
    "net/http"
    // "os"
)
// @BasePath /api/v1
// uploadFile godoc
// @Summary Upload a file to MinIO
// @Schemes
// @Description Upload a file to MinIO storage
// @Accept  multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {string} url_of_the_uploaded_file
// @Router /minio/upload [post]

// @BasePath /api/v1

// func uploadFile(c *gin.Context) {
//     // file, header, err := c.Request.FormFile("file")
//     // if err != nil {
//     //     c.JSON(http.StatusBadRequest, "Failed to upload file")
//     //     return
//     // }
//     // defer file.Close()

//     // objectName := header.Filename
//     // bucketName := "images" // S3 Bucket

//     // // Initialize MinIO client
//     // minioClient, err := minio.New(
//     //     os.Getenv("MINIO_ENDPOINT"),
//     //     &minio.Options{
//     //         Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""),
//     //         Secure: false, // Set this to true if you're using HTTPS
//     //     },
//     // )
//     // if err != nil {
//     //     c.JSON(http.StatusInternalServerError, "Failed to connect to MinIO")
//     //     return
//     // }

//     // // Upload the file to MinIO
//     // _, err = minioClient.PutObject(context.Background(), bucketName, objectName, file, -1, minio.PutObjectOptions{})
//     // if err != nil {
//     //     c.JSON(http.StatusInternalServerError, "Failed to upload file to MinIO")
//     //     return
//     // }

//     // // Generate the file URL
//     // fileURL := "http://" + os.Getenv("MINIO_ENDPOINT") + "/images/" + objectName
//     // // If using HTTPS, you can modify the URL like this:
//     // // fileURL := "https://" + os.Getenv("MINIO_ENDPOINT") + "/images/" + objectName

//     c.JSON(http.StatusOK, "Testing")
// }

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func Helloworld(g *gin.Context)  {
	g.JSON(http.StatusOK,"helloworld")
 }