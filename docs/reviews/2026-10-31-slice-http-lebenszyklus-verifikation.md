# Verifikations-Report: slice-http-lebenszyklus — 2026-10-31

**Rolle:** Verifier (Modul 11), einmal, frischer Kontext. Frage: Bauen wir es richtig? — gegen Plan, DoD und den Vertrag (`SPEC-136`, `SPEC-137`), nicht gegen das Review.

**Gegenstand:** Zweig `slice-http-lebenszyklus`, HEAD `22362d5` (Messpunkt nach der Fix-Runde). Commits: `ef5120a`, `047b1d6`, `8ccf424` (Implementer), `37ae9ce` (Review, F-1 bis F-17), `ac444a9`, `22362d5` (Fix-Runde). Basis des Blocks `023617994ec0`.

**Quellen:** Plan `docs/plan/planning/done/slice-http-lebenszyklus.md` (DoD, Tabelle L1 bis L20, Raten, Fenster W1 bis W6, Raster, §6, Marke 2700 fest); `SPEC-136` (Nr. 2, 4c, 4g, 4h, 8, 10 bis 16 mit Zusätzen und *Ungeprüft*), `SPEC-137`, `SPEC-109` Nr. 2, `SPEC-230` Nr. 8; Review `docs/reviews/2026-10-31-slice-http-lebenszyklus-review.md`; `AGENTS.md` §3.6, §3.7, §3.9 bis §3.18.

**Methode.** Den Beleg des Implementers (`make gates`) habe ich selbst gefahren, nicht übernommen. Alle weiteren Läufe: Container `odata-recorder:race` (Go 1.27.1, `--memory=4g --memory-swap=4g --cpus=4 --network none`, ein schwerer Lauf zugleich unter `flock`), Quelle als Kopie im Scratchpad `verify-lebenszyklus/` (Probe-Tests dort, nicht im Zweig), der Zweig-Worktree ist unverändert. „x von 50“ ist `go test -count=50` in **einem** Prozess; „Prozesse“ ist je ein Prozess des einmal mit `-race` gebauten Testbinaries mit `-test.count=1`. Messender: Verifier, Host unter Fremdlast.

---

## Urteil

**Nicht freigegeben, ein Blocker (B-1): falsches Codeverhalten gegen den Wortlaut von `SPEC-136` Nr. 12 (Zusatz), und zwar in der Klasse, die F-1 meinte.** Alles andere sind Reste mit Grund und Folge (Abschnitt *Reste*). Der Rest des Lebenszyklus ist konform: Nr. 13 und 14 halten unter sieben eigenen Sonden (PA bis PE) und vierzehn Mutationen; `make gates` ist grün; Marke, Messung und Zählstellen stimmen.

## Stufen

| Stufe | Ergebnis |
|---|---|
| (a) Tabelle gegen den Vertrag | Zeilen L1 bis L20 gegen `SPEC-136`/`SPEC-137` gelesen, nicht gegen den Plan. Lücken: **B-1** (Nr. 12 Zusatz, Klasse „Fehler des Accept“), **V-2** (Nr. 16 „keine Senke“). Sonst trägt jede Zusage der Nr. 3a, 3b, 4g, 4h, 8, 11, 13a bis 13d, 14, 15 (Pflicht des Roots, kein Fänger hier), 16 eine Zeile. |
| (b) F-1/F-2, Verbindungszustände und Reihenfolgen | Nr. 14 und die Verfolgung der Verbindungen halten (Sonden PC1 bis PE unten). Nr. 12 Zusatz hält **nicht** für jeden Fehler des Accept (**B-1**). |
| (c) Klassen F-3 bis F-10 (§3.17) | F-4 (Formen) und F-1 (Reihenfolgen) gedeckt, mit einer Mutation an Elementen, die weder Review noch Implementer nannten. F-6 (Senken) hat eine Lücke (**V-2**), F-10 (Fristen) eine nicht genannte Prüfgrenze (**V-5**). F-3, F-5, F-7, F-8, F-9 siehe Tabelle. |
| (d) Raten und Negative | L16 Fall 1: 47 von 50 (Plan 48). W1: 0 von 50, Negativ trägt. L18: Plan-Zahlen der ganzen Suite stimmen nicht (**V-4**). W6: 20 von 20. Keine Rate ohne Zahl im Plan. |
| (e) Offene Punkte (a), (b) | Im Wortlaut an den Architect adressiert, aber mit Ausgang „entschieden“ statt „weiter offen“ (**V-3**); die Randform von B-1 steht gar nicht darin. |
| (f) Marke | 1482 von 2700, nachgerechnet (s. u.). |
| (g) `make gates` | grün, Exit 0 aus Datei (s. u.). |
| (h) DoD, `ODR-E5010`, Zählstellen | Haken wahr; 35 Codes; 34 Paare × 2 Bindungen = 68; stimmt (s. u.). |

## Blocker

### B-1 — `Serve` liefert `nil` für einen echten Fehler des Accept, wenn `Anhalten` unmittelbar folgt (falsches Codeverhalten)

`SPEC-136` Nr. 12: „Ein Fehler des Accept vor dem Aufruf von `Anhalten` ist der Fehler und nie `nil`, auch wenn `Anhalten` unmittelbar folgt.“ Der Code entscheidet am Ergebnis von `Close` in `haltAnLocked` (`durchAnhalten = !errors.Is(Close(), net.ErrClosed)`): ein Fehler des Accept gilt als Fehler von `Anhalten`, solange `Close` nicht `net.ErrClosed` liefert. Das deckt den Fehler, den ein äußeres Schließen auslöst (L7, `TestServeFehlerVorAnhalten`: alle vier Reihenfolgen schließen den Listener von außen), nicht **jeden** Fehler des Accept.

- **Sonde PA** (`probe_test.go`, Scratchpad): ein Listener-Fake, dessen `Accept` ohne Schließen einen Fehler („accept kaputt“) liefert und davor `Anhalten` ruft (Haltepunkt im `Accept`, wie die Reihenfolge 2 des Tests): `Serve` liefert **`nil`**; erwartet der Fehler. Gegenprobe PA0 (derselbe Fake ohne `Anhalten`): `Serve` liefert den Fehler. Der Listener ist hier offen, `Close` ist erfolgreich und kann den Fehler nicht von dem unterscheiden, den es selbst auslöste.
- **Klasse (§3.17):** Die Menge ist „jeder Fehler des Accept vor `Anhalten`“, nicht „jeder Fehler durch äußeres Schließen“. Der Fix von F-1 deckt die zweite. Der Plan nennt in §6 (a) und (b) zwei Formen („`Close` nil nach äußerem Schließen“, „`Schliessen` statt `Anhalten`“); diese dritte steht nirgends.
- **Erreichbarkeit im Betrieb:** gering. `net/http` wiederholt temporäre Fehler des Accept, und `net.Listen`-Listener liefern außer nach dem Schließen praktisch keinen dauerhaften Fehler. Das ist aber ein Grund für ein akzeptiertes Negativ mit Grund und Folge, nicht dafür, dass der Wortlaut es nicht verlangt; der Implementer hat es nicht benannt.
- **Auflösung (nicht meine, ich repariere nicht):** entweder der Code deckt die Klasse (der Listener-Wrapper vermerkt einen Fehler des Accept unter der Sperre vor der Rückkehr, `Anhalten` liest ihn), oder der Architect berichtigt Nr. 12 auf „Fehler durch äußeres Schließen“, und §6 führt die Zeile mit Ausgang *offen* und Adresse (`AGENTS.md` §3.13) bis dahin. Beides ist klein.

## Eigene Sonden und Mutationen (Ergebnis)

| Sonde / Mutation | Ergebnis |
|---|---|
| PC1: HTTP/1.1 Pipelining, zwei GET in einem Schreiben, Handler 1 hängt, `Schliessen(30 s)` | Antwort 1 = 200, Request 2 unbeantwortet (Verbindung zu), `Verarbeiten` 1, `Schliessen` = `nil` |
| PC2: zwei GET in einem Schreiben nach `Anhalten` auf Keep-Alive-Verbindung | Antwort 1 = 503 mit `Connection: close`, Request 2 unbeantwortet, `Verarbeiten` unverändert |
| PC3: `Anhalten` während Request 1 hängt, Request 2 pipelined, kein `Schliessen` | Antwort 1 = 200, Antwort 2 = 503 (Nr. 13b: Zulassung maßgeblich), `Verarbeiten` 1 |
| PD: nach `Anhalten` POST mit angekündigtem 1 MiB, 10 gesendet, dann `Schliessen(30 s)` | 503; `Schliessen` = `nil` in unter 5 s; die Verbindung endet danach (EOF) |
| PE: 15 Läufe × 60 Clients mit wechselnden Verbindungszuständen (neu, 1 bis 4 Requests, Leerlauf, Abbruch), `Anhalten` und `Schliessen(2 s)` zu verschiedenen Zeitpunkten, `-race` | in jedem Lauf: `Schliessen` = `nil`, laufende Handler 0, `Verarbeiten` wächst nach der Rückkehr nicht, 0 offene Client-Verbindungen nach 150 ms, kein Datenrennen |
| F-4 über die Klasse (Generator): 10 Methoden × 4 Body-Formen (+ `CONNECT`) = 41 Schlüssel, je „Form nicht abgewiesen“ als eigener Prozess (Auswahl über Umgebungsvariable); darunter `DELETE` mit `Content-Length: 0`, `TRACE` mit 1 KiB, `get` chunked, `HEAD` chunked, `PROPFIND` ohne Body, `OPTIONS` mit Body — der Befund nannte keines | **41 von 41 rot**, ohne Mutation (nicht vorhandener Schlüssel) 4 von 4 grün |
| F-1: `Serve` prüft `angehalten` statt der Entscheidung (`-count=20`) | 20 von 20 rot (`TestServeFehlerVorAnhalten`) |
| F-10: `Schliessen` kehrt ab Frist 1 min beziehungsweise 1 h sofort mit `nil` zurück | rot (`TestSchliessenNull`, Element `MaxInt64`) |
| F-10: sofortige Rückkehr für Fristen in [3 s, 30 s) und (30 s, 24 h) | **grün** (V-5) |
| F-6: `ErrorLog` auf `os.Stdout` | **grün** (V-2); mit einem Test, der auch den stdout des Kindes liest, rot — die Mutation wirkt, der Fänger sieht sie nicht |
| L16 Fall 1, Prüfen vor `IDs`, Zählen danach (`-count=50 -run TestZulassungAtomar/Fall_1`) | 47 von 50 rot |
| W1: Prüfen und Zählen in zwei Abschnitten ohne Aufruf dazwischen (`-count=50`) | 0 von 50 rot (Negativ bestätigt) |
| L18 `angehalten` ohne Sperre gelesen; `laufend++` ohne Sperre; `laufend--` ohne Sperre (`TestZulassungAtomar`, 30 Prozesse `-race`) | 30 von 30; 30 von 30; 30 von 30; ohne Mutation 0 von 30 |
| L18 `laufend++` zwischen Entsperren und Sperren (Zulassung) | `TestZulassungAtomar` 10 von 30 Prozessen; `TestGleichzeitig` des Kerns 3 von 20; **ganze Suite des Pakets: Datenrennen im Prozess 10 von 10**, Stapel jeweils auf `zulassen`/`verlasse` (ohne Wirkung als Test-Aufruf: `-count=50` ohne Detektor 0 von 50 rot, für `laufend++` wie `laufend--`) |
| L18 `geschlossen` nach dem Entsperren gelesen und gesetzt (`TestSchliessenZweimal`, 25 Prozesse) | **2 von 25** (V-4) |

## Reste (Negativ mit Grund und Folge, kein Blocker)

| ID | Kategorie | Befund | Folge |
|---|---|---|---|
| V-2 | MEDIUM | `SPEC-136` Nr. 16 sagt „in keiner Senke“. L19 liest nur `stderr` des Kindprozesses (und `log.Default()` dorthin). `ErrorLog: log.New(os.Stdout, "", 0)` lässt `TestKeinLogDerBibliothek` grün (Mutation wirkt: mit gelesenem stdout rot). Die Menge der Senken (`stderr`, `log.Default()`, `stdout`, Datei, Telemetrie) ist offen; der Fix von F-6 deckt zwei Elemente. | Code ist richtig; eine künftige Verdrahtung auf stdout bliebe unbemerkt. Fix: `aus.String()` in die Prüfung, Prüfgrenze („stderr und stdout des Kindes“) in den Plan. |
| V-3 | LOW | Plan §6 (Fix-Runde): Ausgang „entschieden (Wortlaut Nr. 12, 14), akzeptiertes Negativ“ für (a) und (b), Adresse Architect nur im Konjunktiv („falls der Satz verschärft werden soll“). `AGENTS.md` §3.13: eine zurückgegebene Randform steht mit Ausgang *offen* und der Adresse. Beide Formen weichen vom Wortlaut ab (Nr. 12 „nie `nil`“); der Code entscheidet sie, der Wortlaut tut es nicht. B-1 kommt dazu. | Der Planner setzt bei der Closure: weiter offen, Adresse Architect, für (a), (b), B-1. Kein Code. |
| V-4 | LOW | Plan L18: für `geschlossen` „25 von 25 Prozessen laut Review“; am jetzigen Code (Lesen und Setzen nach dem Entsperren) 2 von 25. Für die ganze Suite nennt der Plan 5 von 10 und 3 von 10; ich messe für `laufend++` zwischen Entsperren und Sperren 10 von 10 Prozessen mit Datenrennen (Review: 10 von 10, bestätigt). Die Unterschiede sind Mutationsformen (ohne Sperre 30 von 30 in `TestZulassungAtomar`; zwischen Entsperren und Sperren 10 von 30), nicht Messfehler; das Negativ im Plan ist damit eher zu pessimistisch (die ganze Suite im Gate fängt es) und für `geschlossen` zu optimistisch. | Zahlen berichtigen (Planner, §3 L18); `geschlossen`: Nr. 14 sagt gleichzeitiges `Schliessen` nicht zu (*Ungeprüft*), Folge also gering. |
| V-5 | INFO | F-10: die Menge „Fristen“ ist unbeschränkt, geprüft sind 1 ns, 200 ms, 300 ms, 1 s, 2 s, 30 s, `MaxInt64`. Eine Rückkehr ohne Warten für Fristen in [3 s, 30 s) und (30 s, 24 h) bleibt grün. Der Plan nennt die Prüfgrenze nicht (§3.17: akzeptiertes Negativ mit Grund und Folge). | Der Plan führt die Fristen als Menge mit Prüfgrenze; die Folge im Betrieb wäre ein Schließen ohne Warten bei üblichen `SHUTDOWN_TIMEOUT`. |
| V-6 | INFO | DoD-Punkt „Review durchgeführt“ steht auf `[ ]`, obwohl der Report vorliegt; die Closure-Punkte stehen offen. | Der Planner hakt bei der Closure. |
| V-7 | INFO | Nicht erreichbar ohne Listener-Fake, der nach `Close` noch annimmt: eine Verbindung, die `Accept` vor dem Schließen annahm und deren Zustandswechsel `New` nach dem Schnappschuss von `schliesseOhneHandler` meldet, bleibt bis `ReadHeaderTimeout` (10 s) offen; kein Handler läuft. | Echte Listener liefern nach `Close` keine Verbindung; kein Fänger nötig. |

## Größenmarke, Messung (f)

`git diff --numstat 023617994ec0..22362d5` über die Pfade des Blocks: Einfügungen 1432 + Löschungen 50 = **1482**; gleich dem Messpunkt im Plan und der Ausgabe des Gates („ok, Messung 1482 von Marke 2700“, `Messpunkt … messung=1482`). Auslastung 0,55. Marke 2700 fest; `ersetzt: 2100` mit Grund im Block. Die Abweichung zu `AGENTS.md` §3.14 (Schätzung × 1,25 = 2500) steht im Plan §6 als weiter offen (Rang 2 vor Rang 8) und ist keine neue.

## `make gates` (g)

Im Worktree unter `flock` gefahren: Exit **0** aus Datei (`gates.exit`); `d-check: 532 Datei(en) geprüft, 0 Befund(e)`; Stufen `pruefung test`, `pruefung race` (Go 1.27.1), `a-check`, `groessenmarke`, `handbuch` und `migration-check` laufen mit; kein `FAIL`, kein `DATA RACE` in der Ausgabe. Der Stand ist HEAD `22362d5`, der Worktree war sauber.

## DoD, `ODR-E5010`, Zählstellen (h)

- DoD-Haken wahr: Lebenszyklus (L3 bis L16, L18, L20), Grenzen (L1, L2, L19), `ODR-E5010` samt Zählstellen, `make gates`, Doku-Update (der Implementer hat `spec/` nicht angefasst, `git diff bce046f..HEAD -- spec` leer). Offen sind Review-Haken, Closure-Notiz, Register, Risiko-Ausgänge, Paarungen (Planner).
- `model/fehler.go`: `CodeE5010` und Zeile `{CodeE5010, "ODR-E5010", []int{503}, ZielService, "Service is shutting down"}`; `AlleCodes()` 35.
- Zählstellen: `telemetry_test.go` und `regeln_test.go` 35/39/38/37; `httpodata/fehler_test.go` `len(codes) != 34` (34 Paare × 2 Bindungen = 68), Kommentar `klassen_test.go` 34; `spec/spezifikation.md` `SPEC-109` Nr. 2 (Zeile `ODR-E5010`), `SPEC-137`, `SPEC-230` Nr. 8 (35), Historie. `grep` nach „34 Codes“ findet nur Historie und Pläne in `done/`/dem eigenen Plan.

## Plan gegen Code (Diff-Befund)

Der Plan bildet den Code ab (§3.9): `Anhalten`/`Schliessen`/`Serve`, Zähler und Kanal `leer` unter einer Sperre, Verfolgung der Verbindungen (`ConnState`, `ConnContext`), `BaseContext`, `EnableFullDuplex` als erste Anweisung, `ErrorLog` verworfen, Grenzen. Abweichung Plan/Vertrag nur in den Resten oben. Jede Zusage der Tabelle hat einen Fänger; die Lücken liegen in Mengen, nicht in Zeilen (B-1, V-2, V-5).

Rohbytes (§3.18): diese Datei auf Escape-Folgen, BOM und NBSP gegengelesen (`LC_ALL=C grep -P`), ohne Fund.
