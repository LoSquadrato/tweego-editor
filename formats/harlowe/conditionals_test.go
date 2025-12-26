package harlowe

import (
	"testing"
)

// ============================================
// Test 4.1: IF Standalone
// ============================================

func TestIfStandaloneTrue(t *testing.T) {
	state := map[string]interface{}{
		"vita": 100.0,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test: (if: $vita > 50)[Salute OK]
	result, err := handler.ProcessConditionalChain("(if: $vita > 50)[Salute OK]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if !result.ConditionMet {
		t.Error("Expected condition to be true")
	}
	
	if result.ActiveHook != "Salute OK" {
		t.Errorf("Expected 'Salute OK', got '%s'", result.ActiveHook)
	}
	
	if result.HookType != "if" {
		t.Errorf("Expected type 'if', got '%s'", result.HookType)
	}
	
	t.Logf("✅ IF standalone (true): hook = '%s'", result.ActiveHook)
}

func TestIfStandaloneFalse(t *testing.T) {
	state := map[string]interface{}{
		"vita": 30.0,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test: (if: $vita > 50)[Salute OK]
	result, err := handler.ProcessConditionalChain("(if: $vita > 50)[Salute OK]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if result.ConditionMet {
		t.Error("Expected condition to be false")
	}
	
	if result.ActiveHook != "" {
		t.Errorf("Expected empty hook, got '%s'", result.ActiveHook)
	}
	
	t.Log("✅ IF standalone (false): hook nascosto")
}

// ============================================
// Test 4.2: IF + ELSE
// ============================================

func TestIfElseTrue(t *testing.T) {
	state := map[string]interface{}{
		"vita": 100.0,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test: (if: $vita > 50)[Ottimo](else:)[Male]
	result, err := handler.ProcessConditionalChain("(if: $vita > 50)[Ottimo](else:)[Male]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if !result.ConditionMet {
		t.Error("Expected condition to be true")
	}
	
	if result.ActiveHook != "Ottimo" {
		t.Errorf("Expected 'Ottimo', got '%s'", result.ActiveHook)
	}
	
	t.Logf("✅ IF+ELSE (if true): hook = '%s'", result.ActiveHook)
}

func TestIfElseFalse(t *testing.T) {
	state := map[string]interface{}{
		"vita": 30.0,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test: (if: $vita > 50)[Ottimo](else:)[Male]
	result, err := handler.ProcessConditionalChain("(if: $vita > 50)[Ottimo](else:)[Male]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if !result.ConditionMet {
		t.Error("Expected else to be shown (conditionMet = true)")
	}
	
	if result.ActiveHook != "Male" {
		t.Errorf("Expected 'Male', got '%s'", result.ActiveHook)
	}
	
	if result.HookType != "else" {
		t.Errorf("Expected type 'else', got '%s'", result.HookType)
	}
	
	t.Logf("✅ IF+ELSE (if false): hook = '%s'", result.ActiveHook)
}

// ============================================
// Test 4.3: IF + ELSE-IF + ELSE (Catena completa)
// ============================================

func TestIfElseIfElseFirstTrue(t *testing.T) {
	state := map[string]interface{}{
		"vita": 90.0,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test: (if: $vita > 80)[Ottimo](else-if: $vita > 50)[Discreto](else:)[Male]
	result, err := handler.ProcessConditionalChain("(if: $vita > 80)[Ottimo](else-if: $vita > 50)[Discreto](else:)[Male]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if result.ActiveHook != "Ottimo" {
		t.Errorf("Expected 'Ottimo', got '%s'", result.ActiveHook)
	}
	
	t.Logf("✅ IF+ELSE-IF+ELSE (if true): hook = '%s'", result.ActiveHook)
}

func TestIfElseIfElseSecondTrue(t *testing.T) {
	state := map[string]interface{}{
		"vita": 60.0,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test: (if: $vita > 80)[Ottimo](else-if: $vita > 50)[Discreto](else:)[Male]
	result, err := handler.ProcessConditionalChain("(if: $vita > 80)[Ottimo](else-if: $vita > 50)[Discreto](else:)[Male]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if result.ActiveHook != "Discreto" {
		t.Errorf("Expected 'Discreto', got '%s'", result.ActiveHook)
	}
	
	if result.HookType != "else-if" {
		t.Errorf("Expected type 'else-if', got '%s'", result.HookType)
	}
	
	t.Logf("✅ IF+ELSE-IF+ELSE (else-if true): hook = '%s'", result.ActiveHook)
}

func TestIfElseIfElseAllFalse(t *testing.T) {
	state := map[string]interface{}{
		"vita": 20.0,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test: (if: $vita > 80)[Ottimo](else-if: $vita > 50)[Discreto](else:)[Male]
	result, err := handler.ProcessConditionalChain("(if: $vita > 80)[Ottimo](else-if: $vita > 50)[Discreto](else:)[Male]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if result.ActiveHook != "Male" {
		t.Errorf("Expected 'Male', got '%s'", result.ActiveHook)
	}
	
	if result.HookType != "else" {
		t.Errorf("Expected type 'else', got '%s'", result.HookType)
	}
	
	t.Logf("✅ IF+ELSE-IF+ELSE (else): hook = '%s'", result.ActiveHook)
}

// ============================================
// Test 4.4: Multiple ELSE-IF
// ============================================

func TestMultipleElseIf(t *testing.T) {
	state := map[string]interface{}{
		"score": 75.0,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test catena lunga con multiple else-if
	expr := "(if: $score >= 90)[A](else-if: $score >= 80)[B](else-if: $score >= 70)[C](else-if: $score >= 60)[D](else:)[F]"
	result, err := handler.ProcessConditionalChain(expr)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if result.ActiveHook != "C" {
		t.Errorf("Expected 'C', got '%s'", result.ActiveHook)
	}
	
	t.Logf("✅ Multiple ELSE-IF: hook = '%s' (score=75)", result.ActiveHook)
}

// ============================================
// Test 4.5: UNLESS (inverso di IF)
// ============================================

func TestUnlessTrue(t *testing.T) {
	state := map[string]interface{}{
		"ha_chiave": false,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test: (unless: $ha_chiave)[Trova la chiave!]
	// UNLESS mostra hook se condizione è FALSE
	result, err := handler.ProcessConditionalChain("(unless: $ha_chiave)[Trova la chiave!]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if !result.ConditionMet {
		t.Error("Expected unless to show hook (conditionMet = true)")
	}
	
	if result.ActiveHook != "Trova la chiave!" {
		t.Errorf("Expected 'Trova la chiave!', got '%s'", result.ActiveHook)
	}
	
	if result.HookType != "unless" {
		t.Errorf("Expected type 'unless', got '%s'", result.HookType)
	}
	
	t.Logf("✅ UNLESS (condition false): hook = '%s'", result.ActiveHook)
}

func TestUnlessFalse(t *testing.T) {
	state := map[string]interface{}{
		"ha_chiave": true,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test: (unless: $ha_chiave)[Trova la chiave!]
	// UNLESS nasconde hook se condizione è TRUE
	result, err := handler.ProcessConditionalChain("(unless: $ha_chiave)[Trova la chiave!]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if result.ConditionMet {
		t.Error("Expected unless to hide hook (conditionMet = false)")
	}
	
	if result.ActiveHook != "" {
		t.Errorf("Expected empty hook, got '%s'", result.ActiveHook)
	}
	
	t.Log("✅ UNLESS (condition true): hook nascosto")
}

// ============================================
// Test 4.6: Operatori di Confronto
// ============================================

func TestComparisonOperators(t *testing.T) {
	state := map[string]interface{}{
		"vita": 75.0,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	tests := []struct {
		expr     string
		expected bool
		name     string
	}{
		{"(if: $vita > 50)[OK]", true, "Greater than (true)"},
		{"(if: $vita > 100)[OK]", false, "Greater than (false)"},
		{"(if: $vita < 100)[OK]", true, "Less than (true)"},
		{"(if: $vita < 50)[OK]", false, "Less than (false)"},
		{"(if: $vita >= 75)[OK]", true, "Greater or equal (true)"},
		{"(if: $vita >= 76)[OK]", false, "Greater or equal (false)"},
		{"(if: $vita <= 75)[OK]", true, "Less or equal (true)"},
		{"(if: $vita <= 74)[OK]", false, "Less or equal (false)"},
	}
	
	for _, test := range tests {
		result, err := handler.ProcessConditionalChain(test.expr)
		if err != nil {
			t.Errorf("[%s] Error: %v", test.name, err)
			continue
		}
		
		if result.ConditionMet != test.expected {
			t.Errorf("[%s] Expected %v, got %v", test.name, test.expected, result.ConditionMet)
		}
	}
	
	t.Log("✅ All comparison operators work correctly")
}

// ============================================
// Test 4.7: Operatori IS e IS NOT
// ============================================

func TestIsOperator(t *testing.T) {
	state := map[string]interface{}{
		"nome":  "Mario",
		"alive": true,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test IS (true)
	result, err := handler.ProcessConditionalChain(`(if: $nome is "Mario")[Ciao Mario!]`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if !result.ConditionMet {
		t.Error("Expected 'is' to be true")
	}
	
	// Test IS (false)
	result2, err := handler.ProcessConditionalChain(`(if: $nome is "Luigi")[Ciao Luigi!]`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if result2.ConditionMet {
		t.Error("Expected 'is' to be false")
	}
	
	t.Log("✅ IS operator works correctly")
}

func TestIsNotOperator(t *testing.T) {
	state := map[string]interface{}{
		"nome": "Mario",
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test IS NOT (true)
	result, err := handler.ProcessConditionalChain(`(if: $nome is not "Luigi")[Non sei Luigi]`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if !result.ConditionMet {
		t.Error("Expected 'is not' to be true")
	}
	
	// Test IS NOT (false)
	result2, err := handler.ProcessConditionalChain(`(if: $nome is not "Mario")[Non sei Mario]`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if result2.ConditionMet {
		t.Error("Expected 'is not' to be false")
	}
	
	t.Log("✅ IS NOT operator works correctly")
}

// ============================================
// Test 4.8: Operatore CONTAINS
// ============================================

func TestContainsInConditional(t *testing.T) {
	state := map[string]interface{}{
		"inv": []interface{}{"spada", "scudo", "pozione"},
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test CONTAINS (true)
	result, err := handler.ProcessConditionalChain(`(if: $inv contains "spada")[Hai la spada!]`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if !result.ConditionMet {
		t.Error("Expected contains to be true")
	}
	
	if result.ActiveHook != "Hai la spada!" {
		t.Errorf("Expected 'Hai la spada!', got '%s'", result.ActiveHook)
	}
	
	// Test CONTAINS (false)
	result2, err := handler.ProcessConditionalChain(`(if: $inv contains "arco")[Hai l'arco!]`)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	
	if result2.ConditionMet {
		t.Error("Expected contains to be false")
	}
	
	t.Log("✅ CONTAINS in conditionals works correctly")
}

// ============================================
// Test 4.9: Conversione a Boolean (toBool)
// ============================================

func TestToBoolSemantics(t *testing.T) {
	state := map[string]interface{}{
		"num_zero":     0.0,
		"num_positive": 10.0,
		"str_empty":    "",
		"str_full":     "hello",
		"arr_empty":    []interface{}{},
		"arr_full":     []interface{}{"item"},
		"bool_true":    true,
		"bool_false":   false,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	tests := []struct {
		expr     string
		expected bool
		name     string
	}{
		{"(if: $num_zero)[X]", false, "Number 0 = false"},
		{"(if: $num_positive)[X]", true, "Number > 0 = true"},
		{"(if: $str_empty)[X]", false, "Empty string = false"},
		{"(if: $str_full)[X]", true, "Non-empty string = true"},
		{"(if: $arr_empty)[X]", false, "Empty array = false"},
		{"(if: $arr_full)[X]", true, "Non-empty array = true"},
		{"(if: $bool_true)[X]", true, "Boolean true = true"},
		{"(if: $bool_false)[X]", false, "Boolean false = false"},
	}
	
	for _, test := range tests {
		result, err := handler.ProcessConditionalChain(test.expr)
		if err != nil {
			t.Errorf("[%s] Error: %v", test.name, err)
			continue
		}
		
		if result.ConditionMet != test.expected {
			t.Errorf("[%s] Expected %v, got %v", test.name, test.expected, result.ConditionMet)
		}
	}
	
	t.Log("✅ toBool semantics work correctly")
}

// ============================================
// Test 4.10: Nested Conditionals
// ============================================

func TestNestedConditionals(t *testing.T) {
	state := map[string]interface{}{
		"vita":      80.0,
		"ha_chiave": true,
	}
	
	eval := NewHarloweEvaluator(state)
	handler := NewConditionalHandler(eval)
	
	// Test conditional esterno
	outerExpr := "(if: $vita > 50)[Salute OK]"
	outerResult, err := handler.ProcessConditionalChain(outerExpr)
	if err != nil {
		t.Fatalf("Error outer: %v", err)
	}
	
	if outerResult.ActiveHook != "Salute OK" {
		t.Errorf("Expected 'Salute OK', got '%s'", outerResult.ActiveHook)
	}
	
	// Ora simula di processare un nested conditional DENTRO l'hook
	// (Nel vero parser, l'hook "Salute OK" potrebbe contenere altro codice)
	innerExpr := "(if: $ha_chiave)[Puoi aprire la porta]"
	innerResult, err := handler.ProcessConditionalChain(innerExpr)
	if err != nil {
		t.Fatalf("Error inner: %v", err)
	}
	
	if innerResult.ActiveHook != "Puoi aprire la porta" {
		t.Errorf("Expected 'Puoi aprire la porta', got '%s'", innerResult.ActiveHook)
	}
	
	t.Log("✅ Nested conditionals work (sequential processing)")
}