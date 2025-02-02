package main

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	docs "github.com/teomz/Price-Tracker/api-service/docs"
	"github.com/teomz/Price-Tracker/api-service/minio"
	"github.com/teomz/Price-Tracker/api-service/postgresql"
	"github.com/teomz/Price-Tracker/api-service/scraper"
)

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func initialization() *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 10 << 20
	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	{
		minio.Initialize(v1)
		postgresql.Initialize(v1)
		scraper.Initialize(v1)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return r
}
