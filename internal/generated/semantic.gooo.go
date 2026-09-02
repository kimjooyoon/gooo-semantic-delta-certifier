// Code generated from .gooo/semantic-delta-certifier.gooo. DO NOT EDIT.
package generated

type Cell struct {
	Ordinal int
	ID string
	Domain string
	Proof string
	Indicator string
	Name string
}

type CorpusCase struct {
	ID string
	Expected string
	Purpose string
}

const (
	ProtocolSchema = "gooo/semantic-delta-certifier/v1"
	IRSchema = "gooo/semantic-delta-certifier/semantic-ir/v1"
	DenominatorID = "semantic-delta-v1"
)

var StatusValues = []string{"ADDED", "REMOVED", "CHANGED", "UNCHANGED", "UNKNOWN", "REFUTED"}

var StatusPrecedence = []string{"REFUTED", "UNKNOWN", "CLOSED"}

var ProofChoices = []string{"FOUNDATION", "COHERENCE", "REGRESSION"}

var IndicatorClasses = []string{"DRIVER", "OUTCOME", "GUARDRAIL"}

var UnknownFields = []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}

var Cells = []Cell{
	{Ordinal: 1, ID: "gooo://semantic-delta/entity/identity", Domain: "ENTITY", Proof: "FOUNDATION", Indicator: "DRIVER", Name: "EntityIdentity"},
	{Ordinal: 2, ID: "gooo://semantic-delta/entity/digest", Domain: "ENTITY", Proof: "COHERENCE", Indicator: "OUTCOME", Name: "EntityDigest"},
	{Ordinal: 3, ID: "gooo://semantic-delta/entity/release-binding", Domain: "ENTITY", Proof: "REGRESSION", Indicator: "GUARDRAIL", Name: "EntityReleaseBinding"},
	{Ordinal: 4, ID: "gooo://semantic-delta/relation/identity", Domain: "RELATION", Proof: "FOUNDATION", Indicator: "DRIVER", Name: "RelationIdentity"},
	{Ordinal: 5, ID: "gooo://semantic-delta/relation/endpoints", Domain: "RELATION", Proof: "COHERENCE", Indicator: "OUTCOME", Name: "RelationEndpoints"},
	{Ordinal: 6, ID: "gooo://semantic-delta/relation/digest", Domain: "RELATION", Proof: "REGRESSION", Indicator: "GUARDRAIL", Name: "RelationDigest"},
	{Ordinal: 7, ID: "gooo://semantic-delta/effect/identity", Domain: "EFFECT", Proof: "FOUNDATION", Indicator: "DRIVER", Name: "EffectIdentity"},
	{Ordinal: 8, ID: "gooo://semantic-delta/effect/operation", Domain: "EFFECT", Proof: "COHERENCE", Indicator: "OUTCOME", Name: "EffectOperation"},
	{Ordinal: 9, ID: "gooo://semantic-delta/effect/digest", Domain: "EFFECT", Proof: "REGRESSION", Indicator: "GUARDRAIL", Name: "EffectDigest"},
	{Ordinal: 10, ID: "gooo://semantic-delta/capability/identity", Domain: "CAPABILITY", Proof: "FOUNDATION", Indicator: "DRIVER", Name: "CapabilityIdentity"},
	{Ordinal: 11, ID: "gooo://semantic-delta/capability/name", Domain: "CAPABILITY", Proof: "COHERENCE", Indicator: "OUTCOME", Name: "CapabilityName"},
	{Ordinal: 12, ID: "gooo://semantic-delta/capability/digest", Domain: "CAPABILITY", Proof: "REGRESSION", Indicator: "GUARDRAIL", Name: "CapabilityDigest"},
	{Ordinal: 13, ID: "gooo://semantic-delta/artifact/identity", Domain: "ARTIFACT", Proof: "FOUNDATION", Indicator: "DRIVER", Name: "ArtifactIdentity"},
	{Ordinal: 14, ID: "gooo://semantic-delta/artifact/digest", Domain: "ARTIFACT", Proof: "COHERENCE", Indicator: "OUTCOME", Name: "ArtifactDigest"},
	{Ordinal: 15, ID: "gooo://semantic-delta/artifact/release-binding", Domain: "ARTIFACT", Proof: "REGRESSION", Indicator: "GUARDRAIL", Name: "ArtifactReleaseBinding"},
}

var Corpus = []CorpusCase{
	{ID: "normal", Expected: "CLOSED", Purpose: "complete_exact_before_after_corpus"},
	{ID: "unknown", Expected: "UNKNOWN", Purpose: "missing_exact_evidence_corpus"},
	{ID: "refuted", Expected: "REFUTED", Purpose: "contradictory_identity_corpus"},
}
