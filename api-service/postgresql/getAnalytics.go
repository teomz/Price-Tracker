package postgresql

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/teomz/Price-Tracker/api-service/models"
	"github.com/teomz/Price-Tracker/api-service/utilities"
)

// getAnalytics godoc
// @Summary get table
// @Description Return a table from the PostgreSQL database
// @Tags postgres
// @Accept json
// @Produce json
// @Param TaskUser query string true "User calling the API"
// @Param Table query string true "the table to query"
// @Success 200 {object} models.QuerySuccessResponse[models.Analytics]
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /postgresql/getAnalytics [get]
func getAnalytics(g *gin.Context) {

	action := "getAnalytics"

	var allowedQueryTypes = []string{"SELECT"}

	var allowedTables = []string{"analytics_amazon", "analytics_ist"}

	// Load the environment file
	envFile := "../.env"
	err := utilities.LoadEnvFile(envFile)
	if err != nil {
		// log.Println("Error occurred while loading .env file.")
	}

	// Check if the user is authorized
	if err := utilities.CheckUser(g, os.Getenv("AIRFLOW_USER")); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: action,
			Error:  "Wrong User",
		})
		return
	}

	// Create a DB instance
	db, err := createDBInstance()
	defer db.close()
	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: action,
			Error:  err.Error(),
		})
		return
	}

	table := g.DefaultQuery("Table", "default_table")

	query := fmt.Sprintf("SELECT * FROM %s", table)

	// Validate query
	if err := utilities.ValidateQuery(query, allowedQueryTypes, allowedTables); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: action,
			Error:  err.Error(),
		})
		return
	}

	// Execute the query and handle any errors
	rows, err := db.pool.Query(context.Background(), query)
	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: action,
			Error:  err.Error(),
		})
		return
	}
	defer rows.Close()

	if table == "analytics_amazon" || table == "analytics_ist" {
		var getTable []models.Analytics

		for rows.Next() {
			var analyticsRow models.Analytics
			if err := rows.Scan(&analyticsRow.Date, &analyticsRow.Name, &analyticsRow.LatestSale, &analyticsRow.Percent, &analyticsRow.Discount_five_Sale,
				&analyticsRow.Discount_five_Percent, &analyticsRow.Discount_ten_Sale, &analyticsRow.Discount_ten_Percent, &analyticsRow.URL); err != nil {
				g.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Action: action,
					Error:  "Error scanning analyticsRow",
				})
				return
			}
			getTable = append(getTable, analyticsRow)
		}

		// Respond with success and the details of the insert
		g.JSON(http.StatusOK, models.QuerySuccessResponse[models.Analytics]{
			Action: action + " Successful",
			Values: getTable,
		})
	}
}
