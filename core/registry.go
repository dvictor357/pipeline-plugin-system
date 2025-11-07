package core

import (
	"fmt"
	"sync"
)

// Registry manages plugin registration and retrieval with thread-safe storage.
// It allows plugins to be registered by name and retrieved for pipeline construction.
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewRegistry creates a new Registry with an empty plugin map.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry with the given name.
// Returns an error if a plugin with the same name is already registered.
func (r *Registry) Register(name string, plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %q is already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// Get retrieves a plugin by name from the registry.
// Returns an error if the plugin is not found.
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %q not found in registry", name)
	}

	return plugin, nil
}

// BuildPipeline constructs a pipeline from a list of plugin names.
// Returns an error if any plugin name is not found in the registry.
func (r *Registry) BuildPipeline(names []string, strategy ErrorStrategy) (*Pipeline, error) {
	pipeline := NewPipeline(strategy)

	for _, name := range names {
		plugin, err := r.Get(name)
		if err != nil {
			return nil, fmt.Errorf("failed to build pipeline: %w", err)
		}
		pipeline.Use(plugin)
	}

	return pipeline, nil
}
