# Welle welle-odata-v2: OData V2 als Kompatibilitätsumfang (Kann)

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. Diese Welle ist vor ihrem Start angelegt
(Entscheidung des Users vom 2026-10-30); seit dem Start ihrer ersten Slices (Ausnahme je Slice, §2) steht sie in der
Roadmap unter *Offene Wellen*. Ihre Slice-Dateien liegen seit
2026-10-30 in `open/`; die Pläne tragen ihre Start-Trigger mit Abhängigkeiten.

**Zielmeilenstein:** kein Meilenstein-Bezug (Kann-Umfang, nicht Teil der Erstabnahme nach dem Lastenheft).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Ein als V2 deklarierter Dienst wird über einen versionierten Adapter verarbeitet: Metadaten und Standardabfrage nach dem dokumentierten V2-Umfang, ein leeres Ergebnis bleibt im V2-Antwortformat konsistent, eine nicht unterstützte V2-Eigenschaft ergibt einen expliziten Kompatibilitätsfehler (`LH-FA-17`). Das V4-Domainmodell bleibt unberührt, und es gibt keine Vermischung von V2 und V4 in einem Service-Modell (`LH-FA-17.a`, Erweiterungspunkt „OData V2“ der Architektur).
Kann-Umfang: nicht Teil der Erstabnahme, und die Aufzeichnung (`record`) ist von der Version unabhängig. Eine Welle, weil der Closure-Trigger den Rundlauf Import, Workspace und Abfrage für einen V2-Dienst beobachtet, nicht die Parser einzeln.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-mock` liegt in `done/`: der V2-Adapter ersetzt Payload-Schreiber und Query-Parser hinter den Ports, die der Mock-Kern nutzt; Query und Rundlauf setzen den Mock-Kern voraus.
- Eine ADR des Architect zum versionierten Adapter (Aufteilung der Ports je Version, Ort von `odata_version`) ist `Accepted`: erfüllt durch [ADR-0011](../adr/0011-versionierter-adapter-fuer-odata-v2.md) (2026-11-01).
- Die Reihenfolge nach den Muss-Wellen ist Priorisierung, keine Abhängigkeit; sie steht deshalb nicht im Trigger.

**Trigger je Slice (Code-Abhängigkeit, Zusatz zum Trigger):** `slice-v2-import-codecs`, `slice-v2-metadaten`, `slice-v2-payload-lesen` und `slice-v2-query` (der Parser) brauchen nur die ADR (`Accepted`, erfüllt) und die Spezifikation (`SPEC-170` bis `SPEC-177`), nicht `welle-mock`: der Parser kennt das Modell nicht, der Test über beide Seiten wandert in den Rundlauf von `slice-v2-bootstrap`. `slice-v2-payload-schreiben` braucht `slice-odata-json-schreiben` (Teil 1 von `welle-mock`) und den Vertrag der Schreib-Schnittstelle in der Spezifikation. `slice-v2-bootstrap` braucht alle davor, `welle-mock` ganz, `slice-bootstrap-m3` und `slice-handbuch-meldungskatalog`. Der Trigger der Welle bleibt formal.

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün; `make gegenproben` ist grün.
- Ein Test zeichnet einen Referenzdienst mit V2-Antworten auf, importiert die Session, legt einen Workspace an und beantwortet im Modus `mock` die Metadaten und eine Standardabfrage nach dem dokumentierten V2-Umfang; ein leeres Ergebnis bleibt im V2-Format, eine nicht unterstützte V2-Eigenschaft ergibt einen expliziten Kompatibilitätsfehler.
- Ein Dienst der Version V4 verhält sich nach dem Slice-Satz wie vorher: die Tests von `welle-mock` laufen unverändert grün.
- Das Handbuch deckt alles ab, was das Binary dieser Welle kann (V2-Umfang, Kompatibilitätsfehler); das Kataloge-Gate aus `slice-handbuch-meldungskatalog` ist grün.
- Die Closure-Notiz liegt als `welle-odata-v2-results.md` im Ruheort `done/`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-v2-metadaten | Metadaten-Leser für V2 (EDMX/CSDL V2) mit Validierungsbericht und Version je Service | `LH-FA-17`, `LH-RB-01`; `LH-FA-17.a`, `SPEC-002` |
| slice-v2-import-codecs | Der Import nimmt den Satz der Codecs je Dienst entgegen (`baselineimport.Dienst.Codecs`) | `LH-FA-17`, `LH-FA-18`; `SPEC-127` Nr. 1, `SPEC-175` |
| slice-v2-payload-lesen | Leser für V2-Antworten (JSON mit der Hülle `d`) für den Import | `LH-FA-17`; `LH-FA-17.a`, `SPEC-172` Nr. 1 bis 5 |
| slice-v2-payload-schreiben | Schreiber für V2-Antworten für den Mock, leeres Ergebnis, Fehlerobjekt | `LH-FA-17`, `LH-QA-05`; `SPEC-172` Nr. 6 und 7 |
| slice-v2-query | Query-Semantik des V2-Umfangs und expliziter Kompatibilitätsfehler `ODR-E6006` | `LH-FA-17`, `LH-QA-05`; `LH-FA-17.a`, `SPEC-173`, `SPEC-174` |
| slice-v2-bootstrap | Wahl des Satzes je Service-Version im Composition Root, Mock-Ergänzungen und Rundlauftest | `LH-FA-17`, `LH-FA-12`; `LH-FA-17.a`, `SPEC-054`, `SPEC-175` |

Die Slice-Dateien liegen seit 2026-10-30 in `open/`; die Pläne tragen ihre Start-Trigger mit Abhängigkeiten, und `plan-welle` gleicht sie beim Start gegen den Ist-Stand ab. Marken und Betriebsregeln setzt der Planner beim Start (`SPEC-128`).

**Pflicht Handbuch:** Jeder Slice, dessen Ergebnis ein Anwender am Binary sieht (neues Kommando, Konfigurationsschlüssel, Fehlercode, HTTP-Schnittstelle, Betriebsart), trägt in seiner DoD den Punkt „Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün“. In dieser Welle betrifft er `slice-v2-bootstrap` (Wahl der Version je Service, Kompatibilitätsfehler) und die übrigen Slices nur, soweit sie einen Fehlercode einführen.

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: keine Welle.
- Wird blockiert von: `welle-mock` (die Ports und der Kern, hinter denen der V2-Adapter steht; `slice-v2-metadaten` und der Lese-Teil von `slice-v2-payload` ausgenommen, §2) und einer ADR des Architect zum versionierten Adapter.
- **Entschieden, Import je Version** (`SPEC-175` Nr. 4: derselbe Importer, Satz je Dienst; der Slice davor ist `slice-v2-import-codecs`). Die ursprüngliche Frage: `LH-FA-17` verlangt den Rundlauf Import, Workspace, Abfrage für einen V2-Dienst, `LH-FA-18` beschreibt den Import; `baselineimport.New` nimmt genau ein Paar von Codecs, und die Auswahl der Antworten (`SPEC-127` Nr. 8) und die Ketten über Folgelinks (Nr. 12) sind V4-Regeln. Wer den Import je Version anpasst und ob dafür ein Slice vor `slice-v2-bootstrap` nötig ist, steht in keinem Plan der Welle entschieden; `slice-v2-bootstrap` trägt nur die Verdrahtung.
- Unabhängig: `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-expand-batch`.
- Abhängigkeiten der Slices (Pfeil heißt: braucht; die frühere Kette metadaten → payload → query → bootstrap gilt nicht mehr): `slice-v2-import-codecs`, `slice-v2-metadaten`, `slice-v2-payload-lesen` und `slice-v2-query` brauchen nur [ADR-0011](../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted` und die Spezifikation; `slice-v2-payload-schreiben` → `slice-odata-json-schreiben` (`welle-mock`) und den Vertrag der Schreib-Schnittstelle; `slice-v2-bootstrap` → alle davor, `welle-mock`, `slice-bootstrap-m3`, `slice-handbuch-meldungskatalog`.
- **Zustand:** Geliefert ist `slice-v2-import-codecs` (Beleg [done/](done/slice-v2-import-codecs.md)). Gestartet ist `slice-v2-metadaten` (Plan in [in-progress/](in-progress/slice-v2-metadaten.md), Zweig slice-v2-metadaten, im Verifier-Lauf). Gestartet ist `slice-v2-payload-lesen` (Plan in [in-progress/](in-progress/slice-v2-payload-lesen.md), Zweig slice-v2-payload-lesen). Gestartet ist `slice-v2-query` (Plan in [in-progress/](in-progress/slice-v2-query.md), Zweig slice-v2-query).
- **Offene Rückgabe an den Architect** (Adresse: Architect; Ausgang *offen*): `SPEC-110` Nr. 9, Satz „Ohne `SERVICES_FILE` gilt genau ein Service aus `SERVICE_ID`, `SERVICE_ROOT_PATH` … und `UPSTREAM_URL`“, nennt `SERVICE_ODATA_VERSION` nicht; der Nachzug des Vorschlags (Teil C) hat diese Stelle nicht erfasst. Zu klären vor dem Start von `slice-v2-bootstrap`, der die Konfiguration liefert.
- **Register-Sichtung (Eröffnung Schritt 2):** `observations/` gesichtet, Stand dieses Plans. Treffer, beim Start in den Slice-Plänen (§8) mit dem dann aktuellen Zähler als Risiko zu führen: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für Fakes der Codec-Ports; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); die Slices `slice-v2-metadaten`, `slice-v2-payload` (Lesen) und `slice-v2-query` sind Leser fremder Bytes; die Formel nach `SPEC-128` kennt dafür keinen Zuschlag mehr, die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung. Kein offener Eintrag erreicht mit dieser Welle 3×.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Vollständige Kompatibilität zu herstellerspezifischen (SAP-)Erweiterungen — Out-of-Scope von `LH-FA-17`.
- Vermischung von V2 und V4 in einem Service-Modell — `LH-FA-17.a`.
- Änderungen am V4-Pfad (Mock, Replay, Record) über die Ports hinaus, hinter denen der Adapter steht.
- V2-Mutationen (POST, MERGE, PUT, DELETE): außerhalb des Umfangs, ergeben `ODR-E6006` mit der Ursache `methode` (`SPEC-170` Nr. 2, `SPEC-174` Nr. 3).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-odata-v2-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).
