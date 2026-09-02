package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/gooo-semantic-delta-certifier/internal/delta"
	"github.com/kimjooyoon/gooo-semantic-delta-certifier/internal/generator"
)

func main() {
	if len(os.Args) < 2 { fatal("command is required") }
	var err error
	switch os.Args[1] {
	case "generate": err = generate(os.Args[2:])
	case "evaluate": err = evaluate(os.Args[2:])
	case "conformance": err = conformance(os.Args[2:])
	case "enrich": err = enrich(os.Args[2:])
	default: err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil { fatal(err.Error()) }
}

func generate(args []string) error {
	flags := flag.NewFlagSet("generate", flag.ContinueOnError)
	source := flags.String("source", ".gooo/semantic-delta-certifier.gooo", "semantic authority source")
	output := flags.String("output", "internal/generated", "generated output directory")
	if err := flags.Parse(args); err != nil { return err }
	return generator.Generate(*source, *output)
}

func evaluate(args []string) error {
	flags := flag.NewFlagSet("evaluate", flag.ContinueOnError)
	inputPath := flags.String("input", "", "input JSON path")
	outputPath := flags.String("output-dir", "", "absolute caller-owned output directory")
	rootPath := flags.String("repository-root", ".", "input repository root")
	if err := flags.Parse(args); err != nil { return err }
	if *inputPath == "" || *outputPath == "" { return errors.New("input and output-dir are required") }
	output, err := delta.EnsureCallerOutput(*rootPath, *outputPath)
	if err != nil { return err }
	inventory, err := delta.InventoryForRoot(*rootPath)
	if err != nil { return err }
	raw, err := os.ReadFile(*inputPath)
	if err != nil { return err }
	evidence := delta.EvaluateBytes(raw, inventory)
	if err := delta.WriteCaseOutput(output, evidence); err != nil { return err }
	if err := delta.WriteJSON(filepath.Join(output, "semantic-delta-evidence.json"), evidence); err != nil { return err }
	if err := delta.WriteJSON(filepath.Join(output, "conformance-index.json"), indexForSingle(evidence, inventory)); err != nil { return err }
	return os.WriteFile(filepath.Join(output, "delta-dossier.md"), []byte(delta.RenderDossier([]delta.Evidence{evidence}, inventory, evidence.Measurements)), 0o644)
}

func conformance(args []string) error {
	flags := flag.NewFlagSet("conformance", flag.ContinueOnError)
	fixturesPath := flags.String("fixtures", "fixtures/cases", "fixture directory")
	outputPath := flags.String("output-dir", "", "absolute caller-owned output directory")
	rootPath := flags.String("repository-root", ".", "input repository root")
	if err := flags.Parse(args); err != nil { return err }
	if *outputPath == "" { return errors.New("output-dir is required") }
	output, err := delta.EnsureCallerOutput(*rootPath, *outputPath)
	if err != nil { return err }
	inventory, err := delta.InventoryForRoot(*rootPath)
	if err != nil { return err }
	paths, err := filepath.Glob(filepath.Join(*fixturesPath, "*.json"))
	if err != nil { return err }
	sort.Strings(paths)
	if len(paths) != 3 { return fmt.Errorf("fixed corpus requires exactly 3 fixtures, found %d", len(paths)) }
	fixtures := make([]delta.Fixture, 0, len(paths))
	evidences := make([]delta.Evidence, 0, len(paths))
	var firstFailure error
	for _, path := range paths {
		fixture, loadErr := delta.LoadFixture(path)
		if loadErr != nil { if firstFailure == nil { firstFailure = fmt.Errorf("%s: %w", path, loadErr) }; continue }
		evidence := delta.Evaluate(fixture.Input, inventory)
		if validateErr := delta.ValidateEvidence(evidence); validateErr != nil && firstFailure == nil { firstFailure = fmt.Errorf("%s: %w", fixture.CaseID, validateErr) }
		if evidence.State != fixture.ExpectedState && firstFailure == nil { firstFailure = fmt.Errorf("%s: expected %s, observed %s", fixture.CaseID, fixture.ExpectedState, evidence.State) }
		fixtures = append(fixtures, fixture)
		evidences = append(evidences, evidence)
		if writeErr := delta.WriteCaseOutput(output, evidence); writeErr != nil && firstFailure == nil { firstFailure = writeErr }
	}
	measurements := unknownMetrics()
	corpus := corpusEvidence(evidences, inventory, measurements)
	if writeErr := delta.WriteJSON(filepath.Join(output, "semantic-delta-evidence.json"), corpus); writeErr != nil && firstFailure == nil { firstFailure = writeErr }
	if writeErr := delta.WriteJSON(filepath.Join(output, "conformance-index.json"), indexForFixtures(fixtures, evidences, inventory, measurements)); writeErr != nil && firstFailure == nil { firstFailure = writeErr }
	if writeErr := os.WriteFile(filepath.Join(output, "delta-dossier.md"), []byte(delta.RenderDossier(evidences, inventory, measurements)), 0o644); writeErr != nil && firstFailure == nil { firstFailure = writeErr }
	return firstFailure
}

func enrich(args []string) error {
	flags := flag.NewFlagSet("enrich", flag.ContinueOnError)
	evidencePath := flags.String("evidence", "", "corpus evidence JSON")
	metricsPath := flags.String("metrics", "", "GitHub Actions metrics JSON")
	outputPath := flags.String("output-dir", "", "caller-owned output directory")
	if err := flags.Parse(args); err != nil { return err }
	if *evidencePath == "" || *metricsPath == "" || *outputPath == "" { return errors.New("evidence, metrics, and output-dir are required") }
	evidenceRaw, err := os.ReadFile(*evidencePath)
	if err != nil { return err }
	metricsRaw, err := os.ReadFile(*metricsPath)
	if err != nil { return err }
	var corpus delta.CorpusEvidence
	var metrics delta.Metrics
	if err := json.Unmarshal(evidenceRaw, &corpus); err != nil { return err }
	if err := json.Unmarshal(metricsRaw, &metrics); err != nil { return err }
	corpus.Measurements = delta.Measurements(metrics)
	for index := range corpus.Cases { corpus.Cases[index].Measurements = corpus.Measurements }
	if err := delta.WriteJSON(filepath.Join(*outputPath, "semantic-delta-evidence.json"), corpus); err != nil { return err }
	for _, evidence := range corpus.Cases { evidence.Measurements = corpus.Measurements; if err := delta.WriteCaseOutput(*outputPath, evidence); err != nil { return err } }
	fixtures := make([]delta.Fixture, 0, len(corpus.Cases))
	for _, evidence := range corpus.Cases { fixtures = append(fixtures, delta.Fixture{CaseID: evidence.CaseID, ExpectedState: evidence.State, Class: classFor(evidence.State), Input: delta.Input{CaseID: evidence.CaseID}}) }
	if err := delta.WriteJSON(filepath.Join(*outputPath, "conformance-index.json"), indexForFixtures(fixtures, corpus.Cases, corpus.Inventory, corpus.Measurements)); err != nil { return err }
	return os.WriteFile(filepath.Join(*outputPath, "delta-dossier.md"), []byte(delta.RenderDossier(corpus.Cases, corpus.Inventory, corpus.Measurements)), 0o644)
}

func corpusEvidence(evidences []delta.Evidence, inventory delta.Inventory, measurements delta.Measurements) delta.CorpusEvidence {
	sort.Slice(evidences, func(i, j int) bool { return evidences[i].CaseID < evidences[j].CaseID })
	return delta.CorpusEvidence{Schema: delta.EvidenceSchema, DenominatorID: "semantic-delta-v1", DenominatorTotal: 15, CorpusTotal: len(evidences), Cases: evidences, Authority: delta.Authority{RepositoryWrites: 0, CallerOwnedOutput: true}, Inventory: inventory, Measurements: measurements}
}

func indexForFixtures(fixtures []delta.Fixture, evidences []delta.Evidence, inventory delta.Inventory, measurements delta.Measurements) delta.ConformanceIndex {
	byID := map[string]delta.Fixture{}
	for _, fixture := range fixtures { byID[fixture.CaseID] = fixture }
	sort.Slice(evidences, func(i, j int) bool { return evidences[i].CaseID < evidences[j].CaseID })
	cases := make([]delta.CaseSummary, 0, len(evidences))
	for _, evidence := range evidences {
		fixture := byID[evidence.CaseID]
		cases = append(cases, delta.CaseSummary{CaseID: evidence.CaseID, Class: fixture.Class, ExpectedState: fixture.ExpectedState, ObservedState: evidence.State, DeltaCells: len(evidence.DeltaCells), ProofVector: evidence.ProofVector, IndicatorVector: evidence.IndicatorVector})
	}
	return delta.ConformanceIndex{Schema: delta.IndexSchema, DenominatorID: "semantic-delta-v1", DenominatorTotal: 15, CorpusTotal: len(cases), Cases: cases, Authority: delta.Authority{RepositoryWrites: 0, CallerOwnedOutput: true}, Inventory: inventory, Measurements: measurements}
}

func indexForSingle(evidence delta.Evidence, inventory delta.Inventory) delta.ConformanceIndex {
	return delta.ConformanceIndex{Schema: delta.IndexSchema, DenominatorID: "semantic-delta-v1", DenominatorTotal: 15, CorpusTotal: 1, Cases: []delta.CaseSummary{{CaseID: evidence.CaseID, Class: classFor(evidence.State), ExpectedState: "", ObservedState: evidence.State, DeltaCells: len(evidence.DeltaCells), ProofVector: evidence.ProofVector, IndicatorVector: evidence.IndicatorVector}}, Authority: delta.Authority{RepositoryWrites: 0, CallerOwnedOutput: true}, Inventory: inventory, Measurements: evidence.Measurements}
}

func classFor(state string) string { if state == delta.StatusUnknown { return "unknown" }; if state == delta.StatusRefuted { return "refuted" }; return "normal" }

func unknownMetrics() delta.Measurements {
	return delta.Measurements{
		Build: measurement("CI_BUILD", "BUILD_MEASUREMENT_UNAVAILABLE"), Test: measurement("CI_TEST", "TEST_MEASUREMENT_UNAVAILABLE"), Conformance: measurement("CI_CONFORMANCE", "CONFORMANCE_MEASUREMENT_UNAVAILABLE"), Wall: measurement("CI_RUNTIME", "WALL_TIME_UNAVAILABLE"), PeakRSS: measurement("CI_RUNTIME", "PEAK_RSS_UNAVAILABLE"), Cache: measurement("CI_CACHE", "CACHE_OBSERVATION_UNAVAILABLE"),
	}
}

func measurement(stage, reason string) delta.Measurement {
	return delta.Measurement{Value: nil, Unit: "", Status: delta.StatusUnknown, Unknown: &delta.Unknown{Stage: stage, Step: "OBSERVE_CI_MEASUREMENT", Reason: reason, UnknownClass: "ENVIRONMENT_UNAVAILABLE", NextOperation: "PROVIDE_GITHUB_ACTIONS_MEASUREMENT", BlockedBy: []string{"github-actions-run"}}}
}

func fatal(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
