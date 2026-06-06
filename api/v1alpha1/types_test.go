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

func TestSecurityAgentSpecRequiredAPIFieldsHaveKubebuilderValidationMarkers(t *testing.T) {
	root := moduleRoot(t)
	typesPath := filepath.Join(root, "api", "v1alpha1", "types.go")

	source, err := os.ReadFile(typesPath)
	if err != nil {
		t.Fatalf("read SecurityAgent API types from %s: %v", typesPath, err)
	}

	specFields := map[string]string{
		"Global":   "json:\"global\"",
		"Features": "json:\"features\"",
		"Output":   "json:\"output\"",
		"Override": "json:\"override\"",
		"Tests":    "json:\"tests\"",
	}

	lines := strings.Split(string(source), "\n")
	for fieldName, jsonTag := range specFields {
		fieldLine := -1
		for lineIndex, line := range lines {
			if strings.Contains(line, fieldName) && strings.Contains(line, jsonTag) {
				fieldLine = lineIndex
				break
			}
		}
		if fieldLine == -1 {
			t.Fatalf("SecurityAgentSpec must expose %s with %s", fieldName, jsonTag)
		}

		markerLine := fieldLine - 1
		for markerLine >= 0 && strings.TrimSpace(lines[markerLine]) == "" {
			markerLine--
		}
		if markerLine < 0 || strings.TrimSpace(lines[markerLine]) != "// +kubebuilder:validation:Required" {
			t.Fatalf("SecurityAgentSpec.%s must be marked required so manifests expose spec.%s in the generated API schema", fieldName, strings.Split(jsonTag, "\"")[1])
		}
	}
}

func TestSecurityAgentResourceSpecIsRequiredForManifestFields(t *testing.T) {
	resourceType := reflect.TypeOf(v1alpha1.SecurityAgent{})
	specField, ok := fieldByJSONName(resourceType, "spec")
	if !ok {
		t.Fatalf("SecurityAgent resource must expose spec so manifests can define required top-level spec fields")
	}

	for _, tagOption := range strings.Split(specField.Tag.Get("json"), ",")[1:] {
		if tagOption == "omitempty" {
			t.Fatalf("SecurityAgent spec must not be optional because manifests must expose spec.global, spec.features, spec.output, spec.override, and spec.tests; got json tag %q", specField.Tag.Get("json"))
		}
	}
}

func TestSecurityAgentCRDSchemaRequiresTopLevelSpecFields(t *testing.T) {
	root := moduleRoot(t)
	crdPath := filepath.Join(root, "config", "crd", "bases", "securityagents.kube-sentinel.io_securityagents.yaml")

	manifest, err := os.ReadFile(crdPath)
	if err != nil {
		t.Fatalf("SecurityAgent CRD manifest must exist at %s so required spec fields are pinned: %v", crdPath, err)
	}

	requiredFields := []string{"global", "features", "output", "override", "tests"}
	requiredBlock := "\n            required:\n"
	for _, fieldName := range requiredFields {
		requiredBlock += "            - " + fieldName + "\n"
	}

	if !strings.Contains(string(manifest), requiredBlock) {
		t.Fatalf("SecurityAgent CRD schema must require top-level spec fields %v", requiredFields)
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

func TestSecurityAgentCRDReferencesGlobalTargetNamespaceForNamespacedWorkloadTargeting(t *testing.T) {
	root := moduleRoot(t)
	crdPath := filepath.Join(root, "config", "crd", "bases", "securityagents.kube-sentinel.io_securityagents.yaml")

	manifest, err := os.ReadFile(crdPath)
	if err != nil {
		t.Fatalf("SecurityAgent CRD manifest must exist at %s so spec.global.targetNamespace workload targeting is pinned: %v", crdPath, err)
	}

	crd := string(manifest)
	targetNamespaceIndex := strings.Index(crd, "\n                  targetNamespace:\n")
	if targetNamespaceIndex == -1 {
		t.Fatalf("SecurityAgent CRD schema must expose spec.global.targetNamespace for namespaced workload targeting")
	}

	// Scope to the global section (ends at features: sibling key at 14-space indentation)
	targetNamespaceSchema := crd[targetNamespaceIndex:]
	if featuresIndex := strings.Index(targetNamespaceSchema, "\n              features:\n"); featuresIndex != -1 {
		targetNamespaceSchema = targetNamespaceSchema[:featuresIndex]
	}

	if !strings.Contains(targetNamespaceSchema, "description:") {
		t.Fatalf("SecurityAgent CRD schema must describe spec.global.targetNamespace as the field used for namespaced workload targeting")
	}

	description := strings.ToLower(targetNamespaceSchema)
	if !strings.Contains(description, "namespaced workload") || !strings.Contains(description, "target") {
		t.Fatalf("SecurityAgent CRD schema description for spec.global.targetNamespace must reference namespaced workload targeting, got:\n%s", targetNamespaceSchema)
	}
}

func TestSecurityAgentCRDNamespacedWorkloadTargetingReferencesExactSpecPath(t *testing.T) {
	root := moduleRoot(t)
	crdPath := filepath.Join(root, "config", "crd", "bases", "securityagents.kube-sentinel.io_securityagents.yaml")

	manifest, err := os.ReadFile(crdPath)
	if err != nil {
		t.Fatalf("SecurityAgent CRD manifest must exist at %s so spec.global.targetNamespace workload targeting is pinned: %v", crdPath, err)
	}

	crd := string(manifest)
	if !strings.Contains(crd, "\n  scope: Cluster\n") {
		t.Fatalf("SecurityAgent CRD must be cluster-scoped when registered with the cluster")
	}

	targetNamespaceIndex := strings.Index(crd, "\n                  targetNamespace:\n")
	if targetNamespaceIndex == -1 {
		t.Fatalf("SecurityAgent CRD schema must expose spec.global.targetNamespace for namespaced workload targeting")
	}

	targetNamespaceSchema := crd[targetNamespaceIndex:]
	if featuresIndex := strings.Index(targetNamespaceSchema, "\n              features:\n"); featuresIndex != -1 {
		targetNamespaceSchema = targetNamespaceSchema[:featuresIndex]
	}

	if !strings.Contains(targetNamespaceSchema, "spec.global.targetNamespace") {
		t.Fatalf("SecurityAgent CRD namespaced workload targeting guidance must reference spec.global.targetNamespace exactly, got:\n%s", targetNamespaceSchema)
	}
}

func TestSecurityAgentCRDDefinesGlobalTargetNamespaceFieldSchema(t *testing.T) {
	root := moduleRoot(t)
	crdPath := filepath.Join(root, "config", "crd", "bases", "securityagents.kube-sentinel.io_securityagents.yaml")

	manifest, err := os.ReadFile(crdPath)
	if err != nil {
		t.Fatalf("SecurityAgent CRD manifest must exist at %s so spec.global.targetNamespace field schema is pinned: %v", crdPath, err)
	}

	crd := string(manifest)
	targetNamespaceIndex := strings.Index(crd, "\n                  targetNamespace:\n")
	if targetNamespaceIndex == -1 {
		t.Fatalf("SecurityAgent CRD schema must expose spec.global.targetNamespace for namespaced workload targeting")
	}

	// Scope to the global section (ends at features: sibling key at 14-space indentation)
	targetNamespaceSchema := crd[targetNamespaceIndex:]
	if featuresIndex := strings.Index(targetNamespaceSchema, "\n              features:\n"); featuresIndex != -1 {
		targetNamespaceSchema = targetNamespaceSchema[:featuresIndex]
	}

	if !strings.Contains(targetNamespaceSchema, "\n                    type: string\n") {
		t.Fatalf("SecurityAgent CRD schema must define spec.global.targetNamespace as a string field, got:\n%s", targetNamespaceSchema)
	}

	descriptionIndex := strings.Index(targetNamespaceSchema, "\n                    description:")
	if descriptionIndex == -1 {
		t.Fatalf("SecurityAgent CRD schema must describe spec.global.targetNamespace directly, got:\n%s", targetNamespaceSchema)
	}

	description := strings.ToLower(targetNamespaceSchema[descriptionIndex:])
	if !strings.Contains(description, "namespaced workload") || !strings.Contains(description, "target") {
		t.Fatalf("SecurityAgent CRD schema description for spec.global.targetNamespace must reference namespaced workload targeting, got:\n%s", targetNamespaceSchema)
	}
}

func TestSecurityAgentResourceMetadataIsNotANamespaceRestriction(t *testing.T) {
	resourceType := reflect.TypeOf(v1alpha1.SecurityAgent{})

	if _, ok := fieldByJSONName(resourceType, "metadata"); !ok {
		t.Fatalf("SecurityAgent resource must expose Kubernetes metadata separately from spec/status so the resource itself is not scoped through a namespace field")
	}

	if _, ok := fieldByJSONName(resourceType, "namespace"); ok {
		t.Fatalf("SecurityAgent resource must not expose a top-level namespace field; cluster scope belongs to the CRD, not a namespace restriction on the resource itself")
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
