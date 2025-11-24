package api

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"tweego-editor/compiler"
	"tweego-editor/formats/harlowe"
	"tweego-editor/parser"
	"tweego-editor/simulator"
	"tweego-editor/watcher"
)

// Server rappresenta il server API
type Server struct {
	router       *gin.Engine
	compiler     *compiler.TweegoWrapper
	watcher      *watcher.FileWatcher
	watcherMutex sync.Mutex
	wsClients    map[*websocket.Conn]bool
	wsUpgrader   websocket.Upgrader
	port         int
}

// ServerConfig configurazione del server
type ServerConfig struct {
	Port         int
	Compiler     *compiler.TweegoWrapper
	EnableCORS   bool
	Debug        bool
}

// NewServer crea un nuovo server API
func NewServer(config ServerConfig) *Server {
	// Imposta modalità Gin
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS se abilitato
	if config.EnableCORS {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	}

	server := &Server{
		router:    router,
		compiler:  config.Compiler,
		wsClients: make(map[*websocket.Conn]bool),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		port: config.Port,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configura tutti gli endpoint
func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		// Health check
		api.GET("/health", s.healthCheck)

		// Story endpoints
		api.POST("/story/parse", s.parseStory)
		api.POST("/story/compile", s.compileStory)
		api.POST("/story/validate", s.validateStory)

		// Passage endpoints
		api.GET("/story/:file/passages", s.getPassages)
		api.GET("/story/:file/passage/:title", s.getPassage)

		// Path Simulator endpoints
		api.POST("/simulator/validate", s.validatePath)
		api.POST("/simulator/simulate", s.simulatePath)
		api.POST("/simulator/suggest", s.suggestPaths)

		// Watcher endpoints
		api.POST("/watch/start", s.startWatcher)
		api.POST("/watch/stop", s.stopWatcher)
		api.GET("/watch/status", s.getWatcherStatus)

		// Utils endpoints
		api.GET("/formats", s.getFormats)
		api.GET("/version", s.getVersion)
	}

	// WebSocket endpoint
	s.router.GET("/ws", s.handleWebSocket)
}

// Start avvia il server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🚀 Server avviato su http://localhost%s", addr)
	log.Printf("📚 API disponibile su http://localhost%s/api", addr)
	log.Printf("🔌 WebSocket su ws://localhost%s/ws", addr)
	return s.router.Run(addr)
}

// ============================================
// Handlers
// ============================================

// healthCheck verifica lo stato del server
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"version": "0.1.0",
	})
}

// ValidateStoryRequest richiesta di validazione
type ValidateStoryRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

// validateStory valida un file .twee
func (s *Server) validateStory(c *gin.Context) {
	var req ValidateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valida il file
	tweeParser := parser.NewTweeParser(req.FilePath)
	validation := tweeParser.Validate()

	c.JSON(http.StatusOK, validation)
}

// ParseStoryRequest richiesta di parsing
type ParseStoryRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

// parseStory parsa un file .twee
func (s *Server) parseStory(c *gin.Context) {
	var req ParseStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse il file
	tweeParser := parser.NewTweeParser(req.FilePath)
	story, err := tweeParser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Processa con Harlowe format
	harloweFormat := harlowe.NewHarloweFormat()
	
	// Arricchisci i passaggi con info extra
	enrichedPassages := make(map[string]interface{})
	for title, passage := range story.Passages {
		enrichedPassages[title] = gin.H{
			"title":     passage.Title,
			"tags":      passage.Tags,
			"content":   passage.Content,
			"links":     harloweFormat.ParseLinks(passage.Content),
			"variables": harloweFormat.ParseVariables(passage.Content),
			"preview":   harloweFormat.StripCode(passage.Content),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"story": gin.H{
			"title":    story.Title,
			"format":   story.Format,
			"passages": enrichedPassages,
			"count":    len(story.Passages),
		},
	})
}

// CompileStoryRequest richiesta di compilazione
type CompileStoryRequest struct {
	FilePath string `json:"file_path" binding:"required"`
	Format   string `json:"format"`
	Output   string `json:"output"`
}

// compileStory compila un file .twee
func (s *Server) compileStory(c *gin.Context) {
	var req CompileStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default values
	if req.Format == "" {
		req.Format = "harlowe-3"
	}
	if req.Output == "" {
		req.Output = "output.html"
	}

	// Compila
	result, err := s.compiler.Compile(req.FilePath, &compiler.CompileOptions{
		Format: req.Format,
		Output: req.Output,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     result.Success,
		"output_file": result.OutputFile,
		"warnings":    result.Warnings,
	})
}

// getPassages ottiene tutti i passaggi
func (s *Server) getPassages(c *gin.Context) {
	filePath := c.Param("file")
	
	tweeParser := parser.NewTweeParser(filePath)
	story, err := tweeParser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"passages": story.Passages,
	})
}

// getPassage ottiene un singolo passaggio
func (s *Server) getPassage(c *gin.Context) {
	filePath := c.Param("file")
	passageTitle := c.Param("title")
	
	tweeParser := parser.NewTweeParser(filePath)
	story, err := tweeParser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	passage, exists := story.Passages[passageTitle]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Passaggio non trovato"})
		return
	}

	harloweFormat := harlowe.NewHarloweFormat()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"passage": gin.H{
			"title":     passage.Title,
			"tags":      passage.Tags,
			"content":   passage.Content,
			"links":     harloweFormat.ParseLinks(passage.Content),
			"variables": harloweFormat.ParseVariables(passage.Content),
			"preview":   harloweFormat.StripCode(passage.Content),
		},
	})
}

// StartWatcherRequest richiesta avvio watcher
type StartWatcherRequest struct {
	Paths       []string `json:"paths" binding:"required"`
	Format      string   `json:"format"`
	Output      string   `json:"output"`
	AutoCompile bool     `json:"auto_compile"`
}

// startWatcher avvia il file watcher
func (s *Server) startWatcher(c *gin.Context) {
	s.watcherMutex.Lock()
	defer s.watcherMutex.Unlock()

	if s.watcher != nil && s.watcher.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Watcher già in esecuzione"})
		return
	}

	var req StartWatcherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default values
	if req.Format == "" {
		req.Format = "harlowe-3"
	}
	if req.Output == "" {
		req.Output = "output.html"
	}

	// Crea watcher
	config := watcher.WatcherConfig{
		Paths:    req.Paths,
		Compiler: s.compiler,
		CompileOpts: &compiler.CompileOptions{
			Format: req.Format,
			Output: req.Output,
		},
		AutoCompile: req.AutoCompile,
	}

	fw, err := watcher.NewFileWatcher(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := fw.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.watcher = fw

	// Invia eventi ai client WebSocket
	go s.broadcastWatcherEvents()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Watcher avviato",
		"paths":   req.Paths,
	})
}

// stopWatcher ferma il file watcher
func (s *Server) stopWatcher(c *gin.Context) {
	s.watcherMutex.Lock()
	defer s.watcherMutex.Unlock()

	if s.watcher == nil || !s.watcher.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Watcher non in esecuzione"})
		return
	}

	if err := s.watcher.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.watcher = nil

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Watcher fermato",
	})
}

// getWatcherStatus ottiene lo stato del watcher
func (s *Server) getWatcherStatus(c *gin.Context) {
	s.watcherMutex.Lock()
	defer s.watcherMutex.Unlock()

	isRunning := s.watcher != nil && s.watcher.IsRunning()

	c.JSON(http.StatusOK, gin.H{
		"running": isRunning,
	})
}

// getFormats ottiene i formati disponibili
func (s *Server) getFormats(c *gin.Context) {
	formats, err := s.compiler.ListFormats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"formats": formats,
	})
}

// getVersion ottiene la versione di Tweego
func (s *Server) getVersion(c *gin.Context) {
	version, err := s.compiler.GetVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"version": version,
	})
}

// ============================================
// Path Simulator Handlers
// ============================================

// ValidatePathRequest richiesta di validazione path
type ValidatePathRequest struct {
	FilePath string   `json:"file_path" binding:"required"`
	Path     []string `json:"path" binding:"required"`
}

// validatePath valida un percorso
func (s *Server) validatePath(c *gin.Context) {
	var req ValidatePathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse la storia
	tweeParser := parser.NewTweeParser(req.FilePath)
	story, err := tweeParser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Crea simulator e valida
	simulator := simulator.NewPathSimulator(story)
	errors := simulator.ValidatePath(req.Path)

	c.JSON(http.StatusOK, gin.H{
		"valid":  len(errors) == 0,
		"path":   req.Path,
		"errors": errors,
	})
}

// SimulatePathRequest richiesta di simulazione path
type SimulatePathRequest struct {
	FilePath string   `json:"file_path" binding:"required"`
	Path     []string `json:"path" binding:"required"`
}

// simulatePath simula l'esecuzione di un percorso
func (s *Server) simulatePath(c *gin.Context) {
	var req SimulatePathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse la storia
	tweeParser := parser.NewTweeParser(req.FilePath)
	story, err := tweeParser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Crea simulator e simula
	sim := simulator.NewPathSimulator(story)
	result := sim.SimulatePath(req.Path)

	c.JSON(http.StatusOK, result)
}

// SuggestPathsRequest richiesta di suggerimento percorsi
type SuggestPathsRequest struct {
	FilePath      string `json:"file_path" binding:"required"`
	StartPassage  string `json:"start_passage" binding:"required"`
	MaxDepth      int    `json:"max_depth"`
}

// suggestPaths suggerisce percorsi validi
func (s *Server) suggestPaths(c *gin.Context) {
	var req SuggestPathsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default max depth
	if req.MaxDepth == 0 || req.MaxDepth > 10 {
		req.MaxDepth = 5
	}

	// Parse la storia
	tweeParser := parser.NewTweeParser(req.FilePath)
	story, err := tweeParser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Crea simulator e suggerisci
	sim := simulator.NewPathSimulator(story)
	paths := sim.GetSuggestedPaths(req.StartPassage, req.MaxDepth)

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"start_passage": req.StartPassage,
		"max_depth":     req.MaxDepth,
		"paths":         paths,
		"count":         len(paths),
	})
}

// ============================================
// WebSocket
// ============================================

// handleWebSocket gestisce connessioni WebSocket
func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Errore upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	s.wsClients[conn] = true
	log.Printf("🔌 Client WebSocket connesso (totale: %d)", len(s.wsClients))

	// Mantieni la connessione aperta
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(s.wsClients, conn)
			log.Printf("🔌 Client WebSocket disconnesso (totale: %d)", len(s.wsClients))
			break
		}
	}
}

// broadcastWatcherEvents invia eventi del watcher ai client WebSocket
func (s *Server) broadcastWatcherEvents() {
	if s.watcher == nil {
		return
	}

	for event := range s.watcher.Events() {
		message := gin.H{
			"type":      event.Type,
			"path":      filepath.Base(event.Path),
			"full_path": event.Path,
			"timestamp": event.Timestamp,
		}

		// Broadcast a tutti i client connessi
		for client := range s.wsClients {
			if err := client.WriteJSON(message); err != nil {
				log.Printf("Errore invio WebSocket: %v", err)
				client.Close()
				delete(s.wsClients, client)
			}
		}
	}
}