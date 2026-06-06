package registry_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/bhshin/kube-sentinel/internal/registry"
)

func TestUnknownFeatureNameErrorIdentifiesExactSuppliedFeatureName(t *testing.T) {
	const featureName = `req-05-unrecognised-feature\name`

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

	if !strings.Contains(err.Error(), featureName) {
		t.Fatalf("ValidateFeatureName(%q) error %q does not identify the exact unrecognised feature name", featureName, err.Error())
	}
}
