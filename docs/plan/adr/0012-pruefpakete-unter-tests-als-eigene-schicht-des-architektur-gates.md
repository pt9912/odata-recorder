# ADR-0012: Prüfpakete unter tests/ als eigene Schicht des Architektur-Gates

**Status:** Proposed

**Datum:** 2026-11-01

**Autor:** pt9912

**Bezug:** [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [ADR-0002](0002-architektur-gate-a-check.md)

**Schärft:** [`SPEC-039`](../../../harness/vertraege/SPEC-039-architektur-gate.md)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

[ADR-0002](0002-architektur-gate-a-check.md) legt die Schichten von `.a-check.yml` auf den Importgraph der Architektur-Sicht fest; alle liegen unter `src/`.
Der Durchstich-Test liegt unter `tests/integration/`, ruft den Composition Root von außen und nennt dazu Typen des Modells, der Inbound-Ports und `sqlite.Open`.
Ohne Regel für `tests/` ist `make a-check` rot, sobald ein Paket dort mehr als `bootstrap` importiert.
Der erste Ausweg, `tests/**` unter `composition_root` zu führen, behebt das und nimmt zugleich die Regel `tech` aus: ein Test dürfte `modernc.org/sqlite` und `encoding/xml` importieren, ohne dass das Gate etwas meldet (gemessen mit dem gepinnten a-check: 0 statt 2 Befunde `tech-leak`).
Das ist eine Lockerung nach `AGENTS.md` §3.6 und braucht eine Entscheidung.

## Entscheidung

Wir führen `tests/**` als **eigene Schicht `tests`** ohne Rolle, nicht als Composition Root.
Die Schicht nennt Modell, Inbound-Ports und Driven Adapter; den Composition Root nennt sie ohne Kante.
Die Regel `tech` gilt für sie wie für jede Schicht außerhalb des benannten Adapters.
Eine weitere Kante der Schicht (Services, Outbound-Ports, Driving Adapter) ist eine neue Entscheidung und kein Eintrag in einem Slice.
Die Einzelheiten (Kanten, gefangene Verstöße, Gegenproben) stehen in `SPEC-039`.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — `tests/**` unter `composition_root` | eine Zeile, jede Probe grün | nimmt `tech-leak` und alle Richtungsregeln für `tests/` aus: ein Test nennt Dienste, Outbound-Ports und Bibliotheken ungeprüft |
| **B — eigene Schicht `tests` mit drei Kanten (gewählt)** | `tech-leak` und Richtungen bleiben gefangen; der Durchstich-Test bleibt grün | drei Zeilen statt einer; die Kante auf `adapters-driven` gilt für alle Driven Adapter, nicht nur `sqlite` |
| C — Tests nennen nur `bootstrap` (Re-Exporte von `model`, `inbound` und `sqlite.Open`) | engste Fläche | Produktcode nur für einen Test |

## Konsequenzen

- Positiv: `tests/` steht unter einer Regel, die Bibliotheken und Richtungen hält; keine Regel für `src/` ändert sich.
- Negativ: ein Test, der einen Driven Adapter nennt, darf jeden nennen; die Bibliotheksgrenzen (`tech`) halten den Rest.
- Folgepflicht: drei Gegenproben zu `make a-check` (`SPEC-102`, Fälle 171 bis 173).
- Die Architektur-Sicht führt `tests/` nicht: es ist kein Bauteil des Hexagons, sondern Prüfcode.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `a-check` | `tests` nennt nur `model`, `ports-inbound`, `adapters-driven`; `tech-leak` gilt auch dort | `make a-check` |

## Re-Evaluierungs-Trigger

Ein Prüfpaket braucht eine Kante, die diese Entscheidung nicht trägt, oder a-check kennt eine Schicht ohne Rolle nicht mehr.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-01 | Proposed | Review-Befund F-2 zu `slice-m2-referenzdienst` |
