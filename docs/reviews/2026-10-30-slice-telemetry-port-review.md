# Review-Report: slice-telemetry-port — 2026-10-30

**Review-Art:** Code — gegen Plan, `SPEC-230` bis `SPEC-232` (nur die in Teil `P` genannten Nummern), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) und die Hard Rules (`AGENTS.md` §3.7, §3.9 bis §3.18).

**Gegenstand:** `slice-telemetry-port` (Teil `P` der Kette `P` → `L` → `S`), Zweig `slice-telemetry-port`, Diff `a0396b0..3c3a72a` (Handoff-Commit `3c3a72a`; die Planner- und Architect-Commits dazwischen sind Eingang, nicht Gegenstand).

**Skill:** `.harness/skills/reviewer.md` @ 4989da8
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-30

**Eingangs-Kontext** (die Verträge, gegen die geprüft wurde):

- `slice-telemetry-port` (§3 Tabelle `P01` bis `P17`, Mutationen und Raten des Implementers, Raster und Zusatztabelle in §6, Hilfen ohne Nennung und Randform `-0.0` mit Ausgang *offen*)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Go, Standardbibliothek)
- `LH-QA-05`, `LH-QA-04`, `LH-FA-01`
- `SPEC-230` Nr. 1 bis 4 und 8 bis 10, `SPEC-232` Nr. 1 und 5 a (dazu Nr. 2 und 3, soweit der Kern sie berührt)
- `AGENTS.md` (Hard Rules), Betriebsregeln des Plans (Hüllen, Rohbytes, Schwerlauf-Sperre)

**Arbeitsweise.** Alle Mutationen liefen auf Kopien in einem Wegwerfverzeichnis des Scratchpads (`rev-telport/`), in Containern (`odata-recorder:deps-mut` für `go test -count=1` über `ports/support` und `adapters/driven/telemetry`, `odata-recorder:race` für Detektor-Läufe), unter `flock /tmp/odata-recorder-heavy.lock`, mit `--memory=4g --memory-swap=4g --cpus=4`, Host unter hoher Last (Load Average 20 bis 95). Im Repo wurde nichts geändert außer diesem Report. Gefahren: 61 Code-Mutationen (Regeln, Port, Fake, Adapter), 9 Sperren-Varianten für die Raten, 2 `a-check`-Läufe auf Kopien, `make groessenmarke` im Zweig. Jede Mutation war vor dem Lauf auf Änderung geprüft (`cmp`). Eine Mutation (`len(s) != 71` zu `<` in `blobReferenz`) blieb grün und ist äquivalent: `nurZeichen` begrenzt die Länge des Rests ohnehin auf 64.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `TestWertformZeichen` führt 12 Vertreter der Menge „jedes Zeichen außerhalb der Form“, die Menge ist jedes Byte außer den erlaubten (und `SPEC-232` Nr. 2 nennt Klammern, die unter den 12 fehlen). Elf Mutationen an Elementen außerhalb der 12 bleiben grün: Kennung nimmt `(`, Tab oder DEL an; Kleinbuchstabe ab 0x60, Großbuchstabe bis `[`; Methode nimmt Kleinbuchstaben an; Token nimmt `-` oder einen Großbuchstaben in der Mitte an; Name nimmt `-` an; Hex nimmt `F` oder 0x60 an. Ein Wert wie `get` als `http.request.method` oder `(` in einer Kennung ginge durch, ohne dass etwas rot wird. | `AGENTS.md` §3.15 (ein Fall je Element einer Menge), §3.17 (die Klasse, nicht die genannten Elemente); `SPEC-230` Nr. 3; `SPEC-232` Nr. 2; `LH-QA-04`; `implement-slice` Schritt 14 | `src/adapters/driven/telemetry/regeln_test.go` · `zeichen := []string{` | ja — `go test` über das Paket `telemetry` mit einer der elf Mutationen (z. B. `kennungByte` um `c == '('` erweitert) | Zeichenklasse als Auswahl statt als Menge aller Bytes |
| F-2 | MEDIUM | Der Zweig „Code außerhalb der Tabelle“ in `CountError` hat keinen Fänger unter Nebenläufigkeit: `TestCountErrorGleichzeitig` ruft nur Codes aus `model.AlleCodes()`. Gibt die Mutation die Sperre vor `dropped[fehlercode]++` frei, bleiben Detektor (0 von 50), Wirkung ohne Detektor (0 von 50) und `-count=50` (grün) ohne Befund; die Zusage `SPEC-232` Nr. 5 a („verliert keinen Zuwachs“) gilt für beide Zweige. | `AGENTS.md` §3.10 (Zusage über Nebenläufigkeit gilt nur als gefangen, wenn die Mutation in jedem Lauf rot wird), §3.15 (jede Aufrufstelle / jeder Zweig) | `src/adapters/driven/telemetry/telemetry_test.go` · `a.CountError(codes[(g*aufrufe+i)%len(codes)])` | ja — `make test-race` (je Prozess `-test.count=1`) mit der Mutation „Entsperren vor `a.dropped[support.VerwurfFehlercode]++`“ | Nebenläufiger Zweig ohne Fänger |
| F-3 | MEDIUM | Die Zusage „Adapter ist für gleichzeitigen Aufruf sicher“ (`SPEC-232` Nr. 5, Satz 1) hat für `Log`/`verwerfen` und `Snapshot` keine Tabellenzeile, keine Mutation und keine Rate im Plan; der einzige Fänger ist `TestLogGleichzeitig`. Gemessen (50 Prozesse, `-test.count=1`, Host unter Last): `Snapshot` ohne Sperre unter dem Detektor rot in 44 von 50 (0 mit `DATA RACE`, die rote Wirkung ist `fatal error: concurrent map read and map write`), ohne Detektor in 42 von 50; `verwerfen` ohne Sperre Detektor 50 von 50, ohne Detektor 22 von 50; `verwerfen` mit Fenster zwischen Lesen und Schreiben Detektor 50 von 50 (0 mit `DATA RACE`), ohne Detektor 45 von 50, `-count=50` 48 Wiederholungen rot. | `AGENTS.md` §3.10 (Rate mit Zahl, nur „in jedem Lauf rot“ zählt), §3.15 (jede Zusage hat eine Zeile) | `src/adapters/driven/telemetry/telemetry_test.go` · `func TestLogGleichzeitig` | ja — Detektor je Prozess über `TestLogGleichzeitig` mit `Snapshot` ohne `a.mu.Lock()` | Nebenläufigkeits-Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt |
| F-4 | MEDIUM | `FakeSpan.End` hält den gerufenen Span nicht fest, ohne dass ein Test rot wird (Mutation `Aufruf{Op: OpEnd}` ohne `Span: s` bleibt grün). Der Kopf von `Aufruf` sagt „Span ist bei Set und End der gerufene Span“; `TestFakeAufzeichnung` prüft `Span` nur an `Set` (`a[2]`, `a[5]`) und an `StartSpan`, nie an `End` (`a[3]`, `a[4]`). Ein Einbau-Test, der aus dem Fake liest, welcher von zwei Spans endete, bekäme `nil`. | `AGENTS.md` §3.15 (der Fake hält jedes Argument fest); `SPEC-230` Nr. 10; Betriebsregel des Plans (Hüllen) | `src/hexagon/ports/support/supporttest/telemetrie.go` · `Aufruf{Op: OpEnd, Span: s}` | ja — `go test ./src/hexagon/ports/support/...` mit der Mutation | Festgehaltenes Feld des Fakes ohne Fänger |
| F-5 | MEDIUM | `Snapshot.Spans` ist im frischen Adapter ein nicht-nil, leerer Slice (`Spans: []support.SpanRecord{}`), und `TestSnapshotVollstaendig` weist `s.Spans == nil` als Fehler ab; `Snapshot.Kopie` und der Fake ohne gesetzten Stand liefern nil. Die Raster-Zelle `Lesen × leerer Wert` und `SPEC-230` Nr. 9 sagen „`Spans` leer“, ohne nil von leer zu trennen; der Code entscheidet die Randform, der Test hält sie fest. Ein Dienst, der den Stand als JSON ausgibt, sähe mit dem Adapter `[]`, mit dem Fake `null`. | `AGENTS.md` §3.16 (eine im Code entschiedene Zelle ist ein MEDIUM), §3.12; `SPEC-230` Nr. 9 | `src/adapters/driven/telemetry/telemetry.go` · `Spans:     []support.SpanRecord{},` | ja — `go test` über das Paket `telemetry` mit `Spans: nil` | Randform (nil gegen leer) im Code entschieden |
| F-6 | LOW | Die Auslegung „`Log` wendet `MinLevel` nicht an“ steht als Zusage im Kopf von `Log` und im Plan (§3, §6, Ausgang *entschieden*), hat aber keinen Fänger: die Mutation `if level < a.minLevel { return }` am Anfang von `Log` bleibt grün, weil jeder Test mit Attributen bei der Voreinstellung (`MinLevel` 0) und mindestens bei `Info` loggt. | `AGENTS.md` §3.11 (ein Kommentar sagt nur zu, was ein Test prüft), §3.15 | `src/adapters/driven/telemetry/telemetry.go` · `schreibt keine Zeile und wendet MinLevel nicht an` | ja — `go test` über das Paket `telemetry` mit der Mutation | Kommentar-Zusage ohne Test |
| F-7 | LOW | Der Kopf von `FakeSpan` sagt „ID zählt je Fake ab 1“; die Mutationen „kein Zählerschritt“ (alle IDs 0) und „ID konstant 1“ bleiben grün, kein Test liest `FakeSpan.ID`. | `AGENTS.md` §3.11 | `src/hexagon/ports/support/supporttest/telemetrie.go` · `ID zählt je Fake ab 1` | ja — `go test ./src/hexagon/ports/support/...` mit `ID: 1` | Kommentar-Zusage ohne Test |
| F-8 | LOW | Die Zusage P14 („importiert nur `model`, `ports/support` und die Standardbibliothek“) ist weiter als ihr Fänger: `.a-check.yml` erlaubt `adapters-driven` → `ports-outbound`; ein Import von `ports/outbound` im Adapter ergibt `gesamt: 0 Befund(e)`. Der Import von `services/record` ergibt `wrong-direction`, wie im Plan genannt. | `AGENTS.md` §3.15; Plan P14 | `docs/plan/planning/in-progress/slice-telemetry-port.md` · `Der Adapter importiert nur `model`, `ports/support` und die Standardbibliothek` | ja — `make a-check` auf einer Kopie mit `_ ".../ports/outbound"` im Adapter | Zusage weiter als ihr Fänger |
| F-9 | LOW | Die Raster-Zelle `Erzeugen (New) × Wiederholung` nennt `SPEC-232` Nr. 1 als Ort für „jeder Aufruf erzeugt einen eigenen Adapter ohne gemeinsamen Zustand“; der Wortlaut von Nr. 1 sagt das nicht, und es gibt keinen Fänger, der es zusagt (rot würde ein gemeinsamer Zustand nur zufällig in `TestCountErrorJeCode`). Der Code hält die Zusage (kein Paketzustand). | `AGENTS.md` §3.16 (eine Zelle nennt nur einen Ort, dessen Wortlaut die Entscheidung trägt) | `docs/plan/planning/in-progress/slice-telemetry-port.md` · `jeder Aufruf erzeugt einen eigenen Adapter ohne gemeinsamen Zustand` | nein — eine Plan-Prüfung ohne Gate | Raster-Ort trägt die Entscheidung nicht |
| F-10 | INFO | `TestKernSchreibtNichts` hält fest, was in `slice-telemetry-spans` ersetzt wird (`StartSpan` gibt den übergebenen `ctx` zurück, ein unbekannter Name zählt nichts, `Set` und `End` ohne Wirkung); der Test ist die Prüfung der Abgrenzung dieses Teils und gehört dem Slice `S` beim Einbau des Spans. Kein Handlungsbedarf in `P`. | Maintainability | `src/adapters/driven/telemetry/telemetry_test.go` · `func TestKernSchreibtNichts` | nein | Test pinnt vorläufiges Verhalten |
| F-11 | INFO | Der Nullwert `supporttest.FakeTelemetry{}` gerät beim ersten Aufruf in Panik (`zaehler` ist nil, Zuweisung an die Map); der Kopf sagt „gerät nicht in Panik“ und meint den Konstruktor `NewFakeTelemetry`. Betrifft Tests nachfolgender Slices, die den Fake als Literal anlegen. | `AGENTS.md` §3.11 (Hüllen-Betriebsregel: eine Panik des Fakes färbt den benannten Test) | `src/hexagon/ports/support/supporttest/telemetrie.go` · `gerät nicht in Panik` | ja — ein Test mit `&supporttest.FakeTelemetry{}` und `CountError` | Zusage des Kopfes gilt nur für den Konstruktor |

Zu den vom Auftrag ausgenommenen Punkten (`support.Verwuerfe()` und `Snapshot.Kopie()` ohne Nennung in `SPEC-230`, Dauer `-0.0`): der Code entscheidet darüber nicht mehr, als §6 des Plans mit Ausgang *offen* und Adresse Architect nennt. Sie sind kein Finding.

## Auflösung der Tabelle P01 bis P17 und des Rasters (Mutation je Zeile)

| Nr. | Mutation (selbst gefahren) | Ergebnis |
|---|---|---|
| P01 | `Int` mit `Ganz: v + 1`; `Millis` durch `Microsecond`; `Str` mit Art `Int` | rot (`TestAttrKonstruktoren`) |
| P02 | `timestamp` in die Menge `Log`; `befunde` aus der Menge `Log` | rot (`TestSchluesselmengen`) |
| P03 | Token 33; Status `> 100`; Status bis 600; Dauer: Unendlich angenommen / NaN angenommen; Art nicht verglichen; Name beginnt mit Ziffer; Kennung 65 | rot (`TestWertformGrenzen`, `TestWertformGrenzenVonHand`) |
| P04 | `ziffer` um `:` erweitert (Element der 12) rot; elf Elemente außerhalb der 12 grün | siehe F-1 |
| P05 | nur Präfix `ODR-E` geprüft | rot (`TestFehlercodeForm`) |
| P06 | Zuwachs um 2 | rot (`TestCountErrorJeCode`) |
| P07 | Verwurf `fehlercode` wird nicht gezählt | rot (`TestCountErrorUnbekannt`) |
| P08 | Verwurf-Gründe ohne `fehlercode` initialisiert | rot (`TestSnapshotVollstaendig`) |
| P09 | `.Kopie()` entfernt | rot (`TestSnapshotKopie`) |
| P10 | `>=` an 65536; 0 ungültig; Wert im Fehlertext; `MinLevel` über `Error` abgewiesen | rot (`TestNewConfig`) |
| P11 | `Snapshot` nicht gezählt; `Level`, Code, Attribute von `Log` und `Set`/`StartSpan`, `ctx`, Name nicht festgehalten; `Aufrufe()` teilt Speicher; `SetzeStand` und `Snapshot` ohne Kopie | rot (`TestFakeAufzeichnung`, `TestFakeSnapshotStand`); `End` ohne `Span` grün (F-4) |
| P12 | Rückgabe-`ctx` auf `Background`; Eltern nie gefunden; Eltern nicht festgehalten; `nil`-`ctx` nicht abgeleitet | rot (`TestFakeEltern`, `TestFakeAufzeichnung`) |
| P13 | nicht selbst gefahren; Zählprüfung 3, 2, 1 und Formen per Reflexion gelesen | ohne Befund (Lesung) |
| P14 | Import `services/record` im Adapter: `wrong-direction: adapters-driven -> services`; Import `ports/outbound`: 0 Befunde; `git diff a0396b0 -- go.mod go.sum` leer | rot bzw. F-8 |
| P15 | siehe Raten unten | rot in 50 von 50 unter dem Detektor; unbekannter Zweig F-2 |
| P16 | siehe Raten unten | rot in 50 von 50 |
| P17 | Wert vor Schlüssel; Verwurf auch bei ungültigem Wert belegt den Schlüssel; zweiter Wert gilt; nur der erste Verwurf je Aufruf gezählt | rot (`TestVerwurfRangfolge`) |
| Raster `New`/`CountError`/`Snapshot` | alle Zellen mit Ort oder „gilt nicht“ mit Grund; eine Zelle trägt ihren Ort nicht (F-9) | siehe F-9 |
| Raster 5 Zeilen Zusatztabelle | `TestVerwurfRangfolge` führt alle fünf Paare | ohne Befund |

## Raten der nebenläufigen Zusagen (selbst gemessen)

Messender: der Reviewer; Image `odata-recorder:race` (Test-Binary einmal gebaut), 50 Prozesse je Zelle mit `-test.count=1`, Host unter Last (Load Average 20 bis 95). *Detektor*: `go test -race -c`; *Wirkung*: ohne Detektor; *Wiederholung*: ein Prozess ohne Detektor, `-test.count=50`. „DR“ zählt die Prozesse mit `DATA RACE` in der Ausgabe.

| Test | Mutation | Detektor | Wirkung | Wiederholung |
|---|---|---|---|---|
| P15 | keine (Gegenprobe) | 0 von 50 rot | 0 von 50 | grün |
| P15 | `CountError` ohne Sperre | 50 von 50 rot, DR 50 | 50 von 50 | Absturz (`fatal error`) |
| P15 | Lesen, Entsperren, Schreiben | 50 von 50 rot, DR 0 (Wirkungs-Fänger) | 50 von 50 | 50 Wiederholungen rot |
| P15 | Zweig „unbekannter Code“ entsperrt vor `dropped++` (Suite) | 0 von 50 | 0 von 50 | grün (F-2) |
| P16 | keine (Gegenprobe) | 0 von 50 | 0 von 50 | grün |
| P16 | Aufzeichnung ohne Sperre | 50 von 50, DR 50 | 50 von 50 | Absturz |
| P16 | Zähler nach der Freigabe | 50 von 50, DR 50 | 50 von 50 | Absturz |
| P16 | Fenster beim Anhängen (Länge lesen, entsperren, anhängen) | 50 von 50, DR 0 | 50 von 50 | 50 Wiederholungen rot |
| `TestLogGleichzeitig` | keine (Gegenprobe) | 0 von 50 | 0 von 50 | grün |
| `TestLogGleichzeitig` | `verwerfen` ohne Sperre | 50 von 50, DR 50 | 22 von 50 | Absturz |
| `TestLogGleichzeitig` | `verwerfen` mit Fenster zwischen Lesen und Schreiben | 50 von 50, DR 0 | 45 von 50 | 48 Wiederholungen rot |
| `TestLogGleichzeitig` | `Snapshot` ohne Sperre | **44 von 50**, DR 0 | **42 von 50** | Absturz (F-3) |

Die Werte des Implementers für P15 und P16 (50 von 50 unter dem Detektor, 48 von 50 für „ohne Sperre“ und „Fenster“ von P15 ohne Detektor) bestätigen sich unter dem Detektor und in P16 vollständig; für P15 ohne Detektor fand ich 50 von 50 statt 48 von 50 (anderer Laststand, eine breitere Fenster-Variante mit zwei getrennten Sperren-Abschnitten), ohne Folge für die Aussage.

## Größenmarke

`make groessenmarke` im Zweig: `slice-telemetry-port.md: ok, Messung 1759 von Marke 3000`; Messpunkt `zeilen=17 nebenlaeufig=2 fakes=1 produktcode=580 schaetzung=2150 marke=3000 messung=1759`. Bestätigt.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/support/telemetry.go` (Port, Typen, `Spans()`, `Eimer()`, `Verwuerfe()`, `Kopie()`) | geprüft, ohne Befund (außer F-5 für die Seite des Adapters); 6 Operationen, 3 Arten, 11 Pflicht-Attribute, 11 Grenzen und 7 Gründe stimmen mit `SPEC-230` überein |
| `src/adapters/driven/telemetry/regeln.go` (Regeln, Rangfolge, Wertformen) | geprüft, ohne Befund am Code; die Formen gleichen Zeichen für Zeichen dem Wortlaut von `SPEC-230` Nr. 3 (Befund F-1 liegt im Fänger) |
| `src/adapters/driven/telemetry/telemetry.go` (`New`, `CountError`, `Snapshot`, `Log`, `StartSpan`) | geprüft, ohne Befund am Code; `New` weist ab wie `SPEC-232` Nr. 1, ohne Wert im Fehler; alle Zugriffe auf den Zustand liegen unter `mu` |
| Abgrenzung zu `slice-telemetry-log` und `slice-telemetry-spans` | geprüft, ohne Befund: `Log` schreibt keine Zeile, `writer`, `clock` und `minLevel` werden gehalten und nicht gelesen, `StartSpan` gibt den `ctx` zurück und einen Span ohne Wirkung, kein `Handler`, keine Senke, kein Ring, keine Verteilung mit Wirkung, kein Zählen von `spanname`, `zustand` oder `pflicht`, `MinLevel` nicht angewendet |
| `src/hexagon/ports/support/supporttest/telemetrie.go` (Fake) | geprüft; Mutationen an Level, Code, `ctx`, Attributen, Eltern, Kopien rot; verdeckte Eigenschaften: keine Hülle um ein Argument, `ctx` mit derselben Identität, nil-`ctx` bleibt nil in der Aufzeichnung; Ausnahmen F-4, F-7, F-11 |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft; kein Konjunktiv über verworfene Alternativen, kein abwesender Text; Zusagen ohne Test: F-6, F-7 |
| Rohbytes (`AGENTS.md` §3.18) | geprüft, ohne Befund: in allen elf Go-Dateien und im Plan keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP (`LC_ALL=C grep -P`); die Zeichen von P04 entstehen aus `[]byte` |
| `go.mod`, `go.sum`, Binary | geprüft, ohne Befund: `git diff a0396b0 -- go.mod go.sum` leer; kein Eintrag in `src/bootstrap` |
| Plan folgt dem Code (`AGENTS.md` §3.9) | geprüft, ohne Befund: §3 (Tabelle, Mutationen, Raten, Hilfen), §6 (Hilfen und `-0.0` mit Ausgang *offen* und Adresse Architect, `AGENTS.md` §3.13) und §7 (Messungen) nennen, was der Code liefert; Testnamen weichen von den Vorschlägen ab (P13), der Plan nennt sie Vorschläge |
| Spec-Straten | geprüft, ohne Befund: `SPEC-232` Nr. 1 und Nr. 5 f stehen im Wortlaut der Planner-Entscheidung; `spec/architecture.md` unverändert im Diff, die Zeile `TelemetryReader` steht seit dem Architect-Stand |
| Suppression (`AGENTS.md` §3.2) | geprüft, ohne Befund: keine `//nolint` im Diff |
| `src/bootstrap`, Handbuch, Katalog | geprüft, ohne Befund: nicht berührt |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 5 |
| LOW | 4 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Zeichenklasse als Auswahl statt als Menge aller Bytes · Nebenläufiger Zweig ohne Fänger · Nebenläufigkeits-Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt · Festgehaltenes Feld des Fakes ohne Fänger · Randform (nil gegen leer) im Code entschieden · Kommentar-Zusage ohne Test · Zusage weiter als ihr Fänger · Raster-Ort trägt die Entscheidung nicht

## Verdikt

**Merge-blockierend:** ja — fünf MEDIUM (F-1 bis F-5). Ein HIGH liegt nicht vor: der Produktcode folgt dem Wortlaut von `SPEC-230` bis `SPEC-232`; die Befunde betreffen Fänger, eine Randform und die Raten, nicht einen Korrektheitsfehler im kritischen Pfad. Die Fix-Runde bildet je Befund die Menge aus dem Wortlaut der Zusage (`AGENTS.md` §3.17): bei F-1 die Menge der 256 Bytes, bei F-2 und F-3 jeder Zweig und jede Operation unter der Sperre, bei F-4 jeder Operationstyp des Fakes. F-5 geht zur Entscheidung an den Architect (Ort in `SPEC-230` Nr. 9 oder Raster).

**Übergabe:** Findings gehen an den Implementer; F-5 zusätzlich an den Architect (Randform ohne Ort, `AGENTS.md` §3.12, §3.13: Zeile in §6 des Plans mit Ausgang *offen* und Adresse); F-9 an den Planner (Raster-Ort). Die Finding-Klassen gehen in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ist ein Lauf-Beleg und ersetzt keine Verifikation; DoD- und Spec-Konformität prüft der Verifier.
