# Review-Report: slice-telemetry-spans — 2026-11-07

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-telemetry-spans`, HEAD `806b67be`, Basis `13e77e0`; `span.go`, `telemetry.go`, `telemetry_test.go` und die fünf Testdateien `span_*_test.go` im Adapter `src/adapters/driven/telemetry`, Plan §3, §4, §6.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-07

**Eingangs-Kontext:**

- `docs/plan/planning/done/slice-telemetry-spans.md` (im Zweig): §3 Antworten A bis L, Tabelle S01 bis S24, Mengen M1 bis M15, Absatz *Zusagen, Mutationen und Raten*, Runden R1 bis R8, Zweig-Menge Z66 bis Z91, Fenster der Sperre, §6 (zwei Abweichungen mit Ausgang *offen*)
- `SPEC-230` Nr. 5 bis 9, `SPEC-231` Nr. 6, `SPEC-232` Nr. 1 bis 5
- `AGENTS.md` §3.7, §3.9 bis §3.18
- `LH-QA-05`, `LH-QA-04`, `LH-QA-01`, `LH-FA-01`
- ADR: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) (Profil `testpackage`, nur genannt), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Trigger Zeit, nur genannt)

**Unverändert (geprüft mit `git diff 13e77e0..HEAD`):** `zeile.go`, `handler.go`, `regeln.go`, `export_test.go`, `hilfen_test.go`, `zeile_test.go`, `handler_*_test.go`, `.golangci.yml`, `.a-check.yml`. Keine Direktive `//nolint` in den neuen Dateien. Die Änderungen an `spec/` und `harness/` zwischen `13e77e0` und HEAD stammen aus dem Merge von `main`, nicht vom Implementer.
Größenmarke: Summe aus Einfügungen und Löschungen über `src/adapters/driven/telemetry` seit `13e77e0` 2098 gegen Marke 3100, wie berichtet (Produktcode `span.go` 147 Zeilen).

## Eigene Mutationen

Werkzeug `tools/harness/mutationslauf.sh` (Container ohne Netz und Mount, `--memory=4g`, unter `flock`), Host-Last 9 bis 19 von 4 CPUs, andere Zweige liefen zugleich. Eigene Liste, andere Namen als die des Implementers; 46 Mutationen (39 der Klasse `andere`, 7 `nebenläufig`).

**Klasse `andere`, ein Prozess je Mutation, 39 Mutationen, 38 rot** (eine davon `rot-paket`, siehe `Attrs` nie nil; die grüne, `g08`, ist gleichwertig):

| Zusage | Mutation | Ergebnis |
|---|---|---|
| End schreibt über `zeileSchreiben`, nicht `eintrag` | `End` ruft `eintrag` | rot (`TestSpanZeileInhalt`) |
| `MinLevel`: der Wert wird verglichen, die Gleichheit schreibt | `<=`; umgekehrter Vergleich; Prüfung vor der Aufzeichnung | je rot (`TestSpanModi`) |
| Uhr genau zweimal je Span | zweite Ablesung in `End`; zweite in `StartSpan` | je rot (`TestUhrAblesungen`) |
| Folge der Felder (`SPEC-231` Nr. 6) | `ereignis` hinter `span_id`; `duration_ms` vor `parent_id`; `parent_id` immer; `timestamp` ist der Start | je rot (`TestSpanZeileInhalt`) |
| Ende vor Anfang gibt Dauer 0; Eimer | ohne `max`; Vorzeichen vertauscht; `<` statt `<=`; zwölfter Eimer ist der elfte | je rot (`TestEimer`, `TestVerteilung`) |
| Ring: ältester zuerst, Umlauf, Kopie | Umlauf auf 0; Verdrängen an falscher Stelle; Ordnung vertauscht; Attribute rückwärts; Start gleich Ende; Stand ohne `Kopie()` | je rot (`TestRing`, `TestRingKopie`) |
| `Attrs` nie nil | `var attrs` statt `make` | `rot-paket`: `TestRing` bleibt grün, `TestRingKopie` färbt das Paket |
| `pflicht`, `schluessel`, `wert` | `pflicht` höchstens 1; `pflicht` nicht gezählt; `StartSpan` ohne Verwurf; `Set` ohne Verwurf; Regeln von `Log` in `StartSpan`; Regeln von `Log` in `Set`; erster Wert gilt in `Set`; Attribute von `StartSpan` entfallen | je rot |
| fremder Name: ein `spanname`, sonst nichts | zusätzlich `zustand`; zusätzlich `pflicht` | je rot (`TestSpanNamen`) |
| `Count`, `Durations` | `Count` ohne Zuwachs; Verteilung nicht zurückgeschrieben | je rot (`TestVerteilung`) |
| `parent_id` | Gleichheit des Adapters entfällt | rot (`TestElternKind`) |
| zweites `End`, `Set` nach `End` | Zählung von `zustand` entfällt (`End`; `Set`) | je rot (`TestEndDoppelt`) |
| `ID` hat 16 Ziffern | 15 Ziffern | rot (`TestSpanIDs`) |
| Ring füllt sich | `return` nach dem Anhängen entfällt | **grün, gleichwertig**: Zeiger und Länge laufen gleich, bis der Ring voll ist; der Umlauf schreibt dann dieselbe Stelle |

**Klasse `nebenläufig`, `-test.count=1` je Prozess, mit Detektor / ohne Detektor:**

| Fenster | Mutation | Ergebnis | Dauer, Last |
|---|---|---|---|
| Aufzeichnung (S15) | `End` gibt die Sperre zwischen Verteilung und Ring frei | 30/30 und 30/30 | 380 s, 15 von 4 |
| Kopie (S16) | `Snapshot` kopiert in zwei Abschnitten | 30/30 und 30/30 | 228 s, 19 von 4 |
| Verteilung (S16) | `Count` und `Counts` in zwei Abschnitten | 30/30 und 30/30 | 387 s, 16 von 4 |
| `ID` (S17) | Zähler liest und schreibt in zwei Abschnitten unter `idMu` | 50/50 und 50/50 | 7 s, 15 von 4 |
| Markieren (S23 (2)) | **Prüfen und Markieren in zwei Abschnitten** | **50/50 und 26/50**, `flatternd` | 3 s, 15 von 4 |
| Verteilung (S16) | **`Sum` in eigenem Abschnitt vor der Verteilung** | **0/30 und 0/30**, grün | 456 s, 12 von 4 |

Die siebte Mutation (`n06`, `Sum` hinter der Verteilung, vor dem Ring) ist 30/30 und 30/30 rot, aber aus dem Grund der ersten Zeile: sie öffnet auch das Fenster vor dem Ring. Sie zählt nicht als Beleg für `Sum`.
Eigene Raten gelten für 30 Prozesse je Arm bei S15 und S16 (der Auftrag nennt 50); die 50/50 des Implementers sind nicht neu gemessen.

**Handprobe in einer Baumkopie (nicht im Zweig):** Die Attribute eines Eintrags und der Zeile in der Folge des Setzens statt der Tabelle (ein Feld `reihe`, in `StartSpan` und `Set` gefüllt, in `End` durchlaufen): `go test -count=1 ./src/adapters/driven/telemetry/` **grün** (16,1 s).

**Abdeckung unter dem Detektor:** `go test -race -count=1 -coverprofile` über das Paket in der Stufe `race`, 95,4 s bei Last 12 bis 16: 99,2 Prozent der Anweisungen, jede Funktion von `span.go` und die geänderten von `telemetry.go` 100,0 Prozent, kein Block mit Treffer 0 in beiden Dateien (`keinSpan.Set` und `End` tragen keine Anweisung und stehen mit 0,0 Prozent im Bericht von `go tool cover`). `TestSpanOhneEnde` läuft 0,02 s (kein Warten auf den Finalizer), `TestEndSnapshotGleichzeitig` 18,0 s, `TestRing` 4,7 s.

## Findings

### F1 · MEDIUM

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (ein Fänger führt jede Form der Eingabe, die der Code verschieden behandelt; eine Fixture gleicht nie dem Wert der Mutation)
- `pfad`: `src/adapters/driven/telemetry/span_zeile_test.go` · `werteVon(z.pflicht, gueltigSpan)`; ebenso `span_test.go` · `attrs := werteVon(keys, gueltigSpan)`
- `befund`: `SPEC-230` Nr. 7 und `SPEC-231` Nr. 6 sagen zu: Attribute in der Reihenfolge der Tabelle, nicht in der Folge des Setzens (Plan S12 und S21 nennen die Mutation). Jeder Fall setzt die Pflicht-Attribute in der Reihenfolge der Tabelle (`werteVon(z.pflicht, …)`, aufsteigende Maske), und die Handprobe mit der Folge des Setzens lässt das ganze Paket grün; die Mutation `s12-attrs-umgekehrt` kehrt die Tabelle um und prüft die Zusage nicht.
- `verifizierbar`: ja — Probe in Baumkopie, `make test` grün
- `klasse`: Fixture gleicht dem Wert der Mutation (Reihenfolge der Eingabe gleich der Soll-Reihenfolge)

### F2 · MEDIUM

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.10 (Prüfen und Handeln in getrennten Abschnitten bei geordneten Zugriffen; Rate mit Zahl der Prozesse), Skill *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*
- `pfad`: `docs/plan/planning/done/slice-telemetry-spans.md` · `| Sperre des Spans (`Set` gegen `End`, `End` gegen `End`) | `Set`, `End` | S23 (2) mit dem Detektor;`
- `befund`: Die Zeile verweist für das Fenster zwischen Prüfen und Markieren auf den Detektor; eine Mutation, die Prüfen und Markieren in zwei Abschnitte unter derselben Sperre legt, ist kein Datenrennen. Sie färbt `TestEndGleichzeitig` in 50 von 50 Prozessen mit Detektor und in 26 von 50 ohne (`flatternd`), und weder Tabelle noch Absatz nennen diese Mutation oder diese Rate; die genannten 3 und 7 von 50 gelten für zwei andere Mutationen.
- `verifizierbar`: ja — `mutationslauf.sh` mit der Klasse `nebenläufig`
- `klasse`: Fenster der Sperre nennt den Detektor für eine Mutation, die er nicht meldet

### F3 · MEDIUM

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger schmaler als Zusage); Sendung von `slice-telemetry-port` in Plan §1 (*Stand zu einem Zeitpunkt*: Zählung von `Count`, `Counts`, `Sum` und `Spans`)
- `pfad`: `src/adapters/driven/telemetry/span_sperre_test.go` · `func verteilungKaputt(tab []spanZeile, s support.Snapshot) string`
- `befund`: Das Orakel des Stands prüft `Count`, `Counts` und die Einträge im Ring, aber nicht `Sum`; die Mutation, die `Sum` in einem eigenen Abschnitt vor der Verteilung zählt, bleibt in `TestSnapshotEinStand` grün (0 von 30 Prozessen mit und 0 von 30 ohne Detektor), während Plan §1 den Fänger für alle Lücken zwischen den Sammlungen einschließlich `Sum` zusagt.
- `verifizierbar`: ja — `mutationslauf.sh`, Klasse `nebenläufig`
- `klasse`: Orakel des Stands schmaler als die Sendung

### F4 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.9 (Plan folgt jeder Korrektur)
- `pfad`: `docs/plan/planning/done/slice-telemetry-spans.md` · `| M2 Paare Span × Schlüssel (S03) |`
- `befund`: Tabelle M2 und Zeile S03 nennen 320 Prüfungen; der Test zählt 618 (2 × (160 + 149), die fremden zusätzlich mit leerem Wert) und der Absatz *Mengen* nennt 618. Die Zahl der Zählprüfung im Plan stimmt an zwei Stellen nicht mit dem Test überein.
- `verifizierbar`: nein — Lesen des Plans
- `klasse`: Zahl der Menge im Plan weicht vom Test ab

### F5 · INFO

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.10, §3.15 (benanntes Negativ mit Grund und Folge)
- `pfad`: `docs/plan/planning/done/slice-telemetry-spans.md` · `*Nicht gebildete Mutationen:*`
- `befund`: Die Begründungen für S10 (`End` hat kein `ctx`) und S14 (Namen werden in einer Abbildung nachgeschlagen) stimmen mit dem Code überein. Für S06 gilt sie nur eingeschränkt: eine Mutation „Geschwister teilen einen Span" ließe sich mit einem Zustand im Adapter schreiben (das Negativ sagt, sie verlange einen im `ctx`); der Test der Geschwister prüft den Abbruch des einen gegen das andere und ist von dieser Lücke nicht betroffen. Zuständig: Implementer, falls er die Begründung schärfen will.
- `verifizierbar`: nein
- `klasse`: Begründung eines nicht gebildeten Fängers zu eng

### F6 · INFO

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `src/adapters/driven/telemetry/span.go` · `a.ring = append(a.ring, r)`
- `befund`: Das `return` hinter dem Anhängen ist beobachtbar gleichwertig mit dem Durchfallen in den Umlauf (Mutation `g08` grün, Zeiger und Länge laufen gleich); kein Fänger fehlt, die Zeile wird nur von der Abdeckung, nicht von einem Test gehalten.
- `verifizierbar`: ja — `mutationslauf.sh`
- `klasse`: gleichwertige Mutation bei Frühausstieg

## Antworten auf die Schwerpunkte

- **Abweichung S11 (`runtime.SetFinalizer` statt `weak.Pointer`):** die Begründung trägt (`testpackage` nimmt allein `export_test.go` aus, die Datei gehört `slice-mock-lesen`). Der Fänger war weder langsam noch flatternd: `TestSpanOhneEnde` 0,02 s unter dem Detektor bei Last 12 bis 16. Die Mutation, die den Span hält, ist laut Plan rot; eigene Wiederholung unter höherer Last steht aus (Obergrenze des Wartens 5 s).
- **Abweichung `idMu`:** der Zuschnitt trägt; die Mutation mit zwei Abschnitten unter `idMu` ist 50/50 und 50/50 rot, `ohne Sperre` und `Beginn bei 0` stehen im Plan als rot.
- **S23 (2) als akzeptiertes Negativ:** Grund und Folge tragen für die beiden genannten Mutationen (Detektor je Prozess 50/50, ohne Detektor selten). Die Variante aus F2 gehört dazu.
- **S15 und S16:** Masse mit Rate ohne Haltepunkt, wie §3.10 verlangt; bei 30 Prozessen je Arm sind beide Mutationen in jedem Lauf rot (Last 15 bis 19 von 4).
- **`TestKernSchreibtNichts`:** vollständig ersetzt. Der Teil „kein Byte, kein Verwurf, keine Wirkung" geht in S01 (fremder Name), S09, S11 und S13 auf; die Prüfung der Reihenfolge der Schlüssel im Stand (`reihenFest`) bleibt an anderen Stellen.
- **Zählprüfungen gegen Quellen:** `support.Spans()` (`spanTabelle`), `support.Eimer()` (`eimerGrenzen`), `logSchluessel(t)`, `support.Verwuerfe()` (`pruefeVektor`), `NumMethod` der Ports (`TestSpanKeinePanik`) stehen im Code und vergleichen gegen die Quelle, nicht gegen eine feste Zahl; die Zahlen 980 und 49 sind aus `2 · (13 + 11) + 1` abgeleitet.
- **Canary (980 Prüfungen):** der Test läuft als Schleife 4 × 49 × 5 mit den Kontrollen (a) und (b). Alle Einbringungswerte sind für jedes Pflicht-Attribut ungültig (Leerraum, Anführungszeichen, `:` und `@`, oder falsche Art), so dass die Schleife den Weg des Verwurfs prüft, den Weg eines gültigen Werts die Kontrolle (a).
- **Sperrreihenfolge in `End`:** Markierung unter der Sperre des Spans, eine Ablesung, Aufzeichnung unter der Sperre des Stands, Zeile unter der Sperre der Zeile, keine Sperre über `Now()` oder `Write`: im Code so (`span.go` `End`, `aufzeichnen`), gefangen durch S18, S23 (1), S24 laut Plan (nicht neu gemessen).
- **Zusatzzeit:** Paket +7,1 s ohne und +33,9 s mit Detektor laut Plan; eigene Messung 95,4 s mit Detektor bei fremder Last, nicht vergleichbar.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `span.go` (Logik von `StartSpan`, `Set`, `End`, `aufzeichnen`, `naechsteID`) | geprüft, ohne Befund am Code; Findings betreffen Fänger |
| `telemetry.go` (Felder, `New`, `Snapshot`) | geprüft, ohne Befund |
| `zeile.go`, `handler.go`, `regeln.go`, `export_test.go`, `hilfen_test.go` | unverändert, geprüft |
| `.golangci.yml`, `.a-check.yml`, `//nolint` | unverändert beziehungsweise keine Direktive, geprüft, ohne Befund |
| Kommentare in `span.go` und `telemetry.go` (`AGENTS.md` §3.7) | geprüft, ohne Befund (Zusage, Kopplung, Grenze) |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -P` auf BOM, NBSP, Nullbreiten- und Zeilentrenner in `span*.go` ohne Treffer; keine Escape-Folge als Fixture |
| Abdeckung der Blöcke (R1 bis R8) | geprüft, ohne Befund (99,2 Prozent, null Blöcke ohne Treffer) |

## Summary

Der Code von `StartSpan`, `Set`, `End` und `Snapshot` hält die geprüften Zusagen: 38 von 39 Mutationen der Klasse `andere` sind rot (die grüne ist gleichwertig), ebenso die Fenster der Aufzeichnung, der Kopie und der `ID`. Drei MEDIUM betreffen Fänger, nicht den Code: die Reihenfolge der Attribute wird nur mit einer Eingabe geprüft, die der Soll-Reihenfolge gleicht (F1), das Fenster zwischen Prüfen und Markieren hat eine Mutation ohne Rate, die der Detektor nicht meldet (F2), und das Orakel des Stands lässt `Sum` aus (F3). Kein HIGH. Die beiden Abweichungen in §6 sind technisch begründet; ihr Ausgang bleibt *offen* mit Adresse Architect.
