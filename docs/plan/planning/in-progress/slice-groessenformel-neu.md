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

**Eingabe:** der Vorschlag des Architect zur Neukalibrierung (Formel behalten und die Schwelle von Nr. 9 senken, oder neue Faktoren; Vorlage im Plan von `slice-mock-lesen`, Abschnitt Meldung an den Architect, Vorschlag V0 und V1, Entscheidung des Users vom 2026-11-06, Auftrag an den Architect aus der Closure von `welle-mock`). Der Architect liefert den überarbeiteten Vorschlag zum Zeitpunkt des 2026-11-08; der Verweis auf seinen Ort (Bericht oder Anhang zu `SPEC-128`) trägt der Planner hier nach, sobald er vorliegt. Bis dahin ist die Zahl offen.

**Stand der Eingabe am Start (Planner, 2026-11-20):** Der Vorschlag V0 und V1 (16 Messpunkte, Empfehlung V0: Formel behalten, untere Schwelle von Nr. 9 von 0,60 auf 0,45) steht im Plan von `slice-mock-lesen` (Abschnitt Meldung an den Architect, Zeilen um 429 und 474), der seit der Archivierung von `welle-mock` im Archiv `done/welle-mock/archiv.zip` liegt (der Stub unter `done/welle-mock/` trägt ihn nicht); er rechnet nur bis zu den 16 Punkten von `welle-mock`. Seither sind Messpunkte dazugekommen, zwei davon von anderer Art als alle bisherigen: `slice-gate-integration-determinism` (Messung 35 von Marke 1500, Auslastung 0,023, Skripte und Handläufe ohne Go-Code), `slice-guard-schreibwege` (1707 von 2600, Skripte ohne Go-Tests), `slice-expand-parser` (2268 von 3500, 0,65), `slice-gegenproben-zeitbudget` (271 von 1400, 0,19) und `slice-bootstrap-m4-konfig` (Ende ÷ Handoff 2,58); `slice-regeln-verkoerpern` trägt keinen Block (kein Go-Code, Diff 156 Zeilen, Begründung in §4 dort: ein Slice ohne Go-Code sei kein Messpunkt der Faktoren). Die Register-Einträge führen am Start 30 Dateien unter `evidence/` (`schaetzformel-ueberschaetzt-den-slice`, die Zahl 28 oben ist der Stand der Closure von `welle-mock`) und 5 (`auslastung-am-handoff-als-endwert-gelesen`). Ein Beschluss des Users über V0, V1 oder eine dritte Form liegt nicht vor.

**Messpunkte nach den 14 (Eingabe, Geber `slice-gegenproben-zeitbudget`):** `Messpunkt slice-gegenproben-zeitbudget.md: zeilen=9 nebenlaeufig=0 fakes=2 produktcode=15 schaetzung=1135 marke=1400 messung=271` (Stand `Ende`, Auslastung rund 0,19 der Marke, unter der Schwelle 0,60 von `SPEC-128` Nr. 9; Plan §4 dort; der Slice ist ein Werkzeug-Slice mit einem Test von 175 Zeilen und 15 Zeilen Produktcode).

**Auswirkung auf laufende Marken.** Eine Marke, die der Planner vor dem ersten Code-Commit nach der alten Formel setzte, bleibt in ihrem Plan stehen; das Gate `make groessenmarke` rechnet die Marke aus dem Block des Plans (`SPEC-128` Nr. 3 p, Nr. 5), nicht aus einer Angabe. Gilt nach dem Beschluss eine andere Formel, so zeigt `make groessenmarke` für einen Plan in `in-progress/` mit der alten Marke „Marke ungleich Formel“ und färbt rot. Der Beschluss nennt deshalb vor dem Code, wie laufende Pläne behandelt werden: (a) der Planner ersetzt die Marke mit Grund in §4 jedes Plans in `in-progress/` (AGENTS.md §3.14: „eine ältere ersetzt er mit Grund in §4“), (b) die Formel trägt eine Übergangsregel nach Stichtag oder Plan-Datum, (c) die Änderung wartet, bis kein Plan mit Block in `in-progress/` liegt. **Stand am Start 2026-11-20:** außer diesem Plan liegt kein Plan in `in-progress/`, kein Plan dort trägt einen Block, und kein Slice-Worktree läuft (`git worktree list` nennt nur den Hauptbaum): „Marke ungleich Formel“ trifft heute keinen Plan, die Wege (a) und (b) sind gegenstandslos, solange kein anderer Slice vor dem Merge dieses Slice startet; startet einer, geht es nach Auftrag 2. Die Pläne in `open/` führen die Marke als Prosa (kein Block, das Gate misst sie nicht); der Planner rechnet sie beim jeweiligen Start neu. Stand 2026-11-08 (überholt; die drei Pläne liegen in `done/`): Pläne mit Marke in `in-progress/` waren `slice-fallback-policy` (S 1900 / M 2600), `slice-etag-versionen` (S 1910 / M 2600, vorläufig) und `slice-admin-lebenszyklus` (S 1615 / M 2200), alle als Prosa; kein Block, also misst das Gate sie nicht (`SPEC-128` Nr. 3 p). Vorläufige Marken der Pläne in `open/` (z. B. `slice-gate-integration-determinism`) rechnet der Planner nach dem Beschluss neu.

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

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect): folgt.

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

- Laufende Marken zeigen nach der Änderung „Marke ungleich Formel“ und färben `make gates` rot — **Ausgang:** weiter offen: der Architect wählt die Behandlung vor dem Code (Auftrag 2); Stand am Start 2026-11-20: kein anderer Plan liegt in `in-progress/`, das Gate hat nichts zu messen (§1 *Auswirkung auf laufende Marken*).
- Der Beschluss des Users liegt bei Start nicht vor (Start auf Anweisung des Users, §4 *Start belegt*) — **Ausgang:** weiter offen; Adresse: User (Beschluss der Faktoren nach dem Vorschlag des Architect, Teil 1; vor Vertragstext, Marke und Code).
- [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Größenmarke als Gate, schärft `SPEC-128`) hat den Status `Proposed`; der Slice ändert die Zahlen der Formel, die die ADR nicht trägt — **Ausgang:** weiter offen; Adresse: User (die ADR annehmen oder bewusst offen lassen; keine Bedingung dieses Slice).
- Slices ohne Go-Code als Messpunkte der Formel (Auftrag 5) — **Ausgang:** weiter offen; Adresse: Architect (Vorschlag, Teil 1).
- Neue Messpunkte seit `welle-mock` (Auftrag 6) tragen die Faktoren des Beschlusses; der Beschluss des Users bezieht sich auf einen anderen Stand als V0 und V1 — **Ausgang:** weiter offen; Adresse: Architect (Vorschlag nennt Zahl und Klassen der Punkte).
- Die Messpunkte (14, ein Slice mit 0,24) tragen eine Formel nicht, die für künftige Slices anderer Art gilt — **Ausgang:** weiter offen: der Vorschlag nennt die Grenze der Herkunft in `SPEC-128` Nr. 2.
- Die Gegenproben tragen die alten Zahlen an mehreren Stellen — **Ausgang:** weiter offen: der Architect nennt die Fälle (`SPEC-102`), der Fänger ist die Mutation „alte Faktoren“.
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
