package plugin

import "sync"

var (
	// globalRegistry is the global plugin registry
	globalRegistry *DefaultRegistry
	registryOnce   sync.Once
)

// GetGlobalRegistry returns the global plugin registry
func GetGlobalRegistry() *DefaultRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewRegistry()
	})
	return globalRegistry
}

// Register registers a plugin in the global registry
func Register(name string, plugin Plugin) error {
	return GetGlobalRegistry().Register(plugin)
}

// Get retrieves a plugin from the global registry
func Get(name string) (Plugin, bool) {
	return GetGlobalRegistry().Get(name)
}

// Has checks if a plugin exists in the global registry
func Has(name string) bool {
	return GetGlobalRegistry().Has(name)
}

// List returns all registered plugin names from the global registry
func List() []string {
	return GetGlobalRegistry().List()
}