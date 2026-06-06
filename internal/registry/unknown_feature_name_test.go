package registry_test

import (
	"errors"
	"testing"

	"github.com/bhshin/kube-sentinel/internal/registry"
)

func TestUnknownFeatureNameReturnsErrorIdentifyingUnrecognisedFeatureName(t *testing.T) {
	const featureName = "test-unrecognised-feature-name"

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

	if !errors.Is(err, &registry.ValidationError{}) {
		t.Fatalf("ValidateFeatureName(%q) error is not matchable as any ValidationError", featureName)
	}
}
