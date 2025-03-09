package scraper

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/teomz/Price-Tracker/api-service/models"
)

// getCurrency godoc
// @Summary Get currency from scraper
// @Description Get fx rate from scraper by providing the url
// @Tags scraper
// @Accept json
// @Produce json
// @Param TaskUser query string true "User calling the API" example("user123")
// @Success 200 {object} models.SuccessDataResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /scraper/getCurrency [get]
func getCurrency(g *gin.Context) {

	url := "https://sg.finance.yahoo.com/quote/SGD%3DX/"

	price, err := getYahooCurrency(url)
	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "getCurrency",
			Error:  err.Error(),
		})
	}

	g.JSON(http.StatusOK, models.SuccessDataResponse{
		Action:   "getCurrency successful",
		Inserted: []string{price},
	})

}

func getYahooCurrency(url string) (string, error) {

	foundFirst := false
	// Initialize the collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("sg.finance.yahoo.com"),
		colly.MaxDepth(1),
	)

	// Declare a variable to store the price
	var price string

	// Define the OnRequest handler
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	// Define the OnHTML handler
	c.OnHTML("span.base", func(e *colly.HTMLElement) {
		// Extract the price from the span element
		if !foundFirst {
			foundFirst = true
			price = e.Text
		}
	})

	// Define the error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error:", err.Error())
	})

	// Start visiting the URL
	err := c.Visit(url)
	if err != nil {
		return "", err // Return error if URL visit fails
	}

	// Wait for the scraping to complete
	c.Wait()

	// Return the price after scraping
	if price == "" {
		return "", fmt.Errorf("price not found")
	}

	return price, nil
}
