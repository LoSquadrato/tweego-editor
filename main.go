package main

import (
	"encoding/json"
	"fmt"
	"log"
	"tweego-editor/compiler"
	"tweego-editor/formats/harlowe"
	"tweego-editor/parser"
)

func main() {
	fmt.Println("Tweego Editor Backend v0.1.0")
	fmt.Println("================================\n")
	
	// Test del parser
	fmt.Println("📖 Testing Parser...")
	testParser()
	
	fmt.Println("\n================================\n")
	
	// Test del compiler
	fmt.Println("⚙️  Testing Tweego Compiler...")
	testCompiler()
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
	
	// Output JSON della storia completa (solo per debug)
	fmt.Println("\n=== JSON Output (primi 500 caratteri) ===")
	jsonData, err := json.MarshalIndent(story, "", "  ")
	if err != nil {
		log.Printf("Errore serializzazione JSON: %v", err)
		return
	}
	jsonStr := string(jsonData)
	if len(jsonStr) > 500 {
		jsonStr = jsonStr[:500] + "..."
	}
	fmt.Println(jsonStr)
}

func testCompiler() {
	// Crea il compiler (cerca tweego nel PATH)
	tweegoCompiler, err := compiler.NewTweegoCompiler("", "./output")
	if err != nil {
		log.Fatalf("❌ Errore inizializzazione compiler: %v", err)
	}
	
	// Ottieni versione di Tweego
	version, err := tweegoCompiler.GetVersion()
	if err != nil {
		log.Printf("⚠️  Impossibile ottenere versione: %v", err)
	} else {
		fmt.Printf("✓ Tweego version: %s\n", version)
	}
	
	// Elenca formati disponibili
	formats, err := tweegoCompiler.ListFormats()
	if err != nil {
		log.Printf("⚠️  Impossibile elencare formati: %v", err)
	} else {
		fmt.Printf("✓ Formati disponibili:\n")
		for _, format := range formats {
			fmt.Printf("  - %s\n", format)
		}
	}
	
	fmt.Println("\n📦 Compilazione test_story.twee...")
	
	// Compila la storia
	result, err := tweegoCompiler.Compile("test_story.twee", &compiler.CompileOptions{
		Format: "harlowe-3", // Usa Harlowe 3
		Output: "test_output.html",
	})
	
	if err != nil {
		log.Printf("❌ Errore compilazione: %v", err)
		if result != nil && result.ErrorMessage != "" {
			fmt.Printf("\nDettagli errore:\n%s\n", result.ErrorMessage)
		}
		return
	}
	
	// Mostra risultato
	if result.Success {
		fmt.Printf("✅ Compilazione completata con successo!\n")
		fmt.Printf("   Output: %s\n", result.OutputFile)
		
		if len(result.Warnings) > 0 {
			fmt.Printf("\n⚠️  Warning (%d):\n", len(result.Warnings))
			for _, warning := range result.Warnings {
				fmt.Printf("   - %s\n", warning)
			}
		}
		
		if result.Output != "" {
			fmt.Printf("\n📄 Output Tweego:\n%s\n", result.Output)
		}
	} else {
		fmt.Printf("❌ Compilazione fallita\n")
		if result.ErrorMessage != "" {
			fmt.Printf("Errore: %s\n", result.ErrorMessage)
		}
	}
}