package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tweego-editor/compiler"
	"tweego-editor/formats"
	_ "tweego-editor/formats/harlowe" // Registra il formato Harlowe
	"tweego-editor/parser"
)

// TestRunner gestisce i test batch per formato
type TestRunner struct {
	baseDir      string
	compiler     *compiler.TweegoWrapper
	format       string
	formatDir    string
	formatParser formats.StoryFormat
}

// ParsedOutput rappresenta l'output completo del parsing
type ParsedOutput struct {
	Filename string       `json:"filename"`
	ParsedAt string       `json:"parsed_at"`
	Success  bool         `json:"success"`
	Error    string       `json:"error,omitempty"`
	Story    *StoryOutput `json:"story,omitempty"`
}

// StoryOutput output della storia parsata
type StoryOutput struct {
	Title         string                    `json:"title"`
	Format        string                    `json:"format"`
	FormatVersion string                    `json:"format_version"`
	PassageCount  int                       `json:"passage_count"`
	Passages      map[string]*PassageOutput `json:"passages"`
}

// PassageOutput output di un singolo passaggio
type PassageOutput struct {
	Title     string                  `json:"title"`
	Tags      []string                `json:"tags"`
	Content   string                  `json:"content"`
	Links     []string                `json:"links"`
	Variables map[string]interface{}  `json:"variables"`
	Preview   string                  `json:"preview"`
	Literals  *formats.LiteralsResult `json:"literals"`
}

// CompiledOutput rappresenta l'output della compilazione
type CompiledOutput struct {
	Filename   string   `json:"filename"`
	CompiledAt string   `json:"compiled_at"`
	Success    bool     `json:"success"`
	Error      string   `json:"error,omitempty"`
	OutputFile string   `json:"output_file,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
}

// TestSummary riassunto dei test
type TestSummary struct {
	Format         string `json:"format"`
	TotalFiles     int    `json:"total_files"`
	ParseSuccess   int    `json:"parse_success"`
	ParseFailed    int    `json:"parse_failed"`
	CompileSuccess int    `json:"compile_success"`
	CompileFailed  int    `json:"compile_failed"`
	Duration       string `json:"duration"`
}

// NewTestRunner crea un nuovo test runner
func NewTestRunner(baseDir string, comp *compiler.TweegoWrapper) *TestRunner {
	return &TestRunner{
		baseDir:  baseDir,
		compiler: comp,
	}
}

// GetAvailableFormats restituisce i formati con cartelle test disponibili
func (tr *TestRunner) GetAvailableFormats() ([]string, error) {
	entries, err := os.ReadDir(tr.baseDir)
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere cartella test: %w", err)
	}

	var formats []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			formats = append(formats, entry.Name())
		}
	}

	return formats, nil
}

// RunTests esegue i test per un formato specifico
func (tr *TestRunner) RunTests(format string) (*TestSummary, error) {
	startTime := time.Now()

	tr.format = format
	tr.formatDir = filepath.Join(tr.baseDir, format)

	// Verifica che la cartella esista
	if _, err := os.Stat(tr.formatDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("cartella test/%s non trovata", format)
	}

	// Ottieni il parser per questo formato
	tr.formatParser = formats.GetRegisteredFormat(format)
	if tr.formatParser == nil {
		return nil, fmt.Errorf("parser per formato '%s' non registrato", format)
	}

	fmt.Printf("🔧 Usando parser: %s\n", tr.formatParser.GetFormatName())

	// Trova tutti i file .twee
	tweeFiles, err := tr.findTweeFiles()
	if err != nil {
		return nil, err
	}

	if len(tweeFiles) == 0 {
		return nil, fmt.Errorf("nessun file .twee trovato in test/%s", format)
	}

	summary := &TestSummary{
		Format:     format,
		TotalFiles: len(tweeFiles),
	}

	fmt.Printf("\n📁 Trovati %d file .twee in test/%s\n", len(tweeFiles), format)
	fmt.Println(strings.Repeat("─", 50))

	// Processa ogni file
	for _, tweeFile := range tweeFiles {
		filename := filepath.Base(tweeFile)
		fmt.Printf("\n📄 %s\n", filename)

		// 1. Parsing
		parseResult := tr.parseFile(tweeFile)
		if parseResult.Success {
			summary.ParseSuccess++
			fmt.Printf("   ✅ Parsing OK - %d passaggi\n", parseResult.Story.PassageCount)
			
			// Mostra formato estratto
			if parseResult.Story.Format != "" {
				fmt.Printf("   📋 Formato: %s", parseResult.Story.Format)
				if parseResult.Story.FormatVersion != "" {
					fmt.Printf(" v%s", parseResult.Story.FormatVersion)
				}
				fmt.Println()
			}
			
			// Conta i literals totali
			totalArrays, totalDatamaps, totalDatasets := tr.countLiterals(parseResult)
			if totalArrays > 0 || totalDatamaps > 0 || totalDatasets > 0 {
				fmt.Printf("   📊 Literals: %d arrays, %d datamaps, %d datasets\n",
					totalArrays, totalDatamaps, totalDatasets)
			}
		} else {
			summary.ParseFailed++
			fmt.Printf("   ❌ Parsing FAILED: %s\n", parseResult.Error)
		}

		// Salva JSON parsing
		parseJSONPath := tr.getOutputPath(tweeFile, "_parsed.json")
		if err := tr.saveJSON(parseJSONPath, parseResult); err != nil {
			fmt.Printf("   ⚠️  Errore salvataggio JSON: %v\n", err)
		} else {
			fmt.Printf("   💾 %s\n", filepath.Base(parseJSONPath))
		}

		// 2. Compilazione (solo se parsing è riuscito)
		var compileResult *CompiledOutput
		if parseResult.Success {
			// Usa il formato estratto dal file
			compileResult = tr.compileFile(tweeFile, parseResult.Story.Format, parseResult.Story.FormatVersion)
			
			if compileResult.Success {
				summary.CompileSuccess++
				fmt.Printf("   ✅ Compilazione OK → %s\n", filepath.Base(compileResult.OutputFile))
				if len(compileResult.Warnings) > 0 {
					fmt.Printf("   ⚠️  %d warning(s)\n", len(compileResult.Warnings))
				}
			} else {
				summary.CompileFailed++
				fmt.Printf("   ❌ Compilazione FAILED: %s\n", compileResult.Error)
			}
		} else {
			// Se il parsing è fallito, la compilazione fallisce automaticamente
			compileResult = &CompiledOutput{
				Filename:   filename,
				CompiledAt: time.Now().Format(time.RFC3339),
				Success:    false,
				Error:      "Parsing fallito - compilazione saltata",
			}
			summary.CompileFailed++
		}

		// Salva JSON compilazione
		compileJSONPath := tr.getOutputPath(tweeFile, "_compiled.json")
		if err := tr.saveJSON(compileJSONPath, compileResult); err != nil {
			fmt.Printf("   ⚠️  Errore salvataggio log: %v\n", err)
		} else {
			fmt.Printf("   💾 %s\n", filepath.Base(compileJSONPath))
		}
	}

	summary.Duration = time.Since(startTime).String()

	// Stampa riassunto
	fmt.Println()
	fmt.Println(strings.Repeat("═", 50))
	fmt.Printf("📊 RIASSUNTO TEST - %s\n", strings.ToUpper(format))
	fmt.Println(strings.Repeat("═", 50))
	fmt.Printf("   File testati:     %d\n", summary.TotalFiles)
	fmt.Printf("   Parsing OK:       %d/%d\n", summary.ParseSuccess, summary.TotalFiles)
	fmt.Printf("   Compilazione OK:  %d/%d\n", summary.CompileSuccess, summary.TotalFiles)
	fmt.Printf("   Durata:           %s\n", summary.Duration)
	fmt.Println(strings.Repeat("═", 50))

	return summary, nil
}

// findTweeFiles trova tutti i file .twee nella cartella del formato
func (tr *TestRunner) findTweeFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(tr.formatDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".twee") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// parseFile parsa un singolo file .twee
func (tr *TestRunner) parseFile(filePath string) *ParsedOutput {
	result := &ParsedOutput{
		Filename: filepath.Base(filePath),
		ParsedAt: time.Now().Format(time.RFC3339),
	}

	// Valida e parsa
	tweeParser := parser.NewTweeParser(filePath)
	story, err := tweeParser.Parse()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	result.Success = true

	// Converti in output
	storyOutput := &StoryOutput{
		Title:         story.Title,
		Format:        story.Format,
		FormatVersion: story.FormatVersion,
		PassageCount:  len(story.Passages),
		Passages:      make(map[string]*PassageOutput),
	}

	// Processa ogni passaggio usando il parser del formato
	for title, passage := range story.Passages {
		// Estrai literals usando il metodo del parser (tutta la logica è nel formato!)
		literals := tr.formatParser.ExtractAllLiterals(passage.Content)

		passageOutput := &PassageOutput{
			Title:     passage.Title,
			Tags:      passage.Tags,
			Content:   passage.Content,
			Links:     tr.formatParser.ParseLinks(passage.Content),
			Variables: tr.formatParser.ParseVariables(passage.Content),
			Preview:   tr.formatParser.StripCode(passage.Content),
			Literals:  literals,
		}

		storyOutput.Passages[title] = passageOutput
	}

	result.Story = storyOutput
	return result
}

// compileFile compila un singolo file .twee usando il formato estratto
func (tr *TestRunner) compileFile(filePath string, storyFormat string, formatVersion string) *CompiledOutput {
	result := &CompiledOutput{
		Filename:   filepath.Base(filePath),
		CompiledAt: time.Now().Format(time.RFC3339),
	}

	// Determina il formato da usare per la compilazione
	compileFormat := tr.determineCompileFormat(storyFormat, formatVersion)
	
	if compileFormat == "" {
		result.Success = false
		result.Error = fmt.Sprintf("Impossibile determinare formato di compilazione da: format='%s', version='%s'", 
			storyFormat, formatVersion)
		return result
	}

	// Output file
	
	outputFile := tr.getOutputPath(filePath, ".html")

	// Compila
	compileResult, err := tr.compiler.Compile(filePath, &compiler.CompileOptions{
		Format: compileFormat,
		Output: outputFile,
	})

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		if compileResult != nil && compileResult.ErrorMessage != "" {
			result.Error = compileResult.ErrorMessage
		}
		return result
	}

	result.Success = compileResult.Success
	result.OutputFile = compileResult.OutputFile
	result.Warnings = compileResult.Warnings

	return result
}

// determineCompileFormat converte il formato estratto nel formato per Tweego
// Es: "harlowe" + "3.2.3" -> "harlowe-3"
//     "sugarcube" + "2.36.1" -> "sugarcube-2"
func (tr *TestRunner) determineCompileFormat(format string, version string) string {
	if format == "" {
		return ""
	}

	// Normalizza il formato (lowercase)
	format = strings.ToLower(format)
	
	// Se non c'è versione, usa il formato base
	if version == "" {
		return format
	}

	// Estrai la versione major (primo numero)
	versionParts := strings.Split(version, ".")
	if len(versionParts) == 0 {
		return format
	}

	majorVersion := versionParts[0]
	
	// Costruisci il formato per Tweego: "formato-major"
	// Es: "harlowe-3", "sugarcube-2"
	return fmt.Sprintf("%s-%s", format, majorVersion)
}

// getOutputPath genera il percorso del file di output
func (tr *TestRunner) getOutputPath(inputFile string, suffix string) string {
	baseName := strings.TrimSuffix(filepath.Base(inputFile), ".twee")
	return filepath.Join(tr.formatDir, baseName+suffix)
}

// saveJSON salva un oggetto in formato JSON
func (tr *TestRunner) saveJSON(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

// countLiterals conta i literals totali in un ParsedOutput
func (tr *TestRunner) countLiterals(result *ParsedOutput) (int, int, int) {
	totalArrays := 0
	totalDatamaps := 0
	totalDatasets := 0

	if result.Story != nil {
		for _, passage := range result.Story.Passages {
			if passage.Literals != nil {
				totalArrays += len(passage.Literals.Arrays)
				totalDatamaps += len(passage.Literals.Datamaps)
				totalDatasets += len(passage.Literals.Datasets)
			}
		}
	}

	return totalArrays, totalDatamaps, totalDatasets
}