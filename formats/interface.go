package formats

import "tweego-editor/parser"

// StoryFormat definisce l'interface per i diversi formati
type StoryFormat interface {
	// ParseVariables estrae le variabili dal contenuto
	ParseVariables(content string) map[string]interface{}
	
	// ParseLinks estrae i collegamenti dal contenuto
	ParseLinks(content string) []string
	
	// StripCode rimuove il codice lasciando solo il testo
	StripCode(content string) string
	
	// GetFormatName restituisce il nome del formato
	GetFormatName() string
}