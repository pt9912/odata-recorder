# Verifikations-Report 2: slice-telemetry-port (Teil P der Kette P, L, S), Nachverifikation nach Fix-Runde 3 - 2026-10-30

**Rolle:** Verifier (Modul 11), zweite Runde - "Bauen wir es richtig?" gegen DoD und Spezifikation, nicht gegen den Bericht des Implementers und nicht gegen den Diff.

**Gegenstand:** `slice-telemetry-port`, Zweig `slice-telemetry-port`, HEAD `bb4e065` (Fix-Runde 3 auf `f472d90`). Der Commit ändert nur Tests und den Plan: `git diff ed3142b..bb4e065 --stat` nennt vier Dateien (`telemetry_test.go` des Adapters, `telemetry_test.go` und `supporttest/telemetrie_test.go` des Ports, den Plan), keinen Produktcode. Die Probe-Ergebnisse der ersten Runde zum Produktcode gelten damit unverändert und sind neu gefahren.

**Gegen was geprüft:** Plan `docs/plan/planning/done/slice-telemetry-port.md` (§3 P09, P11, P15, P16, Absatz Fix-Runde 3, Raten, Negativ N-1; §4; §6), `SPEC-230` Nr. 9 und 10 in `spec/spezifikation.md`, `AGENTS.md` §3.7, §3.10, §3.13, §3.15, §3.17; Vorverifikation `docs/reviews/2026-10-30-slice-telemetry-port-verifikation.md` (V-1 bis V-8).

**Modell:** claude-sonnet-5-5 - **Datum:** 2026-10-30 - **Arbeitsweise:** Sensoren selbst gefahren (Container `odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, jeder schwere Lauf unter `flock /tmp/odata-recorder-heavy.lock`, gebündelt: eine Sperre je Gruppe); Mutationen und Probe-Pakete ausschließlich auf Kopien im Scratchpad (`.../scratchpad/verif-telemetrie-2/`), im Repo geändert wurde nur dieser Report. Die Sperre war überlaufen (bis zu 25 Minuten Wartezeit je Gruppe); die Messungen sind deshalb knapper als geplant (§9).

---

## Verdikt

**NICHT BESTANDEN** - wegen eines MEDIUM (V-9): Der Wirkungs-Fänger `TestSnapshotZeitpunkt` aus Fix-Runde 3 deckt die Klasse "ein Stand zu einem Zeitpunkt" (`SPEC-230` Nr. 9) nur zum Teil. Von zehn nicht-äquivalenten Fenster- und Sperren-Mutationen an Stellen, die weder der Implementer noch die Vorverifikation nannten (Z1 bis Z9, Z11), bleiben drei in jeder Spalte grün (0 von 50 Prozessen mit Detektor, 0 von 50 ohne, `-test.count=50` grün); ein Orakel mit einer zweiten Kette sieht alle drei in 50 von 50 Prozessen rot (Gegenprobe 0 von 50). Eine davon (Z1: Lücke zwischen den Verwürfen `fehlercode` und `schluessel`) liegt nicht im benannten Negativ N-1, denn beide Zähler ändert der Test. Der Code ist konform; es ist wieder ein Fänger-Defizit, und der Fix ist klein (§5).

Alles andere, was dem Verifier gehört, ist belegt: `make gates` Exit 0, die Stufen `test`, `race` und `lint` ohne Cache grün, Marke 3000 unverändert bei Messung 2551, V-2 (Position x Zustand) und V-4 (Rang-Zeiger) geschlossen, V-3 bis auf eine Restmenge der Argumente (V-10, LOW).

Befunde: 0 HIGH, 1 MEDIUM, 1 LOW, 4 INFO.

## 1. Sensor-Belege (selbst gefahren)

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` im Worktree | **Exit 0** (Exit-Code des Befehls `flock ... bash -c "make gates > Datei; echo $? > Datei"`, aus der Datei gelesen, nicht über `tail`) | a-check `gesamt: 0 Befund(e)`; baseline-verify `v6.18.1 OK - 54 Dateien`; d-check `369 Datei(en) geprüft, 0 Befund(e)`; schema-check und migration-check ok; `groessenmarke slice-telemetry-port.md: ok, Messung 2551 von Marke 3000` |
| Cache der Stufen im Gate-Lauf | Die Schichten von `test`, `race`, `lint` standen im Gate-Lauf als `CACHED` (Quellstand wie beim Implementer). Das Grün des Gate-Laufs ist damit kein neuer Lauf dieser Stufen; ich habe jede Stufe darum einzeln ohne Cache gefahren (nächste drei Zeilen). | `gates.log` |
| Stufe `test` ohne Cache | Exit 0; `docker build --no-cache-filter test --target test`; Schicht `pruefung test /src` lief neu (70 s, nicht `CACHED`): `pruefung test: ok` | `stage-test.log` |
| Stufe `race` ohne Cache | Exit 0; `--no-cache-filter race`, `RACE_IMAGE` wie im Makefile; Schicht `pruefung race /src` lief neu (107 s): `pruefung race: ok` | `stage-race.log` |
| Stufe `lint` ohne Cache | Exit 0; `--no-cache-filter lint`, `LINT_IMAGE` wie im Makefile; golangci-lint: `0 issues.` | `stage-lint.log` |
| `make groessenmarke` (im Gate) | `slice-telemetry-port.md: ok, Messung 2551 von Marke 3000`; Messpunkt `zeilen=17 nebenlaeufig=2 fakes=1 produktcode=580 schaetzung=2150 marke=3000 messung=2551` | Gate-Ausgabe |
| Messung nachgerechnet | `git diff --numstat a0396b0..HEAD` über die zwei Pfade des Blocks: 1286 (Adapter) + 1265 (Port) = 2551. Auslastung 2551/3000 = 0,85; 2551/2150 = 1,19 der Schätzung. Die Marke steht unverändert bei 3000 (Block `groessenmarke` im Plan). | eigene Summe |
| Eigene Läufe (Basis `bb4e065`) | `go test -count=1` der drei Pakete (ohne Probe) grün; mein Probe-Paket (§4) plain und unter `-race` grün; die Stufe `race` oben fährt alle Pakete unter dem Detektor | `probes.out` |
| Rohbytes (`AGENTS.md` §3.18) | dieser Report enthält keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP (`LC_ALL=C grep -P` vor dem Commit); die Probe-Dateien liegen im Scratchpad und sind nicht committet | §9 |

Nicht gefahren: `make gegenproben` und der Stempel (Integrationsschritt, laut Auftrag), `make docs-check` auf diesem Report (der Gate-Lauf d-check war grün auf dem Stand vor dem Report).

## 2. V-1: "ein Stand zu einem Zeitpunkt" - gegen die Klasse aufgelöst

**Menge (aus dem Wortlaut von `SPEC-230` Nr. 9: "`Snapshot()` liefert einen Stand zu einem Zeitpunkt"):** jedes Paar von Zustandsgrößen, die ein Stand nicht zu verschiedenen Zeiten lesen darf. In diesem Teil ändern drei Stellen den Zustand: `CountError` (bekannter Code: `Errors[c]`; unbekannter: `Dropped[fehlercode]`) und `Log` (bis zu zwei Verwürfe in `Dropped`). Die Menge der lesbaren Größen: 33 Codes in `Errors`, 7 Gründe in `Dropped`, 5 Verteilungen, `Spans`. Fenster = Freigabe der Sperre zwischen dem Lesen (oder Zählen) zweier Größen.

**Eigene Fenster-Mutationen** (Id = meine; A bis H und G stammen vom Implementer, W1 bis W7 vom Verifier der ersten Runde und sind hier nicht wiederholt; die Z-Mutationen liegen an Stellen, die keiner nannte). Mutationen als Kopie des Quellstands; Spalte *Repo* = `TestSnapshotZeitpunkt`, jeweils Detektor / ohne Detektor (je 50 Prozesse, `-test.count=1`) / ein Prozess `-test.count=50`; Spalte *Probe* = mein `TestProbeStand2` (zweite Kette, §5) ohne Detektor, 50 Prozesse / `-test.count=50`:

| Id | Mutation | Repo | Probe |
|---|---|---|---|
| Z1 | `Errors` und `Dropped[fehlercode]` unter einer Sperre, die übrigen Gründe unter einer zweiten | **0 / 0 / grün** | rot 50 von 50 / rot |
| Z2 | `Errors[E2001]` und `Dropped` unter einer Sperre, die übrigen 32 Codes unter einer zweiten | **0 / 0 / grün** | rot 50 von 50 / rot |
| Z3 | `Errors` und `Dropped[schluessel, wert]` unter einer Sperre, die übrigen Gründe (auch `fehlercode`) unter einer zweiten | rot 50 / 50 / rot | rot 50 von 50 / rot |
| Z4 | `Dropped[wert]` allein zuerst, der Rest danach | rot 50 / 50 / rot | rot 50 von 50 / rot |
| Z5 | `verwerfen` zählt die Verwürfe eines `Log` in umgekehrter Reihenfolge, mit Freigabe dazwischen (Schreibseite) | rot 50 / 50 / rot | rot 50 von 50 / rot |
| Z6 | die 32 übrigen Codes zuerst, danach `Errors[E2001]` und `Dropped` | **0 / 0 / grün** | rot 50 von 50 / rot |
| Z7 | `Errors` und `Dropped[schluessel]` unter einer Sperre, die übrigen Gründe (auch `wert`) unter einer zweiten | rot 50 / 50 / rot | rot 50 von 50 / rot |
| Z8 | `Dropped` zuerst, dann `Errors` Code für Code in umgekehrter Tabellenreihenfolge, je Code eine Sperre | rot 50 / 50 / rot | rot 50 von 50 / rot |
| Z9 | `Dropped` nach der Freigabe der Sperre kopiert (Lesen ohne Sperre) | rot 50 / 50 / Absturz (`concurrent map`, 50 von 50 mit Absturz) | Absturz 50 von 50 |
| Z11 | `verwerfen` unter einer eigenen zweiten Sperre statt `mu` | rot 50 / 50 / Absturz; Detektor `DATA RACE` 50 von 50, Absturz 50 | Absturz 50 von 50 |
| Z10 | Kontrolle, äquivalent: `Durations` unter eigener Sperre kopiert (Wert nie verändert) | 0 / 0 / grün | 0 von 50 / grün |
| Z12 | Kontrolle, äquivalent: `verwerfen` sammelt lokal, wendet alle Verwürfe unter einer Sperre an | 0 / 0 / grün | 0 von 50 / grün |
| - | Gegenprobe, Stand `bb4e065` | 0 / 0 / grün | 0 von 50 / grün |

Zahl der Prozesse und Messender: je Zelle 50 Prozesse mit `-test.count=1` (Test-Binary einmal gebaut: *Detektor* `go test -race -c`, *ohne* `go test -c`) und ein Prozess mit `-test.count=50`, Messender der Verifier. **Last (Load Average, 1 Minute, zu Beginn der Zeile im Container):** Gegenprobe 4,1; Z1 2,9; Z2 2,5; Z3 2,3; Z4 2,3; Z5 2,3; Z6 2,3; Z7 1,9; Z8 2,1; Z9 2,3; Z10 2,3; Z11 1,9; Z12 1,9. Das liegt unter der Last, bei der der Implementer maß (4,7 bis 9,2); darüber habe ich den Fänger nicht gemessen. Die roten Zellen (50 von 50) sind damit bei geringer Last belegt; ob der Fänger bei hoher Last stabil bleibt, ist offen (V-14).

**Urteil zu den Teilen der Menge:**

- Die sieben Fenster des Implementers und meine Z3, Z4, Z5, Z7, Z8 fängt der Fänger in jedem Prozess, ohne dass der Detektor sie meldet (`DATA RACE` 0): bestätigt, der Wirkungs-Fänger trägt sie. Das Fehlen der Sperre (Z9, Z11) fängt die Laufzeit oder der Detektor.
- **G / Z10 sind äquivalent** (Negativ N-1, Durations und Spans): bestätigt. In diesem Teil ändert kein Aufruf `Durations` oder `Spans` (`StartSpan` gibt `keinSpan{}` zurück, `Snapshot` legt `Spans` leer an, `New` legt `Counts` mit Nullen an und kein Pfad schreibt sie); ein Fenster um sie liefert denselben Stand. Z10 und Z12 bleiben in beiden Fängern grün, wie es für äquivalente Mutationen sein muss.
- **Z1, Z2, Z6 sind es nicht.** Z2 und Z6 liegen am Rand des Negativs ("zwei der 32 Codes neben E2001"): ein Fenster zwischen `E2001` und einem Code, den der Test nie zählt, ist für diesen Test unbeobachtbar, für den Code aber nicht - im Betrieb zählt jeder der 33 Codes. Die Begründung "kein Test zählt sie" ist eine Eigenschaft des Tests, nicht des Codes, und das Negativ nennt für diesen Teil keine Folge im Betrieb (die genannte Folge "keine, solange `StartSpan` ohne Wirkung ist" gilt für `Durations` und `Spans`). Z1 liegt gar nicht im Negativ: `fehlercode` (gezählt über `CountError(0)`) und `schluessel`/`wert` (gezählt über `Log`) ändert der Test beide, und das Orakel setzt sie nicht zueinander in Beziehung. Eine Mutation, die `Dropped` in zwei Abschnitte teilt, wird nur gefangen, wenn der Schnitt zwischen `E2001` und `fehlercode` oder zwischen `schluessel` und `wert` liegt; `fehlercode` gegen `schluessel` und `E2001` gegen jeden anderen Code bleiben offen. Das ist **V-9**.

## 3. V-2: Position x Zustand - gegen die Klasse aufgelöst

**Menge:** Wortlaut "tief (Maps, `Counts`, `Spans` und deren `Attrs`), eine nil-Sammlung bleibt nil, eine leere nicht-nil" (`SPEC-230` Nr. 9): jede Position (6 in `Durations`, 6 in `Spans`) x jeder Zustand (nil, leer, 1, 2, 12 Elemente) je Sammlung = 30 + 30 Paare, dazu die 81 Stände der obersten Ebene.

**Zählprüfung gegen den Bau:** gelesen und nachgerechnet. `eintragLaenge(i, p, z)` gibt an Position `p` den Zustand `z` und an jeder anderen Position `laengen[(z+1+i%4)%5]` (einen der anderen vier Zustände, nie `z`); der Test zählt die Paare am gebauten Stand (`countsGesehen`, `attrsGesehen` aus `s.Durations[...].Counts` und `s.Spans[i].Attrs`, nicht aus Schleifenzählern) und verlangt je 30. Das Produkt ist vollständig und die Prüfung hängt am Bau.

**Eigene Mutationen** (Positionen 0, 2, 3, 4, 5 und alle Zustände, nicht die Elemente des Befunds), `Kopie()` auf Kopien; *Repo* = bestehende Suite aller drei Pakete, *Probe* = mein `TestProbeKopieKlasse` (Saat ab 300100, 486 Stände mit unabhängigen Zuständen je Position):

| Id | Mutation | Repo | Probe |
|---|---|---|---|
| Q1 | `Counts` von `span.3` geteilt, nur bei 12 Elementen | rot (`TestSnapshotKopie`) | grün |
| Q2 | `Attrs` von Position 5 geteilt, nur bei 1 Element | rot | rot |
| Q3 | `Counts` von `span.0` geteilt, nur bei 2 Elementen | rot | grün |
| Q4 | `Attrs` von Position 3: leer wird nil | rot | rot |
| Q5 | `Counts` von `span.2`: nil wird leer | rot | grün |
| Q6 | `Counts` von `span.5`: leer wird nil | rot | grün |
| Q7 | `Attrs` von Position 4 bei 12 Elementen auf 11 gekürzt | rot | grün |
| Q8 | `Attrs` von Position 1: nil wird leer | rot | rot |
| Q10 | `Attrs` von Position 0 geteilt, nur bei 12 Elementen | rot | grün |
| Q11 | `Counts` von `span.5` geteilt, nur bei 1 Element | rot | grün |
| Q12 | `Attrs` von Position 2 geteilt, nur bei 2 Elementen | rot | rot |
| Q9 | `Counts` von `span.4` geteilt, nur bei 1 Element **und** nil-`Dropped` (Kopplung zweier Ebenen) | grün | grün |

Elf von zwölf rot im Repo. Mein Probe ist an den Positionen schwächer als die Produkt-Prüfung des Repos (sechs der Mutationen fängt er nicht): das ist kein Befund, die Prüfung des Repos trägt die Klasse. **Q9** koppelt den Zustand einer Position an den Zustand einer anderen Sammlung; keine natürliche Implementierung tut das, und kein Fänger trifft sie: INFO V-11. V-2 ist **geschlossen**.

## 4. V-3: Fake - gegen die Klasse aufgelöst

**Mengen (Wortlaut `SPEC-230` Nr. 10: "jedes Argument unverändert und ungeprüft", "zählt jede Operation", "für gleichzeitigen Aufruf sicher"):** (a) Lesen: `Anzahl` (6 Operationen als Argument) und `Aufrufe()`; (b) Argumente: `Log` (Stufe, Text, ctx, Attributliste), `StartSpan` (Name, ctx, Attributliste), `Set` (Attributliste), `CountError` (Code); die Attributliste ist selbst die Menge Länge x Attr (Key x Art x Text x Ganz x Gleit).

**Probe-Paket (Saat 300400):** `TestProbeFakeArgumenteKlasse` hält 6000 zufällige Aufrufe über die vier Operationen gegen ein Orakel (die übergebenen Werte): Stufen `MinInt` bis `MaxInt` und zufällige 63-Bit-Werte, 7 Texte mit Zeilenumbruch, Rohbyte `0x80` und `0x00`, Namen leer, ein Leerzeichen, fremd und die fünf aus `Spans()`, Attributlisten nil, leer und 1 bis 12 Elemente aus Key (13 Werte, darunter fremde, leere und Rohbyte), Art (8 Werte, darunter 0, -1, 3, 99, `MaxInt32`), Text (12 Werte, darunter `A B` und Rohbyte), Ganz (10 Werte), Gleit (9 Werte, darunter NaN, -0, +-Inf); Attr verglichen über `math.Float64bits`. Besteht am Stand `bb4e065`.

**Eigene Mutationen am Fake** (Kopien; *Repo* = bestehende Suite, *Probe* = mein Paket `supporttest`):

| Id | Mutation | Repo | Probe |
|---|---|---|---|
| F1 | `Log`: Stufe `MaxInt` wird 0 | rot (`TestFakeArgumenteUngeprueft`) | rot |
| F2 | `StartSpan`: Name " " wird "" | rot | rot |
| F3 | `StartSpan`: Attribute der Art 0 entfallen | rot | rot |
| F4 | `Set`: Attribute mit Gleit +Inf entfallen | rot | rot |
| F5 | aufzeichnen: Attribut mit NaN entfällt | **grün** | rot |
| F6 | aufzeichnen: `mode`/`recorder.mode` mit Großbuchstaben im Text entfällt (gültiger Schlüssel, Art stimmt, Wertform verletzt) | **grün** | rot |
| F7 | `CountError`: `MaxInt-1` wird `MaxInt` | rot | grün (Probe zieht den Wert selten) |
| F8 | `Snapshot` zählt nicht, wenn ein Stand gesetzt ist | rot (`TestFakeSnapshotGleichzeitig`) | grün |
| F9 | `Log`: nil-`ctx` wird `Background` (Identität des ctx) | **grün** | rot (`TestProbeFakeModell`) |
| F10 | zweites `End` nicht aufgezeichnet | rot (`TestFakeAufzeichnung`) | rot |
| F11 | `Set` nach `End` nicht aufgezeichnet | rot | rot |
| F12 | Attributliste auf 2 Elemente gekürzt | **grün** | rot |

(a) Lesen, Nebenläufigkeit: Messender der Verifier, je Zelle 30 Prozesse mit `-test.count=1` und ein Prozess `-test.count=30`, Tests `TestFakeLesenGleichzeitig`, `TestFakeGleichzeitig`, `TestFakeSnapshotGleichzeitig`; Last (Load Average, 1 Minute) 3,0 bis 5,9. *Wirkung* = ohne Detektor:

| Id | Mutation | Detektor, 30 Prozesse | Wirkung, 30 Prozesse | `-count=30` |
|---|---|---|---|---|
| - | Gegenprobe | 0 von 30 | 0 von 30 | grün |
| F13 | `Anzahl(OpSnapshot)` ohne Sperre | **30 von 30 rot, `DATA RACE` 30** | 0 von 30 | grün |
| F14 | `Anzahl(OpEnd)` ohne Sperre | **30 von 30 rot, `DATA RACE` 30** | 0 von 30 | grün |
| F18 | `Anzahl` ganz ohne Sperre | **30 von 30 rot, `DATA RACE` 30** | 0 von 30 | grün |
| F19 | `Aufrufe()` ganz ohne Sperre | **30 von 30 rot, `DATA RACE` 30** | 8 von 30 | rot |
| F16 | `Aufrufe()` liest die Slice-Kopfzeile unter der Sperre, klont nach der Freigabe | 0 von 30 | 0 von 30 | grün (äquivalent: Elemente unterhalb der Länge ändern sich nie) |

Das Fehlen der Sperre in `Anzahl` und `Aufrufe()` meldet der Detektor in 30 von 30 Prozessen; die Wirkung ohne Detektor ist, wie der Plan sagt, ein Messwert und kein Teil der Zusage. **V-3 a ist geschlossen.**

**V-3 b ist nur zum Teil geschlossen: V-10 (LOW).** Die Menge "jedes Argument" ist in `TestFakeArgumenteUngeprueft` nicht ausgeschöpft: vier Elemente der Klasse fängt kein Test des Repos (F5, F6, F9, F12): (1) Attribut mit NaN; (2) gültiger Schlüssel und gültige Art mit verletzter Wertform (der Kern von "ungeprüft": der Fake darf die Regeln des Adapters nicht anwenden); (3) Attributliste länger als 2; (4) `Log` mit nil-`ctx` (der Kopf von `Aufruf.Ctx()` sagt "auch nil", der nil-Fall hat nur `StartSpan`, `AGENTS.md` §3.11). Der Code des Fakes ist konform; fehlend ist je Element eine Zeile im Test. Der Fix ist eine Erweiterung der Listen (`Str("mode", "A B")`, ein Eintrag mit NaN, Listen mit 3 und mit 12 Elementen, `Log` mit nil-`ctx`).

F17 (ID-Zähler des Fakes ohne Sperre) ist in der bestehenden Suite ohne Detektor grün; das deckt das akzeptierte Negativ R-7 der ersten Runde (Eindeutigkeit der Fake-IDs nicht zugesagt). Ob `TestFakeGleichzeitig` den ungesperrten Zugriff unter dem Detektor meldet, habe ich nicht gemessen.

## 5. Fix-Vorschlag für V-9

Mein Fänger `TestProbeStand2` (Scratchpad, nicht committet) ist die Vorlage: 8 Schreiber zählen je Runde nacheinander `CountError(E2001)`, `CountError(E2101)`, `CountError(0)` und ein `Log` mit zwei Verwürfen (`schluessel`, `wert`); 8 Leser prüfen je Stand die Kette `E2101 <= E2001 <= E2101 + 8`, `fehlercode <= E2101 <= fehlercode + 8`, `schluessel <= fehlercode <= schluessel + 8` und `schluessel == wert`. Er fängt Z1 bis Z8 in 50 von 50 Prozessen, Z9 und Z11 als Absturz, und bleibt bei Gegenprobe, Z10 und Z12 in 0 von 50 und in `-count=50` grün. Zur Klasse genügt das nicht ganz: die 33 Codes sind ein Raster von 528 Paaren. Eine Allaussage ("jeder Code") schreibt der Test nach `AGENTS.md` §3.15 und §3.17 als Schleife: Schreiber zählen alle 33 Codes in fester Reihenfolge, das Orakel prüft die Kette über alle Nachbarn. Alternative für die Codes, die kein Test ändert: ein Negativ des Architect mit Grund und Folge im Betrieb (die Diagnose weicht um einen Zähler ab, das Ergebnis des Aufrufers bleibt unberührt).

## 6. V-4 und V-5 bis V-8

- **V-4 geschlossen.** Jeder Rang-Zeiger der Pakete `ports/support` und `adapters/driven/telemetry` (27 Vorkommen, `grep -rnoE "\bP[0-9]{2}\b"`) liegt in P01 bis P17; `P18` und `P19` kommen in beiden Paketen nicht mehr vor. Gelesen, ob die Fundstelle auf die Zeile zeigt, deren Fänger der Test ist: `TestSnapshotZeitpunkt` P15, `TestFakeArgumenteUngeprueft` und `TestFakeSpanJeOperation` P11, `TestFakeLesenGleichzeitig` P16, `TestLogOhneMinLevel` P17. Die Zeiger ab `P18` in `outboundtest/metadata_codec_aufloesen.go` gehören zu einem anderen Slice und sind kein Gegenstand.
- **V-5 bis V-8** (INFO der ersten Runde) bleiben beim Planner und sind unverändert: V-5 (grüne Mutationen R6, T1, X8) optional, V-6 (Änderung des Re-Review-Reports durch `ed3142b`) Kenntnis, V-7 (Nullwert des Adapters) Kenntnis des Architect, V-8 (`filtern` mit `op == ""` ist die Menge `Log`; ein Span mit leerem Namen darf nie bis dorthin kommen) Adresse `slice-telemetry-spans`. Keiner von ihnen steht in §6 des Plans; sie gehen mit der Closure an den Planner.

## 7. Die Adresse an `slice-telemetry-spans` (`AGENTS.md` §3.13)

Der Plan weist die Fenster um `Durations` und `Spans` (Negativ N-1) dem Slice `slice-telemetry-spans` zu und setzt als Adresse den Planner (§6, Ausgang *offen*; der Zweig schreibt nichts in `open/`). Gelesen: `slice-telemetry-spans` liegt in `docs/plan/planning/open/`, ist also eine gültige Adresse (kein Ausschluss in §1 trifft die Sendung, er liegt nicht in `done/`). Er trägt die Sendung **zum Teil**: DoD und Zeile S16 (`TestSnapshotEinStand`: "`Snapshot` ist ein Stand (Verteilung und Ring unter einer Sperre)", Orakel `Count` gleich Summe von `Counts`) decken die Lücke zwischen Verteilung und Ring; die Lücke zwischen `Errors`/`Dropped` und den beiden steht nicht darin, und der Geber `slice-telemetry-port` ist dort nicht genannt. Das ist keine Verletzung von §3.13 (die Zeile in §6 trägt den Ausgang *offen* und die Adresse), aber die Übernahme steht aus: **INFO V-13**, Planner: die Sendung samt Geber in §1 oder DoD von `slice-telemetry-spans` eintragen, sobald der Slice startet, und das Orakel auf die Kette über alle vier Sammlungen ausdehnen.

## 8. Befunde

| ID | Schwere | Befund | Zuständig |
|---|---|---|---|
| V-9 | MEDIUM | **Klasse "Stand zu einem Zeitpunkt" in `Snapshot` nur zum Teil gefangen (Rest von V-1).** Z1 (Lücke zwischen `fehlercode` und `schluessel`/`wert`), Z2 und Z6 (Lücke zwischen `E2001` und den 32 anderen Codes) bleiben in `TestSnapshotZeitpunkt` in jeder Spalte grün (0 von 50 mit und ohne Detektor, `-count=50` grün); eine zweite Kette im Orakel fängt sie in 50 von 50 (§2, §5). Z1 liegt nicht im benannten Negativ N-1 (beide Zähler ändert der Test); Z2 und Z6 liegen im Negativ nur als Eigenschaft des Tests ("kein Test zählt sie"), und das Negativ nennt für sie keine Folge im Betrieb (`AGENTS.md` §3.15: ein akzeptiertes Negativ nennt Grund und Folge). Eine Zusage ohne Fänger ist ein MEDIUM. Fix: Schreiber zählen mindestens einen zweiten bekannten Code, besser alle 33 in einer Schleife, das Orakel prüft die Kette über alle Nachbarn (Vorlage §5); oder ein Negativ des Architect mit Folge. | Implementer (Fänger) oder Architect (benanntes Negativ in `SPEC-232` Ungeprüft) |
| V-10 | LOW | **"jedes Argument ungeprüft" (`SPEC-230` Nr. 10) hat vier Elemente der Menge ohne Fänger:** Attribut mit NaN (F5), gültiger Schlüssel und Art mit verletzter Wertform (F6), Attributliste länger als 2 (F12), `Log` mit nil-`ctx` (F9; der Kopf von `Aufruf.Ctx()` sagt "auch nil", `AGENTS.md` §3.11). Der Fake ist konform (mein Probe über 6000 Aufrufe grün); es fehlen Zeilen in den Listen von `TestFakeArgumenteUngeprueft`. | Implementer |
| V-11 | INFO | Q9 (Zustand einer Position an den einer anderen Sammlung gekoppelt) bleibt in Repo und Probe grün; keine natürliche Implementierung koppelt so, kein Befund. | - |
| V-12 | INFO | Zahlen im Plan: die Aufzählung in §6 ("Messung 2550, §4") und die Messung in §4 (2551) weichen um 1 ab; das Gate meldet 2551 und die Summe stimmt mit 2551 (§1). Nachzug im Plan. | Implementer |
| V-13 | INFO | Die Adresse für N-1 ist gültig (Nehmer in `open/`), die Übernahme steht aus: die DoD von `slice-telemetry-spans` trägt die Lücke Verteilung-Ring, nicht die Lücke zu `Errors`/`Dropped`, und nennt den Geber nicht (§7). | Planner |
| V-14 | INFO | Raten von `TestSnapshotZeitpunkt` bei Load Average 1,9 bis 4,1 gemessen (50 von 50 in allen roten Zellen); der Implementer maß bei 4,7 bis 9,2. Für höhere Last habe ich keine Zahl; die Zusage trägt der Wirkungs-Fänger, nicht der Detektor (der Plan sagt es so). | Planner (Kenntnis) |

## 9. Ungeprüft und Gegengelesen

**Ungeprüft** (Sperre überlaufen, Zeitgrenze von 90 Minuten):

- Rate von `TestSnapshotZeitpunkt` und der Fänger der Fake-Lesezugriffe bei höherer Last als 5,9.
- F17 (ID-Zähler des Fakes ohne Sperre) unter dem Detektor.
- `make docs-check` auf diesem Report; `make gegenproben` und Stempel (Integrationsschritt, laut Auftrag).
- Mutationen an `Log`/`CountError` in Kombination mit `Clock.Now()` (`SPEC-232` Nr. 5 f) gehören zu L und S.

**Gegengelesen:** dieser Report ist auf Rohbytes gegengelesen (`LC_ALL=C grep -P` auf Backslash-u-Folgen, BOM und NBSP, vor dem Commit). Die Probe-Dateien (`zz_probe_test.go`, `zz_snap2_test.go`, `zz_probe2_test.go`) und die Mutationen liegen im Scratchpad und sind nicht Teil des Commits.

## 10. Closure-Vorbedingungen (nur genannt)

- DoD offen (Planner): Closure-Notiz mit Steering-Loop-Lerneintrag (§7 des Plans), Zeile *Ende* im Messpunkt (`SPEC-128` Nr. 9; Gate meldet `messung=2551`, Auslastung 0,85 der Marke und damit im Bereich 0,60 bis 1,00), Beobachtungs-Register fortschreiben, Risiken aus §6 mit Ausgang, die drei Paarungen.
- Diff-Vorlage zu `AGENTS.md` §3.14 an den User (Planner, nicht dieser Zweig).
- V-5 bis V-8 sowie V-13 an den Planner; V-9, V-10 und V-12 an den Implementer.

## 11. Register-Kandidaten (genannt, nicht geschrieben)

- `BEO-REPO/sperrfenster-zwischen-sammlungen-ohne-faenger` (aus Runde 1): ein zweiter Beleg - die Fix-Runde deckt die Fenster zwischen Sammlungen und zwischen den Einträgen, die ein Test ändert, nicht aber zwischen Einträgen, die der Test nie zählt oder nicht zueinander in Beziehung setzt (V-9).
- `BEO-REPO/orakel-prueft-nur-die-zaehler-des-testaufbaus`: ein Wirkungs-Fänger mit Orakel über einen Teil der Zähler lässt Fenster zwischen den übrigen offen (V-9). Zähler 1x.
- `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (aus Runde 1): ein weiterer Beleg "jedes Argument ungeprüft" mit vier fehlenden Elementen der Menge (V-10).
