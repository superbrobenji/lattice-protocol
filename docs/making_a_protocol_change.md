# Making a protocol change

This is a step-by-step guide for adding or changing a wire field, opcode, or adapter type in
`lattice-protocol`, aimed at someone who doesn't already know this repo's codegen mechanism. For
how the generator actually works end to end (what `go generate ./...` runs, why reflection drives
some of it and hand-written code drives the rest, how CI enforces it, and where that enforcement
has a known gap), see [`docs/codegen_flow.md`](codegen_flow.md) — this doc won't re-derive that,
only tell you which steps to take and why each one matters.

## 1. Where does my change belong?

| You want to... | Edit |
|---|---|
| Add or change a field on the wire frame itself (the bytes sent over ESP-NOW) | `message/message.go` (the `MeshMessage` struct) |
| Add a new serial command byte (byte 0 of a `MessageTypeSerialCmdBroadcast` payload — server↔node commands and acks) | `opcodes/opcodes.go` |
| Add a new adapter/sensor category (e.g. a new sensor type) | `adapter/types.go` |

There's a fourth, rarer case: a new *frame-level* message type (as opposed to adapter payload
data carried inside `MeshMessage.Data`) goes in `message/types.go` instead. It's mechanically
closer to opcodes/adapter types than to a `MeshMessage` field change — see the note in step 2.2.

**Real example — a `message/` change:** `AuthPath[8]` was added to `MeshMessage` in `v0.5.0`
(commit `b928b11`, PR #37), a chained HMAC-SHA256 field (truncated to 8 bytes) carried over the
relay-accumulated route path. The whole change was 8 lines in `message/message.go`:

```diff
+	// v4: chained HMAC-SHA256-64 over the relay-accumulated route_path (Phase C, issue #44).
+	AuthPath [8]byte `c:"uint8_t[8]" proto:"17,bytes,optional,authPath"`
```

plus bumping `WireSize` from `242` to `250` and updating its doc comment. That one commit is used
throughout this doc as the running example — including a real mistake in it that's worth learning
from (step 2.1, mesh.options).

## 2. Making the change

The mechanics differ depending on which file you touched, because `cmd/gen-headers/main.go`
generates them differently: `message/message.go`'s fields are picked up automatically via Go
reflection over struct tags, but opcodes, adapter types, and message types are each a hand-written
list of calls in `main.go` — a new Go constant in those files is invisible to the generator until
you also add a line for it there. Getting this backwards is the single easiest mistake to make.
`docs/codegen_flow.md` has the full explanation of why the two mechanisms differ; this section is
just the checklist.

### 2.1 Changing `message/message.go` (a wire field)

1. Add or edit the field on the `MeshMessage` struct with both tags:

   ```go
   YourField [N]byte `c:"uint8_t[N]" proto:"18,bytes,optional,yourFieldName"`
   ```

   `c:"..."` is the generated C type; `proto:"fieldNum,protoType[,optional][,protoName]"` is the
   generated proto3 field. Field order in the struct is wire order — don't reorder existing
   fields for a non-breaking change. Proto field numbers are never reused, including ones freed
   up by a field removal: `AuthPath` uses `17`, not `15`/`16` — those were `SecondaryMasterMac`
   and `SecondaryPublicKey`, removed in the `v0.6.0` wire-shrink and left retired.

2. Update the `WireSize` constant *and its doc comment* to the new total packed size, before you
   regenerate. `WireSize` is not computed from the struct — it's a hand-maintained number that the
   generator bakes as a literal into `c/mesh_message.h`'s `static_assert(sizeof(mesh_message) ==
   N, ...)`. If you regenerate with a stale `WireSize`, `go generate` will succeed and produce a
   header that compiles to the *wrong* asserted size — nothing in this repo's own tooling checks
   the arithmetic, only that the header text matches whatever `WireSize` currently says. The
   mismatch only becomes visible as a hard compile error the first time `lattice-nodes`'
   ESP-IDF build actually compiles against the new header — this repo's CI never invokes a C
   compiler. If you do hit that failure downstream, it means the packed struct size and
   `WireSize` disagree; recompute `WireSize` by hand (sum of every field's `c:` width, same as
   the existing doc-comment arithmetic) and fix whichever one is wrong.

3. Run `go generate ./...` (or `make generate`).

4. Read the diff in `c/mesh_message.h` and `proto/mesh.proto` — for this file, both are
   reflection-driven off your struct tags, so the diff should just reflect the field you added.
   If it doesn't, check your tag syntax rather than editing the generated files by hand.

5. If your field is `bytes`-typed and its size is new or changed, manually add or update the
   matching line in `proto/mesh.options`:

   ```
   mesh.MeshMessage.yourFieldName max_size:N
   ```

   **Nothing generates or checks this file.** `docs/codegen_flow.md` documents this as a known,
   currently-unmitigated gap in CI, and the `AuthPath` example above is live proof it's not
   theoretical: the `v0.5.0` commit that added `AuthPath[8]` never added a
   `mesh.MeshMessage.authPath max_size:8` entry, and as of this writing `proto/mesh.options` still
   has no `authPath` line at all — CI passed anyway, then and now, because `make check` only diffs
   *generated* files and this one isn't generated. Don't assume a green CI run means this file is
   correct; check it yourself every time a `bytes` field's size changes.

6. Decide whether this is a flag-day, no-backcompat release (a `WireSize`/`ProtoVersion` change
   that breaks compatibility with nodes/hubs still on the old version) — see
   [`docs/ecosystem.md`](ecosystem.md) for what that policy means operationally for the other two
   repos. Most `message/` changes are flag-day; that's not a niche case here.

### 2.2 Changing `opcodes/opcodes.go`, `adapter/types.go`, or `message/types.go`

1. Add the new Go constant, following the existing naming and grouping convention in the file.

2. **Also add a matching call in `cmd/gen-headers/main.go`** — a `writeCDefByte` call in
   `writeOpcodesHeader` for a new opcode, `writeCDefInt32` in `writeAdapterTypesHeader` for a new
   adapter type, or `writeCDefUint8` in `writeMeshMessageTypesHeader` for a new message type. This
   step is easy to miss because nothing forces it: these three generated headers are each a fixed
   list of hand-written calls in `main.go`, not a reflection loop like `MeshMessage`. If you skip
   this, `go generate ./...` still runs cleanly and produces *no diff at all* for your new
   constant — it's simply absent from the generated header, silently. `make check` (CI's
   `header-sync` job) will still pass, because there's nothing generated to disagree with the
   committed file. The only symptom is that firmware or server code referencing the new
   `#define`/proto entry won't compile — potentially in a different repo, much later.

   This isn't hypothetical: the `v0.3.0` health-opcode addition (commit `8605ea9`) touched
   `cmd/gen-headers/main.go` alongside `opcodes/opcodes.go` for exactly this reason — 16 lines in
   `main.go`, not zero.

3. Run `go generate ./...` (or `make generate`).

4. Read the diff in the relevant header (`c/opcodes.h`, `c/adapter_types.h`, or
   `c/message_types.h`) and confirm your new `#define` actually appears. An empty diff here after
   adding a constant is the tell that you missed step 2, not a sign everything's fine.

## 3. Which semver bump?

This repo's semver policy — what bumps patch vs. minor, and the one documented historical
exception to it — lives in [`CONTRIBUTING.md`'s "Semver rules"](../CONTRIBUTING.md#semver-rules)
section; check there rather than guessing from the change type.

## 4. Publish it

Once your change, its generated files, and any `proto/mesh.options` update are committed and
merged: see [Release process](release_process.md) for the actual tagging and
cross-repo coordination-handoff steps. That's a separate concern from making the change itself,
and it isn't repeated here.
