# Verifikations-Report: slice-telemetry-handler — 2026-11-07

**Review-Art:** Verifikation (gegen DoD, Spezifikation, Plan und Entscheidungen; nicht Diff-Review)

**Gegenstand:** Zweig `slice-telemetry-handler`, HEAD `915d805` (Fix-Runde nach dem Review `e6b5550`), Basis `a6b822b`. Teil H des geschnittenen Slice `slice-telemetry-log`.

**Rolle:** Verifier · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-07

**Eingang:** Plan `docs/plan/planning/done/slice-telemetry-handler.md` (§3 Entscheidungen 18 und 19, Tabelle H01 bis H16, Mengen M1 bis M17, Runden R1 bis R6, Absatz *Umsetzung* samt Fix-Runde, §6), `SPEC-231` Nr. 4 und 5, `SPEC-232` Nr. 2, 3, 5 c, d, f, g, das Review `docs/reviews/2026-11-07-slice-telemetry-handler-review.md`.
ADR: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Bezug des Slice), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (nur genannt: Entscheidung über den Trigger liegt beim User).

## Urteil

**Konform mit Auflagen.** Die DoD-Punkte, die der Implementer tragen kann (Handler, Senke/Sperre/Geheimnisse, `make gates`), sind belegt; die Tabelle *Zusage · Fänger* deckt jeden Satz von `SPEC-231` Nr. 4 und 5 und `SPEC-232` Nr. 2, 3, 5 c, d, f, g; die Fixes F1, F2, F5 halten gegen die Klasse. Gefunden ist eine Lücke in der Klasse „jede Art von Handler“ bei den Sperr-Fängern (V1, MEDIUM) und zwei kleinere Punkte. Nichts davon bricht eine Zusage im Code: der Code ist richtig, die Fänger sind an einer Stelle enger als die Zusage.

## Belege statt Behauptung

| Prüfung | Ergebnis |
|---|---|
| `make gates`, selbst gefahren (Log: `…/scratchpad/verif-th/gates.log`, unter `flock`) | Exit 0; a-check 0 Befunde; d-check 784 Dateien, 0 Befunde; baseline-verify OK; migration-check ok; zeichen-check ok (1591 Dateien); `make groessenmarke`: `slice-telemetry-handler.md: ok, Messung 1665 von Marke 2000`; Messpunkt `zeilen=16 nebenlaeufig=3 fakes=0 produktcode=120 schaetzung=1415 marke=2000 messung=1665`. Die Stufen `build`, `test`, `race`, `lint` meldeten `CACHED`: dieselben Eingaben liefen vorher grün, das Ergebnis gilt für den Inhalt von HEAD. Danach `git status` des Worktrees leer. |
| Größenmarke | `git diff --numstat a6b822b..HEAD` über die fünf Pfade: 84 + 498 + 129 + 262 + 692 = **1665** (Summe Einfügungen und Löschungen) gegen Marke 2000 und Schätzung 1415; Auslastung 0,83 (Band 0,60 bis 1,00, nichts an den Architect zu melden). Der Plan nennt 1665 nicht in Prosa; das Gate prüft es. |
| Unverändert | `git diff a6b822b..HEAD` nennt allein `handler.go`, vier `handler*_test.go`, den Plan und das Review. `zeile.go`, `regeln.go`, `telemetry.go`, `telemetry_test.go`, `regeln_test.go`, `audit_test.go`, `.golangci.yml`, `.a-check.yml`, `go.mod`, `go.sum`, `spec/`, `src/hexagon`, `src/bootstrap`, `harness/`, `tools/`: kein Eintrag. Das Binary blieb unverändert (Handler wird nicht verdrahtet; Handbuch-Punkt der DoD entfällt wie im Plan begründet). |
| `//nolint` | keine in `handler*.go` (`grep`). |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -P` über `handler*.go`, den Plan und das Review: keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, NBSP, U+2028/2029, Nullbreiten- oder Steuerzeichen. Nicht-ASCII-Bytes nur als Umlaute und `×` in Kommentaren. |
| Abdeckung selbst gemessen (Baumkopie per `docker cp`, Image `odata-recorder:race`, `-race -covermode=atomic`, `--memory=4g`, unter `flock`; Skript und Profile im Scratchpad `verif-th/`) | Gesamtlauf über das Paket: 98,9 %; **alle 21 Blöcke von `handler.go` mit Treffer**, kleinster Wert 8 (`WithAttrs` leer), null Blöcke ohne Treffer. Runden mit je einem Lauf `-run '^<Test>$'`: R1 `Enabled` 162, `Handle` 181, Z40 (`eintrag` unter `MinLevel`) 73; R2 `String` 32, `Int64` 28, `Uint64` 24 mit 16 im Innern und 8 darüber, `Float64` 32, `Duration` 24, übrige Arten 80; R3 `schluessel` 60, `wert` 16; R4 `WithAttrs` leer 8, nichtleer 704; R5 `WithGroup` leer 20, nichtleer 68; R6 `Handle` 168. **Jede Zahl des Berichts stimmt mit dem eigenen Lauf überein.** Der Block „Art außerhalb“ in R2 (80) und die Blöcke von R4 bis R6 werden von dem benannten Test selbst erreicht, nicht von einem anderen. |

## Tabelle Zusage · Fänger gegen den Vertrag

Satz für Satz gegen `SPEC-231` Nr. 4 und 5, `SPEC-232` Nr. 2, 3, 5 c, d, f, g gelesen:

| Satz | Zeile | Befund |
|---|---|---|
| Nr. 4 Attribute sind Schlüssel von `Log`; Beispielzeile `capture` | H01 | gedeckt |
| Nr. 5 `Enabled` folgt `MinLevel`, liest `ctx` nicht, ruft die Uhr nicht, zählt nichts | H14 (324 Aufrufe, 3 `MinLevel`), H07 | gedeckt; e1, e4, e5 rot (siehe unten) |
| `Handle` Text, Stufe, `msg` nach Nr. 3; ignoriert `Time` und `PC`; `MinLevel` selbst; immer `nil` (nil-`ctx`, beendetes `ctx`, Senke, Nullwert) | H04, H07, H08, H12, H16 | gedeckt |
| zuerst Schlüssel, dann Art; Gruppe ein Attribut; leerer Schlüssel; `LogValuer`, Panik; Arten; `Uint64`; `Duration` | H02, H03 | gedeckt; a1, a2, a3, l1 rot |
| `WithAttrs`/`WithGroup` neu, Empfänger unverändert, Kopie, geteilte Sperre/Senke/Stand | H05, H06, H15, H10 | gedeckt außer V1 |
| `WithAttrs` je Eintrag, Kette, doppelter Schlüssel, `LogValuer` einmal, leer | H05 | gedeckt; c1, c3, c5, c6 rot oder rot im Paket |
| `WithGroup` `name.`, auch vor spätere `WithAttrs`, mehrere, leer | H06, `TestWithGroupPunkt` | gedeckt; Verkettung ist als gleichwertig begründet (bestätigt, siehe unten) |
| 232 Nr. 2 kein Header, Body, Query, Konfigurationswert, kein Fehlertext | H09 (280) | gedeckt, vom Reviewer gegen die Klasse aufgelöst |
| 232 Nr. 3 Fehler der Senke, ein `Write` je Zeile | H08, H04, H13 | gedeckt |
| 232 Nr. 5 c, d, f, g | H10, H11, H13 | gedeckt für Wurzel und je eine Ableitung; **nicht für jede Art von Handler (V1)** |
| 232 Nr. 5 g letzter Satz (Zeile und Stand getrennt) | Negativ im Plan | Erlaubnis, keine Zusage; Grund und Folge stehen |

## Die Fixes gegen die Klasse (`AGENTS.md` §3.17)

Eigene Mutationen mit `tools/harness/mutationslauf.sh` (Container ohne Netz und Mount, `--memory=2g`, unter `flock`; Liste `verif-th/liste.json`, 25 Mutationen, alle an Elementen, die weder Befund noch Plan nannten; Last des Hosts 6 bis 10 von 4 CPUs, je Ergebniszeile `last=`).
Klasse `andere`: ein Prozess je Mutation. Klasse `nebenläufig`: 20 Prozesse mit `-race` (`-test.count=1`) und 20 ohne Detektor.

| Fix | Menge aus dem Wortlaut | Mutation an einem nicht genannten Element | Ergebnis |
|---|---|---|---|
| F1 Präfix `name.` | Trenner × Stelle × Länge der Namen | g1 Punkt nur ohne früheren Präfix; g2 Punkt nicht bei Namen der Länge 1; g7 Trenner `s`; g8 Trenner `d` (der Befund nannte ohne Punkt, der Fix `_` und `e`) | alle vier **rot** in `TestWithGroupPunkt` |
| F1 Verkettung (m22) | Reihenfolge, Verkettung, Name | g3 zweiter Punkt weg; g4 Reihenfolge vertauscht; g5 nur letzter Name; g6 Name entfällt | alle **grün** im ganzen Paket: bestätigt, dass die Verkettung keine beobachtbare Wirkung hat, solange der Punkt gehalten ist (Begründung des akzeptierten Negativs im Plan hält) |
| F2 Wiedereintritt | Form der geteilten Liste | w1 `attrs := h.vorgelegt[:0]` (andere Form als f05) | **rot** in `TestHandleWiedereintritt`, ein Prozess, ohne Detektor |
| F5 `Enabled`/`MinLevel` | Art des Handlers × Vergleich | e1 `MinLevel-1`; e4 immer wahr für Handler mit Präfix; e5 immer wahr für Handler mit vorgelegten Attributen | alle **rot** in `TestEnabled` |
| Kopier-Handler | Felder eines abgeleiteten Handlers | c1 Kind verliert Präfix; c5 `Handle` ohne Präfix; c6 `WithGroup` halbiert `vorgelegt`; c3 Reihenfolge der Attribute eines Aufrufs von `WithAttrs` umgekehrt | c1, c5, c6 **rot** in `TestWithGroup`; c3 **rot-paket** (`TestWithAttrs` grün, `TestHandlerDoppelterSchluessel` rot) |
| Art 0, Arten | Arten × Grenzen | a1 Obergrenze `MaxInt32`; a2 nicht abbildbare Art liefert `Str`; a3 `Duration`×2 | alle **rot** in `TestSlogArten` |
| `LogValuer` einmal | Aufrufstellen von `abbilden` | l1 `Handle` löst doppelt auf | **rot** in `TestSlogArten` (Zähler `LogValue`) |

## Befunde

### V1 · MEDIUM — Sperr-Fänger führen nicht jede Art von Handler

- `quelle`: `AGENTS.md` §3.15 (ein Fänger führt jede Form der Eingabe; Aufrufstellen sind eine Menge), §3.17; Plan M17 (4 Arten von Handlern) nennt H10 und H13 nicht
- `pfad`: `handler_sperre_test.go` · `TestHandlerKeineZweiWriteHaltepunkt`, `TestHandlerKeineZweiWriteMasse`, `TestBlockierteSenkeHandle`
- `befund`: `SPEC-232` Nr. 5 g: „Der Handler und jeder aus ihm abgeleitete teilen diese Sperre“; die Zähler warten nicht auf die Senke. H10 (1) und (2) schreiben über Wurzel, ein `WithAttrs`-Kind und ein `WithGroup`-Kind der Wurzel; H13 ruft `Handle` von Wurzel und `WithAttrs`-Kind. Die vierte Art (`WithAttrs` auf `WithGroup`) und jede tiefere Ableitung fehlen. Eigene Mutationen:
  - s2: `WithGroup` kopiert den Adapter samt Sperre, wenn der Empfänger ein Präfix hat; **gruen, auch das ganze Paket**.
  - s3: `WithAttrs` kopiert den Adapter, wenn der Empfänger ein Präfix hat; **gruen im ganzen Paket**.
  - z1: `Handle` nimmt und gibt die Sperre der Zeile vor `eintrag`, wenn der Handler ein Präfix hat (die Zählung wartet dann auf die blockierte Senke); **gruen im ganzen Paket**.
  - Kontrollen: z2 (dasselbe bei vorgelegten Attributen) **rot**, 20 von 20 mit und 20 von 20 ohne Detektor; s1 (Kopie bei vorgelegten Attributen) im Haltepunkt 0 von 20 und 0 von 20, im Paket erst durch eine Zeitüberschreitung von `TestHandlerAbleitenMasse` (2 min) **rot-paket**.
- `einordnung`: Der Code ist richtig (alle Handler tragen denselben `*Adapter`; kein Pfad legt eine Sperre je Art an). Die Lücke ist die Klasse der Zusage („jeder abgeleitete“), nicht ein Fehler. Eine Zusage ohne Fänger für zwei Arten ist ein MEDIUM nach §3.15.
- `Auflage` (Planner entscheidet, Fix-Runde oder Negativ): H10 (1) und H13 als Schleife über M17 (vier Arten, die Zählprüfung gegen `reflect.TypeFor[slog.Handler]` ist dort schon vorhanden); sonst als akzeptiertes Negativ mit Grund (ein einziger Zeiger auf den Adapter, keine Stelle im Code, die eine Sperre je Art anlegt) und Folge (eine künftige Art, die den Adapter kopiert, färbt keinen Test) in §6. Der Plan nennt `M17` für H10 und H13 bisher nicht.

### V2 · LOW — Mutation s1 wird vom Haltepunkt nicht gefangen

- `quelle`: Plan Antwort G (H10 (1) fängt „Sperre je Handler“ in jedem Lauf)
- `befund`: Kopiert nur der Handler mit vorgelegten Attributen den Adapter (s1), steht im Haltepunkt 0 von 20 rot (mit und ohne Detektor); das Paket wird nur durch das Hängen von `TestHandlerAbleitenMasse` rot. Dieselbe Klasse wie V1 (Ableitung zweiter Stufe); mit V1 erledigt.

### V3 · INFO — Sendung F3 liegt beim Planner

- `befund`: Der Plan trägt für die Texte von `upstream` und `ablage` (Orakel aus dem Code) die Zeile mit Ausgang *offen* und Adresse Planner (Eintrag auf `main` im Integrationsschritt: Sendung an den Architect und an `slice-telemetry-einbau-gateway`). Das ist eine Handlung des Planners auf `main`, nicht dieses Zweigs; sie ist vor `done/` zu erledigen. Der Widerspruch zwischen Absatz *Mengen* und §6 ist im Fix-Commit berichtigt.

## Bestätigt ohne Befund

- Plan-vs-Code: `handler.go` entspricht Entscheidung 18 und 19 (Abbildung vor `eintrag`, `Resolve`, Art 0 für den Rest, `WithAttrs` klont und bildet beim Ableiten ab, `Handle` baut je Aufruf neu, `WithGroup("")` und leeres `WithAttrs` liefern den Empfänger); `ctx` wird nirgends gelesen; der Handler führt keine eigene Prüfung der Regeln der Attribute.
- Kommentare in `handler.go` tragen Zusage, Kopplung oder Abgrenzung; keine Chronik (`AGENTS.md` §3.7, §3.11).
- Mutationen der Haltepunkt-Klasse: z2 (Zählung wartet für `WithAttrs`-Handler) 20 von 20 mit und 20 von 20 ohne Detektor rot, Messender dieser Report, Last 6 bis 9 von 4 CPUs; die Raten des Implementers (50 von 50) wurden damit nicht widerlegt, aber nicht in voller Zahl wiederholt.
- Plan §3, §4, §6 und die Zeilen `basis`, `schaetzung`, `marke` sind nach der Fix-Runde nachgezogen (`AGENTS.md` §3.9); die berichtigten Sätze in Antwort G, Fenster der Sperre und §6 stimmen untereinander überein.

## Auflagen an den Planner

1. V1 (und V2): Fix-Runde (H10 (1) und H13 über M17) oder akzeptiertes Negativ mit Grund und Folge in §6, bevor der Slice nach `done/` geht.
2. V3: Sendung zu den Texten von `upstream` und `ablage` auf `main` eintragen (Architect, `slice-telemetry-einbau-gateway`), dann den Ausgang in §6 schließen.
3. Closure: DoD-Haken, §7 (Messpunkt `Ende`, Raten H10, H11, H13 mit Zahl der Prozesse, Zusatzzeit der Stufen, Bericht der Abdeckung samt der hier bestätigten Zahlen, Binary unverändert), Lerneintrag, Register. Auslastung 0,83.
4. Die Entscheidung zum Trigger von [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Zusatzzeit der Stufe `race`, Paket rund 52 s im Gesamtlauf der Abdeckung) liegt beim User.

## Nicht geprüft

- Die 50-von-50-Raten des Implementers je Mutation (n01 bis n11) nicht in voller Zahl wiederholt; Stichprobe z2 und s1 bis s3 mit 20 und 20.
- Der Wortlaut der Texte von `upstream` und `ablage` gegen die Spezifikation (gibt es dort nicht, siehe V3).
