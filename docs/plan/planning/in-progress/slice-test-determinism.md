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

```groessenmarke
pfad: tests/e2e
pfad: harness/mk/test-determinism.mk
pfad: tools/harness/test-determinism.sh
produktcode: 80
fakes: 0
```

Der Block trägt den Teil des Architect: `produktcode` 80 sind Treiber (rund 70 Zeilen, Form von `test-integration.sh`) und Fragment (rund 10); `fakes` 0, weil der Referenzdienst von `tests/e2e` wiederverwendet wird und das Handle kein Fake ist. `basis`, `schaetzung` und `marke` setzt der Planner (§4); die Zahlen nach beiden Formeln stehen in §6.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-test-integration` liegen in `done/` (Replay läuft, Aufbau und Referenzdienst stehen). Der Vertrag `SPEC-038` trägt Umfang und Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice ändert `Makefile` und `harness/mk/` und startet nur, wenn kein anderer Slice der Gate-Kette in `in-progress/` liegt ([Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 7). Geprüft beim Start am 2026-11-06: `slice-bootstrap-m2` und `slice-test-integration` liegen in `done/` (`main` 7a2d4fa); `in-progress/` von `main` trägt außer diesem Slice nichts; die Zweige `adr-gate-laufzeit` und `slice-mutationslauf-werkzeug` tragen in ihrem `in-progress/` keinen Slice (`git worktree list`).

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`), vorläufig, vom Planner gesetzt beim Start am 2026-11-06 aus dem Plan, ohne Tabelle *Zusage · Fänger*: die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade, die der Architect im Block nennt (erwartet `tests/`, `harness/mk/`, `Dockerfile`, `tools/harness/`; Text der Verträge zählt nicht mit), Tests eingeschlossen. Schätzung nach `SPEC-128` Nr. 1 aus der Erwartung des Planners: *a* = 0 nebenläufige Zeilen (100 Läufe nacheinander, ein Prozess zugleich), *b* = 10 andere (Aufbau und Reset je Lauf, Vergleich von Status, Body und Headern, Normalisierung um `VOLATILE_HEADERS`, Vergleich der Cursor-Stände, Zahl der Läufe, Verdrahtung des Ziels, Exit-Code und die Randformen von `SPEC-038`), *f* = 1 (der Referenzdienst von `tests/e2e` wiederverwendet, kein neuer Test-Server erwartet; ein Fake nur, wenn der Architect einen braucht), *p* = 100 (Fragment, Treiber, Stufe): *S* = 85·0 + 80·10 + 200·1 + 100 = **1100**, *M* = **1500**. *S* liegt unter 2223: keine Schnittfrage. Das Ziel nutzt `tests/e2e` von `slice-test-integration` (Aufbau, Referenzdienst); trägt der Reset einen Eingang, der Produktcode verlangt, steigt *p*, und der Planner rechnet neu. Den Block `groessenmarke` und die Tabelle trägt der Plan erst, wenn der Architect sie liefert; bis dahin ist die Marke in dieser Prosa-Form vorläufig und `make groessenmarke` misst sie nicht. Beim Setzen des Blocks rechnet der Planner *S* und *M* neu und ersetzt diese Zahlen mit Grund (`ersetzt:`).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): Die Marke setzt der Planner beim Start nach `AGENTS.md` §3.14 (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Zahl steht erst mit der Tabelle *Zusage · Fänger* des Architect. Der Planner schätzt den Slice als klein.
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
- Laufzeit der 100 Läufe — **Ausgang:** offen; Adresse Implementer, misst vor der Übergabe und trägt den Wert in §4 ein. Schätzung des Architect (nicht gemessen): unter 15 s je Lauf des Ziels (je Lauf neun Requests auf der Schleife, ein Reset in einer Transaktion, zwei Lesevorgänge der Ablage auf `tmpfs`; Vorbereitung wie `TestE2EReplay`), Grenze 120 s (`SPEC-038` Nr. 8 d). Die Schätzung trägt, weil kein Prozessstart je Lauf anfällt; ein Wert über 60 s hieße, dass der Aufbau doch nicht billig ist, und geht an den Architect zurück.
- Einbindung des Ziels in die Gate-Kette oder daneben — **Ausgang:** offen; Adresse User (Entscheidung zu `slice-test-integration` §6 zur Gate-Kette). Der Slice trägt `GATE_CHECKS` nicht ein; der Architect ändert es nicht (`SPEC-038` Nr. 2).
- Formelkonflikt der Größenmarke — **Ausgang:** offen; Adresse User. Nach `SPEC-128` Nr. 1: Tabelle 15 Zeilen, alle `andere` (*b* = 15, *a* = 0), *f* = 0, *p* = 80, also *S* = 80·15 + 200·0 + 80 = **1280** und *M* = 100·⌈135·1280/10000⌉ = 100·18 = **1800**; keine Schnittfrage. Nach `AGENTS.md` §3.14 (Wortlaut: Zeilen × 40, plus 200 je Fake, kein fester Anteil, da kein Leser fremder Bytes; × 1,25, auf Hundert aufgerundet): 15·40 = 600, × 1,25 = 750, Marke **800**. Der Unterschied (1800 gegen 800) kommt aus 80 statt 40 je Zeile, 80 Zeilen Treiber und Fragment und dem Zuschlag 1,35 statt 1,25. Der Architect setzt nichts um; `make groessenmarke` rechnet nach `SPEC-128`, und der Planner setzt `basis`, `schaetzung`, `marke` im Block (§3), wenn der User die Formel entschieden hat oder `SPEC-128` gilt.
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
