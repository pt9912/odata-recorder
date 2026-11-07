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

**Berührte Spec-Stellen:** `SPEC-238` Nr. 1 (`Konfig`, `New`), Nr. 2 (Routen, Zuordnung, `Anfrage`, 404 und 405), Nr. 3 a (Korrelations-ID), Nr. 5 (Fehlerobjekt für 404 und 405), Nr. 6 (Code `ODR-E2102`), Nr. 14 (Katalog) · `SPEC-239` — nicht berührt · `SPEC-240` (Zeile `ODR-E2102`) · `SPEC-109` Nr. 2 (Zeile `ODR-E2102`) · `SPEC-230` Nr. 8 (39 Codes) · `SPEC-114` (Korrelations-ID) · `SPEC-006` · `ARC-007`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie. `SPEC-238` bis `SPEC-240` stehen auf `main` (seit dem Merge von `slice-admin-listener-token`).

**Verantwortlich:** —

**Autor:** pt9912 (Planner). **Datum:** 2026-11-07.

**Herkunft:** Teil 2 des Schnitts von [slice-admin-listener-token](../in-progress/slice-admin-listener-token.md) (Vorschlag des Architect, Rückgabe R1; Schnitt vor dem Code vollzogen vom Planner am 2026-11-07 nach `SPEC-128` Nr. 1, `AGENTS.md` §3.14). Die Teile in der Reihenfolge ihres Starts: 1 `slice-admin-listener-token` (Token-Datei), 2 dieser Slice, 3 [slice-admin-authentisierung](slice-admin-authentisierung.md), 4 [slice-admin-lebenszyklus](slice-admin-lebenszyklus.md), 5 [slice-admin-fehlerabbildung](slice-admin-fehlerabbildung.md), 6 [slice-admin-abweisen-audit](slice-admin-abweisen-audit.md). Der Plan trägt die Zeilen K1 bis K3, Z1, Z3 bis Z8 und Y2 bis Y4 der Tabelle des Architect (13 von 102); sie stehen im Wortlaut in §3.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Paket `src/adapters/driving/admin` führt `Konfig` und `New` (`Listener`, `IDs`, `Audit`, `Telemetrie` und `Token` geprüft, die Routen geprüft und kopiert), ordnet Methode und Pfad einer Route zu (kein Treffer: 404, falsche Methode: 405 mit `Allow`, die eingebaute `GET /healthz` steht in der Zuordnung), reicht der Route die `Anfrage` ohne `Authorization`, vergibt die Korrelations-ID vor allem anderen und schreibt das Fehlerobjekt (an der einen Stelle, die es schreibt) für 404 und 405; dazu der neue Code `ODR-E2102` (400, 404, 405; 39. Code) samt Katalog-Ausnahme, Zähler-Reihe und Tests der Codetabelle. Der Teil authentisiert noch nicht (die Tests senden einen Bearer-Header, den er nicht liest), zählt noch nicht und auditiert noch nicht.

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

- [ ] Konfiguration, Routen und Zuordnung: `Konfig` und `New` nach `SPEC-238` Nr. 1 (nil-Prüfung der vier Ports, Token ab 32 Bytes, `New` ruft keinen Port und startet nichts), Routen nach Nr. 2 a bis d (Methode, Muster, nil-`Handler`, Überlappung samt eingebauter `GET /healthz`), Zuordnung (Literal bytegleich, Parameter auf jedes nichtleere Segment, kein Bereinigen des Pfades), 404 und 405 (`Allow` sortiert, die eingebaute Route eingeschlossen) mit `ODR-E2102`, `Anfrage` mit allen Feldern ohne `Authorization`, `KorrelationsID` vor allem anderen; der Adapter hat keine Standardadresse und ruft nirgends `net.Listen`. Bindung der Fakes und des Adapters an die Ports durch Zusicherungen zur Übersetzungszeit (`bindung_test.go`).
- [ ] Code `ODR-E2102` (`SPEC-240`, 39. Code): `src/hexagon/model/fehler.go` und `fehler_test.go` (Status 400, 404, 405, Ziel `admin`, Text `Admin request invalid`, hinter `ODR-E2101`), die Reihe des Zählers (`SPEC-230` Nr. 8, `regeln_test.go`), die Liste in `audit_test.go` des Audit und die Zeile in `harness/katalog-ausnahmen.txt` **im selben Commit wie `fehler.go`** (`make katalog-check` ist sonst rot, `SPEC-201` Nr. 3 e).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Für diesen Slice: `ODR-E2102` mit Ausnahme in `harness/katalog-ausnahmen.txt` (`SPEC-238` Nr. 14); das Binary bleibt unverändert (kein Aufruf von `admin` aus `bootstrap`, `src/bootstrap` ohne Änderung), in §7 vermerkt.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

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
| Tests | neu | Adapter gegen `ServeHTTP` mit `httptest`; Fakes `zaehlendeIDs`, `routeFake`, `listenerFake`; `FakeAudit` und `FakeTelemetry` aus `supporttest`; `bindung_test.go`; der Adapter importiert weder Adapter noch Service (`make a-check`) |
| `spec/spezifikation.md` | geliefert (Architect, im Zweig von `slice-admin-listener-token`) | `SPEC-238` bis `SPEC-240`, Nachzug `SPEC-109`, `SPEC-230`; dieser Teil schreibt sie nicht |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stammen vom Architect (2026-11-07, Plan von [slice-admin-listener-token](../in-progress/slice-admin-listener-token.md)); der Schnitt gibt jedem Teil ganze Zeilen im Wortlaut, `Nr.` ist eine Kennung und keine Zählung (K Konfiguration, Z Zuordnung, Y Katalog und Code). **Wachsende Mengen:** Die Schleifen über die Requestarten *R* (A1, E4 in `slice-admin-authentisierung`) und die Antwortarten (E2) wachsen mit den Teilen; dieser Teil trägt in *R* die Formen 404 und 405 und die Route mit 2xx, die Zählprüfung der späteren Teile rechnet die Formen der Vorgänger mit.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| K1 | (`SPEC-238` Nr. 1) `New` weist `Listener`, `IDs`, `Audit` und `Telemetrie` je mit nil ab: `model.ErrUngueltig`, kein `Server`, der Text nennt keinen Wert | Schleife über die vier Felder, daneben der Gutfall; `errors.Is(err, model.ErrUngueltig)`, Server nil | Prüfung eines Feldes entfernt | andere |
| K2 | (Nr. 1) Ein `Token` unter 32 Bytes ist ungültig, 32 Bytes sind gültig; der Fehlertext nennt das Token nicht | Fälle 0, 1, 31, 32, 33 Bytes; ein Canary der Länge 31 (aus Bytewerten) steht nicht im Text | `< 32` zu `<= 32` oder zu `< 31` | andere |
| K3 | (Nr. 1) `New` ruft keinen Port und keine Methode des `Listener` und startet keine Goroutine | Fakes zählen jede Operation (`FakeAudit.Aufrufe`, `FakeTelemetry.Aufrufe`, `zaehlendeIDs`, `listenerFake`): alle 0; Zahl der Goroutinen vor und nach `New` gleich | `New` ruft `IDs.KorrelationsID()` zur Probe | andere |
| Z1 | (`SPEC-238` Nr. 2 a bis c) Eine Route mit ungültiger Methode, ungültigem Muster oder nil-`Handler` macht `New` zu `ErrUngueltig`; gültig sind alle fünf Methoden, jedes Zeichen aus `A-Za-z0-9._~-` im Literal, Parameter wie `{id}` und `{a_b}`, mehrere Parameter, und eine leere Liste der Routen | Generator über Methode (leer, `get`, `HEAD`, `OPTIONS`, `GET ` mit Leerzeichen), Muster (leer, `x`, `/`, `/a/`, `//a`, `/a//b`, `/{}`, `/{A}`, `/{1}`, `/{a`, `/a}`, `/{a}{b}`, `/a{b}`, `/a/{b}/{b}`, `/ä`, `/a b`, `/a?b`) und nil-`Handler`; dazu Gutfälle: die fünf Methoden und jedes Zeichen der Menge (66 Zeichen), `/__admin/sessions/{id}/close`; Zählprüfung: Zahl der Fälle gleich Summe der Elemente der Achsen | Zeichenmenge des Namens `[a-z]` zu `[A-Za-z]`; Prüfung der doppelten Namen entfernt; Zeichen `~` aus der Menge entfernt | andere |
| Z3 | (Nr. 2 d) Zwei Routen überlappen bei gleicher Methode, gleicher Segmentzahl und je Stelle gleichem Literal oder einem Parameter, und die eingebaute `GET /healthz` zählt mit; sonst sind sie gültig | Generator über Paare aus 8 Mustern (`/a/b`, `/a/c`, `/a/{x}`, `/{x}/c`, `/a/{y}`, `/a`, `/a/b/c`, `/healthz`) × Methode gleich oder verschieden; Orakel ein Modell des Tests je Segment; `GET /{x}` und `GET /healthz` ungültig, `POST /healthz` gültig | Segmentzahl nicht verglichen; Methode nicht verglichen; eingebaute Route nicht gezählt | andere |
| Z4 | (Nr. 2) Zuordnung: ein Literal passt bytegleich, ein Parameter auf jedes nichtleere Segment (auch `.`, `..`, einen Wert mit `%2F`, UTF-8), die Segmentzahl muss gleich sein | Generator über Pfade (Literal gleich, in Groß- oder Kleinschreibung abweichend; Parameter in {`x`, `.`, `..`, `a%2Fb`, `ä`}; Segmentzahl ±1) × 3 Muster; `routeFake` hält `Pfadwerte`; Orakel Literale des Tests | Vergleich ohne Beachtung der Schreibweise; Parameter nimmt das leere Segment | andere |
| Z5 | (Nr. 2) Der Adapter bereinigt den Pfad nicht: `//`, ein abschließender `/`, `/./` und `/../` passen auf keine Route (404 mit Token) | Menge von 6 Pfaden mit Token: 404 `ODR-E2102`, `routeFake` 0 Aufrufe | `path.Clean` vor der Zuordnung | andere |
| Z6 | (Nr. 2, 5) Passt der Pfad auf keine Route: 404; passt er, aber nicht die Methode: 405 mit `Allow` aller Methoden des Pfades (die eingebaute eingeschlossen), nach Bytefolge, durch `, ` getrennt | Routen GET, POST, DELETE auf einem Muster; Schleife Methoden (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE) × Pfade (registriert, nur anderes Muster, unbekannt, `/healthz`); Orakel Literale (`DELETE, GET, POST`, `GET`) | `Allow` unsortiert oder ohne die eingebaute Route; 404 und 405 vertauscht | andere |
| Z7 | (Nr. 2) `Anfrage` trägt Methode, dekodierten Pfad, `RawQuery` unverändert, `KorrelationsID`, `Pfadwerte` (nie nil, leer ohne Parameter), `Header`, `Body` und `Laenge` (`-1` bei unbekannter Länge) | `routeFake` hält die `Anfrage`; Schleife über die sieben Felder mit je eigener Eingabe (Query mit `%`-Folgen; Länge 0, 5 und chunked) | ein Feld nicht gesetzt; `Pfadwerte` nil bei einer Route ohne Parameter | andere |
| Z8 | (Nr. 2) `Authorization` erreicht keine Route, jede andere Header-Zeile bleibt | Schleife über Header (`Authorization`, `authorization`, `X-A`, ein Header mit zwei Werten) mit einem Canary; die gehaltene `Anfrage` hat nur die Nicht-`Authorization`-Header | Header ohne Filter kopiert | andere |
| Y2 | (`SPEC-240`) `ODR-E2102` trägt 400, 404 und 405, den Zielbereich `admin`, den Text `Admin request invalid`, steht hinter `ODR-E2101` in `AlleCodes()` | `fehler_test.go`: Schleife über Status 0 bis 999, `StatusErlaubt` genau für 400, 404, 405; `NeuesFehlerobjekt`; Literal-Tabelle um die Zeile ergänzt, Zählprüfung gegen `AlleCodes()` | Status 405 aus `Stati` entfernt | andere |
| Y3 | (Nr. 14) `ODR-E2102` steht mit Grund in `harness/katalog-ausnahmen.txt`, das Handbuch bleibt unverändert | benannte Ausgabe: `make katalog-check` grün (rot ab dem Code in `fehler.go`, bis die Zeile steht) | Zeile nicht eingetragen | andere |
| Y4 | (`SPEC-230` Nr. 8) Der Zähler führt `ODR-E2102`: ein Eintrag je Code der Tabelle, `CountError` erhöht ihn | `regeln_test.go` und `telemetry_test.go`: Schleife über `AlleCodes()`; die Literal-Liste der Codes um `ODR-E2102` ergänzt | die Liste nicht ergänzt (Test rot) | andere |

Zeilen gesamt 13: *a* = 0, *b* = 13. Die Zeile Z2 ging in Z1 auf; K4 (Kopie des Tokens) hängt an der Authentisierung und liegt in `slice-admin-authentisierung`, K5 (Server teilen nichts, `Anhalten`) in `slice-admin-lebenszyklus`. Die Tests dieses Teils senden einen gültigen `Authorization`-Header, den der Teil nicht liest; „mit Token“ in Z5 und Z6 heißt das.

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, Auftrag C; das Orakel trägt Statuscodes, Codes, Header und Meldungstexte als Literale aus dem Vertrag, nie als Konstanten des Codes):

| Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|---|
| M-Route (Z1, Z3) | Methode 5 ungültige × Muster 17 ungültige + nil-`Handler`; Muster 8 → 36 Paare × Methode gleich oder verschieden gleich 72 | Generator, Modell des Tests je Segment | `ErrUngueltig` oder gültig | Elemente der Listen im Test, Paare n(n+1)/2 × 2 |
| M-Methode × Pfad (Z6) | Methoden 8 × Pfade 4 | Schleife | Status und `Allow` als Literale | 32 |

**Unbeschränkte Mengen und ihre Prüfgrenze (akzeptierte Negative):** die Zahl der Routen (Prüfgrenze 64; Grund: die Zuordnung ist linear; Folge: bei mehr Routen wächst die Laufzeit je Request linear, kein Fehler); die Zahl der Parameter je Muster (Prüfgrenze 4; Grund: jeder Parameter ist ein Segment; Folge: ein Muster mit mehr Parametern gilt wie eines mit vier, ohne eigenen Fänger).

**Zweige als Menge, Runden unter dem Detektor** (Architect, Auftrag D; Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat null Blöcke mit Zähler 0, der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| `New`: nil-Feld, Token unter 32, ungültige Route, Überlappung | K1, K2, Z1, Z3 | – |
| Zuordnung: Treffer, 404, 405, eingebaute Route `GET /healthz` (Handler gibt 200 mit dem Literal von `SPEC-238` Nr. 3 c; die Zusage „ohne Token“ fängt `slice-admin-authentisierung`, A3, A4) | Z4 bis Z8 | – |

**Aufrufstellen und zählende Fakes** (Architect, Auftrag E; „jeder Fall je Aufrufstelle“ gilt für diese Stellen):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| genau eine, vor allem anderen | `IDs.KorrelationsID` | Z7 (Feld `KorrelationsID` der `Anfrage`); A1 und E4 in `slice-admin-authentisierung` | `zaehlendeIDs` (Hülle um `FestIDs`, test-lokal): zählt alle fünf Operationen, hält die vergebene Kennung |
| `Close` des Listeners | `net.Listener` | K3 | `listenerFake`: zählt `Accept` und `Close`, Fehler des Accept einstellbar |
| genau eine je Route | `Handler` | Z4, Z7, Z8 | `routeFake`: zählt, hält `Anfrage` und `ctx`, liefert Antwort und Fehler einstellbar |
| keine in diesem Teil | `Audit.Append`, `Telemetrie.CountError` | K3 (Zähler 0) | `FakeAudit`, `FakeTelemetry` zählen jede Operation |

Die Fakes sind test-lokal (*f* = 3: `zaehlendeIDs`, `routeFake`, `listenerFake`); `FakeAudit` und `FakeTelemetry` bestehen. Der Test-Server mit Client (echter Listener) gehört zu `slice-admin-lebenszyklus`.

**Bindung an den Port** (Architect, Auftrag F; Lerneintrag V-2): Die Zusicherungen zur Übersetzungszeit stehen in `src/adapters/driving/admin/bindung_test.go`: `var _ support.AuditLog = (*supporttest.FakeAudit)(nil)`, `var _ support.Telemetry = (*supporttest.FakeTelemetry)(nil)`, `var _ support.IDGenerator = (*zaehlendeIDs)(nil)`, `var _ net.Listener = (*listenerFake)(nil)`, `var _ admin.Handler = (*routeFake).handle`, `var _ http.Handler = (*admin.Server)(nil)`; `Konfig` trägt die Felder als Schnittstellentypen des Ports (nicht als Funktionen), so übersetzt eine Mutation einer Signatur das Paket nicht mehr. Der Adapter ruft die Inbound Ports nicht: er kennt nur `Handler`.

**Größe der Sicht des Architect** (`SPEC-128` Nr. 1, Aufteilung durch den Planner): 13 Zeilen (*a* = 0, *b* = 13), *f* = 3, *p* = 110 (Rahmen etwa 90, Code und Katalog etwa 20; das Register `schaetzformel-ueberschaetzt-den-slice` sagt: Produktcode eher darunter, Tests darüber).

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.
- Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (`BEO-REPO/report-der-rolle-faellt-docs-check`); der Planner reicht den Satz an beide weiter.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-admin-listener-token` (Teil 1, Token-Datei) liegt in `done/`; `slice-audit-log` und `slice-telemetry-port` (Code-Kanten der Felder `Audit` und `Telemetrie` von `Konfig`) liegen in `done/` (geprüft am 2026-11-07, Planner, `main` = `aadbef0`). Die Bedingung an Teil 1 ist eine Reihenfolge, keine Code-Abhängigkeit: beide Teile schreiben in `src/adapters/driving/admin/`, und die Messung nach `SPEC-128` Nr. 4 g zählt jede Änderung an den Pfaden der Marke, auch die eines parallelen Zweigs; die Teile 2 bis 6 laufen deshalb nacheinander. Die Spezifikation (`SPEC-238` bis `SPEC-240`) trägt den Vertrag samt Randformen seit dem Zweig von Teil 1 (`AGENTS.md` §3.12) und liegt mit dessen Merge auf `main`. **Vor dem ersten Code-Commit zieht der Zweig `main` nach** (Planner im Start-Commit): `main` trägt seit dem Merge von `slice-mock-lesen` den Code `ODR-E6008` (38. Code), `ODR-E2102` ist der 39.; vergibt `slice-mock-schreiben` einen weiteren Code, gleicht der zweite Merge die Zahl an (Rückgabe in §6).

**Größenmarke** (Planner, 2026-11-07, `SPEC-128` Nr. 1; Prosa-Form, der Block folgt beim Start): 13 Zeilen der Tabelle, *a* = 0, *b* = 13, *f* = 3, *p* = 110: *S* = 85·0 + 80·13 + 200·3 + 110 = 1750, *M* = 2400, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Pfade des Blocks beim Start: `src/adapters/driving/admin/`, `src/hexagon/model/fehler.go`, `src/hexagon/model/fehler_test.go`, `src/adapters/driven/audit/audit_test.go`, `src/adapters/driven/telemetry/regeln_test.go`, `harness/katalog-ausnahmen.txt`; `produktcode` 110, `fakes` 3. `basis`, `schaetzung` und `marke` setzt der Planner beim Start in einem Commit vor dem ersten Code-Commit.

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
| Zuordnung von Pfad und Methode | Nr. 2: kein Treffer ist 404 | Nr. 2: leeres Segment (`//`) und abschließender `/` passen auf keine Route | Nr. 2: Groß- und Kleinschreibung zählt, Parameter nimmt jedes nichtleere Segment | gilt nicht: reine Funktion ohne Port (Nr. 2) | Nr. 12: zustandslos |

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16): gilt nicht für `New`, weil jeder Verstoß dasselbe `model.ErrUngueltig` ohne Wert ist (`SPEC-238` Nr. 1, 2); ein Paar ändert Art und Code nicht, zugesagt ist nur, dass der Text keinen Wert nennt (K1, K2). Für die Zuordnung hat jeder Pfad genau ein Ergebnis (404 oder 405 oder Treffer); die Reihenfolge gegenüber der Authentisierung (401 vor 404 und 405, `SPEC-238` Nr. 3 d, e) entscheidet `slice-admin-authentisierung` (A5).

- Die Dateien, die dieser Teil mit anderen Zweigen teilt: `spec/spezifikation.md` (nur der Architect, hier nicht geschrieben), `src/hexagon/model/fehler.go`, `fehler_test.go`, `src/adapters/driven/audit/audit_test.go`, `src/adapters/driven/telemetry/regeln_test.go` und `harness/katalog-ausnahmen.txt` (mit `slice-mock-schreiben`, das einen weiteren Code vergeben kann, und früher mit `slice-mock-lesen`, dessen `ODR-E6008` auf `main` liegt) — **Ausgang:** offen; Adresse: Planner (Integrationsschritt: wer als zweiter nach `main` merged, zieht `main` nach und gleicht die Zahl der Codes in `fehler_test.go`, `regeln_test.go` und `audit_test.go` sowie die Zeile in `SPEC-230` Nr. 8 an; Stand heute 37 + `ODR-E6008` + `ODR-E2102` gleich 39).
- Die eingebaute Route `GET /healthz` steht in diesem Teil in der Zuordnung und antwortet 200 mit dem Literal, bevor die Authentisierung existiert; die Zusagen „ohne Token“ und „nur `GET`“ (A3, A4) fängt erst `slice-admin-authentisierung`. Grund: die Zeilen A3 und A4 brauchen das 401 als Gegenprobe. Folge: bis Teil 3 in `done/` liegt, hat die Antwort auf `GET /healthz` keinen Fänger der Ausnahme von der Authentisierung; nichts erreicht den Listener, das Binary bleibt unverändert — **Ausgang:** entschieden (Planner, Schnitt): Fänger in `slice-admin-authentisierung`.
- Die Felder `Audit` und `Telemetrie` sind geprüft und gehalten, aber nicht gerufen; ihre Zusagen (Zählen, Audit des Abweisens) liegen in `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` — **Ausgang:** entschieden (Planner, Schnitt): K3 fängt, dass dieser Teil sie nicht ruft (Zähler 0).
- `make katalog-check` ist rot, sobald `fehler.go` den Code `ODR-E2102` trägt, bis die Zeile in den Ausnahmen steht (`SPEC-201` Nr. 3 e) — **Ausgang:** entschieden (Y3): Code und Zeile im selben Commit.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (12×): jeder Zweig der Zuordnung und von `New` hat einen Lauf unter dem Detektor (Z-Liste in §3) — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): die Fakes `zaehlendeIDs`, `routeFake` und `listenerFake` tragen jede Schnittstelle des Originals und zählen jede Operation — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/dienst-erfuellt-den-port-ohne-bindung` (2×): die Bindung steht in `bindung_test.go` (§3) — **Ausgang:** weiter offen → `BEO-REPO/dienst-erfuellt-den-port-ohne-bindung` im Register.
- Größe: die Marke steht als Prosa in §4 (*S* = 1750, *M* = 2400) — **Ausgang:** offen; Adresse: Planner beim Start (Block mit `basis`, `schaetzung` und `marke` in einem Commit vor dem ersten Code-Commit).
- Gate-Kette (Nr. 7 der Übergangsregeln von `welle-benutzerhandbuch`): der Teil ändert weder `GATE_CHECKS` noch ein Gate; `harness/katalog-ausnahmen.txt` ist Eingabe von `make katalog-check`, kein Gate-Fragment — **Ausgang:** entfallen; kein Pfad der Gate-Kette im Plan (§3).

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

**Vorgelagert — offene Beobachtungen sichten:** Register beim Schnitt nachgezogen (Planner, 2026-11-07, `main` = `aadbef0`, Zähler aus den Dateien unter `evidence/`): einschlägig sind `zweig-ohne-lauf-unter-dem-detektor` 12×, `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `dienst-erfuellt-den-port-ohne-bindung` 2×, `betriebsregel-fuer-repo-dateien-verletzt` 9×, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 44×, `randform-erst-im-review-entschieden` 34×, `menge-im-plan-weicht-vom-orakel-des-tests-ab` 3×, `unbeschraenkte-menge-ohne-benannte-pruefgrenze` 11×, `schaetzformel-ueberschaetzt-den-slice` 21×, `report-der-rolle-faellt-docs-check` 5× (Verkörperung, soweit die Schwelle erreicht ist, beim Lese-Schritt der Welle-Closure); kein neues Verzeichnis, kein Zähler gesetzt.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
