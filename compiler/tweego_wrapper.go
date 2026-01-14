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

	// Story format (SEMPRE richiesto ora)
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

	// Argomenti aggiuntivi
	args = append(args, options.AdditionalArgs...)

	// Input file
	args = append(args, inputFile)

	// Esegui tweego
	cmd := exec.Command(tw.tweegoPath, args...)

	// Cattura output e errori
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Tweego usa stderr per informazioni, non solo errori
	output := stderr.String()
	if output == "" {
		output = stdout.String()
	}

	result.Output = output

	// Parse warnings
	if strings.Contains(output, "warning:") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "warning") {
				result.Warnings = append(result.Warnings, line)
			}
		}
	}

	// Se c'è un errore E non è solo warnings
	if err != nil {
		result.ErrorMessage = output
		return result, fmt.Errorf("compilazione fallita: %w", err)
	}

	result.Success = true
	result.OutputFile = outputPath

	return result, nil
}

// GetVersion restituisce la versione di Tweego
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

	// 3. Verifica che il file sia .twee
	if !strings.HasSuffix(strings.ToLower(inputFile), ".twee") {
		return fmt.Errorf("il file deve avere estensione .twee")
	}

	// 4. VALIDAZIONE FORMATO - SEMPRE OBBLIGATORIA
	// Se non è specificato un formato, è un errore
	if options == nil || options.Format == "" {
		return fmt.Errorf("formato non specificato. Il formato è obbligatorio per la compilazione")
	}

	// 5. Verifica che il formato specificato sia supportato da Tweego
	formats, err := tw.ListFormats()
	if err != nil {
		return fmt.Errorf("impossibile verificare formati disponibili: %w", err)
	}

	formatExists := false
	normalizedFormat := strings.ToLower(options.Format)
	
	for _, f := range formats {
		// Confronta case-insensitive e con prefissi
		// Es: "harlowe" matcha "harlowe-3", "harlowe-3.2.3"
		if strings.HasPrefix(strings.ToLower(f), normalizedFormat) || 
		   strings.EqualFold(f, normalizedFormat) {
			formatExists = true
			break
		}
	}

	if !formatExists {
		return fmt.Errorf("formato '%s' non supportato da Tweego. Formati disponibili: %v", 
			options.Format, formats)
	}

	return nil
}