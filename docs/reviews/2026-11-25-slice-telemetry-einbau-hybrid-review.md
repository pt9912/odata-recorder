# Review-Report: slice-telemetry-einbau-hybrid (Teil 2) — 2026-11-25

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-telemetry-einbau-hybrid`, HEAD `0c7ef2c8` (Commits: Handoff `e1925257`, Merge `41f26d86`, Nachzug von §6 `d5b936a8`, errcheck-Fix `0c7ef2c8`), Messbasis `f7371c39`; Teil 2 des Schnitts: Dienst und Verdrahtung des Hybrid-Fallbacks, Zeilen D1 bis D18: Span `odata.hybrid.fallback` und Zähler `odr_hybrid_discarded_total` über `support.FallbackTelemetry` in `src/hexagon/services/hybrid/` (`span.go`, `verteiler.go`, `upstream.go`), `Konfig.Telemetrie` als Pflichtfeld, `verdrahteHybrid` in `src/bootstrap/wurzel.go`. Teil 1 (Port und Adapter) ist nicht Gegenstand.

**Skill:** `.harness/skills/reviewer.md` @ HEAD des Zweigs · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-25 (Läufe bis 2026-11-26)

**Eingangs-Kontext:**

- Plan `slice-telemetry-einbau-hybrid`: Kopf, §1, §3 (Tabelle D1 bis D18, Mengen mit Zählprüfung, Mutationen und Raten des Implementers, Negative (1) bis (16)), §4 (Block mit vier Pfaden, Messpunkte), §6 (Risiken, Raster, die fünf Abweichungen (a) bis (e)), §8.
- `SPEC-320`, `SPEC-321` (Nr. 1 bis 9, *Ungeprüft*), `SPEC-230` Nr. 1, 5, 8, 11, `SPEC-232` Nr. 5, `SPEC-284` Nr. 1 bis 13, `SPEC-261` Nr. 3 bis 5, `SPEC-296` Nr. 5, `SPEC-139` Nr. 13 g, `SPEC-109` Nr. 2.
- `AGENTS.md` §3.6 bis §3.18; `harness/README.md` §Sensors; `LH-FA-12`, `LH-QA-04`, `LH-QA-05`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Bezug des Plans, nur genannt).
- Vorherige Reviews am Modul: `slice-telemetry-einbau-hybrid-port`, `-replay`, `-gateway`, `-record`, `-mock` (Lehren: Menge aus dem Wortlaut der Spezifikation, Allaussage als Schleife, Zählprüfung gegen die Quelle, jede Aufrufstelle ein Fall, Kommentar nennt nur, was ein Test fängt).
- Entscheidung des Architect zu den Abweichungen (a) bis (e) (Zweig `spec-hybrid-nachzug`, Commit `c3b539be`, nur `spec/spezifikation.md`, bei diesem Review noch nicht auf `main`; vom Aufrufer mitgeteilt, nicht selbst gelesen): (a) 16 ist richtig; (b) `SPEC-321` Nr. 8 nennt die Ausnahme des Ziels mit Benutzerteil (1/1/0/0); (c) Canary mit acht Zeichen ist neuer Punkt in *Ungeprüft*; (d) 9 von 14 Stellen am Root, Maß 5 von 14; (e) D4: sechs Mutationen für `Set`. Der Slice ist gegen diese Entscheidung beurteilt, eine Zeile von §6 (a) bis (e) ist kein offener Befund des Implementers.

**Unverändert (geprüft mit `git diff --name-only f7371c39..e1925257`):** außerhalb der vier Pfade des Blocks ändert der Slice nur Plan-Dokumente (`slice-telemetry-einbau-hybrid.md`, `roadmap.md`, `welle-betrieb.md`, die zwei letzten im Start-Commit des Planners). `Makefile`, `harness/`, `Dockerfile`, `.golangci.yml`, `.a-check.yml`, `spec/`, `tools/`, `src/hexagon/ports/`, `src/adapters/` und `tests/` bleiben unberührt (der Merge `41f26d86` bringt die Änderungen von `main` mit). Im Diff der Pfade des Blocks: kein `//nolint`, kein `t.Skip`, kein TODO; kein Backslash-`u`, kein BOM, kein NBSP, kein Nullbreiten- oder Zeilentrennzeichen, kein Steuerzeichen (`LC_ALL=C grep -P` und `-F` auf dem Diff `f7371c39..HEAD` über die vier Pfade und den Plan). Die Datei `telemetrie_hybrid_test.go` enthält keinen Wert mit Escape-Folge; der Canary `kanari7q` ist ein ASCII-Literal.

**Größenmarke:** `git diff --numstat f7371c39..HEAD` über die vier Pfade des Blocks: 1707 Einfügungen, 83 Löschungen, Summe **1790** gegen die Marke 2400, wie berichtet; Produktcode von Hand 163 (`span.go` 50, `verteiler.go` 68, `upstream.go` 38, `wurzel.go` 7); fremde Zeilen 0 (`git diff --numstat f7371c39..073042bc` über dieselben Pfade ist leer). `make groessenmarke`: `slice-telemetry-einbau-hybrid.md: ok, Messung 1790 von Marke 2400`, Exit 0.

**Sensoren dieses Laufs** (alle in einer Sitzung unter `flock /tmp/odata-recorder-heavy.lock`, nacheinander, Last des Hosts 1-Minuten-Mittel 1,95 bis 5,45; Messender: der Reviewer, Modell claude-sonnet-5-5):

| Sensor | Ergebnis |
|---|---|
| `make test` mit `DOCKER_BUILD="docker build --progress=plain --no-cache-filter test"` (Stufe `test` frisch, leerer Go-Cache der Stufe, Container ohne Netz) | `pruefung test: ok`, 348 s in der Stufe, 364 s gesamt, Exit 0; keiner der bekannten Flakes (`TestHandbuchZeitgrenzeUndWarten`, `TestAnfrageBodyBytes`) trat auf |
| `make test-race` mit `--no-cache-filter race` (Stufe frisch) | `pruefung race: ok`, 384 s in der Stufe, 401 s gesamt, Exit 0; keine Flake |
| `make a-check` | `gesamt: 0 Befund(e)`; Hinweis des Werkzeugs: 97 Dateien unter `tools/` liegen in keiner Schicht (Bestand, nicht dieser Diff); aus dem Build-Cache (Eingabe unverändert) |
| `make lint` mit `--no-cache-filter lint` (Stufe frisch) | `0 issues`, Exit 0, 46 s |
| `make katalog-check` | Exit 0 (Stufe `katalog` aus dem Cache; der Diff macht keinen Fehlercode, Schlüssel oder Kommando erreichbar, `harness/katalog-ausnahmen.txt` unverändert) |
| `make groessenmarke` | siehe oben, Exit 0 |
| `make docs-check`, `make zeichen-check` | nach dem Schreiben dieses Reports (Ergebnis im Bericht an den Aufrufer) |

**Abdeckung** (`-race -coverprofile`, je Paket, ein Prozess): `hybrid` **100,0 %**, kein Block mit Zähler 0 (auch nicht in `span.go`, `verteiler.go`, `upstream.go`). `bootstrap` 97,5 %: im neuen Code von `wurzel.go` (das Feld im Aufruf von `NeuerHybrid`) kein Nullblock; in `verdrahteHybrid` die vier Altblöcke 349, 353, 357, 371 wie im Plan; weitere Nullblöcke des Pakets (336, 407, 418, 425, 452) liegen in Bestandscode außerhalb des Diffs.

**Nicht nachgemessen:** `make gates` als Kette, `make test-integration`, `make test-determinism`, `make test-contract`, `make handbuch-check`, `make gegenproben` (nicht gefahren, Auftrag; Stempel nicht angefasst); `make katalog-check` und `make a-check` nicht mit leerem Cache; die Raten des Implementers (20 Prozesse je Zelle) nur für die eigenen Mutationen unten, nicht für die zwölf Zeilen seiner Tabelle; **keine eigene Mutation zu D14** (jede Mutation, die Zustand zwischen zwei Anfragen hält, braucht ein Feld oder eine Paketvariable und ist durch `TestVerteilerFelder` und `TestQuelltext` strukturell abgewiesen, siehe F-8); die Zahl 137 der Mutationen des Implementers (siehe F-11); `SPEC-321` Nr. 6 bis 8 und die Entscheidung des Architect wurden am Wortlaut von `main` nachgezählt, der Zweig `spec-hybrid-nachzug` nicht gelesen.

---

## Eigene Mutationen

Treiber im Scratchpad (`hybrid-reviewer/`, außerhalb des Repos, nie im Worktree): ein Container `odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, `CGO_ENABLED=0`, Wegwerfkopie von `go.mod`, `go.sum`, `src` und `tests`; je Mutation ein Text in einer Datei, danach `go test -count=1` über das **ganze** Paket (damit auch fremde Tests fangen), bei `d12-*` über `bootstrap`. Die Nullprobe (Baum ohne Änderung, beide Pakete) ist grün. Jede Mutation liegt an einem Element, das weder Plan noch Handoff nannten, oder an einer anderen Stelle als dort. Eine Mutation zählt nur, wenn die Meldung des Tests die Zusage nennt (Folge, Zählung, Ergebnis, Buch, Canary, Quelltext), nicht der Aufbau.

**48 Läufe, davon zwei Wiederholungen derselben Mutation im ganzen Paket (46 verschiedene Mutationen): 45 rot aus dem benannten Grund (eine davon durch `panic:` statt `--- FAIL` des benannten Tests, siehe F-9; eine im benannten Test nur nach Rate, im Paket deterministisch, F-8), 1 grün (F-1).**

| Zeile | Mutation (eigene) | roter Test (erster Grund) |
|---|---|---|
| D1 | nil `Telemetrie` liefert `context.Canceled` statt `model.ErrUngueltig`; `New` übernimmt das Feld nicht | `TestNew` (Fehler `context canceled`); die zweite rot durch `panic:` im ersten Test mit Fallback (`TestCodesSchleife`), nicht aus `TestNew` (F-9) |
| D2 | Zählung bei nicht fallback-fähiger Anfrage (Body über `MaxBytes`); Span bei Erfolg von M mit Status 404; Span bei Kontextfehler von M | `TestOhneSpan` und `TestQuelltextTelemetrie` (`Body über MaxBytes, Laenge: Aufruf "Count(ODR-E6003)"`; `Erfolg 404: Aufruf "Start(replay)"`; `Fehler ohne Code 2: Aufruf "Start(upstream)"`) |
| D3 | Reihenfolge Beginn vor Zählung nur bei `replay`; `ctx.Err()` nach statt vor der Stufe; Zählung nach der Stufe | `TestWege`, `TestAbbruch` (`Abbruch in M: Replay 1-mal gerufen, erwartet 0`; Folge R1 bis R8), `TestQuelltextTelemetrie` (Operation zweimal) |
| D4 | `End` nicht bei `abgebrochen`; `End` zweimal bei `abgebrochen`; Body nicht geschlossen nur bei Fehler ohne Code (Replay); Body nicht geschlossen bei Kontextfehler von `Senden` | `TestWege`, `TestAbbruch`, `TestGleichzeitigeSpans` (`Span 7 hat 0 End`), `TestStromSchliessen`, `TestUpstreamBodySchliessen` |
| D5 | Status 500 oder mehr des Upstream als `fehler`; `DeadlineExceeded` vor der Stufe als `fehler`; **`DeadlineExceeded` neben einem Fehler mit Code als `fehler`** | `TestWege` (`U3 Antwort 502`), `TestAbbruch` (`S1 ctx vor R context deadline exceeded: Folge`); die dritte **grün** (F-1) |
| D6 | Art der zweiten Stufe der Kette aus dem Code (`replay`); Upstream läuft innerhalb des Spans `replay` (ineinander verschachtelt); Kette zählt `ODR-E6003` statt `ODR-E5001` | `TestWege` (`Spans [replay=fehler replay=antwort], erwartet [replay=fehler upstream=antwort]`), `TestGleichzeitigeSpans`, `TestVerworfen` |
| D7 | `ODR-E6002` als `ODR-E6001` gezählt; `ODR-E6003` bei `replay` nicht gezählt; zweite Zählung von `ODR-E5001` in der Kette | `TestVerworfen` (`Zählungen "ODR-E6001", erwartet "ODR-E6002"`), `TestMethoden`, `TestGleichzeitigeSpans` (`ODR-E5001:32`, erwartet 16) |
| D8 | Zählung von `ODR-E7001` des Replay; Zählung von `ODR-E5001` bei `Bleiben` nur für `POST` | `TestNichtGezaehlt`, `TestQuelltextTelemetrie` |
| D9 | zweites `Set`; Import von `time` in `span.go`; `CountError` neben `CountDiscarded` | `TestQuelltextTelemetrie`/`TestWege` (`Span 1 hat 2 Set`), `TestQuelltext` (`span.go importiert "time"`), `pruefeBuch` in zehn Tests |
| D10 | `Senden` mit `context.WithoutCancel`; `Replay` mit abgeleitetem `ctx`; die Stufe mit `context.WithoutCancel` | `TestKontext` (`Replay: erhielt nicht das ctx des Aufrufs`), `TestUpstreamKontext`, `TestGleichzeitigeSpans` |
| D11 | Fehler vor der Stufe immer `context.Canceled` (auch bei `DeadlineExceeded`); Fehler der Stufe durch `context.Canceled` ersetzt; `kontextFehler` erkennt `DeadlineExceeded` nur unverpackt | `TestAbbruch` (`Fehler context canceled, erwartet context.DeadlineExceeded`), `TestFehlerOhneCode`, `TestUpstreamKontext` |
| D12 | `StartSpan` mit `context.Background()`; Modus `mock` als `hybrid` verdrahtet; `Telemetrie` in `verdrahteHybrid` nicht gesetzt | `TestKontext` und `TestWurzelReplayMatchLiegtImGesamtspan` (`Zeile 2 hat parent_id "", erwartet … "spn-0000000000000001"`); `TestWurzelHybridHaeltDenAdapterDesProzesses` und `TestM3Lebenszyklus` (`gerufene Konstruktoren map[replay:1 telemetrie:1 upstream:1]`); viele Tests des Pakets (`Build` scheitert) |
| D13 | Methode der Anfrage als Art; Ziel an die Art von `upstream` angehängt; Wert eines Headers der Antwort als Ergebnis; Status 404 als Teil des Ergebnisses | `TestKeinWertInDerTelemetrie` (`Span 1 trägt das Ergebnis "1"`), `TestWege`, `TestMethoden` (`Spans [GET=antwort], erwartet [replay=antwort]`) |
| D15 | `PATCH` zählt `ODR-E6001`; `MERGE` ohne Span `replay` | `TestMethoden` (`PATCH Upstream: Zählungen "ODR-E6001", erwartet "ODR-E6002"`; `Methode "MERGE": Replay 0`) |
| D16 | Ergebnis über `atomic.Value` geteilt (geordnet); Zählung jede dritte verloren (`atomic.Int64`); `End` in einer Goroutine | `TestGleichzeitigeSpans` (`Anfrage 84 (Szenario 0): Spans [replay=fehler], erwartet [replay=antwort]`; `Zähler je Code map[ODR-E5001:11 ODR-E6001:36 ODR-E6003:28]`); Raten unten |
| D17 | Flag `abgebrochen` über `atomic.Bool` geteilt | `TestAbbruchBeendetAllein` (`Anfrage 0: Spans [upstream=abgebrochen], erwartet [upstream=antwort]`) |
| D18 | `End` entfällt bei beendetem `ctx` | `TestWurzelHybridHerunterfahrenBeendetDenSpan` (`Zeilen [… odata.gateway.request]; erwartet die des Fallbacks, dann die des Gesamtspans`) |

## Auflösung der Nebenläufigkeit (Raten, `AGENTS.md` §3.10)

Je Zelle 20 Prozesse `-test.count=1` (je Arm ein Test-Binary), Container `--cpus=4` auf einem Host mit 20 Prozessoren, `GOMAXPROCS` 1 und 4, Arme ohne Detektor (`CGO_ENABLED=0`) und mit `-race`; Messender der Reviewer (Modell claude-sonnet-5-5), Treiber `rate.sh`; Last des Hosts (1-Minuten-Mittel aus `/proc/loadavg` vor und nach jeder Zelle) 1,95 bis 5,45; die Raten gelten als Größenordnung. Spalten: 1 Prozessor ohne Detektor · 1 mit · 4 ohne · 4 mit; in keinem Prozess eine Meldung `DATA RACE` oder `fatal error`.

| Test | Mutation | 1 ohne | 1 mit | 4 ohne | 4 mit |
|---|---|---|---|---|---|
| `TestGleichzeitigeSpans` | Nullprobe | 0/20 | 0/20 | 0/20 | 0/20 |
| `TestAbbruchBeendetAllein` | Nullprobe | 0/20 | 0/20 | 0/20 | 0/20 |
| `TestWurzelHybridHerunterfahrenBeendetDenSpan` | Nullprobe | 0/20 | 0/20 | 0/20 | 0/20 |
| `TestGleichzeitigeSpans` | D16 Ergebnis über `atomic.Value` geteilt, Schreiben und Lesen getrennt (geordnet) | **0/20** | **0/20** | 17/20 | 20/20 |
| `TestGleichzeitigeSpans` | D16 Zählung jede dritte verloren | 20/20 | 20/20 | 20/20 | 20/20 |
| `TestGleichzeitigeSpans` | D16 `End` in einer Goroutine (`go span.End()`) | 20/20 | 17/20 | **4/20** | **0/20** |
| `TestAbbruchBeendetAllein` | D17 Flag `abgebrochen` geteilt | 20/20 | 20/20 | 20/20 | 20/20 |
| `TestWurzelHybridHerunterfahrenBeendetDenSpan` | D18 `End` entfällt bei beendetem `ctx` | 20/20 | 20/20 | 20/20 | 20/20 |

Im ganzen Paket sind die zwei Mutationen mit Rate unter 20/20 in jedem Prozess rot: `go span.End()` durch `TestQuelltext` (`span.go: go-Statement`), `TestWege` und neun weitere Tests; die geordnet geteilte Variable durch `TestVerteilerFelder` (neues Feld) und `TestQuelltext` (Import `sync/atomic`). Die Wirkungs-Fänger D16 bis D18 allein fangen sie nicht in jedem Prozess (F-8).

**Der Treffpunkt der Schranke (D16: alle 96 Anfragen halten in ihrer Stufe, bis alle angekommen sind) ist legitim, er stellt das Rot nicht her.** Er erzeugt den Zustand, den die Zusage meint (*n* gleichzeitig offene Spans); ohne ihn läuft jede Anfrage bis zum Ende, bevor die nächste beginnt, und ein geteilter Span fiele nicht auf (die Mutation „geteilter Span“ des Implementers ist nur bei überlappenden Anfragen falsch). Die Schranke gibt keine Mutation vor: die Nullprobe ist in allen Armen 0/20, und die geordnet geteilte Variable der Reviewer-Liste bleibt bei einem Prozessor 0/20, wo die Schranke das Fenster nicht öffnet (kein Haltepunkt zwischen Schreiben und Lesen, wie Negativ (14) sagt). Die Freigabe der abgebrochenen Anfrage ist deterministisch: `cancel()` geht `gib()` voran, und ein Wartender im `select` wählt, was zuerst geschlossen wurde.

## Zählprüfungen gegen die Quelle (nicht gegen den Code des Lesers)

| Menge | Quelle | Befund |
|---|---|---|
| M-OhneSpan 62 | `SPEC-321` Nr. 7 b, `SPEC-109` Nr. 2 (41 Codes), `SPEC-284` Nr. 2 (4 Argumentfehler): 2 + 6 + 37 + 4 + 3 + 1 + 2 + 3 + 4 | stimmt; 37 gleich 41 minus 4; der Test zählt die 37 gegen Literal und gegen `len(AlleCodes())` |
| M-Wege 15 | `SPEC-321` Nr. 7 a: acht Replay (ctx vor R, Antwort, Antwort 400+, Kontextfehler, Fehler mit Code, ohne Code, `ODR-E5001` ohne und mit Upstream), sieben Upstream | stimmt (8 und 7); Ergebnisse 4 + 4 + 7 |
| M-Abbruch 11, M-Art 5, M-ctx 3, M-Wiederholung 6 | `SPEC-321` Nr. 3 d, 3 f, 7 e | stimmen; je mit Zählprüfung im Test |
| M-Verworfen 8 Fälle, 10 Zählungen | `SPEC-321` Nr. 5: sieben von acht Paaren Code × Art | stimmt; die sieben Paare werden alle erreicht (`ODR-E6001`, `-E6002`, `-E6003` mit `replay` und `upstream`, `ODR-E5001` mit `upstream`); vier Codes gleich `VerworfeneCodes()` |
| M-NichtGezählt 20, 12 Codes von R | `SPEC-296` Nr. 5: 13 Codes (`ODR-E3001`, `ODR-E5001` bis `-E5008`, `ODR-E7001`, `-E7003`, `-E7009`, `-E9001`) ohne `ODR-E5001` | stimmt (1 + 1 + 13 + 3 + 1 + 1) |
| M-Stellen-Hybrid 14 × 4 | `SPEC-321` Nr. 8: drei des Ziels, Pfad, Name und Wert der Query, Name und Wert eines Headers der Anfrage, Body der Anfrage, Name und Wert eines Headers der Antwort, Body der Antwort, zwei Texte der Ursache | 14 stimmt; 56 Zellen, davon zwei ohne Span (Benutzerteil × Antwort und Fehler des Upstream, `SPEC-261` Nr. 5 c: der Test hält 1/1/0/0, wie der Architect entschied) |
| M-Methoden 16 | `SPEC-321` Nr. 7 f: fünf mutierende Methoden × Replay, Upstream, Miss, dazu `GET` bei beiden Wegen | stimmt (5 × 3 + 1) |
| Zähler in D16 48/48/16, 112 Spans | 16 × (3 Szenarien mit `ODR-E6003`, 3 mit `ODR-E6001`, eine Kette) und 1 + 1 + 2 + 1 + 1 + 1 Spans je Gruppe von 6 | stimmt |
| Fixture gleicht nie dem Wert der Mutation | Canary `kanari7q` (Token, acht Zeichen); Fehler und Folgen als Literale | geprüft, keine Fixture gleicht einer Mutation |

Allaussagen als Schleife: „an jeder Stelle“ (14 Stellen, Schleife über `stellen()`), „jeder Code der Tabelle außer den vier“ (Schleife über `model.AlleCodes()`), „jeder Fehler von R außer `ODR-E5001`“ (Schleife über 12 Codes), „jede mutierende Methode“ (Schleife über 5). Der Journal-Fake hält jedes Argument fest (Name, Attribute, `ctx` je Span) und zählt auch `Log` und `CountError`, die der Dienst nie rufen darf.

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Mutation „`DeadlineExceeded` neben einem Fehler mit Code ergibt `fehler` statt `abgebrochen`“ (`case errors.Is(err, context.DeadlineExceeded) && !errors.As(err, new(*inbound.Fehler))` statt `kontextFehler(err)`) bleibt im ganzen Paket grün (Exit 0). Der Plan (D5, `SPEC-321` Nr. 4 b) sagt beide Kontextfehler „auch neben einem Fehler mit Code“ zu; `TestAbbruch` (S6) und `TestOhneSpan` führen neben dem Code allein `context.Canceled`, `DeadlineExceeded` steht nur allein (S1 bis S5). Ein Ablauf der Frist neben `ODR-E5001` oder `ODR-E5003` von R hieße `fehler` statt `abgebrochen`. | `AGENTS.md` §3.15; Skill: *Fänger führt nur eine Form der Eingabe* | `src/hexagon/services/hybrid/telemetrie_test.go` · `gemischt := errors.Join(miss, context.Canceled)`; `src/hexagon/services/hybrid/span.go` · `case kontextFehler(err):` | ja: Mutation `d5-deadline-neben-code-fehler`, Exit 0 im Paket | Fänger führt nur eine Form der Eingabe (Kontextfehler-Art neben Fehler mit Code) |
| F-2 | LOW | Der Kopf von `TestGleichzeitigeSpans` sagt „die Zusage „kein Zustand“ fängt der Detektor von make test-race“; `SPEC-321` *Ungeprüft* (letzter Punkt) sagt das Gegenteil („fängt der Wirkungs-Fänger und der Test über den Quelltext des Pakets, nicht der Detektor“), und gemessen fängt der Detektor eine geordnet geteilte Variable bei einem Prozessor in 0 von 20 Prozessen (Tabelle der Raten). | `AGENTS.md` §3.11, §3.7 | `src/hexagon/services/hybrid/telemetrie_gleichzeitig_test.go` · `die Zusage „kein Zustand“ fängt der Detektor von make test-race` | ja: Rate `atomic` 0/20 mit Detektor bei 1 Prozessor | Kommentar sagt mehr zu als der Test fängt |
| F-3 | LOW | Der Kopf von `TestKeinWertInDerTelemetrie` nennt „die Spans entstehen (1, 1, 2 und 1 je Art)“ und 56 Fälle; für die Stelle „Benutzerteil des Ziels“ gilt 1/1/0/0 (`inhalte.spans`), zwei der 56 Zellen tragen keinen Span und prüfen den Canary leer. Die Ausnahme steht allein am Kommentar von `spans`, nicht im Kopf des Tests. | `AGENTS.md` §3.11 | `src/hexagon/services/hybrid/telemetrie_rand_test.go` · `und die Spans entstehen (1, 1, 2 und 1 je Art)` | nein | Kommentar nennt mehr Elemente als der Test |
| F-4 | LOW | Zwei Zählprüfungen der Tabelle haben keine Assertion: D1 („Zählprüfung 8 gegen `SPEC-284` Nr. 1“; `TestNew` führt elf Fälle ohne Zählung der acht Gründe, der Kopf behauptet „die 8 von SPEC-284 Nr. 1“) und M-Folge (3 Fälle). Jeder der acht Gründe ist mit mindestens einem Fall geführt; die Zahl selbst fängt nichts. | `AGENTS.md` §3.15, §3.11 | `src/hexagon/services/hybrid/konfig_test.go` · `Die Gründe der Abweisung sind die 8 von SPEC-284 Nr. 1`; Plan §3 · `Zählprüfung 8 gegen SPEC-284 Nr. 1` | ja: Mutation, die einen Grund aus der Liste der Fälle streicht, bleibt grün | Zählprüfung im Plan ohne Assertion |
| F-5 | LOW | Die Zelle D4 der Tabelle nennt unverändert „`Set` auf einem Weg ausgelassen (15 Mutationen)“; Abweichung (e) und der Architect sagen sechs (Art × Ergebnis). D15 und M-Methoden sind im Handoff nachgezogen, D4 nicht. **Auflage** für Planner oder Implementer (Folge der Entscheidung des Architect): die Zelle D4 berichtigen, Negativ (15) um „5 von 14“ ergänzen und die Ausgänge von §6 (a) bis (e) auf *entschieden* setzen, sobald `spec-hybrid-nachzug` auf `main` liegt. | `AGENTS.md` §3.9 | `docs/plan/planning/in-progress/slice-telemetry-einbau-hybrid.md` · `Set` auf einem Weg ausgelassen (15 Mutationen) | nein | Berichtigung erreicht nicht jede Stelle |
| F-6 | LOW | Die Zeile von §6 zu den Abweichungen (a) bis (e) steht im Commit `d5b936a8`, nicht im Handoff-Commit `e1925257`, der sie in §3 meldet; `AGENTS.md` §3.13 verlangt den selben Commit wie den Bericht. Der Endzustand (Ausgang *offen*, Adresse Architect) stimmt. | `AGENTS.md` §3.13 | `docs/plan/planning/in-progress/slice-telemetry-einbau-hybrid.md` · `Abweichungen des Implementers von Zahlen und Mengen der Tabelle` | ja: `git log` | Adresse im Folge-Commit |
| F-7 | LOW | Negativ (5) nennt als Maß „6 von 6 Formen (3 Ports mit Panik oder Hängen)“. Hinter dem Beginn des Spans stehen zwei Ports (`Replay`, `Senden`, `SPEC-321` Nr. 6); `Mock` läuft vor dem Beginn, dort gibt es weder Span noch Zählung, und die genannte Folge („der Span endet nicht, die Zählung steht schon“) trifft ihn nicht. Das Maß der Klasse, für die die Folge gilt, ist 4 von 4. | `AGENTS.md` §3.15 | `docs/plan/planning/in-progress/slice-telemetry-einbau-hybrid.md` · `6 von 6 Formen (3 Ports mit Panik oder Hängen) ohne Fänger` | nein | Maß eines akzeptierten Negativs weicht von der Klasse ab |
| F-8 | LOW | Negativ (14) nennt für die geordnet geteilten Mutationen „in jedem Prozess rot“ (die zwei des Implementers) und für ein Fenster, das die Schranke nicht öffnet, kein Maß. Gemessen: geordnet geteilte Variable über `atomic.Value` 0/20, 0/20, 17/20, 20/20 (Spalten wie oben), `go span.End()` 20/20, 17/20, 4/20, 0/20 im Test `TestGleichzeitigeSpans` allein. Beide sind im Paket deterministisch rot (`TestQuelltext`: `go`-Statement, Import `sync/atomic`; `TestVerteilerFelder`: neues Feld), so dass die Zusage „kein Zustand, keine Goroutine“ strukturell getragen ist; die Folge im Betrieb ist ohne diese Tests ungefangen. | `AGENTS.md` §3.15, §3.10 | `docs/plan/planning/in-progress/slice-telemetry-einbau-hybrid.md` · `ein Fenster, das die Schranke nicht öffnet, bliebe offen` | ja: Raten oben | Akzeptiertes Negativ ohne Maß der ungefangenen Klasse |
| F-9 | INFO | Die Mutation „`New` übernimmt `Telemetrie` nicht“ ist rot durch `panic: nil pointer dereference` im ersten Test mit Fallback (`TestCodesSchleife`), nicht durch einen `--- FAIL` eines Tests der Zusage (`TestNew` prüft nur die Ablehnung von nil, nicht das Halten des Wertes). Rot, aber nicht aus dem benannten Test; die Wirkung (Panik beim ersten Fallback) ist die des akzeptierten Negativs (4). | Skill: Hinweis zu `panic:`; `AGENTS.md` §3.10 | `src/hexagon/services/hybrid/verteiler.go` · `maxBytes: k.MaxBytes, telemetrie: k.Telemetrie,` | ja: Mutation `d1-new-speichert-telemetrie-nicht` | rot durch Panik statt durch den benannten Test |
| F-10 | INFO | Abweichungen (a) bis (e) beurteilt gegen die Entscheidung des Architect: (a) richtig, die Zelle des Architect war zu eng, `SPEC-321` Nr. 7 f nennt jede mutierende Methode; (b) Ausnahme des Ziels mit Benutzerteil im Test 1/1/0/0, der Code stimmt, keine Entscheidung im Code; (c) Prüfgrenze acht Zeichen ehrlich geführt (Negativ (15)); (d) 9 von 14 am Root, fünf am Dienst, ehrlich; (e) sechs statt 15 stimmt mit dem Code (eine Stelle für `Set`); Negativ (6) am Wortlaut bestätigt (15, 14, 2). Der Dienst entscheidet keine Randform im Code; jede Zelle des Rasters in §6 nennt einen Ort. Offen bleibt allein die Auflage F-5. | `AGENTS.md` §3.12, §3.13, §3.16 | `docs/plan/planning/in-progress/slice-telemetry-einbau-hybrid.md` · `Abweichungen des Implementers von Zahlen und Mengen der Tabelle` | nein | — |
| F-11 | INFO | Die Zahl „137 Mutationen (138 Einträge)“ des Plans lässt sich an der Tabelle der Mutationen nicht nachzählen (131 bis 136 je nach Zählung der „(2 Tests)“); die Listen des Implementers im Scratchpad tragen 140 Einträge in acht Dateien. Ohne Folge für das Urteil; Rolle: Verifier. | `AGENTS.md` §3.10 | `docs/plan/planning/in-progress/slice-telemetry-einbau-hybrid.md` · `137 Mutationen (138 Einträge` | nein | Zahl im Plan nicht nachzählbar |

**Belege zu F-1.** Die Klasse ist die Menge Kontextfehler-Art × Lage (allein, neben einem Fehler mit Code) × Ort (vor der Stufe, von der Stufe; M, R, `Senden`). `kontextFehler` selbst ist durch den Bestand gedeckt (die Mutation „`DeadlineExceeded` nur unverpackt“ in `kontextFehler` ist rot in `TestFehlerOhneCode`, `TestUpstreamBodySchliessen`, `TestUpstreamKontext`); die Lücke liegt in `ergebnisVon` (Abbildung auf das Ergebnis), die allein `context.Canceled` neben einem Code sieht. Ungefangen sind in dieser Klasse die Zellen `DeadlineExceeded` neben Code bei R (Ergebnis) und bei `Senden` (Ergebnis); die Zelle bei M (kein Span) deckt der Bestand.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/services/hybrid/span.go` (`stufe`: Zählung, Beginn, `ctx.Err()`, Stufe, `Set`, `End` in der Folge von `SPEC-321` Nr. 6; `ergebnisVon`; das `ctx` von `StartSpan` verworfen) | geprüft, mit F-1; eigene Mutationen rot aus dem Grund |
| `src/hexagon/services/hybrid/verteiler.go`, `upstream.go` (Pflichtfeld, nil-Prüfung ohne Wert im Text, Code des verworfenen Versuchs an beide Stufen, Kette: `End` von `replay` vor `Entscheide`, Zählung `ODR-E5001`, Body vor `End` geschlossen) | geprüft, ohne Befund am Code (eigene Mutationen rot; F-9 zur Panik) |
| `src/bootstrap/wurzel.go` (`Telemetrie: h.tel`, ein Adapter je Prozess, nur Modus `hybrid`) | geprüft, ohne Befund (eigene Mutationen rot; Nullblöcke 349, 353, 357, 371 sind Bestand) |
| Tests in `src/hexagon/services/hybrid/` (Journal-Fake, Mengen, Schleifen, Schranken mit Frist, Meldung per `t.Errorf`) | geprüft, mit F-1, F-2, F-3, F-4; Fixtures gleichen nie dem Wert einer Mutation; Schranken tragen eine Frist von 20 s, so dass eine Mutation `--- FAIL` erzeugt (keine Panik außer F-9) |
| Tests in `src/bootstrap/` (`telemetrie_hybrid_test.go`, Fall `hybrid` von `TestWurzelReplayMatchLiegtImGesamtspan`) | geprüft, ohne Befund: Zeilen als Literal, `parent_id` der vier Zeilen, `Verworfen`/`Errors`, Identität des Adapters, drei Modi ohne Hybrid-Dienst, `Schliessen` mit Frist |
| Sensible Inhalte (`SPEC-321` Nr. 8, `LH-QA-04`) | geprüft, ohne Befund: Attribute sind drei feste Mengen; 14 Stellen × 4 Arten; neun Stellen zusätzlich an der Senke des echten Adapters |
| Kommentare im Produktcode (`AGENTS.md` §3.7) | geprüft, ohne Befund: Zusage, Kopplung, Rang-Zeiger; kein Konjunktiv, kein abwesender Text; die Zusagen von `Konfig` und `stufe` haben Fänger (D1, D3, D9, D11) |
| Plan-Nachzug (Kopf, §1, §3, §4, §6, §8) und Raster (§6) | geprüft, mit F-5, F-6, F-7, F-8; Messpunkt 1790 und Produktcode 163 stimmen; jede Zelle des Rasters nennt einen Ort oder „gilt nicht“ mit Grund |
| Katalog und Handbuch (`SPEC-201`, `SPEC-200`) | geprüft, ohne Befund: kein Fehlercode, Schlüssel oder Kommando erreichbar gemacht; `make katalog-check` grün |
| ADR und Hard Rules (Docker-only, Suppression, Gate nicht gelockert, Architektur) | geprüft, ohne Befund: kein ADR berührt, keine Direktive `//nolint`, `make a-check` 0 Befunde, `make lint` 0 Issues, kein Gate geändert |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 7 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe (Kontextfehler-Art neben Code) · Kommentar sagt mehr zu als der Test fängt · Zählprüfung im Plan ohne Assertion · Berichtigung erreicht nicht jede Stelle · Akzeptiertes Negativ ohne Maß der ungefangenen Klasse · Adresse im Folge-Commit.

## Verdikt

**Nicht freigegeben, merge-blockierend: ja, wegen F-1 (MEDIUM).** Der Produktcode stimmt in jedem geprüften Punkt mit `SPEC-321` Nr. 1 bis 9 überein (Pflichtfeld, eine Stelle `stufe` mit den vier Operationen in der Folge B1 bis B4, Kette mit `End` vor `Entscheide`, Zählung nur der vier Codes und nie `CountError`, kein Wert in Span und Zähler, Verhalten bei Ziel mit Benutzerteil nach `SPEC-261` Nr. 5 c, nur Modus `hybrid`); die Gates `test`, `test-race`, `a-check`, `lint`, `katalog-check` und `groessenmarke` sind grün, die Abdeckung des Pakets `hybrid` ist 100,0 %. Offen ist ein Fänger: `DeadlineExceeded` neben einem Fehler mit Code in `TestAbbruch` (und die entsprechende Zeile in `TestOhneSpan`).

**Übergabe:** Findings gehen an den Implementer (F-1 mit Fix gegen die Klasse, `AGENTS.md` §3.17; F-2 bis F-4, F-6 und F-7 als Nachzüge im selben Commit); die Auflage F-5 (Zelle D4, Negativ (15) „5 von 14“, Ausgänge von §6 (a) bis (e) *entschieden*) geht an Planner oder Implementer, sobald der Zweig `spec-hybrid-nachzug` auf `main` liegt; die Finding-Klassen gehen in die Slice-Closure §7. Dieser Report ersetzt keine Verifikation: DoD- und Spec-Konformität prüft der Verifier separat.
