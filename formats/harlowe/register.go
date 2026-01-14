package harlowe

import "tweego-editor/formats"

// init registra automaticamente il formato Harlowe
// Questo viene chiamato quando il package viene importato
func init() {
	formats.RegisterFormat("harlowe", func() formats.StoryFormat {
		return NewHarloweFormat()
	})
	
	// Registra anche varianti comuni del nome
	formats.RegisterFormat("Harlowe", func() formats.StoryFormat {
		return NewHarloweFormat()
	})
}