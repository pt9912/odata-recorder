# Review-Report: slice-m2-referenzdienst — 2026-11-01

**Review-Art:** Code — gegen Plan, `SPEC-139` Nr. 1 bis 4, `SPEC-127`, `SPEC-129`, `SPEC-135` Nr. 9, 10, 13, `SPEC-045`, `SPEC-122` Nr. 2, Hard Rules (`AGENTS.md` §3.6, §3.7, §3.9 bis §3.18) und die Regeln von `a-check`.

**Gegenstand:** Diff `f286e03..b596ff7` (Zweig slice-m2-referenzdienst, 10 Commits): `src/bootstrap/wurzel.go` (+40), die Tests `src/bootstrap/import_workspace_test.go` (+277), `haken_test.go` (+25), `export_test.go` (+3), `tests/integration/m2_test.go` (+439), `tests/integration/referenzdienst_test.go` (+163), `Dockerfile`, `.dockerignore`, `.a-check.yml`, `harness/README.md`, `harness/gegenproben.stempel`, Plan `slice-m2-referenzdienst`.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-01

**Eingangs-Kontext:**

- Slice-Plan `slice-m2-referenzdienst` (§1 Beweis und Nicht-Beweis, §3 Ablauf, Tabelle M1 bis M10 und M9b, Tabelle am Handoff, Raten, §6 Raster und Risiken)
- `SPEC-139` Nr. 1 bis 4 und 10, `SPEC-127` Nr. 1, `SPEC-129`, `SPEC-135` Nr. 9, 10, 13, `SPEC-045`, `SPEC-122` Nr. 2, `SPEC-128`
- `LH-FA-01`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-FA-18`, [`ADR-0002`](../plan/adr/0002-architektur-gate-a-check.md), [`ADR-0007`](../plan/adr/0007-race-detektor-im-gate.md)
- `AGENTS.md` (Hard Rules). Der Bericht des Implementers lag dem Reviewer nicht vor; die Abweichung zu Zeile M2 (Mutation „Record-Dienst im Modus replay“ rot über Status und Body) und die offene Rückgabe zu `ZufallsIDs` sind aus dem Plan rekonstruiert.

**Messweise.** Mutanten nur in Kopien im Scratchpad, nie im Worktree (`git status` des Worktrees vor dem Report: sauber). Image `odata-recorder:race`, `--memory=4g --memory-swap=4g --cpus=4 --network none`, unter `flock /tmp/odata-recorder-heavy.lock`, Exit-Codes aus `PIPESTATUS`. Zwei Fänger-Gruppen je Mutant, beide mit `-race -count=1`: **INT** = `tests/integration -run TestDurchstichM2`, **BOOT** = `src/bootstrap -run "TestImport|TestWorkspace|TestZugaenge|TestKonstruktor"`. Rot heißt: der benannte Test scheitert an der Meldung, die zur Mutation gehört (je Mutant die Meldung gelesen; ein Mutant mit Aufbaufehler wurde verworfen und neu gebaut, siehe `impblobs`).

Grundstand der Kopie (Stand `b596ff7`): `go test -race -count=1 ./tests/... ./src/bootstrap/` Exit 0 (`tests/integration` 1,9 s, `src/bootstrap` 54,6 s). `make a-check` (gepinntes Image, Worktree read-only): 0 Befunde, Exit 0. Gegenproben-Stempel: `tools/harness/gegenproben-eingabe.sh` über den Stand liefert `58ff98ef…d52fd`, gleich `harness/gegenproben.stempel` (ob der Lauf grün war, prüft der Verifier). Größenmarke: `git diff --numstat 1984383..HEAD` über die Pfade des Blocks = **953** (Marke 1700, Schätzung 1190). Rohbytes (§3.18): `LC_ALL=C grep -P` auf `tests/integration/*.go`, `src/bootstrap/import_workspace_test.go`, `src/bootstrap/wurzel.go` und den Plan nach Backslash-`u`-Folgen, BOM, NBSP und Zeilentrennern: kein Treffer; die Dateien tragen nur UTF-8-Umlaute in Kommentaren.

## Beweis des Durchstichs (Frage 1)

`TestDurchstichM2` ist grün und beweist, was §1 verspricht; die Orakel sind unabhängig vom Code:

- **Orakel.** `MetadataHash` entsteht aus `sha256(csdlText)` der Fixture (`csdlHash()`), die Folge 7, 11, 13 aus `zaehlerFolge()` der Fixture; beide hängen nicht an Import oder Replay. Der Vergleich Aufnahme gegen Wiedergabe nimmt die Antwort, die der Client vom Recorder sah (Aufnahme), nicht eine Kopie aus dem Code. Er ist bewusst nicht der Vergleich gegen den Direktzugriff (`slice-protokolltreue`).
- **Header.** `gleich` bildet die Ausnahme-Menge aus der geladenen Konfiguration (`Content-Length`, `Content-Encoding`, `Date`, Hop-by-Hop, `VolatileHeaders`, `RedactHeaders`), vergleicht Mengengröße und Werte aller übrigen Header und verlangt, dass jeder redigierte Header in der Wiedergabe fehlt. `ETag`, `Content-Type`, `OData-Version` liegen außerhalb der Ausnahme.
- **Zähler.** Gegenprobe der Aufnahme (Requests gleich Client-Requests, Verbindungen größer 0, `GEHEIM-A` an jedem Request, `GEHEIM-B` in jeder Aufnahme-Antwort); Zähler vor und nach Durchgang A und B; Durchgang B erst nach `srv.Close()`.
- **Suche.** `durchsuche` läuft rekursiv über `DATA_DIR` und liest jede Datei auf Rohbytes. Im Lauf (Debug-Kopie, `t.Logf`) sah der Sucher vor `Herunterfahren(record)`: `odata-recorder.db`, `odata-recorder.db-wal`, `odata-recorder.db-shm` und 8 Dateien unter `blobs/sha256/`. Der Test behauptet nur die Datenbank (`slices.Contains(… "odata-recorder.db")`), nicht `-wal` und `-shm` (F-5).

Mutanten am Durchstich (INT / BOOT):

| Mutant | Wirkung | INT | BOOT | Rot durch |
|---|---|---|---|---|
| `up` | `verarbeiten` des Replay-Handle ruft vor `Replay` den Upstream (`GET /Products`) | rot | grün | Zähler `[2 16]` statt `[1 7]`, in A und B; keine Status- oder Body-Meldung |
| `up404` | Kontakt nur bei `Products(99)` | rot | grün | Zähler `[2 8]` statt `[1 7]` |
| `recrep` | Root verdrahtet im Modus replay `Record` als `verarbeiten` (Sitzungen nil gesetzt) | rot | rot | Status und Body (`ODR-E5009`, Status 409, erwartet `{"count":7}`); keine Zähler-Meldung (0 Treffer) |
| `norecidx` | `RedactHeaders` der Regelmenge leer | rot | rot | `GEHEIM-A`/`GEHEIM-B` in `odata-recorder.db-wal` (Suche 0) und `odata-recorder.db`; `Set-Cookie` in der Wiedergabe |
| `impvary` | Import erhält `VaryHeaders` leer | rot | rot | „Coverage Products: partial, erwartet Products complete“ |
| `leakrep` | Replay schreibt `Authorization` des Requests nach `DATA_DIR/replay.log` | **grün** | grün | (F-1) |
| `upstart` | Verdrahtung des Replay-Handle ruft den Upstream einmal | **grün** | grün | (F-4) |

Zu Frage 3: Der Zähler-Fänger hat eigene rote Mutationen, unabhängig von Status und Body. `up` und `up404` liefern die Replay-Antwort unverändert zurück; nur der Zähler wird rot, und zwar in der Form, die Zeile M2 zusagt. Umgekehrt bleibt `recrep` ohne Zähler-Meldung (der Record-Dienst weist ohne offene Session vor dem Weiterleiten ab). Die Abweichung im Plan ist damit richtig beschrieben, und die beiden Fänger sind nicht austauschbar.

## Raster Zeilen M1 bis M10, M9b gegen die Klasse (Frage 2 und 4)

Klassen laut Plan: 6 Felder von `SchluesselRegeln`, 13 Felder von `baselineimport.Konfig`, 9 Felder von `workspace.Konfig`, beide Modi, beide Zugänge. Die Tests zählen die Felder gegen `reflect` (6, 13, 9) und schlagen bei einem neuen Feld an.

| Klasse | Elemente | Mutanten des Reviewers | Ergebnis |
|---|---|---|---|
| Regelfelder am Import | 6 | `VaryHeaders`, `VolatileHeaders`, `RedactHeaders`, `RedactQueryParams` je leer; `InvalidEncoding` auf `literal`; `BodyMode` auf `bytes` | 6 von 6 BOOT rot („Regelfeld … an Import und Record nicht gleich“); INT nur bei `VaryHeaders` und gemeinsamem `RedactHeaders` rot, wie im Plan benannt |
| Felder von `baselineimport.Konfig` | 13 | `Dienste` (nur der erste; URL des ersten für alle), `Blobs` (andere Art `roh{waechter}`), `Metadaten` (`db.Metadata()`), `MaxCaptureBody`+1, `Uhr` (eigene `NewSystemuhr`), `IDs` (eigene `NewZufallsIDs`) | rot außer `IDs` (grün, bekannt, nicht doppelt gemeldet). `Sessions`, `Exchanges`, `Snapshots`, `Imports`, `MetadatenCodec`, `DatenCodec` sind nur über den Typ geprüft; ein gleichartiger Ersatz ist äquivalent, ein anderer Typ bricht den Aufbau (nicht mutiert) |
| Felder von `workspace.Konfig` | 9 | `Uhr` (eigene `NewSystemuhr`), `Dienste` (eigener Wrapper), `Blobs` (`roh{waechter}`), `Log` (`slog.Default()`), `IDs` (eigene `NewZufallsIDs`) | rot außer `Log` (F-3) und `IDs` (Klasse bekannt). `Workspaces`, `Snapshots` über den Typ; `Transaktion` (`== any(h.db)`) und `Wartung` (`== hd.Wartung()`) sind durch Identität gefangen (nicht mutiert) |
| Zugänge × Modus (M9b) | 2 × 2 | `Import` im Modus replay mit Attrappe (`impreplay`); typisierter Nullzeiger (`impnullptr`); `Workspaces` nur in `record` (`wsnurrec`); `Workspaces` füllt sein Feld beim ersten Aufruf (`lazyws`) | alle rot, Raten unten; die Zugänge sind in beiden Modi und für beide Zugänge in der Schleife geführt |

**Raten M9b** (selbst gemessen; `-race -c`, danach 30 Prozesse je Mutant mit `-test.run "TestZugaenge|TestKonstruktorFehler" -test.count=1`; Messender: Reviewer, Docker `--cpus=4`, Host mit Last 14 bis 15 durch fremde Läufe):

| Mutant | rot | davon mit `WARNING: DATA RACE` |
|---|---|---|
| ohne Mutation | 0 von 30 | 0 |
| `lazyws` (Feld beim ersten Aufruf gefüllt) | **30 von 30** | 30 |
| `impnullptr` (typisierter Nullzeiger) | 30 von 30 | 0 (direkter Test) |
| `impreplay` (Import im Modus replay) | 30 von 30 | 0 |
| `wsnurrec` (`Workspaces` nur in `record`) | 30 von 30 | 0 |

Die Zusage über Nebenläufigkeit (nur `lazyws`) ist damit nach §3.10 „in jedem Lauf rot“ (30 von 30), und der Detektor meldet das Rennen in jedem Prozess. Die Raten des Plans (30 von 30) stimmen. Die Raten für M1 und M2 habe ich nicht neu gemessen; die Mutanten `up` und `up404` sind deterministisch (kein Rennen beteiligt), M1 liegt in den Plan-Angaben des Implementers.

**M10.** Kopie a des Handversuchs nachgestellt: ein Test `tests/integration/rennen_test.go` mit absichtlichem Datenrennen (zwei Goroutinen schreiben `n++`), `pruefung race /src` im Race-Image: Exit 1, `WARNING: DATA RACE` im Paket `tests/integration`. Das Gate fährt `tests/` also, solange `COPY tests ./tests` in der Stufe `source` steht; Kopie b (ohne die Zeile) habe ich nicht wiederholt, sie ist im Plan als Handversuch mit dem Ausgang „grün trotz Rennen“ geführt, und die Gate-Verdrahtung ist laut `harness/README.md` §Sensors nicht behauptet.

## Findings

| ID | kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `AGENTS.md` §3.15 (Fänger führt jede Form der Eingabe), `LH-FA-14`, Plan Zeile M3 | `tests/integration/m2_test.go` · `d.sucheGeheimnisseNachHerunterfahren(vorher)` und `d.spieleAb()` | Die Suche nach `GEHEIM-A` und `GEHEIM-B` in `DATA_DIR` läuft nur vor dem Replay. Durchgang A und B senden `Authorization` erneut, danach liest keine Suche mehr; der Mutant `leakrep` (Replay schreibt den Header in `DATA_DIR/replay.log`) bleibt in INT und BOOT grün. Die Zusage der Zeile M3 („kein Geheimnis steht in `DATA_DIR`“) nennt keine Phase. | ja — Mutant `leakrep` (INT 0 rot) | Geheimnis-Suche nur in einer Phase des Ablaufs |
| F-2 | MEDIUM | `AGENTS.md` §3.6, [`ADR-0002`](../plan/adr/0002-architektur-gate-a-check.md), `ADR-0006`, `harness/vertraege/SPEC-039-architektur-gate.md` | `.a-check.yml` · `- "tests/**"` unter `composition_root` | Die Zeile nimmt `tests/**` nicht nur von den Schichtregeln aus (das nennt der Plan), sondern auch von `tech-leak`: eine Probedatei `tests/integration/probe_test.go` mit Import von `modernc.org/sqlite` und `encoding/xml` ergibt mit der Zeile 0 Befunde, ohne sie 5 (2 `tech-leak`, 3 `wrong-direction`). Plan §1 und §3 begründen nur die Schicht-Imports („nennt dessen Typen“); `SPEC-039` nennt die Ausnahme nicht. Ob das eine Lockerung nach §3.6 ist, entscheidet nicht der Reviewer; die Nebenwirkung auf `tech` steht in keinem Artefakt. | ja — Probe in zwei Kopien (`make a-check`-Image), Ausgaben wie genannt | Ausnahmefläche im Architektur-Gate ohne Nennung ihrer vollen Wirkung |
| F-3 | MEDIUM | `AGENTS.md` §3.15 (Zusage ohne Fänger ist MEDIUM), `SPEC-139` Nr. 3 und 4 (`Log` des Workspace-Dienstes) | `src/bootstrap/import_workspace_test.go` · `"Log":         got.Log != nil,` | Der Fänger der Zeile `Log` prüft nur auf nicht-nil. Der Mutant `wslog` (nur die Zeile des Workspace-Dienstes, `Log: slog.Default()` statt `h.log`) bleibt grün (INT 0, BOOT 0 rot). `SPEC-139` Nr. 3 nennt den einen übergebenen Logger, Nr. 4 den Verbraucher `Log`; der Plan führt die Zeile als geprüft (M6, „9 Felder“). | ja — Mutant `wslog` | Feld der Konfig nur auf nicht-nil geprüft |
| F-4 | LOW | `AGENTS.md` §3.15 (beide Seiten jeder Grenze), Plan Zeile M2 | `tests/integration/m2_test.go` · `vor := [2]int64{d.ref.verbindungen.Load(), d.ref.requests.Load()}` | Die Grundlinie der Zähler wird nach `Build(replay)`, `Hochfahren` und `Workspaces().Get` gelesen. Ein Kontakt zum Referenzdienst beim Verdrahten des Replay-Handle (Mutant `upstart`: ein `GET` in `verdrahteReplay`) bleibt in INT und BOOT grün; der Stand am Ende der Aufnahme (`[1 7]`) wäre die engere Grundlinie. Zeile M2 sagt Durchgang A zu und wird erfüllt; das Ziel in §1 („im Replay erreicht kein Request den Referenzdienst“) ist weiter. | ja — Mutant `upstart` | Fenster der Zähler-Messung kleiner als die Zusage |
| F-5 | LOW | `AGENTS.md` §3.15 (Gegenprobe des Suchers), Plan Zeile M3 | `tests/integration/m2_test.go` · `!slices.Contains(vorher, "odata-recorder.db")` | Die Gegenprobe des Suchers belegt Datenbank und Blobs, nicht `-wal` und `-shm`, obwohl der Plan „auch `-wal` und `-shm`“ zusagt. Im Lauf waren beide da (Debug-Kopie); bleibt die WAL-Datei aus, sucht der Test weiter grün und die Zusage entfällt still. | ja — Debug-Kopie mit `t.Logf` der Dateiliste (Ausgabe: db, `-shm`, `-wal`, 8 Blobs) | Gegenprobe des Suchers deckt nicht alle zugesagten Dateien |
| F-6 | MEDIUM | `AGENTS.md` §3.9 (Aussage berichtigen: alle Stellen), `SPEC-122` Nr. 2 | `harness/vertraege/SPEC-122-test-race.md` · „heute `go.mod`, `go.sum` und `src/`; ein künftiges `tests/`, sobald die Stufe es aufnimmt“ | Der Vertrag sagt „heute“ nur `src/`; mit diesem Slice nimmt die Stufe `source` `tests/` auf (`COPY tests ./tests`). Der Plan (§1) lässt den Wortlaut ausdrücklich stehen. Die Aussage ist nach dem Slice falsch, und die einzige Absicherung der Verdrahtung ist der Handversuch M10; ein Leser des Vertrags findet `tests/` nicht als geführt. | ja — Lesen von `SPEC-122-test-race.md` Nr. 2 gegen `Dockerfile` | Zustandsaussage „heute“ im Vertrag nach Änderung nicht nachgezogen |
| F-7 | LOW | `AGENTS.md` §3.9 | `docs/plan/planning/in-progress/slice-m2-referenzdienst.md` · Zeile M5 der Tabelle „Zusage · Fänger“, Spalte Mutation: „eigene `clock.NewZufallsIDs()` statt `h.ids`: je rot“ | Dieselbe Mutation steht im Absatz „Offen aus der Messung“ und in §6 als grün (nicht zu fangen). Die Vorgabe-Zeile trägt die ältere Aussage weiter. | ja — Lesen des Plans | Widerspruch zwischen zwei Stellen desselben Plans |
| F-8 | LOW | `AGENTS.md` §3.7 (Konjunktiv über die verworfene Alternative) | `src/bootstrap/import_workspace_test.go` · „ein Zugang, der sein Feld beim ersten Aufruf füllt, schriebe dort, und der Detektor meldet es“; `tests/integration/m2_test.go` · „er könnte den Referenzdienst erreichen, und der Zähler zeigt, dass er es nicht tut“ | Zwei Kommentare tragen eine Kopplung (warum der erste Aufruf in den Goroutinen liegt; warum der Replay-Lauf `UPSTREAM_URL` führt) und formulieren sie im Konjunktiv über die abwesende Alternative. Als LOW eingestuft, weil der Satz davor die indikative Kopplung trägt; liest man den Konjunktiv als „beschreibt die verworfene Alternative“, fiele er unter die HIGH-Liste des Skills — diese Lesart entscheidet nicht der Reviewer allein. | nein — kein Gate liest Kommentare | Kopplungs-Kommentar im Konjunktiv über die Alternative |
| F-9 | INFO | Maintainability, `AGENTS.md` §3.7 (Herkunft in ein auflösbares Feld) | `tests/integration/m2_test.go` · „(M2, M3)“ in `pruefeAufnahme`; `src/bootstrap/import_workspace_test.go` · „fängt M5 und M6“ | Die Zeilenkennungen M1 bis M10 stehen in Kommentaren ohne Slice-Kennung; „M2“ trägt außerdem den Namen des Meilensteins. Die Tabelle liegt im Plan, der nach `done/` und später ins Archiv wandert. Hinweis ohne erwartete Aktion in diesem Slice. | nein | Zeilenkennung im Code ohne auflösbaren Anker |
| F-10 | INFO | `AGENTS.md` §3.15, Plan §3 „Offen aus der Messung“ | `src/bootstrap/import_workspace_test.go` · `"IDs":            got.IDs != nil && got.IDs == bootstrap.IDsVon(hd)` (Import und Workspace) | Nicht doppelt gemeldet: die bekannte Mutation „eigene `clock.NewZufallsIDs()`“ bleibt grün (`impids` INT 0, BOOT 0 rot). Sie trifft auch die Zeile des Workspace-Dienstes (`wsids`, in einem Lauf mit gleicher Mutation an der Zeile des Sessions-Dienstes ebenfalls grün); die offene Rückgabe im Plan nennt Import und `TestWerteErreichenVerbraucher`, den Workspace-Dienst nicht ausdrücklich. Die Empfehlung (akzeptiertes Negativ mit Grund und Folge) trägt für alle drei Stellen gleich. | ja — Mutanten `impids`, `wsids` | Zustandsloser Generator, Einzelinstanz nicht fangbar |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/wurzel.go` (Verdrahtung Import und Workspace, Zugänge am `Handle`, `importDienste`) | geprüft, ohne Befund: Werte laut `SPEC-139` Nr. 3 und 4; `nil` ist der Nullwert, `als[…]` liefert bei Fehler die nil-Schnittstelle; Fehler beider Konstruktoren schließen die Ablage (3 Fälle) |
| `src/bootstrap/import_workspace_test.go` (M5, M6, M9, M9b) | geprüft, Befunde F-3, F-8, F-10; Klasse der 6 Regelfelder und der Felder von `Dienste`, `Blobs`, `Metadaten`, `MaxCaptureBody`, `Uhr` rot gesehen |
| `src/bootstrap/haken_test.go`, `export_test.go` | geprüft, ohne Befund (Haken zählen die Aufrufe, `IDsVon` nur Test-Naht) |
| `tests/integration/m2_test.go` (Orakel, Header-Vergleich, Zähler, Suche) | geprüft, Befunde F-1, F-4, F-5, F-8, F-9; Orakel unabhängig, Header-Menge aus der Konfiguration, Replay mit beendetem Referenzdienst |
| `tests/integration/referenzdienst_test.go` | geprüft, ohne Befund (Zähler, Selbstprobe `TestReferenzdienstSelbst`, absolute `nextLink`, `/Products(1)` gleich Seite 1, `$count` JSON laut Plan) |
| `Dockerfile`, `.dockerignore` (M10) | geprüft, ohne Befund: `COPY tests ./tests` in `source`, `!tests/` im Kontext; `test`, `race`, `lint` erben `source`; Rennen in `tests/` mit `pruefung race` rot |
| `.a-check.yml`, `harness/README.md`, `make a-check` | geprüft, Befund F-2; `make a-check` auf dem Worktree 0 Befunde |
| Hexagon-Richtungen | geprüft, ohne Befund: `wurzel.go` nennt Dienste und Adapter nur als Composition Root; `tests/` ruft nur `bootstrap` und nennt `model`, `ports/inbound`, `sqlite` |
| Größenmarke | geprüft, ohne Befund: 953 von 1700 (`git diff --numstat 1984383..HEAD`) |
| Rohbytes (§3.18) | geprüft, ohne Befund: gelesen wurden `tests/integration/*.go`, `import_workspace_test.go`, `wurzel.go`, der Plan |
| Gegenproben-Stempel | geprüft, ohne Befund: Eingabe-Hash gleich dem Stempel |
| Plan (Tabellen, Raten, Raster §6) | geprüft, Befunde F-6, F-7; Raster: jede Zelle trägt Ort oder „gilt nicht“ mit Grund; keine im Code entschiedene Zelle gefunden |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 4 |
| LOW | 4 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Geheimnis-Suche nur in einer Phase des Ablaufs · Ausnahmefläche im Architektur-Gate ohne Nennung ihrer vollen Wirkung · Feld der Konfig nur auf nicht-nil geprüft · Zustandsaussage „heute“ im Vertrag nach Änderung nicht nachgezogen · Fenster der Zähler-Messung kleiner als die Zusage · Kopplungs-Kommentar im Konjunktiv über die Alternative

## Verdikt

**Merge-blockierend:** ja — nach der Regel „MEDIUM blockiert typischerweise“; kein HIGH. Der Beweis des Durchstichs trägt: Orakel unabhängig, Zähler und Status/Body fangen je eigene Mutationen, M9b 30 von 30 mit Detektor-Meldung. Offen sind vier MEDIUM: F-1 und F-3 sind Fänger-Lücken (Implementer), F-2 und F-6 sind Entscheidungen über Gate und Vertrag (Architect bzw. Planner, `AGENTS.md` §3.13: als Zeile mit Ausgang *offen* und Adresse in §6 des Plans führen, bevor der Bericht übergeht).

**Übergabe:** Findings gehen an den Implementer (F-1, F-3, F-4, F-5, F-8) und an den Architect bzw. Planner (F-2, F-6, F-7, F-10); die Finding-Klassen gehen in die Slice-Closure §7. Dieser Report ersetzt keine Verifikation gegen die DoD.
