package feature

import (
	"encoding/json"
	"sort"

	securityv1alpha1 "github.com/bocopile/kube-sentinel/operator/api/v1alpha1"
)

// featurePriority is the canonical registry priority for every known feature ID
// (docs/PLAN.md §우선순위 Registry). Its key set also serves as the set of known
// feature IDs for unknown-name detection, so the resolver is testable without
// requiring real features to self-register.
var featurePriority = map[string]int{
	"target_preflight":       10,
	"bootstrap":              20,
	"source_security":        50,
	"secret_scan":            50,
	"image_vulnerability":    100,
	"image_integrity":        100,
	"sbom":                   100,
	"kubernetes_manifest":    150,
	"rbac_review":            150,
	"dockerfile_scan":        150,
	"script_scan":            150,
	"applied_cluster_config": 200,
	"secret_reference":       200,
	"trivy_operator_reports": 200,
	"remediation_enrichment": 250,
	"report_export":          300,
}

// profileFeatures is the canonical profile→base feature ID mapping
// (docs/ARCHITECTURE.md §profile→registry feature ID 정본 표).
var profileFeatures = map[securityv1alpha1.ScanProfile][]string{
	securityv1alpha1.ProfileSourceSecurity:         {"source_security", "secret_scan"},
	securityv1alpha1.ProfileImageSupplyChain:       {"image_vulnerability", "image_integrity", "sbom"},
	securityv1alpha1.ProfileKubernetesConfig:       {"kubernetes_manifest", "rbac_review"},
	securityv1alpha1.ProfileBuildAndDeploy:         {"dockerfile_scan", "script_scan"},
	securityv1alpha1.ProfileRBACAndSecretReference: {"applied_cluster_config", "rbac_review", "secret_reference"},
}

// umbrellaFeatures expands convenience umbrella names in features[].name
// (docs/ARCHITECTURE.md umbrella table). A name that is not an umbrella is
// treated as a single feature ID.
var umbrellaFeatures = map[string][]string{
	"trivy": {"image_vulnerability", "image_integrity", "sbom", "trivy_operator_reports"},
	"security_assessment": {
		"source_security", "secret_scan", "kubernetes_manifest", "rbac_review",
		"dockerfile_scan", "script_scan", "applied_cluster_config", "secret_reference",
	},
}

// MergeResult is the resolved enabled feature set plus any inputs that could not
// be resolved (surfaced as ConfigError in ScanRun.status.features[]).
type MergeResult struct {
	// EnabledIDs is the final enabled feature set, sorted by priority ascending
	// then ID lexicographically.
	EnabledIDs []string
	// UnknownProfiles holds spec.profiles[] values with no canonical mapping.
	// (In practice CRD enum validation rejects these before reconcile.)
	UnknownProfiles []string
	// UnknownFeatures holds features[].name values that are neither an umbrella
	// name nor a known feature ID.
	UnknownFeatures []string
	// Configs holds the last-wins opaque config per feature ID.
	Configs map[string]json.RawMessage
}

// MergeFeatures resolves the enabled feature set from profiles and feature
// overrides following docs/ARCHITECTURE.md merge rules:
//  1. (caller) ScanRun.spec.profiles overrides SecurityAssessment.spec.profiles
//  2. expand profiles to the base feature set
//  3. expand features[].name via the umbrella table
//  4. apply enabled=true (union) / enabled=false (difference) in declared order
//  5. features[].config overrides defaults; last item wins
//  6. final set = (profiles ∪ enabled:true) − (enabled:false)
//  7. sort by priority ascending, then ID lexicographically
func MergeFeatures(profiles []securityv1alpha1.ScanProfile, features []securityv1alpha1.FeatureSpec) MergeResult {
	enabled := map[string]bool{}
	configs := map[string]json.RawMessage{}
	var unknownProfiles, unknownFeatures []string

	// 2. expand profiles
	for _, p := range profiles {
		ids, ok := profileFeatures[p]
		if !ok {
			unknownProfiles = append(unknownProfiles, string(p))
			continue
		}
		for _, id := range ids {
			enabled[id] = true
		}
	}

	// 3-5. apply feature overrides in declared order
	for _, f := range features {
		ids, isUmbrella := umbrellaFeatures[f.Name]
		if !isUmbrella {
			if _, known := featurePriority[f.Name]; !known {
				unknownFeatures = append(unknownFeatures, f.Name)
				continue
			}
			ids = []string{f.Name}
		}
		for _, id := range ids {
			if f.Enabled {
				enabled[id] = true
			} else {
				delete(enabled, id)
			}
			if len(f.Config.Raw) > 0 {
				configs[id] = append(json.RawMessage(nil), f.Config.Raw...) // last wins
			}
		}
	}

	// 6-7. collect and sort
	out := make([]string, 0, len(enabled))
	for id := range enabled {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool {
		if featurePriority[out[i]] != featurePriority[out[j]] {
			return featurePriority[out[i]] < featurePriority[out[j]]
		}
		return out[i] < out[j]
	})

	return MergeResult{
		EnabledIDs:      out,
		UnknownProfiles: unknownProfiles,
		UnknownFeatures: unknownFeatures,
		Configs:         configs,
	}
}
