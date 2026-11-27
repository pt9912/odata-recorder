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

**Berührte Spec-Stellen:** `LH-QA-06.a` (Container und Neustart) · `LH-RB-03.a` (Linux und CI, Container-Build, End-to-End-Test) · `SPEC-077` (`DATA_DIR`) · `SPEC-035` (OCI-Runtime-Image) · `SPEC-037` (Test, Integration) · `SPEC-123` (Start: Wiederherstellung, Integritätsprüfung) · `SPEC-059` (`SHUTDOWN_TIMEOUT`) · `LH-FA-02.c` (geordneter Shutdown) · `SPEC-124` Nr. 14 · `ARC-012`.
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

**Ziel:** Ein Laufzeit-Image des Dienstes (Stufe des `Dockerfile`, reproduzierbar, Eingangs-Image per Digest gepinnt) startet ohne externe Infrastruktur mit einem Datenvolumen als `DATA_DIR`; ein frisches, leeres Volumen ist ein gültiger Startzustand (`LH-QA-06.a`). Der Neustarttest fährt gegen dieses Image: leeres Datenvolumen, Aufzeichnung gegen einen Referenzdienst, Neustart des Containers, die Recordings sind danach lesbar (Closure-Kriterium der Welle, `LH-QA-06.a`).

**Entkoppelt von `welle-dev-container`** (Entscheidung des Users, 2026-11-27, Antwort 1; Wortlaut in [welle-betrieb](../welle-betrieb.md) §4): Die Bedingung vom 2026-10-30 (Start erst nach `welle-dev-container`, Neustarttest nur im Daemon eines Arbeitscontainers, Verbot eines Laufs im Daemon des Hosts) ist **aufgehoben**. Der Neustarttest läuft im Docker-Daemon des Hosts, nicht abgeschottet; das ist der benannte Preis der Entscheidung. **Bedingungen an das Werkzeug** (Planner, 2026-11-27, damit der Lauf neben den Gate-Läufen im selben Daemon besteht): er ist ein Werkzeug außerhalb von `GATE_CHECKS`, sein Image trägt einen eigenen Tag (keiner aus `odata-recorder:*` der Gates), alles, was er anlegt (Image, Container, Volumes, Netz), trägt ein eigenes Label, der Lauf läuft unter `flock /tmp/odata-recorder-heavy.lock` und räumt im Daemon des Hosts auf, was er angelegt hat, und nur das, auch nach rotem Lauf und Abbruch.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Arbeitscontainer für Entwickler (`LH-RB-03`) — `welle-dev-container`; Gegenstand ist das Image des Dienstes, nicht die Umgebung, in der gebaut wird.
- Abschottung des Testlaufs durch einen eigenen Daemon — Bestand mit Grund: der User hat den Lauf im Daemon des Hosts entschieden (Antwort 1); kein Folge-Slice, und `welle-dev-container` bekommt daraus keine Sendung.
- Eine CI (Server, Pipeline, Runner) — der Lauf über `make` ist als Ersatz für den CI-Test der Messmethode von `LH-QA-06` anerkannt (User, 2026-11-27, Antwort 4); die Änderung am Lastenheft schreibt der Architect.
- Orchestrierung des Betreibers (Compose-Dateien, Kubernetes), Registry, Veröffentlichung des Images — Out-of-Scope der Welle ([welle-betrieb](../welle-betrieb.md) §6): das Image und sein Start sind Gegenstand, nicht seine Umgebung.
- Lesen von Workspaces und Snapshots über die Prozessgrenze nach dem Neustart — bis `welle-admin-api` gibt es weder Kommando noch Endpunkt, die Workspace oder Import im Container anlegen (die CLI-Kommandos entfallen, Entscheidung des Users, 2026-10-30); Recordings entstehen über den Listener im Modus `record`. Der Test deckt Recordings; die Zusage für Workspaces und Snapshots in `LH-QA-06.a` hat in diesem Slice keinen Fänger (akzeptiertes Negativ mit Folge, siehe §6, Adresse: Planner für `slice-bootstrap-m6`).
- Healthcheck über `GET /healthz` — der Endpunkt gehört dem Admin-Listener, `welle-admin-api`; ein Healthcheck des Images über den OData-Listener wäre ein Vertrag ohne Grundlage.
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

- [ ] Image: eine Stufe des `Dockerfile` (oder ein eigenes Fragment) baut das Laufzeit-Image reproduzierbar aus dem Binary der Stufe `build`; das Datenvolumen ist `DATA_DIR`; der Prozess läuft ohne Root-Rechte; der Start aus einem leeren Volumen gelingt (`LH-QA-06`, `LH-RB-03.a`).
- [ ] Neustarttest: leeres Datenvolumen, Aufzeichnung über den Listener im Modus `record` gegen einen Referenzdienst, Neustart des Containers, Recordings lesbar; ist ein `make`-Ziel außerhalb von `GATE_CHECKS` mit eigenem Tag und Label, läuft unter `flock /tmp/odata-recorder-heavy.lock` im Daemon des Hosts und hinterlässt dort weder Image noch Container, Volume oder Netz seines Labels (`LH-QA-06.a`).
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary): Start des Images, Datenvolumen, Konfiguration über die Umgebung; das Kataloge-Gate ist grün.
- [ ] `make gates` grün; der Neustarttest liegt außerhalb der Kette, sein grüner Lauf steht in §7 mit Commit, Ausgabe, Last des Hosts und Messendem (Ersatz für den CI-Test, `LH-QA-06`).
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`harness/README.md` §Sensors, falls ein Ziel entsteht).
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
| `Dockerfile` | update | Stufe des Laufzeit-Images (Basis per Digest, Nutzer, Volumen, Eintrittspunkt) |
| `harness/mk/` (neues Fragment) und `Makefile` | neu, update | Ziele für Bau und Neustarttest; Werkzeug, nicht in `GATE_CHECKS` (§1) |
| Treiber und Test des Neustarts (Ort: Auftrag 3, `tools/…` oder `tools/harness/…`) | neu | Start, Aufzeichnung, Neustart, Prüfung; Referenzdienst als Fixture |
| `harness/vertraege/` | neu | Vertrag des Ziels (Zusagen, Randformen), falls ein Ziel entsteht ([ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)) |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` | update | Fälle der Gegenprobe zum Ziel ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); der Stempel erst im Integrationsschritt |
| `harness/README.md` | update | Zeilen der Ziele in §Sensors, *Nicht behauptet* |
| Benutzerhandbuch (Ort und Form legt `slice-handbuch-geruest` fest) | update | Kapitel Start im Container (Pflicht Handbuch) |
| Tests | neu | Fälle der Zusagen, Mengen als Schleifen |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Aufträge an den Architect:**

1. **Image.** Basis (Digest-Pin wie im `Dockerfile`), Nutzer ohne Root-Rechte und Besitz des Volumens, Pfad des Volumens und Wert von `DATA_DIR` im Image (Default der Spezifikation ist `data` relativ zum Arbeitsverzeichnis, `SPEC-077`), Eintrittspunkt und Signal (`SIGTERM` führt zum geordneten Shutdown, `SPEC-059`), Konfiguration über die Umgebung, Reproduzierbarkeit (`-trimpath`, kein Zeitstempel).
2. **Start aus leerem Volumen.** Wer das Verzeichnis anlegt (`DATA_DIR` legt der Composition Root an, `SPEC-077`), Rechte bei einem vom Betreiber angelegten Volumen mit fremdem Besitzer, Verhalten bei Volumen, das nicht beschreibbar ist (`ODR-E2001`?).
3. **Werkzeug im Daemon des Hosts.** Entschieden ist (§1): Werkzeug außerhalb von `GATE_CHECKS` nach dem Muster von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), Lauf im Daemon des Hosts, eigener Tag und Label, `flock`, Aufräumen. Zu entscheiden: Namen von Ziel, Tag und Label, was das Aufräumen bei rotem Lauf und Abbruch trifft und was es nie anfasst, wer das Ziel wann fährt (`BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger`), und wie der Lauf mit `AGENTS.md` §3.1 zusammengeht: dort sind Prüfläufe netzlos und ohne schreibbaren Mount, der Test braucht ein beschreibbares Datenvolumen und Verbindungen zwischen seinen Containern (§6).
4. **Was „lesbar“ heißt.** Es gibt keinen Lesepfad über die Prozessgrenze (kein CLI, keine Admin-API): der Test öffnet das Volumen mit dem Adapter der Persistenz nach dem Neustart, oder liest über einen Weg, den der Architect begründet; zu entscheiden, ob Neustart `SIGTERM` (geordnet: Session `complete`) und `SIGKILL` (Wiederherstellung: Session `incomplete`, `SPEC-123` Nr. 5) beide getestet werden.
5. **Referenzdienst.** Wie der Referenzdienst für den Container erreichbar ist (Netz des Tests, ohne externen Dienst), und dass der Test nichts außerhalb seiner eigenen Container und seines eigenen Netzes erreicht.
6. **Vertrag und Gegenproben.** Vertrag des Ziels in `harness/vertraege/`, Fälle der Gegenprobe je Fehlform in Wegwerfverzeichnissen, Zeitbudget (der Architect schätzt, der Implementer misst), und ob ein Fall das Produkt mitfährt (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`: die Eingabe enthält nur Fixtures).
7. **Tabelle, Raster, erwartete Produktcode-Größe** für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 7) | — | — |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Der Neustarttest läuft im Daemon des Hosts unter derselben Sperre; er und seine Mutationen löschen dort nur Objekte seines Labels (siehe §1).
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Fehlform des Volumens, jedes Signal), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-handbuch-meldungskatalog` liegen in `done/` (erfüllt). Der Vertrag des Ziels samt Randformen liegt in `harness/vertraege/`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). **Gate-Kette** (Übergangsregeln Nr. 7; er ändert `Dockerfile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/` und den Stempel): kein anderer Kettenslice liegt in `in-progress/`; Vorgänger der Kette ist [slice-bootstrap-m6](slice-bootstrap-m6.md) in `done/` (Planner, 2026-11-27, Drift-Log der Roadmap). `welle-dev-container` ist keine Bedingung mehr (§1).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-bootstrap-m2` (`welle-record-replay`): der Prozess, den das Image startet: Startbefehl mit `LadeOderMelde`, Verdrahtung, Listener im Modus `record`, geordnetes Herunterfahren und Startwartung; heute ruft der Adapter `cli` (`cli.Run`) keinen Start. Der Container startet mit `start`; `CMD ["start"]` ist Sache dieses Slice (Sendung von `slice-cli-adapter`, `SPEC-138`).
- `slice-handbuch-meldungskatalog` (`welle-benutzerhandbuch`, setzt `slice-handbuch-geruest` voraus): die DoD verlangt ein Handbuch-Kapitel und ein grünes Kataloge-Gate.

**Unabhängig von `welle-mock` / M3, weil** der Neustarttest nur Recordings prüft (Modus `record`); `MockUseCase`, Workspaces mit Mutation und Snapshots sind nicht Gegenstand (§1). Der Trigger der Welle (`welle-mock`) bleibt Sache der Welle; ob der Slice vor ihm läuft, entscheidet der User. **Nicht unabhängig von M2:** ohne `slice-bootstrap-m2` gibt es keinen Dienst im Image. `slice-listener-tls`, die Telemetrie, `slice-audit-log` und `slice-lasttest` sind nicht Bedingung; die Telemetrie ändert das Image nicht.

**Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` (`SPEC-128`), sobald die Tabelle *Zusage · Fänger* vorliegt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht ist dann Image (Bau) gegen Neustarttest (Treiber, Gate oder Werkzeug).
- `in-progress` → `open` (blockiert): der Neustarttest lässt sich nicht fahren, ohne eine harte Regel zu lockern (`AGENTS.md` §3.1, §3.6; siehe §6), oder er kann im Daemon des Hosts nicht auf eigene Objekte begrenzt aufräumen.

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

- Der Neustarttest verlängert die Gate-Kette um einen langen, daemonabhängigen Lauf, dessen Frist unter Last reißt — **Ausgang:** entschieden (Planner, 2026-11-27, §1): er ist ein Werkzeug außerhalb der Kette; die Frist des Laufs selbst trägt Last und Messenden (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
- Ein Werkzeug außerhalb der Kette fährt niemand: nach der Closure bricht der Neustart unbemerkt (`BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger`) — **Ausgang:** offen; Adresse: Architect (Auftrag 3: wer das Ziel wann fährt); für die Closure der Welle trägt [welle-betrieb](../welle-betrieb.md) §3 den Lauf.
- Im Daemon des Hosts kollidiert der Lauf mit Gate-Läufen oder fremden Objekten (feste Tags `odata-recorder:*`, Namen, Netz), oder das Aufräumen löscht Fremdes oder lässt Eigenes stehen — **Ausgang:** offen; Adresse: Architect (Auftrag 3); Fänger: nach grünem, rotem und abgebrochenem Lauf trägt der Daemon kein Objekt des Labels, und ein fremdes Objekt ohne Label steht noch.
- `AGENTS.md` §3.1 verlangt für Prüfläufe netzlos (`--network none`) und ohne schreibbaren Mount; der Neustarttest braucht ein beschreibbares Datenvolumen und Verbindungen zwischen seinen Containern. Die Entscheidung des Users nennt den Daemon des Hosts, nicht diese Regel — **Ausgang:** offen; Adresse: Architect (Auftrag 3: Einordnung im Vertrag oder als ADR; eine Änderung der harten Regel ist eine ADR, `AGENTS.md` §3.6, und geht an den User).
- Zusage-Lücke: `LH-QA-06.a` nennt Recordings, Workspaces und Snapshots, das Closure-Kriterium der Welle nur Recordings; über die Prozessgrenze gibt es für Workspaces und Snapshots keinen Weg bis `welle-admin-api` — **Ausgang:** offen; akzeptiertes Negativ mit Grund und Folge (`AGENTS.md` §3.15): Workspaces und Snapshots über den Neustart prüft bis dahin allein die Persistenz-Ebene (`LH-QA-03`); Adresse: Planner für `slice-bootstrap-m6` (Rundlauf über HTTP, Nehmer ohne Datei).
- Das Volumen gehört einem anderen Nutzer als der Prozess im Image, der Start scheitert oder legt Dateien mit fremdem Besitz an — **Ausgang:** offen; Adresse: Architect (Auftrag 2), Randform vor dem Code.
- `SIGKILL` während der Aufzeichnung hinterlässt eine `open`-Session; der nächste Start setzt sie `incomplete` — **Ausgang:** offen; Adresse: Architect (Auftrag 4), Fänger oder benanntes Negativ.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation, in `harness/vertraege/` oder `gilt nicht` mit Grund):

| Operation | fehlender Wert (`DATA_DIR`, Volumen) | leerer Wert (leeres Volumen, leere Umgebung) | unbekannter oder fremder Wert (Besitz, nicht beschreibbar, Bestand einer anderen Version) | Abbruch (`SIGTERM`, `SIGKILL`) | Wiederholung (mehrfacher Neustart) | Modus des Volumens (`record` auf `replay`-Bestand) |
|---|---|---|---|---|---|---|
| Image bauen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Start aus leerem Volumen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Aufzeichnen im Container | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Neustart des Containers | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Recordings nach dem Neustart lesen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Aufräumen des Testlaufs (Container, Volumen, Image im Daemon) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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
