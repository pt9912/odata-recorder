# Review-Report: slice-guide-groessen-gate — 2026-10-31

**Review-Art:** Code (gegen Plan, Vertrag und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-guide-groessen-gate`, Diff `4e36c01..6d68d6e` (Code-Commits `3435f3f`, `556007f`, `6d68d6e`; Basis `git merge-base HEAD main`).

**Skill:** `.harness/skills/reviewer.md` @ `6d68d6e` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Eingangs-Kontext:**

- `slice-guide-groessen-gate` (Tabelle Zusage · Fänger Nr. 1 bis 15, Raster, §6, Marke 1800, Messung 737)
- [SPEC-040](../../harness/vertraege/SPEC-040-docs-check.md) Nr. 1 bis 11, [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) Fälle 156 bis 170
- [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md) (auf `main` Accepted), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)
- `LH-QA-06`; `AGENTS.md` §3.2, §3.6, §3.7, §3.9 bis §3.18

**Vorgehen.** Alle Mutationen liefen auf Kopien im Scratchpad (`review-guide/`), der Worktree blieb unverändert. Das Binary `d-check` stammt aus dem Image des Pins (`v0.90.0`, Digest aus `d-check.mk`). Die 15 Fälle liefen als Testbinary im Image von golangci-lint (`--memory=4g --cpus=4 --network none`) gegen je eine Mutation. Pin-Mutationen sind Ersatz-Programme vor dem echten Binary, wie im Plan. Mutationen an der Konfiguration wurden neu gebildet, nicht aus dem Plan übernommen (Ausgabe `out.txt`, `out2.txt`).

## Auflösung der Tabelle (je Zeile eine eigene Mutation; rot = Fälle, die rot wurden)

| Z | Mutation (eigene, auch an Elementen, die der Implementer nicht nannte) | rote Fälle |
|---|---|---|
| Grundlinie | keine | keiner (0 von 15) |
| 1 | `file` aus `modules:`; `structure` aus `modules:`; Block `file:` gelöscht; Eintrag `AGENTS.md` in `structure:` gelöscht | 170; 170; 167, 168; 167, 169 |
| 2 (13 Pfade × 2 Codes, 10 Regel-Mutationen) | je Regel `max-lines` bzw. `max-bytes` entfernt (5 × 2); Glob Agents `[a-u]*` (schließt `validator.md` und `verifier.md` aus); Glob Skills `[a-b]*`; Glob Commands `[a-i]*`; Glob `harness/convention.md` | 167 (bzw. 167, 168) je Mutation; 167, 168; 166 bis 169; 167, 168; 166 bis 169 |
| 3 | `.claude/commands/*.md` durch drei Regeln mit den Dateinamen ersetzt | 167, 168 |
| 4 | Pin-Ersatz `>=` (Grenzen minus 1); Maße vertauscht; Bytes nicht geprüft; Zeilen in den Regeln nicht geprüft | 156 bis 161, 163, 164; 13 von 15; 157, 159, 160, 163, 167, 168; 158, 159, 161, 164, 167, 168 |
| 5 | Pin zählt Zeichen | 160 |
| 6 | Pin zählt nur Zeilenvorschübe | 161 |
| 7 | nur der erste Befund je Datei; nur der erste je Regel | 159, 167, 168 je Mutation |
| 8 | Pin ohne `file-no-match` | 162 |
| 9 | `.harness/` ausgeblendet; `.claude/` ausgeblendet | 10 Fälle; 11 Fälle |
| 10 | `hint` in einer Regel | 167, 168 |
| 11 | `sections: one`; Muster `3\.1 `; Muster `3\.\d ` (einstellig); Muster `3\.[12] `; `max-lines` entfernt | 166, 167, 168; 167; 167; 167; 167 |
| 12 | Pin ohne `section-missing` | 165, 169 |
| 13 | `max-lines: 0` in der Regel der Commands | 166, 168, 169 |
| 14 (Negativ) | AGENTS +100 Bytes; 900000 Bytes; 49999 und 50000 Zeilen; Agents 799999 Bytes; Hard Rule 30 | keiner; 167; keiner; keiner; keiner |
| 15 (Negativ) | Binary fehlt; `.d-check.yml` aus `EINGABEN` | alle 15; Prüfsumme gleich (ohne Eintrag), anders (mit Eintrag) |

Gegen die echte Konfiguration auf einer Kopie des Baums, mit dem Binary des Pins: `reviewer.md` +1 Zeile, `conventions.md` +1 Byte, `AGENTS.md` +1 Zeile am Ende, §3.7 +1 Zeile (Befund `section-lines-exceeded` auf Zeile 138) und eine neue 61-Zeilen-Datei in `.claude/agents/` werden rot. Die Rollen-Typen und der Skill in `.harness/skills/` werden trotz `scan.ignore` erreicht. Die Zahlen des Plans (Absatz *Mutationen und Raten*) stimmen mit meiner Messung überein, soweit sie dieselben Mutationen nennen. Grundlinie: `make docs-check` im Worktree grün (514 Dateien, 0 Befunde). Messung der Marke: Summe `git diff --numstat` über die Pfade des Blocks gleich 737.

## Findings

| kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|
| MEDIUM | `AGENTS.md` §3.15 | `harness/mk/gegenproben.mk` · `"$(DCHECK_REF)")` | Der Pin von d-check in der Prüfsumme ist zugesagt ([SPEC-040](../../harness/vertraege/SPEC-040-docs-check.md) Nr. 10 g, [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) Eingaben und Grenze), steht aber weder als Fall noch im akzeptierten Negativ Z15 oder Nr. 10 h (die nennen Kopie des Binary und `.d-check.yml`). Wird das dritte Argument aus dem Aufruf von `gegenproben-eingabe.sh` entfernt, wird kein Fall und kein Gate rot; ein Pin-Wechsel liefe dann ohne neuen Lauf. | nein (Prüfsumme geändert: anderer dritter Pin gibt anderen Wert, Aufruf im Makefile ungeprüft) | Zusage ohne Fänger und ohne benanntes Negativ |
| LOW | `AGENTS.md` §3.11 | `harness/vertraege/SPEC-040-docs-check.md` · `eine Grenze über 50 000 Zeilen oder 800 000 Bytes macht ihn rot` | Für `AGENTS.md` stimmt die Schwelle nicht: Der Stub trägt 50003 Zeilen und 800022 Bytes. Eine Grenze von 50000 Zeilen oder 799999 Bytes bleibt in Fall 167 grün (gemessen: 49999 und 50000 Zeilen, 799999 Bytes keiner rot); für die übrigen zwölf Pfade stimmt sie. | ja | Schwelle in der Zusage ungenau |
| INFO | `AGENTS.md` §3.6 | `.d-check.yml` · `max-lines: 29` | Die Grenze der Hard Rules (29, Ist-Stand §3.7) anzuheben bleibt in allen 15 Fällen grün (gemessen mit 30); das deckt Nr. 10 a allgemein ("Höhe der Grenze"), Z14 nennt aber nur die Grenze von `AGENTS.md` als Gegenstand. Kein Verstoß; Zusatz für den Reviewer am Diff von `.d-check.yml`. | nein | Höhe der Grenze ungeprüft (akzeptiert) |
| INFO | Maintainability | `.d-check.yml` · `Jede Grenze ist das Maximum ihrer Dateien` | Die Grenze ist nur für die größte Datei jeder Gruppe gleich dem Ist-Stand (so in [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md) und Nr. 4 a entschieden). Die kleineren Dateien haben Spielraum, gemessen: `closure-note-reviewer.md` +1 Byte, `verifier.md` +1 Zeile, `plan-welle.md` +1 Byte bleiben grün (rund 300 Zeilen und 25 KB über die drei Gruppen, ohne `AGENTS.md` und `conventions.md`). Der Ausdruck "frieren den Ist-Stand ein" im Plan §1 gilt damit je Gruppe, nicht je Datei. | ja | Gruppenmaximum lässt Spielraum (entschieden) |
| INFO | `AGENTS.md` §3.9 | `docs/plan/adr/README.md` · `Proposed` | Der Zweig führt [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md) und die Index-Zeile als Proposed, auf `main` ist sie Accepted; Plan §1, §2 und §6 (offener Punkt 1) sagen weiter "Proposed, Annahme beim Nutzer". Beim Merge ist die Fassung von `main` zu nehmen und der Plan nachzuziehen; Adresse: Planner (Closure). | ja | Stand der ADR zwischen Zweig und `main` |
| INFO | `AGENTS.md` §3.15 | `tools/gegenproben/guide.go` · `if haben["file"] && haben["structure"]` | Fall 170 prüft Wörter im Rest der Zeile `modules:`; ein Kommentar derselben Zeile mit `file structure` erfüllt ihn, ohne dass die Module aktiv sind. Wirkt nur bei einem Kommentar an dieser Zeile; Fall 167 bleibt der Fänger für den Block. | ja | Wortprüfung statt Modullesung |

Keine Findings der Stufe HIGH.

## Prüfpunkte der Anweisung

1. **Gate-Lockerung (§3.6):** keine. Grenzen sind die gemessenen Ist-Werte aller 13 Dateien (`wc -l`, `wc -c` gegengerechnet; Summe 976 Zeilen und 68861 Bytes stimmt), keine Direktive `nolint`, kein `exempt-paths`, kein `hint`, kein Anheben gegenüber `main` (die Guide-Dateien selbst sind im Diff unverändert). Der Spielraum der kleineren Gruppendateien steht als INFO.
2. **`.d-check.yml`:** `file` und `structure` fassen die genannten Dateien, auch unter `.claude/` und `.harness/skills/` (Läufe R1, R3 und R9; `scan.ignore` wirkt nicht). Symlinks werden nicht gemessen (neuer Symlink in `.claude/commands/` bleibt grün): das ist Nr. 10 c, benannt mit Grund und Folge; kein Finding.
3. **Dockerfile, `.dockerignore`, `gegenproben.mk`:** Das Binary kommt als `/d-check` aus dem Image des Pins (Digest); kein Stage außer `gegenproben` zieht `D_CHECK_IMAGE`, kein Stage kopiert den ganzen Kontext, daher ändert `!.d-check.yml` keinen Cache der anderen Stufen. Die Prüfsumme ändert sich mit `.d-check.yml` (auch mit einem Kommentar: Wert `529cfbe9…` zu `0170e5a4…`) und mit dem dritten Pin (`ce42fd7f…`); ohne den Eintrag in `EINGABEN` bleibt sie gleich (`a9ffeeac…` mit und ohne Änderung). Fehlendes Argument im Makefile: siehe MEDIUM.
4. **Offene Punkte:** Die Abweichung der Stubs von Fall 166 und 167 (zwei Abschnitte statt eines) steht in §6 mit Ausgang *offen* und Adresse Architect im selben Commit wie der Code (`3435f3f`), wie `AGENTS.md` §3.13 es verlangt; [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) nennt weiter einen Abschnitt, bis der Architect entscheidet. Korrekt adressiert; kein Finding.
5. **Kommentare und Köpfe (§3.11):** Köpfe von `guide.go`, `.d-check.yml`, `gegenproben.mk` und `Dockerfile` sagen zu, was die Fälle prüfen; die Aussage "Zeichen jenseits von ASCII führt nur Fall 160" stimmt (Rohbytes in `guide.go` und `guide_test.go` gegengelesen mit `LC_ALL=C grep -P`, keine Escape-Folge als Text). Nur die Schwelle der LOW-Zeile ist ungenau.
6. **`harness/README.md`:** Spalte *Vertrag* der Zeile `make docs-check` hat 117 Zeichen, Spalte *Tut was* der Zeile `make gegenproben` genau 200 (Grenze 200); ein Test mit 218 Zeichen wird mit `section-cell-oversized` rot. Die Zeilen sind wahr: `docs-check` fährt `file` und `structure`, `gegenproben` fährt die Fälle 156 bis 170. *Nicht behauptet* nennt Nr. 10 d und 10 h, aber nicht den Pin in der Prüfsumme (siehe MEDIUM).
7. **Negative Z14 und Z15:** Z14 hat Grund und Folge im Plan und in Nr. 10 a, Z15 in Nr. 10 b, c, g, h; beide sind vollständig bis auf den Pin (MEDIUM).

## Negativbefund (geprüft, ohne Befund)

- `.d-check.yml`, Kommentare und Messwerte: geprüft, ohne Befund außer INFO.
- `Dockerfile`, `.dockerignore`: geprüft, ohne Befund.
- `harness/mk/` und `tools/harness/`: geprüft, Befund nur im MEDIUM.
- `tools/gegenproben/` (`guide.go`, `guide_test.go`, `faelle.go`, `main.go`, `main_test.go`, `strom_test.go`): geprüft, ohne Befund außer INFO (Fall 170); Zahl 133 der getrennten Fälle stimmt (37 + 133 = 170).
- `harness/vertraege/` (SPEC-040, SPEC-102): geprüft, Befund nur LOW.
- `harness/README.md`, `docs/plan/adr/`: geprüft, ohne Befund außer INFO (Stand der ADR).
- `AGENTS.md`, `.claude/`, `.harness/skills/`, `harness/conventions.md`: unverändert im Diff (geprüft).
- Marke: Messung 737 (nachgerechnet) unter der festen Marke 1800; keine Änderung der Marke im Diff.
