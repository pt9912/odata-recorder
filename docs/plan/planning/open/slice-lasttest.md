# Slice slice-lasttest: Lasttest nach `LH-QA-01.a` mit Messpunkt am Gesamtspan für Replay und Mock — Lauf mit Takt, Prozess des Binary, Treiber und Gegenproben (Teil 3 des Schnitts; trägt die Kennung des ungeteilten Slice)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0003](../../adr/0003-lint-profil-golangci.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)

**Berührte Spec-Stellen** (Teil 3 des Schnitts; die Stellen der Teile 1 und 2 stehen in [slice-lasttest-fixtures](../in-progress/slice-lasttest-fixtures.md) und [slice-lasttest-auswertung](slice-lasttest-auswertung.md)): `SPEC-216` ([der Vertrag des Werkzeugs](../../../../harness/vertraege/SPEC-216-lasttest.md), Architect, 2026-11-23: Nr. 1 bis 4, 8, 9, 14 e, 14 g (Architect, 2026-11-24: das Hauptprogramm und die Regel für `forbidigo`), 15, 16 a, 16 d, 16 e, 16 g, 16 i, 16 k) · `LH-QA-01.a` (Messaufbau: 2 vCPU, 4 GiB, 10.000 Exchanges, Bodies bis 64 KiB, 20 Requests/s, 10 Minuten nach Warm-up; Mockserver mit 10.000 Entities) · `SPEC-135` Nr. 8 und `SPEC-195` (der Aufwand je Request, den der Lasttest misst, *Ungeprüft* dort) · `SPEC-139` Nr. 13 f (die Zeilen sind der Weg des Lasttests) · `SPEC-037` Nr. 6 (der Übergang, dem der Aufbau folgt) · `SPEC-102` (Zählung der Fälle der Gegenproben) · spezifikation.md §6 (Abnahmeparameter: „der Messaufbau muss reproduzierbar im Test-Harness hinterlegt sein“) · `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30 (Schnitt vor dem Code in drei Teile durch den Planner am 2026-11-23, Vorschlag des Architect; der Plan des ungeteilten Slice trug Ziel, Tabelle und Raster in einer Datei).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein reproduzierbarer Lasttest nach `LH-QA-01.a` fährt das gebaute Binary mit 2 vCPU und 4 GiB gegen 10.000 aufgezeichnete Exchanges (Replay) und gegen einen Workspace aus 10.000 Entities (Mock), mit 20 Requests/s über 10 Minuten nach Warm-up, und gibt p95 des Gesamtspans `odata.gateway.request` für Replay und Mock aus, mit dem Anteil von `odata.replay.match` bzw. `odata.mock.query`; die Zahlen stehen in der Closure-Notiz samt Last und Messendem des Laufs. Er prüft gegen den Zielwert des Lastenhefts (vorläufig p95 ≤ 100 ms) und berichtet; ein Durchfallen ist ein Carveout mit Folge-Slice, kein stilles Rot.
Die Form ist entschieden (Architect, 2026-11-23, [SPEC-216](../../../../harness/vertraege/SPEC-216-lasttest.md)): ein **Werkzeug außerhalb der Gate-Kette** (`make lasttest`, kein Eintrag in `GATE_CHECKS`), Programm `tests/lasttest/` im Hauptmodul, Treiber `tools/harness/lasttest.sh`, Stufe `lasttest` des `Dockerfile`; zwei Läufe je Betriebsart im selben Container (`--cpus 2 --memory 4g`, netzlos, ohne Mount); der Zielwert ist der Parameter `LASTTEST_ZIEL_MS` (Vorgabe 100), und nur ein Lauf des Profils `abnahme` mit Ziel 100 und beiden Betriebsarten trägt `abnahme=ja`. Erwartung, kein Befund: `SPEC-135` Nr. 8 liest je Request die ganze Session, `SPEC-195` je Collection alle Einträge des Sets; ob `LH-QA-01` hält, ist das Ergebnis dieses Slice, und ein Durchfallen ist der Normalfall der Planung (§6).
**Dieser Teil** liefert den Lauf: den offenen Takt und den Lastgenerator (Zeilen L1 bis L4), den Prozess des Binary und den Ablauf mit Abbruch (L5 bis L7), den Treiber mit den Variablen, den Grenzen von `docker run` und den Exits (L8 bis L10), die Stufe, das Fragment und die Zeile in `harness/README.md` (L11), das Hauptprogramm `main.go` mit der Regel für `forbidigo` in `.golangci.yml` (die Datei greift auf `os.Stdout` zu; die Regel gilt dem Pfad `^tests/lasttest/main\.go$` allein, `AGENTS.md` §3.2, `SPEC-216` Nr. 14 g; Fänger: `make lint`, ohne die Regel meldet es einen Befund) und den Handlauf mit den Zahlen. **Schnitt vor dem Code** (Vorschlag des Architect, Planner, 2026-11-23): die Schätzung des ungeschnittenen Plans (*S* = 6055, Marke 8200, 34 Zeilen) liegt weit über der Schnittgrenze von `SPEC-128` Nr. 1; drei Teile, die nacheinander laufen und je einzeln lieferbar sind: **1** [slice-lasttest-fixtures](../in-progress/slice-lasttest-fixtures.md) (Profile, Referenzdienst, Aufbau, Listen, Umgebung, Uhr; F1 bis F10), **2** [slice-lasttest-auswertung](slice-lasttest-auswertung.md) (Leser, Phasen, p95, Last-Proben, Urteil, Ausgabe; A1 bis A13) und **3** (diese Datei, L1 bis L11). Dieser Teil trägt die Kennung des ungeteilten Slice (Abweichung vom Namensvorschlag `slice-lasttest-lauf`, Planner, 2026-11-23): jeder Verweis „`slice-lasttest` in `done/`“ in den Plänen davor und danach (`welle-betrieb`, die Roadmap, die Pläne mit Bezug auf den Lasttest) bedeutet damit „alle drei Teile in `done/`“, weil dieser Teil als letzter startet und die zwei anderen voraussetzt.

**Übernimmt:** keine Sendung; die Sendung von `slice-telemetry-spans` (Abrufweg des Gesamtspans über die Span-Zeile, `SPEC-231` Nr. 6) liegt bei Teil 2 ([slice-lasttest-auswertung](slice-lasttest-auswertung.md), Leser), dieser Teil füttert den Leser mit dem Strom des Binary und belegt die Zeile am gebauten Binary im Handlauf (Zeile L11).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Profile, Referenzdienst, Aufbau, Listen, Umgebung und Uhr — [slice-lasttest-fixtures](../in-progress/slice-lasttest-fixtures.md) (Teil 1, liegt bei diesem Start in `done/`); der Leser, die Phasen, p95, Proben, Urteil, Exit und Ausgabe — [slice-lasttest-auswertung](slice-lasttest-auswertung.md) (Teil 2, liegt bei diesem Start in `done/`). Dieser Teil ruft ihre Schnittstellen.
- Die Bestätigung des Zielwerts von `LH-QA-01` — Entscheidung des fachlichen Owners, nicht des Repos (Teil von M7, [welle-betrieb](../welle-betrieb.md) §3 und §6); der Slice misst und berichtet gegen den vorläufigen Wert des Lastenhefts, p95 ≤ 100 ms (ausdrückliche Entscheidung des Users vom 2026-11-23, siehe §4).
- Tuning bei verfehltem Zielwert (Indizes, Puffer) — eigener Slice nach der Messung; fällt der Wert durch, ist es ein Carveout mit Folge-Slice (`harness/conventions.md`, Carveout-Disziplin), kein stilles Rot.
- Latenz im Record- und im Pass-through-Betrieb — `LH-QA-01` sagt dort keine Latenz zu (der Upstream bestimmt sie); ebenso der Hybrid-Fallback auf den Upstream.
- Der Einbau der Spans — `slice-telemetry-einbau-gateway` (Gesamtspan), `slice-telemetry-einbau-replay` (Match-Span), `slice-telemetry-einbau-mock` (Query-Span) und `slice-telemetry-einbau-mock-mutation` (Mutations-Span); dieser Slice liest sie.
- Eine Admin-Schnittstelle zum Aufbau der Last (Import, Workspace) — `welle-admin-api`; die Fixtures entstehen im Programm von Teil 1 über `bootstrap.Build` im Modus `record`, `Import().Import` und `Workspaces().Create` (`SPEC-216` Nr. 6 a), wie der Übergang von `SPEC-037` Nr. 6.
- Ein Rauchlauf oder der volle Lauf in `make gates` — entschieden (`SPEC-216` Nr. 1): das Ziel ist ein Werkzeug; das Profil `rauch` ist ein Handlauf der Rollen. Aufnahme in die Kette wäre eine Verschärfung und ein eigener Slice nach dem ersten Lauf (Zeitbudget `SPEC-211` Nr. 4).
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

- [ ] Lauf: das gebaute Binary mit 2 vCPU und 4 GiB, 20 Requests/s, 60 s Warm-up und 10 Minuten Messphase, zwei Läufe je Betriebsart, mit offenem Takt, Prozess und Ablauf mit Abbruch; der Aufbau und die Auswertung kommen aus den Teilen 1 und 2; Ausgabe von p95 des Gesamtspans für Replay und Mock samt Anteil des Match- bzw. Query-Spans, Zahl der Requests, Fehlerzahl, Last des Hosts (Proben) und Zielwert; ungültig ist ein Lauf bei Fehlern, falscher Zählung, Last über der halben Kernzahl und zwei Läufen über der Toleranz (der größere von 10 % des kleineren p95 und 2 ms, `SPEC-216` Nr. 12 c); die Rolle, die ihn fährt, trägt Last und Messenden in die Closure-Notiz (Zeilen L1 bis L7; `LH-QA-01`, Messmethode).
- [ ] Werkzeug und Verdrahtung: der Treiber `tools/harness/lasttest.sh` mit den Variablen und den Grenzen von `docker run`, die Stufe `lasttest` des `Dockerfile`, das Fragment mit dem Ziel `lasttest` ohne Eintrag in `GATE_CHECKS`, die Fälle der Gegenprobe und die Zeile des Ziels in der Tabelle der **Werkzeuge** von `harness/README.md` samt *Nicht behauptet* (Zeilen L8 bis L11; `SPEC-216` Nr. 1, 2, 3, 15); die Regel für `forbidigo` zum Hauptprogramm in `.golangci.yml` (`SPEC-216` Nr. 14 g) gehört zum Hauptprogramm und zählt nicht als eigener Punkt.
- [ ] Bericht und Entscheidung: die Zahlen stehen in der Closure-Notiz; Vergleich gegen den Zielwert; bei Überschreitung ein Carveout nach `docs/plan/carveouts/` mit Frist, Auflösungs-Trigger und Folge-Slice (`harness/conventions.md`), nie ein stilles Rot.
- [ ] `make gates` grün; der Lasttest ist ein Werkzeug außerhalb der Kette, die Kette prüft das Paket mit `make test`, `make test-race`, `make a-check` und `make lint` (`SPEC-216` Nr. 1 und 14).
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`harness/README.md` §Sensors).
- [ ] Führt der Slice ein sichtbares Verhalten des Binary ein (Schlüssel, Code, Kommando): Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Lasttest ist ein Werkzeug außerhalb des Binary).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tests/lasttest/` | update | der Lauf im Paket von Teil 1 und 2 (Schicht `tests`, [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)): Takt, Lastgenerator, Prozess des Binary, Ablauf und Abbruch, `main.go` mit der Funktion `main` (das Paket heißt von Teil 1 an `main` und trägt bis dahin ohne `main.go`, `SPEC-216` Nr. 14 g); im Hauptmodul statt unter `tools/`, weil das Paket den Composition Root und den Adapter `sqlite` braucht und `make test`, `make test-race` und `make a-check` es prüfen (`SPEC-216` Nr. 14 e) |
| `Dockerfile`, `harness/mk/lasttest.mk` | update, neu | Stufe `lasttest` (Programm und Binary der Stufe `build`, netzlos außer `deps`) und das Ziel `lasttest` ohne Eintrag in `GATE_CHECKS`; `Makefile` bleibt unverändert (`include harness/mk/*.mk`) <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) --> |
| `.golangci.yml` | update | eine Regel unter `linters.exclusions.rules` für `forbidigo`, Pfad `^tests/lasttest/main\.go$`, Text `os\.Std(out\|err)`, mit Kommentarblock, dessen erste Zeile mit `# Why:` beginnt (Muster der Regel zu `tests/e2e/uebergang/main.go`, `AGENTS.md` §3.2, [ADR-0003](../../adr/0003-lint-profil-golangci.md)): das Hauptprogramm greift auf `os.Stdout` zu, und `forbidigo` meldet es (`SPEC-216` Nr. 14 g, gemessen: 1 Befund; mit der Regel 0, `make a-check` unverändert grün); die Regel entsteht im Commit mit `main.go` und ändert kein anderes Paket; sie gilt dieser einen Datei, die Schärfe von `forbidigo` für den Rest des Baums bleibt |
| `tools/harness/lasttest.sh` | neu | Treiber auf dem Host: Variablen, `docker info`, Bau, `docker run` mit `--cpus 2 --memory 4g --memory-swap 4g --network none` (`SPEC-216` Nr. 3, `AGENTS.md` §3.1) |
| `harness/vertraege/` | geliefert (Architect, 2026-11-23) | `SPEC-216` und die Zeile in der Liste der Verträge ([ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)); dieser Plan schreibt sie nicht, `SPEC-102` zieht der Implementer in diesem Teil mit den Gegenproben um die Zählung nach |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` | update | Fälle der Gegenprobe zum Treiber mit Ersatz für `docker` (`SPEC-216` Nr. 15, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); der Stempel erst im Integrationsschritt |
| `harness/README.md` | update | Zeile des Ziels in der Tabelle der **Werkzeuge** (nicht der Gates), *Nicht behauptet* (der Lasttest zu `LH-QA-01`: die Zahlen des Produkts, die Verdrahtung, `SPEC-216` Nr. 16) |
| `docs/plan/carveouts/` | neu, falls der Wert durchfällt | Carveout mit Frist und Folge-Slice |
| Tests | neu | Fälle der Zusagen L1 bis L11, Mengen als Schleifen, Ersatz-Dienst und Ersatz für `docker` (`SPEC-216` Nr. 14) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstanden mit dem Vertrag **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Entscheidungen des Architect** (2026-11-23; die Aufträge 1 bis 7 des ungeschnittenen Plans mit ihrem Ort in [SPEC-216](../../../../harness/vertraege/SPEC-216-lasttest.md), soweit dieser Teil sie trägt; der Vertrag trägt die Gründe und die verworfenen Alternativen, dieser Plan wiederholt sie nicht, `AGENTS.md` §3.8):

1. **Gate oder Werkzeug** (`SPEC-216` Nr. 1, 14): ein Werkzeug außerhalb der Kette (`make lasttest`, Zeile unter den Werkzeugen); die Kette prüft das Paket `tests/lasttest` mit `make test`, `make test-race`, `make a-check`, `make lint` und `make groessenmarke`, und die Auswertung rechnet gegen Literale (p95 aus bekannten Werten); der Lauf mit dem echten Binary ist Handlauf (Profil `rauch`). Ein Rauchlauf in der Kette ist verworfen; die Aufnahme später ist eine Verschärfung.
2. **Messung** (Nr. 8 d, 9 e): die Messung sendet nur die Anfragen der Liste und entleert stdout und stderr ohne Unterbrechung; die Auswertung der Zeilen (Nr. 10, 12) liegt in Teil 2.
3. **Reproduzierbarkeit** (Nr. 3, 5): `--cpus 2 --memory 4g --memory-swap 4g --network none`, zwei Läufe je Betriebsart; das Werkzeug setzt keine Sperre, die Rolle nimmt den `flock` der Betriebsregeln; Startwert und Permutationen liegen in Teil 1, Toleranz und Proben der Last in Teil 2.
4. **Last** (Nr. 8): offene Last mit festem Takt (Anfrage *k* zur Zeit *t0* + *k*/20 s), keine Obergrenze der Verbindungen; die Listen liegen in Teil 1.
5. **Zielwert** (Nr. 2): `LASTTEST_ZIEL_MS` (Vorgabe 100, 1 bis 9999) als Variable des Treibers; der Vergleich und `abnahme=ja` liegen in Teil 2.
6. **Tabelle, Raster, Gegenproben, Produktcode-Größe:** unten und in §6; die Größe in §4.
7. **Hauptprogramm und Lint** (`SPEC-216` Nr. 14 g, Architect, 2026-11-24, gemessen): das Paket trägt in Teil 1 und 2 ohne `main.go`; `main.go` kommt mit diesem Teil, braucht `os.Stdout`, und die Regel für `forbidigo` in `.golangci.yml` (Pfad `^tests/lasttest/main\.go$`, `# Why:`) macht `make lint` grün; `make a-check` blieb grün. Der Teil legt `main.go` und die Regel im selben Commit an; nicht gemessen sind `make test` und `make test-race` mit `main.go` und die Stufe `lasttest` selbst (Zeile L11 und der Handlauf, §6).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-23, Zeilen L1 bis L11 der Tabelle des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; jede Zeile steht in genau einem der drei Teile; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). `Nr.` ist die Kennung des ungeschnittenen Plans: **F** gehört zu Teil 1 (Fixtures, Aufbau, Uhr), **A** zu Teil 2 (Leser und Auswertung), **L** zu Teil 3 (Lauf: Prozess, Last, Treiber). *Ersatz-Dienst*: ein HTTP-Server im Prozess des Tests und ein Hilfsprozess, die je Anfrage eine Zeile des Spans mit vorgegebener Dauer schreiben und zählen; *Fake-Uhr* ersetzt die Schnittstelle von Zeit (`SPEC-216` Nr. 14). Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest; `Nr.` bezieht sich auf `SPEC-216`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| L1 | (Nr. 8 a) Offene Last: Anfrage *k* startet zur Zeit *t0* + *k* × Takt aus *k*, unabhängig von Antworten; Warm-up und Messphase ohne Lücke; Zahl der Anfragen gleich Rate × (Warm-up + Messphase) | `TestTakt` (Fake-Uhr, Ersatz-Dienst): Startzeiten der Anfragen 0 bis 4 und der letzten (*k* = 13199) als Literal (0, 50, 100, 150, 200 ms; 659,95 s); die Anfrage 3 hängt 200 ms, die Starts 4 bis 7 folgen im Takt; Zahl der Anfragen am Ersatz 13200 und 260; Drift: ein Overhead von 1 ms je Start der Fake-Uhr verschiebt den Start 13199 nicht; Rate je Prozess unter `make test-race` | der Takt wartet auf die Antwort; der Takt aus dem Ende des vorigen Starts (Drift); Anfrage 13199 fehlt | nebenläufig |
| L2 | (Nr. 8 b) Verbindungen ohne Obergrenze, eine Goroutine je Anfrage; `gleichzeitig_max` und `verspaetung_max_ms` | Wirkungs-Fänger `TestGleichzeitig`: der Ersatz hält 5 Anfragen offen: `gleichzeitig_max` gleich 5; die Fake-Uhr springt 7 ms nach der Fälligkeit einer Anfrage: `verspaetung_max_ms` gleich 7; 20 Läufe, Rate je Prozess | `MaxConnsPerHost` gleich 1; ein Zähler ohne atomare Operation (Detektor); die Verspätung als Mittel | nebenläufig |
| L3 | (Nr. 8 c) Der Body wird vollständig gelesen und verworfen; Frist 10 s; jede Anfrage endet genau einmal als Erfolg (200) oder als Fehler | `TestAntwort`: Body 64 KiB: `bytes` gleich der Summe, die Verbindung wird wiederverwendet (Zahl der neuen Verbindungen höchstens `gleichzeitig_max`); 7 Fehlerformen (404, 500, 503, 204, Transportfehler, Zeitüberschreitung 10 s + 1 der Fake-Uhr, Lesefehler des Bodys) je genau einmal als Fehler gezählt, 1 Gutfall; Zählprüfung 8 | der Body wird nicht gelesen; 204 gilt als Erfolg; die Zeitüberschreitung zählt nicht; ein Fehler zählt doppelt | andere |
| L4 | (Nr. 8 d) Die Messung sendet nur die Anfragen der Liste: keine Anfrage zur Bereitschaft, kein Admin-Aufruf, keine Wiederholung eines Fehlers | `TestNurDieListe`: der Ersatz zählt jede Anfrage nach Methode und Pfad: gleich der Liste (Mock und Replay, `rauch`), sonst 0 (`/healthz`, `/__admin`); ein Status 503 wird nicht wiederholt; bis zur ersten fälligen Anfrage 0 Anfragen und eine Verbindung (die Bereitschaft); Zählprüfung 2 Profile | die Bereitschaft per `GET /`; eine Wiederholung nach einem Fehler | andere |
| L5 | (Nr. 9 b, c, d) Prozess des Binary: bereit bei TCP-Verbindung binnen 10 s, sonst `binary`; endet das Binary früh, ist der Lauf ungültig; Ende mit `SIGTERM`, nach 15 s `SIGKILL`; Exit ≠ 0 und `SIGKILL` sind ein `Hinweis`, die Gültigkeit bleibt | `TestProzess` (Hilfsprozess, Fake-Uhr): 6 Formen: startet nicht (Exit 1), lauscht nie (10 s + 1), stirbt im Lauf (Exit 3), endet auf `SIGTERM` mit 0, mit 1 (Hinweis `binary_exit`), ignoriert `SIGTERM` (nach 15 s `SIGKILL`, Hinweis `binary_kill`); Orakel Literal je Form; 20 Läufe, Rate je Prozess; Zählprüfung 6 | `SIGKILL` sofort; die Bereitschaft per HTTP-Anfrage; Exit ≠ 0 macht ungültig | nebenläufig |
| L6 | (Nr. 9 e, f) stdout und stderr werden von der Bereitschaft bis zum Ende entleert; je Lauf ein frischer Prozess und neue Replay-Sessions | Wirkungs-Fänger `TestStroeme`: der Hilfsprozess schreibt je 8 MiB auf stdout und stderr (Puffer der Pipe 64 KiB): der Lauf endet binnen einer Sekunde; Lauf 1 und Lauf 2 mit verschiedenen Prozess-Kennungen und Session-IDs; 20 Läufe, Rate je Prozess | nur stderr entleert; der Leser beginnt nach der Bereitschaft; der Prozess wird für Lauf 2 wiederverwendet | nebenläufig |
| L7 | (Nr. 4 b, d) Ablauf und Abbruch: Probe, Aufbau, `replay`, `mock`, Urteile; INT oder TERM beenden das Binary, kein Urteil, Exit 2; der Aufbau endet vor dem ersten Binary | `TestAblauf` (Ersatz für Aufbau und Binary): die Folge der Aufrufe als Literal für `beide`, `replay`, `mock` und `rauch` (4); Abbruch in 4 Phasen (Aufbau, Warm-up, Messphase, Nachlauf): `SIGTERM` am Binary einmal (bzw. 0 im Aufbau), keine Urteilszeile, Exit 2; 20 Läufe, Rate je Prozess; Zählprüfung 4 + 4 | der Abbruch lässt das Binary laufen; ein Urteil nach dem Abbruch; der Aufbau läuft neben dem ersten Binary | nebenläufig |
| L8 | (Nr. 2) Die Variablen des Treibers: je Variable ungesetzt (Vorgabe eingesetzt), gültig an den Grenzen, leer, null (`LASTTEST_ZIEL_MS`), Form | Fälle der Gegenprobe mit Ersatz für `docker`: `LASTTEST_BETRIEBSART` (ungesetzt, `replay`, `mock`, `beide`, leer, `alle`, `Replay`) 7; `LASTTEST_PROFIL` (ungesetzt, `abnahme`, `rauch`, leer, `voll`) 5; `LASTTEST_ZIEL_MS` (ungesetzt, `1`, `9999`, leer, `0`, `00`, `10000`, `01`, `1.5`, `-1`, `1 `) 11; `LASTTEST_STARTWERT` (ungesetzt, `0`, `1`, 18 Ziffern, 19 Ziffern, leer, `01`, `-1`) 8; Orakel: Exit 2 mit der Meldung oder `docker run` mit dem eingesetzten Wert; Zählprüfung 31 | die Prüfung auf `null` entfällt; `{0,3}` zu `{0,4}`; die Vorgabe wird nicht eingesetzt; `01` als Startwert gültig | andere |
| L9 | (Nr. 3 b, c) Pflichtoptionen und verbotene Formen von `docker run`, als Liste gelesen | Fälle der Gegenprobe: Pflichtoptionen 5 und `--env` 5 (4 Variablen und `LASTTEST_KERNE`), je ein Fall der Mutation (entfernt) gleich 10; verbotene Formen 18 (Nr. 3 c: 14 Optionen und 4 Optionen mit falschem Wert), je ein Fall; eine Umgebung mit `LASTTEST_KERNE` wird überschrieben (1); Zählprüfung 29 | `--cpus 4`; `--memory 8g`; `--network host`; `-v`; Swap erlaubt; `LASTTEST_KERNE` durchgereicht | andere |
| L10 | (Nr. 3 a, d, e, f) Reihenfolge der Prüfungen, Exit je Ausgang des Containers, Abbruch, nur `info`, `build`, `run` | Fälle der Gegenprobe: Argument (leer, nicht leer) 2; Variable falsch gleich kein Aufruf von `docker`; `docker info` scheitert gleich kein `build`; Bau scheitert gleich kein `run`; Exit des Containers 0, 1, 2, 3, 125, 126, 127, 137, 255 (9) gleich 0, 1, 2, 3, 2, 2, 2, 3, 3; INT, TERM, HUP (3) gleich Exit 2 und `abgebrochen`; Zählprüfung 2 + 3 + 9 + 3 | der Exit des Containers wird verschluckt; 137 gleich 2; ein Abbruch endet mit 143; ein weiterer Unterbefehl von `docker` | andere |
| L11 | (Nr. 1, 14 e) Das Ziel ist ein Werkzeug: kein Eintrag in `GATE_CHECKS`; die Kette prüft das Paket; der Lauf des Profils `rauch` am echten Binary schreibt die Zeilen von Nr. 13 | Ausgabe der Gates: `make gates` grün; `make -n gates` nennt `lasttest.sh` nicht (Handlauf des Verifiers); Handlauf `make lasttest LASTTEST_PROFIL=rauch` mit Exit 0 oder 1 und allen Zeilenarten von Nr. 13 (Beleg im Bericht, Last und Messender) | `GATE_CHECKS += lasttest`; ein Compile-Fehler im Paket färbt `make test` rot (Kontrolle) | andere |

Zeilen: L1 bis L11 gleich 11 (*a* = 5: L1, L2, L5, L6, L7; *b* = 6). Die Klasse `nebenläufig` setzt der Architect für Zeilen, deren Code oder Fänger nebenläufig ist (`SPEC-128` Nr. 3 e); jede dieser fünf Zeilen nennt ihren Fänger als Wirkungs-Fänger mit Rate je Prozess (`AGENTS.md` §3.10) oder als Test unter dem Detektor; L8 bis L10 sind Fälle der Gegenprobe (`tools/gegenproben/`). Größe und Schnitt: §4.

**Mengen mit Zählprüfung** (Schleife oder Generator, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15): die Zeilen nennen in der Spalte *Fänger* ihre Menge und ihre Zählprüfung; hier die Achsen der größeren.

| Menge | Achsen und Elemente | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|
| `docker run` (L9) | 5 Pflichtoptionen + 5 `--env` + 18 verbotene Formen + 1 | Aufruf als Liste | 29 gleich `SPEC-216` Nr. 3 b und c |
| Variablen (L8) | 4 Variablen × Formen (ungesetzt, gültig, leer, null, Form) | Exit 2 und Text, oder `docker run` | 7 + 5 + 11 + 8 gleich 31 |

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; die Zeilen der Zusatztabelle des ungeschnittenen Plans, soweit Fänger dieses Teils sie tragen; die Fänger der Teile 1 und 2 stehen in deren Plänen):

| Paar oder Modus | Erwartung | Fänger dieses Teils |
|---|---|---|
| Betriebsart `replay`, `mock`, `beide` | der Ablauf unterscheidet sich (`mock` fährt kein `$metadata`, `beide` zwei Folgen); die Listen und die Umgebung (F7 bis F9, Teil 1) und die Zeile `Abnahme` (A10, Teil 2) unterscheiden sich ebenfalls | L7 |

**Bestehende Fänger, die der Vertrag nennt** (keine Zeile dieses Plans): der Index `(session_id, match_key)` ohne Tabellenscan und die Gleichheit von Migration und DDL (`make schema-check`, `make migration-check`, Teil 1); die Zeile des Gesamtspans mit `parent_id` (`TestSpanZeileGesamtspan`, `slice-telemetry-spans`, Teil 2).

**Akzeptierte Negative dieses Vertrags** (Grund, Folge und Maß; ausführlich in `SPEC-216` Nr. 16; die Negative zur Auswertung und zur Konstante 100 stehen in [slice-lasttest-auswertung](slice-lasttest-auswertung.md), das zu den Setzungen der Fixtures in [slice-lasttest-fixtures](../in-progress/slice-lasttest-fixtures.md)): (1) Treiber, Stufe, Container und das gebaute Binary zusammen (Nr. 16 a, k): keine der 5 Stufen der Verdrahtung (Bau der Stufe, `docker run`, Grenzen, Binary im Container, Zeile des Spans am Binary) hat einen Fänger in der Kette, 5 von 5 belegt allein der Handlauf (`rauch`, zwei `abnahme`). (2) Setzungen ohne Messung (Nr. 16 g): sieben Setzungen, Warm-up, Toleranz, Schwelle der Last, Seitengröße, Größenklassen, Frist von 10 s und Nachlauf von 10 s; 0 von 7 gemessen, davon trägt Teil 1 zwei (Seitengröße, Größenklassen), Teil 2 drei (Warm-up, Toleranz, Schwelle) und dieser Teil zwei (Frist von 10 s in L3 und L5, Nachlauf von 10 s in L7); Folge: sie können zu eng oder zu weit sein; der Architect prüft sie an den drei ersten gültigen Läufen. (3) Der Treiber teilt die zwei Kerne (Nr. 16 i): `treiber_cpu_s` je Lauf in der Ausgabe, kein Grenzwert. (4) Der Abbruch durch `SIGKILL` am Treiber (Nr. 16 d): der Container bleibt bis zum Ende des Laufs, `--rm` räumt ihn danach; Folge: ein Verzeichnis oder Container ohne Wirkung auf das Ergebnis.

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Der Lauf des Lasttests selbst trägt `--cpus=2 --memory=4g` nach `LH-QA-01.a` und liegt ebenfalls unter `flock /tmp/odata-recorder-heavy.lock`, damit kein anderer schwerer Lauf die Messung stört; die Rolle meldet die Last des Hosts mit der Zahl.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: die Fälle der Gegenprobe), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Eine Zahl aus einem Lauf nennt Zahl der Prozesse, Last und Messenden; „stabil“ oder „schwankt wenig“ steht nur mit Zahl (`AGENTS.md` §3.10).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-lasttest-fixtures](../in-progress/slice-lasttest-fixtures.md) (Teil 1) und [slice-lasttest-auswertung](slice-lasttest-auswertung.md) (Teil 2) liegen in `done/` (der Lauf ruft den Aufbau, die Listen, die Umgebung, die Uhr, den Leser und das Urteil); `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-replay` (Teil 1 und Teil 2 des Schnitts vom 2026-11-22; `slice-telemetry-einbau-record`, Teil 3, ist keine Bedingung), `slice-telemetry-einbau-mock` und `slice-telemetry-einbau-mock-mutation` (Teile 1 und 2 des Schnitts vom 2026-11-21) liegen in `done/` (geprüft am 2026-11-23: alle vier); die Entscheidung des Users zum Zielwert von `LH-QA-01` liegt vor (2026-11-23, siehe unten). Der Vertrag des Werkzeugs samt Randformen liegt in `harness/vertraege/` ([SPEC-216](../../../../harness/vertraege/SPEC-216-lasttest.md), Architect, 2026-11-23), Tabelle und Raster in §3 und §6, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Bedingung der Gate-Kette (Übergangsregeln Nr. 7): dieser Teil ändert sie und startet nacheinander mit jedem Slice, der `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` oder den Stempel ändert; Reihenfolge der Kette: nach [slice-test-security](slice-test-security.md) in `done/` (Planner, 2026-11-24; die Reihenfolge steht im Drift-Log der Roadmap, das Auslassen eines Vorgängers regelt Nr. 7 der Übergangsregeln).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-einbau-gateway` (Teil 1) und `slice-telemetry-einbau-replay` (Teil 2): der Gesamtspan `odata.gateway.request` im HTTP-Adapter ist der Messpunkt nach `LH-QA-01.a` (Teil 1), `odata.replay.match` weist den Anteil der Auswahl aus (Teil 2; Eintrag des Planners vom 2026-11-22, Kante nach dem Schnitt von `slice-telemetry-einbau-gateway`); mit ihnen kommen mittelbar `slice-http-adapter`, `slice-replay-service` und `slice-bootstrap-m2` (M2), ohne die es das Binary im Modus `replay` nicht gibt. Der Capture-Span (Teil 3) ist keine Bedingung: `LH-QA-01.a` nimmt `record` aus.
- `slice-telemetry-einbau-mock` und `slice-telemetry-einbau-mock-mutation`: `odata.mock.query` weist den Anteil der Query-Ausführung aus (Teil 1), `odata.workspace.mutation` den der Mutation (Teil 2); mit ihnen kommen mittelbar `welle-mock` und `slice-bootstrap-m3` (M3), ohne die es den Modus `mock` nicht gibt.
- **Entscheidung des Users zum Zielwert** ([Bericht](../../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §10 Nr. 8): der User hat am 2026-11-23 ausdrücklich beschlossen, gegen den vorläufigen Wert des Lastenhefts zu messen, p95 ≤ 100 ms. Die Bedingung des Starts ist damit erfüllt. Die Bestätigung des Zielwerts durch den Owner bleibt Teil von M7; ein späterer anderer Wert entwertet die Zahlen nicht, ändert aber die Abnahme-Aussage, weil das Werkzeug den Wert als Parameter trägt (Auftrag 5).

**Abhängig von `welle-mock` / M3 und von M2, weil** der Lasttest beide Pfade (Replay und Mock) am Gesamtspan misst und der Lasttest des Mockservers den Modus `mock` braucht; er ist der Grund des Start-Triggers der Welle ([welle-betrieb](../welle-betrieb.md) §2). Ein Teilstart nur für Replay (nach `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-replay`, ohne Mock) ist möglich, ergibt aber nicht die DoD und wird hier nicht geplant; schneidet der User so, geht der Mock-Teil an einen Folge-Slice und der Planner passt diese Datei an. **Nicht Bedingung:** `slice-container-image`, `slice-listener-tls` und `slice-audit-log` (der Audit-Eintrag der Löschung liegt nicht im Messpfad). Reihenfolge-Vorsicht, kein Blocker: ein Lauf von 10 Minuten bei 20 Requests/s schwankt mit der Last des Hosts (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 2×); liegt `welle-dev-container` in `done/`, ist der Lauf dort vor fremder Last geschützt, sonst gilt der `flock` der Betriebsregeln und der Bericht nennt die Last.

**Größenmarke** (`SPEC-128` Nr. 1 und 10 g, `AGENTS.md` §3.14; Teil des Architect, 2026-11-23, als **Vorschlag**: `basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit mit `make groessenmarke`; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Dieser Teil hat 11 Zeilen (L1 bis L11), *a* = 5, *b* = 6, *f* = 2 (Ersatz-Dienst, Ersatz für `docker`), *p* = 768 (Erwartung des Architect für den Code außerhalb der Tests: Prozess 180, Lastgenerator 180, Hauptprogramm 180, Treiber 140, Fragment, Stufe und Zeile in `harness/README.md` 44, Regel für `forbidigo` in `.golangci.yml` 8 und Rest 36; die Erwartung ist geschätzt; am Ende maßen csdl 1506, json 1640 und der Import 1802 Zeilen Produktcode, `SPEC-128` Nr. 2; beim Schnitt am 2026-11-23 waren es 760 ohne die Regel, die 8 Zeilen setzt der Planner am 2026-11-24 nach `SPEC-216` Nr. 14 g ein: ein `# Why:`-Kommentar von drei Zeilen, vier Zeilen der Regel und eine Leerzeile, nach dem Vorbild der Regel zu `tests/e2e/uebergang/main.go`): *S* = 85·5 + 80·6 + 200·2 + 768 = 425 + 480 + 400 + 768 = 2073, 135·2073 = 279855, **Marke 2800** = 100 · ⌈27,9855⌉ (unverändert gegenüber *S* = 2065); Schnittfrage: 135·*S* > 300000, *S* ≥ 2223 (`SPEC-128` Nr. 1): **nein**, der Teil liegt 150 unter der Schnittgrenze (knapp zwei Zeilen: eine weitere Zeile *b* kostet 80 und gibt *S* = 2153, zwei lösen die Schnittfrage neu aus, *S* = 2233; wächst die Tabelle im Code, `AGENTS.md` §3.9, prüft der Planner sie neu). Der ungeschnittene Plan hatte *S* = 6055 und die Marke 8200 und die Schnittfrage weit überschritten; die Summe der drei Teile, 2000 ([slice-lasttest-fixtures](../in-progress/slice-lasttest-fixtures.md), Marke 2700) + 1990 ([slice-lasttest-auswertung](slice-lasttest-auswertung.md), Marke 2700) + 2073 (dieser Teil, Marke 2800) = 6063, liegt 8 über ihm: die Regel in `.golangci.yml` gab es beim Schnitt nicht; sonst keine Abweichung vom Vorschlag des Architect bei den Zahlen. Die Teile laufen nacheinander (1, 2, 3), und nur dieser Teil ändert die Gate-Kette. Block (Teil des Architect: `pfad`, `produktcode`, `fakes`; die Dateien einzeln nennt der Implementer im selben Commit, sobald er sie anlegt, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`):

```groessenmarke
pfad: tests/lasttest/
pfad: tools/harness/lasttest.sh
pfad: harness/mk/lasttest.mk
pfad: Dockerfile
pfad: harness/README.md
pfad: tools/gegenproben/
pfad: tools/harness/gegenproben-eingabe.sh
pfad: .golangci.yml
produktcode: 768
fakes: 2
```

`tests/lasttest/` ist das Verzeichnis der drei Teile (die Zeilen von Teil 1 und 2 liegen vor der Basis dieses Teils und zählen nicht). `Dockerfile`, `harness/README.md` und `tools/gegenproben/` ändern auch andere Slices der Gate-Kette; sie laufen nacheinander (Übergangsregeln Nr. 7), die Messung zählt jede Änderung an den Pfaden zwischen `basis` und `HEAD` (`SPEC-128` Nr. 10 g): der Planner prüft beim Start, dass die Basis hinter dem letzten Slice der Kette liegt, und der Implementer nimmt `main` erst nach dem Handoff auf, wenn es keine Zeile an diesen Pfaden trägt (§6, Eintrag *fremde Zeilen*). Das Verzeichnis `tools/gegenproben/` steht im Block, weil die Namen der neuen Fälle vor dem Code nicht feststehen; der Implementer nennt die Dateien einzeln, sobald er sie anlegt. `.golangci.yml` steht im Block mit der einen Regel dieses Teils; andere Slices, die das Lint-Profil ändern, nennt kein Plan unter `open/` und `in-progress/` (geprüft am 2026-11-24), sonst gälte die Regel für fremde Zeilen wie bei den übrigen Pfaden (§6). Der Stempel `harness/gegenproben.stempel` steht nicht im Block (Integrationsschritt). Wird der Schnitt nicht angenommen, trägt der Plan des ungeteilten Slice einen Block mit allen Pfaden, `produktcode` 2500 und `fakes` 4, `schnitt-grund` und die Marke 8200 (`SPEC-128` Nr. 5 g); er gilt nicht mehr.

**Reihenfolge und Kollisionen** (Architect, 2026-11-23; Prüfung der Pläne unter `open/` und `in-progress/`, `git worktree list`; vom Planner beim Schnitt auf diesen Teil beschränkt): (a) **Dieser Teil** ändert `Dockerfile`, `harness/mk/` (neues Fragment), `harness/README.md`, `tools/gegenproben/` und `tools/harness/gegenproben-eingabe.sh` (dazu `.golangci.yml` mit einer Regel, kein Pfad der Kette nach Nr. 7 der Übergangsregeln): ein Slice der Gate-Kette der Linie C (Übergangsregeln Nr. 7), er läuft nacheinander mit jedem anderen Slice, der diese Dateien ändert: die Pläne unter `open/`, die sie in §3 nennen, sind `slice-test-security` (Ziel am Ende von `GATE_CHECKS`, `Dockerfile`, `harness/README.md`), `slice-container-image` (`Dockerfile`, Fragment, Gegenproben), `slice-worktree-parallel` (`harness/mk/`, `Makefile`, Gegenproben), `slice-coverage-gate` (`Dockerfile`, `harness/README.md`, Gegenproben), `slice-abdeckung-deklaration` (`harness/README.md`, Gegenproben), `slice-handbuch-mock` und `slice-handbuch-telemetrie` (Gegenproben und Stempel); die Reihenfolge steht fest (Planner, 2026-11-24, Drift-Log der Roadmap): nach `slice-test-security`, vor den Teilen von `slice-dev-container-*` und allen weiteren Slices der Kette; der Lasttest hängt an keinem von ihnen außer an dieser Reihenfolge, und keiner an ihm. (b) Die Teile 1 und 2 legen nur Dateien unter `tests/lasttest/` an und kollidieren mit keinem Slice. (c) `welle-betrieb` §4 und §5 (die Liste der Slices und der Abhängigkeiten) und die Roadmap schreibt der Planner beim Schnitt nach (Übergangsregeln Nr. 4); die Kanten: Teil 1 → Teil 2 → Teil 3, Teil 3 nach den vier Einbau-Slices in `done/`. (d) Die Zeile in der Liste der Verträge (`harness/vertraege/README.md`) und `SPEC-216` stehen mit dem Zweig `spec-lasttest`; `SPEC-102` (Zählung der Fälle) zieht der Implementer dieses Teils nach.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke dieses Teils, die der Planner beim Start setzt (`SPEC-128` Nr. 5 d: ein Eintrag mit Grund deckt eine Überschreitung); ein Teil, der die Schnittgrenze nach dem Code überschreitet, wird nicht nachgeschnitten, sondern mit Messung und Grund berichtet (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert): die Spans sind außerhalb des Prozesses nicht lesbar (der Abrufweg ist die Span-Zeile, `SPEC-231` Nr. 6; sie fehlt, solange die Einbau-Slices nicht in `done/` liegen).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor und enthält die Zahlen des Laufs (p95 für Replay und Mock, Last, Messender).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Zielwert von `LH-QA-01` ist vorläufig und nicht freigegeben — **Ausgang:** entschieden (User, 2026-11-23): der Slice misst gegen den vorläufigen Wert p95 ≤ 100 ms; die Bestätigung durch den Owner bleibt Teil von M7 ([Bericht](../../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §10 Nr. 8), ein anderer Wert danach ändert die Abnahme-Aussage, nicht die Zahlen.
- Jede Zahl p95 hängt an Last und Messendem; zwei Läufe unter verschiedener Last sind keine Wiederholung — **Ausgang:** entschieden (Architect, 2026-11-23, `SPEC-216` Nr. 11 und 12 c): das Werkzeug nimmt Proben der Last (24 je Lauf, ungültig über der halben Kernzahl, Teil 2) und gibt sie mit aus, zwei Läufe je Betriebsart liegen in der Toleranz oder das Urteil ist `ungueltig`; den Messenden trägt die Rolle in die Closure-Notiz dieses Teils (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 2×). Die Setzungen (Toleranz, Schwelle, Warm-up) sind ungemessen und tragen den Trigger der Prüfung (`SPEC-216` Nr. 16 g, Adresse: Architect nach dem dritten gültigen Lauf je Betriebsart).
- Ein Lauf von mehr als 10 Minuten reißt die Frist einer Gate-Kette oder Gegenprobe — **Ausgang:** entschieden (Architect, 2026-11-23, `SPEC-216` Nr. 1): das Werkzeug liegt außerhalb der Kette, nach dem Muster von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, 1×); die Kette prüft das Paket ohne Lauf.
- Der Abrufweg der Spans ist nicht spezifiziert ([welle-betrieb](../welle-betrieb.md) §6); ohne lesbaren Gesamtspan misst das Werkzeug an der Gegenseite (Client) und nicht im Prozess — **Ausgang:** entschieden (Planner, 2026-11-07, nach `slice-telemetry-spans`): der Abrufweg ist die Span-Zeile (`SPEC-231` Nr. 6), geliefert vom Adapter und gefangen von `TestSpanZeileGesamtspan`; er trägt, sobald `slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-replay` und `slice-telemetry-einbau-mock` die Spans rufen (Start-Trigger in §4); der Leser liegt in Teil 2. Eine Messung an der Gegenseite entfällt.
- Das Durchfallen des Zielwerts — **Ausgang:** offen; Carveout mit Frist, Auflösungs-Trigger und Folge-Slice (Tuning), nie ein stilles Rot ([welle-betrieb](../welle-betrieb.md) §6). Das Durchfallen ist wahrscheinlich, nicht beziffert: `SPEC-135` Nr. 8 liest je Request die ganze Session (10.000 Exchanges), und `SPEC-195` liest je Collection alle Einträge des Sets; beide Verträge nennen den Lasttest als Beleg (`SPEC-135` Nr. 8 und *Ungeprüft*, `SPEC-195` *Ungeprüft*), und `SPEC-135` nennt einen Port `ListByMatchKey` als Antwort. Adresse der Folge: Planner, mit dem Ergebnis der Closure.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster steht unten, jede Zelle mit Ort in `SPEC-216` oder `gilt nicht` mit Grund — **Ausgang:** entschieden (Architect, 2026-11-23).
- Der Schnitt vor dem Code (Schätzung des ungeschnittenen Plans 6055, Marke 8200, über der Schnittgrenze von 2223, §4) — **Ausgang:** entschieden (Planner, 2026-11-23, nach dem Vorschlag des Architect): Schnitt in drei Teile (Fixtures, Auswertung, Lauf) mit den Marken 2700, 2700 und 2800 (dieser Teil), nachgerechnet aus der Tabelle (keine Abweichung bei den Zahlen); dieser Teil trägt die Kennung `slice-lasttest` statt `slice-lasttest-lauf`; `basis`, `schaetzung` und `marke` setzt der Planner je Teil beim Start (§4); die Pläne und die Kanten stehen in `welle-betrieb` §4 und §5.
- Dieser Teil ändert die Gate-Kette (`Dockerfile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/`) und läuft nacheinander mit den Slices der Linie C, die dieselben Dateien ändern (§4, Übergangsregeln Nr. 7) — **Ausgang:** entschieden (Planner, 2026-11-24): Reihenfolge der Kette nach der Priorität des Users (`welle-admin-api`, `welle-betrieb`, `welle-dev-container`, `welle-expand-batch`, Handbuch-Slices, `welle-sensoren`): `slice-test-security`, dann dieser Teil, dann die Teile von `slice-dev-container-*`; der Start-Trigger (§4) nennt den Vorgänger, das Drift-Log der Roadmap die ganze Reihenfolge.
- Das Hauptprogramm `main.go` braucht `os.Stdout`, `forbidigo` meldet es (`SPEC-216` Nr. 14 g, gemessen: 1 Befund) — **Ausgang:** entschieden (Architect, 2026-11-24, `SPEC-216` Nr. 14 g; Planner, 2026-11-24, nachgezogen nach `AGENTS.md` §3.9): eine Regel unter `linters.exclusions.rules` in `.golangci.yml` für `^tests/lasttest/main\.go$` mit `# Why:` (`AGENTS.md` §3.2), im Commit mit `main.go`; gemessen: `make lint` 0 Befunde, `make a-check` unverändert grün; Fänger: `make lint` (ohne die Regel 1 Befund, mit ihr 0), Größe in §4 (*p* + 8). Nicht gemessen: `make test` und `make test-race` mit `main.go` (der Fall von `tests/e2e/uebergang`, den die Kette trägt) und die Stufe `lasttest`; die zwei Läufe der Tests deckt `make gates` (`make test`, `make test-race`), die Stufe allein der Handlauf (L11, akzeptiertes Negativ 1 in §3).
- Fremde Zeilen im Block (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`): `Dockerfile`, `harness/README.md` und `tools/gegenproben/` sind Pfade, die andere Slices ändern; ein Merge von `main` in den Zweig dieses Teils bringt deren Zeilen in die Messung — **Ausgang:** offen; Adresse: Planner (beim Start: die Basis hinter dem letzten Slice der Kette, beim Handoff: `git diff --numstat` über die Pfade gegen die Zeilen aus `main` trennen und mit Grund berichten).
- Die Dauer des Aufbaus (10.000 Aufnahmen und ein Import dieser Größe), die Größe der Datenbank und die Zeit eines vollen Laufs (rund 50 Minuten nach der Rechnung des Vertrags) sind ungemessen — **Ausgang:** offen; der Handlauf des Implementers dieses Teils liefert die Zahlen (Profil `rauch` und ein Lauf `abnahme`) und trägt sie hier ein; liegt der Aufbau über der Frist von 30 min, geht die Frage an den Architect (`SPEC-216` Nr. 16 e; der Aufbau selbst liegt in Teil 1).
- Eine Sendung von `slice-telemetry-einbau-record` (§6, akzeptiertes Negativ 1: „die Dauer des Capture-Spans misst `slice-lasttest`“) trifft den Ausschluss dieses Teils *Latenz im Record-Betrieb* (`LH-QA-01.a` nimmt `record` aus): dieser Teil ist dafür keine Adresse — **Ausgang:** offen; Adresse: Planner (Closure von `slice-telemetry-einbau-record`: der Geber nennt dort einen anderen Nehmer oder streicht den Verweis; `AGENTS.md` §3.13).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23; die Zeilen dieses Teils; die Zeilen der Teile 1 und 2 stehen in deren Plänen; jede Zelle nennt ihren Ort in `harness/vertraege/` oder `gilt nicht` mit Grund):

| Operation | fehlender Wert (Zielwert, Startwert, Binary) | leerer Wert (leere Fixture, keine Antwort) | unbekannter oder fremder Wert (Fehlerantwort, Zeitüberschreitung, Zielwert 0) | Abbruch (Lauf vorzeitig beendet) | Wiederholung (zweiter Lauf, andere Last) | Betriebsart (Replay, Mock) |
|---|---|---|---|---|---|---|
| Variablen prüfen und den Container starten (Treiber) | `SPEC-216` Nr. 2: ungesetzt gilt die Vorgabe und wird eingesetzt (`LASTTEST_BETRIEBSART` `beide`, `LASTTEST_PROFIL` `abnahme`, `LASTTEST_ZIEL_MS` 100, `LASTTEST_STARTWERT` 1); `docker` fehlt: Nr. 3 a (Exit 2) | `SPEC-216` Nr. 2: gesetzt und leer ist `<VAR> leer`, Exit 2; ein leeres Argument am Treiber Nr. 3 a | `SPEC-216` Nr. 2: `null` (Zielwert 0 und 00), `Form` (Großbuchstabe, Vorzeichen, Dezimalpunkt, zu lang, führende Null außer `0` beim Startwert, Wert außerhalb der Aufzählung), Exit 2; Exit des Containers außerhalb von 0 bis 3: Nr. 3 d | `SPEC-216` Nr. 3 e: INT, HUP, TERM an den Treiber, Exit 2, `--rm`; SIGKILL: Nr. 16 d | `SPEC-216` Nr. 3 a und d: derselbe Aufruf gibt dieselbe Prüfung und denselben Exit | `SPEC-216` Nr. 2 (`LASTTEST_BETRIEBSART`) und Nr. 3 b (die Grenzen sind für beide gleich) |
| Last fahren (20 Requests/s, Warm-up, 10 Minuten) | `SPEC-216` Nr. 9 b: das Binary startet nicht oder lauscht nie: `binary`; Nr. 9 a: nur fünf Schlüssel | `SPEC-216` Nr. 8 c: keine Antwort ist eine Zeitüberschreitung (10 s) und ein Fehler; ein leerer Body ist ein Erfolg bei Status 200; eine leere Liste gibt es nicht (`gilt nicht`: Nr. 5, das Profil hat mindestens 60 Anfragen) | `SPEC-216` Nr. 8 c und 10 h: Status ≠ 200, Transportfehler, Lesefehler: Lauf ungültig (`fehler`); Nr. 10 b: Zeilen anderer Art sind `fremde_zeilen` | `SPEC-216` Nr. 4 d und Nr. 9 d (Signal beendet das Binary, kein Urteil, Exit 2); Nr. 9 c (das Binary endet im Lauf: `binary`) | `SPEC-216` Nr. 12 c (zweiter Lauf, Toleranz), Nr. 9 f (frischer Prozess, neue Sessions), Nr. 11 (andere Last: ungültig über der halben Kernzahl) | `SPEC-216` Nr. 7 a (`replay`), Nr. 7 b (`mock`), Nr. 9 a (Umgebung je Betriebsart) |

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

*Wird bei der Closure gefüllt.*

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30), beim Schnitt am 2026-11-23 erneut; der Planner zieht Treffer und Zähler beim Start nach. Treffer: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, jede Zahl p95 hängt an Last und Messendem und trägt beides) und `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, lange Läufe in der Gate-Kette; Entscheidung des Architect: Gate oder Werkzeug, Muster von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für Testhilfen, die `ctx` und Ströme des Lastgenerators umhüllen; offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
