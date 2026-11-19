# Verifikation slice-bootstrap-m4 (Composition Root im Modus `hybrid`, M4-Test, Stufe 10)

**Rolle:** Verifier (einmal). **Datum:** 2026-11-19. **Stand:** Zweig `slice-bootstrap-m4`, HEAD `a7726090`. **Bezug:** [LH-FA-12](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), `SPEC-110` Nr. 2, 8, 12 a, `SPEC-139` Nr. 12, `SPEC-261`, `SPEC-284`, `SPEC-037` Nr. 7, `SPEC-200`.
**Plan:** [slice-bootstrap-m4](../plan/planning/done/slice-bootstrap-m4.md). **Modell:** claude-sonnet-5-5.
Der Review-Report `2026-11-18-slice-bootstrap-m4-review.md` ist als Liste der Befunde F1 bis F6 gelesen, nicht als Beleg; alles unter *Belege* ist selbst gefahren.

## Urteil

**Bestanden.** Kein HIGH, kein MEDIUM, zwei LOW, vier INFO. Alle Gates sind grün (`make gates` Exit 0, dazu `make test-race` frisch ohne Cache), der Code folgt `SPEC-139` Nr. 12 und `SPEC-110` Nr. 8, und von 20 eigenen Mutationen an Stellen, die ich selbst wählte, sind 18 rot aus dem benannten Grund, eine ist äquivalent (V-5) und eine bricht die Übersetzung (V-B7, zählt nicht). Die Nacharbeit F1 bis F6 deckt die Klasse, nicht nur die genannten Mutationen. Die 412-Sendung aus `welle-etag` ist in Code, Test und Plan getragen; offen bleibt nur, dass der Test den Zustand nach dem `412` nicht zurückliest (V-1, LOW). Die Closure und die DoD-Häkchen stehen aus; sie sind nicht Gegenstand dieser Rolle.

## Findings

| Nr. | Schwere | Ort | Befund |
|---|---|---|---|
| V-1 | LOW | `tests/integration/m4_test.go` · `ohneAusloeser`, Fall `fachliche 412` (Zeile 243); Plan §1 *Übernimmt (Fall 412)* und DoD Punkt 2 | Die Sendung sagt: `ODR-E6010`, 0 Verbindungen **und Zustand unverändert**. Der Fall prüft Status, Code, fehlende Kennzeichnung und 0 Verbindungen und Requests (`keinKontakt`), liest den Zustand der Entity danach nicht zurück. Die Zeile M12 im Plan nennt nur „0 Verbindungen“. Die Hälfte „Zustand unverändert“ trägt allein die Transaktion des Mock (`imWorkspace` kehrt vor `Put` mit `ODR-E6010` zurück, `schreiben.go:209`) und deren Tests aus `slice-if-match`; in diesem Slice hat sie keinen Fänger und kein benanntes Negativ. Der `If-Match`-Wert des Falls ist `"nope"` (nie gültig), nicht ein veralteter ETag einer früheren Version, wie DoD Punkt 2 wörtlich sagt; für die Zusage „kein Fallback bei `412`“ ist das gleichwertig, für die Lesart „veraltet“ nicht. Auflage an den Planner: entweder Rücklesen (`GET` der Entity 1 nach dem `PATCH` gleich dem Stand davor, ein Fall mit einem ETag aus `GET` vor einem ersten `PATCH`) oder die Zeile in DoD und §1 auf „0 Verbindungen“ zurückziehen und die Zustands-Hälfte als akzeptiertes Negativ mit Grund (Transaktion des Mock) und Folge führen. Kein Code betroffen. |
| V-2 | LOW | Plan §7 *Stand des Implementers*, Absatz *Offen* und §3 Absatz *Handbuch* | Zwei Aussagen sind nach Nacharbeit und Review nicht nachgezogen (`AGENTS.md` §3.9): §3 *Handbuch* sagt „die Mutationen am Handbuch … sind nicht gefahren“, der Reviewer hat sie gefahren (Review I3: Zeile gestrichen macht `make katalog-check` rot, „später“ macht `make handbuch-check` rot); §7 *Offen* (2) nennt Platzhalter und nicht ausgewertete Raten, die in §3 inzwischen gefüllt sind. Der Zustand der Zusage M21 ist damit im Plan schlechter beschrieben, als er ist. Nachzug bei der Closure; keine Wirkung auf den Code. |
| V-3 | INFO | Plan DoD Punkt 3, Zeile M21, §6 | Die Beispiel-Datei für den Modus `hybrid` entfällt (Entscheidung des Architect, 2026-11-19, `SPEC-200` Nr. 6 kennt keine Umgebung). DoD Punkt 3 und Zeile M21 tragen die Entscheidung; das akzeptierte Negativ (Sätze über das Verhalten am Binary fängt nur `TestM4` und `TestE2EHybrid`, nicht das Handbuch-Gate) hat Grund und Folge. Ich habe die Sätze des Kapitels `Modus hybrid` gegen `SPEC-139` Nr. 12 b, c, `SPEC-261` Nr. 4 bis 6, `SPEC-284` Nr. 3, 5, 6, 8, 9 und `SPEC-195` Nr. 5 gelesen: kein Widerspruch (Workspace-Header fehlt: `ODR-E5003`, unbekannt: `ODR-E7001`; Kategorien; Vorrang des Replay; Host-Einschränkung; Warnung nur bei `enabled`; `UPSTREAM_TLS_INSECURE_SKIP_VERIFY` ist ein Schlüssel der Tabelle). |
| V-4 | INFO | `tests/integration/m4_test.go` · `ohneAusloeser` mit `offen(up, false)` | Die offenste Policy trägt beide Replay-Kategorien; der Replay geht dem Upstream vor (`SPEC-261` Nr. 4). Eine Mutation „412 ist Auslöser“ wird deshalb am Pfad zum Replay rot (`400 ODR-E5003`, V-I1), nicht an einer Verbindung. Dass der Pfad zum Upstream für `412` gefangen ist, zeigt V-I7 (Umgehung des Replay): rot mit „1 neue Verbindungen am Referenzdienst, erwartet 0“. Kein Befund, nur die Lesart der Zeile. |
| V-5 | INFO | `src/hexagon/services/hybrid/verteiler.go:100` gegen `policy/entscheidung.go` | Die Mutation V-I4 (`ODR-E5002` in `ausloeserVon` als Replay-Miss) bleibt unter `TestM4` grün: der Dienst ruft `Entscheide` nach dem Replay nur bei `ODR-E5001` (`verteiler.go:100`), der Zweig der Policy ist vom Root aus nicht erreichbar. Äquivalent auf der Ebene des Slice; die Policy hat eigene Tests (`slice-fallback-policy`). Kein Befund. |
| V-6 | INFO | Plan DoD, §7 | Alle DoD-Häkchen offen, §7 *Wird bei der Closure gefüllt*, kein Closure-Lerneintrag, Beobachtungs-Register nicht fortgeschrieben (Zähler `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` und `…/gegenprobe-haengt-an-der-laufzeit-des-produkts` stehen für den Integrationsschritt aus). Erwartet vor der Closure. |

## DoD Punkt für Punkt (Plan §2)

| Punkt | Zustand | Beleg |
|---|---|---|
| 1 Verdrahtung `hybrid` (`HybridUseCase`, vier Schlüssel über `NeuePolicy`, `Fallback` in der Tabelle der Verbraucher, `hybrid.Konfig` nach `SPEC-284` Nr. 1) | erfüllt | `wurzel.go` `verdrahteHybrid`: `Dienste` mit `Root` und `UpstreamURL`, `Mock`, `Replay`, `Upstream` (immer gebaut), `Policy` aus `NeuePolicy`, `MaxBytes` gleich `MaxCaptureBody`; `h.verarbeiten = hy.Hybrid`; `Hochfahren` ohne `oeffneErsteSessions` und `Herunterfahren` ohne `CloseAll` (nur `modusRecord`); Snapshot- und Workspace-Dienst entstehen in jedem Modus (`wurzel.go:279` bis `287`), wie `SPEC-139` Nr. 12 a verlangt. Fänger: `M4WahlJeModus`, `M4WerteDerPolicy`, `M4KonfigDesHybrid`, `WerteErreichenVerbraucher`; Mutationen V-B3, V-B4, V-B6, V-B7, V-B8 rot. |
| 2 M4-Test (Mock-Treffer 0; lesend mit/ohne Freigabe; POST mit/ohne Eintrag; Replay-Miss; `ODR-E5002`; `404`, `409`; **`412`**; Host-Fänger Z35; Start und `TestE2EHybrid`) | erfüllt, ein LOW (V-1) | `TestM4` (Fälle `ohneAusloeser`, `lesenderFallback`, `ohneFreigabe`, `strikteAbdeckung`, `mutierend`, `replayMiss`, `erschoepft`, `replayFallback`, `gleichzeitig`) grün im Gate; Stufe 10 grün in `make test-integration` (zehn Stufen, `e2e-abdeckung` „10 Stufen, 10 Deklarationen, 0 ohne Kennung“, unverändert). |
| 3 Handbuch-Kapitel | erfüllt (ohne Beispiel-Datei, entschieden) | Kapitel `Modus hybrid` und die Zeilen `MODE`, `UPSTREAM_URL`, `STRICT_COVERAGE`, `ETAG_ENABLED`, `PUBLIC_BASE_URL`, vier Schlüssel; enthält Wechselwirkung (`SPEC-261` Nr. 4 e), Host-Einschränkung (Nr. 5 c), `X-Recorder-Fallback`, Links unverändert, `ODR-E5003` erst im Replay-Fallback, Body über `MAX_CAPTURE_BODY`; `make handbuch-check` und `make katalog-check` im Gate grün. |
| 4 `make gates` grün | erfüllt | siehe Belege |
| 5 Review mit Report | vorhanden | `2026-11-18-slice-bootstrap-m4-review.md` (F1 bis F6); die Nacharbeit hat kein zweites Review, die Verifikation löst sie auf (unten) |
| 6 bis 10 Doku-Update, Closure, Register, Risiken, Paarungen | offen, vor der Closure | V-6; Risiken in §6 tragen Ausgänge (offen nur: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` und die Reihenfolge zu `slice-bootstrap-m6`, beides mit Adresse) |

## Entscheidungs-Konformität

| Stelle | Ergebnis |
|---|---|
| `SPEC-110` Nr. 2, 8, 12 a (Modus `hybrid` lieferbar, Start ohne Session, Schlüsselmenge) | konform: `pruefeGrundwerte` lässt `hybrid` zu und verlangt `PublicBaseURL` wie `mock`; V-B1 (Verlangen aufgehoben) rot (`TestModus`). |
| `SPEC-139` Nr. 12 a bis d | konform: Bau (a), Start ohne Session und Workspace (b), Herunterfahren ohne Schritt (2) (c; Zusammenspiel mit wartender Anfrage ist akzeptiertes Negativ, `SPEC-037` Nr. 10 n), Werte (d). |
| `SPEC-139` Nr. 2 und 4 mit Architect-Entscheidung 2026-11-19 (`UPSTREAM_CA_FILE` in `hybrid` wie in `record`, auch bei `disabled`) | konform: `vorbereiten` liest die Datei in `record` und `hybrid`; Text in Nr. 2 und Nr. 4 nachgezogen (Diff gelesen); V-B9 (nur bei `enabled`) rot in `TestM4BuildFehler`. |
| `SPEC-261` Nr. 3 c, 4, 5, 6 (Auslöser, Vorrang Replay, Freigabe, Schreibfreigaben) | konform; V-I1 bis V-I3, V-I5 bis V-I7 rot (Menge der Auslöser und Methoden). |
| `SPEC-284` Nr. 1, 3 d, 5 c, 8, 9 | konform: `Konfig`, `Upstream` immer gesetzt, Host der Policy am echten Adapter (M13/Z35, Mutation `Ziel` mit anderem Port und Host rot laut Plan, V-B6 rot für `UpstreamURL` je Service). |
| `SPEC-037` Nr. 7 mit Architect-Änderung (Replay-Miss mit `PATCH` auf unbekannte Entity und `POST` auf Entity, je `ODR-E5001`, 0 Verbindungen; `POST` auf Collection nie Replay-Miss) | konform: `TestE2EHybrid`, Test „Replay-Miss“, Zeilen 184 bis 189 fahren genau diese zwei Fälle; Kopfsatz und *Stand* von `SPEC-037` nennen Stufe 10 als geliefert (zehn Stufen = Zahl der Stufen in Nr. 7); `harness/README.md` sagt „zehn Stufen“ und führt das Negativ der Stufe 10 (Fehler des Baus) unter *Nicht behauptet*. |
| `SPEC-200` (Handbuch, Beispiel-Format) | konform: keine Beispiel-Datei für `hybrid`, Entscheidung des Architect, `SPEC-200` unverändert. |
| 412-Sendung aus `welle-etag` | im Code/Test getragen (V-I1, V-I7 rot), Plan §1, DoD und §3 M12 nennen sie; Zustand nicht zurückgelesen: V-1. |

## Tabelle Zusage · Fänger: je Zusage Mutation und roter Test

Spalte *eigen* sind Mutationen, die ich gewählt und gefahren habe, an Elementen der Menge, die der Plan oder das Review nicht nannte. Treiber: `docs/plan/planning/slice-bootstrap-m4-mutationen/r2drv.sh` (Container ohne Netz, `--memory=4g --cpus=4`, unter `flock`, nacheinander); Listen im Scratchpad (nicht eingecheckt): Paket `src/bootstrap` neun, Paket `tests/integration` zehn Mutationen in zwei Listen. Load des Hosts 3,2 bis 3,7 (`/proc/loadavg`), Messender der Verifier.

| Zeile | Zusage | eigene Mutation | Ergebnis (roter Test, Meldung) |
|---|---|---|---|
| M01 | `hybrid` bindet genau den Hybrid-Dienst; `mock`/`replay` nie; kein Session-Aufbau | V-B5: `Hochfahren` öffnet Sessions in jedem Modus außer `mock` (also auch `replay`, neues Element) | rot, `TestM4WahlJeModus` („Hochfahren im Modus hybrid/replay: Panik …“) und `TestM4Lebenszyklus` |
| M01 | gleiche Zusage, Bindung | V-B7: `h.verarbeiten = rep.Replay` statt `hy.Hybrid` | Übersetzungsfehler (`hy` unbenutzt), zählt nicht; die Mutation des Plans „`hybrid` bindet `MockUseCase`/`ReplayUseCase`“ ist laut Plan und Review rot gesehen (Review I2 hat die Liste neu gefahren) |
| M02 | die vier Schlüssel erreichen `policy.Werte`; Fehler von `NeuePolicy` wird gemeldet | V-B8: Fehler von `NeuePolicy` verworfen | rot, `TestM4BuildFehler` („Fehler model: ungültiger Wert“) |
| M03 | `MaxBytes` gleich `MAX_CAPTURE_BODY` | V-B4: `MaxBytes + 1` (Grenzfall um eins) | rot, `TestM4KonfigDesHybrid` („MaxBytes 12346, erwartet 12345“) |
| M03 | `UpstreamURL` und `Root` je Service | V-B3: `Root` aus `d.ID`; V-B6: `UpstreamURL` des ersten Dienstes für alle (der Plan nannte den ersten Dienst für `Root`, `UpstreamURL` ebenso; V-B3 ist ein neues Element, `ID` statt `Root`) | beide rot, `TestM4WahlJeModus`/`TestM4WerteDerPolicy` bzw. `TestM4KonfigDesHybrid` („Referenzdienst von c: 0 Verbindungen“) |
| M03 | `PUBLIC_BASE_URL` in `hybrid` verlangt (`SPEC-110` Nr. 8) | V-B1: nur im Modus `mock` verlangt | rot, `TestModus` („Ablage 1-mal geöffnet … erwartet je 0“) |
| M22 | `UPSTREAM_CA_FILE` in `hybrid`; Warnung `tls_pruefung_aus` genau einmal bei `hybrid` und `enabled` | V-B9: CA-Datei nur bei `enabled` gelesen; V-B2: Warnung auch im Modus `replay` (neues Element der Menge der vier Modi) | beide rot, `TestM4BuildFehler` und `TestM4UpstreamClientUndWarnung` („1 Einträge tls_pruefung_aus, erwartet 0“) |
| M12 | kein Auslöser, kein Upstream (404, 409, **412**, `ODR-E6005`, `ODR-E6007`) | V-I1 (412), V-I2 (409, nicht genannt), V-I3 (`ODR-E6005`, nicht genannt), V-I6 (`ODR-E6007`, nicht genannt) als Auslöser; V-I7: `412` geht unter Umgehung des Replay direkt an den Upstream | alle rot, `TestM4`, Meldung z. B. „fachliche 412: 200 "upstream", erwartet 412 ODR-E6010 ohne Kennzeichnung“ und „1 neue Verbindungen am Referenzdienst, erwartet 0“ (V-I7) |
| M16 | `OPTIONS`/jede andere Methode erreicht nie den Upstream | V-I5: `OPTIONS` als lesend (Element der Menge der zehn Methoden, im Plan nicht als Mutation genannt) | rot, `TestM4` („OPTIONS route=false hostFrei=true: 200 "upstream", erwartet 501 ODR-E6002“) |
| M17 | `ODR-E5002` erreicht den Upstream nie | V-I4: `ODR-E5002` als Replay-Miss in der Policy | grün, äquivalent (V-5): der Dienst ruft `Entscheide` nach dem Replay nur bei `ODR-E5001` |
| M11 | Zähler nach Methode und Verbindungen im Referenzdienst | V-I8: Verbindung zählt auch bei `StateClosed`; V-I9: Methode zählt doppelt (beide nicht im Plan) | rot, `TestM4` („GET ODR-E6003: 2 neue Verbindungen am Referenzdienst, erwartet 1“; „2 neue Requests mit der Methode GET, erwartet 1“) |
| M19 | Liste der Requests unter Sperre, Wirkungs-Fänger | V-I10: die Liste verliert Einträge (Obergrenze 1000) | rot, `TestReferenzdienstZaehlerGleichzeitig` („1000 festgehaltene Requests, erwartet 128000“) |
| M10, M20 | Stufe 10 am Binary | eigene Mutation an einer Kopie des Baums: `Hybrid` bindet `m.Mock` statt `hy.Hybrid` (`make test-integration`, netzlos) | rot, `TestE2EHybrid/Happy_Path`, `Boundary/POST_mit_Eintrag`, `Boundary/Replay-Miss` (nicht Aufbau: die Stufe startet und antwortet) |

Zusammenfassung: 20 Mutationen gefahren (9 im Paket `src/bootstrap`, 6 plus 4 im Paket `tests/integration`, dazu die Stufe 10 am Binary), davon 18 rot aus dem benannten Grund, 1 äquivalent (V-I4), 1 Übersetzungsfehler (V-B7). Die Mengen der Nacharbeit (F1 bis F6, `AGENTS.md` §3.17) sind gegen den Wortlaut gebildet (Plan §3, *Nacharbeit F1 und F2*: 14 Werte, 39 Zellen mit Verbraucher, 78 geprüfte Zellen); meine Mutationen V-B2, V-B3, V-B5 und V-B6 liegen an Elementen dieser Mengen (Modus, `Root`, Dienst), die weder Befund noch Plan nannten, und sind rot: die Klasse ist gedeckt.

## Belege der Sensoren (selbst gefahren)

**`make gates`, einmal vollständig, Exit 0** (Log im Scratchpad der Sitzung, ein Container zugleich, sonst lief nichts). Darin: `a-check` 0 Befunde; `docs-check` „d-check: 954 Datei(en) geprüft, 0 Befund(e)“; `baseline-verify` v6.18.1 OK, 54 Dateien; `migration-check` ok; `test-contract`, `test-determinism` (100 Läufe, 0 abweichend), `test-integration` zehn Stufen grün; `groessenmarke slice-bootstrap-m4.md: ok, Messung 2222 von Marke 2500` (Messpunkt: 16 Zeilen, 2 nebenläufig, 2 Fakes, Produktcode 130, Schätzung 1820; Formel stimmt mit dem Block in §4 überein); `zeichen: ok, 1972 Datei(en)`; `handbuch-check`, `katalog-check`, `lint-profil`, `lint`, `schema-check` ohne Befund (Kette Exit 0).
Die Stufen `build`, `test` und `race` des Images kamen im Lauf aus dem Layer-Cache (gleiche Eingaben wie der Lauf des Implementers). Deshalb habe ich **`make test-race` danach frisch ohne Cache gefahren** (`DOCKER_BUILD="docker build --progress=plain --no-cache"`): `pruefung race: ok`, Exit 0, kein `DATA RACE`. Die Unit-Tests der Pakete `src/bootstrap` und `tests/integration` laufen außerdem in meinen Mutationsläufen als Nullprobe (grün) und im Paket-Lauf der Container.
`git status` des Worktrees war nach den Läufen leer (`make e2e-abdeckung` meldet „unveraendert“); die Mutationen liefen im Container und an einer Kopie im Scratchpad, die ich gelöscht habe; im Repo liegt nur dieser Report.

## Plan-vs-Code-Diff

| Plan §3 | Code | Abweichung |
|---|---|---|
| `src/bootstrap/` Wahl je Modus, `hybrid.Konfig`, `NeuePolicy` rufen, `konfig*.go` unberührt | `wurzel.go`, `start.go` geändert, `konfig*.go` nicht im Diff | keine |
| `httpodata` ohne Änderung | nicht im Diff | keine |
| Tests `m4_wurzel_test.go`, `wurzel_test.go`, `tests/integration/m4_test.go`, Referenzdienst nach Methode, `tests/e2e/stufe_hybrid_test.go`, `referenz_test.go` | alle im Diff | keine |
| Treiber Stufe 10 (drei Zeilen der Deklaration, `--rm`, kein fester `--name`), `e2e-abdeckung.md`, `harness/README.md` (zehn Stufen, *Nicht behauptet*) | Diff: vier Zeilen im Treiber (`e2e_abdeckung` und `stufe 10`), kein neues `docker run`; README-Zeile und Absatz stimmen | keine |
| `docs/user/` Kapitel, Zeilen der Einstellungen | Diff gelesen (siehe V-3) | keine |
| `SPEC-037`, `SPEC-139`, `SPEC-110` (Architect) | `SPEC-037` Kopfsatz, *Stand*, Nr. 7; `spezifikation.md` Nr. 2, 4 (CA-Datei) | keine; Änderungen stehen mit Ausgang *entschieden* und Architect in §6 |
| Größenmarke: Pfade `src/bootstrap/`, `tests/integration/`, `tests/e2e/`, `tools/harness/test-integration.sh` | Messung 2222 von 2500 | keine; der Merge von `main` in den Zweig liegt unter der Marke und ändert die Pfade nicht (Gate grün) |
| Mutationslisten und Treiber unter `docs/plan/planning/slice-bootstrap-m4-mutationen/` | im Diff, kein Pfad der Marke | keine; die Treiber leiten ihre Pfade aus dem Ort des Skripts ab (Review I4 behoben, ich habe `r2drv.sh` aus dem Ordner gefahren) |

## Ungeprüft

- Gegenproben (`make gegenproben`, Stempel `harness/gegenproben.stempel`): nicht gefahren; der Plan nennt sie als nicht gelaufen, der Integrationsschritt trägt den Stempel.
- Raten der Datenrennen und der Lost Updates (N = 20 Prozesse) habe ich **nicht** neu gemessen; ich habe den Wirkungs-Fänger nur an einer Mutation (V-I10) einmal rot gesehen. Die Zahlen in §3 *Nebenläufigkeit* sind die des Implementers, nicht von mir bestätigt.
- Mutationen am Referenzdienst von `tests/e2e` (`referenz_test.go`, Zeile M11 zweiter Teil) und die Mutationen der Handbuch-Zeile M21: nicht von mir gefahren (der Reviewer hat M21 gefahren, Review I3).
- Die Stufe-10-Mutation „`Start` meldet einen Fehler des Baus mit Exit 0“ (akzeptiertes Negativ, Stufe 10 grün, nur Stufe 4 rot): nicht neu gefahren; ich habe die Aussage nur gegen `harness/README.md` und Plan gelesen.
- Der Inhalt der Handbuch-Sätze, die kein Gate fängt (Klasse K6): gelesen, nicht mechanisch geprüft.
- Das Zusammenspiel von `SHUTDOWN_TIMEOUT` mit einer wartenden Upstream-Anfrage (akzeptiertes Negativ, `SPEC-037` Nr. 10 n).
- Die Closure-Pflichten (Lerneintrag, Register, Paarungen) liegen nicht vor und sind nicht Gegenstand dieser Rolle.
