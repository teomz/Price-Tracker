package minio

import (
	"github.com/gin-gonic/gin"
)

func Initialize(g *gin.RouterGroup) {
	setup()
	minio_group := g.Group("/minio")
	{
		minio_group.POST("/uploadImage", uploadImage)
		minio_group.GET("/getImage", getImage)
		minio_group.DELETE("/deleteImage", deleteImage)
	}

}
