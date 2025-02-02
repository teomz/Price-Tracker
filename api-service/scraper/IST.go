package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/teomz/Price-Tracker/api-service/models"
)

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

func downloadImage(imageURL string) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		fmt.Printf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	fileContents, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to save image: %v", err)
	}

	mimeType := http.DetectContentType(fileContents)

	parts := strings.Split(mimeType, "/")
	var fileExtension string
	// The second part of the MIME type is the subtype
	subtype := parts[1]
	// Map common MIME types to file extensions based on the subtype
	switch subtype {
	case "jpeg", "jpg":
		fileExtension = "jpeg"
	case "png":
		fileExtension = "png"
	}
	// Prepare multipart form-data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Extract filename from URL
	filepath := extractFilename(imageURL)
	fileName := filepath + "." + fileExtension
	fmt.Println(fileName)

	// Add the image as a file field
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	// Write the image content into the file part
	_, err = part.Write(fileContents)
	if err != nil {
		return "", fmt.Errorf("failed to write image content to form: %w", err)
	}

	// Add the extension as a text field
	err = writer.WriteField("extension", fileExtension)
	if err != nil {
		return "", fmt.Errorf("failed to write extension field: %w", err)
	}

	writer.Close()

	// Add additional query parameters
	params := url.Values{}
	params.Add("TaskUser", os.Getenv("AIRFLOW_USER"))
	params.Add("BucketNameKey", os.Getenv("SCRAPER_BUCKET"))

	// Construct the final URL with query parameters
	baseURL := "http://localhost:8080/api/v1/minio/uploadImage"
	finalURL := baseURL + "?" + params.Encode()

	// Make HTTP POST request to upload image
	req, err := http.NewRequest("POST", finalURL, body)
	if err != nil {
		fmt.Printf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if status code is OK
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload image: %s", string(bodyResponse))
	}

	var response models.SuccessResponse
	// Decode the response body to get the saved path
	if err := json.Unmarshal(bodyResponse, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return response.ObjectName, nil
}

func extractFilename(url string) string {
	parts := strings.Split(url, "/")
	filenameParts := strings.Split(parts[len(parts)-1], ".")
	return filenameParts[0]
}
