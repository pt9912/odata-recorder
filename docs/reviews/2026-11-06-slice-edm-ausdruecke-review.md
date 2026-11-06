# Review-Report: slice-edm-ausdruecke — 2026-11-06

**Review-Art:** Code — gegen Plan, Entscheidungen und Hard Rules (nicht gegen die DoD; das ist die Frage des Verifiers).

**Gegenstand:** Zweig `slice-edm-ausdruecke`, Diff `f6daf34..04c6e0e` (HEAD `04c6e0e`), ein Lauf ohne Wiederholungsschleife.

**Skill:** `.harness/skills/reviewer.md` @ 9dd4e5d (die erweiterte Fassung von `main`: MEDIUM-Zeilen „Menge ohne Achsenliste“, „Fänger schmaler als Zusage“, „unbeschränkte Menge ohne Prüfgrenze“, „Fixture gleicht dem Wert der Mutation“, Hinweis zu `panic:`) · <!-- d-check:ignore (Skill-Pfad der Vorlage; die Datei liegt im Repo) -->
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-06

**Eingangs-Kontext:**

- Slice-Plan `slice-edm-ausdruecke` (§3 Tabelle mit 38 Zeilen und „Stand der Fänger beim Handoff“, §4 Messung 2780 von 5000, §6 Raster und zwei offene Zeilen für den Architect)
- `SPEC-166`, `SPEC-167`, `SPEC-109` Nr. 5, `LH-FA-08.b`
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md), [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)
- `AGENTS.md` §3.7, §3.10, §3.11, §3.12, §3.13, §3.15, §3.16, §3.17, §3.18

**Vorgehen.** Jede Zeile der Tabelle und jede Zelle des Rasters wurde mit eigenen Mutationen am Produktcode aufgelöst (Wegwerfkopie, `go test` des Pakets `src/hexagon/model` im Image `odata-recorder:test`, ohne Netz, 2 GB). 31 Mutationen, dazu 10 Mutationen für die Nebenläufigkeit (Detektor je Prozess mit `-test.count=1`, 20 Prozesse je Mutation, Image `odata-recorder:race`). Die Klassen wurden gegen die Achsen Zeichen, Länge, Folge, Seite, Stelle und Zustand geprüft; die Mutationen unten sind die, die grün blieben, und die, die die Aussage der Befunde tragen. Das Rohbyte-Gegenlesen (`AGENTS.md` §3.18) lief als Festtext-Suche mit `printf`-gebautem Muster (`grep -P '\xc3\xbc'` liefert in dieser Umgebung auch bei vorhandenem Byte keinen Treffer und taugt hier nicht).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Zusage „Text sonst unbegrenzt“ (Raster R1, `SPEC-166` Nr. 7: Aufwand linear in der Länge) hat keinen Fänger auf der Achse Länge: keine Textliste enthält einen Wert über 256 Bytes. Mutation `s4`: `vergleicheGleichartig` vergleicht Text nur bis Byte 256, ebenso `s2` bis 1024 und `s3` bis 65536; in allen drei Fällen 0 rote Tests im Paket `model`. Szenario: zwei Werte einer Property, die erst hinter Byte 256 abweichen, vergleichen gleich, `$filter=Name eq '…'` und `$orderby` liefern falsche Treffer. | `AGENTS.md` §3.15, §3.17; Plan §3 *Klassen als Generator* (Zeile 4, 19) | `src/hexagon/model/wert_vergleich.go` · `return sign(strings.Compare(a, b))`; `src/hexagon/model/wert_vergleich_test.go` · `func textListe()` | ja — `go test ./src/hexagon/model/` mit der Mutation | Fänger führt eine Länge der Eingabe nur (Text) |
| F-2 | MEDIUM | Die Grenze „höchstens 1024 Zeichen“ der kanonischen Zahl (`SPEC-166` Nr. 3 c, 7) ist an beiden Seiten nur für positive Texte gefangen. Mutation `n1`: die Länge wird ohne das Vorzeichen gezählt (`-` plus 1024 Ziffern, 1025 Zeichen, gilt als gültig): 0 rote Tests; Mutation `s6` (Grenze 1025 nur bei Vorzeichen): 0 rote Tests. `dezimalListe` und der Generator `randformen` führen kein negatives Element an der Grenze; die Tabelle (Zeile 3) nennt „1024 Zeichen gültig, 1025 `ErrUngueltig`“ ohne die Achse Vorzeichen. Szenario: ein Wert von 1025 Zeichen wird als `Edm.Decimal` verglichen, den `SPEC-126` Nr. 10 b abweist. | `AGENTS.md` §3.15 (beide Seiten jeder Grenze), §3.17 | `src/hexagon/model/wert_vergleich.go` · `len(s) > maxZahlLaenge`; `src/hexagon/model/wert_vergleich_test.go` · `lang("1", 1023, "0", "")` | ja — `go test ./src/hexagon/model/` mit der Mutation | Grenze nur auf einer Seite der Achse Vorzeichen gefangen |
| F-3 | MEDIUM | Die Zelle „`typ` ohne Properties: jeder Name ist unbekannt“ (Raster R6, Spalte *fehlend*, Ort `SPEC-167` Nr. 3 a) hat keinen Fänger: kein Test übergibt einen `EntityType` ohne `Properties`; `typMit()` ist der einzige Aufbau. Mutation `s1`: `eigenschaft` liefert bei leerer `Properties` eine gültige Text-Property: 0 rote Tests. Szenario: `$filter=Foo eq 'x'` auf einem Typ ohne Properties würde angenommen statt mit `400` abgewiesen. | `AGENTS.md` §3.15, §3.16 | `src/hexagon/model/filter_typ.go` · `func (p filterPruefer) eigenschaft(`; `src/hexagon/model/filter_typ_test.go` · `func typMit(` | ja — `go test ./src/hexagon/model/` mit der Mutation | Rasterzelle mit Ort, aber ohne Fänger |
| F-4 | MEDIUM | Die Zusage „Nullwert `WahrheitUnbekannt`“ (`SPEC-166` Kopf, Raster R4 *fehlend*) ist nur symbolisch geprüft, nie am Zero-Value. Mutation `v1`: die Konstanten werden so geordnet, dass `Wahrheit(0)` falsch ist (`WahrheitFalsch Wahrheit = iota`): 0 rote Tests. Szenario: eine nicht gesetzte `Wahrheit` im Interpreter (`var w model.Wahrheit`) wird durch `Nicht(w)` zu wahr und lässt einen Datensatz durch `$filter` (`Durchlaesst`), obwohl die Spezifikation „unbekannt“ zusagt. | `AGENTS.md` §3.15, §3.16; `SPEC-166` Nr. 5 | `src/hexagon/model/wahrheit.go` · `WahrheitUnbekannt Wahrheit = iota`; `src/hexagon/model/wahrheit_test.go` · `func tafelEingaben()` (enthält 3 und -7, nicht 0) | ja — `go test ./src/hexagon/model/` mit der Mutation | Zusage Zero-Value ohne Fänger |
| F-5 | MEDIUM | Der Fänger der Zeile 20 (`TestGleichzeitig`) ist schmaler als die Zusage „die Funktionen … gleichzeitige Aufrufe beeinflussen sich nicht“ (`SPEC-166` Nr. 8 nennt alle Funktionen). Er ruft sechs Funktionen und von `Vergleiche`/`OrdneWerte` nur die Klassen Zahl, Zeitpunkt, Text, keinen Wert `ArtNull`, keine Klasse Datum, Wahrheit, Guid. Mutation: ein unbeschütztes Paketzähler-`++` in `WertAusLiteral` (`q1`), `WahrheitVon` (`q2`), `HatOrdnung` (`q3`), `Durchlaesst` (`q5`), im `null`-Zweig von `Vergleiche` (`q6`) und von `OrdneWerte` (`q8`): Detektor je Prozess `-test.count=1` in 0 von 20 Prozessen rot (ohne Detektor ebenfalls 0 von 20, auch mit `-test.count=20`). Zum Vergleich in 20 von 20 rot: Zähler in `Error()` (`q4`), im Maskenaufruf (`q7`), in `bruchteil` (`q9`) und das Ergebnisfeld `r3`. Der Plan nennt die Differenz nicht als akzeptiertes Negativ mit Grund und Folge; nur die allgemeine Zeile „Datenrennen, das kein Lauf auslöst“ in `SPEC-166` *Ungeprüft* trägt sie. | `AGENTS.md` §3.10, §3.15; [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md) | `src/hexagon/model/filter_typ_test.go` · `func TestGleichzeitig(`; Plan §3 Zeile 20 und *Raten der nebenläufigen Zeile 20* | ja — `go test -race -test.run '^TestGleichzeitig$' -test.count=1` mit der Mutation, 20 Prozesse | Fänger schmaler als Zusage, Differenz nicht benannt |
| F-6 | LOW | Zwei Randformen ohne Ort sind im Code entschieden statt zurückgegeben: ein `FilterTypFehler` mit unbekannter Ursache wird zur Klasse `ungueltig` (`klasse()`-Standardzweig), und ein von Hand gebauter Text `1.0` im `LiteralZahl` wird `literal_bereich`. Beide stehen in §6 mit Ausgang *offen* und Adresse Architect (`AGENTS.md` §3.13 erfüllt); `AGENTS.md` §3.12 verlangt zusätzlich das Anhalten vor dem Code. Kein Fänger für den ersten Fall (Zustand nicht erreichbar), der zweite ist durch `TestBereich` (Text `1.0`) und `TestFilterAusLeser` (Leser liefert `1`) gefangen. | `AGENTS.md` §3.12, §3.16 | `src/hexagon/model/filter_typ_fehler.go` · `jede Ursache außerhalb der sieben gehört zur Klasse`; Plan §6 · `Randform ohne Ort: ein FilterTypFehler` | nein — Entscheidung des Architect, kein Gate | Randform im Code entschieden (gemeldet, nicht angehalten) |
| F-7 | LOW | Zwei Kommentare sagen mehr zu, als ein Test prüft. (a) `AlleFilterTypUrsachen` „in der Reihenfolge von `SPEC-167` Nr. 6“: Mutation `v5` (zwei Einträge vertauscht) lässt 0 Tests rot. (b) Das Kommentar in `TestReihenfolgeVerstoesse` „typ_unvertraeglich zuerst; Bereich vor Operator“: die zweite Hälfte hat keinen Fänger (im Plan §3 als äquivalente Mutation geführt, im Test nicht ausgenommen). | `AGENTS.md` §3.11 | `src/hexagon/model/filter_typ_fehler.go` · `in der Reihenfolge von SPEC-167 Nr. 6`; `src/hexagon/model/filter_typ_test.go` · `Bereich vor Operator` | ja — `go test ./src/hexagon/model/` mit `v5` | Kommentar sagt Reihenfolge zu, Test prüft sie nicht |
| F-8 | INFO | Die Zählprüfungen in fremden Tests (`19*26` in `kontext_test.go`, `4*(26+4+1)` in `reset_test.go`, je eine Zeile in `helfer_test.go` und `beenden_test.go`) sind begründet: `TestTabellenFehlerVollstaendig` und `TestTabellenzeilenVollstaendig` halten die Liste gegen `EintraegeAbbildung`, die Zählung hält die Matrix gegen Schrumpfen. Jede künftige Zeile der Tabelle berührt dieselben vier Dateien; die Pfade stehen in Plan §3, nicht im Block von §4 (die Messung 2780 über die elf Pfade des Blocks ist nachgerechnet und stimmt). | `AGENTS.md` §3.9, §3.14 | `src/hexagon/services/workspace/kontext_test.go` · `faelle != 19*26` | nein | Kopplung einer neuen Tabellenzeile an vier fremde Testdateien |
| F-9 | INFO | Das akzeptierte Negativ zu `r3` (geteiltes Ergebnisfeld, kein Wirkungs-Fänger) ist nachgemessen und trägt Grund und Folge: Detektor 20 von 20 Prozessen, ohne Detektor 0 von 20 (`-test.count=1`) und 0 von 20 Prozessen (`-test.count=20`). Die Raten des Plans stimmen. | `AGENTS.md` §3.10 | Plan §3 · `Folge für r3` | ja — Lauf wie F-5 | — |
| F-10 | INFO | Das vom Implementer gemeldete SQLite-Flake in `make test-race` (`TestTransaktionVertrag/…ReplaceAll`, „cannot commit - no transaction is active“, unter Last) liegt außerhalb des Diffs; nicht untersucht, nur vermerkt. | Maintainability | — | nein | — |

## Mutationen, die rot wurden (Auswahl der Auflösung)

| Zeile / Zelle | Mutation | roter Test |
|---|---|---|
| 3 (Länge ohne Punkt) | `n2`: Länge ohne `.` gezählt | `TestDecimalExakt` |
| 3 (Grenze 1025) | `s5`: `len(s) > maxZahlLaenge+1` | `TestDecimalExakt` |
| 3, 16 (negative Brüche) | `n3`: Vorzeichen kehrt den Bruchteil nicht um | `TestDecimalExakt`, `TestOrdneWerteGleichVergleiche` |
| 7 (Bruchteil 12 Ziffern) | `n4`: höchstens 11 | `TestZeitpunktOrdnung`, `TestFormverstoesse`, `TestOrdnungTotal` |
| 9, 19 (Datum, Länge) | `n5`: `len(s) >= 10` | `TestFormverstoesse` |
| 10 (Seite) | `n6`: Guid-Prüfung nur links (`Vergleiche`); `n7` ebenso in `OrdneWerte` | `TestGuid`, `TestKlassenUnvergleichbar` |
| 11, 19 (`ArtNull` mit Text) | `n8`: Text eines `ArtNull` wird geprüft | `TestFormverstoesse`, `TestNullVergleiche` |
| 12, 13 (Seite) | `n9`, `n10`: nur der linke Operand entscheidet | `TestUndTafel`, `TestOderTafel` |
| 15 (Wert außerhalb) | `n11`: `Durchlaesst` ab `WahrheitWahr` | `TestWahrheitVonUndDurchlaesst` |
| 25 (Folge) | `n12`: offener Typ vor den Properties | `TestNavigationUndOffen` |
| 31 (Seite) | `n14`, `n15`: Guid-Operator nur links bzw. nur rechts | `TestGuidOperatoren`, `TestNullGegenTyp` |
| 30 (Seite) | `n16`, `n17`: Bereich nur rechts bzw. nur links | `TestBereich` |
| 30 (Form) | `n19`: kanonische Form des Zahlliterals nicht verlangt | `TestBereich` |
| 22 (Stelle, Folge) | `n18`: `Links` von `and`/`or` nicht verlangt | `TestBooleschOperand`, `TestMaskeAufrufe`, `TestReihenfolgeVerstoesse` |
| 22 (Stelle) | `n20`: `not` ohne Prüfung des Operanden | `TestBooleschOperand`, `TestReihenfolgeVerstoesse` |
| 21, 23 | `n21`: `null` als Wurzel abgewiesen | `TestWurzel`, `TestBooleschGueltig` |
| 35 | `n22`: `Is` trifft auch `ErrURLUngueltig` | `TestFilterTypFehler` |
| 35 | `v6`: Standardzweig von `klasse()` auf `ErrURLNichtUnterstuetzt` | `TestFilterTypFehler` |

Äquivalent: `n13` (Navigation vor den Properties) — Namen von Properties und Navigation-Properties sind je Typ eindeutig (`SPEC-167` *Ungeprüft*, `SPEC-125` Nr. 11), kein Pfad unterscheidet die Folge.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/wert_vergleich.go`, `wahrheit.go` (Ordnung je Klasse, Form, Null-Semantik, Wahrheitstafeln) | geprüft; Befunde nur auf den Achsen Länge, Vorzeichen und Zero-Value (F-1, F-2, F-4); Orakel der Tests (`math/big`, `time`, reguläre Ausdrücke, Runen) sind vom Code unabhängig |
| `src/hexagon/model/filter_typ.go`, `filter_typ_fehler.go` (Folge der Verstöße, Stellen, Maske, Klassen, Codes) | geprüft; Befund nur zu `typ` ohne Properties (F-3) und den zwei Randformen aus §6 (F-6); die 49 mal 4 Paare der Verstöße, die acht Stellen und die Mutationen der Seite und Folge sind rot |
| `src/hexagon/ports/outbound/abbildung.go`, `abbildung_test.go` | geprüft, ohne Befund (Zeile `ErrFilterNichtBeobachtet` → `ODR-E6003`; 7 mal 4 mal 8 Fälle, auch verpackt und verbunden) |
| `src/hexagon/services/{session,workspace}/*_test.go` | geprüft, ohne Befund außer INFO F-8 |
| `tests/integration/filter_instant_test.go` | geprüft, ohne Befund (Versatz, Tag- und Jahreswechsel, 42 Schreibweisen gegen `time.Time`; [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md): Test unter `tests/` darf Modell und Driven Adapter nennen) |
| [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) | geprüft, ohne Befund: die neuen Dateien des Pakets `model` importieren nur `errors`, `fmt`, `strconv`, `strings`; kein V2- oder V4-Adapterpaket |
| `make a-check` auf dem Zweig | 0 Befunde (der Hinweis zu ungeschichteten Dateien unter `tools/` ist vorhanden und nicht aus diesem Diff) |
| Gate-Suppression (`nolint`, `noqa`, `d-check:ignore`, `t.Skip`) im Diff | geprüft, ohne Befund; `.golangci.yml`, `.a-check.yml`, `Makefile` unverändert |
| Kommentare (`AGENTS.md` §3.7) in den neuen Dateien | geprüft; kein Konjunktiv über verworfene Alternativen, keine Chronik; Befund nur F-7 (§3.11) |
| Rohbytes (`AGENTS.md` §3.18) in den zehn Dateien des Diffs | gegengelesen mit Festtext-Suche: `ungültig` mit Byte-Folge 0xC3 0xBC in Code, Test und Spezifikation, kein 0xC3 0x83 (Doppelkodierung), kein U+FFFD, kein BOM, NBSP oder Nullbreiten-Zeichen; Fixtures mit Sonderbytes entstehen aus `[]byte{…}` oder Hilfen (`nbsp()`, `bom()`) |
| Plan §6 (zwei offene Zeilen für den Architect, Sendung an `slice-query-interpreter`) | geprüft: Ausgang *offen* und Adresse stehen; der Nehmer führt die Sendung in seinem §1 |

## Nicht aufgelöst

- Der Lauf von `make gates` und `make test-race` als Ganzes (Verifier-Sache; hier nur `a-check` und die Pakettests mit Mutationen).
- Die Rate der Nebenläufigkeit wurde für zehn Mutationen mit 20 Prozessen gemessen, nicht mit mehr; die Aussagen 0 von 20 und 20 von 20 tragen F-5, sagen aber nichts über Mutationen in anderen Zweigen.
- Mutationen an `Edm.Date` mit Jahr außerhalb 0001 bis 9999 und an ungültigem UTF-8 (beides in `SPEC-166` *Ungeprüft*).
- Ob der Plan die Zeilen 3, 11, 21 im Wortlaut der Tabelle (14 Dezimalzahlen, acht Vertreter, 7 gültige Wurzeln) gegen die tatsächlichen Mengen (24, neun, 8) nachziehen muss, ist Sache des Planners (`AGENTS.md` §3.9); die Tests zählen ihre Mengen selbst.

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 5 |
| LOW | 2 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Fänger führt eine Länge der Eingabe nur · Grenze nur auf einer Seite einer Achse · Rasterzelle mit Ort, aber ohne Fänger · Zusage Zero-Value ohne Fänger · Fänger schmaler als Zusage, Differenz nicht benannt · Randform im Code entschieden · Kommentar sagt Reihenfolge zu, Test prüft sie nicht

## Verdikt

**Merge-blockierend:** ja — fünf MEDIUM; alle fünf betreffen Tests und Plan (je ein Fall oder ein benanntes Negativ mit Grund und Folge), im Produktcode wurde kein Fehler gefunden. Kein Rollen-Widerspruch, kein HIGH: der Konflikt-Pfad über den Architect entfällt.

**Übergabe:** Findings gehen an den Implementer; F-6 liegt beim Architect (Ort in `SPEC-167` Nr. 6 und Nr. 3 c). Die Finding-Klassen gehen in die Slice-Closure §7.
