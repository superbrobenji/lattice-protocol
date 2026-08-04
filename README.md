# lattice-protocol

[![CI](https://github.com/superbrobenji/lattice-protocol/actions/workflows/ci.yml/badge.svg)](https://github.com/superbrobenji/lattice-protocol/actions/workflows/ci.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

Shared protocol definitions for the Lattice mesh network. Defines the opcode and adapter-type constants consumed by all Lattice services and firmware.

Used by:
- **motionSensorServer** (Go) — imports as a Go module
- **Lattice-nodes** (ESP32/C++) — includes via git submodule

## Packages

| Package | Description |
|---------|-------------|
| `opcodes/` | Serial command opcode constants (Go) |
| `adapter/` | Adapter type identifiers and helpers (Go) |
| `c/` | Generated C headers for firmware — do not edit directly |
| `cmd/gen-headers/` | Generator that writes `c/` from the Go constants |

## Usage

### Go (motionSensorServer)

```go
import (
    "github.com/superbrobenji/lattice-protocol/opcodes"
    "github.com/superbrobenji/lattice-protocol/adapter"
)

payload[0] = opcodes.OpLEDSolid
if adapter.IsOutput(node.AdapterType) { ... }
```

### C (Lattice-nodes — via git submodule at `main/lib/lattice-protocol`)

```c
#include "lib/lattice-protocol/c/opcodes.h"
#include "lib/lattice-protocol/c/adapter_types.h"

payload[0] = OP_LED_SOLID;
```

Set up the submodule:

```sh
git submodule update --init
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full protocol contribution workflow — adding opcodes or adapter types, regenerating C headers, and opening a PR.

## Versioning

This module follows semver. Consumers pin to a tag.

| Tag | Notes |
|-----|-------|
| v0.5.0 | Protocol v4 wire: AuthPath[8] — chained HMAC-SHA256-64 authenticating relay-accumulated route_path (issue lattice-nodes#44). Flag-day, no backcompat. WireSize=250. |
| v0.4.2 | gen-headers emit textual include guards (`LATTICE_<FILE>_H`) on all `c/*.h`; wire format unchanged |
| v0.4.1 | nanopb `max_size` options for proto v3 fields (`routePath`, `authTag`, secondary master) |
| v0.4.0 | Protocol v3 wire format: source-routed downlink (`RoutePath[60]` + `RouteLen`), E2E AEAD auth tag (`AuthTag[16]`), dual-master JOIN_ACK fields (`SecondaryMasterMac`, `SecondaryPublicKey[32]`); `WireSize = 242` |
| v0.3.0 | Add health opcodes 0xB0/0xB1/0xB2 |
| v0.2.1 | Lower go directive to 1.21.0 |
| v0.2.0 | Generated C headers; submodule support |
| v0.1.0 | Initial release |

## License

Copyright (C) 2026 Lattice Contributors

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

See [LICENSE](LICENSE) for the full terms.
