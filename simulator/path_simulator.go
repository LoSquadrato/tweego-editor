package simulator

import (
	"fmt"
	"tweego-editor/formats/harlowe"
	"tweego-editor/parser"
)

// PathSimulator simula un percorso attraverso la storia
type PathSimulator struct {
	story  *parser.Story
	format harlowe.HarloweFormat
}

// VariableChange rappresenta il cambiamento di una variabile
type VariableChange struct {
	Name     string      `json:"name"`
	Previous interface{} `json:"previous"`
	Current  interface{} `json:"current"`
	Delta    interface{} `json:"delta,omitempty"` // Solo per numeri
}

// StepResult risultato di un singolo step
type StepResult struct {
	PassageTitle string                    `json:"passage_title"`
	PassageIndex int                       `json:"passage_index"`
	Changes      map[string]VariableChange `json:"changes"`
	Warnings     []string                  `json:"warnings,omitempty"`
	AvailableLinks []string                `json:"available_links"`
}

// SimulationResult risultato completo della simulazione
type SimulationResult struct {
	Success      bool                   `json:"success"`
	Path         []string               `json:"path"`
	Steps        []StepResult           `json:"steps"`
	FinalState   map[string]interface{} `json:"final_state"`
	Errors       []string               `json:"errors,omitempty"`
	TotalWarnings int                   `json:"total_warnings"`
}

// NewPathSimulator crea un nuovo simulatore
func NewPathSimulator(story *parser.Story) *PathSimulator {
	return &PathSimulator{
		story:  story,
		format: *harlowe.NewHarloweFormat(),
	}
}

// ValidatePath verifica che il path sia valido (passaggi collegati)
func (ps *PathSimulator) ValidatePath(path []string) []string {
	errors := []string{}
	
	// Verifica che tutti i passaggi esistano
	for i, passageTitle := range path {
		if _, exists := ps.story.Passages[passageTitle]; !exists {
			errors = append(errors, fmt.Sprintf("Step %d: passaggio '%s' non esiste", i+1, passageTitle))
		}
	}
	
	// Verifica che i passaggi siano collegati
	for i := 0; i < len(path)-1; i++ {
		currentTitle := path[i]
		nextTitle := path[i+1]
		
		passage, exists := ps.story.Passages[currentTitle]
		if !exists {
			continue // Già segnalato sopra
		}
		
		// Estrai i link dal passaggio corrente
		links := ps.format.ParseLinks(passage.Content)
		
		// Verifica che il prossimo passaggio sia tra i link
		linked := false
		for _, link := range links {
			if link == nextTitle {
				linked = true
				break
			}
		}
		
		if !linked {
			errors = append(errors, fmt.Sprintf(
				"Step %d→%d: '%s' non ha un link diretto a '%s'. Link disponibili: %v",
				i+1, i+2, currentTitle, nextTitle, links,
			))
		}
	}
	
	return errors
}

// SimulatePath simula l'esecuzione di un percorso
func (ps *PathSimulator) SimulatePath(path []string) *SimulationResult {
	result := &SimulationResult{
		Success: true,
		Path:    path,
		Steps:   []StepResult{},
		FinalState: make(map[string]interface{}),
		Errors: []string{},
	}
	
	// Valida il path prima di simulare
	validationErrors := ps.ValidatePath(path)
	if len(validationErrors) > 0 {
		result.Success = false
		result.Errors = validationErrors
		return result
	}
	
	// Stato corrente delle variabili
	currentState := make(map[string]interface{})
	
	// Simula ogni passaggio
	for i, passageTitle := range path {
		passage, exists := ps.story.Passages[passageTitle]
		if !exists {
			// Non dovrebbe mai succedere dopo la validazione
			continue
		}
		
		stepResult := StepResult{
			PassageTitle:   passageTitle,
			PassageIndex:   i + 1,
			Changes:        make(map[string]VariableChange),
			Warnings:       []string{},
			AvailableLinks: ps.format.ParseLinks(passage.Content),
		}
		
		// Estrai variabili definite in questo passaggio
		newVars := ps.format.ParseVariables(passage.Content)
		
		// Calcola i cambiamenti
		for varName, newValue := range newVars {
			previousValue, existed := currentState[varName]
			
			change := VariableChange{
				Name:     varName,
				Previous: previousValue,
				Current:  newValue,
			}
			
			// Calcola delta per valori numerici
			if existed {
				prevNum, prevIsNum := toNumber(previousValue)
				currNum, currIsNum := toNumber(newValue)
				
				if prevIsNum && currIsNum {
					change.Delta = currNum - prevNum
				}
			} else {
				// Variabile nuova
				change.Previous = nil
			}
			
			stepResult.Changes[varName] = change
			currentState[varName] = newValue
		}
		
		// Genera warnings
		stepResult.Warnings = ps.generateWarnings(passage, currentState, stepResult.Changes)
		result.TotalWarnings += len(stepResult.Warnings)
		
		result.Steps = append(result.Steps, stepResult)
	}
	
	// Stato finale
	result.FinalState = currentState
	
	return result
}

// generateWarnings genera warning per un passaggio
func (ps *PathSimulator) generateWarnings(passage *parser.Passage, currentState map[string]interface{}, changes map[string]VariableChange) []string {
	warnings := []string{}
	
	// Warning per valori critici
	for varName, change := range changes {
		// Esempio: vita sotto soglia critica
		if varName == "vita" || varName == "health" || varName == "hp" {
			if val, isNum := toNumber(change.Current); isNum {
				if val <= 0 {
					warnings = append(warnings, fmt.Sprintf("⚠️ %s è a 0 o negativo: %v", varName, val))
				} else if val <= 20 {
					warnings = append(warnings, fmt.Sprintf("⚠️ %s è sotto soglia critica: %v", varName, val))
				}
			}
		}
		
		// Warning per variabili sovrascritte senza essere usate
		if change.Previous != nil && change.Previous != change.Current {
			// TODO: Potremmo verificare se la variabile è stata usata tra i passaggi
		}
	}
	
	return warnings
}

// toNumber converte un valore in numero se possibile
func toNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case string:
		// Prova a parsare stringhe numeriche
		var num float64
		_, err := fmt.Sscanf(v, "%f", &num)
		return num, err == nil
	}
	return 0, false
}

// GetSuggestedPaths suggerisce percorsi validi dato un punto di partenza
func (ps *PathSimulator) GetSuggestedPaths(startPassage string, maxDepth int) [][]string {
	paths := [][]string{}
	
	// BFS per trovare tutti i percorsi possibili
	queue := [][]string{{startPassage}}
	
	for len(queue) > 0 && len(paths) < 10 { // Limit a 10 path per non esplodere
		currentPath := queue[0]
		queue = queue[1:]
		
		if len(currentPath) >= maxDepth {
			paths = append(paths, currentPath)
			continue
		}
		
		lastPassage := currentPath[len(currentPath)-1]
		passage, exists := ps.story.Passages[lastPassage]
		if !exists {
			continue
		}
		
		links := ps.format.ParseLinks(passage.Content)
		
		if len(links) == 0 {
			// Fine del percorso
			paths = append(paths, currentPath)
		} else {
			// Espandi i percorsi
			for _, link := range links {
				newPath := make([]string, len(currentPath))
				copy(newPath, currentPath)
				newPath = append(newPath, link)
				queue = append(queue, newPath)
			}
		}
	}
	
	return paths
}
