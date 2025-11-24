package parser

import "time"

// Passage rappresenta un singolo passaggio Twine
type Passage struct {
	Title      string            `json:"title"`
	Tags       []string          `json:"tags"`
	Content    string            `json:"content"`
	Position   Position          `json:"position"`
	Metadata   map[string]string `json:"metadata"`
	ParsedAt   time.Time         `json:"parsed_at"`
}

// Position rappresenta la posizione del passaggio nell'editor
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Story rappresenta l'intera storia
type Story struct {
	Title     string              `json:"title"`
	Passages  map[string]*Passage `json:"passages"`
	IFID      string              `json:"ifid"`
	Format    string              `json:"format"`
	FormatVersion string           `json:"format_version"`
}
