# Review-Report: slice-telemetry-port, Fix-Runde — 2026-10-30

**Review-Art:** Code, Re-Review der Fix-Runde — jeder Fix gegen die Klasse seiner Zusage (`AGENTS.md` §3.17), gegen Plan, `SPEC-230` bis `SPEC-232` (die in Teil `P` genannten Nummern), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) und die Hard Rules (`AGENTS.md` §3.7 bis §3.18).

**Gegenstand:** `slice-telemetry-port` (Teil `P` der Kette `P` → `L` → `S`), Zweig `slice-telemetry-port`, Diff `3c3a72a..84a8d7e` (Architect-Commit `a557519`, Implementer-Commit `84a8d7e`). Erster Review: [2026-10-30-slice-telemetry-port-review.md](2026-10-30-slice-telemetry-port-review.md) (F-1 bis F-11).

**Skill:** `.harness/skills/reviewer.md`
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-30

**Eingangs-Kontext:** Plan `slice-telemetry-port` (§3 Tabelle `P01` bis `P17` samt Fix-Runde, §6), Diff `3c3a72a..HEAD`, `SPEC-230` bis `SPEC-232`, `AGENTS.md`. Kein Bericht des Implementers; die Raten und Mutationen des Plans sind nachgemessen, nicht übernommen.

**Arbeitsweise.** Mutationen auf Kopien in einem Wegwerfverzeichnis (`rev2-telport/`), in Containern (`odata-recorder:deps-mut`, `odata-recorder:race`) unter `flock /tmp/odata-recorder-heavy.lock`, `--memory=4g --memory-swap=4g --cpus=4`; jede Mutation vor dem Lauf auf Änderung geprüft (`cmp`). Im Repo wurde nichts geändert außer diesem Report. Die Schlange der Sperre war lang (Wartezeit bis über 40 Minuten); was darum nicht lief, steht unter *Ungelaufen*.

**Gefahren:** 51 Mutationen im Lauf 1 (Regeln, Port, Kopie, Fake, Adapter; alle auf Elemente, die weder der erste Review noch der Implementer nannten, soweit die Klasse es zulässt), 17 Zellen mit Raten (drei davon ohne Mutation) (je 30 Prozesse Detektor `-test.count=1`, 30 Prozesse ohne Detektor, ein Prozess `-test.count=30`), Rohbytes-Gegenlesen.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| R-1 | MEDIUM | Der Fix zu F-5 deckt von der Menge „Sammlungen von `Snapshot.Kopie()`“ nur die vier der obersten Ebene. `SPEC-230` Nr. 9 sagt „tief (Maps, `Counts`, `Spans` und deren `Attrs`), eine nil-Sammlung bleibt nil, eine leere nicht-nil“; fünf Mutationen bleiben grün: `Counts` nil zu leer, `Counts` leer zu nil, `Attrs` nil zu leer, `Attrs` leer zu nil, und nur der erste `SpanRecord` bekommt eine eigene `Attrs`-Kopie (`for i := range min(1, len(k.Spans))`). Der Fänger führt ein `SpanRecord` mit gefüllten `Attrs` und ein `Verteilung` mit gefüllten `Counts`; die Fälle nil und leer prüft er nur oben. Ein Dienst-Test, der einen Stand mit `Verteilung{Counts: nil}` setzt und mit `reflect.DeepEqual` liest, bekäme einen anderen Stand als gesetzt; bei zwei Spans teilt die Kopie des zweiten `Attrs` mit dem Original. | `AGENTS.md` §3.17 (die Menge aus dem Wortlaut der Zusage, nicht aus dem Befund), §3.15 (ein Fall je Element einer Menge); `SPEC-230` Nr. 9 | `src/hexagon/ports/support/telemetry_test.go` · `if z := leer.Kopie(); z.Errors == nil` | ja — `go test ./src/hexagon/ports/support/...` mit einer der fünf Mutationen in `Snapshot.Kopie` (z. B. `v.Counts = append([]uint64{}, v.Counts...)`) | Fänger führt eine Ebene und ein Element der Menge |
| R-2 | LOW | `FakeTelemetry.Snapshot` ohne Sperre bleibt unter jedem Lauf grün (Detektor 0 von 30, ohne Detektor 0 von 30, `-test.count=30` grün), weil `TestFakeGleichzeitig` nur `aufzeichnen`-Aufrufe und `Snapshot` mischt und kein Schreiber `stand` oder `gesetzt` anfasst (`SetzeStand` läuft nie gleichzeitig). Der Kopf nennt den Fake „für gleichzeitigen Aufruf sicher“. | `AGENTS.md` §3.10 (Zusage über Nebenläufigkeit), §3.11; `SPEC-230` Nr. 10 | `src/hexagon/ports/support/supporttest/telemetrie_test.go` · `func TestFakeGleichzeitig` | ja — Probe `SetzeStand` gegen `Snapshot` (8 + 8 Goroutinen, je 500 Aufrufe) unter `go test -race`: grün auf dem Stand, `DATA RACE` mit `Snapshot` ohne `f.mu.Lock()`; die Suite mit dieser Mutation ist grün | Zusage „sicher“ weiter als der Fänger (Operationen `Snapshot` und `SetzeStand` nie gleichzeitig) |
| R-3 | LOW | Die Mutation „`nurZeichen` prüft das Byte an Index 127 nicht“ bleibt grün: `TestWertformZeichen` führt Grundwerte von höchstens 71 Bytes (Blob); Kennung (64) und Name (128) werden nur an den ersten 7 bzw. 6 Stellen und am Ende des Grundwerts mit jedem Byte belegt, die Stellen 71 bis 127 eines Namens nie. Die Funktion ist für alle Formen gemeinsam, ein Fehler dort fällt über den Blob an den Stellen bis 70 auf (Index 9, 31 und 63 rot); der Rest bleibt eine Lücke der Menge „jedes Byte an jeder Stelle“. | `AGENTS.md` §3.17, §3.15 | `src/adapters/driven/telemetry/regeln_test.go` · `"name": "Ab.1_x"` | ja — `go test` über das Paket `telemetry` mit `if i != 127 && !ok(s[i])` in `nurZeichen` | Fänger führt nicht alle Stellen der Menge |
| R-4 | LOW | Die Mutation „Dauer ab 0“ als `a.Gleit == 0 \|\| a.Gleit >= 1e-300` bleibt grün: der kleinste positive Wert der Fänger ist `1e-6` (und 1 ns als 1e-6 ms); der Rand „größer 0, kleiner als jeder Wert der Fixture“ (kleinste Gleitkommazahl, subnormale Werte) fehlt. | `AGENTS.md` §3.15 (beide Seiten jeder Grenze); `SPEC-230` Nr. 3 (Dauer) | `src/adapters/driven/telemetry/regeln_test.go` · `[]float64{math.Copysign(0, -1), -1e-6, 0, 1e-6` | ja — `go test` über das Paket `telemetry` mit der Mutation | Grenze 0, positive Seite nur mit einem Vertreter |
| R-5 | LOW | Der Kopf von `TestAdapterImporte` und die Zeile P14 sagen „die Menge der Importe jeder Datei des Pakets“; der Test überspringt jede Datei `_test.go` (`strings.HasSuffix(d, "_test.go")`). Ein Import von `ports/outbound` oder `supporttest` in `export_test.go` bleibt unbeanstandet (gewollt für Testhilfen, aber nicht gesagt). | `AGENTS.md` §3.11 (ein Kommentar sagt nur zu, was ein Test prüft) | `src/adapters/driven/telemetry/telemetry_test.go` · `Menge der Importe jeder Datei des Pakets` | ja — Lesung des Tests | Kopf-Zusage weiter als die Prüfung |
| R-6 | INFO | Die Raten des Plans für „Wirkung ohne Detektor“ (50 von 50) sind in drei Zellen nicht zu reproduzieren: gemessen (30 Prozesse) 29 von 30 für `CountError` unbekannter Zweig ohne Sperre, bekannter Zweig ohne Sperre und bekannter Zweig mit Fenster; der gate-relevante Detektor-Lauf ist in allen Zellen 30 von 30. Kein Handlungsbedarf, weil `make test-race` die Zusage trägt; die Zahl im Plan ist ein Messwert unter anderer Last. | `AGENTS.md` §3.10 (Raten mit Zahl) | `docs/plan/planning/done/slice-telemetry-port.md` · `Raten der Fix-Runde` | ja — Wiederholung der Raten | Rate unter anderer Last |
| R-7 | INFO | Die IDs des Fakes sind unter gleichzeitigem `StartSpan` nicht pünktlich eindeutig, wenn eine Mutation die Zählung in zwei Abschnitte teilt (`f.naechste` gelesen, entsperrt, wieder gesperrt, Zuweisung): grün in jedem Lauf (Detektor 0 von 30). `SPEC-230` Nr. 10 und `SPEC-232` Nr. 5 e sagen die Eindeutigkeit nur für den Adapter zu; der Kopf des Fakes sagt „zählt je Fake ab 1“. Kein Handlungsbedarf, solange kein Test des Fakes Eindeutigkeit voraussetzt. | `AGENTS.md` §3.11 | `src/hexagon/ports/support/supporttest/telemetrie.go` · `ID zählt je Fake ab 1` | ja — Mutation | Eindeutigkeit der Fake-IDs unter Nebenläufigkeit nicht zugesagt |

Offene Punkte, die der Auftrag ausnahm und die hier nicht doppelt stehen: F-10 (`TestKernSchreibtNichts`, Adresse Planner), Zeitbudget der Gegenproben ([ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)), Stempel.

## Auflösung der Fixes F-1 bis F-11 gegen die Klasse

| Finding | Menge (aus dem Wortlaut) | Eigene Mutationen an Elementen, die weder Reviewer noch Implementer nannten | Ergebnis |
|---|---|---|---|
| F-1 | jedes Byte außerhalb der Form je Wertform (6 Textformen × 256 Bytes × alle Stellen des Grundwerts) | Token: Ziffer als erstes Zeichen; Name: `.` und Ziffer als erstes Zeichen, Leerzeichen in der Mitte; Kennung: `~`, `/`; Großbuchstabe ab `@`; Kleinbuchstabe bis `{`; Hex: `/` und `A`; Methode: Ziffer; Blob-Präfix nur 6 Zeichen; `nurZeichen` prüft nur 10 Bytes, überspringt Index 9, 31, 63 | alle 16 rot; `nurZeichen` überspringt Index 127: grün (R-3) |
| F-2 | jeder Zweig von `CountError` unter Nebenläufigkeit | unbekannter Zweig ohne Sperre, unbekannter Zweig mit Fenster, bekannter Zweig ohne Sperre, bekannter Zweig mit Fenster (selbst formuliert) | rot, Raten unten |
| F-3 | jede Operation des Adapters unter der Sperre (`CountError` 2 Zweige, `Log`/`verwerfen`, `Snapshot`) | `Snapshot` ohne Sperre, `Snapshot` gibt vor `Kopie()` frei (neues Fenster), `verwerfen` ohne Sperre, `verwerfen` mit Fenster, unbekannter Zweig ohne Sperre | rot, Raten unten |
| F-4 | die sechs Operationen des Fakes | `End` mit dem Eltern-Span, `Set` mit dem Eltern-Span | rot (`TestFakeSpanJeOperation`, `TestFakeAufzeichnung`) |
| F-5 | alle Sammlungen von `Snapshot` und `Kopie`, alle Ebenen | `Spans` nil im Adapter; Fake: je eine der vier Sammlungen nil im ungesetzten Stand (4 Mutationen); gesetzt nur bei `Errors != nil`; `Spans` nie nil im gesetzten Stand; `Kopie`: `Counts` und `Attrs` nil zu leer und leer zu nil; Schleife nur über den ersten `SpanRecord` | Ebene 1 rot; Ebene 2 (`Counts`, `Attrs`) und der zweite Span: grün (R-1); „ohne ersten“ rot |
| F-6 | 4 Stufen × 4 `MinLevel` | Rückkehr nur bei Stufe `Debug`; nur bei `MinLevel` über `Error` | rot (`TestLogOhneMinLevel`) |
| F-7 | die ID jedes Spans | Zählerschritt 2 statt 1 | rot (`TestFakeSpanID`) |
| F-8 | alle Dateien des Pakets (ohne `_test.go`) | siehe *Zeitfenster Lauf 2* | siehe unten |
| F-9 | zwei Adapter ohne gemeinsamen Zustand | mit den Mutationen des Implementers belegt (gemeinsame Map); eigene Paketvariable für `durations` ist im Teil `P` nicht beobachtbar (kein Schreiber) | durch Teil `S` zu decken; kein Befund |
| F-11 | der Nullwert, jede Operation | `Anzahl` ohne obere Grenze (`>`) und mit falscher unterer Grenze (`op < 2`) | rot (`TestFakeNullwert`, `TestFakeAufzeichnung`) |
| `-0.0` | Vorzeichenbit der Dauer | Mutation `(a.Gleit != 0 \|\| !Signbit)` ist äquivalent (negative Werte fallen an `>= 0`); `+0.0` und `Millis(0)` angenommen, geprüft in `TestWertformGrenzenVonHand` | ohne Befund |
| `Verwuerfe()` | Reihenfolge der sieben Gründe | Vertauschen je eines Paars benachbarter Gründe (5 Mutationen: 0-1, 1-2, 2-3, 3-4, 5-6) | alle rot (`TestVerwuerfe`) |

## Raten der nebenläufigen Zusagen (selbst gemessen)

Messender: der Reviewer; Image `odata-recorder:race`; je Zelle 30 Prozesse mit `-test.count=1` (*Detektor*: `go test -race -c`; *Wirkung*: ohne Detektor; *Wiederholung*: ein Prozess ohne Detektor mit `-test.count=30`; DR = Prozesse mit `DATA RACE`), Host unter Last (Load Average 17 bis 35). Der Plan nennt 50 Prozesse; die Wartezeit der Schlange erlaubte 30.

| Test | Mutation | Detektor | Wirkung | Wiederholung |
|---|---|---|---|---|
| `TestCountErrorGleichzeitig` | keine | 0 von 30 | 0 von 30 | grün |
| `TestCountErrorGleichzeitig` | unbekannter Zweig ohne Sperre | 30 von 30, DR 30 | 29 von 30 | Absturz |
| `TestCountErrorGleichzeitig` | unbekannter Zweig mit Fenster | 30 von 30, DR 0 | 30 von 30 | 26 von 30 Wiederholungen rot |
| `TestCountErrorGleichzeitig` | bekannter Zweig ohne Sperre | 30 von 30, DR 30 | 29 von 30 | Absturz |
| `TestCountErrorGleichzeitig` | bekannter Zweig mit Fenster | 30 von 30, DR 0 | 29 von 30 | 30 von 30 rot |
| `TestAdapterGleichzeitig` | keine | 0 von 30 | 0 von 30 | grün |
| `TestAdapterGleichzeitig` | `Snapshot` ohne Sperre | 30 von 30, DR 0 | 30 von 30 | Absturz |
| `TestAdapterGleichzeitig` | `Snapshot` gibt vor `Kopie()` frei | 30 von 30, DR 0 | 30 von 30 | Absturz |
| `TestAdapterGleichzeitig` | `verwerfen` ohne Sperre | 30 von 30, DR 29 | 30 von 30 | Absturz |
| `TestAdapterGleichzeitig` | `verwerfen` mit Fenster | 30 von 30, DR 0 | 30 von 30 | 30 von 30 rot |
| `TestAdapterGleichzeitig` | unbekannter Zweig ohne Sperre | 30 von 30, DR 29 | 30 von 30 | Absturz |
| `TestFakeGleichzeitig` | keine | 0 von 30 | 0 von 30 | grün |
| `TestFakeGleichzeitig` | `zaehler` mit Fenster | 30 von 30, DR 0 | 30 von 30 | 30 von 30 rot |
| `TestFakeGleichzeitig` | `aufrufe` mit Fenster | 30 von 30, DR 0 | 30 von 30 | 30 von 30 rot |
| `TestFakeGleichzeitig` | `ID` nach der Freigabe gelesen | 30 von 30, DR 30 | 0 von 30 | grün |
| `TestFakeGleichzeitig` | `Snapshot` ohne Sperre | 0 von 30 | 0 von 30 | grün (R-2) |
| `TestFakeGleichzeitig` | `ID` mit Fenster (kein Datenrennen, Doppel-IDs) | 0 von 30 | 0 von 30 | grün (R-7) |

Die Zusagen von F-2 und F-3 sind in jedem Detektor-Prozess gefangen (30 von 30 je Zelle); die Fenster-Mutationen ohne `DATA RACE` fängt der Wirkungs-Fänger im Detektor-Lauf in jedem Prozess.

## Größenmarke

`make groessenmarke` im Zweig, Exit 0: `slice-telemetry-port.md: ok, Messung 2016 von Marke 3000`; Messpunkt `zeilen=17 nebenlaeufig=2 fakes=1 produktcode=580 schaetzung=2150 marke=3000 messung=2016`. Der erste Review maß 1759; die Fix-Runde fügt 257 Zeilen hinzu. Die Marke wurde nicht gesetzt.

## Ungelaufen

Bei der Abgabe dieses Reports lagen in der Schlange der Sperre, ohne Ergebnis: (1) der Teil des Probe-Pakets für den Adapter (Generatoren und eigenes Orakel, Saat ab 100000: Textformen je Byte und Stelle samt Mehrbyte-Zeichen, Dauer über Zufallsbitmuster, Blob, Status und Zahl, `Kopie` über Stände mit nil/leer/gefüllt auf jeder Ebene, `New` × Wiederholung gegen ein Modell je Adapter): sein erster Lauf scheiterte am Aufbau, weil das Schreibwerkzeug ein `\u`-Literal des BOM in das Zeichen umsetzte (`AGENTS.md` §3.18, illegal byte order mark); die Datei ist aus Bytewerten neu gebaut und erneut eingereiht, ohne Ergebnis bei Abgabe. Gelaufen und in R-2 genannt ist der Teil für den Fake (Modell-Orakel über die sechs Operationen und Stand nil/leer: grün auf dem Stand; `SetzeStand` gegen `Snapshot`: grün auf dem Stand, rot mit der Mutation unter dem Detektor). Die Aussagen zu den Klassen F-1 und F-5 stützen sich allein auf die Mutationen; (2) sieben Mutationen zu F-8 (Import je eines weiteren Pakets in einer neuen Datei des Adapters: `adapters/driven/clock`, `services/record`, `ports/outbound`, `adapters/driven/sqlite`, `ports/inbound`, `supporttest`, `net/http` als Gegenprobe); (3) die zwölf Mutationen des ersten Reviews zu F-1 gegen den Fix-Stand (nicht nachgefahren; die Menge der Fix-Runde ist größer, ob sie jede der zwölf fängt, ist nicht gemessen). Wer den Lauf nachholt: die Skripte liegen im Scratchpad `rev2-telport/` (`probe2.sh`, `mb2.sh`, `mb3.sh`).

Nicht geprüft in diesem Lauf: `make gates` (Aufgabe des Implementers und des Verifiers), die Closure-Notiz.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/telemetry/regeln.go` (`Signbit`) | geprüft, ohne Befund: `-0.0` abgewiesen, `+0.0` angenommen, wie `SPEC-230` Nr. 3 |
| `src/adapters/driven/telemetry/telemetry.go` | geprüft, ohne Befund am Code: unverändert im Fix-Diff; `Spans: []support.SpanRecord{}` und `Kopie()` stehen im Wortlaut von `SPEC-230` Nr. 9; alle Zugriffe unter `mu` |
| `src/hexagon/ports/support/supporttest/telemetrie.go` | geprüft: der Nullwert nutzbar (Zählerfeld als Array, `gesetzt`), `Anzahl` mit Grenzprüfung; Befunde R-2, R-7 liegen im Fänger |
| `spec/spezifikation.md` (`SPEC-230` Nr. 3, 4, 9, 10; `SPEC-232` Nr. 1) | geprüft, ohne Befund: die Änderungen sind die Entscheidungen des Architect; Nr. 9 (Stand nie nil) und Nr. 10 (Fake: gesetzter Stand „nil oder leer, wie gesetzt“) widersprechen sich nicht, sie gelten für verschiedene Erzeuger |
| Rohbytes (`AGENTS.md` §3.18) | geprüft, ohne Befund: in `regeln_test.go`, `telemetry_test.go`, `regeln.go`, `telemetry.go`, `telemetrie.go`, `telemetrie_test.go`, `telemetry_test.go` (ports/support) und im Plan keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP (`LC_ALL=C grep -P`) |
| Suppression (`AGENTS.md` §3.2) | geprüft, ohne Befund: keine `//nolint` im Fix-Diff |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft; ohne Konjunktiv über verworfene Alternativen; eine Kopf-Zusage weiter als der Test: R-5 |
| Plan folgt dem Code (`AGENTS.md` §3.9) | geprüft, ohne Befund: P03, P04, P08 bis P11, P14, P15, P17, die Raster-Zeilen `New` und `Lesen`, §6 (Hilfen, `-0.0`, nil gegen leer, F-10 mit Ausgang *offen*) und die Klassen-Aufzählung (28672 Werte = 256 × 112) nennen, was der Code liefert; Zählprüfung und Ort der Zellen stimmen |
| Kern enthält nichts von `L` oder `S` | geprüft, ohne Befund: `Log` schreibt keine Zeile, `StartSpan` bleibt ohne Wirkung, `Handler`, Ring und Verteilung fehlen, `go.mod` und `go.sum` im Diff unverändert |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 4 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fänger führt eine Ebene und ein Element der Menge · Zusage „sicher“ weiter als der Fänger · Fänger führt nicht alle Stellen der Menge · Grenze 0, positive Seite nur mit einem Vertreter · Kopf-Zusage weiter als die Prüfung · Rate unter anderer Last · Eindeutigkeit der Fake-IDs nicht zugesagt

## Verdikt

**Merge-blockierend:** ja, solange R-1 offen ist (MEDIUM). Die Fixes zu F-1 bis F-4, F-6, F-7, F-9 und F-11 decken ihre Klassen: die Mutationen an Elementen, die der Befund nicht nannte, werden rot. F-5 ist auf der obersten Ebene gedeckt, in den Ebenen `Counts` und `Attrs` und beim zweiten `SpanRecord` nicht (R-1). F-8 ist wegen des offenen Laufs unter *Ungelaufen* nur gelesen, nicht gemessen. R-2 bis R-5 sind Lücken von LOW; R-6 und R-7 ohne Handlungsbedarf.

**Übergabe:** R-1 bis R-5 an den Implementer; R-1 mit der Menge „6 Sammlungen × (nil, leer, gefüllt) und jeder `SpanRecord`“ als Klasse (`AGENTS.md` §3.17). Die Finding-Klassen gehen in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ist ein Lauf-Beleg und ersetzt keine Verifikation; DoD- und Spec-Konformität prüft der Verifier.
