# Verifikations-Report: slice-replay-service — 2026-10-31

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? (gegen Plan, DoD und `SPEC-135`) · **Zweig:** `slice-replay-service`, Stand `2ccbaf5` · **Basis des Diffs:** `b919969` · **Bezug:** [LH-FA-04, LH-FA-14, LH-QA-03](../../spec/lastenheft.md), [SPEC-135](../../spec/spezifikation.md), [Plan](../plan/planning/done/slice-replay-service.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [SPEC-128](../../harness/vertraege/SPEC-128-groessenmarke.md).
Es ist die erste Verifikation (Verifikation 1, frischer Kontext). Geprüft sind der Slice nach der Fix-Runde zu F-1 bis F-8 und F-9 des [Review-Reports](2026-10-31-slice-replay-service-review.md) und die Zusagen des Planners aus §6 des Plans.

## Verdikt: NICHT BESTANDEN

- Der Produktcode erfüllt `SPEC-135` in jeder gemessenen Klasse bis auf eine Stelle ohne Wirkung im Betrieb (V-4, LOW). Kein HIGH. Von 72 eigenen Mutanten (Produktcode, Orakel von Hand, Probe-Paket mit eigenen Fakes) sind 59 in der Suite des Slice rot, 2 gleichwertig und **11 grün**: Der Beweis hat Lücken, der Code nicht.
- **Drei Klassen der Fix-Runde (`AGENTS.md` §3.17) sind nicht zu Ende gedeckt**, und jede ist ein MEDIUM, weil die Zusage am Element der Klasse ohne Fänger ist (`AGENTS.md` §3.15): die ASCII-Schreibweise der Header-Namen (V-1: von 26 Buchstaben fangen 4 nichts, B, F, J, Z), „der Dienst hält keinen Zustand und ist gleichzeitig sicher“ für Zustand in der Instanz (V-2: weder Quelltext-Fänger noch Detektor, 0 von 50), und „`Laenge` 0 heißt: nicht gelesen“ für einen Body mit Bytes (V-3).
- Die Wirkungs-Fänger der Transaktion halten: Put in zweiter Transaktion, `Workspaces.Get` in eigener Transaktion und der Cursor vor `Sessions.Get` oder nach `ListBySession` gelesen sind in jedem der 50 Prozesse rot (Abschnitt 4); der Befund des Reviews zu Z49 ist behoben.
- Der Vertrag `SPEC-135` ist an einer Stelle nicht widerspruchsfrei (V-6, Nr. 12: „kein Code für den Client“ neben `ODR-E9001`); der Nachzug liegt beim Architect und ist unterwegs.
- Alle sechs Adressen aus §6 des Plans nehmen an (Abschnitt 7); zwei Nachzüge beim Planner (V-7).

Befunde: 0 HIGH, 3 MEDIUM, 3 LOW, 4 INFO.

## 1. Sensor-Belege (selbst gefahren)

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` im Worktree | Exit 0 (aus Datei `gates.rc`, nicht über eine Pipe); darin `a-check` „gesamt: 0 Befund(e)“, `d-check` „468 Datei(en) geprüft, 0 Befund(e)“, `baseline-verify` OK, `migration-check` ok, `make groessenmarke` „ok, Messung 5451 von Marke 9200“. Viele Stufen `CACHED`, darum die drei Stufen unten einzeln ohne Cache | `…/scratchpad/verif-replay/gates.out`, `gates.rc` |
| Stufe `test` ohne Cache (`--no-cache-filter test`) | Exit 0, `pruefung test: ok` (79 s) | `nc-test.out` |
| Stufe `race` ohne Cache (`--no-cache-filter race`, Image des Targets) | Exit 0, `pruefung race: ok` (114 s) | `nc-race.out` |
| Stufe `lint` ohne Cache | Exit 0, `0 issues` | `nc-lint.out` |
| `make groessenmarke` | Exit 0 im Gate; eigene Nachmessung `git diff --numstat b919969..HEAD` über die zwei Pfade des Blocks: 5451 Einfügungen, 0 Löschungen = 5451 (gleich dem Gate) | siehe Abschnitt 5 |
| Stufe `zeichen` | existiert in diesem Worktree nicht (kein Ziel im `Dockerfile` und in `harness/mk/`) | `grep` |
| `make gegenproben`, Stempel | nicht gefahren (Auftrag) | – |

Die Läufe des Verifiers liefen in Kopien im Scratchpad (`/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/verif-replay/`), im Image `odata-recorder:test` bzw. `odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, je Gruppe unter `flock /tmp/odata-recorder-heavy.lock`; im Worktree liegt nichts außer diesem Report. Last: Host mit weiteren Läufen anderer Zweige unter derselben Sperre.

## 2. Plan gegen Code (Diff `b919969..HEAD`)

- Unter `src/` geändert sind genau die Pfade des Blocks: `src/hexagon/ports/inbound/replay.go` (12 Zeilen), `src/hexagon/services/replay/` (24 Dateien) und `src/hexagon/services/replay/begrenzt/` (2 Dateien); zusammen 27 Dateien, 5451 Zeilen, davon Produktcode 627 (`replay.go` 224, `dienst.go` 105, `auswahl.go` 103, `key.go` 70, `antwort.go` 48, `begrenzt.go` 35, `fehler.go` 30, Port 12), Tests und Fakes 4824. Außerhalb von `src/`: Plan, Roadmap, `spec/spezifikation.md`, der Review-Report und die Pläne der Zusammenführung mit `main`. Nicht berührt: `AGENTS.md`, `harness/`, `tools/`, `.claude/`, `.harness/`, `Dockerfile`, `Makefile`, `src/hexagon/model/` (die Codetabelle bleibt bei 33; kein Test und keine Zählstelle ändert sich).
- „Ausdrücklich NICHT“ eingehalten: kein HTTP-Eingang, kein Hybrid, keine Workspace-Anlage; der Dienst importiert nur `context`, `errors`, `io`, `slices`, `bytes`, `log/slog`, `strings`, Domain und Ports (Quelltext gelesen; `make a-check` 0 Befunde; mein Probe-Fake ruft `StartSpan` und `CountError` mit Panik, beide werden nie gerufen).
- Die Zahlen des Plans stimmen: Tabelle 66 Zeilen (Z1 bis Z66), 9 nebenläufig, 57 andere, 7 Negative (Z59 bis Z64, Z66); alle 66 Testnamen, die der Plan nennt, bestehen im Quelltext (Skript über jedes `Test…` im Plan, 0 fehlende). Z41 hat 4·3·3·4·2·2 = 576 Elemente, davon 48 geliefert (nachgerechnet).
- Fix-Runde: F-7 behoben (der Satz „bis dahin lehnt der Rest jeden Request mit Body ab“ steht nicht mehr im Plan; §4 trägt 4715 und 5451); F-8 behoben (`Puffer.Bytes` liefert unmittelbar `b.buf.Bytes()`, kein Zweig).
- Rohbytes (`AGENTS.md` §3.18): das Verfahren `env LC_ALL=C grep -aP` ist an `git show 4bbe19d^:docs/reviews/2026-10-25-slice-baseline-import-verifikation-3.md` bestätigt (`\xC2\xA0` ergibt 1). In allen geänderten Dateien von `src/`, `spec/`, Plan und Review: kein BOM, kein NBSP, kein Zeichen der Klasse Zero-Width oder Steuerzeichen, kein CR, jede Datei endet mit Zeilenende. Der Arbeitsbaum ist sauber (`git status` leer, keine `.orig`, `.bak`).

## 3. DoD Punkt für Punkt

Alle Häkchen der DoD stehen unausgefüllt (`[ ]`); das Setzen gehört dem Planner bei der Closure. Stand der Belege:

| DoD-Punkt | Beleg liegt vor? |
|---|---|
| Zuordnung und Validierung (`ODR-E5003`, `E5004`, `E7001`, `E7003`, `E7009` ohne Cursoränderung; Identität des Cursors; neue Replay-Session bei 0; Form der ID) | ja: Code gelesen, Suite und Probe (Steuerheader 18 688 Fälle, Generator 32 000 Schritte) ohne Abweichung am Code; Lücken nur im Beweis (V-1, V-3) |
| Auswahl (Match-Key aus dem redigierten Request, vier Strategien, `E5001`, `E5002`, `E5007`, `E5008`, Replayfähigkeit, atomares Fortschreiben, unabhängige Cursors) | ja: Generator gegen eigenen Simulator, 4000 Läufe, 32 000 Schritte, 0 Abweichungen (Abschnitt 5); Wirkungs-Fänger 50 von 50 (Abschnitt 4) |
| Auslieferung (Blob mit Integrität `E3001`, Header nach Nr. 13, `X-Recorder-Match-Redacted`, kein Upstream) | ja: Mutanten e1, e2, f1 bis f3, i1, o1 rot; Policy gegen `LH-FA-04.a` Nr. 3 gelesen (Abschnitt 6) |
| `make gates` grün | ja (Abschnitt 1) |
| Review durchgeführt, Report liegt vor | ja: [Review-Report](2026-10-31-slice-replay-service-review.md); ein Re-Review der Fix-Runde gab es nicht, ihre Fixes löst dieser Report gegen die Klasse auf (Abschnitt 5) |
| Doku-Update öffentlicher Vertrag | ja: `SPEC-135` und die Nachzüge in `SPEC-109`; Doku-Gate grün; Wortlaut-Befund V-6 |
| Handbuch | ja, begründet: kein Binary-Verhalten, kein Code, kein Schlüssel (Plan DoD, Sendung 5) |
| Closure-Notiz mit Lerneintrag · Register · Risiko-Ausgänge · drei Paarungen | offen (Planner): §7 steht auf *Wird bei der Closure gefüllt*, §6 trägt für die meisten Risiken *offen bis zur Closure* |

## 4. Nebenläufigkeit: Wirkungs-Fänger gegen Detektor, mit Raten

Messender: der Verifier. Je Zelle 50 Prozesse, ein Prozess ein Lauf von `-test.count=1`, die Testbinärdatei einmal gebaut; Last: 4 CPUs, 4 GB, Host mit weiteren Läufen unter `flock`. Mutanten in einer Kopie des Dienstes, gesteuert über eine Umgebungsvariable; sie legen eine Kontextvariable an (die sechs Fänger prüfen keine Identität des `ctx`).

| Mutant | Fenster | TestGleichzeitigGleicherCursor | TestZweiReplaySessions | TestReplayGegenReset | TestFensterDerTransaktion | TestFensterLoeschen | TestReplayGegenLoeschen |
|---|---|---|---|---|---|---|---|
| ohne Mutation | – | 0/50 | 0/50 | 0/50 | 0/50 | 0/50 | 0/50 |
| m1: `Put` in zweiter Transaktion, Cursor in der ersten gelesen | Fenster 3 | **50/50** | **50/50** | **50/50** | **50/50** | **50/50** | **50/50** |
| m2: `Workspaces.Get` in eigener Transaktion vor der Hauptransaktion | Fenster 1 | 0/50 | 0/50 | 0/50 | **50/50** | **50/50** | **50/50** |
| m3: Cursor vor `Sessions.Get` in eigener Transaktion gelesen | vor der Transaktion | **50/50** | **50/50** | **50/50** | **50/50** | 0/50 | **50/50** |
| m5: Cursor nach `ListBySession` in eigener Transaktion gelesen | vor der Transaktion | **50/50** | **50/50** | **50/50** | **50/50** | 0/50 | **50/50** |

- Die Zellen mit 0/50 bei einem Mutanten sind nicht die Zuständigen des Fensters (`TestFensterLoeschen` fragt allein das Löschen des Workspace zwischen Schritt 1 und 2); jedes Fenster hat mindestens einen Fänger mit n von n. Z49 (der Befund des Reviews: nur ein Lauf rot) ist behoben: m3 ist in `TestFensterDerTransaktion` in 50 von 50 Prozessen rot.
- **Detektor** (`odata-recorder:race`, `-race`, je Prozess `-test.count=1`, 50 Prozesse): ohne Mutation 0 von 50 in beiden Gruppen (die sechs Fänger und mein ungeordneter Test, Abschnitt 5, V-2). Der Detektor meldet die Fenster-Mutanten nicht (jeder Zugriff liegt unter der Sperre, `SPEC-122` Nr. 8 b); das trägt der Plan so (§3, Z50 bis Z52).
- Negative Z59 bis Z64 und Z66: jedes nennt **Grund und Folge** (Wortlaut gelesen). Z63 („ein Datenrennen, das kein Test auslöst“) deckt die Serialisierung durch den Adapter; es deckt **nicht** den Zustand in der Instanz (V-2).

## 5. Fix-Klassen gegen die Klasse aufgelöst (`AGENTS.md` §3.17)

**Probe-Paket** (`probe_test.go`, Package `replay_test`, eigene Fakes mit eingebetteten Schnittstellen: jede nicht benannte Operation ist ein Panic; Orakel von Hand aus dem Wortlaut von `SPEC-135`, `SPEC-109` Nr. 5 und `SPEC-045`, nie der Dienst). Gegen den Stand `2ccbaf5` läuft es grün bis auf V-4.

| Klasse (Zusage) | Menge, Achsen · Elemente | Ergebnis am Code |
|---|---|---|
| F-1 Replayfähigkeit (`SPEC-045`, Nr. 8, 9) | Strategie {4} × `capture_status` {complete, incomplete, failed, leer, fremd} × Integrität {verified, absent, invalid, leer} × Response-Redaction {none, applied, unsafe, leer, fremd} × Flag {2} × Layout {Einzelner, zwei Exchanges} · 1600 | 0 Abweichungen |
| F-2 Schreibweise der Steuerheader | 3 Namen × jede Position × jedes Byte 0x00 bis 0xFF = 73 · 256 = 18 688 | 0 Abweichungen |
| F-2 `Content-Encoding` Name, Wert; `Content-Type` Name | 16·256 + 8·256 + 12·256 = 9216 | 0 Abweichungen |
| F-2 flüchtige Header (Konfig × Antwort) | Byte b1 × Byte b2 am Namen `X-<b>` = 256 · 256 = 65 536 | 0 Abweichungen |
| F-3 Verbünde und Einzelfehler an jeder Stelle | Stelle {S1, S3, S4, S5, S6, S7, S8} × 551 Fehlerwerte (22 Fehler einzeln und verpackt, mit `ErrNichtGefunden` in beiden Reihenfolgen und verpackt, 441 Tripel `Join(ErrNichtGefunden, B, C)`) = 3857 | 2 Abweichungen, beide `ErrBereitsVorhanden` an S8 (V-4) |
| F-4 `New` je Position | Dienste und Pfade je Liste der Länge 1 bis 4 an jeder Position (leere Kennung, leerer Root, Nullwert), fünf Ports einzeln nil, Aufzählungen | 0 Abweichungen |
| F-5 Kontext nach der Bestätigung | Endegrund {`Canceled`, `DeadlineExceeded`, eigener Grund} × Strategie {4} × Zeitpunkt {nach Bestätigung ohne Eintrag, mit Eintrag, im Eintrag} = 36; zusätzlich Identität des `ctx` an S1, S3, S4, S5, S6, S7, S8, S10 je Strategie | 0 Abweichungen |
| F-6 Strom unverändert | Strategie {4} × {Identität, `Close` 0 bei Erfolg}; Fehlerpfade `Put` (2 Strategien) und Bestätigung (4): `Close` genau 1 | 0 Abweichungen |
| F-9 Zustand des Response-Bodys | Zustand {`present`, `absent`, `discarded`, leer, `x`, `PRESENT`, `Absent`, ` present`, `present `} × Strategie {4} × replayfähig {2} = 72; `Open` 0, `Put` 0, Cursor unverändert, `Body` nil, Status 500 | 0 Abweichungen |
| Auswahl und Cursor (Nr. 9, 10) | Generator, Saat 20261031: 4000 Läufe (0 bis 6 Exchanges in gemischter Ordnung, Keys A, B, leer; zufällig nicht replayfähig; vorbelegter Cursor mit anderer Strategie; Flag) × 8 Schritte = 32 000, Antwort und Cursor (`Position`, `Consumed`, `Strategy`, `ServiceID`, Existenz) gegen einen Simulator aus dem Wortlaut | 0 Abweichungen |
| Header sind eine Kopie (Nr. 13) | Map, jede Wertliste, `ETag` mit zwei Werten | 0 Abweichungen |

**72 Mutanten** (40 am Produktcode in `auswahl.go`, `dienst.go`, `key.go`, `antwort.go`, `replay.go`, 4 Fenster-Mutanten der Transaktion, 1 Öffnen des Blobs nach der Bestätigung, 1 Zustand in der Instanz, 26 je ein Buchstabe der ASCII-Faltung). Rot in der Suite des Slice: 59. Grün in der Suite: 13, davon zwei gleichwertig.

| Mutant (an Elementen, die weder Implementer noch Reviewer nannten) | Suite des Slice | Probe |
|---|---|---|
| a1 bis a6: `Replayfaehig` mit `false`, `true`, `!flag` an jeder der drei Wahlfunktionen (6 von 6) | rot (`TestReplayfaehigkeit`) | rot |
| b2 `c > 'A'`, b5 Rand mit Zeilenvorschub, b7 `EqualFold` im Namensvergleich, b8 und b9 `ToLower` an Antwort und Konfig, b10 Konfig ohne Faltung | rot | rot |
| **b1 `c < 'Z'`, b3 `c >= '@'`, b4 `c <= '['`** | **grün** | rot |
| b6 `EqualFold` für den Wert `identity` | grün, **gleichwertig** (kein Buchstabe von `identity` hat einen Faltungspartner) | grün |
| c1 nur mittlerer Dienst, c2 nur mittleres Muster, c3 Root ungeprüft, c4 Kennung ungeprüft | rot | rot |
| d2 `Canceled` zwischen Bestätigung und Eintrag, d3 `ctx.Err()` nach dem Eintrag, d4 `ctx.Err()` vor dem Eintrag | rot | rot |
| **d1 `DeadlineExceeded` zwischen Bestätigung und Eintrag** | **grün** | rot |
| e1 `struct{ io.ReadCloser }`, e2 `io.NopCloser` | rot | rot |
| f1 nur `discarded`, f2 `discarded` und `x`, f3 ohne den leeren Wert | rot (`TestResponseBodyZustand`) | rot |
| g3 `GegenstandSonstiges` an S8, h1b `fehlt` immer wahr, h2 `==` statt `errors.Is`, i1 Map des Exchange, j2 `position` statt `position+1`, k1 `Position` gleich `SequenceNo`, m1 Root `/` ohne führenden `/`, m2 Grenze 129 | rot | rot |
| **h3 `GegenstandSession` an S6 für alle Fehler außer dem fehlenden Cursor** | **grün** | rot (`ErrBereitsVorhanden`) |
| j1 leerer `MatchKey` passt | grün, **gleichwertig** (der Key des Requests ist nie leer) | grün |
| **l1 `Body` nicht nil und `Laenge` 0 wird gelesen** | **grün** | rot |
| o1 `Blobs.Open` nach der Bestätigung | rot (fünf Tests) | – |
| **r1 Feld `n` im `Replayer`, `r.n++` am Anfang von `Replay`** | **grün (auch `TestPaketOhneZustand`)**, Detektor 0 von 50 | Detektor 50 von 50 |
| 26 Mutanten `c != '<Buchstabe>'` in `asciiKlein`: rot in der Suite für 22 Buchstaben, **grün für B, F, J, Z** | 22 rot, **4 grün** | 26 von 26 rot |

## 6. Entscheidungen und Vertrag

- Die elf Entscheidungen des Architect (§3 des Plans) sind im Code umgesetzt: Rangfolge a bis h (Generator und Probe), Form der Replay-Session-ID, Request-Body durch den Redaktor mit Grenze (Key aus der Ausgabe), Auswahl mit `ODR-E5001` vor allen, Cursor (`first` und `last` schreiben nie, der bestehende Cursor entscheidet), `Open` in `f` (o1 rot), Header als Policy, Auswahl ohne Auslieferung (d2 bis d4 rot), ein Eintrag über `Telemetrie.Log`, keine Fehlerzählung, kein neuer Code.
- Policy der Header gegen `LH-FA-04.a` Nr. 3 („ausgewählte End-to-End-Header; zeit- oder hostabhängige nur nach expliziter Policy“): konform. Entfernt werden redigierte (wertbasiert), `Content-Length` und `Content-Encoding` (Leitung, Adapter setzt sie), die Namen aus `VOLATILE_HEADERS` (die ausdrückliche Policy, `none` liefert alle) und ein gespeichertes `X-Recorder-Match-Redacted`; `Location` mit Marker geht aus. Die Aussage in §6 „der Verifier prüft die Policy“ ist damit erfüllt; das Risiko *Location* bleibt eine Entscheidung des Owners (V-7).
- **`SPEC-135` Nr. 12 und 15, Widerspruchsfreiheit.** Nr. 15 (13 erreichbare Codes, `ODR-E9001` aus dem Zustand des Response-Bodys), Nr. 3 (g) und die Anwendung in `SPEC-109` Nr. 5 sind untereinander und mit dem Code stimmig. Nr. 12 widerspricht ihnen in einem Halbsatz: „`ODR-E9001` (500) … (kein Code für den Client, kein Strom, `Open` wird nicht gerufen)“; `ODR-E9001` ist ein Code, der an den Client geht. Der Folgesatz „Das gilt für Antworten auf `HEAD`, 1xx, 204, 304 und leere Bodies“ stand hinter der Aussage zu `absent` und hängt jetzt hinter der Klausel zu `discarded`; „Das“ bezieht sich dann rückwärts auf `ODR-E9001` (ein `HEAD` wäre ein Fehler), gemeint ist der leere Strom. Beides ist ein Wortlaut-Fehler, nicht des Verhaltens (V-6).

## 7. Adressen aus §6 des Plans (`AGENTS.md` §3.13)

Gelesen: §1 *Ausdrücklich NICHT* und DoD der Nehmer in `open/`. Die Pläne in `open/` bleiben nach der Entscheidung des Users bis zum Start unverändert (`welle-record-replay.md`, *Auftrag beim Start eines Slice aus `open/`*); die Sendungen stehen deshalb mit Ausgang *offen* in §6 des Gebers.

| Sendung | Nehmer | Ausschluss trifft? | Urteil |
|---|---|---|---|
| 1 `Close` genau einmal, Länge und Coding selbst, Fehler ohne Code, `X-Recorder-Match-Redacted` unbekannt | `slice-http-adapter` | nein; DoD „Schließen durch den Adapter“, „Fehlerobjekt nach `SPEC-006`“ trägt es | gültig |
| 2 Verdrahtung der elf Felder; Entscheidung zu `REPLAY_ASSIGNMENT=ordered` | `slice-bootstrap-m2` | nein; DoD „Verdrahtung und Lebenszyklus“; der Punkt `ordered` steht dort nicht und braucht zuerst die Entscheidung des Users | gültig für die Verdrahtung; `ordered` **bedingt** (INFO, I-2) |
| 3 Match-Span `odata.replay.match`, Fehlerzählung am Ort der Auslieferung | `slice-telemetry-einbau-gateway` | nein; DoD nennt „Match-Span im Replay-Dienst“ und den Zähler je Code | gültig |
| 4 Aufwand je Request bei 10 000 Exchanges, Folge `ListByMatchKey` | `slice-lasttest` | nein (Tuning ist ein eigener Slice, die Messung trifft nicht); **aber** die DoD „der Match-Key-Index wird ohne Tabellenscan genutzt“ trifft der Replay nicht (V-7) | gültig mit Nachzug |
| 5 Sichtbares für das Handbuch | `slice-handbuch-record-replay` | nein; Auftrag Nr. 1 bis 5 trägt Aufzeichnen und Wiedergeben | gültig |
| 6 keine neue Pflicht, nur `ODR-E5001` darf den Fallback auslösen | `slice-hybrid-service` | nein; DoD „`ODR-E5001` geht weiter … jeder andere Replay-Fehler unverändert“ deckt es | gültig |

Die Randform *discarded* (Ausgang *entschieden*, Wortlaut des Architect verbatim in `SPEC-135` Nr. 3 g, 12 und 15, im selben Commit wie der Code und der Fänger Z65) ist als Zeile in §6 und im Raster geführt; sie ist eine Randform, die das Review fand, nicht eine, die der Implementer im Code entschied (`AGENTS.md` §3.12).

## 8. Messung der Größenmarke

- Gate: 5451 von Marke 9200; eigene Nachrechnung `git diff --numstat b919969..HEAD` über `src/hexagon/ports/inbound/replay.go` und `src/hexagon/services/replay`: 5451 (alle Einfügungen). Auslastung 59 %.
- Block: *S* = 85·9 + 80·55 + 200·4 + 800 = 6765 und Marke 9200 (nachgerechnet; 135·6765/10000 = 91,33, aufgerundet 92, Marke 9200). Die Tabelle hat 66 Zeilen (57 andere): 85·9 + 80·57 + 200·4 + 800 = 6925; das Gate meldet „Tabelle ergibt 6925, Schätzung 6765“ als Hinweis. Das ist stimmig: `SPEC-128` Nr. 5 c macht im Code denselben Unterschied zum Hinweis, weil die Tabelle mit den Korrekturen wächst und die Marke nach dem ersten Code-Commit fest ist; die Marke aus 6925 wäre 9400. Kein Befund.
- Vierter Messpunkt für `SPEC-128` Nr. 2 (nur genannt): Handoff 4715, Ende 5451 (Faktor 1,156); Ende ÷ *S* = 0,806 (Block) bzw. 0,787 (Tabelle), unter dem Zuschlag 1,35. Produktcode 627 gegen 800 erwartet (Spanne 650 bis 950; unter der Spanne).

## 9. Befunde

| ID | Schwere | Befund | Quelle | Ort | Zuständig | Reproduzierbar |
|---|---|---|---|---|---|---|
| V-1 | MEDIUM | **Die ASCII-Faltung der Header-Namen ist nur an den Buchstaben der Fixtures gefangen.** Zusage: „Header-Namen vergleicht der Dienst überall ohne Beachtung der ASCII-Schreibweise“ (`SPEC-135` Kopf). Die Menge der Klasse ist jeder Buchstabe A bis Z (26) und die Nachbarn der Grenzen (`@`, `[`). Die Fix-Runde zu F-2 führt vier Schreibweisen fester Namen und vier Faltungen; die Buchstaben B, F, J und Z kommen in keinem Namen der Fixtures vor, an dem die Faltung etwas entscheidet. Die Mutanten `c != 'B'`, `'F'`, `'J'`, `'Z'` in `asciiKlein` (26 Mutanten, 22 rot, 4 grün) und `c < 'Z'`, `c >= '@'`, `c <= '['` bleiben in der ganzen Suite grün. Folge im Betrieb: ein konfigurierter Name in `VOLATILE_HEADERS` mit einem solchen Buchstaben (etwa `X-Zeit`) würde in anderer Schreibweise der Antwort nicht entfernt. Der Code ist richtig; der Beweis fehlt. `@` und `[` sind keine Zeichen eines HTTP-Namens, die Grenzelemente sind daher ohne Wirkung im Betrieb, gehören aber zur Klasse (`AGENTS.md` §3.15: beide Seiten jeder Grenze). | `AGENTS.md` §3.15, §3.17, `SPEC-135` Kopf | `src/hexagon/services/replay/dienst.go` · `func asciiKlein` ; Test `TestAntwortHeaderFaltung`, `TestSteuerheaderSchreibweise` | Implementer: ein Fänger als Schleife über alle 256 Bytes je Seite (wie das Probe-Paket: Konfig-Name × Antwort-Name, 65 536 Paare) | ja: `sed` auf Zeile 13 von `dienst.go`, `go test` (`letters2.out`) |
| V-2 | MEDIUM | **„Der Dienst hält keinen Zustand, ist für gleichzeitigen Aufruf sicher“ (`SPEC-135` Nr. 17) hat für Zustand in der Instanz keinen Fänger.** Ein Feld `n int` im `Replayer` und `r.n++` am Anfang von `Replay` bleibt grün: `TestPaketOhneZustand` meldet nur Paketvariable, `init`, go-Anweisung, `sync`, Typen, Methoden, Schreibzugriff auf die Konfig; das Feld und sein Schreibzugriff liegen außerhalb der Regeln. Der Detektor sieht es nicht (0 von 50): die sechs nebenläufigen Fänger bauen je Aufruf einen **eigenen** Dienst (`lauf` ruft `u.replayer()`), und ihre Aufrufe sind geordnet (`SPEC-122` Nr. 8 b). Das akzeptierte Negativ Z63 nennt die Serialisierung durch den Adapter und „ein Datenrennen, das kein Test auslöst“, nicht diesen Fall, und die Raten des Plans nennen allein die Paketvariable. Mein ungeordneter Test (16 Goroutinen, 20 Aufrufe, **ein** `Replayer`, Startsperre, `first`) ist unter dem Detektor in 50 von 50 Prozessen rot gegen den Mutanten und 0 von 50 gegen den Stand. Folge im Betrieb: ein Zähler oder Zwischenspeicher im geteilten Dienst (der Normalfall, ein `Replayer` je Prozess) wäre ein Datenrennen unter Last, das die Suite nie zeigt. | `AGENTS.md` §3.10, §3.15, `SPEC-135` Nr. 17 | `src/hexagon/services/replay/nebenlaeufig_test.go` · `func lauf` ; `TestPaketOhneZustand` | Implementer: ein Test, der **einen** Dienst ohne gegenseitige Ordnung ruft und unter `make test-race` läuft (Rate mit Zahl der Prozesse in §3 des Plans), oder die Regel „kein Feld im `Replayer` außer der Konfig“ in den Quelltext-Fänger; sonst Z54 als akzeptiertes Negativ mit Grund und Folge | ja: `mut3/r1`, `racerun.out` |
| V-3 | MEDIUM | **„`Laenge` 0 heißt kein Body, er wird nicht gelesen“ (`SPEC-135` Nr. 7 a) ist für einen Strom mit Bytes ohne Fänger.** Die Menge der Zusage ist `Body` {nil, leer, mit Bytes} × `Laenge` {-1, 0, >0}; Z3 und Z17 führen `Laenge` 0 mit leerem oder fehlendem Body (Gutfall ohne Wirkung auf den Key). Der Mutant `if a.Body != nil && a.Laenge != 0` zu `if a.Body != nil` bleibt grün; mit einem Strom von 7 Bytes, `Laenge` 0 und `Content-Type` JSON liest er den Body, bildet einen anderen Key und antwortet `ODR-E5001`, gelesen 7 Bytes (Probe `TestVLaenge0`, 2 Fälle, Methode GET und POST). Folge im Betrieb: gering (ein HTTP-Server liefert bei `Content-Length: 0` keinen Strom mit Bytes), aber die Zusage steht im Vertrag. | `AGENTS.md` §3.15, §3.17, `SPEC-135` Nr. 7 a | `src/hexagon/services/replay/key.go` · `if a.Body != nil && a.Laenge != 0 {` | Implementer: ein Fall `Laenge` 0 mit Bytes (Zähler gelesener Bytes 0, Key gleich dem ohne Body) | ja: `mut4/l1` |
| V-4 | LOW | **Der Gegenstand an S8 (und S6) weicht von `SPEC-135` Nr. 18 für `ErrBereitsVorhanden` ab.** Nr. 18 sagt für S8: „`Session` bei `ErrNichtGefunden`, sonst `Sonstiges`“; der Code bildet jeden Fehler von `Cursors.Put` mit `GegenstandSession` ab. Die Tabelle von `SPEC-109` Nr. 5 hat drei Zeilen, die vom Gegenstand abhängen (`ErrNichtGefunden`, `ErrBereitsVorhanden`, `ErrUngueltig` nur für `Import`); Z6 führt nur `ErrNichtGefunden`. `ErrBereitsVorhanden` von `Put` ergibt `ODR-E7010` statt `ODR-E9001` (2 von 3857 Fällen der Probe); der Mutant h3 (`GegenstandSession` auch an S6) ist in der Suite grün. Der Port liefert den Fehler nicht (`SPEC-108` Nr. 11: `ErrNichtGefunden`, `ErrFremderService`, `ErrUngueltig`), eine Wirkung im Betrieb gibt es nicht. | `SPEC-135` Nr. 18, `SPEC-109` Nr. 5, `AGENTS.md` §3.15 | `src/hexagon/services/replay/replay.go` · `return model.Exchange{}, strom, abbilden(err, outbound.GegenstandSession)` | Implementer: Gegenstand nach `errors.Is(err, ErrNichtGefunden)` und ein Fall je gegenstandsabhängiger Zeile; oder Architect: Nr. 18 auf „`Session`“ ändern | ja: `probe-base.out`, `mut4/h3` |
| V-5 | LOW | **Kontext nach der Bestätigung: der Fänger führt nur den Grund `Canceled`** (Antwort auf die Frage des Auftrags: ja, ein Befund, LOW). Der Mutant „Fehler, wenn `ctx.Err()` gleich `DeadlineExceeded`“ zwischen Bestätigung und Eintrag bleibt grün; mit dem Endegrund {`Canceled`, `DeadlineExceeded`, eigener Grund} (36 Elemente) ist die Probe rot. `SPEC-135` Nr. 7 c nennt beide Kontextfehler. Jeder Mutant, der nicht auf den Grund prüft (`ctx.Err() != nil`, `errors.Is(…, Canceled)`: d2 bis d4), ist rot, darum LOW. | `AGENTS.md` §3.17, `SPEC-135` Nr. 14 | `src/hexagon/services/replay/replay.go` · `protokolliere(ctx, …)` ; Test `TestKontextNachDerBestaetigung` | Implementer: Endegrund als Achse der Schleife | ja: `mut2/d1` |
| V-6 | LOW | **`SPEC-135` Nr. 12 ist in einem Halbsatz nicht widerspruchsfrei** (Abschnitt 6): „kein Code für den Client“ neben `ODR-E9001`, und der Folgesatz „Das gilt für Antworten auf `HEAD`, 1xx, 204, 304 und leere Bodies“ bezieht sich nach der neuen Klausel rückwärts auf den Fehler. Das Verhalten ist durch Nr. 3 (g), 15 und `SPEC-109` eindeutig; der Adapter liest Nr. 12 und könnte „kein Code“ falsch lesen. | `SPEC-135` Nr. 12 | `spec/spezifikation.md` · `SPEC-135` Nr. 12 | Architect (Nachzug ist unterwegs): „kein Strom“ statt „kein Code für den Client“, und der Satz zu `HEAD` vor die Klausel zu `discarded` | ja: Text |
| V-7 | LOW | **Zwei Nachzüge beim Planner.** (a) Die DoD von `slice-lasttest` („der Match-Key-Index wird ohne Tabellenscan genutzt“, `LH-QA-01.a`) trifft der Replay nicht: der Dienst liest mit `ListBySession` die ganze Session und befragt den Index `(session_id, match_key)` nie (`SPEC-135` Nr. 8, Z60); Sendung 4 nennt die Folge `ListByMatchKey`, nicht diese DoD-Zeile. Beim Start von `slice-lasttest` ist die Zeile mit dem Replay abzugleichen. (b) §6 führt *Location mit Marker* als *Vorschlag, Owner bestätigt* (Plan §1: Entscheidung vom 2026-10-21, `SPEC-135` Nr. 13) und zugleich *Ausgang: offen, Adresse Owner für die Bestätigung*; beides zugleich gilt nicht, die Closure entscheidet, welches stimmt. | `AGENTS.md` §3.13, Plan §6 | `docs/plan/planning/open/slice-lasttest.md` DoD 1 ; Plan §6 | Planner | ja: Text |
| I-1 | INFO | Betriebsregel „kein `sed`, `head`, `tail`, `cp` auf Repo-Dateien“: keine Wirkung nachweisbar. Alle 27 Dateien unter `src/` sowie `spec/`, Plan und Review sind ohne BOM, NBSP, Zero-Width, CR und Steuerzeichen, enden mit Zeilenende, `gofmt` ist grün (`pruefung test`), der Arbeitsbaum ist sauber. | `AGENTS.md` §3.18 | – | – | ja: Abschnitt 2 |
| I-2 | INFO | `REPLAY_ASSIGNMENT=ordered` (Z62, Sendung 2): die Adresse `slice-bootstrap-m2` trägt den Punkt weder in §1 noch in der DoD; solange der User nicht entschieden hat, ist sie bedingt. Z62 nennt Grund und Folge. Die Entscheidung nennt der Plan unter *Offene Entscheidungen*. | `AGENTS.md` §3.13 | Plan §6 | User | ja: Text |
| I-3 | INFO | `AGENTS.md` §3.14 nennt noch die Faktoren 40 und 65 je andere Zeile; `SPEC-128` Nr. 1 und 2 legen 80 fest und sagen, dass kein Briefing eine Zahl der Faktoren trägt. Das gilt für alle Slices, nicht allein diesen. | `SPEC-128` Nr. 1 | `AGENTS.md` §3.14 | Planner | ja: Text |
| I-4 | INFO | Der Zweig liegt hinter `main`: der Plan `slice-odata-url-pfad` liegt hier noch in `in-progress/` (dort `done/`). Das löst der Integrationsschritt, nicht der Slice. | `AGENTS.md` §3.3 | `docs/plan/planning/` | Planner | ja: `git status` des Zweigs |

## 10. Antworten auf die Fragen des Auftrags

1. **Fixes F-1 bis F-9 gegen ihre Klassen.** F-1 geschlossen (1600 Elemente, 6 Mutanten rot). F-2 **teilweise** (Namen, Rand, Methode, Antwort-Header: 93 440 Fälle ohne Abweichung am Code; der Buchstabe ist nicht Teil der Menge, V-1). F-3 geschlossen für die Verbünde mit `ErrNichtGefunden` (3857 Fälle, nur V-4). F-4 geschlossen. F-5 **teilweise** (V-5, LOW; die Frage zu `DeadlineExceeded`: ja, Befund). F-6 geschlossen. F-7, F-8 geschlossen. F-9 geschlossen (Z65, Menge 8; 72 Fälle der Probe; drei Mutanten rot; Z66 nennt Grund und Folge).
2. **Nebenläufigkeit.** Abschnitt 4: n von n; Detektor 0 von 50 ohne Mutation; die Lücke ist V-2, nicht die Fenster.
3. **Schichten.** `make a-check` 0 Befunde; Dienst ohne Adapter-Import; Codetabelle 33; keine festen Zählstellen betroffen.
4. **Messung.** Abschnitt 8: stimmig.
5. **Wortlaut Nr. 12 und 15.** Abschnitt 6, V-6; Nr. 15 ist widerspruchsfrei.
6. **Offene Zeilen in §6.** Abschnitt 7: alle Adressen nehmen an; Nachzüge V-7, I-2.
7. **Betriebsregel.** I-1: keine Wirkung.

## 11. Ungeprüft

- `make gegenproben` und der Stempel (Auftrag).
- Der Detektor gegen die Fenster-Mutanten (nach `SPEC-122` Nr. 8 b nicht zu melden; allein die Wirkungs-Fänger tragen sie) und die Serialisierung der Transaktionen im SQLite-Adapter (Z63).
- Die Wortlaute von Z59 bis Z64 und Z66 sind auf Grund und Folge gelesen, nicht mutiert (es gibt keine Mutation).
- Das Protokoll (`SPEC-135` Nr. 16) nur über die Suite (Z40) und die Panik-Fakes des Probe-Pakets (kein Aufruf von `StartSpan`, `CountError`); eine eigene Mutation am Eintrag fuhr der Verifier nicht.
- Eine Panik eines Ports, ein Lesefehler des Blob-Stroms nach der Rückkehr (Z64): Negative.

## 12. Übergabe

Implementer: V-1 bis V-5 (je ein Fänger über die Klasse; der Bericht nennt je Befund die Menge, die Zahl ihrer Elemente und eine Mutation an einem Element, das der Befund nicht nannte, `AGENTS.md` §3.17; V-2 mit Rate je Prozess). Architect: V-6 (und V-4, falls Nr. 18 geändert wird). Planner: V-7, I-3, I-4. User: I-2.

Register-Kandidaten (nur genannt, kein Eintrag): *Fänger führt feste Namen statt des Alphabets der Klasse* (zu `universalaussage-im-vertrag-ungeprueft`); *Detektor blind, wenn jeder Test seinen eigenen Dienst baut* (zu `nebenlaeufiger-lauf-nur-im-glueck-gruen`); *Gegenstandsabhängige Fehlerzeilen: ein Fänger für eine von drei Zeilen* (zu `vertrag-sagt-mehr-zu-als-ein-gate-prueft`).
