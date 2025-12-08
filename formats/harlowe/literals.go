package harlowe

// se si aggiungono i metodi vanno aggiornati i call dentro a parser

import (
	"regexp"
	"strconv"
	"strings"
)

// ParseArrayLiteral parsa (a: ...) e restituisce uno slice
// Input: `(a: "spada", "scudo", 10)`
// Output: []interface{}{"spada", "scudo", 10.0}
func ParseArrayLiteral(content string) []interface{} {
	// Regex per estrarre contenuto di (a: ...)
	arrayRegex := regexp.MustCompile(`\(a:\s*([^)]*)\)`)
	match := arrayRegex.FindStringSubmatch(content)
	
	if len(match) < 2 {
		return nil
	}
	
	// Parsa gli elementi
	elements := splitElements(match[1])
	result := make([]interface{}, 0, len(elements))
	
	for _, elem := range elements {
		elem = strings.TrimSpace(elem)
		if elem != "" {
			result = append(result, parseValue(elem))
		}
	}
	
	return result
}

// ParseDatamapLiteral parsa (dm: ...) e restituisce una map
// Input: `(dm: "nome", "Eroe", "vita", 100)`
// Output: map[string]interface{}{"nome": "Eroe", "vita": 100.0}
func ParseDatamapLiteral(content string) map[string]interface{} {
	// Regex per estrarre contenuto di (dm: ...)
	dmRegex := regexp.MustCompile(`\(dm:\s*([^)]*)\)`)
	match := dmRegex.FindStringSubmatch(content)
	
	if len(match) < 2 {
		return nil
	}
	
	// Parsa gli elementi (coppie chiave-valore)
	elements := splitElements(match[1])
	result := make(map[string]interface{})
	
	// Processa a coppie: chiave, valore, chiave, valore...
	for i := 0; i < len(elements)-1; i += 2 {
		key := strings.TrimSpace(elements[i])
		value := strings.TrimSpace(elements[i+1])
		
		// La chiave deve essere una stringa, rimuovi le virgolette
		key = stripQuotes(key)
		
		if key != "" {
			result[key] = parseValue(value)
		}
	}
	
	return result
}

// ParseDatasetLiteral parsa (ds: ...) e restituisce uno slice (valori unici)
// Input: `(ds: "forza", "agilità")`
// Output: []interface{}{"forza", "agilità"}
func ParseDatasetLiteral(content string) []interface{} {
	// Regex per estrarre contenuto di (ds: ...)
	dsRegex := regexp.MustCompile(`\(ds:\s*([^)]*)\)`)
	match := dsRegex.FindStringSubmatch(content)
	
	if len(match) < 2 {
		return nil
	}
	
	// Parsa gli elementi
	elements := splitElements(match[1])
	seen := make(map[interface{}]bool)
	result := make([]interface{}, 0, len(elements))
	
	for _, elem := range elements {
		elem = strings.TrimSpace(elem)
		if elem != "" {
			val := parseValue(elem)
			// Aggiungi solo se non già presente (set = valori unici)
			if !seen[val] {
				seen[val] = true
				result = append(result, val)
			}
		}
	}
	
	return result
}

// FindAllArrayLiterals trova tutti gli (a: ...) in un contenuto
func FindAllArrayLiterals(content string) [][]interface{} {
	arrayRegex := regexp.MustCompile(`\(a:\s*([^)]*)\)`)
	matches := arrayRegex.FindAllStringSubmatch(content, -1)
	
	results := make([][]interface{}, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 2 {
			elements := splitElements(match[1])
			arr := make([]interface{}, 0, len(elements))
			for _, elem := range elements {
				elem = strings.TrimSpace(elem)
				if elem != "" {
					arr = append(arr, parseValue(elem))
				}
			}
			results = append(results, arr)
		}
	}
	
	return results
}

// FindAllDatamapLiterals trova tutti i (dm: ...) in un contenuto
func FindAllDatamapLiterals(content string) []map[string]interface{} {
	dmRegex := regexp.MustCompile(`\(dm:\s*([^)]*)\)`)
	matches := dmRegex.FindAllStringSubmatch(content, -1)
	
	results := make([]map[string]interface{}, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 2 {
			elements := splitElements(match[1])
			dm := make(map[string]interface{})
			for i := 0; i < len(elements)-1; i += 2 {
				key := stripQuotes(strings.TrimSpace(elements[i]))
				value := strings.TrimSpace(elements[i+1])
				if key != "" {
					dm[key] = parseValue(value)
				}
			}
			results = append(results, dm)
		}
	}
	
	return results
}

// FindAllDatasetLiterals trova tutti i (ds: ...) in un contenuto
func FindAllDatasetLiterals(content string) [][]interface{} {
	dsRegex := regexp.MustCompile(`\(ds:\s*([^)]*)\)`)
	matches := dsRegex.FindAllStringSubmatch(content, -1)
	
	results := make([][]interface{}, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 2 {
			elements := splitElements(match[1])
			seen := make(map[interface{}]bool)
			ds := make([]interface{}, 0, len(elements))
			for _, elem := range elements {
				elem = strings.TrimSpace(elem)
				if elem != "" {
					val := parseValue(elem)
					if !seen[val] {
						seen[val] = true
						ds = append(ds, val)
					}
				}
			}
			results = append(results, ds)
		}
	}
	
	return results
}

// splitElements divide gli elementi rispettando le stringhe tra virgolette
// "spada", "scudo con, virgola", 10 -> ["spada", "scudo con, virgola", "10"]
func splitElements(content string) []string {
	var elements []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	
	for _, char := range content {
		switch {
		case (char == '"' || char == '\'') && !inQuotes:
			// Inizio stringa
			inQuotes = true
			quoteChar = char
			current.WriteRune(char)
		case char == quoteChar && inQuotes:
			// Fine stringa
			inQuotes = false
			quoteChar = 0
			current.WriteRune(char)
		case char == ',' && !inQuotes:
			// Separatore (solo fuori dalle stringhe)
			elem := strings.TrimSpace(current.String())
			if elem != "" {
				elements = append(elements, elem)
			}
			current.Reset()
		default:
			current.WriteRune(char)
		}
	}
	
	// Ultimo elemento
	elem := strings.TrimSpace(current.String())
	if elem != "" {
		elements = append(elements, elem)
	}
	
	return elements
}

// parseValue converte una stringa raw in un valore Go tipizzato
// "spada" -> "spada" (string)
// 10 -> 10.0 (float64)
// true -> true (bool)
// $variabile -> "$variabile" (string, per ora - gestito dopo)
func parseValue(raw string) interface{} {
	raw = strings.TrimSpace(raw)
	
	// Booleani
	if raw == "true" {
		return true
	}
	if raw == "false" {
		return false
	}
	
	// Stringa tra virgolette
	if (strings.HasPrefix(raw, `"`) && strings.HasSuffix(raw, `"`)) ||
		(strings.HasPrefix(raw, `'`) && strings.HasSuffix(raw, `'`)) {
		return raw[1 : len(raw)-1]
	}
	
	// Numero
	if num, err := strconv.ParseFloat(raw, 64); err == nil {
		return num
	}
	
	// Altrimenti restituisci come stringa (include $variabili)
	return raw
}

// stripQuotes rimuove le virgolette da una stringa
func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}