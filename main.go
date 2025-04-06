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
	//logFile := "comparison.log" // The log file where differences will be written

	book := parse.ParseMarkdownToBook(inputFile)

	// Regenerate the markdown file
	confirmMDFileName := "Confirm_" + inputFile
	err := parse.RegenerateMdFile(book, confirmMDFileName)
	if err != nil {
		log.Fatalf("Error writing book to MD File: %v", err)
	}

	// Compare the files
	//err = utils.CompareFiles(outputFile, confirmMDFileName, logFile)
	//if err != nil {
	//	log.Fatalf("File written back is not idential: %v", err)
	//}

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
	err = parse.WriteBookToJson(book, outputFile)
	if err != nil {
		log.Fatalf("Error writing book to JSON: %v", err)
	}

	// Regenerate the markdown file
	newMDFileName := "New_" + inputFile
	err = parse.RegenerateMdFile(book, newMDFileName)
	if err != nil {
		log.Fatalf("Error writing book to MD File: %v", err)
	}

	// Log success
	logger.Println("Book successfully updated and written to JSON.")
}
