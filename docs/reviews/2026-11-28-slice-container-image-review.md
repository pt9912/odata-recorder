# Review-Report: slice-container-image (Laufzeit-Image und Neustarttest) — 2026-11-28

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf.

**Gegenstand:** der eigene Anteil in `0158092b..5fd25831` ohne die Merges von `main` (Commits `0194b5d4`, `e28b3195`, `b39db79d`, `90daeec7`): `Dockerfile`, `.dockerignore`, `harness/mk/image.mk`, `tools/harness/test-image.sh`, `tests/e2e/image_test.go`, `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh`, Handbuch-Kapitel, `harness/README.md`, Nachzug in `harness/vertraege/` und Plan.

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`5fd25831`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-28

**Eingangs-Kontext:** Plan `slice-container-image`; `SPEC-322`, `SPEC-323`, `SPEC-102` (Fälle 226 bis 236), `SPEC-211`; `LH-QA-06`, `LH-RB-03`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md); `AGENTS.md` §3; Übergangsregeln Nr. 1, 2.

**Urteil: nicht ohne Nacharbeit.** Kein HIGH, 2 MEDIUM. Der Neustarttest prüft, was er sagt: er ist grün auf dem Stand, rot bei gelöschten Blobs, und er räumt im Daemon des Hosts allein Objekte des eigenen Laufs ab. Offen sind zwei Zusagen ohne Fänger.

**Vorgehen.** Läufe im Daemon des Hosts unter `flock /tmp/odata-recorder-heavy.lock`, Mutationen in Wegwerfkopien (`git archive HEAD`), eigene Objekte mit Präfix `odr-rev-` oder die des Treibers; am Ende kein eigener Container, kein eigenes Volume, kein eigenes Tag. Kein `docker build` mit eigener Speichergrenze (der Bau-Dienst nimmt keine an). Messend: Reviewer, Host mit 20 Kernen und fremder Last.

| Lauf | Ergebnis |
|---|---|
| `make test-image` auf `5fd25831` | grün, 24,0 s (Quelle und Stufe `build` im Cache; Last 4,8 vorher, 6,7 nachher) |
| L1: der Dienst löscht beim Start das Verzeichnis der Blobs | rot: `Phase pruefen`, Import der Session `500 ODR-E3001` |
| L2: Image mit `/bin/sh`, Datei in `/data`, `EXPOSE`, `HEALTHCHECK` | **grün** (F-1) |
| L3: ohne `--no-cache-filter image` | rot: `zweites Binary nicht neu übersetzt` |
| L4: Bau hängt eine Zeitmarke an das Binary | rot: `Binary der zwei Bauten ungleich` |
| Fälle 226 bis 236 gegen vier Treiber-Mutanten (Container ohne Netz, 4 CPU, 4 GiB) | Basis grün; `wait` im Abbruch entfernt: rot; `PASS` ohne Anker: rot; `trap` mit leerer Aktion im Abbruch entfernt: **grün** (F-2); Dateityp der Vertrauensanker nicht geprüft: grün, ohne Befund (Verzeichnis und Link haben Größe 0) |
| `INT` an die Prozessgruppe im zweiten Bau, `TERM` an den Treiber in `aufnehmen`, `HUP` an die Gruppe in `pruefen` | je `abgebrochen`, kein Container, kein Tag |
| Ausgabe des Treibers in eine Pipe, die der Leser schließt | Container und zwei Tags bleiben (F-3) |
| Stufe `gegenproben` des `Dockerfile`, zweimal gebaut (`--no-cache-filter gegenproben`) | baut mit `test-image.sh`; die elf Fälle `testimage/*` in beiden Läufen grün; der Lauf ist rot (F-8) |

**Handbuch, Kapitel `Container-Image`, nachgefahren** (eigenes Tag, Namen und Volumes; Software-Version 1.2.3): Start mit Volume, `stop` und `start`, `stop -t 30`, `version`, Eigentümer `65532` eines neuen Volumes, Übernahme des Bestands durch einen neuen Container, Token über `docker cp` (`0600`), Port ohne `LISTEN` ohne Antwort und mit `LISTEN` mit Antwort, `PUBLIC_BASE_URL` in `mock` und `hybrid` verlangt, `ADMIN_LISTEN` außerhalb von Loopback ohne TLS abgewiesen und mit `ADMIN_ALLOW_PLAINTEXT=true` erreichbar, nach `docker kill` offene Session `incomplete`, nicht beschreibbares Verzeichnis Ausgangswert 1. Ohne Widerspruch. Nicht gefahren: Docker Compose, eingehängte Zertifikats- und Token-Dateien, die Wartezeit von 10 s, der feste Port 8080 (stattdessen ein freier) und `make image` selbst (das geteilte Tag `odata-recorder:image` blieb unberührt; `make -n` und derselbe Bau unter eigenem Tag).

## Findings

Befunde ohne Lösungsvorschlag nach dem Schema des Skills.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `SPEC-322` Nr. 2 („Das Image trägt genau“) und Nr. 8 (kein `HEALTHCHECK`, kein `EXPOSE`) haben keinen Fänger und kein benanntes Negativ: ein Image mit Shell, Fremddatei in `/data`, `EXPOSE` und `HEALTHCHECK` ist grün (L2). Ein solches Image ginge unbemerkt durch die Gate-Kette. | `AGENTS.md` §3.15; `SPEC-322` | `tools/harness/test-image.sh` · `# Die Prüfungen am Image (SPEC-323 Nr. 3).` | ja, `make test-image` bliebe mit L2 grün | Zusage „trägt genau / trägt nicht“ ohne Fänger |
| F-2 | MEDIUM | `SPEC-323` Nr. 6 d sagt, weitere Signale erreichen das Aufräumen nach dem Abbruch nicht. Ohne die Zeile `trap '' INT HUP TERM` bleiben die elf Fälle grün (jeder sendet ein Signal); `dash` betritt den Trap dann beim zweiten Signal erneut und beginnt das Aufräumen von vorn (an einem Nachbau gesehen). | `AGENTS.md` §3.15, §3.16 | `tools/harness/test-image.sh` · `trap '' INT HUP TERM` | ja, Lauf der Fälle 226 bis 236 mit entfernter Zeile bleibt grün | Paar von Ursachen ohne Fänger |
| F-3 | LOW | Schließt der Leser der Ausgabe die Pipe, endet der Treiber an `SIGPIPE` ohne Aufräumen: gemessen blieben der Container `…-pruefung` und die Tags `test-image` und `test-image-client`. `SPEC-323` Nr. 6 nennt `INT`, `HUP`, `TERM` und `SIGKILL`; diese Randform hat keinen Ort. | `AGENTS.md` §3.12, §3.16 | `tools/harness/test-image.sh` · `trap abbruch INT HUP TERM` | ja, `make test-image` mit schließendem Leser | Randform eines Signals ohne Ort im Vertrag |
| F-4 | LOW | Kopfkommentar und Hilfetext von `make gegenproben` zählen die geprüften Treiber auf und nennen den des Neustarttests nicht; Tabelle in `harness/README.md`, `Dockerfile` und `SPEC-102` sind nachgezogen. | `AGENTS.md` §3.9 | `harness/mk/gegenproben.mk` · `die Größengrenzen von docs-check und der Treiber von lasttest` | nein | Aussage nicht an allen Stellen nachgezogen |
| F-5 | INFO | `docker stop --time` ist veraltet: Docker 29.10.0 meldet `Flag --time has been deprecated, use --timeout instead`; die Meldung bleibt im Treiber verborgen. Vertrag und Fall 226 halten die Form fest. Zuständig: Architect. | Maintainability; `SPEC-323` Nr. 4 c | `tools/harness/test-image.sh` · `rufe stop --time 30` | ja, `docker stop --time 1 <Name>` zeigt die Meldung | Veraltete Option eines Werkzeugs |
| F-6 | INFO | Jeder Lauf übersetzt ohne Cache und lässt Einträge im Bau-Cache des Hosts: ein Lauf brachte 7 Einträge und rund 230 MB (`docker system df` davor und danach; die Schicht von `go build` allein 193 MB). `SPEC-323` Nr. 1 nennt als Folge im Betrieb allein die Laufzeit. Zuständig: Architect. | Maintainability; `SPEC-323` Nr. 1 | `tools/harness/test-image.sh` · `bau --no-cache-filter build --no-cache-filter image` | ja, `docker system df` | Plattenwachstum eines Gates nicht benannt |
| F-7 | INFO | Die Antworten, die beide Phasen byte-gleich vergleichen, tragen keine Bodies; dass die aufgezeichneten Bodies den Neustart überstehen, belegt allein der Status 201 des zweiten Imports, dessen Ergebnis nicht gegen den ersten gehalten wird. Fehlende Blobs sind rot (L1). | Maintainability; `SPEC-323` Nr. 4 d | `tests/e2e/image_test.go` · `imageImport("neustart-2", w.Session)` | nein | Beleg über Status statt über Inhalt |
| F-8 | INFO | Die Stufe `gegenproben` war in zwei Läufen rot allein durch `lint/*` mit `Zeit überschritten` (9 und 5 Fälle von 236, wechselnde Menge); die Last stieg während des Laufs von 10 und 6 auf 127 bis 254. Nach Übergangsregeln Nr. 1 ist ein Rot unter Last kein Befund; ob der Lauf ohne diesen Slice anders endet, ist nicht gemessen. Der Stempel passt nicht und ist nach Nr. 2 Sache des Integrationsschritts, dessen `make gegenproben` daran scheitern kann. Zuständig: Planner. | Übergangsregeln Nr. 1, 2; [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) | `harness/gegenproben.stempel` | ja, `make gegenproben` | Gegenproben unter selbst erzeugter Last über der Fallfrist |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tests/e2e/image_test.go` (Neustart ohne Volume und ohne externe Infrastruktur: derselbe Container, `--network none`, frischer Bestand, vier Antworten, `restore`) | geprüft, ohne Befund außer F-7 |
| `Dockerfile`, `.dockerignore` (leere Basis, Nutzer `65532`, allein `DATA_DIR`, kein `VOLUME`; `image history` trägt allein `COPY`, `ENV`, `USER`, Start) | geprüft, ohne Befund außer F-1 |
| Token (Datei in einem Verzeichnis aus `mktemp -d`, Standardeingabe der Phasen; weder Argument noch Image noch Ausgabe) | geprüft, ohne Befund |
| Sicherheit am Host-Daemon (`rm`, `rmi` allein an Namen mit der Kennung des Laufs, kein `prune`, kein Entfernen nach Label) | geprüft, ohne Befund |
| Aufräumen bei Fehler und Signal (rote Läufe L1, L3, L4; drei Signale) | geprüft, ohne Befund außer F-2, F-3 |
| Binary-Vergleich samt Prüfung des Alters | geprüft, ohne Befund |
| Zeitgrenzen und Hänger (Token 30 Versuche, Bereitschaft 30 s, Phase 120 s, Stopp 30 s; `docker` selbst ohne Grenze, benannt in `SPEC-323` Nr. 5, 9 e) | geprüft, ohne Befund |
| Gate-Kosten in `make gates` (24 s) | geprüft, ohne Befund außer F-6 |
| `tools/gegenproben/`, `gegenproben-eingabe.sh` (Ersatz für `docker`, Mengen, Eingabeliste) | geprüft, ohne Befund außer F-2, F-8 |
| `harness/mk/image.mk`, `harness/README.md` | geprüft, ohne Befund außer F-4 |
| `harness/vertraege/` (`SPEC-102`, `SPEC-211`, `SPEC-323`), `spec/spezifikation.md` (Straten) | geprüft, ohne Befund außer F-5, F-6 |
| `docs/user/benutzerhandbuch.md` | geprüft, ohne Befund |
| Plan §3 (Tabelle, Mutationen), §4, §6 | geprüft, ohne Befund |
| Kommentare, Suppressions, Produktcode unter `src/` | geprüft, ohne Befund; `src/` ist im eigenen Anteil nicht berührt |

Nicht gefahren: `make gates`, `make gegenproben`, `make lint`, `make handbuch-check`.

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 2 |
| INFO | 4 |

**Finding-Klassen dieses Laufs:** Zusage „trägt genau / trägt nicht“ ohne Fänger · Paar von Ursachen ohne Fänger · Randform eines Signals ohne Ort im Vertrag · Aussage nicht an allen Stellen nachgezogen · Veraltete Option eines Werkzeugs · Plattenwachstum eines Gates nicht benannt · Beleg über Status statt über Inhalt · Gegenproben unter selbst erzeugter Last über der Fallfrist

## Verdikt

**Merge-blockierend:** ja, wegen F-1 und F-2 (je eine Zusage ohne Fänger, `AGENTS.md` §3.15); je ein Fänger oder ein benanntes Negativ mit Grund und Folge.

**Übergabe:** F-1 bis F-4 an den Implementer (F-1 und F-3 mit Rückgabe an den Architect, wenn der Vertrag den Ort braucht), F-5 bis F-7 an den Architect, F-8 an den Planner für den Integrationsschritt. Der Report ersetzt keine Verifikation.
