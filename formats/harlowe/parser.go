package harlowe

import (
	"regexp"
	"strings"
)

// HarloweFormat implementa StoryFormat per Harlowe
type HarloweFormat struct{}

// NewHarloweFormat crea un nuovo parser Harlowe
func NewHarloweFormat() *HarloweFormat {
	return &HarloweFormat{}
}

// GetFormatName restituisce "Harlowe"
func (h *HarloweFormat) GetFormatName() string {
	return "Harlowe"
}

// ParseLinks estrae i link [[...]] dal contenuto
func (h *HarloweFormat) ParseLinks(content string) []string {
	linkRegex := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)
	
	links := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			// Gestisce sia [[Link]] che [[Testo|Link]]
			linkParts := strings.Split(match[1], "|")
			linkTarget := strings.TrimSpace(linkParts[len(linkParts)-1])
			links = append(links, linkTarget)
		}
	}
	
	return links
}

// ParseVariables estrae variabili (set:, put:) dal contenuto
func (h *HarloweFormat) ParseVariables(content string) map[string]interface{} {
	variables := make(map[string]interface{})
	
	// Regex basilare per (set: $var to value)
	varRegex := regexp.MustCompile(`\(set:\s*\$(\w+)\s+to\s+([^)]+)\)`)
	matches := varRegex.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) > 2 {
			varName := match[1]
			varValue := strings.TrimSpace(match[2])
			variables[varName] = varValue
		}
	}
	
	return variables
}

// StripCode rimuove macro e codice Harlowe
func (h *HarloweFormat) StripCode(content string) string {
	// Rimuovi macro (...)
	macroRegex := regexp.MustCompile(`\([^)]+\)`)
	cleaned := macroRegex.ReplaceAllString(content, "")
	
	// Rimuovi markup HTML
	htmlRegex := regexp.MustCompile(`<[^>]+>`)
	cleaned = htmlRegex.ReplaceAllString(cleaned, "")
	
	// Pulisci spazi multipli
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	
	return strings.TrimSpace(cleaned)
}