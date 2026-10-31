# Review-Report: slice-http-lebenszyklus — 2026-10-31

**Review-Art:** Code — geprüft gegen Plan (`slice-http-lebenszyklus` §1, DoD, §3 mit der Tabelle *Zusage · Fänger* L1 bis L20, *Gemessene Raten*, *Sperren und Fenster* W1 bis W5, den Zählstellen von `ODR-E5010`, §4 mit Block, Basis `023617994ec0` und Marke 2700, §6 mit Raster und Zusatztabelle), `SPEC-136` (Nr. 2, 4c, 4g, 4h, 8, 10 bis 16 mit Zusätzen und *Ungeprüft*), `SPEC-137`, `SPEC-109` Nr. 2, `SPEC-230` Nr. 8 und `AGENTS.md` §3.6, §3.7, §3.9 bis §3.18.

**Gegenstand:** `slice-http-lebenszyklus` (Welle `welle-record-replay`), `git diff bce046f..HEAD`; HEAD `8ccf424`, Code-Commits `ef5120a` (Code und Zählstellen), `047b1d6` (Lebenszyklus, Tests L1 bis L20), `8ccf424` (Plan nach den Messungen). Basis des Blocks `023617994ec0`.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `8ccf424` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Methode.** Der Bericht des Implementers (Plan §3, *Gemessene Raten*) wurde gelesen; keine Mutation, keine Rate und keine Einschätzung daraus ist übernommen. Alle Läufe: Container `odata-recorder:race` (Go 1.27.1, `--memory=4g --cpus=4 --network none`, ein Lauf zugleich unter `flock`), Quelle als Kopie im Scratchpad `review-lebenszyklus/`, der Worktree ist unverändert. rund 90 Mutationen an `server.go` und `model/fehler.go` (Mutationen mit Aufbaufehler sind berichtigt und neu gefahren), 6 Sonden als Tests im Scratchpad. „x von 50“ ist `go test -count=50` in **einem** Prozess; „Prozesse“ ist je ein Prozess des einmal mit `-race` gebauten Testbinaries mit `-test.count=1`, Messender der Reviewer.

**Eingangs-Kontext:**

- Slice-Plan `slice-http-lebenszyklus` in `8ccf424`
- `SPEC-136`, `SPEC-137`, `SPEC-109` Nr. 2, `SPEC-230` Nr. 8, `SPEC-121` Nr. 2e, `SPEC-128`
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md)
- `LH-FA-01`, `LH-FA-02`, `LH-QA-05`
- `AGENTS.md` (Hard Rules §3.6, §3.7, §3.9 bis §3.18); Reviewer-Skill
- Nicht doppelt gemeldet (Adresse im Plan §6): `ARC-021` (`.a-check.yml`), Reihenfolge des Herunterfahrens im Composition Root, gleichzeitiges `Schliessen`, Typed-nil.

---

## Antworten auf die Fragen des Auftrags

**(1) Abweichungen vom Plan.** Zähler und Kanal `leer` unter derselben Sperre statt `WaitGroup`: der Plan (§3, Zeile `httpodata/`, W5) ist im selben Commit nachgezogen; der Kanal wird nur unter der Sperre geöffnet (`meldeLeerLocked`), ein verlorenes Aufwecken gibt es nicht (Mutationen `l9_nur_timer` 4 von 4, `l9_halbe_frist` 50 von 50, `l9_kein_warten` rot). `Serve` ohne Start-Prüfung: Nr. 12 Zusatz „Serve nach Anhalten oder Schliessen liefert nil“ stimmt für alle vier Formen (L7 Element 1 bis 4; `Anhalten` schließt den Listener, `Shutdown` liefert `ErrServerClosed`, beides wird `nil`). „Der Fehler des Accept davor ist nie `nil`“ stimmt **nicht**, wenn `Anhalten` unmittelbar nach dem Fehler läuft: Sonde 1 (der Listener ruft nach dem gescheiterten `Accept` und vor der Rückkehr `Anhalten`) liefert `nil` (F-1).

**(2) Akzeptiertes Negativ W1 und L16 Fall 1.** Eigene Messung: Mutation „Prüfen und Zählen in zwei Abschnitten **ohne** Aufruf dazwischen“ 0 von 50; dieselbe mit `runtime.Gosched()` im Fenster 2 von 50; Fall 1 allein fängt die Mutation „Prüfen vor `IDs`, Zählen danach“ in 43 von 50 (Plan: 47 von 50; Fall 0 in 50 von 50). `AGENTS.md` §3.10 verlangt für die Zusage über Nebenläufigkeit „in jedem Lauf rot“ **oder** einen Wirkungs-Fänger **oder** ein akzeptiertes Negativ mit Grund und Folge; die Rate allein genügte nicht, das Negativ trägt aber Grund (kein Argument des Tests im Fenster), Folge im Betrieb und die Rate 0 von 50 und ist zulässig. Die Haltepunkt-Suche ist ehrlich: im Fenster zwischen den zwei Sperren-Abschnitten liegt kein Aufruf eines Fakes. Kein Finding gegen W1; Fall 1 fängt über Fall 0 hinaus nichts (F-11, INFO).

**(3) Fenster W2 bis W5.** W2: L6 fängt die Mutationen „Flag per Vergleich-und-Tausch, `Close` außerhalb“ und „zweiter Aufrufer kehrt per `TryLock` zurück“ je 50 von 50, **weil** der Test den Listener mit einem `Close` von 30 ms umhüllt (die Rate gilt für diese Hülle). W3, W4: benanntes Negativ mit Grund (`SPEC-121` Nr. 2e) und Folge „keine“; die Begründung trägt (nach dem Setzen kann der Zähler nur sinken). W5: kein Fenster, weil `leer` unter der Sperre geöffnet wird; L9 und die Mutationen oben bestätigen das Aufwecken. Zwei Fenster stehen **nicht** in der Tabelle und haben einen Befund: Accept-Fehler bis zum Lesen des Flags (F-1) und der Zustand „Listener schon geschlossen“ in `Anhalten` (F-5).

**(4) `ODR-E5010` und die Zählstellen.** Tabelle (503, `service`, „Service is shutting down“), `fehler_test.go` (Zeile und `fremd` gleich `{200, 418, 507}`), `telemetry_test.go` und `regeln_test.go` (35, 39, 38 und 37, `alleCodes`), `httpodata/fehler_test.go` (34 Paare, die Schleife über `model.AlleCodes()` führt 34 × 2 Bindungen = 68 Läufe) und der Kommentar in `klassen_test.go` stimmen; `grep` nach `34`, `33`, `E5009` findet keine übersehene Stelle (die übrigen Funde zählen anderes). Mutationen: Code fehlt 15 Tests rot; Status 502, Ziel `record`, zweiter Status 500, `{x}` in der message je rot; die message „Service shutting down“ bleibt grün (I-4). Die Pläne in `open/` tragen 35 beziehungsweise 38 Zähler.

**(5) Kommentare nur mit Zusagen, die ein Test prüft (§3.11).** Kein Kommentar im Konjunktiv, kein abwesender Text. Drei Kommentare und Plan-Zeilen sagen mehr zu, als geprüft ist: `Serve` („sonst der Fehler des Accept“, F-1), `Schliessen` („schließt Verbindungen ohne laufenden Request“, F-2) und die Zeile L19 (Plan: „weder in `log.Default()` noch auf stderr“, F-6).

**(6) Auslastung der Marke.** `git diff --numstat` über die Pfade des Blocks: 1161 (1112 Einfügungen und 49 Löschungen), gleich dem Messpunkt; 1161 von 2700 = 0,43; die Schätzung 1990 lag um den Faktor 1,71 über der Messung (Produktcode 172 statt der erwarteten 140). Meldung an den Architect bei der Closure, kein Finding.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Scheitert `Accept` (Listener von außen geschlossen) und folgt `Anhalten` unmittelbar, liefert `Serve` `nil`; `SPEC-136` Nr. 12 sagt: „ein Fehler des Accept vor dem Aufruf von Anhalten ist der Fehler und nie nil, auch wenn Anhalten unmittelbar folgt“. Sonde 1 (ein Listener-Fake ruft `Anhalten` nach dem Fehler des inneren `Accept`) ist am Code rot; der Wrapper liest `schliesst` erst nach der Rückkehr des inneren `Accept`. Das Paar „`Serve` schlägt fehl und `Anhalten` folgt“ nennt als Fänger L7 Element 5, das `Anhalten` nicht ruft; das Fenster steht nicht in W1 bis W5, obwohl der Fake des Listeners ein Haltepunkt ist. Die Mutation „`Serve` prüft `angehalten`“ bleibt grün. | `SPEC-136` Nr. 12; `AGENTS.md` §3.16 (Paar), §3.10 (Fenster) | `src/adapters/driving/httpodata/server.go` · `if err != nil && l.schliesst.Load() {`; `slice-http-lebenszyklus.md` · „`Serve` schlägt fehl und `Anhalten` folgt“ | ja — Sonde 1; Mutation `l7_nil_wenn_angehalten` grün | Paar von Ursachen ohne Fänger |
| F-2 | MEDIUM | `Schliessen` lässt Verbindungen offen, die keinen zugelassenen Request haben und nicht im Leerlauf sind: eine Verbindung ohne ein Byte und eine mit halbem Kopf bleiben nach der Rückkehr von `Schliessen(30 s)` offen (Sonde 5, 6: nach 1 s kein EOF, Zeitgrenze; `net/http` zählt eine neue Verbindung erst nach 5 s als untätig). `SPEC-136` Nr. 14: „Verbindungen ohne laufenden Request schließt es sofort“. L13 führt nur die Leerlauf-Verbindung; das Raster hat keine Zeile *Verbindung · Zustand* (neu, halber Kopf, Leerlauf), und *Ungeprüft* nennt nur Backlog, 503-Verbindung und ungesendeten Body. | `SPEC-136` Nr. 14; `AGENTS.md` §3.16 (Randform ohne Ort), §3.15 | `src/adapters/driving/httpodata/server.go` · `// Mit einem beendeten Kontext schließt Shutdown die Verbindungen ohne laufenden Request`; `lebenszyklus_test.go` · `TestIdleVerbindung` | ja — Sonde 5, Sonde 6 | Randform ohne Ort und Fänger (Verbindungszustand) |
| F-3 | MEDIUM | Die Zusage „höchstens `frist` ab seinem Aufruf“ (Nr. 14; Nr. 15 rechnet `rest` danach) hat keinen Fänger: die Mutation „Frist ab dem Setzen von `angehalten`“ (`rest = frist - Since(seit)`) lässt die ganze Suite grün, weil L12 `Anhalten` und `Schliessen` ohne Pause ruft. Sonde 2 (`Anhalten`, 300 ms warten, `Schliessen(300 ms)` dauert mindestens 300 ms) ist am Code grün und an der Mutation rot. | `SPEC-136` Nr. 14, 15; `AGENTS.md` §3.15 | `src/adapters/driving/httpodata/lebenszyklus_test.go` · `TestSummeDerFristen` | ja — Mutation `l12_frist_ab_anhalten` grün, Sonde 2 | Zusage ohne Fänger |
| F-4 | MEDIUM | Die Allaussage „jede Form des Requests“ (Nr. 13b) hat in L4 eine Liste von 7 Formen, keine Klasse: Mutationen, die genau eine weitere Form nicht abweisen, bleiben grün — chunked (`ContentLength == -1`), `Expect: 100-continue`, `PUT`, `DELETE`, `PROPFIND`, `TRACE`, `PATCH`, `POST` ohne Body, kleingeschriebene Methode `get`, absolute-form (10 von 10 grün); die Elemente der Liste sind rot (GET, POST 1 KiB, POST 1 MiB und 10 Bytes, HEAD, OPTIONS, CONNECT). Chunked und `Expect` hat der Architect in *Annahmen über net/http* gemessen, der Fänger führt sie nicht. | `SPEC-136` Nr. 13b; `AGENTS.md` §3.15, §3.17 | `src/adapters/driving/httpodata/lebenszyklus_test.go` · `func formen() []form {` | ja — Mutationen `l4x_*` grün | Fänger führt nur eine Form der Eingabe |
| F-5 | MEDIUM | „Ein weiteres und ein gleichzeitiges `Anhalten` sind wirkungslos“ (Nr. 13d) hat für den Listener keinen Fänger: ruft der zweite Aufruf `Close` des Listeners nochmals (Wache `if s.angehalten` entfernt, oder `Close` im Zweig), bleibt die Suite grün (`l6_ohne_wache`, `l6b_close_zweimal_listener`, je ein Lauf). Der Listener-Fake `langsamListener` zählt `Close` nicht; die erste Mutation von L6 („doppelt, Panik“) trifft den Kanal `leer`, nicht den Listener. | `SPEC-136` Nr. 13d; `AGENTS.md` §3.15 (Fake zählt jede Operation) | `src/adapters/driving/httpodata/lebenszyklus_fake_test.go` · `func (l langsamListener) Close() error {` | ja — Mutationen `l6_ohne_wache`, `l6b_close_zweimal_listener` grün | Zusage ohne Fänger (Fake zählt nicht) |
| F-6 | MEDIUM | Die Zusage „weder in `log.Default()` noch auf stderr“ (Nr. 16, Plan-Zeile L19) fängt nur der Teil `log.Default()`: mit `ErrorLog: log.New(os.Stderr, "", 0)` bleibt die Suite grün, die Panik mit dem Wert steht dann auf stderr. Der Test setzt `log.SetOutput` und liest keinen Strom von stderr. | `SPEC-136` Nr. 16; `AGENTS.md` §3.11, §3.15 | `src/adapters/driving/httpodata/lebenszyklus_test.go` · `// L19: der Adapter schreibt kein Log von net/http` | ja — Mutation `l19b_errorlog_stderr` grün | Zusage ohne Fänger (stderr) |
| F-7 | MEDIUM | Der Fänger für die Prüfung auf `leer` vor dem Warten ist nur mit Wahrscheinlichkeit rot: wird die Vorprüfung entfernt, wählt `select` bei bereitem `leer` und abgelaufenem Timer zufällig; `TestSchliessenNull` ist 47 von 50 rot (ein Einzellauf blieb grün). Der Plan nennt die Zeile deterministisch rot und die Mutation nur zusammen mit dem Überlauf („50 von 50“); eine Rate für die Mutation allein fehlt. | `AGENTS.md` §3.10 (Rate nennen) | `src/adapters/driving/httpodata/lebenszyklus_test.go` · `fristen := []time.Duration{0, 1, 30 * time.Second, math.MaxInt64}`; `slice-http-lebenszyklus.md` · „Überlauf (Zeitpunkt addiert): 50 von 50“ | ja — Mutation `l11_ohne_leer_vorpruefung`, `-count=50` | Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt |
| F-8 | MEDIUM | `Schliessen` schließt den Listener nur dann, wenn `Serve` schon läuft: lässt man `Listener.Close` in `Schliessen` weg (das Setzen des Flags bleibt), schließt allein `Shutdown` den Listener, den `Serve` verfolgt. Ist `Serve` noch nicht gestartet, bleibt der Listener offen (Sonde 3 rot an der Mutation). L14 prüft `Dial` nach `Schliessen` direkt nach dem Start: rot nur, wenn die Goroutine von `Serve` noch nicht verfolgt (15 von 50); L7 Element 4 (`Schliessen` vor `Serve`) prüft den Listener nicht. Der Plan nennt keine Rate für L14. | `SPEC-136` Nr. 13a, 14; `AGENTS.md` §3.10, §3.15 | `src/adapters/driving/httpodata/lebenszyklus_test.go` · `func TestSchliessenOhneAnhalten`; `server.go` · `s.haltAnLocked()` in `Schliessen` | ja — Mutation `l14_ohne_listener_close`, `-count=50`; Sonde 3 | Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt |
| F-9 | LOW | Die Zeile L18 nennt `TestZulassungAtomar` als Fänger für „Zustand (`angehalten`, Zähler der Handler)“. Für den Zähler ist der Detektor dort 0 von 25 (`laufend++` nach dem Entsperren) und 0 von 50 (`laufend++` und `laufend--` zwischen Entsperren und Sperren, `laufend--` ohne Sperre), weil Fall 1 fast alle Requests nach `Anhalten` abweist. Rot ist es über `TestGleichzeitig` des Kerns (10 von 10 Prozessen der ganzen Suite). `geschlossen` ohne Sperre fängt `TestSchliessenZweimal` (25 von 25), nicht der genannte Test. | `AGENTS.md` §3.15 (Tabelle nennt den Fänger) | `slice-http-lebenszyklus.md` · `| L18 | Der Zustand (`angehalten`, Zähler der Handler)` | ja — Mutationen `l18b_laufend_ohne_sperre`, `l18_verlasse_ungeschuetzt`, `-race` | Fänger der Tabelle fängt nicht, was die Zeile nennt |
| F-10 | LOW | L9 prüft die Rückkehr unter 5 s und die Antwort des Clients, nicht, dass `Schliessen` auf den Handler gewartet hat: die Mutation „`Schliessen(frist >= 10 s)` kehrt sofort mit `nil` zurück“ lässt `TestSchliessenFrueh` grün (Sonde 4: Dauer 70 µs, `Close` 0 bei der Rückkehr); rot wird sie nur zufällig in `TestSchliessenNull` (Liste der `ctx`-Fehler). | `AGENTS.md` §3.15 (beide Seiten einer Grenze) | `src/adapters/driving/httpodata/lebenszyklus_test.go` · `func TestSchliessenFrueh` | ja — Mutation `l9c_sofort_ab_10s`, Sonde 4 | Fänger prüft eine Seite der Grenze |
| F-11 | INFO | L16 Fall 1 fängt die Mutation, die Fall 0 schon fängt (43 von 50, Plan 47 von 50), und die Mutation ohne Aufruf im Fenster nicht (0 von 50; mit `Gosched` im Fenster 2 von 50). Die Rate des Plans weicht unter Last von der des Reviewers ab. Zuständig: Planner (Zahl in §3). | `AGENTS.md` §3.10 | `slice-http-lebenszyklus.md` · `| L16 Fall 1 | dieselbe Mutation (ein Aufruf im Fenster) | 47 von 50` | ja — `-count=50 -run TestZulassungAtomar/Fall_1` | Rate unter Last |
| F-12 | INFO | L5 („läuft zu Ende“) sieht keinen Abbruch des `ctx` durch `Anhalten`: die Mutation `s.abbruch()` in `Anhalten` ist in `TestLaufenderRequestEndet` 0 von 50 rot (der Fake `wartetFrei` ignoriert `ctx`); rot ist sie in L10, L11, L12 und L20. Zuständig: Implementer, falls die Zeile den `ctx` zusagen soll. | `SPEC-136` Nr. 4g, 13c | `src/adapters/driving/httpodata/lebenszyklus_test.go` · `func TestLaufenderRequestEndet` | ja — Mutation `l5_abbruch_ctx_bei_anhalten` | — |
| F-13 | INFO | `TestServeEndetMitNil` ruft `Anhalten` und `Schliessen(0)` ohne Zeitgrenze; hängt `Schliessen` (Mutation ohne Anhalten), endet der Lauf nach 150 s mit `panic: test timed out` und nennt nur diesen Test, die Ergebnisse der übrigen gehen verloren. | `AGENTS.md` §3.10 | `src/adapters/driving/httpodata/lebenszyklus_test.go` · `_ = srv.Schliessen(0)` | ja — Mutation `l14_ohne_halt` | Hänger bis zur Zeitgrenze von go test |
| F-14 | INFO | Der Zweig `if err == nil { return errors.New("httpodata: Serve endete ohne Fehler") }` in `Serve` ist unerreichbar (`http.Server.Serve` liefert immer einen Fehler) und ohne Test. | Maintainability | `src/adapters/driving/httpodata/server.go` · `Serve endete ohne Fehler` | nein | Toter Zweig |
| F-15 | INFO | Die Zeile L17 nennt die message „Service is shutting down“; kein Test prüft den Wortlaut (`SPEC-109` Nr. 3: Wortlaut kein Vertrag; `TestMessageFest` prüft die Form). Die Mutation „Service shutting down“ bleibt grün. | `SPEC-109` Nr. 3 | `src/hexagon/model/fehler.go` · `"Service is shutting down"` | ja — Mutation `l17_message` | — |
| F-16 | INFO | `EnableFullDuplex` auf dem 503-Pfad hat keine beobachtbare Wirkung: nach dem Verschieben hinter die Zulassung bleibt die Suite grün, weil `Connection: close` das Verwerfen des Bodys schon verhindert. Nr. 4h („auch dem der Nr. 3b“) hat dort keinen Fänger und braucht keinen; auf dem Normalpfad ist er rot (`TestLeihe`, `TestLeiheJederPfad`). Zuständig: Architect, falls der Satz bleiben soll. | `SPEC-136` Nr. 4h | `src/adapters/driving/httpodata/server.go` · `_ = http.NewResponseController(w).EnableFullDuplex()` | ja — Mutation `l4_fullduplex_erst_nach_zulassen` | — |
| F-17 | INFO | `SPEC-109` Nr. 2 führt die Zeile `ODR-E5010` hinter der letzten Zeile der Zwei-Spalten-Tabelle mit leeren Zellen für die rechte Hälfte; die Tabelle ist nicht mehr nach Code geordnet. Vom Planner geschrieben. Zuständig: Planner. | Maintainability | `spec/spezifikation.md` · Zeile mit `ODR-E5010` unter der Tabelle der Codes in `SPEC-109` Nr. 2 | nein | Tabelle nicht geordnet |

## Raten der nebenläufigen Zeilen (`AGENTS.md` §3.10)

Messender: Reviewer; „x von 50“ ist ein Prozess mit `-count=50`, Prozesse mit `-race` und `-test.count=1`. Host unter Last durch fremde Läufe. Spalte *Plan* ist die Zahl des Implementers.

| Zeile | Mutation | rot | Plan |
|---|---|---|---|
| L3 | `Anhalten` ruft `Shutdown`; `Listener.Close` entfällt | 50 von 50; 50 von 50 | 50 von 50 |
| L5 | Abbruch nach dem Dienst bei angehaltenem Server | 50 von 50 | 50 von 50 |
| L6 | `Close` außerhalb der Sperre; `TryLock` mit Rückkehr | 50 von 50; 50 von 50 (mit der 30-ms-Hülle) | 50 von 50 |
| L9 | nur Timer (volle Frist); halbe Frist | 4 von 4 (je 30 s); 50 von 50 (`TestSummeDerFristen`) | 3 von 3 |
| L10 | `nil` statt `ErrFristAbgelaufen`; Rückkehr vor dem Handler | 50 von 50; 50 von 50 | 50 von 50 |
| L11 | Frist 0 als Ende; Frist über 24 h als 0; `int32`; **Vorprüfung entfernt** | rot im Einzellauf; 50 von 50; 50 von 50; **47 von 50** | 50 von 50 (nur mit Überlauf) |
| L12 | `Anhalten` wartet 300 ms; `Schliessen` wartet zweimal; **Frist ab `Anhalten`** | 50 von 50; 50 von 50; **grün** | 50 von 50 |
| L16 | Fall 0: Prüfen vor `IDs`; Fall 1 allein; ohne Aufruf; mit `Gosched` | 50 von 50; 43 von 50; 0 von 50; 2 von 50 | 50, 47, 0 |
| L18 | `angehalten` gelesen ohne Sperre; `angehalten` geschrieben ohne Sperre; Basis ohne Mutation | 25 von 25 Prozessen; 25 von 25; 0 von 25 | 30 von 30; 0 von 30 |
| L18 (Zähler) | `laufend++` ohne Sperre; `laufend++` zwischen Entsperren und Sperren; `laufend--` ohne Sperre | `TestZulassungAtomar` 0 von 25, 0 von 50, 0 von 50; ganze Suite 10 von 10 (`TestGleichzeitig`) | nicht genannt |
| L20 | `BaseContext` fehlt; nur `http.Server.Close` | 50 von 50; 50 von 50 | 50 von 50 |

## Auflösung der Zeilen (eigene Mutationen, rote Tests)

| Zeile | Mutationen | rot |
|---|---|---|
| L1 | `MaxHeaderBytes` 65537; `ReadHeaderTimeout` entfällt | `TestServerVoreinstellungen` |
| L2 | Grenze 128 KiB; Grenze 32 KiB | `TestKopfGrenze` (und L1) |
| L3 | Shutdown; kein `Close`; Warten auf `leer` | `TestAnhaltenSchliesstListener` |
| L4 | GET, POST 1 KiB, POST 1 MiB, HEAD, OPTIONS, CONNECT je nicht abgewiesen; `Connection: close` nicht bei HEAD, nicht bei OPTIONS; Status 500 bei CONNECT; `DisableGeneralOptionsHandler` entfällt; Zähler vor der Prüfung; **10 weitere Formen grün (F-4)** | `TestAngehaltenDreiNullDrei`, `TestIdleVerbindung` |
| L5 | Abbruch nach dem Dienst; `ctx` beendet (nur L10 bis L12, L20) | `TestLaufenderRequestEndet` und weitere |
| L6 | `Close` außerhalb der Sperre; `TryLock`; **zweites `Close` des Listeners grün (F-5)** | `TestAnhaltenWiederholt` |
| L7 | `ErrServerClosed` zurück; Flag immer gesetzt; `net.ErrClosed` als Ende; **`angehalten` statt Flag grün (F-1)** | `TestServeEndetMitNil` |
| L8 | Prüfung nach dem Anhalten; nur −1; `MinInt64` ausgenommen | `TestSchliessenNegativ` |
| L9 | nur Timer; halbe Frist; kein Warten; **sofort ab 10 s grün (F-10)** | `TestSchliessenFrueh`, `TestSummeDerFristen` |
| L10 | `nil`; Join mit `ErrUngueltig`; keine Rückkehr auf den Handler; kein `abbruch` | `TestSchliessenFristAbgelaufen`, `TestSchliessenKontextUngelesenerBody` |
| L11 | Frist 0 ohne Ende; Frist 0 als `nil`; über 24 h als 0; `int32`; **Vorprüfung entfernt (F-7)**; unter 1 ms als 0 grün | `TestSchliessenNull` |
| L12 | wartet 300 ms; wartet zweimal; **Frist ab `Anhalten` grün (F-3)** | `TestSummeDerFristen` |
| L13 | `Schliessen` schließt Idle-Verbindungen nicht; Zähler vor der Prüfung | `TestIdleVerbindung` |
| L14 | `Schliessen` hält nicht an; ohne `Listener.Close` (15 von 50, F-8) | `TestSchliessenOhneAnhalten` |
| L15 | zweites wartet; zweites liefert `ErrFristAbgelaufen` | `TestSchliessenZweimal` |
| L16 | Fall 0: Prüfen vor `IDs`; Fall 1: zwei Abschnitte (grün, Negativ); `Gosched`; kein Warten | `TestZulassungAtomar` |
| L17 | Code fehlt; Status 502; Ziel `record`; zweiter Status; `{x}` in message; message umformuliert (grün) | `TestCodeTabelle`, `TestMessageFest`, `TestFehlerobjektJeCode` und weitere |
| L18 | siehe Raten | `TestZulassungAtomar`, `TestGleichzeitig`, `TestSchliessenZweimal` |
| L19 | `ErrorLog` entfällt; **`ErrorLog` auf stderr grün (F-6)** | `TestKeinLogDerBibliothek` |
| L20 | `BaseContext` fehlt; unabhängiger Kontext; nur `Close` | `TestSchliessenKontextUngelesenerBody` |

Sonden (Tests im Scratchpad, am Code): Sonde 1 (Accept-Fehler, dann `Anhalten`) rot, 2 (Frist ab Aufruf) grün, 3 (Listener nach `Schliessen` vor `Serve`) grün, 4 (Warten auf den Handler) grün, 5 (neue Verbindung nach `Schliessen`) rot, 6 (halber Kopf) rot.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driving/httpodata/server.go` | gegen `SPEC-136` Nr. 2, 3, 4c, 4g, 4h, 8, 10 bis 16 gelesen und mutiert; Befunde F-1, F-2, F-5, F-8, F-14, F-16; Sperre, Zähler und `leer` sonst ohne Befund |
| `src/adapters/driving/httpodata/*_test.go` (Lebenszyklus, Fakes, Zählstellen) | Tabelle L1 bis L20 aufgelöst; Befunde F-3, F-4, F-6, F-7, F-9, F-10, F-12, F-13; Rohbytes (keine Escape-Folge, BOM, NBSP) ohne Befund |
| `src/hexagon/model/` (`fehler.go`, `fehler_test.go`) | geprüft, ohne Befund (Tabelle, `fremd`, Zählung); F-15 |
| `src/adapters/driven/telemetry/` (`telemetry_test.go`, `regeln_test.go`) | geprüft, ohne Befund (35, 39, 38 und 37, Kommentare nachgezogen) |
| `spec/spezifikation.md` (`SPEC-136` Zusätze, `SPEC-137`, `SPEC-109` Nr. 2, `SPEC-230` Nr. 8, Historie) | Code und Wortlaut gelesen; Abweichungen F-1, F-2; F-17 |
| `docs/plan/planning/in-progress/slice-http-lebenszyklus.md` | §3.9: Code, Tabelle, Raten, Negativ W1, Messpunkt im Commit `8ccf424`; Zahlen der Raten bestätigt bis auf F-7, F-11; Größenmarke 1161 von 2700 bestätigt |
| `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/` | in den Commits des Implementers unverändert; keine `//nolint`-Direktive, keine Lockerung (`AGENTS.md` §3.2, §3.6) |

## Summary

0 HIGH · 8 MEDIUM · 2 LOW · 7 INFO. Die Zusagen des Lebenszyklus sind im Verhalten bis auf zwei Randformen eingehalten: ein Accept-Fehler mit unmittelbar folgendem `Anhalten` ergibt `nil` (F-1), und `Schliessen` lässt neue und halbfertige Verbindungen offen (F-2). Sechs MEDIUM sind Zusagen, Mengen oder Raten ohne tragenden Fänger (F-3 bis F-8). Die Rate-Angabe zu W1 genügt `AGENTS.md` §3.10 als akzeptiertes Negativ mit Grund und Folge. Nach der Behebung von F-1 bis F-8 läuft eine zweite Runde nur auf diesen Zeilen.
