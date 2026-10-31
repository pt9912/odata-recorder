# Welle welle-mock: Mock-Modus mit Metadaten, zustandsbehaftetem CRUD, Abfrageoptionen und Snapshots

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. Diese Welle ist vor ihrem Start angelegt
(Entscheidung des Users vom 2026-10-30) und startet gestaffelt (§2): Der Trigger von
Teil 1 ist eingetreten, sie steht in der Roadmap unter *Offene Wellen*. Ihre
Slice-Dateien liegen in `open/`; sie bleiben, wie sie sind, bis der jeweilige Slice
vor dem Code steht (Auftrag in §4).

**Zielmeilenstein:** M3 (siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Ein Test legt aus einer importierten Baseline einen Workspace an und beantwortet im Modus `mock` ohne Upstream `$metadata`, Collection- und Entity-Abfragen mit `$filter`, `$select`, `$orderby`, `$top`, `$skip` und `$count`, wendet POST, PATCH, PUT und DELETE auf den Workspace an, sieht die Zustandsübergänge in folgenden GET-Anfragen, erzeugt und stellt Snapshots wieder her (Replay-Cursor bleiben unverändert) und setzt den Workspace zurück (`LH-FA-06`, `LH-FA-07`, `LH-FA-08`, `LH-FA-09`, `LH-FA-10`, `LH-FA-11`).
Die Welle baut auf dem, was `welle-record-replay` liefert: Baseline mit Coverage-Beweis, Workspace, den Codec zum Lesen, die Services und den HTTP-Adapter. Sie ergänzt den Parser der OData-URL, den Query-Interpreter, den Codec zum Schreiben von OData-JSON und zum Lesen von Client-Bodies, `MockUseCase` und `SnapshotUseCase`.
Eine Welle, weil der Closure-Trigger mehr beobachtet als die DoDs der Slices: die Abnahmematrix der Spezifikation (`SPEC-049`), die Messung von Determinismus und Isolation (`LH-QA-02`, `LH-QA-03`) und den Rundlauf durch alle Schichten im Modus `mock`.
Zwei Stellen der Spezifikation sind vor dem Start von `slice-bootstrap-m3` nachzuziehen (Aufgabe des Architect): die Startkonfiguration lehnt den Modus `mock` heute ab, auch als Default (`SPEC-110` Nr. 8), und kennt `STRICT_COVERAGE` und `PUBLIC_BASE_URL` nicht (Nr. 2).

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

Die Welle startet gestaffelt, in zwei Teilen mit je eigenem Trigger (Umplanung vom 2026-10-30, Drift-Log der Roadmap; Entscheidung des Users zu Frage 7 des [Abgleichs](../../reviews/2026-10-30-lastenheft-wellen-abgleich.md), mit den Übergangsregeln des Parallelbetriebs). Sie bleibt **eine** Welle, weil ihr Closure-Trigger (§3) den Rundlauf durch alle Schichten im Modus `mock`, die Abnahmematrix und die Messung von Determinismus und Isolation beobachtet; zwei Wellen trügen zwei Closure-Notizen für ein Ergebnis, M3.

- **Teil 1** (`slice-odata-url-pfad`, `slice-odata-url-filter`, `slice-edm-ausdruecke`, `slice-query-interpreter`, `slice-odata-json-schreiben`, `slice-odata-json-eingabe`): `slice-odata-codec-csdl`, `slice-odata-codec-json` und `slice-baseline-import` liegen in `done/` (Modell, Codec zum Lesen, Baseline) — eingetreten. Grund: Die sechs Slices sind reine Funktionen und Codecs auf dem Modell; sie haben keine Port-Anbindung, keinen Composition Root und kein sichtbares Verhalten am Binary. **Nicht** Trigger von Teil 1 sind `welle-record-replay` in `done/` und das Handbuch-Gerüst: Die Handbuch-Pflicht (§4) gilt nur für Slices mit sichtbarem Verhalten, diese sechs tragen sie nicht (die Fehlercodes `ODR-E6001` bis `ODR-E6003` stehen im Binary, bevor sie beginnen). Teil 1 läuft als Linie B in einem eigenen Arbeitsbaum, höchstens ein Slice zugleich ([Übergangsregeln](README.md#parallelbetrieb-in-worktrees-übergangsregeln)). Geliefert: [slice-odata-url-pfad](done/slice-odata-url-pfad.md) (Linie B, `SPEC-160` bis `SPEC-163`, Code `ODR-E6005`, die Tabelle der Codes führt 34 Einträge); die übrigen fünf Slices von Teil 1 sind nicht gestartet.
- **Teil 2** (`slice-mock-lesen`, `slice-mock-schreiben`, `slice-snapshot-use-case`, `slice-mock-abnahme`, `slice-bootstrap-m3`), Trigger unverändert:
  - `welle-record-replay` liegt in `done/`: Workspace, Replay, Import, Codec zum Lesen, HTTP-Adapter und Composition Root sind die Grundlage von `MockUseCase`.
  - M2 ist erreicht: der M2-Test aus `slice-bootstrap-m2` ist grün und Teil von `make test`.
  - `slice-handbuch-meldungskatalog` (`welle-benutzerhandbuch`) liegt in `done/`: Die Handbuch-Pflicht aus §4 und das Closure-Kriterium aus §3 setzen das Handbuch-Gerüst und das Kataloge-Gate voraus. Der Slice startet erst nach `slice-bootstrap-m2`, liegt also nicht automatisch vor dem Trigger oben; ohne ihn könnte kein Slice von Teil 2 seine DoD erfüllen.

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`; die Welle schließt erst nach Teil 2, auch wenn Teil 1 früher fertig ist (ein Closure-Trigger, eine Notiz).
- `make gates` ist grün (einschließlich `make test-race` über die neuen Pakete); `make gegenproben` ist grün.
- Der M3-Test aus `slice-bootstrap-m3` ist grün und Teil von `make test`: Er startet den Recorder im Modus `mock` mit einer Baseline aus einem Import, legt einen Workspace an, fragt `$metadata` und eine kombinierte Abfrage mit `$filter`, `$orderby`, `$top`, `$skip` und `$count` ab, legt eine Entity an, ändert und löscht sie und liest nach jedem Schritt den neuen Zustand, erzeugt einen Snapshot, ändert weiter, stellt wieder her und setzt zurück; der Referenzdienst, aus dem die Baseline stammt, ist beendet, und es entstand keine Verbindung zu ihm. Eine Collection-Abfrage auf nicht vollständigem Bestand ergibt im strikten Modus `ODR-E6003`, ohne ihn trägt die Antwort `X-Recorder-Coverage`.
- Die Abnahmetests aus `slice-mock-abnahme` sind grün und Teil von `make test`: die Muss-Zeilen der Matrix `SPEC-049` (`ODV4-*`), 100 identische Läufe nach Reset mit gleichen normalisierten Antworten (`LH-QA-02`), 20 parallele Workspaces ohne Beobachtung fremden Zustands und ein Neustart, nach dem Mutationen und Snapshots vorhanden sind (`LH-QA-03`).
- Das Handbuch deckt alles ab, was das Binary dieser Welle kann (Modus `mock`, Schlüssel, Fehlercodes); das Kataloge-Gate aus `slice-handbuch-meldungskatalog` ist grün.
- Die Closure-Notiz liegt als `welle-mock-results.md` im Ruheort `done/`; ihre *Steering-Loop-Einträge* tragen je verkörperter Regel mit dem Anker `seit welle-mock` das Feld `liegt in`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-odata-url-pfad | Parser der OData-URL, Teil Ressourcenpfad, Key-Syntax und Systemoptionen: Pfad und Optionen werden zum Query-AST des Domain-Modells (Leser fremder Bytes) | `LH-FA-08`, `LH-FA-07`, `LH-RB-01`; `LH-FA-08.a`, `LH-FA-07.b` |
| slice-odata-url-filter | Parser der OData-URL, Teil `$filter`: Grammatik der Muss-Untermenge, typisierte Literale; `400` für ungültige, `501` mit `ODR-E6001` für gültige, nicht unterstützte Ausdrücke | `LH-FA-08`, `LH-QA-05`; `LH-FA-08.a`, `SPEC-022` |
| slice-edm-ausdruecke | Typregeln der acht EDM-Typen und dreiwertige Null-Semantik als reine Funktionen im Domain-Modell: Vergleich, Sortierung, Typprüfung des AST gegen das Modell | `LH-FA-08`, `LH-QA-02`; `LH-FA-08.b`, `SPEC-049` |
| slice-query-interpreter | Query-Interpreter: Filter, stabile Sortierung nach Key, Count vor Paging, Skip und Top, Projektion mit Property-Maske | `LH-FA-08`, `LH-FA-11`, `LH-QA-02`; `LH-FA-08.a`, `SPEC-044` |
| slice-odata-json-schreiben | Codec zum Schreiben von OData-JSON: Collection, Entity, `@odata.context`, `@odata.count`, Fehlerobjekt im OData-V4-Format | `LH-FA-07`, `LH-FA-08`, `LH-RB-01`; `LH-FA-07.b`, `SPEC-006`, `SPEC-126` |
| slice-odata-json-eingabe | Codec zum Lesen von Client-Bodies (POST, PATCH, PUT): Entity-Eingabe mit Typprüfung, fehlend und `null` bleiben unterscheidbar (Leser fremder Bytes) | `LH-FA-07`, `LH-RB-01`; `LH-FA-07.a`, `SPEC-044` |
| slice-mock-lesen | `MockUseCase`, Teil Lesen: `$metadata`, Collection, Key-Lookup, `/$count`; Workspace-Zuordnung, Coverage-Prüfung vor der Abfrage, Header `X-Recorder-Coverage` | `LH-FA-06`, `LH-FA-08`, `LH-FA-11`; `LH-FA-06.a`, `LH-FA-11.a`, `SPEC-044` |
| slice-mock-schreiben | `MockUseCase`, Teil Schreiben: POST, PATCH, PUT, DELETE in der Workspace-Transaktion, Validierung, Statuscodes der Vertragsmatrix, Teilprojektionen | `LH-FA-07`, `LH-FA-09`, `LH-QA-03`; `LH-FA-07.a`, `LH-FA-07.b`, `SPEC-044` |
| slice-snapshot-use-case | `SnapshotUseCase` (Erzeugen, Auflisten) und `Restore` im `WorkspaceUseCase`; Replay-Cursor bleiben unverändert | `LH-FA-10`, `LH-FA-09`; `LH-FA-10.a`, `SPEC-108` |
| slice-mock-abnahme | Abnahmetests des Mock: Golden-Fixtures der Matrix `SPEC-049`, 100 Läufe nach Reset, 20 parallele Workspaces, Neustart | `LH-QA-02`, `LH-QA-03`, `LH-RB-01`; `LH-QA-02.a`, `LH-QA-03.a`, `SPEC-049` |
| slice-bootstrap-m3 | Composition Root für den Modus `mock`, Konfiguration (`STRICT_COVERAGE`, `PUBLIC_BASE_URL`), Mock-Routen im HTTP-Adapter, M3-Test | `LH-FA-12`, `LH-FA-06`, `LH-FA-07`; `SPEC-110`, `SPEC-055`, `SPEC-011` |

**Auftrag beim Start eines Slice aus `open/`:** Die Pläne in `open/` sind angelegt und bleiben, wie sie sind, bis der jeweilige Slice vor dem Code steht. Vor dem Übergang `next` → `in-progress` entsteht der Vertrag samt Tabelle *Zusage · Fänger* und Raster (`AGENTS.md` §3.15, §3.16) in der Spezifikation durch den Architect; vor dem ersten Code-Commit setzt der Planner die Marke nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` und zieht die Sichtung in §8 des Plans nach. Leser fremder Bytes sind `slice-odata-url-pfad`, `slice-odata-url-filter` und `slice-odata-json-eingabe`; die Formel nach `SPEC-128` kennt dafür keinen Zuschlag mehr, die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung.

**Pflicht Handbuch:** Jeder Slice, dessen Ergebnis ein Anwender am Binary sieht (neues Kommando, Konfigurationsschlüssel, Fehlercode, HTTP-Schnittstelle, Betriebsart), trägt in seiner DoD den Punkt „Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün“. In dieser Welle gilt er verbindlich für `slice-bootstrap-m3` (der Modus `mock` und seine Schlüssel erreichen das Binary erst dort) und bedingt für `slice-mock-lesen`, `slice-mock-schreiben` und `slice-snapshot-use-case`, sofern sie einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary bringen. Die Parser, Ausdrücke, der Interpreter, die Codecs und `slice-mock-abnahme` ändern das Binary nicht und tragen ihn nicht. Das ist der Grund, warum die sechs Slices von Teil 1 vor dem Handbuch-Gerüst und vor `welle-record-replay` in `done/` starten dürfen; liefert einer doch einen neuen Fehlercode, Schlüssel oder ein Kommando, gibt der Architect es mit Grund an den Planner zurück (Ausgang *offen* in §6 des Plans, `AGENTS.md` §3.13), und der Slice wartet auf `slice-handbuch-meldungskatalog`.

**Teile und Zusatz für den Start von Teil 1** (Umplanung vom 2026-10-30, Drift-Log der Roadmap; die Pläne in `open/` bleiben dafür unverändert, Entscheidung des Users vom 2026-10-21 zu den Plänen in `open/`): Teil 1 sind die ersten sechs Slices der Tabelle, Teil 2 die übrigen fünf. Die Start-Zeilen der Pläne von `slice-odata-url-pfad`, `slice-odata-json-schreiben` und `slice-odata-json-eingabe` nennen `welle-record-replay` in `done/`; für Teil 1 gilt stattdessen der Trigger aus §2, und der Planner trägt ihn beim Start in den Plan ein und nennt im Start-Commit den Grund. Der Start-Trigger von `slice-odata-url-filter`, `slice-edm-ausdruecke` und `slice-query-interpreter` (der vorige Slice des Astes in `done/`) bleibt. Die Marke setzt der Planner vor dem ersten Code-Commit mit `make groessenmarke`; der Start eines Teil-1-Slice wartet nicht auf `slice-groessenmarke-werkzeug` in `done/`, wohl aber der erste Code-Commit auf die Fix-Runde des Werkzeugs in `main` (die Architect-Phase davor läuft ohne Werkzeug). Der Ast des Query-Interpreters ist eine Kette (`slice-odata-url-pfad` → `slice-odata-url-filter` → `slice-edm-ausdruecke` → `slice-query-interpreter`, jeder braucht den vorigen im Code); die beiden Codecs brauchen nur das Modell und sind gegeneinander und gegen die Kette unabhängig. Es läuft höchstens eine Linie von Teil 1 zugleich, die Slices nacheinander; empfohlene Reihenfolge: die Kette (zwei Leser fremder Bytes zuerst, sie tragen das größte Risiko), danach `slice-odata-json-schreiben`, danach `slice-odata-json-eingabe` (Leser fremder Bytes).
## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: `welle-hybrid` (`HybridUseCase` ruft `MockUseCase` über dessen Inbound Port), `welle-betrieb` (Spans `SPEC-029` und `SPEC-030`, Lasttest des Mockservers), `welle-etag`, `welle-expand-batch` und `welle-odata-v2` (Mock-Kern, Query-Interpreter und Codec); `welle-admin-api` mittelbar über `welle-betrieb` (die Endpunkte für Snapshot und Restore brauchen `SnapshotUseCase`).
- Wird blockiert von: Teil 1 von `slice-odata-codec-csdl`, `slice-odata-codec-json` und `slice-baseline-import` (Modell, Codec zum Lesen; erfüllt); Teil 2 von `welle-record-replay` (Baseline, Workspace, Codec zum Lesen, Services, HTTP-Adapter, Composition Root) und von `slice-handbuch-meldungskatalog` (Gerüst und Kataloge-Gate des Handbuchs, siehe §2). Die Parser- und Interpreter-Slices ändern das Binary nicht und laufen ohne ihn (Teil 1, §2); die Welle schließt nicht ohne ihn. Teil 1 ändert die Gate-Kette nicht (Linie B der Übergangsregeln); er berührt `src/hexagon/model` und `src/adapters/driven/odata/` und die Spezifikation, nicht die Dateien der Linie A außer `spec/spezifikation.md` und dem ADR-Index (Kennungsbereiche der Übergangsregeln).
- Unabhängig: die Werkzeug- und Doku-Wellen (`welle-dev-container`), die keine Produktfunktion liefern.
- Reihenfolge der Slices (Pfeil heißt: braucht): `slice-odata-url-pfad` → `slice-odata-url-filter` → `slice-edm-ausdruecke` → `slice-query-interpreter`; `slice-odata-json-schreiben` und `slice-odata-json-eingabe` brauchen keinen Slice des Query-Astes (nur das Modell) und liegen vor `slice-mock-lesen` bzw. `slice-mock-schreiben`; `slice-mock-lesen` braucht `slice-query-interpreter` und `slice-odata-json-schreiben`; `slice-mock-schreiben` braucht `slice-mock-lesen` (Konstruktor, Routing und Workspace-Zuordnung sind gemeinsam) und `slice-odata-json-eingabe`; `slice-snapshot-use-case` braucht `slice-mock-schreiben` (die Änderungen, die ein Snapshot festhält, entstehen erst dort); `slice-mock-abnahme` braucht `slice-snapshot-use-case`; `slice-bootstrap-m3` zuletzt.
- **Register-Sichtung (Eröffnung Schritt 2):** `observations/` gesichtet. Offen sind sieben Einträge, keiner erreicht mit dieser Welle 3×. Treffer, in den Plänen der Slices als Risiko geführt: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Slices mit vielen Fakes von Repositories und Codec (`slice-mock-lesen`, `slice-mock-schreiben`, `slice-snapshot-use-case`); `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für `slice-mock-schreiben` und `slice-mock-abnahme` (Workspace-Transaktionen, 20 parallele Workspaces); `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) für die ganze Welle, weil der Produktcode wächst, den Fälle der Gegenproben mitfahren; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für jeden Plan mit der Betriebsregel „nur `Edit` und `Write`“. Nicht berührt: `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (1×). Die verkörperten Regeln (`AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug`, §3.15, §3.16, §3.17) gelten für jeden vertragliefernden Slice.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Hybrid-Modus und Upstream-Fallback (`LH-FA-12`) — `welle-hybrid`; der Mock kennt keinen Upstream-Port.
- HTTP-Endpunkte für Workspace, Reset, Snapshot und Restore (`LH-FA-13`) — `welle-admin-api`; der M3-Test erreicht sie über den Composition Root.
- Spans `SPEC-029` und `SPEC-030`, Audit-Log, Metriken, TLS des Listeners — `welle-betrieb`; der Lasttest zu `LH-QA-01` für den Mockserver bleibt bis dahin ungeprüft.
- ETags, `If-Match` (`LH-FA-15`) und `Prefer: return=…` — `welle-etag`; bis dahin gilt das Muss-Verhalten der Vertragsmatrix (`LH-FA-07.b`: POST `201`, PATCH und PUT `204`).
- `$expand`, `$batch` (`LH-FA-16`), `$apply`, `$search`, Actions, Functions, Deep Insert — `welle-expand-batch` bzw. nicht im ersten Abnahmeumfang (Lastenheft §5); der Mock antwortet mit `501` (`ODR-E6001`, `ODR-E6002`).
- OData V2 (`LH-FA-17`) — `welle-odata-v2`.
- Das Service-Dokument (`GET` auf den Service-Root): nicht im Muss-Umfang (Entscheidung des Users, 2026-10-30), kein Vertrag in der Spezifikation, der Mock antwortet mit `501`. Der Architect von `slice-mock-lesen` prüft vor dem ersten Code-Commit am Referenz-Client des M3-Tests, ob eine Mindestfassung mit eigenem Vertrag nötig ist (Auftrag dort in §1); nur dann gehört sie zur Welle.
- Umschreiben von Links (`LINK_REWRITE` über `off`): kein Eintrag im Lastenheft, keine Welle.
- Coverage-Schwelle als Gate — gehört zu `welle-sensoren`.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-mock-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).
