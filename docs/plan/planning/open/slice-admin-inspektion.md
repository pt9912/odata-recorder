# Slice slice-admin-inspektion: Admin-Endpunkte für Exchanges, Coverage und Diagnose (Teil 3 des Schnitts von `slice-admin-sessions`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-FA-11`, `LH-FA-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-005` (Zeilen `exchanges`, `coverage`, `diagnostics`) · `LH-FA-13.a` (Status, Fehlercodes) · `SPEC-293` (Inspektion) mit `SPEC-292` Nr. 2, 4, 7 und 8 (Aufrufe, Pfadwert, JSON, Fehler; gelten für beide Verträge) · `SPEC-236` (Vertrag von `InspectionUseCase`: Nr. 3 bis 8 und *Ungeprüft*) · `SPEC-071`, `SPEC-112` · `SPEC-109` · `SPEC-238` Nr. 6, 9, 15 (Rahmen, nicht geändert) · `ARC-007` · `architecture.md` §2.2, §4.9.
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

**Ziel:** Der Admin-Adapter bildet `GET /__admin/sessions/{id}/exchanges`, `GET /__admin/coverage` und `GET /__admin/diagnostics` aus `SPEC-005` auf `InspectionUseCase` ab: Exchanges einer Session mit den Parametern `limit` und `after`, Coverage einer Baseline mit `baseline_id`, Diagnose; jede Antwort trägt Status, Fehlerobjekt und Korrelations-ID, kein Geheimnis steht in einer Antwort, und die Inspektion verändert nichts (kein Audit-Eintrag, kein Schreiben in einen Port). Der Teil legt `InspektionRouten` an (drei Routen), nutzt die Funktion der Zeit und die JSON-Regeln aus Teil 1 und führt den zählenden Fake `inspektionFake`.

**Vertrag (Architect, 2026-11-21):** `SPEC-293` (Exchanges, Coverage, Diagnose); der Vertrag steht auf `main`, bevor dieser Teil startet.

**Übernimmt (Sendung von `slice-inspection-use-case` und `slice-inspection-coverage-diagnose`, Teil des Schnitts von `slice-admin-sessions`, `AGENTS.md` §3.13):** die Endpunkte für Exchanges, Coverage und Diagnose; beide Slices liegen in `done/` ([slice-inspection-use-case](../done/slice-inspection-use-case.md), [slice-inspection-coverage-diagnose](../done/slice-inspection-coverage-diagnose.md)), und ihr §1 *Ausdrücklich NICHT* weist die Endpunkte, Pfade und Statuscodes dem Admin-Adapter zu.

**Übernimmt (Sendung von `slice-inspection-use-case`, Rückgabe des Architect vom 2026-11-07, `AGENTS.md` §3.13; Teil des Schnitts von `slice-admin-sessions`):** die Parameter und die Antwort auf einen ungültigen Parameter. `GET …/exchanges` trägt `limit` und `after` (Zahlen, `SPEC-236` Nr. 3; `limit` fehlt: Standard, `after` fehlt: Anfang), `GET /__admin/coverage` trägt `baseline_id` (Pflicht, `SPEC-236` Nr. 6). Ein negativer Wert, eine nicht lesbare Zahl und ein fehlendes oder leeres `baseline_id` ergeben `model.ErrUngueltig` ohne Code vom Use Case (oder vom Adapter beim Lesen der Zahl, ohne Aufruf des Use Case); die Antwort darauf ist `ODR-E2102` mit 400 (Entscheidung des Architect im Vertrag des Rahmens, `SPEC-238` Nr. 6 und 15 a; der Code besteht seit `slice-admin-rahmen`); ein eigener Code entsteht hier nicht. Die Fehler `ODR-E7003` (unbekannte Session bei Exchanges) und `ODR-E7002` (unbekannte Baseline) liefert der Use Case mit Code und Status (`SPEC-236` Nr. 8); der Adapter bildet sie ab.

**Übernimmt (Sendung von `slice-admin-sessions`, Teil 1; Geber: `slice-admin-sessions`; Eintrag des Planners vom 2026-11-21 beim Schnitt, `AGENTS.md` §3.13):** (a) die Funktion der Zeit aus Teil 1 (RFC 3339 in UTC mit neun Nachkommastellen, die Nullzeit ist `null`, M-Zeit) für `started_at` und `finished_at`; dieser Teil liefert sie nicht; (b) die JSON-Regeln von `SPEC-292` Nr. 7 und die Abbildung der Fehler von Nr. 8 (Zeilen S9 und S11 in Teil 1); (c) der Fänger S16 (AST-Test über die Dateien der Routen) von Teil 1 und die Zeile I17 dieses Teils prüfen dieselben drei Namen (`NeuesFehlerobjekt`, `CountError`, `ArtAdminAbgewiesen`): die Dateien dieses Teils stehen in der Dateimenge von I17, die Zählprüfung nennt sie; der AST-Test von Teil 1 bildet seine Menge aus dem Verzeichnis (jede Datei `.go` ohne `_test.go` außer den fünf des Rahmens), die Dateien dieses Teils gehören dazu, ohne dass jener Test geändert wird; (d) die Gesamtzahlen über alle drei Teile, die zwei Zeilen dieses Teils als Literal tragen (I3: 42 Fälle mit 7 Treffern, I17: 84 Zellen mit genau 2 Einträgen): läuft dieser Teil vor Teil 2 ([slice-admin-sessions-beenden](../in-progress/slice-admin-sessions-beenden.md)), zählen seine Tests mit den Routen von Teil 1 und 3 allein 28 Fälle mit 5 Treffern (4 Pfadformen) und 57 Zellen mit einem Eintrag, und Teil 2 hebt sie auf 42 mit 7 und 84 mit 2 (dessen §1); läuft dieser Teil nach Teil 2, schreibt er 42 und 84.

**Übernimmt (Sendung von `slice-admin-listener-token`, Pflichten der Endpunkte nach `SPEC-238` Nr. 15, Schnitt vom 2026-11-07, `AGENTS.md` §3.13; Teil des Schnitts von `slice-admin-sessions`):** (a) was die Route als formal ungültig erkennt oder von einem Dienst als `model.ErrUngueltig` ohne Code erhält (ungültiger Seitenparameter `limit` oder `after`: nicht numerisch, negativ, außerhalb des Bereichs von `int`, fehlendes `baseline_id`), gibt sie als Fehler zurück (`ErrUngueltig` oder `*inbound.Fehler{Code: CodeE2102, Status: 400}`); (b) sie ruft `CountError` nie und `Audit.Append` nie für `ODR-E2101` und schreibt kein Fehlerobjekt selbst; (c) die vier Arten von `SPEC-234` Nr. 2 schreibt der Dienst, nie die Route; die Inspektion schreibt keinen Eintrag; (d) eine Erfolgsantwort hat einen Status von 200 bis 299 und trägt Daten als `Antwort.Body`. Der Fänger liegt in diesem Slice (die Routen und ihre Tests); der Rahmen hat keine Route (`slice-admin-abweisen-audit`, Zeile P1, akzeptiertes Negativ). Der Reviewer löst jede Zeile von Nr. 15 mit einer Mutation auf.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Endpunkte für Sessions (`GET`, `POST`, `close`, `DELETE`) — [slice-admin-sessions](../done/slice-admin-sessions.md) (Teil 1) und [slice-admin-sessions-beenden](../in-progress/slice-admin-sessions-beenden.md) (Teil 2); dieser Teil teilt mit ihnen das Paket und die Zeit, nicht den Gegenstand.
- Imports, Workspaces und Snapshots — `slice-admin-imports` und `slice-admin-workspaces`.
- Der Rahmen (Listener, Token, Fehlerobjekt, Abbildung der Port-Fehler) — `slice-admin-listener-token`; dieser Teil trägt die Routen und ihre Abbildung.
- Bildung der Seite (Zeiger `sequence_no`), Coverage einer Baseline, Allowlist der Einstellungen und Zähler je Fehlercode — `InspectionUseCase` (`slice-inspection-use-case`, `slice-inspection-coverage-diagnose`, beide in `done/`); der Adapter bildet ab und liest nie aus der Datenbank (Architektur §2).
- Ein Audit-Eintrag der Inspektion — es gibt keinen: die Inspektion verändert nichts (`SPEC-293`), `AuditLog` ist kein Parameter von `InspektionRouten` (I1, I17).
- Das Zählen der Fehler je Code — der Rahmen von `slice-admin-listener-token`, der das Fehlerobjekt schreibt (`SPEC-230` Nr. 8); die Routen rufen `CountError` nie.

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

- [ ] Exchanges, Coverage und Diagnose über `InspectionUseCase`: die Parameter `limit`, `after` und `baseline_id` nach `SPEC-236` Nr. 3, 4 und 6, die Antwort auf einen ungültigen Parameter (negativer Wert, nicht lesbare Zahl, fehlendes `baseline_id`, doppelter oder leerer bekannter Parameter) ist `ODR-E2102` mit 400 (`SPEC-238` Nr. 15 a, Geber `slice-admin-listener-token`), Fehlerabbildung (`ODR-E7003`, `ODR-E7002`), Form der Antworten nach `SPEC-293` Nr. 6 bis 8, kein Geheimnis in der Antwort (I1 bis I16, N5).
- [ ] Pflichten der Endpunkte nach `SPEC-238` Nr. 15 (a) bis (d) für jede der drei Routen: kein `CountError`, kein `Append` für `ODR-E2101`, kein selbst geschriebenes Fehlerobjekt, Erfolgsstatus 200 bis 299, je Route gefangen (I17 und die Fälle von I16; Geber `slice-admin-listener-token`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/adapters/driving/admin/` | update | `InspektionRouten` (Dateinamen sind Vorschlag: `inspektion.go`, `inspektion_query.go`); die Funktion der Zeit liefert Teil 1; keine Änderung an `admin.go`, `route.go`, `fehler.go`, `lebenszyklus.go` und an den Routen der Teile 1 und 2 (die Dateimenge des AST-Tests bildet Teil 1 aus dem Verzeichnis, eine neue Datei dieses Teils gehört dazu, ohne dass der Test sie nennt) |
| `src/bootstrap/`, `src/hexagon/ports/support/`, `src/adapters/driven/audit/` | keine Änderung | der Teil führt keinen Fehlercode, Schlüssel, kein Kommando und keine Art ein; das Binary bleibt unverändert (I18) |
| `spec/spezifikation.md` | geliefert (Architect, auf `main`) | `SPEC-293`; dieser Teil schreibt sie nicht |
| Tests | neu | Adapter gegen `inspektionFake` und `FakeAudit`: jedes Argument festgehalten, jede Operation gezählt, auch die, die der Adapter nie rufen darf (`AGENTS.md` §3.15); ein Fall je Aufrufstelle für den `ctx` des Requests |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem 2026-11-21 (Architect) mit dem Vertrag in der Spezifikation (`SPEC-293`) vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. `Nr.` ist die Kennung des ungeschnittenen Plans: **I** gehört zu diesem Teil (Exchanges, Coverage, Diagnose), **N5** ist seine nebenläufige Zeile; die Zeilen **S**, **N1**, **N2** stehen in Teil 1, **B**, **N3**, **N4** in Teil 2. Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest.

**Entscheidungen des Architect für diesen Teil** (2026-11-21, vor dem ersten Code-Commit, `AGENTS.md` §3.12; Ort in der Spezifikation; die übrigen Aufträge stehen in §3 von [slice-admin-sessions](../done/slice-admin-sessions.md)):

1. **Statuscodes, JSON-Form, Fehlerabbildung:** `SPEC-293` Nr. 6 bis 9. Erfolg 200; Form der Zeile, der Coverage (`evidence` mit denselben Bytes) und der Diagnose nach dem Vertrag; eine ungültige Evidenz ist `ODR-E9001` (500), keine Teilantwort.
2. **Parameter:** `SPEC-293` Nr. 3 bis 5. Ungültig sind (alle `ODR-E2102`, 400, ohne Aufruf des Dienstes): eine nicht zerlegbare Query, ein doppelter oder leerer bekannter Parameter, eine Zahl außer ASCII-Ziffern, ein Wert über dem größten `int`, ein fehlendes `baseline_id`. `{id}` geht byte-gleich an den Dienst (`SPEC-292` Nr. 4). Ein neuer Code entsteht nicht.
3. **Fehlversuche am Token** (Entscheidung des Users, Welle §5, „Kein Code, bis entschieden“): sie berührt diesen Plan nicht; die Routen laufen erst nach der Authentisierung (allein der Rahmen, `SPEC-238` Nr. 8).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-21, Zeilen I1 bis I18 und N5 des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| I1 | (`SPEC-293` Nr. 1) `InspektionRouten` weist `inspektion` mit nil ab (`model.ErrUngueltig`, keine Routen, kein Wert im Text) und liefert drei Routen in der Reihenfolge des Vertrags; es ruft keinen Port, startet keine Goroutine und hat weder `AuditLog` noch `Telemetry` | wie S1 für das eine Feld; `inspektionFake` zählt 0 Aufrufe; die Funktion hat einen Parameter (Test mit `reflect` auf den Funktionswert) | Prüfung entfernt; zwei Routen vertauscht; `Diagnose` zur Probe gerufen | andere |
| I2 | (Nr. 2) Jede Route ruft genau eine Operation mit `ctx` unverändert und den Argumenten, sonst keine: `Exchanges(ctx, id, Seite)`, `Coverage(ctx, baselineID)`, `Diagnose(ctx)` | `inspektionFake` zählt die drei Operationen: 3 Routen × 3 Zähler = 9 Zellen, Diagonale 1; Wert im `ctx` kommt an, Gleichheit des `ctx` | `ctx` abgeleitet; `Coverage` bei `exchanges` zur Prüfung gerufen | andere |
| I3 | (Nr. 1; `SPEC-238` Nr. 2) Zuordnung: `GET` trifft die drei Pfadformen, jede andere Methode ist 405 mit `Allow: GET` | M-Route: 3 Pfadformen × 7 Methoden = 21 Fälle (3 Treffer, 18 mal 405), zusammen mit den Routen von `SessionRouten` 6 Pfadformen × 7 = 42 Fälle (7 Treffer); Orakel Literale; Zählprüfung gegen die Muster der beiden Konstruktoren | `Allow` fehlt; Muster `/exchange` | andere |
| I4 | (Nr. 5) Der Pfadwert `{id}` von `exchanges` geht byte-gleich an `Exchanges`, ohne Prüfung | M-Pfadwert: 13 Werte wie B3 an einer Route = 13 Fälle; Orakel wie B3 | wie B3 | andere |
| I5 | (Nr. 2, 3) Query: die ganze Query muss zerlegbar sein; unbekannte Schlüssel sind ohne Wirkung (auch mehrfach, auch `limit` an `coverage` und `baseline_id` an `exchanges`, auch in anderer Schreibweise); `diagnostics` liest die Query nicht | M-Query: 6 ungültige Formen (`%zz`, `%`, `a=%`, `a=1;b=2`, `%2`, `limit=1&%zz`) × 2 Routen = 12 Fälle (400, 0 Aufrufe), dieselben 6 an `diagnostics` (200), unbekannte Schlüssel 4 Formen (`x=1`, `x=1&x=2`, `Limit=1`, der bekannte Schlüssel der anderen Route) × 2 Routen = 8; Zählprüfung 26 | Query bei `diagnostics` geprüft; unbekannter Schlüssel abgewiesen; ungültige Prozentfolge in einem unbekannten Schlüssel übergangen | andere |
| I6 | (Nr. 4) Gültige Zahl: Ziffernfolge, führende Nullen, größter `int`; fehlt `limit`, ist `Seite.Limit` 0, fehlt `after`, ist `Seite.Nach` 0; `limit=0` ist 0; Reihenfolge und Stellung ändern nichts | M-Zahl gültig: 8 Werte (`0`, `1`, `007`, `100`, `1000`, `1001`, `%31%30`, größter `int`) × 2 Parameter = 16 Fälle; M-Stellung: 5 Formen (`limit=3`, `after=5`, `limit=3&after=5`, `after=5&limit=3`, `x=1&limit=3&y=2&after=5`) und kein Parameter = 6; Orakel Literale; Zählprüfung 22 | `limit=0` als ungültig; `007` abgewiesen; `after` bei fehlendem `limit` verworfen; `int32` | andere |
| I7 | (Nr. 4) Ungültige Zahl: 400 `ODR-E2102`, 0 Aufrufe | M-Zahl ungültig: 17 Werte (leer, `-1`, `-0`, `+1`, `%2B1`, `%201`, `1%20`, `1+0`, `1.0`, `1e3`, `0x10`, `1_0`, `abc`, eine arabisch-indische Ziffer und eine vollbreite Ziffer (beide aus Bytewerten), größter `int` plus 1, 25 Neunen) × 2 Parameter = 34 Fälle, dazu 4 Paare (beide Parameter ungültig, beide Reihenfolgen, je mit `-1` und `1.0`) = 38; Orakel: Grammatik und `math/big` im Test, ohne Code des Lesers; der Fake würde `ODR-E7003` liefern; Zählprüfung 38 | `strconv.Atoi` allein (nimmt `+1` und `-1`); `ParseInt` mit Basis 0; Überlauf gekappt; Prüfung der Ziffern mit `unicode.IsDigit` | andere |
| I8 | (Nr. 3) Doppelt und leer: ein bekannter Schlüssel zweimal oder mit leerem Wert ist 400, ein unbekannter doppelter nicht | 3 bekannte Schlüssel (`limit`, `after`, `baseline_id`) × 6 Formen (`k=1&k=2`, `k=1&k=1`, `k=1&k=`, `k=&k=1`, `k=1&k`, `k=1&x=2&k=3`) = 18 Fälle, leer (`k=`, `k`) 3 × 2 = 6, `x=1&x=2` an beiden Routen = 2; Zählprüfung 26 | der letzte Wert gilt; der erste gilt; leerer Wert als fehlend | andere |
| I9 | (Nr. 5) `baseline_id` ist Pflicht (fehlt: 400); jeder andere dekodierte Wert geht byte-gleich an `Coverage` | M-baseline: 9 Werte (`x`, `a%20b`, `a+b` (ergibt Leerzeichen), `%00`, `%FF`, `a%2Fb`, `A`, `%20`, 300 mal `a`) und fehlt = 10 Fälle; Orakel die dekodierten Bytes (aus Bytewerten); der Fake würde `ODR-E7002` liefern | Wert getrimmt; leerer Wert als gültig; Wert in Kleinbuchstaben | andere |
| I10 | (Nr. 6) Exchanges: 200, Felder und Reihenfolge, `limit`, `total`, `next` (`null` bei 0), `[]` | M-Felder: jedes der 15 Felder von `inbound.ExchangeZeile` (Zählprüfung `NumField` gleich 15) und der 4 von `ExchangeSeite` (gleich 4) mit eigenem Wert, Orakel die Namen als Literale aus `SPEC-236` Nr. 5; `next` 0 und 7; 0, 1, 3 Zeilen und `nil`; Zählprüfung | zwei Felder vertauscht; `next` als 0; `total` fehlt; Name `sequence` | andere |
| I11 | (Nr. 6; `SPEC-292` Nr. 7) Werte der Zeile: Zeit (`finished_at` Nullzeit `null`), `response_status` 0 ist `null`, Wahrheitswert, Zeichen im Pfad | M-Zeit: 7 Werte an `started_at` und `finished_at` = 14, `response_status` 0 und 200 = 2, `key_redacted` beide Werte = 2, Pfad mit 6 Zeichenformen (`<&>`, `"`, Backslash, Byte `0x01`, Byte `0xFF` (erwartet `U+FFFD`), `ä`; aus Bytewerten) = 6; 24 Fälle | `finished_at` als `0001-…`; `response_status` 0 als 0; Pfad als Bytes durchgereicht | andere |
| I12 | (Nr. 7) Coverage: 200, `{"entries":[…]}`, drei Felder, Reihenfolge, `[]`; `evidence` mit denselben Bytes | M-Evidenz gültig: 6 Formen (kanonisch klein, mit `<&>`, mit `U+2028` als UTF-8-Bytes, mit einer Escape-Folge im Text (aus Bytewerten), 64 Ebenen tief, 64 KiB groß); Orakel `bytes.Contains(body, evidenz)`; 0, 1, 3 Einträge und `nil`; M-Felder: 3 Felder von `CoverageEintrag` (Zählprüfung) | `json.Marshal` des `RawMessage` (maskiert `<`, `>`, `&`); `json.Indent`; `evidence` als Zeichenkette | andere |
| I13 | (Nr. 7) Ungültige Evidenz: 500 `ODR-E9001`, kein Teil der Antwort, kein Wert im Text | M-Evidenz ungültig: 4 Formen (leer, `{`, `{} {}`, `x`); Orakel Literale | ungültige Evidenz eingesetzt; 400 statt 500 | andere |
| I14 | (Nr. 8) Diagnose: 200, `mode`, `settings` mit sortierten Schlüsseln und `{}`, `errors` in der Reihenfolge des Dienstes mit `code` und `count` bis `MaxUint64`, `[]` | `mode` 4 Werte (`record`, `replay`, `mock`, `hybrid`); `settings` mit 0, 1, 3 Einträgen (`b`, `a`, `B`: Reihenfolge der Bytes) und `nil`; `errors` mit `len(model.AlleCodes())` Einträgen und `nil`, `count` 0, 1, `math.MaxUint64`; Zählprüfung gegen `NumField` von `inbound.Diagnose` (3) und `FehlerZaehler` (2) | `settings` unsortiert; `count` als `float64` (verliert Stellen); `null` statt `{}` | andere |
| I15 | (Nr. 8, 11) Akzeptiertes Negativ: die Route filtert `settings` nicht; ein Schlüssel außerhalb der Allowlist, den ein Dienst lieferte, erschiene in der Antwort | ein Test belegt das Durchreichen (`settings` mit `UPSTREAM_URL` erscheint unverändert); gegen den Inhalt kein Fänger. Grund: die Allowlist gehört dem Dienst (`SPEC-236` Nr. 7), eine zweite Liste wäre ein zweiter Ort derselben Entscheidung. Folge: ein fehlerhafter Dienst zeigte einen Schlüssel; der Fänger der Liste liegt beim Dienst. Maß: 18 von 18 erlaubten Schlüsseln der Liste, 0 geprüft in der Route | nicht anwendbar | andere |
| I16 | (Nr. 9; `SPEC-292` Nr. 8) Fehlerabbildung der Inspektion wie S11 | M-Fehler: 3 Routen × 8 allgemeine Ausgänge + `exchanges` 3 (`ODR-E7003`, `ODR-E3001`, `ODR-E9001`) + `coverage` 3 (`ODR-E7002`, `ODR-E3001`, `ODR-E9001`) = 30 Zellen; Orakel Literale; ein Verstoß der Query geht jedem Fehler des Dienstes vor (der Fake liefert `ODR-E7003`) | `ODR-E7002` als 400; Query nach dem Aufruf des Dienstes geprüft | andere |
| I17 | (Nr. 1, 2; `SPEC-292` Nr. 11 c) Die Inspektion verändert nichts: `Append` 0, kein `CountError` und kein Fehlerobjekt der Route | M-Audit: 3 Routen: 12 + 12 + 9 = 33 Zellen, `FakeAudit.Aufrufe()` gleich 0; AST-Test wie S16 über die Dateien der Inspektion; Gesamtzählung aller Teile: 24 + 27 + 33 = 84 Zellen, genau 2 mit Eintrag | Route schreibt `Append`; Route ruft `CountError` | andere |
| I18 | (Nr. 12) Der Teil führt keinen Fehlercode, Schlüssel, Kommando oder keine Art ein, das Binary bleibt unverändert | wie S17 (benannte Ausgabe `make katalog-check`, Importe von `bootstrap`) | `InspektionRouten` aus `bootstrap` gerufen | andere |
| N5 | (Nr. 10) Gleichzeitige Requests (auch zwei Seiten derselben Session) beeinflussen einander nicht, die Routen halten keinen Zustand | 16 Goroutinen je Route, jede mit eigener Eingabe, Orakel: die eigene Antwort; `make test-race`, je Prozess `-test.count=1` | Paketvariable für die letzte Seite oder `next` | nebenläufig |

Zeilen: I1 bis I18, N5 gleich 19 (*a* = 1, *b* = 18).

**Verteilung beim Schnitt** (Planner, 2026-11-21): Die Zeilen sind im Wortlaut des Architect übernommen. Wo eine Zeile auf eine Zeile eines anderen Teils verweist (I1 auf S1, I4 auf B3, I16 auf S11, I17 auf S16 und die Gesamtzählung, I18 auf S17), gilt deren Wortlaut in den Plänen von Teil 1 ([slice-admin-sessions](../done/slice-admin-sessions.md)) und Teil 2 ([slice-admin-sessions-beenden](../in-progress/slice-admin-sessions-beenden.md)). I3 und I17 nennen Gesamtzahlen über alle drei Teile: §1 (*Übernimmt*, Punkt (d)) regelt, welcher Stand sie trägt, wenn dieser Teil vor Teil 2 läuft. Die Fänger der Parameter (I5 bis I9) und der Evidenz (I12, I13) gehören diesem Teil allein.

**Mengen mit Zählprüfung** (Generator oder Schleife, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15; Architect, 2026-11-21, je Teil getrennt; das Orakel trägt Statuscodes, Codes, Namen und Bytes als Literale aus dem Vertrag, nie als Konstanten des Codes und nie als Aufruf des Lesers):

| Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|---|
| M-Route (I3) | Pfadformen 3 (`…/{id}/exchanges`, `/__admin/coverage`, `/__admin/diagnostics`) × Methoden 7 (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS) = 21, davon 3 Treffer; mit den Routen von `SessionRouten` 6 Pfadformen = 42, davon 7 Treffer | Schleife | Status, `Allow` je Pfadform (`GET`, `GET`, `GET`) als Literale | Pfadformen gleich der Zahl der verschiedenen Muster der Routen, Routen 3 gleich der Länge der Liste des Konstruktors (Stand mit Teil 2: 6 und 7 für beide Konstruktoren) |
| M-Pfadwert (I4) | Werte 13 (siehe B3) × Route mit `{id}` 1 (`exchanges`) = 13 | Schleife über die Werte aus Bytewerten | die dekodierten Bytes als Literale | 13 gleich der Länge der Liste, 1 gleich der Zahl der Muster mit `{id}` |
| M-Query (I5) | Prozent- und Trennerformen 6 (`%zz`, `%`, `a=%`, `a=1;b=2`, `%2`, `limit=1&%zz`) × Routen mit Query-Prüfung 2 (`exchanges`, `coverage`) = 12 (400); dieselben 6 Formen an `diagnostics` = 6 (ohne Wirkung); unbekannte Schlüssel 4 × 2 = 8 | Schleife | Status und Argumente am Fake als Literale | Summe der Achsen (26) |
| M-Zahl (I6, I7) | gültige Werte 8 und ungültige 17 (siehe I6, I7) × Parameter 2 (`limit`, `after`) = 50, dazu 4 Paare ungültiger Werte; Stellung 5 Formen und kein Parameter = 6; Achse *Weg zum Aufruf*: `Seite{Limit, Nach}` am Fake | Schleife | Grammatik der Ziffern und `math/big` (nicht `strconv`) | 8 + 17 gleich der Länge der Liste, 2 Parameter; die Zahl des größten `int` aus `math.MaxInt` |
| M-Doppelt (I8) | bekannte Schlüssel 3 × Formen 6 = 18; leer 3 × 2 = 6; unbekannt doppelt 2 | Schleife | 400 und 0 Aufrufe; 200 | 26 |
| M-baseline (I9) | Werte 9 und fehlend | Schleife | die dekodierten Bytes als Literale | 10 |
| M-Zeit (I11) | Werte 7 (`10:00:00.123456789Z`, neun Nullen, `.100000000`, `10:00:00+02:00` gleich `08:00:00Z`, Nullzeit, Jahr 1 mit 1 ns, Jahr 9999) an `started_at` und `finished_at` | Schleife | Literale der Zeichenkette (nie `time.Format` im Orakel) | 7, Zahl der Zeitfelder 2 gegen die Felder `time.Time` von `ExchangeZeile` (`reflect`) |
| M-Felder (I10, I12, I14) | Felder der Typen des Ports: `ExchangeZeile` 15, `ExchangeSeite` 4, `CoverageEintrag` 3, `Diagnose` 3, `FehlerZaehler` 2 = 27, jedes mit eigenem Wert | Schleife über `reflect`, Wert je Feldindex | Namen und Reihenfolge als Literale des Vertrags | `NumField` je Typ gegen die Zahl der Literale (15, 4, 3, 3, 2) |
| M-Fehler (I16) | allgemeine Ausgänge 8 × Routen 3 = 24, Codes der Tabelle von `SPEC-293` Nr. 9: `exchanges` 3, `coverage` 3, `diagnostics` 0 = 6; zusammen 30 | Schleife | Status, Code und Body als Literale aus `SPEC-109` Nr. 2 | 30 gleich 3 × 8 + 6; die Codes gegen `model.AlleCodes()` (jeder Code der Tabelle ist darin) |
| M-Audit (I17) | Zellen Route × Ausgang: `exchanges` 12, `coverage` 12, `diagnostics` 9 = 33, in jeder `FakeAudit.Aufrufe()` gleich 0; mit den 51 Zellen der Teile 1 und 2 gleich 84, genau 2 mit 1 Eintrag, 82 mit 0 | Schleife über M-Fehler und je einen Erfolg | `FakeAudit.Aufrufe()` und `Eintraege()` als Literale | 33 gleich 3 × 9 + 6; 84 gleich 7 × 9 + 21 (Stand mit Teil 2) |
| M-Evidenz (I12, I13) | gültig 6, ungültig 4 | Schleife | `bytes.Contains` und Status | 6 + 4 |

**Unbeschränkte Mengen und ihre Prüfgrenze (akzeptierte Negative, Grund und Folge):** die Länge von `{id}` und `baseline_id` (Prüfgrenze 65 und 300 Bytes; Grund: kein Zweig hängt an der Länge; Folge: jede Länge bis zur Header-Grenze von 64 KiB erreicht den Dienst unverändert); die Zeilen einer Seite und die Einträge der Coverage und der Diagnose (Prüfgrenze 3; Folge: der Body wächst mit der Menge, die der Dienst liefert, bei der Seite höchstens `MaxLimit` Zeilen); die Größe der Evidenz (Prüfgrenze 64 KiB; Folge: der Body wächst mit ihr, ohne Grenze der Route); die Zahl der gleichzeitigen Requests (Prüfgrenze 16, N5; Folge: die Routen halten keinen Zustand, die Zahl ändert nichts).

**Zweige als Menge, Runden unter dem Detektor** (Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat in den neuen Dateien dieses Teils null Blöcke mit Zähler 0; der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Z-1 Konstruktor: nil-Port, Routen in der Reihenfolge | I1 | ein Lauf |
| Z-3 Leser der Query: Zerlegung scheitert, Schlüssel doppelt, leer, fehlend, Grammatik der Zahl, Überlauf | I5 bis I9 | ein Lauf |
| Z-4 Schreiber des JSON: Zeit (Nullzeit, anderer Ort), Liste `nil`, `next` `null`, `response_status` `null`, `evidence` gültig und ungültig, `settings` leer und gefüllt | I10 bis I14 | ein Lauf |
| Z-5 Handler: Erfolg, Fehler des Dienstes | I16 | ein Lauf |
| Z-6 Rückgaben, die nur bei einem Fehler von `json.Marshal` über einer Zeichenkette oder Zahl greifen | kein Fänger (unerreichbar) | der einzige erlaubte Ausreißer, mit Datei und Funktion im Bericht |

Nebenläufigkeit: N5; der Zustand nach dem Konstruktor ist schreibgeschützt.

**Aufrufstellen und zählende Fakes** („jeder Fall je Aufrufstelle“ gilt für diese drei Stellen):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| `Exchanges` | `InspectionUseCase` | I2, I4, I6 bis I8, I10 | `inspektionFake` zählt die drei Operationen, hält `ctx` und jedes Argument (auch `Seite`), liefert Ergebnis und Fehler einstellbar, keine Panik |
| `Coverage` | `InspectionUseCase` | I2, I9, I12, I13 | wie oben |
| `Diagnose` | `InspectionUseCase` | I2, I14 | wie oben |

Der Fake `inspektionFake` ist test-lokal (*f* = 1); `FakeAudit`, `FakeTelemetry` und die Fakes der Kennungen bestehen. `FakeAudit` dient allein der Zählung 0 (I17), der Konstruktor nimmt kein `AuditLog`. Ein Test-Server mit echtem Client gehört nicht dazu: die Tests rufen `ServeHTTP` mit `httptest`.

**Bindung an die Ports** (Lerneintrag V-2): `src/adapters/driving/admin/bindung_test.go` ergänzt `var _ inbound.InspectionUseCase = (*inspektionFake)(nil)`; `InspektionRouten` nimmt den Schnittstellentyp des Ports als Parameter (nicht eine Funktion), so übersetzt eine Mutation einer Signatur das Paket nicht mehr.

**Größe und Schnitt:** steht in §3 von [slice-admin-sessions](../done/slice-admin-sessions.md) (Architect, 2026-11-21; Zahlen des Planner in §4 dort und hier). Dieser Teil: *a* = 1, *b* = 18, *f* = 1, *p* = 300 (Konstruktor 25, Handler 90, Leser der Query 80, Schreiber 105), *S* = 2025, *M* = 2800. Die Teile teilen einen Mechanismus (Zeit, Paket) und laufen nacheinander (§4, *Reihenfolge und Konflikte*).

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

**Start** (`next` → `in-progress`): [slice-admin-sessions](../done/slice-admin-sessions.md) (Teil 1) liegt in `done/`; mit ihm liegen die Bedingungen von Teil 1 in `done/` (`slice-admin-listener-token` samt seinen Teilen `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit`, `slice-audit-log` und `slice-audit-dienste`); dazu `slice-inspection-use-case` und `slice-inspection-coverage-diagnose` (der Schnitt des Architect vom 2026-11-07: Exchanges, dann Coverage und Diagnose). Teil 2 ([slice-admin-sessions-beenden](../in-progress/slice-admin-sessions-beenden.md)) ist keine Bedingung. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): `SPEC-293` liegt vor (Architect, 2026-11-21).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-admin-sessions` (Teil 1): *Code-Abhängigkeit*. Die Funktion der Zeit und die JSON-Regeln liegen dort (`SPEC-292` Nr. 7); dieser Teil nutzt sie im selben Paket.
- `slice-admin-listener-token` und seine Teile: *Code-Abhängigkeit*. Der Rahmen (`Route`, Zuordnung, Fehlerabbildung) trägt die Routen.
- `slice-inspection-use-case`, `slice-inspection-coverage-diagnose`: *Code-Abhängigkeit, Port*. `InspectionUseCase` mit `Exchanges`, `Coverage` und `Diagnose`.
- `slice-audit-log`: nur für den Fake `FakeAudit` der Zählung 0 (I17).

**Nicht gebraucht:** `slice-admin-sessions-beenden` (Teil 2), `welle-mock`, M2, `slice-http-adapter`, `slice-workspace-use-case` und `slice-admin-tls`; der Adapter spricht gegen Fakes, TLS umhüllt den Listener unabhängig von den Routen.

**Reihenfolge und Konflikte** (Planner, 2026-11-21): Die drei Teile des Schnitts von `slice-admin-sessions` schreiben im selben Paket `src/adapters/driving/admin/` und teilen die Zeit; die Messung zählt jede Änderung an den Pfaden (`SPEC-128` Nr. 10 g). Sie laufen nacheinander in derselben Linie, nie zugleich: Teil 1, dann Teil 2, dann dieser Teil; dieser Teil hängt nur an Teil 1 und darf vor Teil 2 laufen (dann gilt §1, *Übernimmt*, Punkt (d)). Dieser Teil ändert nichts unter `src/bootstrap/`, `src/hexagon/ports/support/` und `src/adapters/driven/audit/`. Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-21): der Block unten trägt für diesen Teil 19 Zeilen der Tabelle (I1 bis I18, N5), *a* = 1 (N5), *b* = 18, *f* = 1 (`inspektionFake`), *p* = 300: *S* = 85·1 + 80·18 + 200·1 + 300 = 2025; 135·2025 = 273375, *M* = 100 · ⌈27,34⌉ = 2800; keine Schnittfrage (135·*S* = 273375 ≤ 300000, *S* < 2223). Der Teil liegt 198 unter der Schnittgrenze: jede weitere Zeile der Tabelle kostet 80, drei weitere lösen die Schnittfrage neu aus (*S* = 2265); wächst die Tabelle im Code (`AGENTS.md` §3.9), prüft der Planner sie neu. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. `basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit (`AGENTS.md` §3.14); bis dahin steht die Marke aus (Prosa-Marke und Block ersetzen einander nicht, der Block steht allein).

```groessenmarke
pfad: src/adapters/driving/admin/
produktcode: 300
fakes: 1
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt (`AGENTS.md` §3.14); ein weiterer Schnitt ist nicht vorgesehen, die Schätzung liegt unter der Schnittgrenze (198).
- `in-progress` → `open` (blockiert): eine Randform ohne Ort im Vertrag, den der Implementer nicht entscheidet (`AGENTS.md` §3.12).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): jede Zelle nennt eine Nummer von `SPEC-293` oder `SPEC-292` (der Architect, 2026-11-21; Zeilen der Inspektion des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen) oder „gilt nicht“ mit Grund; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `GET /__admin/sessions/{id}/exchanges`, `{id}` | gilt nicht: ein leeres Segment passt auf keine Route (404 `ODR-E2102`, `SPEC-238` Nr. 2) | gilt nicht: wie fehlend | `SPEC-293` Nr. 5 (byte-gleich; unbekannt `ODR-E7003`, 404) | `SPEC-293` Nr. 10 | `SPEC-293` Nr. 10 (Seitenfolge, `SPEC-236` Nr. 4) |
| `GET …/exchanges`, `limit` | `SPEC-293` Nr. 4 (fehlt: `Seite.Limit` 0, Standard des Dienstes) | Nr. 3 (`limit=` und `limit`: 400) | Nr. 3, 4 (doppelt, nicht numerisch, negativ, Überlauf: 400); über `MaxLimit`: der Dienst kappt (`SPEC-236` Nr. 3) | gilt nicht: die Prüfung ist synchron und ohne Kontext | gilt nicht: kein Zustand (Nr. 10) |
| `GET …/exchanges`, `after` | Nr. 4 (fehlt: `Seite.Nach` 0, am Anfang) | Nr. 3 (400) | Nr. 3, 4 (wie `limit`); hinter der letzten `sequence_no`: leere Seite (`SPEC-236` Nr. 4) | gilt nicht: wie `limit` | gilt nicht: kein Zustand (Nr. 10) |
| `GET /__admin/coverage`, `baseline_id` | Nr. 5 (Pflicht: 400) | Nr. 5 (leer: 400) | Nr. 3, 5 (doppelt: 400; sonst byte-gleich; unbekannte Baseline und Snapshot anderer Art: `ODR-E7002`, 404) | Nr. 10 | Nr. 10 (liest neu) |
| `GET /__admin/diagnostics` | gilt nicht: kein Argument; die Query ist ohne Wirkung (Nr. 2) | Nr. 2 | Nr. 2 (jede Query ohne Wirkung); fremde Methode: `SPEC-238` Nr. 2 | Nr. 10 und `SPEC-236` Nr. 7 (der Dienst prüft `ctx`) | Nr. 10, `SPEC-236` Nr. 7 (jeder Aufruf liest neu) |

Ohne Nummer steht in den Zellen `SPEC-293`. Die Zeilen zu den Sessions stehen in §6 von [slice-admin-sessions](../done/slice-admin-sessions.md) und [slice-admin-sessions-beenden](../in-progress/slice-admin-sessions-beenden.md).

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; der Adapter kennt den Modus nicht, `Diagnose` liefert `mode` als Text, der Fänger führt alle vier Modi, I14):

| Paar | Entscheidung | Ort | Fänger |
|---|---|---|---|
| Kontextfehler × Fehler mit Code | der Abbruch gilt (`errors.Join` und Ursache) | `SPEC-292` Nr. 8, `SPEC-238` Nr. 6 a 1 | I16 (Ausgang `errors.Join(context.Canceled, *Fehler E7003)`) |
| Antwort und Fehler des Dienstes zugleich | der Fehler gilt | `SPEC-292` Nr. 8, `SPEC-238` Nr. 6 b | I16 (Ausgang „Wert und Fehler E7003“) |
| Parameter ungültig × unbekannte Session oder Baseline | der Parameter zuerst: 400, der Dienst wird nicht gerufen | `SPEC-293` Nr. 2, 9 | I7, I8, I9, I16 (der Fake würde `ODR-E7003` und `ODR-E7002` liefern, Zähler 0) |
| `limit` ungültig × `after` ungültig | dieselbe Antwort 400 | `SPEC-293` Nr. 3, 4 | I7 (4 Paare beider Parameter in beiden Reihenfolgen) |
| ein bekannter Schlüssel doppelt oder leer × ein ungültiger Wert desselben Schlüssels | dieselbe Antwort 400 | `SPEC-293` Nr. 3 | I8 (die Formen `k=1&k=` und `k=&k=1` tragen beides) |
| `limit` an `coverage`, `baseline_id` an `exchanges` | unbekannter Schlüssel, ohne Wirkung | `SPEC-293` Nr. 3 | I5 |
| Authentisierung × alles | 401 vor 404 und 405 vor jedem Aufruf einer Route | `SPEC-238` Nr. 3 d | gilt nicht für diesen Plan (Rahmen) |

- Der Use Case liefert für einen ungültigen Seitenparameter und eine leere Baseline-Kennung `model.ErrUngueltig` ohne Code, und `SPEC-109` und §4 der Spezifikation hatten dafür weder Code noch Status; der Code ist `ODR-E2102` mit 400 (entschieden, `SPEC-238` Nr. 15 a; Geber `slice-admin-listener-token`, Teil `slice-admin-rahmen`), die Zahl, die der Adapter nicht lesen kann (Text, Überlauf: ebenfalls `ODR-E2102`; doppelter Parameter: nicht entschieden), entscheidet der Vertrag dieses Slice — **Ausgang:** entschieden (Architect, 2026-11-21): Text, Überlauf, Vorzeichen, Leerraum und der doppelte Parameter sind `ODR-E2102` (400) ohne Aufruf des Dienstes (`SPEC-293` Nr. 3, 4; I6 bis I8); ein neuer Code entsteht nicht.
- Statuscodes der Erfolge und JSON-Form der Antworten standen nicht in `SPEC-005` — **Ausgang:** entschieden (Architect, 2026-11-21): Status, Form und Fehler stehen in `SPEC-293` Nr. 6 bis 9; eine eigene OpenAPI-Datei entsteht nicht (`SPEC-005`).
- Zählen der Fehler (`SPEC-230` Nr. 8): `InspectionUseCase` erhält nie `Telemetry` und zählt `ODR-E7003` und `ODR-E7002` nicht; die Stelle, die den Fehler ausliefert, zählt — das ist der Rahmen von `slice-admin-listener-token` (Sendung dort, §1 *Übernimmt*), nicht jede Route dieses Slice — **Ausgang:** entfallen für diesen Slice (die Routen rufen kein `CountError`, der Rahmen schreibt das Fehlerobjekt); weiter offen bei der Gegenstelle.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (6×): der Fake von `InspectionUseCase` trägt jede Schnittstelle des Originals und zählt jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Marke steht nicht, bevor die Tabelle *Zusage · Fänger* vorliegt — **Ausgang:** offen; Adresse: Planner beim Start (§4): die Tabelle liegt vor, *S* = 2025 liegt 198 unter der Schnittgrenze; der Planner setzt `basis`, `schaetzung` und `marke` vor dem ersten Code-Commit.
- Die Teile teilen die Zeit und das Paket (Schnitt vom 2026-11-21) — **Ausgang:** entschieden (Planner, 2026-11-21): sie laufen nacheinander (§4, *Reihenfolge und Konflikte*), die Sendung der Mittel und der Gesamtzahlen steht in §1 (*Übernimmt*, Geber `slice-admin-sessions`).
- Gesamtzählungen über die drei Teile (I3: 42 Fälle, I17: 84 Zellen): welcher Teil sie zuletzt hebt, hängt von der Reihenfolge von Teil 2 und Teil 3 ab — **Ausgang:** entschieden (Planner, 2026-11-21): der später in `done/` gehende Teil hebt die Literale (§1, Punkt (d)); Annahme des Planners: keine Änderung am Vertrag, nur an der Zählprüfung der Tests; der Architect bestätigt sie beim nächsten Vertrag dieser Linie.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-21; der Planner zieht Treffer und Zähler beim Start nach, der Zähler ist die Zahl der Dateien unter `evidence/`). Der Teil liest Query und Pfad und schreibt JSON gegen Fakes: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (6×) für den Fake des Use Case (beide verkörpert, in `.claude/commands/implement-slice.md` Schritt 14 und `.harness/skills/reviewer.md`), außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (21×, verkörpert). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (32×, `AGENTS.md` §3.9), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (59×, `AGENTS.md` §3.17; „jede Zahl“, „jeder Fehlercode“ als Schleife), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (66×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16). Keine Beobachtung steht mit diesem Slice offen unter 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
