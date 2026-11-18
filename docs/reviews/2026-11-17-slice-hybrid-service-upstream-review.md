# Review-Report: slice-hybrid-service-upstream (Teil 2: Upstream-Fallback) — 2026-11-17

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers)

**Gegenstand:** `slice-hybrid-service-upstream`, Zweig `slice-hybrid-service-upstream`, `75249f56..30d9130a` (`git diff main...HEAD`); Pfade `src/hexagon/services/hybrid/{verteiler,puffer,antwort,fehler,upstream}.go` samt Tests (`upstream_test.go`, `upstream_hilfen_test.go`, `referenz_test.go` und die angepassten Tests von Teil 1), Plan-Abschnitte §1, §3 (Tabelle *Zusage · Fänger*, Bericht, akzeptierte Negative) und §6 (Raster, offene Zeilen).

**Skill:** `.harness/skills/reviewer.md` @ `30d9130a` <!-- d-check:ignore (Zitat des Standes im Lauf) -->
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-17

**Eingangs-Kontext** (die Verträge, gegen die geprüft wurde):

- Slice-Plan `slice-hybrid-service-upstream` (§1 bis §6) und, für das Raster, `slice-hybrid-service` §6
- `SPEC-284` Nr. 1 bis 13 (Teile U und H6), `SPEC-261` Nr. 4, 5 c, 10, `SPEC-120`, `SPEC-121` Nr. 4, 7
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-FA-12`, `LH-FA-04`
- `AGENTS.md` §3.2, §3.6, §3.7, §3.9 bis §3.18

**Vorgehen und Belege.** Gelesen: der gesamte Diff des Produktcodes (`upstream.go` 137 Zeilen, Änderungen an `verteiler.go`, `puffer.go`, `antwort.go`, `fehler.go`), alle Tests des Pakets mit Änderung und der Plan. Gefahren, unter `flock /tmp/odata-recorder-heavy.lock`, Image `odata-recorder:race`, `--network none`, `--memory=4g --memory-swap=4g --cpus=4`, ein Container je Lauf, Messender: der Reviewer, Last des Hosts 4,1 bis 5,3 von 4 (aus der Ausgabe von `tools/harness/mutationslauf.sh`):

1. Die Tabelle *Zusage · Fänger* aus §3 des Plans, Zeile für Zeile, mit eigenen Mutationen, nicht aus dem Bericht des Implementers übernommen: 66 Mutationen in vier Listen über `tools/harness/mutationslauf.sh` (Klasse `andere`, ein Prozess je Stufe, deterministisch; je Mutation prüft das Werkzeug, dass `alt` genau einmal trifft). Ergebnis: 51 rot, 1 rot-paket (nur durch eine Nebenwirkung, siehe F-5), 12 grün (F-1 bis F-7 und das benannte Negativ `s09`, F-10), 2 Übersetzungsfehler (meine Mutationen `r33` und `s01`, mit einer übersetzenden Bedingung als `t01` wiederholt und rot).
2. Das Raster aus §6 von `slice-hybrid-service` (Zeile *Upstream-Fallback*, *Schließen der Ströme*, Spalte U der Zusatztabelle) mit je einer Mutation je Zelle, die das Raster nennt: fehlender Body, fehlende und leere `UpstreamURL`, fremder Host, leere Query, leerer Header-Wert, Fehler des Upstream, Kontextfehler, Wiederholen, Antwort zugleich mit Fehler.
3. Ein eigenes Probe-Paket `probe_rv_test.go` (nur in einer Kopie des Baums im Scratchpad, nicht im Zweig) für die Dienstwahl gegen ein unabhängiges Orakel: 63 Mengen von Roots (Teilmengen aus `/`, `/a`, `/a/b`, `/ab`, `/A`, `/a/b/c`) × 20 Pfade (leer, ohne führenden `/`, `*`, Grenzfälle der Segmentgrenze, Groß- und Kleinschreibung, Escape) = 1260 Fälle. Der Code besteht es unverändert; die Lücken liegen im Fänger, nicht im Code.
4. `go test -race -count=20` über `./src/hexagon/services/hybrid/...` und fünfmal `go test -race -count=1 -run 'TestGleichzeitig|TestUpstream'`: grün (hybrid 9,0 s, policy 98,8 s; 5 von 5 Prozessen grün).
5. Nebenläufigkeit: Mutation s15 (Schreibzugriff auf `v.dienste` in `waehle`, Klasse `nebenläufig`): mit Detektor 30 von 30 Prozessen rot, ohne Detektor nicht gemessen (`n_ohne` 0).
6. Rohbytes: `LC_ALL=C grep -nP` auf Backslash, `u` und vier Hex-Ziffern über `src/hexagon/services/hybrid/*.go` und den Plan: kein Treffer; Zeichen jenseits von ASCII in den neuen Dateien sind nur Umlaute und typografische Anführungszeichen in Kommentaren. Dieser Report ist auf demselben Weg gegengelesen.

Nicht gefahren: `make gates` (der Implementer meldet sie; die Frage der DoD gehört dem Verifier).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Zusage „U erhält einen eigenen, vollständigen Leser über den Puffer“ (`SPEC-284` Nr. 5 c, 8 b; Plan Z12 (U), Z24) ist nur für den Weg nach M gefangen. Die Mutation „U erhält den Leser der Stufe R“ (`st.upstream = st.replay` nach der Entscheidung `Upstream` im Replay-Miss) bleibt grün: kein Fall führt U nach R mit einem Body (die Fälle des Weges `nach R` sind alle `GET` ohne Body), und ein `GET` mit Body erreicht U nach einem Miss. Ein Leser, den R schon gelesen hat, ginge als leerer oder verkürzter Body an den Upstream, mit `Laenge` des vollen Puffers. | `AGENTS.md` §3.15 (ein Fall je Aufrufstelle des Ports) | `verteiler.go` · `Code: c}) == policy.Upstream {`; Test `upstream_test.go` · `"nach R": func() *rig {` (ohne Body); `TestUpstreamProbe` und `TestUpstreamArgumente` (nur Weg nach M) | ja — Mutation `r04-ur-nach-r-reader-verbraucht`, Suite grün | Fänger führt nur eine Aufrufstelle (Weg nach M, nicht nach R) |
| F-2 | MEDIUM | Die Zusage „kein Dienst: `Ziel` leer, U 0“ (`SPEC-284` Nr. 4, Plan Z19) ist für den Root `/` nur mit führendem `/` im Pfad gefangen. Die Mutation `return pfad, true` in `unterRoot` bleibt grün: kein Fall nennt bei einem Dienst mit Root `/` einen Pfad ohne führenden `/` (leer, `odata`, `*`). Mit der Mutation erhielte ein solcher Request `Ziel` und OData-Pfad `/` des Dienstes statt keines Dienstes (die Spezifikation von `SPEC-135` Nr. 4 und `SPEC-121` Nr. 2 b nennt den führenden `/`). | `AGENTS.md` §3.15 (beide Seiten jeder Grenze) | `upstream.go` · `return pfad, strings.HasPrefix(pfad, "/")`; Test `upstream_test.go` · `func TestDienstwahl` (der Fall ohne Dienst nennt nur einen Dienst mit Root `/odata/svc`) | ja — Mutation `r01-root-slash-ohne-praefix`, Suite grün | Fänger führt nur eine Seite der Grenze (Pfad mit führendem `/`) |
| F-3 | MEDIUM | Die Zusage „`Ziel` ist die `UpstreamURL` des gewählten Dienstes“ (`SPEC-284` Nr. 4; Raster *fehlende UpstreamURL: Nr. 4*) ist nur für Dienste mit nichtleerer `UpstreamURL` gefangen. Die Mutationen „nur Dienste mit `UpstreamURL` zur Wahl zulassen“ und „leere `UpstreamURL` des gewählten Dienstes durch die des ersten Dienstes ersetzen“ bleiben grün; jede Liste in jedem Test führt nur Dienste mit URL (Ausnahme: ein einziger Dienst ohne URL). Folge: ein Dienst ohne `UPSTREAM_URL` könnte über den Weg einer Liste mit mehreren Diensten das Ziel eines anderen erhalten, ohne dass etwas rot wird. | `AGENTS.md` §3.15 (ein Fall je Form der Eingabe), §3.16 (Zelle *leerer Wert*) | `upstream.go` · `return v.dienste[gewaehlt].UpstreamURL, rest`; Test `upstream_test.go` · `func TestDienstwahl` (`dienste := []hybrid.Dienst{` ohne leere URL) | ja — Mutationen `u01-wahl-nur-mit-url`, `u02-ziel-fallback-erster-dienst`, Suite grün | Fänger führt nur eine Form der Eingabe (nichtleere `UpstreamURL`) |
| F-4 | MEDIUM | Die Zusage Z10 (U) „ist `ctx` beendet, ist das Ergebnis `ctx.Err()`“ (`SPEC-284` Nr. 3 e) ist nur für `context.Canceled` gefangen. Die Mutationen „`return … context.Canceled`“ und „Prüfung nur auf `ctx.Err() == context.Canceled`“ vor U bleiben grün, weil beide Fälle von `TestUpstreamAbbruch` einen Kontext mit `WithCancel` nutzen und keiner einen mit Frist; dieselben Mutationen vor R (Teil 1, `TestAbbruchZwischenStufen`) sind rot. Mit der Mutation lieferte ein Request mit abgelaufener Frist einen Abbruch statt der Frist, oder riefe U mit abgelaufener Frist. | `AGENTS.md` §3.15 (ein Fall je Element einer Menge: beide Kontextfehler) | `upstream.go` · `return inbound.Antwort{}, ctxErr`; Test `upstream_test.go` · `func TestUpstreamAbbruch` · `abbruchKontext()` | ja — Mutationen `r02-ctx-konstante-canceled`, `t02-u-ctx-nur-canceled`, Suite grün (Gegenprobe vor R: `s02`, `t03` rot) | Fänger führt nur eine Form der Eingabe (`Canceled` ohne `DeadlineExceeded`) |
| F-5 | MEDIUM | Zusage zur Nebenläufigkeit (`SPEC-284` Nr. 13: der Dienst hält nach `New` nur seine Konfig, die er nicht ändert) ist für den neuen Weg U nicht gefangen. Der Plan nennt in §3 `TestVerteilerFelder`, `TestQuelltext` und `TestGleichzeitig` als Fänger („ändert dieser Slice nicht“), aber der Slice führt das Feld `dienste` ein, das jeder Weg liest. Die Mutation „Schreibzugriff `v.dienste[0].UpstreamURL = ziel` in `upstreamFallback`“ bleibt grün: `TestQuelltext` erkennt keinen Schreibzugriff über einen Indexausdruck auf ein Feld des Empfängers, und `TestGleichzeitig` läuft nie durch `upstreamFallback` (kein Fall mit U, 0 von 64 Aufrufen). Der Detektor sähe das Rennen nur, wo ein Lauf den Weg gleichzeitig führt. Ein Schreibzugriff im Weg `waehle` ist gefangen (30 von 30 Prozessen mit Detektor, Mutation `s15`). Ein Negativ mit Grund und Folge steht nicht im Plan. | `AGENTS.md` §3.10 (Nebenläufigkeit: „in jedem Lauf rot“, Zugriff, den kein Lauf ausführt), §3.15 | `upstream.go` · `v.upstream.Senden(ctx, outbound.UpstreamAnfrage{`; Test `nebenlaeufig_test.go` · `func TestGleichzeitig`; Plan §3 · `ändert dieser Slice nicht: kein neues Feld, das ein Aufruf schreibt` | ja — Mutation `s16-upstream-schreibt-dienste` (Klasse `andere`; rot-paket nur für `s17` und nur durch eine Nebenwirkung auf `TestDienstwahl`) | Zusage über Nebenläufigkeit ohne Fänger für den neuen Weg und ohne Negativ |
| F-6 | MEDIUM | Randform im Code entschieden: ein Fehler von `Senden`, der `ErrUpstreamZeit` und `ErrUpstreamNichtErreichbar` zugleich enthält, ergibt `ODR-E4001` mit 504, weil der Code `ErrUpstreamZeit` zuerst prüft; `SPEC-284` Nr. 11 nennt die Rangfolge nicht, und kein Fänger führt die Form (die Mutation „`ErrUpstreamZeit` nur ohne `ErrUpstreamNichtErreichbar`“ bleibt grün). Der Implementer meldet die Zelle in §6 mit Ausgang *offen* und Adresse Architect, hat aber nicht angehalten (`AGENTS.md` §3.12). | `AGENTS.md` §3.12, §3.16 (eine im Code entschiedene Zelle ist ein MEDIUM) | `upstream.go` · `case errors.Is(err, outbound.ErrUpstreamZeit):`; Plan §6 · `Randform ohne Ort in SPEC-284 Nr. 11` | ja — Mutation `r42-zeit-nach-nichterreichbar`, Suite grün | Randform ohne Ort in der Spezifikation (Adresse Architect, offen) |
| F-7 | LOW | Die Zusagen „an Kommas geteilt, Leerzeichen und Tabulatoren am Rand entfernt“ und „ohne Beachtung der ASCII-Schreibweise“ (`SPEC-284` Nr. 8 c) sind für den Rand jenseits der Menge ungefangen: `strings.TrimSpace` statt `Trim(t, " \t")` (kürzt auch Unicode-Leerraum wie NBSP), `strings.ToLower` statt `asciiKlein` (faltet Zeichen jenseits von ASCII, etwa U+212A) für den Namen und für `Connection`: drei Mutationen grün; `TestHeaderFilter` führt nur Leerzeichen und Tabulator als Randzeichen und nur ASCII-Namen. Ein Name, der erst durch die Mutation einem Namen der Liste gleicht, würde fälschlich entfernt; der HTTP-Adapter lässt Namen jenseits von ASCII nicht zu, der Weg über Randzeichen im Wert von `Connection` ist schmal. | `AGENTS.md` §3.15 (beide Seiten jeder Grenze) | `upstream.go` · `strings.Trim(t, " \t")` und `k := asciiKlein(n); !hopByHop(k)` | ja — Mutationen `r03-trimspace`, `r36-tolower-name`, `r37-tolower-connection`, Suite grün | Fänger führt nur eine Seite der Grenze (Randzeichen und Namen innerhalb von ASCII) |
| F-8 | LOW | Zustand im Plan nicht nachgezogen: §6 nennt die Messung der Pfade des Blocks mit 1448 Zeilen (1399 und 49); `git diff --numstat 75249f56..30d9130a` über dieselben Pfade ergibt 1455 (1406 und 49), nach dem Commit `gofmt` und den Lint-Korrekturen. Die Zeilen des Plans (`Bericht`, §3) kamen dazu in einem eigenen Commit (`a16d8cae`) nach den drei Code-Commits, nicht im selben Commit wie der Code. Die Marke 3000 liegt weit entfernt. | `AGENTS.md` §3.9 (Plan folgt jeder Korrektur im selben Commit), §3.14 | `docs/plan/planning/done/slice-hybrid-service-upstream.md` · `gemessen nach dem Code: Summe aus Einfügungen und Löschungen` | ja — `git diff --numstat` | Zahl im Plan nicht nachgezogen; Plan nicht im Commit des Codes |
| F-9 | INFO | Größenmarke: der Block trägt `basis`, `schaetzung` und `marke` nicht; §4 nennt „Marke: noch nicht gesetzt, vorläufig 3000“ und die Bedingung („ergibt *S* ≥ 2223, wird vor dem Code geschnitten“) ist vor dem ersten Code-Commit nicht geprüft worden (§6, Ausgang *offen*, Adresse Planner, vom Implementer ehrlich gemeldet). Nach dem Code liegt die Messung bei 1455 Zeilen, das Ergebnis ändert sich dadurch nicht. Zuständig: Planner. | `AGENTS.md` §3.14 | Plan §4 · `Marke: noch nicht gesetzt, vorläufig 3000` | ja — `make groessenmarke` (grün mit Hinweis, nach Angabe des Implementers; nicht gefahren) | Marke nicht vor dem ersten Code-Commit gesetzt (offen, Adresse Planner) |
| F-10 | INFO | Akzeptierte Negative (a) und (b) bestätigt, ihr Grund trägt: `Pfad` an H5 (Mutation `s09` grün), ein einzelner Aufruf der Probe mit leerem `Ziel` oder ohne `ODR-E5001` ändern das Ergebnis von `Entscheide` nie, denn die Probe ist eine Oder-Verknüpfung über vier Codes und `*policy.Policy` ist ein konkreter Typ. Die benannten Fänger (`Ziel` an allen vier Aufrufen, an H3 und H5; `Pfad` an H3 und in der Probe) sind rot (`r30`, `r31`, `r32`, `s08`, `s10`). Dasselbe gilt für Negativ Z35: der Dienst importiert keinen Adapter, der Prüf-Client ist kein Fänger für `httpupstream`; der Fänger ist der Verbindungszähler am Binary in `slice-bootstrap-m4`. | `AGENTS.md` §3.15 | Plan §3 · `Akzeptierte Negative (Grund und Folge)` | nein | — |
| F-11 | INFO | Die Plan-Zeile „Z19 Mutationen je Stelle sind teils äquivalent“ trägt Grund und Folge und hat die Adresse Planner (§6, Ausgang *offen*); die Zählprüfungen der Schleifen (`224`, `20`, `3·7`, `10`, `5·9`) stimmen mit den Mengen überein. Die Kennzeichnung der Antwort und die Fehlerabbildung (Z25, Z26 (U), Z27, Z28 (U)) hielten alle meine Mutationen (Status, Header, Ursache, Schließen). | `AGENTS.md` §3.13 | Plan §6 · `Z19: Mutationen je Stelle sind teils äquivalent` | nein | — |

---

## Auflösung der Tabelle *Zusage · Fänger* (§3) mit eigenen Mutationen

„rot“ heißt `--- FAIL` des benannten Tests (nie `panic:` oder Zeitlimit). Namen in Klammern verweisen auf meine Listen.

| Zeile | Eigene Mutation | Ergebnis |
|---|---|---|
| Z01 (U) | Prüfung `Upstream == nil` gestrichen (`r22`) | rot `TestNew` |
| Z01 (D) | Liste nicht kopiert (`r23`) | rot `TestZ01DienstwahlBleibt` |
| Z07 | H3 mit `GET` statt `Methode` (`s06`) | rot `TestUpstreamNachMock` |
| Z08 (U) | H5 mit `GET` statt `Methode` (`s07`); Bedingung `c == ODR-E5001` immer wahr (`t01`) | rot `TestUpstreamNachReplayMiss`, `TestUpstreamNurNachE5001` |
| Z09 (U) | Code der Bedingung immer wahr (`t01`) | rot `TestUpstreamNurNachE5001` |
| Z10 (U) | Prüfung vor U entfernt (`r35`); konstanter Fehler `Canceled` (`r02`); nur `Canceled` geprüft (`t02`) | rot; **grün; grün (F-4)** |
| Z11 (U) | Kontextfehler nur mit `==` erkannt (`s11`) | rot `TestUpstreamKontext` |
| Z12 (U) | `Ziel` der Probe leer, alle vier Aufrufe (`r32`); `Pfad` der Probe leer (`s10`) | rot `TestUpstreamProbe` |
| Z19 | erster statt längster Root (`r19`), Segmentgrenze (`s12`), `pfad == root` (`s13`), Rest leer (`r20`), voller Pfad als OData-Pfad (`r21`), `Ziel` des ersten Dienstes (`s18`); `Ziel` an H3, H5 leer (`r30`, `r31`); `Pfad` an H3 (`s08`) | rot `TestDienstwahl`, `TestZielAnStellen` |
| Z19 | Root `/` ohne Prüfung auf führenden `/` (`r01`); leere `UpstreamURL` des gewählten Dienstes (`u01`, `u02`) | **grün; grün (F-2, F-3)** |
| Z20 | `Methode` fest `GET` (`s03`) | rot `TestHostFaenger` |
| Z21 | keine eigene Mutation allein an der Sicherheitsachse gefahren; die Mutationen `s06`, `s07`, `r16` bis `r18` sind in `TestUpstreamNachMock`, `TestUpstreamNachReplayMiss`, `TestMethodenHeaderNieAmUpstream` rot, ob `TestKeineAnfrageOhneFreigabe` sie einzeln fängt, ist nicht gemessen | – |
| Z22 | je Name der drei Header aus der Liste entfernt (`r16`, `r17`, `r18`) | rot `TestMethodenHeaderNieAmUpstream` |
| Z23 | ungefiltert (`r05`); nur Hop-by-Hop (`r06`); `Trailer` fehlt (`r11`); Präfix ohne Bindestrich (`r12`); `Expect`, `Host`, `Content-Length` fehlen (`r13` bis `r15`); `Accept-Encoding` entfernt (`s14`); Wertliste geteilt (`r38`); Token von `Connection` nicht klein (`r10`); nur erstes Token (`r39`); nur erster Wert (`r40`) | rot `TestHeaderFilter` |
| Z23 | `TrimSpace`; `ToLower` für Namen und `Connection` (`r03`, `r36`, `r37`) | **grün (F-7)** |
| Z24 | `RawQuery` leer (`r07`); `Body` und `Laenge` der Anfrage statt des Puffers (`s04`, `s05`); Leser von M geteilt (`r27`); Body bei `Laenge` 0 (`r28`, `r29`); Leser von R (`r04`) | rot `TestUpstreamArgumente`, `TestUpstreamProbe`; **grün für `r04` (F-1)** |
| Z25 | 504 und 502 (`r24`, `r25`); `Ursache` fehlt (`r41`) | rot `TestUpstreamFehler` |
| Z25 | `ErrUpstreamZeit` und `ErrUpstreamNichtErreichbar` zugleich (`r42`) | **grün (F-6)** |
| Z26 (U) | Kennzeichnung `replay` statt `upstream` (`r34`) | rot `TestUpstreamKennzeichnung` |
| Z27 | Header roh (`r08`); `Status` fest (`r09`); Body `nil` (`r43`) | rot `TestUpstreamAntwort` |
| Z28 (U) | `verwirf` entfernt (`r26`) | rot `TestUpstreamBodySchliessen` |
| Z31/Z32, Nr. 13 | Schreibzugriff in `waehle` (`s15`); Schreibzugriff in `upstreamFallback` (`s16`) | rot 30 von 30 Prozessen mit Detektor; **grün (F-5)** |
| Z35 | akzeptiertes Negativ mit Grund und Folge | F-10 |

## Auflösung des Rasters (§6 von `slice-hybrid-service`, Zeile *Upstream-Fallback* und *Schließen der Ströme*)

| Zelle | Mutation | Ergebnis |
|---|---|---|
| fehlender Body, `Nr. 5 b`, `8 b` | `ohneBody` weitergereicht (`r28`, `r29`) | rot |
| fehlende `UpstreamURL`, `Nr. 4` | Wahl nur mit URL; Ziel des ersten Dienstes (`u01`, `u02`) | grün (F-3) |
| leere Query, `Nr. 8 a` | `RawQuery` geändert (`r07`) | rot (beide Seiten: leer und kodiert) |
| leerer Header-Wert, `Nr. 8 c` | Wertlisten geteilt, `X-Empty` entfernt | rot (`r38`, Fixture `"X-Empty": {""}`) |
| fremder Host, `SPEC-261` Nr. 5 | `Ziel` des anderen Dienstes (`s18`) | rot `TestDienstwahl` (`s18`); der Fall „Verbindungen 0 bei fremdem Host“ ist gelesen, nicht mutiert |
| Fehler des Upstream, `Nr. 11` | Abbildung, Ursache, Rangfolge (`r24`, `r25`, `r41`, `r42`) | rot, bis auf die Randform (F-6) |
| Abbruch, `Nr. 3 e`, `Nr. 11` | Prüfung vor U, Form des Kontextfehlers (`r35`, `r02`, `t02`, `s11`) | rot, bis auf die Frist (F-4) |
| Wiederholung, `Nr. 8` | kein zweiter `Senden`, kein R nach U (Zählprüfung `TestUpstreamFehler`) | nicht mutiert; die Zählung `g.u.zahl() != 1` und `g.r.zahl()` in `TestUpstreamFehler` ist gelesen |
| Antwort zugleich mit Fehler, `Nr. 10` | `verwirf` fehlt (`r26`) | rot |

Zusatztabelle *Ergebnis des Mock × Zustand der Policy*, Spalte U: die Zeilen *Coverage*, *nicht unterstützte Operation*, *Replay-Miss* und *anderer Replay-Fehler* sind durch `TestUpstreamNachMock`, `TestUpstreamNachReplayMiss` und `TestUpstreamNurNachE5001` gegen ein unabhängiges Orakel gefangen (`s06`, `s07`, `t01` rot).

## Prüfung gegen die Klasse der Zusage (`AGENTS.md` §3.17)

| Zusage | Menge (Wortlaut der Spezifikation) | Fänger als Schleife? | Mutation an einem Element, das der Plan nicht nannte |
|---|---|---|---|
| Header-Filter (`Nr. 8 c`) | 21 Namen in 4 Schreibweisen, Tokens von `Connection` (3 Werte, 3 Tokens, Leerraum) | ja, Zählprüfung `n >= 60` | `Trailer` (`r11`), `Accept-Encoding` (`s14`), nur erster Wert von `Connection` (`r40`): rot; Rand jenseits der Menge: grün (F-7) |
| Dienstwahl (`Nr. 4`) | Roots × Pfade × zwei Reihenfolgen | ja, Zählprüfung 20 | `/`-Root ohne führenden `/` und leere `UpstreamURL`: grün (F-2, F-3); das eigene Probe-Paket (1260 Fälle) besteht |
| `Ziel` an den sieben Stellen | 7 Formen × 3 Wege | ja | je Stelle `H3`, `H5`, Probe: rot; einzelner Aufruf der Probe: akzeptiertes Negativ (F-10) |
| Kontextfehler (`Nr. 2`, `3 e`) | `Canceled`, `DeadlineExceeded`, verpackt, `Join` an H6; an U nur `Canceled` vor U | an H6 ja, vor U nein | F-4 |
| Aufrufstellen von U | Weg nach M, Weg nach R | `wege()` für `Kontext`, `Fehler`, `Kennzeichnung`, `Antwort`, `BodySchliessen`; für Body und Leser nein | F-1 |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/services/hybrid/upstream.go` (Fehlerabbildung, Antwort, Kennzeichnung, Schließen) | geprüft, ohne Befund, außer F-4 bis F-7 |
| `src/hexagon/services/hybrid/verteiler.go`, `puffer.go` (Ablauf, Probe, Puffer) | geprüft, ohne Befund, außer F-1 |
| `src/hexagon/services/hybrid/antwort.go`, `fehler.go` | geprüft, ohne Befund |
| `src/hexagon/services/hybrid/` Tests (`upstream_test.go`, `upstream_hilfen_test.go`, `referenz_test.go`, `konfig_test.go`, `klassen_test.go`, `quelltext_test.go`, `hilfen_test.go`) | geprüft; Kopfzeilen nennen, was die Tests führen (`AGENTS.md` §3.11), Fakes zählen jede Operation und halten jedes Argument; Befunde F-1 bis F-5, F-7 |
| Kommentare in Code und Tests | geprüft, ohne Befund (keine verworfene Alternative, kein abwesender Text, keine Chronik) |
| `AGENTS.md` §3.2 (Suppression) | geprüft, ohne Befund (keine `//nolint`-Direktive im Diff) |
| `AGENTS.md` §3.4 (Architektur meilensteinfrei) | geprüft, ohne Befund (`spec/architecture.md` nicht berührt) |
| `AGENTS.md` §3.18 (Rohbytes) | geprüft, ohne Befund (kein Wert mit Escape-Folge, Dateien des Pakets gegengelesen) |
| Sicherheitsachse „keine Anfrage ohne Freigabe erreicht den Upstream“ (`LH-FA-12.b`) | geprüft, ohne Befund: jede Mutation, die einen Zweig zu U öffnet (`Bleiben` als `Upstream`, Methode aus Override-Header, H3 und H5 mit anderer Methode) ist rot |
| Geheimnisse in Senke (`LH-QA-04`) | geprüft, ohne Befund: `Error()` nennt nur Code und `message`, `Ursache` bleibt über `errors.Is` erreichbar (`TestUpstreamFehler`, `r41`) |
| Spec-Straten (`SPEC-284`, `SPEC-261`, `SPEC-120`, `SPEC-121`) | geprüft, ohne Befund; keine Aussage des Diffs widerspricht dem Lastenheft |
| `docs/plan/planning/done/slice-hybrid-service.md`, `welle-hybrid.md` | geprüft, ohne Befund (reine Nachzüge der Verweise `open/` nach `in-progress/`) |
| `docs/plan/planning/in-progress/roadmap.md` | nicht gemeldet: die vier Befunde von `make docs-check` sind bekannt, die Roadmap schreibt nur der Integrationsschritt |

## Summary

0 HIGH, 6 MEDIUM, 2 LOW, 3 INFO.

Der Code stimmt mit `SPEC-284` überein; am Code selbst fand ich kein Fehlverhalten, und das eigene Probe-Paket zur Dienstwahl (1260 Fälle) besteht unverändert. Die Befunde liegen im Fänger: F-1 bis F-4 sind Formen der Eingabe, die der Code verschieden behandelt und die ein Fall nicht führt (der Weg nach R mit Body, der Root `/` ohne führenden `/`, ein Dienst ohne `UpstreamURL`, eine Frist statt eines Abbruchs). F-5 ist eine Zusage zur Nebenläufigkeit ohne Fänger für den neuen Weg und ohne Negativ im Plan. F-6 ist eine im Code entschiedene Randform, die der Implementer in §6 mit Adresse Architect gemeldet hat. Alle Mutationen sind deterministisch (Klasse `andere`, ein Prozess), eine Rate gibt es nicht; die Last des Hosts lag bei 4,1 bis 5,3 von 4.

**Konflikt-Pfad:** kein HIGH, kein Rollen-Widerspruch bisher; die MEDIUM sind Annahme oder Begründung, nicht Sequenz. F-6 liegt bereits beim Architect (`AGENTS.md` §3.13).
