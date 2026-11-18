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

**Berührte Spec-Stellen:** `LH-FA-16.a` (Erweiterungspunkt `$expand`, eigener Parser-Contract-Test, unbekannte Operationen `501`) · `SPEC-160` (Nr. 2 Ergebnis, Nr. 3 Veränderlichkeit, Nr. 4 Fehler, Nr. 5 Rangfolge, Nr. 6 Grenzen, Nr. 7 Nebenläufigkeit, Nr. 8 Schlusssatz: Erweiterungsweg für `$expand`) · `SPEC-162` (Nr. 2, Nr. 3 b: `$expand` ist heute nicht unterstützt; Nr. 6 und 7 als Vorbild der Liste) · `SPEC-163` (`ODR-E6005`) · `SPEC-022` (`ODR-E6001`) · `SPEC-049` (Zeile GET: Navigation-Expand, Spalte „Nicht in MVP“) · `SPEC-042` (Fixture `ODV4-E-03` c bleibt unverändert) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkte: separate AST-Knoten, Validatoren und Executors).
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

**Ziel:** Der Parser der OData-URL (`urlparser`, Port `URLCodec`, `SPEC-160`) liest die Systemoption `$expand` aus `raw_query` und überführt sie in einen eigenen Knoten des Query-AST (`model.Optionen`, `src/hexagon/model/url_ast.go`). Er liest nur Text und kennt das Modell nicht (`SPEC-160` Nr. 9): eine Liste im unterstützten Umfang ergibt den Knoten, eine unlesbare Form `400` (`ODR-E6005`), eine lesbare, aber nicht unterstützte Form `501` (`ODR-E6001`), jeweils mit den bestehenden Ursachen von `SPEC-160` Nr. 4 (Vorschlag, §6 A6). Die Option ist **zuschaltbar** (`New` mit einer Option, `SPEC-160` Nr. 8 Schlusssatz): ohne sie bleibt `$expand` `option_nicht_unterstuetzt`, das Binary, `ODV4-E-03` c (`GET /Products?$expand=Foo` ist `501`) und die Zeile `abnahme_matrix_test.go` ändern sich nicht. Der Slice greift auf keine Daten und kein Modell zu (`LH-FA-16.a`: Parser-Contract-Test).

**Schnitt vor dem Code (§4):** Der Validator gegen die Navigationen des Modells (`model.PruefeExpand`, Muster von `SPEC-167`/`PruefeFilter`) ist aus diesem Slice herausgenommen und gehört `slice-expand-validator`. Der Plan führte beide in einem Slice; die Schnittfrage von `SPEC-128` Nr. 1 gilt dafür (§4). Der Validator ist keine Aufgabe des Parsers: `SPEC-160` sagt, der Parser kenne das Modell nicht, und `ARC-011` lässt den Adapter nur Modell und Ports importieren.

**Vorschlag des Planner zum Umfang (Entscheidung beim Architect, §6 A1 bis A9):** `$expand=Name1,Name2` — eine Liste einfacher Namen (Namensform wie `$select`, `SPEC-162` Nr. 6) ohne Leerraum; `*`, Pfade (`a/b`), `$ref`, Typ-Cast, `$levels` und jede verschachtelte Option in Klammern sind lesbar und `501` (`wert_nicht_unterstuetzt`, wie die Pfade von `$select`). Damit hängt der Slice weder an `slice-odata-url-filter` noch an `slice-edm-ausdruecke` (beide in `done/`; die Voraussetzung entfällt, solange der Architect `$filter` in Klammern nicht aufnimmt).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Validator gegen das Modell (Navigation vorhanden, auch geerbt; Property ohne Navigation; Subtyp) — `slice-expand-validator` (Schnitt, §4); dessen §1 führt den Punkt mit der Kennung dieses Slice als Geber.
- Verdrahtung: das Einschalten der Option im Aufbau des Dienstes, der Aufruf des Validators vor der Ausführung und die Änderung von `ODV4-E-03` c — `slice-expand-ausfuehrung` (dessen §1 trägt die Sendung, `AGENTS.md` §3.13); bis dahin liest das Binary `$expand` nicht und antwortet `501`.
- Auflösen der Navigation gegen Daten, Datenbasis der Navigationen und Ausgabe der eingebetteten Entities — `slice-expand-ausfuehrung` (Executor- und Serializer-Contract-Test von `LH-FA-16.a` liegen dort); dieser Slice liefert den Knoten.
- Verschachtelte Optionen im `$expand` (`$filter`, `$select`, `$orderby`, `$top`, `$skip`, `$count`, `$levels`, `$expand`) — Entscheidung des Architect (§6 A1); was er nicht nennt, ergibt `501`.
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
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (Vorschlag A6: keine neue Ursache, kein neuer Code; der Modus `mock` erreicht das Binary erst mit `slice-bootstrap-m3`, die Antwort auf `$expand` zeigt sich erst mit `slice-expand-ausfuehrung`).
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
| `src/hexagon/model/url_ast.go`, `src/hexagon/model/url_ast_expand.go` | update, neu | Knoten für `$expand` im Query-AST, technologieunabhängig; Feld in `Optionen` (Form nach Entscheidung A8) |
| `src/adapters/driven/odata/urlparser/` | update | Lesen der Option (`query.go`, neue Datei `expand.go`), Einhängestelle in `New`; `$expand` verlässt `nichtUnterstuetzteOption` nur mit eingehängter Option |
| `spec/spezifikation.md` | update | Vertrag mit Randformen, Beispiele (Architect): `SPEC-160` Nr. 2 und 8, `SPEC-162` Nr. 3 b, neue Nummer aus dem Bereich der Linie B |
| Tests | neu | `urlparser/expand_*_test.go` (Fälle, Alphabet-Generator mit Orakel, Rangfolge, Parser-Contract, Nebenläufigkeit), `model/url_ast_expand_test.go` |

`src/hexagon/ports/outbound/` bleibt unverändert: die Signatur von `URLCodec.ParseURL` und die Felder von `SPEC-160` ändern sich nicht (`SPEC-160` Nr. 8 Schlusssatz). Die Zeile des früheren Plans, die den Port „für den Validator“ erweiterte, entfällt mit dem Schnitt: der Validator ist eine reine Funktion in `model` (Muster `PruefeFilter`) und braucht keinen Port.

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen als Entwurf des Planner im Plan; der Architect bestätigt oder ändert die Zeilen mit dem Vertrag in der Spezifikation **vor** dem ersten Code-Commit (`AGENTS.md` §3.12). Ändert er den Umfang (§6 A1), setzt der Planner die Marke (§4) vor dem Code neu. Die Randformen entscheidet der Architect, nicht der Implementer.

- Der Parser liest fremde Bytes (den Text der URL): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); Optionstexte mit Escape-Folgen und Zeichen jenseits von ASCII entstehen aus Bytewerten (`AGENTS.md` §3.18).
- Orakel des Generators: der Test baut die erwartete Klasse (Knoten · `wert` · `wert_nicht_unterstuetzt` · `zu_viele`) mit einer eigenen, kurzen Tabelle aus dem Wortlaut der Spezifikation und ruft keinen Code des Lesers (`BEO-REPO/menge-im-plan-weicht-vom-orakel-des-tests-ab`, `BEO-REPO/menge-ohne-ihre-positionen`). Alphabet und Prüfgrenze: die Zeichen `a`, `Z`, `_`, `0`, `,`, `/`, `*`, `(`, `)`, `$`, `.`, `;`, `=`, `%20`, `%2C` und zwei Zeichen jenseits von ASCII aus Bytewerten, Folgen bis Länge 4 vollständig; längere Werte nur als benannte Familien (Grenzen, Wiederholung). Das ist das akzeptierte Negativ der Klasse „jede Zeichenfolge“ (Nr. 24).

**Tabelle *Zusage · Fänger*** (Entwurf des Planner für den Vorschlag des Umfangs; `AGENTS.md` §3.15; Dateien liegen unter `src/adapters/driven/odata/urlparser/`, Mutationen fährt der Implementer vor der Übergabe selbst):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Ohne eingehängte Option bleibt `$expand` mit jedem Wert (leer, ungültig kodiert, mehrfach) `option_nicht_unterstuetzt`, der Wert bleibt ungelesen | `TestExpandOhneEinhaengung`: Schleife über Werte (leer, `%G1`, Alphabet bis Länge 3) und Wiederholungen 1 bis 3; die bestehenden Fälle `query_test.go` (`$expand=a&$expand=b`) und `rang_test.go` bleiben unverändert grün | Die Vorgabe von `New` liest die Option | andere |
| 2 | Mit eingehängter Option ergibt eine gültige Liste `Optionen.Expand` mit den Namen in der Reihenfolge der URL | `TestExpandListe`: Listen der Länge 1 bis zur Grenze aus vier Namen, alle Reihenfolgen der Länge 3 | Die Elemente werden sortiert | andere |
| 3 | Fehlt `$expand`, ist `Expand` der Nullwert und jedes andere Feld gleich dem Ergebnis ohne eingehängte Option | `TestExpandDifferenz`: jede Query der bestehenden Tabellen (`query_test.go`, `rang_test.go`) gegen `New()` und `New(Expand)`, tiefe Gleichheit | Die Einhängung ändert ein anderes Feld (etwa `Sonstige`) | andere |
| 4 | Name und Wert der Option werden genau einmal dekodiert; `%24expand` und die Schreibweisen der Hex-Ziffern gelten wie das rohe Zeichen | `TestExpandDekodierung`: Schleife über Escape-Formen des Namens (roh, `%24`, Groß- und Kleinschreibung der Hex-Ziffern) und des Werts (`%2C` als Trenner, `%252C` bleibt `%2C` und ist `wert`); Fixture aus Bytewerten | Der Wert wird zweimal dekodiert | andere |
| 5 | Eine zweite Option `$expand` ist `option_doppelt`, in jeder Schreibweise des Namens, mit gleichem und mit anderem Wert, auch wenn der zweite Wert ungültig ist | `TestExpandDoppelt`: Paare aus {roh, `%24`} × {gleicher, anderer, ungültiger Wert} × Reihenfolge | Das zweite Vorkommen überschreibt das erste | andere |
| 6 | Ein leerer Wert (`$expand=`) ist `wert` | `TestExpandLeer`: `$expand=` und `$expand` ohne `=` | Leer ergibt den Nullwert ohne Fehler | andere |
| 7 | Ein leeres Element an jeder Stelle (`,a`, `a,`, `a,,b`, `,`) ist `wert` | `TestExpandLeeresElement`: Schleife über die Stelle (Anfang, Mitte, Ende) und die Länge der Liste 1 bis 4 | Leere Elemente werden übergangen | andere |
| 8 | Leerraum und jedes Zeichen außerhalb der Namensform an jeder Stelle des Namens (Anfang, Mitte, Ende) ist `wert`; kein Zeichen jenseits von ASCII ist Teil eines Namens | `TestExpandNamenform`: Schleife über alle 128 ASCII-Zeichen (roh und als `%XX`) und drei Bytefolgen jenseits von ASCII (NBSP, BOM, zwei Bytes `0xC3 0xA9`) × drei Stellen; Orakel: Zeichenklasse aus `SPEC-125` Nr. 8 b | Die Prüfung der Namensform prüft nur das erste Zeichen | andere |
| 9 | Ein Name ist 1 bis 128 Zeichen lang, beginnt mit Buchstabe oder `_`; 129 Zeichen und eine Ziffer am Anfang sind `wert` | `TestExpandNamenGrenze`: beide Seiten der Grenze (1, 128, 129, 0) × erstes Zeichen {Buchstabe, `_`, Ziffer} | Grenze 129 statt 128 | andere |
| 10 | `*`, allein oder neben Namen an jeder Position, ist `wert_nicht_unterstuetzt` | `TestExpandStern`: Schleife über Position (vor, nach, zwischen Namen) und Zahl der Namen 0 bis 3; `**` und `*a` gesondert | `*` wird als Name gelesen | andere |
| 11 | Ein Pfad aus mindestens zwei einfachen Namen (`a/b`, `a/b/c`) ist `wert_nicht_unterstuetzt`; ein leeres Segment (`/a`, `a/`, `a//b`) ist `wert` | `TestExpandPfad`: Segmentzahl 2 bis 5, leeres Segment an jeder Stelle | Ein Pfad wird als erster Name gelesen | andere |
| 12 | Ein Segment, das `$ref`, `$levels`, `$count`, `$value` oder einen qualifizierten Namen (Typ-Cast) trägt, an jeder Stelle des Pfads, ist `wert_nicht_unterstuetzt` oder `wert` nach der Entscheidung A3 | `TestExpandBesondereSegmente`: Schleife über fünf Segmentnamen × Stelle {allein, erstes, letztes Segment}; Orakel: Tabelle der Entscheidung | Ein Segment mit `$` wird als Name gelesen | andere |
| 13 | Eine Klammer im Element macht die Form zu einer verschachtelten Option: `wert_nicht_unterstuetzt`, der Inhalt wird nicht weiter gelesen; ein `)` ohne `(` ist `wert` | `TestExpandKlammer`: Schleife über jede Systemoption (`$filter`, `$select`, `$orderby`, `$top`, `$skip`, `$count`, `$levels`, `$expand`, `$search`) in `a(...)` und die Formen `a(`, `(`, `a()`, `a)`, `)a(` | Klammer-Inhalt wird auf Namen geprüft | andere |
| 14 | Mehr als 32 Elemente sind `zu_viele`, vor der Prüfung der Elemente; 32 sind gültig | `TestExpandZahl`: 31, 32, 33 Elemente, das 33. Element ungültig, das erste ungültig; Ort `query/<n>` | Grenze 33; Prüfung der Elemente vor der Zahl | andere |
| 15 | Ein Name zweimal (`a,a`, nicht benachbart: `a,b,a`) ist `wert`; die Schreibweise zählt (`a,A` sind zwei) | `TestExpandDoppeltesElement`: Abstand 0 bis 3 zwischen den Vorkommen, Groß- und Kleinschreibung | Duplikate werden zusammengeführt; Vergleich ohne Beachtung der Schreibweise | andere |
| 16 | Mehrere Verstöße im Wert: `zu_viele` vor den Elementen; Klasse `ungueltig` vor `nicht_unterstuetzt`; innerhalb der Klasse das erste Element von links | `TestExpandRangImWert`: alle Paare aus {leer, Name ungültig, doppelt, `*`, Pfad, Klammer} an den Stellen i und j (i<j und i>j), erwartete Ursache aus der Tabelle der Zusatztabelle in §6 | Die Klassen werden nach Position statt nach Klasse entschieden | andere |
| 17 | Rangfolge über Optionen: ein Verstoß in `$expand` und ein Verstoß einer anderen Option folgen `SPEC-160` Nr. 5, in beiden Reihenfolgen der Query | `TestExpandRangQuery`: Paare aus {Verstoß von `$expand`} × {`$top=-1`, `$foo=1`, `$select=a,,b`, `$filter` ohne Leser} × {links, rechts}; `rang_test.go` bleibt das Muster | `$expand` wird vor den anderen Optionen entschieden | andere |
| 18 | Der Ort ist `query/<n>` der Option (`n` zählt die nichtleeren Segmente), ohne Teil der Eingabe | `TestExpandOrt`: leere Segmente (`&&`) vor der Option, Option an Stelle 1 bis 4 | `n` zählt die leeren Segmente mit | andere |
| 19 | Kein Teil der Eingabe steht im Fehler: weder Name noch Wert noch Position in Bytes | `TestExpandFehlertext`: Schleife über alle Verstoßarten aus Nr. 5 bis 16, Eingabe mit unverwechselbarem Namen; `Error()` enthält ihn nicht | Der Name steht in der Meldung | andere |
| 20 | Das Ergebnis gehört dem Aufrufer: eine Änderung der Elemente ändert kein späteres Ergebnis, kein Slice wird zwischen Ergebnissen geteilt | `TestExpandKopie`: zweimal gleiche Eingabe, erstes Ergebnis ändern (Element, Namen), zweites vergleichen; Schleife über Listenlängen 1 bis 3 | Ein Paket-Slice wird als Ergebnis geliefert | andere |
| 21 | Gleiche Eingabe ergibt gleiches Ergebnis; ein Fehler gibt `ODataURL{}` | `TestExpandWiederholung`: jede Eingabe der Nr. 2 bis 16 zweimal, tiefe Gleichheit; Fehlerfälle gegen den Nullwert | Ein Zähler im Parser färbt das zweite Ergebnis | andere |
| 22 | Die Summe der Längen aller Texte im Ergebnis bleibt höchstens Eingabe plus 48 (`SPEC-160` Nr. 6 b) | `TestExpandWaechter`: Generator über Listen aus Namen der Länge 1 und 128, Länge der Eingabe bis zur Grenze von 8192 Bytes | Das Element trägt eine Kopie des ganzen Werts | andere |
| 23 | Gleichzeitige Aufrufe auf einer Instanz und auf mehreren beeinflussen sich nicht; jedes Ergebnis gleicht dem eines einzelnen Aufrufs (`SPEC-160` Nr. 7) | `TestExpandGleichzeitig`: 4 Goroutinen × 50 Runden × alle Fälle der Nr. 2 und 16, Ergebnis gegen Einzelaufruf (Wirkungs-Fänger); `make test-race` je Prozess für ein geteiltes Feld | Ein Puffer im Parser wird zwischen Aufrufen geteilt | nebenläufig |
| 24 | Jede Zeichenfolge über dem Alphabet der Grammatik bis zur Prüfgrenze (Länge 4, §3) hat die Klasse des Orakels: Knoten, `wert`, `wert_nicht_unterstuetzt` oder `zu_viele` | `TestExpandAlphabet`: vollständige Aufzählung aller Folgen bis Länge 4 über 17 Zeichen (88 740 Folgen), Orakel ohne Code des Lesers; akzeptiertes Negativ: längere Folgen nur als Familien (Nr. 9, 14, 22), Folge: eine Regel, die erst ab Länge 5 greift, bleibt ungefangen | Der Leser behandelt `;` oder `=` im Element als Trenner | andere |
| 25 | Der Parser liest kein Modell und keinen Dienst: die Richtung der Importe bleibt | `make a-check` (Regeln des Hexagons, Ausgabe grün); ein Import von `model.ServiceModel` oder eines Dienstes im Paket wird rot | `urlparser` importiert `services/mock` | andere |
| 26 | Die Beispiele der Spezifikation (Tabelle der Gültigen, der `400` und der `501`) ergeben die genannte Antwort: eigener Parser-Contract-Test | `TestExpandContract`: Tabelle der Beispiele als Fixture (Eingabe, Knoten oder Ursache und Ort), ein Fall je Zeile; keine Eingabe aus dem Generator | Die Antwort einer Zeile wird vertauscht | andere |
| 27 | Der Knoten trägt die Felder des Vertrags, jedes andere Feld von `Optionen` bleibt unverändert (Name, Typ, Reihenfolge) | `TestOptionenFelder` in `model`: die Feldliste von `Optionen` und des Knotens per `reflect` gegen die Liste des Vertrags | Ein Feld von `Optionen` wird umbenannt oder entfällt | andere |
| 28 | Das Binary bleibt unverändert: `GET /Products?$expand=Foo` ist `501` mit `ODR-E6001` (`ODV4-E-03` c) | die bestehende Zeile in `src/bootstrap/abnahme_matrix_test.go` (c) und `make test-contract`; kein neuer Test | Der Aufbau des Dienstes hängt die Option ein | andere |
| 29 | Akzeptiertes Negativ: die Länge des Werts hat keine eigene Grenze außer `rawQuery` (8192 Bytes) und der Zahl der Elemente | kein Fänger; Grund: die Summe der Längen ist über Nr. 6 b und die Grenze von 8192 begrenzt; Folge: ein Wert aus 32 Namen von 128 Zeichen liegt unter der Grenze der Query und ist gültig | — | andere |
| 30 | Akzeptiertes Negativ: ob `slice-expand-ausfuehrung` den Knoten ohne Änderung des Parsers liest, belegt erst dessen Slice | kein Fänger hier; Grund: der Leser des Knotens liegt nicht in diesem Slice; Folge: eine nötige Änderung der Form dort ist eine Änderung dieses Vertrags (`AGENTS.md` §3.13, Adresse Architect) | — | andere |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-odata-url-pfad`, `slice-odata-url-filter` und `slice-edm-ausdruecke` liegen in `done/` (am 2026-11-18 geprüft); die Voraussetzung für `$filter` in Klammern ist erfüllt, falls der Architect ihn aufnimmt. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-odata-url-pfad` (Typen des Query-AST, Paket `urlparser`, Zerlegung von `raw_query`, Fehlerbildung `400`/`501`; in `done/`). Das Modell mit den Navigation-Properties braucht dieser Slice **nicht** mehr (der Validator ist abgeschnitten). Der Welle-Trigger (`welle-mock` in `done/`) ist Reihenfolge-Vorsicht, keine Code-Abhängigkeit.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Query-Interpreter, JSON-Schreiber, Workspace und Composition Root (der Slice liest Text), M2 und `slice-bootstrap-m3` (das Binary ändert sich erst mit der Verdrahtung), `slice-expand-validator` (liest den Knoten dieses Slice, nicht umgekehrt), `slice-expand-ausfuehrung`, `slice-batch-parser`.
- **Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`; die Faktoren stehen allein in `SPEC-128` Nr. 1): gezählt ist die Tabelle in §3 mit 30 Zeilen, *a* = 1 nebenläufige (Zeile 23), *b* = 29 andere (die beiden akzeptierten Negative zählen als `andere`, `SPEC-128` Nr. 3 e), *f* = 0 (der Slice hat keinen Port-Fake und keinen Test-Server), *p* = 150 (Erwartung des Planner für den Architect, Spanne 100 bis 250: Lesen der Liste rund 70 Zeilen, Knoten und Feld rund 40, Einhängestelle und Verdrahtung in `query.go` rund 40). *S* = 85·1 + 80·29 + 200·0 + 150 = **2555**, 135·2555 = 344925, Marke *M* = 100 · ⌈344925 / 10000⌉ = **3500**. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks (das Paket `urlparser`, der Knoten in `src/hexagon/model/`), Tests eingeschlossen. `basis` ist `HEAD` vor dem Vertrag des Architect: er ändert nur `spec/`, das kein Pfad des Blocks ist (`SPEC-128` Nr. 3 m), die Messung bleibt bis zum ersten Code-Commit 0. Erwartung der Messung (Ende ÷ *S* lag bei den Parser-Slices zwischen 0,69 und 0,78, die Formel überschätzt, `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`): rund 1800, Spanne 1500 bis 2400. Ändert der Architect die Zeilen der Tabelle oder `produktcode`, setzt der Planner Schätzung und Marke **vor** dem ersten Code-Commit neu (`ersetzt` mit Grund, `SPEC-128` Nr. 3 n).
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
- Fakes der Ports: der Slice hat keinen Port-Fake (`f` = 0); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) trifft ihn nicht — **Ausgang:** entfallen, solange der Architect keinen Port aufnimmt.
- Register gesichtet am 2026-11-18, Zähler aus den Dateien unter `observations/BEO-REPO/`: `menge-ohne-ihre-positionen` (3×) und `menge-im-plan-weicht-vom-orakel-des-tests-ab` (3×) für Nr. 8, 11, 12, 24 (Position und Orakel aus dem Wortlaut, §3); `suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) für Nr. 24; `fixture-gleicht-dem-wert-der-mutation` (7×) für Nr. 4 und 8; `betriebsregel-fuer-repo-dateien-verletzt` (18×); `randform-erst-im-review-entschieden` (36×) für das Raster. Die Einträge mit 3× und mehr sind verkörpert oder warten auf die Welle-Closure; hier wird nichts gezählt — **Ausgang:** weiter offen.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten hat für jede Zelle einen Ort oder die offene Entscheidung A1 bis A9 mit Adresse Architect; der Implementer hält an einer Zelle ohne Ort an — **Ausgang:** offen.

**Offene Entscheidungen — Ausgang *offen*, Adresse Architect** (`AGENTS.md` §3.13; der Planner entscheidet sie nicht, der Vorschlag steht dabei, damit die Tabelle in §3 und die Marke in §4 gezählt werden konnten; der Ort ist die neue Nummer des Architect im Bereich der Linie B und die genannten Stellen von `SPEC-160` und `SPEC-162`):

- **A1 Umfang.** Vorschlag: eine Liste einfacher Namen; `*`, Pfad, `$ref`, Typ-Cast, `$levels` und verschachtelte Optionen sind lesbar und `501`. Nennt der Architect `$filter` in Klammern oder `*`, wächst der Umfang (Zeilen der Tabelle, Marke).
- **A2 Einhängestelle.** Vorschlag: `New` mit einer Option ohne Parameter; ohne sie bleibt `$expand` `option_nicht_unterstuetzt` (Binary, `ODV4-E-03` c und `SPEC-160` Nr. 8 Schlusssatz unverändert). Alternative: immer lesbar; dann muss der Dienst jede gültige Form selbst mit `501` abweisen, sonst entstünde eine Antwort ohne Expansion (`LH-RB-01`), und `ODV4-E-03` c ändert sich in diesem Slice.
- **A3 Grammatik.** Trenner `,`, kein Leerraum, `;` trennt nichts; Namensform und 128 Zeichen wie `SPEC-162` Nr. 6; ein `(` in einem Element macht es zu einer verschachtelten Option (`wert_nicht_unterstuetzt`, wie `SPEC-162` Nr. 7), ein `)` ohne `(` ist `wert`; ein Segment `$ref`, `$levels`, `$count`, `$value` oder mit qualifiziertem Namen: `wert_nicht_unterstuetzt` oder `wert` (Nr. 12 der Tabelle).
- **A4 Grenzen.** Vorschlag: 32 Elemente (`zu_viele`, Ort `query/<n>`), Länge des Namens 128; Grund und Folge im Betrieb nennt der Architect.
- **A5 Doppeltes Element.** Vorschlag: `wert`, wie `$select` (`SPEC-162` Nr. 6); die Schreibweise zählt.
- **A6 Fehlerwerte.** Vorschlag: nur die bestehenden Ursachen (`wert`, `zu_viele`, `wert_nicht_unterstuetzt`, `option_doppelt`, `option_nicht_unterstuetzt`); `AlleURLUrsachen` bleibt bei sechzehn, `SPEC-163` und das Handbuch bleiben unverändert (DoD: Binary unverändert).
- **A7 Rangfolge im Wert.** Vorschlag: wie `SPEC-160` Nr. 5 und `SPEC-162` Nr. 6: `zu_viele` vor den Elementen, `ungueltig` vor `nicht_unterstuetzt`, innerhalb der Klasse das erste Element von links.
- **A8 Form des Knotens.** Vorschlag: `Optionen.Expand *ExpandOption{Elemente []ExpandElement}`, `ExpandElement{Name string}` (ein Struct statt `string`, damit verschachtelte Optionen später ein Feld ergänzen, ohne die Leser zu ändern); Reihenfolge der URL, Kopie nach `SPEC-160` Nr. 3; `nil` heißt nicht genannt. Er bestimmt, ob `slice-expand-ausfuehrung` und `slice-expand-validator` ohne Änderung des Parsers lesen.
- **A9 Anwendbarkeit je Ressourcenart.** Vorschlag: wie die anderen Optionen (`SPEC-160` *Ungeprüft*): der Parser liest `$expand` an jeder Ressource (auch `/$metadata`, `/Set/$count`), die Wirkung entscheidet der Dienst.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; „A<n>“ heißt: Ort ist die offene Entscheidung des Architect in der Spezifikation, Adresse Architect; „SPEC-160 …“ nennt einen Ort, dessen Wortlaut die Entscheidung trägt):

| Operation | Option fehlt | leerer Wert; leeres Element | Name ungültig, Länge, Dekodierung | `*`, Pfad, `$ref`, Typ-Cast | Klammer (verschachtelt) | Wiederholung (Option, Element) | Zahl der Elemente | unbekannte oder fremde Navigation | Abbruch (Kontext) |
|---|---|---|---|---|---|---|---|---|---|
| `ParseURL` ohne eingehängte Option | `SPEC-160` Nr. 2 (Nullwert des Feldes) | `SPEC-162` Nr. 3 b (jeder Wert, auch leer, ungelesen) | `SPEC-162` Nr. 2 und 3 b (Wert ungelesen) | `SPEC-162` Nr. 3 b | `SPEC-162` Nr. 3 b | `SPEC-162` Nr. 3 b („beliebig oft“) | gilt nicht: der Wert wird nicht gelesen (`SPEC-162` Nr. 2 Schluss) | gilt nicht: der Parser kennt das Modell nicht (`SPEC-160` Absatz 3, Nr. 9); Zelle des Validators in `slice-expand-validator` | gilt nicht: die Signatur `ParseURL(pfad, rawQuery)` hat keinen Kontext (`SPEC-160`) |
| `ParseURL` mit eingehängter Option | A8 (`SPEC-160` Nr. 2 trägt die anderen Felder) | A3 | A3, A4; Dekodierung `SPEC-162` Nr. 2 | A1, A3 | A1, A3 | A5; Option: A2 und `SPEC-162` Nr. 3 a | A4 | gilt nicht: wie oben | gilt nicht: wie oben |

Zusätzlich zu A9: je Ressourcenart (`/$metadata`, `/Set`, `/Set(Schlüssel)`, `/Set/$count`) liest der Parser dieselbe Option (Vorschlag), Fänger Nr. 3 und 26.

**Zusatztabelle: mehrere Verstöße am selben Wert** (Vorschlag A7, der Architect bestätigt; je Paar das Ergebnis der Rangfolge; Fänger für alle Zeilen: Nr. 16 und 17 der Tabelle in §3):

| Paar (links, rechts im Wert) | Ergebnis | Fänger |
|---|---|---|
| leeres Element, `*` | `wert` (Klasse `ungueltig` vor `nicht_unterstuetzt`) | `TestExpandRangImWert`, beide Reihenfolgen |
| Name ungültig, Pfad | `wert` | wie oben |
| `*`, Pfad | `wert_nicht_unterstuetzt` (erstes der Klasse) | wie oben |
| doppeltes Element, Klammer | `wert` | wie oben |
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
