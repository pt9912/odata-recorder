# Slice slice-admin-abweisen-audit: Admin-Rahmen, Teil 6 — Audit des Abweisens, Geheimnisse und Gleichzeitigkeit

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

**Berührte Spec-Stellen:** `LH-FA-13.a` (jedes Abweisen mit `ODR-E2101` hinterlässt einen Audit-Eintrag ohne Tokenwert) · `LH-QA-04.a` · `SPEC-238` Nr. 8 (Audit des Abweisens, bei jeder Antwort 401, auch der eines `*inbound.Fehler` der Route), Nr. 5 und Nr. 6 a 2 (Header folgen dem Status, ein gültiger `*inbound.Fehler` mit 401 ist eine Antwort 401), Nr. 7 (Zählen bleibt unberührt, D5), Nr. 12 (Gleichzeitigkeit), Nr. 13 (Geheimnisse, acht Ausgabewege), Nr. 14 (Binary und Katalog), Nr. 15 (Pflichten der Endpunkte), *Ungeprüft* (Zeile zum Zyklus im Fehlerbaum, nur als Grenze der Suche `finde`, die dieser Teil nicht ändert, §6) · `SPEC-234` Nr. 2 (fünfte Art `admin_abgewiesen`), Nr. 5, Nr. 7 (der Adapter schreibt die vier Arten des Use Case nie) · `SPEC-110` Nr. 2 (unberührt) · `ARC-007`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie. `SPEC-238` steht auf `main` (seit dem Merge von `slice-admin-listener-token`).

**Verantwortlich:** pt9912 (Zweig `slice-admin-abweisen-audit`, Worktree `/Development/odata-recorder-wt-admin-abweisen-audit`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912 (Planner). **Datum:** 2026-11-07.

**Herkunft:** Teil 6 des Schnitts von [slice-admin-listener-token](../done/slice-admin-listener-token.md) (siehe dort und [slice-admin-rahmen](../done/slice-admin-rahmen.md) für die Reihenfolge der sechs Teile). Der Plan trägt die Zeilen D1 bis D7, H1, N1, Y1, Y5, Y6 und P1 der Tabelle des Architect (13 von 102); sie stehen im Wortlaut in §3. Gegenüber dem Vorschlag des Architect trägt dieser Teil zusätzlich H1 (die Schleife über alle acht Ausgabewege braucht den Eintrag und die Fehlertexte von `Serve` und `Schliessen`) und N1 (gleichzeitige Requests über `Handler`, `Append` und `CountError`).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Jedes Abweisen mit `ODR-E2101` hinterlässt genau einen Audit-Eintrag ohne Tokenwert (`SPEC-238` Nr. 8): die fünfte Art `admin_abgewiesen` in `support.Arten()` (Pflicht allein `error_code`), vom Adapter des Audit angenommen und als Zeile `{"timestamp":"…","correlation_id":"cor-…","art":"admin_abgewiesen","error_code":"ODR-E2101"}` geschrieben, und im Rahmen ein `Append(Eintrag{Art: ArtAdminAbgewiesen, ErrorCode: CodeE2101, KorrelationsID: <ID der Antwort>})` auf der Goroutine des Requests, vor dem Schreiben von Header und Body, ohne Drosselung und Zusammenfassung; ein Fehler von `Append` ändert weder Antwort noch Zählen; bei keiner anderen Antwort ein `Append`. Dazu der Abschluss der Zusagen über den ganzen Rahmen: kein Tokenwert und kein Teilstück auf einem der acht Ausgabewege (Schleife H1) und gleichzeitige Requests teilen nichts (N1). Der Test in `src/bootstrap/audit_betrieb_test.go` führt fünf Arten und nimmt `admin_abgewiesen` ausdrücklich aus (Y6); das Binary bleibt unverändert.

**Übernimmt (Sendungen von `slice-admin-listener-token`, Schnitt vom 2026-11-07, `AGENTS.md` §3.13):** (1) die Änderung der Testdatei `src/bootstrap/audit_betrieb_test.go` — Rückgabe R2 des Architect, vom Planner bestätigt (nur Testdatei, kein Produktcode von `bootstrap`; die Voraussetzung „Binary unverändert“ gilt); (2) den Audit-Eintrag zu `ODR-E2101` aus `LH-FA-13.a`, dessen Fänger `slice-admin-authentisierung` (U1) bis hierher offen lässt; (3) den Eintrag auch bei einer Antwort 401, die ein gültiger `*inbound.Fehler` der Route mit `ODR-E2101` auslöst (`SPEC-238` Nr. 8 „bei **jeder** Antwort 401“, Entscheidung des Architect vom 2026-11-20) — Sendung von `slice-admin-fehlerabbildung` (Geber), Fänger D1; `slice-admin-fehlerabbildung` ruft `Audit` nicht.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die vier Arten einer Löschung und eines fehlgeschlagenen Imports (Session gelöscht, Session durch Aufbewahrung gelöscht, Workspace gelöscht, Import fehlgeschlagen; `ArtSessionGeloescht`, `ArtSessionAufbewahrung`, `ArtWorkspaceGeloescht`, `ArtImportFehlgeschlagen`) — allein der Use Case schreibt sie (`slice-audit-dienste`, Entscheidung `SPEC-234` Nr. 7); der Adapter schreibt keine der vier (D2 fängt es).
- Der Port `AuditLog` und der Adapter des Audit — `slice-audit-log` (`done/`); dieser Teil ruft den Port und ergänzt dort nur die fünfte Art (`audit.go` des Ports, Tests des Adapters).
- Eine Drosselung der Einträge und die Begrenzung der Fehlversuche — das Lastenheft nennt keine (`LH-FA-13.a`); benannte Lücke, Adresse User (Rückgabe R5, §6), kein Code, bis entschieden.
- Der Fänger gegen einen Tokenwert in `<DATA_DIR>/audit.jsonl` des gebauten Prozesses — `slice-bootstrap-m6` (Sendung von `slice-audit-verdrahtung`); dieser Teil fängt am Fake von `AuditLog` und setzt mit der Feldmenge von `SPEC-238` Nr. 8 darauf auf.
- Das Einsortieren der Zelle `admin_abgewiesen` in den Test von Modus × Art und die Verdrahtung des Admin-Adapters — `slice-bootstrap-m6`; hier nimmt der Test die Art aus, weil kein Use Case sie schreibt und kein Aufruf im Binary besteht.
- Die vorigen Teile des Rahmens (`slice-admin-rahmen`, `-authentisierung`, `-lebenszyklus`, `-fehlerabbildung`) und die Token-Datei (`slice-admin-listener-token`) — Voraussetzungen; die Endpunkte, TLS und das Lesen von `ADMIN_*` — `slice-admin-sessions`, `slice-admin-imports`, `slice-admin-workspaces`, `slice-admin-tls`, `slice-bootstrap-m6`.

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

- [ ] Audit des Abweisens: die Art `admin_abgewiesen` als fünfte Zeile von `support.Arten()` (Pflicht allein `error_code`) und im Adapter des Audit (jedes verbotene Feld und das fehlende `error_code` abgewiesen, Zeile Byte für Byte), genau ein `Append` je 401 (auch bei einem gültigen `*inbound.Fehler` der Route mit `ODR-E2101`, direkt und verpackt) mit den Feldern `Art`, `ErrorCode`, `KorrelationsID` und sonst leer, kein `Append` bei einer anderen Antwort und nie eine der vier Arten des Use Case, vor dem Schreiben von Header und Body, auf der Goroutine des Requests ohne Frist, Fehler von `Append` ändert nichts, jede Abweisung ein Eintrag (D1 bis D7, Y5).
- [ ] Geheimnisse und Gleichzeitigkeit über den ganzen Rahmen: kein Tokenwert und kein Teilstück auf einem der acht Ausgabewege von `SPEC-238` Nr. 13 (H1, Zählprüfung 8 gegen die Buchstaben des Vertrags), gleichzeitige Requests teilen nichts und keine Sperre liegt über `Handler`, `Append`, `CountError` oder `Vergleich` (N1).
- [ ] Binary unverändert: `src/bootstrap` ohne Produktionsänderung, der Test `audit_betrieb_test.go` zählt fünf Arten und nimmt `admin_abgewiesen` aus (Y1, Y6); die Pflichten der Endpunkte stehen in den Plänen der Nehmer (P1, akzeptiertes Negativ).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Für diesen Slice: kein neuer Code; das Binary bleibt unverändert (kein Aufruf von `admin` aus `bootstrap`, in `src/bootstrap` ändert sich allein die Testdatei aus Y6), in §7 vermerkt. Der Audit-Eintrag erscheint im Handbuch mit `slice-bootstrap-m6`, sobald der Listener verdrahtet ist.
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
| `src/adapters/driving/admin/` | update | `Audit.Append` bei jeder Antwort 401 (auch der eines Fehlers der Route), vor dem Schreiben, auf der Goroutine des Requests (`SPEC-238` Nr. 8); der Ort ist die eine Schreibstelle `schreibeFehler` (`admin.go`), in der Zählen und Header dem Status folgen; sie erhält die Korrelations-ID der Antwort als Parameter `id` (jeder Aufrufer reicht sie weiter: `bediene` und `rufe` in `admin.go`, `schreibeFehlerDerRoute` in `fehler.go`, `abweise` in `lebenszyklus.go`); die Suche im Fehlerbaum (`finde`, `findeAuf` in `fehler.go`) bleibt unberührt (§6, M-1) |
| `src/hexagon/ports/support/audit.go`, `src/hexagon/ports/support/audit_test.go` | update | Art `ArtAdminAbgewiesen` als fünfte Zeile von `Arten()`; Orakel des Tests auf fünf Arten |
| `src/adapters/driven/audit/audit_test.go` | update (nur Test) | Schleife über Felder × Arten mit der fünften Art und die Zeile des Vertrags |
| `src/bootstrap/audit_betrieb_test.go` | update (nur Test, Rückgabe R2) | zählt vier Arten und die Zellen Modus × Art: fünf Arten, `admin_abgewiesen` ausdrücklich ausgenommen (kein Use Case, kein Aufruf im Binary); kein Produktcode von `bootstrap` ändert sich |
| Tests | neu / update | Adapter gegen `ServeHTTP` und gegen einen echten Listener; `FakeAudit` (hält jedes Argument, zählt `Aufrufe()`, `Vor` als Haltepunkt, `Fehler` einstellbar) und `FakeTelemetry` aus `supporttest`; Pipes für `stdout`, `stderr` und `log.Default()`; der Adapter importiert weder Adapter noch Service (`make a-check`). **Bestehende Prüfungen, die bei einer Antwort 401 `Audit` 0 verlangen, schreibt dieser Teil um** (Vorbehalt von `slice-admin-rahmen` N-2: die Tests „Audit 0“ schreibt ein späterer Teil um): `nurZaehlung` in `helfer_test.go` (Audit 0 an jeder Stelle), `TestZaehlenJedeSchreibstelle` in `schreibstellen_test.go` (Stelle 401) und `pruefeFehlerDerRoute` in `fehlerabbildung_test.go` (das Paar `ODR-E2101` mit 401 von E1); weitere findet der Lauf von `make test`, jede steht im Bericht des Implementers. Umgesetzt: neue Dateien `abweisen_audit_test.go` (D1 bis D7 ohne N1), `abweisen_gleichzeitig_test.go` (N1) und `abweisen_geheimnisse_test.go` (H1, `//go:build linux` wie die Erfassung der Ausgabe); umgeschrieben: `nurZaehlung` (`helfer_test.go`: Audit gleich der Zahl der Zählungen von `ODR-E2101`, bei jedem anderen Code 0, damit auch `pruefeFehlerDerRoute` mit dem Paar `ODR-E2101` und 401 und `TestMethodeUndPfad`) und `TestZaehlenJedeSchreibstelle` (`schreibstellen_test.go`: Stelle 401 ein Aufruf, jede andere keiner); der Lauf von `make test` fand außerdem allein die Tests der Zeilen Y5 und Y6 rot (`support/audit_test.go`, `bootstrap/audit_betrieb_test.go`) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stammen vom Architect (2026-11-07, Plan von [slice-admin-listener-token](../done/slice-admin-listener-token.md)); der Schnitt gibt jedem Teil ganze Zeilen im Wortlaut (D Audit, H Geheimnisse, N Gleichzeitigkeit, Y Katalog und Code, P Pflichten der Endpunkte). Die Mengen *R* und *W* und die Mengen M-Header (U1) und M-Methode × Pfad (U7), über die D1 und D2 laufen, führt der Plan, der die jeweilige Zeile trägt (`slice-admin-authentisierung`, `slice-admin-fehlerabbildung`).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| D1 | (Nr. 8) Bei jeder Antwort 401 genau ein `Append(Eintrag{Art: ArtAdminAbgewiesen, ErrorCode: CodeE2101, KorrelationsID: ID der Antwort})`, alle übrigen Felder leer | Schleife über die 20 Formen von U1, die Zellen von U7 und `Vergleich` gleich `false`, dazu ein Fehler der Route `Fehler{ODR-E2101, 401}` (direkt und verpackt): `FakeAudit.Eintraege()` gleich dem Literal-Eintrag mit der ID, `Aufrufe()` 1 | `ServiceID` gesetzt; Art falsch; `Append` nur im Zweig der Authentisierung, nicht in der einen Schreibstelle | andere |
| D2 | (Nr. 8) Bei keiner anderen Antwort ein `Append`, und nie eine der vier Arten von `SPEC-234` | Schleife über 404, 405, 400, 503, 500, healthz, Erfolg, Kontextfehler: `FakeAudit.Aufrufe()` 0 | `Append` auch bei 404 | andere |
| D3 | (Nr. 8) Der Eintrag trägt keinen Tokenwert, kein Teilstück, keinen Pfad, keine Methode, keine Adresse | Canary im falschen Token, im Pfad, in der Query, in einem Header, in der Methode: der gehaltene `Eintrag` ist als Struktur gleich dem Literal (alle sieben Felder) | der Pfad in `ServiceID` | andere |
| D4 | (Nr. 8) Bis `Append` zurückkehrt, hat der Adapter weder `WriteHeader` noch `Write` gerufen und keinen Header des Fehlerobjekts gesetzt (`X-Correlation-Id` steht vorher in der Map, Nr. 3 a) | `FakeAudit.Vor` prüft einen `ResponseRecorder` (`ServeHTTP` direkt): kein `WriteHeader`, kein `Write`, Body leer, in der Header-Map allein `X-Correlation-Id` | `Append` nach dem Schreiben | andere |
| D5 | (Nr. 8) Das Ergebnis von `Append` ändert weder Antwort noch Zählen | `FakeAudit.Fehler` gesetzt (`ErrUngueltig`, `errors.New`): Antwort Byte für Byte gleich der ohne Fehler, `CountError` 1 | 500 bei einem Fehler des Audits | andere |
| D6 | (Nr. 8) `Append` läuft auf der Goroutine des Requests, ohne Frist: `ServeHTTP` kehrt erst nach `Append` zurück | `Vor` blockiert auf einem Kanal; `ServeHTTP` in einer Goroutine: sie ist nach dem Freigeben zu Ende und nicht davor (Kanäle, kein Warten auf Zeit) | `go h.Audit.Append(...)` | nebenläufig |
| D7 | (Nr. 8) Jede Abweisung schreibt einen Eintrag, ohne Drosselung und Zusammenfassung | 8 Goroutinen × 8 Requests unter dem Detektor: 64 Einträge, die `KorrelationsID` paarweise verschieden | Drosselung nach N Einträgen | nebenläufig |
| H1 | (Nr. 13) Das Token und kein Teilstück erreicht einen der acht Ausgabewege (a Body, b Header, c Eintrag, d Argumente von `Telemetry`, e Fehlertexte von `New`, `Serve`, `Schliessen`, f `log.Default()`, g `stdout` und `stderr`, h Panik und `ErrorLog`) | Schleife Weg × Anlass (richtiges Token; falsches mit Canary; Canary in Pfad, Query, Body, fremdem Header; Token in der Ursache eines Fehlers der Route; im Panikwert; `Schliessen` mit Frist); Teilstücke die ersten, mittleren und letzten 8 Bytes; Canary aus Bytewerten; Zählprüfung 8 Wege gegen Nr. 13 | `fmt.Println(token)` im Fehlerpfad | andere |
| N1 | (Nr. 12) Gleichzeitige Requests teilen nichts, und keine Sperre liegt über `Handler`, `Append`, `CountError` oder `Vergleich` | 32 Requests parallel (401, 200, 404, Fehler, Panik) in 20 Läufen unter dem Detektor: eigene ID je Antwort, Einträge und Zähler gleich der Erwartung; ein blockierter `Handler` hält die übrigen nicht auf | Sperre um den Aufruf des `Handler` | nebenläufig |
| Y1 | (Nr. 14) Das Binary bleibt unverändert: kein Aufruf von `admin` aus `bootstrap`, kein `ADMIN_*` im Laden, `SPEC-110` Nr. 2 lehnt `ADMIN_*` weiter ab | benannte Ausgabe: `git diff <Basis> -- src/bootstrap` ohne Produktionsdatei (nur die Testdatei aus Y6) und `make test` mit den Tests von `SPEC-110` Nr. 2 grün | ein Import von `admin` in `src/bootstrap` | andere |
| Y5 | (Nr. 8) Die Art `admin_abgewiesen` ist die fünfte Zeile von `support.Arten()` (Pflicht allein `error_code`); der Adapter des Audit nimmt den Eintrag an, schreibt die Zeile des Vertrags und weist jedes verbotene Feld und das fehlende `error_code` ab | `support/audit_test.go` (Orakel mit 5 Arten); `driven/audit/audit_test.go`: Schleife über 5 Felder (4 verboten, `error_code` Pflicht) × Art; Zeile Byte für Byte | Pflichtfeld `service_id` statt `error_code` | andere |
| Y6 | (Nr. 8) Der Test in `src/bootstrap/audit_betrieb_test.go` zählt die Arten und die Zellen Modus × Art: er führt fünf Arten und nimmt `admin_abgewiesen` ausdrücklich aus (kein Use Case, kein Aufruf im Binary) | `go test ./src/bootstrap`: grün mit der Ausnahme; die Ausnahme entfernt ergibt rot (Zelle ohne Fall) | die Ausnahme zu breit (nimmt `session_geloescht` mit) | andere |
| P1 | (Nr. 15) Die Pflichten (a) bis (d) der Endpunkte stehen in den Plänen von `slice-admin-sessions`, `slice-admin-imports`, `slice-admin-workspaces` | **Akzeptiertes Negativ.** Grund: der Fänger liegt in den Slices der Endpunkte (ihre Routen, ihre Tests), dieser Slice hat keine Route. Folge: bis dahin hält die Pflicht allein der Plan der Nehmer (Rückgabe R3 in §6, Planner); der Reviewer jener Slices löst jede Zeile von Nr. 15 mit einer Mutation auf | – | andere |

Zeilen gesamt 13: *a* = 3 (D6, D7, N1), *b* = 10. N1 steht hier statt bei `slice-admin-authentisierung`, weil die Schleife `Append` und `CountError` einschließt; die Zusage „keine Sperre über `Vergleich`“ ist damit erst mit diesem Teil gefangen (Risiko in `slice-admin-authentisierung`, §6).

**Ausgabewege (H1)** — die acht Wege von `SPEC-238` Nr. 13 und der Teil, der den Anlass liefert; die Schleife H1 läuft über alle acht, die Zählprüfung gegen die Buchstaben des Vertrags gilt hier:

| Weg | Gegenstand | liefert der Teil | lokaler Fänger im Teil |
|---|---|---|---|
| a Body | Fehlerobjekt und Erfolgsantwort | `slice-admin-authentisierung`, `slice-admin-fehlerabbildung` | U9, E5 |
| b Header | Antwort-Header | `slice-admin-authentisierung` | E2, E3 |
| c Eintrag | Argument von `Audit.Append` | dieser Teil | D1, D3 |
| d Argumente von `Telemetry` | `CountError` | `slice-admin-fehlerabbildung` | C1 |
| e Fehlertexte | `New`, `LadeToken`, `Serve`, `Schliessen` | `slice-admin-rahmen` (K1, K2), `slice-admin-listener-token` (T24), `slice-admin-lebenszyklus` | K2, T24, S2, S7 |
| f `log.Default()` | Log des Prozesses | `slice-admin-fehlerabbildung` | H2 |
| g `stdout` und `stderr` | Ausgabe des Prozesses | `slice-admin-fehlerabbildung`, `slice-admin-listener-token` | H2, T25 |
| h Panik und `ErrorLog` | Wert der Panik, Log des `http.Server` | `slice-admin-fehlerabbildung`, `slice-admin-lebenszyklus` | B8, G1 |

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, Auftrag C):

| Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|---|
| M-Ausgabewege (H1) | 8 Wege × 7 Anlässe × 3 Teilstücke und der ganze Wert | Schleife | Canary fehlt | 8 gleich den Buchstaben in `SPEC-238` Nr. 13 |
| M-Header (U1), M-Methode × Pfad (U7), M-Schreibstellen W | aus den Plänen, die ihre Zeilen tragen (20, 40, 9) | Schleife | Eintrag gleich dem Literal | wie dort |

**Unbeschränkte Menge und ihre Prüfgrenze (akzeptiertes Negativ):** die Zahl gleichzeitiger Requests (Prüfgrenze 32 in N1, 8 × 8 in D7; Grund: der Zustand ist je Request lokal, es gibt keine gemeinsame Sperre; Folge: ein Fehler unter 1000 gleichzeitigen Requests ist durch die Eigenschaft „keine gemeinsame Sperre“ getragen, nicht durch einen eigenen Fänger).

**Zweige als Menge, Runden unter dem Detektor** (Architect, Auftrag D; Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat null Blöcke mit Zähler 0, der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Audit (401, Fehler von `Append`) | D1 bis D7 | – |

**Aufrufstellen und zählende Fakes** (Architect, Auftrag E; „jeder Fall je Aufrufstelle“ gilt für diese Stellen):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| genau eine, beim 401 | `Audit.Append` | D1 bis D7 | `FakeAudit`: `Eintraege()` hält jedes Argument, `Aufrufe()` zählt, `Vor` ist der Haltepunkt; zählt auch die Aufrufe, die nie vorkommen dürfen (404, 405, 400, 503, 500, Erfolg) |
| genau eine, in der Schreibstelle | `Telemetrie.CountError` | D5 (`CountError` 1 bei Fehler des Audits) | `FakeTelemetry` |

**Token-Handling und Canary** (Architect, Auftrag G): Was der Adapter an `AuditLog` übergibt: `Eintrag{Art: ArtAdminAbgewiesen, ErrorCode: CodeE2101, KorrelationsID: <ID>}` und sonst nichts (D1, D3); die Zeile im Audit-Log ist `{"timestamp":"…","correlation_id":"cor-…","art":"admin_abgewiesen","error_code":"ODR-E2101"}`. Der Fänger gegen einen Tokenwert in `<DATA_DIR>/audit.jsonl` des gebauten Prozesses liegt bei `slice-bootstrap-m6` und setzt auf dieser Feldmenge auf.

**Größe der Sicht des Architect** (`SPEC-128` Nr. 1, Aufteilung durch den Planner): 13 Zeilen (*a* = 3, *b* = 10), *f* = 0, *p* = 40.

**Raten der Fänger** (`AGENTS.md` §3.10): D6, D7 und N1 sind Zusagen über Nebenläufigkeit; jede Mutation muss in jedem Lauf rot werden. Der Implementer trägt die Tabelle *Zusage · Mutation · roter Test* und die Raten je Mutation (Zahl der Prozesse, Messender, Last) in einen Absatz von §3 ein, bevor er an den Reviewer übergibt.

**Umsetzung, Mutationen, Mengen und Raten** (Implementer, 2026-11-21; `AGENTS.md` §3.10, §3.15, §3.17). Messendes: `sh tools/harness/mutationslauf.sh` mit zwei Listen (62 und 4 Mutationen), `MUTATIONSLAUF_SPEICHER=4g`, je Aufruf unter `flock /tmp/odata-recorder-heavy.lock`, Image `odata-recorder:race`. Ergebnis: Liste 1 (62): 60 `rot`, 2 `flatternd`, 0 `rot-paket`, 0 `gruen`, 0 `uebersetzungsfehler`, Gesamtdauer 486,1 s, je Mutation 0,3 bis 83,0 s (83 s die vier `N1-sperre-um-…`, die in 60 Prozessen je auf die Frist von 5 s warten); Liste 2 (4: `D3-body`, `D3-ursache`, `D1-kontext-beendet-ohne-eintrag`, `…-route`): 4 `rot`, 36,8 s. Der Code war bei beiden Läufen der des Handoff-Commits; nach Liste 1 kamen allein die Tests `TestAuditUeberEchtenListener` und `TestAuditBeiBeendetemKontext` hinzu (die zweite Liste und zwei Einzelläufe fahren sie). Die vier Mutationen von Y6 liefen im Einzellauf des benannten Tests (`go test -run` in einem Container ohne Netz, Datei ersetzt, dann zurückgesetzt), nicht im Werkzeug: dessen Nullprobe fährt `src/bootstrap` ganz und bricht bei der Zeitgrenze von 2 Minuten je Prozess ab (`Ausgangsstand rot`, die Suite braucht rund 250 s, `SPEC-210` Nr. 5 e); 4 von 4 rot. Last des Hosts am Beginn jeder Mutation (erster Wert von `/proc/loadavg`, `last=l/k`): 1,30 bis 3,68 bei `k` = 4 CPUs des Containers (Liste 2: 3,45 bis 3,49); andere Rollen liefen auf dem Host, ein Rot unter dieser Last ist kein Rot auf ruhigem Host. Die Gründe des Rot wurden an 42 Mutationen im Einzellauf gelesen (Name, Test, erste Meldungen; je Zeile D1 bis D7, H1, Y5 und Y6 mehrere): jede ist aus dem benannten Grund rot, keine weil der Aufbau bricht; eine Mutation (`D5-kontextfehler-bricht-ab`) beendete den Prozess durch eine Panik außerhalb von `ServeHTTP`, die Prüfung fängt sie seitdem mit `rufeSicher` als Fehler des Tests ab.

**Raten der Zusagen über Nebenläufigkeit** (D6, D7, N1; `AGENTS.md` §3.10; Prozesse: je Prozess `-test.count=1`; Messender: `mutationslauf`, Arm mit Detektor = Binary mit `-race`, Arm ohne = ohne; Last wie oben):

| Mutation | Zusage | Prozesse | mit Detektor | ohne Detektor | Last (l/k) |
|---|---|---|---|---|---|
| `D6-goroutine` | D6 | 100 | 50/50 | 50/50 | 3,56/4 |
| `D6-goroutine-mit-warten` | D6 | 100 | 50/50 | 50/50 | 3,56/4 |
| `D7-drosselung-nach-32` | D7 | 100 | 50/50 | 50/50 | 3,68/4 |
| `D7-zusammenfassung-jeder-zweite` | D7 | 100 | 50/50 | 50/50 | 3,54/4 |
| `D7-zaehler-ohne-sperre` | D7 | 100 | 50/50 | 0/50 | 3,54/4 |
| `N1-sperre-um-handler` | N1 | 60 | 30/30 | 30/30 | 3,66/4 |
| `N1-sperre-um-append` | N1 | 60 | 30/30 | 30/30 | 2,04/4 |
| `N1-sperre-um-counterror` | N1 | 60 | 30/30 | 30/30 | 1,36/4 |
| `N1-sperre-um-vergleich` | N1 | 60 | 30/30 | 30/30 | 1,36/4 |
| `N1-geteilter-zaehler-ohne-sperre` | N1 | 100 | 50/50 | 0/50 | 1,30/4 |

Gegenprobe am unveränderten Stand: die Tests dieses Slice (Muster `Audit|Abweisung|Gleichzeitige|Blockierter|KeinToken|H1Probe|Quellen|Eintrag|Kontext|Zaehlen`: die Tests dieses Slice und einige des Rahmens mit gleichem Namensteil) in 40 Prozessen mit Detektor und 40 ohne, je `-test.count=1`: 0 von 40 rot und 0 von 40 rot (Container mit 4 CPUs, Last des Hosts nicht notiert). Jede Mutation mit einer Wirkung (Goroutine, Drosselung, Zusammenfassung, Sperre) ist in jedem Prozess rot, mit und ohne Detektor: D6 fängt der Stapel des Aufrufers in `Vor` (`ServeHTTP` steht darauf, deterministisch; die zweite Probe, `ServeHTTP` ist vor der Freigabe nicht zu Ende, ist ein Kanal ohne Warten und würde bei `go Append` nur mit Wahrscheinlichkeit rot: sie fängt `D6-goroutine` im Einzellauf und `D6-goroutine-mit-warten` nie, dort trägt allein der Stapel; die Rate der zweiten Probe allein ist nicht gemessen), D7 zählt die Einträge, N1 hält einen Aufruf an und verlangt, dass die 32 übrigen zurückkehren. Die zwei Mutationen ohne Wirkung (`D7-zaehler-ohne-sperre`, `N1-geteilter-zaehler-ohne-sperre`: ein gemeinsamer Zähler ohne Sperre) sind Datenrennen: der Detektor meldet sie in 50 von 50 Prozessen, ohne Detektor sieht sie kein Test (0 von 50, `flatternd` ist hier das erwartete Ergebnis, `SPEC-210` Nr. 7 c).

| Zeile | Zusage · Mutation (Name in der Liste) | roter Test |
|---|---|---|
| D1 | genau ein `Append` mit Art, Code und ID der Antwort je 401, sonst leer, auch beim Fehler der Route direkt und verpackt · Art `session_geloescht` (`D1-art-falsch`), `ServiceID` gesetzt (`D1-serviceid-gesetzt`), Code `ODR-E2102` (`D1-code-falsch`), ID leer (`D1-id-leer`), zweimal (`D1-zweimal`), nur im Zweig der Authentisierung (`D1-nur-authentisierung`), eine zweite Aufrufstelle (`D1-zweite-aufrufstelle`), keine Antwort und kein Eintrag bei beendetem Kontext (`D1-kontext-beendet-ohne-eintrag`, `D1-kontext-beendet-ohne-eintrag-route`) | `TestAuditBeiJederAntwort401`, `TestAppendEineAufrufstelle`, `TestAuditBeiBeendetemKontext`; dazu in den umgeschriebenen Prüfungen `TestFehlerobjektFormat` (Paar `ODR-E2101`, 401: `D1-nur-authentisierung-alt`) und `TestAuditUeberEchtenListener` (`D1-id-leer`, `D5-500-bei-fehler` am echten Listener) |
| D2 | bei keiner anderen Antwort ein `Append`, nie eine der vier Arten · `Append` auch bei 404 (`D2-bei-404`), 405, 503, 400, 500 (`D2-bei-405`, `D2-bei-503`, `D2-bei-400`, `D2-bei-500`), beim Fehler der Route mit 404 (`D2-bei-fehler-der-route`), bei `GET /healthz` (`D2-bei-healthz`), vor dem Abbruch bei einem Kontextfehler (`D2-vor-kontextfehler`), Art `session_geloescht` (`D2-vier-arten-des-use-case`) | `TestKeinAuditBeiAndererAntwort`; in den umgeschriebenen Prüfungen `D2-bei-404-alt` in `TestMethodeUndPfad`, `D2-bei-503-alt` in `TestZaehlenJedeSchreibstelle`, `D2-bei-500-alt` in `TestPanikDerRoute` (jede rot: die Umschreibung schwächt D2 nicht) |
| D3 | kein Tokenwert, Teilstück, Pfad, Methode, Adresse im Eintrag · Pfad, Query, Methode, Adresse, Header, Kandidat, Body, Ursache der Route an die ID gehängt (`D3-pfad`, `D3-query`, `D3-methode`, `D3-adresse`, `D3-header`, `D3-kandidat`, `D3-body`, `D3-ursache`) | `TestEintragOhneWerteDesRequests` (der Fund nennt den Ort) |
| D4 | bis `Append` zurückkehrt, weder `WriteHeader` noch `Write` gerufen und kein Header des Fehlerobjekts gesetzt, allein `X-Correlation-Id` in der Map · `Append` nach den Headern (`D4-nach-header`), nach `WriteHeader` (`D4-nach-writeheader`), nach `Write` (`D4-nach-write`) | `TestAuditVorDemSchreiben` |
| D5 | das Ergebnis von `Append` ändert weder Antwort noch Zählen · 500 bei einem Fehler (`D5-500-bei-fehler`), Rückkehr ohne Antwort (`D5-abbruch-bei-fehler`), Abbruch bei `context.Canceled` aus `Append` (`D5-kontextfehler-bricht-ab`), Zählen zweimal bei einem Fehler (`D5-zaehlt-zweimal-bei-fehler`) | `TestAuditFehlerAendertNichts` |
| D6 | `Append` auf der Goroutine des Requests, `ServeHTTP` kehrt erst danach zurück · `go Append` (`D6-goroutine`), Goroutine mit Warten auf ihr Ende (`D6-goroutine-mit-warten`) | `TestAuditAufDerGoroutineDesRequests` |
| D7 | jede Abweisung ein Eintrag · Drosselung nach 32 (`D7-drosselung-nach-32`), nur jeder zweite Eintrag (`D7-zusammenfassung-jeder-zweite`), gemeinsamer Zähler ohne Sperre (`D7-zaehler-ohne-sperre`) | `TestJedeAbweisungEinEintrag` |
| H1 | kein Tokenwert und kein Teilstück auf einem der acht Wege · je Weg eine Mutation: Body (`H1-a-token-im-body`), Header (`H1-b-token-im-header`), Eintrag (`H1-c-token-im-eintrag`), `Telemetry.Log` (`H1-d-token-an-telemetrie`), Fehlertext von `New`, `Schliessen`, `Serve` (`H1-e-token-im-text-von-new`, `…-schliessen`, `…-serve`), `ErrorLog` in `log.Default()` und `log.Printf` (`H1-f-errorlog-in-log-default`, `H1-f-log-printf-token`), `println` auf stderr (`H1-g-stderr-token`), Panik mit dem Token (`H1-h-panik-mit-token`) | `TestKeinTokenAufAchtWegen` (der Fund nennt Anlass und Weg), Probe des Fängers `TestH1ProbeSiehtJedenWeg` |
| N1 | keine Sperre über `Handler`, `Append`, `CountError`, `Vergleich`, gleichzeitige Requests teilen nichts · `s.mu` über den Aufruf (`N1-sperre-um-handler`, `N1-sperre-um-append`, `N1-sperre-um-counterror`, `N1-sperre-um-vergleich`), ein gemeinsamer Zähler ohne Sperre (`N1-geteilter-zaehler-ohne-sperre`) | `TestBlockierterAufrufHaeltDieUebrigenNicht` (vier Stellen), `TestGleichzeitigeRequestsTeilenNichts` |
| Y1 | das Binary bleibt unverändert | benannte Ausgabe: `git diff 58a5b407..HEAD -- src/bootstrap` nennt allein `audit_betrieb_test.go`; kein Fänger außer dieser Ausgabe (akzeptiertes Negativ unten) |
| Y5 | fünfte Zeile von `Arten()` mit `error_code` als Pflicht, Adapter nimmt an, schreibt die Zeile, weist ab · Pflicht `service_id` statt `error_code` (`Y5-pflicht-service-id`, auch gegen den Adapter: `Y5-adapter-pflicht-service-id`), `error_code` und `session_id` Pflicht (`Y5-pflicht-error-code-und-session-id`), fünfte Zeile fehlt (`Y5-fuenfte-zeile-fehlt`), Art `admin_abgelehnt` (`Y5-art-text`), `error_code` keine Pflicht (`Y5-adapter-error-code-nicht-pflicht`), Zeile ohne `error_code` (`Y5-adapter-zeile-ohne-error-code`), verbotene Felder angenommen (`Y5-adapter-verbotene-felder-erlaubt`) | `TestArtenGleichTabelle`, `TestArtTexte` (`support/audit_test.go`), `TestFeldBelegung`, `TestZeileAdminAbgewiesen` (`driven/audit/audit_test.go`) |
| Y6 | der Test in `audit_betrieb_test.go` führt fünf Arten und nimmt `admin_abgewiesen` aus · Ausnahme zu breit, sie nimmt `session_geloescht` mit (`Y6-ausnahme-zu-breit`), Ausnahme entfernt (`Y6-ausnahme-entfernt`), Ausnahme nur im Modus `record` (`Y6-ausnahme-nur-in-record`), Zählung 4 statt 5 (`Y6-zaehlung-vier`) | `TestAuditAufrufstellen` |
| P1 | akzeptiertes Negativ (Zeile P1 oben) | – |

**Mengen** (Generator · Orakel · Zahl der Elemente, die der Test zählt; Zählprüfung gegen die Quelle). *M-401-Formen* (D1, D3 bis D6): Schleife über die 20 Formen von U1, die 40 Zellen von U7 (8 Methoden × 5 Pfade) ohne Token und mit falschem Token und den Fehler der Route `ODR-E2101` mit 401 in 5 Verpackungen ohne und mit Ursache: 20 + 40 + 40 + 10 = 110 (`TestQuellen401Zahl`; die 8 und 5 und 20 kommen aus `u7Methoden`, `u7Pfade`, `u1Formen`, `verpackungen`); Orakel: der Eintrag als Literal (Art und Code aus dem Wortlaut des Vertrags, die ID aus dem Header der Antwort, `TestEintragHatSiebenFelder` hält die sieben Felder). *D2:* die 30 Requestarten *R* (genau die 21 mit Status 401 schreiben einen Eintrag), M-Code × Status (41 Codes der Tabelle und 3 ungültige × 19 Status, davon 1 Paar mit Eintrag: 44 × 19 = 836), 28 Kontextfehler (2 Fehler × 7 Formen × 2 Stellen), 20 Formen von U1 bei angehaltenem Server. *D3:* 8 Orte (Pfad, Query, Header, Body, Kandidat, Methode, Adresse, Ursache) einzeln und zugleich = 9 Fälle, 11 Werte je Fall geprüft (die 8 Orte der Orte-Hilfe und 3 weitere). *D4:* 110 Formen × 3 Stellen der Mutation. *D5:* 110 Formen × 7 Fehler von `Append` = 770 (`ErrUngueltig`, zwei Fehler der Ablage, `errors.New` und `io.ErrShortWrite`, zwei Kontextfehler, ein `*inbound.Fehler`, ein verpackter `ErrUngueltig`). *D7:* 64 nacheinander, 8 × 8 zugleich. *N1:* 20 Läufe × 32 Requests (5 Arten), 4 Stellen × 32. *H1:* 8 Wege × 7 Anlässe = 56 Zellen (das Orakel prüft jedes Fenster von 8 Bytes jedes Wertes: der Wert, sein Anfang, seine Mitte, sein Ende und jedes Teilstück dazwischen). *Y5:* 5 Arten × 5 Felder = 25 Zellen (Test des Ports und des Adapters) und die 66 Werte von `-5` bis `60` für `error_code`. Je Menge eine Mutation an einem Element, das die Zeile des Plans nicht nannte, rot gesehen: D1 `D1-code-falsch` und `D1-zweimal`; D2 `D2-bei-503` und `D2-bei-healthz`; D3 `D3-methode`, `D3-adresse` und `D3-ursache`; D5 `D5-kontextfehler-bricht-ab` (`context.Canceled` aus `Append`); D6 `D6-goroutine-mit-warten`; D7 `D7-zusammenfassung-jeder-zweite`; N1 `N1-sperre-um-counterror`; H1 die Wege a bis f und h (der Plan nannte allein `fmt.Println` im Fehlerpfad, Weg g); Y5 `Y5-art-text`, `Y5-adapter-verbotene-felder-erlaubt`.

**Unter `-race -coverprofile`** (Image `odata-recorder:race`, `go test -race -count=1 -coverprofile` über das Paket `admin`, 99,2 % der Anweisungen): drei Blöcke mit Zähler 0, alle außerhalb dieses Slice: `route.go` Funktion `zerlegePfad`, `tokendatei.go` Funktionen `pruefeUndLies` und `erzeuge`; der Code dieses Slice (`admin.go` `schreibeFehler`, `bediene`, `rufe`, `fehler.go`, `lebenszyklus.go` `abweise`) hat keinen Block mit Zähler 0.

**Akzeptierte Negative** (Grund · Folge · ungefangene Klasse mit Maß; `AGENTS.md` §3.15):
- Y1: Binary unverändert. Grund: ein Import von `admin` in den Composition Root ist erlaubt (`a-check`: `composition_root` darf jeden Adapter importieren), und kein Test der Konfiguration verlangt sein Fehlen; der Plan nennt die Ausgabe von `git diff` als Fänger. Folge: die Zusage hält allein diese Ausgabe, die Reviewer und Verifier selbst fahren. Ungefangen: 1 von 1 Mutation (ein Import `_ ".../adapters/driving/admin"` in `src/bootstrap/konfig.go`): `make a-check` meldet 0 Befunde, `TestKonfig*` und `TestAuditAufrufstellen` bleiben grün.
- H1, Teilstücke unter 8 Bytes: der Fänger prüft Fenster ab 8 Bytes (`TestH1ProbeSiehtJedenWeg` belegt die Grenze: 8 Bytes gefunden, 7 nicht). Folge im Betrieb: ein Fehler, der nur 1 bis 7 Bytes des Tokens ausgäbe, bliebe grün. Ungefangen: die Längen 1 bis 7 von 36 möglichen Längen eines Teilstücks des Tokens (7 von 36); der Wert selbst und jedes Fenster ab 8 Bytes sind gefangen.
- H1, `stdout`: die Dateien auf dem Weg eines Requests (`admin.go`, `fehler.go`, `route.go`, `lebenszyklus.go`) importieren weder `os` noch `fmt` (das Lint-Profil verbietet `fmt.Print` und `println` zusätzlich, `forbidigo`); die Mutationen schreiben über `println` auf stderr und über `log`. Der Kanal ist belegt (`TestErfassungSiehtAusgabe`: ein Schreiben auf Deskriptor 1, 2 und ins Log wird gezählt). Ungefangen: 1 von 2 Strömen ohne Mutation (stdout); die Mutation im Fehlerpfad von `LadeToken` fängt T25.
- D4: `X-Correlation-Id` steht beim Aufruf von `Append` schon in der Header-Map; entschieden, kein Negativ mehr (`SPEC-238` Nr. 8, §6, Ausgang *entschieden*): der Test lässt allein diesen Schlüssel zu, ein zweiter Header in der Map macht ihn rot.
- Gleichwertig, nicht gefangen: `ErrorCode: code` statt `model.CodeE2101` im Eintrag (nur `ODR-E2101` trägt 401, `SPEC-109` Nr. 2); wächst die Tabelle um einen zweiten Code mit 401, schreibt der Adapter weiter `ODR-E2101` (`SPEC-238` Nr. 8 verlangt es).

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

**Start** (`next` → `in-progress`): [slice-admin-fehlerabbildung](../done/slice-admin-fehlerabbildung.md) liegt in `done/` (Code-Abhängigkeit: D2 läuft über die Fehlerantworten 400 und 500, H1 über Panik und Telemetrie; über ihn transitiv die Teile 2 bis 4), dazu `slice-audit-log` in `done/` (`AuditLog`, `FakeAudit`, Adapter des Audit; geprüft am 2026-11-07, Planner, `main` = `aadbef0`). `slice-audit-dienste` ist keine Kante dieses Teils: der Eintrag zu `ODR-E2101` ist eine eigene Art des Adapters, die vier Arten schreibt allein der Use Case. Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-238` Nr. 8, 12, 13, `SPEC-234` Nr. 2) seit dem Zweig von `slice-admin-listener-token` (`AGENTS.md` §3.12).

**Stand der Bedingungen beim Start** (Planner, 2026-11-21, `main` = `34d74b0d` vor dem Start): `slice-admin-fehlerabbildung` (Code-Abhängigkeit, über ihn transitiv die Teile 2 bis 4), `slice-audit-log` und `slice-telemetry-port` liegen in `done/`; `SPEC-238` Nr. 5, 6 a, 8, 12, 13, 15 samt den Klarstellungen vom 2026-11-20 (Nr. 5 und 8: Header und Audit folgen dem Status, auch bei einem Fehler der Route) und `SPEC-234` Nr. 2 stehen auf `main`; in `in-progress/` läuft sonst kein Slice, kein Worktree; der Slice nennt keinen Pfad der Gate-Kette (Übergangsregel 7, §6), kein Konflikt. Die Sendungen aus §1 sind geprüft (`AGENTS.md` §3.13): (1) die Sendung von `slice-admin-fehlerabbildung` (Audit bei einem 401 der Route) steht in §1 (*Übernimmt* (3)), in der Zeile *Admin* von §3 und im Fänger D1, sein §1 schließt sie nicht aus; (2) H1 ist Zeile H1 mit der Tabelle der Ausgabewege (die lokalen Fänger B8, C1, E5, H2 der Teile 2 bis 5 stehen in `done/`); (3) die Pflichten der Endpunkte (P1) liegen unverändert bei `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces`; (4) die Rückgabe M-1 aus `slice-admin-fehlerabbildung` (Folgeprüfung bei einer Änderung an `finde`) steht in §6 mit Ausgang *weiter offen*. `slice-bootstrap-m6` führt die Zelle `admin_abgewiesen` und den Canary-Fänger gegen `audit.jsonl` (Rückgabe R6, entschieden).

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-07, Teil des Planner 2026-11-21, vor dem ersten Code-Commit): 13 Zeilen der Tabelle, *a* = 3 (D6, D7, N1), *b* = 10, *f* = 0, *p* = 40: *S* = 85·3 + 80·10 + 200·0 + 40 = 1095 gleich der Tabelle, *M* = 100 · ⌈135·1095 / 10000⌉ = 100 · ⌈14,7825⌉ = 1500 gleich der Formel; keine Schnittfrage (135·*S* = 147825 ≤ 300000, *S* < 2223), kein `schnitt-grund`. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Pfade des Blocks: `src/adapters/driving/admin/`, `src/hexagon/ports/support/audit.go`, `src/hexagon/ports/support/audit_test.go`, `src/adapters/driven/audit/audit_test.go`, `src/bootstrap/audit_betrieb_test.go`; `produktcode` 40, `fakes` 0 (`FakeAudit` und `FakeTelemetry` bestehen, §3). Basis der Messung: `58a5b407` (Stand von `main` nach den Moves des Starts; der Commit mit dem Block folgt ihr und ändert keinen Pfad unter `src/`, Messung 0).

```groessenmarke
pfad: src/adapters/driving/admin/
pfad: src/hexagon/ports/support/audit.go
pfad: src/hexagon/ports/support/audit_test.go
pfad: src/adapters/driven/audit/audit_test.go
pfad: src/bootstrap/audit_betrieb_test.go
produktcode: 40
fakes: 0
basis: 58a5b407
schaetzung: 1095
marke: 1500
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 1500 (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein Teil, der sich ohne geteilten Mechanismus abtrennen lässt, steht dann vor dem Code in §4, sonst steht die Messung mit Grund im Plan (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert — Carveout?): der Architect ändert `SPEC-238` Nr. 8 oder 13 oder `SPEC-234` Nr. 2 (Rückgabe an den Planner), oder die Änderung von `audit_betrieb_test.go` verlangt Produktcode in `src/bootstrap` (dann gilt die Alternative aus Rückgabe R2: `Arten()` bleibt bei vier, eine zweite Funktion trägt die Art des Adapters).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-07, der Audit-Anteil der Zeile *Authentisierung*): jede Zelle nennt eine Nummer von `SPEC-238` (Rahmen) oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Audit des Abweisens (`Append` beim 401) | `SPEC-238` Nr. 8: fehlender Header ist ein Verstoß und ein Eintrag (Felder `Art`, `ErrorCode`, `KorrelationsID`, sonst leer) | Nr. 8: leerer Wert und `Bearer` ohne Kandidat ebenso | Nr. 8 und 13: falsches Token, anderes Schema, mehrere Zeilen: ein Eintrag ohne Tokenwert, ohne Teilstück, ohne Pfad und Methode; ein Fehler von `Append` ändert nichts | Nr. 8: `Append` hat keinen `ctx` und läuft zu Ende; der Client kann trennen, der Eintrag steht (`SPEC-234` Nr. 5 c) | Nr. 8 und 12: jede Abweisung ein Eintrag, ohne Drosselung und Zusammenfassung |

**Zusatztabelle: Paare von Verstößen am selben Schlüssel und Modi** (`AGENTS.md` §3.16; die Fänger sind Zeilen der Tabelle in §3):

| Schlüssel und Paar oder Modus | Ergebnis (Ort) | Fänger |
|---|---|---|
| Header: zwei Zeilen und falsches Schema | 401, genau ein Eintrag (`SPEC-238` Nr. 4 a, Nr. 8) | D1 (Schleife über U1) |
| Header: ungültige Form und angehaltener Server | 503, kein Eintrag (Nr. 3 b, Nr. 8) | D2 |
| Modus 401 mit Fehler von `Append` | Antwort und Zählen unverändert (Nr. 8) | D5 |
| Modus Erfolg, 404, 405, 400, 500, `GET /healthz`, Kontextfehler | kein Eintrag (Nr. 8) | D2 |

- Rückgabe R2 (Architect, Auftrag K): `src/bootstrap/audit_betrieb_test.go` zählt `support.Arten()` (vier) und die Zellen Modus × Art; die fünfte Art bricht den Test, Produktcode von `bootstrap` ändert sich nicht — **Ausgang:** entschieden (Planner, 2026-11-07): die Änderung der einen Testdatei (die Art ausdrücklich ausgenommen, Zeile Y6) verletzt die Voraussetzung „Binary unverändert“ nicht; die Alternative (`Arten()` bleibt bei vier, eine zweite Funktion trägt die Art, der Audit-Adapter prüft beide) bleibt verworfen, weil sie den Produktcode des Audit-Adapters ändert und die Tabelle von `SPEC-234` Nr. 2 halbiert.
- Rückgabe R5 (User): jede Abweisung mit `ODR-E2101` schreibt einen Eintrag in `audit.jsonl`, ohne Drosselung (`SPEC-238` Nr. 8): ein Client ohne Token füllt die Datei; dazu die offene Begrenzung der Fehlversuche aus `LH-FA-13.a` — **Ausgang:** weiter offen als benannte Lücke des Lastenhefts; Adresse: User (Owner); kein Code, bis entschieden. Sichtbar gehalten in der Roadmap und in `welle-admin-api` (§6 dort).
- Rückgabe R6 (Planner): `slice-bootstrap-m6` ruft `admin.LadeToken` (`SPEC-239`), sortiert die Meldung `ODR-E2001 ADMIN_TOKEN_FILE: <Grund>` unter den Schlüssel (`SPEC-110` Nr. 12), reicht `admin.Konfig`, nimmt die Zelle `admin_abgewiesen` in den Test Modus × Art auf (Y6), ordnet das Herunterfahren ein (`SPEC-136` Nr. 15) und setzt den Canary-Fänger gegen `audit.jsonl` auf der Feldmenge von `SPEC-238` Nr. 8 auf — **Ausgang:** entschieden (Planner, 2026-11-07): eingetragen in §1 und DoD von `slice-bootstrap-m6` mit `slice-admin-listener-token` als Geber.
- Die Zusage „keine Sperre über `Vergleich`“ (N1) ist erst mit diesem Teil gefangen — **Ausgang:** entschieden (Planner, Schnitt): akzeptiert bis zu diesem Teil; die Zeile liegt hier, weil die Schleife `Append` und `CountError` einschließt.
- `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×): D6, D7 und N1 sind Zusagen über Nebenläufigkeit; jede Mutation muss in jedem Lauf rot werden (`AGENTS.md` §3.10) — **Ausgang:** weiter offen → `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): `FakeAudit` hält jedes Argument und zählt jede Operation, auch die, die der Rahmen nie rufen darf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (44×): P1 (Pflichten der Endpunkte) ist ein akzeptiertes Negativ mit Grund und Folge, kein Fänger — **Ausgang:** weiter offen → `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` im Register.
- Größe: der Block in §4 trägt den Teil des Architect (13 Zeilen, *S* = 1095, *M* = 1500 nach `SPEC-128` Nr. 1) und den des Planner (`basis` `58a5b407`, `schaetzung` 1095, `marke` 1500) — **Ausgang:** entfallen: gesetzt in einem Commit vor dem ersten Code-Commit (Planner, 2026-11-21), `make groessenmarke` grün; keine Schnittfrage. Messung nach dem Code (Implementer, 2026-11-21): `git diff --numstat 58a5b407..<Handoff-Commit>` über die fünf Pfade des Blocks, Tests eingeschlossen: 1369 Zeilen (1316 Einfügungen, 53 Löschungen), unter der Marke 1500 (0,91); kein Eintrag *ueberschreitung* nötig.
- Rückgabe M-1 aus `slice-admin-fehlerabbildung` (Zyklus-Schutz in `findeAuf` an der Adresse statt an Typ und Adresse; akzeptiertes Negativ dort, Adresse Planner „bei der nächsten Änderung an `finde` (`fehler.go`)“): dieser Teil ruft `finde` nicht neu und ändert weder `finde` noch `findeAuf` (der Eintrag bei 401 liegt in `schreibeFehler`, hinter der Abbildung) — **Ausgang:** weiter offen (Adresse Planner, Auslöser nicht eingetreten): beim Handoff prüft der Implementer `git diff 58a5b407..HEAD -- src/adapters/driving/admin/fehler.go` auf eine Änderung an `finde` oder `findeAuf` und nennt das Ergebnis im Bericht; berührt der Teil die Suche doch, hält er an und gibt die Folgeprüfung (Schlüssel `(reflect.Type, Adresse)`, `TestZyklusInDerSuche` um die drei Formen als Schleife, Mutation „Schlüssel nur Adresse“ rot, `AGENTS.md` §3.17) an den Planner zurück, statt sie im Code zu entscheiden (`AGENTS.md` §3.12, §3.13). Geprüft beim Handoff (Implementer, 2026-11-21): `git diff 58a5b407..HEAD -- src/adapters/driving/admin/fehler.go` zeigt eine Änderung allein in `schreibeFehlerDerRoute` (der Parameter `id`, dazu der Satz im Kommentar); `finde`, `findeAuf` und `zeigerAdresse` sind unverändert, der Auslöser ist nicht eingetreten.
- Bestehende Prüfungen mit `Audit` 0 bei einer Antwort 401 (`nurZaehlung`, `TestZaehlenJedeSchreibstelle`, `pruefeFehlerDerRoute`, §3 *Tests*): ihre Umschreibung ist Teil dieses Slice (Vorbehalt von `slice-admin-rahmen` N-2) — **Ausgang:** entfallen (Implementer, 2026-11-21): umgeschrieben sind `nurZaehlung` in `helfer_test.go` (Audit gleich der Zahl der Zählungen von `ODR-E2101`, bei jedem anderen Code 0; damit auch `pruefeFehlerDerRoute` mit dem Paar `ODR-E2101`, 401 und alle Prüfungen von `TestMethodeUndPfad`, `TestFehlerVorAntwort`, `TestStatusDerAntwort`) und `TestZaehlenJedeSchreibstelle` in `schreibstellen_test.go` (Stelle 401: ein Aufruf, jede andere: keiner); der Lauf von `make test` fand keine weitere rote Prüfung im Paket `admin`. D2 ist nicht geschwächt: eine Mutation „`Append` auch bei 404, 503, 500“ ist in den umgeschriebenen Prüfungen selbst rot (`D2-bei-404-alt`, `D2-bei-503-alt`, `D2-bei-500-alt`, §3), und die Prüfung des Paares `ODR-E2101` mit 401 verlangt jetzt den einen Eintrag (`D1-nur-authentisierung-alt` rot). Ein Befund des Reviews bleibt möglich; das Risiko selbst ist aufgelöst.
- D4 liest „Header leer“ (Zeile D4, Plan §3) als: bis `Append` zurückkehrt, ist weder `WriteHeader` noch `Write` gerufen, der Body leer, und kein Header des Fehlerobjekts gesetzt (`Content-Type`, `Content-Length`, `WWW-Authenticate`); `X-Correlation-Id` steht in der Header-Map, weil `ServeHTTP` die ID vor allem anderen setzt (`SPEC-238` Nr. 3 a, Nr. 5 „jede Antwort trägt `X-Correlation-Id`“). `TestAuditVorDemSchreiben` lässt allein diesen Schlüssel zu. Das Wort „Header“ in Nr. 8 kann auch die Map meinen (dann stünde die ID erst nach `Append` darin) — **Ausgang:** entschieden (Architect, 2026-11-21): die Lesung des Implementers gilt; sie folgt aus Nr. 3 a (die ID steht vor allem anderen fest und gehört zum Eintrag) und aus dem Zweck von Nr. 8 (nichts geht an den Client, bevor der Eintrag steht); `SPEC-238` Nr. 8 trägt sie jetzt im Wortlaut (Klarstellung, keine Entscheidung geändert), Reihenfolge in `ServeHTTP` und Test bleiben.
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

**Vorgelagert — offene Beobachtungen sichten:** Register beim Schnitt nachgezogen (Planner, 2026-11-07, `main` = `aadbef0`, Zähler aus den Dateien unter `evidence/`): einschlägig sind `nebenlaeufiger-lauf-nur-im-glueck-gruen` 13×, `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 44×, `akzeptiertes-negativ-ohne-folge` 15×, `betriebsregel-fuer-repo-dateien-verletzt` 9×, `randform-erst-im-review-entschieden` 34×, `rate-ohne-detektor-schwankt-mit-der-last`, `schaetzformel-ueberschaetzt-den-slice` 21×, `report-der-rolle-faellt-docs-check` 5× (Verkörperung, soweit die Schwelle erreicht ist, beim Lese-Schritt der Welle-Closure); kein neues Verzeichnis, kein Zähler gesetzt.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
