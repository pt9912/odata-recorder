# Review-Report: slice-audit-log (Teil A, Port und Adapter) — 2026-11-06

**Review-Art:** Code — gegen Plan, Vertrag (`SPEC-234`, `SPEC-235`, `SPEC-123` Nr. 6/8), Hard Rules und die ADRs ([ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)). Nicht gegen die DoD (Verifier).

**Gegenstand:** Zweig `slice-audit-log`, Diff `1067fcd..4a1c5e2` (Code a9cf341/25158e3, Plan 323d9ca, Architect 4a1c5e2); acht Dateien, davon Produktcode `src/hexagon/ports/support/audit.go` und `src/adapters/driven/audit/audit.go`.

**Skill:** `.harness/skills/reviewer.md` in der Fassung von `main` (mit den Zeilen „Menge ohne Achsenliste“, „Fänger schmaler als Zusage“, „unbeschränkte Menge ohne Prüfgrenze“, „Fixture gleicht dem Wert der Mutation“) · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-06

**Eingangs-Kontext:** `slice-audit-log` §3 (Tabelle A01–A21, Raten, Pflicht-Menge-Tabelle), §4 (Messung 1463 von 3000), §6 (Raster, Zusatztabelle); `SPEC-234`, `SPEC-235`; `LH-RB-04.a`, `LH-FA-13.a`, `LH-QA-04.a`; `AGENTS.md` §3.7, §3.10, §3.11, §3.15 bis §3.18. Teile E (`slice-audit-dienste`) und V (`slice-audit-verdrahtung`) sind nicht Gegenstand.

## Methode und Messbedingungen

Wegwerfkopien des Zweigs unter dem Scratchpad (`rev-al`), Perl-Ersetzung im Text, `go test -c` und Lauf des Testbinary im gepinnten Go-Image (`golang:1.27-alpine`, Digest `8a5910f31396…`), netzlos, `--memory=2g`, unter `flock /tmp/odata-recorder-heavy.lock`, Quelle schreibgeschützt, Kopien und Binaries danach gelöscht. Der Baum des Zweigs blieb unverändert. Grundlauf (unveränderter Code) in jedem Satz grün (BASE 0 von 3, BASEP 0 von 5).
Rate = Zahl roter Testbinary-Prozesse von N, je Prozess `-test.count=1`, die Prozesse nacheinander. Rechner mit 20 Kernen, `--cpus=4` (A18 zusätzlich `--cpus=0.3`); Load-Average (1 Minute) während der Läufe 7,6 bis 13,7, weitere Agenten liefen zugleich. Messender: dieser Review.
Statische Mutationen (S1 bis S5) liefen über das ganze Paket mit `-test.count=1`, N = 1 (deterministisch). Die Läufe der Sperre (W*, LOG, A18) führten nur die Tests `TestGleichzeitige*`, `TestSperre*`, `TestAnhaengen`, `TestZeilenende` und die Probe-Tests, nicht das ganze Paket.
`make a-check` im Worktree: 0 Befunde (der Hinweis auf 55 Dateien unter `tools/` ist vorbestehend und nicht Teil des Diffs). `make test-race` und `make gates` habe ich nicht gefahren; das SQLite-Flake ist mir nicht begegnet.

## Findings

### MEDIUM

**M1 — `Close` fehlt: Zusage „schließt sie wieder, hält kein Handle“ ohne Fänger**
- `kategorie`: MEDIUM · `quelle`: `AGENTS.md` §3.15, `SPEC-235` Nr. 2 · `klasse`: Zusage ohne Fänger (Ressourcen)
- `pfad`: `src/adapters/driven/audit/audit.go` · `defer func() { _ = f.Close() }()`
- `befund`: Mutation S2 (Zeile `defer … f.Close()` entfernt) bleibt in allen Tests grün. A21 nennt nur den *Fehler* des Schließens als Negativ, nicht dessen Fehlen. Szenario: die Aufbewahrung ruft `Append` je gelöschter Session; jeder Aufruf behält ein Handle, nach der Grenze der offenen Dateien scheitert `Append` mit `EMFILE`, die Einträge fehlen (das Log meldet es), und der Prozess verliert Handles für alles andere.
- `verifizierbar`: ja (Zahl der Einträge in `/proc/self/fd` vor und nach N `Append`).

**M2 — Log-Eintrag „nach der Freigabe der Sperre“ (`SPEC-235` Nr. 5) ohne Fänger, Fenster fehlt im Raster**
- `kategorie`: MEDIUM · `quelle`: `AGENTS.md` §3.10, §3.15 · `klasse`: Fenster der Sperre ohne Zeile
- `pfad`: `src/adapters/driven/audit/audit.go` · `d.melde(e, true)` in `Append` (nach `d.anhaengen`)
- `befund`: Mutation LOG (der Log-Aufruf läuft in einem `defer` von `anhaengen`, also unter der Sperre) bleibt für die Tests der Suite grün: in den fünf roten Läufen (N = 5) war allein mein Probe-Test `TestProbeLogNachFreigabe` rot (sein Handler ruft ein zweites `Append`), kein Test des Zweigs. Szenario: ein Handler, der blockiert (volle Pipe auf stderr), hält die Sperre und damit jeden `Append`, bei `Startup` vor den Listenern. Die Zusatztabelle in §6 führt Fenster „Sperre bis Uhr“, „Uhr bis Schreiben“, „Schreiben bis Sync“, „Freigabe auf jedem Pfad“; das Fenster „Schreiben bis Log“ steht nicht darin, A20 prüft nur, dass die Sperre frei wird, nicht wann.
- `verifizierbar`: ja.

**M3 — Fenster „Öffnen bis Schreiben“ hat nur A19 als Fänger, mit Rate unter 100 %, Rate nicht benannt**
- `kategorie`: MEDIUM · `quelle`: `AGENTS.md` §3.10 (Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt), §3.15 · `klasse`: Fenster der Sperre; Rate nicht benannt
- `pfad`: `docs/plan/planning/in-progress/slice-audit-log.md` · „Fenster der Sperre: von `Uhr.Now()` bis zum Schreiben“
- `befund`: Die Zeile ordnet das ganze Fenster A19 zu; die Raten im Plan (50 von 50) gelten für die drei Mutationen des Implementers, die die Sperre ganz oder vor dem Schreiben lösen. Löst eine Mutation die Sperre an einer Stelle *innerhalb* des Fensters (Freigabe und sofortiges Wiedernehmen, ohne `Gosched`), wird A19 nur teilweise rot, je 20 Prozesse: nach `Uhr.Now()`/vor der ID: 12 von 20; nach der ID, vor dem Öffnen: 17 von 20; nach dem Öffnen, vor der Zeilenende-Prüfung: 8 von 20; vor dem Schreiben: 10 von 20. Mit `runtime.Gosched()` im Fenster: 20/20, 20/20, 19/20, 20/20. Das Fenster zwischen Zeilenende-Prüfung und Schreiben hat zudem eine eigene Wirkung, die A19 nicht zählt (zwei gleichzeitige `Append` auf einer Datei mit halber Zeile schreiben dann zwei vorangestellte LF, eine Leerzeile); mein Probe-Test dafür (64 Goroutinen, Datei beginnt mit `halb`) wurde mit Yield 20 von 20 rot (zugleich mit A19), er gehört nicht zur Suite. Ein Wirkungs-Fänger mit halber Zeile fehlt; die Rate je Mutation fehlt im Plan.
- `verifizierbar`: ja.

**M4 — Menge „Zeilenende“: die Achse Dateilänge hat keinen Fall für ein Byte ohne LF**
- `kategorie`: MEDIUM · `quelle`: `AGENTS.md` §3.15, §3.17 (Fänger führt nur eine Form der Eingabe) · `klasse`: Menge ohne Achsenliste, Element nicht genannt
- `pfad`: `src/adapters/driven/audit/audit.go` · `if info.Size() == 0 {`; Test `TestZeilenende`
- `befund`: Die Menge hat sieben Zustände, die kürzeste Datei ohne LF am Ende ist `x\r` (2 Bytes). Mutation S1 (`Size() <= 1`) bleibt grün: eine Datei aus genau einem Byte ohne LF (Absturz nach dem ersten Byte) bekommt kein vorangestelltes LF, der nächste Eintrag klebt an dem Byte und die Zeile ist kein JSON. Die Pflicht-Menge-Tabelle nennt die Achsen der Menge nicht (Zustand 7, ohne Achse Länge).
- `verifizierbar`: ja.

**M5 — Menge „Kennungsform“: die Achse Stelle ist nur bei Länge 3 geführt**
- `kategorie`: MEDIUM · `quelle`: `AGENTS.md` §3.15, §3.17 · `klasse`: Menge ohne Achsenliste, Element nicht genannt
- `pfad`: `src/adapters/driven/audit/audit.go` · `for i := range len(s) {`; Test `TestKennungsform`
- `befund`: Das Plan-Element „Stelle 3 (erstes, mittleres, letztes)“ gilt für Werte der Länge 3 (`%sxx`, `x%sx`, `xx%s`). Bei langen Werten prüft der Test nur den letzten Platz (`a`×64 plus LF) und das Mehrbyte-UTF-8 am Ende. Mutation S4 (die Prüfung überspringt für `len(s) >= 32` die Plätze 4 bis `len-5`) bleibt grün, sodass ein ungültiges Byte mitten in einer 64-Byte-Kennung durchgeht. Die frühere Fassung dieser Mutation ohne Längenbedingung wurde nur zufällig rot (`TestLogBeiUngueltig`, Wert `canary svc` hat sein Leerzeichen auf Platz 7), nicht durch den Fänger der Menge. Szenario: eine Kennung mit einem Anführungszeichen in der Mitte bräche das JSON der Zeile (die Zeile hat keine Maskierung).
- `verifizierbar`: ja.

**M6 — Menge „Verformungen der Art“: die Achse Zeichen nennt fünf Leerraumzeichen, nicht NUL, FF und NBSP**
- `kategorie`: MEDIUM · `quelle`: `AGENTS.md` §3.15, §3.17 · `klasse`: Menge ohne Achsenliste, Element nicht genannt
- `pfad`: `src/adapters/driven/audit/audit_test.go` · `for _, w := range []string{" ", "\t", "\n", "\r", "\v"} {`
- `befund`: Mutation S5 (`strings.Trim(string(e.Art), "<NUL><FF><NBSP>")` vor dem Vergleich der Art, NBSP als Bytefolge C2 A0) bleibt grün. Die Achse „Zeichen vor/hinter der Art“ führt fünf Elemente (Folge der Elemente: ASCII-Leerraum); NUL (Byte 0), Seitenvorschub und ein Mehrbyte-Leerraum fehlen. Szenario: ein Aufrufer mit `session_geloescht` plus NUL würde als Art akzeptiert und mit dem NUL in der Zeile geschrieben.
- `verifizierbar`: ja.

### LOW

**L1 — Kommentar an `AuditArt` ist veraltet (Abgrenzung falsch)**
- `kategorie`: LOW · `quelle`: `AGENTS.md` §3.7 · `klasse`: Kommentar nennt abwesenden Text
- `pfad`: `src/hexagon/ports/support/audit.go` · „Der Name weicht vom Vertrag ab, weil Art den Typ der Attribute von Telemetry trägt.“
- `befund`: Der Vertrag nennt den Typ seit 4a1c5e2 `AuditArt`; die Aussage „weicht ab“ stimmt nicht mehr. Der Architect hat als Auflage an den Implementer festgelegt (§6, Zeile *Namenskollision*), dass der Kommentar den Grund nennt. Der zweite Halbsatz (Kollision mit dem Typ `Art` der Attribute, `SPEC-230`) ist die gültige Abgrenzung. Als Kategorie nehme ich LOW nach Vorgabe der Anfrage; die HIGH-Zeile des Skills („Kommentar beschreibt … einen abwesenden Text“) trifft nur dann, wenn man „weicht ab“ als Chronik liest, was ich nicht tue.
- `verifizierbar`: nein.

**L2 — Negativ A21 nennt den Fehler von `Sync`, nicht das Fehlen des Aufrufs**
- `kategorie`: LOW · `quelle`: `AGENTS.md` §3.15 · `klasse`: akzeptiertes Negativ schmaler als Zusage
- `pfad`: `docs/plan/planning/in-progress/slice-audit-log.md` · „ein Fehler von `Sync`, ein kurzes Schreiben und ein Fehler des Schließens sind nicht auslösbar“
- `befund`: Mutation S3 (`return f.Sync()` zu `return nil`) bleibt grün. Die Zusage „erst nach `Sync` `nil`“ (`SPEC-235` Nr. 4 h) hat Grund und Folge im Negativ, aber nur für den Fehler des Aufrufs; der Wegfall des Aufrufs steht nicht darin. Die Folge-Zeile des Vertrags („hängt am Dateisystem“) trägt ihn teilweise. Eine Naht gibt es nicht; das ist als Negativ vertretbar, wenn es den Wegfall nennt.
- `verifizierbar`: nein (kein Lauf des Images kann es sehen).

### INFO

- **I1 — A18 (300 ms)**: die Grenze ist ein Warten auf ein Ereignis, das nicht kommen soll; der Grundlauf kann deshalb nicht flackern (rot nur bei einem zweiten `Uhr.Now()`). Das Durchlassen der Mutation bräuchte einen zweiten Aufruf, der später als 300 ms kommt. Gemessen: Mutation a (`Uhr.Now()` vor der Sperre) und b (Sperre nur um `schreibe`) je 20 von 20 bei `--cpus=4` und 20 von 20 bei `--cpus=0.3`. Grenze angemessen; die Plan-Aussage („nicht gemessen unter Last“) ist mit dieser Messung ergänzbar.
- **I2 — A19/A20**: die Raten des Plans (50/50, 20/20) für die genannten Mutationen halten (siehe M3 für Mutationen innerhalb des Fensters). A20 prüft die Freigabe auf den vier Pfaden (Form liegt vor der Sperre).
- **I3 — Datei mit weiten Rechten / `O_RDWR`**: der Adapter öffnet mit `O_RDWR`; eine vorhandene Datei ohne Leserecht (zum Beispiel `0200`) lässt `Append` scheitern, obwohl Anhängen erlaubt wäre. `SPEC-235` *Ungeprüft* nennt Rechte vorhandener Dateien; keine Aktion erwartet.
- **I4 — `Datei{}` ohne `New`**: ein Nullwert hat `Uhr == nil`, `Append` gerät dann in Panik; der Vertrag lässt nur `New` zu.

## Aufgelöste Zeilen und Zellen (ohne Befund)

- A01 bis A17, A19 (genannte Mutationen), A20, Rechte (`0600`, `O_APPEND`), abgebrochene letzte Zeile (außer M4), Fehler-Log (Soll-Objekt exakt, Canary in Pfad, Fehlertext, ungültiger Eintrag): vom Implementer mit Mutationen belegt (Tabelle in §3 des Plans); meine zusätzlichen Mutationen blieben ungefangen nur an den Stellen der Findings. Menge Zellen der Feldtabelle (4 × 5 × 2): Soll-Orakel getrennt von `Arten()`; Menge Code (−5 bis 60): Orakel getrennt vom Modell.
- Geheimnisse: Zeile und Log tragen nur die Felder des Eintrags; der Fehlertext von `pruefe` nennt Feldnamen, nie Werte; im Log steht beim Fehler der Prüfung weder `grund` noch eine Kennung (`TestLogBeiUngueltig`). Ein Tokenwert hat in keinem Feld Platz (Kennungsform), siehe `SPEC-234` Nr. 3.
- Hard Rules: keine `//nolint` in den acht Dateien; Rohbytes gegengelesen (`LC_ALL=C grep -P` auf Backslash-`u`-Folgen in `src/adapters/driven/audit/*.go` und `src/hexagon/ports/support/audit*.go`: keine Treffer; die Nicht-ASCII-Zeichen sind Umlaute in Kommentaren und Schlüsseln). [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md): keine Änderung unter `src/adapters/driven/sqlite`, `tools/schema` und keine Migration. Gate-Kette (`Makefile`, `harness/`, `Dockerfile`, `.a-check.yml`, `tools/`) unberührt. `make a-check`: 0 Befunde ([ADR-0002](../plan/adr/0002-architektur-gate-a-check.md)).

## Negativbefund-Zeilen (geprüft, ohne Befund)

- `src/hexagon/ports/support/` (`audit.go`, `audit_test.go`): geprüft, ohne Befund außer L1.
- `src/adapters/driven/audit/` (`audit.go`, Tests): geprüft; Befunde M1 bis M6, L2.
- `spec/spezifikation.md` (Diff: `AuditArt`): geprüft, ohne Befund; Typzeile und Historie stimmen mit dem Code überein.
- `docs/plan/planning/in-progress/slice-audit-log.md`: geprüft; Befunde M3 (Zusatztabelle), L2; Messung 1463 von 3000 im Rahmen der Marke.
- Gate-Kette (`Makefile`, `harness/`, `Dockerfile`, `tools/`, `.a-check.yml`): geprüft, ohne Befund (unberührt).

## Nicht aufgelöst

- `make test-race` und `make gates`: nicht gefahren (Verifier).
- Eine Sperre im Gebrauch durch den Dienst (Prozessende zwischen Commit und `Append`, A21-Teil): liegt bei `slice-audit-dienste` (E19), nicht Gegenstand.
- Die Mutationen S1 bis S5, W*, LOG liefen mit N = 1 (statisch) bzw. N = 20 (Sperre); LOG mit N = 5. Probe-Tests und Mutationstexte liegen nur im Scratchpad (`rev-al/zus*`), nicht im Repo.
- Mengen mit Elementen, die ich nicht mutiert habe: Zeit (A04/A10, Elemente aus der Tabelle des Implementers gelesen, nicht eigenständig mutiert), Codes (A09), die Schlüsselmenge des Log-Eintrags je Art.
