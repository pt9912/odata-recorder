# Review-Report: slice-odata-url-filter — 2026-10-31

**Review-Art:** Code — geprüft gegen Plan (`slice-odata-url-filter` §1, §3 mit der Tabelle *Zusage · Fänger* Z1 bis Z35, dem Absatz *Zusage · Mutation · roter Test* und den Raten, §4 mit Block, Marke 4700 und Messung 1911, §6 mit Raster O1 bis O9 und der offenen Rückgabe an den Architect), gegen `SPEC-164` und `SPEC-165` samt Nachbarn (`SPEC-160` Nr. 8, `SPEC-161` Nr. 6, `SPEC-162`, `SPEC-125` Nr. 8 b, `SPEC-122`, `SPEC-128`) und gegen die Hard Rules (`AGENTS.md` §3.7 bis §3.18). Nicht geprüft wird gegen die DoD; das ist die Frage des Verifiers.

**Gegenstand:** `slice-odata-url-filter` (Welle `welle-mock`), Code-Commit `84b24a5` (Implementer) und Merge `e2e9cca` (HEAD); Basis `4e36c01`. Pfade: `src/hexagon/model/filter_ast.go`, `src/hexagon/model/filter_ast_test.go`, `src/adapters/driven/odata/urlparser/filter.go`, `filter_token.go` und die sieben `filter_*_test.go`; daneben der Plan.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `e2e9cca` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Methode.** Der Bericht des Implementers (Plan §3) wurde gelesen; keine Mutation, keine Rate und keine Einschätzung daraus ist übernommen. Alle Mutationen laufen auf Kopien im Scratchpad (`/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/review-urlfilter/`; im Worktree liegt nichts außer diesem Report). Ein Lauf ist `go test -count=1 ./src/adapters/driven/odata/urlparser ./src/hexagon/model` im Image `odata-recorder:deps` (`--network none --memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`, Exit-Code aus Datei); der Basislauf ohne Mutation ist grün. Zusätzlich ein Probe-Paket (nicht im Repo): rund 200 Einzeleingaben der Randformen und ein Generator mit voller Klammerung, 3000 Saaten ab 1, 25 Blättern, eigenem Orakel. Die Raten von Z29/Z30 sind von mir gemessen: ein einmal übersetztes Testprogramm, 20 Prozesse nacheinander mit `-test.run '^TestFilterGleichzeitig$' -test.count=1`, im Image `odata-recorder:deps` (ohne Detektor) und `odata-recorder:race` (cgo, `-race`); `make test-race` selbst habe ich nicht gefahren.

**Eingangs-Kontext:**

- Slice-Plan `slice-odata-url-filter` in `e2e9cca`
- `SPEC-164`, `SPEC-165`, `SPEC-160` Nr. 8, `SPEC-161` Nr. 6, `SPEC-162` Nr. 2 und 8, `SPEC-122` Nr. 8 und 9, `SPEC-128` Nr. 1 und 2
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md)
- [LH-FA-08, LH-QA-05](../../spec/lastenheft.md)
- `AGENTS.md` (Hard Rules); Reviewer-Skill mit den Zeilen *Fänger führt nur eine Form der Eingabe*, *Paar von Ursachen ohne Fänger*, *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*; Report `2026-10-31-slice-replay-service-review` als Maßstab

---

## Antworten auf die Fragen des Auftrags

**Bilanz.** 49 gültige Mutanten (ein weiterer übersetzt nicht und zählt nicht), **40 rot, 9 grün**: `R14d` (F-1), `R01c` und `R23b` (F-2), `R20b` und `R20c` (F-3), `R22f` (F-4), `R33d` (F-5), `R10r` (F-7), `R19a` (F-9). Dazu die Raten von `M29`/`M30` (zwei Mutanten, drei Bedingungen) und die unveränderte Fassung. Kein HIGH.

**(1) `nurNamenzeichen` / `einfacherName`.** Beide Prüfungen sichern denselben Fall; ich habe `nurNamenzeichen` ganz durch `return true` ersetzt (`R10r`): alle Tests grün. Die Prüfung ist über `einfacherName` (Name, Glieder von Pfad und qualifiziertem Namen) vollständig ersetzbar und ändert in keinem Fall Ursache oder Baum: äquivalent. → F-7 (LOW). Eine Mutation an nur einer Stelle bleibt grün, wie der Plan sagt; Mutation an beiden (`R11c`, Zeichen `~`) ist rot (`TestZeichenAussenGrammatik`).

**(2) Zeile 28.** M28 (letzten Baum zwischenspeichern) überlebt zu Recht: der Baum besteht nur aus Werten, ein zwischengespeicherter Baum ist beobachtbar nicht von einem frischen zu unterscheiden; ein unsynchronisierter Cache bricht erst nebenläufig (`R28b`: rot, aber in `TestFilterGleichzeitig`, nicht in `TestFilterWiederholung`). M28b (Zustand der Instanz) ist rot im Zeilen-Fänger. Gefangen ist damit die Hälfte „Zustand bleibt zwischen Aufrufen“, soweit er an der Instanz hängt; Zustand in einer Variablen des Pakets (`R28a`, `klammern` als Paketvariable) fängt `TestFilterWiederholung` **nicht** (a0 und Wiederholung teilen den schmutzigen Zustand), rot werden nur `TestStruktur` und `TestZeichenAussenGrammatik` durch Streuung zwischen den Tests (F-13). Die Hälfte „der Baum teilt nichts“ („Ändern eines Wertes am ersten Ergebnis“, Plan Z28) führt kein Test aus, und die Zusage „keine Slice, keine Map, kein Zeiger“ von `SPEC-165` Nr. 2 hat keinen Fänger (`R33d` grün): F-5 (MEDIUM).

**(3) Guid mit Hex-Buchstaben am Anfang.** Der Code setzt den Wortlaut von `SPEC-164` Nr. 4 b/4 c genau um: `a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11` ist ein Wortlauf mit `-`, also `wert`; `0f8fad5b-…` und die Großschreibung mit Ziffer am Anfang werden gelesen und kleingeschrieben (Probe bestätigt, ebenso `{…}` `wert`). `flAusnahmen15` hält drei Formen (`a`, `A`, `f` am Anfang) an einer Stelle. §6 führt die Rückgabe mit Ausgang *offen* und Adresse Architect, mit Vorschlag und Folge; das genügt `AGENTS.md` §3.13. Kein Finding am Code (F-12 INFO nennt die Folge).

**(4) Kommentare (§3.7, §3.11).** Keine verworfene Alternative, kein abwesender Text. Zwei Zusagen im Kommentar ohne Test: `filter.go` „die Länge geht vor der Form“ (F-1; Mutation grün) und `filter_token.go` „geht nie zurück“ (F-10; Z32 ist das akzeptierte Negativ). Die übrigen Zusagen (Zustandslosigkeit, kein Teilergebnis, Tiefe 32) tragen Tests.

**(5) `New()` bleibt ohne Leser.** `grep` über `src` ohne Tests: `NewFilterLeser` kommt nur in `filter.go` vor, `src/bootstrap` kennt weder `NewFilterLeser` noch `WithFilter` (die Änderung an `src/bootstrap/main/main.go` im Merge stammt von `main`, nicht vom Slice); Gate-Kette, `Makefile`, `harness/` unberührt. Die Verdrahtung gehört `slice-bootstrap-m3` (Sendung in §6, Ausgang *offen*, Adresse Planner). Ohne Befund.

**Größenmarke.** `git diff --numstat 65f75c8..84b24a5` über die drei Pfade des Blocks: 1911 Einfügungen, 0 Löschungen, Summe 1911 gegen Marke 4700; `65f75c8..e2e9cca` gleich (der Merge ändert keinen dieser Pfade). Stimmt mit dem Plan; Produktcode 506 Zeilen gegen *p* = 650. Ohne Befund.

## Zeile für Zeile (Mutation = eigene, Kürzel = Mutant)

| Zeile | Mutation (eigene, auf Kopie) | Ergebnis |
|---|---|---|
| Z1 | U+0000/Folge am Anfang nicht geprüft (`R01a`); Kodierung erst nach der Prüfung des Rands, `text` leer oder mit Leerzeichen am Anfang (`R01c`) | `R01a` rot (`TestLeserArgument`); `R01c` **grün** (F-2) |
| Z2 | `le` gelesen als `lt` (`R02a`, `le` nannte der Implementer nicht); Operanden nur bei `lt` vertauscht (`R02b`) | beide rot (`TestVergleichOperanden`) |
| Z3 | `or` mit vertauschten Operanden (`R03a`) | rot (`TestVerknuepfungen`) |
| Z4 | `ne` auf Stufe `gt` (`R04a`); `le` auf Stufe `eq` (`R04c`) | beide rot (`TestRangfolge`, `TestGeneratorFilter`) |
| Z5 | `eq`/`ne` rechtsassoziativ (`R05a`) | rot (`TestAssoziativitaet`) |
| Z6 | `klammern` nach der Gruppe auf 0 gesetzt statt abgebaut (`R06a`) | rot (`TestKlammern`) |
| Z7 | `maxSchluesselwort` 4, `divby` nicht erkannt (`R07a`); `mod` nicht in der Liste (`R07b`, `mod` nannte der Implementer nicht) | beide rot (`TestSchluesselwoerter`) |
| Z8 | Token `=` am Operatorplatz als `filter` (`R08a`) | rot (`TestOperatorUnbekannt`) |
| Z9 | Operand `mod` als Operatorwort (`R09a`) | rot (`TestOperatorwortAlsName`) |
| Z10 | Namenslänge nach Bytes statt Zeichen (`R10c`); `nurNamenzeichen` entfernt (`R10r`) | `R10c` rot, aber nur in `TestNamenAchsen`/`TestNamenStellungen` (Pfad-Parser), nicht in `TestNamen` (F-11); `R10r` **grün** (F-7) |
| Z11 | `~` als Namenszeichen an beiden Stellen (`R11c`) | rot (`TestZeichenAussenGrammatik`); Kommentarformen am Operandenplatz: F-6 |
| Z12 | ein einzelnes Leerzeichen am Anfang zugelassen (`R12a`) | rot (`TestLeerzeichen`) |
| Z13 | `\r` beendet den Lauf (`R13a2`, mit Panik im Leser); NBSP (C2 A0) als Leerraum (`R13b`) | beide rot (`TestAndererLeerraum`) |
| Z14 | Grenze der Ziffernläufe 513 (`R14c`) | rot (`TestLiteralLaenge`) |
| Z15 | `true`/`false` nicht kanonisch (`R15a`); `+` beginnt keinen Literal-Lauf (`R15b`) | beide rot (`TestLiteraleWieSchluessel`) |
| Z16 | `)` beendet die Zeichenkette nicht (`R16a`) | rot (`TestLiteralEnde`) |
| Z17 | Schreibweise `nULL` als Name (`R17a`) | rot (`TestNull`) |
| Z18 | Wurzel `null` abgelehnt (`R18a`) | rot (`TestWurzelOhneTyp`) |
| Z19 | Pfad ab fünf Gliedern `wert` (`R19a`) | **grün** (F-9) |
| Z20 | Ein-Zeichen-Name vor `(` kein Aufruf (`R20a`); `_`-Name (`R20b`); `$`-Lauf (`R20c`) | `R20a` rot, aber nicht in `TestAufruf`; `R20b`, `R20c` **grün** (F-3) |
| Z21 | `$root` nicht erkannt (`R21a`); `$this` (`R21d`); `@` mit Nicht-Namen (`R21b`) | alle rot (`TestNichtUnterstuetzt`) |
| Z22 | Tiefe nach der Gruppe auf 0 gesetzt statt abgebaut (`R22f`) | **grün** (F-4) |
| Z23 | Paare am selben Token: Aufruf gegen ungültiges Zeichen (`R23b`) | **grün** (F-2); Paare an verschiedenen Token rot (`R35b`, `TestRangImFilter`) |
| Z24 | Name im Ort der Ursache über 128 Zeichen (`R24c`) | rot (nicht in `TestFilterFehler`, aber über `flUrsache` in `TestNamen`) |
| Z25 | Ort der 501-Ursache fest `query/1` (`R25e`) | rot (`TestFilterFehler`, `TestFilterMitLeser`) |
| Z26 | Hex-Ziffer `b` nicht dekodiert (`R26b`, Element ohne Schlüssel im Plan) | rot (`TestGleicherWert`, Generatoren des Pfad-Parsers); nicht in `TestFilterDekodierung` (F-11) |
| Z27 | `)` am Operatorplatz immer als Ende (`R27b`) | rot (`TestStruktur`) |
| Z28 | `klammern` als Paketvariable (`R28a`); letzter Baum gespeichert (`R28b`) | siehe (2) und F-13 |
| Z29 | Tiefenzähler als Paketvariable (`M29`) | siehe Raten |
| Z30 | `map` als Paketvariable (`M30`) | siehe Raten |
| Z31 | Kopie des Rests je Zeichenketten-Doppelanführung (`R31c`); Kopie je Lauf (`R31d`) | beide rot (`TestFilterWaechter`) |
| Z32 | akzeptiertes Negativ, Grund und Folge im Plan und in `SPEC-164` *Ungeprüft* | nicht gefahren |
| Z33 | `or` als `OR` (`R33a`); `Operand` von `not` ohne Wert (`R33c`) | beide rot (`TestFilterKnoten`, `TestFilterKnotenKorpus`); Zusage „keine Slice/Map/Zeiger“: F-5 |
| Z34 | `or` rechtsassoziativ (`R34b`) | rot (`TestGeneratorFilter`) |
| Z35 | bei Pfad und Schlusszeichen `(` weiterlesen und `wert` melden (`R35b`) | rot (`TestStopHinterNichtUnterstuetzt`, `TestRangImFilter`) |

**Raten Z29/Z30 (20 Prozesse je Zelle, ich habe gemessen; Standard-`GOMAXPROCS` im Container mit `--cpus=4`):**

| Variante | ohne Detektor | mit Detektor |
|---|---|---|
| unverändert | 0 von 20 | 0 von 20 |
| `M29` (Tiefenzähler als Paketvariable) | **20 von 20** (Wirkungs-Fänger `TestFilterGleichzeitig`) | **20 von 20** (`DATA RACE`) |
| `M30` (`map` als Paketvariable) | **20 von 20** (`fatal error: concurrent map writes`) | **20 von 20** (`DATA RACE`) |
| `M29` mit `-test.cpu=2` | 20 von 20 | – |
| `M30` mit `-test.cpu=2` | 20 von 20 | – |
| `M29` mit `-test.cpu=1` | **0 von 20** | **0 von 20** |
| `M30` mit `-test.cpu=1` | **0 von 20** | **0 von 20** |

Die Raten des Implementers (20 von 20, unverändert 0 von 10) stimmen für die Bedingung mit mehreren Prozessoren; die Bedingung nennt der Plan nicht (F-8).

**Mengen als Klasse (§3.17).** Der Generator `TestGeneratorFilter` (2000 Läufe, Saat 1 bis 2000) hat ein Orakel, das der Leser nie ist: der Renderer des Tests mit minimaler Klammerung und dem Stufen-Orakel `flStufe`; Mutationen der Rangfolge und Assoziativität (`R04a`, `R04c`, `R05a`, `R34b`, `R02a`, `R15a`, `R17a`) werden in ihm rot. Mein Gegen-Generator (volle Klammerung ohne Stufen-Orakel, Blätter mit Unicode-Namen, Guid, Zeitpunkt mit Versatz, Zeichenketten mit `and (`, Schlüsselwörter in gemischter Schreibweise, 3000 Saaten ab 1) liefert **0 Abweichungen**. Die Blattmenge des Implementer-Generators ist klein und fest (neun Namen, fünf Zeichenketten, je ein Guid, Datum, Zeitpunkt); die Literal-Formen deckt Z15 als Differenzialtest. Elemente, die der Implementer nicht nannte und die rot werden: `le`, `ne`, `mod`, `$this`, `$root`, `~`, `\r`, NBSP, `+`, `TRUE`, `nULL`, `or`-Operandentausch.

**Randformen (Probe, rund 200 Eingaben).** Alle entsprechen `SPEC-164` bis auf die in F-1 und F-6 genannten: leer, nur Leerzeichen, Tabulator, U+0000, Folge `0xFF`, Surrogate, Überlang-NUL, Codepunkt über U+10FFFF (alle `kodierung`/`wert` wie Nr. 1, 2); Namen mit `L`, `Lt`, `Nd` hinter dem ersten Zeichen gültig, `Nl`, `No`, `M`, `Nd` am Anfang `wert`; Name 128/129 Zeichen auch aus Zweibyte-Zeichen; Literal 512/513 Bytes; Tiefe 32/33 für `(`, `not` und gemischt; Kette von 1699 `or`; `not(Id)`, `NOT Id`; `a/b(1)` `filter`; `$it/` `filter`; `Id eq X'AB'` `literal_typ`.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Bei derselben Zeichenkette gelten zwei Verstöße (`zu_lang` und `literal`: ohne schließendes `'` oder mit Rest hinter ihm, jeweils über 512 Bytes); der Code meldet `zu_lang` zuerst, `SPEC-164` Nr. 4 a („Ohne schließendes `'` ist es `literal`“) und Nr. 7 entscheiden das nicht, der Kommentar „die Länge geht vor der Form“ verweist auf keine Stelle. Vertauscht man die beiden `case`-Zweige (`R14d`), bleibt jeder Test grün. Ein ungeschlossenes `'x` mit 600 Bytes Rest ist so `zu_lang`, obwohl der Rest kein Literal ist. | `AGENTS.md` §3.16 (Randform im Code entschieden), §3.11 (Kommentar sagt zu, was kein Test prüft) | `filter.go` · `case len(t.text) > maxLiteral:` und `die Länge geht vor der Form` | ja — `R14d` grün; Probe `'` + 600 Bytes: `zu_lang` | Randform im Code entschieden |
| F-2 | MEDIUM | Zwei Paare von Verstößen am selben Text haben keinen Fänger: unmittelbarer Aufruf mit führendem Leerzeichen und ungültigem UTF-8 (Kodierung geht vor `wert`; `R01c` grün) und ein Wortlauf mit ungültigem Zeichen unmittelbar vor `(` (Aufruf `filter` geht vor `wert`, `a#(1)`; `R23b` grün). Z1 und Z23 führen Paare nur an verschiedenen Token. | `AGENTS.md` §3.16 (Paar von Verstößen), Reviewer-Skill *Paar von Ursachen ohne Fänger* | `filter.go` · `if !gueltigerText(text) {` und `case t.vorKlammer:` | ja — `R01c`, `R23b` grün | Paar von Ursachen am selben Token ohne Fänger |
| F-3 | MEDIUM | `TestAufruf` führt als Namen vor `(` nur Namen mit Buchstabe am Anfang und kein Ein-Zeichen-Wort: ein Aufruf mit `_`-Namen (`R20b`) und mit `$`-Lauf (`$x(1)` müsste nach Nr. 4 c (1) `filter` sein, `R20c`) bleibt grün; die Ein-Zeichen-Mutation `R20a` wird nicht in `TestAufruf`, nur zufällig in `TestFilterFehler` rot. Die Allaussage „jeder Name unmittelbar vor `(`“ steht als Liste von Buchstaben-Namen. | `AGENTS.md` §3.15 und §3.17 (Fänger führt jede Form der Eingabe), Reviewer-Skill *Fänger führt nur eine Form der Eingabe* | `filter_lexik_test.go` · `func TestAufruf` ; `filter.go` · `case t.vorKlammer:` | ja — `R20b`, `R20c` grün | Fänger führt nur eine Form der Eingabe |
| F-4 | MEDIUM | Die Tiefe wird nach dem Schließen einer inneren Gruppe nur dann sicher abgebaut, wenn nichts mehr offen ist: setzt man sie stattdessen auf 0 (`R22f`), bleibt `TestTiefe` grün, weil kein Fall nach dem Schließen einer inneren Gruppe weiter öffnet (z. B. 20 offene `(`, dann `(a)`, dann 20 `not`). Die Grenze 32 ließe sich so beliebig überschreiten. | `AGENTS.md` §3.15, Reviewer-Skill *Fänger führt nur eine Form der Eingabe* | `filter.go` · `p.klammern--` (Zeile danach `p.tiefe--` in `gruppe`) ; `filter_grenzen_test.go` · `func TestTiefe` | ja — `R22f` grün | Fänger führt nur eine Form der Eingabe |
| F-5 | MEDIUM | `SPEC-165` Nr. 2 sagt zu: „Der Baum enthält keine Slice, keine Map und keinen Zeiger“, und der Plan Z28 verspricht „Ändern eines Wertes am ersten Ergebnis verändert das zweite nicht“. Kein Test prüft die Gestalt der sechs Typen (`R33d`: ein Feld `Teile []string` in `FilterEigenschaft` — alles grün) und kein Test ändert ein Ergebnis. | `AGENTS.md` §3.15 (Zusage ohne Fänger ist ein MEDIUM); `SPEC-165` Nr. 2 | `filter_ast.go` · `type FilterEigenschaft struct` ; `filter_ast_test.go` · `func TestFilterKnoten` | ja — `R33d` grün | Zusage ohne Fänger |
| F-6 | MEDIUM | Die Kommentarformen von Z11 stehen nur hinter einem Operanden (`"Id eq 1 "+k`), also am Operatorplatz, wo jedes Token `wert` ist. Am Operandenplatz ist `Id eq -- c` `filter` (Minus-Lauf nach Nr. 4 b/5 c), während Nr. 2 `--` als „Zeichen ohne Rolle (`wert`)“ nennt; `//`, `#`, `;`, `/*` sind dort `wert`. Der Code entscheidet den Widerspruch der beiden Stellen stillschweigend. | `SPEC-164` Nr. 2 gegen Nr. 4 b und 5 c; `AGENTS.md` §3.16 | `filter_lexik_test.go` · `flFehler(t, "Id eq 1 "+k, model.URLWert)` ; `filter.go` · `case c == '-':` | ja — Probe `-- c`, `Id eq -- c`: `filter` | Randform im Code entschieden |
| F-7 | LOW | `nurNamenzeichen` ist ersetzbar, ohne dass etwas rot wird (`R10r`: Körper `return true`, alle Tests grün): `einfacherName` entscheidet jeden Fall gleich. Die Regel „Name“ steht zweimal, in `nurNamenzeichen` und in `namenAnfang` (`dekodieren.go`); eine Änderung an einer Stelle bleibt wirkungslos. | `AGENTS.md` §3.15 (eine Prüfung, die sich entfernen lässt, ohne dass etwas rot wird) | `filter.go` · `func nurNamenzeichen(s string) bool {` | ja — `R10r` grün | Redundante Prüfung ohne Wirkung |
| F-8 | LOW | Die Raten von Z29/Z30 hängen an der Parallelität: bei `GOMAXPROCS` 1 (`-test.cpu=1`) fangen weder der Wirkungs-Fänger noch der Detektor die Mutationen `M29`/`M30` (je 0 von 20), ab 2 sind es 20 von 20. Der Plan nennt Zahl der Prozesse und Messenden, aber nicht diese Bedingung. | `AGENTS.md` §3.10 (Rate mit Zahl der Prozesse und Messendem) | `filter_dienst_test.go` · `func TestFilterGleichzeitig` ; Plan §3 *Raten der nebenläufigen Zeilen* | ja — `-test.cpu=1`, 20 Prozesse | Rate ohne Bedingung der Parallelität |
| F-9 | LOW | Die Menge „Pfad“ (Z19) ist unbeschränkt (beliebig viele Glieder), der Test führt Pfade bis vier Glieder; ein Grenzwert ab fünf Gliedern (`R19a`) bleibt grün. Eine Prüfgrenze mit Grund und Folge nennt der Plan nicht. | `AGENTS.md` §3.17 (unbeschränkte Menge: Prüfgrenze benennen) | `filter_lexik_test.go` · `func TestPfad` ; `filter.go` · `func pfad(s string)` | ja — `R19a` grün | Menge unbeschränkt, Prüfgrenze nicht benannt |
| F-10 | LOW | Der Kommentar von `zerleger` sagt „geht nie zurück“ zu; Z32 ist als akzeptiertes Negativ ausgewiesen (Rücklauf ohne Zuteilung bleibt grün), kein Test prüft die Zusage, und der Kommentar sagt nicht, was offen bleibt. | `AGENTS.md` §3.11 | `filter_token.go` · `und geht nie zurück` | nein — Laufzeit ohne Zuteilung fängt kein Gate | Kommentar sagt zu, was kein Test prüft |
| F-11 | INFO | Z10: `TestNamen` fängt keinen Zeichenzähler nach Bytes (`R10c`); rot werden nur `TestNamenAchsen`/`TestNamenStellungen` des Pfad-Parsers, die denselben `einfacherName` rufen. Ebenso Z26: die Hex-Ziffer `b` (`R26b`) fängt `TestFilterDekodierung` nicht, rot werden die Tests des Pfad-Parsers. Die Wirkung ist gedeckt, die Zeile nicht von ihrem eigenen Fänger. | Maintainability | `filter_lexik_test.go` · `func TestNamen` | ja — `R10c`, `R26b` | — |
| F-12 | INFO | Folge der offenen Rückgabe (3): eine Guid mit `a` bis `f` am Anfang (6 von 16 Anfangsziffern, rund 37 %) ist im Filter `wert`/400, im Schlüssel eine Guid. Der Code setzt den Wortlaut um, §6 führt Ausgang *offen* und Adresse Architect; bei einer Entscheidung ändert sich der Test an `flAusnahmen15`. | `AGENTS.md` §3.13 | `filter_literal_test.go` · `func flAusnahmen15` | ja | — |
| F-13 | INFO | Z28: Zustand in einer Variablen des Pakets (`R28a`) fängt `TestFilterWiederholung` nicht; es genügt die Streuung zwischen den Tests in `TestStruktur`. Die Zusage für Paketvariablen trägt allein der Detektor (`SPEC-122` Nr. 8) und der Wirkungs-Fänger von Z29. | `SPEC-122` Nr. 8 und 9 | `filter_dienst_test.go` · `func TestFilterWiederholung` | ja — `R28a` | — |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/filter_ast.go`, `filter_ast_test.go` | geprüft; bis auf F-5 ohne Befund (Gestalt, Wertempfänger, Operatortexte nach `SPEC-165` Nr. 1 und 3) |
| `src/adapters/driven/odata/urlparser/filter.go`, `filter_token.go` (Grammatik, Rangfolge, Lexik, Tiefe, Literale, Sonderformen) | geprüft gegen `SPEC-164` Nr. 1 bis 9; Randformen der Probe stimmen mit dem Wortlaut überein, Befunde F-1 bis F-4, F-6, F-7, F-10 |
| `filter_*_test.go` (sieben Dateien) | geprüft; Befunde F-3, F-4, F-8, F-9, F-11, F-13; Generator mit unabhängigem Orakel, keine Abweichung der 3000 Saaten meines Gegen-Generators |
| Aufruf des Lesers aus `ParseURL` (`query.go`, unverändert) | geprüft ohne Befund (Z25: `R25e` rot) |
| Composition Root, Gate-Kette, `Makefile`, `harness/` | geprüft, ohne Befund (kein Aufruf von `NewFilterLeser` außerhalb des Pakets, Gate-Kette unberührt) |
| `docs/plan/planning/done/slice-odata-url-filter.md` (§3 Tabelle, Raten, Marke, §6 Raster) | geprüft ohne Befund außer F-8 (Bedingung der Rate) und F-9 (Prüfgrenze); Messung 1911 nachgemessen; Raster ohne leere Zelle, Rückgabe mit Ausgang *offen* und Adresse |
| Spezifikation (`SPEC-164`, `SPEC-165`) | vom Implementer unverändert (`git diff 65f75c8..84b24a5` ohne `spec/`); der Widerspruch von Nr. 2 und Nr. 4 b/5 c geht als F-6 an den Architect |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 6 |
| LOW | 4 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Randform im Code entschieden · Paar von Ursachen am selben Token ohne Fänger · Fänger führt nur eine Form der Eingabe · Zusage ohne Fänger · Redundante Prüfung ohne Wirkung · Rate ohne Bedingung der Parallelität · Menge unbeschränkt, Prüfgrenze nicht benannt · Kommentar sagt zu, was kein Test prüft

## Verdikt

**Merge-blockierend:** ja — sechs MEDIUM-Findings, kein HIGH. F-1 und F-6 nennen Randformen, die der Code entscheidet; sie gehen nach `AGENTS.md` §3.12 an den Architect (F-1: Reihenfolge von `zu_lang` und `literal` an derselben Zeichenkette; F-6: `--` am Operandenplatz), vom Implementer ist dort keine Entscheidung im Code zu treffen. F-2 bis F-5 sind Fänger-Lücken; der Fix deckt die Klasse (`AGENTS.md` §3.17): für F-3 die Menge der Lauf-Anfänge vor `(` (Buchstabe, `_`, `$`, `@`, Ein-Zeichen-Name) × Formen, für F-4 jede Stelle, an der nach dem Schließen einer inneren Gruppe weiter geöffnet wird (Gruppe, `not`), für F-5 die Gestalt aller sechs Typen.

**Übergabe:** Findings gehen an den Implementer, F-1 und F-6 zusätzlich an den Architect (Adresse in §6 des Plans mit Ausgang *offen* im selben Commit wie die Übergabe, `AGENTS.md` §3.13). Die Finding-Klassen gehen in die Slice-Closure §7 und von dort in den Zähler. Der Report ersetzt keine Verifikation; DoD- und Spec-Konformität prüft der Verifier.

**Rohbytes (`AGENTS.md` §3.18):** Dieser Report enthält keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM und kein NBSP; gegengelesen wurde `docs/reviews/2026-10-31-slice-odata-url-filter-review.md` mit `LC_ALL=C grep -P`.
