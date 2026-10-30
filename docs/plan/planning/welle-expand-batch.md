# Welle welle-expand-batch: Navigationen mit `$expand` und `$batch` mit Changesets (Soll)

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. Diese Welle ist vor ihrem Start angelegt
(Entscheidung des Users vom 2026-10-30); bis ihr Trigger eintritt steht sie in der
Roadmap unter *Nächste Wellen*. Ihre Slice-Dateien legt `plan-welle` (Schritt 6)
beim Start an, weil ein Slice-Plan vor seinem Start veraltet.

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

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün; `make gegenproben` ist grün.
- Ein Test im Modus `mock` liest eine Entity mit `$expand` über eine Navigation (verknüpfte Daten sichtbar, leere Navigation leer), führt einen gültigen Batch mit Einzelergebnissen aus, beantwortet einen leeren Batch gültig und übernimmt aus einem fehlerhaften Changeset keine Teiländerung (Zustand vor und nach dem Changeset gleich).
- Je Funktion (`$expand`, `$batch`) liegt ein Parser-, Validator-, Executor- und Serializer-Contract-Test vor (`LH-FA-16.a`); `$apply`, Actions und Functions ergeben weiter `501`.
- Das Handbuch deckt alles ab, was das Binary dieser Welle kann (`$expand`, `$batch`, Fehlercodes); das Kataloge-Gate aus `slice-handbuch-meldungskatalog` ist grün.
- Die Closure-Notiz liegt als `welle-expand-batch-results.md` im Ruheort `done/`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-expand-parser | `$expand` in der OData-URL: Parser und Validator gegen die Navigationen des Modells | `LH-FA-16`; `LH-FA-16.a`, `SPEC-049` |
| slice-expand-ausfuehrung | Auflösen der Navigation im Query-Interpreter und Ausgabe im OData-JSON; Datenbasis der Navigationen nach Entscheidung des Architect | `LH-FA-16`, `LH-FA-11`; `LH-FA-16.a`, `SPEC-044` |
| slice-batch-parser | Parser für `$batch`-Requests (Leser fremder Bytes), leerer Batch | `LH-FA-16`; `LH-FA-16.a` |
| slice-batch-changesets | Ausführung der Einzelrequests und atomare Changesets in einer Workspace-Transaktion, Antwort des Batch | `LH-FA-16`, `LH-QA-03`; `LH-FA-16.a`, `LH-FA-07.a` |

Die Slice-Dateien legt `plan-welle` (Schritt 6) beim Start der Welle an: Ein Slice-Plan altert vor seinem Start, und diese Welle beginnt erst nach `welle-mock`.

**Pflicht Handbuch:** Jeder Slice, dessen Ergebnis ein Anwender am Binary sieht (neues Kommando, Konfigurationsschlüssel, Fehlercode, HTTP-Schnittstelle, Betriebsart), trägt in seiner DoD den Punkt „Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün“. `plan-welle` (Schritt 6) übernimmt ihn in jede Slice-Datei; in dieser Welle betrifft er `slice-expand-ausfuehrung` und `slice-batch-changesets` (sichtbar am Binary), die Parser-Slices nur, soweit sie einen Fehlercode einführen.

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: keine Welle. Eine Berührung ohne Kante: ein `PATCH` mit `If-Match` in einem Changeset (`welle-etag`); wer von beiden später liegt, ergänzt den Fall in seinem Test.
- Wird blockiert von: `welle-mock` (Query-Interpreter, Workspace-Transaktion, Codec zum Schreiben).
- Unabhängig: `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-odata-v2`.
- Reihenfolge der Slices (Pfeil heißt: braucht): `slice-expand-parser` → `slice-expand-ausfuehrung`; `slice-batch-parser` → `slice-batch-changesets`; die beiden Äste sind unabhängig.
- **Register-Sichtung (Eröffnung Schritt 2):** `observations/` gesichtet, Stand dieses Plans. Treffer, beim Start in den Slice-Plänen (§8) mit dem dann aktuellen Zähler als Risiko zu führen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) für die Atomarität der Changesets; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Kein offener Eintrag erreicht mit dieser Welle 3×.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Actions und Functions, `$apply`, `$search`, Delta-Abfragen, Deep Insert — Lastenheft §5 und `SPEC-049` (nicht im MVP); sie bleiben `501`.
- Herstellerspezifische Batch- und Navigations-Erweiterungen — Out-of-Scope von `LH-FA-16`.
- Verschachtelte Optionen im `$expand` (`$filter`, `$select` innerhalb der Navigation) — der unterstützte Umfang ist Entscheidung des Architect vor dem Start; was er nicht nennt, bleibt `501`.
- Aufzeichnen und Abspielen von Batch-Requests — der Record-Pfad leitet sie als HTTP unverändert weiter, der Replay-Pfad bedient sie unabhängig vom Parser (`LH-FA-16.a`).
- Die Frage, woher die Beziehungen der Navigationen stammen (der Import liest eingebettete Navigationen nicht, `SPEC-126` Nr. 9): Aufgabe des Architect vor dem Start; die Welle beginnt erst mit der Antwort.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-expand-batch-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).
