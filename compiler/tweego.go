package compiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TweegoCompiler gestisce la compilazione con Tweego
type TweegoCompiler struct {
	tweegoPath string
	workDir    string
}

// CompileOptions opzioni per la compilazione
type CompileOptions struct {
	Format       string   // Story format (es: "harlowe-3.2.3")
	Output       string   // File output (default: "output.html")
	Head         string   // File HTML da includere in <head>
	StartNode    string   // Passaggio iniziale
	StrictMode   bool     // Modalità strict (warnings = errors)
	Watch        bool     // Watch mode (ricompila automaticamente)
	AdditionalArgs []string // Argomenti aggiuntivi per Tweego
}

// CompileResult risultato della compilazione
type CompileResult struct {
	Success      bool
	Output       string
	ErrorMessage string
	Warnings     []string
	OutputFile   string
}

// NewTweegoCompiler crea un nuovo compiler
func NewTweegoCompiler(tweegoPath string, workDir string) (*TweegoCompiler, error) {
	// Se tweegoPath è vuoto, cerca tweego nel PATH
	if tweegoPath == "" {
		path, err := exec.LookPath("tweego")
		if err != nil {
			return nil, fmt.Errorf("tweego non trovato nel PATH. Installalo da https://www.motoslave.net/tweego/")
		}
		tweegoPath = path
	}

	// Verifica che tweego sia eseguibile
	if _, err := os.Stat(tweegoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tweego non trovato in: %s", tweegoPath)
	}

	// Se workDir non esiste, crealo
	if workDir != "" {
		if err := os.MkdirAll(workDir, 0755); err != nil {
			return nil, fmt.Errorf("impossibile creare workDir: %w", err)
		}
	}

	return &TweegoCompiler{
		tweegoPath: tweegoPath,
		workDir:    workDir,
	}, nil
}

// Compile compila un file .twee in HTML
func (tc *TweegoCompiler) Compile(inputFile string, options *CompileOptions) (*CompileResult, error) {
	result := &CompileResult{
		Success: false,
	}

	// Opzioni di default
	if options == nil {
		options = &CompileOptions{
			Output: "output.html",
		}
	}

	// Costruisci gli argomenti per tweego
	args := []string{}

	// Output file
	outputPath := options.Output
	if !filepath.IsAbs(outputPath) && tc.workDir != "" {
		outputPath = filepath.Join(tc.workDir, outputPath)
	}
	args = append(args, "-o", outputPath)

	// Story format
	if options.Format != "" {
		args = append(args, "-f", options.Format)
	}

	// Head file
	if options.Head != "" {
		args = append(args, "--head", options.Head)
	}

	// Start node
	if options.StartNode != "" {
		args = append(args, "-s", options.StartNode)
	}

	// Strict mode
	if options.StrictMode {
		args = append(args, "--strict")
	}

	// Watch mode
	if options.Watch {
		args = append(args, "-w")
	}

	// Argomenti aggiuntivi
	args = append(args, options.AdditionalArgs...)

	// File input (sempre per ultimo)
	args = append(args, inputFile)

	// Esegui tweego
	cmd := exec.Command(tc.tweegoPath, args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Esegui il comando
	err := cmd.Run()

	// Cattura output
	result.Output = stdout.String()
	stderrStr := stderr.String()

	// Processa warnings
	if stderrStr != "" {
		lines := strings.Split(stderrStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				if strings.Contains(strings.ToLower(line), "warning") {
					result.Warnings = append(result.Warnings, line)
				} else if strings.Contains(strings.ToLower(line), "error") {
					result.ErrorMessage += line + "\n"
				}
			}
		}
	}

	if err != nil {
		result.Success = false
		if result.ErrorMessage == "" {
			result.ErrorMessage = fmt.Sprintf("Errore esecuzione tweego: %v\n%s", err, stderrStr)
		}
		return result, fmt.Errorf("compilazione fallita: %w", err)
	}

	result.Success = true
	result.OutputFile = outputPath

	return result, nil
}

// GetVersion ritorna la versione di Tweego installata
func (tc *TweegoCompiler) GetVersion() (string, error) {
	cmd := exec.Command(tc.tweegoPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("impossibile ottenere versione tweego: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ListFormats elenca i formati disponibili
func (tc *TweegoCompiler) ListFormats() ([]string, error) {
	cmd := exec.Command(tc.tweegoPath, "--list-formats")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("impossibile elencare formati: %w", err)
	}

	formats := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "Story formats:") {
			formats = append(formats, line)
		}
	}

	return formats, nil
}