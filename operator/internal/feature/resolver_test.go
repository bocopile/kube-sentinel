package feature

import (
	"reflect"
	"testing"

	securityv1alpha1 "github.com/bocopile/kube-sentinel/operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestMergeFeaturesProfilesExpandAndOrder(t *testing.T) {
	res := MergeFeatures([]securityv1alpha1.ScanProfile{
		securityv1alpha1.ProfileBuildAndDeploy,   // dockerfile_scan(150), script_scan(150)
		securityv1alpha1.ProfileSourceSecurity,   // source_security(50), secret_scan(50)
		securityv1alpha1.ProfileImageSupplyChain, // image_vulnerability(100), image_integrity(100), sbom(100)
	}, nil)

	want := []string{
		"secret_scan", "source_security", // 50, lex
		"image_integrity", "image_vulnerability", "sbom", // 100, lex
		"dockerfile_scan", "script_scan", // 150, lex
	}
	if !reflect.DeepEqual(res.EnabledIDs, want) {
		t.Fatalf("EnabledIDs = %v, want %v", res.EnabledIDs, want)
	}
	if len(res.UnknownProfiles) != 0 || len(res.UnknownFeatures) != 0 {
		t.Fatalf("unexpected unknowns: profiles=%v features=%v", res.UnknownProfiles, res.UnknownFeatures)
	}
}

func TestMergeFeaturesUmbrellaExpansion(t *testing.T) {
	res := MergeFeatures(nil, []securityv1alpha1.FeatureSpec{
		{Name: "trivy", Enabled: true},
	})
	want := []string{"image_integrity", "image_vulnerability", "sbom", "trivy_operator_reports"}
	if !reflect.DeepEqual(res.EnabledIDs, want) {
		t.Fatalf("EnabledIDs = %v, want %v", res.EnabledIDs, want)
	}
}

func TestMergeFeaturesEnabledFalseRemoves(t *testing.T) {
	res := MergeFeatures(
		[]securityv1alpha1.ScanProfile{securityv1alpha1.ProfileImageSupplyChain}, // image_vuln, image_integrity, sbom
		[]securityv1alpha1.FeatureSpec{
			{Name: "sbom", Enabled: false}, // difference
			{Name: "report_export", Enabled: true},
		},
	)
	want := []string{"image_integrity", "image_vulnerability", "report_export"}
	if !reflect.DeepEqual(res.EnabledIDs, want) {
		t.Fatalf("EnabledIDs = %v, want %v", res.EnabledIDs, want)
	}
}

func TestMergeFeaturesUnknownSeparated(t *testing.T) {
	res := MergeFeatures(
		[]securityv1alpha1.ScanProfile{securityv1alpha1.ProfileSourceSecurity, "BogusProfile"},
		[]securityv1alpha1.FeatureSpec{
			{Name: "source_security", Enabled: true},
			{Name: "not_a_real_feature", Enabled: true},
		},
	)
	if want := []string{"BogusProfile"}; !reflect.DeepEqual(res.UnknownProfiles, want) {
		t.Fatalf("UnknownProfiles = %v, want %v", res.UnknownProfiles, want)
	}
	if want := []string{"not_a_real_feature"}; !reflect.DeepEqual(res.UnknownFeatures, want) {
		t.Fatalf("UnknownFeatures = %v, want %v", res.UnknownFeatures, want)
	}
	// known features still resolve despite the unknowns
	if want := []string{"secret_scan", "source_security"}; !reflect.DeepEqual(res.EnabledIDs, want) {
		t.Fatalf("EnabledIDs = %v, want %v", res.EnabledIDs, want)
	}
}

func TestMergeFeaturesConfigLastWins(t *testing.T) {
	res := MergeFeatures(nil, []securityv1alpha1.FeatureSpec{
		{Name: "sbom", Enabled: true, Config: runtime.RawExtension{Raw: []byte(`{"v":1}`)}},
		{Name: "sbom", Enabled: true, Config: runtime.RawExtension{Raw: []byte(`{"v":2}`)}},
	})
	got := string(res.Configs["sbom"])
	if got != `{"v":2}` {
		t.Fatalf("Configs[sbom] = %s, want {\"v\":2}", got)
	}
}
