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
	utils.MergeDuplicateChapters(book, logger, 0.8, frequentWords)

	// Merge fuzzy sections
	utils.DiscardFuzzyMatchedSections(book, logger, frequentWords)

	// Remove duplicate descriptions
	utils.RemoveDuplicateDescriptions(book, logger)

	// Renumber chapters and sections
	utils.RenumberChaptersAndSections(book)

	// Write updated book to JSON
	err := utils.WriteBookToJson(book, outputFile)
	if err != nil {
		log.Fatalf("Error writing book to JSON: %v", err)
	}

	// Regenerate the markdown file
	newMDFileName := "New_" + inputFile
	err = utils.RegenerateMdFile(book, newMDFileName)
	if err != nil {
		log.Fatalf("Error writing book to MD File: %v", err)
	}

	// Log success
	logger.Println("Book successfully updated and written to JSON.")
}
