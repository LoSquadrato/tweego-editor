package harlowe

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// HarloweEvaluator gestisce l'evaluation di espressioni Harlowe
// Implementa l'interfaccia formats.Evaluator
type HarloweEvaluator struct {
	state           map[string]interface{} // Stato delle variabili
	visitedPassages map[string]int         // Passato dal PathSimulator
	history         []string               // Passato dal PathSimulator
	currentPassage  string                 // Passato dal PathSimulator
}

// NewHarloweEvaluator crea un nuovo evaluator
func NewHarloweEvaluator(state map[string]interface{}) *HarloweEvaluator {
	if state == nil {
		state = make(map[string]interface{})
	}
	return &HarloweEvaluator{
		state:           state,
		visitedPassages: make(map[string]int),
		history:         []string{},
		currentPassage:  "",
	}
}

// ============================================
// INTERFACE IMPLEMENTATION: formats.Evaluator
// ============================================

// GetState restituisce lo stato corrente delle variabili
func (e *HarloweEvaluator) GetState() map[string]interface{} {
	return e.state
}

// SetState imposta lo stato delle variabili
func (e *HarloweEvaluator) SetState(state map[string]interface{}) {
	if state == nil {
		state = make(map[string]interface{})
	}
	e.state = state
}

// SetVisitedPassages imposta la mappa dei passaggi visitati (passata dal PathSimulator)
func (e *HarloweEvaluator) SetVisitedPassages(visited map[string]int) {
	if visited == nil {
		visited = make(map[string]int)
	}
	e.visitedPassages = visited
}

// SetHistory imposta la history dei passaggi (passata dal PathSimulator)
func (e *HarloweEvaluator) SetHistory(history []string) {
	if history == nil {
		history = []string{}
	}
	e.history = history
}

// SetCurrentPassage imposta il passaggio corrente (passato dal PathSimulator)
func (e *HarloweEvaluator) SetCurrentPassage(passageName string) {
	e.currentPassage = passageName
}

// EvaluateCondition valuta una condizione e ritorna true/false
func (e *HarloweEvaluator) EvaluateCondition(condition string) (bool, error) {
	result, err := e.EvaluateExpression(condition)
	if err != nil {
		return false, err
	}

	// Converti il risultato in booleano secondo le regole Harlowe
	switch v := result.(type) {
	case bool:
		return v, nil
	case float64:
		return v != 0, nil // Harlowe: numero != 0 è true
	case int:
		return v != 0, nil
	case string:
		return v != "", nil // Stringa non vuota è true
	case []interface{}:
		return len(v) > 0, nil // Array non vuoto è true
	case map[string]interface{}:
		return len(v) > 0, nil // Datamap non vuoto è true
	default:
		return result != nil, nil
	}
}

// ============================================
// VISITED & HISTORY HELPERS
// ============================================

// visited implementa la logica di (visited: "passaggio")
func (e *HarloweEvaluator) visited(passageName string) float64 {
	if count, exists := e.visitedPassages[passageName]; exists {
		return float64(count)
	}
	return 0
}

// visits ritorna il numero di visite al passaggio corrente
func (e *HarloweEvaluator) visits() float64 {
	return e.visited(e.currentPassage)
}

// ============================================
// 2.1 Property Access: $var's property
// ============================================

// EvaluatePropertyAccess valuta espressioni come $Mago's vita
func (e *HarloweEvaluator) EvaluatePropertyAccess(expression string) (interface{}, error) {
	parts := e.parsePropertyPath(expression)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid property access: %s", expression)
	}

	varName := parts[0]
	properties := parts[1:]

	currentValue, exists := e.state[varName]
	if !exists {
		return nil, fmt.Errorf("variable $%s does not exist", varName)
	}

	for _, propertyName := range properties {
		// CASO SPECIALE: "length" su array
		if propertyName == "length" {
			return e.getArrayLength(currentValue)
		}

		// CASO SPECIALE: ordinali su array (1st, 2nd, last, etc.)
		if e.isOrdinal(propertyName) {
			return e.getArrayElement(currentValue, propertyName)
		}

		// CASO NORMALE: property access su datamap
		datamap, ok := currentValue.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot access property '%s' on non-datamap value", propertyName)
		}

		value, exists := datamap[propertyName]
		if !exists {
			return nil, fmt.Errorf("property '%s' does not exist in path", propertyName)
		}

		currentValue = value
	}

	return currentValue, nil
}

// isOrdinal verifica se una stringa è un ordinale
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

// SetProperty imposta il valore di una property
func (e *HarloweEvaluator) SetProperty(varPath string, value interface{}) error {
	parts := e.parsePropertyPath(varPath)
	if len(parts) < 2 {
		return fmt.Errorf("invalid property path: %s", varPath)
	}

	varName := parts[0]
	properties := parts[1:]

	baseValue, exists := e.state[varName]
	if !exists {
		return fmt.Errorf("cannot set property on non-existent variable $%s. Create it first with (set: $%s to (dm:))", 
			varName, varName)
	}

	_, isDatamap := baseValue.(map[string]interface{})
	if !isDatamap {
		return fmt.Errorf("cannot set property '%s' on $%s: variable is %s, not a datamap. Use (set: $%s to (dm:)) first",
			properties[0], varName, e.GetTypeName(baseValue), varName)
	}

	// Naviga la struttura nested (ora sappiamo che esiste ed è un datamap)
	current := baseValue

	for i, prop := range properties {
		datamap, ok := current.(map[string]interface{})
		if !ok {
			return fmt.Errorf("cannot set property on non-datamap value at path element '%s'", prop)
		}

		if i == len(properties)-1 {
			datamap[prop] = value
			return nil
		}

		// Crea automaticamente nested datamap solo per path intermedi
		if _, exists := datamap[prop]; !exists {
			datamap[prop] = make(map[string]interface{})
		}

		current = datamap[prop]
		
		// Verifica che il nested value sia ancora un datamap
		if _, ok := current.(map[string]interface{}); !ok {
			return fmt.Errorf("cannot set nested property: '%s' is %s, not a datamap", 
				prop, e.GetTypeName(current))
		}
	}

	return nil
}

// parsePropertyPath converte "$Mago's vita's max" in ["Mago", "vita", "max"]
func (e *HarloweEvaluator) parsePropertyPath(path string) []string {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "$") {
		path = path[1:]
	}

	parts := strings.Split(path, "'s")

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

// ReplaceItKeyword sostituisce "it" con il valore corrente
func (e *HarloweEvaluator) ReplaceItKeyword(expression string, varName string) (string, error) {
	currentValue, exists := e.state[varName]
	if !exists {
		return "", fmt.Errorf("variable $%s does not exist for 'it' replacement", varName)
	}

	valueStr := e.valueToString(currentValue)
	itRegex := regexp.MustCompile(`\bit\b`)
	result := itRegex.ReplaceAllString(expression, valueStr)

	return result, nil
}

// valueToString converte un valore in stringa per uso in espressioni
func (e *HarloweEvaluator) valueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v)
	case float64:
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

// Put implementa (put: value into $var)
func (e *HarloweEvaluator) Put(value interface{}, target string) error {
	if strings.Contains(target, "'s") {
		return e.SetProperty(target, value)
	}

	varName := strings.TrimPrefix(target, "$")
	e.state[varName] = value
	return nil
}

// ============================================
// 2.5 (move:) Macro
// ============================================

// Move implementa (move: $source into $dest)
func (e *HarloweEvaluator) Move(source string, dest string) error {
	sourceName := strings.TrimPrefix(source, "$")
	value, exists := e.state[sourceName]
	if !exists {
		return fmt.Errorf("source variable $%s does not exist", sourceName)
	}

	if err := e.Put(value, dest); err != nil {
		return err
	}

	e.state[sourceName] = 0
	return nil
}

// ============================================
// FASE 3: Array/Collection Operations
// ============================================

// concatenateArrays concatena due array
func (e *HarloweEvaluator) concatenateArrays(left, right interface{}) ([]interface{}, error) {
	leftArray, err := e.toArray(left)
	if err != nil {
		return nil, fmt.Errorf("left operand is not an array: %w", err)
	}

	rightArray, err := e.toArray(right)
	if err != nil {
		return nil, fmt.Errorf("right operand is not an array: %w", err)
	}

	result := make([]interface{}, len(leftArray)+len(rightArray))
	copy(result, leftArray)
	copy(result[len(leftArray):], rightArray)

	return result, nil
}

// contains verifica se un array contiene un valore
func (e *HarloweEvaluator) contains(array interface{}, value interface{}) (bool, error) {
	arr, err := e.toArray(array)
	if err != nil {
		return false, fmt.Errorf("operand is not an array: %w", err)
	}

	for _, item := range arr {
		if e.areEqual(item, value) {
			return true, nil
		}
	}

	return false, nil
}

// getArrayLength restituisce la lunghezza di un array
func (e *HarloweEvaluator) getArrayLength(array interface{}) (float64, error) {
	arr, err := e.toArray(array)
	if err != nil {
		return 0, fmt.Errorf("operand is not an array: %w", err)
	}

	return float64(len(arr)), nil
}

// getArrayElement ottiene un elemento da un array
func (e *HarloweEvaluator) getArrayElement(array interface{}, position string) (interface{}, error) {
	arr, err := e.toArray(array)
	if err != nil {
		return nil, fmt.Errorf("operand is not an array: %w", err)
	}

	if len(arr) == 0 {
		return nil, fmt.Errorf("array is empty")
	}

	var index int

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
		num, err := strconv.Atoi(position)
		if err != nil {
			return nil, fmt.Errorf("invalid position: %s", position)
		}
		index = num - 1
	}

	if index < 0 || index >= len(arr) {
		return nil, fmt.Errorf("index %d out of bounds (array length: %d)", index+1, len(arr))
	}

	return arr[index], nil
}

// toArray converte un valore in array
func (e *HarloweEvaluator) toArray(value interface{}) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	default:
		return nil, fmt.Errorf("value is not an array: %T", value)
	}
}

// areEqual confronta due valori
func (e *HarloweEvaluator) areEqual(left, right interface{}) bool {
	leftNum, leftIsNum := toNumber(left)
	rightNum, rightIsNum := toNumber(right)

	if leftIsNum && rightIsNum {
		return leftNum == rightNum
	}

	leftBool, leftIsBool := left.(bool)
	rightBool, rightIsBool := right.(bool)
	if leftIsBool && rightIsBool {
		return leftBool == rightBool
	}

	leftStr, leftIsStr := left.(string)
	rightStr, rightIsStr := right.(string)
	if leftIsStr && rightIsStr {
		return leftStr == rightStr
	}

	return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right)
}

// ============================================
// FIX 3: REVERSE LOOKUP con operatore "of"
// ============================================

// evaluateOf gestisce il reverse lookup: $key of $datamap
// Es: "Slime" of (dm: "Slime", "Water") → "Water"
// Equivalente a: $datamap's $key
func (e *HarloweEvaluator) evaluateOf(expression string) (interface{}, error) {
	// Split per " of "
	parts := strings.Split(expression, " of ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid 'of' expression: %s", expression)
	}
	
	keyExpr := strings.TrimSpace(parts[0])
	datamapExpr := strings.TrimSpace(parts[1])
	
	// Valuta la chiave (può essere variabile, stringa, etc.)
	keyValue, err := e.EvaluateExpression(keyExpr)
	if err != nil {
		return nil, fmt.Errorf("error evaluating key in 'of' expression: %w", err)
	}
	
	// La chiave deve essere convertibile in stringa
	var keyString string
	switch v := keyValue.(type) {
	case string:
		keyString = v
	case float64:
		// Converti numero in stringa (es: 1 → "1")
		if v == float64(int(v)) {
			keyString = fmt.Sprintf("%d", int(v))
		} else {
			keyString = fmt.Sprintf("%v", v)
		}
	case int:
		keyString = fmt.Sprintf("%d", v)
	default:
		keyString = fmt.Sprintf("%v", v)
	}
	
	// Valuta il datamap
	datamapValue, err := e.EvaluateExpression(datamapExpr)
	if err != nil {
		return nil, fmt.Errorf("error evaluating datamap in 'of' expression: %w", err)
	}
	
	// Deve essere un datamap
	datamap, ok := datamapValue.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("right operand of 'of' must be a datamap, got %s", e.GetTypeName(datamapValue))
	}
	
	// Cerca la chiave nel datamap
	value, exists := datamap[keyString]
	if !exists {
		// 🔥 Come dice la documentazione: se la chiave non esiste → ERRORE
		availableKeys := make([]string, 0, len(datamap))
		for k := range datamap {
			availableKeys = append(availableKeys, k)
		}
		return nil, fmt.Errorf("key '%s' not found in datamap. Available keys: %v", keyString, availableKeys)
	}
	
	return value, nil
}


// EvaluateExpression valuta un'espressione
func (e *HarloweEvaluator) EvaluateExpression(expression string) (interface{}, error) {
	expression = strings.TrimSpace(expression)

	// Keyword "visits" - numero visite passaggio corrente
	if expression == "visits" {
		return e.visits(), nil
	}

	// Macro (visited: "passaggio")
	if strings.HasPrefix(expression, "(visited:") {
		return e.evaluateVisitedMacro(expression)
	}

	// Macro (history:)
	if expression == "(history:)" {
		return e.history, nil
	}

	// 🔥 NUOVO FIX 3: Literals inline - PRIMA di "of" per permettere parsing corretto
	// Array literal: (a: ...)
	if strings.HasPrefix(expression, "(a:") || strings.HasPrefix(expression, "(array:") {
		return ParseArrayLiteral(expression, e)
	}
	
	// Datamap literal: (dm: ...)
	if strings.HasPrefix(expression, "(dm:") || strings.HasPrefix(expression, "(datamap:") {
		return ParseDatamapLiteral(expression, e)
	}
	
	// Dataset literal: (ds: ...)
	if strings.HasPrefix(expression, "(ds:") || strings.HasPrefix(expression, "(dataset:") {
		return ParseDatasetLiteral(expression, e)
	}

	// Property access
	if strings.Contains(expression, "'s") {
		return e.EvaluatePropertyAccess(expression)
	}

	// Variabile semplice
	if strings.HasPrefix(expression, "$") {
		varName := strings.TrimPrefix(expression, "$")
		value, exists := e.state[varName]
		if !exists {
			return 0, nil
		}
		return value, nil
	}

	// Numero
	if num, err := strconv.ParseFloat(expression, 64); err == nil {
		return num, nil
	}

	// Stringa con quotes
	if strings.HasPrefix(expression, `"`) && strings.HasSuffix(expression, `"`) {
		return strings.Trim(expression, `"`), nil
	}

	// Booleani
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

	// 🔥 FIX 3: Operatore "of" (reverse lookup)
	if strings.Contains(expression, " of ") {
		return e.evaluateOf(expression)
	}

	// Operazioni con +
	if strings.Contains(expression, "+") {
		return e.evaluatePlus(expression)
	}

	// Operazioni con -
	if strings.Contains(expression, "-") {
		return e.evaluateMinus(expression)
	}

	// Comparazioni
	if strings.Contains(expression, " > ") || strings.Contains(expression, " < ") ||
		strings.Contains(expression, " >= ") || strings.Contains(expression, " <= ") ||
		strings.Contains(expression, " is ") || strings.Contains(expression, " == ") {
		return e.evaluateComparison(expression)
	}

	return nil, fmt.Errorf("cannot evaluate expression: %s", expression)
}

// evaluateVisitedMacro valuta (visited: "passaggio")
func (e *HarloweEvaluator) evaluateVisitedMacro(expression string) (interface{}, error) {
	visitedRegex := regexp.MustCompile(`\(visited:\s*"([^"]+)"\)`)
	matches := visitedRegex.FindStringSubmatch(expression)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid visited macro: %s", expression)
	}

	passageName := matches[1]
	count := e.visited(passageName)

	// Ritorna il conteggio come numero (convertito in bool nelle condizioni)
	return count, nil
}

// evaluateComparison valuta comparazioni
func (e *HarloweEvaluator) evaluateComparison(expression string) (interface{}, error) {
	patterns := []struct {
		op      string
		compare func(left, right float64) bool
	}{
		{" >= ", func(l, r float64) bool { return l >= r }},
		{" <= ", func(l, r float64) bool { return l <= r }},
		{" > ", func(l, r float64) bool { return l > r }},
		{" < ", func(l, r float64) bool { return l < r }},
		{" is ", func(l, r float64) bool { return l == r }},
		{" == ", func(l, r float64) bool { return l == r }},
	}

	for _, p := range patterns {
		if strings.Contains(expression, p.op) {
			parts := strings.SplitN(expression, p.op, 2)
			if len(parts) == 2 {
				left, err1 := e.EvaluateExpression(strings.TrimSpace(parts[0]))
				right, err2 := e.EvaluateExpression(strings.TrimSpace(parts[1]))

				if err1 != nil || err2 != nil {
					return nil, fmt.Errorf("error evaluating comparison operands")
				}

				leftNum, lok := toNumber(left)
				rightNum, rok := toNumber(right)

				if lok && rok {
					return p.compare(leftNum, rightNum), nil
				}

				if p.op == " is " || p.op == " == " {
					return e.areEqual(left, right), nil
				}
			}
		}
	}

	return nil, fmt.Errorf("cannot evaluate comparison: %s", expression)
}

// evaluateContains valuta "X contains Y"
func (e *HarloweEvaluator) evaluateContains(expression string) (interface{}, error) {
	parts := strings.Split(expression, " contains ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid contains expression: %s", expression)
	}

	leftExpr := strings.TrimSpace(parts[0])
	rightExpr := strings.TrimSpace(parts[1])

	leftValue, err := e.EvaluateExpression(leftExpr)
	if err != nil {
		return nil, err
	}

	rightValue, err := e.EvaluateExpression(rightExpr)
	if err != nil {
		return nil, err
	}

	return e.contains(leftValue, rightValue)
}

// ============================================
// FIX 2: DATAMAP MERGE OPERATION
// ============================================

// MergeDatamaps combina due datamap usando l'operatore +
// Le chiavi del datamap di destra sovrascrivono quelle di sinistra
// Es: (dm: "hp", 100) + (dm: "hp", 150, "mp", 50) → {hp: 150, mp: 50}
func (e *HarloweEvaluator) MergeDatamaps(left, right interface{}) (map[string]interface{}, error) {
	// Converti entrambi in datamap
	leftMap, lok := left.(map[string]interface{})
	if !lok {
		return nil, fmt.Errorf("left operand is not a datamap: %T", left)
	}
	
	rightMap, rok := right.(map[string]interface{})
	if !rok {
		return nil, fmt.Errorf("right operand is not a datamap: %T", right)
	}
	
	// Crea nuovo datamap con merge
	result := make(map[string]interface{})
	
	// Copia tutte le chiavi da left
	for key, value := range leftMap {
		result[key] = value
	}
	
	// Sovrascrivi/aggiungi chiavi da right
	for key, value := range rightMap {
		result[key] = value
	}
	
	return result, nil
}

// ============================================
// AGGIORNA evaluatePlus() per supportare datamap merge
// ============================================

// Sostituisci il metodo evaluatePlus() esistente con questo:

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

	// 🔥 FIX 2: Prova merge datamap PRIMA degli array
	if _, ok := left.(map[string]interface{}); ok {
		return e.MergeDatamaps(left, right)
	}

	// Concatenazione array
	if _, ok := left.([]interface{}); ok {
		return e.concatenateArrays(left, right)
	}

	// Somma numerica
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

// ============================================
// TYPE HELPERS
// ============================================

// GetTypeName ritorna il nome del tipo Harlowe di un valore
func (e *HarloweEvaluator) GetTypeName(value interface{}) string {
	if value == nil {
		return "empty"
	}

	switch value.(type) {
	case bool:
		return "boolean"
	case float64, int, int64:
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