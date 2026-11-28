# Verifikation slice-container-image (Laufzeit-Image und Neustarttest) — 2026-11-28

**Rolle:** Verifier (genau ein Lauf). **Stand:** Zweig `slice-container-image`, HEAD `6453c3e4`. **Eingang:** Plan `slice-container-image` (DoD, §3, §6), [Review](2026-11-28-slice-container-image-review.md) und Fix-Runde `ac7d2d2f`.
**Bezug:** `LH-QA-06`, `LH-RB-03`; `SPEC-322`, `SPEC-323`, `SPEC-102`, `SPEC-211`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md). **Messender:** der Verifier, Host x86_64, Docker 29.10.0, 20 Kerne, fremde Last.

## Urteil

**Bestätigt mit Auflagen.** DoD 1 bis 3 sind erfüllt. DoD 4 ist zur Hälfte erfüllt: der Gate-Nachweis passt zum Stand, der Eintrag in §7 fehlt (Closure). Kein HIGH, kein MEDIUM, drei LOW, zwei INFO. Die Fix-Runde deckt die Klassen F-1, F-3, F-4, F-5 und F-7; F-2 ist für die drei Signale gedeckt, eine Achse bleibt ungeführt (V-3).

## Auflagen

1. **Planner, vor `done/`:** §7 nennt den Gate-Lauf mit Commit, Ausgabe, Laufzeit, Last und Messendem (DoD 4). §3 *Messung* nennt `make gates` am Stand `3e72b1fb`, also vor der Fix-Runde; für `6453c3e4` belegt allein der Stempel des Gate-Laufs (11:07:18), dessen Ausgabe der Verifier nicht gesehen hat. Dazu V-1.
2. **Planner, Integrationsschritt (`SPEC-102`, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)):** `harness/gegenproben.stempel` passt nicht zum Stand (`39ea6316…` liegt, `62fd0968…` folgt aus den Eingaben). `make gegenproben` ist an diesem Stand nicht gefahren, und der Befund F-8 (Stufe unter Last rot) ist offen. Der Verifier hat allein das Modul `tools/gegenproben` direkt gefahren; die übrigen Gegenproben der Stufe sind damit nicht belegt.
3. **Planner, Closure:** die Zeile in §6 *Abweichung zu F-7* nennt als Adresse den Reviewer, der Review ist aber gelaufen. Ergebnis der Verifikation unten (F-7): die gefahrene Mutation trägt Zeile 9 von `SPEC-323` Nr. 8; der Vertrag nennt `text/plain` nirgends, an ihm ist nichts zu ändern. Die übrigen Ausgänge *weiter offen* (Laufzeit, Größe, Aufrufe im Handbuch, F-8) und die Closure-Pflichten (Häkchen, §7, Register, Paarungen) stehen aus; die Aufrufe im Handbuch sind unten von Hand gefahren.
4. **Planner und fachlicher Owner, vor `Accepted`:** für den Halbsatz von `LH-RB-03` „in Docker-basierten CI-Umgebungen möglich“ gibt es keinen Lauf (Tabelle unten). Der Ersatz-Satz für den Lauf über `make` steht allein bei `LH-QA-06`. Der Plan der CI auf `main` startet erst nach `Accepted`.

## Findings

| ID | Kategorie | Befund | Quelle |
|---|---|---|---|
| V-1 | LOW | Die Messung nach `SPEC-211` Nr. 4 ist die Summe der fünf Ziele; §3 nennt `make test-image` allein. Aus Werten verschiedener Tage (rund 67 s Zuwachs der vier nach geändertem Quelltext, 20 bis 51 s für den neuen Lauf) liegt die Summe bei rund 90 bis höchstens 120 s, also nicht weit vom Trigger von 120 s; die obere Zahl zählt den Bau der Stufe `integration` doppelt. Das ist eine Rechnung, keine Messung. | `SPEC-211` Nr. 4; DoD 4 |
| V-2 | LOW | Die Tabelle in §3 führt `spec/spezifikation.md` nicht, obwohl der Zweig `SPEC-322` Nr. 4 und die Historie ändert (`b39db79d`). Kopf und §6 nennen die Änderung. | `AGENTS.md` §3.9 |
| V-3 | LOW | Das zweite Signal trifft in den Fällen allein das erste `rm` des Aufräumens. Der Mutant „`trap abbruch` vor der Schleife der `rmi` wieder scharf“ bleibt grün. Der Vertrag sagt „weitere Signale erreichen es nicht“ für den ganzen Durchgang; die Achse *Stelle im Durchgang* ist weder geführt noch als Negativ benannt. Der Code selbst ist richtig (gelesen: `trap ''` steht vor dem Aufräumen und wird nicht zurückgesetzt). | `AGENTS.md` §3.15; `SPEC-323` Nr. 6 d |
| V-4 | INFO | Der Abbruch durch ein Signal bei geschlossenem Leser hat keinen Fall (Mutant „`trap - PIPE` im Abbruch“ grün). Nr. 6 e sagt „für jede Operation“, Zeile 15 nennt grün, rot und Baufehler. | `SPEC-323` Nr. 6 e, Nr. 8 |
| V-5 | INFO | *Nicht behauptet* in `harness/README.md` nennt zum Binary-Vergleich allein „über verschiedene Hosts (Nr. 3)“; die Reichweite steht in `SPEC-323` Nr. 9 g und schließt auch den Bau aus leerem Cache und andere Bau-Dienste aus. | `SPEC-323` Nr. 9 g |

## DoD-Konformität

| DoD | Ergebnis | Beleg |
|---|---|---|
| 1 Image, `make image`, Start aus leerem Verzeichnis | erfüllt | Prüfungen am echten Image grün (eigener Lauf). Von Hand: Start ohne jede Angabe und ohne Netz aus frischem `/data` (`mode=mock`, bereit, Exit 0 beim Stopp), `/data` mit `65532/65532` und `0700`, `StopSignal` leer. `make image` nicht ausgeführt (geteiltes Tag); `make -n image` gelesen und denselben Bau unter eigenem Tag gefahren |
| 2 Neustarttest, Gate, kein Volume, Aufräumen, zweiter Bau, 20 Zeilen | erfüllt | `harness/mk/image.mk` hängt `test-image` an `GATE_CHECKS`; eigener Lauf des Treibers in einer Kopie aus `git archive HEAD` grün mit allen Zeilen bis `test-image: grün` (23,2 s); nach 14 Läufen des Treibers (der grüne, die Kontrolle, die Mutationen, die Pipe) kein Container und kein Tag eines Laufs. Tabelle und Raster in §3 und §6 sind zeilengleich mit `SPEC-323` Nr. 8 und 7. Jede der 20 Zeilen hat einen Fänger |
| 3 Handbuch-Kapitel | erfüllt | Bau, Volume, `LISTEN`, Stopp-Wartezeit, Rechte stehen; Aufrufe von Hand gefahren (unten); `handbuch-check` und `katalog-check` über den Gate-Stempel |
| 4 `make gates` grün, §7 nennt den Lauf | teilweise | Stempel: `gates-passed.head` gleich `6453c3e4`, `gates-passed.diffsha` gleich dem Hash des Arbeitsbaums (`fbacbe70…`). `make gates` nicht neu gefahren. §7 leer |
| Review | erfüllt | Report liegt vor |
| Doku-Update | erfüllt | `harness/README.md` (Gate- und Werkzeug-Zeile, *Nicht behauptet*), `SPEC-102`, `SPEC-211` Nr. 2 b und 4, *Stand* von `SPEC-323`, Index der Verträge; V-2, V-5 |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen | Sache der Closure |

**Größenmarke** (selbst gemessen, `git diff --numstat 0158092b..6453c3e4` über die Pfade des Blocks): 2026 von 2800, wie §6.

**Entscheidungen:** [ADR-0001](../plan/adr/0001-implementierungssprache-go.md): Go, Bau allein in Containern; das Binary läuft in der leeren Basis; ohne Befund. [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md): die Fälle 226 bis 236 liegen außerhalb der Gate-Kette; konform, der Pflichtlauf steht aus (Auflage 2).

## Fix-Runde gegen die Klassen des Reviews

Mutationen in Wegwerfkopien (`git archive HEAD`), je ein Lauf, an Elementen, die weder Review noch §3 so nennen. *GP* ist ein Mutant des Treibers gegen die elf Fälle in einem Container ohne Netz (`--memory=4g --cpus=4`, `sh` ist `dash`); der echte Treiber und das ganze Modul waren davor grün. *Lauf* ist der Treiber gegen den Daemon des Hosts.

| Befund | Menge | Eigene Mutation | Ergebnis |
|---|---|---|---|
| F-1 | Namen im Export, Angaben, Variablen | GP: Verzeichnisse außer `data/` und `tmp/` ausgenommen; alles unter `etc/` erlaubt; `HEALTHCHECK` der Form `NONE` erlaubt; Variablen mit `_` ausgenommen; allein Namen aus Großbuchstaben geprüft | 5 von 5 rot im Fall `image` (`inhalt-6`, `inhalt-3`, `healthcheck-none`, `variable-vorn-ADMIN_LISTEN`, `variable-vorn-data_dir`) |
| F-1 | dieselbe am echten Image | Lauf: `/etc/passwd` kopiert; `HEALTHCHECK NONE`; `ENV SERVICE_ID`; `WORKDIR /work` | 4 von 4 rot: `Inhalt` (`etc/passwd`), `HEALTHCHECK`, `Variable SERVICE_ID`, `Inhalt` (`work/`) |
| F-2 | Signal, Stelle im Aufräumen | GP: vor den `rmi` wieder scharf; Ignorieren allein bei laufendem Aufruf | beide grün: das erste ist V-3, das zweite ist praktisch gleichwertig (an jedem Ort der Fälle lief schon ein Aufruf) |
| F-3 | Wege mal Strom | GP: die Zeile `grün`, eine Zwischenmeldung und `melde` werten den Schreibfehler | 3 von 3 rot im Fall `aufraeumen` |
| F-3 | Abbruch mit geschlossenem Leser | GP: `trap - PIPE` im Abbruch | grün (V-4) |
| F-3 | am echten Daemon | Lauf unverändert, beide Ströme in `head -n 1` | Exit 0; kein Container, kein Tag, `TMPDIR` leer |
| F-4 | Stellen, die die Treiber der Gegenproben aufzählen | gelesen: Kopf und Hilfetext von `harness/mk/gegenproben.mk`, `Dockerfile`, `harness/README.md`, `SPEC-102`, `tools/gegenproben/main.go`, Eingabeliste | jede nennt den Treiber |
| F-5 | Stellen mit der veralteten Option | Suche nach `--time` in Code und lebenden Dokumenten | allein die Berichts-Zeilen des Plans; `SPEC-215` und Zeile 44 des Geschwister-Plans nachgezogen |
| F-7 | Bodies beider Exchanges, verfälscht statt gelöscht | Lauf: bei gestopptem Server den Blob der Antwort auf `/Products` bzw. den von `$metadata` mit gleicher Länge verändert, Eigentümer und Rechte erhalten; Kontrolle: derselbe Weg mit gleichen Bytes | Kontrolle grün; beide rot, Import nach dem Neustart `500 ODR-E3001` |

**Abweichung bei der Mutation zu F-7:** beide Hälften der Aussage in §6 sind bestätigt. Mit `text/plain` verwirft der Recorder den Body beim Aufzeichnen (`ODR-E2401`, Grund `typ`), die Session ist `incomplete`, die Phase `aufnehmen` rot an `close`. Mit `application/json;odata.metadata=none` ist `aufnehmen` grün und `pruefen` rot an der neuen Zeile (`read` 1, `ohne_kontext`).

**Stichproben zur Tabelle in §3** (Lauf): `STOPSIGNAL SIGKILL` rot `Stopp: Exit des Servers` (Zeile 8); `/data` mit Gruppe 0 und mit `0750` je rot `/data Eigentümer und Rechte` (Zeile 4).

## Lastenheft: was der Beleg trägt

| Wortlaut | Trägt | Trägt nicht |
|---|---|---|
| `LH-QA-06`: „läuft auf Linux als Container“ | ja: Image und Neustarttest auf einem Linux-Host | belegt für eine Architektur (x86_64) und eine Laufzeit (Docker 29.10.0) |
| `LH-QA-06`: „kann ohne verpflichtende externe Infrastruktur gestartet werden“ | ja: `--network none`, kein Volume; von Hand auch ohne jede Angabe | der Modus `record` braucht fachlich einen Upstream; im Test ist es ein Referenzdienst im Client-Container |
| `LH-QA-06`: „gleiche Konfiguration und gleiche Eingabedaten führen zu gleichem Testverhalten“ | **nicht durch diesen Slice**: der Neustarttest fährt je Gate-Lauf einmal und vergleicht keine zwei Läufe | der Beleg liegt bei `make test-determinism` (`SPEC-038`, über den Gate-Stempel), am Binary der Test-Stufe, nicht am Laufzeit-Image |
| `LH-QA-06` Messmethode: „frischem Datenverzeichnis“, „Persistenz-/Neustarttest“ | ja: davor genau eine Session und kein Workspace; nach Stopp und Start desselben Containers Exchanges, Import-Report, Workspaces, Snapshots byte-gleich, Bodies über den zweiten Import | eingehängtes Volume, neuer Container auf altem Bestand (nur von Hand: Token derselbe, Start gelingt), `SIGKILL`, `mock` und `replay` am Image, ein zweiter Neustart |
| `LH-QA-06` Messmethode: „Containerbasierter CI-Test“ | als Ersatz nach dem Satz des Lastenhefts: Lauf über `make` | es gibt keine CI; der Lauf läuft auf einem Entwickler-Host mit gefülltem Cache |
| `LH-QA-06.a`: „aus einem Container-Image mit einem Datenvolumen“ | `/data` ist `DATA_DIR`, Eigentümer und Rechte geprüft | das Volume selbst hat keinen Fänger; von Hand: benanntes Volume, Stopp, Start, Ersatz des Containers |
| `LH-RB-03` Vorgabe: „unter Linux … möglich“ | ja: Bau, Tests und Ausführung in Containern; `version` des Binary direkt auf dem Host von Hand | ein Start des Dienstes (`start`) direkt auf dem Host ist nicht gefahren; in der Gate-Kette läuft das Binary allein in Containern |
| `LH-RB-03` Vorgabe: „in Docker-basierten CI-Umgebungen möglich“ | **nein**: kein Lauf in einer CI | Auflage 4 |
| `LH-RB-03` Nachweis: „Zwei Bauten desselben Standes ergeben ein byte-gleiches Programm“ | ja, in den Grenzen von `SPEC-323` Nr. 9 g: zwei Übersetzungen, gleicher Host, Daemon, Quellstand, Basis-Images, Modul-Schicht; die zweite nachweislich neu | andere Hosts, Bau-Dienste, Plattformen, Bau aus leerem Cache; je Lauf eine Stichprobe; die erste Übersetzung darf aus einem früheren Lauf stammen |
| `LH-RB-03` Nachweis: „automatisierter End-to-End-Test auf Linux“ | ja: `make test-image` und `make test-integration` (`SPEC-037`) in der Gate-Kette | in einer CI läuft keiner |
| `LH-FA-02.c` (geordneter Shutdown, im Plan genannt) | `SIGTERM` im Image, Exit 0 | das Abwarten laufender Exchanges ist am Image nicht geprüft |

## Handbuch von Hand (Version 1.2.3, eigene Namen)

`version` gibt `odata-recorder 1.2.3`; Start mit benanntem Volume im Modus `record`; `stop` Exit 0, `start` läuft wieder; `stop -t 30` Exit 0; Token über `docker cp` (`0600`); ein neuer Container auf dem Volume übernimmt den Bestand mit demselben Token; veröffentlichter Port ohne `LISTEN` ohne Antwort, mit `LISTEN=0.0.0.0:8080` mit Antwort (502, der Upstream des Beispiels besteht nicht); `mock` mit `LISTEN` ohne `PUBLIC_BASE_URL` Exit 1 mit `ODR-E2001`. Ohne Widerspruch. Nicht gefahren: Docker Compose, eingehängte Verzeichnisse und Zertifikatsdateien, Admin-Listener außerhalb von Loopback, die Wartezeit von 10 s.

## Nicht gefahren

`make gates` (Stempel passt), `make gegenproben` und die Stufe `gegenproben` (Auftrag), `make image`, `make lint`, `make test`. Wiederholraten der Mutationen sind nicht gemessen: je Mutation ein Lauf. Load Average beim Start der Läufe 3 bis 9. Kein eigener Container, kein Volume und kein Tag ist übrig; fremde Objekte sind nicht berührt. `main` steht bei `39a21107`; ein Probe-Merge (`git merge-tree`) ist ohne Konflikt, der gemergte Stand ist nicht geprüft.
