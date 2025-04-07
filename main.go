package main

import (
	"fmt"
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

	blockedCombinations := utils.LoadBlockedCombinations("blocked_combinations.txt")

	fmt.Println(blockedCombinations)

	// Calculate frequent words map for fuzzy comparison
	frequentWords := utils.CalculateFrequentWords(book)

	utils.RemoveDuplicateChapters(book, logger, frequentWords, 0.9)

	utils.RenumberChaptersAndSections(book)

	utils.ListChaptersAndSections(book, "NoDupeChapters", nil)

	allSections := utils.SortSectionsByNumber(utils.CollectSections(book))

	fmt.Println(len(allSections))

	// Merge duplicate chapters
	duplicateChapters := utils.MergeDuplicateChapters(book, logger, 0.8, .7, frequentWords, blockedCombinations)

	fmt.Println(duplicateChapters)

	utils.ListChaptersAndSections(book, "MergedChapters", duplicateChapters)

	// Renumber chapters and sections
	utils.RenumberChaptersAndSections(book)

	utils.ListChaptersAndSections(book, "Final", nil)
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

	utils.ListChaptersAndSections(book, "Final", nil)

	// Log success
	logger.Println("Book successfully updated and written to JSON.")
}
