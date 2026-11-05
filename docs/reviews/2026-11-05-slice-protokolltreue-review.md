# Review-Report: slice-protokolltreue — 2026-11-05

**Review-Art:** Code (Test-Code, Produktcode 0) — geprüft gegen Plan (`slice-protokolltreue` §1, §3 mit der Tabelle *Zusage · Fänger* Z1 bis Z22 und dem Handoff-Absatz, §4 mit Block und Messpunkt, §6 mit Raster und Zusatztabelle), gegen `LH-QA-07.a` Nr. 1 bis 7, `SPEC-119` Nr. 6, `SPEC-121` Nr. 4 und `AGENTS.md` §3.

**Gegenstand:** `slice-protokolltreue`, Zweig `slice-protokolltreue`, Diff `195afaa..c588bb6` (Basis der Randformen `7c4e1df`): `tests/integration/protokolltreue_test.go` (1028 Zeilen), `vergleich_test.go` (251), `vergleichsdienst_test.go` (125), Plan (+48/-5). Kein Produktcode.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-05

**Methode.** Einmaliger Lauf, kein Bericht des Implementers übernommen. Alle drei Testdateien vollständig gelesen, Spec Nr. 1 bis 7 und `SPEC-119` Nr. 6 dagegen gehalten. Mutationen auf Wegwerfkopien von `src/`, `tests/`, `go.mod` im Scratchpad (`rev-pt/`), Image `odata-recorder:race` (`--network none --memory 4g`, unter `flock /tmp/odata-recorder-heavy.lock`), je Mutante eine Ersetzung, `go test -count=1` über die Tests des Vergleichs. Basislauf der Suite grün. `make lint` einmal gefahren: grün. `make gates` nicht gefahren (Verifier-Frage). Z21 habe ich selbst gemessen (siehe unten). Nicht jede Zeile Z1 bis Z22 habe ich mit einer eigenen Mutation aufgelöst: Z2 bis Z7, Z10 bis Z14, Z18, Z19 sind gelesen und gegen die Mutanten unten gespiegelt, nicht einzeln mutiert; die Mutanten wurden nach Formen der Klassen gewählt (§3.17), die der Fänger nicht nennt.

**Eingangs-Kontext:** Slice-Plan in `c588bb6`; `LH-QA-07`, `LH-FA-01`; `LH-QA-07.a`; `SPEC-119` Nr. 6; `AGENTS.md` §3.7, §3.9 bis §3.18.

---

## Findings

| Nr. | kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|---|
| 1 | MEDIUM | `AGENTS.md` §3.15 (Fänger führt nur eine Form), §3.17 | `tests/integration/vergleich_test.go` · `out[n] = v` (in `relevante`) und `!slices.Equal(dv, pv)` (in `vergleiche`) | Die Klasse „Liste der Feldzeilen je Name, leere Werte eingeschlossen" (`LH-QA-07.a` Nr. 1: gleich viele Zeilen mit denselben Werten) hat Fänger für fehlend, `x`, vertauscht, zusammengeführt und eine dritte Zeile `c`; drei Formen bleiben grün. Mutante a: leere Werte aus einer Liste mit mehr als einer Zeile entfernt; Mutante b: `slices.Compact` auf die Liste (zwei gleiche Zeilen gelten als eine); Mutante c: Werte mit `strings.TrimSpace` verglichen. Alle drei: die gesamte Suite grün (`ok`). Ein Recorder, der eine leere Zeile neben einer gefüllten verwirft, eine doppelte Zeile zusammenlegt oder Leerraum am Rand eines Werts trimmt, fiele dem Vergleich nicht auf; der Fall `header` führt `X-Mehrfach` mit `a`, `b` und `X-Leer` nur allein. | ja (Mutation) | Fänger führt eine Form der Zeilenliste (leer neben gefüllt, gleiche Zeilen, Leerraum am Rand) |
| 2 | MEDIUM | `AGENTS.md` §3.15, §3.11 (Kommentar sagt mehr zu als ein Test prüft); `LH-QA-07.a` Nr. 3, `SPEC-119` Nr. 6 | `tests/integration/vergleich_test.go` · `// entpackeGlied nimmt eine Stufe ab: gzip (auch mehrere Glieder hintereinander)` und `case "gzip", "deflate", "identity":` in `kette`; Fänger `TestNichtEntpackbar`, `TestKetten` | Die Kettenmenge (drei Namen, höchstens zwei Glieder, Name ohne Beachtung der Schreibweise, leerer Rohbody bei unbekannter Kette ist rot, mehrere gzip-Glieder sind ein Strom) ist nur in Teilen gefangen. Mutante a: `g := asciiKlein(t)` zu `g := t` in `kette` (Schreibweise): grün. Mutante b: `len(roh) == 0` vor die Prüfung der Kette (ein leerer Rohbody mit unbekannter Kette gilt als leer statt rot, Spec `SPEC-119` Nr. 6: Grund `encoding`): grün, weil alle 19 Ketten von `kaputteKetten` den Rohbody `x` oder einen kaputten Strom tragen, keine einen leeren. Mutante c: `gr.Multistream(false)` im gzip-Zweig: grün; der Kommentar sagt Mehrfach-Glieder zu, kein Fall führt zwei gzip-Glieder. Die Folge liegt im Test, nicht im Produkt: die Vergleichsfunktion entscheidet das Urteil des Slice (grün oder Fehler des Tests). | ja (Mutation) | Kettenklasse nur in Teilformen gefangen; Kommentar sagt Ungeprüftes zu |
| 3 | LOW | `AGENTS.md` §3.16 (Randform im Code entschieden), Maintainability | `tests/integration/vergleich_test.go` · `// nachName ordnet die Feldzeilen nach dem kleingeschriebenen Namen; Zeilen verschiedener Schreibweisen` | Die Reihenfolge der Zeilen zweier Schreibweisen desselben Namens in einer Map („folgen der Bytefolge der Schreibweisen") steht weder in `LH-QA-07.a` Nr. 1 noch im Raster, nur im Kommentar; ihr Fänger fehlt: Mutante `out[k] = h[n]` statt `append` (die zweite Schreibweise überschreibt die erste): grün. Auf dem Draht tritt das nicht auf (`net/http` kanonisiert), nur an handgebauten Maps. | ja (Mutation) | Randform im Test-Code entschieden, ohne Fänger |
| 4 | LOW | `AGENTS.md` §3.9 (Plan folgt jeder Korrektur), §3.15 (Negativ mit Grund und Folge) | `docs/plan/planning/in-progress/slice-protokolltreue.md` · `| Z13 | Eine Antwort ohne Body (` und `(2) Z13: \`net/http\` sendet bei HEAD, 204 und 304 keinen Body` | Die Tabellenzeile Z13 führt weiter den Plan-Wortlaut („Mutator liefert bei 204 einen Body nach"), die Umsetzung ist handgebaut; die Abweichung steht nur im Handoff-Absatz darunter. Das Negativ HEAD (Go-Client sieht einen Body nach HEAD nicht) trägt dort Grund und Folge („keine, der Server verwirft ihn"), steht aber weder in `LH-QA-07.a` Nr. 3 noch in Nr. 7; „der Server" ist der `net/http`-Server des Recorders, im Text nicht benannt. Nach §3.15 genügt Grund und Folge im Plan; Tabelle und Absatz widersprechen einander im Wortlaut. | nein | Tabellenzeile und Abweichung nicht zusammengeführt |
| 5 | INFO | `AGENTS.md` §3.10 | Plan §3 *Raten* · `Z21 ... rot in 30 von 30 Prozessen` | Rate gemessen und bestätigt (Image `odata-recorder:race`, `-race`, Test-Binary einmal gebaut, 30 Prozesse `-test.count=1`, `TestParallel`, Messender: der Reviewer): Mutation (Paketvariable `zaehlerMut++` in `bediene` des Referenzdienstes) rot in 30 von 30, ohne Mutation 0 von 30. Z21 fängt nur ein Rennen im Test-Server selbst; ein Rennen des Recorders unter dem Vergleich meldet nur der Detektor, soweit ein Lauf es auslöst (`SPEC-122` Nr. 8). | ja | — |
| 6 | INFO | `SPEC-128` Nr. 9 | Plan §4 · `Messpunkt slice-protokolltreue.md: ... messung=1404` | Messung bestätigt: `git diff --numstat 7c4e1df..c588bb6 -- tests/` ergibt 1028 + 251 + 125 = 1404, Produktcode 0, unter der Marke 2700. | ja | Größenmarke |
| 7 | INFO | Maintainability | `tests/integration/vergleich_test.go` · `if t != "" {` in `verbindungsnamen` | Mutante „leere Teile von `Connection` als Name führen" bleibt grün; äquivalent (ein Name `""` kommt in keiner Antwort vor). Kein Finding. | nein | äquivalente Mutante |

## Mutationen (durch den Reviewer)

Je Mutante eine Ersetzung, Paket `tests/integration`, Auswahl der Tests des Vergleichs; Ergebnis aus der Ausgabe von `go test`.

| Zusage | Mutation | Ergebnis |
|---|---|---|
| Z1 (Phrase zählt nicht) | `direkt.phrase != proxy.phrase` im Statusvergleich | rot: `TestStatusNichtPhrase` |
| Z1 (Statuscode, nicht Klasse) | `status/100` verglichen | rot (Status-Tests) |
| Z8 (sieben Namen) | `Age` aus dem Default von `VOLATILE_HEADERS` | rot: `TestFaelleDirektWieErwartet`, `TestVolatileHeaders` |
| Z8 (ein Element, das die Planmutation nicht nannte) | `Traceparent` aus dem Default | rot: dieselben Tests |
| Z9 (Liste der neun) | `Upgrade` durch `Upgrade-X` ersetzt | rot: `TestHopByHopUndConnection` |
| Z15 (Modus roh) | `Content-Encoding` auch im Modus roh ausgenommen | rot: `TestNichtAngefragtesCoding` u. a. |
| Z16 (Transport entpackt nicht) | `DisableCompression: false` im Client | rot: `TestAcceptEncodingErreichtDenDienst`, `TestTextAntworten` |
| Z17 (keine Weiterleitung) | `CheckRedirect` des Clients entfernt | rot: `TestSelbstvergleich`, `TestStatusAbweichung` u. a. |
| Z21 | siehe INFO 5 | rot 30 von 30, ohne 0 von 30 |
| Z5, Z6 (Klasse Zeilenliste) | a, b, c aus Finding 1 | **grün** (Finding 1) |
| Z11, Z14 (Kettenklasse) | a, b, c aus Finding 2 | **grün** (Finding 2) |
| Nr. 1 (Zeilen verschiedener Schreibweisen) | `out[k] = h[n]` | **grün** (Finding 3) |
| Z22 | gelesen: `git diff --stat 195afaa..c588bb6` über `m2_test.go`, `referenzdienst_test.go`, `text_test.go` ist leer | bestätigt |

## Raster (§6 des Plans) gegen `LH-QA-07.a`

Jede der fünfzehn Zellen und die Zusatztabelle gelesen: jede nennt Nr. 1 bis 7 der Spezifikation oder „gilt nicht" mit Grund; die genannten Nummern tragen den Wortlaut der Entscheidung. Eine im Code entschiedene Zelle habe ich nur in der Form von Finding 3 gefunden (Reihenfolge verschiedener Schreibweisen derselben Namens). Die Zellen *Abbruch* und *Upstream nicht erreichbar* (Z20) sind akzeptierte Negative mit Grund und Folge in Nr. 7 a bis c.

## Einordnung gegen Plan, ADR und Hard Rules

- **Plan §3 (Tabelle gegen Umsetzung):** Z1 bis Z19 haben je einen Test; die Mengen der Zählprüfungen (33, 68, 36, 63, 27, 15, 44, 30, 9, 38) stimmen mit den Schleifen überein (68 nachgerechnet aus `relevantMit`). Abweichungen des Handoff-Absatzes (Z13, Z11/Z14 handgebaut, eigener Referenzdienst) sind benannt.
- **§3.7 / §3.11 (Kommentare):** kein Konjunktiv über verworfene Alternativen, kein abwesender Text; Kopfzeile des Tests der direkten Antworten im Indikativ (letzter Commit). Eine Zusage ohne Fänger: der Kommentar von `entpackeGlied` (Finding 2).
- **§3.18 (Rohbytes):** keine Escape-Folge aus Backslash und `u` in den drei Dateien (`grep -P '\\u[0-9a-fA-F]{4}|\\x'` leer); `brotliBytes()` ist aus Bytewerten gebaut; erwartete Header und Bodies stehen ausgeschrieben.
- **§3.2:** keine `//nolint`; `make lint` grün.
- **ADR-0012:** Importe des Pakets `tests/integration` gegen `src/bootstrap` entsprechen den vorhandenen Tests des Pakets; `make a-check` nicht gefahren.
- **Kritischer Pfad (Capture, Redaction, Replay):** nicht berührt; kein Produktcode.

## Geprüft, ohne Befund

- `tests/integration/vergleichsdienst_test.go` — zustandslos (Antwort allein vom Request), Echo von `Accept-Encoding` je Antwort, 16 Routen; ohne Befund.
- `tests/integration/m2_test.go`, `referenzdienst_test.go`, `text_test.go` — ohne Diff; ohne Befund.
- `tests/integration/protokolltreue_test.go` — Veränderer, Bausteine, `TestParallel`; ohne Befund außer den Findings 1 bis 3 (Fänger der Zeilen- und Kettenklasse).
- `docs/plan/planning/in-progress/slice-protokolltreue.md` — §3, §4, §6 gelesen; Befund 4.
- `spec/spezifikation.md` — im Diff unverändert (`LH-QA-07.a` seit `7c4e1df`); ohne Befund.
- `src/` — im Diff unverändert; ohne Befund.
