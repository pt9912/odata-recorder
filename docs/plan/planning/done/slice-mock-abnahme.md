# Slice slice-mock-abnahme: Abnahmetests des Mock: Matrix, Determinismus, Isolation, Neustart

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-QA-02`, `LH-QA-03`, `LH-RB-01`, `LH-FA-08`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-QA-02.a` · `LH-QA-03.a` · `LH-RB-01.a` · `SPEC-049` (Abnahmematrix; Nr. 1 bis 10 vom Architect vor dem Code) · `SPEC-048` (Fault Injection, nur Neustart-Teil) · `architecture.md` §6 (Qualitätssicherung).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-mock-abnahme`, Worktree `/Development/odata-recorder-wt-mock-abnahme`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Abnahmetests (Ablage `src/bootstrap/abnahme_*_test.go`, `SPEC-049` Nr. 1) belegen den Mock gegen die Zusagen des Lastenhefts, die kein Slice allein trägt: die Muss-Zeilen der Abnahmematrix (`SPEC-049`, `ODV4-*`) als Golden-Fixtures, 100 identische Läufe nach Reset mit gleichen normalisierten Antworten, 20 parallele Workspaces ohne Beobachtung fremden Zustands und Cursors, und ein regulärer Neustart, nach dem abgeschlossene Mutationen und Snapshots vorhanden sind.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Soll-Zeilen der Matrix (`Prefer`, `If-Match`, `412`, zusätzliche Funktionen) — `welle-etag` und `welle-expand-batch`.
- Die Fault-Injection-Matrix der Persistenz (`SPEC-048`) — Bestand von `welle-persistenz`; hier prüft nur der Neustart-Test die Folge für Mutationen und Snapshots.
- Der Neustart über die Prozessgrenze (Binary, `SIGTERM`, Start im Modus `mock`) für Mutationen — `slice-bootstrap-m3` (Sendung angenommen in dessen §1 und DoD, Geber dieser Slice); für Snapshots bleibt er das akzeptierte Negativ aus `SPEC-049` Nr. 10 b bis `welle-admin-api`.
- Der Lasttest zu `LH-QA-01` — `welle-betrieb`; er braucht die Spans.
- Neue Produktfunktion — der Slice ändert `src/` nur, wo ein Test einen Fehler findet, und meldet ihn dann an den Slice, dem die Zusage gehört.

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

- [x] Abnahmematrix: die Muss-Zeilen von `SPEC-049` (`ODV4-GET`, `-CRUD`, `-Q`, `-T`, `-H`, `-E`) als Golden-Fixtures, je Fixture-ID ein Test (`LH-RB-01`).
- [x] Determinismus: 100 identische Läufe nach Reset ergeben gleiche normalisierte Antworten (`LH-QA-02`), Vergleich ohne `VOLATILE_HEADERS`.
- [x] Isolation und Neustart: mindestens 20 parallele Workspaces beobachten keinen fremden Zustand und keine fremden Cursor; nach regulärem Neustart sind abgeschlossene Mutationen und Snapshots vorhanden (`LH-QA-03`).
- [x] `make gates` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
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
| `src/bootstrap/abnahme_aufbau_test.go`, `abnahme_matrix_test.go`, `abnahme_determinismus_test.go`, `abnahme_isolation_test.go`, `abnahme_neustart_test.go`, `abnahme_nebenlaeufig_test.go` | neu | Aufbau, Golden-Fixtures der Matrix, 100 Läufe, 20 parallele Workspaces, Neustart und Fänger mit Haltepunkt im Paket `bootstrap_test`, damit `make test` und `make test-race` sie fahren (`SPEC-049` Nr. 1; das Architektur-Gate erlaubt dem Paket `tests` keine Kante auf die Dienste, Rückgabe R1) |
| `spec/spezifikation.md` | update | `SPEC-049` Nr. 1 bis 10: Inhalt der Fixtures und Verträge der Läufe, des Neustarts und der Fenster (Architect, vor dem Code erledigt) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster (§6, `AGENTS.md` §3.16) entstanden mit der Festlegung der Fixtures in `SPEC-049` Nr. 4, vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); sie stehen unten und in §6. Die Fixtures hat der Architect festgelegt, nicht der Implementer.

- Die Läufe sind Teil von `make test`: wie lange die 100 Läufe dauern dürfen, ist als Randform in `SPEC-049` Nr. 5 f entschieden (erwartet unter 15 s, Frist 120 s).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-07; Spalten nach `SPEC-128` Nr. 3 a). Die Zusagen sind die Nummern von `SPEC-049` (Nr. 3 bis 8) und die Zeilen von Nr. 4; die Mutationen gelten für `src/` (Mock, Parser, Interpreter, Schreiber, Dienste, Adapter der Ablage) und für die Hilfen des Tests, jede mit einem Wert, der von dem der Fixture verschieden ist. Die Fänger der Matrix sind Teilfälle von `TestAbnahmeMatrix` mit der Fixture-ID als Namen.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| T01 | `ODV4-GET-01`: Collection liefert alle fünf Einträge mit `KK` | `TestAbnahmeMatrix/ODV4-GET-01` | der Schreiber lässt das letzte Element von `value` weg | andere |
| T02 | `ODV4-GET-02`: Lookup liefert Entity 1 und Entity 4 (`null`), `404` für 99 | `TestAbnahmeMatrix/ODV4-GET-02` | ein Mitglied mit `null` entfällt in der Entity; ein Lookup ohne Treffer liefert 200 | andere |
| T03 | `ODV4-GET-03`: leere Collection `value: []`, Anzahl `0`, Lookup `404` | `TestAbnahmeMatrix/ODV4-GET-03` | `value` fehlt bei leerer Collection; die Anzahl `0` ergibt einen leeren Body | andere |
| T04 | `ODV4-Q-01`: `eq` | `TestAbnahmeMatrix/ODV4-Q-01` | `eq` wertet wie `ne` | andere |
| T05 | `ODV4-Q-02`: `ne` ist die Negation von `eq`, mit `null` | `TestAbnahmeMatrix/ODV4-Q-02` | `ne` schließt `null` aus | andere |
| T06 | `ODV4-Q-03`: `gt` ohne Gleichheit | `TestAbnahmeMatrix/ODV4-Q-03` | `gt` wertet wie `ge` | andere |
| T07 | `ODV4-Q-04`: `ge` mit Gleichheit | `TestAbnahmeMatrix/ODV4-Q-04` | `ge` wertet wie `gt` | andere |
| T08 | `ODV4-Q-05`: `lt` ohne Gleichheit | `TestAbnahmeMatrix/ODV4-Q-05` | `lt` wertet wie `le` | andere |
| T09 | `ODV4-Q-06`: `le` mit Gleichheit | `TestAbnahmeMatrix/ODV4-Q-06` | `le` wertet wie `lt` | andere |
| T10 | `ODV4-Q-07`: `and` und `or` | `TestAbnahmeMatrix/ODV4-Q-07` | `and` wertet wie `or`; `or` wertet wie `and` (zwei Läufe) | andere |
| T11 | `ODV4-Q-08`: `not` mit `null` als falsch bei `eq` | `TestAbnahmeMatrix/ODV4-Q-08` | `not` gibt seinen Operanden unverändert zurück | andere |
| T12 | `ODV4-Q-09`: `$select` an Entity und Collection in der Folge des Modells | `TestAbnahmeMatrix/ODV4-Q-09` | `$select` wird ignoriert; die Mitglieder stehen in der Folge der URL | andere |
| T13 | `ODV4-Q-10` a: `$orderby` mit `desc`, `null` am Ende; b: zweites Element bei Gleichstand im ersten Schlüssel, gegen die Hilfsordnung nach dem Schlüssel | `TestAbnahmeMatrix/ODV4-Q-10` | `desc` ordnet aufsteigend; `null` ordnet bei `desc` zuerst; nur das erste Element von `$orderby` wird ausgewertet; das zweite Element ordnet aufsteigend | andere |
| T14 | `ODV4-Q-11`: `$skip` vor `$top` | `TestAbnahmeMatrix/ODV4-Q-11` | `$top` wirkt vor `$skip` | andere |
| T15 | `ODV4-Q-12`: `@odata.count` vor dem Paging, `/$count` mit `$filter` | `TestAbnahmeMatrix/ODV4-Q-12` | `@odata.count` zählt nach `$top`; `/$count` ignoriert `$filter` | andere |
| T16 | `ODV4-T-01`: `Edm.String` | `TestAbnahmeMatrix/ODV4-T-01` | der Schreiber setzt keine Anführungszeichen um einen String | andere |
| T17 | `ODV4-T-02`: `Edm.Boolean` | `TestAbnahmeMatrix/ODV4-T-02` | `false` wird als `null` geschrieben | andere |
| T18 | `ODV4-T-03`: `Edm.Int32` als Zahl | `TestAbnahmeMatrix/ODV4-T-03` | `Edm.Int32` wird als String geschrieben | andere |
| T19 | `ODV4-T-04`: `Edm.Int64` als Zahl ohne Rundung, mit `IEEE754Compatible` als String | `TestAbnahmeMatrix/ODV4-T-04` | `Edm.Int64` läuft über `float64`; `IEEE754Compatible` bleibt ohne Wirkung | andere |
| T20 | `ODV4-T-05`: `Edm.Decimal` mit Vorzeichen | `TestAbnahmeMatrix/ODV4-T-05` | das Vorzeichen eines Decimal geht verloren | andere |
| T21 | `ODV4-T-06`: `Edm.Guid` in Kleinbuchstaben (Ref des Produkts 3 mit Hex-Buchstaben, R5) | `TestAbnahmeMatrix/ODV4-T-06` | die Guid wird in Großbuchstaben geschrieben | andere |
| T22 | `ODV4-T-07`: `Edm.Date` | `TestAbnahmeMatrix/ODV4-T-07` | das Datum trägt eine Uhrzeit | andere |
| T23 | `ODV4-T-08`: `Edm.DateTimeOffset` mit Bruchteil | `TestAbnahmeMatrix/ODV4-T-08` | der Bruchteil der Sekunde wird abgeschnitten | andere |
| T24 | `ODV4-H-01`: Stufe aus `Accept` (`none`), keine Ablehnung (`text/html`), `full` als `minimal` | `TestAbnahmeMatrix/ODV4-H-01` | `none` ergibt `minimal`; `text/html` ergibt einen Fehler; `full` setzt `full` in den `Content-Type` (drei Läufe) | andere |
| T25 | `ODV4-H-02`: `OData-Version: 4.0` an jeder Antwortart (Collection, Entity, Anzahl, `$metadata`, 201, 204) | `TestAbnahmeMatrix/ODV4-H-02` | das Kennzeichen fehlt bei `204`; es fehlt bei `$metadata` (zwei Läufe) | andere |
| T26 | `ODV4-H-03`: `Content-Type` von `$metadata`, Anzahl, Collection ohne `IEEE754Compatible` | `TestAbnahmeMatrix/ODV4-H-03` | `$metadata` mit `text/xml`; Anzahl mit `application/json`; `IEEE754Compatible` ohne Anfrage (drei Läufe) | andere |
| T27 | `ODV4-E-01`: ungültiger Filter und `$top` negativ sind `400`, `ODR-E6005` | `TestAbnahmeMatrix/ODV4-E-01` | der Filter mit unbekannter Property ergibt 500; `$top=-1` wird als 0 gelesen | andere |
| T28 | `ODV4-E-02`: unbekanntes Set und fehlende Entity (`complete`) sind `404`, `ODR-E6008` | `TestAbnahmeMatrix/ODV4-E-02` | ein unbekanntes Set ergibt 400; die fehlende Entity ergibt `ODR-E6003` | andere |
| T29 | `ODV4-E-03`: `$search`, `$apply`, `$expand`, Service-Dokument `501`, `OPTIONS` `ODR-E6002` | `TestAbnahmeMatrix/ODV4-E-03` | `$search` wird ignoriert; `$expand` wird ignoriert; das Service-Dokument ergibt 200; `OPTIONS` wird wie `GET` beantwortet (vier Läufe) | andere |
| T30 | `ODV4-E-04`: Coverage strikt und lax, Zähler, Lookup (`partial` ist nie `404`), Mutation unabhängig vom Modus | `TestAbnahmeMatrix/ODV4-E-04` | der strikte Modus antwortet 200; lax setzt `X-Recorder-Coverage` nicht; ein Lookup auf `partial` ergibt 404 (drei Läufe) | andere |
| T31 | `ODV4-CRUD-01`: `POST` mit `201`, `Location` aus der Konfiguration, Entity, Zahl und Ordnung nach Zahl | `TestAbnahmeMatrix/ODV4-CRUD-01` | `Location` fehlt; `Location` trägt den Host der Anfrage; `Edm.Int64` im Body fehlerhaft gespeichert (drei Läufe) | andere |
| T32 | `ODV4-CRUD-02`: `PATCH` ändert nur die genannten Mitglieder | `TestAbnahmeMatrix/ODV4-CRUD-02` | `PATCH` ersetzt den Eintrag (nicht genannte werden `null`) | andere |
| T33 | `ODV4-CRUD-03`: `PUT` ersetzt, nicht genannte nullbare werden `null`; c, d: `PUT` an einen `partial`-Eintrag ersetzt ihn und macht ihn `complete` | `TestAbnahmeMatrix/ODV4-CRUD-03` | `PUT` behält die nicht genannten Mitglieder; `PUT` an `partial` ergibt `ODR-E6003`; „der Eintrag bleibt nach dem `PUT` `partial`“ ist hier ungefangen (akzeptiertes Negativ, `SPEC-049` Nr. 10 i, Ersatz-Fänger `TestPutPartial` im Test des Dienstes) | andere |
| T34 | `ODV4-CRUD-04`: `DELETE` setzt einen Tombstone, `404` danach, Anzahl und Collection ohne ihn; f: `POST` über den Tombstone legt neu an | `TestAbnahmeMatrix/ODV4-CRUD-04` | `DELETE` entfernt den Eintrag (das zweite `DELETE` ergibt 204); der Tombstone erscheint in der Collection; `POST` über einen Tombstone ergibt `ODR-E6009`; der neue Eintrag behält Mitglieder des gelöschten | andere |
| T35 | `ODV4-CRUD-05`: `null`, leerer Text und `{}` bleiben unterscheidbar | `TestAbnahmeMatrix/ODV4-CRUD-05` | `null` im `PATCH` gilt als nicht genannt; der leere Text wird `null`; `{}` ändert den Eintrag | andere |
| T36 | `ODV4-CRUD-06`: Duplikat `409`, fehlender Schlüssel und Widerspruch `400`, fehlend `404`, `partial` `422`, Zustand unverändert | `TestAbnahmeMatrix/ODV4-CRUD-06` | `POST` mit vorhandenem Schlüssel überschreibt (201); `PATCH` an `partial` ändert; die Revision wächst bei einem Fehler (drei Läufe) | andere |
| T37 | Nr. 3 e: genau 36 IDs aus Nr. 4, 92 Schritte, jeder gelaufen; die acht Typen von `model.UnterstuetzteEDMTypen()` je in einer Fixture `ODV4-T` | `TestAbnahmeMatrixZaehlung` (Tabelle, Bereiche, Typen), `TestAbnahmeMatrix` (jeder Schritt gelaufen) | eine ID aus der Tabelle des Tests entfernt; ein Schritt entfernt; eine ID doppelt; ein Schritt läuft nicht | andere |
| T38 | Nr. 3 c: der Vergleich nennt die Menge der Header genau (kein weiterer, keiner fehlt, Wert genau) | `TestAbnahmeMatrixVergleich/Header` | der Vergleich prüft nur die Header der Erwartung | andere |
| T39 | Nr. 3 c: der Body ist byte-gleich (ein Zeilenende am Ende ist ein Befund) | `TestAbnahmeMatrixVergleich/Body` | der Vergleich kürzt Leerraum am Ende | andere |
| T40 | Nr. 3 c: bei einem Fehler zählen Status und Code beide | `TestAbnahmeMatrixVergleich/Fehler` | der Vergleich prüft nur den Status; nur den Code (zwei Läufe) | andere |
| T41 | Nr. 3 b: jede ID beginnt im Ausgangszustand (frische Datei) | `TestAbnahmeMatrixAusgangszustand` | der Aufbau teilt die Datei zwischen zwei IDs | andere |
| T42 | Nr. 3 d: Revision nach der ID gleich der Zahl der erfolgreichen Mutationen | `TestAbnahmeMatrix/ODV4-CRUD-04` | die Mutation ruft `Advance` nicht auf | andere |
| T43 | Nr. 9 a: kein `t.Skip`, kein `testing.Short` in den Dateien der Abnahme | `TestAbnahmeKeinSkip` | ein `t.Skip` in `abnahme_neustart_test.go` | andere |
| T44 | Nr. 5 a: Revision nach dem Reset von Lauf i ist 3·(i−1)+1, nach dem Lauf 3·i | `TestAbnahmeDeterminismus` | der Reset erhöht die Revision nicht | andere |
| T45 | Nr. 5 b: Lauf 1 gleich dem Orakel in allen 14 Schritten | `TestAbnahmeDeterminismus` | Schritt D05 ohne die Änderung des `PATCH` (Name unverändert) | andere |
| T46 | Nr. 5 b, c: jeder Lauf ab 2 gleich Lauf 1 (Antworten, Stand nach Reset, Stand nach dem Lauf) | `TestAbnahmeDeterminismus` | die `entity_version` wächst ab Lauf 2 um 2 statt 1 | andere |
| T47 | Nr. 5 c: der Stand nach Reset ist der Ausgangszustand (Version 0, kein Tombstone) in jedem Lauf | `TestAbnahmeDeterminismus` | der Reset lässt den Tombstone von Produkt 3 stehen | andere |
| T48 | Nr. 5 c: der Cursor steht nach dem Reset auf Position 0, vorher auf 5 (Kontrollwert) | `TestAbnahmeDeterminismus` | der Reset ruft `Cursors.ResetAll` nicht | andere |
| T49 | Nr. 5 d: der Vergleich nimmt allein die Namen aus `VOLATILE_HEADERS` aus (ohne Beachtung der Schreibweise), vergleicht Status, Body, Namen, Zahl und Folge der Werte | `TestAbnahmeDeterminismusVergleich` | die Ausnahme trifft `Content-Length`; sie beachtet die Schreibweise; die Folge der Werte zählt nicht (drei Läufe) | andere |
| T50 | Nr. 5 d: jeder Name einer Antwort liegt in `Content-Type`, `OData-Version`, `X-Recorder-Coverage`, `Location` und keiner in `VOLATILE_HEADERS` | `TestAbnahmeDeterminismus` | der Mock setzt `Date` an jede Antwort | andere |
| T51 | Nr. 5 e: ein abweichender Lauf macht den Test rot, die Zeile `Läufe: 100, abweichend: <k>` | `TestAbnahmeDeterminismusAuswertung` | eine Abweichung wird nur protokolliert | andere |
| T52 | Nr. 5 f: die Frist von 120 s ist gesetzt, ihr Ablauf ist `Fatal` | `TestAbnahmeDeterminismusFrist` (Obergrenze beidseitig: Fristen von 125 s und 180 s sind Fatal; abgelaufen; im Lauf abgelaufen), `TestAbnahmeDeterminismus` (der Test setzt die Frist) | der Lauf setzt keine Frist; der Wächter lässt Fristen bis 150 s zu | andere |
| T53 | Nr. 5 a: genau 100 Läufe laufen | `TestAbnahmeDeterminismus` | die Schleife zählt bis 99 | andere |
| T54 | Nr. 6 a, f: 20 Goroutinen starten auf ein gemeinsames Signal, 20 enden, 228 Schritte verglichen | `TestAbnahmeIsolation` | die Menge hat 19 Workspaces; ein Workspace läuft nicht | nebenläufig |
| T55 | Nr. 6 b, c: jeder Schritt von S(k) gleich dem Orakel von k | `TestAbnahmeIsolation` | `EntityRepository.ListSet` ohne Filter auf den Workspace | nebenläufig |
| T56 | Nr. 6 c: Revision je Workspace gleich dem Orakel | `TestAbnahmeIsolation` | `Advance` ohne die Kennung des Workspace in der Bedingung | nebenläufig |
| T57 | Nr. 6 e: Cursor-Liste von `ws-k` gleich dem Orakel, die Resets von 0, 5, 10, 15 berühren die übrigen nicht | `TestAbnahmeIsolation` | `Cursors.ResetAll` setzt die Cursors aller Workspaces zurück | nebenläufig |
| T58 | Nr. 6 b: Snapshot `s1` in allen 20 Workspaces, die Liste nennt nur den eigenen | `TestAbnahmeIsolation` | der Name eines Snapshots ist global eindeutig; `ListWorkspace` ohne Filter | nebenläufig |
| T59 | Nr. 6 b: `Restore` wirkt auf den einen Workspace | `TestAbnahmeIsolation` | `Restore` ersetzt den State des ersten Workspace der Liste | nebenläufig |
| T60 | Nr. 6 d: kein Eintrag von `ws-k` trägt einen Wert von `ws-j` (Schleife über 20·19 Paare) | `TestAbnahmeIsolation` | `Entities.Put` schreibt die Kennung `ws-00` | nebenläufig |
| T61 | Nr. 6 c: die Orakel der 20 Workspaces unterscheiden sich paarweise (Schritte 2 und 5, R6) | `TestAbnahmeIsolationOrakelUnterscheidet` | die Folge S(k) benutzt k nicht | andere |
| T62 | Nr. 6 a: alle 20 Goroutinen sind am Startsignal, bevor eine den ersten Schritt beginnt | `TestAbnahmeIsolationStart` | das Signal entfällt | nebenläufig |
| T63 | Nr. 7 a: nach dem Neustart gibt es neue Instanzen; die alten sind geschlossen und benutzbar nicht mehr | `TestAbnahmeNeustart` | der Aufbau benutzt die alte Datenbank weiter | andere |
| T64 | Nr. 7 d (1): Stand von `ws-1` und `ws-2` nach dem Neustart gleich dem festgehaltenen (Version, `Deleted`, Payload, Revision, Cursors) | `TestAbnahmeNeustart` | der Adapter der Ablage schreibt `entity_version` nicht | andere |
| T65 | Nr. 7 d (2): die vier Antworten sind byte-gleich | `TestAbnahmeNeustart` | der Tombstone von Produkt 3 geht beim Öffnen verloren | andere |
| T66 | Nr. 7 d (3): die Snapshots von `ws-1` (Reihenfolge, Felder), `ws-2` keiner | `TestAbnahmeNeustart` | `CreatedAt` des Snapshots wird beim Lesen durch die Zeit des Öffnens ersetzt | andere |
| T67 | Nr. 7 b: ein Fehlversuch (`ODR-E6009`) und ein beendeter Kontext hinterlassen nichts (Version 1 von Produkt 1, Revision 5) | `TestAbnahmeNeustart` | die Mutation schreibt vor der Prüfung auf ein vorhandenes Produkt; der Rollback fehlt | andere |
| T68 | Nr. 7 d (4): `Restore` von `vor` nach dem Neustart: State nach Schritt 4, Revision 6, Cursors unverändert (`ws-1` Position 3, `ws-2` Position 2, R8) | `TestAbnahmeNeustart` | `Restore` lässt den Tombstone von Produkt 3 aus; es setzt die Cursors zurück | andere |
| T69 | Nr. 7 d (5): Version 2 und Revision 7 danach | `TestAbnahmeNeustart` | das Öffnen setzt die Revision auf 0 | andere |
| T70 | Nr. 7 d (6): ein zweiter Neustart ohne Änderung ergibt denselben Stand (`SPEC-103` Nr. 3) | `TestAbnahmeNeustart` | das Öffnen migriert die Datei erneut und leert den State | andere |
| T71 | Nr. 8 (a): 20 gleichzeitige `POST` mit gleichem Schlüssel: eine 201, 19 `ODR-E6009`, Version 1, Revision plus 1 | `TestAbnahmeNebenlaeufigPost` | die Prüfung auf ein vorhandenes Produkt liest `Konfig.Entities` außerhalb der Transaktion | nebenläufig |
| T72 | Nr. 8 (a): 20 gleichzeitige `DELETE`: eine 204, 19 `ODR-E6008`, Version 1, Revision plus 1 | `TestAbnahmeNebenlaeufigDelete` | der Tombstone wird nach der Prüfung außerhalb der Transaktion geschrieben | nebenläufig |
| T73 | Nr. 8 (b): 20 gleichzeitige `PATCH`: 20 mal 204, Revision plus 20, Version 20, kein `ODR-E9001` | `TestAbnahmeNebenlaeufigPatch` | `erwartet` stammt aus dem Lesen des Workspace in Schritt (f) statt aus der Transaktion | nebenläufig |
| T74 | Nr. 8 (c1): Löschen des Workspace am Haltepunkt: `ODR-E7001`, keine Zeile | `TestAbnahmeNebenlaeufigLoeschen` | `f` liest den Workspace nicht noch einmal und benutzt den aus Schritt (f) | nebenläufig |
| T75 | Nr. 8 (c2): Reset am Haltepunkt: 204, Version 1, Revision des Reset plus 1 | `TestAbnahmeNebenlaeufigReset` | `erwartet` stammt aus Schritt (f) | nebenläufig |
| T76 | Nr. 8 Haltepunkt: hält bis *n* eingetroffen sind, Freigabe von Hand, Frist 10 s ist `Fatal` | `TestAbnahmeHaltepunkt` | der Dekorator gibt bei *n*−1 frei; die Frist entfällt | nebenläufig |
| T77 | Nr. 8: Raten je Mutation von T71 bis T75 sind 20 von 20 | Absatz *Mutationen und Raten* in §3 (Ausgabe von 20 Prozessläufen je Mutation) | jede Mutation von T71 bis T75 in 20 Läufen | nebenläufig |
| T78 | Nr. 8: die Zweige der Tests unter dem Detektor stammen aus dem Abdeckungsprofil | Absatz *Mutationen und Raten* in §3 (Profil von `go test -race -coverprofile -coverpkg`) | der Zweig von `Entities.Put` des Mock hat im Profil 0 Treffer | nebenläufig |
| T79 | Nr. 10 a: HTTP-Adapter, Routen und Rahmung sind nicht Gegenstand | keiner; akzeptiertes Negativ: Grund `SPEC-049` Nr. 1, Folge: eine falsche Route bleibt hier unbemerkt (Nr. 10 a) | – | andere |
| T80 | Nr. 10 b: Binary, `main`, Signale und Start im Modus `mock` | keiner; akzeptiertes Negativ: Grund Nr. 7 a, Folge: der Neustart über die Prozessgrenze fängt für Mutationen `slice-bootstrap-m3`, für Snapshots ist er ohne Fänger (Nr. 10 b) | – | andere |
| T81 | Nr. 10 c: Absturz und Fehler der Ablage | keiner; akzeptiertes Negativ: Grund `SPEC-048`, Folge: ein Absturz in einer Mutation ist hier nicht belegt (Nr. 10 c) | – | andere |
| T82 | Nr. 10 d: Mischung aus Vorher und Nachher beim gleichzeitigen Lesen | keiner; akzeptiertes Negativ: Grund und Folge Nr. 10 d | – | andere |
| T83 | Nr. 10 e bis i: Last, Soll-Zeilen, zwei Vergleichsfunktionen, Kombinationen, „der Eintrag bleibt nach `PUT` `partial`“ (i) | keiner; akzeptiertes Negativ: Grund und Folge in Nr. 10 e bis i | – | andere |

**Mengen** (Zählung gegen die Quelle, `AGENTS.md` §3.15; die Zahlen prüft `TestAbnahmeMatrixZaehlung` bzw. der Test der Menge):

- **M1 Fixture-IDs:** 36 = 3 (`GET`) + 6 (`CRUD`) + 12 (`Q`) + 8 (`T`) + 3 (`H`) + 4 (`E`), aus der Tabelle oben in `SPEC-049`; Fänger T37.
- **M2 Schritte der Fixtures:** 92 (GET 7, CRUD 29, Q 16, T 9, H 13, E 18); Fänger T37.
- **M3 Typen:** die acht EDM-Typen der Zeile EDM, je einer in `ODV4-T-01..08`; der Test bildet die Zählung aus `model.UnterstuetzteEDMTypen()` und vergleicht sie mit 8; Fänger T16 bis T23.
- **M4 Header-Namen einer Antwort:** vier (`Content-Type`, `OData-Version`, `X-Recorder-Coverage`, `Location`); Fänger T50.
- **M5 Workspaces der Isolation:** 20; Fänger T54. **M6 Schritte:** 228 = 20·11 + 4·2; Fänger T54.
- **M7 Fenster der Mutation (`SPEC-192` Nr. 8):** (a) Prüfen und Schreiben bei `POST` und bei `DELETE`, (b) Schreiben und Revision, (c) Zuordnung und Transaktion mit Löschen und mit Reset; fünf Läufe; Fänger T71 bis T75. (d) ist Nr. 6, (e) und (f) sind Nr. 10 d und Nr. 8 (Antwort in der Transaktion) und haben keine eigene Zeile.
- **M8 Läufe der Determinismus-Messung:** 100; Fänger T53.

**Mutationen und Raten** (Implementer, 2026-11-08, `AGENTS.md` §3.10). Messender ist der Implementer-Agent auf dem Host (Docker, `--memory 3g`, unter `flock`); jede Mutation läuft in einer Wegwerfkopie des Baums (ein Text wird genau einmal ersetzt, das Test-Binary des Pakets `src/bootstrap` wird neu übersetzt, der Elternfall läuft mit `-test.run`), die Spalte *roter Test* nennt die Teilfälle, die rot wurden (Namen ohne `TestAbnahmeMatrix/`). Wo die Mutation der Tabelle oben im Code nicht so ausdrückbar war, steht die angewandte Mutation; ein Fänger zählt nur, wenn der benannte Fall rot wurde.

| Nr. | angewandte Mutation | roter Test |
|---|---|---|
| T01 | die Schleife des Schreibers über `Entities` lässt das letzte Element aus | `ODV4-GET-01` (und jede Collection) |
| T02 | (a) der Schreiber lässt ein Mitglied mit `null` aus; (b) ein Lookup ohne Treffer liefert `200` | (a) `GET-01`, `GET-02`, `Q-10`, `H-02`, `H-03`, `CRUD-03`, `CRUD-05`; (b) `GET-02`, `GET-03`, `E-02`, `E-04` |
| T03 | (a) `value` fehlt bei leerer Collection; (b) die Anzahl `0` ergibt einen leeren Body | `GET-03` ((a) auch `Q-12`) |
| T04 bis T09 | `opTrifft`/`Vergleiche`: `eq` wie `ne`; `ne` mit `null` falsch; `gt` wie `ge`; `ge` wie `gt`; `lt` wie `le`; `le` wie `lt` | `Q-01`; `Q-02`; `Q-03`; `Q-04` (und `T-03`); `Q-05`; `Q-06` |
| T10 | `and` wertet wie `or`; `or` wertet wie `and` (zwei Läufe) | `Q-07` (beide) |
| T11 | `not` gibt den Operanden unverändert zurück | `Q-08` |
| T12 | (a) `projiziere` ohne `$select`; (b) Projektion und Schreiber liefern die Mitglieder in der Folge der URL | (a) jede Fixture mit `$select`; (b) `Q-09`. Die Projektion allein (b ohne den Schreiber) bleibt grün: der Schreiber schreibt die Mitglieder in der Folge des Modells |
| T13 | (a) `desc` ordnet aufsteigend; (b) `null` ordnet bei `desc` zuerst; (c) nur das erste Element von `$orderby` wird ausgewertet; (d) das zweite Element ordnet aufsteigend | (a) `Q-10` a, `Q-11`, `CRUD-01`; (b) `Q-10` a; (c) und (d) `Q-10` b |
| T14 | `$top` wirkt vor `$skip` | `Q-11` |
| T15 | (a) `@odata.count` zählt nach `$top`; (b) `/$count` ignoriert `$filter` | `Q-12` (beide) |
| T16 bis T23 | der Schreiber: String ohne Anführungszeichen; `false` als `null`; `Edm.Int32` als String; (a) `Edm.Int64` über `float64`, (b) `IEEE754Compatible` ohne Wirkung; `Edm.Decimal` ohne Vorzeichen; `Edm.Guid` in Großbuchstaben; `Edm.Date` mit Uhrzeit; `Edm.DateTimeOffset` ohne Bruchteil | `T-01`; `T-02`; `T-03`; `T-04` (beide); `T-05`; `T-06` und `CRUD-01` (R5); `T-07`; `T-08` |
| T24 | (a) `none` ergibt `minimal`; (b) `text/html` ergibt einen Fehler; (c) `full` im `Content-Type` | `H-01` (alle drei) |
| T25 | (a) `OData-Version` fehlt bei `204`; (b) fehlt bei `$metadata` | (a) `H-02`, `CRUD-02` bis `CRUD-05`; (b) `H-02`, `H-03` |
| T26 | (a) `$metadata` mit `text/xml`; (b) Anzahl mit `application/json`; (c) `IEEE754Compatible` ohne Anfrage | `H-03` (alle drei, daneben weitere) |
| T27 | (a) Filter-Fehler ergibt `500`/`ODR-E9001`; (b) `$top=-1` wird als 0 gelesen | `E-01` (beide) |
| T28 | (a) unbekanntes Set ergibt `400`; (b) die fehlende Entity bei `complete` ergibt `ODR-E6003` | `E-02` ((b) auch `GET-02`, `GET-03`) |
| T29 | (a) `$search`, (b) `$apply`, (c) `$expand` werden ignoriert; (d) das Service-Dokument wird als `$metadata` beantwortet; (e) `OPTIONS` wie `GET` | `E-03` (fünf Läufe) |
| T30 | (a) strikter Modus antwortet `200`; (b) lax setzt bei `partial` keinen `X-Recorder-Coverage`; (c) Lookup auf `partial` ergibt `404`; (d) die Mutation hängt vom Modus ab | `E-04` (vier Läufe) |
| T31 | (a) `Location` fehlt; (b) Basis der Adressen nicht aus der Konfiguration; (c) `Edm.Int64` im Body gekürzt gespeichert | (a) `H-02`, `CRUD-01`; (b) jede Fixture; (c) `CRUD-01`, `CRUD-02` |
| T32 bis T35 | `PATCH` ersetzt den Eintrag; `PUT` behält die nicht genannten; (a) `DELETE` ohne Tombstone, (b) Tombstone in der Collection; (a) `null` im `PATCH` ignoriert, (b) leerer Text wird `null`, (c) `{}` ändert den Eintrag; Zellen von `SPEC-193` Nr. 2: (a) `PUT` an `partial` ergibt `ODR-E6003`, (b) `POST` über einen Tombstone ergibt `ODR-E6009` | `CRUD-02` (und `H-02`, `CRUD-05`); `CRUD-03`; `CRUD-04` (beide); `CRUD-05` (alle drei); `CRUD-03` c und d (Zelle a, rot an c und d); `CRUD-04` f (Zelle b und „der neue Eintrag behält die `Ref` des gelöschten“, beide rot an f). „Der Eintrag bleibt nach `PUT` `partial`“ wird von `CRUD-03` d **nicht** rot (nur `H-02` f, ein `PATCH` danach): akzeptiertes Negativ, `SPEC-049` Nr. 10 i (Rückgabe R12, entschieden) |
| T36 | (a) `POST` überschreibt; (b) `PATCH` an `partial` ändert; (c) die Revision wächst bei einem Fehler | `CRUD-06` ((c) auch `E-04`, `CRUD-04`) |
| T37 | (a) eine ID entfernt; (b) ein Schritt entfernt; (c) eine ID doppelt; (d) ein Schritt läuft nicht | (a) bis (c) `TestAbnahmeMatrixZaehlung`; (d) `TestAbnahmeMatrix` |
| T38 bis T40 | der Vergleich prüft nur die Header der Erwartung; kürzt Leerraum am Ende; prüft bei einem Fehler nur den Code, nur den Status (zwei Läufe) | `TestAbnahmeMatrixVergleich/Header`; `/Body`; `/Fehler` (beide) |
| T41 | der Aufbau teilt die Umgebung zwischen zwei IDs | `TestAbnahmeMatrixAusgangszustand` |
| T42 | `Advance` wird in der Mutation nicht gerufen | `H-02`, `CRUD-01` bis `CRUD-05` |
| T43 | ein `t.Skip` in `src/bootstrap/abnahme_neustart_test.go` | `TestAbnahmeKeinSkip` |
| T44 | der Reset erhöht die Revision nicht | `TestAbnahmeDeterminismus` |
| T45 | `PATCH` übernimmt den Namen nicht | `TestAbnahmeDeterminismus` (Lauf 1, Schritt D05, gegen das Orakel) |
| T46 | `entity_version` wächst ab Lauf 2 um 2 | `TestAbnahmeDeterminismus` (Lauf 2, Stand nach dem Lauf) |
| T47 | der Reset schreibt den Tombstone von Produkt 3 | `TestAbnahmeDeterminismus` (Lauf 1, Stand nach Reset) |
| T48 | der Reset ruft `Cursors.ResetAll` nicht | `TestAbnahmeDeterminismus` (Lauf 1, Stand nach Reset: Cursor auf 5) |
| T49 | die Ausnahme trifft `Content-Length`; sie beachtet die Schreibweise; die Folge der Werte zählt nicht (drei Läufe) | `TestAbnahmeDeterminismusVergleich` (alle drei) |
| T50 | der Mock setzt `Date` an jede Antwort | `TestAbnahmeDeterminismus` |
| T51 | eine Abweichung wird nur protokolliert | `TestAbnahmeDeterminismusAuswertung` |
| T52 | (a) die Läufe prüfen die Frist nicht; (b) der Test setzt keine Frist; (c) der Ablauf der Frist ist keine Fatal; (d) der Wächter lässt Fristen bis 150 s zu | (a), (c), (d) `TestAbnahmeDeterminismusFrist` ((d) am Fall 125 s); (b) `TestAbnahmeDeterminismus` |
| T53 | die Schleife zählt bis 99 | `TestAbnahmeDeterminismus` (`Läufe: 99`) |
| T54 | (a) die Menge hat 19 Workspaces; (b) der Workspace 7 läuft nicht | `TestAbnahmeIsolation` (beide) |
| T55 bis T60 | `ListSet` ohne Filter auf den Workspace; `Advance` ohne die Kennung in der Bedingung; `ResetAll` setzt alle Cursors zurück; (a) Name des Snapshots global eindeutig, (b) `ListWorkspace` ohne Filter; `Restore` ersetzt den State des ersten Workspace der Liste; `Entities.Put` schreibt die Kennung `ws-00` | `TestAbnahmeIsolation` (jede) |
| T61 | die Folge S(k) benutzt k nicht | `TestAbnahmeIsolationOrakelUnterscheidet` |
| T62 | das Startsignal entfällt | `TestAbnahmeIsolationStart` (Rate unten) |
| T63 | der Neustart gibt die alte Umgebung zurück | `TestAbnahmeNeustart` |
| T64 bis T66 | der Adapter schreibt `entity_version` nicht; der Tombstone geht beim Lesen verloren; `CreatedAt` des Snapshots wird beim Lesen durch die Zeit des Lesens ersetzt | `TestAbnahmeNeustart` ((1) bis (3)) |
| T67 | die Mutation schreibt vor der Prüfung und die Transaktion rollt bei einem Fehler nicht zurück | `TestAbnahmeNeustart` (Fehlversuche, (1)) |
| T68 | (a) `Restore` lässt den Tombstone von Produkt 3 aus; (b) `Restore` setzt die Cursors zurück | `TestAbnahmeNeustart` ((4)) |
| T69 | (a) das Öffnen setzt die Revision auf 0; (b) die Version zählt nach dem Öffnen nicht weiter | `TestAbnahmeNeustart` ((a) bei (1), (b) bei (5)) |
| T70 | das Öffnen löscht Einträge mit `entity_version` ab 2 | `TestAbnahmeNeustart` ((6): Zahl der Einträge 6 gegen 7) |

**Fix-Runde (Befunde F1 bis F4, §3.17).** Mengen und Elementzahlen: *F1* `$orderby`: die Elemente der Ordnung (Richtung des ersten, des zweiten, Zahl der ausgewerteten, Typ der Spalte), vier Mutationen an `ODV4-Q-10` b rot: nur das erste Element, das zweite aufsteigend, alle in der Richtung des ersten, und eine am Befund nicht genannten Element, die Spalte `Boolean` ordnet `true` vor `false`. *F6* die zwei Zellen von `SPEC-193` Nr. 2 (elf von 20 in der Matrix): `PUT` an `partial` ergibt `ODR-E6003` (rot an `CRUD-03` c und d), `POST` über Tombstone ergibt `ODR-E6009` (rot an `CRUD-04` f), der neue Eintrag behält die `Ref` des gelöschten (rot an `CRUD-04` f, ein nicht genanntes Element); der Eintrag bleibt `partial` nach `PUT` (R12, nur `H-02` f rot). *F2* Obergrenze der Frist: Fristen von 125 s und 180 s, die Mutation „Wächter bis 150 s“ rot an `TestAbnahmeDeterminismusFrist`. Die Zahl 88 ist in Zähler, Meldung, Kommentaren und Bereichssumme 92 (7, 29, 16, 9, 13, 18).

**Raten der Fenster (T62, T71 bis T76).** Je Mutation 20 Prozesse mit Detektor (Binary mit `-race`) und 20 ohne, jeder Prozess `-test.count=1` des einen Tests, nacheinander; die Last ist der Eins-Minuten-Wert aus `/proc/loadavg` des Hosts zu Beginn und am Ende der Reihe; Messender ist der Implementer-Agent mit einem Runner im Scratchpad (nicht `tools/harness/mutationslauf.sh`: dessen Nullprobe fährt das Paket `src/bootstrap` mit `-test.timeout=2m`, das Paket braucht aber rund 150 s, Ausgang `Ausgangsstand rot`, Rückgabe R10). Die Zusage gilt als gefangen, weil jede Zeile *rot in 20 von 20* ist:

| Nr. | angewandte Mutation (Wirkung im roten Lauf) | Test | mit Detektor | ohne Detektor | Last |
|---|---|---|---|---|---|
| T71 (Lesen vorab) | `schreibe` liest `Konfig.Entities` vor der Transaktion, `imWorkspace` benutzt den Wert (20 mal `201`, 20 Eintreffende am Haltepunkt der Entities) | `TestAbnahmeNebenlaeufigPost` | 20 von 20 | 20 von 20 | 4,56 / 4,00 |
| T71 (Schreiben danach) | der Eintrag wird nach der Transaktion über `Konfig.Entities` geschrieben (20 mal `201`) | `TestAbnahmeNebenlaeufigPost` | 20 von 20 | 20 von 20 | 3,94 / 2,99 |
| T72 (Lesen vorab) | wie T71 (Lesen vorab), `DELETE` | `TestAbnahmeNebenlaeufigDelete` | 20 von 20 | 20 von 20 | 4,00 / 3,10 |
| T72 (Schreiben danach) | wie T71 (Schreiben danach), `DELETE` (20 mal `204`) | `TestAbnahmeNebenlaeufigDelete` | 20 von 20 | 20 von 20 | 2,99 / 2,58 |
| T73 | `Advance` mit der Revision der Zuordnung statt der der Transaktion (eine `204`, 19 mal `ODR-E9001`) | `TestAbnahmeNebenlaeufigPatch` | 20 von 20 | 20 von 20 | 3,23 / 3,82 |
| T74 | `imWorkspace` liest den Workspace in der Transaktion nicht (`ODR-E6008` statt `ODR-E7001`) | `TestAbnahmeNebenlaeufigLoeschen` | 20 von 20 | 20 von 20 | 3,82 / 3,77 |
| T75 | wie T73 am Haltepunkt des Reset (`ODR-E9001` statt `204`) | `TestAbnahmeNebenlaeufigReset` | 20 von 20 | 20 von 20 | 3,95 / 4,02 |
| T62 | das Startsignal entfällt (20 Goroutinen beginnen vor dem Signal) | `TestAbnahmeIsolationStart` | 20 von 20 | 20 von 20 | 2,28 / 2,34 |
| T76 (a) | der Haltepunkt gibt bei *n*−1 frei | `TestAbnahmeHaltepunkt` | 20 von 20 | 20 von 20 | 2,34 / 2,31 |
| T76 (b) | die Frist des Haltepunkts entfällt | `TestAbnahmeHaltepunkt` | 20 von 20 | 20 von 20 | 2,31 / 3,21 |

Ein Zugriff über `Konfig.Entities` **innerhalb** der Funktion der Transaktion (`Get` für T71, `Put` für T72, je zwei Mutationen) ist ebenfalls *rot in 20 von 20* mit und ohne Detektor (Last 9,32 / 2,05 bis 2,49 / 2,31), aber über die Frist und nicht über die Wirkung: der erste Aufrufer hält am Haltepunkt der Entities die Schreibsperre, die übrigen 19 warten auf `BEGIN IMMEDIATE` und erreichen ihn nie, nach 10 s meldet der Test `Haltepunkt nicht erreicht: 1 von 20 eingetroffen` (Rückgabe R9). Die Datenrennen-Zusage von `AGENTS.md` §3.10 trifft diese Zeilen nicht: keine der Mutationen ist ein Datenrennen, der Detektor meldet keines, die Wirkung zählen die Fänger.

**Raten der Isolation (T55 bis T60, Befund F4).** Je Mutation 20 Prozesse mit Detektor und 20 ohne, `-test.count=1`, der Test `TestAbnahmeIsolation`, nacheinander; Messender der Implementer-Agent mit dem Runner im Scratchpad, Last der Eins-Minuten-Wert des Hosts zu Beginn und am Ende der Reihe. Alle **rot in 20 von 20 mit und ohne Detektor**: T55 (4,72 / 3,80), T56 (3,80 / 3,86), T57 (3,86 / 4,37), T58 (a) Name global (4,37 / 3,53), T58 (b) `ListWorkspace` ohne Filter (3,53 / 3,31), T59 (3,31 / 4,12), T60 (4,12 / 4,17). Die Wirkung ist ein Fremdzustand, den ein Einzellauf nicht zeigt; kein Datenrennen.

**Dauer** (`SPEC-049` Nr. 5 f, erwartet unter 15 s, Frist 120 s). Die 100 Läufe brauchen ohne Detektor 3,3 s (Last 3) bis 6,2 s (Last 8 bis 13), mit Detektor 8,8 s (Last 6,9). Die Abnahme-Tests zusammen brauchen 26 s ohne Detektor (Last 3) und 34 s mit Detektor und Abdeckungsprofil; das Paket `src/bootstrap` gesamt rund 150 s ohne Detektor (Last 5 bis 13).

**Abdeckungsprofil** (T78): `go test -race -coverprofile -coverpkg` über `services/mock`, `services/workspace` und `services/snapshot` mit `-run TestAbnahme` ergibt (neu gemessen nach der Fix-Runde, 92 Schritte) 71,9 % der Anweisungen der drei Pakete. In `mock/schreiben.go`, `workspace/reset.go`, `workspace/wiederherstellen.go` und `snapshot/dienst.go` liegen 213 Blöcke, 142 mit Treffer, 71 ohne (unverändert: `PUT` an `partial` und `POST` über einen Tombstone laufen über Blöcke, die schon Treffer hatten). Die 71 sind Fehlerzweige und kein Zweig eines Fensters: der Fehler eines Ports (`abbilden(err, …)` nach `Get`, `Put`, `Advance`, `ReplaceAll`, `ResetAll`, `Create`, `GetWorkspace`), das Argument `ctx` gleich `nil`, die Form des Payloads (`liesEintrag`), der Typ des Eintrags bei `PATCH` und `PUT` (`typ != mu.eingabe.Typ`, `ODR-E6007`, `SPEC-193` Nr. 2 a), der Fehler des Schreibers, die Kennung einer Baseline statt eines Snapshots und ein Manifest, das die Prüfung nicht besteht. Akzeptiertes Negativ, Grund: Diese Zweige gehören den Tests der Dienste mit Fakes der Ports (`SPEC-129`, `SPEC-192`); Folge: eine Mutation in einem dieser Fehlerzweige fängt die Abnahme nicht. `imWorkspace` wird 640 mal betreten, der Block mit `Entities.Put` hat 484 Treffer (vorher 577 und 421); mit `-run TestAbnahmeKeinSkip` allein hat er 0 Treffer, die Zählung kippt also, wenn die Fenster-Tests nicht laufen.

**Mengen mit Generator und Orakel** (Handoff-Regel von `.claude/commands/implement-slice.md` Schritt 14): *Fixture-IDs* (M1): Generator `abErwarteteIDs` (Bereich × Nummer: 3, 6, 12, 8, 3, 4), Orakel die Tabelle von `SPEC-049`, 36 Elemente, `TestAbnahmeMatrixZaehlung`; *Schritte* (M2): 92, Summe der Bereiche 7, 29, 16, 9, 13, 18, `TestAbnahmeMatrixZaehlung` und der Zähler je ID in `TestAbnahmeMatrix`; *Typen* (M3): die acht von `model.UnterstuetzteEDMTypen()`, je einer in `ODV4-T-01` bis `-08` (Abgleich der gewählten Property mit dem Typ im CSDL), `TestAbnahmeMatrixZaehlung`; *Namen der Header* (M4): vier, `TestAbnahmeDeterminismus`; *Workspaces* (M5): 20, *Schritte der Isolation* (M6): 228 = 20·11 + 4·2 und 380 Paare, Orakel die Einzelläufe von k, `TestAbnahmeIsolation`; *Fenster* (M7): fünf, `TestAbnahmeNebenlaeufig*`; *Läufe* (M8): 100, `TestAbnahmeDeterminismus`.

**Akzeptiertes Negativ V2 (Verifikation, LOW): Obergrenze des Frist-Wächters (T52).** `TestAbnahmeDeterminismusFrist` führt die Fälle `abDetFrist+5s` und 180 s; eine Obergrenze zwischen 120 und 125 s bleibt grün (Mutationen f1 mit 121 s und f2 mit 124 s, vom Verifier gefahren), jede ab 126 s ist rot, die Untergrenze fängt `TestAbnahmeDeterminismus` (f5 rot). Grund: ein Fall mit `abDetFrist+1s` verlangt, dass der Zeitverbrauch bis `Deadline` unter einer Sekunde bleibt, und wird unter Last flatterig; der Test setzt die Frist selbst auf 120 s. Folge im Betrieb: der Wächter schützt allein vor einer künftig längeren Frist, und eine Verlängerung um mehr als 5 s fällt auf; kein Fix-Lauf (Durchstich vor Härtung).

**Messung der Größe** (`make groessenmarke` am Handoff, Basis `27acf86f`, Summe aus Einfügungen und Löschungen der sechs Pfade des Blocks): 2335 Zeilen nach der Fix-Runde (vorher 2325) von der Marke 9700, eine Auslastung von rund 0,24 (erwartet war rund 0,56); die Schätzung 7120 war um den Faktor 3 zu hoch, weil die Tabelle 83 Zeilen zählt und der Slice keinen Produktcode und keine Port-Fakes außer den zwei Dekoratoren trägt (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-snapshot-use-case` liegt in `done/`. Die Spezifikation trägt die Fixtures der Matrix `SPEC-049` mit Inhalt, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand beim Start** (Planner, 2026-11-07, `main` = `08843060`): Trigger Teil 1 erfüllt, `slice-snapshot-use-case` liegt seit dem Merge `56b87d09` in `done/`. Trigger Teil 2 (Fixtures der Matrix `SPEC-049` mit Inhalt in der Spezifikation) **nicht erfüllt**: `SPEC-049` führt die Adressen `ODV4-GET-01..03`, `-CRUD-01..06`, `-Q-01..12`, `-T-01..08`, `-H-01..03`, `-E-01..04` als „geplante Testadressen“ ohne Inhalt; er ist Bedingung für den ersten Code-Commit, nicht für den Start, und liegt beim Architect (Aufträge M1 bis M7). §1 *Ausdrücklich NICHT* und DoD der Sendungen von `slice-snapshot-use-case` gelesen (`AGENTS.md` §3.13): `welle-mock` §5 und `done/slice-snapshot-use-case.md` §7 weisen diesem Slice allein die Bedingung „Trigger in `done/`“ und den Prüfgegenstand Reset-Determinismus, Isolation und Neustart mit Snapshots zu; kein Ausschluss in §1 dieses Slice trifft sie (Snapshots und Restore sind Gegenstand der Neustart-Zeile, nicht ausgeschlossen). Codes: `model.AlleCodes()` zählt auf `main` 39; dieser Slice fügt keinen hinzu, solange kein Test einen Fehler findet (§1, vierter Punkt); `slice-admin-rahmen` (40. Code) und `slice-sqlite-interrupt-transaktion` mergen nach `main`, wer als zweiter merged, gleicht die Zahl an (`welle-mock` §5, Merge-Nähe). Gate-Kette (Nr. 7 der Übergangsregeln): `tests/` ist neu, `Makefile`, `harness/mk/`, `Dockerfile` und Stempel bleiben unberührt; braucht `tests/` im Build-Kontext eine Änderung des `Dockerfile`, geht das als Rückgabe an den Planner. Speicher beim Start: 14 GB verfügbar, Load 13.

**Größenmarke** (`SPEC-128`; die Faktoren stehen allein dort; Architect, 2026-11-07, Auftrag M7). Die Tabelle in §3 hat 83 Zeilen (*a* = 16 nebenläufige: T54 bis T60, T62, T71 bis T78; *b* = 67 andere, darunter fünf akzeptierte Negative T79 bis T83); Port-Fakes *f* = 2 (die Dekoratoren von `WorkspaceRepository` und `EntityRepository` als Haltepunkt, `SPEC-049` Nr. 8; die Ablage ist die echte); Produktcode *p* = 0 (die Tests ändern `src/` nur, wo sie einen Fehler finden, und ein Fix gehört der Marke des Slice, dem die Zusage gehört, `SPEC-049` Nr. 9 c). Richtwert nach `SPEC-128` Nr. 1, das Werkzeug rechnet: *S* = 85·16 + 80·67 + 200·2 + 0 = 7120 und *M* = 100·⌈135·7120/10000⌉ = 9700, also die Schnittfrage (*S* ≥ 2223); der Vorschlag steht in Rückgabe R3 (§6). Der Pfad `spec/spezifikation.md` steht nicht im Block: `SPEC-128` Nr. 3 m schließt `spec/` aus, und die Spezifikation zählt nicht zur Messung. Das Register führt die Überschätzung des Produktcodes und die Unterschätzung der Tests (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`): hier gibt es keinen Produktcode, die Zeilen der Tabelle sind realistisch gezählt (eine Zeile je Fixture-ID, nicht je Schritt), die Formel ist nicht geändert. **Der Block unter diesem Absatz** trägt den Teil des Architect (`pfad`, `produktcode`, `fakes`); `basis`, `schaetzung`, `marke` und, falls der Planner nicht schneidet, `schnitt-grund` setzt der Planner in eigenem Commit vor dem ersten Code-Commit (`AGENTS.md` §3.14).

**Marke und Schnittentscheidung** (Planner, 2026-11-07, Antwort auf Rückgabe R3): Marke **9700** (Schätzung 7120, Basis `27acf86f`, Werkzeug `make groessenmarke`; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat 27acf86f..<letzter Commit des Slice>` über die sechs Pfade des Blocks, Tests eingeschlossen). **Kein vorab benannter Schnitt**, der Grund steht als `schnitt-grund` im Block. Die Prüfung nach `AGENTS.md` §3.14: ein vorab benannter Teil muss ohne geteilten Mechanismus abtrennbar sein. Isolation und Nebenläufigkeit (T54 bis T60, T62, T71 bis T78, *S* = 1760) liest die Hilfen von `abnahme_aufbau_test.go` und teilt damit den Aufbau (Ausgangsdaten, Dienste, Ablage); der Rest (*S* = 5360, Marke 7300) bliebe über der Schnittgrenze, der Schnitt löste die Schnittfrage nicht. Die Gruppe der Matrix lässt sich nicht trennen, weil die Folge D das Orakel aus ihr nimmt (Architect, R3). **Erwartete Auslastung unter der Marke:** das Register führt die Überschätzung (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`), und `slice-snapshot-use-case` lag mit Marke 10600 bei 0,53; die Tests brauchten dort rund 62 statt 80 bis 85 Zeilen je Tabellenzeile. Bei 62 je andere Zeile und rund 65 je nebenläufige liegt das Ende bei etwa 5400 Zeilen (Auslastung rund 0,56); die Formel und die Faktoren bleiben unverändert, die Neukalibrierung von `SPEC-128` ist Sache des Users. Überschreitet die Messung die Marke trotzdem, ist es ein Eintrag mit Grund (`SPEC-128` Nr. 5), kein Schnitt beim Handoff (der Teil teilt den Aufbau und ist deshalb kein vorab benannter Schnitt im Sinn von `AGENTS.md` §3.14).

```groessenmarke
pfad: src/bootstrap/abnahme_aufbau_test.go
pfad: src/bootstrap/abnahme_matrix_test.go
pfad: src/bootstrap/abnahme_determinismus_test.go
pfad: src/bootstrap/abnahme_isolation_test.go
pfad: src/bootstrap/abnahme_neustart_test.go
pfad: src/bootstrap/abnahme_nebenlaeufig_test.go
produktcode: 0
fakes: 2
basis: 27acf86f
schaetzung: 7120
marke: 9700
schnitt-grund: Kein Schnitt vor dem Code. Der trennbare Teil Isolation und Nebenlaeufigkeit (T54 bis T60, T62, T71 bis T78) hat S=1760, der Rest S=5360 bliebe ueber der Schnittgrenze (S >= 2223); der Schnitt senkt die Marke des Rests nicht unter sie. Aufbau (Ausgangsdaten, Hilfen der Dienste und der Ablage), Fixture-Tabelle und Folge D teilen alle Teile, auch der Teil liest den Aufbau. Ein Schnitt der Matrix scheitert am Orakel: die Folge D der Determinismus-Messung nimmt Schritte der Matrix (SPEC-049 Nr. 5 b). Der Slice liefert keinen Produktcode, nur Tests ueber einem Mechanismus.
```

**Aufträge an den Architect zum Start** (Planner, 2026-11-07, `main` = `08843060`; vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Architect arbeitet im Zweig `slice-mock-abnahme`; Ausgang jeweils *offen* in §6; Namen M1 bis M7):

- **M1. Fixtures der Matrix `SPEC-049` mit Inhalt.** Für jede Adresse der Muss-Zeilen (`ODV4-GET-01..03`, `-CRUD-01..06`, `-Q-01..12`, `-T-01..08`, `-H-01..03`, `-E-01..04`: 36 Adressen) legt er Request, erwartete Antwort (Status, Header, Body) und Ausgangsdaten fest; dazu die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) mit den Spalten, die `SPEC-128` Nr. 3 liest, und die Mutation je Zeile. Die Soll-Zeilen bleiben draußen (§1).
- **M2. Determinismus.** Was „normalisiert“ ist (`VOLATILE_HEADERS` und weitere flüchtige Felder), was ein Lauf umfasst (Reset, Anfragenfolge, Vergleichsform) und das Zeitbudget der 100 Läufe in `make test` als Randform mit Ort.
- **M3. Isolation.** Welche Beobachtungen „fremder Zustand“ und „fremder Cursor“ heißen (lesen, mutieren, Reset, Snapshot, Restore), die Menge der 20 Workspaces mit Zählprüfung, und ob Cursor im Mock-Modus überhaupt vorkommen (`LH-QA-03.a`).
- **M4. Neustart.** Im Test-Binary oder als Prozess mit `SIGTERM` (`LH-FA-02.c`); was „abgeschlossen“ bei Mutationen heißt und was nach dem Neustart zu sehen ist (Mutationen, Snapshots, Restore). Die Fault-Injection-Matrix (`SPEC-048`) bleibt draußen.
- **M5. Raster und Zusatztabelle (`AGENTS.md` §3.16)** für Reset, Mutation, Snapshot, Restore und Neustart: je Zelle ein Ort oder „gilt nicht“ mit Grund.
- **M6. Nebenläufigkeit.** Fänger der Zusagen über 20 Workspaces mit Wirkungs-Fänger oder benanntem Negativ (`AGENTS.md` §3.10; Register `rate-ohne-detektor-schwankt-mit-der-last`, `zweig-ohne-lauf-unter-dem-detektor`), Zweige mit Zählprüfung unter `-race -coverprofile`, Raten mit Zahl der Prozesse, Last und Messendem.
- **M7. Pfade, Merge-Nähe, Reports.** Pfade des Blocks (§3: `tests/`, `spec/spezifikation.md`; weitere nur mit Meldung an den Planner) und die Faktoren *a*, *b*, *f*, *p* für die Marke; in `spec/spezifikation.md` nur die Stellen von `SPEC-049` und was sie berühren, kein Umstellen daneben. Reports von Reviewer und Verifier tragen ADR-Kennungen nie nackt, sondern als auflösenden Link (Register `report-der-rolle-faellt-docs-check`); der Implementer schreibt Repo-Dateien nur mit `Edit` und `Write` (Register `betriebsregel-fuer-repo-dateien-verletzt`, 13×).

**Bestätigung des Architect zum Start** (2026-11-07, Aufträge M1 bis M7; Vertrag in `SPEC-049` Nr. 1 bis 10 ohne neue `SPEC`-Nummer, denn der Bereich 160 bis 199 der Linie B ist erschöpft und die Kennungsfrage ist beim User offen, Rückgabe R2; Tabelle, Mengen und Raster in diesem Plan; der Ausgang jedes Auftrags steht in §6). Die Prüfstellen sind in der Spezifikation entschieden, sonst hielte der Implementer an (`AGENTS.md` §3.12).

- **M1. Fixtures.** 36 Adressen in sechs Bereichen mit 92 Schritten (`SPEC-049` Nr. 4; 88 beim Start, vier nach dem Review ergänzt, Rückgabe R11), Ausgangsdaten (Nr. 2) und Form des Vergleichs (Nr. 3); die Antworten sind gegen den Stand des Codes auf `9208698c` gemessen (ein Wegwerf-Probelauf der Anfragen gegen den echten Dienst, nicht eingecheckt) und gegen den Wortlaut von `SPEC-169`, `SPEC-193` und `SPEC-197` gelesen; eine Abweichung des Codes von der Spezifikation hat der Probelauf nicht gezeigt. Die Tabelle *Zusage · Fänger* hat 83 Zeilen (T01 bis T83); die Namen in der Spalte *Fänger* sind Aufträge an den Implementer: jedes Symbol entsteht mit seinem Test, er prüft vor dem Handoff mit `grep -rn "func <Symbol>" src`, dass jedes existiert, und zieht die Tabelle im selben Commit nach, wenn ein Name abweicht (`AGENTS.md` §3.9; Register `tabelle-nennt-test-der-nicht-existiert`).
- **M2. Determinismus.** Lauf, Folge D (14 Schritte als Wiederholung von Schritten aus Nr. 4), Orakel, Stand und Cursor, Normalisierung (`SPEC-049` Nr. 5 a bis e): kein Header des Mock ist flüchtig, die Normalisierung nimmt hier nichts aus und steht im Vergleich, damit ein künftiger flüchtiger Header nicht still ausfällt. Zeitbudget (Nr. 5 f): erwartet unter 15 s, Frist 120 s, Fatal; die Gegenproben fahren das Paket nicht.
- **M3. Isolation.** 20 Workspaces, 228 Schritte, Einzelläufe als Orakel, Cursor-Liste je Schritt (Nr. 6). Cursor kommen im Modus `mock` nicht vor, soweit der Mock sie berührt (er liest und schreibt keinen); sie kommen im Workspace vor, weil Replay und Hybrid sie führen, und der Test setzt je Workspace einen eigenen, den Reset, Restore, Snapshot und Mutation anderer nicht verändern dürfen.
- **M4. Neustart.** Im Prozess des Tests, nicht als Prozess mit `SIGTERM` (Nr. 7 a): der Modus `mock` ist im Binary nicht verdrahtet. Abgeschlossen heißt: der Aufruf kehrt ohne Fehler zurück (Nr. 7 b). Die Prozessgrenze ist Rückgabe R4.
- **M5. Raster und Zusatztabelle.** In §6, jede Zelle mit einer Nummer der Spezifikation oder „gilt nicht“ mit Grund.
- **M6. Nebenläufigkeit.** Fünf Läufe der Fenster mit Haltepunkt an `Konfig.Workspaces` (Stelle S2) und an `Konfig.Entities` (T71 bis T75), die Isolation T54 bis T62, der Haltepunkt selbst (T76); Raten `rot in <k> von 20` mit Last und Messendem (Nr. 8); Zweige aus dem Abdeckungsprofil unter `-race -coverprofile` (T78). Akzeptiertes Negativ: die Mischung aus Vorher und Nachher beim Lesen (T82).
- **M7. Pfade, Merge-Nähe, Reports.** Pfade: die sechs Dateien `src/bootstrap/abnahme_*_test.go` (`tests/` ist nach dem Architektur-Gate nicht möglich, Rückgabe R1); `spec/spezifikation.md` steht nicht im Block (`SPEC-128` Nr. 3 m). Faktoren *a* = 16, *b* = 67, *f* = 2, *p* = 0 (siehe *Größenmarke*). Merge-Nähe: Dieser Slice fügt keinen Fehlercode hinzu (`SPEC-049` Nr. 9 c); die Zahl der Codes bleibt unberührt. Die Reports von Reviewer und Verifier tragen ADR-Kennungen als auflösenden Link, und der Implementer schreibt Repo-Dateien nur mit `Edit` und `Write` (Planner-Satz aus M7 unverändert).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Die Marke ist 9700 (Block, Basis `27acf86f`); ein vorab benannter Schnitt entfällt mit Grund (*Marke und Schnittentscheidung*, Rückgabe R3 in §6).
- `in-progress` → `open` (blockiert — Carveout?): die Festlegungen von `SPEC-049` Nr. 1 bis 10 tragen nicht, weil ein Test eine Randform ohne Ort findet (`AGENTS.md` §3.12, siehe §6).

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

- Die Fixture-IDs `ODV4-*` waren „geplante Testadressen“ ohne Inhalt (`SPEC-049`) — **Ausgang:** entfallen. Der Architect hat Inhalt, Zahl (36 Adressen, 88 Schritte, nach dem Review 92: Rückgabe R11) und Ausgangsdaten vor dem ersten Code-Commit in `SPEC-049` Nr. 2 bis 4 festgelegt (M1).
- Laufzeit der 100 Läufe und des Paralleltests in `make test` und die Frist der Gegenproben, die Pakete des Produkts mitfahren (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`) — **Ausgang:** entfallen für die Entscheidung (Zeitbudget als Randform in `SPEC-049` Nr. 5 f: erwartet unter 15 s, Frist 120 s; die Fälle der Gegenproben arbeiten auf Wegwerfmodulen, ihre Frist ist nicht berührt); auch für die Messung (Closure, Planner, 2026-11-08): die gemessene Dauer steht im Absatz *Mutationen und Raten* (§3), 3,3 bis 6,2 s ohne und 8,8 s mit Detektor für die 100 Läufe, die Abnahme-Tests zusammen 26 s, `make gates` im Lauf des Verifiers grün; das Paket braucht rund 150 s, die Grenze der Nullprobe ist Rückgabe R10 (offen, Adresse Architect).
- Raten der nebenläufigen Fänger (20 Workspaces) — **Ausgang:** entfallen (Closure, Planner, 2026-11-08): Reviewer und Verifier haben selbst gemessen (T56, T57, T59x, T75, T76 (a), B4 und T55x, T58a, T60x: je rot in 20 von 20 mit und ohne Detektor). Ungemessen bleibt T54 (Zählung von 19 Workspaces, kein Datenrennen, kein Fremdzustand): akzeptiertes Negativ, Folge im Betrieb: eine Menge unter 20 fiele erst am Zähler des Tests auf. Die Fenster sind mit Haltepunkt deterministisch gebaut (`SPEC-049` Nr. 8); die Raten je Mutation mit Last und Messendem (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`) und das Abdeckungsprofil unter `-race -coverprofile` als Menge der Zweige (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`) stehen im Absatz *Mutationen und Raten* (§3): rot in 20 von 20 Prozessläufen mit und ohne Detektor.
- Der Neustarttest: im Test-Binary oder als Prozess (`SIGTERM`, `LH-FA-02.c`) — **Ausgang:** entfallen. Entschieden in `SPEC-049` Nr. 7 a: Neustart im Prozess des Tests (Dienste und Ablage geschlossen und neu gebaut); der Modus `mock` ist im Binary nicht verdrahtet. Die Prozessgrenze ist Rückgabe R4.
- Aufträge M1 bis M7 an den Architect zum Start (§4) — **Ausgang:** entfallen. Bestätigung des Architect in §4; die Rückgaben R1 bis R4 stehen unten.
- Größe: Marke 9700 (Schätzung 7120, Basis `27acf86f`), kein vorab benannter Schnitt, `schnitt-grund` im Block — **Ausgang:** entfallen für die Setzung (Planner, vor dem ersten Code-Commit, Rückgabe R3); auch für die Messung (Closure, Planner, 2026-11-08; Meldung an den Architect in §7): erwartet wurde eine Auslastung von etwa 0,56 (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`), gemessen am Handoff sind es 2335 von 9700 (rund 0,24); die Messung steht im Absatz *Mutationen und Raten* (§3), §7 bleibt der Closure.
- Gate-Kette (Nr. 7 der Übergangsregeln): die Tests liegen unter `src/bootstrap/`, im Build-Kontext der Stufe `source`; `Makefile`, `Dockerfile`, `.dockerignore` und `harness/mk/` bleiben unberührt — **Ausgang:** entfallen. Der Stempel der Gegenproben gehört dem Integrationsschritt (`SPEC-049` Nr. 9 d).
- Zahl der Codes: ein Fehler, den ein Test findet, kann einen Code verlangen, und `slice-admin-rahmen` und `slice-sqlite-interrupt-transaktion` mergen anders auf `main` — **Ausgang:** entfallen (Closure, Planner, 2026-11-08): der Slice fügt keinen Code hinzu, `model.AlleCodes()` führt nach dem Merge von `main` (mit `ODR-E2102` aus `slice-admin-rahmen`) 40 Codes, und die Zählstellen auf `main` tragen sie; `make gates` im Zweig nach dem Merge belegt es.

**Rückgaben des Architect** (`AGENTS.md` §3.13; Ausgang jeweils *offen*, Adresse genannt):

- **R1. Ablage der Tests in `src/bootstrap/` statt `tests/`** — Ausgang: entfallen (Planner bestätigt am 2026-11-07: Ablage unter `src/bootstrap/abnahme_*_test.go`; §1, §3 und Block sind nachgezogen; eine Änderung der Architekturregel oder eine ADR beginnt dieser Slice nicht). Das Architektur-Gate erlaubt dem Paket `tests` Kanten auf Modell, Inbound-Ports und Driven Adapter, nicht auf die Dienste; die Abnahme braucht `mock`, `workspace` und `snapshot`. Der Composition Root darf sie importieren. Eine Kante `tests` → Dienste wäre eine Änderung der Architekturregel (`AGENTS.md` §3.6) und braucht eine Entscheidung. Folge: die Pfade des Blocks sind sechs Dateien; der Stempel der Gegenproben ändert sich, weil `src` Eingabe ist.
- **R2. Kennungsfrage der Linie B** — Ausgang: offen; Adresse: User (über den Planner). Nichts entschieden: der Vertrag steht als Nr. 1 bis 10 in `SPEC-049` und nimmt keine Nummer aus 160 bis 199 oder aus einem Bereich anderer Linien; ob die Abnahme eine eigene Nummer bekommt, ist Teil der Frage.
- **R3. Größenmarke und Schnitt** — Ausgang: entfallen (Planner, 2026-11-07: Marke 9700, `schnitt-grund` im Block, Begründung in §4 *Marke und Schnittentscheidung*). *S* = 7120, *M* = 9700 (Schnittfrage). Die Aussage in §4, die Gruppe der Matrix lasse sich ohne geteilten Mechanismus von den übrigen trennen, trifft nicht zu: die Folge D der Determinismus-Messung nimmt Schritte der Matrix als Orakel (`SPEC-049` Nr. 5 b), und alle Teile teilen Aufbau und Ausgangsdaten. Ein trennbarer Teil ist **Isolation und Nebenläufigkeit** (T54 bis T60, T62, T71 bis T78; Dateien `abnahme_isolation_test.go` und `abnahme_nebenlaeufig_test.go`): *a* = 16, *f* = 2, *S* = 1760 und *M* = 2400, unter der Schnittgrenze; er nutzt vom Rest nur die Hilfen von `abnahme_aufbau_test.go` (liest sie, ändert sie nicht) und teilt kein Orakel. Der Rest hat *S* = 5360 (*M* 7300) und bleibt über der Grenze, weil Aufbau, Fixture-Tabelle und Folge D gemeinsam sind. Ob vorab so geschnitten wird oder ein `schnitt-grund` in den Block kommt, entscheidet der Planner; der Block erlaubt `pfad` der sechs Dateien für beide Wege. Ein Pfad `spec/spezifikation.md` ist im Block nicht zulässig (`SPEC-128` Nr. 3 m); der Planner-Auftrag M7 nannte ihn.
- **R4. Neustart über die Prozessgrenze im Modus `mock`** (`SIGTERM`, `LH-FA-02.c`, Start aus `main`) — Ausgang: eingetreten für Mutationen (Planner, 2026-11-07: `slice-bootstrap-m3` nimmt die Sendung in §1, DoD, §3 und §6, `main` = `67e20951`, Geber `slice-mock-abnahme`; Szenario in `make test-integration`); Snapshots über die Prozessgrenze bleiben das akzeptierte Negativ aus `SPEC-049` Nr. 10 b bis `welle-admin-api` (es gibt keinen Lesepfad). Erfüllt für Mutationen (Planner, 2026-11-08): geliefert als Stufe `TestE2EMockNeustart` in [slice-bootstrap-m3-e2e-mock](../done/slice-bootstrap-m3-e2e-mock.md) (Teil E des Schnitts von `slice-bootstrap-m3`). Die Folge in Nr. 10 b („ohne Fänger“) gilt damit nur noch für Snapshots; ihren Wortlaut zieht der Architect nach, dieser Slice ändert ihn nicht. Ursprünglicher Text der Rückgabe: Der Architect hat §1 und die DoD von `slice-bootstrap-m3` gelesen: dort steht der M3-Test (Start im Modus `mock`, Anlegen, Abfragen, Mutieren, Snapshot, Restore, Reset), kein Neustart über die Prozessgrenze und kein Ausschluss, der ihn trifft. Der Architect trägt die Sendung nicht selbst in den Plan eines anderen Slice ein (Pläne in `open/` schreibt der Planner); `SPEC-049` Nr. 10 b führt die Lücke als akzeptiertes Negativ mit Grund und Folge. Nimmt `slice-bootstrap-m3` die Zeile (Neustart mit `SIGTERM`: abgeschlossene Mutationen und Snapshots nach dem Start vorhanden, Geber `slice-mock-abnahme`) in §1 oder die DoD, ist die Lücke geschlossen; sonst bleibt sie das Negativ.

**Rückgaben des Implementers** (`AGENTS.md` §3.13; Ausgang jeweils *offen*, Adresse genannt; der Implementer ändert weder die Erwartungen von `SPEC-049` Nr. 4 noch den Wortlaut der Spezifikation):

- **R5. `ODV4-T-06` fängt die Mutation T21 nicht.** Ausgang: entschieden (Architect, 2026-11-08, `SPEC-049` Nr. 2 und Nr. 4). Alle Guid-Werte der Ausgangsdaten und der Fixture `ODV4-T-06` tragen nur Ziffern; die Schreibweise in Großbuchstaben ändert kein Byte der Antwort. Die Mutation „der Schreiber setzt `Edm.Guid` in Großbuchstaben“ wird allein von `ODV4-CRUD-01` rot (Ref `0000000a-0000-0000-0000-00000000000a`); die Zeile T21 nennt `ODV4-T-06` als Fänger. Entscheidung: der Vorschlag gilt mit einer Änderung der Ausgangsdaten: Produkt 3 trägt `Ref` `0000000c-0000-0000-0000-00000000000c` (Tabelle, `P(3)`, Filter und Antwort von `ODV4-T-06`); die Ausgangsdaten bleiben sonst unverändert. Auftrag an den Implementer: in `src/bootstrap/abnahme_aufbau_test.go` die Zeile von Produkt 3 (`"Ref"`) und in `abnahme_matrix_test.go` den Filter von `ODV4-T-06` auf diese Guid setzen; die erwarteten Antworten leiten sich aus den Ausgangsdaten ab. Danach T21 an `ODV4-T-06` allein rot sehen (Mutation: Großbuchstaben nur für die Guid) und die Zeile T21 behalten.
- **R6. Der Schritt 8 von S(k) unterscheidet die Workspaces nicht.** Ausgang: entschieden (Architect, 2026-11-08, `SPEC-049` Nr. 6 b und c, Zeile T61). `GET /Products/$count` ergibt in jedem Workspace 5 (fünf Produkte, plus eine Entity, minus ein Tombstone); die Orakel unterscheiden sich paarweise in den Schritten 2 und 5, nicht im Schritt 8. `TestAbnahmeIsolationOrakelUnterscheidet` prüft diese zwei Schritte; der Wortlaut „die Zahl dieses Workspace“ (Schritt 8) und „Schritte 5 und 8“ (T61) trifft nicht zu. Entscheidung: der Wortlaut wird berichtigt (Schritt 8 ergibt in jedem Workspace `5`, die Orakel unterscheiden sich in den Schritten 2 und 5); der Test bleibt, kein weiterer Auftrag.
- **R7. `SPEC-049` Nr. 10 b.** Ausgang: entschieden (Architect, 2026-11-08, `SPEC-049` Nr. 10 b nachgezogen: Fänger für Mutationen im Test des Composition Root des Modus `mock`, ohne Fänger nur Snapshots). Die Folge „ohne Fänger“ für den Neustart über die Prozessgrenze gilt nach R4 nur noch für Snapshots (Mutationen trägt `slice-bootstrap-m3`); den Wortlaut zieht der Architect nach, der Implementer ändert ihn nicht.
- **R8. Cursors im Test des Neustarts.** Ausgang: entschieden (Architect, 2026-11-08, `SPEC-049` Nr. 7 c und 7 d (4)). Nr. 7 c nennt keinen Cursor, Nr. 7 d (4) sagt „lässt die Cursors unverändert“; ohne einen Cursor prüft das nichts. Der Test schreibt vor den Schritten je einen Cursor `rep-1`/`ses-1` auf Position 3 (`ws-1`) und 2 (`ws-2`); mit ihnen wird die Mutation T68 (b) rot. Entscheidung: die Fixture des Implementers gilt (Position 3 auf `ws-1`, Position 2 auf `ws-2`, `consumed` 1 bis 3 und 1 bis 2, vor den Schritten über den Port geschrieben); Nr. 7 c und 7 d (4) nennen sie.
- **R9. Der Haltepunkt liest zuerst und hält danach; er hält innerhalb der Transaktion nicht.** Ausgang: entschieden (Architect, 2026-11-08, `SPEC-049` Nr. 8). Nr. 8 lässt offen, ob `Get` vor oder nach dem Halten liest. Hält der Dekorator vor dem Lesen, liest der Dienst nach der Freigabe den neuen Stand, und die Mutationen T74 und T75 (Revision oder Workspace aus der Zuordnung statt aus der Transaktion) blieben in 40 von 40 Läufen grün (gemessen). Der Implementer hat die Entscheidung im Test getroffen: `Get` der Workspaces und `Get`, `List`, `ListSet` der Entities lesen zuerst und halten danach, `Put` und `ReplaceAll` halten vor dem Schreiben. Dazu: der Satz „ein Dienst, der außerhalb der Transaktion prüft oder schreibt, hält dort, bis alle *n* eingetroffen sind“ trifft nur einen Zugriff vor `BEGIN` oder nach `COMMIT`; ein Zugriff über `Konfig.Entities` in der Funktion der Transaktion verklemmt gegen die Schreibsperre und endet nach 10 s als `Haltepunkt nicht erreicht` (rot in 20 von 20, aber 7 bis 10 Minuten je Mutation und Reihe). Entscheidung: die Codelage gilt; Nr. 8 sagt zu, dass `Get` (Workspaces) und `Get`, `List`, `ListSet` (Entities) zuerst lesen und danach halten, `Put` und `ReplaceAll` vor dem Schreiben, und dass der Haltepunkt nur vor `BEGIN` und nach `COMMIT` greift. Ein Zugriff in der Funktion der Transaktion endet als „Haltepunkt nicht erreicht“ nach der Frist und ist als Fänger zulässig, mit der Dauer der Frist als Folge. Kein Auftrag.
- **R10. Laufzeit des Pakets `src/bootstrap` und Nullprobe des Werkzeugs `mutationslauf`.** Ausgang: weiter offen; Adresse: Architect (Lesart von `SPEC-210` Nr. 5 a und 9 c, Zeitgrenze der Nullprobe); der Planner entschied am 2026-11-08, dass kein Slice ihn nimmt (kein Produkt-Slice, Durchstich vor Härtung) und führte die Beobachtung `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` im Register. Die Rollen fahren ihre Mutationen weiter mit einem eigenen Runner. Das Paket braucht mit den Abnahme-Tests rund 150 s ohne Detektor (Last 5 bis 13); die Nullprobe des Werkzeugs fährt das Paket mit `-test.timeout=2m` und meldet `Ausgangsstand rot: src/bootstrap`. Das Werkzeug trägt die Mutationen dieses Pakets deshalb nicht; die Messung dieses Slice lief mit einem Runner im Scratchpad. `make test` und `make test-race` setzen keine Grenze unter den 10 Minuten von `go test` (`SPEC-122` Nr. 5).
- **R11. Befunde F1, F6 und F8 des Reviews vom 2026-11-08 (Fixture und Zählung).** Ausgang: entschieden (Architect, 2026-11-08, `SPEC-049` Nr. 4 und 10 h). *F1:* `ODV4-Q-10` hatte keinen Gleichstand im ersten Schlüssel und als zweiten Schlüssel den Schlüssel der Hilfsordnung (`SPEC-191` Nr. 3 d); die Mutation „nur das erste Element von `$orderby`“ blieb grün. Entscheidung: die Ausgangsdaten bleiben (fünf Produkte, alle Zählungen und Orakel unberührt); `ODV4-Q-10` bekommt den Schritt b `GET /Products?$select=Id&$orderby=Active,Id desc` mit der Folge 5, 2, 4, 3, 1 (`Active`: `null` bei 5, `false` bei 2, `true` bei 1, 3, 4; die drei sind im ersten Schlüssel gleich, das zweite Element ordnet sie absteigend, die Hilfsordnung aufsteigend ergäbe 1, 3, 4). Die Mutationen „nur das erste Element“ (5, 2, 1, 3, 4), „das zweite Element ordnet aufsteigend“ (5, 2, 1, 3, 4) und „alle Elemente in der Richtung des ersten“ (5, 2, 1, 3, 4) werden rot. Der Schritt a bleibt unverändert. *F6:* Entscheidung (a), die zwei Zellen von `SPEC-193` Nr. 2 sind Fixtures: `ODV4-CRUD-03` c und d (`PUT /Orders(1)` ergibt `204`, danach `GET /Orders(1)` `200`) und `ODV4-CRUD-04` f (`POST` über den Tombstone von `Products(3)` ergibt `201`); die Matrix führt elf von 20 Zellen, die übrigen neun bleiben dem Test des Dienstes, als akzeptiertes Negativ mit Grund und Folge in `SPEC-049` Nr. 10 h. *F8:* die Kategorienliste der 71 Blöcke im Absatz *Abdeckungsprofil* nennt den Typ des Eintrags (`ODR-E6007`, `SPEC-193` Nr. 2 a). *Folgen der Zählung:* 36 IDs bleiben, Schritte 92 statt 88 (GET 7, CRUD 29, Q 16, T 9, H 13, E 18); Revision von `ODV4-CRUD-03` und `-04` je 2 statt 1; die Folge D der Determinismus-Messung (nutzt `CRUD-04` a bis d und `E-04` b), die Isolation (228 Schritte aus S(k)), die Neustart-Zeilen und die Marke 9700 (Zeilen der Tabelle: 83) bleiben. Aufträge an den Implementer (`src/bootstrap/abnahme_matrix_test.go`): (1) in `abFixtures` hinter dem Schritt von `ODV4-Q-10` den Schritt `b` `abG("/Products", "$select=Id&$orderby=Active,Id desc")` mit `abKK(abAuswahl("Id", 5, 2, 4, 3, 1))` und den bisherigen Schritt auf den Namen `a`; (2) `ODV4-CRUD-03`: `Revision: 2`, Schritt `c` `abM("PUT", "/Orders(1)", `{"Id":1,"Note":"n"}`)` mit `abK0()`, Schritt `d` `abG("/Orders(1)", "")` mit `abKE` des Bodys `{"@odata.context":"{W}/$metadata#Orders/$entity","@odata.id":"{W}/Orders(1)","Id":1,"Note":"n"}`; (3) `ODV4-CRUD-04`: `Revision: 2`, Schritt `f` `abM("POST", "/Products", `{"Id":3,"Name":"Chain2"}`)` mit `abKL("{W}/Products(3)", …)` und dem Body `{"@odata.context":"{W}/$metadata#Products/$entity","@odata.id":"{W}/Products(3)","Id":3,"Name":"Chain2","Active":null,"Stock":null,"Price":null,"Ref":null,"Released":null,"Updated":null}`; (4) die Zahl 88 in `abPruefeZaehlung` (Zeilen um 319), in den Kommentaren (Zeilen 22, 384, 385) und die Summe der Bereiche auf 92 (GET 7, CRUD 29, Q 16, T 9, H 13, E 18); `abErwarteteIDs` bleibt; (5) Mutationen rot sehen und im Absatz *Mutationen und Raten* eintragen: „nur das erste Element von `$orderby`“ (rot `Q-10` b), „das zweite Element ordnet aufsteigend“ (`Q-10` b), `PUT` an `partial` ergibt `ODR-E6003` (`CRUD-03` c), der Eintrag bleibt nach `PUT` `partial` (`CRUD-03` d), `POST` über Tombstone ergibt `ODR-E6009` (`CRUD-04` f), der neue Eintrag behält `Ref` des gelöschten (`CRUD-04` f; Produkt 3 trägt im Ausgangszustand eine `Ref`); (6) das Abdeckungsprofil neu messen, denn `PUT` an `partial` und `POST` über Tombstone treffen Blöcke von `mock/schreiben.go`, die bisher ohne Treffer waren, und die Zahlen 213, 142, 71 im Absatz *Abdeckungsprofil* nachziehen.
- **R12. `ODV4-CRUD-03` d fängt „der Eintrag bleibt nach `PUT` `partial`“ nicht.** Ausgang: entschieden (Architect, 2026-11-08, `SPEC-049` Nr. 10 i): akzeptiertes Negativ, kein neuer Schritt, kein Test-Auftrag; 92 Schritte, Revisionen und Bereichssummen bleiben. Die Mutation in der gemeinsamen Anlage `legeAn` fängt `ODV4-H-02` f (rot gemessen, siehe oben); eine Mutation allein im `PUT`-Pfad (`entscheide`, Zweig `methodePut`) fängt nur `TestPutPartial` (`SPEC-192`, W37) im Test des Dienstes. Grund: der Schritt `PATCH /Orders(1)` fängt nur diese eine Mutation und kostet Schritt, Revision und Summen; die Zusage hat ihren Fänger im Test ihres Vertrags. Folge: der Pfad von `PUT` allein ist an der Matrix ungefangen. Die Erwartung ist nicht geändert.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Orte sind Nummern von `SPEC-129`, `SPEC-192` bis `SPEC-198`, `SPEC-103` und `SPEC-049`):

| Operation | fehlend oder leer | unbekannt | fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `Reset` | Kennung leer: `SPEC-129` Nr. 10 (b), unverändert an den Port, `ODR-E7001` | `SPEC-129` Nr. 10 (b), `ODR-E7001` | gilt nicht: der Aufruf nennt eine Kennung, keinen Aufrufer; die Wirkung auf fremde State und Cursors: `SPEC-129` Nr. 10 *Wirkung*, `SPEC-049` Nr. 6 e | `SPEC-129` Nr. 2 und Nr. 10 (d): Zustand unverändert | `SPEC-129` Nr. 10 *Revision* (wächst bei jedem Reset), `SPEC-049` Nr. 5 a |
| Mutation (`POST`, `PATCH`, `PUT`, `DELETE`) | Header `X-Recorder-Workspace`: `SPEC-195` Nr. 5; Body: `SPEC-192` Nr. 5 und `SPEC-181` Nr. 2 a; Schlüssel der URL: `SPEC-196` Nr. 7 a; `null` und nicht genannt: `SPEC-193` Nr. 3, `SPEC-049` Nr. 4 (`ODV4-CRUD-03`, `-05`) | Workspace: `SPEC-195` Nr. 8; Set: `SPEC-192` Nr. 4; Eintrag: `SPEC-193` Nr. 2; Property und Typ im Body: `SPEC-181` Nr. 3 bis 5 | Workspace eines anderen Service: `SPEC-195` Nr. 8 (`ODR-E5004`); Mutation eines Workspace gegen die anderen: `SPEC-192` Nr. 8 (d), `SPEC-049` Nr. 6 | `SPEC-195` Nr. 2 und `SPEC-192` Nr. 7, `SPEC-049` Nr. 7 b | `SPEC-193` Nr. 2 (`POST` doppelt, `DELETE` doppelt), `SPEC-194` Nr. 5, gleichzeitig `SPEC-192` Nr. 8 (a), `SPEC-049` Nr. 8 |
| Snapshot erzeugen | Name leer: `SPEC-129` Nr. 21; Workspace leer: Nr. 22 (c) | Workspace: Nr. 22 (c) und Nr. 24, `ODR-E7001` | Name eines anderen Workspace ist frei: Nr. 22 (i); `ErrFremderService`: Nr. 24 (unerreichbar); `SPEC-049` Nr. 6 b Schritt 6 | Nr. 21 (Kontextfehler ohne Code), Nr. 25 (Blob ohne Referenz) | Name vergeben `ODR-E7006`: Nr. 22 (i) |
| Snapshot auflisten | Kennung leer: Nr. 23 (unverändert an den Port) | Nr. 23, `ODR-E7001` | Nr. 23: nie der Snapshot eines anderen Workspace; `SPEC-049` Nr. 6 b Schritt 13 | Nr. 21 | gilt nicht: lesend, eine Wiederholung ändert nichts (Nr. 23) |
| `Restore` | Kennung des Workspace oder des Snapshots leer: `SPEC-129` Nr. 16 und 18 (Rangfolge) | Nr. 18 (`ODR-E7001`, `ODR-E7002`) | Snapshot eines anderen Workspace `ODR-E7002`, Baseline eines anderen Service `ODR-E5004`: Nr. 18 | Nr. 16 (Zustand unverändert) und Nr. 18 (Kontextfehler geht jeder Zeile vor) | Nr. 17 (Revision plus 1, auch bei gleichem State) |
| Neustart | Verzeichnis fehlt: `SPEC-103` Nr. 7 (nicht Gegenstand von Nr. 7, der Test öffnet das vorhandene) | Schemaversion neuer oder Tabellen ohne Version: `SPEC-103` Nr. 4 und 5 (nicht Gegenstand) | fremde Datei: `SPEC-103` Nr. 6 (nicht Gegenstand) | gilt nicht: der reguläre Neustart wird nicht abgebrochen; Abbruch und Absturz sind `SPEC-048` (`SPEC-049` Nr. 10 c) | zweiter Neustart: `SPEC-049` Nr. 7 d (6), `SPEC-103` Nr. 3 und 9 |

**Zusatztabelle Z** (Verstöße am selben Schlüssel und Modus von `STRICT_COVERAGE`, `AGENTS.md` §3.16; Ort `SPEC-197` Nr. 2 und 3, `SPEC-193` Nr. 2, `SPEC-049` Nr. 4):

| Nr. | Operation | Lage des Schlüssels | `StrictCoverage` | Ergebnis | Fänger |
|---|---|---|---|---|---|
| Z1 | Collection | `Orders`, Coverage `partial` | falsch | 200, `X-Recorder-Coverage: partial` | `ODV4-E-04` b (T30) |
| Z2 | Collection | `Orders`, `partial` | wahr | `ODR-E6003` | `ODV4-E-04` a (T30) |
| Z3 | Anzahl | `Orders`, `partial` | wahr | `ODR-E6003` | `ODV4-E-04` c (T30) |
| Z4 | Collection | `Products`, `complete` | wahr | 200, `complete` | `ODV4-E-04` d (T30) |
| Z5 | Lookup ohne Treffer | `Orders`, `partial` | falsch | `ODR-E6003`, kein `404` | `ODV4-E-04` e (T30) |
| Z6 | Lookup ohne Treffer | `Orders`, `partial` | wahr | `ODR-E6003` | `ODV4-E-04` f (T30) |
| Z7 | Lookup, Eintrag `partial` | `Orders(1)` kennt `Note` nicht | falsch | `ODR-E6003` | `ODV4-E-04` g (T30) |
| Z8 | `PATCH`, Eintrag fehlt | `Orders(9)`, `partial` | falsch | `ODR-E6003` | `ODV4-CRUD-06` e (T36) |
| Z9 | `PATCH`, Eintrag vorhanden | `Orders(1)`, `partial` | falsch | `ODR-E6003` | `ODV4-CRUD-06` f (T36) |
| Z10 | `PATCH`, Eintrag fehlt | `Products(99)`, `complete` | falsch | `ODR-E6008` | `ODV4-CRUD-06` d (T36) |
| Z11 | `PATCH`, Eintrag fehlt | `Orders(9)`, `partial` | wahr | `ODR-E6003`, wie Z8 (kein Einfluss des Modus) | `ODV4-E-04` h (T30) |
| Z12 | `PATCH`, Eintrag fehlt | `Products(99)`, `complete` | wahr | `ODR-E6008`, wie Z10 | `ODV4-E-04` i (T30) |
| Z13 | `PUT`, Eintrag vorhanden | `Orders(1)`, `partial` | falsch | 204, danach `GET` 200 (der Zustand `complete` ist an der Matrix nicht beobachtbar, `SPEC-049` Nr. 10 i; Fänger der Zusage „danach `complete`“: `TestPutPartial` im Test des Dienstes) | `ODV4-CRUD-03` c und d (T33) für 204 und `GET` 200 |
| Z14 | `POST`, Tombstone | `Products(3)` nach `DELETE` | falsch | 201, legt neu an, kein `ODR-E6009` | `ODV4-CRUD-04` f (T34) |

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

**Stand der Closure** (Planner, Integrationsschritt, 2026-11-08; `main` = `cd01cba2` nach dem Merge von `slice-admin-rahmen`, in den Zweig gemerged; das reine `git mv` nach `done/` folgt als eigener Commit, danach der Merge nach `main`, kein Push). Belege: Architect vor dem Code `96bdbbc2`, Marke `7d6e18f9`, Tests `d63b4755` bis `0daee175`, Review-Report `docs/reviews/2026-11-08-slice-mock-abnahme-review.md` (`0ecdb6e4`: kein HIGH, ein MEDIUM F1, drei LOW, sechs INFO), Entscheidungen des Architect `f1166c8d` (R11) und `73a1f341` (R12), Fix-Runde `267bdc89`, Messung `028ba949`, Verifikation `docs/reviews/2026-11-08-slice-mock-abnahme-verifikation.md` (`c37f47c7`: bestanden mit Auflagen, V1 MEDIUM Plan-Widerspruch zu R12, V2 LOW Frist-Wächter, V3 INFO Closure). V1 ist im Plan in einem Commit nachgezogen (`c8d6a9f1`: T33, `T32 bis T35`, T83, Z13), V2 als akzeptiertes Negativ im Absatz nach *Mutationen und Raten* (§3). Es gab keinen weiteren Implementer-Lauf nach der Verifikation (Durchstich vor Härtung); der Nachzug ist **nicht nachverifiziert**.

**Gegenstand:** geliefert. Die Abnahmetests des Mock liegen in sechs Dateien `src/bootstrap/abnahme_{aufbau,matrix,determinismus,isolation,neustart,nebenlaeufig}_test.go` (kein Produktcode, `Makefile`, `Dockerfile` und `harness/mk/` unberührt): `TestAbnahmeMatrix` mit 36 Fixtures `ODV4-*` und 92 Schritten (`SPEC-049` Nr. 2 bis 4, Zählung durch `TestAbnahmeMatrixZaehlung`), `TestAbnahmeDeterminismus` (100 Läufe nach Reset, Frist 120 s, Fänger T44 bis T53), `TestAbnahmeIsolation` (20 Workspaces, 228 Schritte, Cursors je Workspace), `TestAbnahmeNeustart` (Dienste und Ablage im Prozess des Tests geschlossen und neu gebaut) und die Fenster der Nebenläufigkeit mit Haltepunkt (`TestAbnahmeNebenlaeufig*`, `TestAbnahmeHaltepunkt`). Die Spezifikation `SPEC-049` Nr. 1 bis 10 schrieb der Architect vor dem Code.

**Größen** (`SPEC-128` Nr. 9): `Messpunkt slice-mock-abnahme.md: zeilen=83 nebenlaeufig=16 fakes=2 produktcode=0 schaetzung=7120 marke=9700 messung=2335`, Stand `Ende`. Auslastung 2335 ÷ 9700 = 0,24, **unter 0,60: Meldung an den Architect** nach `SPEC-128` Nr. 9 (Neurechnung der Faktoren aus allen Messpunkten, Nr. 2): vierte Beobachtung zu `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` in Folge, hier mit *p* = 0 und der Schätzung 7120 gegen eine Messung von 2335 (Faktor 3). Die Marke wurde vor dem ersten Code-Commit gesetzt und bleibt; keine Rückführung, die Messung liegt unter ihr.

**Mutationen:** der Implementer je Zeile der Tabelle (Fix-Runde mit Mengen und Elementzahlen in §3), der Reviewer rund zwei Dutzend eigene (A1 bis B4, T56, T57, T59x, T75, T76a, davon A1, B1 bis B3 grün und Befunde), der Verifier die Klasse der Fix-Runde (m1 bis m10, f1 bis f5; grün m7, f1, f2, beide Negative) und die Raten T55x, T58a, T60x (je 20 von 20 mit und ohne Detektor, Last 4 bis 8, Messender der Verifier-Agent). Raten und Messende der übrigen Zeilen: *Mutationen und Raten* in §3.

**Gate:** `make gates` im Zweig nach dem Merge von `main` und den Closure-Dateien, unter `flock /tmp/odata-recorder-heavy.lock`: Ergebnis im Bericht des Integrationsschritts (Zweig vor dem Merge: Lauf des Verifiers, `EXIT=0`, `d-check` 855 Dateien, `pruefung test` und `pruefung race` ok, `zeichen` ok). Die Größenmarke meldet 2335 von 9700 (ok).

**Akzeptierte Negative:** T79 bis T83 (Grund und Folge in `SPEC-049` Nr. 10 a bis i, darunter neu i: „der Eintrag bleibt nach `PUT` `partial`“, Ersatz-Fänger `TestPutPartial`), V2 (Obergrenze des Frist-Wächters zwischen 120 und 125 s), T54 (19 Workspaces, eine Zählung), Snapshots über die Prozessgrenze (Nr. 10 b) und die neun übrigen Zellen von `SPEC-193` Nr. 2 (Nr. 10 h), je mit Folge im Betrieb.

**Risiko-Ausgänge (§6):** je Punkt im Plan selbst. Entfallen oder eingetreten: Fixture-IDs (M1), Laufzeit, Raten, Neustarttest, Aufträge M1 bis M7, Größe, Gate-Kette, Zahl der Codes (40, kein neuer Code), R1, R3 bis R9, R11, R12 (R4 eingetreten bei `slice-bootstrap-m3`). Weiter offen mit Adresse: R2 (Kennungsfrage der Linie B, Adresse User) und R10 (Zeitgrenze der Nullprobe von `mutationslauf` gegen die Laufzeit des Pakets `src/bootstrap`, Adresse Architect, Register `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`).

**Steering-Loop-Einträge:**

1. **Schätzformel:** Auslastung 0,24, Produktcode 0, Schätzung 7120 gegen 2335. Evidenz zu `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`; Meldung an den Architect (Neukalibrierung, Klasse „nur Tests“ gesondert, Messpunkte nach Klassen getrennt).
2. **Eine Entscheidung des Architect erreicht nicht jede Stelle des Plans:** R12 stand in §6, T33, `T32 bis T35`, T83 und Z13 trugen weiter die alte Aussage (Verifikation V1, MEDIUM; die Zusage hatte in der Tabelle einen Fänger, den es nicht gibt). Evidenz zu `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (dort auch V2). Die Regel steht in `AGENTS.md` §3.9; kein neuer Anker.
3. **Eine Fixture ohne Gleichstand fängt die Klasse der Zusage nicht** (F1: `$orderby` mit einem Schlüssel, R5: Guid nur aus Ziffern). Evidenz zu `BEO-REPO/fixture-gleicht-dem-wert-der-mutation`; die Fix-Runde deckte die Klasse (Verifier: Mutationen m3 und m10 an nicht genannten Elementen rot).
4. **Die Nullprobe des Mutationslaufs hat eine feste Zeitgrenze unter der Laufzeit des Pakets.** Neue Beobachtung `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` (1×), Adresse Architect.
5. **Betriebsregel:** der Planner setzte die zehn Haken der DoD dieses Plans mit `sed -i` statt mit dem Edit-Werkzeug (Diff geprüft: zehn Zeilen, nichts daneben) und rief einmal `python3` auf (vom Command-Guard abgewiesen, keine Wirkung). Evidenz zu `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`. Die Berichte von Implementer, Reviewer und Verifier melden keine Verletzung; der Verifier fuhr Mutationen mit `perl -0pi` nur in Wegwerfkopien im Container.

**Spec-Lücken:** benannt, nicht geschlossen: `SPEC-049` trägt keine eigene Nummer im Bereich der Linie B; die Kennungsfrage (R2) liegt beim User. Geschlossen vom Architect: Nr. 10 b (Neustart über die Prozessgrenze nur für Snapshots ungefangen, R7) und Nr. 10 i (R12).

**Kein Feld `liegt in` in diesem Slice:** nichts wurde in einer Regel, einem Sensor oder einer Spec-Stelle als neue Regel verkörpert, deren Zielort dieser Slice trägt; `SPEC-049` Nr. 1 bis 10 schrieb der Architect (benannte Spec-Lücke ohne `liegt in`). Eine Pflicht zur Anker-Paarung entsteht nicht.

**Register:** Evidenz `slice-mock-abnahme` zu `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (Eintrag 1), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (Eintrag 2), `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` (Eintrag 3), `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (Eintrag 5) und neu `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` (Eintrag 4). Keine neue Evidenz zu `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (Raten bestätigen den Stand).

**Drei Paarungen:** Anker (kein Eintrag trägt `liegt in`, nichts zu paaren) · Folge-Slice (`slice-bootstrap-m3` trägt die Sendung des Neustarts über die Prozessgrenze für Mutationen in §1, DoD und §6, den Start-Trigger „`slice-mock-abnahme` in `done/`“ und neu den offenen Punkt der Ablage des M3-Tests mit Adresse Architect in §6; `slice-batch-changesets` braucht `welle-mock` in `done/`; R2 hat als Adresse den User, R10 den Architect, beide ohne Nehmer-Slice) · Register (die fünf Verzeichnisse liegen mit `evidence/slice-mock-abnahme.md` im Register, das fünfte ist neu).

**Start-Trigger der Folger:** [slice-bootstrap-m3](../done/slice-bootstrap-m3.md) hat als Bedingung, dass `slice-mock-abnahme`, `slice-snapshot-use-case`, `slice-test-integration` und `slice-odata-url-filter` in `done/` liegen; sie tritt mit dem Merge nach `main` ein. Der Spezifikations-Nachzug von `SPEC-110` (Architect) ist Bedingung für den ersten Code-Commit, nicht für den Start.

**Hinweise an den User (nicht entschieden):** (1) Kennungsfrage R2 der Linie B. (2) R10: die Nullprobe von `mutationslauf` trägt `src/bootstrap` nicht; eine Entscheidung über die Zeitgrenze liegt beim Architect, ein Slice dafür gibt es nicht.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-07, `main` = `08843060`, Zähler aus den Dateien unter `evidence/`). Einschlägig: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (8×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (16×) für den Paralleltest, `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (5×) für die Laufzeit der Läufe in der Gate-Kette, `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (13×), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (48×), `BEO-REPO/randform-erst-im-review-entschieden` (34×), `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (26×) und `BEO-REPO/report-der-rolle-faellt-docs-check` (5×). Verkörpert: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×, `AGENTS.md` §3.10). Das Register führt sie als offen oder verkörpert nach eigenem Stand; dieser Start ändert keinen Zustand.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
