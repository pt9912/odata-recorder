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
| 3 | Fehlt `$expand`, ist `Expand` der Nullwert und jedes andere Feld gleich dem Ergebnis ohne eingehängte Option; die Einhängestellen von `New` ändern einander nicht, in jeder Folge (`SPEC-262` Nr. 1: „ändert kein anderes Feld“) | `TestExpandOptionenFolgen`: alle Folgen der Länge 0 bis 3 über {`WithFilter(Leser)`, `WithFilter(nil)`, `WithExpand()`} (40) × 7 Queries (`$filter` und `$expand` allein und in beiden Reihenfolgen, mit fremden Optionen dazwischen, ohne beide), Orakel: die Folge bestimmt allein, ob `$filter` (letzte `WithFilter` mit Leser) und `$expand` (irgendeine `WithExpand`) lesbar sind, der erste Verstoß der Lesefolge gilt; `TestExpandDifferenz`: jede Query ohne `$expand` aus `korpus()` (`nebenlaeufig_test.go`: Gut- und Fehlerfälle aller Klassen, dazu 25 erzeugte) gegen `New()` und `New(WithExpand())`, tiefe Gleichheit samt Fehlertext, `Expand` ist `nil`; je Ressourcenart (`/$metadata`, `/S`, `/S(1)`, `/S/$count`) liest der Parser die Option; die Tabellen von `query_test.go` und `rang_test.go` sind lokale Werte der Tests und bleiben unverändert grün | Die Einhängung ändert ein anderes Feld (etwa `Sonstige`); `WithExpand` setzt `filter` zurück; `WithFilter` setzt `expand` zurück | andere |
| 4 | Name und Wert der Option werden genau einmal dekodiert; `%24expand` und die Schreibweisen der Hex-Ziffern gelten wie das rohe Zeichen | `TestExpandDekodierung`: Schleife über Escape-Formen des Namens (roh, `%24`, Groß- und Kleinschreibung der Hex-Ziffern) und des Werts (`%2C` als Trenner, `%252C` bleibt `%2C` und ist `wert`); Fixture aus Bytewerten | Der Wert wird zweimal dekodiert | andere |
| 5 | Eine zweite Option `$expand` ist `option_doppelt`, in jeder Schreibweise des Namens, mit gleichem und mit anderem Wert, auch wenn der zweite Wert ungültig ist | `TestExpandDoppelt`: Paare aus {roh, `%24`} × {gleicher, anderer, ungültiger Wert} × Reihenfolge | Das zweite Vorkommen überschreibt das erste | andere |
| 6 | Ein leerer Wert (`$expand=`) ist `wert` | `TestExpandLeer`: `$expand=` und `$expand` ohne `=` | Leer ergibt den Nullwert ohne Fehler | andere |
| 7 | Ein leeres Element an jeder Stelle (`,a`, `a,`, `a,,b`, `,`) ist `wert` | `TestExpandLeeresElement`: Schleife über die Stelle (Anfang, Mitte, Ende) und die Länge der Liste 1 bis 4 | Leere Elemente werden übergangen | andere |
| 8 | Leerraum und jedes Zeichen außerhalb der Namensform an jeder Stelle des Namens (Anfang, Mitte, Ende) ist `wert`; ein Zeichen jenseits von ASCII ist Teil eines Namens genau dann, wenn es ein Buchstabe (Kategorie `L`) oder, nicht am Anfang, eine Ziffer (`Nd`) ist (`SPEC-262` Nr. 4 a) | `TestExpandNamenform`: Schleife über alle 128 ASCII-Zeichen außer `,`, `/` und `(` (eigene Zeilen Nr. 7, 11, 13), roh und als `%XX`, und sieben Zeichen jenseits von ASCII aus Bytewerten (é `0xC3 0xA9`, Ж, 中 sind Buchstaben, ٣ ist eine Ziffer und gültig außer am Anfang; NBSP, BOM und ² sind `wert`) × drei Stellen × roh und als `%XX` in beiden Schreibweisen; Orakel: Zeichenklasse aus `SPEC-125` Nr. 8 b als eigene Tabelle, nicht aus dem Code des Lesers | Die Prüfung der Namensform prüft nur das erste Zeichen; jedes Zeichen jenseits von ASCII wird verworfen | andere |
| 9 | Ein Name ist 1 bis 128 Zeichen lang (Zeichen, nicht Bytes), beginnt mit Buchstabe oder `_`; 129 Zeichen und eine Ziffer am Anfang sind `wert` | `TestExpandNamenGrenze`: beide Seiten der Grenze (1, 128, 129, 0) × erstes Zeichen {Buchstabe, `_`, Ziffer} × Breite des Zeichens {ein Byte, zwei Bytes (`0xC3 0xA9`)} | Grenze 129 statt 128; Länge in Bytes statt in Zeichen | andere |
| 10 | `*`, allein oder neben Namen an jeder Position, ist `wert_nicht_unterstuetzt` | `TestExpandStern`: Schleife über Position (vor, nach, zwischen Namen) und Zahl der Namen 0 bis 3; `**` und `*a` gesondert | `*` wird als Name gelesen | andere |
| 11 | Ein Pfad aus mindestens zwei einfachen Namen (`a/b`, `a/b/c`) ist `wert_nicht_unterstuetzt`; ein leeres Segment (`/a`, `a/`, `a//b`) ist `wert` | `TestExpandPfad`: Segmentzahl 2 bis 9, leeres Segment an jeder Stelle; `TestExpandSegmentZahl`: Zahl der Segmente 1 bis 4090 (340 Stufen, Prüfgrenze unten) | Ein Pfad wird als erster Name gelesen | andere |
| 12 | Ein Segment eines Pfads ist ein Name, ein qualifizierter Name (Typ-Cast) oder `$ref`, `$count`, `$value`: der Pfad ist `wert_nicht_unterstuetzt`, an jeder Stelle; jedes andere Segment (`*`, `$levels`, `$`, `a.`, `.a`) macht das Element zu `wert`; ein qualifizierter Name oder `$ref` allein ist `wert` (`SPEC-262` Nr. 4 c, d) | `TestExpandBesondereSegmente`: Schleife über die Segmente {Name, `Ns.Typ`, `a.b.c`, `$ref`, `$count`, `$value`, `*`, `$levels`, `$`, `a.`, `.a`} × Stelle {erstes, mittleres, letztes Segment eines Pfads} und jedes Segment allein; Orakel: Tabelle des Vertrags. Die Menge der Segmente ist die Grammatik eines Segments, getrennt nach Zeichenmenge und Struktur (Orakel `exOrakel`, kein Code des Lesers): `TestExpandSegmentZeichen` (359 Zeichen × 24 Formen × 4 Stellen), `TestExpandSegmentStruktur` (alle Folgen der Länge 0 bis 4 über 11 Zeichen × 4 Stellen), `TestExpandSegmentWorte` (die Wörter mit `$` samt Nachbarn im Abstand eins, 5720 Schreibweisen × 4 Stellen), `TestExpandSegmentLaenge` (Länge je Teil eines qualifizierten Namens, 208); die Zahl der Teile, der Segmente und die Länge als Ganzes sind allein durch `rawQuery` begrenzt (`SPEC-262` Nr. 4 c): `TestExpandSegmentTeile` (107 440), `TestExpandSegmentZahl` (22 940), `TestExpandSegmentKreuz` (1644), `TestExpandSegmentLaengeGesamt` (936) | Ein Segment mit `$` wird als Name gelesen; `*` wird als Segment zugelassen; `-` im Segment; 129 Zeichen im Segment; eine Ziffer nach dem Punkt; `$REF`; `$filter` als Segment | andere |
| 13 | Ein `(` irgendwo im Wert macht den ganzen Wert zu `wert_nicht_unterstuetzt`, der Rest wird nicht gelesen: auch hinter einem leeren oder ungültigen Element und bei mehr als 32 Elementen; ein `)` ohne `(` ist `wert` (`SPEC-262` Nr. 3 b) | `TestExpandKlammer`: Schleife über jede Systemoption (`$filter`, `$select`, `$orderby`, `$top`, `$skip`, `$count`, `$levels`, `$expand`, `$search`) in `a(...)`, über die Stelle des `(` (Anfang, Mitte, Ende des Elements; Element 1 bis 3) und über die Vorbelastung davor {keine, leeres Element, ungültiger Name, doppeltes Element, 33 Elemente}; dazu die Formen `a(`, `(`, `a()`, `a)`, `)a(` und `a($select=Id,Name),b`; Ergebnis immer `wert_nicht_unterstuetzt`, außer `a)` (`wert`) | Das `(` wird erst nach der Zahl oder den Elementen gesucht | andere |
| 14 | Mehr als 32 Elemente sind `zu_viele`, vor der Prüfung der Elemente; 32 sind gültig; ein `(` im Wert geht der Zahl vor (Nr. 13) | `TestExpandZahl`: 31, 32, 33, 34, 65, 200 Elemente × {keines, erstes, letztes Element ungültig} × Ort `query/<n>` (n = 1 bis 3); dazu 31, 32 und 33 leere Elemente | Grenze 33; Prüfung der Elemente vor der Zahl | andere |
| 15 | Ein Element, das einem früheren bytegleich ist (`a,a`, nicht benachbart: `a,b,a`; auch `*,*` und `a/b,a/b`), ist `wert`; die Schreibweise zählt (`a,A` sind zwei) | `TestExpandDoppeltesElement`: Abstand 0 bis 3 zwischen den Vorkommen × Form {Name, `*`, Pfad}, Groß- und Kleinschreibung; `TestExpandDoppeltesElementAbstand`: alle Paare von Stellen (i < j) in Listen der Länge 2 bis 32 (5456) × Form {Name, `*`, Pfad} und Schreibweise geändert (Name, Pfad), 27 280 Fälle | Duplikate werden zusammengeführt; Vergleich ohne Beachtung der Schreibweise; `*,*` ergibt `wert_nicht_unterstuetzt`; Vergleich nur mit den letzten 4, 8 oder 30 Elementen; kein Vergleich an Stelle 31 | andere |
| 16 | Mehrere Verstöße im Wert: ein `(` geht allem vor, danach `zu_viele` vor den Elementen; Klasse `ungueltig` vor `nicht_unterstuetzt`; innerhalb der Klasse das erste Element von links (`SPEC-262` Nr. 3) | `TestExpandRangImWert`: alle Paare aus {leer, Name ungültig, doppelt, `*`, Pfad, Klammer} an den Stellen i und j (i ≠ j aus 0 bis 5, erste und letzte Stelle eingeschlossen; 1080), erwartete Ursache aus der Zusatztabelle in §6 | Die Klassen werden nach Position statt nach Klasse entschieden; die Klammer wird als Element geprüft; `*` oder Pfad am ersten, vorletzten oder dritten Element beendet das Lesen | andere |
| 17 | Rangfolge über Optionen: ein Verstoß in `$expand` und ein Verstoß einer anderen Option folgen `SPEC-160` Nr. 5, in beiden Reihenfolgen der Query | `TestExpandRangQuery`: Paare aus {Verstoß von `$expand`: leeres Element, ungültiger Name, doppeltes Element, `*`, Pfad, Klammer, `zu_viele`, `kodierung`} × {`$top=-1`, `$foo=1`, `$select=a,,b`, `$filter` ohne Leser} × {links, rechts} × {ohne, mit einer fremden Option davor} (128); `rang_test.go` bleibt das Muster | `$expand` wird vor den anderen Optionen entschieden; eine Klammer oder ein Pfad ist sofort der Fehler | andere |
| 18 | Der Ort ist `query/<n>` der Option (`n` zählt die nichtleeren Segmente), ohne Teil der Eingabe | `TestExpandOrt`: leere Segmente (`&&`) vor der Option, Option an Stelle 1 bis 4 | `n` zählt die leeren Segmente mit | andere |
| 19 | Kein Teil der Eingabe steht im Fehler: weder Name noch Wert noch Position in Bytes | `TestExpandFehlertext`: Schleife über alle Verstoßarten aus Nr. 5 bis 16, Eingabe mit unverwechselbarem Namen; `Error()` enthält ihn nicht | Der Name steht in der Meldung | andere |
| 20 | Das Ergebnis gehört dem Aufrufer: eine Änderung der Elemente ändert kein späteres Ergebnis, kein Slice wird zwischen Ergebnissen geteilt | `TestExpandKopie`: zweimal gleiche Eingabe, erstes Ergebnis ändern (Element, Namen), zweites vergleichen; Schleife über Listenlängen 1 bis 3 | Ein Paket-Slice wird als Ergebnis geliefert | andere |
| 21 | Gleiche Eingabe ergibt gleiches Ergebnis; ein Fehler gibt `ODataURL{}` | `TestExpandWiederholung`: jede Eingabe der Nr. 2 bis 16 zweimal, tiefe Gleichheit; Fehlerfälle gegen den Nullwert | Ein Zähler im Parser färbt das zweite Ergebnis | andere |
| 22 | Die Summe der Längen aller Texte im Ergebnis bleibt höchstens Eingabe plus 48 (`SPEC-160` Nr. 6 b) | `TestExpandWaechter`: Zahl {1, 2, 16, 31, 32} × Länge der Namen {1, 128} × {roh, jedes Byte als `%XX`} (20), dazu 32 Namen zu 128 Zeichen mit einer Query von 8191, 8192 und 8193 Bytes (Auffüllung durch eine fremde Option; 8193 ist `zu_lang`) und ein Name aus 128 Zeichen zu zwei Bytes; die Textsumme gleicht genau der Summe der Namen, die Zuteilung bleibt unter 64 · n + 65536 | Das Element trägt eine Kopie des ganzen Werts | andere |
| 23 | Gleichzeitige Aufrufe auf einer Instanz und auf mehreren beeinflussen sich nicht; jedes Ergebnis gleicht dem eines einzelnen Aufrufs; kein Aufruf schreibt ein Feld des Parsers (`SPEC-160` Nr. 7, `SPEC-262` Nr. 8) | `TestExpandGleichzeitig`: 4 Goroutinen × 50 Runden × 490 Eingaben (Listen, Verstöße je Art, 432 Paare der Rangfolge) auf einer gemeinsamen und einer eigenen Instanz, Ergebnis gegen den Einzelaufruf, dessen Elemente kopiert sind (Wirkungs-Fänger, verglichen erst nach dem Ende aller Goroutinen); `make test-race` je Prozess für ein geteiltes Feld ohne Wirkung; `TestExpandParserUnveraendert`: 4 Parser (ohne, mit Leser, mit `WithExpand`, mit beidem) × alle Eingaben von `korpus()`, `exQueries`, den Folgen der Länge 1 bis 3 des Alphabets (je mit und ohne `$top=1` davor), Queries von 8191 bis 8193 Bytes und 32 Namen zu 121 bis 122 Zeichen; der Parser gleicht nach jedem Aufruf einer Kopie vor dem ersten Aufruf (Wirkung statt Rennen, `SPEC-262` Nr. 8 „danach unveränderlich“; Prüfgrenze: ein Schreiben nur bei einer Eingabe außerhalb dieser Menge bleibt ungefangen) | Ein Puffer im Parser wird zwischen Aufrufen geteilt; ein Aufruf schreibt `expand` oder `filter` des Parsers | nebenläufig |
| 24 | Jede Zeichenfolge über dem Alphabet der Grammatik bis zur Prüfgrenze (Länge 4, §3) hat die Klasse des Orakels: Knoten, `wert`, `wert_nicht_unterstuetzt` oder `zu_viele` | `TestExpandAlphabet`: vollständige Aufzählung aller Folgen bis Länge 4 über 17 Zeichen (88 740 Folgen), Orakel ohne Code des Lesers; akzeptiertes Negativ: längere Folgen nur als Familien (Nr. 9, 14, 22), Folge: eine Regel, die erst ab Länge 5 greift, bleibt ungefangen | Der Leser behandelt `;` oder `=` im Element als Trenner | andere |
| 25 | Der Parser liest kein Modell und keinen Dienst: die Richtung der Importe bleibt | `make a-check` (Regeln des Hexagons, Ausgabe grün); ein Import von `model.ServiceModel` oder eines Dienstes im Paket wird rot | `urlparser` importiert `services/mock` | andere |
| 26 | Die Beispiele der Spezifikation (Tabelle der Gültigen, der `400` und der `501`) ergeben die genannte Antwort: eigener Parser-Contract-Test | `TestExpandContract`: Tabelle der Beispiele als Fixture (Eingabe, Knoten oder Ursache und Ort), ein Fall je Zeile; keine Eingabe aus dem Generator | Die Antwort einer Zeile wird vertauscht | andere |
| 27 | Der Knoten trägt die Felder des Vertrags, jedes andere Feld von `Optionen` bleibt unverändert (Name, Typ, Reihenfolge) | `TestOptionenFelder` in `model`: die Feldliste von `Optionen` und des Knotens per `reflect` gegen die Liste des Vertrags | Ein Feld von `Optionen` wird umbenannt oder entfällt | andere |
| 28 | Das Binary bleibt unverändert: `GET /Products?$expand=Foo` ist `501` mit `ODR-E6001` (`ODV4-E-03` c) | die bestehende Zeile in `src/bootstrap/abnahme_matrix_test.go` (c) baut ihren Dienst selbst und sieht den Aufbau in `wurzel.go` nicht (Mutation grün gesehen); der Fänger des Aufbaus ist `TestSatzVersion4OhneExpand` in `src/bootstrap/expand_wurzel_test.go`: `SatzVersion4().URL` liest `$expand` nicht (4 Werte), `$filter` bleibt lesbar; zweiter Fänger, den `make test` fährt und das Mutationswerkzeug nicht: `TestVertragFehlerklassen` in `tests/integration` (`$expand=GEHEIM-Q` ist mit eingehängter Option `400` statt `501`) | Der Aufbau des Dienstes hängt die Option ein | andere |
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
| 3 | Folgen der Einhängestellen | Folgen der Länge 0 bis 3 über 3 Einhängungen × 7 Queries | Zustand aus der Folge (letzte `WithFilter`, irgendeine `WithExpand`) | 40 × 7 = 280 |
| 12 | Segmente | 11 Segmente × 3 Stellen und allein | Tabelle `zulaessig` | 44 |
| 12 | Zeichen im Segment | 359 Zeichen (125 ASCII roh und als `%XX` in beiden Schreibweisen, 7 jenseits von ASCII) × 24 Formen (1 bis 3 Teile × Teil mit dem Zeichen × 4 Inhalte) × 4 Stellen | `exOrakel` | 34 464 |
| 12 | Struktur des Segments | Folgen der Länge 0 bis 4 über 11 Zeichen (`a`, `Z`, `_`, `0`, `.`, `$`, `-`, `*`, é, NBSP, ٣) × 4 Stellen | `exOrakel` | 16 105 × 4 = 64 420 |
| 12 | Wörter mit `$` | 26 Wörter und ihre Nachbarn im Abstand eins (gestrichen, ersetzt, eingefügt aus 7 Zeichen, ein Zeichen groß, qualifiziert) roh und mit `%24`, ohne Doppelte × 4 Stellen | `exOrakel` | 5720 × 4 = 22 880 |
| 12 | Länge je Teil | 6 Paare (Zahl der Teile, Teil) × Länge {1, 127, 128, 129} × Breite {1, 2 Bytes} × 4 Stellen; alle Teile 128 | `exOrakel` | 192 + 16 |
| 12 | Zahl der Teile | 340 Stufen (1 bis 300, jede 97. bis 4083, 4090) × Stelle des Teils (bis 40 jede, darüber erste, mittlere, letzte) × Inhalt (bis 40: 10 Inhalte × 3 Schreibweisen des Trenners, darüber `b` und leer) × 4 Stellen im Pfad; ohne Inhalt; bis 40 verschiedene Namen | `exOrakel` | 98 400 + 7200 + 1360 + 480 = 107 440 |
| 12, 11 | Zahl der Segmente | 340 Stufen × Stelle (bis 40 jede, darüber erste, mittlere, letzte) × Art (11, darüber 3) × Trenner `/`, `%2F` (bis 40); alle Segmente von einer Art; das Element als erstes, mittleres, letztes einer Liste | `exOrakel` | 18 040 + 2700 + 880 + 1320 = 22 940 |
| 12 | Segmente × Teile | 1 bis 6 Segmente × 1 bis 8 Teile × Segment × Teil × {`-`, `b`}, ohne Inhalt; große Paare aus {2, 5, 17, 64, 200}² mit höchstens 4000 Teilen | `exOrakel` | 1512 + 48 + 84 = 1644 |
| 12 | Länge gesamt | 1 bis 30 Teile × Breite {1, 2 Bytes} × {alle 128, einer 129 an erster, mittlerer, letzter Stelle} × 4 Stellen im Pfad | `exOrakel` | 936 |
| 13 | Klammer | 9 Optionen × 3 Stellen × 7 Vorläufe × 2 Nachläufe, 7 feste Formen | immer `501`, außer `a)` | 378 + 7 |
| 14 | Zahl | {31, 32, 33, 34, 65, 200} × 3 Verstöße × 3 Orte; 31, 32, 33 leere Elemente | Zahl > 32 `zu_viele`, sonst Verstoß, sonst Knoten | 54 + 3 |
| 15 | doppeltes Element | Abstand 0 bis 3 × {Name, `*`, Pfad} × {gleich, Schreibweise} | gleich `wert`, sonst Name oder `501` | 20 |
| 15 | Abstand | alle Paare von Stellen i < j in Listen der Länge 2 bis 32 (5456) × {Name, `*`, Pfad} gleich, und {Name, Pfad} in anderer Schreibweise | gleich `wert`, sonst Namen oder `501` | 16 368 + 10 912 |
| 16 | Paare im Wert | 6 × 6 Arten × 30 Stellenpaare (i ≠ j aus 0 bis 5) | Klammer, dann `ungueltig`, dann `501` | 1080 |
| 17 | Paare über Optionen | 8 Verstöße von `$expand` (6 Arten im Wert, `zu_viele`, `kodierung`) × 4 Verstöße anderer Optionen × 2 Seiten × 2 Stellen | `rang` aus `rang_test.go` | 128 |
| 18 | Ort | Stelle 1 bis 4 × {0, 1, 3} leere Segmente × 3 Orte | `query/<n>` | 36 |
| 22 | Wächter | {1, 2, 16, 31, 32} × {1, 128} × {roh, `%XX`}; 32 × 128 mit 8191, 8192, 8193 Bytes; 2-Byte-Name | Textsumme gleich Summe der Namen, Zuteilung | 24 |
| 23 | Nebenläufigkeit | 4 Goroutinen × 50 Runden × 490 Eingaben | Einzelaufruf, kopiert | 98 000 Aufrufe |
| 23 | Parser unverändert | 4 Parser × Eingaben (`korpus()`, 490 `exQueries`, 5219 Folgen der Länge 1 bis 3 mit und ohne `$top=1`, 3 lange Queries, 1 lange Liste) | Kopie des Parsers vor dem ersten Aufruf | 4 × (490 + 10 438 + 4 + Zahl der Eingaben von `korpus()`) |
| 24 | Alphabet | alle Folgen der Länge 1 bis 4 über 17 Zeichen (`a`, `Z`, `_`, `0`, `,`, `/`, `*`, `(`, `)`, `$`, `.`, `;`, `=`, `%20`, `%2C`, é, NBSP) | `exOrakel` | 88 740 |
| 26 | Beispiele | Tabelle der Spezifikation, 27 Zeilen, dazu 4 Ressourcenarten | Wortlaut der Tabelle | 27 + 4 |

Prüfgrenzen mit Folge im Betrieb (akzeptierte Negative, `AGENTS.md` §3.17): Länge 4 über 17 Zeichen (Nr. 24), Zahl der Elemente 200 (Nr. 14), höchstens vier leere Elemente und drei Namen neben `*`; eine Regel, die erst jenseits greift und keine Familie trifft, bleibt ungefangen. Gesehen: die Mutation `n03` (Regel nur bei Länge 9 des Werts) bleibt im Alphabet-Test grün und wird nur zufällig von `TestExpandListe` gefangen (`rot-paket`). Dazu aus der Fix-Runde (Tabelle unten): Länge 4 über 11 Zeichen für die Struktur eines Segments (Folge im Betrieb: ein Verstoß, der erst ab Länge 5 greift und weder in den 24 Formen von `TestExpandSegmentZeichen` noch bei den Wörtern vorkommt, bleibt ungefangen), Länge 3 für die Folgen der Einhängestellen (es gibt drei; ein Zustand, der erst die vierte Option verlangt, bleibt ungefangen) und die Menge der Eingaben von `TestExpandParserUnveraendert` (ein Schreiben, das nur bei einer Eingabe außerhalb von `korpus()`, `exQueries`, den Folgen bis Länge 3 und den vier langen Eingaben und den gültigen Listen von 1 bis 32 Namen greift, bleibt ungefangen; Folge im Betrieb: ein Parser, den ein Aufruf verändert, ist ein Datenrennen oder eine Wirkung auf spätere Aufrufe, `make test-race` und `TestExpandGleichzeitig` fangen die Wirkung bei den 490 Eingaben der Zeile 23).

**Mutationen, vom Implementer gefahren** (`sh tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, ein Lauf mit 52 Mutationen, Klasse `andere`, ein Prozess je Stufe; Host-Last der Ergebniszeilen 7,1 bis 9,6 bei 4 CPUs, Messender der Lauf; Ergebnis 50 `rot`, 1 `rot-paket`, 1 `flatternd`, 0 `gruen`, 0 Übersetzungsfehler). Die Namen der Mutationen tragen die Nr. der Zeile; zu jeder Zeile ein Element, das der Entwurf nicht nannte, ist mit `n…` bezeichnet:

| Nr. | Mutationen (alle `rot`, wo nichts anderes steht) | roter Test |
|---|---|---|
| 1 | `New` liest die Option | `TestExpandOhneEinhaengung` |
| 2 | Elemente sortiert; `WithExpand` setzt nichts | `TestExpandListe`; `TestExpandContract` |
| 3 | die Einhängung lässt `Sonstige` weg; `WithExpand` setzt `filter` zurück; `WithFilter` setzt `expand` zurück | `TestExpandDifferenz`; `TestExpandOptionenFolgen` |
| 4 | Wert zweimal dekodiert; Name zweimal dekodiert | `TestExpandDekodierung` |
| 5 | zweites Vorkommen überschreibt | `TestExpandDoppelt` |
| 6, 7 | leer ergibt den Nullwert; leere Elemente übergangen | `TestExpandLeer`, `TestExpandLeeresElement` |
| 8 | nur das erste Zeichen geprüft; jedes Zeichen jenseits von ASCII verworfen | `TestExpandNamenform` |
| 9 | Grenze 129; Länge in Bytes | `TestExpandNamenGrenze` |
| 10, 11 | `*` als Name; Pfad als erster Name; Pfad höchstens 3 Segmente, `n07`: höchstens 7 | `TestExpandStern`, `TestExpandPfad` |
| 12 | `$levels` als Segment; `*` als Segment; `$ref` kein Segment; Fix-Runde: `$REF`, `$filter` und `$select` als Segment, `-`, 129 Zeichen, Ziffer nach dem Punkt (Tabelle unten) | `TestExpandBesondereSegmente`; `TestExpandSegmentZeichen`, `…Struktur`, `…Worte`, `…Laenge`; Verifikation (Tabelle unten): `TestExpandSegmentTeile`, `…Zahl`, `…Kreuz`, `…LaengeGesamt` |
| 13 | Klammer nach der Zahl; Klammer nach den Elementen; `n05`: letztes Zeichen ungeprüft | `TestExpandKlammer` |
| 14 | Grenze 33; Elemente vor der Zahl; `n04`: nur genau 33 Elemente | `TestExpandZahl` |
| 15 | Duplikate zusammengeführt; Vergleich ohne Schreibweise; `*,*` ergibt `501`; Fix-Runde: Fenster 4, 8, 30, keine Prüfung an Stelle 31 | `TestExpandDoppeltesElement`; `TestExpandDoppeltesElementAbstand` |
| 16 | ein `501` beendet das Lesen; Klammer wie ein Element; Fix-Runde: `*` oder Pfad am ersten, vorletzten, dritten Element beendet das Lesen | `TestExpandRangImWert` |
| 17 | `501` sofort als Fehler; Fix-Runde: Klammer sofort als Fehler, Pfad sofort als Fehler | `TestExpandRangQuery` |
| 18 | Ort ab 0; leere Segmente zählen (in `model.QuerySegmente`) | `TestExpandOrt` |
| 19 | Name im Ort der Meldung | `TestExpandFehlertext` |
| 20 | Paket-Slice als Ergebnis | `TestExpandKopie` |
| 21 | Zähler färbt das zweite Ergebnis | `TestExpandWiederholung` |
| 22 | Element trägt den ganzen Wert | `TestExpandWaechter` |
| 23 | Feld `zaehler` im Parser; Fix-Runde: ein Aufruf schreibt `expand` (nach `$top`, nach 8192 Bytes) oder `filter` (beim Lesen von `$expand`); Verifikation: ein Aufruf mit genau 7, 13 oder 21 gültigen Namen schreibt `filter` oder `expand` | `TestParserOhneZustand`; `TestExpandParserUnveraendert` |
| 24 | `;` trennt; `=` trennt; `n01`: `$` als Präfix eines Namens; `n02`: `)` wie `(`; `n03`: Regel nur bei Länge 9 (`rot-paket`, Alphabet-Test grün) | `TestExpandAlphabet` |
| 25 | Import eines Nachbar-Adapters; Import eines Dienstes (`make a-check`: `wrong-direction`, von Hand gesehen) | `TestImporte` |
| 26 | `$ref` kein Segment (Zeile `Category/$ref`) | `TestExpandContract` |
| 27 | zusätzliches Feld in `Optionen`; im Knoten | `TestOptionenFelder` |
| 28 | `wurzel.go` hängt `WithExpand()` ein: `TestAbnahmeMatrix` bleibt grün (die Tabelle baut ihren Dienst selbst), `TestSatzVersion4OhneExpand` ist rot (4 von 4 Werten, von Hand gefahren: das Paket `src/bootstrap` läuft im Mutationslauf nicht, `Ausgangsstand rot`) | `TestSatzVersion4OhneExpand` |

**Raten der Nebenläufigkeit** (Zeile 23, `TestExpandGleichzeitig`, je Mutation 50 Prozesse mit Detektor und 50 ohne, `-test.count=1`, 4 Goroutinen je Prozess, Host-Last 8,4 / 7,3 / 7,1 bei 4 CPUs, Messender `mutationslauf`; jede Zahl ist eine Größenordnung unter Last): geteilter Puffer der Elemente (Wirkung): mit Detektor 50/50, ohne 50/50; geteilter Zähler mit Wirkung: 50/50 und 50/50; geteilter Zähler ohne Wirkung (nur ein Datenrennen): mit Detektor 50/50, ohne 0/50. Das Datenrennen ohne Wirkung fängt allein `make test-race`; der Wirkungs-Fänger fängt Puffer und Wirkung in jedem Lauf, ohne auf das Rennen zu warten.

**Fix-Runde nach dem Review** (`docs/reviews/2026-11-18-slice-expand-parser-1.md`, Befunde F-1 bis F-7; `AGENTS.md` §3.17). Die Menge ist aus dem Wortlaut von `SPEC-262` gebildet, nicht aus dem Befund; das Orakel ist `exOrakel` (Tabelle der Zeichenklassen, kein Code des Lesers). Mutationen: `sh tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, ein Lauf mit 27 Mutationen der Klasse `andere` (ein Prozess je Stufe), Ergebnis 27 `rot`, 0 `rot-paket`, 0 `gruen`; Host-Last 4,4 bis 4,9 bei 4 CPUs, Messender der Lauf des Implementers. Die Kennungen `f<N>-r…` sind die Mutationen des Reviews (als Gegenprobe der Fänger rot), `f<N>-n…` Mutationen an einem Element, das der Befund nicht nannte:

| Befund | Menge · Achsen · Elementzahl je Achse | Fänger | Mutation an einem Element, das der Befund nicht nannte (rot gesehen) |
|---|---|---|---|
| F-1 | jedes Segment eines Pfads (Name, qualifizierter Name, `$ref`, `$count`, `$value`; `SPEC-262` Nr. 4 c, d), Zeichenmenge und Struktur getrennt. **Zeichen:** 359 Zeichen × 24 Formen (Zahl der Teile 1 bis 3 × Teil mit dem Zeichen × 4 Inhalte) × 4 Stellen = 34 464. **Struktur:** Länge 0 bis 4 über 11 Zeichen (16 105 Folgen) × 4 Stellen = 64 420. **Wörter:** 26 Wörter mit `$` samt Nachbarn im Abstand eins, 5720 Schreibweisen × 4 Stellen = 22 880. **Länge je Teil:** 6 × 4 × 2 × 4 = 192, dazu alle Teile 128 = 16 | `TestExpandSegmentZeichen`, `…Struktur`, `…Worte`, `…Laenge` | `f1-n1` (nur zwei Teile eines qualifizierten Namens geprüft), `f1-n2` (Länge erst ab dem zweiten Teil unbegrenzt), `f1-n3` (`$value` als Präfix), `f1-n4` (leerer mittlerer Teil `a..a`), `f1-n5` (NBSP am Ende abgeschnitten), `f1-n6` (`²` als Teil); die vom Review genannten `f1-r14` bis `f1-r18` sind rot |
| F-2 | Abstand eines doppelten Elements: alle Paare i < j in Listen der Länge 2 bis 32 (31 Längen, 5456 Paare) × Form {Name, `*`, Pfad} gleich (16 368) und {Name, Pfad} in anderer Schreibweise (10 912) = 27 280. Die Menge ist beschränkt (mehr als 32 Elemente sind `zu_viele`, `SPEC-262` Nr. 6) und vollständig, kein Negativ | `TestExpandDoppeltesElementAbstand` | `f2-n1` (Fenster 8), `f2-n2` (keine Prüfung an Stelle 31), `f2-n3` (Fenster 30, Abstand 31 fehlt); `f2-r19` (Fenster 4) ist rot |
| F-3 | Zusammensetzung der Einhängestellen von `New`: Folgen der Länge 0 bis 3 über {`WithFilter(Leser)`, `WithFilter(nil)`, `WithExpand()`} (1 + 3 + 9 + 27 = 40) × 7 Queries = 280. Prüfgrenze Länge 3 (Grund und Folge: §3, Prüfgrenzen) | `TestExpandOptionenFolgen` | `f3-n1` (`WithFilter` setzt `expand` zurück), `f3-n2` (`WithExpand` kippt das Merkmal, zweite Einhängung schaltet aus), `f3-n3` (`WithFilter(nil)` schaltet `expand` ein); `f3-r02` (`WithExpand` setzt `filter` zurück) ist rot |
| F-4 | Paare von Verstößen im Wert: 6 Arten × 6 Arten × 30 Stellenpaare (i ≠ j aus 0 bis 5, erste und letzte Stelle eingeschlossen) = 1080; Rang über Optionen: 8 Verstöße von `$expand` (6 Arten im Wert, `zu_viele`, `kodierung`) × 4 Verstöße anderer Optionen × 2 Seiten × 2 Stellen = 128 | `TestExpandRangImWert`, `TestExpandRangQuery` | `f4-n1` (`*` oder Pfad am vorletzten Element beendet das Lesen), `f4-n2` (Pfad am dritten Element), `f4-n3` (Pfad ist sofort der Fehler); `f4-r26` und `f4-q02` sind rot. Eine Mutation „am letzten Element beendet das Lesen“ ist im Wert gleichwertig (nichts folgt) und nur über Optionen sichtbar: `TestExpandRangQuery` fängt sie, `TestExpandRangImWert` nicht (`rot-paket`, im ersten Lauf gesehen und ersetzt durch `f4-n1`) |
| F-5 | Aufrufe von `ParseURL` auf einem Parser: 4 Parser (ohne, mit Leser, mit `WithExpand`, mit beidem) × (Eingaben von `korpus()`, 490 von `exQueries`, 2 × 5219 Folgen der Länge 1 bis 3, 3 Queries von 8191 bis 8193 Bytes, 1 Liste von 32 langen Namen); nach jedem Aufruf gleicht der Parser einer Kopie. Prüfgrenze: die Menge der Eingaben ist unbeschränkt (§3, Prüfgrenzen) | `TestExpandParserUnveraendert` | `f5-n1` (`leseExpand` setzt `filter` zurück), `f5-n2` (`expand` wird bei 8192 Bytes ausgeschaltet); `f5-z01` (nach `$top` wird `expand` gesetzt) ist rot |
| F-6 | Zeile 28 nennt beide Fänger des Aufbaus (`TestSatzVersion4OhneExpand`, `TestVertragFehlerklassen`) | Text von Zeile 28 | keine (Plan-Zeile) |
| F-7 | Messpunkt beim Handoff: §4 | `make groessenmarke` | keine (Plan-Zeile) |

**Fix-Runde nach der Verifikation** (`docs/reviews/2026-11-18-slice-expand-parser-verifikation.md`, Befunde V-1 bis V-5; `AGENTS.md` §3.17). Die Menge ist aus dem Wortlaut von `SPEC-262` Nr. 4 c und Nr. 6 gebildet: Zahl der Teile eines qualifizierten Namens, Zahl der Segmente eines Pfads und Länge des Segments als Ganzes sind allein durch `rawQuery` (8192 Bytes) begrenzt, also höchstens 4090 Teile oder Segmente mit dem Präfix `N/N/`. Orakel: `exOrakel` (kein Code des Lesers). Mutationen: `sh tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, zwei Läufe (16 und 2 Mutationen der Klasse `andere`, ein Prozess je Stufe), Ergebnis 18 `rot`, 0 `rot-paket`, 0 `gruen`, je am benannten Test; Host-Last 3,1 bis 3,8 von 4 CPUs, Messender der Lauf des Implementers. Die Klasse `andere` ist deterministisch; die Raten der Zeile 23 sind nicht neu gemessen (kein Code der Nebenläufigkeit berührt).

| Befund | Menge · Achsen · Elementzahl je Achse | Fänger | Mutation an einem Element, das der Befund nicht nannte (rot gesehen) |
|---|---|---|---|
| V-1 | Zahl der Teile: 340 Stufen (1 bis 300, jede 97. bis 4083, 4090) × Stelle des Teils (bis 40 Teile jede, darüber 3) × Inhalt (bis 40 Teile 10, darüber 2) × Trenner (3 Schreibweisen, darüber 1) × 4 Stellen im Pfad = 107 440. Zahl der Segmente: 340 Stufen × Stelle (jede bzw. 3) × Art (11, darüber 3) × Trenner (2, darüber 1) + alle Segmente von einer Art + Stelle in der Liste = 22 940. Segmente × Teile: 21 × 36 × 2 + 48 + 21 × 4 = 1644. Länge gesamt: (30 Fälle alle 128 + 87 Fälle mit einem Teil 129) × 2 Breiten × 4 Stellen im Pfad = 936 | `TestExpandSegmentTeile`, `…Zahl`, `…Kreuz`, `…LaengeGesamt` | Teile: `v1-teile9-splitn`, `v1-teile5-splitn` (die Mutation des Verifiers), `v1-teile-ab-7-ungeprueft`, `v1-teile-ab-33-ungeprueft`, `v1-teile-ueber-129-ungueltig`, `v1-teile-ueber-2000-ungueltig`, `v1-segment-laenge-ueber-1000`; Länge: `v1-teil-ab-20-laenge-129`; Segmente: `v1-segmente-ueber-17-ungueltig`, `…ab-12-ungeprueft`, `…ab-100-ungeprueft`, `…ueber-3000-ungueltig`, `v1-segment-ab-3-nur-name`; Kreuz: `v1-gesamtteile-ueber-1000`, `v1-letztes-segment-teile-ab-50` |
| V-2 | gültige Listen von 1 bis 32 Namen (32) × {ohne, mit vorangestelltem `$top=1`} = 64 Eingaben × 4 Parser; Zustand des Parsers nach jedem Aufruf gegen die Kopie | `TestExpandParserUnveraendert` | `v2-schreibt-liste-7` (löscht `filter`), `v2-schreibt-liste-21` und `v2-schreibt-liste-13` (schalten `expand` aus) |

Prüfgrenze mit Grund und Folge im Betrieb (akzeptiertes Negativ, `AGENTS.md` §3.15, §3.17): Die Zahlen von Teilen und Segmenten über 300 sind nur in den Stufen 397, 494, … 4083 und 4090 geprüft, an der ersten, mittleren und letzten Stelle, und die Gesamtform nur in den 21 großen Paaren der Zeile `Segmente × Teile`. Grund: der Aufwand je Eingabe wächst mit der Zahl; alle 4090 Zahlen mit jeder Stelle brauchten ohne Detektor 16 Sekunden (eine Messung, Last 3 bis 4 von 4 CPUs, `go test` im Image `odata-recorder:test`), mit dem Detektor ein Vielfaches. Folge im Betrieb: eine Regel, die nur eine Zahl zwischen zwei Stufen trifft und keine größere (kein Schwellenwert), liefert dort `400` statt `501` oder umgekehrt und bleibt ungefangen; ein Schwellenwert trifft die Stufe 4090 und ist gefangen (`v1-teile-ueber-2000-ungueltig`). Ebenso bleibt ungefangen, was erst bei einer Stelle zwischen der ersten, mittleren und letzten eines Segments mit mehr als 40 Teilen oder Pfads mit mehr als 40 Segmenten greift.

**Fix-Runde nach der Nachverifikation** (`docs/reviews/2026-11-18-slice-expand-parser-nachverifikation.md`, Befunde N-1 bis N-4; `AGENTS.md` §3.17). Die Befunde nennen Kreuzungen der Achsen, die der Fix der Verifikation nicht führte; die Mengen sind aus dem Wortlaut von `SPEC-262` Nr. 3 d, 4 c, 4 e und 6 gebildet. Orakel: `exOrakel`; die Generatoren sind die Proben P1 bis P6 des Verifiers als Schleifen, jeder Test zählt seine Menge gegen die Zahl der Tabelle (`exZaehlt`). Alle sechs Tests laufen zusammen in rund 0,15 s ohne Detektor (ein Lauf, Image `odata-recorder:race`, Host-Last 17 von 4 CPUs, Messender der Lauf des Implementers). Mutationen: `sh tools/harness/mutationslauf.sh`, Klasse `andere` (deterministisch), zwei Läufe: die elf Mutationen `nv01`, `nv05` bis `nv13` und `nv17` des Berichts (11 `rot`, je am Test der Zeile) und acht Mutationen `n<N>-n…` an Elementen, die der Bericht nicht nannte (8 `rot`, 0 `rot-paket`, 0 `gruen`); die Raten der Zeile 23 sind nicht neu gemessen (kein Code der Nebenläufigkeit berührt).

| Befund | Menge · Achsen · Elementzahl je Achse | Fänger | Mutation an einem Element, das der Bericht nicht nannte (rot gesehen) |
|---|---|---|---|
| N-1 | Art (6) × jede Stelle 0 bis k-1 für k in {2, 3, 7, 12, 25, 32} (81 Stellen, 486) + je Paar von Stellen (i, j), i ungleich j, 4 Paare von Verstößen: 4 × (2 + 6 + 42 + 132 + 600 + 992) = 7096; zusammen 7582 | `TestExpandStelleImWert` | `n1-n01` (`*` ab Stelle 27 übergangen), `n1-n02` (Pfad ab Stelle 19 ist `wert`); die Mutationen des Berichts: `nv01`, `nv17` |
| N-2 | Dublette: 23 Formen (`*`, 7 Namenslängen, 15 Pfadformen) × Abstand {0, 1, 5, 20} = 92 Eingaben bei gleichen Elementen; verschiedene Namen gleicher Länge mit gleichem Präfix, Ende oder gleicher Mitte (6 Längen × 3 × 2 = 36) und verschiedene Pfade (3 Längen × 3 = 9): zusammen 137 | `TestExpandDublette` | `n2-n01` (Dublette nach den letzten 16 Bytes), `n2-n02` (keine Dublette ab 39 Schrägstrichen); `nv05`, `nv06` |
| N-3 | Folge der Segmentarten: 11 Arten, alle 121 Paare und 1331 Tripel = 1452 | `TestExpandSegmentFolge` | `n3-n01` (`$ref` nach `$count` ist `wert`), `n3-n02` (`$count` nach `$value` ist `wert`); `nv07`, `nv08` |
| N-4 | Größe an der Stelle: Listen mit 2, 4, 8, 16, 20, 32 Elementen × jede Stelle × 4 Arten (Pfad, Teile, langer Name, Gemisch) × 45, 100, 200 Glieder, je mit und ohne Verstoß am Nachbarn = 1968. Länge des Werts: 31 bis 34 Elemente von 1 bis 128 Zeichen, je roh, mit `(` und mit `,x(` am Ende, und Pfade von 100 bis 4000 Segmenten vor `(` mit 4 Enden = 88 | `TestExpandGroesseAnStelle`, `TestExpandLaengeVorKlammerUndZahl` | `n4-n01` (Element über 300 Bytes ab Stelle 5 ist `wert`), `n4-n02` (`(` nur in den ersten 6000 Bytes gesucht); `nv09` bis `nv13` |
| N-1 (Dekodierung, Probe P6 des Berichts, Abschnitt 3) | 10 Escapes × 21 Stellungen × 2 Längen (kurz, nach 120 Segmenten) = 420 | `TestExpandEscapeInJederStellung` | keine neue; der Bericht fand die Dekodierung gedeckt (keine grüne Mutation) |

Prüfgrenze P7 (akzeptiertes Negativ, `AGENTS.md` §3.17): jede Zahl von Teilen und Segmenten von 301 bis 4090 an erster, mittlerer und letzter Stelle (170 550 Eingaben) ist nicht in der Schleife. Grund: der Verifier maß 30,6 s ohne und 62,5 s mit Detektor (ein Lauf, Container mit 4 CPUs, Host-Last 12 bis 27, `nproc` 3, Messender der Verifier); das Paket brauchte mit dem Detektor 26,3 s, die Zeitgrenze von `make test-race` beträgt 10 Minuten je Paket (`SPEC-122` Nr. 5). Folge im Betrieb: eine Regel, die nur eine Zahl zwischen zwei Stufen trifft und keine größere, bleibt ungefangen (wie in der Prüfgrenze darüber; ein Schwellenwert wird von `TestExpandSegmentTeile` und `…Zahl` gefangen). Dieselbe Folge gilt für die Stellen 6 bis 31 bei den Paaren der Zeile `TestExpandRangImWert` nur dort, wo `TestExpandStelleImWert` sie nicht führt; diese führt jede Stelle der Listenlängen 2, 3, 7, 12, 25 und 32, nicht jede Listenlänge dazwischen (akzeptiertes Negativ: eine Regel, die nur bei einer Listenlänge zwischen diesen greift und keine längere trifft, bleibt ungefangen). Offen bleibt ebenso die Größe der Elemente zwischen 45, 100 und 200 Gliedern.

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
- Handoff (gemessen mit `make groessenmarke` am Stand nach der Fix-Runde der Verifikation, letzter Commit `ef038aa2`): `Messpunkt slice-expand-parser.md: zeilen=30 nebenlaeufig=1 fakes=0 produktcode=150 schaetzung=2555 marke=3500 messung=2004`; Produktcode von Hand gemessen 127 Zeilen (Summe aus Einfügungen und Löschungen der Dateien ohne Tests: `expand.go` 90, `parser.go` 10, `query.go` 14, `url_ast.go` 1, `url_ast_expand.go` 12), Erwartung 150; die Messung liegt in der Spanne der Erwartung (1500 bis 2400); Auslastung 2004 ÷ 3500 = 0,57, unter 0,60, das meldet der Planner in der Closure dem Architect (`SPEC-128` Nr. 9).
- Handoff nach der Fix-Runde der Nachverifikation (`make groessenmarke`, Stand nach dem Plan-Commit dieser Runde): `Messpunkt slice-expand-parser.md: zeilen=30 nebenlaeufig=1 fakes=0 produktcode=150 schaetzung=2555 marke=3500 messung=2268`; Auslastung 2268 ÷ 3500 = 0,65, über 0,60 (`SPEC-128` Nr. 9: der Planner meldet es in der Closure dem Architect); die Messung liegt knapp unter der Spanne der Erwartung nach oben (1500 bis 2400), über ihr liegt sie nicht.

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

- Die Menge der Optionstexte ist unbeschränkt (jede Zeichenfolge) — **Ausgang:** weiter offen; Prüfgrenze Länge 4 über 17 Zeichen und benannte Familien darüber, Folge im Betrieb als akzeptiertes Negativ in Nr. 24 der Tabelle (`AGENTS.md` §3.17).
- Schnitt: `welle-expand-batch` §4 und die Roadmap führen vier Slices, es sind fünf (`slice-expand-validator` ist neu); der Closure-Trigger der Welle verlangt „je Funktion ein Parser-, Validator-, Executor- und Serializer-Contract-Test“ — **Ausgang:** weiter offen; Adresse Planner (Integrationsschritt: Welle-Plan §2 und §4, Roadmap; dieser Zweig schreibt beides nicht, `docs/plan/planning/README.md` Übergangsregeln Nr. 4).
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
| Verstoß in `$expand`, Verstoß in `$top` | der erste der Klasse in der Lesefolge, `ungueltig` zuerst | `TestExpandRangQuery`: je Verstoß der sechs Arten im Wert (leeres Element, ungültiger Name, doppeltes Element, `*`, Pfad, Klammer) sowie `zu_viele` und `kodierung`, mit vier Verstößen anderer Optionen, auf beiden Seiten und an zwei Stellen |

Die Zeilen der Zusatztabelle mit zwei Verstößen im Wert führt `TestExpandRangImWert` für alle sechs Arten an jeder Stelle von 0 bis 5 (1080 Paare), nicht nur für die Stellen 1 bis 4.

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

Stand der Übergabe: Review Runde 1, Verifikation und Nachverifikation liegen unter `docs/reviews/` (`2026-11-18-slice-expand-parser-1.md`, `-verifikation.md`, `-nachverifikation.md`). Die Fix-Runde nach der Nachverifikation (N-1 bis N-4, Commits `7327ba0f`, `4ec1aba2`) hat noch keine Nachverifikation in frischem Kontext. Es fehlen: diese Nachverifikation, die Closure (Steering-Loop-Eintrag, Risiko-Ausgänge, Register-Evidenz einschließlich der Betriebsregel-Verstöße der Subagenten, DoD-Häkchen), `make slice-mv SLICE=slice-expand-parser TO=done` und der Integrationsschritt (origin/main einmergen, Roadmap, `welle-expand-batch` und Rückverfolgung, `make gates` im frischen Worktree, ff-only-Push).

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
