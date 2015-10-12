package giniapi

type Layout struct {
	Pages []PageLayout
}

type PageLayout struct {
	Number    int
	SizeX     float64
	SizeY     float64
	TextZones []TextZone
	Regions   []Region
}

type TextZone struct {
	Paragraphs []Paragraph
}

type PageCoordinates struct {
	W float64
	H float64
	T float64
	L float64
}

type Paragraph struct {
	PageCoordinates
	Lines []Line
}

type Line struct {
	PageCoordinates
	Words []Word
}

type Word struct {
	PageCoordinates
	Fontsize   float64
	FontFamily string
	Bold       bool
	Text       string
}

type Region struct {
	PageCoordinates
	Type string
}
