# Gooo Semantic Delta Certifier

`gooo-semantic-delta-certifier` is a read-only authority boundary for proving
what changed and what stayed the same between two immutable Gooo releases.
It accepts old/new release bindings and semantic IR, matches only exact stable
identities, and emits per-cell `ADDED`, `REMOVED`, `CHANGED`, `UNCHANGED`,
`UNKNOWN`, or `REFUTED` observations for semantic entities, relations, effects,
capabilities, and generated artifact digests.

The semantic authority is the `.gooo` declaration:

```text
.gooo/semantic-delta-certifier.gooo
  -> internal/generated/semantic_ir.json
  -> internal/generated/semantic.gooo.go
  -> evaluator / generator / runtime
```

The declaration fixes a 15-cell denominator, independent proof choices
(`FOUNDATION`, `COHERENCE`, `REGRESSION`), independent indicator classes
(`DRIVER`, `OUTCOME`, `GUARDRAIL`), the six-field UNKNOWN record, and the
precedence `REFUTED > UNKNOWN > CLOSED`. There is no aggregate score or
percentage. A delta result is descriptive; it does not decide whether a
change is an improvement.

## Usage

The runtime writes only below an absolute, caller-owned output directory that
is outside the input repository:

```text
go run ./cmd/gooo-delta evaluate \
  --input fixtures/cases/normal.json \
  --output-dir /absolute/path/to/empty/output
```

The conformance command consumes the fixed normal/UNKNOWN/REFUTED corpus and
writes per-case evidence, `semantic-delta-evidence.json`,
`conformance-index.json`, and a human-readable `delta-dossier.md` to caller
output. Missing matched before/after integers or exact digests are emitted as
`null` with `UNKNOWN`. `FIXED_POINT` is accepted only when it is explicitly
claimed and exact evidence is present.

The input contract treats release and semantic-IR digests as immutable
content-addressed bindings. No fuzzy matching, source-text inference, or
repository mutation is part of the evaluator. Runtime authority records
`repository_writes=0`, and CI outputs are generated outside the repository.

## CI authority

GitHub Actions is the only validation authority. It regenerates the committed
Go projection, runs Go 1.27 formatting/build/vet/test/conformance checks, and
records exact inventory and runner observations. Root `README.md` is excluded
from inventory. Build, test, wall-time, peak-RSS, and cache measurements are
represented as exact values when observed; otherwise they remain `null` with
an UNKNOWN record. Conformance dossier and machine evidence are uploaded as
CI artifacts and packaged as release assets.
