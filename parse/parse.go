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

func ParseMarkdownToBook(filename string) *models.Book {
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

	scanner := bufio.NewScanner(file)

	// Parsing content line by line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip blank lines
		}

		// Go through the reArray and check for matches
		switch {
		case reChapter.MatchString(line):
			// Handle chapters
		case reSection.MatchString(line):
			// Handle sections
		case reDescription.MatchString(line):
			// Handle descriptions
		case rePoint.MatchString(line):
			// Handle points
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	return book
}
