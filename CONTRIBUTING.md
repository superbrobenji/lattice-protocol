# Contributing to lattice-protocol

## What lives here

This repo contains shared protocol definitions for the Lattice mesh network:

- **`opcodes/opcodes.go`** — serial command opcode constants (Go)
- **`adapter/types.go`** — adapter type identifiers and helpers (Go)
- **`message/`** — `MeshMessage` wire-format struct (the packed protocol frame) and message-type constants (Go)
- **`c/`** — C headers generated from the Go constants; never edit these by hand
- **`proto/`** — generated `mesh.proto` plus hand-maintained `mesh.options` nanopb sizing file

Changes here affect all consumers: `lattice-hub` (imports as Go module) and `lattice-nodes` (includes as git submodule). Treat every change as a protocol change.

## Prerequisites

- Go 1.21 or later (`go version`)
- Git

## Adding an opcode

Opcodes are **not** reflected into the C header automatically: `cmd/gen-headers/main.go` builds
`c/opcodes.h` from a fixed list of hand-written `writeCDefByte(...)` calls in `writeOpcodesHeader`
(one call per constant) — it does not scan `opcodes/opcodes.go`. Adding a new opcode means editing
**both** files; skip the second one and `go generate` succeeds with no diff, silently leaving your
new constant out of the generated header while `make check` still passes (see
[`docs/making_a_protocol_change.md`](docs/making_a_protocol_change.md) for the full explanation).

1. Edit `opcodes/opcodes.go`. Add your constant in the appropriate group, following the existing naming convention (`OpXxx` in Go, which becomes `OP_XXX` in the generated C header):

   ```go
   const (
       OpYourNewOpcode = byte(0xXX)
   )
   ```

2. Add a matching call in `cmd/gen-headers/main.go`'s `writeOpcodesHeader` function:

   ```go
   writeCDefByte(f, "OP_YOUR_NEW_OPCODE", opcodes.OpYourNewOpcode, "describe the payload/semantics")
   ```

3. Regenerate C headers:

   ```sh
   make generate
   # equivalent: go generate ./...
   ```

4. Verify the headers are in sync with the Go constants:

   ```sh
   make check
   # Runs: go generate ./... && git diff --exit-code c/ proto/
   # Must exit 0 before you commit.
   ```

5. Run tests:

   ```sh
   go test ./...
   go vet ./...
   ```

6. Commit the Go source, the generator change, and the regenerated `c/` file together in one commit:

   ```sh
   git add opcodes/opcodes.go cmd/gen-headers/main.go c/opcodes.h
   git commit -m "feat(opcodes): add OpYourNewOpcode (0xXX)"
   ```

## Adding an adapter type

Same flow as adding an opcode: edit `adapter/types.go`, add a matching `writeCDefInt32(...)` call in `cmd/gen-headers/main.go`'s `writeAdapterTypesHeader` function, regenerate, and include `c/adapter_types.h` in the commit.

## Changing the wire format

The `MeshMessage` struct (`message/message.go`) is the packed frame sent over ESP-NOW between
nodes and relayed to the master. It's the highest-consequence change type in this repo: one file
you edit, two files generated from it, one hand-maintained file that isn't generated but still
needs to match by hand, and a cross-repo release-coordination step.

For the full mechanical walkthrough — struct-tag syntax, updating `WireSize`, regenerating
`c/mesh_message.h` and `proto/mesh.proto`, and committing everything together — see
[`docs/making_a_protocol_change.md`](docs/making_a_protocol_change.md). This section only covers
the two decisions that are easy to get wrong:

- **Most wire-format changes are flag-day, no-backcompat releases.** A `WireSize`/`ProtoVersion`
  bump breaks compatibility with nodes/hubs still on the old version, with no dual-version support
  period — see [`docs/ecosystem.md`](docs/ecosystem.md#policy-flag-day-releases-no-backcompat) for
  what that means operationally for `lattice-nodes` and `lattice-hub`.
- **`proto/mesh.options` is hand-maintained, and nothing checks it.** If the field you touched is
  a `bytes` field whose size changed, you must manually add or update its `max_size:N` entry in
  `proto/mesh.options`. `make check` only diffs *generated* files against the working tree, so a
  stale or missing entry here will not fail CI no matter how out of sync it gets — treat every
  `bytes` field size change as a reason to check this file by hand.

Once your change, its generated files, and any `proto/mesh.options` update are committed and
merged, coordinate the release — see [`docs/release_process.md`](docs/release_process.md) for the
actual tagging and cross-repo update mechanics.

## Semver rules

| Change | Bump |
|--------|------|
| Add new opcode or adapter type | Patch (`v0.Y.Z+1`) |
| Rename or remove an existing constant | Minor (`v0.Y+1.0`) |
| Breaking protocol redesign | Open an issue first |

> **Historical note:** `v0.3.0` ("add health opcodes 0xB0/0xB1/0xB2", a plain opcode addition)
> was tagged as a minor bump from `v0.2.1`. It predates this table — the table was written
> about an hour later the same day (2026-06-30) and was never applied retroactively. Nothing
> since has broken the rule; treat the table above as current policy.

After merging, create and push a semver tag:

```sh
git tag v0.Y.Z
git push origin v0.Y.Z
```

Update the versioning table in `README.md` with the new tag and a one-line description of what changed.

## Code style

- Run `gofmt` before committing (enforced by `go vet ./...`)
- Constant names: `OpXxx` for opcodes, `AdapterTypeXxx` for adapter types
- Group related constants together, separated by a blank line from unrelated groups

## Pull request process

1. Fork the repo and create a branch: `feature/your-opcode-name` or `fix/your-fix`
2. Follow the steps above for adding an opcode or adapter type
3. Ensure `make check`, `go test ./...`, and `go vet ./...` all pass locally
4. Open a PR against `main` and fill in the PR template
5. All CI jobs (`go-test`, `header-sync`, `go-lint`, CodeQL) must be green; 1 approving review required

## Repository setup

After the first merge to `main`, configure branch protection in GitHub settings (Settings → Branches → Add rule for `main`):

- Require status checks to pass before merging
  - Required checks: `Test` (go-test job), `Header sync` (header-sync job)
- Require approvals: 1
- Dismiss stale pull request approvals when new commits are pushed
- Do not allow bypassing the above settings
