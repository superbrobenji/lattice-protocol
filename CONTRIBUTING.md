# Contributing to lattice-protocol

## What lives here

This repo contains shared protocol definitions for the Lattice mesh network:

- **`opcodes/opcodes.go`** — serial command opcode constants (Go)
- **`adapter/types.go`** — adapter type identifiers and helpers (Go)
- **`c/`** — C headers generated from the Go constants; never edit these by hand

Changes here affect all consumers: `lattice-hub` (imports as Go module) and `Lattice-nodes` (includes as git submodule). Treat every change as a protocol change.

## Prerequisites

- Go 1.21 or later (`go version`)
- Git

## Adding an opcode

1. Edit `opcodes/opcodes.go`. Add your constant in the appropriate group, following the existing naming convention (`OpXxx` in Go, which becomes `OP_XXX` in the generated C header).

   ```go
   const (
       OpYourNewOpcode Opcode = 0xXX
   )
   ```

2. Regenerate C headers:

   ```sh
   make generate
   # equivalent: go generate ./...
   ```

3. Verify the headers are in sync with the Go constants:

   ```sh
   make check
   # Runs: go generate ./... && git diff --exit-code c/ proto/
   # Must exit 0 before you commit.
   ```

4. Run tests:

   ```sh
   go test ./...
   go vet ./...
   ```

5. Commit Go source and generated `c/` files together in one commit:

   ```sh
   git add opcodes/opcodes.go c/opcodes.h
   git commit -m "feat(opcodes): add OpYourNewOpcode (0xXX)"
   ```

## Adding an adapter type

Same flow as adding an opcode, but edit `adapter/types.go` instead and include `c/adapter_types.h` in the commit.

## Changing the wire format

The `MeshMessage` struct (`message/message.go`) is the packed frame sent over ESP-NOW between
nodes and relayed to the master. It has more moving parts than an opcode change: one file you
edit, two files that are generated from it, one hand-maintained file that isn't generated but
still needs to match, and a cross-repo release step.

1. Edit `message/message.go`. Add or change a field on `MeshMessage`, following the existing
   struct-tag format:

   ```go
   YourField [N]byte `c:"uint8_t[N]" proto:"18,bytes,optional,yourFieldName"`
   ```

   `c:"..."` drives the generated C field type; `proto:"fieldNum,protoType[,optional][,protoName]"`
   drives the generated proto3 field. Field order in the struct is the wire order — do not
   reorder existing fields. Proto field numbers are never reused; pick the next unused one.

2. Update the `WireSize` constant and its doc comment to the new total packed size. `WireSize`
   is not computed from the struct automatically — it's a hand-maintained constant, and the
   generator bakes its current value into the C header's compile-time check (next step), so it
   must be correct *before* you regenerate.

3. Regenerate the C header and proto file:

   ```sh
   make generate
   # equivalent: go generate ./...
   ```

   This rewrites `c/mesh_message.h` and `proto/mesh.proto` from `message/message.go`.

4. `c/mesh_message.h` carries a compile-time size check:

   ```c
   static_assert(sizeof(mesh_message) == 200, "mesh_message size changed — update server proto");
   ```

   The `200` is whatever `WireSize` was when you last ran `go generate` — the generator writes
   the literal, it doesn't compute it. If your struct change alters the packed size but you
   didn't update `WireSize` first, this line will bake in the *old*, now-wrong number, and the
   mismatch is only caught when a C compiler builds against this header — i.e. in the
   `lattice-nodes` firmware build, not in this repo's own CI. This repo never compiles the C
   headers itself; `header-sync` only runs `make check` (`go generate ./... && git diff
   --exit-code c/ proto/`), which catches a header that's out of sync with the Go source, not a
   `WireSize` that's internally wrong.

5. If the field you touched is a `bytes` field whose size changed, manually update its entry in
   `proto/mesh.options`:

   ```
   mesh.MeshMessage.yourFieldName max_size:N
   ```

   **This file is not generated, and nothing checks it.** Unlike `c/*.h` and `proto/mesh.proto`,
   it carries no "Code generated ... DO NOT EDIT" banner — `go generate` never writes it. `make
   check` diffs the *generated* files against the working tree; since `mesh.options` is never
   part of what gets generated, there is nothing to diff it against, so a stale `max_size` here
   will not fail CI no matter how out of sync it gets. This is a real, currently-unfixed gap in
   the tooling, not a hypothetical one — treat every `bytes` field size change as a reason to
   check this file by hand.

6. Run tests:

   ```sh
   go test ./...
   go vet ./...
   ```

7. Commit the Go source and every file you touched or regenerated together:

   ```sh
   git add message/message.go c/mesh_message.h proto/mesh.proto proto/mesh.options
   git commit -m "feat(message): <describe the field change>"
   ```

8. Coordinate the release. A wire-format change is a hard dependency for `lattice-hub` and
   `lattice-nodes` — see [`docs/release_process.md`](docs/release_process.md) for the actual
   tagging and cross-repo update mechanics. This section only covers what to change here.

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
