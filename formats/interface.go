package formats

// LiteralInfo contiene il raw e il valore parsato di un literal
type LiteralInfo struct {
	Passage string      `json:"passage,omitempty"` // Passaggio dove è stato trovato
	Raw     string      `json:"raw"`               // Stringa originale es: (a: "spada", "scudo")
	Parsed  interface{} `json:"parsed"`            // Valore parsato es: []interface{}{"spada", "scudo"}
}

// LiteralsResult contiene tutti i literals estratti da un contenuto
type LiteralsResult struct {
	Arrays   []LiteralInfo `json:"arrays"`
	Datamaps []LiteralInfo `json:"datamaps"`
	Datasets []LiteralInfo `json:"datasets"`
}

// StoryFormat definisce l'interface per i diversi formati
type StoryFormat interface {
	// GetFormatName restituisce il nome del formato
	GetFormatName() string

	// ParseLinks estrae i collegamenti dal contenuto
	ParseLinks(content string) []string

	// ParseVariables estrae le variabili dal contenuto
	ParseVariables(content string) map[string]interface{}

	// StripCode rimuove il codice lasciando solo il testo
	StripCode(content string) string

	// === LITERALS ===

	// ParseArrayLiteral parsa un singolo array literal
	// Es: (a: "spada", "scudo") -> []interface{}{"spada", "scudo"}
	ParseArrayLiteral(content string) []interface{}

	// ParseDatamapLiteral parsa un singolo datamap literal
	// Es: (dm: "nome", "Eroe") -> map[string]interface{}{"nome": "Eroe"}
	ParseDatamapLiteral(content string) map[string]interface{}

	// ParseDatasetLiteral parsa un singolo dataset literal
	// Es: (ds: "a", "b") -> []interface{}{"a", "b"}
	ParseDatasetLiteral(content string) []interface{}

	// FindAllArrayLiterals trova tutti gli array literals nel contenuto
	FindAllArrayLiterals(content string) [][]interface{}

	// FindAllDatamapLiterals trova tutti i datamap literals nel contenuto
	FindAllDatamapLiterals(content string) []map[string]interface{}

	// FindAllDatasetLiterals trova tutti i dataset literals nel contenuto
	FindAllDatasetLiterals(content string) [][]interface{}

	// ExtractAllLiterals estrae tutti i literals con raw + parsed
	// Questo è il metodo principale che il runner dovrebbe usare
	ExtractAllLiterals(content string) *LiteralsResult
}