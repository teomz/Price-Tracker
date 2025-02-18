package postgresql

import (
	"github.com/gin-gonic/gin"
)

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func Initialize(g *gin.RouterGroup) {

	postgresql_group := g.Group("/postgresql")
	{
		postgresql_group.POST("/uploadInfo", uploadInfo)
		postgresql_group.GET("/getInfoByDate", getInfoByDate)
	}

}
