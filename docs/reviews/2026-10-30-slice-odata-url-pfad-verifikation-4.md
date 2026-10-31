# Verifikations-Report 4: slice-odata-url-pfad — 2026-10-31

**Rolle:** Verifier, frischer Kontext, vierte Runde, gezielt auf das Delta `55781f7..dad40cf` (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, [SPEC-160](../../spec/spezifikation.md), [SPEC-161](../../spec/spezifikation.md), [SPEC-162](../../spec/spezifikation.md), [SPEC-125](../../spec/spezifikation.md) Nr. 8 b und [AGENTS.md](../../AGENTS.md) §3.10, §3.12, §3.15 bis §3.18). Kein Review, keine Validierung, keine Reparatur.

**Gegenstand:** Zweig `slice-odata-url-pfad`, geprüfter Stand `dad40cf` (Fix-Runde nach [Report 3](2026-10-30-slice-odata-url-pfad-verifikation-3.md)), Arbeitsbaum zu Beginn sauber. Plan: [slice-odata-url-pfad.md](../plan/planning/done/slice-odata-url-pfad.md). Basis der Marke `7535bf7`, Marke 9200.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Methode.** Keine Zahl, Mutation oder Äquivalenzbehauptung des Implementers ist übernommen. Scratchpad `/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/verif-urlpfad-4/` (flüchtig; im Repo liegt nur diese Datei). Die Stellen im Produktcode habe ich selbst gezählt (Abschnitt 3), das Probe-Paket hat ein eigenes Orakel (`regexp` über `\p{L}`, `\p{Nd}`, eigener Dekodierer mit `strconv.ParseUint`; nie der Parser), die Mutationen sind eigene Änderungen an Kopien (`diff -rq` gegen die Basis: genau eine Datei geändert, sonst nicht gezählt). Container: `golang@sha256:8a5910f3…` bzw. `golang:1.27.1-trixie@sha256:8f58fd67…` (Race), `--network none --memory=4g --memory-swap=4g --cpus=4`, je Gruppe eine Sperre `/tmp/odata-recorder-heavy.lock`, kein Container zurückgelassen.

**Der Produktcode ist unverändert:** `git diff --stat 55781f7..dad40cf -- src ':!*_test.go'` ist leer. Geändert sind Plan und drei Testdateien (`dekodierung_test.go` neu, `unicodeklassen_test.go`, `zustand_test.go`); 4 Dateien, +230 −27.

---

## 1. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| <a id="v4-1"></a>V4-1 | **MEDIUM** | **Die Zusage „Name“ ist an den zehn Orten für `L`, `Nd`, `Nl`, `No`, `M` gedeckt, nicht für ihre ASCII-Anteile und nicht für die Länge in Zeichen ([AGENTS.md](../../AGENTS.md) §3.17).** [SPEC-161](../../spec/spezifikation.md) Nr. 4 nennt „Buchstaben, Ziffern, `_`, 1 bis 128 Zeichen“; die Menge `TestUnicodeKlassen` (29 Zeichen) führt keinen ASCII-Vertreter und keinen Namen aus Mehrbyte-Zeichen an der Grenze. Drei Klassen bleiben ungefangen: (a) **Länge in Zeichen statt Bytes:** die Mutation „`zeichen += utf8.RuneLen(r)`“ bleibt an **allen acht** Stellen mit Längengrenze grün (Menge, Rest-Glied, Alias, Pfad-Glied, `$select`, `$orderby`, nacktes Wort, Schlüsselname); ein Name aus 128 mal `ä` (256 Bytes) würde als zu lang abgelehnt; dieselbe Änderung an der gemeinsamen Funktion `namenAnfang` (`zeichen++` zu `zeichen += utf8.RuneLen(r)`, Abschnitt 4.6) bleibt in der ganzen Suite grün. (b) **`_` an nicht erster Stelle:** die Mutation „`_` nur am Anfang“ bleibt grün beim Alias, beim Pfad-Glied (`$select` und `$orderby`) und beim nackten Wort im Schlüssel (3 von 9 Stellen). (c) **`-` im Namen:** die Mutation „`-` an nicht erster Stelle erlaubt“ bleibt grün beim Rest-Glied, beim Pfad-Glied und bei `$orderby` (3 von 9). Alle 14 Mutationen sind nicht gleichwertig (bei (b) und (c) fängt die Suite die Änderung der gemeinsamen Funktion, 4.6, nicht aber die an einer einzelnen Stelle): das Probe-Paket (Abschnitt 3) meldet je Mutation Abweichungen vom Orakel (2 bis 4 je Stelle bei (b), (c); `TestProbeLaenge` rot bei (a)). Ursache: der Plan führt die Klasse als Unicode-Kategorie, die Allaussage „jedes Zeichen eines Namens, jede Länge“ aber hat keine Schleife über ASCII und über die Grenze 128 mit Mehrbyte-Zeichen. | Abschnitt 4, Tabellen 4.1 und 4.2; Schleife von `unicodeklassen_test.go` | Implementer (Fänger); Planner (Menge im Plan, Zeile 28) |
| <a id="v4-2"></a>V4-2 | LOW | **Es gibt eine elfte Stelle, die einen Namen liest, und ihre Klasse ist nicht gedeckt:** `sonstigerName` (`query.go:54`, zweiter Zweig, herstellerspezifische Option, [SPEC-162](../../spec/spezifikation.md) Nr. 3 e, f: nur ASCII-Buchstaben, Ziffern, `_`, `-`, `.`). Die Zählung „neun Stellen im Code“ stimmt für die Regel von [SPEC-161](../../spec/spezifikation.md) Nr. 4, diese Stelle hat eine andere Regel und steht in keiner Menge. Mutationen „ein Zeichen jenseits von ASCII der Kategorie `Nd` / `Mn` / `No` / `Nl` / `Lo` / `Lm` im Namen erlaubt“ bleiben grün (6 von 6; `Ll`/`Lu` fängt `TestFehlerwert`), ebenso die ASCII-Nachbarn der Bereichsgrenzen (`@` und `` ` `` im Namen, `[` und `{`, `:` als Ziffer, `-` am Anfang: 4 von 4). Das Probe-Paket zeigt 4 bis 564 248 Abweichungen je Mutation. | Abschnitt 4, Tabelle 4.3 | Implementer (ein Ort mehr in `unicodeOrte`, ASCII-Grenzen als Schleife) |
| <a id="v4-3"></a>V4-3 | LOW | **Die Gruppe „sonstige“ der Menge Z deckt acht von neunzehn Kategorien (`Pc`, `Pd`, `Po`, `Zs`, `Cf`, `Sm`, `Sk`, `So`).** Es fehlen elf: `Sc`, `Co`, `Ps`, `Pe`, `Pi`, `Pf`, `Zl`, `Zp`, `Cc`, `Cn` und `Cs` (nicht kodierbar). Die Mutationen „Kategorie `Co` als Buchstabe“ (9 von 9 Stellen), „`Sc`“ (6 von 9) und „`Ps`“ (7 von 9) bleiben grün; die Mutationen bilden nicht alle ein plausibles Versehen ab (kein Bibliotheksaufruf liefert „nur `Co`“), eine Schleife über je einen Vertreter aller 29 kodierbaren Kategorien wäre billig (29 × 10 × 4). Gleiches Muster im akzeptierten Negativ (5): `csdl/namen.go` (dort ist `Mn` gefangen, `Mc`, `Me`, `Nl` als Ziffer und `Sc` nicht). | Abschnitt 4, Tabellen 4.1, 4.4 | Implementer |
| <a id="v4-4"></a>V4-4 | LOW | **Ein gültiges U+FFFD ist ungefangen** (`dekodieren.go:44`). Die Mutation „`utf8.ValidString(s) && … && !strings.ContainsRune(s, utf8.RuneError)`“ bleibt grün; `%EF%BF%BD` ist gültiges UTF-8 und müsste gelesen werden ([SPEC-161](../../spec/spezifikation.md) Nr. 2: „kein gültiges UTF-8“ ist `kodierung`). Das Probe-Paket (Folgen mit `%EF%BF%BD`) meldet die Abweichung. Das Fenster der Klasse „gültiges UTF-8 am Rand“ (U+0080, U+07FF, U+0800, U+FFFD, U+FFFE, U+10FFFF) hat keinen Fänger außer den Zeichen der Unicode-Tabelle. | Abschnitt 4, Tabelle 4.5 | Implementer |
| <a id="v4-5"></a>V4-5 | INFO | **V3-1, V3-2, V3-3 sind gegen ihre Klasse geschlossen.** V3-1: alle 35 Mutationen des Implementers (je Stelle Namensregel nur ASCII, mit `M`/`No`, mit `Nl`, Ziffer am Anfang) selbst nachgefahren, 35 rot, die 36. (`S1` mit Ziffer am Anfang) ist gleichwertig (`typPraefix` erlaubt die Ziffer am Anfang); dazu 54 eigene Mutationen an Elementen, die niemand nannte (`Lo`, `Lt`, `Lm` ausgeschlossen, Buchstabe oder Ziffer nur BMP, Ziffer nur ASCII, je an 9 Stellen): 54 von 54 rot. V3-2: `<= 0`, `< 1`, erstes Byte, letztes Byte, Länge unter 20, `!= 0`: rot. V3-3: jede Grenze der Hex-Ziffern (`g`, `G`, `:`, `@`, `` ` ``) und `i+3`, `i+=3`, rot; `|` gegen `+` und der Schnellweg ohne `%` sind gleichwertig (Abschnitt 4). Das Probe-Paket findet am unveränderten Code 0 Abweichungen. | Abschnitte 3, 4 | — |
| <a id="v4-6"></a>V4-6 | LOW | **V3-5 ist geschlossen, Negativ (3) ist wahr für die gemessenen Fälle und zu eng für zwei andere.** Der Wirkungs-Fänger `TestGleichzeitigMitFilter` fängt Zustand hinter dem Feld `filter` (Abschnitt 5): W1 und Wrapper 20 von 20 bei vier Kernen, 0 von 20 mit `GOMAXPROCS=1`, Kontrolle 0 von 20; die Zahlen des Plans stimmen. Zwei Sätze des Negativs gehen aber über das Gemessene hinaus. (a) „braucht mindestens zwei gleichzeitig laufende Threads“: ein Fenster mit einer Abgabestelle (`runtime.Gosched()`, eine umkämpfte Sperre) fängt der Fänger auch mit einem Kern, 20 von 20 (WD, WF). (b) „auf einem Host mit einem Kern bleibt das ungemeldet“: ungeschützter Zustand hinter dem Feld meldet `make test-race` in 20 von 20 Prozessen auch mit `GOMAXPROCS=1` (WA bis WD), der Detektor ist der Fänger der Wahl und steht nicht im Negativ; ungefangen bleibt tatsächlich nur geschützter Zustand mit engem Fenster auf einem Kern (WE: 0 von 20). | Abschnitt 5 | Planner (Wortlaut von Negativ (3)) |
| <a id="v4-7"></a>V4-7 | LOW | **Negativ (5) ist berichtigt und trotzdem nicht vollständig:** die Korrektur des Orts stimmt (`outboundtest/metadata_codec_aufloesen.go:324`, die Fälle `a²`, `a` + U+0663, U+0663 + `a`, `a` + U+0301; `csdl/namen.go:33`). „Ungefangen ist nur die Klasse `Nl`“ stimmt nicht: auch `Mc` und `Me` (zusammen als Ziffer erlaubt), `Nl` als Ziffer und `Sc` bleiben grün (Tabelle 4.4); gefangen sind `ASCII` (`B2_Kodierung`), `M`+`No` zusammen, `Mn` allein, `No` allein, die Ziffer am Anfang. Ohne Sicherheitswirkung, wie das Negativ sagt. | Abschnitt 4, Tabelle 4.4 | Planner (Wortlaut) |
| <a id="v4-8"></a>V4-8 | INFO | **Messung und Zahlen stimmen.** `git diff --numstat 7535bf7..HEAD` über die 14 Pfade des Blocks: Einfügungen 5736, Löschungen 9, **Summe 5745**, gleich Plan und Ausgabe des Gates (`Messung 5745 von Marke 9200`, `zeilen=61 nebenlaeufig=2 fakes=1 produktcode=1700 schaetzung=6790`). Die Tabelle *Zusage · Fänger* hat 61 Zeilen (Zählung der Zeilen der Form `| N |` im Abschnitt: 61). Elementzahlen: 1160 = 29 × 10 × 4, 16 = 4 × 4, 1456 = 4 × 364 (364 = 3 × (10 × 7 + 10 + 4 × 10) + 4), 4800 = 16 × 300; die Tests prüfen jede außer 4800 mit einer Zählprobe (`t.Fatalf`; 16 und 300 sind Konstanten im Code). Rohbytes: in Plan und den drei Testdateien 0 Folgen aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP (`LC_ALL=C grep -P`); die Zeichen der Menge stehen als Bytewerte und werden gegen ihren Codepunkt gehalten. | Abschnitt 2 | — |

**Offen aus Report 3, nicht Auftrag dieser Runde (§3.13):** die Planner-Punkte (g), (i), (j), (k), (l) in §6 des Plans habe ich weder geprüft noch bewertet; die Closure ist nicht Auftrag.

**Verdikt: NICHT BESTANDEN**, wegen V4-1 (MEDIUM). Es ist eine Fänger-Lücke, keine Abweichung des Produktcodes: das Probe-Paket findet am Code **0 Abweichungen** vom Wortlaut in rund 50 Mio Eingaben. V3-1 bis V3-3 und V3-5 sind in der benannten Klasse geschlossen; offen ist die Nachbarklasse derselben Zusage (ASCII-Zeichen und Länge in Zeichen an den Stellen). V4-2 bis V4-7 sind LOW und blockieren allein nicht.

## 2. DoD und Gates

- **`make gates`** im Worktree auf `dad40cf`: **Exit 0** (aus der Datei `gates.rc`, nicht über `| tail` maskiert). `d-check`: 369 Dateien (ohne diesen Report), 0 Befunde; `baseline-verify` 54 Dateien; `groessenmarke slice-odata-url-pfad.md: ok, Messung 5745 von Marke 9200`.
- **Als `CACHED` gelaufene Stufen einzeln ohne Cache nachgefahren** (`docker build --no-cache-filter <Stufe> --target <Stufe>`): `test` Exit 0, `pruefung test: ok` (133 s); `race` Exit 0, `pruefung race: ok` (155 s); `lint` Exit 0, `0 issues.` (zweimal, Paket und Werkzeuge). Gesamt 375 s. `make groessenmarke` allein: `ok, Messung 5745 von Marke 9200`.
- **Messung nachgerechnet:** siehe V4-8. Das Delta `55781f7..dad40cf` ist 230 Einfügungen, 27 Löschungen in Plan und Tests; kein Produktcode.
- **Plan-Zahlen gegen Tests:** `TestUnicodeKlassen` 1160 (Zählprobe 10 Orte, 29 Zeichen, 6/4/19), `TestNulEinheit` 16, `TestHexUngueltig` 364 Formen und 1456 Elemente (die Teilsummen des Plans stimmen), `TestGleichzeitigMitFilter` 16 × 300. Die Aussage „35 Mutationen“ (9 Stellen × 4 Varianten, ohne `S1`-Ziffer am Anfang) habe ich selbst gefahren: 35 rot, 1 gleichwertig.
- **Orte des Pfad-Glieds:** der Plan zählt zehn Orte (neun Stellen im Code, das Pfad-Glied auf zwei Wegen); nachgezählt: `typPraefix` (`literal.go:125`), `namenAnfang` in `mengeOderEntity` (`pfad.go:53`), `einfacherName` in `restForm` (`pfad.go:107`), `sonstigerName` für den Alias (`query.go:56`), `elementePfad` (`query.go:156`, von `$select` und `$orderby` gerufen), `leseSelect` (`query.go:174`), `leseOrderBy` (`query.go:206`), `nacktesWort` (`schluessel.go:87`), `strukturiere` (`schluessel.go:96`): **9 Stellen, 10 Orte**, wie der Plan. Dazu die elfte Stelle von V4-2. Außerhalb von `urlparser` liest allein `csdl/namen.go:33` einen Namen (`grep` nach `unicode.Is…`).
- **Dekodierstellen:** drei Aufrufe von `dekodiere` (`pfad.go:19`, `query.go:77`, `query.go:91`) gleich vier Orte (Segment 1 und 2 teilen den Aufruf, Name, Wert), wie der Plan. `hexWert` steht zusätzlich in `literal.go:104` (Guid), das ist keine Dekodierung.

Alle Punkte der DoD sind `[ ]`, §7 ist „Wird bei der Closure gefüllt“ (Closure ist nicht Auftrag dieser Runde).

## 3. Probe-Paket (Orakel nie der Parser)

`probe_test.go` im Scratchpad, als Test im Paket `urlparser_test` einer Kopie (Kopie enthält Modell, Port `outbound` und `urlparser`; `go.mod` ohne Abhängigkeiten).

| Probe | Menge | Orakel | Ergebnis am unveränderten Code |
|---|---|---|---|
| `TestProbeNamen` | jedes Zeichen ab U+0080 außer Surrogaten (1 111 936) und 27 ASCII-Zeichen ohne Strukturzeichen, × 4 Stellungen (allein, erste, innere, letzte) × 10 Orte (9 Stellen, `S5` auf zwei Wegen), zusammen 44,5 Mio Aufrufe | `^[\p{L}_][\p{L}\p{Nd}_]*$` mit 1 bis 128 Zeichen; Typpräfix `^[\p{L}\p{Nd}_.]+$` | 0 Abweichungen (eine Ausnahme der Probe, kein Befund: `$select=*` ist `Alle`, das Orakel nimmt es aus) |
| `TestProbeLaenge` | Namen aus 1, 127, 128, 129, 130 mal `a`, `ä`, U+3042, U+1D400 an 5 Stellen | dieselbe Regex mit Zeichenzahl | 0 Abweichungen |
| `TestProbeHexPaare` | jedes Paar von Bytes (ohne Steuerzeichen und `/?#&=`) nach `%` an vier Stellen (Segment 1, Segment 2, Name, Wert) in 7 Kontexten (allein, `a` davor, `a` dahinter, `%80%80` dahinter, `%C3` davor, `%C3` dahinter, `%41` beidseits): 332 668 Einheiten × 4 | eigener Dekodierer (`strconv.ParseUint`, `utf8.ValidString`, kein U+0000) | 0 Abweichungen |
| `TestProbeFolgen` | alle Folgen aus 1 bis 3 von 50 Bruchstücken (`%G1`, `%1g`, `%@1`, `` %`1 ``, `%:1`, `%/1`, `%00`, `%C0%80`, `%ED%A0%80`, `%F4%90%80%80`, `%EF%BF%BD`, `%EF%BF%BE`, rohe Bytes `\xc3`, `\xff` …): 127 550 Einheiten × 4 Stellen | wie oben | 0 Abweichungen |
| `TestProbeHerstellerOption` | jedes Zeichen ab U+0021 außer `& = # % $ @`, jedes Byte als `%XX`, × 4 Stellungen, dazu Längen 127 bis 129: 4,45 Mio Fälle | `^[A-Za-z_][A-Za-z0-9_.\-]*$`, 1 bis 128 Bytes | 0 Abweichungen |

## 4. Mutationen

Je Mutation: Kopie der Basis, eine Änderung, `go test -failfast` über das Paket `urlparser` ohne die Probe; war die Suite grün, folgte das Probe-Paket, um „gleichwertig“ von „ungefangen“ zu trennen (Probe 0 Abweichungen = gleichwertig). Nebenläufige Zusagen: Abschnitt 5.

**4.1 Namen an 9 Stellen (Basis `nameM`, je Stelle ersetzt der Aufruf; 144 Mutationen).**

| Variante | Änderung | rot | grün, nicht gleichwertig | gleichwertig |
|---|---|---|---|---|
| VA, VB, VNl, VFirst (die 35 des Implementers) | nur ASCII; `M`+`No`; `Nl` als Buchstabe; Ziffer am Anfang | 35 von 36 | 0 | 1 (`S1`+VFirst) |
| Lo, Lt, Lm | Buchstabe ohne `Lo`, `Lt`, `Lm` | 27 von 27 | 0 | 0 |
| Lsupp, Ndsupp | Buchstabe bzw. Ziffer nur BMP | 18 von 18 | 0 | 0 |
| NdAscii | Ziffer nur `0` bis `9` | 9 von 9 | 0 | 0 |
| UsInner | `_` nur am Anfang | 6 von 9 | **3** (Alias, Pfad-Glied, nacktes Wort) | 0 |
| Dash | `-` an nicht erster Stelle erlaubt | 6 von 9 | **3** (Rest-Glied, Pfad-Glied, `$orderby`) | 0 |
| ByteLen | Zeichenzahl in Bytes | 0 von 9 | **8** | 1 (`typPraefix` hat keine Längengrenze) |
| Sc, Co, Ps | Kategorie als Buchstabe zugelassen | 5 von 27 (rot allein durch die ASCII-Zeichen `$` und `(`) | **22** (Sc 6, Co 9, Ps 7) | 0 |

Die gemeinsame Funktion `namenAnfang` selbst: siehe 4.6.

**4.2** Die Zählung zu V4-1: 3 + 3 + 8 = 14 grüne Mutationen der Klassen (a) bis (c); Probe: `UsInner` 2 (Alias, nacktes Wort) bzw. 4 (Pfad-Glied, zwei Wege) Abweichungen je Stelle, `Dash` 2 bzw. 4, `ByteLen`: `TestProbeLaenge` rot an allen 5 Stellen, an denen die Probe die Länge prüft (Menge, Alias, `$select`, `$orderby`, Schlüsselname; für Rest-Glied, Pfad-Glied und nacktes Wort nicht geprüft).

**4.3 Elfte Stelle (`sonstigerName`, 15 Mutationen).**

| Mutation | Suite | Probe |
|---|---|---|
| O1 Byte ab 0xC0 als Buchstabe | grün | 0, gleichwertig (jede UTF-8-Folge endet mit einem Folgebyte 0x80 bis 0xBF) |
| O3 `@` und `` ` `` im Namen | grün | 10 |
| O4 `[` und `{` im Namen | grün | 16 |
| O5, O5b `.` bzw. `-` nicht erlaubt; O6 Länge `maxName+1` | rot | – |
| O7 `:` als Ziffer | grün | 4 |
| O8 `-` am Anfang | grün | 4 |
| O9a Buchstabe jenseits von ASCII (`IsLetter`) | rot (`TestFehlerwert`) | – |
| O9b `Nd`, O9c `Mn`, O9d `No`, O9e `Nl` jenseits von ASCII im Namen | grün | 1520, 4118, 1830, 478 |
| O9f `Lo`, O9g `Lm` als Buchstabe | grün | 564 248, 1640 |

**4.4 `csdl/namen.go` (Negativ (5), 9 Mutationen).** `C1` nur ASCII, `C2` `M`+`No`, `C4` Ziffer am Anfang, `C5` `Mn`, `C6` `No`: rot (`TestVertrag/B2_Kodierung`, `D6_Namen_und_Eindeutigkeit`). Grün: `C3` `Nl` als Buchstabe (vom Negativ benannt), `C7` `Nl` als Ziffer, `C8` `Mc`+`Me` als Ziffer, `C9` `Sc` als Buchstabe (nicht benannt).

**4.5 Dekodierung (`dekodieren.go`, 15 Mutationen).** Rot (12): `hexWert` mit Obergrenze `g`/`G`, `:`, Untergrenze `@`/`` ` ``, `i+3 >= len`, `i += 3`, Schnellweg ohne `%` liefert `true`, Länge unter 20, NUL nur `!= 0`, `< len-1`, erstes und letztes Byte. Gleichwertig (2): `|` gegen `+` bei der Zusammensetzung der Nibble (Bits überlappen nicht); `utf8.ValidString` statt `gueltigerText` im Schnellweg (kein rohes U+0000 erreicht ihn: `pruefeArgumente` weist jedes rohe Zeichen unter 0x20 ab, Probe 0). **Grün, nicht gleichwertig (1): U+FFFD (V4-4).**

**4.6 Gemeinsame Funktion `namenAnfang`** (eine Änderung an der Funktion wirkt auf alle Stellen außer `typPraefix` zugleich; 3 Mutationen, ganze Suite). SH1 Zeichenzahl in Bytes (`zeichen += utf8.RuneLen(r)` statt `zeichen++`): **grün**, `TestProbeLaenge` rot (Mehrbyte-Namen von 65 bis 128 Zeichen würden abgelehnt) — die einzeilige Änderung an der einen Stelle, die alle Namen zählt, fängt kein Test. SH2 `_` nur am Anfang: rot (`TestGeneratorPfad`, `TestGeneratorKodierung`, `TestGeneratorQuery`). SH3 `-` erlaubt: rot (`TestFehlerwert`, `TestFehlerOhneEingabe`, `TestGeneratorPfad`). Die Klassen (b) und (c) von V4-1 sind also nur gefangen, wenn die gemeinsame Funktion sich ändert; eine Stelle mit eigener Regel (etwa eine Kopie der Funktion an einem Aufruf, wie die Mutationen der Tabelle 4.1 sie bilden) bliebe grün.

## 5. Nebenläufige Zusagen (Raten mit Zahl der Prozesse)

Messender: ich, im Container mit `--cpus=4`, 20 Prozesse je Zelle, `-test.run '^TestGleichzeitigMitFilter$' -test.count=1`, übersetztes Testprogramm. „Rot“ zählt Prozesse mit Exit ungleich 0. `/proc/loadavg` im Lauf nicht festgehalten (fremde Läufe in der Warteschlange); Kontrolle und Mutation liefen in demselben Container nacheinander.

| Mutation (Zustand hinter dem Feld `filter`) | Gestalt-Fänger `TestParserOhneZustand` | vier Kerne | `GOMAXPROCS=2` | `GOMAXPROCS=1` | `-race`, vier Kerne | `-race`, ein Kern |
|---|---|---|---|---|---|---|
| Kontrolle (keine Mutation) | grün | 0 von 20 | – | 0 von 20 | 0 von 20 | 0 von 20 |
| W1 (Feld `last` der Instanz, geschrieben vor, gelesen nach dem Leser) | rot | 20 von 20 | 20 von 20 | 0 von 20 | – | – |
| WG (Wrapper in `WithFilter`, liest `last` nach dem ersten Leseraufruf) | grün | 20 von 20 | 20 von 20 | 0 von 20 | – | – |
| WA (Wrapper, `last` geschrieben und sofort gelesen, ungeschützt) | grün | **0 von 20** | – | 0 von 20 | 20 von 20 | 20 von 20 |
| WB (Wrapper, Ergebnis nach dem Aufruf in ein Feld geschrieben und gelesen) | grün | **0 von 20** | – | 0 von 20 | 20 von 20 | 20 von 20 |
| WC (Wrapper, `last` vor und nach dem Aufruf geschrieben) | grün | **0 von 20** | – | 0 von 20 | 20 von 20 | 20 von 20 |
| WD (WA mit `runtime.Gosched()` im Fenster) | grün | 20 von 20 | – | **20 von 20** | 20 von 20 | 20 von 20 |
| WE (Wrapper, `last` unter Sperre geschrieben, dann unter Sperre gelesen; übrige Suite grün) | nicht einzeln gelaufen (Parser unverändert) | 20 von 20 | – | **0 von 20** | 20 von 20 | 0 von 20 |
| WF (WE mit `runtime.Gosched()` zwischen den Abschnitten) | nicht einzeln gelaufen (Parser unverändert) | 20 von 20 | – | 20 von 20 | 20 von 20 | 20 von 20 |

Die übrige Suite (ohne den Wirkungs-Fänger) ist bei W1 rot (`TestVerschraenktLeser`, ein bestehender Fänger), bei WG rot (der Wrapper ruft den Leser zweimal; Zähl-Tests), bei WE grün (der reine Fall für den Wirkungs-Fänger), bei WF rot (`TestVerschraenktLeser` läuft in die Frist von 10 s). Bei WA bis WD ersetzt meine Mutation `WithFilter(nil)` nicht vollständig (`TestFilterOhneLeser` rot, ein Artefakt der Mutation); diese vier Zeilen werten allein die Spalten der Raten.

Lesart: Die Zahlen des Plans (W1 und Wrapper: 20 von 20 mit mehreren Kernen, 0 von 20 mit einem Kern, Kontrolle 0) sind wahr. Die Klasse „Zustand hinter dem Feld“ ist durch zwei Fänger gedeckt, die sich ergänzen: der Wirkungs-Fänger (Fenster mit Abgabestelle, oder geschützter Zustand bei mehreren Kernen) und der Detektor von `make test-race` (ungeschützter Zustand, in jedem Prozess, auch mit einem Kern). Nicht gefangen bleibt geschützter Zustand mit engem Fenster auf einem Kern (WE, 0 von 20), das benennt das Negativ der Sache nach richtig; zum Wortlaut V4-6. Die Wrapper WA bis WC sind mit dem Wirkungs-Fänger allein nicht rot (0 von 20 bei vier Kernen); sie hängen am Detektor.

## 6. Prüfung der Entscheidungen und der Negative

- **Negativ (3)** (Zustand nach der Rückkehr des Lesers): Entscheidung trägt, Wortlaut zu eng, V4-6.
- **Negativ (4)** (Vervielfältigungsfaktor unter 64 bei `$orderby` und Schlüssel): im Delta unverändert; nicht neu bewertet.
- **Negativ (5)**: Ortskorrektur wahr, Klassenangabe falsch, V4-7.
- **Tabelle *Zusage · Fänger*:** die geänderten Zeilen 14, 24, 26, 28 und das Raster tragen die neuen Mengen; die Zeile 28 nennt 1160 und die Kennungen `S1-V1` bis `S9-V4`; die ältere Fassung (348) bleibt als „Menge der Nachverifikation 3“ in N-2 stehen und ist als solche gekennzeichnet. Eine Zusage ohne Fänger ist in der Zeile 28 die Nachbarklasse von V4-1.
- **Raster der Randformen ([AGENTS.md](../../AGENTS.md) §3.16):** das Delta ändert nur Mengen von Fängern; das Raster in §6 habe ich in dieser Runde nicht erneut geprüft.

## 7. Nicht gelaufen und Vorbehalte

- **`make gegenproben`, Stempel:** nicht gefahren (nicht Auftrag).
- **Die Stufe `go build` von `make build`** habe ich nicht erneut ohne Cache gefahren; der Produktcode ist gegenüber Report 3 unverändert.
- **Schließung der Unicode-Tabelle:** das Probe-Paket nutzt die Tabellen von Go selbst als Orakel (`regexp`, `\p{L}`); eine Abweichung zwischen der Tabelle des Compilers und dem Unicode-Standard sähe es nicht.
- **Äquivalenzen** (`|` gegen `+`, Schnellweg, `typPraefix` ohne Länge, O1, VFirst auf `S1`) sind begründet und von der Probe mit 0 Abweichungen gestützt, nicht bewiesen.
- **Die Raten von Abschnitt 5** sind 20 Prozesse je Zelle; „0 von 20“ ist keine Obergrenze der Wahrscheinlichkeit (bei p = 0,1 liegt die Chance auf 20 Nullen bei rund 12 %).
- **Eine Mutation „Rest-Glied nur ASCII-Name“** oder ähnliche Kombinationen über mehrere Stellen zugleich habe ich nicht gefahren; je Stelle einzeln ist die Menge geschlossen.

## 8. Register-Kandidaten (nur genannt, nicht geschrieben)

1. **Allaussage über „Name“ ohne ASCII-Vertreter und ohne Längengrenze in Zeichen** (V4-1): zweite Runde in Folge, in der die Menge der Klasse aus der Kategorie gebildet ist und der Teil der Klasse fehlt, den die Kategorie nicht nennt (`_`, `-`, Länge). Wiederholung des Musters von V3-1.
2. **Eine Stelle, die dieselbe Zusage mit anderer Regel trägt, steht in keiner Menge** (V4-2): die Zählung „neun Stellen“ stimmt für die Regel, die elfte Stelle (ASCII-Namen) liegt daneben.
3. **Ein Negativ benennt die Klasse der ungefangenen Fälle zu eng** („nur `Nl`“, V4-7; „braucht zwei Threads“, V4-6): zweite Runde mit einer Allaussage im Negativ, die eine Messung überholt.
