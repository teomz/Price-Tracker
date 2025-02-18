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
// @Param TaskUser query string true "User calling the API" example("user123")
// @Param Query body []interface{} true "Array of items to upload" example()
// @Success 200 {object} models.SuccessDataResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /postgresql/uploadInfo [post]
func uploadInfo(g *gin.Context) {

	var allowedQueryTypes = []string{"INSERT"}

	var allowedTables = []string{"omnibus", "sale"}

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

	var req []map[string]interface{}

	// Bind JSON request
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadInfo",
			Error:  "Invalid JSON format",
		})
		return
	}

	query, values, err := buildInsertQuery(req, "omnibus")

	fmt.Println(values)

	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadInfo",
			Error:  err.Error(),
		})
		return
	}

	// Validate query
	if err := utilities.ValidateQuery(query, allowedQueryTypes, allowedTables); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadInfo",
			Error:  err.Error(),
		})
		return
	}

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

	var insertedIDs []string

	for rows.Next() {
		var upc string
		if err := rows.Scan(&upc); err != nil {
			g.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Action: "insertProducts",
				Error:  "Error scanning UPC",
			})
			return
		}
		insertedIDs = append(insertedIDs, upc)
	}

	// Respond with success and the details of the insert
	g.JSON(http.StatusOK, models.SuccessDataResponse{
		Action:   "uploadInfo Successful",
		Inserted: insertedIDs,
	})
}

func buildInsertQuery(item []map[string]interface{}, platform string) (string, []interface{}, error) {
	var queryBuilder strings.Builder

	switch platform {
	case "omnibus":

		queryBuilder.WriteString(`
		INSERT INTO omnibus (
			upc, name, price, version, pagecount, datecreated, publisher, imgpath, 
			isturl, amazonurl, lastupdated, status
		) 
		VALUES
		`)

		values := []interface{}{}

		for i, row := range item {
			omnibus := models.Omnibus{
				UPC:       row["upc"].(string),
				Name:      row["name"].(string),
				Price:     float32(row["price"].(float64)),
				Version:   row["version"].(string),
				PageCount: int(row["pagecount"].(float64)),
				Publisher: row["publisher"].(string),
				ImgPath:   row["imgpath"].(string),
				ISTUrl:    row["isturl"].(string),
				AmazonUrl: row["amazonurl"].(string),
				Status:    row["status"].(string),
			}

			if i > 0 {
				queryBuilder.WriteString(",")
			}
			queryBuilder.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				len(values)+1, len(values)+2, len(values)+3, len(values)+4, len(values)+5, len(values)+6, len(values)+7,
				len(values)+8, len(values)+9, len(values)+10, len(values)+11, len(values)+12))

			// Append the values to the values slice
			values = append(values, omnibus.UPC, omnibus.Name, omnibus.Price, omnibus.Version, omnibus.PageCount, omnibus.DateCreated,
				omnibus.Publisher, omnibus.ImgPath, omnibus.ISTUrl, omnibus.AmazonUrl, omnibus.LastUpdated, omnibus.Status)
		}

		queryBuilder.WriteString(`
			ON CONFLICT (upc)
			DO NOTHING
			RETURNING upc;
		`)

		return queryBuilder.String(), values, nil

	default:
		return "", []interface{}{}, fmt.Errorf("Empty List")
	}
}
