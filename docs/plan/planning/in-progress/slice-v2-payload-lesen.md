# Slice slice-v2-payload-lesen: Leser für V2-Antworten (JSON mit der Hülle `d`) für den Import

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), [ADR-0003](../../adr/0003-lint-profil-golangci.md) (`make lint` bleibt ohne Ausnahme grün, `AGENTS.md` §3.2, §3.6)

**Berührte Spec-Stellen:** `SPEC-170` (Umfang, Typen) · `SPEC-125` Nr. 9 a (Ketten von `BaseType`: ein Modell von Hand kann sie schließen, Zeile N1, Kreis) · `SPEC-172` Nr. 1 bis 5 (Lesen; Nr. 6 und 7 gehören `slice-v2-payload-schreiben`) · `SPEC-178` · `LH-FA-17.a` (versionierter Adapter für V2-Payloads, keine Vermischung von V2 und V4) · `LH-FA-17` Boundary (leeres Ergebnis bleibt im V2-Antwortformat konsistent) · `SPEC-126` (Vertrag des V4-Lesers als Vorbild; Schreibweisen der Typen Nr. 10) · `SPEC-127` Nr. 8 (Auswahl der Antworten durch den Import) · `SPEC-002` · `SPEC-006` (Fehlerobjekt) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkt OData V2).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-v2-payload-lesen)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Leser `jsoncodecv2` (Paket `src/adapters/driven/odata/jsoncodecv2`, `SPEC-172`) **liest** die JSON-Antworten eines V2-Dienstes (Hülle `d`) auf Entity- und Collection-Abfragen für den Import in typisierte Entities gegen ein `model.ServiceModel` mit Version `2.0`: Typen in der Schreibweise von V2 (`/Date(…)/`, `Edm.Int64` und `Edm.Decimal` als String), Anzahl `__count`, Folgelink `__next`, nicht gedeutete Konstrukte als übersprungen gemeldet; das leere Ergebnis bleibt eine leere Menge (`LH-FA-17`, Boundary). Das Modell baut der Test mit `model.NeuesServiceModel("2.0", …)`; der V2-Metadatenleser ist keine Voraussetzung. Der Schnitt von Lesen und Schreiben folgt `AGENTS.md` §3.14: beide teilen allein die Tabelle der Typen in der Spezifikation (`SPEC-170` Nr. 4), keinen Mechanismus im Code; der Rundlauf beider ist der Test von `slice-v2-payload-schreiben`.

**Teilen mit der V4-Suite** (`AGENTS.md` §3.9; `make lint` mit `dupl`, `cyclop`, `gocyclo`, `gocognit`, `unparam` ohne Ausnahme grün, [ADR-0003](../../adr/0003-lint-profil-golangci.md)): Die V2-Suite liegt im Paket `outboundtest` neben der V4-Suite und teilt drei Helfer mit ihr: `uPruefer` (Prüfer der Menge U, aus `b3Pruefer`), `a3Quellfehler` (die sieben Orte eines Fehlers von `src`, aus `jA3`) und `int32Werte` (die handgeschriebenen Fälle von `Edm.Int32`, aus `jfF3`); `b3Alphabet` verliert den Parameter `bis` (jeder Aufruf übergab 0xFF). Das Verhalten der V4-Suite ist unverändert (die Aufrufe und die Fälle sind dieselben, ihre Tests laufen grün); die Dateien der V4-Suite sind damit nicht mehr byte-gleich, siehe J4 in §3 und §6 (Abweichung, *entschieden*: zulässig nach [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) §Konsequenzen, weil `outboundtest` ein Paket ohne Versionsnamen ist und die V4-Tests unverändert bleiben).

**Übernommen von `slice-v2-metadaten`** (Geber; `AGENTS.md` §3.13, Sendung vor dem Nehmer eingetragen):

- Eine Fixture mit einer Property, deren `Unterstuetzt()` falsch ist und deren `EdmType` `Edm.DateTimeOffset(V2)` lautet (V2 nennt `Edm.DateTimeOffset` außerhalb der sieben Typen, `SPEC-170` Nr. 4): erwartet ist `Uebersprungen` mit Grund `typ` (`SPEC-172` Nr. 3); der Wert des Modells `Edm.DateTimeOffset`, das der V2-Metadatenleser aus `Edm.DateTime` bildet, bleibt `datetimeoffset` (`SPEC-170` Nr. 4). Beide Fälle gehören zum Lesen und stehen in der DoD (Lesen) und in §3 (Nr. 1 und 2).
- Der Test `src/adapters/driven/odata/versionstrennung_test.go` (eine Schleife über alle Verzeichnisse unter `src/adapters/driven/odata/`, kein Eintrag je Paket) kommt aus `slice-v2-metadaten`. Fehlt er beim Start dieses Slice, legt dieser Slice ihn nach demselben Wortlaut an; liegt er vor, ändert dieser Slice ihn nicht und die Schleife erfasst `jsoncodecv2` ohne Eintrag.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Das Schreiben der V2-Antworten (Entity, Collection, Anzahl, leeres Ergebnis, Fehlerobjekt, Darstellungskopf; `SPEC-172` Nr. 6 und 7) — `slice-v2-payload-schreiben`; es braucht `slice-odata-json-schreiben` und teilt keinen Mechanismus mit dem Lesen.
- Metadaten-Leser für V2 — `slice-v2-metadaten`; dieser Slice liest Entities gegen ein Modell, er bildet es nicht, und braucht den Leser nicht (das Modell ist derselbe Typ, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)).
- Query-Semantik und Kompatibilitätsfehler für Abfragen — `slice-v2-query`.
- Wahl des Adapters je Service und Konfiguration — `slice-v2-bootstrap`; die Anpassung des Imports an einen Satz von Codecs je Dienst ist `slice-v2-import-codecs`.
- V2-Mutationen und das Lesen von Client-Bodies für V2 — `welle-odata-v2` §6; was die Spezifikation nicht nennt, ergibt den Kompatibilitätsfehler.
- Vollständige Kompatibilität zu herstellerspezifischen (SAP-)Erweiterungen — Out-of-Scope von `LH-FA-17`.

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

- [x] Lesen: Antworten auf Entity- und Collection-Abfrage in typisierte Entities gegen das Modell der Version `2.0` (Hülle `d`, Typen in der Schreibweise von V2, `__count`, `__next`, nicht gedeutete Konstrukte als übersprungen gemeldet); kanonische Schreibweise je Typ; ein Modell anderer Version ist `ErrUngueltig`; V4-Leser und V4-Schreiber unberührt (ihre Tests laufen unverändert grün; die V4-Suite gibt drei Helfer an die V2-Suite ab, §1 *Teilen mit der V4-Suite*); eine nicht unterstützte Property (`Unterstuetzt()` falsch, `EdmType` `Edm.DateTimeOffset(V2)`) ist `Uebersprungen` mit Grund `typ`, und der Test `versionstrennung_test.go` über alle Verzeichnisse von `src/adapters/driven/odata/` liegt vor (aus `slice-v2-metadaten`, sonst von diesem Slice angelegt).
- [x] Randformen des Lesers fremder Bytes (leer, falsche Hülle, V4-Antwort am V2-Leser, Typ-Schreibweisen, Zahlen jenseits von `float64`, doppelte Namen, Größe, Tiefe, Zerlegung des Stroms in Reads): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Generator über die Klasse der Bodies.
- [x] `make gates` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist (der Architect schrieb den Absatz zu Kreisen in `SPEC-172` Nr. 3 samt Historie vom 2026-11-05).
- [x] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (ein V2-Dienst erreicht es erst mit `slice-v2-bootstrap`, das auch das Kapitel zu V2 trägt).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/adapters/driven/odata/jsoncodecv2/` | neu | Leser der V2-Antworten (`SPEC-172` Nr. 1 bis 5); kein Port wird geändert ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)) |
| Tests | neu | Fälle von `LH-FA-17` (Boundary, Happy Path) für das Lesen; Generator über die Klasse der Bodies |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Leser liest fremde Bytes (ein Format, das dieses Repo nicht schreibt): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); Bodies mit Escape-Folgen oder Rohbytes entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Auftrag an den Architect:** erledigt (Entwurf zu [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), `Accepted`). Auftrag 1 (Format) steht in `SPEC-170` Nr. 2 (JSON, Atom außerhalb), 2 und 3 (Hülle, Mitglieder, leeres Ergebnis) in `SPEC-172` Nr. 2 bis 4, 5 (Grenzen, Fehler) in `SPEC-172` Nr. 1, 6 (V4-Body am V2-Leser) in `SPEC-172` Nr. 2; Auftrag 4 (Fehlerobjekt, Darstellungskopf) betrifft das Schreiben und gehört `slice-v2-payload-schreiben` (`SPEC-172` Nr. 6 e und g).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Stand: vom Architect gefüllt, 57 Zeilen, davon 2 `nebenläufig` und 55 `andere`, Klasse nach `SPEC-128` Nr. 3 e). Die Zeilen 1 bis 3 sind die Sendungen von `slice-v2-metadaten` (§1 *Übernommen*) und tragen die Klasse `andere`. Mengen sind Schleifen oder Generatoren aus dem Wortlaut der Zusage (`AGENTS.md` §3.15, §3.17); jede Zeile mit Generator hat eine Zählprüfung (der Test nennt die Zahl der Elemente je Klasse). Orakel sind unabhängig vom Leser: `encoding/json` (`Valid`), `math/big`, `time`, `unicode/utf8`. Werte mit Escape-Folge oder Rohbytes entstehen aus Bytewerten (`esc("…")`, `[]byte{…}`) und die Dateien werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18). Der Fake eines Stroms hält jedes Argument fest und zählt jede Operation (Reads, angeforderte Bytes, `Close`, jede Schnittstelle der Menge O: `io.WriterTo`, `io.ByteReader`, `io.ByteScanner`, `io.RuneReader`, `io.RuneScanner`, `io.Seeker`, `io.ReaderAt`).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Eine Property mit `Unterstuetzt()` falsch und `EdmType` `Edm.DateTimeOffset(V2)` ist `Uebersprungen` mit Grund `typ` (`SPEC-172` Nr. 3), nicht `unbekannte_property` und kein Wert; aus `slice-v2-metadaten` | Fall in N4 mit gültigem Datumswert (damit nicht `wert` greift); erwartet genau `[{Name, typ}]` und keine Property | Prüfung von `Unterstuetzt()` entfernt; Grund `wert` statt `typ` | andere |
| 2 | Der Wert des Modells `Edm.DateTimeOffset` (aus `Edm.DateTime`) bleibt `datetimeoffset` (`SPEC-170` Nr. 4) | Fall in W8, der die `Art` prüft, nicht nur den Text | `Art` `date` oder `string` | andere |
| 3 | Eine Schleife über alle Verzeichnisse von `src/adapters/driven/odata/` hält die Versionstrennung (kein Eintrag je Paket); aus `slice-v2-metadaten` | `versionstrennung_test.go` (kommt per Merge; fehlt er, nach dem Wortlaut des Gebers anlegen); Gegenprobe: `jsoncodecv2` importiert `jsoncodec` | Import von `jsoncodec` in `jsoncodecv2` | andere |
| G1 | Argumente: `ctx` nil, `src` nil, `grenze` kleiner 1 (1 gültig), `entitySet` leer sind `ErrUngueltig` ohne `DatenFehler` und ohne Code; `src` wird nicht gelesen (SPEC-172 Nr. 1, SPEC-126 Nr. 1 a bis d) | Schleife vier Verstöße (`grenze` 0, -1, `MinInt64`) mal zwei Methoden, Strom-Fake zählt Reads (erwartet 0) | `grenze<1` zu `<0`; Prüfung nach dem ersten Read | andere |
| G2 | Ein Modell mit Version ungleich `2.0` (auch der Nullwert, Version leer) ist `ErrUngueltig`, `src` wird nicht gelesen; `2.0` liest (E1) | Schleife Versionen (leer, `1.0`, `3.0`, `4.0`, `2.0 `, `2.00`, `20`) mal zwei Methoden plus Gegenfall `2.0`; Strom-Fake zählt Reads | Versionsprüfung entfernt; Präfix `2.` statt Gleichheit; Nullwert-Modell liefert `entity_set` | andere |
| G3 | Prüffolge: Argumente und Version vor dem Kontext, der Kontext vor `entity_set` (Zusatztabelle Z1) | ein Fall je Paar von Z1 | Kontextprüfung vor der Versionsprüfung; `entity_set` vor dem Kontext | andere |
| G4 | `entitySet` nennt kein Set, ein Set mit abweichender Schreibweise oder mit Container/Namensraum, oder ein Set, dessen Typ nicht auflöst: `entity_set`, Ort der Name, `src` nicht gelesen (SPEC-126 Nr. 1 f) | Schleife Namen mal zwei Methoden; Modell mit Set ohne Typ; Reads gezählt | Ort leer; `src` vor der Prüfung gelesen | andere |
| G5 | Ein beendeter `ctx` endet mit dem Kontextfehler unverändert (`==`), ohne Code, an jeder Prüfstelle (vor jedem Read, einmal nach dem letzten Read); ein Verstoß in den schon gelieferten Bytes geht vor (SPEC-126 Nr. 1 e) | Schleife Stellen (vor Read 1, vor Read k, nach dem letzten) mal zwei Methoden, Abbruch im Read k durch den Fake | Prüfung nach dem letzten Read entfernt; Kontextfehler als `DatenFehler` | andere |
| G6 | Ein Fehler von `src` außer `io.EOF` (auch `io.ErrUnexpectedEOF`, umhüllter `io.EOF`, Kontextfehler von `src` bei lebendem `ctx`) endet das Lesen unverändert, ohne `DatenFehler`, auch nach vollständigem Dokument; Daten im selben Read werden vorher verarbeitet (SPEC-126 Nr. 2) | Schleife Fehlerarten (5) mal Zeitpunkte (vor dem ersten Byte, zwischen Reads, im selben Read mit und ohne Bytes, nach vollständigem Dokument) | Fehler als `syntax`; `errors.Is(io.EOF)` statt `==`; Fehler nach vollständigem Dokument verschluckt | andere |
| G7 | Ergebnis und Fehler hängen nicht von der Zerlegung in Reads ab (ganz, Einzelbytes, jede Zweiteilung an jeder Stelle, leere Reads, `n>0` mit `io.EOF`) (SPEC-126 Nr. 2) | Zweiteilung an jeder Stelle jedes Dokuments des Korpus (Echtdateien, Fälle von W8 bis W11, Hülle); Zählprüfung der Zweiteilungen | Zustand am Ende eines Reads zurückgesetzt; `0, nil` als Ende; `n>0` mit `io.EOF` verworfen | andere |
| G8 | Der Leser nutzt von `src` allein `Read`: kein `Close`, keine Schnittstelle der Menge O (`io.WriterTo`, `io.ByteReader`, `io.ByteScanner`, `io.RuneReader`, `io.RuneScanner`, `io.Seeker`, `io.ReaderAt`), kein Read nach `io.EOF` oder Fehler | Strom-Hülle, die alle acht Methoden führt und zählt; Summe muss 0 sein, in jedem Fall der Suite | Zusicherung `io.WriterTo`; `Close` am Ende | andere |
| G9 | Grammatik: genau ein JSON-Wert nach RFC 8259 streng; jeder Verstoß ist `syntax` (SPEC-126 Nr. 4) | Generator: Folgen über das Token- und das Zahl-Alphabet, eingebettet in `{"d":[ … ]}`; Orakel `json.Valid` plus eigene Dublettenprüfung; Zählprüfung je Klasse | `01`, `1.`, `+1` gültig; Komma vor `}` gültig; Kommentar gültig | andere |
| G10 | Kodierung: BOM (UTF-8, UTF-16 beide) ist `kodierung`, ein unvollständiges BOM `syntax`; UTF-8 streng, Surrogat-Escapes paarweise, U+0000 gültig, U+FEFF und DEL im String Zeichen (SPEC-126 Nr. 3) | Generator über die Intervallgrenzen (U+D7FF, U+E000, U+10FFFF, U+110000, überlang) mal Orte (Name, String, nicht gedeuteter Inhalt); Escapes mit `esc("…")`; Orakel nach den Regeln von Go | einzelnes Surrogat still ersetzt; BOM hinter dem ersten Byte als BOM | andere |
| G11 | Tiefe: höchstens 512 (Wurzel 1), die 513. offene Klammer ist `tiefe`, auch im nicht gedeuteten Inhalt (SPEC-126 Nr. 5) | Schleife Klammerart (Objekt, Array, gemischt) mal Position (komplexe Property, `__deferred`, `__metadata`, Navigation) mal {512, 513} | Grenze 512 zu 513 oder 511; nur Objekte gezählt | andere |
| G12 | Größe: höchstens `grenze` Bytes; das Byte `grenze+1` ist `zu_gross`, dahinter wird nichts gesehen, davor entscheidet ein Verstoß; `MaxInt64` gültig; es werden nie mehr als `grenze+1` Bytes angefordert (SPEC-126 Nr. 5) | Schleife `grenze` (1, 2, Länge-1, Länge, Länge+1, `MaxInt64`) mal Zerlegung (ganz, 1 Byte, zweiteilig); Strom-Fake zählt die angeforderten Bytes; Paare Verstoß davor/dahinter | `>` zu `>=`; Puffer größer als `grenze+1` angefordert; Überlauf bei `grenze+1` | andere |
| G13 | Doppelter Name in jedem Objekt ist `doppelter_name` (dekodiert verglichen, `a` und `A` verschieden, gleich ob die Werte gleich sind), auch im nicht gedeuteten Inhalt (SPEC-126 Nr. 4) | Schleife Objektträger (Wurzel, `d`, Element von `results`, `__metadata`, `__deferred`, Navigation, komplexe Property, unbekannter Inhalt) mal Schreibweise (gleich, Escape, Case, gleicher/anderer Wert), auch mit mehr als 16 Namen | Vergleich ohne Beachtung der Schreibweise; nur die Wurzel geprüft; Liste über 16 Namen nicht geprüft | andere |
| G14 | Rangfolge im Lesen: Lesen vor Deuten, im Lesen der erste Verstoß in Byte-Reihenfolge, Müll hinter der Wurzel `syntax` oder `zu_gross` (SPEC-126 Nr. 13 a) | Zusatztabelle Z2, ein Fall je Paar | Deuten-Fehler geht dem Lesen-Fehler vor | andere |
| G15 | Fehlerform: `*DatenFehler`, `Code()` `ODR-E6101`, `errors.Is(err, ErrDatenUngueltig)`, Ort als JSON-Pointer (`~0`, `~1`, sonst roh) in den drei Hüllen, `Error()` mit `strconv.Quote`, kein Wert im Text (SPEC-126 Nr. 14, SPEC-172 E3) | Schleife Orte (Objektform, Array-Form, `ParseEntity`, `__metadata/type`, fehlender Schlüssel) mal Namen mit `~`, `/`, `\`, `"` und Steuerzeichen | `~0` und `~1` vertauscht; Quote fehlt; Wert im Text | andere |
| G16 | Die Ursachen: jede der 14 Ursachen ohne `kontext` wird irgendwo erreicht, `kontext` nie (E1) | Schleife über die 14 Ursachen (je eine Zeile der Suite liefert sie) plus Zähler über alle Fälle der Suite: `kontext` 0 | `kontext` für eine fehlende Hülle | andere |
| G17 | Wiederholung: gleiche Bytes und Modell geben gleiche Werte; die Reihenfolge von `Entities`, `Properties`, `Uebersprungen` ist die des Dokuments, nie die einer Map (SPEC-126 Nr. 15) | Schleife 50 Wiederholungen mal Dokument mit 40 Properties und 40 übersprungenen Namen (über der Map-Schwelle) | Properties aus einer Map iteriert | andere |
| G18 | Das Ergebnis gehört dem Aufrufer; `Entities`, `Properties`, `Uebersprungen` sind nie `nil`; zwei Elemente gleichen Schlüssels stehen beide im Ergebnis (SPEC-126 Nr. 12) | Ergebnis ändern, zweiter Aufruf gleich; leere Mengen `!= nil`; Dublette | `nil` bei leerem `results`; Dubletten zusammengeführt | andere |
| G19 | Der Leser hat kein Feld, erfüllt `outbound.ODataCodec`, und kein Typ der Bibliothek steht in einer Signatur oder einem Modelltyp (SPEC-126 Nr. 15, Kopf) | `reflect` `NumField()` gleich 0; `var _ outbound.ODataCodec`; Prüfung der Signaturen | Feld `cache` im Leser | andere |
| G20 | Gleichzeitige Aufrufe sind geordnet: jedes Ergebnis gleich dem eines einzelnen Aufrufs (SPEC-126 Nr. 15) | n Goroutinen mal m Dokumente (Echtdateien, Fälle), Vergleich mit dem Einzelaufruf; `make test-race` | Paketvariable für Zähler oder Typenliste; geteilter Puffer | nebenläufig |
| G21 | Gleichzeitige Aufrufe sind unabhängig: der Fehler eines Aufrufs berührt keinen anderen | gültige und ungültige Bodies parallel, je Aufruf Erwartung (Wirkungs-Fänger zählt die Vermischung); `make test-race` | gemeinsamer Fehlerzeiger im Paket | nebenläufig |
| H1 | Eine Wurzel, die kein Objekt ist (Array, Zahl, String, `true`, `false`, `null`), ist `struktur`, Ort leer (SPEC-172 Nr. 2) | Schleife 6 Wurzeln mal zwei Methoden | Wurzel-Array als `d` gelesen | andere |
| H2 | `d` fehlt: `struktur`, Ort `/d`, gleich welche Mitglieder sonst; ein weiteres Mitglied neben `d` ist `struktur`, Ort das Mitglied, vor und hinter `d` (Nr. 2, E3) | Schleife: drei Echt-V4-Bodys, das Fehlerobjekt (Echt 24), Mitglieder `error`, `@odata.context`, `value`, `d2`, leerer Name, jeweils vor und hinter `d` | `d` optional; V4-Wurzel gelesen; weitere Mitglieder ignoriert | andere |
| H3 | `d` als Array (V1): jedes Element ein Objekt, sonst `struktur` Ort `/d/<Index>`; Anzahl und Folgelink Nullwerte; `d` weder Objekt noch Array ist `struktur` `/d` (Nr. 2 a) | Schleife Elementarten (6) mal Index (0, 1, letzter) und `d` als String, Zahl, `true`, `null` | Index im Ort fehlt; Array-Element ohne Prüfung | andere |
| H4 | `d` als Objekt: `results` ist Pflicht und ein Array von Objekten (fehlt oder falscher Typ: `struktur` `/d/results`); ein Element kein Objekt: `/d/results/<Index>`; jedes andere Mitglied außer `__count` und `__next` ist `struktur`, Ort das Mitglied (Nr. 2 a) | Schleife Typen von `results` (5), Elementarten (6) mal Index, fremde Mitglieder (`__metadata`, `__deferred`, `value`, `d`, `results2`, `@odata.count`) | fremdes Mitglied ignoriert; `results` optional | andere |
| H5 | Leeres Ergebnis: `{"d":{"results":[]}}`, `{"d":[]}` und `{"d":{"__count":"0","results":[]}}` sind gültig, `Entities` leer und nicht `nil` (Nr. 2 a, LH-FA-17 Boundary) | Echtdatei 22 und die zwei weiteren Formen, `!= nil` | `nil` statt leerer Menge | andere |
| H6 | `__count`: JSON-Zahl oder String mit JSON-Zahl, ganzzahlig 0 bis `MaxInt64`; sonst `wert` oder `typ`, Ort `/d/__count`; fehlt: Nullwert (Nr. 2 a, E2) | Schleife Schreibweisen (mindestens 30: `"6"`, `6`, `"0"`, `0`, `"9223372036854775807"`, `…808`, `"-1"`, `"1.5"`, `"1e2"`, `1e2`, `"007"`, `""`, `"abc"`, `" 6"`, `"+6"`, `true`, `null`, `{}`, `[]`, Marker) | String-Form nicht gelesen; `"007"` angenommen; Überlauf bei `MaxInt64` | andere |
| H7 | `__next`: ein nichtleerer String unverändert; Rohtext genau `[REDACTED]` ist `Redigiert`; leer ist `wert`; Nicht-String `typ`; ein Marker mit Escape ist ein gewöhnlicher String (Nr. 2 a) | Schleife Werte (absolut, relativ, mit Escapes, Marker, Marker mit `esc("005B")` aus Bytewerten, leer, `null`, Zahl, Objekt) | Marker mit Escape als `Redigiert`; leerer String erlaubt | andere |
| H8 | `ParseEntity`: `d` Objekt ist die Entity; `results`, `__count`, `__next` sind gewöhnliche Namen (Property, sonst `unbekannte_property`); `d` kein Objekt ist `struktur` `/d` (Nr. 2 b) | Modell mit und ohne Property `results`, `__count`, `__next`; `d` als 6 JSON-Typen | `__count` an einer Entity als Anzahl behandelt | andere |
| H9 | Die Methode entscheidet: ein Collection-Body an `ParseEntity` ist `unbekannte_property` `/d/results`; ein Entity-Body an `ParseCollection` ist `struktur`, Ort das erste Mitglied von `d` (Nr. 2 b, E3) | zwei Fälle mit Echt 14 und 08 | der Body bestimmt die Methode | andere |
| H10 | Rangfolge im Deuten nach E3: Hülle vor Inhalt, Mitglieder in Dokumentreihenfolge, `results` an seiner Stelle, dann das Ende (Zusatztabelle Z3) | ein Fall je Paar von Z3 | Schritte (2) und (3) von E3 vertauscht; `__metadata` nicht zuerst | andere |
| N1 | `__metadata` ist ein Objekt (sonst `struktur`, Ort `…/__metadata`); `type` ist ein String, der genau einen Entity-Typ nennt (Typ des Sets oder abgeleiteter, nicht abstrakter), sonst `odata_typ`, Ort `…/__metadata/type`; `Entity.Typ` ist der genannte Typ, die Reihenfolge der Mitglieder ist frei (Nr. 3, E4) | Schleife `__metadata` (6 JSON-Typen); `type` (kein String mal 5, leer, unbekannt, mit `#`, Alias, Leerraum, falsche Schreibweise, außerhalb der Ableitung, abstrakt, unmittelbar/mittelbar abgeleitet, Typ des Sets) mal Stelle (vor/hinter den Properties); Property des abgeleiteten Typs vor `__metadata`; Kreis in `BaseType` (`v2N1Kreis`, Modell `Y` von Hand, Schleife über 7 Typen: Selbstbezug, Ring über 2 und über 3 Typen, Anhang an einem Ring, mal 3 Formen des Bodys = 21 Läufe, Frist 10 s je Lauf): ein Typ in einer geschlossenen Kette, die den Typ des Sets nicht erreicht, ist `odata_typ`, der Leser kehrt zurück (Ringe ab Länge 4 und ein Ring mit dem Typ des Sets: ohne Fänger, akzeptiertes Negativ, siehe unten *Akzeptierte Negative nach der Verifikation*) | `#` akzeptiert; nur unmittelbar abgeleitet; `__metadata` erst nach den Properties gedeutet; Markierung `besucht` in `abgeleitet` entfernt oder nur am ersten Typ gesetzt | andere |
| N2 | Andere Mitglieder von `__metadata` (`uri`, `etag`, `id`, `content_type`, `media_src`, `edit_media`, fremde) sind ohne Wirkung, jeder JSON-Wert, und stehen nirgends im Ergebnis (Nr. 3) | Schleife Mitglieder mal 6 Wertformen; Echt 08 und 40 | `etag` als Property; `__metadata` in `Uebersprungen` | andere |
| N3 | Properties unterstützten Typs stehen in der Reihenfolge des Dokuments; eine fehlende Property ist nicht selektiert (nicht in `Properties`), `null` ist `Art` null bei `Nullable`, sonst `null` (SPEC-126 Nr. 9, 10 e) | Schleife Properties mal {fehlt, `null`, Wert}; Body-Reihenfolge verschieden von der des Modells | fehlend als `null`; `null` bei nicht `Nullable` zugelassen | andere |
| N4 | Eine Property nicht unterstützten Typs (`Edm.Int16`, `Edm.Double`, `Edm.Binary`, `Edm.Time`, komplex, `Edm.DateTimeOffset(V2)`, `Edm.Date`) ist `Uebersprungen` `typ`; ihr Inhalt wird gelesen, nicht gedeutet, jeder JSON-Wert auch `null` (Nr. 3, E4) | Schleife 8 Typen mal 7 JSON-Werte (auch tiefes Objekt) | `Edm.Int16` auf `Edm.Int32` abgebildet; `null` als Fehler | andere |
| N5 | Navigation (Objekt wie `__deferred`, expandierte Entity, `{"results":[…]}`, Array, `null`) ist `Uebersprungen` `navigation`; String, Zahl, `true`, `false` ist `typ`; der Inhalt wird gelesen, nicht gedeutet (Nr. 3) | Schleife 8 Wertformen mal (Navigation zu einer Entity, zu einer Collection); Echt 19 | `null` als `typ`; Inhalt nicht auf Dubletten gelesen | andere |
| N6 | Ein anderer Name ist `unbekannte_property`, Ort das Mitglied, auch einer mit `@`; ein Typ ist nie offen, auch bei `Open` wahr (Nr. 3, E4); Namen gelten Zeichen für Zeichen (Schreibweise zählt, keine Normalisierung) | Schleife Namen (mindestens 12: `@odata.type`, `P@x`, `@`, Case-Varianten, zerlegt/zusammengesetzt, `results`, `__count`, leer) mal Wertformen; Modell mit `Open` wahr | `@`-Namen ignoriert; Grund `offen`; Namensvergleich ohne Schreibweise | andere |
| W1 | Typ mal JSON-Typ der sieben Typen: erlaubt, `typ` oder (bei `null`) Ergebnis nach `Nullable` (Nr. 4, SPEC-126 Nr. 10 a) | Schleife 7 mal 7 Zellen gegen die Tabelle im Test (Orakel aus `SPEC-170` Nr. 4 und `SPEC-126` Nr. 10) | `Edm.Int32` als String angenommen; `Edm.Boolean` als String | andere |
| W2 | `Edm.String` und `Edm.Boolean`: dekodierte Zeichen ohne Normalisierung, leer, U+0000, `\/`; `true`, `false` (Nr. 4) | Schleife Strings (leer, U+0000 aus Bytewerten, zusammengesetzt/zerlegt, `\/`) und Booleans | Unicode-Normalisierung; `"true"` als Boolean | andere |
| W3 | `Edm.Int32`: jede Schreibweise mit ganzzahligem Wert (`1.0`, `1e2`, `100E-0`, `-0`), sonst `wert`; Bereich ±2^31 beide Seiten; als String `typ`; ein Exponent wird nicht ausgeschrieben (SPEC-126 Nr. 10 b) | Generator Zahlformen (Mantisse mal Exponent mal Vorzeichen), Orakel `math/big`, Grenzwerte; Zählprüfung | Bereich 2^31 statt 2^31-1; `1.0` abgelehnt | andere |
| W4 | `Edm.Int64`: Zahl oder String mit gleicher Deutung, Bereich ±2^63 beide Seiten, 2^53+1 behält seine Ziffern, ein String ohne Zahl (`""`, ` 5`, `+5`, `0x10`, `NaN`, `1,5`) ist `wert` | Generator wie W3 plus Stringformen; Orakel `math/big`; Zählprüfung | Rundung über `float64`; `"5"` ungleich `5` | andere |
| W5 | `Edm.Decimal`: kanonische Schreibweise ohne Exponent und ohne Nullen am Ende, `-0` ist `0`, Ziffern jenseits `float64` bleiben, `1`, `1.0`, `1e0`, `10e-1`, `"1"` sind gleich | Generator Zahlformen mit Orakel `math/big` (Rat); Zählprüfung; Gleichheitsklasse | `-0` bleibt `-0`; Nullen am Ende bleiben; Rundung über `float64` | andere |
| W6 | `zahl_zu_lang`: kanonische Schreibweise über 1024 Zeichen einer Dezimalzahl; 1024 gültig, 1025 nicht; ein Exponent (`1e999999999`) wird nie ausgeschrieben; für `Edm.Int32`, `Edm.Int64`, `__count` gilt dasselbe ohne eigene Ursache (SPEC-126 Nr. 5) | Schleife Grenzfälle mal vier Stellen (Decimal, Int32, Int64, `__count`); Speicher- und Zeitwächter der Suite | Grenze 1024 zu 1025; Exponent ausgeschrieben (Wächter schlägt an) | andere |
| W7 | `Edm.Guid`: 36 Zeichen `8-4-4-4-12`, jede Schreibweise der Ziffern, kanonisch klein; jede andere Form `wert` (Nr. 4) | Schleife über die 36 Positionen (Strich fehlt/zusätzlich, Nicht-Hex), Klammern, Leerraum, 35/37 Zeichen | Großbuchstaben bleiben; Klammern akzeptiert | andere |
| W8 | `Edm.DateTime` (Modell `Edm.DateTimeOffset`): `/Date(ms)/` ergibt `Art` `datetimeoffset` und den UTC-Text mit Bruchteil ohne Nullen; Bereich -62135596800000 bis 253402300799999, eine Millisekunde daneben ist `wert` (Nr. 4) | Generator: alle (Jahr 1 bis 9999, Monat) mal {erster Tag 00:00:00.000, letzter Tag 23:59:59.999} = 239 976 Fälle, dazu die Bruchteile 0 bis 999 (1000), die Grenzen ±1 und die Jahre 1900, 2000, 2100; Orakel: Text aus den Feldern mit `Sprintf`, Millisekunden aus `time.Date(…).UnixMilli()`; Zählprüfung 240 976 plus Grenzfälle | Millisekunden abgeschnitten; Bruchteil mit Nullen; Schaltjahrregel 100/400; Obergrenze `<` statt `<=` | andere |
| W9 | Der dekodierte Inhalt ist genau `/Date(` ganze Zahl `)/`; jeder Zusatz (`+0060`, `-0100`, `+0000`, Leerraum), `-0`, `+5`, führende Null, leere Zahl, Dezimalzahl, Exponent, Zahl über `int64` (20 Ziffern), `/date(`, fehlendes `)/` ist `wert`; ein ISO-String ist `wert`; Zahl, `true`, Objekt ist `typ` (Nr. 4, E5) | Schleife mindestens 30 Formen mal beide Methoden | Versatz akzeptiert; `-0` akzeptiert; Überlauf wickelt um | andere |
| W10 | Die Schreibweisen `\/Date(…)\/`, `/Date(…)/` und der Slash als Escape aus Backslash, `u` und `002F` ergeben denselben Text; ein Escape mitten in `Date` ebenso; unabhängig von der Zerlegung (Nr. 4) | Fälle aus Bytewerten (`esc("002F")`, `[]byte{…}`), auf Rohbytes gegengelesen; alle Zweiteilungen | nur der Rohtext verglichen | andere |
| W11 | `null` und der Marker `[REDACTED]` je Typ: der Marker (Rohtext ohne Escape) ist `Art` redigiert für `Edm.String`, `Edm.Guid`, `Edm.Int64`, `Edm.Decimal`, `Edm.DateTime`; für `Edm.Int32` und `Edm.Boolean` als String `typ`; mit Escape ein gewöhnlicher String (für die vier Formtypen `wert`) (E5, SPEC-126 Nr. 10 e, f) | Schleife 7 Typen mal {Marker, Marker mit Escape, `null`} mal {`Nullable`, nicht} | `Edm.DateTime`-Marker als `wert`; `Edm.Int32`-Marker redigiert | andere |
| K1 | Schlüsselliteral: `Edm.String` in `'…'` mit verdoppeltem `'`, jeder andere Typ sein kanonischer Text (`Edm.DateTimeOffset`: `2020-01-01T00:00:00Z`); zusammengesetzt `Name=Literal` durch `,` in der Reihenfolge des Modells, nicht des Bodys (Nr. 5, SPEC-126 Nr. 11) | Schleife 7 Typen als Einzelschlüssel; zusammengesetzt mit 2 und 3 Properties, Body-Reihenfolge permutiert | Body- statt Modellreihenfolge; `'` nicht verdoppelt | andere |
| K2 | Verletzungen des Schlüssels: Wert `null` (auch bei `Nullable`), Marker, String mit U+0000, Property fehlt, Typ ohne Schlüssel sind `schluessel`, Ort dort, wo die Property stünde; ein Fehler des Werts (`typ`, `wert`) geht einem fehlenden Schlüssel vor (SPEC-126 Nr. 11, 13 c, E3) | Schleife Schlüsselposition (1 bis 3) mal Verstoß (4) mal Typ (7); Paar Fehler plus fehlender Schlüssel (Zusatztabelle Z3) | `null` bei `Nullable` zugelassen; Ort leer | andere |
| K3 | Verschiedene Schlüssel haben verschiedene Texte, derselbe Schlüssel denselben, gleich in welcher Schreibweise (`1`, `1.0`, `"1"` bei `Edm.Int64`) und Reihenfolge (SPEC-126 Nr. 11) | Paarvergleich als Schleife über die Schreibweisen je Typ | Text ohne Typ (`1` gleich `'1'`) | andere |
| R1 | Echtfälle (Olingo Referenzszenario): sieben Antworten ergeben die erwarteten Entities (Schlüssel `'1'`, `EntryDate` `1999-01-01T00:00:00Z` bis `2010-12-01T00:00:00Z`, `Age`/`Seats`/`Version`/`Location` `typ`, `ne_*` `navigation`, `__count` 6 aus `"6"`, leer, ETag ignoriert) | Schleife über die Dateien 14, 20, 22, 08, 19, 36, 35 mit den erwarteten Entities als Literalen (nicht aus dem Leser gebildet), plus alle Zweiteilungen | Datum nicht umgerechnet; `__count` als String nicht gelesen; `__metadata` als Property | andere |
| R2 | Echtfälle mit Verstoß: `$select` ohne Schlüssel ist `schluessel`, Ort `/d/results/0/EmployeeId`; das V2-Fehlerobjekt ist `struktur` `/d`; drei V4-Bodys sind `struktur` `/d` | Schleife über Datei 18, 24 und drei V4-Dateien, je mit der Methode der Datei | Fehlerobjekt als leere Collection gelesen | andere |
| J4 | V4-Leser, V4-Schreiber, `csdl`, `urlparser` und `src/hexagon/model` bleiben byte-gleich; in `src/hexagon/ports/outbound` ändern sich allein drei V4-Dateien der Suite (`odata_codec_kodierung.go`, `odata_codec_strom.go`, `odata_codec_werte.go`) durch das Herausziehen von Helfern ohne geändertes Verhalten (§1 *Teilen mit der V4-Suite*); ihre Tests laufen unverändert grün | benannte Ausgabe: `git diff --numstat --diff-filter=MDRT <basis>..HEAD` über `jsoncodec`, `csdl`, `urlparser`, `src/hexagon/model`, `src/hexagon/ports/outbound` nennt allein diese drei Dateien (Stand beim Handoff, Zahlen in §4); `make test` | beliebige Änderung an `jsoncodec/scan.go`; ein Helfer der V4-Suite, der eine Zusage der V4-Suite abschwächt (Mutation am V4-Leser `jsoncodec`, rot in der V4-Suite, siehe Absatz *Mutationen*) | andere |

**Mutationen des Implementers** (`AGENTS.md` §3.10; Stand Handoff, Zusage · Mutation · roter Test). Messender: der Implementer, je Mutante eine Ersetzung im Produktcode `jsoncodecv2` auf einer Wegwerfkopie von `src/` im Scratchpad, Image `odata-recorder:test` (`--network none --memory=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`), `go test -count=1 -run TestVertrag` über das Paket; „rot“ heißt: die genannte Zeile der Suite meldet `--- FAIL` (weitere rote Zeilen sind nicht aufgelistet, sofern sie nicht die Zusage selbst sind). Jede Mutante baut und läuft; eine gleichwertige Mutante ist keine Zeile.

| Zeile | Mutation | roter Test |
|---|---|---|
| Nr. 1 | `!p.Unterstuetzt() \|\|` aus `property` entfernt | N4 |
| Nr. 2 | `lesText(k, model.ArtDateTimeOffset, …)` mit `model.ArtString` | W8 (Art) |
| Nr. 3 | `jsoncodecv2` importiert `jsoncodec` (vom Verifier auf einer Wegwerfkopie gefahren) | `TestKeinPaketImportiertEineAndereVersion` (`versionstrennung_test.go`) |
| G1 | `grenze < 1` zu `grenze < 0` | G1 |
| G2 | Version `2.0` durch Präfix `2.` ersetzt | G2 |
| G3 | Kontextprüfung vor die Argumentprüfung gestellt | G3 |
| G4 | Ort des `entity_set`-Fehlers leer | G4 |
| G5 | zweite Kontextprüfung (nach dem letzten Read) entfernt | G5 |
| G6 | Fehler von `src` als `syntax` gemeldet | G6, G14 |
| G7 | `n == 0` ohne Fehler als Ende gewertet | G7 |
| G8 | `Close` am Ende gerufen, wenn `src` es führt | G8 |
| G9 | Tabulator aus dem Leerraum entfernt | G9 |
| G10 | Führungsbyte-Bereich `0xE1..0xEF` zu `0xE1..0xEE` | G10 |
| G11 | `>= maxTiefe` zu `> maxTiefe` | G11 |
| G12 | `n > rest` zu `n >= rest`; `want = rest + 2` statt `+ 1` | G12 (beide), G1, G8 |
| G13 | Dublettenprüfung abgeschaltet (`false && f.hat(text)`) | G13 |
| G14 | Verstoß in den Bytes vor `grenze+1` unterdrückt, wenn der Read über die Grenze geht (`&& !ueber`) | G7, G12; G14 selbst bleibt grün (das Paar *`zu_gross` gegen Verstoß davor* fängt G12, siehe Z2) |
| G15 | `~0` und `~1` vertauscht | G15 |
| G16 | `kontext` für die fehlende Hülle | G16 |
| G17 | Property vorn statt hinten angefügt (Reihenfolge des Dokuments) | G17, N3 |
| G18 | `Uebersprungen` einer Entity `nil`; `results` leer als `nil` | G18 (beide), N2, H5, R1 |
| G19 | Feld `cache` im Leser | G19 |
| H1 | Prüfung der Wurzel (Objekt) abgeschaltet | H1 |
| H2 | weitere Mitglieder neben `d` ignoriert | H2 |
| H3 | Index im Ort der Array-Elemente fehlt | H3 |
| H4 | fremdes Mitglied in `d` ignoriert | H4 |
| H5 | `results` leer ohne Nicht-`nil` | H5 |
| H6 | `__count` negativ erlaubt (`v < 0` aus) | H6 |
| H7 | `__next` leer erlaubt | H7 |
| H8, H9 | unbekannte Property ignoriert (`return nil`) | H8, H9 |
| H10 | `__metadata` nicht zuerst gedeutet (Name im ersten Durchlauf nie gleich) | H10, N1 |
| N1 | abstrakten Typ zugelassen; Markierung `besucht` weg oder nur am ersten Typ (Kreis) | N1 (Kreis: Frist von 10 s) |
| N2 | `__metadata` im zweiten Durchlauf nicht übersprungen | N1, N2 |
| N3 | `null` bei nicht `Nullable` zugelassen | N3, W1 |
| N4 | `Edm.Date` nicht übersprungen | N4 |
| N5 | `null` bei Navigation als `typ` | N5 |
| N6 | unbekannter Name ignoriert | H8, H9, G16 |
| W1 | `Edm.Int32` als String angenommen | W1, W3 |
| W2 | `Edm.Boolean` als String angenommen | W2 |
| W3 | `ParseInt(…, bits+1)` | W3 (und V4: F3) |
| W4 | Int64 mit 63 Bit gelesen | W4 |
| W5 | `-0` bleibt `-0` | W5 |
| W6 | `laenge >= maxDezimalLaenge` | W6 |
| W7 | Guid nicht in Kleinbuchstaben | W7 |
| W8 | `ms >= zeitMax`; `TrimRight` entfernt | W8 (beide) |
| W9 | `-0` angenommen | W9 |
| W10 | Escape `\/` im String nicht zugelassen | W10 |
| W11 | Marker auch für `Edm.Int32` | W11 |
| K1 | `'` nicht verdoppelt | K1 |
| K2 | `null` als Schlüssel zugelassen | K2 |
| K3 | Schlüsseltext ohne Anführungszeichen für Strings | K3 |
| R1 | Zeitpunkt um eine Stunde verschoben | R1 |
| R2 | Ursache der fehlenden Hülle `kontext` | R2 (zusammen mit G16) |
| J4 | Änderung an `jsoncodec` (benannte Ausgabe des `git diff`) | keine Mutation im Test; die Ausgabe nennt die drei Dateien der Suite |

Nicht gefahren (kein roter Test gesehen): G9 für die übrigen Zeichen der Grammatik (nur der Tabulator ist gemutet; der Generator zählt die Folgen), G14 für jedes Paar von Z2 einzeln (nur das genannte Paar), H8 und H9 mit eigener Mutation (rot nur durch die Mutation von N6), K1 mit Body- statt Modellreihenfolge, R2 mit einem Fehlerobjekt als leere Collection. Das sind Zusagen mit Fänger in der Tabelle; keine ist ein Negativ. Der Verifier hat sie danach selbst gefahren und rot gesehen (G9 jenseits des Tabulators 21 von 21, G14 je Paar von Z2, H8, H9, K1 mit Body-Reihenfolge und R2 mit dem Fehlerobjekt als leere Collection; Report `docs/reviews/2026-11-05-slice-v2-payload-lesen-verifikation.md`).

**Mutationen an den geteilten Helfern** (J4; die V4-Suite ist der Fänger der V4-Seite): Mutation am V4-Leser `jsoncodec` auf einer Wegwerfkopie, Lauf `go test -run TestODataCodecVertrag ./src/bootstrap/` im Image `odata-recorder:test`: UTF-8-Bereich `0xE1..0xEE` rot in B3 (Helfer `uPruefer`, `b3Alphabet`), Fehler von `src` als `syntax` rot in A3 (Helfer `a3Quellfehler`), `ParseInt(…, bits+1)` rot in F3 und F4 (Helfer `int32Werte`). Auf der V2-Seite derselben Helfer: G10, G6 und W3 (siehe oben). Alle sechs Aufrufstellen der drei Helfer sind damit rot gesehen.

**Raten der Zeilen G20 und G21** (`AGENTS.md` §3.10; Messender: der Implementer; Image `odata-recorder:race`, gebaut von `make test-race` (Go mit C-Compiler, cgo), Test-Binary mit `go test -race -c`, danach je Mutante 30 Prozesse nacheinander mit `-test.count=1 -test.run TestVertrag/G20` bzw. `/G21`, rot heißt Exit ungleich null; der Detektor meldet dasselbe Rennen je Prozess einmal, ein `-count=20` zeigt es nicht). Die Mutanten sind Paketvariablen im Leser, denn der Leser hat kein Feld (G19):

| Mutation | G20 (Prozesse rot) | G21 (Prozesse rot) |
|---|---|---|
| C1: Paketvariable `zaehlerM`, in `vorbereiten` erhöht | 30 von 30 | 30 von 30 |
| C2: Lesepuffer als Paketvariable (`[32 << 10]byte`) | 30 von 30 | 30 von 30 |
| C3: Paketvariable `letzter` mit dem Fehler von `lies` beschrieben | 30 von 30 | 30 von 30 |
| C4: Vermischung unter Sperre (Mutex, `letzter` liefert dem nächsten gültigen Aufruf den Fehler eines früheren), also kein Datenrennen | 20 von 20 Läufen eines Prozesses mit `-count=20`, ohne `-race` | 20 von 20 Läufen mit `-count=20`, ohne `-race`; unter `-race`: 30 von 30 Prozesse rot, 0 Meldungen des Detektors |

Der Detektor fängt die Mutationen C1 bis C3 in jedem Prozess (30 von 30 je Zeile und Mutante); die Vermischung unter einer Sperre meldet er nicht (0 Meldungen), sie fängt der Wirkungs-Fänger (G21 vergleicht jedes Ergebnis mit dem eines einzelnen Aufrufs). Der Leser hat keine Sperre und gibt keine zwischen Prüfen und Handeln frei; ein Raster je Fenster der Sperre entfällt. Ungesehen bleibt, was der Detektor nicht meldet: ein Zugriff, den kein Lauf der Suite ausführt (Folge: ein Rennen in einem Pfad, den weder Echtdatei noch Fall der Suite berührt, bliebe grün; die Klasse der Bodies deckt die Pfade des Deuters, nicht jede Kombination).

**Akzeptierte Negative nach der Verifikation** (V-1 des Verifikations-Reports; nicht behoben, keine weitere Fix-Runde: Review und Verifikation laufen je einmal, Entscheidung des Users *Durchstich vor Härtung*). Zwei Sätze von `SPEC-172` Nr. 3 zum Kreis in `BaseType` haben in der Suite keinen Fänger:

- Zusage: „die Länge des Rings ist ohne Wirkung“, also auch ein Ring ab Länge 4 endet mit `odata_typ`. Fänger `N1` (`v2N1Kreis`): Selbstbezug, Ringe über 2 und 3 Typen und ein Anhang an einem Ring, Modell `Y` mit 7 Typen. Mutation, die grün bleibt: der Abbruch wirkt nur bis zur Ringlänge 3 (der Verifier leerte die Markierung bei 4 Einträgen), ein Ring ab Länge 4 läuft ohne Abbruch weiter.
- Zusage: „erreicht die Kette den Typ des Sets, ist der Typ abgeleitet“, auch wenn die Kette in einem Ring liegt, der den Typ des Sets enthält. Das Modell `Y` führt keinen Ring mit dem Typ des Sets. Mutation, die grün bleibt: eine Prüfung des Kreises vor der Suche nach dem Typ des Sets (ein Ring mit dem Typ des Sets würde `odata_typ` statt abgeleitet).

Grund: Der **Code** ist nach dem Probe-Paket des Verifiers richtig (Ringe der Länge 1 bis 12 × Typ des Sets außerhalb oder im Ring × Anhänge 0 bis 4 × beide Methoden, 3456 Läufe, 0 Abweichungen; die drei Mutanten *Markierung entfernt*, *nur am ersten Typ*, *Ring ab Länge 4* sind im Probe-Paket rot). Es fehlt allein der Fänger in der Suite; die Menge der Ringe ist unbeschränkt, die Prüfgrenze steht in §6 (*Prüfgrenze des Kreises*). Folge im Betrieb: eine Änderung an `abgeleitet` (Markierung, Reihenfolge der Prüfung), die nur Ringe ab Länge 4 oder einen Ring mit dem Typ des Sets bricht, bliebe in `make gates` grün. Erreichbar ist ein Kreis nur über ein von Hand gebautes Modell (`model.NeuesServiceModel` prüft ihn nicht, der CSDL-Leser weist ihn ab, `SPEC-125` Nr. 9 a); ein importierter Dienst trägt ihn nicht. Adresse für den Nachzug des Fängers: weiter offen beim Planner im Register und in `welle-odata-v2` (Zuweisung bei der Closure der Welle; kein Plan auf `main` nimmt den Punkt an, deshalb keine Sendung nach `AGENTS.md` §3.13). Dies ist kein behobener Befund.

INFO (V-4 der Verifikation, kein Code): in `v2G6` trägt der Aufruf von `a3Quellfehler` das Label „mitten in einer Zahl“, schneidet aber vor der Ziffer (`"ID":1`, Versatz 5). Die Klasse deckt `v2G6Zeitpunkte` über jede Position (die Mutante *Fehler mitten in einer Zahl verschluckt* ist rot); an der Stelle steht nur ein ungenaues Label.

**Lint (`make lint`, Befund des Reviews):** die elf Meldungen (`cyclop`, `gocyclo` für `v2G18`; `gocognit` für `v2G6` und `v2FaelleG9Bytes`; `dupl` über drei Paare, `unparam` an `b3Alphabet`) sind durch Umbau behoben, ohne `//nolint` und ohne Änderung von `.golangci.yml`: `v2G18Leer`, `v2G9ByteOrte`, `v2G6Zeitpunkte` und die geteilten Helfer `uPruefer`, `a3Quellfehler`, `int32Werte`.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (V2 schreibt Zeitpunkte als `\/Date(…)\/`) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

**Anweisungen des Architect für den Implementer** (Reihenfolge, Echtdateien, Aufruf; unverändert übernommen):

**Reihenfolge für den Implementer:** (1) `git merge main` nach der Integration von `slice-v2-metadaten` (bringt `versionstrennung_test.go` und `csdlv2`; liegt der Test dann, ändert der Implementer ihn nicht); (2) Rohdateien der Echtfälle byte-genau als Fixtures ablegen (`Write`, kein Heredoc, kein `cp` ins Repo), danach mit `cmp` gegen die Quellen prüfen und die SHA-256 unten vergleichen; (3) Suite vor Produktcode, jede Zeile der Tabelle rot sehen; (4) Mutationen je Zeile selbst fahren und in einem Absatz von §3 festhalten (Zusage · Mutation · roter Test, Raten mit Zahl der Prozesse bei G20 und G21 mit `make test-race`, 30 Prozesse je Mutation).

**Echtdateien** (Quelle `/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/odata-testserver/`, Pfad `out-v2-decoded/<Name>.body`, SHA-256 in Klammern, kein Zeilenende am Dateiende):
14-top (109526173fb866c081338e655303379a990cef8a58020b428d27b9f99a61019b, 2132 Bytes) · 20-inlinecount (de5de6fa9b05dcf000e379a484be5cd9e6d5f041876a8f91d5ba470d7a3c7774, 2146) · 22-empty-result (cda900e0bd7fb6118c2c63cc5f77ce2a6cce5524c695b77af4cbe82676a5bcc2, 20) · 08-entity-key-json (ff13bed873e035d32c875933821be535c2b37d84ca8bd060438fa8de68dd7f2a, 1062) · 19-expand (abb5493888680491ab4b44ca67af066b0a7bac894103f016a33b0df8530e0f38, 1299) · 36-rooms-json (68ec21f9c0a9e5f190532b314b1c3f868aeca6ba113c132ec9807b39af44dda8, 481) · 35-teams-json (d5cae4237461728e2af891ab6844f300d9c9bbc4b24b0f9b8fe38d534da03b42, 1017) · 18-select (87ef9da4cafb2ed2c41184cd585c2877ab223df023371186b0aee72d4b431958, 2561) · 24-notfound-entity (d33cbc6f327222e9512b24d278e4522c319ce2ac5124cff56331a587345991e3, 94). V4: `out-v4olingo/04-collection-default.body` (4f1a8cad8b32a97b1cf5ca10e70d7591e79f0d72dc00b7b7b5a7f5835910f196, 441), `out-v4olingo/08-entity-key.body` (33aa541982e0f8527a00bbfa1a4568ea29d4ad1c35314607d807fb1ff577428e, 121), `out-v4net-decoded/21-count-true.body` (e792dd6e4e98fdca734c49e610aaf7e31f515f28ca8c94e33d8f408a9f8f67f9, 375).
Modelle der Echtfälle baut die Suite von Hand mit `model.NeuesServiceModel("2.0", …)` nach `03-metadata.body` (Employee: `EmployeeId` `Edm.String` Schlüssel nicht `Nullable`, `EmployeeName`, `ManagerId`, `RoomId`, `TeamId` String, `Location` komplex, `Age` `Edm.Int16`, `EntryDate` `Edm.DateTimeOffset` `Nullable`, `ImageUrl` String, Navigationen `ne_Manager`, `ne_Team`, `ne_Room`; Room: `Id`, `Name`, `Seats`/`Version` `Edm.Int16`, Navigationen; Team: `isScrumTeam` `Edm.Boolean`); die Tabelle ist die Quelle, nicht `csdlv2` (keine Voraussetzung).

**Aufruf der Suite:** `jsoncodecv2/leser_test.go`: `outboundtest.ODataCodecV2Vertrag(t, func() outbound.ODataCodec { return jsoncodecv2.New() })` (kein Metadaten-Leser nötig).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` (erfüllt), und der Vertrag samt Randformen steht in `SPEC-172` Nr. 1 bis 5 mit den Ersatzwortlauten E1 bis E6 (erfüllt, Commit `20b0d1a`); die Tabelle *Zusage · Fänger* und das Raster stehen in §3 und §6 (erfüllt).

- **Code-Abhängigkeit, mit Grund:** keine auf einen offenen Slice. Der Leser steht hinter dem Port `ODataCodec` (`done/`) und braucht Modell und Fehlermodell (`done/`); das V2-Modell ist derselbe Typ `model.ServiceModel` ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)), die Tests bauen es mit `model.NeuesServiceModel("2.0", …)`, der V2-Metadatenleser ist keine Voraussetzung. Der Schnitt von Lesen und Schreiben ist ein Teil ohne geteilten Mechanismus (`AGENTS.md` §3.14): sie teilen allein die Tabelle der Typen in der Spezifikation (`SPEC-170` Nr. 4, `SPEC-172` Nr. 4), der Rundlauf beider liegt im Schreiben.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Workspace, Composition Root, M2 und `slice-bootstrap-m3` (der Codec arbeitet auf Bytes, Modell und Entities; der Rundlauf über Import und Mock gehört `slice-v2-bootstrap`), `slice-v2-query`, `slice-v2-metadaten`, `slice-odata-json-schreiben`, `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-expand-batch`. Das heißt: **startbar ohne `welle-mock` und ohne M2, sobald [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted` ist.**
**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`), aus der Tabelle in §3 (57 Zeilen *Zusage · Fänger*, davon *a* = 2 nebenläufig (G20, G21) und *b* = 55 andere; *f* = 0 Port-Fakes, denn die Strom-Hüllen der Suite sind Hilfen des Tests; *p* = 1400 Produktcode, die ehrliche Erwartung des Architect gegen den V4-Leser mit 1507 Zeilen, Herleitung: Scanner 517, Baum 24, Leser 113, Zahl 160, Guid 40, Wert 102, Zeit 70, Deuter 370): *S* = 85·2 + 80·55 + 200·0 + 1400 = **5970**, Marke *M* = 100 · ⌈135·5970 / 10000⌉ = **8100**. Die vorläufige Prosa-Marke 6100 (Grobschätzung mit 45 Zeilen und Produktcode 900, die Zahl nach Hebung des Scanners) ist durch den Block ersetzt, weil die Tabelle des Architect 57 Zeilen hat und der Produktcode ohne Hebung 1400 beträgt. Die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `versionstrennung_test.go` steht nicht darin (er kommt per Merge von `slice-v2-metadaten`, `SPEC-128` Nr. 10 g). Das Ende wird zwischen 0,8 und 1,2 mal *S* erwartet (`csdlv2` endete bei 0,80). `basis` ist `20b0d1a`, der Commit der Ersatzwortlaute in `SPEC-172`, gesetzt vom Planner vor dem ersten Code-Commit.

**Schnittfrage** (*S* ≥ 2223, hier gestellt und entschieden): Entity und Collection teilen Scanner, Hülle (`d`, `results`, `__count`, `__next`), Deuter und die Tabelle der Typen; ein Teil ohne geteilten Mechanismus ist nicht abtrennbar, das Schreiben ist bereits vor dem Code geschnitten (`slice-v2-payload-schreiben`). Die Alternative, den JSON-Scanner aus `jsoncodec` zu heben und von beiden Versionen zu teilen, ändert den V4-Pfad (Verschieben zählt doppelt, rund 1060 gegen 517 Zeilen der Kopie), widerspricht der DoD (V4-Leser unberührt) und dem Importsatz von `SPEC-172`; [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) erlaubte ein versionsfreies Paket, wenn die V4-Tests unverändert bleiben, verlangt es nicht. Entschieden: eigener Scanner in `jsoncodecv2`. Damit steht der Grund im Block (`schnitt-grund`), warum kein Teil vor dem Code abgetrennt wird.

```groessenmarke
pfad: src/adapters/driven/odata/jsoncodecv2
pfad: src/hexagon/ports/outbound/outboundtest/odata_codec_v2_vertrag.go
pfad: src/hexagon/ports/outbound/outboundtest/odata_codec_v2_geruest.go
pfad: src/hexagon/ports/outbound/outboundtest/odata_codec_v2_lesen.go
pfad: src/hexagon/ports/outbound/outboundtest/odata_codec_v2_huelle.go
pfad: src/hexagon/ports/outbound/outboundtest/odata_codec_v2_entity.go
pfad: src/hexagon/ports/outbound/outboundtest/odata_codec_v2_werte.go
pfad: src/hexagon/ports/outbound/outboundtest/odata_codec_v2_zeit.go
pfad: src/hexagon/ports/outbound/outboundtest/odata_codec_v2_schluessel.go
pfad: src/hexagon/ports/outbound/outboundtest/odata_codec_v2_echt.go
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_echt_14-top.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_echt_20-inlinecount.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_echt_22-leer.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_echt_08-entity.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_echt_19-expand.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_echt_36-rooms.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_echt_35-teams.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_echt_18-select.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_echt_24-fehler.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v4_fremd_olingo-04.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v4_fremd_olingo-08.json
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v4_fremd_net-21.json
produktcode: 1400
fakes: 0
schnitt-grund: Entity und Collection teilen Scanner, Hülle und Deuter, und die Tabelle der Typen; das Schreiben ist vor dem Code abgetrennt (slice-v2-payload-schreiben); die Hebung des Scanners in ein versionsfreies Paket ändert den V4-Pfad (Verschieben zählt doppelt, rund 1060 gegen 517 Zeilen der Kopie), widerspricht der DoD (V4 unberührt) und dem Importsatz der Spezifikation; Vorbild csdl, json und csdlv2, ungeschnitten geliefert
basis: 20b0d1a
schaetzung: 5970
marke: 8100
```

**Messung beim Handoff** (Metrik des Blocks: Summe aus Einfügungen und Löschungen von `git diff --numstat 20b0d1a..<letzter Commit des Slice>` über die Pfade des Blocks): 4085, also 0,68 mal *S* (5970) und unter der Marke 8100; `make groessenmarke` meldet `ok, Messung 4085 von Marke 8100`. Der Review nennt 7323 (1,23 mal *S*); der Befehl über die Pfade des Blocks und das Gate messen davor, beim Stand `ead28d1`, 4078 und beim Handoff 4085. Der Unterschied zur Zahl des Reviews liegt in der Metrik, nicht im Code: die 7323 des Reviews zählen die ganzen Verzeichnisse `jsoncodecv2` und `outboundtest`, nicht die Pfade des Blocks; darin liegen 3245 Zeilen aus `slice-v2-metadaten` (`metadata_codec_v2_*.go`, drei `v2_*.xml`), die der Merge `9650388` hinter die Basis `20b0d1a` in den Zweig brachte: 7323 − 3245 = 4078 (Stand des Reviews), plus 7 Zeilen der Änderung nach dem Review = 4085 (Ursache vom Verifier geklärt, kein Fehler des Codes; die Messung des Gates ist die der Marke). Der Zuwachs dieser Änderung gegen den Stand des Reviews beträgt 7 Zeilen; die drei V4-Dateien der Suite (J4) liegen nicht im Block und zählen nicht mit (Änderung dort: 56 Einfügungen und 41 Löschungen in `odata_codec_kodierung.go`, `odata_codec_strom.go`, `odata_codec_werte.go`).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 8100 (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat 20b0d1a..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Der Schnitt zwischen Lesen und Schreiben ist vor dem Code gemacht; die Schätzung ist *S* = 5970, die Schnittfrage von `SPEC-128` ist gestellt und entschieden (eigener Scanner, Grund im Block).
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`.

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

- Der Import eines V2-Dienstes braucht mehr als den Leser: die Auswahl der Antworten (`SPEC-127` Nr. 8: heute `application/json` mit `odata.metadata`) und die Ketten über Folgelinks und Anzahl (`SPEC-127` Nr. 12) sind V4-Regeln in `done/`; die Anpassung trägt `slice-v2-import-codecs` (`SPEC-175` Nr. 4) — **Ausgang:** eingetreten: `slice-v2-import-codecs`.
- Das Format der V2-Antworten (JSON oder Atom) steht im Lastenheft nicht — **Ausgang:** entfallen: JSON, Atom außerhalb des Umfangs (`SPEC-170` Nr. 2).
- Menge der Bodies unbeschränkt: Ausgang *entschieden*: Generatoren G9, G10, W3 bis W5, W8 mit Orakel und Zählprüfung, Rest als akzeptiertes Negativ (oben).
- Zeit- und Zahlenschreibweisen: Ausgang *entschieden*: Zeilen W3 bis W11, Raster-Zelle *Typ-Schreibweisen*.
- Fakes: Ausgang *entschieden*: Strom-Fakes zählen jeden Read, jede angeforderte Byte-Zahl, jede Methode der Menge O und `Close`; keine Port-Fakes (f = 0).
- Randformen ohne Ort: Ausgang *entschieden* mit E1 bis E6, eingetragen in `SPEC-172` vor dem ersten Code-Commit (Commit `20b0d1a`).
- Kopplung an die V4-Suite: die neue Suite nutzt Hilfen im selben Paket (`jsuite`, Strom-Hüllen, Generator-Rahmen `gFahre`/`gZaehle`); ändert ein späterer Slice diese Hilfen, bricht der Bau der V2-Suite. Ausgang: *eingetreten* und im Slice gelöst: `make lint` meldete vier `dupl`-Paare und ein `unparam` zwischen den V2-Dateien und den V4-Dateien der Suite; die Lösung teilt drei Helfer (§1 *Teilen mit der V4-Suite*), die Kopplung ist damit größer, nicht kleiner. Folge klein (Bauabbruch, kein stilles Grün); ob die Kopplung ins Beobachtungs-Register wandert: *entschieden* (Planner, Closure): nein; ein Auftreten, die Folge ist laut (Bauabbruch) und die Kopplung ist im Plan (§1) benannt; ein weiteres Auftreten in einem anderen Slice trägt die Beobachtung nach.
- Abweichung vom Plan (J4): die V4-Dateien der Suite `odata_codec_kodierung.go`, `odata_codec_strom.go` und `odata_codec_werte.go` sind nicht byte-gleich geblieben (Helfer herausgezogen, Verhalten unverändert; Zahlen in §4). Ausgang: *entschieden* durch den Architect (2026-11-05): das Teilen ist zulässig, der Wortlaut von J4 und der DoD-Zeile *Lesen* gilt unverändert. Grund: [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) trennt die Adapter-Pakete (kein V2-Paket importiert ein V4-Paket) und sagt in §Konsequenzen, Gemeinsames dürfe in einem Paket ohne Versionsnamen liegen, wenn die V4-Tests unverändert bleiben; `outboundtest` liegt unter `src/hexagon/ports/outbound`, ist keiner der beiden Adapter, und die Helfer ändern kein Verhalten (V4-Tests grün, Mutationen in §3). Es ist kein Schnitt nachzuholen; der Reviewer der nächsten Prüfrunde löst J4 in der geänderten Fassung auf. Die Alternative ohne Änderung der V4-Dateien, eine Ausnahme in `.golangci.yml` (`exclusions.rules` für `dupl`), wäre ein ADR (`AGENTS.md` §3.6) und entfällt.
- Kreis in `BaseType` (Zeile N1, `v2N1Kreis`): ein von Hand gebautes `model.ServiceModel` kann eine geschlossene Kette von `BaseType` tragen (`model.NeuesServiceModel` prüft sie nicht, der CSDL-Leser weist sie mit `nicht_aufloesbar` ab, `SPEC-125` Nr. 9 a). Der Leser folgt der Kette mit einer Markierung je Typ (`abgeleitet`) und behandelt einen Typ, dessen Kette sich schließt, ohne den Typ des Sets zu erreichen, wie jeden Typ außerhalb der Ableitung (`odata_typ`, Regel E4, kein neuer Fall). Ausgang: *entschieden* durch den Architect (2026-11-05), eingetragen in `SPEC-172` Nr. 3 (Historie vom selben Tag): der Leser kehrt bei jedem Modell zurück, ein Typ in einer geschlossenen Kette ohne den Typ des Sets ist `odata_typ`, erreicht die Kette den Typ des Sets, ist der Typ abgeleitet, die Länge des Rings ist ohne Wirkung (Ringe ab Länge 4 gelten wie kürzere). Der Code tut das, es folgt keine Codeänderung; die Prüfgrenze unten bleibt als akzeptiertes Negativ.
- Merge-Konflikt in `spec/spezifikation.md`: `slice-v2-metadaten` ändert `SPEC-171`/`SPEC-172` *Ungeprüft* und die Historie. Ausgang: *eingetreten* beim zweiten Merge von `main` (Closure, 2026-11-05, Merge-Commit `443c7da`, nach `slice-v2-query`, `slice-v2-query-filter`, `slice-text-antworten` und `slice-m2-referenzdienst`): ein Konflikt allein in der Tabelle *Historie* von `spec/spezifikation.md` (zwei Seiten fügten Zeilen am selben Ort ein), gelöst mit beiden Seiten, Zeile dieses Zweigs oben; kein Konflikt in `SPEC-171` oder `SPEC-172`. Der erste Merge (`9650388`) und der Stand beim Handoff (`git merge-tree --write-tree HEAD main`) waren ohne Konflikt.
- Marke: *S* = 5970, *M* = 8100; das Ende wird zwischen 0,8 und 1,2 mal *S* erwartet; `ueberschreitung` nur mit Grund (`SPEC-128` Nr. 5 d). Ausgang: *entfallen*, die Marke ist nicht überschritten; die Messung steht in §4.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Stand: vom Architect gefüllt, keine Zelle ohne Ort, keine „gilt nicht“). `E1` bis `E6` sind die Ersatzwortlaute, die vor dem ersten Code-Commit in `SPEC-172` stehen: Nr. 1 (E1), Nr. 2 a (E2), Nr. 2 c (E3), Nr. 3 (E4), Nr. 4 (E5) und der Absatz *Ungeprüft bleibt beim Lesen außerdem* (E6). Die Zellen sind damit *entschieden*; der Implementer entscheidet im Code nichts und hält an einer Randform ohne Ort an (`AGENTS.md` §3.12).

| Operation | Body leer oder keine Menge | falsche Hülle | unbekannte Property, nicht unterstützter Typ | Typ-Schreibweisen, Zahlen jenseits `float64` | doppelter Name, über Grenze | Wiederholung, Abbruch | Modell (Version, Nullwert) | Mitglieder der Hülle (`__count`, `__next`, `__metadata`) |
|---|---|---|---|---|---|---|---|---|
| Lesen einer Entity | leer, nur Leerraum: `syntax`; nur BOM: `kodierung`/`syntax` (`SPEC-172` Nr. 1, `SPEC-126` Nr. 3, 4); `{"d":{}}`: `schluessel` (`SPEC-126` Nr. 11) | `struktur` `/d` (Nr. 2, E3); Collection-Body: `unbekannte_property` `/d/results` (Nr. 2 b, E3) | `unbekannte_property` oder `Uebersprungen` `typ`/`navigation` (Nr. 3, E4) | Nr. 4 mit E5; `zahl_zu_lang` (Nr. 1, `SPEC-126` Nr. 5) | `SPEC-126` Nr. 4 und 5 (`doppelter_name`, `zu_gross`, `tiefe`) über Nr. 1 | Nr. 1 (`SPEC-126` Nr. 1, 2, 15) | `ErrUngueltig` (Nr. 1, E1) | `__metadata`: Nr. 3, E4; `__count`, `__next` sind Namen (Nr. 2 b) |
| Lesen einer Collection | `{"d":{"results":[]}}`, `{"d":[]}` gültig, `Entities` leer und nicht `nil` (Nr. 2 a); Body leer: `syntax` | `struktur` `/d` (Nr. 2, E3); Entity-Body: `struktur`, Ort erstes Mitglied (E3) | wie Entity (Nr. 3, E4) | wie Entity; `__count` nach E2 | wie Entity | wie Entity | `ErrUngueltig` (Nr. 1, E1) | `__count` (Nr. 2 a, E2), `__next` (Nr. 2 a, `SPEC-126` Nr. 8 b), `__metadata` je Element (Nr. 3, E4) |

**Zusatztabellen (kollidierende Verstöße, Modi)** — Spalte *Fänger* nach `AGENTS.md` §3.16; die Zeilen G3, G14, H10 und K2 der Tabelle in §3 führen sie.

**Z1 — Prüffolge der Argumente** (Zeile G3; Orte `SPEC-126` Nr. 1, E1)

| Paar | Entscheidung | Fänger |
|---|---|---|
| Version falsch und `ctx` beendet | `ErrUngueltig` (Argument-Klasse geht dem Kontext vor) | ein Fall |
| Version falsch und `entitySet` unbekannt | `ErrUngueltig` | ein Fall |
| Nullwert-Modell und beliebiges Set | `ErrUngueltig`, nicht `entity_set` | ein Fall |
| `ctx` beendet und `entitySet` unbekannt | Kontextfehler unverändert (`SPEC-126` Nr. 1 f folgt auf e) | ein Fall |
| `ctx` nil und Version falsch | `ErrUngueltig` (welcher von beiden, nicht zugesagt) | ein Fall |
| Argumentverstoß und Fehler von `src` | `src` nicht gelesen, `ErrUngueltig` | ein Fall (Reads 0) |

**Z2 — Rangfolge im Lesen** (Zeile G14; `SPEC-126` Nr. 13 a)

| Paar | Entscheidung | Fänger |
|---|---|---|
| doppelter Name früher, `syntax` später | `doppelter_name` | ein Fall je Reihenfolge (beide) |
| `kodierung` und `tiefe` | der erste in Byte-Reihenfolge | beide Reihenfolgen |
| `zu_gross` am Byte `grenze+1` gegen Verstoß davor/dahinter | davor entscheidet der Verstoß, dahinter `zu_gross` | zwei Fälle je Verstoßart |
| `syntax` am Ende gegen `struktur` am Anfang | `syntax` (Lesen vor Deuten) | ein Fall |
| `doppelter_name` tief in einer Navigation gegen früheres `unbekannte_property` | `doppelter_name` | ein Fall |
| Müll hinter der Wurzel gegen ungültiges UTF-8 dort | `syntax` | ein Fall |
| Fehler von `src` im selben Read gegen Verstoß des Lesens in den Daten | der Verstoß geht vor, bei unvollständigem Dokument der Fehler von `src` | zwei Fälle |

**Z3 — Rangfolge im Deuten** (Zeilen H10, K2; E3)

| Paar | Entscheidung | Fänger |
|---|---|---|
| `d` fehlt und weiteres Mitglied (V4-Body mit `@odata.context`) | `struktur` `/d` | Echt-V4-Body |
| weiteres Mitglied neben `d` (vor und hinter `d`) und Fehler in `d` | das Mitglied | zwei Fälle |
| `results` falscher Typ und fremdes Mitglied in `d`, beide Reihenfolgen | das erste im Dokument | zwei Fälle |
| `__count` falsch vor `results` und Fehler im Element; `__count` hinter `results` | vor `results`: `__count`; dahinter: das Element | zwei Fälle |
| `__next` falsch, beide Seiten des Elements | wie `__count` | zwei Fälle |
| Fehler in Element 0 und Element 1 | Element 0 | ein Fall |
| `odata_typ` aus `__metadata` und früher im Dokument stehende `unbekannte_property` | `odata_typ` (`__metadata` zuerst) | ein Fall |
| `__metadata` kein Objekt und späterer Fehler einer Property | `struktur` `…/__metadata` | ein Fall |
| Fehler einer Property (`typ`) und fehlender Schlüssel | der Fehler der Property | ein Fall |
| zwei fehlende Schlüssel | der erste in der Reihenfolge des Modells | ein Fall |
| Marker am ersten Schlüssel und fehlender zweiter | der Marker (Ort seine Property) | ein Fall |
| Navigation `typ` und `unbekannte_property`, beide Reihenfolgen | das erste im Dokument | zwei Fälle |

**Z4 — Modi:** Zusagen, die von der Methode abhängen, stehen in H8, H9 (Entity gegen Collection) und mit beiden Methoden in G1, G2, G4, G5, H1, R2; ein Modus ohne Zeile gibt es nicht. Ein Modus der Version (`ParseEntity` gegen `ParseCollection` am selben Body) ist je eine Zeile der Schleife.

**Akzeptierte Negative** (Grund und Folge stehen am Ort):
- Versatz an `Edm.DateTime` ist `wert` (E6 a): Grund: `Edm.DateTime` hat keine Zone (`SPEC-170` Nr. 4), die Spezifikation von V2 kennt den Versatz für `Edm.DateTimeOffset`, kein beobachteter Dienst schreibt ihn; Folge: ein solcher Dienst ist nicht importierbar (`ODR-E6101`).
- `__next` unbelegt (E6 b): Grund: der Referenzdienst paginiert nicht; Folge: Form nach der Spezifikation, Fänger H7 mit künstlichen Bodys.
- Schlüsselliteral gegen URL-Leser (E6 c): Grund: der URL-Leser ist nicht dieser Slice; Folge: Abweichung zeigt sich im Rundlauf von `slice-v2-bootstrap` (Sendung S1).
- Prüfgrenzen der Mengen: G9 und G10 prüfen alle Folgen bis zu einer Länge, die die Zählprüfung nennt; längere Folgen und Alphabete außerhalb der Token sind nicht erschöpfend; Folge: ein Fehler, der nur in einer längeren Folge liegt, bleibt ungesehen, der Probe-Lauf des Reviewers über die Klasse ist das Mittel (`AGENTS.md` §3.17).
- Prüfgrenze des Kreises (N1, `v2N1Kreis`): geprüft sind Selbstbezug, Ringe über 2 und 3 Typen und ein Anhang an einem Ring; Ringe beliebiger Länge sind eine unbeschränkte Menge, und `SPEC-172` Nr. 3 sagt zu, dass die Länge ohne Wirkung ist. Grund: der Leser markiert jeden besuchten Typ, die Länge des Rings geht in keine Entscheidung ein; Folge: eine Fassung, die nur Ringe ab Länge 4 nicht abbricht, bliebe ungesehen; ebenso eine Prüfung des Kreises vor der Suche nach dem Typ des Sets, denn das Modell `Y` führt keinen Ring mit dem Typ des Sets (V-1 der Verifikation, Zusage `SPEC-172` Nr. 3 ohne Fänger, siehe §3 *Akzeptierte Negative nach der Verifikation*; kein neuer Fänger in diesem Slice). Die Frist von 10 s je Lauf hält einen Leser ohne Abbruch an; ein hängender Lauf bleibt als Goroutine bis zum Ende des Prozesses stehen (nur in der Mutante).
- `Edm.Int16`, komplexe Typen und `Edm.Double` echter Dienste werden übersprungen (`SPEC-170` Nr. 3, entschieden): Folge im Betrieb: der Mock zeigt diese Properties nicht; im Referenzszenario sind 3 von 9 Properties von `Employee` (`Age`, `Location`, und `Room.Seats`/`Version`) betroffen; ob ein Register-Eintrag folgt, ist im Punkt *Beobachtungs-Register* weiter unten in §6 entschieden (kein Eintrag).


Die Zelle *unbekannte Property, nicht unterstützter Typ* führt auch die Property mit `Unterstuetzt()` falsch und `EdmType` `Edm.DateTimeOffset(V2)`: `Uebersprungen` mit `typ`, Ort `SPEC-172` Nr. 3, Fänger Nr. 1 der Tabelle in §3 (Sendung von `slice-v2-metadaten`). Die Zelle *Typ-Schreibweisen* führt den Wert des Modells `Edm.DateTimeOffset` aus `Edm.DateTime`: `datetimeoffset`, Ort `SPEC-170` Nr. 4, Fänger Nr. 2.

- Die Sendungen von `slice-v2-metadaten` (Fixture `Edm.DateTimeOffset(V2)`, Test `versionstrennung_test.go`) — **Ausgang:** *entfallen*, geliefert: der Test `src/adapters/driven/odata/versionstrennung_test.go` liegt per Merge von `main` (`9650388`) im Zweig, dieser Slice hat ihn nicht angelegt und nicht geändert; die Fixture `Edm.DateTimeOffset(V2)` steht in `v2Uebersprungen` (`odata_codec_v2_vertrag.go`, Zeile N4 der Tabelle in §3, Nr. 1 bis 3).

- Sendung S1 an `slice-v2-bootstrap` (Geber dieser Slice): die DoD-Zeile des Rundlaufs „Test über beide Seiten“ führt die Zeile „Das Schlüsselliteral des V2-Payload-Lesers (`SPEC-172` Nr. 5) gleicht dem Literal, das der V2-URL-Leser (`SPEC-173` Nr. 3) für denselben Schlüssel bildet, Schleife über die sieben Typen, Geber `slice-v2-payload-lesen`.“ — **Ausgang:** *eingetreten*, getragen vom Nehmer: die DoD-Zeile *Rundlauftest* von `slice-v2-bootstrap` (`open/`) führt das Schlüsselliteral des V2-Payload-Lesers samt „Geber `slice-v2-payload-lesen`“ seit Commit `ed0c5e0` auf `main`, im Zweig per Merge; sein §1 *Ausdrücklich NICHT* schließt die Sendung nicht aus (er verdrahtet die Leser und ändert keinen). Es war nichts nachzutragen (`AGENTS.md` §3.13 gelesen und geprüft: `slice-v2-bootstrap` Zeile 72 der DoD). Der Punkt steht bis zum Rundlauf als akzeptiertes Negativ (E6 c).
- Keine Sendung an `slice-v2-payload-schreiben`: E5 gilt dem Leser, Nr. 6 f und Nr. 7 von `SPEC-172` bleiben unverändert.
- Beobachtungs-Register: im Referenzszenario sind `Age` und `Location` von `Employee` sowie `Seats` und `Version` von `Room` von der Übersprungs-Regel für `Edm.Int16` und komplexe Typen betroffen (`SPEC-170` Nr. 3, entschieden); ein Eintrag in `observations/` ist eine Antwort, keine Aktion dieses Slice — **Ausgang:** *entschieden* (Planner, Closure): kein Eintrag im Register. Grund: das Register führt Lernstoff des Harness (Regeln, Sensoren, Spec-Lücken); die Übersprungs-Regel ist eine Entscheidung der Spezifikation (`SPEC-170` Nr. 3) mit benannter Folge im Betrieb im akzeptierten Negativ oben (der Mock zeigt diese Properties nicht), kein Fehler des Vorgehens. Wer die Typen später unterstützen will, ändert `SPEC-170`, nicht das Register.

Die Zeilen zum Schreiben des Rasters stehen in `slice-v2-payload-schreiben`.

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

**Stand der Closure Phase 1** (Planner, 2026-11-05, nach dem Merge von `main` im Zweig, Merge-Commit `443c7da`; das reine `git mv` nach `done/` folgt als eigener Commit, die Integration in `main` danach). Belege: Review-Report `docs/reviews/2026-11-05-slice-v2-payload-lesen-review.md` (0 HIGH, 2 MEDIUM, 1 LOW, 5 INFO; Commit `ead28d1`), Handoff mit Behebung `617585f`, Entscheidung des Architect zum Kreis `6b9b332`, Verifikations-Report `docs/reviews/2026-11-05-slice-v2-payload-lesen-verifikation.md` auf `6b9b332` (0 HIGH, 0 MEDIUM, 2 LOW, 3 INFO; Commit `cbdddf1`). Der Verifier fuhr `make gates` (Exit 0), die Mutationen und die Raten selbst. Review und Verifikation sind je einmal gelaufen; eine weitere Fix-Runde gibt es nicht (Entscheidung des Users: Durchstich vor Härtung).

**Gegenstand:** geliefert. `jsoncodecv2.New()` (Paket `src/adapters/driven/odata/jsoncodecv2`, `SPEC-172` Nr. 1 bis 5) liest die JSON-Antworten eines V2-Dienstes (Hülle `d`, Collection mit `results`, `__count`, `__next`, Entity) gegen ein `model.ServiceModel` der Version `2.0` in typisierte Entities: Typen in der Schreibweise von V2 (`/Date(…)/`, `Edm.Int64` und `Edm.Decimal` als String), nicht gedeutete Konstrukte als übersprungen gemeldet, das leere Ergebnis eine leere Menge; ein Modell anderer Version ist `ErrUngueltig`. Kein V2-Paket importiert ein V4-Paket (`versionstrennung_test.go`, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)). **Das Binary blieb unverändert:** `jsoncodecv2` ist nirgends verdrahtet, ein V2-Dienst erreicht es erst mit `slice-v2-bootstrap`, das auch das Kapitel zu V2 im Handbuch trägt; kein Fehlercode, Schlüssel oder Kommando kam in den Katalog.

**Größen:** Messung 4085 von Marke 8100 (Auslastung 0,50; `make groessenmarke`: ok, vom Verifier auf `6b9b332` gemessen), Schätzung *S* = 5970 nach `SPEC-128`, Ende ÷ *S* = 0,68 (die Erwartung des Plans lag bei 0,8 bis 1,2); Produktcode 1389 Zeilen in zehn Dateien statt 1400. Die 7323 des Reviews zählten ganze Verzeichnisse, darunter 3245 Zeilen aus `slice-v2-metadaten`, und gelten nicht (Ursache in §4). Die Marke ist nicht verändert; die Closure-Commits ändern in den Pfaden des Blocks nichts. Der Datenpunkt geht an die Neukalibrierung der Faktoren (Rückgabe 2), Register `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`.

**Mutationen:** Tabelle *Zusage · Fänger* (57 Zeilen) mit der Mutation je Zusage und den Raten von G20 und G21 in §3. Der Verifier bestätigte die Raten mit 30 Prozessen je Zeile: C1 bis C3 rot in 30 von 30 (je Zeile), C4 (Sperre, kein Datenrennen) 0 Meldungen des Detektors, rot im Wirkungs-Fänger. Er fuhr 21 Mutanten für G9 jenseits des Tabulators, 4 für G14, 5 für H8, H9, K1, R2 und 28 an den geteilten Helfern (alle sechs Aufrufstellen) selbst: alle rot; grün blieb allein der Kreis (V-1).

**Was nicht gedeckt ist (nicht behoben):** V-1 der Verifikation ist ein akzeptiertes Negativ mit Grund und Folge im Betrieb (§3, *Akzeptierte Negative nach der Verifikation*), kein behobener Befund. Der **Code** ist laut Probe-Paket des Verifiers richtig (3456 Läufe: Ringe der Länge 1 bis 12 × Typ des Sets außerhalb oder im Ring × Anhänge 0 bis 4 × beide Methoden, 0 Abweichungen); es fehlt der Fänger. Eine Mutation bliebe in `make gates` grün, wenn sie nur Ringe ab Länge 4 nicht abbräche oder den Kreis vor der Suche nach dem Typ des Sets prüfte.

**Risiko-Ausgänge (§6):** je Punkt im Plan selbst. Eingetreten: die Sendung S1 (getragen von `slice-v2-bootstrap`, siehe §6), die Kopplung an die V4-Suite (gelöst durch drei geteilte Helfer), der Konflikt in der Historie von `spec/spezifikation.md` (beim Merge der Closure gelöst), `slice-v2-import-codecs` (Auswahl der Antworten). Entschieden: Int16 und komplexe Typen (kein Register-Eintrag), Kopplung (kein Register-Eintrag), Abweichung J4 und Kreis (Architect). Weiter offen, mit Adresse: der Nachzug des Fängers für V-1 (Planner, Zuweisung bei der Closure der Welle; ein benannter Folge-Slice existiert nicht, kein Plan auf `main` nimmt ihn an, deshalb keine Sendung nach `AGENTS.md` §3.13).

**Rückgaben (nicht entschieden; der Planner entscheidet sie nicht):**

1. User: Verkörperung ab 3×. `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` hat mit diesem Slice 19 Belege und ist in `AGENTS.md` §3.17 verkörpert; dies ist der vierte Slice in Folge (`slice-text-antworten`, `slice-v2-query`, `slice-v2-query-filter`, `slice-v2-payload-lesen`), in dem ein Fix die Klasse nicht ganz deckte. Anders als in den drei Slices davor liegt der Fall hier eine Stufe niedriger: der Fix des Reviews (Kreisschutz mit Fänger) führte Selbstbezug, zwei Ringlängen und einen Anhang, und die zwei Sätze der Zusage (Ring ab Länge 4, Ring mit dem Typ des Sets) blieben ungefangen; Befund war ein LOW des Verifiers, die beiden MEDIUM des Reviews hielten gegen die Klasse. Dazu `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (5×, Schwelle erreicht, Ausgang nicht zugewiesen) und `BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo` (6×, verkörpert in `AGENTS.md` §3.10; dies ist der Fall *Handoff ohne Plan-Absatz*, ein Sensor dafür ist nicht beschlossen). Adresse: User, spätestens beim Lese-Schritt der Closure von `welle-odata-v2`.
2. User und Architect: `AGENTS.md` §3.14 (40/65/1,25) gegen `SPEC-128` Nr. 1 (80/85/1,35) und die Neukalibrierung der Faktoren; Datenpunkt dieses Slice: *S* 5970, Messung 4085, 0,68. Der Planner ändert die Faktoren nicht.

**Steering-Loop-Einträge:**

- Lerneintrag, benannte Spec-Lücke, geschlossen: `SPEC-172` Nr. 3 nannte nicht, was der Leser bei einer geschlossenen Kette von `BaseType` tut (ein von Hand gebautes `model.ServiceModel` kann sie tragen, `SPEC-125` Nr. 9 a weist sie nur im CSDL-Leser ab); der Architect entschied nach dem Review: der Leser kehrt bei jedem Modell zurück, ein Typ in einer geschlossenen Kette ohne den Typ des Sets ist `odata_typ`, erreicht die Kette den Typ des Sets, ist er abgeleitet, die Länge des Rings ist ohne Wirkung (Historie vom 2026-11-05 in `spec/spezifikation.md`). Die zweite Entscheidung des Architect ist J4: die V4-Suite gibt drei Helfer an die V2-Suite ab (`outboundtest` ist ein Paket ohne Versionsnamen, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) §Konsequenzen), weil `make lint` ohne Ausnahme grün bleibt (`AGENTS.md` §3.6). Gegenstück: `LH-FA-17` (versionierter Adapter, keine Vermischung von V2 und V4). Kein `liegt in`: verkörpert in der Spezifikation und im Plan, nicht an einem Zielort.
- Lerneintrag, gezählt: der Handoff kam als Zwischenstand ohne Absatz mit Mutationen und Raten im Plan und mit rotem `make lint` (elf Meldungen aus `cyclop`, `gocyclo`, `gocognit`, `dupl`, `unparam`); der Review fand beides (MEDIUM 1 und 2), der Implementer behob beides durch Umbau ohne `//nolint` und ohne Änderung von `.golangci.yml`. Evidenz zu `BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo`; das rote `make lint` am Handoff hat keinen eigenen Eintrag (kein vorhandener passt, ein Auftreten).
- Lerneintrag, gezählt: der Fix zum Kreis (LOW 3 des Reviews) deckte die genannte Mutation (Markierung `besucht`), nicht die Klasse der Zusage (V-1). Evidenz zu `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`; kein neuer Sensor und keine neue Regel mit diesem Slice, der Vorschlag steht in Rückgabe 1.
- Lerneintrag, gezählt: die Formel von `SPEC-128` überschätzt den Slice (0,68). Evidenz zu `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (5×).
- Beobachtung ohne Eintrag: die Zahl des Reviews zur Größenmarke (7323) zählte ganze Verzeichnisse statt der Pfade des Blocks; ein Auftreten, geklärt in §4, gezählt nicht.

**Register (Phase 1):** Evidenz `slice-v2-payload-lesen` zu `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`, `BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo` und `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`; kein neues Verzeichnis. `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` und `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (aus §8) bekamen keine Evidenz: Review und Verifikation melden keine Verletzung der Betriebsregel, und die Strom-Hülle der Suite führt alle acht Methoden der Menge O (G8).

**Drei Paarungen:** Anker (kein Feld `liegt in` in diesem Slice, nichts verkörpert; die Sensor-Paarung hat keinen Eintrag zu prüfen) · Folge-Slice (der Nachzug des Fängers ist keinem Slice zugewiesen, deshalb als *weiter offen* mit Adresse Planner und `welle-odata-v2` geführt, nicht als Folge-Slice behauptet; die Sendung S1 an `slice-v2-bootstrap` existiert als Datei in `open/` und ist dort in der DoD getragen) · Register (jede genannte Kennung `BEO-REPO/<slug>` existiert, jedes `evidence/` ist nicht leer) — die nächste Welle-Closure prüft die Paarungen erneut.

**Offener Rest:** das reine `git mv` nach `done/` und der Merge nach `main` sind nicht Teil dieser Phase; kein Push.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30), beim Start nachgezogen (2026-11-01, Zähler = Zahl der Dateien in `evidence/`). `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, Schwelle erreicht, Verkörperung offen) gilt für Fakes der Codec-Ports und Hüllen von Strömen (jede Schnittstelle des Originals, jeder Read gezählt); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (3×, Schwelle erreicht, Zustand offen) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig. Beide sind Feldlagen, die der Plan als Risiko führt (§3 Betriebsregeln, §6); ein weiteres Auftreten trägt der Slice in `evidence/` ein.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
