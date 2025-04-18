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

// Struct to store chapter number and title
type ChapterInfo struct {
	ChapterNumber string
	Title         string
}

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

func FuzzySimilarity(book *models.Book, logger *log.Logger, frequentWords map[string]int, s1, s2 string, threshold float64) float64 {
	// Log the chapter titles being compared
	logger.Printf("Comparing Chapter Titles: '%s' and '%s'\n", s1, s2)

	// Use JaroWinkler or any other fuzzy matching algorithm
	similarity := smetrics.JaroWinkler(s1, s2, 0.6, 2) // Adjust parameters

	// Log the similarity score
	logger.Printf("Calculated Similarity: %.2f (Threshold: %.2f)\n", similarity, threshold)

	// Log the result of the comparison
	if similarity >= threshold {
		logger.Printf("Match Found: Similarity %.2f is above threshold %.2f\n", similarity, threshold)
		return similarity
	} else {
		logger.Printf("No Match: Similarity %.2f is below threshold %.2f\n", similarity, threshold)
		return 0.0
	}
}

// Function to merge duplicate chapters by fuzzy matching their titles
// This now checks the blocked combinations before merging.
func MergeDuplicateChapters(book *models.Book, logger *log.Logger, firstThreshold, subsequentThreshold float64, frequentWords map[string]int, blockedCombinations map[string]bool) []string {
	var removedChapters []string // List to store removed chapters with their numbers and titles
	firstMatch := true           // Flag to track if it's the first match

	// Loop through the chapters to find duplicates
	for i := 0; i < len(book.Chapters); i++ {
		originalChapter := &book.Chapters[i]
		mergedCount := 0 // Track how many chapters have been merged into the original chapter

		// Compare this chapter with all subsequent chapters
		for j := i + 1; j < len(book.Chapters); j++ {
			duplicateChapter := &book.Chapters[j]

			// Skip merging if this combination is blocked
			if _, blocked := blockedCombinations[originalChapter.ChapterNumber+","+duplicateChapter.ChapterNumber]; blocked {
				logger.Printf("Skipping merge for blocked combination: Chapter %s and Chapter %s\n", originalChapter.ChapterNumber, duplicateChapter.ChapterNumber)
				continue
			}

			// Choose the threshold based on whether it's the first match or not
			var threshold float64
			if firstMatch {
				threshold = firstThreshold // Use the lower threshold for the first match
			} else {
				threshold = subsequentThreshold // Use the higher threshold for subsequent matches
			}

			// Perform fuzzy comparison of the chapter titles
			similarity := FuzzySimilarity(book, logger, frequentWords, originalChapter.Title, duplicateChapter.Title, threshold)

			// Log the similarity score for debugging
			logger.Printf("Comparing Chapter %s (%s) with Chapter %s (%s) - Similarity Score: %.2f (Threshold: %.2f)\n",
				originalChapter.ChapterNumber, originalChapter.Title,
				duplicateChapter.ChapterNumber, duplicateChapter.Title,
				similarity, threshold)

			// If the similarity score exceeds the threshold, merge the chapters
			if similarity > threshold {
				// Log the merge process and store the removed chapter
				logger.Printf("Merging Chapter %s (%s) with Chapter %s (%s) due to fuzzy match: %.2f\n",
					originalChapter.ChapterNumber, originalChapter.Title,
					duplicateChapter.ChapterNumber, duplicateChapter.Title, similarity)

				// Store the removed chapter with its number and title
				removedChapters = append(removedChapters, fmt.Sprintf("Chapter %s: %s", duplicateChapter.ChapterNumber, duplicateChapter.Title))

				// Move sections from duplicate chapter to the original chapter
				for _, duplicateSection := range duplicateChapter.Sections {
					originalChapter.Sections = append(originalChapter.Sections, duplicateSection)

					// Renumber descriptions and points in each section
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

				// Increment the count of merged chapters
				mergedCount++

				// If two chapters have been merged into this original chapter, move on to the next original chapter
				if mergedCount >= 2 {
					break // Exit the loop to move on to the next original chapter
				}
			}
			// After the first match, set the flag to false
			firstMatch = false
		}

		// If two chapters have been merged, stop processing the current chapter
		if mergedCount >= 2 {
			logger.Printf("Finished merging chapters for Chapter %s (%s), moving to the next chapter.\n", originalChapter.ChapterNumber, originalChapter.Title)
			continue // Move on to the next original chapter
		}
	}

	// Log the removed chapters for debugging
	logger.Printf("Removed Chapters: %v\n", removedChapters)

	// Sort the removed chapters by chapter number (if needed)
	sort.Strings(removedChapters)

	return removedChapters
}

// Function to remove duplicate chapters based on fuzzy matching their titles
// This function will now take a user-provided threshold for similarity and remove all sections associated with duplicate chapters.
func RemoveDuplicateChapters(book *models.Book, logger *log.Logger, frequentWords map[string]int, threshold float64) {
	// Loop through the chapters to find duplicates
	for i := 0; i < len(book.Chapters); i++ {
		originalChapter := &book.Chapters[i]

		// Compare this chapter with all subsequent chapters
		for j := i + 1; j < len(book.Chapters); j++ {
			duplicateChapter := &book.Chapters[j]

			// Perform fuzzy comparison of the chapter titles
			similarity := FuzzySimilarity(book, logger, frequentWords, originalChapter.Title, duplicateChapter.Title, threshold)

			// If the chapters are similar enough, remove the duplicate chapter and all of its sections
			if similarity >= threshold {
				// Log the removal process
				logger.Printf("Removing Duplicate Chapter %s (%s) because it matches Chapter %s (%s) with similarity: %.2f\n",
					duplicateChapter.ChapterNumber, duplicateChapter.Title, originalChapter.ChapterNumber, originalChapter.Title, similarity)

				// Remove the duplicate chapter and all associated sections and descriptions
				book.Chapters = append(book.Chapters[:j], book.Chapters[j+1:]...)
				j-- // Adjust the index after removal to prevent skipping a chapter
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

// Function to print chapters and their sections, writing clean output to a .txt file
// Accepts a slice of strings for duplicateChapters (merged chapters) to output them at the top if provided.
func ListChaptersAndSections(book *models.Book, suffix string, duplicateChapters []string) {
	// Create the file name with the suffix appended
	fileName := fmt.Sprintf("ChaptersAndSections_%s.txt", suffix)

	// Open the file for writing (create it from scratch)
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v", err)
		return
	}
	defer logFile.Close()

	// If duplicateChapters is provided (not nil or empty), output them at the top of the file
	if len(duplicateChapters) > 0 {
		fmt.Fprintln(logFile, "Merged Duplicate Chapters:")
		for _, chapter := range duplicateChapters {
			// Print each chapter in the format "Chapter {Number}: {Title}"
			fmt.Fprintln(logFile, chapter)
		}
		// Add a line break before listing the regular chapters
		fmt.Fprintln(logFile)
	}

	// Iterate through the chapters in the book and output them
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

// Function to load blocked chapter combinations from a file
func LoadBlockedCombinations(filePath string) map[string]bool {
	blockedCombinations := make(map[string]bool)

	// Open the blocked combinations file
	file, err := os.Open(filePath)
	if err != nil {
		// If the file does not exist (e.g., the first run), return an empty map
		if os.IsNotExist(err) {
			return blockedCombinations
		}
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines
		if line == "" {
			continue
		}

		// Assume the format is "ChapterX,ChapterY" (e.g., "3,17")
		parts := strings.Split(line, ",")
		if len(parts) == 2 {
			blockedCombinations[parts[0]+","+parts[1]] = true
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	return blockedCombinations
}

// Function to write blocked combinations to a file
func WriteBlockedCombinations(filePath string, blockedCombinations map[string]bool) {
	// Open the file for writing (create it from scratch)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error opening file for writing: %v\n", err)
		return
	}
	defer file.Close()

	// Write each blocked combination to the file
	for combination := range blockedCombinations {
		_, err := file.WriteString(combination + "\n")
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
		}
	}
}
