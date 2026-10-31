# Verifikations-Report: slice-guide-groessen-gate — 2026-10-31

**Rolle:** Verifier (Modul 11), einmal, frischer Kontext. Frage: Bauen wir es richtig, gegen Plan, DoD und Vertrag.

**Gegenstand:** Zweig `slice-guide-groessen-gate`, HEAD `5404023` (Implementer `3435f3f`, `556007f`, `6d68d6e`; Review `b923d61`; Fix-Runde `3661392`, `4ebdd1e`, `5404023`).

**Quellen:** Plan `docs/plan/planning/done/slice-guide-groessen-gate.md` (DoD, Tabelle Z1 bis Z15, Raster, §6, Marke 1800), [SPEC-040](../../harness/vertraege/SPEC-040-docs-check.md) Nr. 1 bis 11, [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) Fälle 156 bis 170, [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md) (auf `main` Accepted, im Zweig Proposed: Sache des Planners bei der Closure, kein Finding), [Review-Report](2026-10-31-slice-guide-groessen-gate-review.md), `AGENTS.md` §3.2, §3.6, §3.9 bis §3.18, `LH-QA-06`.

## Urteil

**Konform, kein Blocker.** Alle Sensoren sind selbst gefahren; die Fix-Runde deckt die Klasse. Offen bleiben ein LOW (Plan-Drift) und drei INFO, alle ohne Wirkung auf das Gate.

## Sensor-Belege (selbst gefahren, nicht aus dem Bericht übernommen)

| Sensor | Ergebnis |
|---|---|
| `make gates` im Worktree (`flock`, Ausgabe und Exit in Datei) | Exit 0 (`gates.exit`); `d-check: 515 Datei(en) geprüft, 0 Befund(e)`; `zeichen: ok`; `groessenmarke slice-guide-groessen-gate.md: ok, Messung 742 von Marke 1800` |
| `make gegenproben` mit dem committeten Stempel | "Eingaben unverändert seit dem letzten grünen Lauf (301c6afc…)", Exit 0 |
| `make gegenproben GEGENPROBEN_STEMPEL=<Scratchpad>` (erzwungener Lauf) | Exit 0, `gegenproben: 170 Fälle, 0 rot`, Schritt 81 s (Zeitbudget 120 s); der geschriebene Stempel ist byte-gleich dem committeten |
| Marke (f) | `git diff --numstat 3309cb4..HEAD` über die 8 Pfade des Blocks: 742 (Summe der Einzelzeilen nachgerechnet); Marke 1800 im Diff nicht geändert; Formel 135 × 1270 → 1800 stimmt |

**Stempel (g):** Der Stempel deckt den Zweig-Stand: Die Prüfsumme aus `gegenproben-eingabe.sh` mit den drei echten Pins ist gleich dem committeten Wert; `5404023` ändert nur den Plan, keinen Eingabepfad. Nach dem Merge von `main` ändert sich die Prüfsumme erneut (nur benannt, `main` trägt `ca6bb3f` und weitere Stände), der Planner zieht den Stempel dort nach.

## Stufe 1: DoD (h)

| DoD-Punkt | Befund |
|---|---|
| 1 `file`/`structure` aktiv, 13 Dateien in 5 Regeln, Grenzen mit Messung im Kommentar, `make docs-check` grün | wahr; `.d-check.yml` trägt die Messung je Regel; `docs-check` 0 Befunde |
| 2 Wirkung erprobt, Zusatzzeit gemessen | Haken offen im Plan; der Inhalt ist wahr (170 Fälle rot-grün gesehen, Zeitbudget in `SPEC-102`, §6 Punkt 4); Haken setzt der Planner |
| 3 Regel zum Anheben in [SPEC-040](../../harness/vertraege/SPEC-040-docs-check.md) Nr. 8 und [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md); Spalte *Vertrag* ≤ 200 Zeichen | wahr (Review Prüfpunkt 6: 117 und 200 Zeichen) |
| 4 `make gates` grün | wahr (Exit 0, siehe oben); **Wortlaut stale:** die Zeile nennt "Messung 737", gemessen sind jetzt 742 (LOW, unten) |
| 5 bis 9 Review, Doku-Update, Closure, Register, Risiken | offen oder Sache von Planner und Closure; Review-Report liegt vor; Risiken in §6 tragen Ausgang (zwei mit *offen* und Adresse) |

## Stufe 2: Tabelle gegen den Vertrag (a)

Z1 bis Z15 gegen [SPEC-040](../../harness/vertraege/SPEC-040-docs-check.md) Nr. 1 bis 11 und [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) gelesen, nicht gegen den Plan allein:

- Jede Aussage der Nr. 1 a, 2, 3, 4 a bis d, 5, 6 a bis e, 7, 9, 11 hat in der Tabelle einen Fänger oder ein benanntes Negativ (Z14: Nr. 8 e und 10 a; Z15: Nr. 10 b, c, g, h). Nr. 6 f, 6 g und 7 haben keinen eigenen Fall; Nr. 6 f steht als Negativ in 10 c, Nr. 7 und 6 g als Eigenschaft des Pins in 10 g. Grund und Folge stehen in Vertrag und Plan: vollständig.
- Die Menge "13 Dateien" stimmt mit dem Dateibaum: 1 + 3 + 6 + 2 + 1 = 13 (`wc` gegengerechnet).
- Die Fälle 156 bis 170 heißen in `guide.go` wie im Vertrag (Test `TestGuideFaelleHabenDieNamenAusDemVertrag`).
- **Abweichung Vertrag/Code (bereits in §6 mit Adresse):** Fall 166 und 167 nennen in SPEC-102 einen Abschnitt `### 3.1`, der Code zwei (`### 3.1`, `### 3.2`). §6 trägt die Zeile mit Ausgang *offen* und Adresse Architect im selben Commit wie der Code; der Code entscheidet nichts still, der Vertrag gilt bis zur Antwort. Korrekt adressiert.
- **INFO:** SPEC-102 Fall 167 sagt "jeder der 13 Pfade mit 50 000 Zeilen zu 16 Bytes"; der Stub von `AGENTS.md` trägt 799 974 Bytes (Abschnittsüberschrift). Dieselbe Stelle wie die Abweichung oben; Adresse Architect, ohne Wirkung auf das Gate.

## Stufe 3: Fix-Runde deckt die Klasse (b, c), `AGENTS.md` §3.17

**Befund MEDIUM (Pin in der Prüfsumme).** Menge aus dem Wortlaut: die Eingaben der Prüfsumme = 3 Pins + 15 vorhandene Pfade in `EINGABEN` (Plan §3 Fix-Runde).

- Pins, 3 Elemente: `gegenproben-eingabe.sh` selbst gefahren. 2 Argumente: Exit 1; 4 Argumente: Exit 1; jeder der drei Pins leer: Exit 1 (3 von 3, auch `A_CHECK_IMAGE`, das der Befund nicht nannte); jeder Pin mit anderem Wert: andere Prüfsumme (3 von 3). Unbenanntes Element: Pin aus einem Leerzeichen besteht die Prüfung (Exit 0, eigene Prüfsumme) — **INFO**, nicht erreichbar über `make` (die Pins sind digestgepinnte Konstanten), akzeptiert als Negativ, Folge keine.
- Pfade, 15 Elemente: akzeptiertes Negativ mit Grund und Folge im Plan (§3 Fix-Runde, Zeile 15) und in `SPEC-040` Nr. 10 h. Klasse gedeckt.
- Mutation "`"$(DCHECK_REF)"` aus `gegenproben.mk`" ist ein Exit 1 von `make gegenproben` (ein Argument weniger), kein Fall: benannt.

**Befund LOW (Schwelle 50 000/800 000 in Fall 167).** Menge: 5 Regeln × 2 Maße = 10 Grenzen. Selbst gefahren auf einer Kopie der echten Konfiguration (Binary `d-check` des Pins, Digest `d8ce395b…`, Testbinary im Image von golangci-lint, `--memory=4g --cpus=4 --network none`):

| Mutation | rote Fälle |
|---|---|
| Grundlinie (nur Kommentar angehängt) | 0 von 15 |
| jede der 10 Grenzen einzeln auf 50 000 Zeilen bzw. 800 000 Bytes (L1 bis L5, B1 bis B5) | 10 von 10 rot, je Fall 167; bei den drei Glob-Gruppen zusätzlich Fall 168 |
| Glob Agents `[a-o]*` (schließt planner, reviewer, validator, verifier aus; vom Implementer und Reviewer nicht genannt) | rot: 167, 168 |
| Glob Commands `[a-h]*` (schließt implement-slice, plan-welle aus; nicht genannt) | rot: 167, 168 |
| Glob Skills `[d-z]*` (schließt closure-note-reviewer aus; nicht genannt) | rot: 167, 168 |
| Hard Rule `max-lines: 50000` | rot: 167 |
| Hard Rule `sections: one` | rot: 166, 167, 168 |
| `harness/conventions.md` auf 799 974 Bytes (unterhalb des Stubs; Negativkontrolle) | 0 rot (richtig: die Grenze bleibt unter dem Stub, das Gate ist dort scharf) |

Der Stub von `AGENTS.md` (50 000 Zeilen, 799 974 Bytes) schwächt Fall 167 nicht: alle 10 Grenzen werden gefangen. Die Spec sagt "über 50 000 Zeilen oder 800 000 Bytes"; tatsächlich ist schon die Grenze **gleich** 50 000/800 000 rot (stärker als zugesagt), also wahr.

## Stufe 4: Gate nicht gelockert, Grenzen gleich Ist-Stand (d), `AGENTS.md` §3.6

Alle 13 Dateien neu gemessen (`wc -l`, `wc -c`): AGENTS 337/24 648; Commands max 221/16 992 (`implement-slice`); Agents max 53/3 759 (`architect`); Skills max 137/7 397 (`reviewer`); `conventions` 228/16 065. Jede Grenze in `.d-check.yml` ist gleich dem Maximum ihrer Gruppe; Summe 976 Zeilen und 68 861 Bytes. Die 13 Guide-Dateien sind im Diff `main...HEAD` unverändert; kein `hint`, kein `exempt-paths`, keine Direktive `nolint`, keine Anhebung gegenüber `main`. Der Spielraum der kleineren Gruppendateien ist entschieden (`SPEC-040` Nr. 4 a) und im Plan als Negativ mit Folge geführt.

## Stufe 5: §6, Adressen, nichts im Code entschieden (e), `AGENTS.md` §3.13

- Zeile "Abweichung beim Handoff" (Stub `### 3.1`/`### 3.2`): Ausgang *offen*, Adresse Architect, im selben Commit wie der Code.
- Zeile "Review-Befund MEDIUM": Pin ist im Code behoben (Skript); der Nachzug von `SPEC-040` Nr. 10 g/h und `SPEC-102` (Eingaben, Randform 10) steht als *offen*, Adresse Architect. Verifiziert: weder SPEC-040 noch SPEC-102 nennen die Arität (`git diff main...HEAD`), der Code entscheidet nicht anstelle der Spec, der Code gilt bis zur Antwort. Korrekt.
- Zeile "Review-Befund LOW": Ausgang *entfallen*, wahr (Stufe 3).
- Die Adresse "Architect" nimmt an: die Sendung steht im Plan dieses Slice als Rückgabe, kein anderer Slice ist Nehmer.

## Findings

| Stufe | Kat. | Befund | Folge |
|---|---|---|---|
| DoD | LOW | Plan Zeile 70 (DoD 4) sagt "Messung 737", `make groessenmarke` meldet 742 (die Fix-Runde fügte 5 Zeilen hinzu); `AGENTS.md` §3.9 verlangt den Nachzug im selben Commit | Planner zieht die Zahl bei der Closure nach; Marke 1800 fest, keine Wirkung |
| Vertrag | INFO | SPEC-102 Fall 167 "zu 16 Bytes" gilt nicht für den Stub von `AGENTS.md` (799 974 Bytes) | mit der offenen Zeile zu Fall 166/167 an den Architect |
| Fix | INFO | Pin aus einem Leerzeichen passiert `gegenproben-eingabe.sh` | nicht über `make` erreichbar (Pins sind Konstanten); akzeptiert |
| Stand | INFO | Stempel gehört zum Zweig-Stand; nach dem Merge von `main` neu zu setzen | Planner bei der Closure |
| ADR | — | Zweig Proposed, `main` Accepted | Sache des Planners, laut Auftrag kein Finding |

Keine Findings der Stufe HIGH. Kein falsches Codeverhalten gefunden.

## Rohbytes (§3.18)

Dieser Report enthält keine Escape-Folge als Text und keine Zeichen jenseits von Umlauten; gegengelesen mit `LC_ALL=C grep -P '\\u[0-9A-Fa-f]{4}'` (kein Treffer).

## Grenze dieser Verifikation

Pin-Mutationen anderer Art als die der Fälle 156 bis 165 (ein anderer Pin steht nicht zur Verfügung) und die Wirkung nach dem Merge mit `main` sind nicht gefahren. Alle Mutationen liefen auf Kopien im Scratchpad (`verify-guide/`); der Worktree-Code blieb unverändert.
