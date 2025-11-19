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

// NewTweeParser crea un nuovo parser
func NewTweeParser(filepath string) *TweeParser {
	return &TweeParser{filepath: filepath}
}

// Parse legge e parsa il file .twee
func (tp *TweeParser) Parse() (*Story, error) {
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