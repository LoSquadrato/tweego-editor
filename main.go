// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"tweego-editor/formats/harlowe"
	"tweego-editor/parser"
)

func main() {
	fmt.Println("Tweego Editor Backend v0.1.0")
	fmt.Println("Testing parser...\n")
	
	// Test del parser
	testParser()
}

func testParser() {
	// Crea un parser per il file di test
	tweeParser := parser.NewTweeParser("test_story.twee")
	
	// Parsa il file
	story, err := tweeParser.Parse()
	if err != nil {
		log.Fatalf("Errore nel parsing: %v", err)
	}
	
	fmt.Printf("✓ Storia parsata con successo!\n")
	fmt.Printf("  Passaggi trovati: %d\n\n", len(story.Passages))
	
	// Inizializza il formato Harlowe
	harlowe := harlowe.NewHarloweFormat()
	
	// Analizza ogni passaggio
	for title, passage := range story.Passages {
		fmt.Printf("=== Passaggio: %s ===\n", title)
		fmt.Printf("Tag: %v\n", passage.Tags)
		
		// Estrai link
		links := harlowe.ParseLinks(passage.Content)
		fmt.Printf("Link trovati: %v\n", links)
		
		// Estrai variabili
		variables := harlowe.ParseVariables(passage.Content)
		if len(variables) > 0 {
			fmt.Printf("Variabili: %v\n", variables)
		}
		
		// Mostra anteprima pulita
		preview := harlowe.StripCode(passage.Content)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		fmt.Printf("Anteprima: %s\n", preview)
		fmt.Println()
	}
	
	// Output JSON della storia completa
	fmt.Println("\n=== JSON Output ===")
	jsonData, err := json.MarshalIndent(story, "", "  ")
	if err != nil {
		log.Printf("Errore serializzazione JSON: %v", err)
		return
	}
	fmt.Println(string(jsonData))
}