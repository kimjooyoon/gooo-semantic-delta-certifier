package delta

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/gooo-semantic-delta-certifier/internal/generated"
)

type observedRecord struct {
	ID string
	Kind string
	Name string
	From string
	To string
	Operation string
	Path string
	Digest string
}

func EvaluateBytes(raw []byte, inventory Inventory) Evidence {
	var input Input
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return invalidEvidence("MALFORMED_INPUT", inventory, "INPUT:MALFORMED_JSON")
	}
	return Evaluate(input, inventory)
}

func Evaluate(input Input, inventory Inventory) Evidence {
	evidence := baseEvidence(input.CaseID, inventory)
	unknowns := []Unknown{}
	contradictions := []string{}
	if input.Schema != InputSchema || input.CaseID == "" {
		contradictions = append(contradictions, "INPUT:INVALID_SCHEMA_OR_CASE_ID")
	}
	validateCanonicalBinding(input.Old, "OLD", &unknowns, &contradictions)
	validateCanonicalBinding(input.New, "NEW", &unknowns, &contradictions)

	oldEntities, oldEntityDuplicates := entityRecords(input.Old.SemanticIR.Entities, "OLD", &unknowns, &contradictions)
	newEntities, newEntityDuplicates := entityRecords(input.New.SemanticIR.Entities, "NEW", &unknowns, &contradictions)
	oldRelations, oldRelationDuplicates := relationRecords(input.Old.SemanticIR.Relations, "OLD", &unknowns, &contradictions)
	newRelations, newRelationDuplicates := relationRecords(input.New.SemanticIR.Relations, "NEW", &unknowns, &contradictions)
	oldEffects, oldEffectDuplicates := effectRecords(input.Old.SemanticIR.Effects, "OLD", &unknowns, &contradictions)
	newEffects, newEffectDuplicates := effectRecords(input.New.SemanticIR.Effects, "NEW", &unknowns, &contradictions)
	oldCapabilities, oldCapabilityDuplicates := capabilityRecords(input.Old.SemanticIR.Capabilities, "OLD", &unknowns, &contradictions)
	newCapabilities, newCapabilityDuplicates := capabilityRecords(input.New.SemanticIR.Capabilities, "NEW", &unknowns, &contradictions)
	oldArtifacts, oldArtifactDuplicates := artifactRecords(input.Old.SemanticIR.GeneratedArtifacts, "OLD", &unknowns, &contradictions)
	newArtifacts, newArtifactDuplicates := artifactRecords(input.New.SemanticIR.GeneratedArtifacts, "NEW", &unknowns, &contradictions)

	evidence.DeltaCells = append(evidence.DeltaCells, compareDomain("ENTITY", oldEntities, newEntities, oldEntityDuplicates, newEntityDuplicates)...)
	evidence.DeltaCells = append(evidence.DeltaCells, compareDomain("RELATION", oldRelations, newRelations, oldRelationDuplicates, newRelationDuplicates)...)
	evidence.DeltaCells = append(evidence.DeltaCells, compareDomain("EFFECT", oldEffects, newEffects, oldEffectDuplicates, newEffectDuplicates)...)
	evidence.DeltaCells = append(evidence.DeltaCells, compareDomain("CAPABILITY", oldCapabilities, newCapabilities, oldCapabilityDuplicates, newCapabilityDuplicates)...)
	evidence.DeltaCells = append(evidence.DeltaCells, compareDomain("ARTIFACT", oldArtifacts, newArtifacts, oldArtifactDuplicates, newArtifactDuplicates)...)
	sort.Slice(evidence.DeltaCells, func(i, j int) bool {
		if evidence.DeltaCells[i].Domain != evidence.DeltaCells[j].Domain {
			return evidence.DeltaCells[i].Domain < evidence.DeltaCells[j].Domain
		}
		return evidence.DeltaCells[i].Identity < evidence.DeltaCells[j].Identity
	})
	evidence.Improvement, improvementUnknown, improvementRefuted := compareImprovement(input.Improvement)
	unknowns = append(unknowns, improvementUnknown...)
	contradictions = append(contradictions, improvementRefuted...)
	evidence.Unknowns = uniqueUnknowns(unknowns)
	evidence.Contradictions = uniqueStrings(contradictions)
	evidence.State, evidence.Reason = resolve(evidence.Contradictions, evidence.Unknowns)
	return evidence
}

func baseEvidence(caseID string, inventory Inventory) Evidence {
	proof := make([]ProofVectorCell, 0, len(generated.Cells))
	indicators := make([]IndicatorVectorCell, 0, len(generated.Cells))
	for _, cell := range generated.Cells {
		proof = append(proof, ProofVectorCell{CellID: cell.ID, ProofChoice: cell.Proof})
		indicators = append(indicators, IndicatorVectorCell{CellID: cell.ID, Indicator: cell.Indicator})
	}
	return Evidence{
		Schema: EvidenceSchema, CaseID: caseID, Precedence: append([]string{}, generated.StatusPrecedence...),
		DenominatorID: generated.DenominatorID, DenominatorTotal: len(generated.Cells),
		DeltaCells: []DeltaCell{}, ProofVector: proof, IndicatorVector: indicators,
		Improvement: ImprovementEvidence{Before: nil, After: nil, ExactDigest: nil, Unknown: nil},
		Unknowns: []Unknown{}, Contradictions: []string{}, Authority: Authority{RepositoryWrites: 0, CallerOwnedOutput: true},
		Inventory: inventory, Measurements: unknownMeasurements(),
	}
}

func invalidEvidence(reason string, inventory Inventory, contradiction string) Evidence {
	evidence := baseEvidence("MALFORMED_INPUT", inventory)
	evidence.State = StatusRefuted
	evidence.Reason = reason
	evidence.Contradictions = []string{contradiction}
	return evidence
}

func unknownMeasurements() Measurements {
	return Measurements{
		Build: unknownMeasurement("CI_BUILD", "BUILD_MEASUREMENT_UNAVAILABLE"),
		Test: unknownMeasurement("CI_TEST", "TEST_MEASUREMENT_UNAVAILABLE"),
		Conformance: unknownMeasurement("CI_CONFORMANCE", "CONFORMANCE_MEASUREMENT_UNAVAILABLE"),
		Wall: unknownMeasurement("CI_RUNTIME", "WALL_TIME_UNAVAILABLE"),
		PeakRSS: unknownMeasurement("CI_RUNTIME", "PEAK_RSS_UNAVAILABLE"),
		Cache: unknownMeasurement("CI_CACHE", "CACHE_OBSERVATION_UNAVAILABLE"),
	}
}

func unknownMeasurement(stage, reason string) Measurement {
	return Measurement{Value: nil, Unit: "", Status: StatusUnknown, Unknown: &Unknown{
		Stage: stage, Step: "OBSERVE_CI_MEASUREMENT", Reason: reason, UnknownClass: "ENVIRONMENT_UNAVAILABLE",
		NextOperation: "PROVIDE_GITHUB_ACTIONS_MEASUREMENT", BlockedBy: []string{"github-actions-run"},
	}}
}

func compareImprovement(claim *ImprovementClaim) (ImprovementEvidence, []Unknown, []string) {
	result := ImprovementEvidence{ClaimedState: "", Status: StatusUnknown, Before: nil, After: nil, PairID: "", BeforeIdentity: "", AfterIdentity: "", ExactDigest: nil, Unknown: nil}
	if claim == nil {
		result.Reason = "MATCHED_BEFORE_AFTER_EVIDENCE_UNAVAILABLE"
		result.Unknown = &Unknown{Stage: "IMPROVEMENT", Step: "REQUIRE_EXACT_PAIR", Reason: result.Reason, UnknownClass: "DIRECT_MISSING", NextOperation: "PROVIDE_MATCHED_BEFORE_AFTER_INTEGER_OR_EXACT_DIGEST", BlockedBy: []string{"improvement-pair"}}
		return result, []Unknown{*result.Unknown}, nil
	}
	result.ClaimedState = claim.ClaimedState
	result.Before, result.After = claim.Before, claim.After
	result.PairID, result.BeforeIdentity, result.AfterIdentity = claim.PairID, claim.BeforeIdentity, claim.AfterIdentity
	if claim.ExactDigest != "" {
		result.ExactDigest = digestPointer(claim.ExactDigest)
	}
	if claim.ClaimedState == "FIXED_POINT" && !claim.Explicit {
		result.Status = StatusRefuted
		result.Reason = "FIXED_POINT_REQUIRES_EXPLICIT_TRUE"
		return result, nil, []string{"IMPROVEMENT:OPERATIONAL_REFUTED_FIXED_POINT_NOT_EXPLICIT"}
	}
	integerPair := claim.Before != nil && claim.After != nil && claim.PairID != "" && claim.BeforeIdentity != "" && claim.BeforeIdentity == claim.AfterIdentity
	digestPair := claim.PairID != "" && validDigest(claim.ExactDigest)
	if integerPair || digestPair {
		result.Status = "CLOSED"
		result.Reason = "EXACT_BEFORE_AFTER_EVIDENCE_BOUND"
		return result, nil, nil
	}
	result.Reason = "MATCHED_BEFORE_AFTER_EVIDENCE_INCOMPLETE"
	result.Unknown = &Unknown{Stage: "IMPROVEMENT", Step: "BIND_EXACT_PAIR", Reason: result.Reason, UnknownClass: "PAIR_EVIDENCE_INCOMPLETE", NextOperation: "PROVIDE_MATCHED_INTEGER_PAIR_OR_EXACT_DIGEST", BlockedBy: []string{"improvement-pair"}}
	return result, []Unknown{*result.Unknown}, nil
}

func resolve(contradictions []string, unknowns []Unknown) (string, string) {
	if len(contradictions) > 0 {
		return StatusRefuted, "KNOWN_CONTRADICTION"
	}
	if len(unknowns) > 0 {
		return StatusUnknown, "EVIDENCE_INCOMPLETE"
	}
	return StatusClosed, "ALL_DECLARED_OBSERVATIONS_BOUND"
}

func compareDomain(domain string, oldRecords, newRecords map[string]observedRecord, oldDuplicates, newDuplicates map[string]bool) []DeltaCell {
	ids := map[string]bool{}
	for id := range oldRecords { ids[id] = true }
	for id := range newRecords { ids[id] = true }
	for id := range oldDuplicates { ids[id] = true }
	for id := range newDuplicates { ids[id] = true }
	ordered := make([]string, 0, len(ids))
	for id := range ids { ordered = append(ordered, id) }
	sort.Strings(ordered)
	result := make([]DeltaCell, 0, len(ordered))
	for _, id := range ordered {
		oldValue, oldOK := oldRecords[id]
		newValue, newOK := newRecords[id]
		cell := DeltaCell{Domain: domain, Identity: id, OldDigest: nil, NewDigest: nil, Unknown: nil}
		if value, ok := oldDuplicates[id]; ok && value { cell.Status, cell.Reason = StatusRefuted, "DUPLICATE_OLD_ID"; result = append(result, cell); continue }
		if value, ok := newDuplicates[id]; ok && value { cell.Status, cell.Reason = StatusRefuted, "DUPLICATE_NEW_ID"; result = append(result, cell); continue }
		if oldOK { cell.OldDigest = digestPointer(oldValue.Digest) }
		if newOK { cell.NewDigest = digestPointer(newValue.Digest) }
		if !oldOK {
			if unknown := requireDigest(newValue.Digest, domain, id); unknown != nil { cell.Status, cell.Reason, cell.Unknown = StatusUnknown, "ADDED_DIGEST_UNAVAILABLE", unknown } else { cell.Status, cell.Reason = StatusAdded, "EXACT_ID_PRESENT_ONLY_IN_NEW" }
			result = append(result, cell); continue
		}
		if !newOK {
			if unknown := requireDigest(oldValue.Digest, domain, id); unknown != nil { cell.Status, cell.Reason, cell.Unknown = StatusUnknown, "REMOVED_DIGEST_UNAVAILABLE", unknown } else { cell.Status, cell.Reason = StatusRemoved, "EXACT_ID_PRESENT_ONLY_IN_OLD" }
			result = append(result, cell); continue
		}
		if oldValue.Kind != newValue.Kind {
			cell.Status, cell.Reason = StatusRefuted, "EXACT_ID_KIND_COLLISION"
			result = append(result, cell); continue
		}
		if unknown := requireDigest(oldValue.Digest, domain, id); unknown != nil {
			cell.Status, cell.Reason, cell.Unknown = StatusUnknown, "OLD_DIGEST_UNAVAILABLE", unknown
			result = append(result, cell); continue
		}
		if unknown := requireDigest(newValue.Digest, domain, id); unknown != nil {
			cell.Status, cell.Reason, cell.Unknown = StatusUnknown, "NEW_DIGEST_UNAVAILABLE", unknown
			result = append(result, cell); continue
		}
		if oldValue.Digest == newValue.Digest {
			if oldValue.signature() != newValue.signature() {
				cell.Status, cell.Reason = StatusRefuted, "DIGEST_CONTENT_COLLISION"
			} else {
				cell.Status, cell.Reason = StatusUnchanged, "EXACT_DIGEST_MATCHED"
			}
		} else {
			cell.Status, cell.Reason = StatusChanged, "EXACT_DIGEST_CHANGED"
		}
		result = append(result, cell)
	}
	return result
}

func (r observedRecord) signature() string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", r.ID, r.Kind, r.Name, r.From, r.To, r.Operation, r.Path)
}

func entityRecords(values []SemanticEntity, side string, unknowns *[]Unknown, contradictions *[]string) (map[string]observedRecord, map[string]bool) {
	result := map[string]observedRecord{}
	duplicates := map[string]bool{}
	if values == nil { appendUnknown(unknowns, domainUnknown(side, "ENTITY")); return result, duplicates }
	for _, value := range values {
		record := observedRecord{ID: value.ID, Kind: value.Kind, Name: value.Name, Digest: value.Digest}
		addRecord(result, duplicates, record, side, "ENTITY", unknowns, contradictions)
	}
	return result, duplicates
}

func relationRecords(values []SemanticRelation, side string, unknowns *[]Unknown, contradictions *[]string) (map[string]observedRecord, map[string]bool) {
	result := map[string]observedRecord{}
	duplicates := map[string]bool{}
	if values == nil { appendUnknown(unknowns, domainUnknown(side, "RELATION")); return result, duplicates }
	for _, value := range values {
		record := observedRecord{ID: value.ID, Kind: value.Kind, From: value.From, To: value.To, Digest: value.Digest}
		addRecord(result, duplicates, record, side, "RELATION", unknowns, contradictions)
	}
	return result, duplicates
}

func effectRecords(values []SemanticEffect, side string, unknowns *[]Unknown, contradictions *[]string) (map[string]observedRecord, map[string]bool) {
	result := map[string]observedRecord{}
	duplicates := map[string]bool{}
	if values == nil { appendUnknown(unknowns, domainUnknown(side, "EFFECT")); return result, duplicates }
	for _, value := range values {
		record := observedRecord{ID: value.ID, Kind: "EFFECT", Operation: value.Operation, Digest: value.Digest}
		addRecord(result, duplicates, record, side, "EFFECT", unknowns, contradictions)
	}
	return result, duplicates
}

func capabilityRecords(values []SemanticCapability, side string, unknowns *[]Unknown, contradictions *[]string) (map[string]observedRecord, map[string]bool) {
	result := map[string]observedRecord{}
	duplicates := map[string]bool{}
	if values == nil { appendUnknown(unknowns, domainUnknown(side, "CAPABILITY")); return result, duplicates }
	for _, value := range values {
		record := observedRecord{ID: value.ID, Kind: "CAPABILITY", Name: value.Name, Digest: value.Digest}
		addRecord(result, duplicates, record, side, "CAPABILITY", unknowns, contradictions)
	}
	return result, duplicates
}

func artifactRecords(values []GeneratedArtifact, side string, unknowns *[]Unknown, contradictions *[]string) (map[string]observedRecord, map[string]bool) {
	result := map[string]observedRecord{}
	duplicates := map[string]bool{}
	if values == nil { appendUnknown(unknowns, domainUnknown(side, "ARTIFACT")); return result, duplicates }
	for _, value := range values {
		record := observedRecord{ID: value.ID, Kind: value.Kind, Path: value.Path, Digest: value.Digest}
		addRecord(result, duplicates, record, side, "ARTIFACT", unknowns, contradictions)
	}
	return result, duplicates
}

func addRecord(result map[string]observedRecord, duplicates map[string]bool, record observedRecord, side, domain string, unknowns *[]Unknown, contradictions *[]string) {
	if record.ID == "" {
		*contradictions = append(*contradictions, side+":"+domain+":MISSING_EXACT_ID")
		return
	}
	if _, exists := result[record.ID]; exists {
		duplicates[record.ID] = true
		*contradictions = append(*contradictions, side+":"+domain+":DUPLICATE_ID:"+record.ID)
		return
	}
	if record.Digest != "" && !validDigest(record.Digest) {
		*contradictions = append(*contradictions, side+":"+domain+":INVALID_DIGEST:"+record.ID)
	}
	result[record.ID] = record
	_ = unknowns
}

func domainUnknown(side, domain string) Unknown {
	return Unknown{Stage: "DELTA_COMPARE", Step: "LOAD_SEMANTIC_DOMAIN", Reason: side+"_"+domain+"_COLLECTION_UNAVAILABLE", UnknownClass: "DIRECT_MISSING", NextOperation: "PROVIDE_SEMANTIC_IR_DOMAIN", BlockedBy: []string{side + "-" + strings.ToLower(domain)}}
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values { if !seen[value] { seen[value] = true; result = append(result, value) } }
	sort.Strings(result)
	return result
}

func uniqueUnknowns(values []Unknown) []Unknown {
	result := []Unknown{}
	seen := map[string]bool{}
	for _, value := range values {
		if value.BlockedBy == nil { value.BlockedBy = []string{} }
		key := value.Stage+"|"+value.Step+"|"+value.Reason+"|"+value.UnknownClass+"|"+value.NextOperation+"|"+strings.Join(value.BlockedBy, ",")
		if !seen[key] { seen[key] = true; result = append(result, value) }
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Stage+result[i].Step+result[i].Reason < result[j].Stage+result[j].Step+result[j].Reason })
	return result
}

