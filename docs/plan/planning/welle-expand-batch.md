# Welle welle-expand-batch: Navigationen mit `$expand` und `$batch` mit Changesets (Soll)

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. Diese Welle ist vor ihrem Start angelegt
(Entscheidung des Users vom 2026-10-30); bis ihr Trigger eintritt steht sie in der
Roadmap unter *Nächste Wellen*. Der Zustand eines Slice ist das Verzeichnis, in
dem seine Datei liegt; die Links der Slice-Liste (§4) nennen es.

**Zielmeilenstein:** kein Meilenstein-Bezug (Soll-Umfang, nicht Teil der Erstabnahme nach dem Lastenheft).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Der Mock löst Navigationen mit `$expand` auf und verarbeitet `$batch` mit atomaren Changesets in einem definierten Umfang (`LH-FA-16`): verknüpfte Daten und Einzelergebnisse sind korrekt sichtbar, ein leerer Batch und eine leere Navigation ergeben eine gültige leere Antwort, ein fehlerhaftes Changeset übernimmt keine Teiländerung.
Jede Funktion hat einen eigenen Parser-, Validator-, Executor- und Serializer-Contract-Test (`LH-FA-16.a`); unbekannte Operationen im Mock liefern `501`.
Soll-Umfang: nicht Teil der Erstabnahme. Eine Welle, weil der Closure-Trigger beide Funktionen zusammen mit der Abnahmematrix beobachtet (`$expand` ist dort „nicht im MVP“, `SPEC-049`) und die Atomarität über die Workspace-Transaktion des Mock, nicht in einem Slice allein.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-mock` liegt in `done/`: `$expand` erweitert den Query-Interpreter und `$batch` führt Changesets in der Workspace-Transaktion des Mock aus.
- Die Reihenfolge nach den Muss-Wellen ist Priorisierung, keine Abhängigkeit; sie steht deshalb nicht im Trigger.

**Trigger je Slice (Code-Abhängigkeit, Zusatz zum Trigger):** Der Trigger der Welle bleibt formal und gilt für den Abschluss.
- [`slice-batch-parser`](done/slice-batch-parser.md) hängt an keinem Code aus `welle-mock` (der Leser braucht Fehlermodell, Konvention der Codec-Ports und Modell); der Welle-Trigger ist für ihn Priorisierung, ein Start davor ist eine Entscheidung des Users.
- [`slice-expand-parser`](done/slice-expand-parser.md) braucht [`slice-odata-url-pfad`](done/welle-mock/slice-odata-url-pfad.md) aus `welle-mock` (Teil 1, reine Funktion), nicht `welle-mock` ganz; nennt der Architect `$filter` in verschachtelten Optionen, auch [`slice-odata-url-filter`](done/welle-mock/slice-odata-url-filter.md) und [`slice-edm-ausdruecke`](done/welle-mock/slice-edm-ausdruecke.md).
- [`slice-expand-validator`](done/slice-expand-validator.md) braucht [`slice-expand-parser`](done/slice-expand-parser.md) (Knoten `$expand`) und das Modell der Metadaten, nicht `welle-mock` ganz.
- [`slice-expand-bausteine`](open/slice-expand-bausteine.md) braucht [`slice-expand-parser`](done/slice-expand-parser.md), [`slice-expand-validator`](done/slice-expand-validator.md) (Knoten und `PruefeExpand`) und den Vertrag zu Datenbasis und Ausgabe (`SPEC-264`, `SPEC-266`), nicht den Mock-Dienst.
- [`slice-expand-ausfuehrung`](open/slice-expand-ausfuehrung.md) braucht `welle-mock`, [`slice-expand-parser`](done/slice-expand-parser.md), [`slice-expand-validator`](done/slice-expand-validator.md) und [`slice-expand-bausteine`](open/slice-expand-bausteine.md) (Beziehungen und Schreiber der eingebetteten Entities).
- [`slice-batch-validator`](next/slice-batch-validator.md) braucht [`slice-batch-parser`](done/slice-batch-parser.md) (Domain-Werte, Fehlerwert und `model.ContentIDGueltig`), nicht `welle-mock`; der Welle-Trigger ist für ihn Priorisierung wie für den Parser.
- [`slice-batch-changesets`](open/slice-batch-changesets.md) braucht `welle-mock` (Workspace-Transaktion), [`slice-batch-parser`](done/slice-batch-parser.md) und [`slice-batch-validator`](next/slice-batch-validator.md).

**Stand `$batch`:** [`slice-batch-parser`](done/slice-batch-parser.md) liefert Port, Domain-Werte, Fehlerwert und Leser, ohne Aufrufer und ohne Route (Beleg: §7 dort). [`slice-batch-validator`](next/slice-batch-validator.md) (Prüfung der Struktur) ist gestartet (Zweig `slice-batch-validator`, kein Kettenslice, nicht Teil der Erstabnahme; Start-Angaben und Marke in seinem Plan); die Sendungen an ihn und an [`slice-batch-changesets`](open/slice-batch-changesets.md) stehen in deren §1. `LH-FA-16` ist damit nicht erfüllt.

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün; `make gegenproben` ist grün.
- Ein Test im Modus `mock` liest eine Entity mit `$expand` über eine Navigation (verknüpfte Daten sichtbar, leere Navigation leer), führt einen gültigen Batch mit Einzelergebnissen aus, beantwortet einen leeren Batch gültig und übernimmt aus einem fehlerhaften Changeset keine Teiländerung (Zustand vor und nach dem Changeset gleich).
- Je Funktion (`$expand`, `$batch`) liegt ein Parser-, Validator-, Executor- und Serializer-Contract-Test vor (`LH-FA-16.a`); `$apply`, Actions und Functions ergeben weiter `501`.
- Das Handbuch deckt alles ab, was das Binary dieser Welle kann (`$expand`, `$batch`, Fehlercodes); das Kataloge-Gate aus [`slice-handbuch-meldungskatalog`](done/slice-handbuch-meldungskatalog.md) ist grün.
- Die Closure-Notiz liegt als `welle-expand-batch-results.md` im Ruheort `done/`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| [slice-expand-parser](done/slice-expand-parser.md) | `$expand` in der OData-URL: Parser und AST | `LH-FA-16`; `LH-FA-16.a`, `SPEC-049` |
| [slice-expand-validator](done/slice-expand-validator.md) | Validator von `$expand` gegen die Navigationen des Modells (Schnitt von [`slice-expand-parser`](done/slice-expand-parser.md)) | `LH-FA-16`; `LH-FA-16.a` |
| [slice-expand-bausteine](open/slice-expand-bausteine.md) | Angaben des `$metadata` und Bestimmung der Beziehung je Navigation, Schreiber eingebetteter Entities (Schnitt von [`slice-expand-ausfuehrung`](open/slice-expand-ausfuehrung.md)) | `LH-FA-16`; `LH-FA-16.a`, `SPEC-264`, `SPEC-266` |
| [slice-expand-ausfuehrung](open/slice-expand-ausfuehrung.md) | `$expand` im Mock ausführen: Verdrahtung, Coverage, Auflösen gegen den Entity-State, Antwort | `LH-FA-16`, `LH-FA-11`; `LH-FA-16.a`, `SPEC-265`, `SPEC-044` |
| [slice-batch-parser](done/slice-batch-parser.md) | Parser für `$batch`-Requests (Leser fremder Bytes), leerer Batch | `LH-FA-16`; `LH-FA-16.a` |
| [slice-batch-validator](next/slice-batch-validator.md) | Prüfung der Struktur eines Batch: Methoden, `Content-ID`, Verweise (Schnitt von [`slice-batch-parser`](done/slice-batch-parser.md)) | `LH-FA-16`; `LH-FA-16.a` |
| [slice-batch-changesets](open/slice-batch-changesets.md) | Ausführung der Einzelrequests und atomare Changesets in einer Workspace-Transaktion, Antwort des Batch | `LH-FA-16`, `LH-QA-03`; `LH-FA-16.a`, `LH-FA-07.a` |

Der Zustand eines Slice ist das Verzeichnis, in dem seine Datei liegt; die Links der Tabelle nennen es. Die Pläne tragen ihre Start-Trigger mit Abhängigkeiten, und `plan-welle` gleicht sie beim Start gegen den Ist-Stand ab. Marken und Betriebsregeln setzt der Planner beim Start (`SPEC-128`).

**Pflicht Handbuch:** Jeder Slice, dessen Ergebnis ein Anwender am Binary sieht (neues Kommando, Konfigurationsschlüssel, Fehlercode, HTTP-Schnittstelle, Betriebsart), trägt in seiner DoD den Punkt „Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün“. In dieser Welle betrifft er [`slice-expand-ausfuehrung`](open/slice-expand-ausfuehrung.md) und [`slice-batch-changesets`](open/slice-batch-changesets.md) (sichtbar am Binary), die Parser- und Validator-Slices und [`slice-expand-bausteine`](open/slice-expand-bausteine.md) nur, soweit sie einen Fehlercode einführen.

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: keine Welle. Eine Berührung ohne Kante: ein `PATCH` mit `If-Match` in einem Changeset (`welle-etag`); wer von beiden später liegt, ergänzt den Fall in seinem Test.
- Wird blockiert von: `welle-mock` (Query-Interpreter, Workspace-Transaktion, Codec zum Schreiben); die Parser- und Validator-Slices hängen nur an Teil 1 beziehungsweise an keinem Code daraus (Trigger je Slice in §2).
- Unabhängig: `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-odata-v2`.
- Reihenfolge der Slices (Pfeil heißt: braucht): [`slice-expand-parser`](done/slice-expand-parser.md) → [`slice-expand-validator`](done/slice-expand-validator.md) → [`slice-expand-bausteine`](open/slice-expand-bausteine.md) → [`slice-expand-ausfuehrung`](open/slice-expand-ausfuehrung.md); [`slice-batch-parser`](done/slice-batch-parser.md) → [`slice-batch-validator`](next/slice-batch-validator.md) → [`slice-batch-changesets`](open/slice-batch-changesets.md); die beiden Äste sind unabhängig.
- **Register-Sichtung (Eröffnung Schritt 2):** `observations/` gesichtet, Stand dieses Plans. Treffer, beim Start in den Slice-Plänen (§8) mit dem dann aktuellen Zähler als Risiko zu führen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) für die Atomarität der Changesets; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Kein offener Eintrag erreicht mit dieser Welle 3×.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Actions und Functions, `$apply`, `$search`, Delta-Abfragen, Deep Insert — Lastenheft §5 und `SPEC-049` (nicht im MVP); sie bleiben `501`.
- Herstellerspezifische Batch- und Navigations-Erweiterungen — Out-of-Scope von `LH-FA-16`.
- Verschachtelte Optionen im `$expand` (`$filter`, `$select` innerhalb der Navigation) — nicht im unterstützten Umfang (`SPEC-262`); sie bleiben `501`.
- Aufzeichnen und Abspielen von Batch-Requests — der Record-Pfad leitet sie als HTTP unverändert weiter, der Replay-Pfad bedient sie unabhängig vom Parser (`LH-FA-16.a`).
- Beziehungen aus aufgezeichneten `$expand`-Antworten, aus der Konfiguration oder aus Namen — Datenbasis sind allein die Referenzbedingungen des `$metadata` (`SPEC-264` Nr. 1); eine so nicht beschriebene Navigation bleibt `501`. Ob dieser Umfang `LH-FA-16` trägt, ist eine offene Frage an den User ([`slice-expand-ausfuehrung`](open/slice-expand-ausfuehrung.md) §6).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-expand-batch-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).
