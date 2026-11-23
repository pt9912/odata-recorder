# Slice slice-dev-container-geruest: Gerüst des Arbeitscontainers

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-dev-container](../welle-dev-container.md).

**Bezug:** `LH-QA-06` (gleiches Testverhalten bei gleicher Konfiguration, Container), `LH-RB-03` (Entwicklung unter Linux und in Docker-Umgebungen), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Folgepflicht: der Arbeitscontainer fährt `make gegenproben`), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Vertrag unter `harness/vertraege/`), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Stufen des `Dockerfile`). Entscheidungen des Architect (2026-11-23): [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) (Privilegierung, Limits, Netz; `Proposed`, die Entscheidung steht beim User aus) und [ADR-0032](../../adr/0032-schluesselverwaltung-des-arbeitscontainers.md) (Schlüsselverwaltung, `Proposed`; gehört dem Zugang, nicht diesem Slice).

**Berührte Spec-Stellen:** `SPEC-215` (neuer Vertrag, [`SPEC-215-arbeitscontainer.md`](../../../../harness/vertraege/SPEC-215-arbeitscontainer.md)) samt Zeile im Index von `harness/vertraege/README.md` · `SPEC-102` (die Fälle zum Gerüst kommen als Nachzug des Implementers hinzu, Auftrag 8) · `AGENTS.md` §3.1 (Docker-only): bleibt unberührt, [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) Entscheidung, Einordnung nach §3.1 (Vorschlag).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Auf dem Host `u-boot` startet ein Container aus einem gepinnten Image mit eigenem Docker-Daemon unter festem Speicher- und CPU-Limit und meldet erst nach einer Bereitschaftsprobe, dass der Daemon in ihm Befehle annimmt; ein Start, der nicht bereit wird, endet nach einer Frist rot.

**Auftrag an den Architect** (Entscheidungen vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Planner entscheidet sie nicht):

1. **Privilegierung** (ADR-Kandidat, sicherheitsrelevant): Ein Docker-Daemon im Container braucht erweiterte Rechte. Der Architect vergleicht die Wege (privilegierter Container, Daemon ohne Root im Container, Laufzeit mit eigenem Kernel-Namensraum, Verzicht auf den Container zugunsten einer Maschine) nach Wirkung auf den Host, Aufwand auf `u-boot` und dem, was `AGENTS.md` §3.1 verlangt (kein lokaler Toolchain-Install; der Host braucht nur Docker und GNU make) — und nennt, ob §3.1 ergänzt werden muss (das wäre eine Änderung der harten Regel mit ADR, nicht stillschweigend). Zu entscheiden ist auch, was der Container auf dem Host sehen darf (Sockets, Mounts, Netz) und was nicht.
2. **Limits**: Wert und Ort von Speicher-, Swap- und CPU-Grenze (Befehlszeile des Außencontainers oder Konfiguration des Daemons), und dass die Grenze auch BuildKit im inneren Daemon trifft, der sich vom Host aus nicht begrenzen lässt; das Verhalten bei Überschreitung (Container endet, Host unberührt) ist eine Zusage mit Fänger.
3. **Vertrag** unter `harness/vertraege/`: Start, Bereitschaftsprobe (was „bereit“ heißt, Frist, Exit-Codes, Wiederholung, Ausgabe), Stopp, wiederholter Start, Abbruch durch Signal. Die Probe ist **erweiterbar**: spätere Slices der Welle fügen Bedingungen hinzu (Zugang), ohne den Vertrag zu ersetzen.
4. **Image und Pins**: Basis per Digest, Versionen der Pakete, Ort der Stufe oder des Skripts (`Dockerfile`-Stufe, `harness/mk/`-Fragment, `tools/harness/`); Bindung in `harness/README.md` (Sensors-Tabelle oder Werkzeuge-Tabelle; ein Ziel, das privilegiert startet, gehört nicht in `GATE_CHECKS` — der Architect entscheidet und begründet es wie [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)).
5. **Zustand**: Was ein Neustart behält (Daemon-Verzeichnis mit Image-Cache für `deps`, Bare-Repos) und was nicht; ein frischer Start ohne Zustand muss möglich sein (Closure-Kriterium der Welle).
6. **Netz**: Die Prüfläufe im inneren Daemon sind netzlos (`--network none`), nur `deps` holt aus dem Netz; welches Netz der Außencontainer bekommt (Spiegel, Proxy, nur Ausgang) ist eine Entscheidung mit Grund.
7. **Raster und Tabelle**: das Raster *Operation × Argumentart* (Operationen Start, Probe, Stopp; Argumente Limit fehlend, leer, null, unbekannt, größer als der Host; Name des Containers schon vergeben; Daemon startet nicht; Abbruch während des Starts; Wiederholung) in §6, jede Zelle mit Ort im Vertrag (`AGENTS.md` §3.16), und die Tabelle *Zusage · Fänger* in §3 (`AGENTS.md` §3.15) — beides vor dem Übergang `next` → `in-progress`.
8. **Prüfbarkeit**: wie ein Fänger ohne privilegierten Start im Gate lebt (Ersatz für `docker` in einer Gegenprobe nach dem Muster der Fälle 103 bis 114 von `SPEC-102`, Fälle in `tools/gegenproben/`) und welcher Teil nur ein Lauf auf `u-boot` belegt; Letzteres steht als akzeptiertes Negativ mit Grund und Folge in §3 und in „Nicht behauptet“ von `harness/README.md`.
9. **Größe**: Schätzung und Marke nach `AGENTS.md` §3.14 (Zeilen der Tabelle, Klassen, Port-Fakes entfallen hier; der Fake ist ein Ersatz-Skript), die der Planner nach dem Vertrag mit dem Werkzeug setzt; wird die Grenze 3000 überschritten, schneidet der Planner vor dem Code.

**Ausgang des Architect** (2026-11-23; Vertrag `SPEC-215`; `AGENTS.md` §3.12, §3.16). Die Entwürfe der zwei ADRs stehen `Proposed`: sie berühren die Sicherheit des Hosts und gehören dem User; jede Stelle des Vertrags, die an [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) hängt, trägt den Vermerk *Vorbehalt A*.

1. **Privilegierung** — [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md), Entscheidung (Vorschlag): `--privileged` mit Daemon im Container, aber ohne Docker-Socket, Bind-Mount, Host-Namensräume und veröffentlichten Port; der innere Daemon hört nur auf dem Unix-Socket. Verglichen sind sieben Wege (gezielte Fähigkeiten, Daemon ohne Root, Laufzeit mit eigenem Kernel, Maschine, Socket des Hosts, nichts tun). `AGENTS.md` §3.1 bleibt unberührt: der Wortlaut gilt für Prüfläufe, und die bleiben in den inneren Containern netzlos und ohne Mount; eine Ergänzung wäre eine eigene ADR. Im Vertrag: Nr. 5 c, d, Nr. 6 (Bedingungen 3 bis 5), Nr. 9. **Ausgang: offen, Adresse User.**
2. **Limits** — [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) und `SPEC-215` Nr. 9: Ort ist die Befehlszeile des Außencontainers (die cgroup trifft BuildKit und jeden inneren Container, Bedingung 4 liest die Grenze im Container nach); Vorgaben 10 GiB, kein Swap, 8 CPUs, 8192 Prozesse. **Korrektur an der Formulierung des Plans:** bei Überschreitung beendet der Kernel Prozesse im Container und lässt den Host unberührt; ob der Container endet, hängt davon ab, ob der Kernel den ersten Prozess wählt (`SPEC-215` Nr. 9 b). „Beendet den Container“ in §2 und im Closure-Trigger der Welle gilt nur, wenn der User das Gruppenende wählt (Frage 4 der ADR). **Ausgang: offen, Adresse User; der Planner zieht §2 und die Welle nach.**
3. **Vertrag** — `SPEC-215` Nr. 4 bis 8 und 10: Start, Probe (fünf Bedingungen, erweiterbar durch Anhängen, Nr. 6 c), Stopp, wiederholter Start, Abbruch durch Signal.
4. **Image und Pins** — `SPEC-215` Nr. 2: ein eigenes Dockerfile `harness/dev/Dockerfile` mit eigenem Kontext (kein Eingriff in die Stufen der Gates), Basis per Digest, Pakete mit Version, andere Quellen verboten, geprüft von `dev-bau.sh` vor dem Bau. Bindung: Zeilen in der Tabelle der **Werkzeuge** von `harness/README.md`, nicht in `GATE_CHECKS` (`SPEC-215` Nr. 1, Alternativen und Grund dort). <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->
5. **Zustand** — `SPEC-215` Nr. 8: zwei benannte Volumes (`/var/lib/docker`, `/srv/git`) überleben Stopp und Neustart; `make dev-stop DEV_ZUSTAND=loeschen` gibt den frischen Start, den der Start als `Zustand neu` meldet.
6. **Netz** — [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) und `SPEC-215` Nr. 9 c: eigenes Netz je Container, Ausgang per NAT, kein Port; die Prüfläufe im inneren Daemon bleiben `--network none`. **Ausgang: offen, Adresse User** (mit 1).
7. **Raster und Tabelle** — §6 und §3; Randformen ohne Ort gibt es nicht (§6).
8. **Prüfbarkeit** — `SPEC-215` Nr. 11 und 12, §3 *Prüfbarkeit*: ein Ersatz für `docker` am Anfang von `PATH`; der echte Start belegt nur ein Handlauf auf `u-boot` (akzeptiertes Negativ, §3 Zeilen 40, 51 bis 54).
9. **Größe** — §3 *Erwartete Größe* und §4: die Schätzung des ungeschnittenen Slice liegt über der Schnittgrenze; der Architect schlägt den Schnitt in drei Teile vor, der Planner schneidet vor dem Code.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- SSH-Server, Schlüsselpaar und Host-Schlüssel — `slice-dev-container-zugang`; das Gerüst läuft ohne Zugang von außen, die Schlüsselverwaltung ist eine eigene Entscheidung mit eigener Prüfbarkeit.
- Bare-Repo, Klon und Einlieferung — `slice-dev-container-einlieferung`; sie brauchen den Zugang. Das Gerüst legt nur das leere Volume `<n>-repos` an (`SPEC-215` Nr. 8 a), damit der Zustand von Anfang an vollständig benannt ist.
- Lauf von `make gates` und `make gegenproben` im Container — `slice-dev-container-lauf`; dieser Slice beweist, dass der Daemon antwortet, nicht, dass die Gates in ihm bestehen.
- Das Dokument für Entwickler — `slice-dev-container-dokument`; es beschreibt den ganzen Weg und lässt sich erst nach ihm schreiben.
- Ein Eingriff in die vorhandenen Gate-Stufen (`Dockerfile`: `test`, `build`, `race`, `lint`, `gegenproben`) — Bestand: der Container fährt sie, wie sie sind; ein Eingriff verschöbe die Prüfung ihrer Gates. Ausgenommen sind die Zeilen, die die neuen Treiber in die Stufe `gegenproben` kopieren und in `EINGABEN` von `tools/harness/gegenproben-eingabe.sh` und in die Positivliste von `.dockerignore` aufnehmen: sie ändern nicht, was die Stufe prüft, sondern was sie liest (Auftrag 8; Änderung des Architect am 2026-11-23, sonst gäbe es keinen Fall der Gegenproben zum Gerüst).

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

- [ ] Die Entscheidungen 1, 2 und 6 des Architect stehen vor dem Code in [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) (Privilegierung, Limits, Netz) und die Entscheidungen 3 bis 5 im Vertrag `SPEC-215` unter `harness/vertraege/`; die ADR hat den Status `Accepted`, bevor der erste Code-Commit entsteht, und der Vorbehalt A des Vertrags ist dann entfallen oder nachgezogen.
- [ ] Start mit Limit und Bereitschaftsprobe nach dem Vertrag: ein frischer Container wird bereit und meldet es; ein Daemon, der nicht bereit wird, endet nach der Frist rot; ein Lauf über dem Speicherlimit lässt den Host unberührt (der Kernel beendet Prozesse im Container; ob der Container endet, ist Entscheidung des Users, Frage 4 der ADR; gemessen auf `u-boot`, Zahl im Bericht).
- [ ] Zu jeder Zusage des Vertrags ein Fänger in der Tabelle *Zusage · Fänger* (§3) und, wo ein Skript oder Gate sie hält, ein Fall der Gegenproben; `make gegenproben` ist gefahren.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: je Ziel eine Zeile in der Tabelle der **Werkzeuge** von `harness/README.md` (Entscheidung des Architect, `SPEC-215` Nr. 1); `AGENTS.md` §3.1 bleibt unberührt, es sei denn, der User entscheidet es in einer eigenen ADR anders.
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
| `docs/plan/adr/` (Privilegierung, Limits und Netz; Schlüsselverwaltung) und ADR-Index | neu / update (Architect, geliefert) | Entscheidungen 1, 2 und 6; Status `Proposed`, `Accepted` vor dem Code (§4) |
| `harness/dev/Dockerfile`, `harness/dev/daemon.json` | neu | Image mit Pins, Adressbereich des inneren Daemons; eigener Kontext (Entscheidung 4) <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) --> |
| `tools/harness/dev-bau.sh`, `dev-start.sh`, `dev-probe.sh`, `dev-stop.sh`, `dev-lib.sh`; `harness/mk/dev.mk` | neu | Treiber und vier Ziele, nicht in `GATE_CHECKS` (`SPEC-215` Nr. 1) <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) --> |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh`, `Dockerfile` (Stufe `gegenproben`, eine eigene `COPY`-Zeile je Treiber), `.dockerignore`, `SPEC-102` | update | Fälle mit Ersatz für `docker` (Entscheidung 8, `SPEC-215` Nr. 11); die Treiber sind Eingabe; der Stempel erst im Integrationsschritt |
| `harness/README.md` | update | je Ziel eine Zeile in der Tabelle der Werkzeuge, „Nicht behauptet“ für den Teil, den nur `u-boot` belegt (`SPEC-215` Nr. 12) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15), die Mengen, die Prüfbarkeit und die erwartete Größe stehen unten und sind vom Architect (2026-11-23) vor dem Übergang `next` → `in-progress` geschrieben. Der Schnitt in drei Teile (§4) teilt die Zeilen so: **Teil I** Bau und Probe (Zeilen 1 bis 20), **Teil II** Start (21 bis 41), **Teil III** Stopp und die Handläufe auf `u-boot` (42 bis 54). Die Zeile nennt in Klammern den Ort im Vertrag; jede Mutation fährt der Implementer selbst (`AGENTS.md` §3.10).

**Tabelle *Zusage · Fänger*.** Jeder Fänger ist ein Fall der Gegenprobe mit einem Ersatz für `docker` (`SPEC-215` Nr. 11), wo nichts anderes steht; „Schleife“ heißt: ein Fall je Element der Menge (unten), die Zählprüfung nennt ihre Zahl. Handläufe und Negative sind benannt.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Keines der vier Ziele hängt an `GATE_CHECKS`; je Ziel eine Zeile in der Tabelle der Werkzeuge, Zelle höchstens 200 Zeichen (Nr. 1) | Zelle: `make docs-check` (`section-cell-oversized`); Kette: Handprobe `make -n gates` nennt keinen Treiber `dev-*.sh`, Zahl der Treffer 0 im Bericht; Existenz der Zeile und ihr Ort (Werkzeuge, nicht Gates) kein Fänger, Negativ Nr. 12 d | Zelle mit 201 Zeichen; `GATE_CHECKS += dev-start` | andere |
| 2 | Das Image heißt `odata-dev:arbeit`, der Kontext des Baus ist `harness/dev`, nie `.` (Nr. 2 a) | Fall `dev-bau/aufruf`: der Ersatz hält die Argumente von `docker build`; Schleife über Tag (genau `odata-dev:arbeit`) und Kontext (letztes Argument genau `harness/dev`) | Tag `odata-recorder:dev`; Kontext `.` <!-- d-check:ignore (geplanter Ordner, entsteht mit dem Code) --> | andere |
| 3 | Jede `FROM`-Zeile außer `FROM scratch` trägt `@sha256:` und genau 64 Ziffern `0-9a-f` (Nr. 2 b, c) | Schleife über M1: 10 Verstöße (Exit 2, `Pin fehlt: <Zeile>`) und 3 Gutfälle (Exit 0, `docker build` gerufen) | Länge ungeprüft (genau 64 zu mindestens 1); nur das erste `FROM` geprüft; Großbuchstaben als Ziffer | andere |
| 4 | Jedes Paket einer `apk add`-Zeile trägt `=<version>` (Nr. 2 b, c) | Schleife über M2: 8 Verstöße (Exit 2, `Pin fehlt: <Zeile>`) und 3 Gutfälle | nur das erste Paket geprüft; Fortsetzungszeile ungelesen; `name=` ohne Version gilt als Pin | andere |
| 5 | Das Dockerfile kennt nur Basis und `apk` als Quelle von Bytes; jede andere Form ist verboten (Nr. 2 b) | Schleife über M3: 9 Verbote (Exit 2, `Quelle ohne Pin: <Wort>`) und 2 Gutfälle (Wortgrenze) | je Wort der Liste aus der Prüfung genommen (8 Mutationen); Teilstring statt Wort | andere |
| 6 | Ein Pin-Verstoß, ein fehlendes Dockerfile und ein Aufruf mit Argument rufen kein `docker` (Nr. 2 c, 4 a) | die Fälle von 3 bis 5 plus Dockerfile fehlt und Argument (auch ein leeres): der Zähler des Ersatzes ist in jedem Fall der Schleife 0 | `docker info` vor der Prüfung | andere |
| 7 | Der Bau: gelungen Exit 0 und `Image odata-dev:arbeit gebaut`; scheitert `docker build` Exit 1 und `Bau gescheitert`; `docker` unerreichbar Exit 2 und `docker nicht erreichbar`; ein Dockerfile ohne `FROM` Exit 1 (Nr. 2 c, d, 4) | je Form ein Fall (4 Fälle); der Ersatz antwortet auf `info` und `build` mit dem Exit des Falls | Exit 0 trotz scheiterndem `build`; Meldung fehlt | andere |
| 8 | Die Probe prüft fünf Bedingungen in fester Reihenfolge; bestehen alle: fünf `ok`-Zeilen, `bereit`, Exit 0 (Nr. 6, 6 a) | Fall `dev-probe/bereit`: `zahl M4 "$n" 5`, Nummern und Namen der Zeilen in Reihenfolge; dieselbe Bibliothek ruft der Start (Zeile 34) | eine Bedingung aus der Liste genommen; zwei vertauscht | andere |
| 9 | Die erste verletzte Bedingung beendet den Durchgang: Exit 1, `ok` für alle davor, `nicht erfüllt` für diese, nichts für spätere; bei zwei verletzten gilt die frühere (Nr. 6 a, Zusatztabelle) | Schleife über k = 1 bis 5 und über alle zehn Paare (k, j) mit k < j; der Ersatz zählt je Bedingung ihre Aufrufe: hinter k keiner | Probe läuft alle Bedingungen und meldet die letzte; falsche Nummer | andere |
| 10 | Bedingung 1 `container`: nicht vorhanden, nicht laufend, ohne Label, Label eines anderen Namens sind verletzt (Nr. 6) | Schleife über M5: 4 Verstöße, 1 Gutfall | Label nicht geprüft; `Running` nicht geprüft | andere |
| 11 | Bedingung 2 `daemon`: Exit ungleich 0, leere Antwort und nur Leerraum sind verletzt (Nr. 6) | Schleife über M6: 3 Verstöße, 1 Gutfall | Antwort nicht gekürzt; Exit ignoriert | andere |
| 12 | Bedingung 3 `eigener-daemon`: gleiche Kennung wie der Host, leere innere und leere Kennung des Hosts sind verletzt (Nr. 6) | Schleife über M7: 3 Verstöße, 1 Gutfall; `info` des Hosts und des Containers antworten getrennt | Vergleich entfernt; leer gilt als ungleich | andere |
| 13 | Bedingung 4 `limit` (Vorbehalt A): Speicher, Swap und CPU gelten im Container wie angefordert; ungesetzt oder `max` ist verletzt (Nr. 6) | Schleife über M8: 10 Verstöße (je Datei ungleich und `max`, dazu `Memory` 0, `NanoCpus` 0, `MemorySwap` ungleich `Memory`, `cat` scheitert), 1 Gutfall; zwei Größen (10g mit 8 CPUs und 512m mit 1 CPU), die Sollwerte aus Rechnung (Bytewerte, `NanoCpus ÷ 10000`) | je Vergleich entfernt (3 Mutationen); `max` als Zahl gelesen | andere |
| 14 | Bedingung 5 `kein-tcp` (Vorbehalt A): ein lauschender Port 2375 oder 2376 ist verletzt, ein Port mit anderer Zahl nicht (Nr. 6) | Schleife über M9: 4 Verstöße, 4 Gutfälle (`:23750`, `:12375`, `:2377`, keine Zeile) | Teilstring-Suche statt Portgrenze; nur 2375 | andere |
| 15 | Die Probe ist lesend: erlaubt sind `container inspect`, `exec` mit den Befehlen der Bedingungen 2 bis 5 und `info` (Nr. 6) | in jedem Fall von 8 bis 14 und in jeder Lage von M10 prüft der Ersatz jeden Aufruf gegen die Menge der erlaubten Formen (Allaussage); Zähler von `run`, `create`, `start`, `stop`, `kill`, `restart`, `rm`, `cp`, `volume`, `network`, `build`, `pull`: je 0 | `docker restart` oder `exec <n> touch` hinzugefügt | andere |
| 16 | Dieselbe Lage gibt dieselbe Ausgabe, denselben Exit und dieselbe Aufrufliste (Nr. 6 b) | Schleife über M10 (6 Lagen): zwei Läufe hintereinander, Vergleich aller drei | ein Zähler oder eine Zeit in der Ausgabe | andere |
| 17 | `DEV_NAME` ist `[a-z][a-z0-9-]{0,30}`; jede Abweichung ist Exit 2 ohne Aufruf von `docker`; die Probe ist der erste Treiber dieser Achse, Start und Stopp kommen in ihren Teilen hinzu (Nr. 3, 4 a) | Schleife über M11 (10 Verstöße, 3 Gutfälle) × die vorhandenen Treiber; `zahl M11 "$n" 10` und die Zahl der Treiber | Länge 31 zu 32 (beide Seiten der Grenze); Großbuchstaben erlaubt | andere |
| 18 | INT, HUP und TERM an `dev-bau` (während `docker build`) und an `dev-probe` (während `inspect` und `exec`) enden mit Exit 2 und `abgebrochen`, ohne weiteren Aufruf (Nr. 10) | Schleife über 3 Phasen × 3 Signale = 9 (M19, Teil Bau und Probe); der Ersatz schläft 1 s im Aufruf, der Fall sendet das Signal und zählt die Aufrufe danach | Handler fehlt (Exit 143); nur INT behandelt | nebenläufig |
| 19 | Reihenfolge: Argument vor Variable vor `docker` (Nr. 4 a); bei zwei Verstößen gilt der frühere | Paare (Argument, ungültiger Name) und (ungültiger Name, `docker` unerreichbar) je Treiber `dev-bau` und `dev-probe`, soweit er die Variable liest; Zähler `docker` 0 | Reihenfolge vertauscht | andere |
| 20 | Befunde stehen auf stderr, Erfolgs- und Fortschrittszeilen auf stdout, jede Zeile mit dem Präfix `dev-<Ziel>: ` (Nr. 4 c) | in jedem Fall der Zeilen 3 bis 19 liest der Fall beide Ströme getrennt und prüft Strom und Präfix jeder Zeile (Allaussage über alle Meldungen); `docker`-Ausgabe unverändert am Strom des Aufrufs | eine Meldung auf dem falschen Strom; Präfix fehlt | andere |
| 21 | `DEV_MEMORY` hat die Form `[1-9][0-9]{0,5}[mg]`; leer, null und jede andere Form sind Exit 2 mit der Meldung ihrer Klasse (`leer`, `null`, `Form`), ohne Aufruf von `docker` (Nr. 3) | Schleife über M12: 20 Verstöße (die Klasse je Element im Orakel) und 3 Gutfälle (der Wert steht in `--memory` und `--memory-swap`) | Einheit `k` erlaubt; führende Null erlaubt; `0` gültig; leer als Vorgabe | andere |
| 22 | `DEV_MEMORY` bleibt unter dem Speicher des Hosts: gleich oder darüber ist Exit 2 (Nr. 3) | der Ersatz meldet `MemTotal` 16 GiB: `15g` (gut), `16g`, `17g`, `16384m` (gleich), `16383m` (gut): beide Seiten der Grenze, Gleichheit, beide Einheiten | `>=` zu `>`; Einheit `m` nicht umgerechnet | andere |
| 23 | `DEV_CPUS` hat die Form `[1-9][0-9]{0,2}` und liegt höchstens bei `NCPU` (Nr. 3) | M13: 11 Verstöße; `NCPU` 4: `4` gut, `5` Exit 2; Gutfälle `1` und `4` in `--cpus` | `>` zu `>=`; `8.0` erlaubt | andere |
| 24 | `DEV_FRIST` hat die Form `[1-9][0-9]{0,2}`, höchstens `600` (Nr. 3) | M14: 8 Verstöße (darunter `601`), 2 Gutfälle (`1`, `600`) | Grenze 601; leer als Vorgabe | andere |
| 25 | Nicht gesetzte Variablen gelten als Vorgabe (`odata-dev`, `10g`, `8`, `90`), gesetzte und leere sind Fehler (Nr. 3) | Schleife über die vier Variablen: je eine nicht gesetzt, die anderen gesetzt, und alle vier nicht gesetzt: die Vorgabe steht in `--name`, `--memory`, `--memory-swap`, `--cpus` und in `warte … Frist 90 s` | Vorgabe geändert (`9g`); nicht gesetzt als leer | andere |
| 26 | Reihenfolge der Prüfungen des Starts: Variablen, `docker` erreichbar, Werte gegen Host, cgroup, Image, Name und Fremdobjekte (Nr. 4 a, 5 a); bei zwei Verstößen gilt der frühere | Schleife über M23: 11 Paare (die sechs Paare ungültiger Variablen und die fünf Folgen der Umgebung, Zusatztabelle §6) | Reihenfolge je Paar vertauscht | andere |
| 27 | cgroup v1 und ein fehlendes Image sind Exit 2 mit Meldung, ohne Wirkung (Nr. 5 a) | Schleife (3): cgroup `1`, cgroup leer, Image fehlt; Zähler von `run`, `volume create`, `network create`, `rm`, `stop` je 0 | cgroup nicht geprüft; Prüfung des Images entfernt | andere |
| 28 | Ein Container des Namens in jedem Zustand und ein fremdes Volume oder Netz verhindern den Start ohne Wirkung (Nr. 5 a) | Schleife über M18: 5 Zustände × mit und ohne Label (10), dazu Volume `<n>-docker` ohne Label, Volume `<n>-repos` ohne Label, Netz ohne Label (3): Exit 2 und `Name vergeben (Zustand <z>)` bzw. `gehört nicht zum Arbeitscontainer`; Zähler von `run`, `rm`, `stop`, `create` je 0 | Zustand `paused` nicht erkannt; Label nicht geprüft; vorhandenen Container entfernt | andere |
| 29 | `docker run` trägt jede der vierzehn Pflichtoptionen mit dem Wert aus Name, Limit und Kennung (Nr. 5 c) | Schleife über M16 (14) × zwei Namen × zwei Limits; der Ersatz hält die Argumente als Liste; `zahl M16 "$n" 14` | je Option entfernt (14 Mutationen); Name statt Kennung | andere |
| 30 | `--memory` und `--memory-swap` tragen für jeden gültigen Wert denselben Wert (kein Swap) (Nr. 5 c, 9) | Schleife über die Gutfälle von M12 (`1m`, `512m`, `15g`) und die Vorgabe | `--memory-swap` doppelt so groß; fehlend | andere |
| 31 | Im Aufruf von `docker run` kommt keine der fünfzehn verbotenen Formen vor; der Container sieht vom Host nur seine Volumes (Nr. 5 d) | Schleife über M17 (15 Formen) × zwei Namen (`a` und `a-b`, ein Präfix-Paar): der Ersatz prüft jedes Wort der Liste gegen jede Form; `zahl M17 "$n" 15`; eine Fixture mit jeder Form macht den Prüfer selbst rot | je Form in den Aufruf eingefügt (15 Mutationen); Präfixvergleich statt Gleichheit bei der Quelle | andere |
| 32 | Zwei Container verschiedener Namen teilen weder Container noch Volume noch Netz, auch bei Namen mit gemeinsamem Präfix (Nr. 8 a) | Wirkungs-Fänger: zwei Starts (`a` und `a-b`; `a` und `b`); der Ersatz sammelt alle Namen der angelegten und eingehängten Objekte beider; die Schnittmenge ist leer; je Start 2 Volumes, 1 Netz, 1 Container | feste Volume-Namen (`odata-dev-docker`); Netz ohne den Namen | nebenläufig |
| 33 | Volumes und Netz entstehen nur, wenn sie fehlen, mit dem Label; `Zustand neu` oder `Zustand vorhanden` folgt dem Volume `<n>-docker` (Nr. 5 b, 8 c) | Schleife über M24: 8 Lagen (Volume `docker`, Volume `repos`, Netz je vorhanden oder fehlend); Zähler von `create` je Objekt, Label im Aufruf, Meldung | `create` immer; Meldung aus dem Netz abgeleitet | andere |
| 34 | Der Start wartet bis zur Bereitschaft und endet beim ersten Bestehen mit Exit 0 (Nr. 5 e) | die Probe besteht im k-ten Durchgang (k = 1, 2, 3, `DEV_FRIST` 10): Exit 0, `bereit`-Zeile, genau k Durchgänge, die Zeile `warte` einmal; dieselben Bedingungen wie Zeile 8 | Schleife läuft bis zur Frist; `bereit` ohne Probe | andere |
| 35 | Besteht die Probe bis zur Frist nicht, endet der Start mit Exit 1 und nennt die zuletzt verletzte Bedingung (Nr. 5 f) | Schleife über Bedingung 1 bis 5 als verletzte, `DEV_FRIST` 2: Exit 1, Nummer und Name in der Meldung, Dauer zwischen Frist und Frist plus 5 s; Gegenstück: bereit im letzten Durchgang vor der Frist, Exit 0 | Frist ohne Wirkung (der Fall reißt seine Frist); `>` zu `>=` | andere |
| 36 | Endet der Container vor der Bereitschaft, endet der Start sofort mit Exit 1 und dessen Exit-Code (Nr. 5 f) | Schleife über Code 137, 1, 0 und den Zeitpunkt (vor dem ersten, nach dem zweiten Durchgang): Dauer unter der Frist, Code in der Meldung | `State.Running` nicht gelesen; Code fest 137 | andere |
| 37 | Ein roter Start schreibt `logs --tail 40`, entfernt nur einen Container mit der eigenen Kennung und lässt Volumes und Netz stehen (Nr. 5 f, g) | Schleife über {Frist abgelaufen, Container endet} × {Kennung gleich, fremd}: Zähler von `logs`, `rm`, `volume rm`, `network rm` | `rm` ohne Kennungsprüfung; `volume rm` im roten Start | andere |
| 38 | Endet `docker run` mit einem Namenskonflikt, räumt der Start nichts ab (der Container gehört einem zweiten Start); endet es aus anderem Grund, räumt er einen halb angelegten Container mit eigener Kennung ab (Nr. 5 g) | Wirkungs-Fänger: der Ersatz antwortet auf `run` mit Exit 125 und `Conflict`, der Container trägt eine fremde Kennung: Exit 2, `docker run gescheitert`, Zähler von `rm` und `stop` je 0; Gegenstück: Exit 125 mit anderem Text, Container mit eigener Kennung: `rm` genau einmal | `rm --force <n>` ohne Kennungsprüfung | nebenläufig |
| 39 | INT, HUP und TERM an den Start enden mit Exit 2 und `abgebrochen`; der Container mit eigener Kennung wird entfernt, auch wenn das Signal während `docker run` eintrifft; Volumes und Netz bleiben (Nr. 10, 5 g) | Schleife über 3 Signale × 4 Phasen (vor `run`, während `run`, in der Probe-Schleife, während `logs`) = 12; „während `run`“: der Ersatz lässt `run` 1 s schlafen und erfolgreich enden, danach `rm` genau einmal; vor `run`: `rm` null | Handler entfernt nicht; Kennung erst nach `run` gesetzt (Signal während `run` ungefangen) | nebenläufig |
| 40 | Der Start auf `u-boot` wird bereit und meldet es; Zeit, `memory.max`, `cpu.max` und Spitze des Speichers stehen im Plan (Nr. 12 a) | kein automatischer Fall: Handlauf `make dev-start`, danach `make dev-probe`; Negativ Nr. 12 a (Grund: kein Daemon und kein Privileg in den Prüfläufen; Folge: der Treiber kann am echten `docker` scheitern, der Handlauf ist der Fänger) | — | andere |
| 41 | Der Start legt auf dem Daemon des Hosts nur Container, zwei Volumes und ein Netz an und zieht oder entfernt kein Image (Nr. 8 d) | Allaussage über die Aufrufe des Ersatzes in jedem Fall des Starts: jeder verändernde Aufruf ist `volume create` (2), `network create` (1), `run` (1) oder (rot) `rm` des eigenen Containers; Zähler von `pull`, `tag`, `rmi`, `image`, `build`, `cp` je 0 | `docker pull` im Start | andere |
| 42 | `DEV_ZUSTAND` ist `behalten` oder `loeschen`; jede andere Form ist Exit 2 ohne Aufruf von `docker`; `DEV_NAME` folgt für den Stopp der Achse von Zeile 17 (Nr. 3) | Schleife über M15 (5 Verstöße; Gutfälle `behalten`, `loeschen`, nicht gesetzt); `löschen` aus Bytewerten gebaut und gegengelesen (`AGENTS.md` §3.18) | `loeschen` als Teilstring; leer als Vorgabe | andere |
| 43 | Der Stopp prüft vor jeder Wirkung, dass Container, Netz und (bei `loeschen`) Volumes das Label tragen; sonst Exit 2 ohne Wirkung (Nr. 7 a) | Schleife über Objektart (Container, Netz, Volume `<n>-docker`, Volume `<n>-repos`) × {`behalten`, `loeschen`} × {Container vorhanden, fehlt} = 14 Lagen (ein fremder Container fehlt nicht), das fremde Objekt ohne Label, die anderen eigen: bei `behalten` bleibt ein fremdes Volume unbeachtet (Exit 0), sonst Exit 2; bei Exit 2 Zähler von `stop`, `rm`, `network rm`, `volume rm` je 0 | Prüfung erst nach dem ersten Schritt; Label nicht geprüft | andere |
| 44 | Die Schritte laufen in der Reihenfolge `stop --time 30`, `rm` (ohne `--force`), `network rm`, bei `loeschen` je `volume rm` (Nr. 7 b) | Aufrufliste des Ersatzes für die Zustände running, exited, created (`stop` immer), Wortlaut der Argumente | `--force`; `--time 10`; `rm` vor `stop` | andere |
| 45 | Ein Container, der nicht besteht, ist Exit 0 mit `nicht vorhanden`, der zweite Stopp hintereinander gleich; bei `loeschen` gehen Netz und Volumes trotzdem (Nr. 7 c, d) | Schleife über M22: Container, Netz, Volumes je vorhanden oder fehlend (8) × `behalten`, `loeschen` (2) = 16; ein Schritt nur für ein vorhandenes Objekt; zweimaliger Aufruf gleiche Ausgabe | `stop` ohne Existenzprüfung (Exit 1); Abbruch bei fehlendem Container vor `loeschen` | andere |
| 46 | `behalten` (Vorgabe) entfernt kein Volume (Nr. 7 b, 8 a) | zwei Fälle: Variable nicht gesetzt, `behalten`: Zähler `volume rm` null, Meldung `Zustand behalten (…)` | `volume rm` bei jedem Stopp | andere |
| 47 | `loeschen` entfernt jedes Volume mit dem Label des Containers, auch ein später hinzugekommenes, und nur diese (Nr. 7 b, 8 a) | Schleife über M21: `volume ls` meldet 0 bis 3 Volumes (das dritte `<n>-ssh`): je ein `volume rm` mit genau diesem Namen; ein Volume ohne Label fehlt in der Liste und bleibt | feste Namen statt der Liste; Label-Filter fehlt im `ls` | andere |
| 48 | Der Stopp von `a` lässt Container, Netz und Volumes von `a-b` und `b` unberührt (Nr. 7 a, 8 a) | Wirkungs-Fänger: der Ersatz hält zwei Container (`a`, `a-b`) mit je zwei Volumes und Netz; nach `dev-stop` mit `loeschen` für `a` kommt kein Name von `a-b` oder `b` in den Argumenten der verändernden Aufrufe vor | Filter nach dem Präfix `a-`; `volume rm` über alle Volumes | nebenläufig |
| 49 | Der erste scheiternde Schritt beendet den Stopp mit Exit 1 und `<Schritt> gescheitert`; kein späterer Schritt läuft (Nr. 7 b) | Schleife über Schritt k (`stop`, `rm`, `network rm`, erstes und zweites `volume rm`) = 5 (M20): der Ersatz lässt Schritt k mit Exit 1 enden; Zähler der späteren Aufrufe je 0 | Fehler des Schritts ignoriert; Meldung ohne Schrittnamen | andere |
| 50 | INT, HUP und TERM an den Stopp enden mit Exit 2 und `abgebrochen`, ohne späteren Schritt; ein zweiter Aufruf stellt den Stand her (Nr. 10, 7 c) | Schleife über 4 Phasen (`stop`, `rm`, `network rm`, `volume rm`) × 3 Signale = 12; der Ersatz schläft 1 s im Schritt; danach ein zweiter Lauf auf dem Rest: Exit 0 | Handler fehlt; nächster Schritt nach dem Signal | nebenläufig |
| 51 | Eine Überschreitung des Speichers im Container lässt den Host unberührt; der Kernel beendet Prozesse im Container (Nr. 9 b) | kein Gate: Handlauf auf `u-boot` mit einem Prozess, der über das Limit wächst, in einem Container mit `DEV_MEMORY=1g`: Exit 137 des Prozesses, Daemon des Hosts antwortet, `free` vor und nach; die Mutation ohne `--memory` wird nicht gefahren (sie belastete den Host); Bedingung 4 (Zeile 13) und Zeile 29 fangen die fehlende Option; Negativ Nr. 12 a | — | andere |
| 52 | Der innere Daemon erreicht das Netz für `deps`, der Adressbereich überlappt kein Außennetz, innere Prüfläufe bleiben `--network none` (Nr. 9 c) | kein automatischer Fall: Handlauf auf `u-boot` (ein `docker pull` im inneren Daemon; ein innerer Container mit `--network none` ohne Ausgang); Negativ Nr. 12 e | — | andere |
| 53 | Zwei Container (`a`, `b`) laufen zugleich ohne Kollision; der Stopp von `a` lässt `b` bereit (Nr. 8 a, 7 a) | Handlauf auf `u-boot`: beide Namen, `make dev-probe` je Container vor und nach dem Stopp von `a` (Zeit, Last und Messender im Plan); darüber hinaus nur die Zeilen 32 und 48; Negativ Nr. 12 g | — | nebenläufig |
| 54 | Nach Start und Stopp trägt der Daemon des Hosts nichts außer der Spur (Image, Container, Volumes und Netz des Arbeitscontainers) (Nr. 8 d) | Handlauf auf `u-boot`: `docker images`, `ps -a`, `volume ls`, `network ls` vor dem Start und nach dem Stopp mit `loeschen`; die Mengen sind gleich bis auf das Image | — | andere |

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
| M12 `DEV_MEMORY` | 20 Verstöße (leer; `0`, `0g`, `0m`, `00g` als null; `10`; `10G`; `10M`; `10gb`; `10k`; `1.5g`; `-1g`; `+1g`; mit Leerraum davor und danach; `010g`; sieben Ziffern; `max`; `unlimited`; `-1`), 3 Gutfälle (`1m`, `512m`, `15g`) | Einheit, Ziffern, Vorzeichen, Leerraum, Wort |
| M13 `DEV_CPUS` | 11 Verstöße (leer; `0`; `00`; `08`; `0.5`; `8.0`; `-1`; `1e1`; mit Leerraum; `abc`; vier Ziffern), dazu Host-Grenze (`4` gut, `5` Exit 2), Gutfälle `1` und `4` | Form, Ziffern, Grenze `NCPU` |
| M14 `DEV_FRIST` | 8 Verstöße (leer; `0`; `601`; `1s`; `1.5`; `-5`; `0600`; vier Ziffern), 2 Gutfälle (`1`, `600`) | Form, Grenze 600 |
| M15 `DEV_ZUSTAND` | 5 Verstöße (leer; `LOESCHEN`; `ja`; `löschen` mit Umlaut aus Bytewerten; `behalten` mit Leerraum), 3 Gutfälle (`behalten`, `loeschen`, nicht gesetzt) | Schreibweise, Umlaut, Leerraum |
| M16 Pflichtoptionen | 14 (`--detach`, `--name`, `--label odata-dev.name`, `--label odata-dev.start`, `--restart no`, `--network`, `--mount` für `<n>-docker`, `--mount` für `<n>-repos`, `--memory`, `--memory-swap`, `--cpus`, `--pids-limit`, `--cgroupns private`, `--privileged`) | Option × Name × Limit |
| M17 verbotene Formen | 15 (die Nummern (1) bis (15) von Nr. 5 d) | Form × Name (Präfix-Paar) |
| M18 Fremdobjekte | 13 (Zustände `running`, `exited`, `created`, `paused`, `restarting` je mit und ohne Label; Volume `docker`, Volume `repos`, Netz je ohne Label) | Zustand × Label; Objektart |
| M19 Abbruch | 33 (11 Phasen × 3 Signale: Bau 1, Probe 2, Start 4, Stopp 4) | Treiber × Phase × Signal |
| M20 Schritte des Stopps | 5 (`stop`, `rm`, `network rm`, erstes und zweites `volume rm`) | Schritt |
| M21 Volumes | 4 (0 bis 3 mit Label, das dritte später hinzugekommen) | Zahl, Herkunft des Namens |
| M22 Lagen des Stopps | 16 (Container, Netz, Volumes je vorhanden oder fehlend = 8, × `behalten`, `loeschen`) | Objekt × Variable |
| M23 Paare des Starts | 11 (sechs Paare der vier Variablen `DEV_NAME`, `DEV_MEMORY`, `DEV_CPUS`, `DEV_FRIST`; fünf Folgen der Umgebung) | Paar |
| M24 Lagen des Zustands | 8 (Volume `docker`, Volume `repos`, Netz je vorhanden oder fehlend) | Objekt |

**Prüfbarkeit** (Auftrag 8). *Im Gate:* die Fälle der Gegenproben rufen die Treiber im Wegwerfverzeichnis mit einem Ersatz für `docker` am Anfang von `PATH` (`SPEC-215` Nr. 11; Muster der Fälle 103 bis 114 von `SPEC-102`); sie belegen Reihenfolge, Argumente von `docker`, Zähler, Exit-Codes und Ströme, nicht den Daemon. Die Fälle mit Wartezeit setzen `DEV_FRIST` auf höchstens 3. Der Implementer misst die Laufzeit von `make gegenproben` je Teil nach der Messvorschrift von `SPEC-102` (Soll nach [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)) und trägt sie ein; die Fälle der Zeilen 18, 32, 38, 39, 48 und 50 sind keine Läufe mit Wahrscheinlichkeit (der Ersatz hält den Aufruf fest, das Signal kommt danach), und der Implementer fährt jede ihrer Mutationen zwanzigmal (`-count=20`) und nennt die Rate (`AGENTS.md` §3.10). *Nur ein Lauf auf `u-boot`:* die Zeilen 40, 51, 52, 53 und 54 (Handläufe mit Zahlen: Zeit bis bereit, `memory.max` und `cpu.max`, Spitze des Speichers, Last des Hosts und Messender, nach `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). *Akzeptierte Negative* (Grund und Folge im Vertrag, `SPEC-215` Nr. 12; Zeile in „Nicht behauptet“ von `harness/README.md`, Wortlaut des Architect: „der Start des Arbeitscontainers auf `u-boot`, seine Grenzen im Container, die Netze des inneren Daemons und die Spur auf dem Daemon des Hosts (`SPEC-215` Nr. 12 a, e), die Antworten des Ersatzes für `docker` (Nr. 12 b), die Verdrahtung der vier Ziele (Nr. 12 d) und SIGKILL (Nr. 12 f)“): 12 a bis 12 i. Ein `u-boot`-Lauf ist ein Handlauf des Implementers und einer des Verifiers.

**Erwartete Größe** (Architect, 2026-11-23, für die Marke des Planner nach `SPEC-128` Nr. 1; Produktcode in Zeilen von Skripten, Dockerfile und Fragment ohne Gegenproben und Tests). *Ungeschnitten:* 54 Zeilen der Tabelle, *a* = 7 (18, 32, 38, 39, 48, 50, 53), *b* = 47, *f* = 0 (der Ersatz für `docker` ist ein Skript im Fall, kein Port-Fake), *p* = 745: *S* = 85·7 + 80·47 + 200·0 + 745 = 5100, Marke 100·⌈135·5100/10000⌉ = 6900; das liegt über der Schnittfrage (*S* ≥ 2223) und die Marke weit über der Grenze 3000. *p* je Teil: Teil I rund 385 (`dev-bau.sh` 100, `dev-lib.sh` 90, `dev-probe.sh` 130, `harness/dev/Dockerfile` 30, `daemon.json` 8, `dev.mk` 15, `harness/README.md` 12), Teil II rund 240 (`dev-start.sh` 230, `dev.mk` 10), Teil III rund 120 (`dev-stop.sh` 110, `dev.mk` 6, `harness/README.md` 4; `dev-lib.sh` wächst nicht). Die Zeile `Dockerfile`, `.dockerignore` und `gegenproben-eingabe.sh` je Teil zählt einige Zeilen und steckt in *p*. Die Formel ist an Go-Slices geeicht; an Skripten belegt sie hier nur `slice-guard-schreibwege` (Messung 1707 von Marke 2600, Geltungsgrenze `SPEC-128` Nr. 2); die Fälle der Gegenproben sind Go-Code mit Fixtures und Shell-Ersatz. <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) -->

Betriebsregeln für alle Rollen dieses Slice: Repo-Dateien nur mit `Edit` und `Write` (kein `sed -i`, `awk`, Python, Heredoc in die Datei; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); jeder `docker run` einer Rolle mit `--memory` und `--memory-swap`, höchstens zwei Container zugleich (Erfahrung vom 2026-10-24: Mutationskopien belegten ein Vielfaches des Bedarfs); Werte mit Escape-Folge oder Rohbytes aus Bytewerten und gegengelesen (`AGENTS.md` §3.18).


## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`; für jeden Teil des Schnitts unten, der erste braucht alle Punkte, die folgenden den Vorgänger in `done/`): `welle-fundament` liegt in `done/`; [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) hat den Status `Accepted` (steht beim User aus; ist sie `Accepted`, zieht der Planner den Vorbehalt A in `SPEC-215` nach oder lässt den Architect es tun, und der Vertrag trägt die Entscheidung des Users, nicht den Vorschlag); der Vertrag `SPEC-215`, die Tabelle *Zusage · Fänger* und das Raster aus §6 stehen vor dem ersten Code-Commit (Architect, 2026-11-23, geliefert); der Planner hat den Schnitt vollzogen und die Marke je Teil gesetzt; kein anderer Slice der Gate-Kette liegt in `in-progress/` (Übergangsregeln Nr. 7: dieser Slice ändert `harness/mk/` (neues Fragment), `Dockerfile`, `.dockerignore`, `harness/README.md`, `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` und den Stempel; seine drei Teile laufen nacheinander, nie zugleich mit einem anderen Kettenslice).

**Reihenfolge und Kollisionen** (Architect, 2026-11-23; die Reihenfolge der Kette legt der Planner beim Start fest). Zur Zeit liegt kein Kettenslice in `in-progress/` (`slice-admin-tls` und `slice-telemetry-einbau-record` nennen die Kette nur in ihren Start-Bedingungen, ihre Pfade liegen unter `src/` und im eigenen Plan; gelesen am 2026-11-23). In `open/` führen Pfade der Kette (Treffer der Pfadsuche nach `harness/mk`, `Dockerfile`, `tools/gegenproben`, `Gate-Kette`, vom Planner gegen Nr. 7 zu prüfen): `slice-worktree-parallel` (Image-Tags und Sperre der Gates; die Tags `odata-recorder:*` dort, dieser Slice nutzt `odata-dev:arbeit` und kollidiert im Tag nicht, in `Dockerfile` und `harness/mk/` teilen beide Dateien), `slice-gegenproben-eichung` (misst, baut nichts um, `SPEC-102` Messvorschrift), `slice-test-security`, `slice-coverage-gate`, `slice-abdeckung-deklaration`, `slice-lasttest`, `slice-handbuch-mock` und `slice-handbuch-telemetrie` (Stufe `handbuch`, `tools/gegenproben/`, Stempel), `slice-container-image` (wartet auf diese Welle, Kante in `welle-dev-container` §5), dazu der Slice (a) von [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md) (Stempel je Gate, `Makefile`, `harness/mk/`, `harness/README.md`; noch kein Slice). Dateien mit Konflikt auf derselben Zeile: die `COPY`-Zeilen der Stufe `gegenproben` im `Dockerfile`, die Liste `EINGABEN` und die Positivliste von `.dockerignore`; der Implementer legt je Treiber eine eigene Zeile an und ändert keine fremde. `harness/mk/dev.mk` ist eine neue Datei; das `Makefile` bindet `harness/mk/*.mk` ein und bleibt unverändert. Nach `done/` wartet nur `slice-dev-container-zugang` auf den letzten Teil (Reihenfolge in `welle-dev-container` §5). <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1; **Vorschlag des Architect, die Marke setzt der Planner vor dem ersten Code-Commit**). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` zählt nicht, `docs/` und `spec/` nennt kein Pfad. Rechenweg je Teil (Zeilen und Klassen aus §3, *Erwartete Größe*; *f* = 0):

| Teil | Zeilen der Tabelle | *a* | *b* | *p* | *S* = 85·*a* + 80·*b* + *p* | Marke 100·⌈135·*S*/10000⌉ | Schnittfrage (*S* ≥ 2223) |
|---|---|---|---|---|---|---|---|
| ungeschnitten | 1 bis 54 | 7 | 47 | 745 | 5100 | 6900 | **ja** |
| Teil I: Bau und Probe | 1 bis 20 | 1 | 19 | 385 | 1990 | 2700 | nein (Abstand 233) |
| Teil II: Start | 21 bis 41 | 3 | 18 | 240 | 1935 | 2700 | nein (Abstand 288) |
| Teil III: Stopp und Handläufe | 42 bis 54 | 3 | 10 | 120 | 1175 | 1600 | nein |

*Rechnung:* Teil I 85 + 1520 + 385 = 1990, 135·1990 = 268650, aufgerundet auf 27 Hunderter; Teil II 255 + 1440 + 240 = 1935, 135·1935 = 261225, 27 Hunderter; Teil III 255 + 800 + 120 = 1175, 135·1175 = 158625, 16 Hunderter; ungeschnitten 595 + 3760 + 745 = 5100, 135·5100 = 688500, 69 Hunderter.

**Schnitt vor dem Code** (Vorschlag, `AGENTS.md` §3.14: über der Schnittgrenze wird vor dem Code geschnitten; der Planner schneidet, legt die Pläne der Teile an, trägt §3.13 nach und führt Roadmap und `welle-dev-container` nach). Die Teile teilen den Mechanismus nur in einer Richtung (der spätere nutzt, was der frühere liefert), keiner schreibt am Code des anderen: **Teil I** liefert Image, Pins, Bau, die Bibliothek `dev-lib.sh`, die Probe und den Ersatz für `docker` samt Rahmen der Fälle (Zeilen 1 bis 20); **Teil II** den Start (21 bis 41) und ruft die Probe; **Teil III** den Stopp und die Handläufe auf `u-boot` mit zwei Containern (42 bis 54). Die Handläufe des Starts (Zeile 40) gehören zu Teil II, weil dort zum ersten Mal ein Container läuft; Aufräumen übernimmt dort von Hand `docker rm --force`. Ein Schnitt in zwei Teile trägt nicht: Teil II und III zusammen liegen bei *S* = 3110, Teil I und II bei 3925. Die Metrik-Pfade je Teil (Block-Entwurf, der Planner setzt `groessenmarke` als Info-Zeichenkette, `basis`, `schaetzung` und `marke` beim Start; der Architect-Teil des Blocks ist `pfad`, `produktcode`, `fakes`):

```text
Teil I:   pfad: harness/dev
          pfad: tools/harness/dev-bau.sh
          pfad: tools/harness/dev-lib.sh
          pfad: tools/harness/dev-probe.sh
          pfad: harness/mk/dev.mk
          pfad: tools/gegenproben
          pfad: tools/harness/gegenproben-eingabe.sh
          pfad: Dockerfile
          pfad: .dockerignore
          pfad: harness/README.md
          produktcode: 385    fakes: 0    (schaetzung 1990, marke 2700)
Teil II:  dieselben Pfade ohne harness/dev, mit tools/harness/dev-start.sh statt dev-bau.sh, dev-lib.sh und dev-probe.sh
          produktcode: 240    fakes: 0    (schaetzung 1935, marke 2700)
Teil III: dieselben Pfade wie Teil II, mit tools/harness/dev-stop.sh statt dev-start.sh
          produktcode: 120    fakes: 0    (schaetzung 1175, marke 1600)
```

Der Block steht hier als `text`, damit das Gate ihn nicht als Block dieses Plans liest (`SPEC-128` Nr. 3 h: genau einer, Info-Zeichenkette `groessenmarke`); der Block im Plan eines Teils trägt die Info-Zeichenkette. Der Plan eines Teils nennt jede Datei einzeln, wie oben.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke; offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten) und die Grenze 3000 vor dem Code, die der Schnitt oben einhält.
- `in-progress` → `open` (blockiert): [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) wird `Rejected` oder `Superseded`, oder `u-boot` erlaubt den gewählten Weg nicht (Befund an den Planner, die Welle wird umgeplant, Drift-Log).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, je Zusage des Vertrags ist eine Mutation rot gesehen (`AGENTS.md` §3.10), und ein Start auf `u-boot` ist mit Zahlen belegt (Zeit bis zur Bereitschaft, Limit, Speicher beim Lauf); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Prüfbarkeit: Der privilegierte Start läuft nicht in `make gates` (Docker-only, netzlos, ohne schreibbaren Mount); was nur `u-boot` belegt, ist ohne Sensor — **Ausgang:** weiter offen; Fänger und akzeptiertes Negativ mit Grund und Folge stehen in §3 (Zeilen 40, 51 bis 54, *Prüfbarkeit*) und in `SPEC-215` Nr. 12; die Zeile für „Nicht behauptet“ steht in §3, der Implementer trägt sie in `harness/README.md` ein.
- Sicherheit: Ein privilegierter Container mit eigenem Daemon vergrößert die Angriffsfläche des Hosts — **Ausgang:** weiter offen; Adresse: User. [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) (`Proposed`) benennt Wirkung, sieben Alternativen und die Begrenzung (kein Socket, kein Bind-Mount, keine Host-Namensräume, kein Port); ohne `Accepted` kein Code.
- Konflikt mit `AGENTS.md` §3.1: der Wortlaut „Host braucht nur Docker und GNU make“ gilt für Prüfläufe — **Ausgang:** weiter offen; Adresse: User. Der Vorschlag der ADR lautet: unberührt (der Arbeitscontainer ist kein Prüflauf, die Prüfläufe in ihm bleiben netzlos und ohne Mount); eine Ergänzung der harten Regel wäre eine eigene ADR.
- Gemeinsamer Daemon-Zustand: Image-Tags `odata-recorder:*` kollidieren auf dem Host, wenn der Container den Daemon des Hosts nutzt statt eines eigenen — **Ausgang:** weiter offen: der Vertrag sagt zu, dass der innere Daemon allein baut und eine andere Kennung trägt als der des Hosts (`SPEC-215` Nr. 6, Bedingung 3, Zeile 12) und dass der Start auf dem Host nur die Spur anlegt (Nr. 8 d, Zeilen 41 und 54); der Nachweis im Lauf von `make gates` (kein Image des Laufs auf dem Host) liegt beim Slice des Laufs.
- Nebenläufigkeit zweier Container auf einem Host (Port, Name, Volume): Randform des Rasters — **Ausgang:** weiter offen; Zeilen 32, 38, 48 und 53 der Tabelle (Name und Volume je Container, Konflikt beim `docker run`, Stopp berührt nur das Eigene, Handlauf mit zwei Containern); einen Port gibt es im Gerüst nicht.
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* unten; jede Zelle trägt einen Ort im Vertrag oder „gilt nicht“ mit Grund (Architect, 2026-11-23); der Implementer hält an einer Randform an, die dort nicht steht — **Ausgang:** weiter offen bis zur Closure.
- Die Überschreitung des Speichers beendet im Container Prozesse, nicht notwendig den Container (`SPEC-215` Nr. 9 b); §2 und der Closure-Trigger der Welle sagen „beendet den Container“ — **Ausgang:** offen; Adresse: User (Frage 4 der ADR zur Privilegierung), danach Planner (zieht `welle-dev-container` §3 nach, §3.13: der Plan nimmt hier nichts an, was die Welle nicht trägt).
- [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) hat den Status `Accepted` noch nicht; ohne ihn startet kein Teil (§4) — **Ausgang:** offen; Adresse: User.
- Die Größe liegt über der Schnittgrenze (§3 *Erwartete Größe*, §4): der Slice ist ungeschnitten nicht startbar — **Ausgang:** offen; Adresse: Planner (Schnitt in drei Teile vor dem Code, Pläne der Teile, Kanten, Sendungen nach §3.13, Roadmap und Welle-Plan). Die Zeilen der Tabelle sind den Teilen zugeordnet (§3).
- Der Architect hat den Ausschluss der Gate-Stufen in §1 um die Zeilen für die Stufe `gegenproben` (`COPY`, `EINGABEN`, `.dockerignore`) ergänzt, damit ein Fall der Gegenproben zum Gerüst entsteht — **Ausgang:** offen; Adresse: Planner (bestätigt oder verwirft; verwirft er, entfallen die Zeilen 3 bis 20 und 21 bis 50 als Fälle der Gegenprobe, und der Architect liefert eine andere Prüfbarkeit).
- Folgen für die anderen Pläne der Welle (§3.13): `slice-dev-container-zugang` nennt in §1 und DoD „eine ADR zur Schlüsselverwaltung“ und die erweiterbare Probe (Auftrag 3 dort) — **Ausgang:** offen; Adresse: Planner (trägt [ADR-0032](../../adr/0032-schluesselverwaltung-des-arbeitscontainers.md) und `SPEC-215` Nr. 6 c in dessen Kopf und §3 ein); der Plan des Zugangs ist nicht Teil dieses Auftrags.
- Die Werte der Pins (Digest der Basis, Versionen der Pakete) hat der Architect nicht abgerufen (netzlos); der Implementer trägt sie ein — **Ausgang:** offen; sind sie nicht erfüllbar (Paket nicht mehr geführt), ist es ein roter Bau des Implementers, kein Anlass, den Pin zu lockern.
- Host-Annahmen (cgroup v2, Speichertreiber `overlayfs`, Docker 29.10.0, 33 362 579 456 Byte Speicher) sind am Host des Architect-Laufs gemessen; ob er `u-boot` ist, belegt der Handlauf von Zeile 40 — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16). Die Zelle nennt den Ort in `SPEC-215` oder „gilt nicht“ mit Grund; „Variable“ meint die Variablen von Nr. 3, `DEV_NAME` hat jede Operation, `DEV_MEMORY`, `DEV_CPUS` und `DEV_FRIST` der Start, `DEV_ZUSTAND` der Stopp.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Bau (`dev-bau`) | Dockerfile fehlt: Nr. 2 c (Exit 2, `harness/dev/Dockerfile fehlt`); Variablen: gilt nicht, der Bau liest keine | Dockerfile ohne `FROM`: kein Pin-Verstoß, der Bau scheitert, Exit 1 (Nr. 2 c, d) | Pin fehlt, Quelle ohne Pin, Argument des Treibers: Nr. 2 b, c und 4 a (Exit 2, kein `docker`); Daemon unerreichbar: Nr. 4 a | INT, HUP, TERM: Nr. 10 (Exit 2, nichts aufzuräumen) | zweiter Bau: Nr. 2 d (setzt den Tag neu, ein laufender Container behält sein Image) |
| Start (`dev-start`) | Variable nicht gesetzt: Vorgabe (Nr. 3); Image fehlt: Nr. 5 a (Exit 2); Zustand fehlt: legt an und meldet `neu` (Nr. 5 b, 8 c) | Variable leer, null: Nr. 3 (Exit 2, Meldung `leer` und `null`, kein `docker`) | nicht parsbarer Wert und Form, Wert über dem Host: Nr. 3 (`Form`, `über dem Speicher`, `über den Prozessoren`); cgroup v1: Nr. 5 a; fremdes Objekt des Namens: Nr. 5 a | Signal während des Starts, Daemon startet nicht (Frist), Container endet: Nr. 10 und 5 f, g (nur der eigene Container wird entfernt, Exit 2, 1 und 1) | Container des Namens vorhanden (jeder Zustand): Nr. 5 a (`Name vergeben`, Exit 2); zweiter Start nach Stopp: Nr. 8 c (`Zustand vorhanden`); zwei Starts zugleich: Nr. 5 g (der Konflikt beim `docker run`, nichts Fremdes entfernt) |
| Probe (`dev-probe`) | Container nicht vorhanden: Nr. 6 a (Bedingung 1 verletzt, Exit 1); Variable nicht gesetzt: Vorgabe | Antwort ohne Inhalt, Exit ohne Ausgabe: Nr. 6 (Bedingungen 2, 3, 5); `DEV_NAME` leer: Nr. 3 | Daemon einer fremden Version: gilt nicht, die Version legt der Pin des Images fest (Nr. 2 b) und die Bedingung 2 prüft die Antwort, nicht die Nummer; Container ohne Label, Daemon des Hosts, Limit gelockert, TCP-Port: Nr. 6 (Bedingungen 1, 3, 4, 5) | INT, HUP, TERM: Nr. 10 (Exit 2) | Probe auf bereitem Container: Nr. 6 b (gleiche Ausgabe, gleicher Exit) |
| Stopp (`dev-stop`) | Container nicht vorhanden: Nr. 7 c, d (Exit 0, `nicht vorhanden`, bei `loeschen` gehen Netz und Volumes trotzdem) | `DEV_ZUSTAND` leer: Nr. 3 (Exit 2, kein `docker`) | `DEV_ZUSTAND` unbekannt: Nr. 3; Objekt ohne das Label des Containers: Nr. 7 a (Exit 2, nichts verändert) | INT, HUP, TERM, ein Schritt scheitert: Nr. 10 und 7 b (Exit 2; Exit 1 und kein späterer Schritt) | Stopp eines gestoppten oder nicht vorhandenen Containers: Nr. 7 c (Exit 0, gleiche Ausgabe) |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Ursachen (je Paar die Entscheidung und der Fänger; die Reihenfolge der Prüfungen steht in `SPEC-215` Nr. 4 a):

| Paar | Entscheidung | Fänger |
|---|---|---|
| zwei ungültige Variablen am Start (sechs Paare aus `DEV_NAME`, `DEV_MEMORY`, `DEV_CPUS`, `DEV_FRIST`) | die frühere der Reihenfolge `DEV_NAME`, `DEV_MEMORY`, `DEV_CPUS`, `DEV_FRIST` wird gemeldet (Nr. 3) | Zeile 26 (sechs Paare) |
| Argument und ungültige Variable | der Aufruf-Fehler (Nr. 4 a) | Zeile 19 |
| ungültige Variable und `docker` unerreichbar | die Variable, kein Aufruf von `docker` (Nr. 4 a) | Zeilen 19 und 26 |
| `docker` unerreichbar und cgroup v1; cgroup v1 und Image fehlt; Image fehlt und Name vergeben; Name vergeben und fremdes Volume | die frühere Prüfung von Nr. 4 a und 5 a | Zeile 26 (fünf Folgen) |
| Limit gesetzt (Bedingung 4 verletzt) und Daemon nicht bereit (Bedingung 2 verletzt) | die frühere Bedingung (2) wird gemeldet (Nr. 6 a) | Zeile 9 (alle zehn Paare) |
| Start abgebrochen und Name schon vergeben | der Start endet vor jeder Wirkung (`Name vergeben` oder `abgebrochen`, beide Exit 2); entfernt wird nichts, weil noch kein Container mit der eigenen Kennung besteht (Nr. 5 a, g, 10) | Zeilen 28 (kein `rm`) und 39 (Phase vor `run`, `rm` null) |
| Frist abgelaufen und Container einer fremden Kennung | kein `rm` (Nr. 5 g) | Zeile 37 |
| fremdes Objekt und Container nicht vorhanden (Stopp) | das fremde Objekt: Exit 2 (Nr. 7 a) | Zeile 43 (Lagen mit Container vorhanden und fehlend) |
| Signal und gescheiterter Schritt (Stopp) | das Signal gilt: Exit 2 (Nr. 10); ein Aufruf stellt den Stand her | Zeile 50 (zweiter Lauf) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30; keiner steht bei 3× und offen). Für diesen Slice zählen: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, `offen`) — die Messungen im Container laufen unter einem festen Limit; Raten und Zeiten nennen Limit, Last und Messenden, und erreicht ein Fund mit dieser Welle 3×, braucht der Eintrag einen Ausgang (Welle-Closure); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) — die Betriebsregeln stehen in §3, ein Auftreten wird als Beleg eingetragen. Die Erfahrung vom 2026-10-24 (Speichermangel des Hosts durch Mutationskopien, Container ohne Limit) hat **keinen Eintrag im Register**; sie liegt als Notiz des Users vor und ist Anlass dieser Welle, nicht gezählt — die Closure entscheidet, ob ein Eintrag mit Beleg `evidence/slice-dev-container-geruest.md` entsteht.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
