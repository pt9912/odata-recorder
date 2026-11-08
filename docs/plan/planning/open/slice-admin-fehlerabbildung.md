# Slice slice-admin-fehlerabbildung: Admin-Rahmen, Teil 5 — Fehler der Routen, Zählen und Kontextfehler

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

**Berührte Spec-Stellen:** `SPEC-238` Nr. 5 (Format des Fehlerobjekts), Nr. 6 (Abbildung der Fehler der Routen, `ODR-E2102`), Nr. 7 (Zählen), Nr. 9 (Kontextfehler), Nr. 13 (Geheimnisse, soweit Fehlerobjekt, Telemetrie, Log und Panik), Nr. 15 (Pflichten der Endpunkte) · `SPEC-109` Nr. 1 bis 6 (Fehlerobjekt, Abbildung der Port-Fehler) · `SPEC-006` (Fehlerobjekt) · `SPEC-230` Nr. 8 (Zähler je Fehlercode) · `SPEC-124` *Ungeprüft*, `SPEC-127`, `SPEC-129` (Code der Antwort auf `model.ErrUngueltig`) · `LH-FA-13.a` · `LH-QA-04.a` · `ARC-007`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie. `SPEC-238` steht auf `main` (seit dem Merge von `slice-admin-listener-token`).

**Verantwortlich:** —

**Autor:** pt9912 (Planner). **Datum:** 2026-11-07.

**Herkunft:** Teil 5 des Schnitts von [slice-admin-listener-token](../done/slice-admin-listener-token.md) (siehe dort und [slice-admin-rahmen](../done/slice-admin-rahmen.md) für die Reihenfolge der sechs Teile). Der Plan trägt die Zeilen B1 bis B9, C1 bis C5, E1, E5, X1 und H2 der Tabelle des Architect (18 von 102); sie stehen im Wortlaut in §3. Gegenüber dem Vorschlag des Architect trägt dieser Teil zusätzlich E1 und E5 (Schleifen über die Fehler, die eine Route liefert) und gibt H1 (Schleife über alle acht Ausgabewege) an `slice-admin-abweisen-audit` ab, wo der Weg des Audit-Eintrags und die Fehlertexte von `Serve` und `Schliessen` bestehen.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Was eine Route zurückgibt, wird zur Antwort (`SPEC-238` Nr. 6): ein gültiger `*inbound.Fehler` bestimmt Status und Fehlerobjekt (auch verpackt oder im Join), `model.ErrUngueltig` ohne Code ist `ODR-E2102` mit 400, jeder andere Fehler und jeder Bruch des Vertrags (ungültiger Code, Status außerhalb der Tabelle, nil-Zeiger, Status der `Antwort` außerhalb von 200 bis 299) ist `ODR-E9001` mit 500, eine Panik der Route wird gefangen (500, der Server bedient weiter), ein Kontextfehler ohne Code erzeugt keine Antwort. Die eine Stelle, die das Fehlerobjekt schreibt, ruft `Telemetrie.CountError(code)` genau einmal je Fehlerobjekt mit Code, vor dem Schreiben der Antwort, an neun Schreibstellen (401, 404, 405, 503, `ErrUngueltig`, `*inbound.Fehler`, Panik, Bruch des Vertrags, anderer Fehler); Erfolg und Kontextfehler zählen nicht, die Routen rufen `CountError` nie. Das Fehlerobjekt nennt keinen Wert aus Request, Antwort, Ursache oder Token, und der Adapter schreibt kein Log und nichts auf `stdout` oder `stderr`.

**Übernimmt (Sendungen von `slice-admin-listener-token`, Schnitt vom 2026-11-07, `AGENTS.md` §3.13):** (1) den Code der HTTP-Antwort auf `model.ErrUngueltig` — Sendung von `slice-session-use-case`; Entscheidung des Architect: `ODR-E2102` mit 400 (`SPEC-238` Nr. 6); die Zeile der Codetabelle liefert `slice-admin-rahmen`, die Abbildung dieser Teil; (2) das Zählen der Fehler mit Code für den Admin-Listener (`SPEC-230` Nr. 8) — Sendung von `slice-inspection-use-case`, Rückgabe 4 des Architect vom 2026-11-07: `InspectionUseCase` erhält nie `Telemetry`, die Stelle, die das Fehlerobjekt schreibt, zählt; (3) den Code für einen ungültigen Seitenparameter und ein fehlendes `baseline_id` (Sendung von `slice-inspection-use-case`): `ODR-E2102`, die Route gibt den Fehler zurück (Pflichten der Endpunkte, `SPEC-238` Nr. 15), dieser Teil bildet ihn ab. Die Adresse der Geber bleibt `slice-admin-listener-token`; dessen Teile tragen die Sendungen.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Audit des Abweisens (Art `admin_abgewiesen`, `Audit.Append`) und die Schleife über alle acht Ausgabewege (H1) — `slice-admin-abweisen-audit`; dieser Teil ruft `Audit` nicht (D2-Zähler 0 gilt dort).
- Die Pflichten der Endpunkte (`SPEC-238` Nr. 15: `ErrUngueltig` und Lesefehler des Bodys, Seitenparameter, `baseline_id`, kein `CountError` in der Route, Erfolgsstatus 200 bis 299) — `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces` tragen sie in ihren Plänen; dieser Teil fängt die Seite des Rahmens (C4, B-Zeilen), die Routen haben hier nur Testrouten.
- Konfiguration, Zuordnung, Code `ODR-E2102` als Zeile der Tabelle — `slice-admin-rahmen`; Authentisierung und Header — `slice-admin-authentisierung`; `Serve`, `Anhalten`, 503 — `slice-admin-lebenszyklus` (dessen Fall G1 bleibt grün, wenn die Panik jetzt der Adapter fängt; §6).
- Eigene Codes für die 400 der Endpunkte (`SPEC-047`): ein eigener Code entsteht nicht (`SPEC-238` Nr. 15 a).
- TLS, das Lesen von `ADMIN_*`, die Verdrahtung — `slice-admin-tls`, `slice-bootstrap-m6`; das Binary bleibt unverändert.

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

- [ ] Fehler der Routen: Abbildung nach `SPEC-238` Nr. 6 (gültiger `*inbound.Fehler` auch verpackt, `ErrUngueltig` als `ODR-E2102` mit 400, jeder andere Fehler und jeder Bruch des Vertrags als `ODR-E9001` mit 500, Fehler vor Antwort, Status der `Antwort` 200 bis 299, Panik gefangen), Format des Fehlerobjekts (drei Felder, Reihenfolge, ohne Zeilenende), Kontextfehler ohne Antwort und ohne Zählen, kein Wert aus Request, Antwort, Ursache oder Token im Fehlerobjekt (B1 bis B9, E1, E5, X1); die Schleifen über *R* und die Antwortarten tragen die Formen „Route mit Fehler“, „Route mit Panik“, 400 und 500, die Zählprüfung der 27 Zellen von E2 gilt.
- [ ] Zählen: genau ein `CountError(code)` je Fehlerobjekt mit Code an den neun Schreibstellen, vor dem Schreiben der Antwort, an einer einzigen Stelle des Pakets; Erfolg und Kontextfehler zählen nicht; Route ohne Port (C1 bis C5); der Adapter schreibt kein Log und nichts auf `stdout` oder `stderr` (H2).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Für diesen Slice: kein neuer Code (`ODR-E2102` und `ODR-E9001` bestehen); das Binary bleibt unverändert (kein Aufruf von `admin` aus `bootstrap`, `src/bootstrap` ohne Änderung), in §7 vermerkt.
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
| `src/adapters/driving/admin/` | update | Abbildung der Fehler der Routen, Fangen der Panik, Behandlung der Kontextfehler, `CountError` in der einen Stelle, die das Fehlerobjekt schreibt (`SPEC-238` Nr. 5, 6, 7, 9, 13) |
| Tests | neu / update | Adapter gegen `ServeHTTP` mit `httptest`; `FakeTelemetry` aus `supporttest` (`Aufrufe()` hält Operation und Argument, `Anzahl(op)` zählt `Log`, `StartSpan`, `Snapshot` mit); `routeFake` liefert Antwort und Fehler einstellbar; AST-Tests für die eine Stelle (C3) und die Typen der `Anfrage` (C4); der Adapter importiert weder Adapter noch Service (`make a-check`) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stammen vom Architect (2026-11-07, Plan von [slice-admin-listener-token](../done/slice-admin-listener-token.md)); der Schnitt gibt jedem Teil ganze Zeilen im Wortlaut (B Fehler der Routen, C Zählen, E Antwort, X Abbruch, H Geheimnisse). **Wachsende Mengen:** dieser Teil vervollständigt die Schleifen über die Requestarten *R* (A1, E4 in `slice-admin-authentisierung`) um „Route mit Fehler“ und „Route mit Panik“ und die Antwortarten von E2 um 400, Fehler der Route und 500; damit gilt dort die Zählprüfung 9 × 3 = 27.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| E1 | (Nr. 5) Das Fehlerobjekt hat genau drei Felder in der Reihenfolge `code`, `message`, `target`, ohne Leerraum und Zeilenende, mit `Content-Type: application/json` und `Content-Length` gleich der Länge | Schleife über `model.AlleCodes()` × `Stati()` ohne 0: `routeFake` liefert `*inbound.Fehler`; Orakel die Literal-Tabelle von `SPEC-109` Nr. 2 im Test; Zählprüfung Tabelle und `AlleCodes()` gleich lang und gleich in den Codes | Feldreihenfolge vertauscht; Zeilenende angehängt | andere |
| E5 | (Nr. 5) Das Fehlerobjekt nennt keinen Wert aus Request, Antwort, Ursache oder Token | Schleife über die Schreibstellen W (C1) × Canary-Orte (Pfad, Query, Header-Wert, Body, Ursache des Fehlers der Route, Panikwert): der Canary steht in keinem Byte der Antwort | Ursache als `message` | andere |
| B1 | (Nr. 6 a 2) Ein gültiger `*inbound.Fehler` bestimmt Status und Fehlerobjekt, auch verpackt oder im Join | Schleife über Verpackung (direkt, `%w`, `%w` dreifach, Join vorn, Join hinten) × 3 Codes mit verschiedenen Status (`ODR-E7003`, `ODR-E4001` 502 und 504, `ODR-E6102`) | `errors.As` durch eine Typ-Zusicherung ersetzt | andere |
| B2 | (Nr. 6 a 2) Ein ungültiger Code, Status 0, ein Status außerhalb von 400 bis 599 und ein Status, den der Code nicht trägt, sind ein Bruch des Vertrags: `ODR-E9001` mit 500 | Schleife über `AlleCodes()` plus Codes 0, -1 und letzter + 1 × Status (0, 100, 200, 399, 600, 999 und jeder Status eines anderen Codes, den dieser nicht trägt); Orakel die Literal-Tabelle | `Code.StatusErlaubt` nicht geprüft; Bereich 400 bis 599 nicht geprüft | andere |
| B3 | (Nr. 6 a 2) Ein nil-Zeiger `*inbound.Fehler` zählt nicht als dieser Fehler | Fälle: nackter nil-Zeiger gleich `ODR-E9001`; nil-Zeiger im Join mit `ErrUngueltig` gleich `ODR-E2102` | der nil-Zeiger wie ein gültiger gelesen (Panik) | andere |
| B4 | (Nr. 6 a 3) `model.ErrUngueltig` ohne `*inbound.Fehler` ergibt `ODR-E2102` mit 400, in jeder Verpackung | Schleife über Verpackung (direkt, `%w`, Join vorn, Join hinten) mit einem Text, der einen Canary nennt: Status 400, Literal, der Canary steht nicht in der Antwort | `errors.Is` durch `==` ersetzt | andere |
| B5 | (Nr. 6 a 2, 3) Neben einem `*inbound.Fehler` gilt dieser | Fälle: `Fehler{ODR-E6101, 422}` mit Ursache `ErrUngueltig`; Join von `ErrUngueltig` und `Fehler{ODR-E7003, 404}`: 422 und 404 | Prüfung von (3) vor (2) | andere |
| B6 | (Nr. 6 a 4) Jeder andere Fehler, auch ein Fehler eines Ports ohne Code, ergibt `ODR-E9001` mit 500 | Menge M-Err: jede exportierte Variable `Err…` der Pakete `model` und `ports/outbound` außer `ErrUngueltig`, gefunden mit `go/parser` im Quelltext (Zählprüfung gegen die Quelle, keine feste Zahl), dazu `errors.New("x")`: 500 und Literal | die Tabelle von `SPEC-109` Nr. 5 im Adapter nachgebaut (`ErrNichtGefunden` gleich 404) | andere |
| B7 | (Nr. 6 b) Liefert die Route Fehler und Antwort, gilt der Fehler | `routeFake` liefert `Antwort{200, Body}` und `Fehler{ODR-E7003, 404}`: 404 und das Fehlerobjekt | die Antwort bevorzugt | andere |
| B8 | (Nr. 6 c) Eine Panik der Route ergibt `ODR-E9001` mit 500, der Server bedient weiter, der Wert erreicht nichts | Menge M-Panik (7): string, `error`, `*inbound.Fehler`, nil, `http.ErrAbortHandler`, int, ein Canary; danach ein zweiter Request mit 200; der Canary steht in keinem der acht Wege (H1) | Panik nicht gefangen; `ErrAbortHandler` durchgelassen | andere |
| B9 | (Nr. 6 d) Ein Status der `Antwort` außerhalb von 200 bis 299 ist `ODR-E9001` mit 500; 200 und 299 gelten | Grenzen 0, 199, 200, 299, 300, 399, 400, 500, 600: Orakel | `>= 200 && <= 299` zu `< 299` | andere |
| C1 | (Nr. 7) Genau ein `CountError(code)` je Fehlerobjekt mit Code: die neun Schreibstellen W (401, 404, 405, 503, `ErrUngueltig`, `*inbound.Fehler` der Route, Panik, Bruch des Vertrags, anderer Fehler) | Schleife über W: `FakeTelemetry.Aufrufe()` hat genau einen `CountError` mit dem Code der Antwort, `Anzahl(Log)`, `Anzahl(StartSpan)` und `Anzahl(Snapshot)` 0; Zählprüfung 9 gegen die Aufzählung in `SPEC-238` Nr. 7 | Zählen im 401-Zweig zusätzlich (doppelt); im Panikzweig vergessen | andere |
| C2 | (Nr. 7) Eine Erfolgsantwort und ein Kontextfehler ohne Code zählen nicht | Schleife über healthz, 200, 201, 204 und die Formen von X1: `CountError` 0 | Zählen am Ende jedes Requests | andere |
| C3 | (Nr. 7) Im Paket gibt es genau eine Stelle, die `model.NeuesFehlerobjekt` ruft, und genau eine, die `CountError` ruft | AST-Test: je 1 Aufrufstelle | eine zweite Stelle (Kopie im 401-Zweig) | andere |
| C4 | (Nr. 7, 15) Die Route erhält weder `Telemetry` noch `AuditLog`: `Handler` und `Anfrage` tragen keinen Port | Zusicherung zur Übersetzungszeit der Signatur `func(context.Context, admin.Anfrage) (admin.Antwort, error)`; AST-Test: kein Feld der `Anfrage` hat einen Typ aus `support` | ein Port als Feld der `Anfrage` | andere |
| C5 | (Nr. 7) `CountError` liegt vor dem Schreiben der Antwort | `FakeTelemetry` in einer Hülle des Tests und ein `ResponseWriter` schreiben in dasselbe Ereignisprotokoll: Eintrag `zaehlen` vor jedem `kopf` und `schreiben` | Zählen nach dem Schreiben | andere |
| X1 | (Nr. 9) Ein Kontextfehler der Route (`Canceled`, `DeadlineExceeded`; direkt, verpackt, im Join, neben `*inbound.Fehler`, neben `ErrUngueltig`) ergibt keine Antwort, kein Zählen, keinen Eintrag; die Verbindung bricht ab | Schleife über 2 Fehler × 5 Verpackungen: der Client sieht einen Transportfehler (kein Status); `CountError` und `Append` 0 | Kontextfehler wie ein Fehler mit 500 | andere |
| H2 | (Nr. 13) Der Adapter schreibt kein Log und nichts auf `stdout` oder `stderr` | `log.SetOutput`, `os.Stdout` und `os.Stderr` auf Pipes: nach allen Requestarten R und allen Fehlern 0 Bytes | `log.Printf` im Fehlerpfad | andere |

Zeilen gesamt 18: *a* = 0, *b* = 18. In B8, C1, E5 und H2 nennen die Zellen „acht Wege“ (H1) und „Eintrag“ (`Append`); dieser Teil fängt am Fehlerobjekt, an `FakeTelemetry` und an den Pipes, die Wege Audit-Eintrag und Fehlertexte von `Serve` und `Schliessen` schließt die Schleife H1 in `slice-admin-abweisen-audit`. Die Zelle „keinen Eintrag“ in X1 prüft `FakeAudit.Aufrufe()` gleich 0 (der Teil ruft `Audit` nicht).

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, Auftrag C; das Orakel trägt Statuscodes, Codes, Feldnamen, Header und Meldungstexte als Literale aus dem Vertrag, nie als Konstanten des Codes):

| Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|---|
| M-Code × Status (E1, B1, B2) | `model.AlleCodes()` × `Stati()` ohne 0; dazu Codes 0, -1, letzter + 1 × Status 0, 100, 200, 399, 600, 999 und jeder nicht getragene | Schleife über die Quelle | die Literal-Tabelle von `SPEC-109` Nr. 2 im Test | Literal-Tabelle und `AlleCodes()` gleich lang und gleich in den Codes, keine feste Zahl |
| M-Err (B6) | jede exportierte Variable `Err…` der Pakete `model` und `ports/outbound` außer `ErrUngueltig` | `go/parser` über die Quelle | 500, Literal `ODR-E9001` | Zahl der Variablen im Quelltext gleich Zahl der Fälle, keine feste Zahl |
| M-Verpackung (B1, B4, X1) | direkt, `%w`, `%w` dreifach, Join vorn, Join hinten gleich 5 | Schleife | wie der Fehler ohne Verpackung | 5 |
| M-Panik (B8) | 7 Werte | Liste | 500, Server lebt | 7 |
| M-Status (B9) | 0, 199, 200, 299, 300, 399, 400, 500, 600 gleich 9 | Liste | 200 bis 299 gelten | 9 |
| M-Schreibstellen W (C1, E5) | 401, 404, 405, 503, `ErrUngueltig`, `*inbound.Fehler`, Panik, Bruch des Vertrags, anderer Fehler gleich 9 | Schleife | Code und Status | 9 gleich der Aufzählung in `SPEC-238` Nr. 7 |

**Zweige als Menge, Runden unter dem Detektor** (Architect, Auftrag D; Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat null Blöcke mit Zähler 0, der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Fehler der Route: Kontextfehler, `*inbound.Fehler` gültig und ungültig, nil-Zeiger, `ErrUngueltig`, anderer Fehler, Fehler und Antwort, Status der Antwort, Erfolg leer und nichtleer | B1 bis B7, B9, E3, X1 | Kontextfehler (X1): ein Lauf |
| Panik der Route | B8 | ein Lauf je Panikwert (7) |
| Schreibstelle des Fehlerobjekts, Zählen | C1 bis C5 | – |

**Aufrufstellen und zählende Fakes** (Architect, Auftrag E; „jeder Fall je Aufrufstelle“ gilt für diese Stellen; der Gegenstand der Fehlerabbildung entfällt, weil der Adapter die Tabelle von `SPEC-109` Nr. 5 nicht anwendet, `SPEC-238` Nr. 6 a):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| genau eine, in der Schreibstelle | `Telemetrie.CountError` | C1, C2, C5 | `FakeTelemetry`: `Aufrufe()` hält Operation und Argument, `Anzahl(op)` zählt `Log`, `StartSpan`, `Snapshot` mit |
| genau eine je Route | `Handler` | B1 bis B9, X1 | `routeFake`: zählt, hält `Anfrage` und `ctx`, Antwort und Fehler einstellbar; der `useCaseFake` dahinter zählt jede Operation |
| keine in diesem Teil | `Audit.Append` | X1, C2 | `FakeAudit.Aufrufe()` gleich 0 |

*f* = 0: alle Fakes bestehen seit `slice-admin-rahmen` und `slice-admin-lebenszyklus`.

**Größe der Sicht des Architect** (`SPEC-128` Nr. 1, Aufteilung durch den Planner): 18 Zeilen (*a* = 0, *b* = 18), *f* = 0, *p* = 70.

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

**Start** (`next` → `in-progress`): [slice-admin-lebenszyklus](../in-progress/slice-admin-lebenszyklus.md) liegt in `done/` (Code-Abhängigkeit: das 503 als neunte Schreibstelle und der `http.Server`, gegen den X1 die abgebrochene Verbindung beobachtet; über ihn transitiv `slice-admin-authentisierung` und `slice-admin-rahmen`), dazu `slice-telemetry-port` (`CountError`, `FakeTelemetry`) in `done/` (geprüft am 2026-11-07, Planner, `main` = `aadbef0`). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-238` Nr. 5 bis 7, 9, 13, 15) seit dem Zweig von `slice-admin-listener-token` (`AGENTS.md` §3.12).

**Größenmarke** (Planner, 2026-11-07, `SPEC-128` Nr. 1; Prosa-Form, der Block folgt beim Start): 18 Zeilen der Tabelle, *a* = 0, *b* = 18, *f* = 0, *p* = 70: *S* = 85·0 + 80·18 + 200·0 + 70 = 1510, *M* = 2100, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Pfad des Blocks beim Start: `src/adapters/driving/admin/`; `produktcode` 70, `fakes` 0. `basis`, `schaetzung` und `marke` setzt der Planner beim Start in einem Commit vor dem ersten Code-Commit.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 2100 (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein Teil, der sich ohne geteilten Mechanismus abtrennen lässt, steht dann vor dem Code in §4, sonst steht die Messung mit Grund im Plan (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert — Carveout?): der Architect ändert `SPEC-238` Nr. 6 oder 7 (Rückgabe an den Planner), oder `FakeTelemetry` reicht für die Zählung je Fehlercode nicht (Rückgabe an den Planner; der Fake gehört `slice-telemetry-port` in `done/`).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-07, Zeile dieses Teils): jede Zelle nennt eine Nummer von `SPEC-238` (Rahmen) oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Abbildung von Port-Fehlern und `model.ErrUngueltig` auf Status und Fehlerobjekt | `SPEC-238` Nr. 6 a 2: Status 0 und nil-Zeiger `*inbound.Fehler` sind ein Bruch des Vertrags, `ODR-E9001` | Nr. 5 und 6 d: leerer `Body` der Antwort ist erlaubt; Nr. 6 a 3: `ErrUngueltig` aus einem leeren Wert ist `ODR-E2102` | Nr. 6 a 2 bis 4: Code außerhalb der Tabelle, nicht getragener Status, anderer Fehler: `ODR-E9001`; Nr. 2: unbekannter Pfad 404, nicht vorgesehene Methode 405 (`slice-admin-rahmen`) | Nr. 9: Kontextfehler ohne Antwort; Nr. 6 c: Panik der Route | Nr. 7 und 12: ein Fehler zählt einmal, kein Zustand zwischen Requests |

**Zusatztabelle: Paare von Verstößen am selben Schlüssel und Modi** (`AGENTS.md` §3.16; die Fänger sind Zeilen der Tabelle in §3):

| Schlüssel und Paar oder Modus | Ergebnis (Ort) | Fänger |
|---|---|---|
| Fehler der Route: `*inbound.Fehler` mit Ursache `ErrUngueltig` | der `*inbound.Fehler` gilt (`SPEC-238` Nr. 6 a 2 vor 3) | B5 |
| Fehler der Route: Join von `ErrUngueltig` und `*inbound.Fehler` | der `*inbound.Fehler` gilt (Nr. 6 a 2 vor 3) | B5 |
| Fehler der Route: nil-Zeiger `*inbound.Fehler` im Join mit `ErrUngueltig` | `ODR-E2102` (Nr. 6 a 2, 3) | B3 |
| Fehler der Route: Fehler und `Antwort` zugleich | der Fehler gilt (Nr. 6 b) | B7 |
| Fehler der Route: ungültiger Code und Status zugleich | `ODR-E9001` mit 500 (Nr. 6 a 2) | B2 |
| Fehler der Route: Kontextfehler neben `*inbound.Fehler` oder `ErrUngueltig` | keine Antwort, kein Zählen (Nr. 9) | X1 |
| Fehler der Route: Panik mit einem `*inbound.Fehler` als Wert | `ODR-E9001` mit 500, der Wert erreicht nichts (Nr. 6 c) | B8 |

- Der `ErrorLog`-Fall von G1 (`slice-admin-lebenszyklus`): die Panik der Route erreicht `net/http` nicht mehr, weil dieser Teil sie fängt — **Ausgang:** offen; Adresse: Planner (beim Start dieses Teils: der Implementer hält die Zeile G1 grün und belegt die Mutation „`ErrorLog` nicht verworfen“ mit einem Fehler außerhalb der Route; der Fall bleibt in `slice-admin-lebenszyklus`).
- Die Pflichten der Endpunkte (`SPEC-238` Nr. 15) hält allein der Plan der Nehmer (C4 fängt die Seite des Rahmens); dieser Teil hat keine Route — **Ausgang:** entschieden (Planner, 2026-11-07, Rückgabe R3): eingetragen in §1 und DoD von `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces` mit `slice-admin-listener-token` als Geber; der Reviewer jener Slices löst jede Zeile von Nr. 15 mit einer Mutation auf.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (12×): jeder Zweig der Abbildung, jede Panik und der Kontextfehler haben einen Lauf unter dem Detektor (Z-Liste in §3) — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): `routeFake`, `useCaseFake` und `FakeTelemetry` zählen jede Operation; eine Panik der Route ist Gegenstand des Tests (B8), nicht eines Fakes — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (33×) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×): „jede Antwort“, „jeder Fehler mit Code“ und „keiner doppelt“ sind Schleifen über Schreibstellen × Fehlerarten (C1, E5), nicht Listen — **Ausgang:** weiter offen → `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` im Register.
- Größe: die Marke steht als Prosa in §4 (*S* = 1510, *M* = 2100) — **Ausgang:** offen; Adresse: Planner beim Start (Block mit `basis`, `schaetzung` und `marke` in einem Commit vor dem ersten Code-Commit).
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

**Vorgelagert — offene Beobachtungen sichten:** Register beim Schnitt nachgezogen (Planner, 2026-11-07, `main` = `aadbef0`, Zähler aus den Dateien unter `evidence/`): einschlägig sind `zweig-ohne-lauf-unter-dem-detektor` 12×, `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `fix-deckt-die-genannte-mutation-nicht-die-klasse` 33×, `universalaussage-im-vertrag-ungeprueft` 14×, `betriebsregel-fuer-repo-dateien-verletzt` 9×, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 44×, `randform-erst-im-review-entschieden` 34×, `menge-im-plan-weicht-vom-orakel-des-tests-ab` 3×, `unbeschraenkte-menge-ohne-benannte-pruefgrenze` 11×, `schaetzformel-ueberschaetzt-den-slice` 21×, `report-der-rolle-faellt-docs-check` 5× (Verkörperung, soweit die Schwelle erreicht ist, beim Lese-Schritt der Welle-Closure); kein neues Verzeichnis, kein Zähler gesetzt.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
