# Review slice-prefer-return, Runde 4 (Fix ad8e4f01)

* Rolle: Reviewer, frischer Kontext. Gegenstand: `git show ad8e4f01` (Fix-Runde 3; `e3c2a864` ist nur ein ADR-Link im Report von Runde 3) gegen die Klasse (`AGENTS.md` §3.17), §3.11, §3.9.
* Vorrunden: `2026-11-18-slice-prefer-return-review.md`, `-review-2.md`, `-review-3.md`.
* Bezug: LH-FA-07, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).
* Messung: `sh tools/harness/mutationslauf.sh` (Image `odata-recorder:race`, netzlos, unter `flock` auf `.wt/heavy.lock`), Klasse `andere`, 12 Mutationen in `src/hexagon/model/rueckgabe.go`, je ein Prozess je Stufe (zwei Stufen: benannter Test, dann Paket), ohne Detektor; Last des Hosts 5,9 bis 10,9 von 4 CPUs (Spalte `last`); Messender: der Reviewer. Die Listen liegen nicht im Repo.
* Menge, vom Reviewer aus `SPEC-285` Nr. 1 bis 3 gebildet: Position des Backslash (Zeilenanfang, Elementanfang, Name, hinter dem Namen, um `=`, Wert als Token, Wort in Anführungszeichen, Paar im Namen, Paar im Rest, hinter `;`, hinter `,`, Zeilenende, dasselbe im Element `return`) × Lage des Elements in Zeile und Anfrage × Länge der Folge × Zeichen hinter der Folge × **Zeichen vor der Folge** × **Kontext des Paares** (erstes oder späteres Paar der Zeile, Paar am Anfang eines Elements, hinter `;`, unmittelbar hinter einem Paar). Der Generator führt die ersten vier Achsen; die beiden fettgedruckten fehlen ihm.

## Findings

### MEDIUM

- kategorie: MEDIUM
- quelle: `AGENTS.md` §3.15, §3.17 (Fänger schmaler als Zusage; Menge ohne Achsenliste)
- pfad: `src/hexagon/model/rueckgabe_test.go` · `TestLiesRueckgabeBackslashStellen` (`stellen := backslashStellen()`, 71 Stellen × 6 Lagen × `k` 0 bis 7 × 7 Zeichen hinter der Folge)
- befund: Der Generator variiert Position, Lage, Länge und das Zeichen *hinter* der Folge, nicht das Zeichen *vor* ihr und nicht den Kontext des Paares. Sieben Mutationen an Elementen der Klasse, die weder ein Befund noch der Bericht des Implementers nannte, bleiben in Stufe 2 (ganzes Paket `model`) grün, zwei weitere werden nur von einem anderen Test des Pakets rot. In `elementEnde`: der Backslash schützt nicht, wenn das Zeichen davor ein Leerzeichen ist (`if z[i] == '\\' && !(i > 0 && z[i-1] == ' ')`), ebenso nach `;`, nach Tabulator, nach `=`; der Backslash schützt nur im ersten Paar der Zeile (Rot nur durch `TestLiesRueckgabeAnfuehrung` und `TestLiesRueckgabeZeit`, nicht durch den benannten Test); ein Paar am Anfang eines Elements, unmittelbar hinter `;` und unmittelbar hinter einem Paar wird ohne Backslash-Schutz gelesen (drei Mutationen, alle grün). In `liesReturnWert`: ein Backslash hinter Tabulator und Wort wird als Ende akzeptiert (Rot nur durch `TestLiesRueckgabeFolgen`). Belege, je ein Eingabewert, der von Spezifikation und Orakel als `Keine` gelesen wird und mit der Mutation `Minimal` ergibt: `x="a \", return=minimal;"`, `x="a;\", return=minimal;"`, `x="a<TAB>\", return=minimal;"`, `x="a=\", return=minimal;"`, `"a\",return=minimal;"` (Paar am Elementanfang), `x=1;"a\", return=minimal;"`, `x="a""b\", return=minimal;"`. Ein Probe-Test mit diesen sieben Werten war für die sieben Mutationen rot und ist wieder gelöscht (`git status` sauber). Das Zeichen vor der Folge ist in der Menge nur durch `,` gedeckt (diese Mutation ist im benannten Test rot); die übrigen Zeichen, die in der Grammatik eine Rolle spielen, sind es nicht.
- verifizierbar: ja (Mutationen oben; Zeile der Liste je Name `r4-leerzeichen-vor-backslash-im-paar`, `r4-semikolon-…`, `r4-tab-…`, `r4-gleich-…`, `r4-paar-am-elementanfang-naiv`, `r4-paar-hinter-semikolon-naiv`, `r4-paar-hinter-paar-naiv`, dazu `r4-zweites-paar-ohne-schutz` und `r4-tab-vor-backslash-hinter-wort` nur rot-paket)
- klasse: Klasse „Stelle der Grammatik" nach den Achsen der Befund-Beispiele gebildet, nicht aus dem Wortlaut (§3.17); dritte Runde in Folge

### LOW

- kategorie: LOW
- quelle: `AGENTS.md` §3.11 (Kommentar sagt nur zu, was ein Test prüft)
- pfad: `src/hexagon/model/rueckgabe_test.go` · `TestLiesRueckgabeBackslashStellen`: „ein Backslash an jeder Stelle der Grammatik"; Plan `slice-prefer-return.md` · Zeile 6b: „die Menge aus der Grammatik von `SPEC-285` Nr. 1 bis 3, jede Position, an der ein Backslash in einer Kopfzeile stehen kann"
- befund: Der Kopf und die Zeile des Plans sagen „jede Stelle"/„jede Position" zu. Der Kopf nennt als ungeprüft zwei Dinge (ein weiteres Zeichen hinter der Folge, mehr als ein Element zwischen Stelle und `return`); das Zeichen vor der Folge und die Lage des Paares (Elementanfang, hinter `;`, unmittelbar hinter einem Paar, späteres Paar der Zeile) nennt weder er noch die Prüfgrenze im Plan. Der Satz „die Funktion liest jede Stelle einheitlich" der Prüfgrenze ist eine Behauptung ohne Fänger; die Mutationen oben zeigen Stellen, an denen eine Lesart unterscheiden würde.
- verifizierbar: ja (Kopf gegen Mutationen)
- klasse: Kommentar und Plan nennen mehr Elemente als der Test (§3.11)

- kategorie: LOW
- quelle: `AGENTS.md` §3.9 (der Slice-Plan folgt jeder Korrektur)
- pfad: `docs/plan/planning/in-progress/slice-prefer-return.md` · „Zeile 6c nennt die Zusage und den Fänger; EA2 und EA8 rot in `TestLiesRueckgabeBackslashStellen`" (N-6) und „Zeile 6b, 6c, die Mengentabelle und die Mutationstabelle nennen je den Test, der rot wird"
- befund: Die Tabelle *Zusage · Fänger* in §3 hat die Zeilen 6a und 6b, aber keine Zeile 6c; Zeile 6b trägt beide Tests und die Mutationen von 6c. Mengentabelle und Mutationstabelle führen 6b und 6c getrennt. N-6 und die Fix-Runde-3-Tabelle verweisen auf eine „Zeile 6c" der Zusage-Tabelle, die es dort nicht gibt.
- verifizierbar: ja (Suche nach `| 6c` in §3 der Zusage-Tabelle)
- klasse: Zeilenkennung im Plan an einer Tabelle nachgezogen, an der anderen nicht

- kategorie: LOW
- quelle: `AGENTS.md` §3.9, §3.7 (Zustandsfeld nennt den Zustand)
- pfad: `docs/plan/planning/in-progress/slice-prefer-return.md` · „Messpunkt (`SPEC-128` Nr. 9, Stand Handoff, nach der Fix-Runde 2)"
- befund: Der Wert `messung=2298` ist nach dem Commit der Fix-Runde 3 gesetzt (`make groessenmarke` meldet auf `HEAD` 2298 von Marke 3500, gleich dem Plan), die Beschriftung nennt weiter die Fix-Runde 2. Die Zahl stimmt, der Zustand, den sie beschreibt, nicht.
- verifizierbar: ja (`make groessenmarke`)
- klasse: Beschriftung eines Zustandsfeldes beim Nachziehen der Zahl nicht mitgezogen

### INFO

- INFO-1 (positiv, Klasse aus Runde 3): Die genannten Mutationen der Runde 3 liegen im Generator: je eine neue Mutation an Elementen, die Befund und Bericht nicht nannten, ist rot im benannten Test, nämlich Backslash vor einem Zeilenende überspringt die Zeile (`r4-zeilenende-backslash-ueberspringt-zeile`, rot) und Backslash nach `,` im Paar schützt nicht (`r4-komma-vor-backslash-im-paar`, rot). Die Zählung 71 × 6 × 8 × 7 = 23856 wird im Test geprüft (`t.Fatalf` bei ungleich 71, `t.Errorf` bei ungleicher Menge).
- INFO-2 (Orakel): `orakelZeilen`, `preferElement`, `preferReturn`, `preferWert`, `preferBackslash` teilen keinen Code mit `LiesRueckgabe`; Lesart und Spezifikation stimmen an den untersuchten Eingaben überein (die sieben Probewerte oben liefern im Orakel `Keine`).
- INFO-3 (gleichwertige Mutation): `r4-leerzeichen-vor-backslash-im-wort` (Backslash nach Leerzeichen im Wort in Anführungszeichen) bleibt auch im Probe-Test grün. Ein Wort mit einem Leerzeichen kann weder `minimal` noch `representation` sein, und ein offenes Wort ergibt wie ein aufgelöstes `Keine`; die Mutation ändert kein Ergebnis und ist kein Befund.
- INFO-4 (Kennungen): Die Kennungen `B1`, `B2` stehen im Plan doppelt: Zeile 6a (Zeichen nach dem Backslash gestrichen, der Backslash bleibt) und Fix-Runde 3 / Zeile 6c (hinter dem Schlusszeichen eines Paares, am Elementanfang). Die Zeilen sind unterscheidbar, die Namen nicht.
- INFO-5 (Steering-Loop, `.harness/skills/reviewer.md` §Pflege): Das Muster „Menge nach den Beispielen des Befunds gebildet, eine Achse fehlt" ist die dritte Runde in Folge (Runde 2: Stellen außerhalb eines Paares, Runde 3: hinter dem Schlusszeichen, Runde 4: Zeichen vor der Folge und Kontext des Paares). Der Zähler ist drei; ein Widerspruch des Implementers liegt nicht vor, der Konflikt-Pfad über den Architect ist damit nicht ausgelöst.
- INFO-6 (Betriebsregel): Der Reviewer hat die temporäre Probe-Datei mit einem Heredoc geschrieben, was die Betriebsregeln des Plans für Repo-Dateien verbieten; die Datei ist gelöscht und nie eingecheckt (`git status` leer); die zwei Korrekturen dieses Reports entstanden mit `sed -i` aus demselben Grund.

## Geprüft, ohne Befund

- `src/hexagon/model/rueckgabe.go` (Kopfkommentar von `LiesRueckgabe`, `elementEnde`, `liesAnfuehrung`): geprüft, ohne Befund am Kommentar; die Mutationen oben betreffen den Test, nicht den Code.
- `src/hexagon/model/rueckgabe_test.go` (`orakelZeilen`, `backslashLagen`, Zählung): geprüft, ohne Befund außer den beiden Befunden oben.
- `docs/plan/planning/in-progress/slice-prefer-return.md` (Mengentabelle Zeile 6b/6c, Mutationstabelle, Fix-Runde 3, Messpunkt): geprüft; Befunde LOW oben, im Übrigen ohne Befund.
- `make groessenmarke`: grün, Messung 2298 von Marke 3500, gleich dem Plan.

## Verdikt

Nicht freigabefähig: ein MEDIUM (Klasse „Backslash an jeder Stelle" ist um die Achsen Zeichen vor der Folge und Kontext des Paares schmaler als die Zusage von `SPEC-285` Nr. 2) und drei LOW offen. Die Fix-Runde 3 deckt die Mutationen der Runde 3 und die Positionen, die sie bildet; sie deckt nicht die Klasse.
