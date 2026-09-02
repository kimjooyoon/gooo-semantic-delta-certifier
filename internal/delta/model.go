package delta

import "github.com/kimjooyoon/gooo-semantic-delta-certifier/internal/generated"

const (
	InputSchema  = "gooo/semantic-delta-certifier/input/v1"
	FixtureSchema = "gooo/semantic-delta-certifier/fixture/v1"
	EvidenceSchema = "gooo/semantic-delta-certifier/evidence/v1"
	IndexSchema = "gooo/semantic-delta-certifier/conformance-index/v1"
	DossierSchema = "gooo/semantic-delta-certifier/dossier/v1"
	StatusClosed  = "CLOSED"
	StatusUnknown = "UNKNOWN"
	StatusRefuted = "REFUTED"
	StatusAdded   = "ADDED"
	StatusRemoved = "REMOVED"
	StatusChanged = "CHANGED"
	StatusUnchanged = "UNCHANGED"
)

type Unknown struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type SemanticEntity struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Digest string `json:"digest"`
	Name   string `json:"name"`
}

type SemanticRelation struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	From   string `json:"from"`
	To     string `json:"to"`
	Digest string `json:"digest"`
}

type SemanticEffect struct {
	ID        string `json:"id"`
	Operation string `json:"operation"`
	Digest    string `json:"digest"`
}

type SemanticCapability struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Digest string `json:"digest"`
}

type GeneratedArtifact struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type SemanticIR struct {
	Schema             string              `json:"schema"`
	IRID               string              `json:"ir_id"`
	Digest             string              `json:"digest"`
	ObservedDigest     string              `json:"observed_digest"`
	Entities           []SemanticEntity    `json:"entities"`
	Relations          []SemanticRelation  `json:"relations"`
	Effects            []SemanticEffect    `json:"effects"`
	Capabilities       []SemanticCapability `json:"capabilities"`
	GeneratedArtifacts []GeneratedArtifact  `json:"generated_artifacts"`
}

type ReleaseBinding struct {
	Schema                  string    `json:"schema"`
	ReleaseID               string    `json:"release_id"`
	Tag                     string    `json:"tag"`
	ReleaseDigest           string    `json:"release_digest"`
	ObservedReleaseDigest   string    `json:"observed_release_digest"`
	SemanticIRDigest        string    `json:"semantic_ir_digest"`
	ObservedSemanticIRDigest string   `json:"observed_semantic_ir_digest"`
	SemanticIR              SemanticIR `json:"semantic_ir"`
}

type ImprovementClaim struct {
	ClaimedState   string `json:"claimed_state"`
	Explicit       bool   `json:"explicit"`
	PairID         string `json:"pair_id"`
	Before         *int   `json:"before"`
	After          *int   `json:"after"`
	BeforeIdentity string `json:"before_identity"`
	AfterIdentity  string `json:"after_identity"`
	ExactDigest    string `json:"exact_digest"`
}

type Input struct {
	Schema  string          `json:"schema"`
	CaseID  string          `json:"case_id"`
	Old     ReleaseBinding  `json:"old"`
	New     ReleaseBinding  `json:"new"`
	Improvement *ImprovementClaim `json:"improvement"`
}

type Fixture struct {
	Schema        string `json:"schema"`
	CaseID        string `json:"case_id"`
	Class         string `json:"class"`
	ExpectedState string `json:"expected_state"`
	Input         Input  `json:"input"`
}

type DeltaCell struct {
	Domain   string  `json:"domain"`
	Identity string  `json:"identity"`
	Status   string  `json:"status"`
	OldDigest *string `json:"old_digest"`
	NewDigest *string `json:"new_digest"`
	Reason   string  `json:"reason"`
	Unknown  *Unknown `json:"unknown"`
}

type ProofVectorCell struct {
	CellID string `json:"cell_id"`
	ProofChoice string `json:"proof_choice"`
}

type IndicatorVectorCell struct {
	CellID string `json:"cell_id"`
	Indicator string `json:"indicator"`
}

type ImprovementEvidence struct {
	ClaimedState string `json:"claimed_state"`
	Status       string `json:"status"`
	Before       *int   `json:"before"`
	After        *int   `json:"after"`
	PairID       string `json:"pair_id"`
	BeforeIdentity string `json:"before_identity"`
	AfterIdentity  string `json:"after_identity"`
	ExactDigest  *string `json:"exact_digest"`
	Reason       string `json:"reason"`
	Unknown      *Unknown `json:"unknown"`
}

type Authority struct {
	RepositoryWrites int `json:"repository_writes"`
	CallerOwnedOutput bool `json:"caller_owned_output"`
}

type Inventory struct {
	DescendantDirs int `json:"descendant_dirs"`
	DescendantFiles int `json:"descendant_files"`
	GoFiles int `json:"go_files"`
	GoPhysicalLines int `json:"go_physical_lines"`
	GoooFiles int `json:"gooo_files"`
	GoooPhysicalLines int `json:"gooo_physical_lines"`
	RootREADMEExcluded bool `json:"root_readme_excluded"`
}

type Measurement struct {
	Value *int64 `json:"value"`
	Unit string `json:"unit"`
	Status string `json:"status"`
	Unknown *Unknown `json:"unknown"`
}

type Measurements struct {
	Build Measurement `json:"build"`
	Test Measurement `json:"test"`
	Conformance Measurement `json:"conformance"`
	Wall Measurement `json:"wall"`
	PeakRSS Measurement `json:"peak_rss"`
	Cache Measurement `json:"cache"`
}

type Evidence struct {
	Schema string `json:"schema"`
	CaseID string `json:"case_id"`
	State string `json:"state"`
	Reason string `json:"reason"`
	Precedence []string `json:"precedence"`
	DenominatorID string `json:"denominator_id"`
	DenominatorTotal int `json:"denominator_total"`
	DeltaCells []DeltaCell `json:"delta_cells"`
	ProofVector []ProofVectorCell `json:"proof_vector"`
	IndicatorVector []IndicatorVectorCell `json:"indicator_vector"`
	Improvement ImprovementEvidence `json:"improvement"`
	Unknowns []Unknown `json:"unknowns"`
	Contradictions []string `json:"contradictions"`
	Authority Authority `json:"authority"`
	Inventory Inventory `json:"inventory"`
	Measurements Measurements `json:"measurements"`
}

type CaseSummary struct {
	CaseID string `json:"case_id"`
	Class string `json:"class"`
	ExpectedState string `json:"expected_state"`
	ObservedState string `json:"observed_state"`
	DeltaCells int `json:"delta_cells"`
	ProofVector []ProofVectorCell `json:"proof_vector"`
	IndicatorVector []IndicatorVectorCell `json:"indicator_vector"`
}

type ConformanceIndex struct {
	Schema string `json:"schema"`
	DenominatorID string `json:"denominator_id"`
	DenominatorTotal int `json:"denominator_total"`
	CorpusTotal int `json:"corpus_total"`
	Cases []CaseSummary `json:"cases"`
	Authority Authority `json:"authority"`
	Inventory Inventory `json:"inventory"`
	Measurements Measurements `json:"measurements"`
}

type CorpusEvidence struct {
	Schema string `json:"schema"`
	DenominatorID string `json:"denominator_id"`
	DenominatorTotal int `json:"denominator_total"`
	CorpusTotal int `json:"corpus_total"`
	Cases []Evidence `json:"cases"`
	Authority Authority `json:"authority"`
	Inventory Inventory `json:"inventory"`
	Measurements Measurements `json:"measurements"`
}

type Metrics struct {
	Build Measurement `json:"build"`
	Test Measurement `json:"test"`
	Conformance Measurement `json:"conformance"`
	Wall Measurement `json:"wall"`
	PeakRSS Measurement `json:"peak_rss"`
	Cache Measurement `json:"cache"`
}

func contractCells() []generated.Cell { return generated.Cells }
