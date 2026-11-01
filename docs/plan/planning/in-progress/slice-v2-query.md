# Slice slice-v2-query: Pfad, Systemoptionen und expliziter Kompatibilitätsfehler der V2-URL (Teil A)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-QA-05`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-170` · `SPEC-173` Nr. 1 bis 4, 6 und 7 (Nr. 5, die Grammatik von `$filter`, gehört `slice-v2-query-filter`) · `SPEC-174` · `SPEC-175` Nr. 3 (Feld `URL` des Satzes) · `SPEC-177` · `SPEC-160` bis `SPEC-162` (Gerüst, das `SPEC-173` Nr. 1 übernimmt) · `SPEC-109` Nr. 2 und 5 · `SPEC-230` Nr. 8 · `SPEC-128` · `LH-FA-17.a` (versionierter Adapter für Query-Semantik, keine Vermischung von V2 und V4) · `LH-FA-17` Negative (nicht unterstützte V2-Eigenschaft: expliziter Kompatibilitätsfehler) · `LH-FA-08.a` und `LH-FA-08.b` (V4-Semantik der Abfrageoptionen, auf die der AST abbildet) · `SPEC-022` (`ODR-E6001`) · `SPEC-049` · `SPEC-002` · `ARC-001`, `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkt OData V2).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-v2-query)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel (Teil A der Teilung, Schnitt entschieden, §4):** Der Parser der V2-URL hinter dem Port der ADR überführt Pfad und `raw_query` einer V2-Abfrage im dokumentierten V2-Umfang (Collection, Key, `$count`, `$orderby`, `$top`, `$skip`, `$select`, `$inlinecount`, `$format` und die V2-Schreibweisen der Literale des Schlüssels) in den Query-AST des Domain-Modells, so dass der Query-Interpreter aus `welle-mock` ihn unverändert ausführt; jede V2-Eigenschaft außerhalb des Umfangs, die der Parser kennt (Option, Pfadform, Format, Literal; etwa `$expand`, `$links`, `$format=atom`, `$batch`), ergibt einen expliziten Kompatibilitätsfehler mit eigenem Code und nie eine scheinbar gültige Antwort (`LH-FA-17`, Negative; `LH-RB-01`). Der Slice liefert die Einhängestelle `WithFilter` samt Typ `FilterLeser` (`SPEC-173`, `SPEC-160` Nr. 8); ohne Leser ist `$filter` mit lesbarem Wert der Kompatibilitätsfehler `filter`. Der Leser der Grammatik von `$filter` (`SPEC-173` Nr. 5) ist `slice-v2-query-filter`. Der Slice liest Text und kennt das Modell nicht (`URLCodec.ParseURL(pfad, rawQuery)`, `SPEC-160` Nr. 9, `SPEC-173`); die Prüfung gegen das Modell ist Sache des Dienstes, der die Query ausführt. Er schreibt keine Antwort. Der Kompatibilitätsfehler ist `ODR-E6006` (HTTP 501, `SPEC-174`, `SPEC-177`) mit der Abbildung in `outbound`; das Feld `URL` im `outbound.CodecSatz` (`SPEC-175` Nr. 3) liefert dieser Slice, wie `slice-v2-import-codecs` es ihm übergeben hat. Der Parser ist eine Kopie des V4-Parsers mit Abweichung, kein Import ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md); Begründung in §3).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Metadaten-Leser und Payload-Codec — `slice-v2-metadaten`, `slice-v2-payload-lesen` und `slice-v2-payload-schreiben`; dieser Slice braucht das V2-Modell und liest und schreibt keine Payloads.
- Wahl des Adapters je Service, Konfiguration, Route und Rundlauftest — `slice-v2-bootstrap`.
- Ausführung der Abfrage — der Query-Interpreter aus `slice-query-interpreter` (`welle-mock`); ein V2-Konstrukt, das der AST nicht tragen kann (V4-Semantik), ist hier ein Kompatibilitätsfehler und ändert den Interpreter nicht (V4-Pfad bleibt unberührt, `welle-odata-v2` §6).
- Grammatik von `$filter` (Operatoren, Funktionen, Literale im Ausdruck, `NewFilterLeser`, die Zuordnung `filter` und `literal_typ` im Ausdruck) — `slice-v2-query-filter` (Teil B; die Literal-Funktion und die Einhängestelle liegen hier und gehen an B, B startet nach diesem Slice in `done/`).
- V2-Mutationen (POST, MERGE, PUT, DELETE) und die Ursache `methode` — `welle-odata-v2` §6 und `slice-v2-bootstrap`: `ParseURL` hat keine Methode, `methode` erzeugt der Mock (`SPEC-174` Nr. 3); was die Spezifikation nicht nennt, ergibt dort den Kompatibilitätsfehler.
- Verdrahtung des Parsers im Satz der Implementierungen und der Fänger „Versions-Kopfzeilen ändern nichts“ (`SPEC-174` Nr. 6) am Dienst — `slice-v2-bootstrap`; der Parser sieht keine Header.
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

- [ ] Parser: Pfad, Schlüsselliterale und Systemoptionen des V2-Umfangs (`SPEC-173` Nr. 1 bis 4, 6, 7) werden zum Query-AST des Domain-Modells; ungültig `ODR-E6005` (400); die Einhängestelle `WithFilter` mit `FilterLeser` liegt, ohne Leser ist `$filter` der Kompatibilitätsfehler `filter`. Der Test über beide Seiten (Parser plus Interpreter) liegt im Rundlauf von `slice-v2-bootstrap`.
- [ ] Kompatibilitätsfehler: `*model.KompatFehler` mit `ODR-E6006` (36 Codes in `model.AlleCodes()`, Abbildung in `outbound`, Feld `URL` im `outbound.CodecSatz`, `SPEC-174`); jede V2-Eigenschaft außerhalb des Umfangs, die der Parser kennt (Option, Pfadform, Format, Literal), ergibt ihn, nie eine scheinbar gültige Antwort und nie `404` oder die V4-Meldung; der Fehler nennt die Ursache und den Ort.
- [ ] Randformen des Lesers fremder Bytes (leere und doppelte Option, Prozent-Kodierung, V4-Syntax am V2-Parser, Literale in V2-Schreibweise, Länge, Wiederholung): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Generator über die Klasse der URLs, Allaussage, echte V2-URLs.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: entfällt, getragen von `slice-v2-bootstrap` (Eintrag von `ODR-E6006` und Kapitel zu V2; dort steht beides in §1 und der DoD). Grund: das Binary ändert sich in diesem Slice nicht, kein Pfad erreicht den Code ohne den Composition Root, und das Gate des Katalogs (`slice-handbuch-meldungskatalog`, `open/`) existiert nicht.
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
| `src/adapters/driven/odata/urlparserv2/` | neu | Parser (`SPEC-173` Nr. 1 bis 4, 6, 7), `urlparserv2.New(opts ...Option)`, `WithFilter`, Typ `FilterLeser`; Kopie des V4-Parsers mit Abweichung, kein Import |
| `src/adapters/driven/odata/urlparserv2/trennung_test.go` | neu | `TestKeinV4Import` (Zeile 16), unabhängig von `versionstrennung_test.go` des Zweigs `slice-v2-metadaten` |
| `src/hexagon/model/kompat_fehler.go` und `kompat_fehler_test.go` | neu | `KompatFehler`, `KompatUrsache`, `AlleKompatUrsachen`, `ErrKompatibilitaet` (`SPEC-174` Nr. 1, 2) |
| `src/hexagon/model/fehler.go` und `fehler_test.go` | update | Code `ODR-E6006` (Status 501, Ziel `mock`, Text `Unsupported OData V2 feature`) hinter `CodeE6005`, `model.AlleCodes()` hat 36 Einträge; der Query-AST bleibt unverändert |
| `src/hexagon/ports/outbound/abbildung.go` und `abbildung_test.go` | update | Abbildung `ErrKompatibilitaet` → `ODR-E6006` (`SPEC-109` Nr. 5) |
| `src/hexagon/ports/outbound/codec_satz.go` | update | Feld `URL URLCodec` im `CodecSatz` (`SPEC-175` Nr. 3; Sendung von `slice-v2-import-codecs`) |
| `src/adapters/driven/telemetry/regeln_test.go` und `telemetry_test.go` | update | 36 Reihen von `odr_errors_total` (`SPEC-230` Nr. 8); jede Stelle mit `35` (Kommentare und Zahlen) |
| `src/adapters/driving/httpodata/klassen_test.go` | update | der Kommentar „34 Paare“ wird 35 |
| `src/adapters/driving/httpodata/fehler_test.go`, `src/hexagon/services/session/beenden_test.go`, `src/hexagon/services/workspace/helfer_test.go`, `src/hexagon/services/workspace/kontext_test.go` und `src/hexagon/services/workspace/reset_test.go` | update | im Plan nicht genannt, vom Lauf gefunden: die Zahl der Paare (35), die Liste der Fehler der Abbildung (`ErrKompatibilitaet`) und die Zahlen 19 × 25 und 4 × (25 + 4 + 1) hängen an der Tabelle der Codes und der Abbildung (`AGENTS.md` §3.9) |
| Tests | neu | Fälle von `LH-FA-17` je Option und je Fehlerklasse; Generator über die Klasse der URLs; Orakel in `orakel_test.go` ohne Import aus dem Produktcode |

**Verdikte des Architect zum Start** (der Plan hatte falsch behauptet oder nicht genannt):

- Die Methode ist keine Sache des Parsers: `ParseURL` hat keine Methode, `methode` erzeugt der Mock (`SPEC-174` Nr. 3, `slice-v2-bootstrap`). §1 und DoD sind berichtigt.
- `outbound/codec_satz.go`, die Tests der Telemetrie und `klassen_test.go` fehlten im Plan; sie wachsen wegen des Feldes `URL` und wegen 36 statt 35 Codes (`AGENTS.md` §3.9).
- `SPEC-173` widersprach sich an neun Stellen (Ursachen `format` und `option`, `$format` mit Schreibweisen und Parametern, `$id`, `$links`, `$orderby` mit Funktion, `$inlinecount`, Fehler des Lesers, Filter, Millisekunde); die Berichtigung steht in `spec/spezifikation.md` (L1 bis L9 des Architect) und ist vor dem ersten Code-Commit eingetragen. Kein neuer ADR: es sind Lücken der Spezifikation, keine Regeln der ADR (`AGENTS.md` §3.8).

**Darf der V2-Parser den V4-Parser nutzen? Nein, er ist eine Kopie mit Abweichung.** [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) schreibt es vor (kein V2-Paket importiert ein V4-Paket; `TestKeinV4Import` prüft es), und `urlparser` exportiert nur `New`, `WithFilter`, `Parser`, `FilterLeser`, `NewFilterLeser`. Ein versionsfreies Paket für das Gemeinsame (rund 175 Zeilen, 3,5 % von *S*: Dekodierung und Zeitfunktionen) lohnt nicht: es verschöbe V4-Produktcode und verlangte neutrale Einträge in `versionstrennung_test.go` eines anderen Zweigs. `model` bleibt der gemeinsame Ort: `model.QuerySegmente` zerlegt die Query für V2 unverändert. Test-Hilfen werden nicht geteilt; `urlparserv2` führt seine eigenen, auch den Zähl-Fake des `FilterLeser`. **Akzeptiertes Negativ:** die Kopien können auseinanderlaufen (kanonischer Text von Datum, Zahl, Guid). Grund: dieselbe Lage wie `SPEC-160` *Ungeprüft*. Folge im Betrieb: ein Schlüssel der URL findet die Entity der Baseline nicht (`404` statt `200`); den Fänger der V2-Seite tragen die Zeilen 28 bis 30 (kanonischer Text gegen die Beispiele von `SPEC-172` Nr. 4 und 5), den Fall je Typ der Rundlauf in `slice-v2-bootstrap`.

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Parser liest fremde Bytes (den Text der URL): die Produktcode-Erwartung des Architect (`produktcode: 1000`, Spanne 900 bis 1150) ist ein fester Anteil der Schätzung (`SPEC-128`); URLs mit Escape-Folgen entstehen aus Bytewerten (`AGENTS.md` §3.18). Die Eingaben der Zeilen 1, 24, 40, 41 und 44 führen U+0085, U+00A0, U+2003, U+3000, U+FEFF und `%FF`.
- Die Mengen der Zeilen 47 und 48 stehen in der Spezifikation (`SPEC-170` Nr. 3, `SPEC-173` Nr. 4), nicht im Code: der Test trägt sie als eigene Listen und zählt sie. „Ort“ in der Tabelle ist der Ort des Fehlers (`pfad`, `pfad/<n>`, `query/<n>`).

**Auftrag an den Architect:** erledigt ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), `Accepted`). Auftrag 1 steht in `SPEC-170` Nr. 2 und 3, 2 in `SPEC-173` Nr. 3 bis 5, 3 in `SPEC-174` Nr. 1 und 4 (Code `ODR-E6006`, Status 501), 4 in `SPEC-173` Nr. 4 und 5, 5 in `SPEC-174` Nr. 6. Der Text der ursprünglichen Fragen bleibt zur Nachprüfung stehen:

1. Der V2-Umfang der Abfrage (Teil des dokumentierten V2-Umfangs aus `slice-v2-metadaten`, Auftrag 1): Optionen, Operatoren, Funktionen (`substringof`, `startswith`, …), Literale (`datetime'…'`, `guid'…'`, Suffixe `L` und `M`), `$inlinecount`, `$count` als Pfad, Key-Syntax.
2. Abbildung auf den Query-AST: welche V2-Konstrukte auf bestehende Knoten fallen und welche der AST nicht trägt (sie sind Kompatibilitätsfehler); wo die Typregeln der V2-Typen liegen (`slice-edm-ausdruecke` ist V4-typisiert, die Abbildung der V2-Typen darauf steht in der Spezifikation, nicht im Code).
3. Kompatibilitätsfehler: Code, Status (vorgeschlagen ist die Linie von `ODR-E6001`, `501`; die Entscheidung trifft der Architect), Meldungstext mit der Eigenschaft, Rangfolge gegen `400` bei mehreren Verstößen.
4. Grenzen (Länge, Tiefe der Ausdrücke) mit Grund und Folge im Betrieb; Umgang mit V4-Syntax am V2-Parser.
5. Version in der Anfrage: `DataServiceVersion`, `MaxDataServiceVersion` und `OData-Version` und was bei einem Widerspruch zur Version des Service gilt.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; 49 Zeilen vom Architect, 2 nebenläufig und 47 andere; „Menge n“ nennt die Elementzahl aus dem Wortlaut, ein Fänger ist eine Schleife oder ein Generator über die Menge, nie eine Auswahl, `AGENTS.md` §3.17; alle Fänger laufen über `ParseURL`, außer wo anderes steht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Form der Argumente (`SPEC-160` Nr. 1): `pfad` leer, ohne `/`, mit rohem `?` oder `#`, mit rohem Steuerzeichen; `rawQuery` mit rohem Steuerzeichen: `model.ErrUngueltig`, weder `*URLFehler` noch `*KompatFehler`, Nullwert. Menge 75 = pfad 37 (leer, ohne `/`, `?`, `#`, 33 Steuerzeichen U+0000 bis U+001F und U+007F) + rawQuery 33 + 5 Gegenfälle (rohes `#` und `?` in rawQuery, `%3F` und `%23` im Pfad, rohes Zeichen jenseits von ASCII im Pfad) | `TestArgumente`: Schleife, `errors.As` auf beide Fehlertypen | U+007F nicht geprüft; `#` im Pfad erlaubt; Grenze `< 0x1F` | andere |
| 2 | Wiederholung und Eigentum (`SPEC-160` Nr. 3): gleiche Eingabe, gleiches Ergebnis (tiefe Gleichheit) und gleicher Fehler; das Ergebnis gehört dem Aufrufer. Menge 68 = 30 Eingaben (20 Erfolge über alle Ressourcenarten und Optionsfelder, 5 `ungueltig`, 5 Kompatibilität) × 2 Aufrufe + 4 Felder (`Schluessel`, `Select.Namen`, `OrderBy`, `Sonstige`) × 2 Elemente (erstes, letztes) | `TestWiederholung`: Schleife; zweiter Aufruf nach Änderung des ersten Ergebnisses | Ergebnis zwischengespeichert; Slice zwischen zwei Ergebnissen geteilt | andere |
| 3 | Form des Ergebnisses (`SPEC-160` Nr. 2): Art je Ressource, `EntitySet` leer bei `$metadata`, `Schluessel` nil ohne Klammer, unbenannter Teil mit leerem Namen, `$top=0` ist gesetzt mit Wert 0, nicht genannte Option ist der Nullwert, `OrderBy` und `Sonstige` nil. Menge 20 Zeilen (4 Arten, 6 Nullwert-Felder, 3 `$top`-Fälle, 4 Schlüsselformen, 3 Reihenfolgen der Optionen ohne Wirkung) | `TestErgebnisForm`: Tabelle mit erwartetem `ODataURL` | `$top=0` als nicht gesetzt; `OrderBy` leere statt nil; Name des unbenannten Teils gefüllt | andere |
| 4 | Grenzen (`SPEC-160` Nr. 6, beide Seiten): pfad 2048, rawQuery 8192, Segmente 16, Segmente der Query 64, Teile 16, Literal 512, Name 128, `$select` 256, `$orderby` 32; Name 128 zusätzlich an drei weiteren Stellen (Set-Name `segment`, Schlüsselname `key`, Optionsname `option_name`). Menge 26 = 8 Grenzen × 2 Seiten (pfad, rawQuery, Segmente, Segmente der Query, Teile, Literal, Elemente von `$select`, Elemente von `$orderby`) + Name 128 an 5 Stellen (Set-Name, Schlüsselname, Optionsname, Eintrag in `$select`, Eintrag in `$orderby`) × 2 Seiten; Ursache und Ort je Fall | `TestGrenzen`: Generator der Eingabe genau auf und über der Grenze | Grenze 2047 oder `>=`; Teile 15; Name an einer Stelle ungeprüft | andere |
| 5 | Gesamtform vor allem anderen (`SPEC-160` Nr. 5, Stufe 0): Reihenfolge `zu_lang` (pfad), `zu_lang` (query), `zu_viele` (pfad), `zu_viele` (query); hinter einem Verstoß der Stufe 0 wird nichts gelesen. Menge 14 = 4 einzeln + 6 Paare + 4 Stufe-0-Verstoß mit zusätzlichem `ungueltig` oder Kompatibilität dahinter | `TestStufeNull`: Paar-Generator | zweiter Verstoß gemeldet; `zu_lang` (query) vor `zu_lang` (pfad) | andere |
| 6 | Wächter (`SPEC-160` Nr. 6 a bis c): Summe der Längen der Texte im Ergebnis höchstens Länge(pfad) + Länge(rawQuery) + 48; `TotalAlloc` höchstens 64 · (Länge(pfad) + Länge(rawQuery)) + 65536. Menge 29 = 6 Zeichenfamilien (`'`, `%`, `,`, `(`, `&`, `/`) × 4 Formen (pfad an der Grenze, rawQuery an der Grenze, beide, ein Pfad aus lauter Familie) + Schlüssel mit 16 benannten Teilen von einem Zeichen + Zeitpunkt ohne Sekunden + 3 große Erfolge; der Zeitpunkt wächst in V2 nicht (die Eingabe `datetime'…'` ist länger als sein Text) | `TestWaechter`: Schleife; Texte des Ergebnisses summiert, `runtime.MemStats` vor und nach | Verkettung in einer Schleife (quadratisch); Wert einer Sonstigen Option verdoppelt | andere |
| 7 | `ODR-E6005` geht jeder Kompatibilität vor, gleich wo sie steht (`SPEC-173` Nr. 1, 6). Menge 38 = alle ausführbaren Stellungen von 6 Kompatibilitätsformen (Pfad: Service-Dokument `/`, Rest-Segment, Literal `1d`; Query: `$expand`, `$format=xml`, `$filter` ohne Leser) mit 5 Verstößen (Pfad: `segment`, `key`; Query: `option_name`, `wert`, `kodierung`): Pfad-Form × Pfad-Verstoß 5 (Rest-Segment und Literal × `segment` und `key` × früher und später, soweit ausführbar: ein `key` hinter einem Rest-Segment oder Literal gibt es nicht, ein `segment` vor dem ersten Segment ebenso; das Service-Dokument verträgt keinen), Pfad-Form × Query-Verstoß 9, Query-Form × Pfad-Verstoß 6, Query-Form × Query-Verstoß 18 (früher und später) | `TestUngueltigVorKompat`: Paar-Generator; Orakel eine unabhängige Funktion des Tests | Kompatibilität vor `ungueltig` gemeldet; Lesen bei der Kompatibilität beendet | andere |
| 8 | Zwei Kompatibilitätsfehler: der erste in der Lesefolge gilt (`SPEC-173` Nr. 1, 6). Menge 21 = geordnete Paare der 6 Formen von Zeile 7, soweit ausführbar: Service-Dokument vor einer Query-Form 3, Literal vor Rest-Segment 1, Pfad-Form (Rest-Segment, Literal) vor einer Query-Form 6, Query-Form vor Query-Form 9 (gleiche Form zweimal eingeschlossen; `$format` und `$filter` zweimal sind behandelte Namen und ergeben `option_doppelt`, `$expand` zweimal die Kompatibilität), zwei Rest-Segmente 1, zwei Literale im Schlüssel 1 | `TestErsteKompat`: Paar-Generator, Ursache und Ort | letzte statt erster gemeldet; Pfad nach Query geprüft | andere |
| 9 | Zwei Verstöße der Klasse `ungueltig`: der erste in der Lesefolge gilt (`SPEC-160` Nr. 5). Menge 18 = 5 Repräsentanten (`segment` im Pfad, `key`, `option_name`, `wert`, `kodierung` im Namen einer Option) als Paare in der Reihenfolge der URL (25 abzüglich 6, in denen ein Query-Verstoß links vom Pfad-Verstoß stünde, und abzüglich 1: `key` hinter `segment` im Pfad ist nicht ausführbar) | `TestErsterUngueltig`: Paar-Generator | zweiter gemeldet; Query vor Pfad | andere |
| 10 | Fehlerwert `*URLFehler` (`SPEC-160` Nr. 4 mit `SPEC-173` Nr. 1): jede der 10 Ursachen der Klasse `ungueltig` hat Code `ODR-E6005`, `errors.Is` genau `ErrURLUngueltig`, Ort wie vorgeschrieben, Nullwert; kein Teil der Eingabe in `Error()` oder `Ort`; nie eine Ursache der Klasse `nicht_unterstuetzt`. Menge 30 = 10 Ursachen × 3 Marker (Name, Literal, Wert in der Eingabe) | `TestUrlFehlerWert`: Schleife; Marker im Text gesucht | Position im Fehler; Klasse vertauscht; Ort um 1 verschoben | andere |
| 11 | Fehlerwert `*KompatFehler` aus dem Parser (`SPEC-173` Nr. 1, `SPEC-174` Nr. 1, 2): die fünf Ursachen des URL-Lesers (`format`, `pfad`, `option`, `filter`, `literal_typ`) mit Ort wie `SPEC-160` Nr. 4; nie `*URLFehler`, `errors.Is` nicht `ErrURLNichtUnterstuetzt`; kein Teil der Eingabe im Fehler; Nullwert. Menge 30 = 5 Ursachen × 3 Positionen (n = 1, 2, 3) × 2 Marker (Name, Wert) | `TestKompatFehlerWert`: Schleife | `*URLFehler` mit Klasse `nicht_unterstuetzt` geliefert; Eingabeteil im Fehler | andere |
| 12 | Modell `KompatFehler` (`SPEC-174` Nr. 1, 2): `KompatUrsache` mit sechs Werten (`format`, `methode`, `pfad`, `option`, `filter`, `literal_typ`), `AlleKompatUrsachen` in dieser Reihenfolge, `Gueltig`; `Error()` ist `ODR-E6006: OData-V2-Kompatibilität: <Ursache>` mit oder ohne Ort in der Schreibweise von `strconv.Quote`; `Code()` ist `CodeE6006`; `errors.Is` trifft `ErrKompatibilitaet` und keinen der vier Sentinel `ErrURLUngueltig`, `ErrURLNichtUnterstuetzt`, `ErrUngueltig`, `ErrDatenUngueltig` (einen Sentinel `ErrNichtUnterstuetzt` gibt es im Modell nicht). Menge 57 = 6 Ursachen × 9 Prüfungen (ohne Ort, mit Ort, Code, Sentinel, 4 Gegen-Sentinel, Zeichenkette der Ursache) + 3 ungültige Ursachenwerte | `TestKompatFehlerModell` in `model` | `methode` fehlt in der Liste; Ort ohne `Quote`; `Is` trifft `ErrURLNichtUnterstuetzt` | andere |
| 13 | Abbildung (`SPEC-109` Nr. 5): jeder Fehler, der `ErrKompatibilitaet` trägt, wird `ODR-E6006`, auch verpackt und verbunden. Menge 384 = 6 Ursachen × 8 Formen (der Sentinel roh und verpackt, der Zeiger roh, einfach und doppelt mit `%w` verpackt, `errors.Join` links und rechts, mit `context.Canceled` verbunden: kein Code) × 8 Gegenstände, dazu 18 Fälle mit einem Fehler der Gruppe `ODR-E9001` daneben; die 16 × 8 Fälle von `ErrURL*` bleiben unverändert | `TestAbbildungKompat` in `outbound` | Zeile fehlt (ergibt `ODR-E9001`); `ErrKompatibilitaet` hinter der Gruppe `ODR-E9001` einsortiert | andere |
| 14 | Tabelle der Codes: `AlleCodes()` hat 36 Einträge, `ODR-E6006` mit Status 501, Zielbereich `mock`, festem Text `Unsupported OData V2 feature` (Entscheidung des Architect; die Texte stehen nur im Code, `SPEC-109` Nr. 3), eingefügt hinter `CodeE6005`; `odr_errors_total` hat 36 Reihen. Menge 39 = 36 Codes (Schleife) + 3 Zahlen (Codes, Reihen, `CodeE9001+1` bleibt ein Code außerhalb) | `TestCodeTabelle` (angepasst), `TestCodesTelemetrie` und die Zahlen in `telemetry_test.go` (angepasst) | Eintrag fehlt; Status 500; `CodeE6006` hinter `CodeE9001` angehängt (dann ist `CodeE9001+1` ein Code) | andere |
| 15 | `outbound.CodecSatz` trägt das Feld `URL URLCodec` (`SPEC-175` Nr. 3; Sendung von `slice-v2-import-codecs`); `*urlparserv2.Parser` erfüllt `outbound.URLCodec`; ein Satz ohne `URL` bleibt gültig (der Importer liest es nie). Menge 3 = Zuweisung des Parsers, Zuweisung eines Fakes, Nullwert des Feldes | `TestSatzTraegtURL` in `urlparserv2` | Parameter von `ParseURL` abweichend (Aufbau bricht, gewollt: die Zusage ist die Signatur); Feld als `*Parser` statt Schnittstelle | andere |
| 16 | Trennung der Versionen ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), Fitness Function): kein Import eines Pakets unter `src/adapters/driven/odata/` ohne den Namen `v2` aus `urlparserv2`, weder im Produktcode noch in den Tests. Menge: alle `.go`-Dateien des Pakets (Zählprüfung: mindestens 7 gelesen) + 4 Gegenproben im Wegwerfbaum (Import von `urlparser`, von `csdl`, von `jsoncodec`, eines eigenen Unterpakets = erlaubt) | `TestKeinV4Import` in `urlparserv2` (30 Zeilen, unabhängig von `versionstrennung_test.go` des Zweigs `slice-v2-metadaten`) | `import _ ".../urlparser"` in `parser.go` und in einer Testdatei | andere |
| 17 | Gleichzeitige Aufrufe (`SPEC-160` Nr. 7): jedes Ergebnis gleich dem eines einzelnen Aufrufs; der Lauf mit erstem Kompatibilitätsfehler ist nie Zustand der Instanz. Menge: 16 Goroutinen × 40 Eingaben (Erfolge, `ungueltig`, Kompatibilität, verschränkt) × 10 Runden, auf einer und auf zwei Instanzen | `TestGleichzeitig` (Wirkungs-Fänger: Ergebnis je Eingabe verglichen, unabhängig vom Detektor); `make test-race` für das Fehlen der Sperre | erster Kompatibilitätsfehler im Feld der Instanz; Puffer als Paketvariable. Rate je Mutation und je Lauf von `go test` bzw. je Prozess von `make test-race` mit Zahl der Prozesse im Plan (`AGENTS.md` §3.10) | nebenläufig |
| 18 | Der eingehängte `FilterLeser` wird aus mehreren Goroutinen gerufen und der Parser verändert nichts an ihm (`SPEC-160` Nr. 7, 8). Menge: 16 Goroutinen × 4 Eingaben (mit Wert, leer, doppelt, `ungueltig` davor) × 10 Runden; Summe der Aufrufe des Zähl-Fakes gleich der Zahl der Eingaben, die ihn erreichen | `TestEinhaengungGleichzeitig` (Zähl-Fake mit atomarem Zähler) | Leser pro Aufruf neu gebildet und nicht gerufen; Aufruf zweimal je `ParseURL` | nebenläufig |
| 19 | Aufrufregeln des Lesers (`SPEC-160` Nr. 8 a): höchstens einmal je `ParseURL`, mit dem dekodierten Wert, an der Stelle von `$filter`, nicht mehr, sobald davor ein Verstoß der Klasse `ungueltig` feststeht. Menge 16 = 6 Stellungen von `$filter` in der Query (erste, mittlere, letzte, nach `&&`, `%24filter`, Escape im Wert) mit genau einem Aufruf + 1 leerer Wert + 1 zweites `$filter` + 5 Verstöße davor (`segment`, `key`, `option_name`, `wert`, `kodierung`) + 1 ungültig kodierter Wert (je 0 Aufrufe) + 2 Kompatibilität davor (Pfad, Option; der Leser wird gerufen, 1 Aufruf) | `TestEinhaengungAufruf`: Zähl-Fake hält Argument fest und zählt jeden Aufruf | Leser mit rohem Wert gerufen; Leser trotz `ungueltig` davor gerufen; bei Kompatibilität davor nicht gerufen | andere |
| 20 | Fehler des Lesers (`SPEC-160` Nr. 8 b bis e, V2 nach L7): Ausdruck und Text stehen im Ergebnis; ein `*URLFehler` der Klasse `ungueltig` oder ein `*KompatFehler` mit Ursache `filter` oder `literal_typ` in der Kette gilt mit seiner Ursache und dem Ort `query/<n>`; alles andere ist Fehler des Aufrufers (`model.ErrUngueltig` ohne Fehlerwert, Nullwert); ein Fehler gewinnt über einen gleichzeitig gelieferten Ausdruck. Menge 36 = 12 gültige Fehler (10 Ursachen + 2) + 19 Aufruferfehler (die 18 und die Kette mit beiden Typen, `SPEC-173` Nr. 1 d) (ohne Kette 1, typisierter nil-Zeiger beider Typen roh und verpackt 4, `*URLFehler` der 6 Ursachen von `nicht_unterstuetzt` 6, `*KompatFehler` mit `format`, `pfad`, `option`, `methode` 4, fremde Ursache je Typ 2, weder Ausdruck noch Fehler 1) + 3 Ausdruck mit Fehler + 2 ohne Leser (`New()`, `WithFilter(nil)`: `filter`) | `TestEinhaengungFehler`: Schleife; Fake liefert je Fall den Fehler | `*URLFehler` der Klasse 501 durchgelassen; typisierter nil-Zeiger als Fehler gewertet; Ort des Lesers übernommen | andere |
| 21 | Rang des Lesers in der Lesefolge (`SPEC-160` Nr. 5 mit `SPEC-173` Nr. 6): ein `filter` des Lesers vor einem `wert` einer späteren Option ergibt `wert`; ein `wert` des Lesers vor einer Kompatibilitätsoption ergibt `wert`; die Reihenfolge der Optionen ändert den Baum nicht. Menge 13 = 6 Bäume + `$filter` an 3 Stellen + 2 Rangfolgen + 2 ohne Leser | `TestEinhaengungRang`: Tabelle | Rang der Klassen im Filter ignoriert | andere |
| 22 | Versions-Kopfzeilen ändern nichts (`SPEC-173` Nr. 7, `SPEC-174` Nr. 6): die Signatur `ParseURL(pfad, rawQuery string)` trägt keine Kopfzeile. Menge 1 (Signatur). **Akzeptiertes Negativ für das Verhalten** (gleiche Antwort mit und ohne `OData-Version: 4.0`, `MaxDataServiceVersion: 1.0`, `DataServiceVersion`, `MinDataServiceVersion`): Grund: der Parser sieht keine Header; Folge: ohne Fänger hier; der Fänger liegt im Rundlauf von `slice-v2-bootstrap` (Sendung, Abschnitt 6) | Signatur: Zeile 15 (compile); Verhalten: Rundlauf in `slice-v2-bootstrap` | Parameter für Kopfzeilen eingeführt (Aufbau bricht, gewollt) | andere |
| 23 | Zerlegung des Pfads (`SPEC-161` Nr. 1 mit `SPEC-173` Nr. 2): vor der Dekodierung an rohem `/`; `//`, abschließendes `/`, `.` und `..` (auch `%2E`), `;`, Leerzeichen sind `segment`; `%2F` trennt nicht; `/` allein ist Kompatibilität `pfad`. Menge 18 = `//`, `/S//`, `/S/`, `/S(1)/`, `.`, `..`, `%2E`, `%2e%2e`, `/S/./x`, `;`, Leerzeichen, `%2F` an 3 Stellen, `/`, 3 Gegenfälle | `TestZerlegung`: Tabelle | `%2F` als Trenner; `//` still angenommen | andere |
| 24 | Dekodierung des Pfads (`SPEC-161` Nr. 2): genau einmal (`%2527` ergibt `%27`), Hex-Schreibweisen und rohes Byte gleich, `+` bleibt, `%5C` und Steuerzeichen außer U+0000 sind Text, ungültige Folgen sind `kodierung` und gehen der Grammatik vor. Menge 28 = 1 + 3 + 10 ungültige Folgen (`%`, `%1`, `%G1`, `%1G`, `S%`, `%FF`, `%C3`, `%C0%80`, `%ED%A0%80`, `%00`) × 2 Segmente + 1 `+` + 1 `%5C` + 1 `%01` + 1 Kodierung vor Grammatik | `TestPfadDekodierung`: Schleife | doppelt dekodiert; `%00` angenommen; `+` als Leerzeichen | andere |
| 25 | Erstes Segment (`SPEC-161` Nr. 3 mit `SPEC-173` Nr. 2): `$metadata` (jedes weitere Segment `pfad_form`); `$batch`, `$entity`, `$all`, `$crossjoin` sind Kompatibilität `pfad`; `$value`, `$links`, `$ref`, `$count`, `$Metadata`, `$` sind `segment`. Menge 16 = 1 + 1 (`$metadata/x`) + 4 + 6 + 1 (`%24metadata`) + 3 Namen (`S`, `S(1)`, `S/$count`) | `TestErstesSegment`: Tabelle | `$links` als Kompatibilität (statt `segment`) am ersten Segment; `$Metadata` angenommen | andere |
| 26 | Name und Klammer (`SPEC-161` Nr. 4): Name nach `SPEC-125` Nr. 8 b (Kategorien `L` und `Nd`, nicht `Nl`, `No`, `M`), 1 bis 128 Zeichen; Rest hinter dem Namen leer oder Klammer; `S()` ist Kompatibilität `pfad`. Menge 70 = 29 Zeichen (6 `L`, 4 `Nd`, 3 `Nl`, 4 `No`, 3 `M`, 9 sonstige) × 2 Stellen + Länge 1, 128, 129 + zusammengesetzt und zerlegt 2 + 7 Reste (`S(`, `S(1`, `S)`, `S-1`, `S.x`, `S 1`, `S()`). Die Zeichen innerhalb der Kategorien sind Vertreter: das Prädikat ist `unicode.Is` (**akzeptiertes Negativ**: die Einordnung eines einzelnen Zeichens prüft das Orakel nicht unabhängig; Folge: ein Zeichen, das die Bibliothek anders einordnet als die Spezifikation, würde gleich eingeordnet) | `TestNamen`: Schleife, Orakel `unicode.Is` | `Nl` als Buchstabe; 129 angenommen; Normalisierung | andere |
| 27 | Körper des Schlüssels (`SPEC-161` Nr. 5): Teile, benannt oder unbenannt, nur eine der beiden Formen, Leerraum, nacktes Wort. Menge 39 = Teile 1, 2, 16, 17 (4) + 7 gültige und gemischte Formen (`1`, `Id=1`, `Id=1,K='x'`, `K='x',Id=1`, `1,2`, `1,Id=2`, `Id=1,2`) + 8 leere Teile und Namen (`1,,2`, `,1`, `1,`, `=1`, `Id`, `Id=`, `Id=1,Id=2`, `1=2`) + 2 Namen mit Leerraum + 9 nackte Wörter (`yes`, `t`, `true`, `false`, `TRUE`, `null`, `INF`, `NaN`, `inf`) + 6 Leerraumstellen (um `,` und `=`, Tabulator, `%20`) + 3 Zeichenketten mit `,` | `TestKoerper`: Tabelle mit Ursache | `1,2` angenommen; Reihenfolge benannter Teile sortiert; `inf` als Literal | andere |
| 28 | Einfache Schlüsselliterale (`SPEC-173` Nr. 3): Zeichenkette, Wahrheitswert, ganze Zahl, `L`, Dezimalzahl mit oder ohne `M`; Art und kanonischer Text. Menge 35 = Zeichenketten 8 (`''`, `'a'`, `'O''Hara'`, `'a,b'`, `'(x)'`, `'a=b'`, `'/'`, `%5C`) + Wahrheitswerte 6 + ganze Zahlen 8 (`0`, `-0`, `007`, `42`, `-42`, `0.0`, `1.50`, `-0.50`) + `L` 6 (`42L`, `42l`, `-7L`, `007L`, `0L`, `-0L`) + `M` 7 (`1.5M`, `1.50m`, `0M`, `-0.0M`, `7M`, `7m`, `-1.25M`); Gegenprobe gegen die Beispiele von `SPEC-172` Nr. 4 (kanonischer Text) | `TestSchluesselLiterale`: Tabelle | Text mit `L` oder `M`; `-0` bleibt `-0`; `l` abgelehnt | andere |
| 29 | `guid'…'` (`SPEC-173` Nr. 3): Präfix in jeder ASCII-Schreibweise, 8-4-4-4-12 Hex-Ziffern in jeder Schreibweise, Text kleingeschrieben; jede Abweichung der Form ist `literal`. Menge 42 = 16 Schreibweisen von `guid` (Bitmasken) + 6 Schreibweisen der Ziffern + 20 Randformen (Länge 35 und 37, Trenner an den 4 Stellen fehlend, Gruppenlängen ± 1 an den 5 Gruppen, Zeichen `g`, ohne Anführung, geschweifte Klammern, nicht geschlossen) | `TestGuid`: Schleife über Bitmasken und Randformen | nur `guid` klein erkannt; Text nicht kleingeschrieben; Gruppenlänge nicht geprüft | andere |
| 30 | `datetime'…'` (`SPEC-173` Nr. 3): Präfix in jeder ASCII-Schreibweise; `JJJJ-MM-TTThh:mm`, `:ss`, `.f` mit 1 bis 7 Ziffern, ohne Zone; Bereiche; Text als UTC `JJJJ-MM-TTThh:mm:ssZ` mit Bruchteil ohne Nullen am Ende. Menge 298 = 256 Schreibweisen von `datetime` (Bitmasken) + 42 Formen (21 lesbar, 21 `literal`; die Aufzählung dahinter nannte 40, der Test führt 42 und die dort genannten Klassen: Sekunde fehlt, `:ss` (2), Bruchteil 1 bis 7 Ziffern (7), Nullen am Ende (3), Jahr 0001, 9999, 0000 (3), Monat 12, 13, 00 (3), Tag 31, 32, 00, 29.02.2024, 29.02.2023, 30.02 (6), Stunde 23, 24, Minute 59, 60, Sekunde 59, 60 (6), `Z`, `+01:00`, `-01:00` (3), Datum allein, `t` klein, Leerzeichen statt `T`, 8 Ziffern Bruchteil (4), Leerzeichen am Rand, `.` ohne Ziffer, `T10` ohne Doppelpunkt (3)). **Akzeptiertes Negativ:** Bruchteil über drei Ziffern wird nicht auf die Millisekunde des Modells gekürzt (L8); Folge: ein Schlüssel mit Mikrosekunden findet keine Entity | `TestDatetime`: Schleife, Orakel `time.Date` des Tests | Präfix nur klein; Schaltjahr nicht geprüft; Zone angenommen | andere |
| 31 | Literale außerhalb sind Kompatibilität `literal_typ` (`SPEC-173` Nr. 3): Suffix `d`, `D`, `f`, `F`; Zahl mit Exponent; nichtleeres Präfix außer `guid` und `datetime`. Menge 75 = 25 Formen (4 Suffixe × 2 Zahlformen, 4 Exponenten (`1e3`, `1E3`, `1.5e-2`, `-1e+3`), 13 Präfixe (`binary`, `X`, `x`, `time`, `datetimeoffset`, `geography`, `geometry`, `duration`, `Ns.Typ`, `Edm.Foo`, ein Präfix mit Zeichen jenseits von ASCII, `guidx`, `datetime2`)) × 3 Stellen (unbenannt, benannt, zweiter Teil) | `TestLiteralTyp`: Schleife; nie `*URLFehler` | `datetime2` als `datetime` gelesen; Exponent als `literal`; nur `binary` und `X` | andere |
| 32 | Alles andere ist `literal` (`SPEC-173` Nr. 3): Menge 31 = `null` (3 Schreibweisen), `+1`, `.5`, `5.`, `42LL`, `42Lx`, `1.5LM`, `0x10`, `1_000`, Guid ohne Präfix im Schlüssel, Datum und Zeitpunkt ohne Präfix, Guid in geschweiften Klammern, `guid'x'`, `guid'` offen, `datetime'2020-01-01'`, `datetime'…Z'`, `datetime'…+01:00'`, `'a'b`, `'a`, `'a''`, `1e`, `1E+`, `-`, `--1`, `1 2`, `%20`, `true1`, `yes`; jede Form mit Ursache `literal` (nicht `literal_typ`, nicht `key`) | `TestLiteralVerstoss`: Tabelle | `null` als Schlüssel angenommen; Guid ohne Präfix angenommen (V4-Form) | andere |
| 33 | Länge geht vor Form (`SPEC-160` Nr. 6, `SPEC-161` Nr. 6): 512 Bytes mit Anführungszeichen gültig, 513 `zu_lang`, Ort `pfad/<n>`. Menge 12 = 6 Formen (Zeichenkette, ungeschlossenes `'`, `X'…'`, `guid'…'`, Ziffernlauf, `datetime'…'`) × 2 Seiten | `TestLiteralLaenge`: Generator | Grenze 511 oder `>=`; Form vor Länge | andere |
| 34 | `$count` und weitere Segmente (`SPEC-161` Nr. 7 mit `SPEC-173` Nr. 2 und L4): `S/$count` ist die Anzahl; `$count` hinter einem Schlüssel und jedes Segment nach `$count` ist `pfad_form`; Rest-Segmente sind Kompatibilität `pfad`; ein Rest-Segment, das die Form verfehlt, ist `segment`; Kodierung und Form werden auch hinter einem Kompatibilitätsfehler geprüft. Menge 25 = 15 Pfadformen außerhalb (`/`, `/$batch`, `/$entity`, `/$all`, `/$crossjoin`, `/S()`, `/S(1)/N`, `/S/N`, `/S(1)/$links/N`, `/S(1)/$value`, `/S(1)/N/$value`, `/S(1)/$ref`, `/S(1)/N/$count`, `/S(1)/Ns.T`, `/S(1)/N(2)`) + 10 (`S/$count`, `S/$count/x`, `S(1)/$count`, `$metadata/x`, `$metadata/`, `S/$count/`, Rest leer, Rest mit Zeichen außerhalb, Rest mit offener Klammer, Rest mit `%FF` hinter `$links`) | `TestRestSegmente`: Tabelle mit Ursache und Ort | `$links` als `segment`; `$count` hinter Schlüssel als Anzahl; Kodierung hinter Kompatibilität nicht geprüft | andere |
| 35 | Ein Funktionsimport ist ein Set-Name (`SPEC-174` Nr. 7): kein Fehler des Parsers. Menge 3 = `/GetProducts`, `/GetProducts?x=1`, `/GetProducts(1)` ergeben Menge bzw. Entity wie jeder Name | `TestFunktionsimport`: Tabelle | Namen mit `Get` oder Liste bekannter Namen abgelehnt | andere |
| 36 | Zerlegung der Query ist `model.QuerySegmente` (`SPEC-162` Nr. 1): `&&`, `;` trennt nichts, `a=b=c`, `=v`, ohne `=`, leere Query. Menge 8 | `TestQueryZerlegung`: Ergebnis des Parsers gegen die Referenz-Zerlegung | an `;` getrennt; eigene Zerlegung mit anderem erstem `=` | andere |
| 37 | Dekodierung der Query (`SPEC-162` Nr. 2): Name und Wert je genau einmal, `+` bleibt, ungültige Folgen sind `kodierung`; der Wert einer Kompatibilitätsoption wird nicht gelesen (`%FF` im Wert von `$expand` ist `option`, nicht `kodierung`); Folge der Prüfungen: Kodierung des Namens, Klasse, Wiederholung, Kodierung des Werts, Grammatik. Menge 21 = Name 7 (darunter `%2524top`, der Name wird einmal dekodiert) + Wert 6 + 4 Paare der Folge + 2 Kompatibilitätswerte + `%2524` + `Name+desc` | `TestQueryDekodierung`: Tabelle | doppelt dekodiert; Wert einer Kompatibilitätsoption dekodiert | andere |
| 38 | Klassen der Namen (`SPEC-162` Nr. 3 mit `SPEC-173` Nr. 4): die sieben behandelten Namen doppelt sind `option_doppelt`; unbekannter Name mit `$` ist `option_name`; Alias und Herstellername sind `Sonstige`; die Schreibweise zählt. Menge 48 = 7 Namen × 3 Schreibweisen des zweiten (21) + 10 `option_name` (`$Top`, `$foo`, `$`, `$links`, `$value`, `$ref`, `$batch`, `$FORMAT`, `$Format`, `$expandx`) + 8 `Sonstige` (`@a`, `@a1`, `sap-client=100`, `x`, `x=`, `api-version=2`, `a.b=1`, `_x=1`) + 9 weitere (`@`, `@1`, `?x`, `=v`, ein Name jenseits von ASCII, `1x=1`, `-x=1`, Name mit 129 Zeichen, mit 128 Zeichen = `Sonstige`) | `TestNamenKlassen`: Schleife | `$Top` als `$top`; `option_doppelt` nur gleiche Schreibweise des Escapes | andere |
| 39 | Optionen außerhalb sind Kompatibilität `option`, mit jedem Wert und beliebig oft (`SPEC-173` Nr. 4 mit L3): `$expand`, `$skiptoken`, `$callback`, `$id`, `$count`, `$search`, `$apply`, `$compute`, `$deltatoken`, `$levels`, `$index`, `$schemaversion`. Menge 79 = 12 Namen × 3 Werte (leer, nichtleer, ungültig kodiert) × 2 Anzahlen (1, 2) + 7 Gegenfälle (die Optionen im Umfang mit gültigem Wert: keine Kompatibilität). Die Liste der 12 trägt der Test aus `SPEC-173`, nicht aus dem Code | `TestOptionenAusserhalb`: Schleife über die Liste des Tests, Zählprüfung 12 | eine der 12 als `option_name`; zweites Vorkommen als `option_doppelt` | andere |
| 40 | `$top` und `$skip` (`SPEC-162` Nr. 4): ASCII-Ziffern, 0 bis 2147483647. Menge 70 = 2 Optionen × 35 Werte (13 Werte: `0`, `00`, `007`, `1`, `2147483647`, `2147483648`, 20 Ziffern, `-1`, `-0`, `+1`, `1.0`, `1e3`, `0x10`; leer; Randleerraum 9 Zeichen (U+0020, U+0009, U+000A, U+000D, U+0085, U+00A0, U+2003, U+3000, U+FEFF) × 2 Seiten = 18; 3 Ziffern jenseits von ASCII) | `TestTopSkip`: Schleife | `TrimSpace` auf dem Wert; Überlauf 2147483648 angenommen; `+1` angenommen | andere |
| 41 | `$inlinecount` (`SPEC-173` Nr. 4 mit L9): `allpages` ergibt `Count{Gesetzt: true, Wert: true}`, `none` ergibt `Count{Gesetzt: true, Wert: false}`, die Schreibweise zählt, jeder andere Wert ist `wert`; doppelt ist `option_doppelt`; zusammen mit `$count=true` gilt die Kompatibilität `option` des `$count`. Menge 34 = 2 gültig + 9 falsch (`ALLPAGES`, `AllPages`, `None`, `all`, `true`, `false`, `1`, `allpages,none`, `allpagesx`) + leer + 18 Randleerraum + 4 (doppelt, mit `$count` davor und dahinter, `Count` ohne Option nicht gesetzt) | `TestInlinecount`: Schleife | Schreibweise ignoriert; `TrimSpace`; `none` als nicht gesetzt | andere |
| 42 | `$select` (`SPEC-162` Nr. 6 mit `SPEC-173` Nr. 4): `*` und Namen in der Reihenfolge der URL; Pfad mit `/` ist Kompatibilität `option` (nicht `format`, nicht `wert_nicht_unterstuetzt`); sonst `wert`; 256 Elemente gültig, 257 `zu_viele`. Menge 21 = 5 gültig + 9 `wert` (leer, `Id,,Name`, `,Id`, `Id,`, `Id, Name`, `Id,(x)`, `Id,Id`, `*,*`, Name mit 129 Zeichen) + 4 Pfade + 2 Grenzen (256 gültig, 257 `zu_viele`) + 1 Schreibweise `id` gegen `Id` | `TestSelect`: Tabelle und Generator der Grenze | Pfad als `wert`; Reihenfolge sortiert; doppelter Name angenommen | andere |
| 43 | `$orderby` (`SPEC-162` Nr. 7 mit `SPEC-173` Nr. 4 und L5): Richtung `asc` und `desc` in jeder ASCII-Schreibweise, genau ein Leerzeichen; ein `(` im Wert (`tolower(Name)`) oder ein Pfad mit `/` ist Kompatibilität `option`, der Wert wird nicht weiter gelesen; sonst `wert`; 32 Elemente gültig, 33 `zu_viele`. Menge 24 = 8 gültig (darunter 32 Elemente) + 10 `wert` (leer, `,A`, `A,`, `A  desc`, `A up`, `A asc desc`, `A mul 2`, `A,A`, `A,A desc`, 33 Elemente) + 6 Kompatibilität (`tolower(Name)`, `tolower(Name) desc`, `A,tolower(B)`, `C/Name`, `C/Name desc`, `A,C/Name`) | `TestOrderBy`: Tabelle | `(` als `wert`; `desc` nur klein; zwei Leerzeichen angenommen | andere |
| 44 | `$format` (`SPEC-173` Nr. 4 mit L2): ohne Beachtung der ASCII-Schreibweise `json`, `application/json` oder `application/json;` mit beliebigem Rest ist lesbar und ohne Eintrag im Ergebnis; jeder andere nichtleere Wert ist Kompatibilität `format`; leer ist `wert`; `$Format` ist `option_name`; doppelt ist `option_doppelt`. Menge 41 = 8 gültige Schreibweisen (`json`, `JSON`, `Json`, `application/json`, `Application/JSON`, `application/json;`, `application/json;odata=verbose`, `application/json;odata=minimalmetadata;streaming=true`) × 2 Namenschreibweisen (`$format`, `%24format`), das Ergebnis gleich dem ohne `$format` (16) + 2 Werte mit Leerraum (`json `, ` json`: `format`) + 10 außerhalb (`atom`, `xml`, `application/atom+xml`, `application/xml`, `text/plain`, `jsonx`, `application/jsonx`, `application/json-patch+json`, `verbose`, `ATOM`) × 2 Namenschreibweisen (20) + leer + `$Format` + doppelt | `TestFormat`: Schleife | `atom` als lesbar; `odata=verbose` als `format`; Eintrag im Ergebnis; `TrimSpace` | andere |
| 45 | `$filter` ohne Leser (`SPEC-160` Nr. 8 mit `SPEC-173` Nr. 4): nichtleerer lesbarer Wert ist Kompatibilität `filter` und wird nicht weiter gelesen; leer ist `wert`; ungültig kodiert ist `kodierung`; die letzte Option `WithFilter` gilt. Menge 10 = `Id eq 1`, `(`, `x`, leer, `%FF`, doppelt, `ungueltig` davor, `New()`, `New(WithFilter(nil))`, `New(WithFilter(l), WithFilter(nil))` | `TestFilterOhneLeser`: Tabelle | ungefilterte Antwort (kein Fehler); Wert gelesen | andere |
| 46 | Gleichwertigkeit (`SPEC-161` Nr. 8, `SPEC-162` Nr. 9): Reihenfolge der Systemoptionen, Schreibweise der Hex-Ziffern, Escape gegen rohes Zeichen, `true`/`false`, `asc`/`desc`, leere Segmente, führende Nullen ändern das Ergebnis nicht; Schreibweise der Namen, Reihenfolge der Elemente, `+` gegen `%20`, Reihenfolge von `Sonstige` ändern es. Menge 136 = 120 Permutationen von 5 Optionen (`$top`, `$skip`, `$select`, `$orderby`, `$inlinecount`) + 16 Paare (12 gleich, 4 verschieden) | `TestGleichwertig`: Permutationsschleife und Tabelle | Reihenfolge der Optionen im Ergebnis; `+` als Leerzeichen | andere |
| 47 | Allaussage: jede V2-Eigenschaft außerhalb des Umfangs, die der Parser kennt, ist ein `*KompatFehler` mit der erwarteten Ursache, nie ein `*URLFehler`, nie ein Erfolg (`SPEC-170` Nr. 3, `LH-FA-17` Negative). Menge 63 Elemente aus dem Wortlaut der Spezifikation (12 Optionen, 15 Pfadformen, 10 Formate, 25 Literalformen, 1 `$filter` ohne Leser) × 3 Einbettungen (allein; hinter gültigem Pfad und Option; vor einer zweiten gültigen Option) = 189; Zählprüfung: Menge 63 und alle fünf Ursachen getroffen. Operatoren und Funktionen: B; Methoden: Mock (`slice-v2-bootstrap`); Typen: kein Modell hier (`SPEC-171` Nr. 5) | `TestAllaussageKompat`: Schleife über Listen des Tests (nicht aus dem Code); Zählprüfung der Listen | ein Element der Listen (je Liste das letzte) nicht erkannt; Erfolg statt Fehler | andere |
| 48 | Generator über die Klasse der URLs: jede URL aus Pfadform (40) × Optionsfolge (0 bis 4 aus 30 Formen) wird vom Leser genauso klassifiziert wie vom Orakel des Tests (eigene Funktion, nicht der Leser; Rangfolge Stufe 0, `ungueltig` vor Kompatibilität, erster in der Lesefolge) und, wo sie gültig ist, gleich dem Renderer. Menge: 2000 Läufe, Saat 1 bis 2000; Zählprüfung: jede der 15 Ursachen (10 `ungueltig`, 5 Kompatibilität) mindestens 20 Mal, Erfolge mindestens 300 | `TestGeneratorURLs`: Saat je Lauf im Fehlertext | Rangfolge der Klassen vertauscht; Orakel und Leser teilen eine Funktion (Gegenprobe: das Orakel steht in `orakel_test.go` ohne Import aus dem Produktcode) | andere |
| 49 | Echte V2-URLs (Referenzdienst Olingo, Erfassung vom 2026-11-01; der Anhang A, den dieser Plan nannte, liegt nicht vor, die Tabelle steht im Test, §6): 29 verschiedene Anfragen (aus den 45 der Erfassung; der Header `Accept` ändert die URL nicht), 18 Erfolge mit erwartetem `ODataURL`, 5 Kompatibilität `pfad`, 1 `option`, 5 `filter` (ohne Leser; mit Leser in B). Menge 29 | `TestEchteV2URLs`: Tabelle im Test (Pfad und Query der Anfragenlisten der Erfassung) | `$format=json` als `format`; `$inlinecount=allpages` als `option`; `Employees('1')/ne_Team` als Erfolg | andere |

**Mutationen und Raten** (Implementer, `AGENTS.md` §3.10; jede Mutation auf einer Wegwerfkopie im Image `odata-recorder:race` gefahren, rot gesehen am 2026-11-01; Zahl in Klammern = Zahl der Mutationen der Zeile; der Test ist der benannte, weitere rote Tests stehen nicht hier):

| Zeile | Mutationen (Funktion: Änderung) | roter Test |
|---|---|---|
| 1 | `pruefeArgumente`: `#` im Pfad erlaubt; Grenze `< 0x1f`; U+007F nicht geprüft (3) | `TestArgumente` |
| 2 | `ParseURL`: Ergebnis in einer Paketvariablen zwischengespeichert (1) | `TestWiederholung` |
| 3 | `leseZahl`: `$top=0` nicht gesetzt; `leseQuery`: `OrderBy` leer statt nil; `leseKey`: Name des unbenannten Teils gefüllt (3) | `TestErgebnisForm` |
| 4 | Grenzen pfad 2047, pfad `>=`, rawQuery 8193, Segmente 15, Query-Segmente 65, Teile 15, Literal 511, `$select` 255, `$orderby` 33, Name 127 in `einfacherName`, Optionsname ungeprüft (11) | `TestGrenzen` |
| 5 | `gesamtform`: Query vor Pfad; `zu_viele` (pfad) vor `zu_lang` (pfad) (2) | `TestStufeNull` |
| 6 | `dekodiere`: Verkettung in einer Schleife (quadratisch); `leseQuery`: Wert einer Sonstigen Option verdoppelt (2) | `TestWaechter` |
| 7 | `ParseURL`: Lesen endet bei der Kompatibilität des Pfads; Kompatibilität vor `ungueltig` der Query gemeldet (2) | `TestUngueltigVorKompat` |
| 8 | `kompat`: letzter statt erster; Query vor Pfad gelesen (2) | `TestErsteKompat` |
| 9 | Query vor Pfad gelesen; `lesePfad`: letzter statt erster Verstoß (2) | `TestErsterUngueltig` |
| 10 | `lesePfad`: Eingabeteil im Ort; Ort um 1 verschoben (2) | `TestUrlFehlerWert` (die zweite auch viele andere) |
| 11 | `ParseURL`: `*URLFehler` mit `filter` statt `*KompatFehler`; Ort verlängert (2) | `TestKompatFehlerWert` |
| 12 | `AlleKompatUrsachen` ohne `methode`; Ort ohne Quote; `Unwrap` trifft `ErrURLNichtUnterstuetzt`; `Code` ist `ODR-E6005`; `Unwrap` verbindet auch `ErrURLUngueltig` (5) | `TestKompatFehlerModell` |
| 13 | Zeile der Abbildung fehlt; Zeile hinter der Gruppe `ODR-E9001` (2) | `TestAbbildungKompat` |
| 14 | Eintrag fehlt; Status 500; Text geändert; `CodeE6006` hinter `CodeE9001`; Zeile vor `CodeE6005` (5) | `TestCodeTabelle`; `TestCountErrorUnbekannt` und die Zahlen der Telemetrie |
| 15, 22 | `ParseURL(pfad string)` ohne `rawQuery`: der Bau bricht (gewollt, 1) | Bau von `TestSatzTraegtURL` |
| 16 | Import von `urlparser` im Produktcode; Import von `csdl` in einer Testdatei (2) | `TestKeinV4Import` |
| 17, 18 | siehe Raten | siehe Raten |
| 19 | `leseOption` mit rohem Wert für `$filter`; Leser trotz Pfad-Verstoß davor gerufen; Leser vorab gerufen (Query-Verstoß davor); bei Kompatibilität davor nicht gerufen (4) | `TestEinhaengungAufruf` |
| 20 | `*URLFehler` der Klasse 501 durchgelassen; typisierter nil-Zeiger gewertet; Ort des Lesers übernommen; beide Typen in der Kette gewertet (4) | `TestEinhaengungFehler` |
| 21 | `fehlerDesLesers`: `*URLFehler` wie Kompatibilität eingereiht (1) | `TestEinhaengungRang` |
| 23 | `%2F` als Trenner; leeres Segment übersprungen (2) | `TestZerlegung` |
| 24 | doppelt dekodiert; U+0000 angenommen; `+` als Leerzeichen (3) | `TestPfadDekodierung` |
| 25 | `$links` als Kompatibilität am ersten Segment; `$Metadata` angenommen (2) | `TestErstesSegment` |
| 26 | `Nl` als Buchstabe; 129 Zeichen angenommen (2); Normalisierung: akzeptiertes Negativ, kein Fänger | `TestNamen` |
| 27 | `1,2` angenommen; benannte Teile sortiert; `inf` als Literal (3) | `TestKoerper` |
| 28 | Text mit Suffix; `-0` bleibt `-0`; `l` abgelehnt (3) | `TestSchluesselLiterale` |
| 29 | `praefixLiteral`: Präfix nur klein; Text nicht kleingeschrieben; Trenner an der vierten Stelle nicht geprüft; Gesamtlänge nicht geprüft (4) | `TestGuid` |
| 30 | Präfix nur klein; Schaltjahr nicht geprüft; Zone angenommen; Bruchteil nur 6 Ziffern (4) | `TestDatetime` |
| 31 | `datetime2` als `datetime`; Exponent nicht erkannt; nur `binary` und `X` (3) | `TestLiteralTyp` |
| 32 | `null` als Schlüssel angenommen; Guid ohne Präfix angenommen (2) | `TestLiteralVerstoss` |
| 33 | Grenze 511 (`>=`); Form vor Länge (2) | `TestLiteralLaenge` |
| 34 | `$links` als `segment`; `$count` hinter Schlüssel als Anzahl; nach der Kompatibilität nicht weiter gelesen (3) | `TestRestSegmente` |
| 35 | Name mit `Get` abgelehnt (1) | `TestFunktionsimport` |
| 36 | an `;` getrennt; Wert nach dem letzten `=` (2) | `TestQueryZerlegung` |
| 37 | Name doppelt dekodiert; Wert einer Kompatibilitätsoption dekodiert (2) | `TestQueryDekodierung` |
| 38 | `$Top` als `$top`; `option_doppelt` nur bei gleichem Escape (2) | `TestNamenKlassen` |
| 39 | `$schemaversion` fehlt; zweites Vorkommen einer Kompatibilitätsoption als doppelt (2) | `TestOptionenAusserhalb` |
| 40 | `TrimSpace` auf dem Wert; Überlauf 2147483648 angenommen; `+1` angenommen (3) | `TestTopSkip` |
| 41 | Schreibweise ignoriert; `TrimSpace`; `none` als nicht gesetzt (3) | `TestInlinecount` |
| 42 | Pfad als `wert`; Namen sortiert; doppelter Name angenommen (3) | `TestSelect` |
| 43 | `(` als `wert`; `desc` nur klein; zwei Leerzeichen angenommen (3) | `TestOrderBy` |
| 44 | `atom` lesbar; `odata=verbose` als `format`; Eintrag im Ergebnis; `TrimSpace` (4) | `TestFormat` |
| 45 | `$filter` ohne Leser ergibt keinen Fehler; die erste Option `WithFilter` gilt (2) | `TestFilterOhneLeser` |
| 46 | `$top` entfällt hinter `$skip`; `+` als Leerzeichen (2) | `TestGleichwertig` |
| 47 | je Liste das letzte Element nicht erkannt: `$schemaversion`, `/S(1)/N(2)`, `ATOM`, `datetime2'a'`, `$filter` (5) | `TestAllaussageKompat` |
| 48 | Rangfolge der Klassen vertauscht (Query vor Pfad, Kompatibilität vor `ungueltig`) und Orakel: `orakel_test.go` importiert aus dem Produktcode nichts | `TestGeneratorURLs` |
| 49 | `$format=json` als `format`; `$inlinecount` als `option`; `ne_Team` als Erfolg; `$filter` ohne Leser als `option` (4) | `TestEchteV2URLs` |

**Raten** (`AGENTS.md` §3.10; 30 Prozesse je Zeile, `-test.count=1` und `-test.count=20`, Image `odata-recorder:race` mit vier CPUs, gemessen vom Implementer): die Prüfung `TestGleichzeitig` (Zeile 17) und `TestEinhaengungGleichzeitig` (Zeile 18) bleibt ohne Mutation in 30 von 30 Prozessen grün. Zeile 17, Wirkungs-Fänger: erster Kompatibilitätsfehler im Feld der Instanz rot in 30 von 30 (mit und ohne Detektor, `-test.count=1` und `-test.count=20`); Puffer der Dekodierung als Paketvariable rot in 30 von 30 (ebenso; dafür führt die Menge sechs Eingaben mit je 300 Escapes verschiedener Zeichen); ein Zähler als Paketvariable ohne Wirkung auf das Ergebnis rot nur mit dem Detektor, 30 von 30 Prozessen mit `-race`, 0 von 30 ohne. Zeile 18: ein Zähler im Parser (Datenrennen ohne Wirkung) und ein gemeinsamer Wert in der Leseroutine rot nur mit dem Detektor, je 30 von 30 mit `-race` und 0 von 30 ohne; der Leser pro Aufruf nicht gerufen und der Leser zweimal gerufen sind Wirkungs-Fehler und rot in jedem Lauf (Zähler 320). Das Fehlen einer Sperre trifft allein `make test-race`.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` (erfüllt), und der Vertrag samt Randformen steht in `SPEC-173` (mit den neun Berichtigungen des Architect) und `SPEC-174` (erfüllt); die Tabelle *Zusage · Fänger* und das Raster stehen im Plan (erfüllt, `AGENTS.md` §3.12). Der Slice ist startbar ohne `welle-mock`.

- **Code-Abhängigkeit, mit Grund:** `slice-odata-url-pfad` und `slice-odata-url-filter` (beide in `done/`: Gerüst des Parsers und Query-AST, auf die V2 abbildet), Modell und Fehlermodell (`done/`). Der Parser kennt das Modell nicht (`SPEC-173`); `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-edm-ausdruecke` und `slice-query-interpreter` sind keine Voraussetzung, der Test über beide Seiten (Parser plus Interpreter) wandert in den Rundlauf von `slice-v2-bootstrap`. Der Schnitt von Pfad und Optionen (`SPEC-173` Nr. 1 bis 4, 6, 7) und `$filter` (Nr. 5), die allein die Literale (Nr. 3) und die Einhängestelle `WithFilter` teilen, ist entschieden (siehe Größenmarke): dieser Slice ist Teil A, `$filter` ist `slice-v2-query-filter`, das nach diesem Slice in `done/` startet (Literal-Funktion und Einhängestelle liegen hier).
- **Was der Slice nicht braucht — und warum:** `welle-mock` ganz, `MockUseCase`, Workspace und HTTP-Adapter (der Parser liest Text; der Rundlauf gehört `slice-v2-bootstrap`), M2 und `slice-bootstrap-m3` (das Binary ändert sich erst mit der Wahl des Adapters), `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-expand-batch`. Der V2-Parser bildet auf den AST der Mock-Slices ab und hängt an dessen Form, nicht am Use Case.
- **Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`; die Faktoren stehen allein in `SPEC-128` Nr. 1, die Zahlen von §3.14 sind dort ältere): gezählt ist die Tabelle in §3 mit 49 Zeilen, *a* = 2 nebenläufige (Zeilen 17 und 18), *b* = 47 andere, *f* = 1 (der Zähl-Fake des `FilterLeser`, `fakeFilterLeser` in einer Testdatei von `urlparserv2`: hält das Argument fest, zählt jeden Aufruf atomar, liefert je Fall Ausdruck und Fehler), *p* = 1000 (Erwartung des Architect, Spanne 900 bis 1150). *S* = 85·2 + 80·47 + 200·1 + 1000 = **5130**, 135·5130 = 692550, Marke *M* = 100 · ⌈692550 / 10000⌉ = **7000**. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks (`src/adapters/driven/odata/urlparserv2/`, die Fehlerwerte und Abbildung in `src/hexagon/model/` und `src/hexagon/ports/outbound/`, die Tests der Telemetrie und `klassen_test.go`), Tests eingeschlossen. Die Marke ersetzt die Prosa-Zeile ohne Zahl, die der Plan bis zum Start trug, mit Grund: die Tabelle und der `produktcode` des Architect liegen jetzt vor, die Zahl ist gezählt, nicht geraten. Erwartung der Messung (Ende ÷ *S* war 0,74 bei `slice-odata-url-pfad`, 0,69 bei `slice-odata-url-filter`, 0,78 bei `slice-v2-metadaten`): rund 3800, Spanne 3400 bis 4400. Das Vorbild lag in V4 ebenfalls in zwei Slices.
  **Schnittfrage** (*S* ≥ 2223): gestellt und beantwortet. Ungeschnitten wäre *S* = 8500 (Marke 11500); geschnitten ist `slice-v2-query-filter` (neue Kennung, Teil B: der Leser von `$filter`, `SPEC-173` Nr. 5; 37 Zeilen, *S* = 3370, *M* = 4600, `produktcode` 400) der Teil ohne geteilten Mechanismus, denn er teilt mit dem Rest allein die Literal-Funktion und die Einhängestelle `WithFilter`; die Grammatik, der Lexer, die Tiefe und die Rangfolge der Operatoren gehören nur ihm. A wird nicht weiter geschnitten (`schnitt-grund` im Block); B startet nach A in `done/`. Der Plan von B liegt auf `main` (Adresse in §6).

```groessenmarke
pfad: src/adapters/driven/odata/urlparserv2/
pfad: src/hexagon/model/kompat_fehler.go
pfad: src/hexagon/model/kompat_fehler_test.go
pfad: src/hexagon/model/fehler.go
pfad: src/hexagon/model/fehler_test.go
pfad: src/hexagon/ports/outbound/abbildung.go
pfad: src/hexagon/ports/outbound/abbildung_test.go
pfad: src/hexagon/ports/outbound/codec_satz.go
pfad: src/adapters/driven/telemetry/regeln_test.go
pfad: src/adapters/driven/telemetry/telemetry_test.go
pfad: src/adapters/driving/httpodata/klassen_test.go
pfad: src/adapters/driving/httpodata/fehler_test.go
pfad: src/hexagon/services/session/beenden_test.go
pfad: src/hexagon/services/workspace/helfer_test.go
pfad: src/hexagon/services/workspace/kontext_test.go
pfad: src/hexagon/services/workspace/reset_test.go
produktcode: 1000
fakes: 1
basis: ab7b3083cb26cc3bd2a8a34ddb77c51c471a5a6b
schaetzung: 5130
marke: 7000
schnitt-grund: Pfad, Schlüsselliterale, Systemoptionen und Kompatibilitätsfehler teilen Dekodierung, Namensregeln, Grenzen, die Rangfolge der Verstöße und die Fehlerwerte. Der Teil ohne geteilten Mechanismus, die Grammatik von $filter, ist als slice-v2-query-filter abgetrennt. KompatFehler, Code und Abbildung (rund 65 Zeilen Produktcode) sind allein kein Slice: der Parser erzeugt den Wert und kann ohne ihn nicht gebaut werden.
```

  `basis` ist der Commit, der den Vertrag abschließt (die Berichtigung von `SPEC-173`); bis zum ersten Code-Commit ändern sich die Pfade nicht (Messung 0).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen): über 7000 Zeilen. Eine Überschreitung schneidet nicht (`SPEC-128` Nr. 10 c); sie bekommt einen Eintrag `ueberschreitung` mit Grund im Block, die Rückführung entscheidet der Planner.
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`, oder der AST trägt ein Konstrukt des Umfangs nicht.

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

- Die Menge der URLs ist unbeschränkt (jede Zeichenfolge) — **Ausgang:** akzeptiertes Negativ mit Prüfgrenze: Zeile 48 prüft 2000 Läufe mit Saat 1 bis 2000 mit Zählprüfung (15 Ursachen, mindestens 20 Treffer je Ursache, mindestens 300 Erfolge); jenseits davon nur die Mengen der Zeilen 4 bis 47. Folge im Betrieb: eine Eingabeform außerhalb der erzeugten Klasse bleibt ungefangen, bis Review und Verifier sie mit eigenem Generator (Saat ab 100000) prüfen (`AGENTS.md` §3.17).
- Der AST trägt ein Konstrukt nicht — **Ausgang:** entfallen: der AST liegt in `done/` (`SPEC-165`), jedes Konstrukt ohne Knoten ist Kompatibilität (`SPEC-173` Nr. 5, in B).
- Zahl der Fehlerfälle: der Kompatibilitätsfehler ist eine Allaussage über jede V2-Eigenschaft außerhalb des Umfangs — **Ausgang:** entschieden: Zeile 47, Schleife über 63 Elemente der Listen der Spezifikation (12 Optionen, 15 Pfadformen, 10 Formate, 25 Literalformen, 1 `$filter` ohne Leser) mal drei Einbettungen, mit Zählprüfung; Operatoren und Funktionen gehören B, Methoden dem Mock.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 4×) — **Ausgang:** entschieden: `fakeFilterLeser` (Zeilen 18 bis 21) hält jedes Argument fest, zählt atomar und liefert je Fall Ausdruck und Fehler; `TestFakeFilterLeser` prüft den Zähl-Fake gegen sich selbst (400 Aufrufe aus 8 Goroutinen, Argumente und Zähler gleich).
- Randformen (`AGENTS.md` §3.12, §3.16) — **Ausgang:** entschieden: das Raster unten hat keine Zelle ohne Ort, die neun Lücken von `SPEC-173` (L1 bis L9 des Architect) stehen vor dem Code in der Spezifikation.
- Die Kopien von V4 laufen auseinander (kanonischer Text von Datum, Zahl, Guid) — **Ausgang:** akzeptiertes Negativ mit Grund und Folge in §3; Fänger Zeilen 28 bis 30 und der Rundlauf in `slice-v2-bootstrap`.
- Das Einfügen von `CodeE6006` hinter `CodeE6005` verschiebt die Zahlenwerte ab `ODR-E6101` um eins — **Ausgang:** entschieden: kein Produktcode wertet die Zahl aus, und die Telemetrie-Tests brauchen `CodeE9001+1` als Code außerhalb der Tabelle, was ein Anhängen ans Ende verbietet; Zeile 14 fängt es.
- Die Zahl „35“ und „34 Paare“ in Kommentaren und Tests der Telemetrie und in `klassen_test.go` bleibt stehen — **Ausgang:** entschieden: nachgezogen in den Dateien des Blocks; die Suche nach `35` und `34 Paare` fand fünf weitere Dateien (Dateitabelle in §3), in denen die Zahl der Paare, die Liste der Fehler der Abbildung und zwei Fallzahlen (19 × 25, 4 × (25 + 4 + 1)) an der Tabelle hängen.
- `versionstrennung_test.go` (Zweig `slice-v2-metadaten`) deckt `urlparserv2` mit — **Ausgang:** entschieden: Zeile 16 hat einen eigenen, kleinen Test, unabhängig von der Merge-Reihenfolge; die Doppelung ist bewusst.
- Eingaben mit Escape-Folge oder Rohbytes — **Ausgang:** entschieden: Werte jenseits von ASCII entstehen aus Codepunktzahlen (`zeichen`) und `%XX` je Byte (`pct`) in `helfer_test.go`; gegengelesen mit `LC_ALL=C grep -P` über alle `.go`-Dateien von `urlparserv2` und über `kompat_fehler.go` und `kompat_fehler_test.go`: keine Folge aus Backslash, `u` und vier Hex-Ziffern, nicht-ASCII nur `ä`, `ö`, `ü`, `ß` und `×` in Kommentaren, Meldungstexten und Tabellen (`AGENTS.md` §3.18).
- Der Anhang A, auf den Zeile 49 verwies (die 29 verschiedenen Anfragen der Erfassung), stand nie im Plan — **Ausgang:** weiter offen, Adresse Planner: die Tabelle steht im Test `TestEchteV2URLs` (18 Ergebnisse, 5 `pfad`, 1 `option`, 5 `filter`, aus den Anfragenlisten der Erfassung); der Planner nimmt sie als Anhang auf oder streicht den Verweis.
- Zeilen 7 bis 9 und 41: die Zählung der Architect-Fassung weicht von der ausführbaren Menge ab (38 statt 41, 18 statt 19, `$format` und `$filter` zweimal sind `option_doppelt`, vier Fälle bei `$inlinecount`) — **Ausgang:** entschieden: der Test zählt die ausführbaren Stellungen und prüft die Zahl; die Zeilen oben tragen die gezählten Mengen (`AGENTS.md` §3.9).
- Der Plan nannte den Sentinel `ErrNichtUnterstuetzt` des Modells (Zeile 12) — **Ausgang:** entfallen: er existiert nicht; der vierte Gegen-Sentinel ist `ErrDatenUngueltig`.
- Lehren aus `slice-v2-metadaten` (beide Seiten jeder Grenze, Kleinschreibung reservierter Anfänge, innere Zeichen eines Bereichsprädikats) — **Ausgang:** eingearbeitet in die Zeilen 4, 26, 29, 30, 33, 40, 41, 43 und 44 (beide Seiten, alle Schreibweisen der Präfixe durch Bitmasken, Randleerraum für jedes Zeichen); die inneren Zeichen von Zeile 26 stehen als akzeptiertes Negativ dort.
- Fänger „Versions-Kopfzeilen ändern nichts“ (Zeile 22) hat hier nur die Signatur — **Ausgang:** weiter offen, Adresse `slice-v2-bootstrap` (Rundlauf, Sendung unten).

**Sendungen und Rückgaben (`AGENTS.md` §3.13).** In diesem Zweig liegt nur dieser Plan; die Pläne der Nehmer liegen auf `main` und sind nicht geprüft. Jede Zeile hat Ausgang *offen*, bis der Nehmer sie trägt.

| An | Gegenstand | Ausgang |
|---|---|---|
| `slice-v2-query-filter` (Plan von A auf `main` anzulegen) | Leser von `$filter` (`urlparserv2.NewFilterLeser()`), Fehlerzuordnung nach `SPEC-173` Nr. 1 (c), (d), Kompatibilität `filter` und `literal_typ` im Ausdruck, `SPEC-173` Nr. 5 mit der Berichtigung zu `has`, `in`, `divby`, Literal-Funktion und Einhängestelle aus A | offen, Adresse `slice-v2-query-filter`; sein §1 darf `$filter` nicht ausschließen |
| `slice-v2-bootstrap` | (a) Verdrahtung `CodecSatz{URL: urlparserv2.New(urlparserv2.WithFilter(urlparserv2.NewFilterLeser()))}` für Version 2.0; (b) Start-Liste um `slice-v2-query-filter` erweitern (das Beispiel `substringof` verlangt B); (c) Rundlauf-Fänger „Versions-Kopfzeilen ändern nichts“ (`SPEC-174` Nr. 6): dieselbe Antwort mit und ohne `OData-Version: 4.0` und `MaxDataServiceVersion: 1.0`; (d) Handbuch: Eintrag `ODR-E6006`; (a) und (d) stehen dort schon, (b) und (c) sind neu | offen, Adresse `slice-v2-bootstrap` |
| `slice-handbuch-meldungskatalog` | `ODR-E6006` ist ein Code, den nur ein V2-Dienst erreicht (Menge H des Gates) | offen, Adresse `slice-handbuch-meldungskatalog` |
| `welle-odata-v2` und Roadmap | Schnitt in A und B, neue Kennung, Marke 7000 und 4600 | offen, Adresse `welle-odata-v2` |
| `slice-v2-payload-schreiben` | das Feld `Schreiber` in `CodecSatz` (nicht A); zur Kenntnis: A ändert `codec_satz.go` im selben Verzeichnis, Konfliktpunkt beim Merge | zur Kenntnis |

**Raster *Operation × Argumentart*** (entschieden; jede Zelle trägt einen Ort):

| Operation | Option fehlt oder leer | unbekannte Option, Funktion oder Methode (außerhalb des Umfangs) | fremde Syntax (V4 am V2-Parser, anderes Format) | Literale in V2-Schreibweise, Typfehler | Wiederholung (doppelte Option), über Grenze (Länge, Tiefe) | Abbruch (Kontext) |
|---|---|---|---|---|---|---|
| Ressource: Collection, Key, `$count` | `/` allein: Kompatibilität `pfad` (`SPEC-173` Nr. 2); keine Query: gültig (`SPEC-162` Nr. 1); `S()`: `pfad` (`SPEC-161` Nr. 4) | Rest-Segment, `$links` (L4), `$batch`, `$entity`: `pfad` (`SPEC-173` Nr. 2); Funktionsimport ist ein Name (`SPEC-174` Nr. 7) | `S(1)/$count`: `pfad_form`; `$value`, `$links`, `$ref` als erstes Segment: `segment` (`SPEC-173` Nr. 2); Guid, Datum, Zeitpunkt ohne Präfix im Schlüssel: `literal` (`SPEC-173` Nr. 3) | `SPEC-173` Nr. 3 (Tabelle, `literal_typ`, `literal`) | `Id=1,Id=2`: `key` (`SPEC-161` Nr. 5 c); Grenzen `SPEC-160` Nr. 6 | gilt nicht: `ParseURL` hat keinen Kontext (`SPEC-160` Nr. 1) |
| Option `$filter` (Grammatik in B; hier nur Einhängestelle) | `$filter=` leer: `wert` (`SPEC-162` Nr. 8 mit `SPEC-173` Nr. 4); ohne Leser: Kompatibilität `filter` (`SPEC-160` Nr. 8) | Grammatik des Ausdrucks: B (`SPEC-173` Nr. 5, L6) | Grammatik: B | Grammatik: B | doppelt: `option_doppelt` (`SPEC-162` Nr. 3 a) | gilt nicht (wie oben) |
| Optionen `$orderby`, `$top`, `$skip`, `$select`, `$inlinecount`, `$format` | leer: `wert` (`SPEC-162` Nr. 4 bis 8 mit `SPEC-173` Nr. 4); `$format` leer: `wert` (L2) | `$expand`, `$skiptoken`, `$callback`, `$id` und die V4-Optionen: Kompatibilität `option` (L3); `$format=xml`, `atom`: `format` (L2); `$orderby` mit `(` oder `/`, `$select` mit `/`: `option` (L1, L5) | `$count`, `$search`, `$apply`: `option` (`SPEC-173` Nr. 4); `application/json;odata=verbose`: lesbar (L2) | `$top=abc`, `$inlinecount=foo`: `wert` (`SPEC-173` Nr. 4, L9) | doppelt: `option_doppelt`; 256, 257 und 32, 33 Elemente: `SPEC-162` Nr. 6, 7; Grenzen `SPEC-160` Nr. 6 | gilt nicht (wie oben) |
| Kompatibilitätsfehler | gilt nicht: er hat immer eine Ursache (`SPEC-174` Nr. 2) | die Menge: `SPEC-170` Nr. 3, Ursachen `SPEC-174` Nr. 2; `methode`: Mock (`SPEC-174` Nr. 3, `slice-v2-bootstrap`) | V4-Syntax am V2-Parser: `SPEC-173` Nr. 4 und 5 | `literal_typ` (`SPEC-174` Nr. 2, `SPEC-173` Nr. 3) | Rangfolge `SPEC-174` Nr. 4 und `SPEC-173` Nr. 6 | gilt nicht (wie oben) |
| Einhängestelle `WithFilter` | kein Leser: `filter` (`SPEC-160` Nr. 8) | Leser liefert `*KompatFehler` oder `*URLFehler`: Ursache gilt, Ort `query/<n>` (L7) | Fehler anderer Art oder Klasse: Fehler des Aufrufers `model.ErrUngueltig` (L7) | gilt nicht: der Leser liest den Text (B) | höchstens ein Aufruf je `ParseURL` (`SPEC-160` Nr. 8 a) | gilt nicht: der Leser hat keinen Kontext (`SPEC-160` Nr. 8) |

**Zusatztabelle** (je Paar von Verstößen an derselben URL und je Modus; `AGENTS.md` §3.16; entschieden):

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| Kompatibilität und `400` an derselben URL | `ODR-E6005` geht vor, gleich wo (`SPEC-173` Nr. 1, 6) | Zeile 7 (41 Paare) |
| zwei Eigenschaften außerhalb des Umfangs | die erste in der Lesefolge (`SPEC-173` Nr. 1, 6) | Zeile 8 (21 Paare) |
| zwei Verstöße der Klasse `ungueltig` | der erste in der Lesefolge (`SPEC-160` Nr. 5) | Zeile 9 (19 Paare) |
| Fehler des Lesers gegen spätere Option | `SPEC-160` Nr. 5 mit `SPEC-173` Nr. 6 | Zeile 21 |
| Kopfzeilen der Anfrage gegen Version des Service | die Kopfzeilen ändern nichts (`SPEC-174` Nr. 6, `SPEC-173` Nr. 7) | Signatur: Zeile 22; Verhalten: Rundlauf in `slice-v2-bootstrap` (Sendung) |
| `$format` gültig gegen außerhalb | `SPEC-173` Nr. 4 mit L2 | Zeile 44 |
| mit Leser gegen ohne Leser (Modus der Einhängestelle) | `SPEC-160` Nr. 8 | Zeilen 19, 20, 45 |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30), beim Start nachgezogen (2026-11-01, Zähler = Zahl der Dateien in `evidence/`). `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, Schwelle erreicht, Verkörperung offen) gilt für die Fakes der Ports; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (3×, Schwelle erreicht, Zustand offen) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist in zwei Zeilen nebenläufig (Zeilen 17 und 18, gleichzeitige Aufrufe von `ParseURL` und des eingehängten Lesers). Beide sind Feldlagen, die der Plan als Risiko führt (§3 Betriebsregeln, §6); ein weiteres Auftreten trägt der Slice in `evidence/` ein.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
