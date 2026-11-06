# Slice slice-test-determinism: `make test-determinism` — 100 Wiederholungen des Replay mit gleichen normalisierten Antworten und Cursor-Ständen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-QA-02`, `LH-FA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-038` (Vertrag `make test-determinism`, `harness/vertraege/SPEC-038-test-determinism.md`: *100 Wiederholungen identischer Inputs; identische normalisierte Outputs und Cursor-Stände*; normalisiert werden nur die Header aus `VOLATILE_HEADERS`) · `LH-QA-02.a` (Determinismus-Randbedingungen) · `LH-FA-09.b` (Reset) · `SPEC-037` Nr. 6 (Reset über die Prozessgrenze) · `SPEC-129` Nr. 10, `SPEC-135` Nr. 5, 13, `SPEC-063` · `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-test-determinism)

**Autor:** pt9912. **Datum:** 2026-10-31.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make test-determinism` (`SPEC-038`) fährt eine feste Replay-Folge 100-mal nach Reset und vergleicht die Läufe: gleiche Statuscodes, fachliche Bodies und relevante Header (normalisiert nur um `VOLATILE_HEADERS`) und gleiche Cursor-Stände. Das ist die Messmethode von `LH-QA-02`: „100 identische Testläufe nach Reset; Vergleich normalisierter Antworten“. Der Slice liefert den Teil für `replay`; der Teil für `mock` hat `slice-mock-abnahme` (100 identische Läufe nach Reset).
Der Reset läuft über das Handle des Composition Root im Prozess von `go test` (Entscheidung des Architect, `SPEC-038` Nr. 7); Aufnahme und Import der Vorbereitung nutzen die Hilfen von `tests/e2e`, `Dockerfile` und `uebergang` bleiben unverändert.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Läufe des Mock — `slice-mock-abnahme` führt sie als eigene DoD; sein Test steht neben diesem, nicht hinter dem Ziel `make test-determinism` (`SPEC-038` Nr. 9); dieser Slice ändert den Plan dort nicht und sendet nichts (`AGENTS.md` §3.13: der Nehmer führt die 100 Läufe schon in seiner DoD, aus der Entscheidung entsteht keine Pflicht).
- Der Replay am gebauten Binary und der Reset über den Admin-Listener — `SPEC-038` Nr. 10 a, b (akzeptierte Negative: `TestE2EReplay` fährt das Binary einmal, der Rundlauf von `slice-bootstrap-m6` enthält das Zurücksetzen; keine Sendung, weil der Dienst derselbe ist).
- Parallele Workspaces und Neustart (`LH-QA-03`) — `slice-mock-abnahme`.
- Zeit- und Hostabhängigkeit der Header jenseits `VOLATILE_HEADERS` — jede andere Abweichung ist nach `SPEC-038` ein Befund, kein Gegenstand.

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

- [ ] Das Ziel `make test-determinism` und der Vertrag `SPEC-038` (der Stand wird zum gelieferten Umfang); Treiber, Fragment und `TestE2EDeterminismus` mit den Vergleichstests nach der Tabelle in §3. Einbindung in die Gate-Kette oder daneben nach der Entscheidung des Users zu `slice-test-integration` §6 (§6).
- [ ] 100 Läufe derselben Replay-Folge nach Reset haben gleiche normalisierte Antworten und gleiche Cursor-Stände; die Normalisierung nimmt allein `VOLATILE_HEADERS` aus.
- [ ] Je Zusage eine Mutation, vom Implementer rot gesehen (`AGENTS.md` §3.10), nach der Tabelle *Zusage · Fänger* in §3 (unter anderem ein Reset, der den Cursor stehen lässt, und eine Abweichung eines Headers außerhalb von `VOLATILE_HEADERS`); die Fänger sind deterministisch, eine Rate entfällt.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-038`, Sensors-Tabelle in `harness/README.md`).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
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
| `harness/mk/` (Fragment `test-determinism.mk`) | neu | Ziel `make test-determinism` (das Root-Makefile bindet `harness/mk/*.mk` selbst ein); kein Eintrag in `GATE_CHECKS` (§6) |
| `tools/harness/test-determinism.sh` | neu | Treiber nach `SPEC-038` Nr. 1, 2: `docker build --target integration`, ein Lauf von `TestE2EDeterminismus`, Wertung wie `test-integration.sh` |
| `tests/e2e/determinismus_test.go` (und Hilfen in `tests/e2e/`) | neu | `TestE2EDeterminismus` (Vorbereitung wie `TestE2EReplay`, Handle im Modus `replay`, 100 Läufe), Vergleichs- und Standfunktionen mit den Tests `TestDeterminismusVergleich` und `TestDeterminismusStand` (ohne `E2E_BINARY` lauffähig, auch in `make test`) |
| `harness/vertraege/SPEC-038-test-determinism.md` | geschrieben (Architect) | Aufbau, Folge, Vergleich, Stände, Reset, Randformen; `Stand` wird beim Liefern angepasst |
| `harness/vertraege/SPEC-037-test-integration.md` | geändert (Architect) | Nr. 6: ein Satz, dass der Reset über die Prozessgrenze fehlt und `SPEC-038` Nr. 7 ihn im Prozess des Tests fährt |
| `harness/README.md` | update | Sensors-Tabelle: Ziel, Vertrag, Bindung — erst mit dem Ziel (`AGENTS.md` §4) |
| `Dockerfile`, `tests/e2e/uebergang/`, `src/` | unverändert | die Stufe `integration` trägt Test-Binary und Binary; kein Produktcode, kein Haken (`SPEC-038` Nr. 1, 7) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten, das Raster *Operation × Argumentart* (`AGENTS.md` §3.16) in §6, beide vor dem ersten Code-Commit (`AGENTS.md` §3.12).

**Zusage · Fänger** (Spalten nach `SPEC-128` Nr. 3 a; die Nr. in der Zelle *Zusage* nennen `SPEC-038`). Alle Fänger sind deterministisch: ein Prozess, nacheinander gesendete Requests, keine Nebenläufigkeit im Test; es gibt keine Rate, ein Fänger wird je Mutation in jedem Lauf rot oder gilt nicht. Die Mutationen an `src/` fährt der Implementer vorübergehend und nimmt sie zurück. `TestDeterminismusVergleich` und `TestDeterminismusStand` laufen ohne Binary (auch in `make test`), `TestE2EDeterminismus` nur im Container von `make test-determinism`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | 100 Läufe; nach Lauf 100 ist die Zahl der Läufe 100 und die `Revision` des Workspace 100 (Nr. 3) | `TestE2EDeterminismus`: Zähler und `Get(...).Revision` nach der Schleife | Schleife bis 99 statt 100 | andere |
| 2 | Vor jedem Lauf läuft der Reset: die `Revision` im Lauf i ist i, für jedes i (Nr. 3 a); Allaussage als Prüfung je Lauf in der Schleife | `TestE2EDeterminismus`: Prüfung je Lauf | der Aufruf `Reset` entfällt in Lauf 1, in Lauf 50 und in Lauf 100 (drei Mutationen) | andere |
| 3 | Der Reset setzt jeden Cursor auf `Position` 0 und leeres `Consumed`, löscht keinen und legt keinen an (Nr. 6, Stand nach Reset) | `TestE2EDeterminismus` (Prüfung nach (a)) und `TestDeterminismusStand`: Schleife über die Verstöße Position ungleich 0, `Consumed` nichtleer, Identität fehlt, zusätzliche Identität | in `src/` setzt `Cursors.ResetAll` nichts zurück; setzt nur `Position`, nicht `Consumed`; löscht die Cursor | andere |
| 4 | Die Replay-Session ist in allen Läufen dieselbe; die Zahl der Cursor wächst nicht (Nr. 3) | `TestE2EDeterminismus`: Stand nach dem Lauf gleich Lauf 1 (Länge der Liste) | der Test nimmt je Lauf eine neue Replay-Session `determinismus-<i>`, der Reset bleibt | andere |
| 5 | Orakel: Lauf 1 hat Status und Body der Aufnahme für die Requests 1 bis 8, `410` mit `ODR-E5002` für 9, `404` mit `ODR-E5001` für 10 (Nr. 4) | `TestE2EDeterminismus` | der Test lässt den Header `X-Recorder-Replay-Session` leer (jede Antwort `400`) | andere |
| 6 | Die Antworten aller Läufe sind gleich Lauf 1: Status, Body, Header außer `VOLATILE_HEADERS` (Nr. 5); Allaussage über die Arten der Abweichung als Schleife, dazu der Fall, dass zwei gleiche Antworten gleich sind | `TestDeterminismusVergleich`: Schleife über Status, ein Byte des Body (erste, mittlere, letzte Stelle), Wert eines Headers (je Header der Probe), zusätzlicher Name, fehlender Name, zusätzlicher Wert, Reihenfolge der Werte, leerer statt gesetzter Wert; je Element Unterschied, gleiche Antworten gleich | der Vergleich lässt den Body aus; vergleicht die Werte eines Namens als Menge; ignoriert einen fehlenden Namen | andere |
| 7 | Sind mehrere Verstöße zugleich da, meldet der Vergleich den ersten in der Reihenfolge Status, Body, Header (Nr. 5; Zusatztabelle in §6) | `TestDeterminismusVergleich`: Schleife über die drei Paare | Reihenfolge der Prüfung vertauscht | andere |
| 8 | Normalisiert wird allein um die Namen aus `VOLATILE_HEADERS`, in jeder Schreibweise; jeder andere Name unterscheidet (Nr. 5) | `TestDeterminismusVergleich`: Schleife über die sieben Namen des Default mal drei Schreibweisen (klein, groß, gemischt), je Name abweichender Wert gleich; Schleife über `Content-Length`, `Content-Encoding`, `Transfer-Encoding`, `Connection`, `ETag`, `X-Recorder-Match-Redacted`, je ein Unterschied | die Ausnahme enthält `Content-Length`; der Namensvergleich beachtet die Schreibweise | andere |
| 9 | Die Ausnahme ist `Konfiguration.VolatileHeaders` des Handle, keine Kopie (Nr. 5) | `TestDeterminismusVergleich`: Ausnahme aus `LadeKonfiguration` mit `VOLATILE_HEADERS=X-Probe` (`X-Probe` gleich, `Date` unterscheidet) und mit `none` (jeder Name unterscheidet) | die Liste ist im Test fest eingetragen | andere |
| 10 | Die Stände nach dem Lauf sind gleich Lauf 1 in jedem Feld: Länge der Liste, `WorkspaceID`, `ReplaySessionID`, `RecordingSessionID`, `ServiceID`, `Strategy`, `Position`, Länge und Elemente von `Consumed` (Nr. 6); Allaussage als Schleife über die Felder | `TestDeterminismusStand`: je Feld eine geänderte Kopie, je Feld Unterschied; gleiche Stände gleich | der Standvergleich lässt `Consumed` aus; vergleicht nur die Länge von `Consumed` | andere |
| 11 | Der Leser ist gültig: der Stand nach Lauf 1 ist genau ein Cursor mit Replay-Session `determinismus`, Recording-Session der Aufnahme, `Position` größer 0, `Consumed` nichtleer (Nr. 6, Kontrollwert) | `TestE2EDeterminismus` | der Leser fragt den Workspace `""` (leere Liste) | andere |
| 12 | Ein Fehler in einem Lauf beendet den Test mit Lauf und Grund (Nr. 8 a) | `TestE2EDeterminismus`: benannte Ausgabe `Lauf <i>` | das Handle wird in Lauf 50 heruntergefahren | andere |
| 13 | Ein abweichender Lauf färbt rot, die übrigen laufen weiter, die Zeile `Läufe: 100, abweichend: <k>` steht (Nr. 8 b) | `TestE2EDeterminismus`: benannte Ausgabe | in `src/adapters/driving/` setzt der Replay-Pfad einen Header `X-Probe` mit einem Zähler, der mit jeder Antwort wächst | andere |
| 14 | Negativ: Treiber (Argumente, Wertung `SKIP`, `FAIL`, kein `PASS`, `--network none`, `--memory`, `--read-only`, Frist, Stufe `integration`; Nr. 2, 8 d, 10 f) | kein automatischer Fänger; Handlauf des Implementers im Bericht | Argument `x` und `""`, `t.Skip` im Test, `--network none` entfernt | andere |
| 15 | Negativ: Binary, `main`, Admin-Reset, andere Strategien und Antwortarten (Nr. 10 a, b, e) | kein Fänger: `TestE2EReplay` fährt das Binary einmal, der Rundlauf des Admin-Listeners das Zurücksetzen | — | andere |

**Stand beim Handoff** (Implementer; Messung: jede Mutation ein Prozess, ein Lauf der Stufe `integration` im Container, vom Implementer; alle Fänger deterministisch, jede Mutation in jedem Lauf rot, eine Rate entfällt). Mutationen an `src/` und am Test fuhr der Implementer auf einer Wegwerfkopie, `src/` und Test im Zweig sind unverändert. Ein Fänger zählt, wenn die genannte Meldung in der Ausgabe steht.

| Nr. | Mutation | roter Test, Meldung |
|---|---|---|
| 1 | Schleife bis 99 | `TestE2EDeterminismus`: `Läufe: 99, erwartet 100` und `Revision nach Lauf 100: 99` |
| 2 | Aufruf `Reset` entfällt in Lauf 1, 50, 100 (je ein Lauf, der Test liest dort `Get`) | `TestE2EDeterminismus`: `Lauf 1: Revision 0, erwartet 1`, `Lauf 50: Revision 49, erwartet 50`, `Lauf 100: Revision 99, erwartet 100` |
| 3 | `ResetAll` ohne Wirkung · setzt nur `Position` · löscht die Cursor (`src/adapters/driven/sqlite/cursors.go`) | `TestE2EDeterminismus`: `Stand nach Reset: Cursor 0: Position 8, erwartet 0` · `Consumed [1 2 3 4 5 6 7 8], erwartet leer` · `Zahl der Cursor 0, in Lauf 1 1`, je `Läufe: 100, abweichend: 99`; die Verstöße im Raster (Position, `Consumed`, fehlende, zusätzliche, jede der drei Identitäten einzeln, Lauf 1 nichtleer) einzeln aus `pruefeStandNachReset` gestrichen: `TestDeterminismusStand/nach_Reset` je `nicht gemeldet` (7 Mutationen) |
| 4 | Test nimmt ab Lauf 2 eine Replay-Session `determinismus-<i>` (ab Lauf 1 scheitert schon der Kontrollwert, Zeile 11: `Lauf 1: Leser ungültig: Cursor determinismus-1/…`) | `TestE2EDeterminismus`: `Lauf 2: Stand nach dem Lauf: Zahl der Cursor 1 gegen 2`, `Lauf 3: Stand nach Reset: Zahl der Cursor 2, in Lauf 1 1`, `abweichend: 99` |
| 5 | `X-Recorder-Replay-Session` leer · `ODR-E5002` zu `ODR-E5099` · `ODR-E5001` zu `ODR-E5098` (die beiden letzten ändern alle Läufe gleich, nur das Orakel fängt sie) | `TestE2EDeterminismus`: `Lauf 1: Orakel, Request 1: 400 …` · `Orakel, Request 9: 410 … erwartet 410 mit ODR-E5002` · `Orakel, Request 10: 404 … erwartet 404 mit ODR-E5001` |
| 6 | Vergleich lässt den Body aus · Werte eines Namens als Menge · fehlender Name ignoriert · zusätzlicher Name ignoriert | `TestDeterminismusVergleich`: `Body erstes Byte … kein Unterschied gemeldet` (und mittleres, letztes, gekürzt, leer) · `Reihenfolge der Werte … kein Unterschied gemeldet` · `fehlender Name <N> …` (sechs Namen) · `zusätzlicher Name …`; jede Zeile in beiden Richtungen |
| 7 | Body vor Status geprüft | `TestDeterminismusVergleich/Reihenfolge`: `Status und Body: Meldung "Body …", erwartet Art Status` |
| 8 | Ausnahme enthält `Content-Length` · Namen des Headers nicht kanonisiert · Namen der Liste nicht kanonisiert | `TestDeterminismusVergleich/flüchtige_Namen` (`Wert von Content-Length … kein Unterschied gemeldet`) · `…/flüchtige_Namen` (Schreibweise klein, groß, gemischt) · `…/Liste_des_Handle` (Liste `x-probe`) |
| 9 | Liste fest in `kopfMenge` eingetragen (der Default) | `TestDeterminismusVergleich/Liste_des_Handle` (`VOLATILE_HEADERS=X-Probe` und `none`) |
| 10 | Standvergleich lässt `Consumed` aus · vergleicht nur dessen Länge · lässt je ein Feld aus (`WorkspaceID`, `ReplaySessionID`, `RecordingSessionID`, `ServiceID`, `Strategy`, `Position`) · prüft die Zahl der Cursor nur einseitig | `TestDeterminismusStand`: `Feld <F> von Cursor <i>: kein Unterschied gemeldet` (je Feld beide Cursor, beide Richtungen); `Cursor fehlt` und `zusätzlicher Cursor: kein Unterschied gemeldet` |
| 11 | Leser fragt den Workspace `""` · Replay-Session leer (kein Cursor) · Kontrollwert prüft die Zahl nur mit `<`, ignoriert `Position` und `Consumed` | `TestE2EDeterminismus`: `Lauf 1: Stand lesen: nicht gefunden: Workspace ` (der Port meldet für `""` einen Fehler, keine leere Liste) · `Lauf 1: Leser ungültig: 0 Cursor, erwartet 1`; `TestDeterminismusStand/Kontrollwert` für die beiden letzten |
| 12 | Handle in Lauf 50 heruntergefahren | `TestE2EDeterminismus`: `Lauf 50: Reset: ODR-E9001: Internal error` |
| 13 | `schreibeAntwort` in `src/adapters/driving/httpodata/antwort.go` setzt `X-Probe` mit wachsendem Zähler | `TestE2EDeterminismus`: `Lauf 2: Request 1: Header X-Probe: Werte ["x"] gegen ["xxxxxxxxx"]` und `Läufe: 100, abweichend: 99` |
| 14 | Handläufe des Treibers (unten) | Argument `x`, `""` und zwei Argumente: `Aufruf:`, Exit 2; `t.Skip`: `rot (SKIP)`, Exit 1; Mutation 1 im Container: `rot (Exit 1)`; SIGTERM am Treiber: Exit 2 |

*Klassen, die der Plan generierbar nennt:* die Arten der Abweichung (Zeile 6) und die Felder (Zeile 10) sind Schleifen. Vergleich der Antworten: 28 Fälle (10 Arten und 6 Namen mal 3 Arten je Name, `Date` ist ausgenommen) in 2 Richtungen, der Test zählt sie (`len(faelle) < 25` ist rot); Namen der Ausnahme: 7 mal 3 Schreibweisen, dazu 6 Namen außerhalb. Orakel dieser Schleifen: `probe()` ist gleich sich selbst. Stände: 10 Feldänderungen mal 2 Cursor mal 2 Richtungen, dazu 2 Fälle der Länge. Nach Reset: 8 Verstöße, Lauf 1 und Kontrollwert (5 Verstöße).

*Handläufe des Treibers* (Zeile 14, Nr. 10 f): Argument `x` und `''`: `Aufruf:`, Exit 2, ohne Container; `t.Skip` vor dem Test: `test-determinism: rot (SKIP)`, Exit 1; Mutation 1: `rot (Exit 1)`; SIGTERM nach 0,3 s und nach 1,2 s: Exit 2, kein Container bleibt. Ein entferntes `--network none` und ein entferntes `--read-only` bleiben grün, der Test braucht beides nicht: das ist das akzeptierte Negativ von `SPEC-038` Nr. 10 f, kein Fänger.

*Akzeptiertes Negativ (Grund, Folge):* die Aufrufstelle `k.VolatileHeaders` in `TestE2EDeterminismus` hat keinen Fänger. Grund: der Vertrag lässt `VOLATILE_HEADERS` im Lauf ungesetzt (Nr. 3), eine feste Liste gleich dem Default bleibt grün; die Funktion `vergleicheAntworten` nimmt die Liste als Argument und ist über `TestDeterminismusVergleich` gedeckt (Zeile 9). Folge: ersetzt jemand die Liste an der Aufrufstelle durch den Default, bemerkt es keiner, bis ein Lauf `VOLATILE_HEADERS` setzt.

```groessenmarke
pfad: tests/e2e
pfad: harness/mk/test-determinism.mk
pfad: tools/harness/test-determinism.sh
produktcode: 80
fakes: 0
basis: 01c547c
schaetzung: 1280
marke: 1800
ersetzt: 1500 | Tabelle Zusage-Faenger des Architect hat 15 Zeilen, alle andere (a = 0, b = 15), f = 0, p = 80: S = 80*15 + 200*0 + 80 = 1280, M = 1800 nach SPEC-128 Nr. 1; die vorlaeufige Marke 1500 kam aus S = 1100 (b = 10, f = 1, p = 100) ohne Tabelle
```

Der Block trägt den Teil des Architect: `produktcode` 80 sind Treiber (rund 70 Zeilen, Form von `test-integration.sh`) und Fragment (rund 10); `fakes` 0, weil der Referenzdienst von `tests/e2e` wiederverwendet wird und das Handle kein Fake ist. `basis`, `schaetzung`, `marke` und `ersetzt` hat der Planner am 2026-11-06 vor dem ersten Code-Commit gesetzt (§4); die Zahlen nach beiden Formeln stehen in §6.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-test-integration` liegen in `done/` (Replay läuft, Aufbau und Referenzdienst stehen). Der Vertrag `SPEC-038` trägt Umfang und Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice ändert `Makefile` und `harness/mk/` und startet nur, wenn kein anderer Slice der Gate-Kette in `in-progress/` liegt ([Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 7). Geprüft beim Start am 2026-11-06: `slice-bootstrap-m2` und `slice-test-integration` liegen in `done/` (`main` 7a2d4fa); `in-progress/` von `main` trägt außer diesem Slice nichts; die Zweige `adr-gate-laufzeit` und `slice-mutationslauf-werkzeug` tragen in ihrem `in-progress/` keinen Slice (`git worktree list`).

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`): Tabelle und Block liegen in §3 (Architect). Die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks (`tests/e2e`, `test-determinism.mk` unter `harness/mk/`, `test-determinism.sh` unter `tools/harness/`), Tests eingeschlossen, Text der Verträge ausgeschlossen; `Dockerfile` und `src/` bleiben unverändert und stehen nicht im Block. Schätzung nach `SPEC-128` Nr. 1: *a* = 0 nebenläufige Zeilen, *b* = 15 andere (alle Zeilen der Tabelle), *f* = 0 (der Referenzdienst von `tests/e2e` wird wiederverwendet, das Handle ist kein Fake), *p* = 80 (Treiber rund 70, Fragment rund 10): *S* = 80·15 + 200·0 + 80 = **1280**, *M* = **1800** (das Programm meldet beide, `make groessenmarke`). *S* liegt unter 2223: keine Schnittfrage. Den Teil des Planner hat der Planner am 2026-11-06 vor dem ersten Code-Commit gesetzt: `basis` ist `01c547c` (der Commit des Architect, der Vertrag, Tabelle, Raster und Block vor dem Code abschließt), `schaetzung` 1280, `marke` 1800; die vorläufige Prosa-Marke 1500 (*S* 1100 ohne Tabelle) ist mit Grund ersetzt (`ersetzt: 1500 | …`). Nach `AGENTS.md` §3.14 im Wortlaut (15·40 = 600, × 1,25 = 750, aufgerundet 800) ergäbe dieselbe Tabelle die Marke 800; der Konflikt ist ein offener Punkt in §6 (Adresse User) und nicht hier gelöst. Der Block gilt für `make groessenmarke`, das die Marke nach `SPEC-128` misst.

**Messung beim Handoff** (Implementer, `git diff --numstat 01c547c..ed86cc9` über die Pfade des Blocks): `test-determinism.mk` 11, `determinismus_test.go` 529, `test-determinism.sh` 73, Summe **613** gegen Schätzung 1280 und Marke 1800. Kein neuer Fake: Referenzdienst und Aufnahme-Hilfen von `tests/e2e` sind wiederverwendet. **Laufzeit der 100 Läufe** (Risiko in §6): der Test braucht 0,74 bis 0,93 s (vier Läufe), `make test-determinism` im Ganzen 2,3 s bei warmem Build-Cache; die Schätzung unter 15 s je Lauf des Ziels trägt, und die Grenze von 60 s ist weit entfernt.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): Die Marke setzt der Planner beim Start nach `AGENTS.md` §3.14 (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Zahl steht mit Tabelle und Block des Architect (Marke 1800, §4). Der Planner schätzt den Slice als klein.
- `in-progress` → `open` (blockiert): entfällt; der Reset ist entschieden (`SPEC-038` Nr. 7, §6).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Reset zwischen den Läufen (Frage an den Architect) — **Ausgang:** entfallen; entschieden in `SPEC-038` Nr. 7 (Handle des Composition Root im Prozess von `go test`, Reset über `Workspaces().Reset`, ohne Binary je Lauf, ohne Haken im Produkt); der Satz in `SPEC-037` Nr. 6 ist nachgezogen. Verworfen: Hilfsprogramm mit Neustart je Lauf, neuer Workspace je Lauf, Zurückstellen bis `slice-bootstrap-m6`; Gründe dort. Die Cursor-Stände liest der Test aus der Ablage (`Cursors().List` über `sqlite.Open`, `SPEC-038` Nr. 6). Folge: Binary, `main` und der Admin-Reset sind in der Messung nicht beteiligt (`SPEC-038` Nr. 10 a, b; akzeptierte Negative, Zeile 15 der Tabelle). Die Rückführung `in-progress` → `open` aus §4 entfällt damit. Der Implementer prüft als Erstes, dass eine zweite Verbindung auf die Ablage neben dem laufenden Handle liest (`SPEC-038` Nr. 6); scheitert das, hält er an und gibt es an den Architect zurück (`AGENTS.md` §3.12).
- Doppelte Messung mit `slice-mock-abnahme` — **Ausgang:** entfallen; entschieden in `SPEC-038` Nr. 9: zwei Tests nebeneinander, nicht hinter einem Ziel; keine Sendung an `slice-mock-abnahme` (dessen DoD führt die 100 Läufe des Mock schon, eine neue Pflicht entsteht nicht, `AGENTS.md` §3.13).
- `LISTEN=127.0.0.1:0` in `SPEC-038` Nr. 3 — **Ausgang:** offen; Adresse Architect. `bootstrap.LadeKonfiguration` weist Port 0 ab (`ODR-E2001 LISTEN: Adresse muss host:port sein, Port 1 bis 65535`); der Test lädt mit `127.0.0.1:8080` und setzt danach `Konfiguration.Listen` auf `127.0.0.1:0`, wie das Hilfsprogramm `uebergang`; die Adresse nennt weiter `Handle.Adresse()`. Der Wortlaut von Nr. 3 ist zu berichtigen; eine Entscheidung im Code fiel nicht an, die Wirkung ist die der Spezifikation.
- Laufzeit der 100 Läufe — **Ausgang:** offen bis zur Closure; Adresse Implementer: gemessen 0,74 bis 0,93 s für den Test, 2,3 s für das Ziel (§4), weit unter 60 s. Schätzung des Architect (nicht gemessen): unter 15 s je Lauf des Ziels (je Lauf neun Requests auf der Schleife, ein Reset in einer Transaktion, zwei Lesevorgänge der Ablage auf `tmpfs`; Vorbereitung wie `TestE2EReplay`), Grenze 120 s (`SPEC-038` Nr. 8 d). Die Schätzung trägt, weil kein Prozessstart je Lauf anfällt; ein Wert über 60 s hieße, dass der Aufbau doch nicht billig ist, und geht an den Architect zurück.
- Einbindung des Ziels in die Gate-Kette oder daneben — **Ausgang:** offen; Adresse User (Entscheidung zu `slice-test-integration` §6 zur Gate-Kette). Der Slice trägt `GATE_CHECKS` nicht ein; der Architect ändert es nicht (`SPEC-038` Nr. 2).
- Formelkonflikt der Größenmarke — **Ausgang:** offen; Adresse User. Nach `SPEC-128` Nr. 1: Tabelle 15 Zeilen, alle `andere` (*b* = 15, *a* = 0), *f* = 0, *p* = 80, also *S* = 80·15 + 200·0 + 80 = **1280** und *M* = 100·⌈135·1280/10000⌉ = 100·18 = **1800**; keine Schnittfrage. Nach `AGENTS.md` §3.14 (Wortlaut: Zeilen × 40, plus 200 je Fake, kein fester Anteil, da kein Leser fremder Bytes; × 1,25, auf Hundert aufgerundet): 15·40 = 600, × 1,25 = 750, Marke **800**. Der Unterschied (1800 gegen 800) kommt aus 80 statt 40 je Zeile, 80 Zeilen Treiber und Fragment und dem Zuschlag 1,35 statt 1,25. Der Architect setzt nichts um; `make groessenmarke` rechnet nach `SPEC-128`, und der Planner hat `basis` (`01c547c`), `schaetzung` 1280 und `marke` 1800 nach `SPEC-128` im Block gesetzt (§3, §4), wie bei `slice-protokolltreue` und `slice-test-integration`; die Formel entscheidet der User, der Block folgt ihr dann.
- Randformen (`AGENTS.md` §3.12, §3.16) — **Ausgang:** entfallen; das Raster steht unten, jede Zelle nennt ihren Ort in `SPEC-038` oder „gilt nicht“ mit Grund.

**Raster Operation × Argumentart** (Nr. = Nummer in `SPEC-038`):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Aufruf des Treibers (Argument) | gilt nicht: der Aufruf ohne Argument ist der Normalfall (Nr. 2) | Nr. 2 (Exit 2) | Nr. 2 (jedes Argument, Exit 2) | Nr. 2, Nr. 8 c (Exit 2) | Nr. 1 (`--rm`, `tmpfs`: kein Zustand zwischen Läufen) |
| Umgebung `E2E_BINARY`, `E2E_UEBERGANG` | Nr. 8 e | Nr. 8 e | Nr. 8 a (Vorbereitung scheitert, `Fatal`) | Nr. 8 c | gilt nicht: ein Lauf je Container (Nr. 1) |
| Vorbereitung (Aufnahme, Import, Handle) | Nr. 8 a | gilt nicht: die Werte bildet der Test fest (Nr. 3) | Nr. 8 a | Nr. 8 a, f | gilt nicht: einmal je Lauf des Ziels (Nr. 3) |
| Reset im Lauf i | gilt nicht: die Kennung stammt aus dem Übergang (Nr. 3, Nr. 8 g) | gilt nicht (Nr. 8 g, `SPEC-129`) | gilt nicht (Nr. 8 g, `SPEC-129`) | Nr. 8 a (Fehler im Lauf) | Nr. 3 a (`Revision` i), Nr. 6 (Stand nach Reset) |
| Request der Folge | gilt nicht: Steuerheader fest gebildet (Nr. 8 g, `SPEC-135` Nr. 5) | gilt nicht (Nr. 8 g) | Nr. 4 (Request 10, nie aufgezeichnet, `404`) | Nr. 8 a (Verbindungsfehler) | Nr. 4 (Requests 6 bis 9, vierter `$count`), Nr. 3 |
| Lesen des Stands | Nr. 6 (keine Cursor: leere Liste, nur mit Kontrollwert) | Nr. 6 (Kontrollwert fängt die leere Liste nach Lauf 1) | Nr. 6 (Kontrollwert, Workspace `""`) | Nr. 8 a (Lesefehler) | Nr. 6 (Lauf 1 gegen jeden Lauf), Nr. 3 (zweite Verbindung je Lesen) |
| Vergleich der Antworten | Nr. 5 (Name nur auf einer Seite) | Nr. 5 (Bytes der Werte; leerer Wert gegen gesetzten) | Nr. 5 (Name außerhalb `VOLATILE_HEADERS` unterscheidet) | gilt nicht: reine Funktion, kein Abbruchpunkt | Nr. 8 b (jeder Lauf gegen Lauf 1, Zeile `Läufe: 100, abweichend: <k>`) |
| Vergleich der Stände | Nr. 6 (Länge der Liste, fehlende Identität) | Nr. 6 (leeres `Consumed` gegen nichtleeres) | Nr. 6 (jedes Feld, zusätzliche Identität) | gilt nicht: reine Funktion | Nr. 8 b |
| Herunterfahren des Handle | gilt nicht: Pflicht jedes Laufs (Nr. 8 f) | gilt nicht (Nr. 8 f) | gilt nicht (Nr. 8 f) | Nr. 8 f (bei jedem Abbruch des Tests) | gilt nicht: einmal (Nr. 8 f) |

Zusatztabelle für den Vergleich, der mehrere Verstöße meldet (`AGENTS.md` §3.16; Nr. 5: der erste in der Reihenfolge Status, Body, Header):

| Paar am selben Request | Meldung | Fänger |
|---|---|---|
| Status und Body verschieden | Status | `TestDeterminismusVergleich`, Schleife über die Paare (Zeile 7 der Tabelle) |
| Status und Header verschieden | Status | dieselbe Schleife |
| Body und Header verschieden | Body | dieselbe Schleife |

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (12×, verkörpert in `AGENTS.md` §3.16; „100 identische Läufe“ und „jede Abweichung ist ein Befund“ sind Allaussagen mit Fänger über alle Fälle); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
