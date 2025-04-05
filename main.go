package main

import (
	"log"
	"os"

	"github.com/VernRussell/merge-outline/parse"
	"github.com/VernRussell/merge-outline/utils"
)

func main() {
	// Parsing the markdown
	inputFile := "ContainerGardening.md"
	outputFile := "ContainerGardening.json"
	book := parse.ParseMarkdownToBook(inputFile)

	// Create logger
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Calculate frequent words map for fuzzy comparison
	frequentWords := utils.CalculateFrequentWords(book)

	// Merge duplicate chapters
	mergeDuplicateChapters(book, logger, 0.8, frequentWords)

	// Merge fuzzy sections
	discardFuzzyMatchedSections(book, logger, frequentWords)

	// Remove duplicate descriptions
	removeDuplicateDescriptions(book, logger)

	// Renumber chapters and sections
	renumberChaptersAndSections(book)

	// Write updated book to JSON
	err := writeBookToJson(book, outputFile)
	if err != nil {
		log.Fatalf("Error writing book to JSON: %v", err)
	}

	// Regenerate the markdown file
	newMDFileName := "New_" + inputFile
	err = regenerateMdFile(book, newMDFileName)
	if err != nil {
		log.Fatalf("Error writing book to MD File: %v", err)
	}

	// Log success
	logger.Println("Book successfully updated and written to JSON.")
}
