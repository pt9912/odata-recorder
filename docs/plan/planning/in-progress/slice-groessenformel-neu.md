# Slice slice-groessenformel-neu: Neue Faktoren der Größenformel in SPEC-128, im Werkzeug und in den Gegenproben

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice und keine Closure-Bedingung, die von der DoD verschieden wäre (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln): er ändert `tools/groessenmarke/`, `tools/gegenproben/` und den Stempel der Gegenproben, damit gilt Nr. 7 (ein Besitzer der Gate-Kette zugleich).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Größe von Slices, der Bezug ist der nächste, den es gibt), [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Größenmarke als Gate), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenproben), Register-Einträge `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (28 Belege, Ausgang *geplant*, Auftrag an den Architect) und `BEO-REPO/auslastung-am-handoff-als-endwert-gelesen` (4×, Ausgang *geplant*, `SPEC-128` Nr. 9).

**Berührte Spec-Stellen:** `SPEC-128` Nr. 1 (Formel und Faktoren), Nr. 2 (Herkunft der Faktoren), Nr. 9 (Schwelle der Auslastung) · `SPEC-102` (Fälle der Gegenproben, die Zahlen der Formel tragen) · die Zahlen in den Plänen unter `in-progress/` (siehe §1, Auswirkung auf laufende Marken).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** pt9912 (Zweig slice-groessenformel-neu)

**Autor:** pt9912 (Planner der Closure von `welle-mock`). **Datum:** 2026-11-08.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Die Faktoren der Schätzformel in `SPEC-128` Nr. 1 (und, falls der Vorschlag es ändert, Zuschlag, Aufrundung und Schnittgrenze) stehen nach dem Beschluss, den der Architect vorlegt und der User fasst, im Vertrag, in `tools/groessenmarke/` und in den Fällen der Gegenproben; die Messpunkte der Closure von `welle-mock` (14 Punkte, Median *E* ÷ Marke 0,56, Median *E* ÷ *S* 0,75, neun von 14 unter 0,60) sind die Eingabe. Hinzu kommt seit dem Start (2026-11-20) die Frage, ob und wie die Formel für Slices ohne Go-Code gilt (§3 Auftrag 5; Eingabe aus den Closures von `slice-gate-integration-determinism`, `slice-guard-schreibwege` und `slice-regeln-verkoerpern`). Die Zahlen des Vorschlags stehen hier nicht; sie stehen im Vorschlag des Architect und, nach dem Beschluss, allein in `SPEC-128` (`AGENTS.md` §3.14).

**Eingabe:** der Vorschlag des Architect zur Neukalibrierung (Formel behalten und die Schwelle von Nr. 9 senken, oder neue Faktoren; Vorlage im Plan von `slice-mock-lesen`, Abschnitt Meldung an den Architect, Vorschlag V0 und V1, Entscheidung des Users vom 2026-11-06, Auftrag an den Architect aus der Closure von `welle-mock`). Der überarbeitete Vorschlag des Architect (Teil 1, 2026-11-20) steht in §3 *Vorschlag des Architect (Teil 1)*; er rechnet an 62 Messpunkten (nicht an 14 oder 16) und nennt seine Zahlen dort. Die Entscheidung über die Faktoren ist die des Users (§3, Entscheidungsfragen F1 bis F5).

**Stand der Eingabe am Start (Planner, 2026-11-20):** Der Vorschlag V0 und V1 (16 Messpunkte, Empfehlung V0: Formel behalten, untere Schwelle von Nr. 9 von 0,60 auf 0,45) steht im Plan von `slice-mock-lesen` (Abschnitt Meldung an den Architect, Zeilen um 429 und 474), der seit der Archivierung von `welle-mock` im Archiv `done/welle-mock/archiv.zip` liegt (der Stub unter `done/welle-mock/` trägt ihn nicht); er rechnet nur bis zu den 16 Punkten von `welle-mock`. Seither sind Messpunkte dazugekommen, zwei davon von anderer Art als alle bisherigen: `slice-gate-integration-determinism` (Messung 35 von Marke 1500, Auslastung 0,023, Skripte und Handläufe ohne Go-Code), `slice-guard-schreibwege` (1707 von 2600, Skripte ohne Go-Tests), `slice-expand-parser` (2268 von 3500, 0,65), `slice-gegenproben-zeitbudget` (271 von 1400, 0,19) und `slice-bootstrap-m4-konfig` (Ende ÷ Handoff 2,58); `slice-regeln-verkoerpern` trägt keinen Block (kein Go-Code, Diff 156 Zeilen, Begründung in §4 dort: ein Slice ohne Go-Code sei kein Messpunkt der Faktoren). Die Register-Einträge führen am Start 30 Dateien unter `evidence/` (`schaetzformel-ueberschaetzt-den-slice`, die Zahl 28 oben ist der Stand der Closure von `welle-mock`) und 5 (`auslastung-am-handoff-als-endwert-gelesen`). Ein Beschluss des Users über V0, V1 oder eine dritte Form liegt nicht vor.

**Messpunkte nach den 14 (Eingabe, Geber `slice-gegenproben-zeitbudget`):** `Messpunkt slice-gegenproben-zeitbudget.md: zeilen=9 nebenlaeufig=0 fakes=2 produktcode=15 schaetzung=1135 marke=1400 messung=271` (Stand `Ende`, Auslastung rund 0,19 der Marke, unter der Schwelle 0,60 von `SPEC-128` Nr. 9; Plan §4 dort; der Slice ist ein Werkzeug-Slice mit einem Test von 175 Zeilen und 15 Zeilen Produktcode).

**Auswirkung auf laufende Marken.** Eine Marke, die der Planner vor dem ersten Code-Commit nach der alten Formel setzte, bleibt in ihrem Plan stehen; das Gate `make groessenmarke` rechnet die Marke aus dem Block des Plans (`SPEC-128` Nr. 3 p, Nr. 5), nicht aus einer Angabe. Gilt nach dem Beschluss eine andere Formel, so zeigt `make groessenmarke` für einen Plan in `in-progress/` mit der alten Marke „Marke ungleich Formel“ und färbt rot. Der Beschluss nennt deshalb vor dem Code, wie laufende Pläne behandelt werden: (a) der Planner ersetzt die Marke mit Grund in §4 jedes Plans in `in-progress/` (AGENTS.md §3.14: „eine ältere ersetzt er mit Grund in §4“), (b) die Formel trägt eine Übergangsregel nach Stichtag oder Plan-Datum, (c) die Änderung wartet, bis kein Plan mit Block in `in-progress/` liegt. **Antwort des Architect (Teil 1):** §3 *Vorschlag des Architect (Teil 1)*, Abschnitt 4. **Stand am Start 2026-11-20:** außer diesem Plan liegt kein Plan in `in-progress/`, kein Plan dort trägt einen Block, und kein Slice-Worktree läuft (`git worktree list` nennt nur den Hauptbaum): „Marke ungleich Formel“ trifft heute keinen Plan, die Wege (a) und (b) sind gegenstandslos, solange kein anderer Slice vor dem Merge dieses Slice startet; startet einer, geht es nach Auftrag 2. Die Pläne in `open/` führen die Marke als Prosa (kein Block, das Gate misst sie nicht); der Planner rechnet sie beim jeweiligen Start neu. Stand 2026-11-08 (überholt; die drei Pläne liegen in `done/`): Pläne mit Marke in `in-progress/` waren `slice-fallback-policy` (S 1900 / M 2600), `slice-etag-versionen` (S 1910 / M 2600, vorläufig) und `slice-admin-lebenszyklus` (S 1615 / M 2200), alle als Prosa; kein Block, also misst das Gate sie nicht (`SPEC-128` Nr. 3 p). Vorläufige Marken der Pläne in `open/` (z. B. `slice-gate-integration-determinism`) rechnet der Planner nach dem Beschluss neu.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Schwelle der Auslastung (`SPEC-128` Nr. 9) zu senken, ohne dass der Beschluss es nennt — das Meldeverfahren (unter 0,60 oder über 1,00 an den Architect) ist Teil des Vorschlags, nicht dieser Plan.
- Eine Änderung des Gates (`make groessenmarke` rechnet, misst und schneidet nicht; Nr. 10) oder seiner Verdrahtung — andere Mechanik; das Gate bleibt, die Zahlen wechseln.
- Das Setzen von Marken für Pläne, die erst starten — das tut der Planner im jeweiligen Start-Commit; dieser Slice ändert die Formel, nicht die Pläne (außer der Behandlung laufender Marken nach dem Beschluss).
- Eine Regel zum Schnitt von Slices (Schnittgrenze *S* ≥ 2223 und Faktoren zum Schnitt in den Plänen) — gehört zum Beschluss, falls der Vorschlag die Schnittgrenze ändert; sonst unverändert.
- Die Gegenprobe zu `make gegenproben` selbst, ihr Zeitbudget und ihr Stempel außer den Fällen, die Zahlen der Formel tragen — die Laufzeit der Gegenproben ist Gegenstand der ADR zum Zeitbudget und ihres Slice, den der User anlegt.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [ ] `SPEC-128` Nr. 1 und 2 tragen die neuen Faktoren (und Nr. 9 die Schwelle, falls der Beschluss sie ändert) mit der Herkunft aus den Messpunkten (14 der Closure von `welle-mock` und die seither hinzugekommenen, §3 Auftrag 6) und der Antwort zu Slices ohne Go-Code (§3 Auftrag 5); `tools/groessenmarke/` rechnet nach der neuen Formel, die Tests des Programms führen die Randwerte der neuen Schnittgrenze (beide Seiten und den Fall *S* gleich der Grenze), jede Zusage mit einer vom Implementer rot gesehenen Mutation (`AGENTS.md` §3.10).
- [ ] Die Fälle der Gegenproben, die Zahlen der Formel tragen (`SPEC-102` Fälle 38 bis 120, soweit sie Marken oder Schätzungen führen), sind nachgezogen und fangen die Mutation „alte Faktoren“; der Stempel der Gegenproben steht nach dem letzten Eingabe-Commit.
- [ ] `make gates` grün; `make gegenproben` einmal gefahren und grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `SPEC-128`, `SPEC-102` (Architect), Zahlen und Verweise in den Plänen, die sie nennen (`AGENTS.md` §3.9), `harness/README.md` falls der Wortlaut des Gates betroffen ist; ein Beleg der Auswirkung auf laufende Marken (Liste der Pläne, die der Planner angepasst hat, oder Nachweis, dass keiner betroffen ist).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; die Register-Einträge `schaetzformel-ueberschaetzt-den-slice` und `auslastung-am-handoff-als-endwert-gelesen` bekommen den Ausgang *verkörpert* mit Zielort `SPEC-128`.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

Der Planner schneidet; der Architect entscheidet (Aufträge unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/vertraege/SPEC-128-groessenmarke.md` | update (Architect) | Faktoren, Herkunft, ggf. Schwelle und Schnittgrenze |
| `tools/groessenmarke/formel.go`, `formel_test.go` und die Tests, die Zahlen der Formel führen (`main_test.go`, `plan_test.go`, `mengen_*_test.go`, `fixtures_test.go`) | update | Konstanten der Formel, Randwerte, Mutationen |
| `harness/vertraege/SPEC-102-*.md`, `tools/gegenproben/` | update | Fälle der Gegenproben mit Zahlen der Formel |
| `tools/harness/gegenproben-eingabe.sh`, `harness/gegenproben.stempel` | update | Stempel nach dem letzten Eingabe-Commit |
| `docs/plan/planning/` (Pläne mit Marke) | update | laufende Marken nach dem Beschluss (Behandlung vor dem Code) |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Teil 1: der Vorschlag aus Auftrag 1, nur der Vorschlag, kein Vertragstext) → **User** (Beschluss der Faktoren, siehe §4 *Start belegt* und §6) → Architect (Teil 2: Vertragstext, Tabelle *Zusage · Fänger*, Raster, Zahlen, Behandlung laufender Marken, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit; die Marke dieses Slice rechnet nach der Formel vor der Änderung) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); schwere Läufe (`make gates`, `make gegenproben`) unter `flock /tmp/odata-recorder-heavy.lock`; jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`; Commit-Messages per `git commit -F <datei>`; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Aufträge an den Architect:**

1. **Vorschlag und Beschluss.** Den überarbeiteten Vorschlag mit den 14 Messpunkten der Closure von `welle-mock` und den 28 Belegen des Registers als Herkunft vorlegen (Faktoren, Zuschlag, Aufrundung, Schnittgrenze, Schwelle von Nr. 9); der Beschluss ist der des Users, nicht des Architect. Der Verweis auf den Ort des Vorschlags geht an den Planner (siehe §1, Eingabe).
2. **Auswirkung auf laufende Marken** (§1): welche der Wege (a) bis (c) gilt, mit Grund; Stichtag oder Kriterium; Wirkung auf `make groessenmarke` für einen Plan mit Block in `in-progress/` (Nr. 5 a).
3. **Randformen** vor dem ersten Code-Commit im Vertrag (Raster in §6): Messpunkte genau auf der neuen Grenze, *S* gleich null, ein Plan ohne Zeilen einer Klasse, Überlauf der Ganzzahl-Rechnung, ein Plan mit der alten Marke.
4. **Tabelle *Zusage · Fänger*** (§3), **Raster** (§6), **erwartete Produktcode-Größe** für die Marke.
5. **Slices ohne Go-Code** (Start 2026-11-20, Geber `slice-gate-integration-determinism`, `slice-guard-schreibwege`, `slice-regeln-verkoerpern`): ob die Formel für Slices aus Skripten, Verträgen und Dokumentation gilt, gilt nicht oder eine zweite Klasse braucht; `slice-gate-integration-determinism` maß 0,023 der Marke, `slice-regeln-verkoerpern` führt keinen Block mit der Begründung, ein Slice ohne Go-Code sei kein Messpunkt der Faktoren, und `slice-guard-schreibwege` hielt die Formel ohne Beleg für Skripte (Messung 1707 von 2600). Der Vorschlag nennt die Antwort mit Herkunft; die Grenze der Formel steht danach in `SPEC-128` Nr. 2.
6. **Neue Messpunkte seit `welle-mock`** in den Vorschlag: alle `Messpunkt …`-Zeilen der Pläne unter `done/` (auch archiviert) und die Dateien unter `evidence/` der beiden Register-Einträge; der Vorschlag nennt die Zahl der Punkte, die Klasse je Punkt (mit und ohne Go-Code) und trennt, was Handoff und was Ende ist (`BEO-REPO/auslastung-am-handoff-als-endwert-gelesen`).

**Vorschlag des Architect (Teil 1, 2026-11-20).** Nur ein Vorschlag: kein Vertragstext in `SPEC-128`, keine Marke, kein Code; die Faktoren beschließt der User (§4 *Start belegt*). Teil 1 der Aufträge 1, 2, 5 und 6 steht hier; die Aufträge 3 und 4 (Randformen, Tabelle *Zusage · Fänger*, Produktcode-Größe) sind Teil 2 und folgen dem Beschluss. Die Zahlen sind aus den Plänen von Hand gelesen und nachgerechnet, nicht vom Gate gemessen; der Rechenstand liegt im Scratchpad des Architect, nicht im Repo.

**Empfehlung: V0, die Formel behalten (85 / 80 / 200 / 1, Zuschlag 1,35, Schnittgrenze unverändert), die untere Schwelle von Nr. 9 von 0,60 auf 0,45 setzen, die Geltungsgrenze der Formel in Nr. 2 nennen und die Spanne Handoff → Ende in Nr. 9 eintragen.** Gründe: (1) Keine Variante verengt die Streuung (Verhältnis p90 ÷ p10 von Ende ÷ Schätzung 1,84 bis 1,89 bei allen dreien); ein anderer Faktor verschiebt nur den Schwerpunkt, und den verschiebt die Schwelle billiger. (2) Die Marke von V0 hält 59 von 61 Enden, an den 32 Punkten der Gruppe *neu*, die in keiner Anpassung standen, 31; V1 hält dort 27, und eine auf *alt* neu angepasste Form hielte 26. (3) V0 ändert keinen Code: Nr. 9 steht nur im Vertrag und in der Rolle des Planner (`0,60` steht in keinem Programm, keinem Fall der Gegenproben und keinem Command), V1 und V2 ändern Programm, 83 Fälle der Gegenproben, Stempel und laufende Marken. (4) Mit 0,45 meldet Nr. 9 sieben von 61 Punkten (11 %) statt 33 (54 %).

**Zu `AGENTS.md` §3.6:** Nr. 9 ist keine Schwelle eines Gates; kein Ziel und kein Programm prüft sie, die Zahl steht allein in `SPEC-128` und [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) trägt keine Zahl der Formel. Eine Senkung der Meldeschwelle lockert damit kein Gate und braucht keine ADR; sie ist der dokumentierte Beschluss des Users mit der Herkunft in Nr. 2 und 9. Der Architect hat das gegen den Wortlaut von §3.6 geprüft (Verdikt: Lockerung legitim, durch diesen Plan und den Vertragstext von Teil 2 dokumentiert).

**1. Messpunkte (Auftrag 6).** 62 Slices tragen eine Messung am Ende und eine Schätzung nach Nr. 1; 35 davon haben eine `Messpunkt …`-Zeile in ihrem Plan (`done/`, auch in den Archiven `archiv.zip`), 27 stehen in den Closure-Tabellen von `welle-record-replay` und `welle-mock`, in den Abschnitten *Größen* der Pläne oder in den 35 Dateien unter `evidence/` der beiden Register-Einträge (alle 35 Dateien sind abgedeckt). Gruppe *alt* (29): die Slices der beiden Closure-Tabellen (17 und 14, `odata-url-pfad` in beiden, `test-contract` ohne Schätzung nicht gezählt); Gruppe *neu* (33): alle übrigen Pläne unter `done/` mit Block. Klasse: **60 mit Go-Code, 1 Skript** (`guard-schreibwege`), **1 Handlauf** (`gate-integration-determinism`, 35 Zeilen gegen 1040 geschätzt); `regeln-verkoerpern` führt keinen Block (Diff 156 Zeilen) und `test-contract` keine Schätzung, beide stehen nicht in der Tabelle. Die Rechnung läuft ohne den Handlauf: **n = 61**. Spalte *H* ist der Stand beim Handoff (Messung des Implementers vor Review und Verifikation), *E* das Ende nach Review, Verifikation und Nachzug; nur **20 Punkte tragen beide Werte**, bei den übrigen kennt der Plan nur das Ende. Die Spalten *z* und *a* sind bei zehn Punkten aus *S* zurückgerechnet (`m2-referenzdienst`, `workspace-use-case`, `odata-url-filter`, `query-interpreter`, `audit-dienste`, `bootstrap-m4`, `handbuch-meldungskatalog`, `handbuch-katalog`, `unsichtbare-zeichen`, `guide-groessen-gate`); *S* ist der Wert des Blocks beim Start, auch wenn die Tabelle danach wuchs (`handbuch-meldungskatalog` 24 → 32 Zeilen, `query-interpreter` 34 → 40), *M* ist nach Nr. 1 neu gerechnet (sieben Pläne trugen eine Marke, die nicht der Formel aus ihrer *S* folgt, weil *S* beim Start kleiner war oder eine ältere Fassung galt: `bootstrap-m2`, `replay-service`, `handbuch-geruest`, `prefer-return`, `sqlite-interrupt-transaktion`, `fallback-policy-entscheidung`, `gegenproben-zeitbudget`; dort weicht Spalte *M* vom Plan ab).

Kennzahlen (n = 61): Ende ÷ *S* Median 0,83 (*alt* 0,77, *neu* 0,90), p10 0,64, p90 1,20; **49 von 61 Enden liegen unter der Schätzung**, 5 unter 0,60 ihrer Schätzung. Nach Größe der Schätzung: Median 0,75 (*S* unter 1500, 12 Punkte), 0,89 (1500 bis 2999, 27), 0,77 (3000 bis 5999, 14), 0,71 (ab 6000, 8): kein Trend, der einen Faktor je Größe trüge. Ende ÷ Marke (V0): Median 0,60, p10 0,47, p90 0,85; unter 0,60: 31, unter 0,50: 8, unter 0,45: 5, unter 0,40: 3; über 1,00: 2 (`telemetry-port` 1,03, `bootstrap-m4-konfig` 1,92).

**Handoff und Ende getrennt** (`BEO-REPO/auslastung-am-handoff-als-endwert-gelesen`): Ende ÷ Handoff bei 20 Punkten: Median 1,16, p25 1,09, p75 1,35, p90 1,66, Maximum 3,09; ohne die zwei Punkte, bei denen die Tabelle nach dem Handoff wuchs (`bootstrap-m4-konfig` 2,58 nach vier Runden mit Fängern der Klassen, `gates-ohne-worktrees` 3,09 mit Tabelle 12 → 16 Zeilen) 18 Punkte mit Median 1,15, p90 1,49, Maximum 1,66. `SPEC-128` Nr. 9 misst schon das Ende („Auslastung = Messung am Ende ÷ Marke“); der Befund der Register-Einträge ist die Lesart der Rollen, nicht der Wortlaut: ein Handoff-Wert ist kein Endwert. Ein Handoff ab 0,70 der Marke (nach V0) kommt in den 20 Paaren zwei Mal vor (`bootstrap-m4-konfig` 0,75 → Ende 1,92, `telemetry-handler` 0,78 → Ende 0,85); in diesen 20 Paaren reißt nach V0 nur `bootstrap-m4-konfig` die Marke (`bootstrap-m2` trug eine ältere Marke von 2900 und riss sie mit 1,06; nach V0 sind es 3400 und 0,90).

| Slice | Gr. | Art | z | f | p | *S* | *M* | *E* | *E* ÷ *S* | *E* ÷ *M* | *H* | *E* ÷ *H* |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `v2-import-codecs` | alt | go | 9 | 1 | 120 | 1045 | 1500 | 758 | 0,73 | 0,51 | 695 | 1,09 |
| `http-lebenszyklus` | alt | go | 20 | 1 | 140 | 1990 | 2700 | 1728 | 0,87 | 0,64 | 1161 | 1,49 |
| `test-integration` | alt | go | 30 | 1 | 330 | 2960 | 4000 | 1931 | 0,65 | 0,48 | 1835 | 1,05 |
| `protokolltreue` | alt | go | 22 | 1 | 0 | 1965 | 2700 | 1517 | 0,77 | 0,56 | 1404 | 1,08 |
| `test-determinism` | alt | go | 15 | 0 | 80 | 1280 | 1800 | 694 | 0,54 | 0,39 | 613 | 1,13 |
| `m2-referenzdienst` | alt | go | 11 | 1 | 100 | 1190 | 1700 | 980 | 0,82 | 0,58 | – | – |
| `v2-query-filter` | alt | go | 37 | 0 | 400 | 3370 | 4600 | 2546 | 0,76 | 0,55 | – | – |
| `bootstrap-m2` | alt | go | 16 | 3 | 550 | 2445 | 3400 | 3075 | 1,26 | 0,90 | 2280 | 1,35 |
| `telemetry-port` | alt | go | 17 | 1 | 580 | 2150 | 3000 | 3093 | 1,44 | 1,03 | – | – |
| `groessenmarke-werkzeug` | alt | go | 44 | 2 | 750 | 4670 | 6400 | 5342 | 1,14 | 0,83 | 3589 | 1,49 |
| `replay-service` | alt | go | 66 | 4 | 800 | 6925 | 9400 | 6017 | 0,87 | 0,64 | – | – |
| `workspace-use-case` | alt | go | 70 | 6 | 950 | 7815 | 10600 | 6540 | 0,84 | 0,62 | – | – |
| `handbuch-geruest` | alt | go | 105 | 2 | 1000 | 9815 | 13300 | 6956 | 0,71 | 0,52 | – | – |
| `v2-query` | alt | go | 49 | 1 | 1000 | 5130 | 7000 | 3781 | 0,74 | 0,54 | – | – |
| `v2-payload-lesen` | alt | go | 57 | 0 | 1400 | 5970 | 8100 | 4085 | 0,68 | 0,50 | – | – |
| `v2-metadaten` | alt | go | 57 | 0 | 1500 | 6070 | 8200 | 4962 | 0,82 | 0,61 | – | – |
| `odata-url-pfad` | alt | go | 61 | 1 | 1700 | 6790 | 9200 | 6517 | 0,96 | 0,71 | – | – |
| `odata-url-filter` | alt | go | 35 | 0 | 650 | 3460 | 4700 | 2388 | 0,69 | 0,51 | – | – |
| `edm-ausdruecke` | alt | go | 38 | 0 | 600 | 3645 | 5000 | 3229 | 0,89 | 0,65 | 2780 | 1,16 |
| `query-interpreter` | alt | go | 34 | 0 | 550 | 3275 | 4500 | 3201 | 0,98 | 0,71 | 2588 | 1,24 |
| `odata-json-schreiben` | alt | go | 42 | 1 | 1000 | 4570 | 6200 | 3367 | 0,74 | 0,54 | 3000 | 1,12 |
| `odata-json-eingabe` | alt | go | 45 | 0 | 750 | 4360 | 5900 | 3692 | 0,85 | 0,63 | – | – |
| `mock-zuordnung` | alt | go | 24 | 8 | 740 | 4265 | 5800 | 2737 | 0,64 | 0,47 | – | – |
| `mock-lesen` | alt | go | 46 | 3 | 1080 | 5370 | 7300 | 4371 | 0,81 | 0,60 | – | – |
| `mock-schreiben` | alt | go | 81 | 4 | 900 | 8220 | 11100 | 4544 | 0,55 | 0,41 | – | – |
| `snapshot-use-case` | alt | go | 81 | 4 | 500 | 7845 | 10600 | 5581 | 0,71 | 0,53 | – | – |
| `mock-abnahme` | alt | go | 83 | 2 | 0 | 7120 | 9700 | 2335 | 0,33 | 0,24 | 2325 | 1,00 |
| `bootstrap-m3` | alt | go | 16 | 1 | 262 | 1747 | 2400 | 1367 | 0,78 | 0,57 | – | – |
| `bootstrap-m3-e2e-mock` | alt | go | 9 | 0 | 40 | 765 | 1100 | 570 | 0,75 | 0,52 | – | – |
| `admin-authentisierung` | neu | go | 22 | 0 | 100 | 1860 | 2600 | 1537 | 0,83 | 0,59 | – | – |
| `admin-lebenszyklus` | neu | go | 19 | 1 | 190 | 1955 | 2700 | 2396 | 1,23 | 0,89 | – | – |
| `admin-listener-token` | neu | go | 24 | 0 | 150 | 2080 | 2900 | 1512 | 0,73 | 0,52 | – | – |
| `admin-rahmen` | neu | go | 13 | 3 | 110 | 1750 | 2400 | 1785 | 1,02 | 0,74 | – | – |
| `audit-dienste` | neu | go | 19 | 2 | 260 | 2195 | 3000 | 2184 | 0,99 | 0,73 | – | – |
| `audit-log` | neu | go | 21 | 1 | 290 | 2185 | 3000 | 1895 | 0,87 | 0,63 | 1463 | 1,30 |
| `audit-verdrahtung` | neu | go | 12 | 2 | 0 | 1360 | 1900 | 1288 | 0,95 | 0,68 | – | – |
| `etag-versionen` | neu | go | 24 | 0 | 260 | 2190 | 3000 | 1999 | 0,91 | 0,67 | 1954 | 1,02 |
| `expand-parser` | neu | go | 30 | 0 | 150 | 2555 | 3500 | 2268 | 0,89 | 0,65 | 1723 | 1,32 |
| `handbuch-meldungskatalog` | neu | go | 24 | 0 | 650 | 2570 | 3500 | 2739 | 1,07 | 0,78 | – | – |
| `if-match` | neu | go | 28 | 0 | 200 | 2460 | 3400 | 2454 | 1,00 | 0,72 | 1478 | 1,66 |
| `inspection-coverage-diagnose` | neu | go | 22 | 1 | 190 | 2155 | 3000 | 1725 | 0,80 | 0,57 | – | – |
| `inspection-use-case` | neu | go | 18 | 1 | 200 | 1845 | 2500 | 1408 | 0,76 | 0,56 | 1222 | 1,15 |
| `prefer-return` | neu | go | 32 | 1 | 300 | 3065 | 4200 | 2399 | 0,78 | 0,57 | – | – |
| `telemetry-handler` | neu | go | 16 | 0 | 120 | 1415 | 2000 | 1705 | 1,20 | 0,85 | 1555 | 1,10 |
| `telemetry-log` | neu | go | 20 | 1 | 260 | 2085 | 2900 | 1600 | 0,77 | 0,55 | 1372 | 1,17 |
| `telemetry-spans` | neu | go | 24 | 0 | 280 | 2230 | 3100 | 2203 | 0,99 | 0,71 | – | – |
| `v2-payload-schreiben` | neu | go | 43 | 1 | 650 | 4300 | 5900 | 3293 | 0,77 | 0,56 | – | – |
| `v4-schreiber-marker-text` | neu | go | 13 | 0 | 30 | 1070 | 1500 | 771 | 0,72 | 0,51 | – | – |
| `sqlite-interrupt-transaktion` | neu | go | 23 | 1 | 90 | 2160 | 3000 | 1796 | 0,83 | 0,60 | – | – |
| `fallback-policy` | neu | go | 9 | 0 | 300 | 1020 | 1400 | 1311 | 1,29 | 0,94 | – | – |
| `fallback-policy-entscheidung` | neu | go | 16 | 0 | 200 | 1485 | 2100 | 1706 | 1,15 | 0,81 | – | – |
| `hybrid-service` | neu | go | 26 | 2 | 330 | 2820 | 3900 | 2241 | 0,79 | 0,57 | – | – |
| `hybrid-service-upstream` | neu | go | 19 | 2 | 300 | 2220 | 3000 | 2126 | 0,96 | 0,71 | – | – |
| `bootstrap-m4-konfig` | neu | go | 6 | 0 | 150 | 630 | 900 | 1729 | 2,74 | 1,92 | 671 | 2,58 |
| `bootstrap-m4` | neu | go | 16 | 2 | 130 | 1820 | 2500 | 2222 | 1,22 | 0,89 | – | – |
| `gegenproben-zeitbudget` | neu | go | 9 | 2 | 15 | 1135 | 1600 | 271 | 0,24 | 0,17 | – | – |
| `gates-ohne-worktrees` | neu | go | 16 | 0 | 375 | 1655 | 2300 | 1057 | 0,64 | 0,46 | 342 | 3,09 |
| `handbuch-katalog` | neu | go | 22 | 0 | 400 | 2160 | 3000 | 2163 | 1,00 | 0,72 | – | – |
| `unsichtbare-zeichen` | neu | go | 36 | 1 | 480 | 3560 | 4900 | 3189 | 0,90 | 0,65 | – | – |
| `guard-schreibwege` | neu | skript | 20 | 0 | 300 | 1900 | 2600 | 1707 | 0,90 | 0,66 | – | – |
| `guide-groessen-gate` | neu | go | 15 | 0 | 70 | 1270 | 1800 | 744 | 0,59 | 0,41 | – | – |
| `gate-integration-determinism` | neu | handlauf | 12 | 0 | 70 | 1040 | 1500 | 35 | 0,03 | 0,02 | – | – |

Spalten: *z* Zeilen der Tabelle, *f* Fakes, *p* Produktcode nach der Erwartung des Blocks, *S* Schätzung nach Nr. 1, *M* Marke nach Nr. 1, *E* Messung am Ende, *H* Messung am Handoff, Quotienten gerundet; die Zeile von `gate-integration-determinism` ist in keiner Kennzahl enthalten.

**2. Slices ohne Go-Code (Auftrag 5): keine zweite Klasse; die Formel gilt für Zeilen mit Test-Code als Fänger, sonst kein Block.** Die Sprache trennt die Punkte nicht, die Art des Fängers trennt sie: `guard-schreibwege` (Skript, Prüflauf mit 1064 Zeilen, Guard 643 Zeilen von Hand) liegt bei 0,90 der Schätzung und 0,66 der Marke, mitten in den Go-Punkten. `gate-integration-determinism` (12 Zeilen, deren Fänger Handläufe und Ausgaben von Gates sind) misst 35 Zeilen bei 1040 geschätzt (0,034), und `regeln-verkoerpern` (16 Zeilen über Guide-Dateien, Diff 156) hätte nach Nr. 1 *S* = 1280 gegeben (0,12): eine Zeile ohne Test-Code erzeugt keine 80 Zeilen. Zwei Punkte tragen keinen Faktor (zwei Parameter, die um 3,6 auseinanderliegen), und ein Block für sie liefert kein Signal; eine dritte Klasse in der Spalte *Klasse* hieße Änderung von Nr. 3 e, des Programms und der Fälle 38 bis 120 für zwei Punkte. Zwei Go-Slices mit niedrigem Wert gehören nicht in diese Klasse: `gegenproben-zeitbudget` (0,24) zählt 2 Fakes mit 400 von 1135 Schätzung bei Fakes von wenigen Zeilen, `mock-abnahme` (0,33) liefert allein Tests ohne Produktcode. **Vorschlag für Nr. 2:** die Formel gilt für einen Slice, dessen Zeilen der Tabelle überwiegend einen Test mit Code als Fänger tragen (Go-Test, Prüflauf-Skript, Fall der Gegenprobe); ein Slice, dessen Fänger überwiegend Handläufe, Ausgaben eines Gates oder Prüfungen des Doku-Gates sind, führt keinen Block (Nr. 3 p: kein Block, Hinweis, Exit 0; Nr. 10 d) und liefert keinen Messpunkt; der Planner nennt den Grund in §4 des Plans, wie `slice-regeln-verkoerpern` es tat. Kosten: nur Text; der Vertrag lässt einen Plan ohne Block schon zu. Akzeptiertes Negativ: ein solcher Slice wächst ohne Marke (Nr. 10 d: Code vor der Marke bleibt ungesehen); Grund: seine Zeilen sind Text und Handläufe, deren Wachstum `make docs-check` je Datei und die Gates begrenzen, und die Zahl der betroffenen Slices ist klein (2 von 63).

**3. Varianten (Auftrag 1), gerechnet an n = 61.** V0 ist die Formel von heute. V1 ist der Vorschlag vom 2026-11-07 (*S* = 45 · *z* + 130 · *f* + 2,5 · *p*, Zuschlag 1,40, Schwelle 0,50), hier an den 61 Punkten geprüft. V2 ist die dritte Form: gewichtete kleinste Quadrate (Gewicht 1 ÷ *S*², also relativer Fehler) auf allen 61 Punkten, *S* = 52 · *z* + 129 · *f* + 2,45 · *p*, mit dem Zuschlag 1,55, bei dem wie bei V0 zwei Punkte über der Marke liegen; Schwelle 0,45.

| | V0 (Formel behalten) | V1 (Vorschlag 2026-11-07) | V2 (neu angepasst) |
|---|---|---|---|
| Änderung | Text in Nr. 2 und 9 | Nr. 1, Programm, Fälle 38 bis 120, Stempel, laufende Marken | wie V1 |
| Ende ÷ Schätzung, Median (p10 bis p90) | 0,83 (0,64 bis 1,20) | 1,04 (0,74 bis 1,40) | 0,97 (0,70 bis 1,29) |
| p90 ÷ p10 (Streuung) | 1,88 | 1,89 | 1,84 |
| Enden unter der Schätzung | 49 von 61 | 24 | 33 |
| Enden über der Marke | 2 (`telemetry-port`, `bootstrap-m4-konfig`) | 5 (`admin-lebenszyklus`, `audit-verdrahtung`, `telemetry-handler`, `bootstrap-m4-konfig`, `bootstrap-m4`) | 2 (`bootstrap-m4-konfig`, `bootstrap-m4` mit 1,01) |
| Über der Marke an *alt* (29) / *neu* (32) | 1 / 1 | 0 / 5 | in der Probe, keine Aussage |
| Ende ÷ Marke, Median (p10 bis p90) | 0,60 (0,47 bis 0,85) | 0,73 (0,53 bis 0,98) | 0,62 (0,43 bis 0,81) |
| Unter 0,60 | 31 | 13 | 27 |
| Schwelle und Meldungen (unten + über 1,00) | 0,45: 5 + 2 = 7 | 0,50: 5 + 5 = 10 | 0,45: 7 + 2 = 9 |
| Meldungen mit der Schwelle von heute (0,60) | 31 + 2 = 33 | 13 + 5 = 18 | 27 + 2 = 29 |

Lesart: **(a)** Die Streuung ist in allen drei Formen gleich; V1 und V2 setzen den Schwerpunkt von Ende ÷ Schätzung von 0,83 auf rund 1,0, und die Auslastung rückt im Median von 0,60 auf 0,73 und 0,62: der Gewinn von V2 ist eine Zahl, kein besseres Signal. **(b)** V1 war in der Probe der 16 Punkte 16 von 16 haltbar; an den Punkten, die nach ihm kamen, reißt es die Marke fünf Mal in 32 (16 %), V0 ein Mal (3 %). Eine Anpassung nur an *alt* (43,6 · *z* + 182,6 · *f* + 2,16 · *p*, Zuschlag 1,40) hielte an *neu* 26 von 32; die Koeffizienten wandern gegenüber der Anpassung an alle Punkte um 20 bis 40 % (*z* 43,6 → 52, *f* 183 → 129, *p* 2,16 → 2,45): jede Anpassung an diesen Daten ist beim nächsten Dutzend Punkte überholt. **(c)** Die Marke hält den Schnitt, nicht die Prognose: V0 hält 59 von 61 (97 %); die zwei Ausreißer haben bekannte Ursachen (`bootstrap-m4-konfig`: *S* 630, vier Review-Runden mit Fängern der Klassen, Ende ÷ Handoff 2,58; `telemetry-port`: Produktcode 580 erwartet) und sind der Fall, für den der Eintrag `ueberschreitung` mit Grund da ist (Nr. 5 d). Geprüft und verworfen: eine Form mit Achsenabschnitt (*E* ≈ 0,67 · *S* + 458, Spanne 0,66 bis 1,30, p90 ÷ p10 1,97, gleiches Bild, dazu ein Strukturbruch in der Formel) und V0 mit Zuschlag 1,40 (ein Punkt weniger über der Marke, sechs statt fünf unter 0,45, dafür dieselben Änderungen an Programm und Gegenproben wie V1). Schnittgrenze (*S* ≥ 2223): bei V0 unverändert; bei V1 und V2 ist sie neu zu setzen (Teil 2).

**Schwelle bei V0.** 0,45 liegt knapp unter dem p10 von Ende ÷ Marke (0,47): rund jeder zehnte Slice meldet nach unten, zwei von 61 nach oben; die Meldung trennt wieder den Ausreißer vom Regelfall (heute: 33 von 61, `slice-mock-lesen` mit 0,599 ist ein Beleg dafür, dass 0,60 der Median ist). 0,40 meldete drei von 61 nach unten (`gegenproben-zeitbudget`, `mock-abnahme`, `test-determinism` mit 0,39); `mock-schreiben` und `guide-groessen-gate` mit 0,41 lägen knapp darüber.

**4. Laufende Marken (Auftrag 2).** Bei **V0** ist die Formel dieselbe: keine Marke ändert sich, das Gate rechnet gleich, die Prosa-Marken der Pläne in `open/` bleiben gültig; die Wege (a) bis (c) sind gegenstandslos. Bei **V1 oder V2** gilt (a): der Planner ersetzt die Marke jedes Plans mit Block in `in-progress/` mit Grund in §4 und dem Eintrag `ersetzt` (Nr. 3 n); (b) scheidet aus, weil das Programm keinen Plan-Zeitpunkt liest (Nr. 10 e: es sieht den Stand, nicht den Verlauf) und eine Übergangsregel ein neues Merkmal der Eingabe wäre, (c) scheidet aus, weil ein wartender Slice-Satz bei drei parallelen Worktrees den Durchsatz hält, den die Änderung kostet. Wirkung auf `make groessenmarke` bis zur Ersetzung: Befund `Marke <m> weicht von der Formel ab` (Nr. 5 b), Gate rot, Exit 1 mit Messpunkt-Zeile. Stand 2026-11-20: kein Plan in `in-progress/` trägt einen Block; die Wirkung ist heute null, und für Pläne, die vor dem Merge dieses Slice starten, rechnet der Planner im Start-Commit nach der dann geltenden Formel.

**5. Zuschnitt dieses Slice nach dem Beschluss.** Bei V0 entfallen die Arbeiten an `tools/groessenmarke/`, an den Fällen 38 bis 120 von `SPEC-102`, am Stempel und der Lauf von `make gegenproben`: der Slice ändert `SPEC-128` Nr. 2 (Herkunft aus 61 Punkten, Geltungsgrenze, Datenlage), Nr. 9 (Schwelle, Spanne Handoff → Ende) und Nr. 10 a (Grund: die Punktezahl), zieht Zahlen und Verweise in den Plänen nach (`AGENTS.md` §3.9) und hält `make gates`. Eine Zusage des Vertrags mit Programm-Fänger entsteht dabei nicht (Nr. 9 ist Regel der Rolle, nicht des Programms): Fänger-Tabelle und Mutationen von Teil 2 entfallen, das Raster der Randformen (Auftrag 3) ist `gilt nicht` mit diesem Grund. Der Planner passt DoD und §3 an, sobald der Beschluss vorliegt (§6). Bei V1 oder V2 bleibt der Plan wie geschrieben; die Marke dieses Slice rechnet der Planner nach der Formel vor der Änderung.

**6. Entscheidungsfragen an den User** (Ausgang offen, Adresse User, §6):

- **F1 Formel.** (a) V0, Faktoren und Zuschlag behalten (empfohlen). (b) V1. (c) V2. (d) V0 mit Zuschlag 1,40.
- **F2 Untere Schwelle von Nr. 9** (bei V0). (a) 0,45 (empfohlen). (b) 0,40. (c) entfällt: die Messpunkt-Zeilen bleiben in den Plänen, eine Meldung nach unten gäbe es nicht mehr; dann verschwände eine Drift aller Slices nach unten ohne Anlass.
- **F3 Slices ohne Test-Code** (Auftrag 5). (a) keine zweite Klasse; Geltungsgrenze in Nr. 2, sonst kein Block (empfohlen). (b) zweite Klasse `Handlauf` in Spalte *Klasse* mit einem Faktor aus zwei Punkten (nicht empfohlen: Änderung von Nr. 3 e, Programm und Fällen, ein Faktor aus 35 und 148 Zeilen).
- **F4 Handoff und Ende in Nr. 9.** (a) die Spanne Handoff → Ende (Median 1,16, p90 1,66) in Nr. 9 nennen, kein neuer Auslöser (empfohlen). (b) zusätzlich ein Hinweis, wenn der Handoff ab 0,70 der Marke liegt (2 Treffer in 20 Paaren, einer davon reißt die Marke; nicht empfohlen: eine Hürde ohne Träger, getragen von zwei Punkten).
- **F5 Laufende Marken** (nur bei V1 oder V2). (a) der Planner ersetzt die Marke je Plan mit `ersetzt` (empfohlen für diesen Fall). (b) Übergangsregel. (c) warten.

**7. Akzeptierte Negative des Vorschlags** (Grund und Folge, für `SPEC-128` Nr. 10 a in Teil 2). (1) Die Streuung bleibt (p10 bis p90 von Ende ÷ Schätzung 0,64 bis 1,20): die Marke ist eine Hülle, keine Prognose; Folge: etwa einer von 30 Slices reißt sie und trägt einen Eintrag `ueberschreitung`. (2) Keine Mindestmarke für kleine Schätzungen: ein Punkt (`bootstrap-m4-konfig`) trüge sie, zwei weitere kleine Punkte (`fallback-policy` 1,29, `gegenproben-zeitbudget` 0,24) zeigen die Gegenrichtung. (3) Die Punkte sind von Hand aus Plänen gelesen, 20 tragen Handoff und Ende, zehn haben zurückgerechnete *z* und *a*, und die Erwartungen des Produktcodes stammen von verschiedenen Händen; ein Fehler in einer Zeile verschiebt keine Kennzahl über die zweite Stelle, ein systematischer Fehler der Lesart wäre über die Tabelle oben nachprüfbar. (4) Die Punkte sind nicht unabhängig (Slices derselben Welle, derselbe Architect); die Prüfung *alt* gegen *neu* ist eine zeitliche Trennung, keine Stichprobe.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect in Teil 2, nach dem Beschluss): folgt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Closure-Notiz `welle-mock-results`); der Vorschlag des Architect liegt vor und der User hat die Faktoren beschlossen (ohne Beschluss kein Start); kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln). Reihenfolge der Kette seit 2026-11-08: `slice-gate-integration-determinism`, `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, dieser Slice. Es laufen höchstens drei Slice-Worktrees zugleich. Der Stempel der Gegenproben (`harness/gegenproben.stempel`) ist an diesen Stand gebunden: der Slice ändert Eingaben der Gegenproben und fährt `make gegenproben` einmal.

**Start belegt (2026-11-20, Planner):** `welle-mock-results` liegt in `done/`; `in-progress/` enthielt nur `roadmap.md`, kein anderer Slice der Kette läuft (`slice-gate-integration-determinism`, `slice-guard-schreibwege`, `slice-regeln-verkoerpern` und `slice-gates-ohne-worktrees` liegen in `done/`, `slice-regeln-verkoerpern` seit dem Merge 809ef414); `git worktree list` nannte vor dem Start nur den Hauptbaum, der Worktree dieses Slice ist der erste (Verzeichnis odata-recorder-wt-groessenformel neben dem Hauptbaum, Zweig `slice-groessenformel-neu`, Basis `main`). **Ausnahme von der Start-Bedingung, auf Anweisung des Users vom 2026-11-20 („slice-groessenformel-neu starten“):** der Beschluss des Users über die Faktoren liegt nicht vor (der Vorschlag V0 und V1 aus `slice-mock-lesen` ist offen beim User, siehe §1 *Stand der Eingabe*). Der Start gilt als Auftrag, den Vorschlag zu überarbeiten; die Bedingung „ohne Beschluss kein Start“ wirkt als **kein Vertragstext und kein Code vor dem Beschluss**: der Architect liefert zuerst nur den Vorschlag (§3, Teil 1), der User beschließt, danach folgt Teil 2. Bleibt der Beschluss aus, gilt die Rückführung `in-progress` → `open` unten.

Gegenstände und Abhängigkeiten, am Start gelesen: Register `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` und `BEO-REPO/auslastung-am-handoff-als-endwert-gelesen` (Gegenstand, *geplant*; `state.md` am Start auf `in-progress/` nachgezogen), berührend `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (*verkörpert*) und `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (*geplant* in `slice-coverage-gate`); `welle-hybrid-results` (*Folge-Slices*: dieser Slice ist Träger der Neukalibrierung, der Datenpunkt von `slice-bootstrap-m4-konfig` steht im Register); `slice-regeln-verkoerpern` (schließt die Formel aus, nennt diesen Slice als Nehmer; kein offener Punkt dort ist an diesen Slice adressiert) und `slice-guard-schreibwege` (§4: eine Messung weit unter der Marke sei ein Befund für diesen Slice; die Messung lag bei 0,66, der Fall trat nicht ein, die Frage der Skripte steht als Auftrag 5); [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) hat den Status `Proposed` (Befund am Start: die ADR, die `SPEC-128` schärft, ist nicht angenommen; sie trägt keine Zahl der Formel, `AGENTS.md` §3.8, der Slice ändert sie nicht, die Annahme ist Entscheidung des User, §6). Dritte Nehmer: `slice-gegenproben-eichung` (`open/`, Kette, sechster) berührt dieselben Eingaben der Gegenproben und den Stempel und folgt diesem Slice (Nr. 7 der Übergangsregeln).

**Größenmarke:** noch nicht gesetzt, Prosa, kein Block; der Architect ist zuerst dran (Vorschlag), danach der Beschluss des Users. Die Marke setzt der Planner vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14; die Schätzung kommt aus der Tabelle des Architect. Erwartet ist ein kleiner bis mittlerer Slice (Konstanten, Randwerte, Fälle der Gegenproben). Wegen der Messung am Ende (`E`) rechnet der Planner die Marke dieses Slice nach der Formel vor der Änderung.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): der Beschluss des Users über die Faktoren fehlt oder wird zurückgezogen.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, `make gegenproben` grün, Closure-Notiz in §7 mit Lerneintrag (die neue Formel mit `liegt in` `SPEC-128` und Anker, oder eine benannte Lücke zur Aussagekraft der Messpunkte).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Laufende Marken zeigen nach der Änderung „Marke ungleich Formel“ und färben `make gates` rot — **Ausgang:** entfallen bei V0 (die Formel bleibt, keine Marke ändert sich); bei V1 oder V2 gilt Weg (a) (Vorschlag §3, Abschnitt 4, Frage F5); Stand am Start 2026-11-20: kein anderer Plan liegt in `in-progress/`, das Gate hat nichts zu messen (§1 *Auswirkung auf laufende Marken*). Adresse: User (F5, nur bei V1 oder V2).
- Der Beschluss des Users liegt bei Start nicht vor (Start auf Anweisung des Users, §4 *Start belegt*) — **Ausgang:** weiter offen; Adresse: User (Beschluss nach dem Vorschlag des Architect, Teil 1, §3: F1 Formel, F2 untere Schwelle, F3 Slices ohne Test-Code, F4 Handoff und Ende in Nr. 9, F5 laufende Marken; Empfehlung V0, 0,45, keine zweite Klasse, Spanne ohne Auslöser; vor Vertragstext, Marke und Code).
- Der Zuschnitt des Slice hängt vom Beschluss: bei V0 entfallen `tools/groessenmarke/`, die Fälle 38 bis 120 von `SPEC-102`, der Stempel und `make gegenproben`, DoD-Punkte 1 (zweiter Teil), 2 und 3 und die Marke dieses Slice (kein Block, Text-Slice) — **Ausgang:** offen; Adresse: Planner (nach dem Beschluss DoD, §3 und §4 anpassen, `AGENTS.md` §3.9; kein Block für diesen Slice wäre eine Entscheidung nach `SPEC-128` Nr. 3 p mit Grund in §4).
- [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Größenmarke als Gate, schärft `SPEC-128`) hat den Status `Proposed`; der Slice ändert die Zahlen der Formel, die die ADR nicht trägt — **Ausgang:** weiter offen; Adresse: User (die ADR annehmen oder bewusst offen lassen; keine Bedingung dieses Slice).
- Slices ohne Go-Code als Messpunkte der Formel (Auftrag 5) — **Ausgang:** Vorschlag liegt vor (§3, Abschnitt 2: keine zweite Klasse, Geltungsgrenze in Nr. 2, sonst kein Block); weiter offen; Adresse: User (F3).
- Neue Messpunkte seit `welle-mock` (Auftrag 6) tragen die Faktoren des Beschlusses; der Beschluss des Users bezieht sich auf einen anderen Stand als V0 und V1 — **Ausgang:** entfallen: der Vorschlag rechnet an 62 Punkten (61 in den Kennzahlen), nennt Klasse je Punkt und trennt Handoff und Ende (20 Paare); der User beschließt auf diesem Stand.
- Die Messpunkte (14, ein Slice mit 0,24) tragen eine Formel nicht, die für künftige Slices anderer Art gilt — **Ausgang:** weiter offen; Adresse: Architect (Teil 2: die Grenze der Herkunft in `SPEC-128` Nr. 2 und Nr. 10 a nach dem Wortlaut von §3, Abschnitt 7).
- Die Gegenproben tragen die alten Zahlen an mehreren Stellen — **Ausgang:** entfallen bei V0 (die Fälle tragen die Zahlen der Formel, V0 ändert keine, und die Schwelle von Nr. 9 steht in keinem Fall); bei V1 oder V2 weiter offen: der Architect nennt in Teil 2 die Fälle (`SPEC-102`), der Fänger ist die Mutation „alte Faktoren“.
- Raster der Randformen (Auftrag 3) ist noch leer — **Ausgang:** weiter offen: der Architect füllt es vor dem ersten Code-Commit (`AGENTS.md` §3.16).
- Größe: die Marke steht als Prosa in §4 — **Ausgang:** offen; Adresse: Planner (vor dem ersten Code-Commit).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `welle-mock` (Entscheidung des Users vom 2026-11-08); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (28 Belege, *geplant*) und `BEO-REPO/auslastung-am-handoff-als-endwert-gelesen` (4×, *geplant*) sind der Gegenstand; `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (19) und `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (7×) berühren Marke und Gegenproben.
