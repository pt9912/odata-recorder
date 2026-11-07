# Review-Bericht: slice-mock-zuordnung — 2026-11-07

**Rolle:** Reviewer (Modul 10) · **Frage:** Ist der Diff gegen Plan, Entscheidungen und Hard Rules wartbar und richtig? — nicht gegen die DoD (Verifier).

**Review-Art:** Code (gegen Plan und Konventionen).

**Gegenstand:** `slice-mock-zuordnung` (Teil 1 des Schnitts von `slice-mock-lesen`), Zweig `slice-mock-zuordnung`, HEAD `3eef52b`, Diff `03c9420..3eef52b` (14 Dateien; Pfade des Blocks `groessenmarke`: 2489 Zeilen bei Marke 5800, Produktcode 322). Dieselbe Summe ergibt `git diff --numstat 2fc460a..3eef52b` (Basis des Blocks).

**Geprüft gegen:** Plan `slice-mock-zuordnung` (§3 Tabelle Z01 bis Z24, Mengen M1 bis M6, Absatz *Mutationen und Raten*, §6 Raster und Risiken R1 bis R7), `SPEC-195` (Nr. 1 bis 12), `SPEC-198` Nr. 4 und 5, `SPEC-109` Nr. 5 und 6, `AGENTS.md` §3.2, §3.6, §3.7, §3.9 bis §3.11, §3.13 bis §3.18. Skill `.harness/skills/reviewer.md`.

**Methode.** Einmal gefahren, keine Wiederholungsschleife. Der Bericht des Implementers ist nicht übernommen: 100 eigene Mutationen mit `tools/harness/mutationslauf.sh` (Image `odata-recorder:race`, `--memory 3g`, `--cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`; Listen im eigenen Verzeichnis `scratchpad/rv-mock`), dazu eine Wegwerfkopie in einem Container ohne Netz (Cache-Mutation, Abdeckung unter dem Detektor). Ergebnis der Mutationen: 86 rot, 3 flatternd (die drei Fänger der Nebenläufigkeit, erwartet: mit Detektor 20/20, ohne 0/10), 2 Übersetzungsfehler (zählen nicht), 9 grün (F-1 bis F-5). Gates im Worktree gefahren: `make test` (Stufe zwischengespeichert, grün), `make lint`, `make a-check`, `make katalog-check`, `make groessenmarke`, `make zeichen-check` je Exit 0; die Dateien des Slice sind ohne Treffer auf Backslash-u-Folgen, BOM, NBSP und Zeilentrenner gegengelesen (`LC_ALL=C grep -P`), `.golangci.yml` und die übrigen Dateien der Gate-Kette sind im Diff unberührt, keine `//nolint`-Direktive.

**Grenze dieses Laufs (ehrlich).** Nicht gefahren: Mutationen an den Zeilen Z08/Z09 im Code von `outbound.Abbilden` (liegt außerhalb des Diffs), die Kontext-Mutationen an S2 und S3 (`z20c`, `z20d`; gefahren sind S4 und S12), Z24 (`SchreibeFehler` im Dienst), die Fakes (Z21) und die Wirkungs-Mutation `z22d`. Der Satz „80 von 80 rot“ des Plans ist nicht nachgefahren; eine seiner Mutationen (`z12c`) ist in der Form, die ich daraus ableite, nicht rot (F-3).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Der Gegenstand der Fehlerabbildung ist an S1 (`ParseURL`), S4 (`Blobs.Open`) und S12 (`SchreibeAnzahl`) nicht festgehalten: Wird er auf `Dienst` oder `Workspace` gesetzt, bleiben `TestURLUrsachen`, `TestKompatUndRest` und `TestMetadatenFehler` grün (Mutationen r18a bis r18d). Kein Fall speist `ErrNichtGefunden` oder `ErrBeschaedigt` an diesen Stellen ein; unter `Dienst` wäre das `ODR-E5005`/404 statt `ODR-E9001`. | `AGENTS.md` §3.15 (Aufrufstellen als Menge: Gegenstand der Fehlerabbildung je Stelle), `SPEC-195` Nr. 11 | `src/hexagon/services/mock/metadaten.go` · `abbilden(err, outbound.GegenstandSonstiges)` (zwei Stellen); `mock.go` · `u, err := d.Codecs.URL.ParseURL(pfad, a.RawQuery)` | ja (`mutationslauf.sh`: r18a/b/c/d `gruen`) | Gegenstand der Fehlerabbildung je Aufrufstelle nicht gefangen |
| F-2 | MEDIUM | Die Zusage „merkt sich weder Modell noch Coverage noch Entity-State, liest in jedem Aufruf neu“ hat keinen Fänger. Eine Wegwerfkopie mit einem Cache `map[string]model.Service` als Feld von `Mocker` (gefüllt in `holeService`, ohne Sperre) lässt `go test` und `go test -race` über das Paket grün: Jeder Test baut den Dienst neu (`u.rufe`), `TestGleichzeitig` füllt den Cache im sequentiellen Lauf von `soll` und liest ihn danach nur. `TestPaketOhneZustand` prüft Paketvariablen, Importe und Typen mit Methoden, nicht die Felder von `Mocker`; Mutation r22f (zusätzliches Feld) bleibt grün. Der Plan nennt „Cache des Modells im Dienst“ als gefangene Mutation von Z22. | `AGENTS.md` §3.15 (Zusage ohne Fänger), §3.10 (Zusage über Nebenläufigkeit), `SPEC-195` Nr. 12 | `src/hexagon/services/mock/zustand_test.go` · `func TestPaketOhneZustand`; Plan §3 · `Cache des Modells im Dienst` | ja (Wegwerfkopie; `mutationslauf.sh` r22f `gruen`) | Zusage „kein Zustand, liest neu“ ohne Wirkungs-Fänger |
| F-3 | MEDIUM | Die Fixtures des fremden Service tragen Kennungen anderer Länge als `svc` (`anderer` 7, `x` 1); die Mutation „nur die Länge von `ServiceID` verglichen“ (`len(w.ServiceID) != len(d.ID)`, r17a) bleibt grün. Der Plan führt `z12c` („nur die Länge verglichen“) als rot; in dieser Form ist sie es nicht. Der Fall „fremder Service, Kennung gleicher Länge“ fehlt. | `AGENTS.md` §3.15 (Fixture gleicht nie dem Wert der Mutation; beide Seiten der Grenze) | `src/hexagon/services/mock/zuordnung_test.go` · `ServiceID: "anderer"`; Plan §3 · `z12c` | ja (r17a `gruen`) | Fixture deckt die Mutation nicht (Länge statt Wert) |
| F-4 | LOW | `asciiKlein` ist die Zusage „nur ASCII-Großbuchstaben“ (`SPEC-195` Nr. 5); ersetzt man den Vergleich durch `strings.ToLower` oder `strings.EqualFold`, bleibt M3 grün (r13a, r13b), weil keine Form einen Namen mit einem Zeichen trägt, das Unicode faltet (etwa U+212A in `Work`). Mit dem HTTP-Adapter ist das nicht erreichbar (Header-Namen sind Token aus ASCII), daher LOW. | `AGENTS.md` §3.15 (jede Form, die der Code verschieden behandelt) | `src/hexagon/services/mock/dienst.go` · `asciiKlein(n) == klein`; `zuordnung_test.go` · `"Name gemischt"` | ja (r13a/r13b `gruen`) | Fänger führt nur eine Form der Eingabe (ASCII-Faltung gegen Unicode-Faltung) |
| F-5 | LOW | Die Liste der 13 Codes in `TestFehlerWert` (`spec195Nr10`) ist eine Handkopie; Plan (§6, Z19) und Test-Kommentar sprechen von der „Liste in `SPEC-195` Nr. 10“. Ändert die Spezifikation die Liste, bleibt der Test grün. | `AGENTS.md` §3.11 (Kommentar sagt nur zu, was ein Test prüft) | `src/hexagon/services/mock/rang_test.go` · `// spec195Nr10 ist die Liste der erreichbaren Codes in SPEC-195 Nr. 10` | nein (kein Lauf liest die Spezifikation) | Zählprüfung gegen eine Kopie statt gegen die Quelle |
| F-6 | LOW | R5 (`SPEC-180` Nr. 6, Nachzug im Merge-Commit) steht in §6, aber weder in der DoD (die Zeile *Doku-Update* nennt nur `SPEC-168` Nr. 10, R2) noch im Kopf *Berührte Spec-Stellen*; wer allein die DoD abhakt, übergeht den Nachzug. R2 steht in DoD, Kopf, §1 und §6. | `AGENTS.md` §3.9 (Kopf und betroffene Stellen im selben Commit), §3.13 | Plan · `Doku-Update, falls ein öffentlicher Vertrag berührt ist; dazu der Nachzug von SPEC-168 Nr. 10 (R2, §1)` | nein | Berichtigung erreicht nicht jede Stelle (Plan-Kopf, DoD) |
| F-7 | INFO | Zwei Roots gleicher Länge, die denselben Pfad treffen, sind nur identische Roots; die Konfiguration schließt sie aus (`SPEC-110` Nr. 7, 9). `>=` statt `>` in `dienstFuer` (r12b) bleibt deshalb grün und ist gleichwertig; weder `SPEC-195` Nr. 4 noch das Raster nennen den Gleichstand mit „gilt nicht“. | Maintainability (Raster, `AGENTS.md` §3.16) | `src/hexagon/services/mock/dienst.go` · `len(d.Root) > len(gewaehlt.Root)` | ja (r12b `gruen`, gleichwertig) | Randform ohne ausdrückliches „gilt nicht“ |

## Zu den Schwerpunkten der Beauftragung

- **Z01 bis Z24 und Mengen M1 bis M6.** Rot (aus dem benannten Test, kein `panic`, kein Zeitlimit, `rot-paket` nirgends): M1 alle 17 Elemente einzeln (r01 bis r07; die acht Port-Felder je einzeln, die drei Satz-Felder je einzeln, `MaxBytes` 0 und 2 als Grenze); M2 (r12a, c bis h, j; r12k ist ein Übersetzungsfehler, n02 deckt dieselbe Zusage rot); M3 (r13c bis f, r13h; Ausnahme F-4); M4 sieben Paarmutationen der Rangfolge (r14a bis r14g: Methode/Workspace, Service/Workspace, Zuordnung/Dienst, Methode/URL, Service/URL, Workspace/URL) und der Platzhalter vor `holeService` (r14h); M5 (r17g, r20a/b, n03); M6 S4 und S12 (r19e, r19f; S2 und S3 nicht gefahren). Z02 (r10a, r10b, n01), Z03 (r11a bis g, beide Seiten der Grenze -1/-2 und 0/1), Z10 (r16a bis e), Z13 (r17c, d, f, g, h), Z14 (r19a, b, g, h, l mit 1 MiB), Z16/Z17 (r18e, r18f, r19c, r19d, r19i), Z18 (r15a bis d), Z23 (r21a, r21b) rot.
- **Rangfolge der Paare Z20/Z11.** Jedes der gefahrenen Paare ist rot; die 27 Paare laufen als Schleife mit Orakel „Code der früheren Stufe, Zähler der späteren Ports“.
- **Z08 16 Ursachen × 3 Pfade, Z09, Fehlerabbildung 11 Codes aus 44 Fehlerzeilen.** Die Mengen werden gegen `model.AlleURLUrsachen()` (16) und `AlleKompatUrsachen()` (6) gezählt, Orakel von Hand; die Abbildung selbst liegt in `outbound.Abbilden` und ist nicht Teil des Diffs. Lücke: F-1.
- **Dienstwahl, Pfad relativ zum Root, Header, Methode, `$metadata`.** Siehe M2, M3, Z10, Z14; BOM und Nicht-UTF-8 entstehen aus Bytewerten (`metadatenInhalte`), 1 MiB als Prüfgrenze benannt (akzeptiertes Negativ mit Grund und Folge in §3).
- **Codecs nil, Aliasing.** r08 (Zeiger des Satzes geteilt) und r09 (Slice nicht geklont) rot in `TestNewKopiert`; r10a/r10b (Satz je Feld) rot.
- **Platzhalter Z18.** r14h, r15a bis d rot; Collection, Entity und Anzahl enden mit `ODR-E6001`, ohne `Entities` und `Coverage`.
- **Fakes (§3.15).** `TestFakesZaehlenAlles` zählt 28 Operationen der acht Ports per Reflexion, hält Argumente und `ctx` fest und speist einen Fehler ein (gelesen; Mutation an den Fakes nicht gefahren).
- **Z22.** Raten nachgefahren: r22a, r22b, r22e (Schreibzugriff auf die Konfig in `Mock`, tief in `metadaten`, auf ein Element von `Dienste`): mit Detektor 20/20 rot, ohne 0/10 (`flatternd`, unter Last 4,1/4); das deckt die Raten des Plans (50/50 und 0/50) und das akzeptierte Negativ (2). Die Abdeckung ist nachgemessen: `go test -race -coverprofile` über das Paket 100,0 %, `TestGleichzeitig` allein 93,5 %, ungedeckt nur `New` (88,9 %) und `pruefeKonfig` (40,0 %). Grund und Folge des Negativs (1) tragen: `New` läuft vor dem ersten Request und gleichzeitig ist nur `Mock`. Offen bleibt die Zusage „kein Zustand“ als Feld von `Mocker` (F-2). Fünf Läufe von `go test -race` über das Paket (vier gemessen, ein Lauf der Abdeckung) sind grün.
- **`TestPaketOhneZustand` (Robustheit).** Liest `*.go` relativ zum Arbeitsverzeichnis; `go test` setzt es auf das Paketverzeichnis, und der Test ist unter `make test`, im Container des Mutationslaufs und unter `-race` grün. `len(r) < 4` hält die Löschung einer Quelldatei fest. Der Test kennt `os`, `net`, `time`, `sync` und `database/sql` als verbotene Importe; das Gate `lint` und `a-check` laufen grün.
- **R7.** Teil 1 ist ohne `ODR-E6008` und `model.AlleRessourcenArten`/`AlleCoverageStatus` vollständig: Der Code nennt `ODR-E6008` nirgends, `model` trägt 37 Codes, `make katalog-check` ist grün (`SPEC-201` Nr. 1: ein Name der Spezifikation, den der Quelltext noch nicht führt, ist kein Befund). Die Spezifikation führt 13 Codes in `SPEC-195` Nr. 10 und 38 in `SPEC-109` Nr. 2; der Plan führt die zwei Codes in L37 von Teil 2.
- **R6.** In §6 korrekt geführt: Ausgang *offen*, Adresse Architect, mit dem Befund (Kennzeichen mit leerem Namen setzt der Dienst unverändert in den Header). Nicht entschieden, nicht geprüft.
- **R2/R5.** R2 steht in DoD, Kopf, §1 und §6; R5 nur in §6 (F-6).
- **Größenmarke.** 2489 gegen 5800; Produktcode 322 gegen 740 geschätzt, Tabelle und Block konsistent; `make groessenmarke` grün.
- **Lint.** `.golangci.yml` und alle Dateien der Gate-Kette im Diff unberührt (§3.2, §3.6), keine `//nolint`; `make lint` Exit 0.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/inbound/mock.go`, `mock_test.go` | geprüft, ohne Befund (r21a, r21b rot) |
| `src/hexagon/services/mock/fehler.go` | geprüft, ohne Befund (Status des Codes, Ursache über `Unwrap`; r20a, r20b rot) |
| `src/hexagon/services/mock/mock.go` (Ablauf, Konfig, Argumente, Methode, Workspace, Service) | geprüft; Befunde F-1 (S1), F-3, F-7 |
| `src/hexagon/services/mock/metadaten.go` (Strom, `HEAD`, Schließen, `ctx`) | geprüft; Befund F-1 (S4, S12) |
| `src/hexagon/services/mock/dienst.go` | geprüft; Befunde F-4, F-7 |
| `src/hexagon/services/mock/*_test.go` | geprüft; Befunde F-2, F-3, F-5; Rohbytes ohne Treffer |
| `docs/plan/planning/in-progress/slice-mock-zuordnung.md` | geprüft; Befund F-6; Plan folgt dem Code (§3.9) bis auf die genannten Stellen |
| `spec/` | im Diff unberührt; Nachzug R2 und R5 steht für den Merge aus |
| `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/` | im Diff unberührt |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 3 |
| INFO | 1 |

**Finding-Klassen dieses Laufs:** Gegenstand der Fehlerabbildung je Aufrufstelle nicht gefangen · Zusage „kein Zustand, liest neu“ ohne Wirkungs-Fänger · Fixture deckt die Mutation nicht (Länge statt Wert) · Fänger führt nur eine Form der Eingabe · Zählprüfung gegen eine Kopie statt gegen die Quelle · Berichtigung erreicht nicht jede Stelle · Randform ohne ausdrückliches „gilt nicht“

## Verdikt

**Merge-blockierend:** ja — drei MEDIUM (F-1 bis F-3), jede mit Mutation belegt und klein im Fix; kein HIGH. F-4 bis F-7 blockieren nicht.

**Übergabe:** Findings gehen an den Implementer; die Finding-Klassen gehen in die Slice-Closure §7. Der Fix deckt die Klasse der Zusage, nicht die genannte Mutation (`AGENTS.md` §3.17): zu F-1 alle Stellen S1, S4, S12 (und die Ports S2, S3 mit einem Fehler, den der Gegenstand verschieden abbildet), zu F-2 die Menge „Zustand des Dienstes“ (Felder, Aufruf zweimal gegen veränderten Zustand), zu F-3 die Menge „fremde Kennung“ (gleiche Länge, gleiches Präfix, Teilzeichenkette). Dieser Bericht ersetzt keine Verifikation.
