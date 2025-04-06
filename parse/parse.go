package parse

import (
	"bufio"
	"encoding/json"
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
				log.Println("Processing: Introduction")
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
				log.Println("Processing: Conclusion")
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

// Function to regenerate markdown from the models.Book object
func RegenerateMdFile(book *models.Book, filename string) error {
	// Open or create the output markdown file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Write the title of the book with markdown formatting
	_, err = file.WriteString(fmt.Sprintf("### **%s**\n\n", book.Title))
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	// Write Introduction if it exists
	if book.Intro != nil {
		_, err = file.WriteString("### **Introduction**\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}

		// Write all descriptions in Introduction
		for _, desc := range book.Intro.Descriptions {
			err = WriteDescriptionToFile(file, desc)
			if err != nil {
				return err
			}
		}

		// Add a blank line after Introduction before Chapter 1 starts
		_, err = file.WriteString("\n") // Adds an extra blank line after Introduction
		if err != nil {
			return fmt.Errorf("error writing blank line after Introduction: %v", err)
		}
	}

	// Write all chapters
	for _, chapter := range book.Chapters {
		// Write the chapter title (for example, Chapter 2)
		_, err = file.WriteString(fmt.Sprintf("### **%s\\. %s**\n", chapter.ChapterNumber, chapter.Title))
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}

		// Add a blank line after the chapter title before the first section
		_, err = file.WriteString("\n") // Blank line between chapter and first section
		if err != nil {
			return fmt.Errorf("error writing blank line after chapter title: %v", err)
		}

		// Write all sections within the chapter
		for _, section := range chapter.Sections {
			// Write the section title
			_, err = file.WriteString(fmt.Sprintf("#### **%s**\n", section.SectionTitle))
			if err != nil {
				return fmt.Errorf("error writing section title: %v", err)
			}

			// Write all descriptions within the section
			for _, desc := range section.Descriptions {
				err = WriteDescriptionToFile(file, desc)
				if err != nil {
					return err
				}
			}

			// Add a blank line after descriptions in the section (optional, depending on the structure)
			_, err = file.WriteString("\n") // Blank line after section descriptions (if desired)
			if err != nil {
				return fmt.Errorf("error writing blank line after section descriptions: %v", err)
			}
		}
	}

	// Write Conclusion if it exists
	if book.Conclusion != nil {
		_, err = file.WriteString("### **Conclusion**\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}

		// Write all descriptions in Conclusion
		for _, desc := range book.Conclusion.Descriptions {
			err = WriteDescriptionToFile(file, desc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper function to write a description to the markdown file
// Function to write a description to the markdown file
func WriteDescriptionToFile(file *os.File, description models.Description) error {
	// Add a blank line before the description
	_, err := file.WriteString("\n")
	if err != nil {
		return fmt.Errorf("error writing blank line before description: %v", err)
	}

	// Write the description header and text on the same line
	_, err = file.WriteString(fmt.Sprintf("**%s**: %s", description.DescriptionHeader, description.DescriptionText))
	if err != nil {
		return fmt.Errorf("error writing description header and text: %v", err)
	}

	// Add a blank line after the description before points
	_, err = file.WriteString("\n") // This adds the blank line after description and before points
	if err != nil {
		return fmt.Errorf("error writing blank line after description: %v", err)
	}

	// Adds an extra blank line before Points
	_, err = file.WriteString("\n") // Adds an extra blank line before Points

	// Write all points in the description
	for _, point := range description.Points {
		// Write the point, starting with "* " and no extra indentation
		_, err = file.WriteString(fmt.Sprintf("* %s\n", point.PointText))
		if err != nil {
			return fmt.Errorf("error writing point: %v", err)
		}
	}

	return nil
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
