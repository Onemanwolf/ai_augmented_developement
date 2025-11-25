package events

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// Registry maintains a mapping of event type names to their Go types.
type Registry struct {
	mu    sync.RWMutex
	types map[string]reflect.Type
}

// NewRegistry creates a new event type registry.
func NewRegistry() *Registry {
	return &Registry{
		types: make(map[string]reflect.Type),
	}
}

// globalRegistry is the default registry instance.
var globalRegistry = NewRegistry()

// Register adds an event type to the global registry.
func Register(eventType string, event interface{}) {
	globalRegistry.Register(eventType, event)
}

// Register adds an event type to the registry.
func (r *Registry) Register(eventType string, event interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := reflect.TypeOf(event)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	r.types[eventType] = t
}

// Deserialize unmarshals JSON data into the appropriate event type.
func Deserialize(eventType string, data []byte) (interface{}, error) {
	return globalRegistry.Deserialize(eventType, data)
}

// Deserialize unmarshals JSON data into the appropriate event type.
func (r *Registry) Deserialize(eventType string, data []byte) (interface{}, error) {
	r.mu.RLock()
	t, ok := r.types[eventType]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}

	event := reflect.New(t).Interface()
	if err := json.Unmarshal(data, event); err != nil {
		return nil, fmt.Errorf("failed to deserialize event %s: %w", eventType, err)
	}

	return event, nil
}

// IsRegistered checks if an event type is registered.
func IsRegistered(eventType string) bool {
	return globalRegistry.IsRegistered(eventType)
}

// IsRegistered checks if an event type is registered.
func (r *Registry) IsRegistered(eventType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.types[eventType]
	return ok
}

// RegisteredTypes returns all registered event type names.
func RegisteredTypes() []string {
	return globalRegistry.RegisteredTypes()
}

// RegisteredTypes returns all registered event type names.
func (r *Registry) RegisteredTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.types))
	for t := range r.types {
		types = append(types, t)
	}
	return types
}
