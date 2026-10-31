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

**Berührte Spec-Stellen:** ein neuer Vertrag unter `harness/vertraege/` (Kennung nach `harness/vertraege/README.md` §Kennungen, vom Architect geschrieben, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)) · `SPEC-101` (Lint-Profil und `lint.sh`, falls der Architect den Ort dort wählt) · `SPEC-102` (Fälle der Gegenprobe, Zeitbudget) · `harness/README.md` §Sensors (Zeile des Ziels, falls ein neues Ziel entsteht). `spec/spezifikation.md` und die Architektur-Sicht bleiben unberührt: der Slice liefert kein Anwendungscode.
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

**Ziel:** Ein Gate in der Kette `make gates` wird rot, wenn eine Datei der benannten Prüfmenge ein Zeichen enthält, das dort nicht hingehört (BOM, NBSP, U+2028, U+2029, U+200B und rohe Steuerzeichen außer Zeilenende und Tabulator), und weist mit einer Gegenprobe nach, dass es rot wird; `AGENTS.md` §3.18 verliert damit den Teil, den ein Gate fangen kann.

**Der Teil, den das Gate nicht fängt, ist benannt:** Die Regel §3.18 beschreibt zwei Wirkungen des Werkzeugs: eine Folge aus Backslash, `u` und vier Hex-Ziffern wird still in das **sichtbare** Zeichen umgesetzt (aus der Folge für die öffnende eckige Klammer wird das Zeichen `[` selbst), und Rohbytes wie BOM oder NBSP gehen im Bericht verloren oder entstehen unbemerkt. Das Gate sieht nur, was in der Datei steht: ein unsichtbares Zeichen, das dort nicht hingehört. Die Umsetzung zu einem sichtbaren Zeichen, das der Autor als Text meinte, hinterlässt eine gültige Datei; kein Zeichenfilter erkennt sie. Dieser Teil bleibt Prosa (`AGENTS.md` §3.18) und Gegenlesen; die Closure nennt ihn in der Tabelle der Welle als „Prosa bleibt“.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung von `AGENTS.md` §3.18 — `AGENTS.md` ändert nur der User; der Slice liefert in der Closure einen Diff-Vorschlag (Zeiger auf das Gate, Kern der Regel bleibt für den nicht gefangenen Teil).
- Die Erkennung der Umsetzung `\u…` → sichtbares Zeichen (siehe oben) — nicht entscheidbar aus dem Ergebnis; ein Fänger dafür wäre ein Test, der den Wert mit dem Wert vergleicht, den der Autor meinte, und den kennt nur der Test selbst.
- Zeichen, die in Dokumenten, Kommentaren und Zeichenketten legitim stehen (deutsche Umlaute, typografische Striche, ein `§` und Ähnliches, UTF-8 überhaupt) — die Prüfung gilt einer festen Liste unsichtbarer Zeichen, keiner Zeichenklasse „kein ASCII“; die Liste und ihre Grenze entscheidet der Architect im Vertrag.
- Die vendorte Baseline unter `.harness/baseline/` und fremde Bäume — der Baum ist Kurs-Inhalt und wird byte-genau mitgeliefert (`harness/conventions.md` §Baseline); ob und wie er in der Prüfmenge fehlt, steht im Vertrag.
- Eine Direktive im geprüften Text, die das Gate umgeht (`AGENTS.md` §3.2) — Ausnahmen für Testfixtures mit Absicht entstehen aus Bytewerten (`esc("…")`, `[]byte{…}`) und brauchen keine Ausnahme; eine Pfad-Ausnahme für Daten-Fixtures (Dateien, die ein BOM tragen müssen) wäre ein Eintrag im Vertrag oder in der Konfiguration des Gates, kein Marker im Text; der Architect entscheidet, ob es sie braucht (Auftrag 3).

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

- [ ] Das Gate (Ort und Name nach der Entscheidung des Architect, Auftrag 1) hängt an `GATE_CHECKS` und wird rot, wenn eine Datei der Prüfmenge ein Zeichen der Liste enthält; die Prüfmenge und die Liste stehen in einem Vertrag unter `harness/vertraege/`, jede Zusage des Vertrags hat eine Zeile in der Tabelle *Zusage · Fänger* (§3) mit einer vom Implementer selbst rot gesehenen Mutation.
- [ ] Das Gate ist durch eine Gegenprobe nach [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) belegt (ein Fall je Zeichen der Liste, Fixtures aus Bytewerten, auf Rohbytes gegengelesen); die Zusatzzeit der Gegenproben ist gemessen und in §7 berichtet.
- [ ] Der Bestand ist sauber oder berichtigt: Dateien der Prüfmenge, die ein Zeichen der Liste heute tragen, sind im Slice berichtigt (keine Ausnahme-Liste ohne Carveout, `AGENTS.md` §3.6); die Zahl der berichtigten Dateien steht in §7.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Zeile des Ziels in `harness/README.md` §Sensors (im selben Commit wie das Ziel) und Zeile im Index `harness/vertraege/README.md`; die Closure legt den Diff-Vorschlag für `AGENTS.md` §3.18 bei (der User ändert, kein Rollenlauf).
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
| ein Vertrag unter `harness/vertraege/` und die Zeile im Index dort | neu (Architect) | Prüfmenge, Liste der Zeichen, Ausnahmen, Ausgabe, Exit-Codes, Randformen, *Ungeprüft* (der nicht gefangene Teil von `AGENTS.md` §3.18) |
| Prüfprogramm (`tools/harness/lint.sh` oder ein eigenes Programm unter `tools/`; Ort: Auftrag 1) | update oder neu | liest Dateien, meldet Datei, Zeile, Spalte und das Zeichen als Kennung `U+XXXX` (nie als Rohzeichen) |
| `Makefile`-Fragment unter `harness/mk/`, `Dockerfile`, `.dockerignore` | update oder neu | Ziel in `GATE_CHECKS`, nur falls der Architect ein eigenes Ziel wählt; Stufe und Allowlist wie beim Muster `groessenmarke` |
| `harness/README.md` | update | Zeile des Ziels in §Sensors und die Zeile in *Nicht behauptet*, die den nicht gefangenen Teil nennt, im selben Commit wie das Ziel (sonst meldet `make docs-check` ein Target ohne Zeile) |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` | update | Fälle zu `SPEC-102` (je Zeichen der Liste ein Fall, Wegwerfverzeichnis, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); `EINGABEN` des Skripts um die neuen Pfade; Stempel `harness/gegenproben.stempel` nach dem letzten Commit unter den Eingabepfaden |
| Testdateien des Prüfprogramms | neu | eine Zeile der Tabelle unten je Test; Mengen als Schleifen über die Liste (`AGENTS.md` §3.17) |
| Bestand der Prüfmenge | update, falls Treffer | gemessen am Stand `0c524e2` mit `grep -rlP` auf Rohbytes (ohne `.git`): **0 Dateien** mit BOM, NBSP, U+2028, U+2029, U+200B oder rohem Steuerzeichen außer Zeilenende und Tabulator; die Messung wiederholt der Implementer am Start |

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Vertrag, Aufträge 1 bis 8, Tabelle *Zusage · Fänger*, Raster, Gegenproben-Fälle, Nachzüge, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorschlag an den User).

**Betriebsregeln für alle Rollen** (Wortlaut und Gründe: [welle-sensoren](../welle-sensoren.md) §4):

- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen; Docker sonst nur über `make` (`AGENTS.md` §3.1). Gate-Läufe laufen im Hauptbaum, weil die Images feste Tags tragen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- **Dieser Slice ist der Fall von `AGENTS.md` §3.18 selbst:** Fixtures mit BOM, NBSP, U+2028, U+2029, U+200B, rohem Steuerzeichen oder einer Folge aus Backslash, `u` und vier Hex-Ziffern entstehen aus Bytewerten (`[]byte{…}`, `esc("…")`); die geschriebene Datei, der Plan und der Bericht werden mit `LC_ALL=C grep -P` auf Rohbytes gegengelesen, und der Bericht nennt, welche Dateien gegengelesen wurden. Ein Fänger, dessen Fixture beim Schreiben still zu einem anderen Zeichen wird, ist kein Fänger.
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

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 6) | — | — |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-groessenmarke-werkzeug` liegt in `done/` (Trigger der Welle), und der Architect hat den Vertrag, die Entscheidungen zu den Aufträgen 1 bis 8 und die Tabelle *Zusage · Fänger* geliefert, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Erster Slice der Welle; kein anderer Slice der Welle läuft gleichzeitig.

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner schreibt nach dem Architect den Block `groessenmarke` ([SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 3) in diese Sektion und setzt die Marke mit `make groessenmarke` im eigenen Commit vor dem ersten Code-Commit; die Metrik nennt der Block. Aufwand nach Einschätzung des Planners: S.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Schätzung des Architect mal 1,25 liegt über 3000 (`AGENTS.md` §3.14), oder der Slice braucht Eingaben von mehr als zwei Schichten der Gate-Kette; dann trennt der Planner vor dem Code den Teil für Dokumente vom Teil für Quelltext.
- `in-progress` → `open` (blockiert): der Architect entscheidet den Ort (Auftrag 1) so, dass das Gate die Dokumente nicht erreicht und eine Host-seitige Auswahl nötig ist, die `slice-guide-groessen-gate` oder `slice-abdeckung-deklaration` mitbräuchten; dann geht die Frage an den Planner (Reihenfolge der Welle).

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

- Der nicht gefangene Teil von `AGENTS.md` §3.18 (die Umsetzung zu einem sichtbaren Zeichen) lässt die Prosa-Regel stehen; wer das Gate für die ganze Regel hält, streicht die Regel zu früh — **Ausgang:** offen bis zur Closure; erwartet weiter offen: → `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (Teil „Umsetzung zu sichtbaren Zeichen“); der Vertrag nennt ihn unter *Ungeprüft*.
- Prüfmenge gegen Build-Kontext: Die Stufen des `Dockerfile` sehen nur, was `.dockerignore` zulässt; Dokumente und `tools/` erreichen die Stufe `lint` nicht — **Ausgang:** offen bis zur Closure; entschieden im Vertrag (Auftrag 1).
- Fehlalarm: ein NBSP oder eine Steuerfolge, die ein Dokument oder eine Fixture absichtlich trägt, färbt das Gate rot, und die naheliegende Antwort ist eine Ausnahme — **Ausgang:** offen bis zur Closure; entschieden im Vertrag (Auftrag 3); eine Ausnahme ist ein Eintrag mit Grund, kein Marker im Text.
- Selbstanwendung: Quelltext und Tests des Prüfprogramms dürfen die verbotenen Zeichen nur als Bytewerte tragen, sonst färbt das Gate sich selbst rot oder, schlimmer, ein Schreibwerkzeug setzt das Fixture still um — **Ausgang:** offen bis zur Closure; der Fänger prüft die Datei des Tests mit (Gegenlesen auf Rohbytes, Betriebsregeln in §3).
- Laufzeit der Gegenproben: jeder Fall verlängert den Lauf, dessen Re-Evaluierungs-Trigger ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): Lauf über 30 s bei geänderter Eingabe; gemessen 56,9 s Schritt) beim User unentschieden ist — **Ausgang:** offen bis zur Closure; der Slice misst die Zusatzzeit und berichtet sie in §7, er entscheidet den Trigger nicht; weiter offen: → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (Adresse: User).
- Die Aufhängung des Ziels in `GATE_CHECKS` ist durch keinen Fall der Gegenproben gezeigt (`harness/README.md` *Nicht behauptet* nennt dies für die bestehenden Ziele) — **Ausgang:** offen bis zur Closure; der Architect entscheidet, ob ein Fall sie zeigt, sonst steht sie als akzeptiertes Negativ mit Grund und Folge im Vertrag.

**Raster der Randformen** (`AGENTS.md` §3.16; füllt der Architect vor dem ersten Code-Commit: jede Zelle nennt ihren Ort im Vertrag oder „gilt nicht“ mit Grund; eine leere Zelle ist ein Fehler des Plans; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Prüfen einer Datei (Inhalt) | (Vertrag, offen) Datei fehlt | (Vertrag, offen) leere Datei | (Vertrag, offen) Zeichen der Liste an Anfang, Mitte, Ende; kein gültiges UTF-8; Zeichen außerhalb der Liste | (Vertrag, offen) Lesefehler mitten in der Datei | (Vertrag, offen) zwei Zeichen derselben Zeile, dasselbe Zeichen mehrfach |
| Bestimmen der Prüfmenge | (Vertrag, offen) Pfad fehlt | (Vertrag, offen) leere Menge | (Vertrag, offen) Binärdatei, Symlink, Pfad außerhalb des Baums | (Vertrag, offen) Abbruch beim Lesen der Liste | (Vertrag, offen) dieselbe Datei zweimal in der Liste |
| Ausnahme | (Vertrag, offen) Eintrag ohne Grund | (Vertrag, offen) leerer Eintrag | (Vertrag, offen) Eintrag, der keine Datei trifft | (Vertrag, offen) gilt nicht, falls Konfiguration | (Vertrag, offen) Eintrag doppelt |
| Aufruf des Programms (Argumente, Umgebung) | (Vertrag, offen) kein Argument | (Vertrag, offen) leeres Argument | (Vertrag, offen) unbekannte Option | (Vertrag, offen) Abbruch durch Signal | (Vertrag, offen) zweiter Lauf auf unverändertem Baum |
| Meldung und Ausgang | (Vertrag, offen) keine Befunde | (Vertrag, offen) Befund ohne Zeichen-Kennung | (Vertrag, offen) Befunde in mehreren Dateien, Reihenfolge | (Vertrag, offen) Ausgabe bricht | (Vertrag, offen) gleiche Befunde bei zweitem Lauf |

**Zusatztabelle Paare und Modi** (Spalte *Fänger*; füllt der Architect, wenn die Operation mehrere Verstöße am selben Schlüssel meldet — hier: zwei verschiedene Zeichen derselben Zeile, dasselbe Zeichen in zwei Dateien, Befund und Lesefehler derselben Datei):

| Paar oder Modus | Entscheidung (Ort im Vertrag) | Fänger |
|---|---|---|
| (Architect) | (Vertrag, offen) | — |

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
