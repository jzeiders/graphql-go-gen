package presets

import (
	"fmt"
	"sync"
)

// Registry manages registered presets
type Registry struct {
	mu      sync.RWMutex
	presets map[string]Preset
}

// NewRegistry creates a new preset registry
func NewRegistry() *Registry {
	return &Registry{
		presets: make(map[string]Preset),
	}
}

// globalRegistry is the default preset registry
var globalRegistry = NewRegistry()

// Register registers a preset with the given name
func (r *Registry) Register(name string, preset Preset) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.presets[name]; exists {
		return fmt.Errorf("preset %q already registered", name)
	}

	r.presets[name] = preset
	return nil
}

// Get retrieves a preset by name
func (r *Registry) Get(name string) (Preset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	preset, exists := r.presets[name]
	if !exists {
		return nil, fmt.Errorf("preset %q not found", name)
	}

	return preset, nil
}

// List returns all registered preset names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.presets))
	for name := range r.presets {
		names = append(names, name)
	}
	return names
}

// Register registers a preset globally
func Register(name string, preset Preset) error {
	return globalRegistry.Register(name, preset)
}

// Get retrieves a preset from the global registry
func Get(name string) (Preset, error) {
	return globalRegistry.Get(name)
}

// List returns all globally registered preset names
func List() []string {
	return globalRegistry.List()
}