# Verifikations-Report 3: slice-odata-url-pfad — 2026-10-31

**Rolle:** Verifier, frischer Kontext, dritte Runde, gezielt auf das Delta `b43b95c..4000b5c` (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, [SPEC-160](../../spec/spezifikation.md), [SPEC-161](../../spec/spezifikation.md), [SPEC-162](../../spec/spezifikation.md), [SPEC-125](../../spec/spezifikation.md) Nr. 8 b und [AGENTS.md](../../AGENTS.md) §3.10, §3.12, §3.15 bis §3.18). Kein Review, keine Reparatur: Code, Plan und Spezifikation sind von mir nicht verändert; geschrieben ist nur diese Datei.

**Gegenstand:** Zweig `slice-odata-url-pfad`, geprüfter Stand `4000b5c` (Spezifikation `d2db433`, Fix-Runde `3a1a109`, `f05cff4`, Messung `4000b5c`), Vorverifikation `b43b95c` (Report 2: N-1 bis N-6). Arbeitsbaum zu Beginn sauber. Basis der Marke `7535bf7`, Marke 9200.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Methode.** Keine Zahl, Mutation oder Äquivalenzbehauptung des Implementers ist übernommen. Scratchpad `/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/verif-urlpfad-3/` (flüchtig; im Repo liegt nur diese Datei): Kopien des Baums per `git archive HEAD`, Container `odata-recorder:race` (`--network none --memory=4g --memory-swap=4g --cpus=4`, je Gruppe ein `flock /tmp/odata-recorder-heavy.lock`-Aufruf). Mutationen: eigener Mutator (Go-Programm im Container), der pro Operator-Treffer genau eine Ersetzung an einer Kopie macht, dann `go test -count=1 -skip TestPV3` über `urlparser`. Probe-Paket `TestPV3*` (nur im Scratchpad), Orakel aus dem Wortlaut von [SPEC-161](../../spec/spezifikation.md) Nr. 2, 4, 5, 6: `regexp` (`\p{L}`, `\p{Nd}`, ASCII-Klassen), `math/big`, `time`, nie der Parser. **Last des Hosts:** `uptime` zu Beginn 2,26; während der Läufe unter der Sperre 7 bis 10, bei den Raten (`/proc/loadavg` im Container) 15,4 bis 19,6 (Minute 1), bei den Läufen mit `GOMAXPROCS=1` 17,6 bis 18,5; fremde Läufe anderer Rollen.

---

## 1. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| <a id="v3-1"></a>V3-1 | **MEDIUM** | **N-2 ist nicht über die Klasse gedeckt ([AGENTS.md](../../AGENTS.md) §3.17).** Die Zusage „ein Buchstabe ist `L`, eine Ziffer `Nd`, nicht `Nl`, nicht `No`, nicht `M`“ gilt an jeder Stelle, an der ein Name gelesen wird ([SPEC-161](../../spec/spezifikation.md) Nr. 4, 5 c, 7, 6 g; [SPEC-162](../../spec/spezifikation.md) Nr. 3 d, 6, 7). Der Code liest Namen an neun Stellen; `TestUnicodeKlassen` (348) führt vier Orte (Typpräfix, Name der Menge, Name eines Schlüsselteils, Alias). Mutation je Stelle, `nameB` (Mark und `No` als Buchstabe bzw. Ziffer zugelassen) bzw. `nameA` (nur ASCII), alles andere unverändert: **`nameB` bleibt in der ganzen Suite grün** an `$select`-Element (`query.go:174`), `$orderby`-Element (`query.go:206`), Pfad-Glied in `$select`/`$orderby` (`query.go:156`), Glied eines qualifizierten Rest-Segments (`pfad.go:107`), nacktes Wort im Schlüssel (`schluessel.go:87`); **`nameA` bleibt grün** am nackten Wort (`Products(ä)` müsste `key` sein, bliebe `literal`). Rot sind beide Mutationen an Typpräfix (`literal.go:127`), Menge (`pfad.go:53`), Schlüsselname (`schluessel.go:96`), Alias (`query.go:56`, nur `nameB`; `nameA` rot). Der Produktcode ist richtig: meine Probe über **alle Codepunkte ab U+0080** (ohne Surrogate) an allen acht Orten und vier Stellungen findet **0 Abweichungen in 22 238 720 Prüfungen**, dazu Pfade in `$select`/`$orderby` und qualifizierte Rest-Glieder **0 in 3 335 808**; dieselbe Probe wird an den grünen Mutanten rot (`query.go:156`: 7394 Abweichungen, die Suite grün). Der Plan nennt die Menge als „29 Zeichen × Orte × Stellungen“ und zählt die vier Orte; die Klasse „jede Stelle, die einen Namen liest“ steht nicht darin. | `sites.out`, `pn-m4.out`, `probe2-base.out` im Scratchpad | **Implementer** (ein Fall je Stelle × Zeichen-Klasse, als Schleife über die neun Stellen; die Menge aus dem Wortlaut bilden), danach Plan §3 nachziehen |
| <a id="v3-2"></a>V3-2 | LOW | **U+0000 am Anfang einer dekodierten Einheit ist ungefangen** (`dekodieren.go:44`, `strings.IndexByte(s, 0) < 0`). Mutation „`<= 0`“ und Mutation „`< 1`“ bleiben grün (2 von 2). Der Mutant nimmt `/%00` (Ursache `segment` statt `kodierung`), `?x=%00`, `?$top=%00`, `?$select=%00`, `?$filter=%00` (`filter`), `?x=1&%00` an; `TestNul` setzt `%00` an sechs Stellen, nie als erstes Zeichen. Zusage: [SPEC-161](../../spec/spezifikation.md) Nr. 2, [SPEC-162](../../spec/spezifikation.md) Nr. 2 (U+0000 ist `kodierung`). Produktcode richtig (Probe `TestPV3Escape`, 0 Abweichungen am Basisstand, 12 am Mutanten). Außerhalb des Deltas, gefunden mit der Aufgabe „Rest von `dekodieren.go`“. | Mutanten `dekodieren.go:44:55` und `:44:57`; Probe | Implementer (Position 0 und Mitte je Stelle) |
| <a id="v3-3"></a>V3-3 | LOW | **Ein Escape mit einer ungültigen Hex-Ziffer vorn ist nur dann gefangen, wenn das Ergebnis zufällig kein UTF-8 ist** (`dekodieren.go:33`). Mutation „zweites `||` zu `&&`“ und Mutation „`s[i+1]` zu `s[i+2]`“ bleiben grün (2 von 2): `%G1` ergibt das Byte `0xF1`, und `%G0%A4%80%80`, `%G1%80%80%80`, `%G2%80%80%80` usw. ergeben gültiges UTF-8 und werden angenommen (42 Abweichungen je Mutant in `TestPV3Escape`, 0 am Basisstand). Die Spezifikation nennt `%G1` ausdrücklich als `kodierung` ([SPEC-161](../../spec/spezifikation.md) Nr. 2); die Fixture der Tests („`%G1` allein“) gleicht dem Wert der Mutation ([AGENTS.md](../../AGENTS.md) §3.15: „Eine Fixture gleicht nie dem Wert der Mutation“). Außerhalb des Deltas. | Mutanten `dekodieren.go:33:42`, `:33:34`; Probe | Implementer (Fortsetzungsbytes `%80%80%80` hinter jeder Form von ungültigem Escape) |
| <a id="v3-4"></a>V3-4 | LOW | **Das akzeptierte Negativ (5) im Plan sagt Falsches über den Fänger.** Es nennt `csdl/leser_test.go`; diese Datei (23 Zeilen) führt kein Zeichen. Der Fänger liegt in `src/hexagon/ports/outbound/outboundtest/metadata_codec_aufloesen.go:324` und führt `a²` (`No`, ungültig), `a٣` (`Nd`, gültig), `٣a` (Ziffer vorn, ungültig), `a` + U+0301 (`M`, ungültig). Der Plan sagt, `M` sei ungefangen und `Ⅷ` werde geführt; gemessen (Mutation an `csdl/namen.go:33`): `IsMark` als Buchstabe **rot**, `No` als Ziffer **rot**, `IsDigit` zu `IsNumber` **rot**, `Nl` (`Ⅷ`, kein Vorkommen im Repo) als Buchstabe **grün**. Ungefangen ist also nur `Nl`; Folge im Betrieb ist ein Name aus einem `Nl`-Zeichen im Import, ohne Sicherheitswirkung (das stimmt mit der Folge des Plans überein, die Zeichenklasse nicht). Grund und Adresse (kein Folge-Slice) bleiben tragfähig. | `csdl.out` | Planner (Wortlaut von (5): Ort `outboundtest/metadata_codec_aufloesen.go`, ungefangen ist `Nl`) |
| <a id="v3-5"></a>V3-5 | LOW | **N-3: die Entscheidung trägt, aber (3) benennt eine Umgehung nicht und überschätzt die Prüfung.** Gestalt (`TestParserOhneZustand`) und Lint (`gochecknoglobals`) fangen ein zweites Feld und eine Variable des Pakets (W1 20 von 20, W1v: Lint Exit 1 `letzterFilterText is a global variable`). Nicht gefangen: Zustand in einem **Objekt hinter dem Interface-Feld `filter`**, das der Parser selbst baut (`WithFilter` hüllt den Leser in einen `merk`-Typ mit Sperre; Feldzahl 1, Kind Interface; Ergebnis des Aufrufs trägt den Text des zuletzt zurückgekehrten Aufrufs): die **ganze Suite ist grün** (`go test` über `urlparser`), `TestVerschraenktLeser` 0 von 20. Ebenso grün in Suite **und** `make lint`-Stufe: ein Zeiger in `var errLetzter = &zustandFehler{}` (lint Exit 0, `0 issues`). Der zweite Fall steht in (3) („Fehlerwerte `Err…`“), der erste nicht, und der Satz „der Review des Diffs prüft, ob ein Feld oder eine Variable des Pakets dazukommt (beides fängt ein Gate)“ trifft den ersten nicht (weder Feld noch Variable). Die Begründung „ein stochastischer Fänger bliebe auf Hosts mit einem Kern ohne Wirkung“ ist gemessen richtig für **einen** Kern (siehe Abschnitt 5: 0 von 20), nicht für mehrere (20 von 20). Folge im Betrieb stimmt: der Parser ist in diesem Slice unverdrahtet, ein späterer Aufrufer liest Ergebnisse gleichzeitiger Aufrufe. Ein Wirkungs-Fänger mit Rate (16 Goroutinen × 300 Aufrufe, Leser kehrt sofort zurück, jeder Aufruf trägt seinen Text) wäre zulässig und fängt W1, den Wrapper und den `Err`-Zeiger (Letzteres in einem Lauf 45 von 4800). | Abschnitt 5 | Implementer oder Planner (Wortlaut von (3) um den Wrapper ergänzen, oder den Wirkungs-Fänger mit Rate eintragen) |
| <a id="v3-6"></a>V3-6 | INFO | **Negativ (4) hält.** Zuteilung am Basisstand für 32 `$orderby`-Elemente bei n = 7978: 640 Bytes (Faktor 0,1); Schlüssel mit 16 Teilen bei n = 1715: 1904 Bytes (Faktor 1,1), dicht mit `%61` bei n = 2047: 5744 (Faktor 2,8). Eine Kopie des Werts je Element käme auf höchstens 32 · n bzw. 16 · n und bleibt unter 64 · n + 65 536; Grund, Folge (Faktor bis 32 ungemeldet, Zuteilung linear) und Adresse (keine nötig) stimmen. | `alloc.out` | – |
| <a id="v3-7"></a>V3-7 | INFO | **Das Delta ist wahr gegen den Wortlaut.** Produktcode unverändert (`git diff --stat b43b95c..4000b5c -- src ':!*_test.go'` leer; geändert sind Spezifikation, Plan und sieben Testdateien). Der Wortlaut von Nr. 4, Nr. 6 g und die Historie stimmen mit dem Code überein: **0 Abweichungen in rund 25,8 Mio Eingaben** der Probe (Abschnitt 3). Die Aussage der Historie „Klarstellung ohne Änderung des Verhaltens“ ist belegt. | Abschnitt 3 | – |
| <a id="v3-8"></a>V3-8 | INFO | **Mutations-Lauf über `literal.go`, `zeit.go`, `dekodieren.go`, `schluessel.go`: 473 Mutationen, 426 rot, 38 grün, 9 bei `go vet` gescheitert (Abschnitt 4).** Von den 38 grünen sind 34 gleichwertig (begründet in Abschnitt 4), 4 nicht (V3-2, V3-3). Die Mutationen, die Report 2 grün fand, sind rot: Mindestlänge der Uhrzeit, Typpräfix mit `IsNumber`/`IsMark` im Lauf; `+INF`, `nan`, `-NaN` als `literal_typ`, Uhrzeit mit `Z` und mit Versatz, `1e3x`, Exponent ohne Ziffern (Zeichenketten maskiert der Mutator, darum von Hand angewendet) je rot in `TestLiteralFormen` (7 von 7). | `mut-*.out` | – |

**Offen aus Report 2, nicht Auftrag dieser Runde (§3.13):** (g) Mutationsskript außerhalb des Repos, (i) bis (l) stehen mit Ausgang *offen* und Adresse in §6 des Plans; nichts davon ist neu und nichts davon habe ich bewertet.

**Verdikt: NICHT BESTANDEN**, wegen V3-1 (MEDIUM). Es ist eine Fänger-Lücke, keine Abweichung des Produktcodes: der Code erfüllt den Wortlaut an allen geprüften Stellen. V3-2 bis V3-5 sind LOW und blockieren allein nicht. Mit V3-1 behoben (und V3-2, V3-3 mitgezogen, wenn der Implementer die Klasse „Dekodierung“ schon anfasst) wäre kein MEDIUM offen.

## 2. DoD und Gates

- **`make gates`** im Worktree auf `4000b5c`: **Exit 0** (aus Datei `gates.exit`, nicht über `| tail` maskiert). `d-check` 368 Dateien, 0 Befunde. `make groessenmarke` im Lauf: `slice-odata-url-pfad.md: ok, Messung 5555 von Marke 9200`; Messpunkt `zeilen=61 nebenlaeufig=2 fakes=1 produktcode=1700 schaetzung=6790 marke=9200 messung=5555`.
- **Als `CACHED` gelaufene Stufen einzeln ohne Cache nachgefahren** (`docker build --no-cache-filter pruefung,build,test,race,lint --target <Stufe>`, eigene Tags, danach entfernt): `test` Exit 0, `pruefung test: ok` (68,2 s); `race` Exit 0, `pruefung race: ok` (105,6 s); `lint` Exit 0, `0 issues.`. Die Stufe `go build` von `make build` habe ich nicht erneut ohne Cache gefahren (ihr Ergebnis hängt an der Quelle, die im Delta unverändert ist).
- **Messung nachgerechnet:** `git diff --numstat 7535bf7..4000b5c` über die 14 Pfade des Blocks: Einfügungen 5546, Löschungen 9, **Summe 5555**; gleich der Zahl im Plan und der Ausgabe des Gates. Das Delta `b43b95c..4000b5c` in `src` ist +558 und −25 Zeilen (583), davon kein Produktcode. Die Zeilenzahl der Tabelle *Zusage · Fänger* bleibt **61** (`zeilen=61` im Messpunkt, unverändert zur Messung 5022 und zur Setzung der Marke); die Marke 9200 steht bei Schätzung 6790 · 1,25 aufgerundet.
- **Plan-Zahlen gegen Tests:** `TestLiteralFormen` zählt 233 (68 + 51 + 43 + 64 + 7; die Teilsummen im Plan habe ich nachgerechnet: U 21 + 1 + 24 + 8 + 10 + 4, E 24 + 21 + 6, S 3 + 21 + 7 + 10 + 2, W 48 + 16), `TestUnicodeKlassen` 348 (29 × 4 + 29 × 4 + 29 × 2 + 29 × 2), `TestVerschraenktLeser` 14 Läufe der Optionen (7 × 2) neben 32 der Pfade, `TestKeyDoppelt` 55 (9 + 46, darunter sechs Paare); alle Tests laufen grün und prüfen ihre Zählprobe selbst.
- **Rohbytes** (`LC_ALL=C grep -P`): in den sieben geänderten Testdateien steht jedes Zeichen jenseits von ASCII, das ein Wert ist, als Bytefolge (`\xd9\xa3` usw.); rohe Umlaute stehen nur in Kommentaren und Meldungstexten. Keine Folge aus Backslash, `u` und vier Hex-Ziffern in Plan, Spezifikations-Delta und Tests; kein BOM, kein NBSP in Plan und Spezifikation. Diese Datei ist ebenfalls gegengelesen (Abschnitt 7).

Alle Punkte der DoD sind `[ ]`, §7 ist „Wird bei der Closure gefüllt“ (Closure ist nicht Auftrag dieser Runde).

## 3. Probe-Paket (Orakel nie der Parser)

| Probe | Klasse | Läufe | Orakel | Ergebnis am Basisstand |
|---|---|---|---|---|
| `TestPV3Literale` | [SPEC-161](../../spec/spezifikation.md) Nr. 6 a bis g | **242 006** Eingaben: 35 Basisformen (alle Literalarten, Uhrzeit, Exponent, `INF`, Typpräfixe) × jede Stelle × 44 Ersatz- und Einfügezeichen (42 verschiedene) (Ziffern, `e`, `E`, `Z`, `T`, Trennzeichen, Vorzeichen, Leerraum, `ä`, `٣`, `²`, `Ⅷ`, U+0301, `ſ`, Kelvin-Zeichen, `ı`, U+FF11, kyrillisches `е`), Streichen, Kürzen, Vertauschen; Uhrzeit systematisch (15 Stunden × 9 Minuten × 10 Sekundenformen × 19 Bruchteile × 8 Anhänge), Exponent systematisch (17 Mantissen × 6 `e` × 6 Vorzeichen × 10 Ziffernfolgen), alle 2^n Schreibweisen von `inf`, `nan`, `true`, `false`, `null` mit Vor- und Nachsatz | Regex aus dem Wortlaut, `math/big` (Zahl), `time` (Datum, Zeitpunkt, UTC-Bereich) | 0 Abweichungen in Ursache, Art und Text |
| `TestPV3Unicode` | [SPEC-161](../../spec/spezifikation.md) Nr. 4, 6 g, [SPEC-125](../../spec/spezifikation.md) Nr. 8 b | **22 238 720**: jeder Codepunkt U+0080 bis U+10FFFF (ohne Surrogate) × 4 Stellungen (allein, erste, innere, letzte) × Ort (Typpräfix, Name der Menge; in erster und innerer Stellung auch Schlüsselname, Alias, `$select`, `$orderby`, Rest-Segment, nacktes Wort) | `\p{L}`, `\p{Nd}`, `_` (Regexp) | 0 Abweichungen |
| `TestPV3NamenPfade` | dieselbe Klasse an Pfaden | **3 335 808**: jeder Codepunkt × `$select=p/a·b`, `$orderby=p/a·b`, `/S/Ns.a·b` | wie oben | 0 Abweichungen |
| `TestPV3Namenlaenge` | Name 128/129 Zeichen, Literal 512 Bytes | Namen aus 1-, 2-, 3-, 4-Byte-Zeichen und `_`, ASCII, `٣` an Länge 126 bis 130; Literale mit 510 bis 513 Bytes und 254 bis 257 Zeichen zu 2 Bytes | Regexp, `utf8` | 0 Abweichungen |
| `TestPV3Nackt` | [SPEC-161](../../spec/spezifikation.md) Nr. 5 c | 42 unbenannte Wörter (alle Schreibweisen von `true`/`null`, `INF`/`NaN` genau, `inf`, `-INF`, Unicode) und als Wert eines benannten Teils | Regel von Nr. 5 c von Hand | 0 Abweichungen |
| `TestPV3Escape` | [SPEC-161](../../spec/spezifikation.md) Nr. 2 | 1463: jede Paarung von 17 Zeichen als erste und zweite Ziffer nach `%`, in fünf Kontexten (allein, mit `%80%80%80`, mit `%A4%80%80`, mit Vor- und Nachsatz); NUL an 18 Stellen | – | 0 Abweichungen |

## 4. Mutationen

**Lauf über vier Dateien.** Mutator-Operatoren: `<=`↔`<`, `>=`↔`>`, `==`↔`!=`, `&&`↔`||`, `<`/`>` zu `<=`/`>=`, jede Zahl ±1, jedes Zeichen-Literal um eins, Verneinung entfernt, `true`↔`false`; außerhalb von Kommentaren und Zeichenketten. Je Mutation eine frische Ersetzung, `go test -count=1 -skip TestPV3 ./src/adapters/driven/odata/urlparser/`.

| Datei | Mutationen | rot | grün | `go vet`/Aufbau |
|---|---|---|---|---|
| `literal.go` | 111 | 102 | 3 | 6 |
| `zeit.go` | 236 | 207 | 27 | 2 |
| `dekodieren.go` | 81 | 72 | 8 | 1 |
| `schluessel.go` | 45 | 45 | 0 | 0 |
| zusammen | **473** | **426** | **38** | **9** |

Die neun Fälle der letzten Spalte: drei sind Artefakte des Mutators (`'9'` zu `'10'`, keine gültige Zeichenkonstante), sechs scheitern an `go vet` (ein `||` zwischen sich ausschließenden `!=`-Vergleichen u. a.); mit `-vet=off` wiederholt sind alle sechs **rot**.

**Die 38 grünen Mutationen**, geprüft von Hand und, wo nicht gleichwertig, mit der Probe:

| Stelle | Anzahl | Urteil |
|---|---|---|
| `zeit.go` Rückgabewerte im Fehlerpfad (`return 0, 0, 0, false` u. a., `0` zu `1`) | 18 | gleichwertig: der Aufrufer wertet die Werte bei `ok == false` nie aus |
| `zeit.go:34` (`||` zu `&&` zwischen `!o1`, `!o2`, `!o3`; `tag < 1` zu `tag < 0`) | 3 | gleichwertig: ein Fehlschlag von `ziffernWert` liefert 0, `jahr < 1`/`monat < 1`/`tag < 1` fangen ihn; Tag 0 fängt `t.Day() != tag` |
| `zeit.go:37` (Stunde, Minute, Sekunde, Nanosekunde von `time.Date` `0` zu `1`) | 4 | gleichwertig: nur `t.Day()` wird gelesen |
| `zeit.go:118` (`len(s) < 17` zu `< 16`) | 1 | gleichwertig: Länge 16 scheitert in `versatzMinuten("")` |
| `zeit.go:133` (Nanosekunde im UTC-Zeitpunkt) | 1 | gleichwertig: `Format` ohne Bruchteil |
| `literal.go:29`, `:36` | 2 | gleichwertig: `'` allein bzw. `i+2 < len(s)` am vorletzten Zeichen enden mit demselben `false` |
| `literal.go:118` (`i > 0` zu `i >= 0`) | 1 | gleichwertig: `zeichenkette` läuft vorher und nimmt jede Form mit `'` am Anfang |
| `dekodieren.go:18` (`return -1` zu `-2`), `:33:57` (zweite Hex-Prüfung auf `s[i+1]`), `:50` (`'Z'` zu `<`, zu `'['`) | 4 | gleichwertig: alle Prüfungen lauten `< 0`; ein ungültiges zweites Zeichen ergibt `1<<4 | -1` = `0xFF`, nie gültiges UTF-8; `Z` und `[` sind für `true`/`false`/`null`/`e`/Hex/`asc`/`desc` ohne Wirkung |
| `dekodieren.go:44:55`, `:44:57` | 2 | **nicht gleichwertig**: [V3-2](#v3-2) |
| `dekodieren.go:33:42`, `:33:34` | 2 | **nicht gleichwertig**: [V3-3](#v3-3) |

**Orte des Namens einzeln** (Mutant ersetzt den Aufruf an genau einer Stelle durch `nameA`, `nameB`, `anfangA`, `anfangB` bzw. ändert die Zeichen-Bedingung von `typPraefix`; Suite wie oben):

| Stelle | `nameA` (nur ASCII) | `nameB` (`M` und `No` zugelassen) | Probe am grünen Mutanten |
|---|---|---|---|
| Typpräfix `literal.go:127` | rot (2 Tests) | rot (1 Test) | – |
| Name der Menge `pfad.go:53` | rot (5) | rot (2) | – |
| Schlüsselname `schluessel.go:96` | rot (6) | rot (1) | – |
| Alias `query.go:56` | rot (1) | rot (1) | – |
| `$select`-Element `query.go:174` | rot (2) | **grün** | rot (40 Meldungen, Ausgabe auf 40 gekappt) |
| `$orderby`-Element `query.go:206` | rot (2) | **grün** | rot |
| Pfad-Glied `query.go:156` | rot (1) | **grün** | rot (7394 Abweichungen in `TestPV3NamenPfade`) |
| Rest-Segment-Glied `pfad.go:107` | rot (2) | **grün** | rot |
| nacktes Wort `schluessel.go:87` | **grün** | **grün** | rot |

## 5. Nebenläufige Zusagen (Raten mit Zahl der Prozesse)

Messender: Verifier. Container `odata-recorder:race` (`--cpus=4`), Binary ohne `-race`; „20 Prozesse“ sind 20 Aufrufe des übersetzten Testprogramms mit `-test.count=1`, „1 × 20“ ein Prozess mit `-test.count=20` (gezählt: Iterationen mit `--- FAIL`). `loadavg` im Lauf 15,4 bis 19,6. Mutanten an Kopien des Arbeitsbaums: W1 (Text von `$filter`, Feld der Instanz, unter einer Sperre nach der Rückkehr des Lesers geschrieben, am Ende von `ParseURL` gelesen), W3 (Zustand aus `$top`), W4 (Zustand aus `$select`), WRAP (Zustand in einem Objekt, das `WithFilter` um den Leser legt; der Parser behält ein Feld vom Kind Interface). `TestPV3Stoch` ist mein Wirkungs-Fänger: 16 Goroutinen × 300 Aufrufe, der Leser kehrt sofort zurück, jeder Aufruf trägt einen eigenen Text, das Ergebnis muss ihn tragen.

| Mutation | `TestVerschraenktLeser` | `TestVerschraenkt` | `TestGleichzeitig` | `TestParserOhneZustand` | `TestPV3Stoch` |
|---|---|---|---|---|---|
| Kontrolle ohne Mutation | **0 von 20**; 1 × 20: 0 | 0 von 20 | 0 von 20 | 0 von 20 | 0 von 20 |
| W1 (nach dem Leser) | 0 von 20; 1 × 20: 0 | 0 von 20 | 0 von 20 | **20 von 20**; 1 × 20: 20 | 20 von 20; 1 × 20: 20 |
| W3 (`$top`) | **20 von 20**; 1 × 20: 20 | 0 von 20 | 20 von 20 | 20 von 20 | 0 von 20 (kein Filter beteiligt) |
| W4 (`$select`) | **20 von 20**; 1 × 20: 20 | 0 von 20 | 20 von 20 | 20 von 20 | 0 von 20 (kein Filter beteiligt) |
| WRAP (Objekt hinter `filter`) | 0 von 20 | 0 von 20 | 0 von 20 | **0 von 20** (Suite insgesamt grün) | 20 von 20; 1 × 20: 20 |
| `Err`-Zeiger als Variable des Pakets | Suite grün | – | – | – | 1 Prozess: 45 von 4800 Aufrufen fremd; `make lint`-Stufe Exit 0 |
| W1 mit `GOMAXPROCS=1` (`--cpus=1`, `loadavg` 17,6 bis 18,5) | – | – | – | – | **0 von 20**; 1 × 20: 0 |
| WRAP mit `GOMAXPROCS=1` | – | – | – | – | **0 von 20**; 1 × 20: 0 |
| Kontrolle mit `GOMAXPROCS=1` | – | – | – | – | 0 von 20 |

Lesart. Die Kontrolle ohne Mutation ist 0 von N in allen Fängern: keine Flakiness. W3 und W4 sind von `TestVerschraenktLeser` erzwungen (beide Aufrufe halten im Leser), unabhängig vom Zufall. W1 fängt allein die Gestalt; sie ist umgehbar (WRAP, `Err`-Zeiger, [V3-5](#v3-5)). Ein Wirkungs-Fänger fängt W1 und WRAP in jedem Prozess bei mehreren Kernen, bei einem Kern in keinem: das trägt den Grund von (3) nur für Hosts mit einem Kern. W1v (Variable des Pakets) ist von `make lint` gefangen: Stufe `lint` Exit 1, `query.go:265:5: letzterFilterText is a global variable (gochecknoglobals)`, 1 Befund.

## 6. Prüfung der Entscheidungen des Architect und der Negative

- **Klasse von Buchstabe und Ziffer** ([SPEC-161](../../spec/spezifikation.md) Nr. 4, 6 g): der Wortlaut trägt die Entscheidung ohne Blick in den Code (Kategorie `L`, `Nd`; nicht `Nl`, nicht `No`, nicht `M`; Verweis auf [SPEC-125](../../spec/spezifikation.md) Nr. 8 b, der dieselbe Klasse nennt). Der Code liest sie so (Abschnitt 3). Offen nur die Abdeckung durch Fänger ([V3-1](#v3-1)).
- **Nr. 6 g (Uhrzeit, Exponent, Sonderwerte)**: aus dem Wortlaut gebildet, ohne den Code: Stunde 00 bis 23, Minute und Sekunde 00 bis 59, Bruchteil 1 bis 12 Ziffern, ASCII-Ziffern, kein Versatz; Exponent `-`? Ziffern (`.` Ziffern)? `e`/`E` `+`/`-`? mindestens eine ASCII-Ziffer; `INF`, `-INF`, `NaN` genau so. Der Code erfüllt jeden Punkt (Probe 242 006 Literale, 0 Abweichungen; Mutationen aus Report 2 jetzt rot, Abschnitt 4).
- **Negative (3), (4), (5)**: (3) hat Grund und Folge, trägt für einen Kern, ist in der Aufzählung unvollständig ([V3-5](#v3-5)); (4) stimmt ([V3-6](#v3-6)); (5) nennt Ort und Inhalt falsch ([V3-4](#v3-4)); alle drei haben eine Adresse oder begründen, warum keine nötig ist.
- **Tabelle *Zusage · Fänger*:** bleibt 61 Zeilen; die Erweiterungen hängen an Zeilen 14, 16, 28, 30, 38 (Plan); ich habe die Zeilen gegen den Vertrag nicht neu aufgelöst, sondern nur die Zeilen der Fix-Runde (Abschnitt 4, 5).

## 7. Nicht gelaufen und Vorbehalte

- **`make gegenproben` und der Stempel:** nicht gefahren, wie beauftragt.
- **Die einzelnen Mutationen des Implementers** (L12a, ZW, W1b, PZ1 usw.) habe ich nicht 1:1 nachgespielt, sondern durch den Mutator über alle vier Dateien und die Orts-Mutanten ersetzt; dass jede im Plan benannte rot ist, folgt daraus nicht Namen für Namen. Der Mutator ersetzt das Skript außerhalb des Repos (Plan §6 (g)) nicht.
- **Nicht gemutet:** `query.go` und `pfad.go` jenseits der Stellen der Namen, `parser.go`, das Modell (`model`), die Fakes. Der Lauf über `literal.go`, `zeit.go`, `dekodieren.go`, `schluessel.go` ist ein Operator-Raster, keine Vollständigkeitsaussage; gleichwertige Mutanten sind von Hand begründet, nicht bewiesen.
- **Raten:** je Mutation 20 Prozesse und ein Prozess mit 20 Iterationen, nicht mit `-race`; der `Err`-Zeiger nur in einem Lauf. Die Last (15 bis 20) stammt von fremden Läufen anderer Rollen; die Kontrolle blieb in jedem Block bei 0.
- **`make build` ohne Cache** nicht erneut gefahren (siehe Abschnitt 2).
- **Gegenlesung dieser Datei** auf Rohbytes (`LC_ALL=C grep -P` auf Backslash, `u` und vier Hex-Ziffern, BOM, NBSP): ohne Treffer; `ä`, `٣`, `²`, `Ⅷ`, `ſ`, `ı` stehen als die erwarteten UTF-8-Bytes in der Datei.

## 8. Register-Kandidaten (nur genannt, nicht geschrieben)

1. `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`: ein weiteres Mal; N-2 wurde über vier Orte gedeckt, nicht über „jede Stelle, an der ein Name gelesen wird“ ([V3-1](#v3-1)).
2. Neu: **eine Fixture, die dem Wert der Mutation gleicht, an einer Dekodierung** (`%G1` allein wird zufällig als ungültiges UTF-8 abgewiesen, [V3-3](#v3-3)).
3. Neu: **ein Gestalt-Fänger für „kein Zustand“ ist durch Zustand hinter einem Interface-Feld umgehbar** ([V3-5](#v3-5)); der Wirkungs-Fänger mit Rate gehört dazu, mit der Grenze „ein Kern“.
