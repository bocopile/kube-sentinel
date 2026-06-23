// Package artifactstore defines the operator-side Artifact Store abstraction.
// The operator WRITES derived artifacts (SBOM, scanner baseline, evidence
// bundle, exported report) here; raw scanner output and normalized findings
// stay in PostgreSQL (see docs/DATABASE.md, docs/ARCHITECTURE.md). Backends
// (filesystem default, S3-compatible, SeaweedFS, PVC) implement this interface
// behind a plugin boundary; only the filesystem backend is targeted for the
// first MVP. Backend implementations are added in M1.
package artifactstore

import (
	"context"
	"io"
)

// ArtifactRef identifies one stored artifact.
type ArtifactRef struct {
	// Type is the artifact kind: sbom, scanner_baseline, evidence_bundle,
	// human_report, artifact_input_manifest, ... (matches artifact_index.artifact_type).
	Type string
	// Key is the backend-relative storage path (digest-based where applicable).
	Key string
	// Checksum is the sha256 of the stored content, when known.
	Checksum string
}

// ArtifactStore is the write+read interface the operator uses. Backends are
// registered behind this boundary so storage choice never leaks into feature or
// reconciler code.
type ArtifactStore interface {
	// Put stores content and returns the resolved reference (key/checksum filled in).
	Put(ctx context.Context, ref ArtifactRef, content io.Reader) (ArtifactRef, error)
	// Get opens stored content for reading.
	Get(ctx context.Context, ref ArtifactRef) (io.ReadCloser, error)
	// List returns artifacts under a key prefix.
	List(ctx context.Context, prefix string) ([]ArtifactRef, error)
	// GenerateDownloadURL returns a retrievable URL for a stored artifact when
	// the backend supports it (else an error the caller may treat as optional).
	GenerateDownloadURL(ctx context.Context, ref ArtifactRef) (string, error)
}
