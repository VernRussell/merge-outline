package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/VernRussell/merge-outline/models"

	"github.com/xrash/smetrics"
)

// CalculateFrequentWords calculates frequent words in a Book object
func CalculateFrequentWords(book *models.Book) map[string]int {
	frequentWords := make(map[string]int)
	// Iterate through the book's chapters, sections, and descriptions to count words
	for _, chapter := range book.Chapters {
		for _, section := range chapter.Sections {
			for _, description := range section.Descriptions {
				words := strings.Fields(description.DescriptionHeader)
				for _, word := range words {
					frequentWords[word]++
				}
			}
		}
	}
	return frequentWords
}

// FuzzySimilarity calculates the similarity between two strings using the Jaro-Winkler distance
func FuzzySimilarity(s1, s2 string) float64 {
	return smetrics.JaroWinkler(s1, s2, 0.7, 4)
}

// mergeDuplicateChapters merges duplicate chapters based on a fuzzy similarity threshold
func MergeDuplicateChapters(book *models.Book, logger *log.Logger, threshold float64, frequentWords map[string]int) {
	// Logic to merge duplicate chapters based on fuzzy similarity
	logger.Println("Merging duplicate chapters based on fuzzy similarity...")
	// Example: iterate over the chapters and merge them based on similarity threshold
	for i := 0; i < len(book.Chapters); i++ {
		for j := i + 1; j < len(book.Chapters); j++ {
			if FuzzySimilarity(book.Chapters[i].Title, book.Chapters[j].Title) > threshold {
				// Merge chapters logic (you can adjust as needed)
				// For now, let's just log the matching titles
				logger.Printf("Merging chapters: %s and %s\n", book.Chapters[i].Title, book.Chapters[j].Title)
			}
		}
	}
}

// discardFuzzyMatchedSections discards sections that match with other sections based on fuzzy similarity
func DiscardFuzzyMatchedSections(book *models.Book, logger *log.Logger, frequentWords map[string]int) {
	// Logic to discard sections based on fuzzy matching
	logger.Println("Discarding fuzzy matched sections...")
	// Example: iterate through sections and discard duplicates
	for i := 0; i < len(book.Chapters); i++ {
		for j := i + 1; j < len(book.Chapters); j++ {
			// Implement fuzzy matching logic and discard based on similarity
		}
	}
}

// removeDuplicateDescriptions removes descriptions with duplicate content from a Book
func RemoveDuplicateDescriptions(book *models.Book, logger *log.Logger) {
	// Logic to remove duplicate descriptions
	logger.Println("Removing duplicate descriptions...")
	// Iterate over descriptions and check for duplicates
	for _, chapter := range book.Chapters {
		for _, section := range chapter.Sections {
			uniqueDescriptions := make(map[string]bool)
			var cleanedDescriptions []models.Description
			for _, description := range section.Descriptions {
				if _, exists := uniqueDescriptions[description.DescriptionHeader]; !exists {
					uniqueDescriptions[description.DescriptionHeader] = true
					cleanedDescriptions = append(cleanedDescriptions, description)
				}
			}
			section.Descriptions = cleanedDescriptions
		}
	}
}

// renumberChaptersAndSections renumbers chapters and sections after modifications
func RenumberChaptersAndSections(book *models.Book) {
	// Logic to renumber chapters and sections after modifications
	// Reset chapter and section numbers
	chapterCount := 1
	for i := range book.Chapters {
		book.Chapters[i].ChapterNumber = fmt.Sprintf("%d", chapterCount)
		sectionCount := 1
		for j := range book.Chapters[i].Sections {
			book.Chapters[i].Sections[j].SectionNumber = fmt.Sprintf("%d.%d", chapterCount, sectionCount)
			sectionCount++
		}
		chapterCount++
	}
}

// writeBookToJson writes the Book object to a JSON file
func WriteBookToJson(book *models.Book, filename string) error {
	// Create and open the file for writing
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Marshal the Book object into a formatted JSON string
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

// regenerateMdFile regenerates the markdown file from the Book object
func RegenerateMdFile(book *models.Book, filename string) error {
	// Logic to regenerate markdown from the Book object
	// For now, this is just a placeholder
	return nil
}
