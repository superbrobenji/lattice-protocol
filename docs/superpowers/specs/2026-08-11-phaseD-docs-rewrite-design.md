# Phase D — Docs Rewrite Design (lattice-protocol)

**Status:** Approved, ready for writing-plans.
**Date:** 2026-08-11
**Parent:** `lattice-nodes` umbrella spec, Phase D — cross-repo by design. This is the
`lattice-protocol` leg, running in parallel with the `lattice-hub` leg. `lattice-nodes`' own
leg is already merged (PR #102).

## Problem

Research (this session, grep-verified against current source at code-review rigor) found
`lattice-protocol`'s core data — opcode/adapter-type/message-type values, the versioning table,
the codegen mechanism — has **zero drift**: every value in Go source, generated C headers, and
tests is in lockstep, and the README's versioning table has been updated correctly at every one
of 5 independent tag releases. This is not a "docs are confidently wrong" situation like
`lattice-nodes` had.

What's actually wrong is narrower:

1. **A stale rename**: `README.md`, `CONTRIBUTING.md`, and `.github/ISSUE_TEMPLATE/bug_report.md`
   all still say the Go consumer is `motionSensorServer` — renamed to `lattice-hub` on 2026-07-01
   (commit `fc7bad9`), in the very PR that was supposed to be the rename. The rename scoped itself
   to internal branding (header guards, go.mod path) and missed the prose.
2. **Structural gaps in existing docs**: README's "Packages" table omits `message/` and `proto/`
   entirely, despite `message/` being the actual wire-format source of truth and the subject of 5
   of the last 8 feature commits. `cmd/gen-headers/`'s description omits that it also generates
   `proto/mesh.proto`, not just `c/`. CONTRIBUTING has no workflow for changing the wire format
   (only opcode/adapter-type addition), despite that being the highest-consequence, most-improvised
   change type in the repo's history. CONTRIBUTING's CI job list is missing `go-lint`. CONTRIBUTING's
   semver policy ("opcode/adapter addition = patch") contradicts the one precedent that exists
   (`v0.3.0`, an opcode-only release, bumped minor).
3. **Missing docs entirely**: no doc explains the 3-repo ecosystem relationship (submodule vs.
   go-import consumption, release-coordination, the "flag-day, no backcompat" convention that's
   implicit in changelog prose but never stated as policy). No doc walks the codegen flow
   (struct tag → `go generate` → `c/`+`proto/` → `static_assert` → CI enforcement) — including that
   `proto/mesh.options` is hand-maintained, not generated, and has no CI drift-detection. No single
   doc consolidates all message types/opcodes/adapter types/wire layout in one place (today spread
   across 3 Go files) — the equivalent of what `lattice-nodes` just built as `server_requirements.md`.

User also requested two non-technical step-by-step guides, since this repo currently assumes deep
Go/codegen fluency in every reader: one for someone who just consumes a release (pins/bumps a
version, zero interest in codegen internals), one for someone making a protocol change who isn't
fluent in this repo's specific codegen mechanism.

## Scope — 9 documents (3 fixes + 6 new)

| Doc | Action | Key content |
|---|---|---|
| `README.md` | Fix | `motionSensorServer`→`lattice-hub` (both occurrences); add `message/`+`proto/` rows to the Packages table; fix `cmd/gen-headers/` description to mention it also writes `proto/mesh.proto`; add a brief, accurate ecosystem paragraph (lattice-nodes vendors `c/` as a submodule, lattice-hub imports the Go module) |
| `CONTRIBUTING.md` | Fix + extend | Fix consumer name; add `go-lint` to the CI job table; reconcile the semver table against the `v0.3.0` precedent (correct the rule or document it as a stated exception — verify which by checking whether `v0.3.0` predates the written rule); add a new "Changing the wire format" workflow section: editing `message/message.go`, `WireSize`/`static_assert` implications, `proto/mesh.options` manual sync (flagged explicitly as hand-maintained with no CI enforcement), cross-repo `PROTO_VERSION` coordination with lattice-nodes/lattice-hub |
| `.github/ISSUE_TEMPLATE/bug_report.md` | Fix | `motionSensorServer`→`lattice-hub` |
| `docs/ecosystem.md` | New | The 3-repo relationship: dependency direction (lattice-protocol is upstream of both), lattice-nodes vendors `c/` via git submodule, lattice-hub imports the Go module via `go.mod`; the "flag-day, no backcompat" convention stated explicitly as policy (currently only implicit in changelog prose); what it means operationally for each consumer to pick up a new tag |
| `docs/codegen_flow.md` | New | Full pipeline walkthrough with a diagram: struct tag (`c:`/`proto:`) on `message/message.go`, `message/types.go`, `opcodes/opcodes.go`, `adapter/types.go` → `go generate` → `cmd/gen-headers/main.go` reflects and writes `c/*.h` + `proto/mesh.proto` → `static_assert(sizeof(mesh_message_t) == WireSize)` catches size drift → CI's `header-sync` job (`make check` = `go generate ./... && git diff --exit-code c/ proto/`) enforces sync. Explicit callout: `proto/mesh.options` is hand-maintained (no "DO NOT EDIT" banner, not reflected over), and CI's diff-check does not protect it from drifting out of sync with new `bytes` fields — a real gap to document as a known caveat, not silently paper over |
| `docs/protocol_reference.md` | New | Consolidated reference: all 6 message types, all 11 opcodes across 5 groups (health, route-report, management, LED/relay, ack) with hex values, all 5 adapter types, the current `MeshMessage` wire layout (200 bytes, protocol v5) — cross-checked against `c/mesh_message.h`'s `static_assert` and `message/message.go` directly, not copied from README's prose |
| `docs/consuming_a_release.md` | New | Step-by-step, zero codegen knowledge assumed. Two paths: **lattice-nodes maintainer** (submodule commands to pin/update to a specific tag, verifying the vendored `c/` headers match the tag, what a `WireSize` bump means for firmware); **lattice-hub maintainer** (`go.mod` version bump, `go mod tidy`, what changed to expect from the version's changelog entry, verifying `go build` catches any breaking Go-level API change) |
| `docs/making_a_protocol_change.md` | New | Step-by-step for a developer not fluent in this repo's codegen: deciding where a change belongs (`message/`, `opcodes/`, or `adapter/`), editing the right file, running `go generate ./...`, reading the generated diff in `c/`+`proto/`, handling a `static_assert` failure, updating `proto/mesh.options` by hand if a `bytes` field changed size, deciding the semver bump per the (now-reconciled) policy. Ends with a pointer to `docs/release_process.md` for the actual tagging/publishing step rather than duplicating it |
| `docs/release_process.md` | New | The **actual current lightweight process**, verified against real repo state (9 git tags `v0.1.0`–`v0.6.0`, no GitHub Releases created, no `CHANGELOG.md` — the README versioning table *is* the changelog): update the README versioning table entry, commit, `git tag vX.Y.Z`, `git push --tags`; the semver decision (cross-ref the reconciled CONTRIBUTING policy); the flag-day coordination handoff to lattice-nodes (submodule bump) and lattice-hub (go.mod bump) maintainers, cross-referencing `docs/ecosystem.md` and `docs/consuming_a_release.md`. Explicitly note GitHub Releases/CHANGELOG.md are *not* currently used, rather than describing a heavier process that doesn't exist |

## Approach

**No separate research phase** — this session's Explore-agent audit already verified every claim
above against current source (git-tag checkouts for the versioning-table claims, direct reads of
`message/message.go`, `opcodes/opcodes.go`, `adapter/types.go`, `cmd/gen-headers/main.go`,
`c/*.h`, `proto/mesh.proto`, `proto/mesh.options`, `.github/workflows/ci.yml`, and `git log`) at
the same rigor Phase A's audit used. That research is the input to the plan directly — no
survey-agent wave needed before writing tasks.

## Review Standard

Same as `lattice-nodes`' Phase D: no "behavior preservation" bar — the bar is **accuracy against
current source**. Each doc's review must independently re-verify a sample of its concrete claims
(opcode/type values against `c/*.h`, the codegen flow against `cmd/gen-headers/main.go` directly,
the semver-table reconciliation against the actual `v0.3.0` tag date vs. the CONTRIBUTING commit
date) rather than checking prose alone. `docs/consuming_a_release.md` and
`docs/making_a_protocol_change.md` additionally need a "could someone with the stated skill level
actually follow this with zero other context" check, the same bar `lattice-nodes`' getting_started.md
was held to.

## Global Constraints

Docs-only phase: no wire-format changes, no code changes (including the semver-table reconciliation
— if `v0.3.0` predates the written rule, document it as historical exception rather than silently
rewriting history; if it postdates the rule, fix the table to match the actual practiced policy).
Every claim about a value (opcode, size, port, field) must be independently verified against source
in this phase, not carried over from the research summary without a spot-check, since the research
summary itself could contain a transcription error.

## Deliverable note

No new GitHub issues needed from this phase — no deferred-future-work item was identified for
lattice-protocol (contrast with lattice-nodes' pre-built-binary tracking issue).
