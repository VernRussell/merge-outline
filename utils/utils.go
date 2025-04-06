package utils

import (
	"bufio"
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
func FuzzySimilarity(book *models.Book, logger *log.Logger, frequentWords map[string]int, s1, s2 string) float64 {
	// Use JaroWinkler or any other fuzzy matching algorithm
	return smetrics.JaroWinkler(s1, s2, 0.7, 4) // Adjust thresholds if needed
}

// Function to merge duplicate chapters by fuzzy matching their titles
func MergeDuplicateChapters(book *models.Book, logger *log.Logger, threshold float64, frequentWords map[string]int) {
	// Loop through the chapters to find duplicates
	for i := 0; i < len(book.Chapters); i++ {
		originalChapter := &book.Chapters[i]

		// Compare this chapter with all subsequent chapters
		for j := i + 1; j < len(book.Chapters); j++ {
			duplicateChapter := &book.Chapters[j]

			// Perform fuzzy comparison of the chapter titles
			similarity := FuzzySimilarity(book, logger, frequentWords, originalChapter.Title, duplicateChapter.Title)

			if similarity > threshold {
				// Log the merge process
				logger.Printf("Merging Chapter %s (%s) with Chapter %s (%s) due to fuzzy match: %.2f\n",
					originalChapter.ChapterNumber, originalChapter.Title, duplicateChapter.ChapterNumber, duplicateChapter.Title, similarity)

				// Move sections from duplicate chapter to the original chapter
				// The sections from the duplicate chapter are added to the original chapter without renumbering
				for _, duplicateSection := range duplicateChapter.Sections {
					originalChapter.Sections = append(originalChapter.Sections, duplicateSection)

					// Renumber descriptions and points in each section, but keep the original section numbers
					for k := 0; k < len(duplicateSection.Descriptions); k++ {
						duplicateSection.Descriptions[k].DescriptionNumber = fmt.Sprintf("%s.%d", duplicateSection.SectionNumber, k+1)

						// Renumber points in the description if necessary
						for p := 0; p < len(duplicateSection.Descriptions[k].Points); p++ {
							duplicateSection.Descriptions[k].Points[p].PointNumber = fmt.Sprintf("%s.%d", duplicateSection.Descriptions[k].DescriptionNumber, p+1)
						}
					}
				}

				// Remove the duplicate chapter
				book.Chapters = append(book.Chapters[:j], book.Chapters[j+1:]...)
				j-- // Adjust the index after removal
			}
		}
	}
}

// discardFuzzyMatchedSections discards sections that match with other sections based on fuzzy similarity
// Function to discard fuzzy matched sections within a chapter of the book
func DiscardFuzzyMatchedSections(book *models.Book, logger *log.Logger, frequentWords map[string]int) {
	// Iterate through each chapter in the book
	for chapterIdx := range book.Chapters {
		chapter := &book.Chapters[chapterIdx]
		sections := chapter.Sections
		remainingSections := []models.Section{}  // Will hold sections that are not discarded
		processed := make([]bool, len(sections)) // Track which sections are processed (discarded)

		// Iterate through each section within the current chapter
		for i := 0; i < len(sections); i++ {
			if processed[i] {
				continue // Skip already processed (discarded) sections
			}

			currentSection := sections[i]

			// Compare current section to all other sections within this chapter
			for j := i + 1; j < len(sections); j++ {
				if processed[j] {
					continue // Skip already processed (discarded) sections
				}

				// Perform fuzzy comparison of the section titles (using SectionTitle)
				titleSimilarity := FuzzySimilarity(book, logger, frequentWords, currentSection.SectionTitle, sections[j].SectionTitle)

				// If the sections are similar enough, discard the second one
				if titleSimilarity > 0.8 { // You can adjust the threshold value here
					// Log that the section is being discarded
					//logger.Printf("Discarding Section %s (%s) due to fuzzy match with Section %s (%s): %.2f\n",
					//	currentSection.SectionNumber, currentSection.SectionTitle,
					//	sections[j].SectionNumber, sections[j].SectionTitle, titleSimilarity)

					// Mark the second section as processed (discarded)
					processed[j] = true
				}
			}

			// Add the non-discarded section to the list
			remainingSections = append(remainingSections, currentSection)
		}

		// Update the chapter with the remaining sections (those not discarded)
		chapter.Sections = remainingSections
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

// Function to compare two files and report differences to a log file
func CompareFiles(file1, file2, logFile string) error {
	// Open the log file for appending, create it if it doesn't exist
	logFileHandle, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}
	defer logFileHandle.Close()

	// Set up the logger to write to the log file
	logger := log.New(logFileHandle, "DIFFERENCE: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Open the first file
	f1, err := os.Open(file1)
	if err != nil {
		return fmt.Errorf("error opening file1: %v", err)
	}
	defer f1.Close()

	// Open the second file
	f2, err := os.Open(file2)
	if err != nil {
		return fmt.Errorf("error opening file2: %v", err)
	}
	defer f2.Close()

	// Create scanners for both files to read line by line
	scanner1 := bufio.NewScanner(f1)
	scanner2 := bufio.NewScanner(f2)

	lineNumber := 1
	differencesFound := false

	// Compare the files line by line
	for scanner1.Scan() || scanner2.Scan() {
		// Read the current line from each file
		line1 := scanner1.Text()
		line2 := scanner2.Text()

		// If one file has a line while the other doesn't
		if scanner1.Err() != nil || scanner2.Err() != nil {
			return fmt.Errorf("error reading files: %v", err)
		}

		// If the lines are different
		if line1 != line2 {
			differencesFound = true
			// Log the differences to the log file
			logger.Printf("Line %d:\n", lineNumber)
			if line1 != "" {
				logger.Printf("  File1: %s\n", line1)
			} else {
				logger.Println("  File1: <No content>")
			}
			if line2 != "" {
				logger.Printf("  File2: %s\n", line2)
			} else {
				logger.Println("  File2: <No content>")
			}
			logger.Println()
		}
		lineNumber++
	}

	// If differences are found, return an error
	if differencesFound {
		return fmt.Errorf("the files are not identical")
	}

	// If no differences, return nil (no error)
	fmt.Println("The files are identical.")
	return nil
}
