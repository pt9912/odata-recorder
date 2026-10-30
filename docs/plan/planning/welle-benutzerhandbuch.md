# Welle welle-benutzerhandbuch: Benutzerhandbuch für Anwender des Recorders

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. **Geplante Wellen bekommen noch keine
Datei:** Sie stehen in der Roadmap unter *Nächste Wellen* und nirgends sonst —
zwei Positionen, nicht drei.

**Zielmeilenstein:** kein Meilenstein-Bezug (Dokumentations-Welle). Sie hängt an M2 (siehe [Roadmap](in-progress/roadmap.md)): Die Kapitel über Aufzeichnen und Wiedergeben beschreiben, was das gebaute Binary nach M2 zeigt, und stehen deshalb erst nach M2.

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Wer den Recorder in Tests einsetzt, erledigt seine Aufgaben mit docs/user/benutzerhandbuch.md allein: bauen und starten, konfigurieren, gegen einen Upstream aufzeichnen, ohne Upstream wiedergeben, eine Fehlermeldung (`ODR-E…`) deuten (`LH-QA-05`).
Das Handbuch ist die Stelle, an der der unterstützte Teilumfang und die Schutzmaßnahmen für Zugangsdaten und Nutzdaten für Anwender stehen (`LH-RB-01`, `LH-RB-04`); das Lastenheft verlangt kein Handbuch als Dokument, diese beiden Randbedingungen verlangen „verbindlich zu dokumentieren“ und „dokumentierte Schutzmaßnahmen“.
Das Handbuch folgt dem Standard `docs/user/benutzerhandbuch-standard.md` und sagt nur zu, was das gebaute Binary zeigt und ein Test oder Gate belegt (Ist-Zustand, `AGENTS.md` §3.11 sinngemäß für Dokumentation).
Eine Welle, weil der Closure-Trigger mehr beobachtet als die DoDs der Slices: Eine Person ohne Vorwissen führt die Aufgaben des Handbuchs allein nach dem Handbuch aus (`slice-handbuch-abnahme`).
Der Schnitt folgt den Gegenständen: Was schon im Code belegt ist (Bauen, Version, Startkonfiguration, Fehlercodes), steht vor M2; was erst M2 liefert (Aufzeichnen, Wiedergeben), steht nach M2.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `slice-groessenmarke-werkzeug` liegt in `done/`: Die Marken der Slices dieser Welle setzt der Planner mit `make groessenmarke` (`SPEC-128`), nicht mit der Prosa-Formel.
- `docs/user/benutzerhandbuch-standard.md` liegt auf `main` (vom User bereitgestellt; im Worktree als eigener Commit vor diesem Plan).

Der Start der Slices ist gestuft: `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` brauchen nur den Start der Welle; `slice-handbuch-record-replay` wartet auf `slice-bootstrap-m2` in `done/` (Trigger dort in §4).

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün und fährt den Abgleich der Fehlercodes aus `slice-handbuch-meldungskatalog`; `make gegenproben` ist grün.
- Der Probelauf aus `slice-handbuch-abnahme` ist gefahren: Eine Rolle ohne Vorwissen hat jede Aufgabe des Handbuchs allein nach dem Handbuch am gebauten Binary ausgeführt; jeder Befund ist behoben oder trägt einen Ausgang in §6 des Slice-Plans.
- Jedes Beispiel und jede Konfiguration im Handbuch, die als Datei vorliegt, läuft gegen das gebaute Binary in der Stufe `test` (Beleg-Form: Entscheidung des Architect in `slice-handbuch-geruest`).
- Die Closure-Notiz liegt als `welle-benutzerhandbuch-results.md` im Ruheort `done/`; ihre *Steering-Loop-Einträge* tragen, falls etwas verkörpert wird, den Anker `seit welle-benutzerhandbuch` mit dem Feld `liegt in`; die Pflicht der Folgewellen aus §6 (Kapitel im Handbuch liefern) steht dort als Vorschlag an den User (Zielort `AGENTS.md` oder `.claude/commands/implement-slice.md`).

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-handbuch-geruest | Gerüst von docs/user/benutzerhandbuch.md, Anpassung des Standards, Kapitel über Belegtes | `LH-RB-01`, `LH-RB-03`, `LH-QA-06` |
| slice-handbuch-meldungskatalog | Abgleich der Fehlercodes `ODR-E…` zwischen Code, Spezifikation und Handbuch | `LH-QA-05` |
| slice-handbuch-record-replay | Kapitel Aufzeichnen und Wiedergeben mit belegten Beispielen | `LH-FA-02`, `LH-FA-04`, `LH-FA-14`, `LH-RB-04` |
| slice-handbuch-abnahme | Probelauf durch eine Rolle ohne Vorwissen, Befunde, Pflege von Version und Stand | `LH-QA-05` |

Kein Slice „Standard“: Der Standard liegt als `docs/user/benutzerhandbuch-standard.md` auf `main`. Was er für dieses Repo anpasst, ist ein Auftrag an den Architect in `slice-handbuch-geruest` (Begründung dort in §1).

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: nichts. Die Folgewellen für Mock, Hybrid und Admin-API liefern ihre Kapitel selbst (§6).
- Wird blockiert von: `slice-groessenmarke-werkzeug` (Start der Welle, §2); `slice-bootstrap-m2` blockiert nur `slice-handbuch-record-replay`, nicht die Welle.
- Unabhängig von: welle-dev-container (deren „Dokument für Entwickler“ ist ein anderes Dokument mit anderer Zielgruppe) und der Welle der Sensoren.
- Reihenfolge der Slices (Pfeil heißt: braucht): `slice-handbuch-geruest` → `slice-handbuch-meldungskatalog` (das Gate liest die Abschnitte, die das Gerüst anlegt) → `slice-handbuch-record-replay` (braucht zusätzlich `slice-bootstrap-m2` in `done/`) → `slice-handbuch-abnahme` zuletzt.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Kapitel über Mock-Modus, Hybrid-Modus und Fallback, Admin-API, Snapshots, Telemetrie, TLS des Listeners — Folgewellen; das Kapitel entsteht in der Welle, die das Verhalten liefert. Ein Kapitel über Verhalten, das das Binary nicht zeigt, widerspricht dem Ist-Zustand. Die Pflicht für jene Wellen trägt die Closure dieser Welle als Vorschlag an den User.
- Ein Dokument für Entwickler des Recorders (Architektur, Bauen für die Weiterentwicklung, Arbeitscontainer) — andere Zielgruppe; `AGENTS.md`, `spec/` und die Welle welle-dev-container tragen es.
- Betriebs- und Sicherheitsdetails, die einem Angreifer helfen (genaue Sicherheitskonfiguration, Token-Verwaltung der Admin-API) — erst mit der Admin-API; der Standard (Abschnitt 10) verlangt Zurückhaltung.
- Screenshots, PDF-Export, Suchfunktion, Übersetzung — der Recorder hat keine Oberfläche; die Abschnitte 6, 7, 11 und 12 des Standards gelten nur, soweit `slice-handbuch-geruest` sie übernimmt.
- Release-Prozess und Versionsnummern-Vergabe — kein Release-Prozess in diesem Repo; das Handbuch nennt die Version, die das gebaute Binary meldet.
- Das Lasttest-Ziel von `LH-QA-01` und jede Zahl zur Performance — ungeprüft (welle-record-replay §6); das Handbuch sagt dazu nichts zu.
- Die Entscheidung über Ort und Bindung des Abgleichs in der Gate-Kette und über die Laufzeit nach [ADR-0005](../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) — Sache des Users, siehe `slice-handbuch-meldungskatalog` §6.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-benutzerhandbuch-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).
