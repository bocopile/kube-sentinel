package registry_test

import (
	"reflect"
	"testing"

	"github.com/bhshin/kube-sentinel/internal/registry"
)

func TestREQP0KubeSentinelProject06DistinctPriorityFeaturesAreListedByAscendingPriority(t *testing.T) {
	features := []registry.Feature{
		{ID: "req-p0-06-zeta", Priority: 1},
		{ID: "req-p0-06-alpha", Priority: 3},
		{ID: "req-p0-06-delta", Priority: 2},
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
		{ID: "req-p0-06-zeta", Priority: 1},
		{ID: "req-p0-06-delta", Priority: 2},
		{ID: "req-p0-06-alpha", Priority: 3},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() ordered distinct-priority features = %#v, want %#v", got, want)
	}
}
