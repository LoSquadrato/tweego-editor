package formats

import (
	"strings"
	"sync"
)

// formatRegistry mantiene i parser registrati
var (
	registry     = make(map[string]func() StoryFormat)
	registryLock sync.RWMutex
)

// RegisterFormat registra un nuovo formato
// Chiamato dai package dei singoli formati nel loro init()
func RegisterFormat(name string, factory func() StoryFormat) {
	registryLock.Lock()
	defer registryLock.Unlock()
	registry[strings.ToLower(name)] = factory
}

// GetRegisteredFormat restituisce il parser per un formato registrato
func GetRegisteredFormat(name string) StoryFormat {
	registryLock.RLock()
	defer registryLock.RUnlock()

	factory, exists := registry[strings.ToLower(name)]
	if !exists {
		return nil
	}
	return factory()
}

// GetAvailableFormats restituisce i nomi dei formati registrati
func GetAvailableFormats() []string {
	registryLock.RLock()
	defer registryLock.RUnlock()

	formats := make([]string, 0, len(registry))
	for name := range registry {
		formats = append(formats, name)
	}
	return formats
}

// IsFormatRegistered verifica se un formato è registrato
func IsFormatRegistered(name string) bool {
	registryLock.RLock()
	defer registryLock.RUnlock()

	_, exists := registry[strings.ToLower(name)]
	return exists
}