package models

import (
	"time"
)


type Omnibus struct {
	UPC         string  `json:"upc"`         // Universal Product Code
	Code        string  `json:"code"`        // A distributor SKU code
	Name        string  `json:"name"`        // Name of the omnibus
	Price       float32 `json:"price"`       // Price of the omnibus
	Version     string  `json:"version"`     // Standard or DM version
	PageCount   int     `json:"pagecount"`   // Total number of pages
	DateCreated string  `json:"releaseddate"`        // Creation date
	Publisher   string  `json:"publisher"`   // Publisher of the omnibus
	ImgPath     string  `json:"imgpath"`     // Path to the image file
	ISTUrl      string  `json:"isturl"`      // URL to IST
	AmazonUrl   string  `json:"amazonurl"`   // URL to Amazon
}

type Sale struct {
	Date     time.Time	`json:"date"`
	UPC      string     `json:"upc"`       // Universal Product Code
	Sale     float32    `json:"sale"`      // Daily sale price
	Platform string     `json:"platform"`  // Platform name (e.g., IST, Amazon)
	Percent  int        `json:"percent"`   // Sale percentage over the original price
}
