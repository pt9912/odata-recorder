# Verifikation slice-hybrid-service, Teil 1 (Mock zuerst und Replay-Fallback)

**Rolle:** Verifier (einmal). **Datum:** 2026-11-09. **Stand:** Zweig `slice-hybrid-service`, HEAD `e8345367` (Fix-Runde nach dem [Review](2026-11-09-slice-hybrid-service-review.md)). **Bezug:** [LH-FA-12](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), `SPEC-284`, `SPEC-261`.
**Plan:** [slice-hybrid-service](../plan/planning/done/slice-hybrid-service.md).

## Urteil

Die Zusagen des Teils 1 sind gehalten, und die Fix-Runde deckt vier der fünf Klassen. Keine Abweichung von Vertrag und Plan, die den Übergang nach `done/` hindert. Der Dienst erreicht in Teil 1 keinen Upstream: das ist strukturell (kein Port, kein Feld, kein `Ziel`) und mit einem eigenen Probe über 187 200 Eingaben belegt. Drei Auflagen vor oder bei der Closure (unten): ein Element der Klasse F-1 fehlt (Länge des Puffers 0, MEDIUM), die Sicherheitszusage hängt für die Standardbibliothek an einer Sperrliste (`net` kommt durch, MEDIUM), und ein *Join* mit zwei Fehlern mit Code hat keinen Fänger (LOW). Dazu ein Hinweis für die Integration mit `slice-if-match`, der nicht diesen Slice betrifft.

## Belege der Sensoren (selbst gefahren)

| Sensor | Ergebnis |
|---|---|
| `make gates` (einmal, unter `flock /tmp/odata-recorder-heavy.lock`, am HEAD, Last 2,9 bis 5,8) | Exit 0, alle Gates grün, darunter `lint`, `a-check`, `test-race`, `katalog-check`, `docs-check`, `groessenmarke`; `git status` danach sauber |
| `groessenmarke` | `slice-hybrid-service.md: ok, Messung 2166 von Marke 3900`; `Messpunkt: zeilen=26 nebenlaeufig=2 fakes=2 produktcode=330 schaetzung=2820 marke=3900 messung=2166` |
| eigene Messung `git diff --numstat 6de6da87..HEAD` über `src/hexagon/ports/inbound/hybrid*.go` und `src/hexagon/services/hybrid` | 2166 (Plan §3: 2166, Befund F-6 des Reviews erledigt); Produktcode 234 Zeilen in fünf Dateien (der Block sagt 330 als Erwartung, die Messung unterschreitet sie) |
| Liste des Implementers über `tools/harness/mutationslauf.sh` (83 Mutationen) | 79 rot, 4 grün (die benannten: Nullwert von `Code`, Probe ohne `ODR-E6002`, Probe ohne `ODR-E5001`, `Pfad` an `Entscheide`), 0 flatternd, 68,7 s, Last 3,9 bis 4,1 von 4 CPUs; entspricht dem Plan |
| Liste der Fix-Runde (14 Mutationen) | 13 rot, 1 Übersetzungsfehler (`f4-feld-typ-geaendert`, zählt nicht), 16,2 s; entspricht dem Plan |
| `go test -race -count=1` über Dienst, Policy und Ports (Image `odata-recorder:race`, `--network none`, `--memory=3g`, Kopie des Baums im Container) | grün |

Die Behauptung des Implementers (Gates grün, 2166, 79 von 83, 13 von 14) ist damit belegt, nicht nur zugesagt.

## Plan-vs-Code-Diff

- Berührt sind `src/hexagon/ports/inbound/hybrid.go` und `hybrid_test.go`, `src/hexagon/services/hybrid/` (`verteiler.go`, `puffer.go`, `antwort.go`, `fehler.go` und elf Testdateien), `spec/spezifikation.md` (eine Zeile: *Ungeprüft* 7 von `SPEC-284`), der Plan, `welle-hybrid.md`, `roadmap.md`, `slice-bootstrap-m4.md`, der neue Plan `slice-hybrid-service-upstream.md` in `open/` und der Review-Report. Keine Änderung an `Makefile`, `Dockerfile`, `harness/`, `tools/`, `.golangci.yml`, `.a-check.yml`, `src/bootstrap`, `src/adapters`, `src/hexagon/model/` (die Zahl der Fehlercodes bleibt 40) und an keinem anderen Pfad. Das Binary bleibt unverändert: der Handbuch- und Katalog-Punkt der DoD braucht in §7 die Zeile, dass es unverändert blieb.
- Das Paket importiert `bytes`, `context`, `errors`, `io`, `math`, `slices`, `model`, `inbound`, `policy`; kein Adapter, kein anderer Dienst, kein `outbound`.
- `SPEC-284` Nr. 1 nennt `Upstream` und `UpstreamURL`; der Code führt beide in Teil 1 nicht. Das ist die Entscheidung des Architect (20743ddd, leeres Ziel in Teil 1) und im Plan §1 und §3 so geschnitten; Teil 2 holt die Felder nach. Hinweis, kein Befund.
- Die Tabelle *Zusage · Fänger* hat 26 Zeilen (Z01 bis Z06, Z08 bis Z18, Z26, Z28 bis Z34, Z36). Jede Nummer von `SPEC-284` hat in Teil 1 eine Zeile oder ist als Teil 2 benannt: Nr. 1 Z01, Nr. 2 Z02 und Z11, Nr. 3 Z03 bis Z10, Nr. 4 (Dienstwahl, Abweichung (a), in Teil 1 ohne Wirkung), Nr. 5 Z12 bis Z16, Nr. 6 Z17 und Z18, Nr. 7 Z06 und Z29, Nr. 9 Z26, Nr. 10 Z28, Nr. 12 Z11, Nr. 13 Z30 bis Z32, *Ungeprüft* Z33, Z34, Z36.
- Gegengelesen auf Rohbytes (`LC_ALL=C grep -P '\\u[0-9A-Fa-f]{4}'`): alle `.go`-Dateien des Pakets und des Ports, kein Treffer.

## DoD gegen den Befund

| DoD-Punkt (Plan §2) | Ergebnis |
|---|---|
| Mock zuerst; Zuordnung unverändert an Mock und Replay | gehalten: Probe Sicherheit (14 Methoden × alle 40 Codes × 3 Pfade × 3 Antworten von R × 20 Policy-Zustände, 100 800 Fälle) ruft den Mock genau einmal, R genau dann, wenn Methode bekannt, Code Auslöser und Kategorie gesetzt (Orakel aus `SPEC-261` Nr. 2, 3, 4 b, nicht aus dem Code), und liefert sonst denselben Fehlerwert |
| Replay-Fallback nur für die Kategorien; Kennzeichnung `replay`; jeder Replay-Fehler unverändert; mutierend lässt den Mock unverändert | gehalten: derselbe Probe (Fehlerwert von R mit `==`, ein Wert `X-Recorder-Fallback: replay`); Mutation „Mock ein zweites Mal gerufen“ rot im Probe und in der Suite (Z29) |
| Body nach Nr. 5, Body der Anfrage nie geschlossen, verworfene Antwort genau einmal | gehalten: Probe Pufferung (86 400 Fälle) und Suite; siehe Sicherheit |
| `make gates` grün | belegt |
| Review, Report liegt vor | belegt (`2026-11-09-slice-hybrid-service-review.md`) |
| Doku-Update (Vertrag, Plan, Roadmap) | `SPEC-284` liegt, Plan §3 und §6 sind nachgezogen |
| Handbuch/Katalog | Binary unverändert; Zeile in §7 steht bei der Closure aus |
| Closure-Notiz, Register, Ausgang je Risiko, Paarungen | offen: §7 steht auf *Wird bei der Closure gefüllt*, die Häkchen sind offen (erwartet, Planner); §6 trägt je Risiko einen Ausgang oder *offen* mit Adresse (19 Ausgänge) |

## Sicherheit: kein Upstream-Zugriff in Teil 1

- **Strukturell.** Der `Verteiler` hält `mock`, `replay`, `policy`, `maxBytes`. Es gibt keinen `UpstreamClient`, kein `Senden`, keine `UpstreamURL`; jeder der Aufrufe von `Entscheide` trägt ein leeres `Ziel`, und `Entscheide` gibt für ein leeres `Ziel` nie `Upstream` zurück (`SPEC-261` Nr. 5 c). Jede Entscheidung außer `Replay` gilt als `Bleiben`. Selbst eine `Upstream`-Entscheidung der Policy führte zum Fehler von M zurück.
- **Eigener Probe** (`zz_probe_test.go`, nur im Scratchpad, nicht committet): Policies mit `Upstream` an und aus, Hosts `h`, `[::1]`, `x:80`, Routenfreigaben für alle fünf mutierenden Methoden, jede Kategorienmenge; Methoden `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `MERGE`, `OPTIONS`, `TRACE`, `CONNECT`, `get`, `post`, `Get`, `PROPFIND`; jeder Code von `model.AlleCodes()` als Fehler von M; R liefert Erfolg, `ODR-E5001` (Replay-Miss) und `ODR-E5002` (erschöpfter Key). Ergebnis: kein Eingabepaar erreicht mehr als M und R (M 1-mal, R höchstens 1-mal, bei `ODR-E5001` und `ODR-E5002` kehrt der Fehler als derselbe Wert zurück). Der Probe ist scharf: sechs von sieben Mutationen (Mock zweimal, Kennzeichnung `upstream`, Fallback über der Grenze, Leser von R das Original, Grenze ohne `+1`, `Laenge` -1 bei leerem Puffer) sind in ihm rot; die siebte (`!= policy.Replay` als `== policy.Bleiben`) ist in Teil 1 äquivalent (Befund F-8 des Reviews).
- **Mutationen auf den Pfad zum Netz.** Ein Import von `net/http`, `os`, `sync`, `time` oder `unsafe` ist durch `TestQuelltext` gesperrt, die übrige Standardbibliothek nicht (siehe V-2).
- **Body-Pufferung** (`SPEC-284` Nr. 5, eigener Probe, Menge aus dem Wortlaut: `MaxBytes` {1, 2, 4, 7} × Policy {keine, nur `coverage_insufficient`, nur `unsupported`, nur Host} × Code {`ODR-E6003`, `ODR-E6001`, `ODR-E6005`} × Länge des Bodys 0 bis 2·`MaxBytes`+2 × Deklaration {-1, 0, n, n-1, n+1, `MaxBytes`, `MaxBytes`+1, 2·`MaxBytes`+5, 2^40} × Fehler des Bodys {keiner, ein Fehler, `context.Canceled`} × Lesemodus von M {nichts, ein Byte, alles} × M mit und ohne Fehler): `Laenge` gleich der Pufferlänge für M und R, eigene Leser, gleiche Bytes, bei `MaxBytes`+1 oder Lesefehler kein Fallback und M erhält Bytes und Rest bzw. Fehler mit der deklarierten `Laenge`, bei `Laenge` über der Grenze ungelesen und unverändert, bei `Laenge` 0 nie gelesen, `Close` des Bodys der Anfrage 0. Der Code besteht alle 86 400 Fälle.

## Vertrag `SPEC-284` gegen den Code (Teil 1)

| Stelle | Ergebnis |
|---|---|
| Nr. 1 `New` (ohne `Upstream`) | gehalten; `Dienste` wird geprüft und nicht gespeichert |
| Nr. 2 Argumente und Kontext | gehalten (`ctx` nil, `Methode` leer, `Laenge` kleiner -1, `Body` nil bei `Laenge` größer 0; kein Wert im Text; Kontext unverändert an H2 und H4) |
| Nr. 3 a bis e | gehalten; Kontextfehler vor dem Code; `ctx.Err()` erst nach der Entscheidung und vor R |
| Nr. 5 a bis e | gehalten, siehe Probe; Lücke im Fänger: V-1 |
| Nr. 6, 7 | gehalten (dieselbe Map an M und R, `ODR-E5003` erst im Replay) |
| Nr. 9 | gehalten (neue Map, ASCII-Vergleich des Namens, Status und Body durchgereicht) |
| Nr. 10 | gehalten (genau ein `Close` bei Fehler mit Antwort, auch bei Kontextfehler; nie bei ausgelieferter Antwort) |
| Nr. 11, 12, 13 | gehalten (derselbe Fehlerwert; Stellen H2, H4; keine Sperre, keine Goroutine) |
| *Ungeprüft* 1 bis 7 | im Vertrag benannt; Plan Z33, Z34, Z36 tragen Grund und Folge |

## Tabelle Zusage · Fänger gegen den Vertrag

Jede Zusage von Nr. 1 bis 13 (Teil 1) hat eine Zeile mit Fänger oder ein benanntes Negativ. Zwei Negative des Plans (`Pfad` an `Entscheide`, Zahl der Aufrufe von `Entscheide`; zusätzlich `Ziel` ungleich leer an H3 und `Upstream` wie `Replay` behandelt) stehen mit Grund und Folge in §6 des Plans, nicht in *Ungeprüft* von `SPEC-284`; Teil 2 nimmt sie über Z19 bis Z21 auf (Befund F-8 des Reviews, Adresse Teil 2). Hinweis, kein neuer Befund. Das Negativ zum typisierten nil-Zeiger (Z36) ist in `SPEC-284` *Ungeprüft* 7 verankert; zu der Zeile in `spec/spezifikation.md` §7 (Historie) gibt es keinen eigenen Eintrag (INFO).

## Fix-Runde gegen die Klasse aufgelöst (`AGENTS.md` §3.17)

Eigene Mutationen (Kopie des Pakets samt Tests, Image `odata-recorder:race`, `--network none`, `--memory=3g`, `--cpus=4`), je mit Prüfung, dass das Muster die Quelle ändert. „Rot“ heißt `--- FAIL` eines Tests ohne Zeitlimit.

| Befund | Menge aus dem Wortlaut von `SPEC-284` | Eigene Mutation an einem Element, das weder Befund noch Fix nannte | Ergebnis |
|---|---|---|---|
| F-1 gelogene `Laenge` (Nr. 5 c: `Laenge` gleich der Länge des Puffers) | `MaxBytes` × Länge des Puffers *n* von **0** bis `MaxBytes` × Deklaration {-1, 1 bis `MaxBytes`, darüber} × Stufe {M, R}; der Fänger führt *n* ab 1 (122 + 38 Fälle) | R behält die deklarierte `Laenge`, wenn sie kleiner ist als der Puffer; M behält sie, wenn sie größer ist; R mit `Laenge` um 1 kleiner genau an der Grenze; `Laenge` genau 1 bei M; Grenze doppelt so hoch (`>` gegen `*2`) | rot (5 Mutationen) |
| F-1, Element *n* = 0 | leerer Body bei deklarierter `Laenge` größer 0 oder -1 (chunked ohne Inhalt) | R erhält `Laenge` -1 statt 0; M erhält die deklarierte statt 0; R erhält `Body` nil | **grün (3 Mutationen), V-1** |
| F-2 leerer `Root` | Liste der Länge 1 bis 5 × Position | geprüft wird nicht Position 2, nicht Position 4, nicht ab Länge 4 | rot (3) |
| F-3 verpackter Fehler (Nr. 3 b) | `errors.As` über Baum: direkt, `%w` ein- und zweifach, `Join` vorn, hinten, verschachtelt, × alle 40 Codes | nur eine Ebene `Unwrap` manuell; nur `Unwrap() error` ohne `Unwrap() []error`; Kontext nur als direkter Wert (`Canceled`, `DeadlineExceeded` einzeln); Kontextprüfung ganz entfernt | rot (6) |
| F-3, Element *mehrere Fehler mit Code im selben Baum* | `errors.As` liefert den ersten | den letzten `*inbound.Fehler` im `Join` nehmen | **grün, V-3** |
| F-4 Zustand im `Verteiler` (Nr. 13) | Felder des Typs (AST), Variablen und Methoden auf Paketebene | Feld `n int`; Feld mit Namen `_` vom Typ `[0]func()`; Feld `chan struct{}`; Feld `f func()`; Gruppenfeld `mock, zz`; eingebetteter Typ; Typ von `maxBytes` auf `int`; Variable `map` auf Paketebene; zweiter Typ mit Methode | rot (9, `TestVerteilerFelder` bzw. `TestQuelltext`) |
| F-4 Detektor je Prozess | 24 Prozesse, `-test.count=1`, Image `race`, Last 3,7 bis 5,8 von 4 | Variable auf Paketebene (Fehler), Feld `n++`, Variable auf Paketebene (Puffer) | mit Detektor 24 von 24, 24 von 24, 24 von 24; ohne Detektor 0 von 24 (kein Fenster); Grundlauf 0 von 24 mit und ohne Detektor |
| F-5 Frist im Fake von M | Aufrufstellen H2 und H4 | `WithoutCancel` an M, `Background` an M, `WithoutCancel` an R | rot als `--- FAIL` (`TestAbbruchZwischenStufen` nach 1,00 s, dazu `TestKontextAnStellen`, `TestGleichzeitigerAbbruch`), kein `panic: test timed out` |

Zwei Mutationen zählen nicht: die erste Fassung von „nur eine Ebene“ (äquivalent, `errors.As` blieb dahinter) und „beliebiger Auslöser im Join“ (Code-Nummern geraten, ohne Wirkung). Die Zeilenzahl der Mutationen: 33 gewertete eigene Mutationen (28 rot, 5 grün: drei zu V-1, eine zu V-3, eine zu V-2) und sieben im Probe (sechs rot, eine äquivalent).

## Befunde und Auflagen

| ID | Kategorie | Befund | Quelle | Pfad | Fänger |
|---|---|---|---|---|---|
| V-1 | MEDIUM | Die Klasse F-1 ist nicht vollständig: der Fänger `TestLaengeGegenBytes` beginnt bei einer Pufferlänge von 1. Das Element *leerer Puffer* (der Client sendet *chunked* ohne Inhalt, `Laenge` -1, oder deklariert mehr, als kommt) hat keinen Fall: R mit `Laenge` -1, M mit der deklarierten `Laenge`, R mit `Body` nil bleiben grün. Der Code ist richtig (Probe: 0 Abweichungen), nur die Zusage „`Laenge` gleich der Länge des Puffers“ ist für *n* = 0 ungefangen. | `AGENTS.md` §3.15, §3.17; `SPEC-284` Nr. 5 c | `klassen_test.go` · `for n := int64(1); n <= grenze; n++`, Zählprüfung `gezaehlt != 122 \|\| ueber != 38` | Mutationen V-1 a bis c, Suite grün |
| V-2 | MEDIUM | Die Zusage „in Teil 1 kein Upstream-Zugriff“ ist für die Standardbibliothek eine Sperrliste: `importErlaubt` lässt jedes Paket ohne Punkt im ersten Segment durch, `TestQuelltext` sperrt nur `sync`, `sync/atomic`, `os`, `net/http`, `time`, `unsafe`. Ein Import von `net` mit `net.Dial` im Replay-Miss bleibt grün (Mutation S1: Suite und `make a-check` sehen es nicht; `.a-check.yml` regelt die Standardbibliothek nicht). Im Ist-Zustand gibt es keinen solchen Zugriff; der Weg dorthin wäre eine Änderung durch einen Menschen oder Agenten, der Review fängt sie dann allein. | `LH-FA-12`; Plan §1 (der Zwischenstand ist der sichere); `AGENTS.md` §3.15 | `quelltext_test.go` · `importErlaubt`, `slices.Contains([]string{"sync", ...})` | Mutation S1, Suite grün |
| V-3 | LOW | Bei einem `errors.Join` mit zwei `*inbound.Fehler` gilt der erste (`errors.As`); die Mutation „der letzte“ bleibt grün. Die Arten von `TestVerpackteFehler` tragen genau einen Fehler mit Code neben Fehlern ohne Code. | `AGENTS.md` §3.17 (Menge „verbunden“ unvollständig) | `klassen_test.go` · `arten` | Mutation V-3, Suite grün |

**Auflagen für den Implementer bzw. Planner (vor `done/`):**

1. V-1: *n* = 0 in `TestLaengeGegenBytes` aufnehmen (deklariert -1 und 1 bis `MaxBytes`, M und R; `Body` von R nicht nil), Zählprüfung anpassen; eine Mutation an einem anderen Element (etwa `Laenge` des Puffers `MaxBytes` bei *n* = 0) rot sehen.
2. V-2: `importErlaubt` für die Standardbibliothek als Zulassungsliste (`bytes`, `context`, `errors`, `io`, `math`, `slices`) führen, oder die Lücke in §6 als akzeptiertes Negativ mit Grund und Folge eintragen. Entscheidung beim Planner.
3. V-3: `TestVerpackteFehler` um einen `Join` mit zwei Fehlern mit Code ergänzen (erster Auslöser, zweiter nicht, und umgekehrt), oder als Negativ benennen.

## Nebenläufigkeit Z31, Z32

Die Raten stehen in der Tabelle der Fix-Runde (Detektor je Prozess 24 von 24, Grundlauf 0 von 24). Die Zusage „kein geteilter Zustand“ ist damit für Datenrennen durch den Detektor gefangen; für geordnete Zugriffe (Kanal als Sperre) durch `TestVerteilerFelder` deterministisch, soweit der Zustand im Typ `Verteiler` steht, und durch `TestQuelltext`, soweit er auf Paketebene steht. Grenze des AST-Tests: gelesen wird nur `verteiler.go`; ein Feld in einem anderen Typ ist ein anderer Typ (`TestQuelltext` prüft Methoden außerhalb von `Verteiler` und `fehlerLeser`). Der Hinweis des Reviews zu F-4 ist damit geschlossen.

## Überschneidung

`git merge-tree --write-tree`:

- **gegen `main`** (79c21077): kein Konflikt (Baum 2f7766bb). `main` hat in `src/` nichts, was der Zweig nicht hat.
- **gegen `slice-if-match`** (e1a41ccc): ein Konflikt, rein textlich, in `spec/spezifikation.md` §7 (Änderungstabelle: beide fügen oben eine Zeile an). Auflösung „beide behalten“. Kennungen kollidieren nicht (`SPEC-283` dort, `SPEC-284` hier). Fehlercodes: `ODR-E6010` kommt aus `slice-if-match` (40 auf 41). Im Wegwerf-Merge (Archiv des Merge-Baums, Konflikt nur in der Doku) ist das Paket `hybrid` grün (`TestCodesSchleife`, Z03 und Z09 über `model.AlleCodes()` mit der Untergrenze 40), `model` und die Ports grün, `go vet` ohne Meldung.
- **Hinweis für die Integration, nicht dieser Slice:** im selben Wegwerf-Merge sind zwei Tests des Pakets `policy` (aus `slice-fallback-policy`, auf `main`) rot, weil sie die genaue Zahl 40 prüfen (`entscheide_test.go:391`, `total_test.go:193`: „`model.AlleCodes()` liefert 41 Codes, erwartet 40“). Dieser Zweig berührt `policy` nicht; der Zusammenstoß gehört zu `slice-if-match` gegen `main` und tritt für den zweiten Integrierenden auf. Adresse: Planner (Folge in der Closure von `slice-if-match` oder in `slice-fallback-policy`-Tests).

## Drei Fragen an den User

Nicht geprüft als Befund (Planner-Vorgabe): der Replay-Fallback auch für mutierende Methoden gilt bis zur Antwort (`SPEC-261` Nr. 4 b); der Probe fährt ihn, M wird dabei genau einmal gerufen. `ODR-E6006` ist kein Auslöser (Probe: kein Fallback, derselbe Fehlerwert).

## Offene Punkte

Die Häkchen der DoD stehen alle offen und §7 ist leer: erwartet, sie gehören in die Closure. Zu belegen sind dort: Gates grün (hier belegt), Review (liegt vor), Verifikation (dieser Report), die Zeile in §7, dass das Binary unverändert blieb, das Beobachtungs-Register (`BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` trifft V-1 erneut), der Ausgang jedes Risikos und die Paarungen.

Scratchpad (nicht committet): `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/ver-hy/` mit Probe, Mutationslisten, Skripten und Ausgaben.
