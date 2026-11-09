# Review-Report: slice-admin-lebenszyklus — 2026-11-08

**Review-Art:** Code (gegen Plan, `SPEC-238` Nr. 1 bis 15, `SPEC-136` Nr. 12 bis 14, `LH-FA-13.a`, Hard Rules `AGENTS.md` §3.2, §3.6, §3.7, §3.9 bis §3.18)

**Gegenstand:** Zweig `slice-admin-lebenszyklus`, Commits `df270bcd`, `08670f19` und `4590f30a` gegen die Basis `fb575e00`, Pfad `src/adapters/driving/admin/` (2092 Zeilen nach `git diff --numstat`, Marke 2700; nachgerechnet: 2061 Einfügungen, 31 Löschungen).

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-08

**Eingangs-Kontext:**

- Plan `slice-admin-lebenszyklus` (Tabelle Zusage · Fänger mit 19 Zeilen, Raster mit 8 Zeilen und Zusatztabelle mit 9 Paaren in §6, Fenster W1 bis W6, Negative N-1 bis N-12, Mutationen und Raten des Implementers in §3)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `LH-FA-13.a`, `LH-QA-04.a`, `LH-RB-04`
- `AGENTS.md` (Hard Rules)

Vorgehen: Code (`lebenszyklus.go`, `admin.go`) und Tests gelesen, gegen den Wortlaut von `SPEC-238` Nr. 3 b, 9, 10, 11, 12 und `SPEC-136` Nr. 12 bis 14 abgeglichen, den Code mit dem Vorbild `httpodata` verglichen (`merkListener`, `merkeZustand` und `markiereHandler` sind dieselbe Bauart). Die Mutationslisten des Implementers dienten als Anhalt; 52 eigene Mutationsläufe an Elementen und Klassen, die die Liste nicht nennt oder nur einmal führt, in der Wegwerfkopie des Pakets `admin` (Race-Image, ohne Netz, Quellen schreibgeschützt eingehängt, `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, ein Container zur Zeit; Prozess = Aufruf des gebauten Testbinary mit `-race` und `-test.count=1`, wo nicht anders vermerkt; Loadavg (1 Minute) 6 bis 19 während der Läufe, drei Slices parallel; Messender: der Reviewer). Ungemutierter Lauf von `go test -race -count=1` über das Paket: grün. Nicht Gegenstand: `TestWettlaufNebenlaeufig` (Teil 1); er ist in einem von sechs Läufen einer Mutation (R44) rot geworden und wurde hier nicht weiter verfolgt.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `SPEC-136` Nr. 14 und S11 sagen zu, dass `Schliessen` die Verbindung mit einem laufenden Request nach dessen Antwort schließt. Die Mutation, die den Aufruf `Shutdown` entfernt, bleibt grün (0 von 1 im Gesamtlauf): kein Test liest nach der Antwort des laufenden Requests, ob die Verbindung geschlossen ist (`TestVerbindungszustaende` und `TestLeerlaufUndAbgewiesene` lesen allein den Status 200). Folge: nach `Schliessen` bleibt die Verbindung bis `IdleTimeout` offen; Verbindungen im Leerlauf schließt `schliesseOhneHandler` unabhängig davon. | `AGENTS.md` §3.15 (Zusage ohne Fänger) | `src/adapters/driving/admin/lebenszyklus.go` · `_ = s.http.Shutdown(beendet)` | ja — `go test -race` mit der Mutation | Zusage ohne Fänger: Zustand der Verbindung nach der Antwort |
| F-2 | MEDIUM | Die Mutation, die die Wartezeit von `Schliessen` auf ein Viertel der Frist kürzt (`time.NewTimer(frist / 4)`), bleibt grün (0 von 1); erst ein Deckel von 150 ms wird rot (über `TestVerbindungszustaende`). N-4 nennt 200 ms als geprüft und sagt, die Ränder 0, 30 s und `math.MaxInt64` fingen ein Schließen ohne Warten; `TestSchliessenKontextNachFrist` (200 ms) misst die Dauer nicht, und `TestSchliessenFristen` (30 s, Freigabe nach 100 ms) verlangt nur, dass die Wartezeit 100 ms übersteigt. Ein Handler, der nach 60 Prozent der Frist endet, bekäme `ErrFristAbgelaufen` statt nil. | `AGENTS.md` §3.15 (Fänger schmaler als Zusage, Prüfgrenze falsch benannt); Plan N-4 | `src/adapters/driving/admin/lebenszyklus.go` · `timer := time.NewTimer(frist)` | ja — `go test -race` mit der Mutation | Prüfgrenze der Frist falsch beschrieben |
| F-3 | MEDIUM | K5 sagt „zwei Server teilen nichts“ und nennt als Mutation „Zustand als Variable des Pakets“. `TestZweiServerTeilenNichts` führt allein `angehalten` und den Listener: der Kontext des Servers samt `abbruch` als Variable des Pakets bleibt für diesen Test grün (0 von 1); rot werden dabei nur Tests, die in demselben Prozess später laufen (`TestSchliessenKontextNachFrist`, `TestSchliessenFristen`, `TestLaufenderRequestEndet`), nicht aus dem Grund von K5. Zähler, `geschlossen` und `leer` sind ebenso nicht als Menge geführt (die Verbindungsmenge als Variable macht K5 rot, 3 von 3, über den Detektor). | `AGENTS.md` §3.15 (Fänger führt eine Form der Eingabe), §3.17 | `src/adapters/driving/admin/grenzen_test.go` · `func TestZweiServerTeilenNichts` | ja — `go test -race -run TestZweiServerTeilenNichts` mit der Mutation | Allaussage über Zustandsfelder an einem Element geprüft |
| F-4 | MEDIUM | S13 nennt `geschlossen` als Zustand unter der Sperre. Liest und schreibt `Schliessen` es außerhalb der Sperre, wird allein `TestSchliessenZweimal` rot, und nur mit dem Detektor: 2 von 10 Prozessen (gezielter Lauf) und 4 von 6 (Gesamtlauf). Das Schreiben außerhalb der Sperre allein ist 6 von 6 rot. Weder ein Wirkungs-Fänger noch ein Negativ benennt das Lesen; eine Nebenläufigkeits-Zusage gilt nach §3.10 nur als gefangen, wenn die Mutation in jedem Lauf rot wird. | `AGENTS.md` §3.10 | `src/adapters/driving/admin/lebenszyklus.go` · `schon := s.geschlossen` | ja — `go test -race` je Prozess mit der Mutation | Nebenläufigkeits-Mutation nur mit Wahrscheinlichkeit gefangen, Rate nicht benannt |
| F-5 | MEDIUM | `Accept` löscht den vermerkten Fehler, wenn danach das Schließen durch `Anhalten` folgt („der letzte Fehler gilt; auf einen Fehler, den net/http wiederholt, folgt ein weiteres Accept“). Die Mutation, `acceptFehler` dann nicht zu löschen, bleibt grün (0 von 1): die Klasse „jede Art des Fehlers“ (S2, 35 und 4 Elemente) führt keinen temporären Fehler, und der Plan nennt die Auslassung als Eigenschaft des Fakes, nicht als Negativ mit Grund und Folge. `SPEC-136` Nr. 12 sagt „nie nil“ für einen Fehler vor `Anhalten` und „der letzte Fehler“; der temporäre Fall steht nirgends. | `AGENTS.md` §3.15, §3.16 (Zelle im Code entschieden), §3.11 (Kommentar über Wiederholung ohne Test) | `src/adapters/driving/admin/lebenszyklus.go` · `if l.s.durchAnhalten && errors.Is(err, net.ErrClosed) {` | ja — `go test -race` mit der Mutation | Art des Fehlers der Klasse ausgelassen (temporärer Fehler des Accept) |
| F-6 | MEDIUM | `Schliessen(frist)` mit negativer Frist auf einem schon geschlossenen Server: die Mutation, die die Prüfung hinter die Rückgabe des zweiten Aufrufs setzt (nil statt `model.ErrUngueltig`), bleibt grün (0 von 1). `TestSchliessenNegativ` führt nur den frischen Server; das Paar (negative Frist, zweites `Schliessen`) steht weder im Raster (Zeile `Schliessen · frist`) noch in der Zusatztabelle. | `AGENTS.md` §3.16 (Paar ohne Fänger) | `src/adapters/driving/admin/lebenszyklus.go` · `if frist < 0 {` | ja — `go test -race` mit der Mutation | Paar von Ursachen ohne Fänger (negative Frist × geschlossen) |
| F-7 | LOW | Der Kommentar von `ServeHTTP` sagt, mit Vollduplex verwerfe net/http den ungelesenen Rest des Bodys nicht; die Zeile gilt für jede Antwort des Servers. Wird `EnableFullDuplex` entfernt, bleibt jeder Test grün (0 von 1), auch die Form „POST angekündigt 1 MiB, 10 gesendet“; `SPEC-238` verlangt es nicht (der Grund in `httpodata`, `SPEC-136` Nr. 3, 4 h, hat im Admin-Adapter keine Entsprechung). | `AGENTS.md` §3.7, §3.11 | `src/adapters/driving/admin/admin.go` · `_ = http.NewResponseController(w).EnableFullDuplex()` | ja — `go test -race` mit der Mutation | Kommentar nennt eine Wirkung, die kein Test zeigt |
| F-8 | LOW | Der Kopf von `TestFuenfzigHandlerEndenGleichzeitig` nennt „S6 Fall 2“, den der Plan nicht führt (S6 hat Fall 0 und Fall 1), und sagt „ohne dass Anhalten die meisten abweist“; der Test ruft `Anhalten` erst, nachdem alle 50 Antworten da sind. | `AGENTS.md` §3.7, §3.11 | `src/adapters/driving/admin/lebenszyklus_test.go` · `hält S13 und S6 Fall 2` | nein | Kommentar nennt mehr als der Test |
| F-9 | LOW | §3 des Plans (Zeile *Tests*) nennt `lebenszyklus_test.go`, `lebenszyklus_accept_test.go`, `grenzen_test.go`, `lebenszyklus_fake_test.go` und die Fakes, nicht `export_test.go` (neue Hilfe `GrenzenFuerTest`, ein Zugriff des Tests auf `s.http`) und nicht `helfer_test.go` (`objekt503`). | `AGENTS.md` §3.9 | `docs/plan/planning/done/slice-admin-lebenszyklus.md` · `Tests | neu / update` | nein | Plan zieht eine geänderte Datei nicht nach |
| F-10 | INFO | Drei Mutationen bleiben grün, ohne dass eine Zusage des Vertrags daran hängt: das Zurücksetzen von `handler` im Zustand `StateIdle` (äquivalent, weil `Shutdown` Verbindungen im Leerlauf selbst schließt), das Auslassen von `delete(s.verb, c)` (Speicherleck, kein Fänger) und `MaxBytesReader` nur für `POST` (der Code behandelt Methoden gleich; G3 führt allein `POST`). Der Kommentar von `merkeZustand` („jede ist von der Annahme bis zum Ende bekannt, und ein Handler gilt nur, bis die Verbindung wieder im Leerlauf ist“) sagt die ersten beiden zu. Zuständig: Implementer, auf Wunsch als Negativ im Plan. | `AGENTS.md` §3.11 | `src/adapters/driving/admin/lebenszyklus.go` · `delete(s.verb, c)` | ja | Zusage im Kommentar ohne Fänger, ohne Folge im Vertrag |
| F-11 | INFO | Die Historie-Zeile in `spec/spezifikation.md` ist die einzige Änderung an der Spezifikation (Diff: eine Zeile; die Kennung des Slice im Text wurde zu „vor dem Code des Lebenszyklus des Admin-Servers“). `SPEC-238` Nr. 1 bis 15 und *Ungeprüft* sind byte-gleich; der Vertrag blieb unberührt. Der Implementer hat eine Datei des Architect geändert; die Abweichung steht in §6 mit Ausgang *offen* und Adresse Architect (`AGENTS.md` §3.13 erfüllt). Zuständig: Architect (Bestätigung der Fassung). | `AGENTS.md` §3.13 | `spec/spezifikation.md` · `vor dem Code des Lebenszyklus des Admin-Servers` | ja — `make docs-check` | Rollengrenze, Abweichung als offen geführt |
| F-12 | INFO | X2/S9 (bekannt, Planner/Architect): die Mutationen „`BaseContext` nicht gesetzt“ und „`abbruch` nicht gerufen“ bleiben im Teilfall GET grün, im Teilfall POST mit ungelesenem Body rot (selbst nachgefahren: `BaseContext` Q3 1 von 1 rot, `context.WithoutCancel` an der Route R13 rot in drei Tests, Abbruch samt Schließen entfernt P10 rot). Der Kontext des Servers ist damit gefangen; die Zusage „der `ctx` endet mit der Frist“ trägt im GET-Fall das Schließen der Verbindung. N-12 nennt Grund und Folge und trägt als Negativ. Offen bleibt allein die Spalte *Mutation* von X2 im Plan (§6, Adresse Planner). | `AGENTS.md` §3.15 | `docs/plan/planning/done/slice-admin-lebenszyklus.md` · `Abweichung vom Plan (Implementer, 2026-11-08): die Zeile X2` | nein | Plan-Spalte nennt eine Mutation, die der Teilfall nicht fängt |

## Bewertung der Negative und der Raten des Implementers

| Negativ oder Rate | Eigene Messung | Urteil |
|---|---|---|
| N-3 (S6 ohne Aufruf im Fenster, 0 von 20) | R25: 0 von 10 mit Detektor; `zulassen` ist eine Funktion mit einem Abschnitt unter einer Sperre | trägt: Grund (kein Argument des Tests liegt im Fenster, der Detektor sieht nichts unter einer Sperre) und Folge benannt, der Prüfpunkt des Reviews („die eine Funktion der Zulassung“) ist erfüllt; mit Aufruf im Fenster ist S6 Fall 0 rot (P4, 1 von 1; deterministisch) und Fall 1 rot in 10 von 10 (R28) |
| N-11 (S13 ohne Detektor, 2 von 20) | R24: 6 von 20 ohne Detektor (`TestZaehlerUnterLast`); mit Detektor R20, R21, R22 je 10 von 10 | trägt: die Zusage steht auf dem Detektor von `make test-race` (Teil der Kette); die Rate ohne Detektor schwankt mit der Last (2 und 6 von 20), die Größenordnung „wenige von 20“ stimmt |
| N-12 / X2 Fall 2 | siehe F-12 | trägt |
| N-5 (gleichzeitiges `Schliessen`) | R44: Schreiben von `geschlossen` außerhalb der Sperre 6 von 6; R27, R27b: Lesen außerhalb 2 von 10 und 4 von 6 | trägt für das Verhalten (Panik, Datenrennen beim Schreiben); das Lesen von `geschlossen` ist F-4 |
| N-4 (Prüfgrenze der Fristen) | R1, R2 | trägt nicht in der beschriebenen Grenze, siehe F-2 |
| N-1, N-2, N-6 bis N-10 | Grund und Folge gelesen; N-10 gegen „`durchAnhalten` immer falsch“ (Implementer) | tragen |

## Auflösung der Tabelle Zusage · Fänger, des Rasters und der Fenster

Je Zeile mindestens eine eigene Mutation (Reihe in Klammern, Tabelle im Anhang); „rot“ heißt: aus dem benannten Test rot, außer wo vermerkt.

| Zeile | Eigene Mutation | Ergebnis |
|---|---|---|
| A2 | `Connection: close` fehlt (Q1) | rot, 1 von 1 (`TestAngehaltenJedeRequestart`, `TestHeaderJeAntwortart`, `TestAngehaltenDreiNullDrei`) |
| K5 | Kontext und `abbruch` als Variable des Pakets (R30b); Verbindungsmenge (R31b) | R30b grün für K5 (F-3); R31b rot 3 von 3 |
| S1 | `http.ErrServerClosed` durchgereicht (P1) | rot, 1 von 1 (`TestServeEndetMitNil`, `TestServeFehlerKlasse`, `TestServeFehlerVorAnhalten`, `TestSchliessenOhneAnhalten`) |
| S2 | Fehler vor `Anhalten` zu nil (P2); `acceptFehler` behalten (R5) | P2 rot, 1 von 1 (`TestServeFehlerVorAnhalten`, `TestServeFehlerKlasse`); R5 grün (F-5) |
| S3 | `Anhalten` wartet auf Handler (P3); Flag und `Close` außerhalb der Sperre (R23) | P3 rot (`TestAnhaltenSchliesstListener`, dazu Zeitlimit des Gesamtlaufs); R23 rot, 10 von 10 (`TestAnhaltenWiederholt`) |
| S4 | Handler liest Body vor dem 503 (Q2) | rot, 1 von 1 (`TestAngehaltenDreiNullDrei`, `TestLeerlaufUndAbgewiesene`) |
| S5 | `Anhalten` bricht `ctx` ab (Q6) | rot, 1 von 1 (`TestLaufenderRequestEndet`) |
| S6 | Fall 0 mit Aufruf im Fenster (P4); Fall 1 (R28); ohne Aufruf (R25) | P4 rot, 1 von 1 (`TestZulassungAtomar`); R28 rot 10 von 10; R25 grün, N-3 |
| S7 | `frist <= 0` (Q7); negative Frist nach `Schliessen` (R6) | Q7 rot (`TestSchliessenNegativ`, `TestSchliessenFristen`); R6 grün (F-6) |
| S8 | Frist durch 4 (R1); Deckel 150 ms (R2); nil nach Ablauf (Q10) | R1 grün (F-2); R2 und Q10 rot (`TestSchliessenFristen`, `TestVerbindungszustaende`) |
| S9 | `BaseContext` nicht gesetzt (Q3); Rückkehr vor dem Ende des Handlers (P5) | Q3 rot, 1 von 1 (nur POST, N-12); P5 rot, 1 von 1 (`TestSchliessenKontextNachFrist`, `TestSchliessenFristen`, `TestKeineSperreUeberAufruf`) |
| S10 | zweites `Schliessen` wartet (Q5) | rot, 1 von 1 (`TestSchliessenZweimal`) |
| S11 | `Anhalten` schließt Leerlauf (Q9); `Shutdown` entfernt (R3); Markierung nach der Zulassung (R8) | Q9 rot (`TestLeerlaufUndAbgewiesene`); R3 grün (F-1); R8 rot (`TestZulassungAtomar`) |
| S12 | `Schliessen` setzt das Flag, schließt den Listener nicht (Q4, P11) | Q4 im Gesamtlauf Zeitlimit (rot, nicht aus dem benannten Test); P11 mit dem benannten Test allein rot, 1 von 1 (`TestSchliessenOhneAnhalten`) |
| S13 | `angehalten` ohne Sperre gelesen (R20); Zähler ohne Sperre verringert (R21); Zähler zwischen Entsperren und Sperren erhöht (R22); `geschlossen` (R27, R44); `durchAnhalten` (R41); Verbindungsmenge (R43); Sperre über Handler (Q8, P12); Sperre über `Vergleich` (P9); `verlasse` ohne `defer` (R7) | R20 bis R22 je 10 von 10 mit Detektor; R44, R41, R43 je 6 von 6; R27 2 von 10 (F-4); R7 rot (`TestHandlerPanikZaehltNichtWeiter`); Q8 im Gesamtlauf Zeitlimit, P12 mit dem benannten Test allein rot 3 von 3 (`TestKeineSperreUeberAufruf`); P9 rot, 1 von 1 (`TestKeineSperreUeberAufruf`) |
| X2 | `context.WithoutCancel` an der Route (R13); Abbruch und Schließen entfernt (P10) | R13 rot (`TestKontextEndetBeimTrennen`, `TestSchliessenKontextNachFrist`, `TestSchliessenFristen`); P10 rot (`TestSchliessenKontextNachFrist`, `TestSchliessenFristen`) |
| G1 | `ErrorLog` nicht gesetzt (R32); `ReadTimeout` 3 s (P6) | R32 rot aus `TestKeinLogDerBibliothek` allein; P6 rot, 1 von 1 (`TestWerteDesServers`) |
| G2 | allgemeiner OPTIONS-Handler an (P8) | rot, 1 von 1 (`TestOptionsStern`, `TestWerteDesServers`, `TestAngehaltenDreiNullDrei`) |
| G3 | Grenze ein Byte mehr (P7); Begrenzung nur `POST` (R12) | P7 rot, 1 von 1 (`TestBodyGrenze`); R12 grün (F-10) |

Raster (8 Zeilen, je Zelle der Test): `Serve · Listener` S2 mit der Auslassung F-5, S1 ohne Befund; `Anhalten` S3; `Schliessen · frist` S7 (F-6), S8 (F-2), S12, S9, S10; `ServeHTTP · angehalten` A2, S4 (50 Elemente, Zählprüfung im Test); `OPTIONS *` G2, S4; `Schliessen · Verbindung` S11 (F-1); `Route · ctx` X2, S9; `Route · Body` G3. Keine Zelle ohne Ort in der Spezifikation; keine im Code entschiedene Zelle außer dem temporären Fehler des Accept (F-5). Zusatztabelle (9 Paare): das Paar „Body über 1 MiB und `Schliessen` mit Ablauf der Frist“ nennt S9 und G3 getrennt, nennt als Ergebnis „Kontextfehler oder `*http.MaxBytesError`“ (Abbildung ist Sache von `slice-admin-fehlerabbildung`) und hat keinen gemeinsamen Fall; kein Befund, weil das Ergebnis offen gelassen ist.

Fenster: W1 N-3 (trägt, siehe oben) und S6 Fall 0/1; W2 S3 (R23 10 von 10; das Schließen des Listeners in `Schliessen` außerhalb der Sperre, R11, wird über `TestServeEndetMitNil` in 3 von 3 Läufen rot, nicht über S3: die Wirkung auf `Serve` fängt es); W3 und W4 benannte Negative, die Begründung stimmt mit dem Code überein (nach dem Setzen weist `zulassen` jeden Request ab, der Zähler kann nur sinken); W5 kein Fenster (Empfang auf einen Kanal, den `meldeLeerLocked` unter derselben Sperre schließt); W6 S2 (P2 rot). Nicht als Fenster geführt, aber im Code: die Schnappschuss-Lücke in `schliesseOhneHandler` (Sperre freigegeben vor dem Schließen); sie ist N-7 und ohne Folge für zugelassene Requests, weil `markiereHandler` vor `zulassen` läuft (Mutation R8 wird rot).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driving/admin/lebenszyklus.go` gegen `SPEC-136` Nr. 12 bis 14 und `SPEC-238` Nr. 9, 10, 11, 12 | geprüft, ohne Befund am Code; F-1 bis F-6 betreffen Fänger |
| `src/adapters/driving/admin/admin.go` (`New`, `ServeHTTP`, `rufe`) gegen `SPEC-238` Nr. 3 a, b, 10 | geprüft, ohne Befund am Code (Reihenfolge ID vor Zulassung vor Authentisierung; keine Sperre über `Handler` und `Vergleich`); F-7 Kommentar |
| Tests `lebenszyklus_test.go`, `lebenszyklus_accept_test.go`, `lebenszyklus_fake_test.go`, `grenzen_test.go`, `fakes_test.go` (Fakes halten Argumente und zählen jede Operation: `listenerFake` zählt `Accept`, `Close`, `Addr`; `routeFake` hält `Anfrage` und `ctx`; `zaehlendeIDs` mit Haltepunkt) | geprüft; F-1 bis F-6, F-8 |
| Hard Rules `AGENTS.md` §3.2 (keine `//nolint`-Direktive im Paket: Suche leer), §3.6 (kein Gate gelockert), §3.18 (kein Escape in den neuen Dateien; Umlaute sind UTF-8) | geprüft, ohne Befund |
| Größenmarke `AGENTS.md` §3.14 und `SPEC-128` (Block in §4, Messung 2092 gegen Marke 2700 nachgerechnet) | geprüft, ohne Befund |
| `spec/spezifikation.md` | geprüft, eine Zeile geändert (F-11) |
| Binary, `src/bootstrap`, andere Pfade unter `src/` | geprüft, unverändert (Diff über die Pfade leer) |

## Anhang: Läufe

Die Reihen R, Q und P sind eigene Mutationen des Reviewers; Q und P fahren das gesamte Paket (ein Prozess), R auch einzelne Tests. Der Hinweis des Skills gilt: `panic:` oder Zeitlimit statt `--- FAIL` des benannten Tests heißt rot, aber nicht aus ihm. Rohdaten und Bündel liegen im Scratchpad des Reviewers (Verzeichnis `rv-lz-5c1e`), nicht im Repo.

| Reihe | Mutation | Prozesse | rot |
|---|---|---|---|
| R1 | Wartezeit `frist / 4` | 1 | 0 (F-2) |
| R2 | Wartezeit auf 150 ms gedeckelt | 1 | 1 (`TestVerbindungszustaende`) |
| R3 | `Shutdown` entfernt | 1 | 0 (F-1) |
| R4 | `EnableFullDuplex` entfernt | 1 | 0 (F-7) |
| R5 | `acceptFehler` bei Schließen durch `Anhalten` nicht gelöscht | 1 | 0 (F-5) |
| R6 | negative Frist nach `Schliessen` ergibt nil | 1 | 0 (F-6) |
| R7 | `verlasse` ohne `defer` | 1 | 1 |
| R8 | `markiereHandler` nach der Zulassung | 1 | 1 |
| R9 | `delete(s.verb, c)` entfällt | 1 | 0 (F-10) |
| R10 | `handler` im Leerlauf nicht zurückgesetzt | 1 | 0 (F-10, äquivalent) |
| R11 | `Schliessen` schließt den Listener außerhalb der Sperre | 3 | 3 (`TestServeEndetMitNil`, `TestSchliessenOhneAnhalten`) |
| R12 | `MaxBytesReader` nur für `POST` | 1 | 0 (F-10) |
| R13 | `context.WithoutCancel` an der Route | 1 | 1 |
| R20 | `angehalten` ohne Sperre gelesen (Detektor) | 10 | 10 |
| R21 | Zähler ohne Sperre verringert (Detektor) | 10 | 10 |
| R22 | Zähler zwischen Entsperren und Sperren erhöht (Detektor) | 10 | 10 |
| R23 | Flag und `Close` außerhalb der Sperre | 10 | 10 |
| R24 | Zähler ohne Sperre verringert, ohne Detektor | 20 | 6 (beim Implementer 2 von 20) |
| R25 | Prüfen und Zählen in zwei Abschnitten ohne Aufruf im Fenster | 10 | 0 (N-3) |
| R27 | `geschlossen` außerhalb der Sperre, nur `TestSchliessenZweimal` | 10 | 2 (F-4) |
| R27b | wie R27, Gesamtlauf | 6 | 4 (F-4) |
| R28 | S6 Fall 1 mit Aufruf im Fenster | 10 | 10 |
| R30 | Kontext und `abbruch` als Variable des Pakets, Gesamtlauf | 1 | 1 (nur andere Tests) |
| R30b | wie R30, nur `TestZweiServerTeilenNichts` | 1 | 0 (F-3) |
| R31 | Verbindungsmenge als Variable des Pakets, Gesamtlauf | 1 | 1 |
| R31b | wie R31, nur `TestZweiServerTeilenNichts` | 3 | 3 |
| R32 | `ErrorLog` nicht gesetzt, nur `TestKeinLogDerBibliothek` | 1 | 1 |
| R41 | `durchAnhalten` im `Accept` ohne Sperre gelesen | 6 | 6 |
| R43 | `merkeZustand` ohne Sperre | 6 | 6 |
| R44 | `geschlossen` nach dem Entsperren geschrieben | 6 | 6 |
| Q1 bis Q10 | A2 `Connection: close`; S4 Body vor dem 503; S9 `BaseContext`; S12 Flag ohne Schließen; S10 zweites `Schliessen` wartet; S5 `Anhalten` bricht `ctx` ab; S7 `frist <= 0`; S13 Sperre über Handler; S11 `Anhalten` schließt Leerlauf; S8 nil nach Ablauf | je 1 | je 1 (Q4 und Q8 im Gesamtlauf mit Zeitlimit) |
| P1 bis P12 | S1; S2; S3; S6 Fall 0; S9; G1 `ReadTimeout`; G3; G2; S13 `Vergleich`; X2 (P10); S12 und S13 Handler mit dem benannten Test allein (P11 1 von 1, P12 3 von 3) | je 1 | je alle |

Von 52 Läufen sind 39 in jedem Prozess rot, 3 nur teilweise (R24 ohne Detektor, R27, R27b), 10 grün: R1, R3, R4, R5, R6, R9, R10, R12, R25 (N-3, erwartet) und R30b.

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 6 |
| LOW | 3 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Zusage ohne Fänger: Zustand der Verbindung nach der Antwort (F-1) · Prüfgrenze der Frist falsch beschrieben (F-2) · Allaussage über Zustandsfelder an einem Element geprüft (F-3) · Nebenläufigkeits-Mutation nur mit Wahrscheinlichkeit gefangen, Rate nicht benannt (F-4) · Art des Fehlers der Klasse ausgelassen (F-5) · Paar von Ursachen ohne Fänger (F-6) · Kommentar nennt eine Wirkung, die kein Test zeigt (F-7, F-8)

## Verdikt

**Merge-blockierend:** ja für F-1 bis F-6. Der Code selbst hat keinen Befund (Verhalten gegen `SPEC-136` Nr. 12 bis 14 und `SPEC-238` Nr. 9 bis 12 gelesen; von 52 Mutationsläufen 39 in jedem Prozess rot, 3 teilweise, 10 grün, davon eine äquivalent und eine als N-3 benannt). Die sechs MEDIUM sind Zusagen ohne Fänger oder mit falsch benannter Prüfgrenze und lassen sich mit je einem Fall oder einem Negativ mit Grund und Folge schließen. Die Findings gehen an den Implementer; F-11 an den Architect, die Spalte *Mutation* von X2 an den Planner (beide stehen in §6 mit Ausgang *offen*). Der Report ersetzt keine Verifikation.
