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
		// CASO SPECIALE: "length" su array
		if propertyName == "length" {
			return e.GetArrayLength(currentValue)
		}
		
		// CASO SPECIALE: ordinali su array (1st, 2nd, last, etc.)
		if e.isOrdinal(propertyName) {
			return e.GetArrayElement(currentValue, propertyName)
		}
		
		// CASO NORMALE: property access su datamap
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

// isOrdinal verifica se una stringa è un ordinale (1st, 2nd, last, etc.)
func (e *HarloweEvaluator) isOrdinal(s string) bool {
	ordinals := []string{"1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "10th", "last"}
	for _, ord := range ordinals {
		if s == ord {
			return true
		}
	}
	return false
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
// FASE 3: Array/Collection Operations
// ============================================

// ConcatenateArrays concatena due array o aggiunge elementi
func (e *HarloweEvaluator) ConcatenateArrays(left, right interface{}) ([]interface{}, error) {
	// Converti entrambi in array se necessario
	leftArray, err := e.toArray(left)
	if err != nil {
		return nil, fmt.Errorf("left operand is not an array: %w", err)
	}
	
	rightArray, err := e.toArray(right)
	if err != nil {
		return nil, fmt.Errorf("right operand is not an array: %w", err)
	}
	
	// Concatena
	result := make([]interface{}, len(leftArray)+len(rightArray))
	copy(result, leftArray)
	copy(result[len(leftArray):], rightArray)
	
	return result, nil
}

// Contains verifica se un array contiene un valore
func (e *HarloweEvaluator) Contains(array interface{}, value interface{}) (bool, error) {
	arr, err := e.toArray(array)
	if err != nil {
		return false, fmt.Errorf("operand is not an array: %w", err)
	}
	
	// Cerca il valore
	for _, item := range arr {
		if e.areEqual(item, value) {
			return true, nil
		}
	}
	
	return false, nil
}

// GetArrayLength restituisce la lunghezza di un array
func (e *HarloweEvaluator) GetArrayLength(array interface{}) (float64, error) {
	arr, err := e.toArray(array)
	if err != nil {
		return 0, fmt.Errorf("operand is not an array: %w", err)
	}
	
	return float64(len(arr)), nil
}

// GetArrayElement ottiene un elemento da un array per indice o posizione
func (e *HarloweEvaluator) GetArrayElement(array interface{}, position string) (interface{}, error) {
	arr, err := e.toArray(array)
	if err != nil {
		return nil, fmt.Errorf("operand is not an array: %w", err)
	}
	
	if len(arr) == 0 {
		return nil, fmt.Errorf("array is empty")
	}
	
	var index int
	
	// Gestisci posizioni speciali
	switch position {
	case "1st":
		index = 0
	case "2nd":
		index = 1
	case "3rd":
		index = 2
	case "4th":
		index = 3
	case "5th":
		index = 4
	case "6th":
		index = 5
	case "7th":
		index = 6
	case "8th":
		index = 7
	case "9th":
		index = 8
	case "10th":
		index = 9
	case "last":
		index = len(arr) - 1
	default:
		// Prova a parsare come numero
		num, err := strconv.Atoi(position)
		if err != nil {
			return nil, fmt.Errorf("invalid position: %s", position)
		}
		index = num - 1 // Harlowe usa indici 1-based
	}
	
	// Verifica bounds
	if index < 0 || index >= len(arr) {
		return nil, fmt.Errorf("index %d out of bounds (array length: %d)", index+1, len(arr))
	}
	
	return arr[index], nil
}

// ============================================
// UTILITY FUNCTIONS per Array Operations
// ============================================

// toArray converte un valore in array se possibile
func (e *HarloweEvaluator) toArray(value interface{}) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	default:
		return nil, fmt.Errorf("value is not an array: %T", value)
	}
}

// ============================================
// HELPER METHODS per Conditionals
// ============================================

// areEqual confronta due valori per uguaglianza
func (e *HarloweEvaluator) areEqual(left, right interface{}) bool {
	// Conversione numerica se necessario
	leftNum, leftIsNum := toNumber(left)
	rightNum, rightIsNum := toNumber(right)
	
	if leftIsNum && rightIsNum {
		return leftNum == rightNum
	}
	
	// Confronto booleani
	leftBool, leftIsBool := left.(bool)
	rightBool, rightIsBool := right.(bool)
	if leftIsBool && rightIsBool {
		return leftBool == rightBool
	}
	
	// Confronto stringhe
	leftStr, leftIsStr := left.(string)
	rightStr, rightIsStr := right.(string)
	if leftIsStr && rightIsStr {
		return leftStr == rightStr
	}
	
	// Confronto generico
	return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right)
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
	
	// Operatore "contains"
	if strings.Contains(expression, " contains ") {
		return e.evaluateContains(expression)
	}
	
	// Operazioni aritmetiche o concatenazione
	if strings.Contains(expression, "+") {
		return e.evaluatePlus(expression)
	}
	
	if strings.Contains(expression, "-") {
		return e.evaluateMinus(expression)
	}
	
	return nil, fmt.Errorf("cannot evaluate expression: %s", expression)
}

// evaluateContains valuta espressioni "X contains Y"
func (e *HarloweEvaluator) evaluateContains(expression string) (interface{}, error) {
	parts := strings.Split(expression, " contains ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid contains expression: %s", expression)
	}
	
	leftExpr := strings.TrimSpace(parts[0])
	rightExpr := strings.TrimSpace(parts[1])
	
	// Valuta entrambe le parti
	leftValue, err := e.EvaluateExpression(leftExpr)
	if err != nil {
		return nil, err
	}
	
	rightValue, err := e.EvaluateExpression(rightExpr)
	if err != nil {
		return nil, err
	}
	
	return e.Contains(leftValue, rightValue)
}

// evaluatePlus valuta operazioni con +
func (e *HarloweEvaluator) evaluatePlus(expression string) (interface{}, error) {
	parts := strings.Split(expression, "+")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid plus expression: %s", expression)
	}
	
	left, err1 := e.EvaluateExpression(strings.TrimSpace(parts[0]))
	right, err2 := e.EvaluateExpression(strings.TrimSpace(parts[1]))
	
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("error evaluating operands")
	}
	
	// Prova concatenazione array
	if _, ok := left.([]interface{}); ok {
		return e.ConcatenateArrays(left, right)
	}
	
	// Altrimenti somma numerica
	leftNum, lok := toNumber(left)
	rightNum, rok := toNumber(right)
	if lok && rok {
		return leftNum + rightNum, nil
	}
	
	// String concatenation
	leftStr, lok := left.(string)
	rightStr, rok := right.(string)
	if lok && rok {
		return leftStr + rightStr, nil
	}
	
	return nil, fmt.Errorf("cannot add %T and %T", left, right)
}

// evaluateMinus valuta operazioni con -
func (e *HarloweEvaluator) evaluateMinus(expression string) (interface{}, error) {
	parts := strings.Split(expression, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid minus expression: %s", expression)
	}
	
	left, err1 := e.EvaluateExpression(strings.TrimSpace(parts[0]))
	right, err2 := e.EvaluateExpression(strings.TrimSpace(parts[1]))
	
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("error evaluating operands")
	}
	
	// Sottrazione numerica
	leftNum, lok := toNumber(left)
	rightNum, rok := toNumber(right)
	if lok && rok {
		return leftNum - rightNum, nil
	}
	
	return nil, fmt.Errorf("cannot subtract %T and %T", left, right)
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

// GetTypeName ritorna il nome del tipo Harlowe di un valore
func (e *HarloweEvaluator) GetTypeName(value interface{}) string {
	switch value.(type) {
	case bool:
		return "boolean"
	case float64, int:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "datamap"
	case map[string]bool:
		return "dataset"
	default:
		return "unknown"
	}
}