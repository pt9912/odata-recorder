# Review-Report: slice-v2-query-filter (Teil B) — 2026-11-05

**Review-Art:** Code (gegen Plan, Spezifikation, ADR und Hard Rules)

**Gegenstand:** `slice-v2-query-filter`, Zweig `slice-v2-query-filter`, `80d7910c..8a316f0` (Basis des Blocks `80d7910c`); gemessen 2351 Zeilen über `src/adapters/driven/odata/urlparserv2/` (Marke 4600), Produktcode `filter.go` 356 und `filter_token.go` 113 Zeilen.

**Skill:** `.harness/skills/reviewer.md` @ `8a316f0`
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-05

**Eingangs-Kontext:**

- `slice-v2-query-filter` (Plan, §1, §3 mit 37 Zeilen, §6 Raster und Zusatztabelle)
- [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) (kein V4-Import), `Accepted`
- `LH-FA-17`, `LH-QA-05`, `LH-RB-01`
- `SPEC-173` Nr. 5 (Berichtigung L6 und Schärfung vom 2026-11-05), `SPEC-164` Nr. 1 bis 4 und 6 bis 9, `SPEC-170` Nr. 3, `SPEC-174`
- `AGENTS.md` (Hard Rules §3.7, §3.9 bis §3.18)

Lauf: `make gates` im Worktree: keine Fehlerzeile im Log, `groessenmarke slice-v2-query-filter.md: ok, Messung 2351 von Marke 4600`, `a-check` 0 Befunde, `d-check` 0 Befunde, kein Datei-Diff danach. Zusätzlich 56 Mutationsläufe auf Wegwerfkopien (Image `odata-recorder:test`, `--memory=4g --cpus=4`, `-count=1`; 5 davon nicht gewertet, siehe unten), 2 Wettlauf-Mutationen im Image `odata-recorder:race` und zwei Differenz-Proben. Nicht geprüft: Verifier-Fragen (DoD), Merge-Verträglichkeit mit dem neueren `main`.

Rohbytes (`AGENTS.md` §3.18): gegengelesen mit `LC_ALL=C grep -P` wurden `filter.go`, `filter_token.go`, alle `fl_*_test.go`, `literale_test.go`, der Plan `slice-v2-query-filter.md` und `spec/spezifikation.md`. Kein BOM, kein NBSP, kein Zeilentrenner, kein NUL, keine Folge Backslash-`u`-Hex im Quelltext; Nicht-ASCII-Bytes stehen nur als Umlaute und `…` in Kommentaren und Meldungstexten. Die Zeichen jenseits von ASCII in den Eingaben entstehen aus Bytewerten (`\xc2\xa0`, `flRune(0xE4)`).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Ein Wortlauf mit `$`, `@` oder einem unzulässigen Zeichen unmittelbar vor `(` ist ohne Fänger. Zwei Mutationen bleiben grün: die Aufruf-Prüfung hinter die Prüfung von `$`/`@` verschoben, und der Aufruf nur für Läufe, die nach Entfernen von `/` und `.` ein einfacher Name sind. Szenario: `$x(1)`, `@p(1)`, `a#(1)` müssen `filter` (`501`) sein (`SPEC-164` Nr. 4 c 1 geht Nr. 4 c 2 bis 4 voraus), die Mutation lieferte `wert` (`400`). Zeile 23 führt 33 gültige Namen × 5 Formen, Zeile 36 nur gültige Namen. | `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe), `SPEC-164` Nr. 4 c | `src/adapters/driven/odata/urlparserv2/filter.go` · `case t.vorKlammer:` | ja, 2 Mutationen auf Wegwerfkopie, `go test` grün | Aufruf-Vorrang vor Namensprüfung nur mit gültigen Namen gefangen |
| F-2 | MEDIUM | Die Form von V4 für Datum, Zeitpunkt, Uhrzeit und Guid ist Position für Position ohne Fänger. Grün bleiben: `s[7] == '-'` entfernt, `s[4] == '-'` entfernt, Ziffernprüfung der Stellen 0 bis 4, 5 bis 7, 8 bis 10 entfernt (je einzeln), bei der Uhrzeit die Ziffern an Stelle 0 bis 2 und 3 bis 5 entfernt (je einzeln), die Hex-Prüfung in `guidForm` entfernt, die Länge von `guidForm` von 36 auf 35 bis 40 geweitet (9 Mutationen). Szenario: `2020-01x01`, `20x0-01-01`, `2020-xx-01`, `1x:30`, `12:xx` und ein Lauf mit Ziffer am Anfang, der die Guid-Form nur beinahe hat (35 Zeichen, `0zeebc99-…`), müssen `literal` sein (`SPEC-173` Nr. 5 *Literal-Lauf ohne Präfix* c und d; das Beispiel mit 35 Zeichen steht wörtlich in `SPEC-164` Nr. 4 b); die Mutationen lieferten `wert`. Zeile 19 führt je Form zwei bis vier Texte, die jeweils an der Grenze nicht trennen. Gefangen ist der Trenner der Guid-Form an Stelle 23 (`u_guid_i23`) und die Form im Ganzen; die Prüfung `s[2] == ':'` der Uhrzeit fuhr der Reviewer nicht. | `AGENTS.md` §3.15, §3.17 (Klasse der Zusage: jede Position der Form) | `src/adapters/driven/odata/urlparserv2/filter.go` · `func formVonV4(s string) bool` und `func guidForm(s string) bool` | ja, 9 Mutationen auf Wegwerfkopie grün | Positionen der Form von V4 als Auswahl statt als Schleife über alle Elemente |
| F-3 | MEDIUM | `INF` und `NaN` sind nur in dieser Schreibweise `literal_typ` (`SPEC-164` Nr. 4 c 7: „genau so“); die Mutation `asciiKlein(s) == "inf" \|\| … "nan"` bleibt grün. Szenario: `Id eq inf` und `nan eq 1` sind gültige Eigenschaften, die Mutation lieferte `501` mit `literal_typ`. Die Tabelle führt `INF`, `NaN`, `-INF` und `-inf` (dort Minus), aber keinen Namen `inf`, `Inf`, `nan`, `NAN`. | `AGENTS.md` §3.15 (beide Seiten der Grenze, Fall „gleich“) | `src/adapters/driven/odata/urlparserv2/filter.go` · `if s == "INF" \|\| s == "NaN" {` | ja, Mutation auf Wegwerfkopie grün | Schreibweise der Wörter `INF`/`NaN` ohne Fänger |
| F-4 | MEDIUM | Die Zusage „ohne Beachtung der ASCII-Schreibweise“ (`SPEC-164` Nr. 3, Zeile 6 der Tabelle) ist nur mit ASCII-Eingaben gefangen. Die Mutation `strings.ToLower` statt `asciiKlein` für die Operatorwörter bleibt grün. Szenario: `Id İn 1` (U+0130, in Bytewerten) muss `wert` sein (`İ` ist kein Schlüsselwort), die Mutation lieferte `filter` (`501`), weil Go U+0130 zu `i` kleinschreibt. Die Bitmasken-Schleife erzeugt nur ASCII-Schreibweisen. | `AGENTS.md` §3.15, §3.17 | `src/adapters/driven/odata/urlparserv2/filter_token.go` · `t.klein = asciiKlein(t.text)` | ja, Mutation auf Wegwerfkopie grün (zweimal gefahren, beide grün) | Fänger der Schreibweise führt nur die ASCII-Menge |
| F-5 | MEDIUM | `SPEC-173` Nr. 5 widerspricht sich in einer Nummer. Der unveränderte Satz „Ein Literal-Lauf, der in der Form von V4 als Guid, Datum, Zeitpunkt oder Uhrzeit **lesbar ist** (ohne Präfix), ist `wert`“ steht neben dem neuen Absatz *Literal-Lauf ohne Präfix* (c): „Ein Lauf, der der Form von V4 nach **beginnt**, ist `wert` … der Rest und die Bereiche werden nicht geprüft“. Für `2020-13-45` und `12:30x` sagt der erste Satz `literal`, der zweite `wert`; die Historie vom 2026-11-05 nennt die Schärfung „Prüfung der Form statt der Lesbarkeit“, der alte Satz wurde nicht angepasst. Der Code folgt dem neuen Absatz; ein Leser des ersten Satzes (Nachfolge-Slice, Handbuch, Verifier) erwartet das Gegenteil. Zur Rolle des Architect. | `SPEC-173` Nr. 5, `AGENTS.md` §3.9 (wer eine Aussage berichtigt, sucht alle Stellen) | `spec/spezifikation.md` · `Ein Literal-Lauf, der in der Form von V4 als Guid, Datum, Zeitpunkt oder Uhrzeit lesbar ist (ohne Präfix), ist` | nein, kein Gate liest den Widerspruch | Stale Satz neben geschärfter Regel in derselben Nummer |
| F-6 | LOW | Die Mutation „Aufruf und Pfad vertauscht“ (Zeile 26) ist ein äquivalenter Mutant (beide Verstöße ergeben `KompatFilter`), die Mutation „`Code()` vertauscht“ (Zeile 27) ist dagegen fahrbar und rot: `CodeE6006` durch `CodeE6005` in `Code()` des Modells, rot in `TestFlFehlerwert`, `TestFlAufruf`, `TestFlAusserhalbAllaussage`, `TestFlDekodierung`, `TestFlGenerator`, `TestFlKommentarformen` und weiteren. §6 nennt beide „nicht gefahren“ und gibt für „äquivalent“ keine Folge an, das akzeptierte Negativ steht nicht im Block *Akzeptierte Negative* von §3. | `AGENTS.md` §3.15 (akzeptiertes Negativ nennt Grund und Folge), §3.9 | `docs/plan/planning/done/slice-v2-query-filter.md` · `sind nicht gefahren.` | nein | Zeile „nicht gefahren“ stimmt für `Code()` nicht, Negativ ohne Folge |
| F-7 | LOW | Die Kommentarzeile von `TestFlGleichzeitig` beschreibt die abwesende Fassung im Konjunktiv: „ein Tiefenzähler im Feld der Instanz lieferte dem einen das Ergebnis des anderen“. Es ist die Beschreibung der Mutation, die der Test fängt, kein Zustand des Codes. Grenzfall gegen die HIGH-Regel „Kommentar beschreibt die verworfene Alternative“; als LOW gemeldet, weil der Satz die Wirkung benennt, die der Fänger zählt, und der Rest des Kopfes (`Der Fänger zählt die Wirkung …`) im Indikativ steht. | `AGENTS.md` §3.7, §3.11 | `src/adapters/driven/odata/urlparserv2/fl_dienst_test.go` · `Tiefenzähler im Feld der Instanz lieferte dem einen das Ergebnis des anderen` | nein | Kommentar im Konjunktiv über die Mutation |
| F-8 | LOW | Der Kopf des Plans nennt „`SPEC-164` … Nr. 1 bis 4 und 6 bis 9 gelten wortgleich“; `SPEC-173` Nr. 5 *Der Leser* zählt Nr. 1, 2, 3, 4, 7, 8 und 9 auf, Nr. 6 ersetzt es durch die Zuordnung zu `*URLFehler` und `*KompatFehler`. Kopf und Spezifikation nennen verschiedene Mengen. | `AGENTS.md` §3.9 | `docs/plan/planning/done/slice-v2-query-filter.md` · `Nr. 1 bis 4 und 6 bis 9 gelten wortgleich` | nein | Mengenangabe an einer Stelle des Plans nicht nachgezogen |
| F-9 | INFO | Die fünf Rückgaben des Implementers (Zeilen 17, 27, 28, 13, 9; §6 Ausgang *offen*, Adresse Architect) sind gegen den Code gelesen: Zeile 17 (hinter `X'a'` ein `,` ergibt `literal_typ`) folgt `SPEC-164` Nr. 6 und `SPEC-173` Nr. 5 *Rangfolge im Ausdruck*; Zeile 27 (Marker im Verstoß selbst) und Zeile 13 (`Id eq1 1`) sind Fall-Wahl des Tests ohne Wirkung auf den Vertrag; Zeile 28 (Leser nur für das erste `$filter`) folgt `SPEC-160` Nr. 8 a; Zeile 9 (Mutation „Name klein geliefert“ statt NFC) ist rot in mindestens sechs Tests. Der Code widerspricht der Spezifikation in keiner dieser fünf Stellen. Die Bestätigung der Tabellenzeilen ist die des Architect. | `AGENTS.md` §3.13 | `docs/plan/planning/done/slice-v2-query-filter.md` · `Abweichungen der Tabelle *Zusage · Fänger* von der Spezifikation` | nein | offene Rückgabe, Entscheidung beim Architect |
| F-10 | INFO | Randformen des Lesers gegen das Raster (§3.16): keine Zelle ist im Code entschieden. Belegt durch zwei Differenz-Proben ohne Mutation (siehe unten): 400000 zufällige Folgen aus 29 Bausteinen gegen den Leser von `urlparser` (Ursachen `filter` und `wert_nicht_unterstuetzt` zusammengefasst): 0 Abweichungen; 300000 zufällige Literal-Läufe gegen das Schlüsselliteral `/S(<lit>)`: 192 Abweichungen, alle Läufe der Form `zz:zz…` (Uhrzeit-Form, `literal` im Schlüssel, `wert` im Filter), die `SPEC-173` Nr. 5 (c) verlangt. | `AGENTS.md` §3.16, §3.12 | `src/adapters/driven/odata/urlparserv2/filter.go` | ja (Probe auf Wegwerfkopie, nicht im Zweig) | Randformen im Code, Proben bestätigen den Wortlaut |

Keine HIGH-Findings. Geprüft gegen die HIGH-Liste: [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) (kein Import von `urlparser`, `csdl`, `jsoncodec`; `TestKeinV4Import` deckt die neuen Dateien mit, `make a-check` 0 Befunde), Gate-Suppression (kein `nolint` in den neuen Dateien), Geheimnis in Fehlertext (`kompatFehler` und `fehler` tragen nur Ursache, Ort leer; `TestFlFehlerwert` sucht Marker in `Error()`; Mutation „Ort gesetzt“ rot), Kommentarklassen (Produktcode `filter.go`, `filter_token.go` im Indikativ, jede Funktion mit Zusage oder Kopplung; eine Fundstelle F-7 in einem Test), Zustandsfeld mit Chronik (§4 *Messung beim Handoff* ist eine Messung, keine Chronik), Spec-Straten (Code gegen `SPEC-173` Nr. 5 und `SPEC-164` gelesen; Widerspruch F-5 liegt innerhalb einer Nummer, nicht zwischen den Straten).

## Mutationen des Reviewers (Wegwerfkopie, `go test -count=1` über `urlparserv2`)

Je Zeile der Tabelle *Zusage · Fänger* (nur die, die der Reviewer selbst fuhr; Mutationen des Implementers aus §3 wurden nicht wiederholt):

| Zeile | Mutation (auch an Elementen, die der Plan nicht nannte) | Ergebnis |
|---|---|---|
| 1 | `lt` als `le` gebaut | rot (`TestFlAssoziativitaet`, `TestFlKlammern`, `TestFlKorpus`, `TestFlRangfolge`, `TestFlGenerator`, `TestFlSchluesselwoerter`) |
| 2 | `and` als `or` gebaut; `not` ohne Knoten | rot (je mindestens 6 Tests) |
| 3 | `ne` auf der Stufe der Vergleiche; `le` auf der Stufe von `eq` | rot (`TestFlRangfolge`, `TestFlGenerator`) |
| 4 | `le` rechtsassoziativ (nur dieser Operator) | rot (`TestFlAssoziativitaet`, `TestFlGenerator`) |
| 5 | Klammer liest ihren Inhalt ab `and` | rot (`TestFlKlammern`, `TestFlKorpus`, `TestFlGenerator`, `TestFlWaechter`) |
| 6 | Schlüsselwörter mit `strings.ToLower` | **grün** (F-4) |
| 7 | unbekanntes Token am Platz eines Operators als `filter` | rot (6 Tests) |
| 9 | Name klein geliefert | rot (mindestens 6 Tests) |
| 12 | `-` nur mit ASCII-Ziffer als Literal-Lauf (`.` nicht) | rot (`TestFlMinusLaeufe`) |
| 13 | `tokRand` als Ende; Anfangsleerzeichen angenommen (je einzeln) | rot (`TestFlLeerzeichen`) |
| 15 | V4-Form vor der Länge geprüft | rot (`TestFlLiteralLaenge`) |
| 16 | Rest hinter dem Literal mitgemessen | rot (`TestFlRestZaehltNicht`) |
| 17 | Leerzeichen nicht als Ende der Zeichenkette | rot (6 Tests) |
| 18 | eigene Literal-Kopie im Filter (Guid groß) | rot (`TestFlEchteV2Filter`, `TestFlGenerator`, `TestFlLiteralEnde`) |
| 19 | `INF`, `NaN` je einzeln nicht ausgenommen | rot (`TestFlV4FormenOhnePraefix`, `TestFlLiteraleWieSchluessel`) |
| 19 | Positionen der Form von V4 (9 Mutationen), `inf`/`nan` kleingeschrieben | **grün** (F-2, F-3) |
| 19 | Trenner der Guid-Form an Stelle 23 | rot (`TestFlV4FormenOhnePraefix`) |
| 20 | `null` nur kleingeschrieben | rot (`TestFlNull`, `TestFlKorpus`, `TestFlGenerator`, `TestFlSchluesselwoerter`) |
| 22 | letztes Glied eines qualifizierten Namens ungeprüft; erstes Glied eines Pfads ungeprüft | rot (`TestFlNichtUnterstuetzt`, `TestFlPfad`, `TestFlZeichenAussenGrammatik`) |
| 23 | Aufruf nach `$`/`@`-Prüfung; Aufruf nur für einfache Namen | **grün** (F-1) |
| 24 | `$it`, `$root`, `$this` je einzeln aus der Liste; `$it` ohne Rest; Alias ohne Namensprüfung | rot (`TestFlNichtUnterstuetzt`) |
| 25 | Grenze 31; Tiefe nach Klammer nicht abgebaut | rot (`TestFlTiefe`) |
| 27 | `Ort` im Kompatibilitätsfehler; Kodierung nach der Grammatik; `Code()` vertauscht | rot (`TestFlFehlerwert` und weitere) |
| 34 | Kopie des Textes je Token | rot (`TestFlWaechter`) |
| 35 | `mul` nicht in der Liste (Element, das der Plan nicht nannte) | rot (`TestFlAusserhalbAllaussage`, `TestFlNichtUnterstuetzt`, `TestFlSchluesselwoerter`, `TestFlGenerator`) |

Nicht gewertet: 4 Mutationen mit Fehler im Mutationsmuster (Übersetzungsfehler oder eine Variable im Muster ersetzt), 1 äquivalente (Länge `>= 9` statt `>= 10` in der Datumsform, weil die Ziffernprüfung der Stellen 8 bis 10 die Länge erzwingt).

### Zeilen 32 und 33 (Nebenläufigkeit), Detektor je Prozess

Messender: der Reviewer. Test-Binary einmal übersetzt (`go test -c`, mit `-race` bzw. ohne), danach 10 Prozesse je Zeile und Modus (`-test.run TestFlGleichzeitig -test.count=1`), Image `odata-recorder:race`, `--cpus=4`, Maschine unter Last (Load 8). Unmutierter Zweig mit Detektor: 10 von 10 Prozessen grün, 0 `DATA RACE`.

| Zeile | Mutation | mit Detektor | ohne Detektor (Wirkungs-Fänger) |
|---|---|---|---|
| 32 | Tiefenzähler als Feld der Instanz, je Aufruf auf 0 gesetzt | 10 von 10 rot, 10 von 10 `DATA RACE` | 10 von 10 rot |
| 33 | `map`-Cache in einer Variablen des Pakets, beim ersten Lesen eines Zahl-Laufs geschrieben (Prüfung auf Fehlen, dann Schreiben) | 10 von 10 rot, 10 von 10 `DATA RACE` | 10 von 10 rot (`concurrent map writes`) |

Das stimmt mit den 30 von 30 des Plans überein (kleinere Stichprobe). Der Fänger für Zeile 33 liest Texte ohne frühere Lesung (`flParallelFrisch`); die Cache-Variante, die erst beim Fehlen schreibt, wird gefangen.

### Differenz-Proben (keine Mutation, nicht im Zweig)

1. 400000 Folgen aus 29 Bausteinen (Namen, Zahl, Zeichenkette, Operatoren, Klammern, `foo(`, `a/b`, `$it`, `@p`, `has`, `add`, `#`, `'`, Tabulator, …) gegen `urlparser.NewFilterLeser()` und `urlparserv2.NewFilterLeser()`, Ursachen `filter` und `wert_nicht_unterstuetzt` zusammengefasst: 0 Abweichungen.
2. 300000 zufällige Läufe aus Literal-Bausteinen (Ziffern, `.`, `-`, `+`, `e`, `E`, `d`, `f`, `L`, `M`, `:`, `T`, `Z`, `INF`) als `Id eq <lit>` gegen `/S(<lit>)`: Abweichungen nur `literal` → `wert` für `zz:zz…` (192).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/odata/urlparserv2/filter.go`, `filter_token.go` | gegen `SPEC-173` Nr. 5 und `SPEC-164` gelesen und mutiert; Befunde F-1 bis F-4 (Fänger), sonst ohne Befund; Nebenläufigkeit: kein Zustand der Instanz und kein Paketzustand nach der Initialisierung (Zeilen 32 und 33 mit 10 von 10 Prozessen gefangen) |
| `src/adapters/driven/odata/urlparserv2/fl_*_test.go` | geprüft; Lücken F-1 bis F-4, Kommentar F-7; Rohbytes gegengelesen, Zeichen jenseits von ASCII aus Bytewerten |
| `src/adapters/driven/odata/urlparserv2/literale_test.go` | geprüft, ohne Befund (Listen als Funktionen, Tests von `slice-v2-query` unverändert) |
| `spec/spezifikation.md` (`SPEC-173` Nr. 5, Historie, *Ungeprüft*) | gegen den Code gelesen; F-5 |
| `docs/plan/planning/done/slice-v2-query-filter.md` | §3, §4, §6 geprüft; F-6, F-8, F-9; Größenmarke Messung 2351 von 4600 stimmt mit `git diff --numstat` überein |
| `docs/plan/planning/in-progress/roadmap.md`, `docs/plan/planning/welle-odata-v2.md` | je eine Zeile, geprüft, ohne Befund |
| `src/hexagon/model/` (nur `Code()` mutiert) | ohne Befund |
