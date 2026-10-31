# Review-Report: slice-replay-service — 2026-10-31

**Review-Art:** Code — geprüft gegen Plan (`slice-replay-service` §1, §3 mit der Tabelle *Zusage · Fänger* Z1 bis Z64, der Zusatztabelle, dem Absatz *Mutationen und Raten des Implementers*, §4 mit Block, Marke 9200 und Basis `b919969`, §6 mit Raster und offener Randform), gegen `SPEC-135` und die Nachbarn (`SPEC-045`, `SPEC-052`, `SPEC-109` Nr. 5, `SPEC-114` Nr. 9, `SPEC-116`, `SPEC-118` Nr. 10, `SPEC-129` Nr. 11 a, `SPEC-122`) und gegen die Hard Rules (`AGENTS.md` §3.1 bis §3.18). Nicht geprüft wird gegen die DoD; das ist die Frage des Verifiers.

**Gegenstand:** `slice-replay-service` (Welle `welle-record-replay`), `git diff 1296515..be43c73` (HEAD `be43c73`). Pfade: `src/hexagon/ports/inbound/replay.go`, `src/hexagon/services/replay/` (Produktcode `replay.go`, `dienst.go`, `auswahl.go`, `antwort.go`, `key.go`, `fehler.go`), Unterpaket `src/hexagon/services/replay/begrenzt/`, Tests und Fakes; daneben die Plan-, Roadmap- und Spezifikationsänderungen des Diffs.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `be43c73` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Methode.** Der Bericht des Implementers (Plan §3) wurde gelesen; keine Mutation, keine Rate und keine Einschätzung daraus ist übernommen. Alle Mutationen laufen auf Kopien im Scratchpad (`/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/review-replay/`; im Repo liegt nichts außer diesem Report), im Image `odata-recorder:test` bzw. `odata-recorder:race` mit `--network none --memory=4g --memory-swap=4g --cpus=4`, Gruppen unter `flock /tmp/odata-recorder-heavy.lock`, je Mutant ein Lauf `go test -count=1 ./src/hexagon/services/replay/...` (Basislauf ohne Mutation: grün). Mutanten, die nicht übersetzen (3), zählen nicht. Die Grundlage der Klassenprüfung (`AGENTS.md` §3.17) sind Mengen als Produkt ihrer Achsen; wo eine Achse im Test nur eine Ausprägung führt, steht ein Mutant an einer anderen.

**Eingangs-Kontext:**

- Slice-Plan `slice-replay-service` in `be43c73` (§1, §3, §4, §6)
- `SPEC-135` Nr. 1 bis 18 und *Ungeprüft*; `SPEC-045`, `SPEC-052`, `SPEC-109` Nr. 5, `SPEC-114` Nr. 9, `SPEC-116`, `SPEC-117`, `SPEC-118` Nr. 10, `SPEC-129` Nr. 11 a, `SPEC-122` Nr. 8, `SPEC-128`
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md)
- [LH-FA-04, LH-FA-14, LH-QA-03](../../spec/lastenheft.md)
- `AGENTS.md` (Hard Rules); Reviewer-Skill mit den Zeilen *Fänger führt nur eine Form der Eingabe*, *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*, *Akzeptiertes Negativ ohne Grund und Folge*; Reports `2026-10-30-slice-workspace-use-case-review*.md` als Maßstab

---

## Antworten auf die Fragen des Auftrags

**(1) Tabelle Z1 bis Z58, Zusatztabelle, Raster.** 112 gültige Mutanten (Produktcode, Fakes und Detektor-Varianten), **93 rot, 19 grün**. Von den 19 grünen sind 4 äquivalent oder tot (I-1), 1 ist mit gültigen Fixtures nicht erreichbar (F-9) und 14 sind Lücken des Beweises (F-1 bis F-8; `b1` ist LOW). Mit mindestens einer eigenen Mutation aufgelöst sind Z1 bis Z41 und Z43 bis Z58 (Z5 nur am Strom, Z16 an einem von 17 Paaren); **nicht gefahren** ist allein Z42 (beschädigter Blob verbraucht nichts: die Zusage trägt `TestFehlerLassenCursorStehen` über Z33, `z33` rot). Zeile für Zeile, Kürzel = Mutant, rot = der Zeilen-Fänger wurde aus dem Grund der Zeile rot:

| Zeile | Mutation (eigene, auf Kopie) | Ergebnis |
|---|---|---|
| Z1 | Fünf-Ports-Prüfung (Transaktion), Nullwert in `Pfade`, `last` nicht angenommen; nur der erste Dienst und nur der erste Pfad geprüft | rot (`TestNewKonfig`); die beiden „nur der erste“ **grün** (F-4) |
| Z2 | `New` ruft `Sessions.Get` | rot |
| Z3 | `Laenge < -2`; `Laenge >= 0` bei nil | rot (`TestArgumente`) |
| Z4 | `ctx` an S1, S3, S4, S5, S6, S7, S8 und S10 je durch `Background()` oder `WithoutCancel` ersetzt | 8 von 8 rot (`TestKontextJederStelle`) |
| Z5 | Strom bei Fehler nicht geschlossen | rot; die Abbildung der Kontextfehler liegt im Port-Paket (`outbound.Abbilden`), nicht im Diff, und ist dort nicht mutiert |
| Z6 | Gegenstand an S8 falsch; `fehlt` hält nur `ErrBeschaedigt` aus | erste rot; zweite **grün** (F-3) |
| Z7 | `Status` fest 404 | rot |
| Z8 | Segmentgrenze fehlt; kürzester Root; „letzter Treffer gewinnt“; „erster Treffer gewinnt“; `rest = ""`; `strings.ToLower` statt ASCII | Segmentgrenze, kürzester Root, „erster“, `rest` rot; „letzter“ und `ToLower` **grün** (F-4, F-2) |
| Z9 | Pfad mit Root im Key; feste Dienst-Kennung | rot |
| Z10 | mehr als ein Wert; leerer Wert | rot (`TestSteuerheader`) |
| Z11 | Grenze 129; 6 Mutanten an den Rändern der Zeichenklassen (`/`, `:`, `@`, `[`, `` ` ``, `{`); `len < 0` | alle rot bis auf `len(id) < 0`: äquivalent (I-1) |
| Z12 | `TrimSpace` auf die Kennung | rot |
| Z13 | bei fehlenden Headern `Sessions.List` | rot (`TestKeineImplizit`) |
| Z14 | `open` wie `incomplete` zugelassen; Service der Session nicht verglichen | rot |
| Z15 | Service des Workspace nicht verglichen | rot |
| Z16 | Session und Body vertauscht | rot (`TestRangfolge`); 16 der 17 Paare nicht einzeln mutiert (der Fänger ist die Menge von 17) |
| Z17 | `Metadata` wahr bei `$metadata`; Grenze der Ausgabe um 1 zu klein; Content-Type nur unter kanonischem Namen | die ersten zwei rot; der dritte **grün** (F-2) |
| Z18 | Trennzeichen `;` statt `,`; `TrimSpace` statt Leerzeichen und Tab; Name nur kanonisch | erste rot; die beiden anderen **grün** (F-2) |
| Z19 | `RequestRedaction` immer `none` | grün, äquivalent (Z61; I-1) |
| Z20 | `X-Recorder-*` vor dem Key entfernt | rot |
| Z21, Z22 | `reject` ignoriert; `E9001` als `E5006` | rot |
| Z23 | `Canceled` als `E5001` | rot (`TestLesefehlerBody`) |
| Z24 | Guard `MatchKey != ""` entfernt | grün, tot (I-1) |
| Z25 | zweiter Aufruf von `ListBySession`; erstes Element statt kleinstes | rot |
| Z26 | `consumed` ignoriert; größte statt kleinste | rot |
| Z27 | `position` statt `position`+1; `E5007` und `E5008` vertauscht; Strict auf der Kandidatenliste statt der Session | rot |
| Z28 | `first` und `last` vertauscht; `first` ohne Ordnung (erstes Element der Liste) | rot |
| Z29 | über Z26 (Verbrauch nach `SequenceNo`) | rot (`TestBeispieleLH`) |
| Z30 | Konfig vor Cursor | rot |
| Z31 | `first` schreibt nie, `last` schreibt; `Position` gleich `SequenceNo`; Replay- und Recording-Session vertauscht | rot |
| Z32 | Identität: `Get` mit vertauschten Sessions; mit fester Replay-Session | rot |
| Z33 | `Put` vor der Auswahl | rot |
| Z34, Z35 | `Status` fest 200; `Open` auch bei `absent` | rot |
| Z36 | `content-encoding` nicht entfernt; Schreibweise der Konfig-Namen | rot |
| Z37 | Header mit erstem Wert Marker entfernt; leere Wertliste entfernt | rot |
| Z38 | Map des Exchange zurückgegeben | rot |
| Z39 | Header immer | rot |
| Z40 | Stufe `INFO`; Attribut `session_id` aus dem Workspace; Eintrag doppelt | rot |
| Z41 | `Replayfaehig(e, true)` in `sequential`; in `strict`; in `first` und `last` | `sequential` rot; die beiden anderen **grün** (F-1) |
| Z43 | kein `Close` bei Fehler; doppeltes `Close`; `Close` nur bei Fehler von `f` | rot |
| Z44 | `ctx.Err()` zwischen Bestätigung und Eintrag | **grün** (F-5) |
| Z45 | Feld `Upstream outbound.UpstreamClient` | rot |
| Z46 bis Z53 | siehe Raten | rot, 50 von 50 |
| Z54 | `sync/atomic`, `sync`, Methode auf `Konfig`, `var zaehler` im Paket | rot (`TestPaketOhneZustand`) |
| Z55 | `Blobs.Verify` zählt nicht; `Cursors.List` zählt nicht; Rollback fehlt | rot |
| Z56 | Strom hinter `struct{ io.ReadCloser }` versteckt; `io.NopCloser` | Hülle versteckt: **grün** (F-6); `NopCloser`: rot nur über `Close` |
| Z57 | `Cursors.ResetAll` im Fenster der Transaktion | rot |
| Z58 | Import von `adapters/driven/sqlite` im Paket | `make a-check`: 1 Befund `app-impurity`, Exit 1; ohne Mutation 0 Befunde, Exit 0 |
| `begrenzt` | `<` statt `<=`; Rest nicht geschrieben; „nil bei leerem Ergebnis“ entfernt | die ersten zwei rot; die dritte **grün** (F-8) |

Die Zusatztabelle (Paare P1 bis P17, Fänger Z16) hat als Menge einen Fänger; eine vertauschte Reihenfolge ist rot. Das Raster F1 (112 Zellen, 16 Zeilen) hat keine leere Zelle und keine Zelle ohne Ort; die Zellen mit Fänger sind über die Zeilen oben aufgelöst. Zwei Zellen haben ihren Ort, aber keinen Fänger: `Methode` × fremd (*die Schreibweise zählt, `get` ist unbekannt*; F-2) und `Content-Encoding`/`Content-Type` × fremd (Name nur kanonisch; F-2). Das Raster *Antwortart × Auslieferung* ist über Z34 bis Z41 aufgelöst; die Zeile *Response-Redaction `applied`* trägt den Fänger nur für `sequential` (F-1).

**(2) Klassen (`AGENTS.md` §3.17).** Mengen gegen den Wortlaut gezählt: Z26 1755 = 585 × 3, Z27 939 = 313 × 3, Z28 3510 = 2 × 585 × 3 (die Tests rechnen mit Zählprüfung nach; das Orakel ist eine Handtabelle). Die Achse *Strategie* ist in Z24, Z25, Z29, Z31, Z39 geführt, **nicht** in Z41 (F-1); die Achse *Position in der Menge* (Dienste der Konfig, Pfade) ist nicht geführt (F-4); die Achse *Schreibweise* (Name, Zeichenfaltung, Rand, Methode) führt nur die kanonische Form (F-2). Der Mutant `Laenge > 0` statt `!= 0` (ein Body mit unbekannter Länge würde nicht gelesen) ist allein über `TestLesefehlerBody` rot; ein nichtleerer, fehlerfreier Body mit `Laenge` -1 hat keinen eigenen Fall.

**(3) Nebenläufigkeit.** Die Wirkungs-Fänger Z46 bis Z53 sind in jedem Lauf rot; Raten unter *Raten*. Der Detektor fängt weder die Fenster noch den Zähler (Z63, akzeptiertes Negativ mit Grund und Folge, bestätigt).

**(4) Abweichungen und Lücken des Implementers.** Z24 `strict`: in `strict` ist der Exchange ohne Key an `position`+1 ein anderer Key (`E5007`); `SPEC-135` Nr. 9 trägt das, der Test führt es (8 Elemente). Der begrenzte Puffer als Unterpaket `begrenzt` folgt `SPEC-135` Nr. 17 (kein weiterer Typ mit Methoden im Paket `replay`); `make a-check` meldet 0 Befunde, `Puffer` ist nicht für gleichzeitigen Gebrauch und sagt das. Die Zusätze zu Z8 (`/A/b`, `/a/B`), Z17 (Eingabegrenze bei kürzerer Ausgabe) und Z40 stehen im Plan und im Test. Der Kontextfehler kehrt „selbst“ zurück (`context.Canceled` bzw. `DeadlineExceeded` als Wert, nicht der Verbund): `SPEC-135` Nr. 7 c sagt „mit dem Kontextfehler selbst“; `TestLesefehlerBody` ist für `Canceled` und `DeadlineExceeded` als `E5001` rot gesehen (Z23). Die Verletzung der Betriebsregel mit `sed` ist im Diff nicht abzulesen (I-3).

**(5) Hard Rules.** Keine `//nolint`-Direktive im Diff; keine Änderung an `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/` oder `AGENTS.md`; `spec/architecture.md` unberührt (`AGENTS.md` §3.4); keine neue ADR, kein Gate gelockert. Docker-only: alle Läufe über `make` oder Images, netzlos. `make a-check` prüft `src/hexagon/services/**` als Schicht `services`; der Dienst importiert keinen Adapter (Z58, Mutation rot gesehen) und kein `net/http`.

**(6) Randform `discarded`.** Siehe F-9.

**(7) Größenmarke.** `git diff --numstat b919969..be43c73` über `src/hexagon/ports/inbound/replay.go` und `src/hexagon/services/replay`: **4715** Einfügungen und Löschungen (4715 + 0), wie der Lauf von `make groessenmarke` („ok, Messung 4715 von Marke 9200“). Die Marke ist gehalten; die Schätzung 6765 wurde um 30 % unterschritten (Handoff ÷ Schätzung 0,70; die Faktoren aus `SPEC-128` Nr. 2 nennen 0,72 bis 1,03). Der Plan trägt die Zahl 4715 nicht (F-7).

---

## Raten

Messender: der Reviewer; Image `odata-recorder:test` (Zeilen *ohne Detektor*) bzw. `odata-recorder:race` (Zeilen *Detektor*); je Zeile **50 Prozesse** mit `-test.count=1` über die sechs Fänger `TestGleichzeitigGleicherCursor`, `TestZweiReplaySessions`, `TestReplayGegenReset`, `TestFensterDerTransaktion`, `TestFensterLoeschen`, `TestReplayGegenLoeschen`; die Binärdatei wird einmal gebaut und 50-mal gestartet. Last: 4 CPUs, 4 GB, Host mit weiteren Läufen unter `flock`.

| Mutation (auf Kopie) | Zeilen | ohne Detektor | Detektor |
|---|---|---|---|
| keine (Normalstand) | – | 0 von 50 | 0 von 50 |
| `Put` in einer zweiten Transaktion (Cursor in der ersten gelesen; Fenster 3) | Z46, Z48, Z51, Z52 | **50 von 50**, jeder der sechs Fänger 50 von 50 | **50 von 50**, jeder 50 von 50 |
| `Open` und `Put` in einer zweiten Transaktion (Fenster 2) | Z46, Z47, Z48, Z51 | **50 von 50**; `TestFensterDerTransaktion`, `TestFensterLoeschen`, `TestGleichzeitigGleicherCursor`, `TestReplayGegenLoeschen`, `TestReplayGegenReset`, `TestZweiReplaySessions` je 50 von 50 | – |
| `Workspaces.Get` in eigener Transaktion (Fenster 1) | Z50, Z53 | **50 von 50**; `TestFensterDerTransaktion`, `TestFensterLoeschen`, `TestReplayGegenLoeschen` je 50 von 50 | – |
| Cursor vor `Sessions.Get` in eigener Transaktion gelesen (Fenster vor der Transaktion) | Z49 | ein Lauf rot (`TestFensterDerTransaktion`, Meldungen *Z49 nach Sessions.Get* und *Z49 nach ListBySession*, `seq:2`) | – |
| Zähler `zaehler++` am Anfang von `Replay` (Paketvariable) | Z54, Z63 | – | **0 von 50** (Z63, akzeptiertes Negativ) |
| `var zaehler` im Paket | Z54 | `TestPaketOhneZustand` rot (Quelltext) | – |

Die Fenster sind Wirkungs-Fänger wie geplant: Kanal, Haltepunkt und FIFO des Stores bestimmen die Reihenfolge, keine Zeit. Die Fänger sind gestaffelt: `Workspaces.Get` in eigener Transaktion trifft nur Z50; Fenster 2 trifft Z50 und Z51; Fenster 3 trifft alle. Die Messung des Implementers (0 von 50 im Normalstand, 50 von 50 gegen `Put` in der zweiten Transaktion, Detektor 0 von 50 gegen den Zähler) ist nachgemessen und trägt. Zu der Rate gegen den Mutanten *Cursor vor `Sessions.Get`* nenne ich nur einen Lauf, nicht 50.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Der Fänger der Replayfähigkeit führt `REPLAY_REDACTED_BODIES` nur im Modus `sequential`: `Replayfaehig(e, true)` an der Stelle von `strict` (`waehleStrict`) und von `first`/`last` (`waehleErster`) lässt alle Tests grün (Exit 0), während derselbe Mutant in `sequential` `TestReplayfaehigkeit` rot macht. Folge der Lücke: ein Exchange mit Response-Redaction `applied` würde bei `REPLAY_REDACTED_BODIES=false` in drei von vier Strategien ausgeliefert, ohne dass etwas rot wird. | `AGENTS.md` §3.15, `SPEC-135` Nr. 8 und 9, `SPEC-045` | `src/hexagon/services/replay/auswahl.go` · `case !model.Replayfaehig(e, redigierteBodies):` und `if !model.Replayfaehig(e, redigierteBodies) {` | ja — Mutation, `go test` | Fänger führt nur eine Form der Eingabe (Achse Strategie in der Menge der Replayfähigkeit fehlt) |
| F-2 | MEDIUM | Die Zusage „Header-Namen vergleicht der Dienst überall ohne Beachtung der ASCII-Schreibweise“, die Randregel „Leerzeichen und Tabulatoren am Rand“ (`Content-Encoding`) und die Zelle *Methode × fremd* („`get` ist unbekannt“) haben keinen Fänger über die Form: `a.Header["Content-Encoding"]` und `a.Header["Content-Type"]` statt `werteVon` (je Exit 0), `strings.TrimSpace` statt `Trim(t, " \t")` (Exit 0), `strings.ToLower` statt `asciiKlein` (Exit 0; das Kelvin-Zeichen U+212A faltet zu `k`, und `x-recorder-workspace` enthält ein `k`) und `strings.ToUpper(a.Methode)` im Key (Exit 0) bleiben grün; `TestContentEncodingRequest` führt den Namen nur kanonisch und den Rand nur mit Leerzeichen, `TestRequestBodyFormen` führt `Content-Type` nur kanonisch, keine Anfrage trägt eine Methode außer `GET`, `HEAD` und `POST` in Großschreibung. Folge: Quelle der Lücke ist nicht der Code (er ist richtig), sondern der Beweis; ein Header-Name aus einer anderen Quelle als `net/http` (der Namen mit Nicht-ASCII ohnehin abweist) wäre nicht gefangen. | `AGENTS.md` §3.15 und §3.17, `SPEC-135` Kopf und Nr. 7 b, Raster F1 | `src/hexagon/services/replay/key.go` · `strings.Trim(t, " \t")` und `werteVon(a.Header, "content-encoding")`; `src/hexagon/services/replay/dienst.go` · `func asciiKlein(s string) string {` | ja — Mutationen, `go test` | Fänger führt nur die kanonische Form (Schreibweise, Faltung, Rand, Methode) |
| F-3 | MEDIUM | Der Fänger der Fehlerabbildung führt an Stelle S6 (`Cursors.Get`) von den Verbünden mit `ErrNichtGefunden` nur `Join(ErrNichtGefunden, ErrBeschaedigt)`. Der Code entscheidet über die Tabelle von `SPEC-109` Nr. 5 (der erste zutreffende Eintrag), die den Verbund je nach zweitem Fehler verschieden abbildet: vor der Zeile `ErrNichtGefunden` (`ErrFremderService`, `ErrBlob*`, `ErrPfad` und andere) ist es ein Fehler, danach (`ErrUngueltig`, `ErrDatenUngueltig` und andere) der fehlende Cursor. Der Mutant `return ok && !errors.Is(err, outbound.ErrBeschaedigt)` in `fehlt` bleibt grün (Exit 0); `Join(ErrNichtGefunden, ErrFremderService)` würde als neue Replay-Session behandelt statt `ODR-E5004` zu liefern. | `AGENTS.md` §3.15 und §3.17, `SPEC-109` Nr. 5, `SPEC-135` Nr. 10 | `src/hexagon/services/replay/replay.go` · `return ok && c == w`; Test `TestAbbildungJederStelle` · `{"Join(ErrNichtGefunden, ErrBeschaedigt)"` | ja — Mutation, `go test` | Fänger führt nur eine Form der Eingabe (Verbund mit `ErrNichtGefunden`: nur ein zweiter Fehler von vielen) |
| F-4 | MEDIUM | Die Mengen der Konfig führen eine Ausprägung der Achse *Position*: `New` prüft die Dienste und die Pfade je Element, aber der Mutant, der nur `Dienste[0]` prüft, und der Mutant, der nur `Pfade[0]` prüft, bleiben grün (je Exit 0), weil `TestNewKonfig` jeden Verstoß als einzigen Eintrag der Liste führt. Ebenso bleibt „der letzte passende Dienst gewinnt“ (Bedingung „wahr“ statt „länger“) grün, weil `TestDienstWahl` die Dienste aufsteigend nach Länge des Roots konfiguriert; die Zusage „bei mehreren der mit dem längsten `Root`“ hängt dort an der Reihenfolge, die der Test setzt. | `AGENTS.md` §3.15 und §3.17, `SPEC-135` Nr. 1 und 4 | `src/hexagon/services/replay/replay.go` · `slices.ContainsFunc(k.Dienste, func(d Dienst) bool`; `src/hexagon/services/replay/dienst.go` · `(!gefunden \|\| len(d.Root) > len(gewaehlt.Root))` | ja — Mutationen, `go test` | Fänger führt nur eine Form der Eingabe (Position und Reihenfolge in der Menge der Dienste und Pfade) |
| F-5 | MEDIUM | Die Zusage „der Dienst prüft `ctx` nach der Bestätigung nicht mehr“ hat nur an einer von zwei Stellen einen Fänger: der Test beendet `ctx` im `Log` der Telemetrie, also nach dem Eintrag. Der Mutant `if ctx.Err() != nil { _ = strom.Close(); return …, ctx.Err() }` zwischen Bestätigung und `protokolliere` bleibt grün (Exit 0); er ließe einen Exchange verbrauchen und gäbe dem Aufrufer einen Fehler statt der Antwort. | `AGENTS.md` §3.15, `SPEC-135` Nr. 14 | `src/hexagon/services/replay/replay.go` · `protokolliere(ctx, r.k.Telemetrie, e, d.ID, rec, ws, a.KorrelationsID)`; Test `TestAbbruchNachDerAuswahl` · `abbruchTel{u.tel, cancel}` | ja — Mutation, `go test` | Fänger führt nur eine Stelle der Menge (Fenster zwischen Bestätigung und Eintrag) |
| F-6 | MEDIUM | Die Zusage „der Strom von `Blobs.Open` wird unverändert weitergereicht“ (`SPEC-135` Nr. 12) hat keinen Fänger: `Body: struct{ io.ReadCloser }{strom}` (verbirgt `WriterTo`) bleibt grün (Exit 0). Die Hülle `strom` trägt `WriteTo` und zählt es (Z56, `TestHuellen` prüft die Hülle selbst), aber kein Replay-Test fragt `writeTos` oder die Identität der Antwort ab; `io.NopCloser` wird allein über den Zähler von `Close` rot. | `AGENTS.md` §3.15, `SPEC-135` Nr. 12 | `src/hexagon/services/replay/replay.go` · `Body: strom}, nil`; Test `TestStatusUndBodyByteGleich` | ja — Mutation, `go test` | Zusage ohne Fänger (Hülle gebaut, am Dienst nicht benutzt) |
| F-7 | LOW | Der Plan folgt dem Code nicht in zwei Stellen (`AGENTS.md` §3.9): §4 nennt die Messung bis „4263 am Stand `45fc11e`“ und endet dort, die Messung beim Handoff (4715 von 9200) steht nirgends; und das Feld `schnitt-grund` des Blocks sagt weiter „bis dahin lehnt der Rest jeden Request mit Body mit ODR-E5001 ab“, obwohl der Teil Request-Body in `d4ebb7b` geschrieben ist und im Slice bleibt. | `AGENTS.md` §3.9 und §3.14 | `docs/plan/planning/done/slice-replay-service.md` · `bis dahin lehnt der Rest jeden Request mit Body mit ODR-E5001 ab` | ja — `make groessenmarke` nennt 4715 | Plan folgt der Korrektur nicht (Zahl und Zustandssatz) |
| F-8 | LOW | Der Kommentar von `Puffer.Bytes` sagt zu: „ohne Bytes ist das Ergebnis nil“. Der Mutant ohne diesen Zweig bleibt grün (Exit 0); `TestGrenze` fragt das nicht, und kein Test, auch keiner über den Key, unterscheidet `nil` und leer. | `AGENTS.md` §3.11 | `src/hexagon/services/replay/begrenzt/begrenzt.go` · `// Bytes liefert die angenommenen Bytes; ohne Bytes ist das Ergebnis nil.` | ja — Mutation, `go test` | Kommentar sagt zu, was kein Test prüft |
| F-9 | LOW | Randform ohne Ort in `SPEC-135` Nr. 12: ein gewählter Exchange mit Response-Body `discarded`. Der Code behandelt jeden Zustand außer `present` wie `absent` (leerer Strom, kein `Open`), und kein Test fängt es: `if e.ResponseBody != model.BodyAbsent {` bleibt grün. Erreichbar ist die Form nicht: `Exchange.Validate` und die Negativtests des Schemas (`make schema-check`, *Verworfener Response-Body bei complete*) lassen sie bei `complete` nicht zu, und nur `complete` ist ein Kandidat. Der Implementer hat die Zelle gemeldet, §6 trägt sie mit Ausgang *offen* und Adresse Architect (`AGENTS.md` §3.12, §3.13). Bewertung: keine Merge-Blockade, kein Failure-Szenario über einen gültigen Speicher, daher LOW (nicht MEDIUM nach §3.16). **Zuständig: Architect**, entweder die Form in Nr. 12 ausdrücklich benennen oder `ODR-E9001` festlegen; danach schreibt der Implementer den Fänger (ein Exchange, der `Validate` umgeht). | `AGENTS.md` §3.12 und §3.16 | `src/hexagon/services/replay/replay.go` · `if e.ResponseBody == model.BodyPresent {`; `spec/spezifikation.md` · `SPEC-135` Nr. 12 | ja — Mutation, `go test`; Entscheidung: nein | Randform ohne Ort (gemeldet, unerreichbar mit gültigen Daten) |

## Informationen (ohne erwartete Aktion des Implementers)

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| I-1 | INFO | Vier grüne Mutanten sind äquivalent oder tot und kein Befund. (1) `len(id) < 1` zu `< 0` in `replaySessionGueltig`: die leere Kennung ist vorher `einzelwert` abgewiesen. (2) Guard `e.MatchKey != ""` entfernt: der Key der Anfrage ist nie leer, `TestLeererKey` bleibt ohne Mutation, die eine Änderung nur an `gleicherKey` ergäbe (Z24 nennt in Plan und Bericht die Mutation „leerer Key gleich leerem Key“, die als einzelne Änderung am Dienst nicht zu bilden ist). (3) `append(c.Consumed, …)` ohne `Clone`: gleichwertig, solange kein Port eine Liste mit Reserve in der Kapazität liefert. (4) `RequestRedaction` immer `none` (das Ergebnis `rr` fließt nur in `KeyRedacted`, das der Dienst nicht liest): vom Implementer als gleichwertig benannt (Z61), nachgemessen. | Maintainability | `src/hexagon/services/replay/dienst.go` · `if len(id) < 1 \|\| len(id) > 128`; `src/hexagon/services/replay/auswahl.go` · `return e.MatchKey != "" && e.MatchKey == key` | ja — Mutation, `go test` | – |
| I-2 | INFO | `TestKeinUpstream` hält `k.NumField() != 11` fest; der Vertrag sagt, die Konfig wachse additiv (`SPEC-135` Nr. 1), und `slice-telemetry-einbau-gateway` trägt Felder nach: der Test bricht dort an dieser Stelle und gehört dann in dessen Plan. Die Regel „anderer Dienst“ erkennt den eigenen Baum an einem Teilstring (`strings.Contains(pfad, "/services/replay")`), der auch `/services/replayfoo` annähme. `New` klont `Dienste` und `Pfade`, nicht die Listen in `Regeln` (`VaryHeaders`, `VolatileHeaders`). | Maintainability | `src/hexagon/services/replay/konfig_test.go` · `k.NumField() != 11` | nein | – |
| I-3 | INFO | Die vom Implementer gemeldete Verletzung der Betriebsregel mit `sed` ist im Diff nicht abzulesen. Nachgesehen: die geänderten Dateien sind unauffällig (kein Rest einer Ersetzung, `gofmt` und `go vet` grün, `make lint` ohne Cache 0 Befunde). | Maintainability | – | nein | – |
| I-4 | INFO | Sensoren des Reviewers: `make gates` Exit 0 (Datei `gates.rc`: 0); `make a-check` 0 Befunde, `make groessenmarke` „ok, Messung 4715 von Marke 9200“, `make docs-check` 0 Befunde. Die gecachten Stufen einzeln ohne Cache nachgefahren (`docker build --no-cache-filter <stufe>`): `test` („pruefung test: ok“), `race` („pruefung race: ok“), `lint` („0 issues“), alle Exit 0. `make gegenproben` und der Stempel nicht berührt. | Maintainability | – | ja — `make gates` | – |
| I-5 | INFO | Rohbytes (`AGENTS.md` §3.18): das Verfahren `env LC_ALL=C grep -aP` ist an `docs/reviews/2026-10-25-slice-baseline-import-verifikation-3.md` Zeile 45 bestätigt (die Zeile trägt `C2 A0`). In `src/hexagon/services/replay/`, `src/hexagon/ports/inbound/replay.go`, dem Plan und dem Diff von `spec/spezifikation.md` findet sich weder ein BOM noch ein NBSP, ein Zeichen der Klasse Zero-Width oder ein Steuerzeichen außer dem Zeilenende; die Fixtures mit `0xFF`, BOM und CRLF entstehen im Test aus Bytewerten. | `AGENTS.md` §3.18 | – | ja — `env LC_ALL=C grep -aP` | – |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/inbound/replay.go` | geprüft, ohne Befund (eine Methode, Signatur gegen `SPEC-135` und `SPEC-121`; der Kommentar trägt Zusage und Verweis) |
| `src/hexagon/services/replay/` Produktcode (`replay.go`, `dienst.go`, `auswahl.go`, `antwort.go`, `key.go`, `fehler.go`) | geprüft; Ablauf, Rangfolge (a bis h), Zuordnung, Auswahl, Cursor, Auslieferung, Header-Policy, Protokoll und Fehlerabbildung gegen `SPEC-135` ohne Abweichung; 93 von 112 Mutanten rot, die 19 grünen sind F-1 bis F-6, F-8, F-9 und I-1; der Code selbst weicht an keiner mutierten Stelle von der Spezifikation ab |
| `src/hexagon/services/replay/begrenzt/` | geprüft; Befund F-8; die Grenze (`<=`, Rest über der Grenze) ist rot gesehen |
| Tests und Fakes (`*_test.go`) | geprüft; Befunde F-1 bis F-6; die Fakes sind stichprobenhaft mutiert (`Blobs.Verify` und `Cursors.List` zählen nicht, Rollback fehlt: je rot); `TestFakesZaehlenJedeOperation` läuft über alle Operationen der Ports |
| Nebenläufigkeit (Z46 bis Z53, Z54, Z63) | geprüft, ohne Befund (Raten 50 von 50 in Prozessen mit und ohne Detektor; Ordnung durch Kanal, Haltepunkt und FIFO; kein `sleep` ordnet Ereignisse, nur ein Warten auf den Zustand der Schlange mit Frist 10 s) |
| Schichten und Imports (`.a-check.yml`) | geprüft, ohne Befund (`make a-check` 0 Befunde; Import eines Adapters im Paket gibt `app-impurity` mit Exit 1; kein `net/http`, kein anderer Dienst) |
| Lint-Ausnahmen, `//nolint`, Gates | geprüft, ohne Befund (`.golangci.yml`, `.a-check.yml`, `Makefile` unverändert; keine Direktive; `make lint` ohne Cache 0 Befunde) |
| `spec/spezifikation.md` | geprüft, ohne Befund (der Vertrag `SPEC-135` und die Nachzüge stehen vor dem ersten Code-Commit; kein Code-Commit ändert `spec/`; keine Aussage widerspricht dem Lastenheft; die Zeile zu `Location` mit Marker folgt der Entscheidung des Owners und ist in §6 offen geführt) |
| `spec/architecture.md`, `harness/`, `AGENTS.md`, `.claude/`, `.harness/` | geprüft, ohne Befund (im Diff unverändert; `AGENTS.md` §3.4 nicht berührt) |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft; Produktcode ohne Konjunktiv über verworfene Alternativen und ohne abwesenden Text; Befund F-8 (Zusage ohne Test) |
| `slice-replay-service` (Plan) | geprüft; Befund F-7; die Sendungen 1 bis 6 und die Randform stehen in §6 mit Ausgang *offen* und Adresse (`AGENTS.md` §3.13); die Tabelle und der Absatz *Mutationen und Raten* stimmen mit den Tests überein (Zählprüfungen in den Tests gesehen: 18 und 5, 22, 135, 939, 3510) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 6 |
| LOW | 3 |
| INFO | 5 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe (Achse Strategie · Schreibweise, Faltung, Rand und Methode · Verbund mit `ErrNichtGefunden` · Position in der Menge) · Fänger führt nur eine Stelle des Fensters · Zusage ohne Fänger (unverändert weitergereicht) · Plan folgt der Korrektur nicht · Kommentar sagt zu, was kein Test prüft · Randform ohne Ort.

## Verdikt

**Merge-blockierend:** ja, wegen der sechs MEDIUM; kein HIGH. Der Produktcode weicht in keiner gemessenen Klasse von `SPEC-135` ab (kein Mutant am Code ist als Fehlverhalten grün geblieben; alle sechs MEDIUM sind Lücken des Beweises, keine Fehler des Verhaltens). Die nebenläufigen Zusagen sind gut gefangen (Raten 50 von 50, Normalstand 0 von 50).

**Übergabe:** F-1 bis F-8 gehen an den Implementer (je ein Fänger über die Klasse: F-1 als Schleife über Strategie × `RedigierteBodies` × Redaction; F-2 über Namen, Faltung, Rand und Methode; F-3 über die Verbünde mit `ErrNichtGefunden`; F-4 über Position und Reihenfolge; F-5 mit einem Haltepunkt zwischen Bestätigung und Eintrag; F-6 über die Identität des Stroms; F-7 und F-8 als Nachzug). F-9 geht an den Architect (Nr. 12). Die Finding-Klassen gehen in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ersetzt keine Verifikation: DoD-Konformität, `make gates` und die Prüfung der Sendungen gegen die Nehmer prüft der Verifier.
