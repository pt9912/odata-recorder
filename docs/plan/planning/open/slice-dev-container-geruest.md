# Slice slice-dev-container-geruest: Gerüst des Arbeitscontainers — Stopp und Handläufe auf `u-boot` (Teil III des Schnitts; trägt die Kennung des ungeteilten Slice)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-dev-container](../welle-dev-container.md).

**Bezug:** `LH-QA-06` (gleiches Testverhalten bei gleicher Konfiguration, Container), `LH-RB-03` (Entwicklung unter Linux und in Docker-Umgebungen), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Folgepflicht: der Arbeitscontainer fährt `make gegenproben`), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Vertrag unter `harness/vertraege/`), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Stufen des `Dockerfile`). Entscheidungen des Architect (2026-11-23): [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) (Privilegierung, Limits, Netz; `Accepted`, Entscheidung des User vom 2026-11-24) und [ADR-0032](../../adr/0032-schluesselverwaltung-des-arbeitscontainers.md) (Schlüsselverwaltung, `Accepted`, Entscheidung des User vom 2026-11-24; gehört dem Zugang, nicht diesem Slice).

**Berührte Spec-Stellen** (Teil III des Schnitts; die Stellen der Teile I und II stehen in [slice-dev-container-bau](../next/slice-dev-container-bau.md) und [slice-dev-container-start](slice-dev-container-start.md)): `SPEC-215` (Vertrag, [`SPEC-215-arbeitscontainer.md`](../../../../harness/vertraege/SPEC-215-arbeitscontainer.md): Nr. 3, 7, 8, 9, 10, 12) samt Zeile im Index von `harness/vertraege/README.md` · `SPEC-102` (die Fälle zum Stopp kommen als Nachzug des Implementers hinzu, Auftrag 8) · `AGENTS.md` §3.1 (Docker-only): bleibt unberührt, [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) Entscheidung, Einordnung nach §3.1 (Vorschlag).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Schnitt vor dem Code in drei Teile durch den Planner am 2026-11-24, Vorschlag des Architect; der Plan des ungeteilten Slice trug Ziel, Tabelle und Raster in einer Datei). **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Auf dem Host `u-boot` startet ein Container aus einem gepinnten Image mit eigenem Docker-Daemon unter festem Speicher- und CPU-Limit und meldet erst nach einer Bereitschaftsprobe, dass der Daemon in ihm Befehle annimmt; ein Start, der nicht bereit wird, endet nach einer Frist rot. **Dieser Teil** liefert den Stopp (`dev-stop`: Schritte in fester Reihenfolge, Zustand `behalten` oder `loeschen`, nur Objekte mit dem Label des Containers, wiederholbar, `SPEC-215` Nr. 3, 7, 8, 10) und die Handläufe auf `u-boot`, die das Gerüst als Ganzes belegen: Speicherüberschreitung lässt den Host unberührt, der innere Daemon erreicht das Netz für `deps`, zwei Container laufen ohne Kollision, der Daemon des Hosts trägt nach Start und Stopp nur die Spur (Zeilen 51 bis 54). **Schnitt vor dem Code** (Vorschlag des Architect, Planner, 2026-11-24): die Schätzung des ungeschnittenen Plans (*S* = 5100, Marke 6900, 54 Zeilen) liegt weit über der Schnittgrenze von `SPEC-128` Nr. 1; drei Teile, die nacheinander laufen und je einzeln lieferbar sind: **I** [slice-dev-container-bau](../next/slice-dev-container-bau.md) (Bau und Probe, Zeilen 1 bis 20), **II** [slice-dev-container-start](slice-dev-container-start.md) (Start, Zeilen 21 bis 41) und **III** (diese Datei, Zeilen 42 bis 54). Dieser Teil trägt die Kennung des ungeteilten Slice (Abweichung vom Namensvorschlag, Planner, 2026-11-24): jeder Verweis „`slice-dev-container-geruest` in `done/`“ in den Plänen davor und danach (`welle-dev-container`, `slice-dev-container-zugang` und seine Folge) bedeutet damit „alle drei Teile in `done/`“, weil dieser Teil als letzter startet und die zwei anderen voraussetzt.

**Auftrag an den Architect** (Entscheidungen vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Planner entscheidet sie nicht; Wortlaut des ungeteilten Plans):

1. **Privilegierung** (ADR-Kandidat, sicherheitsrelevant): Ein Docker-Daemon im Container braucht erweiterte Rechte. Der Architect vergleicht die Wege (privilegierter Container, Daemon ohne Root im Container, Laufzeit mit eigenem Kernel-Namensraum, Verzicht auf den Container zugunsten einer Maschine) nach Wirkung auf den Host, Aufwand auf `u-boot` und dem, was `AGENTS.md` §3.1 verlangt (kein lokaler Toolchain-Install; der Host braucht nur Docker und GNU make) — und nennt, ob §3.1 ergänzt werden muss (das wäre eine Änderung der harten Regel mit ADR, nicht stillschweigend). Zu entscheiden ist auch, was der Container auf dem Host sehen darf (Sockets, Mounts, Netz) und was nicht.
2. **Limits**: Wert und Ort von Speicher-, Swap- und CPU-Grenze (Befehlszeile des Außencontainers oder Konfiguration des Daemons), und dass die Grenze auch BuildKit im inneren Daemon trifft, der sich vom Host aus nicht begrenzen lässt; das Verhalten bei Überschreitung (Prozess im Container endet, Host unberührt; ob der Container endet, wählt der Kernel, `SPEC-215` Nr. 9 b) ist eine Zusage mit Fänger.
3. **Vertrag** unter `harness/vertraege/`: Start, Bereitschaftsprobe (was „bereit“ heißt, Frist, Exit-Codes, Wiederholung, Ausgabe), Stopp, wiederholter Start, Abbruch durch Signal. Die Probe ist **erweiterbar**: spätere Slices der Welle fügen Bedingungen hinzu (Zugang), ohne den Vertrag zu ersetzen.
4. **Image und Pins**: Basis per Digest, Versionen der Pakete, Ort der Stufe oder des Skripts (`Dockerfile`-Stufe, `harness/mk/`-Fragment, `tools/harness/`); Bindung in `harness/README.md` (Sensors-Tabelle oder Werkzeuge-Tabelle; ein Ziel, das privilegiert startet, gehört nicht in `GATE_CHECKS` — der Architect entscheidet und begründet es wie [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)).
5. **Zustand**: Was ein Neustart behält (Daemon-Verzeichnis mit Image-Cache für `deps`, Bare-Repos) und was nicht; ein frischer Start ohne Zustand muss möglich sein (Closure-Kriterium der Welle).
6. **Netz**: Die Prüfläufe im inneren Daemon sind netzlos (`--network none`), nur `deps` holt aus dem Netz; welches Netz der Außencontainer bekommt (Spiegel, Proxy, nur Ausgang) ist eine Entscheidung mit Grund.
7. **Raster und Tabelle**: das Raster *Operation × Argumentart* (Operationen Start, Probe, Stopp; Argumente Limit fehlend, leer, null, unbekannt, größer als der Host; Name des Containers schon vergeben; Daemon startet nicht; Abbruch während des Starts; Wiederholung) in §6, jede Zelle mit Ort im Vertrag (`AGENTS.md` §3.16), und die Tabelle *Zusage · Fänger* in §3 (`AGENTS.md` §3.15) — beides vor dem Übergang `next` → `in-progress`.
8. **Prüfbarkeit**: wie ein Fänger ohne privilegierten Start im Gate lebt (Ersatz für `docker` in einer Gegenprobe nach dem Muster der Fälle 103 bis 114 von `SPEC-102`, Fälle in `tools/gegenproben/`) und welcher Teil nur ein Lauf auf `u-boot` belegt; Letzteres steht als akzeptiertes Negativ mit Grund und Folge in §3 und in „Nicht behauptet“ von `harness/README.md`.
9. **Größe**: Schätzung und Marke nach `AGENTS.md` §3.14 (Zeilen der Tabelle, Klassen, Port-Fakes entfallen hier; der Fake ist ein Ersatz-Skript), die der Planner nach dem Vertrag mit dem Werkzeug setzt; wird die Grenze 3000 überschritten, schneidet der Planner vor dem Code.

**Ausgang des Architect** (2026-11-23; Vertrag `SPEC-215`; `AGENTS.md` §3.12, §3.16). Die zwei ADRs sind `Accepted` (Entscheidung des User vom 2026-11-24: sie berühren die Sicherheit des Hosts und gehören dem User); der Vorbehalt A des Vertrags an [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) ist entfallen, die Stellen tragen die Entscheidung.

1. **Privilegierung** — [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md), Entscheidung (User, 2026-11-24): `--privileged` mit Daemon im Container, aber ohne Docker-Socket, Bind-Mount, Host-Namensräume und veröffentlichten Port; der innere Daemon hört nur auf dem Unix-Socket. Verglichen sind sieben Wege (gezielte Fähigkeiten, Daemon ohne Root, Laufzeit mit eigenem Kernel, Maschine, Socket des Hosts, nichts tun). `AGENTS.md` §3.1 bleibt unberührt: der Wortlaut gilt für Prüfläufe, und die bleiben in den inneren Containern netzlos und ohne Mount; eine Ergänzung wäre eine eigene ADR. Im Vertrag: Nr. 5 c, d, Nr. 6 (Bedingungen 3 bis 5), Nr. 9. **Ausgang: entschieden** (User, 2026-11-24).
2. **Limits** — [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) und `SPEC-215` Nr. 9: Ort ist die Befehlszeile des Außencontainers (die cgroup trifft BuildKit und jeden inneren Container, Bedingung 4 liest die Grenze im Container nach); Vorgaben 10 GiB, kein Swap, 8 CPUs, 8192 Prozesse. **Korrektur an der Formulierung des Plans:** bei Überschreitung beendet der Kernel Prozesse im Container und lässt den Host unberührt; ob der Container endet, hängt davon ab, ob der Kernel den ersten Prozess wählt (`SPEC-215` Nr. 9 b). „Beendet den Container“ gilt nicht für jede Überschreitung: der User hat den Kernel wählen lassen, ein Ende der ganzen Gruppe setzt der Start nicht (Frage 4 der ADR). **Ausgang: entschieden** (User, 2026-11-24; Planner, 2026-11-24: §2 und die Welle tragen den Wortlaut von `SPEC-215` Nr. 9 b).
3. **Vertrag** — `SPEC-215` Nr. 4 bis 8 und 10: Start, Probe (fünf Bedingungen, erweiterbar durch Anhängen, Nr. 6 c), Stopp, wiederholter Start, Abbruch durch Signal.
4. **Image und Pins** — `SPEC-215` Nr. 2: ein eigenes Dockerfile `harness/dev/Dockerfile` mit eigenem Kontext (kein Eingriff in die Stufen der Gates), Basis per Digest, Pakete mit Version, andere Quellen verboten, geprüft von `dev-bau.sh` vor dem Bau. Bindung: Zeilen in der Tabelle der **Werkzeuge** von `harness/README.md`, nicht in `GATE_CHECKS` (`SPEC-215` Nr. 1, Alternativen und Grund dort). <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->
5. **Zustand** — `SPEC-215` Nr. 8: zwei benannte Volumes (`/var/lib/docker`, `/srv/git`) überleben Stopp und Neustart; `make dev-stop DEV_ZUSTAND=loeschen` gibt den frischen Start, den der Start als `Zustand neu` meldet.
6. **Netz** — [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) und `SPEC-215` Nr. 9 c: eigenes Netz je Container, Ausgang per NAT, kein Port; die Prüfläufe im inneren Daemon bleiben `--network none`. **Ausgang: entschieden** (User, 2026-11-24, mit 1).
7. **Raster und Tabelle** — §6 und §3; Randformen ohne Ort gibt es nicht (§6).
8. **Prüfbarkeit** — `SPEC-215` Nr. 11 und 12, §3 *Prüfbarkeit*: ein Ersatz für `docker` am Anfang von `PATH`; der echte Start belegt nur ein Handlauf auf `u-boot` (akzeptiertes Negativ, §3 Zeilen 40, 51 bis 54).
9. **Größe** — §3 *Erwartete Größe* und §4: die Schätzung des ungeschnittenen Slice liegt über der Schnittgrenze; der Architect schlägt den Schnitt in drei Teile vor, der Planner hat vor dem Code geschnitten (§4).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Image, Pins, Bau, Probe und Bibliothek — [slice-dev-container-bau](../next/slice-dev-container-bau.md) (Teil I); der Start mit Limits, Volumes und Netz — [slice-dev-container-start](slice-dev-container-start.md) (Teil II); beide liegen bei diesem Start in `done/`, dieser Teil ruft ihre Bibliothek.
- SSH-Server, Schlüsselpaar und Host-Schlüssel — `slice-dev-container-zugang`; das Gerüst läuft ohne Zugang von außen, die Schlüsselverwaltung ist eine eigene Entscheidung mit eigener Prüfbarkeit.
- Bare-Repo, Klon und Einlieferung — `slice-dev-container-einlieferung`; sie brauchen den Zugang. Das Gerüst legt nur das leere Volume `<n>-repos` an (`SPEC-215` Nr. 8 a), damit der Zustand von Anfang an vollständig benannt ist.
- Lauf von `make gates` und `make gegenproben` im Container — `slice-dev-container-lauf`; dieser Slice beweist, dass der Daemon antwortet, nicht, dass die Gates in ihm bestehen.
- Das Dokument für Entwickler — `slice-dev-container-dokument`; es beschreibt den ganzen Weg und lässt sich erst nach ihm schreiben.
- Ein Eingriff in die vorhandenen Gate-Stufen (`Dockerfile`: `test`, `build`, `race`, `lint`, `gegenproben`) — Bestand: der Container fährt sie, wie sie sind; ein Eingriff verschöbe die Prüfung ihrer Gates. Ausgenommen sind die Zeilen, die die neuen Treiber in die Stufe `gegenproben` kopieren und in `EINGABEN` von `tools/harness/gegenproben-eingabe.sh` und in die Positivliste von `.dockerignore` aufnehmen: sie ändern nicht, was die Stufe prüft, sondern was sie liest (Auftrag 8; Änderung des Architect am 2026-11-23, vom Planner am 2026-11-24 bestätigt, [slice-dev-container-bau](../next/slice-dev-container-bau.md) §6; sonst gäbe es keinen Fall der Gegenproben zum Gerüst).

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

- [ ] Stopp nach dem Vertrag: `DEV_ZUSTAND` mit Form und Vorgabe, Prüfung des Labels vor jeder Wirkung, Schritte in fester Reihenfolge ohne `--force`, Wiederholbarkeit, `behalten` und `loeschen` mit den Volumes des Labels, Stopp berührt nur den eigenen Container, erster gescheiterter Schritt und Signale (Zeilen 42 bis 50 der Tabelle in §3, vom Architect).
- [ ] Handläufe auf `u-boot`: ein Lauf über dem Speicherlimit lässt den Host unberührt (der Kernel wählt den Prozess und beendet ihn im Container; der Container endet nur, wenn es der erste ist, `SPEC-215` Nr. 9 b; gemessen auf `u-boot`, Zahl im Bericht), der innere Daemon erreicht das Netz für `deps` und der Adressbereich überlappt kein Außennetz, zwei Container laufen ohne Kollision, der Daemon des Hosts trägt nach Start und Stopp mit `loeschen` nur die Spur (Zeilen 51 bis 54).
- [ ] Zu jeder Zusage des Teils ein Fänger in der Tabelle *Zusage · Fänger* (§3) und, wo ein Skript oder Gate sie hält, ein Fall der Gegenproben; `make gegenproben` ist gefahren; die Zeile in „Nicht behauptet“ von `harness/README.md` nennt, was nur `u-boot` belegt (`SPEC-215` Nr. 12).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: eine Zeile für `dev-stop` in der Tabelle der **Werkzeuge** von `harness/README.md` (Entscheidung des Architect, `SPEC-215` Nr. 1); `AGENTS.md` §3.1 bleibt unberührt, es sei denn, der User entscheidet es in einer eigenen ADR anders.
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
| `tools/harness/dev-stop.sh`; `harness/mk/dev.mk` | neu; update | der Treiber des Stopps und das Ziel `dev-stop`, nicht in `GATE_CHECKS` (`SPEC-215` Nr. 1); `dev-lib.sh` wächst nicht <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) --> |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh`, `Dockerfile` (Stufe `gegenproben`, eine eigene `COPY`-Zeile für `dev-stop.sh`), `.dockerignore`, `SPEC-102` | update | Fälle mit Ersatz für `docker` (Entscheidung 8, `SPEC-215` Nr. 11); der Treiber ist Eingabe; der Stempel erst im Integrationsschritt |
| `harness/README.md` | update | eine Zeile für `dev-stop` in der Tabelle der Werkzeuge, „Nicht behauptet“ für den Teil, den nur `u-boot` belegt (`SPEC-215` Nr. 12) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15), die Mengen, die Prüfbarkeit und die erwartete Größe stehen unten und sind vom Architect (2026-11-23) vor dem Übergang `next` → `in-progress` geschrieben. Die Zeilen 42 bis 54 des ungeschnittenen Plans hat der Planner beim Schnitt im Wortlaut übernommen, nur die Verweise in den Zeilen 51 und 53 auf Zeilen der Teile I und II tragen die Angabe des Teils (jede Zeile steht in genau einem der drei Teile). Die Zeile nennt in Klammern den Ort im Vertrag; jede Mutation fährt der Implementer selbst (`AGENTS.md` §3.10).

**Tabelle *Zusage · Fänger*.** Jeder Fänger ist ein Fall der Gegenprobe mit einem Ersatz für `docker` (`SPEC-215` Nr. 11), wo nichts anderes steht; „Schleife“ heißt: ein Fall je Element der Menge (unten), die Zählprüfung nennt ihre Zahl. Handläufe und Negative sind benannt. Die Zeilen 1 bis 20 stehen in [slice-dev-container-bau](../next/slice-dev-container-bau.md), die Zeilen 21 bis 41 in [slice-dev-container-start](slice-dev-container-start.md).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 42 | `DEV_ZUSTAND` ist `behalten` oder `loeschen`; jede andere Form ist Exit 2 ohne Aufruf von `docker`; `DEV_NAME` folgt für den Stopp der Achse von Zeile 17 (Nr. 3) | Schleife über M15 (5 Verstöße; Gutfälle `behalten`, `loeschen`, nicht gesetzt); `löschen` aus Bytewerten gebaut und gegengelesen (`AGENTS.md` §3.18) | `loeschen` als Teilstring; leer als Vorgabe | andere |
| 43 | Der Stopp prüft vor jeder Wirkung, dass Container, Netz und (bei `loeschen`) Volumes das Label tragen; sonst Exit 2 ohne Wirkung (Nr. 7 a) | Schleife über Objektart (Container, Netz, Volume `<n>-docker`, Volume `<n>-repos`) × {`behalten`, `loeschen`} × {Container vorhanden, fehlt} = 14 Lagen (ein fremder Container fehlt nicht), das fremde Objekt ohne Label, die anderen eigen: bei `behalten` bleibt ein fremdes Volume unbeachtet (Exit 0), sonst Exit 2; bei Exit 2 Zähler von `stop`, `rm`, `network rm`, `volume rm` je 0 | Prüfung erst nach dem ersten Schritt; Label nicht geprüft | andere |
| 44 | Die Schritte laufen in der Reihenfolge `stop --time 30`, `rm` (ohne `--force`), `network rm`, bei `loeschen` je `volume rm` (Nr. 7 b) | Aufrufliste des Ersatzes für die Zustände running, exited, created (`stop` immer), Wortlaut der Argumente | `--force`; `--time 10`; `rm` vor `stop` | andere |
| 45 | Ein Container, der nicht besteht, ist Exit 0 mit `nicht vorhanden`, der zweite Stopp hintereinander gleich; bei `loeschen` gehen Netz und Volumes trotzdem (Nr. 7 c, d) | Schleife über M22: Container, Netz, Volumes je vorhanden oder fehlend (8) × `behalten`, `loeschen` (2) = 16; ein Schritt nur für ein vorhandenes Objekt; zweimaliger Aufruf gleiche Ausgabe | `stop` ohne Existenzprüfung (Exit 1); Abbruch bei fehlendem Container vor `loeschen` | andere |
| 46 | `behalten` (Vorgabe) entfernt kein Volume (Nr. 7 b, 8 a) | zwei Fälle: Variable nicht gesetzt, `behalten`: Zähler `volume rm` null, Meldung `Zustand behalten (…)` | `volume rm` bei jedem Stopp | andere |
| 47 | `loeschen` entfernt jedes Volume mit dem Label des Containers, auch ein später hinzugekommenes, und nur diese (Nr. 7 b, 8 a) | Schleife über M21: `volume ls` meldet 0 bis 3 Volumes (das dritte `<n>-ssh`): je ein `volume rm` mit genau diesem Namen; ein Volume ohne Label fehlt in der Liste und bleibt | feste Namen statt der Liste; Label-Filter fehlt im `ls` | andere |
| 48 | Der Stopp von `a` lässt Container, Netz und Volumes von `a-b` und `b` unberührt (Nr. 7 a, 8 a) | Wirkungs-Fänger: der Ersatz hält zwei Container (`a`, `a-b`) mit je zwei Volumes und Netz; nach `dev-stop` mit `loeschen` für `a` kommt kein Name von `a-b` oder `b` in den Argumenten der verändernden Aufrufe vor | Filter nach dem Präfix `a-`; `volume rm` über alle Volumes | nebenläufig |
| 49 | Der erste scheiternde Schritt beendet den Stopp mit Exit 1 und `<Schritt> gescheitert`; kein späterer Schritt läuft (Nr. 7 b) | Schleife über Schritt k (`stop`, `rm`, `network rm`, erstes und zweites `volume rm`) = 5 (M20): der Ersatz lässt Schritt k mit Exit 1 enden; Zähler der späteren Aufrufe je 0 | Fehler des Schritts ignoriert; Meldung ohne Schrittnamen | andere |
| 50 | INT, HUP und TERM an den Stopp enden mit Exit 2 und `abgebrochen`, ohne späteren Schritt; ein zweiter Aufruf stellt den Stand her (Nr. 10, 7 c) | Schleife über 4 Phasen (`stop`, `rm`, `network rm`, `volume rm`) × 3 Signale = 12; der Ersatz schläft 1 s im Schritt; danach ein zweiter Lauf auf dem Rest: Exit 0 | Handler fehlt; nächster Schritt nach dem Signal | nebenläufig |
| 51 | Eine Überschreitung des Speichers im Container lässt den Host unberührt; der Kernel beendet Prozesse im Container (Nr. 9 b) | kein Gate: Handlauf auf `u-boot` mit einem Prozess, der über das Limit wächst, in einem Container mit `DEV_MEMORY=1g`: Exit 137 des Prozesses, Daemon des Hosts antwortet, `free` vor und nach; die Mutation ohne `--memory` wird nicht gefahren (sie belastete den Host); Bedingung 4 (Zeile 13, Teil I) und Zeile 29 (Teil II) fangen die fehlende Option; Negativ Nr. 12 a | — | andere |
| 52 | Der innere Daemon erreicht das Netz für `deps`, der Adressbereich überlappt kein Außennetz, innere Prüfläufe bleiben `--network none` (Nr. 9 c) | kein automatischer Fall: Handlauf auf `u-boot` (ein `docker pull` im inneren Daemon; ein innerer Container mit `--network none` ohne Ausgang); Negativ Nr. 12 e | — | andere |
| 53 | Zwei Container (`a`, `b`) laufen zugleich ohne Kollision; der Stopp von `a` lässt `b` bereit (Nr. 8 a, 7 a) | Handlauf auf `u-boot`: beide Namen, `make dev-probe` je Container vor und nach dem Stopp von `a` (Zeit, Last und Messender im Plan); darüber hinaus nur die Zeilen 32 (Teil II) und 48; Negativ Nr. 12 g | — | nebenläufig |
| 54 | Nach Start und Stopp trägt der Daemon des Hosts nichts außer der Spur (Image, Container, Volumes und Netz des Arbeitscontainers) (Nr. 8 d) | Handlauf auf `u-boot`: `docker images`, `ps -a`, `volume ls`, `network ls` vor dem Start und nach dem Stopp mit `loeschen`; die Mengen sind gleich bis auf das Image | — | andere |

Zeilen: 42 bis 54 gleich 13 (*a* = 3: Zeilen 48, 50, 53; *b* = 10). Die Klasse `nebenläufig` setzt der Architect für Zeilen, deren Code oder Fänger nebenläufig ist (`SPEC-128` Nr. 3 e); die Fälle der Zeilen 48 und 50 sind keine Läufe mit Wahrscheinlichkeit (der Ersatz hält den Aufruf fest, das Signal kommt danach), der Implementer fährt jede ihrer Mutationen zwanzigmal (`-count=20`) und nennt die Rate (`AGENTS.md` §3.10); Zeile 53 ist ein Handlauf. Größe und Schnitt: §4.

**Mengen** (`AGENTS.md` §3.15, §3.17): Achsen und Zählprüfung. Der Test schreibt jede als Schleife über Verstöße und Gutfälle und prüft die Zahl (`zahl M<n> "$n" <k>`); ein Element, das ein Verstoß ist, ist Exit 2 (Variable) oder Exit 1 (Schritt) mit der Meldung des Vertrags; ein Gutfall kommt durch.

| Menge | Elemente | Achsen |
|---|---|---|
| M15 `DEV_ZUSTAND` | 5 Verstöße (leer; `LOESCHEN`; `ja`; `löschen` mit Umlaut aus Bytewerten; `behalten` mit Leerraum), 3 Gutfälle (`behalten`, `loeschen`, nicht gesetzt) | Schreibweise, Umlaut, Leerraum |
| M19 Abbruch, dieser Teil | 12 (4 Phasen des Stopps × 3 Signale); die 21 weiteren Elemente der Menge von 33 liegen in den Teilen I (9) und II (12) | Treiber × Phase × Signal |
| M20 Schritte des Stopps | 5 (`stop`, `rm`, `network rm`, erstes und zweites `volume rm`) | Schritt |
| M21 Volumes | 4 (0 bis 3 mit Label, das dritte später hinzugekommen) | Zahl, Herkunft des Namens |
| M22 Lagen des Stopps | 16 (Container, Netz, Volumes je vorhanden oder fehlend = 8, × `behalten`, `loeschen`) | Objekt × Variable |

**Zusatztabelle** (`AGENTS.md` §3.16), Paare von Ursachen (je Paar die Entscheidung und der Fänger; die Reihenfolge der Prüfungen steht in `SPEC-215` Nr. 4 a); die Zeilen der Zusatztabelle des ungeschnittenen Plans, soweit Fänger dieses Teils sie tragen:

| Paar | Entscheidung | Fänger dieses Teils |
|---|---|---|
| fremdes Objekt und Container nicht vorhanden (Stopp) | das fremde Objekt: Exit 2 (Nr. 7 a) | Zeile 43 (Lagen mit Container vorhanden und fehlend) |
| Signal und gescheiterter Schritt (Stopp) | das Signal gilt: Exit 2 (Nr. 10); ein Aufruf stellt den Stand her | Zeile 50 (zweiter Lauf) |

**Prüfbarkeit** (Auftrag 8). *Im Gate:* die Fälle der Gegenproben rufen den Treiber im Wegwerfverzeichnis mit einem Ersatz für `docker` am Anfang von `PATH` (`SPEC-215` Nr. 11; Muster der Fälle 103 bis 114 von `SPEC-102`); sie belegen Reihenfolge, Argumente von `docker`, Zähler, Exit-Codes und Ströme, nicht den Daemon. Die Fälle mit Wartezeit setzen `DEV_FRIST` auf höchstens 3. Der Implementer misst die Laufzeit von `make gegenproben` je Teil nach der Messvorschrift von `SPEC-102` (Soll nach [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)) und trägt sie ein; die Fälle der Zeilen 48 und 50 sind keine Läufe mit Wahrscheinlichkeit (der Ersatz hält den Aufruf fest, das Signal kommt danach), und der Implementer fährt jede ihrer Mutationen zwanzigmal (`-count=20`) und nennt die Rate (`AGENTS.md` §3.10). *Nur ein Lauf auf `u-boot`:* die Zeilen 51, 52, 53 und 54 (Handläufe mit Zahlen: Zeit bis bereit, `memory.max` und `cpu.max`, Spitze des Speichers, Last des Hosts und Messender, nach `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`; die Zeile 40 des Starts liegt in Teil II). *Akzeptierte Negative* (Grund und Folge im Vertrag, `SPEC-215` Nr. 12; Zeile in „Nicht behauptet“ von `harness/README.md`, Wortlaut des Architect: „der Start des Arbeitscontainers auf `u-boot`, seine Grenzen im Container, die Netze des inneren Daemons und die Spur auf dem Daemon des Hosts (`SPEC-215` Nr. 12 a, e), die Antworten des Ersatzes für `docker` (Nr. 12 b), die Verdrahtung der vier Ziele (Nr. 12 d) und SIGKILL (Nr. 12 f)“): 12 a bis 12 i. Ein `u-boot`-Lauf ist ein Handlauf des Implementers und einer des Verifiers. Die Zeile in „Nicht behauptet“ schreibt dieser Teil, weil mit ihm alle vier Ziele und alle Handläufe vorliegen.

**Erwartete Größe** (Architect, 2026-11-23, für die Marke des Planner nach `SPEC-128` Nr. 1; Produktcode in Zeilen von Skripten, Dockerfile und Fragment ohne Gegenproben und Tests): Teil III rund 120 (`dev-stop.sh` 110, `dev.mk` 6, `harness/README.md` 4; `dev-lib.sh` wächst nicht). Die Zeile `Dockerfile`, `.dockerignore` und `gegenproben-eingabe.sh` zählt einige Zeilen und steckt in *p*. Die Formel ist an Go-Slices geeicht; an Skripten belegt sie hier nur `slice-guard-schreibwege` (Messung 1707 von Marke 2600, Geltungsgrenze `SPEC-128` Nr. 2); die Fälle der Gegenproben sind Go-Code mit Fixtures und Shell-Ersatz. <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) -->

Betriebsregeln für alle Rollen dieses Slice: Repo-Dateien nur mit `Edit` und `Write` (kein `sed -i`, `awk`, Python, Heredoc in die Datei; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); jeder `docker run` einer Rolle mit `--memory` und `--memory-swap`, höchstens zwei Container zugleich (Erfahrung vom 2026-10-24: Mutationskopien belegten ein Vielfaches des Bedarfs); Werte mit Escape-Folge oder Rohbytes aus Bytewerten und gegengelesen (`AGENTS.md` §3.18).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-dev-container-start](slice-dev-container-start.md) (Teil II) und mit ihm [slice-dev-container-bau](../next/slice-dev-container-bau.md) (Teil I) liegen in `done/` (der Stopp ruft die Bibliothek und prüft dieselben Objekte); [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) hat den Status `Accepted` (Bedingung von Teil I, hier mit ihm erfüllt); Vertrag, Tabelle *Zusage · Fänger* und Raster aus §6 stehen vor dem ersten Code-Commit (Architect, 2026-11-23, geliefert); der Planner hat die Marke gesetzt; kein anderer Slice der Gate-Kette liegt in `in-progress/` (Übergangsregeln Nr. 7: dieser Teil ändert `harness/mk/dev.mk`, `Dockerfile`, `.dockerignore`, `harness/README.md`, `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` und den Stempel). Nach `done/` wartet nur `slice-dev-container-zugang` auf diesen Teil (Reihenfolge in `welle-dev-container` §5). <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->

**Warum:** Teil II liefert den Start, den der Stopp aufräumt, und `dev.mk`; Teil I die Bibliothek (*Code-Abhängigkeit*). Die Handläufe der Zeilen 51 bis 54 brauchen den Stopp, um den Host sauber zu hinterlassen.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1; Teil des Architect, 2026-11-23, im Block unten; der Planner hat die Zahlen am 2026-11-24 aus der Tabelle nachgerechnet, **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 13 Zeilen der Tabelle (42 bis 54), *a* = 3 (Zeilen 48, 50, 53), *b* = 10, *f* = 0 (der Ersatz für `docker` ist ein Skript im Fall, kein Port-Fake), *p* = 120: *S* = 85·3 + 80·10 + 200·0 + 120 = 255 + 800 + 120 = 1175, 135·1175 = 158625, **Marke 1600** = 100 · ⌈15,8625⌉; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`. Ein Schnitt in zwei Teile trüge nicht: Teil II und III zusammen liegen bei *S* = 3111, Teil I und II bei 3926 (Architect). Summe der drei Teile: 1990 (Teil I) + 1936 (Teil II) + 1175 = 5101; der ungeschnittene Plan (5100) zählte die Zeile von `dev-start` in `harness/README.md` nicht, die Abweichung ist diese eine Zeile (Architect, 2026-11-24, [slice-dev-container-start](slice-dev-container-start.md)). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` zählt nicht, `docs/` und `spec/` nennt kein Pfad.

```groessenmarke
pfad: tools/harness/dev-stop.sh
pfad: harness/mk/dev.mk
pfad: tools/gegenproben
pfad: tools/harness/gegenproben-eingabe.sh
pfad: Dockerfile
pfad: .dockerignore
pfad: harness/README.md
produktcode: 120
fakes: 0
```

Der Block nennt `tools/gegenproben` als Verzeichnis, weil die Namen der neuen Fälle vor dem Code nicht feststehen: der Implementer nennt jede Datei, die er dort anlegt, im selben Commit im Block (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`). `harness/mk/dev.mk` entsteht in Teil I, die Zeilen der Teile I und II liegen vor der Basis dieses Teils und zählen nicht. `Dockerfile`, `.dockerignore`, `harness/README.md` und `tools/gegenproben` ändern auch andere Slices der Gate-Kette; sie laufen nacheinander (Übergangsregeln Nr. 7), und der Planner prüft beim Start, dass die Basis hinter dem letzten Slice der Kette liegt (§6, Eintrag *fremde Zeilen*). Der Stempel steht nicht im Block (Integrationsschritt). <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->

**Reihenfolge und Kollisionen** (Architect, 2026-11-23; die Reihenfolge der Kette legt der Planner fest, Roadmap, Drift-Log 2026-11-24). In `open/` führen Pfade der Kette (Treffer der Pfadsuche nach `harness/mk`, `Dockerfile`, `tools/gegenproben`, `Gate-Kette`, vom Planner gegen Nr. 7 geprüft): `slice-worktree-parallel` (Image-Tags und Sperre der Gates; die Tags `odata-recorder:*` dort, dieser Slice nutzt `odata-dev:arbeit` und kollidiert im Tag nicht, in `Dockerfile` und `harness/mk/` teilen beide Dateien), `slice-gegenproben-eichung` (misst, baut nichts um, `SPEC-102` Messvorschrift, Stempel), `slice-test-security`, `slice-coverage-gate`, `slice-abdeckung-deklaration`, `slice-lasttest` (Teil 3), `slice-handbuch-mock` und `slice-handbuch-telemetrie` (Stufe `handbuch`, `tools/gegenproben/`, Stempel), `slice-container-image` (wartet auf diese Welle, Kante in `welle-dev-container` §5), dazu der Slice (a) von [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md) (Stempel je Gate, `Makefile`, `harness/mk/`, `harness/README.md`; noch kein Slice). Dateien mit Konflikt auf derselben Zeile: die `COPY`-Zeilen der Stufe `gegenproben` im `Dockerfile`, die Liste `EINGABEN` und die Positivliste von `.dockerignore`; der Implementer legt je Treiber eine eigene Zeile an und ändert keine fremde. `harness/mk/dev.mk` ist eine neue Datei (Teil I); das `Makefile` bindet `harness/mk/*.mk` ein und bleibt unverändert. <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein weiterer Schnitt ist nicht vorgesehen, ein Teil, der die Schnittgrenze nach dem Code überschreitet, wird mit Messung und Grund berichtet (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert): [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) wird `Rejected` oder `Superseded`, oder `u-boot` erlaubt den gewählten Weg nicht (Befund an den Planner, die Welle wird umgeplant, Drift-Log).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, je Zusage des Teils ist eine Mutation rot gesehen (`AGENTS.md` §3.10), und ein Start auf `u-boot` ist mit Zahlen belegt (Zeit bis zur Bereitschaft, Limit, Speicher beim Lauf; Zeile 40 in Teil II, Zeilen 51 bis 54 hier); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Prüfbarkeit: Der privilegierte Start läuft nicht in `make gates` (Docker-only, netzlos, ohne schreibbaren Mount); was nur `u-boot` belegt, ist ohne Sensor — **Ausgang:** weiter offen; Fänger und akzeptiertes Negativ mit Grund und Folge stehen in §3 (Zeilen 51 bis 54, *Prüfbarkeit*; Zeile 40 in Teil II) und in `SPEC-215` Nr. 12; die Zeile für „Nicht behauptet“ steht in §3, der Implementer trägt sie in `harness/README.md` ein.
- Gemeinsamer Daemon-Zustand: Image-Tags `odata-recorder:*` kollidieren auf dem Host, wenn der Container den Daemon des Hosts nutzt statt eines eigenen — **Ausgang:** weiter offen: der Vertrag sagt zu, dass der innere Daemon allein baut und eine andere Kennung trägt als der des Hosts (`SPEC-215` Nr. 6, Bedingung 3, Zeile 12 in Teil I) und dass der Start auf dem Host nur die Spur anlegt (Nr. 8 d, Zeile 41 in Teil II, Zeile 54 hier); der Nachweis im Lauf von `make gates` (kein Image des Laufs auf dem Host) liegt beim Slice des Laufs.
- Nebenläufigkeit zweier Container auf einem Host (Port, Name, Volume): Randform des Rasters — **Ausgang:** weiter offen; Zeilen 32 und 38 in Teil II, Zeilen 48 und 53 hier (Name und Volume je Container, Konflikt beim `docker run`, Stopp berührt nur das Eigene, Handlauf mit zwei Containern); einen Port gibt es im Gerüst nicht.
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* unten; jede Zelle trägt einen Ort im Vertrag oder „gilt nicht“ mit Grund (Architect, 2026-11-23); der Implementer hält an einer Randform an, die dort nicht steht — **Ausgang:** weiter offen bis zur Closure.
- Die Überschreitung des Speichers beendet im Container Prozesse, nicht notwendig den Container (`SPEC-215` Nr. 9 b); §2 und der Closure-Trigger der Welle sagten „beendet den Container“ — **Ausgang:** entschieden (User, 2026-11-24, Frage 4 der ADR zur Privilegierung: der Kernel wählt; Planner, 2026-11-24: §2 und `welle-dev-container` §3 tragen den Wortlaut von `SPEC-215` Nr. 9 b).
- Fremde Zeilen im Block (`SPEC-128` Nr. 10 g): `Dockerfile`, `.dockerignore`, `harness/README.md` und `tools/gegenproben` ändern auch andere Slices der Kette — **Ausgang:** offen; Adresse: Planner (beim Start die Basis hinter dem letzten Slice der Kette; beim Handoff die Zeilen aus `main` trennen und mit Grund berichten).
- Der Schnitt vor dem Code (Schätzung des ungeschnittenen Plans 5100, Marke 6900) — **Ausgang:** entschieden (Planner, 2026-11-24, nach dem Vorschlag des Architect): Schnitt in drei Teile mit den Marken 2700, 2700 und 1600 (dieser Teil), nachgerechnet aus der Tabelle (keine Abweichung bei den Zahlen); dieser Teil trägt die Kennung `slice-dev-container-geruest`; `basis`, `schaetzung` und `marke` setzt der Planner je Teil beim Start (§4).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23; die Zeile dieses Teils; die Zeilen der Teile I und II stehen in deren Plänen). Die Zelle nennt den Ort in `SPEC-215` oder „gilt nicht“ mit Grund; „Variable“ meint die Variablen von Nr. 3, `DEV_NAME` hat jede Operation, `DEV_ZUSTAND` der Stopp.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Stopp (`dev-stop`) | Container nicht vorhanden: Nr. 7 c, d (Exit 0, `nicht vorhanden`, bei `loeschen` gehen Netz und Volumes trotzdem) | `DEV_ZUSTAND` leer: Nr. 3 (Exit 2, kein `docker`) | `DEV_ZUSTAND` unbekannt: Nr. 3; Objekt ohne das Label des Containers: Nr. 7 a (Exit 2, nichts verändert) | INT, HUP, TERM, ein Schritt scheitert: Nr. 10 und 7 b (Exit 2; Exit 1 und kein späterer Schritt) | Stopp eines gestoppten oder nicht vorhandenen Containers: Nr. 7 c (Exit 0, gleiche Ausgabe) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, Stand 2026-10-30, beim Schnitt am 2026-11-24 erneut; keiner steht bei 3× und offen). Für diesen Slice zählen: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, `offen`) — die Messungen im Container laufen unter einem festen Limit; Raten und Zeiten nennen Limit, Last und Messenden, und erreicht ein Fund mit dieser Welle 3×, braucht der Eintrag einen Ausgang (Welle-Closure); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) — die Betriebsregeln stehen in §3, ein Auftreten wird als Beleg eingetragen. Die Erfahrung vom 2026-10-24 (Speichermangel des Hosts durch Mutationskopien, Container ohne Limit) hat **keinen Eintrag im Register**; sie liegt als Notiz des Users vor und ist Anlass dieser Welle, nicht gezählt — die Closure entscheidet, ob ein Eintrag mit Beleg `evidence/slice-dev-container-geruest.md` entsteht.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
