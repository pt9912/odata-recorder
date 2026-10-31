# Slice slice-http-adapter: HTTP-Driving-Adapter für Record und Replay

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-01`, `LH-FA-04`, `LH-FA-02`, `LH-QA-05`, `LH-QA-07`, `LH-RB-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-01.a` · `LH-FA-02.b` · `LH-FA-04.b` · `LH-QA-05.a` · `LH-QA-07.a` · `SPEC-006` · `SPEC-109` Nr. 1, 2 · `SPEC-114` Nr. 6 · `SPEC-121` · `SPEC-135` · `SPEC-136` Nr. 1 bis 10, 12 · `ARC-005` · `architecture.md` §2.2, §2.5.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Driving Adapter `src/adapters/driving/httpodata` nimmt HTTP-Requests des OData-Listeners an, bildet Methode, Pfad, `raw_query`, Header und Body-Strom unverändert auf `inbound.Anfrage` ab (`SPEC-136` Nr. 4), ruft den Dienst des Betriebsmodus als Funktionswert (`Verarbeiten`, `SPEC-136` Nr. 2), streamt die Antwort (Nr. 5, 6), schließt den Strom genau einmal, vergibt die Korrelations-ID und schreibt Fehler als Fehlerobjekt (Nr. 7, 8); ein Lesefehler oder ein Kontextfehler bricht die Verbindung ab (Nr. 9). `SPEC-136` steht mit dem Vorabentwurf des Architect in der Spezifikation (Linie A, `SPEC-129` bis `SPEC-159`); `SPEC-137` und die Zählstellen von `ODR-E5010` schreibt [slice-http-lebenszyklus](../open/slice-http-lebenszyklus.md), denn der Kern liefert den Code nicht aus. Der Lebenszyklus (Anhalten, Schließen, 503, Grenzen des Servers: `SPEC-136` Nr. 3b, 11, 13 bis 15) ist vor dem Code dort geschnitten.

**Übernimmt:** Die Korrelations-ID kommt aus `IDGenerator.KorrelationsID()` (`SPEC-114` Nr. 6), nicht aus eigenem Code des Adapters; die Tests setzen `supporttest.FestIDs` (Sendung von `slice-clock-idgenerator`).

**Übernimmt (Record):** Der Adapter füllt `inbound.Anfrage` unverändert (`SPEC-121` Kopf): `Pfad` mit den Escapes des Clients ohne Query, `RawQuery` ohne `?`, alle Header außer `Host`, `Body` (nil ohne Body), `Laenge` (`-1`, `0` oder die angekündigte Länge) und `KorrelationsID`. Er schreibt `Antwort.Status` und `Antwort.Header`, kopiert `Antwort.Body` und ruft `Antwort.Body.Close` genau einmal, auch nach einem Fehler beim Schreiben an den Client und nach dessen Abbruch (`SPEC-121` Nr. 7). Vom Aufruf von `Record` bis zur Rückkehr von `Close` (ohne `Antwort`: bis zur Rückkehr von `Record`) ändert er weder `Anfrage.Header` noch eine seiner Wertlisten und liest `Body` nicht (`SPEC-121` Kopf). Einen `*inbound.Fehler` schreibt er mit dessen `Code` und `Status` als Fehlerobjekt (`SPEC-109` Nr. 1, `SPEC-121` Nr. 14); ein Kontextfehler ohne Code bekommt keine Antwort, `model.ErrUngueltig` ist `ODR-E9001` (Sendung von `slice-record-service`).

**Übernimmt (Replay):** (Sendung von `slice-replay-service`, `SPEC-135`) Die `Anfrage` wird unverändert gefüllt, auch `X-Recorder-*`; der Dienst entnimmt die Zuordnung (`SPEC-052`). Länge und Coding der Antwort setzt der Adapter nicht selbst (`SPEC-136` Nr. 5d): der Dienst liefert weder `Content-Length` noch `Content-Encoding`, und `net/http` rahmt (ohne Länge chunked, bei `HEAD` ohne Länge); eine Länge und ein Coding, die der Record-Dienst setzt, bleiben. Ein Fehler ohne `*Fehler` und ohne Code (Kontextfehler, Abbruch des Clients, `model.ErrUngueltig`) wird wie in `SPEC-121` Nr. 14 behandelt. `X-Recorder-Match-Redacted` kennt der Adapter nicht; der Dienst liefert es in den Headern, und der Adapter schreibt sie unverändert.

**Übernimmt (Race-Detektor):** Die nebenläufigen Tests des Pakets laufen im Ziel mit `go test -race`, das `slice-race-gate` als `make test-race` an `make gates` hängt (Sendung von `slice-race-gate`); dessen Vertrag führt keine Liste der Pakete, sondern fährt alle Pakete der Stufe `source` (`SPEC-122` Nr. 2), so dass `httpodata` ohne Eintrag darunter fällt und dieser Slice dafür nichts ändert.

**Herunterfahren geht weiter:** Die Sendung von `slice-session-beenden` (zwei Schritte des Listeners, 503 auf bestehenden Verbindungen, keine doppelte Frist; `SPEC-124` Nr. 14, `LH-FA-02.c`) und das Risiko „Request-Größe und Zeitgrenzen“ liegen bei [slice-http-lebenszyklus](../open/slice-http-lebenszyklus.md), das sie in seinem §1 annimmt; dieser Slice trägt sie nicht mehr.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Entscheidungen über Cursor, Match-Key, Coverage oder CRUD — der Adapter entscheidet sie nicht (`architecture.md` §2.2).
- Die Wahl des Use Case je Modus — Composition Root (`slice-bootstrap-m2`); der Adapter kennt den Modus nicht.
- Span `odata.gateway.request` und Metriken — es gibt noch keinen `Telemetry`-Port (Welle §6); der Adapter vergibt die Korrelations-ID, die der Port später trägt.
- TLS des Listeners — Folgewelle.
- Der Lebenszyklus (Anhalten, Schließen mit Frist, 503 `ODR-E5010` auf bestehenden Verbindungen, feste Grenzen des Servers) — [slice-http-lebenszyklus](../open/slice-http-lebenszyklus.md): der einzige Teil ohne geteilten Mechanismus mit dem Kern, vor dem Code geschnitten (`AGENTS.md` §3.14).
- Spans und Zähler je Fehlercode am Adapter — `slice-telemetry-einbau-gateway` (nimmt an, siehe dessen §1).
- Absolute-Form-Requests (Entscheidung des Architect in `SPEC-136` Nr. 4c) — der Kern liefert dafür keinen Pfad.

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

- [ ] Abbildung und Streaming: Request und Antwort ohne Veränderung der Bytes, geliehene kontextgebundene Ströme, `Close` genau einmal, Abbruch durch Kontext und Lesefehler; ein Gateway-Status vor dem ersten Byte kommt vom Dienst (`ODR-E4001`, `SPEC-136` Nr. 7c), danach keine Änderung des Status (`LH-FA-02.b`); der Adapter setzt weder Länge noch Coding der Antwort (`SPEC-136` Nr. 5d), `X-Recorder-*` der Anfrage gehen unverändert an den Dienst (Sendung von `slice-replay-service`).
- [ ] Fehlerobjekt mit Code, Status und `X-Correlation-Id` für jeden Code der Tabelle, `ODR-E9001` für Verträge, die der Dienst bricht (`SPEC-136` Nr. 5a, 7, 8).
- [ ] Tests über `httptest` und rohen Client gegen Fakes der Inbound Ports, ohne SQLite; die nebenläufigen laufen im Ziel mit `-race` aus `slice-race-gate` (Sendung von `slice-race-gate`) und mit `-count=50`.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
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
| `src/adapters/driving/httpodata/` | neu | `New`, `Serve`, Handler, Abbildung, Fehlerobjekt, Streaming, Abbruch (der Lebenszyklus: `slice-http-lebenszyklus`) |
| `spec/spezifikation.md` | update | Vertrag `SPEC-136` mit Randformen: Header-Behandlung, Hop-by-Hop, Abbruch, Fehlerobjekt (Vorabentwurf des Architect, vor dem Code geschrieben; Nr. 3b, 11, 13 bis 15 liefert `slice-http-lebenszyklus`, dort auch `SPEC-137`, `SPEC-109` Nr. 2 und `SPEC-230` Nr. 8); `SPEC-121` Nr. 2c und `SPEC-135` Nr. 7d berichtigt (roher `#` im Pfad) |
| Tests des Pakets | neu | `httptest` mit Fake-Ports; Fälle je Fehlercode aus `LH-QA-05.a` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten in diesem Abschnitt, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16); beide stammen aus dem Vorabentwurf des Architect und stehen vor dem ersten Code-Commit (`AGENTS.md` §3.12). Die Tabelle der Zeilen des Lebenszyklus (L1 bis L17) führt [slice-http-lebenszyklus](../open/slice-http-lebenszyklus.md).

- Der Adapter importiert nur Inbound Ports, Querschnitts-Ports und Domain (`a-check`); die Ports `RecordUseCase` und `ReplayUseCase` liefern die beiden Slices davor.
- Der Auftrag an den Architect zur Reihenfolge des Herunterfahrens (Sendung von `slice-session-beenden`) ist beim Vorabentwurf erledigt und liegt bei [slice-http-lebenszyklus](../open/slice-http-lebenszyklus.md) (`SPEC-136` Nr. 13 bis 15).
- Modus-Routing: der Adapter kennt den Modus nicht (`architecture.md` §2.5); `Konfig.Verarbeiten` ist ein Funktionswert, den der Composition Root aus `rec.Record` (Modus `record`) oder `rep.Replay` (Modus `replay`) bindet, beide mit der Signatur `func(ctx, inbound.Anfrage) (inbound.Antwort, error)`; Mock und Hybrid binden dieselbe Signatur später (Vorabentwurf des Architect, `SPEC-136` Nr. 2).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Format nach `SPEC-128` Nr. 3 a; Zeilen A1 bis A34, Vorabentwurf des Architect).

*Nr. n* verweist auf `SPEC-136`. Mengen sind Produkte ihrer Achsen; "×2" heißt: dieselbe Menge läuft mit `fake.Record` und mit `fake.Replay` als `Verarbeiten` (Modus unabhängig). Der Fake des Dienstes ist **ein** Typ, der `RecordUseCase` und `ReplayUseCase` erfüllt, jedes Argument festhält (`ctx`, jedes Feld der `Anfrage` als Kopie, die Map `Header` als Zeiger und als Abbild) und jede Operation zählt, auch `Drain` (muss 0 bleiben). `IDs` ist `supporttest.FestIDs`. Test-Server: der `Server` auf einem Listener `127.0.0.1:0` im Test, Clients per `net/http` und, wo die Bytes der Request-Zeile zählen, per rohem `net.Dial` mit selbst geschriebenen Bytes. Fixtures sind nie der Wert der Mutation.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| A1 | `New` weist nil `Verarbeiten`, nil `IDs`, nil `Listener` ab (`model.ErrUngueltig`, kein Server, Text ohne Wert), Nr. 1 | `TestNewUngueltig`: 3 Fälle (je ein Feld nil, die anderen gesetzt), je `errors.Is(err, model.ErrUngueltig)`, Server nil, Fehlertext enthält den Canary des Feldes nicht; ein Fall mit allen drei gesetzt gelingt | Prüfung auf `Listener` entfernt: Fall 3 liefert einen Server | andere |
| A2 | `Serve` blockiert und endet mit einem Fehler, wenn der Listener von außen schließt; nie `nil`, nie Hängen, Nr. 12 | `TestServeEndet`: Listener von außen schließen, `Serve` kehrt in 5 s mit Fehler `!= nil` zurück; 1 Fall mit einem geschlossen übergebenen Listener | `Serve` liefert `nil` bei jedem Fehler | andere |
| A3 | Methode byte-gleich, Nr. 4a | `TestAnfrageMethode`: Menge 10 Methoden (`GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`, `TRACE`, `PROPFIND`, `get`) × 2 Bindungen = 20; der Fake hält `Methode` | `strings.ToUpper(r.Method)`: Fall `get` rot | andere |
| A4 | `Pfad` ist das Target bis zum ersten `?`, byte-gleich, ohne Dekodieren und Bereinigen, kein Routing, Nr. 2, 4b, 4i | `TestAnfragePfad` mit rohem `net.Dial`: Menge 13 Pfad-Formen × 2 Bindungen = 26: `/`, `/a`, `/a/`, `//a`, `/a/./b`, `/a/../b`, `/a%2Fb`, `/a%2fb`, `/%41`, `/a%20b`, `/a;x=1`, `/a#b` (roher `#`, Bytes `0x23`), `/` + `0xC3 0xA4` (rohe UTF-8-Bytes); dazu Fall 14 (nur Bindung egal): `/a%zz` ergibt 400 von `net/http`, Verarbeiten-Zähler 0, `IDs`-Zähler 0 | Handler hinter `http.ServeMux` oder `r.URL.EscapedPath()` statt `RequestURI`: `//a`, `/a/../b`, `/%41` und `0xC3 0xA4` rot | andere |
| A5 | `RawQuery` ist der Rest hinter dem ersten `?`, byte-gleich; `?` am Ende und kein `?` sind gleich, Nr. 4b | `TestAnfrageQuery` (roh): Menge 12 × 2 = 24: kein `?`, `?` allein, `a=1`, `a=1&a=2`, `$filter=Name%20eq%20%27x%27`, `%zz`, `a=%`, `a?b`, `a=b;c`, `0xC3 0xA4`, `&&`, `=` | `r.URL.RawQuery` statt `RequestURI`: `%zz`, `a?b` rot; Teilung am letzten `?` | andere |
| A6 | Targets ohne führenden `/` gehen mit dem Target bis zum ersten `?` als `Pfad` an den Dienst, ungeteilt, Nr. 4c | `TestAnfrageFremdesTarget` (roh): 3 Fälle (`GET http://h/OData/x?q=1`, `OPTIONS *`, `CONNECT h:443`); der Fake sieht `Pfad` gleich `http://h/OData/x`, `*`, `h:443` und `RawQuery` `q=1`, ``, ``; der Fake antwortet `ODR-E5005` und der Client erhält 404 mit Fehlerobjekt | `Pfad` aus `r.URL.Path` (Fall 1 liefert `/OData/x`): rot | andere |
| A7 | `Header` trägt alle Header, Werte und Reihenfolge, kanonisch, ohne `Host` und `Transfer-Encoding`, Nr. 4d | `TestAnfrageHeader` (roh): 10 Fälle × 2 = 20: ein Wert; zwei Zeilen desselben Namens (Reihenfolge); `x-a` und `X-A` (zusammen, Zeilenreihenfolge); leerer Wert; die drei `X-Recorder-*`; `Expect: 100-continue`; `Connection: close` mit `Upgrade`; `Host` nicht in `Header`; chunked (kein `Transfer-Encoding`, `Laenge` -1); `Content-Length` bei POST | nur `r.Header.Get` je Name: Fall 2 rot; `X-Recorder-*` gestrichen: Fall 5 rot; `Host` eingefügt: Fall 8 rot | andere |
| A8 | `Body` nil und `Laenge` 0 ohne Body, sonst der Strom und die angekündigte Länge, Nr. 4e | `TestAnfrageBodyForm`: 5 Formen × 2 = 10: GET ohne Body; POST `Content-Length: 0`; POST `Content-Length: 1`; POST chunked 5 Bytes in 2 Chunks; POST chunked leer; der Fake hält `Body` (nil oder nicht) und `Laenge` (0, 0, 1, -1, -1) | `Body` immer gesetzt (Fall 1, 2 rot); `Laenge` immer `r.ContentLength` bei `NoBody` ohne Prüfung geändert auf `-1` | andere |
| A9 | Der Request-Body kommt ungelesen, unverändert und ohne Grenze an, Nr. 4e, 4h | `TestAnfrageBodyBytes`: 8 Größen (1, 4095, 4096, 4097, 65535, 65536, 65537, 8 388 609 Bytes) × 2 = 16; der Fake liest alles und vergleicht Byte für Byte mit dem gesendeten (Zufallsmuster, nie ein Wert, den ein Puffer füllen würde); Fall 8 liegt über `MAX_CAPTURE_BODY` (8 MiB) | Adapter liest 1 Byte vor (`Peek`): jeder Fall rot; `http.MaxBytesReader` mit 8 MiB: Fall 8 rot | andere |
| A10 | `KorrelationsID` kommt einmal je Request aus `IDs`, Nr. 3a, 4f | `TestKorrelationsID`: 3 aufeinanderfolgende Requests, `FestIDs`: der Fake sieht `cor-…01`, `…02`, `…03`; der `IDs`-Zähler `KorrelationsID` steht auf 3 und die übrigen vier Operationen auf 0 | zweiter Aufruf von `KorrelationsID` im Fehlerpfad (Zähler 4): rot; Konstante statt `IDs`: rot | andere |
| A11 | `Verarbeiten` genau einmal je Request, sonst kein Aufruf (`Drain` nie), nur der Dienst der Bindung, Nr. 2, 3d | `TestEinAufruf`: 4 Requests × 2 Bindungen; Zähler `Record` bzw. `Replay` gleich 4, der jeweils andere 0, `Drain` 0 | Adapter ruft nach dem Fehler erneut (Wiederholung): Zähler 8; ruft `Drain`: rot | andere |
| A12 | `ctx` ist `r.Context()` ohne Frist und endet, wenn der Client abbricht, Nr. 4g, 9a | `TestKontextClientAbbruch`: GET, der Fake wartet auf `ctx.Done()`; der Test schließt die Verbindung; der Fake sieht `Done` in 5 s, `ctx.Err() == context.Canceled`, `Deadline()` ohne Frist; zweiter Fall: POST mit vollständig gelesenem Body, dann dasselbe | `context.Background()` an den Dienst: der Fake sieht `Done` nie, Test rot per Frist; `context.WithTimeout` im Adapter: `Deadline()` gesetzt, rot | nebenläufig |
| A13 | Status geht unverändert aus, Nr. 5b | `TestAntwortStatus`: Menge 9 (200, 201, 204, 301, 304, 400, 404, 500, 599) × 2 = 18, der Fake liefert Status und leeren Body | Status 599 auf 500 gedeckelt oder 200 verändert: rot | andere |
| A14 | Ein Bruch des Dienstes (Body nil, Status außerhalb 200 bis 599) ergibt `ODR-E9001` mit 500 ohne Header der Antwort, `Close` einmal, wenn ein Body da ist, Nr. 5a | `TestAntwortVerstoss`: 8 Fälle: Body nil mit Status 200; Status 0, 100, 101, 199, 600, 1000 mit Body (je `Close` 1, Header des Fakes `X-Marker` nicht in der Antwort); Body nil mit Status 0 (Paar, `Close` nicht gerufen); die Antwort ist das Fehlerobjekt `ODR-E9001` mit 500 | Prüfung `Status < 100`: Fall 199 rot; Grenze `> 599` auf `> 600`: Fall 600 rot; Header trotzdem übernommen: `X-Marker` sichtbar | andere |
| A15 | Antwort-Header unverändert, alle Werte in Reihenfolge, Nr. 5c | `TestAntwortHeader`: 6 Fälle × 2 = 12: ein Wert; zwei Werte (Reihenfolge); `Set-Cookie` zweimal; leerer Wert; Name `x-lower` (Client sieht `X-Lower`); 64 Header | nur der erste Wert: Fall 2, 3 rot; leere Werte übersprungen: Fall 4 rot | andere |
| A16 | Hop-by-Hop-Header entfallen samt der Namen, die `Connection` nennt, Nr. 5c | `TestAntwortHopByHop` (roher Client, Zeilen der Antwort): Menge 9 feste Namen × 2 Schreibweisen (kanonisch, klein) = 18, je mit eindeutigem Marker-Wert, der in keiner Antwortzeile erscheinen darf; dazu Fall 19: `Connection: X-Foo, x-bar` mit `X-Foo`, `X-Bar` und `X-Keep` (nur `X-Keep` bleibt); Fall 20: `Connection:  X-A ,,<Tab>X-B` mit `X-A`, `X-B` (beide entfallen) | Liste ohne `Upgrade` (Fall `Upgrade` rot) oder ohne `Trailer`: je rot; Liste der `Connection` ignoriert: Fall 19 rot | andere |
| A17 | Kein Sniffing: ohne `Content-Type` wird keiner ergänzt, Nr. 5d | `TestAntwortContentType`: 3 Fälle (fehlt, Body `<html>x</html>`; `Content-Type: text/plain`; `content-type: application/json` klein); Client: Fall 1 ohne den Header, Fälle 2 und 3 unverändert | kein `Header()["Content-Type"] = nil`: Fall 1 liefert `text/html; charset=utf-8` | andere |
| A18 | Die Bytes der Antwort gehen unverändert und vollständig aus, über jede Puffergrenze, Nr. 6a | `TestAntwortBytes`: Menge 9 Größen (0, 1, 4095, 4096, 4097, 32767, 32768, 32769, 1 048 577) × 2 = 18; Vergleich Byte für Byte (Zufallsmuster) | Kopieren bricht bei 32768 ab oder verschluckt das letzte Stück: Größen 32769 und 1 048 577 rot | andere |
| A19 | `Read`-Formen: `(n, nil)`-Folge, `(n, EOF)` zusammen, `(0, nil)` zwischendurch gehen vollständig aus, Nr. 6a | `TestAntwortReadFormen`: 3 Formen (skriptierter Body); der Client sieht alle Bytes; `Close` einmal | `(n, EOF)`: die `n` Bytes verworfen: Form 2 rot; `(0, nil)` als Ende gedeutet: Form 3 rot | andere |
| A20 | HEAD, 204, 304 ohne Body, eine gesetzte `Content-Length` bleibt, Nr. 6d | `TestAntwortOhneBody`: 3 Fälle; der Fake liefert Status, `Content-Length: 7` bei HEAD und leeren Body; der Client sieht den Header bei HEAD, keinen Body in allen drei | Adapter setzt `Content-Length` auf 0 bei HEAD: rot | andere |
| A21 | `Close` genau einmal auf jedem Pfad, nie vor dem letzten `Read`, nie ein `Read` danach, Nr. 6c, 7b | `TestCloseEinmal`: 6 Pfade (Erfolg 200; Erfolg HEAD; Lesefehler; Statusverstoß; 204 mit Bytes; Fehler `ODR-E5001` zugleich mit einer Antwort); der Body-Fake zählt `Close` (je 1), `Read` nach `Close` (0) und `Read` nach EOF | `Close` fehlt im Fehlerpfad (Pfad 3, 4, 6 rot); `defer` und zusätzliches `Close` (Zähler 2): rot | andere |
| A22 | Ein Schreibfehler beendet das Kopieren, `Close` einmal, der Handler kehrt zurück, Nr. 6e | `TestSchreibfehler`: der Body-Fake liefert ohne Ende 1 KiB je `Read`; der Client schließt nach dem ersten Stück; in 5 s: `Close` 1, `Read`-Zähler hört auf zu wachsen (zwei Messungen 200 ms auseinander gleich) | Schreibfehler ignoriert: Fake liefert weiter, Zähler wächst, Test rot per Frist | nebenläufig |
| A23 | Der Handler kehrt erst nach `Close` zurück; bis dahin sind `Header` und `Body` der `Anfrage` gültig und unverändert, Nr. 3, 4h | `TestLeihe`: Antwort chunked; `Close` blockiert auf einem Kanal und liest dann den Rest des Request-Bodys (muss die Bytes liefern, kein Fehler) und vergleicht die Map `Header` mit ihrem Abbild vom Aufruf; der Client liest bis kurz vor dem Ende, das Ende (EOF) kommt nicht, solange `Close` blockiert (300 ms), und kommt nach dem Freigeben | `Close` in einer Goroutine nach der Rückkehr: Ende kommt sofort, rot; Adapter löscht `X-Recorder-*` aus der Map nach dem Aufruf: Vergleich rot | nebenläufig |
| A24 | Streaming: ohne `Content-Length` erreicht ein Stück den Client, bevor das nächste gelesen wird, Nr. 6b | `TestStreaming`: der Body-Fake liefert `abc`, das zweite `Read` blockiert auf einem Kanal, den der Test schließt, sobald der Client `abc` gelesen hat (Frist 5 s); Antwort ohne Länge | kein `Flush`: der Client erhält `abc` erst mit dem zweiten Stück, das nie kommt, Test rot per Frist | nebenläufig |
| A25 | Ein Lesefehler bricht die Verbindung ab, ohne Abschluss der Antwort, ohne Fehlerobjekt, zu jedem Zeitpunkt, Nr. 9b | `TestLesefehler`: 3 Zeitpunkte × 2 Rahmungen (mit und ohne `Content-Length`) = 6: Fehler beim ersten `Read` (der Client erhält keine Antwort, `Do` meldet einen Fehler); Fehler nach dem ersten Stück (Status 200 gelesen, danach Lesefehler `io.ErrUnexpectedEOF`); `(n > 0, Fehler)` in einem Aufruf (die `n` Bytes kommen an, danach Fehler); in allen: kein Fehlerobjekt im Body, `Close` einmal | Fehler als Ende behandelt (`return` ohne Abbruch): der Client sieht ein sauberes EOF, rot; Fehlerobjekt 502 geschrieben: rot | andere |
| A26 | Jeder `*Fehler` mit gültigem Code und Status geht mit diesem Status und dem Fehlerobjekt des Codes aus; Header `Content-Type`, `Content-Length`, `X-Correlation-Id`, Nr. 7a2, 8 | `TestFehlerobjektJeCode`: Schleife über alle Paare (Code, Status) aus `model.AlleCodes()` mit `Stati()` ohne `0`: 33 Paare × 2 Bindungen = 66 (die 34 Codes tragen 33 Stati ungleich 0: `ODR-E2001` und `ODR-E2401` tragen keinen, `ODR-E4001` zwei); erwartet `Status`, Body gleich `json.Marshal(NeuesFehlerobjekt(c))`, genau drei Felder, `Content-Type: application/json`, `Content-Length` gleich der Länge, `X-Correlation-Id` gleich der ID, die der Fake sah; `ODR-E4001` mit 502 und mit 504 | fester Status 500 oder 404: jedes Paar außer einem rot; Reihenfolge der Felder vertauscht: rot; `Content-Length` fehlt: rot | andere |
| A27 | Ein `*Fehler` mit unzulässigem Paar ergibt `ODR-E9001` mit 500, Nr. 7a2 | `TestFehlerUnzulaessig`: Menge 7: (Code 0, 500), (Code 9999, 404), (`ODR-E2001`, 0), (`ODR-E5001`, 200), (`ODR-E5001`, 0), (`ODR-E5001`, 600), (`ODR-E4001`, 500); jeder liefert das Fehlerobjekt `ODR-E9001` mit Status 500 und der Adapter ruft `WriteHeader` nie mit 0 | Prüfung auf `StatusErlaubt` entfernt: (`ODR-E5001`, 200), (`ODR-E4001`, 500) rot; Prüfung `Status >= 400` entfernt: (`ODR-E2001`, 0) löst eine Panik aus, rot | andere |
| A28 | `*Fehler` wird auch verpackt und in `errors.Join` gefunden, Nr. 7a2 | `TestFehlerVerpackt`: 3 Formen (`fmt.Errorf("%w", f)`, `errors.Join(x, f)`, zweifach verpackt) mit `ODR-E5003` (400); Antwort 400 mit `ODR-E5003` | `err.(*inbound.Fehler)` statt `errors.As`: Formen 1 bis 3 rot | andere |
| A29 | Ein Fehler ohne Code ergibt `ODR-E9001` mit 500, auch `model.ErrUngueltig`, Nr. 7a3 | `TestFehlerOhneCode`: 2 Fälle (`model.ErrUngueltig`; `errors.New("x")`); 500 mit `ODR-E9001`; Body ohne den Text des Fehlers | `ErrUngueltig` als 400 gedeutet: Fall 1 rot | andere |
| A30 | Ein Kontextfehler ohne Code bekommt keine Antwort, die Verbindung bricht ab; er gewinnt gegen einen `*Fehler`, Nr. 7a1, 9a | `TestKontextfehler`: Menge 2 × 4 = 8 (`context.Canceled`, `context.DeadlineExceeded`; direkt, `%w`, `errors.Join` mit `*Fehler` `ODR-E5001`, `errors.Join` mit `model.ErrUngueltig`); der Client erhält keine Antwort (`Do` meldet einen Fehler), kein Fehlerobjekt, keine Statuszeile | Prüfung nach dem `*Fehler`: Form 3 liefert 404, rot; Kontextfehler als 500: alle Fälle rot | andere |
| A31 | Das Fehlerobjekt enthält keinen Wert des Requests und keine Ursache, Nr. 8 | `TestFehlerObjektKeineWerte`: Canary in 5 Stellen (Pfad, Query, Header-Wert, Request-Body, `Ursache`) × 2 Fehlerarten (`*Fehler` mit `Ursache` die den Canary trägt; `model.ErrUngueltig` mit Canary) = 10; der Body, alle Header und der Fehlertext des Clients enthalten ihn nicht | `message` aus `err.Error()` statt `Code.Message()`: Ursache-Fälle rot | andere |
| A32 | Bei HEAD entfällt der Body des Fehlerobjekts, die Header bleiben, Nr. 8 | `TestFehlerHead`: 1 Fall (`ODR-E5001`): Status 404, `Content-Length` gleich der Länge des Objekts, kein Body | Body trotzdem geschrieben: der Client meldet es nicht, `net/http` verwirft; Mutation `Content-Length` weggelassen: rot | andere |
| A33 | `X-Correlation-Id` steht nur auf Fehlerantworten, Nr. 8 | `TestKorrelationsHeader`: 2 Fälle (Erfolg 200 ohne den Header, auch wenn die Antwort des Dienstes einen eigenen `X-Correlation-Id` mit anderem Wert trägt: dessen Wert bleibt; Fehler 404 mit dem Header gleich der ID des Fakes) | Header an jeder Antwort: Fall 1 rot | andere |
| A34 | Gleichzeitige Requests teilen nichts, Nr. 10 | `TestGleichzeitig`: 64 Requests gleichzeitig, jeder mit eigenem Pfad, Query, Header-Wert, Body und eigener Antwort des Fakes (Echo mit der Nummer); jede Antwort gehört zu ihrem Request, die 64 `KorrelationsID` sind paarweise verschieden; unter `-race` und mit `-count=50` | Paket-Variable als Puffer oder gemeinsame `Anfrage` wiederverwendet: Antworten vertauscht, rot (Wirkung); Wirkungs-Fänger, nicht der Detektor | nebenläufig |

Zeilen A1 bis A34: **34**, davon **5 nebenläufig** (A12, A22, A23, A24, A34).

**Mengen aus dem Wortlaut** (`AGENTS.md` §3.17), jede als Produkt der Achsen:

| Menge | Achsen | Elemente | Fänger |
|---|---|---|---|
| Pfad-Formen | 13 Formen × 2 Bindungen (+ Fall `/a%zz` vor dem Handler) | 26 + 1 | A4 |
| Query-Formen | 12 Formen × 2 Bindungen | 24 | A5 |
| Methoden | 10 × 2 | 20 | A3 |
| Request-Header-Fälle | 10 × 2 | 20 | A7 |
| Body-Formen | 5 × 2 | 10 | A8 |
| Request-Body-Größen | 8 × 2 | 16 | A9 |
| Status | 9 × 2 | 18 | A13 |
| Status-Verstoß | 8 | 8 | A14 |
| Hop-by-Hop | 9 Namen × 2 Schreibweisen + 2 Connection-Listen | 20 | A16 |
| Antwort-Größen | 9 × 2 | 18 | A18 |
| (Code, Status)-Paare | 33 Paare (alle Stati ungleich 0 der 34 Codes, `ODR-E4001` mit zwei) × 2 | 66 | A26 |
| unzulässige Paare | 7 | 7 | A27 |
| Kontextfehler | 2 Fehler × 4 Hüllen | 8 | A30 |
| Canary | 5 Stellen × 2 Fehlerarten | 10 | A31 |
| Lesefehler | 3 Zeitpunkte × 2 Rahmungen | 6 | A25 |
| Close-Pfade | 6 | 6 | A21 |
| Anhalten-Methoden | 3 | 3 | L4 |

Gleichstand und Grenzen: Status 200 und 599 gelten, 199 und 600 nicht (A13, A14); `Content-Length` 0 und 1 (A8); Puffergrenzen 4095/4096/4097 und 32767/32768/32769 (A9, A18); beide Seiten der Kopfgrenze (L2); `ODR-E4001` mit 502 und 504, beide gleich dem getragenen Status (A26). Die "Aufrufstellen des Ports" sind **eine** (`Verarbeiten`) und `IDs.KorrelationsID`; A10 und A11 zählen beide und alle Operationen des Fakes, auch `Drain`.

Die Zeile *Anhalten-Methoden* und der Verweis auf L2 gehören zur Tabelle von [slice-http-lebenszyklus](../open/slice-http-lebenszyklus.md); die Mengen des Kerns sind die übrigen.

**Raten (`AGENTS.md` §3.10).** Jede nebenläufige Zeile (A12, A22, A23, A24, A34) ordnet mit Kanälen. Der Implementer fährt die Mutation je nebenläufiger Zeile mit `-count=50` in einem Prozess und nennt „x von 50 rot“, für ein Datenrennen die Zahl roter Prozesse mit `-test.count=1` (mindestens 20 Prozesse), den Messenden und trägt beides in diesen Absatz ein; „hoch“ steht nie ohne Zahl. A34 ist ein Wirkungs-Fänger, nicht der Detektor. Die Fenster der Sperre (L16) gehören zu [slice-http-lebenszyklus](../open/slice-http-lebenszyklus.md).

*Raten: wird vom Implementer gefüllt.*

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-record-service` und `slice-replay-service` und `slice-race-gate` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 4800 Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/adapters/driving/httpodata/`, Tests eingeschlossen; Basis ist der Commit des Plans mit `SPEC-136`, Tabelle und Raster, vor dem ersten Code-Commit, Block unten). Schätzung S = 3525 (a=5, b=29, f=2, p=380 nach der Tabelle in §3), Marke 100 · aufrunden(135·3525/10000) = 4800; die Schnittfrage ist gestellt (S ≥ 2223) und der Lebenszyklus vor dem Code geschnitten (`slice-http-lebenszyklus`); Grund für den Kern im Block (`schnitt-grund`). Die ältere Erwartung „rund 700 Zeilen“ und die Rückführung „über 1000“ trugen keine Tabelle; sie sind ersetzt (`ersetzt` im Block).

**Größenmarke.** Schätzung *S* = 85·5 + 80·29 + 200·2 + 380 = 3525 (Zeilen der Tabelle in §3: 5 nebenläufig, nämlich A12, A22, A23, A24, A34, und 29 andere, nachgezählt; zwei Port-Fakes: ein Fake des Dienstes für `RecordUseCase` und `ReplayUseCase` in einem Typ, der `Drain` zählt, und der Test-Server samt rohem Client; Produktcode 380 nach dem Architect: `server.go` 80, `anfrage.go` 70, `antwort.go` 130, `fehler.go` 70, Paketkommentar und Konstanten 30), Marke 100·⌈135·3525/10000⌉ = 4800 (`SPEC-128` Nr. 1). Die Marke ist nach dem ersten Code-Commit fest (`SPEC-128` Nr. 5 c); der Body-Fake mit Kanälen und die Hilfen gehören zu den Zeilen der Tabelle, nicht zu den Fakes.

```groessenmarke
pfad: src/adapters/driving/httpodata/
produktcode: 380
fakes: 2
basis: b50a4aed4a9c4103b8022940a328af01428cb5cd
schaetzung: 3525
marke: 4800
schnitt-grund: Der Kern ist ein Ablauf, den Handler, Fake des Dienstes und Test-Server teilen: Anfrage bilden, Dienst rufen, Antwort oder Fehler schreiben, Close. Ein weiterer Schnitt (Fehlerobjekt, Streaming oder Abbruch) ließe den Adapter ohne Fehlerpfad oder ohne Abbruchsemantik: falsch, nicht klein. Streaming und Abbruch teilen die Kopierschleife, Fehlerobjekt und Abbruch die Reihenfolge der Prüfung von err (der Kontextfehler gewinnt). Der Teil ohne geteilten Mechanismus, der Lebenszyklus (Zustand der Zulassung, Anhalten, Schliessen, 503 mit ODR-E5010, Grenzen des Servers), ist als slice-http-lebenszyklus vor dem Code geschnitten.
ersetzt: 1000 | Die ältere Rückführung über 1000 Zeilen (Erwartung rund 700) trug keine Tabelle Zusage-Fänger und keine Formel; die Marke gilt nach der Tabelle in §3 und SPEC-128 Nr. 1.
```

Messpunkt am Start (`make groessenmarke`, vor dem ersten Code-Commit): `zeilen=34 nebenlaeufig=5 fakes=2 produktcode=380 schaetzung=3525 marke=4800 messung=0`, Tabelle ergibt 3525 gleich der Schätzung des Blocks.
- `in-progress` → `open` (blockiert): entfällt — `RequestURI` und `r.Body` genügen, kein Typ von `net/http` erreicht den Port (Vorabentwurf des Architect).

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

**Risiken mit Ausgang** (Entscheidungen des Architect im Vorabentwurf):

1. **Header-Behandlung** — **Ausgang:** entschieden (Architect), Ort `SPEC-136` Nr. 4d, 5c, 5d, 8. Raster *Header-Art × Behandlung*:

   | Header-Art | Request (Nr. 4d) | Antwort (Nr. 5c, d) |
   |---|---|---|
   | normal, ein Wert | unverändert | unverändert |
   | mehrfach (mehrere Zeilen) | alle Werte in Reihenfolge | alle Werte in Reihenfolge |
   | Schreibweise | kanonisch (`net/http`), nur in der Schreibweise verschiedene Namen unter einem Namen | kanonisch beim Schreiben |
   | leerer Wert | bleibt | bleibt |
   | `Set-Cookie` | unverändert (Dienst entscheidet) | mehrfach, unverändert |
   | `Host` | nie in `Header` (`r.Host`) | – |
   | `Connection` und die Liste | bleibt in `Header` (Dienst filtert) | entfällt, samt der Namen, die ihr Wert nennt |
   | `Transfer-Encoding` | nie in `Header`, `Laenge` -1 | entfällt |
   | `Content-Length` | bleibt in `Header` | bleibt, wenn der Dienst sie liefert; sonst bestimmt `net/http` |
   | `Content-Encoding` | bleibt | bleibt (Record: die Bytes sind gepackt); Replay liefert keins |
   | `Content-Type` fehlt | – | kein Ergänzen (kein Sniffing) |
   | `Date` fehlt | – | `net/http` setzt es |
   | `Expect: 100-continue` | bleibt in `Header`; `net/http` antwortet `100` beim ersten Lesen des Bodys | – |
   | `X-Recorder-*` | bleibt in `Header` | – |
   | `Upgrade` | bleibt in `Header` | entfällt |

2. **Client bricht ab, Antwort bricht ab nach dem ersten Byte** — **Ausgang:** entschieden, Ort `SPEC-136` Nr. 9 (a bis c), 6e, 6f; Fänger A12, A22, A25, A30. Die nebenläufigen Fänger ordnen mit Kanälen; jeder läuft vor der Übergabe mit `-count=50` und unter `make test-race`.
3. **Pfad ohne Service-Root, Prozent-Kodierung, `raw_query` mit ungültiger Kodierung** — **Ausgang:** entschieden, Ort `SPEC-136` Nr. 4b, 4c, 4i: der Adapter reicht unverändert; ein Pfad ohne Service-Root ist `ODR-E5005` (Dienst, 404); eine ungültige Query erreicht den Dienst (`ODR-E5006` nur bei `MATCH_INVALID_ENCODING=reject`); ein Pfad mit ungültigem Escape erreicht ihn nie (`net/http`, 400 Klartext), ein roher `#` ergibt `ODR-E9001`.
4. **Request-Größe und Zeitgrenzen** — **Ausgang:** entschieden, Ort `SPEC-136` Nr. 11; geliefert vom Slice `slice-http-lebenszyklus` (Fänger L1, L2). Request-Body und Antwort sind unbegrenzt (`MAX_CAPTURE_BODY` begrenzt die Aufnahme).
5. **Herunterfahren in zwei Schritten** — **Ausgang:** entschieden, Ort `SPEC-136` Nr. 13 bis 15; die Reihenfolge ist gegen `LH-FA-02.c`, `SPEC-059` und `SPEC-124` Nr. 14 gelesen (Listener nimmt nichts mehr an, `CloseAll` mit `SHUTDOWN_TIMEOUT`, Verbindungen schließen mit dem Rest der Frist, dann die Repositories). Geliefert vom Slice `slice-http-lebenszyklus` (Fänger L3 bis L16); der Root-Teil (Reihenfolge, Rest der Frist) liegt bei `slice-bootstrap-m2` (Sendung in 7.3).
6. **Randformen (`AGENTS.md` §3.12, §3.16)** — **Ausgang:** entschieden; das Raster steht in 5, jede Zelle hat einen Ort in `SPEC-136` oder "gilt nicht" mit Grund.

**Neue Risiken des Architect (jeweils mit Ausgang):**

7. **Annahmen über `net/http`** (Pfad mit ungültigem Escape ergibt 400 vor dem Handler, `ErrAbortHandler` schließt ohne Abschluss, `Content-Type` nil unterdrückt das Sniffing, `Flush` sendet sofort, `ctx` bei ungelesenem Body) — **Ausgang:** offen, Adresse Implementer; je Annahme der Fänger A4 (Fall 13), A25, A17, A24, A12 als erste Tests; weicht ein Verhalten ab, geht die Zelle an den Architect.
8. **Zähler und Gesamtspan fehlen im Adapter** (`SPEC-136` *Ungeprüft*) — **Ausgang:** weiter offen, Adresse `slice-telemetry-einbau-gateway` (nimmt es an: sein Ziel nennt "der Zähler je Fehlercode steigt für jeden Fehler, den der Adapter als Fehlerobjekt schreibt" und den Gesamtspan).
9. **Replay-Antworten gehen chunked** (keine `Content-Length`, `Flush` je Stück) — **Ausgang:** akzeptiertes Negativ: `LH-QA-07.a` zählt `Content-Length` nicht zu den relevanten Headern, `SPEC-135` Nr. 13 b nimmt die Länge bewusst aus; Folge: ein Client sieht `Transfer-Encoding: chunked` statt einer Länge.

**Weitere Zeilen mit Ausgang** (Planner, nach dem Vorabentwurf):

- Die Sendung von `slice-replay-service` „Länge und Coding setzt der Adapter selbst“ (dessen §6 Sendung 1) widerspricht `SPEC-136` Nr. 5d nicht, sobald der Wortlaut gilt: der Adapter setzt weder `Content-Length` noch `Content-Encoding` noch `Transfer-Encoding`, `net/http` rahmt; der Plan (§1 *Übernimmt (Replay)*, DoD) ist nachgezogen — **Ausgang:** entschieden (Architect), Ort `SPEC-136` Nr. 5d.
- Die Zahl roter Läufe je nebenläufiger Zeile (A12, A22, A23, A24, A34) und die Prozesszahl für Datenrennen — **Ausgang:** offen; Adresse Implementer, Absatz *Raten* in §3 (`AGENTS.md` §3.10).
- Eine Zelle des Rasters oder eine Randform ohne Ort in `SPEC-136` — **Ausgang:** offen; Adresse Implementer, der anhält und die Zelle an den Architect zurückgibt (`AGENTS.md` §3.12), mit einer Zeile in diesem Abschnitt im selben Commit wie sein Bericht (`AGENTS.md` §3.13).
- Der Satz in `harness/README.md` §Sensors *Nicht behauptet* („`make a-check` prüft damit erst eine Schicht“) und die Aussage über ein Datenrennen, das kein Lauf auslöst (das Fenster L16 des Lebenszyklus ist ein Beispiel) — **Ausgang:** offen; Adresse Planner bei der Closure dieses Slice bzw. nach dem Merge (`AGENTS.md` §3.9): die Schichten mit Code zählen und den Satz anpassen.
- Der Stempel der Gegenproben nach dem Merge (Vorabentwurf des Architect, Abschnitt Integrationsschritt) — **Ausgang:** offen; Adresse Integrationsschritt, der prüft, ob eine Eingabe der Gegenproben sich geändert hat; dieser Slice ändert weder `Dockerfile` noch `Makefile` noch Gegenproben.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect). Die Zeilen `Anhalten` und `Schliessen` und die Zusatzzeilen mit Fänger `L2` und `L4` gehören zu [slice-http-lebenszyklus](../open/slice-http-lebenszyklus.md) (dessen Tabelle trägt die Fänger); die Zelle nennt den Ort in `SPEC-136`, der für beide Slices gilt.

*Fehlend* heißt nil oder der Nullwert, *leer* ein vorhandener leerer Wert, *fremd* eine Form, die der Vertrag gesondert behandelt. "Abbruch" ist der Abbruch durch den Client (`ctx`), "Wiederholung" derselbe Wert in zwei Aufrufen. "Vor dem Handler" (V) ist `SPEC-136` Nr. 4i.

| Operation · Argument | fehlend | leer | fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `New` · `Verarbeiten` | Nr. 1 | gilt nicht: eine Funktion hat keinen leeren Zustand | gilt nicht: jede Funktion der Signatur gilt; ein Typed-nil ist *Ungeprüft* | gilt nicht: kein Kontext, kein Strom | Nr. 1, 10: zwei `New` mit derselben Funktion geben zwei Server ohne gemeinsamen Zustand |
| `New` · `IDs` | Nr. 1 | gilt nicht: ein Interface hat keinen leeren Zustand | gilt nicht: ein leerer Rückgabewert ist *Ungeprüft* (`SPEC-114` Nr. 6 sichert 36 Zeichen) | gilt nicht: kein Kontext | Nr. 3a: einmal je Request, drei Requests geben drei Werte |
| `New` · `Listener` | Nr. 1 | gilt nicht: ein Interface hat keinen leeren Zustand | Nr. 12: ein geschlossen übergebener `Listener` ergibt einen Fehler von `Serve` | gilt nicht: kein Kontext | gilt nicht: derselbe `Listener` in zwei Servern ist ein Aufruferfehler (*Ungeprüft*) |
| `Serve` · Zustand des `Listener` | gilt nicht: `New` weist nil ab (Nr. 1) | gilt nicht: kein Wert | Nr. 12: von außen geschlossen ergibt einen Fehler | Nr. 12, 13, 14: `Anhalten` und `Schliessen` beenden mit `nil` | Nr. 12 *Ungeprüft*: ein zweites `Serve` |
| Request · Methode | V: `net/http` weist einen Request ohne Methode ab | V: dito | Nr. 4a: jedes Token, byte-gleich | Nr. 9a | gilt nicht: eine Methode wiederholt sich nicht; zwei Requests teilen nichts (Nr. 10) |
| Request · Pfad | Nr. 4c: ein Target ohne `/` geht als `Pfad` an den Dienst (`ODR-E5005`) | Nr. 4b: Pfad und `?` ohne Query ergeben `RawQuery` leer; ein leerer Pfad kommt nicht vor (V) | Nr. 4b, 4c, 4i: Dot-Segmente, `//`, Escapes, rohe Bytes unverändert; ungültiger Escape V; roher `#` ergibt `ODR-E9001` (*Ungeprüft*) | Nr. 9a | Nr. 10: kein Zustand, `Pfad` je Request neu |
| Request · `RawQuery` | gilt nicht: ein fehlendes `?` ist leer (Nr. 4b) | Nr. 4b: `?` allein und kein `?` sind gleich | Nr. 4b, 4i: jede Folge unverändert, auch ungültige Kodierung (Dienst: `ODR-E5006`) | Nr. 9a | Nr. 10 |
| Request · Header | Nr. 4d: die Map ist nie nil | Nr. 4d: ein leerer Wert bleibt | Nr. 4d: `Host`, `Transfer-Encoding` nie in `Header`; `X-Recorder-*`, `Connection`, `Expect` unverändert; Namen kanonisch | Nr. 9a | Nr. 4d: zwei Zeilen desselben Namens, Werte in Reihenfolge |
| Request · Body und `Laenge` | Nr. 4e: `NoBody` ergibt nil und 0 | Nr. 4e: `Content-Length: 0` ergibt nil und 0; chunked leer ergibt Strom und -1 | Nr. 4e: `Content-Length` und chunked; keine Grenze für die Größe | Nr. 9a, 4g: bei ungelesenem Body erst nach dem Lesen (*Ungeprüft*) | gilt nicht: ein Strom wird einmal gelesen, vom Dienst |
| Request · Korrelations-ID | gilt nicht: `IDs` liefert immer einen Wert | Nr. 8 *Ungeprüft*: ein leerer Wert | gilt nicht: die Form sichert `SPEC-114` | gilt nicht: der Aufruf hat keinen Kontext | Nr. 3a: einmal je Request, nie zweimal |
| Antwort · Status | Nr. 5a: 0 ergibt `ODR-E9001` | gilt nicht: eine Zahl hat keinen leeren Zustand | Nr. 5a, 5b: 200 bis 599 unverändert, alles andere `ODR-E9001` | Nr. 9a, 6e | gilt nicht: eine Antwort wird einmal geschrieben |
| Antwort · Header | Nr. 5c: nil oder leer ergibt keine Header | Nr. 5c: ein leerer Wert bleibt | Nr. 5c: Hop-by-Hop und `Connection`-Liste entfallen; `Content-Type` fehlt: kein Sniffing (Nr. 5d); ungültige Namen oder Werte `net/http` (*Ungeprüft*) | Nr. 9a | Nr. 5c: mehrere Werte eines Namens in Reihenfolge, `Set-Cookie` mehrfach |
| Antwort · Body | Nr. 5a: nil ergibt `ODR-E9001`, kein `Close` | Nr. 6a, 6d: ein leerer Body ist erlaubt (kein `Write`) | Nr. 6a: `(n, err)`, `(0, nil)`, Fehler; Nr. 6d: Bytes hinter 204/304 | Nr. 6e, 9a, 9b: Schreibfehler, Abbruch, Lesefehler | Nr. 6c: `Close` genau einmal auf jedem Pfad |
| Fehler · `err` | gilt nicht: `nil` ist kein Fehler (Nr. 3e) | gilt nicht: ein `error` hat keinen leeren Zustand | Nr. 7a: Kontextfehler vor `*Fehler` vor sonst; ungültiges Paar `ODR-E9001`; verpackt und `Join` | Nr. 9a | gilt nicht: ein Fehler wird einmal ausgeliefert |
| `Anhalten` · Zustand | gilt nicht: kein Argument | gilt nicht: kein Argument | Nr. 13b: Request danach 503 | gilt nicht: `Anhalten` hat keinen Kontext und keine Frist | Nr. 13d: wirkungslos, auch gleichzeitig |
| `Schliessen` · `frist` | gilt nicht: eine Zahl hat keinen fehlenden Zustand; `0` ist "nicht warten" (Nr. 14) | Nr. 14: `0` | Nr. 14: negativ ergibt `model.ErrUngueltig` ohne Änderung | Nr. 14: Frist abgelaufen schließt die Verbindungen | Nr. 14: das zweite ergibt `nil` |

**Zusatztabelle: Paare von Verstößen am selben Schlüssel und je Modus** (Spalte *Fänger*):

| Paar oder Modus | Ergebnis | Ort | Fänger |
|---|---|---|---|
| Body nil und Status 0 | `ODR-E9001`, 500, `Close` nicht gerufen | Nr. 5a | A14 (Fall 8) |
| Status ungültig und Body gesetzt | `ODR-E9001`, 500, `Close` einmal, keine Header der Antwort | Nr. 5a, 6c | A14 (Fälle 2 bis 7) |
| Fehler und Antwort zugleich | der Fehler gilt, `Close` einmal | Nr. 7b | A21 (Pfad 6) |
| `*Fehler` und Kontextfehler | Kontextfehler gewinnt, keine Antwort | Nr. 7a1 | A30 (Hülle 3) |
| unzulässiges Paar und Kontextfehler | Kontextfehler gewinnt | Nr. 7a1 | A30 (alle Hüllen mit `model.ErrUngueltig`) und A27 ohne Kontext |
| `model.ErrUngueltig` und Kontextfehler | Kontextfehler gewinnt | Nr. 7a1 | A30 (Hülle 4) |
| Lesefehler und Schreibfehler (Client weg) | das Kopieren endet am ersten, `Close` einmal | Nr. 6e, 9b | A21, A22 |
| angehalten und Fehler des Dienstes | angehalten gewinnt, `Verarbeiten` nicht gerufen | Nr. 3b | L4 |
| angehalten und ungültiger Kopf (zu groß) | `net/http` weist zuerst ab (431), keine 503 | Nr. 4i | L2 |
| Antwort mit `Content-Length` und Lesefehler im Body | Abbruch, nicht abgeschlossen | Nr. 9b | A25 (Rahmung mit Länge) |
| Modus `record` und `replay` | gleiche Abbildung, byte-gleich; nur der eigene Dienst wird gerufen | Nr. 2 | A3 bis A9, A11, A13, A15, A16, A18, A26 (jede Menge ×2) |

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (9×, Ausgang `verkörpert` in `AGENTS.md` §3.10); der Slice ist nebenläufig (Abbrüche, Herunterfahren): die nebenläufigen Fänger ordnen mit Kanälen und laufen im Ziel mit `-race` aus `slice-race-gate`, und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
