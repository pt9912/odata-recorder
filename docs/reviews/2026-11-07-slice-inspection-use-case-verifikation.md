# Verifikation slice-inspection-use-case (Teil 1: Exchanges)

**Rolle:** Verifier (`Bauen wir es richtig?`, gegen Plan, DoD und Spezifikation). **Datum:** 2026-11-07.
**Gegenstand:** Zweig `slice-inspection-use-case`, HEAD `9e6fea7` (Fix-Runde nach dem Review `ed2f0eb`), Basis `81e960d`.
**Eingang:** DoD-Bestätigung und Bericht des Implementers in §3 des Plans, Review `docs/reviews/2026-11-07-slice-inspection-use-case-review.md`.
**Bezug:** `LH-FA-13`, `LH-QA-04`, `LH-QA-05`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).
**Spezifikation:** `SPEC-236` Nr. 1 bis 4 und *Ungeprüft*, `SPEC-109` Nr. 5, `SPEC-124` Nr. 3, `SPEC-105` Nr. 8.

## Urteil

**Bestanden mit Auflagen.** Kein HIGH, kein Befund am Code gegen den Vertrag. Zwei LOW im Plan und im Test (V-1, V-2), eine Auflage an den Planner (V-3) und die Closure-Reste (V-4). Die Fixes aus dem Review (F-1 bis F-6) decken die Klasse, nicht nur die genannten Mutationen (Abschnitt *Klassen*).

## Eigene Sensoren (nicht die Behauptung, die Belege)

| Sensor | Ergebnis |
|---|---|
| `make gates` (einmal, eigenes Log `scratchpad/vi-iuc/gates.log`, unter `flock /tmp/odata-recorder-heavy.lock`) | Exit 0; `d-check` 0 Befunde, `groessenmarke slice-inspection-use-case.md: ok, Messung 1404 von Marke 2500`, `zeichen-check` ok |
| Abdeckung selbst gemessen: `go test -race -covermode=atomic -coverprofile` in einem Container aus `odata-recorder:race` (Baum per `docker cp`, `--network none --memory=4g --cpus=4`, Image nicht überschrieben) | Paket `inspection` 100,0 % der Anweisungen, je Funktion 100,0 % (`mitCode`, `New`, `ctxNil`, `Exchanges`, `wirksamesLimit`, `schneide`, `zeile`); Paket `inbound` 100,0 %; deckt sich mit dem Plan |
| Reihenfolge-Unabhängigkeit: `go test -race -count=3 -shuffle=on` über `inspection` und `inbound` | grün |
| Rohbytes (`AGENTS.md` §3.18): `LC_ALL=C grep -cP` auf Backslash-u-Folge, BOM, NBSP, Zeilentrenner, Nullbreiten- und Steuerzeichen über alle 12 geänderten oder neuen Dateien (`inspection.go`, `inspektion.go`, `fehler.go`, die sechs `*_test.go`, Plan, Review, `slice-inspection-coverage-diagnose.md`) | 0 Treffer in jeder Datei |
| Diff-Umfang `81e960d..HEAD` | nur `ports/inbound/inspection.go`, `services/inspection/` (Dienst, `fehler.go`, 6 Testdateien), der Plan, das Review und eine Zeile in `slice-inspection-coverage-diagnose.md` (Hinweis zur Kennung `SPEC-236`); `ports/outbound/`, SQLite-Adapter, Telemetry, `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/`, `Dockerfile` unberührt |
| `//nolint` | keine im Paket und im Port |
| Arbeitsbaum nach den Läufen | sauber |

## Plan gegen Code

- Port: `Seite{Limit, Nach}`, `ExchangeSeite{Eintraege, Limit, Gesamt, Weiter}`, `ExchangeZeile` (15 Felder, Namen und Typen wie in `SPEC-236`), `InspectionUseCase.Exchanges`; `Coverage` und `Diagnose` fehlen bewusst (Teil 2), das steht im Plan §1, §4 und §6 mit Adresse.
- Dienst: `Konfig{Exchanges}`, `New` (nil: `model.ErrUngueltig`), Prüffolge `ctx` nil, `Nach`, `Limit` vor dem Port (Nr. 2, 3), genau ein Aufruf von `ListBySession` mit der Kennung unverändert, Fehler über `mitCode(err, GegenstandSession)` mit `Stati()[0]` und `Ursache`, Kontextfehler unverändert, Schnitt im Speicher mit `Weiter` aus der `sequence_no` des letzten Eintrags, `Eintraege` nie `nil`. Keine Abweichung vom Vertrag gefunden.
- Importe: Domain und Ports (`a-check` grün, `services` darf nur Modell und Ports importieren).
- Sendung `SPEC-109` Nr. 5, Absatz *Anwendung* (Gegenstand `Session` bei `Exchanges`): der Zweig ändert `SPEC-109` nicht (richtig, `slice-mock-zuordnung` ändert dieselbe Stelle); die Sendung steht in §1 *Übernimmt*, in der DoD (Punkt 4) und in §6 mit Ausgang *offen* und Adresse Planner (Merge-Commit). Der Satz fehlt heute in `spezifikation.md` Z. 854 und ist dort zu ergänzen.

## Tabelle Zusage · Fänger gegen den Vertrag

Jeder Satz von `SPEC-236` Nr. 1 bis 4, soweit Teil 1, hat eine Zeile und einen Fänger, den ich mit einer Mutation aufgelöst habe:

| Satz des Vertrags | Zeile | Fänger | Eigene Mutation (rot) |
|---|---|---|---|
| Nr. 1 `New` weist nil in `Exchanges` ab | K1 | `TestNewExchangesNil` | `vm-new-ohne-konfig` (rot-paket: Panik) |
| Nr. 1 nur `ListBySession`, kein Schreiben | X13 | `TestLesenSchreibtNichts` | `vm-append-bei-portfehler` (Pfad Port-Fehler) |
| Nr. 2 `ctx` nil ohne Code, kein Port | A1 | `TestKontextNil` | durch die Mutationen des Implementers |
| Nr. 2 Kontext unverändert, nie vom Dienst geprüft | X8, X10 | `TestKontextfehlerUnveraendert`, `TestKontextDesAufrufs` | `vm-ctx-vor-port` (Dienst prüft `ctx.Err()` vor dem Port: `TestKontextDesAufrufs`) |
| Nr. 2 Argumentprüfung vor jedem Port | X1 | `TestSeitenparameterNegativ` | `vm-port-bei-limit-negativ` |
| Nr. 3 Standard, Obergrenze, Kappung, Antwort trägt `Limit` | X2 | `TestSeitenModell`, `TestGrenzenDesVertrags` (Literale) | siehe *Klassen*, Grenzwerte |
| Nr. 4 Gesamt, Eintraege, Weiter, leere Seite, Reihenfolge | X3, X4, X5 | `TestSeitenModell`, `TestSeitenFolge`, `TestSequenzMitLuecken` | siehe *Klassen*, Zeiger |
| Nr. 4 Stabilität beim Anhängen | X6 | `TestAnhaengenZwischenSeiten` | Cache der Liste an anderer Stelle (Abschnitt *Nebenläufigkeit*) |
| Nr. 4 Kennung unverändert, Fehler nach Nr. 8 | X7, X9 | `TestKennungUnveraendert`, `TestFehlerDesPortsMitCode` | `vm-blobfehlt-verschluckt`, `vm-status-stati-mitte` |
| Nr. 4/5 Allowlist, kein Geheimnis | X11, X12 | `TestAllowlistIstMengeDerFelder`, `TestFeldUnveraendert`, `TestKeinGeheimnis` | siehe *Klassen*, Allowlist und Canary |
| Nr. 4 *Gesamt zählt alle*, Session gelöscht zwischen zwei Seiten (`ODR-E7003`) | X3, X9 | wie oben (derselbe Weg über `ErrNichtGefunden`) | `vm-gesamt-minus-nach` |

Eine Zusage ohne Fänger habe ich im Code nicht gefunden. Gefunden sind zwei Zusagen des Plans, die mehr sagen, als der Fänger trägt (V-1, V-2).

## Klassen (`AGENTS.md` §3.17): eigene Stichproben an nicht genannten Elementen

Läufe mit `tools/harness/mutationslauf.sh`, Klasse `andere` ein Prozess je Mutation, unter `flock`, `MUTATIONSLAUF_SPEICHER=4g`; Messender: der Verifier; Last des Hosts 6,1 bis 10,5 bei 4 CPUs. Die Liste liegt im Scratchpad (`scratchpad/vi-iuc/liste.json`, `liste2.json`, `liste3.json`), die Namen mit Präfix `vm-`, `vn-`, `vp-` und `wx-` sind keine des Implementers und keine des Reviews. 33 Mutationen: 29 + 3 + 1.

| Klasse (Menge nach dem Wortlaut) | Elemente | Meine Mutationen an nicht genannten Elementen | Ergebnis |
|---|---|---|---|
| Grenzwerte `Limit` (Standard, Obergrenze, Gleichheit) | Konstanten 2, Zweige `0`, `> Max`, sonst, Antwort | `StandardLimit` 200; `MaxLimit` 1100, 500, 2000; Kappungswert `MaxLimit - 1`; Wert im Zweig *sonst* `limit - 1`; `limit <= 1` statt `== 0`; `Limit` der Antwort plus 1; `Weiter` bei `limit == MaxLimit` und `len == limit - 1` | 9 von 9 rot (`TestSeitenModell`, `TestGrenzenDesVertrags`) |
| Zeiger, `Weiter`, `Gesamt` (Seitenbildung) | Filter, Schnitt, `Weiter`, `Gesamt` | Filter über Index statt `sequence_no`; Schnitt bei `limit + 1`; `Weiter = nach + limit`; `rest` erst ab zwei Einträgen; `Gesamt - nach` | 5 von 5 rot (`TestSeitenModell`, `TestSequenzMitLuecken`) |
| Allowlist (15 Felder) | Feld × Wert | `KeyRedacted` negiert, `ResponseStatus + 1`, `SequenceNo + 1`, `FinishedAt.Add(1)` (Nullzeit wird nichtnull) | 4 von 4 rot (`TestFeldUnveraendert`) |
| Kein-Geheimnis (7 Felder × Einbringungsort × Kodierung) | Ziel der Einbringung | Leck in `Method` (`RawQuery`), in `ID` (`MatchKey`), Header-Namen der Anfrage in `Method`, Header-Werte der Antwort in `Method`: andere Ziele als `Path`, das der Implementer prüfte | 4 von 4 rot (`TestKeinGeheimnis`) |
| Port-Fehler (28 Fehler, 2 Stati) | Eintrag × Status | `ErrBlobFehlt` still verschluckt; Status des letzten statt des ersten Eintrags | 2 von 2 rot (`TestFehlerDesPortsMitCode`) |
| Kontext (vor/nach, Identität) | Stelle | Dienst prüft `ctx.Err()` vor dem Port (Port wird nicht gerufen, derselbe Fehler kehrt zurück) | rot (`TestKontextDesAufrufs`) |
| Lesen schreibt nichts, Argumentprüfung vor Port | Pfad × Operation | `Append` im Pfad Port-Fehler; `ListBySession` im Pfad `Limit < 0` | 2 von 2 rot |

Alle 28 Mutationen am Dienst rot (27 der Tabelle und `vm-new-ohne-konfig`); das ist mehr als der Befund nannte, die genannten Mutationen des Reviews (Literale 100 und 1000, Zweige von `TestGleichzeitig`, `Weiter` mit Lücken, Zeiten) sind in der Fix-Runde gedeckt (`g1` bis `g6`, `w-*`, `l1`, `z1` bis `z4` stehen in der Liste im Plan, 80 + 15 = 95 Zeilen, Zählung stimmt).

## Nebenläufigkeit

- `TestGleichzeitig`: 41 Zweige (8 feste, 5 Formen des Kontextfehlers, 28 Fehler mit Code) × 8 Goroutinen × 20 Runden. Eigene Mutationen an Stellen, die keine der 15 `w-*` ist (`s.k = s.k` vor dem Port-Aufruf, `s.k = s.k` zwischen Port-Aufruf und Fehlerprüfung): mit Detektor 20 von 20, ohne Detektor 0 von 50 Prozessläufen, Ergebnis `flatternd` (`-test.count=1` je Prozess, Last 6,1 bis 6,2 bei 4 CPUs). Das deckt sich mit dem Plan.
- Der Plan sagt das ehrlich: Datenrennen 20 von 20 nur mit Detektor, 0 von 50 ohne, `flatternd`; die Wirkung eines Zustands (Cache der Liste) fängt `TestAnhaengenZwischenSeiten`, nicht `TestGleichzeitig`. Meine Gegenprobe mit einem Cache an anderer Stelle als im Plan (Feld `gepuffert`, Ablage im Erfolgsfall, Lesen vor dem Port): `TestAnhaengenZwischenSeiten` rot in 5 von 5 Prozessen mit und 5 von 5 ohne Detektor; `TestGleichzeitig` mit Detektor rot (1 Lauf), ohne Detektor grün (1 Lauf). Damit stimmt die Aufteilung des Plans (Zustand mit Wirkung beim Anhängen-Test, Rennen beim Detektor).
- Es gibt kein Fenster zwischen Prüfen und Handeln (`SPEC-236` Nr. 10); eine Zeile je Fenster entfällt zu Recht.

## Zählprüfungen

- `TestZweigeSindDieWege` hält die Zweige von `TestGleichzeitig` gegen `wegeExchanges()` (13 Wege) und die Fehler mit Code gegen `outbound.EintraegeAbbildung()` plus 1, Kontextfehler 5. Die 13 Wege sind eine von Hand geführte Liste des Tests, keine Zählung gegen den Quelltext.
- `TestWegeImQuelltext` (W1, `go/parser`): zählt Aufrufe über Felder von `Konfig`, Aufrufe von `mitCode` und Stellen mit `ErrUngueltig`; liest `.` als Paketverzeichnis, das gilt unter `make test`, `make test-race`, im Mutationslauf und im Container. Lint und `a-check` grün. Benannte Negative im Plan stimmen: Port-Aufruf über lokale Variable (meine Mutation `vn-get-ueber-lokale-variable`: W1 bleibt grün, `TestLesenSchreibtNichts` ist rot, Ergebnis `rot-paket`); Abbilden ohne `mitCode`.

## Befunde

| # | Stufe | Quelle | Ort | Befund | Auflage |
|---|---|---|---|---|---|
| V-1 | LOW | `AGENTS.md` §3.11, §3.15 | Plan §3, Satz hinter der Tabelle *Wege*: „ein Weg ohne Test ist nach W1 rot“ | W1 zählt nur Port-Aufrufe, Aufrufe von `mitCode` und Stellen `ErrUngueltig`. Eine Verzweigung ohne diese Merkmale und ohne Test bleibt grün: `vn-neuer-zweig-ohne-test` (`if len(alle) > 100000 { alle = alle[:100000] }` vor `schneide`) ist `gruen`, alle Pakete und Tests grün. Die benannte Grenze von W1 im Plan nennt diese Klasse nicht. Der Dienst ist 100,0 % abgedeckt (gemessen), aber die Abdeckung ist kein Gate. | Satz im Plan berichtigen (W1 fängt Aufrufe und Argumentverstöße, nicht jede Verzweigung); die Klasse als benanntes Negativ mit Grund (Abdeckung wird einmal je Slice gemessen, kein Gate) und Folge (eine ungetestete Verzweigung bleibt grün) führen. Adresse: Implementer |
| V-2 | LOW | `AGENTS.md` §3.11 | `src/hexagon/services/inspection/inspektion.go` Kopf von `Inspektion`: „setzt die Operationen von `InspectionUseCase` um“ | Kein Test und keine Zusicherung zur Übersetzungszeit bindet `*Inspektion` an `inbound.InspectionUseCase`. `vp-signatur-weicht-vom-port-ab` (zusätzlicher variadischer Parameter, der Dienst erfüllt den Port nicht mehr) ist `gruen`. Folge: der Bruch fiele erst im Einbau in `slice-admin-sessions` oder `slice-bootstrap-m6` auf. | Eine Zeile `var _ inbound.InspectionUseCase = (*inspection.Inspektion)(nil)` in einer Testdatei des Pakets, oder ein benanntes Negativ im Plan (Folge: Bruch erst beim Einbau). Teil 2 erweitert den Port um zwei Methoden, die Zeile wächst dort mit. Adresse: Implementer |
| V-3 | Auflage an den Planner | `AGENTS.md` §3.7 (Rang-Zeiger) | `docs/plan/planning/in-progress/slice-inspection-coverage-diagnose.md` auf `main`, §6, Punkt „`SPEC-236` steht mit dem Merge von `slice-inspection-use-case` auf `main` für den ganzen Vertrag“ (Z. 205) | Der Satz zitiert „kein Gate prüft die Lücke zwischen Vertrag und Code, `SPEC-236` *Ungeprüft*“. `SPEC-236` *Ungeprüft* hat fünf Punkte (Größe der Session, Geheimnis im Pfad, Schlüsselmenge und Modi, Zählen der Fehler, HTTP-Antwort) und nennt diese Lücke nicht. In Teil 1 ist der Zitatort berichtigt (§6), in Teil 2 steht er unverändert (auch im Zweig). Der andere Verweis in Teil 2 (Zeile N2, Zählen der Fehler des Dienstes) ist richtig. | Planner berichtigt Z. 205 von Teil 2 auf main wie in Teil 1 (Lücke steht allein in §1 und §6) |
| V-4 | Closure-Rest | DoD des Plans | Plan §2 und §7 | Die DoD-Haken, §7 (Closure-Notiz mit Steering-Loop-Eintrag, Messung der Marke, Beobachtungs-Register, Ausgang jedes Risikos) fehlen noch. Die Messung 1404 von 2500 ist bei jeder Änderung nach dieser Verifikation neu zu lesen. Offen mit Adresse Implementer: der Bericht je Befund nach `AGENTS.md` §3.17 für V-1 und V-2 (Menge, Zahl der Elemente, eine Mutation an einem nicht genannten Element). | Bei der Closure |

## Antworten auf die Schwerpunkte der Anfrage

- **Plan gegen Code, Umfang:** deckt sich; nichts außerhalb von Port-Datei, Dienst, Tests, Plan, Reports.
- **Gates:** `make gates` grün, Messung 1404 von 2500, keine `//nolint`, `.golangci.yml` und `.a-check.yml` unverändert.
- **Fixes gegen die Klasse:** gedeckt, siehe *Klassen*; kein Element der Klassen Grenzwerte, Zeiger, Allowlist, Kein-Geheimnis, Port-Fehler und Kontext blieb in meinen 28 Stichproben grün.
- **`TestGleichzeitig`, ehrlich im Plan:** ja, Raten 20/20 und 0/50 und 0/30 stehen mit Last und Messendem im Plan und stimmen mit meinen Läufen überein.
- **Rohbytes:** 0 Treffer, siehe Tabelle; Werte jenseits von ASCII entstehen in den Tests aus Bytewerten.
- **Sendung `SPEC-109` Nr. 5:** in §1, DoD und §6 getragen.

## Eigene Läufe, Zeugen

Die Ablage der Listen und Logs liegt im Scratchpad unter `vi-iuc/` (`gates.log`, `liste.json`, `liste2.json`, `liste3.json`, `mut.out`, `cov.out`, `cache.sh`). Der Zweig blieb unverändert; Mutationen liefen in Kopien im Container. Nichts nach `main`, nichts gepusht.
