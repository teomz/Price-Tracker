package scraper

import (
	"github.com/gin-gonic/gin"
)

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func Initialize(g *gin.RouterGroup) {

	scraper_group := g.Group("/scraper")
	{
		scraper_group.GET("/getScrapedInfo", getScrapedInfo)
		scraper_group.GET("/getScrapedSale", getScrapedSale)
	}

}
