# Verifikation: slice-odata-codec-csdl

* Datum: 2026-10-24
* Rolle: verifier (frischer Kontext; kein Produktcode, kein Test, keine Spezifikation, kein Plan geändert; `.claude/agents/architect.md` ist eine fremde, uncommittete Änderung und blieb ungelesen und unberührt)
* Gegenstand: Spitze `cf3131d`. Basis der Größenmarke `3b6b3b8`; Reviewer `0e11658`; Architect `f7639b9`, `684602d`; Implementer `867e878`, `f0256eb`, `d3a8baa`, `f54cec7`, `d2c566e`, `ed57fcb`, `f6986cf`, `ad8732b`, `7e08438`, `cf3131d`.
* Plan: [`slice-odata-codec-csdl.md`](../plan/planning/done/slice-odata-codec-csdl.md); Review: [`2026-10-24-slice-odata-codec-csdl-review.md`](2026-10-24-slice-odata-codec-csdl-review.md)
* Maßstab: DoD (§2), Tabelle *Zusage · Fänger* A1 bis J2 (§3) und Raster samt Zusatztabelle (§6) gegen `SPEC-125` Nr. 1 bis 14 und *Ungeprüft*; `SPEC-109` Nr. 5; `SPEC-049`; `LH-FA-06.a`, `LH-RB-01`, `LH-QA-07.a`; `AGENTS.md` §3.1 bis §3.16.
* Hinweis zum Skill: `.harness/skills/closure-note-reviewer.md` gilt für Slices in `done/`; dieser Slice liegt in `in-progress/`, der Skill wurde nicht angewandt.
* Belege selbst gefahren: `make gates` (Rückgabecode 0), die Suite des Slice ohne und mit Detektor, rund 250 eigene Mutationen als Wegwerfkopien von `git archive HEAD` unter dem Scratchpad (Quelle `:ro`, `--network none`, `--memory` 2 bis 4 GiB, höchstens drei Container zugleich, Ersetzung mit einem eigenen Perl-Skript, nie im Repo), Probe-Dokumente gegen den Leser, Raten als Testbinary mit N Prozessen zu `-test.count=1`. `git status --short` zeigt vor und nach dem Lauf nur `.claude/agents/architect.md`.

## Urteil

**Nicht bestanden: kein HIGH, vier MEDIUM, sechs LOW, vier INFO.** Die DoD-Punkte 1 bis 4 sind dem Wortlaut nach erfüllt, `make gates` ist grün, und alle 14 MEDIUM und 2 LOW des Reviews sind **für ihre Klasse** geschlossen, mit einem Rest bei F-5 (V-6). Vier Befunde sind Verifier-only-Klasse und unsichtbar für Suite und Review:

1. **V-1 (MEDIUM, Implementer)**: Eine `CDATA`-Sektion vor oder nach der Wurzel wird gelesen, wenn sie leer ist oder nur Leerraum enthält. `SPEC-125` Nr. 4 lässt dort nur Leerraum, Kommentare und Verarbeitungsanweisungen zu; kein Fänger.
2. **V-2 (MEDIUM, Architect und Implementer)**: Die Rangfolge mehrerer Verstöße **im selben Start-Tag** (Tiefe, doppeltes Attribut, Ersatzreferenz, Version, Struktur) und mehrerer `struktur`-Verstöße **im selben Element** (Ort) ist im Code entschieden, hat keine Zelle im Raster und keinen Fänger.
3. **V-3 (MEDIUM, Implementer)**: Kein Dokument der Suite hat einen Namensraum mit Punkt **und** eine Referenz in ihn. Die Mutation `qualify` teilt am ersten statt am letzten Punkt: die Suite bleibt grün, jedes Dokument eines Dienstes wie `Microsoft.OData.….TripPin` wäre `nicht_aufloesbar`.
4. **V-4 (MEDIUM, Implementer)**: Ein leerer `Key` in der Ableitung bei Schlüssel in der Basis ist als gültig getestet, aber nicht als Modell: die Mutation löscht `Keys` der Ableitung, die Suite bleibt grün.

Auflagen vor der Closure: V-1 bis V-4 (nach der Nachbesserung genügt eine Nachverifikation dieser Punkte), V-7 (Planner) und die übrigen LOW nach Ermessen des Planners. Die Größe (4199 gegen Marke 2400) ist die Entscheidung des Planners (Abschnitt 6).

* MEDIUM: V-1, V-2, V-3, V-4
* LOW: V-5, V-6, V-7, V-8, V-9, V-10
* INFO: V-11, V-12, V-13, V-14

## 1. DoD Punkt für Punkt (Plan §2)

| # | DoD | Stand | Beleg |
|---|---|---|---|
| 1 | Port `MetadataCodec`, Modell, `*MetadatenFehler`, Zeile in `SPEC-109` Nr. 5, kein Parsertyp in Signaturen, Vertragssuite | erfüllt | `src/hexagon/ports/outbound/metadata_codec.go` (Signatur nur `context`, `io`, `model`), `src/hexagon/model/metadaten.go`, `service_model.go`, `abbildung.go:61`; `outboundtest.MetadataCodecVertrag` (40 `t.Run`; F5 und H2 in den Paketen `model` und `outbound`, J1 in `.a-check.yml`, J2 in `leser_test.go`); H2 Mutation „Zeile entfällt“ und „nur ein Gegenstand“ rot (Review), `a-check` Fälle in Abschnitt 4 |
| 2 | CSDL-Adapter: Parsen, Referenzen, Vererbung, EDM-Typen, `ODR-E6004` mit Ursache und Ort für die genannten Bedingungen und Grenzen | erfüllt, mit V-1 und V-3 | dreizehn Ursachen, Ort nach `SPEC-125` Nr. 7 bis 9; ein TripPin-artiges Dokument (Alias, Punkt-Namensraum, `EnumType`, `ComplexType`, `Singleton`, `FunctionImport`, `Reference`, Facetten) wird gelesen und liefert 21 Funde in Dokumentreihenfolge; die Grenzen: Mutationen g1 bis g7 (Wegfall von Tiefe, Summe, Kette, Größe, Zeichenmenge, Namenslänge) alle rot |
| 3 | Validierungsbericht mit `unsupported_features`, ein Fall je Element von Nr. 7 | erfüllt | 71 von 72 Mutationen der Zeile G1 rot, die 72. gleichwertig (Abschnitt 3, F-7) |
| 4 | `make gates` grün | erfüllt | Abschnitt 4: Rückgabecode 0, `a-check` 0 Befund, `d-check` 240 Dateien 0 Befund, `baseline-verify` OK, `make test` und `make test-race` über die Stufen von `Dockerfile` (die Schichten waren zwischengespeichert, der Quellstand ist derselbe; zusätzlich die Suite selbst gefahren) |
| 5 | Review durchgeführt, Report liegt vor | erfüllt | `0e11658`; die Folgerunde ist in Abschnitt 3 nachverifiziert |
| 6 | Doku-Update bei öffentlichem Vertrag | erfüllt | `SPEC-125` nach dem Review nachgezogen (Nr. 1, 2, 3, 4, 5, 7, 8, 9, 11), Änderungszeile am Ende der Spezifikation; `spec/architecture.md` ohne Slice-, Wellen-, Hash- oder ADR-Bezug (Suche: kein Treffer); `make docs-check` grün |
| 7 bis 10 | Closure-Notiz, Register, Risiko-Ausgänge, drei Paarungen | offen für die Closure | §7 trägt „Wird bei der Closure gefüllt“; Abschnitt 7 |

Ein Häkchen setze ich nicht.

## 2. Tabelle *Zusage · Fänger*, Raster und Vertrag

**Tabelle gegen `SPEC-125`, Zeile für Zeile.** Zu jeder der 44 Zeilen gibt es einen Test unter dem genannten Namen. Jede Nummer des Vertrags hat Zeilen: Nr. 1 A1, A2; Nr. 2 A3, A4, A5; Nr. 3 B2; Nr. 4 B1, B3, B4, B5, B6; Nr. 5 E1 bis E3, D3, D6; Nr. 6 C1, C2; Nr. 7 G1, G2; Nr. 8 D6, D7; Nr. 9 D1, D2, D3, D4, D7, D8; Nr. 10 D4, D5; Nr. 11 F1 bis F6; Nr. 12 G3, G4; Nr. 13 D9, D10, H1, H2; Nr. 14 I1 bis I3. Die Allaussagen sind Schleifen (D4: 34 Namen × Rolle, D8: 28 Paare, D9, D10, G5, G1 über die Mengen); der Zusatzfall zählt die Mengen (34 und acht, dreizehn Ursachen, 28 Paare, mindestens 400 Dokumente in A4).

**Was in keiner Zeile steht** (die Befunde dieses Berichts): `CDATA` außerhalb der Wurzel (V-1), Verstöße im selben Start-Tag oder Element (V-2), Namensraum mit Punkt als Ziel einer Referenz (V-3), das Modell einer Ableitung mit leerem `Key`, `Path` mit Leerraum und `HasStream` gegen `Open` (V-4), offenes `Collection(` mit Rest und Leerraum je Referenzart (V-5).

**Raster.** Die Zellen nennen Orte, deren Wortlaut die Entscheidung trägt; die Zeilen für `Namespace`/`Alias` `Edm`, `Target` eines Bindings, Ring über 64 Typen und Referenz außerhalb der Wurzel sind seit dem Review nachgetragen. Zellen ohne Ort habe ich in V-1 und V-2 gefunden.

**Zusatztabelle.** P1 bis P27 plus die Schleife über die 28 Paare der acht Durchläufe: alle 28 Vertauschungen rot (Abschnitt 3, F-8). Die Rangfolge **innerhalb eines Start-Tags** fehlt (V-2).

**Akzeptierte Negative.** (a) bis (f) und das Negativ „A4 abgetastet“ nennen Grund und Folge. Zwei Anmerkungen: (b) D8 P8 nennt den Grund (der einfache Durchlauf hat die Eigenschaft durch seinen Bau), die Folge im Betrieb nicht (V-13); das Negativ „CR aus dem Leerraum“ ist **zutreffend**, die Tabellenzeile B5 widerspricht ihm aber noch (V-7).

## 3. Nachverifikation der Befunde des Reviews (Klasse, nicht Mutation)

Je Befund: eigene Mutationen über die Klasse. „rot“ heißt: ein Test der Zeile, die die Mutation treffen soll, wird rot.

| Review | Klasse | Eigene Mutationen | Ergebnis |
|---|---|---|---|
| F-1 | Randformen im Code entschieden (Annotation unter jedem Element; Attribute mit Namensraum; `Target`; `Edm`) | Entscheidung steht in `SPEC-125` Nr. 7, 8, 9 e, 11; Code ändert `bindingZiel` und `schema` | geschlossen; der Code stimmt mit dem neuen Wortlaut überein (Proben) |
| F-2 | A4 schwächer als der Vertrag | natürliche Mutationen am Strom-Teil: BOM mit zwei statt drei Bytes entschieden, `FullRune`-Prüfung entfernt, UTF-16-Marke ohne Abwarten, Rest nicht zurückgehalten, `len(tail) > 1`, `aus` ohne Fehler, `>=` bei der Größe, Warten auf UTF-16 entfernt (9) | 8 rot (A4 oder die Zeile des Fehlers), 1 gleichwertig (`halde` vor dem Lesen nicht geleert: jeder Weg überschreibt sie). `make test`: 961 Dokumente bis 5000 Bytes, 951 174 Läufe der Zweiteilungen, jede Position; `make test-race`: Probe, 60 528 Läufe. Grund und Folge des Negativs stehen im Plan |
| F-3 | Zeichenmengen, beide Seiten jeder Grenze | `0x1F`, `0x0B`, CR, Tabulator, LF, `U+FFFE`, `U+FFFF`, `DEL`, `U+0080` bis `U+009F`, `U+10000`, `U+10FFFF`, `U+FFFD`, `U+D7FF`, `U+E000`, `U+00A0` (z1 bis z15) | 15 von 15 rot. Grün sind drei Punkte im Inneren eines Bereichs, den der Vertrag nicht einzeln nennt (`U+FEFF`, `U+0085`, `U+2028` gesperrt): V-12 |
| F-4 | Leerraum nach der Wurzel | Tabulator, Leerzeichen, LF, `U+00A0`, Wegfall der Prüfung auf `&` (w1 bis w4, w6) | 5 von 5 rot. Gleichwertig: `\x0b` im Satz (das Zeichen ist vorher `syntax`) und CR nicht im Satz (Abschnitt 5, „CR“) |
| F-5 | Ersatzreferenz, jede Referenz in jedem Wert | nur die erste Referenz je Token, nur das letzte Attribut, Dezimal als Hex, `x` als Präfix, CDATA-Ausnahme, Text, Bereichsgrenzen `D800` und `DFFF` (e1, e3 bis e5, e8 bis e12) | 9 von 9 rot. **Rest**: die **gültige** Seite der Grenzen zusammen mit einem Zeichen `U+FFFD` im selben Token (`Term="&#xD7FF;"`, `Term="&#xE000;"` neben `U+FFFD`) bleibt grün, V-6 |
| F-6 | doppeltes Attribut | nur ohne Namensraum, lokaler Name als Schlüssel (d1, d3, d4) | 3 von 3 rot |
| F-7 | Funde: Attribute mit Namensraum an vierzehn Orten, Annotation unter zwölf Eltern, fremde Elemente, Konstrukte | 72 Mutationen: `attributFunde` je Aufrufstelle entfernt (14, `Property` und `NavigationProperty` einzeln), Ausnahmen `xml` und `xmlns`, drei Listen von Attributen, fremdes Element mit Attributen und mit Name, `Annotation` je Elternart (12), fremdes Element je Elternart (12), Konstrukt je Elternart (12), Konstrukt je Name (13) | 71 rot, 1 gleichwertig (Wert `true` gegen `!= "false"`: `boolAttr` lässt nur beide zu) |
| F-8 | Paare der Rangfolge | alle 28 Vertauschungen je zweier der acht Durchläufe | 28 von 28 rot (17 durch den Fall des Paares oder von P12 bis P27, 11 durch den Speicherwächter, Abschnitt 5). **Rest** in anderer Klasse: V-2 |
| F-9 | Fixture gleicht dem Wert der Mutation | `keys > 2`, zweites `EntityContainer` nur im selben und nur im anderen Namensraum, Abstract irgendwo in der Kette, `t.keys` statt `len(refs)`, nur Wurzel oder nur direkte Basis zählt, geerbte und eigene Navigation-Property, V3-Namensraum, Property gegen Navigation-Property (k1 bis k12) | 13 von 13 rot. **Rest** auf der Modell-Seite: V-4 |
| F-10 | Namen, `Edm` | `IsNumber`, ASCII-Ziffern, `IsMark`, nur ASCII, Ziffer am Anfang, Länge in Bytes (Name, Namensraum), Grenzen 127, 129, 255, 257, `Edm` ohne Punkt, `HasPrefix("Edm")`, `EqualFold` (Namespace, Alias), `Edm` als Alias, Glieder ungeprüft, leerer Name (n1 bis n20) | 20 von 20 rot. Grün: nur die Kurzform `Edm.` mit weniger als zwei Zeichen dahinter (künstlich, V-12) |
| F-11 | Kontextfehler × Stelle | jede der drei Stellen (vor dem Aufruf, im Lesen, vor dem Auflösen) mit `Canceled` und mit `DeadlineExceeded` ersetzt, `errors.Join` um den Fehler (c1 bis c4) | 7 von 7 rot |
| F-12 | Konstruktor kopiert vier Eingaben | `namespaces`, `sets`, `bindings`, Typen, `Keys`, `Properties`, `NavigationProperties`, dazu die fünf Zugriffe (f1 bis f12) | 12 von 12 rot; gleichwertig oder künstlich: zweiter Indexeintrag, Prüfung auf leeren Namen |
| F-13, F-14 | Ring über 64 Typen; `Edm.Sub` | `kette` vor der Wiederholung, Grenze 63 und 65, Anfang des Zyklus (r1 bis r4); `Edm.` ohne Punkt, nur `Edm`, nur `Edm.` (n12 bis n15) | rot; Spezifikation Nr. 9 a und Nr. 8 b tragen die Entscheidung |
| F-15, F-16 | Plan folgt der Korrektur | Kopf, §1, §6 Abw. 1 bis 13 | erfüllt (Abschnitt 6); der Name `max` steht noch im Code: V-9 |

## 4. Gates und Belege

* `make gates`: Rückgabecode 0 auf `cf3131d`. `a-check`: 0 Befund (fünf Dateien unter `tools/` liegen in keiner Schicht, wie vorher). `baseline-verify`: OK. `docs-check`: 240 Dateien, 0 Befund (vor diesem Bericht). Stufen `build`, `test`, `race`, `lint`: die Schichten von `Dockerfile` kamen aus dem Zwischenspeicher, weil der Quellstand derselbe ist wie beim Lauf des Implementers; `lint` lief und blieb ohne Befund. Ich habe die Suite zusätzlich selbst gefahren: ohne Detektor `ok csdl 17,7 s`, mit Detektor `ok csdl 102,7 s` (Messlauf mit anderem Wächter-Wert, sonst unverändert), `model` und `outbound` ok.
* Kein `//nolint` in `src/adapters/driven/odata`, `src/hexagon/model`, `src/hexagon/ports/outbound` (Suche).
* **J1** (Fänger der Zeile: benannte Ausgabe von `make a-check`): in einer Kopie `encoding/xml` importiert in `src/hexagon/model` → `core-impurity`, in `ports/outbound` → `port-impurity`, in `services/session` → `app-impurity`, im SQLite-Adapter → `tech-leak`, im Paket `csdl` kein Befund. Die Regel ist tragend (Review I-7), aber kein stehender Fänger: V-8.
* **Abdeckung unter dem Detektor**, `-run 'TestVertrag/^(I2|I3)_'`, `-coverpkg` auf `csdl`: 99,7 %. Nicht ausgeführt sind `lesen.go:70` (`io.ErrNoProgress`) und `strom.go:27` (`stromFehler.Error`). Das deckt sich mit dem Plan.
* **Sendungen** (`AGENTS.md` §3.13): `slice-baseline-import`, `slice-bootstrap-m2`, `slice-odata-codec-json` und `slice-protokolltreue` nennen `slice-odata-codec-csdl`; die Sendung von `slice-protokolltreue` (Markierung unbekannter Erweiterungen) liegt in DoD Punkt 3 und hat ihre Fänger (G1).

## 5. Mutationen und Raten

### Mutationen

Gruppe · Zahl · rot · grün. Grün heißt hier: beobachtbar und ohne roten Test, außer als gleichwertig benannt.

| Gruppe | Zahl | rot | grün |
|---|---|---|---|
| Zeichenmengen (z) | 18 | 15 | 3 (z16 `U+FEFF`, z17 `U+0085`, z18 `U+2028`; V-12) |
| Leerraum nach der Wurzel (w) | 7 | 5 | 2 gleichwertig (`\x0b`; CR) |
| Referenzen (e) | 12 | 9 | 3 (e2 nur `Term` geprüft, künstlich; e6, e7 V-6) |
| doppelte Attribute (d) | 3 | 3 | 0 |
| Funde G1 (g, an, fr, ko, kn) | 72 | 71 | 1 gleichwertig |
| Rangfolge: 28 Vertauschungen (sw) | 28 | 28 | 0 |
| Schlüssel, Container, Ketten (k) | 13 | 13 | 0 |
| Namen (n) | 21 | 20 | 1 künstlich (n15b) |
| Kontext (c) | 7 | 7 | 0 |
| Kette, Summe (r, m) | 6 | 6 | 0 |
| Modell-Konstruktor und Zugriffe (f) | 14 | 12 | 2 gleichwertig oder künstlich |
| Grenzen entfernt (g1 bis g7) | 7 | 7 | 0 |
| Strom und Zerlegung (a4) | 9 | 8 | 1 gleichwertig |
| Start-Tag und Element, Reihenfolge (s) | 7 | 0 | **7** (V-2) |
| Auflösen (q) | 6 | 2 | **4** (q3a, q3b V-5; q5 V-3; q14 V-4) |
| Modell-Felder (mt) | 12 | 9 | **3** (mt7 `Path`, mt10 `HasStream` → `Open`: V-4; mt11 Leerraum in `BaseType`: V-5) |
| Raten-Mutanten (i1, i2, i3) | 10 | 10 | 0 |

Vier weitere Kopien (n5, n15 vor der Ersetzung, q16, a4j) waren Übersetzungsfehler; n5 und n15 sind ersetzt, q16 und a4j nicht gewertet. Alle Mutationen, die eine Grenze, die Tiefe oder die Größe entfernen (g1 bis g7, a4g, m1), liefen mit `--memory=2g` und endeten in Sekunden rot (79 s für alle sieben zusammen); kein Lauf wurde vom Host abgebrochen.

### Raten (Zahl der Prozesse, Messender: der Verifier, Host mit 20 Kernen, Testbinary der Kopie, `-test.count=1`, `--cpus` wie angegeben)

| Zeile | Mutation | Bedingung | Prozesse rot | Unmutiert rot |
|---|---|---|---|---|
| I1 | `EntityTypes` über Map | ohne Detektor, `--cpus=3`, Zeilen I1, F1, F3, F4, F6, B5, G4 | 30 von 30 (I1 in 30) | 0 von 30 |
| I1 | `Properties` über Map | wie oben | 30 von 30 (I1 in 30) | 0 von 30 |
| I1 | `Funde` über Map | wie oben | 30 von 30 (I1 in 30, G4 in 28) | 0 von 30 |
| I1 | `EntitySets` über Map | wie oben | 30 von 30 (I1 in 30) | 0 von 30 |
| I1 | `NavigationBindings` über Map | wie oben | 25 von 30, **I1 nie** (F4 in 24, F6 in 12) | 0 von 30 |
| I1 | `Keys` über Map | wie oben | 14 von 30, **I1 nie** (nur F3 in 14) | 0 von 30 |
| I1 | `Namespaces` über Map | wie oben | 26 von 30, **I1 nie** (nur F4 in 26) | 0 von 30 |
| (ganze Suite) | `Keys`, `NavigationBindings`, `Namespaces` über Map | ohne Detektor, `--cpus=3`, alle Zeilen | je 12 von 12 (A4 und I3 oder I2 in jedem Prozess) | — |
| I2 | Variable des Pakets (`zaehler++`) | Detektor, `--cpus=3`, Zeilen I2, I3 | 15 von 15 | 0 von 15 |
| I2 | Map des Pakets (`cache[grenze]++`) | wie oben | 15 von 15 (in 10 benennen I2 und I3 den Fehler, in 5 endet der Prozess ohne benannten Test) | 0 von 15 |
| I3 | Funde als Variable des Pakets unter Sperre mit `Gosched` zwischen den Abschnitten | ohne Detektor, `--cpus=3` | 30 von 30 (I3 in 30) | — |
| I3 | wie oben | ohne Detektor, `--cpus=0.3` | 10 von 10 | 0 von 30 (`--cpus=0.3`) |
| A4 | BOM mit zwei statt drei Bytes entschieden (a4a); Rest am Lesenende nicht zurückgehalten (a4e) | ohne Detektor, `--cpus=3`, Zeile A4 | je 8 von 8 | 0 von 10 (A4, D8, B5, B4) |
| D8 | Vertauschung `basen`/`ketten` (Speicherwächter) | wie oben, Zeile D8, `--memory=2g` | 10 von 10 | 0 von 10 |
| D8 | Vertauschung `summe`/`vererbung` (Fall des Paares) | wie oben | 10 von 10 | 0 von 10 |
| B5 | Prüfung auf `&` nach der Wurzel entfernt (w6) | wie oben | 10 von 10 | 0 von 10 |
| B4 | nur die erste Referenz je Token (e3) | wie oben | 10 von 10 | 0 von 10 |
| `TestTransaktionVertrag` (Paket `sqlite`) | keine | Detektor, 40 Prozesse `--cpus=0.5`, 40 Prozesse `--cpus=3` | 0 von 80 | — |

Die Reihen sind klein und sagen „nie grün gesehen“, nicht „immer rot“; alle Fänger der Folgerunde sind deterministisch.

### CR „aus dem Leerraum genommen“ (Frage des Auftrags)

Die Gleichwertigkeit ist **richtig**. Mutation w7 (`bytes.Trim(t, " \t\n")`, kein CR im Satz): die ganze Suite bleibt grün (18,6 s), und Probe-Dokumente mit rohem CR vor und nach der Wurzel, CR LF und CR um die Wurzel liefern bei Code und Mutant dasselbe. Der Decoder macht jedes rohe CR in `CharData` zu LF, auch in `CDATA`. Die zweite Beobachtungsstelle, `&#xD;`, ist seit `7e08438` über das Zeichen `&` in den Bytes `syntax` und kommt nicht mehr als Leerraum an. Ein Fall kann die Mutation also nicht fangen; die erste Fassung von `684602d` irrte, der Implementer und der Plan (Absatz *Gleichwertig*, Punkt (a)) sind richtig. Die Folge ist benannt (ein anderer Adapter, der CR nicht zuließe, fiele an den gültigen Fällen „CR um die Wurzel“ und „CR nach der Wurzel“ auf). **Nicht nachgezogen** ist die Tabellenzeile B5, die die Mutation noch „rohes CR rot“ nennt (V-7).

### Speicherwächter und Detektor-Probe (Fragen des Auftrags)

* **Spielraum.** Der Wächter bricht bei 512 MiB Heap ab. Der gesunde Lauf ohne Detektor bleibt zwischen 128 und 192 MiB (bei 128 MiB Abbruch nach 0,6 s bei 136 MiB, bei 192 MiB grün); mit Detektor unter 256 MiB (grün, 102,7 s). Der Abstand zur Grenze ist mindestens das 2,7- bzw. 2fache: kein falscher Alarm auf diesem Host zu erwarten.
* **Ersetzt er einen Fänger?** Für die 11 von 28 Vertauschungen, die ihn auslösen (`basen` mit `ketten`, `summe`, `vererbung`, `typenAufloesen`, `setsAufloesen`; `ketten` mit den sechs späteren), gilt: an den Dokumenten mit Zyklus (alle Paare mit `a2`, P19, D3) ist der Ausgang **kein falscher Wert, sondern eine Schleife ohne Ende** (`kette(t)` auf einem Zyklus, bevor `ketten` ihn meldet). Ein Fänger kann dort nur „der Aufruf endet nicht“ sein; der Wächter macht daraus nach etwa einer Sekunde einen Abbruch mit Rückgabecode 1, statt nach der Zeitgrenze des Laufs. Er ersetzt dort keinen Fänger, der einen Wert prüfen könnte. Ob einzelne der elf auch ohne die Schleife an einem Wertfall rot wären, zeigt der Lauf nicht: der Abbruch verdeckt es, und `lies` fährt alle Dokumente einer Zeile, bevor `-run` einen Unterfall auswählt (ich habe es mit einem Filter versucht; der Wächter schlug trotzdem zu). Er hat zwei Grenzen: (1) Der Abbruch benennt keine Zeile. Die Meldung sagt „die Stapel aller Goroutinen folgen“, der Stapel der laufenden Goroutine ist aber „stack unavailable“; das Ergebnis ist ein Abbruch unter dem Namen `TestVertrag/D8_Rangfolge` ohne das Dokument (V-10). (2) Ein Abbruch verdeckt jeden anderen roten Fall desselben Prozesses, die Meldung der Zeilen davor geht verloren.
* **Detektor-Probe von A4.** Unter `make test-race` fährt A4 für Dokumente bis 5000 Bytes die ersten und letzten acht und 16 verteilte Positionen (60 528 Läufe), unter `make test` jede Position (951 174 Läufe, Zählung gegen Dokumente × (L−1) × 2 gleich). Grund und Folge stehen im Plan (Zeile A4 und Absatz *Akzeptierte Negative*): eine Zweiteilung hat keinen Zustand, den der Detektor sähe; die vollständige Menge fährt `make test`. Beide laufen in `make gates`. Offen bleibt, wie bei jeder Probe, was das Negativ benennt: ein Fehler nur in der Mitte eines Dokuments über 5000 Bytes. Ich habe natürliche Mutationen dieser Art gesucht (a4a bis a4i) und keine gefunden, die erst ab 5000 Bytes sichtbar wird.

## 6. Plan folgt dem Code, Größe, Abweichungen

**§3.9.** Kopf (*Berührte Spec-Stellen*), §1 (Parametername `grenze`, Folgerunde), §3 (Tabelle, Prüfgrenzen, Ausgang des Implementers, Folgerunde, Größe nach beiden Handoffs), §6 (Abweichungen 1 bis 13 mit Ausgang) und die Änderungszeile der Spezifikation sind nachgezogen. Zwei Stellen nicht: die Tabellenzeile B5 (V-7) und die Namen `max` im Code (V-9).

**Abweichungen 1 bis 13.** Jede trägt in §6 einen Punkt mit Ausgang (1 bis 8 einzeln; 9 und 10, 11 und 12 je zusammen; 13 einzeln). Ich habe die Behauptungen „Code bleibt“ und „Code geändert“ am Code geprüft: Punkt 3 (`bindingZiel` liefert `<qualifizierter Container>/<Name>`, Proben und F6), Punkt 10 (`schema` weist `Edm.` ab), Punkt 13 (Referenz außerhalb der Wurzel: `&` im `CharData`-Token der Wurzelebene) stimmen. Punkt 13 deckt `CDATA` nicht, V-1.

**Akzeptierte Negative.** Geprüft in Abschnitt 2. Das Negativ „Typed-nil“, „Pfad eines Bindings“, UTF-16 liegt in `SPEC-125` *Ungeprüft* mit Grund und Folge.

**Größe.** `git diff --numstat 3b6b3b8..7e08438` über `src/hexagon/model`, `src/hexagon/ports/outbound`, `src/adapters/driven/odata/csdl`, `.a-check.yml` (`7e08438` ist der letzte Commit mit einem Pfad der Marke; `cf3131d` ändert nur den Plan): **4198 Einfügungen, 1 Löschung, 4199 Zeilen** gegen die Marke 2400 (+75 %). Produktcode 1505, Tests und Fixtures 2694. Das ist die Zahl des Plans; der Plan beschönigt sie nicht (Handoff 3628, Folgerunde +571, Messung mit Metrik und Pfaden, kein Kürzen, die Rückführung als Entscheidung des Planners benannt). Die Messung nach dem Review, die die Closure nach §4 nachholt, fällt auf diese Zahl plus die Nachbesserung zu V-1 bis V-4.

## 7. Offen für die Closure (Planner)

1. **Größe**: 4199 gegen 2400 in §4 eintragen, die Überschreitung begründen (Tests 2694 gegen 48 Zeilen je Tabellenzeile in der Schätzung; die Formel kennt den Produktcode eines Parsers nicht), Risiko „Größe“ in §6 mit Ausgang; `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` ist ein weiterer Beleg. Die Rückführung (`in-progress` → `next`) ist die Entscheidung des Planners.
2. **Tabellenzeile B5** (V-7) und die erledigten Teile des Plans in einem Commit mit dem Fänger zu V-1 bis V-4 nachziehen.
3. **Register** (`docs/plan/planning/observations/`): `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (V-3, V-4, V-5, V-6 sind vier weitere Belege der Klasse nach dem Fix zu F-5, F-9, F-10, F-12: die Fixtures der Folgerunde decken die genannte Mutation, nicht die Form der Eingabe), `BEO-REPO/randform-erst-im-review-entschieden` (V-1, V-2), `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (hier kein neuer Beleg: 99,7 %, Rest benannt), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (kein Beleg: alle Raten deterministisch), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (V-13).
4. **Risiko-Ausgänge** (§6): Bibliothek *entfallen* (`encoding/xml` genügt); `SetMetadata` *eingetreten* (beide Nehmer tragen die Sendung); Sendung `slice-protokolltreue` *eingetreten* (DoD Punkt 3, G1); Eingabe aus einem fremden Dienst, Randformen, Prüfgrenzen: *weiter offen* bis V-1 bis V-3 behoben sind; Stabilität der Reihenfolge *entfallen* (I1 mit den Raten oben, V-11); Größe *eingetreten*; Bericht des Implementers *entfallen* (der Absatz liegt im Plan).
5. **Lerneintrag** (§7) mit den Finding-Klassen aus Review und Bericht: *Fänger führt nur eine Form der Eingabe* (V-3 bis V-6, nach dem Fix zu F-5, F-9 bis F-12), *Randform im Code entschieden* (V-1, V-2), *Plan folgt nicht der Korrektur* (V-7, V-9).
6. **Move nach `done/`** von Hand erst nach Nachbesserung und Nachverifikation (zwei Commits nach `AGENTS.md` §3.3); `.claude/agents/architect.md` bleibt beim Nutzer.

## Findings

| ID | Schwere | Ort | Aussage | Ausfallszenario | Quelle | Adressat |
|---|---|---|---|---|---|---|
| V-1 | MEDIUM | `src/adapters/driven/odata/csdl/lesen.go:99` (`case xml.CharData:` bei leerem Stapel); `SPEC-125` Nr. 4; Fänger B5 Menge A und B | Eine `CDATA`-Sektion vor oder nach der Wurzel wird gelesen, wenn sie leer ist oder nur Leerraum enthält: `<![CDATA[ ]]>` und `<![CDATA[]]>` vor und nach der Wurzel sind gültig (Proben `cdata_vor`, `cdata_nach`, `cdata_leer_vor`, `kom_cdata`). Der Vertrag sagt „wohlgeformtes XML 1.0“ und „Davor und danach sind Leerraum, Kommentare und Verarbeitungsanweisungen zulässig“; XML 1.0 lässt dort keine `CDATA` zu. Der Decoder liefert den Inhalt als `CharData`, die Prüfung sieht nur Text und `&`; die Fixtures führen `CDATA` nur innerhalb der Wurzel (Menge A von B5, B4). Gleiche Klasse wie die Zeichenreferenz auf Leerraum, die `7e08438` schloss | Ein Dokument, das ein wohlgeformtes XML 1.0 verwirft, wird importiert; ein zweiter Adapter, der die Suite fährt, darf es verwerfen, der Leser nicht: der Vertrag ist für beide gleich. Der Schaden ist klein (kein Datenverlust), die Zusage ist falsch | `AGENTS.md` §3.15, `SPEC-125` Nr. 4 | Implementer (Prüfung: das Token beginnt mit `<![CDATA[`; Fänger: Menge B um {`CDATA` leer, mit Leerraum, mit Text} × {vor, nach der Wurzel}, Mutation „Prüfung entfällt“ rot sehen); Architect (Nr. 4 nennt `CDATA` außerhalb der Wurzel ausdrücklich) |
| V-2 | MEDIUM | `lesen.go:156-176` (`start`: Tiefe, `doppelt`, `wurzel`, `kind`) und `lesen.go:71-77` (`refs` vor `start`); `elemente.go:280-306` (`member`: Name, Type, Wahrheitswert, doppelter Name) und `entityType`; `SPEC-125` Nr. 8, 9 und Zusatztabelle | Mehrere Verstöße im selben Start-Tag oder Element liefern eine Ursache oder einen Ort, den nur die Reihenfolge der `if` im Code bestimmt. Proben: Element in Tiefe 65 mit doppeltem Attribut → `tiefe`, mit Ersatzreferenz → `syntax`; `Edmx Version="3.0" Version="4.0"` → `syntax` (nicht `version`); `Property Name="1" Name="x"` → `syntax`; `Property` mit `Nullable="x"` und fehlendem `Type` → Ort `Ns.T/p@Type`; doppelter Name mit ungültigem `Nullable` → Ort `…@Nullable`; Typ doppelt mit ungültigem `Abstract` → `Ns.T@Abstract`. Die Mutationen s1, s2, s5, s6 (Reihenfolge im Start-Tag) und s7, s8, s9 (Reihenfolge im Element) bleiben alle grün. Die Zusatztabelle führt P1 bis P27 nur für Verstöße an **verschiedenen** Stellen des Stroms (Byte-Reihenfolge) und die acht Durchläufe des Auflösens; `AGENTS.md` §3.16 verlangt eine Zeile „je Paar von Verstößen am selben Schlüssel“ | Zwei Fassungen des Lesers (oder ein Umbau) melden für dasselbe Dokument verschiedene Ursachen und Orte; der Ort steht im Fehlertext des Imports, die Diagnose hängt von einer Reihenfolge, die niemand entschieden hat | `AGENTS.md` §3.12, §3.16 | Architect (Rangfolge innerhalb eines Start-Tags und eines Elements entscheiden und als Zeilen der Zusatztabelle führen); Implementer (Fänger, Mutationen s1 bis s9 rot sehen) |
| V-3 | MEDIUM | `aufloesen.go:29` (`strings.LastIndex(s, ".")`); Fixtures `testdata/*.xml` (`Demo`, `Zoo`, `Shop`) und die Fälle der Suite (`Ns`, `A`, `B`) | Kein gültiges Dokument der Suite hat einen Namensraum mit Punkt, auf den eine Referenz zeigt. Mutation q5 (`strings.Index`, erster Punkt): die ganze Suite bleibt grün (18 s). Probe: `Microsoft.OData.Models.TripPin` mit `BaseType`, Property-Typ, Navigation-Typ und Entity-Set ist beim Code gültig, bei q5 `nicht_aufloesbar`. Nr. 9 und Nr. 8 b sagen „Ein Name vor dem **letzten** `.`“ und „jedes Glied eines Namensraums, durch `.` getrennt“; die Fälle von D6 prüfen nur die **Form** solcher Namen | Eine Änderung an `qualify` bricht jeden realen Dienst (die Namensräume realer Dienste haben Punkte), und die Suite zeigt es nicht. Genau die Form der Eingabe, die in der Praxis die Regel ist | `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe) | Implementer (Fixture: ein Namensraum mit mindestens drei Gliedern, Referenzen aller sechs Arten über Namensraum und Alias; Mutation q5 rot sehen) |
| V-4 | MEDIUM | `aufloesen.go:289` (`modellTyp`: `if len(c.refs) > 0`); `aufloesen.go` Zeile `Path: b.path`; `elemente.go` Zeile `open: b[1]`; Zeilen F2, F6, D7 | Drei Zusagen von Nr. 11 haben keinen Fänger auf der Seite des Modells. (a) Ein leerer `<Key/>` in der Ableitung bei Schlüssel in der Basis ist als **gültig** getestet (F-9 d), aber nicht, dass `Keys` der Ableitung die der Basis bleiben: Mutation q14 (`c.keys > 0` statt `len(c.refs) > 0`) setzt `Keys` der Ableitung leer, grün (Probe `d4_leerer_key_abgeleitet`: `Ns.D[]`). (b) `Path` „bleibt der Wert des Dokuments“: `strings.TrimSpace` darauf (mt7) bleibt grün; die Mutation des Implementers („`Path` normalisiert“) fasst anderes an. (c) `OpenType` und nur `OpenType` macht `Open` wahr: `HasStream="true"` setzt `Open` (mt10), grün; die Fälle von G1 prüfen `HasStream` nur im Bericht | Das Modell eines gültigen Dokuments ist falsch (Schlüssel einer Ableitung leer), ohne dass ein Test es sieht; `slice-odata-codec-json` und `slice-baseline-import` lesen `Keys` | `AGENTS.md` §3.15 | Implementer (je ein Fall: Modell der Ableitung mit leerem `Key`; `Path` mit Leerraum an beiden Seiten; Typ mit `HasStream` und ohne `OpenType` mit `Open` falsch) |
| V-5 | LOW | `aufloesen.go:144` und `:156` (`strings.HasSuffix(s, ")")`); Zeile D5, D2 | Zwei Formen der Eingabe fehlen den Fängern der Referenzen. (a) „Offenes `Collection(`“: das Fixture `Collection(Edm.String` kann die Mutation „Suffix `)` nicht geprüft“ nicht sehen (sie schneidet das letzte Zeichen des Namens ab, `Edm.Strin` bleibt unauflösbar). Mit `Collection(Edm.StringX` (Property) und `Collection(Ns.TX` (Navigation-Property) liefert der Mutant gültig (q3a, q3b grün), der Code `nicht_aufloesbar` (Proben `d2`, `d3`). Fixture gleicht dem Wert der Mutation. (b) Leerraum in Referenzen: nur der Typ einer Property hat den Fall; ` Ns.T` als `BaseType` (mt11, `TrimSpace`) bleibt grün, ebenso `EntityType` eines Sets, Typ einer Navigation-Property, `Target` und `Name` einer `PropertyRef` | Ein Dokument mit Rest hinter der Klammer wird wie ein geschlossenes behandelt; ein `TrimSpace` an einer der fünf Stellen verändert den Vertrag still | `AGENTS.md` §3.15 | Implementer (Fälle: `Collection(Edm.StringX`, `Collection(Ns.TX`; Leerraum je Referenzart in der Schleife über die sechs Arten von D2) |
| V-6 | LOW | `lesen.go:150` (`n >= 0xD800 && n <= 0xDFFF`); Zeile B4, `faelleB4Ersatz` | Rest von F-5: die gültige Seite der Grenzen der Ersatzzeichen ist nur **ohne** ein `U+FFFD` im Token geprüft, aber die Suche in den Bytes läuft nur, wenn der Decoder ein `U+FFFD` geliefert hat. Mutation e6 (`n >= 0xD7FF`) und e7 (`n <= 0xE000`) bleiben grün; Probe: `Term="…&#xD7FF;"` und `Term="…&#xE000;"` neben einem `U+FFFD` sind beim Code gültig, beim Mutanten `syntax` (`ref1_fffd_d7ff`, `ref2_fffd_e000`) | Ein gültiges Dokument mit `U+FFFD` und `&#xD7FF;` im selben Wert würde verworfen | `AGENTS.md` §3.15 | Implementer (Fälle {`U+FFFD` + `&#xD7FF;`, `U+FFFD` + `&#xE000;`} in `faelleB4Ersatz`) |
| V-7 | LOW | Plan §3, Zeile B5, letzte Mutation: „CR aus dem Leerraum genommen (rohes CR rot)“; Absatz *Gleichwertig und nicht beobachtbar* (a) | Die Zeile nennt die Mutation noch als rot, der Plan führt sie an anderer Stelle als gleichwertig (Abschnitt 5 bestätigt: unbeobachtbar). Die Berichtigung erreichte nicht jede Stelle, die die Aussage trägt | Die nächste Änderung sucht den Fänger, der nicht existiert | `AGENTS.md` §3.9 | Architect (Zeile B5) / Planner |
| V-8 | LOW | `.a-check.yml` (`{pattern: "encoding/xml", adapter: "src/adapters/driven/odata"}`); `tools/gegenproben/faelle.go` (`a-check/bibliothek-in-fremdem-adapter` hält nur `modernc.org/sqlite`); Zeile J1 | Die Regel ist tragend (Abschnitt 4), aber ihr Fänger ist ein einmaliger Lauf in einer Kopie. Wird die Zeile aus `.a-check.yml` entfernt, bleiben `make gates` und `make gegenproben` grün: eine Lockerung des Gates (`AGENTS.md` §3.6) ohne roten Fänger | Ein späterer Commit streicht die Regel (etwa beim Umbau der Schichten), und `encoding/xml` kann in den Kern wandern, ohne dass ein Gate es meldet | `AGENTS.md` §3.15, §3.6 | Implementer oder Planner (ein Fall der Gegenprobe `a-check/xml-im-kern` mit `WantText: core-impurity`, oder ein benanntes Negativ mit Grund und Folge) |
| V-9 | LOW | `src/adapters/driven/odata/csdl/leser.go:19` („grenze ist `max`“) und Fehlertext `max kleiner 1` (`leser.go:28`); `src/hexagon/ports/outbound/metadata_codec.go:12` („grenze ist `max` in SPEC-125“) | Die Spezifikation führt seit `f7639b9` allein `grenze`; zwei Kommentare nennen den abgelösten Namen (`AGENTS.md` §3.7: beschreibt abwesenden Text) und der Fehlertext von `ErrUngueltig` nennt `max`. Die Berichtigung erreichte nicht jede Stelle | Der Leser des Ports sucht `max` in der Spezifikation und findet es nicht | `AGENTS.md` §3.7, §3.9 | Implementer |
| V-10 | LOW | `src/hexagon/ports/outbound/outboundtest/metadata_codec_vertrag.go:39-58` (`speicherLimit`, `speicherwaechter`) | Die Meldung des Wächters sagt „die Stapel aller Goroutinen folgen“; der Stapel der Goroutine, die schleift, ist „goroutine running on other thread; stack unavailable“ (Lauf `sw_basen_ketten`). Die Meldung benennt weder Zeile noch Dokument; die Zusage des Textes ist stärker als das, was geschieht (`AGENTS.md` §3.11) | Wer einen roten Lauf des Wächters liest, findet nicht, welcher Fall schleift (elf der 28 Vertauschungen enden so) | `AGENTS.md` §3.11 | Implementer (Meldung: den Namen der laufenden Zeile mitgeben, etwa über eine Variable, die `t.Run` setzt; oder den Satz streichen) |
| V-11 | INFO | Zeile I1; Review I-6 | Reihenfolge von `NavigationBindings`, `Keys` und `Namespaces` hat keinen deterministischen Fänger auf Zeilenebene: I1 wird bei diesen drei Mutationen nie rot; F4 (26 von 30) , F6 (12), F3 (14 von 30) fangen sie mit Wahrscheinlichkeit; die ganze Suite ist 12 von 12 rot (A4, I2, I3 vergleichen Läufe untereinander). Review I-6 gilt weiter; kein neuer Befund | keines, solange die Suite als Ganzes läuft | `AGENTS.md` §3.10 | Planner (Zeile I1 oder Lerneintrag nennt den Aggregat-Fänger) |
| V-12 | INFO | `strom.go:xmlZeichenGueltig`; `elemente.go:schema` | Punktmutationen im Inneren genannter Bereiche bleiben grün und sind vom Vertrag nicht einzeln genannt: `U+FEFF`, `U+0085`, `U+2028` im Kommentar gesperrt (z16 bis z18); `Edm.` mit nur einem oder zwei Zeichen dahinter (n15b); zweiter Indexeintrag im Modell (f14); Ersatzreferenz nur im Attribut `Term` geprüft (e2). Kein Fänger zählt Punkte, die kein Vertragstext nennt | keines für die Zusage | `AGENTS.md` §3.15 | — |
| V-13 | INFO | Plan §3 *Akzeptierte Negative, die der Implementer fand* (b) | Das Negativ zu D8 P8 und „P9 Syntax, dann Tiefe“ nennt den Grund, nicht die Folge im Betrieb (§3.15 verlangt beides). Die Folge ist klein (ein Leser mit Vorprüfung der Namen meldete `struktur` statt `syntax`); sie gehört in einen Satz | keines | `AGENTS.md` §3.15 | Planner |
| V-14 | INFO | `src/adapters/driven/sqlite/repositories_test.go`, `TestTransaktionVertrag` (Review I-5) | Das einmalige Rot unter Last, das der Implementer meldet, ist nicht reproduzierbar: 0 von 80 Prozessen rot (40 bei `--cpus=0.5`, 40 bei `--cpus=3`, mit Detektor, parallel zu zwei weiteren Läufen). Zusammen mit den 80 Prozessen des Reviews 0 von 160 | keines nachgewiesen | `AGENTS.md` §3.10 | Planner (Register, wenn es wiederkehrt) |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Sicherheit des Lesers | `DOCTYPE` in jeder Form ist `doctype` bevor etwas gelesen wird; keine Entität wird expandiert; `CharsetReader` liefert nie einen Reader; Tiefe, Summe, Kette, Größe durchgesetzt (g1 bis g7 rot); Probe mit TripPin-artigem Dokument, Alias und Punkt-Namensraum ohne falsche Ablehnung |
| Fehlerform und Abbildung | `Error()`, `Unwrap`, `Code`; H2 (Review) rot; `ErrMetadatenUngueltig` → `ODR-E6004` bei jedem Gegenstand |
| Nebenläufigkeit | Leser ohne Zustand (`TestLeserOhneZustand`); I2 15 von 15 rot, I3 30 von 30 und 10 von 10 rot, unmutiert 0 von 45 bzw. 0 von 30 |
| Prüfgrenzen | beide Seiten jeder Grenze: Größe, Tiefe, Summe, Kette, Name (Mutationen a4h, g1 bis g7, m1, m2, n7 bis n11, r1 bis r4 rot) |
| Hard Rules `AGENTS.md` §3.1 bis §3.4, §3.13 | kein lokaler Toolchain-Aufruf (alles über `make` und `docker run`), kein `//nolint`, `spec/architecture.md` ohne Slice, Welle, Hash oder ADR-Bezug, Sendungen getragen |
| Kommentare `AGENTS.md` §3.7, §3.11 | Produktcode und Tests tragen Zusage, Kopplung oder Abgrenzung; Ausnahmen V-9 und V-10 |
| Commits | jede Kopfzeile nennt `LH-FA-06` und `LH-RB-01` |

## Verdikt

**Nicht bestanden.** Die DoD ist dem Wortlaut nach erfüllt und die Gates sind grün, aber vier Befunde gehören zu den Klassen, die der Slice selbst nennt (*Randform im Code entschieden*, *Fänger führt nur eine Form der Eingabe*): V-1 ist ein Verstoß des Codes gegen den Vertrag, V-2 eine Randform ohne Zelle und Fänger, V-3 und V-4 sind Lücken, die eine Änderung an zentralem Code unbemerkt ließen. Rückgabe an den **Architect** (V-1 Wortlaut, V-2 Entscheidung, V-7) und den **Implementer** (V-1 bis V-6, V-8 bis V-10); der **Planner** zieht V-7, V-11 bis V-14 und die Größe nach. Nach der Nachbesserung genügt die Nachverifikation von V-1 bis V-4; die Raten und das Raster der übrigen Zeilen sind gemessen und müssen nicht wiederholt werden. `make gates` läuft nach dem Commit dieses Berichts.
