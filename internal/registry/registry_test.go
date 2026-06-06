package registry_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bhshin/kube-sentinel/internal/registry"
)

func TestUnknownFeatureNameReturnsDetectableValidationError(t *testing.T) {
	t.Parallel()

	const featureName = "not-in-feature-registry"

	err := registry.ValidateFeatureName(featureName)
	if err == nil {
		t.Fatalf("ValidateFeatureName(%q) returned nil error, want validation error", featureName)
	}

	var validationError *registry.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("ValidateFeatureName(%q) error %T is not detectable as *registry.ValidationError", featureName, err)
	}

	if validationError.FeatureName != featureName {
		t.Fatalf("validation error feature name = %q, want %q", validationError.FeatureName, featureName)
	}
}

func TestInitialisedKnownFeatureNamesRejectUnknownFeatureName(t *testing.T) {
	knownFeatureNames := []string{
		"known-feature-alpha",
		"known-feature-beta",
	}
	registry.InitialiseKnownFeatureNames(knownFeatureNames)

	const knownFeatureName = "known-feature-alpha"
	if err := registry.ValidateFeatureName(knownFeatureName); err != nil {
		t.Fatalf("ValidateFeatureName(%q) returned error %v, want nil", knownFeatureName, err)
	}

	const unknownFeatureName = "known-feature-gamma"
	err := registry.ValidateFeatureName(unknownFeatureName)
	if err == nil {
		t.Fatalf("ValidateFeatureName(%q) returned nil error, want validation error", unknownFeatureName)
	}

	var validationError *registry.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("ValidateFeatureName(%q) error %T is not detectable as *registry.ValidationError", unknownFeatureName, err)
	}

	if validationError.FeatureName != unknownFeatureName {
		t.Fatalf("validation error feature name = %q, want %q", validationError.FeatureName, unknownFeatureName)
	}
}

func TestListReturnsFeaturesInDeterministicPriorityThenIDOrder(t *testing.T) {
	features := []registry.Feature{
		{ID: "test-registry-order-zeta", Priority: 20},
		{ID: "test-registry-order-alpha", Priority: 10},
		{ID: "test-registry-order-beta", Priority: 10},
		{ID: "test-registry-order-gamma", Priority: -5},
	}

	registeredIDs := make(map[string]bool, len(features))
	for _, feature := range features {
		registeredIDs[feature.ID] = true
		registry.Register(feature.ID, feature.Priority)
	}

	var got []registry.Feature
	for _, feature := range registry.List() {
		if registeredIDs[feature.ID] {
			got = append(got, feature)
		}
	}

	want := []registry.Feature{
		{ID: "test-registry-order-gamma", Priority: -5},
		{ID: "test-registry-order-alpha", Priority: 10},
		{ID: "test-registry-order-beta", Priority: 10},
		{ID: "test-registry-order-zeta", Priority: 20},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() filtered order = %#v, want %#v", got, want)
	}
}
