package scraper

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/teomz/Price-Tracker/api-service/models"
	"github.com/teomz/Price-Tracker/api-service/utilities"
)

var (
	omnibusData = make(map[string]*models.Omnibus) // Shared storage for Omnibus structs
	mutex       = sync.Mutex{}                     // Prevents race conditions
	wg          sync.WaitGroup                     // WaitGroup
	retry       = sync.Mutex{}                     // Mutex to protect shared resources
	retryCheck  = make(map[string]int)
	// visited     = make(map[string]bool)
	// visitng     = sync.Mutex{}
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
	worker := 4

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

	upcChan := make(chan string, worker)            // UPCs from IST
	resultChan := make(chan models.Omnibus, worker) // Completed Omnibus structs
	successChan := make(chan bool, worker)          // Completed Omnibus structs

	// Start amazon worker in a goroutine
	go func() {
		counter := 0
		for upc := range upcChan {
			counter++
			if counter <= worker {
				// For the first two items, just start them directly
				log.Print("Start worker for ", upc)
				go scrapeAmazonLink("https://www.amazon.sg/s?k="+upc, upc, resultChan)
			} else {
				// Wait for one of the previous workers to finish before starting a new one
				<-successChan
				log.Print("Received Success and Starting worker for ", upc)
				go scrapeAmazonLink("https://www.amazon.sg/s?k="+upc, upc, resultChan)
			}
		}
	}()

	// if i%%10

	// Start a goroutine to collect
	go func() {

		for omnibus := range resultChan {
			log.Println("Goroutine Collection for ", omnibus.UPC)
			omnibusList = append(omnibusList, omnibus) // Collect results
			successChan <- true
			wg.Done()
		}

	}()

	// Start Scraper 1 (keeps running and sending UPCs)
	scrapeIST(url, upcChan, resultChan)
	wg.Wait()
	close(upcChan)
	close(resultChan) // All goroutines finished, close the channel

	fmt.Println("Scarping Done")

	// If scraping is successful, return the data
	g.JSON(http.StatusOK, models.SuccessScraperResponse{
		Action: "Scraping successful",
		Data:   omnibusList,
	})
}

func scrapeAmazonLink(source string, upc string, resultChan chan models.Omnibus) {
	const maxRetries = 55

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.amazon.sg", "amazon.sg"),
		colly.MaxDepth(5),
	)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	// This will run after a page is visited and the source is fetched
	c.OnResponse(func(r *colly.Response) {
		log.Println("Successfully fetched:", r.Request.URL)
	})

	var substring string
	foundFirst := false
	c.OnHTML("a.a-link-normal", func(e *colly.HTMLElement) {
		htmlText := e.Text
		// Check if we already found the first match
		if !foundFirst {
			log.Println("Found Link in Amazon")
			// Extract href attribute value
			href := e.Attr("href")
			parts := strings.Split(href, "/ref=")

			// Extract the part before "/ref="
			substring = parts[0]
			amazonurl := "https://www.amazon.sg" + substring
			mutex.Lock()
			if omnibus, exists := omnibusData[upc]; exists {
				if htmlText == "Visit the help section" || strings.Contains(substring, "gp/help/customer/") {
					omnibus.AmazonUrl = ""
				} else {
					omnibus.AmazonUrl = amazonurl
				}
				omnibus.AmazonUrl = amazonurl
				resultChan <- *omnibus // Send completed Omnibus data
				log.Println("Found Link: Send result AmazonURL ", omnibus.AmazonUrl, " for ", upc)
			}
			mutex.Unlock()
			foundFirst = true
			// successChan <- true
			// log.Println("Found Link: Send successChan for ", upc)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error for ", upc)
		retry.Lock()
		retryCount, exists := retryCheck[upc]
		if !exists {
			retryCheck[upc] = 1
			retryCount = 1
		} else {
			retryCount++
			retryCheck[upc] = retryCount
		}
		retry.Unlock()
		// Log the error details
		if retryCount <= maxRetries {
			randomSleepDuration := time.Duration(rand.Intn(6)+5) * time.Second
			time.Sleep(randomSleepDuration)
			log.Printf("Retrying %s (attempt %d/%d)", source, retryCount, maxRetries)
			go scrapeAmazonLink("https://www.amazon.sg/s?k="+upc, upc, resultChan)
		} else {
			mutex.Lock()
			if omnibus, exists := omnibusData[upc]; exists {
				omnibus.AmazonUrl = err.Error()
				resultChan <- *omnibus // Send completed Omnibus data
				log.Println("Error: Send empty AmazonURL for ", upc)
			}
			mutex.Unlock()
			foundFirst = true
			// successChan <- true
			// log.Println("Error: Send successChan for ", upc)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		// log.Println("OnScraped")
		if !foundFirst {
			// Log the error details
			mutex.Lock()
			if omnibus, exists := omnibusData[upc]; exists {
				omnibus.AmazonUrl = ""
				resultChan <- *omnibus // Send completed Omnibus data
				log.Println("OnScrape: Send no result AmazonURL for ", upc)
			}
			mutex.Unlock()
			foundFirst = true
			// successChan <- true
			// log.Println("OnScrape: Send successChan for ", upc)
		}

	})

	c.Visit(source)

}

func scrapeIST(source string, upcChan chan string, resultChan chan models.Omnibus) {

	//visited[source+"?pg=1"] = true
	pgcount := 1

	c := colly.NewCollector(

		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.instocktrades.com", "instocktrades.com"),
	)
	// 	colly.Async(true),
	// )

	// c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})

	// this will get the 2nd link from source
	c.OnHTML("div[class=title]", func(e *colly.HTMLElement) {
		name := e.ChildText("a")
		if (strings.Contains(name, "Omni") && strings.Contains(name, "HC")) ||
			((strings.Contains(name, "Deluxe") || strings.Contains(name, "DLX")) && strings.Contains(name, "HC")) ||
			(strings.Contains(name, "Library") && strings.Contains(name, "HC")) ||
			(strings.Contains(name, "Omni") && strings.Contains(name, "Conan")) ||
			((strings.Contains(name, "Teenage Mutant Ninja Turtles") || strings.Contains(name, "TMNT")) && strings.Contains(name, "HC")) {
			wg.Add(1)
			link := e.Request.AbsoluteURL(e.ChildAttr("a", "href"))
			log.Println(link)
			go scrapeISTInfo(link, upcChan, resultChan)
		}
	})

	c.OnHTML("a.btn.hotaction", func(e *colly.HTMLElement) {
		// link := e.Request.AbsoluteURL(e.Attr("href"))
		// visitng.Lock()
		// if _, exists := visited[link]; !exists {
		// 	visited[link] = true
		// 	visitng.Unlock()
		// 	c.Visit(link)
		// } else {
		// 	visitng.Unlock()
		// }
		link := e.Request.AbsoluteURL(e.Attr("href"))
		parsedURL, _ := url.Parse(link)
		pageParam := parsedURL.Query().Get("pg")
		pageNumber, _ := strconv.Atoi(pageParam)
		if pageNumber > pgcount {
			fmt.Println(pgcount)
			pgcount++
			c.Visit(link) // Visit the next page
		}
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

		if strings.Contains(strings.ToLower(product.Name), " dm ") {
			product.Version = "DM"
		} else if strings.Contains(strings.ToLower(product.Name), "direct market") {
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

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error for ISTurl: ", r.Request.URL)
		wg.Done()
	})

	// Start scraping by visiting the source URL
	c.Visit(source)

}
