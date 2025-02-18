package scraper

import (
	"html"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/teomz/Price-Tracker/api-service/models"
	"github.com/teomz/Price-Tracker/api-service/utilities"
)

var (
	omnibusData = make(map[string]*models.Omnibus) // Shared storage for Omnibus structs
	mutex       = sync.Mutex{}                     // Prevents race conditions
	wg          sync.WaitGroup                     // WaitGroup
)

// getScrapedInfo godoc
// @Summary Get data from scraper
// @Description Get data from scraper by providing the url and platform params
// @Tags scraper
// @Accept json
// @Produce json
// @Param TaskUser query string true "User calling the API"
// @Param URL query string true "URL source for scraping"
// @Success 200 {object} models.SuccessScraperResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /scraper/getScrapedInfo [get]
func getScrapedInfo(g *gin.Context) {
	// platform := g.DefaultQuery("Platform", "default_platform")
	// var err error
	var omnibusList []models.Omnibus // Store all completed JSON objects

	if err := utilities.CheckUser(g, os.Getenv("AIRFLOW_USER")); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "GetScrapedInfo",
			Error:  "Wrong User",
		})
		return
	}

	url := html.UnescapeString(g.DefaultQuery("URL", ""))

	if url == "" {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "GetScrapedInfo",
			Error:  "URL is required for scraping",
		})
		return
	}

	upcChan := make(chan string, 20)            // UPCs from IST
	resultChan := make(chan models.Omnibus, 20) // Completed Omnibus structs

	// Start amazon worker in a goroutine
	go func() {
		for upc := range upcChan {
			go scrapeAmazonLink("https://www.amazon.sg/s?k="+upc, upc, resultChan)
		}
	}()

	// if i%%10

	// Start a goroutine to collect
	go func() {

		for omnibus := range resultChan {
			omnibusList = append(omnibusList, omnibus) // Collect results
			wg.Done()
		}

	}()

	// Start Scraper 1 (keeps running and sending UPCs)
	scrapeIST(url, upcChan, resultChan)
	wg.Wait()
	close(upcChan)
	close(resultChan) // All goroutines finished, close the channel

	// If scraping is successful, return the data
	g.JSON(http.StatusOK, models.SuccessScraperResponse{
		Action: "Scraping successful",
		Data:   omnibusList,
	})
}

func scrapeAmazonLink(source string, upc string, resultChan chan models.Omnibus) {

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.amazon.sg", "amazon.sg"),
		colly.MaxDepth(1),
	)

	var substring string
	foundFirst := false
	c.OnHTML("a.a-link-normal.s-underline-text.s-underline-link-text.s-link-style.a-text-normal[href]", func(e *colly.HTMLElement) {
		// Check if we already found the first match
		if !foundFirst {
			// Extract href attribute value
			href := e.Attr("href")
			parts := strings.Split(href, "/ref=")

			// Extract the part before "/ref="
			substring = parts[0]
			amazonurl := "https://www.amazon.sg" + substring
			mutex.Lock()
			if omnibus, exists := omnibusData[upc]; exists {
				omnibus.AmazonUrl = amazonurl
				resultChan <- *omnibus // Send completed Omnibus data
			} else {
			}
			mutex.Unlock()
			foundFirst = true

		}
	})

	c.OnError(func(r *colly.Response, err error) {
		// Log the error details
		mutex.Lock()
		if omnibus, exists := omnibusData[upc]; exists {
			omnibus.AmazonUrl = "Error"
			resultChan <- *omnibus // Send completed Omnibus data
		} else {
			// Optionally handle the case when omnibus data is not found
		}
		mutex.Unlock()
		foundFirst = true
	})

	c.OnScraped(func(r *colly.Response) {
		if !foundFirst {
			// Log the error details
			mutex.Lock()
			if omnibus, exists := omnibusData[upc]; exists {
				resultChan <- *omnibus // Send completed Omnibus data
			} else {
			}
			mutex.Unlock()
			foundFirst = true
		}

	})

	c.Visit(source)

}

func scrapeIST(source string, upcChan chan string, resultChan chan models.Omnibus) {
	c := colly.NewCollector(

		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.instocktrades.com", "instocktrades.com"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})

	// this will get the 2nd link from source
	c.OnHTML("div[class=title]", func(e *colly.HTMLElement) {
		name := e.ChildText("a")
		if (strings.Contains(name, "Omni") && strings.Contains(name, "HC")) ||
			(strings.Contains(name, "Deluxe") && strings.Contains(name, "HC")) ||
			(strings.Contains(name, "Library") && strings.Contains(name, "HC")) ||
			(strings.Contains(name, "Omni") && strings.Contains(name, "Conan")) {
			wg.Add(1)
			link := e.Request.AbsoluteURL(e.ChildAttr("a", "href"))
			go scrapeISTInfo(link, upcChan, resultChan)
		}
	})

	c.OnHTML("a.btn.hotaction", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		c.Visit(link)

	})
	c.Visit(source)
	c.Wait()
}

// ScrapeIST scrapes product details from a given InstockTrades URL
func scrapeISTInfo(source string, upcChan chan string, resultChan chan models.Omnibus) {

	// Initialize the product struct to hold scraped data
	product := &models.Omnibus{}

	// Create a new Colly collector for instocktrades.com
	c := colly.NewCollector(

		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.instocktrades.com", "instocktrades.com"),
		colly.MaxDepth(1),
	)

	c.OnHTML("div.frame img", func(e *colly.HTMLElement) {
		imagePath, err := downloadImage(e.Attr("src"))
		if err != nil {
			log.Println("Failed to download image:", err)
		} else {
			product.ImgPath = imagePath
		}
	})

	// Handle the product content scraping for the 2nd link
	c.OnHTML("div.productcontent", func(e *colly.HTMLElement) {
		// Set the IST product URL and name
		product.ISTUrl = e.Request.URL.String()
		product.Name = e.ChildText("h1")

		if strings.Contains(strings.ToLower(product.Name), "dm") {
			product.Version = "DM"
		} else {
			product.Version = "Standard"
		}

		// Scrape additional product info (Publisher, UPC)
		scrapeISTProductInfo(e, product)

		// Scrape pricing info (Price)
		scrapeISTPricingInfo(e, product)

		product.DateCreated = time.Now()

		if product.Version == "Standard" {
			mutex.Lock()
			omnibusData[product.UPC] = product // Store in map
			mutex.Unlock()
			upcChan <- product.UPC // Send completed Omnibus data
		} else {
			resultChan <- *product
		}
	})
	// Start scraping by visiting the source URL
	c.Visit(source)

}
