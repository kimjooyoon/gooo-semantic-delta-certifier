package delta

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

func canonicalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

func digestValue(value any) (string, error) {
	raw, err := canonicalJSON(value)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(hash[:]), nil
}

func semanticIRDigest(ir SemanticIR) (string, error) {
	copyIR := ir
	copyIR.Digest = ""
	copyIR.ObservedDigest = ""
	return digestValue(copyIR)
}

func releaseDigest(release ReleaseBinding) (string, error) {
	copyRelease := release
	copyRelease.ReleaseDigest = ""
	copyRelease.ObservedReleaseDigest = ""
	return digestValue(copyRelease)
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || len(value) < 8 || value[:7] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}

func validateCanonicalBinding(release ReleaseBinding, side string, unknowns *[]Unknown, contradictions *[]string) {
	if release.Schema == "" || release.ReleaseID == "" || release.Tag == "" {
		*contradictions = append(*contradictions, side+":MISSING_RELEASE_IDENTITY")
	}
	if release.SemanticIR.Schema != "gooo/semantic-delta-certifier/semantic-ir-input/v1" {
		*contradictions = append(*contradictions, side+":INVALID_SEMANTIC_IR_SCHEMA")
	}
	if release.ReleaseDigest == "" || release.SemanticIR.Digest == "" {
		appendUnknown(unknowns, Unknown{
			Stage: "INPUT_BINDING", Step: "BIND_IMMUTABLE_RELEASE", Reason: side+"_DIGEST_UNAVAILABLE",
			UnknownClass: "DIRECT_MISSING", NextOperation: "PROVIDE_IMMUTABLE_RELEASE_AND_SEMANTIC_IR_DIGESTS", BlockedBy: []string{side + "-release-binding"},
		})
	}
	if release.ReleaseDigest != "" && !validDigest(release.ReleaseDigest) {
		*contradictions = append(*contradictions, side+":INVALID_RELEASE_DIGEST")
	}
	if release.SemanticIR.Digest != "" && !validDigest(release.SemanticIR.Digest) {
		*contradictions = append(*contradictions, side+":INVALID_SEMANTIC_IR_DIGEST")
	}
	if release.SemanticIRDigest != "" && !validDigest(release.SemanticIRDigest) {
		*contradictions = append(*contradictions, side+":INVALID_RELEASE_SEMANTIC_IR_DIGEST")
	}
	if release.ObservedReleaseDigest == "" || release.ObservedSemanticIRDigest == "" {
		appendUnknown(unknowns, Unknown{
			Stage: "INPUT_BINDING", Step: "OBSERVE_IMMUTABLE_DIGESTS", Reason: side+"_OBSERVED_DIGEST_UNAVAILABLE",
			UnknownClass: "DIRECT_MISSING", NextOperation: "PROVIDE_OBSERVED_RELEASE_DIGESTS", BlockedBy: []string{side + "-observed-digests"},
		})
	}
	if release.ObservedReleaseDigest != "" && release.ObservedReleaseDigest != release.ReleaseDigest {
		*contradictions = append(*contradictions, side+":STALE_RELEASE_DIGEST")
	}
	if release.ObservedSemanticIRDigest != "" && release.ObservedSemanticIRDigest != release.SemanticIR.Digest {
		*contradictions = append(*contradictions, side+":STALE_SEMANTIC_IR_DIGEST")
	}
	if release.SemanticIRDigest != "" && release.SemanticIRDigest != release.SemanticIR.Digest {
		*contradictions = append(*contradictions, side+":RELEASE_IR_BINDING_MISMATCH")
	}
	if release.SemanticIR.Digest != "" {
		computed, err := semanticIRDigest(release.SemanticIR)
		if err != nil || computed != release.SemanticIR.Digest {
			*contradictions = append(*contradictions, side+":SEMANTIC_IR_CONTENT_DIGEST_MISMATCH")
		}
	}
	if release.ReleaseDigest != "" {
		computed, err := releaseDigest(release)
		if err != nil || computed != release.ReleaseDigest {
			*contradictions = append(*contradictions, side+":RELEASE_CONTENT_DIGEST_MISMATCH")
		}
	}
}

func appendUnknown(unknowns *[]Unknown, value Unknown) {
	if value.BlockedBy == nil {
		value.BlockedBy = []string{}
	}
	*unknowns = append(*unknowns, value)
}

func requireDigest(digest string, domain, id string) *Unknown {
	if digest != "" {
		return nil
	}
	return &Unknown{
		Stage: "DELTA_COMPARE", Step: "COMPARE_EXACT_DIGEST", Reason: "EXACT_DIGEST_UNAVAILABLE",
		UnknownClass: "DIRECT_MISSING", NextOperation: "PROVIDE_EXACT_SEMANTIC_DIGEST", BlockedBy: []string{domain + ":" + id},
	}
}

func digestPointer(value string) *string {
	if value == "" {
		return nil
	}
	result := value
	return &result
}

func intPointer(value int) *int {
	result := value
	return &result
}

func validUnknown(value *Unknown) error {
	if value == nil {
		return errors.New("UNKNOWN record is nil")
	}
	if value.Stage == "" || value.Step == "" || value.Reason == "" || value.UnknownClass == "" || value.NextOperation == "" || value.BlockedBy == nil {
		return errors.New("UNKNOWN record must preserve six fields")
	}
	return nil
}

func ensureStatus(value string) error {
	for _, item := range generated.StatusValues {
		if item == value {
			return nil
		}
	}
	return fmt.Errorf("unsupported status %q", value)
}
