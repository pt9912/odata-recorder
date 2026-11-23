# Slice slice-dev-container-start: Start des Arbeitscontainers mit Limits, Volumes, Netz und Bereitschaftswarten (Teil II des Schnitts von `slice-dev-container-geruest`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-dev-container](../welle-dev-container.md).

**Bezug:** `LH-QA-06` (gleiches Testverhalten bei gleicher Konfiguration, Container), `LH-RB-03` (Entwicklung unter Linux und in Docker-Umgebungen), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Folgepflicht: der Arbeitscontainer fährt `make gegenproben`), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Vertrag unter `harness/vertraege/`), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Stufen des `Dockerfile`). Entscheidungen des Architect (2026-11-23): [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) (Privilegierung, Limits, Netz; `Proposed`, die Entscheidung steht beim User aus).

**Berührte Spec-Stellen** (Teil II des Schnitts von [slice-dev-container-geruest](slice-dev-container-geruest.md); die Stellen der Teile I und III stehen in deren Plänen): `SPEC-215` (Vertrag, [`SPEC-215-arbeitscontainer.md`](../../../../harness/vertraege/SPEC-215-arbeitscontainer.md): Nr. 3, 4, 5, 8, 9, 10, 11, 12 a) · `SPEC-102` (die Fälle zum Start kommen als Nachzug des Implementers hinzu) · `AGENTS.md` §3.1 (Docker-only): bleibt unberührt, [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) Entscheidung, Einordnung nach §3.1 (Vorschlag).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-dev-container-geruest](slice-dev-container-geruest.md), Vorschlag des Architect vom 2026-11-23). **Datum:** 2026-11-24 (der Plan des ungeteilten Slice vom 2026-10-30 trug Ziel, Tabelle und Raster in einer Datei).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Auf dem Host `u-boot` startet `dev-start` einen Container aus dem Image von Teil I mit eigenem Docker-Daemon unter festem Speicher- und CPU-Limit (Variablen `DEV_MEMORY`, `DEV_CPUS`, `DEV_FRIST`, `SPEC-215` Nr. 3), legt Volumes und Netz nur an, wenn sie fehlen (Nr. 5, 8), und meldet erst nach der Bereitschaftsprobe aus Teil I, dass der Daemon in ihm Befehle annimmt; ein Start, der nicht bereit wird, endet nach der Frist rot und räumt nur den eigenen Container ab (Nr. 5 e bis g). **Schnitt vor dem Code** (Vorschlag des Architect, Planner, 2026-11-24): drei Teile, die nacheinander laufen und je einzeln lieferbar sind: **I** [slice-dev-container-bau](slice-dev-container-bau.md) (Bau und Probe, Zeilen 1 bis 20), **II** (diese Datei, Start, Zeilen 21 bis 41) und **III** [slice-dev-container-geruest](slice-dev-container-geruest.md) (Stopp und Handläufe auf `u-boot`, Zeilen 42 bis 54; trägt die Kennung des ungeteilten Slice). Die Handläufe des Starts (Zeile 40) gehören zu diesem Teil, weil hier zum ersten Mal ein Container läuft; Aufräumen übernimmt dort von Hand `docker rm --force`.

**Entscheidungen des Architect** (2026-11-23, soweit dieser Teil sie trägt; der Wortlaut aller neun Aufträge und ihr Ausgang stehen in [slice-dev-container-geruest](slice-dev-container-geruest.md) §1, der Vertrag trägt die Gründe, dieser Plan wiederholt sie nicht, `AGENTS.md` §3.8):

| Auftrag | Entscheidung | Ort |
|---|---|---|
| 2 Limits | Ort ist die Befehlszeile des Außencontainers (die cgroup trifft BuildKit und jeden inneren Container, Bedingung 4 liest die Grenze im Container nach); Vorgaben 10 GiB, kein Swap, 8 CPUs, 8192 Prozesse; bei Überschreitung beendet der Kernel Prozesse im Container und lässt den Host unberührt, ob der Container endet, ist Entscheidung des Users ([ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) Frage 4, Teil III) | `SPEC-215` Nr. 9 |
| 5 Zustand | zwei benannte Volumes (`/var/lib/docker`, `/srv/git`) überleben Stopp und Neustart; der frische Start meldet `Zustand neu` | `SPEC-215` Nr. 8 |
| 6 Netz | eigenes Netz je Container, Ausgang per NAT, kein Port; die Prüfläufe im inneren Daemon bleiben `--network none`. **Ausgang: offen, Adresse User** (mit Auftrag 1) | `SPEC-215` Nr. 9 c |
| 8 Prüfbarkeit | ein Ersatz für `docker` am Anfang von `PATH`; der echte Start belegt nur ein Handlauf auf `u-boot` (Zeile 40, akzeptiertes Negativ) | `SPEC-215` Nr. 11, 12 a |

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Image, Pins, Bau und Probe sowie die Bibliothek `dev-lib.sh` — [slice-dev-container-bau](slice-dev-container-bau.md) (Teil I, liegt bei diesem Start in `done/`); dieser Teil ruft sie.
- Der Stopp, `DEV_ZUSTAND` und die Handläufe der Zeilen 51 bis 54 (Speicherüberschreitung, Netz, zwei Container, Spur auf dem Host) — [slice-dev-container-geruest](slice-dev-container-geruest.md) (Teil III): sie brauchen den Stopp, um den Host sauber zu hinterlassen.
- SSH-Server, Schlüsselpaar und Host-Schlüssel — `slice-dev-container-zugang`; das Gerüst läuft ohne Zugang von außen.
- Bare-Repo, Klon und Einlieferung — `slice-dev-container-einlieferung`; der Start legt nur das leere Volume `<n>-repos` an (`SPEC-215` Nr. 8 a), damit der Zustand von Anfang an vollständig benannt ist.
- Lauf von `make gates` und `make gegenproben` im Container — `slice-dev-container-lauf`; dieser Teil beweist, dass der Daemon antwortet, nicht, dass die Gates in ihm bestehen.
- Das Dokument für Entwickler — `slice-dev-container-dokument`.
- Ein Eingriff in die vorhandenen Gate-Stufen (`Dockerfile`: `test`, `build`, `race`, `lint`, `gegenproben`) — Bestand: der Container fährt sie, wie sie sind. Ausgenommen sind die Zeilen, die den neuen Treiber in die Stufe `gegenproben` kopieren und in `EINGABEN` von `tools/harness/gegenproben-eingabe.sh` und in die Positivliste von `.dockerignore` aufnehmen: sie ändern nicht, was die Stufe prüft, sondern was sie liest (Auftrag 8; vom Planner am 2026-11-24 bestätigt, [slice-dev-container-bau](slice-dev-container-bau.md) §6).

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

- [ ] Variablen, Prüfungen und `docker run`: `DEV_MEMORY`, `DEV_CPUS` und `DEV_FRIST` mit Form, Null, Host-Grenze und Vorgabe, die Reihenfolge der Prüfungen vor jeder Wirkung, die vierzehn Pflichtoptionen und keine der fünfzehn verbotenen Formen im Aufruf, Name, Volumes und Netz je Container ohne Kollision, Volumes und Netz nur bei Fehlen (Zeilen 21 bis 33 und 41 der Tabelle in §3, vom Architect).
- [ ] Start mit Limit und Bereitschaftsprobe nach dem Vertrag: ein frischer Container wird bereit und meldet es; ein Daemon, der nicht bereit wird, endet nach der Frist rot und räumt nur den eigenen Container ab; Signale und Namenskonflikt räumen nichts Fremdes ab (Zeilen 34 bis 39); ein Start auf `u-boot` wird bereit (Handlauf, Zeile 40: Zeit bis bereit, `memory.max`, `cpu.max`, Spitze des Speichers, Last des Hosts und Messender im Bericht).
- [ ] Zu jeder Zusage des Teils ein Fänger in der Tabelle *Zusage · Fänger* (§3) und, wo ein Skript oder Gate sie hält, ein Fall der Gegenproben; `make gegenproben` ist gefahren.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: eine Zeile für `dev-start` in der Tabelle der **Werkzeuge** von `harness/README.md` (Entscheidung des Architect, `SPEC-215` Nr. 1); `AGENTS.md` §3.1 bleibt unberührt, es sei denn, der User entscheidet es in einer eigenen ADR anders.
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
| `tools/harness/dev-start.sh`; `harness/mk/dev.mk` | neu; update | der Treiber des Starts und das Ziel `dev-start`, nicht in `GATE_CHECKS` (`SPEC-215` Nr. 1); `dev-lib.sh` und `dev-probe.sh` liefert Teil I und ändert dieser Teil nicht <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) --> |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh`, `Dockerfile` (Stufe `gegenproben`, eine eigene `COPY`-Zeile für `dev-start.sh`), `.dockerignore`, `SPEC-102` | update | Fälle mit Ersatz für `docker` (Entscheidung 8, `SPEC-215` Nr. 11); der Treiber ist Eingabe; der Stempel erst im Integrationsschritt |
| `harness/README.md` | update | eine Zeile für `dev-start` in der Tabelle der Werkzeuge (`SPEC-215` Nr. 12) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15), die Mengen, die Prüfbarkeit und die erwartete Größe stehen unten und sind vom Architect (2026-11-23) vor dem Übergang `next` → `in-progress` geschrieben; die Zeilen 21 bis 41 des ungeschnittenen Plans hat der Planner beim Schnitt im Wortlaut übernommen, nur der Verweis in Zeile 34 auf die Zeile 8 trägt die Angabe des Teils (jede Zeile steht in genau einem der drei Teile). Die Zeile nennt in Klammern den Ort im Vertrag; jede Mutation fährt der Implementer selbst (`AGENTS.md` §3.10).

**Tabelle *Zusage · Fänger*.** Jeder Fänger ist ein Fall der Gegenprobe mit einem Ersatz für `docker` (`SPEC-215` Nr. 11), wo nichts anderes steht; „Schleife“ heißt: ein Fall je Element der Menge (unten), die Zählprüfung nennt ihre Zahl. Handläufe und Negative sind benannt. Die Zeilen 1 bis 20 stehen in [slice-dev-container-bau](slice-dev-container-bau.md), die Zeilen 42 bis 54 in [slice-dev-container-geruest](slice-dev-container-geruest.md).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
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
| 34 | Der Start wartet bis zur Bereitschaft und endet beim ersten Bestehen mit Exit 0 (Nr. 5 e) | die Probe besteht im k-ten Durchgang (k = 1, 2, 3, `DEV_FRIST` 10): Exit 0, `bereit`-Zeile, genau k Durchgänge, die Zeile `warte` einmal; dieselben Bedingungen wie Zeile 8 (Teil I) | Schleife läuft bis zur Frist; `bereit` ohne Probe | andere |
| 35 | Besteht die Probe bis zur Frist nicht, endet der Start mit Exit 1 und nennt die zuletzt verletzte Bedingung (Nr. 5 f) | Schleife über Bedingung 1 bis 5 als verletzte, `DEV_FRIST` 2: Exit 1, Nummer und Name in der Meldung, Dauer zwischen Frist und Frist plus 5 s; Gegenstück: bereit im letzten Durchgang vor der Frist, Exit 0 | Frist ohne Wirkung (der Fall reißt seine Frist); `>` zu `>=` | andere |
| 36 | Endet der Container vor der Bereitschaft, endet der Start sofort mit Exit 1 und dessen Exit-Code (Nr. 5 f) | Schleife über Code 137, 1, 0 und den Zeitpunkt (vor dem ersten, nach dem zweiten Durchgang): Dauer unter der Frist, Code in der Meldung | `State.Running` nicht gelesen; Code fest 137 | andere |
| 37 | Ein roter Start schreibt `logs --tail 40`, entfernt nur einen Container mit der eigenen Kennung und lässt Volumes und Netz stehen (Nr. 5 f, g) | Schleife über {Frist abgelaufen, Container endet} × {Kennung gleich, fremd}: Zähler von `logs`, `rm`, `volume rm`, `network rm` | `rm` ohne Kennungsprüfung; `volume rm` im roten Start | andere |
| 38 | Endet `docker run` mit einem Namenskonflikt, räumt der Start nichts ab (der Container gehört einem zweiten Start); endet es aus anderem Grund, räumt er einen halb angelegten Container mit eigener Kennung ab (Nr. 5 g) | Wirkungs-Fänger: der Ersatz antwortet auf `run` mit Exit 125 und `Conflict`, der Container trägt eine fremde Kennung: Exit 2, `docker run gescheitert`, Zähler von `rm` und `stop` je 0; Gegenstück: Exit 125 mit anderem Text, Container mit eigener Kennung: `rm` genau einmal | `rm --force <n>` ohne Kennungsprüfung | nebenläufig |
| 39 | INT, HUP und TERM an den Start enden mit Exit 2 und `abgebrochen`; der Container mit eigener Kennung wird entfernt, auch wenn das Signal während `docker run` eintrifft; Volumes und Netz bleiben (Nr. 10, 5 g) | Schleife über 3 Signale × 4 Phasen (vor `run`, während `run`, in der Probe-Schleife, während `logs`) = 12; „während `run`“: der Ersatz lässt `run` 1 s schlafen und erfolgreich enden, danach `rm` genau einmal; vor `run`: `rm` null | Handler entfernt nicht; Kennung erst nach `run` gesetzt (Signal während `run` ungefangen) | nebenläufig |
| 40 | Der Start auf `u-boot` wird bereit und meldet es; Zeit, `memory.max`, `cpu.max` und Spitze des Speichers stehen im Plan (Nr. 12 a) | kein automatischer Fall: Handlauf `make dev-start`, danach `make dev-probe`; Negativ Nr. 12 a (Grund: kein Daemon und kein Privileg in den Prüfläufen; Folge: der Treiber kann am echten `docker` scheitern, der Handlauf ist der Fänger) | — | andere |
| 41 | Der Start legt auf dem Daemon des Hosts nur Container, zwei Volumes und ein Netz an und zieht oder entfernt kein Image (Nr. 8 d) | Allaussage über die Aufrufe des Ersatzes in jedem Fall des Starts: jeder verändernde Aufruf ist `volume create` (2), `network create` (1), `run` (1) oder (rot) `rm` des eigenen Containers; Zähler von `pull`, `tag`, `rmi`, `image`, `build`, `cp` je 0 | `docker pull` im Start | andere |

Zeilen: 21 bis 41 gleich 21 (*a* = 3: Zeilen 32, 38, 39; *b* = 18). Die Klasse `nebenläufig` setzt der Architect für Zeilen, deren Code oder Fänger nebenläufig ist (`SPEC-128` Nr. 3 e); die Fälle der Zeilen 32, 38 und 39 sind keine Läufe mit Wahrscheinlichkeit (der Ersatz hält den Aufruf fest, das Signal kommt danach), der Implementer fährt jede ihrer Mutationen zwanzigmal (`-count=20`) und nennt die Rate (`AGENTS.md` §3.10). Größe und Schnitt: §4.

**Mengen** (`AGENTS.md` §3.15, §3.17): Achsen und Zählprüfung. Der Test schreibt jede als Schleife über Verstöße und Gutfälle und prüft die Zahl (`zahl M<n> "$n" <k>`); ein Element, das ein Verstoß ist, ist Exit 2 (Variable) mit der Meldung des Vertrags; ein Gutfall kommt durch.

| Menge | Elemente | Achsen |
|---|---|---|
| M12 `DEV_MEMORY` | 20 Verstöße (leer; `0`, `0g`, `0m`, `00g` als null; `10`; `10G`; `10M`; `10gb`; `10k`; `1.5g`; `-1g`; `+1g`; mit Leerraum davor und danach; `010g`; sieben Ziffern; `max`; `unlimited`; `-1`), 3 Gutfälle (`1m`, `512m`, `15g`) | Einheit, Ziffern, Vorzeichen, Leerraum, Wort |
| M13 `DEV_CPUS` | 11 Verstöße (leer; `0`; `00`; `08`; `0.5`; `8.0`; `-1`; `1e1`; mit Leerraum; `abc`; vier Ziffern), dazu Host-Grenze (`4` gut, `5` Exit 2), Gutfälle `1` und `4` | Form, Ziffern, Grenze `NCPU` |
| M14 `DEV_FRIST` | 8 Verstöße (leer; `0`; `601`; `1s`; `1.5`; `-5`; `0600`; vier Ziffern), 2 Gutfälle (`1`, `600`) | Form, Grenze 600 |
| M16 Pflichtoptionen | 14 (`--detach`, `--name`, `--label odata-dev.name`, `--label odata-dev.start`, `--restart no`, `--network`, `--mount` für `<n>-docker`, `--mount` für `<n>-repos`, `--memory`, `--memory-swap`, `--cpus`, `--pids-limit`, `--cgroupns private`, `--privileged`) | Option × Name × Limit |
| M17 verbotene Formen | 15 (die Nummern (1) bis (15) von Nr. 5 d) | Form × Name (Präfix-Paar) |
| M18 Fremdobjekte | 13 (Zustände `running`, `exited`, `created`, `paused`, `restarting` je mit und ohne Label; Volume `docker`, Volume `repos`, Netz je ohne Label) | Zustand × Label; Objektart |
| M19 Abbruch, dieser Teil | 12 (4 Phasen des Starts × 3 Signale); die 21 weiteren Elemente der Menge von 33 liegen in den Teilen I (9) und III (12) | Treiber × Phase × Signal |
| M23 Paare des Starts | 11 (sechs Paare der vier Variablen `DEV_NAME`, `DEV_MEMORY`, `DEV_CPUS`, `DEV_FRIST`; fünf Folgen der Umgebung) | Paar |
| M24 Lagen des Zustands | 8 (Volume `docker`, Volume `repos`, Netz je vorhanden oder fehlend) | Objekt |

**Zusatztabelle** (`AGENTS.md` §3.16), Paare von Ursachen (je Paar die Entscheidung und der Fänger; die Reihenfolge der Prüfungen steht in `SPEC-215` Nr. 4 a); die Zeilen der Zusatztabelle des ungeschnittenen Plans, soweit Fänger dieses Teils sie tragen:

| Paar | Entscheidung | Fänger dieses Teils |
|---|---|---|
| zwei ungültige Variablen am Start (sechs Paare aus `DEV_NAME`, `DEV_MEMORY`, `DEV_CPUS`, `DEV_FRIST`) | die frühere der Reihenfolge `DEV_NAME`, `DEV_MEMORY`, `DEV_CPUS`, `DEV_FRIST` wird gemeldet (Nr. 3) | Zeile 26 (sechs Paare) |
| ungültige Variable und `docker` unerreichbar | die Variable, kein Aufruf von `docker` (Nr. 4 a) | Zeile 26 (Zeile 19 in Teil I für Bau und Probe) |
| `docker` unerreichbar und cgroup v1; cgroup v1 und Image fehlt; Image fehlt und Name vergeben; Name vergeben und fremdes Volume | die frühere Prüfung von Nr. 4 a und 5 a | Zeile 26 (fünf Folgen) |
| Start abgebrochen und Name schon vergeben | der Start endet vor jeder Wirkung (`Name vergeben` oder `abgebrochen`, beide Exit 2); entfernt wird nichts, weil noch kein Container mit der eigenen Kennung besteht (Nr. 5 a, g, 10) | Zeilen 28 (kein `rm`) und 39 (Phase vor `run`, `rm` null) |
| Frist abgelaufen und Container einer fremden Kennung | kein `rm` (Nr. 5 g) | Zeile 37 |

**Prüfbarkeit** (Auftrag 8). *Im Gate:* die Fälle der Gegenproben rufen den Treiber im Wegwerfverzeichnis mit einem Ersatz für `docker` am Anfang von `PATH` (`SPEC-215` Nr. 11; Muster der Fälle 103 bis 114 von `SPEC-102`); sie belegen Reihenfolge, Argumente von `docker`, Zähler, Exit-Codes und Ströme, nicht den Daemon. Die Fälle mit Wartezeit setzen `DEV_FRIST` auf höchstens 3. Der Implementer misst die Laufzeit von `make gegenproben` nach der Messvorschrift von `SPEC-102` (Soll nach [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)) und trägt sie ein; die Fälle der Zeilen 32, 38 und 39 sind keine Läufe mit Wahrscheinlichkeit, der Implementer fährt jede ihrer Mutationen zwanzigmal (`-count=20`). *Nur ein Lauf auf `u-boot`:* die Zeile 40 (Handlauf mit Zahlen: Zeit bis bereit, `memory.max` und `cpu.max`, Spitze des Speichers, Last des Hosts und Messender, nach `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`); ein Handlauf des Implementers und einer des Verifiers. *Akzeptierte Negative* (Grund und Folge im Vertrag, `SPEC-215` Nr. 12): 12 a trifft diesen Teil; die Zeile in „Nicht behauptet“ von `harness/README.md` schreibt Teil III.

**Erwartete Größe** (Architect, 2026-11-23, für die Marke des Planner nach `SPEC-128` Nr. 1; Produktcode in Zeilen von Skripten, Dockerfile und Fragment ohne Gegenproben und Tests): Teil II rund 241 (`dev-start.sh` 230, `dev.mk` 10, `harness/README.md` 1: eine neue Zeile in der Tabelle der Werkzeuge, eine Einfügung ohne Löschung; `SPEC-128` Nr. 4 a zählt jede Zeile unter den Pfaden des Blocks, auch Prosa, nur `harness/vertraege/` nicht, und Teil I und III rechnen ihre Zeilen in `harness/README.md` in *p* ein, Berichtigung des Architect vom 2026-11-24). Die Zeile `Dockerfile`, `.dockerignore` und `gegenproben-eingabe.sh` zählt einige Zeilen und steckt in *p*. Die Formel ist an Go-Slices geeicht; an Skripten belegt sie hier nur `slice-guard-schreibwege` (Messung 1707 von Marke 2600, Geltungsgrenze `SPEC-128` Nr. 2); die Fälle der Gegenproben sind Go-Code mit Fixtures und Shell-Ersatz. <!-- d-check:ignore (geplante Dateien, entstehen mit dem Code) -->

Betriebsregeln für alle Rollen dieses Slice: Repo-Dateien nur mit `Edit` und `Write` (kein `sed -i`, `awk`, Python, Heredoc in die Datei; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); jeder `docker run` einer Rolle mit `--memory` und `--memory-swap`, höchstens zwei Container zugleich (Erfahrung vom 2026-10-24: Mutationskopien belegten ein Vielfaches des Bedarfs); Werte mit Escape-Folge oder Rohbytes aus Bytewerten und gegengelesen (`AGENTS.md` §3.18).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-dev-container-bau](slice-dev-container-bau.md) (Teil I) liegt in `done/` (der Start ruft die Probe und die Bibliothek); [ADR-0031](../../adr/0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md) hat den Status `Accepted` (Bedingung von Teil I, hier mit ihm erfüllt); Vertrag, Tabelle *Zusage · Fänger* und Raster aus §6 stehen vor dem ersten Code-Commit (Architect, 2026-11-23, geliefert); der Planner hat die Marke gesetzt; kein anderer Slice der Gate-Kette liegt in `in-progress/` (Übergangsregeln Nr. 7: dieser Teil ändert `harness/mk/dev.mk`, `Dockerfile`, `.dockerignore`, `harness/README.md`, `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` und den Stempel). Teil III ([slice-dev-container-geruest](slice-dev-container-geruest.md)) startet, wenn dieser Teil in `done/` liegt. <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->

**Warum:** Teil I liefert die Probe und `dev-lib.sh` (*Code-Abhängigkeit*); Teil III ruft den Treiber und die Bibliothek (*Code-Abhängigkeit*).

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1; Teil des Architect, 2026-11-23, im Block unten; der Planner hat die Zahlen am 2026-11-24 aus der Tabelle nachgerechnet, **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 21 Zeilen der Tabelle (21 bis 41), *a* = 3 (Zeilen 32, 38, 39), *b* = 18, *f* = 0 (der Ersatz für `docker` ist ein Skript im Fall, kein Port-Fake), *p* = 241 (240 für `dev-start.sh` und `dev.mk` und 1 für die Zeile in `harness/README.md`, die in der ersten Fassung fehlte; Architect, 2026-11-24, nach `SPEC-128` Nr. 1 und 4 a): *S* = 85·3 + 80·18 + 200·0 + 241 = 255 + 1440 + 241 = 1936, 135·1936 = 261360, **Marke 2700** = 100 · ⌈26,136⌉ (unverändert gegen *S* = 1935); keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`; der Teil liegt 287 unter der Schnittgrenze, jede weitere Zeile kostet 80 (*b*) oder 85 (*a*), die vierte weitere Zeile löst die Schnittfrage aus (*S* = 2256 mit vier Zeilen *b*). Summe der drei Teile: 1990 (Teil I) + 1936 + 1175 (Teil III) = 5101; der ungeschnittene Plan (5100) zählte die Zeile von Teil II nicht, die Abweichung ist diese eine Zeile. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` zählt nicht, `docs/` und `spec/` nennt kein Pfad.

```groessenmarke
pfad: tools/harness/dev-start.sh
pfad: harness/mk/dev.mk
pfad: tools/gegenproben
pfad: tools/harness/gegenproben-eingabe.sh
pfad: Dockerfile
pfad: .dockerignore
pfad: harness/README.md
produktcode: 241
fakes: 0
```

Der Block nennt `tools/gegenproben` als Verzeichnis, weil die Namen der neuen Fälle vor dem Code nicht feststehen: der Implementer nennt jede Datei, die er dort anlegt, im selben Commit im Block (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`). `harness/mk/dev.mk` entsteht in Teil I, die Zeilen von Teil I liegen vor der Basis dieses Teils und zählen nicht. `Dockerfile`, `.dockerignore`, `harness/README.md` und `tools/gegenproben` ändern auch andere Slices der Gate-Kette; sie laufen nacheinander (Übergangsregeln Nr. 7), und der Planner prüft beim Start, dass die Basis hinter dem letzten Slice der Kette liegt (§6, Eintrag *fremde Zeilen*). Der Stempel steht nicht im Block (Integrationsschritt). <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->

**Reihenfolge und Kollisionen** (Architect, 2026-11-23; die Reihenfolge der Kette legt der Planner fest, Roadmap, Drift-Log 2026-11-24): Teil I liefert `harness/mk/dev.mk`, dieser Teil hängt `dev-start` an; die Dateien mit Konflikt auf derselben Zeile sind die `COPY`-Zeilen der Stufe `gegenproben` im `Dockerfile`, die Liste `EINGABEN` und die Positivliste von `.dockerignore`: der Implementer legt für `dev-start.sh` eine eigene Zeile an und ändert keine fremde. Das `Makefile` bindet `harness/mk/*.mk` ein und bleibt unverändert. <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein weiterer Schnitt ist nicht vorgesehen, ein Teil, der die Schnittgrenze nach dem Code überschreitet, wird mit Messung und Grund berichtet (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert): `u-boot` erlaubt den gewählten Weg nicht (der Handlauf der Zeile 40 scheitert am Privileg oder am Daemon im Container; Befund an den Planner, die Welle wird umgeplant, Drift-Log).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, je Zusage des Teils ist eine Mutation rot gesehen (`AGENTS.md` §3.10), und ein Start auf `u-boot` ist mit Zahlen belegt (Zeit bis zur Bereitschaft, Limit, Speicher beim Lauf); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Prüfbarkeit: Der privilegierte Start läuft nicht in `make gates` (Docker-only, netzlos, ohne schreibbaren Mount); der Treiber kann am echten `docker` scheitern — **Ausgang:** weiter offen; Fänger und akzeptiertes Negativ mit Grund und Folge stehen in §3 (Zeile 40, *Prüfbarkeit*) und in `SPEC-215` Nr. 12 a; der Handlauf des Implementers und des Verifiers ist der Beleg.
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* unten; jede Zelle trägt einen Ort im Vertrag oder „gilt nicht“ mit Grund (Architect, 2026-11-23); der Implementer hält an einer Randform an, die dort nicht steht — **Ausgang:** weiter offen bis zur Closure.
- Nebenläufigkeit zweier Container auf einem Host (Name, Volume, Netz): die Fälle 32 und 38 fangen den Start (Name und Volume je Container, Konflikt beim `docker run`); der Stopp und der Handlauf mit zwei Containern liegen in Teil III — **Ausgang:** entschieden (Planner, 2026-11-24): der Start trägt die Zeilen 32 und 38, die Zeilen 48 und 53 tragen Teil III; einen Port gibt es im Gerüst nicht.
- Host-Annahmen (cgroup v2, Speichertreiber `overlayfs`, Docker 29.10.0, 33 362 579 456 Byte Speicher) sind am Host des Architect-Laufs gemessen; ob er `u-boot` ist, belegt der Handlauf von Zeile 40 — **Ausgang:** offen.
- Fremde Zeilen im Block (`SPEC-128` Nr. 10 g): `Dockerfile`, `.dockerignore`, `harness/README.md` und `tools/gegenproben` ändern auch andere Slices der Kette — **Ausgang:** offen; Adresse: Planner (beim Start die Basis hinter dem letzten Slice der Kette; beim Handoff die Zeilen aus `main` trennen und mit Grund berichten).
- Die Überschreitung des Speichers beendet im Container Prozesse, nicht notwendig den Container (`SPEC-215` Nr. 9 b): die Vorgabe `DEV_MEMORY` und die Option `--memory` setzt dieser Teil, die Aussage über das Ende des Containers prüft Teil III (Zeile 51) — **Ausgang:** siehe [slice-dev-container-geruest](slice-dev-container-geruest.md) §6 (offen, Adresse User, Frage 4 der ADR).
- Der Schnitt vor dem Code (Schätzung des ungeschnittenen Plans 5100, Marke 6900) — **Ausgang:** entschieden (Planner, 2026-11-24, nach dem Vorschlag des Architect): Schnitt in drei Teile mit den Marken 2700, 2700 (dieser Teil) und 1600, nachgerechnet aus der Tabelle (keine Abweichung bei den Zahlen); `basis`, `schaetzung` und `marke` setzt der Planner je Teil beim Start (§4).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23; die Zeile dieses Teils; die Zeilen der Teile I und III stehen in deren Plänen). Die Zelle nennt den Ort in `SPEC-215` oder „gilt nicht“ mit Grund; „Variable“ meint die Variablen von Nr. 3, `DEV_NAME` hat jede Operation, `DEV_MEMORY`, `DEV_CPUS` und `DEV_FRIST` der Start.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Start (`dev-start`) | Variable nicht gesetzt: Vorgabe (Nr. 3); Image fehlt: Nr. 5 a (Exit 2); Zustand fehlt: legt an und meldet `neu` (Nr. 5 b, 8 c) | Variable leer, null: Nr. 3 (Exit 2, Meldung `leer` und `null`, kein `docker`) | nicht parsbarer Wert und Form, Wert über dem Host: Nr. 3 (`Form`, `über dem Speicher`, `über den Prozessoren`); cgroup v1: Nr. 5 a; fremdes Objekt des Namens: Nr. 5 a | Signal während des Starts, Daemon startet nicht (Frist), Container endet: Nr. 10 und 5 f, g (nur der eigene Container wird entfernt, Exit 2, 1 und 1) | Container des Namens vorhanden (jeder Zustand): Nr. 5 a (`Name vergeben`, Exit 2); zweiter Start nach Stopp: Nr. 8 c (`Zustand vorhanden`); zwei Starts zugleich: Nr. 5 g (der Konflikt beim `docker run`, nichts Fremdes entfernt) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-24; der Planner zieht Treffer und Zähler beim Start nach). Für diesen Teil zählen: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, `offen`) — die Messungen im Container laufen unter einem festen Limit; Raten und Zeiten nennen Limit, Last und Messenden; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) — die Betriebsregeln stehen in §3, ein Auftreten wird als Beleg eingetragen. Keiner der offenen erreicht mit diesem Teil 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
