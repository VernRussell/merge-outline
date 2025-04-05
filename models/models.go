// models/models.go
package models

type Point struct {
	PointNumber string `json:"pointnumber"`
	PointText   string `json:"pointtext"`
}

type Description struct {
	DescriptionNumber string  `json:"descriptionnumber"`
	DescriptionHeader string  `json:"descriptionheader"`
	DescriptionText   string  `json:"descriptiontext"`
	Points            []Point `json:"points"`
}

type Section struct {
	SectionNumber string        `json:"sectionnumber"`
	SectionTitle  string        `json:"sectiontitle"`
	Descriptions  []Description `json:"descriptions"`
}

type Chapter struct {
	ChapterNumber  string        `json:"chapternumber"`
	OriginalNumber string        `json:"original_number"`
	Title          string        `json:"title"`
	Sections       []Section     `json:"sections"`
	Descriptions   []Description `json:"descriptions,omitempty"`
}

type Extra struct {
	Title        string        `json:"title"`
	Descriptions []Description `json:"descriptions"`
}

type Book struct {
	Title      string    `json:"title"`
	Intro      *Extra    `json:"introduction,omitempty"`
	Chapters   []Chapter `json:"chapters"`
	Conclusion *Extra    `json:"conclusion,omitempty"`
}
