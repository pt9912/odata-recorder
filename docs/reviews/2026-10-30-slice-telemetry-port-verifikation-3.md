# Verifikations-Report 3: slice-telemetry-port (Teil P der Kette P, L, S), Nachverifikation nach Fix-Runde 4 - 2026-10-31

**Rolle:** Verifier (Modul 11), dritte Runde, frischer Kontext - "Bauen wir es richtig?" gegen DoD, Plan und Spezifikation; nicht gegen den Bericht des Implementers, nicht gegen die Mutationslisten der Vorgänger.

**Gegenstand:** `slice-telemetry-port`, Zweig `slice-telemetry-port`, HEAD `4b062f1` (Fix-Runde 4 `abdd9a1`, danach zwei Planner-Commits `bce78b3`, `4b062f1`, die nur den Plan und zwei Dateien in `open/` ändern). Basis des Slice `a0396b0`.

**Gegen was geprüft:** Plan `docs/plan/planning/in-progress/slice-telemetry-port.md` (DoD, §3 Tabelle P01 bis P17, Fix-Runden 1 bis 4, Negativ N-1, §4 Block `groessenmarke` und Stand der Messung, §6, Raster), `SPEC-230` Nr. 1 bis 4 und 8 bis 10 sowie `SPEC-232` Nr. 1 und 5 a in `spec/spezifikation.md`, `AGENTS.md` §3.7, §3.10, §3.11, §3.13, §3.15, §3.17, Code unter `src/hexagon/ports/support/`, `src/hexagon/ports/support/supporttest/` und `src/adapters/driven/telemetry/`, die Pläne `slice-telemetry-spans` und `slice-telemetry-log` in `open/`; Vorberichte `docs/reviews/2026-10-30-slice-telemetry-port-*.md` nur zur Kenntnis.

**Modell:** claude-sonnet-5-5. **Arbeitsweise:** Sensoren selbst gefahren (Gate über `make`; Mutationen und Raten im Container `odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, jede Gruppe unter einer Sperre `flock /tmp/odata-recorder-heavy.lock`); Mutationen und Probe-Pakete ausschließlich auf Kopien im Scratchpad (`/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/verif-telemetrie-3/`), im Repo geändert ist nur dieser Report. Kein Container und keine wartende Sperre blieb zurück.

---

## Verdikt

**NICHT BESTANDEN** - wegen eines MEDIUM (V-15): Zwei Zusagen in Kommentaren des Ports nennen mehr, als die Tests prüfen (`AGENTS.md` §3.11), und eine Mutation je Zusage bleibt grün. Der Kern der Runde ist dagegen gelöst:

- **V-9 (Stand zu einem Zeitpunkt) ist gelöst.** Sieben Fenster-Mutationen an Stellen, die weder der Implementer noch ein Vorverifikator nannte (andere Form der Teilung als die 70 Schnitte: gerade/ungerade, ein einzelner Code am Anfang, in der Mitte und am Ende, drei Abschnitte, Dreiergruppen, Sperre je Verwurf), färbt `TestSnapshotZeitpunkt` in jedem Prozess: 50 von 50 ohne Detektor, 50 von 50 mit Detektor, `-test.count=50` rot, bei Last bis 11,1 (Load Average, §2). Die Gegenprobe bleibt in allen Zellen 0 von 50. Die Kopie nach der Freigabe der Sperre (A4) fällt in 50 von 50 Prozessen mit `fatal error` auf.
- **V-10 (Argumente des Fakes) ist im Kern gelöst**, ein Rest bleibt als LOW (V-18).
- **V-14 ist aufgelöst** (§2): die Rate gilt auch bei Last, die den Plan-Wert 5,4 übersteigt.
- Das Negativ N-1 stimmt in Grund, Folge im Betrieb und Adresse (§5).

Befunde: 0 HIGH, 1 MEDIUM, 5 LOW (davon V-5 aus der Vorrunde weiter offen), 3 INFO.

## 1. Sensor-Belege (selbst gefahren)

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` im Worktree | **Exit 0** (Exit-Code des Befehls in eine Datei geschrieben und aus der Datei gelesen, nicht über `tail`); Load Average vor dem Lauf 4,66, danach 4,45 | a-check `gesamt: 0 Befund(e)`; baseline-verify `v6.18.1 OK - 54 Dateien`; d-check `370 Datei(en) geprüft, 0 Befund(e)`; schema-check und migration-check ok; `groessenmarke slice-telemetry-port.md: ok, Messung 2647 von Marke 3000` |
| Cache im Gate-Lauf | Die Schichten von `test`, `race` und `lint` standen als `CACHED`; das Grün des Gates ist für sie kein neuer Lauf. Jede Stufe darum einzeln ohne Cache (`--no-cache-filter <stufe>`, Argumente wie im Makefile) | `gates.log` |
| Stufe `test` ohne Cache | Exit 0; Schicht `pruefung test /src` lief neu (69,1 s): `pruefung test: ok` | `stage-test.log` |
| Stufe `race` ohne Cache | Exit 0; Schicht `pruefung race /src` lief neu (107,5 s): `pruefung race: ok` | `stage-race.log` |
| Stufe `lint` ohne Cache | Exit 0; `0 issues.` (20,4 s) | `stage-lint.log` |
| Last der drei Stufenläufe | Load Average 4,94 zu Beginn, 5,41 am Ende | `stages.load` |
| `make groessenmarke` (im Gate) | `Messpunkt slice-telemetry-port.md: zeilen=17 nebenlaeufig=2 fakes=1 produktcode=580 schaetzung=2150 marke=3000 messung=2647` | Gate-Ausgabe |
| Zahlen des Plans nachgerechnet | `git diff --numstat a0396b0 -- src/hexagon/ports/support src/adapters/driven/telemetry` summiert auf 2647 (gleich mit `a0396b0..HEAD`); 2647/3000 = 0,88 und 2647/2150 = 1,23 wie in §4 des Plans; Schätzung 85*2 + 80*15 + 200*1 + 580 = 2150 und Marke 100*ceil(135*2150/10000) = 3000 stimmen; die Marke steht unverändert bei 3000 | eigene Summe |
| Zahlen der Tests nachgerechnet | 967 Aufrufe in `TestFakeArgumenteUngeprueft` = 1 + 96 + 82 + 738 + 41 + 9; 36 Zähler und 35 Fenster, 70 Schnitte = 35 * 2, 74 = 70 + 2 + 2: stimmen mit §3 | gelesen und gerechnet |
| `go.mod`, `go.sum`, Binary | `git diff a0396b0..HEAD -- go.mod go.sum src/bootstrap .a-check.yml AGENTS.md .claude .harness harness` leer; kein Paket außerhalb von `ports/support` und `adapters/driven/telemetry` importiert den Port oder den Adapter (`grep`), das Binary ist damit unverändert (Hash-Vergleich nicht neu gefahren, Vorbericht 2 hat ihn belegt und der Produktcode änderte sich seither nicht) | `git diff`, `grep` |
| Eigene Probe auf dem unveränderten Stand | die drei Pakete grün; meine 12 Probe-Tests grün (3 im Adapter, 4 in `ports/support`, 5 in `supporttest`), `go vet` ohne Meldung; Orakel sind damit nicht defekt | `out/basis.txt` |

Nicht gefahren, wie beauftragt: `make gegenproben` und der Stempel. `make docs-check` auf diesem Report: grün, siehe §9.

## 2. V-9 und V-14: "ein Stand zu einem Zeitpunkt" und die Last

**Menge (aus dem Wortlaut von `SPEC-230` Nr. 9, "liefert einen Stand zu einem Zeitpunkt"):** jede Teilung der 36 Zähler, die ein Aufruf dieses Teils ändert (33 Codes von `Errors` in Tabellenreihenfolge, `fehlercode`, `schluessel`, `wert`), in zwei Lesezeitpunkte. Eine Teilung der Kette in zwei nicht leere Seiten trennt mindestens ein Paar Nachbarn; die Menge der Formen ist damit nicht die der 70 Schnitte (zusammenhängende Teilung), sondern jede Teilung (nicht zusammenhängend eingeschlossen), dazu die Schreibseite des `Log`. Elementzahl der Klasse: 2^36 - 2 Teilungen; geprüft sind acht Vertreter, die sich in der Form der Teilung unterscheiden.

**Eigene Mutationen** (Kopie von `telemetry.go`, `Snapshot` bzw. `verwerfen` ersetzt; Id = meine):

| Id | Mutation | Form der Teilung |
|---|---|---|
| A1 | `Errors` der Codes mit geradem Index und `Dropped` unter einer Sperre, die Codes mit ungeradem Index unter einer zweiten | nicht zusammenhängend, 33 Nachbarpaare getrennt |
| A2 | Code Nr. 16 allein in einer zweiten Sperre | ein Element aus der Mitte |
| A6 | Code Nr. 0 allein in einer zweiten Sperre | ein Element am Anfang |
| A7 | Code Nr. 32 allein in einer zweiten Sperre | ein Element am Ende |
| A3 | drei Abschnitte: Codes 0 bis 10, 11 bis 21, 22 bis 32 mit `Dropped` (Sperre je Abschnitt) | zwei Schnitte in einem Stand |
| A5 | `Dropped` zuerst, dann die Codes in elf Dreiergruppen mit Sperre je Gruppe | elf Fenster |
| A8 | `verwerfen` mit Sperre je Verwurf (Schreibseite des `Log`) | Fenster zwischen `schluessel` und `wert` |
| A4 | `Snapshot` gibt die Sperre frei und ruft `Kopie()` danach | Lesen ohne Sperre |

**Eigenes Orakel** (`TestProbeStufenform`, nicht aus dem Code und nicht aus `luft` des Repo-Tests): ein einziger Schreiber zählt je Runde die 33 Codes in Tabellenreihenfolge, `CountError(0)` und ein `Log` mit zwei Verwürfen; ein Stand zu einem Zeitpunkt hat die Stufenform (n+1,...,n+1,n,...,n): nicht steigend, Sprung höchstens 1, höchstens ein Sprung, `schluessel` gleich `wert`. 4 Leser, 15000 Runden.

**Raten** (Messender: der Verifier; je Zelle 50 Prozesse mit `-test.count=1`, Test-Binary einmal gebaut: *plain* = `go test -c`, *Detektor* = `go test -race -c`; *Wiederholung* = ein Prozess mit `-test.count=50`; *Probe* = mein Orakel, plain). Zwei Durchgänge: **natürliche Last** (nur die Last der anderen Linien) und **Zusatzlast** (acht Rechenschleifen im selben Container, die sich die 4-CPU-Quote mit dem Test teilen). Die Spalte *Last* ist das Load Average (1 Minute, aus `/proc/loadavg` am Ende der Zelle; der Wert springt, weil die Schleifen in der Quote gedrosselt werden):

| Id | Durchgang | plain | Detektor | Wiederholung | Probe | Last (plain / Detektor / Wiederholung / Probe) |
|---|---|---|---|---|---|---|
| Gegenprobe | natürlich | 0 von 50 | 0 von 50 | grün | 0 von 50 | 3,5 / 2,7 / 3,0 / 3,0 |
| Gegenprobe | Zusatzlast | 0 von 50 | 0 von 50 | grün | 0 von 50 | 2,6 / 8,6 / 10,1 / 6,8 |
| A1 | natürlich | 50 von 50 | 50 von 50 | rot | 50 von 50 | 3,1 / 3,3 / 3,0 / 3,0 |
| A2 | natürlich | 50 von 50 | 50 von 50 | rot | 50 von 50 | 2,7 / 2,9 / 2,7 / 2,6 |
| A3 | natürlich | 50 von 50 | 50 von 50 | rot | 50 von 50 | 2,6 / 3,0 / 2,9 / 3,9 |
| A5 | natürlich | 50 von 50 | 50 von 50 | rot | 50 von 50 | 3,9 / 4,1 / 4,1 / 4,1 |
| A1 | Zusatzlast | 50 von 50 | 50 von 50 | rot | 50 von 50 | 2,7 / 10,3 / 7,0 / 2,9 |
| A5 | Zusatzlast | 50 von 50 | 50 von 50 | rot | 50 von 50 | 6,2 / 10,4 / 10,2 / 3,9 |
| A6 | Zusatzlast | 50 von 50 | 50 von 50 | rot | 50 von 50 | 2,1 / 11,1 / 5,0 / 5,7 |
| A7 | Zusatzlast | 50 von 50 | 50 von 50 | rot | 50 von 50 | 10,4 / 4,3 / 2,8 / 13,0 |
| A8 | Zusatzlast | 50 von 50 | 50 von 50 | rot | 50 von 50 | 10,1 / 4,8 / 10,7 / 12,5 |
| A4 | natürlich (Last 11,0) | 50 von 50 (`fatal error`, kein `DATA RACE`) | 50 von 50 (`fatal error`) | - | - | `TestSnapshotZeitpunkt` und `TestAdapterGleichzeitig` je plain und Detektor: 4 Zellen, je 50 von 50 |

In keiner Zelle meldet der Detektor `DATA RACE` bei A1 bis A3 und A5 bis A8 (jeder Zugriff liegt unter einer Sperre): die Wirkung trägt der Wirkungs-Fänger, wie der Plan sagt.

**Urteil V-9:** gelöst. Die Menge "jede Teilung" ist durch die Kette aller 36 Zähler und beide Leseordnungen gedeckt; die sieben Vertreter (A1 bis A3, A5 bis A8) mit fünf verschiedenen Formen der Teilung färben den Fänger in 900 von 900 Prozessen (natürliche Last: 4 Mutationen * (50 plain + 50 Detektor) = 400; Zusatzlast: 5 Mutationen * 100 = 500), die Gegenprobe in 0 von 200; A4 (Lesen ohne Sperre) in 200 von 200.

**Urteil V-14 (Adresse Verifier, §6 des Plans):** Die Raten gelten bei Load Average **bis 11,1** an den Zellen der Repo-Tests (Detektor A6), 13,0 an einer Zelle meiner Probe; der Plan nannte 5,4. Über 13,0 habe ich nicht gemessen (die Host-Last anderer Linien liegt in diesen Läufen zwischen 2 und 11). Vorschlag an den Planner: V-14 **entfallen** mit der Zahl "50 von 50 bei Load Average 2,0 bis 11,1 (Verifier, drei Durchgänge, 8 Mutationen)".

**Mutationen der Sperren in Adapter und Fake** (eigene, nicht aus den Vorberichten; Detektor und plain je 50 Prozesse, Zusatz `-count=50` ein Prozess; Last am Zellenende 6,8 bis 11,1):

| Id | Mutation | Detektor | plain | Wiederholung |
|---|---|---|---|---|
| L1 | Fake: `StartSpan` zählt die ID (`naechste++`) ohne Sperre | 50 von 50, 50 mit `DATA RACE` | 0 von 50 | grün |
| L4 | Fake: `Snapshot` liest `gesetzt` und `stand` ohne Sperre | 50 von 50, 50 mit `DATA RACE` | 9 von 50 | rot |
| L7 | Fake: `aufzeichnen` zählt `zaehler[op]++` nach der Freigabe | 50 von 50, 50 mit `DATA RACE` | 50 von 50 | rot |
| L8 | Fake: `aufzeichnen` liest die Länge, gibt frei, hängt danach an (verlorene Aufzeichnung) | 50 von 50, 0 mit `DATA RACE` | 50 von 50 | rot |
| Gegenprobe | unverändert | 0 von 50 | 0 von 50 | grün |

L1 und L4 zeigen: das Fehlen der Sperre fängt allein der Detektor in jedem Prozess (50 von 50); ohne Detektor ist die Wirkung ein Messwert (0 und 9 von 50), wie der Plan es für R-2 und R-7 sagt. L8 ist ein Wirkungs-Fänger (Fenster, der Detektor sieht nichts) und wird in jedem Prozess rot.

## 3. Klasse "Kopie" und "Fake": eigene Mutationen

**Kopie je Sammlung** (`Snapshot.Kopie()` in `telemetry.go`, 12 Mutationen, die übersetzen; `KSh5` nicht, zählt nicht): Errors nil zu leer (KE1), Errors leer zu nil (KE2), Dropped nil zu leer (KD1), Dropped leer zu nil (KD2), Spans nil zu leer (KS1), Spans leer zu nil (KS2), Durations nil zu leer (KU1), Durations leer zu nil (KU2), Errors geteilt (KSh1), Dropped geteilt (KSh2), `Counts` nie geklont (KSh4), `Durations[n]` als ganzer Wert aus dem Original (KSh6): **alle 12 rot** in der bestehenden Suite (`TestSnapshotKopie`, `TestSnapshotKopieBeispiel`, teils `TestFakeSnapshotStand`, `TestSnapshotVollstaendig`). Die Klasse der Sammlungen ist gedeckt.

**Tiefe der Zustände** (Länge von `Attrs` und `Counts`): die Menge der Zustände im Test ist nil, leer, 1, 2, 12. Mein Generator (`TestProbeKopieKlasse`, 3000 Stände, Längen 0 bis 14 je Eintrag, bis zu 8 Einträge je Sammlung, Orakel zwei gleich gebaute Stände) färbt S1 (Kopie lässt `Attrs` der Länge 3 geteilt) und S1b (dasselbe für `Counts`); die bestehende Suite bleibt grün (V-17).

**Fake** (Kopie von `telemetrie.go`): siehe Tabelle in §4; die bestehende Suite fängt L1, L4, L7, L8 (Detektor bzw. Wirkung), K14 (`Aufrufe()` teilt den äußeren Slice), sie fängt nicht K1 bis K1d, K3b, K3c, K6, K12, K4, K4b.

## 4. Alle eigenen Mutationen (46 gültige; 17 in der bestehenden Suite grün)

Läufe: `go test -count=1` über die drei Pakete, plain, einmal je Mutation, mit der bestehenden Suite (ohne meine Probe) und mit meiner Probe; Last 3,2 bis 5,1. *bestehend* = rot/grün der bestehenden Tests, *Probe* = meine Tests.

| Id | Mutation | bestehend | Probe |
|---|---|---|---|
| A1 bis A8 | Fenster (§2) | rot | rot |
| A17 | unbekannter Code legt einen Schlüssel in `Errors` an (`+= 0`) | rot, aber nur weil der nächste Aufruf desselben Codes als bekannt zählt: kein Fänger der Reihenzahl | rot |
| A19 | `Log` zählt höchstens die ersten zwei Verwürfe | **grün** | grün |
| A20 | `Log` legt für jeden fremden Schlüssel einen Eintrag mit 0 in `Dropped` an | **grün** | rot |
| K1 | Fake: `SetzeStand` kopiert nur `Errors` | **grün** | rot |
| K1b | Fake: `Snapshot` liefert `Durations`, `Dropped`, `Spans` geteilt | **grün** | rot |
| K1c | Fake: `SetzeStand` teilt `Counts` und `Attrs` | **grün** | rot |
| K1d | Fake: `Snapshot` teilt `Counts` und `Attrs` | **grün** | rot |
| K3b | Fake: `Aufruf.Eltern` ist nil, wenn der Eltern selbst Eltern hat (dritte Ebene) | **grün** | rot |
| K3c | Fake: `FakeSpan.Eltern` ebenso | **grün** | rot |
| K6 | Fake: der `ctx` von `StartSpan` verliert die Frist (Abbruch und Werte bleiben) | **grün** | rot |
| K12 | Fake: `Anzahl` ohne die Prüfung `op < 0` (Panik bei -1) | **grün** | rot |
| K4 | Fake: `Log` hält nur die ersten 16 Attribute fest | **grün** | rot |
| K4b | Fake: `Set` hält nur die ersten 16 Attribute fest | **grün** | rot |
| K14 | Fake: `Aufrufe()` teilt den äußeren Slice | rot (`TestFakeAufzeichnung`) | grün |
| S1 | `Kopie`: `Attrs` der Länge 3 geteilt | **grün** | rot |
| S1b | `Kopie`: `Counts` der Länge 3 geteilt | **grün** | rot |
| S3 | `Spans()`: `Pflicht` des fünften Spans aus einer Paketvariable | **grün** | rot |
| S4 | `Verwuerfe()` liefert eine Paketvariable | **grün** | rot |
| S5 | `Millis` teilt auf Mikrosekunden gekürzt | rot (`TestAttrKonstruktoren`) | rot |
| S6 | `ErrorCode` liefert für `CodeE9001` den leeren Text | **grün** | rot |
| R1 | Wertform Zahl über `int32` | rot (`TestWertformGrenzen`) | grün |
| R3 | Wertform Dauer endlich nur bis `MaxFloat32` | rot (`TestWertformGrenzen`) | grün |
| KE1 bis KSh6 | Kopie je Sammlung (§3), 12 | rot | rot |
| L1, L4, L7, L8 | Sperren des Fakes (§2) | Detektor rot, L8 auch plain | - |

17 Mutationen bleiben in der bestehenden Suite grün: A19, A20, K1, K1b, K1c, K1d, K3b, K3c, K6, K12, K4, K4b, S1, S1b, S3, S4, S6. A19 färbt meine Probe nicht (ich habe keinen Fänger für drei Verwürfe geschrieben); die übrigen 16 färbt sie. Die Befunde V-15 bis V-19 ordnen diese 17 nach Zusage und Schwere.

## 5. N-1 und die Adresse an `slice-telemetry-spans`

**Grund.** Der Code ändert in diesem Teil `Durations` und `Spans` nicht: `New` legt `Counts` mit Nullen an, kein Pfad schreibt sie (`StartSpan` gibt `keinSpan{}` zurück), `Snapshot` legt `Spans` leer an. Die Gründe senke, pflicht, spanname und zustand zählt kein Aufruf: `Log` ruft `verwerfen` allein mit dem, was `filtern` liefert (`schluessel`, `wert`), `CountError` zählt allein `fehlercode`. Ein Fenster der Sperre um sie liefert denselben Stand: **bestätigt**, N-1 ist äquivalent (ein Fenster allein um `Durations` oder `Spans` habe ich nicht neu gefahren; der Befund am Code genügt, weil kein Pfad den Wert ändert).

**Folge im Betrieb.** "Keine, solange `StartSpan` ohne Wirkung ist und kein Aufruf die vier Gründe zählt": stimmt; im Produkt ruft kein Dienst den Port (§1), die Folge ist also leer. Sobald `slice-telemetry-spans` `Durations` füllt, gilt die Klasse für sie - der Plan sagt das und weist den Fänger zu.

**Adresse.** `slice-telemetry-spans` liegt in `docs/plan/planning/open/`. Gelesen: §1 *Übernimmt (Sendung von `slice-telemetry-port`, Klasse Stand zu einem Zeitpunkt)* nennt Geber, die 36 Zähler, die unbeobachtbaren Fenster und den Fänger über die Kette der vier Sammlungen einschließlich der vier Gründe; die DoD (zweiter Punkt) trägt "die Lücken zwischen `Errors`, `Durations`, `Dropped` und `Spans`, einschließlich der Gründe senke, pflicht, spanname, zustand; Geber `slice-telemetry-port`"; die Zeile S16 führt "die Kette der Zähler über `Errors`, `Durations`, `Dropped` und `Spans` (Sendung von `slice-telemetry-port`, §1)". Kein Punkt unter *Ausdrücklich NICHT* trifft die Sendung (der erste nennt `CountError` als Code von `slice-telemetry-port`; der Fänger ruft `CountError` nur auf, er ändert ihn nicht). Auch F-10 (Ersatz des Teils von `TestKernSchreibtNichts`) und V-8 stehen als *Übernimmt* bzw. *Hinweis* in `slice-telemetry-spans`, F-10 zusätzlich in `slice-telemetry-log` (§1, Geber benannt). **Die Sendung hat eine Adresse** (`AGENTS.md` §3.13).

## 6. Befunde

Zuständigkeit: Implementer = Test oder Code im Slice; Planner = Plan und Adresse; Architect = Entscheidung in der Spezifikation; User = Entscheidung des Owners.

| Nr. | Schwere | Zuständig | Befund |
|---|---|---|---|
| **V-15** | **MEDIUM** | Implementer | **Zwei Zusagen in Kommentaren nennen mehr, als der Test prüft (`AGENTS.md` §3.11).** (a) `FakeTelemetry.StartSpan` (`supporttest/telemetrie.go`, Kommentar): "den übergebenen ableitet (Abbruch, Frist und Werte bleiben)". Menge: Abbruch, Frist, Werte = 3 Elemente; `TestFakeEltern` "Wert und Abbruch bleiben" prüft 2, die Frist nicht. Mutation K6 (der Rückgabe-`ctx` verliert `Deadline`, Abbruch und Werte bleiben) bleibt grün (bestehende Suite 0 Fehler), mein `TestProbeFakeCtx` wird rot. Es ist dieselbe Klasse wie die Betriebsregel "eine Hülle um ein Argument verdeckt keine Eigenschaft des Originals" (§3 des Plans). (b) `support.Spans()` (Kommentar): "jeder Aufruf liefert frische Slices". Menge: 1 äußerer Slice und 5 `Pflicht`-Slices = 6; `TestSpansDaten` verändert allein `got[0].Pflicht[0]` (1 von 6). Mutation S3 (`Pflicht` des fünften Spans aus einer Paketvariable) bleibt grün; `Eimer()` ist gedeckt (alle Elemente liegen in einem Slice). Fix gegen die Klasse: der Test von (a) prüft die Frist (`context.WithDeadline` im Eingabe-`ctx`, `Deadline()` des Rückgabe-`ctx` gleich); der Test von (b) ändert jedes Element jeder Ebene (Schleife über 5 Spans, jeder `Name` und jedes `Pflicht`-Element) oder der Kommentar sagt weniger. Wer weitere Kommentare mit "jeder", "alle", "bleiben" im Paket hat, schreibt sie gegen ihren Test auf. |
| V-16 | LOW | Implementer | **Fake: `SetzeStand` und `Snapshot` kopieren nach `SPEC-230` Nr. 10 "als Kopie"; der Fänger prüft `Errors`.** Menge: 6 Sammlungen (`Errors`, `Durations`, `Dropped`, `Spans`, `Counts`, `Attrs`) * 2 Richtungen (beim Setzen, beim Lesen) = 12; `TestFakeSnapshotStand` ändert allein `Errors` (2 von 12). Mutationen K1, K1b, K1c, K1d (je eine Richtung, nur `Errors` bzw. nur oben geklont) bleiben grün, `TestProbeFakeStandTief` färbt alle vier. Der Code ruft eine Zeile `Kopie()`, die an anderer Stelle tief geprüft ist; die Mutation ist die handgebaute Teilkopie, deshalb LOW. Fix: der Test ruft `kopieAendern`-artig alle sechs Sammlungen in beide Richtungen. |
| V-17 | LOW | Implementer | **Die Menge der Zustände von `Attrs` lässt die reale Länge 3 aus.** `SPEC-232` Nr. 4 begrenzt den Ring auf "höchstens drei Attribute"; `TestSnapshotKopie` führt für `Attrs` und `Counts` die Zustände nil, leer, 1, 2, 12 (Zählprüfung 5). Mutation S1 (`Attrs` der Länge 3 geteilt) bleibt grün, mein Generator mit Längen 0 bis 14 färbt sie (200 Einträge mit `Attrs` der Länge 3 in 3000 Ständen). `Counts` hat im Betrieb stets 12 (gedeckt). Fix: 3 in `eintragLaengen` aufnehmen (6 Zustände, 36 Paare) oder die Prüfgrenze benennen (§3.17: unbeschränkte Menge, Grund und Folge im Betrieb). |
| V-18 | LOW | Planner / Implementer | **Prüfgrenzen der unbeschränkten Mengen des Fakes fehlen im Plan (`AGENTS.md` §3.17).** Der Fänger von P11 führt Attributlisten bis Länge 12, Eltern-Tiefe 1 und 2 (`TestFakeEltern`), `Anzahl` für 0 und 7. Nicht benannt sind die Grenzen: (i) Liste länger als 12 (K4, K4b: 16 abgeschnitten, grün); (ii) Tiefe der Eltern-Kette ab 3 (K3b, K3c: grün); (iii) `Anzahl(-1)` und jeder Wert unter 0 (K12: die Prüfung `op < 0` ist entfernbar, grün). Die Zeile F-11 des Plans nennt "Anzahl außerhalb" als Menge, geprüft sind 0 und `OpSnapshot+1`, nicht die untere Seite. Folge im Betrieb: keine, der Fake dient Tests; ein Dienst-Test mit mehr als 12 Attributen oder einem Enkel-Span bekäme eine falsche Aufzeichnung. Entweder Fall je Grenze (Liste 13 und 41, Tiefe 3, `Anzahl(-1)`) oder ein akzeptiertes Negativ mit Grund und Folge im Plan. |
| V-19 | LOW | Implementer | **"Die Zahl der Reihen ist fest; kein Wert eines Aufrufers bildet einen Schlüssel" (`SPEC-230` Nr. 8) hat nur am frischen Adapter einen Fänger.** `TestSnapshotVollstaendig` zählt 33, 5, 7 nach `New`; nach Aufrufen mit fremdem Schlüssel oder unbekanntem Code zählt kein Test die Schlüssel. Mutation A20 (`Log` legt je fremdem Schlüssel einen Eintrag mit 0 an) bleibt grün, `TestProbeFesteReihen` (400 Aufrufe mit zufälligen Codes und Schlüsseln, danach 33, 7, 5) färbt sie. Eine Variante mit Zuwachs ungleich 0 fängt `TestVerwurfRangfolge` über die Schleife "kein anderer Grund ungleich 0" bzw. P07; die Lücke ist die Mutation mit 0. Fix: `len(Errors) == 33`, `len(Dropped) == 7`, `len(Durations) == 5` nach den Aufrufen in P07 und P17. |
| V-5 (weiter offen) | LOW | Implementer / Planner | **`Log` zählt bei mehr als zwei Verwürfen je Aufruf**: A19 (nur die ersten zwei gezählt) bleibt grün, P17 führt höchstens zwei Verwürfe je Aufruf. Der Plan trägt V-5 als *weiter offen* (Adresse Implementer, optional). `SPEC-230` Nr. 4 sagt, dass jedes verletzende Attribut zählt; die Zahl der Attribute je Aufruf ist unbeschränkt (13 Schlüssel plus Dubletten). Ich führe sie wegen der dritten Runde nach der Nennung als LOW, nicht als INFO: ein Fall mit drei und mit 13 verletzenden Attributen schließt sie, oder der Plan benennt die Prüfgrenze. |
| V-20 | INFO | Planner | **V-14 ist aufgelöst** (§2): 50 von 50 bei Load Average bis 11,1 (13,0 an der Probe). §6 des Plans kann V-14 mit Ausgang *entfallen* und dieser Zahl führen; der Satz im Absatz "Raten der Fenster-Fänger" ("gelten für Load Average bis 5,4") wird dann auf die neue Obergrenze gezogen. |
| V-21 | INFO | Planner | **Zwei Aussagen zur Last stehen gegeneinander in §3.** Der Absatz nach der Tabelle der Fix-Runde 3 sagt, die Rate sei "unter Load Average 4,7 bis 9,2" gemessen, der Absatz der Fix-Runde 4 sagt "bis 5,4"; die Tabelle der Fix-Runde 3 gehört zum ersetzten Orakel (der Kopf nennt es). Kein Fehler im Code; die ältere Messung steht als Chronik neben der geltenden. Der Planner kann sie in der Closure entfernen oder als ersetzt kennzeichnen. |
| V-22 | INFO | Implementer | **`support.Verwuerfe()` liefert (anders als `Spans()` und `Eimer()`) keinen Hinweis auf frische Slices**, und kein Test prüft es (Mutation S4: Paketvariable, grün; mein `TestProbeFrisch` färbt sie). Weder die Spezifikation noch der Kommentar sagen etwas zu; `New` und `Snapshot` rufen die Funktion. Eine Entscheidung ist nicht nötig; wer sie als frisch zusagt, schreibt den Kommentar und den Fall. S6 (`ErrorCode` für den letzten Code leer) ist eine künstliche Mutation, weil `ErrorCode` keine Verzweigung nach dem Code hat; geprüft sind 1 von 33 Codes plus drei außerhalb, ohne Folge. |

Gelöst gegenüber Vorrunde (kein Befund): V-9, V-10 (Rest siehe V-16, V-18), V-12 (Zahl der Messung: 2647 in §4, Gate und eigener Summe gleich), V-13 (Adresse `slice-telemetry-spans` getragen, §5), V-14 (Messung). Offen und nicht neu bewertet: V-7 (Nullwert des Adapters, Adresse Architect), V-8 (Hinweis an `slice-telemetry-spans`).

## 7. Konformität gegen die DoD des Plans

| DoD-Punkt | Stand |
|---|---|
| 1 Ports, Typen, Fake | erfüllt bis auf V-15 bis V-18 (Fänger, nicht Verhalten): Operationen, 13 und 11 Schlüssel, neun Wertformen, Rangfolge, `Spans()`, `Eimer()` entsprechen `SPEC-230` Nr. 1 bis 4 und 10; die Klasse der Wertformen haben die Vorrunden aufgelöst, ich habe R1 und R3 an ihr gefahren (beide rot) |
| 2 Kern des Adapters | erfüllt: `New` (9 Fälle, `TestNewUnabhaengig`), `CountError` (33 + 4/5 Codes), `Snapshot` vollständig und Kopie, `Log` ohne Zeile, `StartSpan` ohne Wirkung, unter `make test-race` ohne Befund (Gate und Stufe ohne Cache) |
| 3 `make gates` grün | erfüllt (Exit 0; `test`, `race`, `lint` zusätzlich ohne Cache) |
| 4 Review durchgeführt | erfüllt: zwei Reports (`2026-10-30-slice-telemetry-port-review.md`, `-review-2.md`) |
| 5 Doku-Update | erfüllt: `SPEC-230` bis `SPEC-232`; Binary unverändert (§1) |
| 6 Handbuch | entfällt, Binary unverändert (kein Importeur des Ports) |
| 7 Closure-Notiz, 8 Register, 9 Ausgänge in §6, 10 Paarungen | Aufgabe des Planners nach dieser Verifikation; §7 des Plans trägt "Wird bei der Closure gefüllt" |

Messpunkt-Zeile (`SPEC-128` Nr. 9): Handoff-Wert 2647 nachgerechnet und gleich dem Gate; die Ende-Zeile setzt der Planner.

## 8. Gegenprobe der Orakel und Grenzen der Aussagen

- Meine Orakel sind unabhängig vom Code: `TestProbeStufenform` (Stufenform, Einzelschreiber), `TestProbeKopieKlasse` (zwei gleich gebaute Stände aus derselben Saat, 3000 Stände), `TestProbeMillis` (`float64(d)/1e6`, 2010 Werte), `TestProbeErrorCodeAlle` (33 Codes gegen `Code.String()`), `TestProbeFrisch`, `TestProbeFakeLaengen` (Längen 0 bis 40), `TestProbeFakeTiefe` (Tiefe 1 bis 8), `TestProbeFakeCtx`, `TestProbeFakeStandTief`, `TestProbeFakeAnzahl` (-5 bis 12), `TestProbeFesteReihen`, `TestProbeSnapshotGetrennt`. Alle grün auf dem unveränderten Stand; jede rot unter mindestens einer Mutation (Tabelle §4).
- Das Load Average ist ein Host-Wert und schwankt in den Zellen stark (2,0 bis 13,0); Zeilen nennen den Wert am Zellenende. Bei Zusatzlast wirkt die Quote des Containers (8 Rechenschleifen in 4 CPUs), nicht allein das Load Average. Mehr als 13,0 habe ich nicht gemessen.
- Die Zahl der Mutationen (46 gültige: 27 in der ersten Gruppe, 12 in der Kopie-Gruppe, A6 bis A8, L1, L4, L7, L8) zählt nicht, was nicht übersetzte (KSh5).

## 9. Ungeprüft und Gegengelesen

**Ungeprüft:** `make gegenproben` und `harness/gegenproben.stempel` (beauftragt nicht gefahren); der Binary-Hash (nicht neu verglichen, §1); die Wertformen P03 bis P05 über die Stichproben R1 und R3 hinaus (die Vorrunden haben die Klassen aufgelöst, 61440 und 28672 Werte); die Fenster um `Durations` und `Spans` (N-1, unbeobachtbar); die Raten der Fake-Tests `TestFakeGleichzeitig`, `TestFakeSnapshotGleichzeitig` und `TestFakeLesenGleichzeitig` gegen die Mutationen des Implementers (nur meine L1, L4, L7, L8); Last über 13,0; die Zusatzzeit der Stufen `test` und `race` gegen die Entscheidung über den Re-Evaluierungs-Trigger von [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Entscheidung des Users).

**Gegengelesen:** dieser Report auf Rohbytes (`LC_ALL=C grep -P`: keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP) und mit `make docs-check` im Worktree (Exit 0, `d-check: 371 Datei(en) geprüft, 0 Befund(e)`, Report als nicht versionierte Datei eingeschlossen); die Rohbytes-Prüfung fand 0 Treffer für die Escape-Folge, das BOM und das NBSP.

## 10. Vorbedingungen der Closure (nur genannt, nicht gesetzt)

Fix-Runde 5 für V-15 (MEDIUM) und nach Wahl V-16 bis V-19; Planner zieht §3 (Absatz Fix-Runde 5), §4 (Stand der Messung) und §6 (V-14, V-16 bis V-19) im selben Commit nach (`AGENTS.md` §3.9); danach Closure-Notiz mit Lerneintrag, Messpunkt-Zeile Ende, Register.

## 11. Register-Kandidaten (genannt, nicht geschrieben)

- Klasse "Kommentar sagt mehr zu, als der Test prüft", Elemente einer Aufzählung im Kommentar (Abbruch, Frist, Werte; jedes Slice-Element) - zu `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` gehörig, vom Zähler des Registers zu klären.
- Klasse "handgebaute Teilkopie in einem Fake": Fake kopiert eine von vier Sammlungen und der Test ändert dieselbe (V-16) - zu `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` gehörig.
- Klasse "Zusatzlast im selben Container statt Host-Last": Messung der Rate bei Last ohne Eingriff in andere Linien.
