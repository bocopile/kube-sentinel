package feature

import (
	"fmt"
	"sort"
	"sync"
)

// registry is the global, priority-ordered feature registry. Features call
// Register() from their init() so the reconciler needs no per-feature wiring.
var registry = struct {
	sync.RWMutex
	byID map[string]Feature
}{byID: map[string]Feature{}}

// Register adds a feature to the global registry. Panics on duplicate ID so
// build-time wiring mistakes fail loudly. Call from a feature package init().
func Register(f Feature) {
	registry.Lock()
	defer registry.Unlock()
	if _, exists := registry.byID[f.ID()]; exists {
		panic(fmt.Sprintf("feature %q already registered", f.ID()))
	}
	registry.byID[f.ID()] = f
}

// Get returns a feature by ID.
func Get(id string) (Feature, bool) {
	registry.RLock()
	defer registry.RUnlock()
	f, ok := registry.byID[id]
	return f, ok
}

// All returns every registered feature sorted by Priority ascending, then ID
// lexicographically (matches the resolver ordering in docs/ARCHITECTURE.md).
func All() []Feature {
	registry.RLock()
	defer registry.RUnlock()
	out := make([]Feature, 0, len(registry.byID))
	for _, f := range registry.byID {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Priority() != out[j].Priority() {
			return out[i].Priority() < out[j].Priority()
		}
		return out[i].ID() < out[j].ID()
	})
	return out
}

// Resolve returns the enabled feature set for the given IDs, sorted by priority.
// Unknown IDs are returned separately so the caller can mark them ConfigError.
func Resolve(enabledIDs []string) (resolved []Feature, unknown []string) {
	seen := map[string]bool{}
	for _, id := range enabledIDs {
		if seen[id] {
			continue
		}
		seen[id] = true
		if f, ok := Get(id); ok {
			resolved = append(resolved, f)
		} else {
			unknown = append(unknown, id)
		}
	}
	sort.Slice(resolved, func(i, j int) bool {
		if resolved[i].Priority() != resolved[j].Priority() {
			return resolved[i].Priority() < resolved[j].Priority()
		}
		return resolved[i].ID() < resolved[j].ID()
	})
	return resolved, unknown
}
