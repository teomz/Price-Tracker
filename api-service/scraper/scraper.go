package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"github.com/gocolly/colly"
	"github.com/teomz/Price-Tracker/api-service/models"
)

// ScrapeIST scrapes product details from a given InstockTrades URL
func ScrapeIST(source string) models.Omnibus {
	// Initialize the product struct to hold scraped data
	product := models.Omnibus{}

	// Create a new Colly collector for instocktrades.com
	c := colly.NewCollector(
		colly.AllowedDomains("www.instocktrades.com", "instocktrades.com"),
	)

	// Handle the product content scraping
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
		scrapeISTProductInfo(e, &product)

		// Scrape pricing info (Price)
		scrapeISTPricingInfo(e, &product)

		product.DateCreated = time.Now().Format("2006-01-02")


	})

	// Start scraping by visiting the source URL
	if err := c.Visit(source); err != nil {
		fmt.Println("Error visiting source:", err)
	}

	// Return the scraped product
	return product
}

// scrapeProductInfo extracts Publisher and UPC from the product content
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