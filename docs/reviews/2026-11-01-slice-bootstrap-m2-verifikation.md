# Verifikationsbericht: slice-bootstrap-m2 — 2026-11-01

**Art:** Verifikation (Modul 11) — Konformität gegen die DoD des Plans, `SPEC-139`, `SPEC-138` Nr. 8, `SPEC-114` Nr. 2 und die Hard Rules `AGENTS.md` §3.10, §3.12, §3.15 bis §3.18. Nicht das Review (Diff gegen Plan) und nicht die Validierung.

**Gegenstand:** Zweig `slice-bootstrap-m2`, geprüfter Stand `1e03bfe081e1837fe7dc8c6ee06f946db5b22120` (Basis der Größenmarke `e8c027c`; letzter Code-Commit `0036af7`, danach nur Plan und Spec). Review-Report: `docs/reviews/2026-11-01-slice-bootstrap-m2-review.md` (R-1 bis R-9).

**Messweise.** Keine Behauptung des Implementers übernommen; alles hier ist selbst gefahren. Mutanten nur in Scratchpad-Kopien von `src/` (Worktree-Code unverändert), Image `odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`. Rot heißt: ein benannter Test fällt mit Testmeldung (kein Aufbaufehler; ein Aufbaufehler ist nicht gezählt und neu gefahren). 56 Mutanten je einmal über die Tests des Pakets `src/bootstrap`; Raten der nebenläufigen Zeilen: Testbinary einmal mit `go test -race -c`, **10 Prozesse** mit `-test.count=1` je Mutante (der Plan nennt 30; 10 genügen zur Gegenprobe, jede Abweichung wäre sichtbar). Durchstich am gebauten Binary aus `make gates` (Image `odata-recorder:build`, Binary enthält `errServeBeendet`, `dienst_abweichend` und `tls_pruefung_aus` des Standes) in `busybox:1.36` ohne Netz mit `httpd`-CGI als Upstream.

## Sensor-Belege (selbst gefahren)

| Sensor | Ergebnis |
|---|---|
| `make gates` (flock) | Exit 0; enthält `docs-check`, `schema-check`, `migration-check`, `build`, `test`, `test-race`, `groessenmarke`, `handbuch-check`, `zeichen-check`, `a-check`, `lint-profil`, `lint` |
| `make groessenmarke` | `Messpunkt slice-bootstrap-m2.md: zeilen=16 nebenlaeufig=3 fakes=3 produktcode=550 schaetzung=2445 marke=2900 messung=2759`; Hinweis „Tabelle ergibt 2445, Schätzung 2120“ (bekannt, Planner bei Closure) |
| Messung von Hand | `git diff --numstat e8c027c..0036af7 -- src/bootstrap/` Summe 2759 (2758 + 1); gleich bei `HEAD` (`git diff 0036af7..HEAD -- src` leer). Produktcode 642 Zeilen (316 + 188 + 135 + 3) gegen Spielraum 653 |
| `go test -race -count=20` über S1, S2, B5, B6, `TestServeAusfall`, `TestWarte`, `TestStartFehlerPaare`, `TestAufrufstellen`, `TestZugaengeGleichzeitig` | grün, 109 s (der Plan nennt `-count=50`, 298 s; 20 ist meine Stichprobe) |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -P` auf Backslash-u mit vier Hex-Ziffern, BOM, NBSP, Zeilentrenner, Nullbreite, Steuerzeichen über alle `src/bootstrap/*.go`, `main.go`, Plan und Review: 0 Treffer. Die Warnung `Zertifikatsprüfung zum Upstream abgeschaltet` ist in `start.go`, `start_test.go` und `spezifikation.md` bytegleich (`c3 bc`). `make zeichen-check` grün |

## Durchstich am gebauten Binary (unabhängig)

Alles in einem Container ohne Netz; Upstream `httpd` mit `slow.sh` (3 s) und `fast.sh`.

- `MODE=record SHUTDOWN_TIMEOUT=8s start`: schneller Request 200; langsamer Request unterwegs, `SIGTERM`: der langsame Request endet mit seiner Antwort (`{"value":[{"ID":2}]}`), ein neuer Request 0,5 s danach: Verbindung abgelehnt, Exit 0 nach rund 2 s. Ablage danach (Probe über den sqlite-Adapter): eine Session `complete`, zwei Exchanges `complete`. `blobs` ist `drwx------`.
- Doppelsignal bei laufendem Request: zweites `SIGINT` Exit 130, zweites `SIGTERM` Exit 143 (`SPEC-138` Nr. 7 c).
- Startfehler, je Exit 1 und genau eine Zeile ohne Wert: `MODE=mock` und `MODE` leer `ODR-E2001 MODE: …`; `DATA_DIR` ist Datei und unter Datei `ODR-E2001 DATA_DIR: …`; `UPSTREAM_CA_FILE` fehlt `ODR-E2001 UPSTREAM_CA_FILE: Datei nicht lesbar` (der Pfad steht nicht in der Meldung); `LISTEN` belegt `ODR-E2001 LISTEN: Adresse nicht bindbar`.
- `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true`, `record`: genau eine WARN-Zeile `start: Zertifikatsprüfung zum Upstream abgeschaltet ereignis=tls_pruefung_aus` vor `start: bereit`, Exit 0. `replay`: Request ohne Header 400, Exit 0 nach `SIGTERM`, keine Warnung.
- Logzeilen tragen nur Zeit, Stufe, Text und die Attribute `mode`, `listen` (Nr. 8) beziehungsweise `ereignis`; kein Pfad, keine URL, kein Fehlertext. `stdout` bleibt leer.

Ergebnis: Durchstich belegt, kein Befund.

## Findings

| ID | Kategorie | Quelle | Pfad | Befund | Beleg |
|---|---|---|---|---|---|
| V-1 | MEDIUM | `AGENTS.md` §3.15 (Zusage ohne Fänger), `SPEC-139` Nr. 7 („jeder weitere und jeder gleichzeitige Aufruf wartet auf dessen Ende und liefert denselben Wert“), Plan B5 | `src/bootstrap/lebenszyklus.go` · `Herunterfahren` | Der zweite Aufruf, der `nil` statt des Wertes des ersten liefert, bleibt in allen Läufen grün (`B5_zweiter_nil`: `erster := false; Do(…erster = true…); if !erster { return nil }`). `TestHerunterfahrenReihenfolge` ruft „nacheinander“ und „gleichzeitig“ nur mit Ergebnis `nil` auf; die Fehler-Fälle (`Fehler/…`) rufen einmal. Folge: ein Aufrufer, der nach einem gescheiterten Herunterfahren ein zweites Mal ruft (der Plan nennt „wiederholt“ als Zelle des Rasters), erhielte Erfolg. Das gleichzeitige Warten ohne Datenrennen (zweiter Aufruf kehrt früh zurück, Ergebnis geschützt) ist ebenfalls unsichtbar; die Variante mit ungeschütztem `h.ergebnis` färbt der Detektor (`B5_gleichzeitig_wartet_nicht`: rot in 10 von 10 Prozessen mit `DATA RACE`, grün ohne `-race`) | Mutanten `B5_zweiter_nil`, `B5_gleichzeitig_wartet_nicht` |
| V-2 | MEDIUM | `AGENTS.md` §3.15 (Allaussage), `SPEC-138` Nr. 8 f („`errors.Is` gegen `ctx.Err()`“) | `src/bootstrap/start.go` · `abschluss` | Die Klasse „Kontextfehler eines Schritts, auch umhüllt“ ist nur mit dem blanken `ctx.Err()` gefangen: `errors.Is(e, ctx.Err())` durch `e == ctx.Err()` ersetzt bleibt grün (`S2_gleich_statt_is`). Alle Fakes von `TestAbbruchInJederPhase` liefern `ctx.Err()` unumhüllt; die Spec nennt `errors.Is` ausdrücklich, und die echten Ports (`sqlite`, `Abbilden`, `SPEC-109` Nr. 6: „auch wenn ein Fehler beides enthält“) können umhüllen. Folge im Betrieb: ein Signal während eines Schritts, dessen Fehler den Kontextfehler umhüllt, ergäbe Exit 1 mit Meldung statt Exit 0 | Mutant `S2_gleich_statt_is` |
| V-3 | MEDIUM | `AGENTS.md` §3.15/§3.16, `SPEC-139` Nr. 2 („Fehler der Ablage kommt unverändert zurück und wird nach `SPEC-109` Nr. 5 abgebildet“), Raster `Build` | `src/bootstrap/wurzel.go` · `build`, Fehler von `Oeffne` | Der Fehler von `Oeffne` außerhalb des Kontextfehlers durch `errors.New("ablage")` ersetzt bleibt grün (`N7_ablage_fehler_flach`): `TestStartFehler` (Phase Ablage) liefert `errPhase`, dessen Abbildung ohnehin `ODR-E9001` ist, also gleich dem Mutanten; `TestModus` prüft den Fehler der Ablage nicht. Die Zelle des Rasters (`DATA_DIR` fehlt, `odata-recorder.db` ist Verzeichnis, Datei beschädigt: abgebildet) hat einen Ort und keinen Fänger. Folge: ein Fehler der Ablage (etwa `ErrBusy`) erschiene als `ODR-E9001` statt seines Codes | Mutant `N7_ablage_fehler_flach` |
| V-4 | LOW | `AGENTS.md` §3.15 (eine Fixture gleicht nie dem Wert der Mutation) | `src/bootstrap/wurzel_test.go` · `TestDiensteAnlegen` | Die Attribute des Logeintrags werden mit `strings.Contains` geprüft: `ereignis=dienst_abweichend` trifft auch `dienst_abweichend2` (Mutant `B2_ereignis` grün), ebenso `error_code=ODR-E2001` und `service_id=a` ihre Verlängerungen. `SPEC-139` Nr. 8 c nennt die Werte fest. Wirkung gering: Auswertungen suchen nach dem Wert | Mutant `B2_ereignis` |
| V-5 | LOW | `AGENTS.md` §3.15 (Zusage ohne Fänger, hier ohne beobachtbare Wirkung im Betrieb) | `wurzel.go`, `lebenszyklus.go`, `start.go` | Vier Zusagen von `SPEC-139` haben keinen Fänger, jede ohne sichtbare Wirkung in dieser Welle: (a) `log` nil verwirft (Nr. 1): `slog.Default()` statt Discard bleibt grün (`N1`); (b) die Attribute `mode`, `listen` von `start: bereit` (Nr. 8): feste Werte bleiben grün (`N2`); (c) die Telemetrie des Replay-Dienstes schreibt nach `io.Discard` (Nr. 3): `os.Stderr` bleibt grün (`N3`, der Adapter schreibt noch nichts, `SPEC-232`); (d) `Herunterfahren` unter `WithoutCancel(ctx)` statt `Background()` (Nr. 7): grün (`S1_background`), weil kein Test einen Wert im `ctx` des Herunterfahrens trägt; die Abbruchwirkung ist durch S1 gefangen. Entweder als akzeptiertes Negativ mit Grund und Folge in §6 eintragen oder je ein Fall (ein Logger-Fake, ein Wert im `ctx`) | Mutanten `N1`, `N2`, `N3`, `S1_background` |
| V-6 | LOW | `AGENTS.md` §3.7 (ein Kommentar beschreibt, was da ist) | `src/bootstrap/start_test.go` · Kommentar von `TestStartFehlerPaare` | Der Kommentar sagt, die Stelle des Ausfalls von `Serve` sei „offen (Plan §6)“; Plan §6 und `SPEC-139` Nr. 8 haben sie entschieden („zuletzt“). Beim Nachzug des Fängers (V-7) mit zu berichtigen | Lesen |
| V-7 | INFO | bekannt, nicht doppelt gezählt | `TestStartFehlerPaare` | Fänger der Zeile „Ausfall von `Serve` zuletzt“ fehlt (Plan §6, Verifier-Fix-Runde): `B10_serve_zuerst` (Zeile des Ausfalls vor die des Herunterfahrens) ist grün. Offene Rückgabe R-8 (`AGENTS.md` §3.14 gegen `SPEC-128`) und Block-Schätzung 2120 gegen Tabelle 2445 ebenfalls bekannt | Mutant `B10_serve_zuerst` |
| V-8 | INFO | DoD | Plan §2 | Haken 4 bis 9 stehen offen (Review durchgeführt, Doku-Update, Closure-Notiz, Register, Risiko-Ausgänge, Paarungen). Der Review-Report liegt vor, Haken 4 ist setzbar; die übrigen gehören der Closure. §6 trägt noch „weiter offen“ (Risiko 2, R-8) | Lesen |

Kein HIGH. Die Mutanten von V-1 bis V-5 sind in Scratchpad-Kopien gefahren; der Worktree-Code ist unverändert.

## Auflösung der Tabelle *Zusage · Fänger* (jede Zeile mit eigener Mutation, gegen die Klasse)

Eigene Mutationen heißt: nicht die der Plan-Tabelle, sondern Elemente der Klasse, die dort oder im Review nicht genannt sind. Rot = Test mit Testmeldung. „Rate“ = Prozesse mit `-race -c`, 10 je Mutante; ungemutiert 0 von 10 rot.

| Zeile | Klasse (Menge) | Eigene Mutationen und Ergebnis |
|---|---|---|
| S1 | Quellen des `ctx` im Herunterfahren; Anhalten vor CloseAll | `Anhalten` entfällt: rot (`TestHerunterfahrenOhneAbbruchDesAusloesers`, `TestHerunterfahrenReihenfolge`, `TestServeAusfall`, `TestStartFehler`); `ctx` statt `WithoutCancel`: Rate 10 von 10; `Background()` statt `WithoutCancel(ctx)`: grün (V-5 d) |
| S2 | Schritt × Ort × Modus (18) | Vorprüfung nur im Modus `record` (Element `replay`, nicht genannt): rot `TestAbbruchInJederPhase`, Rate 10 von 10; alle Vorprüfungen entfernt: Rate 10 von 10; `==` statt `errors.Is`: grün (V-2) |
| S3 | Kontext × Verstoß (4) | `LadeOderMelde(…, os.Stderr)`: rot (`TestKonfigFehlerBeiAbbruch`, `TestStartLiestUmgebung`, `TestCADatei`) |
| S4 | Phase × Kontext (30); Fehlerzeilen | erste Session erst nach dem Aufruf als begonnen: rot (`TestStartFehler`, `TestAbbruchInJederPhase`); Wert (`LISTEN`) in der Meldung: rot (`TestStartFehler`, `TestStartFehlerPaare`) |
| B1 | benachbarte Paare, beide Modi, Dienstreihenfolge | Wartung und Dienste nur im Modus `replay` vertauscht: rot; Dienste rückwärts: rot; Sessions rückwärts: rot (`TestStartReihenfolge`) |
| B2 | Dienst × Zustand der Ablage; Meldungen und Logeinträge | `Get`-Fehler verschluckt: rot (`TestHandleRandformen`); Root nur nach Länge verglichen (Fixture `/x` gegen `/a`): rot; Version nur nach Länge: rot; `service_id` aus dem Root: rot; Stufe WARN: rot; `ereignis` um ein Suffix verlängert: grün (V-4) |
| B3 | drei Verbraucher des einen Wächters | zweiter Wächter für die Wartung (nicht genannt): rot `TestEinWaechter`; für die Pipeline (`record`): rot; für den Replay-Dienst (`replay`): rot |
| B4 | Schalter × Modus (4); Warnung genau einmal | Warnung doppelt: rot (`TestPruefenUndWarnung`); `Pruefen` invertiert: rot (`TestPruefenUndWarnung`, `TestWerteErreichenVerbraucher`) |
| B5 | Reihenfolge, genau ein `Schliessen`, Warten auf `Serve` | `Schliessen` im Modus `replay` entfällt: rot (4 Tests); Warten auf `Serve` entfällt: Rate 10 von 10; zweiter Aufruf liefert `nil`: grün (V-1); gleichzeitiger Aufruf wartet nicht, Ergebnis ungeschützt: grün ohne `-race`, Rate 10 von 10 mit Detektor |
| B6 | Dauer von `CloseAll` {0, <, =, >} × Modus; Frist aus `SHUTDOWN_TIMEOUT` | Frist aus `RetentionMaxAge`: rot (`TestFristRest`); ohne Untergrenze 0: rot; `session.Frist` aus `SQLiteBusyTimeout` (nicht genanntes Feld): rot (`TestWerteErreichenVerbraucher`) |
| B7 | 20 Felder; Uhr an 4 + 3 Stellen; Logger | Uhr eigene `NewSystemuhr()` je Stelle: Wartung, Wächter, Telemetrie, Session, Recorder je rot (5 von 5; die Wartung im Modus `replay` teilt die Zeile mit `record`); Logger der Wartung `slog.Default()`: rot; `RedactHeaders` mit `RedactQueryParams` vertauscht: rot; `replay.Dienst.Root` aus der ID: rot; `UpstreamURL` aus dem Root: rot; `Pfade` des Replay-Dienstes `nil`: rot. Die Menge der Uhren (7 Stellen) ist vollständig: `grep` über alle `Konfig` in `hexagon/services` und `adapters` findet `Uhr` oder `Clock` in Wartung, Recorder, Session, Telemetrie und (als Argument) Wächter; Workspace und Import gehören dem Folge-Slice |
| B8 | Modus; `UPSTREAM_CA_FILE` (8) | `Build` nimmt alles außer `mock` an: rot (`TestModus`); `Build` liest die CA-Datei auch in `replay`: rot (`TestCADatei`); Fehler der Ablage abgeflacht: grün (V-3) |
| B9 | `blobs` fehlt / vorhanden × Modus (8) | `0o701`: rot; `os.Mkdir` statt `MkdirAll`: rot (`TestBlobVerzeichnisRechte`, `TestDiensteAnlegen`) |
| B10 | 13 Paare der Quellen | Fehler des Herunterfahrens als ein Wert angehängt (`append(fehler, err)`): rot; `abschluss` schreibt rückwärts: rot; `Fehlgeschlagen > 1`: rot; `Serve` zuerst: grün (V-7, bekannt) |
| B11 | Aufrufstellen × Modus (6 + 5) | `Create` mit `Background()`: rot; `Get`, `Startup`, `OpenInitialSession` je mit `WithoutCancel`: rot (3); `Open` mit `Background()`: rot; Netz `tcp6`: rot (`TestAufrufstellen`; viele Folgetests fallen mit, weil der Container kein IPv6 bindet) |
| B12 | fünf Zugänge gleichzeitig mit `Hochfahren` | Sperre in `Adresse` entfällt; `Hochfahren` schreibt `h.record`, `h.sessions`, `h.wartung` oder `h.fehler` ohne Sperre (vier nicht genannte Zugänge): je Rate 10 von 10 (`DATA RACE`) |
| Handle | Zustände von `Hochfahren` | zweites `Hochfahren` zugelassen: rot (`TestHandleRandformen`) |

Raten (10 Prozesse, `-test.count=1`): ungemutiert 0 von 10; S1 `ctx` 10 von 10; S2 ohne Vorprüfung 10 von 10; S2 nur `record` 10 von 10; B5 ohne Warten auf `Serve` 10 von 10; B12 fünf Mutanten je 10 von 10. Das deckt sich mit den 30 von 30 des Plans.

## Raster §6 (Operation × Argumentart)

Jede Zelle trägt einen Ort (`SPEC-139` Nr., `SPEC-138` Nr. 8, `SPEC-110`, `SPEC-103`, `SPEC-106`, `SPEC-136` Nr. 14/15) oder „gilt nicht“ mit Grund; stichprobenhaft aufgelöst (`SPEC-110` Nr. 4, 9, 11; `SPEC-136` Nr. 14; `SPEC-139` Nr. 2, 6, 7, 8, 10). Nichts im Code entschieden: die beiden Randformen des Implementers (Schließen von `Fehler()`, `UPSTREAM_CA_FILE` in `replay`) stehen in `SPEC-139`/`SPEC-138` und haben Fänger (`TestServeAusfall`, `TestWarte`, `TestCADatei`). Zellen ohne Fänger: Fehler der Ablage abgebildet (V-3), „wiederholt“ für `Herunterfahren` mit Fehlerwert (V-1).

## Fixes des Reviews (R-1 bis R-9) gegen die Klasse

- R-1 (Uhr): Menge 4 + 3 vollständig (siehe B7); fünf Stellen einzeln mutiert, alle rot.
- R-2 (Paare): 13 Paare als Schleife; drei nicht genannte Mutanten rot; die Stelle von `Serve` offen (V-7).
- R-3 (beide Felder): beide Zeilen und beide Logeinträge in fester Folge gefangen; das Attribut `ereignis` nur als Präfix (V-4).
- R-4 (Aufrufstellen): sechs Stellen als Schleife; fünf Mutanten (`Create`, `Get`, `Startup`, `OpenInitialSession`, `Open`) rot, die zwei Fehlerformen `Background` und `WithoutCancel` je einzeln getrennt; `tcp6` rot.
- R-5 (`0700`): vier Zustände × zwei Modi; `0o701` und `Mkdir` rot.
- R-6 (Nebenläufigkeit): alle fünf Zugänge vom Detektor gefangen, vier davon mit nicht genannten Elementen.
- R-7 (offene Randformen): entschieden und gefangen.
- R-8: offen mit Adresse, bekannt.
- R-9: Kommentar von `gibFrei` nennt nur die Kopplung.

## Plan-vs-Code-Diff (Basis `e8c027c`)

Plan und Code stimmen überein: 13 Felder der Naht `bauteile` wie §3; Zugänge des `Handle` wie `SPEC-139`; `Start` ruft `LadeOderMelde`, `main` setzt `Konfig.Start` (am Binary belegt); Reihenfolge Ablage → Wartung → Dienste → Sessions → Listener, Herunterfahren `Anhalten` → `CloseAll` → `Schliessen(rest)` → Warten → Ablage; Plan-Kopf, §1, §3, §4, §6 nachgezogen; keine Datei außerhalb des Plans geändert (`src/bootstrap/*`, `main.go`, `spec/spezifikation.md` und Planungsdokumente). Hexagon: `a-check` grün, kein Hexagon-Paket importiert `bootstrap`. Commits tragen `LH-FA-*`. Abweichungen: keine über §6 hinaus.

## Negativbefund (geprüft)

- DoD Haken 1 bis 3: erfüllt; Haken 4 bis 9 offen, Closure (V-8).
- Sensoren (`make gates`, Größenmarke, `-race -count=20`): belegt, ohne Befund.
- Durchstich am gebauten Binary (Start, Request, `SIGTERM` mit laufendem Request, Exit 0/1/130/143, Logzeilen ohne Werte): ohne Befund.
- Tabelle *Zusage · Fänger* S1: Befund V-5 d, sonst ohne Befund.
- S2: Befund V-2, sonst ohne Befund. S3: ohne Befund. S4: ohne Befund.
- B1, B3, B4, B6, B9, B11, B12, Handle: ohne Befund.
- B2: Befund V-4. B5: Befund V-1. B7: ohne Befund (V-5 c, N3 betrifft `SPEC-139` Nr. 3, nicht die Uhr). B8: Befund V-3. B10: Befund V-7 (bekannt).
- Raster §6: Orte vollständig, nichts im Code entschieden; Zellen ohne Fänger siehe V-1, V-3.
- Rohbytes (`AGENTS.md` §3.18): ohne Befund (gegengelesen: alle Dateien von `src/bootstrap/`, `main.go`, Plan, Review).
- Hard Rules §3.9, §3.13: Plan nachgezogen, offene Rückgabe R-8 mit Adresse in §6; ohne weiteren Befund.
- Kommentare (`AGENTS.md` §3.7, §3.11): Befund V-6, sonst ohne Befund.

## Summary

0 HIGH, 3 MEDIUM, 2 LOW, 2 INFO (V-7 ist die bekannte Lücke, V-8 der Stand der DoD). Der Durchstich trägt. Die drei MEDIUM sind Zusagen, die ein Mutant unbemerkt passiert: der Wert eines wiederholten `Herunterfahren` (V-1), umhüllter Kontextfehler in `abschluss` (V-2), der Code eines Fehlers der Ablage (V-3). Vorschlag an den Planner: V-1 bis V-3 mit dem Fänger der `Serve`-Zeile (V-7) in der angekündigten Fix-Runde bündeln; V-5 als akzeptiertes Negativ mit Grund und Folge eintragen oder gefangen lassen.
