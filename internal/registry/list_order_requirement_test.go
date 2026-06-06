package registry_test

import (
	"reflect"
	"testing"

	"github.com/bhshin/kube-sentinel/internal/registry"
)

func TestRegisteredFeaturesAreListedByAscendingPriorityThenLexicographicID(t *testing.T) {
	features := []registry.Feature{
		{ID: "req-06-order-charlie", Priority: 30},
		{ID: "req-06-order-bravo", Priority: 20},
		{ID: "req-06-order-delta", Priority: 30},
		{ID: "req-06-order-alpha", Priority: 10},
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
		{ID: "req-06-order-alpha", Priority: 10},
		{ID: "req-06-order-bravo", Priority: 20},
		{ID: "req-06-order-charlie", Priority: 30},
		{ID: "req-06-order-delta", Priority: 30},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() ordered registered features = %#v, want %#v", got, want)
	}
}
