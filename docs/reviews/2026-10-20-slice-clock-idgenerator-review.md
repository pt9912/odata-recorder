# Review: slice-clock-idgenerator

* Datum: 2026-10-20
* Rolle: reviewer (frischer Kontext, kein Self-Review)
* Review-Art: Code gegen Plan + Konventionen
* Gegenstand: `git diff fe372f7..HEAD` (e8d5972 Code, 7e97c49 Spec-Entscheidung Nr. 10, 5576700 Test-Nachzug)
* Plan: `docs/plan/planning/in-progress/slice-clock-idgenerator.md`
* Verträge: `SPEC-114` (`spec/spezifikation.md`), `LH-QA-02`, `LH-FA-13`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
* Maßstab: Plan, SPEC-114, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), AGENTS.md §3.7, §3.9 bis §3.16, Reviewer-Skill. Nicht geprüft: DoD (Verifier).
* Methode: Wegwerfkopie von `src/`, `go.mod` im Scratchpad, Image `odata-recorder:test` (`--network none`, Kopie nur lesend gemountet); rund 45 Mutationen (je Zeile der Tabelle Zusage·Fänger und je Rasterzelle). Die `a-check`-Mutation lief im Arbeitsbaum und wurde per `git checkout` zurückgenommen; der Arbeitsbaum ist sauber. Kein `-race` möglich (Image: "requires cgo").

## Zusammenfassung

Keine HIGH. Zwei MEDIUM, drei LOW, drei INFO. Die Tabelle ist überwiegend scharf: die Mutationen der Zeilen UTC, Rundung, Rücksprung, Nullzeit/Jahr 0/10000, frische Uhr (`gesehen`-Marke), Höchstwert fortschreiben, Port-Signaturen, `NewFestUhr` (Ortsjahr, Nullzeit, Grenzen, Untergrenze, `In(+02:00)`-Nullzeit), `Advance` (negativ, Bereich, Stand vor Panik, Obergrenze), `Now` ändert Stand, `%032x`, globaler Zähler, Präfixe (Code und Fake), 1 Byte Zufall, feste Kennung, Form (31 Zeichen), Tabelle ohne Methode, Methode ohne Tabellenzeile (beide Richtungen von `IDVertrag`), Seed je Aufruf: alle rot aus dem benannten Grund. `a-check` meldet bei Import von `adapters/driven/sqlite` in `clock` `lateral-adapter` rot. Zwei Plan-Zusagen sind nicht gefangen (F1, F2).

## Findings

### F1 MEDIUM — "Quelle wird unter der Sperre gelesen" hat keinen Fänger
* kategorie: MEDIUM
* quelle: AGENTS.md §3.15 (Zusage ohne Fänger), §3.11
* pfad: `docs/plan/planning/in-progress/slice-clock-idgenerator.md` · "die Quelle wird unter der Sperre gelesen"; Code `src/adapters/driven/clock/clock.go` · `t := u.quelle().UTC()`
* befund: Der Plan sagt, `TestSystemuhrNebenlaeufig` fange das Lesen der Quelle unter der Sperre. Mutation gefahren: `t := u.quelle().UTC()` vor `u.mu.Lock()` verschoben: alle Tests grün (Gate-Test-Set der drei Pakete). Die Klammer bleibt dabei monoton; die Aussage ist weder beobachtbar noch im Code oder der Spec zugesagt, steht aber als Fänger im Plan.
* Fehlerfall: Plan und Test weichen auseinander; wer die Zeile liest, hält die Lesereihenfolge für gefangen.
* verifizierbar: ja (Mutation im Scratch)
* klasse: Zusage des Plans ohne Fänger

### F2 MEDIUM — "keine feste Folge" (Nr. 7): Mutation "math/rand mit festem Seed" ist nur in einer Form rot
* kategorie: MEDIUM
* quelle: AGENTS.md §3.15, SPEC-114 Nr. 7
* pfad: `docs/plan/planning/in-progress/slice-clock-idgenerator.md` · "Mutation: `math/rand` mit festem Seed: rot"; Test `src/adapters/driven/clock/clock_test.go` · `TestZufallsIDsUnabhaengig`
* befund: Mutation A (neue Quelle `rand.NewSource(1)` bei jedem Aufruf) ist rot. Mutation B (ein paketweiter, mutexgeschützter `math/rand`-Generator mit festem Seed) ist grün: zwei Instanzen und 64 × 500 Aufrufe liefern verschiedene Kennungen, nur die Folge ist über Prozessstarts gleich. Der Plan nennt als akzeptiertes Negativ nur "anderer zufälliger Seed", nicht den festen Seed in der paketweiten Form.
* Fehlerfall: nach einem Neustart vergibt der Dienst dieselbe Kennungsfolge; `ErrBereitsVorhanden` trifft erzeugte Kennungen, obwohl Nr. 8 es als nicht eintretend zusagt.
* verifizierbar: ja
* klasse: Zusage des Plans ohne Fänger

### F3 LOW — Plan behauptet beobachtete Läufe, die das Image nicht trägt; falscher Fänger genannt
* kategorie: LOW (Konventions-Anker: AGENTS.md §3.11 / §3.7, Zusage nur was geprüft ist)
* quelle: AGENTS.md §3.11
* pfad: Plan §3 · "bei `-count=20` und unter `-race` rot gesehen"
* befund: Im Image `odata-recorder:test` scheitert `go test -race` ("requires cgo"). Mutation "Sperre in `Systemuhr.Now` entfernt", 20 Läufe der Pakettests: 11 rot, 9 grün; rot wird nur `TestVertrag` (Teil `Quelle in +02:00`, `UhrVertrag`), `TestSystemuhrNebenlaeufig` nie. Der Plan ordnet die Wirkung dem falschen Test zu; das akzeptierte Negativ bleibt gültig, die genannte Beobachtung ist nicht reproduzierbar.
* verifizierbar: ja
* klasse: Plan nennt Fänger oder Lauf, der die Mutation nicht fängt

### F4 LOW — Randform "Zähler-Darstellung" erst im Code, dann in der Spec entschieden
* kategorie: LOW (Konventions-Anker: AGENTS.md §3.12, §3.16; bereits im Slice repariert)
* quelle: AGENTS.md §3.12
* pfad: `src/hexagon/ports/support/supporttest/fest.go` · `fmt.Sprintf("%s-%032d", …)` (in e8d5972); `spec/spezifikation.md` SPEC-114 Nr. 10 "dezimal" (in 7e97c49)
* befund: Der erste Code-Commit entschied Dezimal statt Hex im Code; Nr. 10 hielt es erst im Nachzug fest, und der Test verglich zuerst gegen das Format des Codes (5576700). Der Nachzug ist in Spec, Plan §3 und Raster erfolgt; kein offener Zustand.
* verifizierbar: nein (Historie)
* klasse: Randform im Code vor der Spezifikation entschieden (Beobachtung `BEO-REPO/randform-erst-im-review-entschieden`, Wiederholung)

### F5 LOW — Raster-Zelle "Zufallsquelle fällt aus" ohne Fänger und ohne akzeptiertes Negativ
* kategorie: LOW (Konventions-Anker: AGENTS.md §3.15 verlangt eine Zeile je Randform; hier ohne Auswirkung auf Verhalten, da die Laufzeit den Prozess beendet)
* quelle: AGENTS.md §3.15, §3.16
* pfad: Plan §6 Raster · "Zufallsquelle fällt aus: Prozess endet, kein Fehlerwert (Nr. 7)"; Code `clock.go` · `_, _ = rand.Read(b[:])`
* befund: Die Zelle ist in §6 entschieden, §3 führt dazu weder Fänger noch ein akzeptiertes Negativ; die Spec-Zeile *Ungeprüft* nennt die Güte der Quelle, nicht den Ausfall. Der Kommentar "rand.Read endet bei Ausfall der Quelle den Prozess" sagt eine Laufzeit-Eigenschaft zu, die kein Test prüft; zudem ist "endet … den Prozess" ungrammatisch (beendet).
* verifizierbar: nein
* klasse: Raster-Zelle ohne Fänger oder benanntes Negativ

### F6 INFO — Präfixe "verschieden" nur über die Literale der Testtabelle gefangen
* kategorie: INFO
* quelle: SPEC-114 Nr. 6, 8
* pfad: `src/hexagon/ports/support/supporttest/vertrag.go` · `praefixe()`
* befund: Ändert man Code und Testtabelle gemeinsam auf dasselbe Präfix (`exc`→`ses`), bleibt `TestVertrag` grün; eine Einzeländerung ist rot. `TestFestIDsFolge` hält die Literale der Fakes. Kein Fänger für "verschieden" als Menge; die Aussage ist durch die festen Literale praktisch getragen. Keine Aktion erwartet.
* verifizierbar: ja
* klasse: Allaussage über Literale statt über die Menge

### F7 INFO — Nullwert `Systemuhr{}` und Plan-Zahlen
* kategorie: INFO
* quelle: Maintainability; AGENTS.md §3.9, §3.14
* pfad: `clock.go` · `type Systemuhr struct`; Plan §4 "rund 700 Zeilen"
* befund: (a) Der Nullwert `clock.Systemuhr{}` hat `quelle == nil` und paniert bei `Now`; Konstruktor `NewSystemuhr` ist der einzige Weg (außerhalb des Pakets nicht abgesichert). (b) Metrik aus Plan §4 (`git diff --numstat fe372f7..HEAD` über `src/…`, Tests und `spec/spezifikation.md`): 609 + 3 = 612 Zeilen gegen Marke 1000, Schätzung 700: keine Rückführung. Plan nennt den Ist-Wert noch nicht (kommt mit Closure). `harness/gegenproben.stempel` (2 Zeilen) liegt nicht unter den Plan-Pfaden und zählt nicht.
* verifizierbar: ja (Zeile (b))
* klasse: —

### F8 INFO — Schichten und Zeitquelle
* kategorie: INFO
* quelle: `.a-check.yml`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
* pfad: `src/adapters/driven/clock/clock.go` · Importe
* befund: `make a-check` in der Mutation (Import `adapters/driven/sqlite`) rot mit `lateral-adapter`; `clock` importiert sonst nur `ports/support`. Der gleiche Lauf listet Dateien unter `tools/` als ungeschichtet; das gilt unabhängig vom Diff.
* verifizierbar: ja
* klasse: —

## Auflösung der Zeilen (Kurz)

* Zeilen UTC, Rundung, Rücksprung, Bereich, Port, `NewFestUhr`, `Advance`, Präfix/Form, Kollision, `FestIDs` (Folge, Dezimal, je Operation, Paar), Fakes im Vertrag, Schichten: Mutation rot (siehe Zusammenfassung).
* Nebenläufigkeit: `NoClamp` rot in `TestSystemuhrNebenlaeufig`; Sperre entfernen nur über `TestVertrag` und nicht sicher (F3); `FestUhr`-Sperre entfernen: `TestFestUhrNebenlaeufig` rot; `FestIDs`-Sperre entfernen: rot (Absturz durch Map-Schreiben).
* Raster: alle Zellen mit Ort tragen eine rote Mutation außer "Zufallsquelle fällt aus" (F5) und "feste Folge" (F2).
* Panik-Pfade der Fakes: Stand bleibt unverändert (Mutation "Stand vor Panik" rot; `d < 0` und Bereich vor der Zuweisung).
* `IDVertrag` in beide Richtungen: rot bei sechster Methode im Port und bei Tabellenzeile ohne Methode.
* Kommentare sonst: Zusagen decken sich mit den Tests; keine verworfenen Alternativen; `Now`-Kommentar von `Systemuhr` erwähnt die Sperre nicht.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/support/` | geprüft, ohne Befund |
| `src/hexagon/ports/support/supporttest/` | geprüft, Befund nur F6 |
| `src/adapters/driven/clock/` | geprüft, Befunde F1, F2, F5, F7 |
| `spec/spezifikation.md` (SPEC-114, Historie) | geprüft, ohne Befund |
| Plan-Nachzug §1, §3, §6, Kopf | geprüft, Befunde F1 bis F3 |
| `harness/gegenproben.stempel` | geprüft, ohne Befund (reine Eingabeänderung) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 3 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Zusage des Plans ohne Fänger · Plan nennt Fänger oder Lauf, der die Mutation nicht fängt · Randform im Code vor der Spezifikation entschieden · Raster-Zelle ohne Fänger oder benanntes Negativ

## Verdikt

**Merge-blockierend:** ja, nach der Regel (MEDIUM F1, F2, AGENTS.md §3.15). Sachlich ohne Laufzeitfehler: beide MEDIUM betreffen Zusagen, die im Plan stehen, aber nicht gefangen sind, kein Verhaltensfehler im Code. Übergang ohne Code-Änderung möglich, wenn der Implementer die Zeilen als akzeptiertes Negativ mit Grund führt oder die Fänger ergänzt; die Entscheidung liegt bei ihm.

**Übergabe:** Findings an den Implementer; Klassen in die Closure §7.
