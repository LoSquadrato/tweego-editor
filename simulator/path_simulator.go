package simulator

import (
	"fmt"
	"tweego-editor/formats"
	"tweego-editor/parser"
)

// PathSimulator simula un percorso attraverso la storia
type PathSimulator struct {
	story           *parser.Story
	format          formats.StoryFormat
	visitedPassages map[string]int
	history         []string
}

// VariableChange rappresenta il cambiamento di una variabile
type VariableChange struct {
	Name     string      `json:"name"`
	Previous interface{} `json:"previous"`
	Current  interface{} `json:"current"`
	Delta    interface{} `json:"delta,omitempty"`
}

// StepResult risultato di un singolo step
type StepResult struct {
	PassageTitle   string                    `json:"passage_title"`
	PassageIndex   int                       `json:"passage_index"`
	Changes        map[string]VariableChange `json:"changes"`
	Warnings       []string                  `json:"warnings,omitempty"`
	AvailableLinks []string                  `json:"available_links"`
}

// SimulationResult risultato completo della simulazione
type SimulationResult struct {
	Success       bool                   `json:"success"`
	Path          []string               `json:"path"`
	Steps         []StepResult           `json:"steps"`
	FinalState    map[string]interface{} `json:"final_state"`
	Errors        []string               `json:"errors,omitempty"`
	TotalWarnings int                    `json:"total_warnings"`
}

// NewPathSimulator crea un nuovo simulatore
func NewPathSimulator(story *parser.Story) *PathSimulator {
	formatName := story.Format
	if formatName == "" {
		formatName = "harlowe"
	}

	format := formats.GetRegisteredFormat(formatName)
	if format == nil {
		format = formats.GetRegisteredFormat("harlowe")
	}

	return &PathSimulator{
		story:           story,
		format:          format,
		visitedPassages: make(map[string]int),
		history:         []string{},
	}
}

// ValidatePath verifica che il path sia valido
func (ps *PathSimulator) ValidatePath(path []string) []string {
	errors := []string{}

	for i, passageTitle := range path {
		if _, exists := ps.story.Passages[passageTitle]; !exists {
			errors = append(errors, fmt.Sprintf("Step %d: passaggio '%s' non esiste", i+1, passageTitle))
		}
	}

	for i := 0; i < len(path)-1; i++ {
		currentTitle := path[i]
		nextTitle := path[i+1]

		passage, exists := ps.story.Passages[currentTitle]
		if !exists {
			continue
		}

		links := ps.format.ParseLinks(passage.Content)

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
		Success:    true,
		Path:       path,
		Steps:      []StepResult{},
		FinalState: make(map[string]interface{}),
		Errors:     []string{},
	}

	validationErrors := ps.ValidatePath(path)
	if len(validationErrors) > 0 {
		result.Success = false
		result.Errors = validationErrors
		return result
	}

	// Reset history e visited
	ps.visitedPassages = make(map[string]int)
	ps.history = []string{}

	// Stato corrente delle variabili
	currentState := make(map[string]interface{})

	// Simula ogni passaggio
	for i, passageTitle := range path {
		passage, exists := ps.story.Passages[passageTitle]
		if !exists {
			continue
		}

		// 1. Aggiorna history e visited
		ps.visitedPassages[passageTitle]++
		ps.history = append(ps.history, passageTitle)

		stepResult := StepResult{
			PassageTitle:   passageTitle,
			PassageIndex:   i + 1,
			Changes:        make(map[string]VariableChange),
			Warnings:       []string{},
			AvailableLinks: ps.format.ParseLinks(passage.Content),
		}

		// 2. Salva stato PRIMA del processing
		stateBefore := ps.copyState(currentState)

		// 3. Crea evaluator con lo stato corrente
		eval := ps.format.CreateEvaluator(currentState)
		eval.SetVisitedPassages(ps.visitedPassages)
		eval.SetHistory(ps.history)
		eval.SetCurrentPassage(passageTitle)

		// 4. CHIAVE: Processa il contenuto usando il formato
		//    Questo modifica lo stato dell'evaluator
		if err := ps.format.ProcessPassageContent(passage.Content, eval); err != nil {
			// Log error ma continua
			fmt.Printf("⚠️  Warning processing passage %s: %v\n", passageTitle, err)
		}

		// 5. Ottieni il nuovo stato dall'evaluator
		newState := eval.GetState()

		// 6. Calcola i cambiamenti
		for varName, newValue := range newState {
			previousValue, existed := stateBefore[varName]

			change := VariableChange{
				Name:     varName,
				Previous: previousValue,
				Current:  newValue,
			}

			if existed {
				prevNum, prevIsNum := toNumber(previousValue)
				currNum, currIsNum := toNumber(newValue)

				if prevIsNum && currIsNum {
					change.Delta = currNum - prevNum
				}
			} else {
				change.Previous = nil
			}

			stepResult.Changes[varName] = change
		}

		// 7. Aggiorna lo stato corrente
		currentState = newState

		// 8. Genera warnings
		stepResult.Warnings = ps.generateWarnings(passage, currentState, stepResult.Changes)
		result.TotalWarnings += len(stepResult.Warnings)

		result.Steps = append(result.Steps, stepResult)
	}

	result.FinalState = currentState

	return result
}

// copyState crea una copia profonda dello stato
func (ps *PathSimulator) copyState(state map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range state {
		copy[k] = v
	}
	return copy
}

// generateWarnings genera warning per un passaggio
func (ps *PathSimulator) generateWarnings(passage *parser.Passage, currentState map[string]interface{}, changes map[string]VariableChange) []string {
	warnings := []string{}

	for varName, change := range changes {
		if varName == "vita" || varName == "health" || varName == "hp" {
			if val, isNum := toNumber(change.Current); isNum {
				if val <= 0 {
					warnings = append(warnings, fmt.Sprintf("⚠️ %s è a 0 o negativo: %v", varName, val))
				} else if val <= 20 {
					warnings = append(warnings, fmt.Sprintf("⚠️ %s è sotto soglia critica: %v", varName, val))
				}
			}
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
		var num float64
		_, err := fmt.Sscanf(v, "%f", &num)
		return num, err == nil
	}
	return 0, false
}

// GetSuggestedPaths suggerisce percorsi validi dato un punto di partenza
func (ps *PathSimulator) GetSuggestedPaths(startPassage string, maxDepth int) [][]string {
	paths := [][]string{}

	queue := [][]string{{startPassage}}

	for len(queue) > 0 && len(paths) < 10 {
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
			paths = append(paths, currentPath)
		} else {
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