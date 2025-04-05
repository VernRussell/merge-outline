// extract/merge.go
package extract

import (
	"log"

	"github.com/VernRussell/merge-outline/models" // Adjust import path accordingly
)

// mergeDuplicateChapters merges chapters with the same title (or another criterion) to avoid duplication.
func MergeDuplicateChapters(book *models.Book, logger *log.Logger) {
	// Use a map to track chapters by their title (or other criterion)
	seenChapters := make(map[string]*models.Chapter)

	// Iterate through each chapter and merge duplicates
	var uniqueChapters []models.Chapter
	for _, chapter := range book.Chapters {
		// If the chapter title is already in the map, merge it
		if existingChapter, exists := seenChapters[chapter.Title]; exists {
			// Merge logic: e.g., append sections from the duplicate chapter to the existing one
			existingChapter.Sections = append(existingChapter.Sections, chapter.Sections...)
			// Optionally merge other fields (like descriptions) if needed
			existingChapter.Descriptions = append(existingChapter.Descriptions, chapter.Descriptions...)
		} else {
			// If the chapter is unique, add it to the map and the unique list
			seenChapters[chapter.Title] = &chapter
			uniqueChapters = append(uniqueChapters, chapter)
		}
	}

	// Update the book with the unique chapters
	book.Chapters = uniqueChapters

	// Log the result
	logger.Printf("Merged duplicate chapters. Total chapters: %d", len(book.Chapters))
}
