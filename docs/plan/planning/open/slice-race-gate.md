# Slice slice-race-gate: Race-Detektor im Gate

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-QA-03` (nebenläufige Tests, Isolation), `LH-QA-06` (gleiches Testverhalten, Container, gepinntes Image), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, Stufen des `Dockerfile`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenprobe des neuen Ziels), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) (Build ohne cgo bleibt; der SQLite-Adapter ist reines Go und läuft unter `-race`). Die ADR „Race-Detektor im Gate“ schreibt der Architect beim Start (§4); ihre Kennung vergibt der ADR-Index.

**Berührte Spec-Stellen:** Vertrag des neuen Ziels in `harness/vertraege/` (neue Kennung, Architect vor dem Code) · `SPEC-036` (Stufen `test` und `build` bleiben ohne cgo; gelesen) · `SPEC-102` (Gegenproben) · `SPEC-119`, `SPEC-120`, `SPEC-121` je *Ungeprüft* (dort steht, dass das Gate kein `-race` fährt; Nachzug durch den Architect).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-21.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Eine eigene Stufe des `Dockerfile` baut mit cgo auf einem per Digest gepinnten Go-Image, das den C-Compiler mitbringt, und fährt dort ohne Netz `go test -race` über die nebenläufigen Pakete; ein neues Ziel hängt sie an `make gates`. Die übrigen Stufen bleiben ohne cgo. Damit fängt das Gate ein Datenrennen, das die Tests heute nur im Glücksfall sehen, und `slice-http-adapter` entsteht mit dem Race-Detektor im Gate.

**Übernimmt (Register):** Den Ausgang `geplant` von `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (5×; Entscheidung des Users vom 2026-10-21): diese Closure setzt ihn auf `verkörpert`, mit dem Ziel und der Regel unten als Zielorten.

**Übernimmt (Regel):** Die Regel für `AGENTS.md` §3.10 aus zwei Closure-Notizen, dort je als Vorschlag ohne Zielort. Sie kommt mit diesem Slice (Entscheidungen des Users vom 2026-10-21), mit Herkunfts-Anker `seit welle-record-replay`:

- Aus `slice-clock-idgenerator` (§7, *Steering-Loop-Eintrag*): Wo ein Vertrag eine Zusage über Nebenläufigkeit macht (Sperre, Zähler, Reihenfolge), gilt sie nur als gefangen, wenn die Mutation in einer Wiederholung (`-count=20`) in jedem Lauf rot wird; sonst bekommt sie einen Wirkungs-Fänger, der die Wirkung zählt statt auf ein Datenrennen zu warten, oder steht als akzeptiertes Negativ mit Grund im Plan; der Plan nennt keinen Lauf (`-race`), den das Image nicht fährt.
- Aus `slice-record-service` (§7, *Steering-Loop-Eintrag* (b)), drei Teile: (i) eine Zusage der Form „unter einer Sperre, ohne sie dazwischen freizugeben“ bekommt im Raster je Fenster der Sperre eine Zeile (zwischen welchen Schritten, auf welcher Seite), mit Fänger oder benanntem Negativ; (ii) ein Fänger, der seine Mutation nur mit Wahrscheinlichkeit fängt, nennt seine Rate je Mutation und je Lauf von `go test`, und „hoch“ steht nur mit Zahl; (iii) vor einem Test-Haken oder einem Negativ sucht der Plan einen Haltepunkt an einem Argument, das der Test übergibt (`ctx`, Strom, Fake).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- cgo im Binary und in den Stufen `test`, `build` und `lint` — Bestand: das Binary bleibt reines Go und ohne Toolchain quer baubar ([ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), `SPEC-036`); nur die neue Stufe baut mit cgo. Ob der Fall der Gegenprobe cgo in einer Stufe der Gegenproben braucht, entscheidet der Architect im Vertrag (§6).
- Die Behebung eines Rennens, das das neue Ziel in Code eines Slice in `done/` findet, ohne Entscheidung des Planners — anderer Vorgang: das Rennen ist ein Befund an den Planner (`AGENTS.md` §3.13: ein Slice in `done/` ist keine Adresse), der zwischen Behebung hier und Folge-Slice entscheidet (§6); kein Paket verlässt dafür das Ziel (`AGENTS.md` §3.6), und ein Rot geht nur mit `CO-<NNN>` in die Kette.
- Coverage-Schwelle als Gate — Welle §6; dieser Slice fügt eine Prüfung hinzu, keine Schwelle.
- Das Paket des HTTP-Adapters — `slice-http-adapter` (er entsteht nach diesem Slice und nimmt sein Paket in das Ziel auf, falls der Vertrag eine Liste führt; Sendung dort in §1 und DoD).

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

- [ ] Ziel mit Race-Detektor: die ADR „Race-Detektor im Gate“ ist Accepted und im ADR-Index, der Vertrag des Ziels steht vor dem Code in `harness/vertraege/` (Zusagen, Randformen, Grenzen); eine eigene Stufe des `Dockerfile` auf einem per Digest gepinnten Go-Image mit C-Compiler, `CGO_ENABLED=1`, `RUN --network=none`, fährt `go test -race` über die nebenläufigen Pakete nach dem Vertrag; ein neues Ziel hängt sie an `GATE_CHECKS` und hat seine Zeile in `harness/README.md` §Sensors; die übrigen Stufen bauen weiter ohne cgo (`LH-QA-06`, `LH-QA-03`).
- [ ] Fänger des Ziels: ein Datenrennen in einer Kopie der Eingabe macht das Ziel rot (Gegenprobe, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); jede weitere Zusage des Vertrags hat ihren Fänger (`AGENTS.md` §3.15); die heutigen nebenläufigen Pakete laufen unter `-race` grün oder jedes gefundene Rennen ist als Befund entschieden (§6).
- [ ] Regel und Nachzug: die Regel aus §1 *Übernimmt (Regel)* steht in `AGENTS.md` §3.10 mit `(seit welle-record-replay)`, mit beiden Herkünften: der Satz aus `slice-clock-idgenerator` und die drei Teile aus `slice-record-service` (je Fenster einer Sperre eine Zeile im Raster mit Fänger oder benanntem Negativ; Rate je Mutation und je Lauf von `go test`, „hoch“ nur mit Zahl; vor einem Test-Haken oder Negativ ein Haltepunkt an einem Argument, das der Test übergibt); jede Stelle, die sagt, das Gate fahre kein `-race`, ist nachgezogen — `SPEC-119`, `SPEC-120` und `SPEC-121` *Ungeprüft* (Architect, Grund und Folge je verbleibendem Negativ), die Pläne außerhalb von `done/` (`AGENTS.md` §3.9) und Kommentare in Tests (`AGENTS.md` §3.11).
- [ ] `make gates` grün, `make gegenproben` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Vertrag des Ziels, `harness/README.md` §Sensors (Zeile und *Nicht behauptet*), ADR-Index.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; der Ausgang von `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` wechselt von `geplant` auf `verkörpert` (Zielorte: das Ziel und `AGENTS.md` §3.10).
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `docs/plan/adr/` | neu | ADR „Race-Detektor im Gate“ samt Zeile im Index; schreibt der Architect beim Start, vor dem Code. Sie trägt Entscheidung, Alternativen (Wiederholungen im Gate, Wirkungs-Fänger allein, `-race` außerhalb des Gates) und Gründe, keine Regeldetails (`AGENTS.md` §3.8); ihr `Schärft:` zeigt auf den Vertrag |
| `harness/vertraege/` | neu | Vertrag des Ziels (nächste freie Kennung nach dem Index dort): welche Pakete, welches Image, woran das Ziel rot wird, was es nicht prüft; die Randformen aus §6 entschieden; Zeile im Index des Ordners |
| `Dockerfile` | update | neue Stufe auf einem eigenen, per Digest gepinnten Go-Image derselben Go-Version wie `deps`, mit C-Compiler; die Module kommen aus `deps` (`/go/pkg/mod`), der Lauf ist `RUN --network=none`; die übrigen Stufen bleiben bei `CGO_ENABLED=0` |
| `harness/mk/go-gates.mk` oder ein neues Fragment unter `harness/mk/` | update / neu | das Ziel samt Pin des Images, an `GATE_CHECKS` gehängt; Name des Ziels im Vertrag |
| `harness/README.md` | update | Zeile in §Sensors mit Bindung an die neue ADR; *Nicht behauptet* nachgezogen |
| `tools/gegenproben/` | update | Fall: ein Datenrennen in einer Kopie der Eingabe macht das Ziel rot (Stempel `harness/gegenproben.stempel` folgt) |
| `AGENTS.md` | update | §3.10: die Regel aus §1 *Übernimmt (Regel)* mit beiden Herkünften, samt den drei Teilen aus `slice-record-service`; Anker `(seit welle-record-replay)` |
| `spec/spezifikation.md` | update | `SPEC-119`, `SPEC-120`, `SPEC-121` *Ungeprüft*: die Sätze über das fehlende `-race`, nachgezogen vom Architect |
| Pläne außerhalb von `done/`, Kommentare in Tests | update | jede Stelle, die sagt, das Gate fahre kein `-race` (§6, Risiko *Nachzug*) |

- Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster in §6 (`AGENTS.md` §3.16) schreibt der Architect mit dem Vertrag vor dem ersten Code-Commit (`AGENTS.md` §3.12). Erwartete Zusagen, als Suchliste und nicht als Vertrag: die Stufe baut mit cgo auf dem gepinnten Image; sie läuft ohne Netz; sie fährt `-race` über genau die Pakete des Vertrags; ein Datenrennen macht sie rot; die übrigen Stufen bauen ohne cgo; das Ziel hängt an `make gates`.
- Wer die nebenläufigen Pakete heute sind, steht im Vertrag; Kandidaten nach Goroutinen und Sperren im Code: `src/hexagon/services/record/`, `src/hexagon/services/record/capture/`, `src/adapters/driven/httpupstream/`, `src/adapters/driven/clock/`, `src/adapters/driven/sqlite/` und die Vertragstests unter `src/hexagon/ports/`.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): Kein Slice der Welle muss vorher in `done/` liegen (der Slice ändert `Dockerfile`, Gate-Kette und Gegenproben, nicht den Code der Pakete); er startet, sobald die WIP-Grenze frei ist, und liegt vor `slice-http-adapter`, dessen Start ihn in `done/` verlangt. Vor dem ersten Code-Commit sind die ADR „Race-Detektor im Gate“ Accepted und der Vertrag samt Randformen geschrieben, beides vom Architect (`AGENTS.md` §3.12).

**Größenmarke** (`AGENTS.md` §3.14): über **600** Zeilen nach dem Review (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `Dockerfile`, `harness/mk/`, `tools/gegenproben/` und die Dateien unter `src/`, die der Slice ändert, Tests eingeschlossen; `<Basis>` ist der Commit, der ADR und Vertrag vor dem Code abschließt; Dokumente zählen nicht mit).

**Schätzung** (`AGENTS.md` §3.14), vorläufig, ohne Tabelle: rund 6 Zeilen × 85 = **510** für den Stand beim Handoff, ohne Port-Fake und Test-Server. Die Zeilen zählen nebenläufig, weil ihre Fänger nebenläufig sind: die Gegenprobe baut selbst ein Datenrennen (Entscheidung des Users vom 2026-10-21; §3.14: nebenläufig zählt eine Zeile, wenn der Code oder sein Fänger nebenläufig ist). Gegen die Marke steht die Schätzung × 1,25 = 637,5, rund **640**, 6 % über der Marke; die Marke ist gegen 510 gesetzt, vor der Form der Schätzung vom 2026-10-21 (§6, Risiko *Marke gegen die Schätzung × 1,25*). Steht die Tabelle, rechnet der Planner die Schätzung vor dem ersten Code-Commit neu; liegt sie × 1,25 dann über der Marke, wird vor dem Code geschnitten, nicht nach dem Code. Nach dem Code wird gemessen und mit Grund berichtet.

**Schnitt beim Handoff** (`AGENTS.md` §3.14): keiner vorab benannt. Stufe, Ziel und Gegenprobe hängen am selben Vertrag, und die Gegenprobe ist der Fänger der Zusage „ein Datenrennen macht das Ziel rot“ (`AGENTS.md` §3.15); ein Schnitt dort ließe die Zusage ohne Fänger. Nennt der Vertrag einen Teil, der zuletzt in eigenem Commit entstehen kann und keinen Mechanismus mit dem Rest teilt, trägt der Planner ihn vor dem ersten Code-Commit hier ein.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke (sie schneidet vor dem Code, und beim Handoff ist kein Teil vorab benannt, oben, `AGENTS.md` §3.14); offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten), etwa wenn das Ziel in heutigem Code Rennen findet, deren Behebung eigene Liefer-Punkte wird (§6).
- `in-progress` → `open` (blockiert): kein gepinntes Go-Image derselben Version mit C-Compiler, das ohne Netz baut, oder der Race-Detektor läuft auf der Plattform des Bau-Hosts nicht (§6); dann entscheidet der Architect in der ADR, nicht der Implementer im `Dockerfile`.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` fährt das neue Ziel und ist grün, `make gegenproben` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Rennen im heutigen Code: das Ziel kann in Paketen aus Slices in `done/` (`capture`, `record`, `httpupstream`, `clock`, `sqlite`) ein Rennen finden, das die Tests bisher nur im Glücksfall sahen — **Ausgang:** offen; jedes Rennen ist ein Befund an den Planner (`AGENTS.md` §3.13), der zwischen Behebung in diesem Slice und einem Folge-Slice entscheidet; das Ziel geht nicht mit einem bekannten Rot in die Kette, außer mit `CO-<NNN>` (Modul 7).
- Laufzeit unter `-race`: die Läufe werden langsamer und brauchen mehr Speicher; Tests mit echten Zeitgrenzen (Frist von 20 ms und Zeitgrenze von 10 s in A8 von `slice-upstream-adapter`, die Fenster von 250 ms aus `slice-capture-pipeline`) können unter `-race` kippen, ohne dass ein Rennen vorliegt — **Ausgang:** offen; der Implementer misst den Faktor im Image, ein Test, der nur unter `-race` ohne Rennen rot wird, ist ein Befund an den Planner, keine Lockerung der Zeitgrenze im Gate.
- Kosten: ein längerer Lauf von `make gates` und ein weiteres gepinntes Image (vom User beim Annehmen benannt) — **Ausgang:** offen; die Closure nennt die Dauer von `make gates` vor und nach dem Slice.
- Image: das Eingangs-Image von `deps` ist ein Alpine-Image ohne C-Compiler, und ein Paketmanager im Bau bräuchte Netz; die Stufe braucht ein eigenes Go-Image derselben Version mit C-Compiler, per Digest gepinnt — **Ausgang:** offen; der Architect wählt es in der ADR.
- Plattform: `deps` und die davon abgeleiteten Stufen laufen auf der Plattform des Bau-Hosts, und den Race-Detektor gibt es nicht auf jeder Plattform — **Ausgang:** offen; der Vertrag nennt, auf welchen Plattformen das Ziel läuft und was es auf den übrigen tut.
- Gegenprobe: das Programm der Gegenproben läuft heute in einer Stufe auf `LINT_IMAGE` ohne cgo; ein Fall mit `-race` braucht cgo dort oder eine eigene Stufe — **Ausgang:** offen; der Architect entscheidet im Vertrag, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) bleibt.
- Pakete nach diesem Slice: `slice-session-wartung`, `slice-baseline-import`, `slice-workspace-use-case`, `slice-replay-service`, `slice-http-adapter` und `slice-bootstrap-m2` bringen nebenläufige Pakete — **Ausgang:** offen; der Vertrag entscheidet, ob das Ziel eine Liste führt oder alle Pakete fährt. Die Sendung an `slice-http-adapter` steht dort in §1 und DoD (mit der Kennung dieses Slice); für die übrigen zieht die Closure dieses Slice deren Pläne nach (`AGENTS.md` §3.9, §3.13 beim Nachzug geprüft).
- Nachzug an jeder Stelle (`BEO-REPO/berichtigung-erreicht-nicht-jede-stelle`, 2×): der Satz „das Gate fährt kein `-race`“ steht in `SPEC-119`, `SPEC-120` und `SPEC-121` *Ungeprüft*, in §6 oder §8 der Pläne von `slice-session-wartung`, `slice-baseline-import`, `slice-workspace-use-case`, `slice-replay-service`, `slice-http-adapter` und `slice-bootstrap-m2`, in §3 von `slice-upstream-adapter` und im Kopfkommentar von A15 in `src/adapters/driven/httpupstream/client_test.go` — **Ausgang:** offen; die Closure sucht vor dem Abschluss in `spec/`, `src/`, `harness/` und den Plänen außerhalb von `done/` nach `-race`, „Race-Detektor“ und `CGO_ENABLED` und nennt jede Fundstelle mit ihrem Nachzug.
- Teile aus der Closure-Notiz von `slice-record-service` §7 (b) (je Fenster einer Sperre eine Zeile im Raster, Rate je Mutation und je Lauf, Haltepunkt an einem Argument des Tests) — **Ausgang:** entfallen; vom User am 2026-10-21 angenommen, sie stehen in §1 *Übernimmt (Regel)* und in der DoD und gehen mit diesem Slice in `AGENTS.md` §3.10.
- Klasse der Zeilen eines Werkzeug-Vertrags — **Ausgang:** entfallen; × 85, weil die Gegenproben dieses Slice selbst Rennen bauen (Entscheidung des Users vom 2026-10-21); allgemein in `AGENTS.md` §3.14: nebenläufig zählt eine Zeile, wenn der Code oder sein Fänger nebenläufig ist. §4 rechnet damit.
- Marke gegen die Schätzung × 1,25: nach `AGENTS.md` §3.14 schätzt die Formel den Stand beim Handoff, und gegen die Marke steht die Schätzung × 1,25; die vorläufige Schätzung ergibt so rund 640 gegen die Marke 600, die gegen 510 und vor dieser Form gesetzt ist; einen Teil, den ein Schnitt vor dem Code nähme, nennt der Plan nicht (§4) — **Ausgang:** offen; entschieden wird mit der Neuberechnung aus der Tabelle vor dem ersten Code-Commit. Ob eine Marke, die vor der Form vom 2026-10-21 gesetzt ist, dabei neu gesetzt werden darf oder gilt, sagt §3.14 nicht; Frage an den User (sie betrifft ebenso die Pläne im Auftrag der Welle, [welle-record-replay](../welle-record-replay.md) §4).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* als Gerüst; jede Zelle nennt die Frage, der Ort ist der Vertrag, den der Architect vor dem ersten Code-Commit schreibt; eine Zelle ohne Ort hält den Implementer an — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Ziel fahren (Pakete) | ein Paket des Vertrags fehlt im Baum: rot oder grün? (Vertrag) | Paket ohne Test oder Liste leer: rot oder grün? (Vertrag) | ein Paket ohne Nebenläufigkeit, ein Paket außerhalb von `src/`, eines der Werkzeuge unter `tools/`: gehört es dazu? (Vertrag) | Zeitgrenze von `go test` unter `-race`, Abbruch des Laufs: Zeitgrenze und Meldung (Vertrag) | zweiter Lauf ohne Änderung: Schicht- und Test-Cache, läuft `-race` erneut? `-count`? (Vertrag) |
| Stufe bauen (Image) | Pin oder Digest fehlt: Bau scheitert? (Vertrag) | gilt nicht: das Image ist ein Pflichtargument des Ziels, ohne Wert gilt der Pin der Datei — zu bestätigen im Vertrag | Plattform ohne Race-Detektor (Vertrag, Risiko *Plattform*) | ein Modul fehlt in `deps`, der Lauf bräuchte Netz: rot mit welcher Meldung? (Vertrag) | gleicher Digest, gleiche Module: gleiches Ergebnis (`LH-QA-06`; Vertrag) |
| Gegenprobe des Ziels | der Fall mit Rennen fehlt: der Stempel deckt ihn nicht (Vertrag, `SPEC-102`) | gilt nicht: der Fall ist fest im Programm | ein Rennen nur im Testcode, ein Rennen in einem Paket außerhalb der Liste: rot? (Vertrag) | Abbruch der Gegenprobe: wie die übrigen Fälle (`SPEC-102`, zu bestätigen) | Rennen, das der Detektor nur in einem Teil der Läufe sieht: Rate des Falls (Vertrag; `AGENTS.md` §3.10) |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Ursachen am selben Lauf:

| Paar | Entscheidung | Fänger |
|---|---|---|
| Rennen und Zeitgrenze im selben Test | offen: welche Meldung gewinnt (Vertrag) | offen (Vertrag) |
| Rennen in zwei Paketen | offen: meldet das Ziel beide (Vertrag) | offen (Vertrag) |
| Rennen und Fehler eines Tests ohne Rennen | offen: beide gemeldet, Exit-Code (Vertrag) | offen (Vertrag) |

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

*Wird bei der Closure gefüllt.* Mit diesem Slice wird verkörpert (die Regel in `AGENTS.md` §3.10 und das neue Ziel; Auslöser `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen`); der Steering-Loop-Eintrag trägt dann das Feld `liegt in`, und der Zielort trägt den Anker `seit welle-record-replay`.

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle Einträge). Gegenstand dieses Slice ist `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (5×, Ausgang `geplant` mit dieser Kennung). Für die Planung zählen außerdem: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (2×) — der Nachzug des Satzes „das Gate fährt kein `-race`“ trägt die Liste der Stellen und die Suche vor der Closure (§6, Risiko *Nachzug an jeder Stelle*); `BEO-REPO/akzeptiertes-negativ-ohne-folge` (2×) — was nach dem Nachzug in `SPEC-119` bis `SPEC-121` *Ungeprüft* bleibt (Rennen außerhalb der Pakete des Ziels, Raten unter `-race`), steht dort mit Grund und Folge; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) — der Vertrag des Ziels nennt nur, was eine Gegenprobe fängt; `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (verkörpert in `AGENTS.md` §3.16) — „alle nebenläufigen Pakete“ ist eine Allaussage und braucht ihren Fänger über alle Pakete; `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (verkörpert in `AGENTS.md` §3.14) — Marke und Schätzung in §4. Ohne Bezug: `BEO-REPO/befund-an-rolle-ohne-artefakt` (1×) und `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×; der Fall der Gegenprobe führt keine Zeichen über ASCII).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
