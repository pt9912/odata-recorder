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

**Verantwortlich:** pt9912 (Zweig `slice-admin-fehlerabbildung`, Worktree `/Development/odata-recorder-wt-admin-fehlerabbildung`; der Zweig entsteht nach dem Start-Commit von `main`)

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
- Konfiguration, Zuordnung, Code `ODR-E2102` als Zeile der Tabelle — `slice-admin-rahmen`; Authentisierung und Header — `slice-admin-authentisierung`; `Serve`, `Anhalten`, 503 — `slice-admin-lebenszyklus` (dessen Fall G1 bleibt grün: er fährt die Panik jetzt außerhalb der Route, die der Route fängt der Adapter; sein Fall S13 fährt neben der Panik der Route den Kontextfehler; §3 und §6).
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
| `src/adapters/driving/admin/` | update | Abbildung der Fehler der Routen (`fehler.go`: `rufeHandler` fängt die Panik, `abbilde` bestimmt Code und Status, `finde` durchsucht den Baum der Fehler), Behandlung der Kontextfehler, `CountError` in der einen Stelle, die das Fehlerobjekt schreibt (`schreibeFehler` in `admin.go`, jetzt Methode des `Server`), Prüfung des Status der `Antwort` in `rufe` (`SPEC-238` Nr. 5, 6, 7, 9, 13) |
| Tests | neu / update | Adapter gegen `ServeHTTP` mit `httptest` und gegen einen laufenden Server (Panik, Kontextfehler); `FakeTelemetry` aus `supporttest` (`Aufrufe()` hält Operation und Argument, `Anzahl(op)` zählt `Log`, `StartSpan`, `Snapshot` mit); `routeFake` liefert Antwort und Fehler einstellbar; AST-Test für die eine Stelle (C3), Reflexion für die Typen der `Anfrage` (C4), Kindprozess für stdout und stderr (H2, wie G1); neue Dateien `fehlercodes_test.go`, `fehlerabbildung_test.go`, `schreibstellen_test.go`, `kontextfehler_test.go`, `ausgabe_test.go`; geändert: die Schleifen über *R* und E2 (`authentisierung_test.go`), A2 (`grenzen_test.go`), Fall G1 und S13 von `slice-admin-lebenszyklus` (siehe unten), Z6 (`zuordnung_test.go`); der Adapter importiert weder Adapter noch Service (`make a-check`) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stammen vom Architect (2026-11-07, Plan von [slice-admin-listener-token](../done/slice-admin-listener-token.md)); der Schnitt gibt jedem Teil ganze Zeilen im Wortlaut (B Fehler der Routen, C Zählen, E Antwort, X Abbruch, H Geheimnisse). **Wachsende Mengen:** dieser Teil vervollständigt die Schleifen über die Requestarten *R* (A1, E4 in `slice-admin-authentisierung`) um „Route mit Fehler“ und „Route mit Panik“ und die Antwortarten von E2 um 400, Fehler der Route und 500; damit gilt dort die Zählprüfung 9 × 3 = 27. Umgesetzt: *R* wächst von 27 auf 30 (Route mit `ErrUngueltig` 400, Route mit Fehlercode `ODR-E7003` 404, Route mit Panik 500; die Route mit Fehler ohne Code 500 besteht), die Antwortarten von E2 sind 9 (401, 404, 405, 200, healthz, 500, 503 und die neuen 400 und Fehler der Route mit Code; die Panik ist eine weitere Form der 500) × 3 Header = 27.

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

**Umsetzung, Mutationen, Mengen und Raten** (Implementer, 2026-11-20; `AGENTS.md` §3.10, §3.15, §3.17). Messendes: `sh tools/harness/mutationslauf.sh` mit einer Liste von 46 Mutationen, `MUTATIONSLAUF_SPEICHER=4g`, ein Aufruf unter `flock /tmp/odata-recorder-heavy.lock`, Image `odata-recorder:race`. Ergebnis: 46 `rot`, 0 `rot-paket`, 0 `gruen`, 0 `flatternd`, 0 `uebersetzungsfehler`, Gesamtdauer 76,4 s, je Mutation 0,9 bis 5,9 s (5,9 s die Mutation `S13-verlasse-nicht-deferred`, die auf die Frist von 5 s des Tests wartet); ein erster Lauf derselben Liste auf einem früheren Stand der Tests gab dasselbe Ergebnis (46 `rot`, 80,7 s). Alle Mutationen gehören der Klasse `andere`: **1 Prozess je Mutation** (Stufe 1, der benannte Test; Stufe 2 lief nie, weil jede Mutation in Stufe 1 rot war), keine Rate mit und ohne Detektor; der Plan macht keine Zusage über Nebenläufigkeit (`SPEC-238` Nr. 12 für `CountError` liegt bei N1 in `slice-admin-abweisen-audit`). Last des Hosts am Beginn jeder Mutation (erster Wert von `/proc/loadavg`, `last=l/k` der Ergebniszeile): 4,30 bis 5,79 bei `k` = 4 CPUs des Containers (im ersten Lauf 2,83 bis 3,94); ein Rot unter dieser Last ist kein Rot auf ruhigem Host, die Tests sind deterministisch. Einen Wirkungs-Fänger mit Rate braucht keine Zeile; die Mutationen sind Verhaltens-Mutationen (`SPEC-210` Nr. 12 a). Die Gründe des Rot wurden an sieben Mutationen im Einzellauf gelesen (B3, B8 mit `ErrAbortHandler`, E5 mit Panikwert, G1, H2, S13, X1): jedes ist aus dem benannten Grund rot, nicht weil der Aufbau bricht.

| Zeile | Zusage · Mutation (Name in der Liste) | roter Test |
|---|---|---|
| E1 | drei Felder in fester Reihenfolge ohne Zeilenende, Länge gleich Body · Feldreihenfolge in `FehlerInhalt` vertauscht (`E1-feldreihenfolge`); `\n` angehängt (`E1-zeilenende`); `Content-Length` + 1 (`E1-laenge`) | `TestFehlerobjektFormat` |
| E5 | kein Wert aus Request, Antwort, Ursache, Token · Ursache als Header (`E5-ursache-im-header`); Pfad als Header bei 404 (`E5-pfad-im-header`); Token im `WWW-Authenticate` (`E5-token-im-www`); Panikwert als Body (`E5-panikwert-im-body`) | `TestKeinWertImFehlerobjekt` (Fund nennt den Ort; `TestOrteGetrennt` hält die acht Werte ohne gemeinsames Teilstück von 8 Bytes) |
| B1 | gültiger `*inbound.Fehler` auch verpackt · Unwrap nicht verfolgt (`B1-unwrap`); Äste eines Join nicht verfolgt (`B1-join`); Typ-Zusicherung statt Suche (`B1-typzusicherung`) | `TestGueltigerFehlerVerpackt` |
| B2 | ungültiger Code, Status 0, Status außerhalb 400 bis 599, nicht getragener Status = `ODR-E9001` · `StatusErlaubt` nicht geprüft (`B2-statuserlaubt`); Untergrenze 400 zu 0 (`B2-bereich`) | `TestBruchDesVertrags` |
| B3 | nil-Zeiger zählt nicht · Prüfung des nil-Zeigers entfernt (`B3-nilzeiger`) | `TestNilZeiger` |
| B4 | `ErrUngueltig` ohne Fehler mit Code = `ODR-E2102` mit 400, in jeder Verpackung · `errors.Is` durch `==` (`B4-is-durch-gleich`); Status 404 (`B4-status`) | `TestErrUngueltig` |
| B5 | neben einem `*inbound.Fehler` gilt dieser · Prüfung von (3) vor (2) (`B5-reihenfolge`) | `TestFehlerVorErrUngueltig` |
| B6 | jeder andere Fehler = `ODR-E9001` · Tabelle von `SPEC-109` Nr. 5 nachgebaut, `ErrNichtGefunden` gleich 404 (`B6-tabelle-nachgebaut`) | `TestAndereFehler` |
| B7 | Fehler vor Antwort · die Antwort bevorzugt (`B7-antwort-bevorzugt`) | `TestFehlerVorAntwort` |
| B8 | Panik gefangen, `ErrAbortHandler` eingeschlossen, Server bedient weiter · Panik nicht gefangen (`B8-panik-nicht-gefangen`); `ErrAbortHandler` durchgelassen (`B8-abort-durchgelassen`) | `TestPanikDerRoute` |
| B9 | Status der Antwort 200 bis 299 gilt · obere Grenze `>= 299` (`B9-obergrenze`); untere Grenze 201 (`B9-untergrenze-201`) und 199 (`B9-untergrenze-199`) | `TestStatusDerAntwort` |
| C1 | genau ein `CountError` je Fehlerobjekt an neun Stellen · doppelt im 401-Zweig (`C1-401-doppelt`); nicht bei `ODR-E9001` (Panik, Bruch, anderer Fehler; `C1-E9001-nicht-gezaehlt`), nicht bei 503, 405, 400, nicht bei der Antwort aus einem `*inbound.Fehler` (`C1-503-`, `C1-405-`, `C1-400-nicht-gezaehlt`, `C1-fehler-der-route-nicht-gezaehlt`); `Log` statt `CountError` (`C1-log-statt-counterror`) | `TestZaehlenJedeSchreibstelle` |
| C2 | Erfolg und Kontextfehler zählen nicht · Zählen nach jeder Erfolgsantwort der Route (`C2-erfolg-zaehlt`); nach `GET /healthz` (`C2-healthz-zaehlt`); Kontextfehler zählt (`X1-zaehlt`) | `TestErfolgZaehltNicht`, `TestKontextfehlerOhneAntwort` |
| C3 | genau eine Stelle für `NeuesFehlerobjekt` und eine für `CountError` · zweiter Aufruf von `NeuesFehlerobjekt` im 401-Zweig (`C3-zweites-NeuesFehlerobjekt`); zweite Nennung von `CountError` in `abweise` (`C3-zweites-CountError`) | `TestEineStelleSchreibtUndZaehlt` |
| C4 | die Route erhält keinen Port · `Tele support.Telemetry` als Feld der `Anfrage` (`C4-port-als-feld`) | `TestRouteOhnePort` |
| C5 | `CountError` vor dem Schreiben · Zählen nach dem Schreiben, als `defer` (`C5-zaehlen-nach-dem-schreiben`) | `TestZaehlenVorDemSchreiben` |
| X1 | Kontextfehler ohne Antwort, ohne Zählen, ohne Eintrag · wie ein Fehler mit 500 (`X1-wie-500`); zählt (`X1-zaehlt`); nur direkt gefunden (`X1-nur-direkt`); `DeadlineExceeded` nicht erkannt (`X1-nur-canceled`); neben einem `*inbound.Fehler` gilt der Fehler (`X1-neben-fehler-gilt-der-fehler`) | `TestKontextfehlerOhneAntwort` |
| H2 | kein Log, nichts auf stdout oder stderr · `log.Printf` im 503 (`H2-log-im-503`); `println` im Fehlerpfad (`H2-println-im-fehlerpfad`) und bei der Panik (`H2-println-bei-panik`) | `TestAusgabeLeer` (Kindprozess; `TestAusgabeFaengerProbe` belegt, dass stdout und stderr beim Elternprozess ankommen) |
| G1 (aus `slice-admin-lebenszyklus`) | `ErrorLog` verwirft · `ErrorLog` schreibt in `log.Default()` (`G1-errorlog-nicht-verworfen`), mit der Panik des Generators der Korrelations-ID, also außerhalb der Route | `TestKeinLogDerBibliothek` |
| S13 (aus `slice-admin-lebenszyklus`) | ein Request ohne Antwort verlässt die Zulassung · `verlasse` nicht per `defer` (`S13-verlasse-nicht-deferred`) | `TestHandlerPanikZaehltNichtWeiter` (Panik der Route und Kontextfehler) |

**Mengen** (Generator · Orakel · Zahl der Elemente, die der Test zählt; Zählprüfung gegen die Quelle): *M-Code × Status* (E1, B1, B2): Schleife über die Literal-Tabelle von `SPEC-109` Nr. 2 im Test (41 Zeilen, gleich lang und gleich in Code, Status ohne 0 und Zielbereich wie `model.AlleCodes()`, `TestSpecCodesGegenQuelle`); E1: alle 42 Paare aus Code und Status, die die Tabelle trägt; B2: 44 Codes (41 und 0, -1, letzter + 1) × 19 Status (0, 100, 200, 399, 600, 999 und die 13 verschiedenen Status der Tabelle) = 836 Paare, davon 42 getragen (E1) und 794 Brüche; *M-Err* (B6): `go/parser` über `model` und `ports/outbound`, Orakel 500 und Literal, 29 Variablen im Quelltext (11 von `model`, 18 von `ports/outbound`) gleich 29 Fälle des Tests (Zählprüfung ohne feste Zahl, `TestAndereFehler`), je 5 Verpackungen = 145 Fälle, dazu `errors.New`; die 11 Variablen von `model` sind die echten Werte, für die 18 von `ports/outbound` (kein Adapter darf das Paket importieren, `make a-check`) baut der Test einen Fehler mit dem Text aus dem `errors.New("…")` des Quelltextes; *M-Verpackung* 5 (B1: 5 × 4 Paare = 20 Fälle; B4: 5; X1: 2 Fehler × 7 Formen = 14, die 5 Formen des Plans und zwei weitere, die Ursache eines `*inbound.Fehler` und die Nachbarschaft eines nil-Zeigers); *M-Panik* 7 auf einem laufenden Server; *M-Status* 9; *M-Schreibstellen* 9 (C1, C5, E5 und die Schleife des Kindprozesses von H2); E5: 9 Stellen × 8 Orte = 72 Prüfungen; *R* 30 und die Antwortarten von E2 9 × 3 = 27 (Zählprüfungen im Test). Die Zählprüfung der Aufzählung der neun Stellen gegen `SPEC-238` Nr. 7 ist die Zahl 9 im Test: die Spezifikation liegt im Container der Tests nicht vor.

**Unter `-race -coverprofile`** (Image `odata-recorder:race`, `go test -race -count=1 -coverprofile` über das Paket `admin`, 99,2 % der Anweisungen): drei Blöcke mit Zähler 0, alle außerhalb dieses Slice und ohne Zeile des Plans: `route.go` Funktion `zerlegePfad` (Fehler von `url.PathUnescape`: `EscapedPath` liefert nie ein ungültiges Segment), `tokendatei.go` Funktion `pruefeUndLies` (Fehler von `os.ReadFile` nach bestandener Rechteprüfung) und Funktion `erzeuge` (Fehler beim Schreiben der Hilfsdatei). Der Code dieses Slice (`fehler.go`, `rufe`, `schreibeFehler`, `abweise`) hat keinen Block mit Zähler 0: je Zweig der Tabelle in §3 läuft mindestens ein Test unter dem Detektor (Kontextfehler: 14 Läufe, Panik: 7 Läufe, je Panikwert einer).

**Akzeptierte Negative** (Grund · Folge · ungefangene Klasse mit Maß; `AGENTS.md` §3.15):
- Die 18 Variablen von `ports/outbound` laufen in B6 als Fehler mit ihrem Text, nicht als der Wert der Variablen (Grund: `make a-check` verbietet dem Adapter und seinen Tests den Import des Pakets). Folge: eine Abbildung, die einen dieser Fehler an der Identität erkennt, ist im Adapter nicht schreibbar; eine am Text, wie `B6-tabelle-nachgebaut`, fängt der Test. Ungefangen: 18 von 29 Variablen in der Form „Wert der Variable“, keine in der Form „Text“.
- Die Grenzen 399 und 600 der Prüfung „Status 400 bis 599“ fängt allein `Code.StatusErlaubt`, weil kein Code der Tabelle einen Status außerhalb von 400 bis 599 trägt (außer 0, den die untere Grenze fängt, `B2-bereich` rot). Folge: wächst die Tabelle um einen Code mit Status außerhalb des Bereichs, trägt der Adapter ihn aus, ohne dass ein Test rot wird; `SPEC-109` Nr. 2 ist dann zu ändern. Ungefangen: die Mutation der oberen Grenze (`<= 599` zu `<= 699`), 1 von 2 Grenzen des Bereichs, gleichwertig im heutigen Stand.
- Die Ausgabe auf stdout ist nicht mutiert (die Dateien auf dem Weg eines Requests, `admin.go`, `fehler.go`, `route.go` und `lebenszyklus.go`, importieren weder `os` noch `fmt`; die Mutationen schreiben über `println` und `log`, beide auf stderr). Der Kanal ist über `TestAusgabeFaengerProbe` für stdout und stderr belegt. Ungefangen: 1 von 2 Strömen ohne Mutation am Code des Adapters.
- Ein nil-Zeiger eines anderen Fehlertyps mit einer `Unwrap`, die auf nil zugreift, im Baum des Fehlers: `finde` kennt allein `*inbound.Fehler` (`SPEC-238` *Ungeprüft*: Typed-nil). Folge: eine Panik des Adapters außerhalb der Route, `net/http` schließt die Verbindung ohne Antwort und ohne Zählen. Ungefangen: jeder Typ außer `*inbound.Fehler`, 1 von n.
- `runtime.Goexit` einer Route ist kein Wert einer Panik und nicht Teil von M-Panik: der Aufruf von `Handler` kehrt nie zurück, der Request endet ohne Antwort (wie ein Handler, der nie zurückkehrt, `SPEC-238` *Ungeprüft*). 1 Form, nicht gefangen.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-admin-lebenszyklus](../done/slice-admin-lebenszyklus.md) liegt in `done/` (Code-Abhängigkeit: das 503 als neunte Schreibstelle und der `http.Server`, gegen den X1 die abgebrochene Verbindung beobachtet; über ihn transitiv `slice-admin-authentisierung` und `slice-admin-rahmen`), dazu `slice-telemetry-port` (`CountError`, `FakeTelemetry`) in `done/` (geprüft am 2026-11-07, Planner, `main` = `aadbef0`). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-238` Nr. 5 bis 7, 9, 13, 15) seit dem Zweig von `slice-admin-listener-token` (`AGENTS.md` §3.12).

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect und des Planner, 2026-11-20, vor dem ersten Code-Commit): 18 Zeilen der Tabelle, *a* = 0, *b* = 18, *f* = 0, *p* = 70: *S* = 85·0 + 80·18 + 200·0 + 70 = 1510 gleich der Tabelle (18 Zeilen, alle `andere`), *M* = 100 · ⌈135·1510 / 10000⌉ = 100 · ⌈20,385⌉ = 2100 gleich der Formel; keine Schnittfrage (135·*S* = 203850 ≤ 300000, *S* < 2223), kein `schnitt-grund`. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Pfad des Blocks: `src/adapters/driving/admin/`; `produktcode` 70, `fakes` 0 (alle Fakes bestehen, §3). Basis der Messung: `5b3233b0` (Stand von `main` nach den Moves des Starts; der Commit mit dem Block folgt ihr und ändert keinen Pfad unter `src/`, Messung 0). **Stand der Bedingungen beim Start** (Planner, 2026-11-20, `main` = `222abb0a` vor dem Start): `slice-admin-lebenszyklus` und `slice-telemetry-port` liegen in `done/`; `SPEC-238` Nr. 5 bis 7, 9, 13 und 15 stehen auf `main`; in `in-progress/` läuft sonst kein Slice, kein Worktree; der Slice ändert keinen Pfad der Gate-Kette (Übergangsregel 7, §6), kein Konflikt. Die Sendungen aus §1 sind bei den Nehmern geführt: H1 bei `slice-admin-abweisen-audit` (Zeile H1 und Tabelle der Ausgabewege), Nr. 15 bei `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces` (§1, Geber `slice-admin-listener-token`).

```groessenmarke
pfad: src/adapters/driving/admin/
produktcode: 70
fakes: 0
basis: 5b3233b0
schaetzung: 1510
marke: 2100
```

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

- Der `ErrorLog`-Fall von G1 (`slice-admin-lebenszyklus`): die Panik der Route erreicht `net/http` nicht mehr, weil dieser Teil sie fängt — **Ausgang:** entfallen (Implementer, 2026-11-20): `TestKeinLogDerBibliothek` bleibt grün und fängt `ErrorLog` weiter, weil der Kindprozess die Panik jetzt aus dem Generator der Korrelations-ID wirft, also außerhalb der Route, die der Adapter nicht fängt (`net/http` fängt sie und schreibt `http: panic serving …` an `ErrorLog`); die Mutation `G1-errorlog-nicht-verworfen` ist rot, aus diesem Grund (stderr des Kindprozesses trägt die Zeile). Der Fall bleibt in `slice-admin-lebenszyklus`, geändert sind Szenario und Kopfkommentar des Tests; ebenso fährt S13 (`TestHandlerPanikZaehltNichtWeiter`) die Panik der Route (500) und den Kontextfehler (Abbruch).
- Der Wortlaut von `SPEC-238` Nr. 6 a nennt `errors.As` und `errors.Is`; beide entpacken einen nil-Zeiger `*inbound.Fehler` und greifen dabei auf nil zu (Panik), die Randform „nil-Zeiger im Join“ (B3) ist damit nicht erfüllbar. Der Adapter durchsucht den Baum der Fehler selbst (`finde` in `fehler.go`) und überspringt den nil-Zeiger; was daraus für Nachbarformen folgt, hat der Code aus dem Wortlaut „zählt nicht als dieser Fehler“ gelesen und der Plan prüft es: ein gültiger `*inbound.Fehler` neben einem nil-Zeiger im Join gilt (B3), ein Kontextfehler neben einem nil-Zeiger und als Ursache eines `*inbound.Fehler` bricht ab (X1), ein ungültiger `*inbound.Fehler` neben `ErrUngueltig` ist `ODR-E9001` und nicht `ODR-E2102` (B5, Nr. 6 a 2 vor 3). Vorschlag an den Architect: Nr. 6 a sagt „die Suche folgt `errors.Is` und `errors.As`, entpackt aber keinen nil-Zeiger `*inbound.Fehler`“ und nennt die drei Nachbarformen — **Ausgang:** offen; Adresse: Architect (keine Änderung des Codes verlangt; ohne Klarstellung bleibt das Wort „`errors.As`“ im Vertrag gegen den Code).
- Die Pflichten der Endpunkte (`SPEC-238` Nr. 15) hält allein der Plan der Nehmer (C4 fängt die Seite des Rahmens); dieser Teil hat keine Route — **Ausgang:** entschieden (Planner, 2026-11-07, Rückgabe R3): eingetragen in §1 und DoD von `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces` mit `slice-admin-listener-token` als Geber; der Reviewer jener Slices löst jede Zeile von Nr. 15 mit einer Mutation auf.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (12×): jeder Zweig der Abbildung, jede Panik und der Kontextfehler haben einen Lauf unter dem Detektor (Z-Liste in §3) — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): `routeFake`, `useCaseFake` und `FakeTelemetry` zählen jede Operation; eine Panik der Route ist Gegenstand des Tests (B8), nicht eines Fakes — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (33×) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×): „jede Antwort“, „jeder Fehler mit Code“ und „keiner doppelt“ sind Schleifen über Schreibstellen × Fehlerarten (C1, E5), nicht Listen — **Ausgang:** weiter offen → `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` im Register.
- Größe: der Block in §4 trägt den Teil des Architect (18 Zeilen, *S* = 1510, *M* = 2100 nach `SPEC-128` Nr. 1) und den des Planner (`basis` `5b3233b0`, `schaetzung` 1510, `marke` 2100) — **Ausgang:** entfallen: gesetzt in einem Commit vor dem ersten Code-Commit (Planner, 2026-11-20), `make groessenmarke` grün; keine Schnittfrage. Messung nach dem Code (Implementer, 2026-11-20): `git diff --numstat 5b3233b0..<letzter Commit des Slice>` über `src/adapters/driving/admin/`, Tests eingeschlossen: 1423 Zeilen (1321 Einfügungen, 102 Löschungen), unter der Marke 2100; kein Eintrag *ueberschreitung* nötig.
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
