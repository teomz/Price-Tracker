package scraper


import (
	"github.com/teomz/Price-Tracker/component/models"
	"github.com/gocolly/colly"
	"fmt"
	"strings"
)



func ScrapeComicGeek(source string) models.Omnibus {

	product := models.Omnibus{}

	c := colly.NewCollector(
		colly.AllowedDomains("leagueofcomicgeeks.com"),
	)

	c.OnHTML("div.page-details", func(e *colly.HTMLElement) {
		// Extract the comic title (from h1 tag inside the div)
		product.Name = e.ChildText("h1")

		// Extract publisher and release date from the "div.header-intro"
		parts := strings.SplitN(strings.TrimSpace(e.ChildText("div.header-intro a")), " ", 2)
		product.Publisher = parts[0]
		// Extract release date from the second anchor tag inside the "header-intro" (in the format "Released Dec 31st, 2024")
		releaseDate := e.ChildText("div.header-intro a[style='font-weight: normal;']")
		product.DateCreated = releaseDate

	})


	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r.StatusCode, "Error:", err)
	})

	// Visit the source URL
	err := c.Visit(source)
	if err != nil {
		fmt.Println("Error visiting source:", err)
	}

	return product
	

}

