package scraper

import (
	"fmt"
	"log"
	"math/rand"
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
	wgSale         sync.WaitGroup
	saleRetry      = sync.Mutex{} // Mutex to protect shared resources
	saleRetryCheck = make(map[string]int)
)

// getScrapedSale godoc
// @Summary Get sale from scraper
// @Description Get sale from scraper by providing the url and platform params
// @Tags scraper
// @Accept json
// @Produce json
// @Param TaskUser query string true "User calling the API" example("user123")
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

	worker := 2
	urlChan := make(chan models.SaleUrls, worker)
	resultChan := make(chan models.Sale, worker*2) // Completed Sale structs
	amazonChan := make(chan models.SaleUrls, worker)
	successChan := make(chan bool, worker) // Completed Omnibus structs

	go func() {

		for sale := range resultChan {
			saleList = append(saleList, sale) // Collect results
			wgSale.Done()
		}

	}()

	go func() {
		for url := range urlChan {
			if url.Amazon != "" && url.IST != "" {
				wgSale.Add(1)
			}
			if url.Amazon != "" {
				amazonChan <- url
			}
			if url.IST != "" {
				go getISTSale(url.IST, resultChan, url.UPC)
			}
		}

	}()

	go func() {
		counter := 0
		for amazon := range amazonChan {
			counter++
			if counter <= worker {
				// For the first two items, just start them directly
				go getAmazonSale(amazon.Amazon, resultChan, amazon.UPC, successChan)
			} else {
				// Wait for one of the previous workers to finish before starting a new one
				<-successChan
				go getAmazonSale(amazon.Amazon, resultChan, amazon.UPC, successChan)
			}
		}

	}()

	for _, url := range urlList {
		wgSale.Add(1)
		urlChan <- url
	}

	wgSale.Wait()
	close(urlChan)
	close(amazonChan)
	close(successChan)
	close(resultChan)

	// worker := 1
	// urlChan := make(chan models.FlatData, worker)
	// resultChan := make(chan models.Sale, worker) // Completed Sale structs
	// successChan := make(chan bool, worker)       // Completed Omnibus structs

	// var flattened []models.FlatData

	// // Iterate over each record and flatten it
	// for _, data := range urlList {
	// 	if data.Amazon != "" {
	// 		flattened = append(flattened, models.FlatData{
	// 			Source: "amazon",
	// 			UPC:    data.UPC,
	// 			URL:    data.Amazon,
	// 		})
	// 	}
	// 	if data.IST != "" {
	// 		flattened = append(flattened, models.FlatData{
	// 			Source: "ist",
	// 			UPC:    data.UPC,
	// 			URL:    data.IST,
	// 		})
	// 	}
	// }

	// go func() {
	// 	counter := 0
	// 	for url := range urlChan {
	// 		counter++
	// 		if counter <= worker {
	// 			// For the first two items, just start them directly
	// 			if url.Source == "amazon" {
	// 				go getAmazonSale(url.URL, resultChan, url.UPC, successChan)
	// 			} else {
	// 				go getISTSale(url.URL, resultChan, url.UPC, successChan)
	// 			}
	// 		} else {
	// 			// Wait for one of the previous workers to finish before starting a new one
	// 			<-successChan
	// 			if url.Source == "amazon" {
	// 				go getAmazonSale(url.URL, resultChan, url.UPC, successChan)
	// 			} else {
	// 				go getISTSale(url.URL, resultChan, url.UPC, successChan)
	// 			}
	// 		}
	// 	}

	// }()

	// go func() {
	// 	for sale := range resultChan {
	// 		log.Printf("Recording %s: Received Price $%f for %s ", sale.Platform, sale.Sale, sale.UPC)
	// 		saleList = append(saleList, sale) // Collect results
	// 		successChan <- true
	// 		wgSale.Done()
	// 	}
	// }()

	// for _, url := range flattened {
	// 	wgSale.Add(1)
	// 	urlChan <- url
	// }

	// wgSale.Wait()
	// close(urlChan)
	// close(successChan)
	// close(resultChan)

	// If scraping is successful, return the data
	g.JSON(http.StatusOK, models.SuccessSaleResponse{
		Action: "Scraping Sale successful",
		Data:   saleList,
	})
}

func getAmazonSale(url string, resultChan chan models.Sale, upc string, successChan chan bool) {
	log.Println("Scraping Amazon link:", url)

	var record models.Sale
	foundFirst := false
	maxRetries := 50
	haveRetried := false

	record.Date = time.Now()
	record.Platform = "Amazon"
	record.UPC = upc
	record.Sale = float32(-1)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.amazon.sg", "amazon.sg"),
		colly.MaxDepth(1),
	)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL, " ", upc)
	})

	// This will run after a page is visited and the source is fetched
	c.OnResponse(func(r *colly.Response) {
		log.Println("Successfully fetched:", r.Request.URL, " ", upc)
		log.Println(r.StatusCode)
	})

	c.OnHTML("span.a-price", func(e *colly.HTMLElement) {
		if !foundFirst {
			priceWhole := e.ChildText("span.a-price-whole")
			priceFraction := e.ChildText("span.a-price-fraction")
			// Construct the full price
			price := fmt.Sprintf("%s.%s", strings.Replace(priceWhole, ".", "", 1), strings.Replace(priceFraction, ".", "", 1))
			log.Printf("Price Found for %s : %s", upc, price)
			sale, err := strconv.ParseFloat(price, 32)
			if err != nil {
				record.Sale = float32(-1)
			} else {
				record.Sale = float32(sale)
			}
			//fmt.Println(record.Sale)
			foundFirst = true
			// log.Printf("%s: Sending Price $%f for %s ", record.Platform, record.Sale, record.UPC)
			// resultChan <- record
			// successChan <- true
		}

	})

	c.OnHTML("title", func(e *colly.HTMLElement) {
		titleText := e.Text
		if strings.Contains(titleText, "Server Busy") {
			log.Println("Detected 'Server Busy' page. Retrying...")

			saleRetry.Lock()
			retryCount, exists := saleRetryCheck[upc]
			if !exists {
				saleRetryCheck[upc] = 1
				retryCount = 1
			} else {
				retryCount++
				saleRetryCheck[upc] = retryCount
			}
			saleRetry.Unlock()

			if retryCount <= maxRetries {
				randomSleepDuration := time.Duration(rand.Intn(6)+5) * time.Second
				time.Sleep(randomSleepDuration)
				log.Printf("Retrying %s (attempt %d/%d)", upc, retryCount, maxRetries)
				haveRetried = true
				go getAmazonSale(url, resultChan, upc, successChan)
			}
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println(err.Error())
		// Log the error details
		record.Sale = float32(-1)
		// log.Printf("%s: Sending Price $%f for %s ", record.Platform, record.Sale, record.UPC)
		// resultChan <- record
		// successChan <- true
	})

	// c.OnScraped(func(r *colly.Response) {
	// 	if !foundFirst {
	// 		log.Println("OnScraped:", upc)
	// 		// Log the error details
	// 		record.Sale = float32(-1)
	// 		foundFirst = true
	// 		log.Printf("%s: Sending Price $%f for %s ", record.Platform, record.Sale, record.UPC)
	// 		resultChan <- record
	// 		successChan <- true
	// 	}

	// })

	c.Visit(url)

	if !haveRetried {
		log.Printf("%s: Sending Price $%f for %s ", record.Platform, record.Sale, record.UPC)
		resultChan <- record
		successChan <- true
	}

}

func getISTSale(url string, resultChan chan models.Sale, upc string) {

	log.Println("Scraping IST link")

	var record models.Sale

	record.Date = time.Now()
	record.UPC = upc
	record.Platform = "IST"
	record.Sale = float32(-1)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.instocktrades.com", "instocktrades.com"),
		colly.MaxDepth(1),
	)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL, " ", upc)
	})

	// This will run after a page is visited and the source is fetched
	c.OnResponse(func(r *colly.Response) {
		log.Println("Successfully fetched:", r.Request.URL)
		log.Println(r.StatusCode)
	})

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

	log.Printf("%s: Sending Price $%f for %s ", record.Platform, record.Sale, record.UPC)
	resultChan <- record

}
