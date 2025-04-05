package parse

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/VernRussell/merge-outline/models"
)

var (
	reChapter      = regexp.MustCompile(`^### \*\*(\d+)(?:\\\.)? (.+)\*\*`)    // Chapter regex
	reSection      = regexp.MustCompile(`^(\d+\.\d+) (.+)|#### \*\*(.*?)\*\*`) // Section regex
	reDescription  = regexp.MustCompile(`\*\*(.*?)\*\*`)                       // Description regex
	rePoint        = regexp.MustCompile(`^\* (.+)`)                            // Point regex
	reIntroduction = regexp.MustCompile(`^### \*\*Introduction\*\*`)           // Introduction regex
	reConclusion   = regexp.MustCompile(`^### \*\*Conclusion\*\*`)             // Conclusion regex
)

// ParseMarkdownToBook parses the markdown file and returns a Book object
func ParseMarkdownToBook(filename string) *models.Book {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Initialize the Book object
	book := &models.Book{
		Chapters: []models.Chapter{},
	}

	var currentChapter *models.Chapter
	var currentSection *models.Section
	var currentDescription *models.Description

	// Declare variables to track chapter, section, description numbers, and mode
	//sectionNum := 0 // For tracking section number
	//descNum := 0    // For tracking description number
	pointNum := 0 // For tracking point number
	//mode := ""      // Parsing mode (empty to start)
	//chapterNum := 0 // For tracking chapter number

	scanner := bufio.NewScanner(file)

	// Parsing content line by line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip blank lines
		}

		// Check for matches and process accordingly
		switch {
		case reChapter.MatchString(line):
			// New chapter found
			matches := reChapter.FindStringSubmatch(line)

			// Check if the regex matched the expected number of elements
			if len(matches) >= 3 {
				currentChapter = &models.Chapter{
					ChapterNumber:  matches[1],
					OriginalNumber: matches[1],
					Title:          matches[2],
				}
				//sectionNum = 1 // Reset section number when starting a new chapter
			} else {
				log.Printf("Skipping invalid chapter line: %s\n", line)
			}

		case reSection.MatchString(line):
			// New section found
			matches := reSection.FindStringSubmatch(line)

			// Ensure the regex captures the expected elements
			if len(matches) >= 3 {
				currentSection = &models.Section{
					SectionNumber: matches[1],
					SectionTitle:  matches[2],
				}
				currentChapter.Sections = append(currentChapter.Sections, *currentSection)
				pointNum = 1 // Reset point number when starting a new section
			} else {
				log.Printf("Skipping invalid section line: %s\n", line)
			}

		case reDescription.MatchString(line):
			// New description found
			matches := reDescription.FindStringSubmatch(line)

			// Ensure the regex captures the expected elements
			if len(matches) >= 2 {
				currentDescription = &models.Description{
					DescriptionNumber: matches[1],
					DescriptionHeader: matches[2],
				}
				currentSection.Descriptions = append(currentSection.Descriptions, *currentDescription)
			} else {
				log.Printf("Skipping invalid description line: %s\n", line)
			}

		case rePoint.MatchString(line):
			// New point found
			matches := rePoint.FindStringSubmatch(line)
			if len(matches) >= 2 {
				currentDescription.Points = append(currentDescription.Points, models.Point{
					PointNumber: fmt.Sprintf("%d", pointNum),
					PointText:   matches[1],
				})
				pointNum++ // Increment point number
			} else {
				log.Printf("Skipping invalid point line: %s\n", line)
			}

		case reIntroduction.MatchString(line):
			// Handle Introduction section
			book.Intro = &models.Extra{
				Title: "Introduction",
			}

		case reConclusion.MatchString(line):
			// Handle Conclusion section
			book.Conclusion = &models.Extra{
				Title: "Conclusion",
			}
		}
	}

	// Add the last chapter if any exists
	if currentChapter != nil {
		book.Chapters = append(book.Chapters, *currentChapter)
	}

	// Check for errors while reading the file
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	return book
}
