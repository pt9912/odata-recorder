# Review-Report: slice-odata-codec-json — 2026-10-24

**Review-Art:** Code — geprüft gegen Plan (`slice-odata-codec-json` §1, §3 mit der Tabelle *Zusage · Fänger* A1 bis L1 samt F6 bis F8 und Z1 und dem Bericht des Implementers, §4, §6 mit Raster und Zusatztabelle der Paare), Vertrag `SPEC-126` und Hard Rules (`AGENTS.md` §3.1 bis §3.17), nicht gegen die DoD (Verifier).

**Gegenstand:** `slice-odata-codec-json` (Welle `welle-record-replay`), `git diff bc05c62..0acc49a` mit HEAD `0acc49a` (Implementer `534a557` = alles außer Teil Z, `0acc49a` = Teil Z: `Edm.Guid`, `Edm.Date`, `Edm.DateTimeOffset`). Pfade: `src/hexagon/model`, `src/hexagon/ports/outbound` (mit `outboundtest/`), `src/adapters/driven/odata/jsoncodec`, `src/bootstrap`, eine Zeile in `src/hexagon/services/session/beenden_test.go`, der Plan.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `0acc49a` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-24

**Methode.** Der Bericht des Implementers (Plan §3) wurde gelesen, keine Mutation und keine Rate übernommen. Eigene Werkzeuge im Scratchpad (nichts im Repo außer diesem Report): ein Mutationsläufer (`rv/runner/rvmut.go`, Text-Datei je Mutation mit Datei, alt, neu; Wegwerfkopie von `git archive`-Stand, Quelle `:ro`, `--network none`, `--memory=4g`, Mutationen ohne Grenze mit `--memory=2g` und Zeitgrenze, höchstens drei Container zugleich) und ein **Probe-Paket aus Generatoren** (`rv/probe/`, 5 Dateien, 23 prüfende Tests, rund 8,8 Millionen Läufe, Laufzeit 75 Sekunden). Die Mutationen sind 237 Definitionen (davon 15 Bäume für die Raten); die Zeilen der Tabelle wurden mit der ganzen Suite `ODataCodecVertrag` (über einen Wrapper im Paket `revtest`, gleich `src/bootstrap/odata_codec_test.go`) und mit den Zeilen-Filtern gefahren. Vor dem Report: `make gates` Rückgabecode 0, Repo unverändert (die uncommittete Änderung an `.claude/agents/architect.md` blieb ungelesen und unberührt).

**Eingangs-Kontext:**

- Slice-Plan `slice-odata-codec-json` in `0acc49a` (§1, §3, §4, §6, §8)
- `SPEC-126` (Nr. 1 bis 15, *Ungeprüft*), `SPEC-125` Nr. 11, `SPEC-109` Nr. 5, `SPEC-051`, `SPEC-049`
- `LH-FA-18`, `LH-RB-01`, `LH-FA-11`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `AGENTS.md` (Hard Rules §3.1 bis §3.17), Register `suite-ohne-generator-fuer-klassen-der-spezifikation`, `werkzeug-eingabe-setzt-escape-in-zeichen-um`

---

## Antworten auf die Fragen des Auftrags

**(1) Tabelle A1 bis L1, F6 bis F8, Z1; Raster; Zusatztabelle.** Jede Zeile hat einen Test unter dem genannten `t.Run`-Namen und ist von mindestens einer eigenen Mutation aus dem benannten Grund rot gesehen (Tabelle *Mutationen je Zeile* unten); 50 Zeilen rot, kein Aufbau-Rot. Der Code selbst ist gegen das Probe-Paket **ohne Abweichung** (0 von 8,8 Millionen Läufen), ein Defekt im Produktcode ist nicht gefunden. Grün bleiben **14 Mutationen der ganzen Suite**, die zu fünf Klassen gehören (F-1 bis F-4, F-6); das Probe-Paket macht 13 davon rot (die vierzehnte, `{"a":1,:2}`, braucht Token-Folgen der Länge 8). Zusatztabelle: alle 139 ausdrückbaren Paare (95 der Collection, 44 der Entity) sind durch mindestens einen Fall belegt (Instrumentierung von `hPaar`), sechs Vertauschungen benachbarter Schritte des Ablaufs von Nr. 13 c sind rot, und 120 000 zufällige Dokumente über alle 15 Stellen mit beliebigen Platzierungen, mehreren Verstößen und `grenze` stimmen mit einem eigenen Orakel überein. **H2 281 statt 292 und H1 mit zwölf Fällen `syntax`: Defekte der Tabelle des Architect, nicht des Codes** (I-1). Die Zelle „Set mit nicht auflösbarem Typ“ hat der Code entschieden (F-5).

**(2) §3.17, Probe-Paket, Register-Eintrag `suite-ohne-generator-fuer-klassen-der-spezifikation`.** Das Probe-Paket führt Generatoren über die Klassen (1) Zahlformen (Grammatik 960 799 Folgen, Werte gegen `math/big`), (2) Guid (alle Einzeländerungen), (3) Date (alle Monate × Tage an rund 65 Jahren, Orakel `time`), (4) DateTimeOffset (168 171 Kombinationen, Orakel eigene Kalenderrechnung nach Hinnant), (5) Schlüssel (alle 40 320 Permutationen von acht Typen, alle 255 Teilmengen fehlender Teile, Injektivität), (6) Kontext (353 456 Einzeländerungen, Orakel regulärer Ausdruck), (7) Paare (120 000 Zufallsdokumente), (8) Kodierung (alle 65 536 Escapes, alle 1 048 576 Paare hoch × tief, alle Ein- und Zwei-Byte-Folgen, Drei- und Vier-Byte-Folgen an Grenzbytes), (9) Grammatik (45 302 Ein-Byte-Änderungen gegen `json.Valid`, 5 380 839 Token-Folgen bis Länge 7, 320 Escape-Bytes und Literale), (10) Pointer (6 230 Namen), (11) Zerlegung (100 182 Läufe mit 25 zufälligen Zerlegungen je Body, Bodys bis 456 KB), (12) Grenzen (Größe, Dezimallänge 1016 bis 1032 in 16 Formen). **Bewertung:** Die Suite müsste für die Klassen (8), (9) und die Zahlgrammatik, die Dezimallänge und den Kontext einen Generator führen: 14 überlebende Mutationen (F-1 bis F-4, F-6) liegen alle in Klassen, die die Suite nur über handgeschriebene Mengen fährt, und jedes dieser Probe-Pakete ist klein (4 bis 100 Zeilen je Klasse, 0,1 bis 12 Sekunden). Der Architect hat die Folge benannt („erst das Probe-Paket“); sie tritt ein: jede Runde findet in diesen Klassen neue Mutationen, und `AGENTS.md` §3.16 verlangt für eine Allaussage des Vertrags („Jeder Verstoß ist `syntax`“) den Fänger über alle Fälle oder die Prüfgrenze als akzeptiertes Negativ. Das ist ein weiterer Beleg für den Register-Eintrag (bisher 1×, offen).

**(3) Raten, akzeptierte Negative, Detektor.** Siehe Tabelle *Raten*; die Zahlen des Implementers (I1 30 von 30, J1 30 von 30, J2 20 von 20 und 100 von 100) sind reproduziert. Akzeptierte Negative: *L1 quadratisch* bestätigt (Kopie je Anhängen endet nach 100 s an der Zeitgrenze, Zeile L1; Grund und Folge stehen im Plan); *Set mit unbekanntem Typ ohne Test* trägt eine falsche Begründung (F-5); *teilweises BOM* steht weder im Plan noch in `SPEC-126` *Ungeprüft* und bleibt als Mutation grün (F-3); *äquivalente Mutation `Properties` nil* bestätigt (grün, weil jede gelesene Entity ihren Schlüssel als Property trägt). Abdeckung unter dem Detektor (`go test -race -coverpkg`): 99,9 %, ein Block ohne Lauf, `wert.go` · das `return model.Wert{}, model.DatenTyp` hinter dem `switch` von `lesWert` (nicht erreichbar, `Unterstuetzt()` lässt nur die acht Typen durch); gleich dem Bericht. Die Suite braucht unter dem Detektor mit Abdeckung 234 Sekunden.

**(4) Hard Rules.** Kein `encoding/json`, kein `reflect`, keine Fremdbibliothek im Leser (`jsoncodec` importiert `context`, `errors`, `io`, `strconv`, `strings`, `time`, `unicode/utf8`, `model`); keine Variable des Pakets (Leser ohne Zustand, 9 Mutationen des Zustands in 30 Prozessen rot); `make gates` grün, `a-check` ohne Befund; keine `//nolint`. Kommentare (§3.7, §3.11): keine verworfene Alternative, kein abwesender Text; die Köpfe nennen, was die Suite fängt (Kopf von `dezimalLaenge` sagt „vier Formen“ und führt vier, der Zweig mit Ziffern beiderseits des Punktes fehlt: F-4). §3.9: Plan folgt dem Code bis auf F-8. **Gegenlesen der Suite-Bytes** (`werkzeug-eingabe-setzt-escape-in-zeichen-um`): kein Steuerzeichen, kein NBSP, kein U+2028, kein BOM im Quelltext; außer den deutschen Meldungen sind die Zeichen jenseits ASCII absichtlich roh (`é` als `C3 A9`, `e` + U+0301 als `65 CC 81` in der Tabelle der Pointer, `ä`; Bytes gelesen); alle `\u`-Folgen stehen als `esc("…")` oder in Raw-Strings. **Beleg für das Register:** bei diesem Lauf hat das Werkzeug dreimal eine eigene Eingabe verändert (zuletzt beim Schreiben dieses Satzes) (der Escape für die öffnende eckige Klammer, geschrieben als Backslash, `u` und `005b`, wurde in einem Bash-Heredoc zu `[`; die Escapes für U+00A0, U+0660 und U+FF11 wurden in einem Write-Inhalt zu rohen Zeichen); das Probe-Paket prüfte zunächst das Falsche und fiel durch die Abweichung von Suite und Probe auf. Der Eintrag steht damit bei 3×.

**(5) Größe.** `git diff --numstat bc05c62..0acc49a` über die Pfade der Marke: **5612** Zeilen (Produktcode 1640, Tests und Suite 3972), vor Teil Z (`534a557`) **5144** (1447 und 3697), Teil Z allein **468**; gleich den Zahlen des Implementers. Marke 4500: +24,7 % mit Z, +14,3 % ohne Z. Grund: nicht der Produktcode (1640 gegen erwartete 1500, Spanne 1250 bis 1800), sondern die Suite mit 81 Zeilen je Zeile der Tabelle (3972 für die 49 Zeilen der Tabelle) gegen 64 beim Leser des CSDL und 40 der Formel; 5612 gegen die Formel allein (2050) ist der Faktor 2,74, gegen die Schätzung mit Produktcode (3550) 1,58. Die größten Dateien: `odata_codec_paare.go` 658 Zeilen (551 Fälle der Zusatztabelle), `odata_codec_allaussagen.go` 500, `odata_codec_lesen.go` 437, `odata_codec_strom.go` 382, `odata_codec_typen.go` 380. **Z als Schnitt:** der Stand `534a557` übersetzt und besteht die ganze Suite (26 Sekunden, 46 Zeilen: ohne F6 bis F8 und Z1); Z teilt mit dem Rest `hexWert` und `ziffer` aus `scan.go` und die Zuordnung in `lesWert`. Z ist 8,3 % des Slice; ohne Z liegt der Slice weiter 644 Zeilen über der Marke, der Schnitt löst die Überschreitung nicht. Die Entscheidung trifft der Planner; der Befund steht als INFO (I-4).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Allaussage „jeder Verstoß der Grammatik ist `syntax`“ (Nr. 4) hat keinen Fänger über die Klasse: acht Mutationen bleiben in der **ganzen Suite** grün — die Ziffer nimmt `_` an; eine Zahl nimmt `1e5e5` und `1.5.5` an; `[1}` und `{"a":1]` (Klammerart) werden gelesen; `{"a" "b":1}` und `{"a":1,:2}` werden gelesen; `tRUE` gilt als Literal; der Escape `\'` gilt. Zeile B2 führt die Bytes an sechs Orten und eine Liste von rund 60 Verstößen, keine Zahl-, Literal-, Klammer- oder Escape-Position. Das Probe-Paket (Zahlen 960 799 Folgen, Token-Folgen 5 380 839, Ein-Byte-Änderungen 45 302, Escapes und Literale 320) macht sieben der acht rot; `{"a":1,:2}` fängt es erst bei Länge 8 (Prüfgrenze des Pakets). Der Code selbst weicht in keinem Lauf ab. Zuständig: **Implementer** (Fänger), **Architect** (Allaussage ohne Prüfgrenze im Plan, Klasse (9) bewusst der Probe überlassen). | `AGENTS.md` §3.15, §3.16, §3.17 | `jsoncodec/scan.go` · `case (c == 'e' \|\| c == 'E') && s.z != zExp:` · `outboundtest/odata_codec_lesen.go` · `func jfB2Verstoesse() []jfall` | ja — Mutationen `ziffer`, `zahlFortsetzung`, `nachWert`, `objektByte`, `literalByte`, `escapeByte` | Allaussage des Vertrags ohne Fänger über die Klasse |
| F-2 | MEDIUM | Zeile B3: die gültigen Zeichen führen keinen Lead-Byte-Bereich von `E1` bis `EC` und von `F1` bis `F3` (kein Zeichen von U+1000 bis U+CFFF, keines der Ebenen 4 bis 15). Die Mutationen `c <= 0xF2`, `c >= 0xF2` und `c >= 0xE2` im Lauf der Folge bleiben in der ganzen Suite grün (3 von 3); der ganze Zweig `0xF1..0xF3` hat keinen gültigen Fall. Probe: alle Ein- und Zwei-Byte-Folgen und Drei- und Vier-Byte-Folgen an Grenzbytes rot. Wiederholung des Musters von F-3 im Review von `slice-odata-codec-csdl` (je Grenze nur eine Seite). Zuständig: **Implementer** (ein Fall je Bereich), **Architect** (Menge „gültig roh“ aus Grenzen gebildet, nicht aus den Zweigen). | `AGENTS.md` §3.15 | `jsoncodec/scan.go` · `case c >= 0xF1 && c <= 0xF3:` · `outboundtest/odata_codec_lesen.go` · `func b3Gueltig() []string` | ja | Fänger führt nur eine Form der Eingabe |
| F-3 | MEDIUM | Ein teilweises BOM (`EF BB` mit beliebigem dritten Byte statt `BF`) hat keinen Fänger: die Mutation „`EF BB` gefolgt von jedem Byte ist `kodierung`“ bleibt in der ganzen Suite grün; B3 führt nur die drei vollständigen BOM, B2 nur Einzelbytes vor der Wurzel (`EF`, `FF`, `FE` allein sind gefangen). `SPEC-126` Nr. 3 macht allein das vollständige BOM zu `kodierung`. Das Negativ steht weder im Plan (§3, §6) noch in *Ungeprüft*, es hat weder Grund noch Folge. Probe über alle 6 216 Folgen von 1 bis 4 Bytes aus `{EF BB BF FF FE 20}`, ganz und Byte für Byte: rot. Zuständig: **Implementer** (Fall), **Architect** (Zelle der Raster-Zeile „nur BOM“). | `AGENTS.md` §3.15 | `jsoncodec/scan.go` · `case s.bom == 0xEF && s.bomN == 2 && c == 0xBF` · `outboundtest/odata_codec_lesen.go` · `for _, bom := range []string{` | ja | Akzeptiertes Negativ ohne Grund und Folge |
| F-4 | MEDIUM | Zeile F5, Länge 1024: der Kopf nennt vier Formen (ganzer Teil, Bruchteil, Exponent, Vorzeichen), der Code hat drei Zweige der Länge. Der Zweig mit Ziffern beiderseits des Dezimalpunkts (`123.456`, `12345e-3`) hat keinen Grenzfall: die Mutation `laenge = n + 2` bleibt in der ganzen Suite grün (1 von 1), `12345e-1022` liegt im Zweig des reinen Bruchteils. Probe über 16 Längen von 1016 bis 1032 in 16 Formen (ganz, Bruch, gemischt, Exponent, mit Vorzeichen, als String): rot. Zuständig: **Implementer** (Fall), **Architect** (die Menge der Tabelle nennt „gemischt“ nicht). | `AGENTS.md` §3.15, §3.17 | `jsoncodec/zahl.go` · `laenge = n + 1` · `outboundtest/odata_codec_werte.go` · `func dezimalLaenge() []wertFall` | ja | Fänger führt nur eine Form der Eingabe |
| F-5 | MEDIUM | Randform im Code entschieden: Ein Entity-Set, dessen Typ das Modell nicht kennt, liefert der Leser als `entity_set` (Ort der Name). `SPEC-126` Nr. 1 f nennt „nennt kein Entity-Set des Modells“, `SPEC-125` Nr. 11 schließt den Fall aus; das Raster hat keine Zelle. Der Bericht (Abweichung 3) begründet das fehlende Testen mit „die Fixture nicht zu bauen“ und setzt die Zelle auf „gilt nicht“ — das ist eine Entscheidung des Architect, und die Fixture ist zu bauen: `model.NeuesServiceModel("4.0", nil, nil, []model.EntitySet{{Name: "X", EntityType: "T.Fehlt"}}, nil)` (Probe: beide Methoden `entity_set`, Ort `X`). Zuständig: **Architect** (Zelle und Entscheidung: `entity_set` oder `model.ErrUngueltig`), **Implementer** (Fall). | `AGENTS.md` §3.12, §3.16 | `jsoncodec/deuten.go` · `d.setTyp, ok = d.info(s.EntityType)` · Plan §3 · „weil die Fixture nicht zu bauen ist“ | ja | Randform im Code entschieden |
| F-6 | LOW | Zeile K1: der Pointer führt Steuerzeichen, `/` und `~`, aber keines der übrigen Zeichen (0x20 bis 0x7E außer `/`, `~`; ab 0x80); die Mutation „`\` wird `~2`“ im `Replacer` bleibt in der ganzen Suite grün. `SPEC-126` Nr. 14 sagt zu, dass nur `~` und `/` ersetzt werden. Probe (6 230 Namen aus allen ASCII-Zeichen, Unicode und Zufallsfolgen, `Error()` gegen `strconv.Quote`): 0 Abweichungen am Code, die Mutation rot. Zuständig: **Implementer**. | `AGENTS.md` §3.15 | `jsoncodec/deuten.go` · `ersatz := strings.NewReplacer("~", "~0", "/", "~1")` · `outboundtest/odata_codec_allaussagen.go` · `func jfK1() []jfall` | ja | Fänger führt nur eine Form der Eingabe |
| F-7 | LOW | Der Plan (§3 *Speicherwächter*) nennt „ein endloser Strom aus Leerraum (Zeile A5/B5)“ als Fall, den der Wächter mit benanntem Grund meldet. Die Mutation „Größengrenze entfernt“ (`ueber := false`) füllt den Heap nicht (der Scanner verwirft Leerraum): der Lauf hängt, die Zeile A5 wird erst an der Zeitgrenze rot (60 Sekunden im Lauf mit Grenze, `go test` ohne Angabe 10 Minuten), nicht aus dem benannten Grund. Der Bericht führt die Mutation nicht. Die Wächter für Dezimalzahl, Ganzzahl und `@odata.count` (Heap 956 MiB) feuern wie berichtet, die Tiefe endet im Fall. Zuständig: **Implementer**. | `AGENTS.md` §3.10 | `jsoncodec/leser.go` · `ueber := int64(n) > rest` · Plan §3 · „ein endloser Strom aus Leerraum (Zeile A5/B5)“ | ja | Fänger wird nicht aus dem benannten Grund rot |
| F-8 | LOW | Der Plan folgt dem Code nicht in zwei Punkten (`AGENTS.md` §3.9): (a) die Zahlen von *Gemessen* stammen nicht vom Stand `0acc49a`: 18 600 Bodys und 3 675 160 Läufe (HEAD: 18 645 und 3 683 056), 12 345 Fehlerfälle (HEAD: 12 390), Spitze 84 MiB (gemessen 94 MiB, schwankt); (b) die Zeilen H1 und H2 der Tabelle nennen weiter 270 und 292 Fälle und „Schleife über die Zeilen“, gebaut sind 270 und 281 (Abweichungen 1 und 2 stehen nur im Absatz des Berichts, ohne Punkt in §6). Zuständig: **Implementer** (a), **Planner** (b, nach der Entscheidung des Architect, I-1). | `AGENTS.md` §3.9 | Plan §3 · „**Gemessen:** Zeile A4: 18 600 Bodys bis 4096 Bytes“ | ja — `go test -v -run 'TestODataCodecVertrag/^(A4\|H3)_'` | Plan folgt nicht der Korrektur |
| I-1 | INFO | Defekte der Tabelle des Architect, nicht des Codes: (a) elf der 292 Paare von H2 sind in keinem Body ausdrückbar (S1 mit S2 am selben Mitglied, in Collection und Entity; S6 mit S7 bis S15, weil ohne `value` kein Element existiert); gebaut und belegt sind 139 Paare in 281 Fällen. (b) Von den 16 Fällen „Verstoß des Lesens hinter dem schließenden `}`“ (vier Ursachen × S2, S6, S14) sind zwölf (`kodierung`, `tiefe`, `doppelter_name`) kein eigener Fall: hinter dem Ende der Wurzel ist das erste Byte `syntax`, die Zahl 270 zählt sie mit. Beide Abweichungen hat der Implementer richtig zurückgegeben; eine Entscheidung im Code liegt nicht vor. | `AGENTS.md` §3.16 | Plan §6 · „**H2 zusammen 226 + 66 = 292, H1 270, die Tabelle 562 Fälle.**“ | nein | Tabelle des Architect zählt Unausdrückbares |
| I-2 | INFO | Äquivalente oder nur in der Laufzeit sichtbare Mutanten (kein Befund): das Feld `fertig` des Scanners (`ende` gleich ohne es), der Sonderfall `math.MaxInt64` in `lies` (der Rest wird nie zum Überlauf), die Umschaltung `namenLinear` (ohne sie 8,8 statt 0,3 Sekunden für Zeile L1, Ergebnis gleich), `strings.NewReplacer` für `d.pfad` (nur `value` und Ziffern), die Rückgabe von `seite` in `sammlung` (der Aufrufer `ParseCollection` setzt den Nullwert), `Properties: nil` (bestätigt), die Zweige `zPunkt` im Fall der Ziffer nach `.`. | Maintainability | `jsoncodec/scan.go` · `fertig  bool` | nein | — |
| I-3 | INFO | Zeile L1 hat eine Obergrenze von 256 MiB bei gemessenen 94 MiB (für 100 000 Elemente, 4,5 MiB Body): 2,5 KB Zusatz je Element (336 MiB) ist rot, 1 KB (rund +100 MiB) und 0,6 KB je Member bleiben grün. Die quadratische Kopie je Anhängen (100 Sekunden Zeitgrenze) bestätigt das akzeptierte Negativ. Der Fänger sieht Wachstum ab etwa dem Dreifachen. | `AGENTS.md` §3.15 | `outboundtest/odata_codec_speicher.go` · `const l1Grenze = 256 << 20` | ja | — |
| I-4 | INFO | Größe: 5612 Zeilen gegen Marke 4500 (+24,7 %), ohne Teil Z 5144 (+14,3 %); Z allein 468 (Schätzung des Architect rund 600). Der zweite Messpunkt für `groessenmarke-ueberschritten-ohne-rueckfuehrung`: Ende gegen Schätzung × 1,25 gleich 1,26 (4437), gegen die Schätzung 1,58. Ein Schnitt von Z löst die Überschreitung nicht; der Stand `534a557` ist grün. Die Entscheidung trifft der Planner. | `AGENTS.md` §3.14 | Plan §6 · „*Messung des Implementers beim Handoff*“ | ja | Marke überschritten, berichtet |
| I-5 | INFO | Raten (Tabelle unten) reproduziert und um fünf Zweige unter dem Detektor ergänzt: Zähler im Zweig des Kontextfehlers, des Fehlers von `src`, von `zu_gross`, von `entity_set` und in `zeitpunktText` sind in je 30 von 30 Prozessen rot (ebenso ein Zähler allein in `ParseEntity`). | `AGENTS.md` §3.10 | `outboundtest/odata_codec_allaussagen.go` · `func jJ1(` | ja | — |
| I-6 | INFO | Register: `suite-ohne-generator-fuer-klassen-der-spezifikation` bekommt mit F-1 bis F-4 und F-6 einen weiteren Beleg (fünf Klassen, 14 überlebende Mutationen, Probe-Paket als Beleg); `werkzeug-eingabe-setzt-escape-in-zeichen-um` einen dritten (Antwort 4); `groessenmarke-ueberschritten-ohne-rueckfuehrung` den zweiten Messpunkt (I-4). Die Eintragung gehört dem Lese-Schritt der Closure, nicht diesem Report. | `AGENTS.md` §3.17 | Plan §8 · „`suite-ohne-generator-fuer-klassen-der-spezifikation` (1×, offen)“ | nein | — |
| I-7 | INFO | Teil Z ist im Code richtig: Guid, Date und DateTimeOffset ohne Abweichung in 238 000 Läufen der Generatoren (Orakel regulärer Ausdruck, `time`, eigene Kalenderrechnung); 25 Mutationen der Zeilen F6 bis F8 und Z1 rot. | `AGENTS.md` §3.10 | `jsoncodec/zeit.go` · `func zeitpunktText(text string) (string, bool)` | ja | — |
| I-8 | INFO | Das Probe-Paket liegt nicht im Repo (Auftrag: nur der Report); es liegt im Scratchpad unter `rv/probe/` (`probe_util_test.go`, `probe_wert_test.go`, `probe_struktur_test.go`, `probe_paare_test.go`, `probe_typ_test.go`) und nutzt allein `jsoncodec`, `csdl`, `model` und die Standardbibliothek. Der Verifier fährt es einmal gegen den Code und einmal gegen die Mutationen der Tabelle unten; das Scratchpad ist flüchtig, die Klassen und Orakel stehen in Antwort (2). | `AGENTS.md` §3.17 | — | nein | — |

## Mutationen je Zeile

„rot“: eine Mutation der Zeile wird aus dem benannten Grund rot; Kennungen aus meinem Lauf. Alle 237 Definitionen liegen im Scratchpad (`rv/runner/m*.txt`).

| Zeile | Eigene Mutationen | Ergebnis |
|---|---|---|
| A1 | je Prüfung entfernt oder verschoben (`grenze < 0`, `grenze < 2`, Set vor Kontext, Kontext nach `lies`, `src` vor Set, `ToLower` am Set) | 6 von 6 rot |
| A2 | Prüfung nur beim ersten Read; ohne Prüfung vor dem Deuten; `errors.Join(err)`; Prüfung nach dem Read vor dem Verstoß | 4 von 4 rot |
| A3 | Fehler als `syntax`; nach vollständigem Body verschluckt; unvollständig als `syntax` | 3 von 3 rot |
| A4 | Zahl, Literal, UTF-8, Escape, hohes Surrogat, BOM am Ende eines Reads | 6 von 6 rot |
| A5 | `Close`; `grenze + 2`; Read nach `io.EOF` | 3 von 3 rot |
| B1 | leere Eingabe gültig; Ende der Zahl nicht gezählt (rot in C1); `fertig` ausgelassen | 2 rot (B1, C1), 1 äquivalent |
| B2 | NBSP als Leerraum, Komma vor `}` / `]`, `\a`, führende Null, `+1`, Ziffer `_`, `1e5e5`, `1.5.5`, `[1}`, `tRUE`, `\'`, `{"a" "b":1}`, `{"a":1,:2}` | 6 rot, 8 grün (F-1) |
| B3 | 10 Mutationen der Ränder und Surrogate, 5 des BOM, 3 der Lead-Bytes | 14 rot, 4 grün (F-2, F-3) |
| B4 | 511, nur Objekte, Grenze entfernt (2 GiB, Fall rot) | 3 von 3 rot |
| B5 | `>=`, Reihenfolge, Summe, `grenze+2`, `MaxInt64`; Grenze entfernt | 4 rot, 1 äquivalent, 1 Zeitgrenze (F-7) |
| B6 | `ToLower`, nur 2 Ebenen, Map ohne Altbestand, nie Map | 3 rot, 1 äquivalent |
| C1 bis C3 | Wurzel, `value`, Elemente, Wurzelmitglied, Methode | 8 von 8 rot |
| D1, D2 | Pfadende, `EqualFold`, `*`, Navigation, `TrimSpace`, Klammer | 6 von 6 rot |
| E1 bis E5 | Link, Marker, Anzahl, `@odata.type`, Annotationen, Mitglieder | 16 von 16 rot |
| F1 bis F5 | String, Boolean, Int32, Int64, Decimal | 15 rot, 1 grün (F-4) |
| F6 bis F8, Z1 | 5 + 7 + 12 + 1 | 25 von 25 rot |
| F9 bis F11 | `null`, JSON-Typ, Marker | 10 rot, 1 grün (Boolean aus String: F2 fängt) |
| G1 bis G3 | Literal, Reihenfolge, Verletzungen | 11 von 11 rot (G2 `>= 1` fängt G1) |
| H1, H2 | Lesen vor Deuten; 6 Vertauschungen im Ablauf | 7 von 7 rot |
| H3, H4 | kein `DatenFehler`, Ort ohne `/`, Unterstütztes in `Properties`, `Nullable`, Länge | 6 rot, 1 äquivalent |
| I1 bis I3 | vier Maps, drei geteilte oder behaltene Slices, Dublette entfernt oder gemeldet | 9 von 9 rot, Raten |
| J1, J2 | neun Zustände des Pakets, Puffer unter Sperre | 10 von 10 rot, Raten |
| K1 bis K4 | Ort, Code, `Unwrap`, Tabelle, Typ der Bibliothek, Feld | 8 rot, 2 grün (F-6, I-2) |
| L1 | Zusatz je Element oder Member, quadratisch, lineare Namensuche | 1 rot (336 MiB), 1 Zeitgrenze, 3 grün (I-3) |

## Raten

Alle Reihen: Reviewer, Testbinary der Wegwerfkopie im Container (`--memory` 2 bis 4 GiB), ein Prozess je Lauf mit `-test.count=1`, falls nicht anders genannt; Detektor `go test -race` (cgo, Image `golang:1.27.1-trixie`), nicht über `make test-race`.

| Zeile | Mutation | Bedingung | Prozesse rot | Unmutiert grün |
|---|---|---|---|---|
| J1 | Zähler des Pakets in `lies` | Detektor | 30 von 30 | 3 von 3 |
| J1 | Map des Pakets in `lies` | Detektor | 30 von 30 | — |
| J1 | gemeinsamer Puffer des Pakets | Detektor, Zeitgrenze 60 s | 12 von 12 | — |
| J1 | Zähler nur in `ParseEntity` | Detektor | 30 von 30 | — |
| J1 | Zähler im Zweig des Kontextfehlers / des Fehlers von `src` / von `zu_gross` / von `entity_set` / in `zeitpunktText` | Detektor | je 30 von 30 | — |
| J2 | Puffer des Pakets, zweimal gesperrt (kein Rennen, fremde Daten) | ohne Detektor | 20 von 20 | — |
| J2 | wie oben | Detektor | 20 von 20 | — |
| J2 | wie oben | ohne Detektor, `-test.count=20` | 100 von 100 Läufen in 5 Prozessen | — |
| I1 | `Properties` und `Uebersprungen` über Map | ohne Detektor | 30 von 30 | — |
| I1 | `Entities` über Map | ohne Detektor | 30 von 30 | — |
| I1 | nur `Uebersprungen` über Map | ohne Detektor | 30 von 30 | — |
| I1 | nur `Properties` über Map | ohne Detektor | 30 von 30 | — |
| I1 | nur `Properties` über Map | ohne Detektor, `-test.count=20` | 100 von 100 Läufen in 5 Prozessen | — |

Unmutiert: die Suite ist in jedem gefahrenen Lauf grün (Detektor 3 von 3, ohne 26 Sekunden, Detektor mit Abdeckung 1 von 1).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/odata/jsoncodec` (Produktcode, 1640 Zeilen) | geprüft; Befund nur über die Fänger (F-1 bis F-6), im Code kein Defekt: 8,8 Millionen Läufe der Generatoren ohne Abweichung |
| `src/hexagon/model` (`entity.go`, `daten_fehler.go`, Tests) | geprüft, ohne Befund (K1, K3; Zeilen 4 bis 7) |
| `src/hexagon/ports/outbound` (`odata_codec.go`, `abbildung.go`, `abbildung_test.go`) | geprüft, ohne Befund (K2: Wegfall der Zeile rot) |
| `src/hexagon/ports/outbound/outboundtest` | geprüft; Befunde F-1 bis F-4, F-6, F-7 betreffen Fänger dieser Suite |
| `src/bootstrap/odata_codec_test.go` | geprüft, ohne Befund |
| `.a-check.yml` | geprüft, ohne Befund (`make a-check` grün; der Leser importiert nur `model`) |
| `src/hexagon/services/session/beenden_test.go` | geprüft, ohne Befund (eine Zeile `ErrDatenUngueltig`) |
| Sendungen (`AGENTS.md` §3.13) | geprüft, ohne Befund: Nachprüfung des Planner im Plan §3 gelesen, `slice-baseline-import` trägt den Eintrag |
| Fehlertexte und Orte | geprüft, ohne Befund: kein Wert einer Property in `Error()` und `Ort`, nur Namen (`LH-FA-14`) |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft, ohne Befund |
| Speicher und Zeit (Betriebsregel Speicher) | geprüft: Wächter feuert bei Dezimalzahl, Ganzzahl, `@odata.count` (Heap 956 MiB); F-7 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 5 |
| LOW | 3 |
| INFO | 8 |

**Finding-Klassen dieses Laufs:** Allaussage des Vertrags ohne Fänger über die Klasse (F-1) · Fänger führt nur eine Form der Eingabe (F-2, F-4, F-6) · Akzeptiertes Negativ ohne Grund und Folge (F-3) · Randform im Code entschieden (F-5) · Fänger wird nicht aus dem benannten Grund rot (F-7) · Plan folgt nicht der Korrektur (F-8)

## Verdikt

**Merge-blockierend:** ja — 5 MEDIUM. Kein HIGH: der Leser ist in allen geprüften Formen vertragsgemäß (0 Abweichungen in 8,8 Millionen Läufen, gegen unabhängige Orakel), die Findings betreffen Fänger, die eine Mutation grün lassen, und eine Randform, die der Code entscheidet.

**Übergabe:** An den **Implementer**: F-1 (Fänger über die Klasse als Schleife oder Generator, Prüfgrenze mit Grund und Folge, wo die Menge unbeschränkt bleibt), F-2, F-3, F-4, F-6, F-7 und den Fall zu F-5; jeder Fix deckt die Klasse (`AGENTS.md` §3.17), nicht die genannte Mutation. An den **Architect**: F-5 (Zelle und Entscheidung), die Mengen zu F-2 und F-4, die Allaussage zu F-1 (Generator in der Suite oder Prüfgrenze), I-1 (Zahlen von H1 und H2). An den **Planner**: F-8 (b) und die Entscheidung über Teil Z (I-4). Die Finding-Klassen gehen in die Slice-Closure §7. Der Report ersetzt keine Verifikation.
