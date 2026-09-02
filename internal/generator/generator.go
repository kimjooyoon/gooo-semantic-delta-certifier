package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Cell struct {
	Ordinal int `json:"ordinal"`
	ID string `json:"id"`
	Domain string `json:"domain"`
	Proof string `json:"proof"`
	Indicator string `json:"indicator"`
	Name string `json:"name"`
}

type CorpusCase struct {
	ID string `json:"id"`
	Expected string `json:"expected"`
	Purpose string `json:"purpose"`
}

type Authority struct {
	Schema string
	Protocol string
	DenominatorID string
	DenominatorTotal int
	StatusValues []string
	StatusPrecedence []string
	ProofChoices []string
	IndicatorClasses []string
	UnknownFields []string
	Cells []Cell
	Corpus []CorpusCase
}

type semanticIR struct {
	Schema string `json:"schema"`
	Source string `json:"source"`
	Protocol string `json:"protocol"`
	DenominatorID string `json:"denominator_id"`
	DenominatorTotal int `json:"denominator_total"`
	StatusValues []string `json:"status_values"`
	StatusPrecedence []string `json:"status_precedence"`
	ProofChoices []string `json:"proof_choices"`
	IndicatorClasses []string `json:"indicator_classes"`
	UnknownFields []string `json:"unknown_fields"`
	Cells []Cell `json:"cells"`
	Corpus []CorpusCase `json:"corpus"`
}

func Generate(sourcePath, outputDir string) error {
	raw, err := os.ReadFile(sourcePath)
	if err != nil { return err }
	authority, err := Parse(string(raw))
	if err != nil { return err }
	if err := os.MkdirAll(outputDir, 0o755); err != nil { return err }
	result := semanticIR{Schema: "gooo/semantic-delta-certifier/semantic-ir/v1", Source: sourcePath, Protocol: authority.Protocol, DenominatorID: authority.DenominatorID, DenominatorTotal: authority.DenominatorTotal, StatusValues: authority.StatusValues, StatusPrecedence: authority.StatusPrecedence, ProofChoices: authority.ProofChoices, IndicatorClasses: authority.IndicatorClasses, UnknownFields: authority.UnknownFields, Cells: authority.Cells, Corpus: authority.Corpus}
	jsonBytes, err := renderJSON(result)
	if err != nil { return err }
	jsonBytes = append(jsonBytes, '\n')
	if err := os.WriteFile(filepath.Join(outputDir, "semantic_ir.json"), jsonBytes, 0o644); err != nil { return err }
	goBytes := []byte(renderGo(authority))
	return os.WriteFile(filepath.Join(outputDir, "semantic.gooo.go"), goBytes, 0o644)
}

func renderJSON(result semanticIR) ([]byte, error) {
	var builder strings.Builder
	builder.WriteString("{\n")
	fmt.Fprintf(&builder, "  \"schema\": %q,\n", result.Schema)
	fmt.Fprintf(&builder, "  \"source\": %q,\n", result.Source)
	fmt.Fprintf(&builder, "  \"protocol\": %q,\n", result.Protocol)
	fmt.Fprintf(&builder, "  \"denominator_id\": %q,\n", result.DenominatorID)
	fmt.Fprintf(&builder, "  \"denominator_total\": %d,\n", result.DenominatorTotal)
	writeJSONStrings(&builder, "status_values", result.StatusValues)
	writeJSONStrings(&builder, "status_precedence", result.StatusPrecedence)
	writeJSONStrings(&builder, "proof_choices", result.ProofChoices)
	writeJSONStrings(&builder, "indicator_classes", result.IndicatorClasses)
	writeJSONStrings(&builder, "unknown_fields", result.UnknownFields)
	builder.WriteString("  \"cells\": [\n")
	for index, cell := range result.Cells {
		value, err := json.Marshal(cell); if err != nil { return nil, err }
		value = []byte(objectWithSpaces(string(value)))
		fmt.Fprintf(&builder, "    %s", value)
		if index+1 < len(result.Cells) { builder.WriteString(",") }
		builder.WriteString("\n")
	}
	builder.WriteString("  ],\n  \"corpus\": [\n")
	for index, item := range result.Corpus {
		value, err := json.Marshal(item); if err != nil { return nil, err }
		value = []byte(objectWithSpaces(string(value)))
		fmt.Fprintf(&builder, "    %s", value)
		if index+1 < len(result.Corpus) { builder.WriteString(",") }
		builder.WriteString("\n")
	}
	builder.WriteString("  ]\n}\n")
	return []byte(builder.String()), nil
}

func objectWithSpaces(value string) string {
	value = strings.ReplaceAll(value, "\":", "\": ")
	return strings.ReplaceAll(value, ",\"", ", \"")
}

func writeJSONStrings(builder *strings.Builder, name string, values []string) {
	fmt.Fprintf(builder, "  %q: [", name)
	for index, value := range values { if index > 0 { builder.WriteString(", ") }; fmt.Fprintf(builder, "%q", value) }
	builder.WriteString("],\n")
}

func Parse(source string) (Authority, error) {
	result := Authority{}
	for lineNumber, rawLine := range strings.Split(source, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") { continue }
		fields := strings.Fields(line)
		if len(fields) == 0 { continue }
		if fields[0] == "@gooo" {
			values := keyValues(fields[1:])
			result.Schema = values["schema"]
			continue
		}
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "cell ") && !strings.HasPrefix(line, "denominator ") && !strings.HasPrefix(line, "corpus ") {
			key, value, ok := strings.Cut(line, "=")
			if !ok { return Authority{}, fmt.Errorf("line %d: invalid declaration", lineNumber+1) }
			switch strings.TrimSpace(key) {
			case "protocol": result.Protocol = strings.TrimSpace(value)
			case "status_values": result.StatusValues = csv(value)
			case "status_precedence": result.StatusPrecedence = csv(value)
			case "proof_choices": result.ProofChoices = csv(value)
			case "indicator_classes": result.IndicatorClasses = csv(value)
			case "unknown_fields": result.UnknownFields = csv(value)
			default: return Authority{}, fmt.Errorf("line %d: unknown declaration %q", lineNumber+1, key)
			}
			continue
		}
		switch fields[0] {
		case "denominator":
			values := keyValues(fields[1:]); result.DenominatorID = values["id"]
			var parseErr error
			result.DenominatorTotal, parseErr = strconv.Atoi(values["total"])
			if parseErr != nil { return Authority{}, fmt.Errorf("line %d: invalid denominator total", lineNumber+1) }
		case "cell":
			values := keyValues(fields[1:]); ordinal, parseErr := strconv.Atoi(values["ordinal"])
			if parseErr != nil { return Authority{}, fmt.Errorf("line %d: invalid cell ordinal", lineNumber+1) }
			result.Cells = append(result.Cells, Cell{Ordinal: ordinal, ID: values["id"], Domain: values["domain"], Proof: values["proof"], Indicator: values["indicator"], Name: values["name"]})
		case "corpus":
			values := keyValues(fields[1:]); result.Corpus = append(result.Corpus, CorpusCase{ID: values["id"], Expected: values["expected"], Purpose: values["purpose"]})
		default: return Authority{}, fmt.Errorf("line %d: unknown declaration %q", lineNumber+1, fields[0])
		}
	}
	if err := validate(result); err != nil { return Authority{}, err }
	return result, nil
}

func keyValues(fields []string) map[string]string {
	result := map[string]string{}
	for _, field := range fields { key, value, ok := strings.Cut(field, "="); if ok { result[key] = value } }
	return result
}

func csv(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts { if strings.TrimSpace(part) != "" { result = append(result, strings.TrimSpace(part)) } }
	return result
}

func validate(authority Authority) error {
	if authority.Schema == "" || authority.Protocol == "" || authority.DenominatorID == "" || authority.DenominatorTotal != 15 { return errors.New("INVALID_AUTHORITY_HEADER") }
	if len(authority.Cells) != authority.DenominatorTotal || len(authority.Corpus) != 3 { return errors.New("FIXED_DENOMINATOR_OR_CORPUS_MISMATCH") }
	ids := map[string]bool{}
	for index, cell := range authority.Cells {
		if cell.Ordinal != index+1 || cell.ID == "" || cell.Domain == "" || cell.Proof == "" || cell.Indicator == "" || cell.Name == "" { return fmt.Errorf("INVALID_CELL_%d", index+1) }
		if ids[cell.ID] { return fmt.Errorf("DUPLICATE_CELL_%s", cell.ID) }; ids[cell.ID] = true
		if !contains(authority.ProofChoices, cell.Proof) || !contains(authority.IndicatorClasses, cell.Indicator) { return fmt.Errorf("INVALID_CELL_VECTOR_%d", index+1) }
	}
	if !equal(authority.StatusPrecedence, []string{"REFUTED", "UNKNOWN", "CLOSED"}) || len(authority.StatusValues) != 6 || len(authority.UnknownFields) != 6 { return errors.New("INVALID_STATUS_OR_UNKNOWN_CONTRACT") }
	for _, expected := range []string{"normal", "unknown", "refuted"} { found := false; for _, item := range authority.Corpus { if item.ID == expected { found = true } }; if !found { return errors.New("MISSING_CORPUS_CLASS") } }
	return nil
}

func contains(values []string, target string) bool { for _, value := range values { if value == target { return true } }; return false }

func equal(left, right []string) bool { if len(left) != len(right) { return false }; for index := range left { if left[index] != right[index] { return false } }; return true }

func renderGo(authority Authority) string {
	var builder strings.Builder
	builder.WriteString("// Code generated from .gooo/semantic-delta-certifier.gooo. DO NOT EDIT.\npackage generated\n\n")
	builder.WriteString("type Cell struct {\n\tOrdinal int\n\tID string\n\tDomain string\n\tProof string\n\tIndicator string\n\tName string\n}\n\n")
	builder.WriteString("type CorpusCase struct {\n\tID string\n\tExpected string\n\tPurpose string\n}\n\n")
	fmt.Fprintf(&builder, "const (\n\tProtocolSchema = %q\n\tIRSchema = %q\n\tDenominatorID = %q\n)\n\n", authority.Protocol, "gooo/semantic-delta-certifier/semantic-ir/v1", authority.DenominatorID)
	writeStringSlice(&builder, "StatusValues", authority.StatusValues)
	writeStringSlice(&builder, "StatusPrecedence", authority.StatusPrecedence)
	writeStringSlice(&builder, "ProofChoices", authority.ProofChoices)
	writeStringSlice(&builder, "IndicatorClasses", authority.IndicatorClasses)
	writeStringSlice(&builder, "UnknownFields", authority.UnknownFields)
	builder.WriteString("var Cells = []Cell{\n")
	for _, cell := range authority.Cells { fmt.Fprintf(&builder, "\t{Ordinal: %d, ID: %q, Domain: %q, Proof: %q, Indicator: %q, Name: %q},\n", cell.Ordinal, cell.ID, cell.Domain, cell.Proof, cell.Indicator, cell.Name) }
	builder.WriteString("}\n\nvar Corpus = []CorpusCase{\n")
	for _, item := range authority.Corpus { fmt.Fprintf(&builder, "\t{ID: %q, Expected: %q, Purpose: %q},\n", item.ID, item.Expected, item.Purpose) }
	builder.WriteString("}\n")
	return builder.String()
}

func writeStringSlice(builder *strings.Builder, name string, values []string) {
	fmt.Fprintf(builder, "var %s = []string{", name)
	for index, value := range values { if index > 0 { builder.WriteString(", ") }; fmt.Fprintf(builder, "%q", value) }
	builder.WriteString("}\n\n")
}

func SortedCells(cells []Cell) []Cell { result := append([]Cell{}, cells...); sort.Slice(result, func(i, j int) bool { return result[i].Ordinal < result[j].Ordinal }); return result }
