package feature

import "testing"

// fakeFeature is a minimal Feature for registry ordering tests.
type fakeFeature struct {
	id       string
	priority int
}

func (f fakeFeature) ID() string                          { return f.id }
func (f fakeFeature) Priority() int                       { return f.priority }
func (f fakeFeature) Validate(FeatureContext) []Condition { return nil }
func (f fakeFeature) Preflight(FeatureContext) []CheckResult {
	return nil
}
func (f fakeFeature) Build(FeatureContext) DesiredState    { return DesiredState{} }
func (f fakeFeature) Collect(FeatureContext) []ArtifactRef { return nil }
func (f fakeFeature) Normalize(FeatureContext) []Finding   { return nil }

// withCleanRegistry swaps in an isolated registry for the duration of a test so
// the global state is not polluted across tests.
func withCleanRegistry(t *testing.T, fn func()) {
	t.Helper()
	registry.Lock()
	saved := registry.byID
	registry.byID = map[string]Feature{}
	registry.Unlock()
	defer func() {
		registry.Lock()
		registry.byID = saved
		registry.Unlock()
	}()
	fn()
}

func TestRegisterAndGet(t *testing.T) {
	withCleanRegistry(t, func() {
		Register(fakeFeature{id: "source_security", priority: 50})
		got, ok := Get("source_security")
		if !ok {
			t.Fatalf("expected source_security to be registered")
		}
		if got.Priority() != 50 {
			t.Fatalf("priority = %d, want 50", got.Priority())
		}
		if _, ok := Get("missing"); ok {
			t.Fatalf("did not expect missing feature to resolve")
		}
	})
}

func TestRegisterDuplicatePanics(t *testing.T) {
	withCleanRegistry(t, func() {
		Register(fakeFeature{id: "sbom", priority: 100})
		defer func() {
			if recover() == nil {
				t.Fatalf("expected panic on duplicate registration")
			}
		}()
		Register(fakeFeature{id: "sbom", priority: 100})
	})
}

func TestAllOrdersByPriorityThenID(t *testing.T) {
	withCleanRegistry(t, func() {
		// Register out of order; expect priority asc, then ID lexicographic.
		Register(fakeFeature{id: "report_export", priority: 300})
		Register(fakeFeature{id: "sbom", priority: 100})
		Register(fakeFeature{id: "image_integrity", priority: 100})
		Register(fakeFeature{id: "target_preflight", priority: 10})

		want := []string{"target_preflight", "image_integrity", "sbom", "report_export"}
		got := All()
		if len(got) != len(want) {
			t.Fatalf("len = %d, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i].ID() != want[i] {
				t.Fatalf("order[%d] = %q, want %q", i, got[i].ID(), want[i])
			}
		}
	})
}

func TestResolveSeparatesUnknown(t *testing.T) {
	withCleanRegistry(t, func() {
		Register(fakeFeature{id: "secret_scan", priority: 50})
		Register(fakeFeature{id: "source_security", priority: 50})

		resolved, unknown := Resolve([]string{"source_security", "nope", "secret_scan", "source_security"})
		if len(resolved) != 2 {
			t.Fatalf("resolved = %d, want 2", len(resolved))
		}
		// priority equal (50) → lexicographic: secret_scan before source_security
		if resolved[0].ID() != "secret_scan" || resolved[1].ID() != "source_security" {
			t.Fatalf("resolved order = [%s, %s]", resolved[0].ID(), resolved[1].ID())
		}
		if len(unknown) != 1 || unknown[0] != "nope" {
			t.Fatalf("unknown = %v, want [nope]", unknown)
		}
	})
}
