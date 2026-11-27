# Carveouts — odata-recorder

Aktive Carveouts mit Auflösungs-Trigger. Aufgelöste Carveouts wandern nach
`done/` (reiner `git mv`). Diese Liste ist abgeleitet; maßgeblich sind die
Carveout-Dateien.

## Aktive Carveouts

| ID | Titel | Gate | Trigger | Folge-Slice |
|---|---|---|---|---|
| [CO-001](CO-001-lasttest-replay-ohne-p95.md) | Replay ohne p95, das Binary endet an der Speichergrenze | `make lasttest` | Aufruf `abnahme` mit `abnahme=ja` | [slice-replay-speicher-und-suche](../planning/open/slice-replay-speicher-und-suche.md) |

## Aufgelöste Carveouts

(noch keine)

## Konventionen

- Jeder aktive Carveout braucht: Trigger, Folge-Slice, letzten Prüf-Termin.
- Bei Welle-Closure: Carveout-Audit zwingend — welche gültig, welche aufgelöst?
- Siehe Baseline-Regelwerk `modul-07-carveouts.md`.
