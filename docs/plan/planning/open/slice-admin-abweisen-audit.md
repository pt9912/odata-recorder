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

**Berührte Spec-Stellen:** `LH-FA-13.a` (jedes Abweisen mit `ODR-E2101` hinterlässt einen Audit-Eintrag ohne Tokenwert) · `LH-QA-04.a` · `SPEC-238` Nr. 8 (Audit des Abweisens), Nr. 12 (Gleichzeitigkeit), Nr. 13 (Geheimnisse, acht Ausgabewege), Nr. 14 (Binary und Katalog), Nr. 15 (Pflichten der Endpunkte) · `SPEC-234` Nr. 2 (fünfte Art `admin_abgewiesen`), Nr. 5, Nr. 7 (der Adapter schreibt die vier Arten des Use Case nie) · `SPEC-110` Nr. 2 (unberührt) · `ARC-007`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie. `SPEC-238` steht auf `main` (seit dem Merge von `slice-admin-listener-token`).

**Verantwortlich:** —

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

**Übernimmt (Sendungen von `slice-admin-listener-token`, Schnitt vom 2026-11-07, `AGENTS.md` §3.13):** (1) die Änderung der Testdatei `src/bootstrap/audit_betrieb_test.go` — Rückgabe R2 des Architect, vom Planner bestätigt (nur Testdatei, kein Produktcode von `bootstrap`; die Voraussetzung „Binary unverändert“ gilt); (2) den Audit-Eintrag zu `ODR-E2101` aus `LH-FA-13.a`, dessen Fänger `slice-admin-authentisierung` (U1) bis hierher offen lässt.

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

- [ ] Audit des Abweisens: die Art `admin_abgewiesen` als fünfte Zeile von `support.Arten()` (Pflicht allein `error_code`) und im Adapter des Audit (jedes verbotene Feld und das fehlende `error_code` abgewiesen, Zeile Byte für Byte), genau ein `Append` je 401 mit den Feldern `Art`, `ErrorCode`, `KorrelationsID` und sonst leer, kein `Append` bei einer anderen Antwort und nie eine der vier Arten des Use Case, vor dem Schreiben von Header und Body, auf der Goroutine des Requests ohne Frist, Fehler von `Append` ändert nichts, jede Abweisung ein Eintrag (D1 bis D7, Y5).
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
| `src/adapters/driving/admin/` | update | `Audit.Append` im 401-Zweig, vor dem Schreiben, auf der Goroutine des Requests (`SPEC-238` Nr. 8) |
| `src/hexagon/ports/support/audit.go`, `src/hexagon/ports/support/audit_test.go` | update | Art `ArtAdminAbgewiesen` als fünfte Zeile von `Arten()`; Orakel des Tests auf fünf Arten |
| `src/adapters/driven/audit/audit_test.go` | update (nur Test) | Schleife über Felder × Arten mit der fünften Art und die Zeile des Vertrags |
| `src/bootstrap/audit_betrieb_test.go` | update (nur Test, Rückgabe R2) | zählt vier Arten und die Zellen Modus × Art: fünf Arten, `admin_abgewiesen` ausdrücklich ausgenommen (kein Use Case, kein Aufruf im Binary); kein Produktcode von `bootstrap` ändert sich |
| Tests | neu / update | Adapter gegen `ServeHTTP` und gegen einen echten Listener; `FakeAudit` (hält jedes Argument, zählt `Aufrufe()`, `Vor` als Haltepunkt, `Fehler` einstellbar) und `FakeTelemetry` aus `supporttest`; Pipes für `stdout`, `stderr` und `log.Default()`; der Adapter importiert weder Adapter noch Service (`make a-check`) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stammen vom Architect (2026-11-07, Plan von [slice-admin-listener-token](../done/slice-admin-listener-token.md)); der Schnitt gibt jedem Teil ganze Zeilen im Wortlaut (D Audit, H Geheimnisse, N Gleichzeitigkeit, Y Katalog und Code, P Pflichten der Endpunkte). Die Mengen *R* und *W* und die Mengen M-Header (U1) und M-Methode × Pfad (U7), über die D1 und D2 laufen, führt der Plan, der die jeweilige Zeile trägt (`slice-admin-authentisierung`, `slice-admin-fehlerabbildung`).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| D1 | (Nr. 8) Bei jeder Antwort 401 genau ein `Append(Eintrag{Art: ArtAdminAbgewiesen, ErrorCode: CodeE2101, KorrelationsID: ID der Antwort})`, alle übrigen Felder leer | Schleife über die 20 Formen von U1, die Zellen von U7 und `Vergleich` gleich `false`: `FakeAudit.Eintraege()` gleich dem Literal-Eintrag mit der ID, `Aufrufe()` 1 | `ServiceID` gesetzt; Art falsch | andere |
| D2 | (Nr. 8) Bei keiner anderen Antwort ein `Append`, und nie eine der vier Arten von `SPEC-234` | Schleife über 404, 405, 400, 503, 500, healthz, Erfolg, Kontextfehler: `FakeAudit.Aufrufe()` 0 | `Append` auch bei 404 | andere |
| D3 | (Nr. 8) Der Eintrag trägt keinen Tokenwert, kein Teilstück, keinen Pfad, keine Methode, keine Adresse | Canary im falschen Token, im Pfad, in der Query, in einem Header, in der Methode: der gehaltene `Eintrag` ist als Struktur gleich dem Literal (alle sieben Felder) | der Pfad in `ServiceID` | andere |
| D4 | (Nr. 8) Bis `Append` zurückkehrt, hat der Adapter weder Header noch Body geschrieben | `FakeAudit.Vor` prüft einen `ResponseRecorder` (`ServeHTTP` direkt): Header leer, Body leer | `Append` nach dem Schreiben | andere |
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

**Start** (`next` → `in-progress`): [slice-admin-fehlerabbildung](slice-admin-fehlerabbildung.md) liegt in `done/` (Code-Abhängigkeit: D2 läuft über die Fehlerantworten 400 und 500, H1 über Panik und Telemetrie; über ihn transitiv die Teile 2 bis 4), dazu `slice-audit-log` in `done/` (`AuditLog`, `FakeAudit`, Adapter des Audit; geprüft am 2026-11-07, Planner, `main` = `aadbef0`). `slice-audit-dienste` ist keine Kante dieses Teils: der Eintrag zu `ODR-E2101` ist eine eigene Art des Adapters, die vier Arten schreibt allein der Use Case. Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-238` Nr. 8, 12, 13, `SPEC-234` Nr. 2) seit dem Zweig von `slice-admin-listener-token` (`AGENTS.md` §3.12).

**Größenmarke** (Planner, 2026-11-07, `SPEC-128` Nr. 1; Prosa-Form, der Block folgt beim Start): 13 Zeilen der Tabelle, *a* = 3, *b* = 10, *f* = 0, *p* = 40: *S* = 85·3 + 80·10 + 200·0 + 40 = 1095, *M* = 1500, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Pfade des Blocks beim Start: `src/adapters/driving/admin/`, `src/hexagon/ports/support/audit.go`, `src/hexagon/ports/support/audit_test.go`, `src/adapters/driven/audit/audit_test.go`, `src/bootstrap/audit_betrieb_test.go`; `produktcode` 40, `fakes` 0. `basis`, `schaetzung` und `marke` setzt der Planner beim Start in einem Commit vor dem ersten Code-Commit.

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
- Größe: die Marke steht als Prosa in §4 (*S* = 1095, *M* = 1500) — **Ausgang:** offen; Adresse: Planner beim Start (Block mit `basis`, `schaetzung` und `marke` in einem Commit vor dem ersten Code-Commit).
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
