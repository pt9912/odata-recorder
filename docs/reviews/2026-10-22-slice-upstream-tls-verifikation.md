# Verifikation: slice-upstream-tls

* Datum: 2026-10-22
* Rolle: verifier (frischer Kontext; kein Code, keine Spezifikation, kein Plan, kein Lastenheft, keine `AGENTS.md` geändert; `.claude/agents/architect.md` ist eine fremde, uncommittete Änderung und blieb ungelesen und unberührt)
* Gegenstand: Spitze `ccce210`, `git diff c944425..HEAD -- src/` (Basis nach Plan §4: der Commit vor dem Cherry-Pick). Die 21 Commits seit `c944425`, älteste zuerst: Cherry-Pick `d359c9e` von `0716718` → Fänger `0e2eb13`, `e2d000b`, `125e243`, `f2fbce1`, `5f23c5c` → Plan `36c57e2` → Stempel `217cb4b` → Architect `d6f4181` → Kommentar von `fehler.go` `3a4e238` → Stempel `cffa310` → Architect `6701532` → Review `229b4de` → Architect `eaa6da4` → Implementer `38ed3f5`, `6691470`, `8bcb22e` → Stempel `a7da8c4` → Architect `8d2a653` → Implementer `07632ad` → Stempel `ccce210`. Die Commits des Architect an Spezifikation und Plan sind der Vertrag, gegen den gelesen wird.
* Plan: `docs/plan/planning/in-progress/slice-upstream-tls.md`
* Maßstab: DoD (§2), Tabelle *Zusage · Fänger* (§3) und Raster (§6) gegen `SPEC-120` Nr. 1, 7, 8 und *Ungeprüft* (samt Historie), `SPEC-110` Nr. 10, 11, `SPEC-109` Nr. 3, `SPEC-121` Nr. 11, 12; Lastenheft [LH-RB-02](../../spec/lastenheft.md#lh-rb-02--http-schnittstelle), [LH-FA-01](../../spec/lastenheft.md#lh-fa-01--transparente-weiterleitung-muss); `AGENTS.md` §3.1 bis §3.16; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)
* Reihenfolge: erst die eigene Prüfung gegen Plan, DoD und Spezifikation, danach der Review-Report (`docs/reviews/2026-10-22-slice-upstream-tls-review.md`, Abschnitt 9).
* Belege selbst gefahren: `make gates` auf `ccce210` (rc 0); `make gegenproben`; `make test` mit `--no-cache-filter test` und `make lint` mit `--no-cache-filter lint` (je rc 0); 35 eigene Mutanten (Abschnitt 3), vier davon zusätzlich als Ersatz einer Datei unter `src/` mit `make test` und danach `git checkout -- <Pfad>`; Wegwerf-Proben zur Standardbibliothek im gepinnten Image (Abschnitt 4); `git diff --numstat` (Abschnitt 8); `git patch-id` (Abschnitt 8). Die Mutanten liefen auf Kopien von `git archive HEAD` (Stufe `deps` des `Dockerfile` als Image, Go 1.27.1, `--network none`, Quelle schreibgeschützt, der Go-Cache in einem benannten Volume); der Arbeitsbaum blieb unberührt (`git status --short` zeigt vor und nach allen Läufen nur `.claude/agents/architect.md`, `git diff --stat -- src` ist leer).

## Urteil

**Freigegeben: kein HIGH, kein MEDIUM, ein LOW, fünf INFO.** Der Code ist nach allen Läufen richtig. Jede Mutation der Tabelle *Zusage · Fänger*, von mir nachgebaut, ist aus dem benannten Grund rot oder gleichwertig und begründet (Abschnitt 3); F-1 und F-2 des Reviews sind im Stand geschlossen, I-1 bis I-3 soweit dieser Slice es kann (Abschnitt 9). Die Annahmen über die Bibliothek habe ich am Quelltext im gepinnten Image und mit Proben bestätigt (Abschnitt 4).

* LOW: V1 (Implementer, Planner) die Zusage „Blöcke anderer Art bleiben ohne Wirkung“ hat für die Wirkung im Pool keinen Fänger; der Block mit PEM-Headern hat das Paar, der Block anderer Art nicht.
* INFO: V2 (Implementer) der Fänger zu HTTP/1.1 über TLS läuft nur mit `Pruefen=true`; V3 (Implementer) der Kommentar von `Konfig` verweist auf den Kommentar von `New`, der `Pruefen` nicht nennt; V4 (Planner) Wortlaut der Mutation „nur den Ablauf“ in A10; V5 (Architect) „je Prozess einmal“ in *Ungeprüft* gilt für den erfolgreichen Ladevorgang; V6 (Planner) der Nullwert von `Pruefen` bleibt bis zum Fänger in `slice-bootstrap-m2` ein offener Rest zu [LH-RB-02](../../spec/lastenheft.md#lh-rb-02--http-schnittstelle).

**Sache des Planners bei der Closure (genannt, nicht bewertet):** die Messung der Größe nach dem Review (Abschnitt 8 gibt meine Zahl, die Marke ist 500); die Closure-Notiz mit Steering-Loop-Lerneintrag (§7: „Wird bei der Closure gefüllt“); das Beobachtungs-Register (§2 und §7); der Ausgang der vier Risiken in §6, die noch *offen* tragen (Cherry-Pick gegen einen geänderten Rest, Zwischenstand, Loopback und `localhost` in der Stufe `test`, Größe); die drei Paarungen (Anker · Folge-Slice · Register) bei der nächsten Welle-Closure; der Move nach `done/` (zuerst der Inhalt, dann der reine `git mv`, `AGENTS.md` §3.3).

## 1. DoD Punkt für Punkt

| # | DoD | Stand | Beleg |
|---|---|---|---|
| 1 | `CAPEM` nach `SPEC-120` Nr. 1 (Leseregel, `ErrUngueltig` auch bei `Pruefen=false`, Blöcke ohne Wirkung, jedes lesbare Zertifikat ein Anker, Block mit PEM-Headern, am Start ein Verstoß) | erfüllt; Fänger-Lücke V1 (LOW) | `leseAnker` ruft `AppendCertsFromPEM` einmal; `TestNewCAPEM` mit fünf Formen × beide Werte von `Pruefen`, sechs Gutfällen, `drei Zertifikate` (A, B, C je gegen den eigenen Server), `Block mit Headern (A) vor E`, `CAPEM` nil und leer; `TestKonfigCADatei/Block mit PEM-Headern`; Mutanten m1a bis m1d, m10, m13 bis m18 rot, m24 grün (V1) |
| 2 | TLS und Transport nach Nr. 7 und 8 (Kette und Hostname gegen Anker des Systems und `CAPEM`, ergänzt nicht ersetzt, `Pruefen=false` prüft nichts, jeder Fehler der Prüfung `ErrUpstreamNichtErreichbar` ohne Werte, HTTP/1.1 über TLS) | erfüllt | `TestTLS` (vier Fehlerfälle, acht Gutfälle), `TestTransportTLS`, `tlsFehlerFaelle` in `alleFehlerFaelle` (`TestFehlerOhneWerte`); Mutanten m2a bis m9, m11, m19, m20, m23 rot, m12 und m21 gleichwertig |
| 3 | erster Code-Commit `git cherry-pick -x 0716718`, sonst nichts; jede Änderung danach in eigenem Commit; der Bericht des Handoffs nennt beide Messungen | erfüllt | `d359c9e`: gleiche `git patch-id --stable` wie `0716718` (`544cfb44…`), Trailer `(cherry picked from commit 0716718…)`, vier Dateien, 291 Einfügungen, 5 Löschungen; kein Commit mit `src/` zwischen `c944425` und `d359c9e`; jeder spätere Commit mit `src/` berührt eine Datei oder zwei (`git log --stat c944425..HEAD -- src`). Beide Messungen (296 nach dem Cherry-Pick, 316 beim Handoff) stehen im Plan §6 *Größe*; ich habe sie nachgemessen (Abschnitt 8). Der Bericht des Handoffs selbst ist kein Artefakt im Repo |
| 4 | `make gates` grün | erfüllt | eigener Lauf auf `ccce210`, rc 0 (a-check 0 Befunde, d-check 185 Dateien 0 Befunde, `baseline-verify` OK, `schema-check`, `migration-check`, `build`, `test`, `lint-profil`, `lint`); `make test` und `make lint` zusätzlich ungecacht rc 0 (`pruefung test: ok`, „0 issues“) |
| 5 | Review durchgeführt, Report liegt vor | erfüllt | `docs/reviews/2026-10-22-slice-upstream-tls-review.md` (`229b4de`), 0 HIGH, 1 MEDIUM, 1 LOW, 4 INFO |
| 6 | Doku-Update (`SPEC-120` Nr. 1, 7 und `SPEC-110` Nr. 11: Nachzug nach Review und Verifikation) | erfüllt | Nachzug nach dem Review: `eaa6da4` (Nr. 1 Nullwert, Nr. 8 Wortlaut, *Ungeprüft* Zweig `NewCertPool`) und `8d2a653` (Nr. 1 jedes lesbare Zertifikat), je mit Zeile in der Historie; `SPEC-120` Nr. 7 und `SPEC-110` Nr. 11 blieben unverändert und stimmen (Abschnitt 5). Kein Befund dieser Verifikation verlangt eine Änderung der Spezifikation |
| 7 | Closure-Notiz mit Steering-Loop-Lerneintrag | Sache des Planners | §7 trägt „Wird bei der Closure gefüllt“ |
| 8 | Beobachtungs-Register fortgeschrieben | Sache des Planners | Abschnitt 10 nennt die Belege dieses Laufs |
| 9 | Jedes Risiko aus §6 trägt einen Ausgang | zu neun von dreizehn getragen, vier Sache des Planners | *entfallen* mit Beleg: Block mit PEM-Headern, Annahme `AppendCertsFromPEM`, Fall *f*, Text der Bibliothek, Nullwert von `Konfig`, Wortlaut Nr. 8, Zweig `NewCertPool`, Ort der Lesart, Kommentare von `Konfig` und `leseAnker`; *offen* mit Stand: Cherry-Pick, Zwischenstand, Loopback, Größe |
| 10 | Die drei Paarungen sind getragen | Sache des Planners | von der nächsten Welle-Closure geprüft |

Ein Häkchen setze ich nicht.

## 2. Tabelle *Zusage · Fänger* gegen den Vertrag

Jeder Satz von `SPEC-120` Nr. 1, 7, 8 und *Ungeprüft*, `SPEC-110` Nr. 10 und 11 gegen §3 und §6 des Plans gelesen. Jeder Fänger unter seinem Namen steht im Code: `TestNewCAPEM` (A1), `TestKonfigCADatei` (A1 am Start), `TestTLS` samt `tlsFehlerFaelle` (A10), `TestTransportTLS` (A11 TLS), `TestFehlerOhneWerte` über `alleFehlerFaelle` samt `tlsFehlerFaelle` (A14 TLS); die Fälle tragen die Namen, die der Plan nennt (`TLS unbekannte Stelle`, `TLS Hostname`, `TLS Hostname gegen Anker des Systems`, `TLS abgelaufen`, `ohne Prüfung über localhost, A als CAPEM`, `drei Zertifikate`, `ungültiger Block danach`, `Schlüssel danach`, `Block mit Headern (A) vor E`).

| Vertrag | Zusage, Randform, Zelle | Zeile · Fänger | Stand |
|---|---|---|---|
| Nr. 1 | nichtleeres `CAPEM` ohne lesbares Zertifikat: `ErrUngueltig`, kein Client, auch bei `Pruefen=false` | A1 · `TestNewCAPEM` (Text, `PRIVATE KEY`, `CERTIFICATE` ungültig, mit Header, `X509 CERTIFICATE`; je beide Werte von `Pruefen`; `c != nil` geprüft) | gedeckt: m10, m13, m14b, m15, m16 rot |
| Nr. 1 | nil und leer sind gleich, nur die Anker des Systems | A1 · `CAPEM []byte(nil)`, `CAPEM []byte{}` gegen S | gedeckt: m6, m13 rot |
| Nr. 1 | Text, Blöcke anderer Art, ungültige Blöcke, Blöcke mit PEM-Headern neben einem lesbaren Zertifikat ohne Wirkung | A1 · `zwischen Text`, `Schlüssel davor/danach`, `ungültiger Block davor/danach`, `Block mit Headern (A) vor E` | gedeckt für Text, ungültige Blöcke und Header (m17 rot); **Block anderer Art mit gültigem Inhalt im Pool nicht** (m24 grün, V1) |
| Nr. 1 | jedes lesbare Zertifikat ist ein Anker, wie viele es sind und was davor oder danach steht | A1 · `drei Zertifikate` (je gegen A, B, C), `ungültiger Block danach`, `Schlüssel danach`, die Fälle davor | gedeckt: m1a (`gegen B`, `gegen C` rot), m1b (`gegen A`, `gegen B` rot, die Fälle mit Block danach grün), m1c, m1d rot |
| Nr. 1 | `CAPEM` ergänzt die Anker des Systems und ersetzt sie nicht | A10 · `A als CAPEM gegen S` | gedeckt: m5 rot (allein dieser Fall) |
| Nr. 1 | Nullwert von `Pruefen` ist falsch und prüft nichts; der Root übergibt den Wert des Starts | A10 (`Pruefen=false`); `TestKonfigDefaults`, `TestKonfigTLSWiderspruch`; Übergabe: DoD von `slice-bootstrap-m2` | gedeckt, soweit in diesem Slice prüfbar; Rest V6 |
| Nr. 1, `SPEC-110` Nr. 11 | derselbe Leser am Start; ein Block mit PEM-Headern allein ist ein Verstoß | A1 · `TestKonfigCADatei/Block mit PEM-Headern` (Bytes der Gutfall-Datei plus ein Header) | gedeckt: m18 rot, allein dieser Fall; m26 und m27 (übrige Formen am Start) grün wie das Negativ in §3 sagt |
| Nr. 1 | ein abgelaufenes Zertifikat ist lesbar | A10 · `TLS abgelaufen` (E als `CAPEM`, `New` liefert einen Client) | gedeckt |
| Nr. 7 | Kette und Hostname gegen Anker des Systems und `CAPEM`: unbekannte Stelle, Hostname, Hostname gegen die Anker des Systems, abgelaufen | A10 · vier Fehlerfälle, `ohne CAPEM gegen S`, `A als CAPEM gegen A` | gedeckt: m3, m4, m6, m7, m8 (allein `TLS Hostname gegen Anker des Systems`), m9 (allein `TLS abgelaufen`) rot |
| Nr. 7 | `Pruefen=false` prüft nichts, `CAPEM` ohne Wirkung | A10 · fünf Fälle ohne Prüfung | gedeckt: m2e (alle fünf rot), m2a (Kette), m2b (Hostname), m2d (Ablauf), m2c, m2f (mit `CAPEM`) rot |
| Nr. 7 | jeder Fehler der Prüfung ist `ErrUpstreamNichtErreichbar` | A10 · `pruefeArt` prüft die Art und dass keine zweite Art dazukommt | gedeckt: m23 (als `ErrUpstreamZeit`) rot |
| Nr. 7 | HTTP/1.1 auch über TLS | A11 TLS · `TestTransportTLS` (Server mit `EnableHTTP2`) | gedeckt: m11 rot (`HTTP/2.0`); m12 (ohne die Zuweisung `Protocols`) grün und gleichwertig (Abschnitt 4); Lücke in der Matrix V2 |
| Nr. 8 | kein Fehler nennt einen Wert der Anfrage, ein Byte der Antwort oder den Text eines Fehlers von `Body`, auch nicht bei einem Fehler der Prüfung | A14 · `TestFehlerOhneWerte` über alle Tabellen samt `tlsFehlerFaelle`, `bodyFehlerFaelle` | gedeckt: m20 (`Pfad`, `RawQuery` bei `x509`) rot an genau den vier TLS-Fällen und an keinem anderen, m19 (Text von `Body`) rot an den drei Body-Fällen |
| Nr. 8 | der Text der Bibliothek zu Verbindung, Namensauflösung und TLS ist zugelassen | *Ungeprüft*, Plan §3 und §6 (Fall *f*) | m21 grün und gleichwertig, die Begründung ist am Weg des Textes nachgeprüft (Abschnitt 5) |
| *Ungeprüft* | Zweig `anker = x509.NewCertPool()`, Inhalt der Anker, TLS-Versionen, HTTP/2 | akzeptierte Negative mit Grund und Folge | gedeckt als Negativ; Gegenstand von V5 nur der Wortlaut „je Prozess einmal“ |
| Raster §6 | jede Zelle von `New` · `CAPEM`, `Pruefen`, Zusatzzeilen Modus und Allaussage | Abschnitt 5 | keine leere Zelle, keine Zelle ohne Ort |

## 3. Eigene Mutationen

35 Ansätze, je auf einer frischen Kopie von `git archive HEAD` (`src/`, `go.mod`, `go.sum`), je eine Ersetzung einer Datei, dann `gofmt -l`, `go vet` und `go test -count=1` über `httpupstream` und `bootstrap`. 29 rot aus dem benannten Grund, 6 grün (4 gleichwertig oder benannt, 2 Lücken: V1, V2). Rot zählt nur, wenn die Ausgabe den Fall nennt, den die Mutation treffen soll. Verworfen: ein Ansatz von m2b (ein leerer `ServerName` bei IP-Zielen färbte auch die Fälle gegen `127.0.0.1` rot; ersetzt durch die Fassung, die nur DNS-Namen prüft) und ein Übersetzungsfehler in m20 (ungenutzte Variable, behoben).

| Gruppe | Mutant | Mutation | Ergebnis |
|---|---|---|---|
| Leseregel | m1a | nur das erste lesbare Zertifikat | rot: `drei Zertifikate/A, B, C gegen B`, `gegen C` |
| | m1b | nur das letzte lesbare | rot: `gegen A`, `gegen B`; `ungültiger Block danach` und `Schlüssel danach` grün (wie der Plan sagt) |
| | m1c | nur der erste Block | rot: `Schlüssel davor`, `ungültiger Block davor`, `gegen B`, `gegen C`, `Block mit Headern (A) vor E` |
| | m1d | nur der letzte Block | rot: `ungültiger Block danach`, `Schlüssel danach`, `gegen A`, `gegen B` |
| | m10 | `CAPEM` nur bei `Pruefen=true` gelesen | rot: fünf Formen × `Pruefen=false` in `TestNewCAPEM` |
| | m13 | Prüfung auf leer entfernt (nil ergibt `ErrUngueltig`) | rot: `CAPEM []byte(nil)`, `CAPEM []byte{}`; dazu jeder Test, der einen Client ohne `CAPEM` baut (der Aufbau bricht, die benannten Fälle sind darunter) |
| | m14b | Art des Blocks nicht geprüft (Header geprüft) | rot: `X509 CERTIFICATE` allein, beide Werte von `Pruefen` |
| | m14 | Art und Header nicht geprüft | rot: `X509 CERTIFICATE` allein, `mitHeader` allein, `Block mit Headern (A) vor E` |
| | m15 | Inhalt nicht mit `ParseCertificate` geprüft | rot: `CERTIFICATE` ungültig, beide Werte |
| | m16 | Header nicht geprüft (eigener Leser) | rot: `mitHeader` allein (beide Werte), `Block mit Headern (A) vor E` |
| | m17 | Block mit Headern zählt nicht für die Prüfung, wirkt aber im Pool | rot: allein `Block mit Headern (A) vor E`; `mitHeader` allein bleibt grün (wie der Plan sagt) |
| | m18 | der Start nimmt einen Block mit Headern an | rot: `TestKonfigCADatei/Block mit PEM-Headern`, sonst nichts |
| | m24 | **Block anderer Art mit gültigem Inhalt zählt nicht für die Prüfung, wirkt aber im Pool** | **grün (V1)** |
| Modus, Kette, Hostname, Ablauf | m3 | Prüfung immer aus | rot: die vier TLS-Fehlerfälle in `TestTLS` und in `TestFehlerOhneWerte`, `Block mit Headern (A) vor E` |
| | m2e | `Pruefen=false` prüft doch alles | rot: alle fünf Fälle ohne Prüfung |
| | m2a | `Pruefen=false` prüft die Kette (nicht den Hostnamen) | rot: gegen A, über `localhost`, gegen E, `A als CAPEM` gegen E; `über localhost, A als CAPEM` grün (die Kette gilt dort) |
| | m2b | `Pruefen=false` prüft den Hostnamen (nur DNS-Namen) | rot: `über localhost`, `über localhost, A als CAPEM` |
| | m2d | `Pruefen=false` prüft den Ablauf | rot: beide Fälle gegen E (V4) |
| | m2c | `Pruefen=false` prüft alles, sobald `CAPEM` gesetzt ist | rot: `über localhost, A als CAPEM`, `A als CAPEM, gegen E` |
| | m2f | `Pruefen=false` prüft den Hostnamen, sobald `CAPEM` gesetzt ist | rot: allein `über localhost, A als CAPEM` |
| | m7 | Hostname nicht geprüft | rot: `TLS Hostname`, `TLS Hostname gegen Anker des Systems` |
| | m8 | Hostname nicht geprüft, solange `CAPEM` fehlt | rot: allein `TLS Hostname gegen Anker des Systems` |
| | m9 | Ablauf nicht geprüft | rot: allein `TLS abgelaufen` |
| | m4 | `CAPEM` geprüft, aber kein Anker | rot: `A als CAPEM gegen A`, `TestTransportTLS`, alle Gutfälle von `TestNewCAPEM` |
| | m5 | `CAPEM` ersetzt die Anker | rot: allein `A als CAPEM gegen S` |
| | m6 | Anker des Systems nicht geladen (leer ergibt einen leeren Pool) | rot: `ohne CAPEM gegen S`, `CAPEM nil`, `CAPEM {}` |
| Protokoll | m11 | HTTP/2 angeboten | rot: `TestTransportTLS` (`HTTP/2.0`) |
| | m12 | Zuweisung von `Protocols` entfernt | grün, gleichwertig (Abschnitt 4) |
| | m25 | HTTP/2 nur bei `Pruefen=false` angeboten | **grün (V2)** |
| Fehlertexte | m20 | Fehler der Prüfung mit `Pfad` und `RawQuery` im Text (bei `x509`) | rot: genau `TLS unbekannte Stelle`, `TLS Hostname`, `TLS Hostname gegen Anker des Systems`, `TLS abgelaufen` in `TestFehlerOhneWerte`, kein anderer Fall |
| | m21 | Text der Bibliothek selbst im Fehler (bei `x509`) | grün, gleichwertig (Abschnitt 5) |
| | m19 | Text des Fehlers von `Body` im Fehler | rot: `a Body liefert einen Fehler`, `b Body-Fehler vor der Frist`, `f Body, Verbindung` |
| | m23 | Fehler der Prüfung als `ErrUpstreamZeit` | rot: die vier TLS-Fälle in `TestTLS`, `Block mit Headern (A) vor E` |
| Start | m26 | Start ohne Prüfung der Art | grün: das Negativ in §3 (beide Leser rufen dieselbe Funktion) |
| | m27 | Start ohne `ParseCertificate` | grün: dasselbe Negativ |

**Vier Mutanten (m1a, m2a, m18, m19) zusätzlich als Ersatz einer Datei unter `src/` mit `make test`** (wie der Auftrag es nennt), danach `git checkout -- <Pfad>`: dieselben roten Fälle, `pruefung test` bricht mit Exit 1 ab; nach jedem Lauf `git diff --stat -- src` leer.

**Fänger mit Wahrscheinlichkeit:** keine in diesem Slice (TLS auf Loopback, Zeitgrenze 10 s, Mutationen rein datengetrieben). Zur Kontrolle: die Tests von TLS, `CAPEM` und A14 dreißigmal (`-count=30`) unter `--cpus=1`: grün (3,3 s).

## 4. Annahmen über die Bibliothek (gepinntes Image)

Gelesen im Image `golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414` (`go version`: go1.27.1 linux/amd64), `docker run --rm --network none`, ohne Mount, ohne Host-Toolchain.

| Annahme | Pfad im Image, Zeilen | Befund |
|---|---|---|
| `AppendCertsFromPEM` überspringt Blöcke anderer Art und mit Headern | `/usr/local/go/src/crypto/x509/cert_pool.go` L223 bis L252 (`if block.Type != "CERTIFICATE" \|\| len(block.Headers) != 0 { continue }`) | bestätigt |
| ein ungültiger Block wird übersprungen, jedes lesbare Zertifikat wird aufgenommen, `ok` wird beim ersten gesetzt und die Schleife läuft weiter | dort L233 bis L251 | bestätigt: kein Abbruch nach dem ersten |
| dasselbe Zertifikat mehrfach: gleichwertig mit einem Vorkommen | `cert_pool.go` `addCertFunc` L198 bis L215 (`haveSum`), `ok = true` auch beim zweiten Vorkommen | bestätigt; Probe: `ok=true` zweimal, `Pool.Equal` nach dem zweiten Aufruf wahr |
| ein `-----BEGIN` mit Text oder Einrückung davor in derselben Zeile ist Text | `/usr/local/go/src/encoding/pem/pem.go` `Decode` L89 bis L180 (`beginIndex > 0 && rest[beginIndex-1] != '\n'`) | bestätigt; Probe: `ok=false` für „Text davor, gleiche Zeile“ und „Einrückung davor“, `ok=true` bei Text in einer eigenen Zeile |
| `X509 CERTIFICATE` und ein Block mit Header zählen nicht, auch bei gültigem Inhalt | Probe im Image (Programm in einem Wegwerf-Verzeichnis, `go run`) | `ok=false` für beide; ein Header-Block vor einem gewöhnlichen Block: `ok=true` (der zweite zählt) |
| bei gesetztem `TLSClientConfig`, ohne `ForceAttemptHTTP2` und ohne `Protocols` bietet der Transport HTTP/2 nicht an | `/usr/local/go/src/net/http/transport.go` `onceSetNextProtoDefaults` L439 bis L474 und `protocols` L476 bis L501 (`case !t.ForceAttemptHTTP2 && (t.TLSClientConfig != nil …)`) | bestätigt; deshalb ist m12 gleichwertig, die Zuweisung sagt die Zusage ausdrücklich und sichert gegen eine Änderung von Code oder Bibliothek |
| die Anker des Systems werden je Prozess einmal gelesen | `/usr/local/go/src/crypto/x509/root.go` L36 bis L41 (`once.Do(initSystemRoots)`), L43 bis L68 | bestätigt für den erfolgreichen Ladevorgang; nach einem Fehler ist `systemRoots` nil, und jeder Aufruf von `SystemCertPool` liest erneut (Probe, V5) |
| `SystemCertPool` meldet einen Fehler, wenn eine vorhandene Datei in `SSL_CERT_FILE` nicht lesbar ist und die Verzeichnisse keinen Anker liefern; eine fehlende Datei ist kein Fehler | `root.go` `loadOnDiskRoots` L153 bis L203 (`firstErr` nur bei `!os.IsNotExist`; Fehler nur bei `roots.len() == 0`) | bestätigt; Probe mit `SSL_CERT_FILE` auf ein Verzeichnis und leerem `SSL_CERT_DIR`: Pool nil, Fehler `read /tmp/d: is a directory`; mit fehlender Datei: Pool nicht nil, kein Fehler |
| ein Pool nil ist ein Panic in `New`, sobald `CAPEM` ein lesbares Zertifikat trägt | Probe | `AppendCertsFromPEM` auf nil: Panic bei lesbarem Zertifikat, `ok=false` ohne |
| `SystemCertPool` liefert eine Kopie | `cert_pool.go` L121 bis L127 (`sysRoots.Clone()`) | bestätigt: `AppendCertsFromPEM` im Adapter verändert den Pool des Prozesses nicht |

## 5. Raster §6 und Vertrag

**Raster `New` · `CAPEM`, `Pruefen`:** jede der sechs Zellen hat einen Ort, dessen Wortlaut die Entscheidung trägt: fehlend (Nr. 1 „leer ohne Datei“, „ergänzt … und ersetzt sie nicht“; „Der Nullwert von `Pruefen` ist falsch und prüft kein Zertifikat“), leer („nichtleeren“), unbekannt („es braucht mindestens einen Block `CERTIFICATE` ohne PEM-Header“, am Start `SPEC-110` Nr. 11 „eine Datei mit nur einem solchen Block ist ein Verstoß“), fremd („Ohne Wirkung bleiben …“, „jedes lesbare Zertifikat darin ist ein Anker“, „ein `-----BEGIN` mit Text oder Einrückung davor … ist Text“, „Ein abgelaufenes Zertifikat ist lesbar“, Nr. 7 „`CAPEM` bleibt ohne Wirkung“, *Ungeprüft* zum Bestand des Systems), Abbruch („gilt nicht für Werte“: `New` ist synchron und trägt keinen Kontext), Wiederholung (Nr. 9 „teilen nichts außer den Verbindungen des Clients“). Keine Zelle ist im Code entschieden. Die Zelle „dasselbe Zertifikat mehrfach“ trägt ihren Grund (`haveSum`) und ihre Folge (keine) und ist am Quelltext bestätigt (Abschnitt 4).

**Zusatzzeile Modus:** Pruefen an/aus × `CAPEM` × Anker des Systems × Hostname × abgelaufen: die Zellen sind durch A10 gedeckt (m2a bis m2f, m7 bis m9), darunter die beiden Zellen, die der Review als Lücke nannte. **Zusatzzeile Allaussage:** durch A14 gedeckt (m19, m20).

**Vertrag gegen die Nachbarn:** `SPEC-109` Nr. 3 („`message` ist ein kurzer fester Text je Code, nie ein Wert aus Request, Daten oder Konfiguration … Die Ursache steht im Log, nie im Body“), `SPEC-121` Nr. 11 (der Exchange ohne Antwort trägt `ResponseStatus` 0, keinen Text) und Nr. 12 („Kein Eintrag enthält einen Fehlertext …“) tragen die Folge, die *Ungeprüft* für den Text der Bibliothek nennt (er erreicht weder Client noch Protokoll noch Aufzeichnung). `rang` in `fehler.go` verwirft den Fehler der Bibliothek ganz. `SPEC-110` Nr. 11 stimmt mit Nr. 1 überein, und `pruefeCADatei` ruft dieselbe Funktion wie `leseAnker` (`x509.NewCertPool().AppendCertsFromPEM(roh)`). Der Wert des Starts (`SPEC-110` Nr. 10): `k.UpstreamTLSVerify = !skip && (!verifyGesetzt || verify)`, falsch genau bei `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` (ein Widerspruch beider Schlüssel ist ein Verstoß), und `TestKonfigDefaults` und `TestKonfigTLSWiderspruch` halten ihn.

**Allaussagen und Kommentare (§3.7, §3.11).** Gelesen gegen die Tests: `tls.go` (jeder Satz: Fänger A1, A10, A11; „kein Test fängt diesen Zweig“ stimmt), `client.go` (`Konfig`: Nullwert, V3; `New`), `fehler.go` (Kopf: „Kein Fehler dieses Pakets nennt …“ durch A14 gehalten, m19 und m20; „dieses Paket trägt ihn nicht“ benennt selbst, dass nur der Marker im Fake gehalten ist), `tls_test.go` (`TestMain`: Umgebung vor dem ersten Gebrauch; `tlsFehlerFaelle`: die drei Formen aus Nr. 7 haben je einen Fall, der Code unterscheidet die Fehler der Bibliothek nicht; `TestNewCAPEM`: „auch eines vor einem Block ohne Wirkung“ durch `ungültiger Block danach` und `Schlüssel danach`), `fehler_test.go` (die Liste „A2, A5, A8, A9, A10 … A14 alle“ stimmt mit `alleFehlerFaelle`), `konfig_test.go` (kein neuer Kommentar). Kein Konjunktiv über eine verworfene Alternative, kein abwesender Text.

## 6. Hard Rules

Kein `//nolint` unter `src/`; Importe nur Standardbibliothek, `model` und `ports/outbound`; `make a-check` 0 Befunde; niemand außerhalb des Pakets importiert `httpupstream`; Docker-only (alle Läufe über `make` und `docker run --network none`, kein Interpreter auf dem Host); `.golangci.yml`, `.a-check.yml`, `Dockerfile`, `Makefile`, `harness/mk/` und `tools/` im Diff `c944425..HEAD` unverändert (außerhalb von `src/`, `spec/` und `docs/` ändert der Slice nur `harness/gegenproben.stempel`); `AGENTS.md` unverändert. Commit-Köpfe: jeder der 21 Commits von `c944425` (ausgenommen) bis `ccce210` nennt `LH-RB-02` und `LH-FA-01`, keiner trägt eine `SPEC-`-Kennung im Kopf. Gegenproben-Stempel: `make gegenproben` meldet „Eingaben unverändert seit dem letzten grünen Lauf (e9a8907f…)“ und lässt den Baum unverändert. Regel §3.3: der Move des Starts (`next/` nach `in-progress/`) ist ein reiner Move-Commit (`4162116`, 100 % Ähnlichkeit), die Verweise stehen in `5a4620e`.

## 7. Plan gegen Code (§3.9), Kommentare, Folge-Slices

**Dateitabelle §3:** `spec/spezifikation.md` (Nr. 1, 8, *Ungeprüft*, Historie; `SPEC-110` Nr. 11 vor dem Code, `01b02d7`), `tls.go` neu und `client.go` (`New` ruft `tlsTransport`, Kommentar von `Konfig` nennt den Nullwert), `fehler.go` nur Kommentar (Diff: sechs Einfügungen, drei Löschungen, alles Kommentar), `tls_test.go` neu, `fehler_test.go` (`tlsFehlerFaelle` in `alleFehlerFaelle`, Kommentar), `konfig_test.go` (ein Fall in `TestKonfigCADatei`), `konfig_dienste.go` unberührt: jede Zeile stimmt mit dem Diff. Die Aussage des Plans „der Cherry-Pick ändert nur Kommentar und Rumpf von `New`“ stimmt (acht Einfügungen, zwei Löschungen in `client.go`).

**Kopf.** `Bezug` (`LH-RB-02`, `LH-FA-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md)) und `Berührte Spec-Stellen` nennen den Stand nach Review (Nullwert, jedes lesbare Zertifikat, Wortlaut der Mengen, Fehler beim Laden der Anker des Systems). §1 *Übernimmt* nennt die Fälle zu mehreren Zertifikaten und zu den zwei Zellen der Modus-Matrix; §3 trägt dieselben Commits; die *Stand*-Sätze der Risiken zu `Konfig` (Nullwert) und zu `anker = x509.NewCertPool()` stimmen mit den Kommentaren im Code überein (`07632ad`).

**Folge-Slice `slice-bootstrap-m2` (§3.13).** `docs/plan/planning/open/slice-bootstrap-m2.md` liegt in `open/`; sein *Ausdrücklich NICHT* (Modus `mock` und `hybrid`, CLI-Kommandos, Lasttest, Vergleichstest) trifft die Sendung nicht; die Sendung (der Root übergibt `Pruefen` aus dem Wert des Starts) steht in §1 *Übernimmt (Upstream)* und im ersten Punkt der DoD, mit der Kennung `slice-upstream-tls`, und zwar im selben Commit (`eaa6da4`), der die Zeile „Übergaben an Folge-Slices“ im Plan dieses Slice setzt. Die Sendung nennt die Mutationen („der Root übergibt immer falsch“, „immer wahr“) und je einen der beiden Fälle, den sie färben; fahrbar ist das erst mit dem Root, denn kein Code ruft `httpupstream.New`. Der Start von `slice-bootstrap-m2` verlangt `slice-upstream-tls` in `done/` (§4 dort).

**Risiko-Ausgänge §6** (Beleg je Zeile in Abschnitt 1, Zeile 9): die neun *entfallen* tragen ihren Beleg und stimmen mit Spezifikation, Code und Tests überein (ich habe die Annahme zu `AppendCertsFromPEM`, den Zweig `NewCertPool` und die Texte der Bibliothek selbst geprüft, Abschnitte 4 und 5).

## 8. Cherry-Pick, Größe

**Cherry-Pick.** `git patch-id --stable`: `d359c9e` und `0716718` je `544cfb4414ef68a2a33b1cc0ca0581e9108685b5`; `d359c9e` trägt den Trailer `(cherry picked from commit 0716718769289e7e92a9798ce597bdb8d7f143a2)`; Dateien: `client.go`, `fehler_test.go`, `tls.go`, `tls_test.go`, 291 Einfügungen, 5 Löschungen wie `0716718`. Ein Konflikt ist nicht aufgetreten (der Plan §6 nennt einen automatischen Merge von `client.go` und `fehler_test.go`).

**Größe** (`git diff --numstat c944425..<Commit> -- src/`, Einfügungen plus Löschungen, Produktcode und Tests getrennt):

| Commit | gesamt | Produktcode | Tests |
|---|---|---|---|
| `d359c9e` (Cherry-Pick) | 296 | 48 | 248 |
| `5f23c5c` (Handoff, letzter Fänger vor dem Plan) | 316 | 48 | 268 |
| `3a4e238` (Kommentar von `fehler.go`) | 325 | 57 | 268 |
| `38ed3f5` (mehrere Zertifikate) | 348 | 57 | 291 |
| `6691470` (Modus-Matrix) | 353 | 57 | 296 |
| `07632ad` und `ccce210` (Spitze) | **356** (346 Einfügungen, 10 Löschungen) | 60 | 296 |

Die Zahlen des Plans (296, 316, 325 mit 57 und 268) stimmen; der Zuwachs „9 in `tls_test.go`, 11 in `konfig_test.go`“ stimmt. Die Marke ist 500 (§4), vom Planner vor dem ersten Code-Commit gesetzt (`db188bb`); die Messung nach dem Review gehört dem Planner.

## 9. Abgleich mit dem Review

| Review | Stand im Code und in der Spezifikation | Geschlossen |
|---|---|---|
| F-1 (MEDIUM): mehrere lesbare Zertifikate ohne Fänger, Kommentar „jedes lesbare Zertifikat“ ohne Satz | `8d2a653` setzt den Satz am Ende von Nr. 1; `38ed3f5` trägt `drei Zertifikate` (A, B, C je gegen den eigenen Server), `ungültiger Block danach`, `Schlüssel danach`; m1a, m1b, m1c, m1d von mir nachgebaut: rot an den Fällen, die der Plan nennt | ja |
| F-2 (LOW): zwei Zellen der Modus-Matrix | `6691470`: `TLS Hostname gegen Anker des Systems` (m8 rot, allein dieser Fall), `ohne Prüfung über localhost, A als CAPEM` (m2f rot, allein dieser Fall) | ja |
| I-1: Nullwert von `Konfig` | `SPEC-120` Nr. 1 nennt den Nullwert und die Pflicht des Aufrufers; der Kommentar von `Konfig` ist ein Rang-Zeiger darauf (`07632ad`); die Sendung steht im Nehmer (Abschnitt 7) | ja, soweit dieser Slice es kann; Rest V6 |
| I-2: Wortlaut von Nr. 8 | (a) „Wert der Anfrage“ und „Byte der Antwort“ sind definiert, die zugelassenen Texte je Klasse benannt; (b) der Text eines Fehlers von `Body` steht in der Liste der verbotenen Werte; (c) „`Laenge` ist keiner“. Der Inhalt ist unverändert, A14 hält ihn (m19, m20) | ja |
| I-3: Zweig `NewCertPool` | *Ungeprüft* trägt Bedingung, Messung auf dem gepinnten Image (bestätigt, Abschnitt 4) und Folge; der Kommentar von `leseAnker` zeigt dorthin und nennt die Bedingung nicht mehr; Plan §3 und §6 | ja; Wortlaut V5 |
| I-4: Äquivalente Mutationen | `p2` = m12, `e5` = m21 von mir nachgebaut: grün; `k2`, `k3` = m26, m27: grün; die Begründungen tragen | — |

## 10. Lernpunkte für die Closure

**Finding-Klassen dieses Laufs:** Fänger für die Wirkung im Pool nur für eine der zwei Sorten Block ohne Wirkung (V1: die Variante *zählt nicht für die Prüfung, wirkt aber im Pool* stand für den Block mit Headern im Plan, für den Block anderer Art nicht, weil die Fixture `PRIVATE KEY` mit 1, 2, 3 sie nicht trennt und der Fall `X509 CERTIFICATE` allein nur die Prüfung trifft) · Zusage ohne Bedingung, Fänger nur in einem Modus (V2) · Allaussage der Spezifikation über die Bibliothek gilt für den Erfolgsfall (V5).

**Beobachtungs-Register (`docs/plan/planning/observations/BEO-REPO/`).** Der Planner entscheidet, was zählt. V1 ist ein Beleg zu `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (Klasse „Fixture gleicht dem Wert der Mutation“ beziehungsweise „ein Fall je Element“); V2 und V5 sind Einzelfunde ohne bestehenden Eintrag. Die Zahlen im Plan §8 stimmen mit den Dateien unter `evidence/` (in der Reihenfolge der Verzeichnisse: 3, 2, 3, 11, 5, 15, 6, 17, 2 Belege).

**Messwerte für die Formel (§3.14):** Schätzung 326 beim Handoff, Marke 500, gemessen 316 beim Handoff, 356 an der Spitze (+12,7 % ab dem Handoff, +9,5 % ab 325); die Schätzung des Plans nach dem Review (366 bis 379) lag über dem Ist.

## Befunde

| ID | Kategorie | Befund | Quelle | Pfad | Zuständig |
|---|---|---|---|---|---|
| V1 | LOW | `SPEC-120` Nr. 1 sagt zu: „Ohne Wirkung bleiben Blöcke anderer Art“; das Raster nennt die Zelle *fremd* mit „Text, Blöcke anderer Art, ungültige Blöcke oder Blöcke mit PEM-Headern neben einem lesbaren Zertifikat: ohne Wirkung“. Für die Wirkung auf den Pool hat nur der Block mit Headern einen Fänger (`Block mit Headern (A) vor E`, m17 rot). Für den Block anderer Art stehen `Schlüssel davor/danach` (Inhalt 1, 2, 3: kein Zertifikat, trennt die Mutation nicht, wie der Plan selbst sagt) und `X509 CERTIFICATE` allein (trifft nur die Prüfung). Gemutet bleibt `go test` grün (m24): die Prüfung bleibt, die Zertifikate aus Blöcken anderer Art mit gültigem Inhalt kommen zusätzlich in den Pool. Folge im Betrieb, wenn ein eigener Leser das täte: ein Anker, den die Leseregel nicht zusagt, ohne Meldung (die Richtung, in der ein Fehler Vertrauen erweitert, nicht verengt). Heute ruft der Code die Bibliothek einmal, und kein natürlicher Mutant des Codes erreicht das; darum LOW, wie F-2 des Reviews. Ein Fall würde sie trennen: `X509 CERTIFICATE` mit dem Zertifikat von A vor einem gewöhnlichen Block (E), gegen A mit `Pruefen`: `ErrUpstreamNichtErreichbar`. | `AGENTS.md` §3.15, §3.16; `SPEC-120` Nr. 1 | `src/adapters/driven/httpupstream/tls_test.go` · `fremd := pem.EncodeToMemory(&pem.Block{Type: "X509 CERTIFICATE", Bytes: a.Certificate().Raw})` | Implementer (der Fall), Planner (Zeile A1 in §3: die Mutation ergänzen, oder das Negativ mit Grund und Folge führen) |
| V2 | INFO | `SPEC-120` Nr. 7 sagt „Der Client spricht HTTP/1.1, auch über TLS“ ohne Bedingung; `TestTransportTLS` läuft nur mit `Pruefen=true`. Gemutet bleibt grün (m25): HTTP/2 wird nur bei `Pruefen=false` angeboten. Kein natürlicher Mutant des heutigen Codes (`Protocols` wird unbedingt gesetzt), darum INFO. | `AGENTS.md` §3.15 | `src/adapters/driven/httpupstream/tls_test.go` · `proto, err := tlsAbruf(t, s, "127.0.0.1", true, pemVon(s.Certificate()))` | Implementer (optional: derselbe Fall mit `Pruefen=false`) |
| V3 | INFO | Der Kommentar von `Konfig` sagt „welche davon New anwendet, nennt der Kommentar von New“. Nach dem Cherry-Pick wendet `New` alle drei Werte an; sein Kommentar nennt `Frist` und `CAPEM` beim Namen und `Pruefen` nur über „TLS und Protokoll setzt tlsTransport“. Der Satz ist nicht falsch, aber indirekt; kein Handlungsbedarf, wenn der Architect ihn so lässt. | `AGENTS.md` §3.7 | `src/adapters/driven/httpupstream/client.go` · `// Konfig trägt die Werte aus SPEC-120 Nr. 1; welche davon New anwendet, nennt der Kommentar von New.` | Implementer |
| V4 | INFO | Plan §3 A10 nennt zur Mutation *`Pruefen=false` prüft nur den Ablauf* „der Fall gegen E rot“. Gemutet (m2d) sind beide Fälle gegen E rot, `ohne Prüfung gegen E` und `ohne Prüfung, A als CAPEM, gegen E` (der Review nennt sie als „beide Fälle gegen E“). Der Satz ist ungenau, nicht falsch. | `AGENTS.md` §3.9 | `docs/plan/planning/in-progress/slice-upstream-tls.md` · `nur den Ablauf (der Fall gegen E rot)` | Planner |
| V5 | INFO | `SPEC-120` *Ungeprüft* begründet, warum kein Test den Zweig `NewCertPool` erreicht, mit „die Bibliothek liest den Bestand je Prozess einmal“. Das gilt für den erfolgreichen Ladevorgang (`root.go` L36 bis L41: `once.Do`); nach einem Fehler setzt `initSystemRoots` den Pool auf nil, und jeder Aufruf von `SystemCertPool` liest erneut (Probe: ein erster Aufruf mit `SSL_CERT_FILE` auf ein Verzeichnis liefert den Fehler, ein zweiter mit geänderter Umgebung einen Pool). Die Folgerung „ein Fänger bräuchte einen eigenen Prozess“ bleibt richtig, denn `TestMain` lädt beim ersten Gebrauch einen Anker und danach ist der Bestand zwischengespeichert; der Wortlaut sagt es aber ohne die Bedingung. | `SPEC-120` *Ungeprüft*; `AGENTS.md` §3.11 | `spec/spezifikation.md` · `denn die Bibliothek liest den Bestand je Prozess einmal und die Tests setzen ihn vor dem ersten Gebrauch auf einen Anker` | Architect |
| V6 | INFO | [LH-RB-02](../../spec/lastenheft.md#lh-rb-02--http-schnittstelle) lässt die Abschaltung der Zertifikatsprüfung nur mit ausdrücklicher Konfiguration zu. Der Nullwert von `Konfig` prüft nichts (Entscheidung des Architect zu I-1: die Option bleibt positiv, der Aufrufer setzt `Pruefen` immer). Im Stand dieses Slice ruft kein Code `httpupstream.New` (kein Import außerhalb des Pakets); der Fänger der Übergabe (Mutationen *Root übergibt immer falsch*, *immer wahr*) ist ein Punkt der DoD von `slice-bootstrap-m2`. Bis dahin ist der Rest offen und benannt. | [LH-RB-02](../../spec/lastenheft.md#lh-rb-02--http-schnittstelle); `SPEC-120` Nr. 1 | `docs/plan/planning/open/slice-bootstrap-m2.md` · erster Punkt der DoD (Verdrahtung und Lebenszyklus) | Planner (bei der Closure von `slice-bootstrap-m2` den Fänger prüfen) |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tls.go` (`tlsTransport`, `leseAnker`) | Code nach Nr. 1 und 7; 25 Mutanten rot aus dem benannten Grund (m1a bis m11, m13 bis m17), m12 gleichwertig, m24 und m25 grün (V1, V2) |
| `client.go` (`New`, `Konfig`) | Code nach Nr. 1; Befund V3 (Kommentar) |
| `fehler.go` (Kopfkommentar) | nur Kommentar im Diff; jeder Satz gegen m19, m20, m21 gelesen; ohne Befund |
| `tls_test.go`, `fehler_test.go` | Fixtures trennen (`X509 CERTIFICATE`, `mitHeader`, E, B, C), `TestMain` setzt die Umgebung vor dem ersten Gebrauch, `TestTLS` läuft außerhalb einer Blase; dreißig Läufe unter `--cpus=1` grün; Befunde V1, V2 |
| `konfig_test.go` (`TestKonfigCADatei`) | Fixture trägt die Bytes der Gutfall-Datei plus einen Header; m18 rot allein an diesem Fall; ohne Befund |
| Spezifikation `SPEC-120` Nr. 1, 7, 8, *Ungeprüft*; `SPEC-110` Nr. 10, 11; Historie | gegen Code, Nachbarn (`SPEC-109` Nr. 3, `SPEC-121` Nr. 11, 12) und Quelltext der Bibliothek gelesen; die vier Zeilen der Historie vom 2026-10-22 stimmen mit dem Text; Befund V5 (Wortlaut) |
| Plan §1, §3, §4, §6 | Dateitabelle, Kopf, Stand-Sätze, Zahlen und Ausgänge stimmen mit dem Code; Befund V4 (Wortlaut) |
| Cherry-Pick `d359c9e` gegen `0716718` | gleiche `git patch-id`, ohne Befund |
| Folge-Slices, Übergaben (§3.13) | Sendung an `slice-bootstrap-m2` im selben Commit, kein Ausschluss trifft, Nehmer in `open/`; ohne Befund |
| Gegenproben, Gates, Suppression | Stempel gleicht dem Baum; `make gates` rc 0; `make test` und `make lint` ungecacht rc 0; kein `//nolint`, kein Gate gelockert |
| Größe | 356 (60, 296) gegen die Marke 500; die Messung nach dem Review gehört dem Planner |
