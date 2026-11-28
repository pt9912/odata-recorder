# Slice slice-betrieb-spans-nachweis: Ein Test belegt Record, Replay und Mock mit den fünf Spans und den Zählern

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md) — Kriterium der Closure (§3 dort), das kein Slice der Welle als einen Test geliefert hat.

**Bezug:** `LH-QA-05`, `LH-QA-04`, `LH-FA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md).

**Berührte Spec-Stellen:** `SPEC-026` bis `SPEC-030` (Spans und Pflicht-Attribute) · `SPEC-230` (Zähler je Fehlercode) · `SPEC-294` Nr. 5 (welche Fehler gezählt werden). Der Slice ändert keine Stelle.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-betrieb-spans-nachweis`, Worktree `/Development/odata-recorder-wt-spans`; der Zweig entsteht von `main` nach den Start-Commits, der Start liegt auf `main`)

**Autor:** pt9912 (Planner, Vorbereitung der Closure von `welle-betrieb`). **Datum:** 2026-11-28.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Ein Test fährt über den Composition Root die Modi `record`, `replay` und `mock`, findet die Spans `SPEC-026` bis `SPEC-030` mit ihren Pflicht-Attributen ohne Query-Literale und Bodies und sieht je ausgelöstem Fehlercode den Zähler des Adapters der Telemetrie steigen.

**Anlass** (Planner, 2026-11-28, gelesen, nicht gefahren): Die Teile sind einzeln belegt, in keinem Test zusammen. Spans je Modus: `TestWurzelGesamtspanJedeAnfrage`, `TestWurzelCaptureLiegtImGesamtspan`, `TestWurzelReplayMatchLiegtImGesamtspan`, `TestWurzelMockSpanLiegtImGesamtspan`, `TestWurzelMutationSpanZeile` (alle `src/bootstrap/`). Zähler: `TestFehlerzaehlerJederFall` (`src/adapters/driving/httpodata/`, Fake der Telemetrie); mit dem echten Adapter über den Root lesen `TestWurzelHybridUpstreamNichtErreichbar` (`ODR-E4001`, Modus `hybrid`) und `TestAdminAbweisenBegrenzt` (`ODR-E2101`, Admin-Listener) den Zähler, kein Test liest ihn für einen Code, den ein Client am OData-Listener in `record`, `replay` oder `mock` erhält.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung am Produktcode oder an der Spezifikation — der Slice belegt vorhandenes Verhalten; zeigt der Test eine Abweichung, hält der Implementer an (§4).
- Der Modus `hybrid` — das Kriterium der Welle nennt ihn nicht; seine Spans und Zähler tragen die Tests von [`slice-telemetry-einbau-hybrid`](../done/slice-telemetry-einbau-hybrid.md).
- Ein Lauf am gebauten Binary — das Kriterium verlangt „ein Test“; der Stand der Zähler ist im Prozess am Adapter lesbar.
- Ersatz oder Kürzung der Tests oben — sie bleiben als Fänger ihrer Verträge.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] Ein Test über `record`, `replay` und `mock` findet jeden der fünf Spans mit jedem Pflicht-Attribut; ein Canary als Query-Literal und in Request- und Response-Body steht in keiner Zeile der Senke.
- [ ] Derselbe Test vergleicht je Modus den Stand der Zähler mit genau den Codes, die sein Client als Fehlerobjekt erhielt; mindestens ein Code je Modus.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/bootstrap/telemetrie_betrieb_test.go` | neu | eine Testfunktion, `TestBetriebSpansUndZaehler`; Helfer der vorhandenen Testdateien (`baueWurzelMit`, `mockDaten`, `zeilen`, `keinVerwurf`, `gwClient`); der Upstream in `record` ist ein `httptest`-Server, `replay` fährt über der Aufzeichnung dieses `record`-Laufs |

Der Bericht nennt je Zusage der DoD eine Mutation (ein Span fehlt, ein Pflicht-Attribut fehlt, der Canary als Attribut, ein Zähler steigt nicht) und den roten Lauf.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Planner beim Start, 2026-11-28; kein Architect-Lauf, weil der Slice keinen neuen Vertrag liefert: jede Zusage steht schon in `SPEC-026` bis `SPEC-030`, `SPEC-230`, `SPEC-232`, `SPEC-290` und `SPEC-294` bis `SPEC-296`). Die Spalten sind die von `SPEC-128` Nr. 3 a. Fänger ist in jeder Zeile die eine Testfunktion in `src/bootstrap/telemetrie_betrieb_test.go` (Name nach Wahl des Implementers, hier `TestBetriebSpansUndZaehler`); er führt die Anfragen über den OData-Listener. Mutationen am Produktcode laufen nur in einer Kopie im Container, nie als Änderung im Zweig.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| BS1 | (`SPEC-026`, `SPEC-294` Nr. 4) In `record`, `replay` und `mock` trägt je Anfrage ein Span `odata.gateway.request` `http.request.method`, `http.response.status_code` und `recorder.mode` mit den Werten der Anfrage und des Modus, bei Erfolg und bei Fehlerobjekt | `TestBetriebSpansUndZaehler` | `recorder.mode` fest auf einen Modus; Status nicht gesetzt | andere |
| BS2 | (`SPEC-027`, `SPEC-295` Nr. 4) `record` schreibt `odata.record.capture` mit `recorder.session_id` der offenen Session und `recorder.capture_status` `complete` | `TestBetriebSpansUndZaehler` | der Span entfällt; Status bleibt `unbestimmt` | andere |
| BS3 | (`SPEC-028`, `SPEC-296`) `replay` schreibt `odata.replay.match` mit `recorder.match_strategy` und `recorder.match_result` | `TestBetriebSpansUndZaehler` | der Span entfällt; ein Attribut fehlt | andere |
| BS4 | (`SPEC-029`, `SPEC-290` Nr. 4 a, b) `mock` schreibt `odata.mock.query` mit `recorder.entity_set` (Name aus dem Modell, nicht der Platzhalter) und `recorder.coverage_status` (nicht `nicht_gelesen`) bei einer Anfrage an eine Menge | `TestBetriebSpansUndZaehler` | `Set` des Namens entfällt; `Set` der Coverage entfällt | andere |
| BS5 | (`SPEC-030`, `SPEC-290` Nr. 4 c, d) Eine Mutation in `mock` schreibt `odata.workspace.mutation` mit `recorder.workspace_id` und `recorder.mutation_type` | `TestBetriebSpansUndZaehler` | der Span entfällt; falsche Art | andere |
| BS6 | (`LH-QA-04.a`) Ein Canary als Literal der Query steht in `record`, `replay` und `mock` in keiner Zeile der Senke | `TestBetriebSpansUndZaehler` | der Gesamtspan oder ein Dienst gibt die Query in ein Attribut oder Feld | andere |
| BS7 | (`LH-QA-04.a`) Ein Canary im Request-Body steht in `record`, `replay` und `mock` in keiner Zeile der Senke | `TestBetriebSpansUndZaehler` | ein Dienst gibt ein Stück des Request-Bodys aus | andere |
| BS8 | (`LH-QA-04.a`) Ein Canary im Response-Body steht in `record`, `replay` und `mock` in keiner Zeile der Senke | `TestBetriebSpansUndZaehler` | ein Dienst gibt ein Stück des Response-Bodys aus | andere |
| BS9 | (`SPEC-230` Nr. 8, `SPEC-294` Nr. 5) `record`: der Stand der Zähler über alle Codes ist genau die Menge der Codes der Fehlerobjekte, die der Client erhielt, mit Anzahl; mindestens ein Code | `TestBetriebSpansUndZaehler` | `CountError` entfällt; zählt doppelt | andere |
| BS10 | (wie BS9) `replay` | `TestBetriebSpansUndZaehler` | wie BS9 | andere |
| BS11 | (wie BS9) `mock` | `TestBetriebSpansUndZaehler` | wie BS9 | andere |

Ein Canary zählt nur, wenn der Test zeigt, dass er den Prozess erreicht hat (beim Upstream, in der Antwort an den Client oder im Fehlerfall an der Stelle, die ihn liest); je Ort ein eigener Wert, und der Wert erfüllt die Form `Token`, damit nicht allein die Formprüfung des Adapters ihn fernhält (`BEO-REPO/fixture-gleicht-dem-wert-der-mutation`).

**Mutationen und Raten** (Implementer, 2026-11-28; Wegwerfkopie in einem Container mit 4 CPU und 4 GiB ohne Netz, Last des Hosts 4 bis 7, ohne Detektor, je Mutation ein Prozess von `go test -count=5`). Fänger jeder Zeile ist `TestBetriebSpansUndZaehler`; 40 Mutationen, jede rot in 5 von 5 Läufen, jede aus der Prüfung, die ihre Zusage hält.

| Zusage | Mutation | Zahl | Roter Befund des Tests |
|---|---|---|---|
| BS1 | `recorder.mode` fest `record`; Methode fest `GET`; Status nicht gesetzt; Gesamtspan entfällt; Gesamtspan entfällt allein bei Fehlerobjekt | 5 | Attribute des Gesamtspans bzw. Zahl der Gesamtspans |
| BS2 | Capture-Span entfällt; Status bleibt `unbestimmt`; Status fest `complete`; `recorder.session_id` ist die Kennung des Exchange | 4 | Span unter dem Gesamtspan |
| BS3 | Match-Span entfällt; `recorder.match_result` entfällt im `Set`; Ergebnis fest `treffer`; Strategie fest `sequential` | 4 | Span unter dem Gesamtspan |
| BS4 | `Set` des Namens entfällt; `Set` der Coverage entfällt; Name fest `Products`; Coverage fest `complete` | 4 | Span unter dem Gesamtspan |
| BS5 | Mutations-Span entfällt; Art fest `post`; `recorder.workspace_id` ist die Kennung des Dienstes | 3 | Span unter dem Gesamtspan |
| BS6, BS7, BS8 | der HTTP-Adapter gibt in genau einem Modus die Query, den Request-Body oder den Antwort-Body als Text eines Eintrags aus (3 Orte × 3 Modi) | 9 | „die Senke trägt" den Wert des Orts im Modus der Mutation |
| BS9, BS10, BS11 | je Modus: `CountError` entfällt; zählt doppelt (6). Je Code nur das erste Mal (rot in `replay` und `mock`); falscher Code (rot in allen drei Modi) | 8 | Stand der Zähler ungleich den erhaltenen Codes |
| „keine weitere Zeile, kein Verwurf" (Kopf des Tests) | Eintrag mit unbekanntem Schlüssel; zweiter Span unter dem Gesamtspan; Span ohne Eltern | 3 | Verwurf `schluessel`; Spans unter dem Gesamtspan; Zahl der Zeilen der Dienste |

Die Mutationen zu BS6 bis BS8 sitzen im HTTP-Adapter mit einer Bedingung auf den Modus, nicht in den Diensten; der Fänger durchsucht die Senke und hängt nicht an der Stelle, die schreibt.

**Läufe des unveränderten Tests** (derselbe Container): ohne Detektor 30 von 30 in einem Prozess und 10 von 10 Prozessen, 0,52 bis 0,64 s je Lauf; mit Detektor 30 von 30 und 10 von 10, 0,83 bis 1,04 s je Lauf (Last 6 bis 7). Das Paket `src/bootstrap` allein unter dem Detektor (`go test -race -count=1`, ein Lauf, 4 CPU, Last vorher 6,6, nachher 3,8): 422,4 s von 600 s; der Lauf ist kein Vergleich mit den 415,9 s aus §6.

**Akzeptierte Negative** (`AGENTS.md` §3.15):

- Der Test sucht den ganzen Wert in jeder Groß- und Kleinschreibung, kein Teilstück und keine Kodierung. Grund: die abgeleiteten Formen sind unbeschränkt. Folge: eine Stelle, die einen Wert gekürzt oder kodiert ausgibt, bleibt hier ungefangen.
- Gesucht wird allein in der Senke des Adapters des Prozesses, nicht im Stand `Snapshot`, im Logger von `Build` und auf `stdout` und `stderr` neben der Senke. Grund: BS6 bis BS8 sagen die Senke zu. Folge: ein Wert dort bleibt hier ungefangen.
- Die Orte sind Literal der Query, Request-Body und Antwort-Body; Pfad, Namen in der Query und Header führt der Test nicht. Grund: die DoD nennt diese Orte nicht; die Tests der Dienste und des Adapters führen sie. Folge: wie dort.
- Gezählt werden 5 der 41 Codes (`ODR-E4001`, `ODR-E5001`, `ODR-E5003`, `ODR-E6008`, `ODR-E7003`); 36 bleiben ungeführt. Grund: die Zählung läuft für jeden Code durch eine Stelle (`SPEC-294` Nr. 5), die Codes führt `TestFehlerzaehlerJederFall` mit dem Fake. Folge: eine Ausnahme je Code am echten Adapter für einen ungeführten Code bleibt hier ungefangen.
- Die Attribute führen nicht jeden Wert ihrer Menge: `capture_status` ohne `failed`, `match_result` ohne `erschoepft`, `anderer_key` und `nicht_replayfaehig`, `match_strategy` allein `last`, `coverage_status` ohne `unknown`. Grund: die Zusage ist das Pflicht-Attribut je Span, die Mengen halten die Tests der Dienste. Folge: ein falscher Wert allein auf einem ungeführten Weg bleibt hier ungefangen.

```groessenmarke
pfad: src/bootstrap/telemetrie_betrieb_test.go
produktcode: 0
fakes: 0
basis: a819be8e
schaetzung: 880
marke: 1200
```

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein Code-Vorgänger; die Slices der Welle liegen in `done/`. Kein Kettenslice.

**Stand beim Start** (Planner, 2026-11-28, `main` bei `4c935d28`, durch Lesen und Suchen, ohne Testlauf): Die Bedingung ist erfüllt; die übrigen Slices der Welle liegen in `done/`. Der Vertrag liegt in der Spezifikation, keine Frage an den Architect ist offen; er wird erst mit der Rückführung unten gebraucht. Die Helfer in `src/bootstrap` bestehen (`baueWurzelHaken`, `mockDaten`, `replayDaten`, `fakeUp`, `TelemetrieVon`). *Kollision:* in `in-progress/` liegt daneben allein [slice-batch-parser](../in-progress/slice-batch-parser.md); sein Zweig und sein Worktree ändern keinen Pfad unter `src/bootstrap/`. Der Zweig `vertrag-ci` trägt beim Start keinen Commit über `main`. Der Slice ändert weder `harness/README.md` noch eine Datei der Gate-Kette (Übergangsregeln Nr. 7); der Test läuft in `make test` und `make test-race` mit.

**Größenmarke** (`SPEC-128` Nr. 1, `AGENTS.md` §3.14; Planner beim Start, vor dem ersten Code-Commit): elf Zeilen der Tabelle (BS1 bis BS11), *a* = 0, *b* = 11, *f* = 0 (der Test schreibt keinen Fake eines Ports), *p* = 0: *S* = 880, Marke **1200**, keine Schnittfrage. Den Teil des Blocks, den sonst der Architect setzt (`pfad`, `produktcode`, `fakes`), hat der Planner gesetzt. `basis` ist `a819be8e` (letzter Commit von `main` vor dem Commit des Blocks); Messbasis nach `SPEC-128` Nr. 10 g: der Pfad entsteht erst mit diesem Slice, Fremdzeilen beim Start 0. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Ändert der Implementer eine weitere Datei (etwa `export_test.go` oder eine vorhandene Datei mit Helfern), nennt er sie im selben Commit in §3 und im Block. Für einen Schnitt beim Handoff ist kein Teil vorab benannt; eine Überschreitung steht mit Grund im Block.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß): über 1200 Zeilen (Metrik wie oben). Eine Überschreitung schneidet nicht (`SPEC-128` Nr. 10 c); sie bekommt einen Eintrag `ueberschreitung` mit Grund im Block.
- `in-progress` → `open` (blockiert): der Test findet ein Verhalten, das den Spec-Stellen widerspricht, oder er ist ohne Änderung am Produktcode nicht zu schreiben; Rückgabe an den Planner.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

Kein neuer Vertrag; ein Raster der Randformen entfällt (`AGENTS.md` §3.12 gilt für Slices mit neuem Vertrag).

- Raster *Operation × Argumentart*: entfällt, der Slice liefert keine Operation und ändert kein Verhalten; eine Randform ohne Ort in der Spezifikation geht an den Architect (`AGENTS.md` §3.12). Beim Schreiben des Tests ist keine aufgetreten, und kein Verhalten widersprach den Spec-Stellen — **Ausgang:** offen; Adresse: Planner (Closure).
- Das Paket `bootstrap` liegt unter dem Detektor bei 415,9 s von 600 s (`BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze`); der Test kommt hinzu, mit Detektor 0,83 bis 1,04 s, ohne 0,52 bis 0,64 s (§3) — **Ausgang:** weiter offen; Adresse: Planner (Closure).
- Der Capture-Span endet erst mit dem Schließen des Body; ein Test, der vorher liest, flackert. Der Test fährt je Modus den Prozess herunter, bevor er die Senke liest; 80 von 80 Läufen mit und ohne Detektor grün (§3) — **Ausgang:** weiter offen; Adresse: Planner (Closure).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten** (Planner beim Start, 2026-11-28; Zähler gleich der Zahl der Dateien unter `evidence/`, alle im Kürzel `BEO-REPO`):

- **Einschlägig, offen oder geplant:** `paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (3×, offen; Risiko in §6), `nullprobe-grenze-unter-der-paketlaufzeit` (17×, geplant in `slice-mutationslauf-modulwurzel`; Mutationen in `src/bootstrap` laufen mit eigenem Treiber).
- **Zu beachten:** `fixture-gleicht-dem-wert-der-mutation` (17×; je Attribut und je Ort des Canary ein eigener Wert), `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (93×) und `menge-ohne-ihre-positionen` (28×; die Mengen Spans, Pflicht-Attribute, Orte des Canary und Modi als Schleifen oder je eigene Prüfung, `AGENTS.md` §3.15), `allaussage-an-jeder-stelle-mit-faenger-an-erster-und-zweiter-stelle` (5×), `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (31×), `hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×), `betriebsregel-fuer-repo-dateien-verletzt` (25×; Übergangsregeln Nr. 8).
- **Planung:** `marke-beim-start-nicht-gesetzt` (1×; Marke mit diesem Start gesetzt), `schaetzformel-ueberschaetzt-den-slice` (34×; bei einem reinen Test-Slice liegt das Ende eher weit unter der Marke).

Kein Eintrag verlangt vor diesem Slice einen eigenen Slice.
