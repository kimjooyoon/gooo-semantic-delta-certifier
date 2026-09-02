# Semantic Delta Certifier Protocol v1

## Authority

`.gooo/semantic-delta-certifier.gooo` is the sole semantic declaration. The
generator projects its fixed denominator, status vocabulary, precedence,
proof choices, indicator classes, UNKNOWN fields, and corpus declaration into
`internal/generated/`. Go code may evaluate, generate, and render these
declarations, but it may not redefine them.

The denominator is exactly 15 cells. Each cell has one proof choice and one
indicator class as independent coordinates. These coordinates are vectors,
not a combined measure.

## Inputs and exact matching

An input contains an immutable `old` and `new` release binding. Each binding
contains a release digest and a semantic-IR digest, plus observed copies of
those digests. The evaluator checks the SHA-256 shape, observed equality, and
the canonical content digest of each semantic IR and release binding.

Semantic records are matched only by their exact stable `id` within their
domain. A record present only on the new side is `ADDED`; only on the old side
is `REMOVED`; equal exact digests are `UNCHANGED`; unequal exact digests are
`CHANGED`. A missing digest is `UNKNOWN` and is never converted into a zero or
an inferred match. Duplicate IDs, invalid digests, kind collisions, digest
collisions, and stale bindings are `REFUTED`.

## Resolution and improvement evidence

The top-level resolution uses `REFUTED > UNKNOWN > CLOSED`. Every UNKNOWN
preserves `stage`, `step`, `reason`, `unknown_class`, `next_operation`, and
`blocked_by`. Improvement is not evaluated as good or bad. Its evidence block
contains `before`, `after`, and exact digest fields; absent matched integer or
digest evidence remains `null` plus `UNKNOWN`. A `FIXED_POINT` claim is
accepted only when `explicit=true` and exact evidence is bound; otherwise the
claim is either UNKNOWN or REFUTED when its explicitness contradicts the
contract.

## Output boundary

The evaluator never writes the input repository. All generated dossier and
machine evidence files are written beneath the caller-owned output directory.
The CI workflow supplies the runner measurements and uploads the resulting
files as an artifact. The tag release workflow packages the same outputs as a
release asset and emits an asset SHA-256 digest.
