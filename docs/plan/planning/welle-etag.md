# Welle welle-etag: ETags, bedingte Änderungen und Prefer (Soll)

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. Diese Welle ist vor ihrem Start angelegt
(Entscheidung des Users vom 2026-10-30); bis ihr Trigger eintritt steht sie in der
Roadmap unter *Nächste Wellen*. Ihre Slice-Dateien liegen seit
2026-10-30 in `open/`; die Pläne tragen ihre Start-Trigger mit Abhängigkeiten.

**Zielmeilenstein:** kein Meilenstein-Bezug (Soll-Umfang, nicht Teil der Erstabnahme nach dem Lastenheft).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Der Mock liefert je Entity ein stabiles Versionskennzeichen (ETag aus dem Versionszähler der Entity, `entity_version` im Schema), prüft `If-Match` und `If-None-Match` bei PATCH, PUT und DELETE und weist ein veraltetes ETag mit `412` ab, ohne den Zustand zu ändern (`LH-FA-15`). Dazu kommt `Prefer: return=minimal|representation` für mutierende Requests, der Soll-Teil der CRUD-Vertragsmatrix (`LH-FA-07.b`, Zeile Header in `SPEC-049`).
Ohne aktivierte ETag-Unterstützung behauptet der Mock keinen Concurrency-Schutz (`LH-FA-15.a`).
Soll-Umfang: nicht Teil der Erstabnahme (Lastenheft, Absatz *Erstabnahme*). Eine Welle, weil der Closure-Trigger den Rundlauf über alle Mutationsarten und die Header-Zeilen der Matrix beobachtet, nicht nur die Einzelfälle der Slices.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-mock` liegt in `done/`: ETag, `If-Match` und `Prefer` setzen `MockUseCase` mit Mutation voraus.
- Die Reihenfolge nach den Muss-Wellen (`welle-hybrid`, `welle-betrieb`, `welle-admin-api`) ist Priorisierung, keine Abhängigkeit; sie steht deshalb nicht im Trigger.

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün; `make gegenproben` ist grün.
- Ein Test im Modus `mock` liest eine Entity zweimal mit gleichem ETag, ändert sie mit aktuellem `If-Match` (neues ETag sichtbar), weist dieselbe Änderung mit dem alten ETag mit `412` ab (Zustand unverändert) und führt `Prefer: return=minimal` und `return=representation` für POST, PATCH und PUT aus; die als Soll geführten Zeilen der Matrix `SPEC-049` (Header und Fehler) sind grün.
- Liegt `welle-hybrid` in `done/`, deckt der Test auch den Fall, dass ein `412` im Modus `hybrid` keinen Fallback auslöst (`LH-FA-12.b`).
- Das Handbuch deckt alles ab, was das Binary dieser Welle kann (ETag, `If-Match`, `Prefer`, Fehlercode für `412`); das Kataloge-Gate aus `slice-handbuch-meldungskatalog` ist grün.
- Die Closure-Notiz liegt als `welle-etag-results.md` im Ruheort `done/`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-etag-versionen | ETag aus dem Versionszähler der Entity: stabil beim Lesen, neue Version bei jeder Änderung, Header in den Antworten | `LH-FA-15`, `LH-QA-02`; `LH-FA-15.a`, `LH-QA-02.a` |
| slice-if-match | `If-Match` und `If-None-Match` bei PATCH, PUT und DELETE, `412` ohne Zustandsänderung, atomare Versionserhöhung | `LH-FA-15`; `LH-FA-15.a` |
| slice-prefer-return | `Prefer: return=minimal` und `return=representation` für mutierende Requests | `LH-FA-07`; `LH-FA-07.b`, `SPEC-049` |

Die Slice-Dateien liegen seit 2026-10-30 in `open/`; die Pläne tragen ihre Start-Trigger mit Abhängigkeiten (alle drei starten nach `welle-mock`), und `plan-welle` gleicht sie beim Start gegen den Ist-Stand ab. Marken und Betriebsregeln setzt der Planner beim Start (`SPEC-128`).

**Pflicht Handbuch:** Jeder Slice, dessen Ergebnis ein Anwender am Binary sieht (neues Kommando, Konfigurationsschlüssel, Fehlercode, HTTP-Schnittstelle, Betriebsart), trägt in seiner DoD den Punkt „Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün“. In dieser Welle betrifft er alle drei Slices (ETag-Header, `412`, `Prefer`).

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: keine Welle. Eine Berührung ohne Kante: ein `PATCH` mit `If-Match` in einem Changeset von `welle-expand-batch` und ein `412` im Modus `hybrid`; wer von beiden später liegt, ergänzt den Fall in seinem Test.
- Wird blockiert von: `welle-mock` (`MockUseCase` mit Mutation, `entity_version` im Schema).
- Unabhängig: `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-expand-batch`, `welle-odata-v2`.
- Reihenfolge der Slices (Pfeil heißt: braucht): `slice-etag-versionen` → `slice-if-match`; `slice-prefer-return` unabhängig davon.
- **Register-Sichtung (Eröffnung Schritt 2):** `observations/` gesichtet, Stand dieses Plans. Treffer, beim Start in den Slice-Plänen (§8) mit dem dann aktuellen Zähler als Risiko zu führen: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für „Versionserhöhung erfolgt atomar“ bei gleichzeitigen Änderungen (Wirkungs-Fänger, `AGENTS.md` §3.10); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Kein offener Eintrag erreicht mit dieser Welle 3×.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Verteilte Versionskoordination über mehrere Instanzen — Out-of-Scope von `LH-FA-15`.
- Erzeugen von ETags im Record und Replay — aufgezeichnete ETags bleiben Header der Exchanges.
- Fehlercode für `412` und sein Eintrag in der Fehlertabelle — Aufgabe des Architect vor dem Start (die Tabelle führt heute keinen Code für `412`); nicht Gegenstand der Slices.
- `Prefer`-Werte außer `return=minimal` und `return=representation`.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-etag-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).
