# Slice slice-dev-container-bau: Image, Pins, Bau und Bereitschaftsprobe des Arbeitscontainers (Teil I des Schnitts von `slice-dev-container-geruest`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-dev-container](../welle-dev-container.md).

**Bezug:** `LH-QA-06` (gleiches Testverhalten bei gleicher Konfiguration, Container), `LH-RB-03` (Entwicklung unter Linux und in Docker-Umgebungen), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Folgepflicht: der Arbeitscontainer fährt `make gegenproben`), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Vertrag unter `harness/vertraege/`), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Stufen des `Dockerfile`). Entscheidungen des Architect (2026-11-23): [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) (Privilegierung, Limits, Netz; `Accepted`, Entscheidung des User vom 2026-11-24).

**Berührte Spec-Stellen** (Teil I des Schnitts von [slice-dev-container-geruest](../open/slice-dev-container-geruest.md); die Stellen der Teile II und III stehen in deren Plänen): `SPEC-215` (Vertrag, [`SPEC-215-arbeitscontainer.md`](../../../../harness/vertraege/SPEC-215-arbeitscontainer.md): Nr. 1 bis 4, 6, 10, 11, 12 b, 12 d) samt Zeile im Index von `harness/vertraege/README.md` · `SPEC-102` (die Fälle zum Bau und zur Probe kommen als Nachzug des Implementers hinzu) · `AGENTS.md` §3.1 (Docker-only): bleibt unberührt, [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) Entscheidung, Einordnung nach §3.1 (Vorschlag).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Stand beim Start** (Planner, 2026-11-27, am Stand `8dfcf9b9`, Bedingung für Bedingung nach §4): `welle-fundament` liegt in `done/` ([Closure-Notiz](../done/welle-fundament-results.md)), ebenso `welle-record-replay` (Trigger von [welle-dev-container](../welle-dev-container.md) §2); [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) und [ADR-0032](../../adr/0032-schluesselverwaltung-des-arbeitscontainers.md) sind `Accepted`; `SPEC-215` samt Index-Zeile, die Tabelle *Zusage · Fänger* (Zeilen 1 bis 20) und das Raster stehen auf `main`, ohne Platzhalter außer den Pin-Werten, die der Implementer einträgt (`SPEC-215` Nr. 2 b); §6 führt keine offene Frage an den Architect; in `in-progress/` lag kein anderer Slice; der Vorgänger der Gate-Kette [slice-lasttest](../done/slice-lasttest.md) liegt in `done/`. **Reihenfolge gegen Teil D:** Der Vorschlag der letzten Closure, Teil D von [slice-bootstrap-m6](../next/slice-bootstrap-m6.md) vor diesem Slice einzureihen, setzt die Ausnahme des Users vom Welle-Trigger voraus; sie liegt nicht vor, Teil D startet nicht, und dieser Slice läuft vor ihm.

**Zurückgestellt** (`in-progress` → `open`, Planner, 2026-11-27): Der User hat am 2026-11-27 entschieden, dass die Welle den Meilenstein M7 verlässt und dieser Slice nach `open/` zurückgeht (Antwort 1, Wortlaut in [welle-betrieb](../welle-betrieb.md) §4). Es ist kein Code entstanden: der Zweig `slice-dev-container-bau` trägt keinen Commit über den Start `7a8361ad` hinaus; seinen Worktree räumt der Orchestrator ab. Beim erneuten Start gilt: die Angaben *Stand beim Start* oben sind Stand des ersten Starts; der Planner prüft die Bedingungen neu, setzt `basis`, `schaetzung` und `marke` neu und sichtet das Register neu.

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-dev-container-geruest](../open/slice-dev-container-geruest.md), Vorschlag des Architect vom 2026-11-23). **Datum:** 2026-11-24 (der Plan des ungeteilten Slice vom 2026-10-30 trug Ziel, Tabelle und Raster in einer Datei).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Image `odata-dev:arbeit` baut aus einem eigenen Kontext (`harness/dev`) mit Basis per Digest und Paketen mit Version (`dev-bau`, `SPEC-215` Nr. 2), die Bibliothek `dev-lib.sh` prüft die Variablen und die Aufruf-Form (Nr. 3, 4), und die Probe (`dev-probe`) prüft in fester Reihenfolge fünf Bedingungen an einem laufenden Container, lesend und durch Anhängen erweiterbar (Nr. 6); die Fälle der Gegenprobe mit einem Ersatz für `docker` belegen Argumente, Zähler, Exits und Ströme (Nr. 11). Der Teil startet keinen Container und legt auf dem Daemon des Hosts nichts an außer dem Image. **Schnitt vor dem Code** (Vorschlag des Architect, Planner, 2026-11-24): die Schätzung des ungeschnittenen Plans (*S* = 5100, Marke 6900, 54 Zeilen) liegt weit über der Schnittgrenze von `SPEC-128` Nr. 1; drei Teile, die nacheinander laufen und je einzeln lieferbar sind: **I** (diese Datei, Zeilen 1 bis 20), **II** [slice-dev-container-start](../open/slice-dev-container-start.md) (Start, Zeilen 21 bis 41) und **III** [slice-dev-container-geruest](../open/slice-dev-container-geruest.md) (Stopp und Handläufe auf `u-boot`, Zeilen 42 bis 54; der Teil trägt die Kennung des ungeteilten Slice, Abweichung vom Namensvorschlag, damit jede Zeile „`slice-dev-container-geruest` in `done/`“ in den Plänen davor und danach (`welle-dev-container`, `slice-dev-container-zugang`) „alle drei Teile in `done/`“ heißt, weil Teil III als letzter startet und die zwei anderen voraussetzt). <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) -->

**Entscheidungen des Architect** (2026-11-23, soweit dieser Teil sie trägt; der Wortlaut aller neun Aufträge und ihr Ausgang stehen in [slice-dev-container-geruest](../open/slice-dev-container-geruest.md) §1, der Vertrag trägt die Gründe, dieser Plan wiederholt sie nicht, `AGENTS.md` §3.8):

| Auftrag | Entscheidung | Ort |
|---|---|---|
| 1 Privilegierung | `--privileged` mit eigenem Daemon, ohne Docker-Socket, Bind-Mount, Host-Namensräume und Port; der innere Daemon hört nur auf dem Unix-Socket; `AGENTS.md` §3.1 bleibt unberührt. **Ausgang: entschieden** (User, 2026-11-24, [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) `Accepted`) | `SPEC-215` Nr. 5 c, d, 6, 9 |
| 3 Vertrag | Start, Probe (fünf Bedingungen, erweiterbar durch Anhängen), Stopp, wiederholter Start, Abbruch durch Signal | `SPEC-215` Nr. 4 bis 8 und 10 |
| 4 Image und Pins | ein eigenes Dockerfile `harness/dev/Dockerfile` mit eigenem Kontext (kein Eingriff in die Stufen der Gates), Basis per Digest, Pakete mit Version, andere Quellen verboten, geprüft von `dev-bau.sh` vor dem Bau; Bindung: Zeilen in der Tabelle der **Werkzeuge** von `harness/README.md`, nicht in `GATE_CHECKS` <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) --> | `SPEC-215` Nr. 1, 2 |
| 8 Prüfbarkeit | ein Ersatz für `docker` am Anfang von `PATH`; der echte Start belegt nur ein Handlauf auf `u-boot` (Teile II und III) | `SPEC-215` Nr. 11, 12 |

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Start des Containers mit Limits, Volumes, Netz und Bereitschaftswarten, die Variablen `DEV_MEMORY`, `DEV_CPUS`, `DEV_FRIST` — [slice-dev-container-start](../open/slice-dev-container-start.md) (Teil II): sie brauchen die Probe und die Bibliothek dieses Teils, nicht umgekehrt.
- Der Stopp, `DEV_ZUSTAND` und die Handläufe auf `u-boot` mit zwei Containern — [slice-dev-container-geruest](../open/slice-dev-container-geruest.md) (Teil III).
- SSH-Server, Schlüsselpaar und Host-Schlüssel — `slice-dev-container-zugang`; das Gerüst läuft ohne Zugang von außen, die Schlüsselverwaltung ist eine eigene Entscheidung mit eigener Prüfbarkeit.
- Bare-Repo, Klon und Einlieferung — `slice-dev-container-einlieferung`; sie brauchen den Zugang.
- Lauf von `make gates` und `make gegenproben` im Container — `slice-dev-container-lauf`; dieser Teil beweist weder, dass ein Daemon antwortet, noch, dass die Gates in ihm bestehen.
- Das Dokument für Entwickler — `slice-dev-container-dokument`.
- Ein Eingriff in die vorhandenen Gate-Stufen (`Dockerfile`: `test`, `build`, `race`, `lint`, `gegenproben`) — Bestand: der Container fährt sie, wie sie sind; ein Eingriff verschöbe die Prüfung ihrer Gates. Ausgenommen sind die Zeilen, die die neuen Treiber in die Stufe `gegenproben` kopieren und in `EINGABEN` von `tools/harness/gegenproben-eingabe.sh` und in die Positivliste von `.dockerignore` aufnehmen: sie ändern nicht, was die Stufe prüft, sondern was sie liest (Auftrag 8; Änderung des Architect am 2026-11-23, vom Planner am 2026-11-24 bestätigt, §6; sonst gäbe es keinen Fall der Gegenproben zum Gerüst).

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

- [ ] Die Entscheidungen 1, 2 und 6 des Architect stehen vor dem Code in [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) (Privilegierung, Limits, Netz) und die Entscheidungen 3 bis 5 im Vertrag `SPEC-215` unter `harness/vertraege/`; die ADR hat den Status `Accepted` (User, 2026-11-24), und der Vorbehalt A des Vertrags ist entfallen.
- [ ] Image, Pins und Bau: `harness/dev/Dockerfile` mit Digest und Paketpins, `dev-bau` prüft vor dem Bau Pin, Quelle und Aufruf und baut `odata-dev:arbeit` im Kontext `harness/dev`; Bibliothek `dev-lib.sh` mit `DEV_NAME`, Reihenfolge der Prüfungen, Strömen und Abbruch (Zeilen 1 bis 7 und 17 bis 20 der Tabelle in §3, vom Architect). <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) -->
- [ ] Probe nach dem Vertrag: fünf Bedingungen in fester Reihenfolge, lesend, dieselbe Lage gibt dieselbe Ausgabe; die erste verletzte Bedingung beendet den Durchgang (Zeilen 8 bis 16); zu jeder Zusage des Teils ein Fänger in der Tabelle *Zusage · Fänger* (§3) und ein Fall der Gegenproben; `make gegenproben` ist gefahren.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: je Ziel dieses Teils eine Zeile in der Tabelle der **Werkzeuge** von `harness/README.md` (Entscheidung des Architect, `SPEC-215` Nr. 1); `AGENTS.md` §3.1 bleibt unberührt, es sei denn, der User entscheidet es in einer eigenen ADR anders.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/vertraege/SPEC-215-arbeitscontainer.md`, Index in `harness/vertraege/README.md` | neu / update (Architect, geliefert) | Zusagen des Gerüsts, vor dem Code |
| `docs/plan/adr/` (Privilegierung, Limits und Netz; Schlüsselverwaltung) und ADR-Index | neu / update (Architect, geliefert) | Entscheidungen 1, 2 und 6; Status `Accepted` (User, 2026-11-24) |
| `harness/dev/Dockerfile`, `harness/dev/daemon.json` | neu | Image mit Pins, Adressbereich des inneren Daemons; eigener Kontext (Entscheidung 4) <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) --> |
| `tools/harness/dev-bau.sh`, `dev-probe.sh`, `dev-lib.sh`; `harness/mk/dev.mk` | neu | die Treiber und die Ziele `dev-bau` und `dev-probe` dieses Teils, nicht in `GATE_CHECKS` (`SPEC-215` Nr. 1); die Teile II und III hängen `dev-start` und `dev-stop` an dieselbe Datei <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) --> |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh`, `Dockerfile` (Stufe `gegenproben`, eine eigene `COPY`-Zeile je Treiber), `.dockerignore`, `SPEC-102` | update | Fälle mit Ersatz für `docker` (Entscheidung 8, `SPEC-215` Nr. 11); die Treiber sind Eingabe; der Stempel erst im Integrationsschritt |
| `harness/README.md` | update | je Ziel dieses Teils eine Zeile in der Tabelle der Werkzeuge (`SPEC-215` Nr. 12) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15), die Mengen, die Prüfbarkeit und die erwartete Größe stehen unten und sind vom Architect (2026-11-23) vor dem Übergang `next` → `in-progress` geschrieben; die Zeilen 1 bis 20 des ungeschnittenen Plans hat der Planner beim Schnitt im Wortlaut übernommen, nur der Verweis in Zeile 8 auf die Zeile 34 trägt die Angabe des Teils (jede Zeile steht in genau einem der drei Teile). Die Zeile nennt in Klammern den Ort im Vertrag; jede Mutation fährt der Implementer selbst (`AGENTS.md` §3.10).

**Tabelle *Zusage · Fänger*.** Jeder Fänger ist ein Fall der Gegenprobe mit einem Ersatz für `docker` (`SPEC-215` Nr. 11), wo nichts anderes steht; „Schleife“ heißt: ein Fall je Element der Menge (unten), die Zählprüfung nennt ihre Zahl. Handläufe und Negative sind benannt. Die Zeilen 21 bis 41 stehen in [slice-dev-container-start](../open/slice-dev-container-start.md), die Zeilen 42 bis 54 in [slice-dev-container-geruest](../open/slice-dev-container-geruest.md).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Keines der vier Ziele hängt an `GATE_CHECKS`; je Ziel eine Zeile in der Tabelle der Werkzeuge, Zelle höchstens 200 Zeichen (Nr. 1) | Zelle: `make docs-check` (`section-cell-oversized`); Kette: Handprobe `make -n gates` nennt keinen Treiber `dev-*.sh`, Zahl der Treffer 0 im Bericht; Existenz der Zeile und ihr Ort (Werkzeuge, nicht Gates) kein Fänger, Negativ Nr. 12 d | Zelle mit 201 Zeichen; `GATE_CHECKS += dev-start` | andere |
| 2 | Das Image heißt `odata-dev:arbeit`, der Kontext des Baus ist `harness/dev`, nie `.` (Nr. 2 a) | Fall `dev-bau/aufruf`: der Ersatz hält die Argumente von `docker build`; Schleife über Tag (genau `odata-dev:arbeit`) und Kontext (letztes Argument genau `harness/dev`) | Tag `odata-recorder:dev`; Kontext `.` <!-- d-check:ignore (geplanter Ordner, entsteht mit dem Code) --> | andere |
| 3 | Jede `FROM`-Zeile außer `FROM scratch` trägt `@sha256:` und genau 64 Ziffern `0-9a-f` (Nr. 2 b, c) | Schleife über M1: 10 Verstöße (Exit 2, `Pin fehlt: <Zeile>`) und 3 Gutfälle (Exit 0, `docker build` gerufen) | Länge ungeprüft (genau 64 zu mindestens 1); nur das erste `FROM` geprüft; Großbuchstaben als Ziffer | andere |
| 4 | Jedes Paket einer `apk add`-Zeile trägt `=<version>` (Nr. 2 b, c) | Schleife über M2: 8 Verstöße (Exit 2, `Pin fehlt: <Zeile>`) und 3 Gutfälle | nur das erste Paket geprüft; Fortsetzungszeile ungelesen; `name=` ohne Version gilt als Pin | andere |
| 5 | Das Dockerfile kennt nur Basis und `apk` als Quelle von Bytes; jede andere Form ist verboten (Nr. 2 b) | Schleife über M3: 9 Verbote (Exit 2, `Quelle ohne Pin: <Wort>`) und 2 Gutfälle (Wortgrenze) | je Wort der Liste aus der Prüfung genommen (8 Mutationen); Teilstring statt Wort | andere |
| 6 | Ein Pin-Verstoß, ein fehlendes Dockerfile und ein Aufruf mit Argument rufen kein `docker` (Nr. 2 c, 4 a) | die Fälle von 3 bis 5 plus Dockerfile fehlt und Argument (auch ein leeres): der Zähler des Ersatzes ist in jedem Fall der Schleife 0 | `docker info` vor der Prüfung | andere |
| 7 | Der Bau: gelungen Exit 0 und `Image odata-dev:arbeit gebaut`; scheitert `docker build` Exit 1 und `Bau gescheitert`; `docker` unerreichbar Exit 2 und `docker nicht erreichbar`; ein Dockerfile ohne `FROM` Exit 1 (Nr. 2 c, d, 4) | je Form ein Fall (4 Fälle); der Ersatz antwortet auf `info` und `build` mit dem Exit des Falls | Exit 0 trotz scheiterndem `build`; Meldung fehlt | andere |
| 8 | Die Probe prüft fünf Bedingungen in fester Reihenfolge; bestehen alle: fünf `ok`-Zeilen, `bereit`, Exit 0 (Nr. 6, 6 a) | Fall `dev-probe/bereit`: `zahl M4 "$n" 5`, Nummern und Namen der Zeilen in Reihenfolge; dieselbe Bibliothek ruft der Start (Zeile 34, Teil II) | eine Bedingung aus der Liste genommen; zwei vertauscht | andere |
| 9 | Die erste verletzte Bedingung beendet den Durchgang: Exit 1, `ok` für alle davor, `nicht erfüllt` für diese, nichts für spätere; bei zwei verletzten gilt die frühere (Nr. 6 a, Zusatztabelle) | Schleife über k = 1 bis 5 und über alle zehn Paare (k, j) mit k < j; der Ersatz zählt je Bedingung ihre Aufrufe: hinter k keiner | Probe läuft alle Bedingungen und meldet die letzte; falsche Nummer | andere |
| 10 | Bedingung 1 `container`: nicht vorhanden, nicht laufend, ohne Label, Label eines anderen Namens sind verletzt (Nr. 6) | Schleife über M5: 4 Verstöße, 1 Gutfall | Label nicht geprüft; `Running` nicht geprüft | andere |
| 11 | Bedingung 2 `daemon`: Exit ungleich 0, leere Antwort und nur Leerraum sind verletzt (Nr. 6) | Schleife über M6: 3 Verstöße, 1 Gutfall | Antwort nicht gekürzt; Exit ignoriert | andere |
| 12 | Bedingung 3 `eigener-daemon`: gleiche Kennung wie der Host, leere innere und leere Kennung des Hosts sind verletzt (Nr. 6) | Schleife über M7: 3 Verstöße, 1 Gutfall; `info` des Hosts und des Containers antworten getrennt | Vergleich entfernt; leer gilt als ungleich | andere |
| 13 | Bedingung 4 `limit`: Speicher, Swap und CPU gelten im Container wie angefordert; ungesetzt oder `max` ist verletzt (Nr. 6) | Schleife über M8: 10 Verstöße (je Datei ungleich und `max`, dazu `Memory` 0, `NanoCpus` 0, `MemorySwap` ungleich `Memory`, `cat` scheitert), 1 Gutfall; zwei Größen (10g mit 8 CPUs und 512m mit 1 CPU), die Sollwerte aus Rechnung (Bytewerte, `NanoCpus ÷ 10000`) | je Vergleich entfernt (3 Mutationen); `max` als Zahl gelesen | andere |
| 14 | Bedingung 5 `kein-tcp`: ein lauschender Port 2375 oder 2376 ist verletzt, ein Port mit anderer Zahl nicht (Nr. 6) | Schleife über M9: 4 Verstöße, 4 Gutfälle (`:23750`, `:12375`, `:2377`, keine Zeile) | Teilstring-Suche statt Portgrenze; nur 2375 | andere |
| 15 | Die Probe ist lesend: erlaubt sind `container inspect`, `exec` mit den Befehlen der Bedingungen 2 bis 5 und `info` (Nr. 6) | in jedem Fall von 8 bis 14 und in jeder Lage von M10 prüft der Ersatz jeden Aufruf gegen die Menge der erlaubten Formen (Allaussage); Zähler von `run`, `create`, `start`, `stop`, `kill`, `restart`, `rm`, `cp`, `volume`, `network`, `build`, `pull`: je 0 | `docker restart` oder `exec <n> touch` hinzugefügt | andere |
| 16 | Dieselbe Lage gibt dieselbe Ausgabe, denselben Exit und dieselbe Aufrufliste (Nr. 6 b) | Schleife über M10 (6 Lagen): zwei Läufe hintereinander, Vergleich aller drei | ein Zähler oder eine Zeit in der Ausgabe | andere |
| 17 | `DEV_NAME` ist `[a-z][a-z0-9-]{0,30}`; jede Abweichung ist Exit 2 ohne Aufruf von `docker`; die Probe ist der erste Treiber dieser Achse, Start und Stopp kommen in ihren Teilen hinzu (Nr. 3, 4 a) | Schleife über M11 (10 Verstöße, 3 Gutfälle) × die vorhandenen Treiber; `zahl M11 "$n" 10` und die Zahl der Treiber | Länge 31 zu 32 (beide Seiten der Grenze); Großbuchstaben erlaubt | andere |
| 18 | INT, HUP und TERM an `dev-bau` (während `docker build`) und an `dev-probe` (während `inspect` und `exec`) enden mit Exit 2 und `abgebrochen`, ohne weiteren Aufruf (Nr. 10) | Schleife über 3 Phasen × 3 Signale = 9 (M19, Teil Bau und Probe); der Ersatz schläft 1 s im Aufruf, der Fall sendet das Signal und zählt die Aufrufe danach | Handler fehlt (Exit 143); nur INT behandelt | nebenläufig |
| 19 | Reihenfolge: Argument vor Variable vor `docker` (Nr. 4 a); bei zwei Verstößen gilt der frühere | Paare (Argument, ungültiger Name) und (ungültiger Name, `docker` unerreichbar) je Treiber `dev-bau` und `dev-probe`, soweit er die Variable liest; Zähler `docker` 0 | Reihenfolge vertauscht | andere |
| 20 | Befunde stehen auf stderr, Erfolgs- und Fortschrittszeilen auf stdout, jede Zeile mit dem Präfix `dev-<Ziel>: ` (Nr. 4 c) | in jedem Fall der Zeilen 3 bis 19 liest der Fall beide Ströme getrennt und prüft Strom und Präfix jeder Zeile (Allaussage über alle Meldungen); `docker`-Ausgabe unverändert am Strom des Aufrufs | eine Meldung auf dem falschen Strom; Präfix fehlt | andere |

Zeilen: 1 bis 20 gleich 20 (*a* = 1: Zeile 18; *b* = 19). Die Klasse `nebenläufig` setzt der Architect für Zeilen, deren Code oder Fänger nebenläufig ist (`SPEC-128` Nr. 3 e); Zeile 18 ist keine Wahrscheinlichkeit (der Ersatz hält den Aufruf fest, das Signal kommt danach), der Implementer fährt ihre Mutationen zwanzigmal (`-count=20`) und nennt die Rate (`AGENTS.md` §3.10). Größe und Schnitt: §4.

**Mengen** (`AGENTS.md` §3.15, §3.17): Achsen und Zählprüfung. Der Test schreibt jede als Schleife über Verstöße und Gutfälle und prüft die Zahl (`zahl M<n> "$n" <k>`); ein Element, das ein Verstoß ist, ist Exit 2 (Pin, Variable) oder Exit 1 (Bedingung) mit der Meldung des Vertrags; ein Gutfall kommt durch.

| Menge | Elemente | Achsen |
|---|---|---|
| M1 `FROM` | 10 Verstöße (Tag ohne Digest; weder Tag noch Digest; Digest mit 63 und mit 65 Zeichen; Großbuchstabe im Digest; `@sha256:` ohne Ziffern; `@sha512:` mit 128 Zeichen; zweites `FROM` ohne Digest; `--platform=…` vor einem Image ohne Digest; Nicht-Hex-Zeichen `g` im Digest), 3 Gutfälle (gültig; gültig mit `--platform=…`; `FROM scratch`) | Form des Digests, Stelle (erste oder spätere Zeile), Flag |
| M2 `apk add` | 8 Verstöße (`bash`; `bash=`; `bash=1.0 git`; `git bash=1.0`; Fortsetzungszeile; hinter `&&`; zweite `apk add`-Zeile ohne; `--no-cache bash`), 3 Gutfälle (zwei Pakete mit Version und Option; zwei Zeilen mit Version; keine `apk`-Zeile) | Stelle des Pakets, Zeilenform, Option |
| M3 Quellen | 9 Verbote (`curl`, `wget`, `pip`, `npm`, `apt-get`, `go install`, `git clone`, `ADD http://`, `ADD https://`), 2 Gutfälle (`pipeline` als Teilwort; `ADD x /y`); 8 Formen, `zahl M3 "$n" 8` für die Formen | Wort, Zeilenform (`RUN`, `ADD`) |
| M4 Bedingungen | 5 (`container`, `daemon`, `eigener-daemon`, `limit`, `kein-tcp`) | Reihenfolge |
| M5 bis M9 | M5 4+1, M6 3+1, M7 3+1, M8 10+1, M9 4+4 (Verstöße plus Gutfälle, wie in den Zeilen 10 bis 14) | je Bedingung ihre Formen: Zustand des Containers, Antwort, Vergleichsseite, Datei und Abweichung, Schreibweise der Adresse |
| M10 Lagen der Probe | 6 (bereit; verletzt 1 bis 5) | Lage × Treiber |
| M11 `DEV_NAME` | 10 Verstöße (leer; `A`; `aB`; `1a`; `-a`; `a_b`; `a.b`; `a b`; `a/b`; 32 Zeichen), 3 Gutfälle (`a`; 31 Zeichen; `a-1`) | Form, Länge (beide Seiten der Grenze 31 und 32), Treiber (Probe, Start, Stopp) |
| M19 Abbruch, dieser Teil | 9 (3 Phasen: Bau 1, Probe 2, × 3 Signale); die 24 weiteren Elemente der Menge von 33 liegen in den Teilen II (12) und III (12) | Treiber × Phase × Signal |

**Zusatztabelle** (`AGENTS.md` §3.16), Paare von Ursachen (je Paar die Entscheidung und der Fänger; die Reihenfolge der Prüfungen steht in `SPEC-215` Nr. 4 a); die Zeilen der Zusatztabelle des ungeschnittenen Plans, soweit Fänger dieses Teils sie tragen:

| Paar | Entscheidung | Fänger dieses Teils |
|---|---|---|
| Argument und ungültige Variable | der Aufruf-Fehler (Nr. 4 a) | Zeile 19 |
| ungültige Variable und `docker` unerreichbar | die Variable, kein Aufruf von `docker` (Nr. 4 a) | Zeile 19 (Zeile 26 in Teil II für den Start) |
| Limit gesetzt (Bedingung 4 verletzt) und Daemon nicht bereit (Bedingung 2 verletzt) | die frühere Bedingung (2) wird gemeldet (Nr. 6 a) | Zeile 9 (alle zehn Paare) |

**Prüfbarkeit** (Auftrag 8). *Im Gate:* die Fälle der Gegenproben rufen die Treiber im Wegwerfverzeichnis mit einem Ersatz für `docker` am Anfang von `PATH` (`SPEC-215` Nr. 11; Muster der Fälle 103 bis 114 von `SPEC-102`); sie belegen Reihenfolge, Argumente von `docker`, Zähler, Exit-Codes und Ströme, nicht den Daemon. Der Implementer misst die Laufzeit von `make gegenproben` nach der Messvorschrift von `SPEC-102` (Soll nach [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)) und trägt sie ein; der Fall der Zeile 18 fährt der Implementer jede Mutation zwanzigmal (`-count=20`). *Nur ein Lauf auf `u-boot`:* gibt es in diesem Teil nicht (kein Container startet); die Handläufe liegen in den Teilen II und III. *Akzeptierte Negative* (Grund und Folge im Vertrag, `SPEC-215` Nr. 12): 12 b (die Antworten des Ersatzes für `docker`) und 12 d (die Verdrahtung der Ziele) treffen diesen Teil; die Zeile in „Nicht behauptet“ von `harness/README.md` schreibt Teil III.

**Erwartete Größe** (Architect, 2026-11-23, für die Marke des Planner nach `SPEC-128` Nr. 1; Produktcode in Zeilen von Skripten, Dockerfile und Fragment ohne Gegenproben und Tests): Teil I rund 385 (`dev-bau.sh` 100, `dev-lib.sh` 90, `dev-probe.sh` 130, `harness/dev/Dockerfile` 30, `daemon.json` 8, `dev.mk` 15, `harness/README.md` 12). Die Zeile `Dockerfile`, `.dockerignore` und `gegenproben-eingabe.sh` zählt einige Zeilen und steckt in *p*. Die Formel ist an Go-Slices geeicht; an Skripten belegt sie hier nur `slice-guard-schreibwege` (Messung 1707 von Marke 2600, Geltungsgrenze `SPEC-128` Nr. 2); die Fälle der Gegenproben sind Go-Code mit Fixtures und Shell-Ersatz. <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) -->

Betriebsregeln für alle Rollen dieses Slice: Repo-Dateien nur mit `Edit` und `Write` (kein `sed -i`, `awk`, Python, Heredoc in die Datei; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); jeder `docker run` einer Rolle mit `--memory` und `--memory-swap`, höchstens zwei Container zugleich (Erfahrung vom 2026-10-24: Mutationskopien belegten ein Vielfaches des Bedarfs); Werte mit Escape-Folge oder Rohbytes aus Bytewerten und gegengelesen (`AGENTS.md` §3.18).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-fundament` liegt in `done/`; [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) hat den Status `Accepted` (Entscheidung des User vom 2026-11-24, erfüllt; der Vertrag trägt sie, der Vorbehalt A ist entfallen); der Vertrag `SPEC-215`, die Tabelle *Zusage · Fänger* und das Raster aus §6 stehen vor dem ersten Code-Commit (Architect, 2026-11-23, geliefert); der Planner hat die Marke gesetzt; kein anderer Slice der Gate-Kette liegt in `in-progress/` (Übergangsregeln Nr. 7: dieser Teil ändert `harness/mk/` (neues Fragment), `Dockerfile`, `.dockerignore`, `harness/README.md`, `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` und den Stempel); Reihenfolge der Kette: nach [slice-container-image](slice-container-image.md) in `done/` (Planner, 2026-11-27, nach der Entkopplung; vorher: nach [slice-lasttest](../done/slice-lasttest.md); die Reihenfolge der Kette steht im Drift-Log der Roadmap, das Auslassen eines Vorgängers regelt Nr. 7 der Übergangsregeln). Teil II ([slice-dev-container-start](../open/slice-dev-container-start.md)) startet, wenn dieser Teil in `done/` liegt, Teil III ([slice-dev-container-geruest](../open/slice-dev-container-geruest.md)) nach Teil II.

**Warum:** Teil II ruft die Probe und die Bibliothek dieses Teils (*Code-Abhängigkeit*), Teil III den Treiber und die Bibliothek beider (*Code-Abhängigkeit*); die Reihenfolge I, II, III folgt den Schnittstellen.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1; Teil des Architect, 2026-11-23, im Block unten; der Planner hat die Zahlen am 2026-11-24 aus der Tabelle nachgerechnet, **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 20 Zeilen der Tabelle (1 bis 20), *a* = 1 (Zeile 18), *b* = 19, *f* = 0 (der Ersatz für `docker` ist ein Skript im Fall, kein Port-Fake), *p* = 385: *S* = 85·1 + 80·19 + 200·0 + 385 = 85 + 1520 + 385 = 1990, 135·1990 = 268650, **Marke 2700** = 100 · ⌈26,865⌉; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`; der Teil liegt 233 unter der Schnittgrenze, jede weitere Zeile kostet 80 (*b*) oder 85 (*a*), die dritte weitere Zeile löst die Schnittfrage aus (*S* = 2230 mit drei Zeilen *b*). Summe der drei Teile: 1990 + 1936 (Teil II) + 1175 (Teil III) = 5101; der ungeschnittene Plan (5100) zählte die Zeile von `dev-start` in `harness/README.md` nicht, die Abweichung ist diese eine Zeile (Architect, 2026-11-24, [slice-dev-container-start](../open/slice-dev-container-start.md)). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` zählt nicht, `docs/` und `spec/` nennt kein Pfad.

```groessenmarke
pfad: harness/dev/Dockerfile
pfad: harness/dev/daemon.json
pfad: tools/harness/dev-bau.sh
pfad: tools/harness/dev-lib.sh
pfad: tools/harness/dev-probe.sh
pfad: harness/mk/dev.mk
pfad: tools/gegenproben
pfad: tools/harness/gegenproben-eingabe.sh
pfad: Dockerfile
pfad: .dockerignore
pfad: harness/README.md
produktcode: 385
fakes: 0
```

**Zurückgenommen** (Planner, 2026-11-27, mit der Rückführung): `basis`, `schaetzung` und `marke` des ersten Starts (`8dfcf9b9`, 1990, 2700) gelten nicht mehr; der Planner setzt sie beim erneuten Start vor dem ersten Code-Commit. Die Formel ist für Skripte mit Prozessen schwach belegt (§8, `BEO-REPO/schaetzformel-unterschaetzt-werkzeug-mit-prozessen`): eine Messung über der Marke ist nach dem Code mit Messung und Grund zu berichten (`AGENTS.md` §3.14), nicht durch einen Schnitt beim Handoff zu heilen.

Der Block nennt `harness/dev` als zwei Dateien (der Architect nannte das Verzeichnis; es gehört diesem Teil allein, und die zwei Dateien sind bekannt), `tools/gegenproben` als Verzeichnis, weil die Namen der neuen Fälle vor dem Code nicht feststehen: der Implementer nennt jede Datei, die er dort anlegt, im selben Commit im Block (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`). `Dockerfile`, `.dockerignore`, `harness/README.md` und `tools/gegenproben` ändern auch andere Slices der Gate-Kette; sie laufen nacheinander (Übergangsregeln Nr. 7), und der Planner prüft beim Start, dass die Basis hinter dem letzten Slice der Kette liegt (§6, Eintrag *fremde Zeilen*). Der Stempel steht nicht im Block (Integrationsschritt). <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) -->

**Reihenfolge und Kollisionen** (Architect, 2026-11-23; die Reihenfolge der Kette legt der Planner fest, Roadmap, Drift-Log 2026-11-24). Dieser Teil ändert die Pfade der Kette (siehe Start); die Pläne unter `open/`, die dieselben Pfade ändern, laufen nacheinander mit ihm. Dateien mit Konflikt auf derselben Zeile: die `COPY`-Zeilen der Stufe `gegenproben` im `Dockerfile`, die Liste `EINGABEN` und die Positivliste von `.dockerignore`; der Implementer legt je Treiber eine eigene Zeile an und ändert keine fremde. `harness/mk/dev.mk` ist eine neue Datei; das `Makefile` bindet `harness/mk/*.mk` ein und bleibt unverändert. <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein weiterer Schnitt ist nicht vorgesehen, ein Teil, der die Schnittgrenze nach dem Code überschreitet, wird mit Messung und Grund berichtet (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert): [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) wird `Rejected` oder `Superseded`, oder `u-boot` erlaubt den gewählten Weg nicht (Befund an den Planner, die Welle wird umgeplant, Drift-Log).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, je Zusage des Teils ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor. Der Beleg eines Starts auf `u-boot` mit Zahlen gehört zu den Teilen II und III.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Sicherheit: Ein privilegierter Container mit eigenem Daemon vergrößert die Angriffsfläche des Hosts — **Ausgang:** entschieden (User, 2026-11-24). [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) (`Accepted`) benennt Wirkung, sieben Alternativen und die Begrenzung (kein Socket, kein Bind-Mount, keine Host-Namensräume, kein Port).
- Konflikt mit `AGENTS.md` §3.1: der Wortlaut „Host braucht nur Docker und GNU make“ gilt für Prüfläufe — **Ausgang:** entschieden (User, 2026-11-24, Entscheidung 2 in [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md), vom Planner beim Start am 2026-11-27 nachgetragen): §3.1 bleibt unberührt (der Arbeitscontainer ist kein Prüflauf, die Prüfläufe in ihm bleiben netzlos und ohne Mount); eine Ergänzung der harten Regel wäre eine eigene ADR.
- [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) hat den Status `Accepted`, Bedingung jedes Teils (§4) — **Ausgang:** entfallen (User, 2026-11-24, die Bedingung ist erfüllt).
- Die Werte der Pins (Digest der Basis, Versionen der Pakete) hat der Architect nicht abgerufen (netzlos); der Implementer trägt sie ein — **Ausgang:** offen; sind sie nicht erfüllbar (Paket nicht mehr geführt), ist es ein roter Bau des Implementers, kein Anlass, den Pin zu lockern.
- Der Architect hat den Ausschluss der Gate-Stufen in §1 um die Zeilen für die Stufe `gegenproben` (`COPY`, `EINGABEN`, `.dockerignore`) ergänzt, damit ein Fall der Gegenproben zum Gerüst entsteht — **Ausgang:** entschieden (Planner, 2026-11-24): bestätigt. Die Zeilen ändern, was die Stufe `gegenproben` liest, nicht was sie prüft (kein Gate wird gelockert, `AGENTS.md` §3.6); ohne sie hätten die Zeilen 3 bis 50 keinen Haltepunkt in der Kette (der Vertrag Nr. 11 verlangt Fälle mit Ersatz für `docker`, Muster der Fälle 103 bis 114 von `SPEC-102`); je Treiber eine eigene Zeile, keine fremde geändert (§4). Das gilt für alle drei Teile.
- Fremde Zeilen im Block (`SPEC-128` Nr. 10 g): `Dockerfile`, `.dockerignore`, `harness/README.md` und `tools/gegenproben` ändern auch andere Slices der Kette; ein Merge von `main` in den Zweig bringt deren Zeilen in die Messung — **Ausgang:** offen; Adresse: Planner (beim erneuten Start die Basis hinter dem letzten Slice der Kette; beim Handoff die Zeilen aus `main` trennen und mit Grund berichten).
- Die Stufe `gegenproben` braucht neben den Treibern auch `harness/dev/Dockerfile`: die Fälle laufen gegen Kopien beider (`SPEC-215` Nr. 11), der Ausschluss in §1 nennt nur die Treiber — **Ausgang:** entschieden (Planner, 2026-11-27, aus dem Wortlaut des Vertrags, keine neue Entscheidung): die Zeile in `COPY`, `EINGABEN` und `.dockerignore` für `harness/dev/Dockerfile` gehört zu denselben Zeilen wie die der Treiber; sie ändert, was die Stufe liest, nicht was sie prüft. <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* unten; jede Zelle trägt einen Ort im Vertrag oder „gilt nicht“ mit Grund (Architect, 2026-11-23); der Implementer hält an einer Randform an, die dort nicht steht — **Ausgang:** weiter offen bis zur Closure.
- Folgen für die anderen Pläne der Welle (§3.13): `slice-dev-container-zugang` nennt in §1 und DoD „eine ADR zur Schlüsselverwaltung“ und die erweiterbare Probe (Auftrag 3 dort) — **Ausgang:** entschieden (Planner, 2026-11-24): der Kopf und der Start-Trigger des Zugangs tragen [ADR-0032](../../adr/0032-schluesselverwaltung-des-arbeitscontainers.md) (`Accepted`, User, 2026-11-24) und `SPEC-215` Nr. 6 c, der Start nennt die drei Teile.
- Der Schnitt vor dem Code (Schätzung des ungeschnittenen Plans 5100, Marke 6900, über der Schnittgrenze von 2223) — **Ausgang:** entschieden (Planner, 2026-11-24, nach dem Vorschlag des Architect): Schnitt in drei Teile mit den Marken 2700 (dieser Teil), 2700 und 1600, nachgerechnet aus der Tabelle (keine Abweichung bei den Zahlen); Teil III trägt die Kennung `slice-dev-container-geruest`; `basis`, `schaetzung` und `marke` setzt der Planner je Teil beim Start (§4).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23; die Zeilen dieses Teils; die Zeilen der Teile II und III stehen in deren Plänen). Die Zelle nennt den Ort in `SPEC-215` oder „gilt nicht“ mit Grund; „Variable“ meint die Variablen von Nr. 3, `DEV_NAME` hat jede Operation.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Bau (`dev-bau`) | Dockerfile fehlt: Nr. 2 c (Exit 2, `harness/dev/Dockerfile fehlt`); Variablen: gilt nicht, der Bau liest keine | Dockerfile ohne `FROM`: kein Pin-Verstoß, der Bau scheitert, Exit 1 (Nr. 2 c, d) | Pin fehlt, Quelle ohne Pin, Argument des Treibers: Nr. 2 b, c und 4 a (Exit 2, kein `docker`); Daemon unerreichbar: Nr. 4 a | INT, HUP, TERM: Nr. 10 (Exit 2, nichts aufzuräumen) | zweiter Bau: Nr. 2 d (setzt den Tag neu, ein laufender Container behält sein Image) |
| Probe (`dev-probe`) | Container nicht vorhanden: Nr. 6 a (Bedingung 1 verletzt, Exit 1); Variable nicht gesetzt: Vorgabe | Antwort ohne Inhalt, Exit ohne Ausgabe: Nr. 6 (Bedingungen 2, 3, 5); `DEV_NAME` leer: Nr. 3 | Daemon einer fremden Version: gilt nicht, die Version legt der Pin des Images fest (Nr. 2 b) und die Bedingung 2 prüft die Antwort, nicht die Nummer; Container ohne Label, Daemon des Hosts, Limit gelockert, TCP-Port: Nr. 6 (Bedingungen 1, 3, 4, 5) | INT, HUP, TERM: Nr. 10 (Exit 2) | Probe auf bereitem Container: Nr. 6 b (gleiche Ausgabe, gleicher Exit) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-24; der Planner zieht Treffer und Zähler beim Start nach). Für diesen Teil zählen: `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) — die Betriebsregeln stehen in §3, ein Auftreten wird als Beleg eingetragen; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, `offen`) — Raten und Zeiten nennen Limit, Last und Messenden. Die Erfahrung vom 2026-10-24 (Speichermangel des Hosts durch Mutationskopien, Container ohne Limit) hat **keinen Eintrag im Register**; sie liegt als Notiz des Users vor und ist Anlass dieser Welle, nicht gezählt. Keiner der offenen erreicht mit diesem Teil 3×.

Beim Start erneut gesichtet (Planner, 2026-11-27, Zähler = Dateien unter `evidence/`; die Zahlen und der Schlusssatz oben sind der Stand des Schnitts). Offen und für den Teil einschlägig: `BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache` (20×, ohne Zielort): der Bericht nennt, ob die Stufe `gegenproben` frisch lief; `BEO-REPO/schaetzformel-unterschaetzt-werkzeug-mit-prozessen` (2×): Ersatz-Skript für `docker` und Signal-Fälle sind die beschriebene Form, **eine Messung über der Marke aus diesem Grund wäre der dritte Beleg** und damit eine Lücke mit eigenem Ausgang bei der Closure; `BEO-REPO/allaussage-an-jeder-stelle-mit-faenger-an-erster-und-zweiter-stelle` (5×): Zeilen 15 und 20 sind Allaussagen; `BEO-REPO/image-tag-race-von-allen-worktrees-geteilt` (2×): `odata-dev:arbeit` ist ein fester Tag im Daemon des Hosts, `make dev-bau` läuft unter der Sperre der Übergangsregeln. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (23×), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (15×, `AGENTS.md` §3.10), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (88×, `AGENTS.md` §3.15). Geplant: `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×, §6).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
