# odata-recorder

**English** · [Deutsch](README.de.md)

> **Records the communication between your OData client and an OData service, replays it reproducibly without the service, and builds a stateful OData mock service from it.**

## What is odata-recorder?

`odata-recorder` is a tool for software developers and test automation. It sits
between an OData client and an existing OData service, records the HTTP/OData
exchanges in recording sessions, and replays them deterministically later. From
the metadata and the recorded data it also provides a locally modifiable mock
service with CRUD operations, query options, isolated workspaces and snapshots.
Integration tests then run without a permanently reachable production or test
server. It is aimed at developers, testers and CI systems. The first acceptance
scope is OData V4 with JSON and XML metadata. The requirements are in the
[Lastenheft](spec/lastenheft.md) (requirements specification, in German).

## What can I do today?

You can build the binary from source with `make build` (image
`odata-recorder:build`) and start it with the command `start`. It offers:

- the modes `record` (forward to the upstream and record), `replay` (answer from
  the recordings), `mock` (CRUD and query options on imported data) and `hybrid`
  (the mock first, recordings or upstream only where you allow it);
- an admin interface for sessions, imports, workspaces and snapshots;
- HTTPS for the service and for the admin interface.

How to use it is described in the [user manual](docs/user/benutzerhandbuch.md)
(in German). There is no release yet, and the first acceptance against the
requirements is still open.

The [Lastenheft](spec/lastenheft.md) (draft) defines the requirements with
acceptance criteria, the [specification](spec/spezifikation.md) the technical
details, and the [architecture](spec/architecture.md) the components. All three
are written in German. `make help` lists all targets.

## Why odata-recorder?

Integration tests against OData services often need a reachable service with a
suitable data set. That makes local environments and CI runs expensive, slow and
hard to reproduce, and tests can change the service unintentionally. For many of
these tests the service itself is not the subject, but the behavior observable
from the client's point of view. That is the behavior `odata-recorder` records and
makes repeatable offline.

## Core idea

A recording does not necessarily contain a complete data set. The recorder does
not pretend to be complete: incomplete coverage stays visible, and a storage
failure is reported instead of being presented as a successful recording. A
complete simulation of arbitrary business logic is not required.

## What makes it trustworthy?

- **Process:** [`AGENTS.md`](AGENTS.md) (hard rules every change must follow), [`harness/README.md`](harness/README.md) (source precedence: which source wins on conflict, and the gates).
- **Contracts:** [`spec/lastenheft.md`](spec/lastenheft.md) (`LH-*` IDs with acceptance criteria), then [`spec/spezifikation.md`](spec/spezifikation.md) and [`spec/architecture.md`](spec/architecture.md).
- **Gates:** `make gates` runs the gates; which ones exist and what each checks is listed in [`harness/README.md` §Sensors](harness/README.md#sensors-feedback-gates).
- **Auditability:** decisions in [`docs/plan/adr/`](docs/plan/adr/), planning in [`docs/plan/planning/`](docs/plan/planning/).
- **Traceability:** `make doc-trace` prints the requirements traceability matrix (requirement, decisions, slices); it is a report, not a gate.

## License

[MIT](LICENSE)
