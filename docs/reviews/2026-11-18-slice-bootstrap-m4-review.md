# Review-Report: slice-bootstrap-m4 — 2026-11-18

**Review-Art:** Code — geprüft wird der Diff `9b3a0c64..a8b36d7e` gegen Plan, Entscheidungen und Hard Rules.

**Gegenstand:** Zweig `slice-bootstrap-m4`, Stand `a8b36d7e` (Teil 2 von `welle-hybrid`: Modus `hybrid` im Composition Root, M4-Test, Stufe 10, Handbuch).

**Skill:** `.harness/skills/reviewer.md` · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-18

**Eingangs-Kontext:**

- Slice-Plan `slice-bootstrap-m4` (Tabelle *Zusage · Fänger* M01 bis M03, M10 bis M22, Raster in §6)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `LH-FA-12`, `LH-QA-02`; `SPEC-139` Nr. 9 und 12, `SPEC-284`, `SPEC-261`, `SPEC-037` (Nr. 2 bis 5, 7, 10 n), `SPEC-211` Nr. 2 und 3, `SPEC-110`, `SPEC-200`, `SPEC-201`
- `AGENTS.md` (Hard Rules, besonders §3.7, §3.9 bis §3.18)

**Messung:** eigene Mutationen im Container ohne Netz (`--memory=4g --cpus=4`) unter `flock .wt/heavy.lock`, Load des Hosts 4 bis 5 (`/proc/loadavg`), Messender der Reviewer. Listen und Treiber: `docs/plan/planning/slice-bootstrap-m4-mutationen/` (`liste-m4rev-1.json` bis `liste-m4rev-3.json`). Die Stufe-10-Mutationen und die Handbuch-Mutationen liefen an einer Kopie des Baums außerhalb des Repos (`make test-integration`, `make handbuch-check`, `make katalog-check`).

---

## Findings

| Nr. | kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|---|
| F1 | MEDIUM | `AGENTS.md` §3.15, §3.17; `SPEC-139` Nr. 4 und 12 a | `src/bootstrap/wurzel.go` · `func (h *Handle) baueReplay(` | Der Replay-Dienst des Modus `hybrid` erhält `Pfade` (`REDACT_JSON_PATHS`), `RedigierteBodies` (`REPLAY_REDACTED_BODIES`) und `MaxBytes` vom Root. Vier Mutationen, die einen Wert nur im Modus `hybrid` verwerfen (R1 `baueReplay` mit leerem `vorlaeufig`, R9 `ReplayRedactedBodies` aus, R11 `MaxCaptureBody` fest, R12 `v.pfade` nil), bleiben grün unter `TestM4*` und `TestWerteErreichenVerbraucher` (Paket `src/bootstrap`) und unter `TestM4` (`tests/integration`, R1, R9, R11 grün); nur die Mutation an `Strategie` (R10) wird rot. Keine Zeile der Tabelle (M02, M03) nennt diese Werte für `hybrid`, `TestWerteErreichenVerbraucher` prüft den Replay-Dienst nur im Modus `replay`. | ja | Zusage gilt nur für einen Teil der Modi |
| F2 | MEDIUM | `AGENTS.md` §3.15; `SPEC-139` Nr. 11 a, 12 a | `src/bootstrap/wurzel.go` · `StrictCoverage: k.StrictCoverage,` | Die Mutation `StrictCoverage` im Modus `hybrid` fest wahr (R8, `StrictCoverage: k.StrictCoverage \|\| k.Modus == modusHybrid`) bleibt grün unter `TestM4*`, `TestWerteErreichenVerbraucher` und `TestM4` (`tests/integration`). Jede Fixture setzt `StrictCoverage` auf `true`, den Wert der Mutation. Das Handbuch sagt für `hybrid` zu, mit `false` gebe es für eine Collection-Abfrage über unvollständigen Bestand keinen Fallback; kein Fall führt `STRICT_COVERAGE=false` im Modus `hybrid`. | ja | Fixture gleicht dem Wert der Mutation |
| F3 | MEDIUM | `AGENTS.md` §3.10 | `docs/plan/planning/in-progress/slice-bootstrap-m4.md` · `*Datenrennen* auf einem einfachen Zähler und auf der Liste der Requests ohne Sperre: nicht ausgewertet` | §3 nennt die Raten der Datenrennen auf den Zählern des Referenzdienstes nicht („nicht ausgewertet“), und der Absatz trägt kein akzeptiertes Negativ für sie. Gemessen (N = 20 Prozesse `-test.count=1` mit Detektor, `ratedrv.sh`, `liste-m4s-rate2.json`, Load 4 bis 5): `requests` plain M4 19 von 20, Selbsttest 20 von 20; `verbindungen` plain M4 20 von 20, Selbsttest **0 von 20**; Zähler je Methode plain M4 **17 von 20**, Selbsttest **19 von 20**; ohne Detektor je 0 von 20. Die Zusage „Zähler atomar“ ist damit nach `AGENTS.md` §3.10 nicht gefangen (nicht in jedem Lauf rot); ob die Vereinigung beider Tests in einem Prozess 20 von 20 erreicht, ist nicht gemessen. | ja | Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt |
| F4 | LOW | `AGENTS.md` §3.9, §3.13 | `harness/vertraege/SPEC-037-test-integration.md` · `die Szenarien über Admin-Listener und Verbindungszähler kommen als weitere Stufen` | Der Kopfsatz der Festlegung nennt den Verbindungszähler weiter als künftige Stufe, obwohl `TestE2EHybrid` (Stufe 10) ihn liefert; derselbe Commit setzt den Satz *Stand* auf „zehn Stufen“ und lässt den Kopfsatz stehen. Die Änderung des Satzes *Stand* in einem Vertrag des Architect trägt in §6 den Ausgang „entschieden“ ohne Adresse oder Bestätigung des Architect. | ja | Berichtigung nicht an allen Stellen nachgezogen |
| F5 | LOW | `AGENTS.md` §3.15 | `docs/plan/planning/in-progress/slice-bootstrap-m4.md` · `Stufe 10 fängt nur Fehler der Konfigurationsprüfung, akzeptiertes Negativ` | Das akzeptierte Negativ nennt keine Folge im Betrieb. Gemessen an einer Kopie (`make test-integration`): `Start` liefert bei Fehler von `LadeOderMelde` Exit 0 gibt Stufe 1 und Stufe 10 rot; `Start` liefert bei Fehler des Baus Exit 0 gibt allein Stufe 4 (`TestE2EPersistenz`) rot, Stufe 10 grün, wie der Plan sagt. Die Zeile M10 nennt als Mutation „Exit 0 bei Fehler“ ohne die Unterscheidung. | ja | Akzeptiertes Negativ ohne Folge |

## Offen

(wird fortgeschrieben)

## Geprüft, ohne Befund

(wird fortgeschrieben)
