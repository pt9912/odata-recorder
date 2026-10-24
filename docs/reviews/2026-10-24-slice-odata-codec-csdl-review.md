# Review-Report: slice-odata-codec-csdl — 2026-10-24

**Review-Art:** Code — geprüft gegen Plan (`slice-odata-codec-csdl` §1, §3 mit der Tabelle *Zusage · Fänger* A1 bis J2 und dem Ausgang des Implementers, §4, §6 mit Raster und Zusatztabelle P1 bis P22), Vertrag `SPEC-125` und Hard Rules (`AGENTS.md` §3.1 bis §3.16), nicht gegen die DoD (Verifier).

**Gegenstand:** `slice-odata-codec-csdl` (Welle `welle-record-replay`), `git diff 3b6b3b8..9b9a4fa` mit HEAD `9b9a4fa`. Pfade: `src/hexagon/model`, `src/hexagon/ports/outbound` (mit `outboundtest/` und `testdata/`), `src/adapters/driven/odata/csdl`, `.a-check.yml`, eine Zeile in `src/hexagon/services/session/beenden_test.go`, der Plan.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `9b9a4fa` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-24

**Methode.** Der Bericht des Implementers (Plan §3) wurde gelesen, aber keine Mutation übernommen. Eigene Mutationen: 256 Einzelmutationen plus 5 Ersatzmutationen für Übersetzungsfehler, dazu 11 Mutanten für die Raten-Reihen. Jede in einer Wegwerfkopie aus `git archive 9b9a4fa` (`go.mod`, `go.sum`, `src/`), Ersetzung per eigenem Werkzeug im Scratchpad, nie im Repo; Lauf `go test -count=1` über `csdl`, `model`, `outbound` im Image `odata-recorder:deps-verif` (`docker run --network none`, Quelle `:ro`, `--memory=4g --memory-swap=4g`, höchstens 3 Container zugleich). Rot zählt nur, wenn ein Test der Zeile rot wird, die die Mutation treffen soll. Ergebnis: 190 rot, 58 grün (davon 9 äquivalent oder nicht beobachtbar, s. INFO), 8 Übersetzungsfehler (5 ersetzt, 3 gleichwertig zu schon gefahrenen), 0 Lauf OOM-gekillt oder über die Zeitgrenze. Keine Mutation entfernte eine Grenze oder Tiefe; die Grenzen-Mutationen (`>=`, ±1, Zähler beim Schließen, Summe je Typ) endeten alle in Sekunden rot. Raten: Testbinary je Mutant, N Prozesse mit `-test.count=1`; mit Detektor im Image `odata-recorder:race` (cgo). Messender: der Reviewer, Host mit 20 Kernen, Last 2 bis 4. Zusätzlich 24 Probe-Dokumente (Ersatzreferenzen, Zyklen, Namen, Größen) gegen den Leser. J1: Kopie des Repos mit `a-check`-Image (Pin des Repos), sieben Fälle.

**Eingangs-Kontext:**

- Slice-Plan `slice-odata-codec-csdl` in `9b9a4fa` (§1, §3, §4, §6, §8)
- `SPEC-125`, `SPEC-109` Nr. 5, `SPEC-107` Nr. 14/15, `SPEC-049`, `SPEC-002`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `LH-FA-06.a`, `LH-RB-01`, `LH-QA-07.a`
- `AGENTS.md` (Hard Rules §3.1 bis §3.16)

---

## Antworten auf die Fragen des Auftrags

**(1) Tabelle A1 bis J2 und Raster.** Jede der 44 Zeilen hat einen Test unter dem genannten `t.Run`-Namen. Von den Mutationen des Implementers ist jede, die ich nachgefahren habe, aus dem genannten Grund rot. Meine eigenen Mutationen lassen 49 beobachtbare Formen grün; sie sind in den Findings F-1 bis F-14 nach Zeile gebündelt (Zeilen A2, B2, B4, B5, C1, D1, D6, D7, D8, F5, G1). Zeilen ohne grüne Eigenmutation: A3, A5, B1, B3, B6, C2, D2, D3 (außer Zyklus über 64 Typen, F-13), D4 (alle 34 Namen und 4 weitere Typen je einzeln vertauscht: 38 von 38 rot), D5, D9, D10, E1, E2, E3, F1, F2, F3, F4, F6, G2, G3, G4, G5, H1, H2, I1, I2, I3, J1, J2. Die Zellen des Rasters nennen alle einen Ort, dessen Wortlaut die Entscheidung trägt, bis auf die in F-1 genannten; die Zusatztabelle ist unvollständig (F-8).

**(2) Raten.** Siehe Tabelle am Ende. Alle Fänger-Mutationen der Zeilen I1, I2, I3 sind in jedem gemessenen Prozess rot, unmutiert in jedem grün; auch unter Drosselung (`--cpus` 0,2 und 0,3).

**(3) Abweichungen des Implementers (Plan §3, Punkte 1 bis 8).** Punkt 1 bis 4 sind im Code entschiedene Randformen (F-1). Punkt 5 ist ein Code-Fehler der ersten Fassung, vom Vertrag gedeckt (`SPEC-125` Nr. 4 b), keine Lücke. Punkt 6 (Parametername) F-15. Punkt 7 ist keine Abweichung: der Plan nennt in §3 `outboundtest/` mit den Fixtures; die Marken-Pfade enthalten den Ort. Punkt 8 ist eine Abschwächung der Zusage (F-2).

**(4) Akzeptierte Negative.** (a) bis (f) des Implementers nennen Grund und Folge. Nicht gedeckt: das Negativ „A4 abgetastet“ (F-2). Die Negative der Architect-Liste (Typed-nil, Pfad eines Bindings, UTF-16) tragen Grund und Folge in `SPEC-125` *Ungeprüft*.

**(5) Kommentare §3.7, §3.11.** Produktcode und Tests tragen Zusage, Kopplung oder Abgrenzung; keine verworfene Alternative, keine Chronik. Kopf von `zeileA4` nennt, was offen bleibt (Dokumente über 5000 Bytes), und sagt „abgetastet“: gedeckt.

**(6) Sicherheit des XML-Lesers.** `DOCTYPE` in jeder Form ist `doctype`, bevor etwas gelesen wird; keine Entität wird expandiert (`Decoder.Entity` nicht gesetzt, `&foo;` ist `syntax`); `CharsetReader` liefert nie einen Reader; Tiefe 64, Größe `max+1`, Summe, Kette durchgesetzt. Probe gegen den Leser (Container 4 GiB): 1 000 000 Annotationen 2,2 s gesamt, 300 000 Attribute an einem Element 133 ms, 200 000 leere Typen 487 ms, 20 000 Typen in Ketten zu 64 → `zu_viele_properties` in 81 ms. Kein quadratisches Verhalten gefunden. Die Eingabe wird vollständig in `strom.alles` mitgehalten (INFO I-4).

**(7) Datenrennen, Fehlerabbildung.** Der Leser hält keinen Zustand (`J2`), `make test-race`-Image: unmutiert 15 von 15 Prozessen grün. `ErrMetadatenUngueltig` → `ODR-E6004` bei jedem Gegenstand: H2 fängt Wegfall der Zeile und Beschränkung auf einen Gegenstand (beide Mutationen rot).

**(8) Größe.** `git diff --numstat 3b6b3b8..9b9a4fa` über die Pfade der Marke: 3628 Zeilen (3627 Einfügungen, 1 Löschung), gleich der Messung des Implementers; Marke 2400, 51 % darüber. Ehrlich berichtet, mit Grund (INFO I-3).

**(9) `TestTransaktionVertrag` (Paket `sqlite`).** Nicht reproduzierbar: 40 von 40 Prozessen grün bei `--cpus=2`, 40 von 40 bei `--cpus=0,15` (Testbinary, `-test.count=1`, Reviewer, Host 20 Kerne). Der Test liegt nicht im Diff (INFO I-5).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Vier Randformen sind im Code entschieden, ohne dass `SPEC-125` eine Zelle trägt, und der Implementer hat nicht angehalten (§3.12): (a) Fund `annotation` auch als Kind von `Key`, `PropertyRef`, `NavigationPropertyBinding`, `Edmx`, `DataServices` (Nr. 7 nennt sieben Eltern); (b) Attribut mit Namensraum als `erweiterung` an gemeldeten Elementen, aber nicht an einem fremden Element; (c) `Target` eines Bindings bleibt der Name im Container, mit Container nur wenn so geschrieben (`Categories` neben `Demo.Container/Products` in demselben Modell, Nr. 11 sagt „qualifiziert“); (d) Ort bei `Edm` als `Namespace`/`Alias`. Zuständig: **Architect** (Spezifikation entscheidet, danach Fänger). | `AGENTS.md` §3.12, §3.16 | `src/adapters/driven/odata/csdl/elemente.go` · `func (l *lauf) annotation(p rahmen, se xml.StartElement)` und `func (l *lauf) erweiterung(` | nein | Randform im Code entschieden |
| F-2 | MEDIUM | Der Implementer hat die Zeile A4 des Architect geändert („jede Zweiteilung für Dokumente bis 200 Bytes“ → „Zweiteilungen an den ersten und letzten fünf und an 16 Stellen“), `SPEC-125` Nr. 2 sagt weiter „jede Zweiteilung“, und das Negativ nennt weder Grund noch Folge. Eine Zusage der Spezifikation hat einen schwächeren Fänger als der Vertrag zusagt. Zuständig: **Architect** (Vertragstext oder Negativ mit Grund und Folge), **Implementer** (Fänger, wenn der Vertrag bleibt). | `AGENTS.md` §3.15 | `docs/plan/planning/in-progress/slice-odata-codec-csdl.md` · „| A4 | Ergebnis unabhängig von der Zerlegung in Reads“; `outboundtest/metadata_codec_vertrag.go` · `func zerlegungen(` | ja | akzeptiertes Negativ ohne Grund und Folge |
| F-3 | MEDIUM | Fänger der Zeichenmengen von Nr. 3 führt je Grenze nur eine Seite: Mutation „U+FFFF zugelassen“, „0x1F“, „0x0B“, „U+10000 gesperrt“, „U+10FFFF gesperrt“, „DEL gesperrt“, „C1-Bereich gesperrt“ und „CR gesperrt“ bleiben alle grün (je 1 von 1 Prozess, deterministisch). Fixtures führen nur NUL, 0x01, U+FFFE als ungültig und kein CR, DEL, U+10FFFF als gültig. Zuständig: **Implementer**. | `AGENTS.md` §3.15 | `outboundtest/metadata_codec_lesen.go` · `{"Surrogat", "\xED\xA0\x80"}, {"U+FFFE", "\xEF\xBF\xBE"}, {"NUL", "\x00"}` · Code `csdl/strom.go` · `r == 0xFFFE \|\| r == 0xFFFF` | ja | Fänger führt nur eine Form der Eingabe |
| F-4 | MEDIUM | Leerraum nach der Wurzel (Nr. 4: Leerraum, Kommentare, PI): `bytes.Trim` ohne `\r` und `bytes.TrimSpace` bleiben grün; kein Fixture mit CR, keines mit NBSP oder U+0085 nach der Wurzel. Zuständig: **Implementer**. | `AGENTS.md` §3.15 | `csdl/lesen.go` · `bytes.Trim(t, " \t\r\n")` · Fixture `{"Leerraum um die Wurzel", …"\n\t"…}` | ja | Fänger führt nur eine Form der Eingabe |
| F-5 | MEDIUM | Ersatzreferenz (Nr. 4 b): Mutation „nur die erste Referenz je Token prüfen“ und „nur das letzte Attribut prüfen“ bleiben grün (je 1 von 1). Jede Fixture hat genau eine Referenz je Token und das Attribut `Term` allein. Probe: `a&#65;&#xD800;` und `Term="a&#xD800;" Qualifier="q"` sind im Original `syntax`, kein Test hält es. Zuständig: **Implementer**. | `AGENTS.md` §3.15 | `csdl/lesen.go` · `func hatErsatzRef(roh []byte) bool`, `ersetzt = ersetzt \|\| strings.Contains(a.Value, "�")` | ja | Fänger führt nur eine Form der Eingabe |
| F-6 | MEDIUM | Doppeltes Attribut (Nr. 4): Mutation „nur Attribute ohne Namensraum“ und Mutation „Schlüssel ist der lokale Name“ bleiben grün. Kein Fixture hat ein doppeltes Attribut mit Namensraum (gleicher Name über zwei Präfixe derselben URI: im Original `syntax`) und keines hat zwei Attribute gleichen lokalen Namens in verschiedenen Namensräumen (muss gültig sein). Zuständig: **Implementer**. | `AGENTS.md` §3.15 | `csdl/lesen.go` · `func doppelt(attr []xml.Attr) bool` | ja | Fänger führt nur eine Form der Eingabe |
| F-7 | MEDIUM | Zeile G1, Allaussage „Attribut mit Namensraum an einem ausgewerteten oder gemeldeten Element“: das Fänger-Raster führt 7 von 14 Positionen. Neun Mutationen grün: Meldung entfällt an `Key`, `NavigationPropertyBinding`, `Edmx`, `DataServices`, `Annotations`, `Annotation`, an gemeldeten Elementen (`konstrukt`), an fremden Elementen, und die Ausnahme für `xml:` (z. B. `xml:lang`) entfällt. Ort eines fremden Elements mit `Name` (`ortKind` gegen `p.ort`) ebenso grün. Abweichung 2 des Implementers hat damit keinen Fänger. Zuständig: **Implementer** (Fänger), **Architect** (Abweichung 2, F-1 b). | `AGENTS.md` §3.15 | `csdl/elemente.go` · `case a.Name.Space != "xmlns" && a.Name.Space != nsXML:`, `func (l *lauf) attributFunde(` | ja | Allaussage ohne Fänger über alle Fälle |
| F-8 | MEDIUM | Zusatztabelle (Paare von Verstößen) lässt Paare des Ablaufs von Nr. 9 aus: Vertauschen von `summe`↔`vererbung`, `vererbung`↔`typenAufloesen`, `vererbung`↔`setsAufloesen`, `typenAufloesen`↔`setsAufloesen`, `setsAufloesen`↔`schluessel` bleibt grün (5 von 28 Paaren, je 1 von 1). Die Rangfolge von Nr. 9 c gegen d, e, b sowie e gegen d und f ist behauptet, nicht gefangen. Zuständig: **Architect** (Zeilen in der Zusatztabelle), **Implementer** (Fall). | `AGENTS.md` §3.16 | `slice-odata-codec-csdl.md` · „**Zusatztabelle: Paare von Verstößen am selben Dokument**“; Code `csdl/aufloesen.go` · `d.basen, d.ketten, d.summe, d.vererbung` | ja | Paar von Ursachen ohne Fänger |
| F-9 | MEDIUM | Fixture gleicht dem Wert der Mutation oder führt nur eine Form: (a) D6 „zwei Key“ verwendet dieselbe `PropertyRef` `Id` in beiden, die Prüfung „doppelte PropertyRef“ meldet denselben Ort; Mutation `keys > 2` bleibt grün. (b) D6 „zweites EntityContainer“ nur in einem zweiten Schema; Mutation „nur in anderem Namensraum“ grün. (c) D1: abstrakte Basis ohne Schlüssel mit nicht abstrakter Ableitung ohne Schlüssel muss `ohne_schluessel` sein; Mutation „Abstract irgendwo in der Kette befreit“ grün. (d) D7: leerer `Key` in der Ableitung bei Schlüssel in der Basis (Mutation `t.keys > 0` statt `len(refs)`), Schlüssel in der Mitte bei Wurzel ohne Schlüssel (Mutation „nur die Wurzel zählt“), geerbte Property gegen eigene Navigation-Property gleichen Namens: alle grün. (e) C1: anderer Name als `Edmx` im V3-Namensraum muss `wurzel` sein; Mutation „jeder Name im V3-Namensraum ist `version`“ grün. Zuständig: **Implementer**. | `AGENTS.md` §3.15 | `outboundtest/metadata_codec_aufloesen.go` · `schlecht("zwei Key", eins(`<EntityType Name="T"><Key><PropertyRef Name="Id"/></Key><Key><PropertyRef Name="Id"/>` | ja | Fixture gleicht dem Wert der Mutation |
| F-10 | MEDIUM | Zeile D6, Namen (Nr. 8 b): Mutationen `unicode.IsDigit`→`IsNumber`, →`'0'..'9'` und `IsLetter`→`IsLetter\|\|IsMark` bleiben grün (die Fixtures kennen weder `a²`, noch eine Nd-Ziffer außerhalb ASCII, noch ein Zeichen der Kategorie M). `Edm` als reservierter Name: `HasPrefix`, `EqualFold` bleiben grün (kein Fixture `Edmonton`, `edm`, `EdmX`). Die Länge des Namensraums in Bytes statt Zeichen bleibt grün (die Prüfgrenzen-Tabelle sagt „in Zeichen (`ä`), nicht in Bytes“, das Fixture ist ASCII). Zuständig: **Implementer**. | `AGENTS.md` §3.15 | `csdl/namen.go` · `unicode.IsDigit(r)`, `utf8.RuneCountInString(s) > maxNamensraum`; `csdl/elemente.go` · `if ns == "Edm" {` | ja | Fänger führt nur eine Form der Eingabe |
| F-11 | MEDIUM | Zeile A2: Kontextfehler × Stelle. Mutation „im Lesen wird jeder Kontextfehler zu `context.Canceled`“ bleibt grün (1 von 1): `DeadlineExceeded` kommt nur vor dem Aufruf, nie mitten im Lesen oder vor dem Auflösen vor. Der Vertrag sagt „der Fehler des Kontexts, unverändert (`==`)“ an jeder Aufrufstelle. Zuständig: **Implementer**. | `AGENTS.md` §3.15 | `outboundtest/metadata_codec_vertrag.go` · `func zeileA2(` · `csdl/strom.go` · `s.ende = &stromFehler{art: stromKontext, err: err}` | ja | Fänger führt nur eine Form der Eingabe |
| F-12 | MEDIUM | Zeile F5: Der Konstruktor kopiert vier Eingaben; der Fänger ändert nur `Keys`, `Properties` und `Name` der Typen. Mutationen „`namespaces`, `sets` oder `bindings` nicht kopiert“ bleiben grün (3 von 3). Zuständig: **Implementer**. | `AGENTS.md` §3.15 | `src/hexagon/model/service_model_test.go` · `{"Eingabe des Konstruktors"`; `service_model.go` · `version: version, namespaces: kopie(namespaces), sets: kopie(sets)` | ja | Fänger führt nur eine Form der Eingabe |
| F-13 | MEDIUM | Randform im Code entschieden: Ein Zyklus über mehr als 64 Typen ist `kette` (Probe: 65 und 70 Typen im Ring, Ort `Ns.T1`), ein Zyklus über 64 oder weniger Typen ist `nicht_aufloesbar`. `SPEC-125` Nr. 9 a nennt „Zyklus ist `nicht_aufloesbar`“ und „65. ist `kette`“ ohne die Rangfolge der beiden bei einem Ring; das Raster hat dafür keine Zelle, D3 keinen Fall. Zuständig: **Architect**. | `AGENTS.md` §3.12, §3.16 | `csdl/aufloesen.go` · `case slices.Contains(gesehen, c):` vor `case len(gesehen) >= maxKette:` | ja | Randform im Code entschieden |
| F-14 | MEDIUM | Randform im Code entschieden: `Namespace="Edm.Sub"` wird gelesen (Probe: gültig), jede Referenz auf einen Typ darin (`Edm.Sub.T`) ist `nicht_aufloesbar`, weil `Edm.` allein den Typen von Nr. 10 gehört. `SPEC-125` Nr. 8 b reserviert nur `Edm`. Zuständig: **Architect**. | `AGENTS.md` §3.12 | `csdl/aufloesen.go` · `if strings.HasPrefix(s, "Edm.") {` in `einfacherTyp` | ja | Randform im Code entschieden |
| F-15 | LOW | Der Parametername heißt im Port und im Leser `grenze`, `SPEC-125` zeigt in Signatur und Text `max`; der Plan trägt die Abweichung nur im Ausgang des Implementers (§3), nicht in Kopf, §1 oder §6 (`AGENTS.md` §3.9). Zuständig: **Architect**. | `AGENTS.md` §3.9 | `src/hexagon/ports/outbound/metadata_codec.go` · `ParseMetadata(ctx context.Context, src io.Reader, grenze int64)` | nein | Plan folgt nicht der Korrektur |
| F-16 | LOW | §6 trägt für die acht Abweichungen des Implementers keinen Punkt mit Ausgang; sie stehen nur im Absatz „Abweichungen“ von §3. Das Risiko-Gerüst von §6 („jedes Risiko trägt einen Ausgang“) sieht sie nicht. Zuständig: **Planner** (Plan), **Architect**. | `AGENTS.md` §3.9 | `slice-odata-codec-csdl.md` · „*Abweichungen von Vertrag und Tabelle, zu entscheiden vom Architect*“ | nein | Plan folgt nicht der Korrektur |
| I-1 | INFO | Zeile A1: Die Reihenfolge der Prüfungen (a) `ctx`, (b) `src`, (c) `max` ist nicht beobachtbar, alle drei sind `model.ErrUngueltig` ohne Code; die Mutationen „b vor c“ und „a vor b“ bleiben grün. Beobachtbar ist allein „Argument vor Kontext“ (rot). Zuständig: Architect (Nr. 1 sagt mehr zu, als ein Test unterscheiden kann). | `AGENTS.md` §3.15 | `csdl/leser.go` · `case ctx == nil:` | nein | Zusage ohne beobachtbare Wirkung |
| I-2 | INFO | Äquivalente Mutanten (kein Befund): Wegfall der Prüfung `ctx.Err()` vor dem ersten Read in `ParseMetadata` (`strom.Read` prüft vor jedem Read, auch dem ersten: der Aufruf ist redundant), `AutoClose` am Decoder (nur im nicht strikten Modus wirksam), `&#X41;` und `Contains` statt `HasPrefix` für `CDATA` (die Bibliothek bzw. die Token-Grenzen decken es). | Maintainability | `csdl/leser.go` · `if err := ctx.Err(); err != nil {` (erste Prüfung) | nein | — |
| I-3 | INFO | Größe: 3628 Zeilen gegen Marke 2400 (+51 %); Tests 2125, Produktcode 1503 (Zahlen des Implementers, Summe nachgemessen). Die Rückführung `in-progress` → `next` (§4) ist die Entscheidung des Planners; der Implementer hat nicht gekürzt und berichtet mit Grund (kein Maßstab in der Formel für den Code eines Parsers). | `AGENTS.md` §3.14 | `slice-odata-codec-csdl.md` · „*Größe* (`AGENTS.md` §3.14): gemessen mit“ | ja | Marke überschritten, berichtet |
| I-4 | INFO | `strom.alles` hält jedes gelieferte Byte des Dokuments für die Suche nach Ersatzreferenzen; der Leser hält so bis zu `max` Bytes zusätzlich zum Modell. `SPEC-125` Nr. 5 nennt als Folge nur „höchstens `max + 1` Bytes aus `src`“. Kein Fänger über die Speichergröße. | `AGENTS.md` §3.15 | `csdl/strom.go` · `s.alles = append(s.alles, p[:v]...)` | nein | Folge im Betrieb unvollständig |
| I-5 | INFO | `TestTransaktionVertrag` (Paket `sqlite`, nicht im Diff): in 80 Prozessen nicht rot geworden (40 bei `--cpus=2`, 40 bei `--cpus=0,15`); das einmalige Rot unter Last ist nicht bestätigt, kein Flake nachgewiesen. Der Wortlaut „cannot commit - no transaction is active“ ist ein Fehler der Bibliothek bei Transaktionsende; Zuständigkeit offen (Eintrag im Beobachtungs-Register, wenn es wiederkehrt). | `AGENTS.md` §3.10 | `src/adapters/driven/sqlite/repositories_test.go` · `func TestTransaktionVertrag` | ja | Rate nicht nachgewiesen |
| I-6 | INFO | Zeilen I1 bis F-Zeilen: Die Reihenfolge von `Namespaces`, `NavigationBindings` und `Keys` fängt Zeile I1 nicht (Dokument mit einem Namensraum, ohne Bindings, ohne Composite-Schlüssel). Die Zeilen F4 (28 von 30), F6 (6 von 30), F3 (17 von 30), B5 (23 von 30) fangen sie nur mit Wahrscheinlichkeit je Prozess; als Ganzes ist die Suite in 30 von 30 Prozessen rot (I2, I3, A4 vergleichen Läufe untereinander). Kein Befund, weil der Aggregat-Fänger deterministisch ist; die Zeile in der Tabelle benennt ihn nicht. | `AGENTS.md` §3.10 | `outboundtest/metadata_codec_funde.go` · `func zeileI1(` | ja | Fänger nur mit Wahrscheinlichkeit, Aggregat sicher |
| I-7 | INFO | J1: Die Regel `encoding/xml` → `src/adapters/driven/odata` wirkt: Import in `src/hexagon/model` → `core-impurity`, in `ports/outbound` → `port-impurity`, in `services/session` → `app-impurity`, im SQLite-Adapter → `tech-leak`, im Adapter `csdl` → 0 Befund; ohne die Regel (Kopie) bleibt der Import im Kern ungemeldet („0 Befund“), die Regel ist also tragend. | `AGENTS.md` §3.6 | `.a-check.yml` · `{pattern: "encoding/xml", adapter: "src/adapters/driven/odata"}` | ja | — |

## Raten (Zahl der Prozesse, Messender)

Alle Reihen: Reviewer, Host mit 20 Kernen, Testbinary der Wegwerfkopie, ein Prozess je Lauf mit `-test.count=1` (`TestVertrag`).

| Zeile | Mutation | Bedingung | Prozesse rot | Unmutiert grün |
|---|---|---|---|---|
| I1 | `EntityTypes` über Map | ohne Detektor, `--cpus=3` | 30 von 30 (`I1` rot in 30) | nicht gemessen (Vorlauf ohne Mutation: Suite grün) |
| I1 | `Properties` über Map | wie oben | 30 von 30 | — |
| I1 | `Funde` über Map | wie oben | 30 von 30 | — |
| I1 | `EntitySets` über Map | wie oben | 30 von 30 (`I1` in 30) | — |
| I1 | `Namespaces` über Map | wie oben | 30 von 30, `I1` selbst nie, `F4` in 28 | — |
| I1 | `NavigationBindings` über Map | wie oben | 30 von 30, `I1` nie, `F6` in 6, `B5` in 23 | — |
| I1 | `Keys` über Map | wie oben | 30 von 30, `I1` nie, `F3` in 17 | — |
| I2 | Variable des Pakets (`zaehler++`) | Detektor, `--cpus=3` | 15 von 15 | 15 von 15 |
| I2 | Map des Pakets (`cache[grenze]++`) | Detektor, `--cpus=3` | 15 von 15 (`fatal: concurrent map writes` in 13, Detektor in 2) | 15 von 15 |
| I2 | Variable des Pakets | Detektor, `--cpus=0,3` | 3 von 3 (klein: „nie grün gesehen“) | — |
| I3 | Liste der Funde als Variable des Pakets unter Sperre | ohne Detektor, `--cpus=3` | 30 von 30 (`I3` selbst rot in 30) | — |
| I3 | wie oben | ohne Detektor, `--cpus=0,2` | 10 von 10 (`I3` rot in 10) | — |

Abdeckung unter dem Detektor (`go test -race -coverprofile`, `-run 'TestVertrag/(I2|I3)'`): 99,7 % der Statements von `csdl`, Rest wie vom Implementer genannt (`io.ErrNoProgress` in `tokenFehler`, `stromFehler.Error`); nachgemessen, deckt sich.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model` (`metadaten.go`, `service_model.go`, Tests) | geprüft; Befund F-12 (Konstruktor), sonst ohne Befund: Form des Fehlertexts, `Unwrap`, `Code`, Lese-Sicht je Zugriff gefangen |
| `src/hexagon/ports/outbound` (`metadata_codec.go`, `abbildung.go`, `abbildung_test.go`) | geprüft, ohne Befund (H2: Wegfall und Beschränkung der Zeile rot) |
| `src/hexagon/ports/outbound/outboundtest` | geprüft; Befunde F-3 bis F-11 betreffen Fänger dieser Suite |
| `src/adapters/driven/odata/csdl` | geprüft; Randformen im Code F-1, F-13, F-14; Sicherheit des Lesers ohne Befund (Probe, Grenzen, Entitäten) |
| `.a-check.yml` | geprüft, ohne Befund (I-7) |
| `src/hexagon/services/session/beenden_test.go` | geprüft, ohne Befund (eine Zeile `ErrMetadatenUngueltig`) |
| Plan `slice-odata-codec-csdl` | geprüft; Befunde F-2, F-15, F-16 |
| Sendungen (`AGENTS.md` §3.13) | geprüft, ohne Befund: `slice-baseline-import` und `slice-bootstrap-m2` tragen die Sendung, die Änderung an `slice-baseline-import` §4 ist eine Zeile in der Marke |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft, ohne Befund |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 14 |
| LOW | 2 |
| INFO | 7 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe (F-3 bis F-6, F-10 bis F-12) · Fixture gleicht dem Wert der Mutation (F-9) · Randform im Code entschieden (F-1, F-13, F-14) · Allaussage ohne Fänger über alle Fälle (F-7) · Paar von Ursachen ohne Fänger (F-8) · akzeptiertes Negativ ohne Grund und Folge (F-2) · Plan folgt nicht der Korrektur (F-15, F-16)

## Verdikt

**Merge-blockierend:** ja — 14 MEDIUM. Kein HIGH: der Leser ist in den geprüften Formen korrekt (alle Probe-Dokumente entsprechen dem Vertrag), die Findings betreffen Fänger, die eine Mutation grün lassen, und Randformen, die der Code entscheidet.

**Übergabe:** An den **Implementer**: F-3, F-4, F-5, F-6, F-9, F-10, F-11, F-12 (Fänger ergänzen, jede grüne Mutation rot sehen) und mit dem Architect F-2, F-7, F-8. An den **Architect**: F-1, F-13, F-14, F-15 (Entscheidung in `SPEC-125` vor dem Fänger), F-8 (Zeilen der Zusatztabelle), F-2. An den **Planner**: F-16. Die Finding-Klassen gehen in die Slice-Closure §7. Der Report ersetzt keine Verifikation.
