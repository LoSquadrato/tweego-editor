package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TweeParser gestisce il parsing dei file .twee
type TweeParser struct {
	filepath string
}

// ValidationError rappresenta un errore di validazione
type ValidationError struct {
	Type    string `json:"type"`    // "error" o "warning"
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

// ValidationResult risultato della validazione
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// NewTweeParser crea un nuovo parser
func NewTweeParser(filepath string) *TweeParser {
	return &TweeParser{filepath: filepath}
}

// Validate valida il file .twee prima del parsing
func (tp *TweeParser) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// 1. Verifica che il file esista
	if _, err := os.Stat(tp.filepath); os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "error",
			Message: fmt.Sprintf("File non trovato: %s", tp.filepath),
		})
		return result
	}

	// 2. Verifica che il file sia leggibile
	file, err := os.Open(tp.filepath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "error",
			Message: fmt.Sprintf("Impossibile leggere il file: %v", err),
		})
		return result
	}
	defer file.Close()

	// 3. Verifica sintassi e contenuto base
	scanner := bufio.NewScanner(file)
	lineNum := 0
	hasPassages := false
	hasStoryData := false
	startPassage := ""
	foundPassages := map[string]bool{}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Rileva passaggi
		if strings.HasPrefix(line, "::") {
			hasPassages = true

			// Estrai titolo passaggio
			passageHeaderRegex := regexp.MustCompile(`^::\s*(.+?)(?:\s+\[([^\]]*)\])?(?:\s+\{.*\})?$`)
			matches := passageHeaderRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				title := strings.TrimSpace(matches[1])
				foundPassages[title] = true

				// Rileva StoryData
				if title == "StoryData" {
					hasStoryData = true
				}
			}
		}

		// Estrai start passage da StoryData
		if hasStoryData && strings.Contains(line, `"start"`) {
			startRegex := regexp.MustCompile(`"start"\s*:\s*"([^"]+)"`)
			matches := startRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				startPassage = matches[1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "error",
			Message: fmt.Sprintf("Errore lettura file: %v", err),
		})
		return result
	}

	// 4. Verifica che ci siano passaggi
	if !hasPassages {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "error",
			Message: "Nessun passaggio trovato nel file .twee",
		})
	}

	// 5. Verifica che il passaggio Start esista (se definito)
	if startPassage != "" && !foundPassages[startPassage] {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "error",
			Message: fmt.Sprintf("Passaggio iniziale '%s' definito in StoryData ma non trovato", startPassage),
		})
	}

	// 6. Warning se non c'è StoryData
	if !hasStoryData {
		result.Warnings = append(result.Warnings, ValidationError{
			Type:    "warning",
			Message: "Nessun passaggio StoryData trovato (opzionale ma raccomandato)",
		})
	}

	return result
}

// Parse legge e parsa il file .twee
func (tp *TweeParser) Parse() (*Story, error) {
	// Valida prima di parsare
	validation := tp.Validate()
	if !validation.Valid {
		errMsg := "Validazione fallita:\n"
		for _, err := range validation.Errors {
			errMsg += fmt.Sprintf("  - %s\n", err.Message)
		}
		return nil, fmt.Errorf(errMsg)
	}

	file, err := os.Open(tp.filepath)
	if err != nil {
		return nil, fmt.Errorf("errore apertura file: %w", err)
	}
	defer file.Close()

	story := &Story{
		Passages: make(map[string]*Passage),
	}

	scanner := bufio.NewScanner(file)
	var currentPassage *Passage
	var contentBuilder strings.Builder

	// Regex per il formato :: Title [tags] {"position":"x,y"}
	passageHeaderRegex := regexp.MustCompile(`^::\s*(.+?)(?:\s+\[([^\]]*)\])?(?:\s+\{.*\})?$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Nuova intestazione passaggio
		if strings.HasPrefix(line, "::") {
			// Salva il passaggio precedente se esiste
			if currentPassage != nil {
				currentPassage.Content = strings.TrimSpace(contentBuilder.String())
				story.Passages[currentPassage.Title] = currentPassage
				contentBuilder.Reset()
			}

			// Parsa la nuova intestazione
			matches := passageHeaderRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentPassage = &Passage{
					Title:    strings.TrimSpace(matches[1]),
					Tags:     []string{},
					Metadata: make(map[string]string),
				}

				// Estrai i tag se presenti
				if len(matches) > 2 && matches[2] != "" {
					tags := strings.Split(matches[2], " ")
					for _, tag := range tags {
						tag = strings.TrimSpace(tag)
						if tag != "" {
							currentPassage.Tags = append(currentPassage.Tags, tag)
						}
					}
				}
			}
		} else if currentPassage != nil {
			// Aggiungi al contenuto del passaggio corrente
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	// Salva l'ultimo passaggio
	if currentPassage != nil {
		currentPassage.Content = strings.TrimSpace(contentBuilder.String())
		story.Passages[currentPassage.Title] = currentPassage
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("errore lettura file: %w", err)
	}

	return story, nil
}