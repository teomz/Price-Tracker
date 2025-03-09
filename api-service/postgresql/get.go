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

// getInfoByDate godoc
// @Summary get data by date
// @Description Return an array of upc from the PostgreSQL database filtered by date
// @Tags postgres
// @Accept json
// @Produce json
// @Param TaskUser query string true "User calling the API"
// Date query string true "Date of item creation YYYY-MM-DD"
// @Success 200 {object} models.QuerySuccessResponse[string]
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /postgresql/getInfoByDate [get]
func getInfoByDate(g *gin.Context) {

	action := "getInfoByDate"

	var allowedQueryTypes = []string{"SELECT"}

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

	// req, ok := g.GetQuery("Date")

	// if !ok {
	// 	g.JSON(http.StatusBadRequest, models.ErrorResponse{
	// 		Action: action,
	// 		Error:  "Invalid Date Query",
	// 	})
	// 	return
	// }

	// _, err = time.Parse("2006-01-02", req)
	// if err != nil {
	// 	g.JSON(http.StatusBadRequest, models.ErrorResponse{
	// 		Action: action,
	// 		Error:  err.Error(),
	// 	})
	// 	return
	// }

	query := "SELECT upc FROM omnibus WHERE datecreated >= (SELECT MAX(datecreated) FROM omnibus) - INTERVAL '1 month' AND datecreated <= (SELECT MAX(datecreated) FROM omnibus)"

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

	var getIDs []string

	for rows.Next() {
		var upc string
		if err := rows.Scan(&upc); err != nil {
			g.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Action: action,
				Error:  "Error scanning UPC",
			})
			return
		}
		getIDs = append(getIDs, upc)
	}

	// Respond with success and the details of the insert
	g.JSON(http.StatusOK, models.QuerySuccessResponse[string]{
		Action: action + " Successful",
		Values: getIDs,
	})
}

// getInfoByPublisher godoc
// @Summary get data by date
// @Description Return an array of upc from the PostgreSQL database filtered by date
// @Tags postgres
// @Accept json
// @Produce json
// @Param TaskUser query string true "User calling the API"
// @Param Publisher query string true "Publisher of item"
// @Success 200 {object} models.QuerySuccessResponse[models.SaleUrls]
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /postgresql/getInfoByPublisher [get]
func getInfoByPublisher(g *gin.Context) {

	action := "getInfoByPublisher"

	var allowedQueryTypes = []string{"SELECT"}

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

	req, ok := g.GetQuery("Publisher")

	if !ok {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: action,
			Error:  "Invalid Publisher Query",
		})
		return
	}

	if err := utilities.IsAllowedPublisher(req); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: action,
			Error:  err.Error(),
		})
		return
	}

	query := `SELECT upc, amazonurl, isturl FROM omnibus WHERE publisher ILIKE '%' || $1 || '%'`

	// Validate query
	if err := utilities.ValidateQuery(query, allowedQueryTypes, allowedTables); err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: action,
			Error:  err.Error(),
		})
		return
	}

	// Execute the query and handle any errors
	rows, err := db.pool.Query(context.Background(), query, req)
	if err != nil {
		g.JSON(http.StatusBadRequest, models.ErrorResponse{
			Action: action,
			Error:  err.Error(),
		})
		return
	}
	defer rows.Close()

	var getLinks []models.SaleUrls

	for rows.Next() {
		var saleurls models.SaleUrls
		if err := rows.Scan(&saleurls.UPC, &saleurls.Amazon, &saleurls.IST); err != nil {
			g.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Action: action,
				Error:  "Error scanning UPC",
			})
			return
		}
		getLinks = append(getLinks, saleurls)
	}

	// Respond with success and the details of the insert
	g.JSON(http.StatusOK, models.QuerySuccessResponse[models.SaleUrls]{
		Action: action + " Successful",
		Values: getLinks,
	})
}
