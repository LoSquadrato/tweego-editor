package harlowe

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// HarloweEvaluator gestisce l'evaluation di espressioni Harlowe
type HarloweEvaluator struct {
	state map[string]interface{} // Stato delle variabili
}

// NewHarloweEvaluator crea un nuovo evaluator
func NewHarloweEvaluator(state map[string]interface{}) *HarloweEvaluator {
	if state == nil {
		state = make(map[string]interface{})
	}
	return &HarloweEvaluator{
		state: state,
	}
}

// ============================================
// 2.1 Property Access: $var's property
// ============================================

// EvaluatePropertyAccess valuta espressioni come $Mago's vita
func (e *HarloweEvaluator) EvaluatePropertyAccess(expression string) (interface{}, error) {
	// Parse del path completo: $Mago's stats's forza -> ["Mago", "stats", "forza"]
	parts := e.parsePropertyPath(expression)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid property access: %s", expression)
	}
	
	varName := parts[0]
	properties := parts[1:]
	
	// Inizia con la variabile base
	currentValue, exists := e.state[varName]
	if !exists {
		return nil, fmt.Errorf("variable $%s does not exist", varName)
	}
	
	// Attraversa tutte le properties
	for _, propertyName := range properties {
		// Il valore corrente deve essere un datamap
		datamap, ok := currentValue.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot access property '%s' on non-datamap value", propertyName)
		}
		
		// Accedi alla property
		value, exists := datamap[propertyName]
		if !exists {
			return nil, fmt.Errorf("property '%s' does not exist in path", propertyName)
		}
		
		currentValue = value
	}
	
	return currentValue, nil
}

// ============================================
// 2.2 Property Set: (set: $var's prop to value)
// ============================================

// SetProperty imposta il valore di una property, creando datamap se necessario
func (e *HarloweEvaluator) SetProperty(varPath string, value interface{}) error {
	// Parse del path: $Mago's vita -> ["Mago", "vita"]
	parts := e.parsePropertyPath(varPath)
	if len(parts) < 2 {
		return fmt.Errorf("invalid property path: %s", varPath)
	}
	
	varName := parts[0]
	properties := parts[1:]
	
	// Se la variabile non esiste, creala come datamap vuoto
	if _, exists := e.state[varName]; !exists {
		e.state[varName] = make(map[string]interface{})
	}
	
	// Naviga/crea la struttura nested
	current := e.state[varName]
	
	for i, prop := range properties {
		datamap, ok := current.(map[string]interface{})
		if !ok {
			return fmt.Errorf("cannot set property on non-datamap value")
		}
		
		// Se è l'ultima property, imposta il valore
		if i == len(properties)-1 {
			datamap[prop] = value
			return nil
		}
		
		// Altrimenti, naviga o crea il datamap nested
		if _, exists := datamap[prop]; !exists {
			datamap[prop] = make(map[string]interface{})
		}
		current = datamap[prop]
	}
	
	return nil
}

// parsePropertyPath converte "$Mago's vita's max" in ["Mago", "vita", "max"]
func (e *HarloweEvaluator) parsePropertyPath(path string) []string {
	// Rimuovi spazi e il dollaro iniziale
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "$") {
		path = path[1:]
	}
	
	// Split per 's
	parts := strings.Split(path, "'s")
	
	// Trim ogni parte
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}

// ============================================
// 2.3 Operatore "it"
// ============================================

// ReplaceItKeyword sostituisce "it" con il valore corrente della variabile
func (e *HarloweEvaluator) ReplaceItKeyword(expression string, varName string) (string, error) {
	// Ottieni il valore corrente della variabile
	currentValue, exists := e.state[varName]
	if !exists {
		return "", fmt.Errorf("variable $%s does not exist for 'it' replacement", varName)
	}
	
	// Converti il valore in stringa per la sostituzione
	valueStr := e.valueToString(currentValue)
	
	// Sostituisci "it" con il valore
	// NOTA: sostituiamo solo "it" come parola intera, non dentro altre parole
	itRegex := regexp.MustCompile(`\bit\b`)
	result := itRegex.ReplaceAllString(expression, valueStr)
	
	return result, nil
}

// valueToString converte un valore in una stringa per uso in espressioni
func (e *HarloweEvaluator) valueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v) // Stringhe vanno con quotes
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

// ============================================
// 2.4 (put:) Macro
// ============================================

// Put implementa la macro (put: value into $var)
func (e *HarloweEvaluator) Put(value interface{}, target string) error {
	// Il target può essere:
	// 1. Una variabile semplice: $var
	// 2. Una property: $var's prop
	// 3. Un elemento di array: $var's 1st
	
	if strings.Contains(target, "'s") {
		// È una property o array access
		return e.SetProperty(target, value)
	}
	
	// È una variabile semplice
	varName := strings.TrimPrefix(target, "$")
	e.state[varName] = value
	return nil
}

// ============================================
// 2.5 (move:) Macro
// ============================================

// Move implementa la macro (move: $source into $dest)
func (e *HarloweEvaluator) Move(source string, dest string) error {
	// Ottieni il valore dalla source
	sourceName := strings.TrimPrefix(source, "$")
	value, exists := e.state[sourceName]
	if !exists {
		return fmt.Errorf("source variable $%s does not exist", sourceName)
	}
	
	// Metti nella destinazione
	if err := e.Put(value, dest); err != nil {
		return err
	}
	
	// Rimuovi dalla source (impostala a 0 come default Harlowe)
	e.state[sourceName] = 0
	
	return nil
}

// ============================================
// Utility: Evaluate Simple Expression
// ============================================

// EvaluateExpression valuta un'espressione semplice (es: "90 + 10")
func (e *HarloweEvaluator) EvaluateExpression(expression string) (interface{}, error) {
	// Trim spaces
	expression = strings.TrimSpace(expression)
	
	// Se contiene property access, valutalo
	if strings.Contains(expression, "'s") {
		return e.EvaluatePropertyAccess(expression)
	}
	
	// Se è una variabile semplice
	if strings.HasPrefix(expression, "$") {
		varName := strings.TrimPrefix(expression, "$")
		value, exists := e.state[varName]
		if !exists {
			return 0, nil // Default Harlowe
		}
		return value, nil
	}
	
	// Se è un numero
	if num, err := strconv.ParseFloat(expression, 64); err == nil {
		return num, nil
	}
	
	// Se è una stringa con quotes
	if strings.HasPrefix(expression, `"`) && strings.HasSuffix(expression, `"`) {
		return strings.Trim(expression, `"`), nil
	}
	
	// Se è un booleano
	if expression == "true" {
		return true, nil
	}
	if expression == "false" {
		return false, nil
	}
	
	// Operazioni aritmetiche semplici (per supportare "it + 10")
	if strings.Contains(expression, "+") {
		parts := strings.Split(expression, "+")
		if len(parts) == 2 {
			left, err1 := e.EvaluateExpression(strings.TrimSpace(parts[0]))
			right, err2 := e.EvaluateExpression(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				// Somma numerica
				leftNum, lok := toNumber(left)
				rightNum, rok := toNumber(right)
				if lok && rok {
					return leftNum + rightNum, nil
				}
			}
		}
	}
	
	return nil, fmt.Errorf("cannot evaluate expression: %s", expression)
}

// toNumber converte un valore in float64
func toNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

// GetState restituisce lo stato corrente
func (e *HarloweEvaluator) GetState() map[string]interface{} {
	return e.state
}