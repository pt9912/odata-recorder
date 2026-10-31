# Review-Report: slice-cli-adapter — 2026-10-31

**Review-Art:** Code — gegen Plan, `SPEC-138`, Hard Rules (`AGENTS.md` §3.7, §3.9 bis §3.18) und die Regeln von `a-check`.

**Gegenstand:** Diff `ba5caa1..63f65fb` (Zweig slice-cli-adapter): Paket `src/adapters/driving/cli`, `src/bootstrap/main`, `src/bootstrap/bootstrap.go`, gelöschtes `bootstrap_test.go`, Plan `slice-cli-adapter`.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Eingangs-Kontext:**

- Slice-Plan `slice-cli-adapter` (Tabelle *Zusage · Fänger* C1 bis C14, Raster und Zusatztabelle in §6, Marke 2500, Messung 1182)
- `SPEC-138`, `SPEC-110` (Zusätze), `ARC-006`, `ADR-0001`, `ADR-0007`
- `LH-FA-02` (Buchstabe c), `LH-RB-03`
- `AGENTS.md` (Hard Rules)

**Messweise.** Mutanten in Kopien im Scratchpad (nicht im Worktree), Image `odata-recorder:race`, `--memory=4g --cpus=4 --network none`, `frist` der Tests auf 200 ms gesetzt (wie vom Implementer, damit hängende Mutanten schnell rot werden). Der Reviewer fuhr rund 80 Mutanten (je Tabellenzeile mehrere, dazu Elemente, die der Plan nicht nennt) einmal; die nebenläufigen Zeilen C7 bis C12 danach mit `go test -count=20` je Mutante in **einem** Prozess und zählte die Läufe des roten Tests. Das Datenrennen von C9 lief als `-race`-Binary in **30 Prozessen** mit `-test.count=1`. Das ungemutierte Paket: grün in 20 von 20 Läufen (`-count=20`) und in 30 von 30 Prozessen unter `-race`, ohne Bericht `DATA RACE`. Nicht gefahren: `make gates` (Frage des Verifiers).

## Auflösung der Tabelle (eine Mutation je Zeile, Menge als Klasse)

| Zeile | Mutationen des Reviewers (rot) | Rate | Klasse der Menge |
|---|---|---|---|
| C1 | ohne Argument ruft `Start`; `start` unbekannt; `start` bei `nil` ruft `Start`; `version` ruft `Start` | 4 von 4 | geschlossen (8) |
| C2 | `EqualFold`; `TrimSpace`; `HasPrefix`; `HasSuffix`; nur `args[0]`; letztes Argument in der Meldung; stdout; Exit-Code 1; ohne Zeilenende; `version` mit Zusatz angenommen; **nicht im Plan:** NUL entfernt, `\r` abgeschnitten, `Start` zusätzlich angenommen | 13 von 13 | Klasse 1558 hält (NUL, `\r`, Suffix an Elementen, die der Plan nicht nannte, rot) |
| C3 | `TrimSpace`; leer zu `dev`; stderr; ohne Zeilenende; ohne Präfix | 5 von 5 | **Lücke, siehe R-1** |
| C4 | nil-Writer nicht ersetzt (stdout, stderr); nil an `Start`; Schreibfehler ergibt 1 (Version, Verstoß) | 5 von 5 | geschlossen (18) |
| C5 | zweimal gerufen; Writer vertauscht; `context.Background()`; Kontext schon abgebrochen; Adapter schreibt | 5 von 5 | 4 Eigenschaften |
| C6 | Durchreichen; `<= 0`; `& 0xFF`; `!= 1`; **nicht im Plan:** `int32`, `int16`, `% 256`, Sonderfall 128, Sonderfall 2^62 | 9 von 9 | Klasse 2365 hält |
| C7 | Anmeldung bei Verstoß, bei `version`, bei `start`+`nil`; doppelt; nach dem Aufruf von `Start` | 5 Mutanten je 20 von 20 | geschlossen (5) |
| C8 | Signal ignoriert; Kontext von Anfang an abgebrochen; Erzwingen beim ersten Signal; SIGINT, SIGTERM nicht angemeldet; erstes SIGHUP ignoriert (zwei Formen); Abbruch fehlt in der Schleife | 8 Mutanten je 20 von 20 | geschlossen (8) |
| C9 | `Run` kehrt nach dem Abbruch zurück (Code 1 und Code 0); Ergebnis über gemeinsame Variable | 2 Mutanten je 20 von 20; Datenrennen 30 von 30 Prozessen mit `DATA RACE` (ungemutiert 0 von 30) | geschlossen (6) |
| C10 | Erzwingen fehlt; Code des ersten Signals; immer 143; immer 130; erst beim dritten; Ausgabe beim Erzwingen; Warten auf `Start` nach dem Erzwingen; Zuordnung SIGHUP zu 130 | 8 Mutanten je 20 von 20 | geschlossen (10) |
| C11 | Abmeldung fehlt; doppelt (zwei Formen); vor der Rückkehr (zwei Formen); nur auf dem Normalpfad | 6 Mutanten je 20 von 20 | Fake-Quelle; echte Quelle **siehe R-3** |
| C12 | Paketzustand „erzwungen“; Paketkontext; Paketmerker „erstes Signal“ | 3 Mutanten je 20 von 20 | geschlossen (4) |
| C13 | Import von `src/bootstrap`; von `src/hexagon/model` (jeweils lauffähig, Test rot, nicht Aufbau) | 2 von 2 | alle Importe |
| C14 | Mutation ohne Vorabprüfung des vorbelegten Signals: grün | 0 von 20; 0 von 50 je `-cpu 1, 2, 4, 20` | siehe R-4 |

Bestätigt wurde damit: Die Fänger von C7 bis C12 sind geordnet (Kanäle, blockierte Fakes) und werden in jedem Lauf rot; das Datenrennen von C9 fängt der Detektor in jedem Prozess.

## Findings

| ID | kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|---|
| R-1 | MEDIUM | `AGENTS.md` §3.15 (Fänger führt jede Form der Eingabe), `SPEC-138` Nr. 3 | `src/adapters/driving/cli/kommando_test.go` · `versionen := []string{"1.2.3", "dev", "", " x ", "a\tb"}` | Nr. 3 sagt „`Version` unverändert (auch leer oder mit Leerraum)“; der Fänger führt Randleerzeichen und Tab, aber kein Zeilenende, kein NUL, keine ungültige UTF-8-Folge. Drei Mutationen bleiben grün: `strings.TrimRight(k.Version, "\n\r")`, NUL entfernt, `strings.ToValidUTF8(k.Version, "")` (je 1 von 1 Lauf grün). Eine Version mit Zeilenende käme verändert auf stdout. | ja (Mutation im Scratchpad) | Fänger führt nur eine Form der Eingabe |
| R-2 | MEDIUM | `AGENTS.md` §3.15 (Zusage ohne Fänger), `SPEC-138` Nr. 1 | `src/adapters/driving/cli/cli.go` · `const signalKanal = 2` | Nr. 1 sagt zu: Anmeldung bei der Laufzeit mit „Kanalgröße 2“. Die Tabelle hat dafür keine Zeile und kein benanntes Negativ; die Mutation `make(chan os.Signal)` (ohne Puffer, `signal.Notify` verwirft dann Signale, die eintreffen, während die Schleife nicht empfängt) bleibt grün (1 von 1 Lauf). Die echten Signale sind nur „im Lauf“ geprüft, nie vorbelegt (C8 vorbelegt gilt nur für die Fake-Quelle). | ja | Zusage ohne Fänger |
| R-3 | MEDIUM | `AGENTS.md` §3.15, `SPEC-138` Nr. 1 und 7 d | `src/adapters/driving/cli/cli.go` · `return kanal, func() { signal.Stop(kanal) }` | Die Abmeldung der Laufzeit-Quelle (`signal.Stop`) hat keinen Fänger: Die Mutation `_ = kanal` statt `signal.Stop(kanal)` bleibt grün (1 von 1). C11 zählt nur die Abmeldung der `FakeSignale`. Das Negativ steht weder in C14 noch in *Ungeprüft* von `SPEC-138`; Grund und Folge fehlen damit. | ja | Zusage ohne Fänger |
| R-4 | MEDIUM | `AGENTS.md` §3.15 (akzeptiertes Negativ nennt Grund und Folge), `SPEC-138` Nr. 7 a und 7 b | `docs/plan/planning/in-progress/slice-cli-adapter.md` · „ein Fänger kann die Vorabprüfung nicht von einem Abbruch durch die Schleife trennen“ | Die Zusage „ein vor dem Aufruf eingetroffenes Signal bricht den Kontext, den `Start` erhält“ und „der Abbruch geschieht genau einmal“ haben keinen Fänger: Die Mutation ohne Vorabprüfung ist grün in 0 von 20 Läufen und 0 von 50 je `-cpu 1, 2, 4, 20`; das Reihen der Goroutine macht den Abbruch früher als ihren Beginn, aber ohne Garantie. Der Plan nennt den Grund, nicht die **Folge im Betrieb**, und beide Punkte stehen in §6 mit Ausgang *offen*, Adresse Architect (korrekt adressiert, `AGENTS.md` §3.13); `SPEC-138` führt sie noch nicht in *Ungeprüft*. | ja | Akzeptiertes Negativ ohne Folge |
| R-5 | MEDIUM | `AGENTS.md` §3.16, `SPEC-138` Nr. 1 | `docs/plan/planning/in-progress/slice-cli-adapter.md` · Raster, Zeile „Signalquelle (`Signale`, `nil` = System)“ | Das Raster hat für die Signalquelle keine Zelle für eine Quelle, die einen geschlossenen Kanal oder eine Abmeldung gleich `nil` liefert. Aus dem Code gelesen (nicht gefahren): Ein geschlossener Kanal liefert sofort `nil`, der Abbruch folgt, und die Schleife liefert beim nächsten Durchlauf 143, ohne dass ein Signal kam; eine `nil`-Abmeldung bricht den `defer` mit einer Panik. Der Vertrag nennt beides weder als Verstoß noch als Ungeprüftes. | nein (nicht gemutet, nur gelesen) | Randform ohne Ort |
| R-6 | INFO | `AGENTS.md` §3.10 | `docs/plan/planning/in-progress/slice-cli-adapter.md` · „Die Fänger von C7 bis C12 liefen mit `-count=50` in einem Prozess“ | Die Raten des Implementers gelten für `frist` 100 ms und einen Prozess; sie sind mit Zahl genannt und stimmen mit der Nachmessung des Reviewers (200 ms, 20 Läufe, ein Prozess; Datenrennen 30 Prozesse) überein. Die gelieferte Suite fährt `frist` 5 s; die Rate dort ist nicht gemessen und nur durch die kürzere Frist ersetzt, nicht widerlegt. | ja | Rate unter anderer Frist als geliefert |
| R-7 | INFO | `SPEC-138` *Ungeprüft*, Handoff | `src/bootstrap/main/main.go` · `cli.Konfig{Version: version, Stdout: os.Stdout, Stderr: os.Stderr}` | `Start` und `Signale` sind in `main` nicht gesetzt; `start` endet dort mit Exit-Code 2 und `Unbekanntes Kommando: start`. Das entspricht Plan §1 und dem benannten Negativ (kein Test mit echtem SIGTERM am Binary; `main` ist eine Zeile). Verifier: die Zusage „Binary startbar“ (`LH-RB-03.a`) ist für `version` belegt, für `start` bis `slice-bootstrap-m2` nicht. | nein | Verdrahtung am Binary ungefangen |

Kein HIGH. Kein Verstoß gegen `a-check` erkennbar: Das Paket importiert nur die Standardbibliothek (C13, Mutation rot); der Lauf von `make a-check` ist Sache des Verifiers.

## Weitere Prüfungen (ohne Befund)

- Größenmarke: Summe aus Einfügungen und Löschungen über die vier Pfade des Plans ist 1182 (cli 1090, `bootstrap.go` 39, `bootstrap_test.go` 49, `main.go` 4), Marke 2500 gehalten.
- `AGENTS.md` §3.9: Plan-Kopf, §1, §3 (Tabelle, Gemessen), §4 (Messung), §6 sind im Commit des Implementers nachgezogen; Folge-Slices als Adressen in §6 mit Ausgang *offen*.
- `AGENTS.md` §3.18: kein Backslash-u in Quellen und Plan; NBSP und BOM der Formen entstehen aus Bytewerten (`kommando_test.go`, Aufbau `string([]byte{...})`), Datei auf Rohbytes gegengelesen (`LC_ALL=C grep -P` über `src/adapters/driving/cli/*.go` und den Plan: nur das Zeichen × aus Kommentaren).
- `AGENTS.md` §3.7 und §3.11: Kommentare in `cli.go` und den Tests beschreiben den Zustand; kein Kommentar sagt mehr zu, als ein Test prüft (Ausnahme R-2/R-3: der Kommentar zu `signalKanal` und die Abmeldung der Laufzeit stehen im Vertrag, nicht im Fänger).

## Negativbefund (geprüft, ohne Befund)

- `src/adapters/driving/cli/cli.go` · geprüft, kein Befund außer R-2, R-3, R-5.
- `src/adapters/driving/cli/fakes_test.go` · geprüft, kein Befund (Fakes halten jedes Argument und zählen jede Operation, auch Rückkehr und Abmeldung).
- `src/adapters/driving/cli/kommando_test.go` · geprüft, Befund R-1 (Rest ohne Befund).
- `src/adapters/driving/cli/signale_test.go` · geprüft, kein Befund (nur Wartepunkte als `select` mit Zeitgrenze; C7 bis C12 in jedem Lauf rot).
- `src/bootstrap/` (`bootstrap.go`, `main/main.go`, gelöschtes `bootstrap_test.go`) · geprüft, kein Befund; `exitOK` bleibt für `konfig.go`, die Gegenproben hängen weiter an `bootstrap.go`.
- `spec/spezifikation.md` `SPEC-138` und `harness/vertraege/SPEC-036-test.md` · gegen Code gelesen, kein Widerspruch zum Lastenheft erkennbar.
- `docs/plan/planning/in-progress/slice-cli-adapter.md` · geprüft, Befunde R-4, R-6.
