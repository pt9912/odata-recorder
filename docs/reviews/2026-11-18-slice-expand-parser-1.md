# Review-Report: slice-expand-parser — 2026-11-18

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers)

**Gegenstand:** `slice-expand-parser`, Zweig `slice-expand-parser`, Diff `1ceb6b83..cb5f0ee0` (Commits `1a9efb6c`, `cb5f0ee0`; `2ac8a8d3` und `1ceb6b83` sind Plan und Vertrag). Pfade: `src/adapters/driven/odata/urlparser/` (`expand.go`, `parser.go`, `query.go`, fünf neue `expand_*_test.go`, geänderte `zustand_test.go` und `grenzen_test.go`), `src/hexagon/model/` (`url_ast.go`, `url_ast_expand.go`, `url_ast_expand_test.go`), `src/bootstrap/expand_wurzel_test.go`, `docs/plan/planning/open/slice-expand-ausfuehrung.md` (eine DoD-Zeile), der Plan `slice-expand-parser` (§3, §6).

**Skill:** `.harness/skills/reviewer.md` @ `9dd4e5db` <!-- d-check:ignore (Zitat des Standes im Lauf) -->
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

**Eingangs-Kontext** (die Verträge, gegen die geprüft wurde):

- Slice-Plan `slice-expand-parser` (§3 Tabelle *Zusage · Fänger* Nr. 1 bis 30, Mengen, Mutationen, Raten; §6 Raster und A1 bis A9; Marke 3500)
- `SPEC-262` (Vertrag von `$expand`), `SPEC-263` (Vertrag für `slice-expand-validator`), `SPEC-160` Nr. 3 bis 8, `SPEC-162` Nr. 2 und 3
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-FA-16.a`, `LH-RB-01`
- `AGENTS.md` §3.7 bis §3.18

**Vorgehen und Belege.** Jede Zeile der Tabelle (Nr. 1 bis 28) wurde mit einer eigenen Mutation aufgelöst; die Kennungen `r01` bis `r39`, `q01` bis `q03`, `z01`, `z02`, `o01`, `o02`, `b01` sind die des Reviewers, sie decken sich nicht mit den `n01` bis `n05` des Implementers. Gefahren mit `sh tools/harness/mutationslauf.sh` (Image `odata-recorder:race`, ohne Netz und Mount, Klasse `andere` ein Prozess je Stufe), je unter `flock` auf dem gemeinsamen Schloss; Host-Last 2,0 bis 4,4 von 4 CPUs, Messender der Reviewer. Der Lauf mit 39 Mutationen (`r01` bis `r39`): 30 `rot`, 1 `rot-paket`, 7 `gruen`, 1 `flatternd`, 0 Übersetzungsfehler; dazu drei kleine Läufe: `q01` bis `q03` (2 `rot`, 1 `rot-paket`), `z01` und `z02` (2 `rot-paket`), `o01` und `o02` (2 `rot`).

Die Klasse wurde getrennt nach **Zeichenmenge** (welche Zeichen ein Name, ein Segment, ein Teil eines qualifizierten Namens trägt) und **Struktur** (Position, Abstand, Zahl, Reihenfolge, Zusammensetzung der Optionen) aufgelöst. Das Orakel der Tests (`exOrakel`, `exIstName`, `exIstPfad`, `exSegmentOrakel` in `expand_hilfe_test.go`) liest keinen Code des Lesers; es ist eine eigene Tabelle aus dem Wortlaut von `SPEC-262` Nr. 3 und 4 (gelesen).

| Zeile | Mutation des Reviewers | Ergebnis | roter Test |
|---|---|---|---|
| 1 | `New` setzt `expand` (r01); `$expand` fällt aus der Klasse der nicht unterstützten Namen (r35) | rot, rot | `TestExpandOhneEinhaengung` |
| 2 | Elemente umgekehrt (r03) | rot | `TestExpandListe` |
| 3 | `WithExpand` setzt `filter` zurück (r02) | **grün** | keiner (F-3) |
| 4 | Wert ein zweites Mal dekodiert (r04) | rot | `TestExpandDekodierung` |
| 5 | zweites Vorkommen überschreibt (r05) | rot | `TestExpandDoppelt` |
| 6 | leerer Wert ergibt `nil` (r06) | rot | `TestExpandLeer` |
| 7 | leere Elemente übergangen (r07) | rot | `TestExpandLeeresElement` |
| 8 | `-` im Namen (r08); NBSP am Ende des Namens (r09) | rot, rot | `TestExpandNamenform` |
| 9 | 129 Zeichen als Name (r10) | rot | `TestExpandNamenGrenze` |
| 10 | `**` als `*` (r11) | rot | `TestExpandStern` |
| 11 | Pfad erst ab drei Segmenten (r12); nur drei Segmente geprüft (r13) | rot, rot | `TestExpandPfad` |
| 12 | `$REF` (r14); `$filter` und `$select` als Segment (r15); `-` im Segment (r16); 129 Zeichen im Segment (r17); Ziffern als Teil nach dem Punkt (r18) | **grün** ×5 | keiner (F-1) |
| 13 | `(` nur in den ersten 100 Bytes gesucht (r22) | rot | `TestExpandKlammer` |
| 14 | Grenze 34 (r23); Zahl nicht bei einem ungültigen Element (r24); leere Elemente zählen nicht (r25) | rot ×3 | `TestExpandZahl` |
| 15 | Fenster der Duplikate vier Elemente (r19); nur Namen verglichen (r20); Vergleich ohne Schreibweise (r21) | **grün**, rot, rot | `TestExpandDoppeltesElement`, `TestExpandContract` (r19: F-2) |
| 16 | `*` oder Pfad am ersten Element beendet das Lesen (r26) | rot-paket | nicht `TestExpandRangImWert` (F-4) |
| 17 | `*` oder Pfad ist sofort der Fehler (q01); `(` ist sofort der Fehler (q02) | rot; rot-paket | q01: `TestExpandRangQuery`; q02 nicht aus ihm (F-4) |
| 18 | Ort zählt leere Segmente (o01, in `model.QuerySegmente`); Ort ab 0 (o02) | rot, rot | `TestExpandOrt` |
| 19 | Eingabe im Ort der Meldung, je Aufrufstelle: doppelt (r27), `zu_viele` (r28), Pfad und `*` (r29), Klammer (r30), ungültiger Name (r31) | rot ×5 | `TestExpandFehlertext` |
| 20 | Paket-Slice als Ergebnis (r32) | rot | `TestExpandKopie` |
| 21 | Zähler färbt jedes zweite Ergebnis (r33) | rot | `TestExpandWiederholung` |
| 22 | Element trägt ein Zeichen mehr aus dem Wert (q03) | rot | `TestExpandWaechter` |
| 23 | Puffer geteilt, mit Wirkung (r38, Klasse `nebenläufig`); Zähler ohne Wirkung (r39); `expand` zur Laufzeit geschrieben (z01, z02) | rot; flatternd; rot-paket ×2 | `TestExpandGleichzeitig`; z01, z02: F-5 |
| 24 | `;` trennt (r34) | rot | `TestExpandAlphabet` |
| 25 | Import von `services/mock` in `expand.go` (von Hand im Arbeitsbaum, zurückgenommen) | rot | `make a-check`: `wrong-direction: adapters-driven -> services` |
| 26 | Zeile `Category,category` zusammengeführt (r21); `$Expand` wie `$expand` (r36) | rot, rot | `TestExpandContract` |
| 27 | `Expand` vor `Sonstige` (r37) | rot | `TestOptionenFelder` |
| 28 | `wurzel.go` hängt `WithExpand()` ein (b01, von Hand in einer Kopie des Baums außerhalb des Repos, `make test`) | rot | `TestSatzVersion4OhneExpand` (4 von 4 Werten); `TestAbnahmeMatrix` bleibt grün |

**Raten** (Zeile 23, `TestExpandGleichzeitig`, je Prozess `-test.count=1`, 20 Prozesse je Arm, Last 2,0 bis 4,4 von 4 CPUs, Messender der Reviewer; Größenordnung): geteilter Puffer mit Wirkung (r38): mit Detektor 20 von 20, ohne 20 von 20; Zähler ohne Wirkung (r39): mit Detektor 20 von 20, ohne 0 von 20. Das stimmt mit den 50er-Raten im Plan überein; das Datenrennen ohne Wirkung fängt allein `make test-race`, wie der Plan es sagt.

**Zeile 28 und die Bestandstests.** `src/bootstrap` läuft im Mutationswerkzeug nicht (`Ausgangsstand rot: src/bootstrap`, `SPEC-210` Nr. 5 e), wie der Plan sagt. Die Mutation b01 lief deshalb in einer Kopie des Baums mit `make test`: in `src/bootstrap` ist allein `TestSatzVersion4OhneExpand` rot (alle vier Werte), `TestAbnahmeMatrix` und `TestM3FilterLeserInJedemAufruf` bleiben grün. Zusätzlich ist `tests/integration` `TestVertragFehlerklassen` rot (`$expand=GEHEIM-Q`: `400` statt `501`); der Plan nennt diesen zweiten Fänger nicht (INFO F-6). Die Änderungen an `TestParserOhneZustand` und `textsumme` sind keine Lockerung (§3.6): der Test hält beide Felder nach Name und Art, ein drittes Feld und ein Feld anderer Art bleiben rot, und `textsumme` zählt nur mehr Text.

**Rohbytes** (§3.18): gelesen wurden `expand*.go`, `query.go`, `parser.go`, `zustand_test.go`, `grenzen_test.go`, `expand_wurzel_test.go`, `url_ast_expand*.go` und der Plan mit `LC_ALL=C grep -nP` auf BOM, NBSP und Backslash-u-Folgen: keine Fundstelle; Zeichen jenseits von ASCII stehen in den Tests allein als `\x`-Bytefolgen in Konstanten (`exE`, `exNBSP`, `exBOM`, `exShe`, `exHan`, `exArab3`, `exHoch2`), in den Kommentaren stehen die Zeichen selbst. Dieser Report wurde nach dem Schreiben mit demselben Muster gegengelesen.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Zeichenmenge im Segment eines Pfads hat keinen Fänger: `-` im Segment (r16), 129 Zeichen im Segment (r17), Ziffern als Teil nach dem Punkt (r18), `$REF` (r14) und `$filter` als Segment (r15) bleiben grün, auch im ganzen Paket. `TestExpandNamenform` und `TestExpandNamenGrenze` führen nur Elemente ohne `/`; `TestExpandBesondereSegmente` führt elf feste Segmente (zwei Nicht-Elemente mit `$`); der Generator (17 Zeichen, Länge 4) erreicht kein Segment mit Punkt und Folgeteil (`a/a.0` hat fünf Zeichen). Szenario: ein Umbau von `expandSegment` mit eigener Namensprüfung lässt `a/b-c` als `501` durch, `SPEC-262` Nr. 4 c verlangt `400`. | `AGENTS.md` §3.15, §3.17; `SPEC-262` Nr. 4 a, c | `expand.go` · `func expandSegment(s string) bool`; `expand_grenzen_test.go` · `func TestExpandBesondereSegmente` | ja — `sh tools/harness/mutationslauf.sh` mit r14 bis r18 | Fänger schmaler als Zusage: Zeichenmenge im Segment eines Pfads |
| F-2 | MEDIUM | Der Abstand eines doppelten Elements ist nur bis 3 geführt; ein Vergleich mit den letzten vier Elementen (r19, `teile[max(0, i-4):i]`) bleibt grün, obwohl `SPEC-262` Nr. 4 e jedes frühere Element meint. Die Prüfgrenze „Abstand 3“ steht weder in der Liste der Prüfgrenzen mit Folge im Betrieb (§3) noch als akzeptiertes Negativ. Szenario: `a,b,c,d,e,a` liefert zwei gleiche Namen an den Validator, den `SPEC-263` Nr. 4 als Programmfehler abweist. | `AGENTS.md` §3.15, §3.17 | `expand_grenzen_test.go` · `func TestExpandDoppeltesElement`; Plan §3 · `Prüfgrenzen mit Folge im Betrieb` | ja — Mutation r19 | Menge unbeschränkt, Prüfgrenze nicht als akzeptiertes Negativ benannt |
| F-3 | MEDIUM | `WithExpand` zusammen mit `WithFilter` hat keinen Fänger: eine Option, die `filter` zurücksetzt (r02), bleibt grün. `TestExpandDifferenz` vergleicht `New()` mit `New(WithExpand())`; die Zusammensetzung des Betriebs (`WithFilter` und `WithExpand`, beide Reihenfolgen) kommt in keinem Test vor. Szenario: nach der Verdrahtung antwortet `$filter` mit `501` (`filter`), und der erste Test, der es sieht, liegt in einem anderen Slice. | `AGENTS.md` §3.15; `SPEC-262` Nr. 1 („ändert kein anderes Feld“) | `expand_test.go` · `func TestExpandDifferenz`; `parser.go` · `func WithExpand() Option` | ja — Mutation r02 | Fänger führt nur eine Form der Eingabe: Zusammensetzung der Optionen von `New` |
| F-4 | LOW | `TestExpandRangImWert` setzt die Verstöße nur an die Stellen 1 bis 4 (das erste Element ist immer `n1`), und `TestExpandRangQuery` führt vier der sechs Arten von Verstößen von `$expand` (nicht die Klammer und nicht den Pfad). Die Mutationen „`*` oder Pfad am ersten Element beendet das Lesen“ (r26) und „`(` ist sofort der Fehler“ (q02) werden aus diesen Tests nicht rot (`rot-paket`), sondern aus `TestExpandContract`, `TestExpandDoppeltesElement` und `TestExpandDoppelt`. Rot, aber nicht aus dem Test, den Plan §6 für die Zeilen der Zusatztabelle nennt. | `AGENTS.md` §3.16 (Zusatztabelle: Fänger je Paar) | `expand_grenzen_test.go` · `elemente := []string{"n1", "f2", "f3", "f4", "f5"}`; `expand_grenzen_test.go` · `func TestExpandRangQuery` | ja — Mutationen r26, q02 | Paar von Verstößen an Stelle 0 und je Art nicht im benannten Fänger der Zusatztabelle |
| F-5 | LOW | Das Merkmal `expand` des Parsers ist ein `bool`; ein Schreiben im Aufruf (z01: nach `$top` wird `expand` gesetzt) macht weder `TestExpandOhneEinhaengung` noch `TestParserOhneZustand` rot (beide `rot-paket`, gefangen allein von den Bestandstests `TestGeneratorVerstoss` und `TestRangInDerOption`). Der Test hält Name und Art der Felder, `SPEC-262` Nr. 8 sagt „danach unveränderlich“. | `AGENTS.md` §3.15; `SPEC-262` Nr. 8; Plan §3 Zeile 23 | `zustand_test.go` · `func TestParserOhneZustand`; `expand_test.go` · `func TestExpandOhneEinhaengung` | ja — Mutationen z01, z02 | Zusage über eine Gestalt, Wirkung nur im Paket gefangen |
| F-6 | INFO | Zeile 28: neben `TestSatzVersion4OhneExpand` fängt `tests/integration` `TestVertragFehlerklassen` den Aufbau (b01: `400` statt `501`); der Plan nennt nur den Unit-Fänger. Kein Handlungsbedarf, nur eine zweite Belegstelle, die `make test` fährt und das Mutationswerkzeug nicht. | Maintainability | Plan `slice-expand-parser` §3 · `der Fänger des Aufbaus ist` | ja — `make test` mit b01 | — |
| F-7 | INFO | Die Messung der Marke liegt bei 1298 (Summe aus Einfügungen und Löschungen von `d158068d..cb5f0ee0` über die Pfade des Blocks), die Marke bei 3500, die Erwartung des Plans bei 1500 bis 2400; die Messung liegt unter der Spanne. | `AGENTS.md` §3.14 | Plan `slice-expand-parser` §4 · `Spanne 1500 bis 2400` | ja — `make groessenmarke` | — |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/odata/urlparser/expand.go`, `query.go`, `parser.go` (Code gegen `SPEC-262` Nr. 1 bis 8) | geprüft, ohne Befund: Folge der Prüfungen, Klammer vor Zahl vor Elementen, Klasse `ungueltig` vor `nicht_unterstuetzt`, Ort `query/<n>`, `option_doppelt` auch bei `%24expand`, `$Expand` als `option_name`; Kommentare tragen Zusage, Kopplung oder Rang-Zeiger (§3.7) |
| `src/hexagon/model/url_ast.go`, `url_ast_expand.go`, `url_ast_expand_test.go` | geprüft, ohne Befund: `Expand` als letztes Feld, Feldliste per `reflect`, Mutation r37 rot |
| `src/bootstrap/expand_wurzel_test.go` | geprüft, ohne Befund: Fänger des Aufbaus, Mutation b01 rot; der Kopf sagt zu, was der Test prüft (§3.11), und nennt, was offen bleibt |
| `zustand_test.go`, `grenzen_test.go` (geänderte Bestandstests) | geprüft, ohne Lockerung (§3.6); siehe F-5 für die Grenze der Gestalt |
| Köpfe der `expand_*_test.go` (§3.11) | geprüft: die Mengen in den Köpfen stimmen mit den Schleifen überein (125 ASCII-Zeichen, 5224 Folgen je Wiederholung, 432 Paare, 36 Orte, 88 740 Folgen); der Kopf von `TestExpandNamenform` sagt „an jeder Stelle des Namens“ und meint ein Element ohne `/` (siehe F-1) |
| Orakel der Tests (`expand_hilfe_test.go`) | geprüft, ohne Befund: kein Aufruf des Lesers; die Tabelle der Zeichenklassen ist endlich und benannt |
| Plan §6, §3.13 (Adressen) | geprüft, ohne Befund: `slice-expand-ausfuehrung` trägt `TestSatzVersion4OhneExpand` und die Verdrahtung in DoD und §1 mit Geber `slice-expand-parser`; `slice-expand-validator` führt den Schnitt mit Geber in §1; die Planner-Adresse (vier gegen fünf Slices der Welle) steht in §6 mit Ausgang *offen* |
| Plan §3, §6 gegen die Spezifikation (§3.9, §3.16) | geprüft, ohne Befund: Raster und Zusatztabelle haben je Zelle einen Ort in `SPEC-262`; keine Zelle im Code entschieden |
| `docs/plan/planning/open/slice-expand-ausfuehrung.md` | geprüft, ohne Befund |
| Rohbytes und unsichtbare Zeichen (§3.18) | geprüft, ohne Befund (siehe oben) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 2 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fänger schmaler als Zusage (Zeichenmenge im Segment) · Menge unbeschränkt ohne Prüfgrenze (Abstand der Duplikate) · Fänger führt nur eine Form der Eingabe (Zusammensetzung der Optionen) · Paar von Verstößen nicht im benannten Fänger der Zusatztabelle · Zusage über eine Gestalt, Wirkung nur im Paket gefangen

## Verdikt

**Merge-blockierend:** ja — drei MEDIUM (F-1 bis F-3), jedes mit einer Mutation, die im ganzen Paket grün bleibt. Keine HIGH: der Code selbst weicht an keiner geprüften Stelle von `SPEC-262` ab; die sieben grünen Mutationen sind Lücken der Fänger, keine Fehler des Lesers.

**Übergabe:** Findings gehen an den Implementer. Ein Fix zählt gegen die Klasse (§3.17): für F-1 die Menge der Zeichen, Längen und Teile jedes Segments eines Pfads, für F-2 der Abstand der Duplikate bis 32 oder die Prüfgrenze als akzeptiertes Negativ, für F-3 die Zusammensetzungen von `WithFilter` und `WithExpand` in beiden Reihenfolgen; sein Bericht nennt je Befund die Menge, die Zahl ihrer Elemente und eine Mutation an einem Element, das dieser Report nicht nannte. Die Finding-Klassen gehen in die Slice-Closure §7. Dieser Report ersetzt keine Verifikation — DoD und Spezifikation prüft der Verifier.
