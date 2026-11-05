# Review-Report: slice-text-antworten — 2026-11-05

**Review-Art:** Code — geprüft gegen Plan, Spezifikation (`SPEC-118` Nr. 1, 3, 11, `SPEC-119` Nr. 2 und 8, `SPEC-121` Nr. 5, `SPEC-135` Nr. 7 c), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) und Hard Rules (`AGENTS.md` §3), nicht gegen die DoD (Verifier). Ein Lauf, keine Wiederholungsschleife.

**Gegenstand:** `slice-text-antworten` (Welle `welle-record-replay`), Zweig `slice-text-antworten`, Diff `0686957..07d780d` samt Plan (Tabelle *Zusage · Fänger* Z1 bis Z17, Raster §6, Absatz *Belege des Implementers* in §3).

**Skill:** `.harness/skills/reviewer.md` @ HEAD `07d780d` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-05

**Eingangs-Kontext:**

- Slice-Plan `slice-text-antworten` (§1, §3, §4, §6) und der Diff
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Go, Tests in der Stufe `test`)
- `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-QA-07`
- `SPEC-118` Nr. 1 bis 11, `SPEC-119` Nr. 2 und 8, `SPEC-121` Nr. 5, `SPEC-135` Nr. 7 c
- `AGENTS.md` (Hard Rules, besonders §3.7, §3.10, §3.15, §3.17, §3.18)

---

## Methode

Jede Zeile Z1 bis Z17 wurde mit mindestens einer eigenen Mutation aufgelöst, möglichst an einem Element, das der Plan nicht nannte (§3.17). Die Mutationen liefen auf Wegwerfkopien des Worktrees (Perl-Ersetzung, `go test -count=1` im Image der Stufe `deps`, netzlos, `--memory 4g`); nichts davon liegt im Repo. Ausgangsstand: alle Pakete `model`, `record`, `record/capture`, `replay`, `tests/integration` grün. Datenrennen mit dem Image von `make test-race` (`golang:1.27.1-trixie`, cgo, `-race`), je Prozess `-test.count=1`.

| Zeile | Eigene Mutation (Element, das der Plan nicht nannte, kursiv) | Ergebnis |
|---|---|---|
| Z1 | Bedingung `Zaehlung \|\| Wert` zu `&&` | rot (`TestTextNurMitZiel`) |
| Z2 | Rand nur ohne Tabulator getrimmt; `HasPrefix` zu `Contains` (*`xtext/plain`*, *`application/text`*) | rot, rot |
| Z3 | JSON nur ohne `Wert` (*Ziel `Wert` verdrängt `application/json`*) | rot |
| Z4 | Name ohne Tabulator getrimmt; Wert ohne Tabulator getrimmt; `utf8` zusätzlich erlaubt; Parameter ohne `=` übersprungen (*alle vier nicht genannt*) | rot (alle vier, `TestTextZeichensatz`) |
| Z5 | `a` als Ziffer akzeptiert (*Byte `0x61`*) | rot (`TestTextZaehlungBytes`) |
| Z6 | Name mit `TrimSpace` verglichen (*Name `password ` trifft*) | rot |
| Z7 | `ende` schreibt den Marker nur bei `mPos != 0`; halbe Folge am Ende nur ohne Treffer ein Verstoß | rot, rot |
| Z8 | Bereich je Anfangsbyte verschoben: `C2` bis `DF` zu `C3`, `E1` bis `EF` zu `EE`, `F1` bis `F3` zu `F2`, `E0` unten `A1`, `ED` oben `9E`, `F4` oben `8E`; Folgebyte oben `BE` (*sechs von sieben nicht genannt*) | rot (alle sieben, nur der Generator) |
| Z9 | `url.QueryUnescape`; `//$count` ohne Gegenstand; kein `$`-Anfang; erstes statt letztes `/`; Namen ohne `/` davor als Ziel | rot (alle fünf) |
| Z10 | `n <= 1` zu `n <= 2` | rot |
| Z11 | kein neuer Code in dieser Zeile (die Rangfolge steckt im unveränderten `lauf`); die Tests laufen auf Text grün | nicht mutiert |
| Z12 | Pfad mit `%24` zu `$` zurückgesetzt (*Roh-Pfad*); `Wert` nur bei `GET`; Pfad kleingeschrieben; `Zaehlung` nur bei `GET` | rot; rot (nur über `Name` ohne `Wert`, siehe F-3); grün; grün |
| Z13 | siehe unten (Raten) | siehe unten |
| Z14 | Replay setzt `Wert` und `Name` aus dem Pfad | rot |
| Z15 | Option invertiert (`RedigierteBodies` in `wurzel.go`); Option fest `true` (`replay.go`) | rot, rot |
| Z16 | Route liefert auch `2/Password`; `$count` mit `charset=utf-8` | rot, rot |
| Z17 | `Name: "x"` am Request-Aufruf des Record-Dienstes | rot (der Tee weist ab) |

**Raten (§3.10), eigene Messung, je Prozess `-test.count=1`, Fänger `TestTeeTextGleichzeitig` und `TestTeeReichtZieleDurch`:**

| Mutation | `-race` | ohne `-race` |
|---|---|---|
| `trifft` als Paketvariable (die genannte) | rot in 20 von 20 | rot in 20 von 20 |
| `mPos` und `mAbweich` als Paketvariablen | rot in 20 von 20 | rot in 20 von 20 |
| `rest`, `lo`, `hi` (Zustand der UTF-8-Prüfung) als Paketvariablen | rot in 0 von 20 | rot in 0 von 20 |
| `rest`, `lo`, `hi` als Paketvariablen, ganzes Paket `capture` | rot in 0 von 3 | nicht gemessen |

Z15, Option fest `true`: `TestTextRoundTrip` und `TestDurchstichM2` rot in 15 von 15 Prozessen, ohne Mutation 0 von 15 (ohne `-race`). Die Zahlen des Plans (30 von 30, Basis 0 von 30) sind für die genannten Mutationen vereinbar mit der eigenen Messung.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Der Fänger der Nebenläufigkeit führt nur ASCII-Inhalte. Macht eine Änderung `rest`, `lo` und `hi` des Wert-Automaten zu Paketvariablen, bleibt `TestTeeTextGleichzeitig` in 20 von 20 Prozessen mit und 20 von 20 ohne Detektor grün, ebenso das ganze Paket `capture` unter `-race` (3 von 3); zwei gleichzeitige Aufnahmen mit Mehrbyte-Folgen könnten sich dann gegenseitig als `utf8` verwerfen oder ungültige Folgen durchlassen. | `AGENTS.md` §3.10 (Nebenläufigkeit rot in jedem Lauf), §3.15 (jede Form der Eingabe) | `src/hexagon/services/record/capture/text_test.go` · `roh := "wert-" + strconv.Itoa(i) + strings.Repeat("x", 1+i)` | ja — Mutation wie beschrieben, `make test-race` | Nebenläufigkeits-Fänger führt nur eine Form der Eingabe |
| F-2 | MEDIUM | Die Zusagen „nie gepuffert“ (`SPEC-118` Nr. 11 b) und „Speicher hängt nicht von der Länge des Bodys ab“ (Nr. 6) haben für Text keinen Fänger: puffert der Treffer-Zweig den Inhalt (verwirft ihn nur beim Schreiben), bleibt jeder Test grün; `TestBodySpeicher` misst nur JSON, und die Tabelle Z1 bis Z17 führt keine Zeile dafür. Szenario: ein `$value` von `MAX_CAPTURE_BODY` (bis 1 GiB) liegt im Speicher des Proxys. | `AGENTS.md` §3.15 (Zusage ohne Fänger), `SPEC-118` Nr. 6 und 11 b | `src/hexagon/model/redaktor_text.go` · `if !w.mAbweich && w.mPos < len(Marker) && c == Marker[w.mPos] {` | ja — Mutation „`w.roh = append(w.roh, c)` vor dieser Zeile“ bleibt grün | Zusage ohne Fänger (Speichergrenze des neuen Typs) |
| F-3 | LOW | Plan §6 führt „die Methode ist gleichgültig“ für den Aufruf im Record-Dienst; nur `GET` und `HEAD` laufen. Die Mutation `Zaehlung: zaehlung && a.Methode == "GET"` und eine Kleinschreibung des Pfads an der Aufrufstelle bleiben grün (`Wert` nur bei `GET` wird über den Zustand `Name` ohne `Wert` rot, nicht über die Zusage). Szenario schwach, da `PUT`/`DELETE` auf `$value` in der Regel `204` liefern. | `AGENTS.md` §3.15 (Fänger je Aufrufstelle und Form) | `src/hexagon/services/record/ende.go` · `zaehlung, wert, name := model.TextZiel(x.e.Path)` | ja | Fänger führt nur eine Form der Eingabe (Methode, Schreibweise) |
| F-4 | LOW | Kommentare in den neuen Tests nennen Plan-Zeilen (`Z1 bis Z11 und Z9`, `Z12`, `Z17 und Z13`) ohne Slice-Kennung; Z9 steht doppelt, und nach `done/` und Archiv löst die Nummer allein nicht mehr auf. | `AGENTS.md` §3.7 (Herkunft in **ein** auflösbares Feld) | `src/hexagon/model/redaktor_text_test.go` · `// Die Tests dieser Datei fangen die Zusagen Z1 bis Z11 und Z9` | nein — kein Gate | Plan-Zeilen-Nummer ohne auflösbare Herkunft im Code-Kommentar |
| F-5 | INFO | Rückgabe 1 (`SPEC-118` Nr. 11 b, Marker erst am Stromende und nur bei Erfolg): Der Code lässt bei Treffer vor jedem Verstoß das `dst` leer (Nr. 11 b, erster Halbsatz) und schreibt den Marker in `ende()`. Eine Lesart „Marker mit dem letzten gelesenen Byte“ gäbe bei `zu_gross` und bei einem Lesefehler nach vollständigen Daten einen Marker im `dst` (vgl. Nr. 9 „bei `eingabe` für alle gelieferten“). Der Aufrufer verwirft `dst` in beiden Fällen (Nr. 9), nur der Wortlaut ist offen. Nicht bestätigt; Entscheidung beim Architect. | `SPEC-118` Nr. 9, 11 b | `src/hexagon/model/redaktor_text.go` · `w.aus.text(Marker)` | nein | Auslegung offen, Adresse Architect |
| F-6 | INFO | Rückgabe 2 (halbe Mehrbyte-Folge am Ende ist `utf8`): Der Code folgt dem Wortlaut „sonst `utf8`“ für ungültiges UTF-8; ohne Treffer stehen die Bytes der halben Folge im `dst` (Nr. 9 lässt das für eine Mehrbyte-Folge vor dem Verstoß zu). In sich schlüssig und von `TestTextWertUTF8` samt Generator gehalten; nicht bestätigt. | `SPEC-118` Nr. 4, 9, 11 b | `src/hexagon/model/redaktor_text.go` · `func (w *wertRedaktor) ende() *UnsicherFehler {` | nein | Auslegung offen, Adresse Architect |
| F-7 | INFO | Rückgabe 3 (`TextZiel`): `/Name/$value` ist `Wert` mit `Name` (Suffix `/<Name>/$value`, der Schrägstrich gehört dem Muster), `x/$value` ohne Schrägstrich hat kein Ziel, die Query bleibt Sache des Aufrufers; der Record-Dienst übergibt den Pfad ohne Query, der Adapter bildet ihn aus `RequestURI` mit den Escapes. Daneben beobachtet: `/Products/1/$value` (Schlüssel als Segment) ergibt `Wert` mit `Name` `1`, also ein Ziel für einen Entity-`$value`, den §1 des Plans ausschließt; und ein zweifach kodierter Name (`%2570assword`) trifft das Muster nicht (einmal dekodiert, wie Nr. 11 d). Beides folgt dem Wortlaut; Entscheidung beim Architect. | `SPEC-118` Nr. 11 d | `src/hexagon/model/redaktor_text.go` · `i := strings.LastIndex(rest, "/")` | nein | Auslegung offen, Adresse Architect |
| F-8 | INFO | Rohbytes (§3.18): Die gemeldete Datei und alle neuen und geänderten Textdateien des Diffs sind mit `LC_ALL=C grep -P` gegengelesen: nur `ä`, `ö`, `ü`, `ß` in Kommentaren und Meldungen, kein BOM, kein NBSP, kein Steuerzeichen, keine Folge aus Backslash, `u` und vier Hex-Ziffern; Testdaten ab `0x80` stehen als `\x`-Folgen. Die Zählung der Mutationen im Plan (Summe der Zellen etwa 111) ist plausibel; die Marke wurde nachgemessen: `git diff --numstat 73e140b..HEAD` über die vier Pfade ergibt 1420 (1395 und 25), unter der Marke 2500. | `AGENTS.md` §3.18, `SPEC-128` | alle Dateien des Diffs | ja — `make groessenmarke` | — |
| F-9 | INFO | Zwei Mutationen sind gleichwertig und nicht als Fehler gemeint: Das Weglassen von `schreibeVoll` in Zählung und Wert ohne Treffer bleibt grün, weil `lauf` den Puffer nach jedem Read leert (Obergrenze ein Read). Dazu `slices.Clone(e.ContentType)` zu `e.ContentType` in `bodyEingang()`: grün, ein Fänger für die Aliasierung fehlt, die Zusage stammt aber nicht aus diesem Slice. | `AGENTS.md` §3.15 | `src/hexagon/services/record/capture/pipeline.go` · `ContentType: slices.Clone(e.ContentType)` | ja | gleichwertige Mutation / Bestand |

## Antworten auf die Fragen des Auftrags

- **Zeilen Z1 bis Z17:** je mindestens eine Mutation aufgelöst, 16 von 17 rot an der Zeile selbst; Z11 hat keinen neuen Code. Offen sind nur F-1 und F-2 (MEDIUM) und F-3 (LOW).
- **Allaussagen (§3.17):** Zeichensatz (46 Formen mal 2 Ziele), jedes Byte der Zählung (768), UTF-8 über 232 448 Läufe mit dem Orakel `utf8.ValidString`, Pfadformen (41 370) sind Schleifen; sieben Mutationen an Elementen, die der Plan nicht nannte, wurden gefangen. Die Klasse „Zustand des Wert-Automaten unter Nebenläufigkeit“ ist nicht gedeckt (F-1).
- **Nebenläufigkeit Z13/Z15:** Z13 fängt die genannte Mutation (`trifft`) und `mPos`/`mAbweich` in 20 von 20 Prozessen, mit und ohne Detektor (Wirkungs-Fänger); nicht den UTF-8-Zustand (F-1). Z15 mutiert keine Datenrennen, Rate 15 von 15 gegen 0 von 15.
- **Drei Rückgaben an den Architect:** nicht bestätigt, Code gegen den Wortlaut beurteilt: F-5, F-6, F-7. Kein Befund über HIGH; kein Geheimnis erreicht eine Senke (Treffer schreibt keinen Rohbyte, auch nicht bei Verstoß).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/redaktor.go` (Ziele, `Gueltig`, Typ *Text*, Grund `charset`) | geprüft, ohne Befund |
| `src/hexagon/model/redaktor_text.go` (`TextZiel`, Zeichensatz, Zählung, Wert, UTF-8) | geprüft, Befunde F-1, F-2, F-5 bis F-7 |
| `src/hexagon/services/record/ende.go`, `capture/` (Durchreichen, Abweisung im `Tee`, Grund `charset`) | geprüft, Befund F-3 und F-9 |
| `src/hexagon/services/replay/` (nur Test Z14) | geprüft, ohne Befund |
| `tests/integration/` (Referenzdienst, M2-Fixture, Round-Trip) | geprüft, ohne Befund |
| Plan (§1 bis §6, Belege, Raster, offene Rückgaben mit Adresse) | geprüft, ohne Befund außer F-4 |
| Suppression, `//nolint`, `d-check:ignore` im Diff | geprüft, ohne Befund |

## Summary

Kein HIGH. Zwei MEDIUM (Z13-Fänger führt nur ASCII; Speicherzusage für Text ohne Fänger), zwei LOW, vier INFO (drei Rückgaben an den Architect, eine Messung). Der Code folgt Plan, `SPEC-118` Nr. 11 und `SPEC-119` Nr. 2; die Tabelle ist bis auf die genannten Lücken durch Mutationen gedeckt.
