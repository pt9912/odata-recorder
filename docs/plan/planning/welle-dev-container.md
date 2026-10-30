# Welle welle-dev-container: Arbeitscontainer für Entwickler

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. **Geplante Wellen bekommen noch keine
Datei:** Sie stehen in der Roadmap unter *Nächste Wellen* und nirgends sonst —
zwei Positionen, nicht drei.

**Zielmeilenstein:** kein Meilenstein-Bezug (Werkzeug-Welle; sie liegt außerhalb des Pfads von M2, siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Ein Entwickler startet auf dem Host `u-boot` einen Arbeitscontainer mit eigenem Docker-Daemon und festem Speicher- und CPU-Limit, liefert einen Zweig per SSH in ein Bare-Repo des Containers ein und fährt dort `make gates` und `make gegenproben` — ohne den Daemon des Hosts zu belegen und ohne dass zwei Läufe sich gegenseitig die Zeitgrenzen reißen.
Der Anlass sind vier Beobachtungen: der Host lief mehrfach in den Speichermangel (Mutationsläufe belegten ein Vielfaches des Bedarfs des unmutierten Tests), Gate- und Gegenprobenläufe mehrerer Rollen teilen sich einen Daemon samt festen Image-Tags (`odata-recorder:*`) und stören sich, die Gegenproben reißen ihre Frist je Fall unter Last (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`), und Slices sollen parallel in getrennten Arbeitsbäumen laufen.
Die Welle erfüllt zugleich die Folgepflicht aus [ADR-0005](../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): „Der Arbeitscontainer für Entwickler fährt `make gegenproben`“.
Sie ist ein Werkzeug der Harness (Verträge unter `harness/vertraege/`, [ADR-0004](../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)), kein Produktverhalten: Das Lastenheft führt keine Anforderung an einen Arbeitscontainer; der Bezug ist `LH-QA-06` und `LH-RB-03` (Reproduzierbarkeit, Entwicklung in Docker-Umgebungen) und die Folgepflicht der ADR.
Die Welle trifft drei Entscheidungen nicht selbst, sondern gibt sie dem Architect auf (jede als ADR vor dem Code): ob und wie der Container privilegiert läuft (Verhältnis zu `AGENTS.md` §3.1, „Host braucht nur Docker und GNU make“), das Netz des Containers (Prüfläufe sind netzlos, `deps` holt aus dem Netz) und die Verwaltung der Schlüssel.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-fundament` liegt in `done/` ([Closure-Notiz](done/welle-fundament-results.md)): `make gates` und `make gegenproben` bestehen und sind das, was der Container fährt.
- Die Welle ist unabhängig von `welle-record-replay` und läuft neben ihr; sie ändert weder `src/` noch die Verträge des Produkts.

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- Das *Mehr* gegenüber den DoDs: Auf `u-boot` startet ein frischer Container (ohne Zustand aus einem früheren Lauf) mit gesetztem Speicher- und CPU-Limit und meldet Bereitschaft; ein Zweig wird von einem Klon außerhalb des Containers per SSH eingeliefert; im Container laufen `make gates` und `make gegenproben` grün auf diesem Stand; der Daemon des Hosts hat dabei kein Image und keinen Container des Laufs angelegt.
- Zwei solche Container laufen gleichzeitig, je mit einem anderen Zweig, und beide Läufe sind grün; keine Frist je Fall der Gegenproben reißt (Zeit je Fall im Bericht neben dem Wert aus dem Einzellauf). Ein Lauf, der das Speicherlimit überschreitet, beendet den Container, nicht den Host.
- `make gates` im Hauptbaum ist grün (die Welle ändert Dockerfile, Makefile-Fragmente oder Gegenproben; die Gegenproben sind gefahren).
- Die Closure-Notiz liegt als `welle-dev-container-results.md` im Ruheort `done/`; ihre *Steering-Loop-Einträge* behandeln mindestens `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` und `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (Ausgang aus dem Lauf im Container) und die Folgepflicht aus [ADR-0005](../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) als erfüllt.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| [slice-dev-container-geruest](open/slice-dev-container-geruest.md) | Gerüst des Arbeitscontainers: Image, eigener Docker-Daemon, Limits, Start mit Bereitschaftsprobe | `LH-QA-06`, `LH-RB-03` |
| [slice-dev-container-zugang](open/slice-dev-container-zugang.md) | Zugang zum Container: SSH und Schlüsselpaar | `LH-RB-03`, `LH-RB-04` |
| [slice-dev-container-einlieferung](open/slice-dev-container-einlieferung.md) | Bare-Repo, Klon und Einlieferung per Zweig | `LH-RB-03`, `LH-QA-06` |
| [slice-dev-container-lauf](open/slice-dev-container-lauf.md) | `make gates` und `make gegenproben` im Container | `LH-QA-06`, `LH-QA-03` |
| [slice-dev-container-dokument](open/slice-dev-container-dokument.md) | Dokument für Entwickler | `LH-RB-03` |

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: nichts in der Roadmap. Sie macht die parallele Arbeit mehrerer Rollen in getrennten Arbeitsbäumen tragfähig, ist aber für keine Welle Trigger.
- Wird blockiert von: — (`welle-fundament` ist geschlossen).
- Unabhängig: [welle-record-replay](welle-record-replay.md). Berührungspunkte sind Dateien, nicht Verträge: `Dockerfile`, `Makefile`/`harness/mk/`, `harness/README.md` (Sensors-Tabelle), `tools/gegenproben/` und `SPEC-102`, die auch Slices von `welle-record-replay` ändern (zuletzt `slice-groessenmarke-werkzeug`); wer zuerst merged, hält die Zeilen des anderen im selben Commit nach (`AGENTS.md` §3.9).
- Reihenfolge der Slices (Pfeil heißt: braucht): `slice-dev-container-geruest` → `slice-dev-container-zugang` → `slice-dev-container-einlieferung` → `slice-dev-container-lauf` → `slice-dev-container-dokument`. Jeder braucht den Code des Vorgängers in `done/`: der Zugang sitzt im Container des Gerüsts, die Einlieferung braucht den Zugang, der Lauf braucht einen eingelieferten Stand, das Dokument beschreibt den ganzen Weg.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Eine CI (Server, Pipeline, Runner) — die Welle liefert den Arbeitscontainer für Entwickler und ihre Rollen, nicht die Pipeline des Projekts; `harness/README.md` führt „eine CI gibt es noch nicht“.
- Das Container-Image des Recorders als Auslieferung (`LH-QA-06`, Dienst als Container) — Produkt, Folgewelle mit dem Betrieb; der Arbeitscontainer baut das Produkt, er ist nicht dessen Image.
- Der Betrieb des Containers über `u-boot` hinaus (andere Hosts, macOS, Windows, Cloud) — nicht gefordert; ein anderer Host ist eine eigene Entscheidung.
- Die Orchestrierung der Agenten-Rollen im Container (Planner bis Verifier laufen darin, Aufträge, Berichte) — die Welle stellt den Ort bereit, nicht den Ablauf; die Rollen-Typen unter `.claude/agents/` bleiben, wie sie sind.
- Eine Schwelle oder ein Gate lockern, damit der Lauf im Container besteht — `AGENTS.md` §3.6; ein Fall, der im Container reißt, ist ein Befund (Zuschnitt des Falls), kein Anlass für eine längere Frist.
- Änderungen an `src/` und an Produktverträgen — die Welle berührt sie nicht.
- Backup, Wiederherstellung und Aufbewahrung der Bare-Repos des Containers — der Container ist ersetzbar, der Stand lebt im Klon des Entwicklers.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-dev-container-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).
