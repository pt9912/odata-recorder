# Slice slice-replay-speicher-und-suche: Replay unter `LH-QA-01.a` messbar machen — Ursache des Speicherverbrauchs messen, Suche nach dem Match-Key, erneuter Abnahmelauf

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md). Die Welle schließt nicht, solange kein p95 für Replay in der Größe `abnahme` vorliegt (§3 der Welle).

**Bezug:** `LH-QA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md); Carveout [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md).

**Berührte Spec-Stellen** (Architect, Vertrag vor dem Code, auf `main` seit dem Merge `6e31d761`; Commits `91beaf53` und `b986dcc2`; nach dem Review `f738ad21`, auf `main` seit `a4b79211`): `SPEC-105` Nr. 14 (`ListByMatchKey` mit der größten `sequence_no`, beide aus demselben Stand; Ordnung nach dem Lesen, keine temporäre Struktur), dazu Nr. 12 und 13 · `SPEC-135` *Ungeprüft* (Warten auf den Schreiber, Verhalten unverändert) · `SPEC-135` Nr. 8 (Kandidaten), Nr. 9 (Auswahl, `strict`), Nr. 18 (Stelle S3) und *Ungeprüft* (Speicher und Dauer als Zahl, viele Exchanges eines Keys, Zahl gleichzeitiger Requests) · `SPEC-004` (Index) · `SPEC-216` Nr. 6 g (Berichtigung des Wortlauts), Nr. 1, 12 d und 16 (der Lasttest, unverändert) · `SPEC-296` Nr. 8 (Senke, unverändert) · `LH-QA-01.a`, `LH-FA-04.a`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-replay-speicher-und-suche).

**Autor:** pt9912 (Planner, Closure von [slice-lasttest](../done/slice-lasttest.md)). **Datum:** 2026-11-26.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Unter dem Messaufbau von `LH-QA-01.a` (2 vCPU, 4 GiB, 10.000 Exchanges, 20 Requests/s) läuft der Replay-Dienst die Messung zu Ende und liefert ein p95; die Ursache des heutigen Speicherverbrauchs ist vorher gemessen und nicht vermutet.

**Ausgangslage** (Aufruf `abnahme` auf `e7bfebb4`, 2026-11-26, Verifier; Bericht `2026-11-26-slice-lasttest-verifikation`): `replay` ist in beiden Läufen `ungueltig` mit `grund=binary`; das Binary wurde je Lauf einmal vom OOM-Killer an der Grenze von 4 GiB beendet (Exit 137, `anon-rss` 4147816 kB und 4147740 kB); 13099 und 13109 von 13200 Anfragen sind Fehler, 102 und 92 Gesamtspans, bis zu 185 und 183 gleichzeitig offene Anfragen. Es gibt kein p95. **Die Ursache ist nicht belegt.** Verdacht, nicht Befund: `SPEC-135` Nr. 8 liest je Anfrage die ganze Session (`ListBySession`).

**Übernimmt:** `slice-lasttest` (liegt in `done/`): (a) den Produktbefund V-8 seiner Verifikation, (b) den aus seiner DoD genommenen Punkt *p95 des Gesamtspans für Replay in der Größe `abnahme` samt Anteil des Match-Spans*, (c) die Auflösung oder Verlängerung von [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Ein anderer Zielwert oder eine Änderung am Lastenheft — Entscheidung des Users vom 2026-11-24 ([welle-betrieb](../welle-betrieb.md) §3): eine Lockerung ist eine Änderung am Lastenheft und kein Mittel dieses Slice.
- Änderungen am Lasttest-Werkzeug (`tests/lasttest/`, `tools/harness/lasttest.sh`, `SPEC-216`) — es verhält sich wie zugesagt (Verifikation); der erneute Lauf misst mit demselben Werkzeug, damit die Läufe vergleichbar sind. Braucht die Ursachenmessung etwas daran, geht das an den Planner zurück.
- Die übrigen Aufrufer von `ListBySession` (Record, Session beenden, Inspektion, Import) — sie liegen nicht im Messpfad von `LH-QA-01`; ihr Speicherbedarf ist eigens benannt (`SPEC-124` *Ungeprüft*).
- Tuning der Latenz, wenn der erneute Lauf ein p95 über 100 ms liefert — eigener Slice nach diesem Ergebnis, unter [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md).
- Eine Begrenzung gleichzeitiger Anfragen oder eine Abweisung unter Last — sie änderte das sichtbare Verhalten des Binary und ist nicht entschieden (`SPEC-135` *Ungeprüft*; offene Frage an den User in §6, sie blockiert nicht).
- Eine Migration, ein neues Schema oder ein neuer Index — der Index `(session_id, match_key)` besteht (`make schema-check`, `make migration-check`).
- Weitere Eingriffe in den Speicher (die Kandidaten K2 bis K5 aus §3: Pool und Seiten-Cache der Verbindungen, Transaktion, HTTP-Server, `GOMEMLIMIT`), soweit sie nicht die Suche selbst sind — erklärt die Messung den Verbrauch nicht allein mit dem Lesen der Session (K1), wird die Suche dennoch geliefert, und der Rest steht unter [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) mit neuem Folge-Slice (Architect, 2026-11-27).

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] **Ursache gemessen, zuerst und am unveränderten Code** (Aufbau und Auswertung in §3, *Messung*): die Größen L, C und s stehen als Zahlen mit Methode, Last und Messendem im Plan; die Zahlen sagen, ob das Lesen der ganzen Session (K1) die Grenze von 4 GiB und die Rate von 20/s erklärt, und welcher Anteil auf anderes fällt. Vor dieser Messung entsteht keine Behebung. Derselbe Aufbau läuft nach der Behebung noch einmal.
- [ ] **Behebung:** `RecordingRepository.ListByMatchKey` (`SPEC-105` Nr. 14) und der Replay-Dienst nach `SPEC-135` Nr. 8, 9 und 18; die Zeilen R01 bis R21 der Tabelle in §3 haben je einen Fänger und eine rot gesehene Mutation (`AGENTS.md` §3.10); Strategien, Fehlercodes und Rangfolge bleiben. Sie wird auch dann geliefert, wenn K1 den Verbrauch nicht allein erklärt (Rest unter [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) mit neuem Folge-Slice).
- [ ] **Erneuter Abnahmelauf:** der Verifier fährt `make lasttest` im Profil `abnahme` mit den Vorgaben einmal auf einem Commit, der die Behebung trägt; p95 für Replay und Mock, Last, Messender und Commit stehen in der Closure-Notiz. Bei `abnahme=ja` wird [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) in das Verzeichnis `done/` der Carveouts bewegt und die Bindung in `harness/README.md` zurückgesetzt; sonst bleibt der Carveout aktiv und nennt den nächsten Folge-Slice.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-105`, `SPEC-135`).
- [ ] Führt der Slice ein sichtbares Verhalten des Binary ein: Handbuch-Kapitel im selben Slice nachgezogen; sonst sagt §7, dass kein Schlüssel, Code oder Kommando hinzukam.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen (im Wellen-Betrieb von der Closure der Welle).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort.

**Antworten des Architect** (2026-11-27, vom Planner beim Start aufgenommen; der Vertrag steht auf `main`, Kopf): ein Slice, Messung und Behebung bleiben zusammen. Der Index `(session_id, match_key)` besteht (`idx_exchange_match`, Fänger `make schema-check` und `make migration-check`); keine Migration. Die Spezifikation sagt zum Speicher allein, dass er mit der Zahl der Treffer wächst (`SPEC-105` Nr. 14); Speicher und Dauer als Zahl, viele Exchanges eines Keys und die Zahl gleichzeitiger Requests stehen in `SPEC-135` *Ungeprüft*. Die Tabelle, das Raster (§6) und die Mengen unten hat der Planner aus den Antworten und dem Wortlaut des Vertrags gesetzt; trägt eine Zelle den Fall nicht, hält der Implementer an und gibt sie an den Architect zurück (`AGENTS.md` §3.12).

### Messung (DoD 1)

- **Wann:** zuerst, am unveränderten Code, vor jeder Behebung; derselbe Aufbau läuft nach der Behebung noch einmal.
- **Aufbau:** ein Testprogramm ruft `replay.Replayer.Replay` direkt über die echte SQLite-Ablage und die Blob-Ablage, mit einer Recording-Session von 10.000 Exchanges; netzlos im Container mit `--cpus=2 --memory=4g --memory-swap=4g`, in einer Wegwerfkopie des Repos. Es wird nicht committet und zählt nicht in der Größenmarke; Zahlen, Methode, Last des Hosts und Messender stehen in diesem Abschnitt.
- **Größen:** *L* = gehaltener Speicher je Anfrage (nach einem GC: `HeapAlloc` bei gehaltener Liste, dazu `TotalAlloc`); *C* = CPU-Zeit je Anfrage (`Getrusage` über 50 Anfragen); *s* = Zuwachs je gleichzeitiger Anfrage für *k* = 1, 8, 32 und 64, je *k* ein eigener Prozess (`VmHWM`, `HeapSys`).
- **Auswertung:** *s* × 185 (die höchste Zahl gleichzeitig offener Anfragen des Aufrufs `abnahme`) gegen die rund 3,96 GiB (4,25 GB), bei denen der OOM-Killer das Binary beendete (`anon-rss` 4147816 kB, `kB` des Kernels gleich KiB); *C* × 20 je Sekunde gegen die zwei vCPU.
- **Kandidaten**, keiner belegt: K1 ist das Lesen der ganzen Session je Anfrage (`src/hexagon/services/replay/replay.go:118`; `src/adapters/driven/sqlite/recording.go:121`, mit den Stellen der Dekodierung `:151`, `:179`, `:183`). Die weiteren Stellen der Kandidaten K2 bis K5: `src/adapters/driven/sqlite/tx.go:56` (die Transaktion), `src/adapters/driving/httpodata/server.go:104` (der Server ohne Grenze gleichzeitiger Requests), `src/adapters/driven/sqlite/sqlite.go:81` (der Pool ohne Grenze der Verbindungen) und die fehlende Setzung `GOMEMLIMIT`. Zeilennummern am Stand `0ee80d43`.
- **Folge:** erklärt K1 den Verbrauch nicht allein, wird die Suche dennoch geliefert; der Rest steht unter [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) mit neuem Folge-Slice (§4).

#### Ergebnis am unveränderten Code (vor jeder Behebung)

**Messender:** Implementer, 2026-11-27, Commit `17086a9a`. **Methode:** der Aufbau von `tests/lasttest` (Profil `abnahme`, Startwert 1: Record über `bootstrap.Build`, Import, Workspace; 10 000 Exchanges, 101 MiB in `DATA_DIR`) und ein Messprogramm, das `replay.Replayer.Replay` mit der Konfiguration des Composition Root über `sqlite` und `blobs` ruft (Einzel-Lookups, je Aufruf eine neue Replay-Session, Header `Accept`, `Accept-Encoding`, `User-Agent` wie beim Client des Lasttests); go 1.27.1, `GOMAXPROCS` 2, ohne `GOMEMLIMIT` und `GOGC`. **Container:** netzlos, `--cpus=2 --memory=4g --memory-swap=4g`, ein Container zur Zeit. **Host:** 20 CPU, 31 GB, fremde Datenbank-Container; Last (1 min) während der zwei Durchgänge 7,2 bis 4,9.

| Größe | Wie | Wert |
|---|---|---|
| *L* gehaltener Speicher | `HeapAlloc` nach `runtime.GC` beim Eintritt in die Transaktion (der Dienst hält das Gelesene), Zuwachs gegen den Stand vor dem Aufruf; 10 Aufrufe in zwei Prozessen | 15,60 MiB (15,595 bis 15,609) |
| *L* `TotalAlloc` | je `Replay` | 45,9 MiB in rund 708 000 Allokationen |
| *C* | `Getrusage` über 50 Aufrufe nacheinander, zwei Prozesse | 107,5 und 106,1 ms CPU je Anfrage (Wand 96,3 und 95,1 ms) |
| *s*, gemeinsamer Start | `VmHWM` minus `VmRSS` beim Start, geteilt durch *k*; je *k* ein Prozess, zwei Durchgänge | `VmHWM` bei *k* = 1: 42,6 und 42,2 MiB; 8: 226,7 und 219,5; 32: 883,2 und 848,3; 64: 1550,5 und 1770,2 (Start 13,2 bis 13,5); *s* bei *k* = 8 bis 64: 24,0 bis 27,4 MiB |
| *s*, alle *k* zugleich gehalten (Zusatz: die Transaktion beginnt erst, wenn alle *k* gelesen haben) | wie oben | `VmHWM` 8: 248,3 und 244,0 MiB; 32: 909,1 und 907,5; 64: 1726,4 und 1684,5; gehaltener Heap 125, 500 und 999 MiB; *s* 26,1 bis 29,4 MiB |

**Auswertung.**

- **Speicher:** *s* × 185 = 4,3 bis 5,3 GiB (4,7 bis 5,7 GB) und liegt über den rund 3,96 GiB (4,25 GB), bei denen der OOM-Killer das Binary beendete. Allein der gehaltene Heap (15,6 MiB × 185 = 2,8 GiB) liegt darunter; der Rest bis `VmHWM` ist die Reserve des GC und der Speicher von SQLite.
- **CPU:** *C* × 20 je Sekunde = 2,1 CPU-Sekunden je Sekunde, mehr als die zwei vCPU hergeben. Bei der Rate des Lasttests läuft ein Rückstand an Anfragen auf; so entstehen die offenen Anfragen, deren Speicher sich summiert.
- **K1** (Lesen der ganzen Session je Anfrage) erklärt demnach beide Grenzen. Welcher Anteil je Anfrage auf anderes fällt (K2 bis K5), zeigt derselbe Aufbau nach der Behebung.

#### Ergebnis nach der Behebung (derselbe Aufbau)

**Messender:** Implementer, 2026-11-27, Quellstand des Handoff-Commits (Produktcode), dieselben Daten (Kopie des Stands vor der ersten Messung), dieselben Bedingungen; Last (1 min) 8,6 bis 7,6.

| Größe | vorher | nachher |
|---|---|---|
| *L* gehaltener Speicher | 15,60 MiB | unter der Auflösung der Messung (−0,014 bis −0,006 MiB); gehaltener Heap bei 64 zugleich 0,4 und 0,5 MiB |
| *L* `TotalAlloc` je `Replay` | 45,9 MiB, rund 708 000 Allokationen | 0,055 bis 0,063 MiB, 564 bis 585 Allokationen |
| *C* CPU je Anfrage | 107,5 und 106,1 ms (Wand 96,3 und 95,1) | 0,76 und 0,78 ms (Wand 10,8 und 12,5) |
| `VmHWM`, gemeinsamer Start, *k* = 1, 8, 32, 64 | 42, 220 bis 227, 848 bis 883, 1550 bis 1770 MiB | 14,3; 16,9 und 17,0; 23,1 und 23,6; 29,5 und 32,3 MiB |
| `VmHWM`, alle *k* zugleich gehalten | 42 bis 43, 244 bis 248, 908 bis 909, 1685 bis 1726 MiB | 14,7; 16,8 und 17,0; 20,9 und 21,3; 26,0 und 27,1 MiB |
| *s* je gleichzeitiger Anfrage (*k* = 8 bis 64) | 24,0 bis 29,4 MiB | 0,20 bis 0,46 MiB |
| *s* × 185 | 4,3 bis 5,3 GiB | höchstens 85 MiB |
| *C* × 20 je Sekunde | 2,1 CPU-Sekunden | 0,016 CPU-Sekunden |

**Anteil des Rests:** von *s* und von *C* fällt höchstens ein Fünfzigstel auf anderes als das Lesen der Session (K2 bis K5). K1 erklärt den Verbrauch; ein Rest unter [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) aus dem Speicher ist in diesem Aufbau nicht zu sehen. Der Aufbau fährt den Dienst direkt: HTTP-Server, Telemetrie-Adapter und die Rate des Lasttests gehören nicht dazu, ein p95 liefert allein der Aufruf `abnahme`.

**Nach der Fix-Runde** (Ordnung im Adapter statt in der Abfrage; derselbe Aufbau, frische Kopie derselben Daten, Messender Implementer, Last (1 min) 7,9 bis 6,1): *C* 0,72 und 0,67 ms CPU je Anfrage (Wand 9,6 ms), `TotalAlloc` 0,055 bis 0,060 MiB in 564 bis 585 Allokationen, gehaltener Zuwachs unter der Auflösung; `VmHWM` bei *k* = 1, 8, 32, 64 mit gemeinsamem Start 14,1; 16,7 und 16,9; 23,4 und 23,8; 28,9 und 31,2 MiB; *s* 0,19 bis 0,50 MiB, *s* × 185 höchstens 93 MiB. Die Zahlen bleiben damit in der Streuung der Messung davor. **Sie gelten für einen Treffer je Key** (die Daten des Lasttests, Review I-2): Sortieren und Speicher bei vielen Treffern sind nicht gemessen (`SPEC-135` *Ungeprüft*).

**Nebenbefund, außerhalb des Slice** (§6): bei *k* = 64 zugleich freigegebenen Transaktionen enden einzelne Aufrufe mit `ODR-E9001`, Ursache `database is locked (5) (SQLITE_BUSY)` nach rund 5 s; vorher in 2 von 3 Läufen (6 und 4 Aufrufe von 64, ohne Ursache protokolliert), nachher in 1 von 2 Läufen (2 von 64), nach der Fix-Runde in 1 von 2 Läufen (4 von 64). Auch die Wandzeit bei *k* = 32 (rund 2,2 s bei 0,09 s CPU) stammt aus dem Warten der Schreiber.

### Dateien

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/outbound/recording_repository.go` | update | `ListByMatchKey` (`SPEC-105` Nr. 14) |
| `src/adapters/driven/sqlite/recording.go` | update | die Operation auf dem Index, beide Werte aus demselben Stand |
| `src/hexagon/services/replay/` (`replay.go`, `auswahl.go`) | update | Suche nach dem Key, `strict` mit der größten Nummer (`SPEC-135` Nr. 8, 9, 18) |
| `src/hexagon/ports/outbound/outboundtest/vertrag.go`; `src/adapters/driven/sqlite/` (Tests `export_test.go`, neu `list_by_match_key_test.go`; `repositories_test.go` bleibt unverändert) | update und neu | Port-Vertrag, Dekodierung, Index, Speicher, Fenster, Abbruch an jedem Schritt |
| `src/hexagon/services/replay/` (Tests) | update und neu | der Fake `exchanges` liefert echt (Filter nach Key, größte Nummer) und meldet jeden Aufruf von `ListBySession` als Verstoß; die bestehenden Tests, die `Exchanges.ListBySession` nennen, ziehen auf die neue Operation um, keine Zelle und kein Fänger entfällt (`AGENTS.md` §3.6, `BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger`) |
| `src/hexagon/services/record/`, `session/`, `inspection/`, `baselineimport/` (Tests der Fakes) | update | Stub der neuen Operation; er zählt den Aufruf wie jede andere Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); in `record` auch der Wrapper `jExchanges` |
| `src/bootstrap/telemetrie_replay_test.go` | update | der Wrapper `ursacheExchanges` setzt den Fehler an der neuen Operation ein |

Erwartung des Architect: Produktcode grob 80 bis 120 Zeilen, Tests grob 300 bis 450.

### Zusage · Fänger

(`AGENTS.md` §3.15; die Zeilen R01 bis R11 sind der Port samt Adapter, R12 bis R20 der Dienst, R21 der Composition Root.) Die Zahl der Zellen einer Schleife steht im Test als Literal und wird gegen die gezählten verglichen; kein Orakel kommt aus dem Code. Der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| R01 | (`SPEC-105` Nr. 14, 12) Geliefert werden genau die Exchanges der Session mit byte-gleichem `match_key`, nach `sequence_no` aufsteigend, jedes Feld wie bei `Get` | neuer Fall in `RunRecordingVertrag`: drei Treffer zwischen Exchanges anderer Keys, Kennungen gegen die Reihenfolge der Nummern, Nachbar-Keys (Präfix, Verlängerung, andere Schreibweise); je Treffer Gleichheit mit `Get`. Die Ordnung (der Adapter ordnet nach dem Lesen, `SPEC-105` Nr. 14) hält ein Adaptertest, in dessen Datei die Folge der Zeilen von der Folge der Nummern abweicht (`TestListByMatchKeyOrdnetNachDerNummer`) | ohne Bedingung auf `match_key`; Ordnung nach der Kennung; ohne Sortierung; absteigend; Präfix statt Gleichheit; Vergleich ohne Beachtung der Schreibweise; ein Feld nicht gelesen | andere |
| R02 | (Nr. 14) der zweite Wert ist die größte `sequence_no` der Session, unabhängig vom Key; `0` ohne Exchanges | derselbe Vertrag: die größte Nummer liegt auf einem fremden Key; ein Key ohne Treffer; eine Session ohne Exchanges | Maximum allein über die Treffer; Zahl der Treffer; Zahl der Exchanges der Treffer plus eins | andere |
| R03 | (Nr. 14) eine unbekannte Session, auch die leere Kennung, ist `ErrNichtGefunden`; eine Session ohne Exchanges ist eine leere Liste mit `0` und kein Fehler | derselbe Vertrag, beide Kennungen und die leere Session | keine Prüfung der Existenz; Fehler auch bei der Session ohne Exchanges | andere |
| R04 | (Nr. 14) ein leerer `match_key` trifft keinen Exchange, auch keinen ohne Key, und ist kein Fehler; ein Key ohne Treffer ist eine leere Liste | derselbe Vertrag: Session mit Exchanges mit und ohne Key | der leere Key als Vergleich mit dem fehlenden Wert gelesen; `ErrNichtGefunden` ohne Treffer | andere |
| R05 | (Nr. 14) Exchanges anderer Sessions mit demselben Key werden nie geliefert, ihre Nummern zählen nicht | zweite Session mit demselben Key und höherer Nummer | ohne Bedingung auf die Session an der Liste; ohne Bedingung auf die Session am Maximum | andere |
| R06 | (Nr. 13) ein abgebrochener Kontext endet mit dem Fehler des Kontexts, auch bei unbekannter Session | `recordingFall9` erweitert, dazu die unbekannte Session mit abgebrochenem Kontext | der Kontext geht nicht an die Abfrage; der Fehler des Kontexts als leere Liste; `ErrNichtGefunden` vor dem Kontext | andere |
| R07 | (Nr. 14) beide Werte stammen aus demselben Stand: ein gleichzeitiges `Append` ist in beiden ganz oder gar nicht enthalten | Adaptertest mit Haltepunkt am `ctx` (die Technik von `transaktion_haltepunkt_test.go`): Schleife über jeden Aufruf von `Err`, an dem der Test über eine zweite Verbindung anhängt, je mit gleichem und mit fremdem Key; Orakel: Liste und Maximum sind der Stand davor oder danach, nie gemischt. Ein Fenster, in das kein Aufruf von `Err` fällt, steht als akzeptiertes Negativ mit Grund, Folge und Maß in diesem Abschnitt (`AGENTS.md` §3.10) | zwei Abfragen ohne gemeinsamen Stand (rot in jedem Lauf, sonst Negativ) | nebenläufig |
| R08 | (Nr. 14) ein gleichzeitiges `Delete` der Session ist in beiden ganz oder gar nicht enthalten | wie R07, je Aufruf von `Err` ein `Delete` der Session; Orakel: der volle Stand oder `ErrNichtGefunden`, nie Treffer mit `0` und nie eine leere Liste mit Maximum | Prüfung der Existenz nach den Abfragen in eigenem Stand | nebenläufig |
| R09 | (Nr. 14) die Operation dekodiert allein die gelieferten Exchanges | Adaptertest: `request_headers_json` eines Exchange mit anderem Key derselben Session ist per SQL unlesbar gesetzt; `ListByMatchKey` gelingt, `ListBySession` scheitert (Nachweis, dass die Zeile unlesbar ist) | die ganze Session lesen und im Adapter filtern | andere |
| R10 | (Nr. 14, `SPEC-004`) die Treffer werden über den Index gefunden, und keine Abfrage lässt die Datenbank sortieren oder eine temporäre Struktur anlegen | Adaptertest: `EXPLAIN QUERY PLAN` jeder Abfrage der Operation (Text über `export_test.go`) steht als Literal: die Treffer über `idx_exchange_match` nach Session und Key, die größte Nummer nach der Session; kein `USE TEMP`, kein `AUTOMATIC` | Index umgangen; Filter nach dem Key im Adapter statt in der Abfrage; `ORDER BY` zurück, mit und ohne `INDEXED BY` | andere |
| R11 | (Nr. 14) der Speicher wächst mit der Zahl der Treffer, nicht mit der Größe der Session | Adaptertest mit `testing.AllocsPerRun` bei fester Zahl der Treffer: Session mit 10 gegen 1000 Exchanges; die zulässige Abweichung steht als Literal im Test | die ganze Session lesen und verwerfen | andere |
| R12 | (`SPEC-135` Nr. 8) der Dienst ruft `ListByMatchKey` genau einmal, vor der Transaktion und außerhalb, mit der Kennung aus dem Header unverändert und dem Key aus Nr. 7, und `ListBySession` nie | Dienst-Test (heute `TestListBySession`, je Strategie): der Fake zählt beide Operationen und hält die Argumente fest; sein `ListBySession` ist in jedem Test ein Verstoß | zusätzlich `ListBySession`; zweiter Aufruf; Aufruf in `f`; gekürzte Kennung; anderer Key | andere |
| R13 | (Nr. 18, `AGENTS.md` §3.15) die Stelle S3 erhält den `ctx` des Aufrufs | `stellen_test.go` (Tabelle der Stellen) | `context.Background`; der `ctx` von `StartSpan` | andere |
| R14 | (Nr. 18, 8, 3) Fehler an S3: Gegenstand `Session`; `ErrNichtGefunden` ist `ODR-E7003`, jeder andere nach Nr. 15, ein Kontextfehler ohne Code; danach keine Transaktion und kein Blob | `stellen_test.go`, `gegenstand_test.go` und die Fälle S3 in `telemetrie_span_test.go` | Gegenstand `Sonstiges`; der Fehler als leere Liste weitergereicht (`ODR-E5001`) | andere |
| R15 | (Nr. 8) Kandidaten sind allein die gelieferten Exchanges mit byte-gleichem Key, ein leerer passt nie; der Dienst nimmt keine Ordnung der Liste an | Dienst-Test: der Fake liefert absichtlich einen fremden und einen leeren Key mit und die Liste aufsteigend, absteigend und gemischt, je Strategie | Filter im Dienst entfällt; erster oder letzter nach der Stelle in der Liste | andere |
| R16 | (Nr. 9 `strict`) *p* + 1 über der größten Nummer der Session ist `ODR-E5002` | Dienst-Test: *p* + 1 gleich Maximum plus 1 und plus 2, die größte Nummer je Kandidat und fremd | die größte Nummer aus den Kandidaten statt aus dem Port; Grenze um eins verschoben | andere |
| R17 | (Nr. 9) *p* + 1 nicht über der größten Nummer und kein Kandidat ist `ODR-E5007` | Dienst-Test: *p* + 1 gleich dem Maximum und darunter | `ODR-E5002` ohne Kandidat; echt kleiner statt kleiner gleich | andere |
| R18 | (Nr. 9) ist *p* + 1 ein Kandidat, wird er gewählt, ist er nicht replayfähig, ist es `ODR-E5008` | Dienst-Test: replayfähig und nicht, je mit *p* + 1 gleich dem Maximum und darunter | der nächste replayfähige Kandidat statt `ODR-E5008`; `consumed` befragt | andere |
| R19 | (Nr. 9) ohne Kandidaten ist es `ODR-E5001`, in jeder Strategie, vor `ODR-E5002`, `ODR-E5007` und `ODR-E5008` | Dienst-Test: vier Strategien, *p* + 1 unter, gleich und über dem Maximum, dazu die Session ohne Exchanges | `strict` ohne Kandidaten liefert `ODR-E5002` oder `ODR-E5007` | andere |
| R20 | (Nr. 9, 10) `sequential`, `first` und `last` wählen und schreiben den Cursor wie bisher | die bestehenden Tests (`auswahl_test.go`, `szenarien_test.go`, `nebenlaeufig_test.go`, `zustand_test.go`) gegen den Fake der neuen Operation; keine Zelle entfällt | `consumed` nicht befragt; `last` nimmt den ersten | andere |
| R21 | (`SPEC-296` Nr. 8) der eingesetzte Fehler an S3 trifft die neue Operation, und sein Text erreicht keine Senke | `TestWurzelReplayKeinRequestWertInDerSenke`: der Wrapper überschreibt `ListByMatchKey`, sein Zähler ist 1 | der Wrapper überschreibt allein `ListBySession` (Zähler 0, rot) | andere |
| R22 | (`SPEC-135` Nr. 8) der Aufwand je Anfrage wächst nicht mit der Session: keine Abfrage der Operation liest die Session von ihrem Anfang | Adaptertest `TestListByMatchKeyDurchlaeuftNichtDieSession`: in der Datei sind die ersten Blätter der drei B-Bäume von `exchanges` genullt; ein Durchlauf der Session scheitert an jedem, die Operation gelingt | `max(sequence_no + 0)`; `count(*)`; `max` über einen Ausdruck; eine weitere Abfrage über die Session | andere |

**Messpunkte der Tabelle:** 22 Zeilen, *a* = 2 (R07, R08), *b* = 20. R22 kam in der Fix-Runde nach dem Review hinzu; die Marke ist seit dem ersten Code-Commit fest (`SPEC-128` Nr. 5 c), das Gate meldet die gewachsene Tabelle als Hinweis.

**Zum Stand der Umsetzung** (Implementer, 2026-11-27):

- **Haltepunkt zu R07 und R08.** `database/sql` und der Treiber fragen einen offenen Kontext allein über `Done` ab, nie über `Err`. Der Haltepunkt zählt deshalb `Done` und `Err`; sein `Done` liefert `nil`, so dass keine Goroutine entsteht und jeder Aufruf zwischen zwei Schritten der Operation liegt. Die Schleife läuft über alle diese Aufrufe und verlangt, dass beide Stände vorkommen. Kein Fenster zwischen den Abfragen bleibt ohne Haltepunkt; ein Negativ zu R07 und R08 entsteht nicht.
- **R10 und die Ordnung** (`SPEC-105` Nr. 14). Die Abfrage der Treffer trägt kein `ORDER BY` und kein `INDEXED BY`; SQLite wählt `idx_exchange_match`, und `ListByMatchKey` ordnet nach dem Lesen mit `slices.SortFunc`. Mit `ORDER BY` allein wählt SQLite den Unique-Index `(session_id, sequence_no)` und liest jede Zeile der Session, mit `ORDER BY` und `INDEXED BY` sortiert es in einem temporären B-Baum, ohne beschreibbares Temp-Verzeichnis mit `disk I/O error` (Review F-2). Beides ist mit dem Treiber des Moduls (modernc) als rot gesehen.
- **R22 und die beschädigte Datei.** `EXPLAIN QUERY PLAN` nennt für `max(sequence_no)`, `max(sequence_no + 0)` und `count(*)` dieselbe Zeile; den Unterschied hält `TestListByMatchKeyDurchlaeuftNichtDieSession` (erste Blätter genullt). `TestListByMatchKeyBeschaedigteDatei` führt eine genullte Wurzel je Index durch die Operation: `ErrBeschaedigt` aus beiden Abfragen (`SPEC-103` Nr. 6, Review I-4).
- **R12** hält `TestListByMatchKey` (der Test hieß `TestListBySession`); R12 und R15 bis R19 stehen in `auswahl_test.go`.
- **Zusätzliche Fälle über die Tabelle hinaus:** `TestListByMatchKeyAbbruch` (Abbruch an jedem Aufruf des Kontexts: Fehler des Kontexts oder der volle Stand; an den ersten fünf der Fehler) und `TestListByMatchKeyGespeicherterLeererKey` (ein per SQL leer gespeicherter Key).

**Mutationen** (Implementer, 2026-11-27; `tools/harness/mutationslauf.sh` mit `MUTATIONSLAUF_SPEICHER=4g`, 4 CPU, Last des Hosts je Mutation 9 bis 25; R21 von Hand mit `go test` in einer Wegwerfkopie, weil das Werkzeug für `src/bootstrap` den Ausgangsstand als rot meldet). Jede genannte Mutation ist rot gesehen. Nach der Fix-Runde liefen die Mutationen des Adapters (R01 bis R11, R22, beschädigte Datei) neu auf dem geänderten Stand: 41 von 41 rot, Last 7 bis 19; R12 bis R21 stehen auf unverändertem Code.

| Zeile | Mutationen (je rot) | Roter Test |
|---|---|---|
| R01 | ohne Bedingung auf den Key; Ordnung nach der Kennung; ohne Sortierung; absteigend; Key als Anfang; gespeicherter Key als Anfang des Arguments; `COLLATE NOCASE`; `raw_query` nicht gelesen; `key_redacted` nicht gelesen (diese allein vor der Fix-Runde) | `TestRecordingVertrag`, `TestListByMatchKeyOrdnetNachDerNummer` |
| R02 | Maximum der Treffer; Zahl der Treffer; Zahl plus eins | `TestRecordingVertrag` |
| R03 | keine Prüfung der Existenz; Fehler bei Session ohne Exchanges | `TestRecordingVertrag` |
| R04 | leerer Key als fehlender Wert; `ErrNichtGefunden` ohne Treffer; ohne `?2 <> ''` | `TestRecordingVertrag`, `TestListByMatchKeyGespeicherterLeererKey` |
| R05 | Liste ohne Bedingung auf die Session; Maximum ohne Bedingung auf die Session | `TestRecordingVertrag` |
| R06 | Operation ohne `ctx`; Abfrage der größten Nummer ohne `ctx`; Abfrage der Treffer ohne `ctx`; Abbruch als leere Liste; `ErrNichtGefunden` vor dem Kontext | `TestRecordingVertrag`, `TestListByMatchKeyAbbruch` |
| R07 | ohne `BEGIN`; Stand zwischen den Abfragen freigegeben (`ROLLBACK`) | `TestListByMatchKeyGegenAppend` |
| R08 | ohne `BEGIN`; Stand zwischen den Abfragen freigegeben | `TestListByMatchKeyGegenDelete` |
| R09 | ganze Session gelesen und im Adapter gefiltert | `TestListByMatchKeyDekodiertAlleinDieTreffer` |
| R10 | `ORDER BY` zurück; `INDEXED BY` mit `ORDER BY` (temporärer B-Baum); `+match_key`; `+session_id` in der Abfrage des Maximums; Abfrage ohne Bedingung auf den Key | `TestListByMatchKeyNutztDenIndex` |
| R11 | ganze Session gelesen und verworfen | `TestListByMatchKeySpeicher` |
| R12 | zusätzlich `ListBySession`; zweiter Aufruf; Aufruf in `f`; Kennung vorn und hinten gekürzt; anderer Key | `TestListByMatchKey` |
| R13 | `context.Background`; `context.WithoutCancel`; der `ctx` von `StartSpan` | `TestKontextJederStelle` |
| R14 | Gegenstand `Sonstiges`; Fehler als leere Liste; Blob nach dem Fehler; Transaktion nach dem Fehler (diese fangen `TestKontextfehlerJederStelle`, `TestSpanPlatzhalter` und `TestSpanAbbruch`, nicht `TestAbbildungJederStelle`) | `TestAbbildungJederStelle`, `TestKontextfehlerJederStelle` |
| R15 | Filter entfällt; `first`, `last` und `sequential` nach der Stelle in der Liste | `TestKandidatenAlleinDerKey` |
| R16 | Maximum aus den Kandidaten; Grenze um eins nach oben | `TestStrictGroessteNummer` |
| R17 | Grenze um eins nach unten; `ODR-E5002` ohne Kandidat | `TestStrictGroessteNummer` |
| R18 | nächster Kandidat statt `ODR-E5008`; `consumed` befragt | `TestStrictGroessteNummer` |
| R19 | je eine der vier Strategien ohne `ODR-E5001` | `TestOhneKandidatenIstE5001` |
| R20 | `consumed` nicht befragt; `last` nimmt den ersten | `TestAuswahlSequential`, `TestAuswahlFirstLast` |
| R21 | der Wrapper überschreibt die neue Operation nicht (Status 200, Zähler 0) | `TestWurzelReplayKeinRequestWertInDerSenke` |
| R22 | `max(sequence_no + 0)`; `count(*)`; `max(CAST(sequence_no AS INTEGER))`; eine weitere Abfrage `count(*)` über die Session; Treffer mit `+match_key` über die Session | `TestListByMatchKeyDurchlaeuftNichtDieSession` |
| beschädigte Datei | die Abfrage der größten Nummer ohne `klassifiziere`; die Abfrage der Treffer ohne `klassifiziere` | `TestListByMatchKeyBeschaedigteDatei` |

**Raten** (R07 und R08, Messender Implementer, Last des Hosts 16 bis 25 vor und 9 bis 19 nach der Fix-Runde): je Mutation mit Detektor 10 von 10 Prozessen rot, ohne Detektor 20 von 20 (vier Mutationen, in beiden Läufen). Die Fänger sind Wirkungs-Fänger ohne Goroutine; der Detektor trägt nichts bei.

**Akzeptierte Negative**, je mit Grund, Folge und Maß:

- *R15, „ein leerer `MatchKey` passt nie“ im Dienst:* die Prüfung auf den leeren Key neben der Byte-Gleichheit ist ohne Wirkung, weil der Key nach `SPEC-135` Nr. 7 nie leer ist; sie zu entfernen fängt kein Test (1 von 1 Mutation dieser Klasse). Folge: keine im Betrieb.
- *R22, ein Teil der Session ohne ihr erstes Blatt:* eine Abfrage, deren Weg mit der Session wächst und in keinem der drei Bäume das erste Blatt erreicht, bleibt ungefangen, wenn sie keine der zwei geprüften Abfragen ist oder deren Plan gleich bleibt (gemessen an 1 von 1 Mutation: eine weitere Abfrage `count(*) … AND sequence_no > 1000`, grün). Grund: der Fänger zählt nicht die Schritte von SQLite, `database/sql` gibt sie nicht her. Folge: CPU je Anfrage wüchse mit der Session; das zeigt allein der Lasttest.
- *R11, Speicher von SQLite:* `testing.AllocsPerRun` zählt den Go-Heap, nicht den Speicher des Treibers. Folge: ein Wachstum innerhalb von SQLite bei gleichem Plan bleibt ungefangen.
- *R06, Abbruch während des Lesens der Zeilen:* nach der Abfrage der Treffer sagt der Test „Fehler des Kontexts oder voller Stand“ zu, weil `database/sql` die Zeilen in einer eigenen Goroutine schließt. Folge: keine verkürzte Liste, aber ein spät abgebrochener Aufruf kann gelingen.
- *R06, echter asynchroner Abbruch* (Review F-5): selten endet die Operation mit `interrupted (9)` des Treibers statt mit dem Fehler des Kontexts (Messender Reviewer: 2 von 6 400 und 1 von 24 000 Aufrufen); im Dienst wird daraus `ODR-E9001` statt des Fehlers ohne Code (`SPEC-105` Nr. 13, `SPEC-135` Nr. 2). Grund: `ListBySession` verhält sich ebenso (1 von 24 000, Reviewer), es ist keine Verschlechterung, und eine Behebung beträfe jede Abfrage des Adapters; der Haltepunkt von R06 erreicht den Weg nicht, kein Test führt ihn. Folge: ein abgebrochener Request kann als Fehler 500 gezählt werden.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24):

- Repo-Dateien nur mit `Edit` und `Write`; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `--cpus=4`, in der Messung `--cpus=2`. Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Ein Wert mit Zeichen jenseits von ASCII oder mit Backslash-u-Folge entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9). Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/`, `.harness/skills/`, Roadmap, Welle-Pläne und Register schreibt keine Rolle im Zweig (Übergangsregeln Nr. 4).
- Den Aufruf `abnahme` (rund 50 Minuten) fährt allein der Verifier, genau einmal (`SPEC-216` Nr. 1).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-lasttest` liegt in `done/` und `make lasttest` ist auf `main`; die Antworten des Architect aus §3 liegen vor; der Planner hat die Marke gesetzt. **Geprüft beim Start** (Planner, 2026-11-27, Stand `0ee80d43`): `slice-lasttest` liegt in `done/` und ist nach `main` gemergt (`f6ba4ac5`); der Vertrag ist nach `main` gemergt (`6e31d761`); die Antworten stehen in §3. Der Trigger von [welle-betrieb](../welle-betrieb.md) ist eingetreten, eine Ausnahme braucht der Start nicht. Der Slice hat unter den offenen die höchste Priorität (an ihm hängen M5, [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) und die Bestätigung des Zielwerts).

**Größenmarke** (`SPEC-128` Nr. 1, `AGENTS.md` §3.14): 21 Zeilen der Tabelle in §3, *a* = 2, *b* = 19, *f* = 1 (der Fake `exchanges` des Replay-Dienstes bekommt Verhalten; die Stubs der vier anderen Dienste und die Wrapper sind je wenige Zeilen, die der Compiler verlangt, und zählen nicht als Fake), *p* = 100 (Mitte der Erwartung des Architect, 80 bis 120): *S* = 85·2 + 80·19 + 200·1 + 100 = 1990. Keine Schnittfrage (*S* < 2223, Abstand 233). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Das Messprogramm aus §3 liegt in einer Wegwerfkopie und zählt nicht.

```groessenmarke
pfad: src/hexagon/ports/outbound/recording_repository.go
pfad: src/hexagon/ports/outbound/outboundtest/vertrag.go
pfad: src/adapters/driven/sqlite/recording.go
pfad: src/adapters/driven/sqlite/repositories_test.go
pfad: src/adapters/driven/sqlite/export_test.go
pfad: src/adapters/driven/sqlite/list_by_match_key_test.go
pfad: src/hexagon/services/replay/replay.go
pfad: src/hexagon/services/replay/auswahl.go
pfad: src/hexagon/services/replay/auswahl_test.go
pfad: src/hexagon/services/replay/body_test.go
pfad: src/hexagon/services/replay/body_laenge_test.go
pfad: src/hexagon/services/replay/fakes_test.go
pfad: src/hexagon/services/replay/fakes_pruefen_test.go
pfad: src/hexagon/services/replay/gegenstand_test.go
pfad: src/hexagon/services/replay/helfer_test.go
pfad: src/hexagon/services/replay/konfig_test.go
pfad: src/hexagon/services/replay/export_test.go
pfad: src/hexagon/services/replay/nebenlaeufig_test.go
pfad: src/hexagon/services/replay/stellen_test.go
pfad: src/hexagon/services/replay/szenarien_test.go
pfad: src/hexagon/services/replay/telemetrie_dienst_test.go
pfad: src/hexagon/services/replay/telemetrie_span_test.go
pfad: src/hexagon/services/replay/zuordnung_test.go
pfad: src/hexagon/services/replay/zustand_test.go
pfad: src/hexagon/services/record/helfer_test.go
pfad: src/hexagon/services/record/telemetrie_hilfen_test.go
pfad: src/hexagon/services/session/helfer_test.go
pfad: src/hexagon/services/session/beenden_test.go
pfad: src/hexagon/services/inspection/helfer_test.go
pfad: src/hexagon/services/inspection/port_test.go
pfad: src/hexagon/services/baselineimport/helfer_test.go
pfad: src/hexagon/services/baselineimport/fakes_test.go
pfad: src/bootstrap/telemetrie_replay_test.go
produktcode: 100
fakes: 1
basis: 0ee80d43
schaetzung: 1990
marke: 2700
```

**Gesetzt** vom Planner beim Start am 2026-11-27 vor dem ersten Code-Commit: `basis` `0ee80d43`, `schaetzung` 1990, `marke` 2700; mit `make groessenmarke` am Plan vor dem Eintrag nachgerechnet (`Marke steht aus: Schätzung 1990, Marke 2700`).

Der Block nennt die Dateien einzeln (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`). `list_by_match_key_test.go` ist angelegt; die vorgeschlagene `match_key_test.go` entsteht nicht (die Fälle R12 und R15 bis R19 stehen in `auswahl_test.go`) und ist aus dem Block genommen. Der Implementer hat `replay/export_test.go` (die Auswahl bekommt die größte Nummer) und `inspection/port_test.go` (die Zählung der neuen Operation) in den Block aufgenommen.

**Kollisionen** (Planner, 2026-11-27): [slice-bootstrap-m6-transport](../in-progress/slice-bootstrap-m6-transport.md) läuft zugleich; die zwei Blöcke haben keine Datei gemeinsam (dort `src/bootstrap/` mit `konfig_admin.go`, `konfig.go`, `lebenszyklus.go`, `start.go`, `wurzel.go` und sieben Testdateien, hier allein `telemetrie_replay_test.go`). Keiner von beiden ändert `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md` oder `tools/gegenproben/` im Zweig (Übergangsregeln Nr. 7); die Bindung von [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) in `harness/README.md` ändert allein die Closure. Wer als zweiter merged, nimmt zuvor `main` auf und fährt `make gates` neu (Nr. 5). Der Aufruf `abnahme` hält die Sperre rund 50 Minuten und hält in dieser Zeit die Gates der anderen Linie an.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `open` (blockiert): die Suche nach `SPEC-105` Nr. 14 lässt sich ohne Migration oder ohne ein anderes sichtbares Verhalten des Binary nicht liefern (Befund an den Architect); [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) bleibt aktiv. Dass die Messung den Verdacht nicht trägt, ist **kein** Grund der Rückführung: die Suche wird geliefert, der Rest steht unter dem Carveout mit neuem Folge-Slice (Architect, 2026-11-27).
- `in-progress` → `next` (zu groß): der Umfang liegt vor dem Code über der Schnittgrenze von `SPEC-128` Nr. 1.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag.

Die DoD ist vollständig, `make gates` ist grün, der Aufruf `abnahme` liefert für Replay einen gültigen Lauf mit p95, und die Closure-Notiz mit Lerneintrag nennt die Zahlen und den Stand von [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Planner aus dem Vertrag, 2026-11-27):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `RecordingRepository.ListByMatchKey(ctx, session_id, match_key)` | gilt nicht: zwei Zeichenketten ohne Zustand „fehlt“ neben „leer“ | `session_id`: `ErrNichtGefunden`; `match_key`: keine Treffer, kein Fehler (`SPEC-105` Nr. 14; R03, R04) | unbekannte Session `ErrNichtGefunden`; Key ohne Treffer leere Liste; derselbe Key in anderer Session nie (Nr. 14; R03 bis R05) | Fehler des Kontexts ohne Änderung (`SPEC-105` Nr. 13; R06) | gilt nicht für das Ergebnis: die Operation ändert nichts; gleichzeitiges `Append` und `Delete` ganz oder gar nicht (Nr. 14; R07, R08) |
| `Replay`, Stelle S3 und Auswahl | Header der Zuordnung fehlt: `ODR-E5003` vor S3 (`SPEC-135` Nr. 3, 5; unverändert) | ein leerer `MatchKey` eines Exchange passt nie (Nr. 8; R15) | Session unbekannt: `ODR-E7003` vor S3 (Nr. 6 a); zwischen S1 und S3 gelöscht: `ODR-E7003` (Nr. 8; R14); Key ohne Kandidat: `ODR-E5001` (Nr. 9; R19) | Kontextfehler ohne Code an der Stelle (Nr. 18; R13, R14) | je Strategie nach dem Cursor (Nr. 9, 10, 17; R20, unverändert) |

**Zusatztabelle** (`AGENTS.md` §3.16):

| Paar oder Modus | Ergebnis | Fänger |
|---|---|---|
| keine Kandidaten und *p* + 1 über der größten Nummer (`strict`) | `ODR-E5001` (`SPEC-135` Nr. 9) | R19 |
| keine Kandidaten und *p* + 1 nicht über der größten Nummer (`strict`) | `ODR-E5001` | R19 |
| Session gelöscht (S3) und Key ohne Treffer | `ODR-E7003`, die Auswahl läuft nicht (Nr. 3, 8) | R14 |
| Modus `replay` und `hybrid` | derselbe Dienst, dieselbe Auswahl; der Dienst hat keinen Parameter des Modus (Nr. 9) | R12 bis R20; die Tests von `hybrid` bleiben unverändert |
| Allaussage „der Dienst ruft `ListBySession` nie“ (Nr. 8) | über alle Fälle | R12: der Fake meldet jeden Aufruf in jedem Test des Pakets |

- Die Ursache ist nicht die Suche: der Lauf trennt Speicher je Anfrage nicht von der Zahl gleichzeitiger Anfragen, und das Kernel-Protokoll nennt allein den Prozess — **Ausgang:** offen bis zur Closure; Stand: die Messung (§3) weist Speicher und CPU zu mindestens 98 % dem Lesen der Session zu; Adresse: Planner.
- Messung und Vertrag der Behebung hängen aneinander — **Ausgang:** entschieden (Architect, 2026-11-27): der Vertrag der Suche steht unabhängig vom Ergebnis der Messung (`91beaf53`), die Messung läuft vor der Behebung (DoD 1).
- **Frage an den User** (nicht blockierend): soll das Binary die Zahl gleichzeitiger Requests begrenzen (sichtbares Verhalten, `SPEC-135` *Ungeprüft*)? Empfehlung des Architect: nach der Messung entscheiden. Zahlen nach der Behebung (§3): *s* 0,19 bis 0,50 MiB je gleichzeitiger Anfrage (ohne HTTP-Server, ein Treffer je Key), *C* unter 1 ms; unter 64 zugleich freigegebenen Transaktionen enden einzelne Aufrufe mit `ODR-E9001` (`SQLITE_BUSY`) — **Ausgang:** offen; Adresse: User, vorgelegt vom Planner mit den Zahlen aus DoD 1.
- Der Port bekommt eine Operation, und viele bestehende Tests nennen `ListBySession`: ein Umzug kann einen Fänger lockern (`BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger`, `BEO-REPO/ausgabeformat-wechsel-zieht-tests-ausserhalb-des-blocks-mit`). Stand: der Name der Operation ist in den bestehenden Tests ersetzt, die Zählliterale sind um die neue Operation gewachsen (34 auf 35 Operationen, 27 auf 28 verbotene Ports, 140 auf 145 Prüfungen, 30 auf 31 Methoden), keine Zelle ist entfallen; `TestListBySession` ist in `TestListByMatchKey` aufgegangen (12 auf 24 Elemente) — **Ausgang:** offen; Adresse: Reviewer (Diff der bestehenden Tests).
- Auch nach der Behebung liegt p95 über 100 ms, oder das Binary erreicht die Grenze aus einem zweiten Grund — **Ausgang:** offen; Folge: [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) bleibt aktiv mit neuem Folge-Slice.
- Der Aufruf `abnahme` hält `flock /tmp/odata-recorder-heavy.lock` rund 50 Minuten (gemessen: 49 Minuten, Aufbau 275,5 s, 69 MiB); die Wurzel von Docker (`/docker`) war beim letzten Lauf zu 96 % belegt (23 GiB frei), und fremde Container liefen — **Ausgang:** offen; Adresse: Verifier (Bedingungen im Bericht), Planner (Zeitfenster).
- Seit dem letzten Aufruf `abnahme` änderte sich das Binary (Admin-Listener auf `main`) — der erneute Lauf gilt für den neuen Stand und fährt auch `mock`, dessen Ergebnis `bestanden` er nicht voraussetzt — **Ausgang:** offen; Adresse: Verifier.
- Die Setzungen des Lasttests sind ungemessen; ihre Prüfung nach drei gültigen Läufen je Betriebsart (`SPEC-216` Nr. 16 g) hat bisher zwei Läufe `mock` und keinen `replay` — **Ausgang:** offen; Adresse: Architect nach dem dritten gültigen Lauf.
- Nebenbefund der Messung (§3), unabhängig von der Suche und schon am unveränderten Code: geben 64 Transaktionen zugleich frei, wartet der Schreiber über den Busy-Handler von SQLite, und einzelne Replay enden nach rund 5 s mit `ODR-E9001`, Ursache `SQLITE_BUSY`. `SPEC-135` Nr. 17 sagt zwei gleichzeitige `Replay` zu, eine Zahl darüber nicht; Der Slice ändert die Transaktion nicht (§1) — **Ausgang:** entschieden (Architect, 2026-11-27, Review I-1): `SPEC-135` *Ungeprüft* beschreibt das Warten auf den Schreiber samt `ODR-E9001` nach `SQLITE_BUSY_TIMEOUT`; Verhalten unverändert.
- `SPEC-216` Nr. 6 g nannte `make schema-check` als Beleg des Plans der Abfrage; geprüft wird dort eine andere Abfrage (Review F-4) — **Ausgang:** entschieden (Architect, 2026-11-27): der Wortlaut ist berichtigt, den Plan der Abfragen von `ListByMatchKey` prüft der Test des Adapters (R10).
- Die Abfrage der Treffer sortierte in SQLite und brauchte ab einer Größe eine temporäre Datei (Review F-2) — **Ausgang:** entschieden (Architect, `SPEC-105` Nr. 14) und behoben: keine Sortierung in der Datenbank, der Adapter ordnet nach dem Lesen (R01, R10). Der Dienst als Ganzes ohne beschreibbares Temp-Verzeichnis steht in `SPEC-105` *Ungeprüft*.
- Echter asynchroner Abbruch liefert selten den Fehler des Treibers statt des Kontexts (Review F-5, ebenso `ListBySession`) — **Ausgang:** offen; Negativ in §3; Adresse: Architect (ob `SPEC-105` Nr. 13 für alle Abfragen des Adapters durchgesetzt wird).
- CO-001 führt weiter „Ursache: nicht belegt“, §3 belegt sie (Review I-3) — **Ausgang:** offen; Adresse: Planner (Closure).
- `tools/harness/mutationslauf.sh` meldet für `src/bootstrap` den Ausgangsstand als rot und fährt dann keine Mutation der Liste; `go test` über das Paket war in einer Wegwerfkopie grün; die Ursache ist nicht untersucht — **Ausgang:** offen; Adresse: Planner.
- Ein Lauf von `make gates` in der Fix-Runde war in `make test-race` rot: `TestLeserHaeltNieAn` (`tests/lasttest/leser_test.go`, „nach 1s kein Wert“), bei hoher Last des Hosts (Messender Implementer). Der Test liegt außerhalb des Slice (§1), sein Fänger hängt an einer Zeitgrenze — **Ausgang:** offen; Adresse: Planner.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für
Steering-Loop-Regeln.

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`.

**Vorgelagert — offene Beobachtungen sichten:** bei der Anlage (2026-11-26): `BEO-REPO/kleines-profil-gruen-abnahmegroesse-scheitert` (1×; das Profil `rauch` war grün, die Größe `abnahme` beendet das Binary) und `BEO-REPO/fanger-endet-an-der-speichergrenze-des-containers` (1×; ein Fänger zum Speicher deckt die Größen nur bis zur Grenze seines Containers). Beim Start erneut gesichtet (Planner, 2026-11-27, Zähler = Dateien unter `evidence/`): die zwei oben unverändert, dazu offen `BEO-REPO/abnahme-ohne-weg-fuer-lauf-ohne-messwert` (1×; §4 und DoD 3 nennen den Weg ohne Wert), `BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger` (1×) und `BEO-REPO/ausgabeformat-wechsel-zieht-tests-ausserhalb-des-blocks-mit` (4×; der Port zieht Tests von fünf Diensten und dem Root mit, §4). Keiner der offenen Einträge erreicht mit diesem Slice allein neu 3×. Verkörpert und zu beachten: `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (86×, `AGENTS.md` §3.15), `berichtigung-erreicht-nicht-jede-stelle` (53×, §3.9), `randform-erst-im-review-entschieden` (44×, §3.16), `nebenlaeufiger-lauf-nur-im-glueck-gruen` (27×, §3.10; R07, R08), `rate-ohne-detektor-schwankt-mit-der-last` (15×, §3.10; die Messung nennt die Last), `sqlite-transaktionsvertrag-rot-unter-last` (7×) und `betriebsregel-fuer-repo-dateien-verletzt` (23×).

Alle berührten Sub-Areas GF.
