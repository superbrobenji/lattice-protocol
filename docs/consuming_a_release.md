# Consuming a `lattice-protocol` release

This is a copy-paste, step-by-step guide for maintainers of the two downstream repos —
`lattice-nodes` and `lattice-hub` — who need to pin to a new `lattice-protocol` release, or bump
an existing pin. It assumes **zero knowledge of this repo's codegen pipeline**: you should not
need to run `go generate`, touch `cmd/gen-headers/`, or understand how `c/*.h` gets produced to
follow either path below. If you want that background anyway, see [`docs/codegen_flow.md`](codegen_flow.md)
and [`docs/ecosystem.md`](ecosystem.md).

Pick your path:

- [I maintain `lattice-nodes`](#i-maintain-lattice-nodes) (ESP32 firmware, git submodule consumer)
- [I maintain `lattice-hub`](#i-maintain-lattice-hub) (Go server, Go module consumer)

Both paths end with the same warning: **if the release you're pinning to changed `ProtoVersion`,
this is a flag-day release** — see [`docs/ecosystem.md`](ecosystem.md#policy-flag-day-releases-no-backcompat)
for what that means and why there's no compatibility shim.

---

## I maintain `lattice-nodes`

### Step 1: Find the current pin

`lattice-nodes` vendors this repo as a **git submodule**, declared in `lattice-nodes/.gitmodules`:

```ini
[submodule "firmware/main/lib/lattice-protocol"]
	path = firmware/main/lib/lattice-protocol
	url = https://github.com/superbrobenji/lattice-protocol
```

The submodule lives at `firmware/main/lib/lattice-protocol` inside the `lattice-nodes` repo. To
see what it's currently pinned to, run this from the `lattice-nodes` repo root:

```bash
git submodule status firmware/main/lib/lattice-protocol
```

This prints a commit SHA and (if it matches a tag exactly) the tag name in parentheses, e.g.:

```
 99cd30ccf4a5803c46810ce54563e1e6ff14f596 firmware/main/lib/lattice-protocol (v0.6.0)
```

That SHA/tag is a **fixed pin, not a live tracking pointer** — the vendored `c/*.h` headers only
change when someone explicitly bumps this pointer and commits it. Cloning or building
`lattice-nodes` without doing so gets you whatever was pinned, not whatever is newest on
`lattice-protocol`'s `main` branch.

### Step 2: Pin to a specific release tag

From the `lattice-nodes` repo root, `cd` into the submodule, fetch tags, and check out the one you
want:

```bash
cd firmware/main/lib/lattice-protocol
git fetch --tags
git checkout vX.Y.Z    # e.g. git checkout v0.6.0
cd ../../../..          # back to the lattice-nodes repo root
```

Alternatively, if you just want the latest commit on whatever branch the submodule is configured
to track (not recommended if you want a specific tag — this repo doesn't set a tracking branch by
default, so `--remote` will do nothing useful unless one is configured):

```bash
git submodule update --remote firmware/main/lib/lattice-protocol
```

For pinning to an exact release tag, the explicit `git fetch --tags && git checkout vX.Y.Z` form
above is the reliable one. Either way, the submodule checkout now has a **detached HEAD** at the
new commit — that's expected for a submodule pin.

Now record the new pin in the parent (`lattice-nodes`) repo — a submodule bump is really just a
change to what commit the parent repo points at:

```bash
git add firmware/main/lib/lattice-protocol
git commit -m "chore: bump lattice-protocol submodule to vX.Y.Z"
```

### Step 3: Verify the vendored headers match the tag

Confirm the checked-out submodule commit actually resolves to the tag you meant to pin:

```bash
git -C firmware/main/lib/lattice-protocol describe --tags
```

This should print exactly `vX.Y.Z` (no `-N-gHASH` suffix — a suffix means you're on a commit
*after* that tag, not the tag itself). Then spot-check that the vendored headers under
`firmware/main/lib/lattice-protocol/c/` — `opcodes.h`, `adapter_types.h`, `message_types.h`,
`mesh_message.h` — actually changed if you expected them to:

```bash
git -C firmware/main/lib/lattice-protocol log --oneline -1
git status firmware/main/lib/lattice-protocol   # should be clean once the parent-repo commit lands
```

### Step 4: What a wire-size bump means for your firmware build

`c/mesh_message.h` ends with:

```c
static_assert(sizeof(mesh_message) == 200, "mesh_message size changed — update server proto");
```

The `200` is `lattice-protocol`'s `WireSize` constant baked into the generated header at release
time. If the release you just pinned to changed the on-wire struct size (a `WireSize` bump), this
`static_assert` is a **C-compile-time check** — it will fail your very next `idf.py build` with a
compile error naming this line, not a silent runtime bug. That's by design: `lattice-protocol`'s
own CI never compiles this header (see [`docs/codegen_flow.md`](codegen_flow.md) for why), so the
first place a `WireSize` mismatch can actually be caught is here, in your firmware build.

If you hit this failure, it almost always means the release is a flag-day wire-format change (see
above) — check the version's entry in this repo's own
[README versioning table](../README.md#versioning) for what changed before rebuilding. There is no code change required in `lattice-nodes` to fix the
`static_assert` itself (it's generated and correct for the new size) — what you likely need to
check is whether any of your own firmware code hardcodes assumptions about field widths (e.g.
buffer sizes derived from the old `RoutePath`/`AuthPath` lengths). `lattice-nodes`' own
[`docs/server_requirements.md`](https://github.com/superbrobenji/lattice-nodes/blob/main/docs/server_requirements.md)
documents the current `mesh_message` layout and its `static_assert` in more detail if you need a
reference for what "in sync" looks like.

Once it builds clean, rebuild and reflash every node before it talks to a hub running the new
protocol version — see the flag-day note in [`docs/ecosystem.md`](ecosystem.md#policy-flag-day-releases-no-backcompat)
for why partial rollout means partial mesh outage, not graceful degradation.

---

## I maintain `lattice-hub`

### Step 1: Bump the module version

`lattice-hub` imports this repo as an ordinary Go module dependency, declared in
`server/orchestrator/go.mod` (the sibling `server/sidecar/go.mod` module does not depend on
`lattice-protocol` at all — it has no mesh wire-format code, so it needs no change). From the
`server/orchestrator/` directory:

```bash
go get github.com/superbrobenji/lattice-protocol@vX.Y.Z
```

This rewrites the `require github.com/superbrobenji/lattice-protocol vX.Y.Z` line in
`server/orchestrator/go.mod` and updates `go.sum` accordingly.

### Step 2: Tidy the module

```bash
go mod tidy
```

Run this from `server/orchestrator/` as well. It reconciles `go.mod`/`go.sum` with what's actually
imported — if the new release added or removed transitive dependencies, this is what picks that
up.

### Step 3: Read the changelog before you assume anything

Before touching code, read the target version's entry in this repo's own
[README versioning table](../README.md#versioning). Every entry states whether the release is wire-format-breaking ("flag-day, no backcompat") or
additive, and summarizes what changed — e.g. renamed/removed constants, resized struct fields, new
opcodes. This tells you what kind of failure (if any) to expect in Step 4, and whether you'll also
need to update a hardcoded protocol-version check on the hub side (see
[`docs/ecosystem.md`](ecosystem.md#policy-flag-day-releases-no-backcompat) for the flag-day
mechanism — `lattice-hub`'s `server/orchestrator/mesh/server.go` does the equivalent check to
firmware's `PROTO_VERSION` guard).

### Step 4: Let the compiler do the rest

```bash
go build ./...
```

Run from `server/orchestrator/` — that's the Go module that imports `lattice-protocol` (`lattice-hub`
has no root `go.mod`; `server/orchestrator/` and `server/sidecar/` are separate modules, and
`server/sidecar` doesn't import this one, so it's unaffected). Any breaking Go-level API change in
the new release — a renamed or
removed constant, a changed struct field, a changed function signature — surfaces immediately as a
compile error naming the exact file and line in your code that needs updating. Fix each one, then
re-run `go build ./...` until it's clean.

One thing the compiler **won't** catch: a wire-format-only change (e.g. a new `ProtoVersion` with
no Go API changes) builds cleanly but changes what the server accepts or emits at runtime. If
Step 3 told you this release is a flag-day bump, update the hardcoded protocol-version check
before deploying — don't rely on `go build` succeeding as proof the upgrade is safe.

Finally, run the existing test suite before deploying (same `server/orchestrator/` directory):

```bash
go test ./...
```

Deploy the new hub build only after this passes. As with the firmware side, if this is a flag-day
release, coordinate the hub deploy with reflashing any nodes still running old firmware — see the
flag-day section of [`docs/ecosystem.md`](ecosystem.md#policy-flag-day-releases-no-backcompat).

---

## See also

- [`docs/ecosystem.md`](ecosystem.md) — how the three repos relate, and the flag-day versioning
  policy referenced throughout this doc.
- [`docs/codegen_flow.md`](codegen_flow.md) — how `c/*.h` and `proto/mesh.proto` are generated from
  Go source, if you want to understand what's actually inside a release rather than just consuming
  one.
- [`README.md`](../README.md#versioning) — the versioning table with a one-line changelog entry per
  tagged release.
