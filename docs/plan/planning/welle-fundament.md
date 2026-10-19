# Welle welle-fundament: Code-Gates auf einem Go-Skelett

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. **Geplante Wellen bekommen noch keine
Datei:** Sie stehen in der Roadmap unter *Nächste Wellen* und nirgends sonst —
zwei Positionen, nicht drei.

**Zielmeilenstein:** M1 (siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-18.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

`make gates` fährt Build, Test, Lint und `a-check` gegen echten Go-Code unter `src/`, und jedes dieser Gates ist durch eine Gegenprobe als rot beim Verstoß belegt.

Eine Welle, weil der Closure-Trigger mehr beobachtet als die DoDs der Slices: das Zusammenspiel aller Gates auf demselben Code und den Nachweis, dass jedes Gate einen Verstoß findet.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- [ADR-0001](../adr/0001-implementierungssprache-go.md) accepted (Sprache).
- [ADR-0002](../adr/0002-architektur-gate-a-check.md) accepted (`a-check`).
- [ADR-0003](../adr/0003-lint-profil-golangci.md) accepted (Lint-Profil).

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün auf dem Go-Skelett und führt dabei `make build`, `make test`, `make lint` und `make a-check` aus.
- `make gegenproben` ist grün: jede Gegenprobe weist ihr Gate in einer Kopie des Arbeitsbaums als rot nach.
- Die Closure-Notiz liegt als `welle-fundament-results.md` im Ruheort `done/`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-go-skelett-build-test | Go-Skelett mit Build und Test in Containern | `LH-QA-06` |
| slice-lint-stufe | Lint-Stufe für den Code samt eigener Prüfungen | `LH-QA-06` |
| slice-gegenproben | Gegenproben für `a-check` und Lint, außerhalb der Gate-Kette | `LH-QA-06` |
| slice-lint-gegenproben-modul | `tools/gegenproben` unter dem Lint-Gate | `LH-QA-06` |

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: Welle welle-persistenz (braucht die Container-Stufen für Build, Test und Lint, bevor Repositories entstehen).
- Blockiert: Welle welle-dev-container (braucht `make gates` und `make gegenproben` als Gegenstand).
- Wird blockiert von: —.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Fachlogik (Record, Replay, Mock, Import) — gehört in die Wellen nach der Persistenz; das Skelett trägt nur Composition Root und Version.
- SQLite-Treiber, Migration und Repositories — eigene ADR und Welle welle-persistenz; vorher gibt es keine zugelassene Fremdbibliothek.
- Coverage-Schwelle und Integrationstests — eine Schwelle über einem Skelett misst nichts und wäre Zeremonie; sie kommen mit der ersten Fachlogik.
- Arbeitscontainer, in dem Gates und Gegenproben laufen, ohne den Hauptzweig zu blockieren — eigene Welle welle-dev-container, weil sie `make gates` samt Gegenproben schon voraussetzt.
- Produkt-Image und Release — erst sinnvoll, wenn der Prozess etwas tut.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-fundament-results.md` (entsteht bei der Closure, Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).
