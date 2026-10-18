# ADR-0003: Lint-Profil mit golangci-lint

**Status:** Accepted

**Datum:** 2026-10-18

**Autor:** pt9912

**Bezug:** [ADR-0001](0001-implementierungssprache-go.md), [ADR-0002](0002-architektur-gate-a-check.md)

**Schärft:** [`SPEC-101`](../../../spec/spezifikation.md#7-festlegungen-der-harness-werkzeuge)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Mit Go (ADR-0001) braucht das Repo eine statische Prüfung des Codes.
`pgwire-recorder` und `d-check` verwenden golangci-lint v2 mit demselben SOLID-nahen Profil aus 28 Lintern; `d-check` ergänzt `nolintlint`.
`AGENTS.md` §3.2 verbietet Unterdrückungen im Code, Ausnahmen leben in der zentralen Konfigurationsdatei.

## Entscheidung

Wir wählen **golangci-lint v2** mit dem Profil `.golangci.yml` an der Wurzel.
Das Profil ist das von `d-check` (mit `nolintlint` scharf) auf der Linter-Menge von `pgwire-recorder`: 5 Standard-Linter, 23 SOLID-nahe Linter und `nolintlint`.
Architektur-Regeln prüft `a-check` (ADR-0002), deshalb ist `depguard` nicht aktiv.
Bibliotheken lässt `gomodguard_v2` nur mit Eintrag im Profil zu; heute ist keine zugelassen, jede Aufnahme braucht eine ADR.
Das Image ist per Digest gepinnt.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — nur `go vet` und `staticcheck` | kein eigenes Profil | keine Komplexitäts-, Schnittstellen- und Kontext-Regeln; Unterdrückungen unkontrolliert |
| B — eigenes, kleineres Profil | weniger Befunde am Anfang | dritte Variante neben zwei bewährten Profilen |
| **C — Profil von `d-check` und `pgwire-recorder` (gewählt)** | erprobt, gleiche Schwellen in allen Schwesterprojekten, `nolintlint` bindet §3.2 | strenge Schwellen (Komplexität 15, Funktionslänge 100) |

## Konsequenzen

- Positiv: Ein Profil für alle Projekte; Ausnahmen sind sichtbar und begründet (`# Why:`).
- Negativ: Das Gate prüft heute nur das Profil gegen das Schema von golangci-lint, keinen Code.
- Folgepflicht: Mit dem ersten Code-Slice die Stufe im Dockerfile, die Prüfung des Codes samt Gegenprobe und die ADR für jede zugelassene Bibliothek.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| golangci-lint | `config verify` gegen `.golangci.yml` | `make lint-profil` |

## Re-Evaluierungs-Trigger

Wenn golangci-lint ein neues Hauptschema einführt oder ein Schwesterprojekt sein Profil wechselt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-18 | Accepted | Entscheidung des Owners im Gespräch |
