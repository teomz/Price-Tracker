package models

import (
	"fmt"
	"time"
)

type Omnibus struct {
	UPC         string    `json:"upc"`          // Universal Product Code
	Name        string    `json:"name"`         // Name of the omnibus
	Price       float32   `json:"price"`        // Price of the omnibus
	Version     string    `json:"version"`      // Standard or DM version
	PageCount   int       `json:"pagecount"`    // Total number of pages
	DateCreated string    `json:"releaseddate"` // Creation date
	Publisher   string    `json:"publisher"`    // Publisher of the omnibus
	ImgPath     string    `json:"imgpath"`      // Path to the image file
	ISTUrl      string    `json:"isturl"`       // URL to IST
	AmazonUrl   string    `json:"amazonurl"`    // URL to Amazon
	LastUpdated time.Time `json:"LastUpdated"`  // Last Update on Info
	Status      string    `json:"status"`       // Hot , Cold , Archive
}

type Sale struct {
	Date        time.Time `json:"date"`
	UPC         string    `json:"upc"`         // Universal Product Code
	Sale        float32   `json:"sale"`        // Daily sale price
	Platform    string    `json:"platform"`    // Platform name (e.g., IST, Amazon)
	Percent     int       `json:"percent"`     // Sale percentage over the original price
	LastUpdated time.Time `json:"LastUpdated"` // Last Update on Sale

}

type ErrorResponse struct {
	Action string `json:"action"`
	Error  string `json:"error"`
}

type SuccessResponse struct {
	Action     string `json:"action"`
	BucketName string `json:"bucketname"`
	ObjectName string `json:"objectname"`
}

type SuccessDataResponse struct {
	Action   string   `json:"action"`
	Inserted []string `json:"inserted"`
}

// Struct for request validation
type QueryRequest struct {
	Query  string        `json:"query" binding:"required"`
	Values []interface{} `json:"values" binding:"required"`
}

type SuccessScraperResponse struct {
	Action string    `json:"action"`
	Data   []Omnibus `json:"data"`
}

func (o Omnibus) String() string {
	return fmt.Sprintf(
		"Omnibus: {UPC: %s, Name: %s, Publisher: %s, Price: $%.2f, Version: %s, PageCount: %d, Released: %s, IST URL: %s, Amazon URL: %s, ImgPath: %s, Last Updated: %s}",
		o.UPC, o.Name, o.Publisher, o.Price, o.Version, o.PageCount, o.DateCreated, o.ISTUrl, o.AmazonUrl, o.ImgPath, o.LastUpdated.Format(time.RFC3339),
	)
}

func (s Sale) String() string {
	return fmt.Sprintf(
		"Sale: {Date: %s, UPC: %s, Sale Price: $%.2f, Platform: %s, Discount: %d%%,  Last Updated: %s}",
		s.Date.Format(time.RFC3339), s.UPC, s.Sale, s.Platform, s.Percent, s.LastUpdated.Format(time.RFC3339),
	)
}
