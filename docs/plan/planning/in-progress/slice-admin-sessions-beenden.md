# Slice slice-admin-sessions-beenden: Admin-Endpunkte `close` und `DELETE` für Sessions, Audit `session_beendet` (Teil 2 des Schnitts von `slice-admin-sessions`)

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

**Berührte Spec-Stellen:** `SPEC-005` (Zeilen `close` und `DELETE /__admin/sessions/{id}`) · `LH-FA-13.a` (Beenden, Fehlercodes) · `LH-RB-04.a` · `SPEC-124` (Vertrag von `SessionUseCase`: `Beenden`, `Delete`) · `SPEC-071`, `SPEC-072`, `SPEC-097`, `SPEC-111` · `SPEC-109` · `SPEC-292` (Nr. 1, 2, 4, 6 bis 13; Teil 1 liefert Nr. 3, 5 und die Zeit) · `SPEC-234` Nr. 2 und 7 (die Art `session_beendet`) · `SPEC-238` Nr. 6, 9, 15 (Rahmen, nicht geändert) · `ARC-007` · `architecture.md` §2.2, §4.9.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-admin-sessions](../done/slice-admin-sessions.md), Vorschlag des Architect vom 2026-11-21). **Datum:** 2026-11-21.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem** Abschnitt des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Admin-Adapter bildet `POST /__admin/sessions/{id}/close` und `DELETE /__admin/sessions/{id}` aus `SPEC-005` auf `SessionUseCase.Beenden` und `SessionUseCase.Delete` ab. `close` wartet auf den Dienst (auch über `SHUTDOWN_TIMEOUT` hinaus), schreibt nach der Tatsache den Eintrag `session_beendet` und antwortet 200 mit der Session; `DELETE` antwortet 200 mit dem Bericht der Bereinigung (`blobs_removed`, `cleanup_deferred`) und schreibt nie einen Eintrag (den Eintrag `session_geloescht` schreibt der Dienst); jede Antwort trägt Status, Fehlerobjekt und Korrelations-ID, kein Geheimnis steht in einer Antwort. Der Teil erweitert `SessionRouten` aus Teil 1 von zwei auf vier Routen, nutzt die Funktion der Zeit und `sessionsFake` aus Teil 1 (hier um Zustand und Haltepunkte erweitert) und hebt die Zählstelle der Arten von `support.Arten()` von 6 auf 7.

**Vertrag (Architect, 2026-11-21):** `SPEC-292` Nr. 4 (Pfadwert `{id}`), Nr. 6 bis 10 (Antworten, Fehler, Beenden, Löschen) und Nr. 11 bis 13 (Audit, Nebenläufigkeit, Geheimnisse); der Vertrag steht auf `main`, bevor dieser Teil startet.

**Übernimmt (Sendung von `slice-session-use-case` und `slice-session-beenden`, Teil des Schnitts von `slice-admin-sessions`, `AGENTS.md` §3.13):** die Endpunkte `close` und `DELETE` für Sessions; beide Slices liegen in `done/` ([slice-session-use-case](../done/welle-record-replay/slice-session-use-case.md), [slice-session-beenden](../done/welle-record-replay/slice-session-beenden.md)), und ihr §1 *Ausdrücklich NICHT* weist die Endpunkte der Folgewelle zu. Die Endpunkte `POST` und `GET` liegen in Teil 1.

**Übernimmt (Sendung von `slice-admin-sessions`, Teil 1; Geber: `slice-admin-sessions`; Eintrag des Planners vom 2026-11-21 beim Schnitt, `AGENTS.md` §3.13):** (a) die Mittel aus Teil 1, die dieser Teil erweitert: die Funktion der Zeit (RFC 3339, neun Nachkommastellen, S9), `sessionsFake` (alle sieben Operationen gezählt; dieser Teil ergänzt den Zustand mit eigener Sperre für N3 und die Haltepunkte für B6, B7 und N4) und den Konstruktor `SessionRouten` (vier Routen in der Reihenfolge von `SPEC-292` Nr. 1); (b) die Zählstelle der Arten: `len(support.Arten())` gleich 7 in `audit_test.go` des Ports, in `audit_test.go` des Adapters und in `audit_betrieb_test.go` (B14); dazu der Fänger zu `V-4` aus Teil 1 (akzeptiertes Negativ (6) dort, Eintrag des Planners bei der Closure von Teil 1, `AGENTS.md` §3.13): `TestKennungsform` (`audit_test.go`) und `TestLogBeiFehler` (`log_sperre_test.go`) im Adapter des Audit laufen über `support.Arten()` statt über feste Arten und führen damit die Arten `session_angelegt` und `session_beendet`; die Mutationen „die Prüfung der Form der Kennung für die Art überspringen“ und „das Feld `session_id` im Log eines gescheiterten `Append` weglassen“ sind je Art rot (in Teil 1 blieben beide für `session_angelegt` grün, 2 von 9); (c) der Fänger S16 (AST-Test: kein Aufruf von `NeuesFehlerobjekt`, `CountError` und `ArtAdminAbgewiesen`) liest die Dateien der Routen: die Menge bildet der Test aus dem Verzeichnis (jede Datei `.go` ohne `_test.go` außer den fünf des Rahmens), eine neue Datei dieses Teils gehört dazu, ohne dass der Test sie nennt, und die Zählprüfung hält Rahmen plus Routen gleich allen Dateien; (d) die Zeilen B4, B5, B9 und B14 verweisen auf S4, S8, S11 und S15 und lesen deren Wortlaut im Plan von Teil 1 ([slice-admin-sessions](../done/slice-admin-sessions.md)); (e) die Gesamtzahlen über alle drei Teile, die zwei Zeilen von Teil 3 als Literal tragen (I3: 42 Fälle mit 7 Treffern, I17: 84 Zellen mit genau 2 Einträgen): wer von Teil 2 und Teil 3 als zweiter in `done/` geht, hebt die Literale auf den Endwert; läuft dieser Teil zuerst (Reihenfolge 1, 2, 3), hebt Teil 3 sie (dessen §1), läuft Teil 3 zuerst, zählen seine Tests 28 Fälle mit 5 Treffern (4 Pfadformen) und 57 Zellen mit einem Eintrag, und dieser Teil hebt sie auf 42 mit 7 und 84 mit 2.

**Übernimmt (Sendung von `slice-admin-listener-token`, Pflichten der Endpunkte nach `SPEC-238` Nr. 15, Schnitt vom 2026-11-07, `AGENTS.md` §3.13; Teil des Schnitts von `slice-admin-sessions`):** (a) was ein Dienst als `model.ErrUngueltig` ohne Code liefert, gibt die Route als Fehler zurück (die Abbildung auf `ODR-E2102` mit 400 gehört dem Rahmen); die Route liest weder Query noch Body (B4) und prüft die Form von `{id}` nicht (B3); (b) sie ruft `CountError` nie und `Audit.Append` nie für `ODR-E2101` und schreibt kein Fehlerobjekt selbst; (c) die vier Arten von `SPEC-234` Nr. 2 schreibt der Dienst, nie die Route; die Art für Beenden ist eine Zeile im Vertrag dieses Endpunkts (`SPEC-292` Nr. 11, B11); (d) eine Erfolgsantwort hat einen Status von 200 bis 299 und trägt Daten als `Antwort.Body`. Der Fänger liegt in diesem Slice (die Routen und ihre Tests); der Rahmen hat keine Route (`slice-admin-abweisen-audit`, Zeile P1, akzeptiertes Negativ). Der Reviewer löst jede Zeile von Nr. 15 mit einer Mutation auf.

**Übernimmt (Sendung von `slice-audit-verdrahtung` und `slice-audit-log`, ehemals an `slice-admin-sessions`; beim Schnitt von Teil 1 an diesen Teil gereicht, weil allein `DELETE` eine Löschung auslöst; Eintrag des Planners vom 2026-11-21, §1 *Ausdrücklich NICHT* und DoD dieses Teils vorher gelesen, kein Ausschluss trifft die Sendung):** das Handbuch-Kapitel zur Datei `audit.jsonl` im Datenverzeichnis, sobald ein gebautes Binary eine Löschung auslösen kann (Bedingung in der DoD, Ausgang in §6).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `GET` und `POST /__admin/sessions`, die Funktion der Zeit, der Leser des Bodys, die Art `session_angelegt` — [slice-admin-sessions](../done/slice-admin-sessions.md) (Teil 1, liegt bei diesem Start in `done/`); dieser Teil erweitert und ändert sie nicht.
- Exchanges, Coverage und Diagnose — `slice-admin-inspektion` (Teil 3, [slice-admin-inspektion](../open/slice-admin-inspektion.md)); er teilt mit diesem Teil das Paket und die Zeit, nicht den Gegenstand.
- Imports, Workspaces und Snapshots — `slice-admin-imports` und `slice-admin-workspaces`.
- Der Rahmen (Listener, Token, Fehlerobjekt, Abbildung der Port-Fehler) — `slice-admin-listener-token`; dieser Teil trägt die Routen und ihre Abbildung.
- Statusregel, Rangfolge der Fehler und Bereinigung der Blobs — `SessionUseCase` (`done/`); der Adapter bildet ab und schreibt nie direkt in die Datenbank (Architektur §2).
- Der Port `AuditLog` — `welle-betrieb` (`slice-audit-log`); der Teil ruft ihn.
- Der Audit-Eintrag der Löschung einer Session (Art `session_geloescht`) — `SessionUseCase.Delete` schreibt ihn (`slice-audit-dienste`); der Adapter schreibt die vier Arten von `SPEC-234` Nr. 2 nie, auch nicht für denselben Request (Entscheidung `SPEC-234` Nr. 7). Der Adapter auditiert nur eigene Arten (Anlegen in Teil 1, Beenden hier).
- Ein Endpunkt für die Wartung — `SPEC-005` führt keinen; die Startwartung läuft im Composition Root (Ausgang in §6 von Teil 1).

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

- [ ] Sessions beenden und löschen: `close` (`ODR-E7003`, `ODR-E7011`; die Antwort wartet auf `Beenden`, auch über `SHUTDOWN_TIMEOUT` hinaus, ohne Frist und ohne Ablösung des Kontexts) und `DELETE` (`ODR-E7003`, `ODR-E7004`, `ODR-E7009`; der Fehler der Bereinigung ändert die Antwort nicht; Antwort 200 mit dem Bericht); der Pfadwert `{id}` geht byte-gleich an den Dienst; Statuscodes und Form der Antworten nach Vertrag (B1 bis B9).
- [ ] Audit: `close` schreibt über `AuditLog` genau einen Eintrag `session_beendet` ohne Geheimnisse, nach dem Erfolg von `Beenden` und nie bei einem Fehler; `DELETE` schreibt nie einen Eintrag, so entsteht keine Doppelung mit dem Eintrag, den die Löschung im Use Case schreibt; ein Fehler von `Append` ändert die Antwort nicht; die Art ist die siebente Zeile von `support.Arten()` und die Zählstellen der Tests ziehen mit; die Schleifen von `TestKennungsform` und `TestLogBeiFehler` im Adapter des Audit laufen über `support.Arten()` (Sendung `V-4` aus Teil 1, §1) (B10 bis B16, N3, N4).
- [ ] Pflichten der Endpunkte nach `SPEC-238` Nr. 15 (a) bis (d) für `close` und `DELETE`: je Route gefangen, mit dem Fänger S16 aus Teil 1 über die Dateien beider Teile (Geber `slice-admin-listener-token`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, was unverändert blieb (`bootstrap` außer dem Test, `cmd`, Katalog und Handbuch; der Admin-Listener erreicht das Binary erst mit `slice-bootstrap-m6`). Hinweis aus `slice-audit-verdrahtung` (§6, an diesen Teil gereicht): die Datei `audit.jsonl` im Datenverzeichnis ist für den Anwender sichtbar, ihr Kapitel folgt mit dem ersten Slice, dessen Binary eine Löschung auslöst; das ist dieser Teil, sobald der Listener verdrahtet ist (`slice-bootstrap-m6`) und ein Beispiel gegen das gebaute Binary die Datei erzeugen kann; sonst nennt §7 die Adresse (Risiko in §6).
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
| `src/adapters/driving/admin/` | update | `SessionRouten` erweitert um `close` und `DELETE` (Dateinamen sind Vorschlag: `sessions_beenden.go`, `sessions_loeschen.go`); die Zeit und die Form der Session liefert Teil 1; keine Änderung an `admin.go`, `route.go`, `fehler.go`, `lebenszyklus.go` |
| `src/hexagon/ports/support/audit.go`, `src/hexagon/ports/support/audit_test.go` | update | die Art `session_beendet` (`ArtSessionBeendet`), siebente Zeile von `Arten()` (`SPEC-292` Nr. 11); der Test zählt 6 Arten und wächst auf 7 |
| `src/adapters/driven/audit/audit_test.go`, `src/adapters/driven/audit/log_sperre_test.go`, `src/bootstrap/audit_betrieb_test.go` | update (nur Tests) | Zählstellen der Arten (6 → 7) und Zeile der Art; die Schleifen von `TestKennungsform` und `TestLogBeiFehler` über `support.Arten()` (Sendung `V-4` aus Teil 1); `bootstrap` bleibt außer diesem Test unverändert |
| `spec/spezifikation.md` | geliefert (Architect, auf `main`) | `SPEC-292`; dieser Teil schreibt sie nicht |
| Tests | neu | Adapter gegen `sessionsFake` (aus Teil 1, erweitert) und `FakeAudit`: jedes Argument festgehalten, jede Operation gezählt, auch die, die der Adapter nie rufen darf (`AGENTS.md` §3.15); ein Fall je Aufrufstelle für den `ctx` des Requests |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem 2026-11-21 (Architect) mit dem Vertrag in der Spezifikation (`SPEC-292`) vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. `Nr.` ist die Kennung des ungeschnittenen Plans: **B** gehört zu diesem Teil (`close`, `DELETE`, Audit `session_beendet`), **N3** und **N4** sind seine nebenläufigen Zeilen; die Zeilen **S** und **N1**, **N2** stehen in Teil 1, **I** und **N5** in Teil 3. Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest.

**Entscheidungen des Architect für diesen Teil** (2026-11-21, vor dem ersten Code-Commit, `AGENTS.md` §3.12; Ort in der Spezifikation; die übrigen Aufträge stehen in §3 von [slice-admin-sessions](../done/slice-admin-sessions.md)):

1. **Statuscodes, JSON-Form, Fehlerabbildung:** `SPEC-292` Nr. 6 bis 8. Erfolg: 200 für `close` und `DELETE`; `DELETE` antwortet 200 mit dem Bericht der Bereinigung (kein 204: der Client sähe sonst nirgends, dass Blobs zurückgestellt sind).
2. **`close`:** `SPEC-292` Nr. 9. Die Antwort kommt erst nach `Beenden`, ohne Frist und ohne Goroutine; der Kontext bleibt der des Requests (keine Ablösung, `SPEC-124` Nr. 11 e genügt); die Dauer ist unbegrenzt, weil der Rahmen kein `WriteTimeout` setzt (`SPEC-238` Nr. 10). Folge im Betrieb: ein Client mit Timeout unter `SHUTDOWN_TIMEOUT`, der trennt, macht die Session `incomplete`.
3. **Audit:** `SPEC-292` Nr. 11 und `SPEC-234` Nr. 2 und 7. Die Art `session_beendet`, geschrieben nach der Tatsache, nie bei einem Fehler, mit den Feldern aus dem Ergebnis des Dienstes; ein Fehler von `Append` ändert nichts. `DELETE` schreibt nie einen Eintrag: der Dienst schreibt `session_geloescht`.
4. **Fehlversuche am Token** (Entscheidung des Users, Welle §5, „Kein Code, bis entschieden“): sie berührt diesen Plan nicht (allein der Rahmen, `SPEC-238` Nr. 8). Ändert die Entscheidung die Zahl der Arten von `support.Arten()`, ändert sich allein die Zählstelle der Tests (B14).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-21, Zeilen B1 bis B16, N3 und N4 des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| B1 | (Nr. 1, `SPEC-238` Nr. 2) Mit Teil 2 liefert `SessionRouten` vier Routen in der Reihenfolge des Vertrags; `POST /__admin/sessions/{id}/close` und `DELETE /__admin/sessions/{id}` treffen, jede andere Methode ist 405 mit `Allow`; `GET /__admin/sessions/{id}` ist 405 mit `Allow: DELETE` | M-Route: 3 Pfadformen × 7 Methoden = 21 Fälle (4 Treffer, 17 mal 405); Orakel Literale für `Allow` (`GET, POST`; `DELETE`; `POST`); Zählprüfung gegen die Zahl der Muster | `Allow` ohne `DELETE`; Muster `/close` fehlt | andere |
| B2 | (Nr. 2) `close` ruft genau `Beenden(ctx, id)`, `DELETE` genau `Delete(ctx, id)`, mit `ctx` unverändert und `id` byte-gleich, sonst keine Operation | `sessionsFake`: 2 Routen × 7 Zähler = 14 Zellen, Diagonale 1, sonst 0; Wert im `ctx` kommt an, Gleichheit des `ctx` | `ctx` abgeleitet; `Get` zur Prüfung der Existenz gerufen; `CloseAll` gerufen | andere |
| B3 | (Nr. 4) Der Pfadwert `{id}` geht byte-gleich an den Dienst, ohne Prüfung von Form, Länge und Schreibweise, und steht in keiner Antwort | M-Pfadwert: 13 Werte (`x`, `ses-` und 32 Ziffern, `SES-1`, `a.b`, `.`, `..`, `a%2Fb`, `%C3%A4`, `%FF`, `%00`, 65 mal `a`, `a+b`, `%20`) × 2 Routen = 26 Fälle; Orakel: die dekodierten Bytes als Literale aus Bytewerten; ein Canary steht in keinem Byte von Body und Header; Zählprüfung 13 × 2 | Form geprüft (`[a-z0-9-]`); Länge auf 64 gekürzt; `strings.ToLower`; `url.QueryUnescape` (`+` zu Leerzeichen) | andere |
| B4 | (Nr. 3) `close` und `DELETE` lesen weder Query noch Body | wie S4 für die zwei Routen: 6 Queries × 2 Routen = 12 Fälle, dazu 3 ungültige Bodies je Route = 6; Orakel Status 200 und dieselben Argumente; Zählprüfung 18 | Query geprüft; Body gelesen | andere |
| B5 | (Nr. 6, 7 b) `close` antwortet 200 mit der Session des Dienstes und ihrem neuen Status | M-Felder wie S8 (vier Felder mit eigenen Werten); Status `complete`, `incomplete`, `x`; Orakel Literale | Status 202; Session aus dem Pfad statt aus dem Ergebnis | andere |
| B6 | (Nr. 9 a) Die Antwort von `close` entsteht erst nach der Rückkehr von `Beenden`: bis dahin ist nichts geschrieben, und die Route setzt keine Frist | Wirkungs-Fänger (deterministisch, Kanäle): `Beenden` blockiert, der Test sieht, dass `ServeHTTP` nicht zurückgekehrt ist und der Recorder leer ist; nach der Freigabe 200 mit derselben Antwort wie ohne Wartezeit; `ctx.Deadline()` im Fake nicht gesetzt; das fehlende `WriteTimeout` fängt `grenzen_test.go` des Rahmens | `Beenden` in einer Goroutine und Antwort 202 sofort; `context.WithTimeout(ctx, 30*time.Second)`; `WriteHeader` vor dem Aufruf | andere |
| B7 | (Nr. 9 c, 11 a) Endet der `ctx` während `Beenden`, läuft der Aufruf mit dem `ctx` des Requests unverändert zu Ende, die Route schreibt den Eintrag und gibt das Ergebnis zurück; ein Kontextfehler von `Beenden` bricht die Verbindung ab, ohne Eintrag | Der Fake sieht `ctx.Done()` nach `cancel()` des Requests (kein `WithoutCancel`) und liefert die Session: Antwort 200, 1 Eintrag; ein zweiter Lauf liefert `context.Canceled`: Abbruch, 0 Einträge; ein Wert im `ctx` kommt an | `context.WithoutCancel(ctx)` (der Fake sieht kein `Done`); Eintrag auch bei Kontextfehler | andere |
| B8 | (Nr. 10, 6) `DELETE` antwortet 200 mit `{"blobs_removed":<Entfernt>,"cleanup_deferred":<Zurueckgestellt>}`, den zwei Feldern des Berichts in dieser Reihenfolge | M-Bericht: 6 Fälle (0 und `false`, 3 und `false`, 0 und `true`, 7 und `true`, größter `int` und `false`, 1 und `true`); Orakel Literale; Zählprüfung gegen `reflect.TypeOf(inbound.Loeschbericht{}).NumField()` gleich 2 | 204 ohne Body; Felder vertauscht; `cleanup_deferred` als Zahl | andere |
| B9 | (Nr. 8; `SPEC-238` Nr. 6) Fehlerabbildung von `close` und `DELETE` wie S11 | M-Fehler: 8 allgemeine Ausgänge × 2 Routen + `close` 4 (`ODR-E7003`, `ODR-E7011`, `ODR-E3001`, `ODR-E9001`) + `DELETE` 5 (`ODR-E7003`, `ODR-E7009`, `ODR-E7004`, `ODR-E3001`, `ODR-E9001`) = 25 Zellen; Orakel Literale | `ODR-E7011` als 404; `ErrReferenziert` (Fehler eines Ports ohne Code) als 409 übersetzt | andere |
| B10 | (Nr. 10, 11 c; `SPEC-234` Nr. 7) `DELETE` schreibt nie einen Eintrag, auch nicht `session_geloescht` | M-Audit für `DELETE`: 1 Erfolg + 8 + 5 = 14 Zellen, in jeder `FakeAudit.Aufrufe()` gleich 0 | Eintrag `session_geloescht` nach dem Erfolg | andere |
| B11 | (Nr. 11 a, b, d) Nach einem Erfolg von `Beenden` ruft `close` `Append` genau einmal: Art `session_beendet`, `ServiceID` und `SessionID` aus dem Ergebnis (nicht aus dem Pfad), `KorrelationsID` der Antwort, übrige Felder leer, vor der Rückkehr des Handlers; auch bei Status `incomplete` | Der Fake liefert eine Session mit anderer `ID` als der Pfad; Eintrag Feld für Feld; Status `complete` und `incomplete` je 1 Eintrag; Haltepunkt `Vor`: `Beenden` ist gelaufen, der Recorder leer | `SessionID` aus dem Pfad; Eintrag bei `incomplete` unterdrückt; Eintrag vor `Beenden` | andere |
| B12 | (Nr. 11 c) Kein Eintrag bei jedem anderen Ausgang von `close`; über alle vier Routen genau zwei Zellen mit Eintrag | M-Audit `close`: 1 + 8 + 4 = 13 Zellen, Orakel 1 Eintrag beim Erfolg, sonst 0; Zählprüfung über die vier Routen: 24 (Teil 1) + 13 + 14 = 51 Zellen, genau 2 mit Eintrag (`POST` und `close`) | Eintrag bei `ODR-E7011`; Eintrag bei einem Fehler nach `Drain` | andere |
| B13 | (Nr. 11 e) Ein Fehler von `Append` ändert die Antwort von `close` nicht | `FakeAudit.Fehler` gesetzt: Antwort Byte für Byte gleich der ohne Fehler | Fehler von `Append` als 500 | andere |
| B14 | (`SPEC-234` Nr. 2; `SPEC-235` Nr. 3; `SPEC-292` Nr. 11) `support.Arten()` führt `session_beendet` als siebente Zeile, Pflicht `service_id` und `session_id`, und der Adapter schreibt die Zeile | wie S15; `len(support.Arten())` gleich 7 | wie S15 | andere |
| B15 | (Nr. 11 Folge) Akzeptiertes Negativ: scheitert `close` nach dem Beginn von `Drain`, steht kein Eintrag, obwohl die Session keine Exchanges mehr annimmt und `open` bleibt | kein Fänger. Grund: der Eintrag belegt eine Tatsache (`SPEC-234` Nr. 5 a), und ein Zwischenzustand ist keine. Folge: Fehlerantwort und Zähler des Codes sind die Spur; ein erneutes `close` ist zulässig (`SPEC-124` Nr. 11). Maß: 3 von 3 Fehlerpfaden nach `Drain` (`Drain` selbst, `ListBySession`, `SetStatus`) | nicht anwendbar | andere |
| B16 | (Nr. 11, *Ungeprüft*) Akzeptiertes Negativ: eine Session, die ein anderer Weg mit einer Kennung außerhalb der Form von `SPEC-234` Nr. 3 anlegte, hat nach `close` keinen Eintrag, weil der Adapter des Ports ihn abweist | kein Fänger. Grund: `FakeAudit` hält jeden Eintrag ungeprüft fest, die Abweisung liegt bei `SPEC-235` Nr. 4 a. Folge: der Fehler steht im Log, der Eintrag fehlt. Maß: 0 von n Sessions aus `Create` (jede erfüllt die Form, `SPEC-124` Nr. 4) | nicht anwendbar | andere |
| N3 | (Nr. 12) Gleichzeitige `close` und `DELETE` beeinflussen einander nicht; zwei gleichzeitige `close` derselben Session: eines 200 mit Eintrag, das andere 409 ohne | 16 Goroutinen je Route mit eigener `id`; ein Fake mit Zustand (Sperre im Fake) liefert dem zweiten `close` derselben Session `ODR-E7011`; Orakel: 1 Erfolg, 1 Eintrag; unter `make test-race`, je Prozess `-test.count=1` | Paketvariable für die letzte Session; Eintrag aus einer geteilten Variable | nebenläufig |
| N4 | (Nr. 12) Keine Sperre über `Beenden` und `Delete`: zwei `close` verschiedener Sessions sind zugleich im Dienst | Wirkungs-Fänger: `Beenden` jeder Kennung kehrt erst zurück, wenn beide eingetreten sind (Kanal); Frist 10 s nur als Abbruch im Fehlschlag | `sync.Mutex` um `Beenden`; um `Delete` | nebenläufig |

Zeilen: B1 bis B16, N3, N4 gleich 18 (*a* = 2, *b* = 16).

**Verteilung beim Schnitt** (Planner, 2026-11-21): Die Zeilen sind im Wortlaut des Architect übernommen. Wo eine Zeile auf eine Zeile eines anderen Teils verweist (B4 auf S4, B5 auf S8, B9 auf S11, B14 auf S15), gilt deren Wortlaut im Plan von Teil 1 ([slice-admin-sessions](../done/slice-admin-sessions.md), bei diesem Start in `done/`). B1 setzt Teil 1 voraus (die Muster von `GET` und `POST /__admin/sessions` liefert er) und zählt 3 Pfadformen. B12 zählt über die vier Routen 51 Zellen: Teil 3 hat keine Zelle mit Eintrag, die Zahl gilt unabhängig davon, ob Teil 3 vor diesem Teil lief. Die Gesamtzahlen aller drei Teile (Zeilen I3 und I17 in Teil 3) stehen in §1 (*Übernimmt*, Punkt (e)).

**Mengen mit Zählprüfung** (Generator oder Schleife, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15; Architect, 2026-11-21, je Teil getrennt; das Orakel trägt Statuscodes, Codes, Namen und Bytes als Literale aus dem Vertrag, nie als Konstanten des Codes und nie als Aufruf des Lesers):

| Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|---|
| M-Route (B1) | Pfadformen 3 (`/__admin/sessions`, `…/{id}`, `…/{id}/close`) × Methoden 7 (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS) = 21, davon 4 Treffer | Schleife | Status, `Allow` je Pfadform (`GET, POST`; `DELETE`; `POST`) als Literale | Pfadformen gleich der Zahl der verschiedenen Muster der Routen (3), Routen 4 gleich der Länge der Liste des Konstruktors |
| M-Pfadwert (B3) | Werte 13 (siehe B3) × Routen mit `{id}` 2 (`close`, `DELETE`) = 26 | Schleife über die Werte aus Bytewerten | die dekodierten Bytes als Literale | 13 gleich der Länge der Liste, 2 gleich der Zahl der Muster mit `{id}` |
| M-Query (B4) | Queries 6 (leer, `limit=x`, `a=1&a=2`, `%zz`, `%`, `a=1;b=2`) × Routen 2 = 12 (ohne Wirkung); dazu 3 ungültige Bodies je Route = 6 | Schleife | Status und Argumente am Fake als Literale | Summe der Achsen (18) |
| M-Felder (B5, B8) | Felder: `Session` 4 (jedes mit eigenem Wert), `Loeschbericht` 2 | Schleife über `reflect`, Wert je Feldindex | Namen und Reihenfolge als Literale des Vertrags | `NumField` je Typ gegen die Zahl der Literale (4, 2) |
| M-Bericht (B8) | 6 Werte von `Loeschbericht` | Schleife | Literale | 6 |
| M-Fehler (B9) | allgemeine Ausgänge 8 × Routen 2 = 16, Codes der Tabelle von `SPEC-292` Nr. 8: `close` 4, `DELETE` 5 = 9; zusammen 25 | Schleife | Status, Code und Body als Literale aus `SPEC-109` Nr. 2 | 25 gleich 2 × 8 + 9; die Codes gegen `model.AlleCodes()` (jeder Code der Tabelle ist darin) |
| M-Audit (B10, B12) | Zellen Route × Ausgang: `close` 13, `DELETE` 14 = 27; mit den 24 aus Teil 1 über die vier Routen 51; genau 2 mit 1 Eintrag, 49 mit 0 | Schleife über M-Fehler und je einen Erfolg | `FakeAudit.Aufrufe()` und `Eintraege()` als Literale | 51 gleich 4 × 9 + 15 (Erfolge und Codes); 2 Zellen mit Eintrag |

**Unbeschränkte Mengen und ihre Prüfgrenze (akzeptierte Negative, Grund und Folge):** die Länge von `{id}` (Prüfgrenze 65 Bytes; Grund: kein Zweig hängt an der Länge; Folge: jede Länge bis zur Header-Grenze von 64 KiB erreicht den Dienst unverändert); die Zahl der gleichzeitigen Requests (Prüfgrenze 16, N3; Folge: die Routen halten keinen Zustand, die Zahl ändert nichts).

**Zweige als Menge, Runden unter dem Detektor** (Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat in den neuen Dateien dieses Teils null Blöcke mit Zähler 0; der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Z-1 Konstruktor: Routen in der Reihenfolge (Erweiterung auf vier) | B1 | ein Lauf |
| Z-4 Schreiber des JSON: Bericht der Löschung, Session der Antwort von `close` | B5, B8 | ein Lauf |
| Z-5 Handler: Erfolg, Fehler des Dienstes, Abbruch, Audit (Erfolg, Fehler von `Append`) | B7, B9 bis B13 | ein Lauf |
| Z-6 Rückgaben, die nur bei einem Fehler von `json.Marshal` über einer Zahl oder Zeichenkette greifen | kein Fänger (unerreichbar) | der einzige erlaubte Ausreißer, mit Datei und Funktion im Bericht |

Nebenläufigkeit: N3 und N4; der Zustand nach den Konstruktoren ist schreibgeschützt.

**Aufrufstellen und zählende Fakes** („jeder Fall je Aufrufstelle“ gilt für diese drei Stellen):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| `Beenden` (`POST …/close`) | `SessionUseCase` | B2, B5 bis B7, B11 | `sessionsFake` (aus Teil 1) zählt jede der sieben Operationen, hält `ctx` und jedes Argument, liefert Ergebnis und Fehler einstellbar, ein Haltepunkt je Operation (für B6, B7, N4), keine Panik; der Haltepunkt blockiert bis zur Freigabe |
| `Delete` (`DELETE …/{id}`) | `SessionUseCase` | B2, B8, B10 | wie oben |
| `Append` nach `Beenden` | `AuditLog` | B11 bis B13 | `FakeAudit` (besteht): `Vor`, `Fehler`, `Aufrufe`, `Eintraege` |

Der Fake `sessionsFake` stammt aus Teil 1 (*f* = 0 für diesen Teil: er wird erweitert, nicht neu geschrieben); `sessionsFake` trägt für N3 einen Zustand mit eigener Sperre, nicht der Adapter. `FakeAudit`, `FakeTelemetry` und die Fakes der Kennungen bestehen. Ein Test-Server mit echtem Client gehört nicht dazu: die Tests rufen `ServeHTTP` mit `httptest`; die Zusagen über das Trennen der Verbindung prüft der Kontext des Requests (B7), nicht der Socket (*Ungeprüft*).

**Bindung an die Ports** (Lerneintrag V-2): `src/adapters/driving/admin/bindung_test.go` bindet `sessionsFake` an `inbound.SessionUseCase` (aus Teil 1); `SessionRouten` nimmt die Schnittstellentypen der Ports als Parameter (nicht Funktionen), so übersetzt eine Mutation einer Signatur das Paket nicht mehr.

**Größe und Schnitt:** steht in §3 von [slice-admin-sessions](../done/slice-admin-sessions.md) (Architect, 2026-11-21; Zahlen des Planner in §4 dort und hier). Dieser Teil: *a* = 2, *b* = 16, *f* = 0, *p* = 130 (Handler 70, Bericht 15, Art des Audit 10, Konstruktor 35), *S* = 1580, *M* = 2200. Die Teile teilen einen Mechanismus (Zeit, `sessionsFake`, `SessionRouten`) und laufen nacheinander (§4, *Reihenfolge und Konflikte*).

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-admin-sessions](../done/slice-admin-sessions.md) (Teil 1) liegt in `done/`; mit ihm liegen die Bedingungen von Teil 1 in `done/` (`slice-admin-listener-token` samt seinen Teilen `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit`, `slice-audit-log` und `slice-audit-dienste`). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): `SPEC-292` liegt vor (Architect, 2026-11-21).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-admin-sessions` (Teil 1): *Code-Abhängigkeit*. Die Funktion der Zeit, `sessionsFake`, der Konstruktor `SessionRouten` mit zwei Routen und die Zählstelle der Arten (6) liegen dort; dieser Teil erweitert sie im selben Paket.
- `slice-admin-listener-token` und seine Teile: *Code-Abhängigkeit*. Der Rahmen (`Route`, Zuordnung, Fehlerabbildung, `Audit`) trägt die Routen; ohne ihn gibt es keinen Ort für sie.
- `slice-audit-log`: *Code-Abhängigkeit, Port*. `AuditLog` für den Eintrag von `close`.
- `slice-audit-dienste`: der Eintrag der Löschung entsteht erst im Use Case, und die DoD verlangt „ohne Doppelung“.
- `slice-session-use-case`, `slice-session-beenden`: `SessionUseCase.Beenden` und `.Delete` (beide in `done/`).

**Nicht gebraucht:** `slice-inspection-use-case` und `slice-inspection-coverage-diagnose` (Bedingung von Teil 3), `welle-mock`, M2, `slice-http-adapter`, `slice-workspace-use-case` und `slice-admin-tls`; der Adapter spricht gegen Fakes, TLS umhüllt den Listener unabhängig von den Routen.

**Reihenfolge und Konflikte** (Planner, 2026-11-21): Die drei Teile des Schnitts von `slice-admin-sessions` schreiben im selben Paket `src/adapters/driving/admin/` und teilen die Zeit, `sessionsFake` und `SessionRouten`; die Messung zählt jede Änderung an den Pfaden (`SPEC-128` Nr. 10 g). Sie laufen nacheinander in derselben Linie, nie zugleich: Teil 1, dann dieser Teil, dann Teil 3 ([slice-admin-inspektion](../open/slice-admin-inspektion.md)); Teil 3 hängt nur an Teil 1 und darf vor diesem Teil laufen. Die Dateien `src/hexagon/ports/support/audit*.go`, `src/adapters/driven/audit/audit_test.go` und `src/bootstrap/audit_betrieb_test.go` ändern allein die Teile 1 und 2, Teil 3 ändert sie nicht. In `src/bootstrap/` ändert dieser Teil allein `audit_betrieb_test.go` (Zählstelle der Arten, 6 → 7); Slices, die `src/bootstrap/` mit dem Verzeichnis als Pfad messen (`slice-listener-tls-verdrahtung`), zählen diese Zeilen als fremde mit. Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-21): der Block unten trägt für diesen Teil 18 Zeilen der Tabelle (B1 bis B16, N3, N4), *a* = 2 (N3, N4), *b* = 16, *f* = 0 (`sessionsFake` aus Teil 1 wird erweitert), *p* = 130: *S* = 85·2 + 80·16 + 200·0 + 130 = 1580; 135·1580 = 213300, *M* = 100 · ⌈21,33⌉ = 2200; keine Schnittfrage (*S* < 2223). Der Teil liegt 643 unter der Schnittgrenze: jede weitere Zeile der Tabelle kostet 80, neun weitere lösen die Schnittfrage neu aus (*S* = 2300); wächst die Tabelle im Code (`AGENTS.md` §3.9), prüft der Planner sie neu. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. `basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit (`AGENTS.md` §3.14); bis dahin steht die Marke aus (Prosa-Marke und Block ersetzen einander nicht, der Block steht allein).

```groessenmarke
pfad: src/adapters/driving/admin/
pfad: src/hexagon/ports/support/audit.go
pfad: src/hexagon/ports/support/audit_test.go
pfad: src/adapters/driven/audit/audit_test.go
pfad: src/adapters/driven/audit/log_sperre_test.go
pfad: src/bootstrap/audit_betrieb_test.go
produktcode: 130
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt (`AGENTS.md` §3.14); ein weiterer Schnitt ist nicht vorgesehen, die Schätzung liegt weit unter der Schnittgrenze.
- `in-progress` → `open` (blockiert): eine Randform ohne Ort im Vertrag, den der Implementer nicht entscheidet (`AGENTS.md` §3.12); die Fragen des doppelten Audit-Eintrags beim Löschen und des Kontexts von `close` sind entschieden (`SPEC-292` Nr. 9, 10, 11).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): jede Zelle nennt eine Nummer von `SPEC-292` (der Architect, 2026-11-21; Zeilen `close`, `DELETE` und Audit des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen) oder „gilt nicht“ mit Grund; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `POST /__admin/sessions/{id}/close`, `{id}` | gilt nicht: ein leeres Segment passt auf keine Route (404 `ODR-E2102`, `SPEC-238` Nr. 2) | gilt nicht: wie fehlend | Nr. 4 (jeder nichtleere Wert byte-gleich an den Dienst: unbekannt `ODR-E7003`, 404; Form außerhalb von `SPEC-124` Nr. 4: derselbe Weg) | Nr. 9 (keine Frist, `ctx` unverändert, läuft zu Ende mit Eintrag; ein Kontextfehler bricht ab, ohne Eintrag) | Nr. 12 (zweites `close`: `ODR-E7011`, kein Eintrag; gleichzeitig: ein Eintrag) |
| `DELETE /__admin/sessions/{id}`, `{id}` | gilt nicht: wie `close` | gilt nicht: wie `close` | Nr. 4 und 10 (`ODR-E7003`, `ODR-E7009`, `ODR-E7004`) | Nr. 2, 10 (ein Kontextfehler bricht ab; die Bereinigung ruft der Dienst auch bei beendetem Kontext, `SPEC-124` Nr. 2) | Nr. 12 (zweites `DELETE`: `ODR-E7003`) |
| Audit-Eintrag von `POST` und `close` (in diesem Teil: `close`) | `SPEC-292` Nr. 11 b (jedes Feld je Art gesetzt nach `SPEC-234` Nr. 2) | Nr. 11 e (eine leere Kennung weist `Append` ab, ohne Wirkung auf die Antwort) | Nr. 11 e und *Ungeprüft* (Kennung außerhalb der Form: der Adapter des Ports weist ab) | Nr. 11 a, d (`Append` ohne `ctx`: ein beendeter Kontext ändert den Eintrag nicht, `SPEC-234` Nr. 5 c) | Nr. 12 (je Erfolg genau ein Eintrag, keiner bei `ODR-E7010` und `ODR-E7011`) |

Die Zeilen zu `GET` und `POST /__admin/sessions` stehen in §6 von [slice-admin-sessions](../done/slice-admin-sessions.md), die der Inspektion in [slice-admin-inspektion](../open/slice-admin-inspektion.md). Ohne Nummer steht in den Zellen `SPEC-292`.

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; der Adapter kennt den Modus nicht, `SPEC-124` Nr. 3):

| Paar | Entscheidung | Ort | Fänger |
|---|---|---|---|
| `close`: unbekannt × nicht offen; `DELETE`: unbekannt, offen, Quelle einer Baseline | der Dienst: die Rangfolge von `SPEC-124` Nr. 7 und 11 | `SPEC-292` Nr. 8 | gilt nicht für den Adapter (Rang im Dienst); B9 prüft jede Zelle einzeln |
| Kontextfehler × Fehler mit Code | der Abbruch gilt (`errors.Join` und Ursache) | `SPEC-292` Nr. 8, `SPEC-238` Nr. 6 a 1 | B9 (Ausgang `errors.Join(context.Canceled, *Fehler E7003)`) |
| Antwort und Fehler des Dienstes zugleich | der Fehler gilt | Nr. 8, `SPEC-238` Nr. 6 b | B9 (Ausgang „Wert und Fehler E7003“) |
| Erfolg des Dienstes × Fehler von `Append` | der Erfolg gilt, die Antwort bleibt | Nr. 11 e | B13 |
| zwei gleichzeitige `close` derselben Session | ein Erfolg mit Eintrag, das andere `ODR-E7011` ohne | Nr. 12, `SPEC-124` Nr. 13 | N3 |
| Authentisierung × alles | 401 vor 404 und 405 vor jedem Aufruf einer Route | `SPEC-238` Nr. 3 d | gilt nicht für diesen Plan (Rahmen) |

- Wer schreibt den Audit-Eintrag einer Löschung: der Adapter oder `SessionUseCase` (`slice-audit-log`)? Beide ergäben einen doppelten Eintrag — **Ausgang:** entschieden (Architect, 2026-11-21): allein der Use Case schreibt `session_geloescht` (`SPEC-234` Nr. 7, `SPEC-124` Nr. 7 d1); `DELETE` schreibt nie einen Eintrag (`SPEC-292` Nr. 10, 11 c), gefangen in B10 und M-Audit; Anlegen schreibt der Adapter in Teil 1, Beenden hier (`SPEC-292` Nr. 11, B11).
- `close` wartet auf `Beenden` und kann länger dauern als `SHUTDOWN_TIMEOUT` und als jedes Schreib-Timeout des Servers (`SPEC-124` *Ungeprüft*); bricht der Client ab, wird der Kontext des Requests abgebrochen — **Ausgang:** entschieden (Architect, 2026-11-21): die Antwort kommt nach `Beenden`, ohne Frist und ohne Ablösung des Kontexts; `SPEC-124` Nr. 11 e trägt den Abbruch (`SPEC-292` Nr. 9, B6, B7); die Folge für einen ungeduldigen Client (Session `incomplete`) steht im Vertrag.
- Statuscodes der Erfolge und JSON-Form der Antworten standen nicht in `SPEC-005` — **Ausgang:** entschieden (Architect, 2026-11-21): Status, Form und Fehler stehen in `SPEC-292` Nr. 6 bis 8; eine eigene OpenAPI-Datei entsteht nicht (`SPEC-005`).
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (6×): der Fake von `SessionUseCase` trägt jede Schnittstelle des Originals und zählt jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Marke steht nicht, bevor die Tabelle *Zusage · Fänger* vorliegt — **Ausgang:** offen; Adresse: Planner beim Start (§4): die Tabelle liegt vor, *S* = 1580 liegt unter der Schnittgrenze; der Planner setzt `basis`, `schaetzung` und `marke` vor dem ersten Code-Commit.
- Die Teile teilen die Zeit, `sessionsFake` und `SessionRouten` im selben Paket (Schnitt vom 2026-11-21) — **Ausgang:** entschieden (Planner, 2026-11-21): sie laufen nacheinander (§4, *Reihenfolge und Konflikte*), die Sendung der Mittel und der Gesamtzahlen steht in §1 (*Übernimmt*, Geber `slice-admin-sessions`).
- Gesamtzählungen über die drei Teile (I3: 42 Fälle, I17: 84 Zellen): welcher Teil sie zuletzt hebt, hängt von der Reihenfolge von Teil 2 und Teil 3 ab — **Ausgang:** entschieden (Planner, 2026-11-21): der später in `done/` gehende Teil hebt die Literale (§1, Punkt (e)); Annahme des Planners: keine Änderung am Vertrag, nur an der Zählprüfung der Tests; der Architect bestätigt sie beim nächsten Vertrag dieser Linie.
- Handbuch-Kapitel zu `audit.jsonl`: die Bedingung der Sendung („sobald der Listener verdrahtet ist“, `slice-bootstrap-m6`) tritt nach diesem Teil ein, der Teil lässt Katalog und Handbuch unverändert und `bootstrap` bis auf einen Test — **Ausgang:** weiter offen; Adresse: Planner (Frage an den User, ob `slice-bootstrap-m6` das Kapitel in §1 und DoD übernimmt, weil dort das erste Binary entsteht, das eine Löschung auslösen kann).
- Zählen der Fehler (`SPEC-230` Nr. 8): die Stelle, die den Fehler ausliefert, zählt — das ist der Rahmen von `slice-admin-listener-token` (Sendung dort, §1 *Übernimmt*), nicht jede Route dieses Slice — **Ausgang:** entfallen für diesen Slice (die Routen rufen kein `CountError`, der Rahmen schreibt das Fehlerobjekt); weiter offen bei der Gegenstelle.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-21; der Planner zieht Treffer und Zähler beim Start nach, der Zähler ist die Zahl der Dateien unter `evidence/`). Der Teil schreibt Routen gegen Fakes: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (6×) für die Fakes der Use Cases (beide verkörpert, in `.claude/commands/implement-slice.md` Schritt 14 und `.harness/skills/reviewer.md`), außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (21×, verkörpert). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (32×, `AGENTS.md` §3.9), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (59×, `AGENTS.md` §3.17), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (66×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16). Keine Beobachtung steht mit diesem Slice offen unter 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
