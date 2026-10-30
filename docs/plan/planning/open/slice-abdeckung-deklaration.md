# Slice slice-abdeckung-deklaration: Abdeckungs-Deklaration über den Tests, Anforderung → Test maschinenlesbar

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-sensoren](../welle-sensoren.md) — dritter Slice der Folge, erster von Teil 2; Start, wenn `welle-record-replay` und `slice-guide-groessen-gate` in `done/` liegen.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an Gates), die Anforderungen des Lastenhefts als Gegenstand der Deklaration (`LH-FA-01` bis `LH-FA-18` mit den Akzeptanzkriterien *Happy Path*, *Boundary*, *Negative*; `LH-QA-*` und `LH-RB-*` mit *Messmethode*), `AGENTS.md` §3.15 und §3.16 (Tabelle *Zusage · Fänger*, Raster), Register-Einträge `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×), `BEO-REPO/randform-erst-im-review-entschieden` (20×), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×), Retirement-Check `docs/reviews/2026-10-25-retirement-check-hard-rules.md` §5 (Vorschlag `slice-plan-form-sensor`) und §6, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md).

**Berührte Spec-Stellen:** ein neuer Vertrag unter `harness/vertraege/` (Kennung nach `harness/vertraege/README.md` §Kennungen, vom Architect, Vorbild: der Vertrag `SPEC-048` des Nachbar-Repos `pgwire-recorder`, übernommen nur nach Entscheidung) · `SPEC-102` (Fälle, Zeitbudget) · `SPEC-040` (falls `trace` im Doku-Gate aktiv wird) · `.d-check.yml` (Block `trace`) · `harness/README.md` §Sensors. `spec/lastenheft.md` bleibt unberührt (es ist die Quelle der Anforderungen, nicht Gegenstand der Änderung).
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

**Ziel:** Jeder Test, der eine Anforderung des Lastenhefts trägt, sagt es maschinenlesbar über der Funktion (Form und Pfade nach dem Vertrag), ein Werkzeug erzeugt daraus Tabellen *Anforderung → Test* je Pfad (Happy, Boundary, Negative bzw. Messung), und ein Gate in der Kette `make gates` wird rot, wenn eine Deklaration eine Fehlform hat oder eine erzeugte Tabelle nicht dem erzeugten Stand entspricht; die Vollständigkeit je Anforderung ist sichtbar (`make doc-trace` oder der vom Architect gewählte Weg).

**Das Vorbild** liegt im Nachbar-Repo `pgwire-recorder` (nur gelesen, dort nichts geändert): Deklaration `// Abdeckung: LH-FA-13/Boundary — <Kurzbeschreibung>` über dem Test, Skript `tools/test/abdeckung.sh` (schreibt die Tabellen `abdeckung-unit`, `-e2e`, `-gate`, `-gesamt`, `-vollstaendig` unter docs/user des Nachbar-Repos; `--check` prüft nur), Vertrag `SPEC-048`, Make-Fragment `abdeckung.mk` (`abdeckung-check` in `GATE_CHECKS`, `abdeckung-gegenprobe` in den Gegenproben), Block `trace` in `.d-check.yml`, ADRs 0028 (Abdeckung), 0033 (Gate-Nachweise, geteilte Messung). **Nichts davon ist hier gesetzt:** Dieses Repo ist ein Go-Hexagon mit anderem Lastenheft (Akzeptanzkriterien *Happy Path*, *Boundary*, *Negative* bei `LH-FA`, *Messmethode* bei `LH-QA`), deutscher Sprache, dem Gegenproben-Rahmen nach [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) und Werkzeugen in Go; der Architect entscheidet je Punkt (Aufträge), was passt und was nicht.

**Messwerte des Planners am Stand `0c524e2`** (zur Orientierung; der Implementer misst am Start neu): 618 Testfunktionen in 105 Dateien `*_test.go` unter `src/` und `tools/`; keine Deklaration; das Verzeichnis docs/user (Rang 6 der Source Precedence, „falls vorhanden“) existiert nicht; `make doc-trace` und der Block `trace` sind in `.d-check.yml` nicht gesetzt; `make e2e-abdeckung` (Erzeuger des Werkzeugs `ai-harness-init`, Quelle `tools/harness/selbstpruefung.sh`, kein Gate) besteht; es gibt keinen E2E-Test.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung am Lastenheft (Anforderungen, Akzeptanzkriterien, Überschriften) — Sache des Owners; passt die Form der Anforderungen nicht zum Vertrag (`LH-QA-*` ohne Akzeptanzkriterien, Teilkennungen wie `LH-QA-07.a`), nennt der Architect die Lücke, ändert sie nicht.
- Eine Änderung von `AGENTS.md`, Commands, Rollen-Typen oder Skills — nur als Diff-Vorschlag in der Closure; insbesondere die Frage, ob §3.15 und §3.16 mit der Deklaration kürzer werden, beantwortet der Architect (Auftrag 3), der User ändert.
- Das Nachrüsten der Deklaration über **alle** Bestandstests (618 Funktionen), falls der Architect den Umfang des Backfills größer als den Rest des Slice findet — dann trennt der Planner vor dem Code (`AGENTS.md` §3.14) einen Folge-Slice ab; Adresse: Planner (Rückgabe des Architect, Ausgang *offen* in §6, `AGENTS.md` §3.13); es gibt heute keinen Slice, der den Punkt annimmt.
- Die Nachweisart E2E und das Werkzeug `make e2e-abdeckung` — es gibt keinen E2E-Test; der M2-Test aus `slice-bootstrap-m2` ist der erste. Wie sich beide Erzeuger zueinander verhalten, benennt der Architect (Auftrag 2); verändert wird `e2e-abdeckung` nicht.
- Die Aktivierung von `make doc-trace-gate` (`--require-complete`, rot bei jeder Waise) — solange Slices der Produkt-Welle in `open/` liegen und M2 nicht erreicht ist, hat nahezu jede Anforderung eine Waise; der Slice liefert die beratende Sicht (`make doc-trace`) und benennt den Zeitpunkt, an dem das Gate sinnvoll wird (Auftrag 5).
- Die Anweisungsabdeckung des Codes (Prozent der Anweisungen) — ein anderer Vorgang, ein anderes Gate: `slice-coverage-gate`. Die Deklaration sagt, welcher Test welche Anforderung trägt; sie sagt nicht, wie viel Code er erreicht.
- Der Beweis, dass ein Test tatsächlich tut, was seine Deklaration sagt — nicht berechenbar; die Deklaration ist eine Zusage des Autors, das Gate prüft ihre Form und die Existenz der Anforderung, nicht ihre Wahrheit. Das steht als *Ungeprüft* im Vertrag, mit Folge: Reviewer und Verifier lösen die Deklaration weiter gegen die Anforderung auf.

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

- [ ] Der Vertrag (Architect) legt Form der Deklaration, Pfade, Nachweisarten, erzeugte Tabellen, Ausgang und Fehlformen fest; das Werkzeug erzeugt die Tabellen und prüft sie mit `--check` (oder der vom Architect gewählten Form), das Ziel hängt an `GATE_CHECKS`, und jede Zusage des Vertrags hat eine Zeile in der Tabelle *Zusage · Fänger* (§3) mit einer vom Implementer selbst rot gesehenen Mutation.
- [ ] Die Gegenprobe nach [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) belegt jede Fehlform (ein Fall je Fehlform und je Pfad der Menge) und die Aktualität der Tabellen; die Zusatzzeit der Gegenproben ist gemessen und in §7 berichtet.
- [ ] Der Umfang der Pflicht ist umgesetzt, wie der Architect ihn in Auftrag 6 entschieden hat (neue Tests, Tests mit Anforderung, alle Tests), und die erzeugten Tabellen stehen im Repo; die Sicht auf die Vollständigkeit (`make doc-trace` mit `trace` in `.d-check.yml`, oder der benannte Ersatz) läuft, ohne ein Gate zu sein, das heute rot ist.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Vertrag und Index `harness/vertraege/README.md`, Zeile(n) in `harness/README.md` §Sensors und *Nicht behauptet*, `SPEC-102` (Fälle, Zeitbudget); die Closure legt die Diff-Vorschläge für `AGENTS.md` §3.15 und §3.16, für `implement-slice` und für den Reviewer-Skill bei (der User ändert).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag, darin die Antwort zu `slice-plan-form-sensor` (Auftrag 3) und die Tabelle der Welle zu §3.15 und §3.16.
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
| ein Vertrag unter `harness/vertraege/` und die Zeile im Index dort | neu (Architect) | Form der Deklaration, Pfade, Nachweisarten, Tabellen, Ausgang, Fehlformen, *Ungeprüft* |
| `docs/plan/adr/` | neu (Architect) | Entscheidung über die Übernahme des Musters (Abdeckung je Anforderung und Pfad) und über den Ort der erzeugten Tabellen; Zeile im ADR-Index; weitere ADR nur, wenn eine Entscheidung (Backfill, Beratung gegen Gate) sie verlangt |
| Werkzeug (`tools/`, Sprache und Ort: Auftrag 1) | neu | liest die Deklarationen in `*_test.go`, erzeugt die Tabellen, prüft mit `--check`; Tabellen kommen als Tar-Strom oder Datei ohne schreibbaren Mount des Repos (Muster `schema-generate`, `AGENTS.md` §3.1) |
| Fragment unter `harness/mk/`, `Dockerfile`, `.dockerignore` | neu, update | Ziele (Erzeugen: Werkzeug, kein Gate; Prüfen: Gate in `GATE_CHECKS`), Stufe nach Muster `groessenmarke` oder `pruefung` |
| erzeugte Tabellen (Ort: Auftrag 2) | neu (erzeugt) | je Pfad und Nachweisart eine Tabelle und die Gesamtsicht; im Repo geführt, damit `--check` vergleichen kann |
| `.d-check.yml` | update | Block `trace` (Anforderungen aus `spec/lastenheft.md`, Slices in `done/`, Nachweis-Datei der vollständig belegten Anforderungen); `make doc-trace` beratend |
| Testdateien `*_test.go` | update | Deklarationen im Umfang des Auftrags 6; keine Änderung der Testlogik |
| `harness/README.md` | update | Zeilen der Ziele in §Sensors; *Nicht behauptet* nennt, was das Gate nicht prüft (Wahrheit der Deklaration, Vollständigkeit) |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` | update | Fälle zu `SPEC-102`: je Fehlform ein Fall in einem Wegwerfverzeichnis ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); `EINGABEN` um die neuen Pfade; Stempel nach dem letzten Commit unter den Eingabepfaden |
| Testdateien des Werkzeugs | neu | eine Zeile der Tabelle unten je Test; Mengen als Schleifen (`AGENTS.md` §3.17) |

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Vertrag, Entscheidungen zu den Aufträgen 1 bis 8, Tabelle *Zusage · Fänger*, Raster, Gegenproben-Fälle, Nachzüge, erwartete Produktcode-Größe, Antwort zu `slice-plan-form-sensor`) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit; legt, falls der Architect es verlangt, `slice-plan-form-sensor` oder den Backfill als Slice an) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorschlag an den User).

**Betriebsregeln für alle Rollen** (Wortlaut und Gründe: [welle-sensoren](../welle-sensoren.md) §4):

- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen; Docker sonst nur über `make` (`AGENTS.md` §3.1). Gate-Läufe laufen im Hauptbaum, weil die Images feste Tags tragen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis. Das Setzen von Deklarationen über Hunderte Tests ist ein Lauf des Werkzeugs oder eine Folge von `Edit`, kein Skript auf den Repo-Dateien.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); die Deklaration trägt ein „ — “ (Gedankenstrich) im Muster des Vorbilds: der Gedankenstrich ist ein Zeichen jenseits von ASCII, die Fixtures der Fehlformen bilden ihn aus Bytewerten.
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (`AGENTS.md` §3.17): jeder Pfad, jede Nachweisart, jede Fehlform, jede Art von Testfunktion.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen.
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; der Vertrag steht in `harness/vertraege/`, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), `AGENTS.md` §3.8; das Vorbild im Nachbar-Repo wird gelesen, nicht kopiert):

1. **Übernahme des Musters, Punkt für Punkt.** Form der Deklaration (Zeile über `func Test…`, Trenner, Fortsetzungszeilen, Fehlformen: eingerückt, ohne Trenner, nicht unmittelbar vor einem Test); Pfade je Anforderungsart (`LH-FA`: Happy, Boundary, Negative — das Lastenheft schreibt *Happy Path*; `LH-QA`/`LH-RB`: Messung, geteilte Messung; Teilkennungen wie `LH-QA-07.a`); welche Funktionen zählen (`Test…`, Beispiele, `Benchmark…`, `Fuzz…`, `TestMain`, Hilfsfunktionen, Teiltests über `t.Run`); Sprache und Ort des Werkzeugs (Go nach Muster `tools/groessenmarke` oder ein Skript; Treiber auf dem Host, `docker cp`, netzlos, ohne schreibbaren Mount, `AGENTS.md` §3.1, [ADR-0001](../../adr/0001-implementierungssprache-go.md)). Das Ergebnis je Punkt: übernommen, geändert (mit Grund) oder gestrichen (mit Grund und Folge).
2. **Nachweisarten und Ort.** Unit, E2E, Gate: Welche Arten gibt es in diesem Repo (kein E2E-Test bis `slice-bootstrap-m2`; Gate-Nachweise durch Gegenproben, die einen Test ersetzen), wie der Ort eines Tests die Art bestimmt, und wie sich das Werkzeug zu `make e2e-abdeckung` verhält (ein Erzeuger, zwei, oder getrennt mit Grund). Ort der erzeugten Tabellen: docs/user (Rang 6 der Source Precedence, besteht nicht) oder ein anderer; ob sie im Repo stehen müssen (Vergleich mit `--check`) oder als Tar-Strom des Gates vergleichbar sind.
3. **Das Verhältnis zur Tabelle *Zusage · Fänger* und zu `slice-plan-form-sensor`** (Entscheidung des Planners: kein fünfter Slice, die Frage ist dieser Auftrag, [welle-sensoren](../welle-sensoren.md) §4). Zu beantworten mit Grund und Folge: (a) Die Deklaration verbindet **Anforderung des Lastenhefts → Test**; die Tabelle des Plans verbindet **Zusage des Vertrags (`SPEC-…`, Nummer) → Fänger**. Kann die Deklaration beide Ebenen tragen (Kennungen aus `harness/vertraege/` und `spec/spezifikation.md` als Gegenstand neben `LH-*`; ein Test nennt die Spec nach oben, nicht umgekehrt), und was ändert das an der Pflicht der Pläne, die Tabelle zu führen? (b) Welche Stücke von `AGENTS.md` §3.15 und §3.16 trägt die Deklaration (Zusage ohne Fänger, Fall je Pfad), welche nicht (Fänger sind eine **Mutation** rot; Ort der Zelle im Raster; Allaussage mit Schleife)? (c) Braucht die Form der Pläne (Tabelle *Zusage · Fänger* und Raster in §6 mit Spalte *Fänger*, keine leere Zelle) eine `structure`-Regel auf die Slice-Pläne? Der Retirement-Check schätzt sie auf M und nennt die Grenzen: Wirkung bei mehreren Tabellen je Abschnitt ungeprüft; Archiv-Stubs und Pläne ohne Vertrag brauchen `skip-pattern` oder `exempt-paths`. Antwort „ja“: Rückgabe an den Planner mit Wert und Grund (§6, Ausgang *offen*, Adresse Planner); der Planner legt `slice-plan-form-sensor` an. Antwort „nein, getragen“ oder „nein, nicht berechenbar“: der Grund steht im Vertrag unter *Ungeprüft*.
4. **Randformen und Pfade** der Deklaration vor dem ersten Code-Commit im Vertrag (Raster in §6): fehlender, leerer, unbekannter Wert je Teil der Deklaration; Anforderung nicht im Lastenheft; Pfad passt nicht zur Art; Widerspruch zwischen Deklarationen (geteilte Messung); doppelte Deklaration derselben Anforderung und desselben Pfads an einem Test; Testdatei ohne Test; Test ohne Deklaration im Umfang des Auftrags 6.
5. **`trace` und `make doc-trace`.** Schlüssel des Blocks am Pin des Images (am Image gelesen, nicht aus dem Vorbild geschlossen: `requirements`, `slices`, `coverage`), beratend oder Gate; wenn beratend, der beobachtbare Zeitpunkt, ab dem `--require-complete` gilt (Vorschlag: ein Meilenstein, z. B. M2, oder „keine Waise unter den `Muss`-Anforderungen“), mit Grund; wie `make doc-trace` ohne Docker-Mount-Schreiben läuft.
6. **Umfang der Pflicht und Schnitt.** Pflicht für neue Tests ab einem Stichtag, für Tests mit Anforderungsbezug, oder für alle; Gate-Form dafür (Ratsche: die Zahl der Tests ohne Deklaration darf nur sinken, oder Pflicht je Datei); Zahl der Deklarationen im Bestand (618 Funktionen am Start) und die geschätzte Größe des Backfills in der Metrik von `AGENTS.md` §3.14. Liegt die Schätzung × 1,25 über 3000, trennt der Planner vor dem Code (Backfill als eigener Slice, Adresse: Planner).
7. **Gegenproben.** Fälle zu `SPEC-102` (Nummern fortlaufend; je Fehlform und je Pfad der Menge ein Fall; Fixtures aus Bytewerten), Zeitbudget (der Architect schätzt, der Implementer misst), ob ein Fall das Produkt mitfährt (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`: die Eingabe enthält nur Testdateien als Fixtures, keine Pakete des Produkts).
8. **Tabelle *Zusage · Fänger*** (§3), **Raster** (§6), **Paare** von Verstößen am selben Schlüssel, **Nachzüge** nach `AGENTS.md` §3.9 (`harness/README.md`, `SPEC-102`, Index der Verträge, ADR-Index), ob die Größenregel für einen Leser fremder Bytes gilt (Go-Quelltext ist ein Format, das dieses Repo nicht schreibt: × 65 statt × 40, `AGENTS.md` §3.14) und die **erwartete Produktcode-Größe** für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 8) | — | — |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` liegt in `done/` (der Backfill über die Bestandstests berührt sonst die Dateien, die M2 schreibt; [welle-sensoren](../welle-sensoren.md) §2) und `slice-guide-groessen-gate` liegt in `done/` (Reihenfolge der Welle: beide ändern `.d-check.yml`, `harness/README.md`, die Gegenproben und den Stempel), und der Architect hat den Vertrag, die Entscheidungen zu den Aufträgen 1 bis 8 und die Antwort zu `slice-plan-form-sensor` geliefert, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner schreibt nach dem Architect den Block `groessenmarke` ([SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 3) in diese Sektion und setzt die Marke mit `make groessenmarke` im eigenen Commit vor dem ersten Code-Commit. Aufwand nach Einschätzung des Planners: M, bei Backfill über alle Tests eher L; das Werkzeug entscheidet nach der Tabelle des Architect.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Schätzung des Architect mal 1,25 liegt über 3000 (`AGENTS.md` §3.14), oder der Backfill (Auftrag 6) trägt mehr als die Hälfte; dann trennt der Planner vor dem Code den Backfill als eigenen Slice.
- `in-progress` → `open` (blockiert): der Architect findet für die Erzeugung der Tabellen keinen Weg ohne schreibbaren Mount und ohne neue Host-Abhängigkeit über `bash`, `git` und `docker` hinaus (Muster `schema-generate` reicht nicht); die Frage geht an den User.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, `make gates` und `make gegenproben` grün im Hauptbaum, Review- und Verifikations-Report unter `docs/reviews/`, Closure-Notiz in §7. Der Lerneintrag nennt den Sensor mit `liegt in` (Ziel, mit dem Anker `seit welle-sensoren`), die Antwort zu `slice-plan-form-sensor` und zu den Teilen von `AGENTS.md` §3.15 und §3.16, die das Gate trägt (Tabelle der Welle), samt Diff-Vorschlägen an den User.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Deklaration ist eine Zusage des Autors; ein Test kann `Happy` deklarieren und etwas anderes prüfen. Das Gate sieht Form und Existenz, nicht Wahrheit; wer die Tabelle für einen Beleg hält, irrt — **Ausgang:** offen bis zur Closure; erwartet weiter offen: → `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (die Tabelle sagt mehr zu, als das Gate prüft); der Vertrag nennt es unter *Ungeprüft*, Reviewer und Verifier lösen weiter gegen die Anforderung auf.
- Zwei Ebenen, eine Tabelle: Anforderung → Test (Deklaration) gegen Zusage des Vertrags → Fänger (Plan, `AGENTS.md` §3.15) — wird die zweite stillschweigend durch die erste ersetzt, geht die Mutation je Zusage verloren — **Ausgang:** offen bis zur Closure; entschieden in Auftrag 3.
- `slice-plan-form-sensor`: Das Ergebnis „Form der Pläne nicht getragen“ legt einen Slice fest, den diese Welle nicht plant — **Ausgang:** offen bis zur Closure; Adresse: Planner; ein Slice entsteht nur auf Rückgabe des Architect (Auftrag 3), nicht auf Vorrat.
- Der Backfill über 618 Testfunktionen ist größer als der Rest des Slice oder wirft die Marke — **Ausgang:** offen bis zur Closure; entschieden in Auftrag 6, geschnitten vor dem Code, nicht nach der Messung (`AGENTS.md` §3.14, `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`).
- Der Ort docs/user besteht nicht; die Source Precedence nennt ihn als Rang 6 („falls vorhanden“); ein neuer Ort für erzeugte Dateien ist eine Strukturentscheidung — **Ausgang:** offen bis zur Closure; entschieden in Auftrag 2 (ADR, falls der Ort Rang oder Scan-Regeln von `.d-check.yml` berührt).
- `make doc-trace` ist beratend und hat heute fast nur Waisen (Slices der Produkt-Welle in `open/`, M2 nicht erreicht); ein Bericht, der nur Waisen zeigt, wird nicht gelesen — **Ausgang:** offen bis zur Closure; entschieden in Auftrag 5 (Zeitpunkt des Gates beobachtbar).
- Die Aufhängung der Ziele in `GATE_CHECKS` ist durch keinen Fall der Gegenproben gezeigt (heute für die bestehenden Ziele in *Nicht behauptet*) — **Ausgang:** offen bis zur Closure; der Architect entscheidet, ob ein Fall sie zeigt, sonst akzeptiertes Negativ mit Grund und Folge im Vertrag.
- Laufzeit der Gegenproben: viele Fehlformen heißen viele Fälle; der Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Lauf über 30 s bei geänderter Eingabe; gemessen 56,9 s Schritt) ist beim User unentschieden — **Ausgang:** offen bis zur Closure; der Slice misst die Zusatzzeit und berichtet sie, er entscheidet den Trigger nicht; weiter offen: → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (Adresse: User).

**Raster der Randformen** (`AGENTS.md` §3.16; füllt der Architect vor dem ersten Code-Commit: jede Zelle nennt ihren Ort im Vertrag oder „gilt nicht“ mit Grund; eine leere Zelle ist ein Fehler des Plans; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Lesen einer Deklaration (Kopfzeile) | (Vertrag, offen) Deklaration ohne Trenner; ohne Anforderung; ohne Pfad | (Vertrag, offen) leere Kurzbeschreibung; leere Anforderungsliste | (Vertrag, offen) eingerückt; Leerzeile vor dem Test; Fortsetzungszeile; Deklaration vor Nicht-Test | (Vertrag, offen) Datei bricht mitten in der Deklaration | (Vertrag, offen) zwei Deklarationen über einem Test |
| Auflösen von Anforderung und Pfad | (Vertrag, offen) Anforderung nicht im Lastenheft | (Vertrag, offen) Pfad leer | (Vertrag, offen) Pfad passt nicht zur Art (`Messung` bei `LH-FA`); Teilkennung `LH-QA-07.a`; Kennung aus `harness/vertraege/` | (Vertrag, offen) Lastenheft nicht lesbar | (Vertrag, offen) dasselbe Paar an zwei Tests; Widerspruch bei geteilter Messung |
| Bestimmen der Testmenge | (Vertrag, offen) kein `_test.go` | (Vertrag, offen) Testdatei ohne Test | (Vertrag, offen) Test ohne Deklaration im Umfang; Beispiel, Benchmark, Fuzz, `TestMain`; Datei in `testdata/`, `vendor/`, `.harness/` | (Vertrag, offen) Lesefehler | (Vertrag, offen) dieselbe Datei über zwei Wurzeln (Module unter `tools/`) |
| Erzeugen der Tabellen | (Vertrag, offen) Ziel-Verzeichnis fehlt | (Vertrag, offen) keine Deklaration (leere Tabellen) | (Vertrag, offen) senkrechter Strich in der Kurzbeschreibung (Maskierung); Sortierung bei gleichen Schlüsseln | (Vertrag, offen) Abbruch beim Schreiben (halbe Tabelle) | (Vertrag, offen) zweiter Lauf ohne Änderung schreibt nichts |
| Prüfen (`--check`) | (Vertrag, offen) Tabelle fehlt | (Vertrag, offen) Tabelle leer, Stand nicht leer | (Vertrag, offen) Tabelle von Hand geändert; veraltet nach Änderung des Tests | (Vertrag, offen) Abbruch beim Vergleich | (Vertrag, offen) zweiter Lauf, gleiche Meldung |
| Vollständigkeit je Anforderung | (Vertrag, offen) Anforderung ohne Deklaration | (Vertrag, offen) nur ein Pfad von drei | (Vertrag, offen) Nachweisart mehrfach für einen Pfad | (Vertrag, offen) gilt nicht, reine Auswertung | (Vertrag, offen) Anforderung in zwei Nachweisarten |

**Zusatztabelle Paare und Modi** (Spalte *Fänger*; füllt der Architect: je Paar von Verstößen am selben Schlüssel — Fehlform der Deklaration und unbekannte Anforderung am selben Test, doppelte Deklaration und Widerspruch, `--check` und Fehlform —, die Reihenfolge der Gründe je Element und die Modi Erzeugen/Prüfen):

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

*Wird bei der Closure gefüllt.* Die Closure trägt zusätzlich: die Antwort zu `slice-plan-form-sensor` (Auftrag 3), die gemessene Zusatzzeit der Gegenproben (Schritt und Gesamtlauf, vorher und nachher, Zahl der Läufe, Zustand des Hosts), die Zahl der Deklarationen und der Tests ohne Deklaration, und die Diff-Vorschläge an den User.

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
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, verkörpert in `AGENTS.md` §3.15), `akzeptiertes-negativ-ohne-folge` (10×, §3.15), `randform-erst-im-review-entschieden` (20×, §3.16), `universalaussage-im-vertrag-ungeprueft` (9×, §3.16): der Gegenstand; die Deklaration soll Teile dieser Regeln mechanisch tragen (Auftrag 3), und der Vertrag des Werkzeugs führt zu jeder Zusage einen Fänger, was das Werkzeug nicht prüft, steht unter *Ungeprüft* mit Grund und Folge.
- `suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, verkörpert in `implement-slice` Schritt 14, Implementer-Typ, Reviewer-Skill) und `fix-deckt-die-genannte-mutation-nicht-die-klasse` (5×, §3.17): Mengen des Werkzeugs (Pfade, Nachweisarten, Fehlformen, Funktionsarten) als Generatoren und Schleifen.
- `berichtigung-erreicht-nicht-jede-stelle` (11×, §3.9): die Form der Deklaration steht im Vertrag, im Werkzeug, in den Gegenproben, in den erzeugten Tabellen und in den Tests; Nachzug aller Stellen im selben Commit.
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (4×, §3.18): der Gedankenstrich der Deklaration und die Fixtures der Fehlformen aus Bytewerten.
- `gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, offen), `stempel-gehoert-zu-keinem-committeten-stand` (1×, offen), `betriebsregel-fuer-repo-dateien-verletzt` (1×, offen), `hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×, offen): wie in den Slices zuvor (Gegenprobe ohne Produkt, Stempel, Betriebsregeln in §3, Hüllen tragen jede Schnittstelle).
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (16×; Ausgang `geplant`, `slice-groessenmarke-werkzeug`): der Backfill (Auftrag 6) ist die Stelle, an der die Marke reißen kann; Schnitt vor dem Code, Marke mit dem Werkzeug.
- `rate-ohne-detektor-schwankt-mit-der-last` (2×, offen): trifft nur zu, falls ein Fänger nur mit Wahrscheinlichkeit rot wird; ein Treffer wäre der dritte und macht aus der Notiz eine Lücke (eigener Slice).
Unter der Schwelle ohne Bezug: die übrigen Einträge.

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF (Default `*`, Greenfield: Neues Repo ohne Bestandscode; die Spec führt, der Code folgt).
