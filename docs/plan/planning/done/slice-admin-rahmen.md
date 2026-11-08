# Slice slice-admin-rahmen: Admin-Rahmen, Teil 2 — Konfiguration, Routen, Zuordnung und der Code `ODR-E2102`

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-238` Nr. 1 (`Konfig`, `New`), Nr. 2 (Routen, Zuordnung, `Anfrage`, 404 und 405), Nr. 3 a (Korrelations-ID), Nr. 5 (Fehlerobjekt für 404 und 405), Nr. 6 (Code `ODR-E2102`), Nr. 14 (Katalog) · `SPEC-239` — nicht berührt · `SPEC-240` (Zeile `ODR-E2102`) · `SPEC-109` Nr. 2 (Zeile `ODR-E2102`) · `SPEC-230` Nr. 8 (40 Codes) · `SPEC-114` (Korrelations-ID) · `SPEC-006` · `ARC-007`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie. `SPEC-238` bis `SPEC-240` stehen auf `main` (seit dem Merge von `slice-admin-listener-token`).

**Verantwortlich:** pt9912 (Zweig `slice-admin-rahmen`, Worktree `/Development/odata-recorder-wt-admin-rahmen`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912 (Planner). **Datum:** 2026-11-07.

**Herkunft:** Teil 2 des Schnitts von [slice-admin-listener-token](../done/slice-admin-listener-token.md) (Vorschlag des Architect, Rückgabe R1; Schnitt vor dem Code vollzogen vom Planner am 2026-11-07 nach `SPEC-128` Nr. 1, `AGENTS.md` §3.14). Die Teile in der Reihenfolge ihres Starts: 1 `slice-admin-listener-token` (Token-Datei), 2 dieser Slice, 3 [slice-admin-authentisierung](../done/slice-admin-authentisierung.md), 4 [slice-admin-lebenszyklus](../done/slice-admin-lebenszyklus.md), 5 [slice-admin-fehlerabbildung](../open/slice-admin-fehlerabbildung.md), 6 [slice-admin-abweisen-audit](../open/slice-admin-abweisen-audit.md). Der Plan trägt die Zeilen K1 bis K3, Z1, Z3 bis Z8 und Y2 bis Y4 der Tabelle des Architect (13 von 102); sie stehen im Wortlaut in §3.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Paket `src/adapters/driving/admin` führt `Konfig` und `New` (`Listener`, `IDs`, `Audit`, `Telemetrie` und `Token` geprüft, die Routen geprüft und kopiert), ordnet Methode und Pfad einer Route zu (kein Treffer: 404, falsche Methode: 405 mit `Allow`, die eingebaute `GET /healthz` steht in der Zuordnung), reicht der Route die `Anfrage` ohne `Authorization`, vergibt die Korrelations-ID vor allem anderen und schreibt das Fehlerobjekt (an der einen Stelle, die es schreibt) für 404 und 405; dazu der neue Code `ODR-E2102` (400, 404, 405; 40. Code) samt Katalog-Ausnahme, Zähler-Reihe und Tests der Codetabelle. Der Teil authentisiert noch nicht (die Tests senden einen Bearer-Header, den er nicht liest), zählt noch nicht und auditiert noch nicht.

**Übernimmt (Sendung von `slice-admin-listener-token`, Schnitt vom 2026-11-07, `AGENTS.md` §3.13):** den Code der HTTP-Antwort auf `model.ErrUngueltig`, soweit er die Codetabelle betrifft (Entscheidung des Architect: `ODR-E2102`, `SPEC-238` Nr. 6, `SPEC-240`); die Abbildung von `ErrUngueltig` auf diesen Code liegt in `slice-admin-fehlerabbildung`. Die Sendungen von `slice-session-use-case` (Code für `model.ErrUngueltig`) und `slice-inspection-use-case` (Code für einen ungültigen Seitenparameter) liegen mit ihren Adressen bei `slice-admin-listener-token`; die Teile tragen sie.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Authentisierung, `GET /healthz` ohne Token, Fehlerobjekt außer 404 und 405, Header der Antwort — `slice-admin-authentisierung` (baut auf der Zuordnung dieses Teils auf: 401 vor 404 und 405 braucht die Zuordnung zuerst).
- `Serve`, `Anhalten`, `Schliessen`, Grenzen des `http.Server` — `slice-admin-lebenszyklus`; dieser Teil hat keinen `http.Server` und nimmt `ServeHTTP` direkt (`httptest`).
- Abbildung der Fehler der Routen, Panik, Zählen (`CountError`) — `slice-admin-fehlerabbildung`; die Felder `Telemetrie` und `Audit` sind geprüft und gehalten, aber nicht gerufen (K3).
- Audit des Abweisens und die Art `admin_abgewiesen` — `slice-admin-abweisen-audit`.
- Die Token-Datei (`LadeToken`) — `slice-admin-listener-token` (Teil 1); dieser Teil prüft nur die Länge des übergebenen `Token` (K2).
- Die Endpunkte — `slice-admin-sessions`, `slice-admin-imports`, `slice-admin-workspaces`; die Tests registrieren eine Testroute.
- TLS, `ADMIN_*` aus der Umgebung, Verdrahtung — `slice-admin-tls`, `slice-bootstrap-m6`; das Binary bleibt unverändert.

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

- [x] Konfiguration, Routen und Zuordnung: `Konfig` und `New` nach `SPEC-238` Nr. 1 (nil-Prüfung der vier Ports, Token ab 32 Bytes, `New` ruft keinen Port und startet nichts), Routen nach Nr. 2 a bis d (Methode, Muster, nil-`Handler`, Überlappung samt eingebauter `GET /healthz`), Zuordnung (Literal bytegleich, Parameter auf jedes nichtleere Segment, kein Bereinigen des Pfades), 404 und 405 (`Allow` sortiert, die eingebaute Route eingeschlossen) mit `ODR-E2102`, `Anfrage` mit allen Feldern ohne `Authorization`, `KorrelationsID` vor allem anderen; der Adapter hat keine Standardadresse und ruft nirgends `net.Listen`. Bindung der Fakes und des Adapters an die Ports durch Zusicherungen zur Übersetzungszeit (`bindung_test.go`).
- [x] Code `ODR-E2102` (`SPEC-240`, 40. Code; Zahlen der Tests 39 auf 40, `ODR-E6009` von `slice-mock-schreiben` ist der 39.): `src/hexagon/model/fehler.go` und `fehler_test.go` (Status 400, 404, 405, Ziel `admin`, Text `Admin request invalid`, hinter `ODR-E2101`), die Reihe des Zählers (`SPEC-230` Nr. 8, `regeln_test.go`), die Liste in `audit_test.go` des Audit und die Zeile in `harness/katalog-ausnahmen.txt` **im selben Commit wie `fehler.go`** (`make katalog-check` ist sonst rot, `SPEC-201` Nr. 3 e).
- [x] `make gates` grün (Verifier im Zweig, `EXIT=0`; Planner auf `main` nach dem Merge).
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [x] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Für diesen Slice: `ODR-E2102` mit Ausnahme in `harness/katalog-ausnahmen.txt` (`SPEC-238` Nr. 14); das Binary bleibt unverändert (kein Aufruf von `admin` aus `bootstrap`, `src/bootstrap` ohne Änderung), in §7 vermerkt.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/adapters/driving/admin/` | neu | `Konfig`, `New`, Routen, Zuordnung, `Anfrage` und `Antwort`, die eine Stelle, die das Fehlerobjekt schreibt (404, 405), Aufruf von `IDs.KorrelationsID` (`SPEC-238` Nr. 1, 2, 3 a, 5) |
| `src/hexagon/model/fehler.go`, `src/hexagon/model/fehler_test.go` | update | Code `ODR-E2102` hinter `ODR-E2101` (`CodeE2102`, Zeile der Tabelle mit 400, 404, 405, Ziel `admin`, Text `Admin request invalid`; Literal-Tabelle und Zählung im Test) |
| `harness/katalog-ausnahmen.txt` | update | Zeile `ODR-E2102` mit Grund, direkt unter `ODR-E2101` |
| `src/adapters/driven/telemetry/regeln_test.go`, `src/adapters/driven/audit/audit_test.go` | update (nur Tests) | Literal-Liste der Codes um `ODR-E2102` |
| `src/hexagon/services/mock/lesen_fehler_test.go`, `src/adapters/driving/httpodata/fehler_test.go` | update (nur Tests) | Folge des Codes `ODR-E2102` in bestehenden Tests, die die Codes oder ihre Stati zählen: `codeListen` führt `ODR-E2102` unter den übrigen Codes (sonst ist `TestRueckgabenMenge` rot); `TestFehlerobjektJeCode` rechnet die Zahl der Paare mit den drei Stati des Codes (Auftrag S2) |
| Tests | neu | Adapter gegen `ServeHTTP` mit `httptest`; Fakes `zaehlendeIDs`, `routeFake`, `listenerFake`; `FakeAudit` und `FakeTelemetry` aus `supporttest`; `bindung_test.go`; der Adapter importiert weder Adapter noch Service (`make a-check`) |
| `spec/spezifikation.md` | geliefert (Architect, im Zweig von `slice-admin-listener-token`) | `SPEC-238` bis `SPEC-240`, Nachzug `SPEC-109`, `SPEC-230`; dieser Teil schreibt sie nicht |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stammen vom Architect (2026-11-07, Plan von [slice-admin-listener-token](../done/slice-admin-listener-token.md)); der Schnitt gibt jedem Teil ganze Zeilen im Wortlaut, `Nr.` ist eine Kennung und keine Zählung (K Konfiguration, Z Zuordnung, Y Katalog und Code). **Wachsende Mengen:** Die Schleifen über die Requestarten *R* (A1, E4 in `slice-admin-authentisierung`) und die Antwortarten (E2) wachsen mit den Teilen; dieser Teil trägt in *R* die Formen 404 und 405 und die Route mit 2xx, die Zählprüfung der späteren Teile rechnet die Formen der Vorgänger mit.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| K1 | (`SPEC-238` Nr. 1) `New` weist `Listener`, `IDs`, `Audit` und `Telemetrie` je mit nil ab: `model.ErrUngueltig`, kein `Server`, der Text nennt keinen Wert | Schleife über die vier Felder, daneben der Gutfall; `errors.Is(err, model.ErrUngueltig)`, Server nil | Prüfung eines Feldes entfernt | andere |
| K2 | (Nr. 1) Ein `Token` unter 32 Bytes ist ungültig, 32 Bytes sind gültig; der Fehlertext nennt das Token nicht | Fälle 0, 1, 31, 32, 33 Bytes; ein Canary der Länge 31 (aus Bytewerten) steht nicht im Text | `< 32` zu `<= 32` oder zu `< 31` | andere |
| K3 | (Nr. 1) `New` ruft keinen Port und keine Methode des `Listener`, öffnet keinen Socket (kein Aufruf von `Listen`) und startet keine Goroutine | Fakes zählen jede Operation (`FakeAudit.Aufrufe`, `FakeTelemetry.Aufrufe`, `zaehlendeIDs`, `listenerFake` mit `Accept`, `Close` und `Addr`): alle 0; Zahl der Goroutinen vor und nach `New` gleich; AST-Test (`go/parser`) über die Dateien des Pakets ohne `_test.go`: 0 Aufrufe mit dem Namen `Listen`, `ListenTCP`, `ListenAndServe` oder `ListenAndServeTLS`, Zählprüfung der vier Namen im Test | `New` ruft `IDs.KorrelationsID()` zur Probe; `net.Listen("tcp", ":8081")` in `New` | andere |
| Z1 | (`SPEC-238` Nr. 2 a bis c) Eine Route mit ungültiger Methode, ungültigem Muster oder nil-`Handler` macht `New` zu `ErrUngueltig`, der Text nennt weder Methode noch Muster; gültig sind alle fünf Methoden, jedes Zeichen aus `A-Za-z0-9._~-` im Literal, Parameter wie `{id}` und `{a_b}`, mehrere Parameter, und eine leere Liste der Routen | Generator mit Orakel aus den Zeichenmengen des Vertrags: ungültige Methoden 5 (leer, `get`, `HEAD`, `OPTIONS`, `GET ` mit Leerzeichen) gegen ein gültiges Muster; feste ungültige Muster 12 (leer, `x`, `/`, `/a/`, `//a`, `/a//b`, `/{}`, `/{a`, `/a}`, `/{a}{b}`, `/a{b}`, `/a/{b}/{b}`); Schleife über alle 256 Bytewerte *c*: `/a` + *c* gültig genau für die 66 Zeichen der Menge, `/{` + *c* + `}` gültig genau für `a-z` (26), `/{a` + *c* + `}` gültig genau für `a-z` und `_` (27); nil-`Handler` 1; Gutfälle: die fünf Methoden, `/__admin/sessions/{id}/close`, mehrere Parameter, leere Liste; Canary (aus Bytewerten) in Methode und in Muster der ungültigen Fälle steht nicht im Text; Zählprüfung: Zahl der Fälle gleich der Summe der Elemente der Achsen | Zeichenmenge des Namens `[a-z]` zu `[A-Za-z]`; `_` im Rest des Namens entfernt oder Ziffer im Rest erlaubt; Prüfung der doppelten Namen entfernt; Zeichen `~` aus der Menge entfernt oder `:` ergänzt; das Muster im Text des Fehlers | andere |
| Z3 | (Nr. 2 d) Zwei Routen überlappen bei gleicher Methode, gleicher Segmentzahl und je Stelle gleichem Literal oder einem Parameter, und die eingebaute `GET /healthz` zählt mit; sonst sind sie gültig; der Text nennt kein Muster | Generator über Paare aus 120 Mustern (alle Folgen von ein bis vier Segmenten über Literal `a`, Literal `b` und Parameter) × Methode gleich oder verschieden; Orakel ein Modell des Tests je Segment; dazu `GET /{x}` und `GET /healthz` ungültig, `POST /healthz` und `GET /healthz/x` gültig; ein Canary im Muster steht nicht im Text | Segmentzahl nicht verglichen; Methode nicht verglichen; eingebaute Route nicht gezählt | andere |
| Z4 | (Nr. 1, 2) Zuordnung: ein Literal passt bytegleich auf das dekodierte Segment, ein Parameter auf jedes nichtleere Segment (auch `.`, `..`, einen Wert mit `%2F`, UTF-8 und Bytes, die kein UTF-8 sind), die Segmentzahl muss gleich sein; `New` kopiert die Liste der Routen; Status und Body der Route (auch leer) gehen durch, ein Fehler der Route ist vorläufig `ODR-E9001` mit 500 (die Abbildung liefert `slice-admin-fehlerabbildung`) | Generator über Pfade (Literal gleich; in Groß- oder Kleinschreibung abweichend; als Prozentfolge `%61` und `%41`; Parameter in {`x`, `.`, `..`, `a%2Fb`, `%C3%A4`, `%FF`, `%00` und zwölf Formen, in denen Pfad- und Query-Dekodierung sich unterscheiden: `+`, `a+b`, `%2B`, `%2b`, `%20`, `;`, `%3B`, `%25`, `&=`, `%3F`, `%23`, `~`} je Parameterstelle und das leere Segment (404); Segmentzahl ±1) × 3 Muster (`/a`, `/a/{id}`, `/{p}/c/{q}`); `routeFake` hält `Pfadwerte`; Orakel Literale des Tests; Zählprüfung: Fälle gleich Summe der Elemente der Achsen; Antworten der Route 204 ohne Body, 200 und 201 mit Body, ein `errors.New` (500, `ODR-E9001`); nach `New` wird die Liste der Routen überschrieben (Methode, Muster, Handler), die Zuordnung bleibt | Vergleich ohne Beachtung der Schreibweise; Literal gegen das rohe Segment; Parameter nimmt das leere Segment; Status der Route nicht geschrieben; Liste ohne Kopie gehalten | andere |
| Z5 | (Nr. 2) Der Adapter bereinigt den Pfad nicht, und ein Pfad ohne führenden `/` passt auf keine Route | Routen `GET /a/{id}` und `POST /{x}`; 8 Pfade (`//`, `/a/`, `//a`, `/a//b`, `/a/./b`, `/a/../a/b`, `*`, leer) × Methode (GET, POST) mit Token-Header: 404 `ODR-E2102`, `routeFake` 0 Aufrufe; Zählprüfung 16 | `path.Clean` vor der Zuordnung; Prüfung des führenden `/` entfernt (`POST *` trifft `/{x}`) | andere |
| Z6 | (Nr. 2, 5) Passt der Pfad auf keine Route: 404; passt er, aber nicht die Methode: 405 mit `Allow` aller Methoden der Routen, die auf den Pfad passen (die eingebaute eingeschlossen), nach Bytefolge, durch `, ` getrennt; beide tragen das Fehlerobjekt `ODR-E2102` mit `Content-Type` und `Content-Length` und ohne Wert aus dem Request; die eingebaute `GET /healthz` antwortet 200 mit `{"status":"ok"}` | Routen `GET /a/b`, `POST /a/{id}`, `DELETE /a/{id}`, `PUT /b`; Schleife Methoden (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE) × Pfade (`/a/b`, `/b`, ein unbekannter Pfad mit Canary aus Bytewerten, `/healthz`); Orakel Literale: `/a/b` GET, POST, DELETE die Route (bei POST und DELETE `id` gleich `b`), sonst 405 mit `DELETE, GET, POST`; `/b` PUT die Route, sonst 405 mit `PUT`; unbekannt 404; `/healthz` GET 200 mit dem Literal, sonst 405 mit `GET`; Body, `Content-Type: application/json` und `Content-Length` als Literale, der Canary steht in keinem Byte der Antwort; Zusatzfälle: Methode `get` auf `/b` (405 mit `PUT`), auf `/a/b` (405 mit `DELETE, GET, POST`) und auf `/healthz` (405 mit `GET`), `Get` auf `/healthz` (405 mit `GET`), leere Liste der Routen (`GET /healthz` 200, `GET /x` 404, `POST /healthz` 405 mit `GET`); Zählprüfung 32 + 7 | `Allow` unsortiert, nur vom ersten Muster oder ohne die eingebaute Route; 404 und 405 vertauscht; Methode ohne Beachtung der Schreibweise verglichen; `Content-Length` fehlt | andere |
| Z7 | (Nr. 2, 3 a) `Anfrage` trägt Methode, dekodierten Pfad, `RawQuery` unverändert, `KorrelationsID`, `Pfadwerte` (nie nil, leer ohne Parameter), `Header`, `Body` und `Laenge` (`-1` bei unbekannter Länge); `IDs.KorrelationsID()` läuft je Request genau einmal, vor allem anderen, und kein anderer Aufruf von `IDs` | `routeFake` hält die `Anfrage`; Schleife über die acht Felder mit je eigener Eingabe (Pfad mit `%2F`, dekodiert `/`; Query mit `%`-Folgen; Länge 0, 5 und chunked); Schleife über Requestarten (Route 2xx, 404, 405, `GET /healthz`): `zaehlendeIDs` zählt `KorrelationsID` 1 und die vier anderen Operationen 0, die gehaltene Kennung gleich `Anfrage.KorrelationsID`; ein Haltepunkt in `KorrelationsID` sieht `routeFake`, `FakeAudit` und `FakeTelemetry` bei 0 Aufrufen | ein Feld nicht gesetzt; `Pfadwerte` nil bei einer Route ohne Parameter; `KorrelationsID()` nach der Zuordnung oder ein zweites Mal gerufen | andere |
| Z8 | (Nr. 2) `Authorization` erreicht keine Route, jede andere Header-Zeile bleibt, und `Header` ist eine Kopie | Schleife über Header (`Authorization` in allen 8192 Groß-/Kleinschreibungen als Schlüssel der Map, `X-A`, ein Header mit zwei Werten, dazu 48 Nachbarnamen neben `Authorization`: je Zeichen gestrichen, `x` eingefügt oder ersetzt, `Proxy-Authorization`, `X-Authorization`, `Authorization-Info`, `Auth`, `AuthX`, `Auth-Info`, `Authentication`, `WWW-Authenticate`) mit einem Canary; die gehaltene `Anfrage` hat nur die Nicht-`Authorization`-Header; die Route überschreibt einen Wert und löscht einen Namen ihrer Kopie: die Header des Requests bleiben gleich | Header ohne Filter kopiert; Filter nur für `Authorization` in kanonischer Schreibweise; die Map des Requests ohne Kopie weitergereicht | andere |
| Y2 | (`SPEC-240`) `ODR-E2102` trägt 400, 404 und 405, den Zielbereich `admin`, den Text `Admin request invalid`, steht hinter `ODR-E2101` in `AlleCodes()` | `fehler_test.go`: Schleife über Status 0 bis 999, `StatusErlaubt` genau für 400, 404, 405; `NeuesFehlerobjekt`; Literal-Tabelle um die Zeile ergänzt, Zählprüfung gegen `AlleCodes()` | Status 405 aus `Stati` entfernt | andere |
| Y3 | (Nr. 14) `ODR-E2102` steht mit Grund in `harness/katalog-ausnahmen.txt`, das Handbuch bleibt unverändert | benannte Ausgabe: `make katalog-check` grün (rot ab dem Code in `fehler.go`, bis die Zeile steht) | Zeile nicht eingetragen | andere |
| Y4 | (`SPEC-230` Nr. 8) Der Zähler führt `ODR-E2102`: ein Eintrag je Code der Tabelle, `CountError` erhöht ihn | `regeln_test.go` und `telemetry_test.go`: Schleife über `AlleCodes()`; die Literal-Liste der Codes um `ODR-E2102` ergänzt | die Liste nicht ergänzt (Test rot) | andere |

Zeilen gesamt 13: *a* = 0, *b* = 13. Die Zeile Z2 ging in Z1 auf; K4 (Kopie des Tokens) hängt an der Authentisierung und liegt in `slice-admin-authentisierung`, K5 (Server teilen nichts, `Anhalten`) in `slice-admin-lebenszyklus`. Die Tests dieses Teils senden einen gültigen `Authorization`-Header, den der Teil nicht liest; „mit Token“ in Z5 und Z6 heißt das.

**Aufträge an den Architect zum Start** (Start, Planner, 2026-11-07, `main` = `4d441211`; vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Architect arbeitet im Zweig `slice-admin-rahmen`; Ausgang jeweils *offen* in §6). Der Vertrag (`SPEC-238` bis `SPEC-240`) und die Tabellen dieses Abschnitts stammen von ihm; wo etwas vollständig ist, genügt die Bestätigung mit Nummer. Die Aufträge heißen S1 bis S9, damit sie nicht mit den Buchstaben „Auftrag C bis F“ der Architect-Tabellen unten verwechselt werden.

- **S1. Sätze und Zeilen (`AGENTS.md` §3.15).** Jeder Satz von `SPEC-238` Nr. 1, 2, 3 a, 5, 6, 14 und der Zeile `SPEC-240`, auch jeder Nebensatz (Token „ab 32 Bytes“, „kein Bereinigen des Pfades“, `Allow` „sortiert, die eingebaute Route eingeschlossen“, `Pfadwerte` „nie nil“, `Laenge` „-1“), hat eine Zeile K1 bis K3, Z1, Z3 bis Z8, Y2 bis Y4 oder ein benanntes Negativ mit Grund und Folge; die Spalten sind die, die `SPEC-128` Nr. 3 liest. Jedes Test-Symbol der Tabelle existiert vor dem Handoff (Register `tabelle-nennt-test-der-nicht-existiert`).
- **S2. Zählprüfung der Zahl der Codes (Y2, Y4).** Der 40. Code ist `ODR-E2102`: `model.AlleCodes()` auf `main` hat 39 Einträge (`ODR-E6008`, `ODR-E6009` stehen darin). Die Tests des Modells, der Telemetrie (`regeln_test.go`) und des Audit (`audit_test.go`) zählen gegen `model.AlleCodes()`, **nie gegen eine feste Zahl im Produktcode**; eine Literal-Liste im Test hält die Zahl gegen die Quelle (`TestFehlercodeForm`), und der Architect nennt jede Stelle, die „39“ als Literal trägt (Suche über `src/` und `tests/`). `SPEC-230` Nr. 8 steht auf 40; der Architect bestätigt es.
- **S3. Katalog und Handbuch (`SPEC-201` Nr. 3 c, e; `SPEC-238` Nr. 14).** `ODR-E2102` steht im selben Commit wie `fehler.go` in `harness/katalog-ausnahmen.txt` (Zeile `<Code> | <Grund>`, direkt unter `ODR-E2101`); das Handbuch bleibt unverändert, weil das Binary den Listener nicht verdrahtet. Der Architect bestätigt diese Entscheidung gegen die Handbuch-Pflicht der DoD und gegen `make handbuch-check` und `make katalog-check`; er nennt, ab welchem Slice der Code das Handbuch betritt (`slice-bootstrap-m6`) und trägt die Adresse in §6 ein.
- **S4. Nummernlage der Linie B (`SPEC-238` bis `SPEC-240` vergeben).** Dieser Teil braucht keine neue `SPEC`-Nummer; die Sätze stehen in `SPEC-238`. Braucht der Architect doch eine, geht die Frage als Rückgabe an den Planner, bevor er sie schreibt (Bereich der Linie ist erschöpft, die Kennungsfrage liegt beim User, `slice-snapshot-use-case` §6).
- **S5. Raster und Zusatztabelle (§6, `AGENTS.md` §3.16).** Die Zellen von Konfiguration und Zuordnung tragen je eine Nummer von `SPEC-238` oder „gilt nicht“ mit Grund; der Architect bestätigt, dass `SPEC-238` Nr. 2 die Randformen der Zuordnung (`.`, `..`, `%2F`, UTF-8, leeres Segment, abschließender `/`, Groß- und Kleinschreibung) mit Wortlaut trägt, nicht nur als Beispiel.
- **S6. Pflicht-Mengen mit Zählprüfung, Zweige und Runden, zählende Fakes, Bindung (§3, Register `menge-im-plan-weicht-vom-orakel-des-tests-ab`, `unbeschraenkte-menge-ohne-benannte-pruefgrenze`, `zweig-ohne-lauf-unter-dem-detektor`, `hilfe-ersetzt-argument-verdeckt-eigenschaften`, `panik-eines-fakes-verdeckt-den-benannten-test`, `dienst-erfuellt-den-port-ohne-bindung`).** Der Architect bestätigt die Mengen M-Route und M-Methode × Pfad samt Achsen und Zahlen (das Orakel trägt Statuscodes, Header und Texte als Literale aus dem Vertrag), die Prüfgrenzen (64 Routen, 4 Parameter), die Zweig-Tabelle mit Zählprüfung (null Blöcke mit Zähler 0 unter `-race -coverprofile`), die Aufrufstellen mit ihren zählenden Fakes und die Zusicherungen in `bindung_test.go`; fehlt eine, ergänzt er sie.
- **S7. Merge-Nähe (Dateien mit anderen Zweigen).** `slice-snapshot-use-case` und `slice-sqlite-interrupt-transaktion` schreiben ebenfalls in `spec/spezifikation.md`, `slice-snapshot-use-case` meldet zudem, ob ein Code entsteht (der nächste wäre der 41.). Der Architect schreibt in `spec/spezifikation.md` nur die Stellen, die dieser Teil berührt (`SPEC-238`, Zeile `SPEC-240`, Nachzug `SPEC-109`, `SPEC-230` Nr. 8), und stellt keinen Absatz daneben um; wer als zweiter nach `main` merged, zieht `main` nach und gleicht die Zahl der Codes an (§6).
- **S8. Größe und Pfade (`AGENTS.md` §3.14, `SPEC-128` Nr. 1).** Der Architect bestätigt *a* = 0, *b* = 13, *f* = 3, *p* = 110 und die Pfade des Blocks (§4); er setzt weder `basis` noch `schaetzung` noch `marke`, den Block setzt der Planner nach der Bestätigung in eigenem Commit vor dem ersten Code-Commit. Pfade über die Liste hinaus meldet er dem Planner. Die Auslastung der letzten Teile lag bei 0,41 bis 0,71 (Register `schaetzformel-ueberschaetzt-den-slice`); er benennt das, ohne die Formel zu ändern.
- **S9. Reports der Rollen und Betriebsregeln.** Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link (Register `report-der-rolle-faellt-docs-check`); der Planner reicht den Satz an beide weiter. Der Implementer schreibt Repo-Dateien nur mit `Edit` und `Write` (Register `betriebsregel-fuer-repo-dateien-verletzt`, 12×).

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, Auftrag C; das Orakel trägt Statuscodes, Codes, Header und Meldungstexte als Literale aus dem Vertrag, nie als Konstanten des Codes):

| Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|---|
| M-Route (Z1) | Achsen: ungültige Methode 5 · feste ungültige Muster 12 · Bytewert im Literal 256 (66 gültig) · erstes Zeichen des Namens 256 (26 gültig) · weiteres Zeichen des Namens 256 (27 gültig) · nil-`Handler` 1 | Generator, Modell des Tests je Segment | `ErrUngueltig` oder gültig | Summe der Elemente der Achsen (786), gültige 66 + 26 + 27 gegen die Zeichenmengen des Vertrags; die Zahl 256 ist die Länge der Schleife über `byte` |
| M-Überlappung (Z3) | Muster 120 (Folgen von 1 bis 4 Segmenten über {`a`, `b`, Parameter}: 3 + 9 + 27 + 81) → 7260 Paare (n(n+1)/2) × Methode gleich oder verschieden = 14520, dazu die 4 Fälle der eingebauten Route; Prüfgrenze: vier Segmente (der Code vergleicht jede Stelle gleich, eine fünfte Stelle brächte keine neue Art) | Generator, Modell des Tests je Segment | `ErrUngueltig` oder gültig | Paare n(n+1)/2 × 2 = 14520 aus n = 120, dazu 4 |
| M-Zuordnung (Z4, Z5) | Z4: Pfadformen (Literal gleich, Schreibweise, `%61`, `%41`, Segmentzahl ±1: 6 Formen) × 3 Muster, dazu je Parameterstelle (0, 1 und 2 Stellen) 19 Werte (7 Formen aus Punkten, Prozentfolgen und Bytes, 12 Formen, in denen Pfad- und Query-Dekodierung sich unterscheiden) und das leere Segment (404): 18 + 3 × 20 = 78; Z5: 8 Pfade × 2 Methoden = 16 | Schleife | Treffer oder 404 und `Pfadwerte` als Literale des Tests | Länge der Listen im Test: 78, Z5: 16 |
| M-Methode × Pfad (Z6) | Methoden 8 × Pfade 4 = 32, dazu 7 Zusatzfälle (`get` auf `/b`, `get` auf `/a/b`, `get` und `Get` auf `/healthz`; leere Liste der Routen: 3) | Schleife | Status, `Allow`, Body und Header als Literale | 32 + 7 |
| M-Grenze (Z4) | 64 Routen `GET /r<i>/{a}/{b}/{c}/{d}` (i von 0 bis 63, 4 Parameter) | Schleife | die Anfrage `/r<i>/1/2/3/4` trifft genau die Route i, `routeFake` hält `i` und die vier Werte | 64 gegen die Länge der Liste |

**Unbeschränkte Mengen und ihre Prüfgrenze (akzeptierte Negative):** die Zahl der Routen (Prüfgrenze 64; Grund: die Zuordnung ist linear; Folge: bei mehr Routen wächst die Laufzeit je Request linear, kein Fehler); die Zahl der Parameter je Muster (Prüfgrenze 4; Grund: jeder Parameter ist ein Segment; Folge: ein Muster mit mehr Parametern gilt wie eines mit vier, ohne eigenen Fänger).

**Zweige als Menge, Runden unter dem Detektor** (Architect, Auftrag D; Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat null Blöcke mit Zähler 0, der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Z-1 `New`: nil-Feld (vier), Token unter 32, ungültige Route (Methode, Muster, Handler), Überlappung, Kopie der Routen | K1, K2, Z1, Z3, Z4 | ein Lauf |
| Z-2 Zuordnung: Pfad ohne führenden `/`, Treffer, 404, 405, eingebaute Route `GET /healthz` (Handler gibt 200 mit dem Literal von `SPEC-238` Nr. 3 c; die Zusage „ohne Token“ fängt `slice-admin-authentisierung`, A3, A4) | Z4 bis Z6 | ein Lauf |
| Z-3 Aufbau der `Anfrage`: Filter von `Authorization`, `Laenge` bekannt und `-1`, `Pfadwerte` mit und ohne Parameter | Z7, Z8 | ein Lauf |
| Z-4 Schreiben: Fehlerobjekt für 404, 405 und vorläufig 500, Erfolg mit und ohne Body | Z4, Z6 | ein Lauf |
| Z-5 Dekodieren eines Segments scheitert (`SPEC-238` Nr. 2: tritt nicht auf) | kein Fänger (N-4) | – ; der einzige erlaubte Ausreißer der Zählprüfung, mit Datei und Funktion im Bericht |

Nebenläufigkeit: gilt nicht in diesem Teil; die Zusage „gleichzeitige Requests teilen nichts“ (`SPEC-238` Nr. 12) fängt N1 in `slice-admin-abweisen-audit`, hier ist der Zustand nach `New` schreibgeschützt (Z4: Kopie der Routen).

**Aufrufstellen und zählende Fakes** (Architect, Auftrag E; „jeder Fall je Aufrufstelle“ gilt für diese Stellen):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| genau eine, vor allem anderen | `IDs.KorrelationsID` | Z7 (Feld `KorrelationsID` der `Anfrage`); A1 und E4 in `slice-admin-authentisierung` | `zaehlendeIDs` (Hülle um `FestIDs`, test-lokal): zählt alle fünf Operationen, hält die vergebene Kennung |
| `Accept`, `Close` und `Addr` des Listeners | `net.Listener` | K3 (alle 0 in diesem Teil; `Serve` und `Anhalten` rufen sie in `slice-admin-lebenszyklus`) | `listenerFake`: zählt `Accept`, `Close` und `Addr`, Fehler des Accept einstellbar, keine Panik |
| genau eine je Route | `Handler` | Z4, Z7, Z8 | `routeFake`: zählt, hält `Anfrage` und `ctx`, liefert Antwort und Fehler einstellbar |
| keine in diesem Teil | `Audit.Append`, `Telemetrie.CountError` | K3 (Zähler 0) | `FakeAudit`, `FakeTelemetry` zählen jede Operation |

Die Fakes sind test-lokal (*f* = 3: `zaehlendeIDs`, `routeFake`, `listenerFake`); `FakeAudit` und `FakeTelemetry` bestehen. Der Test-Server mit Client (echter Listener) gehört zu `slice-admin-lebenszyklus`.

**Bindung an den Port** (Architect, Auftrag F; Lerneintrag V-2): Die Zusicherungen zur Übersetzungszeit stehen in `src/adapters/driving/admin/bindung_test.go` (Paket `admin_test`): `var _ support.AuditLog = (*supporttest.FakeAudit)(nil)`, `var _ support.Telemetry = (*supporttest.FakeTelemetry)(nil)`, `var _ support.IDGenerator = (*zaehlendeIDs)(nil)`, `var _ net.Listener = (*listenerFake)(nil)`, `var _ admin.Handler = (*routeFake)(nil).handle` (der Methodenwert; `(*routeFake).handle` wäre ein Methodenausdruck mit dem Empfänger als erstem Argument und passte nicht), `var _ http.Handler = (*admin.Server)(nil)`; `Konfig` trägt die Felder als Schnittstellentypen des Ports (nicht als Funktionen), so übersetzt eine Mutation einer Signatur das Paket nicht mehr. Der Adapter ruft die Inbound Ports nicht: er kennt nur `Handler`.

**Mutationen und Raten** (Implementer, 2026-11-07; `AGENTS.md` §3.10, §3.15, §3.17). Jede Mutation lief auf einer frischen Kopie der Quellen im Container (netzlos, `--memory=4g --memory-swap=4g --cpus=4`, `go test -count=1 -run <Test>`), ein Prozess je Mutation; Messender: der Implementer; Last des Hosts: `load average` 55 bis 69 bei 4 Rechenkernen je Container. Alle Fänger sind deterministisch (kein Zeit-, Zufalls- oder Nebenläufigkeitsanteil; nur `TestNewRuftNichts` misst Goroutinen und fängt die blockierende): rot in 1 von 1 Prozessläufen je Mutation, eine Einzelzahl und nur eine Größenordnung. Erste Runde: 110 Läufe von Mutationen, 106 rot und vier grün (Summe der Aufzählung; die Fassung „108 Läufe“ widersprach ihr). Vier blieben grün: zweimal dieselbe Mutation „Parameter nimmt das leere Segment“ unter `TestZuordnung` (Folge: das leere Segment als Wert ergänzt, §6), „nur das letzte Segment verglichen“ (Folge: Muster `/b/c` ergänzt, §6), beide danach rot gesehen, und „Goroutine endet vor der Messung“ unter `TestNewRuftNichts`, das Negativ N-7 (der Test „keine go-Anweisung in `New`“ fängt den Rumpf von `New`).

| Zeile | Mutation (Produktcode) | roter Test |
|---|---|---|
| K1 | `case k.Listener == nil:` zu `case false:` (ebenso `IDs`, `Audit`, `Telemetrie`: 4 von 4) | `TestNewNilFelder` |
| K2 | `< tokenMinBytes` zu `<=` und zu `< tokenMinBytes-1`; das Token im Text | `TestNewTokenLaenge` |
| K3 | in `New`: `IDs.KorrelationsID`, `Audit.Append`, `Telemetrie.CountError`, `Listener.Addr`, `.Close`, `.Accept` (6 von 6); `go func() { select {} }()`; `net.Listen("tcp", ":8081")`; `go func() {}()` | `TestNewRuftNichts` (7 von 7 der ersten Art); `TestKeinListenImPaket`; `TestNewStartetKeineGoroutine` (die letzte Mutation lässt `TestNewRuftNichts` grün, N-7) |
| Z1 | Name `[a-z]` zu `[A-Za-z]`; `_` im Rest entfernt; Ziffer im Rest erlaubt; doppelte Namen nicht geprüft; `~` entfernt; `:` ergänzt; `_` im Literal entfernt; nil-`Handler` nicht geprüft; `HEAD` ergänzt; `PATCH` entfernt; leeres Segment gültig; führender `/` nicht geprüft; Muster oder Methode im Text des Fehlers | `TestRoutenGrammatik` (`PATCH`: `TestRoutenGutfaelle`; Text: `TestRoutenTextOhneWert`) |
| Z1, nicht genannte Elemente | `-`, `.`, `9` und `Z` aus dem Literal entfernt; `z` aus dem ersten und aus dem weiteren Zeichen des Namens entfernt | `TestRoutenGrammatik` |
| Z3 | Segmentzahl nicht verglichen; Methode nicht verglichen; eingebaute Route nicht gezählt; Literale nicht verglichen; Parameter nur auf einer Seite; nur erstes Segment; nur letztes Segment (zuerst grün, `/b/c`); Muster im Text | `TestRoutenUeberlappung`, `TestEingebauteRouteUeberlappt`, `TestRoutenTextOhneWert` |
| Z4 | Vergleich ohne Schreibweise; Literal gegen das rohe Segment; Parameter nimmt das leere Segment (zuerst grün, leeres Segment als Wert); Segmentzahl `<`; Status der Route nicht geschrieben; Body nie geschrieben; Fehler der Route als `ODR-E2102`; die Liste der Routen nicht kopiert (Handler aus der Liste des Aufrufers gelesen) | `TestZuordnung`, `TestAntwortDerRoute`, `TestZuordnungKopiertRouten` |
| Z4, nicht genannte Elemente | Wert `\xFF` als Nicht-UTF-8 abgewiesen; Wert `..` abgewiesen | `TestZuordnung` |
| Z3, Fix-Runde (Befund 1; Menge: jede Stelle 1 bis 4 bei jeder Länge 1 bis 4, 120 Muster, 14520 Paarfälle) | nur die ersten zwei Segmente verglichen; dritte Stelle übersprungen (nicht im Befund); vierte Stelle übersprungen (nicht im Befund); mittleres Segment (Index 1) übersprungen; `\|\|` statt `&&` je Seite | `TestRoutenUeberlappung` (die letzte auch `TestEingebauteRouteUeberlappt`) |
| Z4, Fix-Runde (Befund 2; Menge: 19 Werte × 3 Parameterstellen, 78 Fälle) | `url.QueryUnescape` statt `PathUnescape` (rot am Wert `a+b` und `+`); `+` vor dem Dekodieren zu `%20` (nicht im Befund); Wert mit `;` nach dem Dekodieren abgewiesen (nicht im Befund) | `TestZuordnung` |
| Z8, Fix-Runde (Befund 3; Menge: 8192 Schreibweisen, 48 Nachbarnamen, 3 feste Fälle = 8243) | Filter auf Suffix `authorization` (rot an `X-Authorization`); Präfix `auth` (rot an `Authentication`); Teilfolge `orizat` (rot an `Authorization-Info`); nur drei Schreibweisen (rot an `aUTHORIZATION`, nicht im Befund); Filter nach Länge 13 (rot an `authorizatiox`, nicht im Befund) | `TestHeaderOhneAuthorization` |
| Z5 | `//` zu `/` ersetzt; `/./` zu `/` ersetzt; abschließender `/` entfernt; Prüfung des führenden `/` entfernt (`POST *` trifft `/{x}`) | `TestKeinBereinigen` |
| Z6 | `Allow` unsortiert, absteigend sortiert, nur vom ersten Muster, ohne die eingebaute Route, mit `,` statt `, `; 404 und 405 vertauscht (beide Richtungen); Methode ohne Beachtung der Schreibweise (zuerst nur mit `get` auf `/b` grün, §6); `Content-Length` fehlt; `Content-Type` anders; Body der eingebauten Route anders; Wert aus dem Request im Body | `TestMethodeUndPfad`, `TestMethodeSchreibweise` |
| Z7 | je Feld der `Anfrage` nicht gesetzt (`Methode`, `Pfad`, `RawQuery`, `KorrelationsID`, `Header`, `Body`, `Laenge`, `Pfadwerte`); `Pfad` kodiert; `RawQuery` neu kodiert; `Laenge` nie `-1`; `Pfadwerte` nil ohne Parameter; `KorrelationsID` ein zweites Mal, nach allem, gar nicht; `IDs.WorkspaceID` zusätzlich; `ctx` nicht der des Requests | `TestAnfrageFelder`, `TestPfadwerteNieNil`, `TestKorrelationsID`, `TestCtxDerRoute` |
| Z8 | Header ohne Filter; Filter nur für `Authorization`; Filter für `Authorization` und `authorization`, nicht `AUTHORIZATION`; Map des Requests weitergereicht; Slices nicht geklont; nur der erste Wert eines Headers kopiert | `TestHeaderOhneAuthorization`, `TestHeaderKopie`, `TestAnfrageFelder` |
| Y2 | Status 405, 400 oder 404 aus `Stati` entfernt, 406 ergänzt; Text mit Punkt; Ziel `internal`; Zeile hinter `ODR-E2401` | `TestCodeTabelle` |
| Y4 | Zeile `ODR-E2102` aus der Tabelle entfernt; die Liste in `regeln_test.go`, `audit_test.go`, `lesen_fehler_test.go` je ohne `ODR-E2102`; `+ 2` zu `+ 1` in `httpodata/fehler_test.go` | `TestCodeTabelle`, `TestErrorCode`, `TestFehlercodeForm`, `TestRueckgabenMenge`, `TestFehlerobjektJeCode` (alle fünf bei entfernter Zeile; je Liste der eigene Test) |
| Y3 | Zeile in `harness/katalog-ausnahmen.txt` fehlt | `make katalog-check` rot, gesehen: `katalog: Fehlercodes: ODR-E2102 fehlt im Handbuch` |

**Fix-Runde nach dem Review** (Implementer, 2026-11-07; `AGENTS.md` §3.17): 13 Mutationen in einer Wegwerfkopie (`odata-recorder:test`, ohne Netz, `--memory 3g --memory-swap 3g --cpus 4`, unter `flock`, `go test -count=1` über das Paket), ein Prozess je Mutation, Messender: der Implementer; Last des Hosts: nicht gemessen (die Läufe warteten auf die Sperre, 10 bis 22 s je Lauf); deterministische Fänger, rot in 1 von 1 Läufen, eine Einzelzahl und nur eine Größenordnung; 13 rot, 0 grün. Die Befunde 1 bis 3 sind die Klassen „Stelle × Länge“ (Z3), „Wertformen, in denen zwei Dekodierer sich unterscheiden“ (Z4) und „Nachbarnamen und Schreibweisen“ (Z8); die Mengen und Zahlen stehen in den drei Zeilen oben. Das Feld `treffer.gefund` (Befund 5) ist entfernt: `ServeHTTP` entscheidet nach `route` und `allow`. Rohbytes: `route_test.go`, `zuordnung_test.go` und `anfrage_test.go` sind mit `LC_ALL=C grep -P` auf BOM, NBSP, Nullbreitenzeichen und Backslash-u-Folgen gegengelesen.

**Klassen mit Generator** (`implement-slice` Schritt 14): *M-Route* (Z1) Generator `z1Generator`, Orakel die Zeichenmengen des Vertrags als Literale (66, 26, 27), 786 Fälle als Summe der Achsen (5 + 12 + 3 × 256 + 1), davon 119 gültig (66 + 26 + 27), jeder der 256 Bytewerte je Position; *M-Überlappung* (Z3) Generator `z3Muster` über 7260 Paare × 2 = 14520 Fälle aus 120 Mustern (ein bis vier Segmente), Orakel `modellUeberlappt` je Segment, dazu 4 Fälle der eingebauten Route (Prüfgrenze vier Segmente, Folge: eine Abweichung erst ab der fünften Stelle bliebe unbemerkt, der Code vergleicht jede Stelle in derselben Schleife); *M-Zuordnung* (Z4) `z4Generator`, 78 Fälle (6 Formen × 3 Muster + 3 Parameterstellen × 20 Werte), Orakel Literale der Pfade und der dekodierten Werte (aus Bytewerten gebaut), Z5 16 Fälle; *M-Header* (Z8) `z8Schreibweisen` (2^13 = 8192) und `z8Nachbarn` (48), 8243 Fälle samt den drei festen; *M-Methode × Pfad* (Z6) 32 + 7 Fälle mit Literalen für Status, `Allow`, Body und Header; *M-Grenze* 64 Routen mit je 4 Parametern. Alle Zählprüfungen rechnen die Summe der Achsen aus den Listen im Test und vergleichen sie mit der Zahl der gelaufenen Fälle; Zahlen der Codes zählen die Tests gegen `model.AlleCodes()` (kein festes 39 oder 40).

**Zweige unter dem Detektor** (Z-Liste): `go test -race -coverprofile` (Image `odata-recorder:race`, Go 1.27.1) über das Paket `admin` zählt in `admin.go` und `route.go` einen Block mit Zähler 0: `route.go:128.4,129.1`, die Rückgabe bei einem Segment, das sich nicht dekodieren lässt (Z-5, N-4, der eine erlaubte Ausreißer). Die Blöcke `tokendatei.go:108.3,109.1` und `tokendatei.go:132.3,133.1` gehören zu Teil 1 und stehen nicht in der Liste dieses Teils. Rohbytes: `admin.go`, `route.go` und alle Test-Dateien des Pakets sind mit `LC_ALL=C grep -P` auf BOM, NBSP, Nullbreitenzeichen und Backslash-u-Folgen gegengelesen (0 Treffer); die Werte jenseits von ASCII (`%C3%A4` dekodiert, `%FF`, `%00`) entstehen in `z4Werte` aus `[]byte{…}`.

**Größe der Sicht des Architect** (`SPEC-128` Nr. 1, Aufteilung durch den Planner): 13 Zeilen (*a* = 0, *b* = 13), *f* = 3, *p* = 110 (Rahmen etwa 90, Code und Katalog etwa 20; das Register `schaetzformel-ueberschaetzt-den-slice` sagt: Produktcode eher darunter, Tests darüber).

**Bestätigungen und Ergänzungen des Architect zu S1 bis S9** (Architect, 2026-11-07, im Zweig `slice-admin-rahmen` vor dem ersten Code-Commit; die Ergänzungen stehen im Wortlaut in den Zeilen K3, Z1, Z3 bis Z8, in den Mengen, den Zweigen und der Bindung oben; die Tabelle bleibt bei 13 Zeilen). Der Vertrag bekam zwei Sätze nach: `SPEC-238` Nr. 1 („kopiert das `Token` und die `Routen`“) und Nr. 2 (Pfad beginnt mit `/`, `%61` auf das Literal `a`, Wert aus UTF-8 und Nicht-UTF-8, Dekodieren scheitert nie, Methode des Requests bytegleich, der Treffer ist eindeutig); keine neue Nummer, keine Zeile im Änderungsverzeichnis (der Eintrag vom 2026-11-07 trägt den Vertrag, und eine Zeile mehr am Ende der Tabelle stieße in den Zweigen von `slice-snapshot-use-case` und `slice-sqlite-interrupt-transaktion` zusammen).

**S1. Abdeckung der Sätze** (jeder Satz von `SPEC-238` Nr. 1, 2, 3 a, 5, 6, 14 und der Zeile `SPEC-240`, auch jeder Nebensatz):

| Satz im Vertrag | Zeile |
|---|---|
| Nr. 1: `Listener`, `IDs`, `Audit`, `Telemetrie` nil | K1 |
| Nr. 1: `Token` unter 32 Bytes (32 gültig) | K2 |
| Nr. 1: eine Route nach Nr. 2; der Text nennt keinen Wert, nie das Token | Z1, Z3 (Canary in Methode und Muster); K1, K2 |
| Nr. 1: `New` kopiert das `Token` | K4 (`slice-admin-authentisierung`) |
| Nr. 1: `New` kopiert die `Routen` | Z4 |
| Nr. 1: kein Port, keine Methode des `Listener`, kein Socket, keine Goroutine | K3 |
| Nr. 1: `Vergleich` nil heißt `vergleiche` | U4, U5 (`slice-admin-authentisierung`) |
| Nr. 1: keine Standardadresse, bedient den übergebenen `Listener`, zwei Server teilen nichts | K3 (kein Aufruf von `Listen`), K5 (`slice-admin-lebenszyklus`, darf den AST-Test von K3 mitnehmen) |
| Nr. 1: `Konfig` wächst additiv | N-1 |
| Nr. 2: `Routen` darf leer sein, dann nur `GET /healthz` | Z1 (leere Liste gültig), Z6 (Zusatzfall) |
| Nr. 2 a, b, c | Z1 |
| Nr. 2 d (Überlappung, eingebaute Route) | Z3 |
| Nr. 2 Zuordnung: Beginn mit `/`, Zerlegung, Dekodieren, Literal, Parameter, Segmentzahl | Z4, Z5 |
| Nr. 2: kein Bereinigen, leeres Segment, abschließender `/` | Z5 |
| Nr. 2: `%2F` im Wert eines Parameters | Z4, Z7 |
| Nr. 2: Dekodieren scheitert nie | N-4 |
| Nr. 2: 404, 405, der `Handler` der einen Route | Z4, Z6 |
| Nr. 2 `Anfrage`: acht Felder, `Pfadwerte` nie nil, `Laenge` -1 | Z7 |
| Nr. 2: `Header` ohne `Authorization`, als Kopie | Z8 |
| Nr. 2: `Body` begrenzt (Nr. 10) | G3 (`slice-admin-lebenszyklus`), N-5 |
| Nr. 2 `Antwort`: Status 200 bis 299, Body | Z4 (Durchreichen), B9 (`slice-admin-fehlerabbildung`), E3 (`slice-admin-authentisierung`) |
| Nr. 3 a: `KorrelationsID` genau einmal, vor allem anderen, sonst keine Operation von `IDs` | Z7, A1 (`slice-admin-authentisierung`) |
| Nr. 5: Fehlerobjekt von 404 und 405 (drei Felder, `Content-Type`, `Content-Length`, kein Wert), `Allow` | Z6 |
| Nr. 5: `X-Correlation-Id` an jeder Antwort | E4 (`slice-admin-authentisierung`); bis dahin trägt die Antwort sie nicht (N-6) |
| Nr. 6: der Code `ODR-E2102` | Y2 |
| Nr. 14: Zeile in den Ausnahmen; kein Schlüssel, kein Kommando | Y3 |
| Nr. 14: das Binary bleibt unverändert | N-3 |
| `SPEC-240`: 400, 404, 405, `admin`, Text, hinter `ODR-E2101` | Y2 |
| `SPEC-240`: kein Zustandswechsel | Z5, Z6 (die Route wird nicht gerufen) |

Akzeptierte Negative (Grund und Folge):

- N-1 `Konfig` wächst additiv (Nr. 1): kein Fänger. Grund: die Aussage gilt künftigen Änderungen. Folge: eine Umbenennung bricht Aufrufer beim Übersetzen (`bindung_test.go`, Composition Root), nie still.
- N-2 Dieser Teil zählt und auditiert nicht (Nr. 7, 8): kein Test „Zähler 0 je Request“. Grund: `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` müssten ihn umschreiben (C1, C2). Folge: ein vorzeitiges `CountError` bliebe bis C1 grün; K3 fängt nur den Aufruf in `New`.
- N-3 Das Binary bleibt unverändert (Nr. 14): kein Fänger. Grund: `src/bootstrap` steht nicht im Block, kein Test liest den Diff. Folge: das Review liest `git diff --stat`; die Verdrahtung kommt mit `slice-bootstrap-m6` und eigenem Fänger.
- N-4 Ein Segment, das sich nicht dekodieren lässt (Nr. 2): kein Fänger. Grund: `EscapedPath` liefert nur gültige Kodierungen, ein Request erreicht den Zweig nicht. Folge: der Zweig (Segment passt nicht) ist unerreichbar und der eine erlaubte Ausreißer der Zählprüfung (Z-5).
- N-5 Der `Body` ist in diesem Teil nicht begrenzt (Nr. 10): kein Fänger. Grund: die Grenze von 1 MiB gehört zum `http.Server` (G3). Folge: bis dahin erreicht kein Request den Adapter, das Binary bleibt unverändert.
- N-6 Die Antworten dieses Teils tragen weder `X-Correlation-Id` noch bei Erfolg `Content-Type` und `Content-Length` (E3, E4 in `slice-admin-authentisierung`): die Tests prüfen ihr Fehlen nicht. Folge: wie N-5.

Akzeptierte Negative aus der Verifikation (Planner, 2026-11-08, Auflage A1 des Verifiers; Entscheidung: Durchstich vor Härtung, kein zweiter Fix-Lauf; die Mutationen sind im Report, Abschnitt 4, mit ihrem Ergebnis belegt):

- N-8 Z4, Literal gegen längere Segmente (Verifier-Befund 2): `HasPrefix`, `HasSuffix` und `Contains` statt Gleichheit bleiben grün, weil das Literal der Muster ein Zeichen ist und kein Fall ein Segment einsetzt, das es enthält (`ab`, `xa` gegen `a`). Grund: das Muster der Zuordnung ist ein Vergleich über Segmentzahl und Segment, der Code vergleicht mit `==`; Durchstich vor Härtung, kein Produktfehler heute. Folge: eine Änderung des Vergleichs auf eine Teilfolge ließe `GET /ab` die Route `/a` treffen, ohne dass ein Test rot wird; `passt` in `route.go` ist Prüfpunkt jedes Reviews, das diese Datei ändert.
- N-9 Z4, zweimaliges Dekodieren (Verifier-Befund 3): ein zweites `PathUnescape` je Segment (`%2541` ergäbe `A`) bleibt grün, weil kein Wert der Menge nach dem ersten Dekodieren wieder die Form `%XX` trägt. Grund: wie N-8. Folge: ein Pfadwert, der ein Name wird, könnte aus `%252e%252e` den Wert `..` bilden; der Code dekodiert heute genau einmal (`zerlegePfad`, vom Verifier gelesen), nichts erreicht den Listener (N-3); eine Änderung an `zerlegePfad` hat diese Zeile als Prüfpunkt im Review. Der Planner sichtet die Zeile beim Start der Endpunkt-Slices (Sichtung nach `plan-welle`, keine Sendung).
- N-10 Z8, mehrzeilige und leere Nachbar-Werte (Verifier-Befund 1, Rest): `Authorization` mit zwei Zeilen wird von der Mutation „nur einzeilige gefiltert“ nicht gefangen — diesen Teil des Befundes trägt `slice-admin-authentisierung` (Zeile A7, Sendung von diesem Slice, dort in §1 und DoD eingetragen); Nachbarn mit leerem erstem Wert, mit leerer Liste und Namen ab 20 Zeichen: kein Fänger. Grund: der Filter prüft den Namen, nie den Wert (`anfrage` in `admin.go`). Folge: eine Änderung, die Header nach ihrem Wert verwirft, bliebe grün; bis A7 gilt: in diesem Teil erreicht kein Request den Adapter (N-3), das Binary bleibt unverändert.
- N-11 Prüfgrenzen der unbeschränkten Mengen Z4 und Z8 (Verifier-Befund 4): Länge und Zeichen eines Parameterwertes (geprüft bis 5 Bytes; Mutationen „nur bis 8 oder 64 Bytes“ und „ASCII-Großbuchstaben gewandelt“ bleiben grün) sowie Name und Wert eines Headers (Namen bis 19 Zeichen, Wert `1`) haben keine obere Grenze im Test. Grund: der Code kennt keine Grenze (`parseSegment`, `anfrage`), die Menge ist unbeschränkt. Folge: eine künftige Grenze für Wertlängen oder eine Umwandlung der Schreibweise im Wert bliebe grün; die Begrenzung der Request-Größe liegt beim `http.Server` (`slice-admin-lebenszyklus`, G3).

**Last des Hosts der Fix-Runde (Verifier, Auflage A4):** nicht gemessen vom Implementer; die Fänger sind deterministisch (kein Zeit-, kein Nebenläufigkeitsanteil), eine Einzelzahl je Mutation ist Größenordnung (`AGENTS.md` §3.10); der Verifier maß dieselben Läufe bei Load Average 4 bis 7 auf 4 CPUs je Container.

**S2. Zahl der Codes.** `model.AlleCodes()` hat auf `main` 39 Einträge; `ODR-E2102` ist der 40. `SPEC-230` Nr. 8 steht auf 40 (bestätigt, `spec/spezifikation.md`). Kein Produktcode zählt die Codes. Die Stellen im Test, die die Zahl oder die Liste tragen (Suche über `src/` und `tests/` nach `AlleCodes`, `39` und den Listen der Codes; `tests/` führt keine):

- `src/hexagon/model/fehler_test.go`: Literal-Tabelle von `TestCodeTabelle` (Zeile `ODR-E2102`: Stati 400, 404, 405, Ziel `admin`), die Zahl 39 in Bedingung und Meldung (Zeilen 81 und 82), eine Prüfung „hinter `ODR-E2101`“ mit Text `Admin request invalid` und Status 0 bis 999 (nur 400, 404, 405 erlaubt).
- `src/adapters/driven/audit/audit_test.go`: die Liste `codeTexte` (um `ODR-E2102`), die Literale 39 in den Zeilen 381 (Kommentar), 389 (Kommentar), 394 bis 395, 406 bis 407, 426 bis 427 (dort `66-39` und die 27 verworfenen Werte werden 66-40 und 26), 450 bis 451.
- `src/adapters/driven/telemetry/regeln_test.go`: `alleCodes()` (um `ODR-E2102`); `TestFehlercodeForm` hält die Zahl gegen `model.AlleCodes()`.
- `src/hexagon/services/mock/lesen_fehler_test.go`: `codeListen` (Liste `u` der übrigen Codes um `ODR-E2102`; ohne sie ist `TestRueckgabenMenge` rot). Diese Stelle fehlte im Plan.
- `src/adapters/driving/httpodata/fehler_test.go`: `TestFehlerobjektJeCode` rechnet `len(model.AlleCodes()) - 2 + 1` Paare (zwei Codes nur mit Status 0, `ODR-E4001` mit zwei Stati); `ODR-E2102` trägt drei Stati, die Rechnung bekommt `+ 2`. Auch diese Stelle fehlte im Plan.
- Keine Stelle: die Telemetrie, die Inspektion und die Tests von `httpodata` über `codeStatusPaare` zählen gegen `model.AlleCodes()`; `klassen_test.go` nennt „35 Paare“ nur im Kommentar (Bestand, nicht Teil dieses Slice).

**S3. Katalog und Handbuch.** Bestätigt: `ODR-E2102` steht im selben Commit wie `fehler.go` in `harness/katalog-ausnahmen.txt` (`<Code> | <Grund>`, direkt unter `ODR-E2101`, Grund: das Binary verdrahtet den Listener nicht); das Handbuch bleibt unverändert, `make handbuch-check` nennt den Code nicht und `make katalog-check` führt ihn über die Ausnahme (`SPEC-201` Nr. 3 e). Den Code betritt das Handbuch mit `slice-bootstrap-m6`, der die Ausnahme auflöst und die Handbuch-Pflicht in seiner DoD trägt (gelesen, 2026-11-07).

**S4. Nummernlage.** Keine neue `SPEC`-Nummer.

**S5. Raster.** Bestätigt: `SPEC-238` Nr. 2 trägt die Randformen der Zuordnung im Wortlaut (`.`, `..`, `%2F`, leeres Segment, abschließender `/`, Groß- und Kleinschreibung); ergänzt wurden UTF-8 und Nicht-UTF-8, der Pfad ohne führenden `/`, die Prozentfolge im Literal, die Methode des Requests und die Kopie der Routen. Das Raster in §6 nennt die Zellen mit den Nummern.

**S6. Mengen, Zweige, Fakes, Bindung.** Bestätigt mit den Ergänzungen oben: Mengen M-Route (786 Fälle, Zeichenmengen aus dem Vertrag), M-Überlappung, M-Zuordnung, M-Methode × Pfad (32 + 7; Implementer-Ergänzung, §6), M-Grenze (64 Routen, 4 Parameter) mit den Prüfgrenzen unten; Zweige Z-1 bis Z-5 mit Zählprüfung (null Blöcke mit Zähler 0 unter `-race -coverprofile`, Ausreißer Z-5); Aufrufstellen und Fakes (`listenerFake` zählt auch `Addr`, kein Fake wirft eine Panik; die Panik einer Route fängt `slice-admin-fehlerabbildung`, B8); Bindung mit dem Methodenwert von `routeFake`.

**S7. Merge-Nähe.** In `spec/spezifikation.md` stehen nur zwei Zeilen geändert: `SPEC-238` Nr. 1 und der Absatz *Zuordnung* in Nr. 2. Das Änderungsverzeichnis, `SPEC-109`, `SPEC-230` Nr. 8 und `SPEC-240` waren vollständig und bleiben. Wer als zweiter nach `main` merged, gleicht die Zahl der Codes an den Stellen aus S2 an (fünf Dateien).

**S8. Größe.** Bestätigt: *a* = 0 (keine nebenläufige Zusage; Nr. 12 liegt bei N1 in `slice-admin-abweisen-audit`), *b* = 13 (K1 bis K3, Z1, Z3 bis Z8, Y2 bis Y4), *f* = 3, *p* = 110 (Rahmen etwa 90, Code und Katalog etwa 20; ein größerer Produktcode höbe nur die Marke, der gemessene Produktcode von `slice-admin-listener-token` lag mit 174 gegen 150 um 16 Prozent darüber), also *S* = 1750 und *M* = 2400, keine Schnittfrage. Der Block führt die Pfade des Planners und zwei weitere (Rückgabe R-1): `src/hexagon/services/mock/lesen_fehler_test.go`, `src/adapters/driving/httpodata/fehler_test.go`. Auslastung: die Messpunkte der letzten Teile liegen bei 0,41 bis 0,71 (`schaetzformel-ueberschaetzt-den-slice`); bei 0,41 bis 0,71 von 2400 wären es 984 bis 1704 Zeilen. Unter 0,60 meldet der Planner dem Architect bei der Closure (`SPEC-128` Nr. 9), der die Faktoren dann aus allen Messpunkten neu rechnet; die Formel bleibt jetzt, wie sie ist.

**S9. Reports und Betriebsregeln.** Zur Kenntnis genommen; der Satz zu den Kennungen von ADRs geht durch den Planner an Reviewer und Verifier. Dieser Architect-Lauf hatte kein `Edit`: die Zeilen in `spec/spezifikation.md` und in diesem Plan sind mit `Write` in Hilfsdateien geschrieben und eingespielt (`git apply` bzw. Zusammensetzen mit `sed -n` und `cat`), kein `sed -i`, kein Heredoc, kein `>>`.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.
- Auftrag an Reviewer und Verifier (Planner, S9): Reports tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link (Pfad ab dem Ort des Reports), sonst meldet `make docs-check` `id-unlinked` (`BEO-REPO/report-der-rolle-faellt-docs-check`); beide nennen im Report, dass sie `make docs-check` über den Report gefahren haben.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-admin-listener-token` (Teil 1, Token-Datei) liegt in `done/`; `slice-audit-log` und `slice-telemetry-port` (Code-Kanten der Felder `Audit` und `Telemetrie` von `Konfig`) liegen in `done/` (geprüft am 2026-11-07, Planner, `main` = `aadbef0`). Die Bedingung an Teil 1 ist eine Reihenfolge, keine Code-Abhängigkeit: beide Teile schreiben in `src/adapters/driving/admin/`, und die Messung nach `SPEC-128` Nr. 4 g zählt jede Änderung an den Pfaden der Marke, auch die eines parallelen Zweigs; die Teile 2 bis 6 laufen deshalb nacheinander. Die Spezifikation (`SPEC-238` bis `SPEC-240`) trägt den Vertrag samt Randformen seit dem Zweig von Teil 1 (`AGENTS.md` §3.12) und liegt mit dessen Merge auf `main`. **Vor dem ersten Code-Commit zieht der Zweig `main` nach** (Planner im Start-Commit): `main` trägt seit dem Merge von `slice-mock-lesen` den Code `ODR-E6008` (38. Code), `ODR-E2102` ist der 40.: `slice-mock-schreiben` (`done/`, Merge nach `main` am 2026-11-07) vergab `ODR-E6009` als 39. Code, `SPEC-230` Nr. 8 steht auf 40 (Integrationsschritt des Planners); die Tests des Modells, des Audit und der Telemetrie zählen auf `main` 39 und gehen mit `ODR-E2102` auf 40 (Y2).

**Stand beim Start** (Planner, 2026-11-07, `main` = `4d441211`): erfüllt. `slice-admin-listener-token`, `slice-audit-log` und `slice-telemetry-port` liegen in `done/`; `SPEC-238` bis `SPEC-240` stehen auf `main`; `src/adapters/driving/admin/` trägt die Token-Datei (`tokendatei.go`), `model.AlleCodes()` zählt 39 Codes (`ODR-E6008`, `ODR-E6009` darin, `ODR-E2102` fehlt). §1 *Ausdrücklich NICHT* und DoD der Nachbarn gelesen (`AGENTS.md` §3.13): `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` nehmen die Punkte an, die §1 ihnen zuweist (Authentisierung, `Serve`, Abbildung und Zählen, Audit des Abweisens); `slice-bootstrap-m6` nimmt die Handbuch-Pflicht des Codes an. Die Gate-Kette bleibt unberührt (Nr. 7 der Übergangsregeln), das Binary bleibt unverändert, der Zweig-Start hängt an keinem anderen Slice in Arbeit (`slice-snapshot-use-case`, `slice-sqlite-interrupt-transaktion` teilen nur `spec/spezifikation.md`, siehe S7). Speicher beim Start: 11 GB verfügbar. **Marke:** der Block `groessenmarke` steht in diesem Abschnitt (Basis = Commit des Architect `0504b729`).

**Größenmarke** (Planner, 2026-11-07, `SPEC-128` Nr. 1; vom Architect bestätigt, S8): 13 Zeilen der Tabelle, *a* = 0, *b* = 13, *f* = 3, *p* = 110: *S* = 85·0 + 80·13 + 200·3 + 110 = 1750, *M* = 2400, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Der Block steht unten; `basis` ist der Commit des Architect, der den Vertrag abschließt, die Pfade sind die sechs des Planners und die zwei der Rückgabe R-1 (acht).

```groessenmarke
pfad: src/adapters/driving/admin/
pfad: src/hexagon/model/fehler.go
pfad: src/hexagon/model/fehler_test.go
pfad: src/adapters/driven/audit/audit_test.go
pfad: src/adapters/driven/telemetry/regeln_test.go
pfad: harness/katalog-ausnahmen.txt
pfad: src/hexagon/services/mock/lesen_fehler_test.go
pfad: src/adapters/driving/httpodata/fehler_test.go
produktcode: 110
fakes: 3
basis: 0504b729a1f998b4c070d9d63f22173c5147af7b
schaetzung: 1750
marke: 2400
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 2400 (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein Teil, der sich ohne geteilten Mechanismus abtrennen lässt, steht dann vor dem Code in §4, sonst steht die Messung mit Grund im Plan (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert — Carveout?): der Architect ändert den Vertrag von `SPEC-238` Nr. 1 oder 2 (Rückgabe an den Planner), oder die Nummer `ODR-E2102` stößt beim Nachziehen von `main` mit einem Code eines anderen Zweigs zusammen.

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-07, Zeilen dieses Teils): jede Zelle nennt eine Nummer von `SPEC-238` (Rahmen) oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Konfiguration und Routen (`New`) | `SPEC-238` Nr. 1: nil in `Listener`, `IDs`, `Audit`, `Telemetrie` | Nr. 1 und 2: Token unter 32 Bytes, leeres Muster; eine leere Liste der Routen ist gültig | Nr. 2 a bis d: Methode, Muster, nil-`Handler`, Überlappung | gilt nicht: `New` hat keinen `ctx` und ruft keinen Port (Nr. 1) | Nr. 1: zwei Server teilen nichts |
| Zuordnung von Pfad und Methode | Nr. 2: kein Treffer ist 404; ein Pfad ohne führenden `/` (`*`) passt auf keine Route | Nr. 2: leeres Segment (`//`), abschließender `/` und der leere Pfad passen auf keine Route | Nr. 2: Groß- und Kleinschreibung zählt (Pfad und Methode des Requests), ein Parameter nimmt jedes nichtleere Segment (auch UTF-8 und Nicht-UTF-8), ein Literal passt auf das dekodierte Segment | gilt nicht: reine Funktion ohne Port (Nr. 2) | Nr. 12: zustandslos |

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16): gilt nicht für `New`, weil jeder Verstoß dasselbe `model.ErrUngueltig` ohne Wert ist (`SPEC-238` Nr. 1, 2); ein Paar ändert Art und Code nicht, zugesagt ist nur, dass der Text keinen Wert nennt (K1, K2). Für die Zuordnung hat jeder Pfad genau ein Ergebnis (404 oder 405 oder Treffer); die Reihenfolge gegenüber der Authentisierung (401 vor 404 und 405, `SPEC-238` Nr. 3 d, e) entscheidet `slice-admin-authentisierung` (A5).

- Die Dateien, die dieser Teil mit anderen Zweigen teilt: `spec/spezifikation.md` (nur der Architect, hier nicht geschrieben), `src/hexagon/model/fehler.go`, `fehler_test.go`, `src/adapters/driven/audit/audit_test.go`, `src/adapters/driven/telemetry/regeln_test.go` und `harness/katalog-ausnahmen.txt` (mit `slice-mock-schreiben`, das einen weiteren Code vergeben kann, und früher mit `slice-mock-lesen`, dessen `ODR-E6008` auf `main` liegt) — **Ausgang:** entschieden (Planner, Integrationsschritt, 2026-11-08): `main` (`6dbe5d02`) in den Zweig gemerged ohne Konflikt; `model.AlleCodes()` zählt im Zweig 40 Codes, die Tests der Modelle, der Telemetrie und des Audit zählen gegen `AlleCodes()` und tragen kein Literal der Zahl, diese Stellen brauchten kein Nachgleichen, eine sechste Literal-Liste im Test von `slice-snapshot-use-case` schon (dritter Punkt unten) (Beleg: `make gates` auf `main` nach dem Merge; Verifier-Probe `git merge-tree` ohne Konflikt) (Auftrag war: Integrationsschritt: wer als zweiter nach `main` merged, zieht `main` nach und gleicht die Zahl der Codes in `fehler_test.go`, `regeln_test.go` und `audit_test.go` sowie die Zeile in `SPEC-230` Nr. 8 an; Stand beim Start: `main` zählt 39 mit `ODR-E6008` und `ODR-E6009`, mit `ODR-E2102` 40; kommt der Code von `slice-snapshot-use-case` hinzu, ist er der 41., und die Zahlen folgen `model.AlleCodes()`).
- Zwei weitere Dateien teilen die Zahl der Codes: `src/hexagon/services/mock/lesen_fehler_test.go` (Liste der übrigen Codes) und `src/adapters/driving/httpodata/fehler_test.go` (Zahl der Paare aus Code und Status); jeder Zweig, der einen Code vergibt (`slice-snapshot-use-case` meldet es), führt sie mit — **Ausgang:** entschieden (Planner, 2026-11-08): beide Dateien sind im Zweig mit `ODR-E2102` nachgezogen und zählen gegen `AlleCodes()`; der Merge von `main` brauchte keinen Nachzug (wie der Punkt davor; die Stellen stehen im Block *Bestätigungen des Architect*, S2).
- Hinweis an den Integrations-Planner (Merge-Nähe, S2 und S7): `slice-snapshot-use-case` und `slice-sqlite-interrupt-transaktion` teilen mit diesem Zweig die Zahl der Codes; wer von den Zweigen, die einen Code vergeben, als zweiter nach `main` merged, gleicht die Zahl an **fünf** Test-Dateien an: `src/hexagon/model/fehler_test.go`, `src/adapters/driven/telemetry/regeln_test.go`, `src/adapters/driven/audit/audit_test.go`, `src/hexagon/services/mock/lesen_fehler_test.go`, `src/adapters/driving/httpodata/fehler_test.go`, dazu die Zeile in `SPEC-230` Nr. 8; die Zahlen folgen `model.AlleCodes()` — **Ausgang:** entschieden (Planner, 2026-11-08): dieser Zweig merged nach `slice-snapshot-use-case` (39 Codes, keiner neu) und `slice-sqlite-interrupt-transaktion` (keiner neu); `ODR-E2102` ist der 40. Code, die fünf Dateien folgen `AlleCodes()` (eine sechste Stelle, die Liste `nie()` in `src/hexagon/services/snapshot/fehler_test.go` aus `slice-snapshot-use-case`, trägt eine Liste der Codes als Literal: der Planner hat `ODR-E2102` im Integrationsschritt eingetragen, `make test` war davor rot), `SPEC-230` Nr. 8 nennt keine feste Zahl (geprüft: keine Literalzahl der Codes in der Spezifikation außerhalb der Änderungstabelle, die beide Seiten behält).
- Aufträge S1 bis S9 an den Architect zum Start (§3): Sätze und Zeilen, Zählprüfung gegen `model.AlleCodes()`, Katalog und Handbuch, Nummernlage, Raster, Mengen mit Zweigen und Fakes, Merge-Nähe, Größe, Reports — **Ausgang:** S1 bis S8 entschieden (Architect, 2026-11-07, Block *Bestätigungen des Architect* in §3 und die Zeilen oben); S9 entschieden (Planner, 2026-11-07): der Satz zu den Kennungen von ADRs steht als Auftrag an Reviewer und Verifier in §3 (Betriebsregeln); Adresse für den Rest: Implementer (Tests und Zählprüfungen der Zeilen).
- Rückgabe R-1 an den Planner (Architect, 2026-11-07, `AGENTS.md` §3.13): der Block `groessenmarke` führt zwei Pfade mehr als die Liste in §4, weil der Code `ODR-E2102` zwei bestehende Tests rot macht, die die Codes zählen (S2): `src/hexagon/services/mock/lesen_fehler_test.go` und `src/adapters/driving/httpodata/fehler_test.go`; beide mit wenigen Zeilen, *p* und die Zeilen der Tabelle bleiben — **Ausgang:** erledigt (Planner, 2026-11-07): beide Pfade stehen im Block in §4, im Commit, der `basis`, `schaetzung` und `marke` setzt, vor dem ersten Code-Commit.
- Die eingebaute Route `GET /healthz` steht in diesem Teil in der Zuordnung und antwortet 200 mit dem Literal, bevor die Authentisierung existiert; die Zusagen „ohne Token“ und „nur `GET`“ (A3, A4) fängt erst `slice-admin-authentisierung`. Grund: die Zeilen A3 und A4 brauchen das 401 als Gegenprobe. Folge: bis Teil 3 in `done/` liegt, hat die Antwort auf `GET /healthz` keinen Fänger der Ausnahme von der Authentisierung; nichts erreicht den Listener, das Binary bleibt unverändert — **Ausgang:** entschieden (Planner, Schnitt): Fänger in `slice-admin-authentisierung`.
- Die Felder `Audit` und `Telemetrie` sind geprüft und gehalten, aber nicht gerufen; ihre Zusagen (Zählen, Audit des Abweisens) liegen in `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` — **Ausgang:** entschieden (Planner, Schnitt): K3 fängt, dass dieser Teil sie nicht ruft (Zähler 0).
- `make katalog-check` ist rot, sobald `fehler.go` den Code `ODR-E2102` trägt, bis die Zeile in den Ausnahmen steht (`SPEC-201` Nr. 3 e) — **Ausgang:** entschieden (Y3): Code und Zeile im selben Commit.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (12×): jeder Zweig der Zuordnung und von `New` hat einen Lauf unter dem Detektor (Z-Liste in §3) — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): die Fakes `zaehlendeIDs`, `routeFake` und `listenerFake` tragen jede Schnittstelle des Originals und zählen jede Operation — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/dienst-erfuellt-den-port-ohne-bindung` (2×): die Bindung steht in `bindung_test.go` (§3) — **Ausgang:** weiter offen → `BEO-REPO/dienst-erfuellt-den-port-ohne-bindung` im Register.
- Größe: der Block steht in §4 (*S* = 1750, *M* = 2400, `basis` = `0504b729`, acht Pfade) — **Ausgang:** entschieden (Planner, 2026-11-07): Block in eigenem Commit vor dem ersten Code-Commit.
- Gate-Kette (Nr. 7 der Übergangsregeln von `welle-benutzerhandbuch`): der Teil ändert weder `GATE_CHECKS` noch ein Gate; `harness/katalog-ausnahmen.txt` ist Eingabe von `make katalog-check`, kein Gate-Fragment — **Ausgang:** entfallen; kein Pfad der Gate-Kette im Plan (§3).
- Abweichung des Implementers (`AGENTS.md` §3.13), Z6: sieben Zusatzfälle statt vier (`get` auf `/b`, `get` auf `/a/b`, `get` und `Get` auf `/healthz`, leere Liste 3). Grund: die Mutation „Methode ohne Beachtung der Schreibweise verglichen“ bleibt bei `get` auf `/b` grün, denn `get` ist auch ohne Beachtung der Schreibweise nicht `PUT`; erst `get` auf einem Pfad mit einer Route `GET` fängt sie. Die Zeile Z6 und die Menge M-Methode × Pfad (32 + 7) tragen die Fälle in diesem Commit — **Ausgang:** entschieden (Planner, 2026-11-08, Integrationsschritt): die Fälle bleiben; sie prüfen den Wortlaut von `SPEC-238` Nr. 2 („Groß- und Kleinschreibung zählt“, Raster in diesem Abschnitt), ergänzen ihn nicht, und der Verifier hat Z6 gegen den Vertrag gelesen (Report Abschnitt 3: Code und Fänger stimmen, keine Abweichung im Code). Eine Spec-Änderung ist nicht nötig; die Bestätigung des Architect ist nicht Bedingung für `done/`.
- Abweichung des Implementers, Z4: das leere Segment an jeder Parameterstelle (404) als achter Wert neben den sieben (3 Fälle mehr, 42 statt 39); in der Fix-Runde zwölf Werte mehr, in denen Pfad- und Query-Dekodierung sich unterscheiden (19 Werte, 78 Fälle). Grund: die Mutation „Parameter nimmt das leere Segment“ blieb bei den sieben Werten grün und wurde nur von Z5 (`/a/`, `//a`) gefangen; Z4 trägt die Zusage „jedes nichtleere Segment“ nun selbst — **Ausgang:** entschieden (Planner, 2026-11-08): wie der Punkt davor (der Wortlaut „jedes nichtleere Segment“ steht in `SPEC-238` Nr. 2); die Grenzen der Menge nennen N-8 bis N-11.
- Abweichung des Implementers, Z3: 120 Muster (alle Folgen von ein bis vier Segmenten über Literal `a`, `b` und Parameter; 7260 Paare, 14520 Paarfälle) statt der acht des Architect; erst `/b/c` ergänzt (9 Muster), in der Fix-Runde die Klasse „jede Stelle bei jeder Länge“. Grund: die Mutation „nur das letzte Segment verglichen“ blieb mit den acht Mustern grün, weil kein Paar gleicher Länge nur im ersten Segment verschieden ist und im letzten übereinstimmt — **Ausgang:** entschieden (Planner, 2026-11-08): wie der Punkt davor; der Verifier fuhr 7 von 7 fremde Mutationen der Klasse rot (Report Abschnitt 4), Z3 ist als Klasse gedeckt.
- Fund, Teil 1 (`slice-admin-listener-token`, `done/`), nicht Gegenstand dieses Teils: `lade` in `tokendatei.go` meldet bei gleichzeitigem Erzeugen gelegentlich `ODR-E2001 ADMIN_TOKEN_FILE: keine reguläre Datei`. Ursache (gelesen, nicht behoben): zwischen `os.Stat` (fehlt) und `os.Lstat` in `istFehlend` kann ein anderer Aufruf die Datei verlinken; dann ist `Stat` gescheitert und `Lstat` gelingt, `statFehler` bildet es auf `grundKeineDatei` ab. Folge: `TestWettlaufNebenlaeufig` (T19) ist unter Last rot (Beobachtung: 3 von 5 Läufen im Container bei Last 55 bis 69). Der Implementer ändert den Code von Teil 1 nicht, denn `SPEC-239` Nr. 1 a entscheidet „fehlt“ gegen „Symlink ohne Ziel“ und nennt den Wettlauf zwischen den beiden Aufrufen nicht — **Ausgang:** weiter offen → `BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel` im Register (Planner, 2026-11-08): benannte Spec-Lücke `SPEC-239` Nr. 1 a und Nr. 3 d (Randform „Datei erscheint zwischen `Stat` und `Lstat`“ ohne Ort); Adresse: Architect (entscheidet die Randform, `AGENTS.md` §3.12), danach ein Fix-Slice des Planners für `tokendatei.go`; der Test `TestWettlaufNebenlaeufig` bleibt bis dahin unter Last gelegentlich rot und ist kein Befund dieses Slice. Der Verifier sah ihn in seinem Gate-Lauf grün.
- N-7 (Implementer, Zeile K3): eine Goroutine, die eine von `New` gerufene Funktion startet und vor der Messung endet, bleibt für `TestNewRuftNichts` unsichtbar; der Test „keine go-Anweisung in `New`“ fängt nur den Rumpf von `New` — **Ausgang:** entschieden (akzeptiertes Negativ, Grund: ohne Haltepunkt sieht kein Test eine beendete Goroutine; Folge: eine kurzlebige Goroutine in `baueRouten` bleibt grün; das Review liest die Datei `admin.go`).

- Verifier-Befunde 1 bis 4 (Report `docs/reviews/2026-11-08-slice-admin-rahmen-verifikation.md` Abschnitt 4; Z8 mehrzeilige `Authorization`, Z4 Literal-Nachbarn, Z4 zweimaliges Dekodieren, Prüfgrenzen) — **Ausgang:** entschieden (Planner, 2026-11-08): Befund 1 geht an `slice-admin-authentisierung` (dort in §1, DoD und Zeile A7 eingetragen, Geber `slice-admin-rahmen`, vor dieser Zeile, `AGENTS.md` §3.13), der Rest von 1 und die Befunde 2 bis 4 sind akzeptierte Negative N-8 bis N-11 in §3; kein zweiter Fix-Lauf (Durchstich vor Härtung).

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

**Closure** (Planner, Integrationsschritt, 2026-11-08; Zweig `slice-admin-rahmen`, Stand `edeab666`, danach `main` (`6dbe5d02`) eingemerged). Liefer-Punkte: drei (Konfiguration/Routen/Zuordnung, Code `ODR-E2102` mit Katalog-Ausnahme, Gates), das Limit gilt.

**Was geliefert ist:** `Konfig`, `New`, Routen, Zuordnung mit 404 und 405 (`ODR-E2102`), `Anfrage` und der Code `ODR-E2102` (40. Code) in `src/adapters/driving/admin/` und `src/hexagon/model/fehler.go`; `Allow`, Zählprüfungen und Bindung stehen in den Tests des Pakets. Das Binary ist unverändert (`src/bootstrap` und `tests/` ohne Diff, Verifier Abschnitt 1); `ODR-E2102` steht in `harness/katalog-ausnahmen.txt`, das Handbuch bleibt (`slice-bootstrap-m6` trägt die Handbuch-Pflicht).

**Mutationen:** Implementer 110 Läufe der ersten Runde (106 rot, vier grün, alle vier bearbeitet: zwei Ergänzungen Z4 und Z3, N-7) und 13 der Fix-Runde (13 rot); Reviewer 32 eigene; Verifier 19 eigene (Z3 7 von 7 rot, Z4 und Z8 11 grün: N-8 bis N-11 und Sendung A7). Alle Fänger sind deterministisch, jede Rate ist 1 von 1 und Größenordnung (Last des Hosts: Implementer Load Average 55 bis 69, Verifier 4 bis 7, jeweils bei 4 CPUs je Container).

**Gate:** Verifier im Zweig `make gates` `EXIT=0` (849 Dateien, 0 Befunde, Zweige unter dem Detektor: ein Ausreißer Z-5, N-4); Planner auf `main` nach dem Merge: der erste Lauf war rot an `TestRueckgabenSnapshot` (sechste Zählstelle, siehe Steering-Loop-Eintrag 4), nach dem Nachzug grün (Ergebnis im Bericht des Integrationslaufs).

**Größe:** Schätzung *S* 1750, Marke 2400, Messung 1785 (`make groessenmarke`: ok), Ende ÷ *S* 1,02, Auslastung 0,74 der Marke. Der Punkt liegt über 0,60: keine Meldung an den Architect (`SPEC-128` Nr. 9). Er bestätigt, dass die Schätzformel bei einem Slice mit Zuordnungs-Generatoren (Z1, Z3, Z4, Z8 mit großen Mengen) nicht überschätzt; Evidenz zu `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` wird nicht angelegt, weil der Punkt kein Auftreten der Beobachtung ist.

**Akzeptierte Negative:** N-1 bis N-7 (Handoff) und N-8 bis N-11 (Verifier, jetzt mit Grund und Folge in §3); Befund 1 des Verifiers geht an `slice-admin-authentisierung` (Zeile A7, dort in §1 und DoD, Geber dieser Slice).

**Risiko-Ausgänge (§6):** je Punkt im Plan. Entschieden: Zahl der Codes (drei Punkte; Merge ohne Nachzug), Aufträge S1 bis S9, R-1, healthz, Felder `Audit` und `Telemetrie`, Katalog, Größe, Gate-Kette, die drei Abweichungen Z6/Z4/Z3 (Planner: die Fälle bleiben, sie prüfen den Wortlaut von `SPEC-238` Nr. 2), die Verifier-Befunde 1 bis 4, N-7. Weiter offen: das Register (`zweig-ohne-lauf-unter-dem-detektor`, `hilfe-ersetzt-argument-verdeckt-eigenschaften`, `panik-eines-fakes-verdeckt-den-benannten-test`, `dienst-erfuellt-den-port-ohne-bindung`) und der Fund Teil 1 (`BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel`, Adresse Architect).

**Steering-Loop-Einträge:**

1. **Die Klasse eines Fixes wurde an der Achse des Befundes gezogen, nicht an der Zusage.** Die Fix-Runde deckte Z4 mit zwölf Formen, in denen Pfad- und Query-Dekodierung sich unterscheiden, und Z8 mit Schreibweisen und Nachbarnamen; der Verifier fand an Elementen, die weder der Befund noch der Fix nannte, elf grüne Mutationen: die Form „dekodiert ergibt wieder eine Prozentfolge“ (Achse: Wert nach dem Dekodieren), ein Literal, das im Segment enthalten ist (Achse: Länge des Segments gegen das Literal), `Authorization` mit zwei Zeilen (Achse: Zahl der Zeilen je Name). Z3 war als Klasse gedeckt (7 von 7 rot). Evidenz zu `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`; die Lücke liegt im Bilden der Menge: die Zusage „bytegleich“ und „dekodiert genau einmal“ nennt die Nachbarn der Gleichheit und die Wiederholung der Operation, die Fixtures führten beide nicht. Kein neuer Zielort: die Regel (`AGENTS.md` §3.17) steht, der Eintrag zählt hoch.
2. **Unbeschränkte Mengen ohne Prüfgrenze im Plan** (Länge des Parameterwertes, Name und Wert eines Headers): Evidenz zu `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze`; der Planner trägt die Grenzen als N-11 ein.
3. **Benannte Spec-Lücke:** `SPEC-239` Nr. 1 a und Nr. 3 d (Teil 1) nennen den Wettlauf zwischen `Stat` und `Lstat` beim gleichzeitigen Erzeugen der Token-Datei nicht; `TestWettlaufNebenlaeufig` ist unter Last gelegentlich rot (3 von 5 Läufen bei Load Average 55 bis 69, im Gate des Verifiers grün). Register `BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel` (neu), Adresse Architect, danach ein Fix-Slice des Planners; hier kein `liegt in`, die Lücke liegt in der versionierten Spezifikation.

4. **Eine Zählstelle, die der Plan nicht nannte, weil sie im Zweig eines anderen Slice entstand.** Der Plan nannte fünf Test-Dateien mit der Zahl oder Liste der Codes (S2, Suche am Start über `src/`); `slice-snapshot-use-case` brachte danach `nie()` in `src/hexagon/services/snapshot/fehler_test.go` mit einer Literal-Liste, die jeden Code einer von zwei Listen zuweist. Der Verifier las „folgen `AlleCodes()`“ und sah die Probe `git merge-tree` ohne Konflikt; der Fehler zeigte erst `make gates` auf `main`. Gelernt: ein Merge ohne Textkonflikt sagt nichts über Zählstellen, die ein anderer Zweig neu einführte. Evidenz zu `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle`; Folge für den Integrationsschritt: nach dem Merge ein Suchlauf nach `"ODR-E2101"` (Nachbar-Code) über `src/` vor dem ersten `make gates`.

**Kein Feld `liegt in` in diesem Slice:** nichts wurde in einer Regel, einem Sensor oder einer Spec-Stelle neu verkörpert, deren Zielort dieser Slice trägt. Eine Pflicht zur Anker-Paarung entsteht nicht.

**Register:** Evidenz `slice-admin-rahmen` zu `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (Eintrag 1), `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (Eintrag 2) und `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (Eintrag 4); neues Verzeichnis `BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel` (Eintrag 3). Die weiteren Register-Zeilen aus §8 bestätigen den Stand ohne neue Form.

**Drei Paarungen:** Anker (kein Eintrag trägt `liegt in`, nichts zu paaren) · Folge-Slice (`slice-admin-authentisierung` führt die Sendung A7 in §1, DoD und Tabelle; `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` liegen als Dateien in `open/`) · Register (die drei genannten Verzeichnisse tragen `evidence/slice-admin-rahmen.md` oder liegen neu im Register).

**Start-Trigger des Folgers:** [slice-admin-authentisierung](../done/slice-admin-authentisierung.md) (Teil 3) hat als Bedingung, dass dieser Slice in `done/` liegt; sie tritt mit dem Merge nach `main` ein. Der Start ist Sache des Planners (Block `groessenmarke`, `basis`, `schaetzung`, `marke` vor dem ersten Code-Commit).

**Hinweise an den User (nicht entschieden):** (1) Der Fund in Teil 1 (`SPEC-239` Randform) wartet auf den Architect. (2) Das Image-Tag `odata-recorder:test` ist über alle Worktrees geteilt und kann den Stand eines anderen Zweigs tragen (Verifier Abschnitt 1); `BEO-REPO/image-tag-race-von-allen-worktrees-geteilt` zählt es bereits.

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

**Vorgelagert — offene Beobachtungen sichten:** Register beim Schnitt nachgezogen (Planner, 2026-11-07, `main` = `aadbef0`) und beim Start erneut (Planner, 2026-11-07, `main` = `4d441211`, Zähler aus den Dateien unter `evidence/`): einschlägig sind `zweig-ohne-lauf-unter-dem-detektor` 16×, `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `dienst-erfuellt-den-port-ohne-bindung` 2×, `betriebsregel-fuer-repo-dateien-verletzt` 12×, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 47×, `randform-erst-im-review-entschieden` 34×, `menge-im-plan-weicht-vom-orakel-des-tests-ab` 3×, `unbeschraenkte-menge-ohne-benannte-pruefgrenze` 11×, `schaetzformel-ueberschaetzt-den-slice` 25×, `report-der-rolle-faellt-docs-check` 5× (Verkörperung, soweit die Schwelle erreicht ist, beim Lese-Schritt der Welle-Closure); kein neues Verzeichnis, kein Zähler gesetzt.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
