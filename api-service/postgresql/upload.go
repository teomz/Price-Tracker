package postgresql

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/teomz/Price-Tracker/api-service/models"
	"github.com/teomz/Price-Tracker/api-service/utilities"
)

// uploadInfo godoc
// @Summary Upload data to PostgreSQL
// @Description Upload an array of items (Omnibus) to the PostgreSQL database
// @Tags postgres
// @Accept json
// @Produce json
// @Param TaskUser query string true "User calling the API"
// @Param sales body []models.Omnibus true "Array of items to upload"
// @Success 200 {object} models.SuccessDataResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /postgresql/uploadInfo [post]
func uploadInfo(g *gin.Context) {

	// Load the environment file
	envFile := "../.env"
	err := utilities.LoadEnvFile(envFile)
	if err != nil {
		log.Println("Error occurred while loading .env file.")
	}

	// Check if the user is authorized
	if err := utilities.CheckUser(g, os.Getenv("AIRFLOW_USER")); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadInfo",
			Error:  "Wrong User",
		})
		return
	}

	// Create a DB instance
	db, err := createDBInstance()
	defer db.close()
	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadInfo",
			Error:  err.Error(),
		})
		return
	}

	// Bind the incoming request body to the Omnibus slice
	var omnibus []models.Omnibus
	if err := g.ShouldBindJSON(&omnibus); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadInfo",
			Error:  "Invalid request body",
		})
		return
	}

	// Prepare the SQL query with a dynamic builder for the bulk insert
	query, values := buildInsertQuery(omnibus)

	// Execute the query and handle any errors
	rows, err := db.pool.Query(context.Background(), query, values...)
	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadInfo",
			Error:  err.Error(),
		})
		return
	}
	defer rows.Close()

	// To track results
	var insertedIDs []string
	var failedItems []models.Omnibus

	// Scan the result to see which rows were inserted successfully
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Fatal("Error scanning result:", err)
		}
		insertedIDs = append(insertedIDs, id)
	}

	// For the items that weren't inserted, add them to the failedItems
	if len(insertedIDs) < len(omnibus) {
		for _, row := range omnibus {
			// If the UPC was not found in the inserted IDs, mark it as failed
			if !contains(insertedIDs, row.UPC) {
				failedItems = append(failedItems, row)
			}
		}
	}

	// Respond with success and the details of the insert
	g.JSON(http.StatusOK, models.SuccessDataResponse{
		Action:   "uploadInfo Successful",
		Inserted: insertedIDs,
		Failed:   failedItems,
	})
}

func contains(ids []string, item string) bool {
	for _, id := range ids {
		if id == item {
			return true
		}
	}
	return false
}

func buildInsertQuery(omnibus []models.Omnibus) (string, []interface{}) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		INSERT INTO omnibus (UPC, code, name, price, version, pagecount, datecreated, publisher, imgpath, isturl, amazonurl, cgnurl, lastupdated, status)
		VALUES
	`)
	values := []interface{}{}

	for i, row := range omnibus {
		if i > 0 {
			queryBuilder.WriteString(",")
		}
		queryBuilder.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			len(values)+1, len(values)+2, len(values)+3, len(values)+4, len(values)+5, len(values)+6, len(values)+7,
			len(values)+8, len(values)+9, len(values)+10, len(values)+11, len(values)+12, len(values)+13, len(values)+14))
		values = append(values, row.UPC, row.Code, row.Name, row.Price, row.Version, row.PageCount, row.DateCreated, row.Publisher,
			row.ImgPath, row.ISTUrl, row.AmazonUrl, row.CGNUrl, row.LastUpdated, row.Status)
	}

	queryBuilder.WriteString(`
		ON CONFLICT (UPC, CODE)
		DO NOTHING
		RETURNING UPC;
	`)

	return queryBuilder.String(), values
}
