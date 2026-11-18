# Review-Report: slice-bootstrap-m4 — 2026-11-18

**Review-Art:** Code — geprüft wird der Diff `9b3a0c64..a8b36d7e` gegen Plan, Entscheidungen und Hard Rules.

**Gegenstand:** Zweig `slice-bootstrap-m4`, Stand `a8b36d7e` (Teil 2 von `welle-hybrid`: Modus `hybrid` im Composition Root, M4-Test, Stufe 10, Handbuch).

**Skill:** `.harness/skills/reviewer.md` · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-18

**Eingangs-Kontext:**

- Slice-Plan `slice-bootstrap-m4` (Tabelle *Zusage · Fänger* M01 bis M03, M10 bis M22, Raster in §6)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `LH-FA-12`, `LH-QA-02`; `SPEC-139` Nr. 9 und 12, `SPEC-284`, `SPEC-261`, `SPEC-037` (Nr. 2 bis 5, 7, 10 n), `SPEC-211` Nr. 2 und 3, `SPEC-110`, `SPEC-200`, `SPEC-201`
- `AGENTS.md` (Hard Rules, besonders §3.7, §3.9 bis §3.18)

Der Report entsteht in Blöcken und wird nach jedem Block committet; Abschnitte, die noch fehlen, stehen unter *Offen*.

---

## Findings

| Nr. | kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|---|
| F1 | MEDIUM | `AGENTS.md` §3.15, §3.17; `SPEC-139` Nr. 4 und 12 a | `src/bootstrap/wurzel.go` · `func (h *Handle) baueReplay(` | Der Replay-Dienst des Modus `hybrid` erhält `Pfade` (`REDACT_JSON_PATHS`) und `RedigierteBodies` (`REPLAY_REDACTED_BODIES`) vom Root; die Mutation, die sie nur im Modus `hybrid` verwirft (R1 `verdrahteHybrid` ruft `baueReplay` mit leerem `vorlaeufig`, R9 `ReplayRedactedBodies` aus, R11 `MaxCaptureBody` fest, R12 `v.pfade` nil), bleibt grün unter `TestM4*` und `TestWerteErreichenVerbraucher`. Keine Zeile der Tabelle (M02, M03) nennt diese Werte für `hybrid`; `TestWerteErreichenVerbraucher` prüft den Replay-Dienst nur im Modus `replay`. | ja | Zusage gilt nur für einen Teil der Modi |
| F2 | MEDIUM | `AGENTS.md` §3.15; `SPEC-139` Nr. 11 a, 12 a | `src/bootstrap/wurzel.go` · `StrictCoverage: k.StrictCoverage,` | Die Mutation `StrictCoverage` im Modus `hybrid` fest wahr (R8) bleibt grün. Die Fixture von `TestM4*` und `TestWerteErreichenVerbraucher` setzt `StrictCoverage` auf `true`, den Wert der Mutation. Das Handbuch sagt für `hybrid` zu: mit `false` gibt es für eine Collection-Abfrage über unvollständigen Bestand keinen Fallback. | ja | Fixture gleicht dem Wert der Mutation |

## Offen

(wird fortgeschrieben)

## Geprüft, ohne Befund

(wird fortgeschrieben)
