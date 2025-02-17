package scraper

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/teomz/Price-Tracker/api-service/models"
	"github.com/teomz/Price-Tracker/api-service/utilities"
)

var (
	wgSale sync.WaitGroup
)

// getScrapedSale godoc
// @Summary Get sale from scraper
// @Description Get sale from scraper by providing the url and platform params
// @Tags scraper
// @Accept json
// @Produce json
// @Param TaskUser query string true "User calling the API"
// @Param URLS body []models.SaleUrls true "URL sources for scraping"
// @Success 200 {object} models.SuccessSaleResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /scraper/getScrapedSale [post]
func getScrapedSale(g *gin.Context) {

	var saleList []models.Sale // Store all completed JSON objects

	if err := utilities.CheckUser(g, os.Getenv("AIRFLOW_USER")); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "getScrapedSale",
			Error:  "Wrong User",
		})
		return
	}

	var urlList []models.SaleUrls

	if err := g.ShouldBindJSON(&urlList); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "getScrapedSale",
			Error:  err.Error(),
		})
		return
	}

	resultChan := make(chan models.Sale, 20) // Completed Sale structs

	go func() {

		for sale := range resultChan {
			saleList = append(saleList, sale) // Collect results
			wgSale.Done()
		}

	}()

	for i, url := range urlList {

		if i%20 == 0 && i != 0 {
			fmt.Println("Sleep...........")
			time.Sleep(2)
		}
		if url.Amazon != "" {
			wgSale.Add(1)
			go getAmazonSale(url.Amazon, resultChan, url.UPC)
		}
		if url.IST != "" {
			wgSale.Add(1)
			go getISTSale(url.IST, resultChan, url.UPC)
		}

	}

	wgSale.Wait()
	close(resultChan)

	// If scraping is successful, return the data
	g.JSON(http.StatusOK, models.SuccessSaleResponse{
		Action: "Scraping Sale successful",
		Data:   saleList,
	})
}

func getAmazonSale(url string, resultChan chan models.Sale, upc string) {
	log.Println("Scraping Amazon link")

	var record models.Sale
	foundFirst := false

	record.Date = time.Now().Format("2006-01-02")
	record.Platform = "Amazon"
	record.UPC = upc

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.amazon.sg", "amazon.sg"),
		colly.MaxDepth(1),
	)

	c.OnHTML("span.a-price", func(e *colly.HTMLElement) {
		if !foundFirst {
			priceWhole := e.ChildText("span.a-price-whole")
			priceFraction := e.ChildText("span.a-price-fraction")

			// Construct the full price
			Price := fmt.Sprintf("%s.%s", strings.Replace(priceWhole, ".", "", 1), strings.Replace(priceFraction, ".", "", 1))
			sale, err := strconv.ParseFloat(Price, 32)
			if err != nil {
				record.Sale = float32(-1)
			} else {
				record.Sale = float32(sale)
			}
			//fmt.Println(record.Sale)
			foundFirst = true
		}

	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("error in amazon")
		// Log the error details
		record.Sale = float32(-1)
	})

	c.Visit(url)

	resultChan <- *&record

}

func getISTSale(url string, resultChan chan models.Sale, upc string) {

	log.Println("Scraping IST link")

	var record models.Sale

	record.Date = time.Now().Format("2006-01-02")
	record.UPC = upc
	record.Platform = "IST"

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.instocktrades.com", "instocktrades.com"),
		colly.MaxDepth(1),
	)

	c.OnHTML("div.pricing", func(e *colly.HTMLElement) {

		salePriceText := e.ChildText("div.price")
		re := regexp.MustCompile(`\$(\d+\.\d{2})`)
		match := re.FindStringSubmatch(salePriceText)
		if len(match) > 1 {
			price, err := strconv.ParseFloat(match[1], 32)
			if err != nil {
				log.Println("Error converting price:", err)
				record.Sale = float32(-1)
			}
			record.Sale = float32(price)
		} else {
			record.Sale = float32(-1)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		// Log the error details
		log.Println("error in ist")

		record.Sale = float32(-1)
	})

	c.Visit(url)

	resultChan <- *&record

}
