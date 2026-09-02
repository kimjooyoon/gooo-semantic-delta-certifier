package delta

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/gooo-semantic-delta-certifier/internal/generated"
)

func WriteJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil { return err }
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

func LoadFixture(path string) (Fixture, error) {
	raw, err := os.ReadFile(path)
	if err != nil { return Fixture{}, err }
	var fixture Fixture
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fixture); err != nil { return Fixture{}, err }
	if fixture.Schema != FixtureSchema || fixture.CaseID == "" || fixture.Input.CaseID != fixture.CaseID {
		return Fixture{}, errors.New("INVALID_FIXTURE_CONTRACT")
	}
	return fixture, nil
}

func InventoryForRoot(root string) (Inventory, error) {
	absolute, err := filepath.Abs(root)
	if err != nil { return Inventory{}, err }
	result := Inventory{RootREADMEExcluded: true}
	err = filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil { return walkErr }
		if path == absolute { return nil }
		relative, err := filepath.Rel(absolute, path)
		if err != nil { return err }
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == ".github-cache") {
			return filepath.SkipDir
		}
		if relative == "README.md" { return nil }
		if entry.IsDir() {
			result.DescendantDirs++
			return nil
		}
		if !entry.Type().IsRegular() { return nil }
		result.DescendantFiles++
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension != ".go" && extension != ".gooo" { return nil }
		data, err := os.ReadFile(path)
		if err != nil { return err }
		lines := physicalLines(data)
		if extension == ".go" { result.GoFiles++; result.GoPhysicalLines += lines }
		if extension == ".gooo" { result.GoooFiles++; result.GoooPhysicalLines += lines }
		return nil
	})
	return result, err
}

func physicalLines(data []byte) int {
	if len(data) == 0 { return 0 }
	lines := 0
	for _, value := range data { if value == '\n' { lines++ } }
	if data[len(data)-1] != '\n' { lines++ }
	return lines
}

func EnsureCallerOutput(repoRoot, outputDir string) (string, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil { return "", err }
	output, err := filepath.Abs(outputDir)
	if err != nil { return "", err }
	relative, err := filepath.Rel(root, output)
	if err != nil { return "", err }
	if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))) {
		return "", errors.New("CALLER_OUTPUT_MUST_BE_OUTSIDE_INPUT_REPOSITORY")
	}
	if err := os.MkdirAll(output, 0o755); err != nil { return "", err }
	entries, err := os.ReadDir(output)
	if err != nil { return "", err }
	if len(entries) != 0 { return "", errors.New("CALLER_OUTPUT_MUST_BE_EMPTY") }
	return output, nil
}

func WriteCaseOutput(outputDir string, evidence Evidence) error {
	return WriteJSON(filepath.Join(outputDir, evidence.CaseID+".delta.json"), evidence)
}

func BuildIndex(evidences []Evidence, inventory Inventory, measurements Measurements) ConformanceIndex {
	sort.Slice(evidences, func(i, j int) bool { return evidences[i].CaseID < evidences[j].CaseID })
	cases := make([]CaseSummary, 0, len(evidences))
	for _, evidence := range evidences {
		class := "normal"
		if evidence.State == StatusUnknown { class = "unknown" }
		if evidence.State == StatusRefuted { class = "refuted" }
		cases = append(cases, CaseSummary{CaseID: evidence.CaseID, Class: class, ExpectedState: "", ObservedState: evidence.State, DeltaCells: len(evidence.DeltaCells), ProofVector: evidence.ProofVector, IndicatorVector: evidence.IndicatorVector})
	}
	return ConformanceIndex{Schema: IndexSchema, DenominatorID: generatedDenominatorID(), DenominatorTotal: len(contractCells()), CorpusTotal: len(cases), Cases: cases, Authority: Authority{RepositoryWrites: 0, CallerOwnedOutput: true}, Inventory: inventory, Measurements: measurements}
}

func generatedDenominatorID() string { return generated.DenominatorID }

func RenderDossier(evidences []Evidence, inventory Inventory, measurements Measurements) string {
	sort.Slice(evidences, func(i, j int) bool { return evidences[i].CaseID < evidences[j].CaseID })
	var builder strings.Builder
	builder.WriteString("# Gooo semantic delta dossier\n\n")
	builder.WriteString("schema: gooo/semantic-delta-certifier/dossier/v1\n\n")
	builder.WriteString("This dossier describes exact before/after observations. It does not judge improvement.\n\n")
	builder.WriteString("## Corpus\n\n")
	builder.WriteString("| case | state | delta cells | unknowns | contradictions |\n| --- | --- | ---: | ---: | ---: |\n")
	for _, evidence := range evidences {
		fmt.Fprintf(&builder, "| %s | %s | %d | %d | %d |\n", evidence.CaseID, evidence.State, len(evidence.DeltaCells), len(evidence.Unknowns), len(evidence.Contradictions))
	}
	builder.WriteString("\n## Per-cell observations\n\n")
	for _, evidence := range evidences {
		fmt.Fprintf(&builder, "### %s — %s\n\n", evidence.CaseID, evidence.State)
		builder.WriteString("| domain | exact identity | status | old digest | new digest | reason |\n| --- | --- | --- | --- | --- | --- |\n")
		for _, cell := range evidence.DeltaCells {
			oldDigest, newDigest := "null", "null"
			if cell.OldDigest != nil { oldDigest = *cell.OldDigest }
			if cell.NewDigest != nil { newDigest = *cell.NewDigest }
			fmt.Fprintf(&builder, "| %s | `%s` | %s | `%s` | `%s` | %s |\n", cell.Domain, cell.Identity, cell.Status, oldDigest, newDigest, cell.Reason)
		}
		builder.WriteString("\n")
	}
	builder.WriteString("## Authority and observations\n\n")
	fmt.Fprintf(&builder, "- denominator: `%s` (%d cells)\n", generatedDenominatorID(), len(contractCells()))
	builder.WriteString("- repository_writes: `0`\n- caller_owned_output: `true`\n")
	fmt.Fprintf(&builder, "- inventory: %d dirs, %d files, %d Go files/%d physical lines, %d Gooo files/%d physical lines\n", inventory.DescendantDirs, inventory.DescendantFiles, inventory.GoFiles, inventory.GoPhysicalLines, inventory.GoooFiles, inventory.GoooPhysicalLines)
	fmt.Fprintf(&builder, "- build/test/wall/peak_rss/cache measurements: `%s` / `%s` / `%s` / `%s` / `%s`\n", measurements.Build.Status, measurements.Test.Status, measurements.Wall.Status, measurements.PeakRSS.Status, measurements.Cache.Status)
	builder.WriteString("\nProof and indicator vectors are preserved independently in the machine-readable evidence.\n")
	return builder.String()
}

func ValidateEvidence(evidence Evidence) error {
	if evidence.DenominatorTotal != len(contractCells()) { return errors.New("DENOMINATOR_TOTAL_MISMATCH") }
	if evidence.Precedence[0] != StatusRefuted || evidence.Precedence[1] != StatusUnknown || evidence.Precedence[2] != StatusClosed { return errors.New("PRECEDENCE_MISMATCH") }
	if evidence.Authority.RepositoryWrites != 0 || !evidence.Authority.CallerOwnedOutput { return errors.New("AUTHORITY_BOUNDARY_MISMATCH") }
	if evidence.State != StatusClosed && evidence.State != StatusUnknown && evidence.State != StatusRefuted { return errors.New("INVALID_RESOLUTION_STATE") }
	for _, unknown := range evidence.Unknowns { if err := validUnknown(&unknown); err != nil { return err } }
	for _, cell := range evidence.DeltaCells {
		if err := ensureStatus(cell.Status); err != nil { return err }
		if cell.Status == StatusUnknown {
			if err := validUnknown(cell.Unknown); err != nil { return err }
		}
		if cell.Status != StatusUnknown && cell.Unknown != nil { return errors.New("NON_UNKNOWN_CELL_HAS_UNKNOWN_RECORD") }
	}
	return nil
}
