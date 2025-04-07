package merge

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/VernRussell/merge-outline/models" // Adjust import path accordingly
	"github.com/VernRussell/merge-outline/utils"
	"github.com/xrash/smetrics"
)

// Function to mark sections as "Keep" or "Unchecked" based on the number of sections in the chapter
// and return the updated sectionMap
func markSectionsAsKeepOrUnchecked(sections []models.Section, book *models.Book) map[string]models.SectionState {
	// Initialize the sectionMap to hold section states
	sectionMap := make(map[string]models.SectionState)

	// Iterate through chapters in the book
	for _, chapter := range book.Chapters {
		// If the chapter has 4 or fewer sections, mark them as "Keep"
		if len(chapter.Sections) <= 4 {
			for _, section := range chapter.Sections {
				// Only mark as "Keep" if the section hasn't been marked already
				if state, exists := sectionMap[section.SectionNumber]; exists {
					// Skip if it's already marked as "Keep" or "Unchecked"
					if state.State != "" {
						continue
					}
				}

				sectionMap[section.SectionNumber] = models.SectionState{
					Section: section,
					State:   "Keep",
				}
			}
		} else {
			// Mark all sections in other chapters as "Unchecked"
			for _, section := range chapter.Sections {
				// Only mark as "Unchecked" if the section hasn't been marked already
				if state, exists := sectionMap[section.SectionNumber]; exists {
					// Skip if it's already marked as "Keep" or "Unchecked"
					if state.State != "" {
						continue
					}
				}

				sectionMap[section.SectionNumber] = models.SectionState{
					Section: section,
					State:   "Unchecked",
				}
			}
		}
	}

	// Return the sectionMap with updated states
	return sectionMap
}

// Function to find duplicates and mark them appropriately, preserving the "Keep" state
func findDuplicatesAndMark(book *models.Book, sectionMap map[string]models.SectionState, threshold float64, logger *log.Logger, frequentWords map[string]int) map[string]models.SectionState {
	// Flag for checking if a "Keep" section was part of the duplicate sections
	var maybeFlag bool

	// Iterate through the sections and compare
	for i, sectionState := range sectionMap {
		// Skip sections already marked as "Duplicate" or "Maybe"
		if sectionState.State == "Duplicate" || sectionState.State == "Maybe" {
			continue
		}

		// Compare the current section with all others
		for j, otherSectionState := range sectionMap {
			// Skip comparing to itself
			if i == j {
				continue
			}

			// Check if sections are marked as "Unchecked" or "Keep"
			if sectionState.State == "Unchecked" || sectionState.State == "Keep" {
				// Check for fuzzy similarity between the sections
				similarity := fuzzySectionSimilarity(book, logger, frequentWords, sectionMap[i].Section.SectionTitle, sectionMap[j].Section.SectionTitle)

				// If similarity exceeds threshold, mark as "Duplicate"
				if similarity >= threshold {
					if sectionState.State == "Keep" || otherSectionState.State == "Keep" {
						maybeFlag = true
					}
					// Mark sections as "Duplicate"
					sectionMap[sectionState.Section.SectionNumber] = models.SectionState{
						Section: sectionState.Section,
						State:   "Duplicate",
					}
					sectionMap[otherSectionState.Section.SectionNumber] = models.SectionState{
						Section: otherSectionState.Section,
						State:   "Duplicate",
					}
				}
			}
		}
	}

	// If any section marked "Keep" was part of a duplicate, mark the chapter as "Maybe"
	if maybeFlag {
		for i, sectionState := range sectionMap {
			if sectionState.State == "Duplicate" {
				// Mark the chapter as "Maybe"
				sectionMap[i] = models.SectionState{
					Section: sectionState.Section,
					State:   "Maybe",
				}
			}
		}
	}

	// Return the sectionMap with updated states
	return sectionMap
}

// Function to perform fuzzy similarity between two n-grams with a threshold
func fuzzySectionSimilarity(book *models.Book, logger *log.Logger, filteredWords map[string]int, word1, word2 string) float64 {
	// Perform fuzzy comparison between the two words (n-grams)
	similarity := smetrics.JaroWinkler(word1, word2, 0.7, 4)

	// Log the similarity score for debugging
	//if similarity > 0.75 { // Only log matches above the threshold
	//	logger.Printf("Fuzzy Similarity: '%s' vs '%s' = %f\n", word1, word2, similarity)
	//}

	return similarity
}

// Function to sort sectionMap by SectionNumber and write to SortedSectionMap.txt
func sortSectionMapAndWriteToFile(sectionMap map[string]models.SectionState) error {
	// Open the file for writing (create it if it doesn't exist)
	file, err := os.OpenFile("SortedSectionMap.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Convert the map to a slice for sorting
	var sortedSections []models.SectionState
	for _, sectionState := range sectionMap {
		sortedSections = append(sortedSections, sectionState)
	}

	// Sort the slice based on SectionNumber (numerically)
	sort.Slice(sortedSections, func(i, j int) bool {
		// Split the section numbers into parts (e.g., "16.1" -> [16, 1])
		sectionPartsI := strings.Split(sortedSections[i].Section.SectionNumber, ".")
		sectionPartsJ := strings.Split(sortedSections[j].Section.SectionNumber, ".")

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

	// Write the sorted sections to the file
	for _, sectionState := range sortedSections {
		// Write section number, title, and state to the file
		_, err := fmt.Fprintf(file, "State: %s - Section %s: %s\n", sectionState.State, sectionState.Section.SectionNumber, sectionState.Section.SectionTitle)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	// Success
	return nil
}

// Function to list chapters and their sections with states ("Keep" or "Unchecked")
func ListChaptersAndSectionsWithState(book *models.Book, sectionMap map[string]models.SectionState) {
	// Open the file to write the results
	file, err := os.Create("ChaptersAndSectionsWithState.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Iterate through the chapters
	for _, chapter := range book.Chapters {
		// Write the chapter title
		_, err := file.WriteString(fmt.Sprintf("Chapter %s: %s\n", chapter.ChapterNumber, chapter.Title))
		if err != nil {
			log.Fatal(err)
		}

		// Sort sections by their SectionNumber
		sort.Slice(chapter.Sections, func(i, j int) bool {
			return chapter.Sections[i].SectionNumber < chapter.Sections[j].SectionNumber
		})

		// Write the sections for the current chapter with their states
		for _, section := range chapter.Sections {
			// Retrieve the state for the section from the sectionMap
			sectionState := sectionMap[section.SectionNumber]
			_, err := file.WriteString(fmt.Sprintf("    Section %s: %s - State: %s\n", sectionState.Section.SectionNumber, sectionState.Section.SectionTitle, sectionState.State))
			if err != nil {
				log.Fatal(err)
			}
		}

		// Add an extra line between chapters for better readability
		_, err = file.WriteString("\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Function to iterate through chapters and sections, mark sections and find duplicates
func ProcessChaptersAndSections(book *models.Book, threshold float64, logger *log.Logger, frequentWords map[string]int) {

	allSections := utils.SortSectionsByNumber(utils.CollectSections(book))

	// First, mark sections as "Keep" or "Unchecked" based on chapter section count
	sectionMap := markSectionsAsKeepOrUnchecked(allSections, book)

	sortSectionMapAndWriteToFile(sectionMap)

	//ListChaptersAndSectionsWithState(book, sectionMap)

	// Next, find duplicates and mark them as "Duplicate", and flag chapters with "Maybe"
	sectionMap = findDuplicatesAndMark(book, sectionMap, threshold, logger, frequentWords)

	ListChaptersAndSectionsWithState(book, sectionMap)

	sortSectionMapAndWriteToFile(sectionMap)

	// Output the chapters and sections
	utils.ListChaptersAndSectionsWithoutDuplicates(book, sectionMap)
}
