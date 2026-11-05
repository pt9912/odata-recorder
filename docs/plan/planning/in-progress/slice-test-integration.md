# Slice slice-test-integration: `make test-integration` — das gebaute Binary im Container, Record → Upstream stoppen → Replay

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-01`, `LH-FA-02`, `LH-FA-03`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-RB-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-037` (Vertrag `make test-integration`, `harness/vertraege/SPEC-037-test-integration.md`: *Record → Upstream stoppen → Replay; Mock CRUD → Reset*, Referenz-Upstream im Container) · `LH-RB-03.a` (Nachweis: reproduzierbarer Container-Build und End-to-End-Test) · `SPEC-049` (Fixture-Rahmen, Header- und Fehlerzeilen) · `SPEC-045` (Capture-/Replay-Fähigkeitsmatrix) · `ARC-012` · `architecture.md` §6 (*End-to-End*: Test-OData-Server starten → Requests aufzeichnen → Server stoppen → Replay).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-test-integration)

**Autor:** pt9912. **Datum:** 2026-10-31.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make test-integration` (`SPEC-037`) startet das gebaute Binary als Prozess in einem Container ohne Außennetz gegen einen Referenzdienst im selben Container, schickt echte HTTP-Requests im Modus `record`, beendet den Referenzdienst und spielt dieselben Requests im Modus `replay` ab. Die Szenarien sind die Akzeptanzkriterien des Lastenhefts (Happy Path, Boundary, Negative) von `LH-FA-01`, `LH-FA-02`, `LH-FA-03`, `LH-FA-04`, `LH-FA-05` und `LH-FA-14`; welche Zeile welches Kriterium trägt, steht in der [Rückverfolgungs-Matrix](../rueckverfolgung.md) (Teil A, Spalte *E2E-Szenario*). Es ist der erste Lauf, der die Prozessgrenze überschreitet: Binary, Startkonfiguration aus der Umgebung, Exit-Code und Signal. Den ganzen Pfad im Prozess von `go test` fährt zuerst der M2-Test aus `slice-bootstrap-m2` (`make test`); dieser Slice prüft, was dort nicht gilt, weil das Binary nicht läuft (Abgrenzung: `SPEC-037` Nr. 1). Import und Workspace-Anlage laufen im Übergang (`SPEC-037` Nr. 6), weil das Binary dafür keinen Eingang hat.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Teil *Mock CRUD → Reset* von `SPEC-037` — es gibt vor `slice-bootstrap-m3` keinen Mock-Code; das Szenario ist DoD-Unterpunkt von `slice-bootstrap-m3` (dort eingetragen, `AGENTS.md` §3.13). Dasselbe gilt für das Szenario des Verbindungszählers im Modus `hybrid` (`slice-bootstrap-m4`) und für die Szenarien über den Admin-Listener (`slice-bootstrap-m6`).
- Fehler des Speichers außer `ENOSPC` (Schreibfehler, Leseverlust) — `SPEC-037` Nr. 8 fährt den vollen Speicher (`tmpfs` von 1 MiB); die anderen Arten prüft die Persistenz-Ebene im Prozess von `go test`. Die Negative von `LH-FA-02` und `LH-FA-14` sind in diesem Slice (Speicher voll, nicht redigierbarer Body), nicht bei `slice-test-security`.
- Vergleichstest Direktzugriff gegen Proxy (`LH-QA-07`) — `slice-protokolltreue`; er läuft im Prozess von `go test` gegen den Referenzdienst von `tests/integration`, nicht gegen den von `tests/e2e` (`SPEC-037` Nr. 4).
- Neustart des Volumens im Image (`LH-QA-06`) — `slice-container-image`; 100 Wiederholungen — `slice-test-determinism`; Canary-Secrets in allen Artefakten — `slice-test-security`; Last — `slice-lasttest`.
- Import und Workspace über die Prozessgrenze, `SIGKILL` und die Wiederherstellung — `SPEC-037` Nr. 10 a und b (akzeptierte Negative mit Grund und Folge); den Rundlauf über den Admin-Listener fährt `slice-bootstrap-m6` als Stufe (dort schon in der DoD).

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

- [ ] Das Ziel `make test-integration` und der Vertrag `SPEC-037` (der Stand „entsteht mit dem Slice, der den Code liefert“ wird zum gelieferten Umfang): je Stufe ein Container mit `--network none`, darin Binary und Referenzdienst; Einbindung in die Gate-Kette oder daneben nach der Entscheidung zu §6.
- [ ] Die Szenarien Record → Upstream stoppen → Replay für die Kriterien, die die Matrix diesem Slice zuweist: `LH-FA-01` (Happy Path, Boundary, Negative), `LH-FA-02` (alle drei, Negative: Speicher voll), `LH-FA-03` (alle drei, darunter der Neustart des Prozesses auf demselben Datenverzeichnis), `LH-FA-04`, `LH-FA-05` (je alle drei) und `LH-FA-14` (alle drei, Negative: nicht redigierbarer Body und Speicher voll); jede Stufe trägt ihre Deklaration für `make e2e-abdeckung` (`SPEC-037` Nr. 3).
- [ ] Je Szenario eine Mutation, vom Implementer rot gesehen (`AGENTS.md` §3.10; Beispiele: Replay kontaktiert den Referenzdienst, ein Header mit Geheimnis erreicht die Datei, die Reihenfolge wiederholter Requests kippt).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-037`, Sensors-Tabelle in `harness/README.md`).
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
| Fragment `test-integration.mk` unter `harness/mk/` (neu), `repo.mk` | neu / update | Ziel `make test-integration` (`SPEC-037` Nr. 2); die drei Marker von `make e2e-abdeckung` (Nr. 3); das Root-Makefile bindet `harness/mk/*.mk` per Glob ein und bleibt unverändert; Bindung in der Sensors-Tabelle |
| `Dockerfile` | update | Stufe `integration` aus `source` und dem Binary der Stufe `build`: Test-Binary und Übergang (`SPEC-037` Nr. 1); das Laufzeit-Image ist `slice-container-image` |
| `tools/harness/test-integration.sh` | neu | Treiber und Quelle der Deklarationen (`SPEC-037` Nr. 2, 3) |
| `tests/e2e/` | neu | Stufen, Referenzdienst, Datensuche, Prozess-Helfer; `tests/e2e/uebergang/` das Hilfsprogramm (`SPEC-037` Nr. 4 bis 9); `tests/integration/` bleibt unberührt (`slice-protokolltreue` ändert es zugleich) |
| `e2e-abdeckung.md` unter `docs/plan/planning/` | neu (erzeugt) | Ausgabe von `make e2e-abdeckung`; kein Gate |
| `harness/vertraege/SPEC-037-test-integration.md` | update | Umfang und Randformen (mit dem Architect-Commit geliefert); der Stand wird mit dem Code gesetzt |
| `harness/README.md` | update | Sensors-Tabelle: Ziel, Vertrag, Bindung |
| `tests/e2e/fixture/` | neu | das Metadaten-Dokument des Referenzdienstes und sein Hash, von Test und Übergang gelesen (`SPEC-037` Nr. 6) |
| `.golangci.yml` | update | eine Ausnahme für `forbidigo` mit `# Why:` allein für `tests/e2e/uebergang/main.go` (der Einstieg des Hilfsprogramms verbindet den Prozess mit stdout und stderr, wie die Einstiege unter `tools/`); keine Schwelle sinkt |
| `tools/gegenproben/` | update | Fälle zur Verdrahtung, nur falls das Ziel an `GATE_CHECKS` hängt (heute ohne Fall: `harness/README.md` *Nicht behauptet*) |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15, Spalten nach `SPEC-128` Nr. 3 a). Der Fänger nennt den Test oder die benannte Ausgabe; die Stufen sind die Tests von `SPEC-037` Nr. 7. Eine Allaussage ist eine Schleife über ihre Menge, die Zahl der Elemente steht dabei. Der Implementer trägt je Zeile die rot gesehene Mutation und die Raten in einen Absatz dieses Abschnitts ein (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Die Stufen laufen ohne Außennetz (`SPEC-037` Nr. 1) | `TestE2EProzess`, Fall *außen*: `Dial` nach `192.0.2.1:80` scheitert mit `ENETUNREACH` (Fehlerart, nicht nur Fehler) | Treiber ohne `--network none`, mit `--network bridge` | andere |
| 2 | Ein Test, der in `-test.list` fehlt, ist Exit 2 mit `Test <Name> nicht gefunden` (Nr. 2) | benannte Ausgabe des Treibers; Lauf mit umbenanntem Test | Test in `tests/e2e` umbenennen, Zeile im Treiber unverändert | andere |
| 3 | Eine Stufe ohne `--- PASS` oder mit `--- SKIP` ist rot (Nr. 2) | benannte Ausgabe `kein PASS` bzw. `SKIP`; je ein Lauf mit leerem und mit übersprungenem Test | `t.Skip()` am Anfang einer Stufe; Stufe ohne Treffer des `-test.run` | andere |
| 4 | Der Treiber lehnt jedes Argument ab, auch ein leeres: `Aufruf:`, Exit 2, ohne Container (Nr. 2) | benannte Ausgabe; Läufe mit `x` und mit `''` | Argumentprüfung entfernt (zwei Läufe: unbekannt, leer) | andere |
| 5 | Jede Stufe trägt Kopfzeile, Deklaration und Anker (Nr. 3) | `make e2e-abdeckung`: Exit 0 mit 7 Stufen und 7 Deklarationen; Exit 1 mit `Stufe ohne Deklaration` oder `Deklaration ohne Stufe` | Deklaration einer Stufe gelöscht; Anker im Treiber geändert | andere |
| 6 | Eine ungültige Konfiguration ist Exit 1, `ODR-E2001` auf stderr, jeder Verstoß genannt, kein Listener (Nr. 5; `SPEC-110` Nr. 1, 3, 8, 12) | `TestE2EProzess`, Schleife über vier Umgebungen: leer, `MODE=Record`, beide Verstöße zugleich (`MODE=Record` mit `SERVICE_ROOT_PATH=/odata/`, zwei Zeilen), `MODE=record` ohne `UPSTREAM_URL` | `main` setzt `Start` nicht (Exit 2); `exitCode` liefert 0 | andere |
| 7 | `SIGTERM` beendet den Prozess mit Exit 0 binnen Frist, danach `connection refused` (Nr. 5) | `TestE2EProzess`, Fall *SIGTERM* | `Signale` in `main` meldet nichts an (Prozess endet durch das Signal) | nebenläufig |
| 8 | Status, Body und Header jeder Antwortart des Referenzdienstes erreichen den Client unverändert (`LH-FA-01`, Happy Path; Nr. 7) | `TestE2EWeiterleitung`, Schleife über 5 Antwortarten (`$metadata`, Seite, Entität mit `ETag`, `$count`, `$value`) | `ETag` entfällt bei der Entität; `Content-Type` von `text/plain` auf JSON | andere |
| 9 | Eine leere erfolgreiche Antwort bleibt erhalten (`LH-FA-01`, Boundary) | `TestE2EWeiterleitung`, Schleife über `/odata/Leer` (200) und `/odata/Nichts` (204) | leerer Body wird `{}`; 204 wird 200 | andere |
| 10 | Ein gestoppter Referenzdienst ist `502` mit `ODR-E4001` und ein Eintrag `ereignis=upstream`, `grund=nicht_erreichbar` (`LH-FA-01`, Negative; `SPEC-121` Nr. 12) | `TestE2EWeiterleitung`, Fall *Upstream gestoppt* | Status 500 statt 502; Eintrag ohne `grund` | andere |
| 11 | Ein `GET` und zwei `POST` mit verschiedenen Bodies sind im Replay je mit ihrer eigenen Antwort lesbar (`LH-FA-02`, Happy Path) | `TestE2EAufzeichnung`, Fall *Rundlauf*: 3 Requests, Antwort gleich Aufnahme, die zwei `POST` verschieden | Match-Key ignoriert den Request-Body (beide `POST` liefern dieselbe Antwort) | andere |
| 12 | Ohne Request- und Response-Body ist die Aufnahme leer und nicht fehlgeschlagen (`LH-FA-02`, Boundary) | `TestE2EAufzeichnung`, Schleife über `GET /odata/Nichts` (204) und `POST /odata/Leer` (200 leer): Replay gleich Aufnahme, nicht `ODR-E5001` | leerer Body wird als `failed` aufgezeichnet (Replay `ODR-E5001`) | andere |
| 13 | Bei vollem Speicher erhält der Client die Antwort unverändert, stderr trägt `ereignis=ablage` mit `error_code`, und der Replay dieses Requests ist `ODR-E5001` (`LH-FA-02`, Negative; Nr. 8) | `TestE2EAufzeichnung`, Fall *Speicher voll*: zwei Requests bei vollem `tmpfs`, Vorbedingung `ENOSPC` der Ballast-Datei | Fehler beim Anhängen verschluckt (kein Eintrag); Anhängen ohne Prüfung als vollständig | nebenläufig |
| 14 | Nach `SIGTERM` und Neustart auf demselben `DATA_DIR` sind alle Austausche der ersten Session im Replay gleich der Aufnahme (`LH-FA-03`, Happy Path) | `TestE2EPersistenz`, Fall *Neustart*: Schleife über alle Requests der Session (4) | zweiter Prozess mit leerem `DATA_DIR`; der Start löscht Sessions | nebenläufig |
| 15 | Eine Session ohne Request ist leer und es wird nichts erfunden (`LH-FA-03`, Boundary) | `TestE2EPersistenz`, Fall *leer*: `zeige` meldet `exchanges=0`, Replay eines Requests ist `ODR-E5001` | Replay liefert für die leere Session eine Standardantwort | nebenläufig |
| 16 | Beschädigte Daten sind Exit 1 mit `ODR-E3001`, die Datei bleibt byte-gleich (`LH-FA-03`, Negative; `SPEC-103` Nr. 6) | `TestE2EPersistenz`, Schleife über 2 Beschädigungen (Text als Datei, `NOTADB`; SQLite-Kopf mit zerstörten Seiten, `CORRUPT`); Vergleich der Bytes | Datei wird neu angelegt; Exit 0 mit Meldung | andere |
| 17 | Jeder aufgezeichnete Request ist bei gestopptem Referenzdienst gleich der Aufnahme (`LH-FA-04`, Happy Path) | `TestE2EReplay`, Schleife über alle aufgezeichneten Requests (8) | Replay liefert die Antwort des Nachbarn (Index + 1) | andere |
| 18 | Der Replay kontaktiert den Referenzdienst nie: bei laufendem Server und danach gestoppt bleibt der Zähler (Verbindungen, Requests) gleich (`LH-FA-04`; Nr. 7) | `TestE2EReplay`, Zähler vor und nach beiden Durchgängen und nach jedem Negativ-Fall | Replay leitet bei fehlendem Treffer an den Upstream weiter | nebenläufig |
| 19 | `$count` gibt in Aufnahmereihenfolge 7, 11, 13, und mit neuer Replay-Session wieder von vorn (`LH-FA-04`, Boundary) | `TestE2EReplay`, Fall *Sequenz*: zwei Replay-Sessions, je 3 Requests | Reihenfolge umgekehrt; ein Cursor für alle Replay-Sessions | andere |
| 20 | Kein Treffer ist `404` mit `ODR-E5001`; jeder der drei Steuerheader fehlend, leer oder doppelt ist `400` mit `ODR-E5003`; der vierte `$count` ist `410` mit `ODR-E5002` (`LH-FA-04`, Negative; `SPEC-135` Nr. 5) | `TestE2EReplay`, Schleife 3 Header mal 3 Formen (9) und die zwei weiteren Fälle | `X-Recorder-Replay-Session` wird nicht geprüft (fehlend ergibt 200) | andere |
| 21 | Andere Reihenfolge der Query-Parameter und andere Schreibweise der Prozent-Hexziffern sind ein Treffer (`LH-FA-05`, Happy Path; `SPEC-116`) | `TestE2EMatching`, Schleife über 2 Abweichungen | der Key hängt von der Reihenfolge ab | andere |
| 22 | `$filter=Id eq 1` und `$filter=Id eq 2` bleiben zwei Requests mit zwei Antworten (`LH-FA-05`, Boundary) | `TestE2EMatching`, Fall *Filter*: zwei Aufnahmen, zwei verschiedene Antworten im Replay | der Key lässt den Wert von `$filter` aus | andere |
| 23 | Eine ungültige Prozent-Kodierung ist mit `reject` `ODR-E5006` (400) und mit `literal` ein wörtlicher Vergleich (`LH-FA-05`, Negative) | `TestE2EMatching`, Schleife über 2 Modi (je ein Prozess) | `reject` liefert einen Treffer; `literal` gleicht `%zz` mit `%7a%7a` ab | andere |
| 24 | Kein Geheimnis steht in `DATA_DIR` oder in stdout und stderr; `GEHEIM-A` erreicht den Referenzdienst, `GEHEIM-C` den Client (`LH-FA-14`, Happy Path; Nr. 9) | `TestE2ESchutz`, Schleife über 3 Geheimnisse mal 4 Orte (Dateien vor dem Herunterfahren, danach, stdout, stderr) | `REDACT_HEADERS`-Default ohne `Authorization`; `Set-Cookie` ungeschützt abgelegt | andere |
| 25 | Die Suche findet, was da ist: ein Kontrollwert in einer Datei unter `DATA_DIR` und auf stderr wird gefunden (Nr. 9) | `TestE2ESchutz`, Fall *Kontrollwert* | Suche überspringt `-wal` und Unterverzeichnisse | andere |
| 26 | Ein Body ohne sensible Felder bleibt für den Replay erhalten (`LH-FA-14`, Boundary) | `TestE2ESchutz`, Fall *Nutzlast*: `POST` mit JSON ohne sensible Felder, Replay mit demselben Body liefert die Antwort | der Redaktor ersetzt jedes Feld | andere |
| 27 | Ein nicht redigierbarer Body lässt den Status des Clients unverändert, sein Marker steht nirgends in `DATA_DIR`, stderr trägt `error_code=ODR-E2401` mit `richtung=request` (`LH-FA-14`, Negative) | `TestE2ESchutz`, Fall *unredigierbar*: Marker in ungültigem JSON | die Redaktion fällt auf die Rohbytes zurück | andere |
| 28 | Bei vollem Speicher nach Beginn der Antwort bleibt der Status, kein Geheimnis steht vor oder nach dem Herunterfahren in einer Datei, und ein Spurwert steht nach dem Herunterfahren in keiner (`LH-FA-14`, Negative; Nr. 8 und 9) | `TestE2ESchutz`, Fall *Speicher voll*: zwei Requests, je mit `Authorization` (Geheimnis, vor und nach rot bei Fund), der zweite ein `POST` mit dem Body-Feld `password` (Geheimnis), dazu der Spurwert `X-Test-Marker` (nach dem Herunterfahren rot, davor `t.Logf`) | die Antwort wird nach dem Fehler beim Schreiben abgebrochen; `Authorization` aus `REDACT_HEADERS` entfernt (das Geheimnis liegt im `-wal`, vor dem Herunterfahren rot); Vorbedingung gestrichen (der Spurwert liegt in der Datenbank) | nebenläufig |
| 29 | Der Übergang legt Import und Workspace an, eine unbekannte Session ist Exit 1 mit `ODR-E7003`, ein fehlendes oder leeres Argument Exit 2 (Nr. 6) | `TestE2EReplay`, Fall *Übergang* (4 Aufrufe: gültig, unbekannt, fehlend, leer) | der Übergang meldet Exit 0 bei unbekannter Session | andere |

**Zusatztabelle, Zusage vom Modus abhängig** (`AGENTS.md` §3.16): die Zusage „der Upstream wird nicht kontaktiert“ gilt im Modus `replay`, die Pflicht `UPSTREAM_URL` im Modus `record` (`SPEC-110` Nr. 9).

| Modus | Zusage | Fänger |
|---|---|---|
| `replay` | kein Kontakt, auch mit gesetzter `UPSTREAM_URL` | Zeile 18 |
| `record` | `UPSTREAM_URL` ist Pflicht: ohne sie `ODR-E2001` | Zeile 6, vierte Umgebung der Schleife |

**Mutationslauf des Implementers** (`AGENTS.md` §3.10, §3.17). Je Mutation eine Wegwerfkopie des Arbeitsbaums, Bau der Stufe `integration`, die Stufe im Container (`--network none --memory 2g`); die Rate ist die Zahl der roten von *N* Prozessläufen (je ein Container, `-test.count=1`). Einen Detektor braucht keine Zeile: keine Zusage hier ist ein Datenrennen, die Fänger der nebenläufigen Zeilen (7, 13, 14, 15, 18, 28) zählen die Wirkung (Exit-Code, Zähler am Referenzdienst, Zeilen auf stderr, Bytes unter `DATA_DIR`). Unmutiert sind alle sieben Stufen je 20 Läufe grün, `Speicher voll` (Zeilen 13 und 28) in 30 Läufen je Stufe. Nach der Auflage des Architect (Zeile 28 mit Geheimnissen) unmutiert 30 von 30 Läufen grün, der Spurwert steht in 30 von 30 im `-wal` vor dem Herunterfahren, kein Geheimnis in einem. Eine Mutation zählt nur mit der Ausgabe, die ihren Grund nennt; die Tabelle nennt sie.

| Nr | Mutation (Datei) | roter Test, Ausgabe | Rate |
|---|---|---|---|
| 1 | Treiber mit `--network bridge` (`tools/harness/test-integration.sh`) | `TestE2EProzess/außen`; `Stufe 1: TestE2EProzess rot (Exit 1)`, die übrigen sechs grün | 1 Lauf |
| 2 | Test `TestE2EMatching` umbenannt, Treiber unverändert | `Stufe 6: Test TestE2EMatching nicht gefunden`, Exit 2, Stufe 7 nicht gelaufen | 1 Lauf |
| 3 | `t.Skip()` am Anfang von `TestE2EWeiterleitung`; `-test.run` des Treibers ohne Treffer (`^<Name>X$`) | `Stufe 2: TestE2EWeiterleitung rot (SKIP)`; alle sieben Stufen `rot (kein PASS)` | 1 Lauf je |
| 4 | Argumentprüfung des Treibers entfernt, Läufe mit `x` und mit `''` | statt `Aufruf:` und Exit 2 laufen alle sieben Stufen, Exit 0 (unmutiert: `Aufruf:`, Exit 2, ohne Container, beide Läufe) | 2 Läufe |
| 5 | Deklaration von `LH-FA-05` gelöscht; Anker der Deklaration `TestE2EReplay` geändert | `make e2e-abdeckung`: `Stufe ohne Deklaration`, Exit 1; `Deklaration ohne Stufe`, Exit 1 (unmutiert Exit 0 mit 7 Stufen und 7 Deklarationen) | 2 Läufe |
| 6 | `main` setzt `Start` nicht (`src/bootstrap/main/main.go`); `exitCode` liefert 0 (`src/adapters/driving/cli/cli.go`) | `TestE2EProzess/ungültige_Konfiguration`, alle vier Umgebungen: `Exit 2, erwartet 1` und `Verstöße [], erwartet [MODE]`; `Exit 0, erwartet 1` | 1 Lauf je |
| 7 | `Signale` meldet `SIGTERM` nicht an (`cli.go`) | `TestE2EProzess/SIGTERM`: `SIGTERM: Exit -1, erwartet 0` | 20 von 20 |
| 8 | `ETag` entfällt; `Content-Type` `text/plain` wird JSON; `OData-Version` entfällt, ein Header, den die Zeile nicht nennt (`src/adapters/driving/httpodata/antwort.go`) | `TestE2EWeiterleitung/<Art>`: `Header …, erwartet …` bei `Seite` und `Entität`; bei `$count` und `$value`; bei allen fünf Arten | 1 Lauf je |
| 9 | leerer Body wird `{}` (mit entferntem `Content-Length`; ohne diesen bleibt die Mutation wirkungslos, der erste Anlauf war grün); 204 wird 200 (`antwort.go`) | `TestE2EWeiterleitung/leer`: `/odata/Leer: 200 "{}", erwartet 200 ohne Body`; `/odata/Nichts: 200 "", erwartet 204 ohne Body` | 1 Lauf je |
| 10 | Status 500 statt 502; Eintrag ohne `grund` (`src/hexagon/services/record/protokoll.go`) | `TestE2EWeiterleitung/Upstream_gestoppt`: `500 …, erwartet 502 mit ODR-E4001`; `0 von 1 Zeilen auf stderr mit ["ereignis=upstream" "grund=nicht_erreichbar"]` | 1 Lauf je |
| 11 | Match-Key ohne Request-Body (`src/hexagon/model/schluessel.go`) | `TestE2EAufzeichnung/Rundlauf`: `POST B` und `POST A: Wiedergabe weicht ab: Body` (die Antworten sind vertauscht, weil der Replay in umgekehrter Reihenfolge fragt) | 1 Lauf |
| 12 | Exchange mit leerem Request- und Response-Body wird `failed` (`src/hexagon/services/record/capture/abschluss.go`) | `TestE2EAufzeichnung/leere_Aufnahme`, beide Fälle: Wiedergabe `422 ODR-E5008` statt 204 bzw. 200 | 1 Lauf |
| 13 | Fehler beim Anhängen nicht protokolliert (`ende.go`); `haengeAn` gibt bei einem Fehler von `Append` nil zurück (`anhaengen.go`) | `TestE2EAufzeichnung/Speicher_voll`: `0 von 2 Zeilen auf stderr mit ["ereignis=ablage" "error_code=ODR-"]` | 20 von 20 je |
| 14 | zweiter Prozess auf leerem `DATA_DIR` (Test); der Start im Modus `replay` löscht die Datenbank (`src/bootstrap/wurzel.go`); der Replay lässt den letzten Exchange der Session aus, ein Element, das keine andere Zeile nennt (`replay.go`) | `TestE2EPersistenz/Neustart`: alle vier Requests `404 ODR-E7003`; alle vier `404 ODR-E7003`; nur `/odata/Products/$count`: `404 ODR-E5001` | 20 von 20 je |
| 15 | Replay liefert für die leere Session eine Standardantwort `200` (`replay.go`) | `TestE2EPersistenz/leer`: `Replay /odata/Products(1): 200 "", erwartet 404 mit ODR-E5001`, ebenso `$metadata` | 20 von 20 |
| 16 | `Open` einer beschädigten Datei legt sie neu an (`wurzel.go`) | `TestE2EPersistenz/beschädigt/NOTADB` und `/CORRUPT`: `Binary endet nicht binnen 10s` | 1 Lauf |
| 17 | Replay liefert Status und Body des Nachbarn (`replay.go`) | `TestE2EReplay/Aufnahme_bei_laufendem_Server` und `…_gestopptem_Server`: `Request 0 … Request 6: Wiedergabe weicht ab: Body` | 1 Lauf |
| 18 | Replay ruft bei einem Fehler den Upstream (`wurzel.go`) | `TestE2EReplay/Fehlerfälle`: `kein Treffer: Zähler (Verbindungen, Requests) [2 9], vor dem Replay [1 8]` und jeder weitere Fall | 20 von 20 |
| 19 | Reihenfolge der Wahl umgekehrt (`auswahl.go`); ein Cursor für alle Replay-Sessions (`replay.go`) | `TestE2EReplay/Sequenz`: `$count Nr. 1: 200 "13", erwartet "7"`; `TestE2EReplay/Aufnahme_…`: `410 ODR-E5002` statt 200 | 1 Lauf je |
| 20 | `X-Recorder-Replay-Session` ungeprüft (`dienst.go`); ein doppelter Wert aller drei Header gilt (Elemente, die die Zeile nicht nennt); ein leerer Wert gilt | `TestE2EReplay/Fehlerfälle`: `X-Recorder-Replay-Session fehlend, leer, doppelt: 500, erwartet 400 mit ODR-E5003`; `X-Recorder-Workspace`, `-Recording-Session`, `-Replay-Session doppelt`; `… leer` für Workspace und Recording-Session | 1 Lauf je |
| 21 | Key hängt von der Reihenfolge der Query-Paare ab; Großschreibung der Hexziffern ist keine Ziffer mehr, ein Element, das die Zeile nicht nennt (`schluessel.go`, `redaction.go`) | `TestE2EMatching/Abweichungen`: `Reihenfolge der Query-Parameter: …Status`; `Schreibweise der Hexziffern: …Status` | 1 Lauf je |
| 22 | Key ohne den Wert von `$filter` (`schluessel.go`) | `TestE2EMatching/Filter`: `Id eq 2` und `Id eq 1: Wiedergabe weicht ab: Body` (der Replay fragt `Id eq 2` zuerst) | 1 Lauf |
| 23 | `reject` ohne Wirkung (`schluessel.go`); ungültige Hexziffern gelten als `zz` (`redaction.go`) | `TestE2EMatching/ungültige_Kodierung/reject`: `reject, %zz: 200 …`; `…/literal`: `literal, %7a%7a: 410 …, erwartet 404 mit ODR-E5001` | 1 Lauf je |
| 24 | Standardmenge von `REDACT_HEADERS` ohne `Authorization`, ohne `Set-Cookie`, ohne `Cookie` (`src/bootstrap/konfig_werte.go`) | `TestE2ESchutz/Geheimnisse`: `GEHEIM-A`, `GEHEIM-C`, `GEHEIM-B … steht in Dateien vor dem Herunterfahren` und nach dem Herunterfahren | 1 Lauf je |
| 25 | Suche überspringt `-wal`; überspringt Unterverzeichnisse; liest höchstens 1 MiB je Datei (`tests/e2e/suche_test.go`) | `TestE2ESchutz/Kontrollwert` und `/Geheimnisse`: `die Suche findet den Kontrollwert KONTROLLE-WAL-3c in kontrolle.db-wal nicht`; `…TIEF-2b in tief/unten/kontrolle.bin`; `…GROSS-5e in gross.bin` | 1 Lauf je |
| 26 | der Redaktor ersetzt jedes Feld (`src/hexagon/model/redaktor_json.go`) | `TestE2ESchutz/Nutzlast`: Wiedergabe `404 ODR-E5001` statt 201 | 1 Lauf |
| 27 | die Redaktion fällt auf die Rohbytes zurück (`src/hexagon/model/redaktor.go`) | `TestE2ESchutz/unredigierbar`: `Marker MARKER-UNREDIGIERBAR-7e3b … in blobs/sha256/…`, dazu keine Zeile `ODR-E2401` | 1 Lauf |
| 28 | die Antwort wird nach dem Fehler beim Anhängen abgebrochen (`ende.go`); `Authorization` aus der Standardmenge von `REDACT_HEADERS` entfernt (`konfig_werte.go`); Vorbedingung gestrichen: kein Ballast, das Anhängen gelingt und der Spurwert liegt in der Datenbank (Test) | `TestE2ESchutz/Speicher_voll`: `Get "http://127.0.0.1:…/odata/Products": EOF`; `Geheimnis GEHEIM-A-… und GEHEIM-D-… vor dem Herunterfahren in odata-recorder.db-wal`, nach dem Herunterfahren in `odata-recorder.db`; `… liegt nach dem Herunterfahren in einer Datei` | 30 von 30 je |
| 29 | der Übergang meldet Exit 0 bei einem Fehler (`tests/e2e/uebergang/uebergang.go`) | `TestE2EReplay/Übergang`: `unbekannte Session: Exit 0 …, erwartet 1 mit ODR-E7003` | 1 Lauf |

Zeile 28 trägt seit der Entscheidung des Architect zum Spurwert (§6, *Speicher voll und der Marker im `-wal`*) zwei Geheimnisse (Header, Body-Feld) und ihre Mutation (`Authorization` entfernt, 30 Läufe, rot vor dem Herunterfahren) neben der Mutation der Vorbedingung; der Implementer trägt beides in die Tabelle der Mutationen nach. Die Mutation der Zeile 28, *der Marker liegt in einer Teil-Datei*, hat im Produkt keinen Haken (Nr. 8 verbietet einen); die Zeile trägt stattdessen die Mutation der Vorbedingung. Ein Fänger gegen eine Teil-Datei des Produkts ist ein akzeptiertes Negativ mit Folge: bleibt nach dem Herunterfahren ein Rest, der den Marker nicht trägt, aber ihn aus einer Datei ableiten ließe, fängt ihn keine Zeile. Elemente, die die Tabelle der Mutationen nicht nannte (`AGENTS.md` §3.17), sind in den Zeilen 8, 14, 20, 21, 24 und 25 rot gesehen.

Der Block der Größenmarke (`SPEC-128` Nr. 3 h bis o); den Teil des Planner (`basis`, `schaetzung`, `marke`) trägt der Planner beim Setzen der Marke (§4).

```groessenmarke
pfad: tests/e2e
pfad: tools/harness/test-integration.sh
pfad: harness/mk/test-integration.mk
pfad: Dockerfile
pfad: repo.mk
produktcode: 330
fakes: 1
basis: 9a6fb9212d8a39cbc45177ab2c0a8aa2001e3d9a
schaetzung: 2880
marke: 3900
ersetzt: 3400 | die vorläufige Marke beim Start (Schätzung 2490) kannte weder die Tabelle Zusage-Fänger des Architect noch die Formel von SPEC-128; gezählt sind a = 6, b = 23, f = 1, p = 330, S = 2880 ergibt 3900
schnitt-grund: Der Rundlauf Record, Upstream stoppen, Replay ist der Lieferwert (Festlegung von SPEC-037); jede Hälfte ohne die andere lieferte das Ziel ohne seine Festlegung. Alle Zeilen teilen Treiber, Stufe des Dockerfile, Test-Server, Prozess-Helfer und Datensuche (der Aufbau), die Replay-Zeilen dazu den Übergang; was sich abtrennen ließe, ist eine Kette (Aufbau und Record-Zeilen, dann Übergang und Replay-Zeilen), keine Abtrennung ohne geteilten Mechanismus, und sie träfe die Start-Trigger von slice-bootstrap-m3, slice-bootstrap-m4, slice-bootstrap-m6 und slice-test-determinism. Die Überschreitung der Schnittschwelle beträgt weniger als ein Drittel, und die Zeilen sind Wiederholungen derselben Form (ein Test mit gemeinsamen Helfern).
```

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht oben, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16); beide entstanden mit dem Vertrag unter `harness/vertraege/` vor dem ersten Code-Commit (`AGENTS.md` §3.12).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-cli-adapter` liegen in `done/` (das Binary startet, hört auf Signale und bedient `record` und `replay`). Der Vertrag `SPEC-037` trägt Umfang und Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice ändert `Makefile`, `harness/mk/` und `Dockerfile` und gehört damit zur Gate-Kette: Er startet nur, wenn kein anderer Slice der Gate-Kette in `in-progress/` liegt ([Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 7). Eine Reihenfolge zu `slice-protokolltreue` gibt es nicht: dessen Vergleichstest läuft im Prozess von `go test` gegen den Referenzdienst von `tests/integration`. Geprüft beim Start am 2026-11-05: außer diesem Slice liegt in `in-progress/` allein `slice-protokolltreue` (Linie A, Pfade unter `tests/`, ändert die Gate-Kette nicht); die Zweige `adr-gate-laufzeit` und `slice-mutationslauf-werkzeug` tragen in ihrem `in-progress/` keinen Slice (`git worktree list`), `slice-mutationslauf-werkzeug` liegt in `open/` von `main`.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`): Tabelle und Block liegen in §3 (Architect, 2026-11-05). Die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks (`tests/e2e`, `tools/harness/test-integration.sh`, `test-integration.mk` unter `harness/mk/`, `Dockerfile`, `repo.mk`), Tests eingeschlossen, Text der Verträge ausgeschlossen. Der Pfad ist `tests/e2e`, nicht `tests/`: `slice-protokolltreue` ändert `tests/integration` zugleich, und die Messung zählt jede Änderung an den Pfaden (`SPEC-128` Nr. 10 g); `Dockerfile` und `repo.mk` könnte ein anderer Slice berühren, dann zählt die Messung es mit. Schätzung nach `SPEC-128` Nr. 1: *a* = 6 nebenläufige Zeilen (7, 13, 14, 15, 18, 28), *b* = 23 andere, *f* = 1 (Test-Server), *p* = 330 (Dockerfile-Stufe 25, Treiber 100 bis 140, Fragment und `repo.mk` 20, Übergang 140): *S* = 85·6 + 80·23 + 200·1 + 330 = **2880**, *M* = **3900** (das Programm meldet beide, `make groessenmarke`). *S* liegt über 2223: Schnittfrage. **Entscheidung des Architect: kein Schnitt, `schnitt-grund` im Block.** Wo sich ein Schnitt legen ließe (Aufbau mit den Record-Zeilen 1 bis 10, 13, 24, 25, 27, 28: *S* 1561 und *M* 2200; Übergang mit den Replay-Zeilen 11, 12, 14 bis 23, 26, 29: *S* 1275 und *M* 1800), ist er eine Kette, die die Start-Trigger von `slice-bootstrap-m3`, `slice-bootstrap-m4`, `slice-bootstrap-m6` und `slice-test-determinism` träfe und das Ziel ohne seine Festlegung ließe; die Entscheidung des Planner, doch zu schneiden, wäre ihm vorbehalten (`SPEC-128` Nr. 10 c). Ein vorab benannter Teil für den Schnitt beim Handoff (`AGENTS.md` §3.14) existiert nicht: jeder Teil teilt den Aufbau. Den Teil des Planner trägt der Planner vor dem ersten Code-Commit: `basis` ist `9a6fb92` (der Commit des Architect, der Vertrag, Tabelle und Raster abschließt), `schaetzung` 2880, `marke` 3900; die Zahlen 2490 und 3400 der vorläufigen Fassung sind mit Grund ersetzt (`ersetzt: 3400 | …`). Nach `AGENTS.md` §3.14 in der dort stehenden Fassung ergäbe dieselbe Tabelle (85 und 40 je Zeile, kein Produktcode für einen Nicht-Leser, 1,25, aufgerundet auf volle Hundert) 1630 · 1,25 = 2037,5, also eine Marke von 2100 und keinen Schnitt; die Abweichung ist ein offener Punkt in §6.

**Messung beim Handoff** (Implementer, Commit `4f4d347`, Basis `9a6fb92`): Summe aus Einfügungen und Löschungen von `git diff --numstat <basis>..HEAD` über die Pfade des Blocks: 1835 (1835 Einfügungen, 0 Löschungen); aufgeteilt `tests/e2e` 1681 (Tests und Helfer 1526, Übergang 123, `fixture` 32), `tools/harness/test-integration.sh` 130, `harness/mk/test-integration.mk` 11, `Dockerfile` 8, `repo.mk` 5. Das Programm meldet `Messpunkt slice-test-integration.md: zeilen=29 nebenlaeufig=6 fakes=1 produktcode=330 schaetzung=2880 marke=3900 messung=1835` und `ok, Messung 1835 von Marke 3900`. Der Produktcode des Planers (330) wurde nicht von Hand nachgemessen: die Pfade des Blocks tragen 154 Zeilen Treiber, Fragment, `repo.mk` und Dockerfile-Stufe und 123 Zeilen Übergang, zusammen 277 (Schätzung 330). Auslastung 1835 ÷ 3900 = 0,47, unter 0,60: der Implementer meldet es dem Architect für die Neurechnung der Faktoren (`SPEC-128` Nr. 9); die Messung nach dem Review gilt als Ende. **Datenpunkt des Architect (2026-11-05):** Messpunkt *Handoff*, Auslastung 0,47, unter der Grenze 0,60 von `SPEC-128` Nr. 9 — Nr. 9 misst allein am *Ende*; das Ende lag nachgemessen 17 bis 29 % über dem Handoff (`AGENTS.md` §3.14), also bei 2147 bis 2367 und einer Auslastung von 0,55 bis 0,61. Ob die Faktoren neu zu rechnen sind, entscheidet erst die Messung nach dem Review; der Architect setzt nichts um und ändert `SPEC-128` Nr. 2 nicht (Datenpunkte: *a* = 6, *b* = 23, *f* = 1, *p* = 330 schätzte 277, S = 2880, Messung 1835 am Handoff; die Auflagen aus §6 ändern die Zahl noch).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach der Tabelle in §3 setzt (Metrik und Pfade siehe oben); ein Schnitt vor dem Code ist mit `schnitt-grund` entschieden, nach dem Code gilt die Messung mit Grund (`ueberschreitung`).
- `in-progress` → `open` (blockiert): Binary und Referenzdienst lassen sich im Container ohne Außennetz nicht verbinden (entschieden in `SPEC-037` Nr. 1; ein Befund dagegen geht an den Architect).

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

- Gate-Kette oder daneben: ein Ziel in der Kette verlängert jeden Lauf von `make gates`, eines daneben verrottet, wenn niemand es ruft (Vorbild der Wahl: `make gegenproben`, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); eine Fallzeit, die mit dem Produkt wächst, hat schon einmal Gegenproben gerissen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, 2×). **Rat des Architect:** daneben, mit Pflichtlauf vor jeder Welle-Closure und bei Änderungen an `src/bootstrap` und `src/adapters/driving`; Grund: das Ziel baut ein Image und startet 7 Container hintereinander (je Stufe frisch, bis 120 s), und die Zahl der Stufen wächst mit M3, M4 und M6; Aufnahme in die Kette ist die eine Zeile `GATE_CHECKS += test-integration` im Fragment, sobald die gemessene Laufzeit unter einem Zeitbudget liegt (neue Gates sind erlaubt, `AGENTS.md` §3.6), und trägt dann Fälle der Gegenproben zur Verdrahtung. Der Vertrag (`SPEC-037` Nr. 1 bis 3) gilt in beiden Fällen und trägt die Einbindung nicht. **Stand:** gebaut daneben, das Fragment hängt nicht an `GATE_CHECKS`; ein Lauf mit warmem Cache und sieben Stufen dauert 43 s. — **Ausgang:** offen; Adresse User (Entscheidung vor dem Start des Slice; bis dahin baut der Implementer das Ziel daneben, die Umstellung ist eine Zeile).
- Formelkonflikt `AGENTS.md` §3.14 gegen `SPEC-128` Nr. 1 und 2: dieselbe Tabelle (*a* = 6, *b* = 23, *f* = 1) ergibt nach `SPEC-128` (85 je nebenläufige, 80 je andere Zeile, 200 je Fake, *p* = 330, Zuschlag 1,35) *S* = 2880 und *M* = 3900 mit Schnittfrage ab *S* ≥ 2223, nach `AGENTS.md` §3.14 (85 und 40 je Zeile, 200 je Fake, Produktcode nur für einen Leser fremder Bytes, Zuschlag 1,25, Schnitt erst über 3000) 1630 · 1,25 = 2037,5, aufgerundet **2100**, und keinen Schnitt (mit *p* = 330 wären es 1960 · 1,25 = 2450 und 2500). `SPEC-128` Nr. 1 sagt, die Faktoren stünden allein dort und kein Briefing trage eine Zahl davon; `AGENTS.md` trägt sie noch, und `make groessenmarke` rechnet nach `SPEC-128` (Rang 2 gegen Rang 8 der Source Precedence). Der Architect setzt nichts um: dieser Plan folgt dem Programm (3900). — **Ausgang:** offen; Adresse User (welche Fassung gilt; die Nachziehung von `AGENTS.md` §3.14 ist dann ein eigener Commit).
- Auslastung am Handoff 0,47 (Messung 1835 von Marke 3900): `SPEC-128` Nr. 9 verlangt bei einer Auslastung unter 0,60 am *Ende* die Neurechnung der Faktoren durch den Architect. Datenpunkt in §4 eingetragen; der Handoff ist nicht das Ende, die Messung nach dem Review entscheidet (0,55 bis 0,61 bei einem Zuwachs von 17 bis 29 %). Der Architect setzt keine Formel um; der Konflikt von `AGENTS.md` §3.14 gegen `SPEC-128` (der vorige Punkt) bleibt beim User. — **Ausgang:** offen; Adresse Architect nach der Messung am Ende (Neurechnung nur, wenn sie unter 0,60 liegt).\n- Netzlos mit zwei Servern — **Ausgang:** entfallen; entschieden in `SPEC-037` Nr. 1: ein Container je Stufe mit `--network none`, der Referenzdienst als Server im Prozess des Tests, Binary als Kindprozess, Verbindung über die Schleife; das Fehlen von `--network none` fängt Zeile 1.
- Überlappung mit dem M2-Test und mit `slice-container-image` — **Ausgang:** entfallen; abgegrenzt in `SPEC-037` Nr. 1 (*Abgrenzung*): der M2-Test fährt den Rundlauf über die Schnittstellen des Composition Root, dieser Vertrag das gebaute Binary (Umgebung, Kommando, Exit-Code, Signal, `main`); das Laufzeit-Image, das Volumen und der Neustart des Containers bleiben bei `slice-container-image`, das kein Ziel von hier übernimmt (keine Sendung, `AGENTS.md` §3.13 greift nicht).
- Fehler-Injektion (Negative von `LH-FA-02` und `LH-FA-14`) — **Ausgang:** entfallen; entschieden in `SPEC-037` Nr. 8 und 7: `ENOSPC` über ein `tmpfs` von 1 MiB mit Ballast, und der nicht redigierbare Body als Eingabefall, beides ohne Haken im Produkt. Die Matrix führt für `LH-FA-14` (Negative) noch `slice-test-security` als Träger der Fehler-Injektion; dessen Plan führt sie nicht, der Planner zieht die Matrix nach. Akzeptiertes Negativ: andere Speicherfehler als `ENOSPC` (Schreibfehler, Leseverlust) ohne Fänger am Binary; Folge: die Persistenz-Ebene im Prozess von `go test` prüft sie (`SPEC-037` Nr. 8, *Folge im Betrieb*).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster unten, jede Zelle mit Ort in `SPEC-037` oder „gilt nicht“ mit Grund — **Ausgang:** entfallen (das Raster steht vor dem ersten Code-Commit; eine Zelle ohne Ort, die der Implementer findet, geht an den Architect zurück, `AGENTS.md` §3.12).
- Akzeptierte Negative aus `SPEC-037` Nr. 10, je mit Grund und Folge dort: Import und Workspace über die Prozessgrenze (a), `SIGKILL` und zweites Signal (b), Last (c), Antwortarten außerhalb des Referenzdienstes (d), Plattform (e), Verdrahtung außer dem Netz-Fall (f); dazu die eigene Fassung des Referenzdienstes (`SPEC-037` Nr. 4) — **Ausgang:** weiter offen (Negative sind Grenzen, kein Befund; die Closure nennt sie).
- Die Gate-Kette hat einen Besitzer zugleich (Übergangsregeln Nr. 7) — **Ausgang:** entfallen; der Start am 2026-11-05 hat geprüft, dass kein Slice der Linie C in `in-progress/` liegt (§4 *Start*). Offen bleibt: ein anderer Slice der Gate-Kette startet erst nach diesem.
- Quelle der Session-ID und Weg von `zeige` (`SPEC-037` Nr. 6): der Test liest die Sessions der Ablage mit dem Adapter `sqlite` (`Sessions().List`), `zeige` liest den Status über das Handle und die Zahl mit `Recording().ListBySession`; kein Inbound-Port liefert sie. — **Ausgang:** entschieden (Architect, 2026-11-05): die Lesart des Codes gilt und steht als Entscheidung in `SPEC-037` Nr. 6 (Alternative Port oder Kommando verworfen: Haken im Produkt, Nr. 8; Folge: Fehler, die nur der Leser der Ablage zeigt, fängt der Weg nicht, Teil von Nr. 10 a). Keine Auflage.
- Session ohne Request und Import (Zeile 15, `SPEC-037` Nr. 7): der Import einer Session ohne `$metadata` ist `ODR-E6004` (`SPEC-127` Nr. 7), also entsteht aus der leeren Session kein Workspace; der Test legt ihn aus einer zweiten Session an und nennt die leere im Header; die Bereitschaft der leeren Session ist die Verbindung, die Zeilen 14 und 17 zählen das `GET $metadata` mit (4 und 8 Requests). — **Ausgang:** entschieden (Architect, 2026-11-05): die Lesart des Codes gilt; `SPEC-037` Nr. 5 (Bereitschaft, das `GET` zählt mit, Ausnahme Session ohne Austausch) und Nr. 7 (Workspace aus zweiter Session) tragen sie. Keine Auflage.
- Speicher voll und der Marker im `-wal` (Zeilen 13 und 28, Nr. 8 und 9): bei vollem `tmpfs` bleibt der Marker aus dem Request-Header in `odata-recorder.db-wal` stehen, solange der Prozess läuft (30 von 30 Läufen); nach `SIGTERM` und Löschen des Ballasts steht er in keiner Datei (30 von 30). — **Ausgang:** entschieden (Architect, 2026-11-05): **kein Befund am Produkt.** Der Marker steht im Header `X-Test-Marker`, den keine Redaktionsregel nennt (`REDACT_HEADERS` ohne ihn, `SPEC-060`); die Redaktion liegt vor der Ablage (`model.RedigiereHeader` in `Schluessel` bildet `vorab.Header`, `src/hexagon/model/schluessel.go:93`; `Exchange.RequestHeaders` übernimmt ihn, `src/hexagon/services/record/recorder.go:112`; die Antwort-Header redigiert `ende.go:66`, die Bodies der Redaktor in `capture/aufnahme.go` vor dem `CaptureWriter`), also ist der Fund im `-wal` die redigierte Darstellung eines Headers, den das Produkt auch bei gelungener Ablage speichert, und `LH-FA-14` und `LH-FA-14.a` sagen zu ihm nichts (Out-of-Scope: Erkennung ohne Konfiguration). `SPEC-037` Nr. 9 trennt jetzt *Geheimnis* (Suche vor und nach dem Herunterfahren, auch bei vollem Speicher) von *Spurwert* (nach dem Herunterfahren; davor `t.Logf`), mit Grund und Folge. Das Urteil „Produkt hält das Geheimnis bei vollem Speicher aus Dateien heraus“ trug der Test bisher nicht, weil sein Wert kein Geheimnis war. **Auflage an den Implementer:** `speicherVollSchutz` führt zwei Requests mit `Authorization` (Geheimnis) und der zweite ein `POST` mit dem Body-Feld `password`, sucht beide Geheimnisse vor und nach dem Herunterfahren (rot bei Fund) und behält den Spurwert; die Rate des Fängers in 30 Läufen und die Mutation `Authorization` aus `REDACT_HEADERS` (rot vor dem Herunterfahren) tragen §3. Zeigt dieser Lauf ein Geheimnis im `-wal`, ist es ein Befund am Produkt gegen `LH-FA-14` (Negative) und geht an den Planner und den User, nicht in den Code.
- Zeile 6, *beide Verstöße zugleich*: der Test setzt `MODE=Record` und `SERVICE_ROOT_PATH=/odata/` (zwei Zeilen `ODR-E2001`, Schlüssel `MODE` und `SERVICE_ROOT_PATH`). — **Ausgang:** entschieden (Architect, 2026-11-05): die beiden Verstöße sind `MODE=Record` (`SPEC-110` Nr. 3) und `SERVICE_ROOT_PATH=/odata/` (abschließender Schrägstrich, `SPEC-110` Nr. 7), in `SPEC-037` Nr. 7 genannt. **Auflage an den Implementer:** `stufe_prozess_test.go` Zeile 52 setzt `SERVICE_ROOT_PATH=/odata/` statt `odata/` (ein einziger Verstoß gegen eine genannte Regel statt zwei, fehlender und abschließender Schrägstrich); die Mutationen der Zeile 6 bleiben.
- Referenzdienst und `$filter` (Nr. 4): der Dienst antwortet auf `$filter=Id eq 1` und `Id eq 2` mit je einer Seite des Eintrags (`Anvil`, `Bolt`, gleich den Entitäten der anderen Routen, sonst scheitert der Import mit `ODR-E6101`). — **Ausgang:** entschieden (Architect, 2026-11-05): die Lesart des Codes gilt, `SPEC-037` Nr. 4 nennt die Route und die Konsistenz der Einträge; andere Werte von `n` sind nicht zugesagt und haben keinen Fänger (kein Szenario fragt sie; Folge: keine). Keine Auflage.
- Ausnahme in `.golangci.yml` (`AGENTS.md` §3.6): `forbidigo` meldet `os.Stdout` und `os.Stderr` im Einstieg des Übergangs; die Regel mit `# Why:` gilt allein `tests/e2e/uebergang/main.go`. — **Ausgang:** entschieden (Architect, 2026-11-05): **keine Lockerung, keine ADR.** `AGENTS.md` §3.2 legt Ausnahmen nach `exclusions.rules` mit `# Why:` und verlangt für sie nichts weiter; §3.6 verlangt eine ADR für die Senkung einer Schwelle oder Strenge, und hier sinkt nichts: `forbidigo` gilt unverändert für allen bestehenden und jeden anderen Code, die Regel nimmt eine neue Datei aus, deren Aufgabe es ist, den Prozess mit seinen Strömen zu verbinden (die Absicht der Regel, Schreiben über einen übergebenen `io.Writer`, bleibt: `uebergang.go` bekommt ihn). Sie ist dieselbe dauerhafte Einstiegs-Ausnahme wie die der Programme unter `tools/` und des Composition Root, kein Carveout auf Zeit (kein Trigger, kein Folge-Slice). Die Grenze: gilt sie mehr als dieser einen Datei (Muster, das ein Verzeichnis nimmt), ist es eine Lockerung. Keine Auflage.
- Die Tests von `tests/e2e` enden in `make test` und `make test-race` mit `SKIP` (ohne `E2E_BINARY`); gelaufen sind dort Übersetzung und `go vet`, der Treiber wertet `SKIP` als rot (Zeile 3). — **Ausgang:** entschieden (Architect, 2026-11-05): akzeptiertes Negativ mit Grund und Folge in `SPEC-037` Nr. 10 g (Grund: die Stufen brauchen das Binary und den Container ohne Netz; Folge: `SKIP` ist kein `PASS`, das Urteil trägt `make test-integration`); Abkürzung: eine Build-Marke, die die Dateien aus `make test` nähme, nähme sie auch aus `go vet` und kostete eine zweite Aufrufform der Stufe `integration`, ohne dass etwas Wiederkehrendes fehlte. Die Closure nennt es unter den Negativen. Keine Auflage.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16). Jede Zelle nennt eine Nummer von `SPEC-037` (oder `SPEC-110`, `SPEC-135`, wo `SPEC-037` auf sie zeigt) oder „gilt nicht“ mit Grund.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Treiber aufrufen (Argumente) | Nr. 2: kein Argument ist der Lauf aller Stufen | Nr. 2: leeres Argument ist `Aufruf:`, Exit 2 | Nr. 2: jedes Argument ist Exit 2 | Nr. 2: INT, TERM enden den `docker run`, Exit 2 | Nr. 2: jede Stufe in frischem Container, zwei Läufe teilen nichts |
| Stufe fahren (Name des Tests) | Nr. 2: Name nicht in `-test.list` ist Exit 2 | Nr. 2: kein `--- PASS` (leerer oder übersprungener Test) ist rot | Nr. 2: Zeichen außerhalb von `[A-Za-z0-9_]` sind „nicht gefunden“ | Nr. 2: Frist 120 s ist rot | Nr. 2: `-test.count=1`, jede Stufe einmal je Lauf |
| Binary starten (Umgebung) | Nr. 5 mit `SPEC-110` Nr. 1 und 8: leere Umgebung, Default `mock`, Exit 1 mit `ODR-E2001` | Nr. 5 mit `SPEC-110` Nr. 1: `MODE=` gilt als ungesetzt | Nr. 5 mit `SPEC-110` Nr. 3: `MODE=Record`, Exit 1; mehrere Verstöße zugleich in einer Meldung (`SPEC-110` Nr. 12) | gilt nicht: ein Signal vor dem Ende des Starts ist im Prozess von `go test` geprüft (`SPEC-138` Nr. 8 c) und lässt sich am Prozess nicht zeitlich setzen; Folge in Nr. 10 b | Nr. 5 und Nr. 7: Neustart auf demselben `DATA_DIR`, nie zwei Prozesse zugleich |
| Binary beenden (`SIGTERM`) | gilt nicht: das Signal ist fest `SIGTERM`, kein Argument | gilt nicht: dito | Nr. 10 b: `SIGKILL` und zweites Signal werden nicht gesendet | Nr. 5: hängt der Prozess über `SHUTDOWN_TIMEOUT` und 5 s hinaus, `SIGKILL` und die Stufe ist rot | gilt nicht: der Test sendet das Signal einmal; ein zweites ist Nr. 10 b |
| Referenzdienst stoppen | gilt nicht: Stopp ohne Argument | gilt nicht: dito | gilt nicht: der Test stoppt nur seinen eigenen Server | gilt nicht: der Test stoppt zwischen Requests, keiner läuft; ein Stopp unter Last ist `LH-QA-01.a` und nicht Gegenstand | Nr. 4: ein zweites `Close` ist wirkungslos |
| Übergang (Session-ID) | Nr. 6: fehlendes Argument ist Exit 2 | Nr. 6: leeres Argument ist Exit 2 | Nr. 6: unbekannte Session ist Exit 1 mit `ODR-E7003` | gilt nicht: der Test bricht den Übergang nie ab; der Aufruf ist kurz und die Ablage danach nur dem nächsten Prozess zugänglich | gilt nicht: kein Szenario ruft ihn zweimal auf derselben Ablage mit derselben Session; ein zweiter Import ist Gegenstand von `SPEC-127` im Prozess von `go test` |
| Platte füllen (Fehler-Injektion) | gilt nicht: die Größe des `tmpfs` ist fest 1 MiB | gilt nicht: dito | Nr. 8: erreicht die Ballast-Datei `ENOSPC` nicht, ist die Stufe rot | Nr. 8: der Speicherfehler trifft den Request mitten im Betrieb, Status des Clients unverändert (Zeilen 13, 28) | Nr. 8: mindestens zwei Requests bei vollem Speicher |
| Geheimnisse suchen (`DATA_DIR`, stdout, stderr) | Nr. 9: leeres Verzeichnis zählt nur mit Kontrollwert | Nr. 9: leere Datei zählt nur mit Kontrollwert | Nr. 9: `-wal`, `-shm`, Blobs und Unterverzeichnisse gehören zur Suche | gilt nicht: die Suche liest nur und läuft nach dem Ende der Prozesse oder bei stehendem Zustand | Nr. 9: vor und nach dem Herunterfahren |

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (11×, verkörpert in `AGENTS.md` §3.15; die Negative von `LH-FA-02` und `LH-FA-14` ohne E2E-Szenario sind solche Negative und brauchen Grund und Folge) und `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (2×, offen: ein Fall, der das Produkt mitfährt, wächst mit ihm; gilt für die Einbindung des Ziels, §6); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
