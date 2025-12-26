package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"tweego-editor/api"
	"tweego-editor/compiler"
	"tweego-editor/formats/harlowe"
	"tweego-editor/parser"
	"tweego-editor/test"
	"tweego-editor/watcher"

	_ "tweego-editor/formats/harlowe"
)

func main() {
	fmt.Println("Tweego Editor Backend v0.1.0")
	fmt.Println("================================\n")
	
	// Mostra menu
	fmt.Println("Scegli una modalità:")
	fmt.Println("1. Test Parser")
	fmt.Println("2. Test Compiler")
	fmt.Println("3. Watch Mode (auto-ricompila)")
	fmt.Println("4. API Server (REST + WebSocket)")
	fmt.Println("5. Test Batch (per formato)")
	fmt.Print("\nScelta (1/2/3/4/5): ")
	
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	fmt.Println()
	
	switch choice {
	case "1":
		fmt.Println("📖 Testing Parser...")
		testParser()
	case "2":
		fmt.Println("⚙️  Testing Tweego Compiler...")
		testCompiler()
	case "3":
		fmt.Println("👀 Starting Watch Mode...")
		testWatcher()
	case "4":
		fmt.Println("🌐 Starting API Server...")
		startAPIServer()
	case "5":
		fmt.Println("🧪 Test Batch per Formato...")
		testBatch()
	default:
		fmt.Println("❌ Scelta non valida")
	}
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
	// Crea il wrapper per Tweego
	tweegoWrapper, err := compiler.NewTweegoWrapper("", "./output")
	if err != nil {
		log.Fatalf("❌ Errore inizializzazione wrapper: %v", err)
	}
	
	// Ottieni versione di Tweego
	version, err := tweegoWrapper.GetVersion()
	if err != nil {
		log.Printf("⚠️  Impossibile ottenere versione: %v", err)
	} else {
		fmt.Printf("✓ Tweego version: %s\n", version)
	}
	
	// Elenca formati disponibili
	formats, err := tweegoWrapper.ListFormats()
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
	result, err := tweegoWrapper.Compile("test_story.twee", &compiler.CompileOptions{
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

func testWatcher() {
	// Crea il wrapper per Tweego
	tweegoWrapper, err := compiler.NewTweegoWrapper("", "./output")
	if err != nil {
		log.Fatalf("❌ Errore inizializzazione wrapper: %v", err)
	}
	
	// Configurazione watcher
	config := watcher.WatcherConfig{
		Paths: []string{"."}, // Monitora directory corrente
		Compiler: tweegoWrapper,
		CompileOpts: &compiler.CompileOptions{
			Format: "harlowe-3",
			Output: "test_output.html",
		},
		AutoCompile: true,
	}
	
	// Crea e avvia il watcher
	fw, err := watcher.NewFileWatcher(config)
	if err != nil {
		log.Fatalf("❌ Errore creazione watcher: %v", err)
	}
	
	if err := fw.Start(); err != nil {
		log.Fatalf("❌ Errore avvio watcher: %v", err)
	}
	
	fmt.Println("\n✨ Watch mode attivo!")
	fmt.Println("💡 Modifica test_story.twee per vedere la ricompilazione automatica")
	fmt.Println("🛑 Premi CTRL+C per uscire\n")
	
	// Ascolta eventi
	for event := range fw.Events() {
		fmt.Printf("📢 Evento: %s - %s\n", event.Type, event.Path)
	}
}

func startAPIServer() {
	// Crea il wrapper per Tweego
	tweegoWrapper, err := compiler.NewTweegoWrapper("", "./output")
	if err != nil {
		log.Fatalf("❌ Errore inizializzazione wrapper: %v", err)
	}
	
	// Configurazione server
	config := api.ServerConfig{
		Port:       8080,
		Compiler:   tweegoWrapper,
		EnableCORS: true,
		Debug:      true,
	}
	
	// Crea e avvia server
	server := api.NewServer(config)
	
	fmt.Println("\n✨ API Server ready!")
	fmt.Println("📚 Documentazione endpoint:")
	fmt.Println("   GET  /api/health")
	fmt.Println("   POST /api/story/validate")
	fmt.Println("   POST /api/story/parse")
	fmt.Println("   POST /api/story/compile")
	fmt.Println("   GET  /api/story/:file/passages")
	fmt.Println("   GET  /api/story/:file/passage/:title")
	fmt.Println("   POST /api/simulator/validate")
	fmt.Println("   POST /api/simulator/simulate")
	fmt.Println("   POST /api/simulator/suggest")
	fmt.Println("   POST /api/watch/start")
	fmt.Println("   POST /api/watch/stop")
	fmt.Println("   GET  /api/watch/status")
	fmt.Println("   GET  /api/formats")
	fmt.Println("   GET  /api/version")
	fmt.Println("   WS   /ws (WebSocket)")
	fmt.Println()
	
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Errore avvio server: %v", err)
	}
}

func testBatch() {
	reader := bufio.NewReader(os.Stdin)
	
	// Crea il wrapper per Tweego
	tweegoWrapper, err := compiler.NewTweegoWrapper("", "")
	if err != nil {
		log.Fatalf("❌ Errore inizializzazione wrapper: %v", err)
	}

	// Crea test runner
	runner := test.NewTestRunner("./test", tweegoWrapper)

	// Ottieni formati disponibili
	formats, err := runner.GetAvailableFormats()
	if err != nil {
		log.Fatalf("❌ Errore lettura cartelle test: %v", err)
	}

	if len(formats) == 0 {
		fmt.Println("❌ Nessuna cartella di test trovata in ./test")
		fmt.Println("   Crea una cartella come ./test/harlowe con file .twee")
		return
	}

	// Mostra formati disponibili
	fmt.Println("📂 Cartelle test disponibili:")
	for i, format := range formats {
		fmt.Printf("   %d. %s\n", i+1, format)
	}
	fmt.Print("\nScegli formato (numero o nome): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Determina il formato scelto
	var selectedFormat string
	
	// Prova a interpretare come numero
	var num int
	if _, err := fmt.Sscanf(input, "%d", &num); err == nil && num > 0 && num <= len(formats) {
		selectedFormat = formats[num-1]
	} else {
		// Cerca per nome
		for _, f := range formats {
			if strings.EqualFold(f, input) {
				selectedFormat = f
				break
			}
		}
	}

	if selectedFormat == "" {
		fmt.Printf("❌ Formato '%s' non trovato\n", input)
		return
	}

	fmt.Printf("\n🚀 Avvio test per formato: %s\n", selectedFormat)

	// Esegui i test
	summary, err := runner.RunTests(selectedFormat)
	if err != nil {
		log.Fatalf("❌ Errore esecuzione test: %v", err)
	}

	// Risultato finale
	if summary.ParseFailed == 0 && summary.CompileFailed == 0 {
		fmt.Println("\n✅ Tutti i test passati!")
	} else {
		fmt.Println("\n⚠️  Alcuni test falliti - controlla i file JSON per i dettagli")
	}
}