package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/teomz/Price-Tracker/api-service/models"
)

var (
	omnibusData = make(map[string]*models.Omnibus) // Shared storage for Omnibus structs
	mutex       = sync.Mutex{}                     // Prevents race conditions
	wg          sync.WaitGroup                     // WaitGroup to track Scraper 2 tasks
)

func test() {
	url := "https://www.instocktrades.com/newreleases"
	var omnibusList []models.Omnibus // Store all completed JSON objects

	upcChan := make(chan string, 10)            // UPCs from IST
	resultChan := make(chan models.Omnibus, 10) // Completed Omnibus structs

	// Start amazon worker in a goroutine
	go func() {
		for upc := range upcChan {
			wg.Add(1) // Add to WaitGroup for each goroutine spawned
			go scrapeAmazonLink("https://www.amazon.sg/s?k="+upc, upc, resultChan)
		}
	}()

	// Start a goroutine to collect results
	go func() {
		for omnibus := range resultChan {
			omnibusList = append(omnibusList, omnibus) // Collect results
		}
	}()

	// Start Scraper 1 (keeps running and sending UPCs)
	scrapeIST(url, upcChan)

	// Wait for all goroutines to complete
	wg.Wait()

	// Close the channels once processing is complete
	close(upcChan)
	close(resultChan) // All goroutines finished, close the channel

	// If scraping is successful, return the data
	fmt.Println(omnibusList)
}

func scrapeISTProductInfo(e *colly.HTMLElement, product *models.Omnibus) {
	selection := e.DOM
	info := selection.Find("div.prodinfo")
	infoNodes := info.Children().Nodes

	for _, infoNode := range infoNodes {
		infoText := selection.FindNodes(infoNode).Text()
		values := strings.Split(infoText, ":")

		if len(values) > 1 {
			label := strings.TrimSpace(values[0])
			value := strings.TrimSpace(values[1])

			// Match and set the relevant product fields
			switch label {
			case "Publisher":
				product.Publisher = value
			case "UPC":
				if upc, err := strconv.ParseInt(value, 10, 64); err == nil {
					product.UPC = strconv.FormatInt(upc, 10) // Store UPC as string
				}
			case "Page Count":
				if pagecount, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64); err == nil {
					product.PageCount = int(pagecount) // Store as int
				}
			}
		}
	}
}

// scrapePricingInfo extracts the price from the product content
func scrapeISTPricingInfo(e *colly.HTMLElement, product *models.Omnibus) {
	selection := e.DOM
	pricing := selection.Find("div.pricing")
	pricingNodes := pricing.Children().Nodes

	for _, pricingNode := range pricingNodes {
		pricingText := selection.FindNodes(pricingNode).Text()
		if strings.Contains(pricingText, ":") {
			values := strings.Split(pricingText, ":")

			if len(values) > 1 {
				label := strings.TrimSpace(values[0])
				value := strings.TrimSpace(values[1])

				// Remove the dollar sign from the price and parse it
				priceStr := strings.Replace(value, "$", "", -1)
				if label == "Was" {
					if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
						product.Price = float32(price)
					}
				}
			}
		}
	}
}

func scrapeAmazonLink(source string, upc string, resultChan chan models.Omnibus) {
	c := colly.NewCollector(
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
			}
			mutex.Unlock()
			foundFirst = true
			fmt.Println(amazonurl)
			wg.Done()
		}
	})

	c.Visit(source)
}

func scrapeIST(source string, upcChan chan string) {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.instocktrades.com", "instocktrades.com"),
		colly.MaxDepth(2),
	)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting:", r.URL)
	})

	// This will get the 2nd link from source
	c.OnHTML("div[class=title]", func(e *colly.HTMLElement) {
		name := e.ChildText("a")
		if strings.Contains(name, "Omni") && strings.Contains(name, "HC") {
			link := e.Request.AbsoluteURL(e.ChildAttr("a", "href"))
			scrapeISTInfo(link, upcChan)
		}
	})

	c.OnHTML("a.btn.hotaction", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		c.Visit(link)
	})
	c.Visit(source)
}

// ScrapeIST scrapes product details from a given InstockTrades URL
func scrapeISTInfo(source string, upcChan chan string) {
	// Initialize the product struct to hold scraped data
	product := &models.Omnibus{}

	// Create a new Colly collector for instocktrades.com
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.instocktrades.com", "instocktrades.com"),
		colly.MaxDepth(2),
	)

	c.OnHTML("div.frame img", func(e *colly.HTMLElement) {
		imagePath := e.Attr("src")
		product.ImgPath = imagePath
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

		mutex.Lock()
		omnibusData[product.UPC] = product // Store in map
		mutex.Unlock()

		log.Println("Scraped IST Data for:", product.UPC)

		if product.Version == "Standard" {
			upcChan <- product.UPC // Send completed Omnibus data
		}
		fmt.Println(product.ISTUrl)
	})

	// Start scraping by visiting the source URL
	c.Visit(source)
}
