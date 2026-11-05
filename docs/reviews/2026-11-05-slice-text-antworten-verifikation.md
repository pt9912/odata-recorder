# Verifikationsbericht: slice-text-antworten — 2026-11-05

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan und DoD, nicht gegen den Diff (Reviewer) und nicht gegen den Bedarf (Validator). Ein Lauf, keine Wiederholungsschleife.

**Gegenstand:** `slice-text-antworten` (Welle `welle-record-replay`, Linie A), Zweig `slice-text-antworten`, geprüfter Stand **HEAD `8081cbf`** (Code `07d780d`, Review `2a23bd0`, Architect `67a9ee1`, Fix-Runde `8081cbf`), Basis des Blocks `73e140b`. Bezug: `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-QA-07`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

**Geprüft gegen:** DoD, §3 (Tabelle *Zusage · Fänger* Z1 bis Z18, *Belege des Implementers*, Raten), §4 (Block, Marke), §6 (Raster, Zusatztabelle, Risiken) des Plans; `SPEC-118` Nr. 1, 3, 11 (mit der Schärfung vom 2026-11-05 zu 11 b und 11 d), `SPEC-119` Nr. 2 und 8, `SPEC-121` Nr. 5, `SPEC-135` Nr. 7 c; `AGENTS.md` §3.10, §3.15 bis §3.18; Review `docs/reviews/2026-11-05-slice-text-antworten-review.md`.

**Methode.** Beleg des Implementers nicht übernommen. Selbst gefahren: `make gates` (ein Lauf, Exit 0), ein frischer `go test -race -count=1 ./src/... ./tests/...` auf einer unveränderten Wegwerfkopie (die Stufen `test` und `race` des `make gates` kamen aus dem Docker-Cache, `CACHED`; der frische Lauf ersetzt den Beleg), 34 eigene Mutationen auf Wegwerfkopien (die Wiederholung im zweiten Modus nicht mitgezählt), Rohbyte-Gegenlesen. Kopien und Skripte im Scratchpad `.../scratchpad/m/` und `.../scratchpad/bin/`, Image `odata-recorder:race` (`golang:1.27.1-trixie`, cgo), `--network none`, `--memory 3g` bis `4g`, Kopie als `/src` eingehängt; im Repo liegt davon nichts.

---

## Zahlen (selbst gemessen)

| Größe | Wert |
|---|---|
| `make gates` auf `8081cbf` | Exit 0 (baseline-verify 54 Dateien, d-check `607 Datei(en), 0 Befund(e)`, groessenmarke, handbuch, lint, zeichen `1192 Datei(en)`, migration-check, schema-check) |
| `make groessenmarke` | `slice-text-antworten.md: ok, Messung 1737 von Marke 2500` (der Plan nennt 1735, siehe V-5) |
| Frischer `go test -race -count=1` über alle Pakete und `tests/` | alle `ok`, kein Datenrennen, kein roter Test |
| Basis der Fänger zu Z13 (unverändert) | ohne `-race` 0 von 20, mit `-race` 0 von 10 Prozessen rot |

## Plan-vs-Code-Diff

| Plan sagt | Code sagt | Urteil |
|---|---|---|
| `model.TextZiel(pfad) (zaehlung, wert bool, name string)` als einzige Stelle der Ableitung; `BodyEingang.Zaehlung/Wert/Name`, `Gueltig()`, Typ *Text*, Grund `charset` | `redaktor_text.go`, `redaktor.go`; Aufrufstelle nur `ende.go` (`model.TextZiel(x.e.Path)`) | konform (`SPEC-118` Nr. 1, 3, 11) |
| Ziele nur für die Antwort, nie für den Request; `Tee` weist Ziel bei `Request` und ungültige Kombinationen mit `ErrUngueltig` ab | `ende.go` setzt `Zaehlung/Wert/Name` nur im `Response`-`Eingang`; `capture/pipeline.go` prüft vor dem ersten Read und vor der Goroutine | konform (`SPEC-119` Nr. 2, `SPEC-121` Nr. 5) |
| Marker erst am Ende des Stroms, bei Verstoß nichts im `dst`; halbe Folge am Ende `utf8` | `wertRedaktor.ende()` schreibt den Marker nur ohne Verstoß; `ende()` meldet `utf8` bei `rest > 0` | konform (Schärfung 11 b) |
| Ziffer am Anfang des dekodierten Namens setzt kein Ziel; ungültige Kodierung: Rohtext | `TextZiel`: `name = roh`, dann dekodiert, dann Ziffernprüfung auf `name[0]` | konform (Schärfung 11 d) |
| Request-Weg unverändert (`SPEC-135` Nr. 7 c), kein Produktcode im Replay | `git diff` über `src/hexagon/services/replay/` nennt nur `text_test.go` | konform |
| Referenzdienst `/Products/$count` als `text/plain`, Route `$value`; M2-Fixture | `tests/integration/referenzdienst_test.go`, `m2_test.go`, `text_test.go`; Tests grün | konform |
| Keine Suppression, keine Plan-Zeilen in Code-Kommentaren | `grep` auf `nolint`, `d-check:ignore`, `Z<N>` im Diff von `src/` und `tests/`: 0 Treffer (F-4 des Reviews behoben) | konform |
| Marke 2500 (Metrik `numstat`, vier Pfade) | Gate 1737 | konform |

## DoD

| Punkt | Zustand |
|---|---|
| Liefer-Punkt 1 (Redaktor, Ableitung, `Tee`) | erfüllt; `SPEC-118` Nr. 1, 3, 11, `SPEC-119` Nr. 2, 8, `SPEC-121` Nr. 5 gegen den Code gelesen |
| Liefer-Punkt 2 (Referenzdienst, M2, Round-Trip) | erfüllt; `TestDurchstichM2`, `TestTextRoundTrip` im frischen Race-Lauf grün (vor/nach dem Wechsel: Beleg des Implementers in §6, selbst nicht wiederholt; die Closure nennt beide Läufe, offen bis dahin) |
| Liefer-Punkt 3 (Zeile · Mutation · roter Test) | erfüllt für Z1 bis Z18 mit den unten genannten Lücken (V-1, V-2); Mutationen ab hier selbst gesehen |
| `make gates` grün | selbst gefahren, Exit 0 |
| Review, Doku-Update, Closure-Notiz, Register, §6-Ausgänge, Paarungen | offen, Sache der Closure; kein Befund gegen die Reihenfolge |

## Auflösung der Fixes gegen die Klasse (`AGENTS.md` §3.17)

### F-1: Zustand des Wert-Automaten unter Nebenläufigkeit

**Klasse (aus dem Wortlaut von `SPEC-119` Nr. 8 und `SPEC-118` Nr. 8/11 b):** jedes Feld des Wert-Automaten und jeder Zeiger, den er hält — `trifft`, `rest`, `lo`, `hi`, `mPos`, `mAbweich`, `aus` (7 Elemente) und jede Teilmenge. Die Tests der Fix-Runde fahren Fänger nach Wirkung (Kanäle), nicht nach Detektor.

Gemessen mit dem übersetzten Testbinary, je Prozess `-test.count=1`, Auswahl `TestTeeText|TestTeeReichtZiele` (für `trifft` `TestTeeTextGleichzeitig|TestTeeReichtZieleDurch`, siehe V-3), Mutation = Paketvariable statt Feld:

| Mutation | ohne `-race` (Prozesse) | mit `-race` (Prozesse) |
|---|---|---|
| `trifft` | 20 von 20 | 10 von 10 |
| `mPos` | 20 von 20 | 10 von 10 |
| `mAbweich` | 20 von 20 | 10 von 10 |
| `mPos` und `mAbweich` (die im Review gemessene Mutation) | 20 von 20 | 10 von 10 |
| `rest` | 20 von 20 | 10 von 10 |
| `lo` | 20 von 20 | 10 von 10 |
| `hi` | 20 von 20 | 10 von 10 |
| `lo` und `hi` | 20 von 20 | 10 von 10 |
| `rest`, `lo`, `hi` (die im Review ungefangene Mutation) | 20 von 20 | 10 von 10 |
| `aus` (Zeiger auf die Ausgabe; vom Implementer nicht genannt) | 10 von 10 | nicht gemessen |
| keine Mutation (Basis) | 0 von 20 | 0 von 10 |

Antwort auf die Frage des Auftrags: Die Mutation „`mPos` und `mAbweich` als Paketvariablen“ ist jetzt **in jedem gemessenen Lauf rot**, mit und ohne Detektor (20 von 20, 10 von 10), allein durch `TestTeeTextMarkerZustand` (Wirkungs-Fänger); die Rate der Messung des Implementers (30 von 30) ist mit meiner vereinbar. Die Klasse ist gedeckt; das Element `aus` ist neu belegt.

### F-2: nie gepuffert / Speichergrenze für Text

**Klasse:** jede Art des Inhalts mal jeder Zweig (Zählung, Wert ohne Treffer, Wert mit Treffer), jedes Byte, das ein Zweig halten könnte.

| Mutation (alle am Code in `redaktor_text.go`, Fänger `TestTextSpeicher`) | Ergebnis |
|---|---|
| Treffer-Zweig hält jedes zweite Byte (nicht genannt) | rot |
| Zweig ohne Treffer hält Bytes ab `0x80` (nicht genannt; der Plan nennt die Variante für den Treffer-Zweig) | rot |
| Zählung hält nur das Byte `9` (nicht genannt) | **grün** |

Die dritte Mutation bleibt grün, weil der Fänger die Zählung mit einem Strom aus lauter `7` füllt: die Klasse „Ziffern“ ist ein Element, nicht die Menge `0` bis `9`. Siehe V-2.

### F-3: Methode × Pfadform an der Aufrufstelle (`ende.go`)

**Klasse:** jede Methode (acht Elemente) × jede Pfadform (sieben, darunter vier Schreibweisen ohne Ziel); `TestTextAntwortMethodeUndSchreibweise` ist die Schleife (56 Fälle).

| Mutation (alle nicht vom Review genannt außer der ersten) | Ergebnis |
|---|---|
| `Zaehlung` nur bei `GET` (Review) | rot |
| `Wert` nicht bei `DELETE` | rot |
| `Zaehlung` nicht bei `TRACE` | rot |
| `Wert` nicht bei `PROPFIND` | rot |
| Pfad kleingeschrieben an der Aufrufstelle (Review) | rot |
| Pfad ohne abschließenden `/` an der Aufrufstelle (nicht genannt) | grün |

Die letzte ist keine Zusage dieses Slice: `TextZiel` selbst bindet das Suffix byte-gleich, und ein Pfad mit abschließendem `/` hat nie ein Ziel; ob der Record-Dienst den Pfad so übergibt, steht in `SPEC-121` Nr. 5 nicht. Kein Befund, nur genannt.

### Ziffernregel in `TextZiel`

**Klasse (`SPEC-118` Nr. 11 d):** jede Ziffer `0` bis `9` am Anfang des dekodierten Namens × Form (roh, kodiert `%3<Ziffer>`, mit Rest, mit Rest ungültiger Kodierung); dazu jede Stelle im Namen (nur die erste zählt) und die Nachbarn der Klasse.

| Mutation | Ergebnis |
|---|---|
| Ziffer `0` ausgenommen, einzeln, dann `1` bis `9` (10 Mutationen) | rot (10 von 10) |
| Prüfung am letzten statt ersten Byte | rot |
| Prüfung nur bei Name mit mehr als einem Byte (`/Products/1/$value` bliebe Ziel) | rot |
| Prüfung am Rohtext statt am dekodierten Namen (`%31`) | rot |
| Prüfung entfällt, sobald der Rohtext ein `%` enthält | rot |
| Obere Nachbar-Grenze: `:` zählt als Ziffer (Plan: genannt) | rot (Plan) |
| **Untere Nachbar-Grenze: `/` (0x2F) zählt als Ziffer** (nicht genannt) | **grün** |

`/` kann nur als dekodiertes `%2F` am Anfang eines Namens stehen: `/P/%2Fabc/$value` hat `Name` `/abc` und ist `Wert`. Die Fänger (`TestTextZiel`, `TestTextZielGenerator`, `TestTextZielRein`) führen diesen Fall nicht; das Orakel des Generators kennt ihn nicht, weil keines der 19 Segmente mit `%2F` beginnt. Siehe V-1.

## Befunde

| ID | Kategorie | Befund | Quelle | Klasse |
|---|---|---|---|---|
| V-1 | MEDIUM | Die Grenze der Ziffernregel hat nur eine Seite: die obere Nachbarin (`:`) ist gefangen, die untere (`/`, nur über `%2F` erreichbar) nicht. Mutation `name[0] >= '/' && name[0] <= '9'` in `model/redaktor_text.go` lässt `TestTextZiel`, `TestTextZielGenerator` und `TestTextZielRein` grün. Wirkung der Mutation: `/P/%2Fabc/$value` verlöre sein Ziel (fail-closed, kein Leck). Die Zusage steht in `SPEC-118` Nr. 11 d; der Plan verspricht „beide Seiten jeder Grenze“ (`AGENTS.md` §3.15). | `AGENTS.md` §3.15, §3.17; `SPEC-118` Nr. 11 d | `src/hexagon/model/redaktor_text_test.go` (`TestTextZiel`, Generator) | Grenze der Ziffernklasse nur an einer Seite gefangen |
| V-2 | LOW | Die Zusage „Speicher hängt nicht vom Inhalt ab“ für die Zählung ist an einem Element der Klasse gefangen (Strom aus `7`); eine Mutation, die nur das Byte `9` hält, bleibt grün. Folge im Betrieb nur bei einer Zählung aus lauter `9`; Hinweis, kein Leck. | `AGENTS.md` §3.15 (jede Form der Eingabe) | `TestTextSpeicher`, Fall „Zählung“ | Menge `0`..`9` nicht als Schleife |
| V-3 | LOW | `TestTeeTextUTF8Zustand` hängt statt zu scheitern, wenn die Aufnahme A vor ihrem Ende nichts schreibt (Mutation `trifft` als Paketvariable, die der Zustand eines vorherigen Tests auf `true` lässt: A schreibt erst am Ende, B wartet auf A). Die Mutation wird rot erst durch die Zeitgrenze von `go test` (Standard 10 Minuten), nicht durch `t.Errorf`. Mit den benannten Fängern der Mutation (`TestTeeTextGleichzeitig`, `TestTeeReichtZieleDurch`) rot 20 von 20, in der Auswahl `TestTeeText|…` hängt sie. Die Zusage Z13 ist gefangen, die Form ist unsauber. | `AGENTS.md` §3.10 (Test zählt, wenn er aus dem benannten Grund rot wird) | `src/hexagon/services/record/capture/text_test.go` · `TestTeeTextUTF8Zustand`, `aDa` | rot nur über die Zeitgrenze |
| V-4 | LOW | In `spec/spezifikation.md` §7 *Historie* steht die neue Zeile vom 2026-11-05 (Schärfung zu `SPEC-118` Nr. 11 b und d) **vor** der Tabellenkopfzeile `\| Datum \| Änderung \|`: die Zeile gehört nicht zur Tabelle (zeigt als Fließtext). Der Architect schreibt die Spezifikation (Commit `67a9ee1`); `make docs-check` meldet es nicht. | `AGENTS.md` §3.9 (Zeilen folgen der Korrektur an ihren Ort), `spec/spezifikation.md` Zeile 2668 | `spec/spezifikation.md` §7 | Tabellenzeile vor der Kopfzeile |
| V-5 | INFO | Der Plan §4 nennt als Messung nach der Fix-Runde 1735 (1710/25); `make groessenmarke` liefert auf `8081cbf` 1737. Unter der Marke 2500, ohne Folge. | `SPEC-128` | Plan §4 |
| V-6 | INFO | Raten des Plans (30 Prozesse je Zeile, Messender: Implementer) stimmen mit der eigenen Messung (20 und 10 Prozesse) überein; die Zahl der Prozesse und der Messende stehen im Plan (`AGENTS.md` §3.10). Das Fehlen von V-3 im Plan ist das einzige, was dort fehlt. | `AGENTS.md` §3.10 | Plan §3 |

Keine DoD-Verletzung. Kein HIGH. Die Rückgaben F-5 bis F-7 des Reviews stehen in §6 mit Ausgang *eingetreten* und der Entscheidung des Architect; Code und Fänger folgen dem geschärften Wortlaut (Marker am Ende, halbe Folge `utf8`, `/Name/$value`, Ziffer am Anfang).

## Rohbytes (§3.18)

Gegengelesen mit `LC_ALL=C grep -P` die Dateien des Diffs `241b3a3..8081cbf` unter `src/`, `tests/`, `spec/spezifikation.md` (nur die Zeilen des Diffs), `docs/reviews/…-review.md` und `docs/plan/planning/in-progress/slice-text-antworten.md`: nur `ä ö ü Ä Ö Ü ß § · × ³ ² é µ` und Anführungszeichen; kein BOM, kein NBSP, keine Zeilentrenner, kein Steuerzeichen, keine Folge aus Backslash, `u` und vier Hex-Ziffern in den neuen Zeilen (die fünf Treffer in `spec/spezifikation.md` stehen in unveränderten Zeilen zu `SPEC-115` und `SPEC-118` Nr. 4). `make zeichen-check` meldet `1192 Datei(en)`, ok. Dieser Bericht wurde nach dem Schreiben auf Rohbytes gegengelesen (siehe Übergabe).

## Zusagen ohne Bestätigung

- Der Lauf von `TestDurchstichM2` „vor und nach dem Wechsel der Fixture“ (§6 Risiko 1) ist Beleg des Implementers; selbst gesehen nur der Lauf nach dem Wechsel (grün, `tests/integration` im frischen Race-Lauf). Die Closure nennt beide Läufe.
- Die Mutationszahl 144 des Plans (111 + 33) ist nicht einzeln nachgezählt; die Stichprobe von 34 eigenen Mutationen bestätigt 31 rot; die drei grünen sind V-1, V-2 und die Randnotiz zu F-3.

## Summary

Kein HIGH, ein MEDIUM (V-1: untere Nachbar-Grenze der Ziffernklasse), drei LOW (V-2, V-3, V-4), zwei INFO. Die Fixes zu F-1 und F-3 decken die Klasse (F-1: 8 Elemente plus `aus`, rot in jedem Lauf, mit und ohne Detektor; F-3: Schleife Methode × Pfadform); F-2 deckt die Klasse bis auf das Element „einzelne Ziffer“ (V-2). `make gates` und der frische Race-Lauf sind grün. Entscheidungs-Konformität zu `SPEC-118` Nr. 1/3/11, `SPEC-119` Nr. 2/8, `SPEC-121` Nr. 5, `SPEC-135` Nr. 7 c und ADR-0001: bestätigt.
