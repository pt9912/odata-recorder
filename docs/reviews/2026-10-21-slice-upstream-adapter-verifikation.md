# Verifikation: slice-upstream-adapter

* Datum: 2026-10-21
* Rolle: verifier (frischer Kontext; kein Code, keine Spezifikation, kein Plan, kein Lastenheft, keine `AGENTS.md` geändert)
* Gegenstand: Spitze `1fab875`, `git diff f76ba07..1fab875 -- src/` (`src/adapters/driven/httpupstream/` und `normalisiereURL` in `src/bootstrap/`). Verlauf `f76ba07` (Plan vor dem Code) → `da6e121` → `0716718` (TLS-Teil, nicht Gegenstand) → `b6d5270` (Revert) → `229d65a` (A19) → `7cecc39` (Review) → `6fdb5be`, `1fab875` (Fänger zu F-1 bis F-10)
* Plan: `docs/plan/planning/in-progress/slice-upstream-adapter.md`
* Maßstab: DoD (§2), Tabelle *Zusage · Fänger* (§3) und Raster (§6) gegen `SPEC-120` Kopf, Nr. 1 ohne `CAPEM`, Nr. 2 bis 6, 8, 9 und *Ungeprüft*, `SPEC-110` Nr. 7 und 12; Lastenheft [LH-FA-01](../../spec/lastenheft.md), [LH-RB-02](../../spec/lastenheft.md), [LH-QA-07](../../spec/lastenheft.md) (nur gelesen); `AGENTS.md` §3.7, §3.9 bis §3.16; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)
* Belege selbst gefahren: 133 eigene Ansätze auf Kopien von `src/` (Abschnitt 3), sieben Wegwerf-Sonden zu den grünen Mutanten (nicht committet, Abschnitt 3), Raten der Fänger mit Wahrscheinlichkeit und Zeitläufe (Abschnitt 4), `git diff --numstat` (Abschnitt 6), `git merge-tree` samt Bau und Test des Ergebnisses (Abschnitt 7), `make gegenproben`, `make test` und `make lint` ungecacht, `make gates` (Abschnitt 10). Jedes `go test` lief im per Digest gepinnten Image der Stufe `deps` (Tag `odata-recorder:deps-mut`, Go 1.27.1), `--network none`, die Quelle schreibgeschützt, Mutanten in Kopien im Scratch-Verzeichnis; der Arbeitsbaum blieb unberührt (`git status` zeigt nur die fremde Änderung an `.claude/agents/architect.md`).

## Urteil

**Nicht freigegeben: kein HIGH, drei MEDIUM.** Der Code ist nach allen Läufen richtig: jede Sonde ist auf `1fab875` grün, und die Zusagen, die ich neu prüfte (`Laenge` und `Frist` bis `math.MaxInt64`, alle 256 Bytes als Token, die Formen des Hosts), hält er. Die Fixes zu F-1 bis F-10 tragen: jede Mutation aus dem Review, von mir nachgebaut, ist rot (Abschnitt 4). Die Befunde sind Lücken der Fänger an der Allaussage A14 und an Zusagen über das Fehlen einer Grenze.

* MEDIUM (Implementer, Fänger): V1 die Allaussage „höchstens Host und Port aus `Ziel`“ ist für den Fehler eines Verstoßes im Ziel nicht belegt; V2 „kein Byte der Antwort“ ist für die Fehlerwege mit gelesenem Kopf (101, `ctx`, Frist) nur im Body belegt; V3 „ohne Obergrenze“ für `Laenge` (Nr. 4) hat keine Zeile und keinen Fänger.
* LOW: V4 „keine Obergrenze“ der `Frist` (Raster-Zelle); V5 (Architect) Prozent-Kodierung und IPvFuture im Host von `Ziel`; V6 die Nicht-Token-Zeichen sind nur zu einem Teil belegt; V7 Kommentare sagen mehr zu, als V1, V2 und V9 tragen; V8 (Planner) I-4 des Reviews hat keinen Träger.
* INFO: V9 (Architect) `Methode` im Fehlertext; V10 (Planner) `slice-upstream-tls` §1 nennt A19 nicht; V11 die zwei grünen Mutanten des Implementers und der Fall `Ziel` mit `0x7F`; V12 Größe 2655 unter der Marke 2900.

Von 33 grünen Mutanten sind 8 gleichwertig, erlaubt oder künstlich, 25 tragen ein Finding; zu V1 bis V6 trennt je eine Sonde jeden dieser Mutanten (Abschnitt 3), zu V9 steht die Entscheidung des Architect aus.

## 1. DoD Punkt für Punkt

| DoD | Stand | Beleg |
|---|---|---|
| Adapter `httpupstream` nach `SPEC-120` Kopf, Nr. 1 ohne `CAPEM`, 2 bis 6, 8, 9; Fänger A1, A2 bis A9, A12 bis A19 | Code erfüllt; Fänger lückenhaft in Nr. 8 (V1, V2), Nr. 4 (V3), Nr. 1 (V4), Nr. 2 (V5, V6) | jeder Satz aus Abschnitt 2 hat Zeile und Fänger bis auf die genannten; 94 Mutanten rot aus dem benannten Grund; Sonden grün auf dem Code |
| Transport nach Nr. 7 ohne TLS (A11 `TestTransport`) | erfüllt | `Proxy` nil, `DisableCompression`; `http.ProxyFromEnvironment` rot, Kompression an rot; kein anderes Paket importiert `httpupstream` |
| TLS-Teil mit reinem Revert von `0716718` genommen, Messung in §6, `Konfig` ohne Zusage zu `Pruefen` und `CAPEM` | erfüllt | Diff-Zeilen von `b6d5270` sind die Umkehrung von `0716718`; Baum gleicht `da6e121` bis auf die Fälle aus `a736bb8`; kein `crypto/tls`, `httptest`, `TestMain`, `tlsFehlerFaelle`; `CAPEM []byte` steht nur als Feld von `Konfig`; Messung 2545 (522, 2023) stimmt; Kommentar von `Konfig` neutral |
| `make gates` grün | erfüllt | eigener Lauf auf `1fab875`, rc 0; `make test` und `make lint` zusätzlich ungecacht (Abschnitt 10) |
| Review durchgeführt, Report liegt vor | erfüllt | `docs/reviews/2026-10-21-slice-upstream-adapter-review.md` (`7cecc39`) |
| Doku-Update (`SPEC-120`: Nachzug nach Review und Verifikation) | offen | V5 und V9 verlangen eine Entscheidung des Architect in `SPEC-120` Nr. 2 und Nr. 8; ein Nachzug davor wäre vorzeitig |
| Closure-Notiz, Register, Ausgänge, Paarungen | offen, beim Planner | §7 „Wird bei der Closure gefüllt“; sechs Risiken in §6 tragen noch *offen* (Abschnitt 9) |

Ein Häkchen setze ich nicht.

## 2. Tabelle *Zusage · Fänger* gegen den Vertrag

Jeder Satz von `SPEC-120` (ohne `CAPEM` aus Nr. 1 und Nr. 7) und `SPEC-110` Nr. 7 gegen §3 und §6 des Plans gelesen. Jeder Fänger unter seinem Namen steht im Code: `TestNewArgumente` (A1), `TestSendenArgumente` und `TestSendenArgumenteGegenseite` (A2, A19), `TestRequestZeile` (A3), `TestHeaderGenau` (A4), `TestRahmung` (A5), `TestStreaming` und `TestKeinVorrat` (A6), `TestAntwort` (A7), `TestFrist` (A8), `TestNichtErreichbar` (A9), `TestTransport` (A11), `TestKontext` (A12), `TestBodyFehler` (A13), `TestFehlerOhneWerte` (A14), `TestGleichzeitig` (A15), `TestRangfolge` (A16), `TestNacheinander` (A17), `TestBodyBlockiert` (A18), `TestKonfigURL` (A19 Start); die drei Haken in `export_test.go` (`SetzeFrist`, `SetzeTransport`, `Transport`) stehen wie im Plan, die Fixtures mit Steuerbytes und Bytes ab `0x80` als Go-Escapes (kein rohes Byte, alle Dateien gültiges UTF-8).

| Vertrag | Zusagen, Randformen, Raster-Zellen | Zeilen | Stand |
|---|---|---|---|
| Kopf | der Adapter sendet, was er bekommt; Leihe der Anfrage | A4; Pflicht des Aufrufers, *Ungeprüft* | gedeckt |
| Nr. 1 | `Frist` unter 1 ns ungültig, 1 ns gültig; `Konfig` trägt drei Felder; keine Obergrenze | A1; V4 | Obergrenze ohne Fänger (V4) |
| Nr. 2 | je Regel und je Element ein Fall; Argument vor Kontext; beide Seiten von Port, Byte-Prüfung, Steuerzeichen, `Laenge`; die vier Bytes im Host | A2, A3, A19 | gedeckt bis auf V5 (Prozent-Kodierung im Adapter) und V6 (Trenner des Tokens) |
| Nr. 2 | „mehrere Verstöße ergeben denselben Fehler“ | Zusatzzeile „gilt nicht: ein Fehler ohne Rangfolge“ | tragend: das `switch` in `pruefe` liefert einen Fehler, jede Umordnung der Fälle liefert `ErrUngueltig` |
| Nr. 3 | Request-Zeile byte-gleich, `//` absolut, genau die Header, Ausnahmen in jeder Schreibweise, `User-Agent`, kein Zusatz | A3, A4 | gedeckt |
| Nr. 4 | Rahmung je Methode, Strom, Lesen bis `io.EOF`, Fehler von `Body`, Ende des Lesens, blockiertes `Read`; „ohne Obergrenze“ | A5, A6, A13, A18 | gedeckt bis auf V3 |
| Nr. 5 | Antwort unverändert, 101, Zwischenantworten, Namen und Werte, `HEAD`/204/304, abgeschnitten, an `ctx` gebunden | A7, A9, A12 | gedeckt |
| Nr. 6 | Frist bis zum Kopf, beide Seiten, Kopf und Ablauf zugleich | A8 | gedeckt (Grenze ±1 ns, im Augenblick, Body danach) |
| Nr. 7 ohne TLS | kein Proxy, HTTP/1.1 | A11, A3 | gedeckt |
| Nr. 8 | Rangfolge, Paare mit und ohne Kopf, jede Art vor dem Kopf | A9, A13, A16 | gedeckt; die fünf Umordnungen von `rang` und sieben Mutanten in `ergebnis` rot |
| Nr. 8 | Allaussage „nennt `Pfad`, `RawQuery`, Header, Bytes; höchstens Host und Port aus `Ziel`“ | A14 über 106 Fälle | V1, V2, V9 |
| Nr. 9 | gleichzeitige Aufrufe, nacheinander | A15, A17 | gedeckt, A15 mit Rate (Abschnitt 4) |
| `SPEC-110` Nr. 7 | die vier Bytes im Namen, Zone, `0x80`, Prozent-Kodierung, IPvFuture | A19 beide Leser | Start gedeckt; Adapter nur für die vier Bytes (V5) |

Die Raster-Zellen: keine leere, keine ohne Ort; die Zelle `Ziel` · fremd nennt für die Prozent-Kodierung „Form der Bibliothek (*Ungeprüft*)“ (V5), die Zelle `Body, Laenge` · fremd „keine Obergrenze“ ohne Fänger (V3).

## 3. Eigene Mutationen

133 Ansätze, je auf einer frischen Kopie von `src/` (Stand `1fab875`, `git archive`), je eine Ersetzung, die ein Go-Programm im Image genau einmal ausführt (sonst Abbruch), dann `go test -count=1` über `httpupstream` oder `bootstrap`. 94 rot aus dem benannten Grund, 33 grün, 6 verworfen (zwei Build-Fehler, einer rot aus fremdem Grund, zwei rot ohne benannten Test an der Zeitgrenze, einer ohne Aussage).

| Gruppe | Ansätze | rot | grün | Bemerkung |
|---|---|---|---|---|
| Byte-Prüfung je Eingabe (`Pfad`, `RawQuery`, Pfad von `Ziel`) × `0x21`, `0x7E`, `0x80`, `0xFF`, `0xC3`; global `< 0x22`, `>= 0x80`, `> 0x80`, `>= 0x7E`, `< 0x20`, `0x7F` entfernt; `0x7F` je Eingabe; zweistellig im Pfad von `Ziel` | 25 | 23 | 2 | grün: `0xC3` allein im Pfad von `Ziel` (künstlich, `0x80` und `0xFF` decken die Seite), `0x7F` einstellig im Pfad von `Ziel` (V11) |
| Schema; die fünf Umordnungen von `rang`; sieben Mutanten in `ergebnis`; vorab beendeter `ctx` (fünf); `Via`, Eintrag im leeren Kopf, `nil` statt leerer Map | 22 | 20 | 2 | grün: Vorabprüfung entfernt und `nil` statt leerer Map (beide gleichwertig, V11) |
| A14: `Ziel`, Host und Port an jeden Fehler, Fehlertext je Feld, `errBody`, Fehler mit Kopf | 21 | 9 | 12 | grün: V1 (zwei), V2 (sechs), V9 (eins); erlaubt oder künstlich (drei: Host und Port an jeden Fehler in zwei Fassungen, Zeiger des Lesers) |
| Token-Zeichen: fünf Nachbarn der Bereiche, sechs innere Grenzen; elf Trenner und `0x7F`/`0x80` als Token | 19 | 11 | 8 | grün: V6 (acht) |
| Start `normalisiereURL`: erneut lesen mit `%25`, nur `%7F`, ohne Klammern, IPvFuture, Hostname ohne Auflösen | 5 | 5 | 0 | |
| Adapter: Prozent-Kodierung und IPvFuture im Host von `Ziel` | 4 | 0 | 3 | grün: V5; ein Ansatz rot aus fremdem Grund (`/%zz`), verworfen |
| Obergrenzen: `Laenge` über 1 GiB und 1 KiB, `Frist` über 24 h und 1 h, Klemme auf 30 min | 5 | 0 | 5 | grün: V3, V4 |
| A19 in beiden Lesern (je ein Byte weniger, `u.Host`, zusätzlich `%`, `'`, `~`, `0x80`, ohne Prüfung); Leser: Unterbrechung als `io.EOF` | 14 | 14 | 0 | |
| Leser: Anforderung ohne `len(p)`, ohne Obergrenze, Stückgröße 1, Puffer | 4 | 1 | 1 | grün: ohne Obergrenze (gleichwertig); zwei verworfen |
| Stichproben zu DoD 1 und 2 (Proxy, Kompression, `Frist`, `Content-Length`, Schließen, 101, Beenden der Frist, Weiterleitung, Host ohne Port) | 14 | 11 | 0 | drei verworfen |
| Summe | 133 | 94 | 33 | 6 verworfen |

**Die 8 grünen ohne Finding:** `0xC3` im Pfad von `Ziel`; `0x7F` einstellig im Pfad von `Ziel`; Vorabprüfung von `ctx` entfernt (`RoundTrip` beendet vor dem Verbinden mit demselben Kontextfehler, der Leser endet mit `ende`); `Header` `nil` statt leerer Map (der Port nennt „Header ohne Einträge“, beides gilt); Host und Port aus `Ziel` an jeden Fehler, in zwei Fassungen (erlaubt, wie A14 sagt); ein Fehlertext mit dem Zeiger des Lesers (künstlich); Stückgröße ohne Obergrenze (gleichwertig: `len(p)` bleibt Grenze).

**Sonden** (nicht committet, je auf dem Code grün und auf jedem Mutanten der Gruppe rot): `TestSondeZielVerstoss` (V1: sechs Ziel-Verstöße mit dem Marker im Pfad; beide Mutanten rot), `TestSondeKopfMarker` (V2: 101, `ctx` mit Kopf, Frist mit Kopf mit Marker in Status und Header, dazu der 101-Fall des Listeners; die fünf einzelnen Mutanten rot, der sechste ist ihre Summe), `TestSondeLaengeMax`, `TestSondeFristMax`, `TestSondeFristKlemme` (V3, V4: `Laenge` `math.MaxInt64` geht als `Content-Length: 9223372036854775807` auf die Leitung, `Frist` `math.MaxInt64` gültig, in der Blase `Frist` 24 h mit Kopf nach 23 h; alle fünf Mutanten rot), `TestSondeProzentHost` (V5: sieben Formen `ErrUngueltig`, `%25` und `%80` gültig; drei Mutanten rot), `TestSondeToken` (V6: alle 256 Bytes in der Mitte von `Methode` und Header-Name, Token genau nach RFC 9110; acht Mutanten rot).

## 4. Fixes zu F-1 bis F-10, Raten, Zeitläufe

**Alle Mutationen aus dem Review sind rot** (nachgebaut, nicht übernommen).

| Finding | Mutation | Ergebnis |
|---|---|---|
| F-1 | `< 0x22`, `0x7E` verboten, `0x80` verboten, dazu je Eingabe und je Byte (`0x21`, `0x7E`, `0x80`, `0xFF`); `0x7F` in `Pfad` und `RawQuery` angenommen; im Pfad von `Ziel` zweistellig | rot in `TestRequestZeile` und `TestSendenArgumente` aus dem Grund des Bytes |
| F-2 | nur `http` | rot: `gültiges Ziel "https://127.0.0.1:1"` |
| F-3 | Fehler von `Body` mit Kopf; 101 vor `ctx` und Frist; 101 zwischen `ctx` und Frist | rot: (j), (h), (i) |
| F-4 | Vorab-`ctx` als nicht erreichbar; mit `Pfad`, `RawQuery`, Header im Text | rot: `schon beendet, gültige Argumente` in `TestKontext` und `TestFehlerOhneWerte` |
| F-5 | `Ziel` an jeden Fehler | rot: 85 von 106, grün 21 wie der Plan sagt (Rest: V1) |
| F-6 | `Via` an jede Antwort; Eintrag im leeren Kopf | rot: `TestAntwort`, Fall (f) |
| F-7 | `/`, `:`, `@`, `[`, `{` als Token; dazu sechs innere Grenzen | rot an Methode und Header-Name (Rest: V6) |
| F-8 | erneut lesen mit `%25`, nur `%7F`, ohne Klammern, IPvFuture, Hostname ohne Auflösen | rot in `TestKonfigURL` (im Adapter Rest: V5) |
| F-9 | Kommentare von `geheim`, `mitGeheim`, `TestFehlerOhneWerte` | tragen jetzt nur, was die Fixtures belegen (Rest: V7) |
| F-10 | Dateitabelle | `ergebnis` samt 101 in `fehler.go`, wie der Code |

**Fänger mit Wahrscheinlichkeit.** A4 (c), letzter Fall, mit `slices.Sort(namen)` entfernt: 28 von 100 Prozessen rot (Plan 50 von 200, Review 42 von 200), unmutiert 0 von 100. A15, eine `http.Header`-Map für alle Aufrufe: 60 von 60 rot (Plan 200 von 200), unmutiert 0 von 60. Beide Raten stehen im Plan mit Zahl; die Größenordnung stimmt.

**Zeitläufe, kein Rot ohne Codefehler:** `-count=50` über `httpupstream` grün (6,7 s); Testbinary mit `-test.count=20` unter `--cpus=0.2` grün (3 min). Alle Fälle über Loopback und `host.invalid` enden unter dem Messwert von 10 ms (`go test -v`: 0,00 s).

## 5. Die zwei Punkte des Implementers

**(a) Prozent-Kodierung eines ASCII-Bytes außer `%25` und IPvFuture im Host von `Ziel`.** `SPEC-120` Nr. 2 trägt sie nicht ausdrücklich: das „wie `UPSTREAM_URL`, `SPEC-110` Nr. 7“ steht an der Regel der vier Bytes, und die Historie vom 2026-10-21 nennt für `Senden` nur sie. Der Plan trägt sie ebenfalls nicht als Zusage des Adapters (A19: „Dazu am Start die Form der Bibliothek“, die Zusatzzeile *Annahme* nennt nur `TestKonfigURL`); die Raster-Zelle `Ziel` · fremd nennt „Form der Bibliothek (*Ungeprüft*)“, und *Ungeprüft* sagt zu abgewiesenen Formen weder Folge noch Fänger. Im Adapter bleiben Mutanten grün, die diese Formen annehmen (erneut lesen mit `%25` im Host, nur `%7F`, IPvFuture: drei Mutanten, `go test` rc 0); die Sonde `TestSondeProzentHost` trennt sie. Der Code verhält sich wie der Start (Sonde grün). Das ist ein Finding, aber ein LOW, weil der Code die Entscheidung der Bibliothek überlässt und die Wirkung der Abweichung 502 statt 500 ist: V5.

**(b) `Ziel` mit `0x7F` im Pfad.** Der Plan benennt es. A2 nennt den Fall mit „das weist schon `url.Parse` ab“ und die Mutation „`0x7F` vor `url.Parse` aus `Ziel` genommen und in `ohneTrenner` angenommen (`Ziel` mit `0x7F` rot, dazu die anderen `0x7F`-Fälle)“, also an zwei Stellen. Gemessen: `0x7F` nur in `ohneTrenner` für den Pfad von `Ziel` angenommen bleibt grün; zusätzlich vor `url.Parse` herausgenommen ist rot im Fall `http://127.0.0.1:{p}/a\x7fb`. Gleichwertig im Sinne der Zusage, kein Finding (V11).

## 6. Größe

`git diff --numstat f76ba07..1fab875 -- src/` (Tests eingeschlossen): **2655** (2651 Einfügungen, 4 Löschungen); Produktcode 526, Tests 2129. Die Zahlen des Implementers stimmen; die Marke 2900 (§4) hält mit 245 Zeilen (8,4 %) Spielraum. Messpunkte nachgerechnet: `da6e121` 2527 (522, 2005), `0716718` 2813, `a736bb8` 2831 (566, 2265), `b6d5270` 2545 (522, 2023), `229d65a` 2567 (526, 2041), `6fdb5be` 2655; `6fdb5be` allein 140 (114, 26), davon 5 in `src/bootstrap/`, ab der Basis 88 (3,4 % auf 2567). §6 nennt jede dieser Zahlen und den Grund (Löschungen treffen Zeilen, die seit der Basis neu sind); §4 trägt die Marke mit der Entscheidung des Planners beim Handoff (Schnitt, der TLS-Teil nahm 286 Zeilen) und der Regel, dass die Marke nach dem Review weder angehoben noch gesenkt wird. Der Zuwachs im Review liegt unter der Spanne 15 bis 30 %, mit der §4 rechnete (nur Tests, keine neue Produktzeile); das Ende gegen den Stand nach dem Revert (2545) ist +4,3 %. Die Fänger zu V1 bis V6 kommen hinzu, geschätzt 60 bis 150 Zeilen (Schätzung, keine Messung): das Ende liegt dann unter 2900.

## 7. Cherry-Pick-Verträglichkeit

`git merge-tree --write-tree --merge-base 0716718^ 1fab875 0716718`: rc 0, kein Konflikt, Baum `aa00d73ab49370c70421b4a78d2aea332a13a5f4`. Der Diff von `1fab875` zu diesem Baum gleicht dem von `0716718` (vier Dateien: `client.go`, `fehler_test.go`, `tls.go`, `tls_test.go`; 291 Einfügungen, 5 Löschungen). Das Ergebnis, aus dem Baum exportiert, besteht `go vet` und `go test` über `httpupstream` und `bootstrap` mit `--network none`; `tlsFehlerFaelle` steht in `alleFehlerFaelle`. Die Änderung der Fänger an `fehler_test.go` (`6fdb5be`, drei Zeilen Kommentar) und die neuen Fälle in `kontextFehlerFaelle` berühren die Zeile nicht, die `0716718` ändert. `merge-tree` schreibt nur Objekte in den Objektspeicher, keine Referenz, keinen Commit. Die Fänger zu V1 bis V6 können `fehler_test.go` oder `client.go` berühren: der Lauf wiederholt sich nach dem letzten Commit des Slice (`slice-upstream-tls` §6, Risiko *Cherry-Pick gegen einen geänderten Rest*).

## 8. Plan gegen Code, Kommentare, Folge-Slices

**Plan gegen Code (`AGENTS.md` §3.9).** Dateitabelle in §3: jede Datei steht (`client.go`, `rahmen.go`, `leser.go`, `antwort.go`, `fehler.go`, `frist.go`; Tests mit den genannten Zeilen je Datei; 101 in `ergebnis` wie nach F-10). Kopf: `Bezug` und `Berührte Spec-Stellen` nennen den Stand (TLS-Teil und `CAPEM` ausgenommen, Host entschieden); der Kommentar von `Konfig` lautet wörtlich wie in §3 und stimmt vor und nach dem Cherry-Pick. §3 (A2, A3, A7, A12, A14, A16, A19) und §6 (Raster `Ziel`, Zusatzzeilen zu 101 und zum Fehler von `Body` mit Kopf, Annahme zu `url.Parse`, Messung) sind im Commit der Fänger nachgezogen. §4 trägt noch „Gemessen 2567“ als Stand vor dem Review; §6 trägt 2655: gewollt, kein Widerspruch.

**Kommentare (§3.7, §3.11).** Kein Konjunktiv über eine verworfene Alternative, kein abwesender Text; die Grenze des Lesers (`lauf`) nennt ihr fehlendes Test-Gegenstück. Drei Kommentare sagen mehr zu, als ein Test fängt: V7.

**Folge-Slices.** `slice-bootstrap-m2` §1 *Übernimmt (Upstream)* nennt die Sendung mit der Kennung `slice-upstream-adapter`, kein *Ausdrücklich NICHT* trifft sie, der Start verlangt `slice-upstream-adapter` und `slice-upstream-tls` in `done/` (§4). `slice-upstream-tls` §1 *Übernimmt* nennt den TLS-Teil mit dieser Kennung; dessen Ausschluss „der übrige Adapter“ zählt die Fänger A2 bis A18, nicht A19 (V10).

## 9. Lernpunkte für die Closure

**Finding-Klassen dieses Laufs:** Allaussage ohne Fänger über alle Fälle (V1, V2: der Marker steht in `mitGeheim`, aber der Fall überschreibt den Wert, oder der Fake trägt kein Feld, das der Fehlertext nennen könnte) · Zusage über das Fehlen einer Grenze ohne Fänger (V3, V4) · Komplement einer Menge nicht belegt (V6) · Randform mit Ort *Ungeprüft*, abgewiesene Form ohne Folge (V5) · Kommentar sagt mehr zu als die Fixture trägt (V7) · Befund ohne Träger (V8).

**Beobachtungs-Register (`docs/plan/planning/observations/BEO-REPO/`):**

| Beobachtung | Stand | Berührt durch |
|---|---|---|
| `vertrag-sagt-mehr-zu-als-ein-gate-prueft` | 16×, verkörpert in `AGENTS.md` §3.15 | Review F-1 bis F-8, dazu V1 bis V6; ein Beleg `slice-upstream-adapter` (ein Vorgang zählt einmal) |
| `universalaussage-im-vertrag-ungeprueft` | 5×, verkörpert in `AGENTS.md` §3.16 | Review F-4, F-5, dazu V1, V2; der Weg des Fehlers ist ein Paar *Fehlerweg × Wert*, das der Plan nicht aufzählt |
| `akzeptiertes-negativ-ohne-folge` | offen, 2× | V5 (Ort *Ungeprüft* ohne Folge für die abgewiesene Form) und V1 („grün bleiben die 21“ ohne Grund und Folge): erreicht mit diesem Vorgang 3×, die Closure weist einen Ausgang zu |
| `befund-an-rolle-ohne-artefakt` | 1× | V8 (I-4 des Reviews an den Architect); der Planner entscheidet, ob es zählt, denn der Architect führt Artefakte |
| `berichtigung-erreicht-nicht-jede-stelle` | 2× | V10, falls der Planner die Liste in `slice-upstream-tls` §1 als Auftreten zählt |
| `nebenlaeufiger-lauf-nur-im-glueck-gruen` | 5×, geplant `slice-race-gate` | kein neuer Beleg: kein Rot ohne Codefehler; Raten A4 (c) 28 von 100, A15 60 von 60 |
| `groessenmarke-ueberschritten-ohne-rueckfuehrung` | 11× | kein Beleg, die Marke hält; Datenpunkt für die Formel: nach dem Schnitt beim Handoff (2545) misst das Ende 2655, +4,3 %, nicht +25 % |

**Lehre für §3.15 und §3.16.** Bei einer Allaussage über Werte listet der Plan je Fehlerweg die Werte, die dieser Weg nennen könnte (Ziel, Pfad, Query, Header, Status, Header der Antwort, Body), und der Fänger belegt jede Zelle; ein Fall, der den Wert selbst setzt, trägt den Marker an der Stelle, die der Fehlertext nennen könnte. Für „keine Obergrenze“ und „jedes Byte, das kein Token ist“ ist der Fänger ein Fall am größten Wert beziehungsweise eine Schleife über alle 256 Bytes.

**Risiken in §6, die noch keinen Ausgang tragen** (jedes bekommt genau einen bei der Closure; die Belege dieses Laufs): Nebenläufigkeit (Raten, `-count=50`, `--cpus=0.2`, kein Rot ohne Codefehler; Datenrennen bleiben ohne `-race` ungeprüft, *weiter offen* oder *entfallen* mit dieser Begründung); Randformen (V5 ist eine Zelle mit Ort ohne Folge: *eingetreten* oder *entfallen* nach der Entscheidung des Architect); Annahmen über die Standardbibliothek (an Go 1.27.1 gepinnt, *weiter offen*); Loopback und Namensauflösung in der Stufe `test` (`127.0.0.1`, `host.invalid` und die Blase laufen ungecacht grün, *entfallen* für die Formen dieses Slice; `localhost` geht mit dem TLS-Teil an `slice-upstream-tls`); Record-Dienst gegen Fake, Adapter gegen `SPEC-120` (erster gemeinsamer Lauf erst im M2-Test, *weiter offen*); Größe (die Marke hält, *entfallen*, wenn sie nach den Fängern zu V1 bis V6 noch hält). Vier Risiken tragen schon einen Ausgang (Entscheidungen des Architect, akzeptierte Negative, Host, TLS-Block).

## 10. Gegenproben-Stempel und Gates

`make gegenproben` auf `1fab875`: „Eingaben unverändert seit dem letzten grünen Lauf (d0aa871b…)“, `git status` unverändert. `make gates` auf `1fab875`, vor diesem Bericht: rc 0. `make test` mit `--no-cache-filter test`: `pruefung test: ok` (46,6 s, gofmt, vet, alle Tests). `make lint` mit `--no-cache-filter lint`: „0 issues“ in allen Läufen von `lint.sh`, rc 0; kein `//nolint` unter `src/`, kein Gate, keine Konfiguration des Lints oder von a-check gelockert (`git diff f76ba07..1fab875` über `.golangci.yml`, `.a-check.yml`, `Dockerfile`, `Makefile`, `harness/mk/`: leer bis auf den Stempel der Gegenproben).

## Befunde

| ID | Kategorie | Befund | Quelle | Pfad | Zuständig |
|---|---|---|---|---|---|
| V1 | MEDIUM | Die Allaussage „höchstens Host und Port aus `Ziel`“ ist für den Fehler eines Verstoßes im Ziel nicht belegt. Die 21 Fälle, die `Ziel` selbst und ohne Marker setzen (Plan A14: „grün bleiben die 21“), überschreiben das Segment `/S3CR3T-ziel`, das `mitGeheim` anhängt; der Fall mit Zugangsdaten trägt den Marker im Passwort. Gemutet nennt `ungueltig("Ziel")` das Ziel mit `url.URL.Redacted()` (Passwort ersetzt, Pfad bleibt) oder den Pfad aus `Ziel`: `go test` rc 0, alle 106 Fälle grün. Die Gegenprobe Host und Port an jeden Fehler bleibt grün, wie A14 sagt. Die Sonde `TestSondeZielVerstoss` (Marker im Pfad von sechs Verstößen) ist auf dem Code grün und auf beiden Mutanten rot. | `SPEC-120` Nr. 8; `AGENTS.md` §3.15, §3.16 (Allaussage über alle Fälle) | `src/adapters/driven/httpupstream/client.go` · `return ziel{}, ungueltig("Ziel")` | Implementer (Marker im Pfad der Ziel-Verstöße), Planner (Zeile A14: „grün bleiben die 21“ ist kein Negativ mit Grund und Folge) |
| V2 | MEDIUM | „Kein Byte der Antwort“ ist für die Fehlerwege mit gelesenem Kopf nur im Body belegt: `fakeKopf` trägt keinen Status-Text und einen leeren Header, der 101-Fall des Listeners trägt `Upgrade: x` und den Marker erst nach dem Kopf, im Strom. Gemutet nennt der Fehler bei 101 `res.Status`, `res.Header` oder den Wert von `Upgrade` (drei Mutanten), bei `ctx` mit Kopf oder Frist mit Kopf Status und Header (zwei Mutanten) und bei 101 beide Texte zugleich: alle grün; der Mutant, der den Body in den Text liest, ist rot (c, d, h). Status-Text und Header-Werte sind Bytes des Upstreams. Die Sonde `TestSondeKopfMarker` trennt die fünf einzelnen, auch den 101-Fall des Listeners. | `SPEC-120` Nr. 5, 8; `AGENTS.md` §3.15, §3.16 (Allaussage) | `src/adapters/driven/httpupstream/fehler.go` · `} else if res.StatusCode == http.StatusSwitchingProtocols {` | Implementer |
| V3 | MEDIUM | `SPEC-120` Nr. 4 sagt für `Laenge` größer 0 „ohne Obergrenze“ zu, die Raster-Zelle `Body, Laenge` · fremd führt es; die Tabelle hat dazu keine Zeile, der größte Wert in den Tests ist 3. Gemutet weist `Senden` ein `Laenge` über 1 GiB oder über 1 KiB mit `ErrUngueltig` ab: `go test` rc 0 (zwei Mutanten). Der Code erfüllt die Zusage: Die Sonde `TestSondeLaengeMax` (`Laenge` `math.MaxInt64`, Kopf `Content-Length: 9223372036854775807`, `ErrUpstreamNichtErreichbar`, nicht `ErrUngueltig`) ist auf dem Code grün und auf beiden Mutanten rot (an der Zeitgrenze, weil sie erst auf den Listener wartet; ein Fall prüft den Fehler vor dem Warten). | `SPEC-120` Nr. 4; `AGENTS.md` §3.15 | `src/adapters/driven/httpupstream/client.go` · `case a.Laenge < -1:` | Implementer (Fänger), Planner (Zeile A5) |
| V4 | LOW | Die Raster-Zelle `New` · `Frist` führt „keine Obergrenze“ (Ort: Nr. 1 „kleiner als 1 ns“); `UPSTREAM_TIMEOUT` lässt nach `SPEC-110` Nr. 4 bis 63 Bit zu (`30d`). Die größte Frist in den Tests ist 1 h. Gemutet: `ErrUngueltig` über 24 h, über 1 h, und der Timer auf 30 min geklemmt: alle drei grün. Die Sonden (`Frist` `math.MaxInt64` gültig; in der Blase `Frist` 24 h, Kopf nach 23 h) sind auf dem Code grün und auf allen drei Mutanten rot. | `SPEC-120` Nr. 1; `SPEC-110` Nr. 4; `AGENTS.md` §3.15 | `src/adapters/driven/httpupstream/frist.go` · `return time.AfterFunc(c.frist, f).Stop` | Implementer |
| V5 | LOW | Prozent-Kodierung eines ASCII-Bytes außer `%25` und IPvFuture im Host von `Ziel` (Abschnitt 5 (a)): `SPEC-120` Nr. 2 und der Plan tragen sie nicht als Zusage des Adapters, die Zelle `Ziel` · fremd nennt „Form der Bibliothek (*Ungeprüft*)“, *Ungeprüft* sagt zur abgewiesenen Form weder Folge noch Fänger. Gemutet im Adapter (erneut lesen mit `%25` im Host, nur `%7F`, IPvFuture angenommen): drei Mutanten grün, beim Start rot (`TestKonfigURL`). Die Wirkung wäre `ErrUpstreamNichtErreichbar` (502) statt `ErrUngueltig` (500). Die Sonde `TestSondeProzentHost` trennt alle drei. | `SPEC-120` Nr. 2 und *Ungeprüft*; `SPEC-110` Nr. 7; `AGENTS.md` §3.16 | `src/adapters/driven/httpupstream/client.go` · `u, err := url.Parse(s)` | Architect (Nr. 2 zieht die Form mit oder nennt sie in *Ungeprüft* mit Folge), danach Implementer (zwei Fälle in `argumentFaelle`) |
| V6 | LOW | Die Nicht-Token-Zeichen sind nur zu einem Teil belegt: das Fix zu F-7 deckt die Nachbarn der Bereiche (`/`, `:`, `@`, `[`, `{`); die übrigen Trenner `"`, `)`, `,`, `;`, `<`, `=`, `>`, `?`, `\`, `]`, `}` und `0x7F` und die Bytes ab `0x80` als Token angenommen bleiben grün (acht Mutanten). Gemutet liefert eine solche Methode oder ein solcher Header-Name `ErrUpstreamNichtErreichbar` statt `ErrUngueltig`; `net/http` weist sie ab, ein Request erreicht den Upstream nicht. Der Kommentar von `istToken` sagt „HTTP-Token (RFC 9110)“ zu. Die Sonde `TestSondeToken` (Schleife über alle 256 Bytes) ist auf dem Code grün und auf allen acht Mutanten rot. | `SPEC-120` Nr. 2; `AGENTS.md` §3.11, §3.15 | `src/adapters/driven/httpupstream/client.go` · `func istToken(s string) bool {` | Implementer |
| V7 | LOW | Drei Kommentare sagen mehr zu, als ein Test fängt: der Kopf von `fehler.go` („Kein Fehler dieses Pakets nennt einen Wert der Anfrage oder ein Byte der Antwort“; V1, V2, V9 bleiben grün), `ungueltig` und `pruefe` („nennt das Feld, nie seinen Wert“; `Ziel` und `Methode`), `istToken` (V6). Das Fix zu F-9 hat `geheim`, `mitGeheim` und `TestFehlerOhneWerte` berichtigt, nicht die gleichlautende Zusage im Produktcode. | `AGENTS.md` §3.7, §3.11 | `src/adapters/driven/httpupstream/fehler.go` · `// Kein Fehler dieses Pakets nennt einen Wert der Anfrage oder ein Byte der Antwort` | Implementer |
| V8 | LOW | I-4 des Reviews (der Grund beim Start nennt die verletzte Regel zum Host nicht, Frage an den Architect) hat keinen Träger: weder §6 des Plans noch `SPEC-110` Nr. 7 und 12 noch die Welle führen die Frage oder einen Ausgang. `konfig_dienste.go` meldet weiter „keine absolute http- oder https-URL ohne Zugangsdaten, Query und Fragment“. | Review `7cecc39` I-4; Baseline-Regelwerk `modul-08-agentenrollen.md` | `docs/plan/planning/in-progress/slice-upstream-adapter.md` · `Randform ohne Ort, gemeldet vom Implementer beim Handoff` | Planner (Träger in §6 mit Ausgang), Architect (Antwort) |
| V9 | INFO | Der Wert der `Methode` steht in keinem Fehlertext der Fälle: ein Mutant, der ihn dem Fehler des Verstoßes anhängt, bleibt grün (die Fälle tragen keinen Marker in `Methode`). `SPEC-120` Nr. 8 nennt `Methode` in der Aufzählung der verbotenen Werte nicht, der zweite Satz („er nennt die Art und höchstens Host und Port aus `Ziel`“) schließt sie aus. Welche Lesart gilt, ist eine Frage an den Architect; danach trägt jeder Methoden-Fall den Marker. | `SPEC-120` Nr. 8 | `src/adapters/driven/httpupstream/client.go` · `return ziel{}, ungueltig("Methode")` | Architect |
| V10 | INFO | `slice-upstream-tls` §1 *Ausdrücklich NICHT* nennt für den übrigen Adapter die Fänger „A2 bis A9“, A11 und „A12 bis A18“, nicht A19 (nach dem Handoff, `229d65a`); die Aussage „ändert ihn nicht“ stimmt, die Aufzählung ist veraltet. | `AGENTS.md` §3.9 | `docs/plan/planning/open/slice-upstream-tls.md` · `A12 bis A18)` | Planner |
| V11 | INFO | Die zwei bewusst grünen des Implementers (Zahl 29 rot, 2 grün; die Liste der beiden steht in keinem Artefakt im Repo) und die Kandidaten, die ich geprüft habe: `0x7F` einstellig im Pfad von `Ziel` (gleichwertig, weil `url.Parse` vorher abweist; der Plan benennt es und die Zwei-Stellen-Mutation, Abschnitt 5 (b)); Vorabprüfung von `ctx` entfernt (gleichwertig: `RoundTrip` endet vor dem Verbinden mit demselben Kontextfehler, der Body wird nicht gelesen); `Header` `nil` statt leerer Map (gleichwertig nach dem Port). Jede der drei ist nach Quelltext und Lauf gleichwertig. | `AGENTS.md` §3.10 | `src/adapters/driven/httpupstream/client.go` · `if err := ctx.Err(); err != nil {` | — |
| V12 | INFO | Größe 2655 über die Pfade aus §3 (Abschnitt 6) gegen die Marke 2900. | `AGENTS.md` §3.14 | `docs/plan/planning/in-progress/slice-upstream-adapter.md` · `über **2900** Zeilen nach dem Review` | Planner |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `client.go`: `New`, `Senden`, `pruefe`, `leseZiel`, Bau von Request und Headern | Code nach Nr. 1 bis 3; Mutanten zu Port, Byte-Prüfung, Schema, Token-Bereichen, `User-Agent`, `Host`, Rahmung rot; Fänger V1, V3 bis V6 |
| `leser.go`, `rahmen.go`, `frist.go`: Leser, Rahmung, Frist | Code nach Nr. 4, 6; Unterbrechung als `io.EOF`, ohne Beenden der Frist, Frist über den Body, `Content-Length` entfernt: rot; ohne Befund |
| `fehler.go`, `antwort.go`: Rangfolge, Body-Hülle | die fünf Umordnungen von `rang` und sieben Mutanten in `ergebnis` rot; Fänger V2 |
| Revert `b6d5270` gegen `0716718`, Kommentar von `Konfig` und `New`, kein Import durch ein anderes Paket | ohne Befund |
| Start: Name des Hosts (A19) | beide Leser: alle Mutanten rot; ohne Befund (Rest im Adapter: V5) |
| Zwischenstand bis zum Cherry-Pick | akzeptiertes Negativ in §3 mit Grund und Folge; kein Lauf erreicht den Adapter (`src/bootstrap/` importiert nur `model`) |
| Übergaben (`AGENTS.md` §3.13) | `slice-bootstrap-m2` und `slice-upstream-tls`: Sendung mit Kennung des Gebers, kein Ausschluss trifft, beide in `open/`; V10 |
| Cherry-Pick von `0716718` | kein Konflikt, Ergebnis baut und besteht (Abschnitt 7) |
| Zeitabhängigkeit | kein Rot ohne Codefehler in `-count=50` und unter `--cpus=0.2` |
| Gegenproben, Gates, Suppression | Stempel aktuell; `make gates`, `make test` und `make lint` ungecacht rc 0; kein `//nolint`, kein Gate gelockert |

Für die Closure §7: V1 bis V6 sind Belege zu `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, V1 und V2 auch zu `BEO-REPO/universalaussage-im-vertrag-ungeprueft`, V1 und V5 zu `BEO-REPO/akzeptiertes-negativ-ohne-folge`, V8 zu `BEO-REPO/befund-an-rolle-ohne-artefakt` (nach Entscheidung des Planners).
