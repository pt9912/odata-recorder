# Review-Report: slice-race-gate — 2026-10-22

**Review-Art:** Code — geprüft gegen Plan (`slice-race-gate` §1, §3 mit der Tabelle *Zusage · Fänger* Z1 bis Z11 und N1 bis N3, §4, §6 mit Raster und Zusatztabelle), Verträge (`SPEC-122` Nr. 1 bis 10, `SPEC-102` Fälle 34 bis 36 mit Aufbau und Zeitbudget, `SPEC-036`), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md) und Hard Rules (`AGENTS.md` §3.1 bis §3.16), nicht gegen die DoD (Verifier). [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md) (Accepted) und das Zeitbudget von 120 s in `SPEC-102` sind Entscheidungen des Nutzers und nicht Gegenstand.

**Gegenstand:** `slice-race-gate` (Welle `welle-record-replay`), `git diff a7def59..HEAD` mit HEAD `790122c`. Implementer: `4acb295`, `2727440`, `16fee8f`, `947a1aa`, `49b4a87`, `f498de2`, `a0841a0`, `dcc720e`, `0376b4d`, `92668e6`; Architect: `cb49d16`, `790122c`. Der Diff berührt `Dockerfile` (Stufe `race`), `harness/mk/go-gates.mk`, `harness/mk/gegenproben.mk`, `harness/README.md`, `harness/gegenproben.stempel`, `tools/pruefung/`, `tools/gegenproben/`, drei Test-Kommentare unter `src/` (nur Kommentarzeilen, kein Code), die Verträge und den Plan. Größe nach `git diff --numstat 73e123e..HEAD` über `Dockerfile`, `harness/mk`, `tools/pruefung`, `tools/gegenproben`, `src`: 422 Zeilen, Marke 800 (Schätzung 620); die Messung nach dem Review trägt der Planner in §4 ein.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `790122c` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-22

**Methode.** Jede Zeile der Tabelle *Zusage · Fänger* (Z1 bis Z11, N1 bis N3) und jede Zelle von Raster und Zusatztabelle ist mit eigenen Mutationen aufgelöst; die Mutationstabelle des Implementers lag mir nicht vor, und was der Plan an Mutationen nennt, habe ich nachgefahren, nicht übernommen. 74 Mutationen, alle in einer Wegwerfkopie aus `git archive HEAD` (ohne den `.git` des Repos; für `make gegenproben`, dessen Prüfsumme einen Git-Baum verlangt, ein leeres `git init` in der Kopie), je als Patch mit `git apply` auf eine frische Kopie gelegt; der Repo-Baum und der Stempel `harness/gegenproben.stempel` blieben unberührt (`git status` vor und nach dem Lauf: allein `.claude/agents/architect.md`, das gehört dem Nutzer). Ebene der Einheitstests: die Stufen `pruefung` und `gegenproben` des `Dockerfile` (der erste Schritt von `make test` und `make gegenproben`, 4 bis 40 s je Mutation); Ebene der Fälle: `make gegenproben` in der Kopie (36 Fälle, unmutiert 36 ok, 85 s); für Z5 bis Z8 trägt die Kopie des `Dockerfile` zusätzlich eine Stufe `pruefung` ohne `go vet` und `go test`, damit der Test von Z1 nicht vor dem Fall rot wird; Stufe `race` und Stufe `test` über `make test-race` und `make test`. Rot zählt nur, wenn die Ausgabe den Test oder Fall nennt, den die Mutation treffen soll. Raten (Tabelle unten): ein eigenes Dockerfile auf dem Image von `RACE_IMAGE` (`go test -race -c`, danach das Testbinary als eigener Prozess mit `-test.count=1`, Zähler der roten Läufe und der Läufe mit `WARNING: DATA RACE`). Zwei Läufe von `make gegenproben` zugleich auf dem 20-Kern-Host rissen die Fallfrist von 120 s für die lint-Fälle („Zeit überschritten“); ich habe die Läufe danach einzeln wiederholt und zähle für Z5 bis Z8 nur die Urteile der Fälle `race/…`.

**Eingangs-Kontext:**

- Slice-Plan `slice-race-gate` in `790122c` (§1, §3, §4, §6) und der Geber der Sendung `slice-upstream-adapter` (§3, A15)
- [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)
- [LH-QA-03](../../spec/lastenheft.md#lh-qa-03--isolation-und-konsistenz), [LH-QA-06](../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit)
- `SPEC-122`, `SPEC-102` (Fälle 34 bis 36, Aufbau, Zeitbudget), `SPEC-036` (Randformen 1, 2, 7, 10, 11, 13, 14)
- Vorherige Reports am Modul: `2026-10-22-slice-upstream-tls-review` (Klassen „Fänger trennt die Mutation nicht (ein Fall je Element)“, „Akzeptiertes Negativ ohne Grund für den Zweig“)
- `AGENTS.md` (Hard Rules)

---

## Antworten auf die Fragen des Auftrags

**(1) Tabelle *Zusage · Fänger* (Z1 bis Z11, N1 bis N3).** Z1 bis Z9 und N1 bis N3: jede genannte Mutation ist aus dem genannten Grund rot, jede Fixture trennt (Z3: `TestRaceBefund` trennt `go`-Exit 2 von Exit 1 nur mit dem Fall `Exit 2`; Z1: `TestRaceOhneBefund` vergleicht die ganze Liste der Aufrufe, `gofmt` und `go vet` vor `go test` werden allein darüber rot; Z6: je eine Mutation für jeden der drei Teilbäume, jede rot an der fehlenden Zeile `FAIL` des Pakets). Z10 und Z11 haben Lücken: Z10 deckt die Stufe `pruefung` nicht (F-1), Z11 deckt `Wegwerfmodul` in den Fällen 35 und 36 und `Env` in Fall 34 nicht (F-2). Eine Mutation, die der Plan nicht nennt, deckt in Z3 eine Grenze auf (F-3). Die Zahlen und Mutationen stehen in der Tabelle unten.

**(2) Raster (§6) und Zusatztabelle.** Jede Zelle nennt einen Ort, dessen Wortlaut die Entscheidung trägt, und der Code trifft sie so (`pruefung race` verzweigt weder nach Muster noch nach Ursache; `Wurzel` fehlt oder ist kein Verzeichnis: Randform 2, Mutationen z4b und z4f rot). Keine im Code entschiedene Zelle gefunden. Die Zusatzzeilen *Rennen und Zeitgrenze* (Fall 36: z5, z8a, z8b rot), *Rennen in zwei Paketen* (Fall 35: z6a bis z6d rot) und *Rennen und roter Test* (viertes Paket von Fall 35, Fixture) haben ihren Fänger; die Zeile *Weitere Paare* begründet das fehlende Paar mit Gleichwertigkeit („der Code verzweigt nicht nach der Ursache“), und die Mutation `l.code > 0` bleibt grün (F-3). Die Zeilen *Plattform* haben nach dem Plan keinen Fänger und kein Host, ich habe sie nicht gefahren; die Zeile *Cache* ist am wiederholten `make gates` bestätigt (die Stufe `race` steht dort `CACHED`).

**(3) Vertrag gegen Code.** `SPEC-122` Nr. 1 bis 3, 5, 10: der Code gleicht dem Wortlaut (Argumente, `CGO_ENABLED=1` nur für den Aufruf, beide Ströme im Befund, keine Ursachenunterscheidung, kein `-timeout`, kein `-failfast`, `-count=1`, drei Aufrufformen, `Aufruf:` mit drei Formen). Nr. 4: Digest und Tag des Pins in `harness/mk/go-gates.mk` gleichen dem Vertrag; der Versionsvergleich der ersten Zeile von `/usr/local/go/VERSION` und seine vier Meldungen (`Go-Version von race … weicht von deps … ab`, `head: cannot open`, `base name (${RACE_IMAGE}) should not be blank`, `stat /bin/sh: no such file or directory`) sind gefahren; jeder `RUN` der Stufe trägt `--network=none` (zwei); die Stufen `deps`, `source`, `test`, `build` bleiben ohne cgo (rot an `cgo: C compiler "gcc" not found`), `pruefung` nicht (F-1). Nr. 6: ein weiterer `make gates` ohne Änderung fährt die Stufe nicht erneut. Nr. 8: die Messung „1 roter Lauf von 20“ mit `-test.count=20` ist nachgefahren (5 Prozesse: je 1 `FAIL`, 19 `PASS`, 1 `WARNING: DATA RACE`). Nr. 9: die Verdrahtung hat keinen dauerhaften Fänger, wie der Vertrag sagt (n1: `make docs-check` bleibt ohne `test-race` in `GATE_CHECKS` grün, `make -n gates` führt die Stufe dann nicht); der Beleg des Implementers ist nachgefahren (n1b: ein Rennen in einer Datei unter `src/hexagon/model` der Kopie lässt `make test-race` rot werden, `make test` bleibt grün). `SPEC-102` Fälle 34 bis 36: Name, Gate, Argumente, Fixture und Erwartung gleichen dem Vertrag; das Zeitbudget (85 s im unmutierten Lauf gegen 120 s) hält. `SPEC-036` Randformen 1, 2, 7, 10, 11, 13, 14 gelten für `race` wörtlich.

**(4) Raten (`AGENTS.md` §3.10).** Der Bericht des Implementers lag mir nicht vor; ich habe die Raten selbst gemessen und nenne Größe und Ergebnis in der Ratentabelle. Fall 35 und 36: 200 von 200 Prozessläufen je Teilbaum (Fixtures nach `tools/gegenproben/faelle.go` nachgeschrieben), Fall 36 50 von 50 mit beiden Meldungen. A15: die Mutation „Map am Client für alle Aufrufe als Header“ 200 von 200 mit und ohne Detektor; die Form ohne falsche Antwort unter dem Detektor 200 von 200, ohne ihn 12 von 200 (F-4). B19: ohne Detektor 100 von 100 rot.

**(5) Die drei Test-Kommentare (`AGENTS.md` §3.11).** `git diff a7def59..HEAD -- src`: drei Dateien, nur Kommentarzeilen, kein Bestandscode verändert. A15: „rot in 200 von 200 Läufen“ stimmt für die genannte Mutation (a15m1 unter `-race`: 200 von 200); der Satz zur Wirkung ohne Detektor trägt sie nicht (F-4); „20 Aufrufe“ und „jede Antwort trägt ihre Werte“ stimmen mit dem Test. B19: „meldet make test-race als Datenrennen“ stimmt (100 von 100, alle mit `DATA RACE`); „hoher Wahrscheinlichkeit“ steht ohne Zahl (F-5). B8: „der Detektor von make test-race meldet es nicht, weil jeder Zugriff weiter unter der Sperre liegt“ stimmt (Mutation b8a: 22 von 30 rot unter `-race`, keiner der 22 trägt `DATA RACE`; ohne `-race` 2 von 30); „mit Wahrscheinlichkeit, nicht mathematisch sicher“ stimmt (I-2).

**(6) Haltepunkt-Zusage.** `make test-race` auf dem Quellstand von HEAD, zweimal erzwungen (`--no-cache-filter race`, Image lokal): beide Läufe `pruefung race: ok`, Stufe 58,7 s und 59,0 s, Wanduhr 72,6 s und 72,2 s; kein Befund, kein Datenrennen, kein `checkptr`, keine Zeitgrenze, `modernc.org/*` ohne Meldung (ein Lauf mit Befund hätte den Text von `go test` gezeigt). Keine Abweichung zwischen den Läufen.

**(7) Kommentare in `Dockerfile`, `harness/mk/`, `tools/pruefung/`, `tools/gegenproben/`.** Die Kommentare von `Dockerfile` (Stufe `race`, Stufe `pruefung`, Stufe `gegenproben`), `go-gates.mk` und `gegenproben.mk` beschreiben, was da ist; die Aussagen „bricht ab, wenn seine Go-Version … abweicht“, „jeder RUN läuft ohne Netz“, „die Stufen deps bis build bleiben bei CGO_ENABLED=0“ stimmen mit den Läufen überein (die letzte als Zusage mit der Lücke F-1). Die Kommentare von `tools/pruefung/` tragen nur Zusagen, die ein Test fängt (Z1 bis Z4). In `tools/gegenproben/` ist „Kopie“ für die Fälle 34 bis 36 ungenau (F-6), und der Satz „jeder schreibt sein Wegwerfmodul selbst“ im Kopf von `raceFaelle` und „die Fälle zu `pruefung race` gegen ein Wegwerfmodul statt der Eingaben“ in der Stufe `gegenproben` sagen zu, was für zwei der drei Fälle kein Test fängt (F-2).

**(8) Gate-Kette und Laufzeit.** `make -n gates` führt die Ziele in der Reihenfolge `a-check`, `baseline-verify`, `docs-check`, `build`, `test`, `test-race`, `lint-profil`, `lint`, `schema-check`, `migration-check`; der Plan nennt keine Reihenfolge, `test-race` hängt hinter `test` (Zeile `GATE_CHECKS += build test test-race`). `make gates` bei warmem Cache: 11,1 s und 10,5 s, jeweils Rückgabecode 0; ein erster Lauf nach meinen Mutationsläufen brauchte 24,3 s (Ursache nicht untersucht). Der Satz in `harness/README.md` *Nicht behauptet* („dass die Stufen test, build und race … pruefung rufen und test-race an GATE_CHECKS hängt (kein Fall der Gegenproben fängt es)“) stimmt mit N1 und mit dem Wortlaut des Plans überein (n1, n1b, zd1 bis zd3).

**(9) Plan folgt dem Code (`AGENTS.md` §3.9).** Kopf, §1, §3 (Dateitabelle, Tabelle), §4 (Pfade, Marke, Schätzung) und §6 (Ausgänge, Messungen) stimmen mit dem Stand überein (Z9 an die gesehenen Meldungen angeglichen; Zeitbudget 120 s eingetragen; die zwei Läufe von `make test-race` und die 11 s von `make gates` gleichen meinen). Die Pfade der Marke sind vollständig: jede geänderte Datei des Diffs unter `Dockerfile`, `harness/mk/`, `tools/pruefung/`, `tools/gegenproben/` und `src/` steht in §3 oder §4; `harness/README.md`, `harness/vertraege/`, der Stempel und die Dokumente zählen nach §4 nicht. 422 Zeilen gegen die Marke 800.

**(10) Adressaten.** F-1 und F-4 (mit F-5) tragen Architect und Implementer zugleich: der Wortlaut steht im Vertrag oder im Auftrag des Plans (§3, *Kommentare in Tests*), die Umsetzung beim Implementer. F-2, F-3, F-6: Implementer; die Gleichwertigkeitsbegründung zu F-3 steht beim Architect.

---

## Mutationen je Zeile

Ergebnis „rot“ nennt den Test oder Fall, an dem die Mutation rot wird; „grün“ heißt: nichts wurde rot.

| Zeile | Mutation | Ergebnis |
|---|---|---|
| Z1 | z1a `-race` entfernt | rot `TestRaceOhneBefund`, `TestRaceMuster` (Liste der Aufrufe) |
| Z1 | z1b `-count=1` entfernt | rot `TestRaceOhneBefund`, `TestRaceMuster` |
| Z1 | z1c `-failfast` ergänzt | rot `TestRaceOhneBefund`, `TestRaceMuster` |
| Z1 | z1d `-timeout=1h` als festes Argument | rot `TestRaceOhneBefund`, `TestRaceMuster` |
| Z1 | z1e Standardmuster `./src/...` | rot `TestRaceOhneBefund` |
| Z1 | z1f Muster verworfen (immer `./...`) | rot `TestRaceMuster` |
| Z1 | z1g `gofmt -l .` vor `go test` | rot `TestRaceOhneBefund`, `TestRaceMuster`, `TestProgrammNichtAufrufbar/race_ohne_go` |
| Z1 | z1h `go vet ./...` vor `go test` | rot `TestRaceOhneBefund`, `TestRaceMuster`, `TestProgrammNichtAufrufbar/race_ohne_go`, `TestRaceBefund` |
| Z1 | z1i `-trimpath` entfernt | rot `TestRaceOhneBefund`, `TestRaceMuster` |
| Z1 | z1j `-buildvcs=false` entfernt | rot `TestRaceOhneBefund`, `TestRaceMuster` |
| Z1 | z1k Muster vor die festen Argumente | rot `TestRaceOhneBefund`, `TestRaceMuster` |
| Z1 | z1l Argumente mit `-` verworfen | rot `TestRaceMuster` |
| Z2 | z2a Umgebung unverändert durchgereicht | rot `TestRaceSetztCgo` (Vorher `0` und leer) |
| Z2 | z2b Umgebung des Kindes allein `CGO_ENABLED=1` | rot `TestRaceSetztCgo` (`PRUEFUNG_PROBE` fehlt) |
| Z2 | z2c `CGO_ENABLED=0` gesetzt | rot `TestRaceSetztCgo` |
| Z2 | z2d Umgebung des Aufrufers hinter den Eintrag | rot `TestRaceSetztCgo` |
| Z3 | z3a nur stderr von `go test` im Befund | rot `TestRaceBefund` (`FAIL x` fehlt) |
| Z3 | z3b nur stdout von `go test` im Befund | rot `TestRaceBefund` (`WARNING: DATA RACE` fehlt) |
| Z3 | z3c `ok`-Zeile nach stderr | rot `TestRaceOhneBefund`, `TestStroeme` |
| Z3 | z3d Exit 1 von `go` gilt als ok | rot `TestRaceBefund`, `TestStroeme` |
| Z3 | z3e Exit-Code von `go` durchgereicht | rot `TestRaceBefund` (nur der Fall `Exit 2`; `Exit 1` bleibt grün) |
| Z3 | z3f Präfix `go test -race:` entfernt | rot `TestRaceBefund` |
| Z3 | z3g Text `pruefung race ok` | rot `TestRaceOhneBefund`, `TestStroeme` |
| Z3 | z3h Befund mit Exit 2 | rot `TestRaceBefund`, `TestStroeme` |
| Z3 | z3i Sonderfall `DATA RACE` mit Exit 3 | rot `TestRaceBefund` |
| Z3 | z3j `l.code > 0` statt `l.code != 0` | grün (F-3) |
| Z4 | z4a `race` ohne `Wurzel` angenommen | rot `TestAufrufFehlerStartenNichts` |
| Z4 | z4b Prüfung der Wurzel entfällt | rot `TestWurzelFehlt` |
| Z4 | z4c `go` nicht aufrufbar gibt Exit 1 | rot `TestProgrammNichtAufrufbar/race_ohne_go` |
| Z4 | z4d leeres Argument für `race` angenommen | rot `TestAufrufFehlerStartenNichts` |
| Z4 | z4e `race` schreibt eine Datei in `Wurzel` | rot `TestSchreibtNichts` |
| Z4 | z4f `Wurzel` als Datei angenommen | rot `TestWurzelFehlt` |
| Z5 | z5 `-race` entfernt (Ebene der Fälle) | rot Fall `race/rennen-in-jedem-teilbaum` und `race/rennen-und-zeitgrenze` (`WARNING: DATA RACE` fehlt), Fall 34 ok |
| Z6 | z6a Standardmuster nur `./src/hexagon/...` | rot `race/rennen-in-jedem-teilbaum` (`FAIL` des Pakets unter `adapters` fehlt) |
| Z6 | z6b Standardmuster ohne `hexagon` | rot `race/ohne-rennen`, `race/rennen-in-jedem-teilbaum` (`FAIL` unter `hexagon` fehlt) |
| Z6 | z6c Standardmuster ohne `adapters` | rot `race/ohne-rennen`, `race/rennen-in-jedem-teilbaum` (`FAIL` unter `adapters` fehlt) |
| Z6 | z6d Standardmuster ohne `bootstrap` | rot `race/ohne-rennen`, `race/rennen-in-jedem-teilbaum` (`FAIL` unter `bootstrap` fehlt) |
| Z7 | z7a Befund bei jedem Lauf | rot `race/ohne-rennen` |
| Z7 | z7b cgo nicht gesetzt | rot `race/ohne-rennen`, `race/rennen-in-jedem-teilbaum`, `race/rennen-und-zeitgrenze` (`-race requires cgo`) |
| Z8 | z8a `-timeout=1h` hinter die Muster | rot `race/rennen-und-zeitgrenze` („Zeit überschritten“ nach 120 s) |
| Z8 | z8b Befund nur mit stderr | rot `race/rennen-in-jedem-teilbaum`, `race/rennen-und-zeitgrenze` (`go test` schreibt beide Meldungen nach stdout) |
| Z8 | z8c Befund nur mit stdout | grün in allen 36 Fällen; rot in `TestRaceBefund` (z3b), die Fälle tragen es nicht |
| Z9 | `RACE_IMAGE` `golang:1.27.3` | rot, `Go-Version von race (go1.27.3) weicht von deps (go1.27.1) ab` |
| Z9 | `RACE_IMAGE` `golang:1.27.2` | rot, `Go-Version von race (go1.27.2) weicht von deps (go1.27.1) ab` |
| Z9 | `RACE_IMAGE=` leer | rot, `base name (${RACE_IMAGE}) should not be blank` |
| Z9 | Build-Argument weggelassen | rot, `exec: "/bin/sh": stat /bin/sh: no such file or directory` |
| Z9 | Image `debian:bookworm-slim` (ohne Go) | rot, `head: cannot open '/usr/local/go/VERSION'` |
| Z9 | Image von `deps` (Alpine, ohne Compiler) | rot, `cgo: C compiler "gcc" not found` |
| Z9 | z9g Vergleich auf `[ -n "$r" ]` geschwächt, Image `golang:1.27.3` | grün (die Meldung der Zeilen oben hängt am Vergleich) |
| Z10 | z10a `ENV CGO_ENABLED=1` in Stufe `test` | rot `make test`, `cgo: C compiler "gcc" not found` |
| Z10 | z10b in Stufe `build` | rot `make build`, dieselbe Meldung |
| Z10 | z10d in Stufe `source` | rot `make test` und `make build` |
| Z10 | z10e in Stufe `deps` | rot `make test` |
| Z10 | z10c in Stufe `pruefung` | grün: `make test` (`pruefung test: ok`) und `make build` (`pruefung build: ok`) (F-1) |
| Z11 | z11a Fall `race/ohne-rennen` umbenannt | rot `TestFaelleDerPruefungHabenDieNamenAusDemVertrag` |
| Z11 | z11b Fall `race/rennen-in-jedem-teilbaum` umbenannt | rot `TestFaelleDerPruefungHabenDieNamenAusDemVertrag` |
| Z11 | z11c Fall `race/rennen-und-zeitgrenze` gestrichen | rot `TestFaelleDerPruefungHabenDieNamenAusDemVertrag` |
| Z11 | z11d `Wegwerfmodul` ohne Wirkung (immer Kopie) | rot `TestWegwerfmodulKopiertDieEingabenNicht` |
| Z11 | z11i `Env` erreicht das Gate nicht | rot `TestFallEnvErreichtDasGate` |
| Z11 | z11j ein Zeichen über ASCII in einer Quelle der Fälle | rot `TestRaceFaelleFuehrenNurASCII` (vier Quellen) |
| Z11 | z11g Fall 34 ohne `Wegwerfmodul` | rot `race/ohne-rennen` (Bau-Fehler der kopierten Pakete) |
| Z11 | z11e Fall 35 ohne `Wegwerfmodul` | grün: 36 Fälle, 0 rot (F-2) |
| Z11 | z11f Fall 36 ohne `Wegwerfmodul` | grün: 36 Fälle, 0 rot (F-2) |
| Z11 | z11h Fall 34 ohne `Env` | grün: 36 Fälle, 0 rot; die Stufe `gegenproben` setzt `CGO_ENABLED=0` selbst (F-2) |
| N1 | n1 `test-race` aus `GATE_CHECKS` | `make docs-check` grün (0 Befunde), `make -n gates` führt die Stufe nicht |
| N1 | n1b Rennen in einer Datei unter `src/hexagon/model` der Kopie | rot `make test-race` (`WARNING: DATA RACE`, `FAIL … src/hexagon/model`), `make test` grün |
| N3 | zd1 `COPY --from=deps /go/pkg/mod` entfernt | rot, `go` holt `modernc.org/sqlite` und scheitert am Netz (Vertrag Nr. 4) |
| N3 | zd2 `--network=none` am Lauf entfernt | grün (Vertrag Nr. 9: kein Fänger, Grund und Folge dort) |
| N3 | zd3 `ENV`-Zeile der Stufe entfernt | grün (Vertrag Nr. 9) |
| Kommentar A15 | a15m1 eine `http.Header`-Map am Client für alle Aufrufe als Header | rot unter `-race` 200 von 200 (mit `DATA RACE`), ohne `-race` 200 von 200 (falsche Antwort) |
| Kommentar A15 | a15m2 eine `http.Header`-Map am Client in jedem `Senden` beschrieben, nicht verwendet | rot unter `-race` 200 von 200 (`DATA RACE`), ohne `-race` 12 von 200 |
| Kommentar B19 | b19m Sperre der Folge entfernt | rot unter `-race` 100 von 100 (`DATA RACE`), ohne `-race` 100 von 100 (Wirkung) |
| Kommentar B8 | b8a Sperre in `nimmAn` zwischen Prüfung und Annahme freigegeben und neu genommen | rot über die Wirkung unter `-race` 22 von 30 (kein `DATA RACE`), ohne `-race` 2 von 30 |
| Kommentar B8 | b8b Sperre in `sperre` zwischen Setzen des Drain und Blick auf die laufenden freigegeben | grün 0 von 30 mit und ohne `-race` (I-2) |

Von 74 Mutationen blieben 10 grün: z3j, z8c (rot im Test), z10c, z11e, z11f, z11h, zd2, zd3 (der Vertrag nennt beide als Negativ), z9g (der geschwächte Vergleich ist die Gegenprobe zu Z9) und b8b (I-2). Unerwartet grün sind z3j, z10c, z11e, z11f, z11h.

## Raten

Je Lauf ein eigener Prozess mit `-test.count=1`; „mit DR“ zählt die Läufe, deren Ausgabe `WARNING: DATA RACE` trägt.

| Fänger | Mutation | Detektor | Läufe | rot | davon mit DR |
|---|---|---|---|---|---|
| Fall 35, Paket unter `src/hexagon/` (Rennen im Testcode) | Fixture | `-race` | 200 | 200 | 200 |
| Fall 35, Paket unter `src/adapters/` (Rennen im Code, vom Test gerufen) | Fixture | `-race` | 200 | 200 | 200 |
| Fall 35, Paket unter `src/bootstrap/` | Fixture | `-race` | 200 | 200 | 200 |
| Fall 36 (`-test.timeout=1s`, `select {}`) | Fixture | `-race` | 50 | 50 | 50 (alle mit `test timed out`) |
| Fall 35, `-test.count=20` im selben Prozess | Fixture | `-race` | 5 Prozesse | je 1 von 20 Wiederholungen rot | 1 je Prozess (`SPEC-122` Nr. 8 bestätigt) |
| A15 `TestGleichzeitig` | keine | `-race` | 200 | 0 | 0 |
| A15 | a15m1 | `-race` / ohne | 200 / 200 | 200 / 200 | 200 / 0 |
| A15 | a15m2 | `-race` / ohne | 200 / 200 | 200 / 12 | 200 / 0 |
| B19 `TestSequenzGleichzeitig` | keine | `-race` | 100 | 0 | 0 |
| B19 | b19m | `-race` / ohne | 100 / 100 | 100 / 100 | 100 / 0 |
| B8 `TestAnnahmeGegenDrainGleichzeitig` (200 Runden je Prozess) | keine | `-race` | 20 | 0 | 0 |
| B8 | b8a | `-race` / ohne | 30 / 30 | 22 / 2 | 0 / 0 |
| B8 | b8b | `-race` / ohne | 30 / 30 | 0 / 0 | 0 / 0 |

Die Fixtures der Fälle 35 und 36 habe ich nach `zaehleOhneKante`, `rennenImTest`, `rennenImCode` und `rennenUndBlockade` in `tools/gegenproben/faelle.go` nachgeschrieben; Textgleichheit mit den Quellen des Programms ist nicht maschinell geprüft, die Fälle selbst liefen in `make gegenproben` (36 ok). Die Zahl des Implementers (`SPEC-102`: „die Zahl der roten Läufe steht in seinem Bericht“) kenne ich nicht.

---

## Findings

Kein Fix wird vorgeschlagen (Skill: Reviewer kategorisiert, Implementer entscheidet). Kategorie nach `AGENTS.md` §3.15 und §3.16 und nach dem Skill `v6.18.1` · `regelwerk/modul-10-review-harness.md` §Ziel-Form: Reviewer-Skill.

| ID | Kategorie | Ort | Aussage | Ausfallszenario | Quelle | Verifizierbar | Klasse | Adressat |
|---|---|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `Dockerfile:26` · `FROM deps AS pruefung`; `harness/vertraege/SPEC-122-test-race.md:38`; Plan §3 Zeile Z10 | `SPEC-122` Nr. 4 und Z10 sagen zu: `CGO_ENABLED=1` in einer der Stufen `deps`, `source`, `pruefung`, `test`, `build` lässt `make build` und `make test` rot werden (`cgo: C compiler "gcc" not found`). Gemutet ist es so für `deps`, `source`, `test`, `build`; `ENV CGO_ENABLED=1` in der Stufe `pruefung` lässt `make test` (`pruefung test: ok`) und `make build` (`pruefung build: ok`) grün. Z10 nennt als Mutation nur `build` oder `test`, der Vertrag nennt als Messung nur die Images von `build` und `test`. | Wer die Stufe `pruefung` auf cgo stellt, sieht kein Rot, obwohl der Vertrag eines zusagt; das Gate trägt dort keinen Fänger, und die Zusage über „deps bis build“ ist für ein Element der Menge falsch. | `AGENTS.md` §3.15; `SPEC-122` Nr. 4 | ja — Mutation, `make test` und `make build` rc 0 | Allaussage über eine Menge ohne Fänger je Element | Architect (Wortlaut Nr. 4 und Zeile Z10) |
| F-2 | MEDIUM | `tools/gegenproben/faelle.go:385` · `Wegwerfmodul: true,` (Fall 35); `:398` (Fall 36); `:379` · `Env: []string{"CGO_ENABLED=0"}` (Fall 34); Tests `tools/gegenproben/main_test.go:290`, `:299` | `TestWegwerfmodulKopiertDieEingabenNicht` und `TestFallEnvErreichtDasGate` prüfen, dass die Felder an einem synthetischen Fall wirken; dass die Fälle 35, 36 `Wegwerfmodul` und Fall 34 `Env` tragen, fängt kein Test. Gemutet (`Wegwerfmodul: false` in Fall 35, in Fall 36; `Env` aus Fall 34): `make gegenproben` rc 0, 36 Fälle, 0 rot. Fall 34 ohne `Wegwerfmodul` ist rot (Bau-Fehler). | Mit `Wegwerfmodul: false` kopiert der Fall `src/` der Eingaben und legt sein `go.mod` darüber (Fall 34 zeigt es: Bau-Fehler `cannot find module providing package …`); in Fall 35 und 36 stehen die erwarteten Texte trotzdem in der Ausgabe, der Fall bleibt grün und prüft nicht mehr, was `SPEC-102` (Aufbau: Wegwerfmodul statt Kopie) und die Kommentare im Kopf von `raceFaelle` und in der Stufe `gegenproben` zusagen. `Env` in Fall 34 hat in der Stufe keine Wirkung, weil sie `CGO_ENABLED=0` selbst setzt. | `AGENTS.md` §3.15, §3.11; Plan Z11 | ja — drei Mutationen, `make gegenproben` rc 0 | Fänger prüft den Mechanismus, nicht seine Verwendung je Fall | Implementer |
| F-3 | MEDIUM | `tools/pruefung/main.go:181` · `if l.code != 0 {` in `pruefeRace`; `tools/pruefung/main_test.go:290` · `func TestRaceBefund`; `SPEC-122` Nr. 3 (`signal: killed`); Plan §6 Zusatztabelle, Zeile *Weitere Paare* | `TestRaceBefund` führt die Exit-Codes 1 und 2; `starte` liefert für einen durch ein Signal beendeten Prozess `-1` (`ExitError.ExitCode`). Gemutet `l.code > 0` statt `l.code != 0`: alle Tests der Stufe `pruefung` grün. Die Zusatztabelle begründet das fehlende Paar *Ende durch den Speicher* mit Gleichwertigkeit („der Code verzweigt nicht nach der Ursache“); die Mutation zeigt eine Form des Exit-Codes, die kein Test führt. | Wird `go` durch ein Signal beendet (Speichergrenze: Nr. 3 nennt das Ende durch den Speicher als Befund), meldet der Code heute `go test -race:` und Exit 1; nach einer Änderung der Prüfung auf `> 0` meldet `pruefung race` `ok` und Exit 0 für einen Lauf, der nichts geprüft hat, und kein Test wird rot. Dieselbe Form trägt `pruefeTest` (nicht im Diff). | `AGENTS.md` §3.15; `SPEC-122` Nr. 3 | ja — Mutation z3j, Tests rc 0 | Fänger führt nur eine Form der Eingabe (Exit-Code) | Implementer (Fänger); Architect (Gleichwertigkeitsgrund) |
| F-4 | MEDIUM | `src/adapters/driven/httpupstream/client_test.go:45` · `ohne den Detektor fängt der Test es nur mit der Rate` (der Satz läuft in Zeile 46 weiter: `aus §3 des Plans von slice-upstream-adapter`) | Der Satz verbindet zwei Formen: das Datenrennen „ohne dass eine Antwort falsch sein muss“ und die Rate aus §3 des Gebers (200 von 200, Mutation: die Map wird als Header benutzt, jede Antwort ist falsch). Gemessen: die genannte Mutation ist mit und ohne Detektor 200 von 200 rot (ohne ihn über falsche Antworten, keine Meldung); die Form ohne falsche Antwort (eine `http.Header`-Map am Client in jedem `Senden` beschrieben) ist unter dem Detektor 200 von 200 rot, ohne ihn 12 von 200. Die Rate dieser Form steht nirgends. | Wer dem Kommentar glaubt, hält ein geteiltes Feld ohne falsche Antwort ohne Detektor für so sicher gefangen wie die Mutation aus §3 des Gebers (200 von 200); gefangen sind 6 %. Fällt die Stufe `race` aus der Kette oder wird sie übersprungen, bleiben solche Rennen in 94 von 100 Läufen ohne Meldung, und der Kommentar sagt das Gegenteil. | `AGENTS.md` §3.11, §3.10; Skill (MEDIUM *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*) | ja — a15m1, a15m2 mit je 200 Prozessläufen | Kommentar nennt die Rate einer anderen Mutation | Architect (Wortlaut im Plan §3, *Kommentare in Tests*); Implementer (Kommentar) |
| F-5 | LOW | `src/hexagon/services/record/anhaengen_test.go:81` · `Detektor fängt der Test es mit hoher Wahrscheinlichkeit, nicht sicher` | Der Kommentar nennt „hoch“ ohne Zahl; `AGENTS.md` §3.10: „hoch“ steht nur mit Zahl. Gemessen (Sperre der Folge entfernt, ohne Detektor): 100 von 100 Prozessläufen rot, unmutiert 0 von 100. | Der Leser kann „hoch“ nicht prüfen; die Stichprobe (100 von 100) trägt „hoch“, widerlegt „nicht sicher“ nicht. | `AGENTS.md` §3.10 | ja — b19m, 100 Läufe | „Hoch“ ohne Zahl | Architect (Wortlaut im Plan §3); Implementer |
| F-6 | LOW | `tools/gegenproben/faelle.go:18` · `ein Eingriff in die Kopie der Eingaben`; `:23`, `:29`; `tools/gegenproben/main.go:140` · `fahre legt je Fall eine Kopie von wurzel an` | Für die Fälle 34 bis 36 gibt es keine Kopie; `legeVerzeichnisAn` legt ein leeres Verzeichnis an. Die Kommentare von `Fall`, `Neu`, `Argumente` und `fahre` beschreiben nur den Fall mit Kopie. | Ein Leser, der `Fall` oder `fahre` liest, sucht die Kopie für Fall 35 und findet sie nicht; der Satz am Feld `Wegwerfmodul` und an `legeVerzeichnisAn` sagt es. | `AGENTS.md` §3.7 | nein — Lesen | Kommentar beschreibt den Regelfall, nicht den neuen Fall | Implementer |
| I-1 | INFO | Plan §3 *Rate*; `SPEC-102` Fälle 34 bis 36 | Der Bericht des Implementers mit der Zahl der roten Läufe der Fixtures und mit seinen Mutationen lag mir nicht vor; die Zahl in `client_test.go` („200 von 200“) gleicht meiner Messung der genannten Mutation, die Raten der Fälle 35 und 36 sind meine eigenen (200 von 200, 50 von 50). | Ohne den Bericht kann der Verifier die Zahl des Implementers nicht gegen meine halten. | `AGENTS.md` §3.10 | nein | – | Verifier |
| I-2 | INFO | `src/hexagon/services/record/wahl_test.go:416` · `der Detektor von make test-race meldet es nicht` | Der neue Satz stimmt: b8a ist unter `-race` 22 von 30 rot, ohne einen Lauf mit `DATA RACE`; ohne `-race` 2 von 30 (der Plan von `slice-record-service` nennt 13 von 100 und 19 von 100). Unter dem Detektor ist die Rate höher als ohne; der Kommentar nennt keine Rate, `slice-race-gate` nennt keine für B8. Die Mutation b8b (Unlock und Lock zwischen Setzen des Drain und Blick auf die laufenden) ist mit und ohne `-race` 0 von 30 rot: nimmt `nimmAn` nach dem Setzen des Drain nichts mehr an, hat dieses Fenster keine Wirkung; die Form des Plans von `slice-record-service` (zählen, danach in eigener Sperre setzen) habe ich nicht gefahren. | – | `AGENTS.md` §3.10 | ja — b8a, b8b | – | Architect (Seite von `Drain`, `SPEC-121` Nr. 2e), Planner |
| I-3 | INFO | Plan §4 *Größenmarke* | 422 Zeilen nach `git diff --numstat 73e123e..HEAD` über die Pfade von §4 (Marke 800, Schätzung 620); der Slice hat Zeilen, die keine reinen Funktionen sind: die Messung nach dem Review trägt der Planner ein (`AGENTS.md` §3.14). | – | `AGENTS.md` §3.14 | ja — `git diff --numstat` | – | Planner |
| I-4 | INFO | `SPEC-102` Randform 6 · Fallfrist 120 s | Zwei `make gegenproben` zugleich auf dem 20-Kern-Host lassen die neun lint-Fälle an der Frist von 120 s scheitern („Zeit überschritten“); ein Lauf allein braucht 85 s im Ganzen (Fälle 58 s). Die Fälle 34 bis 36 teilen den Prozessorkern-Pool mit ihnen. | – | `SPEC-102` Zeitbudget | ja — zwei Läufe zugleich | – | Architect |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `Dockerfile` (Stufe `race`, Kommentare der Stufen `race`, `pruefung`, `gegenproben`) | geprüft; Befund F-1 (Z10, Stufe `pruefung`); sonst ohne Befund: Pin per Digest gleich dem Vertrag, zwei `RUN` mit `--network=none`, Vergleich der ersten Zeile, Umgebung wie `deps` und `lint`, sieben Läufe zu Z9 mit den Meldungen des Plans |
| `harness/mk/go-gates.mk`, `harness/mk/gegenproben.mk` | geprüft, ohne Befund (`test-race` in `GATE_CHECKS` hinter `test`, Zeile in `make help`, `RACE_IMAGE ?=` mit Digest, Kommentare beschreiben den Stand) |
| `harness/README.md` (Zeile `make test-race` in §Sensors, *Nicht behauptet*, Zeile `make gegenproben`) | geprüft, ohne Befund (Bindung auf [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md); der Satz zur Verdrahtung gleicht N1 und dem Wortlaut des Plans) |
| `tools/pruefung/main.go`, `main_test.go` | geprüft; Befund F-3; sonst ohne Befund (34 Mutationen zu Z1 bis Z4 rot aus dem genannten Grund) |
| `tools/gegenproben/main.go`, `faelle.go`, `main_test.go` | geprüft; Befunde F-2, F-6; sonst ohne Befund (Fälle 34 bis 36 gleichen `SPEC-102` in Name, Argumenten, Fixture, Erwartung; Fixtures ohne Zeichen über ASCII, `TestRaceFaelleFuehrenNurASCII` rot an z11j; Raten der Fixtures 200 von 200) |
| `src/` (`client_test.go`, `anhaengen_test.go`, `wahl_test.go`) | geprüft; Befunde F-4, F-5; sonst ohne Befund (nur Kommentarzeilen im Diff, kein Bestandscode verändert; B8 stimmt mit b8a) |
| `SPEC-122`, `SPEC-102` (Fälle 34 bis 36, Aufbau, Zeitbudget), `SPEC-036` | geprüft; Befund F-1 (Nr. 4); sonst Code und Vertrag gleich (Nr. 1 bis 3, 5, 6, 8, 9, 10; Zeitbudget 85 s im unmutierten Lauf) |
| Plan `slice-race-gate` §1, §3, §4, §6 (§3.9) | geprüft, ohne Befund (Kopf, Dateitabelle, Tabelle, Raster, Marke, Ausgänge und Messungen stimmen mit Code und Läufen überein; I-3 für die Messung nach dem Review) |
| Hard Rules `AGENTS.md` §3.1, §3.2, §3.13 | geprüft, ohne Befund (kein `//nolint` im Diff, Docker-only, keine Zuweisung an einen anderen Slice im Diff) |
| `harness/gegenproben.stempel` | geprüft, ohne Befund (`make gegenproben` in der unveränderten Kopie: „Eingaben unverändert“, Prüfsumme `e8c2cd40…`) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 4 |
| LOW | 2 |
| INFO | 4 |

**Finding-Klassen dieses Laufs:** Allaussage über eine Menge ohne Fänger je Element · Fänger prüft den Mechanismus, nicht seine Verwendung je Fall · Fänger führt nur eine Form der Eingabe (Exit-Code) · Kommentar nennt die Rate einer anderen Mutation · „Hoch“ ohne Zahl · Kommentar beschreibt den Regelfall, nicht den neuen Fall

Für die Closure §7: F-1 und F-2 belegen `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (Zusage ohne Fänger, beide über ein Element einer Menge), F-3 dasselbe Muster in der Form „beide Seiten jeder Grenze“; F-4 und F-5 gehören zu `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (die Rate steht nicht oder gehört zu einer anderen Mutation); die Zählung folgt aus den Dateien unter `observations/`.

## Verdikt

**Merge-blockierend:** ja — vier MEDIUM (F-1 bis F-4). Kein HIGH; der Konflikt-Pfad über den Architect ist nicht ausgelöst. Der Code ist nach allen Läufen richtig: `make test-race` ist zweimal erzwungen grün, jede im Plan genannte Mutation der Zeilen Z1 bis Z9 und der Fälle 34 bis 36 ist rot, `make gates` ist grün. F-1 bis F-3 sind Lücken von Zusage und Fänger, F-4 und F-5 Sätze eines Kommentars, den der Plan im Wortlaut vorgab, F-6 ein ungenauer Kommentar.

**Übergabe:** Architect: F-1 (Wortlaut `SPEC-122` Nr. 4 und Z10), F-4 und F-5 (Wortlaut §3 des Plans), die Gleichwertigkeit zu F-3, I-2 (Seite von `Drain`), I-4. Implementer: F-2, F-3, F-4 (Kommentar nach dem Wortlaut des Architect), F-5, F-6. Planner: I-3 (Messung nach dem Review), die Beobachtungen für die Closure §7. Verifier: I-1. Die Finding-Klassen gehen in die Slice-Closure §7. Dieser Report ersetzt keine Verifikation gegen die DoD (Verifier).
