# Slice slice-unsichtbare-zeichen: Prüfung auf unsichtbare Zeichen in Quelltext, Werkzeugen und Dokumenten

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-sensoren](../welle-sensoren.md) — erster Slice der Folge, erster von Teil 1; Start, wenn `slice-groessenmarke-werkzeug` in `done/` liegt (Trigger von Teil 1). Er liegt vor `slice-workspace-use-case` (Reihenfolge, kein Blocker, [welle-sensoren](../welle-sensoren.md) §5).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe: Container, netzlos; das Lastenheft führt keine Anforderung an Gates), `AGENTS.md` §3.18 (die Regel, deren Teil das Gate trägt), Register-Eintrag `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (4×, verkörpert in `AGENTS.md` §3.18), Retirement-Check `docs/reviews/2026-10-25-retirement-check-hard-rules.md` §5, [ADR-0003](../../adr/0003-lint-profil-golangci.md) (Lint-Profil; eigene Prüfungen in `tools/harness/lint.sh`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenproben außerhalb der Gate-Kette).

**Berührte Spec-Stellen:** [SPEC-233](../../../../harness/vertraege/SPEC-233-unsichtbare-zeichen.md) (neuer Vertrag `make zeichen-check`, Kennung nach `harness/vertraege/README.md` §Kennungen, Text des Architect, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)) · [SPEC-102](../../../../harness/vertraege/SPEC-102-gegenproben.md) (Fälle 131 bis 155 stehen; Zeitbudget, Grenze, Stand und die übrigen Nachzüge im Commit des Implementers) · `harness/README.md` §Sensors (Zeile des Ziels, Zeile von `make gegenproben`, *Nicht behauptet*; im Commit des Implementers). `SPEC-101` bleibt unverändert: `lint.sh` ist nicht beteiligt, `tools/zeichen` fällt als Modul unter `tools/` unter dessen Regel, sobald die Stufe `lint` es kopiert. `spec/spezifikation.md` und die Architektur-Sicht bleiben unberührt: der Slice liefert kein Anwendungscode.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Gate `make zeichen-check` ([SPEC-233](../../../../harness/vertraege/SPEC-233-unsichtbare-zeichen.md)) in der Kette `make gates` wird rot, wenn eine Textdatei des Arbeitsbaums außer `.harness/baseline/` ein Zeichen enthält, das dort nicht hingehört (BOM, NBSP, U+2028, U+2029, U+200B und rohe Steuerzeichen außer Zeilenvorschub und Tabulator, dazu der Wagenrücklauf; als Klasse mit 253 Zeichen, Nr. 7), oder ein Byte, das kein gültiges UTF-8 ist, und weist mit einer Gegenprobe nach, dass es rot wird; `AGENTS.md` §3.18 verliert damit den Teil, den ein Gate fangen kann.

**Der Teil, den das Gate nicht fängt, ist benannt:** Die Regel §3.18 beschreibt zwei Wirkungen des Werkzeugs: eine Folge aus Backslash, `u` und vier Hex-Ziffern wird still in das **sichtbare** Zeichen umgesetzt (aus der Folge für die öffnende eckige Klammer wird das Zeichen `[` selbst), und Rohbytes wie BOM oder NBSP gehen im Bericht verloren oder entstehen unbemerkt. Das Gate sieht nur, was in der Datei steht: ein unsichtbares Zeichen, das dort nicht hingehört. Die Umsetzung zu einem sichtbaren Zeichen, das der Autor als Text meinte, hinterlässt eine gültige Datei; kein Zeichenfilter erkennt sie. Dieser Teil bleibt Prosa (`AGENTS.md` §3.18) und Gegenlesen; die Closure nennt ihn in der Tabelle der Welle als „Prosa bleibt“.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung von `AGENTS.md` §3.18 — `AGENTS.md` ändert nur der User; der Slice liefert in der Closure einen Diff-Vorschlag (Zeiger auf das Gate, Kern der Regel bleibt für den nicht gefangenen Teil).
- Die Erkennung der Umsetzung `\u…` → sichtbares Zeichen (siehe oben) — nicht entscheidbar aus dem Ergebnis; ein Fänger dafür wäre ein Test, der den Wert mit dem Wert vergleicht, den der Autor meinte, und den kennt nur der Test selbst.
- Zeichen, die in Dokumenten, Kommentaren und Zeichenketten legitim stehen (deutsche Umlaute, typografische Striche, ein `§` und Ähnliches, UTF-8 überhaupt) — die Prüfung gilt einer festen Menge unsichtbarer Zeichen (Menge Z, `SPEC-233` Nr. 7), keiner Zeichenklasse „kein ASCII“; ihre Grenze (Variationsselektoren, Füllzeichen, Homoglyphen) steht im Vertrag unter *Ungeprüft* (Nr. 16 b).
- Die vendorte Baseline unter `.harness/baseline/` und fremde Bäume — der Baum ist Kurs-Inhalt und wird byte-genau mitgeliefert (`harness/conventions.md` §Baseline); der Teilbaum `.harness/baseline` fehlt in der Prüfmenge des Programms (`SPEC-233` Nr. 5 a), sonst nichts.
- Eine Direktive im geprüften Text, die das Gate umgeht (`AGENTS.md` §3.2), und eine Ausnahme-Liste oder -Konfiguration — Testfixtures mit Absicht entstehen aus Bytewerten (`esc("…")`, `[]byte{…}`) und brauchen keine Ausnahme; der Architect hat entschieden, dass es keine Ausnahme gibt (`SPEC-233` Nr. 5 f); eine künftige Ausnahme wäre ein Carveout (`AGENTS.md` §3.6).

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Das Gate `make zeichen-check` (Programm `zeichen`, Stufe `zeichen` des `Dockerfile`, Treiber `tools/harness/zeichen.sh`, Fragment `harness/mk/zeichen.mk`; Auftrag 1 entschieden) hängt an `GATE_CHECKS` und wird rot, wenn eine Datei der Prüfmenge ein Zeichen der Menge Z oder ungültiges UTF-8 enthält; die Prüfmenge und Z stehen in `SPEC-233`, jede Zusage des Vertrags hat eine Zeile in der Tabelle *Zusage · Fänger* (§3) mit einer vom Implementer selbst rot gesehenen Mutation. <!-- d-check:ignore (harness/mk/zeichen.mk entsteht mit dem Code des Implementers) -->
- [ ] Das Gate ist durch eine Gegenprobe nach [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) belegt (die Fälle 131 bis 155 von `SPEC-102`; Fall 132 führt eine Datei je Zeichen von Z, Fixtures aus Bytewerten, auf Rohbytes gegengelesen); die Zusatzzeit der Gegenproben ist gemessen und in §7 berichtet.
- [ ] Der Bestand ist sauber oder berichtigt: die Datei der Prüfmenge, die ein Zeichen von Z heute trägt (Messung des Architect am Stand `7cd9c60`: 1 Datei, §3), ist im Slice berichtigt (keine Ausnahme-Liste ohne Carveout, `AGENTS.md` §3.6); die Zahl der berichtigten Dateien steht in §7.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Zeile des Ziels in `harness/README.md` §Sensors, Zeile von `make gegenproben` und *Nicht behauptet* (im selben Commit wie das Ziel; Wortlaut in §3, Abschnitt *Nachzüge und Wortlaute*) und die Nachzüge in `SPEC-102` (Festlegung, Aufbau, Eingaben, Randformen 2, 12 und 13, Pflicht zum Fahren, Stand, Grenze, Zeitbudget); die Zeile im Index `harness/vertraege/README.md` steht; die Closure legt den Diff-Vorschlag für `AGENTS.md` §3.18 bei (der User ändert, kein Rollenlauf).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

Der Planner schneidet; der Architect entscheidet (Aufträge unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/vertraege/SPEC-233-unsichtbare-zeichen.md` und die Zeile im Index dort | geliefert (Text des Architect, vom Planner geschrieben) | Prüfmenge, Menge Z, keine Ausnahme, Ausgabe, Exit-Codes, Randformen, *Ungeprüft* (der nicht gefangene Teil von `AGENTS.md` §3.18) |
| `tools/zeichen/` (`go.mod` ohne Abhängigkeiten, Programm, Tests) | neu | `zeichen pruefen <Wurzel>`: liest Bytes, meldet Datei, Zeile, Spalte und das Zeichen als Kennung `U+XXXX` (nie als Rohzeichen); sequentiell (`SPEC-233` Nr. 12 d) |
| `tools/harness/zeichen.sh` | neu | Treiber: `git ls-files`, Tar-Strom, `docker create`/`cp`/`start`/`rm` (`SPEC-233` Nr. 2 bis 4); `#!/usr/bin/env bash`, `set -euo pipefail` |
| `harness/mk/zeichen.mk`, `Dockerfile`, `.dockerignore` | neu und update | Ziel `zeichen-check` in `GATE_CHECKS`; Stufe `zeichen`, Kopie von `tools/zeichen` in die Stufe `lint`, Einträge in der Stufe `gegenproben`; zwei Zeilen der Allowlist (Wortlaute: Abschnitt *Nachzüge und Wortlaute* unten) <!-- d-check:ignore (harness/mk/zeichen.mk entsteht mit dem Code des Implementers) --> |
| `harness/README.md` | update | Zeile des Ziels in §Sensors, Zeile von `make gegenproben` und die Ergänzung in *Nicht behauptet*, im selben Commit wie das Ziel (sonst meldet `make docs-check` ein Target ohne Zeile) |
| `.golangci.yml` | update (Implementer) | der Pfad der Regel für `os.Stdout` und `os.Stderr` im Einstieg eines Programms unter `tools/` nennt `zeichen` neben `gegenproben`, `groessenmarke` und `pruefung`; sonst meldet `forbidigo` `tools/zeichen/main.go` (§6) |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh`, `harness/mk/gegenproben.mk` | update | die Fälle 131 bis 155 von `SPEC-102` (Teil G, zuletzt in eigenem Commit, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); `EINGABEN` des Skripts um `tools/harness/zeichen.sh` und `tools/zeichen`; Kopfkommentar und Hilfe-Zeile des Ziels; Stempel `harness/gegenproben.stempel` nach dem letzten Commit unter den Eingabepfaden |
| Testdateien des Prüfprogramms | neu | eine Zeile der Tabelle unten je Test; Mengen als Schleifen über die Liste (`AGENTS.md` §3.17) |
| Bestand der Prüfmenge | update (1 Datei) | gemessen vom Architect am Stand `7cd9c60` auf Rohbytes (Klassen `Cc`, `Cf`, `Zl`, `Zp`, `Zs`, U+FFFE, U+FFFF; ohne `.git`, ohne `.harness/baseline/`, ohne `archiv.zip`): **1 Datei**, `docs/reviews/2026-10-25-slice-baseline-import-verifikation-3.md`, Zeile 45, zwei U+00A0 in den Spalten 1367 und 1384 (die Beschreibung der Fixtures des Verifiers); Berichtigung: jedes Zeichen im Text durch die Kennung `U+00A0` ersetzen, mit `Edit`, danach auf Rohbytes prüfen; die Messung wiederholt der Implementer am Start. Fällt das Zeichen beim Schreiben still um, steht das in §6 mit Ausgang *offen*, Adresse Planner und User |

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Vertrag, Aufträge 1 bis 8, Tabelle *Zusage · Fänger*, Raster, Gegenproben-Fälle, Nachzüge, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorschlag an den User).

**Betriebsregeln für alle Rollen** (Wortlaut und Gründe: [welle-sensoren](../welle-sensoren.md) §4):

- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen; Docker sonst nur über `make` (`AGENTS.md` §3.1). Gate-Läufe laufen im Hauptbaum, weil die Images feste Tags tragen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- **Dieser Slice ist der Fall von `AGENTS.md` §3.18 selbst:** Fixtures mit BOM, NBSP, U+2028, U+2029, U+200B, rohem Steuerzeichen oder einer Folge aus Backslash, `u` und vier Hex-Ziffern entstehen aus Bytewerten (`[]byte{…}`, `esc("…")`); die geschriebene Datei, der Plan und der Bericht werden mit `env LC_ALL=C grep -aP` auf Rohbytes gegengelesen (nicht mit der Shell-Funktion `grep` des Agenten-Shells; Befund des Architect: sie lenkt auf `ugrep` um und meldet `-P` mit Rohbytes falsch-negativ; `env LC_ALL=C grep -aP` trifft an Zeile 45 des Reports `docs/reviews/2026-10-25-slice-baseline-import-verifikation-3.md` die zwei Zeichen, der Planner hat es nachgemessen), und der Bericht nennt, welche Dateien gegengelesen wurden. Ein Fänger, dessen Fixture beim Schreiben still zu einem anderen Zeichen wird, ist kein Fänger.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen.
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; der Vertrag steht in `harness/vertraege/`, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), eine ADR nur, wenn eine Entscheidung sie verlangt, `AGENTS.md` §3.8):

1. **Ort in der Kette.** Kandidaten: (a) eine eigene Prüfung in `tools/harness/lint.sh` (läuft in der Stufe `lint`, im Modul-Wurzelverzeichnis; sieht nur, was `.dockerignore` in den Kontext lässt, also keine Dokumente und nicht `tools/`), (b) ein eigenes Programm und Ziel in `GATE_CHECKS` nach dem Muster `groessenmarke` (Stufe, `docker cp`, netzlos, ohne Mount), (c) beides, je Prüfmenge eines. Zu entscheiden mit Grund und Folge: welche Prüfmenge welchen Ort braucht, wie die Dateiliste in den Container kommt, und was passiert, wenn ein Pfad nicht in den Kontext fällt (Lücke oder rot). Der Auftrag des Users nennt `src/**/*.go`, `tools/**` und Dokumente; der Architect bestätigt oder grenzt ein.
2. **Die Liste der Zeichen** (BOM, NBSP, U+2028, U+2029, U+200B, rohe Steuerzeichen) — fest oder als Klasse (Unicode-Kategorien `Cf`, `Zl`, `Zp`, `Zs` außer dem Leerzeichen, `Cc` außer Zeilenende und Tabulator): Was ist mit U+200C, U+200D, U+2060, dem weichen Trennstrich U+00AD, den Richtungszeichen U+202A bis U+202E und U+2066 bis U+2069, dem Wagenrücklauf (CRLF-Dateien), U+007F und U+0080 bis U+009F? Eine Klasse ist eine Menge, für die der Fänger eine Schleife über alle Elemente ist (`AGENTS.md` §3.15, §3.17); eine feste Liste nennt ihre Grenze als akzeptiertes Negativ mit Grund und Folge. BOM am Dateianfang und in der Mitte sind zwei Randformen.
3. **Prüfmenge und Ausnahmen.** Welche Dateien (Endungen, Pfade, Binärdateien wie die `archiv.zip` unter `done/`, die vendorte Baseline, Daten-Fixtures mit absichtlichem BOM), und wie eine Ausnahme ausgedrückt wird, ohne Marker im geprüften Text (`AGENTS.md` §3.2) und ohne stille Senkung (`AGENTS.md` §3.6): Eintrag im Vertrag, mit Grund. Eine leere Prüfmenge ist rot oder grün, mit Grund.
4. **Ausgabe und Ausgang.** Form der Meldung (Datei, Zeile, Spalte, Kennung `U+XXXX`, nie das Rohzeichen), Exit-Codes (Befund, Aufrufstelle, nicht lesbare Datei), Reihenfolge bei mehreren Befunden an einer Stelle, Verhalten bei einer Datei, die kein gültiges UTF-8 ist.
5. **Gegenproben.** Fälle zu `SPEC-102` (Nummern fortlaufend nach dem letzten Fall; je Zeichen der Liste und je Randform ein Fall), die Frage, ob ein Fall das Produkt mitfährt (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`: die Eingabe enthält keine Pakete des Produkts), und die Zusatzzeit im Abschnitt *Zeitbudget* von `SPEC-102`; der Architect schätzt sie, der Implementer misst.
6. **Tabelle *Zusage · Fänger*** (§3), **Raster** (§6, jede Zelle mit Ort in der Spezifikation oder im Vertrag oder „gilt nicht“ mit Grund) und die **Paare** von Verstößen am selben Ort (zwei Zeichen derselben Zeile, dasselbe Zeichen mehrfach, ein Zeichen in zwei Dateien).
7. **Nachzüge** nach `AGENTS.md` §3.9: `harness/README.md` (Sensors, *Nicht behauptet*), `SPEC-101` oder `SPEC-102`, wo der Ort sie berührt; ob eine ADR zum Gate nötig ist (Muster: Bindung-Spalte in `harness/README.md`).
8. **Erwartete Größe** des Produktcodes (Zeilen des Prüfprogramms ohne Tests) für die Marke des Planners; ob das Prüfprogramm nebenläufig ist (zählt × 85).

**Entscheidungen des Architect zu den Aufträgen 1 bis 8** (Ort im Vertrag `SPEC-233`; keine neue ADR: kein Accepted-ADR wird berührt, die Wahl folgt dem Muster von [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) — der Host wählt, ein Programm im Container urteilt —, die Bindung in `harness/README.md` ist [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)):

| Auftrag | Entscheidung | Ort |
|---|---|---|
| 1 Ort in der Kette | Variante (b): eigenes Programm, Stufe, Treiber und Ziel. Variante (a) und (c) entfallen: eine Prüfmenge, ein Ort. Die Dateiliste kommt per `git ls-files` und Tar-Strom über `docker cp -` in den Container. Ein Pfad außerhalb des Kontexts ist weder Lücke noch Rot: die Prüfmenge ist, was der Treiber liefert, und eine leere Menge ist Exit 2. Der Auftrag des Users (`src/**/*.go`, `tools/**`, Dokumente) ist bestätigt und erweitert auf alle Dateien des Arbeitsbaums außer `.harness/baseline/` und Binärdateien. Die Rückführung aus §4 („Dokumente nicht erreicht“) tritt nicht ein. Die Host-Auswahl braucht nur dieses Gate: `slice-guide-groessen-gate` läuft im d-check-Modul mit eigenem Mount, `slice-abdeckung-deklaration` liest in der Stufe `source`. | Nr. 1, 2, 3 |
| 2 Liste der Zeichen | Eine Klasse, keine Aufzählung: `Cc` ohne Tabulator und Zeilenvorschub (63), `Cf` (170), `Zl`, `Zp`, `Zs` ohne U+0020 (16), dazu U+FFFE und U+FFFF; 253 Zeichen, Unicode 17.0.0. Entscheidungen im Einzelnen: U+200C bis U+200F, U+2060 bis U+2064, U+00AD, U+180E, die Richtungszeichen und die Tag-Zeichen sind alle in `Cf`; U+0085 und U+0080 bis U+009F sowie U+007F in `Cc`; der Wagenrücklauf rot (LF-Dateien); das BOM an jeder Stelle. Eine Klasse mit Elementzahl macht den Fänger zur Schleife. Eine feste Liste ließe U+202F und U+3000 offen, die Schreibwerkzeuge gern erzeugen. Grenze und Grund: Nr. 16 b (Variationsselektoren, Füllzeichen, Homoglyphen). | Nr. 7, 16 b |
| 3 Prüfmenge, Ausnahmen | Ausgenommen sind allein `.harness/baseline/` (Teilbaum), Symlinks und andere Einträge, Binärdateien (NUL in den ersten 8000 Bytes, Regel von `git`); `archiv.zip` fällt darunter. Die Binärerkennung kennt eine Rückausnahme für UTF-16-BOM. **Keine Ausnahme-Möglichkeit:** keine Option, Konfiguration oder Marker (`AGENTS.md` §3.2). Eine künftige Ausnahme ist ein Carveout und ein Eintrag in Nr. 5 f (`AGENTS.md` §3.6). NBSP ist überall rot, auch in Dokumenten; in Prosa steht `U+00A0`. Eine leere Prüfmenge ist rot (Exit 2). | Nr. 2, 5, 6, 11 b |
| 4 Ausgabe und Ausgang | Befund: `zeichen: <Pfad>:<Zeile>:<Spalte>: U+HEX Kategorie [Kurzname]`. Die Ausgabe ist ASCII-maskiert, Zeichen nur als Kennung. Exit 0, 1 (Befund), 2 (Aufruf, Eingabe, Lesefehler, leere Menge; 2 geht vor 1). Ungültiges UTF-8 ist ein Befund je Byte. Reihenfolge: Pfad bytegleich, dann Stelle. | Nr. 8 bis 12 |
| 5 Gegenproben | `SPEC-102` Fälle 131 bis 155 (25 Fälle). Programm-Fälle 131 bis 150, Treiber-Fälle 151 bis 155. Die Eingabe enthält kein Paket des Produkts. Zusatzzeit: Schätzung 0 bis 14 s im Gesamtlauf, Absatz *Zeitbudget* der Nachzüge unten. | `SPEC-102`, `SPEC-233` Gegenprobe |
| 6 Tabelle, Raster, Paare | Tabelle und Mengen unten, Raster und Zusatztabelle in §6. | §3, §6 |
| 7 Nachzüge, ADR | Abschnitt *Nachzüge und Wortlaute* unten. Keine ADR (Begründung oben). Die Bindung in `harness/README.md` ist [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md). | §3 |
| 8 Größe | a = 0 (sequentiell; Nr. 12 d), b = 36, f = 1, p = 480. §4. | Nr. 12 d |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Text des Architect; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Klasse: alle Zeilen `andere` (die Signal-Läufe von Zeile 7 sind deterministisch: der Ersatz sendet das Signal und endet; zählte man Zeile 7 als nebenläufig, wäre S = 3565 und die Marke bliebe 4900). Zeilen und Spalten sind frei von `|` im Zelleninhalt.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Das Ziel `zeichen-check` hängt an `GATE_CHECKS`; die Stufe `zeichen` prüft (gofmt, vet, test) und baut ohne Netz; der Treiber ruft sie (Nr. 1) | akzeptiertes Negativ: kein Fall der Gegenproben fährt `make` oder `docker build`, das Doku-Gate sieht nur die Zeile in `harness/README.md`; Folge: fehlte die Verdrahtung, bliebe `make gates` grün, ohne dass ein Zeichen geprüft würde; Beleg beim Handoff: `make -n gates` nennt `--target zeichen` und `tools/harness/zeichen.sh`, und `make zeichen-check` ist rot auf einem Baum mit eingepflanztem NBSP (Nr. 16 h) | Zeile `GATE_CHECKS += zeichen-check` entfernt; Stufe `zeichen` ohne `go test`; `COPY tools/zeichen` in der Stufe `lint` entfernt | andere |
| 2 | Der Treiber übergibt jede Datei, die `git ls-files -z --cached --others --exclude-standard` nennt und die im Arbeitsbaum existiert, und nichts sonst (Nr. 2 a bis c, e) | Fall 151: acht Arten Pfad (getrackt, ungetrackt, ignoriert, im Arbeitsbaum gelöscht, Name mit Leerzeichen, Name mit Zeilenumbruch, Symlink samt hängendem, mitgelieferte Datei unter `.harness/baseline/`); der Tar-Strom, den der Ersatz für `docker` festhält, nennt genau die Pfade der Erwartung | `--others` entfernt; `--exclude-standard` entfernt; Filter der gelöschten Pfade entfernt; Liste ohne `-z`; `.git` in die Liste | andere |
| 3 | Der Treiber endet mit 2 und ruft `docker` nicht, wenn kein Git-Arbeitsbaum, kein `git`, keine Datei oder ein Argument vorliegt (Nr. 2 d) | Fall 155: vier Läufe (Verzeichnis ohne Repo, Repo ohne Datei, `PATH` ohne `git`, ein Argument); der Zähler des Ersatzes für `docker` steht auf 0 | Prüfung auf leere Liste entfernt; `kein Git-Arbeitsbaum` mit Exit 0 | andere |
| 4 | `docker` wird genau in der Folge `create --network none <Image> pruefen /work`, `cp - <Container>:/work`, `start -a <Container>`, `rm -f <Container>` gerufen, je einmal, ohne Mount und ohne `run` (Nr. 3 a) | Fall 152: das Protokoll des Ersatzes hält jeden Aufruf mit allen Argumenten fest und zählt auch `run`, `--volume` und `--mount`, die der Treiber nie rufen darf; ein Lauf mit gesetztem `ZEICHEN_IMAGE` | `--network none` entfernt; `run -v` statt `create`; anderes Ziel als `/work`; `ZEICHEN_IMAGE` ignoriert; `start` ohne `-a`; `rm` entfernt | andere |
| 5 | Scheitern von `create`, `tar` oder `cp` endet mit 2 und der Meldung der Stelle; `start` wird dann nicht gerufen (Nr. 3 c) | Fall 154, Läufe 3 bis 5 (`create` endet mit 1, `cp` endet mit 1, `PATH` ohne `tar`) | `pipefail` entfernt (Fehler von `tar` unbemerkt); Exit der Übergabe verworfen | andere |
| 6 | Der Exit des Treibers ist 0 für 0, 1 für 1 und 2 für jeden anderen Exit von `start` (Nr. 4 a) | Fall 153: zehn Läufe (0, 1, 2, 3, 125, 126, 127, 130, 137, 255) | Durchreichen des rohen Exit (3 bleibt 3, 255 bleibt 255); 1 auf 0 abgebildet | andere |
| 7 | Ein entstandener Container wird in jedem Fall entfernt (Ende 0, Exit 1, Fehler von `cp` und `tar`, HUP, INT, TERM mit Exit 2); bei Fehler von `create` gibt es kein `rm` (Nr. 4 b) | Fall 154: acht Läufe; das Protokoll zählt `rm -f` je Lauf | `trap` entfernt; `rm` nur im Erfolgsfall; Signal ohne Exit 2 | andere |
| 8 | Z hat 253 Elemente in den Teilmengen 63, 170, 1, 1, 16 und 2 nach Unicode 17.0.0; jedes andere Zeichen liegt außerhalb (Nr. 7 a, b) | `TestMengeZ`: Durchlauf über alle 1 112 064 Skalarwerte (U+0000 bis U+10FFFF ohne U+D800 bis U+DFFF) gegen die Zahlen und `unicode.Version`; die 29 Kurznamen-Zeichen und die 14 Zeichen des Auftrags (U+FEFF, U+00A0, U+2028, U+2029, U+200B, U+0000, U+0001, U+001F, U+007F, U+0080, U+009F, U+000D, U+000B, U+000C) als Anker | `Cf` ausgelassen; `Zs` samt U+0020; Tabulator oder Zeilenvorschub in `Cc`; Bereich `Cc` nur bis U+009E; U+FFFE oder U+FFFF entfernt; `Zl` ausgelassen | andere |
| 9 | Jedes Element von Z ist an jeder Stelle ein Befund mit Zeile und Spalte nach Nr. 9 | `TestStellungen`: Schleife über die 253 Elemente mal 7 Stellungen (Dateianfang, Zeilenanfang, mittig hinter zwei Umlauten, vor dem Zeilenvorschub, Dateiende ohne Zeilenvorschub, einziges Zeichen nach dem letzten Zeilenvorschub, zweimal benachbart) mal 4 Leser (ganz, ein Byte je Aufruf, halbe Puffer, Daten mit EOF; `testing/iotest`); der Leser mit einem Byte je Aufruf trennt jedes Zeichen an jeder Stelle | Prüfung nur am Zeilenanfang; letztes Zeichen der Datei nicht geprüft; nur das erste Zeichen je Zeile; Dekodierung je Leseaufruf ohne Übertrag; Spalte in Bytes | andere |
| 10 | Jedes Element von Z ist in einer Datei des Prozesses ein Befund; die Ausgabe nennt jedes genau einmal (Nr. 7, 10 a) | Fall 132: 253 Dateien (eine je Element; für U+0000 hinter 8000 Bytes), stderr gleich der erwarteten Liste samt Zeile `253 Befund(e) in 253 Datei(en)`, kein Byte jenseits von ASCII in den Befund-Zeilen | Programm kennt nur die sechs Zeichen des Auftrags; Prüfmenge nur `*.go` und `*.md`; Steuerzeichen nur unter U+0020 | andere |
| 11 | Alles andere passiert, auch U+FFFD und die Grenzfälle der Nr. 16 b (Nr. 7 e, 8 b) | `TestNichtZ`: Schleife über 208 Zeichen (95 druckbares ASCII, Tabulator, Zeilenvorschub, 94 Zeichen U+00A1 bis U+00FF ohne U+00AD, 17 benannte: U+2013, U+2014, U+2018, U+2019, U+201C, U+201D, U+2026, U+20AC, U+2192, U+2713, U+FFFD, U+FE0F, U+034F, U+3164, U+2800, U+4E2D, U+1F600) mal 7 Stellungen; Fall 131 (sauber: Exit 0, `4 Datei(en)`) | Klasse „jedes Byte ab 0x80“ (Umlaute rot); Tabulator als Befund; U+FFFD als ungültig; U+FE0F in Z aufgenommen | andere |
| 12 | Das BOM U+FEFF am Dateianfang und in der Mitte ist je ein Befund mit derselben Kennung (Nr. 7 c) | Fall 133 (`1:1` und `2:2`, beide `U+FEFF Cf BOM`) | BOM nur am Dateianfang geprüft; BOM am Dateianfang abgestreift (`TrimPrefix`) | andere |
| 13 | Der Wagenrücklauf ist ein Befund, auch vor dem Zeilenvorschub, und beendet keine Zeile (Nr. 7 d, 9) | Fall 135: CRLF-Zeilen, NBSP vor CR (`1:2`, `2:2`, `2:3`) | CR vor LF abgestreift; CR als Zeilenende gezählt | andere |
| 14 | Nur U+000A beginnt eine Zeile; U+2028, U+2029 und U+0085 stehen in der Zeile; die Spalte zählt Zeichen, nicht Bytes (Nr. 9) | Fall 136 (`1:2`, `1:4`, `1:6`, `2:2`) und Fall 134 (hinter zwei Umlauten Spalte 3, nicht 5) | U+2028 als Zeilenende gezählt; Spalte in Bytes | andere |
| 15 | Je Zeichen von Z nennt die Meldung die Kategorie, und für die 29 Zeichen der Tabelle den Kurznamen (Nr. 7 f, 10 a) | `TestKurznamen`: 29 Zeichen mit Kurzname, 28 verschiedene Namen (U+FFFE und U+FFFF heißen beide NONCHAR, `SPEC-233` Nr. 7 f), ASCII, jedes Zeichen in Z, Kategorie je Element der 253 gleich der Tabelle von `unicode`; Fall 132 liest jede Zeile | Namen von U+2028 und U+2029 vertauscht; Kategorie `Cn` für U+FEFF | andere |
| 16 | Die Ausgabe führt nie ein Zeichen jenseits von ASCII aus einer Datei, einem Pfad oder einem Fehlertext: Pfad und Grund sind maskiert (`\xHH`, `\\`), das Zeichen nur als Kennung (Nr. 10 b) | `TestAusgabeAscii`: Schleife über alle 256 Bytewerte der Maskierung; Fall 150 (sieben Namen: Leerzeichen, Zeilenumbruch, Rückwärtsschrägstrich, Tabulator, Umlaut, U+00A0, Byte 0xFF); Fall 132 | Pfad roh ausgegeben; `\\` nicht maskiert; Grund des Lesefehlers roh | andere |
| 17 | Die Summenzeilen stimmen: `ok` mit *n* geprüft, *b* binär, *s* andere; sonst *k* Befunde in *d* Dateien (Nr. 10 c, d) | Fälle 131 (4, 0, 0), 139 (Zählung binär), 141 (Zählung andere), 149 (*k* und *d*) | *n* und *b* vertauscht; der Teilbaum `.harness/baseline` mitgezählt; *d* zählt Befunde statt Dateien | andere |
| 18 | Jedes ungültige Byte ist ein Befund `ungültiges UTF-8 (Byte 0xHH)`; das Programm ersetzt, überspringt und rät nichts (Nr. 8 a) | `TestUngueltigesUTF8`: Schleife über 90 Elemente (64 Fortsetzungsbytes 0x80 bis 0xBF, C0, C1, 11 Anfänge F5 bis FF, 6 abgebrochene Folgen C2 / E2 82 / F0 9F 98 je am Dateiende und vor LF, 3 überlange E0 80 80 / F0 80 80 80 / C0 AF, 2 Surrogate ED A0 80 / ED BF BF, 2 über U+10FFFF F4 90 80 80 / F7 BF BF BF) mit der Zahl und den Spalten der Erwartung je Element; Fall 137 | Dekodierung über Umwandlung nach `[]rune` oder `strings.ToValidUTF8` (Fehler werden U+FFFD); nur das erste Byte einer ungültigen Folge gemeldet; Fehler am Dateiende verschluckt | andere |
| 19 | Eine Datei mit `FF FE` oder `FE FF` am Anfang ist nie binär und hat genau einen Befund `UTF-16-BOM`; einzelnes `FF`, `FE`, `FF FF`, `FE FE` sind ungültiges UTF-8 (Nr. 6 b) | `TestUTF16Bom`: Schleife über 8 Elemente (`FF FE x`, `FE FF x`, `FF FE 00 00`, `FF FE` als ganze Datei, `FF`, `FE`, `FF FF`, `FE FE`), die ersten vier mit NUL im Inhalt; Fall 138 | UTF-16-Datei als binär übergangen (NUL-Regel zuerst); Rest nach dem BOM weiter gemeldet | andere |
| 20 | Binär ist eine Datei mit NUL in den ersten 8000 Bytes; ab Index 8000 ist NUL das Zeichen U+0000; der Name macht nicht binär (Nr. 5 e, 6 a) | `TestBinaerGrenze`: Schleife über NUL-Stellen 0, 1, 7998, 7999 (binär) und 8000, 8001 (Text, Befund `U+0000`), Länge 0, Länge 1 (NUL), Länge 7999 mit NUL am Ende, Länge 8000 mit NUL am Ende; Fall 139 (`text.zip` mit Text und NBSP: Befund; eine Datei `*.md` mit NUL: binär) | Grenze 7999 oder 8001; NUL überall gesucht; `*.zip` nach Endung ausgenommen | andere |
| 21 | Name, Endung und Verzeichnis haben keine Wirkung (Nr. 5 e) | Fall 146: 21 Dateinamen (`a.go`, `a.md`, `a.mk`, `a.sh`, `a.yml`, `a.yaml`, `a.json`, `a.xml`, `a.sql`, `a.txt`, `go.mod`, `go.sum`, `Makefile`, `Dockerfile`, `LICENSE`, `.dockerignore`, `.gitattributes`, `.claude/settings.json`, `.githooks/commit-msg`, `a.zip`, `a.stempel`), je mit einem NBSP: 21 Befunde in 21 Dateien | Prüfung nur für `go`, `md`, `mk`, `sh`; Punkt-Verzeichnisse übersprungen; Dateien ohne Endung übersprungen | andere |
| 22 | Der Teilbaum `.harness/baseline/` bleibt ungelesen und ungezählt; die Grenze ist das Verzeichnis (Nr. 5 a) | Fall 140 (vier Pfade: einer ausgenommen, drei Befunde) und `TestAusschluesse` (dazu die Datei `.harness/baseline` und `.harness/baseline` als Teilpfad `x/…`) | Ausschluss als Präfix ohne Schrägstrich; Ausschluss auf jeder Tiefe; Ausschluss fehlt; alle Punkt-Verzeichnisse ausgenommen | andere |
| 23 | Symlinks und andere Einträge werden nicht gefolgt und nicht gelesen und als andere gezählt (Nr. 5 b, c) | `TestSymlinks`: 6 Elemente (auf Datei, auf Verzeichnis, hängend, Schleife, Ziel außerhalb der Wurzel mit NBSP, Pipe); Fall 141 (vier Symlinks, Exit 0, `4 andere`; zweiter Lauf mit NBSP in `t.md`: ein Befund, nicht `l.md`) | `os.Stat` statt `Lstat` (folgt); Symlink auf Verzeichnis betreten; Pipe geöffnet (Test hängt); Symlink nicht gezählt | andere |
| 24 | Eine leere Datei ist eine geprüfte Textdatei ohne Fundstelle (Nr. 5 d) | Fall 142 (`1 Datei(en) geprüft`, Exit 0) und `TestBinaerGrenze` Länge 0 | leere Datei als binär oder als `keine Textdatei geprüft`; nicht gezählt | andere |
| 25 | Eine Prüfung, die nichts prüft, ist rot: keine Textdatei geprüft endet mit 2 (Nr. 11 b) | Fall 143: drei Wurzeln (leer, nur eine Binärdatei, nur `.harness/baseline/` mit einem Zeichen) | Exit 0 bei *n* = 0 | andere |
| 26 | Fehlt die Wurzel oder ist sie eine Datei: Eingabefehler `Wurzel fehlt oder ist kein Verzeichnis`, Exit 2 (Nr. 11 b) | Fall 144: zwei Aufrufe | fehlende Wurzel als leere Menge behandelt (Text `keine Textdatei geprüft` statt `Wurzel fehlt`) | andere |
| 27 | Jede Aufrufform außer `zeichen pruefen <Wurzel>` ist `Aufruf: …`, Exit 2; es gibt keine Option und keine Ausnahme-Angabe (Nr. 5 f, 13) | Fall 145 und `TestAufruf`: dieselben 8 Aufrufe (keine Argumente; `pruefen`; `pruefen` und leeres Argument; drei Argumente; nur die Wurzel; `check <Wurzel>`; `--help`; `pruefen --ausnahme <Wurzel>`) | drittes Argument akzeptiert; `--help` mit Exit 0; Option `--ausnahme` liest eine Datei; leeres Argument als aktuelles Verzeichnis | andere |
| 28 | Ein Lesefehler (Datei oder Verzeichnis, auch mitten in der Datei) ist ein Eingabefehler mit Exit 2; Fundstellen vor dem Fehler bleiben; die übrigen Dateien werden gelesen (Nr. 11 b, c) | `TestLesefehler`: Fehler nach 0, 1, 7999, 8000 und 20 000 Bytes (Leser `io.MultiReader` aus Daten und `iotest.ErrReader`) und ein Fehler beim Öffnen einer Datei (Haltepunkt der Öffnung im Test; als root erzeugt chmod keinen); Befund vor dem Fehler | Fehler als Dateiende gelesen; Lauf bricht beim Fehler ab; Befunde vor dem Fehler verworfen | andere |
| 29 | Exit 0 ohne Befund, 1 mit Befund, 2 bei Eingabefehler, auch mit Befund (Nr. 11 a) | `TestExit`: Matrix Befund (ja, nein) mal Eingabefehler (keiner, Lesefehler, leere Menge, Wurzel fehlt); Fälle 131 (0), 132 (1), 143 (2) | Rang vertauscht (1 vor 2); Eingabefehler als 1 | andere |
| 30 | Ein Befund bricht den Lauf nicht ab; alle Dateien werden gelesen (Nr. 11 c) | Fall 149 (Befunde in der ersten und in späteren Dateien; die Summenzeile nennt alle) | Rückkehr beim ersten Befund; Abbruch je Datei nach dem ersten Befund | andere |
| 31 | Dateien erscheinen in der Folge ihrer Pfade, bytegleich, `/` als Byte; Befunde einer Datei nach Stelle (Nr. 12 a) | Fall 147 (Pfade `b.md`, `a/b.md`, `a-b.md`, `a.md`, `A.md`, `ä.md`, je ein Befund; Erwartung `A.md`, `a-b.md`, `a.md`, `a/b.md`, `b.md`, `ä.md`) und `TestReihenfolge` | Verzeichnisdurchlauf (`WalkDir`) statt Sortierung des ganzen Pfads (`a/b.md` käme vor `a-b.md`); Vergleich nach Locale | andere |
| 32 | Zweiter Lauf auf unverändertem Baum: gleiche Ausgabe, gleicher Exit; die Wurzel bleibt bytegleich (Nr. 12 c) | Fall 147: zweiter Lauf, Vergleich beider Läufe und der Prüfsumme der Wurzel | Ausgabe enthält Zeit oder Zähler; Programm schreibt eine Datei in die Wurzel | andere |
| 33 | Mehrere Befunde an einem Schlüssel sind je ein eigener Befund in der Folge der Stelle (Nr. 12 a, b) | siehe die Zusatztabelle in §6 (Zeilen P1 bis P5) | siehe Zusatztabelle | andere |
| 34 | Keine Grenze für Dateigröße und Zeilenlänge (Nr. 6 c) | `TestGrosseDatei`: Datei von 3 MiB und eine Zeile von 200 000 Zeichen mit NBSP am Ende (`1:200001`); Fall 148 | `bufio.Scanner` mit der Voreinstellung (64 KiB Zeile); `os.ReadFile` mit Grenze | andere |
| 35 | Quelltext und Tests des Moduls tragen kein Zeichen von Z; das Programm prüft sein eigenes Modul (Nr. 14) | `TestSelbstanwendung`: das Programm über `.` des Moduls, Tests eingeschlossen, endet mit 0 und *n* ≥ 1; zusätzlich liegt jede Datei des Moduls in der Prüfmenge von `make zeichen-check` | ein Rohzeichen (NBSP) in einer Testdatei des Moduls | andere |
| 36 | Der Bestand ist sauber: nach der Berichtigung der zwei NBSP in einer Datei meldet das Ziel auf `HEAD` keinen Befund (Nr. 2, 5) | benannte Ausgabe des Gates: `make zeichen-check` grün, die Zeile `zeichen: ok, …` steht in §7 des Plans | die zwei NBSP in `docs/reviews/2026-10-25-slice-baseline-import-verifikation-3.md` Zeile 45 bleiben: das Ziel ist rot mit `…:45:1367: U+00A0 Zs NBSP` und `…:45:1384: U+00A0 Zs NBSP` | andere |

**Mutationen und Raten** (`AGENTS.md` §3.10; gefüllt vom Implementer am 2026-10-31). Verfahren: eine Kopie von `tools/zeichen` (bzw. des Treibers `tools/harness/zeichen.sh`) im Scratchpad, genau eine Änderung am geprüften Code, `go test -count=1 ./...` im Container (`--network none --memory=4g --memory-swap=4g --cpus=4`, `golang:1.27-alpine` der Stufe `deps`); für die Fälle 131 bis 155 wird das Programm aus der Kopie gebaut und die Gruppe im Image der Stufe `gegenproben` gefahren (Treiber: die Kopie des Treibers, `bash`, `git`, `tar` des Images). Zeilen und Fänger dieser Liste sind die Nummern der Tabelle oben; „Test“ ist ein Test von `tools/zeichen`, „Fall“ eine Nummer der Gegenprobe. **Rate:** das Programm ist sequentiell (`SPEC-233` Nr. 12 d), kein Fänger hängt von einer Wahrscheinlichkeit ab; jede aufgeführte Mutation wurde in 3 von 3 Prozessen (Tests) bzw. 2 von 2 (Fälle) rot gesehen, wenn nichts anderes steht, gemessen vom Implementer auf einem Host mit 20 Kernen bei einem Load Average von 11 bis 18 (fremde Last). Zählprüfungen der Mengen stehen in den Tests (Z: 1 112 064 Skalarwerte, 63, 170, 1, 1, 16, 2 und 253; 7 Stellungen; 4 Leser; 208; 90; 8; 10; 21; 6; 8; im Treiber 10, 8 und 4 Läufe).
- **1** Verdrahtung (akzeptiertes Negativ, Beleg statt Fänger): `GATE_CHECKS += zeichen-check` entfernt: `make -n -k gates` nennt 0 statt 2 Aufrufe (`--target zeichen`, `tools/harness/zeichen.sh`; mit der Zeile Zeilen 35 und 36 der Ausgabe); `make zeichen-check` ist rot auf dem Bestand mit den zwei NBSP (Exit 2 von `make`, `1367` und `1384`). Nicht gefahren: Stufe ohne `go test`, `COPY tools/zeichen` in der Stufe `lint` entfernt (kein Fänger, wie der Plan sagt). `-n` braucht `-k`, weil das Ziel `schema-check` eine Zeile mit `make` ausführt.
- **2** Fall 151: `--others` entfernt; `--exclude-standard` entfernt; Filter der gelöschten Pfade entfernt; Liste ohne `-z` (dazu rot in 152, 153, 154); `.git` in der Liste (dazu rot in 155); nicht im Plan: hängender Symlink aus der Menge (`-e` allein) und `.harness/baseline/*` im Treiber verworfen — alle rot. `-e` durch `-f` ersetzt bleibt grün: ein Verzeichnis (Gitlink) im Arbeitsbaum gehört dann nicht zur Liste; akzeptiertes Negativ, Grund: im Repo gibt es keinen, Folge: ein Zeichen in einem Gitlink-Verzeichnis würde nicht geprüft.
- **3** Fall 155: Prüfung der leeren Liste entfernt; `kein Git-Arbeitsbaum` mit Exit 0; `git nicht aufrufbar` entfernt; Prüfung des Arguments entfernt; `Prüfmenge leer` mit Exit 0 — alle rot in `treiber-ohne-repo-und-ohne-dateien`.
- **4** Fall 152: `--network none` entfernt; `create -v …`; Ziel `/w` statt `/work`; `ZEICHEN_IMAGE` ignoriert; `start` ohne `-a`; `rm` entfernt (dazu rot in 154); `cp` nach `/tmp`; nicht im Plan: `DOCKER` ignoriert (rot erst mit dem dritten Lauf, der `DOCKER` setzt und `PATH` ohne `docker` lässt).
- **5** Fall 154 (Läufe 3 bis 5): `pipefail` entfernt; Exit der Übergabe verworfen (`|| true`); Fehler von `create` mit Exit 0; Fehler von `tar` verschluckt (`2>/dev/null || true`, dazu rot in 151, 152, 153).
- **6** Fall 153: roher Exit (255 bleibt 255); Exit 1 auf 0 abgebildet (dazu 154); jeder andere Exit 1 statt 2; nicht im Plan: nur 0 durchgereicht (1 wird 2, dazu 154).
- **7** Fall 154 (dazu 152): `trap … EXIT` entfernt; `rm` nur im Erfolgsfall (nach `start`); `trap … HUP INT TERM` entfernt; `exit 0` im Trap; `trap` vor `create` (dann `rm` bei Fehler von `create`); nicht im Plan: Trap nur für zwei der drei Signale (HUP und INT, HUP und TERM, INT und TERM) — alle rot.
- **8** Test `TestMengeZ`: `Cf` ausgelassen; `Zs` samt U+0020 (dazu `TestNichtZ`); Tabulator nicht ausgenommen (dazu `TestNichtZ`, `TestSelbstanwendung`); `Cc` nur bis U+009E; U+FFFE oder U+FFFF entfernt; `Zl` ausgelassen; nicht im Plan: `Zp` ausgelassen, U+0095 aus `Cc`, U+3000 aus `Zs` — alle rot (dazu `TestKurznamen`, `TestStellungen`, `TestStellungenInDatei`, Fall 132).
- **9** `TestStellungen` (253 × 7 × 4 = 7084 Läufe) und `TestStellungenInDatei` (252 × 7 Dateien, ohne U+0000): Prüfung nur am Zeilenanfang; letztes Zeichen der Datei ungeprüft; nur das erste Zeichen je Zeile; Dekodierung je Leseaufruf ohne Übertrag (`Read` statt `ReadRune`); Spalte in Bytes — alle rot.
- **10** Fall 132: Programm kennt von `Cf` nur U+FEFF und U+200B; `Zs` nur U+00A0 (nicht im Plan); `Cc` nur unter U+0020; Prüfmenge nur `*.go`, `*.md`, `*.mk`, `*.sh` — alle rot in `menge-z` (dazu `sauber`, `dateiarten`, `binaer-grenze`).
- **11** `TestNichtZ` (208 × 7 × 4 Läufe) und Fall 131: Klasse mit U+00E9 und U+00E4 (Umlaute rot; U+00E4 liegt in Fall 131, U+00E9 nur im Test); Tabulator als Befund; U+FFFD als ungültig (dazu 137); U+FE0F und U+3164 in Z (nur Test); nicht im Plan: das Leerzeichen U+0020 in `Zs` (Fall 131 trägt dafür Leerzeichen im Text) — alle rot.
- **12** Fall 133 und `TestStellungenInDatei`: BOM am Dateianfang abgestreift (nur Dateiebene); BOM nur am Dateianfang geprüft (dazu `TestBomAnfangUndMitte`, Fall 132).
- **13** Fall 135: CR vor LF abgestreift; CR als Zeilenende gezählt (dazu `TestZeilenUndSpalten`).
- **14** Fälle 134 und 136: U+2028, U+2029 oder U+0085 als Zeilenende (je 136); Spalte in Bytes (134, 135, 136, 149).
- **15** `TestKurznamen` und Fall 132: Namen von U+2028 und U+2029 vertauscht; Kategorie `Cn` für U+FEFF; nicht im Plan: kein Name für U+2060, `NONCHARX` für U+FFFE, `BOM2` für U+FEFF, `kurzname` leer (dann rot in zwölf Fällen).
- **16** `TestAusgabeAscii` (256 Bytewerte, 254 Namen), Fälle 150 und 147: Pfad roh (dazu `TestReihenfolge`); `\\` nicht maskiert (dazu `TestLeereMengeUndWurzel`); Grund des Lesefehlers roh, Pfad der Lesefehler roh, Wurzel roh (je `TestLesefehlerBeiOeffnenUndVerzeichnis`, `TestLeereMengeUndWurzel`); nicht im Plan: Byte 0x7F als druckbar, Byte 0x20 als nicht druckbar.
- **17** Fälle 131, 139, 141, 142, 149: *n* und *b* vertauscht (rot in 131, 139, 141, 142); *d* zählt Befunde (rot in 149 und fünf weiteren); *s* zählt *b* mit und der Teilbaum `.harness/baseline` mitgezählt (nur `TestBinaerGrenze`, `TestSauber` bzw. `TestAusschluesse`: kein Fall führt `ok` mit diesem Teilbaum).
- **18** `TestUngueltigesUTF8` (90 Elemente × 4 Leser) und Fall 137: nur das erste Byte einer ungültigen Folge gemeldet; Fehler am Dateiende verschluckt (dazu `TestUTF16Bom`); Fortsetzungsbytes übergangen; nicht im Plan: C1 nicht erkannt.
- **19** `TestUTF16Bom` und Fall 138: NUL-Regel vor der BOM-Regel; Rest nach dem BOM weiter gemeldet; `FE FF` nicht erkannt; jedes `FF` am Anfang, jedes `FE` am Anfang als BOM erkannt (`FF FF`, `FE FE`, `FF`, `FE`) — alle rot.
- **20** `TestBinaerGrenze` und Fall 139: Grenze 7999 oder 8001; NUL im ganzen Kopf von 24 000 Bytes gesucht; `*.zip` nach Endung binär (dazu `TestDateiarten`, Fall 146); nicht im Plan: Datei der Länge 1 mit NUL nicht binär.
- **21** Fall 146 und `TestDateiarten`: nur `go`, `md`, `mk`, `sh`; Punkt-Verzeichnisse übersprungen; Dateien ohne Endung übersprungen.
- **22** `TestAusschluesse` und Fall 140: Ausschluss als Präfix ohne Schrägstrich; auf jeder Tiefe; fehlt; alle Punkt-Verzeichnisse ausgenommen. Die Mutation „die Datei `.harness/baseline` wird mit ausgenommen“ ist nur im Test rot (Fall 140 führt keine solche Datei).
- **23** `TestSymlinks` (6 Elemente) und Fall 141: `os.Stat` statt `Lstat` (folgt); Symlink auf Verzeichnis betreten; Symlink nicht gezählt. Die Pipe wird nur im Test geprüft: „Pipe geöffnet“ lässt `TestSymlinks` nach 10 s mit „die Prüfung hängt“ rot werden, Fall 141 bleibt grün.
- **24** Fall 142 und `TestBinaerGrenze` (Länge 0): leere Datei als binär; nicht gezählt (rot in 16 Fällen).
- **25** Fall 143, `TestExit`, `TestLeereMengeUndWurzel`: Exit 0 bei *n* = 0.
- **26** Fall 144, `TestLeereMengeUndWurzel`: fehlende Wurzel als leere Menge.
- **27** Fall 145 und `TestAufruf` (8 Formen): drittes Argument akzeptiert; `--help` mit Exit 0; leeres Argument akzeptiert; `check` als Kommando akzeptiert. Die Option `--ausnahme` ist durch die Form „drei Argumente“ mitgefangen.
- **28** `TestLesefehler` (Fehler nach 0, 1, 7999, 8000, 20 000 Bytes), `TestLesefehlerBeiOeffnenUndVerzeichnis`: Fehler als Dateiende gelesen (dazu `TestExit`); Lauf bricht beim Fehler ab; Befunde vor dem Fehler verworfen (Zähler); Fehler eines Verzeichnisses ignoriert; Fehler beim Öffnen ignoriert.
- **29** `TestExit` und Fall 143: Eingabefehler als 1 (143); Rang vertauscht (1 vor 2) nur in `TestExit` und `TestLesefehler`, weil es keinen Fall mit Befund und Eingabefehler gibt.
- **30** Fall 149: Rückkehr beim ersten Befund einer Datei (dazu 133, 135, 136, 148); Abbruch nach der ersten Datei mit Befund (rot in 15 Fällen).
- **31** Fall 147 und `TestReihenfolge`: Verzeichnisdurchlauf statt Sortierung; Vergleich nach Kleinbuchstaben (dazu `TestDateiarten`).
- **32** Fall 147: das Programm legt eine Datei in der Wurzel an (rot in 6 Fällen); Zähler in der Ausgabe (rot in 12 Tests, darunter `TestReihenfolge`).
- **33** Fall 149 und `TestPaare`: dasselbe Zeichen nicht doppelt gemeldet (c.md).
- **34** Fall 148 und `TestGrosseDatei`: Zeile über 65 536 Spalten abgeschnitten (beide); Datei über 1 MiB abgeschnitten (nur Test; die Zeile des Falls ist kleiner).
- **35** `TestSelbstanwendung`: eine Testdatei im Modul mit den Bytes C2 A0 (NBSP) — rot.
- **36** Benannte Ausgabe des Gates: **der Fänger ist gesehen** (`make zeichen-check` auf dem Bestand: `…:45:1367: U+00A0 Zs NBSP`, `…:45:1384: U+00A0 Zs NBSP`, `2 Befund(e) in 1 Datei(en)`); der Zustand `zeichen: ok, …` steht aus, weil der Bestand nicht berichtigt ist (§6).

**Mengen und Elementzahlen** (für den Absatz unter der Tabelle, `AGENTS.md` §3.17):

| Menge | Elementzahl |
|---|---|
| Z | 253 (63, 170, 1, 1, 16, 2) |
| Stellungen | 7 |
| Leser | 4 |
| Rest, passiert | 208 (95 + 2 + 94 + 17) |
| ungültiges UTF-8 | 90 (64, 2, 11, 6, 3, 2, 2) |
| UTF-16-BOM | 8 |
| NUL-Stellen und Längen | 10 |
| Dateinamen | 21 |
| Symlinks und Einträge | 6 |
| Aufrufformen | 8 |
| Exit-Werte des Treibers | 10 |
| Läufe zum Aufräumen | 8 |
| Arten Pfad im Treiber | 8 |

Jede Allaussage („jedes Element“, „an jeder Stelle“, „jede Aufrufform“) ist eine Schleife über die Menge. Einen Fänger, der nur mit Wahrscheinlichkeit rot wird, gibt es nicht; die Rate ist 100 % je Prozess. Die Nebenläufigkeit ist nicht Teil der Zusagen (Nr. 12 d).

**Fixtures** entstehen aus Bytewerten (`\x`-Folgen, `[]byte{…}`, Umwandlung aus Zahlen), nie aus Rohzeichen oder `\u`-Folgen. Die Dateien von Test und Gegenprobe werden auf Rohbytes gegengelesen (`env LC_ALL=C grep -aP`, nicht die Shell-Funktion), und der Bericht nennt die Dateien.

**Nachzüge und Wortlaute** (Text des Architect; `AGENTS.md` §3.9: alle im Commit des Implementers, außer was als geliefert vermerkt ist). Der Implementer schreibt die Dateien nach diesen Wortlauten; weicht er ab, steht das in §6 mit Ausgang *offen* und Adresse.

| Nr. | Datei | Änderung |
|---|---|---|
| 1 | `harness/vertraege/SPEC-233-unsichtbare-zeichen.md` | geliefert (Planner-Commit vor der Marke) |
| 2 | `harness/vertraege/README.md` | geliefert: Zeile `SPEC-233` im Index |
| 3 | `harness/mk/zeichen.mk` | neu, Wortlaut unten <!-- d-check:ignore (Datei entsteht mit dem Code des Implementers) --> |
| 4 | `tools/harness/zeichen.sh` | neu: Verhalten nach Nr. 2 bis 4 des Vertrags; `#!/usr/bin/env bash`, `set -euo pipefail` (wegen `read -d` für den Filter der gelöschten Pfade) |
| 5 | `tools/zeichen/` | neu: `go.mod` (Modul `zeichen`, ohne Abhängigkeiten), Programm und Tests |
| 6 | `Dockerfile` | Stufe `zeichen`, `COPY tools/zeichen` in der Stufe `lint`, Einträge in der Stufe `gegenproben`, Kommentare, Wortlaut unten |
| 7 | `.dockerignore` | zwei Zeilen: `!tools/zeichen` und `!tools/harness/zeichen.sh` |
| 8 | `tools/harness/gegenproben-eingabe.sh` | in `EINGABEN` anfügen: `tools/harness/zeichen.sh` und `tools/zeichen` |
| 9 | `harness/mk/gegenproben.mk` | Kopfkommentar und `##`-Text des Ziels um `zeichen` ergänzen, Wortlaut unten |
| 10 | `harness/README.md` | Sensors-Zeile, Zeile des Ziels `gegenproben`, *Nicht behauptet*, Wortlaut unten |
| 11 | `tools/gegenproben/` | Fälle 131 bis 155 (Teil G, eigener Commit, zuletzt); Struktur `Gate` um `gateZeichen`/`gateZeichenTreiber`, ein Hilfsaufruf `l.programm(args…)` im Muster von `l.treiber`, Feld `ZeichenTreiber` in `Umgebung` trifft der Implementer; die Eingabe für `TestMengeZ` ist `unicode.Version` = 17.0.0, gemessen mit Go 1.27.1 |
| 12 | `harness/gegenproben.stempel` | nach dem letzten Commit unter den Eingabepfaden |
| 13 | `SPEC-101` | keine Änderung: `lint.sh` bleibt unberührt; `tools/zeichen` fällt unter „jedes Verzeichnis `tools/<Name>` mit eigener `go.mod`“, sobald die Stufe `lint` das Modul kopiert |
| 14 | dieser Plan | Planner und Implementer: §1, §3, §4, §6 und der Kopf folgen jeder Korrektur (`AGENTS.md` §3.9) |
| 15 | `docs/plan/planning/welle-sensoren.md` | Planner bei der Closure: in der Tabelle der Gegenstände den Träger „Prüfung auf BOM, …“ durch `make zeichen-check` (`SPEC-233`) ersetzen |

`harness/mk/zeichen.mk` (wörtlich): <!-- d-check:ignore (Datei entsteht mit dem Code des Implementers) -->

```make
# harness/mk/zeichen.mk — unsichtbare Zeichen in den Dateien des Arbeitsbaums (SPEC-233).
# `zeichen-check` hängt an GATE_CHECKS. Der Host wählt die Dateien mit `git` und übergibt sie an den Container
# (tools/harness/zeichen.sh); das Programm der Stufe zeichen des Dockerfile läuft ohne Netz und ohne Mount.
# Netz braucht nur die Stufe deps des Dockerfile.
DOCKER ?= docker
DOCKER_BUILD ?= $(DOCKER) build --progress=plain

.PHONY: zeichen-check
zeichen-check: ## Gate: kein BOM, NBSP, Zeilentrenner, Nullbreiten- oder Steuerzeichen und kein ungültiges UTF-8 in den Dateien des Arbeitsbaums außer .harness/baseline (SPEC-233) · seit welle-sensoren
	$(DOCKER_BUILD) --target zeichen -t odata-recorder:zeichen .
	bash tools/harness/zeichen.sh

GATE_CHECKS += zeichen-check
```

`Dockerfile`, neue Stufe nach der Stufe `groessenmarke`:

```dockerfile
# --- zeichen: das Programm aus tools/zeichen sucht in den Dateien eines Arbeitsbaums nach Zeichen, die dort
# nicht hingehören (SPEC-233). Es wird geprüft (gofmt, vet, Test) und gebaut; das Ziel `make zeichen-check`
# führt es aus dieser Stufe aus, ohne Netz und ohne Mount.
FROM deps AS zeichen
WORKDIR /zc
COPY tools/zeichen ./
RUN --network=none f="$(gofmt -l .)" && { [ -z "$f" ] || { echo "nicht gofmt-formatiert:"; echo "$f"; exit 1; }; } \
 && go vet ./... && go test ./... && go build -trimpath -buildvcs=false -o /usr/local/bin/zeichen .
WORKDIR /work
ENTRYPOINT ["/usr/local/bin/zeichen"]
```

In der Stufe `lint`: Zeile `COPY tools/zeichen tools/zeichen` nach `COPY tools/pruefung tools/pruefung`; der Kommentar der Stufe nennt `tools/zeichen` neben `tools/gegenproben`, `tools/groessenmarke` und `tools/pruefung`.

In der Stufe `gegenproben`: `COPY --from=zeichen /usr/local/bin/zeichen /usr/local/bin/zeichen` nach der Zeile für `groessenmarke`; die Zeile `COPY tools/harness/lint.sh … tools/harness/` um `tools/harness/zeichen.sh` ergänzen; der Kopfkommentar nennt `zeichen` samt Treiber neben `groessenmarke`.

`harness/mk/gegenproben.mk`: im Kopfkommentar „… pruefung (test, build, race, handbuch) und groessenmarke samt Treiber (SPEC-102)“ wird „… groessenmarke und zeichen samt Treibern (SPEC-102)“; in der Hilfe-Zeile des Ziels dieselbe Ergänzung („… handbuch), groessenmarke und zeichen, nur bei geänderter Eingabe …“).

`harness/README.md`:

- Neue Zeile in der Tabelle der Gates, nach `make handbuch-check`:

```
| `make zeichen-check` | Unsichtbare Zeichen (`SPEC-233`): keine Datei des Arbeitsbaums außer `.harness/baseline/` trägt BOM, NBSP, Zeilen- oder Absatztrenner, Nullbreiten-, Format- oder rohes Steuerzeichen (253 Zeichen, auch den Wagenrücklauf) oder ungültiges UTF-8; Binärdateien ausgenommen; der Host wählt die Dateien mit `git`, das Programm der Stufe `zeichen` urteilt (netzlos, ohne Mount) | [ADR-0004](../docs/plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) |
```

- Die Zeile `make gegenproben`: „… `pruefung` und `groessenmarke`: …“ wird „… `pruefung`, `groessenmarke` und `zeichen`: …“.
- *Nicht behauptet*, angehängt vor „sowie der Lasttest zu `LH-QA-01`“: „, die Verdrahtung von `make zeichen-check` (Stufe `zeichen`, Kopie in die Stufe `lint`, Eintrag in `GATE_CHECKS`, Aufruf von `docker` und `tar`, `SPEC-233` Nr. 16 h), die Umsetzung einer Folge aus Backslash, `u` und vier Hex-Ziffern in ein sichtbares Zeichen durch das Schreibwerkzeug (`AGENTS.md` §3.18, `SPEC-233` Nr. 16 a) und Zeichen außerhalb der Menge (Variationsselektoren, Füllzeichen, Homoglyphen, `SPEC-233` Nr. 16 b)“.

Der Vertrag trägt den Nachzug: Der Implementer ersetzt die Zeile **Stand** von `SPEC-233` durch „geliefert sind das Programm `zeichen`, der Treiber, die Stufe `zeichen` des `Dockerfile`, das Ziel in `GATE_CHECKS`, die Zeile des Ziels in `harness/README.md` und die Fälle 131 bis 155 von `SPEC-102`.“

**Nachzüge an `SPEC-102`** (der Absatz und die Fälle 131 bis 155 stehen; diese Edits liefert der Implementer im Commit des Codes, `AGENTS.md` §3.9):

| Stelle | alt | neu |
|---|---|---|
| Festlegung, Satz 2 | `… `make test-race` und `make handbuch-check`; sie gehören …` | `… `make test-race`, `make handbuch-check` und `make zeichen-check`; sie gehören …` |
| Aufbau, Satz „Die Fälle 38 bis 130 lesen …“ | `Die Fälle 38 bis 130 lesen stdout und stderr getrennt, die Fälle 1 bis 37 gemeinsam (Randform 13).` | `Die Fälle 38 bis 155 lesen stdout und stderr getrennt, die Fälle 1 bis 37 gemeinsam (Randform 13).` |
| Eingaben, erste Zeile | `… `tools/harness/groessenmarke.sh`, `tools/gegenproben/`, `tools/pruefung/`, `tools/groessenmarke/`, `Dockerfile` und `.dockerignore`` | `… `tools/harness/groessenmarke.sh`, `tools/harness/zeichen.sh`, `tools/gegenproben/`, `tools/pruefung/`, `tools/groessenmarke/`, `tools/zeichen/`, `Dockerfile` und `.dockerignore`` |
| Randform 2 | `… `groessenmarke`, `git`) … Für `git` gilt das nur, solange die Fälle 115 bis 120 es verlangen; …` | `… `groessenmarke`, `zeichen`, `git`, `tar`) … Für `git` gilt das nur, solange die Fälle 115 bis 120 und 151 bis 155 es verlangen; für `tar`, solange die Fälle 151 bis 155 es verlangen; …` |
| Randform 12 | `Die Fälle 103 bis 120 setzen `TMPDIR` …` | `Die Fälle 103 bis 120 und 151 bis 155 setzen `TMPDIR` …` |
| Randform 13 | `… den Fällen 38 bis 130 getrennt. Eine Erwartung „nennt“ der Fälle 38 bis 130 …` | `… den Fällen 38 bis 155 getrennt. Eine Erwartung „nennt“ der Fälle 38 bis 155 …` |
| Pflicht zum Fahren | `… am Programm `groessenmarke` oder seinem Treiber, an diesem Vertrag oder an `SPEC-128`, …` | `… am Programm `groessenmarke` oder seinem Treiber, am Programm `zeichen` oder seinem Treiber, an diesem Vertrag, an `SPEC-128` oder an `SPEC-233`, …` |
| Stand | `… und `make handbuch-check` (die Fälle 121 bis 130).` | `… `make handbuch-check` (die Fälle 121 bis 130) und `make zeichen-check` (die Fälle 131 bis 155).` |
| Grenze, neuer Absatz | — | `Die Fälle 131 bis 155 fangen weder die Stufe `zeichen`, den Eintrag in `GATE_CHECKS` und die Kopie von `tools/zeichen` in die Stufe `lint` noch den Aufruf von `docker` und `tar` durch das Gate (`SPEC-233` Nr. 16 h); die Argumente von `docker` fangen die Fälle 152 und 154. Weicht das `tar` des Hosts von dem des Images (GNU tar 1.35) ab, fängt keine Gegenprobe die Abweichung.` |
| Zeitbudget, neuer Absatz | — | `Die Fälle 131 bis 150 starten je einen Prozess ohne Compiler, die Fälle 151 bis 155 25 Läufe des Treibers (1, 2, 10, 8 und 4) mit je `git`, `tar` und dem Ersatz-Skript; Schätzung: 1 bis 4 s Rechenzeit, parallel zu den langen Fällen 1 bis 37. Dazu kommt die Stufe `zeichen` (`gofmt`, `go vet`, `go test`, Bau; geschätzt 10 bis 12 s bei kaltem Cache, parallel zu den Stufen `pruefung` und `groessenmarke`). Gesamtschätzung der Zusatzzeit: 0 bis 14 s im Lauf bei geänderter Eingabe; der Implementer misst (Schritt `RUN gegenproben` und Gesamtlauf, vorher und nachher, gleicher Host-Zustand, Zahl der Läufe und Load Average genannt) und trägt den Wert hier ein. Die letzte Messung steht bei 114,9 s gegen 120 s Budget; liegt ein Lauf darüber, misst er neu und trägt den Wert mit Grund ein (Absatz oben).` |

Der Fall des Wagenrücklaufs ist entschieden: Er ist rot, auch in CRLF-Dateien (`SPEC-233` Nr. 7 d); ein Arbeitsbaum mit `autocrlf` wird rot, Dateien müssen LF sein, im Bestand gibt es kein CR (Architect).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-groessenmarke-werkzeug` liegt in `done/` (Trigger der Welle), und der Architect hat den Vertrag, die Entscheidungen zu den Aufträgen 1 bis 8 und die Tabelle *Zusage · Fänger* geliefert, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Erster Slice der Welle; kein anderer Slice der Welle läuft gleichzeitig.

**Größenmarke** (`AGENTS.md` §3.14): Metrik: die Summe aus Einfügungen und Löschungen von `git diff --numstat <basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` zählt nie ([SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 4 a). Der Block unten trägt den Teil des Architect (`pfad`, `produktcode`, `fakes`, `schnitt-grund`) und den des Planner (`basis`, `schaetzung`, `marke`, `SPEC-128` Nr. 3 k): gesetzt vom Planner am 2026-10-31 im eigenen Commit vor dem ersten Code-Commit, `basis` ist der Plan-Commit davor, mit `make groessenmarke` geprüft (Marke gleich Formel, Schätzung gleich Tabelle, Messung 0); a und b hat der Planner in der Tabelle in §3 nachgezählt (36 Zeilen, alle `andere`). Die Marke ist nach dem ersten Code-Commit fest (`SPEC-128` Nr. 5 c). Aufwand nach Einschätzung des Planners: L.

```groessenmarke
pfad: tools/zeichen
pfad: tools/harness/zeichen.sh
pfad: tools/harness/gegenproben-eingabe.sh
pfad: tools/gegenproben
pfad: harness/mk/zeichen.mk
pfad: harness/mk/gegenproben.mk
pfad: harness/README.md
pfad: harness/gegenproben.stempel
pfad: Dockerfile
pfad: .dockerignore
produktcode: 480
fakes: 1
basis: 3cdbdca1aa699950846957e051156c3860f5986c
schaetzung: 3560
marke: 4900
schnitt-grund: Programm, Treiber, Stufe und Ziel sind ein Gate ohne trennbaren Teil, und ein Gate ohne Gegenprobe widerspricht ADR-0005; vorab benannter Handoff-Schnitt: Teil G (Gegenproben-Fälle 131 bis 155 mit tools/gegenproben, gegenproben-eingabe.sh, gegenproben.mk und Stempel), zuletzt in eigenem Commit, davor gemessen
```

**Zahlen des Architect** (Eingaben a, b, f, p; der Planner zählt a und b in der Tabelle in §3 nach): a (nebenläufig) = 0, weil das Programm sequentiell ist (`SPEC-233` Nr. 12 d) und die Signal-Läufe von Zeile 7 deterministisch sind; b (andere) = 36, jede Zeile der Tabelle eine Zusage, Randform oder Fallfamilie, nicht zusammengelegt, um unter die Schnittfrage zu kommen (`SPEC-128` Nr. 10 b); f = 1, das Ersatz-Skript für `docker` (hält Argumente fest, zählt Aufrufe, legt den Tar-Strom ab), `git` und `tar` sind echt, die Leser stammen aus `testing/iotest`; p = 480: Programm 380 (Aufruf und Ausgabe 60, Baum und Ausschlüsse 70, Lesen, Dekodieren und Position 120, Menge Z und Kurznamen 90, Maskierung 40), Treiber 55, Fragment `zeichen.mk` 14, Stufe im `Dockerfile` samt Einträgen 31.

**Schnittfrage** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1): *S* = 85·0 + 80·36 + 200·1 + 480 = 3560; 135·3560 = 480 600, geteilt durch 10 000 gibt 48,06, aufgerundet 49, also die Marke 4900; die Schnittfrage gilt (135·*S* > 300 000, *S* ≥ 2223). Geschnitten wird vor dem Code nicht, aus dem Grund im Block: Programm, Treiber, Stufe und Ziel bilden ein Gate, ein Teil ohne den anderen prüft nichts (das Programm braucht den Treiber für seine Prüfmenge, der Treiber das Programm für das Urteil), und ein Gate ohne Gegenprobe widerspricht [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) und dem Ziel der Welle. Der oben genannte Schnitt „Dokumente gegen Quelltext“ ist kein Schnitt ohne geteilten Mechanismus: er teilt Programm, Treiber und Gegenprobe und brauchte im Treiber eine Auswahl nach Pfadklasse. Verworfen ist auch ein Schnitt „Programm“ gegen „Treiber“: das Programm ohne Treiber ist kein Gate. **Vorab benannter Schnitt beim Handoff, Teil G:** die Gegenproben-Fälle 131 bis 155 und ihre Hilfen (`tools/gegenproben`, `tools/harness/gegenproben-eingabe.sh`, `harness/mk/gegenproben.mk`, `harness/gegenproben.stempel`), zuletzt in einem eigenen Commit geschrieben und davor gemessen; er teilt keinen Code mit dem Rest, denn er ruft das Programm als Prozess.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Schätzung des Architect mal 1,25 liegt über 3000 (`AGENTS.md` §3.14), oder der Slice braucht Eingaben von mehr als zwei Schichten der Gate-Kette; dann trennt der Planner vor dem Code den Teil für Dokumente vom Teil für Quelltext. **Ausgang:** der Satz trifft formal zu (3560 × 1,25 = 4450), der Slice wird nicht zurückgeführt: `AGENTS.md` §3.14 und `SPEC-128` Nr. 1 lassen bei fehlendem Teil ohne geteilten Mechanismus den Grund zu, er steht im Block oben (Schnittfrage), samt Teil G als Schnitt beim Handoff. Der Schnitt „Dokumente gegen Quelltext“ ist nicht mechanismusfrei; Einspruch des Users: §6.
- `in-progress` → `open` (blockiert): der Architect entscheidet den Ort (Auftrag 1) so, dass das Gate die Dokumente nicht erreicht und eine Host-seitige Auswahl nötig ist, die `slice-guide-groessen-gate` oder `slice-abdeckung-deklaration` mitbräuchten; dann geht die Frage an den Planner (Reihenfolge der Welle). **Ausgang:** nicht eingetreten; das Gate erreicht die Dokumente, der Host wählt die Dateien mit `git ls-files` und übergibt sie als Tar-Strom, und die Auswahl braucht nur dieses Gate (`slice-guide-groessen-gate` läuft im d-check-Modul mit eigenem Mount, `slice-abdeckung-deklaration` liest in der Stufe `source`).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, `make gates` und `make gegenproben` grün im Hauptbaum, Review- und Verifikations-Report unter `docs/reviews/`, Closure-Notiz in §7. Der Lerneintrag nennt den Sensor mit `liegt in` (Ziel oder Skript, mit dem Anker `seit welle-sensoren`) und den Teil von `AGENTS.md` §3.18, den das Gate nicht trägt (benannte Lücke), samt Diff-Vorschlag an den User.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Der nicht gefangene Teil von `AGENTS.md` §3.18 (die Umsetzung zu einem sichtbaren Zeichen) lässt die Prosa-Regel stehen; wer das Gate für die ganze Regel hält, streicht die Regel zu früh — **Ausgang:** offen bis zur Closure; erwartet weiter offen: → `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (Teil „Umsetzung zu sichtbaren Zeichen“); der Vertrag nennt ihn unter *Ungeprüft* (`SPEC-233` Nr. 16 a, entschieden).
- Prüfmenge gegen Build-Kontext: Die Stufen des `Dockerfile` sehen nur, was `.dockerignore` zulässt; Dokumente und `tools/` erreichen die Stufe `lint` nicht — **Ausgang:** offen bis zur Closure; entschieden im Vertrag (Auftrag 1 geschlossen): der Host liefert die Dateien per Tar-Strom, die Stufe `lint` ist an der Prüfmenge nicht beteiligt (`SPEC-233` Nr. 1 und 2).
- Fehlalarm: ein NBSP oder eine Steuerfolge, die ein Dokument oder eine Fixture absichtlich trägt, färbt das Gate rot, und die naheliegende Antwort ist eine Ausnahme — **Ausgang:** offen bis zur Closure; entschieden im Vertrag (Auftrag 3): es gibt keine Ausnahme (`SPEC-233` Nr. 5 f), der Bestand ist 1 Datei und wird im Slice berichtigt (§3).
- Selbstanwendung: Quelltext und Tests des Prüfprogramms dürfen die verbotenen Zeichen nur als Bytewerte tragen, sonst färbt das Gate sich selbst rot oder, schlimmer, ein Schreibwerkzeug setzt das Fixture still um — **Ausgang:** offen bis zur Closure; entschieden: `SPEC-233` Nr. 14 und Zeile 35 der Tabelle; der Fänger prüft die Datei des Tests mit (Gegenlesen auf Rohbytes, Betriebsregeln in §3).
- Laufzeit der Gegenproben: jeder Fall verlängert den Lauf, dessen Re-Evaluierungs-Trigger ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): Lauf über 30 s bei geänderter Eingabe; gemessen 56,9 s Schritt) beim User unentschieden ist — **Ausgang:** offen bis zur Closure; der Slice misst die Zusatzzeit und berichtet sie in §7, er entscheidet den Trigger nicht; die Schätzung des Architect (0 bis 14 s im Lauf bei geänderter Eingabe, Absatz *Zeitbudget* in §3) misst der Implementer; weiter offen: → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (Adresse: User).
- Die Aufhängung des Ziels in `GATE_CHECKS` ist durch keinen Fall der Gegenproben gezeigt (`harness/README.md` *Nicht behauptet* nennt dies für die bestehenden Ziele) — **Ausgang:** weiter offen als akzeptiertes Negativ: Zeile 1 der Tabelle und `SPEC-233` Nr. 16 h nennen Grund und Folge; der Beleg beim Handoff ist `make -n gates` samt rotem `make zeichen-check` auf einem Baum mit eingepflanztem NBSP.
- Schnitt statt Rückführung (Übergabe des Architect, offene Entscheidung 1): Der Satz der Rückführung in §4 trifft formal zu (3560 × 1,25 = 4450); der Planner führt nicht zurück, sondern setzt den Grund und den Schnitt Teil G in §4 (`AGENTS.md` §3.14, `SPEC-128` Nr. 1) — **Ausgang:** weiter offen; Adresse: User, nur bei Einspruch (dann `in-progress` → `next` vor dem ersten Code-Commit); ohne Einspruch bis zur Closure gilt der Grund.
- Kennung `SPEC-233` (offene Entscheidung 2): `harness/vertraege/README.md` §Kennungen verlangt die höchste vergebene Nummer plus eins; 232 war über alle Zweige das Maximum — **Ausgang:** weiter offen; Adresse: User, nur falls die Reihe anders laufen soll; vor dem Merge noch einmal über alle Zweige `git grep SPEC-233` (parallele Worktrees).
- Keine ADR zum Gate (offene Entscheidung 3): kein Accepted-ADR wird berührt, die Bindung in `harness/README.md` ist [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) — **Ausgang:** weiter offen; Adresse: User, nur falls eine ADR gewünscht ist (Titel „Unsichtbare Zeichen sind ein Gate, der Host wählt die Dateien, ein Programm im Container urteilt“, Status Proposed, `Schärft: SPEC-233`, Alternativen: (a) `lint.sh` — Dokumente unerreichbar, `grep -P` locale-abhängig; (b) Mount — daemon-abhängig, nicht das Muster des Repos; (c) gewählt: `git ls-files` plus Tar-Strom); die Entscheidung bleibt dieselbe.
- Wagenrücklauf rot (offene Entscheidung 4, entschieden vom Architect, `SPEC-233` Nr. 7 d): ein Arbeitsbaum mit `autocrlf` wird rot, Dateien müssen LF sein; im Bestand gibt es kein CR — **Ausgang:** weiter offen; Adresse: User, nur bei Einspruch gegen die Entscheidung.
- Zwei Berichtigungen des Planners am Text des Architect, nachgerechnet gegen den Vertrag (`AGENTS.md` §3.9): Zeile 15 der Tabelle sagte „29 Namen, eindeutig“, die Kurznamen von U+FFFE und U+FFFF sind beide NONCHAR (`SPEC-233` Nr. 7 f), also 28 verschiedene; der Absatz *Zeitbudget* sagte „rund 35 Läufe“ des Treibers, die Fälle 151 bis 155 führen 25 (1, 2, 10, 8, 4) — **Ausgang:** weiter offen; Adresse: Architect, nur bei Einspruch; der Vertrag `SPEC-233` bleibt unberührt.
- Das Gegenlesen mit der Shell-Funktion `grep` des Agenten-Shells kann falsch-negativ sein (Befund des Architect: `ugrep`, `-P` mit Rohbytes 0 Treffer; `env LC_ALL=C grep` 1 Treffer an derselben Datei) — **Ausgang:** weiter offen; Adresse: Planner (Closure: als weiteres Vorkommen im Eintrag `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` oder als eigener Eintrag führen); der Diff-Vorschlag für `AGENTS.md` §3.18 nennt `env LC_ALL=C grep -aP`; für BOM und NBSP macht das Gate das Gegenlesen überflüssig.
- Eine Berichtigung im Bestand (§3, Zeile „Bestand der Prüfmenge“) kann beim Schreiben still umfallen, weil das Werkzeug ein Zeichen umsetzt — **Ausgang:** *eingetreten*, offen; Adresse: Planner und User. Der Implementer kann die Zeichen nicht ausgeben: ein `Edit` auf `docs/reviews/2026-10-25-slice-baseline-import-verifikation-3.md` Zeile 45 mit zwei U+00A0 im `old_string` lehnt das Werkzeug ab (`old_string` und `new_string` kommen als dieselbe Zeichenfolge an, „No changes to make“), und eine Probe mit `Write` zeigt, dass ein ausgegebenes NBSP als Leerzeichen (Byte 0x20) in der Datei steht (Rohbytes gegengelesen: `xxd`, `a b` ist `61 20 62`). Die Datei ist unverändert (`sha256sum` vor und nach den Versuchen gleich, `git diff` leer). Die Berichtigung (je Zeichen ein Leerzeichen, Bytes C2 A0 zu 20, laut Auftrag des Laufs) macht der User von Hand oder der Planner mit einem Werkzeug, das Bytes schreibt; eine Ausnahme per Pfad gibt es nicht (`SPEC-233` Nr. 5 f, `AGENTS.md` §3.2), ein Skript im Container nicht (Auftrag). Bis dahin ist `make zeichen-check` und mit ihm `make gates` rot, mit genau zwei Befunden: `docs/reviews/2026-10-25-slice-baseline-import-verifikation-3.md:45:1367: U+00A0 Zs NBSP` und `…:45:1384: U+00A0 Zs NBSP`, `2 Befund(e) in 1 Datei(en)` (Zeile 36 der Tabelle: der Fänger ist damit gesehen, der Zustand „sauber“ steht aus).
- Die Reihenfolge der Optionen von `tar` in `SPEC-233` Nr. 3 b (`tar -c --null -T - --no-recursion -f -`) lässt GNU tar 1.35 mit Exit 2 enden („`--no-recursion` hat keinen Einfluss“: die Option steht hinter `-T -`, einem Dateioperanden, und wirkt nur auf Folgendes; gemessen am Host, `make zeichen-check`, Meldung `zeichen.sh: Übergabe scheitert`). Der Treiber ruft `tar -c --null --no-recursion -T - -f -`: dieselben Optionen, die Option vor der Liste — **Ausgang:** *eingetreten*, offen; Adresse: Architect (Nr. 3 b von `SPEC-233` berichtigen; die Fälle 151 bis 155 fahren das `tar` des Images und fangen die Reihenfolge, `SPEC-102` Grenze).
- Die Ausnahme in `.golangci.yml` für `os.Stdout` und `os.Stderr` in den `main.go` der Programme unter `tools/` (Regel mit `# Why:`) nannte `tools/zeichen/main.go` nicht, und `forbidigo` meldet beide Ströme im Einstieg des neuen Programms. Der Pfad der Regel nennt jetzt `zeichen` neben `gegenproben`, `groessenmarke` und `pruefung`; die Schwelle bleibt, kein Linter wird ausgeschaltet (`AGENTS.md` §3.6, eine Erweiterung der Pfade derselben Regel für denselben Grund, kein neuer Eintrag) — **Ausgang:** *eingetreten*, offen; Adresse: Architect (Entscheidung nachträglich bestätigen; `SPEC-101` nennt die Regel nicht).
- Die Programmliste der Gegenproben (`SPEC-102` Randform 2) nennt nach dem Text des Architect `zeichen`, `git` und `tar`, nicht `bash`; der Treiber `zeichen.sh` startet mit `#!/usr/bin/env bash`, und die Fälle 151 bis 155 rufen `bash` — **Ausgang:** *weiter offen*; Adresse: Architect, nur bei Einspruch; fehlte `bash` im Image, wären die Fälle rot (Exit 127) statt `gegenproben: bash nicht aufrufbar`, das Image der Stufe trägt es (gemessen: `/usr/bin/bash`).

**Raster der Randformen** (`AGENTS.md` §3.16; füllt der Architect vor dem ersten Code-Commit: jede Zelle nennt ihren Ort im Vertrag oder „gilt nicht“ mit Grund; eine leere Zelle ist ein Fehler des Plans; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12). Gefüllt vom Architect; „Nr.“ nennt eine Nummer von `SPEC-233`:

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Prüfen einer Datei (Inhalt) | gilt nicht als Datei: das Programm liest den Baum, keine Liste; verschwindet eine Datei zwischen Aufzählung und Öffnen, ist es `nicht lesbar` (Nr. 11 b) | leere Datei: geprüft und gezählt, keine Fundstelle (Nr. 5 d) | Zeichen von Z an Dateianfang, Zeilenanfang, Mitte, Zeilenende, Dateiende (Nr. 7, 9); kein gültiges UTF-8 (Nr. 8); `FF FE`/`FE FF` am Anfang (Nr. 6 b); NUL vor und nach Index 8000 (Nr. 6 a); Zeichen außerhalb von Z (Nr. 7 e, 16 b) | Lesefehler mitten in der Datei: Eingabefehler, Fundstellen davor bleiben (Nr. 11 b) | zwei Zeichen derselben Zeile, dasselbe Zeichen mehrfach, dasselbe Zeichen in zwei Dateien (Nr. 12 b; Zusatztabelle) |
| Bestimmen der Prüfmenge im Treiber (Liste von `git`) | im Arbeitsbaum gelöschter Pfad (Nr. 2 b); kein Git-Arbeitsbaum, kein `git` (Nr. 2 d) | leere Liste (Nr. 2 d) | ignorierte, ungetrackte, getrackte Dateien, Symlinks (auch hängend), Name mit Leerzeichen oder Zeilenumbruch, `.git` (Nr. 2 a, c, e) | `tar` oder `cp` scheitert mitten im Strom (Nr. 3 c) | gilt nicht: derselbe Pfad zweimal in der Liste (ein Merge-Konflikt nennt ihn je Stufe) legt dieselbe Datei zweimal in den Strom; der Container entpackt zu einer Datei (Nr. 1 c) |
| Bestimmen der Prüfmenge im Programm (Baum) | Wurzel fehlt oder ist eine Datei (Nr. 11 b) | leere Wurzel: Eingabefehler `keine Textdatei geprüft` (Nr. 11 b) | Binärdatei (Nr. 6 a), Symlink samt Ziel außerhalb der Wurzel, Pipe (Nr. 5 b, c), `.harness/baseline` und Beinahe-Treffer (Nr. 5 a) | Verzeichnis nicht lesbar: Eingabefehler, Lauf geht weiter (Nr. 11 b, c) | gilt nicht: der Baum kennt jede Datei einmal; ein Symlink auf sie ist ein anderer Eintrag (Nr. 5 b) |
| Ausnahme | gilt nicht: es gibt keine Ausnahme; ein Eintrag ohne Grund kann nicht entstehen (Nr. 5 f) | gilt nicht: wie links (Nr. 5 f) | gilt nicht: wie links; die Option `--ausnahme` ist ein Aufruffehler (Nr. 13) | gilt nicht: keine Konfiguration, die abbräche (Nr. 5 f) | gilt nicht: kein Eintrag, der doppelt stehen könnte (Nr. 5 f) |
| Aufruf des Programms (Argumente, Umgebung) | kein Argument (Nr. 13) | leeres Argument (Nr. 13) | unbekanntes Kommando, unbekannte Option, drittes Argument (Nr. 13) | gilt nicht: das Programm hält keinen Zustand und schreibt nichts in die Wurzel; ein Abbruch lässt nichts zurück (Nr. 12 c) | zweiter Lauf auf unverändertem Baum: gleiche Ausgabe, gleicher Exit (Nr. 12 c); die Umgebung hat keine Wirkung, das Programm liest keine (Nr. 12 c) |
| Aufruf des Treibers (Argumente, Umgebung) | kein Git-Arbeitsbaum, kein `git`, keine Dateien (Nr. 2 d) | Argument leer: Aufruffehler wie jedes Argument (Nr. 2 d) | ein Argument (Nr. 2 d); `create` scheitert wegen fehlendem Image (Nr. 3 c) | HUP, INT, TERM: Container entfernt, Exit 2 (Nr. 4 b); KILL (Nr. 16 g) | gilt nicht: der Treiber hält außer dem Container keinen Zustand, und den entfernt er (Nr. 4 b) |
| Meldung und Ausgang | keine Befunde: `ok`-Zeile (Nr. 10 c) | gilt nicht: jede Befund-Zeile trägt `U+HEX`, ein Byte oder `UTF-16-BOM` (Nr. 10 a) | Befunde in mehreren Dateien und Reihenfolge (Nr. 12 a); Befund und Eingabefehler (Nr. 11 a) | Schreibfehler auf einem Strom: nicht ausgewertet, akzeptiertes Negativ (Nr. 16 i) | gleiche Befunde beim zweiten Lauf (Nr. 12 c) |

**Zusatztabelle Paare und Modi** (Spalte *Fänger*; füllt der Architect, wenn die Operation mehrere Verstöße am selben Schlüssel meldet — hier: zwei verschiedene Zeichen derselben Zeile, dasselbe Zeichen in zwei Dateien, Befund und Lesefehler derselben Datei):

| Paar oder Modus | Entscheidung (Ort im Vertrag) | Fänger |
|---|---|---|
| P1 zwei verschiedene Zeichen derselben Zeile (benachbart: U+200B, U+00A0) | zwei Befunde in der Folge der Spalten (Nr. 12 a, b) | Fall 149 (Datei `a.md`, Zeile 1) und `TestPaare` |
| P2 dasselbe Zeichen mehrfach (dreimal benachbart; dreimal verteilt) | je Vorkommen ein Befund, keine Zusammenfassung (Nr. 12 b) | Fall 149 (`c.md`: `1:2`, `1:3`, `1:4`) und `TestPaare` |
| P3 dasselbe Zeichen in zwei Dateien | je Datei in der Folge der Pfade (Nr. 12 a) | Fall 149 (`a.md`, `b.md`, `c.md`) |
| P4 Zeichen von Z und ungültiges UTF-8 in derselben Datei | beide, in der Folge der Stelle (Nr. 8, 12 a) | Fall 149 (`a.md`: `2:2` ungültiges Byte, `2:4` NBSP) und `TestPaare` |
| P5 NBSP unmittelbar vor CR vor LF | zwei Befunde, Spalte *n* und *n*+1 (Nr. 7 d, 9) | Fall 135 |
| P6 Befund und Lesefehler derselben Datei | die Befunde davor bleiben, die Fehlermeldung folgt, Exit 2 (Nr. 11 a, b; 12 a) | `TestLesefehler`, `TestExit` |
| P7 Befund in `.harness/baseline/` und Befund daneben | nur der außerhalb (Nr. 5 a) | Fall 140 |
| P8 Zeichen in einer Datei und in einem Symlink auf sie | einmal, bei der Datei (Nr. 5 b) | Fall 141 (zweiter Lauf) |
| P9 Zeichen in einer Binärdatei neben einem in einer Textdatei | nur die Textdatei; die Binärdatei wird gezählt (Nr. 6 a) | Fall 139 |
| P10 UTF-16-BOM mit späterem NBSP in derselben Datei | nur der Befund `UTF-16-BOM`; der Rest wird nicht gelesen (Nr. 6 b) | Fall 138 (`nbsp.md`) und `TestUTF16Bom` |
| Modus Befund gegen Eingabefehler | 2 geht vor 1; die Befunde stehen trotzdem in der Ausgabe (Nr. 11 a) | `TestExit`, Fälle 132 (1), 143 (2) |
| Modus Exit des Programms im Treiber | 0 und 1 gehen durch, jeder andere Exit wird 2 (Nr. 4 a) | Fall 153 (10 Werte) |
| Modus Entstehung des Containers gegen Fehler (`create`, `cp`, `tar`, `start`, Signal) | Container entfernt, außer wenn `create` scheiterte (Nr. 4 b) | Fall 154 (8 Läufe) |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.* Die Closure trägt zusätzlich: die gemessene Zusatzzeit der Gegenproben (Schritt und Gesamtlauf, vorher und nachher, Zahl der Läufe, Zustand des Hosts), die Zahl der im Bestand berichtigten Dateien und den Diff-Vorschlag für `AGENTS.md` §3.18 an den User.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, Stand 2026-10-31, 35 Einträge; Zahlen aus den Dateien unter `evidence/`). Für diesen Slice zählen:
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (4×, verkörpert in `AGENTS.md` §3.18): der Gegenstand; das Gate trägt den unsichtbaren Teil, der Rest bleibt Prosa (§6).
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (29×), `akzeptiertes-negativ-ohne-folge` (11×), `randform-erst-im-review-entschieden` (22×), `universalaussage-im-vertrag-ungeprueft` (12×): Tabelle, Raster und Allaussagen („jedes Zeichen der Liste“) mit Fänger über alle Elemente.
- `gegenprobe-haengt-an-der-laufzeit-des-produkts` (2×, offen), `stempel-gehoert-zu-keinem-committeten-stand` (1×, offen), `betriebsregel-fuer-repo-dateien-verletzt` (2×, offen), `hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, offen): Gegenproben ohne Produkt in der Eingabe, Stempel nach dem letzten Commit, Betriebsregeln in §3, Hüllen von Argumenten (falls das Prüfprogramm einen Strom oder Dateisystem-Fake nutzt) tragen jede Schnittstelle des Originals.
- `rate-ohne-detektor-schwankt-mit-der-last` (3×, offen): trifft zu, falls ein Fänger nur mit Wahrscheinlichkeit rot wird; ein weiterer Treffer hier wäre der vierte.
- `berichtigung-erreicht-nicht-jede-stelle` (14×, §3.9): die Zeile in `harness/README.md`, die Zeile im Index der Verträge und die Gegenproben-Eingaben ziehen im selben Commit nach.
- `fix-deckt-die-genannte-mutation-nicht-die-klasse` (9×, §3.17): die Liste der Zeichen ist eine Menge; der Fix deckt sie ganz.
Unter der Schwelle ohne Bezug: die übrigen Einträge.

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF (Default `*`, Greenfield: Neues Repo ohne Bestandscode; die Spec führt, der Code folgt).
