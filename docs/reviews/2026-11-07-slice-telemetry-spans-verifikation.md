# Verifikation slice-telemetry-spans (Spans, Verteilung, Ring und Span-Zeile des Adapters von `Telemetry`)

**Rolle:** Verifier (`Bauen wir es richtig?`, gegen Plan, DoD und Spezifikation). **Datum:** 2026-11-07.
**Gegenstand:** Zweig `slice-telemetry-spans`, HEAD `24a66324` (Fix-Runde nach dem Review `c793026e`), Basis `13e77e0`, Abzweig von `main` bei `eea875b`.
**Eingang:** DoD-Bestätigung und Bericht des Implementers in §3 des Plans (Tabelle S01 bis S24, Mengen M1 bis M15, Absatz *Zusagen, Mutationen und Raten*, Zweig-Menge Z66 bis Z91, Runden R1 bis R8, Fix-Runde F1 bis F6), Review `docs/reviews/2026-11-07-slice-telemetry-spans-review.md` (F1 bis F6).
**Bezug:** `LH-QA-05`, `LH-QA-04`, `LH-QA-01`, `LH-FA-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).
**Spezifikation:** `SPEC-230` Nr. 5 bis 9, `SPEC-231` Nr. 6, `SPEC-232` Nr. 1 bis 6, `SPEC-128` Nr. 9.

## Urteil

**Bestanden mit Auflagen.** Kein HIGH, kein MEDIUM und kein Widerspruch des Produktcodes zum Vertrag: `span.go` (147 Zeilen) und die Änderungen in `telemetry.go` stimmen Satz für Satz mit `SPEC-230` Nr. 5 bis 9, `SPEC-231` Nr. 6 und `SPEC-232` Nr. 1 bis 5 überein. `make gates` ist grün, das Binary blieb unverändert, die Fixes F1 bis F6 decken ihre Klasse: 32 eigene Mutationen an Elementen, die Plan und Fix-Runde nicht nannten, sind alle rot (null grün). Gefunden habe ich nur Nachzüge im Plan (§3.9) und die zwei noch offenen Abweichungen, die der Architect bestätigen muss: vier LOW (V-1 bis V-4) und eine Register-Meldung (V-5). Die Auflagen stehen am Ende.

## Eigene Sensoren (nicht die Behauptung, die Belege)

| Sensor | Ergebnis |
|---|---|
| `make gates` (einmal, eigenes Log `scratchpad/ver-spans/gates.log`, unter `flock /tmp/odata-recorder-heavy.lock`) | Exit 0. In der Kette: `a-check` (0 Befunde), `d-check` (821 Dateien, 0 Befunde), `lint`, `test`, `test-race`, `build`, `handbuch-check`, `katalog-check`, `schema-check`, `migration-check`, `zeichen-check` (1661 Dateien), `baseline-verify`; `groessenmarke slice-telemetry-spans.md: ok, Messung 2203 von Marke 3100`. **Einschränkung:** die Schichten `test` und `race` des Dockerfile kamen aus dem Layer-Cache (`CACHED`), weil der Quellstand der Fassung des Implementers gleicht; ein `CACHED` Schritt ist ein früherer grüner Lauf desselben Inhalts, kein neuer. Deshalb habe ich das Paket selbst gefahren (nächste Zeilen) |
| Paket `src/adapters/driven/telemetry` selbst gefahren (Image `odata-recorder:race`, Baumkopie per `docker cp`, `--network none --memory=4g --cpus=4`, Last des Hosts 8 bis 11 von 4 CPUs) | ohne Detektor **14,0 s**, mit Detektor **73,9 s**, mit Detektor und Profil **90,6 s**, alle `ok` |
| Abdeckung selbst gemessen (`go test -race -count=1 -covermode=atomic -coverprofile`) | **99,2 Prozent** der Anweisungen, wie im Plan; in `span.go` und `telemetry.go` **kein Block mit Treffer 0**; allein `keinSpan.Set` und `End` stehen mit 0,0 Prozent im Funktionsbericht (kein Block, keine Anweisung) |
| Zweig-Runden R1 bis R8 selbst gefahren (`-race -count=1 -run '^<Test>$' -coverprofile`, ein Lauf je Runde, alle `ok`) | Treffer wie im Plan: R1 Z70 32, Z71 16; R2 Z66 2, Z67 23; R5 Z77 20, Z78 5; R6 Z80 24; R7 Z82 133134, Z83 10; R8 Z87 40, Z88 50, Z89 25. Abweichung: R2 Z68 hat **13** statt 11 Treffer, weil die Fix-Runde F5 zwei Geschwister unter einem Eltern-Span ergänzt hat (Plan: Zahlen der Fassung zum Handoff); jeder Block der Zweige hat in seiner Runde einen Treffer |
| Zusatzzeit (Plan: 72,0 s mit Detektor, Review: 95,4 s) | **kein Widerspruch.** Die 72,0 s des Plans sind der Lauf ohne Profil (meine Messung 73,9 s); die 95,4 s des Reviews sind der Lauf mit `-coverprofile` bei Last 12 bis 16 (meine Messung 90,6 s). Der Absatz im Plan nennt das Profil nicht und stammt von vor F2 und F3 (400 Spans, Uhr mit 40000 Zeiten); die Zahl hält trotzdem (V-4) |
| S11 unter Last (`TestSpanOhneEnde -count=20`, zwölf Rechenschleifen im Container neben dem Test, 4 CPUs) | ohne Detektor 20 Läufe in 0,8 s, mit Detektor 20 Läufe in 2,3 s, alle grün, kein Hängen; die Warteschleife (bis 5 s) ist nie in die Nähe ihrer Grenze gekommen |
| Größenmarke | 2203 von Marke 3100 (Summe aus `git diff --numstat 13e77e0..HEAD` über den Adapter; darin 2 Zeilen `regeln_test.go`, die von `main` über den Merge kamen, ohne sie 2201), Auslastung **0,71**: innerhalb 0,60 bis 1,00, keine Meldung an den Architect nach `SPEC-128` Nr. 9 |
| Pfade | gegen `main` (`eea875b`) berührt der Zweig allein `src/adapters/driven/telemetry/`, den Plan, den Review-Report und `spec/spezifikation.md` (10 Zeilen, der Vertrag des Architect vor dem Code). `src/bootstrap/`, `docs/user/`, `harness/`, `.golangci.yml`, `.a-check.yml`, `Makefile`, `Dockerfile`, `tools/` unverändert; `zeile.go`, `handler.go`, `regeln.go`, `export_test.go` unverändert (`git diff 13e77e0..HEAD`); `regeln_test.go` trägt eine Zeile von `main` (`ODR-E6008`, Merge von `slice-mock-lesen`), keine des Slice. Das Binary blieb unverändert, der Handbuch-Punkt der DoD entfällt wie in §3 und §7 geplant |
| `//nolint`, Gate-Lockerung | keine Direktive in `span*.go`, `telemetry*.go`; `.golangci.yml` und `.a-check.yml` ohne Diff |
| Rohbytes (`AGENTS.md` §3.18): `LC_ALL=C grep -P` auf BOM, NBSP, Zeilentrenner, Nullbreiten- und Steuerzeichen, Backslash-u-Folge | kein Treffer in `span.go`, `span*_test.go`, `telemetry.go`, `telemetry_test.go` und im Plan; die Nicht-ASCII-Bytes sind Umlaute, `ß`, `§` und `×` in Kommentaren (Zählung der Zeichen: ä 195, ü 153, ö 13, Ä 6, ß 6, Ü 1, × 28, § 2). Keine Datei trägt CR. Dieser Report ist gegengelesen (siehe Ende) |

## Plan gegen Code

- Dateien wie in §3: `span.go` neu (Typ `span`, `StartSpan`, `naechsteID`, `Set`, `End`, `aufzeichnen`), `telemetry.go` (Felder `pflicht`, `grenzen`, `idMu`, `spanZaehler`, `ring`, `ringNaechster`; der wirkungslose `StartSpan` entfällt; `Snapshot` kopiert den Ring, den ältesten voran, unter der einen Sperre), fünf neue Testdateien, in `telemetry_test.go` entfällt der Teil von `TestKernSchreibtNichts`. Keine Datei über die Liste in Antwort K hinaus.
- Sätze gegen Code, ohne Befund: fremder Name endet in `StartSpan` vor jeder Prüfung (`SPEC-230` Nr. 5, V-8) und liest die Uhr nicht; `Set` und `End` nach dem Ende zählen `zustand` einmal je Aufruf; die Markierung des ersten `End` steht vor der Ablesung der Uhr; `ParentID` allein bei `p.a == a`; ein Eltern-Span, der endete, bleibt Eltern (der Zeiger hängt im `ctx`, nicht an einer Liste); Dauer `max(ende - start, 0)`; Eimer „erster mit Dauer <= Grenze“, zwölfter für jede längere; `Attrs` in der Folge der Tabelle (die Schleife läuft über `s.pflicht`, nicht über die Abbildung), nie nil; `pflicht` zählt `len(pflicht) - len(attrs)` im selben Abschnitt wie Verteilung und Ring; die Span-Zeile geht über `zeileSchreiben` (Felder `ereignis`, `span_id`, `parent_id` nur mit Eltern, `duration_ms` mit drei Nachkommastellen, dann die Attribute), `timestamp` ist die Ablesung des Endes, `MinLevel` wird mit `slog.LevelInfo < minLevel` verglichen (Gleichheit schreibt), Aufzeichnung geht der Zeile voraus, die Sperren liegen nie ineinander und keine über `Clock.Now()` oder `Write`; `ID` `spn-` und 16 Ziffern aus einem Zähler je Adapter, der bei 1 beginnt; der Adapter hält keinen offenen Span.
- Abweichungen aus §6, beide noch **offen** mit Adresse Architect: (1) S11 mit `runtime.SetFinalizer` statt `weak.Pointer` (Begründung trägt: `testpackage` nimmt allein `export_test.go` aus, die Datei gehört `slice-mock-lesen`, eine Änderung des Profils wäre eine Lockerung nach `AGENTS.md` §3.6); der Fänger ist nicht flatternd (siehe oben) und seine Mutation ist rot (eigene `s11-timer-haelt-span`). (2) Zähler der `ID`s als `uint64` unter `idMu`: `SPEC-232` Nr. 5 e („paarweise verschieden“) hält, die Sperre liegt über keinem anderen Aufruf. Ich sehe an beiden keinen Widerspruch zum Vertrag; die Bestätigung ist Sache des Architect (Auflage A-2).

## Tabelle Zusage · Fänger gegen den Vertrag

Jeder Satz von `SPEC-230` Nr. 5 bis 9, `SPEC-231` Nr. 6 und `SPEC-232` Nr. 1 bis 5 hat eine Zeile S01 bis S24 oder ein benanntes Negativ (N1 bis N3); ich habe die Sätze einzeln gegen die Nummern gelegt und keinen ohne Fänger gefunden. Die 24 Tests existieren unter den Namen der Tabelle (`TestSpanNamen` bis `TestSpanZeileSperre`, dazu `TestRingKopie`, `TestSpanZeileGesamtspan` und `TestAttributFolge`), die Zählprüfungen stehen gegen die Quelle (M3: 8 Namen und Disjunktheit gegen `support.Spans()`, M2: `zulaessig` 11, `fremd` 149, Prüfungen `2 * (160 + 149)` = 618, M4: Teilmengen aus `1 << len(Pflicht)`, F1: `fak` aus der Länge von `Pflicht`).

### Eigene Mutationen (32 distinkte Mutationen, `tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, `--memory=4g`, unter `flock`, Last 8 bis 18 von 4 CPUs)

Liste 1 (21 Einträge, Klasse *andere*): 19 rot, 1 `rot-paket` (`u-set-nach-end-liest-uhr`, s. V-3), 1 Übersetzungsfehler (meine Mutation hatte `angenommen` ungenutzt; in Liste 2 korrigiert wiederholt, rot). Liste 2 (6, *andere*): alle rot. Liste 3 (4, *nebenläufig*, je 8 Prozesse mit und 8 ohne Detektor, `-test.count=1`): alle **8 von 8 und 8 von 8** rot. Dazu `s11-timer-haelt-span` und `v-folge-teilmenge-tauscht`, beide rot. **Grün: keine.**

| Klasse (Befund) | Mutation an einem nicht genannten Element | Ergebnis |
|---|---|---|
| F1 Setzfolge (Menge: 14 Anordnungen × 3 Wege = 42) | Tausch von Attribut 2 und 3 bei einem Span mit drei; Tausch bei zwei Attributen; Tausch allein in der Zeile bei zwei Attributen; Tausch bei einer Teilmenge von zwei aus drei | alle rot (`TestAttributFolge`; die Teilmenge `TestPflichtAttribute`) |
| Uhr (Now() je Span zweimal) | `StartSpan` liest zweimal; `Set` vor dem Ende liest; `Set` nach dem Ende liest; Start unter `idMu`; Ende unter `zeilenMu` | rot (`TestUhrAblesungen`; `Set` nach dem Ende `TestEndDoppelt`, V-3; die beiden Sperren-Mutationen `TestUhrRuftAdapter`, 8 von 8 und 8 von 8) |
| MinLevel | `LevelInfo+1 < minLevel`; `LevelInfo-1 < minLevel` | rot (`TestSpanModi`) |
| Eimer | Grenze `+1 ns`; Eimer 6 übersprungen; zwölfter Eimer fällt in den elften | alle rot (`TestEimer`) |
| Verteilung, Sum, Count, Counts | `Sum` ohne `odata.mock.query`; `Count` ohne `odata.record.capture`; `Count` in eigenem Abschnitt vor `Counts` | rot (`TestVerteilung`; die letzte `TestEndSnapshotGleichzeitig`, 8 von 8 und 8 von 8) |
| Stand (F3, Kette) | `Durations` in eigenem Abschnitt vor Ring und Rest | rot (`TestSnapshotEinStand`, 8 von 8 und 8 von 8) |
| Ring | Umlauf in Schritt zwei; Standard 1023; `Start` des Eintrags gleich Ende | rot (`TestRing`) |
| parent_id | Eltern ohne `odata.replay.match`; Eltern ohne `odata.workspace.mutation` | rot (`TestElternKind`) |
| `ID` | `%017d` statt `%016d` | rot (`TestSpanIDs`) |
| Canary, Aufrufstelle `StartSpan` (der Befund nannte `Set`) | `StartSpan` legt die rohen Attribute ab | rot (`TestCanarySpans`) |
| Aufrufstellen | fremder Name liefert `Background()`; nil-`ctx` bleibt nil; `StartSpan` und `Set` zählen die Verwürfe nicht | alle rot |
| Zustand | zweites `End` zählt nicht; `Set` nach dem Ende zählt nur mit Attribut | rot (`TestEndDoppelt`) |
| S11 | ein Zeitgeber hält den Span | rot (`TestSpanOhneEnde`, 6 s, die Warteschleife läuft aus) |

**Fixes gegen die Klasse (§3.17):**

- **F1 (Reihenfolge der Attribute):** die Menge „Setzfolge“ besteht aus den Anordnungen der Pflicht-Attribute je Span (3! + 4 × 2! = 14, 42 Läufe mit den drei Wegen), der Fänger ist eine Schleife darüber mit Eintrag, Schlüsselfolge der Zeile und Lage in den Rohbytes. Teilmengen (zwei von drei) in fremder Setzfolge sind nicht in der Menge; sie laufen in `TestPflichtAttribute` in der Folge der Tabelle. Die Mutation `v-folge-teilmenge-tauscht` ist dort rot, eine Folge-des-Setzens-Mutation allein bei Teilmengen wäre es auch (die Abbildung `werte` hat keine Setzfolge; jede Mutation, die sie einführt, ist an der vollen Menge sichtbar). Keine Lücke.
- **F2 (Fenster Prüfen/Markieren):** die Menge ist jedes Paar aus 16 `End` und 8 `Set` an 400 Spans; die Masse ohne Detektor ist flatternd, der Detektor je Prozess und der Haltepunkt hinter der Uhr tragen die Zusage. Die Rate 34 von 50 und 16 von 50 ohne Detektor habe ich nicht neu gemessen (akzeptiertes Negativ des Plans, Grund und Folge stehen in §3; der Grund trägt: kein Argument, an dem ein Test halten könnte). Ich fand keine weitere Stelle in `End` oder `Set`, an der das Fenster zwischen Prüfen und Handeln liegt: eine Mutation „Attribute nach dem Unlock lesen“ ist gleichwertig (nach der Markierung schreibt kein `Set` mehr in `werte`) und deshalb nicht gebildet.
- **F3 (`Sum` im Orakel):** das Orakel `verteilungKaputt` vergleicht `Count`, `Counts`, Einträge je Name und `Sum` gegen die Dauern im Ring über die 5 Namen je Stand; die Uhr der Masse läuft vorwärts, die Dauern sind ungleich 0. Meine Mutationen `Count` in eigenem Abschnitt und `Durations` zuerst sind je 8 von 8 mit und ohne Detektor rot; eine nicht genannte Stelle der Kette (`Durations` gegen `Spans`) ist damit gedeckt.
- **F5 (Geschwister):** zwei Geschwister mit verschiedenen Attributen unter einem Eltern-Span; die Mutation `f5-geschwister-teilen-werte` ist laut Plan rot, die Menge „Geschwister“ ist zwei Elemente (ein Eltern-Span, zwei Kinder gleichen Namens); `p-eltern-*` an nicht genannten Namen ist rot.
- **F4, F6:** Plan 618 stimmt mit dem Test (`2 * (160 + 149)`); die gleichwertige Mutation `f6-return-nach-append-fehlt` ist gleichwertig: ohne das `return` zeigt `ringNaechster` bis zur Grenze auf den eben angehängten Platz (nachgerechnet für `MaxSpans` 1 und 3), der Schreibvorgang ist derselbe.

**S18 (Sperre über `Now()`), S23 (1), S24:** Haltepunkt, färben in jedem Lauf; meine zwei Sperren-Mutationen (`idMu`, `zeilenMu` über der Ablesung) sind 8 von 8 und 8 von 8 rot, unabhängig von den drei Mutationen des Plans.

## Befunde

| Nr. | Stufe | Befund | Auflage |
|---|---|---|---|
| V-1 | LOW | Die Zeile S23 der Tabelle (Plan, Spalte *Fänger*) nennt noch „50 Spans“; Test, Menge M12 und Raster nennen 400 (F2). Die Zeilen *S23 (Fix F2)* im Mutationsblock tragen es richtig. `AGENTS.md` §3.9: dieselbe Aussage an allen Stellen | A-1 |
| V-2 | LOW | `TestAttributFolge` (F1) steht allein im Absatz *Fix-Runde*; die Zeilen S12 und S21 der Tabelle nennen ihn nicht als Fänger, und die Mutationen `f1-*`, `f2-*`, `f3-*`, `f5-*` zählen nicht in die „65 Mutationen“ der ersten Tabelle (die 65 stimmen: 65 verschiedene Namen im Block). Die Tabelle *Zusage · Fänger* ist die Quelle des Reviewers; ein Fänger, den nur der Absatz kennt, wird bei der nächsten Änderung übersehen | A-1 |
| V-3 | LOW | `u-set-nach-end-liest-uhr` (`Set` nach dem Ende liest die Uhr) wird in `TestEndDoppelt` rot (S09, Ablesungen 3 statt 2), **nicht** in `TestUhrAblesungen`; M15 führt `Set` allein vor dem Ende. Die Zusage ist gedeckt (S09), die Zeile S20 und die Menge M15 nennen die Aufrufstelle „`Set` nach dem Ende“ nicht, obwohl S20 „nie von `Set`“ sagt. Keine Lücke im Betrieb | A-1 (S20 verweist auf S09) |
| V-4 | LOW | Der Absatz *Zusatzzeit* ist von vor F2 und F3 und nennt das Profil nicht. Meine Messung nach den Fixes (Last 8 bis 11): +6,7 s ohne und +35,8 s mit Detektor gegenüber 7,3 s und 38,1 s vor dem Slice (14,0 s, 73,9 s), also nahe den +7,1 s und +33,9 s des Plans; der Review maß mit Profil. Für §7 gilt: 14,0 s und 73,9 s, mit Profil 90,6 s | A-4 |
| V-5 | Register | `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`: die Fix-Runde (Commit `24a66324`) hat `permutationen` (`span_hilfen_test.go`, 15 Zeilen) und `TestAttributFolge` (`span_zeile_test.go`, 66 Zeilen) laut Angabe des Aufrufers per Heredoc (`cat >>`) angehängt, gegen die Betriebsregel „auf Repo-Dateien nur `Edit` und `Write`“ in §3. **Der Inhalt stimmt:** beide Dateien enden mit genau einem `\n`, ohne CR, BOM oder Steuerzeichen; `gofmt`, `go vet` (Stufe `test` des Gate-Laufs), `lint` und `go test` sind grün; die Funktionen tun, was der Kommentar sagt (Tests oben). Die Verletzung hat keine Wirkung auf die Dateien, ist aber eine weitere Zählung der Beobachtung | A-3 |

## DoD

| Punkt | Stand |
|---|---|
| Spans (`SPEC-230` Nr. 5 bis 7) | erfüllt: S01 bis S06, S09 bis S11, S14, S19 bis S20, S23 mit Fängern, 32 eigene Mutationen rot |
| Verteilung, Ring, Span-Zeile, Stand (`SPEC-230` Nr. 8 und 9, `SPEC-231` Nr. 6, `SPEC-232` Nr. 5 b und e) | erfüllt: S07, S08, S12, S13, S15 bis S18, S21, S24; Gate grün; Raten: Implementer 50 von 50 und 50 von 50 (S15, S16), Verifier 8 von 8 und 8 von 8 an nicht genannten Mutationen |
| Keine Geheimnisse und kein Ausfall (`SPEC-232` Nr. 2 und 3) | erfüllt: S22 mit 980 Prüfungen, eigene Mutation an `StartSpan` rot; S13 und S14 |
| `make gates` grün | erfüllt (Einschränkung zu `test` und `race` oben, die ich durch eigene Paket-Läufe ersetzt habe) |
| Review durchgeführt | erfüllt: `docs/reviews/2026-11-07-slice-telemetry-spans-review.md` |
| Doku-Update, Handbuch | Spezifikation vom Architect vor dem Code; Binary unverändert, Handbuch entfällt |
| Closure-Notiz, Register, Ausgänge in §6, Paarungen | offen, Sache des Planners bei der Closure (A-3, A-4, A-5) |

## Auflagen

- **A-1 (Planner, vor der Closure, `AGENTS.md` §3.9):** die Zeile S23 auf 400 Spans ziehen (V-1); in die Spalte *Fänger* von S12 und S21 `TestAttributFolge` (F1) aufnehmen und die Mutationen der Fix-Runde in den Block *Zusage · Mutation · roter Test* stellen oder die Zählung „65“ mit „und N der Fix-Runde“ benennen (V-2); die Zeile S20 verweist für „`Set` nach dem Ende liest die Uhr nicht“ auf S09 (V-3).
- **A-2 (Architect):** die zwei Abweichungen in §6 bestätigen (S11 mit `SetFinalizer`, Zähler der `ID`s unter `idMu`); ich trage beide mit der Begründung des Implementers und meiner Messung.
- **A-3 (Planner, Register):** `betriebsregel-fuer-repo-dateien-verletzt` bekommt einen Beleg durch `evidence/slice-telemetry-spans.md` (Zähler 9× auf 10×; dieser Zweig schreibt das Register nicht): Fix-Runde, Heredoc `cat >>` für `permutationen` und `TestAttributFolge`, Inhalt geprüft und in Ordnung. Dieselbe Closure trägt die offenen Punkte aus §6 (Hybrid-Span, Auftrag an `slice-lasttest`, Pfade neben den Zweigen, Satz zu Reports) und die Paarungen.
- **A-4 (Planner, Closure):** Messpunkt-Zeile 2203 von 3100 (0,71), Zusatzzeit 14,0 s und 73,9 s (V-4), Raten von S15 und S16 aus dem Plan (50 von 50) mit dem Messenden Implementer, und dass das Binary unverändert blieb; die Entscheidung über den Re-Evaluierungs-Trigger von [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) bleibt beim User.
- **A-5 (Integrationsschritt):** der Merge nach `main` zieht `eea875b` nach; `regeln_test.go` liegt auf beiden Seiten gleich (die Zeile stammt von `main`), kein Konflikt zu erwarten.

## Gegengelesen

Dieser Report ist nach dem Schreiben mit `LC_ALL=C grep -P` auf Backslash-u-Folge, BOM, NBSP, Zeilentrenner und Steuerzeichen gegengelesen; Kennungen von ADRs stehen als auflösende Links.
