# Review-Report: slice-upstream-adapter — 2026-10-21

**Review-Art:** Code — geprüft gegen Plan, Spezifikation (`SPEC-120` Kopf und Nr. 1 ohne `CAPEM`, Nr. 2 bis 6, 8, 9 und *Ungeprüft*; `SPEC-110` Nr. 7 und Nr. 12), Entscheidungen des Architect (`5947116`) und Hard Rules (`AGENTS.md` §3.7, §3.10 bis §3.16 in der Fassung von `482fdcb`), nicht gegen die DoD (Verifier).

**Gegenstand:** `slice-upstream-adapter` (Welle `welle-record-replay`), `git diff f76ba07..b5c9dd2 -- src/`: das Paket `src/adapters/driven/httpupstream/` und die Prüfung des Hosts in `src/bootstrap/` (A19). Der TLS-Teil (`0716718`) gehört zu `slice-upstream-tls` und ist nicht Gegenstand; geprüft ist nur, dass sein Revert `b6d5270` sauber ist. Summe aus `git diff --numstat f76ba07..b5c9dd2 -- src/`: 2567 (Wert aus §6 des Plans bestätigt).

**Skill:** `.harness/skills/reviewer.md` @ HEAD `9909bf6` · **Modell:** claude-opus-5-5 · **Datum:** 2026-10-21

**Methode.** Jede Zeile der Tabelle *Zusage · Fänger* (A1 bis A19 ohne A10 und ohne die TLS-Teile von A1, A11, A14), jede Zelle des Rasters und jede Zeile der Zusatztabelle ist mit mindestens einer eigenen Mutation aufgelöst; die Mutationen des Implementers sind nicht übernommen. 160 Mutationen liefen (kompilierend), 136 rot, 24 grün; 17 der grünen tragen ein Finding, 7 sind gleichwertig oder als Negativ benannt (I-1). Sieben Ansätze brachen mit Build-Fehler ab, griffen nicht oder trafen die Zusage nicht; sie sind korrigiert wiederholt oder verworfen und zählen nicht. Alle Läufe: `go test -count=1` im per Digest gepinnten Image der Stufe `deps` (Tag `odata-recorder:deps-mut`, Go 1.27.1), in einem Container mit `--network none`, auf einer Kopie von `src/` (Stand `b5c9dd2`) im Scratch-Verzeichnis; der Arbeitsbaum blieb unberührt. Jede grüne Mutation mit Finding ist über das ganze Paket (`-run ''`) bestätigt, nicht nur über den genannten Test. Fänger mit Wahrscheinlichkeit: Testbinary mit `-test.count 1` in 200 eigenen Prozessen, je mutiert und unmutiert.
Ergänzend, ohne Mutation: `-count=50` über `httpupstream` (grün, 6,6 s); `-count=300` für den Fall „Kopf im Augenblick des Ablaufs“ (grün); Wegwerf-Probe zu `url.Parse` im gepinnten Image (`http://a%41b`, `http://a%3Cb`, `http://[v1.x]`, `http://[127.0.0.1]` abgewiesen, `http://a%25b` angenommen); Vergleich von `b6d5270` mit der Umkehrung von `0716718`; Prüfsumme der Gegenproben-Eingaben gegen `harness/gegenproben.stempel` (gleich, ohne `make gegenproben`). Der Code ist nach diesen Läufen richtig; die Findings sind Fänger-Lücken.

**Eingangs-Kontext:**

- Slice-Plan `slice-upstream-adapter` (§1, §3 mit A1 bis A19 und den akzeptierten Negativen, §4 Schnitt beim Handoff und Entscheidungen des Architect, §6 Raster, Zusatztabelle, Risiken) in `b5c9dd2`; der Nehmer `slice-upstream-tls` (§1, §3 A10, §6 Zwischenstand)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Go, Standardbibliothek, `testing/synctest`), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md) (Kante `adapters-driven → ports-outbound`), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) (Lint-Profil, keine `//nolint`)
- [LH-FA-01](../../spec/lastenheft.md), [LH-RB-02](../../spec/lastenheft.md), [LH-QA-07](../../spec/lastenheft.md), [LH-FA-14](../../spec/lastenheft.md) (nur gelesen)
- `SPEC-120`, `SPEC-110` Nr. 7, 12, `SPEC-109` Nr. 5, `SPEC-121` Kopf, Nr. 4, 11
- Vorherige Findings am benachbarten Modul: Reports `2026-10-21-slice-record-service-review` (F-12: `Body` nil aus `Senden`; hier durch A7 (d) gefangen) und `2026-10-21-slice-capture-pipeline-review` (Klassen „Grenze nur auf einer Seite belegt“, „Paar von Ursachen ohne Fänger“, „Allaussage ohne Fänger über alle Fälle“, „Zelle des Rasters ohne Fänger“)
- `AGENTS.md` (Hard Rules)

---

## Antworten auf die Fragen des Auftrags

**(a) Revert.** `b6d5270` ist zeilengenau die Umkehrung von `0716718` (Vergleich der Diff-Zeilen); der Baum danach gleicht `da6e121` bis auf die Testfälle aus `a736bb8`. Kein TLS-Rest: kein `tls.go`, kein `tls_test.go`, kein `TestMain`, kein `crypto/tls`, `crypto/x509`, `httptest`, `SSL_CERT_*`, `localhost` oder `tlsFehlerFaelle`; `https` steht nur noch in der Schema-Prüfung von `leseZiel` und ihrem Kommentar. Der Kommentar von `New` nennt nur `Frist`, Proxy und Kompression, und jede dieser drei Zusagen fängt eine Mutation (A1, A11, A4 und A7).

**(b) Zwischenstand.** Wie in `5947116` entschieden: `SPEC-120` nennt ihn nicht (kein Wort zu Zwischenstand, Cherry-Pick oder `slice-upstream-tls`); §3 des Plans führt ihn als akzeptiertes Negativ mit Grund (Schnitt beim Handoff) und Folge (keine im Betrieb, weil der Composition Root keinen Adapter verdrahtet — geprüft: `src/bootstrap/` importiert nur `model`); der Kommentar von `Konfig` lautet wörtlich wie in §3 vorgegeben und sagt keine Wirkung von `Pruefen` und `CAPEM` zu. Dass der Satz im Commit zu A19 und nicht im ersten Commit nach dem Revert kam, ändert nichts: dazwischen liegt nur `0587370` ohne Code. Eine Folge des Schnitts nennt der Plan nicht: das Element `https` der Schema-Prüfung aus Nr. 2 hat seit dem Revert keinen Fänger mehr (F-2).

**(c) Tabelle *Zusage · Fänger*.** Jede Zeile ist rot aufgelöst (Tabelle unten). Grün bleiben Mutationen, deren Eingabe kein Fall führt: Grenzen nur von einer Seite (F-1, F-7), ein Element einer Menge (F-2), Paare mit dem Kopf (F-3), der schon beendete Kontext bei gültigen Argumenten (F-4).

**(d) Raster und Zusatztabelle.** Jede Zelle mit Ort ist durch eine Mutation der Zeilen A1 bis A19 rot, bis auf vier: `Senden` · `ctx` · fremd „schon beendet: Kontextfehler“ (F-4), `UpstreamAntwort` · `Header` · leer „Kopf ohne Header-Zeile: Header ohne Einträge“ (F-6), Start · fremd „Prozent-Kodierung eines ASCII-Bytes außer `%25`, IPvFuture: ungültig“ (F-8) und `Senden` · `Ziel` mit dem gültigen Schema `https` (F-2). Die Zeilen „Annahme“ der Zusatztabelle tragen, was sie sagen; „im Augenblick des Ablaufs“ ist unmutiert in 300 Läufen grün.

**(e) Formen der Eingabe, Grenzen, Gleichwerte (`AGENTS.md` §3.15).** Ein Fall je Element: erfüllt für die Ausnahmen der Header, die Methoden der Rahmung, die Elemente aus Nr. 8, die vier Bytes aus A19 (je Leser), die Token-Zeichen; nicht für das Schema `https` (F-2). Beide Seiten jeder Grenze: erfüllt für Port, `Frist`, `Laenge`, Steuerzeichen in Header-Werten (0x08/0x09/0x0A, 0x1F/0x20, 0x7E/0x7F/0x80); nicht für die Byte-Prüfung in `Pfad`, `RawQuery` und dem Pfad von `Ziel` (F-1) und nicht für die Bereiche der Token-Zeichen nach außen (F-7). Gleichwert: der einzige Vergleich zweier Werte mit eigener Zusage ist Kopf gegen Ablauf der Frist, gefangen durch A8 (e) „im Augenblick“; die übrigen Vergleiche im Code gehen gegen Konstanten oder sind ohne Wirkung auf den Vertrag (`cap(puffer) < n`).

**(f) Paare gleichzeitiger Ursachen.** Vor dem Kopf trägt jedes Paar aus Nr. 8 einen deterministischen Fall (A13 (b), A16 (a) bis (g), A18), jede Vertauschung ist rot. Mit dem Kopf fehlen drei Paare: Fehler von `Body` × Kopf, 101 × `ctx`, 101 × Ablauf der Frist (F-3).

**(g) Raten.** Fänger mit Wahrscheinlichkeit sind benannt und gemessen: A4 (c) letzter Fall, Plan 25 %, gemessen 42 von 200 Prozessen rot, unmutiert 0 von 200; A15, Plan 200 von 200, gemessen 200 von 200, unmutiert 0 von 200. A12 „späteres `Read`“ hängt an Bytes im Puffer des Clients und war 200 von 200 rot; ohne Befund. Ein unbenannter Fänger mit Rate ist nicht gefunden.

**(h) Akzeptierte Negative.** Alle sechs in §3 nennen Grund und Folge (Datenrennen; Ende zugleich mit einer Anforderung, Mutation „Prüfung vor dem `Read` der Quelle entfernt“ grün wie benannt; Reihenfolge der Map bei `User-Agent`; die Fälle aus *Ungeprüft*; Leerzeichen im Pfad von `UPSTREAM_URL`; Zwischenstand). Die Fälle in `SPEC-120` *Ungeprüft* und `SPEC-110` *Ungeprüft* nennen je ihre Folge.

**(i) Allaussage A14.** Über die Tabelle `alleFehlerFaelle` (90 Fälle) rot für jede meiner vier Mutationen (Text der Bibliothek, `http.Client`, `ErrUngueltig` mit Wert, Text von `errBody`). Nicht über alle Fehlerwege: der Weg „schon beendeter `ctx`, gültige Argumente“ hat keinen Fall (F-4), und die Hälfte „höchstens Host und Port aus `Ziel`“ hat keine Zeile und keinen Marker (F-5).

**(j) Randformen im Code (`AGENTS.md` §3.12, §3.16).** Keine im Code entschiedene Randform gefunden. Gelesen: Größe eines Stücks (`maxStueck`, ohne Zusage im Vertrag), `Read` mit leerem Puffer, `Close` des Lesers ohne Wirkung (`Body` ist ein `io.Reader`, den der HTTP-Adapter besitzt), Schema in Großschreibung (die URL-Syntax, `url.Parse` liest es klein), Host mit Zone im Kopf `Host` (*Ungeprüft*). Jede Zelle des Rasters nennt einen Ort, dessen Wortlaut sie trägt.

---

## Mutationen je Zeile

| Zeile | rot | grün |
|---|---|---|
| A1 | `< 1 ns` zu `<= 1 ns` (1 ns); Prüfung entfernt (0 und −1 ns); `< 0` (0) | — |
| A2 | Schema, leerer Host, Zugangsdaten, nur mit Namen (`@`), Query/Fragment über `url.URL` (`?`, `#`), Port entfernt, Port `>= 2`, `>= 0`, `< 65535`, `<= 65536`, Doppelpunkt ohne Port, Pfad von `Ziel`, Token der Methode, `CONNECT` entfernt, `CONNECT` ohne Schreibweise, Token als „ohne Leer- und Steuerzeichen“, `\|` und `!` aus der Menge, `0`→`1`, `Z`→`Y`, `/` am Anfang des Pfads, `#` und `?` im Pfad, `#` in der Query, Query-Prüfung entfernt, Trenner `< 0x20`, `0x7F` entfernt, `>= 0x7F`, Header-Name und -Wert entfernt, Tabulator verboten, `< 0x1F`, `< 0x21`, `0x7F` entfernt, `>= 0x7F`, `>= 0x7E`, Tabulator-Ausnahme auch für 0x0A und 0x08, `Laenge < -2`, `< 0`, `Body` nil bei jeder `Laenge`, Kontext vor den Argumenten; `ctx` nil und `Body` nil entfernt: rot über Panic (I-2) | Trenner `< 0x22`, `0x7E` verboten, `0x80` verboten (F-1); Token-Bereich `/`, `:`, `@`, `[`, `{` (F-7); nur Schema `http` (F-2); vorab beendeter `ctx` als nicht erreichbar (F-4) |
| A3 | Pfad über `EscapedPath`, `ForceQuery`, Methode groß, Pfad von `Ziel` vergessen, keine absolute Form, absolute Form nur bei `Pfad` | — |
| A4 | Kompression, Standard-`User-Agent`, `Host` aus der Map, nur erster Wert, Namen kanonisch, Ausnahmen kanonisch, `User-Agent` nur kanonisch, Leerraum-Wert durchgereicht, leere Liste als Treffer; Sortierung der Schreibweisen entfernt: 42 von 200 Prozessen | leere Wertliste als Eintrag der Map (gleichwertig, I-1) |
| A5 | `ContentLength` nur bei −1, `chunked` der Bibliothek überlassen, `Body` bei 0 durchgereicht, `io.LimitReader` | — |
| A6 | `io.ReadAll` vor dem Senden, Antwort ganz gelesen, Lesen auf Vorrat (`TestKeinVorrat`, A18 grün wie im Plan) | — |
| A7 | Kompression, `http.Client` (Weiterleitung), `Body` nil bei 204, Bytes nach der Länge im Kopf bei `HEAD`, Fehler des Bodys als `io.EOF` | zusätzlicher Antwort-Header (F-6); 1xx statt 101 (gleichwertig, I-1) |
| A8 | Frist nicht angewendet, nach dem Kopf nicht beendet, Ergebnis des Beendens übergangen, −1 ns (1 ns davor), +1 ns (im Augenblick) | +1 ns gegen den Fall 1 ns danach (wie im Plan, 200 Läufe) |
| A9 | alles als Zeit, Fehler der Bibliothek unverpackt, 101 zurückgegeben, 101 nicht geschlossen | — |
| A11 | `http.ProxyFromEnvironment` | — |
| A12 | Kontext als nicht erreichbar, Deadline als Zeit, Request mit `context.WithoutCancel` (Blase: deadlock), Prüfung im Antwort-Body entfernt (200 von 200) | — |
| A13 | Fehler von `Body` entfernt, Frist vor dem Fehler von `Body` | — |
| A14 | Text der Bibliothek (14 Fälle), `http.Client` (30), `ErrUngueltig` mit Wert (20), Text von `errBody` (3) | vorab beendeter `ctx` mit `Pfad` und `RawQuery` im Text (F-4); `Ziel` im Text (F-5) |
| A15 | eine Header-Map am `Client`: 200 von 200 Prozessen | — |
| A16 | `ctx` vor dem Kopf entfernt, `Body` vor `ctx`, Frist vor `ctx`, Frist entfernt, `ctx` nach dem Kopf entfernt, Frist vor `ctx` beim Kopf, Body des Kopfs offen | Fehler von `Body` beim Kopf, 101 vor `ctx` und Frist (F-3) |
| A17 | Header-Map am `Client`, Rahmung am `Client` | `Content-Length` in der Header-Map (gleichwertig, I-1) |
| A18 | Bibliothek liest `Body` direkt, Lesen nach dem Ende, Unterbrechung als Fehler von `Body`, Ergebnis nach dem Ende nicht verworfen, Schließen ohne Ende, Fehler ohne Ende, 101 ohne Ende | Prüfung vor dem `Read` der Quelle (benanntes Negativ) |
| A19 Adapter | Prüfung entfernt, je ein Byte weniger, auf den Text, auf `u.Host`, nur außerhalb der Klammern, ohne Auflösen der Prozent-Kodierung, ab 0x80 abgewiesen, nur 0x80 abgewiesen | — |
| A19 Start | dieselben Mutationen in `normalisiereURL`: Prüfung entfernt, je ein Byte weniger, auf den Text, auf `u.Host`, nur außerhalb der Klammern, ohne Auflösen, ab 0x80, nur 0x80 | Prozent-Kodierung eines ASCII-Bytes angenommen (F-8) |
| Raster `UpstreamAntwort` | — | Eintrag in leerem Kopf (F-6) |

---

## Findings

Kein Fix wird vorgeschlagen (Skill: Reviewer kategorisiert, Implementer entscheidet). Die Kategorie folgt `AGENTS.md` §3.15 und §3.16 (Zusage ohne Fänger: MEDIUM); LOW steht wie in den Reports zu `slice-capture-pipeline` und `slice-record-service` für Lücken ohne Wirkung auf Ablage, Client oder Geheimnis.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Byte-Prüfung von `Pfad`, `RawQuery` und dem Pfad von `Ziel` ist nur auf der verbotenen Seite belegt (0x20, 0x7F): „`< 0x22`“, „0x7E verboten“ und „0x80 verboten“ bleiben über das ganze Paket grün, kein Fall führt `!`, `~` oder 0x80 in einem dieser Felder. Gemutet endet ein Request mit `~` oder `!` im Pfad oder mit `À` (0xC3 0x80) mit `ErrUngueltig` und damit `ODR-E9001` statt beim Upstream; der Plan nennt zu A2 „je Grenze verschoben“, belegt aber nur die schrumpfende Richtung. Zuständig: Implementer. | `SPEC-120` Nr. 2, 3; `LH-FA-01`; `AGENTS.md` §3.15 | `src/adapters/driven/httpupstream/client.go` · `if s[i] < 0x21 \|\| s[i] == 0x7F \|\| strings.IndexByte(verboten, s[i]) >= 0 {` | ja — drei Mutationen, `go test` rc 0 | Grenze nur auf einer Seite belegt |
| F-2 | MEDIUM | Das Schema `https` ist in Nr. 2 gültig, aber seit dem Revert von `0716718` führt kein Fall ein `https`-Ziel; „nur `http`“ in `leseZiel` bleibt grün. Gemutet endet jeder Request an einen `https`-Upstream mit `ErrUngueltig`; erst der Cherry-Pick in `slice-upstream-tls` (dort A10) fängt es wieder, und der Zwischenstand in §3 nennt diese Lücke nicht. Zuständig: Implementer (Fänger) oder Planner (Negativ des Zwischenstands). | `SPEC-120` Nr. 2; `LH-RB-02`; `AGENTS.md` §3.15 | `src/adapters/driven/httpupstream/client.go` · `(u.Scheme != "http" && u.Scheme != "https")` | ja — Mutation, `go test` rc 0 | Fänger trennt die Mutation nicht (ein Fall je Element) |
| F-3 | MEDIUM | Drei Paare mit dem Kopf haben keine Zusatzzeile und keinen Fall: Fehler von `Body` × Kopf (Nr. 4: der Fehler zählt nur, solange der Kopf nicht da ist; „Fehler von `Body` auch mit Kopf“ bleibt grün), 101 × `ctx` und 101 × Ablauf der Frist („101 vor `ctx` und Frist“ bleibt grün). Gemutet ersetzt ein Upload-Fehler die frühe Antwort des Upstreams (etwa 401) durch den Fehler von `Body`, und eine 101 bei abgebrochenem Client oder abgelaufener Frist wird 502 statt Kontextfehler oder 504. Zuständig: Implementer. | `SPEC-120` Nr. 4, 5, 6, 8; `AGENTS.md` §3.16 (Paar) | `src/adapters/driven/httpupstream/fehler.go` · `} else if res.StatusCode == http.StatusSwitchingProtocols {` | ja — zwei Mutationen, `go test` rc 0 | Paar von Ursachen ohne Fänger |
| F-4 | MEDIUM | Der eigene Weg für einen schon beendeten `ctx` bei gültigen Argumenten hat keinen Fall: „als `ErrUpstreamNichtErreichbar`“ und „mit `Pfad` und `RawQuery` im Text“ bleiben grün; die Zusatzzeile „Argument × beendeter `ctx`“ deckt nur die Reihenfolge, und `alleFehlerFaelle` enthält den Weg nicht, so dass auch A14 ihn nicht erreicht (ebenso der Kommentar „Kein Fehler dieses Pakets nennt einen Wert“). Gemutet trägt der Fehler eines vom Client schon abgebrochenen Requests dessen Query-Token in die Diagnose (`LH-FA-14`), oder der Dienst meldet `ODR-E4001` für einen Client, der fort ist. Zuständig: Implementer. | `SPEC-120` Nr. 2, 8; `LH-FA-14`; `AGENTS.md` §3.15, §3.16 (Allaussage) | `src/adapters/driven/httpupstream/client.go` · `return outbound.UpstreamAntwort{}, kontextFehler(err)` | ja — zwei Mutationen, `go test` rc 0 | Form der Eingabe ohne Fall |
| F-5 | MEDIUM | Nr. 8 sagt zu, ein Fehler nenne „höchstens Host und Port aus `Ziel`“; A14 nimmt diese Hälfte nicht auf, und kein Fehlerfall trägt einen Marker im Pfad von `Ziel`. „`Ziel` an jeden Fehler gehängt“ bleibt grün. Gemutet steht der Pfad des Upstreams aus `UPSTREAM_URL` in jedem Fehlertext und damit im Protokoll. Zuständig: Implementer (Fänger), Planner (Zeile A14). | `SPEC-120` Nr. 8; `AGENTS.md` §3.15, §3.16 (Allaussage) | `src/adapters/driven/httpupstream/helfer_test.go` · `a.Pfad = "/" + geheim + "-pfad"` | ja — Mutation, `go test` rc 0 | Allaussage ohne Fänger über alle Fälle |
| F-6 | MEDIUM | `TestAntwort` prüft die erwarteten Namen der Antwort, nicht dass keine weiteren stehen; die Zelle „Kopf ohne Header-Zeile: Header ohne Einträge“ hat keinen Fall. „`Via` an jede Antwort“ und „Eintrag in leerem Kopf“ bleiben grün. Gemutet zeichnet der Record-Dienst einen Header auf, den der Upstream nie sandte, und Replay liefert ihn aus (A7 „Antwort unverändert“, `LH-QA-07`). Zuständig: Implementer. | `SPEC-120` Nr. 5; `LH-QA-07`; `AGENTS.md` §3.15, §3.16 | `src/adapters/driven/httpupstream/antwort_test.go` · `if w := ant.Header[name]; !reflect.DeepEqual(w, soll) {` | ja — zwei Mutationen, `go test` rc 0 | Zelle des Rasters ohne Fänger |
| F-7 | LOW | Die Bereiche der Token-Zeichen sind nur nach innen belegt: `/`, `:`, `@`, `[` oder `{` als Token angenommen bleibt grün (fünf Mutationen), kein Fall führt ein Zeichen direkt neben einem Bereich; der Kommentar von `istToken` sagt „HTTP-Token (RFC 9110)“ zu. Gemutet liefert ein solcher Name `ErrUpstreamNichtErreichbar` statt `ErrUngueltig`; über den HTTP-Adapter erreicht ihn kein Request, weil `net/http` nur Token annimmt. Zuständig: Implementer. | `SPEC-120` Nr. 2; `AGENTS.md` §3.11, §3.15 | `src/adapters/driven/httpupstream/client.go` · `alnum := b >= '0' && b <= '9' \|\| b >= 'a' && b <= 'z' \|\| b >= 'A' && b <= 'Z'` | ja — fünf Mutationen, `go test` rc 0 | Grenze nur auf einer Seite belegt |
| F-8 | LOW | Die Zelle Start · fremd nennt „Prozent-Kodierung eines ASCII-Bytes außer `%25`, IPvFuture: ungültig“ (neuer Satz in `SPEC-110` Nr. 7); `TestKonfigURL` führt keine dieser Formen, und eine Zeile *Annahme* fehlt. „Bei einem Fehler von `url.Parse` mit `%25` erneut lesen“ bleibt grün. Der Code unterscheidet die Formen nicht, die Entscheidung liegt in `url.Parse`; ein Sprung des Images, der sie annimmt, bliebe unbemerkt. Zuständig: Implementer. | `SPEC-110` Nr. 7; `AGENTS.md` §3.16 | `src/bootstrap/konfig_test.go` · `"http://h:0", "http://h:65536", "http://h:99999", "http://h:", "http://h:/x"` | ja — Mutation und Probe, `go test` rc 0 | Zelle des Rasters ohne Fänger |
| F-9 | LOW | Der Kommentar von `geheim` sagt, der Marker stehe „in jedem Wert einer Anfrage und jeder Antwort, den ein Fehlertext nicht nennen darf“; er fehlt im Pfad von `Ziel` (F-5) und in den Bodies von `rahmungFehlerFaelle`. Zuständig: Implementer. | `AGENTS.md` §3.11 | `src/adapters/driven/httpupstream/helfer_test.go` · `// geheim steht in jedem Wert einer Anfrage und jeder Antwort, den ein Fehlertext nicht nennen darf` | nein — Lesen | Kommentar sagt mehr zu als die Fixture trägt |
| F-10 | LOW | Die Dateitabelle in §3 des Plans legt 101 in `antwort.go`; der Code prüft 101 in `ergebnis` in `fehler.go`, `antwort.go` kennt 101 nicht. Zuständig: Implementer (Plan). | `AGENTS.md` §3.9 | `docs/plan/planning/in-progress/slice-upstream-adapter.md` · `` `antwort.go` (101, Body-Hülle `` | nein — Lesen | Plan-Satz folgt dem Code nicht |
| I-1 | INFO | Sieben grüne Mutationen ohne Finding: gleichwertig („leere Wertliste als Eintrag der Map“ — die Bibliothek schreibt zu einem Namen ohne Werte keine Zeile; „`Content-Length` in der Header-Map am `Client`“ — die Bibliothek schreibt ihn nicht aus der Map; „1xx statt 101“ — die Bibliothek liefert außer 101 keine 1xx; „Backtick im Bereich `a`–`z`“ — schon Token-Zeichen; „Vorabprüfung von `ctx` entfernt“ — `RoundTrip` endet vor dem Verbinden mit demselben Fehler) oder im Plan benannt („+1 ns“ gegen den Fall 1 ns danach; Prüfung vor dem `Read` der Quelle). | Maintainability | `src/adapters/driven/httpupstream/client.go` · `case len(werte) > 0:` | ja — Mutationen | äquivalente Mutation |
| I-2 | INFO | Drei Mutationen sind über einen Panic rot, nicht über eine Zusicherung: Prüfung `ctx` nil entfernt (nil-Dereferenz in `ctx.Err()`), Prüfung `Body` nil entfernt und nur für `Laenge > 0` (nil-Dereferenz in der Goroutine des Lesers, der Prozess endet). Der Plan nennt das zu A2; der Test sieht den Fehler, den Grund nennt erst der Stack. | `AGENTS.md` §3.10 | `src/adapters/driven/httpupstream/client.go` · `case a.Body == nil && a.Laenge != 0:` | ja — Mutationen | Rot über Panic statt Zusicherung |
| I-3 | INFO | Größe bestätigt: 2567 über `src/` ab `f76ba07`, Marke 2900; die Fänger zu F-1 bis F-8 kommen in die Messung nach dem Review. | `AGENTS.md` §3.14 | `docs/plan/planning/in-progress/slice-upstream-adapter.md` · `über **2900** Zeilen nach dem Review` | ja — `git diff --numstat` | Marke vor dem Review gemessen |
| I-4 | INFO | Ein Host mit `"`, `<`, `>` oder `]` meldet beim Start den Grund „keine absolute http- oder https-URL ohne Zugangsdaten, Query und Fragment“, der die verletzte Regel nicht nennt (ebenso schon beim Port); der Plan hat das so gewollt (§3, „mit dem Grund der übrigen Verstöße der Form“). Ob der Grund die Regel nennen soll, ist eine Frage an den Architect. | `SPEC-110` Nr. 7, 12 | `src/bootstrap/konfig_dienste.go` · `p.melde(schluessel, "keine absolute http- oder https-URL ohne Zugangsdaten, Query und Fragment")` | nein | Meldung nennt die verletzte Regel nicht |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/httpupstream/client.go` — `Konfig`, `New`, `Senden`, `pruefe`, `leseZiel`, Bau von Request und Headern | geprüft; Befunde F-1, F-2, F-4, F-7, sonst ohne Befund (alle Mutationen von A1 bis A4 und A19 außer den genannten rot) |
| `src/adapters/driven/httpupstream/leser.go`, `rahmen.go`, `frist.go` — Leser des Request-Bodys, Rahmung, Frist | geprüft, ohne Befund (jede Mutation von A5, A6, A8, A18 rot; Grenze des Lesers als Negativ benannt) |
| `src/adapters/driven/httpupstream/fehler.go`, `antwort.go` — Rangfolge, Body-Hülle | geprüft; Befunde F-3, F-4, sonst ohne Befund (jede Vertauschung vor dem Kopf rot; kein Weg gefunden, auf dem ein Wert der Anfrage den Fehlertext erreicht) |
| Tests des Pakets (`helfer_`, `fehler_`, `anfrage_`, `antwort_`, `frist_`, `client_test.go`, `export_test.go`) | geprüft; Befunde F-1 bis F-9, I-2; jeder Testname aus §3 existiert, drei Haken in `export_test.go` wie im Plan, Fixtures mit Steuerbytes und Bytes ab 0x80 als Go-Escapes (kein rohes Byte), `-count=50` grün |
| `src/bootstrap/konfig_werte.go`, `konfig_test.go` — Name des Hosts in `UPSTREAM_URL` (A19) | geprüft; Befunde F-8, I-4, sonst ohne Befund (dieselben Fälle und Mutationen wie im Adapter rot) |
| Revert `b6d5270` gegen `0716718`; Kommentar von `Konfig` und `New` | geprüft, ohne Befund (reine Umkehrung, kein TLS-Rest, kein Kommentar sagt eine Wirkung von `Pruefen` oder `CAPEM` zu) |
| Zwischenstand nach `5947116` — Plan §3, §4, `SPEC-120`, Composition Root | geprüft; Befund F-2, sonst ohne Befund |
| Akzeptierte Negative in §3, `SPEC-120` *Ungeprüft*, `SPEC-110` *Ungeprüft* | geprüft, ohne Befund (Grund und Folge je Fall) |
| Suppression, Import-Richtung, Lint (`AGENTS.md` §3.2, [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md)) | geprüft, ohne Befund (kein `//nolint`; Importe nur Standardbibliothek, `ports/outbound` und `model`) |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft; Befunde F-7 (Kommentar von `istToken`), F-9, sonst ohne Befund (keine verworfene Alternative, kein abwesender Text) |
| `harness/gegenproben.stempel` | geprüft, ohne Befund (Prüfsumme der Eingaben gleich dem Stempel) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 6 |
| LOW | 4 |
| INFO | 4 |

**Finding-Klassen dieses Laufs:** Grenze nur auf einer Seite belegt · Fänger trennt die Mutation nicht (ein Fall je Element) · Paar von Ursachen ohne Fänger · Form der Eingabe ohne Fall · Allaussage ohne Fänger über alle Fälle · Zelle des Rasters ohne Fänger · Kommentar sagt mehr zu als die Fixture trägt · Plan-Satz folgt dem Code nicht · äquivalente Mutation · Rot über Panic statt Zusicherung · Marke vor dem Review gemessen · Meldung nennt die verletzte Regel nicht

Für die Closure §7: F-1 bis F-8 sind Belege zu `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (F-4 und F-5 auch zu `BEO-REPO/universalaussage-im-vertrag-ungeprueft`). „Grenze nur auf einer Seite belegt“, „Paar von Ursachen ohne Fänger“, „Allaussage ohne Fänger über alle Fälle“ und „Zelle des Rasters ohne Fänger“ standen schon in den Reports zu `slice-capture-pipeline` und `slice-record-service`.

## Verdikt

**Merge-blockierend:** ja — sechs MEDIUM. Kein HIGH; der Konflikt-Pfad über den Architect ist nicht ausgelöst. Der Code ist nach allen Läufen richtig; die MEDIUM sind Fänger-Lücken.

**Übergabe:** Implementer: F-1 bis F-10. Planner: F-2, falls die Lücke als Negativ des Zwischenstands statt mit Fänger getragen wird; die Zeile A14 (F-5); die Messung nach dem Review (I-3). Architect: I-4. Die Finding-Klassen gehen in die Slice-Closure §7. Dieser Report ersetzt keine Verifikation gegen die DoD (Verifier).
