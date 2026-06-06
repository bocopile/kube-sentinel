package kubesentinel_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoTestAllPackagesPasses(t *testing.T) {
	t.Parallel()

	if os.Getenv("KUBE_SENTINEL_GO_TEST_ALL_CHILD") == "1" {
		return
	}

	root := projectRoot(t)
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(),
		"KUBE_SENTINEL_GO_TEST_ALL_CHILD=1",
		"GOCACHE="+filepath.Join(t.TempDir(), "go-build"),
		"GOMODCACHE="+filepath.Join(t.TempDir(), "go-mod"),
	)

	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)
	if err != nil {
		t.Fatalf("go test ./... failed: %v\n%s", err, output)
	}

	if strings.Contains(output, "[no test files]") || strings.Contains(output, "[no tests to run]") {
		t.Fatalf("go test ./... must pass concrete package test suites, but some packages have no tests:\n%s", output)
	}
}

func TestGoTestAllPackagesExitsWithCodeZero(t *testing.T) {
	t.Parallel()

	if os.Getenv("KUBE_SENTINEL_GO_TEST_ALL_CHILD") == "1" {
		return
	}

	root := projectRoot(t)
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(),
		"KUBE_SENTINEL_GO_TEST_ALL_CHILD=1",
		"GOCACHE="+filepath.Join(t.TempDir(), "go-build"),
		"GOMODCACHE="+filepath.Join(t.TempDir(), "go-mod"),
	)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		exitCode := -1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		t.Fatalf("go test ./... must exit with code 0, got %d: %v\n%s", exitCode, err, outputBytes)
	}
}

func projectRoot(t *testing.T) string {
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
