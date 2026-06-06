package v1alpha1_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bhshin/kube-sentinel/api/v1alpha1"
)

func TestSecurityAgentSpecExposesRequiredAPIFields(t *testing.T) {
	specType := reflect.TypeOf(v1alpha1.SecurityAgentSpec{})

	expectedFields := map[string]reflect.Type{
		"global":   reflect.TypeOf(v1alpha1.GlobalConfig{}),
		"features": reflect.TypeOf(v1alpha1.FeaturesConfig{}),
		"output":   reflect.TypeOf(v1alpha1.OutputConfig{}),
		"override": reflect.TypeOf(v1alpha1.OverrideConfig{}),
		"tests":    reflect.TypeOf(v1alpha1.TestsConfig{}),
	}

	for apiFieldName, expectedType := range expectedFields {
		t.Run(apiFieldName, func(t *testing.T) {
			field, ok := fieldByJSONName(specType, apiFieldName)
			if !ok {
				t.Fatalf("SecurityAgentSpec does not expose spec.%s", apiFieldName)
			}

			if field.Type != expectedType {
				t.Fatalf("spec.%s has type %s, want %s", apiFieldName, field.Type, expectedType)
			}
		})
	}
}

func TestSecurityAgentSpecSerializesRequiredAPIFields(t *testing.T) {
	spec := v1alpha1.SecurityAgentSpec{
		Global: v1alpha1.GlobalConfig{
			Enabled: true,
		},
		Features: v1alpha1.FeaturesConfig{
			PodSecurity: true,
		},
		Output: v1alpha1.OutputConfig{
			Format: "json",
		},
		Override: v1alpha1.OverrideConfig{
			Namespaces: []string{"default"},
		},
		Tests: v1alpha1.TestsConfig{
			Enabled: true,
		},
	}

	encoded, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal SecurityAgentSpec: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("unmarshal SecurityAgentSpec: %v", err)
	}

	for _, requiredField := range []string{"global", "features", "output", "override", "tests"} {
		if _, ok := fields[requiredField]; !ok {
			t.Fatalf("SecurityAgentSpec JSON does not expose spec.%s; got fields %v", requiredField, keys(fields))
		}
	}
}

func TestSecurityAgentSpecRequiredAPIFieldsAreNotMarkedOmitEmpty(t *testing.T) {
	specType := reflect.TypeOf(v1alpha1.SecurityAgentSpec{})

	for _, requiredField := range []string{"global", "features", "output", "override", "tests"} {
		field, ok := fieldByJSONName(specType, requiredField)
		if !ok {
			t.Fatalf("SecurityAgentSpec does not expose spec.%s", requiredField)
		}

		for _, tagOption := range strings.Split(field.Tag.Get("json"), ",")[1:] {
			if tagOption == "omitempty" {
				t.Fatalf("spec.%s must be exposed as a required API field, but json tag %q marks it omitempty", requiredField, field.Tag.Get("json"))
			}
		}
	}
}

func TestSecurityAgentCRDIsClusterScoped(t *testing.T) {
	root := moduleRoot(t)
	crdPath := filepath.Join(root, "config", "crd", "bases", "securityagents.kube-sentinel.io_securityagents.yaml")

	manifest, err := os.ReadFile(crdPath)
	if err != nil {
		t.Fatalf("SecurityAgent CRD manifest must exist at %s so the applied CRD scope is pinned: %v", crdPath, err)
	}

	crd := string(manifest)
	if !strings.Contains(crd, "\n  scope: Cluster\n") {
		t.Fatalf("SecurityAgent CRD must be cluster-scoped with spec.scope: Cluster")
	}

	if strings.Contains(crd, "\n  scope: Namespaced\n") {
		t.Fatalf("SecurityAgent CRD must not restrict SecurityAgent resources to a namespace")
	}

	if strings.Contains(crd, "\n  namespace:") {
		t.Fatalf("SecurityAgent CRD metadata must not set metadata.namespace because CRDs are cluster-scoped resources")
	}
}

func fieldByJSONName(specType reflect.Type, jsonName string) (reflect.StructField, bool) {
	for i := 0; i < specType.NumField(); i++ {
		field := specType.Field(i)
		if strings.Split(field.Tag.Get("json"), ",")[0] == jsonName {
			return field, true
		}
	}

	return reflect.StructField{}, false
}

func keys(fields map[string]json.RawMessage) []string {
	fieldNames := make([]string, 0, len(fields))
	for fieldName := range fields {
		fieldNames = append(fieldNames, fieldName)
	}

	return fieldNames
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
