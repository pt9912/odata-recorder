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
| `SPEC-201` | `make katalog-check` | [`SPEC-201-handbuch-katalog.md`](SPEC-201-handbuch-katalog.md) |
| `SPEC-210` | `tools/harness/mutationslauf.sh` (kein `make`-Ziel) | [`SPEC-210-mutationslauf.md`](SPEC-210-mutationslauf.md) |
| `SPEC-211` | `make test-integration`, `make test-determinism`, `make test-contract`, `make test-security` in der Gate-Kette, dazu die Gleichzeitigkeit und das Zeitbudget mit `make test-image` (`SPEC-323`); der Abbruch an die Prozessgruppe ist nur als Handlauf ohne Mutation gesehen, TERM allein an `make` und die Wartezeit sind ungefangen (Negativ mit Grund und Folge, Nr. 5 e) | [`SPEC-211-container-tests-in-der-kette.md`](SPEC-211-container-tests-in-der-kette.md) |
| `SPEC-212` | `make a-check`, `make docs-check`: Arbeitsbäume unter `.wt/` außerhalb der Prüfmenge | [`SPEC-212-arbeitsbaeume-ausserhalb-der-pruefmenge.md`](SPEC-212-arbeitsbaeume-ausserhalb-der-pruefmenge.md) |
| `SPEC-214` | PreToolUse-Guard `.claude/hooks/pretooluse-command-guard.sh` und sein Prüflauf `tools/harness/guard-faelle.sh` (kein `make`-Ziel, kein Gate, Nr. 10 d) | [`SPEC-214-command-guard.md`](SPEC-214-command-guard.md) |
| `SPEC-215` | `make dev-bau`, `make dev-start`, `make dev-probe`, `make dev-stop` (Arbeitscontainer; keine Gates; Privilegierung, Limits und Netz entscheidet [ADR-0031](../../docs/plan/adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md); Nr. 12 a: der echte Start ist nur ein Handlauf auf `u-boot`) | [`SPEC-215-arbeitscontainer.md`](SPEC-215-arbeitscontainer.md) |
| `SPEC-216` | `make lasttest` (Lasttest nach `LH-QA-01.a`: p95 des Gesamtspans für Replay und Mock; Messung außerhalb der Gate-Kette, kein Gate; der Lauf mit dem echten Binary ist nur ein Handlauf, Nr. 16 a) | [`SPEC-216-lasttest.md`](SPEC-216-lasttest.md) |
| `SPEC-233` | `make zeichen-check` | [`SPEC-233-unsichtbare-zeichen.md`](SPEC-233-unsichtbare-zeichen.md) |
| `SPEC-323` | `make test-image`, `make image` (Neustarttest des Laufzeit-Image `SPEC-322`; ohne Volume, im Daemon des Hosts; Betrieb mit Volume ungeprüft, Nr. 9 c) | [`SPEC-323-test-image.md`](SPEC-323-test-image.md) |
