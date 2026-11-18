# Slice slice-expand-parser: `$expand` in der OData-URL, Parser und AST

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-expand-batch](../welle-expand-batch.md).

**Bezug:** `LH-FA-16`, `LH-RB-01`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-16.a` (Erweiterungspunkt `$expand`, eigener Parser-Contract-Test, unbekannte Operationen `501`) · `SPEC-160` (Nr. 2 Ergebnis, Nr. 3 Veränderlichkeit, Nr. 4 Fehler, Nr. 5 Rangfolge, Nr. 6 Grenzen, Nr. 7 Nebenläufigkeit, Nr. 8 Einhängestelle, Nr. 9 e) · `SPEC-262` (Vertrag von `$expand`) · `SPEC-263` (Prüfung gegen das Modell, Vertrag für `slice-expand-validator`) · `SPEC-162` (Nr. 2, Nr. 3 a und b: Klasse von `$expand` mit und ohne Einhängestelle, Nr. 9; Nr. 6 und 7 als Vorbild der Liste) · `SPEC-163` (`ODR-E6005`) · `SPEC-022` (`ODR-E6001`) · `SPEC-049` (Zeile GET: Navigation-Expand, Spalte „Nicht in MVP“) · `SPEC-042` (Fixture `ODV4-E-03` c bleibt unverändert) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkte: separate AST-Knoten, Validatoren und Executors).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-expand-parser)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Parser der OData-URL (`urlparser`, Port `URLCodec`, `SPEC-160`) liest die Systemoption `$expand` aus `raw_query` und überführt sie in einen eigenen Knoten des Query-AST (`model.Optionen`, `src/hexagon/model/url_ast.go`). Er liest nur Text und kennt das Modell nicht (`SPEC-160` Nr. 9): eine Liste im unterstützten Umfang ergibt den Knoten, eine unlesbare Form `400` (`ODR-E6005`), eine lesbare, aber nicht unterstützte Form `501` (`ODR-E6001`), jeweils mit den bestehenden Ursachen von `SPEC-160` Nr. 4 (`SPEC-262`, §6 A6). Die Option ist **zuschaltbar** (`New` mit einer Option, `SPEC-160` Nr. 8 Schlusssatz): ohne sie bleibt `$expand` `option_nicht_unterstuetzt`, das Binary, `ODV4-E-03` c (`GET /Products?$expand=Foo` ist `501`) und die Zeile `abnahme_matrix_test.go` ändern sich nicht. Der Slice greift auf keine Daten und kein Modell zu (`LH-FA-16.a`: Parser-Contract-Test).

**Schnitt vor dem Code (§4):** Der Validator gegen die Navigationen des Modells (`model.PruefeExpand`, Muster von `SPEC-167`/`PruefeFilter`) ist aus diesem Slice herausgenommen und gehört `slice-expand-validator`. Der Plan führte beide in einem Slice; die Schnittfrage von `SPEC-128` Nr. 1 gilt dafür (§4). Der Validator ist keine Aufgabe des Parsers: `SPEC-160` sagt, der Parser kenne das Modell nicht, und `ARC-011` lässt den Adapter nur Modell und Ports importieren.

**Umfang (entschieden in `SPEC-262`, §6 A1 bis A9):** `$expand=Name1,Name2` — eine Liste einfacher Namen (Namensform wie `$select`, `SPEC-162` Nr. 6) ohne Leerraum; `*`, Pfade (`a/b`, auch mit Typ-Cast und `$ref`) und jeder Wert mit einer Klammer (verschachtelte Optionen, `$levels`) sind lesbar und `501` (`wert_nicht_unterstuetzt`). Damit hängt der Slice weder an `slice-odata-url-filter` noch an `slice-edm-ausdruecke` (beide in `done/`): `$filter` in Klammern ist nicht im Umfang.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Validator gegen das Modell (Navigation vorhanden, auch geerbt; Property ohne Navigation; Subtyp) — `slice-expand-validator` (Schnitt, §4); dessen §1 führt den Punkt mit der Kennung dieses Slice als Geber.
- Verdrahtung: das Einschalten der Option im Aufbau des Dienstes, der Aufruf des Validators vor der Ausführung und die Änderung von `ODV4-E-03` c — `slice-expand-ausfuehrung` (dessen §1 trägt die Sendung, `AGENTS.md` §3.13); bis dahin liest das Binary `$expand` nicht und antwortet `501`; den Aufbau ohne die Option hält `TestSatzVersion4OhneExpand`.
- Auflösen der Navigation gegen Daten, Datenbasis der Navigationen und Ausgabe der eingebetteten Entities — `slice-expand-ausfuehrung` (Executor- und Serializer-Contract-Test von `LH-FA-16.a` liegen dort); dieser Slice liefert den Knoten.
- Verschachtelte Optionen im `$expand` (`$filter`, `$select`, `$orderby`, `$top`, `$skip`, `$count`, `$levels`, `$expand`) — entschieden (§6 A1, `SPEC-262` Nr. 3 b): ein Wert mit `(` ergibt `501`.
- `$expand` im V2-Parser (`urlparserv2`: Kompatibilitätsfehler `option`, `SPEC-174`) und die Leseregel des Imports (`SPEC-127` Nr. 8 i: `$expand` ist dort lesbar) — andere Verträge, hier unberührt.
- `$apply`, `$search`, Actions, Functions, Deep Insert — nicht im ersten Abnahmeumfang (Lastenheft §5); sie bleiben `501`.
- `$batch` — `slice-batch-parser`; die beiden Äste sind unabhängig.
- Herstellerspezifische Navigations-Erweiterungen — Out-of-Scope von `LH-FA-16`.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Grammatik, AST und Einhängestelle: `$expand` (Liste im Umfang des Architect, Dekodierung, Trennzeichen, Grenzen, Rangfolge der Verstöße) wird zu einem Knoten in `model.Optionen`; die Option ist zuschaltbar, ohne sie bleibt `$expand` `option_nicht_unterstuetzt`; jedes bestehende Feld und jedes Ergebnis ohne `$expand` bleibt gleich.
- [ ] Randformen des Optionstextes (leer, leeres Element, doppelte Option, doppeltes Element, `*`, Pfad, `$ref`, Typ-Cast, Klammer, Zahl der Elemente, Länge des Namens, Dekodierung, Groß- und Kleinschreibung): je Zelle des Rasters (§6) ein Ort in der Spezifikation und ein Fänger der Tabelle (§3), als Generator über das Alphabet der Grammatik mit einem Orakel ohne Code des Lesers.
- [ ] Parser-Contract-Test (`LH-FA-16.a`): ein eigener Test fährt den Parser gegen die Beispiele der Spezifikation (Tabelle als Fixture); `make a-check` hält die Richtung (Parser importiert kein Modell des Dienstes, keinen Dienst).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-160`, `SPEC-162`, die neue Nummer des Architect).
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (A6, `SPEC-262`: keine neue Ursache, kein neuer Code; der Modus `mock` erreicht das Binary erst mit `slice-bootstrap-m3`, die Antwort auf `$expand` zeigt sich erst mit `slice-expand-ausfuehrung`).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/model/url_ast.go`, `src/hexagon/model/url_ast_expand.go` | update, neu | Knoten für `$expand` im Query-AST, technologieunabhängig; Feld in `Optionen` (Form nach A8, `SPEC-262`) |
| `src/adapters/driven/odata/urlparser/` | update | Lesen der Option (`query.go`, neue Datei `expand.go`), Einhängestelle in `New`; `$expand` verlässt `nichtUnterstuetzteOption` nur mit eingehängter Option |
| `spec/spezifikation.md` | update | Vertrag mit Randformen, Beispiele (Architect): `SPEC-262` und `SPEC-263` (neu), Nachzug in `SPEC-160` (Typen, Nr. 2, 3, 6, 8, 9) und `SPEC-162` (Nr. 2, 3, 9) |
| Tests | neu | `urlparser/expand_*_test.go` (Fälle, Alphabet-Generator mit Orakel, Rangfolge, Parser-Contract, Nebenläufigkeit), `model/url_ast_expand_test.go`, `src/bootstrap/expand_wurzel_test.go` (Aufbau ohne die Option) |
| `src/adapters/driven/odata/urlparser/zustand_test.go`, `grenzen_test.go` | update | `TestParserOhneZustand` führt die zwei Felder des Parsers (`SPEC-262` Nr. 8), `textsumme` zählt die Namen von `Expand` |

`src/hexagon/ports/outbound/` bleibt unverändert: die Signatur von `URLCodec.ParseURL` und die Felder von `SPEC-160` ändern sich nicht (`SPEC-160` Nr. 8 Schlusssatz). Die Zeile des früheren Plans, die den Port „für den Validator“ erweiterte, entfällt mit dem Schnitt: der Validator ist eine reine Funktion in `model` (Muster `PruefeFilter`) und braucht keinen Port.

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen als Entwurf des Planner im Plan; der Architect bestätigt oder ändert die Zeilen mit dem Vertrag in der Spezifikation **vor** dem ersten Code-Commit (`AGENTS.md` §3.12). Ändert er den Umfang (§6 A1), setzt der Planner die Marke (§4) vor dem Code neu. Die Randformen sind entschieden (`SPEC-262`): der Architect hat die Zeilen 8, 9, 12 bis 16 an den Vertrag angepasst, Zeilenzahl und `produktcode` bleiben; der Implementer entscheidet keine.

- Der Parser liest fremde Bytes (den Text der URL): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); Optionstexte mit Escape-Folgen und Zeichen jenseits von ASCII entstehen aus Bytewerten (`AGENTS.md` §3.18).
- Orakel des Generators: der Test baut die erwartete Klasse (Knoten · `wert` · `wert_nicht_unterstuetzt` · `zu_viele`) mit einer eigenen, kurzen Tabelle aus dem Wortlaut der Spezifikation und ruft keinen Code des Lesers (`BEO-REPO/menge-im-plan-weicht-vom-orakel-des-tests-ab`, `BEO-REPO/menge-ohne-ihre-positionen`). Alphabet und Prüfgrenze: die Zeichen `a`, `Z`, `_`, `0`, `,`, `/`, `*`, `(`, `)`, `$`, `.`, `;`, `=`, `%20`, `%2C` und zwei Zeichen jenseits von ASCII aus Bytewerten (ein Buchstabe, die Bytes `0xC3 0xA9`, und ein Nicht-Buchstabe, NBSP; nach `SPEC-262` Nr. 4 a ist der erste ein Namenszeichen, der zweite nicht), Folgen bis Länge 4 vollständig; längere Werte nur als benannte Familien (Grenzen, Wiederholung). Das ist das akzeptierte Negativ der Klasse „jede Zeichenfolge“ (Nr. 24).

**Tabelle *Zusage · Fänger*** (Entwurf des Planner, vom Architect mit `SPEC-262` bestätigt und in den Zeilen 8, 9, 12 bis 16 angepasst; `AGENTS.md` §3.15; Dateien liegen unter `src/adapters/driven/odata/urlparser/`, Mutationen fährt der Implementer vor der Übergabe selbst):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Ohne eingehängte Option bleibt `$expand` mit jedem Wert (leer, ungültig kodiert, mehrfach) `option_nicht_unterstuetzt`, der Wert bleibt ungelesen | `TestExpandOhneEinhaengung`: Schleife über Werte (leer, `%G1`, Alphabet bis Länge 3) und Wiederholungen 1 bis 3; die bestehenden Fälle `query_test.go` (`$expand=a&$expand=b`) und `rang_test.go` bleiben unverändert grün | Die Vorgabe von `New` liest die Option | andere |
| 2 | Mit eingehängter Option ergibt eine gültige Liste `Optionen.Expand` mit den Namen in der Reihenfolge der URL | `TestExpandListe`: Listen der Länge 1 bis 32, je in aufsteigender und absteigender Folge (64), und alle 24 Reihenfolgen der Länge 3 aus vier Namen | Die Elemente werden sortiert | andere |
| 3 | Fehlt `$expand`, ist `Expand` der Nullwert und jedes andere Feld gleich dem Ergebnis ohne eingehängte Option | `TestExpandDifferenz`: jede Query ohne `$expand` aus `korpus()` (`nebenlaeufig_test.go`: Gut- und Fehlerfälle aller Klassen, dazu 25 erzeugte) gegen `New()` und `New(WithExpand())`, tiefe Gleichheit samt Fehlertext, `Expand` ist `nil`; je Ressourcenart (`/$metadata`, `/S`, `/S(1)`, `/S/$count`) liest der Parser die Option; die Tabellen von `query_test.go` und `rang_test.go` sind lokale Werte der Tests und bleiben unverändert grün | Die Einhängung ändert ein anderes Feld (etwa `Sonstige`) | andere |
| 4 | Name und Wert der Option werden genau einmal dekodiert; `%24expand` und die Schreibweisen der Hex-Ziffern gelten wie das rohe Zeichen | `TestExpandDekodierung`: Schleife über Escape-Formen des Namens (roh, `%24`, Groß- und Kleinschreibung der Hex-Ziffern) und des Werts (`%2C` als Trenner, `%252C` bleibt `%2C` und ist `wert`); Fixture aus Bytewerten | Der Wert wird zweimal dekodiert | andere |
| 5 | Eine zweite Option `$expand` ist `option_doppelt`, in jeder Schreibweise des Namens, mit gleichem und mit anderem Wert, auch wenn der zweite Wert ungültig ist | `TestExpandDoppelt`: Paare aus {roh, `%24`} × {gleicher, anderer, ungültiger Wert} × Reihenfolge | Das zweite Vorkommen überschreibt das erste | andere |
| 6 | Ein leerer Wert (`$expand=`) ist `wert` | `TestExpandLeer`: `$expand=` und `$expand` ohne `=` | Leer ergibt den Nullwert ohne Fehler | andere |
| 7 | Ein leeres Element an jeder Stelle (`,a`, `a,`, `a,,b`, `,`) ist `wert` | `TestExpandLeeresElement`: Schleife über die Stelle (Anfang, Mitte, Ende) und die Länge der Liste 1 bis 4 | Leere Elemente werden übergangen | andere |
| 8 | Leerraum und jedes Zeichen außerhalb der Namensform an jeder Stelle des Namens (Anfang, Mitte, Ende) ist `wert`; ein Zeichen jenseits von ASCII ist Teil eines Namens genau dann, wenn es ein Buchstabe (Kategorie `L`) oder, nicht am Anfang, eine Ziffer (`Nd`) ist (`SPEC-262` Nr. 4 a) | `TestExpandNamenform`: Schleife über alle 128 ASCII-Zeichen außer `,`, `/` und `(` (eigene Zeilen Nr. 7, 11, 13), roh und als `%XX`, und sieben Zeichen jenseits von ASCII aus Bytewerten (é `0xC3 0xA9`, Ж, 中 sind Buchstaben, ٣ ist eine Ziffer und gültig außer am Anfang; NBSP, BOM und ² sind `wert`) × drei Stellen × roh und als `%XX` in beiden Schreibweisen; Orakel: Zeichenklasse aus `SPEC-125` Nr. 8 b als eigene Tabelle, nicht aus dem Code des Lesers | Die Prüfung der Namensform prüft nur das erste Zeichen; jedes Zeichen jenseits von ASCII wird verworfen | andere |
| 9 | Ein Name ist 1 bis 128 Zeichen lang (Zeichen, nicht Bytes), beginnt mit Buchstabe oder `_`; 129 Zeichen und eine Ziffer am Anfang sind `wert` | `TestExpandNamenGrenze`: beide Seiten der Grenze (1, 128, 129, 0) × erstes Zeichen {Buchstabe, `_`, Ziffer} × Breite des Zeichens {ein Byte, zwei Bytes (`0xC3 0xA9`)} | Grenze 129 statt 128; Länge in Bytes statt in Zeichen | andere |
| 10 | `*`, allein oder neben Namen an jeder Position, ist `wert_nicht_unterstuetzt` | `TestExpandStern`: Schleife über Position (vor, nach, zwischen Namen) und Zahl der Namen 0 bis 3; `**` und `*a` gesondert | `*` wird als Name gelesen | andere |
| 11 | Ein Pfad aus mindestens zwei einfachen Namen (`a/b`, `a/b/c`) ist `wert_nicht_unterstuetzt`; ein leeres Segment (`/a`, `a/`, `a//b`) ist `wert` | `TestExpandPfad`: Segmentzahl 2 bis 9 (Prüfgrenze), leeres Segment an jeder Stelle | Ein Pfad wird als erster Name gelesen | andere |
| 12 | Ein Segment eines Pfads ist ein Name, ein qualifizierter Name (Typ-Cast) oder `$ref`, `$count`, `$value`: der Pfad ist `wert_nicht_unterstuetzt`, an jeder Stelle; jedes andere Segment (`*`, `$levels`, `$`, `a.`, `.a`) macht das Element zu `wert`; ein qualifizierter Name oder `$ref` allein ist `wert` (`SPEC-262` Nr. 4 c, d) | `TestExpandBesondereSegmente`: Schleife über die Segmente {Name, `Ns.Typ`, `a.b.c`, `$ref`, `$count`, `$value`, `*`, `$levels`, `$`, `a.`, `.a`} × Stelle {erstes, mittleres, letztes Segment eines Pfads} und jedes Segment allein; Orakel: Tabelle des Vertrags | Ein Segment mit `$` wird als Name gelesen; `*` wird als Segment zugelassen | andere |
| 13 | Ein `(` irgendwo im Wert macht den ganzen Wert zu `wert_nicht_unterstuetzt`, der Rest wird nicht gelesen: auch hinter einem leeren oder ungültigen Element und bei mehr als 32 Elementen; ein `)` ohne `(` ist `wert` (`SPEC-262` Nr. 3 b) | `TestExpandKlammer`: Schleife über jede Systemoption (`$filter`, `$select`, `$orderby`, `$top`, `$skip`, `$count`, `$levels`, `$expand`, `$search`) in `a(...)`, über die Stelle des `(` (Anfang, Mitte, Ende des Elements; Element 1 bis 3) und über die Vorbelastung davor {keine, leeres Element, ungültiger Name, doppeltes Element, 33 Elemente}; dazu die Formen `a(`, `(`, `a()`, `a)`, `)a(` und `a($select=Id,Name),b`; Ergebnis immer `wert_nicht_unterstuetzt`, außer `a)` (`wert`) | Das `(` wird erst nach der Zahl oder den Elementen gesucht | andere |
| 14 | Mehr als 32 Elemente sind `zu_viele`, vor der Prüfung der Elemente; 32 sind gültig; ein `(` im Wert geht der Zahl vor (Nr. 13) | `TestExpandZahl`: 31, 32, 33, 34, 65, 200 Elemente × {keines, erstes, letztes Element ungültig} × Ort `query/<n>` (n = 1 bis 3); dazu 31, 32 und 33 leere Elemente | Grenze 33; Prüfung der Elemente vor der Zahl | andere |
| 15 | Ein Element, das einem früheren bytegleich ist (`a,a`, nicht benachbart: `a,b,a`; auch `*,*` und `a/b,a/b`), ist `wert`; die Schreibweise zählt (`a,A` sind zwei) | `TestExpandDoppeltesElement`: Abstand 0 bis 3 zwischen den Vorkommen × Form {Name, `*`, Pfad}, Groß- und Kleinschreibung | Duplikate werden zusammengeführt; Vergleich ohne Beachtung der Schreibweise; `*,*` ergibt `wert_nicht_unterstuetzt` | andere |
| 16 | Mehrere Verstöße im Wert: ein `(` geht allem vor, danach `zu_viele` vor den Elementen; Klasse `ungueltig` vor `nicht_unterstuetzt`; innerhalb der Klasse das erste Element von links (`SPEC-262` Nr. 3) | `TestExpandRangImWert`: alle Paare aus {leer, Name ungültig, doppelt, `*`, Pfad, Klammer} an den Stellen i und j (i<j und i>j), erwartete Ursache aus der Zusatztabelle in §6 | Die Klassen werden nach Position statt nach Klasse entschieden; die Klammer wird als Element geprüft | andere |
| 17 | Rangfolge über Optionen: ein Verstoß in `$expand` und ein Verstoß einer anderen Option folgen `SPEC-160` Nr. 5, in beiden Reihenfolgen der Query | `TestExpandRangQuery`: Paare aus {Verstoß von `$expand`} × {`$top=-1`, `$foo=1`, `$select=a,,b`, `$filter` ohne Leser} × {links, rechts}; `rang_test.go` bleibt das Muster | `$expand` wird vor den anderen Optionen entschieden | andere |
| 18 | Der Ort ist `query/<n>` der Option (`n` zählt die nichtleeren Segmente), ohne Teil der Eingabe | `TestExpandOrt`: leere Segmente (`&&`) vor der Option, Option an Stelle 1 bis 4 | `n` zählt die leeren Segmente mit | andere |
| 19 | Kein Teil der Eingabe steht im Fehler: weder Name noch Wert noch Position in Bytes | `TestExpandFehlertext`: Schleife über alle Verstoßarten aus Nr. 5 bis 16, Eingabe mit unverwechselbarem Namen; `Error()` enthält ihn nicht | Der Name steht in der Meldung | andere |
| 20 | Das Ergebnis gehört dem Aufrufer: eine Änderung der Elemente ändert kein späteres Ergebnis, kein Slice wird zwischen Ergebnissen geteilt | `TestExpandKopie`: zweimal gleiche Eingabe, erstes Ergebnis ändern (Element, Namen), zweites vergleichen; Schleife über Listenlängen 1 bis 3 | Ein Paket-Slice wird als Ergebnis geliefert | andere |
| 21 | Gleiche Eingabe ergibt gleiches Ergebnis; ein Fehler gibt `ODataURL{}` | `TestExpandWiederholung`: jede Eingabe der Nr. 2 bis 16 zweimal, tiefe Gleichheit; Fehlerfälle gegen den Nullwert | Ein Zähler im Parser färbt das zweite Ergebnis | andere |
| 22 | Die Summe der Längen aller Texte im Ergebnis bleibt höchstens Eingabe plus 48 (`SPEC-160` Nr. 6 b) | `TestExpandWaechter`: Zahl {1, 2, 16, 31, 32} × Länge der Namen {1, 128} × {roh, jedes Byte als `%XX`} (20), dazu 32 Namen zu 128 Zeichen mit einer Query von 8191, 8192 und 8193 Bytes (Auffüllung durch eine fremde Option; 8193 ist `zu_lang`) und ein Name aus 128 Zeichen zu zwei Bytes; die Textsumme gleicht genau der Summe der Namen, die Zuteilung bleibt unter 64 · n + 65536 | Das Element trägt eine Kopie des ganzen Werts | andere |
| 23 | Gleichzeitige Aufrufe auf einer Instanz und auf mehreren beeinflussen sich nicht; jedes Ergebnis gleicht dem eines einzelnen Aufrufs (`SPEC-160` Nr. 7) | `TestExpandGleichzeitig`: 4 Goroutinen × 50 Runden × 490 Eingaben (Listen, Verstöße je Art, 432 Paare der Rangfolge) auf einer gemeinsamen und einer eigenen Instanz, Ergebnis gegen den Einzelaufruf, dessen Elemente kopiert sind (Wirkungs-Fänger, verglichen erst nach dem Ende aller Goroutinen); `make test-race` je Prozess für ein geteiltes Feld ohne Wirkung | Ein Puffer im Parser wird zwischen Aufrufen geteilt | nebenläufig |
| 24 | Jede Zeichenfolge über dem Alphabet der Grammatik bis zur Prüfgrenze (Länge 4, §3) hat die Klasse des Orakels: Knoten, `wert`, `wert_nicht_unterstuetzt` oder `zu_viele` | `TestExpandAlphabet`: vollständige Aufzählung aller Folgen bis Länge 4 über 17 Zeichen (88 740 Folgen), Orakel ohne Code des Lesers; akzeptiertes Negativ: längere Folgen nur als Familien (Nr. 9, 14, 22), Folge: eine Regel, die erst ab Länge 5 greift, bleibt ungefangen | Der Leser behandelt `;` oder `=` im Element als Trenner | andere |
| 25 | Der Parser liest kein Modell und keinen Dienst: die Richtung der Importe bleibt | `make a-check` (Regeln des Hexagons, Ausgabe grün); ein Import von `model.ServiceModel` oder eines Dienstes im Paket wird rot | `urlparser` importiert `services/mock` | andere |
| 26 | Die Beispiele der Spezifikation (Tabelle der Gültigen, der `400` und der `501`) ergeben die genannte Antwort: eigener Parser-Contract-Test | `TestExpandContract`: Tabelle der Beispiele als Fixture (Eingabe, Knoten oder Ursache und Ort), ein Fall je Zeile; keine Eingabe aus dem Generator | Die Antwort einer Zeile wird vertauscht | andere |
| 27 | Der Knoten trägt die Felder des Vertrags, jedes andere Feld von `Optionen` bleibt unverändert (Name, Typ, Reihenfolge) | `TestOptionenFelder` in `model`: die Feldliste von `Optionen` und des Knotens per `reflect` gegen die Liste des Vertrags | Ein Feld von `Optionen` wird umbenannt oder entfällt | andere |
| 28 | Das Binary bleibt unverändert: `GET /Products?$expand=Foo` ist `501` mit `ODR-E6001` (`ODV4-E-03` c) | die bestehende Zeile in `src/bootstrap/abnahme_matrix_test.go` (c) baut ihren Dienst selbst und sieht den Aufbau in `wurzel.go` nicht (Mutation grün gesehen); der Fänger des Aufbaus ist `TestSatzVersion4OhneExpand` in `src/bootstrap/expand_wurzel_test.go`: `SatzVersion4().URL` liest `$expand` nicht (4 Werte), `$filter` bleibt lesbar | Der Aufbau des Dienstes hängt die Option ein | andere |
| 29 | Akzeptiertes Negativ: die Länge des Werts hat keine eigene Grenze außer `rawQuery` (8192 Bytes) und der Zahl der Elemente | kein Fänger; Grund: die Summe der Längen ist über Nr. 6 b und die Grenze von 8192 begrenzt; Folge: ein Wert aus 32 Namen von 128 Zeichen liegt unter der Grenze der Query und ist gültig | — | andere |
| 30 | Akzeptiertes Negativ: ob `slice-expand-ausfuehrung` den Knoten ohne Änderung des Parsers liest, belegt erst dessen Slice | kein Fänger hier; Grund: der Leser des Knotens liegt nicht in diesem Slice; Folge: eine nötige Änderung der Form dort ist eine Änderung dieses Vertrags (`AGENTS.md` §3.13, Adresse Architect) | — | andere |

**Abweichungen der Umsetzung vom Entwurf der Tabelle** (Ort: die Zeilen oben tragen den Ist-Stand): die Zeilen 2, 3, 8, 11, 14, 22 und 23 führen größere Mengen als der Entwurf (Elementzahl in der Mengen-Tabelle unten); Zeile 28 bekommt einen Test, weil die Tabelle der Abnahme ihren Dienst selbst aufbaut und die Zeile des Aufbaus nicht sieht (Mutation grün gesehen, siehe unten); `TestParserOhneZustand` (`zustand_test.go`) führt zwei Felder (`filter` als Interface, `expand` als Bool) statt eines (`SPEC-262` Nr. 8), und `textsumme` (`grenzen_test.go`) zählt die Namen von `Expand` mit. Weitere Tests: `src/bootstrap/expand_wurzel_test.go` (neu, außerhalb der Pfade der Größenmarke).

**Mengen** (`AGENTS.md` §3.17; Orakel ist eine eigene Tabelle im Test, `exOrakel` in `expand_hilfe_test.go`, ohne Code des Lesers; Zeichen jenseits von ASCII als Konstanten aus Bytewerten in `\x`-Schreibweise, die Dateien der Änderung mit `LC_ALL=C grep -P` auf BOM, NBSP und Backslash-u-Folgen gegengelesen: keine Fundstelle):

| Nr. | Menge | Generator · Achsen | Orakel | Elemente |
|---|---|---|---|---|
| 1 | Wert ohne Einhängung | 2 Parser × 3 Wiederholungen × (5 Sonderwerte + 5219 Folgen bis Länge 3) + 2 ohne `=` | feste Antwort `option_nicht_unterstuetzt` | 31 346 |
| 2 | Listen | Länge 1 bis 32 × {auf, ab} · 3 aus 4 Namen | Eingabefolge | 64 + 24 |
| 4 | Dekodierung | Schreibweisen von Name und Wert, Hex-Ziffern groß und klein, Escapes | Tabelle von Hand | 24 |
| 5 | Wiederholung der Option | 3 Schreibweisen² × 4 erste × 5 zweite Werte × 2 Stellen | Tabelle von Hand | 360 |
| 7 | leeres Element | 1 bis 4 Namen × alle Stellen, `,`, `,,`, `a,,` | `wert` | 17 |
| 8 | Zeichen im Namen | 125 ASCII-Zeichen × 3 Stellen × {roh, `%XX` groß, klein} (ohne Steuerzeichen, `&`, `%` roh und ohne `%00`) und 7 Zeichen jenseits von ASCII × 3 Stellen × 3 Schreibweisen | `exIstName` (Tabelle der Buchstaben und Ziffern) | 270 + 744 + 63 |
| 9 | Länge des Namens | {0, 1, 128, 129} × {Buchstabe, `_`, Ziffer} × {1 Byte, 2 Bytes} | Zeichenzahl, nicht Bytes | 24 |
| 10, 11 | `*` und Pfad | Position × 0 bis 3 Namen; Segmentzahl 2 bis 9 × leeres Segment an jeder Stelle | `501` oder `wert` | 10 + 5; 52 + 2 |
| 12 | Segmente | 11 Segmente × 3 Stellen und allein | Tabelle `zulaessig` | 44 |
| 13 | Klammer | 9 Optionen × 3 Stellen × 7 Vorläufe × 2 Nachläufe, 7 feste Formen | immer `501`, außer `a)` | 378 + 7 |
| 14 | Zahl | {31, 32, 33, 34, 65, 200} × 3 Verstöße × 3 Orte; 31, 32, 33 leere Elemente | Zahl > 32 `zu_viele`, sonst Verstoß, sonst Knoten | 54 + 3 |
| 15 | doppeltes Element | Abstand 0 bis 3 × {Name, `*`, Pfad} × {gleich, Schreibweise} | gleich `wert`, sonst Name oder `501` | 20 |
| 16 | Paare im Wert | 6 × 6 Arten × 12 Stellenpaare | Klammer, dann `ungueltig`, dann `501` | 432 |
| 17 | Paare über Optionen | 4 × 4 Verstöße × 2 Seiten | `rang` aus `rang_test.go` | 32 |
| 18 | Ort | Stelle 1 bis 4 × {0, 1, 3} leere Segmente × 3 Orte | `query/<n>` | 36 |
| 22 | Wächter | {1, 2, 16, 31, 32} × {1, 128} × {roh, `%XX`}; 32 × 128 mit 8191, 8192, 8193 Bytes; 2-Byte-Name | Textsumme gleich Summe der Namen, Zuteilung | 24 |
| 23 | Nebenläufigkeit | 4 Goroutinen × 50 Runden × 490 Eingaben | Einzelaufruf, kopiert | 98 000 Aufrufe |
| 24 | Alphabet | alle Folgen der Länge 1 bis 4 über 17 Zeichen (`a`, `Z`, `_`, `0`, `,`, `/`, `*`, `(`, `)`, `$`, `.`, `;`, `=`, `%20`, `%2C`, é, NBSP) | `exOrakel` | 88 740 |
| 26 | Beispiele | Tabelle der Spezifikation, 27 Zeilen, dazu 4 Ressourcenarten | Wortlaut der Tabelle | 27 + 4 |

Prüfgrenzen mit Folge im Betrieb (akzeptierte Negative, `AGENTS.md` §3.17): Länge 4 über 17 Zeichen (Nr. 24), Segmentzahl 9 (Nr. 11), Zahl der Elemente 200 (Nr. 14), höchstens vier leere Elemente und drei Namen neben `*`; eine Regel, die erst jenseits greift und keine Familie trifft, bleibt ungefangen. Gesehen: die Mutation `n03` (Regel nur bei Länge 9 des Werts) bleibt im Alphabet-Test grün und wird nur zufällig von `TestExpandListe` gefangen (`rot-paket`).

**Mutationen, vom Implementer gefahren** (`sh tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, ein Lauf mit 52 Mutationen, Klasse `andere`, ein Prozess je Stufe; Host-Last der Ergebniszeilen 7,1 bis 9,6 bei 4 CPUs, Messender der Lauf; Ergebnis 50 `rot`, 1 `rot-paket`, 1 `flatternd`, 0 `gruen`, 0 Übersetzungsfehler). Die Namen der Mutationen tragen die Nr. der Zeile; zu jeder Zeile ein Element, das der Entwurf nicht nannte, ist mit `n…` bezeichnet:

| Nr. | Mutationen (alle `rot`, wo nichts anderes steht) | roter Test |
|---|---|---|
| 1 | `New` liest die Option | `TestExpandOhneEinhaengung` |
| 2 | Elemente sortiert; `WithExpand` setzt nichts | `TestExpandListe`; `TestExpandContract` |
| 3 | die Einhängung lässt `Sonstige` weg | `TestExpandDifferenz` |
| 4 | Wert zweimal dekodiert; Name zweimal dekodiert | `TestExpandDekodierung` |
| 5 | zweites Vorkommen überschreibt | `TestExpandDoppelt` |
| 6, 7 | leer ergibt den Nullwert; leere Elemente übergangen | `TestExpandLeer`, `TestExpandLeeresElement` |
| 8 | nur das erste Zeichen geprüft; jedes Zeichen jenseits von ASCII verworfen | `TestExpandNamenform` |
| 9 | Grenze 129; Länge in Bytes | `TestExpandNamenGrenze` |
| 10, 11 | `*` als Name; Pfad als erster Name; Pfad höchstens 3 Segmente, `n07`: höchstens 7 | `TestExpandStern`, `TestExpandPfad` |
| 12 | `$levels` als Segment; `*` als Segment; `$ref` kein Segment | `TestExpandBesondereSegmente` |
| 13 | Klammer nach der Zahl; Klammer nach den Elementen; `n05`: letztes Zeichen ungeprüft | `TestExpandKlammer` |
| 14 | Grenze 33; Elemente vor der Zahl; `n04`: nur genau 33 Elemente | `TestExpandZahl` |
| 15 | Duplikate zusammengeführt; Vergleich ohne Schreibweise; `*,*` ergibt `501` | `TestExpandDoppeltesElement` |
| 16 | ein `501` beendet das Lesen; Klammer wie ein Element | `TestExpandRangImWert` |
| 17 | `501` sofort als Fehler | `TestExpandRangQuery` |
| 18 | Ort ab 0; leere Segmente zählen (in `model.QuerySegmente`) | `TestExpandOrt` |
| 19 | Name im Ort der Meldung | `TestExpandFehlertext` |
| 20 | Paket-Slice als Ergebnis | `TestExpandKopie` |
| 21 | Zähler färbt das zweite Ergebnis | `TestExpandWiederholung` |
| 22 | Element trägt den ganzen Wert | `TestExpandWaechter` |
| 23 | Feld `zaehler` im Parser | `TestParserOhneZustand` |
| 24 | `;` trennt; `=` trennt; `n01`: `$` als Präfix eines Namens; `n02`: `)` wie `(`; `n03`: Regel nur bei Länge 9 (`rot-paket`, Alphabet-Test grün) | `TestExpandAlphabet` |
| 25 | Import eines Nachbar-Adapters; Import eines Dienstes (`make a-check`: `wrong-direction`, von Hand gesehen) | `TestImporte` |
| 26 | `$ref` kein Segment (Zeile `Category/$ref`) | `TestExpandContract` |
| 27 | zusätzliches Feld in `Optionen`; im Knoten | `TestOptionenFelder` |
| 28 | `wurzel.go` hängt `WithExpand()` ein: `TestAbnahmeMatrix` bleibt grün (die Tabelle baut ihren Dienst selbst), `TestSatzVersion4OhneExpand` ist rot (4 von 4 Werten, von Hand gefahren: das Paket `src/bootstrap` läuft im Mutationslauf nicht, `Ausgangsstand rot`) | `TestSatzVersion4OhneExpand` |

**Raten der Nebenläufigkeit** (Zeile 23, `TestExpandGleichzeitig`, je Mutation 50 Prozesse mit Detektor und 50 ohne, `-test.count=1`, 4 Goroutinen je Prozess, Host-Last 8,4 / 7,3 / 7,1 bei 4 CPUs, Messender `mutationslauf`; jede Zahl ist eine Größenordnung unter Last): geteilter Puffer der Elemente (Wirkung): mit Detektor 50/50, ohne 50/50; geteilter Zähler mit Wirkung: 50/50 und 50/50; geteilter Zähler ohne Wirkung (nur ein Datenrennen): mit Detektor 50/50, ohne 0/50. Das Datenrennen ohne Wirkung fängt allein `make test-race`; der Wirkungs-Fänger fängt Puffer und Wirkung in jedem Lauf, ohne auf das Rennen zu warten.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-odata-url-pfad`, `slice-odata-url-filter` und `slice-edm-ausdruecke` liegen in `done/` (am 2026-11-18 geprüft); `$filter` in Klammern ist nicht im Umfang (`SPEC-262` Nr. 3 b). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-odata-url-pfad` (Typen des Query-AST, Paket `urlparser`, Zerlegung von `raw_query`, Fehlerbildung `400`/`501`; in `done/`). Das Modell mit den Navigation-Properties braucht dieser Slice **nicht** mehr (der Validator ist abgeschnitten). Der Welle-Trigger (`welle-mock` in `done/`) ist Reihenfolge-Vorsicht, keine Code-Abhängigkeit.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Query-Interpreter, JSON-Schreiber, Workspace und Composition Root (der Slice liest Text), M2 und `slice-bootstrap-m3` (das Binary ändert sich erst mit der Verdrahtung), `slice-expand-validator` (liest den Knoten dieses Slice, nicht umgekehrt), `slice-expand-ausfuehrung`, `slice-batch-parser`.
- **Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`; die Faktoren stehen allein in `SPEC-128` Nr. 1): gezählt ist die Tabelle in §3 mit 30 Zeilen, *a* = 1 nebenläufige (Zeile 23), *b* = 29 andere (die beiden akzeptierten Negative zählen als `andere`, `SPEC-128` Nr. 3 e), *f* = 0 (der Slice hat keinen Port-Fake und keinen Test-Server), *p* = 150 (Erwartung des Planner, vom Architect bestätigt, Spanne 100 bis 250: Lesen der Liste rund 70 Zeilen, Knoten und Feld rund 40, Einhängestelle und Verdrahtung in `query.go` rund 40). *S* = 85·1 + 80·29 + 200·0 + 150 = **2555**, 135·2555 = 344925, Marke *M* = 100 · ⌈344925 / 10000⌉ = **3500**. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks (das Paket `urlparser`, der Knoten in `src/hexagon/model/`), Tests eingeschlossen. `basis` ist `HEAD` vor dem Vertrag des Architect: er ändert nur `spec/`, das kein Pfad des Blocks ist (`SPEC-128` Nr. 3 m), die Messung bleibt bis zum ersten Code-Commit 0. Erwartung der Messung (Ende ÷ *S* lag bei den Parser-Slices zwischen 0,69 und 0,78, die Formel überschätzt, `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`): rund 1800, Spanne 1500 bis 2400. Ändert der Architect die Zeilen der Tabelle oder `produktcode`, setzt der Planner Schätzung und Marke **vor** dem ersten Code-Commit neu (`ersetzt` mit Grund, `SPEC-128` Nr. 3 n).
- **Schnittfrage** (*S* ≥ 2223, hier 2555): gestellt und vor dem Code in zwei Schritten beantwortet. (1) Abgetrennt ist `slice-expand-validator` (Validator gegen die Navigationen des Modells): er teilt mit dem Parser allein den Knoten (`ExpandOption`, rund 25 Zeilen, hier geliefert) und hat sonst eigene Fehlerwerte, eigene Grammatik der Prüfung und eine eigene Eingabe (`ServiceModel`); das Vorbild ist die Trennung von `slice-odata-url-filter` (Leser) und `slice-edm-ausdruecke` (Typprüfung `SPEC-167`). Ungeschnitten wäre *S* mit rund 40 Zeilen rund 3700 (Marke 5000). (2) Der Rest lässt sich nicht weiter abtrennen: Dekodierung, Namensform, Klassifikation der Elemente (Name · `*` · Pfad · Klammer), Rangfolge, Grenzen und Fehlerwerte sind **ein** Mechanismus; die Einhängestelle und der Knoten sind seine Eingabe und Ausgabe, kein Teil, den ein eigener Commit vorab liefern könnte. *S* liegt um 332 über der Schnittgrenze (vier Zeilen); eine Zeile weniger je Zusage ist kein Schnitt.

```groessenmarke
pfad: src/adapters/driven/odata/urlparser/
pfad: src/hexagon/model/url_ast.go
pfad: src/hexagon/model/url_ast_expand.go
pfad: src/hexagon/model/url_ast_expand_test.go
produktcode: 150
fakes: 0
basis: d158068de091b91c6327c509da4a0833991d893b
schaetzung: 2555
marke: 3500
schnitt-grund: Der Validator ist als slice-expand-validator abgetrennt (teilt nur den Knoten). Der Rest teilt Dekodierung, Namensform, Klassifikation der Elemente, Rangfolge, Grenzen und Fehlerwerte in einem Mechanismus; Einhängestelle und Knoten sind dessen Ein- und Ausgabe, kein abtrennbarer Teil.
```

`make groessenmarke` am Stand vor dem Code (2026-11-18, `HEAD` d158068d): `groessenmarke slice-expand-parser.md: ok, Messung 0 von Marke 3500`. Messpunkte (`SPEC-128` Nr. 9; Handoff und Ende trägt der Planner beim Handoff und bei der Closure nach):

- Start: `Messpunkt slice-expand-parser.md: zeilen=30 nebenlaeufig=1 fakes=0 produktcode=150 schaetzung=2555 marke=3500 messung=0`

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 3500 Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Eine Überschreitung schneidet nicht (`SPEC-128` Nr. 10 c); sie bekommt einen Eintrag `ueberschreitung` mit Grund im Block.
- `in-progress` → `open` (blockiert — Carveout?): der Umfang von `$expand` (§6 A1 bis A3) ist nicht entschieden, oder `slice-odata-url-pfad` hat die Form des Query-AST anders geschnitten, als dieser Plan annimmt (am 2026-11-18 geprüft: `model.Optionen` trägt je Systemoption ein Feld, `urlparser.New(opts ...Option)` nimmt Optionen entgegen).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Menge der Optionstexte ist unbeschränkt (jede Zeichenfolge) — **Ausgang:** offen; Prüfgrenze Länge 4 über 17 Zeichen und benannte Familien darüber, Folge im Betrieb als akzeptiertes Negativ in Nr. 24 der Tabelle (`AGENTS.md` §3.17).
- Schnitt: `welle-expand-batch` §4 und die Roadmap führen vier Slices, es sind fünf (`slice-expand-validator` ist neu); der Closure-Trigger der Welle verlangt „je Funktion ein Parser-, Validator-, Executor- und Serializer-Contract-Test“ — **Ausgang:** offen; Adresse Planner (Integrationsschritt: Welle-Plan §2 und §4, Roadmap; dieser Zweig schreibt beides nicht, `docs/plan/planning/README.md` Übergangsregeln Nr. 4).
- Sendung an `slice-expand-ausfuehrung`: Einschalten der Option im Aufbau, Aufruf des Validators und Änderung von `ODV4-E-03` c (`SPEC-042`, `abnahme_matrix_test.go`) — **Ausgang:** weiter offen; dessen §1 und DoD tragen die Sendung seit diesem Commit (`AGENTS.md` §3.13, Geber `slice-expand-parser`).
- Fakes der Ports: der Slice hat keinen Port-Fake (`f` = 0); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) trifft ihn nicht — **Ausgang:** entfallen, der Architect nimmt keinen Port auf (`SPEC-262`).
- Register gesichtet am 2026-11-18, Zähler aus den Dateien unter `observations/BEO-REPO/`: `menge-ohne-ihre-positionen` (3×) und `menge-im-plan-weicht-vom-orakel-des-tests-ab` (3×) für Nr. 8, 11, 12, 24 (Position und Orakel aus dem Wortlaut, §3); `suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) für Nr. 24; `fixture-gleicht-dem-wert-der-mutation` (7×) für Nr. 4 und 8; `betriebsregel-fuer-repo-dateien-verletzt` (18×); `randform-erst-im-review-entschieden` (36×) für das Raster. Die Einträge mit 3× und mehr sind verkörpert oder warten auf die Welle-Closure; hier wird nichts gezählt — **Ausgang:** weiter offen.
- Umsetzung: die Tabelle der Abnahme (`abnahme_matrix_test.go`) baut ihren Dienst selbst, und die Zeile des Aufbaus in `wurzel.go` hatte keinen Fänger für die Option (Mutation grün gesehen) — **Ausgang:** entschieden, `TestSatzVersion4OhneExpand` fängt sie (Mutation rot gesehen, Zeile 28); die Anpassung dieses Tests bei der Verdrahtung trägt `slice-expand-ausfuehrung` in seiner DoD (Geber dieser Slice, `AGENTS.md` §3.13).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten hat für jede Zelle einen Ort in `SPEC-262` (A1 bis A9 sind entschieden); der Implementer hält an einer Zelle ohne Ort an — **Ausgang:** entschieden (Raster unten, `SPEC-262`).

**Entscheidungen des Architect — Ausgang *entschieden*, Ort `SPEC-262` (Nummer der Linie B) und die genannten Stellen von `SPEC-160` und `SPEC-162`; der Validator steht in `SPEC-263`** (`AGENTS.md` §3.13). Keine ADR: die Entscheidungen schärfen die Spezifikation, das Lastenheft und keine angenommene ADR ändern sich (`AGENTS.md` §3.8); eine verworfene Alternative steht mit Grund, wo es eine gab.

- **A1 Umfang.** Entschieden (`SPEC-262` Nr. 3 b, Nr. 4): eine Liste einfacher Namen; `*`, Pfad (Segmente: Name, qualifizierter Name, `$ref`, `$count`, `$value`) und jeder Wert mit `(` (verschachtelte Optionen, `$levels`) sind lesbar und `501`; `$filter` in Klammern ist nicht im Umfang. Grund: `LH-FA-16.a` verlangt `501` für unbekannte Operationen, und ein Leser der Optionen in Klammern wäre ein eigener Slice. Zeilenzahl der Tabelle und Marke bleiben.
- **A2 Einhängestelle.** Entschieden (`SPEC-262` Nr. 1): `New(WithExpand())`; ohne sie bleibt `$expand` `option_nicht_unterstuetzt` (Binary, `ODV4-E-03` c, `SPEC-160` Nr. 8 und `SPEC-049` unverändert). Verworfen: immer lesbar — der Dienst müsste jede gültige Form selbst mit `501` abweisen, sonst entstünde eine Antwort ohne Expansion (`LH-RB-01`), und `ODV4-E-03` c änderte sich in diesem Slice.
- **A3 Grammatik.** Entschieden (`SPEC-262` Nr. 3 und 4): Trenner `,` nach der Dekodierung (`%2C` trennt), kein Leerraum, `;` trennt nichts; Name nach `SPEC-125` Nr. 8 b (Unicode-Buchstaben zählen, 128 Zeichen); ein `(` irgendwo im Wert macht den ganzen Wert `501` und er wird nicht weiter gelesen (wie `$orderby`, `SPEC-162` Nr. 7), weil ein `,` in der Klammer sonst falsch zerlegt würde; ein `)` ohne `(` ist `wert`; in einem Pfad sind `$ref`, `$count`, `$value` und qualifizierte Namen `501`, `*`, `$levels` und ein leeres Segment `wert`, ein qualifizierter Name oder `$ref` allein `wert`.
- **A4 Grenzen.** Entschieden (`SPEC-262` Nr. 6, `SPEC-160` Nr. 6): 32 Elemente (`zu_viele`, Ort `query/<n>`), Name 128 Zeichen. Grund: jedes Element löst der Dienst gegen Daten auf, `$orderby` hat dieselbe Grenze. Folge im Betrieb: mehr als 32 Navigationen ergeben `400`.
- **A5 Doppeltes Element.** Entschieden (`SPEC-262` Nr. 4 e): `wert`, wie bei `$select`; die Schreibweise zählt; die Regel gilt auch für `*` und Pfade.
- **A6 Fehlerwerte.** Entschieden (`SPEC-262`, Kopf): nur die bestehenden Ursachen; `AlleURLUrsachen` bleibt bei sechzehn, `SPEC-163` und das Handbuch bleiben unverändert.
- **A7 Rangfolge im Wert.** Entschieden (`SPEC-262` Nr. 3, 8): `(` zuerst (`501`, nichts weiter gelesen), dann `zu_viele` vor den Elementen, `ungueltig` vor `nicht_unterstuetzt`, innerhalb der Klasse das erste Element von links. Akzeptiertes Negativ gegen `SPEC-160` Nr. 5: ein ungültiges Element in einem Wert mit `(` wird als `501` gemeldet; Grund und Folge stehen in `SPEC-262` Nr. 3 b.
- **A8 Form des Knotens.** Entschieden (`SPEC-262` Kopf und Nr. 5, `SPEC-160` Typen): `Optionen.Expand *ExpandOption{Elemente []ExpandElement}`, `ExpandElement{Name string}`, als **letztes** Feld von `Optionen` (kein bestehendes Feld verschiebt sich); `nil` heißt nicht genannt, genannt heißt mindestens ein Element in der Reihenfolge der URL, Kopie nach `SPEC-160` Nr. 3. `slice-expand-validator` liest `Elemente[i].Name` (`SPEC-263`).
- **A9 Anwendbarkeit je Ressourcenart.** Entschieden (`SPEC-262` Nr. 7): der Parser liest `$expand` an jeder Ressource, die Wirkung entscheidet der Dienst (`SPEC-160` *Ungeprüft*).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; „A<n>“ heißt: Ort ist die offene Entscheidung des Architect in der Spezifikation, Adresse Architect; „SPEC-160 …“ nennt einen Ort, dessen Wortlaut die Entscheidung trägt):

| Operation | Option fehlt | leerer Wert; leeres Element | Name ungültig, Länge, Dekodierung | `*`, Pfad, `$ref`, Typ-Cast | Klammer (verschachtelt) | Wiederholung (Option, Element) | Zahl der Elemente | unbekannte oder fremde Navigation | Abbruch (Kontext) |
|---|---|---|---|---|---|---|---|---|---|
| `ParseURL` ohne eingehängte Option | `SPEC-160` Nr. 2 (Nullwert des Feldes) | `SPEC-162` Nr. 3 b (jeder Wert, auch leer, ungelesen) | `SPEC-162` Nr. 2 und 3 b (Wert ungelesen) | `SPEC-162` Nr. 3 b | `SPEC-162` Nr. 3 b | `SPEC-162` Nr. 3 b („beliebig oft“) | gilt nicht: der Wert wird nicht gelesen (`SPEC-162` Nr. 2 Schluss) | gilt nicht: der Parser kennt das Modell nicht (`SPEC-160` Absatz 3, Nr. 9); Zelle des Validators in `slice-expand-validator` | gilt nicht: die Signatur `ParseURL(pfad, rawQuery)` hat keinen Kontext (`SPEC-160`) |
| `ParseURL` mit eingehängter Option | `SPEC-262` Nr. 5 (`nil`) | `SPEC-262` Nr. 3 a und Nr. 4 d | `SPEC-262` Nr. 4 a und d, Nr. 6; Dekodierung `SPEC-162` Nr. 2 | `SPEC-262` Nr. 4 b und c | `SPEC-262` Nr. 3 b | `SPEC-262` Nr. 2 (Option) und Nr. 4 e (Element) | `SPEC-262` Nr. 3 c und Nr. 6 | gilt nicht: wie oben | gilt nicht: wie oben |

Zusätzlich zu A9: je Ressourcenart (`/$metadata`, `/Set`, `/Set(Schlüssel)`, `/Set/$count`) liest der Parser dieselbe Option (`SPEC-262` Nr. 7), Fänger Nr. 3 und 26.

**Zusatztabelle: mehrere Verstöße am selben Wert** (A7, `SPEC-262` Nr. 3; je Paar das Ergebnis der Rangfolge; Fänger für alle Zeilen: Nr. 16 und 17 der Tabelle in §3):

| Paar (links, rechts im Wert) | Ergebnis | Fänger |
|---|---|---|
| leeres Element, `*` | `wert` (Klasse `ungueltig` vor `nicht_unterstuetzt`) | `TestExpandRangImWert`, beide Reihenfolgen |
| Name ungültig, Pfad | `wert` | wie oben |
| `*`, Pfad | `wert_nicht_unterstuetzt` (erstes der Klasse) | wie oben |
| `(` im Wert, ungültiges Element (leer, ungültiger Name, doppelt) | `wert_nicht_unterstuetzt` (die Klammer geht vor, der Rest wird nicht gelesen, `SPEC-262` Nr. 3 b) | `TestExpandKlammer` und `TestExpandRangImWert` |
| `*`, `*` (doppeltes Element) | `wert` (Klasse `ungueltig` vor `nicht_unterstuetzt`) | `TestExpandDoppeltesElement` |
| zu viele Elemente, leeres Element | `zu_viele` (Stufe vor den Elementen) | `TestExpandZahl` |
| `$expand` doppelt, `$expand` mit `*` | `option_doppelt` (`ungueltig` vor `nicht_unterstuetzt`, gleich wo es steht) | `TestExpandDoppelt` |
| Verstoß in `$expand`, Verstoß in `$top` | der erste der Klasse in der Lesefolge, `ungueltig` zuerst | `TestExpandRangQuery` |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register am 2026-11-18 gesichtet (Anlage 2026-10-30); die Treffer mit Zähler stehen in §6. Der Slice ist bis auf Nr. 23 nicht nebenläufig. Keiner der offenen Einträge erreicht mit diesem Slice 3×, soweit er nicht schon verkörpert ist; das Beobachtungs-Register schreibt der Integrationsschritt.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
