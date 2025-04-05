// main.go
package main

import (
	"log"
	"os"

	"github.com/VernRussell/merge-outline/extract"
	"github.com/VernRussell/merge-outline/utils" // Adjust import path
	// Adjust import path
)

func main() {
	// Provide the path to your JSON file
	inputFile := "./ContainerGardening.json"
	outputFile := "./ContainerGardeningUpdated.json"

	// Load the book data
	book, err := utils.LoadBook(inputFile)
	if err != nil {
		log.Fatalf("Error loading book: %v", err)
	}

	// Create a logger for output
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Merge duplicate chapters
	extract.MergeDuplicateChapters(book, logger)

	// Remove duplicate descriptions based on the header
	extract.RemoveDuplicateDescriptions(book, logger)

	// Write the updated book to a JSON file
	err = utils.WriteBookToJson(book, outputFile)
	if err != nil {
		log.Fatalf("Error writing book to JSON: %v", err)
	}

	// Extract description numbers and headers and print them
	extract.ExtractDescriptionNumbersAndHeaders(book)

	// Optionally log success
	logger.Println("Book successfully updated and written to JSON.")
}
