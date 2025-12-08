package harlowe

import (
	"regexp"
	"strings"

	"tweego-editor/formats"
)

// Regex compilate per performance (compilate una volta sola)
var (
	arrayRegex   = regexp.MustCompile(`\(a:\s*[^)]*\)`)
	datamapRegex = regexp.MustCompile(`\(dm:\s*[^)]*\)`)
	datasetRegex = regexp.MustCompile(`\(ds:\s*[^)]*\)`)
	linkRegex    = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	setRegex     = regexp.MustCompile(`\(set:\s*\$(\w+)\s+to\s+([^)]+)\)`)
	macroRegex   = regexp.MustCompile(`\([^)]+\)`)
	htmlRegex    = regexp.MustCompile(`<[^>]+>`)
)

// HarloweFormat implementa StoryFormat per Harlowe
type HarloweFormat struct{}

// init registra il formato Harlowe all'avvio
func init() {
	formats.RegisterFormat("harlowe", func() formats.StoryFormat {
		return NewHarloweFormat()
	})
}

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
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	links := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			// Gestisce sia [[Link]] che [[Testo->Link]] che [[Testo|Link]]
			linkText := match[1]

			// Prima controlla -> (formato Harlowe)
			if strings.Contains(linkText, "->") {
				parts := strings.Split(linkText, "->")
				linkTarget := strings.TrimSpace(parts[len(parts)-1])
				links = append(links, linkTarget)
			} else if strings.Contains(linkText, "|") {
				// Formato |
				parts := strings.Split(linkText, "|")
				linkTarget := strings.TrimSpace(parts[len(parts)-1])
				links = append(links, linkTarget)
			} else {
				// Link semplice [[Link]]
				links = append(links, strings.TrimSpace(linkText))
			}
		}
	}

	return links
}

// ParseVariables estrae variabili (set:, put:) dal contenuto
func (h *HarloweFormat) ParseVariables(content string) map[string]interface{} {
	variables := make(map[string]interface{})

	matches := setRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			varName := match[1]
			varValue := strings.TrimSpace(match[2])

			// Prova a parsare il valore
			variables[varName] = h.parseVariableValue(varValue)
		}
	}

	return variables
}

// parseVariableValue converte il valore stringa nel tipo Go appropriato
func (h *HarloweFormat) parseVariableValue(value string) interface{} {
	value = strings.TrimSpace(value)

	// È un array literal?
	if strings.HasPrefix(value, "(a:") {
		return ParseArrayLiteral(value)
	}

	// È un datamap literal?
	if strings.HasPrefix(value, "(dm:") {
		return ParseDatamapLiteral(value)
	}

	// È un dataset literal?
	if strings.HasPrefix(value, "(ds:") {
		return ParseDatasetLiteral(value)
	}

	// Altrimenti usa parseValue da literals.go
	return parseValue(value)
}

// StripCode rimuove macro e codice Harlowe
func (h *HarloweFormat) StripCode(content string) string {
	// Rimuovi macro (...)
	cleaned := macroRegex.ReplaceAllString(content, "")

	// Rimuovi markup HTML
	cleaned = htmlRegex.ReplaceAllString(cleaned, "")

	// Pulisci spazi multipli
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	return strings.TrimSpace(cleaned)
}

// === LITERALS - Wrapper per le funzioni in literals.go ===

// ParseArrayLiteral parsa un singolo array literal
func (h *HarloweFormat) ParseArrayLiteral(content string) []interface{} {
	return ParseArrayLiteral(content)
}

// ParseDatamapLiteral parsa un singolo datamap literal
func (h *HarloweFormat) ParseDatamapLiteral(content string) map[string]interface{} {
	return ParseDatamapLiteral(content)
}

// ParseDatasetLiteral parsa un singolo dataset literal
func (h *HarloweFormat) ParseDatasetLiteral(content string) []interface{} {
	return ParseDatasetLiteral(content)
}

// FindAllArrayLiterals trova tutti gli array literals nel contenuto
func (h *HarloweFormat) FindAllArrayLiterals(content string) [][]interface{} {
	return FindAllArrayLiterals(content)
}

// FindAllDatamapLiterals trova tutti i datamap literals nel contenuto
func (h *HarloweFormat) FindAllDatamapLiterals(content string) []map[string]interface{} {
	return FindAllDatamapLiterals(content)
}

// FindAllDatasetLiterals trova tutti i dataset literals nel contenuto
func (h *HarloweFormat) FindAllDatasetLiterals(content string) [][]interface{} {
	return FindAllDatasetLiterals(content)
}

// ExtractAllLiterals estrae tutti i literals con raw + parsed
// Metodo principale che il runner usa - tutta la logica Harlowe è qui
func (h *HarloweFormat) ExtractAllLiterals(content string) *formats.LiteralsResult {
	result := &formats.LiteralsResult{
		Arrays:   []formats.LiteralInfo{},
		Datamaps: []formats.LiteralInfo{},
		Datasets: []formats.LiteralInfo{},
	}

	// === ARRAYS ===
	rawArrays := arrayRegex.FindAllString(content, -1)
	for _, raw := range rawArrays {
		parsed := ParseArrayLiteral(raw)
		if parsed != nil {
			result.Arrays = append(result.Arrays, formats.LiteralInfo{
				Raw:    raw,
				Parsed: parsed,
			})
		}
	}

	// === DATAMAPS ===
	rawDatamaps := datamapRegex.FindAllString(content, -1)
	for _, raw := range rawDatamaps {
		parsed := ParseDatamapLiteral(raw)
		if parsed != nil {
			result.Datamaps = append(result.Datamaps, formats.LiteralInfo{
				Raw:    raw,
				Parsed: parsed,
			})
		}
	}

	// === DATASETS ===
	rawDatasets := datasetRegex.FindAllString(content, -1)
	for _, raw := range rawDatasets {
		parsed := ParseDatasetLiteral(raw)
		if parsed != nil {
			result.Datasets = append(result.Datasets, formats.LiteralInfo{
				Raw:    raw,
				Parsed: parsed,
			})
		}
	}

	return result
}