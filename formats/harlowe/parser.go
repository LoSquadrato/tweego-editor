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
	// Regex per catturare tutto il contenuto di (set: ...)
	// Usa una regex più permissiva per catturare nested parentheses
	setRegex := regexp.MustCompile(`\(set:\s*`)
	indices := setRegex.FindAllStringIndex(content, -1)
	
	for _, idx := range indices {
		start := idx[1] // Dopo "(set: "
		
		// Trova la parentesi chiusa corrispondente
		end := h.findMatchingParen(content, idx[0]) // idx[0] è la posizione di '('
		if end == -1 {
			continue
		}
		
		assignmentContent := content[start:end]
		
		// Split per virgole (gestendo nested)
		assignments := smartSplitComma(assignmentContent)
		
		for _, assignment := range assignments {
			// Usa literals.go per parsare l'assignment
			if err := ParseAssignment(assignment, eval); err != nil {
				// Log error ma continua
				_ = err // Ignora errori per ora
			}
		}
	}
}

// findMatchingParen trova la parentesi chiusa che corrisponde alla parentesi aperta in position
func (h *HarloweFormat) findMatchingParen(content string, openPos int) int {
	if openPos >= len(content) || content[openPos] != '(' {
		return -1
	}
	
	depth := 1
	inString := false
	
	for i := openPos + 1; i < len(content); i++ {
		char := content[i]
		
		// Gestione stringhe
		if char == '"' && (i == 0 || content[i-1] != '\\') {
			inString = !inString
			continue
		}
		
		if inString {
			continue
		}
		
		// Conta parentesi
		if char == '(' {
			depth++
		} else if char == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	
	return -1 // Non trovata
}

// parsePutMacro gestisce (put: value into $var)
func (h *HarloweFormat) parsePutMacro(content string, eval *HarloweEvaluator) {
	// Regex per catturare (put: ... into ...)
	putRegex := regexp.MustCompile(`\(put:\s*`)
	indices := putRegex.FindAllStringIndex(content, -1)
	
	for _, idx := range indices {
		start := idx[1] // Dopo "(put: "
		
		// Trova la parentesi chiusa
		end := h.findMatchingParen(content, start-1)
		if end == -1 {
			continue
		}
		
		putContent := content[start:end]
		
		// Split per " into "
		parts := strings.Split(putContent, " into ")
		if len(parts) != 2 {
			continue
		}
		
		valueExpr := strings.TrimSpace(parts[0])
		target := strings.TrimSpace(parts[1])
		
		// Parse value usando literals.go
		value, err := ParseValue(valueExpr, eval)
		if err != nil {
			value = valueExpr // Fallback
		}
		
		// Esegui put
		eval.Put(value, target)
	}
}

// parseMoveMacro gestisce (move: $source into $dest)
func (h *HarloweFormat) parseMoveMacro(content string, eval *HarloweEvaluator) {
	// Regex per catturare (move: ... into ...)
	moveRegex := regexp.MustCompile(`\(move:\s*`)
	indices := moveRegex.FindAllStringIndex(content, -1)
	
	for _, idx := range indices {
		start := idx[1] // Dopo "(move: "
		
		// Trova la parentesi chiusa
		end := h.findMatchingParen(content, start-1)
		if end == -1 {
			continue
		}
		
		moveContent := content[start:end]
		
		// Split per " into "
		parts := strings.Split(moveContent, " into ")
		if len(parts) != 2 {
			continue
		}
		
		source := strings.TrimSpace(parts[0])
		dest := strings.TrimSpace(parts[1])
		
		// Esegui move
		eval.Move(source, dest)
	}
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