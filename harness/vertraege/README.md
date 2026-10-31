# Verträge der Harness-Werkzeuge

Dieser Ordner legt die Werkzeuge der Prüfumgebung fest: was ihre Prüfung zusagt, woran sie rot wird und was sie nicht prüft.
Keine Kennung dieses Ordners ist eine Zusage des Produkts, und keine erweitert eine Anforderung des Lastenhefts.
Die Schnittstelle eines Werkzeugs ist sein `make`-Target.

## Rang

Der Ordner teilt Rang 2 der Source Precedence mit `spec/spezifikation.md`: die Spezifikation für das Produkt, dieser Ordner für die Werkzeuge.
Das Lastenheft bleibt die Decke.
Grundlage ist [`MR-001`](../conventions.md#mr-001).

## Kennungen

Eine Datei je Kennung, `SPEC-<NNN>-<kurztitel>.md`; die Überschrift nennt Kennung und Titel.
Die Reihe der Kennungen zählt über beide Orte fortlaufend: Die nächste freie Nummer ist die höchste vergebene in `spec/spezifikation.md` und in diesem Ordner, plus eins.

## Verträge

| Kennung | Werkzeug | Datei |
|---|---|---|
| `SPEC-036` | `make build`, `make test` | [`SPEC-036-test.md`](SPEC-036-test.md) |
| `SPEC-037` | `make test-integration` | [`SPEC-037-test-integration.md`](SPEC-037-test-integration.md) |
| `SPEC-038` | `make test-determinism` | [`SPEC-038-test-determinism.md`](SPEC-038-test-determinism.md) |
| `SPEC-039` | `make a-check` | [`SPEC-039-architektur-gate.md`](SPEC-039-architektur-gate.md) |
| `SPEC-040` | `make docs-check` | [`SPEC-040-docs-check.md`](SPEC-040-docs-check.md) |
| `SPEC-041` | `make test-security` | [`SPEC-041-test-security.md`](SPEC-041-test-security.md) |
| `SPEC-042` | `make test-contract` | [`SPEC-042-test-contract.md`](SPEC-042-test-contract.md) |
| `SPEC-102` | `make gegenproben` | [`SPEC-102-gegenproben.md`](SPEC-102-gegenproben.md) |
| `SPEC-104` | `make migration-check` | [`SPEC-104-migration-check.md`](SPEC-104-migration-check.md) |
| `SPEC-101` | `make lint-profil` | [`SPEC-101-lint-profil.md`](SPEC-101-lint-profil.md) |
| `SPEC-122` | `make test-race` | [`SPEC-122-test-race.md`](SPEC-122-test-race.md) |
| `SPEC-128` | `make groessenmarke` | [`SPEC-128-groessenmarke.md`](SPEC-128-groessenmarke.md) |
| `SPEC-200` | `make handbuch-check` | [`SPEC-200-handbuch.md`](SPEC-200-handbuch.md) |
