# Verifikation slice-groessenformel-neu (Text-Slice, Beschluss V0)

Verifier, frischer Kontext, einmal verifiziert. Stand HEAD `4e7d18eb`, Basis `ec1ab4ef`, Worktree `odata-recorder-wt-groessenformel` (vor und nach der Prüfung `git status` leer). Diff: `harness/vertraege/SPEC-128-groessenmarke.md`, `docs/plan/planning/done/slice-groessenformel-neu.md`, `docs/reviews/2026-11-20-slice-groessenformel-neu-review.md`.
Gegen: DoD des Plans, Beschluss des Users vom 2026-11-20 (F1 bis F5), `SPEC-128`, `AGENTS.md` §3.6, §3.7, §3.9, §3.13, §3.15, §3.18, [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md), [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md), [MR-001](../../harness/conventions/MR-001-harness-vertraege-ausserhalb-der-spezifikation.md) und die Anker-Paarung in `.harness/baseline/v6.18.1/regelwerk/grundlagen-traceability.md`.
Alle Mutationen liefen auf Kopien im Scratchpad (`groessenformel-verifier/m-*`, `clone`); der Worktree ist unverändert.

## Urteil

**Freigegeben: ja, mit Auflagen für die Closure.** Kein HIGH, kein MEDIUM. Vier LOW (L1 bis L4) und drei INFO; keiner verlangt eine weitere Prüfrunde, die Auflagen stehen als Schritte unten (Abschnitt *Was der Planner in der Closure tut*).

## Sensoren

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` unter `flock`, Stand `4e7d18eb` | grün, Exit 0 | `groessenformel-verifier/gates.log`; darin `baseline-verify: v6.18.1 OK`, `d-check: 972 Datei(en) geprüft, 0 Befund(e)`, `zeichen: ok, 2007 Datei(en)`, `groessenmarke slice-groessenformel-neu.md: Hinweis: kein Block`, Exit 0 |
| `make docs-check` (einzeln) | Exit 0, 972 Dateien, 0 Befunde | `groessenformel-verifier/dc.log` |
| `make zeichen-check` (einzeln) | Exit 0 | `zeichen: ok, 2007 Datei(en)` |
| Rohbytes (`LC_ALL=C grep -P` auf BOM, NBSP, U+200B bis U+200F, U+2028 bis U+202E, U+2060 bis U+206F, Steuerzeichen, Backslash-u-Folge) | 0 Treffer in `SPEC-128`, Plan, Review-Report | gegengelesen: alle drei geänderten Dateien |
| `git diff --stat ec1ab4ef..HEAD` über die Zeile `EINGABEN` von `tools/harness/gegenproben-eingabe.sh` | leer; Stempel `harness/gegenproben.stempel` bleibt gültig, `make gegenproben` nicht zu fahren | Klon-Lauf, Abschnitt Zeile 2 |

Die Behauptung des Plans (§3 *Zuschnitt* Nr. 3), `make gates` laufe im Worktree vor dem Review, trägt keinen Beleg im Plan; ich habe den Lauf selbst gefahren (grün).

## Plan gegen Diff

- DoD Punkt 1: `SPEC-128` Nr. 1 byte-gleich (der erste Hunk des Diffs beginnt in Nr. 2, Zeile 41); Nr. 2 mit Herkunft und Geltungsgrenze, Nr. 9 mit 0,45 und Spanne, Nr. 10 a (vier Negative) und 10 k geliefert. Der Beschluss F1 (V0), F2 (0,45), F3 (keine zweite Klasse), F4 (nur Spanne, kein Auslöser), F5 (entfällt) steht so im Text; Nr. 9 führt keinen Auslöser beim Handoff (`Handoff und Ende`: "löst keine Meldung aus").
- DoD Punkt 2 (keine Eingabe der Gegenproben): bestätigt, siehe oben.
- DoD Punkt 5 (Doku-Update): `SPEC-102` und `harness/README.md` nicht im Diff; die Zahl 0,60 steht außerhalb von `done/`, `docs/reviews/` und den Beleg-Dateien des Registers nur in `SPEC-128` (als Vergleichszahl) und im Plan (geprüft mit `grep -rn` über `.claude`, `.harness/skills`, `AGENTS.md`, `harness/README.md`, `spec`, `tools`, `open/`, `next/`, Roadmap: kein Träger der Schwelle außer `SPEC-128`).
- DoD Punkte 3 bis 4, 6 bis 9: noch nicht abgehakt, das ist die Closure.
- `AGENTS.md` §3.9: Kopf (`Bezug`, `Berührte Spec-Stellen`), §1, §3, §4, §6 tragen den Beschluss; kein Rest "neue Faktoren in Programm und Gegenproben" im Plan (er steht noch in Roadmap-Zeile und `state.md`, beides Sache des Integrationsschritts, in §6 mit Adresse).
- `AGENTS.md` §3.6: Nr. 9 ist keine Schwelle eines Gates (kein Ziel, kein Programm, kein Fall der Gegenproben, kein Command führt 0,60 oder 0,45); [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) trägt keine Zahl der Formel oder der Schwelle (`grep` ohne Treffer), Status `Proposed`. Das Verdikt des Architect (greift nicht) trägt; `make groessenmarke` ist im Gate-Lauf grün und rechnet unverändert.
- `AGENTS.md` §3.7: die neuen Absätze beschreiben den Zustand; Herkunft steht als `seit slice-groessenformel-neu`, die Stand-Zelle als Datum. Kein Konjunktiv über verworfene Fassungen außer den benannten Alternativen V1 und V2 in Nr. 2, die dort als Herkunft der Faktoren gehören.
- `AGENTS.md` §3.18: Rohbytes siehe oben; Zeile 4 wurde aus Bytewerten gefahren.
- [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md): `SPEC-128` und der Plan liegen nicht unter den Größengrenzen der Guide-Dateien (`make docs-check` grün).

## Zusage und Fänger aufgelöst (eigene Mutationen)

Methode: Kopie von `tools/groessenmarke` in `groessenformel-verifier/m-<name>`, `go test ./...` im Image `odata-recorder:deps` (`--network none --memory=4g --memory-swap=4g --cpus=4`, unter `flock`). Unmutiert: `ok`.

| Zeile | Mutation (Kopie) | Ergebnis |
|---|---|---|
| 1, Element nicht genannt | `faktorAndere` 80 auf 81 (`z1a`) | rot |
| 1, nicht genannt | `faktorNebenlaeufig` 85 auf 84 (`z1b`) | rot |
| 1, nicht genannt | `faktorFake` 200 auf 199 (`z1c`) | rot |
| 1, Aufrundung, gleicher Wert | `+ 9999` auf `+ 10000` (`z1g`, S = 2000 gibt 2700 statt 2800) | rot |
| 1, Schnittgrenze, Gleichheit | `schnittGrenze` 300000 auf 300105 (`z1h`, S = 2223 liegt genau darauf) und auf 300135 (`z1f`) | rot, rot |
| 1, äquivalent | `+ 9999` auf `+ 9998` (`z1d`); `>` auf `>=` (`z1e`); `schnittGrenze` 299970 (`z1i`) | grün; äquivalent: 135·S ist stets ein Vielfaches von 5, 135·S = 300000 hat keine ganzzahlige Lösung, und 135·2222 = 299970 liegt auf der Grenze bei strengem Vergleich |
| 2 | `faktorAndere` 81 in `formel.go`, committet im Klon | `git diff --stat ec1ab4ef..HEAD` über `EINGABEN`: 1 Datei, 1+/1-; ohne Mutation 0 Zeilen |
| 2, Gegenrichtung | `Zuschlag 1,35` auf `1,36` in `SPEC-128`, committet im Klon | Diff über `EINGABEN`: 0 Zeilen (die Zusage "`harness/vertraege/` gehört nicht zu den Eingaben" trägt; kein Code liest `SPEC-128`, nur Pfad-Ausschlüsse `harness/vertraege` in `groessenmarke.sh`, `plan.go`) |
| 3 | Zeile 84 `hinweis(stderr, name, "kein Block ...")` entfällt (`z3a`) | rot: `TestM10StroemeDesProgramms` ("abfrage Hinweis kein Block": "Hinweis: kein Block" fehlt im Strom stderr) und `TestM3ProsaZeile`, nicht Aufbau |
| 3 | stdout `keine-messung` auf `messen` (`z3d`) | rot |
| 3 | Exit nach `keine-messung` auf 1 (`z3e`) | rot |
| 3, Nachsatz | `der Plan wird nicht gemessen` auf `der Plan wird gemessen` (`z3b`) | grün; INFO 1 |
| 4 | NBSP aus Bytewerten (`printf '\302\240'`, `awk`) in Nr. 9 von `SPEC-128` im Klon, gegengelesen (1 Treffer) | `make zeichen-check`-Treiber: Exit 1, `SPEC-128-groessenmarke.md:125:64: U+00A0 Zs NBSP`; unmutiert Exit 0 |
| 4, nicht genannt | BOM am Dateianfang von `SPEC-128` | Exit 1, `1:1: U+FEFF Cf BOM` |
| 4, nicht genannt | U+200B im Plan (Zeile 14) | Exit 1, `14:5: U+200B Cf ZWSP` |
| 4a | Link auf den Plan in Nr. 2 auf `slice-gibt-es-nicht.md` | `make docs-check` Exit 2: `SPEC-128-groessenmarke.md:47 ... target-missing` |
| 4a, Anker | alle Vorkommen von `seit slice-groessenformel-neu` in `SPEC-128` entfernt | `make docs-check` Exit 0, 0 Befunde (der Anker hat keinen Fänger, wie der Plan sagt; akzeptiertes Negativ, Grund und Folge stehen in Zeile 4a und §6) |
| 4a, Move | `make slice-mv SLICE=slice-groessenformel-neu TO=done` im Klon | `rc=0`, zwei Commits (reiner Move, Verweis-Nachzug: 4 Dateien eingehend, 0 ausgehend); beide Links in `SPEC-128` (Nr. 2 und 9) zeigen danach auf `../../docs/plan/planning/done/slice-groessenformel-neu.md`, ebenso Roadmap und Review-Report; `make docs-check` danach: nur der Befund `planning-drift` der Roadmap (siehe Closure), kein toter Link |

Die Menge der Zeile 1 ist die Formel der Nr. 1 (drei Faktoren, Zuschlag, Aufrundung, Schnittgrenze und beide Seiten der Grenze): jedes nicht äquivalente Element ist rot.

### Zeile 5 bis 7 und die Zahlen

Nachgerechnet aus den 62 Zeilen der Tabelle in §3 des Plans (`awk` auf Kopie, ungerundete Werte, Rang-Perzentile nach Nr. 2):

- *a* = (*S* − 80·*z* − 200·*f* − *p*) ÷ 5 ist in allen 62 Zeilen eine ganze Zahl zwischen 0 und *z*; *M* = 100·⌈135·*S*/10000⌉ stimmt in allen 62 Zeilen; *E* ÷ *S*, *E* ÷ *M*, *E* ÷ *H* der Tabelle stimmen in allen Zeilen auf zwei Stellen.
- Klassen: 60 `go`, 1 `skript`, 1 `handlauf` (62); n = 61 ohne den Handlauf.
- *E* ÷ *S*: Median 0,826 (Rang 31), p10 0,642 (Rang 7), p90 1,205 (Rang 55); 49 von 61 unter 1; p90 ÷ p10 = 1,878.
- *E* ÷ *M*: Median 0,599, p10 0,472, p90 0,853; unter 0,60: 31; unter 0,45: 5 (`test-determinism`, `mock-schreiben`, `mock-abnahme`, `gegenproben-zeitbudget`, `guide-groessen-gate`); über 1,00: 2 (`telemetry-port` 1,031, `bootstrap-m4-konfig` 1,921); 59 von 61 gehalten. Meldungen bei 0,45: 5 + 2 = 7; bei 0,60: 31 + 2 = 33. Alt 29 und neu 32; je ein Punkt über der Marke.
- Marke des Plans (aus den Plänen unter `done/` und den Archiven gelesen: `bootstrap-m2` 2900, `replay-service` 9200, `handbuch-geruest` 6900, `prefer-return` 3500, `sqlite-interrupt-transaktion` 2700, `fallback-policy-entscheidung` 2000, `gegenproben-zeitbudget` 1400): über 1,00 vier Punkte (1,06 und 1,008 kommen hinzu), unter 0,45 fünf, unter 0,60 achtundzwanzig, also 9 und 32 Meldungen. Stimmt.
- V1 (45/130/2,5, Zuschlag 1,40): hält 56 von 61, davon 27 an *neu*, p90 ÷ p10 = 1,877 (1,88). V2 (52/129/2,45, Zuschlag 1,55): 59 von 61, p90 ÷ p10 1,844 (1,84). Anpassung an *alt* (43,6/182,6/2,16, 1,40): 26 an *neu*. Nr. 1: 31 an *neu*. Stimmt.
- Handoff: 20 Paare; *E* ÷ *H* Median 1,16, p25 1,09, p75 1,35, p90 1,66, Maximum 3,09; ohne `bootstrap-m4-konfig` und `gates-ohne-worktrees` 18 Punkte, Median 1,15, Maximum 1,66; Handoff ab 0,70 der Marke: `telemetry-handler` 0,78 und `bootstrap-m4-konfig` 0,75, einer reißt die Marke. Stimmt.
- Die Zahlen, die der Auftrag nannte (62/61 Punkte, Median 0,83, 0,47, 7 bzw. 9 Meldungen, V1 1,88), passen zur Tabelle und zum Text von Nr. 2 und 9.

## Akzeptierte Negative: Grund, Folge, Maß

| Negativ | Grund | Folge | Maß | Urteil |
|---|---|---|---|---|
| Nr. 10 a 1 (Streuung) | Hülle, keine Vorhersage | etwa einer von 30 reißt die Marke, Eintrag `ueberschreitung` | p10 bis p90 0,64 bis 1,20; 2 von 61 | trägt |
| Nr. 10 a 2 (keine Mindestmarke) | ein Punkt trüge sie, zwei zeigen die Gegenrichtung | kleiner Slice mit vielen Runden reißt die Marke | drei benannte Punkte | trägt |
| Nr. 10 a 3 (von Hand gelesen) | Pläne, zehn zurückgerechnet | "keine Kennzahl über die zweite Stelle" | 20 Paare, 10 zurückgerechnet; Folgeaussage ohne Maß, siehe L1 | trägt mit L1 |
| Nr. 10 a 4 (nicht unabhängig) | Wellen, ein Architect | Zahlen sind Größenordnungen | ohne Zahl | trägt (kein Maß möglich) |
| Nr. 10 k (Schwelle und Geltungsgrenze ohne Programm) | Programme lesen weder Closure noch Fänger-Art | Meldung kann unterbleiben; Block mit nutzloser Marke | keine Zahl der betroffenen Slices im Vertrag (Plan Teil 1: "2 von 63"; Nr. 2 *Geltungsgrenze* nennt zwei Punkte) | Grund und Folge tragen, Maß fehlt, L2 |
| Zeile 4a, Anker | Anker-Paarung nicht verdrahtet | fehlender Anker fällt erst in Closure oder Retirement-Check auf | von Hand geprüft, siehe unten | trägt |
| Raster, Geltungsgrenze gemischter Fall (I3 des Reviews) und Rundung der Auslastung (I4) | Urteil des Planner; Nr. 9 sagt nichts | siehe L3 | – | benannt, ohne eigene Zeile, L3 |

## Anker-Paarung, von Hand

Beide Enden, nach der Regel (`grundlagen-traceability.md` §Herkunfts-Anker, Sensor 1):

1. **Anker im Ziel.** `SPEC-128` Nr. 2: Absatz *Prüfung an 62 Messpunkten (Stand 2026-11-20; seit slice-groessenformel-neu; ...)* (Zeile 47); Nr. 9: *Die Schwelle 0,45 (seit slice-groessenformel-neu)* (Zeile 125). Das Ziel ist eine Datei ohne Suffix, "irgendwo in ihr" genügt: erfüllt. Die Form ist die der Ausnahme ohne Welle (`seit slice-<Kennung>`); die 3x-Schwelle ist erreicht (28 und 4 Belege, 35 und 5 Dateien unter `evidence/`).
2. **Pfad im Feld.** `harness/vertraege/SPEC-128-groessenmarke.md` existiert ab Repo-Wurzel (`ls`). Ein Suffix ` §...` entfällt, wie der Plan es will: die Nummern von `SPEC-128` sind keine Überschriften.
3. **Feld in §7: noch nicht geschrieben.** §7 trägt nur die Anlage-Zeile; das Feld entsteht mit der Closure. Die Anker-Paarung ist kein Gate (kein Träger in `harness/mk/` oder `tools/harness/`), der Verifier prüft daher beide Enden: das erste und zweite Ende liegen, das dritte schreibt die Closure (Schritt 2 unten).
4. **Link auf den Plan.** Er überlebt den Wechsel nach `done/` (Klon-Lauf oben); er hängt nicht an der Ruheort-Regel, weil `slice-mv` repo-weit nachzieht. Er hängt aber an der Datei: wandert der Plan je in ein Archiv, bricht er; `make docs-check` sähe es (`target-missing`).

**Billigung der Closure-Form im Plan (§3 *Auftrag an die Closure*): ja, mit zwei Berichtigungen.** (a) Das Pflichtfeld verlangt den Zielort in Backticks, Feld und Zielort auf einer Zeile (Vorlage `slice.template.md` §7: der Gedankenstrich, dann "liegt in", dann der Zielort in Backticks); der Plan schreibt `liegt in harness/vertraege/SPEC-128-groessenmarke.md` ohne Backticks. (b) Der Auslöser nennt je Eintrag `BEO-REPO/<slug>` und die Slices der drei Belege (`Auslöser: BEO-... (<slice>, <slice>, <slice> — 3x)`); bei 28 Belegen wählt der Planner drei Kennungen aus `evidence/`, die Zahl der Belege kann er daneben nennen.

## Findings

Alle LOW oder INFO; Fundstellen im Wortlaut des Worktrees `4e7d18eb`.

**L1 — Nr. 10 a 3 nennt kein Maß für "keine Kennzahl über die zweite Stelle" (Zusage ohne Maß, `AGENTS.md` §3.15).**
Fundstelle: `SPEC-128` Zeile 131. Nachgemessen: ein Lesefehler, der einen Punkt in den Nachbarrang verschiebt, ändert den Median höchstens um 0,005, aber p90 von *E* ÷ *S* von 1,205 auf 1,149 oder 1,221 (Ränge 54 und 56) und p10 von *E* ÷ *M* von 0,472 auf 0,460 oder 0,483. Der Satz sagt nicht, was "über die zweite Stelle" misst; bei p90 ändert sich schon die zweite Dezimalstelle um fünf bis sechs Einheiten. Folge im Betrieb: keine Entscheidung kippt (0,45 liegt unter dem sechsten Punkt 0,460), aber der Satz ist ohne Maß nicht belegt. Closure Schritt 5.

**L2 — Nr. 10 k trägt kein Maß** (die Zeilen 5 und 6 der Tabelle *Zusage · Fänger* sind damit Negative mit Grund und Folge, aber ohne Zahl; das Review hatte es als INFO geführt, der Nachzug a54c3092 behob es nicht).
Fundstelle: `SPEC-128` Zeile 142. Das Maß liegt bereits im Vertrag verstreut (Nr. 2 *Geltungsgrenze*: zwei Punkte, 0,034 und 0,12; Nr. 9: 7 von 61 Meldungen), steht aber nicht in Nr. 10 k. Folge: der Leser erfährt nicht, wie oft die Meldung unterbleiben kann. Closure Schritt 5.

**L3 — Zwei Lücken (I3, I4 des Reviews) sind als "benannte Lücken" im Plan geführt, ohne eigene Zeile mit Ausgang *offen* und Adresse** (`AGENTS.md` §3.13 und §3.16: eine Randform ohne Ort).
Fundstelle: Plan §6, Zeile "Die Review- und Verifikationsrolle prüft einen Text-Slice ..." ("bleiben als benannte Lücken stehen"). Gemeint: Satz 1 und 2 der *Geltungsgrenze* sind nicht komplementär (Zeilen mit Diff oder akzeptiertem Negativ als Fänger stehen in keiner Klasse); die Rundung der Auslastung vor dem Vergleich mit 0,45 und 1,00 steht nicht in Nr. 9 (der Vertrag selbst zählt 1,008 als "über 1,00", die Pläne schreiben 1,01). Das Raster nennt für beide als Ort einen Wortlaut, der die Entscheidung nicht trägt. Folge: die nächste Runde findet beides wieder. Closure Schritt 5.

**L4 — Nr. 2 sagt, gegen die Tabelle im Plan ließen sich "die Zahlen dieses Absatzes und von Nr. 9 und 10 a nachrechnen"; für 9 und 32 Meldungen "gegen die Marke des Plans" stimmt das nicht.**
Fundstelle: `SPEC-128` Zeile 47 und Zeile 125. Die Tabelle trägt *M* nach Nr. 1, nicht die Marke des Plans; die sieben abweichenden Marken (2900, 9200, 6900, 3500, 2700, 2000, 1400) stehen nur in den Plänen unter `done/` (zum Teil in `archiv.zip`) und nicht im Plan. Ich habe sie dort gelesen und die Zahlen bestätigt. Folge: ein späterer Prüfer kann 9 und 32 nur über die Pläne nachrechnen; die Tabellenzeile 7 des Plans nennt diese Zahlen auch nicht. Closure Schritt 5 (Spalte oder ein Satz mit den sieben Marken).

**INFO 1 — Zeile 3, Nachsatz des Hinweises.** Die Mutation `der Plan wird nicht gemessen` auf `der Plan wird gemessen` bleibt grün; der Fänger prüft `Hinweis: kein Block`, nicht den Nachsatz. Die Zusage der Zeile (Hinweis, Exit 0, keine Messung) ist durch `keine-messung` und den Exit gedeckt (`z3d`, `z3e` rot); der Nachsatz ist Text. Kein Handlungsbedarf, das Programm ändert der Slice nicht.

**INFO 2 — Zeile 4 nennt Backslash-u-Folgen in der Zusage** ("BOM, NBSP, Backslash-u-Folgen"); `make zeichen-check` fängt BOM, NBSP und Nullbreite (rot gesehen), eine Backslash-u-Folge als Text nicht (`harness/README.md` *Nicht behauptet*, `SPEC-233` Nr. 16 a). Der Fänger dafür ist das Gegenlesen auf Rohbytes, das ich gefahren habe (0 Treffer). Die Zeile könnte "Gate" und "Gegenlesen" je nennen; kein Befund am Vertrag.

**INFO 3 — der Vertrag hängt für Rechenbelege am Plan.** `SPEC-128` zeigt auf `slice-groessenformel-neu.md` als Ort der Tabelle der 62 Punkte; ein Archivieren des Plans in ein Archiv bräche den Link (Gate fängt es). Der Plan ist ohne Welle und bleibt in `done/`; nur ein Hinweis.

## Was der Planner in der Closure tut

Im Zweig (Nr. 4 der Übergangsregeln):

1. DoD abhaken, soweit erfüllt: Punkte 1, 2, 3 (`make gates` grün, von mir selbst gefahren), 4 (Review-Report liegt), 5 (Doku-Update), 6 bis 9 mit der Closure.
2. §7 füllen (Vorlage `slice.template.md` §7): *Was hat funktioniert*, *Was ging anders* (V0 statt V1 mit Programm, Zuschnitt auf Text-Slice, Architect-Berichtigung M1), zwei *Steering-Loop-Einträge* mit dem Feld "liegt in" und dem Zielort `harness/vertraege/SPEC-128-groessenmarke.md` in Backticks (Backticks, eine Zeile, kein Suffix), Auslöser `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` und `BEO-REPO/auslastung-am-handoff-als-endwert-gelesen` mit je drei Slice-Kennungen; die Nummern 2, 9 und 10 a nennt der Text.
3. §7 *Beobachtungs-Register*: entscheiden, ob aus der Lücke von Nr. 10 k (Meldung nach Nr. 9 und Geltungsgrenze ohne Sensor) eine neue Beobachtung entsteht (Beleg `evidence/slice-groessenformel-neu.md`, 1x), und die Antwort schreiben; neue Belege der beiden Einträge nur bei einem Auftreten.
4. §7 *Risiken aus §6*: jedes mit einem Ausgang; die drei offenen (Roadmap und `state.md`: *eingetreten*/*entfallen* erst nach dem Integrationsschritt, sonst *weiter offen* mit Adresse; Anker: *eingetreten*, wenn das Feld steht; [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) `Proposed`: *weiter offen*, Adresse User).
5. L1 bis L4 und I3, I4: entweder ein Text-Commit des Architect vor der Closure (Maß in Nr. 10 k und Nr. 10 a 3, Marken der sieben Pläne in Nr. 9 oder als Spalte der Tabelle, Rundung und gemischter Fall in Nr. 9 und Nr. 2) oder je eine Zeile in §6 mit Ausgang *offen* und Adresse (Architect, nächste Änderung von `SPEC-128`) und die Bedingung, dass die Closure sie nennt (`AGENTS.md` §3.13). Dazu §5 berichtigen: "Zielort ohne Feld für den Anker" (der Anker steht in `SPEC-128`, das Feld in §7).
6. Review und Verifikation in §6 und §7 mit Kennung ohne Pfad und Endung nennen: `2026-11-20-slice-groessenformel-neu-review` und `2026-11-20-slice-groessenformel-neu-verifikation`; diese Datei nach `docs/reviews/2026-11-20-slice-groessenformel-neu-verifikation.md` legen und `make docs-check` danach fahren.
7. Erst der Inhalt, dann der reine Move (`AGENTS.md` §3.3): `make slice-mv SLICE=slice-groessenformel-neu TO=done` (zwei Commits; im Klon gesehen: Plan, Roadmap, Review-Report und `SPEC-128` werden nachgezogen).

Im Integrationsschritt (Hauptbaum, nicht im Zweig):

8. Roadmap: Zeile 19 (`In Arbeit: slice-groessenformel-neu`) streichen und, liegt dann nichts in `in-progress/`, den Ruhe-Marker "Nichts in Arbeit." unter *Offene Wellen* setzen (sonst `planning-drift` von `make docs-check`, im Klon gesehen); Zeile 30 (noch "neue Faktoren in `SPEC-128` Nr. 1, `tools/groessenmarke` und Gegenproben", "Beschluss steht aus") auf V0, Text-Slice, Beschluss vom 2026-11-20 und `done/` umschreiben; im Drift-Log eine Umplanung eintragen (Zuschnitt von Faktoren-Änderung auf Text-Slice, kein Stempel, keine Marke), keine Schließung (`AGENTS.md` §3.7).
9. Register: `state.md` beider Einträge auf *verkörpert*, Zielort `SPEC-128` Nr. 2 und 9, Herkunft `slice-groessenformel-neu` (der Wortlaut "Beschluss steht aus" und "nicht verkörpert" entfällt).
10. Merge `git merge --no-ff slice-groessenformel-neu`, danach `make gates` im Hauptbaum unter der Sperre; `make gegenproben` nicht fahren, solange kein anderer Zweig `EINGABEN` ändert (die Summe der Zweige ist Sache des Integrationsschritts).
11. Dem Retirement-Check unterliegt Nr. 9 ab jetzt (Anker gesetzt): eine künftige Lockerung konsultiert die Herkunft.

## Offene Punkte für den User

1. [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) steht auf `Proposed` und trägt weder Formel noch Schwelle; annehmen oder bewusst offen lassen (Plan §6, Adresse User, keine Bedingung dieses Slice).
2. L1 bis L4 und I3, I4: nachziehen (ein weiterer Text-Commit des Architect) oder als *offen* mit Adresse in §6 stehen lassen; der Verifier hält beides für zulässig, das Maß in Nr. 10 k und die Spalte der Plan-Marken sind die zwei Stellen, die einen späteren Prüfer am meisten Arbeit sparen.
3. Die Schwelle 0,45 liegt 0,02 unter dem p10 (0,472) und 0,01 unter dem sechsten Punkt (0,460; der fünfte liegt bei 0,413): ein Punkt bei 0,46 meldet nicht, einer bei 0,44 meldet; eine Entscheidung des Users, keine Befundlage.
