# Slice slice-cli-adapter: CLI-Driving-Adapter — Kommandoauswahl, Signale und Exit-Codes

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-02`, `LH-RB-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0007](../../adr/0007-race-detektor-im-gate.md)

**Berührte Spec-Stellen:** `SPEC-138` (neu, Vertrag dieses Slice: Kommandos, Signale, Exit-Codes) · `ARC-006` (CLI Driving Adapter: Kommandoauswahl, Signale, Exit-Code) · `ARC-012` (Composition Root) · `LH-FA-02.c` (Beenden des Prozesses: SIGTERM/SIGINT) · `LH-RB-03.a` (das Binary ist auch nativ unter Linux startbar) · `SPEC-110` (Startkonfiguration: es gibt keine Kommandozeilen-Optionen, die Konfiguration kommt aus der Umgebung) · `SPEC-059` (`SHUTDOWN_TIMEOUT`) · `architecture.md` §5 (Zeile *Prozess-Signal*).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-cli-adapter)

**Autor:** pt9912. **Datum:** 2026-10-31.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Driving Adapter `src/adapters/driving/cli` (`ARC-006`) wählt das Kommando aus den Argumenten (`version` und `start`, `SPEC-138` Nr. 2; der Bestand von `version` liegt in `bootstrap.Run`), übersetzt SIGTERM und SIGINT in den Abbruch eines Kontexts, mit dem der Composition Root den geordneten Shutdown nach `LH-FA-02.c` auslöst, und bildet das Ergebnis auf den Exit-Code des Prozesses ab; `src/bootstrap/main` ruft den Adapter. Der Prozess hat damit seinen Eingang, bevor `slice-bootstrap-m2` Listener und Dienste dahinter hängt. Der Start ist ein Funktionswert `Konfig.Start`, den der Composition Root setzt; `main` setzt ihn in diesem Slice nicht, `start` ist bis `slice-bootstrap-m2` ein unbekanntes Kommando (Exit-Code 2) und kein Kommando ohne Wirkung (`LH-RB-01`).

**Übernimmt:** aus `slice-bootstrap-m2` die Zeile `src/bootstrap/main/` *Kommando zum Start* (Kommandoauswahl, Exit-Codes; Sendung von `slice-fehlermodell-konfiguration`, `AGENTS.md` §3.13); `slice-bootstrap-m2` führt den Punkt in §1 *Übernimmt* und im Start-Trigger nach diesem Slice (im selben Commit wie dieser Plan nachgezogen). Das Laden der Konfiguration (`LadeOderMelde`) bleibt dort.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Kommandos über Start und Version hinaus — entfallen (Entscheidung des Users, 2026-10-30, [welle-record-replay](../welle-record-replay.md) §6): `LH-FA-13` verlangt eine Verwaltungsschnittstelle, keine Kommandos, und die Spezifikation nennt keine Kommandozeilen-Optionen (`SPEC-110`).
- Hilfetext und Optionen (`--help`, `-h`, `--version`) — entfallen (Entscheidung des Users, 2026-10-30, [welle-record-replay](../welle-record-replay.md) §6; `SPEC-110`: keine Kommandozeilen-Optionen); `--help` ist ein Aufrufverstoß mit Exit-Code 2 (`SPEC-138` Nr. 4).
- Listener öffnen und schließen, 503 während des Herunterfahrens — `slice-http-adapter` (zwei getrennte Schritte); dieser Slice liefert nur den Auslöser.
- Verdrahtung von Konfiguration, Diensten und `CloseAll` — `slice-bootstrap-m2`; der Adapter kennt weder Dienste noch Listener (`architecture.md` §2: Driving Adapter importieren nur Modell und Ports).
- Der Admin-Listener — `slice-admin-listener-token` ([welle-admin-api](../welle-admin-api.md)).

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

- [x] Kommandoauswahl und Exit-Codes im Adapter `src/adapters/driving/cli` nach `SPEC-138` Nr. 1 bis 5 und 9; `bootstrap.Run` entfällt (`bootstrap.go` behält Paketkommentar und `exitOK`), `src/bootstrap/main` ruft `cli.Run`.
- [x] Signale nach `SPEC-138` Nr. 6, 7: Anmeldung vor dem Aufruf, erstes Signal bricht den Kontext, zweites erzwingt mit 130/143, Abmeldung genau einmal; die Randformen stehen als Raster in §6 und sind vor dem Code in der Spezifikation entschieden (`AGENTS.md` §3.12, §3.16).
- [x] `make gates` grün, darunter `make a-check` (der Adapter importiert nur die Standardbibliothek, `SPEC-138` Nr. 1) und `make test-race` (die Signalbehandlung ist nebenläufig, `AGENTS.md` §3.10).
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist (Aufrufmuster, Exit-Codes, Signalverhalten).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/adapters/driving/cli/` | neu | Kommandoauswahl, Signal in Kontextabbruch, Exit-Code (`ARC-006`) |
| `src/bootstrap/main/` | update | ruft den Adapter statt `bootstrap.Run` |
| `src/bootstrap/bootstrap.go` | update | `Run`, `exitAufruf`, `kommandoName` und Importe entfallen; Paketkommentar angepasst, `exitOK` bleibt (`konfig.go` nutzt es); die Datei bleibt, weil Fälle der Gegenproben (`lint/direktive-nolint`) Text an sie anhängen und `slice-bootstrap-m2` sie füllt |
| `src/bootstrap/bootstrap_test.go` | gelöscht | `TestRun`, `TestRunNilWriter` ziehen in den Adapter |
| `spec/spezifikation.md` | update | `SPEC-138` (Vertrag zu Aufrufmustern, Exit-Codes und Signalverhalten), `SPEC-110` (zwei Zusätze) |
| `harness/vertraege/SPEC-036-test.md` | update | Randformen des Aufrufs verweisen auf `SPEC-138` Nr. 2 bis 4 |
| `spec/architecture.md` | update | Zeile von `ARC-006` |
| Testdateien neben dem Adapter | neu | Happy Path, Boundary und Negative je Aufrufmuster; Signal-Tests ordnen mit Kanälen (`AGENTS.md` §3.10); externes Testpaket `cli_test` (Lint-Regel `testpackage`) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag `SPEC-138` **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen, nennt §6.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15). Die Tests stehen in `src/adapters/driving/cli/` im externen Testpaket `cli_test`. Fakes: `FakeStart` und `FakeSignale` (siehe unten).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| C1 | Nr. 2: kein Argument (`nil`, leere Liste) und `version` geben die Version aus; `start` führt `Start` aus, wenn gesetzt, sonst ist es unbekannt (Menge W: Aufrufmuster {`nil`, leere Liste, `version`, `start`} × Start {gesetzt, `nil`} = 8) | `TestKommandoWahl`: je Fall Exit-Code, stdout, Zahl der Aufrufe von `FakeStart` | ohne Argument ruft `Start`; `start` wird nicht erkannt; `start` bei `nil` ruft `Start` (nil-Aufruf); `version` ruft `Start`: je eine rot | andere |
| C2 | Nr. 4: jedes andere Aufrufmuster ergibt Exit-Code 2, nichts nach stdout, genau `Unbekanntes Kommando: <args[0]>` und Zeilenende nach stderr, `Start` nicht gerufen, keine Anmeldung. Menge A × P: 24 Formen × 4 Positionen = 96, dazu 4 Sonderfälle = 100. Die 24 Formen A: `foo`, leer, ein Leerzeichen, Tab, `Start`, `START`, `Version`, `VERSION`, `start` plus Leerzeichen, Leerzeichen plus `start`, `version` plus Zeilenvorschub, `starts`, `sta`, `versio`, `versionx`, `--help`, `-h`, `help`, `--version`, `-v`, `--`, `version` plus U+00A0, U+FEFF plus `start`, `version` plus NUL (Rohbytes aus Bytewerten gebaut, `AGENTS.md` §3.18). Die 4 Positionen P: [f], [f, `x`], [`version`, f], [`start`, f] (Start gesetzt). Die 4 Sonderfälle: [`start`] mit Start `nil`, [`start`, `x`] mit Start `nil`, [`version`, `x`, `y`], [`start`, `x`, `y`] | `TestAblehnung`: Schleife über A × P und die Sonderfälle; die erwartete Meldung wird aus `args[0]` gebaut. Dazu `TestAblehnungKlasse` als Generator mit demselben Orakel über die Klasse „Rest vor und hinter dem Kommando“: je `start` und `version` jedes Byte 0x00 bis 0xFF davor, dahinter und als Zusatzargument (2 · 3 · 256 = 1536), jede Schreibweise mit einem umgeschalteten Buchstaben (5 + 7 = 12) und jedes echte Präfix (4 + 6 = 10): 1558 Elemente, Zählprüfung im Test | `strings.EqualFold` (rot bei `Start`, `VERSION`); Leerraum entfernt (rot bei `start ` und `version`+LF); `HasPrefix` (rot bei `starts`, `versionx`); nur `args[0]` geprüft (rot bei P2 bis P4); Meldung nennt `args[1]` (rot); Meldung nach stdout; Exit-Code 1; ohne Zeilenende; `Start` trotz Verstoß gerufen; Anmeldung vor der Prüfung (Zähler 1): je eine rot | andere |
| C3 | Nr. 3: `odata-recorder <Version>` und Zeilenende auf stdout, `Version` unverändert, nichts auf stderr, Exit-Code 0, keine Anmeldung. Menge: `Version` {`1.2.3`, `dev`, leer, ` x ` mit Randleerzeichen, `a`+Tab+`b`} × {`nil`, `version`} = 10 | `TestVersionAusgabe`; dazu `TestVersionKlasse` als Generator über die Klasse „Version unverändert“ (Orakel: stdout gleich `odata-recorder ` + Version + Zeilenende): jedes Byte 0x00 bis 0xFF allein, davor, dahinter und in der Mitte (1024), 5 Mehrbyte-Folgen, 3 Längen (100, 1000, 100000), 3 Formatfolgen: 1035 Elemente, Zählprüfung im Test | `TrimSpace` der Version (rot bei ` x `); leere Version durch `dev` ersetzt (rot); Ausgabe nach stderr; ohne Zeilenende; ohne Präfix; Anmeldung bei `version`: je eine rot | andere |
| C4 | Nr. 1: ein nil-Writer verwirft die Ausgabe, `Start` erhält nie `nil`, ein Schreibfehler ändert den Code nicht. Menge Z: Writer {nil, schreibt mit Fehler} × Seite {stdout, stderr, beide} × Kommando {`version`, Verstoß `foo`, `start`} = 18 | `TestWriterRandformen`: Code gleich dem Fall mit gutem Writer (0, 2, 0); `FakeStart` hält die Writer und prüft ungleich nil | nil-Writer nicht ersetzt (Panic); nil an `Start` weitergereicht; Schreibfehler ergibt Exit-Code 1: je eine rot | andere |
| C5 | Nr. 6: `Start` wird genau einmal gerufen, mit den Writern aus `Konfig` (Identität, zwei verschiedene Puffer) und einem abbrechbaren, noch nicht abgebrochenen Kontext; der Adapter schreibt selbst nichts. Menge: 4 Eigenschaften (Zahl der Aufrufe, Identität stdout/stderr, `ctx.Done()` ungleich nil und `ctx.Err()` gleich nil beim Aufruf, Adapter-Ausgabe leer) | `TestStartAufruf` mit `FakeStart` (hält `ctx` und beide Writer, zählt Aufrufe) | zweimal gerufen; Writer vertauscht; `context.Background()` (`Done()` gleich nil); Kontext schon abgebrochen; Adapter schreibt `gestartet` nach stdout: je eine rot | andere |
| C6 | Nr. 5: Rückgabe 0 → 0, jede andere → 1. Menge: Rückgaben {0, 1, 2, 3, 255, 256, 257, -1, MaxInt, MinInt} = 10 (ohne Signal) | `TestExitAbbildung`; dazu `TestExitAbbildungKlasse` als Generator mit dem Orakel „0 bleibt 0, sonst 1“ über -1024 bis 1024 (2049), je Zweierpotenz 2^k für k = 11 bis 62 die Werte ±(2^k-1), ±2^k, ±(2^k+1) (312) und MaxInt, MaxInt-1, MinInt, MinInt+1 (4): 2365 Elemente, Zählprüfung im Test | Durchreichen (rot bei 2, 3, 255, 256, 257); `> 0` statt `!= 0` (rot bei -1, MinInt); `& 0xFF` (rot bei 256); `== 1` (rot bei 2) | andere |
| C7 | Nr. 7 a: Die Anmeldung geschieht genau einmal je `start`, vor dem Aufruf von `Start`, und sonst nie. Menge: Pfade {`start`, ohne Argument, `version`, Verstoß, `start` bei `nil`} = 5, erwartet Anmeldungen {1, 0, 0, 0, 0}; `FakeStart` liest im Aufruf den Zähler von `FakeSignale` | `TestAnmeldung` | Anmeldung auch bei `version` oder Verstoß; Anmeldung nach dem Aufruf von `Start` (Start sieht 0); zweimal angemeldet: je eine rot | nebenläufig |
| C8 | Nr. 6, 7 a, b: Das erste Signal bricht den Kontext ab, genau einmal, und nicht davor; `Start` kehrt mit 0 zurück, `Run` liefert 0 (kein Erzwingen). Menge: Quelle {Fake-Kanal, echtes Signal an den Prozess} × Signal {SIGTERM, SIGINT} × Zeitpunkt {vor dem Aufruf von `Start` vorbelegt, im Lauf}: Fake 4, echt nur „im Lauf“ 2 = 6; im Test um das erste Signal SIGHUP als fremdes (Fake, vorbelegt und im Lauf) auf 8 erweitert, weil die Quelle jedes Signal liefern darf und ein erstes SIGHUP sonst ungefangen bliebe. Echte Signale: Der Test registriert vorher einen eigenen Kanal (`signal.Notify`), damit ein fehlender Fänger den Testprozess nicht beendet | `TestErstesSignal`: `FakeStart` meldet „bereit“, prüft `ctx.Err()` gleich nil, wartet auf `ctx.Done()`, liefert 0; Wartezeiten nur als `select` mit Zeitgrenze | Signal ignoriert (Zeitgrenze, rot); Kontext von Anfang an abgebrochen (`Err()` ungleich nil vor dem Signal); Erzwingen schon beim ersten Signal (Run liefert 143/130); SIGINT nicht angemeldet; SIGTERM nicht angemeldet (nur die zwei echten Fälle fangen die Anmeldung) | nebenläufig |
| C9 | Nr. 7 b: Nach dem ersten Signal wartet `Run` auf `Start` und liefert dessen Code (Nr. 5). Menge: Rückgabe nach dem Signal {0, 1, 2} × erstes Signal {SIGTERM, SIGINT} = 6; `FakeStart` bleibt nach `ctx.Done()` blockiert, bis der Test es freigibt | `TestWartetAufStart` | `Run` kehrt nach dem Abbruch sofort zurück (rot bei 1, 2: Code ungleich); Ergebnis über eine gemeinsame Variable statt über den Kanal (Datenrennen: `make test-race`, je Prozess) | nebenläufig |
| C10 | Nr. 5, 7 c: Das zweite Signal vor der Rückkehr von `Start` erzwingt das Ende mit 130 (zweites SIGINT) und sonst 143, ohne Warten, ohne Ausgabe. Menge: erstes Signal {SIGTERM, SIGINT} × zweites {SIGTERM, SIGINT, SIGHUP als fremdes} × Quelle: Fake 6, echt (ohne SIGHUP) 4 = 10; `FakeStart` bleibt blockiert, der Cleanup gibt es frei | `TestZweitesSignal` | Erzwingen fehlt (Zeitgrenze, rot); Code des ersten Signals; immer 143 (rot bei zweitem SIGINT); immer 130 (rot bei zweitem SIGTERM); Erzwingen erst beim dritten Signal; Ausgabe beim Erzwingen: je eine rot | nebenläufig |
| C11 | Nr. 7 d: Die Abmeldung geschieht genau einmal auf jedem Pfad nach der Anmeldung, auch beim Erzwingen, und nie vor der Rückkehr von `Start`. Menge: Pfade {`Start` liefert 0, 1, 2; erstes Signal dann 0; erstes Signal dann 1; Erzwingen} = 6 mit Abmeldungen gleich 1, dazu {`version`, Verstoß, `start` bei `nil`} = 3 mit Abmeldungen gleich 0; `FakeStart` liest im Lauf den Zähler (muss 0 sein) | `TestAbmeldung` | Abmeldung fehlt; doppelt; vor der Rückkehr von `Start`; nicht auf dem Pfad des Erzwingens: je eine rot | nebenläufig |
| C12 | Nr. 9: kein Zustand zwischen Aufrufen. Menge: Paare nacheinander im selben Prozess {Erzwingen, dann normales `start`}, {erstes Signal, dann normales `start`}, {normal, dann Signal}, {Erzwingen, dann Erzwingen} = 4 | `TestZweiAufrufe`: Der zweite Aufruf sieht `ctx.Err()` gleich nil, Anmeldung gleich 1, sein eigenes Ergebnis | Kontext, Kanal oder Merker „erzwungen“ auf Paketebene: Der zweite Lauf sieht den abgebrochenen Kontext oder hängt: rot | nebenläufig |
| C13 | Nr. 1: Das Paket importiert nur die Standardbibliothek (kein Segment mit Punkt im ersten Pfadteil). Menge: alle Importe der Nicht-Test-Dateien (heute etwa 6) | `TestNurStandardbibliothek` (`go/build.ImportDir`, Schleife über `Imports`) | Import von `src/bootstrap` (etwa für `ExitKonfiguration`) oder von `src/hexagon/model`: rot (strenger als `a-check`, das `model` erlaubte) | andere |
| C15 | Nr. 1, 7 d (aus dem Review): Die Abmeldung der Laufzeit-Quelle nimmt die Anmeldung zurück. Wirkung: Nach der Rückkehr von `Run` beendet ein SIGTERM oder SIGINT den Prozess durch das Signal (Menge: {SIGTERM, SIGINT} = 2) | `TestLaufzeitAbmeldung`: der Testprozess startet sich erneut (`os.Executable`, `-test.run`); das Kind ruft `Run` mit der Quelle `nil` und sendet danach das Signal an sich; erwartet: Ende durch das Signal, nicht Exit-Code 92 | `signal.Stop` fehlt (`_ = kanal`); Abmeldung leer: je rot | nebenläufig |
| C16 | Nr. 1 (aus dem Review): Der Kanal der Laufzeit-Quelle hat Pufferplatz; zwei verschiedene Signale unmittelbar nacheinander erzwingen das Ende (Menge: (SIGTERM, SIGINT), (SIGINT, SIGTERM) = 2, echte Quelle; Code 130 oder 143, weil der Kernel gleiche Signale zusammenfasst und die Laufzeit die Reihenfolge bestimmt) | `TestSignaleHintereinander` | Kanal ohne Puffer (`make(chan os.Signal)`): rot | nebenläufig |
| C14 | Akzeptierte Negative, je mit Grund und Folge in *Ungeprüft* von `SPEC-138` (die Posten des Implementers unten, die Kanalgröße und die Randform der Signalquelle stehen dort erst nach dem Nachzug des Architect, §6): Fenster vor der Anmeldung und Signal bei `version`/Verstoß; zweites Signal zugleich mit der Rückkehr (auch beide vorbelegt); Panic in `Start`; andere Signale als SIGTERM/SIGINT; Goroutine nach dem Erzwingen; Steuerzeichen am Terminal; Verdrahtung des Start-Pfads in `main` (nur `pruefung build` mit `version` fängt `main`); vom Implementer ergänzt (offen, Adresse Architect, siehe §6): „genau einmal“ in Nr. 7 b ist am Kontext nicht beobachtbar (ein zweiter Abbruch ist wirkungslos), und die Vorabprüfung des vorbelegten Signals (Nr. 7 a) ist von außen nicht zu trennen vom Abbruch durch die Schleife (Mutation ohne Vorabprüfung: 0 von 50 Läufen rot) | benannt, kein Test | nicht anwendbar | andere |

**Fixtures und Mutationen.** Keine Fixture gleicht dem Wert der Mutation:
- Die Rückgaben von `Start` (C6, C9) liegen außerhalb {143, 130, 1}.
- Das zweite Signal im Paar (SIGTERM, SIGINT) ergibt 130, im Paar (SIGINT, SIGTERM) 143.
- `Version` ` x ` hat Randleerzeichen.

**Raten (§3.10).** Alle Mutationen von C7 bis C12 sind durch Kanäle und blockierte Fakes geordnet. Sie müssen **in jedem Lauf** rot werden. Der Implementer misst:
- je Mutation `-count=50` mit Zahl der Läufe;
- die Mutation des gemeinsamen Zustands in C9 je Prozess mit `make test-race` (`-test.count=1`, mindestens 20 Prozesse);
- alles mit Zahl, nicht „hoch“.

Die Zusagen halten keine Sperre; die Zeilen *je Fenster der Sperre* aus §3.10 entfallen.

**Gemessen (Implementer, Mutanten in Kopien des Quellstands, Image `odata-recorder:race`, `--memory=4g --cpus=4`, ohne Netz).** Die Mutation steht in der Spalte *Mutation* der Tabelle; hier der rote Test und die Rate. C1 bis C6 und C13 sind nicht nebenläufig: ein Lauf je Mutation, rot. Die Fänger von C7 bis C12 liefen mit `-count=50` in einem Prozess, die Zeitgrenze `frist` der Wartepunkte auf 100 ms gesetzt (das ungemutierte Paket bleibt so in 50 von 50 Läufen grün, in `-race` einmal grün).

| Zusage | Mutation (Auswahl) | roter Test | Rate |
|---|---|---|---|
| C1 | ohne Argument ruft `Start`; `start` unbekannt; `start` bei `nil` ruft `Start`; `version` ruft `Start` | `TestKommandoWahl` (`start` bei `nil`: Panic des Fakes, Testprozess rot) | 4 von 4 |
| C2 | `EqualFold`; `TrimSpace`; `HasPrefix`; nur `args[0]` geprüft; Meldung nennt das letzte Argument, nach stdout, ohne Zeilenende; Exit-Code 1; `Start` trotz Verstoß; Anmeldung vor der Prüfung; **nicht genannt:** Suffix `\r` abgeschnitten, `start x` angenommen | `TestAblehnung`; die zwei nicht genannten fängt allein `TestAblehnungKlasse` (die 100er-Menge ließ sie grün) | 12 von 12 |
| C3 | `TrimSpace` der Version; leere Version zu `dev`; stderr; ohne Zeilenende; ohne Präfix; Anmeldung bei `version`; **nicht genannt:** `TrimRight` von `\n\r`, NUL entfernt, `ToValidUTF8`, Kürzung auf 255 Zeichen, `%` entfernt | `TestVersionAusgabe`; die fünf nicht genannten fängt allein `TestVersionKlasse` (50 von 50 Läufen) | 11 von 11 |
| C4 | nil-Writer nicht ersetzt (stdout, stderr); nil an `Start`; Schreibfehler ergibt 1 (Version, Verstoß) | `TestWriterRandformen` (nil-Writer: Panic, Testprozess rot) | 5 von 5 |
| C5 | zweimal gerufen; Writer vertauscht; `context.Background()`; Kontext schon abgebrochen; Adapter schreibt `gestartet` | `TestStartAufruf` (je ein Teiltest) | 5 von 5 |
| C6 | Durchreichen; `<= 0`; `& 0xFF`; `!= 1`; **nicht genannt:** 128 und Bit 40 gesetzt ergeben 0 | `TestExitAbbildung`; 128 fängt nur `TestExitAbbildungKlasse` | 6 von 6 |
| C7 | Anmeldung bei Verstoß, bei Zusatz, bei `start` mit `nil`; zweimal; nach dem Aufruf von `Start` (`Verzoegerung`) | `TestAnmeldung`, `TestAblehnung` | 5 Mutationen je 50 von 50 |
| C8 | Signal ignoriert (vorbelegt und im Lauf); Kontext von Anfang an abgebrochen; Erzwingen beim ersten Signal; SIGINT, SIGTERM nicht angemeldet; **nicht genannt:** erstes SIGHUP ignoriert | `TestErstesSignal` | 7 Mutationen je 50 von 50 |
| C8 (vorbelegt) | Vorabprüfung des vorbelegten Signals entfernt (die Schleife bricht ab) | keiner | **0 von 50** (benanntes Negativ in C14) |
| C9 | `Run` kehrt nach dem Abbruch zurück | `TestWartetAufStart` | 50 von 50 |
| C9 (Datenrennen) | Ergebnis über gemeinsame Variable statt Kanal | `make test-race`-Detektor, jeder Prozess `-race -test.count=1` | 30 von 30 Prozessen (30 mit Bericht `DATA RACE`) |
| C10 | Erzwingen fehlt; Code des ersten Signals; immer 143; immer 130; erst beim dritten Signal; Ausgabe beim Erzwingen | `TestZweitesSignal` | 6 Mutationen je 50 von 50 |
| C11 | Abmeldung fehlt; doppelt; nach dem ersten Signal vor der Rückkehr; nach dem Start ohne Signal; nur auf dem Normalpfad | `TestAbmeldung` | 5 Mutationen je 50 von 50 |
| C12 | Paketzustand „erzwungen“; Paketmerker „erstes Signal“ | `TestZweiAufrufe` | 2 Mutationen je 50 von 50 |
| C13 | Import von `src/bootstrap`; von `src/hexagon/model` | `TestNurStandardbibliothek` | 2 von 2 |
| C15 | `signal.Stop` der Laufzeit-Quelle fehlt (`_ = kanal`); Abmeldung leer | `TestLaufzeitAbmeldung` (Kind endet mit Exit-Code 92 statt durch das Signal) | 2 Mutationen je 50 von 50 |
| C16 | Kanal der Laufzeit-Quelle ohne Puffer | `TestSignaleHintereinander` | 50 von 50 |
| C16 (Größe) | Kanalgröße 1 statt 2 | keiner | **0 von 50** (benanntes Negativ, siehe unten) |

Das ungemutierte Paket: 30 von 30 Prozessen grün unter `-race -test.count=1`; mit den Fängern C15, C16 und `TestVersionKlasse` 50 von 50 Läufen grün bei `frist` 5 s und einmal grün unter `-race`.

**Akzeptierte Negative aus dem Review** (`AGENTS.md` §3.15):
- *Kanalgröße genau 2* (`SPEC-138` Nr. 1). Grund: Die Schleife empfängt ständig; ein Puffer von 1 hält dasselbe Paar Signale (eines in der Hand der Schleife, eines im Puffer), und ein drittes Signal fasst der Kernel mit dem zweiten gleicher Art zusammen. Kein Lauf unterscheidet 1 von 2 (Mutation: 0 von 50). Folge im Betrieb: keine, solange die Schleife empfängt; fällt ein Signal in die kurze Zeit der Verarbeitung des ersten, bleibt es im Puffer. Der Puffer 0 ist gefangen (C16).
- *Vorabprüfung des vorbelegten Signals* (`SPEC-138` Nr. 7 a; Mutation ohne sie: 0 von 50, auch 0 von 50 je `-cpu 1, 2, 4, 20` beim Reviewer). Grund: Ob die Goroutine von `Start` vor oder nach dem Abbruch durch die Schleife beginnt, bestimmt die Laufzeit. Folge im Betrieb: Ein Signal, das zwischen Anmeldung und Aufruf eintrifft, kann `Start` einen Augenblick mit nicht abgebrochenem Kontext starten lassen; `Start` sieht den Abbruch spätestens beim ersten Prüfen (`SPEC-138` Nr. 8 c), vor dem ersten Startschritt, der etwas öffnet.

- *Mehrbytige Zeichen in `Version`* (`TestVersionKlasse`; Verifikation V-3, `AGENTS.md` §3.17). Prüfgrenze: Die Klasse „jedes Zeichen“ ist unbeschränkt (rund 1,1 Millionen Codepunkte); der Fänger führt jedes Byte 0x00 bis 0xFF in vier Positionen und fünf mehrbytige Folgen. Grund: Das genügt für jede bytebasierte oder generische Mutation an der Version (`TrimX`, `ToValidUTF8`, Kürzen, Entfernen von Steuerzeichen, `ToLower`, Abschneiden ab 0x1f; vom Verifier an vier nicht genannten Elementen rot gesehen). Ungefangen bleibt eine Mutation, die genau ein mehrbytiges Zeichen entfernt (U+2028, `e2 80 a8`: Mutation grün). Folge im Betrieb: ein auf ein einziges Zeichen zielender Eingriff in die Ausgabe von `version` bliebe unbemerkt; der Code behandelt `Version` als undurchsichtigen Wert und kennt keine Zeichenklasse.

**Hinweise für den Implementer** (kein Produktionscode):

- **Gerüst von `Run`.** Die Auswahl steht in einer kleinen Funktion; `start` läuft so:
  - Anmelden, Abmelden per `defer`. Ein schon vorbelegtes Signal nimmt eine Vorabprüfung ohne Warten aus dem Kanal und bricht den Kontext ab, bevor `Start` läuft (`SPEC-138` Nr. 7 a). `Start` läuft dann in einer Goroutine und schreibt sein Ergebnis in einen Kanal der Größe 1, damit sie nach dem Erzwingen nicht blockiert.
  - Eine Schleife mit einem `select` über Ergebnis und Signal und dem Merker *erstes Signal gesehen*. Beim ersten Signal bricht `cancel()` den Kontext ab; bei jedem weiteren liefert die Schleife 130 bei SIGINT und sonst 143. `cyclop` liegt bei 15 und `funlen` bei 100.
- **Fakes.** `FakeStart` hält `ctx`, beide Writer und den Zähler der Aufrufe. Es blockiert auf `ctx.Done()` oder auf einen Kanal und liefert einen konfigurierten Wert. `FakeSignale` zählt Anmeldungen und Abmeldungen und liefert den Kanal; ihr Feld `Verzoegerung` hält die Anmeldung auf, damit ein `Start` neben der Anmeldung den Zähler noch auf 0 sieht (C7). `FakeStart` liest bei Aufruf und Rückkehr die Zähler von `FakeSignale` (C7, C11) und zählt auch die Rückkehren. Beide zählen jede Operation (§3.15).
- **Bootstrap.** `exitOK` bleibt in `bootstrap.go`, weil `konfig.go` es nutzt. Der Paketkommentar lautet etwa: *„Package bootstrap ist der Composition Root des OData Recorder: Startkonfiguration (`SPEC-110`) und Verdrahtung.“*
- **Lint.** `forbidigo` verbietet `os.Stdout`/`os.Stderr` außerhalb von `src/bootstrap`. `main` setzt `Stdout`/`Stderr`; im Paket `cli` steht keiner. `gochecknoglobals` verbietet Variablen auf Paketebene, nur Konstanten sind erlaubt.
- **Echte Signale.** `syscall.Kill(syscall.Getpid(), sig)` erst nach der Meldung „bereit“ von `FakeStart`. Dann ist die Anmeldung aktiv. Der Cleanup des eigenen Kanals meldet erst nach einer Frist ab, damit ein gesendetes, noch nicht zugestelltes Signal nicht nach der Abmeldung den Testprozess beendet. Die Tests laufen nicht parallel.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-fehlermodell-konfiguration` liegt in `done/` (erfüllt: `LadeOderMelde`). Die Spezifikation trägt den Vertrag (Aufrufmuster, Exit-Codes, Signalverhalten) samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12); Raster (§6), Tabelle (§3) und `SPEC-138` stehen im Commit des Architect. Reihenfolge: vor `slice-bootstrap-m2`, dessen Start-Trigger diesen Slice nennt; beide ändern `src/bootstrap/main/` und `src/bootstrap/bootstrap.go` und laufen deshalb nacheinander (Linie A, [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 3).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke aus dem Block unten (Metrik nach `AGENTS.md` §3.14: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die `pfad`-Zeilen des Blocks, Tests eingeschlossen; Formel und Prüfung `SPEC-128`). Die Schätzung steht im Block, nachgerechnet aus der Tabelle in §3 (*Zusage · Fänger*): 6 nebenläufige Zeilen (C7 bis C12) und 8 andere (C1 bis C6, C13, C14), 2 Fakes (`FakeStart`, `FakeSignale`; der Fehler-Writer aus C4 ist ein Hilfstyp von wenigen Zeilen und zählt nicht), Produktcode 230 (`cli.go` etwa 120, `main.go` 10, `bootstrap.go` 30 Löschungen, `bootstrap_test.go` 55 Löschungen, Rest 15). *S* = 85·6 + 80·8 + 200·2 + 230 = 1780, Marke 100 · ⌈135·1780 / 10000⌉ = 2500. Die Schnittfrage entsteht erst ab *S* ≥ 2223; der Slice bleibt ungeschnitten.
- `in-progress` → `open` (blockiert): der Adapter erreicht den Start nicht ohne Import von `src/bootstrap` (entschieden: `Konfig.Start` ist ein Funktionswert, siehe §6).

```groessenmarke
pfad: src/adapters/driving/cli
pfad: src/bootstrap/main
pfad: src/bootstrap/bootstrap.go
pfad: src/bootstrap/bootstrap_test.go
produktcode: 230
fakes: 2
basis: 4e14497
schaetzung: 1780
marke: 2500
```

**Marke** (`AGENTS.md` §3.14): gesetzt am 2026-10-31 vom Planner in eigenem Commit vor dem ersten Code-Commit; `basis` ist `4e14497`, der Commit mit `SPEC-138` und Plan (Vertrag `SPEC-128`, Formel *S* = 85·*a* + 80·*b* + 200·*f* + *p*, Marke = 100 · ⌈135·*S* / 10000⌉). Nachgerechnet aus der Tabelle in §3: 14 Zeilen, davon 6 `nebenläufig` (C7 bis C12) und 8 `andere` (C1 bis C6, C13, C14); das stimmt mit der Zählung des Architect überein. Nach dem ersten Code-Commit ist die Marke fest; eine Abweichung der Schätzung von der Tabelle ist dann nur ein Hinweis (`SPEC-128` Nr. 5 c).

**Gemessen** (Implementer, `make groessenmarke` am Commit c6bd5c4, nach der Fix-Runde des Reviews; beim Handoff am Commit e03912b waren es 1182): 1296 Zeilen von der Marke 2500; die Marke bleibt, `ueberschreitung` entfällt. Die Tabelle hat mit C15 und C16 jetzt 16 Zeilen, das Gate meldet dazu den Hinweis *Tabelle ergibt 1950, Schätzung 1780* (`SPEC-128` Nr. 5 c). Verteilung: `cli.go` 121, `fakes_test.go` 245 (zwei Fakes, Schätzung 400), `kommando_test.go` 368, `signale_test.go` 470, `bootstrap.go` 39, `bootstrap_test.go` 49, `main.go` 4.

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

- Wie erreicht der Adapter den Start, ohne `src/bootstrap` zu importieren? `a-check` erlaubt einem Driving Adapter nur Modell, Inbound Ports und Querschnitts-Ports; der Composition Root importiert alle. — **Ausgang:** entschieden (`SPEC-138` Nr. 1, 6, 8). `Konfig.Start` ist ein Funktionswert des Composition Root; das Paket importiert nur die Standardbibliothek; `.a-check.yml` bleibt.
- Randformen des Signals (Signal vor dem Ende des Starts, zweites Signal, Signal nach einem Konfigurationsfehler, Signal während `CloseAll`) und der Aufrufmuster (leeres Argument, Flag wie `--help`, `version` mit Zusatzargument; Bestand von `bootstrap.Run`) — **Ausgang:** entschieden, Raster und Zusatztabelle unten (`AGENTS.md` §3.12, §3.16).
- Die Spezifikation hat zu Aufrufmustern und Exit-Codes keinen Vertrag (benannte Spec-Lücke) — **Ausgang:** entschieden (`SPEC-138`).
- Exit-Codes und Signalverhalten sind am Binary sichtbar; das Handbuch-Kapitel dazu käme aus `slice-handbuch-record-replay`, dessen Plan den Punkt nicht führt (`AGENTS.md` §3.13) — **Ausgang:** entschieden; die Sendung steht in §1 *Übernimmt (Start-Befehl)* und der DoD von `slice-handbuch-record-replay` (Text des Architect: Aufruf `start`, Exit-Codes 0, 1, 2, 130, 143 und Signalverhalten nach `SPEC-138` Nr. 2 bis 7 in §1 *Übernimmt* und DoD des Nehmers; ein gebundenes Beispiel für `start` ist ein Lauf, der endet: Konfigurationsfehler, Exit 1, Meldung nach `SPEC-110` Nr. 12).
- Dateikollision mit `slice-bootstrap-m2` in `src/bootstrap/main/` und `src/bootstrap/bootstrap.go` — **Ausgang:** offen; durch die Reihenfolge im Start-Trigger gelöst, wenn beide nacheinander laufen.
- Die Pflichten des Starts (`SPEC-138` Nr. 8: Rückgabe statt `os.Exit`, Herunterfahren unter `context.WithoutCancel(ctx)`, Abbruch in jeder Phase, Konfigurationsfehler 1) trägt `slice-bootstrap-m2`; die Zeilen S1 bis S4 gehören in dessen Tabelle *Zusage · Fänger* — **Ausgang:** entschieden; die Zeilen S1 bis S4 stehen im Plan von `slice-bootstrap-m2` (§3, §6) (`AGENTS.md` §3.13).
- Quelle der Kommandos für den Katalog: `slice-handbuch-meldungskatalog` nennt den Aufruf-Zweig in `src/bootstrap/bootstrap.go` als Quelle und behauptet, für Kommandos gebe es keine Spezifikation; beides gilt mit `SPEC-138` nicht mehr (Menge: `start` nur mit gesetztem `Start`, und `version`) — **Ausgang:** entschieden; nachgezogen im Plan von `slice-handbuch-meldungskatalog` (Zeilen mit `Run`, §1 Ziel, Nr. 7, Nr. 8).
- `SPEC-138` Nr. 1 nennt die Kanalgröße „genau 2“; kein Lauf trennt Puffer 1 von 2 (Mutation: 0 von 50, Negativ in §3), und *Ungeprüft* von `SPEC-138` führt es noch nicht (Verifikation V-2) — **Ausgang:** offen; Adresse Architect, zusammen mit den drei Zeilen darunter (Nr. 7 a, Nr. 7 b, Signalquelle). Der Spec-Text ist Sache des Architect.
- Die Prüfgrenze für mehrbytige Zeichen in `TestVersionKlasse` (Verifikation V-3) — **Ausgang:** akzeptiert; Grund und Folge in §3.
- `Konfig.Start` ist in `main` `nil`, `start` ergibt bis `slice-bootstrap-m2` Exit-Code 2 und ist am Binary sichtbar (Verifikation R-7); `slice-bootstrap-m2` führt den Punkt in §1 *Übernimmt (Start-Befehl)* (bei der Closure gelesen: der Nehmer liegt in `open/`, kein Ausschluss trifft die Sendung, `AGENTS.md` §3.13) — **Ausgang:** offen; Adresse `slice-bootstrap-m2`.
- SIGTERM am gebauten Binary und die Verdrahtung von `Start` und `Signale` in `main`: siehe die Zeile *Der Test mit einem echten SIGTERM* oben — **Ausgang:** akzeptiert.
- Die Pläne `slice-container-image` und `slice-listener-tls` nennen `Run` als Skelett (`heute ist Run ein Skelett …`); der Wortlaut wird zu `heute ruft der Adapter cli (cli.Run) keinen Start`; `CMD ["start"]` ist Sache von `slice-container-image` — **Ausgang:** entschieden; die Wortlaute in `slice-container-image` und `slice-listener-tls` sind berichtigt.
- Der Wortlaut `Unbekanntes Kommando: version` für `version x` ist irreführend, aber vom Handbuch gebunden; eine Änderung braucht die Beispiele unter `docs/user/beispiele/` und `SPEC-138` Nr. 4 — **Ausgang:** akzeptiert, kein Folge-Slice.
- Der Test mit einem echten SIGTERM am gebauten Binary fehlt (benanntes Negativ C14: `pruefung build` fährt nur `version`; eine falsche Verdrahtung von `Start` und `Signale` in `main` bleibt bis zum Lauf des Binary im Container ungefangen, `main` ist eine Zeile). `slice-test-integration` ist nicht auf Annahme geprüft und nicht zugewiesen (`AGENTS.md` §3.13) — **Ausgang:** akzeptiert; Grund und Folge in *Ungeprüft* von `SPEC-138`.
- Die Zeile *Nicht behauptet* in `harness/README.md` nennt den Code „ein Skelett aus Composition Root und Version“; das war schon vor diesem Slice veraltet — **Ausgang:** entschieden; die Zeile ist berichtigt (`harness/README.md`).
- `SPEC-138` Nr. 7 b sagt, der erste Abbruch geschehe „genau einmal“; am Kontext ist das nicht beobachtbar, denn ein zweiter `cancel()` ist wirkungslos, und kein Fänger kann die Zahl der Aufrufe zählen — **Ausgang:** offen; Adresse Architect (Vorschlag: „genau einmal“ streichen oder als Negativ in *Ungeprüft* von `SPEC-138` führen; der Code ruft `cancel()` je Schleifendurchlauf nur beim ersten Signal).
- `SPEC-138` Nr. 7 a sagt, ein vor dem Aufruf eingetroffenes Signal breche den Kontext, den `Start` erhält. Der Code sichert das durch eine Vorabprüfung des Kanals vor dem Start der Goroutine; ein Fänger kann die Vorabprüfung nicht von einem Abbruch durch die Schleife trennen, der knapp nach dem Start von `Start` läuft (die Mutation ohne Vorabprüfung bleibt in 50 von 50 Läufen grün, weil die Goroutine nach dem Abbruch beginnt). Der Test prüft die Zusage selbst (`Start` sieht den Abbruch beim Aufruf) — **Ausgang:** offen; Adresse Architect (Vorschlag: den Satz als benanntes Negativ in *Ungeprüft* von `SPEC-138` führen oder die Zusage auf „spätestens beim ersten Prüfen“ abschwächen).
- `SPEC-138` Nr. 1 nennt für die Signalquelle `Signale` weder einen geschlossenen Kanal noch eine Abmeldung gleich `nil` (Randform ohne Ort im Raster, Zeile *Signalquelle*, Zelle *unbekannt oder fremd*). Aus dem Code gelesen: Ein geschlossener Kanal liefert sofort `nil`, der Abbruch folgt und die Schleife liefert beim nächsten Durchlauf 143, ohne dass ein Signal kam; eine `nil`-Abmeldung bricht den `defer` mit einer Panik. Der Code ist unverändert (`AGENTS.md` §3.12) — **Ausgang:** offen; Adresse Architect (Entscheidung: Verstoß der Quelle mit Wortlaut, oder Ungeprüftes in `SPEC-138`; der Code folgt danach).

### Raster *Operation × Argumentart*

Orte sind Nummern von `SPEC-138`.

| Operation | fehlend | leer | unbekannt oder fremd | Zusatz | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|
| Kommandowahl (`args`) | Nr. 2 (`nil` und leere Liste: Version) | Nr. 4 (leeres Element) | Nr. 4 (24 Formen, Fänger C2) | Nr. 4 (hinter `version`, hinter `start`, hinter unbekanntem) | gilt nicht: Die Wahl ist synchron, ohne Kontext und vor jeder Anmeldung; ein Signal dort: *Ungeprüft* | Nr. 9 |
| Ausgabe (`Stdout`, `Stderr`) | Nr. 1 (nil verwirft, `Start` erhält nie nil) | gilt nicht: Ein Writer ohne Daten ist ein gültiger Writer; eine leere `Version`: Nr. 3 | Nr. 1 (jeder `io.Writer`; ein Schreibfehler ändert den Code nicht) | gilt nicht: Der Adapter schreibt je Strom höchstens eine Zeile | gilt nicht: synchron | Nr. 9 |
| `Start` (Funktionswert) | Nr. 1, 4 (`nil`: `start` ist unbekannt, Code 2) | gilt nicht: `Start` erhält nur `ctx` und zwei Writer, nie leer oder nil (Nr. 1, 6); die Rückgabe 0 ist Nr. 5 | Nr. 5 (jede Rückgabe außer 0 → 1) | gilt nicht: genau ein Aufruf (Nr. 6) | Nr. 7 b, 8 b bis e | Nr. 9 |
| Signalquelle (`Signale`, `nil` = System) | Nr. 1, 7 a | gilt nicht: Kein Signal ist kein Ereignis | Nr. 7 a, c (ein anderes Signal als SIGTERM und SIGINT, das die Quelle liefert, zählt als Signal und ergibt als zweites 143; angemeldet werden nur die zwei) | Nr. 7 c (weitere Signale) | Nr. 7 b | Nr. 7 d, 9 |

### Zusatztabelle: Zeitpunkt × Signalnummer, Paare und Verstoß-Paare

| Zeitpunkt oder Paar | erstes Signal | zweites Signal | Fänger |
|---|---|---|---|
| vor der Anmeldung (Prozessstart bis `Notify`) | Standardverhalten, Nr. *Ungeprüft* | – | benanntes Negativ C14 |
| nach der Anmeldung, vor dem Aufruf von `Start` (im Kanal vorbelegt) | Nr. 7 a, 6: `Start` sieht den abgebrochenen Kontext | beide vorbelegt: Nr. *Ungeprüft*, eines von beiden | C8 (Fall vorbelegt); Negativ C14 |
| vor dem Ende des Starts (Phase) | Nr. 7 b, 8 c | Nr. 7 c | C8; C10; `slice-bootstrap-m2` S2 |
| im Lauf | Nr. 7 b, 8 c | Nr. 7 c | C8, C9; C10 |
| während `CloseAll` oder des Herunterfahrens (selbst eingeleitet oder nach dem ersten Signal) | selbst eingeleitet: Nr. 8 e, beschleunigt nichts | Nr. 7 c: 128 plus Nummer, Sessions bleiben `open` | `slice-bootstrap-m2` S1, S4; C10 |
| nach einem Konfigurationsfehler (`Start` lieferte 1) | während des Ladens: Nr. 8 d, Code 1; danach: Nr. 7 d, nicht beobachtet | – | `slice-bootstrap-m2` S3; C6, C11 |
| nach der Rückkehr von `Start` | Nr. 7 d | – | C11 (Abmeldung genau einmal) |
| Paar SIGINT, SIGTERM / SIGTERM, SIGINT / gleiche Art | Nr. 7 b | Nr. 5, 7 c: der Code folgt dem zweiten Signal | C10 (Paare mit verschiedenen Arten in 4 von 10 Fällen) |
| Verstoß-Paar: unbekanntes erstes Argument mit Zusatz | – | – | Nr. 4: die Meldung nennt das erste Argument; C2 (Position „erstes Argument mit Zusatz“) |
| Verstoß-Paar: `start` bei `Start` gleich `nil` mit Zusatz | – | – | Nr. 4: Meldung `start`; C2 |
| Modus (`record`/`replay`) | gilt nicht: Der Adapter kennt den Modus nicht; Wirkung im Start: `slice-bootstrap-m2` S2 | – | – |

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

- **Was hat funktioniert:** (1) Vertrag `SPEC-138`, Tabelle, Raster und Zusatztabelle standen vor dem Code (Basis `4e14497`), die Marke setzte der Planner danach (`1bb9474`); Produktcode `cli.go` 121 Zeilen. (2) Drei Prüfrunden (Review, Fix-Runde, Verifikation): kein HIGH, kein falsches Codeverhalten; die fünf MEDIUM des Reviews (R-1 bis R-5) waren Fänger und Zellen, nicht Code. (3) Die Fix-Runde deckte die Klassen, nicht die genannten Mutationen (`AGENTS.md` §3.17): der Verifier sah vier nicht genannte Mutationen an `Version`, vier an der Abmeldung und zwei an der Anmeldung rot. (4) R-4 und R-5 hat der Implementer nicht im Code entschieden, sondern mit Ausgang *offen* und Adresse in §6 zurückgegeben (`AGENTS.md` §3.12, §3.13).
- **Was ging anders als geplant:** (1) Der Review-Report des Reviewers machte `make gates` rot (V-1): zwei ADR-Kennungen in einer Kopfzeile ohne Link; weder Reviewer noch Implementer ließen nach dem Report `make docs-check` laufen, der Verifier fand es. (2) Die Tabelle wuchs von 14 auf 16 Zeilen (C15, C16 aus dem Review); das Gate meldet dazu den Hinweis *Tabelle ergibt 1950, Schätzung 1780*, die Marke blieb. (3) Drei Zusagen von `SPEC-138` sind am Code nicht trennbar (Kanalgröße genau 2, „genau einmal“ in Nr. 7 b, Vorabprüfung in Nr. 7 a): Negative im Plan, Spec-Nachzug beim Architect.
- **Nicht durch einen zweiten Kontext geprüft:** Der letzte Bericht eines zweiten Kontexts (Verifikation, [Verifikation](../../../reviews/2026-10-31-slice-cli-adapter-verifikation.md), Stand `f905715`) lautet **nicht bestanden** wegen eines einzigen Befunds, V-1 (`make gates` rot, reine Doku), sonst kein HIGH. Eine zweite Verifikation findet bewusst nicht statt (Entscheidung des Auftrags, Durchstich vor Härtung). Ungeprüft von einem zweiten Kontext sind: der Link-Nachzug im Review-Report (`32ffcde`), die Nachzüge in §3 und §6 (V-2, V-3) und diese Notiz; `make gates` im Worktree ist der Beleg für V-1 (siehe *DoD mit Beleg*). Benannt, nicht geschlossen.
- **Lerneintrag:** (1) Der Code war wieder nie der Befund: die Kosten des Slice sind Fänger (Tests und Fakes 1083 von 1204 Zeilen im Paket) und Zellen der Spezifikation. (2) Ein Report des Reviewers oder Verifiers liegt im Repo und fällt unter dieselben Gates wie der Plan: eine nackte `ADR-*`-Kennung in einem `docs/reviews/`-Report macht `docs-check` und damit `make gates` rot, und die Rolle, die ihn schreibt, fährt nach dem Schreiben kein Gate. Der Fall ist ein weiterer Beleg für `BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo` im weiteren Sinn (der Report ist nicht Teil des gegateten Stands, bis jemand ihn fährt); als eigener Kandidat vorgeschlagen (siehe *Steering-Loop-Eintrag*). (3) Die Prüfgrenze einer unbeschränkten Menge (Zeichen) gehört schon in den Plan, wenn der Generator über Bytes statt über Codepunkte läuft (`BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze`): hier entstand sie erst als V-3 der Verifikation. (4) Eine Zusage, die ein Wert nur mit einer Zahl nennt („genau 2“, „genau einmal“), braucht schon beim Schreiben des Vertrags die Frage, welche Mutation sie von der Nachbarzahl trennt; drei Zusagen von `SPEC-138` scheiterten daran (`BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`). **Benannte Lücke, kein `LH-*`:** `SPEC-138` *Ungeprüft* führt die Negative „Kanalgröße genau 2“ und „Vorabprüfung des vorbelegten Signals samt genau einmal“ noch nicht, und die Randform der Signalquelle (geschlossener Kanal, nil-Abmeldung) hat keinen Ort (Adresse: Architect, §6). Neuer Sensor: keiner außerhalb des Pakets (neue Fänger: `TestVersionKlasse`, `TestLaufzeitAbmeldung`, `TestSignaleHintereinander`, `TestAblehnungKlasse`, `TestExitAbbildungKlasse`).
- **Steering-Loop-Eintrag:** Gezählt, nicht verkörpert. Mit diesem Slice wurde nichts in `AGENTS.md`, einem Command, einem Rollen-Typ oder einem Skill geschrieben, daher steht kein `liegt in`. Vorschläge an den Integrationsschritt für das Register (der Planner schreibt es hier nicht): (a) neuer Kandidat `BEO-REPO/report-der-rolle-faellt-docs-check` (1×, offen; Beleg V-1 dieses Slice; Diff-Vorlage an den User, noch nicht geschrieben: `.harness/skills/reviewer.md` und der Verifier-Typ nennen als letzten Schritt `make docs-check` nach dem Schreiben des Reports); (b) weitere Belege (je ein Beleg `evidence/slice-cli-adapter.md`) zu `fix-deckt-die-genannte-mutation-nicht-die-klasse` (R-1 bis R-3: Version, Abmeldung, Kanal; vier Elemente, die der Befund nicht nannte), `unbeschraenkte-menge-ohne-benannte-pruefgrenze` (V-3), `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (Kanalgröße genau 2, genau einmal, Vorabprüfung), `randform-erst-im-review-entschieden` (R-5: geschlossener Kanal und nil-Abmeldung ohne Ort, vom Implementer zurückgegeben statt im Code entschieden: ein Fall, in dem die Regel `AGENTS.md` §3.12 hielt) und `akzeptiertes-negativ-ohne-folge` (kein Beleg: alle Negative nennen Folge).
- **Beobachtungs-Register (`../observations/`):** geschrieben im Integrationsschritt auf `main`: je `evidence/slice-cli-adapter.md` unter `BEO-REPO/` zu `fix-deckt-die-genannte-mutation-nicht-die-klasse`, `unbeschraenkte-menge-ohne-benannte-pruefgrenze`, `vertrag-sagt-mehr-zu-als-ein-gate-prueft`, `randform-erst-im-review-entschieden`; neu `report-der-rolle-faellt-docs-check`. Die Zählung folgt aus den Dateien.
- **Größe (Messpunkte, `AGENTS.md` §3.14):** 1296 Zeilen nach `make groessenmarke` (Summe aus Einfügungen und Löschungen von `git diff --numstat 4e14497..<letzter Commit des Slice>` über die vier Pfade des Blocks, Tests eingeschlossen) gegen die Marke 2500: **gehalten**, 52 % der Marke; die Marke bleibt, keine `ueberschreitung`. Hinweis des Gates: *Tabelle ergibt 1950, Schätzung 1780* (16 Zeilen statt 14; C15 und C16 kamen mit dem Review dazu, `SPEC-128` Nr. 5 c, kein Befund). Messpunkte: Handoff (`e03912b`) 1182, nach der Fix-Runde (`c6bd5c4`) 1296; Ende ÷ Handoff 1,10, Ende ÷ Schätzung 0,73; Produktcode `cli.go` 121 gegen 230 der Schätzung (mit Löschungen in `bootstrap.go` und `bootstrap_test.go`).
- **DoD mit Beleg:** (1) Kommandoauswahl und Exit-Codes: `src/adapters/driving/cli/cli.go`, `bootstrap.go` nur Paketkommentar und `exitOK`, `main.go` ruft `cli.Run`; Zeilen C1 bis C6, C12, C13. (2) Signale: Zeilen C7 bis C11, C15, C16; Raster in §6 vor dem Code entschieden, die drei offenen Zellen stehen mit Adresse in §6. (3) `make gates` Exit 0 im Worktree des Planners nach dem Link-Nachzug (Exit-Code aus Datei, unter der Sperre), darin `make groessenmarke` mit Messung 1296 von Marke 2500. (4) Review und Verifikation: [Review](../../../reviews/2026-10-31-slice-cli-adapter-review.md), [Verifikation](../../../reviews/2026-10-31-slice-cli-adapter-verifikation.md). (5) Doku: `SPEC-138` und `SPEC-110` (zwei Zusätze), `ARC-006`; **Handbuch:** das Binary ändert sich sichtbar nicht (`version` bleibt bytegleich, `start` ist bis `slice-bootstrap-m2` ein unbekanntes Kommando); das Kapitel zu Aufruf, Exit-Codes und Signalverhalten ist Sache von `slice-handbuch-record-replay`, der Katalog Sache von `slice-handbuch-meldungskatalog` (§6, Ausgang *offen*, Nehmer-Nachzüge macht der Integrationsschritt auf `main`). (6) Closure-Notiz: dieser Abschnitt.
- **Offener Rest:** (1) Architect: `SPEC-138` *Ungeprüft* um die Negative Kanalgröße genau 2 (V-2), „genau einmal“ (Nr. 7 b, Nr. 7 a Vorabprüfung) und die Randform der Signalquelle (R-5) ergänzen oder die Zusagen abschwächen; danach der Code, falls R-5 ein Verhalten verlangt. (2) `slice-bootstrap-m2`: `Konfig.Start` in `main` setzen (Sendung steht in dessen §1 *Übernimmt (Start-Befehl)*), S1 bis S4 in seine Tabelle (steht dort vorab). (3) erledigt im Integrationsschritt auf `main`: Register geschrieben, Nehmer-Pläne nachgezogen (`slice-handbuch-record-replay`: Kapitel Aufruf, Exit-Codes 0, 1, 2, 130, 143, Signalverhalten; `slice-handbuch-meldungskatalog`: Quelle `SPEC-138` statt `Run`; `slice-container-image` und `slice-listener-tls`: Wortlaut zu `Run`), Roadmap und Welle, `harness/README.md` *Nicht behauptet* („Skelett“). (4) Akzeptiert: SIGTERM am gebauten Binary (Verdrahtung in `main`; keine Adresse, `slice-test-integration` nicht zugewiesen), mehrbytige Zeichen in `Version` (V-3).
- **Folge-Slices:** [slice-bootstrap-m2](../open/slice-bootstrap-m2.md) nimmt die Sendungen an (§1 *Übernimmt (Start-Befehl)* und Tabelle S1 bis S4 vorab gelesen, kein Ausschluss trifft sie). Die übrigen Zeilen in §6 (Handbuch-Kapitel, Katalog, `slice-container-image`, `slice-listener-tls`) sind Sendungen des Planners an Pläne in `open/`; ihr Nachzug steht im Offenen Rest (3).
- **Risiken aus §6:** alle mit Ausgang — entschieden: Zugang zum Start (`Konfig.Start`), Randformen, Spec-Lücke; akzeptiert: Wortlaut bei `version x`, SIGTERM am Binary, mehrbytige Zeichen in `Version`; weiter offen: die vier Zeilen an den Architect (Kanalgröße, „genau einmal“, Vorabprüfung, Signalquelle), `Konfig.Start` (`slice-bootstrap-m2`), Pflichten des Starts S1 bis S4 (`slice-bootstrap-m2`), Handbuch-Kapitel, Katalog, `slice-container-image`/`slice-listener-tls` und `harness/README.md` (Planner, Integrationsschritt), Dateikollision mit `slice-bootstrap-m2` (durch die Reihenfolge gelöst, sobald der Nehmer startet).
- **Drei Paarungen:** Anker (kein Zielort-Feld, nichts verkörpert; die Sensor-Paarung hat keinen Eintrag zu prüfen) · Folge-Slice (`slice-bootstrap-m2` in `open/` nimmt an; Nachzüge der Nehmer im Offenen Rest) · Register (geschrieben, siehe oben; die Paarung prüft die Welle-Closure von `welle-record-replay` nach dem Integrationsschritt).

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9) und, weil die Signalbehandlung nebenläufig ist, `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (8×, verkörpert in `AGENTS.md` §3.10); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
