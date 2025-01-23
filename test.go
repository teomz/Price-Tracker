package main

import (
	"github.com/teomz/Price-Tracker/api-service/scraper"
	"fmt"
)

func main() {

	source := "https://www.instocktrades.com/products/jul240881/daredevil-by-bendis-maleev-omnibus-hc-vol-02-new-ptg"
	product := scraper.ScrapeIST(source)
	fmt.Printf("Product details: %s\n", product)

}