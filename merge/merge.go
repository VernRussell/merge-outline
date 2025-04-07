package merge

import (
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/VernRussell/merge-outline/models" // Adjust import path accordingly
	"github.com/VernRussell/merge-outline/utils"
	"github.com/xrash/smetrics"
)

// Full list of common stopwords to exclude
var stopwords = []string{
	"i", "you", "he", "she", "it", "we", "they",
	"me", "him", "her", "us", "them", "my", "your",
	"his", "her", "its", "our", "their", "mine", "yours",
	"hers", "ours", "theirs", "a", "an", "the", "and",
	"but", "for", "nor", "so", "yet", "to", "with", "about", "as", "at",
	"by", "from", "of", "on", "in", "out", "over", "under",
	"again", "further", "then", "once", "here", "there", "when",
	"where", "why", "how", "all", "any", "both", "each", "few",
	"more", "most", "other", "some", "such", "no", "nor", "not",
}

// Struct to hold n-grams and their counts
type Keyword struct {
	Pair  string
	Count int
}

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
				titleSimilarity := utils.FuzzySimilarity(book, logger, frequentWords, currentSection.SectionTitle, sections[j].SectionTitle, 0.8)

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

// Function to clean up and normalize chapter titles
func NormalizeTitle(title string) string {
	// Remove extra spaces, punctuation, and convert to lowercase
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	normalizedTitle := re.ReplaceAllString(strings.TrimSpace(title), " ")
	return strings.ToLower(normalizedTitle)
}

// Compare only titles to check if they are duplicates
func AreChaptersDuplicates(chap1, chap2 *models.Chapter) bool {
	// Compare only the normalized titles for exact matches
	return NormalizeTitle(chap1.Title) == NormalizeTitle(chap2.Title)
}

// Function to perform fuzzy matching and check if similarity is > 0.75
func AreChaptersFuzzyDuplicates(chap1, chap2 *models.Chapter) bool {
	// Calculate Jaro-Winkler distance between normalized titles
	// Threshold set to 0.75, prefixLength set to 4
	similarity := smetrics.JaroWinkler(NormalizeTitle(chap1.Title), NormalizeTitle(chap2.Title), 0.7, 4)
	return similarity > 0.75
}

// Sample cleanText function (removes stopwords and splits into words)
func CleanMWText(text string) []string {
	words := strings.Fields(text)
	var cleanWords []string
	for _, word := range words {
		// Remove punctuation and make the word lowercase for comparison
		word = strings.ToLower(strings.TrimSpace(word))
		if !IsStopword(word) {
			cleanWords = append(cleanWords, word)
		}
	}
	return cleanWords
}

// Helper function to check if a word is a stopword
func IsStopword(word string) bool {
	for _, stopword := range stopwords {
		if word == stopword {
			return true
		}
	}
	return false
}

// Function to create a map of chapters 16 and beyond
func GenerateChaptersToInclude(book *models.Book) []string {
	// Create a slice to store the ChapterNumbers
	var chaptersToInclude []string

	// Iterate over chapters starting from Chapter 16
	for i := 15; i < len(book.Chapters); i++ {
		chapter := &book.Chapters[i]
		chaptersToInclude = append(chaptersToInclude, chapter.ChapterNumber) // Append chapter number to the slice
	}
	return chaptersToInclude
}

// Instead of using the keywords slice, pass filteredWords.
func CompareChaptersWithFrequentWords(book *models.Book, logger *log.Logger, wordFrequency map[string]int) map[string]string {
	// Create a map to store the matched chapter titles with chapter number
	matchedChapters := make(map[string]string)

	// Iterate over chapters starting from chapter 16
	for i := 15; i < len(book.Chapters); i++ {
		chapter16 := &book.Chapters[i]
		wordsInChapter16 := cleanText(chapter16.Title)

		// Iterate over words in chapter 16
		for _, word := range wordsInChapter16 {
			// Check if the word exists in the wordFrequency map and the chapter hasn't been matched yet
			if wordFrequency[word] > 0 && !isInMap(matchedChapters, chapter16.Title) {
				// Iterate over the first 15 chapters to compare
				for j := 0; j < 15 && j < len(book.Chapters); j++ {
					chapter15 := &book.Chapters[j]
					wordsInChapter15 := cleanText(chapter15.Title)

					// Check if the word exists in chapter 15 and hasn't been matched yet
					for _, word15 := range wordsInChapter15 {
						if word == word15 && !isInMap(matchedChapters, chapter15.Title) {
							// Log the match between chapter 16 and chapter 15
							logger.Printf("Matching Chapters Found: Word '%s' in Chapter %s: %s and Chapter %s: %s\n",
								word, chapter16.ChapterNumber, chapter16.Title, chapter15.ChapterNumber, chapter15.Title)

							// Mark both chapters as matched by adding to the map
							matchedChapters[chapter16.Title] = chapter16.ChapterNumber
							matchedChapters[chapter15.Title] = chapter15.ChapterNumber
						}
					}
				}
			}
		}
	}

	return matchedChapters
}

// Helper function to check if a chapter title is already in the map
func isInMap(m map[string]string, chapterTitle string) bool {
	_, exists := m[chapterTitle]
	return exists
}

// Function to clean and split text into words
func cleanText(text string) []string {
	// Remove non-alphanumeric characters and convert to lowercase
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	text = re.ReplaceAllString(text, " ")
	return strings.Fields(text)
}

// Helper function to check if a chapter is already in the matchedChapters slice
func IsInSlice(slice []string, chapterTitle string) bool {
	for _, title := range slice {
		if title == chapterTitle {
			return true
		}
	}
	return false
}

// Function to process n-grams, compare fuzzy matches, and merge results
func ProcessChapters(book *models.Book, logger *log.Logger, ngramSizes []int, chaptersToInclude []string, frequentWords map[string]int) {
	// Collect n-grams for the filtered chapters
	chapterResults := collectKeywordNgrams(book, chaptersToInclude, ngramSizes)

	// Call the checkAndMergeChapters function to regress through chapters and merge
	checkAndMergeChapters(book, chapterResults, chaptersToInclude, frequentWords, logger)

	// Optionally, print the merged chapters (if needed)
	//printKeywordNgrams(book, chapterResults, chaptersToInclude, logger)
}

// Function to check and merge chapters
func checkAndMergeChapters(book *models.Book, chapterResults map[string][]Keyword, chaptersToInclude []string, frequentWords map[string]int, logger *log.Logger) {
	// Create a map to track which chapters have been merged
	chapterMerged := make(map[string]bool)

	// Create a slice to hold the chapters that need to be removed after merging
	chaptersToRemove := []int{}

	// Start with the last chapter in the list (e.g., Chapter 160)
	for chapterNumber := len(book.Chapters) - 1; chapterNumber >= 0; chapterNumber-- {
		chapter := &book.Chapters[chapterNumber]
		keywords := chapterResults[chapter.ChapterNumber]

		// Skip this chapter if it has already been merged or is not in the chaptersToInclude slice
		if chapterMerged[chapter.ChapterNumber] || !contains(chaptersToInclude, chapter.ChapterNumber) {
			continue
		}

		mergeFound := false

		// Iterate through all previous chapters (from the current chapter to Chapter 1)
		for i := 0; i < chapterNumber; i++ {
			targetChapter := &book.Chapters[i]
			targetKeywords := chapterResults[targetChapter.ChapterNumber]

			// Compare n-grams from the current chapter with previous chapters
			for _, k := range keywords {
				for _, targetK := range targetKeywords {
					// Fuzzy similarity (use a threshold of 0.8 similarity)
					similarity := fuzzySimilarity(book, logger, frequentWords, k.Pair, targetK.Pair)
					if similarity > 0.8 { // Threshold for fuzzy match
						// Log the merge before performing it
						logger.Printf("Merging Chapter %s (%s) with Chapter %s (%s)\n", chapter.ChapterNumber, chapter.Title, targetChapter.ChapterNumber, targetChapter.Title)

						// Merge the n-grams: Append the sections of the current chapter to the target chapter's sections
						targetChapter.Sections = append(targetChapter.Sections, chapter.Sections...) // Merge the sections

						// Mark both chapters as merged
						chapterMerged[chapter.ChapterNumber] = true
						chapterMerged[targetChapter.ChapterNumber] = true

						// Add the merged chapter to the removal list
						chaptersToRemove = append(chaptersToRemove, chapterNumber)

						mergeFound = true
						break // Exit the n-gram comparison loop once a merge is found
					}
				}

				if mergeFound {
					break // Exit the section loop once a merge is found
				}
			}

			if mergeFound {
				break // Exit the chapter comparison loop once a merge is found
			}
		}

		// After merging, check **subsequent chapters**, not previous ones
		if !chapterMerged[chapter.ChapterNumber] {
			continue // Skip the chapter if no merge was found
		}
	}

	// After the merge process, remove the chapters that were merged
	// Sort the chaptersToRemove in descending order to remove chapters safely
	sort.Sort(sort.Reverse(sort.IntSlice(chaptersToRemove)))

	// Remove the chapters from book.Chapters
	for _, chapterIndex := range chaptersToRemove {
		chapter := &book.Chapters[chapterIndex]
		logger.Printf("Removed Chapter %s (%s) after merging\n", chapter.ChapterNumber, chapter.Title) // Log with chapter number and title
		book.Chapters = append(book.Chapters[:chapterIndex], book.Chapters[chapterIndex+1:]...)
	}
}

// Collect n-grams for the chapters (doubles, triples, etc.)
// Function to check if a chapter is in the chaptersToInclude slice
func contains(chaptersToInclude []string, chapterNumber string) bool {
	for _, chapter := range chaptersToInclude {
		if chapter == chapterNumber {
			return true
		}
	}
	return false
}

func collectKeywordNgrams(book *models.Book, chaptersToInclude []string, ngramSizes []int) map[string][]Keyword {
	chapterResults := make(map[string][]Keyword)

	// Iterate over chapters
	for i := 0; i < len(book.Chapters); i++ {
		chapter := &book.Chapters[i]

		// Skip chapters that are not in the provided list (using contains function)
		if !contains(chaptersToInclude, chapter.ChapterNumber) {
			continue
		}

		wordsInChapter := cleanText(chapter.Title)

		// If the title is too short after cleaning, skip it
		if len(wordsInChapter) < 2 {
			continue
		}

		// Map to store n-grams (pairs, triples, etc.) and their counts for the chapter
		keywordCounts := make(map[string]int)

		// Collect n-grams based on dynamic sizes (from the ngramSizes list)
		for _, n := range ngramSizes { // Loop for each specified n-gram size
			for i := 0; i <= len(wordsInChapter)-n; i++ {
				var keyword string
				valid := true

				// Check if all words in the n-gram are non-stopwords
				for j := 0; j < n; j++ {
					if IsStopword(wordsInChapter[i+j]) {
						valid = false
						break
					}
					keyword += wordsInChapter[i+j] + " "
				}
				keyword = strings.TrimSpace(keyword)

				// If the keyword is valid (not containing stopwords), increment its count
				if valid && keyword != "" {
					keywordCounts[keyword]++
				}
			}
		}

		// Store the sorted n-grams and their counts for the chapter
		var keywords []Keyword
		for k, count := range keywordCounts {
			keywords = append(keywords, Keyword{Pair: k, Count: count})
		}

		// Sort keywords by count in descending order
		sort.Slice(keywords, func(i, j int) bool {
			return keywords[i].Count > keywords[j].Count
		})

		// Only add chapters with valid n-grams
		if len(keywords) > 0 {
			chapterResults[chapter.ChapterNumber] = keywords
		}
	}

	return chapterResults
}

// Function to compare n-grams between Chapters 1-15 and Chapters 117+ and merge them if fuzzy matches are found
func compareAndMergeNgrams(book *models.Book, chapterResults map[string][]Keyword, chaptersToInclude map[string]bool, frequentWords map[string]int, logger *log.Logger) {
	// Create a map to track which chapters have been merged
	chapterMerged := make(map[string]bool)

	// Iterate over n-grams from Chapters 117+ and compare with Chapters 1-15
	for chapterNumber, keywords := range chapterResults {
		// Only process Chapters 117+ (from Chapter 117 onward)
		if chapterNumber >= "117" {
			for _, k := range keywords {
				// First, compare with n-grams from Chapters 1-15
				var matched bool
				for targetChapterNumber, targetKeywords := range chapterResults {
					// Skip same chapters (i.e., don't compare chapter 117+ with itself)
					if targetChapterNumber == chapterNumber {
						continue
					}

					// Only compare with Chapters 1-15 first
					if targetChapterNumber >= "1" && targetChapterNumber <= "15" {
						for _, targetK := range targetKeywords {
							// Fuzzy similarity (use a threshold of 0.8 similarity)
							similarity := fuzzySimilarity(book, logger, frequentWords, k.Pair, targetK.Pair)
							if similarity > 0.8 { // Threshold of fuzzy match
								// Merge Chapter 117+ with Chapter 1-15
								if !chapterMerged[chapterNumber] {
									logger.Printf("Merging Chapter %s with Chapter %s\n", chapterNumber, targetChapterNumber)
									chapterResults[targetChapterNumber] = append(chapterResults[targetChapterNumber], keywords...)
									chapterMerged[chapterNumber] = true
								}
								matched = true
							}
						}
					}
				}

				// If no match is found with Chapters 1-15, compare with Chapter 125
				if !matched {
					if chapterNumber == "139" { // If Chapter 139 doesn't match in Chapters 1-15, try Chapter 125
						for targetChapterNumber, targetKeywords := range chapterResults {
							if targetChapterNumber == "125" { // Compare only with Chapter 125
								for _, targetK := range targetKeywords {
									similarity := fuzzySimilarity(book, logger, frequentWords, k.Pair, targetK.Pair)
									if similarity > 0.8 {
										// Merge Chapter 139 with Chapter 125
										if !chapterMerged[chapterNumber] {
											logger.Printf("Merging Chapter %s with Chapter %s\n", chapterNumber, targetChapterNumber)
											chapterResults[targetChapterNumber] = append(chapterResults[targetChapterNumber], keywords...)
											chapterMerged[chapterNumber] = true
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// Function to perform fuzzy similarity between two n-grams with a threshold
func fuzzySimilarity(book *models.Book, logger *log.Logger, filteredWords map[string]int, word1, word2 string) float64 {
	// Perform fuzzy comparison between the two words (n-grams)
	similarity := smetrics.JaroWinkler(word1, word2, 0.7, 4)

	// Log the similarity score for debugging
	//if similarity > 0.75 { // Only log matches above the threshold
	//	logger.Printf("Fuzzy Similarity: '%s' vs '%s' = %f\n", word1, word2, similarity)
	//}

	return similarity
}

// Function to find duplicates, marking sections as "good", "duplicate", or "unique"
// Function to find duplicates, marking sections as "good", "duplicate", or "unique"
func FindAndMarkSections(sections []models.Section, book *models.Book, logger *log.Logger, frequentWords map[string]int, threshold float64) map[string]models.SectionState {
	sectionMap := make(map[string]models.SectionState)
	//processed := make(map[string]bool) // Keeps track of sections already marked as "good" or "duplicate"

	// Sort sections by their section number in ascending order
	//sortSectionsByNumber(sections)

	// Iterate through the sections in sorted order
	for i := 0; i < len(sections); i++ {
		currentSection := &sections[i]
		// Initialize the section state to "unchecked"
		if _, exists := sectionMap[currentSection.SectionNumber]; !exists {
			sectionMap[currentSection.SectionNumber] = models.SectionState{
				Section: *currentSection,
				State:   "unchecked", // Default state
			}
		}

		// Skip sections that are already marked as "duplicate"
		if sectionMap[currentSection.SectionNumber].State == "duplicate" {
			continue
		}

		// If this section has not been processed yet, process it
		if sectionMap[currentSection.SectionNumber].State == "unchecked" {
			// Compare the current section to all previous sections in sorted order
			markedAsDuplicate := false
			for j := 0; j < i; j++ {
				otherSection := &sections[j]

				// Skip if already processed as "good" or "duplicate"
				if sectionMap[otherSection.SectionNumber].State == "duplicate" || sectionMap[otherSection.SectionNumber].State == "good" {
					continue
				}

				// Use FuzzySimilarity to check similarity between section titles
				similarity := fuzzySimilarity(book, logger, frequentWords, currentSection.SectionTitle, otherSection.SectionTitle)

				// Debugging: Print the similarity score for each comparison
				log.Printf("Comparing '%s' with '%s' | Similarity: %.2f\n", currentSection.SectionTitle, otherSection.SectionTitle, similarity)

				// If similarity exceeds the threshold, mark as "good" and "duplicate"
				if similarity >= threshold {
					// Mark the current section as "good" and the matched one as "duplicate"
					sectionMap[currentSection.SectionNumber] = models.SectionState{
						Section: *currentSection,
						State:   "good",
					}
					sectionMap[otherSection.SectionNumber] = models.SectionState{
						Section: *otherSection,
						State:   "duplicate",
					}
					markedAsDuplicate = true
					break // Exit inner loop after finding a match
				}
			}

			// If no match was found, mark it as "unique"
			if !markedAsDuplicate {
				sectionMap[currentSection.SectionNumber] = models.SectionState{
					Section: *currentSection,
					State:   "unique",
				}
			}
		}
	}

	// Convert the map to a slice to allow sorting
	var sortedSections []models.SectionState
	for _, sectionState := range sectionMap {
		sortedSections = append(sortedSections, sectionState)
	}

	// Sort the slice by SectionNumber (ascending) and then by State ("good", "duplicate", "unique")
	sort.Slice(sortedSections, func(i, j int) bool {
		// First compare by SectionNumber
		if sortedSections[i].Section.SectionNumber != sortedSections[j].Section.SectionNumber {
			return sortedSections[i].Section.SectionNumber < sortedSections[j].Section.SectionNumber
		}
		// Then compare by State ("good" comes before "duplicate", then "unique")
		stateOrder := map[string]int{
			"good":      1,
			"duplicate": 2,
			"unique":    3,
		}
		return stateOrder[sortedSections[i].State] < stateOrder[sortedSections[j].State]
	})

	// Convert the sorted slice back to a map for return
	sortedSectionMap := make(map[string]models.SectionState)
	for _, sectionState := range sortedSections {
		sortedSectionMap[sectionState.Section.SectionNumber] = sectionState
	}

	return sortedSectionMap
}
