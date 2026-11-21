# Slice slice-admin-sessions: Admin-Endpunkte `GET` und `POST /__admin/sessions`, Audit `session_angelegt` (Teil 1 des Schnitts)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-FA-02`, `LH-RB-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-005` (Zeilen `GET` und `POST /__admin/sessions`) · `LH-FA-13.a` (Status, Fehlercodes) · `LH-RB-04.a` · `SPEC-124` (Vertrag von `SessionUseCase`: `List`, `Create`) · `SPEC-066`, `SPEC-100` · `SPEC-109` · `SPEC-292` (Nr. 1 bis 8 und 11 bis 13 für diese zwei Routen; Nr. 9 und 10 gehören Teil 2), neu vom Architect · `SPEC-234` Nr. 2 und 7 (die Art `session_angelegt`) · `SPEC-236` *Ungeprüft* (Verweis) · `SPEC-238` Nr. 6, 9, 15 (Rahmen, nicht geändert) · `ARC-007` · `architecture.md` §2.2, §4.9.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-admin-sessions`, Worktree `/Development/odata-recorder-wt-admin-sessions`; der Zweig besteht seit dem Vertrags-Commit `54090d97`, der Implementer nimmt `main` ab dem Start-Commit per `git merge main` vor dem ersten Code-Commit auf)

**Autor:** pt9912. **Datum:** 2026-10-30; vor dem Code in drei Teile geschnitten am 2026-11-21 (Planner, nach dem Vorschlag des Architect): dieser Teil, [slice-admin-sessions-beenden](../open/slice-admin-sessions-beenden.md) (Teil 2) und [slice-admin-inspektion](../open/slice-admin-inspektion.md) (Teil 3).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Admin-Adapter bildet `GET /__admin/sessions` und `POST /__admin/sessions` aus `SPEC-005` auf `SessionUseCase.List` und `SessionUseCase.Create` ab: Auflisten (bei keiner Session eine leere Liste) und Anlegen mit strengem Body (`service_id` Pflicht, `id` optional); jede Antwort trägt Status, Fehlerobjekt und Korrelations-ID, `POST` schreibt nach dem Erfolg einen Audit-Eintrag `session_angelegt` ohne Geheimnisse. Der Teil legt `SessionRouten` mit zwei Routen an, die Funktion der Zeit (RFC 3339 in UTC mit neun Nachkommastellen) und den zählenden Fake `sessionsFake`, die Teil 2 und Teil 3 nutzen, und hebt die Zählstelle der Arten von `support.Arten()` von 5 auf 6.

**Vertrag (Architect, 2026-11-21):** `SPEC-292` (die Routen, Antworten, Anlegen; Audit für Anlegen und Beenden) und `SPEC-293` (Exchanges, Coverage, Diagnose). Der ungeschnittene Umfang liegt über der Schnittgrenze von `SPEC-128` Nr. 1 (Schätzung 5715 gegen 2223); der Architect schlug in §3 (*Größe und Schnitt*) drei Teile vor, der Planner hat sie vor dem Code vollzogen: dieser Teil trägt `GET` und `POST` (Zeilen S1 bis S18, N1, N2), [slice-admin-sessions-beenden](../open/slice-admin-sessions-beenden.md) trägt `close` und `DELETE` (B1 bis B16, N3, N4), [slice-admin-inspektion](../open/slice-admin-inspektion.md) trägt Exchanges, Coverage und Diagnose (I1 bis I18, N5).

**Übernimmt (Sendung von `slice-session-use-case`, Teil des Schnitts, `AGENTS.md` §3.13):** die Endpunkte `POST` und `GET` für Sessions aus [slice-session-use-case](../done/welle-record-replay/slice-session-use-case.md), deren §1 *Ausdrücklich NICHT* sie der Folgewelle zuweist. `DELETE` und `close` (aus `slice-session-use-case` und `slice-session-beenden`) übernimmt Teil 2.

**Übernimmt (Sendung von `slice-admin-listener-token`, Pflichten der Endpunkte nach `SPEC-238` Nr. 15, Schnitt vom 2026-11-07, `AGENTS.md` §3.13; die zwei Routen dieses Teils, die Routen der Teile 2 und 3 tragen sie in deren §1):** (a) was die Route als formal ungültig erkennt oder von einem Dienst als `model.ErrUngueltig` ohne Code erhält (leerer `service_id`, Form der `id`, Lesefehler des `Body`, auch über 1 MiB), gibt sie als Fehler zurück (`ErrUngueltig` oder `*inbound.Fehler{Code: CodeE2102, Status: 400}`); der Code ist `ODR-E2102` (Entscheidung des Architect, `SPEC-238` Nr. 6 und 15 a), ein eigener Code entsteht nicht; (b) sie ruft `CountError` nie und `Audit.Append` nie für `ODR-E2101` und schreibt kein Fehlerobjekt selbst; (c) die vier Arten von `SPEC-234` Nr. 2 schreibt der Dienst, nie die Route; die Art für Anlegen ist eine Zeile im Vertrag dieses Endpunkts (`SPEC-292` Nr. 11, S12); (d) eine Erfolgsantwort hat einen Status von 200 bis 299 und trägt Daten als `Antwort.Body`. Der Fänger liegt in diesem Slice (die Routen und ihre Tests, S16); der Rahmen hat keine Route (`slice-admin-abweisen-audit`, Zeile P1, akzeptiertes Negativ). Der Reviewer löst jede Zeile von Nr. 15 mit einer Mutation auf.

**Sendung an die Teile 2 und 3 (Geber dieses Teils, Eintrag des Planners vom 2026-11-21 beim Schnitt, `AGENTS.md` §3.13; Nehmer vor dem Geber eingetragen):** (1) *an* [slice-admin-sessions-beenden](../open/slice-admin-sessions-beenden.md) *und* [slice-admin-inspektion](../open/slice-admin-inspektion.md): die Mittel dieses Teils, die jene nutzen oder erweitern, nämlich die Funktion der Zeit (S9, M-Zeit), die JSON-Regeln von `SPEC-292` Nr. 7 und die Abbildung der Fehler (S11), der Konstruktor `SessionRouten` (Teil 2 erweitert ihn auf vier Routen), `sessionsFake` (Teil 2 erweitert ihn um Zustand und Haltepunkte), der AST-Test S16 (seine Dateimenge wächst mit jedem Teil) und die Zählstelle der Arten (6; Teil 2 hebt sie auf 7); (2) *an* [slice-admin-sessions-beenden](../open/slice-admin-sessions-beenden.md): `close` und `DELETE`, der Eintrag `session_beendet` und das Handbuch-Kapitel zu `audit.jsonl`, das `slice-audit-verdrahtung` und `slice-audit-log` diesem Slice zuwiesen (der erste Slice, dessen Binary eine Löschung auslöst, trägt `DELETE`); (3) *an* [slice-admin-inspektion](../open/slice-admin-inspektion.md): die Endpunkte für Exchanges, Coverage und Diagnose und die Antwort auf einen ungültigen Parameter (`limit`, `after`, `baseline_id`; Geber `slice-inspection-use-case` und `slice-inspection-coverage-diagnose`). Die Nehmer haben die Sendungen in §1 und DoD, kein *Ausdrücklich NICHT* dort schließt sie aus.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `POST …/close` und `DELETE …/{id}` samt Audit `session_beendet` — `slice-admin-sessions-beenden` (Teil 2); der Teil erweitert `SessionRouten` und `sessionsFake` und liegt nach diesem Teil, weil er sie im selben Paket braucht.
- Exchanges, Coverage und Diagnose — `slice-admin-inspektion` (Teil 3).
- Imports, Workspaces und Snapshots — `slice-admin-imports` und `slice-admin-workspaces`.
- Der Rahmen (Listener, Token, Fehlerobjekt, Abbildung der Port-Fehler) — `slice-admin-listener-token`; dieser Slice trägt die Routen und ihre Abbildung.
- Statusregel, Rangfolge der Fehler und Bereinigung der Blobs — `SessionUseCase` (`done/`); der Adapter bildet ab und schreibt nie direkt in die Datenbank (Architektur §2).
- Der Port `AuditLog` — `welle-betrieb` (`slice-audit-log`); der Slice ruft ihn.
- Der Audit-Eintrag der Löschung einer Session (Art `session_geloescht`) — `SessionUseCase.Delete` schreibt ihn (`slice-audit-dienste`); der Adapter schreibt die vier Arten von `SPEC-234` Nr. 2 nie, auch nicht für denselben Request (Entscheidung `SPEC-234` Nr. 7, Geber: `slice-audit-log` und `slice-audit-dienste`). Der Adapter auditiert nur eigene Arten (Anlegen hier, Beenden in Teil 2), die der Vertrag des Adapters in die Tabelle einträgt.
- Ein Endpunkt für die Wartung — `SPEC-005` führt keinen; die Startwartung läuft im Composition Root (Ausgang in §6).

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

- [ ] Sessions auflisten und anlegen: `GET` (leere Liste, Kennung und Status) und `POST` (optionale `id`, strenger Body nach `SPEC-292` Nr. 5, `ODR-E7010`, `ODR-E5005`); die Funktion der Zeit, `sessionsFake` und der Konstruktor `SessionRouten` für die Teile 2 und 3; Statuscodes und Form der Antworten nach Vertrag (S1 bis S11, N1, N2).
- [ ] Audit: `POST` schreibt über `AuditLog` genau einen Eintrag `session_angelegt` ohne Geheimnisse, nach dem Erfolg von `Create` und nie bei einem Fehler; `GET` schreibt nie einen; ein Fehler von `Append` ändert die Antwort nicht; die Art ist die sechste Zeile von `support.Arten()` und die Zählstellen der Tests ziehen mit (S12 bis S15, S18).
- [ ] Pflichten der Endpunkte nach `SPEC-238` Nr. 15 (a) bis (d) für `GET` und `POST`: kein `CountError`, kein `Append` für `ODR-E2101`, kein selbst geschriebenes Fehlerobjekt, Erfolgsstatus 200 bis 299, ein formal ungültiger Body ist `ODR-E2102` mit 400, je Route gefangen (S16; Geber `slice-admin-listener-token`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`). Das Kapitel zu `audit.jsonl` (Hinweis aus `slice-audit-verdrahtung`) trägt Teil 2, weil allein `DELETE` eine Löschung auslöst.
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
| `src/adapters/driving/admin/` | update | `SessionRouten` mit `GET` und `POST /__admin/sessions` (Teil 2 erweitert auf vier Routen), der Leser des Bodys, die Funktion der Zeit und der Schreiber des JSON in `sessions.go`, `sessions_body.go` und `json.go`, keine Änderung an `admin.go`, `route.go`, `fehler.go`, `lebenszyklus.go`; `InspektionRouten` liefert Teil 3 |
| `src/hexagon/ports/support/audit.go`, `src/hexagon/ports/support/audit_test.go` | update | die Art `session_angelegt` (`ArtSessionAngelegt`), sechste Zeile von `Arten()` (`SPEC-292` Nr. 11); der Test zählt 5 Arten und wächst auf 6 (Teil 2 hebt auf 7) |
| `src/adapters/driven/audit/audit_test.go`, `src/bootstrap/audit_betrieb_test.go` | update (nur Tests) | Zählstellen der Arten (`[5]bool`, `len(support.Arten()) != 5`) und die Zeile der Art; das Binary bleibt unverändert |
| `spec/spezifikation.md` | geliefert (Architect, auf `main`) | `SPEC-292`, `SPEC-293`; Nachzug `SPEC-005`, `SPEC-234` Nr. 2 und 7, `SPEC-236` *Ungeprüft*, Historie |
| Tests | neu | Adapter gegen Fakes von `SessionUseCase` und `AuditLog`: jedes Argument festgehalten, jede Operation gezählt, auch die, die der Adapter nie rufen darf (`AGENTS.md` §3.15); ein Fall je Aufrufstelle für den `ctx` des Requests (`sessions_fake_test.go`, `sessions_route_test.go`, `sessions_body_test.go`, `sessions_antwort_test.go`, `sessions_audit_test.go`, `sessions_quelle_test.go`; `bindung_test.go` bindet `sessionsFake` an den Port) |
| `src/adapters/driving/admin/abweisen_audit_test.go` | update (nur Test) | `TestAppendEineAufrufstelle` zählt `Append` je Datei: eine Stelle in `admin.go` (Schreibstelle des Fehlerobjekts) und eine in `sessions.go` (Eintrag nach dem Anlegen); Teil 2 trägt `close` in dieselbe Datei und hebt den Zähler |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle steht unten (Architect, 2026-11-21; Spalten nach `SPEC-128` Nr. 3 a). `Nr.` ist eine Kennung und keine Zählung: **S** gehört zu diesem Teil (`GET` und `POST /__admin/sessions`, Audit `session_angelegt`, JSON-Form), **B** zu Teil 2 (`close`, `DELETE`, Audit `session_beendet`), **I** zu Teil 3 (Exchanges, Coverage, Diagnose), **N** sind die nebenläufigen Zeilen (N1, N2 in diesem Teil, N3, N4 in Teil 2, N5 in Teil 3). Der Planner hat den Schnitt am 2026-11-21 vollzogen: jeder Teil trägt ganze Zeilen im Wortlaut des Architect (*Größe und Schnitt* unten); `Zusage` nennt die Nummer des Vertrags (`SPEC-292` oder `SPEC-293`), die Zellen *Fänger* nennen die Menge mit ihrer Zählprüfung.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Zeilen S1 bis S18, N1 und N2; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| S1 | (`SPEC-292` Nr. 1) `SessionRouten` weist `sessions` und `audit` je mit nil ab (`model.ErrUngueltig`, keine Routen, der Text nennt keinen Wert) und liefert die Routen des Teils in der Reihenfolge des Vertrags (Teil 1: `GET` und `POST /__admin/sessions`); es ruft keinen Port und startet keine Goroutine | Schleife über die zwei Felder, daneben der Gutfall mit (Methode, Muster) als Literalen; `sessionsFake` (alle sieben Operationen) und `FakeAudit` zählen 0 Aufrufe; Zahl der Goroutinen vor und nach dem Aufruf gleich; ein Canary aus Bytewerten steht nicht im Text | Prüfung eines Feldes entfernt; zwei Routen vertauscht; `sessions.List` zur Probe gerufen; `go func() {}()` | andere |
| S2 | (Nr. 2) Jede Route ruft genau eine Operation von `sessions` mit dem `ctx` des Handlers (dieselbe Identität) und den Argumenten unverändert, sonst keine: `List` mit `ctx`, `Create` mit `ctx`, `serviceID` und `id`; `Get`, `OpenInitialSession`, `Beenden`, `Delete` und `CloseAll` nie; von `audit` nur `Append` | `sessionsFake` hält `ctx` und jedes Argument und zählt alle sieben Operationen: 2 Routen × 7 Zähler = 14 Zellen, Orakel 1 auf der Diagonale, sonst 0; ein Wert im `ctx` des Requests (`context.WithValue`) kommt beim Fake an, und `ctx` ist nicht abgeleitet (Vergleich auf Gleichheit) | `ctx` mit `context.WithTimeout` abgeleitet; `Get` vor `Create` zur Prüfung gerufen; `serviceID` und `id` vertauscht; `List` zweimal gerufen | andere |
| S3 | (Nr. 1, `SPEC-238` Nr. 2) Zuordnung des Teils: `GET` und `POST` treffen `/__admin/sessions`, jede andere Methode ist 405 mit `Allow: GET, POST` | M-Route: 1 Pfadform × 7 Methoden (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS) = 7 Fälle (2 Treffer, 5 mal 405); Orakel Literale für Status und `Allow`; Zählprüfung: die Pfadformen sind die Muster der gelieferten Routen | `Allow` ohne `POST`; Muster `/__admin/session` | andere |
| S4 | (Nr. 3) `GET` und `POST` lesen die Query nicht: jede Form ist ohne Wirkung; `GET` liest den Body nicht | 6 Queries (leer, `limit=x`, `a=1&a=2`, `%zz`, `%`, `a=1;b=2`) × 2 Routen (`POST` mit gültigem Body) = 12 Fälle, dazu `GET` mit 3 ungültigen Bodies = 3; Orakel: Status 200 und 201 und dieselben Argumente am Fake wie ohne Query; der Body von `GET` zählt jeden Lesezugriff (0); Zählprüfung 15 | Query mit `url.ParseQuery` geprüft und bei Fehler 400; Body von `GET` gelesen | andere |
| S5 | (Nr. 5 a bis f) Gültiger Body: `service_id` und `id` kommen dekodiert, unverändert und ungekürzt an `Create`; fehlender und leerer `id` ergeben `id` gleich `""`; Reihenfolge, Leerraum, Escapes, Nicht-ASCII und `Content-Type` ändern nichts; 1048576 Bytes sind gültig | M-Body gültig: 8 Formen (`{"service_id":"svc"}`, mit `id`, `id` zuerst, `"id":""`, Leerraum davor und danach, Escape im Wert und Nicht-ASCII im Wert (beide aus Bytewerten), Body von genau 1048576 Bytes) und 3 weitere `Content-Type` (keiner, `text/plain`, `application/x-www-form-urlencoded`) zur ersten Form = 11 Fälle; Orakel Literale der Argumente; Zählprüfung 11 | Werte getrimmt; `Content-Type` verlangt; `id` gleich `null` als leer gelesen; Escape nicht aufgelöst; Grenze `>=` statt `>` | andere |
| S6 | (Nr. 5 a bis c, g) Ungültiger Body, Teil Form, Schlüssel, Zeichen: 400 `ODR-E2102`, `Create` 0 Aufrufe, `Append` 0 Aufrufe, kein Wert im Fehlerobjekt | M-Body ungültig, Achsen: Wertform 7 (leer, nur Leerraum, `null`, `[]`, `"x"`, `1`, `true`), Struktur 7 (`{`, nicht geschlossen, Komma vor `}`, einfache Anführungszeichen, ` x` dahinter, `{}` dahinter, Doppelpunkt fehlt), Schlüssel 8 (`{}`, nur `id`, `extra`, `Service_Id`, doppelt mit gleichem Wert, doppelt mit anderem Wert, `id` doppelt, doppelt über einen Escape im Schlüssel (aus Bytewerten)), Zeichen 3 (BOM, ungültiges UTF-8 im Wert, rohes NUL in einer Zeichenkette) = 25 Fälle; der Fake würde `ODR-E5005` liefern; Orakel Literale; Zählprüfung: Summe der Achsen | Prüfung einer der vier Achsen entfernt (je eine Mutation); bei doppeltem Schlüssel gilt der letzte; unbekannter Schlüssel ignoriert; Schreibweise der Schlüssel nicht beachtet | andere |
| S7 | (Nr. 5 a, d, g) Ungültiger Body, Teil Werte, Lesen, Paare: `service_id` leer, `null`, Zahl, Wahrheitswert, Array, Objekt; `id` `null`, Zahl, Wahrheitswert, Array, Objekt; Body mit 1048577 Bytes; Body kürzer als `Content-Length`; jedes Paar zweier Verstoßarten im selben Body ist dieselbe Antwort | Werte 11, Lesen 2 (1048577 Bytes; Länge 100 angekündigt, 10 gesendet), Paare 10 (alle Paare der fünf Arten Schlüssel unbekannt, `service_id`-Wert, `id`-Wert, doppelter Schlüssel, Zeichen) = 23 Fälle; je 400, 0 Aufrufe von `Create` und `Append`; Orakel Literale; Zählprüfung: Summe | `service_id` leer an den Dienst gereicht; Lesefehler als 500; Grenze 1048577 angenommen; nur eine Verstoßart je Body geprüft | andere |
| S8 | (Nr. 6, 7 b) `POST` antwortet 201 mit `{"id":…,"service_id":…,"created_at":…,"status":…}` in dieser Reihenfolge: die vier Felder der Session des Dienstes unverändert (auch ein unbekannter `status`), mit `Content-Type` und `Content-Length`, ohne `Location` | M-Felder: jedes der vier Felder von `model.Session` trägt einen eigenen Wert (Zählprüfung gegen `reflect.TypeOf(model.Session{}).NumField()`), Orakel das Literal des Bodys; `status` `open`, `complete`, `incomplete`, `x`; die Header gleich der Menge von `SPEC-238` Nr. 5 | zwei Felder vertauscht; `service_id` aus dem Body statt aus dem Ergebnis; Status 200; `Location` gesetzt | andere |
| S9 | (Nr. 7 c) Die Zeit ist RFC 3339 in UTC mit genau neun Nachkommastellen, ein anderer Ort wird umgerechnet, die Nullzeit ist `null` | M-Zeit: 7 Werte (`10:00:00.123456789Z`, neun Nullen, `.100000000`, `10:00:00+02:00` gleich `08:00:00Z`, Nullzeit, Jahr 1 mit 1 ns, Jahr 9999) an `created_at` der Antwort von `POST` und eines Eintrags von `GET` = 14 Fälle; Orakel Literale | `time.RFC3339Nano` (kürzt Nullen); keine Umrechnung nach UTC; Nullzeit als `0001-01-01…` | andere |
| S10 | (Nr. 6, 7 d) `GET` antwortet 200 mit `{"sessions":[…]}` in der Reihenfolge des Dienstes, ohne Session `[]`, auch bei `nil` | 0, 1, 3 Sessions (Kennungen in der Reihenfolge `c`, `a`, `b`, nicht sortiert) und `nil` = 4 Fälle; Orakel Literale; Prüfgrenze 3 (N-1) | Liste sortiert; `null` bei `nil`; das Feld heißt `items` | andere |
| S11 | (Nr. 8; `SPEC-238` Nr. 6, 15 a) Jeder Fehler des Dienstes erreicht den Rahmen unverändert: Codes mit Status je Route; `ErrUngueltig` ist 400 `ODR-E2102`; ein Fehler eines Ports ohne Code ist 500 `ODR-E9001` (nicht 404); ein Kontextfehler bricht die Verbindung ab; ein Fehler gilt neben einer Antwort; ein verpackter Fehler mit Code gilt | M-Fehler: je Route 8 allgemeine Ausgänge (`model.ErrUngueltig`, `errors.New`, `model.ErrNichtGefunden`, `context.Canceled`, `context.DeadlineExceeded`, `errors.Join(context.Canceled, *Fehler E7003)`, `fmt.Errorf("x: %w", *Fehler E7003)`, Wert und Fehler E7003 zugleich) und die Codes der Tabelle von Nr. 8 (`GET` 2, `POST` 4): 2 × 8 + 6 = 22 Zellen; Orakel Literale für Status, Code und Body; `FakeTelemetry` zählt je Fehlerantwort genau 1 | Fehler mit `%v` verpackt (verliert den Code); `ErrNichtGefunden` zu 404 übersetzt; Fehler bei vorhandener Antwort verworfen; 404 für `ODR-E7010` hartkodiert | andere |
| S12 | (Nr. 11 a, b, d) Nach einem Erfolg von `Create` ruft `POST` `Append` genau einmal: Art `session_angelegt`, `ServiceID` und `SessionID` aus dem Ergebnis des Dienstes, `KorrelationsID` gleich der ID der Antwort, die übrigen Felder leer, vor der Rückkehr des Handlers | Der Fake vergibt eine andere `id` als der Body nennt; `FakeAudit.Aufrufe()` gleich 1; Eintrag Feld für Feld gegen Literale (alle sieben Felder); Haltepunkt `FakeAudit.Vor`: `Create` ist gelaufen, der Recorder hat nichts geschrieben (kein `WriteHeader`, kein `Write`, kein Header des Fehlerobjekts) | Eintrag aus dem Body; `SessionID` leer; `Append` vor `Create`; zweimal; `ErrorCode` gesetzt; `Append` nach `WriteHeader` | andere |
| S13 | (Nr. 11 c; `SPEC-234` Nr. 7) Kein Eintrag bei jedem anderen Ausgang von `GET` und `POST`: Fehler des Dienstes, Verstoß vor dem Aufruf, Kontextfehler; `GET` schreibt nie einen | M-Audit: Zellen Route × Ausgang für `GET` (1 Erfolg + 8 + 2 Codes = 11) und `POST` (1 + 8 + 4 = 13) = 24; Orakel: 1 Eintrag genau in der Zelle `POST` mit Erfolg, sonst `FakeAudit.Aufrufe()` gleich 0; die 48 Fälle von S6 und S7 zählen mit demselben Orakel mit | Eintrag vor der Prüfung des Fehlers; Eintrag bei `ODR-E7010`; Eintrag bei `GET` | andere |
| S14 | (Nr. 11 e) Ein Fehler von `Append` ändert Status, Body und Header der Antwort nicht | `FakeAudit.Fehler` gesetzt: Antwort Byte für Byte gleich der ohne Fehler; `Append` 1 Aufruf | Fehler von `Append` als 500 | andere |
| S15 | (`SPEC-234` Nr. 2, 4 a und 7, `SPEC-235` Nr. 3; `SPEC-292` Nr. 11) `support.Arten()` führt `session_angelegt` als sechste Zeile mit Pflicht `service_id` und `session_id` und sonst verboten; der Adapter des Ports schreibt `{"timestamp":…,"correlation_id":…,"art":"session_angelegt","service_id":…,"session_id":…}` | `audit_test.go` des Ports und des Adapters: Schleife über die 5 Felder je Art, Orakel Literale; `len(support.Arten())` gleich 6; die Zählstelle in `audit_betrieb_test.go` zieht mit | Pflichtfeld `session_id` entfernt; `error_code` erlaubt; Text der Art `session_angelegt ` mit Leerzeichen | andere |
| S16 | (Nr. 2; `SPEC-238` Nr. 15 b) Die Route zählt nicht, schreibt kein Fehlerobjekt und keinen Eintrag `admin_abgewiesen`: sie ruft `NeuesFehlerobjekt`, `CountError` und `ArtAdminAbgewiesen` nie | AST-Test (`go/parser`) über die Dateien der Routen: 0 Aufrufe der drei Namen, Zählprüfung der drei Namen im Test; Verhalten: ein Fehler `*Fehler{ODR-E2101, 401}` erzeugt genau 1 Eintrag `admin_abgewiesen` und genau 1 `CountError` (der Rahmen), nie 2 | Route ruft `model.NeuesFehlerobjekt`; Route schreibt `ArtAdminAbgewiesen` | andere |
| S17 | (Nr. 14) Der Teil führt keinen Fehlercode, Schlüssel oder Kommando ein, das Binary bleibt unverändert | benannte Ausgabe: `make katalog-check` grün; die Importe des Pakets `bootstrap` nennen `admin` nicht; `git diff --stat` über `src/bootstrap` nennt allein `audit_betrieb_test.go` | `SessionRouten` aus `bootstrap` gerufen | andere |
| S18 | (Nr. 11 Folge) Akzeptiertes Negativ: ein Prozess, der zwischen `Create` und `Append` stirbt, hinterlässt eine Session ohne Eintrag; `OpenInitialSession` beim Start und `CloseAll` beim Herunterfahren schreiben keinen Eintrag | kein Fänger. Grund: kein Test beendet den Prozess dazwischen, und die beiden anderen Stellen sind kein Request des Adapters. Folge: die Einträge fehlen, der Betreiber liest die Sessions mit `GET`. Maß: 2 von 2 Stellen, an denen der Prozess selbst Sessions anlegt und beendet, ohne Eintrag | nicht anwendbar | andere |
| N1 | (Nr. 12) Gleichzeitige Requests derselben Route beeinflussen einander nicht: jede Antwort und jeder Eintrag gehört zum eigenen Request, die Routen halten keinen Zustand | 16 Goroutinen je Route (`GET`, `POST`) zugleich, jede mit eigener `id`, die der Fake in `service_id` zurückgibt; Orakel: Antwort und Eintrag tragen die eigene `id`; Datenrennen: `make test-race`, je Prozess `-test.count=1` | Paketvariable hält die letzte Session, die Antwort liest sie; Eintrag aus einer geteilten Variable | nebenläufig |
| N2 | (Nr. 12) Keine Sperre über einem Aufruf von `sessions` oder `audit`: zwei Requests derselben Route sind zugleich im Dienst | Wirkungs-Fänger: `sessionsFake.Create` kehrt erst zurück, wenn ein zweiter Aufruf eingetreten ist (Kanal); `FakeAudit.Vor` ebenso; die Frist von 10 s ist nur der Abbruch im Fehlschlag | `sync.Mutex` um den Aufruf von `Create`; um `Append` | nebenläufig |

Zeilen: Teil 1 (S1 bis S18, N1, N2) 20 (*a* = 2, *b* = 18). Die Zeilen der Teile 2 (B1 bis B16, N3, N4: 18) und 3 (I1 bis I18, N5: 19) stehen im Wortlaut in [slice-admin-sessions-beenden](../open/slice-admin-sessions-beenden.md) und [slice-admin-inspektion](../open/slice-admin-inspektion.md); ungeschnitten waren es 57 (*a* = 5, *b* = 52).

**Mutationen und Raten** (Implementer, 2026-11-21; `AGENTS.md` §3.10, §3.15, §3.17). Messende: `sh tools/harness/mutationslauf.sh` (`SPEC-210`) unter `flock /tmp/odata-recorder-heavy.lock`, `MUTATIONSLAUF_SPEICHER=4g`, Image `odata-recorder:race`, `r/N` als rote Prozesse von gestarteten je Arm (`-test.count=1`), *Last* ist der erste Wert von `/proc/loadavg` des Hosts zu Beginn der Mutation (Container `--cpus=4`; drei Linien teilen den Host). Die Mutationen mit Test in `src/bootstrap` fuhr eine Kopie des Programms im Scratchpad mit `testZeit` 10 min (die Nullprobe von 2 min trägt die Laufzeit des Pakets nicht); das Werkzeug im Repo blieb unverändert. Der Lauf über die Klasse `andere` (90 Mutationen am Endstand) gab 89 `rot` und 1 `rot-paket` (die Mutation der Wertform „Array mit Paaren“: `TestS6S7UngueltigerBody` bleibt grün, weil `[]` allein am leeren `service_id` scheitert; den Fall führt `TestS6ArrayMitPaaren`, der rot wird). Je Zeile die Mutationen (Auswahl der Liste) und der Test, der aus dem benannten Grund rot wurde:

| Nr. | Mutation | roter Test |
|---|---|---|
| S1 | Prüfung von `sessions` oder `audit` entfernt; Routen vertauscht; `List` zur Probe gerufen; `go func() {}()` und `go func() { select {} }()` | `TestS1Konstruktor`, `TestS1S16QuelleDerRouten` |
| S2 | `context.WithCancel` über `ctx` bei `List` und bei `Create`; `context.Background()`; `Get` vor `Create`; `serviceID` und `id` vertauscht; `List` zweimal; `OpenInitialSession` statt `Create` | `TestS2AufrufeUndKontext` |
| S3 | Muster `/__admin/session`; `POST` als `PUT` | `TestS3Zuordnung` |
| S4 | Query geprüft (`GET`, `POST`); Body von `GET` gelesen | `TestS4QueryUndBody` |
| S5 | `service_id` oder `id` getrimmt; `Content-Type` verlangt (jeder, nur JSON); `id` gleich `null` als leer; `bodyBytes` ±1 (Rahmen); `utf8.Valid` entfernt | `TestS5GueltigerBody`, `TestS5Grenze`, `TestS5ZeichenImWert`, `TestS6S7UngueltigerBody` |
| S6 | Ende nach dem Objekt nicht geprüft; nicht geschlossenes Objekt angenommen; fremder Schlüssel ignoriert; Schreibweise der Schlüssel egal; bei doppeltem Schlüssel gilt der letzte; BOM abgeschnitten; rohes NUL und Tabulator ersetzt; Wertform Array mit Paaren | `TestS6S7UngueltigerBody`, `TestS5Leerraum`, `TestS5ZeichenImWert`, `TestS6ArrayMitPaaren` |
| S7 | `service_id` leer durchgereicht; Typ des Wertes nicht geprüft; Lesefehler als 500; Kontextfehler wie ein Lesefehler | `TestS6S7UngueltigerBody`, `TestS7KontextfehlerBeimLesen` |
| S8 | Felder vertauscht; `service_id` aus dem Body; 200 statt 201; `Location` gesetzt (im Rahmen: `Antwort` trägt keinen Header, die Route kann es nicht); Status geprüft; `created_at` fehlt; Leerraum im JSON | `TestS8FelderDerSession` |
| S9 | `time.RFC3339Nano`; ohne Umrechnung nach UTC; Nullzeit nicht `null`; Nullzeit als Jahr 1; sechs Nachkommastellen | `TestS9Zeit` |
| S10 | Liste umgekehrt; `null` bei `nil`; Feld `items` | `TestS10Liste` |
| S11 | Fehler neu gebildet (`List`, `Create`); Wert gewinnt neben dem Fehler (beide); `ErrNichtGefunden` zu 404; `ODR-E7010` mit 404; Kontextfehler zu 400; Fehler in `errors.Join` verpackt (Identität) | `TestS11S13Fehler`, `TestS11Unveraendert` |
| S12 | Eintrag aus dem Body; `session_id` oder `service_id` leer oder vertauscht; `Append` vor `Create`; zweimal; `error_code`, `workspace_id` oder `import_id` gesetzt; Korrelations-ID leer; falsche Art; ohne Eintrag; `go h.audit.Append` | `TestS12Eintrag` |
| S13 | Eintrag vor der Prüfung des Fehlers; bei `ODR-E7010`; bei `GET`; bei einem Verstoß vor dem Aufruf | `TestS11S13Fehler`, `TestS6S7UngueltigerBody` |
| S14 | Fehler von `Append` als Fehler der Antwort | `TestS14AppendFehler` |
| S15 | `session_id` oder `service_id` aus den Pflichtfeldern; `error_code` erlaubt; Text mit Leerzeichen; Position mit der fünften Zeile getauscht; Zeile entfernt; Zeile des Adapters ohne `session_id`, Korrelations-ID nicht geschrieben (14, Port und Adapter); `Arten()` mit fünf und mit sieben Zeilen, Zelle der Art im Test von `bootstrap` nicht ausgenommen (3, Kopie des Programms) | `TestArtenGleichTabelle`, `TestArtTexte`, `TestArtenLiefertKopie`, `TestFeldBelegung`, `TestZeileSessionAngelegt`, `TestAuditAufrufstellen` |
| S16 | Route ruft `NeuesFehlerobjekt`; nennt `ArtAdminAbgewiesen`; schreibt `admin_abgewiesen`; Status 300; leerer Body; zweite Stelle für `Append` | `TestS1S16QuelleDerRouten`, `TestS16RahmenSchreibtDasAbweisen`, `TestS16ErfolgsstatusUndBody`, `TestAppendEineAufrufstelle` |
| S17 | `bootstrap.go` importiert `admin` | `TestS17BinaryUnveraendert` (`git diff --stat` über `src/bootstrap`: allein `audit_betrieb_test.go`) |
| S18 | akzeptiertes Negativ, kein Fänger (Tabelle) | — |

**Raten der Klasse `nebenläufig`** (`n_detektor` und `n_ohne` Prozesse je Mutation; Fänger von N1 ist der Detektor, die Wirkung ohne Detektor nur bei weitem Fenster):

| Mutation | Test | mit Detektor | ohne Detektor | Last |
|---|---|---|---|---|
| N1: Paketvariable hält die letzte Session, die Antwort liest sie (enges Fenster) | `TestN1GleichzeitigeRequests` | 50/50 | 0/50 | 4,8 |
| N1: dasselbe mit 20000 Lesungen im Fenster | dasselbe | 50/50 | 50/50 | 4,9 |
| N1: Handler der Route teilt `Anfrage` zwischen Requests (eng) | dasselbe | 50/50 | 4/50 | 4,9 |
| N1: dasselbe mit 20000 Lesungen im Fenster | dasselbe | 50/50 | 50/50 | 4,5 |
| N1: `GET` hält die letzte Liste in einer Paketvariable | dasselbe | 50/50 | 0/50 | 1,4 |
| N2: Sperre um `List`, um `Create`, um `Append` (je eine Mutation) | `TestN2KeineSperreUeberDemAufruf` | 10/10, 10/10, 10/10 | 10/10, 10/10, 10/10 | 18,1; 18,2; 11,6 |
| S12: `Append` in einer eigenen Goroutine | `TestS12Eintrag` | 50/50 | 50/50 | 6,5 |
| S1: kurzlebige Goroutine, Fänger Zahl der Goroutinen (nicht der Fänger der Zusage) | `TestS1Konstruktor` | 12/50 | 35/50 | 6,5 |

Vor dem Treffpunkt (alle 16 Requests sind zugleich im Aufruf des Dienstes, bevor einer fortfährt) fing der Detektor die zwei engen Mutationen von N1 nur in 49/50 Prozessen (Last 19 bis 21); der Treffpunkt hebt ihn auf 50/50 (Last 1,4 bis 4,9). Die kurzlebige Goroutine von S1 fängt zuverlässig allein der Test der Quelle (`go`-Anweisung: 0 in `json.go`, `sessions.go`, `sessions_body.go`; rot in 1 von 1 Läufen, deterministisch).

**Mengen mit Generator, Orakel und Zahl** (der Handoff führt sie als Schleife, Zählprüfung gegen die Quelle im Test): M-Route 7; M-Query 15; M-Body 59 (11 gültig, 48 ungültig, je Achse die Zahl des Plans: 7, 7, 8, 3, 11, 2, 10); M-Zeit 14 (7 Werte × 2 Routen); M-Felder 4 und Status 4; M-Fehler 22 (2 × 8 + 6) und M-Audit 24 (2 × 9 + 6); N1 16 je Route. Dazu die Klassen hinter den Zeilen, gebildet aus dem Wortlaut von `SPEC-292` Nr. 5 (kein neuer Zeilenbedarf, *b* bleibt 18, *S* 2110): jeder Bytewert 0 bis 255 vor und hinter dem Objekt (512 Fälle, gültig genau die vier Zeichen des Leerraums: 8) in `TestS5Leerraum`; jedes Steuerzeichen unter 0x20 roh in einer Zeichenkette (32), jedes ASCII-Zeichen von 0x20 bis 0x7f außer Anführungszeichen und Backslash (94), 7 gültige und 6 ungültige UTF-8-Folgen in `TestS5ZeichenImWert`; Arrays, die sich als Paare lesen ließen (3) in `TestS6ArrayMitPaaren`; Kontextfehler beim Lesen (3 Formen) in `TestS7KontextfehlerBeimLesen`; 35 Werte der Zeichen in Antworten (`TestS8ZeichenDerWerte`, Orakel: `encoding/json` liest die Antwort).

**Befund der Messung und Stand der Zweige.** Unter `go test -race -coverprofile` über das Paket `admin` hat der Code dieses Teils (`sessions.go`, `sessions_body.go`, `json.go`) null Blöcke mit Zähler 0; die drei Blöcke mit Zähler 0 sind Altcode (`route.go` in `zerlegePfad`, `tokendatei.go` in `pruefeUndLies` und `erzeuge`); `support`: 100 %. Z-6 hat keinen Ausreißer (der Fehler von `json.Marshal` ist verworfen und hat keine Rückgabe; der von `io.ReadAll` ist ein Zweig von `leseAnlegen`, den S7 führt). Messung gegen die Marke: Summe aus Einfügungen und Löschungen über die Pfade des Blocks gegen `02721bba` am Endstand des Teils 1764 von 2900 (kein fremder Merge in den Pfaden, `ueberschreitung` entfällt). `TestAppendEineAufrufstelle` (Rahmen, `abweisen_audit_test.go`) zählt `Append` nun je Datei: eine Stelle in `admin.go`, eine in `sessions.go`; Teil 2 trägt `close` in `sessions.go` und hebt die Zahl.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12; alle sechs sind entschieden, die Entscheidungen stehen darunter und gelten für alle drei Teile):

1. **OpenAPI-Datei zu `SPEC-005`** (Voraussetzung nach Welle §6): Statuscodes der Erfolge (`200`, `201`, `204`), JSON-Form der Antworten (Felder von Session und Exchange), Fehlerabbildung je Code nach `SPEC-109`.
2. **`close`:** Antwort erst nach `Beenden`; HTTP-Timeouts und Abbruch des Clients (der Kontext des Requests wird abgebrochen, während `Drain` wartet); ob der Adapter den Kontext ablöst oder `SPEC-124` Nr. 11 (Lesen und Statuswechsel unter einem Kontext ohne den Abbruch des Aufrufers) genügt.
3. **Audit:** Eintrag auch bei fehlgeschlagenem Schreibaufruf (etwa `ODR-E7011`)? Felder; Reihenfolge (vor oder nach der Aktion); Verhalten, wenn das Audit selbst scheitert; und wer schreibt was, damit die Löschung einer Session keinen doppelten Eintrag erzeugt (`slice-audit-log` lässt `SessionUseCase` seine Löschungen auditieren).
4. **Parameter:** Paginierung und Coverage aus dem Vertrag `SPEC-236` (Nr. 3, 4, 6; `limit`, `after`, `baseline_id`), die Antwort auf einen ungültigen Parameter (Verweis auf den Code von `slice-admin-listener-token`, Ausnahme nur mit neuem Code samt Katalog und Handbuch), Form der Zahl (Vorzeichen, Leerraum, Überlauf, doppelter Parameter); Form von `{id}` im Pfad (`SPEC-124` Nr. 4), Escapes.
5. **Body von `POST /__admin/sessions`** (`service_id`, `id`): fehlende, leere und unbekannte Felder, falsche Typen, doppelte Schlüssel, Größe, `Content-Type`.
6. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Entscheidungen des Architect zu den Aufträgen 1 bis 6** (2026-11-21, vor dem ersten Code-Commit, `AGENTS.md` §3.12; die Ausgänge in §6 stehen auf *entschieden* mit dem Ort):

1. **Statuscodes, JSON-Form, Fehlerabbildung** (Auftrag 1): `SPEC-292` Nr. 6 bis 8 und `SPEC-293` Nr. 6 bis 9. Erfolg: 201 für `POST`, sonst 200; `DELETE` antwortet 200 mit dem Bericht der Bereinigung (kein 204: der Client sähe sonst nirgends, dass Blobs zurückgestellt sind). Eine eigene OpenAPI-Datei entsteht nicht: der Text der beiden Verträge ist das, gegen das die Tests laufen, und eine zweite Fassung driftete (`SPEC-005` sagt es nun); die Zeile in §6 der Welle („Voraussetzung, nicht Ergebnis“) ist damit für diese Endpunkte erfüllt, der Planner hat sie beim Schnitt nachgezogen.
2. **`close`** (Auftrag 2): `SPEC-292` Nr. 9. Die Antwort kommt erst nach `Beenden`, ohne Frist und ohne Goroutine; der Kontext bleibt der des Requests (keine Ablösung, `SPEC-124` Nr. 11 e genügt); die Dauer ist unbegrenzt, weil der Rahmen kein `WriteTimeout` setzt (`SPEC-238` Nr. 10). Folge im Betrieb: ein Client mit Timeout unter `SHUTDOWN_TIMEOUT`, der trennt, macht die Session `incomplete`. (Teil 2.)
3. **Audit** (Auftrag 3): `SPEC-292` Nr. 11 und `SPEC-234` Nr. 2 und 7. Zwei neue Arten (`session_angelegt`, `session_beendet`), geschrieben nach der Tatsache, nie bei einem Fehler, mit den Feldern aus dem Ergebnis des Dienstes; ein Fehler von `Append` ändert nichts. `DELETE` schreibt nie einen Eintrag: der Dienst schreibt `session_geloescht`, so entsteht keine Doppelung. (`session_angelegt` hier, `session_beendet` in Teil 2.)
4. **Parameter** (Auftrag 4): `SPEC-293` Nr. 3 bis 5. Ungültig sind (alle `ODR-E2102`, 400, ohne Aufruf des Dienstes): eine nicht zerlegbare Query, ein doppelter oder leerer bekannter Parameter, eine Zahl außer ASCII-Ziffern, ein Wert über dem größten `int`, ein fehlendes `baseline_id`. `{id}` geht byte-gleich an den Dienst (`SPEC-292` Nr. 4). Ein neuer Code entsteht nicht. (Teil 3.)
5. **Body von `POST`** (Auftrag 5): `SPEC-292` Nr. 5. Strikt: genau `service_id` (Pflicht, nichtleere Zeichenkette) und `id` (optional, Zeichenkette), kein anderer und kein doppelter Schlüssel, UTF-8 ohne BOM, ein Wert und sonst nichts, höchstens 1 MiB (Rahmen); `Content-Type` wird nicht gelesen.
6. **Tabelle, Raster, Mengen, Nachzug** (Auftrag 6): die Tabelle oben, das Raster und die Zusatztabelle in §6, die Mengen unten. Nachgezogen sind `SPEC-005`, `SPEC-234`, `SPEC-236` und die Historie.
7. **Fehlversuche am Token** (Entscheidung des Users, Welle §5, „Kein Code, bis entschieden“): sie berührt diesen Plan nicht. Sie betrifft allein den Rahmen (`SPEC-238` Nr. 8, den Eintrag `admin_abgewiesen`); die Routen dieses Plans laufen erst nach der Authentisierung, und ihre zwei Arten entstehen nur nach einem Erfolg. Ändert die Entscheidung die Zahl der Arten von `support.Arten()`, ändert sich allein die Zählstelle der Tests (S15, B14).

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect; je Teil getrennt, die Mengen der Teile 2 und 3 stehen in deren Plänen; das Orakel trägt Statuscodes, Codes, Namen und Bytes als Literale aus dem Vertrag, nie als Konstanten des Codes und nie als Aufruf des Lesers):

| Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|---|
| M-Route (S3) | Pfadformen 1 (`/__admin/sessions`) × Methoden 7 (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS) = 7, davon 2 Treffer (über alle Teile 6 Pfadformen, 42 Fälle, 7 Treffer: B1, I3) | Schleife | Status, `Allow` (`GET, POST`) als Literale | Pfadformen gleich der Zahl der verschiedenen Muster der Routen (1), Routen 2 gleich der Länge der Liste des Konstruktors |
| M-Query (S4) | Queries 6 (leer, `limit=x`, `a=1&a=2`, `%zz`, `%`, `a=1;b=2`) × Routen 2 (`GET`, `POST`) = 12 (ohne Wirkung); dazu 3 ungültige Bodies an `GET` = 3 | Schleife | Status und Argumente am Fake als Literale | Summe der Achsen (15) |
| M-Body (S5 bis S7) | gültig 8 + 3 `Content-Type` = 11; ungültig: Wertform 7, Struktur 7, Schlüssel 8, Zeichen 3, Werte 11, Lesen 2, Paare 10 = 48; Achse *Grenze*: 1048576 Bytes gültig, 1048577 ungültig | Generator `bFormen` mit Orakel je Achse | gültig oder 400, Argumente von `Create` als Literale | 11 + 48 = 59, Summe der Achsen |
| M-Zeit (S9) | Werte 7 (siehe S9) an `created_at` der Antwort von `POST` und eines Eintrags von `GET` | Schleife | Literale der Zeichenkette (nie `time.Format` im Orakel) | 7, Zahl der Zeitfelder 1 gegen die Felder `time.Time` von `model.Session` (`reflect`) |
| M-Felder (S8) | Felder von `model.Session` 4, jedes mit eigenem Wert | Schleife über `reflect`, Wert je Feldindex | Namen und Reihenfolge als Literale des Vertrags | `NumField` gegen die Zahl der Literale (4) |
| M-Fehler (S11) | allgemeine Ausgänge 8 × Routen 2 = 16, Codes der Tabelle von `SPEC-292` Nr. 8: `GET` 2, `POST` 4 = 6; zusammen 22 | Schleife | Status, Code und Body als Literale aus `SPEC-109` Nr. 2 | 22 gleich 2 × 8 + 6; die Codes gegen `model.AlleCodes()` (jeder Code der Tabelle ist darin) |
| M-Audit (S13) | Zellen Route × Ausgang: `GET` 11, `POST` 13 = 24; genau 1 mit 1 Eintrag, 23 mit 0 | Schleife über M-Fehler und je einen Erfolg | `FakeAudit.Aufrufe()` und `Eintraege()` als Literale | 24 gleich 2 × 9 + 6; 1 Zelle mit Eintrag |

**Unbeschränkte Mengen und ihre Prüfgrenze (akzeptierte Negative, Grund und Folge):** die Zahl der Sessions in `GET /__admin/sessions` (Prüfgrenze 3; Grund: der Adapter formt jede Session gleich; Folge: der Body wächst linear mit der Zahl der Sessions, `SPEC-005` kennt keine Seitenbildung dafür); die Zahl der gleichzeitigen Requests (Prüfgrenze 16, N1; Folge: die Routen halten keinen Zustand, die Zahl ändert nichts).

**Zweige als Menge, Runden unter dem Detektor** (Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat in den neuen Dateien dieses Teils null Blöcke mit Zähler 0; der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Z-1 Konstruktor: nil-Port, Routen in der Reihenfolge | S1 | ein Lauf |
| Z-2 Leser des Bodys: Lesefehler, UTF-8 und BOM, Wertform, Schlüssel unbekannt und doppelt, Typ je Wert, Pflichtfeld, Rest hinter dem Wert | S5, S6, S7 | ein Lauf |
| Z-4 Schreiber des JSON: Zeit (Nullzeit, anderer Ort), Liste `nil` | S8 bis S10 | ein Lauf |
| Z-5 Handler: Erfolg, Fehler des Dienstes, Audit (Erfolg, Fehler von `Append`) | S11 bis S14 | ein Lauf |
| Z-6 Rückgaben, die nur bei einem Fehler von `json.Marshal` oder `io.ReadAll` über einer Zeichenkette oder Zahl greifen | kein Fänger (unerreichbar) | der einzige erlaubte Ausreißer, mit Datei und Funktion im Bericht |

Nebenläufigkeit: N1 und N2; der Zustand nach den Konstruktoren ist schreibgeschützt.

**Aufrufstellen und zählende Fakes** („jeder Fall je Aufrufstelle“ gilt für diese drei Stellen):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| `List` (`GET /__admin/sessions`) | `SessionUseCase` | S2, S10, S11 | `sessionsFake` zählt jede der sieben Operationen, hält `ctx` und jedes Argument, liefert Ergebnis und Fehler einstellbar, ein Haltepunkt je Operation (für N2 und die Teile 2 und 3), keine Panik |
| `Create` (`POST /__admin/sessions`) | `SessionUseCase` | S2, S5 bis S8, S12 | wie oben |
| `Append` nach `Create` | `AuditLog` | S12 bis S14 | `FakeAudit` (besteht): `Vor`, `Fehler`, `Aufrufe`, `Eintraege` |

Der Fake ist test-lokal (*f* = 1: `sessionsFake`; der Fake `inspektionFake` gehört Teil 3); `FakeAudit`, `FakeTelemetry` und die Fakes der Kennungen bestehen. Ein Test-Server mit echtem Client gehört nicht dazu: die Tests rufen `ServeHTTP` mit `httptest`.

**Bindung an die Ports** (Lerneintrag V-2): `src/adapters/driving/admin/bindung_test.go` ergänzt `var _ inbound.SessionUseCase = (*sessionsFake)(nil)` und `var _ support.AuditLog = (*supporttest.FakeAudit)(nil)` (besteht); `SessionRouten` nimmt die Schnittstellentypen der Ports als Parameter (nicht Funktionen), so übersetzt eine Mutation einer Signatur das Paket nicht mehr.

**Größe und Schnitt** (Architect, 2026-11-21; `SPEC-128` Nr. 1, `AGENTS.md` §3.14). Ungeschnitten: *a* = 5, *b* = 52, *f* = 2, *p* = 730 (Produktcode erwartet: Routen und Handler 270, Leser von Body und Query 140, Schreiber des JSON mit Bericht 190, Arten des Audit 30, Konstruktoren 100), also *S* = 85·5 + 80·52 + 200·2 + 730 = 5715 und *M* = 7800: **Schnittfrage, *S* ≥ 2223**. Der Architect schneidet vor dem Code in drei Teile nach Endpunkten; jeder Teil trägt ganze Zeilen der Tabelle:

| Teil | Vorschlag des Namens | Zeilen | *a* | *b* | *f* | *p* | *S* | *M* |
|---|---|---|---|---|---|---|---|---|
| 1 | `slice-admin-sessions` (dieser Plan) | S1 bis S18, N1, N2 | 2 | 18 | 1 | 300 | 2110 | 2900 |
| 2 | `slice-admin-sessions-beenden` | B1 bis B16, N3, N4 | 2 | 16 | 0 | 130 | 1580 | 2200 |
| 3 | `slice-admin-inspektion` | I1 bis I18, N5 | 1 | 18 | 1 | 300 | 2025 | 2800 |

Die Namen sind Vorschlag, der Planner vergibt sie und legt die zwei neuen Pläne beim Start an (Block, Basis und Marken setzt er, nicht der Architect). *p* je Teil: Teil 1 Routen und Handler 110, Leser des Bodys 60, Schreiber des JSON 70, Arten des Audit 20, Konstruktor 40; Teil 2 Handler 70, Bericht 15, Art des Audit 10, Konstruktor 35; Teil 3 Konstruktor 25, Handler 90, Leser der Query 80, Schreiber 105. Pfade des Blocks je Teil: `src/adapters/driving/admin`, `src/hexagon/ports/support` (nur Teile 1 und 2), `src/adapters/driven/audit` (nur Teile 1 und 2), `src/bootstrap/audit_betrieb_test.go` (nur Teile 1 und 2). **Die Teile teilen einen Mechanismus**: Teil 1 liefert die Funktion der Zeit, `sessionsFake` und den Konstruktor `SessionRouten`, die Teil 2 erweitert, und die Zeit nutzt Teil 3; die Teile schreiben im selben Paket, und die Messung zählt jede Änderung an den Pfaden (`SPEC-128` Nr. 4 g): sie laufen **nacheinander** (1, 2, 3; Teil 3 hängt nur an der Zeit aus Teil 1 und darf vor Teil 2 laufen), nicht parallel. Eine Teilung ohne diesen Mechanismus gibt es nicht; die Reihenfolge ist keine Code-Abhängigkeit im Sinne der Wellenpläne (wie bei `slice-admin-listener-token`). Auslastung der Teile: die Messpunkte der letzten Teile liegen bei 0,41 bis 0,74 der Marke (Register `schaetzformel-ueberschaetzt-den-slice`); die Faktoren bleiben, wie sie sind. Katalog und Handbuch: kein Teil bringt einen Code, Schlüssel oder ein Kommando in den Katalog; in §7 steht, dass das Binary unverändert blieb.

**Geprüft und geschnitten vom Planner (2026-11-21):** Die Zahlen des Architect sind aus der Tabelle nach `SPEC-128` Nr. 1 nachgerechnet und stimmen: ungeschnitten 57 Zeilen (*a* = 5, *b* = 52), *S* = 425 + 4160 + 400 + 730 = 5715, 135·5715 = 771525, *M* = 7800. Teil 1: 20 Zeilen, *S* = 85·2 + 80·18 + 200·1 + 300 = 2110, 135·2110 = 284850, *M* = 100 · ⌈28,485⌉ = 2900. Teil 2: 18 Zeilen, *S* = 85·2 + 80·16 + 0 + 130 = 1580, 135·1580 = 213300, *M* = 2200. Teil 3: 19 Zeilen, *S* = 85·1 + 80·18 + 200·1 + 300 = 2025, 135·2025 = 273375, *M* = 2800. Die Summe der Teile (*S* 5715) gleicht dem ungeschnittenen *S*, die Zeilen 20 + 18 + 19 gleichen 57, *p* 300 + 130 + 300 gleicht 730 (die Zerlegung des Architect nach Routen und Handlern 110 + 70 + 90 = 270, Leser 60 + 80 = 140, Schreiber 70 + 15 + 105 = 190, Arten 20 + 10 = 30, Konstruktoren 40 + 35 + 25 = 100). Alle drei Teile liegen unter der Schnittgrenze (*S* < 2223, 135·*S* ≤ 300000), kein `schnitt-grund` nötig. **Abstand zur Schnittgrenze:** dieser Teil liegt 113 darunter (2223 − 2110), jede weitere Zeile der Tabelle kostet 80 und jede Erhöhung des Produktcodes 1: **zwei weitere Zeilen** (*S* = 2270) lösen die Schnittfrage neu aus, eine nicht (*S* = 2190); Teil 3 liegt 198 darunter (drei weitere Zeilen, *S* = 2265), Teil 2 liegt 643 darunter (neun weitere Zeilen, *S* = 2300). Wächst die Tabelle im Code (`AGENTS.md` §3.9), prüft der Planner sie neu; der Implementer meldet jede weitere Zeile in §3. Die Zeilen sind im Wortlaut verteilt (die Verweise zwischen den Teilen: B4, B5, B9 und B14 lesen S4, S8, S11 und S15 in diesem Plan, I1 und I18 lesen S1 und S17); Mengen, Raster und Zusatztabelle sind je Teil getrennt, die Gesamtzahlen über alle Teile (I3: 42 Fälle, I17: 84 Zellen) regelt §1 der Teile 2 und 3. Die Reihenfolge (1, 2, 3; Teil 3 darf vor Teil 2) und die Pfade des Blocks sind in §4 festgelegt.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* die Endpunkte aus `slice-session-use-case` und `slice-session-beenden` (beide `done/`): angenommen, §1 *Übernimmt* trägt sie mit der Kennung des Gebers; `slice-session-wartung` nennt „Admin-Endpunkte für Sessions und Wartung“, `SPEC-005` kennt für die Wartung keinen (§6). *Von diesem Slice an:* keine. *Vom Planner am 2026-11-07 nachgetragen* (Rückgabe 3 des Architect von `slice-inspection-use-case`, gelesen: dessen §1 *Ausdrücklich NICHT*, DoD und §4 sowie die von `slice-admin-listener-token`): die Parameter `limit`, `after`, `baseline_id` und die Antwort auf einen ungültigen Parameter; `slice-inspection-use-case` schließt die Endpunkte aus (§1 *Ausdrücklich NICHT*, Punkt 1) und nimmt nichts davon; der Code für `model.ErrUngueltig` liegt bei `slice-admin-listener-token`. *Beim Schnitt am 2026-11-21 gelesen* (§1 *Ausdrücklich NICHT*, DoD und §4 der Nehmer): an Teil 2 die Endpunkte `close` und `DELETE`, der Eintrag `session_beendet`, das Handbuch-Kapitel zu `audit.jsonl` (Geber `slice-audit-verdrahtung`, `slice-audit-log`) und die Mittel dieses Teils; an Teil 3 die Endpunkte für Exchanges, Coverage und Diagnose samt Parametern und die Zeit dieses Teils; die Pflichten von `SPEC-238` Nr. 15 (Geber `slice-admin-listener-token`) tragen alle drei Teile für ihre Routen; alle Sendungen stehen im selben Commit in §1 und DoD der Nehmer, kein *Ausdrücklich NICHT* dort schließt eine aus. `slice-admin-imports` und `slice-admin-workspaces` tragen die Pflichten von Nr. 15 für ihre Routen in §1 und DoD (gelesen: beide Pläne); ihr Vertrag (Tabelle, Raster, Status und Form der Antworten, Aufträge an den Architect) steht aus, der Planner startet sie nicht vor dem Architect.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-admin-listener-token` samt seinen Teilen `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` (Schnitt vor dem Code, 2026-11-07; der Rahmen ist erst mit dem letzten Teil vollständig), `slice-audit-log` und `slice-audit-dienste` (`welle-betrieb`) liegen in `done/` (`slice-audit-dienste`, weil der Eintrag der Löschung erst dort im Use Case entsteht und die DoD „ohne Doppelung“ verlangt). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): `SPEC-292` liegt vor (Architect, 2026-11-21). `slice-inspection-use-case` und `slice-inspection-coverage-diagnose` sind keine Bedingung dieses Teils (sie gehören Teil 3).

**Stand der Bedingungen beim Start** (Planner, 2026-11-21, `main` = `02721bba` nach den Moves): `slice-admin-listener-token` und seine fünf Teile (`slice-admin-rahmen`, `-authentisierung`, `-lebenszyklus`, `-fehlerabbildung`, `-abweisen-audit`), `slice-audit-log`, `slice-audit-dienste`, `slice-session-use-case` und `slice-session-beenden` liegen in `done/` (geprüft an den Dateien, die zwei letzten unter `done/welle-record-replay/`); `SPEC-292` und `SPEC-293` samt Tabelle, Raster und Mengen und der Nachzug von `SPEC-005`, `SPEC-234` Nr. 2 und 7 und `SPEC-236` *Ungeprüft* stehen auf `main` (Merge `a791be3b`). Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`. In `in-progress/` läuft sonst allein `slice-listener-tls-verdrahtung` (Worktree `/Development/odata-recorder-wt-listener-tls-verdrahtung`; Pfade: `src/bootstrap/`, `src/adapters/driving/httpodata/`, `src/hexagon/ports/support/tlstest/`): keine Datei gemeinsam; dieser Teil ändert in `src/bootstrap/` allein `audit_betrieb_test.go` (wenige Zeilen: Zählstelle der Arten und Zeile der Art), die der Block jenes Slice mit dem Verzeichnis als Pfad als fremde Zeilen mitzählt (`SPEC-128` Nr. 10 g; der Implementer jenes Slice misst vor dem Merge und trägt, wenn nötig, `ueberschreitung` ein). Die Sendungen an die Teile 2 und 3 und ihre Annahme (`AGENTS.md` §3.13) stehen in §1 und in deren Plänen, am Schnitt geprüft; `docs/user/benutzerhandbuch.md` ändert dieser Teil nicht (Binary unverändert). Der Plan lag beim Start vollständig vor: Tabelle S1 bis S18, N1, N2, Mengen und Raster mit je einem Ort je Zelle, kein Platzhalter.

**Warum:** Rahmen und Authentisierung kommen aus `slice-admin-listener-token`, der Port `AuditLog` für den Schreibendpunkt aus `slice-audit-log` (beide *Code-Abhängigkeit*; der Audit-Port ist über `slice-admin-listener-token` schon vorausgesetzt und hier genannt, weil dieser Slice ihn für `POST` ruft). `SessionUseCase` liegt seit `slice-session-use-case` und `slice-session-beenden` in `done/`.

**Nicht gebraucht:** `welle-mock`, M2, `slice-http-adapter`, `slice-workspace-use-case` und `slice-admin-tls`; der Adapter spricht gegen Fakes, TLS umhüllt den Listener unabhängig von den Routen.

**Reihenfolge und Konflikte** (Planner, 2026-11-21): Die drei Teile des Schnitts schreiben im selben Paket `src/adapters/driving/admin/` und teilen die Zeit, `sessionsFake` und `SessionRouten`; die Messung zählt jede Änderung an den Pfaden (`SPEC-128` Nr. 10 g). Sie laufen nacheinander in derselben Linie (Worktree), nie zugleich: dieser Teil zuerst, dann [slice-admin-sessions-beenden](../open/slice-admin-sessions-beenden.md), dann [slice-admin-inspektion](../open/slice-admin-inspektion.md); Teil 3 hängt nur an der Zeit aus diesem Teil und darf vor Teil 2 laufen. Die Dateien `src/hexagon/ports/support/audit*.go`, `src/adapters/driven/audit/audit_test.go` und `src/bootstrap/audit_betrieb_test.go` ändern allein die Teile 1 und 2.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-21, Teil des Planner 2026-11-21, vor dem ersten Code-Commit; Messung 0 beim Start): der Block unten trägt für diesen Teil 20 Zeilen der Tabelle (S1 bis S18, N1, N2), *a* = 2 (N1, N2), *b* = 18, *f* = 1 (`sessionsFake`), *p* = 300 (Routen und Handler 110, Leser des Bodys 60, Schreiber des JSON 70, Art des Audit 20, Konstruktor 40): *S* = 85·2 + 80·18 + 200·1 + 300 = 2110 gleich der Tabelle, 135·2110 = 284850, *M* = 100 · ⌈28,485⌉ = 2900 gleich der Formel; keine Schnittfrage (135·*S* = 284850 ≤ 300000, *S* < 2223), kein `schnitt-grund`. Der Teil liegt 113 unter der Schnittgrenze: jede weitere Zeile der Tabelle kostet 80, **zwei** weitere lösen die Schnittfrage neu aus (*S* = 2270; eine weitere gibt 2190); wächst die Tabelle im Code (`AGENTS.md` §3.9), prüft der Planner sie neu. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `basis` ist der Commit des Verweis-Nachzugs nach dem Move nach `in-progress/` (`02721bba`). Pfade: die Dateien, die der Architect nannte (`src/adapters/driving/admin` als Verzeichnis, weil die Tests neue Dateien anlegen, deren Namen der Plan nicht vorwegnimmt; `audit.go`, `audit_test.go`, der Test des Adapters und `audit_betrieb_test.go` einzeln, damit fremde Änderungen in `src/hexagon/ports/support/` (`tlstest/`) und `src/bootstrap/` nicht mitzählen); die Folge für die Messung steht in `SPEC-128` Nr. 10 g.

```groessenmarke
pfad: src/adapters/driving/admin/
pfad: src/hexagon/ports/support/audit.go
pfad: src/hexagon/ports/support/audit_test.go
pfad: src/adapters/driven/audit/audit_test.go
pfad: src/bootstrap/audit_betrieb_test.go
produktcode: 300
fakes: 1
basis: 02721bba
schaetzung: 2110
marke: 2900
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke (2900), die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 gesetzt hat (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus dem Block, Tests eingeschlossen). Die Tabelle *Zusage · Fänger* lag vor (§3); der Planner hat sie am 2026-11-21 vor dem Code geschnitten: ungeschnitten *S* = 5715 (über der Schnittgrenze *S* ≥ 2223), dieser Teil *S* = 2110 (*M* = 2900), Teil 2 *S* = 1580 (*M* = 2200), Teil 3 *S* = 2025 (*M* = 2800); ein weiterer Schnitt dieses Teils ist nicht vorgesehen.
- `in-progress` → `open` (blockiert — Carveout?): eine Randform ohne Ort im Vertrag, den der Implementer nicht entscheidet (`AGENTS.md` §3.12); die Fragen des doppelten Audit-Eintrags beim Löschen und des Kontexts von `close` sind entschieden (`SPEC-292` Nr. 9, 10, 11).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): jede Zelle nennt eine Nummer von `SPEC-292` (der Architect, 2026-11-21; Zeilen `GET`, `POST` und Audit des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen) oder „gilt nicht“ mit Grund; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `GET /__admin/sessions` | gilt nicht: kein Argument, Query und Body ohne Wirkung (`SPEC-292` Nr. 3) | `SPEC-292` Nr. 3 (leere Query, leerer Body: ohne Wirkung) | Nr. 3 (unbekannte und ungültige Query, Body: ohne Wirkung); fremde Methode: `SPEC-238` Nr. 2 (405 mit `Allow`, `SPEC-292` Nr. 1) | Nr. 2 (`ctx` unverändert), `SPEC-238` Nr. 6 a 1 und 9 (Kontextfehler bricht die Verbindung ab) | Nr. 12 (jeder Aufruf liest neu, `SPEC-124` Nr. 6) |
| `POST /__admin/sessions`, `service_id` | Nr. 5 d (Pflicht: 400 ohne Aufruf) | Nr. 5 d (nichtleer: 400 ohne Aufruf) | Nr. 5 c, d (falscher Typ, fremder Schlüssel: 400); unbekannter Service: Nr. 5 f und `SPEC-124` Nr. 8 (`ODR-E5005`, 404) | Nr. 5 a (Lesefehler; Kontextfehler), Nr. 8 | Nr. 12 (ohne `id` entsteht eine weitere Session) |
| `POST /__admin/sessions`, `id` | Nr. 5 d, f (der Dienst vergibt die Kennung) | Nr. 5 d, f (leer wie fehlend) | Nr. 5 d (falscher Typ: 400); Form ungültig: Nr. 5 f und `SPEC-124` Nr. 4 (`model.ErrUngueltig`, 400); vergeben: `ODR-E7010` (409) | Nr. 5 a, 8 | Nr. 12 (`ODR-E7010`, kein Eintrag) |
| `POST /__admin/sessions`, Body als Ganzes | Nr. 5 b (leer, nur Leerraum: 400) | Nr. 5 b und d (`{}`: `service_id` fehlt, 400) | Nr. 5 b, c (jede andere Form, jeder fremde Schlüssel: 400); `Content-Type`: Nr. 5 e (nicht gelesen) | Nr. 5 a | Nr. 12 |
| Audit-Eintrag von `POST` und `close` (in diesem Teil: `POST`) | `SPEC-292` Nr. 11 b (jedes Feld je Art gesetzt nach `SPEC-234` Nr. 2) | Nr. 11 e (eine leere Kennung weist `Append` ab, ohne Wirkung auf die Antwort) | Nr. 11 e und *Ungeprüft* (Kennung außerhalb der Form: der Adapter des Ports weist ab) | Nr. 11 a, d (`Append` ohne `ctx`: ein beendeter Kontext ändert den Eintrag nicht, `SPEC-234` Nr. 5 c) | Nr. 12 (je Erfolg genau ein Eintrag, keiner bei `ODR-E7010` und `ODR-E7011`) |

Die Zeilen zu `close` und `DELETE` stehen in §6 von [slice-admin-sessions-beenden](../open/slice-admin-sessions-beenden.md), die zu Exchanges, Coverage und Diagnose in [slice-admin-inspektion](../open/slice-admin-inspektion.md). Ohne Nummer steht in den Zellen `SPEC-292`.

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; der Adapter kennt den Modus nicht: `Create` steht in allen Modi, `SPEC-124` Nr. 3):

| Paar | Entscheidung | Ort | Fänger |
|---|---|---|---|
| Body ungültig × Fehler des Dienstes (unbekannter Service, vergebene `id`) | der Body zuerst: 400, der Dienst wird nicht gerufen | `SPEC-292` Nr. 2, 5 g | S6, S7 (der Fake würde `ODR-E5005` liefern, Zähler 0) |
| zwei Verstöße im selben Body | dieselbe Antwort, welcher gilt, ist nicht zugesagt | Nr. 5 g | S7 (10 Paare der fünf Verstoßarten) |
| `id` mit ungültiger Form × unbekannter Service | der Dienst: die Form zuerst (`model.ErrUngueltig`, 400) | `SPEC-124` Nr. 3, 8 (Rang im Dienst) | gilt nicht für den Adapter; der Fänger liegt beim Dienst, S11 prüft nur das Durchreichen |
| Kontextfehler × Fehler mit Code | der Abbruch gilt (`errors.Join` und Ursache) | `SPEC-292` Nr. 8, `SPEC-238` Nr. 6 a 1 | S11 (Ausgang `errors.Join(context.Canceled, *Fehler E7003)`) |
| Antwort und Fehler des Dienstes zugleich | der Fehler gilt | Nr. 8, `SPEC-238` Nr. 6 b | S11 (Ausgang „Wert und Fehler E7003“) |
| Erfolg des Dienstes × Fehler von `Append` | der Erfolg gilt, die Antwort bleibt | Nr. 11 e | S14 |
| Authentisierung × alles | 401 vor 404 und 405 vor jedem Aufruf einer Route | `SPEC-238` Nr. 3 d | gilt nicht für diesen Plan (Rahmen) |

- Wer schreibt den Audit-Eintrag einer Löschung: der Adapter oder `SessionUseCase` (`slice-audit-log`)? Beide ergäben einen doppelten Eintrag — **Ausgang:** entschieden (Architect, 2026-11-21): allein der Use Case schreibt `session_geloescht` (`SPEC-234` Nr. 7, `SPEC-124` Nr. 7 d1); `DELETE` schreibt nie einen Eintrag (`SPEC-292` Nr. 10, 11 c), gefangen in Teil 2 (B10, M-Audit); Anlegen schreibt der Adapter (`SPEC-292` Nr. 11, S12), Beenden Teil 2.
- Statuscodes der Erfolge und JSON-Form der Antworten stehen nicht in `SPEC-005`; die OpenAPI-Datei ist Voraussetzung (Welle §6) — **Ausgang:** entschieden (Architect, 2026-11-21): Status, Form und Fehler stehen in `SPEC-292` Nr. 6 bis 8 und `SPEC-293` Nr. 6 bis 9; eine eigene OpenAPI-Datei entsteht nicht (`SPEC-005`); der Planner hat die Zeile der Welle beim Schnitt nachgezogen.
- Ein Endpunkt für die Wartung existiert in `SPEC-005` nicht; `slice-session-wartung` nennt „Admin-Endpunkte für Sessions und Wartung“ als Folgewelle — **Ausgang:** entfallen: die Wartung läuft beim Start im Composition Root (`ARC-012`), `SPEC-005` führt keinen Endpunkt, und die Sendung trifft nur den Teil der Sessions, den dieser Slice trägt.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (6×): der Fake von `SessionUseCase` und `AuditLog` trägt jede Schnittstelle des Originals und zählt jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Marke steht nicht, bevor die Tabelle *Zusage · Fänger* vorliegt — **Ausgang:** entschieden (Planner, 2026-11-21): der Block in §4 trägt `basis` `02721bba`, `schaetzung` 2110 und `marke` 2900 (nachgerechnet nach `SPEC-128` Nr. 1, keine Schnittfrage, 113 unter der Schnittgrenze), gesetzt vor dem ersten Code-Commit (`AGENTS.md` §3.14); `make groessenmarke` ist grün.
- Die Teile teilen die Zeit, `sessionsFake` und `SessionRouten` im selben Paket (Schnitt vom 2026-11-21) — **Ausgang:** entschieden (Planner, 2026-11-21): sie laufen nacheinander (1, 2, 3; Teil 3 darf vor Teil 2); die Sendung der Mittel steht in §1 (*Sendung an die Teile 2 und 3*) und in §1 der Nehmer.
- Gesamtzählungen über die drei Teile (I3: 42 Fälle, I17: 84 Zellen) — **Ausgang:** entschieden (Planner, 2026-11-21): der später in `done/` gehende von Teil 2 und Teil 3 hebt die Literale (§1 der Teile 2 und 3); dieser Teil trägt keine Gesamtzahl.
- Zählen der Fehler (`SPEC-230` Nr. 8): die Stelle, die den Fehler ausliefert, zählt — das ist der Rahmen von `slice-admin-listener-token` (Sendung dort, §1 *Übernimmt*), nicht jede Route dieses Slice — **Ausgang:** entfallen für diesen Slice (die Routen rufen kein `CountError`, der Rahmen schreibt das Fehlerobjekt); weiter offen bei der Gegenstelle.
- Randform ohne Ort in `SPEC-292` Nr. 5 b und d, vom Implementer beim Code bemerkt (`AGENTS.md` §3.12, §3.13): (1) ein BOM (U+FEFF) **innerhalb** einer Zeichenkette: Nr. 5 b sagt „UTF-8 ohne BOM (die Bytes `EF BB BF`)“ und der Fänger S6 führt das BOM am Anfang; (2) das Escape eines einzelnen Surrogats in einem Wert (Backslash, `u`, vier Hex-Ziffern von D800 bis DFFF): `encoding/json` dekodiert es zu U+FFFD, Nr. 5 d sagt „dekodiert, unverändert“. Der Code nimmt beides an (der Wert trägt U+FEFF bzw. U+FFFD an `Create`), kein Fänger fixiert es, damit die Entscheidung nicht im Test steht. **Ausgang:** offen — Adresse: Architect (`SPEC-292` Nr. 5 b und d: abweisen mit `ODR-E2102` oder annehmen); ändert die Entscheidung den Code, trifft sie `sessions_body.go` und je einen Fall in `bUngueltig` oder `bGueltig`.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge) und am Tag des Schnitts und des Starts erneut (2026-11-21; Treffer und Zähler beim Start nachgezogen, der Zähler ist die Zahl der Dateien unter `evidence/`). Offen und für den Slice einschlägig: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (6×) für die Fakes der Use Cases (beide verkörpert, in `.claude/commands/implement-slice.md` Schritt 14 und `.harness/skills/reviewer.md`), außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (21×, verkörpert). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (32×, `AGENTS.md` §3.9; die Zeilen von `SPEC-005` sind Stellen), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (59×, `AGENTS.md` §3.17; „jeder Endpunkt“, „jeder Fehlercode“ als Schleife), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (66×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16). Keine Beobachtung steht mit diesem Slice offen unter 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
