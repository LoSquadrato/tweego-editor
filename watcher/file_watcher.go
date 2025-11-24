package watcher

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"tweego-editor/compiler"
	"tweego-editor/parser"
)

// FileWatcher monitora cambiamenti ai file
type FileWatcher struct {
	watcher       *fsnotify.Watcher
	watchedPaths  []string
	compiler      *compiler.TweegoWrapper
	compileOpts   *compiler.CompileOptions
	debounceTime  time.Duration
	eventChan     chan WatchEvent
	stopChan      chan bool
	isRunning     bool
}

// WatchEvent rappresenta un evento del watcher
type WatchEvent struct {
	Type      string    // "created", "modified", "deleted", "renamed"
	Path      string    // Path del file
	Timestamp time.Time // Quando è successo
}

// WatcherConfig configurazione per il watcher
type WatcherConfig struct {
	Paths         []string                  // Path da monitorare
	Compiler      *compiler.TweegoWrapper   // Wrapper Tweego da usare
	CompileOpts   *compiler.CompileOptions  // Opzioni compilazione
	DebounceTime  time.Duration             // Tempo di debounce (default: 500ms)
	OnEvent       func(WatchEvent)          // Callback per eventi
	AutoCompile   bool                      // Compila automaticamente (default: true)
}

// NewFileWatcher crea un nuovo file watcher
func NewFileWatcher(config WatcherConfig) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("errore creazione watcher: %w", err)
	}

	// Default debounce time
	if config.DebounceTime == 0 {
		config.DebounceTime = 500 * time.Millisecond
	}

	fw := &FileWatcher{
		watcher:      watcher,
		watchedPaths: config.Paths,
		compiler:     config.Compiler,
		compileOpts:  config.CompileOpts,
		debounceTime: config.DebounceTime,
		eventChan:    make(chan WatchEvent, 100),
		stopChan:     make(chan bool),
		isRunning:    false,
	}

	// Aggiungi i path da monitorare
	for _, path := range config.Paths {
		if err := watcher.Add(path); err != nil {
			return nil, fmt.Errorf("errore aggiunta path %s: %w", path, err)
		}
		log.Printf("👀 Watching: %s", path)
	}

	return fw, nil
}

// Start avvia il file watcher
func (fw *FileWatcher) Start() error {
	if fw.isRunning {
		return fmt.Errorf("watcher già in esecuzione")
	}

	fw.isRunning = true
	log.Println("🚀 File watcher avviato!")

	// Map per debouncing
	debounceMap := make(map[string]*time.Timer)

	go func() {
		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}

				// Ignora file non .twee
				if !strings.HasSuffix(event.Name, ".twee") {
					continue
				}

				// Determina tipo evento
				var eventType string
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					eventType = "created"
				case event.Op&fsnotify.Write == fsnotify.Write:
					eventType = "modified"
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					eventType = "deleted"
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					eventType = "renamed"
				default:
					continue
				}

				watchEvent := WatchEvent{
					Type:      eventType,
					Path:      event.Name,
					Timestamp: time.Now(),
				}

				log.Printf("📝 File %s: %s", eventType, filepath.Base(event.Name))

				// Invia evento al canale
				fw.eventChan <- watchEvent

				// Debounce per ricompilazione
				if timer, exists := debounceMap[event.Name]; exists {
					timer.Stop()
				}

				debounceMap[event.Name] = time.AfterFunc(fw.debounceTime, func() {
					// Auto-compila se modificato o creato
					if (eventType == "modified" || eventType == "created") && fw.compiler != nil {
						fw.recompile(event.Name)
					}
					delete(debounceMap, event.Name)
				})

			case err, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("❌ Errore watcher: %v", err)

			case <-fw.stopChan:
				log.Println("🛑 File watcher fermato")
				return
			}
		}
	}()

	return nil
}

// Stop ferma il file watcher
func (fw *FileWatcher) Stop() error {
	if !fw.isRunning {
		return fmt.Errorf("watcher non in esecuzione")
	}

	fw.isRunning = false
	fw.stopChan <- true
	
	if err := fw.watcher.Close(); err != nil {
		return fmt.Errorf("errore chiusura watcher: %w", err)
	}

	close(fw.eventChan)
	return nil
}

// Events restituisce il canale degli eventi
func (fw *FileWatcher) Events() <-chan WatchEvent {
	return fw.eventChan
}

// IsRunning verifica se il watcher è attivo
func (fw *FileWatcher) IsRunning() bool {
	return fw.isRunning
}

// AddPath aggiunge un path da monitorare
func (fw *FileWatcher) AddPath(path string) error {
	if err := fw.watcher.Add(path); err != nil {
		return fmt.Errorf("errore aggiunta path: %w", err)
	}
	fw.watchedPaths = append(fw.watchedPaths, path)
	log.Printf("👀 Watching: %s", path)
	return nil
}

// RemovePath rimuove un path dal monitoraggio
func (fw *FileWatcher) RemovePath(path string) error {
	if err := fw.watcher.Remove(path); err != nil {
		return fmt.Errorf("errore rimozione path: %w", err)
	}
	
	// Rimuovi dalla lista
	for i, p := range fw.watchedPaths {
		if p == path {
			fw.watchedPaths = append(fw.watchedPaths[:i], fw.watchedPaths[i+1:]...)
			break
		}
	}
	
	log.Printf("👁️  Stopped watching: %s", path)
	return nil
}

// recompile ricompila il file quando viene modificato
func (fw *FileWatcher) recompile(filePath string) {
	if fw.compiler == nil || fw.compileOpts == nil {
		return
	}

	log.Printf("🔄 Ricompilazione: %s", filepath.Base(filePath))
	
	// Valida il file prima di compilare
	tweeParser := parser.NewTweeParser(filePath)
	validation := tweeParser.Validate()
	
	if !validation.Valid {
		log.Printf("❌ Validazione fallita per %s:", filepath.Base(filePath))
		for _, err := range validation.Errors {
			log.Printf("   - %s", err.Message)
		}
		
		// Invia evento di errore ma non bloccare il watcher
		fw.eventChan <- WatchEvent{
			Type:      "validation_error",
			Path:      filePath,
			Timestamp: time.Now(),
		}
		return
	}
	
	// Mostra warning se presenti
	if len(validation.Warnings) > 0 {
		for _, warn := range validation.Warnings {
			log.Printf("⚠️  %s", warn.Message)
		}
	}
	
	start := time.Now()
	result, err := fw.compiler.Compile(filePath, fw.compileOpts)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("❌ Compilazione fallita (%v): %v", elapsed, err)
		if result != nil && result.ErrorMessage != "" {
			log.Printf("   %s", result.ErrorMessage)
		}
		
		// Invia evento di errore
		fw.eventChan <- WatchEvent{
			Type:      "compile_error",
			Path:      filePath,
			Timestamp: time.Now(),
		}
		return
	}

	if result.Success {
		log.Printf("✅ Compilato con successo in %v", elapsed)
		if len(result.Warnings) > 0 {
			log.Printf("⚠️  %d warning(s)", len(result.Warnings))
		}
		
		// Invia evento di successo
		fw.eventChan <- WatchEvent{
			Type:      "compile_success",
			Path:      filePath,
			Timestamp: time.Now(),
		}
	}
}