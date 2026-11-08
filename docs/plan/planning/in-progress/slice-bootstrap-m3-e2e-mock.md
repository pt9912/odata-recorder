# Slice slice-bootstrap-m3-e2e-mock: Szenarien des Modus mock am gebauten Binary

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-12`, `LH-FA-09`, `LH-RB-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-037` Nr. 2, 3, 5 bis 7 und 10 (Bereitschaft im Modus `mock`, `reset`, Szenarien; Vertrag von `make test-integration`, `harness/vertraege/`) · `SPEC-139` Nr. 11 d und f (Neustart im Modus `mock`) · `SPEC-049` Nr. 7 a und 10 b (Neustart über die Prozessgrenze, Negativ der Snapshots) · `SPEC-110` Nr. 8 und 14 (Default `mock`, `PUBLIC_BASE_URL` aus `LISTEN`) · `SPEC-195` Nr. 5 und 8 (Zuordnung durch den Header).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-bootstrap-m3-e2e-mock`, Worktree `/Development/odata-recorder-wt-bootstrap-m3-e2e-mock`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-11-08.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung**; jeder Ausschluss trägt
seinen Grund.

**Ziel:** Das gebaute Binary im Modus `mock` ist als Stufen von `make test-integration` belegt: *Mock CRUD → Reset* (`TestE2EMock`) und *Mutation → `SIGTERM` → Start auf demselben `DATA_DIR` → Mutation lesbar* (`TestE2EMockNeustart`), dazu der Befehl `reset` des Übergangs und je Stufe die drei Zeilen im Treiber (`SPEC-037` Nr. 3).

**Übernimmt (Sendung von `slice-bootstrap-m3`, Teil E des vor dem Code geschnittenen Plans, Planner, 2026-11-08; ursprünglicher Geber `slice-mock-abnahme`, `SPEC-049` Nr. 7 a und 10 b, Rückgabe R4):** den Neustart über die Prozessgrenze im Modus `mock` für Mutationen und das Szenario *Mock CRUD → Reset*; die Zeilen E01 bis E09 der Tabelle in §3 (Entscheidung des Architect, 2026-11-08: Stufen `TestE2EMock` und `TestE2EMockNeustart`, Randformen in `SPEC-139` Nr. 11 d und f und `SPEC-037` Nr. 7 und 10 i und j). Kein Ausschluss dieses Slice trifft die Sendung. Der Geber trägt den Eintrag in seinem §1 erst nach diesem (`AGENTS.md` §3.13).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Verdrahtung des Modus `mock` im Composition Root, Konfiguration, M3-Test in `tests/integration/`, Handbuch und der Fall `leer` von `TestE2EProzess` — `slice-bootstrap-m3`; ohne dessen Binary gibt es nichts, was die Stufen fahren (Reihenfolge, kein geteilter Mechanismus: E fährt das Binary von außen).
- Snapshots über die Prozessgrenze — es gibt bis `welle-admin-api` keinen Lesepfad dafür; akzeptiertes Negativ in `SPEC-049` Nr. 10 b und `SPEC-037` Nr. 10 j (E09).
- `SIGTERM` während einer Mutation am Binary — akzeptiertes Negativ in `SPEC-037` Nr. 10 i (E08); getrennt belegt durch `SPEC-136` Nr. 15 und `SPEC-192` Nr. 7.
- Änderungen an `Makefile`, `Dockerfile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/` und am Stempel — Nr. 7 der Übergangsregeln; die Stufe `integration` baut `tests/e2e` und `tests/e2e/uebergang` als Ganzes. Braucht ein Szenario doch eine Zeile dort, ist das eine Rückgabe an den Planner, kein Code dieses Slice.
- Modus `hybrid` und Szenarien von `welle-hybrid` — `slice-bootstrap-m4`.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] Die Stufen `TestE2EMock` und `TestE2EMockNeustart` in `tests/e2e/` samt den drei Zeilen je Stufe in `tools/harness/test-integration.sh` und der Befehl `reset` des Übergangs in `tests/e2e/uebergang/` mit Test; `make test-integration` fährt beide grün (`SPEC-037`, Zeilen E01 bis E07 der Tabelle).
- [ ] Das Abdeckungs-Dokument `docs/plan/planning/e2e-abdeckung.md` ist aus `make e2e-abdeckung` neu erzeugt und führt die beiden Stufen; die Negative E08 und E09 stehen mit Grund und Folge in der Spezifikation.
- [ ] `make gates` grün; `make test-integration` als Pflichtlauf vor der Closure grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel: entfällt, kein Ergebnis dieses Slice ist am Binary für einen Anwender neu (Stufen und `reset` sind Prüfwerkzeug); das Handbuch-Gate bleibt grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tests/e2e/` | neu | Stufen `TestE2EMock` und `TestE2EMockNeustart` (`stufe_mock_test.go`); `lauf_test.go`: der Aufruf des Übergangs nimmt eine Umgebung, `importiere` übergibt `E2E_UPSTREAM_URL` (Eingabe des Übergangs, entschieden in `SPEC-037` Nr. 6) |
| `tests/e2e/uebergang/` | update | Befehl `reset` (`uebergang.go`) mit Test `TestAufrufformen` und `TestReset` (`uebergang_test.go`); `handle` liest `E2E_UPSTREAM_URL` |
| `tools/harness/test-integration.sh` | update | drei Zeilen je Stufe (`SPEC-037` Nr. 3), Stufe 8 und 9 |
| `docs/plan/planning/e2e-abdeckung.md` | update (erzeugt) | entsteht aus `make e2e-abdeckung` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster (§6) hat der Architect am 2026-11-08 vor dem ersten Code-Commit geschrieben (Zeilen E01 bis E09 des Plans von `slice-bootstrap-m3`, hier übernommen; die Randformen sind in der Spezifikation entschieden, `AGENTS.md` §3.12). Die Klasse `nebenläufig` trägt E04 (Beenden und Start des Servers); jede andere Zeile, auch ein akzeptiertes Negativ, ist `andere` (`SPEC-128` Nr. 3 e). Der Implementer sieht die Mutation vor der Übergabe rot und trägt die Raten, falls ein Fänger nur mit Wahrscheinlichkeit rot wird, in einen Absatz *Mutationen* dieser Sektion ein (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| E01 | `TestE2EMock`: Mock CRUD am gebauten Binary: Start im Modus `mock` ohne `UPSTREAM_URL` und `PUBLIC_BASE_URL`, `POST` `201` mit `Location` aus `LISTEN`, `GET`, `PATCH` `204`, `DELETE` `204`, `GET` `404` `ODR-E6008`, nach jedem Schritt der Zustand; die Collection ist `complete`; ein nicht unterstützter Ausdruck von `$filter` ist `501` (`SPEC-037` Nr. 7) | Stufe 8 des Treibers | das Binary lehnt `mock` ab; `Location` mit festem Host ⇒ rot | andere |
| E02 | Reset und Zuordnung am Binary: `SIGTERM`, `reset` des Übergangs, Start: die Collection gleicht der Antwort vor den Mutationen; Request ohne `X-Recorder-Workspace`, mit leerem Wert, mit zwei gleichen Werten und mit zwei Werten unter zwei Schreibweisen des Namens (am Binary von zwei gleichen Werten nicht zu unterscheiden, `SPEC-037` Nr. 10 l) ist je `400` `ODR-E5003` (`SPEC-195` Nr. 5), mit unbekanntem Wert `404` `ODR-E7001` | `TestE2EMock` (Schleife über die fünf Formen) | Reset ohne Wirkung ⇒ rot; `TestZuordnung` im Prozess (Form *zwei Schreibweisen*, 6 Werte) für eine `werteVon`, die nur eine Schreibweise liest; die Form „leerer Wert“ wie „fehlend“ gelesen und beantwortet mit `404` ⇒ rot | andere |
| E03 | Kein Upstream: nach dem Stoppen des Referenzdienstes zählt er in keinem Schritt eine Verbindung oder einen Request mehr (`LH-FA-12`) | **Negativ (Zähler):** der Zähler in `TestE2EMock` und `TestE2EMockNeustart` ist Kontrollwert, kein Fänger (`SPEC-037` Nr. 10 k); die Zusage fängt der Start ohne `UPSTREAM_URL` | `UPSTREAM_URL` in jedem Modus Pflicht ⇒ rot am Start (`ODR-E2001`) | andere |
| E04 | `TestE2EMockNeustart`: Mutationen (`POST`, `PATCH`, `DELETE` an drei Entities), `SIGTERM` (Exit 0), Start auf demselben `DATA_DIR`: die Zustände sind lesbar, ein zweiter Neustart ohne Änderung ebenso, jeder Prozess endet mit Exit 0; `SIGTERM` kommt erst nach der Antwort der letzten Mutation und des letzten Lesens, bereit ist das Binary im Modus `mock` mit der Verbindung ohne Request (`SPEC-037` Nr. 5 und 7, `SPEC-049` Nr. 7, `SPEC-139` Nr. 11 f) | Stufe 9 des Treibers; die Antworten vor dem Signal sind das Orakel | der Start leert `DATA_DIR`; das Beenden schließt die Transaktion nicht ab ⇒ die Mutation fehlt, rot | nebenläufig |
| E05 | `DATA_DIR` besteht nicht: der Start legt es an und ist bereit, jeder Request mit Workspace ist `404` `ODR-E7001` (`SPEC-139` Nr. 11 f) | `TestE2EMockNeustart`, Boundary | das Anlegen entfällt ⇒ der Start scheitert ⇒ rot | andere |
| E06 | Befehl `reset` des Übergangs: fehlendes oder leeres Argument Exit 2, unbekannter Workspace und die Kennung einer Session Exit 1 mit `ODR-E7001`, ein `DATA_DIR`, das nicht besteht, Exit 1 ohne es anzulegen, sonst `revision=<Zahl>` (`SPEC-037` Nr. 6) | Test der Funktion `lauf` in `tests/e2e/uebergang/` (Schleife über die Argumentformen beider Argumente); gültiger Aufruf in `TestE2EMock` | Exit-Codes vertauscht; `reset` legt das fehlende `DATA_DIR` an ⇒ rot | andere |
| E07 | Stufen im Treiber: je Stufe drei Zeilen, der Name steht in `-test.list`, ohne `PASS` ist die Stufe rot; `make e2e-abdeckung` führt Stufe 8 und 9 (`SPEC-037` Nr. 2, 3) | ein Lauf von `make test-integration`; `docs/plan/planning/e2e-abdeckung.md` aus `make e2e-abdeckung` erzeugt | eine Anker-Zeile mit falschem Namen ⇒ `nicht gefunden`, Exit 2 | andere |
| E08 | **Negativ:** `SIGTERM` während einer Mutation am Binary; Grund und Folge in `SPEC-037` Nr. 10 i und `SPEC-139` *Ungeprüft* | kein Fänger am Binary; getrennt belegt: `SPEC-136` Nr. 15 (Frist) und `SPEC-192` Nr. 7 (Atomarität) | — | andere |
| E09 | **Negativ:** Snapshots über die Prozessgrenze; Grund und Folge in `SPEC-037` Nr. 10 j und `SPEC-049` Nr. 10 b | im Prozess: `SPEC-049` Nr. 7 d | — | andere |

**Mutationen** (Implementer, 2026-11-08; `AGENTS.md` §3.10). Jede Mutation lief in einer Kopie der Quellen (`src/`, `tests/`), schreibgeschützt eingehängt, im Container ohne Netz mit `--memory 4g`; die Stufen fuhren mit `-test.count=1` in einem Prozess je Lauf. Das Maß ist der Test, der aus dem genannten Grund rot wird; der Aufbau brach in keinem Fall (jede Mutation baut).

| Zeile | Zusage | Mutation | roter Test |
|---|---|---|---|
| E01 | `Location` aus `LISTEN` | `basisAusListen` liefert `localhost` statt des Hosts aus `LISTEN` | `TestE2EMock/CRUD`: `Location` ungleich `http://127.0.0.1:<Port>/odata/Products(3)` |
| E01 | `501` mit `ODR-E6001` für Aufruf und Operatorwort in `$filter` (2 Elemente) | Status von `ODR-E6001` im Katalog 400 statt 501 | `TestE2EMock/CRUD`, beide Elemente |
| E01, E03 | Start im Modus `mock` ohne `UPSTREAM_URL` | `UPSTREAM_URL` in jedem Modus Pflicht | `TestE2EMock` und beide Teile von `TestE2EMockNeustart`: `Binary nicht bereit`, `ODR-E2001` |
| E02 | Reset wirkt | `reset` des Übergangs ruft `Get` statt `Reset` | `TestE2EMock`: Collection nach dem Reset weicht ab; `TestReset`: Revisionen 0, 0 |
| E02 | leerer Wert ist `400` (fünf Formen × fünf Requests = 25, Zählprüfung im Test) | `w[0] == ""` in `einzelwert` entfällt | `TestE2EMock/Zuordnung`: 5 Elemente der Form *leer* `404` statt `400` |
| E02 | zwei Werte sind `400` (dieselben 25) | `len(w) != 1` wird `len(w) < 1` | `TestE2EMock/Zuordnung`: *zwei gleiche* und *zwei Schreibweisen*, 10 Elemente; der `DELETE` ändert den Zustand, die Collection danach weicht ab |
| E04 | Mutationen überleben `SIGTERM` und zwei Neustarts (Orakel: Antworten vor dem Signal; 4 Lesezugriffe × 2 Neustarts = 8 Elemente) | `Herunterfahren` im Modus `mock` löscht `-wal` und `-shm`, statt die Ablage zu schließen | `TestE2EMockNeustart/Mutationen…`: alle 8 Elemente weichen ab |
| E04 | dasselbe | der zweite Start im Modus `mock` leert `DATA_DIR` (Marker-Datei) | `TestE2EMockNeustart/Mutationen…`; auch `TestE2EMock` (Reset nach dem Neustart: `ODR-E7001`) |
| E05 | fehlendes `DATA_DIR` wird angelegt | `MkdirAll` wird `Stat` in `legeDatenverzeichnisAn` | `TestE2EMockNeustart/DATA_DIR_besteht_nicht`: `Binary nicht bereit`, `ODR-E2001` |
| E06 | Exit 1 und 2 getrennt (3 Befehle × 6 Formen + 2 = 20 Elemente von `TestAufrufformen`) | die Konstanten `exitFehler` und `exitAufruf` vertauscht | `TestAufrufformen` (alle 20) |
| E06 | `reset` legt `DATA_DIR` nicht an | die Prüfung `os.Stat` in `zuruecksetzen` entfällt | `TestReset`: `reset hat das DATA_DIR angelegt` |
| E06 | leeres Argument ist Exit 2 (je Stelle) | `args[1] == ""` entfällt; getrennt `args[2] == ""` entfällt | `TestAufrufformen`, je 3 Elemente (*DATA_DIR leer* bzw. *zweites Argument leer* je Befehl) |
| E06 | ein Argument zu viel ist Exit 2 | `len(args) != 3` wird `len(args) < 3` | `TestAufrufformen`, 3 Elemente |
| E07 | der Name der Stufe steht im Test-Binary | `stufe 9 TestE2EMockNeustar` (ein Zeichen weniger) | Treiber: `Stufe 9: Test TestE2EMockNeustar nicht gefunden`, Exit 2 |
| E07 | die Anker-Zeile gehört zur Stufe | Anker `TestE2EMokNeustart` | `make e2e-abdeckung`: Exit 2 (Anker löst in der Region nicht auf) |

**Raten.** Die Zeile E04 ist die Klasse `nebenläufig`; ihr Fänger ist ein Wirkungs-Fänger (der Zustand nach dem Neustart gegen die Antworten vor dem Signal), kein Detektor: der Test ist sequenziell und `SIGTERM` kommt erst nach der letzten Antwort, ein Datenrennen im Binary sieht er nicht, das ist `make test-race` für die Pakete im Prozess. Gemessen von mir, auf dem Host mit `uptime` Last 18,4 bis 19,3 bei 20 Kernen (andere Zweige liefen): `TestE2EMockNeustart` mit der Mutation *`-wal` löschen* rot in 30 von 30 Prozessläufen, mit der Mutation *zweiter Start leert* rot in 10 von 10; ohne Mutation rot in 0 von 30 Läufen (beide Stufen in einem Prozess je Lauf). Die Zeile gilt als gefangen.

**Mengen** (`implement-slice` Schritt 14): Header-Zuordnung 5 Formen × 5 Requests = 25 (`TestE2EMock/Zuordnung` zählt), Orakel die Konstanten Status und Code je Form; Argumentformen von `uebergang` 3 Befehle × 6 Formen + 2 = 20 (`TestAufrufformen` zählt 18 plus die Schleife über 2); Neustart 4 Lesezugriffe × 2 Neustarts = 8 (Orakel: die Antworten vor dem Signal); `DATA_DIR` ohne Ablage 4 Pfade (`TestE2EMockNeustart` zählt).

**Grenzen, die ich beim Bauen fand** (Adresse in §6): (1) Der Übergang `import` löst den `@odata.nextLink` gegen die `UPSTREAM_URL` der Aufnahme auf (`SPEC-127` Nr. 12 e); sie kennt er nicht, die Kette endet in `link_nicht_aufloesbar` und die Coverage ist `partial` statt `complete` (`ODR-E6003` bei `STRICT_COVERAGE`, dem Default). Der Test übergibt sie als `E2E_UPSTREAM_URL`; `SPEC-037` Nr. 6 nennt diese Eingabe nicht. (2) Der Zähler des Referenzdienstes (E03) kann nach dem Stoppen nicht rot werden: ohne `UPSTREAM_URL` kennt das Binary keinen Upstream, und eine Verbindung zu einem geschlossenen Server zählt der Server nicht. Die Mutation *Pflicht `UPSTREAM_URL`* wird dennoch rot, am Start, nicht am Zähler. (3) *Zwei Schreibweisen* des Namens (E02) kommen beim Server von Go als ein Name mit zwei Werten an (der Server kanonisiert), und sind am Binary von *zwei gleiche Werte* nicht zu unterscheiden; eine Mutation von `werteVon`, die nur eine Schreibweise liest, fängt kein Fall am Binary (Gegenstand im Prozess: `SPEC-195` Nr. 5, `src/hexagon/services/mock/zuordnung_test.go`; ob dieser Test die Mutation fängt, habe ich nicht gemessen).

**Entschieden** (Architect, 2026-11-08, Ausgang der drei Punkte von §6): (1) `E2E_UPSTREAM_URL` bleibt, `SPEC-037` Nr. 6 nennt sie samt Alternative; (2) E03 ist akzeptiertes Negativ, `SPEC-037` Nr. 10 k; (3) E02 *zwei Schreibweisen*: Fänger im Prozess `TestZuordnung` (gelesen: `werteVon` der Mock-Dienste liest alle Schreibweisen, die Form hat 6 Zellen und fordert `ODR-E5003`; eine Mutation, die nur eine liest, ergäbe einen Wert und den Erfolg), am Binary Negativ `SPEC-037` Nr. 10 l.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): `slice-bootstrap-m3` liegt in `done/` (das Binary im Modus `mock` und der Nachzug von `SPEC-037` Nr. 5 bis 7 und 10 sowie `SPEC-139` Nr. 11 liegen damit in `main`). Beim Start liest der Planner §1 *Ausdrücklich NICHT* und die DoD der Nachbarn neu (`AGENTS.md` §3.13).

**Gate-Kette (Nr. 7 der Übergangsregeln):** Linie B. Dieser Slice ändert `tools/harness/test-integration.sh` und `tests/e2e/`, kein Pfad aus Nr. 7 (Architect, 2026-11-08, im Plan von `slice-bootstrap-m3` bestätigt: die Stufe `integration` baut `tests/e2e` und `tests/e2e/uebergang` als Ganzes). Braucht ein Szenario eine Änderung des `Dockerfile`, geht das als Rückgabe an den Planner. Stand beim Start (Planner, 2026-11-08): `slice-bootstrap-m3` liegt in `done/`; zugleich in `in-progress/` ist allein `slice-test-contract` (Kettenslice, `Makefile` und `harness/mk/`, rührt den Treiber nicht an); `slice-gate-integration-determinism`, der den Treiber hätte berühren können, ist nach `open/` zurückgestellt und liest diesen Slice beim erneuten Start. `slice-bootstrap-m4` und `slice-bootstrap-m6` führen je ein weiteres Szenario im selben Treiber; der Planner ordnet die Merges beim Start (Risiko in §6).

Größenmarke (Architect, 2026-11-08, `SPEC-128` Nr. 1; Planner-Rechnung gleich): *a* = 1 (E04), *b* = 8, *f* = 0, *p* = 40: *S* = 85 + 640 + 0 + 40 = 765, *M* = 1100, keine Schnittfrage. Der Block trägt den Teil des Architect; `basis`, `schaetzung` und `marke` hat der Planner beim Start am 2026-11-08 in eigenem Commit vor dem ersten Code-Commit gesetzt (nachgerechnet nach `SPEC-128` Nr. 1: 85·1 + 80·8 + 200·0 + 40 = 765, 135·765 = 103275, aufgerundet auf volle Hundert 1100; *S* liegt unter der Grenze 2223, keine Schnittfrage). Basis `fc6d1a6b` ist `main` vor dem Start-Commit; die Planungs-Commits dazwischen ändern allein `docs/`. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `docs/` zählt nicht.

```groessenmarke
pfad: tests/e2e/
pfad: tools/harness/test-integration.sh
produktcode: 40
fakes: 0
basis: fc6d1a6b
schaetzung: 765
marke: 1100
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt.
- `in-progress` → `open` (blockiert — Carveout?): ein Szenario braucht eine Zeile in `Dockerfile`, `Makefile` oder `harness/mk/` (Nr. 7) oder das Binary von `slice-bootstrap-m3` fehlt.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make test-integration` sind grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-08, aus dem Plan von `slice-bootstrap-m3` übernommen): jede Zelle nennt eine Nummer der Spezifikation oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Start auf vorhandenem `DATA_DIR` | `SPEC-139` Nr. 11 f und `SPEC-110` Nr. 11: ein fehlendes Verzeichnis wird angelegt, der Start gelingt | Nr. 11 f: eine Ablage ohne Baseline und ohne Metadaten ist `ODR-E7001` und `ODR-E6004` | Nr. 11 f: die Ablage eines früheren Laufs in einem anderen Modus wird übernommen | Nr. 11 d: eine Mutation beim Herunterfahren ist fertig oder hinterlässt nichts; am Binary ein Negativ (`SPEC-037` Nr. 10 i) | Nr. 11 f und `SPEC-049` Nr. 7 d: ein zweiter Neustart ohne Änderung ergibt dasselbe |
| Request mit `X-Recorder-Workspace` (Modus `mock`) | `SPEC-195` Nr. 5: `ODR-E5003` | Nr. 5: `ODR-E5003` | Nr. 8: unbekannter Wert `ODR-E7001`; Nr. 5: doppelter Wert `ODR-E5003`; ein Workspace eines anderen Service (`ODR-E5004`) gilt nicht: die Stufen führen einen Service, im Prozess von `go test` geprüft (`SPEC-195` Nr. 8) | gilt nicht: der Abbruch durch den Client ist Gegenstand des Adapters im Prozess (`SPEC-136`) | `SPEC-037` Nr. 7: derselbe Request nach dem zweiten Neustart gibt dasselbe |
| Befehl `reset`, Argument Workspace-ID | `SPEC-037` Nr. 6: ein fehlendes Argument ist Exit 2 | Nr. 6: ein leeres ebenso | Nr. 6: ein unbekannter Workspace und die Kennung einer Session sind Exit 1 mit `ODR-E7001` | gilt nicht: ein Hilfsprogramm des Tests, das Signal beendet den Prozess | `SPEC-129` Nr. 10: jeder Reset erhöht die Revision |
| Befehl `reset`, Argument `DATA_DIR` | Nr. 6: ein fehlendes Argument ist Exit 2 | Nr. 6: ein leeres ebenso | Nr. 6 und `SPEC-139` Nr. 2: ein Verzeichnis, das nicht besteht, ist Exit 1 und wird nicht angelegt | gilt nicht: wie oben | `SPEC-129` Nr. 10: wie oben |
| Stufe im Treiber (`stufe <n> <Name>`) | `SPEC-037` Nr. 2: ein Name, der nicht in `-test.list` steht, ist `nicht gefunden`, Exit 2 | Nr. 2: ein Zeichen außerhalb des Alphabets ebenso | Nr. 2: ebenso; die Zahl `<n>` ist nur Ausgabe und kein Argument des Tests (Entscheidung unten) | Nr. 2: die Frist von 120 s je Stufe | gilt nicht: der Treiber führt jede Stufe einmal je Lauf |

- Neustart über die Prozessgrenze im Modus `mock` (Sendung, `SPEC-049` Nr. 7 a): Randformen — **Ausgang:** entschieden (Architect, 2026-11-08): Verzeichnis fehlt wird angelegt (E05), Signal während einer Mutation ist ein akzeptiertes Negativ (E08), Snapshots bleiben das Negativ (E09).
- Randformen der Sendung, die der Architect am 2026-11-08 nachgezogen hat: Zeitpunkt von `SIGTERM` (erst nach der letzten Antwort, E04; das Fenster davor ist E08), Bereitschaft im Modus `mock` (Verbindung ohne Request, `SPEC-037` Nr. 5), fehlendes `DATA_DIR` am Binary (wird angelegt, E05) und am `reset` (Exit 1, E06), doppelter Neustart (E04), die Formen des Headers (E02) — **Ausgang:** entschieden, in `SPEC-037` Nr. 5 bis 7 und `SPEC-139` Nr. 11 f.
- Nummern der Stufen im Treiber `tools/harness/test-integration.sh` und das Abdeckungs-Dokument: `slice-bootstrap-m4` und `slice-bootstrap-m6` fügen im selben Treiber Stufen an. Die Zahl `<n>` in `stufe <n> <Name>` ist nur Ausgabe; der Schlüssel ist der Name des Tests (`SPEC-037` Nr. 2, `-test.list`), und `docs/plan/planning/e2e-abdeckung.md` zählt die Stufen nach der Reihenfolge im Skript und ist erzeugt. Entscheidung: dieser Slice nimmt 8 und 9 als nächste freie Zahlen; wer später in `main` merged, nummeriert seine Literale hinter die vorhandenen um und erzeugt das Dokument aus `make e2e-abdeckung` neu, statt einen Konflikt von Hand aufzulösen; die Reihenfolge der Merges ordnet der Planner. Kein Fänger gleicht die Zahl im Skript mit der Position im Dokument ab — akzeptiertes Negativ: Grund, die Zahl ist nur eine Beschriftung der Ausgabe; Folge, eine falsche Zahl irritiert den Leser der Ausgabe, belegt aber nichts Falsches. — **Ausgang:** entschieden (Architect, 2026-11-08); der Integrationsschritt liegt beim Planner (Merge-Reihenfolge, kein Slice als Nehmer).
- Die Stufe 9 beendet und startet den Server und liegt in der Klasse `nebenläufig` (E04) — **Ausgang:** offen (die Raten stehen in §3, Absatz *Raten*: 30 von 30 und 10 von 10 rot, ohne Mutation 0 von 30; den Ausgang setzt der Planner bei der Closure); Muster der Fänger wie bei den Stufen 1 bis 7 (Bereitschaft statt Schlaf).
- Der Übergang `import` braucht die `UPSTREAM_URL` der Aufnahme, damit die Coverage `complete` wird (`SPEC-127` Nr. 12 e); `SPEC-037` Nr. 6 nennt die Eingabe nicht, und E01 fordert `complete` (`SPEC-037` Nr. 7). Der Implementer hat sie als Umgebungsvariable `E2E_UPSTREAM_URL` des Hilfsprogramms eingeführt (Abweichung vom Wortlaut von Nr. 6, im Code entschieden, nicht in der Spezifikation) — **Ausgang:** entfallen (entschieden): `SPEC-037` Nr. 6 nennt die Eingabe `E2E_UPSTREAM_URL`; ein Argument ist verworfen (Gründe dort). Der Code gilt unverändert.
- Die Zeile E03 (Zähler des Referenzdienstes nach dem Stoppen) kann keine Mutation am Zähler fangen, E02 *zwei Schreibweisen* ist am Binary von *zwei gleiche Werte* nicht zu unterscheiden (§3, *Grenzen*, Punkte 2 und 3) — **Ausgang:** entfallen (entschieden): E03 ist akzeptiertes Negativ in `SPEC-037` Nr. 10 k; bei E02 fängt `TestZuordnung` im Prozess die Mutation von `werteVon`, am Binary steht das Negativ in Nr. 10 l. Auftrag an den Implementer: die Mutation *`werteVon` liest nur eine Schreibweise* (Mock-Dienst) einmal fahren und `TestZuordnung/zwei Schreibweisen` rot sehen, Zahl der roten Zellen (erwartet 6) in die Mutationstabelle.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für
Steering-Loop-Regeln (das Feld `liegt in` steht **nur**, wenn mit diesem Slice
wirklich etwas verkörpert wurde).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage nicht einzeln gesichtet; der Planner sichtet beim Start (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` ist einschlägig: die Stufen fahren das gebaute Binary).

Alle berührten Sub-Areas GF.
