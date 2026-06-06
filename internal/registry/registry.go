package registry

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Feature represents a registered feature with an ID and execution priority.
type Feature struct {
	ID       string
	Priority int
}

// ValidationError is returned when a feature name is not found in the registry.
type ValidationError struct {
	FeatureName string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("unrecognised feature name: %q", e.FeatureName)
}

func (e *ValidationError) Is(target error) bool {
	var t *ValidationError
	if errors.As(target, &t) {
		return e.FeatureName == t.FeatureName
	}
	return false
}

var (
	mu            sync.RWMutex
	globalFeatures = map[string]Feature{}
)

// Register adds a feature to the global registry with the given priority.
func Register(id string, priority int) {
	mu.Lock()
	defer mu.Unlock()
	globalFeatures[id] = Feature{ID: id, Priority: priority}
}

// ValidateFeatureName returns a *ValidationError if the feature name is not registered.
func ValidateFeatureName(name string) error {
	mu.RLock()
	defer mu.RUnlock()
	if _, ok := globalFeatures[name]; !ok {
		return &ValidationError{FeatureName: name}
	}
	return nil
}

// InitialiseKnownFeatureNames registers the given names as known features so ValidateFeatureName accepts them.
func InitialiseKnownFeatureNames(names []string) {
	mu.Lock()
	defer mu.Unlock()
	for _, name := range names {
		if _, exists := globalFeatures[name]; !exists {
			globalFeatures[name] = Feature{ID: name}
		}
	}
}

// List returns all registered features ordered by ascending priority then lexicographic ID.
func List() []Feature {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]Feature, 0, len(globalFeatures))
	for _, f := range globalFeatures {
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Priority != result[j].Priority {
			return result[i].Priority < result[j].Priority
		}
		return result[i].ID < result[j].ID
	})
	return result
}
