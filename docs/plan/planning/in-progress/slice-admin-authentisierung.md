# Slice slice-admin-authentisierung: Admin-Rahmen, Teil 3 — Ablauf, Authentisierung und Antwort

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-QA-04`, `LH-RB-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Bearer-Token, zeitkonstanter Vergleich, 401 ohne Zustandsänderung) · `LH-QA-04.a` · `SPEC-238` Nr. 3 (Ablauf: a, c, d, e, f), Nr. 4 (Authentisierung), Nr. 5 (Header und Erfolgsantwort) · `SPEC-005` (Zeile `GET /healthz`, Schlusssatz) · `SPEC-064` (`ODR-E2101`) · `SPEC-006` (Fehlerobjekt) · `SPEC-114` Nr. 6 (Korrelations-ID) · `ARC-007`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie. `SPEC-238` steht auf `main` (seit dem Merge von `slice-admin-listener-token`).

**Verantwortlich:** pt9912 (Zweig `slice-admin-authentisierung`, Worktree `/Development/odata-recorder-wt-admin-authentisierung`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912 (Planner). **Datum:** 2026-11-07.

**Herkunft:** Teil 3 des Schnitts von [slice-admin-listener-token](../done/slice-admin-listener-token.md) (siehe dort und [slice-admin-rahmen](../done/slice-admin-rahmen.md) für die Reihenfolge der sechs Teile). Der Plan trägt die Zeilen K4, A1, A3 bis A6, U1 bis U9 und E2 bis E4 der Tabelle des Architect (18 von 102); sie stehen im Wortlaut in §3. Gegenüber dem Vorschlag des Architect liegen A2 (Zeile für den angehaltenen Server) bei `slice-admin-lebenszyklus`, E1 und E5 (Schleifen über die Fehler der Routen) bei `slice-admin-fehlerabbildung` und N1 (gleichzeitige Requests) bei `slice-admin-abweisen-audit`: jede dieser Zeilen braucht Code, den erst der genannte Teil liefert.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Rahmen aus `slice-admin-rahmen` bedient den Ablauf eines Requests nach `SPEC-238` Nr. 3: `GET /healthz` ohne Token (nur `GET`, nur genau dieser Pfad), die Authentisierung vor der Zuordnung (`Authorization: Bearer <token>`, genau eine Zeile, Schema in jeder ASCII-Schreibweise, Kandidat Byte für Byte, `Konfig.Vergleich` mit `vergleiche` auf SHA-256 und `subtle.ConstantTimeCompare`), fehlendes oder falsches Token als 401 `ODR-E2101` mit `WWW-Authenticate: Bearer` ohne Aufruf der Route und ohne Zustandsänderung, den Aufruf des `Handler` genau einmal je gültigem Request und die Antwort nach Nr. 5 (Header `Content-Type`, `Content-Length`, `X-Correlation-Id`, `Allow`, `WWW-Authenticate`; Erfolgsantwort der Route). `New` kopiert das übergebene Token. Das 401 schreibt in diesem Teil noch keinen Audit-Eintrag. **Übernimmt (Sendung von `slice-admin-rahmen`, Verifier-Befund 1, 2026-11-08, `AGENTS.md` §3.13):** den Fänger zu Z8 des Rahmens für `Authorization` mit mehreren Zeilen — die gehaltene `Anfrage` trägt auch dann keine `Authorization`-Zeile, wenn der Request zwei Zeilen führt (Zeile A7).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der angehaltene Server (503 `ODR-E5010` vor der Authentisierung), `Serve`, `Anhalten`, `Schliessen`, Grenzen des `http.Server` — `slice-admin-lebenszyklus`; ohne `Anhalten` hat die Zeile A2 keinen Aufrufer.
- Abbildung der Fehler der Routen, Panik, Zählen (`CountError`) und die Schleifen über die Fehlerobjekte (E1, E5) — `slice-admin-fehlerabbildung`; der Teil schreibt das Fehlerobjekt für 401 über die eine Stelle aus `slice-admin-rahmen`.
- Der Audit-Eintrag zu `ODR-E2101` (Art `admin_abgewiesen`, `SPEC-238` Nr. 8) und die Zeile N1 (gleichzeitige Requests, die `Append` und `CountError` einschließt) — `slice-admin-abweisen-audit`; bis dahin schreibt der Adapter keinen Eintrag (Risiko in §6, nichts erreicht den Listener).
- Konfiguration, Routen, Zuordnung, Code `ODR-E2102` — `slice-admin-rahmen` (Voraussetzung dieses Teils); die Token-Datei — `slice-admin-listener-token` (Teil 1).
- Die Begrenzung von Fehlversuchen — das Lastenheft nennt sie nicht (`LH-FA-13.a`); benannte Lücke, Adresse User (§6), kein Code.
- Die Endpunkte, TLS und das Lesen von `ADMIN_*` — `slice-admin-sessions`, `slice-admin-imports`, `slice-admin-workspaces`, `slice-admin-tls`, `slice-bootstrap-m6`; das Binary bleibt unverändert.

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

- [ ] Ablauf: `KorrelationsID` genau einmal je Request vor allem anderen (A1), `GET /healthz` ohne Token und nur `GET` auf genau diesem Pfad (A3, A4), Authentisierung vor Zuordnung (A5), `Handler` genau einmal je Request mit gültiger Zuordnung und Token (A6); die Schleifen über die Requestarten *R* tragen die Formen dieses Teils (healthz, 14 ungültige Formen von `Authorization`, Route mit 2xx, 404, 405).
- [ ] Authentisierung: `Authorization: Bearer <token>` nach `SPEC-238` Nr. 4 (genau eine Zeile, Schema in jeder ASCII-Schreibweise, Kandidat Byte für Byte), `Vergleich` genau einmal für einen nichtleeren Kandidaten und allein sein Ergebnis entscheidet, `vergleiche` mit `subtle.ConstantTimeCompare` auf SHA-256 und sonst kein Vergleich des Tokens im Paket, die Authentisierung für jede Methode und jeden Pfad außer `GET /healthz`, 401 `ODR-E2101` mit `WWW-Authenticate: Bearer` ohne Aufruf der Route und ohne Zustandsänderung; `New` kopiert das Token (K4).
- [ ] Sendung von `slice-admin-rahmen` (Befund 1 des Verifiers): `Authorization` mit zwei Zeilen (auch gleichlautend) erreicht keine Route, die `Anfrage` der Route trägt keine Zeile davon (Zeile A7).
- [ ] Antwort: `Allow` nur bei 405, `WWW-Authenticate` nur bei 401, `X-Correlation-Id` an jeder Antwort mit der ID des Requests, Erfolgsantwort der Route mit Status, `Content-Type` und `Content-Length` (E2 bis E4); die Zählprüfung der 27 Zellen von E2 erreicht `slice-admin-fehlerabbildung`.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Für diesen Slice: kein neuer Code (`ODR-E2101` steht in `harness/katalog-ausnahmen.txt`); das Binary bleibt unverändert (kein Aufruf von `admin` aus `bootstrap`, `src/bootstrap` ohne Änderung), in §7 vermerkt.
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
| `src/adapters/driving/admin/` | update | Ablauf von `ServeHTTP` (healthz, Authentisierung vor Zuordnung, Aufruf des `Handler`), `vergleiche`, Header der Antwort und Erfolgsantwort (`SPEC-238` Nr. 3 a, c, d, e, f, Nr. 4, Nr. 5) |
| Tests | neu / update | Adapter gegen `ServeHTTP` mit `httptest`; die Fakes `zaehlendeIDs` und `routeFake` aus `slice-admin-rahmen` (der `useCaseFake` hinter der Testroute zählt jede Operation, U8); `Vergleich` als zählende Funktion des Tests; AST-Test für `vergleiche` (U6); der Adapter importiert weder Adapter noch Service (`make a-check`) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stammen vom Architect (2026-11-07, Plan von [slice-admin-listener-token](../done/slice-admin-listener-token.md)); der Schnitt gibt jedem Teil ganze Zeilen im Wortlaut (A Ablauf, U Authentisierung, E Antwort, K Konfiguration). **Wachsende Mengen:** Die Requestarten *R* (A1, E4) und die Antwortarten (E2) wachsen mit den Teilen. *R* dieses Teils: healthz, jede der 14 Formen von U1, 404, 405, Route mit 2xx; `slice-admin-lebenszyklus` ergänzt 503, `slice-admin-fehlerabbildung` die Route mit Fehler und die Route mit Panik. Die Antwortarten von E2 dieses Teils: 401, 404, 405, 200, healthz × die drei Header; 503 (Lebenszyklus) und 400, Fehler der Route, 500 (Fehlerabbildung) kommen dazu, die Zählprüfung 9 × 3 = 27 gilt am Ende. Jeder Teil trägt seine Formen in die Schleifen ein; die Elemente der Vorgänger rechnet seine Zählprüfung mit.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| K4 | (Nr. 1) `New` kopiert das `Token` | Quellslice nach `New` mit 0-Bytes überschrieben: das ursprüngliche Token authentisiert (200), das überschriebene nicht (401) | Kopie entfernt | andere |
| A1 | (Nr. 3 a) `KorrelationsID` wird genau einmal je Request gerufen, vor allem anderen; die anderen Operationen von `IDs` nie | `zaehlendeIDs` (fünf Operationen gezählt); Schleife über Requestarten R (healthz, jede Form von U1, 404, 405, Route 2xx, Route mit Fehler, Route mit Panik, 503): Zähler 1 und 0; Haltepunkt in `KorrelationsID` sieht `Vergleich`, Audit und Route noch bei 0 | ID erst nach der Authentisierung; ein zweiter Aufruf im Fehlerpfad | andere |
| A3 | (Nr. 3 c) `GET /healthz` antwortet 200 mit `{"status":"ok"}` und liest `Authorization` nicht | Schleife über Authorization-Formen (fehlt, leer, falsches Token, richtiges Token, zwei Zeilen, `Basic x`): 200 und Literal; `Vergleich` und Audit 0 | healthz hinter der Authentisierung | andere |
| A4 | (Nr. 3 c) Nur `GET` auf genau `/healthz` ist frei; die Query ist gleichgültig; `/healthz/`, `/HEALTHZ`, `/healthz/x`, `/healthz%2F` und jede andere Methode gehen durch die Authentisierung | Pfade × Methoden (GET und 6 andere): ohne Token 401, außer `GET /healthz` mit Query (keine, `?a=1`, `?`) 200; mit Token 404 oder 405 mit `Allow: GET` | `HasPrefix` statt Gleichheit; Methode nicht geprüft | andere |
| A5 | (Nr. 3 d, e) Die Authentisierung liegt vor der Zuordnung: ohne Token 401 für registrierten und unbekannten Pfad und falsche Methode, mit Token 404 und 405 | Menge von Z6 × (ohne Token, mit Token): Orakel Status | Zuordnung vor der Authentisierung | andere |
| A6 | (Nr. 3 f) Der `Handler` läuft genau einmal je Request mit gültiger Zuordnung und Token | `routeFake` zählt: Schleife über alle Routen des Test-Servers (mindestens 5): 1 | Wiederholung nach einem Fehler | andere |
| A7 | (Nr. 2 `Header` ohne `Authorization`; Sendung von `slice-admin-rahmen`, Z8, Befund 1) Die `Anfrage` der Route trägt keine `Authorization`-Zeile, auch nicht bei zwei Zeilen im Request | Schleife über Zeilen (eine, zwei verschiedene, zwei gleichlautende, zwei mit leerem erstem Wert) in allen Schreibweisen des Schlüssels aus Z8: `routeFake` hält die `Anfrage`; Orakel: kein Schlüssel mit `Authorization` in `Header`, jeder Nachbar bleibt | Filter entfernt nur einzeilige `Authorization`; Slice der Länge 2 nicht gefiltert | andere |
| U1 | (Nr. 4 a, b) Jede ungültige Form von `Authorization` ergibt 401 ohne Aufruf von `Vergleich` und ohne Route | Menge M-Header (14): fehlt; leerer Wert; `Bearer` allein; `Basic <token>`; `Token <token>`; `Bearer<token>`; Schema, Tab, Token; Token ohne Schema; zwei Zeilen beide gültig; zwei Zeilen eine gültig; zwei gleichlautende Zeilen; nur Leerzeichen als Wert; `Bearer Bearer <token>`; `X-Authorization` statt `Authorization`; Zählprüfung 14; Orakel Literal des Vertrags | Prüfung auf genau eine Zeile zu `>= 1`; Schema nicht geprüft | andere |
| U2 | (Nr. 4 b) Das Schema `Bearer` gilt in jeder ASCII-Schreibweise und mit einem oder mehreren Leerzeichen | Schema (`Bearer`, `bearer`, `BEARER`, `bEaReR`) × Leerzeichen (1, 2, 5): 200, Kandidat gleich dem Token | Schema exakt verglichen; nur ein Leerzeichen | andere |
| U3 | (Nr. 4 b) Der Kandidat ist der Rest Byte für Byte: Groß- und Kleinschreibung, innerer Leerraum, Präfix, Suffix und das letzte Byte zählen | Schleife über Abweichungen (Token + `x`; ohne letztes Byte; erstes Byte klein statt groß; Rest mit führendem Tab; Token mit innerem Leerzeichen als erwartetes Token, das gültig ist): Orakel Literale | Kandidat getrimmt oder in Kleinbuchstaben | andere |
| U4 | (Nr. 4 c) `Vergleich(Token, Kandidat)` wird für einen nichtleeren Kandidaten genau einmal gerufen, und allein sein Ergebnis entscheidet | `Vergleich` als zählende Funktion des Tests: Rückgabe `true` bei falschem Kandidaten ergibt 200, `false` bei richtigem 401; die Argumente (erwartet, gegeben) Byte für Byte; ein Aufruf; für M-Header (U1) 0 Aufrufe | der Aufruf durch `==` ersetzt (Zähler 0); die Argumente vertauscht | andere |
| U5 | (Nr. 4 e) `vergleiche` liefert `true` genau bei gleichen Werten | Generator über Abweichungen (gleich; kürzer; länger; gleich lang an erster, mittlerer, letzter Stelle abweichend) × Längen 32, 33, 64; Orakel `string(a) == string(b)` im Test | Rückgabe invertiert; nur die Länge verglichen | andere |
| U6 | (Nr. 4 e) `vergleiche` ruft `subtle.ConstantTimeCompare` auf `sha256.Sum256` beider Werte, und das Paket vergleicht das Token sonst nirgends mit `==`, `bytes.Equal`, `bytes.Compare` oder `strings.EqualFold` | AST-Test (`go/parser`) über die Dateien des Pakets ohne `_test.go`: Aufrufe suchen; Zählprüfung 1 Aufruf von `ConstantTimeCompare`, 2 von `Sum256`, 0 der verbotenen | `bytes.Equal` statt `ConstantTimeCompare`; `Sum256` entfernt | andere |
| U7 | (Nr. 4 d) Die Authentisierung gilt für jede Methode und jeden Pfad außer `GET /healthz` | Methoden (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE) × Pfade (registriert, unbekannt, `/`, `/__admin`, `*` bei OPTIONS) ohne Token: 401; Zählprüfung 8 × 5 | Ausnahme für HEAD oder OPTIONS | andere |
| U8 | (Nr. 4; `LH-FA-13.a`) Ein fehlendes oder falsches Token ergibt 401 ohne Zustandsänderung und ohne Aufruf des Use Case | `routeFake` und der `useCaseFake` hinter der Testroute (zählt jede Operation, Lesen wie Schreiben) nach jeder Form von U1, jeder Zelle von U7 und `Vergleich` gleich `false`: Summe 0 | die Route vor der Prüfung gerufen | andere |
| U9 | (Nr. 4, 5) Das 401 trägt das Fehlerobjekt `ODR-E2101` und `WWW-Authenticate: Bearer` | Fall mit den Literalen `{"error":{"code":"ODR-E2101","message":"Admin authentication missing or invalid","target":"admin"}}`, Status 401 und Header | Header `WWW-Authenticate` fehlt | andere |
| E2 | (Nr. 5) `WWW-Authenticate` nur bei 401, `Allow` nur bei 405, `Connection: close` nur bei 503 | Antwortarten (401, 404, 405, 503, 400, Fehler der Route, 500, 200, healthz) × die drei Header: 27 Zellen, Orakel Literale | `Connection: close` an jeder Fehlerantwort | andere |
| E3 | (Nr. 5) Eine Erfolgsantwort trägt den Status der Route; ein nichtleerer `Body` mit `Content-Type: application/json` und `Content-Length`, ein leerer ohne Body | Fälle 200 mit Body, 201 mit Body, 204 ohne Body, 200 mit leerem Body; Orakel Header | `Content-Type` fehlt; Body bei 204 geschrieben | andere |
| E4 | (Nr. 5; `LH-FA-13.a`; `SPEC-114` Nr. 6) Jede Antwort trägt `X-Correlation-Id` gleich der ID des Requests | Schleife über R (A1): der Header ist die Kennung, die `zaehlendeIDs` für diesen Request vergab; zwei Requests tragen verschiedene | Header fehlt bei Erfolg, healthz oder 503 | andere |

Zeilen gesamt 18: *a* = 0, *b* = 18. In A1 und E2 nennen die Zellen die Endformen des ungeschnittenen Plans; dieser Teil führt die Formen, die sein Code erzeugt (siehe *Wachsende Mengen*), `slice-admin-lebenszyklus` und `slice-admin-fehlerabbildung` ergänzen den Rest.

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, Auftrag C; das Orakel trägt Statuscodes, Codes, Header und Meldungstexte als Literale aus dem Vertrag, nie als Konstanten des Codes):

| Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|---|
| M-Header (U1) | 14 ungültige Formen von `Authorization` (Liste in U1) | Liste im Test | 401, Literal `ODR-E2101`, 0 Aufrufe von `Vergleich`, Route, `useCaseFake` | 14 gleich der Zahl der Formen in U1 und in `SPEC-238` Nr. 4 (a, b) |
| M-Schema (U2) | Schreibweise 4 × Leerzeichen 3 gleich 12 | Schleife | 200, Kandidat gleich dem Token | 12 |
| M-Abweichung (U3, U5) | 5 Abweichungen × Längen 32, 33, 64 | Schleife | `string(a) == string(b)` im Test | 15 |
| M-Methode × Pfad (U7, A4, A5) | Methoden 8 × Pfade 5 (U7); die Menge Methoden 8 × Pfade 4 aus `slice-admin-rahmen` (Z6) für A5 | Schleife | Status und `Allow` als Literale | 40 und 32 |

**Unbeschränkte Menge und ihre Prüfgrenze (akzeptiertes Negativ):** die Länge des Wertes von `Authorization` (Prüfgrenze 32, 33 und 8192 Bytes; Grund: ab 33 Bytes entscheidet allein die Gleichheit; Folge: ein Wert nahe 64 KiB wird von `net/http` mit 431 abgewiesen, nicht vom Adapter).

**Zweige als Menge, Runden unter dem Detektor** (Architect, Auftrag D; Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat null Blöcke mit Zähler 0, der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Ablauf: `GET /healthz`, Authentisierung (fehlt, mehrere, Schema, leer, `Vergleich` falsch, `Vergleich` wahr), 404, 405, Route | A1, A3 bis A6, U1 bis U4 | – |
| `vergleiche` | U5, U6 | – |
| Erfolg leer und nichtleer | E3 | – |

**Aufrufstellen und zählende Fakes** (Architect, Auftrag E; „jeder Fall je Aufrufstelle“ gilt für diese Stellen):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| genau eine, vor allem anderen | `IDs.KorrelationsID` | A1, E4 (Schleife über R) | `zaehlendeIDs`: zählt alle fünf Operationen, hält die vergebene Kennung |
| genau eine, nach der Authentisierung | `Vergleich` | U4 | zählende Funktion des Tests: hält beide Argumente, Rückgabe einstellbar |
| genau eine je Route | `Handler` | A6, U8 | `routeFake`: zählt, hält `Anfrage` und `ctx`; der `useCaseFake` dahinter zählt jede Operation (U8) |
| keine in diesem Teil | `Audit.Append`, `Telemetrie.CountError` | A3 (Audit 0) | `FakeAudit`, `FakeTelemetry` zählen jede Operation, auch die, die der Teil nie rufen darf |

**Zeitkonstanter Vergleich** (Architect, Auftrag H): Haltepunkt am Argument: `Konfig.Vergleich` (U4) belegt, dass der Adapter die Funktion ruft und allein ihr Ergebnis gilt; die Quelle der Funktion belegt der AST-Test U6 (`subtle.ConstantTimeCompare` auf SHA-256, keine verbotenen Vergleiche); das Zeitverhalten selbst misst kein Test (akzeptiertes Negativ in `SPEC-238` *Ungeprüft*, Grund und Folge dort).

**Größe der Sicht des Architect** (`SPEC-128` Nr. 1, Aufteilung durch den Planner): 18 Zeilen (*a* = 0, *b* = 18), *f* = 0 (die Fakes liegen seit `slice-admin-rahmen`), *p* = 100.

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

**Start** (`next` → `in-progress`): [slice-admin-rahmen](../done/slice-admin-rahmen.md) liegt in `done/` (Code-Abhängigkeit: `Konfig`, Zuordnung, die eine Stelle für das Fehlerobjekt, die Fakes `zaehlendeIDs` und `routeFake`). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-238` Nr. 3 bis 5) seit dem Zweig von `slice-admin-listener-token` (`AGENTS.md` §3.12). Stand der Bedingungen beim Schnitt (Planner, 2026-11-07, `main` = `aadbef0`): `slice-audit-log` und `slice-telemetry-port` liegen in `done/`; der Teil ruft beide Ports nicht.

**Größenmarke** (Planner, 2026-11-07, `SPEC-128` Nr. 1; Prosa-Form; der Block mit `basis`, `schaetzung` und `marke` steht vor dem ersten Code-Commit, Start 2026-11-08): 18 Zeilen der Tabelle, *a* = 0, *b* = 18, *f* = 0, *p* = 100: *S* = 85·0 + 80·18 + 200·0 + 100 = 1540, *M* = 2100, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Pfad des Blocks beim Start: `src/adapters/driving/admin/`; `produktcode` 100, `fakes` 0. Start (Planner, 2026-11-08): *S* = 1540 und *M* = 2100 bleiben, die Schnittgrenze 2223 wird nicht erreicht, es wird vor dem Code nicht geschnitten. Stand der Bedingungen beim Start (`main` = `c469ef1e`): `slice-admin-rahmen` liegt in `done/`; `slice-audit-log` und `slice-telemetry-port` liegen in `done/`; in `in-progress/` liegen sonst nur `slice-handbuch-katalog-schluessel-kommandos` (Linie der Gate-Kette) und der parallel startende `slice-bootstrap-m3`; dieser Slice ändert keinen Pfad der Gate-Kette (Übergangsregel 7, §6), also kein Konflikt. Basis der Messung: `c469ef1e` (die Start-Commits ändern allein Pläne).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 2100 (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein Teil, der sich ohne geteilten Mechanismus abtrennen lässt, steht dann vor dem Code in §4, sonst steht die Messung mit Grund im Plan (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert — Carveout?): der Architect ändert `SPEC-238` Nr. 3 oder 4 (Rückgabe an den Planner), oder `slice-admin-rahmen` liefert `Konfig` oder die Zuordnung nicht in der Form, die der Ablauf braucht.

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
| Authentisierung (Header `Authorization`) | `SPEC-238` Nr. 4 a: kein Header ist ein Verstoß, 401 | Nr. 4 b: leerer Wert, `Bearer` ohne Kandidat | Nr. 4 b und c: anderes Schema, mehrere Zeilen, Kandidat ungleich dem Token (`Vergleich` falsch) | Nr. 8: `Append` hat keinen `ctx` und läuft zu Ende; der Client kann trennen, der Eintrag steht (`SPEC-234` Nr. 5 c); der Eintrag gehört zu `slice-admin-abweisen-audit` | Nr. 4 und Nr. 12: jeder Request für sich, kein Zustand; Nr. 8: jede Abweisung ein Eintrag (ebenfalls dort) |
| `GET /healthz` | Nr. 3 c: kein `Authorization` nötig, 200 | Nr. 3 c: leere Query und leerer Header gleichgültig, 200 | Nr. 3 c und d: falsches Token gleichgültig; andere Methode, `/healthz/`, andere Schreibweise gehen durch die Authentisierung | Nr. 3 b und Nr. 9: angehalten ist 503, ein trennender Client lässt den Request enden (beides `slice-admin-lebenszyklus`) | Nr. 3 c und Nr. 12: zustandslos, je Request eine ID, nichts gezählt, nichts auditiert |

**Zusatztabelle: Paare von Verstößen am selben Schlüssel und Modi** (`AGENTS.md` §3.16; die Fänger sind Zeilen der Tabelle in §3):

| Schlüssel und Paar oder Modus | Ergebnis (Ort) | Fänger |
|---|---|---|
| Header: fehlender Header und unbekannter Pfad | 401, nicht 404 (`SPEC-238` Nr. 3 d, e) | A5 |
| Header: falsches Token und nicht vorgesehene Methode | 401, nicht 405 (Nr. 3 d, e) | A5 |
| Header: zwei Zeilen und falsches Schema | 401 (Nr. 4 a); der eine Eintrag gehört zu `slice-admin-abweisen-audit` (D1) | U1 |
| Modus `GET /healthz` ohne Token, mit falschem, mit richtigem Token | 200 in allen drei (Nr. 3 c) | A3 |
| Modus `GET /healthz?x=1` | 200 (Nr. 3 c) | A4 |
| Modus `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS` auf `/healthz` ohne Token | 401 (Nr. 3 c, d) | A4, U7 |
| Modus dieselben Methoden auf `/healthz` mit Token | 405 mit `Allow: GET` (Nr. 3 d, e, Nr. 5) | A4, Z6 (`slice-admin-rahmen`) |
| Modus jede der fünf Methoden einer Route mit Token | der `Handler` der Route, genau einmal (Nr. 3 f) | A6 |

Die Paare „ungültige Form und angehaltener Server“ und „`GET /healthz` bei angehaltenem Server“ gehören zu `slice-admin-lebenszyklus` (A2).

- „Der Vergleich erfolgt zeitkonstant“ (`LH-FA-13.a`) hat keinen Fänger, der aus dem benannten Grund rot wird — **Ausgang:** entschieden (Architect, 2026-11-07, `SPEC-238` Nr. 4 e): `Vergleich` ist ein Feld von `Konfig` (Haltepunkt am Argument, Fänger U4), die Funktion `vergleiche` ruft `subtle.ConstantTimeCompare` auf SHA-256 beider Werte, die Quelle belegt der AST-Test U6; das Zeitverhalten selbst misst kein Test (akzeptiertes Negativ mit Grund und Folge in `SPEC-238` *Ungeprüft*).
- Das 401 schreibt in diesem Teil keinen Audit-Eintrag, obwohl `LH-FA-13.a` einen Eintrag ohne Tokenwert verlangt. Grund: der Eintrag ist Zusage von `slice-admin-abweisen-audit` (Art `admin_abgewiesen` in `support.Arten()`, Zeilen D1 bis D7). Folge: bis dieser Teil in `done/` liegt, ist die Anforderung für das Abweisen nicht belegt; der Listener ist nirgends verdrahtet (kein Composition Root vor `slice-bootstrap-m6`), also entsteht keine erreichbare Abweisung ohne Eintrag — **Ausgang:** entschieden (Planner, Schnitt): Fänger in `slice-admin-abweisen-audit`.
- `LH-FA-13.a` nennt keine Begrenzung von Fehlversuchen; ein vom Betreiber gesetztes Token von nur 32 Zeichen wäre durch Raten angreifbar — **Ausgang:** weiter offen als benannte Lücke des Lastenhefts; Adresse: User (Owner); kein Code, bis entschieden.
- Die Schleifen über *R* und die Antwortarten (A1, E2, E4) wachsen mit den Teilen — **Ausgang:** entschieden (Planner, Schnitt): jeder Teil trägt seine Formen ein (§3, *Wachsende Mengen*); die volle Zählprüfung (27 Zellen in E2) gilt mit `slice-admin-fehlerabbildung`.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (12×): jeder Zweig der Authentisierung hat einen Lauf unter dem Detektor (Z-Liste in §3) — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): die Fakes `routeFake` und `useCaseFake` zählen jede Operation, auch die, die der Rahmen nie rufen darf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/akzeptiertes-negativ-ohne-folge` (15×): das Zeitverhalten des Vergleichs ist ein Negativ mit Grund und Folge in `SPEC-238` *Ungeprüft*, der Haltepunkt ist U4 — **Ausgang:** weiter offen → `BEO-REPO/akzeptiertes-negativ-ohne-folge` im Register.
- Größe: die Marke steht als Prosa in §4 (*S* = 1540, *M* = 2100) — **Ausgang:** offen; Adresse: Planner (Start 2026-11-08: Marke bestätigt, Basis `c469ef1e`; der Block mit `basis`, `schaetzung` und `marke` folgt in einem Commit vor dem ersten Code-Commit).
- Gate-Kette (Nr. 7 der Übergangsregeln von `welle-benutzerhandbuch`): der Teil ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette im Plan (§3).

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

**Vorgelagert — offene Beobachtungen sichten:** Register beim Schnitt nachgezogen (Planner, 2026-11-07, `main` = `aadbef0`, Zähler aus den Dateien unter `evidence/`): einschlägig sind `zweig-ohne-lauf-unter-dem-detektor` 12×, `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `akzeptiertes-negativ-ohne-folge` 15×, `betriebsregel-fuer-repo-dateien-verletzt` 9×, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 44×, `randform-erst-im-review-entschieden` 34×, `menge-im-plan-weicht-vom-orakel-des-tests-ab` 3×, `unbeschraenkte-menge-ohne-benannte-pruefgrenze` 11×, `schaetzformel-ueberschaetzt-den-slice` 21×, `report-der-rolle-faellt-docs-check` 5× (Verkörperung, soweit die Schwelle erreicht ist, beim Lese-Schritt der Welle-Closure); kein neues Verzeichnis, kein Zähler gesetzt.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
