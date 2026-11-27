# Carveouts — odata-recorder

Aktive Carveouts mit Auflösungs-Trigger. Aufgelöste Carveouts wandern nach
`done/` (reiner `git mv`). Diese Liste ist abgeleitet; maßgeblich sind die
Carveout-Dateien.

## Aktive Carveouts

(keine)

## Aufgelöste Carveouts

| ID | Titel | Gate | Aufgelöst durch | Folge-Slice |
|---|---|---|---|---|
| [CO-001](CO-001-lasttest-replay-ohne-p95.md) | Replay ohne p95, das Binary endet an der Speichergrenze | `make lasttest` | Aufruf `abnahme` mit `abnahme=ja` auf `9c3f5f10` (2026-11-27) | [slice-replay-speicher-und-suche](../planning/in-progress/slice-replay-speicher-und-suche.md) |

## Konventionen

- Jeder aktive Carveout braucht: Trigger, Folge-Slice, letzten Prüf-Termin.
- Bei Welle-Closure: Carveout-Audit zwingend — welche gültig, welche aufgelöst?
- Siehe Baseline-Regelwerk `modul-07-carveouts.md`.
