package registry_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/bhshin/kube-sentinel/internal/registry"
)

func TestUnknownEmptyFeatureNameErrorIdentifiesSuppliedFeatureName(t *testing.T) {
	const featureName = ""

	err := registry.ValidateFeatureName(featureName)
	if err == nil {
		t.Fatalf("ValidateFeatureName(%q) returned nil error, want ValidationError", featureName)
	}

	var validationError *registry.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("ValidateFeatureName(%q) error %T is not a ValidationError", featureName, err)
	}

	if validationError.FeatureName != featureName {
		t.Fatalf("ValidationError.FeatureName = %q, want %q", validationError.FeatureName, featureName)
	}

	if !strings.Contains(err.Error(), `""`) {
		t.Fatalf("ValidateFeatureName(%q) error %q does not identify the exact supplied Feature name", featureName, err.Error())
	}
}
