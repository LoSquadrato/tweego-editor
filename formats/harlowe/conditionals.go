package harlowe

import (
	"fmt"
	"regexp"
	"strings"
)

// ============================================
// CONDITIONAL HANDLER - Fase 4
// ============================================

// ConditionalResult rappresenta il risultato di un conditional
type ConditionalResult struct {
	ConditionMet bool   // true se la condizione è soddisfatta
	ActiveHook   string // Il contenuto del hook attivo
	HookType     string // "if", "else-if", "else", "unless"
}

// ConditionalHandler gestisce la logica dei conditionals
type ConditionalHandler struct {
	eval *HarloweEvaluator
}

// NewConditionalHandler crea un nuovo handler
func NewConditionalHandler(eval *HarloweEvaluator) *ConditionalHandler {
	return &ConditionalHandler{
		eval: eval,
	}
}

// ============================================
// METODO PRINCIPALE: ProcessConditionalChain
// ============================================

// ProcessConditionalChain processa un'intera catena di conditionals
// Input: "(if: $vita > 80)[A](else-if: $vita > 50)[B](else:)[C]"
// Output: SOLO l'hook attivo (es: "B" se vita = 60)
func (ch *ConditionalHandler) ProcessConditionalChain(expression string) (*ConditionalResult, error) {
	// 1. Identifica il tipo di conditional iniziale
	if strings.HasPrefix(expression, "(if:") {
		return ch.processIfChain(expression)
	}
	
	if strings.HasPrefix(expression, "(unless:") {
		return ch.processUnlessChain(expression)
	}
	
	// (else-if:) e (else:) non possono essere standalone
	if strings.HasPrefix(expression, "(else-if:") || strings.HasPrefix(expression, "(else:") {
		return nil, fmt.Errorf("(else-if:) and (else:) require a preceding (if:) or (unless:)")
	}
	
	return nil, fmt.Errorf("not a conditional expression: %s", expression)
}

// ============================================
// CATENA IF
// ============================================

// processIfChain processa (if:)...(else-if:)*...(else:)?
func (ch *ConditionalHandler) processIfChain(expression string) (*ConditionalResult, error) {
	// Regex per catturare (if: condition)[hook]
	ifRegex := regexp.MustCompile(`^\(if:\s*([^)]+)\)\[([^\]]*)\]`)
	ifMatch := ifRegex.FindStringSubmatch(expression)
	
	if len(ifMatch) < 3 {
		return nil, fmt.Errorf("invalid if statement syntax")
	}
	
	ifCondition := strings.TrimSpace(ifMatch[1])
	ifHook := ifMatch[2]
	remainingExpr := expression[len(ifMatch[0]):] // Resto dopo (if:)[...]
	
	// Valuta la condizione IF
	conditionResult, err := ch.EvaluateCondition(ifCondition)
	if err != nil {
		return nil, fmt.Errorf("error evaluating if condition: %w", err)
	}
	
	// Se IF è TRUE → mostra hook, ignora resto
	if conditionResult {
		return &ConditionalResult{
			ConditionMet: true,
			ActiveHook:   ifHook,
			HookType:     "if",
		}, nil
	}
	
	// IF è FALSE → controlla se c'è (else-if:) o (else:)
	if strings.HasPrefix(remainingExpr, "(else-if:") {
		return ch.processElseIfChain(remainingExpr, false) // previousHookShown = false
	}
	
	if strings.HasPrefix(remainingExpr, "(else:)") {
		return ch.processElse(remainingExpr, false) // previousHookShown = false
	}
	
	// Nessun hook attivo
	return &ConditionalResult{
		ConditionMet: false,
		ActiveHook:   "",
		HookType:     "if",
	}, nil
}

// ============================================
// CATENA ELSE-IF
// ============================================

// processElseIfChain processa (else-if:)...(else-if:)*...(else:)?
func (ch *ConditionalHandler) processElseIfChain(expression string, previousHookShown bool) (*ConditionalResult, error) {
	// Regex per catturare (else-if: condition)[hook]
	elseIfRegex := regexp.MustCompile(`^\(else-if:\s*([^)]+)\)\[([^\]]*)\]`)
	match := elseIfRegex.FindStringSubmatch(expression)
	
	if len(match) < 3 {
		return nil, fmt.Errorf("invalid else-if statement syntax")
	}
	
	condition := strings.TrimSpace(match[1])
	hook := match[2]
	remainingExpr := expression[len(match[0]):] // Resto dopo (else-if:)[...]
	
	// LOGICA: Se hook precedente era MOSTRATO, salta questo
	if previousHookShown {
		return &ConditionalResult{
			ConditionMet: false,
			ActiveHook:   "",
			HookType:     "else-if",
		}, nil
	}
	
	// Hook precedente era nascosto → valuta condizione
	conditionResult, err := ch.EvaluateCondition(condition)
	if err != nil {
		return nil, fmt.Errorf("error evaluating else-if condition: %w", err)
	}
	
	// Se condizione TRUE → mostra hook, ignora resto
	if conditionResult {
		return &ConditionalResult{
			ConditionMet: true,
			ActiveHook:   hook,
			HookType:     "else-if",
		}, nil
	}
	
	// Condizione FALSE → controlla se c'è altro (else-if:) o (else:)
	if strings.HasPrefix(remainingExpr, "(else-if:") {
		return ch.processElseIfChain(remainingExpr, false) // Ricorsione!
	}
	
	if strings.HasPrefix(remainingExpr, "(else:)") {
		return ch.processElse(remainingExpr, false)
	}
	
	// Nessun hook attivo
	return &ConditionalResult{
		ConditionMet: false,
		ActiveHook:   "",
		HookType:     "else-if",
	}, nil
}

// ============================================
// ELSE
// ============================================

// processElse processa (else:)[hook]
func (ch *ConditionalHandler) processElse(expression string, previousHookShown bool) (*ConditionalResult, error) {
	// Regex per catturare (else:)[hook]
	elseRegex := regexp.MustCompile(`^\(else:\)\[([^\]]*)\]`)
	match := elseRegex.FindStringSubmatch(expression)
	
	if len(match) < 2 {
		return nil, fmt.Errorf("invalid else statement syntax")
	}
	
	hook := match[1]
	
	// LOGICA: Mostra SOLO se tutti i precedenti erano nascosti
	if !previousHookShown {
		return &ConditionalResult{
			ConditionMet: true,
			ActiveHook:   hook,
			HookType:     "else",
		}, nil
	}
	
	// Hook precedente era mostrato → nasconde
	return &ConditionalResult{
		ConditionMet: false,
		ActiveHook:   "",
		HookType:     "else",
	}, nil
}

// ============================================
// CATENA UNLESS
// ============================================

// processUnlessChain processa (unless:)...(else-if:)*...(else:)?
// UNLESS è l'inverso di IF: esegue il hook se la condizione è FALSE
func (ch *ConditionalHandler) processUnlessChain(expression string) (*ConditionalResult, error) {
	unlessRegex := regexp.MustCompile(`^\(unless:\s*([^)]+)\)\[([^\]]*)\]`)
	match := unlessRegex.FindStringSubmatch(expression)
	
	if len(match) < 3 {
		return nil, fmt.Errorf("invalid unless statement syntax")
	}
	
	condition := strings.TrimSpace(match[1])
	hook := match[2]
	remainingExpr := expression[len(match[0]):]
	
	// Valuta la condizione
	conditionResult, err := ch.EvaluateCondition(condition)
	if err != nil {
		return nil, fmt.Errorf("error evaluating unless condition: %w", err)
	}
	
	// UNLESS: mostra hook se condizione è FALSE (inverso di IF!)
	if !conditionResult {
		return &ConditionalResult{
			ConditionMet: true,
			ActiveHook:   hook,
			HookType:     "unless",
		}, nil
	}
	
	// Condizione era TRUE → nasconde hook, controlla else-if/else
	if strings.HasPrefix(remainingExpr, "(else-if:") {
		return ch.processElseIfChain(remainingExpr, false)
	}
	
	if strings.HasPrefix(remainingExpr, "(else:)") {
		return ch.processElse(remainingExpr, false)
	}
	
	// Nessun hook attivo
	return &ConditionalResult{
		ConditionMet: false,
		ActiveHook:   "",
		HookType:     "unless",
	}, nil
}

// ============================================
// EVALUATION LOGIC - UNICA VERSIONE
// ============================================

// EvaluateCondition valuta una condizione booleana
// Supporta TUTTI gli operatori Boolean di Harlowe
func (ch *ConditionalHandler) EvaluateCondition(condition string) (bool, error) {
	condition = strings.TrimSpace(condition)
	
	// Operatore "not" (unario, va valutato per primo)
	if strings.HasPrefix(condition, "not ") {
		innerCondition := strings.TrimSpace(condition[4:])
		result, err := ch.EvaluateCondition(innerCondition)
		if err != nil {
			return false, err
		}
		return !result, nil
	}
	
	// Operatore "and"
	if strings.Contains(condition, " and ") {
		return ch.evaluateAnd(condition)
	}
	
	// Operatore "or"
	if strings.Contains(condition, " or ") {
		return ch.evaluateOr(condition)
	}
	
	// Operatore "is not an"
	if strings.Contains(condition, " is not an ") {
		return ch.evaluateIsNotA(condition, " is not an ")
	}
	
	// Operatore "is not a"
	if strings.Contains(condition, " is not a ") {
		return ch.evaluateIsNotA(condition, " is not a ")
	}
	
	// Operatore "is an"
	if strings.Contains(condition, " is an ") {
		return ch.evaluateIsA(condition, " is an ")
	}
	
	// Operatore "is a"
	if strings.Contains(condition, " is a ") {
		return ch.evaluateIsA(condition, " is a ")
	}
	
	// Operatore "does not match"
	if strings.Contains(condition, " does not match ") {
		result, err := ch.evaluateMatches(condition, " does not match ")
		if err != nil {
			return false, err
		}
		return !result, nil
	}
	
	// Operatore "matches"
	if strings.Contains(condition, " matches ") {
		return ch.evaluateMatches(condition, " matches ")
	}
	
	// Operatore "does not contain"
	if strings.Contains(condition, " does not contain ") {
		parts := strings.Split(condition, " does not contain ")
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid 'does not contain' syntax")
		}
		
		// Riutilizza la logica di contains
		containsExpr := parts[0] + " contains " + parts[1]
		result, err := ch.eval.evaluateContains(containsExpr)
		if err != nil {
			return false, err
		}
		return !result.(bool), nil
	}
	
	// Operatore "is in"
	if strings.Contains(condition, " is in ") {
		return ch.evaluateIsIn(condition)
	}
	
	// Operatore "is not"
	if strings.Contains(condition, " is not ") {
		return ch.evaluateIsNot(condition)
	}
	
	// Operatore "is"
	if strings.Contains(condition, " is ") {
		return ch.evaluateIs(condition)
	}
	
	// Operatore "contains"
	if strings.Contains(condition, " contains ") {
		result, err := ch.eval.evaluateContains(condition)
		if err != nil {
			return false, err
		}
		return result.(bool), nil
	}
	
	// Operatori di confronto: >=, <=, >, <
	if strings.Contains(condition, ">=") {
		return ch.evaluateComparison(condition, ">=")
	}
	if strings.Contains(condition, "<=") {
		return ch.evaluateComparison(condition, "<=")
	}
	if strings.Contains(condition, ">") {
		return ch.evaluateComparison(condition, ">")
	}
	if strings.Contains(condition, "<") {
		return ch.evaluateComparison(condition, "<")
	}
	
	// Valuta come espressione booleana diretta
	result, err := ch.eval.EvaluateExpression(condition)
	if err != nil {
		return false, err
	}
	
	// Converti in bool
	return ch.toBool(result), nil
}

// ============================================
// OPERATORI BASE
// ============================================

// evaluateIs valuta "X is Y"
func (ch *ConditionalHandler) evaluateIs(condition string) (bool, error) {
	parts := strings.Split(condition, " is ")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid 'is' syntax: %s", condition)
	}
	
	left, err := ch.eval.EvaluateExpression(strings.TrimSpace(parts[0]))
	if err != nil {
		return false, err
	}
	
	right, err := ch.eval.EvaluateExpression(strings.TrimSpace(parts[1]))
	if err != nil {
		return false, err
	}
	
	return ch.eval.areEqual(left, right), nil
}

// evaluateIsNot valuta "X is not Y"
func (ch *ConditionalHandler) evaluateIsNot(condition string) (bool, error) {
	parts := strings.Split(condition, " is not ")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid 'is not' syntax: %s", condition)
	}
	
	left, err := ch.eval.EvaluateExpression(strings.TrimSpace(parts[0]))
	if err != nil {
		return false, err
	}
	
	right, err := ch.eval.EvaluateExpression(strings.TrimSpace(parts[1]))
	if err != nil {
		return false, err
	}
	
	return !ch.eval.areEqual(left, right), nil
}

// evaluateComparison valuta operatori di confronto (>, <, >=, <=)
func (ch *ConditionalHandler) evaluateComparison(condition string, operator string) (bool, error) {
	parts := strings.Split(condition, operator)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid comparison syntax: %s", condition)
	}
	
	leftExpr := strings.TrimSpace(parts[0])
	rightExpr := strings.TrimSpace(parts[1])
	
	// Valuta entrambe le parti
	leftVal, err := ch.eval.EvaluateExpression(leftExpr)
	if err != nil {
		return false, err
	}
	
	rightVal, err := ch.eval.EvaluateExpression(rightExpr)
	if err != nil {
		return false, err
	}
	
	// Converti in numeri
	leftNum, lok := toNumber(leftVal)
	rightNum, rok := toNumber(rightVal)
	
	if !lok || !rok {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}
	
	// Esegui confronto
	switch operator {
	case ">":
		return leftNum > rightNum, nil
	case "<":
		return leftNum < rightNum, nil
	case ">=":
		return leftNum >= rightNum, nil
	case "<=":
		return leftNum <= rightNum, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// ============================================
// OPERATORI LOGICI
// ============================================

// evaluateAnd valuta "X and Y"
func (ch *ConditionalHandler) evaluateAnd(condition string) (bool, error) {
	parts := strings.Split(condition, " and ")
	if len(parts) < 2 {
		return false, fmt.Errorf("invalid 'and' syntax: %s", condition)
	}
	
	// Valuta tutte le parti
	for _, part := range parts {
		result, err := ch.EvaluateCondition(strings.TrimSpace(part))
		if err != nil {
			return false, err
		}
		if !result {
			return false, nil // Short-circuit: se uno è false, tutto è false
		}
	}
	
	return true, nil
}

// evaluateOr valuta "X or Y"
func (ch *ConditionalHandler) evaluateOr(condition string) (bool, error) {
	parts := strings.Split(condition, " or ")
	if len(parts) < 2 {
		return false, fmt.Errorf("invalid 'or' syntax: %s", condition)
	}
	
	// Valuta tutte le parti
	for _, part := range parts {
		result, err := ch.EvaluateCondition(strings.TrimSpace(part))
		if err != nil {
			return false, err
		}
		if result {
			return true, nil // Short-circuit: se uno è true, tutto è true
		}
	}
	
	return false, nil
}

// ============================================
// OPERATORI AVANZATI
// ============================================

// evaluateIsIn valuta "X is in Y" (inverso di contains)
func (ch *ConditionalHandler) evaluateIsIn(condition string) (bool, error) {
	parts := strings.Split(condition, " is in ")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid 'is in' syntax: %s", condition)
	}
	
	leftExpr := strings.TrimSpace(parts[0])
	rightExpr := strings.TrimSpace(parts[1])
	
	// Invertiamo: right contains left
	containsExpr := rightExpr + " contains " + leftExpr
	result, err := ch.eval.evaluateContains(containsExpr)
	if err != nil {
		return false, err
	}
	
	return result.(bool), nil
}

// evaluateMatches valuta "X matches Y"
func (ch *ConditionalHandler) evaluateMatches(condition string, separator string) (bool, error) {
	parts := strings.Split(condition, separator)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid 'matches' syntax: %s", condition)
	}
	
	leftExpr := strings.TrimSpace(parts[0])
	rightExpr := strings.TrimSpace(parts[1])
	
	// Valuta entrambe le parti
	leftVal, err := ch.eval.EvaluateExpression(leftExpr)
	if err != nil {
		return false, err
	}
	
	rightVal, err := ch.eval.EvaluateExpression(rightExpr)
	if err != nil {
		return false, err
	}
	
	// Se rightVal è un tipo (string "boolean", "number", ecc.)
	rightStr, isStr := rightVal.(string)
	if isStr {
		leftType := ch.eval.GetTypeName(leftVal)
		return leftType == rightStr, nil
	}
	
	// Altrimenti confronta i valori direttamente
	return ch.eval.areEqual(leftVal, rightVal), nil
}

// evaluateIsA valuta "X is a Y" o "X is an Y"
func (ch *ConditionalHandler) evaluateIsA(condition string, separator string) (bool, error) {
	parts := strings.Split(condition, separator)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid 'is a/an' syntax: %s", condition)
	}
	
	leftExpr := strings.TrimSpace(parts[0])
	rightExpr := strings.TrimSpace(parts[1])
	
	// Valuta la parte sinistra
	leftVal, err := ch.eval.EvaluateExpression(leftExpr)
	if err != nil {
		return false, err
	}
	
	// La parte destra dovrebbe essere un tipo (o valutarsi come tale)
	rightVal, err := ch.eval.EvaluateExpression(rightExpr)
	if err != nil {
		return false, err
	}
	
	rightType, isStr := rightVal.(string)
	if !isStr {
		rightType = ch.eval.GetTypeName(rightVal)
	}
	
	leftType := ch.eval.GetTypeName(leftVal)
	return leftType == rightType, nil
}

// evaluateIsNotA valuta "X is not a Y" o "X is not an Y"
func (ch *ConditionalHandler) evaluateIsNotA(condition string, separator string) (bool, error) {
	result, err := ch.evaluateIsA(condition, separator)
	if err != nil {
		return false, err
	}
	return !result, nil
}

// ============================================
// UTILITY
// ============================================

// toBool converte un valore in booleano (semantica Harlowe)
func (ch *ConditionalHandler) toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0 // In Harlowe, numeri != 0 sono true
	case int:
		return v != 0
	case string:
		return v != "" // Stringhe non vuote sono true
	case []interface{}:
		return len(v) > 0 // Array non vuoti sono true
	case map[string]interface{}:
		return len(v) > 0 // Datamap non vuoti sono true
	default:
		return false
	}
}