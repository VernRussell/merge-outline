// extract/extract.go
package extract

import (
	"fmt"
	"log"

	"github.com/VernRussell/project-name/models" // Adjust import path
	"github.com/xrash/smetrics"
)

// Function to extract and print description numbers and headers
func ExtractDescriptionNumbersAndHeaders(book *models.Book) {
	for _, chapter := range book.Chapters {
		for _, section := range chapter.Sections {
			for _, description := range section.Descriptions {
				// Print description number and header
				fmt.Printf("%s: %s\n", description.DescriptionNumber, description.DescriptionHeader)
			}
		}
	}
}

// Fuzzy comparison function using Jaro-Winkler or other algorithm
func FuzzySimilarity(s1, s2 string) float64 {
	return smetrics.JaroWinkler(s1, s2, 0.7, 4) // Adjust thresholds if needed
}

// Function to remove duplicate descriptions based on DescriptionHeader
func RemoveDuplicateDescriptions(book *models.Book, logger *log.Logger) {
	for _, chapter := range book.Chapters {
		for _, section := range chapter.Sections {
			uniqueDescriptions := []models.Description{}
			seenDescriptions := make(map[string]bool)

			// Iterate through each description in the section
			for _, description := range section.Descriptions {
				if !seenDescriptions[description.DescriptionHeader] {
					seenDescriptions[description.DescriptionHeader] = true
					uniqueDescriptions = append(uniqueDescriptions, description)
				} else {
					logger.Printf("Removing duplicate description: %s\n", description.DescriptionHeader)
				}
			}

			// Update the section with the unique descriptions
			section.Descriptions = uniqueDescriptions
		}
	}
}
