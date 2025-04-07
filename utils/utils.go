package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
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

// Function to collect all sections into a single list (with number and title)
func CollectSections(book *models.Book) []models.Section {
	var sections []models.Section
	// Iterate through chapters and sections
	for _, chapter := range book.Chapters {
		for _, section := range chapter.Sections {
			sectionInfo := models.Section{
				SectionNumber: section.SectionNumber,
				SectionTitle:  section.SectionTitle,
			}
			sections = append(sections, sectionInfo)
		}
	}
	return sections
}

// Function to sort sections by SectionNumber (numerically)
func SortSectionsByNumber(sections []models.Section) []models.Section {
	// Sort sections by SectionNumber numerically
	sort.Slice(sections, func(i, j int) bool {
		// Split the section numbers into parts (e.g., "16.1" -> [16, 1])
		sectionPartsI := strings.Split(sections[i].SectionNumber, ".")
		sectionPartsJ := strings.Split(sections[j].SectionNumber, ".")

		// Convert to integer for proper numerical comparison
		numI, _ := strconv.Atoi(sectionPartsI[0])
		numJ, _ := strconv.Atoi(sectionPartsJ[0])

		// First compare the major number (e.g., 16 vs. 17)
		if numI != numJ {
			return numI < numJ
		}

		// If major numbers are the same, compare the minor numbers (e.g., 1 vs. 2)
		minorI, _ := strconv.Atoi(sectionPartsI[1])
		minorJ, _ := strconv.Atoi(sectionPartsJ[1])

		return minorI < minorJ
	})

	return sections
}

// Function to print chapters and their sections
// Function to print chapters and their sections, writing clean output to a .txt file
func ListChaptersAndSections(book *models.Book) {
	// Open the file for writing (create it if it doesn't exist)
	logFile, err := os.OpenFile("ChaptersAndSections.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v", err)
		return
	}
	defer logFile.Close()

	// Iterate through the chapters in the book
	for _, chapter := range book.Chapters {
		// Write the chapter number and title to the file
		fmt.Fprintf(logFile, "Chapter %s: %s\n", chapter.ChapterNumber, chapter.Title)

		// Iterate through the sections in the current chapter
		for _, section := range chapter.Sections {
			// Write the section number and title to the file, slightly indented
			fmt.Fprintf(logFile, "    Section %s: %s\n", section.SectionNumber, section.SectionTitle)
		}

		// Add an extra line between chapters for better readability
		fmt.Fprintln(logFile)
	}
}

// Function to print chapters and their sections, excluding duplicates
func ListChaptersAndSectionsWithoutDuplicates(book *models.Book, sectionMap map[string]models.SectionState) {
	// Open the file for writing (create it if it doesn't exist)
	logFile, err := os.OpenFile("ChaptersAndSectionsWithoutDuplicates.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v", err)
		return
	}
	defer logFile.Close()

	// Iterate through the chapters in the book
	for _, chapter := range book.Chapters {
		// Write the chapter number and title to the file
		fmt.Fprintf(logFile, "Chapter %s: %s\n", chapter.ChapterNumber, chapter.Title)

		// Iterate through the sections in the current chapter
		for _, section := range chapter.Sections {
			// Get the state of the current section from the sectionMap
			sectionState, exists := sectionMap[section.SectionNumber]
			if !exists {
				// If the section is not in the map, consider it unique by default
				sectionState = models.SectionState{
					Section: section,
					State:   "unique", // Default to "unique" if not marked
				}
			}

			// Only write the section if it is not marked as a duplicate
			if sectionState.State != "duplicate" {
				// Write the section number and title to the file, slightly indented
				fmt.Fprintf(logFile, "    Section %s: %s\n", section.SectionNumber, section.SectionTitle)
			}
		}

		// Add an extra line between chapters for better readability
		fmt.Fprintln(logFile)
	}
}
