# The Lattice ecosystem

`lattice-protocol` is the **upstream** repo of a three-repo system. It defines the wire format
and shared constants; it does not depend on either of the other two repos. The dependency graph
is one-directional:

```
lattice-protocol  (this repo — wire format source of truth)
      ^                              ^
      |                              |
  git submodule                  Go module
      |                              |
lattice-nodes                  lattice-hub
(ESP32 firmware)              (Go mesh server)
```

- **`lattice-nodes`** — the ESP32 firmware that runs on every mesh node. Vendors the generated
  `c/*.h` headers from this repo as a git submodule.
- **`lattice-hub`** — the Go server the master node talks to over USB serial (enrollment
  approval, REST API, dashboards). Imports this repo as a Go module.

Both consumers derive their wire-format types from the same source (`message/message.go` in this
repo, generated into `c/mesh_message.h` for firmware and consumed directly as Go structs for the
server) — that shared origin is what keeps a firmware node and a hub server able to parse each
other's frames at all.

## How each repo consumes this one

### `lattice-nodes`: git submodule

`lattice-nodes` vendors this repo as a git submodule at
`firmware/main/lib/lattice-protocol` (declared in `lattice-nodes/.gitmodules`, pointing at
`https://github.com/superbrobenji/lattice-protocol`). At the time of writing it is pinned to
`v0.6.0` (commit `99cd30c`).

A git submodule pin is a commit reference, not a live import: the vendored copy of `c/*.h` inside
`lattice-nodes` only changes when someone in that repo explicitly runs a submodule update and
commits the new pointer. Cloning or building `lattice-nodes` without updating the submodule
reference gets you whatever headers were pinned at that commit — including the `mesh_message`
struct layout, its `static_assert(sizeof(mesh_message) == 200, ...)`, and the opcode/adapter-type
constants — not whatever is newest on `lattice-protocol`'s `main` branch.

### `lattice-hub`: Go module

`lattice-hub` imports this repo as an ordinary Go module dependency. Its
`server/orchestrator/go.mod` declares (the sibling `server/sidecar/go.mod` module does not depend
on `lattice-protocol` — it has no mesh wire-format code):

```
require github.com/superbrobenji/lattice-protocol v0.6.0
```

As with the submodule pin, this is a fixed version — `go build`/`go test` resolve against exactly
`v0.6.0` until someone bumps the `require` line and runs `go mod tidy`.

Both consumers, in other words, are pinned to a specific tagged release of this repo at any given
time; neither auto-tracks `main`.

## Policy: flag-day releases, no backcompat

**Flag-day** is this repo's explicit versioning policy: a wire-format change that alters the
`ProtoVersion`/`PROTO_VERSION` value is not backward compatible, and no dual-version support
period is provided. A node or server built against the old protocol version and one built against
the new version cannot successfully exchange mesh messages after the cutover — there is no
negotiation, no version range, no shim.

This is already the repo's practice, though the `README.md` versioning table only labels it
consistently for recent rows: the `v0.5.0` and `v0.6.0` entries are explicitly annotated
"Flag-day, no backcompat" (matching the commit messages that introduced them), but earlier
wire-breaking releases — e.g. `v0.4.0`, which changed `WireSize` to 242 — carry no such label, and
most rows in the table have neither a wire-format change nor an annotation. The label is not a
reliable signal on its own; check the `WireSize`/field changes described in each row to determine
whether a given release is wire-breaking. This document states the underlying policy explicitly
for the first time: **any change to `ProtoVersion` is a flag day**, labelled or not.

The mechanism is enforced independently in both consumers, not by this repo: `lattice-nodes`
defines `PROTO_VERSION` as a `constexpr` in `firmware/main/src/mesh/MeshMessenger.h` and
`Mesh.cpp` drops any received message where `msg.proto_version != PROTO_VERSION`; `lattice-hub`'s
`server/orchestrator/mesh/server.go` does the equivalent check (`if msg.ProtoVersion != 5 { ...
dropping }`). Firmware or servers left on the old version are silently ignored by anything running
the new version, not gracefully downgraded.

Why flag-day rather than versioned compatibility shims: this is a single-deployment mesh of
resource-constrained embedded devices (the wire frame itself is capped at `WireSize = 200` bytes
to fit the ESP-NOW payload limit), not a multi-tenant service with independently-operated clients.
Maintaining parallel decode paths for old and new wire formats on-device would cost flash and RAM
budget that these nodes don't have, for a compatibility window nobody outside this project's own
rollout needs. This rationale is inferred from the codebase's constraints, not stated anywhere in
its own history — treat it as an observation, not a documented decision.

## What a new tag means operationally

### If you maintain `lattice-nodes`

1. Bump the submodule: `cd firmware/main/lib/lattice-protocol && git fetch --tags && git checkout vX.Y.Z` (or `git submodule update --remote`), then commit the updated submodule pointer from the parent repo.
2. The vendored `c/*.h` headers change with it — re-check `c/mesh_message.h`'s `static_assert(sizeof(mesh_message) == ...)` still matches what the firmware build expects, and re-check any opcode/adapter-type constants your firmware code references against `c/opcodes.h` / `c/adapter_types.h`.
3. If the tag changed `ProtoVersion`/`WireSize` (a flag-day release per above), update `PROTO_VERSION` in `firmware/main/src/mesh/MeshMessenger.h` to match.
4. Rebuild and **fully reflash every node** before it talks to a hub running the new protocol version. This is not optional for a flag-day bump: a node still running old firmware will have its messages silently dropped by a hub on the new version (and vice versa), so partial rollout means partial mesh outage, not graceful degradation.

### If you maintain `lattice-hub`

1. Bump the `require github.com/superbrobenji/lattice-protocol` line in `server/orchestrator/go.mod` to the new tag, then run `go mod tidy`. (`server/sidecar/go.mod` does not import this repo, so it needs no change.)
2. Rebuild (`go build ./...`) — a breaking Go-level API change in the new tag (renamed/removed constants, changed struct fields) will fail the build immediately; a wire-format-only change (new `ProtoVersion`) will build cleanly but change what the server accepts at runtime.
3. If the tag is a flag-day release, update the hardcoded protocol-version check (e.g. `server/orchestrator/mesh/server.go`'s `if msg.ProtoVersion != 5`) to the new value before deploying.
4. Deploy the new hub build. The same flag-day caveat applies from the server side: any node still running old firmware in the field will have its messages dropped by the `ProtoVersion` check until it is reflashed — the hub bump and the node reflash need to be coordinated, not independently scheduled.
