// utils/utils.go
package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/VernRussell/merge-outline/models"
)

// Function to write the book data to a JSON file
func WriteBookToJson(book *models.Book, filename string) error {
	// Open (or create) the file to write the JSON data
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Marshal the book object into JSON format (use json.MarshalIndent for pretty printing)
	jsonData, err := json.MarshalIndent(book, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshalling book to JSON: %v", err)
	}

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing JSON to file: %v", err)
	}

	return nil
}
