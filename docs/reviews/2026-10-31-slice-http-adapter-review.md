# Review-Report: slice-http-adapter — 2026-10-31

**Review-Art:** Code — geprüft gegen Plan (`slice-http-adapter` §1, §3 mit der Tabelle *Zusage · Fänger* A1 bis A34, den Mengen, dem Absatz *Raten*, §4 Marke 4800, §6 mit Raster und Ausgängen), `SPEC-136`, `ARC-005`, `ARC-021`, `ADR-0002` und `AGENTS.md` §3.6, §3.7, §3.9 bis §3.18.

**Gegenstand:** `slice-http-adapter` (Welle `welle-record-replay`), `git diff 4fe4dc0..bea0afc` (HEAD `bea0afc`, Code-Commits `81d4bf5`, `6495d84`, `063d2bf`, `bea0afc`). Pfade: `src/adapters/driving/httpodata/` (Produktcode 268 Zeilen, Tests 1486), `.a-check.yml`, `spec/spezifikation.md` (`SPEC-136`), Plan.

**Skill:** `.harness/skills/reviewer.md` @ `bea0afc` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Eingangs-Kontext:**

- Plan `slice-http-adapter` (in-progress), `SPEC-136` (Nr. 1 bis 16, *Ungeprüft*), `SPEC-121`, `SPEC-135`
- `ARC-005`, `ARC-021`, `architecture.md` §2 (Importtabelle) und `ADR-0002` (a-check), `ADR-0001`
- `LH-FA-01`, `LH-FA-02`, `LH-QA-05`, `LH-QA-07`
- `AGENTS.md` (Hard Rules)

**Messverfahren.** Alle Läufe in Docker (`--memory=4g --cpus=4 --network none`, unter `flock /tmp/odata-recorder-heavy.lock`), Worktree unverändert. 192 Mutationen, je in einer Kopie des Pakets mit `go test -count=1` (Image `odata-recorder:test`); Grundlauf grün. Je Tabellenzeile A1 bis A34 mindestens eine eigene Mutation, dazu Mutationen an Elementen, die der Implementer nicht nannte (`AGENTS.md` §3.17). Nebenläufige Zeilen: `-count=50` in **einem** Prozess auf `odata-recorder:test`, Messender ich, „x von 50“ zählt `--- FAIL`-Zeilen; Datenrennen: `go test -race -c` im Image `odata-recorder:race`, dann 20 Prozesse mit `-test.count=1`, Messender die Ausgabe `WARNING: DATA RACE`. Die Größenmarke ist reproduziert: Summe `git diff --numstat` über `src/adapters/driving/httpodata/` gleich 1754 bei Marke 4800.

---

## Urteil zu `.a-check.yml`

`ports-support` hat `direction: outbound` verloren. Das ist eine **Berichtigung gemäß `ARC-021`**, keine Lockerung nach `AGENTS.md` §3.6, und braucht keine ADR.

- Quelle: `architecture.md` `ARC-021`: „jede Schicht darf sie nutzen“; die Importtabelle nennt die Querschnitts-Ports in den erlaubten Zellen von Application, Driving Adapters und Driven Adapters. Die Konfiguration trug `direction: outbound` und widersprach damit der Sicht, die `ADR-0002` spiegeln lässt („Schichten und Kanten der Datei spiegeln den Importgraph aus `spec/architecture.md` §2“). Eine ADR schärft die Sicht nicht gegen sich selbst; es ist ein Konfigurationsfehler.
- Die Kantenliste (`edges`) ist unverändert und bleibt die Quelle der Erlaubnis; entfallen ist nur die Richtungsprüfung für diese eine Schicht.
- Gegenprobe (Kopie des Baums, `a-check` am gepinnten Digest, Grundlauf auf `bea0afc`: 0 Befunde). Verbotene Kanten werden weiter gemeldet: `httpodata` → `ports/outbound`: `port-direction-mismatch`; `httpodata` → `adapters/driven/sqlite`: `lateral-adapter`; `httpodata` → `services`: `wrong-direction`; `ports/support` → `ports/outbound`: `wrong-direction`; `ports/support` → `adapters/driven/clock`: `port-impurity`; `ports/inbound` → `ports/support`: `wrong-direction`. Alle Läufe Exit 1 mit den genannten Meldungen.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | INFO | Die Änderung an `ports-support` stellt die Konfiguration auf `ARC-021` (jede Schicht darf die Querschnitts-Ports nutzen); alle verbotenen Kanten bleiben gemeldet (Gegenprobe oben). Der Plan führt die Frage in §6 mit Ausgang *offen*, Adresse Architect; dieses Urteil ist der Beleg für dessen Entscheidung. | `ARC-021`, `ADR-0002`, `AGENTS.md` §3.6 | `.a-check.yml` · `# Querschnitts-Ports (ARC-021): jede Schicht darf sie nutzen, daher ohne direction.` | ja — `make a-check` (0 Befunde) und Kopie mit verbotenen Importen | Konfiguration gleicht Sicht an |
| F-2 | LOW | Kein Fall der Gegenproben in `tools/gegenproben/*.go` nennt `ports-support`; dass die Richtungsprüfung für die Schicht entfallen ist, die Kanten aber halten, fängt nur die Kopie dieses Reviews. | `ADR-0002` (Folgepflicht Gegenprobe), `AGENTS.md` §3.15 | `.a-check.yml` · `ports-support:` (kein Treffer für `ports-support` in `tools/gegenproben`) | ja — `make gegenproben` mit einem Fall `httpodata` → `ports/outbound` | Gate-Änderung ohne Fall der Gegenprobe |
| F-3 | MEDIUM | A32 prüft bei HEAD Status und `Content-Length`, nicht die übrigen Header. Mutationen „`X-Correlation-Id` nur ohne HEAD“ (`A32-headnocid`) und „`Content-Type` nur ohne HEAD“ (`A32-headnoct`) bleiben grün, obwohl `SPEC-136` Nr. 8 „die Header bleiben“ zusagt; ein HEAD-Client verlöre die Korrelations-ID. | `AGENTS.md` §3.15, `SPEC-136` Nr. 8 | `fehler_test.go` · `func TestFehlerHead` | ja — `go test` mit den beiden Mutationen | Fänger führt nur eine Form (Header-Menge bei HEAD) |
| F-4 | MEDIUM | A9 nennt als Mutation „Adapter liest 1 Byte vor (`Peek`): jeder Fall rot“; der Plan führt sie als „nicht gefahren“ ohne Grund und Folge. Gefahren (`bufio.NewReader(r.Body)`, `Peek(1)`, `a.Body = br`) bleibt `go test` grün (Exit 0), ebenso die Hülle ohne `Peek`; „ungelesen“ und „nicht umhüllt“ (`SPEC-136` Nr. 4e) haben keinen Fänger, und die Tabelle behauptet einen. | `AGENTS.md` §3.15 (akzeptiertes Negativ nennt Grund und Folge), `SPEC-136` Nr. 4e | `slice-http-adapter.md` · `Adapter liest 1 Byte vor (`Peek`): jeder Fall rot` | ja — Mutation `Peek(1)` in `anfrageAus` | Mutation der Tabelle bleibt grün, Negativ ohne Grund |
| F-5 | MEDIUM | A15 führt keine zwei gleichen Werte desselben Headernamens; die Mutation „gleiche Werte zusammenfassen“ (`slices.Contains(ziel.Values(name), wert)`, `A15-dedupe`) bleibt grün. `SPEC-136` Nr. 5c sagt „alle Werte in ihrer Reihenfolge“; ein doppelt gesandter Wert (`Set-Cookie`, `Vary`) ginge verändert aus (`LH-QA-07`). | `AGENTS.md` §3.15 (Fall gleicher Werte) | `antwort_test.go` · `func TestAntwortHeader` | ja — Mutation `A15-dedupe` | Fänger führt nur verschiedene Werte |
| F-6 | MEDIUM | A16 führt die `Connection`-Liste nur unter dem kanonischen Namen und in einer Zeile: `name == "Connection"` statt `EqualFold` (`A16-connexact`) und „nur der erste Wert von `Connection`“ (`A16-first`, zwei Zeilen) bleiben grün; `SPEC-136` Nr. 5c verlangt „ohne Beachtung der ASCII-Schreibweise“ und nennt „ein Wert von `Connection`“. Ein Name aus einer zweiten `Connection`-Zeile oder aus `connection:` ginge an den Client. | `AGENTS.md` §3.15, §3.17 | `antwort_test.go` · `func TestAntwortHopByHop` | ja — Mutationen `A16-connexact`, `A16-first` | Fänger führt nur eine Form (Schreibweise und Zeilenzahl der Liste) |
| F-7 | MEDIUM | A13 führt neun Status; die Zusage ist „200 bis 599 unverändert“, eine endliche Menge von 400. Mutationen an Status außerhalb der neun (`206`→200, `302`→301, `401`→403: `A13-206`, `A13-302`, `A13-401`) bleiben grün; `A13-200` und `A13-clamp` werden rot. | `AGENTS.md` §3.15, §3.17 (Allaussage als Schleife über die Menge) | `antwort_test.go` · `func TestAntwortStatus` | ja — Mutationen `A13-206`, `A13-401` | Klasse als Liste statt als Schleife |
| F-8 | MEDIUM | `SPEC-136` Nr. 4b sagt „jedes Byte ab 0x80 bleibt“ und „nicht dekodiert, kodiert“; A4/A5 führen nur das gültige UTF-8-Paar 0xC3 0xA4 und feste Escape-Formen. Grün bleiben: `strings.ToValidUTF8` auf `Pfad` und auf `RawQuery` (`A4-utf8`, `A5-utf8`), `%2e`→`.` (`A4-ctl`), `%2B`→`%2b` in der Query (`A5-lowerhex`), Kürzen eines `?` am Ende der Query (`A5-qlast`), `%7e`/`%00`/`\`/`+` im Pfad (`A4-hex`, `A4-nul`, `A4-bslash`, `A4-plus`), `%2F`→`%2f` und `+` in der Query (`A5-pct`, `A5-plus`). Ein Normalisieren ungültiger Bytes änderte den Match-Key (`SPEC-116`). | `AGENTS.md` §3.15, §3.17 | `anfrage_test.go` · `func TestAnfragePfad` und `func TestAnfrageQuery` | ja — die genannten Mutationen | Menge „jedes Byte, jede Escape-Form“ als Auswahl geführt |
| F-9 | MEDIUM | A7 führt keine Header-Namen, die ein Adapter plausibel gezielt anfasst: `Authorization`, `Cookie`, `Content-Type`, `Accept` entfernen (`A7-auth`, `A7-cookie`, `A7-ct`, `A7-accept`) bleibt grün; `X-Recorder-*`, `Expect`, `Connection`, `Content-Length`, `Host` werden rot. `SPEC-136` Nr. 4d: „alle Namen mit allen Werten“, und `LH-FA-14` führt gerade `Authorization` als Geheimnis. | `AGENTS.md` §3.15, `SPEC-136` Nr. 4d | `anfrage_test.go` · `func TestAnfrageHeader` | ja — Mutationen `A7-auth`, `A7-cookie` | Fänger führt nur eine Form (nur gewählte Namen) |
| F-10 | MEDIUM | `http.Server.ErrorLog` ist auf `io.Discard` gesetzt; das steht weder in `SPEC-136` noch in §6, hat keinen Fänger (Entfernen der Zeile: Exit 0; `panic(errors.New(...))` statt `http.ErrAbortHandler`: Exit 0) und widerspricht der Zusage in *Ungeprüft* („`net/http` fängt sie, schreibt eine Meldung“). Panik des Dienstes oder ein Typed-nil beendet die Verbindung ohne eine Zeile in irgendeiner Senke. | `AGENTS.md` §3.12 (Randform im Code entschieden), Maintainability | `server.go` · `ErrorLog: log.New(io.Discard, "", 0),` | ja — Mutation `X-errorlog`, `X-abortgeneric` | Verhalten im Code entschieden, ohne Ort und ohne Fänger |
| F-11 | LOW | `DisableGeneralOptionsHandler` und `EnableFullDuplex` stehen in §6 Risiko 7 in einer Zeile mit dem Ausgang „geprüft vom Implementer“; die Klausel „Offen, Adresse Architect“ steckt im Fließtext derselben Zeile, nicht als eigene Zeile mit Ausgang *offen*. Das Verhalten ist in `SPEC-136` Nr. 2, 4c (jeder Request, auch `OPTIONS *`) und Nr. 3, 4h (Leihe bis `Close`) entschieden; es fehlen nur die Mittel in Nr. 2 und 4h. Fänger vorhanden (A6 Fall 2: `A6-opts` rot in `TestAnfrageFremdesTarget`; A23: `A23-nofd` rot in `TestLeihe`). | `AGENTS.md` §3.13 | `slice-http-adapter.md` · `**Offen, Adresse Architect:** `SPEC-136` Nr. 2 und Nr. 4h nennen beide Mittel nicht` | nein | Offener Punkt in Zeile mit anderem Ausgang |
| F-12 | LOW | A23 prüft die Leihe von `Header` und `Body` nur auf dem Erfolgspfad. Schließt der Adapter `r.Body` im Fehlerpfad (`A23-bodyclose-fehler`), bleibt alles grün, obwohl Nr. 4e „weder gelesen noch geschlossen“ ohne Einschränkung sagt. | `AGENTS.md` §3.15 (Aufrufstellen) | `server.go` · `schreibeFehler(w, r, id, antwort, err)` | ja — Mutation `A23-bodyclose-fehler` | Fänger führt nur eine Form (Pfad) |
| F-13 | LOW | A34 fängt als Wirkungs-Fänger den gemeinsamen Puffer (50 von 50 im `-count=50`-Prozess), aber nicht gemeinsamen Zustand mit schmalem Fenster: globale ID (`A34-sharedid`) und globale `Anfrage` (`A34-sharedanfrage`) bleiben im Wirkungslauf 0 von 50 grün; erst der Detektor fängt sie (20 von 20 Prozessen). Der Plan nennt „gemeinsame Anfrage wiederverwendet: Antworten vertauscht, rot (Wirkung)“. | `AGENTS.md` §3.10 | `slice-http-adapter.md` · `Paket-Variable als Puffer oder gemeinsame `Anfrage` wiederverwendet: Antworten vertauscht, rot (Wirkung)` | ja — `A34-sharedanfrage` mit `-count=50`, dann `-race` | Nebenläufige Zusage trägt der Detektor, nicht die genannte Wirkung |
| F-14 | LOW | Zwei Kommentare nennen, was `net/http` ohne die Einstellung täte (`beantwortete ihn sonst selbst`; `ohne Vollduplex verwirft net/http … den ungelesenen Rest`). Gemeint ist eine Kopplung, die Fassung mit „sonst“ und „ohne“ liest sich als verworfene Alternative. | `AGENTS.md` §3.7 | `server.go` · `net/http beantwortete ihn sonst selbst.` und `antwort.go` · `ohne Vollduplex verwirft` | nein | Kommentar im Konjunktiv über die Alternative |
| F-15 | LOW | Unbeschränkte Mengen nennen keine Prüfgrenze mit Grund und Folge: Request-Body bis 8 MiB + 1 (Grenze 16 MiB, `A9-max16m`, grün), Antwort bis 1 MiB + 1 (Kappen bei 2 MiB, `A18-cap2m`, grün), Werte je Name (Kappen bei 64, `A15-cap`, grün), `Content-Length` bei 204/304 (`A20-204`, `A20-304`, grün; `SPEC-136` Nr. 6d schweigt dazu). | `AGENTS.md` §3.17 | `slice-http-adapter.md` · `Mengen aus dem Wortlaut` | nein | Unbeschränkte Menge ohne Prüfgrenze |
| F-16 | INFO | Die Bereichsprüfungen in `schreibeFehler` sind gegen `StatusErlaubt` redundant: `Status >= 399`, `<= 599` entfernt, `<= 600` (`A27-ge399`, `A27-le599`, `A27-le600`) bleiben grün; der Zweig `f != nil` hat keinen Fänger (kein Fall mit einem Typed-nil `*inbound.Fehler`). Der Plan nennt als Mutation „Prüfung `Status >= 400` entfernt“; sie wird rot (`A27-ge400`). | Maintainability | `fehler.go` · `f != nil && f.Status >= 400 && f.Status <= 599` | ja | Redundante Prüfung ohne eigenen Fänger |
| F-17 | INFO | Zuständigkeit Architect: `SPEC-136` Nr. 2/4h nennen `DisableGeneralOptionsHandler`, `EnableFullDuplex` und die Folge für die Wiederverwendung der Verbindung (`net/http` leert oder schließt bei ungelesenem Body) nicht; ob sie dort stehen, entscheidet er (siehe F-11). | `AGENTS.md` §3.12 | `spec/spezifikation.md` · `SPEC-136` Nr. 4h | nein | Mittel nicht in der Spezifikation |

---

## Zu den vier Prüfpunkten

1. `.a-check.yml`: Berichtigung nach `ARC-021`, siehe Urteil oben (F-1, F-2). Nicht zurückgedreht.
2. Die zwei im Code gelösten Annahmen: Das Verhalten steht in `SPEC-136` (Nr. 2, 4c, 3, 4h); nur die Mittel fehlen. §6 trägt sie mit Ausgang *offen* und Adresse Architect (Risiko 7), aber im Fließtext einer Zeile mit Ausgang „geprüft“ (F-11, F-17). Beide haben einen Fänger, der aus dem benannten Grund rot wird (`A6-opts` → `TestAnfrageFremdesTarget`; `A23-nofd` → `TestLeihe`, 50 von 50).
3. Akzeptierte Negative: A5 (`r.URL.RawQuery` byte-gleich, äquivalent; gemessen grün — bestätigt: `A5-urlraw` Exit 0) trägt Grund. A26 `Content-Length` bei GET (äquivalent) trägt Grund, A32 fängt es (`A26-cl`, `A32-nocl` → `TestFehlerHead` rot). A32 „Body bei HEAD“ (`A32-headbody` grün) trägt Grund und Folge. A31 „`message` aus `err.Error()`“ nicht gefahren: Grund (die Meldung stammt aus dem Modell), Folge fehlt (INFO); die erreichbaren Varianten (`A31-err-hdr`, `A31-hdr-echo`, `A31-path-hdr`) sind rot. **A9 `Peek` trägt weder Grund noch Folge und bleibt grün (F-4). A31 gefahren: `A31-query-body` ist grün, weil `Content-Length` den Überhang abschneidet (äquivalent).**
4. Der Adapter enthält keine Grenzen des Servers (`ReadHeaderTimeout`, `IdleTimeout`, `MaxHeaderBytes`) und kein `Anhalten`/`Schliessen`; die Importe sind Standardbibliothek, `model`, `ports/inbound`, `ports/support` (`make a-check` auf `bea0afc`: 0 Befunde).

## Auflösung der Tabelle (eine eigene Mutation je Zeile; Test, der rot wurde)

Rot aus dem benannten Grund (Testname nennt die Abweichung), `-count=1`, wenn nicht anders vermerkt. Tests, die zusätzlich rot wurden, nenne ich nicht.

| Zeile | Mutation (Kopie) | Roter Test |
|---|---|---|
| A1 | `Listener`/`IDs`/`Verarbeiten`-Prüfung je entfernt | `TestNewUngueltig` (3 von 3) |
| A2 | `Serve` liefert `nil` | `TestServeEndet` |
| A3 | `ToUpper`, `ToLower` der Methode | `TestAnfrageMethode` |
| A4 | `EscapedPath()`, `URL.Path`, `;` gekürzt, `/` am Ende gekürzt | `TestAnfragePfad` (grün bleiben F-8) |
| A5 | Teilung am letzten `?`, `&&`→`&`, `?` ohne Query behalten | `TestAnfrageQuery` (grün bleiben F-8) |
| A6 | `URL.Path`; `DisableGeneralOptionsHandler` entfernt | `TestAnfrageFremdesTarget` |
| A7 | erster Wert, Werte verbunden, Reihenfolge umgekehrt, `X-Recorder-*`, `Host`, `Content-Length`, `Connection`/`Upgrade`, `Expect`, leerer Wert, Map nil | `TestAnfrageHeader` (grün bleiben F-9) |
| A8 | `Body` immer gesetzt; `Laenge` -1 und 0 | `TestAnfrageBodyForm` |
| A9 | `LimitReader` 8 MiB und 64 KiB | `TestAnfrageBodyBytes` (16 MiB und `Peek` grün: F-4, F-15) |
| A10 | zweiter Aufruf, Konstante, Aufruf im Fehlerpfad | `TestKorrelationsID`, `TestKorrelationsHeader` |
| A11 | zweiter Aufruf; Wiederholung nach Fehler und nach Kontextfehler | `TestEinAufruf` |
| A12 | `context.Background()`, `WithTimeout`, `WithoutCancel` | `TestKontextClientAbbruch` (`WithCancel`/`WithValue` grün, äquivalent bis auf `Value`) |
| A13 | auf 598 gedeckelt; 200→201 | `TestAntwortStatus` (grün bleiben F-7) |
| A14 | `< 199`, `< 100`, `> 600`, `> 598`, Prüfung auf Body nil entfernt, Header trotzdem, 502 statt 500 | `TestAntwortVerstoss` |
| A15 | `Set` statt `Add`, leere Werte übersprungen, nur erster Wert, umgekehrt, `Set-Cookie` einzeln | `TestAntwortHeader` (grün bleiben F-5, F-15) |
| A16 | je einer der neun Namen gestrichen (9 von 9), `Connection`-Liste ignoriert, Tab nicht entfernt, kein `ToLower`, nur erster Teil | `TestAntwortHopByHop` (grün bleiben F-6) |
| A17 | Unterdrückung des Sniffings entfernt; Prüfung am Dienst-Header | `TestAntwortContentType` |
| A18 | Kappen bei 32768 | `TestAntwortBytes` (2 MiB grün: F-15) |
| A19 | `(n, EOF)` verworfen, `(n, Fehler)` verworfen, `(0, nil)` als Ende | `TestAntwortReadFormen` |
| A20 | `Content-Length: 0` bei HEAD; bei HEAD entfernt | `TestAntwortOhneBody` (204/304 grün: F-15) |
| A21 | `Close` überall, im Statusverstoß, im Fehlerpfad entfernt; doppelt; vor dem Kopieren | `TestCloseEinmal` |
| A22 | Schreibfehler ignoriert; Schreib- und Flushfehler ignoriert | `TestSchreibfehler` (Flushfehler allein ignoriert: grün, äquivalent, der nächste `Write` meldet den Fehler) |
| A23 | `Close` in Goroutine; ohne `EnableFullDuplex`; `X-Recorder-*` gelöscht; `User-Agent` gelöscht; Header ergänzt oder verändert; `r.Body` gelesen oder geschlossen | `TestLeihe` (Fehlerpfad grün: F-12) |
| A24 | ohne `Flush`; `Flush` nur für das erste Stück; nur für jedes zweite | `TestStreaming` und `TestLeihe` |
| A25 | Lesefehler als Ende; Fehler vor `n` Bytes; kein `Flush` vor Abbruch; `Flush` vor erstem Byte; kein Abbruch | `TestLesefehler` |
| A26 | Status 404, 500; `ODR-E4001` fest 502; `Content-Type`, `X-Correlation-Id`, `Content-Length` entfernt; Charset; Einrückung; Zeilenende; Länge +1 | `TestFehlerobjektJeCode` (`X-Correlation-Id` kleingeschrieben: grün, Fixture gleicht dem Wert) |
| A27 | `StatusErlaubt`, `>= 400`, Obergrenze 499, Rückfall 502 | `TestFehlerUnzulaessig`, `TestFehlerobjektJeCode` |
| A28 | Typ-Assertion; ein Unwrap | `TestFehlerVerpackt` |
| A29 | `ErrUngueltig`→400; andere Fehler→502 | `TestFehlerOhneCode` |
| A30 | Prüfung entfernt; nach `*Fehler`; nur `Canceled`; nur `DeadlineExceeded`; `==` statt `errors.Is`; 500 statt Abbruch | `TestKontextfehler` |
| A31 | `err.Error()` im Header; Pfad im Header; alle Request-Header im Header | `TestFehlerObjektKeineWerte` |
| A32 | `Content-Length` entfernt; `0`; Status 200 bei HEAD | `TestFehlerHead` (Body und andere Header: F-3) |
| A33 | Header an jeder Antwort; nur wenn leer; Wert des Dienstes entfernt | `TestKorrelationsHeader` |
| A34 | gemeinsamer Puffer; globale ID; globale `Anfrage` | `TestGleichzeitig` nur beim Puffer (F-13) |

## Raten der nebenläufigen Zeilen (`AGENTS.md` §3.10)

Messender: ich; Wirkung `-count=50` in einem Prozess (4 CPUs, 4 GiB, `go test -run 'Test(KontextClientAbbruch|Schreibfehler|Leihe|Streaming|Gleichzeitig)$'`); Grundlauf ohne Mutation: 0 von 50 rot.

| Zeile | Mutation | rot (Wirkung) | Detektor (20 Prozesse, `-test.count=1`) |
|---|---|---|---|
| A12 | `context.Background()` | 50 von 50 | — |
| A12 | `WithTimeout` | 50 von 50 | — |
| A22 | Schreibfehler ignoriert | 50 von 50 | — |
| A22 | Schreib- und Flushfehler ignoriert | 50 von 50 | — |
| A23 | `Close` in Goroutine | 50 von 50 | — |
| A23 | ohne `EnableFullDuplex` | 50 von 50 | — |
| A23 | `X-Recorder-*` gelöscht | 50 von 50 | — |
| A23 | `User-Agent` gelöscht (Header, den der Implementer nicht nannte) | 50 von 50 | — |
| A24 | ohne `Flush` | 50 von 50 (`TestStreaming` und `TestLeihe`) | — |
| A24 | `Flush` nur für das erste Stück | 50 von 50 | — |
| A24 | `Flush` nur für jedes zweite Stück | 50 von 50 | — |
| A34 | gemeinsamer Puffer | 50 von 50 | 20 von 20 |
| A34 | globale ID | 0 von 50 | 20 von 20 |
| A34 | globale `Anfrage` | 0 von 50 | 20 von 20 |
| A34 | globale ID erst im Fehlerpfad gelesen | 0 von 50 in `TestGleichzeitig` (rot in `TestFehlerobjektJeCode`, `-count=1`) | 0 von 20 |
| — | ohne Mutation | 0 von 50 | 0 von 20 |

Die Raten des Implementers (50 von 50 für A12, A22, A23, A24, A34 mit Puffer; Detektor 20 von 20) sind bestätigt. Die Lücke liegt bei A34 mit schmalem Fenster (F-13).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driving/httpodata/server.go` (Konfiguration, `Serve`, Ablauf) | geprüft; Befund F-10, F-14 |
| `src/adapters/driving/httpodata/anfrage.go` | geprüft; Fänger-Befunde F-8, F-9 (Code selbst ohne Befund) |
| `src/adapters/driving/httpodata/antwort.go` | geprüft; Fänger-Befunde F-5, F-6, F-7, F-15 (Code selbst ohne Befund) |
| `src/adapters/driving/httpodata/fehler.go` | geprüft; Befunde F-3, F-12, F-16 |
| Importe und Grenzen (kein Server-Limit, kein `Anhalten`/`Schliessen`, keine Importe außerhalb Standardbibliothek, `model`, `ports/inbound`, `ports/support`) | geprüft, ohne Befund |
| `spec/spezifikation.md` `SPEC-136` und die Berichtigungen an `SPEC-121` Nr. 2c, `SPEC-135` Nr. 7d | geprüft, ohne Befund (Widerspruch zum Code: nur F-10) |
| Größenmarke (1754 gegen 4800) und Plan §1/§3/§6 gegen den Code (`AGENTS.md` §3.9) | geprüft, ohne Befund |
| Rohbytes (`AGENTS.md` §3.18): kein `\u`+Hex in den Tests, Umlaute als UTF-8 | geprüft, ohne Befund |
| `harness/`, `Makefile`, `Dockerfile` | nicht berührt; nicht geprüft |

## Summary

0 HIGH, 8 MEDIUM, 6 LOW, 3 INFO. Der Code bricht keine Zusage von `SPEC-136`, die ein Test aufträgt; die MEDIUM-Befunde sind Zusagen oder Formen, deren Mutation grün bleibt (F-3 bis F-9) und ein im Code entschiedenes Verhalten ohne Ort und ohne Fänger (F-10). `.a-check.yml` ist eine Berichtigung nach `ARC-021` ohne ADR-Pflicht (F-1). Kein Finding wurde auf Einspruch herabgestuft; ein Rollen-Widerspruch liegt nicht vor.
