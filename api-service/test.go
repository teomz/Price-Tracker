package main

import (
	"github.com/teomz/Price-Tracker/api-service/scraper"
	"fmt"
)

func main() {

	source := "https://leagueofcomicgeeks.com/comic/4511150/daredevil-by-chip-zdarsky-omnibus-vol-2-hc"
	product := scraper.ScrapeComicGeek(source)
	fmt.Printf("Product Name: %s, Date: %s, Publisher: %s\n", product.Name, product.DateCreated, product.Publisher)

}