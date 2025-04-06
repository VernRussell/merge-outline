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
	// Open file and read content
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	book := &models.Book{
		Chapters: []models.Chapter{},
	}

	var currentChapter *models.Chapter
	var currentSection *models.Section
	var currentDescription *models.Description
	sectionNum := 0
	descNum := 0
	pointNum := 0
	mode := ""      // Initialize mode variable (empty for now)
	chapterNum := 0 // Initialize chapterNum variable
	//	lastType := ""  // Track the last type processed
	//isInIntroduction := false // Flag to track if we're in Introduction or Conclusion

	// Define the re-Array with all the regex patterns
	reArray := []struct {
		name  string
		regex *regexp.Regexp
	}{
		{"Introduction", reIntroduction},
		{"Conclusion", reConclusion},
		{"Chapter", reChapter},
		{"Section", reSection},
		{"Description", reDescription},
		{"Point", rePoint},
	}

	scanner := bufio.NewScanner(file)

	// Parsing content line by line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip blank lines
		}

		// Go through the reArray and check for matches
		for _, r := range reArray {
			if r.regex.MatchString(line) {
				// When a match is found, report where we are
				fmt.Printf("\nMatched: %s\n", r.name)
				// Reset the relevant objects to nil before processing the new match
				switch r.name {
				case "Chapter":
					// Reset Chapter-related state
					currentSection = nil
					currentDescription = nil
					sectionNum = 0
					descNum = 0
					pointNum = 0
				case "Section":
					// Reset Section-related state
					currentDescription = nil
					descNum = 0
					pointNum = 0
				case "Description":
					// Reset Description-related state
					pointNum = 0
				}
				// Break out once a match is found
				break
			}
		}

		// Process line based on mode
		switch {
		case reIntroduction.MatchString(line):
			if mode != "intro" {
				mode = "intro"
				if book.Intro == nil {
					book.Intro = &models.Extra{
						Title: "Introduction",
					}
				}
				fmt.Println("Processing: Introduction")
			}
			//isInIntroduction := true
			continue

		case reConclusion.MatchString(line):
			if mode != "conclusion" {
				mode = "conclusion"
				if book.Conclusion == nil {
					book.Conclusion = &models.Extra{
						Title: "Conclusion",
					}
					currentSection = nil
				}
				fmt.Println("Processing: Conclusion")
			}
			//isInIntroduction = false
			continue

		case reChapter.MatchString(line):
			var matches []string
			matches = reChapter.FindStringSubmatch(line) // Find chapter title
			if matches != nil {
				mode = "chapter"
				chapterNum++
				chapterID := fmt.Sprintf("%d", chapterNum)
				chapter := models.Chapter{ChapterNumber: chapterID, OriginalNumber: matches[1], Title: matches[2]}
				book.Chapters = append(book.Chapters, chapter)
				currentChapter = &book.Chapters[len(book.Chapters)-1]
				currentSection = nil // Reset currentSection when a new chapter starts
				currentDescription = nil
				sectionNum = 0
				descNum = 0
				pointNum = 0
				fmt.Printf("Processing: Chapter %d - %s\n", chapterNum, matches[2])
			}
			continue

		case reSection.MatchString(line):
			var matches []string
			matches = reSection.FindStringSubmatch(line)
			if matches != nil && currentChapter != nil {
				sectionNum++
				sectionID := fmt.Sprintf("%s.%d", currentChapter.ChapterNumber, sectionNum)
				// Ensure that the title is properly captured here

				sectionTitle := matches[len(matches)-1] // Capture the section title (second capture group)
				if sectionTitle == "" {
					sectionTitle = matches[len(matches)-2]
				}

				section := models.Section{SectionNumber: sectionID, SectionTitle: sectionTitle}
				currentChapter.Sections = append(currentChapter.Sections, section)
				currentSection = &currentChapter.Sections[len(currentChapter.Sections)-1]
			}
			continue

		case reDescription.MatchString(line):
			var matches []string
			matches = reDescription.FindStringSubmatch(line)
			if matches != nil {
				descNum++
				descID := ""
				descriptionHeader := matches[1]
				index := strings.Index(line, descriptionHeader)
				descriptionText := line[index+len(descriptionHeader):]
				desc := models.Description{DescriptionNumber: "", DescriptionHeader: descriptionHeader, DescriptionText: descriptionText, Points: []models.Point{}}

				// Handle case where currentSection is nil (i.e., not within a Chapter/Section)
				if currentSection == nil {
					if mode == "intro" {
						// Add description to Introduction
						descID = fmt.Sprintf("Intro.%d", descNum)
						book.Intro.Descriptions = append(book.Intro.Descriptions, desc)
						currentDescription = &book.Intro.Descriptions[len(book.Intro.Descriptions)-1]
					} else if mode == "conclusion" {
						// Add description to Conclusion
						descID = fmt.Sprintf("C.%d", descNum)
						book.Conclusion.Descriptions = append(book.Conclusion.Descriptions, desc)
						currentDescription = &book.Conclusion.Descriptions[len(book.Conclusion.Descriptions)-1]
					}
				} else {
					// If currentSection is not nil, add description to Section
					descID = fmt.Sprintf("%s.%d", currentSection.SectionNumber, descNum)
					currentSection.Descriptions = append(currentSection.Descriptions, desc)
					currentDescription = &currentSection.Descriptions[len(currentSection.Descriptions)-1]
				}

				// Set the description ID
				currentDescription.DescriptionNumber = descID

				fmt.Printf("Processing: Description %d\n", descNum)
				fmt.Printf("Descriptions: %d ", descNum)

				// Reset point number after adding description
				pointNum = 0
			}
			continue

		case rePoint.MatchString(line):
			var matches []string
			matches = rePoint.FindStringSubmatch(line)
			if matches != nil && currentDescription != nil {
				pointNum++
				pointID := fmt.Sprintf("%s.%d", currentDescription.DescriptionNumber, pointNum)
				pt := models.Point{PointNumber: pointID, PointText: matches[1]}
				currentDescription.Points = append(currentDescription.Points, pt)

				fmt.Printf("Processing: Point %d\n", pointNum)
				fmt.Printf("Points: %d ", pointNum)
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	return book
}
