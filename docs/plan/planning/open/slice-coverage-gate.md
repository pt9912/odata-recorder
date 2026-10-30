# Slice slice-coverage-gate: Schwelle der Anweisungsabdeckung der Unit-Tests

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-sensoren](../welle-sensoren.md) — vierter und letzter Slice der Folge; Start, wenn `slice-abdeckung-deklaration` in `done/` liegt.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Prüfbarkeit des Quellcodes), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, Tests in der Stufe `test`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenproben außerhalb der Gate-Kette; Laufzeit), [ADR-0007](../../adr/0007-race-detektor-im-gate.md) (Muster: ein zweiter Lauf der Suite mit eigener Stufe), `AGENTS.md` §3.6 (Schwellen werden nicht ohne ADR gesenkt), Register-Eintrag `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×: ein Zweig, den kein Test ausführt, bleibt von jedem Detektor unbemerkt), Retirement-Check `docs/reviews/2026-10-25-retirement-check-hard-rules.md`, Vorbild im Nachbar-Repo `pgwire-recorder` (nur gelesen): Vertrag `SPEC-056` (Schwelle in einer Zeile `**Schwelle:** …`), dort die ADR Nr. 44.

**Berührte Spec-Stellen:** `SPEC-036` (Vertrag von `make test`, Nachzug durch den Architect) · ein neuer Vertrag oder Abschnitt zur Schwelle unter `harness/vertraege/` (Architect) · `SPEC-102` (Fälle, Zeitbudget) · `harness/README.md` §Sensors und *Nicht behauptet* („ein Coverage-Gate“ wird nachgezogen). Produkt-Spezifikation und Architektur-Sicht bleiben unberührt.
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

**Ziel:** Die Stufe `test` des `Dockerfile` (`pruefung test`) misst die Anweisungsabdeckung der Unit-Tests (Go: `-coverprofile`, `-covermode=atomic`, `-coverpkg` über die Pakete des Produkts) und endet rot, wenn sie unter der vom Architect gewählten Schwelle liegt (feste Schwelle oder Ratsche); der Bericht nennt die Blöcke ohne Lauf, und die Gegenprobe zeigt das Rot unter, das Grün an der Schwelle.

**Hintergrund:** `welle-record-replay` schließt „Coverage-Schwelle als Gate“ aus (Out-of-Scope, „kommt mit mehr Code; hier entstehen die Tests der Verträge, nicht eine Schwelle“); dieser Slice ist die Stelle, die den Punkt aufnimmt. Die Lücke steht in `harness/README.md` heute unter *Nicht behauptet*. Der Code wächst noch (Slices der Produkt-Welle in `open/`); der Architect entscheidet, ob eine feste Schwelle gegen ein wachsendes Ziel oder eine Ratsche (die Abdeckung darf nur steigen) die tragfähigere Form ist (Auftrag 2).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Mutationstests und der Beweis, dass ein ausgeführter Block auch geprüft wird — die Anweisungsabdeckung sagt, dass Code läuft, nicht dass ein Test ihn prüft (`AGENTS.md` §3.10 verlangt je Zusage eine Mutation, die der Implementer selbst rot sieht; ein Mutationswerkzeug braucht Netz für `deps` und ist in [welle-sensoren](../welle-sensoren.md) §6 ausgeschlossen). Das steht als *Ungeprüft* im Vertrag, mit Folge: ein Gate bei 100 % ist kein Beleg für Zusagen.
- Zweig- oder Pfadabdeckung und Abdeckung unter `-race` — Go misst Anweisungen; `make test-race` fährt dieselben Tests mit Detektor und bleibt ein eigenes Ziel ([ADR-0007](../../adr/0007-race-detektor-im-gate.md)). Ein Block, den kein Test ausführt, wird vom Detektor nie gemeldet (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`): die Liste der Blöcke ohne Lauf ist der Beitrag dieses Slice dazu.
- Die Änderung des Lastenhefts (eine Anforderung an die Prüfbarkeit des Quellcodes, wie sie das Nachbar-Repo als `LH-QA-07` führt) — Sache des Owners; der Architect benennt die Lücke (Auftrag 1), der Bezug dieses Slice bleibt `LH-QA-06`.
- Eine Abdeckungs-Auswertung je Anforderung — `slice-abdeckung-deklaration` (Anforderung → Test); dieser Slice misst Code, nicht Anforderungen.
- Die Abdeckung durch E2E-Tests, Integrationstests mit Prozessen oder den Lasttest zu `LH-QA-01` — es gibt keinen E2E-Test; die Nachweisart ist Gegenstand der Produktwellen.
- Eine Änderung von `AGENTS.md` und der Rollen-Dateien — nur als Diff-Vorschlag in der Closure.

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

- [ ] Die Stufe `test` misst die Anweisungsabdeckung nach dem Vertrag (Pakete, Modus, Ausnahmen, Rundung, Form der Schwelle: Auftrag 1 bis 3), vergleicht sie mit der Schwelle und endet rot darunter; die Schwelle steht an einer Stelle, im Vertrag, und der Bericht nennt Gesamtwert und Blöcke ohne Lauf; jede Zusage des Vertrags hat eine Zeile in der Tabelle *Zusage · Fänger* (§3) mit einer vom Implementer selbst rot gesehenen Mutation.
- [ ] Eine Gegenprobe nach [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) belegt das Rot unter der Schwelle, das Grün an und über ihr, die Ausnahmen und die Randformen des Profils (leer, fehlend, unlesbar); die Zusatzzeit der Gegenproben und der Stufe `test` ist gemessen und in §7 berichtet.
- [ ] Die Schwelle am Stand des Slice ist mit Messung gesetzt (Wert, Stand `git`, Zahl der Pakete und Blöcke), das Verhältnis zum wachsenden Code ist entschieden (Auftrag 2), und `harness/README.md` nennt in *Nicht behauptet* nicht mehr „ein Coverage-Gate“, sondern den Rest, den das Gate nicht trägt.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Vertrag zur Schwelle, `SPEC-036`, `SPEC-102`, Zeilen in `harness/README.md` (Sensors, *Nicht behauptet*); die Closure legt den Diff-Vorschlag für `AGENTS.md` bei, soweit eine Regel durch das Gate berührt ist (der User ändert).
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
| ein Vertrag oder Abschnitt unter `harness/vertraege/` (Schwelle) und `SPEC-036` | neu, update (Architect) | Schwelle (Wert oder Ratsche), Pakete und Ausnahmen, Modus, Rundung, Profil, Ausgang, Randformen, *Ungeprüft* |
| `docs/plan/adr/` | neu (Architect) | Entscheidung über die Form der Schwelle (fest oder Ratsche) mit Grund und Folge und über die Bindung des Gates; Zeile im ADR-Index (Muster: das Nachbar-Repo hält Messung und Schwelle in zwei ADRs) |
| `tools/pruefung/` | update | Programm der Stufe `test`: Aufruf von `go test` mit Profil, Auswertung, Vergleich mit der Schwelle, Liste der Blöcke ohne Lauf; Tests je Zeile der Tabelle unten |
| `Dockerfile` | update | Stufe `test`: `go test` mit `-coverprofile`, `-covermode=atomic`, `-coverpkg`; die Laufzeit der Stufe steigt, das Gate bleibt netzlos außer `deps` |
| Ort der Schwelle (Zeile im Vertrag, wie im Vorbild, oder eine Datei) | neu, update | eine Stelle, ein Wert; wer die Schwelle senkt, schreibt eine ADR (`AGENTS.md` §3.6); eine Ratsche braucht außerdem den Stand, von dem sie steigt |
| `harness/README.md` | update | Zeile von `make test` in §Sensors (Spalte Vertrag ≤ 200 Zeichen), *Nicht behauptet*: „ein Coverage-Gate“ entfällt, der Rest (Zweig-, Pfad-, Mutationswahrheit) wird benannt, im selben Commit |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh`, `harness/mk/gegenproben.mk` | update | Fälle zu `SPEC-102`: Profile und Kopien der Stufe (unter, an, über der Schwelle; fehlendes und leeres Profil); ein Fall fährt das Produkt nicht mit ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); `EINGABEN` um die neuen Pfade; Stempel nach dem letzten Commit unter den Eingabepfaden |

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Vertrag, Entscheidungen zu den Aufträgen 1 bis 7, Tabelle *Zusage · Fänger*, Raster, Gegenproben-Fälle, Nachzüge, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorschlag an den User).

**Betriebsregeln für alle Rollen** (Wortlaut und Gründe: [welle-sensoren](../welle-sensoren.md) §4):

- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (die Zeitgrenzen der Gegenproben und von `pruefung` sind für einen ruhigen Host bemessen; das Profil mit `-covermode=atomic` verlängert `go test` zusätzlich); Docker sonst nur über `make` (`AGENTS.md` §3.1). Gate-Läufe laufen im Hauptbaum, weil die Images feste Tags tragen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures (Profile, Kopien) mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Mengen sind Generatoren oder Schleifen (`AGENTS.md` §3.17): jede Grenze der Schwelle (−1 Einheit, genau, +1 Einheit in der Auflösung der Rundung), jede Art von Ausnahme, jede Form eines Profils. Ein Fänger an der Rundung führt den Fall, in dem zwei verglichene Werte gleich sind (`AGENTS.md` §3.15).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen.
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice. **Die Schwelle wird nie gesenkt, um den Slice grün zu bekommen** (`AGENTS.md` §3.6): ein Rot am Stand des Slice ist ein Befund an den Planner.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; der Vertrag steht in `harness/vertraege/`, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), `AGENTS.md` §3.8; das Vorbild im Nachbar-Repo wird gelesen, nicht kopiert):

1. **Gegenstand und Lücke.** Was die Abdeckung misst: Pakete (`src/` ganz, mit `-coverpkg` über alle Pakete des Moduls, auch solche ohne Test; die Module unter `tools/` mit eigenem `go.mod`, jedes einzeln oder gar nicht), erzeugter Code, die Composition Root `bootstrap` (wenig Logik, wenige Tests), `main`; welche Ausnahmen es gibt und wie sie ausgedrückt werden — eine Ausnahme ist eine Absenkung für einen Teil und verlangt einen Carveout mit Trigger und Folge-Slice (`AGENTS.md` §3.6), keine Liste ohne Frist. Das Lastenheft führt keine Anforderung an die Prüfbarkeit des Quellcodes (das Nachbar-Repo führt `LH-QA-07`): der Architect benennt die Lücke für den Owner, der Bezug bleibt `LH-QA-06`.
2. **Feste Schwelle oder Ratsche — mit Grund und Folge.** Feste Schwelle (Vorbild: eine Zeile `**Schwelle:** …` im Vertrag, Prozent der Anweisungen, ganzzahlig): einfach, aber gegen ein wachsendes Ziel entweder zu locker (Wert unter dem Ist-Stand) oder bei jedem Slice neu zu rechnen. Ratsche (die Abdeckung darf nur steigen): braucht einen gespeicherten Stand, der selbst Gegenstand eines Commits ist, und eine Regel für den Fall, dass legitimer Code die Abdeckung senkt (Slice mit Boilerplate, gelöschter Test). Zu entscheiden mit Wert (Messung am Stand des Slice, nicht geschätzt), Rundung, Auflösung (ganze Prozent oder Zehntel), Verhalten bei einem Slice der Produkt-Welle, der die Abdeckung unter die Schwelle drückt (rot, bis Tests da sind, oder Absenken mit ADR). Ob der Slice vor oder nach dem Abschluss von `welle-record-replay` beginnt, ist keine Frage des Triggers (er ist beobachtbar gesetzt), aber der Architect nennt den Abstand der Schwelle zum erwarteten Wachstum.
3. **Mechanik in Go.** `-coverprofile`, `-covermode=atomic`, `-coverpkg`; Zusammenführen mehrerer Pakete und Module; Auswertung (`go tool cover -func` oder eigenes Lesen des Profils); Liste der Blöcke ohne Lauf (Format, Länge der Ausgabe, Reihenfolge); Zusammenspiel mit `go vet` und `gofmt` in `pruefung test` (der Stufe bleibt der Verstoß gegen Format und Vet, der Schwelle ihr Ausgang); Laufzeit gegen die Zeitgrenze von `pruefung` und gegen die Bauzeit der Stufe (der Architect schätzt, der Implementer misst; `-covermode=atomic` ist langsamer als `set`); Verhältnis zu `make test-race` (zwei Läufe der Suite oder einer; hier entscheidet der Architect mit Grund, `go test -race` mit `-coverprofile` und `-covermode=atomic` ist möglich, kostet aber die Stufe `race`).
4. **Randformen** vor dem ersten Code-Commit im Vertrag (Raster in §6): Profil fehlt, ist leer oder unlesbar; Paket ohne Test (zählt mit 0 % oder nicht, mit Grund); Paket ohne Anweisung; Schwelle fehlt, ist 0, ist über 100 oder nicht numerisch; Zeile der Schwelle zweimal; Abdeckung genau an der Schwelle (die Gleichheit der beiden Werte, `AGENTS.md` §3.15); Abbruch (Zeitgrenze, Kill) mitten im Lauf; ein roter Test (Abdeckung wird dann nicht bewertet oder doch).
5. **Gegenproben.** Fälle zu `SPEC-102` (Nummern fortlaufend; Profile und Kopien der Stufe), Zeitbudget im Abschnitt *Zeitbudget* (der Architect schätzt, der Implementer misst), ob ein Fall das Produkt mitfährt (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`: die Eingabe der Fälle sind Profile und Wegwerfmodule, nicht die Pakete des Produkts; sonst wächst die Zeit des Falls mit dem Produkt und die Frist je Fall reißt).
6. **Tabelle *Zusage · Fänger*** (§3), **Raster** (§6), **Paare** (zwei Verstöße am selben Schlüssel: Abdeckung unter der Schwelle und roter Test; Ausnahme und Schwelle), **Nachzüge** nach `AGENTS.md` §3.9 (`harness/README.md` Sensors und *Nicht behauptet*, `SPEC-036`, `SPEC-102`, ADR-Index), und was an der Stufe `test` für die Aufhängung in `GATE_CHECKS` gezeigt ist (heute durch keinen Fall der Gegenproben).
7. **Erwartete Produktcode-Größe** für die Marke des Planners; ob das Prüfprogramm nebenläufig ist (zählt × 85); ob das Lesen des Profils ein Leser fremder Bytes im Sinn von `AGENTS.md` §3.14 ist (× 65).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 6) | — | — |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-abdeckung-deklaration` liegt in `done/` (Reihenfolge der Welle: beide ändern `harness/README.md`, die Gegenproben, den Stempel und die Gate-Kette), und der Architect hat den Vertrag, die Entscheidungen zu den Aufträgen 1 bis 7 und die Tabelle geliefert, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner schreibt nach dem Architect den Block `groessenmarke` ([SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 3) in diese Sektion und setzt die Marke mit `make groessenmarke` im eigenen Commit vor dem ersten Code-Commit. Aufwand nach Einschätzung des Planners: M.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Schätzung des Architect mal 1,25 liegt über 3000 (`AGENTS.md` §3.14); dann trennt der Planner vor dem Code die Module unter `tools/` vom Produkt (`src/`).
- `in-progress` → `open` (blockiert): die Abdeckung am Stand des Slice liegt so niedrig, dass eine ehrliche Schwelle ein Rot erzeugt, das nur Tests der Produkt-Welle beheben; dann entscheidet der Planner mit dem User über Ratsche, Stand der Messung oder einen Carveout (`AGENTS.md` §3.6), nicht der Implementer.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, `make gates` und `make gegenproben` grün im Hauptbaum, Review- und Verifikations-Report unter `docs/reviews/`, Closure-Notiz in §7. Der Lerneintrag nennt den Sensor mit `liegt in` (`Makefile:test` oder die Stufe, mit dem Anker `seit welle-sensoren`), die Liste der Blöcke ohne Lauf am Stand des Slice (Zahl) und den Diff-Vorschlag an den User; mit diesem Slice schließt die letzte Sensor-Zeile der Tabelle der Welle.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Ein Gate bei „X Prozent“ ist kein Beleg für Zusagen: Code läuft, ohne dass ein Test sein Ergebnis prüft; wer die Zahl für Qualität hält, streicht Fänger zu früh — **Ausgang:** offen bis zur Closure; erwartet weiter offen: → `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`; der Vertrag nennt es unter *Ungeprüft*.
- Die Schwelle läuft gegen ein wachsendes Ziel: Die Produkt-Welle (`slice-workspace-use-case`, `slice-replay-service`, `slice-http-adapter`, `slice-bootstrap-m2`) fügt Code hinzu, dessen Tests im selben Slice entstehen, aber ein Slice, der die Abdeckung unter die Schwelle drückt, färbt das Gate rot, und die naheliegende Antwort ist, die Schwelle zu senken — **Ausgang:** offen bis zur Closure; entschieden in Auftrag 2 (Form und Regel für den Fall); eine Senkung ist eine ADR (`AGENTS.md` §3.6).
- Laufzeit: `-covermode=atomic` und `-coverpkg` verlängern die Stufe `test`; die Zeitgrenzen von `pruefung` sind für einen ruhigen Host bemessen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`: die Zeit wächst mit dem Produkt) — **Ausgang:** offen bis zur Closure; der Implementer misst die Stufe vorher und nachher, die Zeitgrenze steht im Vertrag.
- Laufzeit der Gegenproben: die Fälle verlängern den Lauf, dessen Re-Evaluierungs-Trigger ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): Lauf über 30 s bei geänderter Eingabe; gemessen 56,9 s Schritt) beim User unentschieden ist — **Ausgang:** offen bis zur Closure; der Slice misst die Zusatzzeit und berichtet sie, er entscheidet den Trigger nicht; weiter offen: → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (Adresse: User).
- Ausnahmen (erzeugter Code, Composition Root) werden als Liste geführt und wachsen — **Ausgang:** offen bis zur Closure; entschieden in Auftrag 1 (Carveout mit Trigger und Folge-Slice statt Liste ohne Frist).
- Module unter `tools/` haben eigenes `go.mod`; `-coverpkg` über Module hinweg ist nicht möglich, und `make test-race` meldet sie ebenfalls nicht (`SPEC-122` Nr. 8 und 9) — **Ausgang:** offen bis zur Closure; entschieden in Auftrag 1; ein ausgenommenes Modul steht als benannter Rest unter *Nicht behauptet*.
- Die Aufhängung des Gates in `GATE_CHECKS` ist durch keinen Fall der Gegenproben gezeigt — **Ausgang:** offen bis zur Closure; der Architect entscheidet, ob ein Fall sie zeigt (Auftrag 6), sonst akzeptiertes Negativ mit Grund und Folge.

**Raster der Randformen** (`AGENTS.md` §3.16; füllt der Architect vor dem ersten Code-Commit: jede Zelle nennt ihren Ort im Vertrag oder „gilt nicht“ mit Grund; eine leere Zelle ist ein Fehler des Plans; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Lesen der Schwelle | (Vertrag, offen) Zeile fehlt | (Vertrag, offen) Wert leer | (Vertrag, offen) 0, über 100, nicht numerisch, mit Dezimalstelle bei ganzzahliger Form; eingerückte Zeile | (Vertrag, offen) gilt nicht, Lesen einer Datei | (Vertrag, offen) Zeile zweimal |
| Messen der Abdeckung (Profil) | (Vertrag, offen) Profil fehlt | (Vertrag, offen) Profil leer; Paket ohne Anweisung | (Vertrag, offen) Paket ohne Test; Zeile in fremdem Format; Block doppelt; Pfad außerhalb des Moduls | (Vertrag, offen) Zeitgrenze, Kill mitten im Lauf | (Vertrag, offen) zweiter Lauf, gleicher Wert |
| Vergleichen mit der Schwelle | (Vertrag, offen) Abdeckung nicht bestimmbar | (Vertrag, offen) Abdeckung 0 | (Vertrag, offen) unter, genau an (Gleichheit), über der Schwelle; Rundung an der Grenze | (Vertrag, offen) gilt nicht, reine Rechnung | (Vertrag, offen) zweiter Lauf, gleiches Urteil |
| Ausnahmen | (Vertrag, offen) Eintrag ohne Grund | (Vertrag, offen) leerer Eintrag | (Vertrag, offen) Eintrag trifft kein Paket; Eintrag trifft ein Paket mit Tests | (Vertrag, offen) gilt nicht, Konfiguration | (Vertrag, offen) Eintrag doppelt |
| Liste der Blöcke ohne Lauf | (Vertrag, offen) kein Block ohne Lauf | (Vertrag, offen) alle Blöcke ohne Lauf | (Vertrag, offen) sehr viele Blöcke (Länge der Ausgabe) | (Vertrag, offen) Ausgabe bricht | (Vertrag, offen) gleiche Reihenfolge beim zweiten Lauf |
| Verhältnis zu rotem Test, `vet`, `gofmt` | (Vertrag, offen) kein Test im Modul | (Vertrag, offen) leeres Modul | (Vertrag, offen) roter Test bei hoher Abdeckung; Verstoß gegen Format bei hoher Abdeckung | (Vertrag, offen) Zeitgrenze der Suite | (Vertrag, offen) zweiter Lauf, gleiches Urteil |

**Zusatztabelle Paare und Modi** (Spalte *Fänger*; füllt der Architect: je Paar von Verstößen am selben Schlüssel — Abdeckung unter der Schwelle **und** roter Test, Abdeckung unter der Schwelle **und** Verstoß gegen `gofmt`, Ausnahme **und** unbekanntes Paket — und je Modus (feste Schwelle, Ratsche) eine Zeile):

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

*Wird bei der Closure gefüllt.* Die Closure trägt zusätzlich: Gesamtwert der Abdeckung und Zahl der Blöcke ohne Lauf am Stand des Slice, die gemessene Zusatzzeit der Stufe `test` und der Gegenproben (Schritt und Gesamtlauf, vorher und nachher, Zahl der Läufe, Zustand des Hosts) und die Diff-Vorschläge an den User.

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, Stand 2026-10-30, 19 Einträge; Zahlen aus den Dateien unter `evidence/`). Für diesen Slice zählen:
- `zweig-ohne-lauf-unter-dem-detektor` (1×, offen): die Liste der Blöcke ohne Lauf ist das Mittel, das der Eintrag nennt; ein weiterer Vorgang hier wäre der zweite.
- `gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, offen): trifft zu und ist das größte Risiko des Slice (Fall, der das Produkt mitfährt); `stempel-gehoert-zu-keinem-committeten-stand` (1×, offen), `betriebsregel-fuer-repo-dateien-verletzt` (1×, offen), `hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×, offen): Stempel nach dem letzten Commit, Betriebsregeln in §3, Hüllen von Argumenten (Profile, Fake von `go test`) tragen jede Schnittstelle des Originals.
- `rate-ohne-detektor-schwankt-mit-der-last` (2×, offen) und `panik-eines-fakes-verdeckt-den-benannten-test` (2×, offen): trifft zu, falls ein Fänger an der Laufzeit oder an einer Panik eines Fakes hängt; ein weiterer Vorgang macht aus der Notiz eine Lücke (eigener Slice).
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×), `akzeptiertes-negativ-ohne-folge` (10×), `randform-erst-im-review-entschieden` (20×), `universalaussage-im-vertrag-ungeprueft` (9×): Tabelle, Raster, und der Satz „jede Anweisung der Pakete zählt“ mit Fänger über alle Elemente der Menge.
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (16×; Ausgang `geplant`, `slice-groessenmarke-werkzeug`): Marke mit dem Werkzeug, Schnitt vor dem Code.
- `berichtigung-erreicht-nicht-jede-stelle` (11×, §3.9): die Schwelle steht an einer Stelle; Zeilen in `harness/README.md`, `SPEC-036` und Gegenproben ziehen im selben Commit nach.
Unter der Schwelle ohne Bezug: die übrigen Einträge.

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF (Default `*`, Greenfield: Neues Repo ohne Bestandscode; die Spec führt, der Code folgt).
