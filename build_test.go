package kubesentinel_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoBuildAllPackages(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	packages := goListPackages(t, root)
	if len(packages) == 0 {
		t.Fatalf("go list ./... returned no packages; go build ./... must cover every module package")
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(),
		"GOCACHE="+filepath.Join(t.TempDir(), "go-build"),
		"GOMODCACHE="+filepath.Join(t.TempDir(), "go-mod"),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./... failed: %v\n%s", err, output)
	}
}

func TestGoBuildAllPackagesCompletesWithoutBuildDiagnostics(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = root
	cmd.Env = goBuildDefaultCacheEnv()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./... exited non-zero: %v\n%s", err, output)
	}
	if len(output) > 0 {
		t.Fatalf("go build ./... must complete without build diagnostics or compilation errors, got:\n%s", output)
	}
}

func goBuildDefaultCacheEnv() []string {
	var env []string
	for _, variable := range os.Environ() {
		if strings.HasPrefix(variable, "GOCACHE=") || strings.HasPrefix(variable, "GOMODCACHE=") {
			continue
		}
		env = append(env, variable)
	}
	return env
}

func goListPackages(t *testing.T, root string) []string {
	t.Helper()

	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(),
		"GOCACHE="+filepath.Join(t.TempDir(), "go-list-build"),
		"GOMODCACHE="+filepath.Join(t.TempDir(), "go-list-mod"),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list ./... failed before build verification: %v\n%s", err, output)
	}

	return strings.Fields(string(output))
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find module root from %s", dir)
		}
		dir = parent
	}
}
