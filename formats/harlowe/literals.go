package harlowe

import (
	"fmt"
	"strconv"
	"strings"
)

// ============================================
// LITERALS PARSER - Fase 1 + Fase 2
// ============================================

// ParseValue è il punto di ingresso principale per parsare qualsiasi valore Harlowe
// Gestisce ricorsivamente: literals, variabili, property access, operazioni
func ParseValue(expression string, eval *HarloweEvaluator) (interface{}, error) {
	expr := strings.TrimSpace(expression)
	
	// 1. Array Literal: (a: ...)
	if strings.HasPrefix(expr, "(a:") || strings.HasPrefix(expr, "(array:") {
		return ParseArrayLiteral(expr, eval)
	}
	
	// 2. Datamap Literal: (dm: ...)
	if strings.HasPrefix(expr, "(dm:") || strings.HasPrefix(expr, "(datamap:") {
		return ParseDatamapLiteral(expr, eval)
	}
	
	// 3. Dataset Literal: (ds: ...)
	if strings.HasPrefix(expr, "(ds:") || strings.HasPrefix(expr, "(dataset:") {
		return ParseDatasetLiteral(expr, eval)
	}
	
	// 4. String Literal con quotes
	if strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`) {
		return strings.Trim(expr, `"`), nil
	}
	
	// 5. Number Literal
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num, nil
	}
	
	// 6. Boolean Literal
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}
	
	// 7. Variable or Property Access: $var o $var's prop
	if strings.HasPrefix(expr, "$") {
		if strings.Contains(expr, "'s") {
			// Property access - usa evaluator
			return eval.EvaluatePropertyAccess(expr)
		}
		// Variabile semplice
		return eval.EvaluateExpression(expr)
	}
	
	// 8. Operazioni aritmetiche o espressioni complesse
	// Delega all'evaluator
	return eval.EvaluateExpression(expr)
}

// ============================================
// ARRAY LITERAL: (a: val1, val2, ...)
// ============================================

func ParseArrayLiteral(expression string, eval *HarloweEvaluator) ([]interface{}, error) {
	// Estrai contenuto tra parentesi
	content := extractMacroContent(expression)
	if content == "" {
		return []interface{}{}, nil // Array vuoto
	}
	
	// Split per virgole (gestendo virgole dentro stringhe/nested)
	elements := smartSplitComma(content)
	
	// Parse ricorsivo di ogni elemento
	result := []interface{}{}
	for _, elem := range elements {
		elem = strings.TrimSpace(elem)
		if elem == "" {
			continue
		}
		
		// RICORSIONE: ParseValue gestisce anche nested structures
		value, err := ParseValue(elem, eval)
		if err != nil {
			// Se il parsing fallisce, salva come stringa
			value = elem
		}
		
		result = append(result, value)
	}
	
	return result, nil
}

// ============================================
// DATAMAP LITERAL: (dm: key1, val1, key2, val2, ...)
// ============================================

func ParseDatamapLiteral(expression string, eval *HarloweEvaluator) (map[string]interface{}, error) {
	content := extractMacroContent(expression)
	if content == "" {
		return make(map[string]interface{}), nil // Datamap vuoto
	}
	
	elements := smartSplitComma(content)
	result := make(map[string]interface{})
	
	// Processa a coppie: key, value, key, value...
	for i := 0; i < len(elements); i += 2 {
		if i+1 >= len(elements) {
			// Numero dispari di elementi - errore
			return nil, fmt.Errorf("datamap has odd number of elements")
		}
		
		keyExpr := strings.TrimSpace(elements[i])
		valueExpr := strings.TrimSpace(elements[i+1])
		
		// Parse key (deve essere stringa)
		key := strings.Trim(keyExpr, `"`)
		
		// RICORSIONE: Parse value (può essere nested datamap!)
		value, err := ParseValue(valueExpr, eval)
		if err != nil {
			// Fallback a stringa
			value = valueExpr
		}
		
		result[key] = value
	}
	
	return result, nil
}

// ============================================
// DATASET LITERAL: (ds: val1, val2, ...)
// ============================================

func ParseDatasetLiteral(expression string, eval *HarloweEvaluator) (map[string]bool, error) {
	content := extractMacroContent(expression)
	if content == "" {
		return make(map[string]bool), nil // Dataset vuoto
	}
	
	elements := smartSplitComma(content)
	result := make(map[string]bool)
	
	// Dataset = set di valori unici
	for _, elem := range elements {
		elem = strings.TrimSpace(elem)
		if elem == "" {
			continue
		}
		
		// Parse value
		value, err := ParseValue(elem, eval)
		if err != nil {
			value = elem
		}
		
		// Converti in stringa per usare come chiave
		key := fmt.Sprintf("%v", value)
		result[key] = true
	}
	
	return result, nil
}

// ============================================
// UTILITY FUNCTIONS
// ============================================

// extractMacroContent estrae il contenuto tra parentesi di una macro
// Input:  "(a: 1, 2, 3)"
// Output: "1, 2, 3"
func extractMacroContent(macro string) string {
	// Trova la prima parentesi aperta
	start := strings.Index(macro, ":")
	if start == -1 {
		return ""
	}
	start++ // Salta i due punti
	
	// Trova l'ultima parentesi chiusa
	end := strings.LastIndex(macro, ")")
	if end == -1 || end <= start {
		return ""
	}
	
	content := macro[start:end]
	return strings.TrimSpace(content)
}

// smartSplitComma split per virgole ma rispetta stringhe e nested structures
func smartSplitComma(content string) []string {
	result := []string{}
	current := ""
	depth := 0
	inString := false
	escapeNext := false
	
	for i := 0; i < len(content); i++ {
		char := content[i]
		
		// Gestione escape
		if escapeNext {
			current += string(char)
			escapeNext = false
			continue
		}
		
		if char == '\\' {
			escapeNext = true
			current += string(char)
			continue
		}
		
		// Gestione stringhe
		if char == '"' {
			inString = !inString
			current += string(char)
			continue
		}
		
		if inString {
			current += string(char)
			continue
		}
		
		// Gestione parentesi (per nested structures)
		if char == '(' {
			depth++
			current += string(char)
			continue
		}
		
		if char == ')' {
			depth--
			current += string(char)
			continue
		}
		
		// Split per virgola solo a depth 0
		if char == ',' && depth == 0 {
			trimmed := strings.TrimSpace(current)
			if trimmed != "" {
				result = append(result, trimmed)
			}
			current = ""
			continue
		}
		
		current += string(char)
	}
	
	// Aggiungi ultimo elemento se non vuoto
	trimmed := strings.TrimSpace(current)
	if trimmed != "" {
		result = append(result, trimmed)
	}
	
	return result
}

// ============================================
// HELPER: Parse Assignment
// ============================================

// ParseAssignment parsa un'intera assegnazione: "$var to value" o "$var's prop to value"
func ParseAssignment(assignment string, eval *HarloweEvaluator) error {
	// Split per " to "
	parts := strings.Split(assignment, " to ")
	if len(parts) != 2 {
		return fmt.Errorf("invalid assignment syntax: %s", assignment)
	}
	
	varPath := strings.TrimSpace(parts[0])
	valueExpr := strings.TrimSpace(parts[1])
	
	// Verifica che sia una variabile
	if !strings.HasPrefix(varPath, "$") {
		return fmt.Errorf("assignment target must be a variable: %s", varPath)
	}
	
	// Gestisci operatore "it"
	if strings.Contains(valueExpr, "it") {
		varName := extractVarName(varPath)
		if replaced, err := eval.ReplaceItKeyword(valueExpr, varName); err == nil {
			valueExpr = replaced
		}
	}
	
	// Parse value (RICORSIVO!)
	value, err := ParseValue(valueExpr, eval)
	if err != nil {
		return fmt.Errorf("error parsing value: %w", err)
	}
	
	// Set value
	if strings.Contains(varPath, "'s") {
		// Property assignment
		return eval.SetProperty(varPath, value)
	} else {
		// Simple variable
		varName := strings.TrimPrefix(varPath, "$")
		eval.state[varName] = value
		return nil
	}
}

// extractVarName estrae il nome base della variabile
// "$Mago's vita" -> "Mago"
// "$hero" -> "hero"
func extractVarName(varPath string) string {
	varPath = strings.TrimPrefix(varPath, "$")
	if idx := strings.Index(varPath, "'s"); idx != -1 {
		return varPath[:idx]
	}
	return varPath
}

// ============================================
// TYPE CONVERSION UTILITIES
// ============================================

// ConvertToFloat converte un valore in float64 se possibile
func ConvertToFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

// ConvertToString converte un valore in stringa
func ConvertToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		// Rimuovi ".0" se è un intero
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%v", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}