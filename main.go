package main

import (
	"fmt"
	"log"
	"os"

	"github.com/VernRussell/merge-outline/merge"
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

	chaptersToInclude := merge.GenerateChaptersToInclude(book)
	fmt.Println(chaptersToInclude)

	matchedChapters := merge.CompareChaptersWithFrequentWords(book, logger, frequentWords)
	fmt.Println(matchedChapters)

	// List of n-gram sizes you want to process (e.g., 2, 3, and 7)
	ngramSizes := []int{2, 3, 7}

	// Call the processChapters function to collect, compare, and print the n-grams
	merge.ProcessChapters(book, logger, ngramSizes, chaptersToInclude, frequentWords)

	// Remove duplicate descriptions
	merge.RemoveDuplicateDescriptions(book, logger)

	// Merge fuzzy sections
	merge.DiscardFuzzyMatchedSections(book, logger, frequentWords)

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

	utils.ListChaptersAndSections(book)

	// Log success
	logger.Println("Book successfully updated and written to JSON.")
}
