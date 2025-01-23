// main.go in minio package
package minio

import (
    "github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
	docs "github.com/teomz/Price-Tracker/api-service/docs"
	
)

func Initialize() *gin.Engine {
    r := gin.Default()

	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	{
	   eg := v1.Group("/example")
	   {
		eg.GET("/helloworld",Helloworld)
	   }
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
    return r
}

