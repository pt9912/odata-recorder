# Verifikation: slice-admin-lebenszyklus — 2026-11-08

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-238` Nr. 1 bis 15, `SPEC-136` Nr. 13 c und 14, die Tabelle *Zusage · Fänger* (19 Zeilen und die Aufträge der Fix-Runde) und die Hard Rules; nicht Diff-Review (siehe `2026-11-08-slice-admin-lebenszyklus-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-admin-lebenszyklus`, HEAD `8af67837` (Fix-Runde nach dem Review; Architect `319aa096`), Worktree `/Development/odata-recorder-wt-admin-lebenszyklus`. Plan: `docs/plan/planning/done/slice-admin-lebenszyklus.md`. Bezug: [LH-FA-13](../../spec/lastenheft.md), [LH-QA-04](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden mit Auflagen: zwei MEDIUM, drei LOW, keine Fix-Runde am Produktcode.** Der Produktcode (`lebenszyklus.go`, `admin.go`) erfüllt `SPEC-238` Nr. 3 b, 9, 10, 11 und `SPEC-136` Nr. 12 bis 14 in allen Formen, die ich gegen ihn gefahren habe. Alle Raten des Plans für S3, S6, S13 und F-3 stimmen mit meiner Messung überein oder sind schärfer. Zwei Fixes der Fix-Runde decken ihre Klasse nicht: F-6 (negative Frist „ändert nichts“: zwei von sechs Zustandsteilen sind ungefangen) und F-1 (eine Abweichung vom Vertrag steht als *entfallen* im Plan, ohne Entscheidung des Architect). Beide sind Test- und Planlücken; der Code selbst ist für die benannten Formen richtig. `make gates` ist im eigenen Lauf grün (§1).

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| `make gates` | einmal, unter `flock /tmp/odata-recorder-heavy.lock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/vadl-gates.out`: `EXIT 0`; `d-check: 924 Datei(en) geprüft, 0 Befund(e)`; `migration-check: ok`; Stufen `test`, `build`, `race`, `lint` ohne Befund (die Schichten von `test` und `race` kamen aus dem Cache des Implementers: gleiche Quelle, gleicher Inhalt); `groessenmarke`: `slice-admin-lebenszyklus.md` zeilen=19 produktcode=190 schaetzung=1955 marke=2700 messung=2330; `zeichen: ok, 1849 Datei(en)`. Der Lauf beschrieb die geteilten Tags `odata-recorder:*` (unter dem Lock). |
| Race-Lauf auf HEAD | eigenes Image `verif-adl-8af:race` (Stufe `race`, `--target race`, Quellen des Zweigs; Schicht war im Cache des Implementers, daher fuhr ich das Testbinary selbst); Baseline 3 Prozesse `-race -test.count=1` über das ganze Paket `admin`: 0 rot |
| Plan-vs-Code-Diff | `git diff --stat fb575e00..HEAD` über `src/hexagon`, `src/bootstrap`, `harness`, `Makefile`, `*.mk`, `.golangci.yml`, `.a-check.yml`: leer. Produktcode allein `src/adapters/driving/admin/lebenszyklus.go` (neu, 233 Zeilen) und `admin.go` (+51/-). Der Kommentar von `ServeHTTP` trägt den Wortlaut des Auftrags F-7 („`EnableFullDuplex` erlaubt dem Handler, Lesen und Schreiben zu mischen; kein Test fängt das Entfernen (N-13).“). `spec/spezifikation.md`: die Zeile zu Nr. 11 (Randformen a, b) und die Historie; keine weitere Stelle. |
| Fehlercodes | 40 Codes in `src/hexagon/model/fehler.go` (`grep -c '"ODR-[A-Z][0-9]{4}"'`), unverändert; `ODR-E5010` ist bekannt, kein Katalog-Eintrag nötig. |
| Größe | `git diff --numstat fb575e00..HEAD -- src/adapters/driving/admin`: 2299 + 31 = **2330** von Marke 2700 (Messung des Plans bestätigt). |
| Merge-Nähe | `git merge-tree --write-tree HEAD <Zweig>` gegen `main` (`daefe91`), `slice-etag-versionen`, `slice-fallback-policy-entscheidung`: je genau ein Textkonflikt, `spec/spezifikation.md`, und zwar die Änderungstabelle (Historie, Zeilen ab 4018/4068/4020); kein Konflikt im Code. Dieser Slice berührt `wurzel.go`, `konfig.go` und `konfig_werte.go` nicht (sie ändert allein `slice-etag-versionen`); beide Zweige tragen den Admin-Adapter nicht. |
| N-15 | `src/adapters/driving/httpodata/server.go` Zeile 162 (`l.s.acceptFehler = nil` durch `_ = err` ersetzt): `go test ./src/adapters/driving/httpodata` bleibt grün; das Negativ stimmt, siehe V5. |
| Mutations-Container | alle Mutationsläufe in einem Container je Lauf aus dem eigenen Image `verif-adl-8af:race` (`--network none --memory 4g --cpus 4`, Wegwerfkopie in `/w`, Original in `/pristine`), unter `flock`; nie `docker kill` nach Image; ein Prozess ist ein Aufruf des Testbinary mit `-race -test.count=1` (Ausnahme N-11: ohne Detektor), je Prozess `timeout 120`. Loadavg (1 Minute) 3 bis 17 während der Läufe (drei Slices parallel), die Raten sind Größenordnungen. |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Serve, Anhalten, Schliessen (A2, K5, S1 bis S13) | **erfüllt, Haken fehlt; Auflagen V1, V2** | §3 und §4; alle 19 Zeilen haben einen Fänger, der aus dem benannten Grund rot wird. |
| Grenzen und `ctx` (G1 bis G3, X2) | **erfüllt, Haken fehlt** | G1 bis G3 und X2 in den Läufen grün; N-12 bestätigt (§5). |
| `make gates` grün | **erfüllt** | eigener Lauf, §1 |
| Review durchgeführt, Report liegt vor | erfüllt (Haken fehlt) | `docs/reviews/2026-11-08-slice-admin-lebenszyklus-review.md` (0 HIGH, 6 MEDIUM, 3 LOW, 3 INFO); F-1 bis F-8 in `8af67837` umgesetzt, Review der Fixes ist dieser Bericht. |
| Doku-Update | erfüllt | `SPEC-238` Nr. 11 um Randform a und b und Historie (Architect); kein Handbuch-Kapitel nötig. |
| Katalog / Binary | erfüllt | kein neuer Code, `src/bootstrap` unverändert (Diff leer). |
| Closure-Notiz, Register, Paarungen | **offen, Sache des Planners** | §7 ist `Wird bei der Closure gefüllt`; kein DoD-Haken gesetzt. |
| Ausgang je Risiko (§6) | **offen** | Zeilen mit *offen* oder *weiter offen*: G1 `ErrorLog` (Adresse Planner von `slice-admin-fehlerabbildung`), die vier Register-Zeilen (`BEO-REPO/…`) und die Zeile „Aufträge der Fix-Runde … Ausgang: offen; Adresse: Implementer“, die nach dieser Runde auf *entschieden* gehört; dazu V2. |

Liefer-Punkte: zwei, unter dem Limit von drei.

## 3. Tabelle *Zusage · Fänger* gegen den Vertrag

Geprüft Zeile für Zeile gegen `SPEC-238` Nr. 3 b, 9, 10, 11, 12 und `SPEC-136` Nr. 12 bis 14. Die 19 Zeilen (A2, K5, S1 bis S13, X2, G1 bis G3) decken: Nr. 3 b (A2, S4), Nr. 9 (X2, S9), Nr. 10 (G1 bis G3), Nr. 11 samt Randform a (S2) und b (S7), Nr. 12 (S13), `SPEC-136` Nr. 12 (S1, S2), 13 a bis d (S3 bis S6), 14 (S7 bis S12). Nr. 13 bis 15 von `SPEC-238` haben in diesem Slice keinen Code (Geheimnisse: die Fehler von `Serve` und `Schliessen` tragen weder Wert noch Token, `ungueltig("frist ist negativ")` nennt die Frist nicht). Jede Zeile nennt einen Fänger, dessen Symbol im Quelltext steht (`grep -n "^func Test"` über `lebenszyklus_test.go`, `lebenszyklus_accept_test.go`, `grenzen_test.go`); die vier Fänger der Mutationen (S3 bis S6, S13) sind unten mit Rate belegt.

**Eine Zusage hat keinen Fänger für ihre ganze Klasse:** „eine negative Frist ändert nichts“ (Nr. 11 b), siehe V1. **Eine Abweichung steht ohne Entscheidung:** das Schließen einer Verbindung mit halbem Body nach der Antwort, siehe V2.

### V1 · MEDIUM — F-6 deckt die Klasse „ändert nichts“ nicht

- `quelle`: `AGENTS.md` §3.17 (der Fix deckt die Klasse der Zusage), §3.15; `SPEC-238` Nr. 11 (b): „eine negative `frist` ist `model.ErrUngueltig` und ändert nichts, auch auf einem angehaltenen und auf einem schon geschlossenen Server“.
- `menge`: die Teile des Zustands, die ein `Schliessen` ändern kann: `angehalten`, `geschlossen`, Listener, Kontext des Servers (`abbruch`), Verbindungen, Zähler; 6 Elemente. F-6 (3 Zustände × 2 Fristen) prüft die Wirkung auf `angehalten` (503), Listener (`Dial`), Verbindungen (Leerlauf bleibt bedient) und das Ausbleiben von `Serve`-Ende.
- `mutation, die der Befund nicht nannte` (alle vor die Rückgabe des Fehlers gesetzt, `lebenszyklus.go`, Prozesse mit `-race`, `TestSchliessenNegativ`):

| Mutation | Ergebnis |
|---|---|
| `s.geschlossen = true` vor dem Fehler | **0 von 5 rot** |
| `s.abbruch()` vor dem Fehler (beendet den `ctx` der laufenden Requests) | **0 von 5 rot** |
| `_ = s.listener.Close()` vor dem Fehler | 5 von 5 rot |
| `s.schliesseOhneHandler()` vor dem Fehler | 5 von 5 rot |
| Grenze `frist < -1` statt `frist < 0` | 5 von 5 rot |

- `folge`: ein Aufruf mit negativer Frist (Programmfehler des Root) markierte den Server als geschlossen; das folgende `Schliessen(30 s)` kehrte mit `nil` zurück, ohne zu warten und ohne den Listener zu schließen. Der Test sieht das nicht, weil er nach dem Fehler nur `Schliessen(0)` mit `nil` erwartet und im Zustand *frisch* kein `Dial` und keinen hängenden Handler prüft.
- `Auflage an den Implementer` (eine kleine Fix-Runde, nur Tests): in `TestSchliessenNegativ` für *frisch* nach der negativen Frist (a) `Dial` gelingt, (b) ein Request mit hängendem Handler läuft weiter und sein `ctx` endet nicht, (c) ein folgendes `Schliessen(0)` ergibt `ErrFristAbgelaufen`, nicht `nil`; beide Mutationen aus der Tabelle müssen rot werden (Bericht mit Prozessen und Messendem), die Zeile S7 in §3 des Plans und das Raster nachziehen.

### V2 · MEDIUM — F-1: Abweichung vom Vertrag mit Ausgang *entfallen* statt Entscheidung

- `quelle`: `SPEC-136` Nr. 14 (über `SPEC-238` Nr. 11): „Verbindungen ohne laufenden Request schließt `Schliessen` sofort, die mit einem laufenden nach dessen Antwort“; `AGENTS.md` §3.13 (eine offene Entscheidung steht in §6 mit Ausgang *offen* und Adresse), §3.12 (der Implementer entscheidet nicht im Code oder im Test).
- `befund`: der Implementer meldet in §6 („Abweichung bei F-1“) und im Bericht, dass eine Verbindung mit halbem Body nach der Antwort erst schließt, wenn der Server den Rest erhalten hat; der Test sendet deshalb die fehlenden 97 Bytes. Das ist richtig beobachtet und von mir bestätigt (Sonde `TestSondeHalberBodyOhneRest`, Container `vadl-8af-probe`, nicht im Repo): Handler hängt, `Schliessen(30 s)`, Freigabe, Antwort 200 kommt, `Schliessen` kehrt mit `nil` zurück; der Client sendet den Rest nicht; **4 s später ist die Verbindung noch offen** (`i/o timeout` beim Lesen, nicht EOF). Nach der Quelle von `net/http` liest der Server den ungelesenen Rest (bis 256 KiB) nach der Antwort und endet frühestens mit dem `ReadTimeout` von 30 s. Der Vertrag sagt „nach dessen Antwort“; er nimmt diese Form weder in Nr. 11 noch in *Ungeprüft* aus, nur die Rückgabe von `Schliessen` hält sie nicht auf (Nr. 14, Satz „Die Frist gilt den zugelassenen Handlern …“). Der Plan führt die Zeile als **Ausgang: entfallen** (Implementer); ein Negativ (N-16) mit Grund und Folge steht nicht, ein Architect hat nicht entschieden.
- `Auflage an den Architect`, danach an den Planner (Plan nachziehen im selben Commit, §3.9): entweder `SPEC-238` Nr. 11 / *Ungeprüft* nennt die Form („eine Verbindung mit ungelesenem Rest eines Bodys bleibt bis zum Rest oder bis `ReadTimeout` offen“) und der Plan trägt N-16 mit Folge (höchstens eine Verbindung je solchem Request bis 30 s nach `Schliessen`, ohne Handler und ohne Wirkung auf den Rest, wie N-7), oder der Code schließt sie (`SetReadDeadline` auf der Verbindung). Der Ausgang der Zeile in §6 ist bis dahin *offen*, Adresse Architect.

### V3 · LOW — Die Zählprüfung „7 Zustandsfelder“ ist ein Literal, K5 verspricht mehr

- `quelle`: Plan §3, Zeile K5: „Zählprüfung gegen die Zustandsfelder von `Server`“; `AGENTS.md` §3.15 (Zählprüfung gegen die Quelle).
- `befund`: `TestZweiServerTeilenNichts` prüft `len(felder) == 7` gegen die eigene Liste, nicht gegen die Felder des Typs (`Server` trägt zwölf Zustandsfelder: `listener`, `http`, `abbruch`, `mu`, `angehalten`, `durchAnhalten`, `acceptFehler`, `geschlossen`, `laufend`, `leer`, `leerZu`, `verb`; die Felder der Konfiguration sind schreibgeschützt). Eine Mutation am achten Element, das der Plan nicht nennt: `mu` als Variable des Pakets (zwei Server teilen die Sperre): **0 von 10 rot**. Gegenprobe: `http` als Variable des Pakets 10 von 10 rot, `abbruch` allein 10 von 10 rot.
- `folge`: ein künftiges Zustandsfeld, das zwei Server teilen, bliebe grün; die geteilte Sperre allein ist ohne Wirkung auf die Zusagen (ein Server wartet, solange der andere die Sperre hält, keine Sperre liegt über einem Aufruf von Handler oder Vergleich), aber sie widerspricht „zwei Server teilen nichts“.
- `Auflage an den Planner`: entweder die Zeile K5 nennt die Zählprüfung als Literal mit Grund und `mu` als benanntes Negativ (kein Fänger, Folge: keine Wirkung auf eine Zusage), oder der Test zählt die Felder von `Server` per `reflect` (eine Zeile) und das Negativ entfällt.

### V4 · LOW — Das Plan-Negativ N-11 nennt eine Rate, die ich nicht bestätige

- `befund`: N-11 sagt „2 von 20 Prozessen von `TestZaehlerUnterLast`“ für den Zähler ohne Sperre ohne Detektor. Gemessen (`verlasse` verringert den Zähler vor dem Sperren, Testbinary ohne `-race`, 20 Prozesse, Loadavg 8 bis 17): **13 von 20 rot**. Mit dem Detektor 12 von 12.
- `folge`: die Schlussfolgerung („die Zusage steht auf dem Detektor von `make test-race`“) bleibt richtig; die Zahl des Plans ist eine Größenordnung unter meiner. Die Rate hängt an der Last (`AGENTS.md` §3.10: eine Einzelzahl gilt als Größenordnung); der Plan sollte beide Messungen nennen.
- `Auflage an den Planner`: in N-11 „2 von 20 (Implementer), 13 von 20 (Verifier)“ mit Messendem und Last.

### V5 · LOW — N-15 nennt eine Adresse ohne Nehmer

- `quelle`: `AGENTS.md` §3.13 (eine Adresse nimmt an, bevor sie genannt wird).
- `befund`: N-15 sagt „keine Folgepflicht“ und im selben Absatz „ein künftiger Slice am Paket `httpodata` übernimmt die Klasse aus `lebenszyklus_accept_test.go`“. Letzteres ist eine Zuweisung an einen Slice, den es nicht gibt. Das Negativ selbst stimmt (Mutation in `httpodata/server.go:162` grün).
- `Auflage an den Planner`: den Halbsatz „ein künftiger Slice … übernimmt“ streichen (das Negativ steht dann allein, ehrlich) oder die Klasse in der DoD oder §1 eines benannten Slice führen.

Sonst: Zusage ohne Fänger und ohne Negativ habe ich nicht gefunden. Die Negative N-1 bis N-14 tragen Grund und Folge; N-3, N-4, N-5, N-12 und N-14 habe ich nachgemessen (§5).

## 4. Fix-Runde gegen die Klasse aufgelöst (§3.17)

Je Befund die Menge aus dem Wortlaut der Zusage, ihre Elemente, und Mutationen an Elementen, die weder der Befund noch der Fix nannten (Messender: Verifier, Prozesse mit `-race -test.count=1`, Regex auf die Tests der Menge).

| Befund | Menge (Wortlaut) | Elemente | Mutation an einem ungenannten Element | Ergebnis |
|---|---|---|---|---|
| F-1 (S11) | „Verbindungen mit laufendem Request schließen nach dessen Antwort“: Verbindungszustände mit Handler | 2 (laufender Request, halber Body); die Menge der Zustände insgesamt 5 | `Shutdown` entfällt (`_ = beendet` statt des Aufrufs) | 10 von 10 rot, beide Elemente („nach der Antwort … erwartet EOF“) |
| | | | `markiereHandler` markiert nur Requests ohne Body (`handler = r.ContentLength == 0`): halber Body wird sofort geschlossen | 5 von 5 rot (`TestVerbindungszustaende`) |
| | | | `StateNew` legt die Verbindung mit `handler: true` an (neu ohne Byte und halber Kopf bleiben offen) | 5 von 5 rot |
| | | | `StateActive` löscht die Verbindung aus der Menge | **0 von 5, äquivalent**: `Shutdown` schließt Verbindungen im Leerlauf selbst, eine laufende schließt `net/http` nach der Antwort (Klasse N-14) |
| F-2 (S8) | „höchstens `frist`“ nur als Untergrenze: Wartezeit bis zum Ende von `ctx` | 4 Teilfälle (2 Fristen × GET, POST) | `frist * 9 / 10`; `frist - 20 ms`; `frist - 2 ms` (der Befund nannte `/ 4`, der Fix `/ 2`) | je 5 von 5 rot (`TestSchliessenKontextNachFrist`); die Schranke ist bis auf wenige ms scharf, keine Obergrenze (N-4 stimmt) |
| F-3 (K5) | Zustandsfelder, die zwei Server nicht teilen | 7 genannt, 12 vorhanden (V3) | je als Variable des Pakets: `angehalten`, `geschlossen`, `laufend`, `leerZu` allein, `durchAnhalten`, `acceptFehler`, `abbruch` allein, `http` | je **10 von 10 rot**, und zwar im Teilfall des eigenen Feldes (z. B. `geschlossen` im Fall „geschlossen“, `laufend` und `leerZu` im Fall „Zähler samt leer“, `abbruch` im Fall „Kontext samt abbruch“) |
| | | | `mu` als Variable des Pakets | **0 von 10 rot** (V3) |
| F-5 (S2) | „letzter Fehler des Accept; temporäre Fehler wiederholt `net/http`“: Folge eines temporären Fehlers | 4 | Bedingung `durchAnhalten` aus dem Löschen entfernt | 3 von 3 rot (`TestServeFolgeTemporaererFehler` und drei weitere) |
| | | | `Accept` reicht den Fehler ohne `Temporary()` weiter (`errors.New(err.Error())`): `Serve` endet am temporären Fehler | 3 von 3 rot (`TestServeFolgeTemporaererFehler`, Frist 5 s) |
| | | | nur `Schliessen` löscht, `Anhalten` nicht | 3 von 3 rot |
| F-6 (S7) | „ändert nichts“: Zustände × Fristen, und die Teile des Zustands | 6 Elemente, 6 Zustandsteile (V1) | Zustand vor dem Fehler gesetzt: `geschlossen`; `abbruch()` | **je 0 von 5 rot** (V1) |
| | | | Listener geschlossen vor dem Fehler; `schliesseOhneHandler()` vor dem Fehler; Grenze `frist < -1` | je 5 von 5 rot |
| F-7 | Kommentar | – | gelesen: Wortlaut gleich dem Auftrag | erfüllt |
| F-8 | Kopf von `TestFuenfzigHandlerEndenGleichzeitig` | – | gelesen: nennt allein S13, `Anhalten` läuft nach den 50 Antworten | erfüllt |

**Bewertung der zwei Hänger der Mutationsläufe des Implementers** („Paketvariablen, Fristen im Test gesetzt“): im Repo steht keine Paketvariable für Fristen in den Tests (`grep` über die Testdateien: allein lokale `fristen`-Listen). Der Hänger ist eine Eigenschaft des Mutationsaufbaus (ein Zustandsfeld als Variable des Pakets teilen alle Server und alle Tests eines Prozesses, ein Haltepunkt eines Tests kann dann nie freigegeben werden). Ich fuhr dieselbe Klasse (acht Felder als Variable, je 10 Prozesse, dazu `mu`): **0 Zeitüberschreitungen** bei 120 s je Prozess; jede Wartestelle der Tests trägt eine Frist (`warteBis` 5 s, `antwort` 5 s Lesefrist, `serveEnde`, `erwarte`). Ein Hänger bleibt rot (Go bricht nach 10 min ab); kein Befund.

## 5. Nebenläufige Raten und Negative selbst gefahren (§3.10)

Messender: Verifier, Last des Hosts Loadavg 3 bis 17, `-race -test.count=1` und je Prozess das ganze Paket `admin` (Regex `.`), wenn nicht anders angegeben. „rot“ steht für den benannten Test.

| Zusage | Mutation | roter Test | rot von N | Plan |
|---|---|---|---|---|
| S3 | `Close` des Listeners in `haltAnLocked` außerhalb der Sperre (`Unlock` / `Close` / `Lock`) | `TestAnhaltenWiederholt` (dazu `TestSchliessenOhneAnhalten`, `TestServeEndetMitNil`, `TestServeFehlerVorAnhalten`) | 12 von 12 | 20 von 20 |
| S6 Fall 0 und 1 | `zulassen`: Sperre zwischen Prüfen und Zählen frei, mit `KorrelationsID` im Fenster | `TestZulassungAtomar/Fall_0`, `…/Fall_1`, `TestAngehaltenDreiNullDrei`, `TestKorrelationsHeader` | 12 von 12 | 20 von 20 |
| S6 (N-3) | dasselbe Fenster ohne Aufruf (nur `Unlock`; `Lock`) | keiner | **0 von 12** | 0 von 20 |
| S13 | `verlasse`: Zähler vor dem Sperren verringert (Detektor) | `TestZaehlerUnterLast`, `TestFuenfzigHandlerEndenGleichzeitig`, `TestZulassungAtomar`, `TestSchliessenFristen`, `TestVerbindungszustaende` | 12 von 12 | 20 von 20 |
| S13 | `zulassen`: `angehalten` vor dem Sperren gelesen | `TestZulassungAtomar` und weitere | 12 von 12 | 20 von 20 |
| S13 (N-5) | `schon := s.geschlossen` außerhalb der Sperre gelesen, Schreiben unter der Sperre | `TestSchliessenZweimal` (Detektor) | **5 von 12** | 2 von 10 / 4 von 6 (Reviewer): gleiche Größenordnung |
| S13 (N-11) | `verlasse` ohne Sperre, Testbinary **ohne** `-race`, 20 Prozesse | `TestZaehlerUnterLast` | **13 von 20** | 2 von 20 (V4) |
| K5 / F-3 | je Zustandsfeld als Variable des Pakets | `TestZweiServerTeilenNichts` (eigener Teilfall) | 10 von 10 für 8 Felder, 0 von 10 für `mu` | je 10 von 10 |
| X2 / S9 (N-12) | `BaseContext` entfernt; `abbruch` im Ablauf der Frist nicht gerufen (das `defer` bleibt) | `TestSchliessenKontextNachFrist`, nur die POST-Teilfälle (beide Fristen); die GET-Teilfälle bleiben grün | 10 von 10; 10 von 10 | N-12 bestätigt |
| N-14 | `delete(s.verb, c)` im Zustand `StateActive` | keiner | 0 von 5 | äquivalent, bestätigt |

Die Zusagen über Nebenläufigkeit S3, S6 (Fall 0 und 1), S13 (Detektor) werden in jedem Lauf rot; S6 ohne Aufruf im Fenster (N-3), das Lesen von `geschlossen` außerhalb der Sperre (N-5) und der Zähler ohne Detektor (N-11) sind die benannten Negative und stehen im Plan mit Grund und Folge. N-4 (Prüfgrenze der Fristen) stimmt jetzt mit dem überein, was `TestSchliessenKontextNachFrist` fängt (§4, F-2); N-13 (`EnableFullDuplex` ohne Fänger) bestätigt: ohne den Aufruf bleiben `TestAngehaltenDreiNullDrei` und eine Sonde über sechs Größen des angekündigten, nicht gesendeten Bodys (100 Bytes bis 1 MiB, mit Grenze 262144 und 262145) grün, weil ein 503 mit `Connection: close` den Body ohnehin nicht verwirft.

## 6. Auflagen

**An den Implementer (kleine Fix-Runde, nur Tests, danach kein Review nötig, wenn die Mutationen rot sind):** V1.

**An den Architect:** V2 (Entscheidung zu halbem Body nach der Antwort; Wortlaut in `SPEC-238` Nr. 11 / *Ungeprüft* oder Code-Änderung).

**An den Planner (Plan nachziehen, `AGENTS.md` §3.9, ein Commit):** V2 (N-16 oder Änderung, Ausgang der Zeile *offen* bis zur Entscheidung), V3 (K5), V4 (N-11), V5 (N-15); die Zeile „Aufträge der Fix-Runde … Ausgang: offen; Adresse: Implementer“ auf *entschieden* setzen, wenn V1 liegt; Closure-Notiz, Register, Ausgänge der Risiken und die DoD-Haken wie bei jedem Slice mit der Closure.

## 7. Messung und Gegenlesen

- Eigenes Image `verif-adl-8af:race`, Container `vadl-8af-mut` und `vadl-8af-probe` (`--memory` 4g und 3g), nach dem Bericht entfernt; kein Docker-Kill nach Image.
- Das Sonden-Paket (`TestSondeAngekuendigterBody`, `TestSondeHalberBodyOhneRest`) lag nur im Wegwerfbaum des Containers; im Repo steht davon nichts.
- Dieser Bericht ist nach dem Schreiben auf Rohbytes gegengelesen (`LC_ALL=C grep -nP '\\u[0-9a-fA-F]{4}|\xEF\xBB\xBF|\xC2\xA0'`): 0 Treffer über diese Datei.
