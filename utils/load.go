// utils/load.go
package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/VernRussell/merge-outline/models" // Adjust import path
)

// Function to load the book data from a JSON file
func LoadBook(filename string) (*models.Book, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var book models.Book
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&book)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return &book, nil
}
