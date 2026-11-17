# Slice slice-hybrid-service-upstream: HybridUseCase, Teil 2 (Upstream-Fallback)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-hybrid](../welle-hybrid.md).

**Bezug:** `LH-FA-12`, `LH-FA-04`, `LH-FA-11`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` (Entscheidungsmatrix, Upstream-Fallback) · `LH-FA-04.b`, `LH-FA-04.c` (`X-Recorder-*`-Header) · `SPEC-120` (`UpstreamClient`, Fehler) · `SPEC-121` Nr. 4 und 7 (Header-Listen) · `SPEC-092` (`LINK_REWRITE`) · `SPEC-261` Nr. 4, 5 c, 10 (Policy und Pflichten des Aufrufers) · `SPEC-284` Nr. 1, 3 c bis f, 4, 5 c, 8, 9, 10, 11, 12 (Vertrag des Dienstes, Teile U und H6) · `ARC-002`, `ARC-003` · `architecture.md` §2.2, §4.6.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-08.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung**.

**Ziel:** Der Dienst `hybrid.Verteiler` aus [slice-hybrid-service](../done/slice-hybrid-service.md) (Teil 1: Mock zuerst und Replay-Fallback) bekommt die Stufe U: bei dem Ergebnis `Upstream` von `Entscheide` (nach dem Fehler des Mock oder nach `ODR-E5001` des Replay) sendet er über den Outbound Port `UpstreamClient` an den Upstream, mit der Kopie der Header nach `SPEC-284` Nr. 8 c, dem Body aus dem Puffer von Teil 1, dem `Ziel` des Dienstes (dieselbe Zeichenkette für `Entscheide` und `Senden`), der Kennzeichnung `X-Recorder-Fallback: upstream` und der Abbildung der Fehler des Upstream (`ODR-E4001`, `ODR-E9001`); ohne Freigabe der Policy gibt es keinen Aufruf des Upstream (`LH-FA-12.b`). Der Vertrag steht in `SPEC-284`; der Slice ist die zweite Hälfte des Plans, den der Planner vor dem Code geschnitten hat.

**Übernimmt:** die Stufe U von `slice-hybrid-service` (dessen §1 *Ausdrücklich NICHT* diesen Slice nennt, Schnitt vom 2026-11-08): das Feld `Konfig.Upstream` samt Prüfung in `New` und das Feld `Dienst.UpstreamURL`, den Zweig `Upstream` im Schalter von `SPEC-284` Nr. 3 c und d, `UpstreamAnfrage` mit Filter der Header, die Antwort und die Fehlerabbildung (Nr. 8, 9, 10, 11). Dazu aus den Abweichungen von `slice-hybrid-service` §6 (Geber `slice-hybrid-service`, Planner 2026-11-08, `AGENTS.md` §3.13): (a) die **Dienstwahl** nach `SPEC-284` Nr. 4 (längster `Root` an einer Segmentgrenze, `Ziel` gleich `UpstreamURL`, OData-Pfad, kein Dienst: `Ziel` leer) samt dem Fänger von Z01 aus Teil 1 („Dienstwahl bleibt“, Mutation „Liste nicht kopiert“, `New` hält keine Referenz auf `Dienste`), denn ohne `UpstreamURL` hat die Wahl in Teil 1 keine beobachtbare Wirkung; (b) die Aufrufstelle H5 (`Entscheide` nach dem Replay-Miss, Nr. 3 d) und die Übergabe des `Ziel` an allen sieben Stellen (vier Aufrufe der Probe, H1, H3, H5); (c) die Fänger, die Teil 1 als akzeptiertes Negativ trägt: das Argument `Pfad` an `Entscheide` (voller Pfad samt Root) und die Zahl der Aufrufe der Probe. Dazu die Pflichten des Aufrufers aus `SPEC-261` Nr. 10 (Sendung von `slice-fallback-policy-entscheidung`, die dort `slice-hybrid-service` nennt; der Teil 1 reicht sie hierher weiter, weil sie das Senden betreffen): (a) der Dienst gibt der Policy dieselbe Zeichenkette als `Ziel`, die er sendet, und liest sie mit `url.Parse` wie die Policy (Fänger: der Referenzdienst empfängt unter dem Host, den die Policy freigab, aus der Sendung von `slice-fallback-policy`); (b) er ruft `Entscheide` vor jedem Senden; (c) er sendet `X-HTTP-Method`, `X-HTTP-Method-Override` und `X-Method-Override` nie. Und das Risiko der Upstream-Links einer Fallback-Antwort aus `welle-hybrid` §6 (`LINK_REWRITE` ist nach `SPEC-092` nur `off`; §6).
Die Sendung „der Replay-Fallback einer mutierenden Methode ändert den Zustand des Mock nicht“ liegt bei Teil 1 (kein Upstream beteiligt).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Mock zuerst, Replay-Fallback, Pufferung des Bodys, Kennzeichnung `replay` und Prüfung von `ctx` vor R — `slice-hybrid-service` (Teil 1); dieser Slice fügt U, die Dienstwahl und das `Ziel` hinzu und ändert kein Verfahren der Pufferung: er ändert nur das Argument `Ziel` der vier Aufrufe von `Entscheide` in der Probe (aus leer wird die `UpstreamURL` des gewählten Dienstes) und der Aufrufe H1, H3 und H5; die Bedingung der Probe, der Puffer und die Leser bleiben (§6, entschieden).
- Die Policy (Hauptschalter, Allowlist, Routenfreigaben) und ihre Syntax — `slice-fallback-policy` und `slice-fallback-policy-entscheidung`; der Dienst ruft sie und entscheidet nichts selbst.
- Der Modus `hybrid` im Composition Root, Konfiguration, Handbuch-Kapitel und der Verbindungszähler am gebauten Binary (Z35) — `slice-bootstrap-m4`; der Dienst läuft hier gegen Fakes und einen Prüf-Client.
- Speichern einer Fallback-Antwort und Datenabgleich — Out-of-Scope von `LH-FA-12`.
- Umschreiben absoluter Upstream-Links (`LINK_REWRITE`) — `SPEC-092` kennt nur `off`.
- Spans, Audit und Metriken des Fallbacks — `welle-betrieb`.

**Prüfung der Sendung von `slice-hybrid-service`** (Planner, 2026-11-08, §1 *Ausdrücklich NICHT* und DoD dieses Slice gelesen): Dienstwahl mit Fänger von Z01, H5, Übergabe des `Ziel` und die Fänger von `Pfad` und Probe angenommen; kein Punkt von *Ausdrücklich NICHT* trifft sie (der Eintrag „Dienstwahl ohne `UpstreamURL`“ ist gestrichen), DoD Punkt 1 und die Tabelle in §3 (Z01 (D), Z12 (U), Z19) tragen sie mit der Kennung des Gebers.

**Sendungen** (`AGENTS.md` §3.13): an `slice-bootstrap-m4` dieselben beiden Punkte, die `slice-hybrid-service` dort trägt (Konfig mit `Upstream` immer gesetzt und `Dienst.UpstreamURL`; Handbuch-Kapitel nennt Links, `X-Recorder-Fallback`), der Nehmer führt sie in §1 und DoD mit den Kennungen beider Teile.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] Upstream-Fallback über `UpstreamClient` nur nach Freigabe der Policy: alle `X-Recorder-*`-Header entfernt, Body aus dem Puffer von Teil 1 (kein Fallback über `MAX_CAPTURE_BODY`), ohne Freigabe kein Aufruf des Upstream (Zähler 0); `Entscheide` läuft vor jedem Senden und bekommt als `Ziel` dieselbe Zeichenkette, die der Dienst sendet; die Dienstwahl nach `SPEC-284` Nr. 4 (längster `Root`, `Ziel` an allen sieben Stellen, Fänger Z01 (D), Z12 (U), Z19; Sendung von `slice-hybrid-service`); die drei Header zum Überschreiben der Methode gehen nie an den Upstream, Fänger: Schleife über die drei Namen × Schreibweisen am Referenzdienst, der die empfangenen Header hält (`SPEC-261` Nr. 10 a bis c, Sendung von `slice-fallback-policy-entscheidung`; Host-Fänger aus der Sendung von `slice-fallback-policy`).
- [ ] Antwort und Ströme: Kennzeichnung `upstream`, Header der Antwort ohne Hop-by-Hop, Status und Strom unverändert, eine verworfene Antwort genau einmal geschlossen und eine ausgelieferte nie; Fehler des Upstream als `ODR-E4001` (504, 502) und `ODR-E9001` (500), ohne Wiederholen und ohne Rückfall (`SPEC-284` Nr. 9 bis 11).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen; sonst steht in §7, dass das Binary unverändert blieb (der Modus `hybrid` erreicht es erst mit `slice-bootstrap-m4`).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; **kein Zaehler wird gesetzt**. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area?

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/services/hybrid/` | update | Feld `Upstream` und `UpstreamURL`, Zweig `Upstream`, `UpstreamAnfrage` mit Filter, Antwort, Fehlerabbildung neben dem Code von Teil 1 |
| `src/hexagon/services/hybrid/` (Dienstwahl und `Ziel`) | update | Wahl des Dienstes nach `SPEC-284` Nr. 4 (Kopie von `Dienste` in `New`), `Ziel` als viertes Argument von `Entscheide` an den vier Aufrufen der Probe und an H1, H3, H5; keine Zeile der Pufferung wird umgeschrieben |
| Tests | neu | Fake von `UpstreamClient` (hält jedes Argument, zählt jede Operation, Zählstrom) und der Referenzdienst als `httptest`-Server mit Verbindungszähler hinter einem Prüf-Client auf `net/http`; beide sind Test-Hilfen im Paket `hybrid`, kein Adapter (`ARC-002`) |

Die Tabelle *Zusage · Fänger* ist der Teil des Architect aus `slice-hybrid-service` (Zeilen Z07, Z19 bis Z25, Z27, Z35 mit den Nummern der Vorlage dort) und je eine Upstream-Seitenzeile zu Z01, Z08, Z09, Z10, Z11, Z12, Z26 und Z28; die Seitenzeilen tragen die Nummer der Zeile von Teil 1 und das Kürzel `(U)`. Dazu Z01 (D), der Fänger der Dienstwahl, den `slice-hybrid-service` zurückgegeben hat (§6 dort, Abweichung a). Die Ströme und Fänger über Mengen laufen als Schleife mit Zählprüfung gegen die Quelle der Menge (`model.AlleCodes()`, die Namen aus `SPEC-121` Nr. 4), Orakel ist eine unabhängige Fassung der Regel aus `SPEC-261` Nr. 4. Z13 (eigener Leser für U) fängt Z24, Z17 (Map der Anfrage unverändert nach U) Z23, Z05 (`Ziel` an `Entscheide`) Z19. Die Zeilen Z01 bis Z06, Z12 bis Z18, Z29 bis Z34 und Z36 von Teil 1 laufen unverändert weiter.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z01 (U) · SPEC-284 Nr. 1 | `New` weist ein nil in `Upstream` ab (kein Dienst, `model.ErrUngueltig`) und nimmt eine leere `UpstreamURL` an | Tabellentest; beide Seiten der Grenze | Prüfung von `Upstream` gestrichen; leere `UpstreamURL` abgewiesen | andere |
| Z01 (D) · SPEC-284 Nr. 1, 4 | `New` hält keine Referenz auf `Dienste`: ändert oder leert der Aufrufer die Liste (Roots und `UpstreamURL`) nach `New`, wählt der Dienst weiter nach der Liste, die er bekam (Geber: `slice-hybrid-service`, Fänger „Dienstwahl bleibt“) | Test: Liste nach `New` in jedem Element ändern und kürzen, dann Wahl und `Ziel` am Fake U vergleichen | Liste nicht kopiert | andere |
| Z07 · SPEC-284 Nr. 3 c, 8 | Upstream nach M: `ODR-E6001` und `ODR-E6003` bei `Upstream`, freigegebenem Host, `GET` und `HEAD`, ohne Replay-Kategorie rufen U genau einmal und R nie; mutierend nur mit Route und nur für die drei Auslöser von M | Schleife Methoden × {Route vorhanden, nicht} × Auslöser am Fake U | R immer vor U; U ohne `Entscheide` | andere |
| Z08 (U) · SPEC-284 Nr. 3 d | Replay-Miss: R liefert `ODR-E5001`, dann gilt `Entscheide` mit diesem Code: `GET` und `HEAD` mit Host rufen U (1); jede mutierende Methode, auch mit Route, und jede nie freigegebene Methode lassen `ODR-E5001` unverändert zurückkehren, U 0 | Schleife über `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `MERGE`, `OPTIONS`, `TRACE`, `get` | `Entscheide` mit `ODR-E6003` statt `ODR-E5001`; mutierend zugelassen | andere |
| Z09 (U) · SPEC-284 Nr. 3 d | Jeder Replay-Fehler außer `ODR-E5001` (jeder Code von `AlleCodes()` als Fehler von R, offenste Policy, `GET`) kehrt als derselbe Wert zurück, U 0, auch der erschöpfte Key `ODR-E5002` | Schleife über alle Codes mit Zählprüfung; Fake U zählt `Senden` | `ODR-E5002` zu U; Liste der Ausnahmen um `ODR-E5003` verkürzt | andere |
| Z10 (U) · SPEC-284 Nr. 3 e | Abbruch vor U: ist `ctx` nach R (`ODR-E5001`) beendet, läuft U nicht, das Ergebnis ist `ctx.Err()` ohne Code; ebenso nach M, wenn `Entscheide` `Upstream` ergibt | zwei Fälle am Fake U | Prüfung vor U entfernt | andere |
| Z11 (U) · SPEC-284 Nr. 2, 12 | Die Aufrufstelle H6 erhält `ctx` unverändert (Wert und `Done` identisch), und ein Kontextfehler von `Senden` (auch verpackt) kehrt unverändert ohne Code zurück, nicht als `ODR-E4001` oder `ODR-E9001` | Identität und Kontextfehler am Fake U | `context.Background()` an `Senden`; Abbildung vor der Kontextprüfung | andere |
| Z12 (U) · SPEC-284 Nr. 5 a, 5 c | Probe mit `Ziel`: eine Policy mit nur einer Host-Freigabe (keine Replay-Kategorie) und nicht leerem `Ziel` macht `POST` mit Body fallback-fähig: der Body wird gepuffert, M und U erhalten je einen eigenen Leser über den vollständigen Puffer; ein leeres `Ziel` der Probe macht dieselbe Anfrage nicht fähig (kein Puffer, Body ungelesen an M) | Fall am Fake M und Fake U mit Zählreader; Probe ruft `Entscheide` viermal | Ziel der Probe leer gelassen; Probe ohne den Code `ODR-E5001` | andere |
| Z19 · SPEC-284 Nr. 4 | Dienstwahl und Ziel: der längste `Root` an einer Segmentgrenze (`/a` passt nicht auf `/ab`); `Ziel` ist die `UpstreamURL` des gewählten Dienstes, byte-gleich an `Entscheide` an allen sieben Stellen (die vier Aufrufe der Probe, H1, H3, H5; ein Policy-Wrapper am Fake hält jedes Argument, `Pfad` ist der volle Pfad samt Root) und an `Senden` (Formen `http://h`, `HTTP://H:8080`, `https://h:0443`, `http://h/`, `http://h/p`, `http://[::1]:80`, nicht lesbar `http://h ` mit Leerzeichen); kein Dienst: `Ziel` leer, U 0 | Schleife Roots × Pfade und Schleife über die Ziel-Formen × sieben Stellen mit Zählprüfung (sieben Aufrufe der Argumente, vier davon in der Probe); Fake U hält `Ziel`; Orakel: Gleichheit mit der Zeichenkette der Konfig | `Ziel` gekürzt, kleingeschrieben oder um den Pfad verlängert; erster statt längster `Root`; `Ziel` des Nachbardienstes; `Ziel` an einer der sieben Stellen leer (je Stelle eine Mutation, H5 und H3 eingeschlossen); `Pfad` gleich dem OData-Pfad; Probe mit weniger als vier Aufrufen | andere |
| Z20 · SPEC-284 Nr. 8 a | Host-Fänger: der Referenzdienst empfängt unter dem Host (Verbindung und `Host`-Header), den die Policy freigab; `GET` freigegeben ergibt 1 Verbindung mit Methode `GET`, `POST` mit Route 1 Verbindung mit Methode `POST` und Body byte-gleich | Referenzdienst mit Verbindungszähler und Zähler nach Methode, Prüf-Client auf `net/http`, Policy mit Host gleich der Adresse des Referenzdienstes | `Ziel` aus dem Request-Host; `Ziel` des anderen Dienstes; leerer Body-Leser | andere |
| Z21 · SPEC-284 Nr. 3 c, 8 | Sicherheitsachse: keine Anfrage ohne Freigabe erreicht den Upstream (Verbindungszähler 0): Methoden `POST`, `PUT`, `PATCH`, `DELETE`, `MERGE`, `OPTIONS`, `TRACE`, `get` × Codes `ODR-E6001`, `ODR-E6002`, `ODR-E6003`, `ODR-E5001` × Policy-Zustände (alle Schlüssel leer, Hauptschalter aus bei gesetzten Listen, fremder Host, Route für anderen Pfad, Route für andere Methode, Replay-Kategorie gesetzt bei Miss); ein Header `X-HTTP-Method: GET` an einer `POST` ohne Route: 0, `X-HTTP-Method: DELETE` an einer `GET`: am Referenzdienst `GET` | Dreifach-Schleife (Menge × Menge × Menge) gegen den Referenzdienst, Zählprüfung gegen die Mengen; Orakel: unabhängige Funktion „darf senden“ nach `SPEC-261` Nr. 4 | `Senden` vor `Entscheide`; `Bleiben` bei Fehler als `Upstream`; Methode aus dem Override-Header; Hauptschalter ignoriert | andere |
| Z22 · SPEC-284 Nr. 8 c | Die drei Header zum Überschreiben der Methode gehen nie an den Upstream: drei Namen × Schreibweisen (kanonisch, klein, groß, gemischt) × {`GET`, `POST` mit Route}; andere Header (`Authorization`, `Accept`, `If-Match`, `Content-Type`, `Accept-Encoding: br`) kommen mit allen Werten und in Reihenfolge an | Referenzdienst hält die empfangenen Header, 24 Fälle der Schleife | Namensliste um einen Namen verkürzt; Vergleich mit Beachtung der Schreibweise; alle Header gestrichen | andere |
| Z23 · SPEC-284 Nr. 8 c | Header-Filter: jeder Name mit Präfix `X-Recorder-` in jeder Schreibweise entfällt, `X-Recorderx` und `X-Recorder` bleiben; die neun Hop-by-Hop-Namen, die Namen aus `Connection` (mit Leerraum, leere Teile), `Expect`, `Host`, `Content-Length` entfallen; die Kopie ist eine neue Map mit neuen Wertlisten; die Map der Anfrage ist nach U unverändert | Schleife über Namen beider Seiten der Grenze; Orakel: Liste aus `SPEC-121` Nr. 4; Änderung der Kopie lässt die Anfrage unverändert | Präfix `X-Recorder` ohne Bindestrich; Namen aus `Connection` nicht ausgewertet; Wertliste geteilt; Filter löscht in der Original-Map | andere |
| Z24 · SPEC-284 Nr. 8 a, 8 b | `Senden` erhält `Methode`, den OData-Pfad (Rest nach dem Root, beim Root `/` den Pfad), `RawQuery` unverändert (`%2F`, `+`, leer), Body und `Laenge` aus dem Puffer (eigener, vollständiger Leser), ohne Body `nil` und 0 | Schleife Roots × Pfade × Queries am Fake U | voller Pfad; Query dekodiert; `Laenge` -1; Leser der Stufe M weitergereicht | andere |
| Z25 · SPEC-284 Nr. 8, 11 | Fehler von `Senden`: `ErrUpstreamZeit` ergibt `ODR-E4001` mit 504, `ErrUpstreamNichtErreichbar` `ODR-E4001` mit 502, jeder andere Fehler samt `model.ErrUngueltig` `ODR-E9001` mit 500, die Ursache erreicht `errors.Is`; `Senden` einmal, kein R nach U, der Fehler ersetzt den früheren; ein Kontextfehler bleibt ohne Code | Tabellentest fünf Fehler × {nach M, nach R}; die Fakes zählen | Status vertauscht; Wiederholen; Rückfall auf R; früherer Fehler ausgeliefert | andere |
| Z26 (U) · SPEC-284 Nr. 9 | Kennzeichnung: eine Antwort aus U trägt `X-Recorder-Fallback: upstream`; ein Header gleichen Namens der Stufe (jede Schreibweise) wird ersetzt (genau ein Wert); die Map der Stufe bleibt unverändert | Schleife Vorbelegung des Headers; Map der Stufe vorher und nachher | angehängt statt ersetzt; in der Map der Stufe gesetzt | andere |
| Z27 · SPEC-284 Nr. 9, 10 | Antwort aus U: Status (200, 3xx, 4xx, 5xx) und Strom (derselbe Wert, nie gelesen, nie geschlossen) gehen unverändert aus; Header ohne die Hop-by-Hop-Namen und die Namen aus `Connection` der Antwort, `Set-Cookie` mehrfach, `Location`, `Content-Length`, `Content-Encoding` unverändert; absolute Upstream-Links in Body und `Location` byte-gleich | Fake U mit Zählstrom (`Read` 0, `Close` 0) und Header-Menge; Orakel: Liste aus `SPEC-121` Nr. 7 | Hop-by-Hop nicht entfernt; Body gelesen; `Location` umgeschrieben; `Content-Encoding` entfernt | andere |
| Z28 (U) · SPEC-284 Nr. 10 | Liefert U Fehler und Antwort mit Body zugleich, schließt der Dienst den Body genau einmal und liefert den Fehler (auch einen Kontextfehler) unverändert abgebildet nach Z25 bzw. ohne Code; die ausgelieferte Antwort von U schließt er nie | Fehlerart × Zählstrom | `Close` fehlt; `Close` doppelt; `Close` bei Erfolg | andere |
| Z35 · SPEC-284 *Ungeprüft* 3, 5, 6 | Dass `httpupstream` den Host verbindet, den die Policy verglich; `UpstreamAntwort.Body` nil; Änderung von `Header` oder `Body` durch den Adapter während des Aufrufs: akzeptierte Negative im Dienst. Grund: der Dienst importiert keinen Adapter (Z20 nutzt einen Prüf-Client); die beiden anderen sind Pflichten des Nachbarn (`SPEC-120` Nr. 5, `SPEC-121` Kopf). Folge: weicht der Adapter von `url.Parse` ab, ginge ein freigegebenes Ziel an einen anderen Host; sonst Panic in der Goroutine des Handlers oder Datenrennen | keiner im Dienst; der Verbindungszähler am gebauten Binary in `slice-bootstrap-m4` (DoD Punkt 2) | – | andere |

**Betriebsregeln für alle Rollen dieses Slice:** dieselben wie in `slice-hybrid-service` §3 (Repo-Dateien nur mit `Edit` und `Write`; Werte mit Backslash-u-Folge aus Bytewerten und gegengelesen, `AGENTS.md` §3.18; jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container, Mutationen nacheinander; Generator-Pflicht; der Handoff ist ein Commit für Code und Plan; Abweichungen in §6 mit Ausgang *offen*; `AGENTS.md`, `.claude/` und `.harness/skills/` schreibt keine Rolle).

**Aufträge an den Architect:** keine; der Vertrag `SPEC-284` trägt U (Nr. 3 c, d, 4, 8 bis 12). Offen ist die Rückgabe in §6 (Probe und `Ziel` zwischen den Teilen).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-hybrid-service](../done/slice-hybrid-service.md) (Teil 1) liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-284`) bereits. **Warum:** Teil 2 fügt U zum Dienst, den Teil 1 liefert (*Code-Abhängigkeit*: Datei, Schalter, Puffer, Kopie der Header, Kennzeichnung). **Gate-Kette** (Nr. 7 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)): Linie B, Kettenpfade (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel) werden nicht geändert.

**Größenmarke** (Architect, 2026-11-08, Teil 2 des Schnitts; `basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit mit `make groessenmarke`): 17 Zeilen (*a* = 0, *b* = 17), *f* = 2 (Fake `UpstreamClient`, Referenzdienst mit Prüf-Client), Produktcode rund 250 (Feld `Upstream` und Prüfung in `New`, Zweig `Upstream`, `UpstreamAnfrage` mit Filter, Antwort, Fehlerabbildung): *S* = 80·17 + 200·2 + 250 = 2010, *M* = 2800, keine Schnittfrage (Abstand 213).

**Neu gerechnet** (Planner, 2026-11-08, nach der Übernahme der Dienstwahl von `slice-hybrid-service`): die Tabelle trägt 19 Zeilen (*a* = 0, *b* = 19; neu Z01 (D) und Z12 (U); Z19 wuchs, ohne dass eine Zeile entstand, weil H5, `Pfad` und die Probe Mutationen derselben Zusage sind). Der Produktcode wächst von 250 auf rund 300: Dienstwahl mit Kopie der Liste und Rest-Pfad rund 35, `Ziel` an den sieben Stellen und H5 rund 15. *S* = 80·19 + 200·2 + 300 = 2220, *M* = 100 · ⌈135·2220 / 10000⌉ = 3000. **Schnittfrage nach der Formel: nein** (135·2220 = 299700 ≤ 300000), aber der Abstand zur Schnittgrenze 2223 beträgt **3 Zeilen** und liegt weit unter der Genauigkeit der Schätzung des Produktcodes. Ehrlich beantwortet: die Formel entscheidet (`SPEC-128` Nr. 1), und sie verneint; sie trägt keine Sicherheit. Ein Schnitt vorab ist nicht möglich, ohne den Mechanismus zu teilen: die Dienstwahl liefert das `Ziel`, das U sendet und das die Probe an `Entscheide` gibt (Z19, Z20, Z12 (U) hängen an beiden), ein dritter Teil schriebe die Aufrufe von Teil 2 um; ihn herauszulösen, kostete zudem einen dritten Plan samt Review, Verifikation und Closure. **Bedingung beim Start:** der Planner rechnet vor dem ersten Code-Commit mit dem Stand von Teil 1 nach der Closure (tatsächliche Zeilen von Port, Konfig und Probe) und mit der dann geltenden Tabelle nach; ergibt *S* ≥ 2223, wird nach `AGENTS.md` §3.14 vor dem Code geschnitten und der Schnitt steht in diesem Abschnitt. Wächst die Tabelle um eine weitere Zeile (80) oder der Produktcode um 3 Zeilen, ist die Frage mit „ja“ zu beantworten.

```groessenmarke
pfad: src/hexagon/ports/inbound/
pfad: src/hexagon/services/hybrid/
produktcode: 300
fakes: 2
```

**Rückführungen — vorab benennen:**

- `in-progress` → `next` (zu groß): über der Marke, die der Planner beim Start nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). **Marke: noch nicht gesetzt**, vorläufig 3000.
- `in-progress` → `open` (blockiert): Teil 1 liefert die Probe auf Fallback-Fähigkeit nicht als vier Aufrufe von `Entscheide` mit einem Argument `Ziel`, sodass Teil 2 mehr als dieses Argument ändern müsste.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

**Raster *Operation × Argumentart*:** das Raster *Upstream-Fallback* und das *Schließen der Ströme* stehen in §6 von [slice-hybrid-service](../done/slice-hybrid-service.md) (Architect, 2026-11-08, jede Zelle mit Ort in `SPEC-284`); die Spalte U der Zusatztabelle dort (Fänger Z07, Z21) gehört zu diesem Slice. Eine Zelle ohne Ort hält den Implementer an (`AGENTS.md` §3.12).

- Eine Fallback-Antwort kann absolute Upstream-Links tragen, `LINK_REWRITE` kennt nur `off` (`SPEC-092`; Risiko aus `welle-hybrid` §6, mit `slice-hybrid-service` an diesen Slice gegangen) — **Ausgang:** weiter offen bis zur Closure; entschieden: die Antwort trägt `X-Recorder-Fallback` und bleibt unverändert (`SPEC-284` Nr. 9, Fänger Z26 (U), Z27); der Hinweis im Handbuch-Kapitel von `slice-bootstrap-m4` ist dort in die DoD eingetragen.
- **Probe auf Fallback-Fähigkeit und `Ziel` zwischen den Teilen** (Rückgabe an den Architect, Planner 2026-11-08): Teil 1 führt weder `Dienst.UpstreamURL` noch ein `Ziel`; das leere `Ziel` gibt die Policy nie frei (`SPEC-261` Nr. 5 c), das Ergebnis `Upstream` entsteht also nicht. Teil 2 fügt das Feld und die Dienstwahl hinzu und ändert **kein Verfahren der Pufferung**, sondern das Argument `Ziel` der vier Aufrufe von `Entscheide` in der Probe und der Aufrufe H1, H3, H5 (Fänger Z12 (U), Z19). — **Ausgang:** entschieden (Architect, 2026-11-08, `SPEC-284` Nr. 4, 5 a); die Blockbedingung „Teil 1 liefert die Probe so, dass Teil 2 Zeilen der Pufferung umschreiben müsste“ entfällt, die Rückführung in §4 nennt die verbleibende Bedingung.
- **Dienstwahl und Fänger von Z01** (Geber `slice-hybrid-service`, Abweichung (a) in dessen §6; Annahme durch den Planner 2026-11-08): `Dienst.UpstreamURL` fehlt in Teil 1, die Wahl hätte dort keine beobachtbare Wirkung; Teil 2 übernimmt Wahl und Fänger (Z01 (D), Z19, §1 *Übernimmt*). Dasselbe gilt für H5, `Pfad` und die Zahl der Probe-Aufrufe (Abweichungen (b), (c)). — **Ausgang:** eingetreten (Übernahme, Größe in §4 neu gerechnet, Abstand 3 zur Schnittgrenze); Rückfrage nach `AGENTS.md` §3.13 beantwortet: Teil 1 braucht keine Dienstwahl ohne Wirkung.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`: der Fake von `UpstreamClient` trägt jede Schnittstelle des Originals, zählt jeden Aufruf und hält jedes Argument fest; ein Fake, der panikt, verdeckt den benannten Test nicht — **Ausgang:** weiter offen → beide im Register.
- **FRAGEN AN DEN USER** (die drei aus `slice-fallback-policy-entscheidung`, beim User; dieser Slice trägt keine neue): FRAGE 1 (Replay-Fallback auch mutierend, bis dahin ja; betrifft Teil 1), FRAGE 2 (`ODR-E6006` als Auslöser, bis dahin nein), FRAGE 3 (Lücke im Lastenheft). Die Antwort ändert `SPEC-261` und, soweit sie U färbt (Zeilen Z07, Z21), Tabelle und Raster dieses Slice — **Ausgang:** weiter offen; Adresse: User.
- Gate-Kette (Nr. 7 der Übergangsregeln): der Slice ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen.
- Größe: der Teil des Architect ist geliefert und vom Planner neu gerechnet (§4: *S* 2220, *M* 3000, Abstand 3 zur Schnittgrenze); `basis`, `schaetzung` und `marke` setzt der Planner beim Start, vorher mit dem Stand von Teil 1 nachgerechnet — **Ausgang:** offen; Adresse: Planner.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register.

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`.

**Vorgelagert — offene Beobachtungen sichten:** wie `slice-hybrid-service` §8 (Planner, 2026-11-08): Fakes (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`), Zweige ohne Lauf (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`), Zusagen ohne Fänger (`BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, `BEO-REPO/akzeptiertes-negativ-ohne-folge`), Randformen (`BEO-REPO/randform-erst-im-review-entschieden`) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`. Der Slice schreibt keine Zähler.

Alle berührten Sub-Areas GF.
