# Slice slice-container-image: Laufzeit-Image mit Datenvolumen, Start aus leerem Verzeichnis, Neustarttest

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-06`, `LH-RB-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-322` (Laufzeit-Image) · `SPEC-323` (Neustarttest, `harness/vertraege/`) · `SPEC-102` und `SPEC-211` (`harness/vertraege/`, Nachzug im liefernden Commit) · `SPEC-200` Nr. 14 e (Handbuch) · `LH-QA-06.a` (Container und Neustart) · `LH-RB-03.a` (Linux und CI, Container-Build, End-to-End-Test) · `SPEC-077` (`DATA_DIR`) · `SPEC-035` (OCI-Runtime-Image) · `SPEC-037` (Test, Integration) · `SPEC-123` (Start: Wiederherstellung, Integritätsprüfung) · `SPEC-059` (`SHUTDOWN_TIMEOUT`) · `LH-FA-02.c` (geordneter Shutdown) · `SPEC-124` Nr. 14 · `SPEC-005` (Admin-Endpunkte) · `SPEC-106` Nr. 5 und `SPEC-127` Nr. 8 (Bodies beim zweiten Import) · `SPEC-215` Nr. 7 b (`harness/vertraege/`, allein die Option von `docker stop`) · `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-container-image`, Worktree `/Development/odata-recorder-wt-image`; der Zweig entsteht von `main` nach dem Start-Commit, der Start liegt auf `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Laufzeit-Image des Dienstes (Stufe `image` des `Dockerfile`, `SPEC-322`) startet ohne externe Infrastruktur; `/data` ist sein Datenverzeichnis (`DATA_DIR`), und ein frisches, leeres Datenverzeichnis ist ein gültiger Startzustand (`LH-QA-06.a`). Der Neustarttest (`SPEC-323`) fährt gegen dieses Image, **ohne eingehängtes Volume**: frisches Datenverzeichnis des neuen Containers, Aufzeichnung gegen einen Referenzdienst, über den Admin-Listener (`SPEC-005`) Import, Workspace und Snapshot, Stopp und Start desselben Containers; Recordings, Workspaces und Snapshots sind danach lesbar (Closure-Kriterium der Welle, `LH-QA-06.a`). Er ist das Ziel `test-image` aus `SPEC-323` und ein **Gate** an `GATE_CHECKS`; er ist der Lauf über `make`, den `LH-QA-06` ohne CI als Ersatz anerkennt. Derselbe Lauf belegt den reproduzierbaren Container-Build: zwei Bauten desselben Standes ergeben ein byte-gleiches Binary (`LH-RB-03`, `LH-RB-03.a`, `SPEC-323` Nr. 3; Nachtrag des Planners vom 2026-11-27 aus der Bewertung des Architect, `2e289d4d`).

**Entkoppelt von `welle-dev-container`** (Entscheidung des Users, 2026-11-27, Antwort 1 der vier Fragen; Wortlaut in [welle-betrieb](../welle-betrieb.md) §4): Die Bedingung vom 2026-10-30 (Start erst nach `welle-dev-container`, Neustarttest nur im Daemon eines Arbeitscontainers, Verbot eines Laufs im Daemon des Hosts) ist **aufgehoben**. Der Neustarttest läuft im Docker-Daemon des Hosts, nicht abgeschottet; das ist der benannte Preis der Entscheidung. **Gate und kein Volume** (Entscheidung des Users, 2026-11-27, Antworten 1 und 2 der sechs Fragen in [welle-betrieb](../welle-betrieb.md) §4): Die frühere Bedingung des Planners (Werkzeug außerhalb von `GATE_CHECKS`) ist aufgehoben. Wie der Lauf neben den Gate-Läufen im selben Daemon besteht, legt `SPEC-323` fest: Namen und Label je Lauf, `--network none`, Speicherlimits, Aufräumen auf jedem Weg, keine Sperre im Treiber (die Rolle nimmt `flock` um `make`, Nr. 2 c dort).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Arbeitscontainer für Entwickler (`LH-RB-03`) — `welle-dev-container`; Gegenstand ist das Image des Dienstes, nicht die Umgebung, in der gebaut wird.
- Abschottung des Testlaufs durch einen eigenen Daemon — Bestand mit Grund: der User hat den Lauf im Daemon des Hosts entschieden (Antwort 1 der vier Fragen); kein Folge-Slice, und `welle-dev-container` bekommt daraus keine Sendung.
- Eine CI (Server, Pipeline, Runner) — der Lauf über `make` ist als Ersatz für den CI-Test der Messmethode von `LH-QA-06` anerkannt, solange es keine CI gibt (User, 2026-11-27, Antwort 4 der vier und Antwort 3 der sechs Fragen; Lastenheft 0.9.0).
- Ein eingehängtes Volume im Neustarttest, der Ersatz des Containers auf altem Bestand — Bestand mit Grund: User, Antwort 1 der sechs Fragen; die Prüfläufe laufen ohne Mount (`SPEC-322` *Ungeprüft*, `SPEC-323` Nr. 9 c).
- Erzwungenes Ende des Containers (`SIGKILL`) und der Modus `mock` im Image — benannte Negative des Vertrags (`SPEC-323` Nr. 9 a, 9 b).
- Orchestrierung des Betreibers (Compose-Dateien, Kubernetes), Registry, Veröffentlichung des Images — Out-of-Scope der Welle ([welle-betrieb](../welle-betrieb.md) §6): das Image und sein Start sind Gegenstand, nicht seine Umgebung.
- Healthcheck-Anweisung des Images — `GET /healthz` liegt am Admin-Listener (geliefert); keine Anforderung dieses Slice verlangt eine Anweisung im Image, und ein Healthcheck über den OData-Listener wäre ein Vertrag ohne Grundlage.
- TLS im Image (`LISTEN_TLS_*`) — `slice-listener-tls-verdrahtung` (Teil 2 des Schnitts von `slice-listener-tls`); das Image reicht Konfiguration durch, es setzt kein Zertifikat.

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

- [x] Image: die Stufe `image` des `Dockerfile` nach `SPEC-322` (leere Basis, Binary der Stufe `build`, Nutzer `65532`, `DATA_DIR=/data`, kein `VOLUME`, Eintrittspunkt mit `start`) und das Ziel `image` (`SPEC-323` Nr. 1); der Start aus einem leeren Datenverzeichnis gelingt (`LH-QA-06`, `LH-RB-03.a`).
- [x] Neustarttest nach `SPEC-323`: das Ziel `test-image` hängt an `GATE_CHECKS`, hängt kein Volume ein, nimmt im Modus `record` über beide Listener Recordings, einen Workspace und einen Snapshot auf, stoppt und startet denselben Container und liest alle drei mit dem Stand von vor dem Neustart (`LH-QA-06.a`); es räumt im Daemon des Hosts auf, was sein Lauf angelegt hat, und nur das (Nr. 6: Container, Tags, Verzeichnis; nicht den Bau-Cache, Nr. 6 f); der Lauf baut die Stufe `image` ein zweites Mal (`--no-cache-filter build` und `--no-cache-filter image`, Tag `odata-recorder:test-image-zweit-<L>`), prüft, dass das zweite Binary neu übersetzt ist, und vergleicht die beiden byte-gleich (`SPEC-323` Nr. 2 b und 3); die 20 Zeilen von §3 haben ihre Fänger.
- [x] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand): Bau des Image, Volume an `/data`, `LISTEN`, Stopp-Wartezeit und Rechte (`SPEC-323` Nr. 10, `SPEC-322`); `make handbuch-check` führt die Aufrufe von `docker` darin nicht aus (`SPEC-200` Nr. 14 e); das Kataloge-Gate ist grün.
- [x] `make gates` grün, darin der Neustarttest; §7 nennt seinen Lauf mit Commit, Ausgabe, Laufzeit (Messung nach `SPEC-211` Nr. 4), Last des Hosts und Messendem (Ersatz für den CI-Test, `LH-QA-06`). **Mit Abweichung** (Planner, Closure): gemessen ist die Laufzeit der ganzen Kette; die Summe der fünf Container-Tests nach `SPEC-211` Nr. 4 ist nicht gemessen, §3 *Zeitbudget* führt eine Rechnung, die Messung geht an die Closure von [welle-betrieb](../welle-betrieb.md).
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update: `harness/README.md` §Sensors (`test-image` als Gate, `image` als Werkzeug), `SPEC-102`, `SPEC-211` und der *Stand* von `SPEC-323` im liefernden Commit.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `Dockerfile`, `.dockerignore` | update | Stufe `image` nach `SPEC-322`; was der Bau-Kontext zusätzlich braucht |
| Fragment `image.mk` unter `harness/mk/` | neu | die Ziele `test-image` (an `GATE_CHECKS`) und `image` (`SPEC-323` Nr. 1) |
| `tools/harness/test-image.sh` | neu | Treiber: Aufruf, Namen (drei Image-Tags, darunter der zweite Bau), Bau, Prüfungen am Image (Angaben, Variablen, Inhalt über `export`) samt Alter und Vergleich der zwei Binaries (`-nt`, `cmp`), Ablauf (`stop -t`), Fristen, Aufräumen jedes Tags, geschlossener Leser, Exit (`SPEC-323` Nr. 2 bis 6) |
| `tests/e2e/image_test.go` | neu | die Client-Phasen `aufnehmen` und `pruefen` mit dem Referenzdienst, im zweiten Import die zwei gelesenen Exchanges (`SPEC-323` Nr. 4) |
| `harness/vertraege/` (`SPEC-215` Nr. 7 b), `docs/plan/planning/open/slice-dev-container-geruest.md` Zeile 44 | update (Architect) | allein `docker stop -t` statt der veralteten Form (§6, Review F-5) |
| `tools/gegenproben/` (neue Fälle mit Ersatz für `docker`: `testimage.go`, `testimage_test.go`; Anschluss in `main.go`, `faelle.go`, `strom_test.go`, `arbeitsbaeume_test.go`), `tools/harness/gegenproben-eingabe.sh` | neu, update | Fälle „GP“ der Tabelle (`SPEC-102`, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), Fälle 226 bis 236); der Stempel erst im Integrationsschritt |
| `harness/mk/gegenproben.mk` | update | Kopf und Hilfetext von `make gegenproben` nennen den Treiber des Neustarttests (Review F-4) |
| `harness/vertraege/README.md` | update | Zeilen `SPEC-211` und `SPEC-323` im Index (Nachzug mit dem *Stand*) |
| `harness/README.md` | update | `test-image` in der Tabelle der Gates, `image` in der Tabelle der Werkzeuge, *Nicht behauptet* |
| `harness/vertraege/` (`SPEC-102`, `SPEC-211`, `SPEC-323` *Stand*) | update | Nachzug im liefernden Commit (`SPEC-211` Nr. 2 b: fünf statt vier Container-Tests); der Vertrag selbst ist geliefert |
| Benutzerhandbuch | update | Kapitel zum Image (Pflicht Handbuch, `SPEC-323` Nr. 10) |
| `spec/spezifikation.md` | update (Architect) | `SPEC-322` Nr. 4 berichtigt (die Variable `PATH` des Bau-Dienstes) mit Zeile in der Historie, `b39db79d` (§6; bei der Closure nachgetragen, Verifikation V-2) |

Die Dateinamen sind der Vorschlag von `SPEC-323` (Kopf) und, für Test und Gegenproben, des Planners; wer sie ändert, zieht Tabelle und Block im selben Commit nach.

**Vertrag** (geliefert, Architect, 2026-11-27): `SPEC-322` (Image) und `SPEC-323` (Neustarttest) tragen die Antworten auf die Aufträge, die dieser Plan an den Architect gestellt hatte: Image und Datenverzeichnis (`SPEC-322` Nr. 1 bis 8), Lauf im Daemon des Hosts, Netz, Namen, Aufräumen und Abbruch (`SPEC-323` Nr. 1, 2, 5, 6), Prüfungen, Ablauf und Referenzdienst (Nr. 3, 4), Randformen (Nr. 7), Zusagen und Fänger (Nr. 8), Ungeprüftes (Nr. 9) und Größe (Nr. 10). Die Randformen entscheidet der Architect dort, nicht der Implementer (`AGENTS.md` §3.12).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; aus `SPEC-323` Nr. 8 übernommen, Architect; „GP“ ist ein Fall der Gegenprobe mit Ersatz für `docker` in einem Wegwerfverzeichnis, „Lauf“ das Ziel `test-image`, „Nr.“ in den Zellen meint `SPEC-323`; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Nutzer `65532:65532` | Lauf, Nr. 3 | `USER` entfernt | andere |
| 2 | Eintrittspunkt und Standardargument | Lauf, Nr. 3 und 4 a | `CMD` entfernt | andere |
| 3 | `DATA_DIR=/data`, außer `PATH` keine weitere Variable (also kein `LISTEN`, `ADMIN_LISTEN`, `MODE`), kein `VOLUME`, `EXPOSE`, `HEALTHCHECK` | Lauf, Nr. 3; GP: je Variable und je Angabe ein Fall | `VOLUME /data`; `EXPOSE`; `HEALTHCHECK`; eine weitere `ENV` | andere |
| 4 | `/data` mit Eigentümer und Rechten | Lauf, Nr. 3 | `--chown` entfernt | andere |
| 5 | Vertrauensanker und `/tmp`; kein weiterer Eintrag im Image | Lauf, Nr. 3; GP: je Ort ein zusätzlicher Name | `COPY` entfernt; eine weitere Datei im Image | andere |
| 6 | Version im Image | Lauf, Nr. 3 | `VERSION` nicht gereicht | andere |
| 7 | frisches Datenverzeichnis | Phase `aufnehmen` | `aufnehmen` zweimal gegen denselben Server | andere |
| 8 | geordneter Stopp mit Exit 0 | Lauf, Nr. 4 c; GP | `STOPSIGNAL SIGKILL` im Image; Exit nicht gewertet | andere |
| 9 | Recordings samt ihren Bodies überstehen den Neustart | Phase `pruefen` (Exchanges; neuer Import mit zwei gelesenen Exchanges) | neuer Container statt `start`; eine Antwort auf `/Products`, die der Import überspringt | andere |
| 10 | Workspace übersteht den Neustart | Phase `pruefen` | wie Nr. 9 | andere |
| 11 | Snapshot übersteht den Neustart | Phase `pruefen` (Liste, `restore`) | wie Nr. 9 | andere |
| 12 | Token übersteht den Neustart | Phase `pruefen` (Token von vorher) | Treiber reicht `pruefen` ein anderes Token | andere |
| 13 | Optionen der Aufrufe und verbotene Formen (Nr. 2 d, e) | GP: Liste der Aufrufe | je eine Option entfernt oder hinzugefügt | andere |
| 14 | eigene Namen und Label je Lauf | GP: zwei Läufe, getrennte Namen | feste Namen | andere |
| 15 | Aufräumen auf grün, rot und Baufehler, auch bei geschlossenem Leser der Ausgabe | GP je Weg und je Strom | `rmi` oder `rm` entfernt; SIGPIPE nicht ignoriert | andere |
| 16 | Aufräumen bei INT, HUP, TERM, Exit 2; ein weiteres Signal ändert nichts | GP je Signal, und je Signal als zweites während des Aufräumens | `trap` entfernt; das zweite Signal nicht ignoriert | andere |
| 17 | Exit-Codes (Nr. 6 b, c) | GP je Ursache | Aufräumfehler verschluckt | andere |
| 18 | Wertung der Phase (`SKIP`, `FAIL`, ohne `PASS`) | GP je Form | `SKIP` nicht gewertet | andere |
| 19 | Token in keiner Ausgabe und keinem Argument | GP: Ersatz liefert ein Zeichen-Muster als Token, Ausgabe und Aufrufliste ohne Treffer | Token als `--env` | andere |
| 20 | zwei Bauten ergeben ein byte-gleiches Binary, aus zwei Übersetzungen | Lauf, Nr. 3; GP: Ersatz liefert zwei verschiedene Dateien, dazu eine zweite Datei gleich alt und eine älter | Bauzeit über `-ldflags` in das Binary; Vergleich nicht gewertet; `--no-cache-filter image` entfernt; Alter nicht gewertet | andere |

**Nach dem Review nachgezogen** (Architect, 2026-11-28, `SPEC-323` Nr. 8): die Zeilen 3, 5, 9, 15 und 16 sagen mehr zu als am Stand der Übergabe; die Zahl der Zeilen bleibt 20. Ihre Fänger und rot gesehenen Mutationen stehen unter *Fix-Runde zum Review* darunter; die Tabelle *Mutationen je Zeile* berichtet den Stand der Übergabe.

**Mutationen je Zeile** (`AGENTS.md` §3.10; Implementer, 2026-11-27 und 2026-11-28, Host mit 20 Kernen und fremder Last, Messender: der Implementer; jede Zeile ist Klasse `andere`, je Mutation ein Lauf). *Lauf* ist der Treiber gegen den Daemon des Hosts in einer Wegwerfkopie; *GP* ist ein Mutant des Treibers gegen die Fälle 226 bis 236 (`TestTestImageFaelleLaufenGegenDenTreiber` in einem Container ohne Netz, 49 Mutanten, 49 rot am Stand der Übergabe; die elf Fälle gegen den echten Treiber in 30 von 30 Wiederholungen grün, Last 3,6 bis 7,0, am Stand vor `--no-cache-filter image` und vor der Zerlegung der Fälle für den Lint). Die Läufe zu Nr. 1 bis 12 liegen vor dem Nachzug für den Lint (Werte der Phase `aufnehmen` über `t.Log`); `make test-image` ist danach grün gesehen.

| Nr. | Mutation | Rot gesehen |
|---|---|---|
| 1 | `USER` entfernt | Lauf: `rot: Nutzer`; GP: Prüfung entfernt, Fall `image` |
| 2 | `CMD` entfernt | Lauf: `rot: Standardargument`; GP: Prüfung von `cmd` oder `entrypoint` entfernt, Fall `image` |
| 3 | `VOLUME /data` | Lauf: `rot: Volume`; GP: je Prüfung (Volume, `LISTEN`, `ADMIN_LISTEN`, `MODE`, `DATA_DIR`) entfernt, Fall `image` |
| 4 | `--chown` entfernt | Lauf: `rot: /data Eigentümer und Rechte`; GP: Prüfung entfernt, Fall `image` |
| 5 | `COPY` der Anker, `COPY` von `/tmp` entfernt | Lauf: `rot: docker cp /etc/ssl/certs/ca-certificates.crt`, `rot: docker cp /tmp`; GP: Größe der Anker oder Rechte von `/tmp` nicht geprüft, Fall `image` |
| 6 | Treiber reicht `VERSION` nicht an den Bau, `VERSION=9.9.9` | Lauf: `rot: Version`; GP: Vergleich entfernt, Fall `image` |
| 7 | Phase `aufnehmen` zweimal | Lauf: `kein frisches Datenverzeichnis` (Session `complete`) |
| 8 | `STOPSIGNAL SIGKILL`; Exit nicht gewertet | Lauf: `rot: Stopp: Exit des Servers`; GP: Fälle `stopp`, `aufraeumen` |
| 9 | neuer Container statt `start`, mit dessen Token | Lauf: `Sessions nach dem Neustart` (allein eine neue `open`) |
| 10, 11 | hinter dem gemerkten Stand ein weiteres `reset` und ein weiterer Snapshot (Mutation der Phase `aufnehmen`; unter der Mutation von Nr. 9 hält die Phase `pruefen` schon an den Sessions an) | Lauf: `Workspaces nach dem Neustart`, `Snapshots nach dem Neustart` |
| 12 | Treiber reicht `pruefen` ein anderes Token | Lauf: `GET /__admin/sessions: 401 ODR-E2101` |
| 13 | `--memory-swap` entfernt; `-p`, `--user`, `--env LISTEN`, ein Argument am Server hinzugefügt; `--network` entfernt, `host` | GP: Fall `gut` (Aufrufe Argument für Argument) |
| 14 | feste Kennung; Label eines Containers entfernt | GP: Fälle `zwei-laeufe`, `gut` |
| 15 | `rm`, `rmi`, Entfernen des Verzeichnisses entfernt; Baufehler ohne Aufräumen | GP: Fälle `gut`, `image`, `exit-codes` und weitere |
| 16 | `trap` entfernt, ohne `HUP`, `kill` des laufenden Aufrufs entfernt | GP: Fall `abbruch` |
| 17 | Aufräumfehler verschluckt, Baufehler 1, Daemon-Frage 1, roter Lauf 2, Log entfernt, Argument-Prüfung entfernt | GP: Fälle `aufraeumen`, `exit-codes`, `phasen`, `stopp`, `aufruf` |
| 18 | Exit, `SKIP`, `FAIL`, `PASS` je nicht gewertet | GP: Fall `phasen` |
| 19 | Token als `--env`, auf stdout, nicht auf der Standardeingabe | GP: Fälle `gut`, `token` |
| 20 | Zeitstempel über `-ldflags -X` in eine verwendete Variable des Binary; Vergleich nicht gewertet; `--no-cache-filter build` oder `--no-cache-filter image` entfernt; Alter nicht gewertet, gleich alt erlaubt, jedes andere Alter erlaubt, Seiten vertauscht | Lauf: `rot: Binary der zwei Bauten ungleich` (Zeitstempel im Binary), `rot: zweites Binary nicht neu übersetzt` (ohne `--no-cache-filter image`, Binary ohne Zeitstempel); GP: Fälle `image` (Vergleich; Alter: `binary-gleich-alt`, `binary-aelter`), `gut` (die beiden Optionen), alle elf (Seiten vertauscht) |

Ohne Zeile der Tabelle sind vier weitere Mutanten rot: 31 Versuche, Pause entfernt, Frage nach dem Server entfernt (Fall `token`, `SPEC-323` Nr. 4 a) und die Werte nicht an `pruefen` gereicht (Fall `gut`). Zur Zeile 20: `-ldflags -buildid=<Zeit>` ändert die Bytes des Binary nicht und ist ersetzt; dass `--no-cache-filter image` nötig ist, fängt auch bei reproduzierbarem Bau die Prüfung der neuen Übersetzung (der Befund in §6).

**Fix-Runde zum Review** (`AGENTS.md` §3.10, §3.17; Implementer, 2026-11-28, Host mit 20 Kernen und fremder Last, Messender: der Implementer; jede Zeile Klasse `andere`). *GP* ist ein Mutant des Treibers gegen die elf Fälle (`TestTestImageFaelleLaufenGegenDenTreiber`) in einem Container ohne Netz mit `--memory=4g --cpus=4`, `sh` ist `dash`, je Mutant ein Lauf: 34 Mutanten, 34 rot aus dem genannten Grund, der echte Treiber davor und danach grün, Load Average 2,3 bis 5,4. Raten (je Lauf ein Prozess, Load Average 8,4 bis 4,7): die elf Fälle gegen den echten Treiber 0 von 10 rot; das zweite Signal (`trap ''` entfernt, ohne `INT`, ohne `HUP`, ohne `TERM`) und SIGPIPE nicht ignoriert je 5 von 5 rot. *Lauf* ist `make test-image` in einer Wegwerfkopie gegen den Daemon des Hosts, je ein Lauf, Load Average 5,8 bis 8,3; unverändert grün (54 s); nach keinem Lauf blieb ein Container oder Tag.

| Nr. | Menge (Elemente) | Mutation | Rot gesehen |
|---|---|---|---|
| 3 | Angaben: `EXPOSE` (`{}` erlaubt), `HEALTHCHECK` (drei Formen); 18 fremde Variablen je vor und hinter den erlaubten, `PATH=x` als Wert; zwei erlaubte Formen dazu | GP: Prüfung von `EXPOSE` oder `HEALTHCHECK` entfernt, `{}` beim `HEALTHCHECK` erlaubt oder bei den Ports verboten; Variablenprüfung entfernt, allein die drei alten Namen, Anfang `PATH` oder Ende `DATA_DIR` erlaubt, Teil eines erlaubten Namens erlaubt, Groß und klein gleich, Name bis zum letzten `=`, `PATH` verboten; falsches Feld im Format. Lauf: `EXPOSE`, `HEALTHCHECK`, `ENV PATHX=1` | Fall `image` (`expose`, `healthcheck`, `healthcheck-leer`, `erlaubt-2`, `variable-vorn-UPSTREAM_URL`, `-PATHX`, `-XDATA_DIR`, `-DATA_DI`, `-data_dir`, `variable-wert`), Fall `gut` (Format); Lauf: `rot: EXPOSE`, `rot: HEALTHCHECK`, `rot: Variable PATHX` |
| 5 | 46 fremde Namen: sieben Orte, je erlaubtem Namen (18) Anfang und Ende, je Name mit Punkt (3) literal; dazu einer am Anfang | GP: Prüfung entfernt, `grep` ohne `-x` oder ohne `-F`, ein Name mehr oder einer weniger erlaubt, allein die ersten 18 Zeilen, erste Zeile ausgelassen, Export eines anderen Containers, Ausgabe ohne Namen, Fehler von `export` nicht gewertet. Lauf: `/bin/sh` im Image, `data/fremd` im Image | Fall `image` (`inhalt-0`, `-1`, `-5`, `-21`, `-vorn`, Ausgabe), `gut`, `exit-codes`; Lauf: `rot: Inhalt` mit `bin/`, `bin/sh` bzw. `data/fremd` |
| 9 | Report des zweiten Imports | Lauf: Referenzdienst liefert `/Products` als `application/json;odata.metadata=none` (der Import überspringt die Antwort) | `rot: Phase pruefen`, `image_test.go:297`: `read` 1, `skipped` `ohne_kontext`, Status 201 |
| 15 | drei Wege (grün, rot am Image, Baufehler) je Standardausgabe und Fehlerstrom auf geschlossener Pipe (6) | GP: `trap '' PIPE` entfernt, erst im Aufräumen, erst hinter den Bauten, `trap 'exit 2' PIPE` | Fall `aufraeumen` (`gruen-pipe-1`: Exit -1 bzw. 2, `TMPDIR` nicht leer) |
| 16 | drei zweite Signale | GP: `trap ''` im Abbruch entfernt, ohne `INT`, ohne `HUP`, ohne `TERM`, `trap -` statt `trap ''` | Fall `abbruch` (`zweites-INT`, `-HUP`, `-TERM`: zwei Meldungen `abgebrochen`; bei `trap -` bleibt `TMPDIR`) |
| 13 (F-5) | `docker stop` | GP: `--time 30`, `-t 10` | Fall `gut`, Aufruf 18 |

Akzeptierte Negative der Fix-Runde, je mit Grund und Folge: (a) die fremden Namen und Variablen sind eine unbeschränkte Menge, die Prüfgrenze sind die oben genannten Bildungen; Namen mit Leerzeichen, Steuerzeichen oder Zeichen jenseits von ASCII haben keinen Fall (der Ersatz zerlegt seine Liste an Leerzeichen). GNU `tar` 1.35 schreibt unter `LC_ALL=C` Zeilenumbruch und Umlaut als Escape-Folge (gemessen am Host), ein solcher Name steht also nicht in der Menge; Folge: ein anderer `tar` am Host, der das nicht tut, fiele erst an einem solchen Namen auf. (b) Ein Tar-Strom von `export`, den `tar -t` nicht liest (`Tar-Strom Inhalt`), hat keinen Fall; Grund: der Ersatz liefert allein lesbare Ströme, wie bei den Strömen von `cp`; Folge: 0 von 1 Form gefangen.

**Prüfung der neuen Übersetzung** (`SPEC-323` Nr. 3, Prüfung 1; 2026-11-28, 20 Kerne, Messender: der Implementer). Das Alter hat drei Formen: jünger ist grün (jeder grüne Fall, das zweite Binary des Ersatzes eine Sekunde jünger), gleich alt und eine Sekunde älter sind rot (Fall `image`, 29 Abweichungen). GP in einem Container ohne Netz mit `--memory=4g --cpus=4`, `sh` ist `dash`, je Mutant ein Lauf, Load Average 5,4 bis 6,9: die elf Fälle gegen den echten Treiber grün, auch in 10 von 10 Wiederholungen; sieben Mutanten, sieben rot: Prüfung entfernt und gleich alt erlaubt (`binary-gleich-alt: Exit 0 statt 1`), jedes andere Alter erlaubt (`binary-aelter: Exit 0 statt 1`), Seiten vertauscht (alle elf Fälle), Vergleich entfernt (`binary-ungleich`), je eine der zwei Optionen entfernt (`gut`, Aufruf 9). Lauf gegen den Daemon des Hosts (`make test-image` in einer Wegwerfkopie, Treiber allein ohne `--no-cache-filter image`, Binary ohne Zeitstempel): 3 von 3 Läufen `rot: zweites Binary nicht neu übersetzt`, Load Average 3,9 bis 6,4, danach weder Container noch Tag eines Laufs.

**Messung** (`SPEC-211` Nr. 4; 2026-11-27, 20 Kerne, Wanduhr mit `date`, Messender: der Implementer): `make test-image` mit Quelle und Stufe `build` im Cache 21,4 s, 19,9 s und 19,9 s (Load Average 3,6 bis 4,4); der erste Lauf des Standes, mit neu gebauter Stufe `integration`, 51 s (Load Average 3,6 bis 4,9). Diese drei Läufe liegen vor `--no-cache-filter image`; die Option baut die vier `COPY` der Stufe `image` neu, ein Lauf mit ihr in einer Wegwerfkopie dauerte 25 s (Load Average 1,4 bis 4,7). `make gates` am Stand `3e72b1fb` (mit `main` ab `84b66635`) war grün, darin `test-image` mit allen fünf Zeilen bis `test-image: grün`, die ganze Kette 1147 s (2026-11-28, Load Average 5,9 am Beginn und 6,8 am Ende); die Zeit des Neustarttests darin ist nicht einzeln gemessen. Nicht gemessen: ein Lauf ohne die Basis-Images und `make -j gates`.

**Messung, Stand nach der Fix-Runde** (eingetragen vom Planner bei der Closure aus den Ausgaben der Rollen; Host mit 20 Kernen, Docker 29.10.0; der Absatz darüber ist der Stand vor der Fix-Runde). `make gates` am Stand `6453c3e4`: Exit 0, 1170 s, Load Average 9,10 am Beginn (2026-11-28T10:47:48+01:00) und 8,37 am Ende (Messender: der Implementer der Fix-Runde; Ausgabe in §7). `make test-image` einzeln: in der Fix-Runde 54 s in einer Wegwerfkopie (Load Average 5,8 bis 8,3, Implementer; warum länger als die übrigen Läufe, ist nicht gemessen), beim Verifier der Treiber in einer Kopie aus `git archive HEAD` 23,2 s, beim Reviewer vor der Fix-Runde 24,0 s (Load Average 4,8 vorher, 6,7 nachher).

**Zeitbudget der Container-Tests** (`SPEC-211` Nr. 4; Verifikation V-1; **eine Rechnung, keine Messung**). Die Messvorschrift verlangt die Summe der fünf Ziele `test-integration`, `test-determinism`, `test-contract`, `test-security` und `test-image`, einzeln nacheinander; diese Messung ist nicht gefahren, auch nicht vom Implementer, dem Nr. 4 sie beim Hinzufügen eines Ziels zuweist. Gerechnet aus Werten verschiedener Tage und Stände, Zuwachs von `make gates` nach geändertem Quelltext: vier Ziele rund 67 s (`SPEC-211` Nr. 4) plus `test-image` 20 bis 25 s bei Quelle und Stufe `build` im Cache ergibt rund 87 bis 92 s. Mit den höheren Werten des neuen Ziels (51 s, darin der Bau der Stufe `integration`, den die 67 s schon enthalten; 54 s, Ursache nicht gemessen) ergibt dieselbe Addition 118 bis 121 s. **Trigger** (`SPEC-211` Nr. 4): liegt der Zuwachs nach geändertem Quelltext bei **gültiger** Messung über 120 s, legt der Planner dem User drei Wege vor: Beschleunigung, Anhebung der Zahl mit neuer Messung, Herausnahme eines Ziels mit Pflichtlauf und Träger (eine Lockerung, ADR). Er ist mit dieser Rechnung weder als eingetreten festgestellt noch ausgeschlossen. Adresse der Messung: Planner, Closure von [welle-betrieb](../welle-betrieb.md) (dort §2).

**Akzeptierte Negative aus der Verifikation** (`AGENTS.md` §3.15; eingetragen vom Planner bei der Closure; die Mutationen hat der Verifier je einmal gefahren, der Text ist nicht nachgeprüft):

- **V-3, Stelle des zweiten Signals** (`SPEC-323` Nr. 6 d): der Fall `abbruch` sendet das zweite Signal allein beim ersten Entfernen des Aufräumens. *Grund:* der Ersatz für `docker` löst es am ersten `rm` aus; kein Fall führt die Stelle im Durchgang als Achse. *Folge:* eine Änderung, die das Ignorieren erst hinter dieser Stelle aufhebt, bleibt grün; ein weiteres Signal an einer späteren Stelle begänne dann das Aufräumen neu oder beendete es und ließe Tags und Verzeichnis eines Laufs liegen. *Maß:* 1 von 10 Schritten des Durchgangs geführt (6 `rm`, 3 `rmi`, das Verzeichnis); gesehen grün: `trap abbruch` vor der Schleife der `rmi` wieder scharf. Am Code gilt die Zusage: `trap ''` steht vor dem Aufräumen und wird nicht zurückgesetzt (gelesen).
- **V-4, Abbruch bei geschlossenem Leser** (`SPEC-323` Nr. 6 e „für jede Operation“, Nr. 8 Zeile 15): der Abbruch durch ein Signal hat bei geschlossener Standardausgabe oder geschlossenem Fehlerstrom keinen Fall. *Grund:* die Fälle mit Pipe führen die Wege grün, rot am Image und Baufehler, der Fall `abbruch` läuft mit offenen Strömen. *Folge:* eine Änderung, die `SIGPIPE` im Abbruch wieder wirken lässt, bleibt grün; ein Abbruch einer Ausgabe in eine geschlossene Pipe endete dann vor dem Aufräumen und ließe Container und Tags liegen. *Maß:* 6 von 8 Zellen Weg mal Strom geführt, Abbruch 0 von 2; gesehen grün: `trap - PIPE` im Abbruch.

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Der Neustarttest läuft im Daemon des Hosts; die Sperre nimmt die Rolle um `make`, nicht der Treiber (`SPEC-323` Nr. 2 c). Lauf und Mutationen entfernen dort nur Objekte des eigenen Laufs (Nr. 6).
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Variable, jede verbotene Form, jeder Weg des Aufräumens, jedes Signal, jede Form der Phasen-Wertung), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-handbuch-meldungskatalog` liegen in `done/` (erfüllt). Der Vertrag samt Randformen liegt vor dem ersten Code-Commit vor (`AGENTS.md` §3.12; erfüllt: `SPEC-322` und `SPEC-323` auf `main`, `082c560c`). **Gate-Kette** (Übergangsregeln Nr. 7; er ändert `Dockerfile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/` und den Stempel): kein anderer Kettenslice liegt in `in-progress/`; Vorgänger der Kette ist [slice-bootstrap-m6](../done/slice-bootstrap-m6.md) in `done/` (Planner, 2026-11-27, Drift-Log der Roadmap). `welle-dev-container` ist keine Bedingung mehr (§1).

**Stand der Bedingungen** (Planner, 2026-11-27, bei der Closure von Teil D): der Vorgänger [slice-bootstrap-m6](../done/slice-bootstrap-m6.md) liegt in `done/`, und in `in-progress/` liegt kein anderer Kettenslice ([slice-abnahme-belege](../done/slice-abnahme-belege.md) und [slice-lasttest-leser-zeitgrenze](../done/slice-lasttest-leser-zeitgrenze.md) sind keine); **erfüllt, auf `main` nach dem Merge des Zweigs `slice-bootstrap-m6`**. Er ist der nächste Slice der Gate-Kette. Kein Pfad des Blocks steht im Block eines laufenden Slice; das Handbuch (nicht im Block) schreibt auch `slice-abnahme-belege`. `basis`, `schaetzung` und `marke` setzt der Planner beim Start.

**Stand beim Start** (Planner, 2026-11-27, `main` ab `34c574b5`): Jede Bedingung ist erfüllt. (1) [slice-bootstrap-m2](../done/welle-record-replay/slice-bootstrap-m2.md) und [slice-handbuch-meldungskatalog](../done/slice-handbuch-meldungskatalog.md) liegen in `done/`, ebenso die Slices der dritten Abhängigkeit unten. (2) `SPEC-322` und `SPEC-323` stehen auf `main` (`082c560c`, Zeile 20 aus `2e289d4d`); die 20 Zeilen der Tabelle und das Raster aus §3 und §6 sind byte-gleich `SPEC-323` Nr. 8 und 7 und tragen keinen Platzhalter; offene Fragen an den Architect vor dem ersten Code-Commit gibt es nicht. (3) In `in-progress/` liegt daneben kein Kettenslice: [slice-abnahme-belege](../done/slice-abnahme-belege.md) und [slice-admin-abweisen-begrenzen](../done/slice-admin-abweisen-begrenzen.md) ändern keinen Pfad der Nr. 7 der Übergangsregeln. **Zugleich laufende Zweige** (beim Start gelesen: `git diff --name-only main...slice-abnahme-belege`; der Zweig `slice-admin-abweisen-begrenzen` hat noch keinen eigenen Commit, gelesen sind die offenen Änderungen seines Arbeitsbaums und der Block seines Plans): keine Datei des Blocks gemeinsam, `harness/README.md` ändert allein dieser Slice. Gemeinsam ist allein das Benutzerhandbuch (Zeile `Stand:` im Kopf; die Kapitel sind verschieden: Image hier, *Schutz der Daten* und *Admin-Schnittstelle* dort). Unter `harness/vertraege/` ändert [slice-abnahme-belege](../done/slice-abnahme-belege.md) `SPEC-042`, dieser Slice `SPEC-102`, `SPEC-211`, `SPEC-323` und den Index.

**Abhängigkeiten — Grund je Eintrag:**

- `slice-bootstrap-m2` (`welle-record-replay`): der Prozess, den das Image startet: Startbefehl mit `LadeOderMelde`, Verdrahtung, Listener im Modus `record`, geordnetes Herunterfahren und Startwartung. Der Container startet mit `start`; `CMD ["start"]` ist Sache dieses Slice (Sendung von `slice-cli-adapter`, `SPEC-138`; `SPEC-322` Nr. 6).
- `slice-handbuch-meldungskatalog` (`welle-benutzerhandbuch`, setzt `slice-handbuch-geruest` voraus): die DoD verlangt ein Handbuch-Kapitel und ein grünes Kataloge-Gate.
- `welle-mock`, `slice-admin-imports`, `slice-admin-workspaces`, `slice-bootstrap-m6-listener` (alle in `done/`, erfüllt): Workspaces und Snapshots entstehen und werden gelesen über den Admin-Listener im Binary (`LH-QA-06.a`).

`slice-listener-tls`, die Telemetrie, `slice-audit-log` und `slice-lasttest` sind nicht Bedingung; die Telemetrie ändert das Image nicht. `slice-bootstrap-m6` (Teil D) liefert dem Test keinen Code; er liegt allein in der Reihenfolge der Gate-Kette davor.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect: `produktcode` 220 und `fakes` 1 aus `SPEC-323` Nr. 10, 20 Zeilen der Klasse `andere` aus §3; Pfade vom Planner nach §3; **`basis`, `schaetzung` und `marke` hat der Planner beim Start vor dem ersten Code-Commit gesetzt**, 2026-11-27, `AGENTS.md` §3.14): 20 Zeilen, *a* = 0, *b* = 20, *f* = 1, *p* = 220: *S* = 2020, **Marke 2800**, keine Schnittfrage (203 unter der Schnittgrenze; vor dem Start mit `make groessenmarke` an dieser Datei bestätigt, `Marke steht aus: Schätzung 2020, Marke 2800`). `basis` ist der reine Move nach `in-progress/`. Kein Teil ist vorab als Schnitt beim Handoff benannt: Treiber, Client-Phasen und Gegenproben tragen denselben Vertrag, und das Handbuch zählt die Messung nicht. Die Schätzung des Architect für den Umfang ohne Handbuch ist rund 800 Zeilen (`SPEC-323` Nr. 10, nicht gemessen). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; jede weitere Datei, die der Implementer anlegt oder ändert, nennt er im selben Commit im Block.

```groessenmarke
pfad: Dockerfile
pfad: .dockerignore
pfad: harness/mk/image.mk
pfad: tools/harness/test-image.sh
pfad: tests/e2e/image_test.go
pfad: tools/gegenproben/testimage.go
pfad: tools/gegenproben/testimage_test.go
pfad: tools/gegenproben/main.go
pfad: tools/gegenproben/faelle.go
pfad: tools/gegenproben/strom_test.go
pfad: tools/gegenproben/arbeitsbaeume_test.go
pfad: tools/harness/gegenproben-eingabe.sh
pfad: harness/mk/gegenproben.mk
pfad: harness/README.md
produktcode: 220
fakes: 1
basis: 0158092b
schaetzung: 2020
marke: 2800
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nur vor dem ersten Code-Commit, wenn sich die Tabelle ändert und *S* die Schnittgrenze von `SPEC-128` Nr. 1 erreicht (`AGENTS.md` §3.14); die Naht ist dann Image (Bau) gegen Neustarttest (Treiber und Gate). Eine Messung über der Marke nach dem Code führt nicht zurück, sie steht mit Grund als `ueberschreitung` im Block.
- `in-progress` → `open` (blockiert): der Neustarttest lässt sich nicht fahren, ohne eine harte Regel zu lockern (`AGENTS.md` §3.1, §3.6), oder er kann im Daemon des Hosts nicht auf eigene Objekte begrenzt aufräumen.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Der Neustarttest verlängert die Gate-Kette um einen daemonabhängigen Lauf, dessen Frist unter Last reißen kann — **Ausgang:** weiter offen; Register `BEO-REPO/zeitbudget-der-kette-ohne-messung-nach-neuem-ziel`; Adresse: Planner, Closure von [welle-betrieb](../welle-betrieb.md) (Messung der fünf Container-Tests nach `SPEC-211` Nr. 4; §3 *Zeitbudget* führt bis dahin eine Rechnung samt Trigger). Eine gerissene Frist ist an den Läufen dieses Slice nicht berichtet; der zweite Bau nach `SPEC-323` Nr. 3 übersetzt das Programm einmal mehr. Die Aufnahme in die Kette ist entschieden (User, Antwort 2 der sechs Fragen; `SPEC-323` Nr. 1: Zeit geschätzt 10 bis 20 s bei vorhandenen Images, nicht gemessen).
- Ein Werkzeug außerhalb der Kette fährt niemand (`BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger`) — **Ausgang:** entfallen: der Neustarttest ist ein Gate (`SPEC-323` Nr. 1).
- Im Daemon des Hosts kollidiert der Lauf mit Gate-Läufen oder fremden Objekten, oder das Aufräumen löscht Fremdes oder lässt Eigenes stehen — **Ausgang:** entfallen: entschieden (Architect, `SPEC-323` Nr. 2 b, 5, 6; Fänger Zeilen 14 bis 17 in §3). Die Hälfte „lässt Eigenes stehen“ trat im Review auf (F-3, geschlossener Leser) und ist im Slice behoben (Zeile *Review F-3* unten); nach den Läufen von Review und Verifikation blieb kein Container und kein Tag eines Laufs. Benannt bleibt im Vertrag: SIGKILL am Treiber lässt Container, Tags und Verzeichnis liegen (Nr. 6 d dort).
- `AGENTS.md` §3.1 verlangt netzlose Prüfläufe ohne schreibbaren Mount — **Ausgang:** entfallen: entschieden (Architect, `SPEC-323` Nr. 2 d und e; User, Antwort 1 der sechs Fragen): `--network none` und kein Mount, ohne Ausnahme von der Regel.
- Zusage-Lücke: `LH-QA-06.a` nennt Recordings, Workspaces und Snapshots, der Plan prüfte nur Recordings, weil es keinen Weg über die Prozessgrenze gab — **Ausgang:** entfallen: entschieden (Planner, 2026-11-27): der Admin-Listener ist im Binary geliefert, der Neustarttest dieses Slice prüft alle drei (§1, DoD 2); das akzeptierte Negativ und die Adresse „Planner für `slice-bootstrap-m6`“ entfallen. [slice-bootstrap-m6](../done/slice-bootstrap-m6.md) hatte die Sendung nie angenommen und nennt den Neustarttest im Container als ausgeschlossen (§1 dort); seine zwei Starts im Prozess ersetzen ihn nicht.
- Der Umfang ist größer als beim ersten Entwurf (drei Datenarten, Admin-Listener im Container, Gegenproben), und ein Treiber mit echten Prozessen, Ersatz für `docker` und Signal-Fällen liegt nach dem Register über der Schätzung (`BEO-REPO/schaetzformel-unterschaetzt-werkzeug-mit-prozessen`) — **Ausgang:** entfallen: das Ende liegt bei 2026 Zeilen auf *S* = 2020 und unter der Marke 2800 (bei der Closure mit `make groessenmarke` gemessen, §7); dieser Vorgang ist kein dritter Beleg des Eintrags. Gemessen am Handoff (Implementer, 2026-11-28, mit der Fix-Runde zu Zeile 20): 1718 Zeilen über die Pfade des Blocks; mit der Fix-Runde zum Review (dazu `harness/mk/gegenproben.mk`) 2026 Zeilen, unter der Marke 2800, kein Eintrag `ueberschreitung`.
- Das Benutzerhandbuch schreiben zugleich [slice-abnahme-belege](../done/slice-abnahme-belege.md) und [slice-admin-abweisen-begrenzen](../done/slice-admin-abweisen-begrenzen.md) (andere Kapitel, gemeinsam die Zeile `Stand:`) — **Ausgang:** entfallen: `main` mit beiden ist vor dem Handoff aufgenommen (Implementer, 2026-11-28), das Handbuch ohne Konflikt.
- `SPEC-323` Nr. 4 d nannte in einem Satz das Token von vor dem Neustart und `--env` — **Ausgang:** entfallen: entschieden (Architect, 2026-11-28): Nr. 4 d sagt jetzt, dass das Token über die Standardeingabe geht und `--env` allein die Werte aus der Phase `aufnehmen` trägt. Folge: keine Fix-Runde, der Treiber steht so.
- Das Volume gehört einem anderen Nutzer als der Prozess im Image — **Ausgang:** entfallen: entschieden (Architect, `SPEC-322` Nr. 5). Ohne Fänger am Image bleiben als benanntes Negativ der Eigentümer eines neuen Volumes und ein nicht beschreibbares Verzeichnis (`SPEC-322` *Ungeprüft*; User, Antwort 1 der sechs Fragen); das Review hat beide von Hand gesehen (Eigentümer `65532`, Ausgangswert 1).
- `SIGKILL` während der Aufzeichnung hinterlässt eine `open`-Session — **Ausgang:** entfallen: entschieden (Architect, `SPEC-323` Nr. 9 a): benanntes Negativ, die Wiederherstellung ist am Dienst geprüft (`SPEC-123` Nr. 5).
- Randformen (`AGENTS.md` §3.12, §3.16): eine Zelle ohne Ort geht an den Architect — **Ausgang:** entfallen: entschieden (Architect, `SPEC-323` Nr. 7, Raster unten). Eine Randform ohne Ort fand erst das Review (F-3, `SIGPIPE`); sie ist im Slice entschieden und im Register geführt (§7).
- Ob `docker cp` den Eigentümer so ausgibt, wie `SPEC-323` Nr. 3 liest (Nr. 9 f dort) — **Ausgang:** entfallen: gemessen (2026-11-27, Implementer, Docker 29.10.0): der Eintrag trägt `65532/65532` und `drwx------`; Nr. 9 f nennt die Messung.
- Docker setzt dem Image zusätzlich die Variable `PATH` (gemessen 2026-11-27 mit `image inspect`), `SPEC-322` Nr. 4 sagte „Das Image setzt allein `DATA_DIR=/data`“ — **Ausgang:** entfallen: entschieden (Architect, 2026-11-28): der Wortlaut ist berichtigt (das `Dockerfile` setzt allein `DATA_DIR`, der Bau-Dienst fügt `PATH` hinzu), das Image bleibt; `ENV PATH=` entfernte die Variable nicht, sondern setzte sie leer, und die Konfiguration kennt sie nicht. Die Prüfung nach `SPEC-323` Nr. 3 erlaubt weitere Variablen. Folge: keine Fix-Runde. (Stand dieser Entscheidung; die Fix-Runde zu F-1 erlaubt seither außer `DATA_DIR` allein `PATH`.)
- `SPEC-323` Nr. 3 nannte für den zweiten Bau allein `--no-cache-filter build`; damit übersetzt die Stufe `build` neu, das zweite Image trägt aber das Binary eines früheren Baus, der Vergleich nach Zeile 20 läse zweimal dasselbe (gemessen 2026-11-28, Implementer, Docker 29.10.0: mit Zeitstempel im Binary druckt der Schritt `go build` im zweiten Bau eine neue Prüfsumme, im Image liegt die alte; `COPY --link` und `--provenance=false` ändern das nicht, Messung ohne den Detektor je ein Lauf). Der Treiber gibt deshalb zusätzlich `--no-cache-filter image` — **Ausgang:** entfallen: entschieden (Architect, 2026-11-28, nach eigener Messung, `SPEC-323` Nr. 3): beide Optionen bleiben, der Vergleich liest die Binaries zweier Images, und der Treiber prüft zusätzlich, dass das zweite Binary jünger ist als das erste. Reichweite des Belegs zu `LH-RB-03`: `SPEC-323` Nr. 9 g (gleicher Host, Daemon, Quellstand, Basis-Images; nicht über Hosts, nicht aus leerem Cache, eine Stichprobe je Lauf). Folge, **Fix-Runde** (Implementer), geliefert (2026-11-28): der Treiber prüft bei `cmp` mit `-nt` (rot mit `zweites Binary nicht neu übersetzt`), der Ersatz setzt beiden Dateien feste Zeitstempel und der Fall `testimage/image` führt die Formen gleich alt und älter (29 Abweichungen, auch in `SPEC-102`), *Stand* von `SPEC-323` und Kopf des Treibers sind nachgezogen; Mutationen in §3 unter Zeile 20. Offen allein der Stempel der Gegenproben im Integrationsschritt. **Reviewer:** hält Treiber und Fälle gegen `SPEC-323` Nr. 3 und Zeile 20.
- Vier Formen des Treibers nannte `SPEC-323` beim Handoff nicht ausdrücklich: (1) scheitert ein Aufruf von `docker` außer dem Bau und den Phasen (`image inspect`, `create`, `cp`, `run … version`, `start`, `stop`, `inspect`), ist der Exit 1, bei nicht erreichbarem Daemon (Frage mit `docker version`) 2; Nr. 6 c nennt `create` und `start` nicht. (2) Die Container der Prüfungen am Image tragen Namen, Label und `--network none`, der von `version` die Grenzen des Servers; Nr. 5 nennt Grenzen allein für Server und Client. (3) Ein Signal während des Aufräumens beendet kein laufendes Entfernen, das Aufräumen läuft danach von vorn, Exit 2 (gelesen aus Nr. 6 d; die Zelle *Aufräumen × Wiederholung* nennt je Objekt einmal, wirksam entfernt wird je Objekt einmal). (4) Der Treiber fragt vor jedem `rmi` und `rm`, das scheitert, ob das Objekt besteht; eine gescheiterte Frage zählt als bestehend — **Ausgang:** entfallen: entschieden (Architect, 2026-11-28): alle vier bestätigt und im Vertrag (`SPEC-323` Nr. 2 d und 5 für (2), Nr. 6 b für (4), Nr. 6 c für (1), Nr. 6 d und die Zelle *Aufräumen × Wiederholung* für (3)); Fänger sind die Zeilen 13 und 15 bis 17 der Tabelle (Fälle 226, 229, 233, 234, 235). Akzeptiertes Negativ zu (3): ein Signal an die Prozessgruppe während des Aufräumens hat keinen Fall; Grund: der Ersatz für `docker` signalisiert allein den Treiber; Folge: höchstens eine überzählige Meldung `Aufräumen gescheitert`, der zweite Durchgang entfernt das Objekt (aus dem Quelltext gelesen, nicht gefahren). Folge: keine Fix-Runde; der Reviewer hält den Treiber gegen den Wortlaut.
- `COPY` übernimmt die Rechte der Einträge eines Verzeichnisses, nicht die des kopierten Verzeichnisses selbst (gemessen 2026-11-27: `/data` und `/tmp` standen sonst auf `0755`) — **Ausgang:** entfallen: entschieden (Implementer, innerhalb von `SPEC-322` Nr. 2): beide Verzeichnisse entstehen in der Hilfsstufe `image-verzeichnisse` unter je einem eigenen Elternverzeichnis; die Prüfungen nach Nr. 3 fangen die Rechte.
- Die Aufrufe von `docker` im Kapitel des Handbuchs führt kein Gate aus (`SPEC-200` Nr. 14 e) — **Ausgang:** weiter offen; Register `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`. Review und Verifikation haben die Aufrufe an diesem Stand von Hand gefahren, ohne Widerspruch (Version 1.2.3; nicht gefahren: Docker Compose, eingehängte Dateien); eine spätere Änderung des Kapitels oder des Images prüft kein Lauf.
- Modus `mock` und `replay` am Image und ein zweiter Neustart haben keinen Fänger (`SPEC-323` Nr. 7, 9 b) — **Ausgang:** entfallen: entschieden (Architect): benannte Negative.
- Review F-1: „trägt genau“ (`SPEC-322` Nr. 2) und kein `HEALTHCHECK`, `EXPOSE` (Nr. 8) ohne Fänger; zur selben Klasse gehört „setzt allein `DATA_DIR`“ (Nr. 4) — **Ausgang:** entfallen: entschieden (Architect, 2026-11-28, `SPEC-323` Nr. 3 und 9 h): Fänger statt Negativ. **Fix-Runde** (Implementer), geliefert (2026-11-28, §3 *Fix-Runde zum Review*): (1) das Format von `image inspect` nennt zusätzlich `Config.ExposedPorts` und `Config.Healthcheck`, rot `EXPOSE` bzw. `HEALTHCHECK`; (2) außer `DATA_DIR` und `PATH` ist jede Variable rot `Variable <Name>`, der Name ganz verglichen; (3) `docker export` des Containers `pruefung`, jeder Name aus `tar -t` ganz und buchstäblich in der Menge von 18, sonst rot `Inhalt` mit dem Namen. Gegenproben: der Ersatz führt `export`; Fall `image` mit `ExposedPorts`, `Healthcheck`, weiterer Variable (`UPSTREAM_URL`, `PATHX`, `XPATH`, `DATA_DIRX`; die bisher erlaubten `MODEL` und `LISTEN_TLS_CERT` werden rot) und je Ort ein zusätzlicher Name (Wurzel `bin/sh`, `data/`, `tmp/`, `etc/`, `etc/ssl/certs/`, `dev/`, ein leeres Verzeichnis, dazu je ein Name, der einen erlaubten als Anfang bzw. als Ende trägt); Fälle `gut`, `exit-codes` und die Zahlen in `SPEC-102` folgen. Ein Lauf in einer Wegwerfkopie mit Shell, Datei in `/data`, `EXPOSE` und `HEALTHCHECK` je einzeln rot.
- Review F-2: das zweite Signal während des Aufräumens ohne Fänger (`SPEC-323` Nr. 6 d) — **Ausgang:** entfallen: entschieden (Architect, 2026-11-28, Zeile 16): Fänger; die ungeführte Stelle im Durchgang steht als akzeptiertes Negativ in §3 (V-3). **Fix-Runde** (Implementer), geliefert (2026-11-28, §3): Fall `abbruch` zusätzlich drei Läufe: als erstes Signal TERM während des ersten Baus, ein zweites während des ersten `rm` des Aufräumens, je INT, HUP und TERM; der Ersatz führt das zweite `rm` danach aus. Erwartet: Exit 2, genau eine Meldung `abgebrochen`, je Container und Tag ein Entfernen, nichts bleibt. Mutationen: `trap ''` im Abbruch entfernt; je ein Signal aus seiner Liste entfernt.
- Review F-3: SIGPIPE beendet den Treiber ohne Aufräumen; Randform ohne Ort — **Ausgang:** entfallen: behoben; der Abbruch bei geschlossenem Leser ohne Fall steht als akzeptiertes Negativ in §3 (V-4). Entschieden (Architect, 2026-11-28, `SPEC-323` Nr. 6 e, Raster Zeile *Aufräumen*, Zeile 15): der Treiber ignoriert SIGPIPE, Ablauf, Aufräumen und Exit sind die des Laufs. **Fix-Runde** (Implementer), geliefert (2026-11-28, §3; sechs statt zwei Läufe, je Weg grün, rot und Baufehler jeder Strom): SIGPIPE am Anfang ignorieren; Gegenproben mit Läufen, in denen die Standardausgabe (grüner Lauf, Exit 0) bzw. der Fehlerstrom (roter Lauf am Image, Exit 1) das Schreibende einer Pipe ist, deren Leseende vor dem Start geschlossen ist: alle Aufrufe wie ohne die Pipe, nichts bleibt, `TMPDIR` leer. Mutation: das Ignorieren entfernt.
- Review F-5: `docker stop --time` ist veraltet — **Ausgang:** entfallen: entschieden (Architect, 2026-11-28, `SPEC-323` Nr. 4 c; `SPEC-215` Nr. 7 b und Zeile 44 von [slice-dev-container-geruest](../open/slice-dev-container-geruest.md) nachgezogen): `-t 30`. **Fix-Runde** (Implementer), geliefert (2026-11-28, §3): Treiber und `tiSoll` in den Gegenproben.
- Review F-6: jeder Lauf lässt rund 230 MB und 7 Einträge im Bau-Cache des Hosts — **Ausgang:** entfallen: entschieden (Architect, 2026-11-28, `SPEC-323` Nr. 6 f): akzeptiertes Negativ, DoD 2 nennt die Reichweite des Aufräumens. Folge: keine Fix-Runde.
- Review F-7: dass die Bodies den Neustart überstehen, belegt allein der Status 201 des zweiten Imports — **Ausgang:** entfallen: entschieden (Architect, 2026-11-28, `SPEC-323` Nr. 4 d, 9 i, Zeile 9): der Status genügt nicht, weil ein übersprungener Exchange nie geöffnet wird; mit der Zahl der gelesenen Exchanges genügt er (`Open` prüft die Datei gegen ihre Referenz, `SPEC-106` Nr. 5). **Fix-Runde** (Implementer), geliefert (2026-11-28, §3): `TestImagePruefen` verlangt im Body des 201 `"exchanges":{"read":2,"skipped":[],"total":2}`; die unveränderte Aufnahme trägt diese Zahlen. Mutation in einer Wegwerfkopie abweichend vom Vorschlag, siehe die nächste Zeile.
- Abweichung zu F-7: die vorgeschlagene Mutation (`/Products` als `text/plain`) erreicht die neue Zeile nicht. Der Recorder verwirft den Body schon beim Aufzeichnen (`ODR-E2401`, Grund `typ`), die Session ist `incomplete`, und die Phase `aufnehmen` ist an `close` rot (ein Lauf, 2026-11-28). Gefahren ist stattdessen `application/json;odata.metadata=none` (erfasst, vom Import mit `ohne_kontext` übersprungen, `SPEC-323` Nr. 8 Zeile 9: „eine Antwort auf `/Products`, die der Import überspringt“) — **Ausgang:** entfallen: der Verifier hat beide Hälften bestätigt; die gefahrene Mutation trägt Zeile 9 von `SPEC-323` Nr. 8 (`aufnehmen` grün, `pruefen` rot an der neuen Zeile). Der Vertrag nennt `text/plain` nirgends, an ihm ist nichts zu ändern; der Vorschlag stand allein in der Zeile *Review F-7* dieses Plans als Anweisung an die Fix-Runde (Architect, `c7af0a70`).
- Fix-Runde zusammen (Implementer), geliefert (2026-11-28; offen allein der Stempel im Integrationsschritt und die Zeile zu F-7 oben): je Punkt oben die Mutation rot sehen und mit Zahl der Prozesse und Messendem unter §3 nachtragen; Mengen als Schleife (`AGENTS.md` §3.17); `make test-image` grün; im selben Commit `SPEC-102` (Fälle, Zahlen), der *Stand* von `SPEC-323`, der Kopfkommentar des Treibers und *Nicht behauptet* in `harness/README.md` um `SPEC-323` Nr. 9 h und 9 i — **Ausgang:** entfallen: geliefert in `ac7d2d2f` und von der Verifikation gegen die Klassen der Befunde geprüft; der Stempel steht in der Zeile *Review F-8*.
- Verifikation V-5: *Nicht behauptet* in `harness/README.md` nennt zum Binary-Vergleich allein „über verschiedene Hosts (Nr. 3)“; `SPEC-323` Nr. 9 g schließt dazu den Bau aus leerem Cache, andere Bau-Dienste, Kernzahlen und Plattformen aus — **Ausgang:** weiter offen; Register `BEO-REPO/ein-negativ-benennt-die-ungefangene-klasse-zu-eng`; Adresse: Implementer, eigener Commit ohne Slice (ein Halbsatz mit Verweis auf Nr. 9 g), vor der Closure von [welle-betrieb](../welle-betrieb.md); die Sendung steht dort in §2. Der Planner schreibt `harness/README.md` nicht.
- Review F-4 (Kopf und Hilfetext von `make gegenproben`) — **Ausgang:** entfallen: Kopf und Hilfetext in `harness/mk/gegenproben.mk` nennen den Treiber des Neustarttests (Implementer, 2026-11-28).
- Review F-8 (Stufe `gegenproben` unter selbst erzeugter Last) — **Ausgang:** weiter offen; Register `BEO-REPO/gegenproben-erzeugen-die-last-selbst`. Gesehen: zwei Bauten der Stufe im Review rot allein durch `lint/*` mit `Zeit überschritten` (9 und 5 von 236 Fällen, wechselnde Menge; Last von 10 und 6 auf 127 bis 254), die elf Fälle `testimage/*` in beiden grün; ob der Lauf ohne diesen Slice anders endet, ist nicht gemessen. `harness/gegenproben.stempel` passt nicht zum Stand (`39ea6316…` liegt, `62fd0968…` folgt aus den Eingaben, Verifikation Auflage 2); `make gegenproben` ist an diesem Stand nicht gefahren, die Fix-Runde hat die Stufe nicht gebaut (Fälle des Moduls in einem Container ohne Netz, `go test ./...`, grün). Adresse: Planner, Integrationsschritt (`make gegenproben` auf ruhigem Host, Stempel allein; Übergangsregeln Nr. 2); [slice-gegenproben-eichung](../open/slice-gegenproben-eichung.md) §1 führt den Befund als Sendung. **Offene Frage an den User:** wird die Eichung samt der Entscheidung in ihrem §6 (drei Wege) vor diesen Lauf gezogen, oder bleibt die Lücke bei einem roten Lauf offen? Solange `make gegenproben` nicht grün ist, schließt [welle-betrieb](../welle-betrieb.md) nicht.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; aus `SPEC-323` Nr. 7 übernommen, Architect; „Nr.“ ohne Kennung meint `SPEC-323`, *Vol.* Modus und Bestand des Datenverzeichnisses; die Zeile *Aufzeichnen* trägt Import, Workspace und Snapshot mit):

| Operation | fehlend | leer | unbekannt oder fremd | Abbruch | Wiederholung | Vol. |
|---|---|---|---|---|---|---|
| Image bauen | Stufe oder Quelle fehlt: Nr. 6 c (Exit 2) | `VERSION` leer oder ungesetzt: `dev` (Nr. 2 a) | Fehler beim Bau: Nr. 6 c; fremder Inhalt der Stufe: Nr. 3 | Nr. 6 d | Cache; gleiches Binary nach Nr. 3, gleiche Kennung des Image nicht zugesagt (`SPEC-322` *Ungeprüft*) | gilt nicht: der Bau liest keinen Bestand |
| Start | ohne Volume: `SPEC-322` Nr. 5 | leeres Verzeichnis: `SPEC-322` Nr. 5, Nr. 4 b; `DATA_DIR` leer: `SPEC-322` Nr. 5 | nicht beschreibbar, fremder Eigentümer: `SPEC-322` Nr. 5 (ohne Fänger); Schema: `SPEC-103` Nr. 4, 5; Token endet nicht: Nr. 4 a | `SPEC-138` Nr. 7; im Test Nr. 6 d | Nr. 4 c | `SPEC-139` Nr. 11 f, `SPEC-322` Nr. 5 |
| Aufzeichnen | Referenzdienst fehlt: Phase rot (Nr. 4 e) | leere Aufnahme: gilt nicht, `SPEC-037` Nr. 7 | fehlgeschlagener Import oder Aufruf: Phase rot (Nr. 4 e) | `SPEC-139` Nr. 7; erzwungen: `SPEC-322` Nr. 7 (ohne Fänger) | gilt nicht: je Lauf eine Aufnahme auf frischem Bestand | nur `record` (Nr. 2 d) |
| Neustart | Server vorher beendet: Stopp liefert dessen Exit, ungleich 0 ist rot (Nr. 4 c) | gilt nicht: kein Wert | Exit ungleich 0 oder erzwungenes Ende (137): rot (Nr. 4 c) | `SIGTERM`: Nr. 4 c; `SIGKILL`: ohne Fänger (Nr. 9) | ein Neustart; ein zweiter ohne Fänger (`SPEC-139` Nr. 10) | derselbe Container, `record` |
| Lesen nach dem Neustart | fehlender Wert: rot (Nr. 4 d) | leere Liste: rot (Nr. 4 d) | abweichende Bytes, 401, 404: rot (Nr. 4 d) | Nr. 6 d | gilt nicht: Lesen ändert nichts, `restore` läuft einmal | Modus `mock` und `replay` auf dem Bestand: ohne Fänger am Image (Nr. 9) |
| Aufräumen | Objekt nie entstanden: Nr. 6 a | gilt nicht: kein Wert | Objekte anderer Läufe: nie berührt (Nr. 6 d) | Nr. 6 d; geschlossener Leser der Ausgabe: Nr. 6 e | je Objekt ein Entfernen; nach einem Signal im Aufräumen von vorn, jedes weitere ohne Wirkung (Nr. 6 d) | gilt nicht: kein Volume (Nr. 2 e) |

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

**Closure** (Planner, 2026-11-28; Zweig `slice-container-image`, Messbasis `0158092b`, `main` bei `39a21107` vor der Closure aufgenommen, Merge `84ffa38a`). Start `6d1a7cbe`, Lieferung `0194b5d4`, Entscheidungen des Architect vor dem Review `b39db79d`, Prüfung der neuen Übersetzung `90daeec7`, Review-Report `c5b74993`, Entscheidungen zum Review `c7af0a70`, Fix-Runde `ac7d2d2f`, Merge von `main` `6453c3e4`, Verifikations-Report `8217dcc8`. Review (`2026-11-28-slice-container-image-review`: 0 HIGH, 2 MEDIUM, 2 LOW, 4 INFO; nicht ohne Nacharbeit wegen F-1 und F-2) und Verifikation (`2026-11-28-slice-container-image-verifikation`: bestätigt mit Auflagen; DoD 1 bis 3 erfüllt, DoD 4 teilweise; 0 HIGH, 0 MEDIUM, 3 LOW, 2 INFO) liefen je einmal; die Fix-Runde hat allein die Verifikation geprüft. Merge nach `main`, `make gates`, `make gegenproben` und der Stempel sind der Integrationsschritt (Übergangsregeln Nr. 2 und 5).

**Auflagen der Verifikation** (kein Code; den Text der Closure hat niemand nachgeprüft): **1:** Gate-Lauf unten, Messung auf dem Stand nach der Fix-Runde und Zeitbudget in §3 (V-1). **2:** offen für den Integrationsschritt (§6, Review F-8). **3:** jeder Punkt in §6 trägt einen der drei Ausgänge; die Zeile *Abweichung zu F-7* ist *entfallen*. **4:** offene Frage an den fachlichen Owner in der [Roadmap](../in-progress/roadmap.md), Zeile M7. **V-2:** Zeile in der Tabelle von §3. **V-3, V-4:** akzeptierte Negative in §3. **V-5:** §6, Adresse Implementer.

**Was geliefert ist:** das Laufzeit-Image (Stufe `image`, `make image`, `SPEC-322`): leere Basis, Binary, Nutzer `65532`, `/data` als `DATA_DIR`, kein `VOLUME`; das Gate `make test-image` (`SPEC-323`): Prüfungen am Image (Angaben, Variablen, Inhalt, Rechte, Version), zwei Übersetzungen mit byte-gleichem Programm, Aufzeichnung über beide Listener, Stopp und Start desselben Containers ohne Volume, danach Recordings, Workspaces und Snapshots lesbar, Aufräumen im Daemon des Hosts; die Gegenproben 226 bis 236; das Handbuch-Kapitel zum Image. `src/` ist nicht berührt.

**Gate-Lauf** (Ersatz für den CI-Test, `LH-QA-06`). `make gates` im Worktree am Stand `6453c3e4`, unter der Sperre der schweren Läufe, Beginn 2026-11-28T10:47:48+01:00, **Exit 0, 1170 s**, Load Average 9,10 am Beginn und 8,37 am Ende, Host mit 20 Kernen, Docker 29.10.0. Messender: der Implementer der Fix-Runde; der Planner hat die Ausgabe aus dessen Scratchpad gelesen (sie liegt nicht im Repo), `.harness/state/gates-passed.head` trägt denselben Commit, und die Verifikation sah den Hash des Nachweises gleich dem des Arbeitsbaums. Aus der Ausgabe:

```text
baseline-verify: v6.18.1 OK — 54 Dateien (Integritaet + Vollstaendigkeit, netzlos)
d-check: 1466 Datei(en) geprüft, 0 Befund(e)
#22 511.0 pruefung race: ok
groessenmarke slice-container-image.md: ok, Messung 2026 von Marke 2800
test-image: Lauf 3349265-1795860280
test-image: Image geprüft
test-image: zwei Bauten, Binary byte-gleich
test-image: Phase aufnehmen grün
test-image: Server gestoppt (Exit 0) und gestartet
test-image: Phase pruefen grün
test-image: grün
test-contract: grün
test-determinism: grün (TestE2EDeterminismus)
Stufe 11: TestE2EAdmin grün
test-security: grün
zeichen: ok, 2675 Datei(en) geprüft, 6 Binärdatei(en) und 5 andere Einträge übergangen
```

Nach `6453c3e4` ändern sich allein Dateien unter `docs/` und `spec/` (Report, Merge von `main`, Closure); **diesen Stand hat kein Lauf der Kette gedeckt**, `make docs-check` lief auf dem Merge `84ffa38a` grün. Welche übrigen Stufen aus dem Schichten-Cache kamen, ist nicht ausgewertet.

**Was ungeprüft bleibt** (Grenzen des Belegs): (1) kein Lauf in einer CI (`LH-RB-03`, Auflage 4); (2) eine Architektur (x86_64) und eine Laufzeit; (3) `LH-QA-06` „gleiches Testverhalten“ belegt nicht dieser Slice, sondern `make test-determinism` am Binary der Test-Stufe, nicht am Laufzeit-Image; (4) eingehängtes Volume, neuer Container auf altem Bestand, `SIGKILL`, `mock` und `replay` am Image, ein zweiter Neustart (`SPEC-323` Nr. 9); (5) Reichweite des Binary-Vergleichs `SPEC-323` Nr. 9 g; (6) `docker compose`, eingehängte Zertifikats- und Token-Dateien und die Wartezeit von 10 s aus dem Handbuch sind nicht gefahren; (7) Wiederholraten der Mutationen der Verifikation (je ein Lauf).

**Größe** (`SPEC-128` Nr. 9). Zeile des Programms (`make groessenmarke`, 2026-11-28, Plan in `in-progress/`, nach dem Merge `84ffa38a`): `Messpunkt slice-container-image.md: zeilen=20 nebenlaeufig=0 fakes=1 produktcode=220 schaetzung=2020 marke=2800 messung=2026`. Auslastung 2026 ÷ 2800 = **0,72**, Ende ÷ *S* = 1,00; zwischen 0,45 und 1,00, keine Meldung an den Architect. Keine Überschreitung. Handoff vor der Fix-Runde 1718 (§6).

**Mutationen:** §3 (Implementer: 49 Mutanten gegen die Fälle am Stand der Übergabe, Läufe je Zeile, 34 Mutanten der Fix-Runde); Review: eigene Läufe am Image und vier Treiber-Mutanten; Verifikation: Mutationen je Befund an Elementen, die weder Review noch §3 nennen (Report).

**Abweichungen vom Plan:** zusätzlich zum Vertrag beim Start `--no-cache-filter image` und die Prüfung des Alters (`SPEC-323` Nr. 3); `spec/spezifikation.md` geändert (`SPEC-322` Nr. 4); DoD 4 ohne die Messung nach `SPEC-211` Nr. 4 (§3 *Zeitbudget*). Keine Rückführung aus §4 griff.

**Steering-Loop-Einträge:**

1. **Neuer Sensor:** das Gate `make test-image` (`SPEC-323`): Image, Neustart mit Bestand, byte-gleiches Programm aus zwei Übersetzungen.
2. **Lernsignal:** ein grüner Vergleich oder Status belegt nichts, solange nicht belegt ist, dass der Weg dorthin gegangen wurde. Dreimal in diesem Slice: der Vergleich zweier Bauten las zweimal dasselbe Binary aus dem Cache (gefunden vom Implementer, jetzt prüft der Treiber das Alter); der Status 201 des zweiten Imports kam auch für einen übersprungenen Austausch, dessen Body nie geöffnet wird (Review F-7, jetzt die Zahl der gelesenen); die dafür vorgeschlagene, nie gefahrene Mutation des Architect hielt schon eine frühere Phase an und erreichte die neue Zeile nicht (gefunden in der Fix-Runde). Praktisch: zu jedem Gleichheits- oder Statusbeleg die Mutation fahren, die beide Seiten zu derselben Quelle macht oder den Weg überspringt, und sehen, dass sie an der neuen Zeile rot wird.
3. **Kein `liegt in`-Feld:** keine Regel ist mit diesem Slice verkörpert.

**Lerneintrag:** neuer Sensor (Eintrag 1) mit dem Lernsignal (Eintrag 2). Eine Spec-Lücke bleibt nicht offen: die Randformen ohne Ort hat der Architect im Slice im Vertrag entschieden.

**Reviewer-Skill:** keine Änderung. F-1 und F-2 sind an `AGENTS.md` §3.15 und §3.16 gefunden, F-3 an §3.12.

**Register** (Evidenz mit der Closure im Zweig; kein Zähler gesetzt, er folgt aus den Dateien): Evidenz `slice-container-image` zu `gate-lauf-stufen-aus-dem-schicht-cache` (zweiter Bau), `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (F-1, F-2, V-3, V-4, Aufrufe im Handbuch), `menge-ohne-ihre-positionen` (V-3), `randform-erst-im-review-entschieden` (F-3), `berichtigung-erreicht-nicht-jede-stelle` (F-4, V-2), `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (V-5), `bericht-der-rolle-liegt-nicht-im-repo` (Ausgabe des Gate-Laufs, Auflage 1), `gegenproben-erzeugen-die-last-selbst` (F-8; zweiter Vorgang, Stand fortgeschrieben); neu `zeitbudget-der-kette-ohne-messung-nach-neuem-ziel` (V-1). **Schwellen:** kein neuer Übertritt durch diesen Slice; die Einträge mit 2× aus §8 haben keinen dritten Beleg. Den Lese-Schritt führt die Closure von [welle-betrieb](../welle-betrieb.md).

**Risiko-Ausgänge (§6):** 29 Punkte, je ein Ausgang: entfallen 25, eingetreten 0, weiter offen 4 (Zeitbudget und Frist; Aufrufe im Handbuch ohne Gate; V-5; Gegenproben und Stempel).

**Sendungen (`AGENTS.md` §3.13):** an [welle-betrieb](../welle-betrieb.md) §2: Messung nach `SPEC-211` Nr. 4, Halbsatz zu V-5, Lauf der Gegenproben mit Stempel; an [slice-gegenproben-eichung](../open/slice-gegenproben-eichung.md) §1: der Befund F-8; an die [Roadmap](../in-progress/roadmap.md), M7: die Frage zu `LH-RB-03` und der Handbuch-Satz zur Betreiberpflicht, mit diesem Merge fällig.

**Folge-Slice:** keiner neu. [slice-gegenproben-eichung](../open/slice-gegenproben-eichung.md) und [slice-ci-github-actions](../open/slice-ci-github-actions.md) liegen in `open/`.

**Drei Paarungen** (hier geprüft; die Closure der Welle prüft sie erneut): Anker (kein Eintrag trägt ein Zielort-Feld) · Folge-Slice (die zwei genannten liegen in `open/`) · Register (die in §6 bis §8 genannten Verzeichnisse tragen `evidence/`, die neun mit neuer Datei `evidence/slice-container-image.md`).

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

**Vorgelagert — offene Beobachtungen sichten** (Planner, beim Start, 2026-11-27; der Zähler ist die Zahl der Dateien unter `evidence/`).

Offen und mit dem Slice berührt:

- `BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache` (20×, Schwelle überschritten, ohne Zielort): ein grüner Lauf belegt die neuen Fälle der Gegenprobe und die Prüfungen am Image nur, wenn Stufe und Treiber wirklich liefen; DoD 4 nennt die Ausgabe.
- `BEO-REPO/schaetzformel-unterschaetzt-werkzeug-mit-prozessen` (2×): §6.
- `BEO-REPO/image-tag-race-von-allen-worktrees-geteilt` (2×): `make image` setzt das feste Tag `odata-recorder:image`, das alle Arbeitsbäume teilen; `make test-image` nutzt Tags je Lauf (`SPEC-323` Nr. 2 b). Mutationsläufe prüfen nie gegen ein geteiltes Tag fremden Standes.
- `BEO-REPO/rollen-images-ohne-aufraeumen` (1×): jede Rolle entfernt die Images und Container ihrer eigenen Läufe, nur diese.
- `BEO-REPO/exit-aussage-ohne-messstelle` (1×): Zeile 17 misst den Exit am Treiber, nicht allein am `make`-Aufruf.
- `BEO-REPO/sendeseite-der-marker-ohne-faenger` (1×): Zeile 19 hält auch, dass das Muster des Tokens beim Client ankommt, sonst ist die Suche stumm.

Geplant oder verkörpert, weiter zu beachten: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (7×, geplant in `slice-coverage-gate`), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (89×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (21×, §3.15), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (15×, §3.10), `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (23×, Übergangsregeln Nr. 8) und `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (4×, Übergangsregeln Nr. 2).

Ein dritter Vorgang der Einträge mit 2× braucht bei der Closure einen Ausgang im Register.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
