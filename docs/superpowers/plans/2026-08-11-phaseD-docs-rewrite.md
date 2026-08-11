# Phase D Docs Rewrite (lattice-protocol) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the 3 docs with real drift (a stale rename, structural gaps) and write 7 new docs
so lattice-protocol's documentation covers the ecosystem relationship, the codegen flow, a
consolidated protocol reference, release consumption/making-a-change/release-process guides, all
verified against current source.

**Architecture:** No code changes anywhere in this plan — every task edits or creates only
Markdown/template files. Each task is independently verifiable against real source (Go files,
generated headers, git tags, CI config) already surveyed in the design spec.

**Tech Stack:** Markdown docs, Mermaid for the codegen-flow diagram.

## Global Constraints

- Docs-only phase: no wire-format changes, no code changes, no changes to `message/`, `opcodes/`,
  `adapter/`, `cmd/gen-headers/`, `c/`, or `proto/` source files.
- Every concrete claim (opcode value, field name, path, tag date, CI job name) must be
  independently verified against current source by the task's implementer — do not copy a claim
  from the design spec without spot-checking it against the file it cites, since the spec's
  research summary itself could contain a transcription error.
- The semver-table reconciliation (Task 2) must determine, by checking `git log -1 --format=%ad
  <sha-of-v0.3.0-tag>` against `git log --follow -p -- CONTRIBUTING.md` (or equivalent), whether
  the `v0.3.0` minor-bump precedent predates or postdates the written "opcode addition = patch"
  rule, and document accordingly — do not guess.
- Cross-references between new docs (e.g. `making_a_protocol_change.md` pointing to
  `release_process.md`) use relative Markdown links: `[release process](release_process.md)`.
- Every doc lives under `docs/` except `README.md`, `CONTRIBUTING.md`, and
  `.github/ISSUE_TEMPLATE/bug_report.md`, which are edited in place at repo root.

---

### Task 1: Fix `README.md`

**Files:**
- Modify: `README.md`

**Interfaces:**
- Produces: the corrected README other tasks' cross-references assume exists (e.g. Task 4's
  ecosystem doc may be linked from here).

- [ ] **Step 1: Fix the stale consumer name**

Grep `README.md` for `motionSensorServer` (2 occurrences per the spec — verify the count is still
2 before editing, source drift is possible). Replace each with `lattice-hub`, keeping the
surrounding sentence grammatical (check whether it needs a link, matching how other repo names in
the README are formatted).

- [ ] **Step 2: Add `message/` and `proto/` to the Packages table**

Read `message/message.go` and `message/types.go` to write an accurate one-line description of
`message/` (the `MeshMessage` wire-format struct — the actual protocol frame definition — plus
message-type constants). Read `proto/mesh.proto` and `proto/mesh.options` to write an accurate
one-line description of `proto/` (generated `.proto` + hand-maintained `mesh.options` sizing
file). Add both as new rows in the existing Packages table, matching its current column format.

- [ ] **Step 3: Fix the `cmd/gen-headers/` description**

Read `cmd/gen-headers/main.go` and confirm it writes `proto/mesh.proto` in addition to `c/*.h`
(the spec cites a `writeMeshProto` function — verify this function or its current equivalent
still exists and is called). Update the README's description of `cmd/gen-headers/` to mention
both outputs.

- [ ] **Step 4: Add a brief ecosystem paragraph**

Add 2-4 sentences (near the top, e.g. after the intro or in a new "Ecosystem" subsection) stating:
this repo is the shared wire-format source of truth; `lattice-nodes` vendors the generated `c/`
headers as a git submodule; `lattice-hub` imports the Go module via `go.mod`. Verify the
submodule-vendoring claim by checking `lattice-nodes`' `.gitmodules` or `firmware/main/lib/`
structure if accessible; if not directly verifiable from this repo, phrase it as consistent with
what `docs/ecosystem.md` (Task 4) will state in full, and cross-link to it once Task 4 exists
(the doc doesn't need to exist yet to write this paragraph — a relative link `[ecosystem
doc](docs/ecosystem.md)` is valid Markdown regardless of task order).

- [ ] **Step 5: Self-review and commit**

Re-read the full README for internal consistency (no leftover reference to the old name, no
broken table formatting). Commit:

```bash
git add README.md
git commit -m "docs: fix stale consumer name and package table gaps in README"
```

---

### Task 2: Fix `CONTRIBUTING.md` and `.github/ISSUE_TEMPLATE/bug_report.md`

**Files:**
- Modify: `CONTRIBUTING.md`
- Modify: `.github/ISSUE_TEMPLATE/bug_report.md`

**Interfaces:**
- Produces: the "Changing the wire format" workflow section other docs (`making_a_protocol_change.md`,
  Task 7) may reference or build on.

- [ ] **Step 1: Fix the stale consumer name in both files**

`CONTRIBUTING.md` line ~11 and `.github/ISSUE_TEMPLATE/bug_report.md` line ~26 (per the spec's
research — verify current line numbers, they may have shifted) both say `motionSensorServer`.
Replace with `lattice-hub` in both.

- [ ] **Step 2: Add `go-lint` to the CI job table**

Read `.github/workflows/ci.yml` and list its actual current job names (the spec found `go-test`,
`header-sync`, `go-lint`, plus CodeQL in a separate workflow — verify this is still accurate).
Add any missing job to `CONTRIBUTING.md`'s CI job table, matching its existing format.

- [ ] **Step 3: Reconcile the semver policy table**

Run `git log --format='%ai' -1 v0.3.0` and `git log --follow -p -- CONTRIBUTING.md` (or `git log
-p --follow CONTRIBUTING.md | grep -B5 -A5 "opcode or adapter type"`) to determine whether the
`v0.3.0` minor-bump precedent predates or postdates when the "opcode/adapter addition = Patch"
rule was written. If the rule postdates `v0.3.0`: leave the rule as the current policy but add a
one-line historical note ("v0.3.0 predates this policy and used a minor bump"). If the rule
predates `v0.3.0`: the precedent broke the rule after it was written — flag this directly, and
either correct the table to say "Minor" (if that's the actual practiced convention) or note it as
a one-off exception with rationale if inferable from the `v0.3.0` commit message. Do not silently
pick one without checking the actual dates.

- [ ] **Step 4: Add the "Changing the wire format" workflow section**

Following the format of the existing "Adding an opcode" / "Adding an adapter type" sections in
`CONTRIBUTING.md`, add a new section covering: editing `message/message.go` (adding/changing a
field), the `WireSize` constant and what happens if the struct's packed size doesn't match it
(cite the actual `static_assert` mechanism — read `c/mesh_message.h` to quote it precisely),
manually updating `proto/mesh.options` when a `bytes` field's size changes (explicitly note: this
file has no "DO NOT EDIT" banner, is NOT regenerated, and CI's `header-sync` job does not detect
if it drifts out of sync — this is a real known gap, state it as such, not as a solved problem),
and the cross-repo coordination step (point to `docs/release_process.md`, Task 8, for the actual
tagging mechanics — this section covers what to change, not how to publish it).

- [ ] **Step 5: Self-review and commit**

```bash
git add CONTRIBUTING.md .github/ISSUE_TEMPLATE/bug_report.md
git commit -m "docs: fix CONTRIBUTING gaps, add wire-format-change workflow"
```

---

### Task 3: New `docs/ecosystem.md`

**Files:**
- Create: `docs/ecosystem.md`

**Interfaces:**
- Consumes: none.
- Produces: the doc that `docs/release_process.md` (Task 8) and README.md (Task 1) may link to.

- [ ] **Step 1: Write the 3-repo relationship section**

State clearly: `lattice-protocol` is upstream of both `lattice-nodes` and `lattice-hub`. Verify
and describe the two consumption mechanisms precisely — `lattice-nodes` vendors the generated
`c/*.h` headers as a git submodule (describe what pinning a submodule to a tag means in practice);
`lattice-hub` imports this repo as a Go module via `go.mod` (`require
github.com/superbrobenji/lattice-protocol vX.Y.Z`). If you have read access to either sibling
repo's directory on this machine (check `/Users/benji/projects/personal/lattice-nodes` and
`/Users/benji/projects/personal/lattice-hub` — they are sibling directories), verify the submodule
path and the go.mod require line directly rather than describing them from memory.

- [ ] **Step 2: State the flag-day convention as explicit policy**

The spec found this convention is currently only implicit in changelog prose ("Flag-day, no
backcompat" appears in commit messages/README's versioning table). Write it out as an explicit,
named policy: what "flag-day" means operationally (a new `PROTO_VERSION`/`ProtoVersion` value
means old firmware/servers on the previous version are incompatible with new ones — no
dual-version support period), and why this repo chooses that model over versioned compatibility
shims (infer the rationale from context — small embedded devices, a single-deployment system — or
state it as an observed convention without inventing a rationale you can't source).

- [ ] **Step 3: Describe what a new tag means operationally for each consumer**

Two short subsections: "If you maintain lattice-nodes" (a new tag means re-running the submodule
update, re-vendoring headers, verifying `WireSize`/`static_assert` alignment, and — critically —
a full reflash cycle before any node can talk to a hub running the new protocol version).
"If you maintain lattice-hub" (a new tag means a `go.mod` bump, `go mod tidy`, rebuilding, and
the same flag-day caveat — old firmware in the field will be dropped by the `ProtoVersion` check
until reflashed).

- [ ] **Step 4: Self-review and commit**

```bash
git add docs/ecosystem.md
git commit -m "docs: add 3-repo ecosystem relationship doc"
```

---

### Task 4: New `docs/codegen_flow.md`

**Files:**
- Create: `docs/codegen_flow.md`

**Interfaces:**
- Consumes: none.
- Produces: the doc `docs/making_a_protocol_change.md` (Task 7) points to for the mechanical
  details of running codegen.

- [ ] **Step 1: Read the actual codegen source before writing anything**

Read `cmd/gen-headers/main.go` in full. Confirm: it reflects over `message.MeshMessage` (and the
constant packages `message/types.go`, `opcodes/opcodes.go`, `adapter/types.go`) using struct tags
(`c:`, `proto:`), and writes every file in `c/` plus `proto/mesh.proto`. Note the exact function
names involved if the file has clearly separated write-functions (the spec mentions
`writeMeshProto` — verify this name is still accurate, function names can drift).

- [ ] **Step 2: Write the pipeline walkthrough with a Mermaid diagram**

Structure: struct tag on a Go field/constant → `go generate ./...` (confirm this exact invocation
by reading the `//go:generate` directive comments in `message/types.go` and `opcodes/opcodes.go`)
→ `cmd/gen-headers/main.go` reflects and writes `c/*.h` + `proto/mesh.proto` → the generated
`static_assert(sizeof(mesh_message_t) == WireSize, ...)` in `c/mesh_message.h` catches any
size-drift bug at C-compile time → CI's `header-sync` job runs `make check`, which the spec
describes as `go generate ./... && git diff --exit-code c/ proto/` (verify this is the actual
`make check` target by reading the `Makefile`) to enforce that generated output is never
hand-edited or stale. Render this as a Mermaid flowchart (`graph TD` or `graph LR`), since GitHub
renders Mermaid natively in `.md` files — no separate diagram tool needed.

- [ ] **Step 3: Add the `proto/mesh.options` caveat explicitly**

Write a clearly-labeled "Known gap" or "Manual step" callout: `proto/mesh.options` has no "DO NOT
EDIT" banner (verify by reading the file's header), is not written by `cmd/gen-headers/main.go`
(verify by grepping `main.go` for `mesh.options` — confirm it's absent from the generator's write
list), and is not covered by CI's diff-check in any way that would catch a new `bytes` field
whose `max_size` wasn't manually added. State this as a real, currently-unmitigated risk, not as
something CI already handles.

- [ ] **Step 4: Self-review and commit**

Verify the Mermaid diagram is syntactically valid (check bracket/arrow syntax against Mermaid's
flowchart grammar; if `mmdc` is available locally, render it to confirm — not required, but do a
careful manual syntax check at minimum).

```bash
git add docs/codegen_flow.md
git commit -m "docs: add codegen pipeline flow doc with mesh.options caveat"
```

---

### Task 5: New `docs/protocol_reference.md`

**Files:**
- Create: `docs/protocol_reference.md`

**Interfaces:**
- Consumes: none.
- Produces: none (a leaf reference doc).

- [ ] **Step 1: Extract every value from source directly, not from the design spec's prose**

Read `message/types.go` for all message-type constants (spec says 6: verify by counting). Read
`opcodes/opcodes.go` for all opcode constants (spec says 11 across 5 groups: health, route-report,
management, LED/relay, ack — verify names, hex values, and grouping directly from the source
comments/grouping, not from the spec's summary). Read `adapter/types.go` for all adapter-type
constants (spec says 5: Unknown/Serial/PIR/LED/Relay — verify int32 values). Read
`message/message.go` for the full `MeshMessage` struct field-by-field (field name, type, byte
size, the `c:`/`proto:` tags, any doc comment explaining the field's purpose) and read
`c/mesh_message.h`'s `static_assert` to confirm the total wire size (spec says 200 bytes,
protocol v5 — verify against the actual current tag/constant, since a new tag could have shipped
since the spec was written).

- [ ] **Step 2: Write the consolidated reference**

Four sections: Message Types (table: name, numeric value, one-line meaning), Opcodes (table
grouped by the 5 groups: name, hex value, one-line meaning), Adapter Types (table: name, numeric
value, one-line meaning), Wire Layout (a field-by-field table of `MeshMessage`: field name, byte
offset if determinable, size, purpose — plus the total `WireSize` and protocol version number
prominently stated at the top of this section).

- [ ] **Step 3: Self-review and commit**

Cross-check every number written against the source file it came from one more time before
committing — this doc's entire value is being more trustworthy than reading 3 scattered Go files,
so a transcription error here is worse than not having the doc at all.

```bash
git add docs/protocol_reference.md
git commit -m "docs: add consolidated protocol reference (message types, opcodes, adapter types, wire layout)"
```

---

### Task 6: New `docs/consuming_a_release.md`

**Files:**
- Create: `docs/consuming_a_release.md`

**Interfaces:**
- Consumes: none (may link to `docs/ecosystem.md` and `docs/protocol_reference.md` once they
  exist — safe to link regardless of task execution order).
- Produces: none.

- [ ] **Step 1: Write the "I maintain lattice-nodes" path**

Zero codegen knowledge assumed. Step-by-step: locate the submodule reference in `lattice-nodes`
(describe checking `.gitmodules` and the vendored path — if you have read access to the sibling
`lattice-nodes` directory on this machine, verify the exact path and current pinned commit/tag
directly rather than describing generically), the exact `git submodule update --remote` or
`git submodule update --init` + manual checkout-to-tag commands to pin to a specific
`lattice-protocol` release tag, how to verify the vendored `c/` headers now match that tag
(`git -C <submodule-path> describe --tags`), and what a `WireSize` bump means for firmware (a
build-time `static_assert` failure if the firmware's own expectations are out of sync — point to
`lattice-nodes`' own `docs/memory_usage.md` or build docs if relevant, but don't fabricate a cross-repo
link you haven't verified exists).

- [ ] **Step 2: Write the "I maintain lattice-hub" path**

Step-by-step: the exact `go get github.com/superbrobenji/lattice-protocol@vX.Y.Z` command, running
`go mod tidy`, reading the target version's changelog entry in this repo's own README versioning
table to know what changed, and running `go build ./...` to let the Go compiler surface any
breaking API change immediately (struct field renames, removed constants) as compile errors.

- [ ] **Step 3: Self-review and commit**

Verify every command shown is syntactically correct Go/git tooling (these are being read by
someone who explicitly doesn't want to debug a typo'd command).

```bash
git add docs/consuming_a_release.md
git commit -m "docs: add step-by-step release-consumption guide for lattice-nodes and lattice-hub maintainers"
```

---

### Task 7: New `docs/making_a_protocol_change.md`

**Files:**
- Create: `docs/making_a_protocol_change.md`

**Interfaces:**
- Consumes: links to `docs/codegen_flow.md` (Task 4) and `docs/release_process.md` (Task 8) —
  safe to write these links regardless of task order, since filenames are fixed by this plan.
- Produces: none.

- [ ] **Step 1: Write the decision step**

A short section: "where does my change belong" — `message/message.go` for a new/changed wire
field, `opcodes/opcodes.go` for a new serial command, `adapter/types.go` for a new adapter
category. One example of each from real history if a clear one exists (e.g. cite the actual
`AuthPath[8]` addition as the `message/` example, since the spec already found this exact
commit — `v0.5.0`, chained HMAC route auth).

- [ ] **Step 2: Write the mechanical walkthrough**

Step-by-step: edit the chosen file, add/update the struct tag, run `go generate ./...`
(cross-reference `docs/codegen_flow.md` for what this actually does rather than re-explaining the
full pipeline here), read the diff in `c/` and `proto/` to sanity-check the generated output,
handle a `static_assert` failure if one occurs (explain what it means: the packed struct size
doesn't match the hardcoded `WireSize` — you likely need to update `WireSize` itself and consider
whether that's an intentional wire-breaking change), manually update `proto/mesh.options` if a
`bytes`-typed field's size changed (point to Task 4's explicit callout on this being unenforced
by CI).

- [ ] **Step 3: Write the semver-decision step**

Point to the (now-reconciled, per Task 2) semver policy in `CONTRIBUTING.md` rather than
duplicating the table here — one sentence plus a link.

- [ ] **Step 4: End with a pointer, not a duplicate**

Close with: "Ready to publish? See [Release Process](release_process.md) for the actual tagging
and coordination-handoff steps." Do not duplicate the tagging mechanics here — that content
belongs solely in Task 8's doc, per this plan's Global Constraints on cross-references.

- [ ] **Step 5: Self-review and commit**

```bash
git add docs/making_a_protocol_change.md
git commit -m "docs: add step-by-step guide for making a protocol change"
```

---

### Task 8: New `docs/release_process.md`

**Files:**
- Create: `docs/release_process.md`

**Interfaces:**
- Consumes: links to `docs/ecosystem.md` (Task 3) and `docs/consuming_a_release.md` (Task 6).
- Produces: none.

- [ ] **Step 1: Verify the actual current release mechanism before writing**

Run `git tag` (confirm the 9 tags `v0.1.0`–`v0.6.0` the spec cites are still the complete list —
a new tag may have shipped since). Run `gh release list` (confirm it's still empty — no GitHub
Releases used). Confirm `CHANGELOG.md` still doesn't exist at repo root. Confirm the README's
versioning table is genuinely the only changelog (read it).

- [ ] **Step 2: Write the actual lightweight process, not an aspirational one**

Step-by-step: update the README versioning table with the new entry (what format the existing
rows use — copy it exactly), commit that change, `git tag vX.Y.Z`, `git push origin vX.Y.Z` (or
`git push --tags` — check which convention recent tags actually used by looking at whether any CI
workflow triggers on tag push, which would indicate a preferred one-tag-at-a-time convention).
Explicitly state in the doc: "This repo does not use GitHub Releases or a CHANGELOG.md — the
README versioning table is the changelog" so a reader doesn't go looking for infrastructure that
doesn't exist.

- [ ] **Step 3: Write the coordination-handoff section**

After tagging: what to communicate to the `lattice-nodes` and `lattice-hub` maintainers (link to
`docs/ecosystem.md` for why this matters — the flag-day model — and to
`docs/consuming_a_release.md` for the exact commands they'll run). One or two sentences, not a
duplicate of either doc.

- [ ] **Step 4: Self-review and commit**

```bash
git add docs/release_process.md
git commit -m "docs: add release process doc (git-tag + README-table convention)"
```

---

## Execution Note

All 8 tasks touch disjoint file sets (Task 1: `README.md` only; Task 2: `CONTRIBUTING.md` +
`.github/ISSUE_TEMPLATE/bug_report.md`; Tasks 3-8: one new file each under `docs/`) — fully
parallelizable via worktrees, no sequencing dependency required despite the cross-reference links
between Tasks 4/6/7/8 (Markdown links to a not-yet-created file are valid; the target simply needs
to exist by the time this branch merges, which it will once all 8 tasks land).
