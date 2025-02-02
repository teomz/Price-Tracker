package postgresql

import (
	"context"
	"log"
	"net/http"
	"os"

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
// @Param Query body models.QueryRequest true "Array of items to upload"
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

	var req models.QueryRequest

	// Bind JSON request
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadInfo",
			Error:  "Invalid JSON format",
		})
		return
	}

	// Validate query
	if err := utilities.ValidateQuery(req.Query, allowedQueryTypes, allowedTables); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: "uploadInfo",
			Error:  err.Error(),
		})
		return
	}

	// Execute the query and handle any errors
	rows, err := db.pool.Query(context.Background(), req.Query, req.Values...)
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

// func contains(ids []string, item string) bool {
// 	for _, id := range ids {
// 		if id == item {
// 			return true
// 		}
// 	}
// 	return false
// }

// func buildInsertQuery(omnibus []models.Omnibus) (string, []interface{}) {
// 	var queryBuilder strings.Builder
// 	queryBuilder.WriteString(`
// 		INSERT INTO omnibus (UPC, code, name, price, version, pagecount, datecreated, publisher, imgpath, isturl, amazonurl, cgnurl, lastupdated, status)
// 		VALUES
// 	`)
// 	values := []interface{}{}

// 	for i, row := range omnibus {
// 		if i > 0 {
// 			queryBuilder.WriteString(",")
// 		}
// 		queryBuilder.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
// 			len(values)+1, len(values)+2, len(values)+3, len(values)+4, len(values)+5, len(values)+6, len(values)+7,
// 			len(values)+8, len(values)+9, len(values)+10, len(values)+11, len(values)+12, len(values)+13, len(values)+14))
// 		values = append(values, row.UPC, row.Code, row.Name, row.Price, row.Version, row.PageCount, row.DateCreated, row.Publisher,
// 			row.ImgPath, row.ISTUrl, row.AmazonUrl, row.CGNUrl, row.LastUpdated, row.Status)
// 	}

// 	queryBuilder.WriteString(`
// 		ON CONFLICT (UPC, CODE)
// 		DO NOTHING
// 		RETURNING UPC;
// 	`)

// 	return queryBuilder.String(), values
// }
