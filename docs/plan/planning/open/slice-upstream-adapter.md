# Slice slice-upstream-adapter: Upstream-Adapter `httpupstream`

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-01`, `LH-RB-02`, `LH-QA-07`, [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, nur Standardbibliothek: `net/http`), [ADR-0002](../../adr/0002-architektur-gate-a-check.md) (Kante `adapters-driven → ports-outbound`), [ADR-0003](../../adr/0003-lint-profil-golangci.md) (Lint-Profil formt den Schnitt)

**Berührte Spec-Stellen:** `SPEC-120` Nr. 1 bis 9 und *Ungeprüft* (der Vertrag dieses Slice, vor dem Code geschrieben; der Kopf mit Port und Fehlern liegt bei `slice-record-service`) · `SPEC-109` Nr. 5 (Abbildung der Port-Fehler, gelesen, nicht geändert) · `SPEC-110` Nr. 6, 10 · `SPEC-012`, `SPEC-057`, `SPEC-086`, `SPEC-087` · `LH-FA-01.a` (Fehlermodi) · `LH-RB-02.a` · `ARC-004`, `ARC-008`, `ARC-013` · `architecture.md` §2.3, §4.1. Mit dem Vertrag nachgezogen: `LH-RB-02.a`, `SPEC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-21.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Driven Adapter `httpupstream` (`ARC-008`) setzt den Outbound Port `UpstreamClient` mit `net/http` der Standardbibliothek nach `SPEC-120` Nr. 1 bis 9 um: Request-Zeile und Header byte-gleich, Rahmung nach `Laenge`, Strom in beide Richtungen, Antwort unverändert, Frist bis zum Kopf, TLS mit `CAPEM` und `Pruefen`, HTTP/1.1 ohne Proxy aus der Umgebung, Fehlerarten ohne Werte im Text. Geprüft wird er gegen einen rohen Listener und TLS-Server auf Loopback, ohne den Record-Dienst; sein Lieferwert ist der Nachweis zu `LH-RB-02` (HTTP, HTTPS, Zertifikatsfehler) und der Transport, den `slice-bootstrap-m2` verdrahtet.

**Übernimmt:** `slice-record-service`, Teil A: der Vertrag `SPEC-120` Nr. 1 bis 9, die Fänger A1 bis A15 (§3) und die Zeilen des Rasters zu `httpupstream` samt Zusatzzeilen (§6), unverändert übernommen; Schnitt vor dem ersten Code-Commit, Grund in [slice-record-service](../done/slice-record-service.md) §4.

**Übernimmt (Port):** Der Outbound Port `UpstreamClient` mit `UpstreamAnfrage`, `UpstreamAntwort`, `ErrUpstreamNichtErreichbar`, `ErrUpstreamZeit` und ihrer Zeile in `SPEC-109` Nr. 5 entsteht in `slice-record-service` (`src/hexagon/ports/outbound/`, Kopf von `SPEC-120`); dieser Slice implementiert ihn und ändert ihn nicht. Weicht der Port im Code vom Kopf von `SPEC-120` ab, geht das an den Architect (§4, Rückführung *blockiert*), nicht in den Adapter (Sendung von `slice-record-service`).

**Übergaben an Folge-Slices** (der Nehmer führt die Sendung in seinem §1 mit der Kennung dieses Gebers; geprüft: kein *Ausdrücklich NICHT* des Nehmers trifft sie, seine DoD nennt Verdrahtung und Start): `slice-bootstrap-m2` (Verdrahtung von `httpupstream.New` nach `SPEC-120` Nr. 1: `Frist` aus `UPSTREAM_TIMEOUT`, `Pruefen` die wirksame Zertifikatsprüfung nach `SPEC-110` Nr. 10, `CAPEM` der Inhalt von `UPSTREAM_CA_FILE`; die Warnung beim Start bei abgeschalteter Prüfung, `LH-RB-02.a`). Der M2-Test dort ist der erste Lauf von Dienst und Adapter zusammen.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Port `UpstreamClient`, seine Fehler und deren Abbildung auf `ODR-E4001` (`SPEC-109` Nr. 5) — `slice-record-service`; der Adapter implementiert den Port, er definiert ihn nicht.
- Welche Header weitergehen (`X-Recorder-*`, Hop-by-Hop, `Expect`, `Accept-Encoding`), Wahl von Service und Session, Aufzeichnung — `slice-record-service` (`SPEC-121` Nr. 2, 4, 5); der Adapter sendet, was er bekommt (`SPEC-120` Kopf).
- Verdrahtung, Lesen von `UPSTREAM_CA_FILE` und die Warnung beim Start bei abgeschalteter Prüfung — `slice-bootstrap-m2` (Composition Root); der Adapter erhält Bytes und Schalter, keine Datei und keinen Logger.
- Trailer und HTTP/2 zum Upstream — akzeptiertes Negativ in `SPEC-120` *Ungeprüft* mit Folge (Trailer gehen am Recorder verloren, OData nutzt sie nicht; ein Upstream, der HTTP/1.1 ablehnt, ist nicht erreichbar); kein Nehmer, weil kein Fänger und keine Abnahme sie verlangt.
- Upstream-Fallback und Allowlist (`FALLBACK_*`) — Hybrid, nicht in der Welle (Welle §6).
- Span und Metriken des Upstream-Aufrufs — Folgewelle (Welle §6); der Adapter protokolliert nichts, jeder Fehler geht als Wert an den Aufrufer.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Adapter `httpupstream` nach `SPEC-120` Nr. 1 bis 6, 8, 9: Request-Zeile und Header byte-gleich, Rahmung nach `Laenge`, Streaming in beide Richtungen, Antwort unverändert (kein Entpacken, keine Weiterleitung), Frist bis zum Kopf, Fehlerarten in ihrer Rangfolge ohne Werte im Text, gleichzeitige Aufrufe (`LH-FA-01`, `LH-QA-07`; Fänger A1 bis A9, A12 bis A15).
- [ ] TLS und Transport nach `SPEC-120` Nr. 7: Kette und Hostname gegen System und `CAPEM`, `Pruefen=false` prüft nichts, ein Zertifikatsfehler ist `ErrUpstreamNichtErreichbar`, HTTP/1.1 auch über TLS, kein Proxy aus der Umgebung — der Nachweis zu `LH-RB-02` (HTTP, HTTPS, Zertifikatsfehler; Fänger A10, A11).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-120`: Nachzug nach Review und Verifikation).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `spec/spezifikation.md` | update | `SPEC-120` steht **vor dem Code** (geschrieben mit dem Plan von `slice-record-service`); im Slice nur Nachzug nach Review und Verifikation |
| `src/adapters/driven/httpupstream/` | neu | `client.go` (`Konfig`, `New`, `Senden`), `fehler.go` (Art des Fehlers und Rangfolge: Nr. 8), `frist.go` (Frist bis zum Kopf: Nr. 6) |
| Tests von `httpupstream` (`httpupstream_test`, `export_test.go`) | neu | roher Listener (`net.Listen` auf `127.0.0.1:0`, liest die Bytes des Requests bis zur Leerzeile und den Body, schreibt eine Antwort von Hand), `httptest.NewTLSServer` und `httptest.NewUnstartedServer` mit `EnableHTTP2` für TLS und Protokoll; `export_test.go` setzt den Haken der Frist und gibt den Transport frei; Fänger A1 bis A15 nach `LH-FA-01`, `LH-RB-02`, `LH-QA-07` |

Der Vertrag steht in `SPEC-120` **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten, das Raster *Operation × Argumentart* samt Zusatztabelle mit Fänger-Spalte in §6 (`AGENTS.md` §3.16). Die Fänger sind Namen geplanter Tests; der Implementer legt sie unter diesem Namen an oder nennt den Namen, den er stattdessen vergibt, im Bericht. Die Namen der Typen, Funktionen und Felder in `SPEC-120` gehören zum Vertrag (`slice-bootstrap-m2` ruft `New`, der Record-Dienst den Port); eine Abweichung ist eine Änderung der Spezifikation, kein Implementer-Entscheid.

- Importe: `httpupstream` `net/http`, `net/url`, `crypto/tls`, `crypto/x509`, Outbound Ports und Domain. `.a-check.yml` bleibt: die Kanten `adapters-driven → ports-outbound` und `adapters-driven → model` stehen schon darin.
- Lint-Profil, das den Schnitt formt: `noctx` (der Adapter baut Requests mit Kontext), `funlen` 100 Zeilen und 60 Anweisungen, `cyclop` 15 (die Prüfungen aus Nr. 2 und die Rangfolge aus Nr. 8 bestehen aus Teilfunktionen), `gochecknoglobals` (der Transport hängt am `Client`, nicht an einer Paketvariable; `SPEC-120` Nr. 9), `ireturn` (Rückgaben vom Port-Typ sind erlaubt), `testpackage` (externe Testpakete, Zugriff über `export_test.go`). Keine `//nolint` (`AGENTS.md` §3.2).
- Test-Haken im Produktcode, ehrlich benannt: der Haken der Frist (ein Feld, das den Timer erzeugt; im Betrieb nil, dann `time.AfterFunc`; gesetzt nur über `export_test.go`), sein Kommentar sagt das.
- Der Race-Detektor braucht cgo, das Dockerfile baut mit `CGO_ENABLED=0`: das Gate fährt kein `-race`. Die nebenläufigen Fänger ordnen mit Kanälen; der Test wartet auf Ereignisse (Zeitgrenze 10 s), nie auf eine Dauer, außer wo der Ablauf einer Frist die Bedingung ist, und dann nur in der sicheren Richtung (eine kurze Frist von `20 ms` gegen einen Listener, der den Kopf nie sendet; `1 h` gegen einen Austausch, den der Test selbst beendet). A15 fängt seine Mutation nur mit hoher Wahrscheinlichkeit, nicht mathematisch; er steht in der Tabelle als solcher. Vor der Übergabe läuft jeder nebenläufige Test mit `-count=50`.
- Testdaten von Hand: die erwartete Request-Zeile und die erwarteten Kopfzeilen stehen je Fall ausgeschrieben, nicht berechnet. Fixtures mit Steuerbytes (LF, CR, NUL, `0x7F` in A2) stehen als Go-Escapes im Quelltext; nach dem Schreiben prüft der Implementer mit `grep`, dass sie als ASCII-Text stehen und nicht als Zeichen (`BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um`).

**Zusage · Fänger** (jede Zeile löst der Reviewer mit einer Mutation auf; „Mutation“ heißt: der Fänger wird aus dem genannten Grund rot, weil die Eingaben sich in genau dem Merkmal unterscheiden, das die Zusage nennt; je Menge ein Fall je Element, je Grenze beide Seiten, je Form der Eingabe, die der Code verschieden behandelt, ein Fall; *U* Nr. n ist `SPEC-120` Nr. n, *R* Nr. n ist `SPEC-121` Nr. n). Die Zeilen A1 bis A15 sind aus `slice-record-service` unverändert übernommen:

| # | Zusage | Fänger |
|---|---|---|
| A1 | `New` prüft `Frist` und `CAPEM`; 1 ns gültig; `CAPEM` auch bei `Pruefen=false` geprüft (*U* Nr. 1) | `TestNewArgumente`: `Frist` 0 und -1 ns; `CAPEM` Text ohne PEM, `CAPEM` mit einem PEM-Block der Art `PRIVATE KEY`, je mit `Pruefen` wahr und falsch: `ErrUngueltig`, kein Client. Gutfälle: `Frist` 1 ns, `CAPEM` nil und leer, `CAPEM` mit dem Zertifikat des TLS-Servers. Mutationen: `< 1` zu `<= 1` (Fall 1 ns rot); `CAPEM` nur bei `Pruefen` geprüft (Fälle mit falsch rot); je Prüfung entfernt (ihr Fall rot) |
| A2 | `Senden` weist ungültige Argumente ab, ohne zu verbinden und ohne `Body` zu lesen; Argument vor Kontext (*U* Nr. 2) | `TestSendenArgumente`, ein Fall je Regel: `ctx` nil; `Ziel` leer, `/x`, `ftp://h`, `http://`, `http://u:p@h`, `http://h?q`, `http://h#f`; `Methode` leer, `GE T`, `GET` mit LF; `Pfad` leer, `x`, `/a?b`, `/a#b`, `/a b`, `/a` mit `0x7F`; `RawQuery` `a b`, `a#b`, `a` mit NUL; Header-Name `Bad Name` und leer; Wert mit CR, mit LF, mit NUL; `Laenge` -2; `Body` nil bei `Laenge` 5 und -1: `ErrUngueltig`, der rohe Listener sah keine Verbindung, der Body-Zähler steht bei 0. Beendeter `ctx` mit ungültigem `Ziel`: `ErrUngueltig`. Gutfälle: `Body` gesetzt bei `Laenge` 0, leerer Header-Wert, `Pfad` `/`. Mutationen: je Prüfung entfernt (ihr Fall erreicht den Listener oder endet mit `ErrUpstreamNichtErreichbar`, rot); Kontext vor den Argumenten (letzter Fall rot) |
| A3 | Request-Zeile byte-gleich: Pfad aus `Ziel`, dann `Pfad`, `?` nur bei nichtleerer Query, kein Escape geändert; Methode byte-gleich (*U* Nr. 3) | `TestRequestZeile` (roher Listener, Request-Zeile als Text): `Ziel` `http://127.0.0.1:p` und `http://127.0.0.1:p/svc/v1`; `Pfad` `/`, `/Products(1)`, `/a%2Fb`, `/a%2fb`, `/%24metadata`, `/P('a%27b')`; `RawQuery` leer, `$top=1&$filter=a%20eq%20'x'`, `a+b`, `%zz`, `a=%2F`; Methoden `GET`, `get`, `PATCH`, `MERGE`; die erwartete Zeile je Fall von Hand ausgeschrieben. Mutationen: Pfad über `URL.Path` statt roh (`%2F`, `%2f` rot); `?` bei leerer Query (rot); Methode groß geschrieben (`get` rot); Pfad aus `Ziel` vergessen (rot) |
| A4 | Genau die Header der Map: jeder Wert je Name in Reihenfolge, leere Werte gesendet, leere Wertliste nicht; `Host` aus `Ziel`; `Host`, `Content-Length`, `Transfer-Encoding`, `Trailer` der Map nicht gesendet; `User-Agent` mit dem ersten Wert, ohne ihn keiner; kein `Accept-Encoding` und kein anderer Zusatz (*U* Nr. 3) | `TestHeaderGenau` (roher Listener, Kopfzeilen als Menge, je Name als Liste): Eingabe `Accept` [`a`, `b`], `X-Leer` `[""]`, `X-Ohne` [] (leere Liste), `Prefer` [`odata.maxpagesize=1`], `Connection` [`X-Hop`], `Host` [`fremd`], `Content-Length` [`99`], `Transfer-Encoding` [`gzip`], `Trailer` [`X-T`], kein `User-Agent`: empfangen genau `Host: 127.0.0.1:p`, `Accept: a`, `Accept: b`, `X-Leer: `, `Prefer`, `Connection: X-Hop`, sonst nichts; zweiter Fall `User-Agent` [`u1`, `u2`]: genau `User-Agent: u1`. Mutationen: Kompression der Bibliothek an (`Accept-Encoding: gzip`, rot); Standard-`User-Agent` (rot); `Host` aus der Map (rot); nur der erste Wert je Name (`Accept: b` fehlt, rot) |
| A5 | Rahmung nach `Laenge`; `Body` wird bis `io.EOF` gelesen, auch nach den `Laenge` Bytes; bei `0` nicht gelesen (*U* Nr. 4) | `TestRahmung` (roher Listener): `Laenge` -1 mit `abc`: `Transfer-Encoding: chunked`, Body `abc`; `Laenge` 3 mit `abc`: `Content-Length: 3`, und der Body-Leser meldet nach den 3 Bytes einen weiteren Aufruf, der `io.EOF` lieferte; `Laenge` 0 mit gesetztem Body `x`: weder Länge noch Body am Listener, Body-Zähler 0. Mutationen: `ContentLength` nicht gesetzt (Fall 3 kommt chunked, rot); `Body` bei `Laenge` 0 durchgereicht (die Bibliothek sendet ihn chunked, rot); `io.LimitReader` um den Body (der Aufruf nach den 3 Bytes fehlt, rot) |
| A6 | Strom in beide Richtungen, nichts ganz gepuffert (*U* Nr. 4, 5) | `TestStreaming`: Request-Body aus einer Pipe: der Test schreibt `teil1`, der Listener meldet `teil1` auf einem Kanal, erst danach schreibt der Test `teil2` und schließt; Response: der Listener schreibt Kopf und `teil1` (chunked) und wartet auf einen Kanal, `Senden` kehrt zurück, `Read` liefert `teil1`, dann gibt der Test frei und `teil2` folgt. Mutationen: `io.ReadAll` des Request-Bodys vor dem Senden (der Test hängt, rot an der Zeitgrenze); Antwort ganz gelesen vor der Rückkehr (rot an der Zeitgrenze) |
| A7 | Antwort unverändert: Status, Werte je Name in Reihenfolge, Namen kanonisch, Body roh (gzip bleibt), 3xx kommt zurück, 1xx außer 101 nicht, `HEAD`, 204, 304 liefern sofort `io.EOF`, `Body` nie nil (*U* Nr. 5) | `TestAntwort`: (a) 201 mit `set-cookie: a=1`, `Set-Cookie: b=2`, `Content-Encoding: gzip` und gzip-Bytes: Status 201, `Set-Cookie` [`a=1`, `b=2`], Body byte-gleich den gzip-Bytes; (b) 302 mit `Location: /x`: Status 302, der Listener sah genau einen Request; (c) `103 Early Hints`, dann 200: Status 200; (d) 204, 304 und Antwort auf `HEAD`: `Body` nicht nil, erstes `Read` liefert 0 und `io.EOF`. Mutationen: Kompression der Bibliothek an (Body entpackt, rot); Standard-Weiterleitung ((b) rot: zwei Requests); `Body` nil bei 204 (rot) |
| A8 | Frist bis zum Kopf, nicht für den Body; treffen Kopf und Ablauf zusammen, entscheidet das Beenden der Frist (*U* Nr. 6) | `TestFrist`: (a) echter Timer, `Frist` 20 ms, der Listener sendet nie einen Kopf: `ErrUpstreamZeit`, Rückkehr vor der Zeitgrenze 10 s. Mit dem Haken der Frist (der Test löst sie von Hand aus und bestimmt, was ihr Beenden meldet): (b) Kopf da, dann löst der Test aus, dann sendet der Listener den Body: `Read` liefert den ganzen Body; (c) der Test löst aus, bevor der Listener den Kopf sendet: `ErrUpstreamZeit`; (d) Kopf da, aber das Beenden meldet „schon abgelaufen“: `ErrUpstreamZeit`, und der Body ist geschlossen (Body-Zähler am Fake). Mutationen: Frist nicht angewendet ((a) rot an der Zeitgrenze); Frist nach dem Kopf nicht beendet ((b) rot: Body bricht ab); Frist um den ganzen Aufruf samt Body ((b) rot); Ergebnis des Beendens übergangen ((d) rot: Erfolg) |
| A9 | Fehler vor dem Kopf sind `ErrUpstreamNichtErreichbar`, nicht `ErrUpstreamZeit` (*U* Nr. 8) | `TestNichtErreichbar` (`Frist` 1 h): geschlossener Port (Listener geöffnet, Adresse gemerkt, geschlossen); Host `host.invalid`; Listener schließt nach dem Request ohne Antwort; Listener antwortet `kein http` und Leerzeile: je `errors.Is(err, ErrUpstreamNichtErreichbar)` wahr und `errors.Is(err, ErrUpstreamZeit)` falsch. Mutationen: alles als `ErrUpstreamZeit` (rot); Fehler der Bibliothek unverpackt (rot) |
| A10 | TLS: Kette und Hostname gegen System und `CAPEM`; `Pruefen=false` prüft nichts; Fehler der Prüfung ist `ErrUpstreamNichtErreichbar` (*U* Nr. 7) | `TestTLS` (`httptest.NewTLSServer`, Zertifikat für `127.0.0.1`, `::1`, `example.com`): `Pruefen` ohne `CAPEM`: `ErrUpstreamNichtErreichbar`; mit dem Zertifikat als `CAPEM`: 200; mit `CAPEM` über `https://localhost:p`: `ErrUpstreamNichtErreichbar` (Hostname); `Pruefen=false` ohne `CAPEM`, über `127.0.0.1` und über `localhost`: 200. Mutationen: Prüfung immer aus (Fälle 1 und 3 rot); `CAPEM` übergangen (Fall 2 rot); Hostname nicht geprüft (Fall 3 rot) |
| A11 | HTTP/1.1 auch über TLS; kein Proxy aus der Umgebung (*U* Nr. 7) | `TestTransport`: TLS-Server mit `EnableHTTP2`: der Handler sieht `HTTP/1.1`; über `export_test.go` ist `Proxy` des Transports nil. Mutationen: HTTP/2 erzwungen (rot); `http.ProxyFromEnvironment` (Feld rot). Grenze: die Umgebung selbst setzt der Test nicht, weil die Bibliothek sie je Prozess einmal liest; der Fänger ist das Feld |
| A12 | Ist `ctx` beendet, sein Fehler, auch wenn die Frist zugleich ablief; im Body endet `Read` mit einem Fehler, der nicht `io.EOF` ist (*U* Nr. 5, 8) | `TestKontext`: Listener hält den Kopf; `cancel()`: `errors.Is(err, context.Canceled)`, kein Fehler des Ports; `context.WithTimeout` des Aufrufers 20 ms bei `Frist` 1 h: `context.DeadlineExceeded`, nicht `ErrUpstreamZeit`; Haken löst die Frist aus und danach `cancel()`, bevor der Kopf kommt: Fehler des Kontexts; nach dem Kopf `cancel()`, dann `Read`: Fehler ungleich nil und ungleich `io.EOF`. Mutationen: Kontextfehler als `ErrUpstreamNichtErreichbar` (Fall 1 rot); Deadline des Aufrufers als `ErrUpstreamZeit` (Fall 2 rot); Frist vor dem Kontext geprüft (Fall 3 rot) |
| A13 | Ein Fehler von `Body` ist weder `ErrUpstreamNichtErreichbar` noch `ErrUpstreamZeit`, `errors.Is` erreicht ihn; er geht der Frist vor (*U* Nr. 4, 8) | `TestBodyFehler`: `Body` liefert `ab`, dann `errBody`, der Listener wartet auf den ganzen Body: `errors.Is(err, errBody)`, keiner der Port-Fehler; derselbe Fall, und der Haken löst die Frist aus, nachdem `errBody` geliefert ist: weiter `errBody`. Mutationen: Body-Fehler als nicht erreichbar (rot); Frist vor dem Body-Fehler (Fall 2 rot) |
| A14 | Allaussage: kein Fehler von `Senden` nennt `Pfad`, `RawQuery` oder einen Header-Wert (*U* Nr. 8) | `TestFehlerOhneWerte`: jeder Fehlerfall aus A8 (a), A9, A10, A12 und A13 mit `Pfad` `/S3CR3T-pfad`, `RawQuery` `token=S3CR3T-q`, Header `Authorization: S3CR3T-h`: `err.Error()` enthält keinen der drei. Mutation: der Fehler der Bibliothek mit seiner URL unverändert verpackt (rot) |
| A15 | Gleichzeitige Aufrufe sind sicher (*U* Nr. 9) | `TestGleichzeitig`: 20 Aufrufe zugleich mit je eigenem Pfad und Body gegen einen Echo-Server: jede Antwort trägt ihren Pfad und Body. Ohne Race-Detektor nicht mathematisch sicher (akzeptiertes Negativ); `-count=50` |

Nicht gefangen und als akzeptiertes Negativ benannt, je mit Grund und Folge:

- Datenrennen zwischen gleichzeitigen `Senden` und dem `Read` des Bodys — das Gate fährt kein `-race` (`CGO_ENABLED=0`); die Tests ordnen mit Kanälen, A15 fängt seine Mutation nur mit hoher Wahrscheinlichkeit, `-count=50` vor jeder Übergabe; Folge im Betrieb (`SPEC-120` *Ungeprüft*): eine Anfrage mit Werten einer anderen oder ein Panic.
- Der Inhalt der Vertrauensanker des Systems, HTTP/2, Trailer, ein Upstream, der nach dem Kopf hängt, die Wiederverwendung von Verbindungen, eine Länge des Bodys, die ohne Fehler von `Laenge` abweicht — `SPEC-120` *Ungeprüft*, Grund und Folge je Fall dort; die Folge ist meist `ErrUpstreamNichtErreichbar` (eine Stelle, die weder das Image noch `CAPEM` kennt; ein Upstream, der HTTP/1.1 ablehnt; eine ruhende Verbindung, die der Upstream schloss).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-record-service` liegt in `done/` (der Port `UpstreamClient` samt Fehlern und der Zeile in `SPEC-109` Nr. 5 ist im Code, nach Review und Verifikation). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-120`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12); er ist geschrieben. Dieser Slice liegt vor `slice-bootstrap-m2`, dessen Start ihn in `done/` verlangt.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über **1900** Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; `<Basis>` ist der Commit, mit dem dieser Slice nach `in-progress/` wandert; gemessen beim Handoff des Implementers und nach dem Review, beide Male entscheidet der Planner dort). Schätzung nach dem Vorschlag aus `slice-capture-pipeline` §7 (a): 15 Zeilen der Tabelle × 1,25 × 85 (nebenläufig und strombasiert), rund 1600.
- **Der Schnitt, den eine Rückführung nimmt, vorab benannt:** die TLS-Konfiguration (`CAPEM` und `Pruefen`, `SPEC-120` Nr. 1 und 7, Fänger A10 und die `CAPEM`-Fälle von A1) wird ein eigener Slice vor `slice-bootstrap-m2`; bis dahin prüft der Adapter gegen die Vertrauensanker des Systems. Was das für `Konfig` heißt, entscheidet dann der Architect in `SPEC-120`.
- `in-progress` → `open` (blockiert): der Implementer findet eine Randform ohne Ort in `SPEC-120` oder im Raster in §6, oder eine Zelle, deren Satz die Entscheidung nicht trägt, und gibt sie an den Architect (`AGENTS.md` §3.12); ebenso, wenn der Port im Code vom Kopf von `SPEC-120` abweicht (§1 *Übernimmt (Port)*).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

Der Vertrag ist entschieden in `SPEC-120` (*U*); *R* Nr. n ist `SPEC-121` Nr. n. Jede Zelle ist gegen den Wortlaut der Stelle gelesen, nicht gegen ihre Überschrift; wo eine Zelle einen Satz zitiert, trägt er die Entscheidung. Eine Zelle, die der Implementer ohne Ort findet oder deren Satz die Entscheidung nicht trägt, geht an den Architect (`AGENTS.md` §3.12); sie wird nicht im Code entschieden. Raster und Zusatzzeilen sind aus `slice-record-service` übernommen, die Zeilen unverändert; die Zeile *Darstellung: Escapes im Pfad* trägt hier die Hälfte zu *U* Nr. 3, die Hälfte zu *R* Nr. 2b bleibt dort.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16). *Fehlend* heißt `nil` oder Nullwert, *leer* ein vorhandener leerer Wert, *unbekannt* ein Wert außerhalb der Menge, *fremd* eine Form, die der Vertrag gesondert behandelt. **(a)** Abbruch eines Arguments: gilt nicht für Werte (kein Kontext, kein Strom). **(b)** Wiederholung: gleicher Wert in zwei Aufrufen; ohne eigenen Ort gilt *U* Nr. 9 („teilen nichts außer …“).

| Operation · Argument | fehlend | leer | unbekannt | fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|
| `httpupstream.New` · `Frist` | `0`: `ErrUngueltig` (*U* Nr. 1) | gilt nicht: Zahl, `0` ist fehlend | negativ: `ErrUngueltig` (*U* Nr. 1) | 1 ns gültig, keine Obergrenze (*U* Nr. 1 „kleiner als 1 ns“) | (a) | (b) |
| `New` · `CAPEM`, `Pruefen` | nil: nur das System (*U* Nr. 1) | leer gleich nil (*U* Nr. 1 „leer ohne Datei“) | kein PEM-Zertifikat: `ErrUngueltig`, auch bei `Pruefen=false` (*U* Nr. 1) | `Pruefen=false`: keine Prüfung, `CAPEM` ohne Wirkung (*U* Nr. 7) | (a) | (b) |
| `Senden` · `ctx` | nil: `ErrUngueltig` (*U* Nr. 2) | gilt nicht: Schnittstelle | gilt nicht | schon beendet: Kontextfehler nach der Argumentprüfung (*U* Nr. 2, 8) | vor dem Kopf: Kontextfehler vor Frist und Verbindung (*U* Nr. 8); im Body: `Read`-Fehler (*U* Nr. 5) | derselbe `ctx` für zwei Aufrufe: (b) |
| `Senden` · `Ziel` | leer: `ErrUngueltig` (*U* Nr. 2) | gleich fehlend | relativ, `ftp`, ohne Host: `ErrUngueltig` (*U* Nr. 2) | Zugangsdaten, Query, Fragment: `ErrUngueltig`; Pfad im `Ziel` steht vor `Pfad` (*U* Nr. 3); Host nicht auflösbar: `ErrUpstreamNichtErreichbar` (*U* Nr. 8) | (a) | (b) |
| `Senden` · `Methode` | leer: `ErrUngueltig` (*U* Nr. 2) | gleich fehlend | kein Token: `ErrUngueltig` (*U* Nr. 2) | Kleinschreibung, `MERGE`: byte-gleich (*U* Nr. 3); `HEAD`: Body liefert sofort `io.EOF` (*U* Nr. 5) | (a) | (b) |
| `Senden` · `Pfad` | leer: `ErrUngueltig` (*U* Nr. 2 „nicht mit `/` beginnt“) | `/`: gültig (*U* Nr. 3) | ohne `/` am Anfang: `ErrUngueltig` | `?`, `#`, Byte unter `0x21`, `0x7F`: `ErrUngueltig` (*U* Nr. 2); Escapes byte-gleich (*U* Nr. 3) | (a) | (b) |
| `Senden` · `RawQuery` | leer: kein `?` (*U* Nr. 3) | gleich fehlend | gilt nicht: jeder Text ohne die verbotenen Bytes | `#`, Byte unter `0x21`, `0x7F`: `ErrUngueltig` (*U* Nr. 2); `%zz`, `+`: byte-gleich (*U* Nr. 3) | (a) | (b) |
| `Senden` · `Header` | nil: nur `Host` (*U* Nr. 3 „genau die Einträge“) | leerer Wert: gesendet; leere Wertliste: nicht gesendet (*U* Nr. 3) | Name kein Token, Wert mit CR, LF, NUL: `ErrUngueltig` (*U* Nr. 2) | `Host`, `Content-Length`, `Transfer-Encoding`, `Trailer`: nicht gesendet; `User-Agent`: erster Wert; Hop-by-Hop: gesendet (*U* Nr. 3) | (a) | mehrere Werte eines Namens: in Reihenfolge (*U* Nr. 3) |
| `Senden` · `Body`, `Laenge` | `Body` nil bei `Laenge` 0: kein Body; bei anderer: `ErrUngueltig` (*U* Nr. 2) | `Laenge` 0 mit `Body`: nicht gelesen (*U* Nr. 4) | `Laenge` unter -1: `ErrUngueltig` (*U* Nr. 2) | Länge weicht ohne Fehler ab: *Ungeprüft* mit Folge (*U*) | Fehler von `Body`: eigener Fehler vor der Frist (*U* Nr. 4, 8) | gelesen bis `io.EOF` (*U* Nr. 4) |
| `UpstreamAntwort.Body` | nie nil (*U* Nr. 5) | `HEAD`, 204, 304: sofort `io.EOF` (*U* Nr. 5) | gilt nicht | gzip bleibt, 3xx kommt zurück, 1xx außer 101 nicht (*U* Nr. 5) | `ctx` beendet: `Read`-Fehler, die Frist wirkt nicht (*U* Nr. 5, 6) | schließt der Aufrufer einmal (*R* Nr. 8c) |

**Zusatzzeilen für Kombinationen zweier Merkmale, Rangfolgen, Darstellung, Nebenläufigkeit, Modi und Allaussagen** (`AGENTS.md` §3.16); der Ort nennt die Nummer und den Satz, der die Entscheidung trägt:

| Fall | Ort (Satz) | Fänger |
|---|---|---|
| Frist × Kontext des Aufrufers × Deadline des Aufrufers | *U* Nr. 8: „ist `ctx` beendet, … auch wenn die Frist zugleich ablief“ | A12 |
| Body-Fehler × Frist | *U* Nr. 8: „dann ein Fehler von `Body`; dann `ErrUpstreamZeit`“ | A13 |
| Kopf × Ablauf der Frist zugleich | *U* Nr. 6: „entscheidet, ob das Beenden der Frist gelang“ | A8 (d) |
| TLS × `Pruefen=false` × `CAPEM` × Hostname | *U* Nr. 1, 7 | A10 |
| Darstellung: Escapes im Pfad (`%2F`, `%2f`) | *U* Nr. 3: „kein Escape wird hinzugefügt, entfernt oder umgeschrieben“ | A3 |
| Darstellung: Query `+`, `%zz`, `%2F` | *U* Nr. 3 | A3 |
| Darstellung: Namen der Antwort-Header | *U* Nr. 5: „die Namen in kanonischer Schreibweise“ | A7 (a) |
| Nebenläufigkeit: gleichzeitige `Senden` | *U* Nr. 9 | A15 (nicht mathematisch sicher) |
| Modus: `Pruefen` an und aus | *U* Nr. 7 | A10 |
| Allaussage: kein Fehlertext von `Senden` mit Werten | *U* Nr. 8 | A14 |

**Risiken:**

- Nebenläufigkeit: gleichzeitige `Senden`, Abbruch von `ctx` während `Read`, Frist gegen Kopf — **Ausgang:** offen bis zur Closure; das Gate fährt kein `-race` (`CGO_ENABLED=0`): die Fänger ordnen mit Kanälen (§3), A15 ist als nicht mathematisch sicher benannt, und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50`.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster steht oben vor dem Code, mit Zusatztabelle und Fänger-Spalte; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen bis zur Closure.
- Loopback, `localhost` und Namensauflösung in der Stufe `test` (`RUN --network=none`): roher Listener auf `127.0.0.1:0`, TLS-Server, `https://localhost:p` (A10) und `host.invalid` (A9); dieser Slice fährt als erster Tests über Loopback im Gate — **Ausgang:** offen; der Implementer erprobt je Form einen Fall in `make test`, bevor er die übrigen schreibt; scheitert eine Form an der Umgebung statt am Code, geht sie an den Architect (die Zeile der Tabelle wird geändert, nicht still übersprungen).
- Der Record-Dienst ist gegen einen Fake des Ports geprüft, der Adapter gegen `SPEC-120`; zusammen laufen sie erst im M2-Test von `slice-bootstrap-m2` — **Ausgang:** offen; A1 bis A15 halten den Adapter an `SPEC-120`, der Fake in `slice-record-service` folgt denselben Nummern; eine Abweichung, die der M2-Test findet, geht als Befund an den Slice, der sie verantwortet (`slice-bootstrap-m2` §6).
- Akzeptierte Negative ohne benannte Folge in `SPEC-120` *Ungeprüft*: der Inhalt der Vertrauensanker des Systems, HTTP/2 (ein Upstream, der nur HTTP/2 spricht), die Wiederverwendung von Verbindungen, das Ausbleiben von Datenrennen — **Ausgang:** entfallen; je Fall Grund und Folge in `SPEC-120` *Ungeprüft*, vor dem ersten Code-Commit nachgetragen (`BEO-REPO/akzeptiertes-negativ-ohne-folge`). Dort ausdrücklich: ein fehlender oder nicht lesbarer Bestand an Vertrauensankern des Systems gilt als leer; das folgt aus Nr. 1 (`New` scheitert nur an `Frist` und `CAPEM`) und hat keinen Fänger.
- **Größe:** 15 Zeilen in der Tabelle *Zusage · Fänger*, Schätzung rund 1600, Marke 1900 (§4) — **Ausgang:** offen; gemessen beim Handoff des Implementers und nach dem Review, der Schnitt einer Rückführung ist in §4 vorab benannt.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Das Register unter `../observations/BEO-REPO/` trägt acht Einträge; dieser Plan antwortet auf jeden, der ihn trifft:
- `randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16): Raster mit Zellen je Operation und Argument, Zusatztabelle mit dem tragenden Satz und einer Spalte *Fänger*, vor dem Code (§6).
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15): Tabelle *Zusage · Fänger* mit 15 Zeilen; die Eingaben trennen die Mutation (ein Fall je Regel aus Nr. 2, beide Seiten der Grenze 1 ns, je Schreibweise und Escape-Form).
- `universalaussage-im-vertrag-ungeprueft` (über der Schwelle, ohne Ausgang): die Allaussage zu den Fehlertexten hat ihre Zeile über alle Fehlerfälle (A14).
- `nebenlaeufiger-lauf-nur-im-glueck-gruen` (über der Schwelle, ohne Ausgang): der Slice ist nebenläufig; kein `-race` im Gate, Kanal-Ordnung, Fristen nur in der sicheren Richtung, `-count=50`, A15 als nicht mathematisch sicher benannt (§3).
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (10×, ohne Ausgang): dieser Slice ist der Schnitt vor dem ersten Code-Commit von `slice-record-service`; Marke, Messpunkt beim Handoff und der Schnitt einer Rückführung stehen in §4.
- `berichtigung-erreicht-nicht-jede-stelle` (1×): ändert eine Prüfrunde `SPEC-120`, trägt dieselbe Aussage auch `LH-FA-01.a` (Fehlermodi), `LH-RB-02.a` (Zertifikatsfehler), `SPEC-012` (Frist bis zum Kopf), `SPEC-109` Nr. 5 und `SPEC-121` Nr. 9, 11; die Berichtigung sucht dort mit.
- `akzeptiertes-negativ-ohne-folge` (1×): §3 nennt je akzeptiertes Negativ Grund und Folge, `SPEC-120` *Ungeprüft* je Fall ebenso (vor dem Code nachgetragen, §6).
- `werkzeug-eingabe-setzt-escape-in-zeichen-um`: die Fixtures mit Steuerbytes stehen als Go-Escapes, geprüft mit `grep` nach dem Schreiben (§3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
