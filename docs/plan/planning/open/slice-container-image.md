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

**Berührte Spec-Stellen:** `SPEC-322` (Laufzeit-Image) · `SPEC-323` (Neustarttest, `harness/vertraege/`) · `SPEC-102` und `SPEC-211` (`harness/vertraege/`, Nachzug im liefernden Commit) · `SPEC-200` Nr. 14 e (Handbuch) · `LH-QA-06.a` (Container und Neustart) · `LH-RB-03.a` (Linux und CI, Container-Build, End-to-End-Test) · `SPEC-077` (`DATA_DIR`) · `SPEC-035` (OCI-Runtime-Image) · `SPEC-037` (Test, Integration) · `SPEC-123` (Start: Wiederherstellung, Integritätsprüfung) · `SPEC-059` (`SHUTDOWN_TIMEOUT`) · `LH-FA-02.c` (geordneter Shutdown) · `SPEC-124` Nr. 14 · `SPEC-005` (Admin-Endpunkte) · `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

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

- [ ] Image: die Stufe `image` des `Dockerfile` nach `SPEC-322` (leere Basis, Binary der Stufe `build`, Nutzer `65532`, `DATA_DIR=/data`, kein `VOLUME`, Eintrittspunkt mit `start`) und das Ziel `image` (`SPEC-323` Nr. 1); der Start aus einem leeren Datenverzeichnis gelingt (`LH-QA-06`, `LH-RB-03.a`).
- [ ] Neustarttest nach `SPEC-323`: das Ziel `test-image` hängt an `GATE_CHECKS`, hängt kein Volume ein, nimmt im Modus `record` über beide Listener Recordings, einen Workspace und einen Snapshot auf, stoppt und startet denselben Container und liest alle drei mit dem Stand von vor dem Neustart (`LH-QA-06.a`); es räumt im Daemon des Hosts auf, was sein Lauf angelegt hat, und nur das (Nr. 6); der Lauf baut die Stufe `image` ein zweites Mal (`--no-cache-filter build`, Tag `odata-recorder:test-image-zweit-<L>`) und vergleicht die beiden Binaries byte-gleich (`SPEC-323` Nr. 2 b und 3); die 20 Zeilen von §3 haben ihre Fänger.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand): Bau des Image, Volume an `/data`, `LISTEN`, Stopp-Wartezeit und Rechte (`SPEC-323` Nr. 10, `SPEC-322`); `make handbuch-check` führt die Aufrufe von `docker` darin nicht aus (`SPEC-200` Nr. 14 e); das Kataloge-Gate ist grün.
- [ ] `make gates` grün, darin der Neustarttest; §7 nennt seinen Lauf mit Commit, Ausgabe, Laufzeit (Messung nach `SPEC-211` Nr. 4), Last des Hosts und Messendem (Ersatz für den CI-Test, `LH-QA-06`).
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `harness/README.md` §Sensors (`test-image` als Gate, `image` als Werkzeug), `SPEC-102`, `SPEC-211` und der *Stand* von `SPEC-323` im liefernden Commit.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `Dockerfile`, `.dockerignore` | update | Stufe `image` nach `SPEC-322`; was der Bau-Kontext zusätzlich braucht |
| Fragment `image.mk` unter `harness/mk/` | neu | die Ziele `test-image` (an `GATE_CHECKS`) und `image` (`SPEC-323` Nr. 1) |
| `tools/harness/test-image.sh` | neu | Treiber: Aufruf, Namen (drei Image-Tags, darunter der zweite Bau), Bau, Prüfungen am Image samt Vergleich der zwei Binaries (`cmp`), Ablauf, Fristen, Aufräumen jedes Tags, Exit (`SPEC-323` Nr. 2 bis 6) |
| `tests/e2e/image_test.go` | neu | die Client-Phasen `aufnehmen` und `pruefen` mit dem Referenzdienst (`SPEC-323` Nr. 4) |
| `tools/gegenproben/` (neue Fälle mit Ersatz für `docker`), `tools/harness/gegenproben-eingabe.sh` | neu, update | Fälle „GP“ der Tabelle (`SPEC-102`, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); der Stempel erst im Integrationsschritt |
| `harness/README.md` | update | `test-image` in der Tabelle der Gates, `image` in der Tabelle der Werkzeuge, *Nicht behauptet* |
| `harness/vertraege/` (`SPEC-102`, `SPEC-211`, `SPEC-323` *Stand*) | update | Nachzug im liefernden Commit (`SPEC-211` Nr. 2 b: fünf statt vier Container-Tests); der Vertrag selbst ist geliefert |
| Benutzerhandbuch | update | Kapitel zum Image (Pflicht Handbuch, `SPEC-323` Nr. 10) |

Die Dateinamen sind der Vorschlag von `SPEC-323` (Kopf) und, für Test und Gegenproben, des Planners; wer sie ändert, zieht Tabelle und Block im selben Commit nach.

**Vertrag** (geliefert, Architect, 2026-11-27): `SPEC-322` (Image) und `SPEC-323` (Neustarttest) tragen die Antworten auf die Aufträge, die dieser Plan an den Architect gestellt hatte: Image und Datenverzeichnis (`SPEC-322` Nr. 1 bis 8), Lauf im Daemon des Hosts, Netz, Namen, Aufräumen und Abbruch (`SPEC-323` Nr. 1, 2, 5, 6), Prüfungen, Ablauf und Referenzdienst (Nr. 3, 4), Randformen (Nr. 7), Zusagen und Fänger (Nr. 8), Ungeprüftes (Nr. 9) und Größe (Nr. 10). Die Randformen entscheidet der Architect dort, nicht der Implementer (`AGENTS.md` §3.12).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; aus `SPEC-323` Nr. 8 übernommen, Architect; „GP“ ist ein Fall der Gegenprobe mit Ersatz für `docker` in einem Wegwerfverzeichnis, „Lauf“ das Ziel `test-image`, „Nr.“ in den Zellen meint `SPEC-323`; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Nutzer `65532:65532` | Lauf, Nr. 3 | `USER` entfernt | andere |
| 2 | Eintrittspunkt und Standardargument | Lauf, Nr. 3 und 4 a | `CMD` entfernt | andere |
| 3 | `DATA_DIR=/data`, kein `VOLUME`, kein `LISTEN`, `ADMIN_LISTEN`, `MODE` | Lauf, Nr. 3; je Variable ein Fall in GP | `VOLUME /data` | andere |
| 4 | `/data` mit Eigentümer und Rechten | Lauf, Nr. 3 | `--chown` entfernt | andere |
| 5 | Vertrauensanker und `/tmp` | Lauf, Nr. 3 | `COPY` entfernt | andere |
| 6 | Version im Image | Lauf, Nr. 3 | `VERSION` nicht gereicht | andere |
| 7 | frisches Datenverzeichnis | Phase `aufnehmen` | `aufnehmen` zweimal gegen denselben Server | andere |
| 8 | geordneter Stopp mit Exit 0 | Lauf, Nr. 4 c; GP | `STOPSIGNAL SIGKILL` im Image; Exit nicht gewertet | andere |
| 9 | Recordings überstehen den Neustart | Phase `pruefen` (Exchanges, neuer Import) | neuer Container statt `start` | andere |
| 10 | Workspace übersteht den Neustart | Phase `pruefen` | wie Nr. 9 | andere |
| 11 | Snapshot übersteht den Neustart | Phase `pruefen` (Liste, `restore`) | wie Nr. 9 | andere |
| 12 | Token übersteht den Neustart | Phase `pruefen` (Token von vorher) | Treiber reicht `pruefen` ein anderes Token | andere |
| 13 | Optionen der Aufrufe und verbotene Formen (Nr. 2 d, e) | GP: Liste der Aufrufe | je eine Option entfernt oder hinzugefügt | andere |
| 14 | eigene Namen und Label je Lauf | GP: zwei Läufe, getrennte Namen | feste Namen | andere |
| 15 | Aufräumen auf grün, rot und Baufehler | GP je Weg | `rmi` oder `rm` entfernt | andere |
| 16 | Aufräumen bei INT, HUP, TERM, Exit 2 | GP je Signal | `trap` entfernt | andere |
| 17 | Exit-Codes (Nr. 6 b, c) | GP je Ursache | Aufräumfehler verschluckt | andere |
| 18 | Wertung der Phase (`SKIP`, `FAIL`, ohne `PASS`) | GP je Form | `SKIP` nicht gewertet | andere |
| 19 | Token in keiner Ausgabe und keinem Argument | GP: Ersatz liefert ein Zeichen-Muster als Token, Ausgabe und Aufrufliste ohne Treffer | Token als `--env` | andere |
| 20 | zwei Bauten ergeben ein byte-gleiches Binary | Lauf, Nr. 3; GP: Ersatz liefert zwei verschiedene Dateien | Bauzeit über `-ldflags` in das Binary; Vergleich nicht gewertet | andere |

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

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-handbuch-meldungskatalog` liegen in `done/` (erfüllt). Der Vertrag samt Randformen liegt vor dem ersten Code-Commit vor (`AGENTS.md` §3.12; erfüllt: `SPEC-322` und `SPEC-323` auf `main`, `082c560c`). **Gate-Kette** (Übergangsregeln Nr. 7; er ändert `Dockerfile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/` und den Stempel): kein anderer Kettenslice liegt in `in-progress/`; Vorgänger der Kette ist [slice-bootstrap-m6](../in-progress/slice-bootstrap-m6.md) in `done/` (Planner, 2026-11-27, Drift-Log der Roadmap). `welle-dev-container` ist keine Bedingung mehr (§1).

**Stand der Bedingungen** (Planner, 2026-11-27, bei der Closure von Teil D): der Vorgänger [slice-bootstrap-m6](../in-progress/slice-bootstrap-m6.md) liegt in `done/`, und in `in-progress/` liegt kein anderer Kettenslice ([slice-abnahme-belege](../in-progress/slice-abnahme-belege.md) und [slice-lasttest-leser-zeitgrenze](../in-progress/slice-lasttest-leser-zeitgrenze.md) sind keine); **erfüllt, auf `main` nach dem Merge des Zweigs `slice-bootstrap-m6`**. Er ist der nächste Slice der Gate-Kette. Kein Pfad des Blocks steht im Block eines laufenden Slice; das Handbuch (nicht im Block) schreibt auch `slice-abnahme-belege`. `basis`, `schaetzung` und `marke` setzt der Planner beim Start.

**Abhängigkeiten — Grund je Eintrag:**

- `slice-bootstrap-m2` (`welle-record-replay`): der Prozess, den das Image startet: Startbefehl mit `LadeOderMelde`, Verdrahtung, Listener im Modus `record`, geordnetes Herunterfahren und Startwartung. Der Container startet mit `start`; `CMD ["start"]` ist Sache dieses Slice (Sendung von `slice-cli-adapter`, `SPEC-138`; `SPEC-322` Nr. 6).
- `slice-handbuch-meldungskatalog` (`welle-benutzerhandbuch`, setzt `slice-handbuch-geruest` voraus): die DoD verlangt ein Handbuch-Kapitel und ein grünes Kataloge-Gate.
- `welle-mock`, `slice-admin-imports`, `slice-admin-workspaces`, `slice-bootstrap-m6-listener` (alle in `done/`, erfüllt): Workspaces und Snapshots entstehen und werden gelesen über den Admin-Listener im Binary (`LH-QA-06.a`).

`slice-listener-tls`, die Telemetrie, `slice-audit-log` und `slice-lasttest` sind nicht Bedingung; die Telemetrie ändert das Image nicht. `slice-bootstrap-m6` (Teil D) liefert dem Test keinen Code; er liegt allein in der Reihenfolge der Gate-Kette davor.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect: `produktcode` 220 und `fakes` 1 aus `SPEC-323` Nr. 10, 20 Zeilen der Klasse `andere` aus §3; Pfade vom Planner nach §3): noch nicht gesetzt. Aus Tabelle und Block ergeben sich *S* = 2020 und die Marke 2800, keine Schnittfrage (203 unter der Schnittgrenze; geprüft vom Planner am 2026-11-27 mit dem Werkzeug an dieser Datei, `Marke steht aus`; vor Zeile 20 waren es *S* = 1940 und 2700). `basis`, `schaetzung` und `marke` setzt der Planner erst beim Start, vor dem ersten Code-Commit (`AGENTS.md` §3.14), und rechnet neu, falls sich die Tabelle geändert hat. Die Schätzung des Architect für den Umfang ohne Handbuch ist rund 800 Zeilen (`SPEC-323` Nr. 10, nicht gemessen). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; jede weitere Datei, die der Implementer anlegt oder ändert, nennt er im selben Commit im Block.

```groessenmarke
pfad: Dockerfile
pfad: .dockerignore
pfad: harness/mk/image.mk
pfad: tools/harness/test-image.sh
pfad: tests/e2e/image_test.go
pfad: tools/gegenproben/testimage.go
pfad: tools/gegenproben/testimage_test.go
pfad: tools/gegenproben/main.go
pfad: tools/harness/gegenproben-eingabe.sh
pfad: harness/README.md
produktcode: 220
fakes: 1
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; reicht *S* beim Start über die Schnittgrenze von `SPEC-128` Nr. 1, wird vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht ist dann Image (Bau) gegen Neustarttest (Treiber und Gate).
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

- Der Neustarttest verlängert die Gate-Kette um einen daemonabhängigen Lauf, dessen Frist unter Last reißen kann — **Ausgang:** weiter offen; Adresse: Implementer (misst nach `SPEC-211` Nr. 4 mit Last und Messendem; `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`; der zweite Bau nach `SPEC-323` Nr. 3 übersetzt das Programm einmal mehr, Zeit nicht gemessen). Die Aufnahme in die Kette ist entschieden (User, Antwort 2 der sechs Fragen; `SPEC-323` Nr. 1: Zeit geschätzt 10 bis 20 s bei vorhandenen Images, nicht gemessen).
- Ein Werkzeug außerhalb der Kette fährt niemand (`BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger`) — **Ausgang:** entfallen: der Neustarttest ist ein Gate (`SPEC-323` Nr. 1).
- Im Daemon des Hosts kollidiert der Lauf mit Gate-Läufen oder fremden Objekten, oder das Aufräumen löscht Fremdes oder lässt Eigenes stehen — **Ausgang:** entschieden (Architect, `SPEC-323` Nr. 2 b, 5, 6; Fänger Zeilen 14 bis 17 in §3). Benannt bleibt: SIGKILL am Treiber lässt Container, Tags und Verzeichnis liegen (Nr. 6 d dort).
- `AGENTS.md` §3.1 verlangt netzlose Prüfläufe ohne schreibbaren Mount — **Ausgang:** entschieden (Architect, `SPEC-323` Nr. 2 d und e; User, Antwort 1 der sechs Fragen): `--network none` und kein Mount, ohne Ausnahme von der Regel.
- Zusage-Lücke: `LH-QA-06.a` nennt Recordings, Workspaces und Snapshots, der Plan prüfte nur Recordings, weil es keinen Weg über die Prozessgrenze gab — **Ausgang:** entschieden (Planner, 2026-11-27): der Admin-Listener ist im Binary geliefert, der Neustarttest dieses Slice prüft alle drei (§1, DoD 2); das akzeptierte Negativ und die Adresse „Planner für `slice-bootstrap-m6`“ entfallen. [slice-bootstrap-m6](../in-progress/slice-bootstrap-m6.md) hatte die Sendung nie angenommen und nennt den Neustarttest im Container als ausgeschlossen (§1 dort); seine zwei Starts im Prozess ersetzen ihn nicht.
- Der Umfang ist größer als beim ersten Entwurf (drei Datenarten, Admin-Listener im Container, Gegenproben) — **Ausgang:** weiter offen; Adresse: Planner (beim Start `basis`, `schaetzung` und `marke` in den Block, §4; heute *S* = 2020, Marke 2800).
- Das Volume gehört einem anderen Nutzer als der Prozess im Image — **Ausgang:** entschieden (Architect, `SPEC-322` Nr. 5). Ohne Fänger am Image bleiben der Eigentümer eines neuen Volumes und ein nicht beschreibbares Verzeichnis (`SPEC-322` *Ungeprüft*; User, Antwort 1 der sechs Fragen).
- `SIGKILL` während der Aufzeichnung hinterlässt eine `open`-Session — **Ausgang:** entschieden (Architect, `SPEC-323` Nr. 9 a): benanntes Negativ, die Wiederherstellung ist am Dienst geprüft (`SPEC-123` Nr. 5).
- Randformen (`AGENTS.md` §3.12, §3.16): eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, `SPEC-323` Nr. 7, Raster unten).
- Ob `docker cp` den Eigentümer so ausgibt, wie `SPEC-323` Nr. 3 liest, ist nicht gemessen (Nr. 9 f dort) — **Ausgang:** weiter offen; Adresse: Implementer (hält bei Abweichung an und gibt an den Architect zurück, `AGENTS.md` §3.12).
- Die Aufrufe von `docker` im Kapitel des Handbuchs führt kein Gate aus (`SPEC-200` Nr. 14 e) — **Ausgang:** weiter offen; Adresse: Reviewer und Verifier (lesen und fahren sie von Hand).
- Modus `mock` und `replay` am Image und ein zweiter Neustart haben keinen Fänger (`SPEC-323` Nr. 7, 9 b) — **Ausgang:** entschieden (Architect): benannte Negative.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; aus `SPEC-323` Nr. 7 übernommen, Architect; „Nr.“ ohne Kennung meint `SPEC-323`, *Vol.* Modus und Bestand des Datenverzeichnisses; die Zeile *Aufzeichnen* trägt Import, Workspace und Snapshot mit):

| Operation | fehlend | leer | unbekannt oder fremd | Abbruch | Wiederholung | Vol. |
|---|---|---|---|---|---|---|
| Image bauen | Stufe oder Quelle fehlt: Nr. 6 c (Exit 2) | `VERSION` leer oder ungesetzt: `dev` (Nr. 2 a) | Fehler beim Bau: Nr. 6 c; fremder Inhalt der Stufe: Nr. 3 | Nr. 6 d | Cache; gleiches Binary nach Nr. 3, gleiche Kennung des Image nicht zugesagt (`SPEC-322` *Ungeprüft*) | gilt nicht: der Bau liest keinen Bestand |
| Start | ohne Volume: `SPEC-322` Nr. 5 | leeres Verzeichnis: `SPEC-322` Nr. 5, Nr. 4 b; `DATA_DIR` leer: `SPEC-322` Nr. 5 | nicht beschreibbar, fremder Eigentümer: `SPEC-322` Nr. 5 (ohne Fänger); Schema: `SPEC-103` Nr. 4, 5; Token endet nicht: Nr. 4 a | `SPEC-138` Nr. 7; im Test Nr. 6 d | Nr. 4 c | `SPEC-139` Nr. 11 f, `SPEC-322` Nr. 5 |
| Aufzeichnen | Referenzdienst fehlt: Phase rot (Nr. 4 e) | leere Aufnahme: gilt nicht, `SPEC-037` Nr. 7 | fehlgeschlagener Import oder Aufruf: Phase rot (Nr. 4 e) | `SPEC-139` Nr. 7; erzwungen: `SPEC-322` Nr. 7 (ohne Fänger) | gilt nicht: je Lauf eine Aufnahme auf frischem Bestand | nur `record` (Nr. 2 d) |
| Neustart | Server vorher beendet: Stopp liefert dessen Exit, ungleich 0 ist rot (Nr. 4 c) | gilt nicht: kein Wert | Exit ungleich 0 oder erzwungenes Ende (137): rot (Nr. 4 c) | `SIGTERM`: Nr. 4 c; `SIGKILL`: ohne Fänger (Nr. 9) | ein Neustart; ein zweiter ohne Fänger (`SPEC-139` Nr. 10) | derselbe Container, `record` |
| Lesen nach dem Neustart | fehlender Wert: rot (Nr. 4 d) | leere Liste: rot (Nr. 4 d) | abweichende Bytes, 401, 404: rot (Nr. 4 d) | Nr. 6 d | gilt nicht: Lesen ändert nichts, `restore` läuft einmal | Modus `mock` und `replay` auf dem Bestand: ohne Fänger am Image (Nr. 9) |
| Aufräumen | Objekt nie entstanden: Nr. 6 a | gilt nicht: kein Wert | Objekte anderer Läufe: nie berührt (Nr. 6 d) | Nr. 6 d | Entfernen ist je Objekt einmal | gilt nicht: kein Volume (Nr. 2 e) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Slice bringt einen langen Lauf in die Gate-Kette oder daneben: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, jede Frist und Rate trägt Last und Messenden) sind offen; `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (1×) berührt den Stempel der Gegenproben, den dieser Slice nur im Integrationsschritt setzt; offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
