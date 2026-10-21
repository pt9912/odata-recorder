# Verifikation: slice-record-service

* Datum: 2026-10-21
* Rolle: verifier (frischer Kontext; kein Code, keine Spezifikation, kein Plan, kein Lastenheft, keine `AGENTS.md` geändert)
* Gegenstand: `HEAD` `166332f`, Verlauf `7637843` (Vertrag) → `51bffa1` (Schnitt) → `647f988` → `0f1d866` (Implementierung) → `6c08a67` (Planner) → `485682d` (Review) → `567abe7` (Architect) → `166332f` (Implementer)
* Plan: `docs/plan/planning/in-progress/slice-record-service.md`
* Maßstab: DoD (§2) und Tabelle *Zusage · Fänger* (§3) gegen `SPEC-121`, Kopf und Nr. 4, 5, 8 von `SPEC-120`, `SPEC-109` Nr. 5 und die Nachzüge `LH-FA-01.a`, `LH-FA-02.b`, `LH-FA-02.c`, `LH-FA-13.a`, `LH-QA-07.a`, `LH-RB-02.a`, `SPEC-005`, `SPEC-012`, `SPEC-053`, `spec/architecture.md` §1, §2, §2.2, §4.1, §4.9; Lastenheft [LH-FA-01](../../spec/lastenheft.md), [LH-FA-02](../../spec/lastenheft.md), [LH-FA-13](../../spec/lastenheft.md), [LH-QA-07](../../spec/lastenheft.md) (nur gelesen); `AGENTS.md` §3.7, §3.9 bis §3.16; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)
* Belege selbst gefahren: 18 eigene Mutationen auf Kopien von `src/` (Abschnitt 3), sieben Wegwerf-Sonden zu den überlebenden Mutanten und zum Vorschlag des Implementers (Abschnitte 3 und 4), fünf Zeitläufe (Abschnitt 4), `git diff --numstat` (Abschnitt 9), `make gegenproben`, `make gates` auf `166332f` vor diesem Bericht (Abschnitt 10). Jedes `go test` lief im per Digest gepinnten Image der Stufe `deps` des Dockerfiles (`golang@sha256:8a5910…`, Go 1.27.1), `--network none`, die Quelle schreibgeschützt, Mutanten in Kopien im Scratchpad; der Arbeitsbaum blieb unberührt (`git status` leer nach jedem Lauf).

## Urteil

**Nicht freigegeben: kein HIGH, vier MEDIUM.** Der Code ist nach allen Läufen richtig: jede Sonde ist auf `166332f` grün, kein Zeitlauf wird ohne Codefehler rot. Die Befunde sind Lücken der Fänger, ein nicht benanntes Fenster und eine fehlende Spur.

* MEDIUM (Implementer, Fänger): V1 Regel aus F-18 nur mit `br` und in einer Header-Zeile belegt; V2 Bericht „je Session“ ohne trennenden Fall; V3 „die Zahl bleibt“ nach einem Fehler von `ListBySession` ohne trennenden Fall; V4 Anhänge verschiedener Sessions „unabhängig“ ohne Fänger, die Zusatzzeile nennt B18 und B27, die es nicht prüfen.
* LOW: V5 (Architect) die Drain-Seite des Fensters aus `SPEC-121` Nr. 2e ist weder Fänger noch benanntes Negativ; V6 (Planner) I-4 des Reviews hat keinen Träger im Repo.
* INFO: V7 Bewertung des Haltepunkts am ersten `Done()`; V8 Grenze von `testing/synctest`; V9 Plan-Wortlaut neben dem Code; V10 Wortwahl „gehört“ neben „besitzt“; V11 Größe 4148 über der Marke 3600 (Entscheidung beim Planner).

Jeder der sechs grünen Mutanten ist nicht äquivalent: eine trennende Sonde ist auf dem Code grün (20 Läufe) und auf dem Mutanten rot (Abschnitt 3).

## 1. DoD Punkt für Punkt

| DoD | Stand | Beleg |
|---|---|---|
| Port `UpstreamClient` samt Fehlern und ihrer Zeile in `SPEC-109` Nr. 5 | erfüllt | `upstream_client.go` gleich dem Kopf von `SPEC-120` (Typen, Felder, Signatur), `errors.go` mit `ErrUpstreamNichtErreichbar`, `ErrUpstreamZeit`, Zeile in `abbildung.go` und `TestAbbildungTabelle` (beide → `ODR-E4001` für jeden Gegenstand), Wortlaut `SPEC-109` Nr. 5 gleich |
| `Record` nach Nr. 1 bis 9, 11, 12, 14, 15 | Code erfüllt; Fänger lückenhaft in Nr. 4 (V1) und Nr. 15 (V4) | Prüfreihenfolge `ODR-E5005` → `ODR-E5006` → Session → Annahme unter einer Sperre (`nimmAn`), Kennung und Zeit danach, Header nach Nr. 4, Aufnahme nach Nr. 5, Ende nach Nr. 8 in der Reihenfolge (a) bis (g), ohne Antwort nach Nr. 11 samt Rangfolge beim Liefern, Protokoll nach Nr. 12, `*Fehler` nach Nr. 14 |
| Anhängen, Fault Injection und `Drain` nach Nr. 10, 13 | Code erfüllt; Fänger lückenhaft in Nr. 10 (V3, V4) und Nr. 13 (V2) | `haengeAn` (Zahl beim ersten Anhängen und nach `ErrUngueltigeSequenz`, Sperre je Session), `endeLauf` (Bericht, `leer`), `Drain` (Sperre dauerhaft, Frist, Abbruch über den Kontext, Warten ohne Frist), `context.WithoutCancel` für das Anhängen |
| `make gates` grün | erfüllt | eigener Lauf auf `166332f`, rc 0 (Abschnitt 10) |
| Review durchgeführt, Report liegt vor | erfüllt | `docs/reviews/2026-10-21-slice-record-service-review.md` (`485682d`) |
| Doku-Update | erfüllt | `SPEC-120`, `SPEC-121`, `SPEC-109` Nr. 5 und die Nachzüge (Abschnitt 2) |
| Closure-Notiz, Register, Ausgänge, Paarungen | offen, beim Planner | §7 „Wird bei der Closure gefüllt“; §6 trägt noch drei Ausgänge *offen* |

## 2. Vertrag unabhängig gelesen, Nachzüge, Widersprüche

`SPEC-121` gegen den Code, Satz für Satz: Kopf (Typen, `record.Recorder`, Besitz der `Anfrage`, Schreibweise der Header-Namen), Nr. 1 bis 15 und *Ungeprüft* stimmen mit `recorder.go`, `wahl.go`, `header.go`, `ende.go`, `anhaengen.go`, `drain.go`, `protokoll.go` überein. Zwei Sätze geprüft, die nur aus dem Code der Pipeline folgen: „dorthin führt nur ein `Body` nil aus `Senden`“ (*Ungeprüft*, Fehler von `Tee`) trägt `pruefeEingang` in `capture/pipeline.go` (Fehler nur bei `ctx` nil, `src` nil, falscher Richtung, `Metadata` beim Request; der Dienst setzt alle außer `src` der Antwort gültig); „`Abschluss` scheitert hier nicht“ (Nr. 8) trägt `capture/abschluss.go` (Fehler nur bei falschen Richtungen). Die Raster-Zelle `Record` · `Methode` · unbekannt („`Senden` liefert `ErrUngueltig`“) trägt `SPEC-116` Nr. 2: `BildeSchluessel` prüft die Methode nicht.

Der Fake des Ports gegen `SPEC-120` Nr. 4, 5, 8: Lesen des Bodys nach Skript bis `io.EOF` oder `n` Bytes, `Laenge` 0 liest nicht; `Body` nie nil bei einer Antwort, Lesen an `ctx` gebunden; Kontextfehler vor dem Fehler von `Body` vor dem Fehler des Plans. Abweichung ohne Wirkung: einige Antworten tragen kleine Header-Namen (V9).

Widersprüche normativer Sätze, gesucht in `spec/` (Lastenheft nur gelesen) und in §1, §6 und DoD der fünf Nehmer nach `wirkungslos`, `incomplete`, `abgebrochen`, `Teil-Exchange`, `Gateway`, `Client-Abbruch`, `Accept-Encoding`, `Hop-by-Hop`, `UPSTREAM_TIMEOUT`, `sequence_no`, `Beginns`:

| Stelle | Aussage | gegen `SPEC-121` |
|---|---|---|
| `LH-FA-01.a` Nr. 1, 2, Fehlermodi | Session-Wahl `ODR-E5009`; Header nach Nr. 4, 7; 504 bei Zeit, sonst 502, nach Client-Abbruch kein Empfänger | Nr. 2d, 4, 7, 11: gleich |
| `LH-FA-02.a` Nr. 4, 5; `LH-FA-02.b`; `LH-FA-02.c` | ohne Antwort angehängt, `incomplete` oder `failed`; `complete`, wenn beide Bodies vor dem Fehler oder Abbruch vollständig erfasst waren | Nr. 11, 13: gleich |
| `LH-FA-13.a` Status, `close` | Header wählt immer; `close` auf beendeter Session `ODR-E7011` ohne Änderung; abgebrochene Exchanges mit dem Status aus ihren Ergebnissen; `complete` nur ohne unangehängten | Nr. 2d, 13; `SPEC-005`, `SPEC-111`: gleich |
| `LH-QA-07.a` | Client des Vergleichstests nur `gzip`, `deflate`, `identity`; jedes Coding ausgeschlossen: direkt 406, über den Proxy ohne Coding | Nr. 4: gleich (die Frage an die Messmethode ist I-4, V6) |
| `LH-RB-02.a`, `SPEC-012`, `SPEC-053`, `SPEC-109` Nr. 2 und 5 | Zertifikatsfehler 502; Frist bis zum Kopf; Abbruch nach vollständiger Erfassung `complete`; `ODR-E4001` mit 502 oder 504 | `SPEC-120` Nr. 6, 7, 8; Nr. 11, 13: gleich |
| `architecture.md` §1, §2, §2.2, §4.1, §4.9 | kein Typ einer HTTP-Bibliothek im Core und in Port-Signaturen; Port-Zeile und Sequenzen nach Nr. 7, 13 | gleich; keine Welle, kein Slice, kein Hash, keine ADR (`AGENTS.md` §3.4); §4.1 ordnet „Persist“ vor „Response“, der Satz unter dem Diagramm nimmt die Zeitordnung aus |
| `LH-FA-02.b` „besitzt … geliehene … Streams“ | der Adapter besitzt die Ströme und leiht sie | Kopf: die `Anfrage` „gehört … dem Dienst“; inhaltlich dieselbe Leihe, Wortwahl verschieden (V10) |
| Lastenheft [LH-FA-01](../../spec/lastenheft.md), [LH-FA-02](../../spec/lastenheft.md), [LH-QA-07](../../spec/lastenheft.md) | Gateway-Fehler und Protokoll; Speicherfehler sichtbar, kein vorgetäuschter Austausch; keine fachliche Veränderung | kein Widerspruch; Lastenheft und `AGENTS.md` seit `7637843^` unverändert (`git diff --stat` leer) |

Kein Paar normativer Sätze widerspricht einander.

## 3. Tabelle *Zusage · Fänger*, Raster und eigene Mutationen

**Existenz.** Jeder Fänger aus B1 bis B31 steht unter seinem Namen im Code; jede Zelle der 15 Raster-Zeilen nennt ihren Ort oder „gilt nicht“ mit Grund; jeder Fänger der 38 Zusatzzeilen steht im Code, drei Zusatzzeilen nennen ausdrücklich keinen (akzeptierte Negative: `Body` nil aus `Senden`, Änderung der `Anfrage`, Fehler von `Abschluss`). Keine Zelle ist leer. Die Eingaben der Tabelle stehen in den Tests, wie §3 sie nennt, mit zwei Abweichungen im Wortlaut (V9). `TestImporte` liest 18 Dateien aus `record`, `inbound` und `outbound` (Grenze 18 erreicht).

**Verfahren.** Ein Go-Programm im gepinnten Image kopiert `HEAD` je Mutante, ersetzt genau einmal (sonst bricht es ab), fährt `go test -count=1 -timeout=300s` über das Paket `record`. Alle 18 bauten, keine hing; M11 endet an der Zeitgrenze der Tests (10 s) und an der Erkennung des Stillstands von `synctest`. Gezählt wird nur Rot aus dem benannten Grund. Keine Mutation ist wörtlich aus §3 des Plans oder aus dem Review-Bericht übernommen; M14 und M15 sind Varianten zweier dort genannter (Fehlertext und Pfad an anderer Stelle des Eintrags).

| # | Zusage (Ort) | Mutation | Ergebnis |
|---|---|---|---|
| M01 | Nr. 4: entfällt ein Glied mit nichtleerem Coding, entfallen die Glieder `identity` | `*` löst das Entfallen nicht aus | **grün** (V1); Sonde `*, identity;q=0` → `identity` rot |
| M23 | ebenso | nur `br`, `zstd`, `compress` lösen aus | **grün** (V1); Sonde `x-gzip, gzip, identity;q=0` → `gzip` rot |
| M02 | ebenso, über alle Werte | Entfallen je Header-Zeile statt über alle | **grün** (V1); Sonde Zeilen `br` und `identity;q=0` → `identity` rot |
| M03 | ebenso | Glieder `identity` entfallen nur nach dem entfallenen Glied | rot: `TestAcceptEncoding`, `identity;q=0, br` |
| M25 | Nr. 2d: mehrere offene, nicht im `Drain` | `ODR-E5003` aus den Kandidaten vor dem Filter auf `Drain` | rot: `TestSessionOhneHeader`, Fall „s2 im Drain“ |
| M05 | Nr. 11: Rangfolge beim Liefern, Kontext zuerst | Kontext des Clients vor dem Anhängen festgehalten | rot: `TestOhneAntwortRangfolge`, je Port-Fehler Fall Client (Eintrag `upstream`, Code statt Kontextfehler) |
| M06 | ebenso, `Drain` vor den Fehlern des Ports | Abbruch durch `Drain` vor dem Anhängen festgehalten | rot: `TestOhneAntwortRangfolge`, Fall Drain (502 bzw. 500, `grund` nicht `drain`) |
| M08 | Nr. 13: `Abgebrochen`, deren Kontext ein `Drain` beendete, jeden einmal | Zählung beim Abbruch je laufendem Lauf statt am Ende | rot: `TestDrainBericht` (Client-Fall 1 statt 0), `TestDrainWiederholt` (2 statt 1) |
| M09 | Nr. 13: Bericht bei der Rückkehr | Stand vom Beginn von `Drain` geliefert | rot: `TestDrainAbbruch`, `TestDrainBericht`, `TestDrainWiederholt`, `TestAnhaengenNachAbbruch` |
| M10 | Nr. 13: der Bericht zählt je Session | ein Zähler für den Dienst | **grün** (V2); Sonde rot (`Drain(s1)` meldet den Verlust aus `s2`) |
| M11 | Nr. 13: Frist `0` heißt nicht warten, dann abbrechen | Frist `0` kehrt ohne Abbruch zurück | rot: `TestAnhaengenNachAbbruch`, `TestDrainBericht` (Kontext endet nicht), `TestDrainWiederholt` |
| M12 | Nr. 2e, 13: `Drain` setzt die Sperre und sieht die laufenden in einem Schritt | laufende gezählt und `Drain` gesetzt in zwei Sperrabschnitten | rot mit Wahrscheinlichkeit: `TestAnnahmeGegenDrainGleichzeitig` in 12 von 20 und 99 von 100 Läufen (V5) |
| M14 | Nr. 12: kein Fehlertext | Fehlertext des Ports im Attribut `grund` | rot: `TestKeinGeheimnis` (Marker, geschlossene Menge), `TestProtokoll` |
| M15 | Nr. 12: kein Pfad, fester Text | OData-Pfad im Text des Eintrags `upstream` | rot: `TestKeinGeheimnis` (`PFAD-7f`, fester Text) |
| M18 | Nr. 12: je `Ergebnis.Fehler` ein Eintrag | nur der Request wird protokolliert | rot: `TestResponseAufnahme`, `TestNichtAngehaengt`, `TestProtokoll` |
| M19 | Nr. 10: Zahl beim ersten Anhängen gelesen | `ListBySession` vor jedem Anhängen | rot: `TestSequenz` (`ListBySession` 2) |
| M20 | Nr. 10, 15: Anhänge an verschiedene Sessions unabhängig | eine Sperre der Anhänge für alle Sessions | **grün** (V4); Sonde rot (`Close` von `s2` wartet hinter dem angehaltenen `Append` von `s1`) |
| M22 | Nr. 10: scheitert das Lesen der Zahl, bleibt sie | Zahl vor der Fehlerprüfung gesetzt (0, bekannt) | **grün** (V3); Sonde rot (nach dem Lesefehler trägt der nächste Exchange 1 statt 4 und geht verloren) |

Ergebnis: 18 gelaufen, 12 rot (M12 nur mit Wahrscheinlichkeit), 6 grün. Die Sonden (`TestSondeAE`, `TestSondeBerichtJeSession`, `TestSondeZahlNachLesefehler`, `TestSondeAnhaengeUnabhaengig`) liefen auf `166332f` mit `-count=20` grün und je auf ihrem Mutanten rot; sie sind nicht committet.

## 4. Fänger mit Wahrscheinlichkeit, Zeitläufe, Haltepunkt am ersten `Done()`

**Benannt.** B19 (`TestSequenzGleichzeitig`) und `TestAnnahmeGegenDrainGleichzeitig` (B8) stehen in §3 (Abschnitt Race-Detektor und Tabelle), in der Zusatztabelle und in `SPEC-121` *Ungeprüft* als „nicht mathematisch sicher“; die Kommentare der Tests sagen dasselbe. B16 (a) ist durch den angehaltenen `Commit` deterministisch (F-13 erledigt).

**Rot ohne Codefehler: keines.**

| Lauf | Ergebnis |
|---|---|
| Paket `record`, `-count=50` | grün, 4,1 s |
| Testbinary, `-test.count=50`, `--cpus=0.05` | grün, 66 s |
| `TestAnnahmeGegenDrainGleichzeitig`, `TestSequenzGleichzeitig`, `TestEnde`, `TestAnnahmeGegenDrain`, `-test.count=500` | grün |
| die beiden Fänger mit Wahrscheinlichkeit, `GOMAXPROCS=1`, `-test.count=200` | grün |
| dieselben, `--cpus=0.05`, `-test.count=200` | grün, 3 min 27 s |

**Vorschlag des Implementers** (Haltepunkt am ersten `Done()` des Aufrufer-Kontexts in `context.WithCancelCause`), geprüft mit einer Sonde (`haltCtx`: `Done` hält beim ersten Aufruf an; währenddessen `Drain(s1, 1 h)`, danach Freigabe, erwartet `ODR-E5009`):

| Lauf | Ergebnis |
|---|---|
| Sonde auf dem Code | grün in 50 von 50 |
| Sonde gegen „Prüfung auf `Drain` in eigener Sperre nach dem Lesen der Session“ (die Mutation aus F-1) | rot in 50 von 50 |
| Sonde gegen „Sperre in `nimmAn` zwischen Prüfung und Annahme freigegeben“ | grün in 50 von 50 |
| `TestAnnahmeGegenDrainGleichzeitig` gegen dieselbe enge Mutation | rot in 10 von 50 |

Bewertung (die Entscheidung hat der Architect, V7): Der Haltepunkt trägt für jede Lücke, die den Aufruf von `context.WithCancelCause` in `Record` umschließt, deterministisch und ohne Test-Haken im Produktcode. Er trägt nicht für eine Lücke nach diesem Aufruf (in `nimmAn`), und er hängt an zwei Dingen, die kein Vertrag zusagt: an der Reihenfolge in `Record` (Ableitung des Kontexts zwischen `pruefe` und `nimmAn`) und daran, dass `context.WithCancelCause` `parent.Done()` synchron ruft (Implementierung von `propagateCancel`, keine dokumentierte Zusage); ein Fake, der `Done()` früher ruft, verschiebt den Halt. Die Kopplung an die Reihenfolge lässt sich sichtbar machen: rückt die Ableitung hinter die Annahme, wartet `Drain` im Halt, und die Sonde scheitert an ihrer Zeitgrenze. Mit dem *Ungeprüft*-Satz verträgt er sich im Wortlaut („kein Aufruf eines Ports“, „kein Test-Haken“ bleiben wahr: `Done` ist ein Aufruf am Argument), nicht in der Folgerung: „fangen … nur mit Wahrscheinlichkeit“ gälte dann nur noch für den Teil des Fensters nach der Ableitung des Kontexts. Übernimmt der Architect den Haltepunkt, sind *Ungeprüft*, §3 (B8, akzeptierte Negative) und die Zusatzzeile in diese zwei Teile zu trennen.

## 5. `testing/synctest`

Im gepinnten Image Go 1.27.1, `GOEXPERIMENT` leer; Dockerfile, Makefile und die Fragmente unter `harness/mk/` setzen es nicht. `synctest.Test` und `synctest.Wait` sind seit Go 1.25 allgemein verfügbar, Standardbibliothek ([ADR-0001](../plan/adr/0001-implementierungssprache-go.md)). §3 benennt die Grenze Blasenzeit statt Wanduhr („die Fristen (`20 ms`, `1 h`) laufen auf der Uhr der Blase“, „Rot an der Zeitgrenze heißt in der Blase …“); jede Blase baut ihre Umgebung mit `neu(t)` innerhalb von `synctest.Test`. Nicht benannt: ein Hang an einem `sync.Mutex` gilt nicht als dauerhaft blockiert, `synctest.Wait` wartet dann bis zum Timeout von `go test` (V8, wie I-2 des Reviews).

## 6. Plan gegen Code, Kommentare, akzeptierte Negative

**Plan gegen Code (`AGENTS.md` §3.9).** Jede Datei, jeder Typ und jeder Testname aus §3 steht im Code; `export_test.go` gibt genau die vier reinen Funktionen frei; kein Test-Haken in `record`. Abweichungen im Wortlaut: V9.

**Kommentare (§3.7, §3.11).** Die Klassen sind getragen; kein Konjunktiv über eine verworfene Alternative. Die „nicht erreichbar“-Kommentare nennen ihren Ort: `ohneTee` („kein Test erreicht den Zweig (SPEC-121 Ungeprüft)“), `schliesseAb` (`Abschluss` auf Nr. 8, zweiter `BildeSchluessel` auf Nr. 6; F-16 erledigt). Zwei Kommentare sagen mehr zu, als ein Test fängt: `acceptEncoding` („entfällt ein Glied mit nichtleerem Coding“, V1) und `DrainBericht` in `inbound/record.go` („zählt je Session“, V2).

**Akzeptierte Negative, je mit Grund und Folge:**

| Negativ | Grund | Folge | Ort |
|---|---|---|---|
| F-1: Fenster zwischen Rückkehr des Lesens der Session und Annahme | kein Aufruf eines Ports im Fenster, kein Test-Haken | Annahme nach der Rückkehr von `Drain`, Session womöglich `complete` gegen Nr. 13 | §3, *Ungeprüft* |
| F-12: `Body` nil aus `Senden` | Bruch von `SPEC-120` Nr. 5, A7 (d) fängt „nie nil“ | mit Body: `Tee` scheitert; ohne Body: Panic im Handler, `Drain` wartet ohne Ende | §3, *Ungeprüft* |
| F-16: Fehler von `Abschluss` | prüft nur die Richtungen | nicht angehängt, protokolliert | §3, Nr. 8 |
| F-17: Adapter ändert die `Anfrage` | Pflicht des Adapters, Sendung | Key aus geänderten Headern, fehlende Bytes, Rennen | §3, Kopf, *Ungeprüft* |
| B19 und Datenrennen | kein `-race` (`CGO_ENABLED=0`) | falsche `sequence_no` (nicht angehängt), Annahme nach `Drain`, falscher Bericht, Ende des Prozesses | §3, *Ungeprüft* |
| Fehler von `Tee` | nur über `Body` nil | nicht angehängt, Abbruch der Request-Aufnahme, `ODR-E9001` mit 500 | §3, *Ungeprüft* |

Jedes trägt Grund und Folge. Nicht benannt ist die Drain-Seite des Fensters aus Nr. 2e (V5).

## 7. Review-Findings

| Finding | Behandlung | Stand |
|---|---|---|
| F-1 | Wortlaut Nr. 2e, Negativ mit Folge, `TestAnnahmeGegenDrainGleichzeitig` | erledigt; Drain-Seite V5, Haltepunkt V7 |
| F-2, F-3, F-4, F-5, F-6 | Status 201, 304, 404, 502; Methode `patch`; `FinishedAt` in der Blase und ohne Antwort; `Laenge` der gzip-Fälle; zwei Werte von `Content-Type` und `Content-Encoding` je Richtung; `/$metadataX`, `/$metadata/x`, `/x/$metadata` | erledigt |
| F-7 | (a) `connection` klein; (b) `content-encoding` klein; (c) `Connection: accept-encoding`; (d) Glieder `identity`; (e) Tabulator vor `;` | erledigt; (d) nur mit `br` als Auslöser (V1) |
| F-8, F-9 | `TestOhneAntwortRangfolge`, `errors.Join` beider Port-Fehler; `TestDrainBericht` | erledigt |
| F-10 | geschlossene Mengen in `pruefeWerte` | erledigt |
| F-11 | `TestImporte` liest `ports/outbound`, Zeile B31 | erledigt |
| F-12, F-16, F-17, F-18 | Architect `567abe7`, Code F-18 und Kommentar F-16 in `166332f` | erledigt; F-18 Fänger V1 |
| F-13, F-14, F-15 | B16 (a) deterministisch; drei Raster-Zellen; `capture` ohne Antwort, `ablage` mit Gegenstand `Sonstiges` | erledigt |
| I-1, I-2, I-5, I-6 | ohne Handlung | — |
| I-3 | Messung nach dem Review in §4 | erledigt (Abschnitt 9) |
| I-4 | Frage an den Validator | ohne Träger im Repo (V6) |
| I-7 | `LH-FA-13.a`, `SPEC-005` | erledigt |
| I-8 | fünfte Drain-Phase in `dienstFaelle` | erledigt |

## 8. Übergaben (`AGENTS.md` §3.13)

| Nehmer | Sendung in §1 mit Kennung des Gebers | Ausschluss trifft | DoD nennt Gegenstand | in `done/` |
|---|---|---|---|---|
| `slice-session-wartung` | `Drain` über den Inbound Port, Bericht, Session-Regel, `CloseAll` gleichzeitig | nein | ja (beenden mit Warten, Status, `CloseAll`) | nein |
| `slice-http-adapter` | `Anfrage` unverändert, kein Ändern und Lesen bis `Close` (F-17), `Close` genau einmal, `*Fehler` mit `Code` und `Status` | nein | ja (Fehlerobjekt, Ströme) | nein |
| `slice-replay-service` | Key aus den Headern wie empfangen, gespeicherte Header und Blob, ohne Antwort nie Kandidat | nein | ja (Match-Key, `SPEC-045`) | nein |
| `slice-bootstrap-m2` | `record.New`, ein `*slog.Logger`, `CloseAll` über `SessionUseCase` | nein | ja (Verdrahtung) | nein |
| `slice-upstream-adapter` | Port implementieren, nicht ändern | nein (sein Ausschluss betrifft das Definieren des Ports) | ja | nein |

Die Sendung zu F-17 kam im selben Commit wie die Entscheidung (`567abe7`). Der Hinweis in `slice-upstream-adapter` §6 auf `slice-bootstrap-m2` §6 trägt dort der Satz „jeder Fund geht als Befund an den Slice, der ihn verantwortet“.

## 9. Größe

`git diff --numstat 7637843..HEAD` über `src/hexagon/ports/outbound/`, `src/hexagon/ports/inbound/`, `src/hexagon/services/record/` (Tests eingeschlossen): **4148** (4148 Einfügungen, 0 Löschungen; 997 Produktcode, 3151 Tests). Beim Handoff (`..0f1d866`): 3693 (968, 2725). `166332f` allein unter `src/`: 649 (49, 600). Spezifikation seit der Basis: 23 (`647f988` 8, `567abe7` 17). Die Zahlen des Implementers stimmen. Die Marke 3600 ist um 548 (15,2 %) überschritten; 4148 liegt unter der Spanne 4250 bis 4800, die §6 nach Review und Verifikation erwartet, die Fänger zu V1 bis V4 kommen hinzu. Die Entscheidung trifft der Planner (V11).

## 10. Gegenproben-Stempel und Gates

`make gegenproben` auf `166332f`: „Eingaben unverändert seit dem letzten grünen Lauf (0feef8d8…)“; der Stempel deckt `src` und änderte sich nicht, `git status` leer. `make gates` auf `166332f`: rc 0 (`GATE_CHECKS`: docs-check, a-check, baseline-verify, schema-check, migration-check, lint-profil, lint, build, test).

## Befunde

| ID | Kategorie | Befund | Quelle | Pfad | Zuständig |
|---|---|---|---|---|---|
| V1 | MEDIUM | Die Regel aus F-18 ist nur mit dem Auslöser `br` und in einer Header-Zeile belegt. Grün bleiben: `*` löst nicht aus (M01), nur `br`, `zstd`, `compress` lösen aus (M23), Entfallen je Zeile (M02). Gemutet geht aus `*, identity;q=0`, `x-gzip, identity;q=0` oder den Zeilen `br` und `identity;q=0` der Wert `identity;q=0` zum Upstream, der darauf 406 senden darf: genau die fachliche Veränderung, die Nr. 4 ausschließt. Der Kommentar von `acceptEncoding` sagt „ein Glied mit nichtleerem Coding“ zu. | `SPEC-121` Nr. 4; `LH-QA-07`; `AGENTS.md` §3.11, §3.15, §3.16 (Paar) | `src/hexagon/services/record/header_test.go` · `{ae("br, identity;q=0"), "identity", true},` | Implementer |
| V2 | MEDIUM | „Der Bericht zählt je Session“ hat keinen trennenden Fall: jeder Bericht in den Tests betrifft eine Session, deren Nachbarn nichts verloren. Gemutet (ein Zähler für den Dienst, M10) meldet `Drain(s1)` den Verlust aus `s2`, und `SessionUseCase` setzt `s1` auf `incomplete`; ein Zähler am falschen Schlüssel fiele ebenso wenig auf. Die Raster-Zelle `Drain` · `sessionID` · „ohne Exchange in diesem Dienst: Bericht 0 und 0“ (B26) läuft in einer Umgebung ohne Verlust. | `SPEC-121` Nr. 13; `LH-FA-13.a`; `AGENTS.md` §3.15 | `src/hexagon/services/record/drain.go` · `return r.sessions[sessionID].bericht, nil` | Implementer |
| V3 | MEDIUM | „Scheitert das Lesen der Zahl …, die Zahl bleibt“: der Fall „`ListBySession` scheitert“ in `TestNichtAngehaengt` hat keine vorhandenen Exchanges, die Zahl 0 gleicht dem Wert des Mutanten. Gemutet (Zahl vor der Fehlerprüfung gesetzt, M22) trägt der nächste Exchange einer Session mit drei vorhandenen die Nummer 1, `Append` weist ihn ab, ein zweiter Exchange geht verloren. | `SPEC-121` Nr. 10; `SPEC-105` Nr. 8; `AGENTS.md` §3.15 (Fixture gleich dem Mutantenwert) | `src/hexagon/services/record/anhaengen.go` · `sz.zahl, sz.zahlBekannt = len(liste), true` | Implementer |
| V4 | MEDIUM | „Anhänge … an verschiedene unabhängig“ und „Exchanges teilen nichts außer … ihrer Session“ haben keinen Fänger; die Zusatzzeile „Nebenläufigkeit: zwei Sessions zugleich“ nennt B18 und B27, die die Sessions nacheinander anhängen. Gemutet (eine Sperre für alle Sessions, M20) wartet `Close` eines Exchange von `s2` hinter dem `Append` von `s1`, bei einer belegten Ablage bis `SQLITE_BUSY_TIMEOUT`, und mit ihm das Ende der Antwort an den Client und `Drain`. Ein Recording-Fake, der nur die `Append` einer Session anhält, trennt es ohne Zeitabhängigkeit. | `SPEC-121` Nr. 10, 15; `AGENTS.md` §3.11, §3.15 | `docs/plan/planning/in-progress/slice-record-service.md` · `Nebenläufigkeit: zwei Sessions zugleich` | Implementer (Fänger), Planner (Zusatzzeile) |
| V5 | LOW | Die Drain-Seite von Nr. 2e (Sperre setzen und die laufenden sehen in einem Schritt) steht weder in §3 noch in *Ungeprüft*; benannt ist nur das Fenster „zwischen der Rückkehr des Lesens der Session und der Annahme“. M12 (zwei Sperrabschnitte in `sperre`) fängt nur `TestAnnahmeGegenDrainGleichzeitig`, in 12 von 20 und 99 von 100 Läufen; in `sperre` liegt kein Aufruf, an dem ein Test anhalten könnte. Folge wie bei F-1. | `SPEC-121` Nr. 2e, 13, *Ungeprüft*; `AGENTS.md` §3.15 | `src/hexagon/services/record/drain.go` · `sz.drain = true` | Architect (Negativ), Planner (Zeile B8) |
| V6 | LOW | I-4 des Reviews (trägt die Messmethode von `LH-QA-07` die Verengung auf `gzip`, `deflate`, `identity`?) ging an den Validator; weder §6 noch die Welle noch ein anderes Artefakt führen die Frage oder einen Ausgang. | Review `485682d` I-4; Baseline-Regelwerk `modul-08-agentenrollen.md` | `docs/plan/planning/in-progress/slice-record-service.md` · `Randformen, Negative und Wortlaut aus dem Review` | Planner |
| V7 | INFO | Haltepunkt am ersten `Done()` (Abschnitt 4): deterministisch für Lücken um `context.WithCancelCause`, blind für eine Lücke in `nimmAn`, gekoppelt an die Reihenfolge in `Record` und an die Implementierung von `propagateCancel`; mit dem Wortlaut von *Ungeprüft* vereinbar, mit seiner Folgerung nur nach Teilung des Fensters. | `SPEC-121` Nr. 2e, *Ungeprüft* | `src/hexagon/services/record/recorder.go` · `lctx, abbrechen := context.WithCancelCause(ctx)` | Architect |
| V8 | INFO | Grenze von `testing/synctest` nicht im Plan: ein Hang an `sync.Mutex` zeigt sich als Timeout von `go test`, nicht als schneller Fehler. | [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); Review I-2 | `docs/plan/planning/in-progress/slice-record-service.md` · `läuft er in einer Blase von` | Planner |
| V9 | INFO | Plan-Wortlaut neben dem Code: §3 nennt den Body des Upstream-Fakes „als Pipe“, er ist eine Folge von Schritten; §6 sagt „der Fake folgt `SPEC-120` Nr. 5“, einige Antworten tragen kleine Header-Namen (gewollt, sie prüfen die Schreibweise); B10 nennt `accept-encoding` [`br`] und `Accept-Encoding` [`gzip`] → `gzip`, der Test prüft `br, deflate` und `gzip` → `gzip, deflate` (stärker); B18 sagt „zwei Sessions zugleich“, der Test hängt nacheinander an (V4). | `AGENTS.md` §3.9, §3.11 | `docs/plan/planning/in-progress/slice-record-service.md` · `Body als Pipe` | Planner |
| V10 | INFO | `SPEC-121` Kopf: die `Anfrage` „gehört … dem Dienst“; `LH-FA-02.b`: der Adapter „besitzt“ die Ströme und gibt sie „geliehen“ weiter. Gemeint ist dieselbe Leihe; ein Leser könnte aus „gehört“ ein Schließen durch den Dienst ableiten, das Nr. 7 nicht zusagt. | `SPEC-121` Kopf; `LH-FA-02.b` | `spec/spezifikation.md` · `gehört vom Aufruf von` | Architect |
| V11 | INFO | 4148 über die Pfade aus §3 gegen die Marke 3600 (Abschnitt 9). | `AGENTS.md` §3.14 | `docs/plan/planning/in-progress/slice-record-service.md` · `über **3600** Zeilen` | Planner |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Ports `inbound`, `outbound` (Typen, Fehler, Abbildung) | gleich dem Kopf von `SPEC-120` und `SPEC-121`, ohne Befund |
| Wahl von Service, Kodierung, Session, Annahme (`wahl.go`) | Code nach Nr. 2; M25 rot; Annahme-Seite von Nr. 2e siehe Abschnitt 4 |
| Header (`header.go`) | Code nach Nr. 4, 7 einschließlich der Regel aus F-18; Fänger V1 |
| Aufnahme, Ende, ohne Antwort (`ende.go`) | Code nach Nr. 5 bis 9, 11; kein Weg gefunden, auf dem ein Rohbyte oder ein unredigierter Header die Ablage erreicht; M05, M06 rot |
| Anhängen, `Drain` (`anhaengen.go`, `drain.go`) | Code nach Nr. 10, 13; M08, M09, M11, M19 rot; Fänger V2 bis V5 |
| Protokoll (`protokoll.go`) | Code nach Nr. 12; M14, M15, M18 rot |
| Spezifikation und Sicht | kein Widerspruch; Lastenheft und `AGENTS.md` unverändert |
| Übergaben | fünf Nehmer, je mit Kennung des Gebers, kein Ausschluss trifft, keiner in `done/` |
| Zeitabhängigkeit | kein Rot ohne Codefehler in fünf Läufen bis `--cpus=0.05` |
| Gegenproben, Gates | Stempel aktuell; `make gates` rc 0 |

Für die Closure §7: V1 bis V4 sind Belege zu `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, V5 zu `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen`.
