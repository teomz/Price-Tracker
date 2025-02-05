package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Sale struct {
	Date        string    `json:"date"`
	UPC         string    `json:"upc"`         // Universal Product Code
	Sale        float32   `json:"sale"`        // Daily sale price
	Platform    string    `json:"platform"`    // Platform name (e.g., IST, Amazon)
	Percent     int       `json:"percent"`     // Sale percentage over the original price
	LastUpdated time.Time `json:"LastUpdated"` // Last Update on Sale

}

func (s Sale) String() string {
	return fmt.Sprintf(
		"Sale: {Date: %s, UPC: %s, Sale Price: $%.2f, Platform: %s, Discount: %d%%,  Last Updated: %s}",
		s.Date, s.UPC, s.Sale, s.Platform, s.Percent, s.LastUpdated.Format(time.RFC3339),
	)
}

func getAmazonSale(url string) Sale {
	var record Sale
	foundFirst := false

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("www.amazon.sg", "amazon.sg"),
		colly.MaxDepth(1),
	)

	c.OnHTML("span.a-price", func(e *colly.HTMLElement) {
		if !foundFirst {
			record.Date = time.Now().Format("2006-01-02")
			priceWhole := e.ChildText("span.a-price-whole")
			priceFraction := e.ChildText("span.a-price-fraction")

			// Construct the full price
			Price := fmt.Sprintf("%s.%s", strings.Replace(priceWhole, ".", "", 1), strings.Replace(priceFraction, ".", "", 1))
			fmt.Println(Price)
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

	c.Visit(url)

	return record

}

func get() {
	url := "https://www.amazon.sg/New-Avengers-Omnibus-Vol-Printing/dp/130295914X/"
	sale := getAmazonSale(url)
	fmt.Println(sale)
}
