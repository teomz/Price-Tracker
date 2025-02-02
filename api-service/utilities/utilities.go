package utilities

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	//ErrUserIDMissing        = "User ID is required"
	ErrNoFileUploaded       = "No file is uploaded"
	ErrFileOpenFailed       = "Failed to open file"
	ErrFileReadFailed       = "Failed to read file content"
	ErrInvalidFileExtension = "Invalid file extension %v. Only allowed extensions are: %v"
	ErrInvalidMimeType      = "Invalid MIME type. Allowed types are: %v"
	SuccessFileValid        = "File is valid"
	ErrInvalidUser          = "Invalid user"
)

// Validate_File validates the file upload based on extensions and MIME types
func Validate_File(g *gin.Context, ext_list []string, mime_list []string) (string, error) {
	// Convert the extension list and MIME list into maps for quick lookups
	validExts := make(map[string]bool)
	for _, ext := range ext_list {
		validExts[strings.ToLower(ext)] = true
	}

	validMimes := make(map[string]bool)
	for _, mime := range mime_list {
		validMimes[mime] = true
	}

	// Get the file extension and convert it to lowercase
	//ext := strings.ToLower(filepath.Ext(file.Filename))

	ext := g.DefaultPostForm("extension", "jpeg")

	// Check if the file extension is in the allowed list using map lookup
	if !validExts[ext] {
		return "", fmt.Errorf(ErrInvalidFileExtension, ext, ext_list)
	}

	// Open the file to check the MIME type
	_, fileContent, err := GetFile(g)
	if err != nil {
		return "", err
	}

	defer fileContent.Close()

	// Read the first 512 bytes for MIME type detection
	buf := make([]byte, 512)
	_, err = fileContent.Read(buf)
	if err != nil {
		return "", fmt.Errorf(ErrFileReadFailed)
	}

	// Detect the MIME type of the file
	mimeType := http.DetectContentType(buf)

	// Check if the MIME type is in the allowed list using map lookup
	if !validMimes[mimeType] {
		return "", fmt.Errorf(ErrInvalidMimeType, mime_list)
	}

	// If everything is valid, return a success response
	return SuccessFileValid, nil
}

func GetFile(g *gin.Context) (*multipart.FileHeader, multipart.File, error) {
	//Retrieve file from request
	file, err := g.FormFile("file")
	if err != nil {
		return nil, nil, fmt.Errorf(ErrNoFileUploaded)
	}

	fileContent, err := file.Open()
	if err != nil {
		return nil, nil, fmt.Errorf(ErrFileOpenFailed)
	}

	return file, fileContent, nil
}

func CheckUser(g *gin.Context, user_env string) error {
	userID := g.DefaultQuery("TaskUser", "default_user")
	if userID != user_env {
		return fmt.Errorf(ErrInvalidUser)
	}
	return nil
}

func LoadEnvFile(envFilePath string) error {
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		log.Printf("No .env file found at: %s\n", envFilePath)
		return nil // No error, just log and return
	} else {
		// Load the .env file
		err := godotenv.Load(envFilePath)
		if err != nil {
			log.Printf("Error loading .env file: %v", err)
			return err // Return the error if loading fails
		}
	}
	return nil
}

func ValidateQuery(query string, allowedQueryTypes []string, allowedTables []string) error {
	// Convert to uppercase for case-insensitive matching
	upperQuery := strings.ToUpper(strings.TrimSpace(query))

	// Ensure query starts with allowed types (basic SQL injection protection)
	valid := false
	for _, qType := range allowedQueryTypes {
		if strings.HasPrefix(upperQuery, qType) {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("invalid query type")
	}

	// Ensure the table name exists in the whitelist
	tableValid := false
	for _, table := range allowedTables {
		if strings.Contains(query, table) {
			tableValid = true
			break
		}
	}
	if !tableValid {
		return errors.New("unauthorized table access")
	}

	// Basic check to prevent dangerous operations
	if strings.Contains(upperQuery, "DROP TABLE") || strings.Contains(upperQuery, "DELETE FROM") {
		return errors.New("dangerous query detected")
	}

	return nil
}
