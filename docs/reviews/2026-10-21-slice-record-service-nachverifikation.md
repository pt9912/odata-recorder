# Nachverifikation: slice-record-service

* Datum: 2026-10-21
* Rolle: verifier (frischer Kontext; kein Code, keine Spezifikation, kein Plan, kein Lastenheft, keine `AGENTS.md` geändert)
* Gegenstand: `edbf37d` (Architect) und `122f086` (Implementer) gegen die Befunde V1 bis V11 der Verifikation `docs/reviews/2026-10-21-slice-record-service-verifikation.md` (`f3b4860`); Produktcode seit `f3b4860` unverändert (`git diff --stat f3b4860..HEAD -- src/` nennt nur fünf Testdateien)
* Plan: `docs/plan/planning/done/slice-record-service.md`
* Maßstab: `SPEC-121` Kopf, Nr. 2e, 9, 10, 13, 15 und *Ungeprüft*; `LH-FA-02.b` (Abschnitt der Spezifikation); Lastenheft [LH-FA-01](../../spec/lastenheft.md), [LH-FA-02](../../spec/lastenheft.md), [LH-FA-13](../../spec/lastenheft.md), [LH-QA-02](../../spec/lastenheft.md) (nur gelesen); §1 und DoD der Nehmer `slice-session-wartung` und `slice-http-adapter`; `AGENTS.md` §3.9, §3.11, §3.13 bis §3.16; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)
* Belege selbst gefahren: 20 Mutationen auf Kopien von `src/` (Abschnitt 1), sieben Läufe zu den Raten der Fänger mit Wahrscheinlichkeit (Abschnitt 2), fünf Zeitläufe (Abschnitt 3), die Stufen `lint` und `test` des Dockerfiles ungecacht (Abschnitt 3), `make gegenproben` (Abschnitt 6), `git diff --numstat` (Abschnitt 7), `make gates` auf `122f086` (Abschnitt 8). Jedes `go test` lief im Image der Stufe `deps` des Dockerfiles (`golang@sha256:8a5910…`, Go 1.27.1, aus dem Cache neu gebaut, gleiche ID), `--network none`, die Quelle schreibgeschützt (`git archive HEAD` im Scratchpad), Mutanten in Kopien. Der Arbeitsbaum blieb bis auf diesen Bericht unberührt (`git status` leer nach jedem Lauf).

## Urteil

**Freigegeben: kein HIGH, kein MEDIUM.** Die vier MEDIUM V1 bis V4 sind behoben: Die sieben früheren Mutanten sind jetzt aus dem benannten Grund rot (M12 wie benannt nur mit Wahrscheinlichkeit), dazu sieben eigene Mutationen zu V1 bis V4. V5, V7, V8, V9 und V10 sind erledigt. `TestAnnahmeGegenDrainNachLesen` prüft, was B8 sagt. Seine beiden Kopplungen prüft er selbst: sechs Kopplungs-Mutationen machen ihn rot, keine lässt ihn still grün. Kein Zeitlauf wurde ohne Codefehler rot, `make lint` meldet 0 issues.

Neu ist ein LOW: V12 (Planner) – der Plan sagt „hoher Wahrscheinlichkeit“ für die Lücke in `nimmAn`, deren Mutation nur in 10 bis 19 von 100 Läufen rot wird. Neu ist auch ein INFO: V13 (Architect) – das erneute Lesen der Zahl nach einem gescheiterten Lesen folgt aus `SPEC-121` Nr. 10, steht dort aber nicht wörtlich. Beim Planner offen bleiben V6 (I-4 ohne Spur) und V11 (Größe 4325 gegen die Marke 3600). DoD-Häkchen und §7 trägt der Planner.

## 1. Mutationen (V1 bis V4, V5, V7)

**Verfahren** wie in der Verifikation: Ein Go-Programm kopiert die Quelle je Mutante und ersetzt jede Stelle genau einmal; trifft eine Ersetzung nicht genau eine Stelle, bricht es ab. Dann läuft `go test -count=1 -timeout=600s` über das Paket `record`, bei den Kopplungen nur `TestAnnahmeGegenDrainNachLesen`. Alle 20 bauten, keine hing. „An der Zeitgrenze“ heißt: die 10 s des Tests, nicht der Timeout von `go test`. Gezählt wird nur Rot aus dem benannten Grund. M01, M02, M10, M12, M20, M22 und M23 sind wörtlich die Ersetzungen aus `f3b4860`.

| # | Zeile in §3 | Mutation | Ergebnis |
|---|---|---|---|
| M01 | B10 (V1) | `*` löst das Entfallen der Glieder `identity` nicht aus | rot: `TestAcceptEncoding`, `*, identity;q=0` → `identity;q=0` |
| M02 | B10 | Entfallen je Header-Wert statt über alle Werte | rot: beide Fälle über zwei Werte (Zeilen `br` und `identity;q=0`; Schreibweisen `Accept-Encoding` [`identity;q=0`] und `accept-encoding` [`x-foo`]) |
| M23 | B10 | nur `br`, `zstd`, `compress` lösen aus | rot: `x-gzip, gzip, identity;q=0`, `*, …`, `x-foo, …`, `sdch, IDENTITY`, die zwei Schreibweisen |
| N01 | B10 | `x-gzip` entfällt ohne auszulösen (wie ein leeres Coding) | rot: `x-gzip, gzip, identity;q=0` → `gzip, identity;q=0` |
| N02 | B10 | nur Glieder des ersten Werts lösen aus | rot: nur der Fall der zwei Schreibweisen (der Auslöser steht im zweiten Wert) |
| N03 | B10 | Codings mit dem Präfix `x-` lösen nicht aus | rot: `x-gzip, …`, `x-foo, …`, die zwei Schreibweisen |
| M10 | B28 (V2) | ein Zähler für den Dienst statt je Session | rot: `TestDrainBericht`, „Bericht s1 trägt {Abgebrochen:1 NichtAngehaengt:1} aus s2“, ebenso `unbekannt` |
| N05 | B28 | nur `NichtAngehaengt` ein Zähler für den Dienst | rot: „Bericht s1 trägt {Abgebrochen:0 NichtAngehaengt:1} aus s2“, ebenso `unbekannt` |
| N06 | B28 | `Drain` liefert die Summe der Berichte aller Sessions | rot: `s1` und `unbekannt` je {1 1} |
| M22 | B20 (V3) | Zahl vor der Fehlerprüfung von `ListBySession` gesetzt (0, bekannt) | rot: `TestNichtAngehaengt`, Fall `ListBySession`: „danach 0 angehängt, erwartet einer mit der Nummer 4“ (der nächste trug 1 und wurde abgewiesen), `ListBySession` 1-mal statt 2; Fall ohne vorhandene: 1-mal statt 2 |
| N07 | B20 | nach jedem Fehler von `Append` wird die Zahl neu gelesen | rot: sieben Fälle „`ListBySession` 2-mal, erwartet 1“ |
| M20 | B19 (V4) | eine Sperre der Anhänge für alle Sessions | rot an der Zeitgrenze: `TestAnhaengeUnabhaengig`, „Exchange von s2, während Append von s1 hält, kam nicht innerhalb von 10s“ |
| N09 | B19 | erstes Lesen der Zahl samt `Append` unter einer Sperre des Dienstes | rot an der Zeitgrenze: derselbe Grund |
| M12 | B8, Zusatzzeile (V5) | `Drain` zählt die laufenden und setzt den Drain in zwei Sperrabschnitten | rot im ganzen Paket (ein Lauf): nur `TestAnnahmeGegenDrainGleichzeitig`, Runde 101; Raten in Abschnitt 2 |

**Kopplungen von `TestAnnahmeGegenDrainNachLesen` (V7).** K1, K2, K3 und K5 sind die Kopplungen, die B8 nennt; K4 und K6 sind eigene.

| # | Mutation | Ergebnis |
|---|---|---|
| K1 | Ableitung des Kontexts vor dem Lesen der Session | rot: „im Halt 0-mal gelesen, erwartet 1“ |
| K2 | Ableitung nach der Annahme | rot an der Zeitgrenze: „Drain im Halt endet nicht innerhalb von 10s“ |
| K3 | Ableitung ruft `Done` des `ctx` nicht (`context.WithCancelCause(context.WithoutCancel(ctx))`) | rot an der Zeitgrenze: „erstes Done kam nicht“ |
| K4 | `Record` prüft vor der Ableitung mit `select` auf `ctx.Done()` | rot: „Halt nicht in context.WithCancelCause, gerufen aus Record“ |
| K5 | Session-Fake ruft in `List` `ctx.Done()` | rot: derselbe Grund (Halt in `List`) |
| K6 | Ableitung in eigenem Abschnitt unter `r.mu` | rot an der Zeitgrenze: `Drain` wartet im Halt an der Sperre |

Jede Änderung einer der beiden Kopplungen macht den Test also rot. Laufen Reihenfolge und Halt auseinander, meldet er die Lage des Halts. Fehlt der Halt oder hängt `Drain` dahinter fest, endet er an der Zeitgrenze. Nie bleibt er still grün. K4 und K6 sind Änderungen, die Nr. 2e nicht brechen; der Test meldet sie trotzdem laut. Das ist der Preis der Kopplung, und B8 benennt ihn.

## 2. `TestAnnahmeGegenDrainNachLesen` gegen B8, Raten der Fänger mit Wahrscheinlichkeit

**Was B8 sagt, prüft er.** Der Halt liegt nach der Rückkehr des Lesens der Session. Das belegen zwei Prüfungen: Der Stapel zeigt `context.WithCancelCause` direkt aus `(*Recorder).Record`, also ist `pruefe` zurückgekehrt, und der Session-Fake wurde einmal gelesen. Der Halt liegt vor der Annahme, denn `Drain(s1, 1 h)` kehrt im Halt mit 0 und 0 zurück. Nach der Freigabe kommen `ODR-E5009` mit 409, kein `Senden`, keine bezogene Kennung und kein angehängter Exchange. `started_at` prüft er nicht; das tut `TestAnnahmeGegenDrain`.

**Die Kopplung an die Standardbibliothek** steht im Image: `(*cancelCtx).propagateCancel` in Go 1.27.1 ruft `parent.Done()` als erste Anweisung synchron (`context.go`, Zeile 479). Es ruft es noch einmal in `parentCancelCtx` und in der Goroutine für fremde Eltern-Kontexte. `sync.Once` lässt nur den ersten Aufruf anhalten, und die Goroutine endet mit `abbrechen(errEnde)`. Unter `-trimpath` (Stufe `test`) bleiben die Funktionsnamen im Stapel gleich. Das zeigt `pruefung test: ok`.

**`containedctx`.** `haltCtx` trägt kein Feld vom Typ `context.Context`. Die Stufe `lint` lief ungecacht und meldete „0 issues“ für `src`, `tools/gegenproben` und `tools/pruefung` (Abschnitt 3). `.golangci.yml` nimmt `containedctx` für Testdateien nicht aus.

**Raten** (Läufe von `go test -count=N -run` über eine Kopie, gegen die Zahlen in §3 und §6 des Plans):

| Test gegen Mutation | Plan | dieser Lauf |
|---|---|---|
| `TestAnnahmeGegenDrainNachLesen` auf dem Code | grün in 50 von 50 | grün in 50 von 50; dazu 1000 von 1000 (Abschnitt 3) |
| … gegen F-1 (Prüfung in eigener Sperre nach dem Lesen, Annahme ohne Prüfung) | rot in 50 von 50 | rot in 50 von 50 („nach der Rückkehr von Drain angenommen“) |
| … gegen die Lücke in `nimmAn` (Sperre kurz freigegeben) | grün in 50 von 50 (blind) | grün in 50 von 50 |
| … gegen M12 (Seite von `Drain`) | grün in 50 von 50 (blind) | grün in 50 von 50 |
| `TestAnnahmeGegenDrainGleichzeitig` gegen F-1 | 50 von 50; 99 von 100 | 100 von 100 |
| … gegen die Lücke in `nimmAn` | 13 von 100; 10 von 50; 19 von 100 | 10 von 100 |
| … gegen M12 | 12 von 20; 99 von 100; 99 von 100 | 96 von 100 |

Die Zahlen des Plans stimmen in der Größenordnung mit meinen überein. Die Lücke in `nimmAn` wird je Lauf nur selten rot (V12).

## 3. Zeitläufe und ungecachte Stufen

**Rot ohne Codefehler: keines.**

| Lauf | Ergebnis |
|---|---|
| Paket `record`, `-count=50` | grün, 4,4 s |
| Testbinary, `-test.count=50`, `--cpus=0.05`, Arbeitsverzeichnis im Paket | grün, 72 s |
| `TestAnnahmeGegenDrainNachLesen`, `TestAnhaengeUnabhaengig`, `TestDrainBericht`, `TestNichtAngehaengt`, `TestAcceptEncoding`, `TestAnnahmeGegenDrainGleichzeitig`, `-test.count=200`, `--cpus=0.05` | grün, 3 min 45 s |
| dieselben, `GOMAXPROCS=1`, `-test.count=200` | grün |
| `TestAnnahmeGegenDrainNachLesen`, `-test.count=1000` | grün |
| Stufe `lint`, `--no-cache-filter lint` | „0 issues.“ dreimal, 19 s |
| Stufe `test`, `--no-cache-filter test` | `pruefung test: ok` (gofmt, vet, Tests mit `-trimpath`), 49 s |

Ein erster Lauf des Testbinarys mit `--cpus=0.05` startete aus `/out` statt aus dem Paketverzeichnis. Er wurde nur in `TestImporte` rot, an `open ../../ports/inbound`. Der Fehler lag in meinem Aufbau, nicht im Code; der Lauf aus dem Paketverzeichnis ist der in der Tabelle.

## 4. `SPEC-121` nach `edbf37d`

| Stelle | Wortlaut | gegen |
|---|---|---|
| Kopf (Zeile 1089) | „Der Adapter leiht dem Dienst die `Anfrage` … ; ihren Strom besitzt und schließt er (`LH-FA-02.b`)“ | `LH-FA-02.b` (Zeile 347): „besitzt und schließt … gibt sie nur als geliehene … weiter“; gleich (V10). `Anfrage.Body` ist ein `io.Reader`, der Dienst kann ihn nicht schließen. |
| Nr. 2e (Zeile 1097) | „unter ihr beginnt auch `Drain` (Nr. 13): er setzt die Session in den Drain und sieht ihre laufenden Exchanges, ohne die Sperre dazwischen freizugeben“ | `drain.go` `sperre`: `drain = true` und `len(sz.laufend)` unter einem `r.mu`; Nr. 13 Punkt 1 verweist jetzt auf 2e |
| *Ungeprüft* (Zeile 1124) | Sperre „auf beiden Seiten“, in `Record` „soweit beide nach der Ableitung … liegen“, in `Drain` zwischen Setzen und Blick; Grund (kein Aufruf, kein Test-Haken); Halt am ersten `Done` für eine Lücke, die die Ableitung einschließt; Folge: angenommen nach dem Beginn von `Drain`, läuft nach dessen Rückkehr, fehlt im gelieferten Bericht, Session womöglich `complete` | Nr. 2e, 9, 13: kein Widerspruch. Beide Lücken (M12, Lücke in `nimmAn`) führen genau zu dieser Folge: angenommen, nachdem `Drain` `0` laufende sah. „Nur mit Wahrscheinlichkeit“ ohne „hoch“ stimmt mit Abschnitt 2. |
| Plan §3, akzeptierte Negative (Zeile 149) | dieselben zwei Seiten, Grund, deterministischer Teil, Folge | gleich *Ungeprüft*; Grund und Folge stehen da. „hoher Wahrscheinlichkeit“ siehe V12 |
| Plan B8, Zusatzzeilen (Zeilen 121, 225, 226) | deterministischer Teil bis zur Ableitung, Rest mit Wahrscheinlichkeit, akzeptiertes Negativ | gleich *Ungeprüft* |
| `slice-session-wartung` §1 | „`Drain` … kehrt erst zurück, wenn kein Exchange der Session mehr läuft und keiner mehr beginnt“; `CloseAll` ruft `Drain` für mehrere Sessions zugleich | Nr. 13; das Restfenster, in dem das nicht gilt, steht in *Ungeprüft*. Den Bericht je Session fängt jetzt `TestDrainBericht` (M10, N05, N06). |
| `slice-http-adapter` §1 | „Vom Aufruf von `Record` bis zur Rückkehr von `Close` … ändert er weder `Anfrage.Header` … und liest `Body` nicht (`SPEC-121` Kopf)“ | Kopf; ohne „gehört“, keine Änderung nötig |

„gehört“ neben `Anfrage` steht in `spec/` nur noch in der Historie-Zeile des Reviews, wie der Architect angibt. In den Plänen unter `open/` und `in-progress/` steht es nur im Zitat seiner Suche. Lastenheft und `AGENTS.md` sind seit `7637843` unverändert (`git diff --stat` leer). Die Historie trägt die Zeile vom 2026-10-21 nach der Verifikation.

## 5. Plan gegen Code (`AGENTS.md` §3.9), Kommentare (§3.11)

**`edbf37d`** zieht im Plan den Kopf (Berührte Spec-Stellen), §3 (Test-Haken, Abschnitt Race-Detektor, B8, akzeptierte Negative) und §6 (Zusatzzeilen, Risiko Nebenläufigkeit, Ausgang V5, V7, V10) im selben Commit mit der Spezifikation nach.

**`122f086`** zieht §3 (Testdateien, Grenze der Blase, B8, B10, B18, B19, B20, B26, B28), §4 (Messung) und §6 (drei Zusatzzeilen, Risiken) nach. Je Zeile gegen den Code und meine Läufe:

| Zeile | Plan | Befund |
|---|---|---|
| B8 | NachLesen, Lage des Halts, vier Kopplungs-Mutationen samt Grund | stimmt Wort für Wort mit K1, K2, K3, K5 |
| B10 | neun neue Fälle, vier Mutationen | alle neun im Test; M01, M23 und M02 wie angegeben (M23 macht zusätzlich den Fall der zwei Schreibweisen rot); die `q=0`-Mutation nicht nachgefahren |
| B18 | „zwei Sessions abwechselnd“ | `TestSequenz`: `s1, s2, s1, s2, s2` (V9 erledigt) |
| B19 | `TestAnhaengeUnabhaengig`, zwei Mutationen an der Zeitgrenze | M20 und N09 wie angegeben |
| B20 | drei vorhandene, Fall ohne; Neulesen nach `ErrUngueltigeSequenz` und nach Fehler von `ListBySession`, sonst einmal | wie im Test; M22, N07 rot. Siehe V13. |
| B26, B28 | unbekannte Kennung und Bericht je Session | Block `j` in `TestDrainBericht`; M10, N05, N06 rot |
| Zusatzzeilen (Zeilen 229, 238, 250) | `Accept-Encoding` über alle Werte; Bericht über eine andere Session; zwei Sessions zugleich mit B19 `TestAnhaengeUnabhaengig` und B27 | getragen (V4 Planner-Teil erledigt) |
| Zeile 105 | Grenze der Blase an einem `sync.Mutex`; NachLesen und `TestAnhaengeUnabhaengig` außerhalb der Blase | V8 erledigt; K6 und M20 bestätigen das Warten an einer Sperre, rot an der Zeitgrenze |
| Zeilen 98, 280 | Body des Fakes als Folge von Schritten (`schritt`, `dann`); kleine Header-Namen benannt | V9 erledigt |

**Kommentare.** Alle stimmen mit dem, was die Tests fangen:

* `haltCtx` und `gerufenAus` beschreiben, was sie tun. Das gilt für den ersten passenden Rahmen.
* Der Kopf von `TestAnnahmeGegenDrainNachLesen` nennt beide Kopplungen, ihre Prüfung, das Rot an der Zeitgrenze bei K2 und K3 und die Blindheit nach der Ableitung. Das bestätigen Abschnitt 1 und 2.
* Der Kopf von `TestAnnahmeGegenDrainGleichzeitig` sagt „mit Wahrscheinlichkeit“ ohne „hoch“.
* Die Köpfe von `TestAnhaengeUnabhaengig` und `TestNichtAngehaengt`, der Kommentar an `haltAppend` und der neue Block in `TestDrainBericht` stimmen.
* Zwei Kommentare sagten in `f3b4860` mehr zu, als ein Test fing. Beide sind jetzt gedeckt: `acceptEncoding` („ein Glied mit nichtleerem Coding“, M01, M23, N01 bis N03) und `DrainBericht` („zählt je Session“, M10, N05, N06).

## 6. Gegenproben-Stempel

`make gegenproben` auf `122f086`: „Eingaben unverändert seit dem letzten grünen Lauf (`61b3d53e…`) — nichts läuft neu“. Der Stempel ist der von `122f086`, er änderte sich nicht, und `git status` blieb leer. Er ist aktuell. [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) nimmt die Gegenproben aus der Gate-Kette.

## 7. Größe (nur Bericht)

`git diff --numstat 7637843..HEAD` über `src/hexagon/ports/outbound/`, `src/hexagon/ports/inbound/` und `src/hexagon/services/record/`, Tests eingeschlossen: **4325**. Das sind 4325 Einfügungen und 0 Löschungen, davon 997 Produktcode und 3328 Tests. `122f086` allein unter `src/` ändert 221 Zeilen (199 Einfügungen, 22 Löschungen, nur Tests). Die Spezifikation hat seit der Basis kumuliert 26 Zeilen (15 und 11). Je Commit sind es 8 (`647f988`), 17 (`567abe7`) und 9 (`edbf37d`); die Summe 34 liegt über 26, weil Zeilen in mehreren Commits geändert wurden. Die Zahlen des Implementers stimmen. Die Marke 3600 ist um 725 (20,1 %) überschritten; der Wert liegt in der Spanne 4250 bis 4800, die §6 erwartet. Die Entscheidung trifft der Planner (V11).

## 8. Gate-Lauf

`make gates` auf `122f086` vor diesem Bericht: rc 0. Die Kette lief über docs-check, a-check, baseline-verify, schema-check, migration-check, lint-profil, lint, build und test. `lint` und `test` kamen dort aus dem Cache, ihr ungecachter Lauf steht in Abschnitt 3. `make gates` mit diesem Bericht im Baum: Ergebnis im Handback an den Planner.

## Stand der Befunde V1 bis V11

| ID | Stand | Beleg |
|---|---|---|
| V1 MEDIUM | erledigt | M01, M02, M23, N01, N02, N03 rot; Kommentar von `acceptEncoding` gedeckt |
| V2 MEDIUM | erledigt | M10, N05, N06 rot (`TestDrainBericht`, Block je Session); Kommentar von `DrainBericht` gedeckt |
| V3 MEDIUM | erledigt | M22, N07 rot (drei vorhandene Exchanges); Wortlaut siehe V13 |
| V4 MEDIUM | erledigt | M20, N09 rot an der Zeitgrenze (`TestAnhaengeUnabhaengig`); Zusatzzeile nennt B19 und B27 |
| V5 LOW | erledigt | `SPEC-121` Nr. 2e, *Ungeprüft*, Plan §3 mit Grund und Folge; M12 rot in 96 von 100; Wortwahl siehe V12 |
| V6 LOW | offen | Planner: I-4 hat weiter keine Spur im Plan (`grep -c I-4` 0) |
| V7 INFO | erledigt | übernommen; K1 bis K6 rot (Abschnitt 1) |
| V8 INFO | erledigt | Plan Zeile 105 |
| V9 INFO | erledigt | Abschnitt 5 |
| V10 INFO | erledigt | Abschnitt 4 |
| V11 INFO | offen | Planner: 4325 gegen 3600 (Abschnitt 7) |

## Befunde dieser Nachverifikation

| ID | Kategorie | Befund | Quelle | Pfad | Zuständig |
|---|---|---|---|---|---|
| V12 | LOW | Drei Stellen des Plans sagen „hoher Wahrscheinlichkeit“ für die Lücke nach der Ableitung in `nimmAn`. Zeile 105 sagt „Zwei Fänger fangen ihre Mutation nur mit hoher Wahrscheinlichkeit … `TestAnnahmeGegenDrainGleichzeitig` in B8 (… auf beiden Seiten: in `Record` zwischen Prüfung und Annahme nach der Ableitung …)“. Zeile 149 (akzeptierte Negative) sagt „fängt ein Freigeben der Sperre dort nur mit hoher Wahrscheinlichkeit“. Zeile 282 (Ausgang F-1) sagt „danach mit hoher Wahrscheinlichkeit“. Gemessen sind je `go test`-Lauf 13 von 100, 10 von 50, 19 von 100 (Plan B8) und 10 von 100 (Abschnitt 2). Die Stufe `test` fährt jeden Test einmal, also bleibt die Mutation in mehr als vier von fünf Gate-Läufen grün. „Hoch“ trifft nur die Seite von `Drain` (96 bis 99 von 100) und den Ablauf mit `-count=50` vor der Übergabe. Vor `edbf37d` schloss das Fenster F-1 ein (rot in 50 von 50); die Einengung auf den Teil nach der Ableitung ließ das Adjektiv stehen. `SPEC-121` *Ungeprüft*, B8 und der Kopf des Tests sagen richtig „mit Wahrscheinlichkeit“ und nennen die Zahlen. Vorschlag: „hoher“ streichen oder die Raten je Lauf je Seite nennen. | `AGENTS.md` §3.11, §3.15; `SPEC-121` *Ungeprüft* | `docs/plan/planning/done/slice-record-service.md` · `fängt ein Freigeben der Sperre dort nur mit hoher Wahrscheinlichkeit` | Planner (Wortlaut im Plan; stammt aus `edbf37d`) |
| V13 | INFO | `SPEC-121` Nr. 10 nennt das Lesen der Zahl „beim ersten Anhängen einer Session in diesem Dienst und nach `ErrUngueltigeSequenz`“. Dass nach einem gescheiterten Lesen das nächste Anhängen wieder liest, folgt aus „die Zahl bleibt“ (ungelesen) und „Der Exchange erhält die Zahl plus eins“, steht aber nicht wörtlich da. Wörtlich ist das zweite Anhängen nicht mehr das erste, und Nr. 9 zählt ein gescheitertes Anhängen als Anhängen. `TestNichtAngehaengt` legt die Lesart jetzt fest (`ListBySession` zweimal), ebenso die Fänger-Spalte von B20; die Zusage-Spalte von B20 nennt nur `ErrUngueltigeSequenz`. Code (`!sz.zahlBekannt`) und Test sind richtig. Vorschlag: „solange die Zahl nicht gelesen ist und nach `ErrUngueltigeSequenz`“. | `SPEC-121` Nr. 9, 10; `AGENTS.md` §3.12, §3.16 | `spec/spezifikation.md` · `beim ersten Anhängen einer Session in diesem Dienst und nach` | Architect |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Produktcode | seit `f3b4860` unverändert |
| Frühere Mutanten M01, M02, M10, M12, M20, M22, M23 | alle rot aus dem benannten Grund (M12 mit Wahrscheinlichkeit, wie benannt) |
| Eigene Mutationen zu V1 bis V4 | sieben, alle rot |
| `TestAnnahmeGegenDrainNachLesen` | prüft B8; beide Kopplungen selbst geprüft (K1 bis K6 rot); `containedctx` und `make lint` ohne Befund |
| Zeitabhängigkeit | kein Rot ohne Codefehler in fünf Läufen bis `--cpus=0.05` und `GOMAXPROCS=1` |
| `SPEC-121` Kopf, Nr. 2e, 13, *Ungeprüft* | widerspruchsfrei untereinander, zu `LH-FA-02.b`, zum Plan und zu den Nehmern; akzeptierte Negative mit Grund und Folge |
| Übergaben (§3.13) | `slice-session-wartung` und `slice-http-adapter` brauchen keine neue Sendung; keiner liegt in `done/` |
| Gegenproben, Gates | Stempel aktuell; `make gates` rc 0 |

## Zuständigkeiten

* **Planner:** V12 (drei Stellen im Plan), V6, V11; DoD-Häkchen, Ausgänge in §6, §7. Kandidaten für das Register: V12 als Beleg zu `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (die Einengung des Fensters erreichte das Adjektiv nicht), V1 bis V4 der Verifikation zu `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, V5 zu `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen`.
* **Architect:** V13 (Wortlaut von Nr. 10, ohne Folge für Code oder Test).
* **Implementer:** nichts.
