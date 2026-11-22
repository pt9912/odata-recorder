# Slice slice-v2-rundlauf: Rundlauf gegen einen Referenzdienst V2, Referenzdienst V2 als Test-Server und Kapitel zu V2 im Handbuch (Teil B der Schlussverdrahtung von OData V2)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-FA-12`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-170` Nr. 2, 3, 4, 6 · `SPEC-171` Nr. 2, 4 · `SPEC-172` Nr. 3, 5, 6, 7 · `SPEC-173` Nr. 3, 7 · `SPEC-174` Nr. 1 bis 7 · `SPEC-175` Nr. 4 bis 7 · `SPEC-127` Nr. 2, 5, 8, 12, 14 · `SPEC-195` Nr. 2, 3, 7 · `SPEC-192` Nr. 3, 7 · `SPEC-198` Nr. 4 · `SPEC-168` Nr. 5 · `SPEC-125` Nr. 14 · `SPEC-200` Nr. 1 und 6 d (das Kapitel zu V2 und sein Beispiel) · `SPEC-128` Nr. 3 m (das Handbuch zählt nicht) · `LH-FA-17.a` (versionierter Adapter, keine Vermischung von V2 und V4) · `LH-FA-12.a` (Betriebsarten) · `ARC-005`, `ARC-012` · `architecture.md` §2.5 (Modus-Routing), §7 (Erweiterungspunkt OData V2).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Plan aus dem Abschnitt *Zuschnitt Teil B* von [slice-v2-bootstrap](../in-progress/slice-v2-bootstrap.md) angelegt, Vertrag und Zuschnitt vom Architect am 2026-11-22). **Datum:** 2026-11-22.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Zuschnitt (Architect, 2026-11-22, vor dem Code, `AGENTS.md` §3.14):** Die Tabelle *Zusage · Fänger* des ganzen Gegenstands ergab *S* = 4125 (Teil A 2120, Teil B 2005, beide aus der Tabelle nach `SPEC-128` Nr. 1 gerechnet); *S* ≥ 2223 ist die Schnittfrage, und die zwei Teile teilen keinen Code: Teil A ([slice-v2-bootstrap](../in-progress/slice-v2-bootstrap.md)) ist Produktcode im Composition Root mit Tests am Root, **dieser Slice ist Teil B**: ein Test mit Prozessen, ein Referenzdienst als Test-Server und das Kapitel zu V2; er ändert keine Zeile Produktcode. Er startet, wenn Teil A in `done/` liegt (§4).

**Ziel:** Der Rundlauf der Welle (Closure-Trigger von [welle-odata-v2](../welle-odata-v2.md)): ein Test zeichnet einen Referenzdienst V2 auf, importiert die Session, legt einen Workspace an und beantwortet im Modus `mock` die Metadaten und eine Standardabfrage nach dem dokumentierten V2-Umfang; ein leeres Ergebnis bleibt im V2-Format, eine nicht unterstützte Eigenschaft ergibt den Kompatibilitätsfehler `ODR-E6006`, ein V4-Dienst verhält sich wie vorher, und das Handbuch führt V2 als Kapitel. Die Zusagen stehen als Zeilen B01 bis B20 in §3; der Vertrag steht in `SPEC-170` bis `SPEC-175` (vom Architect, auf `main`).

**Übernimmt (Rundlauf, Referenzdienst und Kapitel):** der Rundlauf Aufzeichnen, Importieren, Workspace, Abfragen im Modus `mock` gegen einen Referenzdienst V2, der Referenzdienst V2 als Test-Server (Auftrag 6 des Architect an [slice-v2-bootstrap](../in-progress/slice-v2-bootstrap.md), dort entschieden) und das Kapitel zu V2 im Handbuch. Der Gegenstand stand im ungeschnittenen Plan von [slice-v2-bootstrap](../in-progress/slice-v2-bootstrap.md); der Architect hat ihn dort vor dem Code als Teil B abgetrennt (Geber: `slice-v2-bootstrap`, Teil A).

**Übernimmt (Sendungen, `AGENTS.md` §3.13; Eintrag des Planners vom 2026-11-22 beim Anlegen, die Umadressierung von Teil A auf diesen Teil):** die Pläne in `done/` nennen als Nehmer `slice-v2-bootstrap`; sie sind Zeitdokumente und bleiben stehen. Was zum Rundlauf, zum Referenzdienst oder zum Kapitel gehört, geht an diesen Slice, Verdrahtung, Konfiguration, Anlage und Kataloge bleiben bei Teil A (dessen §1). Dieser Plan ist der Nehmer; sein §1 *Ausdrücklich NICHT* und seine DoD sind vorher geschrieben und gelesen, kein Ausschluss trifft eine Sendung.

1. **Geber [slice-v2-query](../done/slice-v2-query.md) und [slice-v2-query-filter](../done/slice-v2-query-filter.md):** der Rundlauf-Fänger „Versions-Kopfzeilen ändern nichts“ (`SPEC-174` Nr. 6, `SPEC-173` Nr. 7; B12; im Plan des Gebers Zeile 22 und Sendung (c)); der Test über beide Seiten, Parser plus Interpreter (B06, B14); die Kopie der Grammatik gegen V4, gefangen von B06 und B14 gegen Literale in der Form von `SPEC-173` Nr. 3 (im Plan von `slice-v2-query-filter` §6 *Die Kopie der Grammatik … laufen auseinander*). Der Handbuch-Eintrag von `ODR-E6006` ist nicht hier: er steht in Teil A (A22).
2. **Geber [slice-v2-payload-lesen](../done/slice-v2-payload-lesen.md) (Sendung S1):** das Schlüsselliteral des V2-Payload-Lesers (`SPEC-172` Nr. 5) gleicht dem Literal, das der V2-URL-Leser (`SPEC-173` Nr. 3) für denselben Schlüssel bildet, Schleife über die sieben Typen (B14).
3. **Geber [slice-v2-payload-schreiben](../done/slice-v2-payload-schreiben.md) und [slice-v2-import-codecs](../done/slice-v2-import-codecs.md), dazu `slice-mock-zuordnung`:** der Darstellungskopf der Antwort des Mock und die Wirkung von `NurLesend` sind am Binary belegt (B04, B09). Der Mock liest beides seit `slice-mock-zuordnung` und `slice-mock-schreiben` (`SPEC-195` Nr. 7, `SPEC-192` Nr. 3, `SPEC-198` Nr. 8); Teil A ändert den Mock nicht und prüft, dass der Satz je Dienst `NurLesend` trägt (A16). Der Architect nennt im Zuschnitt `slice-v2-payload-schreiben` und `slice-mock-zuordnung` als Geber; `slice-v2-import-codecs` führt denselben Fänger der Wirkung als akzeptiertes Negativ „liegt dort“ (sein §3, *Akzeptiertes Negativ*, Nehmer dort `slice-v2-bootstrap`) und ist vom Planner als Geber ergänzt.
4. **Geber [slice-v2-payload-lesen](../done/slice-v2-payload-lesen.md), [slice-v2-metadaten](../done/slice-v2-metadaten.md), [slice-v2-query](../done/slice-v2-query.md) und [slice-v2-payload-schreiben](../done/slice-v2-payload-schreiben.md):** das Kapitel zu V2 im Handbuch mit Aufgabe, Voraussetzung, Konfiguration (`SERVICE_ODATA_VERSION`, `odata_version`), Umfang (Tabelle der Bereiche und ihrer Folge `ODR-E6006`), Betriebsarten (`hybrid` startet nicht), Aufzeichnen mit JSON (`Accept: application/json` oder `$format=json`; ein Atom-Austausch bleibt Replay-fähig und liefert beim Import `kein_json`), Antwortformat und Kennzeichen, `Edm.DateTime` als UTC, `Accept` und die Versions-Kopfzeilen ohne Wirkung, Funktionsimport wie ein unbekanntes Set, Fehlerantworten im Format des Dienstes; ohne Kennung der Planung (`SPEC-200` Nr. 1); die Aussagen belegt der Rundlauf (B20). Die Zeilen von `SERVICE_ODATA_VERSION` und `ODR-E6006` in den Tabellen trägt Teil A (A22).
5. **Das gebundene Beispiel zu V2 im Handbuch** (Start mit `MODE=hybrid` und `SERVICE_ODATA_VERSION=2.0`, Ausgangswert 1, in der Fehlerausgabe die Zeile `ODR-E2001 MODE: Modus hybrid für Version 2.0 nicht unterstützt`; Geber [slice-v2-bootstrap](../in-progress/slice-v2-bootstrap.md), Teil A, im Zuschnitt des Architect) — **bedingt:** das Format von `SPEC-200` kennt heute keine Umgebung (Nr. 6 d); liegt [slice-handbuch-mock](slice-handbuch-mock.md) (liefert die Umgebung im Beispiel-Format, nimmt die Sendung nicht an) beim Start dieses Teils in `done/`, trägt dieser Teil das Beispiel; sonst bleibt das Kapitel Prosa und Tabelle, und das Beispiel ist ein akzeptiertes Negativ mit Grund (Format ohne Umgebung) und Folge (die Aussage zu `hybrid` mit V2 belegt nur ein Test des Roots, A07; keine Wiederkehr zu erwarten, weil jedes Kapitel mit Umgebung dasselbe Negativ trägt, `slice-listener-tls-verdrahtung`). Es entsteht kein Eintrag im Register und kein Folge-Slice. **Stand bei der Anlage** (Planner, 2026-11-22): `slice-handbuch-mock` liegt in `open/`; die Bedingung wird beim Start dieses Teils gegen den Ist-Stand entschieden (§4). Eine Zuweisung an `slice-handbuch-mock` besteht nicht: dessen §1 nennt das Kapitel zu V2 nicht, und sein *Ausdrücklich NICHT* lässt Kapitel zu Verhalten anderer Wellen in der Welle entstehen, die das Verhalten liefert; kein Ausschluss trifft diese Sendung.

**Mock-Ergänzungen:** entfallen, schon geliefert (§1 von Teil A, Beleg dort); der Rundlauf belegt ihre Wirkung am Binary (B04, B09).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Konfiguration (`SERVICE_ODATA_VERSION`, `odata_version`), Anlage des Service mit der Version, Wahl des Satzes im Root, die Zeilen der Kataloge — [slice-v2-bootstrap](../in-progress/slice-v2-bootstrap.md) (Teil A), den dieser Slice in `done/` voraussetzt. Dieser Slice ändert keine Zeile Produktcode (`produktcode: 0` im Block); findet der Rundlauf einen Fehler im Produktcode, behebt ihn dieser Slice nicht (§6, Rückgabe).
- Metadaten-Leser, Payload-Codec, URL-Parser samt Leser von `$filter`, Schreiber und Mock-Kern — `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-v2-payload-schreiben`, `slice-v2-query`, `slice-v2-query-filter` und `welle-mock` (alle in `done/`); der Rundlauf prüft sie am Binary und ändert keinen.
- Aufzeichnen im Modus `record` als Änderung: die Aufzeichnung ist von der Version unabhängig (`welle-odata-v2` §1); B01 belegt das, kein Teil des Record-Pfads ändert sich.
- V2-Mutationen und der Modus `hybrid` für einen V2-Dienst — `welle-odata-v2` §6 bzw. `welle-hybrid`; die Mutationen sind außerhalb des Umfangs und ergeben `ODR-E6006` mit der Ursache `methode` (B09), `hybrid` mit `2.0` verhindert den Start (A07, Teil A; das Beispiel in Sendung 5 nur bedingt).
- Änderungen am V4-Pfad (Mock, Replay, Record) über die Ports hinaus, hinter denen der Adapter steht — `welle-odata-v2` §6; B11 und B17 belegen, dass er sich nicht ändert.
- Der Nachzug der schmaleren Fänger der V2-Slices (`welle-odata-v2` §5, *Offen aus* `slice-v2-payload-lesen`, `slice-v2-query-filter` und `slice-v2-query`) — Adresse bleibt der Planner bei der Closure der Welle; kein Plan nimmt ihn an, und dieser Slice nimmt ihn nicht an: der Rundlauf prüft die Klassen am Binary (B06, B14), er schließt die Prüfgrenzen der Fänger der Leser nicht.
- Eine Erweiterung von `SPEC-200` um eine Umgebung im Beispiel-Format — [slice-handbuch-mock](slice-handbuch-mock.md) (Gate-Kette, Linie C); dieser Slice berührt die Gate-Kette nicht (§3).
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

- [ ] Der Rundlauf in `make test` (`LH-FA-17`): der Referenzdienst V2 (13 Austausche E1 bis E13) und die Tests B01 bis B19 aus §3, je Zeile ein Fänger mit einer roten Mutation, die der Implementer selbst gesehen hat; B17 ist die Suite des Bestands ohne Änderung (die Tests von `welle-mock` laufen unverändert grün). Der Bericht nennt je Zusage Zusage · Mutation · roter Test und die Laufzeit des Rundlaufs gemessen, mit der Last des Hosts vor und nach jedem Lauf (`AGENTS.md` §3.10, Risiko *Laufzeit* in §6). Der Implementer trägt die Tabelle der Mutationen und die Raten (B18) mit Zahl der Prozesse und Messendem in einen Absatz von §3 ein.
- [ ] Das Kapitel zu V2 im Handbuch (B20, Sendungen 4 und 5): Ist-Zustand, ohne Kennung der Planung (`SPEC-200` Nr. 1), die Aussagen belegt der Rundlauf; das gebundene Beispiel nur, wenn die Bedingung der Sendung 5 beim Start gilt; `make handbuch-check` und `make katalog-check` grün.
- [ ] Die Sendungen 1 bis 3 sind getragen: der Fänger der Versions-Kopfzeilen (B12), der Test über beide Seiten und die Kopie der Grammatik (B06, B14), das Schlüsselliteral-Paar (B14), Darstellungskopf und `NurLesend` am Binary (B04, B09); jede mit der Kennung des Gebers im Bericht.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen: hier der zweite Liefer-Punkt (das Kapitel zu V2 ist der Gegenstand dieses Teils; die Zeilen der Tabellen trägt Teil A); das Kataloge-Gate ist grün.
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
| `tests/integration/v2_referenzdienst_test.go` | neu | der Referenzdienst V2: ein `httptest`-Server mit festen Bytes für die 13 Austausche E1 bis E13, ohne Import der Pakete `csdlv2`, `jsoncodecv2`, `urlparserv2`; Port-Fake im Sinne von `SPEC-128` (B19) |
| `tests/integration/v2_rundlauf_test.go` | neu | Aufzeichnen, Import, Coverage, Antworten und Abfragen im Modus `mock`, Kompatibilitätsfehler, Methoden, Kopfzeilen, `Accept`, Schlüsselliteral, Dokument anderer Version, Replay (B01 bis B09, B12 bis B16) |
| `tests/integration/v2_mehrere_test.go` | neu | ein Prozess mit einem V4- und einem V2-Dienst: Vergleich mit den Einzelläufen (B11) und gleichzeitige Anfragen (B18) |
| `src/bootstrap/v2_ursache_test.go` | neu | Ursache und Ort des Fehlers und Rangfolge der Paare am Port des Mock im Root: der Haken `NeuerMock` umhüllt `mock.New` (B10); der Wächter um den Port ist der zweite Fake dieses Slice |
| `docs/user/benutzerhandbuch.md`, `docs/user/beispiele/` | update, neu | das Kapitel zu V2 (B20); die Beispiel-Datei zu `hybrid` mit `2.0` nur, wenn die Bedingung der Sendung 5 beim Start gilt; das Handbuch zählt nicht in die Größenmarke (`SPEC-128` Nr. 3 m) |
| `src/` außer `src/bootstrap/v2_ursache_test.go`, `Makefile`, `Dockerfile`, `harness/`, `tools/` | keine Änderung | kein Produktcode (`produktcode: 0`); die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Slice nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` |

Die Zuordnung der Zeilen zu den Dateien ist der Vorschlag des Planners nach den vier Pfaden des Blocks (§4); B17 ist die Suite des Bestands ohne Änderung, B19 prüft die Testdateien von Teil B und liegt bei dem Referenzdienst, B20 ist das Handbuch. Eine weitere neue Datei nennt der Implementer dem Planner vor dem Anlegen (§6, Ausgang *offen*), damit der Block sie führt.

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation (`SPEC-170` bis `SPEC-175`) **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind vom Architect entschieden (Zuschnitt vom 2026-11-22), nicht vom Implementer. Die Zeilen B01 bis B20 und die Mengen sind die des Architect im Wortlaut; Teil A trägt A01 bis A22 ([slice-v2-bootstrap](../in-progress/slice-v2-bootstrap.md)).

**Der Referenzdienst V2** (Test-Server, Auftrag 6 des Architect an Teil A; Port-Fake im Sinne von `SPEC-128`; eine Datei `tests/integration/v2_referenzdienst_test.go`). Er ist ein `httptest`-Server mit festen Bytes und **importiert keines der Pakete `csdlv2`, `jsoncodecv2`, `urlparserv2`**: das Orakel ist nicht der Code (B19; `AGENTS.md` §3.15 „Eine Fixture gleicht nie dem Wert der Mutation“). Der Dienst liegt unter `/odata`; er zählt Verbindungen und Requests wie der Referenzdienst von V4 und liefert JSON nur, wenn `Accept` `application/json` nennt oder `$format=json` in der Query steht, sonst Atom.

| Nr. | Anfrage an den Referenzdienst (Aufnahme-Folge) | Antwort | Ausgang beim Import (`SPEC-127` Nr. 8) |
|---|---|---|---|
| E1 | `GET /odata/$metadata` | EDMX 1.0 (CSDL 2008/09), `Content-Type: application/xml;charset=utf-8`: Entity-Typ `Product` mit `Id` (Int32, Schlüssel), `Name` (String, nullbar), `Active` (Boolean), `Qty` (Int64), `Price` (Decimal), `Created` (DateTime), `Ref` (Guid) und `Weight` (**`Edm.Double`**, nicht unterstützt), Sets `Products` und `Archive` (beide `Product`), dazu sieben Typen `KString`, `KBoolean`, `KInt32`, `KInt64`, `KDecimal`, `KGuid`, `KDateTime` mit je einem Schlüssel des Typs im Namen und `Name`, je ein Set gleichen Namens | gelesen (Metadaten) |
| E2 | `GET /odata/Products`, `Accept: application/json` | `{"d":{"results":[<drei Entities mit __metadata, alle Properties inklusive Weight>]}}`, `application/json;charset=utf-8`, `DataServiceVersion: 2.0` | gelesen (Collection) |
| E3 | `GET /odata/Products(1)`, `Accept: application/json` | `{"d":{…}}` | gelesen (Entity) |
| E4 | `GET /odata/Products` ohne `Accept` | Atom-Feed, `application/atom+xml;charset=utf-8` | übersprungen: `kein_json` |
| E5 | `GET /odata/Products?$inlinecount=allpages`, `Accept: application/json` | `{"d":{"__count":"3","results":[…]}}` | übersprungen: `query` |
| E6 | `GET /odata/Archive`, `Accept: application/json` | `{"d":{"results":[]}}` | gelesen (Collection, leer) |
| E7 bis E13 | je `GET /odata/K<Typ>(<Literal der URL>)`, `Accept: application/json` | `{"d":{…}}` mit dem Schlüssel des Typs | gelesen (Entity) ×7 |

Zählprüfung gegen die Quelle: 13 Austausche; gelesen 11 (Metadaten 1, Collection 2, Entity 8), übersprungen 2 (`kein_json` 1, `query` 1), Summe 13 (`SPEC-127` Nr. 14). Die drei Entities von `Products` tragen die sieben Typen in der Schreibweise von V2 (`Edm.Int64` und `Edm.Decimal` als Zeichenkette, `Edm.DateTime` als `\/Date(<ms>)\/`, eine nullbare Property mit `null` in einer Entity); die URL-Literale der Schlüssel (E7 bis E13) sind je Typ mindestens eines der Formen aus `SPEC-173` Nr. 3 (Menge 11 Formen: String `'a''b'`, Boolean `true`, Int32 `1`, Int64 `7` und `7L`, Decimal `1.5` und `1.5M`, Guid `guid'<36 Zeichen>'`, DateTime `datetime'2020-01-02T03:04:05'`, dazu `l`- und `m`-Suffix klein: `7l`, `1.5m`; Zählprüfung 11).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15, Teil B). Mengen als Generatoren aus dem Wortlaut der Zusage, nie eine Liste der Fälle, die einfielen; **Ursachen** = {`format`, `methode`, `pfad`, `option`, `filter`, `literal_typ`} (6, `SPEC-174` Nr. 2); **Antwortarten** = {`$metadata`, Collection, Entity, Anzahl (`/$count`), Fehler `ODR-E6006`} (5).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| B01 | Aufzeichnen gegen den Referenzdienst V2 schreibt jeden der 13 Austausche; die Bytes von Request und Response in den Blobs gleichen den Konstanten der Fixture, unabhängig von der Version (`SPEC-170` Nr. 6, Modus `record`) | `TestV2Aufzeichnung`: Schleife über E1 bis E13, Zählprüfung 13; Bytes aus den Konstanten | Root lässt den V2-Dienst aus `record.Dienste` (13 rot, `ODR-E5009`); Aufnahme verwirft Antworten mit `application/json;charset=utf-8` (11 rot: E2, E3, E5, E6, E7 bis E13) | andere |
| B02 | Beenden und Import: `Neu`; der Report zählt 11 gelesen (Metadaten 1, Collection 2, Entity 8) und 2 übersprungen (`kein_json` 1, `query` 1), Summe 13 (`SPEC-127` Nr. 8 und 14, `SPEC-175` Nr. 4 b und c) | `TestV2Import`: Zahlen aus der Fixture (Zählprüfung 13, 11, 2) | Import mit dem Satz `4.0` (`ODR-E6004`, alle rot); `kein_json` als gelesen gezählt (Fehler des Codecs); `$inlinecount` gelesen | andere |
| B03 | Coverage nach dem Import: `Products` ist `complete` (ungefilterte Abfrage ohne Folgelink), `Archive` ist `complete` (leere Kette), jeder der sieben Sätze `K<Typ>` ist `partial` (nur Einzelabfrage), je Set aus `SPEC-127` Nr. 12 abgeleitet und gegen den Report gelesen, nicht gegen den Code | `TestV2Coverage`: Schleife über 9 Sets, Zählprüfung 9 | Kette nach `__next` statt `@odata.nextLink` gesucht; Status für ein leeres Set `partial` | andere |
| B04 | Der Mock beantwortet im Modus `mock` nach Workspace `$metadata`, Collection, Entity und Anzahl mit Body, `Content-Type` und Kennzeichen der Version; Body, Typ und Kopfzeile gleichen den Konstanten (`SPEC-172` Nr. 6, `SPEC-198` Nr. 4, `SPEC-175` Nr. 7): `application/json;charset=utf-8` mit `DataServiceVersion: 2.0`, bei der Anzahl `text/plain;charset=utf-8` mit denselben Ziffern, bei `$metadata` das aufgezeichnete Dokument. Menge 12 Zellen: 4 Arten × {Body, `Content-Type`, Kennzeichen} | `TestV2Antwortarten`: Tabelle 4 × 3, Zählprüfung 4 und 3; Orakel Konstanten | Kennzeichen `OData-Version: 4.0` (4 rot); Typ `application/json;odata.metadata=minimal` (3 rot); Body in der Form von V4 (3 rot) | andere |
| B05 | Die sieben Typen stehen in Collection und Entity in der Schreibweise von V2; die nicht unterstützte Eigenschaft `Weight` erscheint nie (`SPEC-170` Nr. 4, `SPEC-172` Nr. 3 und 6 f). Menge 14: 7 Typen × {Wert, `null` bei nullbarer Property}, dazu `Weight` in beiden Antwortarten | `TestV2Typen`: Schleife über 7 und 2, Zählprüfung 14; Orakel Konstanten | `Edm.Int64` als Zahl (2 rot); `Edm.DateTime` ohne Hülle `Date` (2 rot); `Weight` geschrieben (2 rot) | andere |
| B06 | Standardabfrage nach dem Umfang: `$filter`, `$orderby`, `$top`, `$skip`, `$select`, `$inlinecount` (`allpages` mit `__count` als Zeichenkette, `none`) ergeben die erwartete Menge in der erwarteten Reihenfolge; `$filter` mit `eq`, `ne`, `gt`, `ge`, `lt`, `le`, `and`, `or`, `not`, Klammern (`SPEC-170` Nr. 2). Menge 16 Anfragen: 6 Optionen und 10 Filterformen | `TestV2Abfragen`: Schleife, Zählprüfung 6 und 10; Orakel aus der Fixture | `$skip` ignoriert (1 rot); `__count` als Zahl (1 rot); `lt` und `le` vertauscht (2 rot); `not` ohne Wirkung (1 rot) | andere |
| B07 | Ein leeres Ergebnis bleibt im V2-Format: `{"d":{"results":[]}}`, mit `$inlinecount=allpages` `{"d":{"__count":"<n>","results":[]}}` mit der Gesamtzahl ohne `$top` und `$skip`, nie `null`, nie ein fehlendes `results` (`SPEC-172` Nr. 6 c und l). Menge 8: Wege {Filter ohne Treffer (n = 0), `$skip` über das Ende (n = 3), `$top=0` (n = 3), leeres Set `Archive` (n = 0)} × {ohne, mit `$inlinecount`} | `TestV2Leer`: Schleife, Zählprüfung 4 und 2 | `results` fehlt (8 rot); `__count` zählt die Entities der Seite (2 rot: `$skip`, `$top=0`); `null` statt `[]` (8 rot) | andere |
| B08 | Eine Anfrage außerhalb des Umfangs ist `501` mit `ODR-E6006`, ohne einen Wert der Anfrage im Body, mit `X-Correlation-Id`; der Zustand des Workspace (Entities, Revisionen, Coverage) bleibt (`SPEC-174` Nr. 1 und 2, `SPEC-170` Nr. 3). Menge 12: 6 Ursachen × 2 Repräsentanten (`$format=atom`, `$format=xml`; `POST`, `MERGE`; `/` und `/Products(1)/Category`; `$expand=Category`, `$skiptoken=x`; `substringof('a',Name)`, `Price add 1 gt 2`; `Price gt 1.5f`, `Id eq X'AA'`) | `TestV2Kompat`: Schleife, Zählprüfung 6 und 2; Status, Code, Body, Zustand vor und nach | V2-Dienst mit dem Satz `4.0` (`ODR-E6001`/`ODR-E6005`, alle rot); Antwort `404` statt `501` (12 rot); Zustand durch `POST` geändert (2 rot) | andere |
| B09 | Jede Methode außer `GET` und `HEAD` ist an einem V2-Dienst `ODR-E6006` mit der Ursache `methode`, der Body wird nicht gelesen, nichts wird geschrieben; `HEAD` ist `GET` ohne Body (`SPEC-174` Nr. 3, `SPEC-195` Nr. 7, `SPEC-192` Nr. 3). Menge 24: {`POST`, `PUT`, `PATCH`, `MERGE`, `DELETE`, `OPTIONS`} × {Collection, Entity, `/$count`, `$metadata`}; dazu `HEAD` an den vier | `TestV2Methoden`: Schleife, Zählprüfung 6, 4 und 4; Status, Code, Zustand | `NurLesend` falsch im Satz (`ODR-E6002` oder eine Ausführung, 24 rot); `HEAD` mit Body (4 rot) | andere |
| B10 | Ursache und Ort des Fehlers je Fall und die Rangfolge der Paare am Port des Mock im Root: `errors.As` auf `*model.KompatFehler`; Menge 12 Fälle aus B08 und 4 Paare: `ODR-E5003` vor Kompatibilitätsfehler, `ODR-E6005` vor Kompatibilitätsfehler, Kompatibilitätsfehler des URL-Lesers vor `methode`, `methode` vor `ODR-E7001` (`SPEC-174` Nr. 4, `SPEC-195` Nr. 3) | `TestV2Ursache` in `src/bootstrap`: der Haken `NeuerMock` umhüllt `mock.New`, hält jeden Fehler fest (Port-Fake, zählt jeden Aufruf), Schleife, Zählprüfung 12 und 4 | `option` und `filter` vertauscht (4 rot); `methode` hinter dem Lesen des Workspace (1 Paar rot); `ODR-E6005` hinter `ODR-E6006` (1 Paar rot) | andere |
| B11 | Der V4-Pfad ist byte-gleich: ein V4-Dienst neben einem V2-Dienst im selben Prozess antwortet wie ein V4-Dienst allein, ein V2-Dienst wie ein V2-Dienst allein, mit der Kopfzeile seiner Version (`OData-Version: 4.0` bzw. `DataServiceVersion: 2.0`); kein Service beeinflusst den anderen (`SPEC-175` Nr. 6, `LH-FA-17.a`). Menge 12: 6 Antwortarten (`$metadata`, Collection, Entity, Anzahl, `$filter`, Fehler `ODR-E6008`) × 2 Dienste | `TestV2NebenV4`: ein Prozess mit `SERVICES_FILE` und zwei Upstreams, Import beider, zwei Vergleichsläufe mit je einem Dienst; Zählprüfung 6 und 2 | ein Satz für beide Dienste, der letzte gewinnt (12 rot); `ETAG_ENABLED` für den V2-Dienst wirksam (1 rot); Metadaten-Cache je Prozess | andere |
| B12 | Die Versions-Kopfzeilen der Anfrage ändern nichts: dieselbe Antwort (Status, Header außer `X-Correlation-Id`, Body) mit und ohne Kopfzeile (`SPEC-174` Nr. 6, `SPEC-173` Nr. 7). Menge 85: {`DataServiceVersion`, `MaxDataServiceVersion`, `MinDataServiceVersion`, `OData-Version`} × {`2.0`, `4.0`, `1.0`, `3.0`} und alle vier zugleich (17 Formen) × 5 Antwortarten | `TestV2Kopfzeilen`: Schleife, Zählprüfung 4, 4, 17 und 5 | `MaxDataServiceVersion: 1.0` ändert die Form (5 rot); die Antwort trägt die Version der Anfrage (5 rot); je Name eine Mutation | andere |
| B13 | `Accept` ändert nichts: auch `application/atom+xml` ergibt JSON (`SPEC-174` Nr. 5). Menge 15: {fehlt, `application/json`, `application/atom+xml`, `application/xml`, `*/*`} × {Collection, Entity, Anzahl} | `TestV2Accept`: Schleife, Zählprüfung 5 und 3 | Atom bei `Accept: application/atom+xml` (3 rot) | andere |
| B14 | Das Schlüsselliteral des V2-Payload-Lesers gleicht dem des V2-URL-Lesers: der Mock findet jede importierte Entity über `GET /<Set>(<Literal>)` für jeden der sieben Typen (`SPEC-172` Nr. 5, `SPEC-173` Nr. 3). Menge 11 Formen der Literale über 7 Typen | `TestV2SchluesselLiteral`: Schleife über die 11 Formen, Zählprüfung 7 und 11; Status 200 und die Entity | Leser schreibt die Guid in Großbuchstaben (`404`, die Form `guid'…'` rot); Zeitpunkt ohne Normalisierung auf UTC (die Form `datetime'…'` rot); Suffix `L` nicht gelesen (`7L` rot) | andere |
| B15 | Ein Dokument anderer Version als der deklarierten ist `ODR-E6004`, nichts ist geschrieben: ein V2-Dokument unter einem Service `4.0` und ein V4-Dokument unter einem Service `2.0` (`SPEC-175` Nr. 4 d, `SPEC-171` Nr. 2). Menge 2 Richtungen | `TestV2Kreuzdokument`: Schleife über 2; Zustand vor und nach (kein Import, keine Baseline, `metadata_blob` leer) | Prüfung der Version am Leser entfernt (2 rot); der Importer nimmt den Satz eines anderen Dienstes | andere |
| B16 | Replay: ein V2-Service im Modus `replay` beantwortet jeden der 13 Austausche byte-gleich, auch den Atom-Austausch (`SPEC-170` Nr. 6, `SPEC-175` Nr. 5). Menge 13 | `TestV2Replay`: Schleife über E1 bis E13, Zählprüfung 13 | Root verdrahtet `replay` für `2.0` nicht (13 rot); Atom wird verworfen (1 rot) | andere |
| B17 | Eine Antwort des Mock für einen Dienst `4.0` bleibt, wie sie war: die Tests von `welle-mock` laufen unverändert grün (`LH-FA-17.a`, Risiko des Plans) | die Suite des Bestands ohne Änderung, `make test` | Satz `2.0` für einen V4-Dienst (die Suite rot) | andere |
| B18 | Gleichzeitige Anfragen an einen V2- und einen V4-Dienst desselben Prozesses beeinflussen sich nicht: jede Antwort gleicht der Einzelantwort, kein Datenrennen (`SPEC-175` Nr. 6; die Implementierungen haben keinen Zustand, `SPEC-125` Nr. 14). Menge: 16 Goroutinen (je 8 je Dienst) × 5 Antwortarten, Start an einer Schranke | `TestV2Gleichzeitig`: Wirkungs-Fänger (jede Antwort gegen die Einzelantwort, Zählprüfung 16 und 5) und `make test-race` je Prozess (`-test.count=1`) | der Root hält den zuletzt gewählten Satz in einem Feld des Handle und nutzt ihn je Anfrage (Wirkung rot in jedem Lauf; der Detektor meldet das Rennen mit einer Rate, die der Bericht je Prozesslauf nennt); Tabelle der Sätze beim ersten Zugriff gefüllt | nebenläufig |
| B19 | Das Orakel ist nicht der Code: die Dateien des Referenzdienstes und der Erwartungen importieren keines der Pakete `csdlv2`, `jsoncodecv2`, `urlparserv2` (`AGENTS.md` §3.15, Fixture gleicht nie dem Wert der Mutation) | `TestV2OrakelOhneCode`: AST über die Testdateien von Teil B | ein Test baut eine Erwartung mit `jsoncodecv2.NewSchreiber()` (rot) | andere |
| B20 | Das Handbuch führt V2 als Kapitel (Inhalt oben, Sendung 4): Ist-Zustand, ohne Kennung der Planung, die Aussagen belegt der Rundlauf; das Beispiel nach Sendung 5, wenn die Bedingung gilt; die Kataloge bleiben grün | `make handbuch-check` und `make katalog-check`, Ausgabe `ok` (benannte Ausgabe) | Kapitel nennt `SPEC-175` (Befund der Menge von Kennungen); Aussage über Geplantes (Befund) | andere |

Der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein (`AGENTS.md` §3.10); die Laufzeit des Rundlaufs steht dort gemessen, mit der Last des Hosts vor und nach jedem Lauf. Die Zeile B18 ist die nebenläufige: jede Wirkung wird in jedem Lauf rot, ein Datenrennen zählt der Detektor von `make test-race` je Prozess (`-test.count=1`).

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (hier: die Bytes des Referenzdienstes, etwa `\/Date(<ms>)\/`, und jede Folge aus Backslash und Hex-Ziffern) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen; der Bericht nennt je Menge die Zählprüfung.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `next` → `in-progress`): [slice-v2-bootstrap](../in-progress/slice-v2-bootstrap.md) (Teil A) liegt in `done/`, dazu die Trigger des Gebers: [slice-v2-import-codecs](../done/slice-v2-import-codecs.md), [slice-v2-metadaten](../done/slice-v2-metadaten.md), [slice-v2-payload-lesen](../done/slice-v2-payload-lesen.md), [slice-v2-payload-schreiben](../done/slice-v2-payload-schreiben.md), [slice-v2-query](../done/slice-v2-query.md), [slice-v2-query-filter](../done/slice-v2-query-filter.md), [welle-mock](../done/welle-mock/welle-mock.md), [slice-bootstrap-m3](../done/welle-mock/slice-bootstrap-m3.md) und [slice-handbuch-meldungskatalog](../done/slice-handbuch-meldungskatalog.md) liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-170` bis `SPEC-175`), die Tabelle *Zusage · Fänger* und das Raster stehen in §3 und §6 (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** Teil A (ohne die Wahl des Satzes je Version im Root bliebe jeder Dienst ein V4-Dienst, und der Rundlauf liefe nicht); die V2-Slices davor (die Adapter und der Import, die der Root verdrahtet); `welle-mock` und `slice-bootstrap-m3` (der Modus `mock` im Root, Workspace und Mock-Kern, `Dienst.Etag`); `slice-handbuch-meldungskatalog` (das Kataloge-Gate, das B20 hält). Der Referenzdienst V4 und der Rundlauf von V4 (`slice-m2-referenzdienst`, `slice-bootstrap-m3`) liegen in `done/` und sind das Muster.
- **Bedingung des Beispiels, keine Start-Bedingung:** [slice-handbuch-mock](slice-handbuch-mock.md) in `done/` entscheidet, ob dieser Teil das gebundene Beispiel trägt (Sendung 5); der Planner prüft den Stand beim Start und trägt ihn in §1 ein.
- **Was der Slice nicht braucht:** `welle-etag`, `welle-expand-batch`, `welle-hybrid`, `welle-betrieb`, `welle-admin-api`; die Reihenfolge nach den Muss-Wellen ist Priorisierung.

**Stand der Bedingungen bei der Anlage** (Planner, 2026-11-22, `main` = `9e9819c5`): Teil A liegt in `open/` und startet am selben Tag; die V2-Slices, `welle-mock`, `slice-bootstrap-m3` und `slice-handbuch-meldungskatalog` liegen in `done/` (geprüft an den Dateien); `slice-handbuch-mock` liegt in `open/`. Den Stand beim Start nennt der Planner im Commit des Starts.

**Reihenfolge und Konflikte** (Planner, 2026-11-22): Dieser Teil legt in `src/bootstrap/` eine neue Testdatei an und ändert sonst keine Datei dort; mit `tests/integration/` teilt er das Verzeichnis, aber keine Datei (der Block nennt die Dateien einzeln, `SPEC-128` Nr. 10 g). `docs/user/benutzerhandbuch.md` ändern auch die Handbuch-Slices (`slice-handbuch-mock`, `slice-handbuch-record-replay`, `slice-handbuch-telemetrie`, `slice-handbuch-etag-zeile`, `slice-handbuch-abnahme`); welcher zuerst nach `done/` geht, ist Priorisierung, und der später gemergte nimmt `main` auf (Nr. 5 der Übergangsregeln). Die Gate-Kette (Nr. 7) berührt der Slice nicht.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-22): die Tabelle in §3 trägt 20 Zeilen (B01 bis B20), *a* = 1 (B18, nebenläufig), *b* = 19 (B17 und B20 sind Zeilen mit Fänger aus dem Bestand bzw. einer Ausgabe des Gates und zählen als `andere`), *f* = 2 (der Referenzdienst V2 und der Wächter um den Mock-Port in B10), *p* = 0 (kein Produktcode): *S* = 85·1 + 80·19 + 200·2 + 0 = 2005; 135·2005 = 270675, *M* = 100 · ⌈27,0675⌉ = 2800; keine Schnittfrage (135·*S* = 270675 ≤ 300000, *S* < 2223, 218 darunter). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; das Handbuch zählt nicht (`SPEC-128` Nr. 3 m). Die Pfade nennen Dateien einzeln, kein Verzeichnis: jede weitere Datei meldet der Implementer dem Planner vor dem Anlegen (§6). `basis`, `schaetzung` und `marke` setzt der Planner beim Start dieses Teils vor dem ersten Code-Commit (`AGENTS.md` §3.14, `SPEC-128` Nr. 3 k); bis dahin steht die Marke aus. Die Fänger sind überwiegend Go-Tests mit Prozessen und einem Test-Server (Geltungsgrenze von `SPEC-128` Nr. 2 erfüllt: der Slice führt einen Block).

```groessenmarke
pfad: tests/integration/v2_referenzdienst_test.go
pfad: tests/integration/v2_rundlauf_test.go
pfad: tests/integration/v2_mehrere_test.go
pfad: src/bootstrap/v2_ursache_test.go
produktcode: 0
fakes: 2
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nur vor dem Code und nur über der Schnittgrenze (`SPEC-128` Nr. 1); die Schätzung liegt darunter, ein weiterer Schnitt ist nicht vorgesehen: Referenzdienst, Rundlauf und Kapitel teilen die Bytes der Austausche E1 bis E13, auf die B20 die Aussagen des Kapitels stützt. Eine Messung über der Marke nach dem Code führt nicht zurück (`AGENTS.md` §3.14): der Slice schließt mit der Messung und dem Grund im Eintrag `ueberschreitung`.
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`, oder [slice-v2-bootstrap](../in-progress/slice-v2-bootstrap.md) liegt nicht in `done/`, oder der Rundlauf findet einen Fehler im Produktcode, den kein Slice annimmt (§6, Rückgabe).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor. Der Rundlauf ist ein Closure-Kriterium von [welle-odata-v2](../welle-odata-v2.md) (§3); mit diesem Slice in `done/` ist es erfüllt, die Closure der Welle bleibt Sache von `close-welle`.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Tests von `welle-mock` müssen unverändert grün laufen (`welle-odata-v2` §3); ein V4-Dienst neben einem V2-Dienst im selben Prozess antwortet wie allein — **Ausgang:** offen; Fänger B17 (Suite des Bestands, `make test`) und B11 (byte-gleich neben V2); die 37 Stellen, die einen `Dienst` ohne Version bilden, fängt A11 in Teil A.
- Mehrere Services verschiedener Versionen im selben Prozess dürfen sich nicht beeinflussen (Roots, Metadaten, Caches, Sätze) — **Ausgang:** offen; die Wahl je Dienst fängt Teil A (A13, A16, A19, A20, A21), im Prozess hier B11 und B18.
- Der Rundlauf ist ein Test mit Prozessen, Fristen und einem Referenzdienst; er erbt die Zeitgrenzen von `make test` und `make gegenproben` — **Ausgang:** weiter offen → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` im Register (7×, Stand: geplant in `slice-coverage-gate`); der Bericht nennt die Laufzeit des Rundlaufs gemessen, mit der Last des Hosts vor und nach jedem Lauf (`AGENTS.md` §3.10).
- Der Referenzdienst V2 und der Wächter um den Mock-Port sind Fakes: sie halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 4×, verkörpert) — **Ausgang:** offen; Fänger B10 und B19.
- Das Orakel ist nicht der Code: eine Fixture gleicht nie dem Wert der Mutation (`BEO-REPO/fixture-gleicht-dem-wert-der-mutation`, 9×, verkörpert) — **Ausgang:** offen; Fänger B19 (AST über die Testdateien von Teil B).
- Die Zeile B18 ist nebenläufig: ihre Wirkung muss in jedem Lauf rot werden, das Datenrennen zählt der Detektor je Prozess (`BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen`, 17×, verkörpert; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 9×, verkörpert) — **Ausgang:** offen; Fänger B18, der Bericht nennt Raten mit Last des Hosts und Messendem.
- Das Kapitel zu V2 sagt mehr zu, als ein Test prüft (`BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, 69×, verkörpert in `AGENTS.md` §3.15) — **Ausgang:** offen; Fänger B20 und die Zuordnung jeder Aussage des Kapitels zu einer Zeile des Rundlaufs im Bericht; Aussagen ohne Beleg bleiben aus dem Kapitel (`SPEC-200` Nr. 1).
- Das gebundene Beispiel zu `hybrid` mit `2.0` hängt an `slice-handbuch-mock` (Format ohne Umgebung, `SPEC-200` Nr. 6 d) — **Ausgang:** offen; Adresse: Planner beim Start dieses Teils (Stand von `slice-handbuch-mock` prüfen, Entscheidung in §1 eintragen); liegt der Slice dann nicht in `done/`, gilt als akzeptiertes Negativ: Grund: das Format kennt keine Umgebung; Folge: die Aussage zu `hybrid` mit V2 belegt nur der Test des Roots (A07), das Kapitel steht ohne gebundenes Beispiel zu ihr; Maß: 1 von 11 Punkten des Kapitels nach Sendung 4 (Betriebsarten), keine Wiederkehr zu erwarten, weil jedes Kapitel mit Umgebung dasselbe Negativ trägt (`slice-listener-tls-verdrahtung`); kein Eintrag im Register, kein Folge-Slice.
- Der Rundlauf findet einen Fehler im Produktcode von Teil A oder der V2-Slices, und dieser Slice ändert keinen Produktcode (`produktcode: 0`) — **Ausgang:** offen; Adresse: Planner (Rückgabe nach `AGENTS.md` §3.13, der Implementer hält an und trägt eine Zeile mit Ausgang *offen* und Adresse ein; ein Fehler in Teil A geht an dessen Zweig, solange er nicht in `done/` liegt, sonst in einen eigenen Slice).
- Marke und Block: `basis`, `schaetzung` und `marke` fehlen bei der Anlage — **Ausgang:** offen; Adresse: Planner beim Start dieses Teils (nach `SPEC-128` Nr. 1 nachrechnen: *S* = 2005, *M* = 2800); weitere Dateien nennt der Implementer dem Planner vor dem Anlegen (Zeile mit Ausgang *offen*, Adresse Planner).
- Das Handbuch ist eine Datei, die mehrere Slices schreiben (`docs/user/benutzerhandbuch.md`): Konflikte beim Merge — **Ausgang:** offen; Adresse: Planner (Reihenfolge beim Start, §4).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster steht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, 2026-11-22): alle Zellen gefüllt; offen ist allein, ob der Implementer eine ungenannte Randform findet (Rückgabe an den Architect).
- Der Funktionsimport eines V2-Dienstes im Mock ist ein unbekanntes Set statt `ODR-E6006` (`SPEC-174` Nr. 7) — **Ausgang:** akzeptiertes Negativ, siehe Zusatztabelle unten: Grund: der Referenzdienst führt keinen Funktionsimport; Folge: der Client sieht den Fehler eines unbekannten Sets; Maß: 1 von 6 Ursachenklassen des Rasters.

**Raster *Operation × Argumentart* (Teil B).** Der Dienst `Mock` an einem V2-Dienst; jede Zelle nennt ihren Ort oder „gilt nicht“ mit Grund.

| Operation | Methode außerhalb | Kopfzeile der Version | `Accept` | Ressource, Option, Filter, Literal, Format außerhalb | leeres Ergebnis | Funktionsimport | Abbruch, Wiederholung |
|---|---|---|---|---|---|---|---|
| Anfrage lesen (URL, Methode) | `ODR-E6006` `methode`, nach dem URL-Leser, vor Workspace (`SPEC-174` Nr. 3, `SPEC-195` Nr. 7) | ohne Wirkung (`SPEC-174` Nr. 6) | ohne Wirkung (`SPEC-174` Nr. 5) | `ODR-E6006` mit Ursache (`SPEC-174` Nr. 2, `SPEC-170` Nr. 3) | gilt nicht: die Antwort kennt keinen Fehler | wie ein unbekanntes Set, akzeptiertes Negativ (`SPEC-174` Nr. 7) | Abbruch: `SPEC-195` Nr. 2 (Kontextfehler ohne Code); Wiederholung: die Anfrage hat keinen Zustand (`SPEC-195`, `SPEC-192` Nr. 7 für Mutationen, hier nie ausgeführt) |
| Antwort schreiben (Schreiber des Satzes) | gilt nicht: nichts wird geschrieben | gilt nicht: der Schreiber sieht keine Kopfzeile (`SPEC-172` Nr. 6 g) | gilt nicht (`SPEC-172` Nr. 6 m) | gilt nicht: der Fehler entsteht vorher | `{"d":{"results":[]}}` (`SPEC-172` Nr. 6 c) | gilt nicht: kein Set | Abbruch: `SPEC-168` Nr. 5; Wiederholung: gleiche Bytes (`SPEC-172` Nr. 6) |
| Import, Aufzeichnung lesen | gilt nicht: der Import liest nur `GET` (`SPEC-127` Nr. 8 a) | gilt nicht: der Import liest keine Anfrage-Kopfzeile außer `Accept` im Key | Atom: `kein_json` (`SPEC-127` Nr. 8 g, `SPEC-175` Nr. 4 b) | `$inlinecount`: `query` (`SPEC-175` Nr. 4 c); Property außerhalb: übersprungen mit Grund `typ` (`SPEC-172` Nr. 3) | leere Collection: gelesen (`SPEC-172` Nr. 2 a, `SPEC-127` Nr. 12) | gilt nicht: der Pfad ist kein Set (`SPEC-127` Nr. 8 b) | Abbruch: `SPEC-127` Nr. 2; Wiederholung: `SPEC-127` Nr. 5 |

**Zusatztabelle (Teil B).**

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| V2-Service in `record` | wie V4, von der Version unabhängig (`SPEC-170` Nr. 6) | B01 |
| V2-Service in `replay` | wie V4 (`SPEC-170` Nr. 6) | B16 |
| V2-Service in `mock` | `SPEC-170` bis `SPEC-174`; Methode außer `GET` und `HEAD`: `ODR-E6006` `methode` (`SPEC-174` Nr. 3) | B04 bis B10 |
| V4-Service neben einem V2-Service | unverändert (`SPEC-175` Nr. 6) | B11, B18 |
| Kopfzeilen und `Accept` | ohne Wirkung (`SPEC-174` Nr. 5 und 6) | B12, B13 |
| Funktionsimport eines V2-Dienstes im Mock | unbekanntes Set statt `ODR-E6006` (`SPEC-174` Nr. 7, akzeptiertes Negativ) | benannt, kein Fänger: der Referenzdienst führt keinen Funktionsimport (`csdlv2` meldet ihn als Fund `konstrukt`, `SPEC-171` Nr. 4); Folge: der Client sieht den Fehler eines unbekannten Sets; Maß: 1 von 6 Ursachenklassen des Rasters |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-22); der Planner zieht Treffer und Zähler beim Start nach. Einschlägig (Zähler: die Dateien unter `evidence/`): `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (7×, geplant in `slice-coverage-gate`) für den Rundlauf; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, verkörpert) für die Fakes; `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` (9×, verkörpert) für das Orakel; `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (17×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (9×, beide verkörpert) für B18; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (69×, verkörpert) für das Kapitel; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (21×, verkörpert) für die Betriebsregel; `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (2×, offen) für den Block (Dateien einzeln, keine fremden Zeilen). Kein offener Eintrag erreicht mit diesem Slice 3×; die darüber liegenden sind verkörpert oder geplant.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
