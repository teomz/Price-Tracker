// main.go in minio package
package minio

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	docs "github.com/teomz/Price-Tracker/api-service/docs"
)

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func Initialize() *gin.Engine {
	r := gin.Default()

	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	{
		minio_group := v1.Group("/minio")
		{
			minio_group.POST("/uploadImage", uploadImage)
			minio_group.GET("/getImage", getImage)
			minio_group.DELETE("/deleteImage", deleteImage)
		}
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
