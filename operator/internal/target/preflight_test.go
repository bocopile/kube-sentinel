package target

import (
	"reflect"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestClassifyAPIError(t *testing.T) {
	gr := schema.GroupResource{Resource: "pods"}
	cases := []struct {
		name       string
		err        error
		wantPhase  string
		wantReason string
	}{
		{"unauthorized", apierrors.NewUnauthorized("bad token"), PhaseAuthFailed, ReasonAuthFailed},
		{"forbidden", apierrors.NewForbidden(gr, "x", nil), PhasePermissionDenied, ReasonPermissionDenied},
		{"network", &fakeNetErr{}, PhaseUnreachable, ReasonUnreachable},
		{"timeout", apierrors.NewServerTimeout(gr, "list", 1), PhaseUnreachable, ReasonUnreachable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			phase, reason := ClassifyAPIError(tc.err)
			if phase != tc.wantPhase || reason != tc.wantReason {
				t.Fatalf("ClassifyAPIError(%v) = (%s,%s), want (%s,%s)", tc.err, phase, reason, tc.wantPhase, tc.wantReason)
			}
		})
	}
}

type fakeNetErr struct{}

func (*fakeNetErr) Error() string { return "dial tcp 10.0.0.1:6443: connect: connection refused" }

func TestNamespaceVisibilityWithAllowlist(t *testing.T) {
	visible, missing := NamespaceVisibility(
		[]string{"app", "platform", "ghost"},
		[]string{"default", "kube-system", "app", "platform"},
	)
	if want := []string{"app", "platform"}; !reflect.DeepEqual(visible, want) {
		t.Fatalf("visible = %v, want %v", visible, want)
	}
	if want := []string{"ghost"}; !reflect.DeepEqual(missing, want) {
		t.Fatalf("missing = %v, want %v", missing, want)
	}
}

func TestNamespaceVisibilityEmptyAllowlistExcludesSystem(t *testing.T) {
	visible, missing := NamespaceVisibility(
		nil,
		[]string{"kube-system", "app", "kube-public", "platform", "default"},
	)
	// system (kube-*) excluded, rest sorted; "default" is not kube-* so it stays.
	if want := []string{"app", "default", "platform"}; !reflect.DeepEqual(visible, want) {
		t.Fatalf("visible = %v, want %v", visible, want)
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %v, want empty", missing)
	}
}

func TestIsSystemNamespace(t *testing.T) {
	for _, n := range []string{"kube-system", "kube-public", "kube-node-lease"} {
		if !isSystemNamespace(n) {
			t.Fatalf("%q should be system", n)
		}
	}
	for _, n := range []string{"default", "app", "platform"} {
		if isSystemNamespace(n) {
			t.Fatalf("%q should not be system", n)
		}
	}
}

func TestFailResultShape(t *testing.T) {
	r := FailResult(PhaseAuthFailed, CondKubeconfigValid, ReasonInvalidSecret, "secret gone")
	if r.Phase != PhaseAuthFailed {
		t.Fatalf("phase = %s", r.Phase)
	}
	if len(r.Conditions) != 1 || r.Conditions[0].OK || r.Conditions[0].Type != CondKubeconfigValid {
		t.Fatalf("unexpected conditions: %+v", r.Conditions)
	}
}
