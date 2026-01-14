package formats

// LiteralInfo contiene il raw e il valore parsato di un literal
type LiteralInfo struct {
	Passage string      `json:"passage,omitempty"`
	Raw     string      `json:"raw"`
	Parsed  interface{} `json:"parsed"`
}

// LiteralsResult contiene tutti i literals estratti da un contenuto
type LiteralsResult struct {
	Arrays   []LiteralInfo `json:"arrays"`
	Datamaps []LiteralInfo `json:"datamaps"`
	Datasets []LiteralInfo `json:"datasets"`
}

// ============================================
// EVALUATOR INTERFACE (NUOVO!)
// ============================================

// Evaluator definisce l'interfaccia per gli evaluator dei vari formati
// L'evaluator riceve lo stato e il contesto dal PathSimulator
type Evaluator interface {
	// Gestione stato variabili
	GetState() map[string]interface{}
	SetState(state map[string]interface{})

	// Valutazione espressioni
	EvaluateExpression(expression string) (interface{}, error)
	EvaluateCondition(condition string) (bool, error)

	// Contesto per valutazione (passato dal PathSimulator)
	SetVisitedPassages(visited map[string]int)
	SetHistory(history []string)
	SetCurrentPassage(passageName string)
}

// ============================================
// STORY FORMAT INTERFACE (AGGIORNATO!)
// ============================================

// StoryFormat definisce l'interface per i diversi formati
type StoryFormat interface {
	// GetFormatName restituisce il nome del formato
	GetFormatName() string

	// CreateEvaluator crea un nuovo evaluator per questo formato
	// NUOVO METODO per supportare il PathSimulator format-agnostic
	CreateEvaluator(initialState map[string]interface{}) Evaluator

	// ProcessPassageContent processa il contenuto di un passaggio
	// modificando lo stato dell'evaluator passato.
	// NUOVO METODO per permettere al PathSimulator di processare
	// i passaggi senza duplicare la logica di parsing
	ProcessPassageContent(content string, eval Evaluator) error


	// ParseLinks estrae i collegamenti dal contenuto
	ParseLinks(content string) []string

	// ParseVariables estrae le variabili dal contenuto
	ParseVariables(content string) map[string]interface{}

	// StripCode rimuove il codice lasciando solo il testo
	StripCode(content string) string

	// === LITERALS ===

	// ParseArrayLiteral parsa un singolo array literal
	ParseArrayLiteral(content string) []interface{}

	// ParseDatamapLiteral parsa un singolo datamap literal
	ParseDatamapLiteral(content string) map[string]interface{}

	// ParseDatasetLiteral parsa un singolo dataset literal
	ParseDatasetLiteral(content string) []interface{}

	// FindAllArrayLiterals trova tutti gli array literals nel contenuto
	FindAllArrayLiterals(content string) [][]interface{}

	// FindAllDatamapLiterals trova tutti i datamap literals nel contenuto
	FindAllDatamapLiterals(content string) []map[string]interface{}

	// FindAllDatasetLiterals trova tutti i dataset literals nel contenuto
	FindAllDatasetLiterals(content string) [][]interface{}

	// ExtractAllLiterals estrae tutti i literals con raw + parsed
	ExtractAllLiterals(content string) *LiteralsResult
}