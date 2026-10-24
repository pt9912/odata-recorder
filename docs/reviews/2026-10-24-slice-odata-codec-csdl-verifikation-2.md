# Nachverifikation: slice-odata-codec-csdl

* Datum: 2026-10-24
* Rolle: verifier (frischer Kontext, nicht der Verifier der ersten Runde; kein Produktcode, kein Test, keine Spezifikation, kein Plan geändert; `.claude/agents/architect.md` ist eine fremde, uncommittete Änderung und blieb unberührt)
* Gegenstand: Spitze `dfc5822`. Seit der ersten Verifikation (`16999b9`): Architect `83b6b8b`; Implementer `4c7019d` (Code und Fänger), `58d7de3` (Gegenprobe Fall 37 und Stempel), `dfc5822` (Plan §3).
* Plan: [`slice-odata-codec-csdl.md`](../plan/planning/in-progress/slice-odata-codec-csdl.md); erste Verifikation: [`2026-10-24-slice-odata-codec-csdl-verifikation.md`](2026-10-24-slice-odata-codec-csdl-verifikation.md)
* Maßstab: die zehn Befunde V-1 bis V-10 **für ihre Klasse**, nicht für die genannte Mutation; `SPEC-125` Nr. 4 b und d, 6, 8 f, 9, 10, 11; `SPEC-102` Fall 37; `AGENTS.md` §3.9 bis §3.16.
* Belege selbst gefahren: `make gegenproben` mit eigenem Stempel (37 Fälle, 0 rot) und in einer Wegwerfkopie ohne die Zeile `encoding/xml` (37 Fälle, 1 rot); die Suite des Slice ohne und mit Detektor; rund 100 eigene Mutationen als Wegwerfkopien von `git archive HEAD` unter dem Scratchpad (Quelle `:ro`, `--network none`, `--memory` 2 bis 4 GiB, höchstens drei Container zugleich, Ersetzung mit einem eigenen Perl-Skript, nie im Repo); ein eigenes Probe-Paket (rund 4000 Dokumente aus Generatoren nach `SPEC-125`, im Wegwerf-Baum als Test neben dem Leser); Raten als Testbinary mit N Prozessen zu `-test.count=1`. `git status --short` zeigt vor und nach dem Lauf nur `.claude/agents/architect.md`.

## Urteil

**Bestanden: kein HIGH, kein MEDIUM, vier LOW, vier INFO neu.** Die Auflagen der ersten Verifikation (V-1 bis V-4, V-7) sind erfüllt: V-1 bis V-4 sind für ihre Klasse geschlossen, nicht nur für die genannte Mutation. Der Leser verhält sich in allen Klassen des Auftrags nach `SPEC-125`: rund 4000 Probe-Dokumente (CDATA an 15 Stellen außerhalb der Wurzel, jedes Paar der Verstöße im Start-Tag an 20 Rahmen von Elementen und in Tiefe 65, neun Namensräume mit Punkten in sechs Positionen und Alias, Keys, Open, HasStream und Path in Ketten, 7 Leerraumarten in 6 Referenzarten, über 30 Klammerformen, 21 Verweise auf Ersatzzeichen in 15 Positionen) liefern am Code durchweg die Ursache und den Ort der Spezifikation. Offen sind **Lücken in den Fängern**, keine im Code: V-5 ist teilweise offen (LOW), V-8 ist für die genannte Mutation geschlossen, aber nicht für die Aufweichung der Regel (LOW), und die Klasse „Verweise in Namensräume mit dem Anfang `Edm`“ hat keinen Fänger (LOW). Die Größe (4333 gegen Marke 2400) ist die Entscheidung des Planners (Abschnitt 5).

| V | Stand | Beleg |
|---|---|---|
| V-1 | **geschlossen** | 7 Mutationen der Prüfung (entfällt, nur nach, nur vor, nur leer, Offset, nicht hinter Kommentar) 6 rot; die grüne (`c8`: nicht unmittelbar hinter einer PI ohne Leerraum) ist N-6 (INFO); 105 Probe-Dokumente (7 Formen × 15 Stellen) richtig |
| V-2 | **geschlossen** für die Reihenfolge; Rest als N-4 (LOW, ein Paar ohne Zeile) | 9 von 9 Mutationen der Reihenfolge im Start-Tag rot (`s1`, `s2`, `s3`, `s8`, `s9`, `s12`, `s13`, `s16b`, `s18`), dazu `x1`, `x3` für P35; 9 Formen von Attributen an 20 Rahmen richtig |
| V-3 | **geschlossen** für `LastIndex`; Rest als N-1 (LOW) | `Index`, Zerlegung nach dem zweiten Punkt, Grenze auf drei Glieder, Container ohne Auflösung: 4 von 4 rot; 27 Dokumente mit Namensräumen ab zwei Gliedern richtig |
| V-4 | **geschlossen** | 10 von 10 Mutationen (Keys, Open, Abstract, HasStream, Path in vier Trimm-Formen) rot; 9 Pfade und 5 Ketten richtig |
| V-5 | **teilweise offen (N-2, LOW)** | Leerraum je Referenzart: 14 von 15 Mutationen rot, `w8` (Leerraum im `Collection(` der Property) grün; Rest hinter der Klammer: `)` als Rest grün (`p6`, `p7`) |
| V-6 | **geschlossen** für die Grenzen; Positionen N-5 (INFO) | alle 8 Mutationen der Grenzen (`D7FF`, `D801`, `E000`, `DFFE`, `>`, nur hoch, nur tief, `FFFD` gesperrt) rot |
| V-7 | **geschlossen** | Plan §3, Zeile B5: „CR aus dem Satz genommen: gleichwertig (unten)“ |
| V-8 | **geschlossen** für die Streichung; Aufweichung N-3 (LOW) | eigener Lauf: 37 Fälle, 0 rot; ohne die Zeile `encoding/xml` Fall 37 rot (`ROT  a-check/xml-im-kern: Verstoß nicht gemeldet`, `37 Fälle, 1 rot`) |
| V-9 | **geschlossen** bis auf N-7 (INFO) | in Kommentaren, Fehlertext und Testfallnamen kein `max` mehr; Rest: Feld `strom.max`, Feld `fall.max`, Zeile A5 des Plans |
| V-10 | **geschlossen** | Vertauschung `summe`/`ketten`: `panic: Speicherwächter: Heap 609 MiB über 512 MiB in der Zeile "D8 Rangfolge", eine Schleife des Lesers wächst ohne Ende` |

Neu: **N-1** (LOW), **N-2** (LOW, Rest von V-5), **N-3** (LOW, Rest von V-8), **N-4** (LOW), **N-5** (INFO, Rest von V-6), **N-6** (INFO), **N-7** (INFO, Rest von V-9), **N-8** (INFO).

## 1. V-1 bis V-6 für ihre Klasse

Methode: erst ein Probe-Paket aus der Spezifikation, nicht aus dem Code, gegen die Spitze (alle Proben richtig, also kein Verhalten gegen den Vertrag), dann Mutationen am Code, jede in einer eigenen Kopie gegen die Suite (ohne `A4`, das eine Zerlegung prüft und keine der Mutationen trifft) und gegen das Probe-Paket. „Suite grün, Probe rot“ heißt: der Code der Spitze gehört zur Zusage, aber kein Fänger der Suite fasst die Mutation. Alle Mutationen mit grüner Suite habe ich zusätzlich mit der vollen Suite (mit `A4`) wiederholt: weiterhin grün.

### V-1: CDATA außerhalb der Wurzel

Probe: `CDATA` in sieben Formen (leer, Leerraum, Text, Zeilenumbruch und Tabulator, mit `&#xD800;`, zwei hintereinander, zwei mit Leerraum dazwischen) an 15 Stellen: vor der Deklaration, ohne Deklaration, hinter der Deklaration (mit und ohne Leerraum), hinter Kommentar, hinter PI, zwischen Kommentaren, nach der Wurzel (unmittelbar, mit Leerraum, hinter Kommentar, hinter PI, vor einem Kommentar, nach CRLF). Alle `syntax`. Innerhalb der Wurzel (vor und hinter `DataServices`, im `Schema`) sind alle gültig.

| Mutation | Suite | Probe |
|---|---|---|
| `c1` Prüfung entfällt | rot (`B5`) | rot |
| `c2` nur nach der Wurzel | rot | rot |
| `c3` nur vor der Wurzel | rot | rot |
| `c4` nur das leere `CDATA` | rot | rot |
| `c6` Präfix ohne `<` | rot | rot |
| `c9` nicht unmittelbar hinter einem Kommentar | rot | rot |
| `c8` nicht unmittelbar hinter einer PI ohne Leerraum | **grün** | rot (N-6) |

Die Fixtures der Suite setzen zwischen Deklaration und Wurzel Leerraum; eine `CDATA` unmittelbar hinter `?>` hat keinen Fall. Der Code prüft das Präfix des Tokens und kennt keine Nachbarschaft, die Mutation ist künstlich (N-6, INFO).

### V-2: Rangfolge im Start-Tag

Probe: 20 Rahmen von Elementen (Wurzel mit falscher Version, falschem Namen, V3-Namensraum und ohne Version; zweites `DataServices`; `Schema`, `EntityType`, `Key`, `PropertyRef`, `Property`, `NavigationProperty`, `EntityContainer`, `EntitySet`, `NavigationPropertyBinding`, `Annotations`; Typ mit nicht auflösbarer Basis), je mit einem Inhaltsverstoß und neun Formen von Attributen (doppelt, doppelt über zwei Präfixe derselben URI, doppelte Namensraumdeklaration, Ersatzreferenz im Wert, in `xmlns`, beide in jeder Reihenfolge, mit `U+FFFD` davor): alle `syntax`. In Tiefe 65 mit jeder der neun Formen `syntax`, ohne `tiefe`; ein Edm-Element in Tiefe 65 ohne Attribut `tiefe`; ein zweites Wurzelelement `syntax`; `Foo` mit V3-Namensraum `wurzel`.

| Mutation (Reihenfolge im Start-Tag) | Suite | Probe | roter Fall der Suite |
|---|---|---|---|
| `s1` Tiefe vor dem doppelten Attribut | rot | rot | `D8` P28 |
| `s2` doppeltes Attribut nicht an der Wurzel | rot | rot | `D8` P30, P31 |
| `s3` nur bei Stapel kürzer als 4 | rot | rot | `D8` P28, P32 |
| `s12` Referenz nach `start` | rot | rot | `D8` P29 |
| `s13` Referenz erst nach der Tiefe | rot | rot | `D8` P29 |
| `s16b` doppeltes Attribut erst nach dem Inhalt | rot | rot | `D8` P28, P30, P31, P32 |
| `s18` `version` vor `wurzel` | rot | rot | `D8` P34 |
| `s8` nicht an `Property`/`NavigationProperty` | rot | rot | `D8` P32 |
| `s9` nur Elemente ohne Edm-Namensraum | rot | rot | |
| `s4` nicht an `Schema` | **grün** | rot | |
| `s5` nicht an `EntityType` | **grün** | rot | |
| `s6` nicht an `Key` und `PropertyRef` | **grün** | rot | |
| `s7` nicht an Container, Set, Binding | **grün** | rot | |
| `s10` nicht an `Annotation` und `Annotations` | **grün** | rot | |
| `s14` Referenz bei der Wurzel erst nach dem Inhalt | **grün** | rot | |
| `s15` Referenz bei `Schema`/`EntityType` erst nach dem Inhalt | **grün** | rot | |

Die Reihenfolge ist im Code ein einziger Ort (`start` und `token`); jede natürliche Umstellung (Zeilen 1 bis 7 der Tabelle) ist rot. Die grünen Zeilen sind Mutationen **nach Elementart**, die der Code nicht kennt (`start` prüft vor der Verzweigung nach Name); sie zeigen, dass die Fänger an `Edmx`, `DataServices`, `Property` und in Tiefe 65 stehen, nicht an `Schema`, `EntityType` und den übrigen. Dazu kommt ein Paar ohne Zeile: **Ersatzreferenz gegen Inhalt** (`s14`, `s15`) hat nur über die Tiefe (P29) einen Fänger, `SPEC-125` Nr. 4 d sagt aber „zuerst `syntax` (doppeltes Attribut, Ersatzreferenz), danach Tiefe oder Inhalt“ (N-4). Mehrere `struktur`-Verstöße in einem Element (Nr. 8 f): `D8` P35 prüft „Ort ist einer der Orte“ an sieben Dokumenten; Mutationen des Orts im ersten Zweig (`x1` Property ohne `Type`, `x3` Set ohne `EntityType`) sind rot, die im zweiten Zweig (`x2`, `x4`, `x5`) sind in P35 nie erreichbar und rot in `D6`, wie das Negativ (e) sagt.

### V-3: Namensräume mit Punkten

Probe: neun Namensräume (`A.B`, `A.B.C`, `Microsoft.OData.Models.TripPin`, `a.b.c.d.e`, `Ns.Edm`, `Edm2.X`, `Edmonton.Models`, `X.Edmx`, `Ns1.N_2`) mit und ohne Alias, Referenz über Namensraum und über Alias, in allen sechs Positionen (Basistyp, Typ einer Property, `Collection(…)` einer Property, Typ einer Navigation-Property, `Collection(…)` einer Navigation-Property, Typ des Sets, Ziel mit Container) und im Modell geprüft: alle richtig. Zwei Schemas `A` und `A.B` mit `A.B.C` und `A.B`: richtig (`A.B` + `C`, `A` + `B`).

Mutationen: `Index` statt `LastIndex` (`q1`), Zerlegung nach dem zweiten Punkt (`q2`), Grenze auf höchstens drei Punkte (`q3`), Container ohne Auflösung des Alias (`q6`): 4 von 4 rot (`F4`, `D2`). Gleichwertig: `q7` (`Index` statt `LastIndex` beim `/` des Ziels: ein zweiter Schrägstrich ist ohnehin ein unbekannter Name), `q4` (kein Alias hat einen Punkt). **Grün: `q5`**, `strings.HasPrefix(s, "Edm")` statt `"Edm."` in `einfacherTyp`, und **`p10`** (`Contains` statt `HasPrefix`): ein Verweis in `Edmonton.Models`, `Edm2.X` oder `Ns.Edm` wird `nicht_aufloesbar`. Nr. 8 b nennt `Edmonton`, `Edm2` und `Ns.Edm` ausdrücklich als zulässige Namensräume; `D6` prüft nur, dass das `Schema` angenommen wird, nie einen Verweis hinein (N-1).

### V-4: Keys, Path, HasStream, Open

Probe: Kette A (Key), B (ohne), C (leerer `<Key/>`), D (leerer `<Key/>`), E (ohne): alle `[Id]`; abstrakte Basis ohne Schlüssel, Schlüssel in der Mitte mit zwei `PropertyRef` in der Reihenfolge `Y, X`, leere `<Key/>` darunter: `[Y X]` bis unten; abstrakte Basis mit leerem `<Key/>` und Schlüssel in der Ableitung; `OpenType` in der Basis mit und ohne Wiederholung, `OpenType="false"` in der Ableitung (bleibt offen), `Abstract` nur am Typ selbst; `HasStream` (wahr, falsch, an Basis und Ableitung): nie offen, Fund `konstrukt` nur an den Typen mit `true`; `Path` in neun Formen (Leerzeichen, Tabulator, `U+00A0`, `U+2003`, `&amp;`): unverändert.

Mutationen: Keys unbedingt überschrieben (`k1`), nur eigene oder direkte Basis (`k2`), leerer Key an der Ableitung löscht (`k10`), `Open` letzter Wert (`k3`), nur direkte Basis (`k4`), `Abstract` geerbt (`k5`), `HasStream` setzt `Open` (`k6`), `Path` mit `TrimSpace` (`k7`), mit `Trim(" ")` (`k8`), mit `TrimRight(" \t")` (`k9`): **10 von 10 rot** (`F2`, `F6`).

### V-5: Leerraum und Klammerformen

Probe: sieben Leerraumarten (Leerzeichen, Tabulator, Zeilenumbruch, `U+00A0`, `U+2003`, `U+0085`, CR) vorn, hinten und im Namen in sechs Referenzarten, im Ziel eines Bindings (sechs Stellen), im Namen einer `PropertyRef`, in `Edm.`-Typen: alle `nicht_aufloesbar`. über 30 Klammerformen, je für Property und Navigation-Property: nur die geschlossene Form löst auf.

| Mutation | Suite | Probe |
|---|---|---|
| `w1` bis `w7`, `w9` bis `w15`: `TrimSpace` an `entityRef`, Basistyp, Typ der Property, Typ der Navigation-Property, Set, Ziel, `PropertyRef`, Inneres des `Collection(` der Navigation-Property, `einfacherTyp`, Container und Name des Ziels, Name und Namensraum in `qualify` | rot (14 von 14) | rot |
| `p1`, `p2` Suffix `)` nicht geprüft (Property, Navigation-Property) | rot | rot |
| `p9` verschachteltes `Collection(Collection(…))` in der Navigation-Property | rot | rot |
| **`w8`** `TrimSpace` im Inneren des `Collection(` der **Property** | **grün** | rot |
| **`p6`**, **`p7`** `TrimRight(…, ")")`: `Collection(Edm.String))` und `Collection(Ns.T))` | **grün** | rot |
| **`p4`**, **`p5`** Präfix ohne Beachtung der Schreibweise: `collection(Edm.String)` | **grün** | rot (`p4`) / nicht erfasst (`p5`) |
| **`p12`** `Edm.string`, `Edm.INT32` ohne Beachtung der Schreibweise | **grün** | nicht erfasst |

Der Fix der ersten Runde führte `Collection(Edm.StringX` und die Formen `)X` und `) `; die Form, in der der Rest wieder eine schließende Klammer ist, fehlt, ebenso das Innere der Property-Klammer (die der Navigation-Property hat den Fall: `w9` rot) und die Schreibweise von `Collection(` und nach `Edm.` (N-2).

### V-6: Ersatzzeichen

Probe: 21 Verweise (`D7FF`, `E000`, `FFFD`, `10FFFF`, `10000`, dezimal und hexadezimal in beiden Schreibweisen der Ziffern, mit führenden Nullen; ungültig `D800`, `DFFF`, `DBFF`, `DC00`, `DAAA` in allen Schreibweisen) in 15 Positionen (Attribut von `Annotation` als erstes und zweites, einfache Anführungszeichen, Attribut einer fremden Element, `xmlns`, `Schema`, `EntityType`, Wurzel, `DataServices`; Text in `Schema`, Wurzel, `DataServices`, `Annotation`, hinter einem Element, fremdes Element) in zehn Nachbarschaften (allein, `U+FFFD` davor und danach, Buchstaben, `&#65;`, `&amp;`, `&lt;`, `&#xFFFD;`): gültige Verweise gültig, ungültige `syntax`. Zwei Verweise im selben Token in beiden Reihenfolgen, `U+FFFD` vor einem ungültigen Verweis, `&amp;#xD800;` (kein Verweis): richtig. Kommentar, PI und `CDATA` mit den 21 Verweisen **und** `U+FFFD` im selben Token: gültig.

| Mutation | Suite | Probe |
|---|---|---|
| `e1` bis `e8`: Untergrenze `D7FF`, `D801`; Obergrenze `E000`, `DFFE`; `>`/`<`; nur hoher, nur tiefer Teil; `FFFD` gesperrt | rot (8 von 8) | rot |
| `e10`, `e11`, `e19`, `e22`, `e23`: Basis, `x`-Präfix, CDATA nicht ausgenommen, `#` allein, Länge des Verweises | rot | rot |
| `e13` nur das letzte Attribut | rot | grün (Probe schwächer) |
| `e16` Text nicht geprüft | rot | rot |
| **`e14`** Werte der `xmlns`-Attribute nicht geprüft | **grün** | rot |
| **`e15`** nur Attribute ohne Namensraum | **grün** | rot |
| **`e17`** Text nur ab Stapel 2 | **grün** | rot |
| **`e18`** Text in fremden Elementen nicht | **grün** | rot |
| **`e24`** Kommentar auf Ersatzverweise geprüft | **grün** | rot |
| **`e25`** PI auf Ersatzverweise geprüft | **grün** | rot |

Die Grenzen (V-6 im Wortlaut) sind geschlossen. Die Fänger der Verweise führen die Position `Annotation` und den Text eines `Schema`; die Position der Mutationen `e14` bis `e18` und die **gültige** Seite im Kommentar und in der PI (`e24`, `e25`: ein Verweis in einem Kommentar ist keiner, auch neben `U+FFFD`) haben keinen Fall (N-5, INFO: eine Stelle im Code, Mutationen nach Position sind künstlich). `e9` (32 Bit gegen 16) und `e20` (`Contains` gegen `HasPrefix` bei `CDATA`) sind gleichwertig.

## 2. V-8, V-9, V-10

**V-8.** `make gegenproben` mit einem eigenen Stempel (`GEGENPROBEN_STEMPEL` auf eine Datei im Scratchpad, das Repo blieb unberührt): `gegenproben: 37 Fälle, 0 rot`, der berechnete Stempel gleich dem committeten (`23a07ecf…`). In einer Wegwerfkopie (`git archive`, eigenes `git init`, weil das Ziel `git` braucht) ohne die Zeile `{pattern: "encoding/xml", …}` in `.a-check.yml`: `ROT  a-check/xml-im-kern: Verstoß nicht gemeldet`, `gegenproben: 37 Fälle, 1 rot`. Der Commit `58d7de3` ändert nur `tools/gegenproben/faelle.go` (5 Zeilen) und den Stempel (je 1), keine Datei unter `harness/` außer dem Stempel und nichts in `.a-check.yml`, wie der Auftrag verlangt. **Klasse:** Fall 37 fängt die Streichung der Zeile, nicht ihre Aufweichung. In einer Kopie mit der Datei `src/adapters/driven/sqlite/xmlleak.go` (`import _ "encoding/xml"`) meldet `make a-check` mit der Regel der Spitze `tech-leak: Tech encoding/xml außerhalb src/adapters/driven/odata`; mit `adapter: "src/adapters"` in der Regel meldet es **0 Befunde**, und kein Fall der Gegenprobe wird rot (Fall 37 prüft den Kern, der Fall `a-check/bibliothek-in-fremdem-adapter` die Zeile für `modernc.org/sqlite`). Eine Lockerung des Gates (`AGENTS.md` §3.6) bleibt damit unbemerkt (N-3).

**V-9.** In `src/adapters/driven/odata`, `src/hexagon/model`, `src/hexagon/ports/outbound` steht `max` nur noch als Feld oder Aufruf: `strom.max` (`strom.go:37`, `:104`, `:110`), `fall.max` der Fixture-Struktur (`metadata_codec_vertrag.go`, `…_lesen.go`, `…_funde.go`), der eingebaute `max(` (`metadata_codec_vertrag.go:546`) und fremde Pakete (`redaktor*.go`). Kommentare, Fehlertext (`grenze kleiner 1`) und Testfallnamen sagen `grenze`. Das Feld `fall.max` ist als Entscheidung vertretbar (ein Feld eines Tests, in keiner Meldung), `strom.max` ist ein Feld im Code des Lesers, das der Auftrag nicht nannte und das der Spezifikation widerspricht, wo sie `grenze` sagt: kein Kommentar, kein Fehlertext, INFO. Nicht nachgezogen ist **die Zeile A5 der Tabelle im Plan** (`slice-odata-codec-csdl.md:197`): „`LimitReader(max)`; `max+2`“, wo alle anderen Zeilen `grenze` führen (N-7).

**V-10.** Vertauschung von `summe` und `ketten` in `loese` (Kopie, `--memory=2g`, `timeout 150`): `panic: Speicherwächter: Heap 609 MiB über 512 MiB in der Zeile "D8 Rangfolge", eine Schleife des Lesers wächst ohne Ende`. Die Meldung nennt die Zeile (keinen Unterfall, und das sagt sie auch nicht: „in der Zeile“), der Satz über die Stapel ist gestrichen. Geschlossen.

## 3. Plan und Spezifikation gegen den Code (`AGENTS.md` §3.9)

* **Spezifikation gegen Code.** Nr. 4 (CDATA), 4 b (beide Seiten der Grenzen, auch neben `U+FFFD`), 4 d (Reihenfolge `fertig`, doppeltes Attribut, Tiefe, Wurzel oder Element; die Ersatzreferenz geht `start` in `token` voraus), 6 ((a) vor (b)), 8 f (nicht zugesagt, `D8` P35), 9 (letzter Punkt, Leerraum zählt), 10 (Klammer als letztes Zeichen, Rest im Namen), 11 (`Keys` der Ableitung, `Open` nur durch `OpenType`, `Path` unverändert): alle am Code durch Proben bestätigt; keine Abweichung.
* **Plan gegen Code.** Die Zeilen V-1 bis V-6, V-8, V-10 der Tabelle von dfc5822 nennen Mutationen und rote Tests, die ich nachgefahren habe (V-2: `s1`/`s3`/`s12`/`s13`/`s16b`/`s18` rot in den genannten Fällen; V-5: `TrimSpace` je Referenzart rot; V-6: `D7FF`, `E000` rot). Die Raten stimmen (Abschnitt 4). Der Plan beschönigt nichts: Abdeckung „nicht neu gemessen“, die Zuordnung von `s3` und `s4` ist offen benannt, der Fehler `gocyclo` des Implementers steht im Text, die Größe 4333 wird ohne Kürzen berichtet. Zwei Dinge sagt er nicht: die Zeile V-5 führt `TrimSpace` je Referenzart, nicht das Innere von `Collection(` der Property und nicht den Rest `)` (N-2); die Zeile A5 trägt noch `max` (N-7).
* **Zeile B5 (V-7).** „CR aus dem Satz genommen: gleichwertig (unten)“, der Auftrag des Architect „ohne die Mutation CR“ ist erfüllt.
* **Ein Commit (N-8).** `AGENTS.md` §3.9 verlangt Code und Plan **im selben Commit**. Die Runde liegt in drei Commits: Code und Fänger `4c7019d`, Gegenprobe und Stempel `58d7de3`, Plan `dfc5822` (alle vom selben Handoff, alle auf demselben Stand der Übergabe). Die früheren Runden hielten dasselbe Muster (`7e08438`/`cf3131d`, `867e878`/`ad8732b`); der Zustand der Spitze ist zusammenhängend, die Regel dem Wortlaut nach nicht erfüllt.
* **Abdeckung.** `go test -race -count=1 -coverprofile -coverpkg=csdl`, Image der Stufe `race`, 274 s: **99,7 %**. Nicht ausgeführt: `tokenFehler` zu `io.ErrNoProgress` (`lesen.go:70`) und `stromFehler.Error` (`strom.go:27`), beide im Plan und in `SPEC-125` *Ungeprüft* benannt. Die neuen Zweige liegen in Teilausdrücken (`HasPrefix(…, "<![CDATA[")` als dritter Operand von `||`) und in einer Umstellung; die Anweisungsabdeckung unterscheidet sie nicht, die Mutationen `c1` bis `c9` und `s1` bis `s3` tun es (rot).

## 4. Raten (Messender: der Verifier; Host mit 20 Kernen; Testbinary der Kopie, N Prozesse zu `-test.count=1`; `--cpus=3`, `--memory=2g`)

Unmutiert: `B5`, `B4`, `D5`, `D2`, `F2`, `F4`, `F6`, `D8` zusammen **0 von 20** rot. Alle Fänger der Runde sind deterministisch (kein Nebenläufiger, keine Map-Reihenfolge im Fänger).

| Zeile | Mutation | Prozesse rot |
|---|---|---|
| B5 | `CDATA`-Prüfung entfällt (`c1`) | 20 von 20 |
| B4 | Untergrenze `D7FF` (`e1`) | 20 von 20 |
| B4 | Obergrenze `E000` (`e3`) | 20 von 20 |
| B4 | `>`/`<` statt `>=`/`<=` (`e5`) | 20 von 20 |
| D5 | Suffix `)` der Property nicht geprüft (`p1`) | 20 von 20 |
| D5 | Suffix `)` der Navigation-Property nicht geprüft (`p2`) | 20 von 20 |
| D2 | `TrimSpace` je Art: Basistyp, Typ der Property, Set, Ziel, `PropertyRef` (`w2`, `w3`, `w5`, `w6`, `w7`) | je 20 von 20 |
| F4 | `Index` statt `LastIndex` (`q1`) | 20 von 20 |
| F4 | höchstens drei Punkte (`q3`) | 20 von 20 |
| F2 | leerer `Key` löscht die Keys (`k10`) | 20 von 20 |
| F2 | `HasStream` setzt `Open` (`k6`) | 20 von 20 |
| F6 | `Path` mit `TrimSpace` (`k7`) | 20 von 20 |
| D8 P28 | Tiefe vor dem doppelten Attribut (`s1`) | 20 von 20 |
| D8 P29 | Referenz nach `start` (`s12`), Referenz nach der Tiefe (`s13`) | je 20 von 20 |
| D8 P30, P31 | doppeltes Attribut nicht an der Wurzel (`s2`) | 20 von 20 |
| D8 P28, P32 | nur bei Stapel kürzer als 4 (`s3`) | 20 von 20 |
| D8 P32 | doppeltes Attribut erst nach dem Inhalt (`s16b`) | 20 von 20 |
| D8 P34 | `version` vor `wurzel` (`s18`) | 20 von 20 |
| D8 P35 | Ort des Typs statt des Members (`x1`), Ort des Containers statt des Sets (`x3`) | je 20 von 20 |
| D8 P35 | Ort im zweiten Zweig (`x2`, `x4`, `x5`) | **0 von 20** in `D8` (rot in `D6`, siehe oben; das Negativ (e) gilt) |
| I1 | `EntityTypes` über die Map `index` (`i1`) | Zeile I1 20 von 20; `F1`, `F3`, `F4`, `F6`, `B5`, `G4` zusammen 20 von 20 |
| I2 | Zähler als Variable des Pakets (`i2`), Detektor | I2 und I3 zusammen 15 von 15 (unmutiert 0 von 15); ohne Detektor grün (Suite und `A4` zusammen) |
| I3 | Funde als Variable des Pakets unter Sperre mit `Gosched` zwischen den Abschnitten (`i3`) | ohne Detektor 30 von 30 (in jedem die Zeile I2) |

Die Reihen sind klein und sagen „nie grün gesehen“, nicht „immer rot“; die Zeile I1 hat den Rest aus der ersten Verifikation (V-11: `Keys`, `NavigationBindings`, `Namespaces` über eine Map fängt die Zeile nie) unverändert.

## 5. Größe und Gegenprobe getrennt (`AGENTS.md` §3.14)

`git diff --numstat 3b6b3b8..4c7019d -- src/hexagon/model src/hexagon/ports/outbound src/adapters/driven/odata/csdl .a-check.yml`: **4332 Einfügungen, 1 Löschung, 4333 Zeilen**; derselbe Wert bis `58d7de3` und bis `HEAD` (`4c7019d` ist der letzte Commit mit einem Pfad der Marke). Die Marke steht bei 2400, also 1933 darüber (+81 %). Gegenüber der ersten Verifikation (4199 bei `7e08438`) sind es +134 netto (169 Einfügungen, 35 Löschungen im Bereich `7e08438..4c7019d`); der Plan nennt 4333 und die Pfade und beschönigt nicht. **`tools/gegenproben`: 5 Zeilen** (`faelle.go`) plus der Stempel (2 Zeilen), nicht in den Pfaden der Marke; der Fall steht auch in `harness/vertraege/SPEC-102-gegenproben.md` (Architect, 1 Zeile). Die Rückführung (`in-progress` nach `next`) ist die Entscheidung des Planners.

## 6. Findings

| ID | Schwere | Ort | Aussage | Ausfallszenario | Quelle | Adressat |
|---|---|---|---|---|---|---|
| N-1 | LOW | `aufloesen.go` (`einfacherTyp`: `strings.HasPrefix(s, "Edm.")`); Zeile D6, D5 | Kein Fall löst einen Verweis in einen Namensraum auf, dessen Name wie `Edm` beginnt. `SPEC-125` Nr. 8 b nennt `Edmonton`, `Edm2` und `Ns.Edm` als zulässige Namensräume; `D6` prüft nur die Annahme des `Schema`. Mutation `q5` (`HasPrefix(s, "Edm")`) und `p10` (`Contains(s, "Edm.")`) bleiben grün (mit voller Suite wiederholt); Proben `Edmonton.Models.T`, `Edm2.X.T`, `Ns.Edm.T` sind bei `q5` bzw. `p10` `nicht_aufloesbar` | Eine Änderung der Prüfung auf `Edm.` lehnt jedes Dokument mit einem Namensraum wie `Edmonton` ab (Fixture gleicht dem Wert der Mutation: das Dokument mit `Edmonton` hat keinen Verweis hinein) | `AGENTS.md` §3.15 | Implementer (D5 oder D6: Verweis auf einen Typ in `Edmonton`, `Edm2`, `Ns.Edm`, `EdmX.Foo` in den sechs Positionen; `q5` und `p10` rot sehen) |
| N-2 | LOW | `aufloesen.go` (`propTyp`, `navTyp`, `einfacherTyp`); Zeile D5, D2 | Rest von V-5. (a) `Collection(Edm.String))` und `Collection(Ns.T))` (Rest ist wieder `)`) sind ungeprüft: `TrimRight(…, ")")` (`p6`, `p7`) bleibt grün. (b) Leerraum im Inneren von `Collection(` der **Property** (`Collection( Edm.String)`, `Collection(Edm.String )`) hat keinen Fall, der der Navigation-Property schon (`w9` rot, `w8` grün). (c) Schreibweise: `collection(Edm.String)` (`p4`, `p5`) und `Edm.string` (`p12`) sind ungeprüft; Nr. 10 sagt „Schreibweise und Leerraum zählen“ und führt `edm.string`, nicht die Schreibweise nach dem Präfix | Ein Umbau der Klammer-Prüfung oder des Vergleichs der Namen nimmt Dokumente an, die der Vertrag ablehnt; der Vertrag ist für jeden Adapter gleich | `AGENTS.md` §3.15 | Architect (ein Satz in Nr. 10: `Collection(…))`, `collection(…)`, `Edm.string` lösen nicht auf); Implementer (D5: die Formen für Property und Navigation-Property; D2 oder D5: Leerraum im Inneren der Klammer der Property; `p4` bis `p7`, `p12`, `w8` rot sehen) |
| N-3 | LOW | `.a-check.yml` Zeile `{pattern: "encoding/xml", adapter: "src/adapters/driven/odata"}`; `tools/gegenproben/faelle.go` Fall 37 | Fall 37 fängt die Streichung der Zeile, nicht die Aufweichung ihres Feldes `adapter`. Eine Kopie mit `adapter: "src/adapters"` meldet für `import _ "encoding/xml"` im SQLite-Adapter 0 Befunde (mit der Regel der Spitze `tech-leak`), und `make gegenproben` bleibt grün | Ein späterer Commit öffnet die Regel auf alle Adapter, und `encoding/xml` kann in den SQLite-Adapter wandern, ohne dass ein Gate es meldet; eine Lockerung nach `AGENTS.md` §3.6 ohne roten Fänger | `AGENTS.md` §3.6, §3.15 | Implementer (ein weiterer Fall: Datei im SQLite-Adapter mit `encoding/xml`, `WantText: tech-leak`; der Architect trägt ihn in `SPEC-102` ein) oder Planner (benanntes Negativ mit Folge) |
| N-4 | LOW | `SPEC-125` Nr. 4 d; Plan §6 Zusatztabelle; `D8` P28 bis P35 | Das Paar **Ersatzreferenz gegen Inhalt** (`wurzel`, `version`, `struktur`) hat keine Zeile und keinen Fänger: P29 deckt Ersatzreferenz gegen Tiefe. `s14` (Wurzel) und `s15` (`Schema`, `EntityType`) bleiben grün, ebenso das doppelte Attribut nach Elementart (`s4` bis `s7`, `s10`: `Schema`, `EntityType`, `Key`, `PropertyRef`, Container, Set, Binding, `Annotation(s)`; Fänger stehen an `Edmx`, `DataServices`, `Property`, Tiefe 65). Der Code hat je einen Ort (`token`, `start`), die Mutationen nach Elementart sind künstlich; `AGENTS.md` §3.16 verlangt aber „je Paar von Verstößen am selben Schlüssel“ eine Zeile | Wird die Prüfung auf Ersatzzeichen in `kind` oder `wurzel` verteilt, ändert sich die Ursache für Dokumente mit zwei Verstößen, und nur P29 sieht es | `AGENTS.md` §3.16 | Architect/Implementer (P-Zeilen: Ersatzreferenz mit `Version="3.0"`, falschem Wurzelnamen, ungültigem `Namespace`; doppeltes Attribut an `Schema`, `EntityType`, `Key`, `EntitySet`; oder ein benanntes Negativ mit Grund und Folge) |
| N-5 | INFO | `lesen.go` (`refs`); Zeilen B4, B5 | Positionen der Ersatzverweise ohne Fänger: Wert von `xmlns`-Attributen (`e14`), Attribute mit Namensraum (`e15`), Text in der Wurzel und in fremden Elementen (`e17`, `e18`), die **gültige** Seite im Kommentar und in der PI (`e24`, `e25`: ein Verweis dort ist keiner). Eine Stelle im Code, Mutationen nach Position sind künstlich; Proben bestätigen den Code | keines für den Code | `AGENTS.md` §3.15 | Implementer nach Ermessen des Planners (Kommentar und PI mit `&#xD800;` neben `U+FFFD` als gültiger Fall) |
| N-6 | INFO | Zeile B5 | `CDATA` unmittelbar hinter `?>` ohne Leerraum hat keinen Fall (`c8` grün); der Code kennt keine Nachbarschaft | keines | `AGENTS.md` §3.15 | Implementer nach Ermessen |
| N-7 | INFO | `strom.go:37` (Feld `max`); `fall.max` der Fixture-Struktur; Plan Zeile A5 (`:197`: `LimitReader(max)`, `max+2`) | Rest von V-9. Das Feld `fall.max` hat der Implementer bewusst stehen lassen (Feld eines Tests, in keiner Meldung): vertretbar. Das Feld `strom.max` und die Zeile A5 des Plans nennen den abgelösten Namen | gering | `AGENTS.md` §3.7, §3.9 | Implementer (`strom.max`), Planner (Zeile A5) |
| N-8 | INFO | Commits `4c7019d`, `58d7de3`, `dfc5822` | `AGENTS.md` §3.9 verlangt Plan und Code im selben Commit; die Runde liegt in drei Commits, wie die vorigen Runden. Die Spitze ist zusammenhängend | keines | `AGENTS.md` §3.9 | Planner (Regel oder Praxis klären: ein Handoff als ein Commit oder die Regel um „im selben Handoff“ schärfen) |

## 7. Offen für die Closure (Planner)

1. **Größe**: 4333 gegen 2400 in §4 eintragen (+81 %, Tests und Fixtures überwiegen, die Formel kennt den Produktcode eines Parsers nicht); die Rückführung ist die Entscheidung des Planners. `tools/gegenproben` 5 Zeilen (+ Stempel 2) getrennt.
2. **Register** (`docs/plan/planning/observations/`): `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (N-1, N-2, N-3: drei weitere Belege nach V-3 bis V-6; die Fixtures der Runde decken die genannte Form, nicht die Menge: `Edm`-Anfang, `)` als Rest, Aufweichung statt Streichung); `BEO-REPO/randform-erst-im-review-entschieden` (kein neuer Beleg: V-1 und V-2 sind im Vertrag und im Code geschlossen); `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (kein neuer Beleg: 99,7 %, Rest benannt); `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (kein Beleg: alle Raten dieser Runde deterministisch); `BEO-REPO/akzeptiertes-negativ-ohne-folge` (V-13 ist behoben: die Folge steht im Plan, Abschnitt *Akzeptierte Negative* (b)).
3. **Risiko-Ausgänge** (§6): wie in der ersten Verifikation (Bibliothek *entfallen*, `SetMetadata` *eingetreten*, Sendung `slice-protokolltreue` *eingetreten*, Stabilität der Reihenfolge *entfallen*, Größe *eingetreten*, Bericht des Implementers *entfallen*); zusätzlich: *Eingabe aus einem fremden Dienst* und *Randformen* **entfallen** für den Code (V-1 bis V-3, die Proben bestätigen), mit dem Rest N-1, N-2 als Fänger-Lücke; *Prüfgrenzen* entfallen (g1 bis g7 der ersten Verifikation, unverändert).
4. **Lerneintrag** (§7): (a) *Fänger führt nur eine Form der Eingabe* hält bei V-5, bei der Klasse der Verweise in `Edm`-ähnliche Namensräume (N-1) und bei der Aufweichung einer Regel statt ihrer Streichung (N-3): die Menge zu einer Zusage („jede Referenzart“, „jeder Rest hinter der Klammer“, „jede Position“) wird in der Runde eine Liste der Formen, die dem Fix einfielen; der Fix hat die genannte Mutation und die ihr ähnlichen. (b) *Randform im Code entschieden* ist mit V-1 und V-2 für diesen Slice geschlossen: Spezifikation (Nr. 4 d, 8 f) und Fänger (P28 bis P35) tragen sie. (c) *Plan folgt nicht der Korrektur*: V-7 geschlossen, N-7 (Zeile A5) und N-8 (drei Commits) sind der Rest. (d) Ein Probe-Paket aus Generatoren über die Klassen der Spezifikation, einmal gegen den Code und dann gegen die Mutationen gefahren, fand in einer Stunde die Lücken, die zehn Mutationen von Hand nicht fanden; die Suite selbst führt keinen solchen Generator, und die Grenze (Fänger in der Suite ohne Generator) ist eine Entscheidung des Planners.
5. **Move nach `done/`** von Hand nach der Entscheidung zu N-1 bis N-4 (Nachbesserung oder benanntes Negativ), in zwei Commits nach `AGENTS.md` §3.3; `.claude/agents/architect.md` bleibt beim Nutzer.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Code gegen Vertrag | rund 4000 Probe-Dokumente nach `SPEC-125` Nr. 4 b, 4 d, 6, 9, 10, 11: kein Verhalten gegen den Vertrag |
| Sicherheit des Lesers | unverändert gegenüber der ersten Verifikation: die Änderung an `token` und `start` berührt `DOCTYPE`, Entitäten, Zeichensatz und Grenzen nicht (`git diff` von `4c7019d`: 9 Zeilen in `lesen.go`, Kommentare in `leser.go`, `strom.go`) |
| Suite mit und ohne Detektor | unmutiert grün: ohne Detektor `ok csdl 29 s` (mit `A4`), mit Detektor `ok csdl 274 s`, `model` und `outbound` ok |
| Kein `//nolint` | keine neue Direktive in den Dateien der Runde |
| Hard Rules `AGENTS.md` §3.1, §3.2, §3.4, §3.13 | alles über `make` und `docker run`, keine Direktive, `spec/architecture.md` unberührt, Sendungen getragen |
| Commits | jede Kopfzeile nennt `LH-FA-06` und `LH-RB-01` |

## Verdikt

**Bestanden.** V-1 bis V-4 sind für ihre Klasse geschlossen, V-7 und V-10 geschlossen, V-8 und V-9 für die genannte Form geschlossen; V-5 ist zum Teil offen. Rückgabe an den **Planner** mit den LOW N-1 bis N-4 zur Entscheidung (Nachbesserung durch Implementer und Architect oder benanntes Negativ), den INFO N-5 bis N-8 und der Größe. Eine weitere Nachverifikation ist nur nötig, wenn der Planner N-1 bis N-4 nachbessern lässt. `make gates` läuft nach dem Commit dieses Berichts.
