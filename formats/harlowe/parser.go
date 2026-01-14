package harlowe

import (
	"fmt"
	"regexp"
	"strings"
	"tweego-editor/formats"
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

// CreateEvaluator crea un nuovo evaluator per Harlowe
// Implementa formats.StoryFormat interface
func (h *HarloweFormat) CreateEvaluator(initialState map[string]interface{}) formats.Evaluator {
	return NewHarloweEvaluator(initialState)
}

// ParseLinks estrae i link [[...]] dal contenuto
func (h *HarloweFormat) ParseLinks(content string) []string {
	linkRegex := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)
	
	links := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			// Gestisce sia [[Link]] che [[Testo->Link]]
			linkParts := strings.Split(match[1], "->")
			if len(linkParts) == 1 {
				// Prova anche con <-
				linkParts = strings.Split(match[1], "<-")
			}
			linkTarget := strings.TrimSpace(linkParts[len(linkParts)-1])
			links = append(links, linkTarget)
		}
	}
	
	return links
}

// ParseVariables estrae variabili (set:, put:, move:) dal contenuto
// USA ARCHITETTURA MODULARE: Parser → Literals → Evaluator
func (h *HarloweFormat) ParseVariables(content string) map[string]interface{} {
	// Crea evaluator con stato vuoto
	eval := NewHarloweEvaluator(nil)
	
	// 1. Parse (set:) macro
	h.parseSetMacro(content, eval)
	
	// 2. Parse (put:) macro
	h.parsePutMacro(content, eval)
	
	// 3. Parse (move:) macro
	h.parseMoveMacro(content, eval)
	
	// Restituisci stato finale
	return eval.GetState()
}

// parseSetMacro gestisce (set: $var to value, $var2 to value2, ...)
func (h *HarloweFormat) parseSetMacro(content string, eval *HarloweEvaluator) {
	setRegex := regexp.MustCompile(`\(set:\s*`)
	indices := setRegex.FindAllStringIndex(content, -1)
	
	for _, idx := range indices {
		start := idx[1]
		end := h.findMatchingParen(content, idx[0])
		if end == -1 {
			continue
		}
		
		assignmentContent := content[start:end]
		assignments := smartSplitComma(assignmentContent)
		
		for _, assignment := range assignments {
			if err := ParseAssignment(assignment, eval); err != nil {
				_ = err
			}
		}
	}
}

// findMatchingParen trova la parentesi chiusa corrispondente
func (h *HarloweFormat) findMatchingParen(content string, openPos int) int {
	if openPos >= len(content) || content[openPos] != '(' {
		return -1
	}
	
	depth := 1
	inString := false
	
	for i := openPos + 1; i < len(content); i++ {
		char := content[i]
		
		if char == '"' && (i == 0 || content[i-1] != '\\') {
			inString = !inString
			continue
		}
		
		if inString {
			continue
		}
		
		if char == '(' {
			depth++
		} else if char == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	
	return -1
}

// parsePutMacro gestisce (put: value into $var)
func (h *HarloweFormat) parsePutMacro(content string, eval *HarloweEvaluator) {
	putRegex := regexp.MustCompile(`\(put:\s*`)
	indices := putRegex.FindAllStringIndex(content, -1)
	
	for _, idx := range indices {
		start := idx[1]
		end := h.findMatchingParen(content, start-1)
		if end == -1 {
			continue
		}
		
		putContent := content[start:end]
		parts := strings.Split(putContent, " into ")
		if len(parts) != 2 {
			continue
		}
		
		valueExpr := strings.TrimSpace(parts[0])
		target := strings.TrimSpace(parts[1])
		
		value, err := ParseValue(valueExpr, eval)
		if err != nil {
			value = valueExpr
		}
		
		eval.Put(value, target)
	}
}

// parseMoveMacro gestisce (move: $source into $dest)
func (h *HarloweFormat) parseMoveMacro(content string, eval *HarloweEvaluator) {
	moveRegex := regexp.MustCompile(`\(move:\s*`)
	indices := moveRegex.FindAllStringIndex(content, -1)
	
	for _, idx := range indices {
		start := idx[1]
		end := h.findMatchingParen(content, start-1)
		if end == -1 {
			continue
		}
		
		moveContent := content[start:end]
		parts := strings.Split(moveContent, " into ")
		if len(parts) != 2 {
			continue
		}
		
		source := strings.TrimSpace(parts[0])
		dest := strings.TrimSpace(parts[1])
		
		eval.Move(source, dest)
	}
}

// StripCode rimuove macro e codice Harlowe
func (h *HarloweFormat) StripCode(content string) string {
	macroRegex := regexp.MustCompile(`\([^)]+\)`)
	cleaned := macroRegex.ReplaceAllString(content, "")
	
	htmlRegex := regexp.MustCompile(`<[^>]+>`)
	cleaned = htmlRegex.ReplaceAllString(cleaned, "")
	
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	
	return strings.TrimSpace(cleaned)
}

// ============================================
// LITERALS METHODS (per interface StoryFormat)
// ============================================

// ParseArrayLiteral parsa un singolo array literal
func (h *HarloweFormat) ParseArrayLiteral(content string) []interface{} {
	eval := NewHarloweEvaluator(nil)
	result, err := ParseArrayLiteral(content, eval)
	if err != nil {
		return []interface{}{}
	}
	return result
}

// ParseDatamapLiteral parsa un singolo datamap literal
func (h *HarloweFormat) ParseDatamapLiteral(content string) map[string]interface{} {
	eval := NewHarloweEvaluator(nil)
	result, err := ParseDatamapLiteral(content, eval)
	if err != nil {
		return make(map[string]interface{})
	}
	return result
}

// ParseDatasetLiteral parsa un singolo dataset literal
func (h *HarloweFormat) ParseDatasetLiteral(content string) []interface{} {
	eval := NewHarloweEvaluator(nil)
	result, err := ParseDatasetLiteral(content, eval)
	if err != nil {
		return []interface{}{}
	}
	
	// Converti map[string]bool in []interface{}
	arr := make([]interface{}, 0, len(result))
	for key := range result {
		arr = append(arr, key)
	}
	return arr
}

// FindAllArrayLiterals trova tutti gli array literals nel contenuto
func (h *HarloweFormat) FindAllArrayLiterals(content string) [][]interface{} {
	regex := regexp.MustCompile(`\(a:|array:\s*[^\)]+\)`)
	matches := regex.FindAllString(content, -1)
	
	results := [][]interface{}{}
	for _, match := range matches {
		arr := h.ParseArrayLiteral(match)
		if len(arr) > 0 {
			results = append(results, arr)
		}
	}
	return results
}

// FindAllDatamapLiterals trova tutti i datamap literals nel contenuto
func (h *HarloweFormat) FindAllDatamapLiterals(content string) []map[string]interface{} {
	regex := regexp.MustCompile(`\(dm:|datamap:\s*[^\)]+\)`)
	matches := regex.FindAllString(content, -1)
	
	results := []map[string]interface{}{}
	for _, match := range matches {
		dm := h.ParseDatamapLiteral(match)
		if len(dm) > 0 {
			results = append(results, dm)
		}
	}
	return results
}

// FindAllDatasetLiterals trova tutti i dataset literals nel contenuto
func (h *HarloweFormat) FindAllDatasetLiterals(content string) [][]interface{} {
	regex := regexp.MustCompile(`\(ds:|dataset:\s*[^\)]+\)`)
	matches := regex.FindAllString(content, -1)
	
	results := [][]interface{}{}
	for _, match := range matches {
		ds := h.ParseDatasetLiteral(match)
		if len(ds) > 0 {
			results = append(results, ds)
		}
	}
	return results
}

// ExtractAllLiterals estrae tutti i literals con raw + parsed
func (h *HarloweFormat) ExtractAllLiterals(content string) *formats.LiteralsResult {
	result := &formats.LiteralsResult{
		Arrays:   []formats.LiteralInfo{},
		Datamaps: []formats.LiteralInfo{},
		Datasets: []formats.LiteralInfo{},
	}
	
	eval := NewHarloweEvaluator(nil)
	
	// Arrays
	arrayRegex := regexp.MustCompile(`\((a|array):[^\)]+\)`)
	arrayMatches := arrayRegex.FindAllString(content, -1)
	for _, raw := range arrayMatches {
		parsed, err := ParseArrayLiteral(raw, eval)
		if err == nil {
			result.Arrays = append(result.Arrays, formats.LiteralInfo{
				Raw:    raw,
				Parsed: parsed,
			})
		}
	}
	
	// Datamaps
	datamapRegex := regexp.MustCompile(`\((dm|datamap):[^\)]+\)`)
	datamapMatches := datamapRegex.FindAllString(content, -1)
	for _, raw := range datamapMatches {
		parsed, err := ParseDatamapLiteral(raw, eval)
		if err == nil {
			result.Datamaps = append(result.Datamaps, formats.LiteralInfo{
				Raw:    raw,
				Parsed: parsed,
			})
		}
	}
	
	// Datasets
	datasetRegex := regexp.MustCompile(`\((ds|dataset):[^\)]+\)`)
	datasetMatches := datasetRegex.FindAllString(content, -1)
	for _, raw := range datasetMatches {
		parsed, err := ParseDatasetLiteral(raw, eval)
		if err == nil {
			result.Datasets = append(result.Datasets, formats.LiteralInfo{
				Raw:    raw,
				Parsed: parsed,
			})
		}
	}
	
	return result
}

// ProcessPassageContent processa il contenuto di un passaggio
// modificando lo stato dell'evaluator passato
func (h *HarloweFormat) ProcessPassageContent(content string, eval formats.Evaluator) error {
	// Cast a HarloweEvaluator per accedere ai metodi specifici
	harloweEval, ok := eval.(*HarloweEvaluator)
	if !ok {
		return fmt.Errorf("evaluator non è di tipo HarloweEvaluator")
	}

	// Parse le varie macro che modificano lo stato
	h.parseSetMacro(content, harloweEval)
	h.parsePutMacro(content, harloweEval)
	h.parseMoveMacro(content, harloweEval)

	return nil
}