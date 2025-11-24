package compiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TweegoWrapper gestisce l'integrazione con Tweego (wrapper esterno)
type TweegoWrapper struct {
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

// NewTweegoWrapper crea un nuovo wrapper per Tweego
func NewTweegoWrapper(tweegoPath string, workDir string) (*TweegoWrapper, error) {
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

	return &TweegoWrapper{
		tweegoPath: tweegoPath,
		workDir:    workDir,
	}, nil
}

// Compile compila un file .twee in HTML usando Tweego
func (tw *TweegoWrapper) Compile(inputFile string, options *CompileOptions) (*CompileResult, error) {
	result := &CompileResult{
		Success: false,
	}

	// Validazione pre-compilazione
	if err := tw.validateBeforeCompile(inputFile, options); err != nil {
		result.ErrorMessage = err.Error()
		return result, err
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
	if !filepath.IsAbs(outputPath) && tw.workDir != "" {
		outputPath = filepath.Join(tw.workDir, outputPath)
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
	cmd := exec.Command(tw.tweegoPath, args...)
	
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
func (tw *TweegoWrapper) GetVersion() (string, error) {
	cmd := exec.Command(tw.tweegoPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("impossibile ottenere versione tweego: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ListFormats elenca i formati disponibili in Tweego
func (tw *TweegoWrapper) ListFormats() ([]string, error) {
	cmd := exec.Command(tw.tweegoPath, "--list-formats")
	
	// Cattura sia stdout che stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Esegui il comando (ignoriamo l'errore perché tweego usa stderr)
	_ = cmd.Run()
	
	// Tweego scrive i formati su stderr
	output := stderr.String()
	if output == "" {
		output = stdout.String()
	}
	
	// Se non c'è nessun output, allora c'è un vero errore
	if output == "" {
		return nil, fmt.Errorf("nessun output da tweego --list-formats")
	}

	formats := []string{}
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Salta linee vuote, header e separatori
		if line == "" || 
		   strings.HasPrefix(line, "Available formats:") ||
		   strings.HasPrefix(line, "ID") ||
		   strings.HasPrefix(line, "---") ||
		   strings.HasPrefix(line, "Story formats:") {
			continue
		}
		
		// Estrai solo l'ID del formato (prima colonna)
		fields := strings.Fields(line)
		if len(fields) > 0 {
			formatID := fields[0]
			formats = append(formats, formatID)
		}
	}

	return formats, nil
}

// validateBeforeCompile valida file e opzioni prima della compilazione
func (tw *TweegoWrapper) validateBeforeCompile(inputFile string, options *CompileOptions) error {
	// 1. Verifica che il file esista
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("file input non trovato: %s", inputFile)
	}

	// 2. Verifica integrità del file (checksum base)
	fileInfo, err := os.Stat(inputFile)
	if err != nil {
		return fmt.Errorf("impossibile leggere info file: %w", err)
	}
	
	if fileInfo.Size() == 0 {
		return fmt.Errorf("file input vuoto: %s", inputFile)
	}

	// 3. Verifica che il formato esista (se specificato)
	if options != nil && options.Format != "" {
		formats, err := tw.ListFormats()
		if err != nil {
			return fmt.Errorf("impossibile verificare formato: %w", err)
		}

		formatExists := false
		for _, f := range formats {
			if f == options.Format {
				formatExists = true
				break
			}
		}

		if !formatExists {
			return fmt.Errorf("formato '%s' non riconosciuto. Formati disponibili: %v", options.Format, formats)
		}
	}

	// 4. Verifica che il file sia .twee
	if !strings.HasSuffix(strings.ToLower(inputFile), ".twee") {
		return fmt.Errorf("il file deve avere estensione .twee")
	}

	return nil
}