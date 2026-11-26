# Slice slice-bootstrap-m6-listener: Admin-Listener im Composition Root: `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE`, Token-Datei, Bindung und Herunterfahren beider Listener, Handbuch und Katalog (Teil B des Schnitts von `slice-bootstrap-m6`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-RB-02`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen** (Teil B des Schnitts von [slice-bootstrap-m6](slice-bootstrap-m6.md)): `LH-FA-13.a` · `SPEC-300` Nr. 1, 2 (für `ADMIN_LISTEN`, `ADMIN_TOKEN_FILE` und den Präfix), 4, 7 bis 9, 11, 12 und 15 ohne TLS (Admin-Listener im Composition Root, Architect 2026-11-23) · `SPEC-110` Nr. 2, 7, 10 bis 13 (Schlüssel dieser Version, `ADMIN_LISTEN`, Zusammenhang, Dateien, Meldung) · `SPEC-139` Nr. 6 bis 8 (Hochfahren, Herunterfahren, Start) · `SPEC-136` Nr. 15 · `SPEC-008`, `SPEC-056`, `SPEC-077` · `SPEC-238` Nr. 3 c, 8, 11 · `SPEC-239` (Token-Datei) · `SPEC-299` Nr. 3 und 6 (der Zwischenstand, Zeile B19) · `SPEC-037` Nr. 10 (Fall `ADMIN_LISTEN` von `TestE2EProzess`) · `SPEC-200`, `SPEC-201` Nr. 4, 9 e, 11 (Handbuch und Katalog) · `SPEC-005` · `ARC-007`, `ARC-012` · `architecture.md` §2.5, §4.2, §4.5, §4.9.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-bootstrap-m6](slice-bootstrap-m6.md), Vorschlag des Architect vom 2026-11-23). **Datum:** 2026-11-23.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Composition Root liest `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE`, lädt oder erzeugt die Token-Datei als ersten Schritt von `Hochfahren` (vor der Startwartung und vor jedem Listener), reicht `admin.Konfig` mit allen Feldern an den Admin-Adapter und startet den Admin-Listener in jedem Modus neben dem OData-Listener (beide Bindungen vor dem ersten `Serve`; geordnetes Herunterfahren beider, Fehlerkanal, `GET /healthz` ohne Token); der Listener läuft auf Loopback, ein `ADMIN_LISTEN` außerhalb von Loopback endet mit `ODR-E2001 ADMIN_LISTEN` (Zeile B19). Handbuch und Katalog tragen die zwei Schlüssel, die Endpunkte und die zwölf Codes (`SPEC-300` Nr. 1, 2, 4, 7 bis 9, 11, 12, 15). **Zwischenstand** (Architect, 2026-11-23): bis Teil C ([slice-bootstrap-m6-transport](slice-bootstrap-m6-transport.md)) bleiben `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT` „nicht unterstützt“, während `SPEC-110` Nr. 2 sie als geprüft führt (die Spezifikation steht vor dem Code, wie bei `SPEC-299`); der Listener bindet in diesem Stand nur auf Loopback (Zeile B19), und das Handbuch dieses Teils nennt die drei Schlüssel nicht. **Schnitt vor dem Code** (Vorschlag des Architect, Planner, 2026-11-23): die Schätzung des ungeschnittenen Plans (*S* = 4366) liegt über der Schnittgrenze von `SPEC-128` Nr. 1; Teil A [slice-bootstrap-m6-endpunkte](../next/slice-bootstrap-m6-endpunkte.md) (Endpunkte im Root, A01 bis A07) liegt vor diesem Teil, Teil B (diese Datei, B01 bis B19), Teil C (C01 bis C08) und Teil D [slice-bootstrap-m6](slice-bootstrap-m6.md) (M6-Test, D01 bis D06) folgen.

**Übernimmt:** das Lesen und Prüfen von `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE` samt Aufheben ihrer Ablehnung in `SPEC-110` Nr. 2 (Sendung von `slice-admin-listener-token`, dessen §1 *Ausdrücklich NICHT* den Root nennt; die drei Schlüssel des Transportschutzes liegen bei Teil C) und den Admin-Listener, den `slice-bootstrap-m2` der Folgewelle zuweist (`AGENTS.md` §3.13).

**Übernimmt außerdem** (Sendung von `slice-admin-listener-token`, Rückgaben R2 und R6 des Architect, Schnitt vom 2026-11-07, `AGENTS.md` §3.13; beim Schnitt vom 2026-11-23 an diesen Teil gelegt, §1 *Ausdrücklich NICHT* und DoD dieses Slice vorher gelesen, kein Ausschluss trifft die Sendung): der Composition Root (a) ruft `admin.LadeToken` (`SPEC-239`) mit dem Pfad aus `ADMIN_TOKEN_FILE` (Default `<DATA_DIR>/admin.token`, als erster Schritt von `Hochfahren`, vor der Startwartung und vor jedem Listener, `SPEC-300` Nr. 4) und sortiert die Meldung `ODR-E2001 ADMIN_TOKEN_FILE: <Grund>` unter den Schlüssel ein (`SPEC-110` Nr. 12); (b) reicht `admin.Konfig` mit `Listener`, `Token`, `IDs`, `Audit`, `Telemetrie` und den Routen der Endpunkt-Slices (die Routen liefert Teil A, Zeile A01); (d) ordnet den Admin-Listener in die Reihenfolge des Herunterfahrens ein (`SPEC-136` Nr. 15; `Anhalten` und `Schliessen` des Adapters liefert `slice-admin-lebenszyklus`); (e) setzt den Canary-Fänger gegen `<DATA_DIR>/audit.jsonl` auf der Feldmenge von `SPEC-238` Nr. 8 auf (`Art`, `ErrorCode`, `KorrelationsID`, sonst leer; Zeile `{"timestamp":"…","correlation_id":"cor-…","art":"admin_abgewiesen","error_code":"ODR-E2101"}`; Zeilen B15 und B16). Die Zelle `admin_abgewiesen` im Test Modus × Art (Punkt (c) der Sendung) liegt bei Teil D (Zeile D03). Dazu sind zwölf der dreizehn Zeilen von `harness/katalog-ausnahmen.txt` aufzulösen (Berichtigung des Architect, 2026-11-23: die Sendung nannte allein `ODR-E2102`): das Binary erreicht den Listener mit diesem Teil, also trägt das Handbuch `ODR-E2101`, `ODR-E2102`, `ODR-E6101`, `ODR-E6102`, `ODR-E7002`, `ODR-E7004` bis `ODR-E7008`, `ODR-E7010` und `ODR-E7011`; `ODR-E7012` bleibt Ausnahme mit berichtigtem Grund (`SPEC-300` Nr. 15; Zeile B17).

**Übernimmt außerdem** (Sendung von `slice-audit-verdrahtung`, beim Schnitt an diesen Teil gelegt): den Fänger gegen einen Tokenwert in `<DATA_DIR>/audit.jsonl`. Erst hier trägt der Prozess den Wert aus `ADMIN_TOKEN_FILE`; `Build` erhielt bis dahin kein Feld dafür (`SPEC-139` *Ungeprüft*; Zeile B16).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Dienste, Routen und Einstellungen der Diagnose im Root und der Ersatz der vier Wächter — [slice-bootstrap-m6-endpunkte](../next/slice-bootstrap-m6-endpunkte.md) (Teil A, liegt bei diesem Start in `done/`); dieser Teil ruft die Routen von A auf und ändert sie nicht.
- Die Schlüssel `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT`, die vier Aufrufe von `SPEC-299` Nr. 8, die Vereinigung der Verstöße beider TLS-Paare, die Warnung am Start, die Umhüllung des Listeners, das Handbuch zum TLS und die Zelle Modus × Weg — [slice-bootstrap-m6-transport](slice-bootstrap-m6-transport.md) (Teil C, Zeilen C01 bis C08; §1 und DoD dort tragen die Sendung von `slice-admin-tls`). Der Zwischenstand von Zeile B19 bis dahin steht in §6.
- Der M6-Test: Rundlauf über beide Listener, Abweisung ohne Zustandsänderung, die Zellen Modus × Art (`admin_abgewiesen`, `session_angelegt`, `session_beendet`, `import_veroeffentlicht`, `workspace_angelegt`, `workspace_zurueckgesetzt`, `snapshot_erzeugt`, `workspace_wiederhergestellt`) und die Stufe 11 `TestE2EAdmin` mit den Negativen von `LH-FA-14` — [slice-bootstrap-m6](slice-bootstrap-m6.md) (Teil D, Zeilen D01 bis D06).
- `HYBRID_*`, `FALLBACK_*` und der Modus `hybrid` — geliefert von `slice-bootstrap-m4-konfig` und `slice-bootstrap-m4` (beide in `done/`); dieser Teil ändert sie nicht und fährt `hybrid` als einen der vier Modi jeder Menge. `LISTEN_TLS_*` — geliefert von `slice-listener-tls-verdrahtung` (`done/`).
- Die Begrenzung der Fehlversuche (`LH-FA-13.a` nannte keine) — entschieden (User, 2026-11-24, Variante A, `SPEC-301`); Träger ist der Folge-Slice [slice-admin-abweisen-begrenzen](slice-admin-abweisen-begrenzen.md), Vorlage in §6 von [slice-bootstrap-m6](slice-bootstrap-m6.md); dieser Teil führt jede Abweisung als einen Eintrag (`SPEC-238` Nr. 8, Zeile B15).
- Der Handshake des Admin-Adapters über TLS (Zeilen AT11 bis AT17 von `slice-admin-tls`) — [slice-admin-tls-betrieb](slice-admin-tls-betrieb.md), startet nach allen vier Teilen.
- Neue Endpunkte oder Fachlogik — der Teil verdrahtet, was die Slices der Welle liefern. Das Umschalten des Modus zur Laufzeit — `LH-FA-13` schließt es aus; der Rundlauf (Teil D) wechselt den Modus durch Neustart.

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

- [ ] Verdrahtung und Konfiguration: der Admin-Listener läuft in jedem Modus neben dem OData-Listener; die Werte von `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE` kommen aus dem Wert des Starts bis in Adapter und Listener (Fänger, dass der Root sie weiterreicht und den Nullwert nicht stehen lässt; Muster aus `slice-upstream-tls`); Reihenfolge Token-Datei, Startwartung, Bindung beider Listener vor dem ersten `Serve` (Berichtigung des Architect, 2026-11-23, `SPEC-300` Nr. 4 und 7: die Zeile im Plan davor setzte die Startwartung vor die Token-Datei), ein Fehler der Token-Datei öffnet keinen Listener und lässt die Startwartung ungelaufen; geordnetes Herunterfahren beider Listener (`Anhalten` und `Schliessen` des Admin-Adapters, `SPEC-136` Nr. 15); `admin.LadeToken` mit dem Pfad aus `ADMIN_TOKEN_FILE`, die Meldung `ODR-E2001 ADMIN_TOKEN_FILE: <Grund>` unter dem Schlüssel (`SPEC-110` Nr. 12) und `admin.Konfig` mit allen Feldern (Geber `slice-admin-listener-token`, Rückgabe R6); ein `ADMIN_LISTEN` außerhalb von Loopback endet im Zwischenstand mit `ODR-E2001 ADMIN_LISTEN` (Zeilen B01 bis B13 und B19 der Tabelle in §3, vom Architect).
- [ ] Fänger des Tokens und der Abweisung am Prozess: `GET /healthz` ohne Token in jedem Modus, jede Abweisung mit 401 `ODR-E2101` genau eine Zeile `admin_abgewiesen` auf der Feldmenge von `SPEC-238` Nr. 8, und der Tokenwert aus `ADMIN_TOKEN_FILE` steht in keinem Byte von `<DATA_DIR>/audit.jsonl`, `stdout`, `stderr`, einer anderen Datei unter `DATA_DIR` und einer Antwort, auch nicht als Hälfte (Canary; Geber: `slice-audit-verdrahtung` und `slice-admin-listener-token`; Zeilen B14 bis B16).
- [ ] Jeder andere Schlüssel mit dem Präfix `ADMIN_` bleibt `ODR-E2001` („nicht unterstützt“); die drei Schlüssel von `SPEC-299` bleiben es bis Teil C; `HYBRID_*`, `FALLBACK_*` und `LISTEN_TLS_*` sind geliefert (ihre Slices liegen in `done/`); die Spezifikation führt den Nachzug von `SPEC-110` Nr. 2 und den Vertrag `SPEC-300` (geliefert, 2026-11-23; Zeile B03).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-300`, `SPEC-110`, `SPEC-139`, `SPEC-136` Nr. 15, `SPEC-238` stehen vor dem Start auf `main`; dieser Teil schreibt keine Spezifikation).
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand: Admin-Listener, Token und Token-Datei, die Schlüssel `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE`, Endpunkte, die zwölf Fehlercodes, die aus `harness/katalog-ausnahmen.txt` entfallen (`ODR-E7012` bleibt mit berichtigtem Grund; Satz und Stellen: `SPEC-201` Nr. 9 e und 11, `SPEC-300` Nr. 15), belegt durch ein Beispiel als Datei gegen das gebaute Binary (`SPEC-200`); der Präfix `ADMIN_` ist nicht mehr in `nichtGelieferte`, Zeile B17; die drei Schlüssel von Teil C nennt das Handbuch noch nicht); das Kataloge-Gate ist grün. Voraussetzung ist `slice-handbuch-meldungskatalog` in `done/` (Trigger von `welle-mock`; liegt dort).
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
| `src/bootstrap/` (`wurzel.go`, `konfig.go`, neu `konfig_admin.go`, `konfig_werte.go`, `lebenszyklus.go`, `start.go`) | update | `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE` lesen, prüfen und verdrahten; zweiter Listener neben dem OData-Listener; Token-Datei; Reihenfolge von Start und Herunterfahren (`SPEC-300`) |
| `src/bootstrap/` (Tests: `export_test.go`, `haken_test.go`, `lebenszyklus_test.go`, `konfig_test.go`, neu `admin_listen_test.go`, `admin_token_test.go`, `admin_start_test.go`, `admin_herunterfahren_test.go`, `admin_geheimnis_test.go`) | update und neu | Naht (`NeuerAdmin`, Eingabe des Transports, Token); die Fänger der Zeilen unten |
| `tests/e2e/stufe_prozess_test.go` | update | der Fall `ADMIN_LISTEN` von `TestE2EProzess` trägt `0.0.0.0:<Port>` statt `127.0.0.1:1`, weil `127.0.0.1:1` mit dem Lesen des Schlüssels ein gültiger Wert ist (im Zwischenstand endet `0.0.0.0` mit `ODR-E2001 ADMIN_LISTEN`, Zeile B19) |
| `harness/katalog-ausnahmen.txt`, Handbuch (`docs/user/benutzerhandbuch.md`), `SPEC-201` Nr. 9 e | update | zwölf der dreizehn Zeilen entfallen, ihre Codes stehen im Handbuch, `ODR-E7012` bleibt mit berichtigtem Grund; die zwei Schlüssel, die Endpunkte, das Token und das Kapitel zum Admin-Listener (`SPEC-300` Nr. 15) |
| `spec/spezifikation.md`, `harness/vertraege/SPEC-037-test-integration.md` | geliefert (Architect, 2026-11-23) | `SPEC-300` neu; Nachzug von `SPEC-110` Nr. 2, 7, 10 bis 13, `SPEC-139`, `SPEC-136` Nr. 15, `SPEC-238` Nr. 14, `SPEC-299` Nr. 14; `SPEC-037` Nr. 7 und 10 a; dieser Teil schreibt sie nicht |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Nachzug in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Sie stehen unten (Architect, 2026-11-23, Zeilen B01 bis B19 der Tabelle des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; jede Zeile steht in genau einem der vier Teile).

**Entscheidungen des Architect** (2026-11-23; die Teile des ungeschnittenen Plans, die den Listener betreffen; Auftrag → Ort in der Spezifikation):

| Auftrag | Entscheidung | Ort |
|---|---|---|
| 1 Nachzug von `SPEC-110` | die fünf Schlüssel `ADMIN_*` sind geprüft (die drei des Transportschutzes in diesem Stand noch „nicht unterstützt“, siehe Zwischenstand), jeder andere `ADMIN_*` bleibt „nicht unterstützt“; `ADMIN_LISTEN` hat die Form von `LISTEN` (Default `127.0.0.1:8081`); Port 0 ist ungültig, ein Test bindet `127.0.0.1:0` in einem Literal von `Build` und liest `AdminAdresse()`, der Prozess-Test wählt vorher einen Port; die Token-Datei liest `Hochfahren`, nie das Laden | `SPEC-110` Nr. 2, 7, 10 bis 13; `SPEC-300` Nr. 2, 4 |
| 3 Reihenfolge beider Listener | Token-Datei als erster Schritt von `Hochfahren`; beide Listener binden vor dem ersten `Serve`; ein Bind-Fehler schließt den anderen; `Anhalten` Admin, dann OData, beide vor `CloseAll`; `Schliessen` OData mit `rest`, dann Admin mit dem Rest der Frist | `SPEC-300` Nr. 4, 7, 8, 9; `SPEC-139` Nr. 6, 7 |
| neu: leerer `AdminListen` im Literal | kein Admin-Listener, keine Token-Datei: die Tests der Slices davor binden keinen Port 8081 | `SPEC-300` Nr. 11 |
| neu: Katalog | zwölf von dreizehn Ausnahmen entfallen (nicht nur `ODR-E2102`), `ODR-E7012` bleibt (kein Endpunkt, `SPEC-236` Nr. 9) | `SPEC-300` Nr. 15 |

**Zusage · Fänger** (`AGENTS.md` §3.15; Architect, 2026-11-23; `Nr.` ist die Kennung des ungeschnittenen Plans, `A` Endpunkte im Root ([slice-bootstrap-m6-endpunkte](../next/slice-bootstrap-m6-endpunkte.md)), `B` Admin-Listener (diese Datei), `C` Transportschutz ([slice-bootstrap-m6-transport](slice-bootstrap-m6-transport.md)), `D` M6-Test ([slice-bootstrap-m6](slice-bootstrap-m6.md)); die Spalte *Zusage* nennt den Ort in `SPEC-300` und den Verträgen; die Tests laufen im Paket `bootstrap_test` ohne Netz außer dem Loopback des Prozesses, Pfad und Namen der Naht in `export_test.go` bestimmt der Implementer, die Mengen stehen mit Achsen unten; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10; **nicht in der Tabelle und nicht in der Marke:** die Begrenzung der Fehlversuche, Ausgang *entschieden* (Folge-Slice [slice-admin-abweisen-begrenzen](slice-admin-abweisen-begrenzen.md), §6 von [slice-bootstrap-m6](slice-bootstrap-m6.md))):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| B01 | (`SPEC-300` Nr. 2, `SPEC-110` Nr. 7) `ADMIN_LISTEN`: ungesetzt, leer oder nur Leerraum ist `127.0.0.1:8081`; die Form ist die von `LISTEN` (derselbe Prüfer, Port 1 bis 65535, Port 0 ungültig); der Grund nennt nie den Wert | `TestAdminListenLesen`: Achsen Schlüssel {`LISTEN`, `ADMIN_LISTEN`} × Menge der Werte, die der Test von `LISTEN` in `konfig_test.go` führt (gültig und ungültig, Zählprüfung gleiche Zahl) × Modus 4; dazu ungesetzt, leer, Leerraum, Port 0, Canary im Wert (steht in keiner Zeile) | Default `8080` (rot); Port 0 gültig (rot); Prüfer ohne Bereich (rot); Wert in der Meldung (rot); nur im Modus `record` gelesen (rot: 3 Modi) | andere |
| B02 | (`SPEC-300` Nr. 2) `ADMIN_TOKEN_FILE`: ungesetzt, leer oder Leerraum ist `<DATA_DIR>/admin.token`, gesetzt der Pfad ohne Leerraum am Rand, ein relativer Pfad bleibt relativ; die Konfiguration trägt nie den Inhalt, und das Laden liest und legt die Datei nicht an | `TestAdminTokenDateiLesen`: 5 Wertarten (ungesetzt, leer, Leerraum, absoluter Pfad, relativer Pfad) × 4 Modi = 20 Zellen; Orakel das Feld `AdminTokenFile`; nach dem Laden besteht keine Datei `admin.token`; kein Feld der Konfiguration trägt Bytes des Tokens (Canary im Dateiinhalt, `fmt.Sprintf("%+v")` der Konfiguration) | Default ohne `DATA_DIR` (rot); Leerraum bleibt (rot); das Laden legt die Datei an (rot); der Inhalt steht in der Konfiguration (rot) | andere |
| B03 | (`SPEC-300` Nr. 2, `SPEC-110` Nr. 2) Jeder andere Schlüssel mit dem Präfix `ADMIN_` ist gesetzt `ODR-E2001 <Name>: wird von dieser Version nicht unterstützt`, leer kein Verstoß; die fünf geprüften Schlüssel sind es nie; die Schreibweise zählt | `TestAdminPraefix`: 8 Namen (`ADMIN_TOKEN`, `ADMIN_`, `ADMIN_LISTEN2`, `ADMIN_TLS`, `ADMIN_TLS_CA`, `ADMIN_PASSWORD`, `ADMIN_TOKEN_FILE_X`, `ADMIN_X`) × {gesetzt, leer} = 16 Zellen, die 5 geprüften Schlüssel (nie „nicht unterstützt“) und 3 andere Schreibweisen (`admin_listen`, `XADMIN_LISTEN`, `ADMIN`, ignoriert) = 24 Zellen, je Zelle als Literal | die Regel entfällt (rot: 8 Zellen); sie trifft die fünf (rot: 5); die Schreibweise zählt nicht (rot: 1); ein leerer Wert ist ein Verstoß (rot: 8) | andere |
| B04 | (`SPEC-300` Nr. 4, `SPEC-239`) `Hochfahren` lädt oder erzeugt die Token-Datei als ersten Schritt; ein zweiter Start auf demselben `DATA_DIR` lädt dasselbe Token; es erreicht `admin.Konfig.Token` und sonst nichts | `TestAdminToken`: 4 Modi × {Datei fehlt (erzeugt: 64 Zeichen `0-9a-f` und `\n`, `0600`), Datei gültig mit 32 Zeichen, Datei ein Symlink auf eine gültige Datei} = 12 Zellen und der zweite Start je Modus (gleiche Bytes, Datei unverändert) = 4; Orakel Dateiinhalt, getrimmt gleich dem `Token` des Haken `NeuerAdmin` | das Token nicht weitergereicht (rot: `admin.New` scheitert, 12 Zellen); je Start neu erzeugt (rot: 4); Pfad ignoriert (rot: die Zellen mit gesetztem Pfad); das Laden erzeugt die Datei (rot) | andere |
| B05 | (`SPEC-300` Nr. 4, `SPEC-239` Nr. 6) Ein Verstoß der Token-Datei ist genau die Zeile `ODR-E2001 ADMIN_TOKEN_FILE: <Grund>` ohne Pfad und Inhalt und beendet `Hochfahren` vor der Startwartung: kein Listener, kein Service, keine Session, kein `audit.jsonl` | `TestAdminTokenVerstoss`: 7 Gründe (`Datei nicht lesbar`, `ist ein Verzeichnis`, `keine reguläre Datei`, `Gruppen- oder Weltrechte`, `mehr als eine Zeile`, `weniger als 32 Zeichen`, `Datei nicht anlegbar`) × 4 Modi = 28 Zellen; Orakel die Zeile als Literal, Canary im Pfad steht in keiner Zeile; die Haken `Listen`, `Startup` und `Create` der Dienste zählen 0, die Ablage hat keine Session, `audit.jsonl` besteht nicht | der Schritt nach der Startwartung (rot: 28 Zellen, `Startup` zählt 1); der Fehler ignoriert (rot: 28); der Pfad in der Zeile (rot: 28); `ErrUngueltig` statt `TokenFehler` abgebildet (rot) | andere |
| B06 | (`SPEC-300` Nr. 5 d) `admin.Konfig` hat jedes Feld gesetzt: `Listener` der gebundene, `Token`, `IDs`, `Audit` und `Telemetrie` die einen Instanzen, `Routen` aus A01, `Vergleich` nil | `TestAdminKonfig`: Haken `NeuerAdmin` hält die Konfig; Menge der 7 Felder über `reflect.VisibleFields` (Zählprüfung 7) × 4 Modi = 28 Zellen; Orakel Zeigergleichheit mit `h.ids`, `h.audit`, `h.tel` über die Naht, Adresse des Listeners gleich `AdminAdresse()`, `Routen` gleich A01 | je Feld der Nullwert (7 Mutationen, je rot in 4 Modi); ein zweiter Generator (rot: 4); ein zweiter Adapter der Telemetrie (rot: 4) | andere |
| B07 | (`SPEC-300` Nr. 11) Ein Literal mit leerem `AdminListen` hat keinen Admin-Listener: ein Bind (Haken `Listen` zählt 1), keine Token-Datei, `AdminAdresse()` leer, `Herunterfahren` kennt einen Server | `TestAdminOhneListener`: 4 Modi × {Build, Hochfahren, Herunterfahren} = 12 Beobachtungen; kein `admin.token` unter `DATA_DIR` | der leere Wert gilt als Default (rot: der Zähler zählt 2, 4 Modi); das Token wird dennoch geladen (rot: 4) | andere |
| B08 | (`SPEC-300` Nr. 7) Beide Listener binden vor dem ersten `Serve`; ein Bind-Fehler ist `ODR-E2001 LISTEN` bzw. `ADMIN_LISTEN: Adresse nicht bindbar`, beide Zeilen nach Schlüssel sortiert, der gebundene Listener wird geschlossen, kein `Serve` läuft, die Adressen sind leer; bei Erfolg tragen beide Adressen verschiedene Ports | `TestAdminBindung`: Menge M-Bind {beide frei, `LISTEN` belegt, `ADMIN_LISTEN` belegt, beide belegt, beide dieselbe Adresse} × 4 Modi = 20 Zellen und `ctx` vor dem Bind abgebrochen (Kontextfehler unverändert, 4 Zellen); Orakel Zeilen als Literal, der Fake des Listeners zählt `Close`, die Fakes der Server zählen `Serve` | der OData-Server startet vor dem Bind des Admin-Listeners (rot: `Serve` zählt 1 bei „Admin belegt“, 4); nur die erste Zeile gemeldet (rot: 4); die Folge `LISTEN` vor `ADMIN_LISTEN` (rot: 4); der gebundene Listener bleibt offen (rot: 12) | andere |
| B09 | (`SPEC-300` Nr. 8) Herunterfahren: `Anhalten` Admin, `Anhalten` OData, `CloseAll` (nur `record`), `Schliessen` OData, `Schliessen` Admin, Warten auf beide `Serve`, Ablage zu; ein Fehler hält die späteren nicht auf, der `errors.Join` folgt der Folge | `TestAdminHerunterfahren`: das Protokoll der Fakes in einer Folge; Menge 4 Modi × {kein Fehler, Fehler aus `CloseAll`, `Schliessen` OData, `Schliessen` Admin, Ablage} = 20 Zellen (`CloseAll` entfällt außerhalb von `record`: 16 mit Fehlerquelle, 4 ohne); Orakel die Folge und die Folge der Fehler als Literal | `Anhalten` Admin nach `CloseAll` (rot: 5 Zellen von `record`); Admin vor OData geschlossen (rot: 20); ein Fehler des Admin-Schließens hält die Ablage auf (rot); `Anhalten` Admin fehlt (rot: 20) | andere |
| B10 | (`SPEC-300` Nr. 8) `Schliessen` des Admin-Servers erhält `SHUTDOWN_TIMEOUT` minus der Zeit seit dem ersten `Anhalten`, gemessen an der einen Uhr, nicht unter 0: die Dauer ist eine Frist, nicht die Summe | `TestAdminFristRest` wie `TestFristRest`: Dauer von `Schliessen` OData {0, kleiner, gleich, größer als die Frist} × Dauer von `CloseAll` {0, kleiner, gleich} im Modus `record` und {0, kleiner, gleich, größer} in den drei anderen = 12 + 12 = 24 Zellen; die Uhr ist die `FestUhr`, Orakel `rest2` als Literal | `rest2` gleich der Frist (rot); negativ (rot: die Zellen größer als die Frist); die Summe zweier Fristen (rot) | andere |
| B11 | (`SPEC-300` Nr. 8, `SPEC-238` Nr. 11) Der Admin-Server ist vor `CloseAll` angehalten, ein zugelassener Request wird beantwortet, ein späterer ist 503 `ODR-E5010`, auch `GET /healthz` | `TestAdminZulassung`: im Modus `record` hält ein Kanal `CloseAll`, in der Zeit sendet der Test `POST /__admin/sessions` (503, keine Session) und `GET /healthz` (503); ein vor `Anhalten` zugelassener Request einer Test-Route (Haken `NeuerAdmin` hängt sie an) endet mit 200; 4 Modi × {zugelassen, später} = 8 Zellen; je Prozess unter `make test-race` | `Anhalten` Admin nach `CloseAll` (rot in jedem Lauf: 201 und eine Session nach `CloseAll`); zugelassener Request abgebrochen (rot: 4); `GET /healthz` bleibt 200 (rot: 4) | nebenläufig |
| B12 | (`SPEC-300` Nr. 9) `Fehler()` liefert höchstens einen Wert, den ersten Fehler eines der beiden `Serve`, und ist erst geschlossen, wenn beide zurückgekehrt sind, nie zweimal | `TestAdminFehlerKanal`: Menge der Ausgänge {Admin fällt aus, OData fällt aus, beide zugleich, einer mit Fehler und einer mit `nil`, beide `nil` nach `Anhalten`} × 4 Modi = 20 Zellen; die Fakes warten auf Kanäle, 100 Wiederholungen je Zelle; Orakel Zahl der Werte (höchstens 1), Zustand des Kanals (offen, bis beide zurück), kein Senden auf einen geschlossenen Kanal | Kapazität 2 und zwei Werte (rot: Zahl 2); `close` nach dem ersten `Serve` (rot in jedem Lauf: der Kanal ist zu früh zu); Senden nach `close` (rot: Panik) | nebenläufig |
| B13 | (`SPEC-300` Nr. 9, `SPEC-139` Nr. 8) Fällt `Serve` eines Listeners vor dem Signal aus (Fehler oder `nil` ohne `Anhalten`), liefert `Start` 1, beide Listener fahren herunter, und die letzte Zeile ist `ODR-E9001: Internal error` | `TestAdminAusfall`: Menge {Admin, OData} × {Fehler, `nil` ohne Anhalten} × 4 Modi = 16 Zellen; Orakel Exit, die Zeilen (Folge, letzte Zeile), beide Server angehalten | der Ausfall des Admin-Servers ignoriert (rot: 8); die Zeile `ODR-E9001` vor den Zeilen des Herunterfahrens (rot: 16) | andere |
| B14 | (`SPEC-238` Nr. 3 c am Prozess) `GET /healthz` ohne Token ist in jedem Modus 200 mit `{"status":"ok"}` und `X-Correlation-Id`; `/healthz/`, `/HEALTHZ` und `POST /healthz` sind 401 | `TestAdminHealthz`: 3 Pfade/Methoden × 4 Modi = 12 Zellen am Listener des `Handle` | `/healthz` hinter dem Token (rot: 4); in einem Modus nicht bedient (rot: 3); die Schreibweise zählt nicht (rot: 4) | andere |
| B15 | (`SPEC-300` Nr. 12 b, `SPEC-238` Nr. 8) Jede Abweisung mit 401 `ODR-E2101` schreibt genau eine Zeile `admin_abgewiesen` mit den vier Schlüsseln `timestamp`, `correlation_id`, `art`, `error_code` in dieser Folge, die `correlation_id` gleich `X-Correlation-Id` der Antwort; der Zähler `ODR-E2101` der Diagnose steigt um 1 | `TestAdminAbweisung`: Menge der Anlässe {kein Header, falsches Token, Schema `Basic`, zwei Header} × 4 Modi = 16 Zellen; Orakel die geparste Zeile (Schlüssel und Folge als Literal), Zählprüfung Zeilen gleich Abweisungen | ein fünftes Feld (rot: 16); `correlation_id` neu erzeugt (rot: 16); zwei Zeilen je Abweisung (rot); keine Zeile (rot); der Zähler nicht gezählt (rot) | andere |
| B16 | (`SPEC-300` Nr. 12 a) Das Token steht in keinem Byte von `audit.jsonl`, `stdout`, `stderr`, einer anderen Datei unter `DATA_DIR` und einer Antwort (Body und Header), auch nicht als Hälfte | `TestAdminKeinGeheimnis`: Menge Weg {Datei `audit.jsonl`, `stderr`, `stdout`, andere Dateien unter `DATA_DIR`, Antworten} × Anlass {kein `Authorization`, falsches Token, Token bis auf das letzte Byte, erste Hälfte, Token plus ein Byte, gültiges Token auf einer schreibenden Route} × 4 Modi = 5 × 6 × 4 = 120 Zellen; Suchmuster ganz, erste und zweite Hälfte; Kontrolle je Weg: der eingespielte Kontrollwert wird gefunden (5 Zellen) | das Token in der Meldung eines Fehlers von `LadeToken` (rot: Weg `stderr`, 24 Zellen), als `art` des Eintrags (rot: 24), in einem Header der 401-Antwort (rot: 24), im Body (rot: 24), in einer Zeile der Telemetrie (rot: 24) | andere |
| B17 | (`SPEC-300` Nr. 15, `SPEC-201` Nr. 4, 9 e, 11) Das Handbuch trägt `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE`, die zwölf Codes und die Endpunkte; der Präfix `ADMIN_` ist nicht mehr in `nichtGelieferte`; zwölf der dreizehn Ausnahmen entfallen, `ODR-E7012` bleibt mit berichtigtem Grund; `SPEC-201` Nr. 9 e nennt die Klasse nicht mehr | `make katalog-check` (`katalog: ok`) und `make handbuch-check` grün; je Mutation das Rot der benannten Art (Exit 1) | ein Code bleibt Ausnahme und steht im Handbuch (rot: Art 9, je Code); ein Schlüssel ohne Zeile im Handbuch (rot: Art 4); `ADMIN_` bleibt in `nichtGelieferte` und das Handbuch nennt `ADMIN_LISTEN` (rot) | andere |
| B18 | (akzeptierte Negative, `SPEC-300` *Ungeprüft*) (1) Eine Session, die ein Client in einem Modus ohne Recorder anlegt, bleibt über das Herunterfahren `open` und wird beim nächsten Start `incomplete`; (2) die Meldung eines Ausfalls nennt den Listener nicht; (3) ein Verstoß der Token-Datei und einer der Konfiguration erscheinen in zwei Läufen | kein Fänger. Haltepunkt gesucht: (1) `CloseAll` ist ein Aufruf des Dienstes, den der Test zählen könnte, aber die Zusage ist, dass er nicht läuft; (2) die Zeile ist festgelegt (`SPEC-139` Nr. 8); (3) die Reihenfolge ist entschieden (Nr. 4). Grund: (1) kein Nutzen für einen Aufruf von `CloseAll` ohne Recorder, (2) kein zweiter Wortlaut ohne Gewinn, (3) das Laden legt nichts an. Folge: (1) ein Zustand `open` bis zum Neustart (Maß: 1 von 3 Statusarten, nur nach `POST` in einem Modus ohne Recorder), (2) der Betreiber sieht den Ausfall, nicht den Listener, (3) ein zweiter Lauf. Maß: 3 Punkte | nicht anwendbar | andere |
| B19 | (`SPEC-300` Nr. 3, Zwischenstand bis Teil C) Solange die drei Schlüssel von `SPEC-299` noch abgelehnt sind, ruft der Root die Prüfung am Laden und in `Build` mit leerem Paar und falschem Flag: ein `ADMIN_LISTEN` außerhalb von Loopback ist `ODR-E2001 ADMIN_LISTEN` (der Zusammenhang, `SPEC-299` Nr. 6), und der Listener bindet nie an eine fremde Adresse ohne Schutz | `TestAdminLoopbackZwischenstand`: 5 Hosts außerhalb von Loopback (`0.0.0.0:8081`, `[::]:8081`, `:8081`, `recorder.internal:8081`, `10.0.0.5:8081`) × Stelle {Laden, `Build`} × 4 Modi = 40 Zellen; Orakel die eine Zeile als Literal, kein Bind (Haken `Listen` zählt 0), `DATA_DIR` nicht angelegt (Laden) bzw. keine Datenbank (`Build`) | der Aufruf entfällt (rot: 40, der Listener bindet `0.0.0.0`); der Aufruf nur am Laden (rot: 20, `Build` mit Literal); nur im Modus `record` (rot: 30) | andere |

Die Zeilen B01 bis B19 sind 19: *a* = 2 (B11, B12), *b* = 17. **Mengen mit Achsen** (Orakel als Literal, nie aus dem Code und nie als Aufruf des Codes; die Zahl der Zellen steht im Test als Literal und wird gegen die gezählten verglichen, `AGENTS.md` §3.15): M-Modus (`record`, `replay`, `mock`, `hybrid`, 4); M-Bind (5 Ausgänge, Achsen Seite, Zahl der belegten, gleiche Adresse); M-Abweisung (Anlass × Weg der Ausgabe). **Unbeschränkte Menge und ihre Prüfgrenze (akzeptiertes Negativ):** Schlüssel mit dem Präfix `ADMIN_` (B03): jeder Name; Prüfgrenze: die 8 benannten Namen und 3 Schreibweisen; Grund: die Regel ist ein Präfixvergleich, ein Name, den sie anders behandelte, zeigt sich am Rand (`ADMIN_` allein, `ADMIN_X`); Folge: eine Ausnahme der Regel für einen Namen im Abstand von mindestens 2 Zeichen zu den acht bliebe grün; Maß: 8 von unbeschränkt, 3 Arten (Präfix allein, Suffix, Schlüssel mit anderem Rest). Die Mengen M-Route, M-Einstellung (Teil A), M-Flag, M-Weg (Teil C) und M-Art (Teil D) stehen in den Plänen der Teile, die sie führen.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18). Das gilt auch für die Beispiele des Handbuch-Kapitels.
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-bootstrap-m6-endpunkte](../next/slice-bootstrap-m6-endpunkte.md) (Teil A des Schnitts) liegt in `done/`; mit ihm liegen die Bedingungen von Teil A in `done/` (die Teile von `slice-admin-listener-token`, `slice-admin-tls`, `slice-inspection-use-case`, `slice-inspection-coverage-diagnose`, `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces` je mit ihren Teilen, `slice-test-security`, `slice-bootstrap-m2`, `slice-m2-referenzdienst`, `slice-bootstrap-m3`, `slice-audit-log`, `slice-telemetry-port`, `slice-telemetry-einbau-record` und beide Teile des Hybrid-Schnitts), dazu `slice-handbuch-meldungskatalog` für das Handbuch-Kapitel. Die Spezifikation trägt den Nachzug von `SPEC-110` und den Vertrag `SPEC-300`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12; geliefert am 2026-11-23). Teil C startet erst, wenn dieser Teil in `done/` liegt.

**Warum:** der Teil öffnet den zweiten Listener und liest die Token-Datei; er braucht die Routen und Dienste von Teil A (*Code-Abhängigkeit*: B06 prüft `Routen` gleich A01) und die Ports `AuditLog` und `Telemetry` (*Code-Abhängigkeit*). Teil B und Teil A ändern dieselben Dateien in `src/bootstrap/` (`wurzel.go`, `export_test.go`) und laufen nacheinander.

**Größenmarke** (`SPEC-128` Nr. 1 und 10 g, `AGENTS.md` §3.14; Teil des Architect 2026-11-23 im Block unten; der Planner hat die Zahlen am 2026-11-23 aus der Tabelle nachgerechnet, **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 19 Zeilen der Tabelle (B01 bis B19), *a* = 2 (B11, B12), *b* = 17, *f* = 2 (der Fake des Admin-Servers samt Haken `NeuerAdmin` und der Client mit Token), *p* = 260 (zwei Schlüssel, Token-Schritt, Bindung beider Listener, Herunterfahren, Kanal): *S* = 85·2 + 80·17 + 200·2 + 260 = 170 + 1360 + 400 + 260 = 2190, 135·2190 = 295650, **Marke 3000** = 100 · ⌈29,565⌉; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`. **Der Abstand zur Schnittgrenze beträgt nur 33:** eine weitere Zeile der Tabelle (80) lässt *S* = 2270 und löst die Schnittfrage aus; wächst die Tabelle im Code (`AGENTS.md` §3.9) oder beim Start, prüft der Planner sie neu und schneidet dann vor dem Code B19 (Zwischenstand) oder B17 (Handbuch und Katalog) als eigenen Teil (Architect, 2026-11-23). Keine Abweichung vom Vorschlag des Architect. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen.

```groessenmarke
pfad: src/bootstrap/konfig.go
pfad: src/bootstrap/konfig_admin.go
pfad: src/bootstrap/konfig_werte.go
pfad: src/bootstrap/lebenszyklus.go
pfad: src/bootstrap/start.go
pfad: src/bootstrap/wurzel.go
pfad: src/bootstrap/export_test.go
pfad: src/bootstrap/haken_test.go
pfad: src/bootstrap/lebenszyklus_test.go
pfad: src/bootstrap/konfig_test.go
pfad: src/bootstrap/admin_listen_test.go
pfad: src/bootstrap/admin_token_test.go
pfad: src/bootstrap/admin_start_test.go
pfad: src/bootstrap/admin_herunterfahren_test.go
pfad: src/bootstrap/admin_geheimnis_test.go
pfad: tests/e2e/stufe_prozess_test.go
pfad: harness/katalog-ausnahmen.txt
produktcode: 260
fakes: 2
```

Der Block nennt die Dateien einzeln (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`); die Dateien des Handbuchs und der Spezifikation stehen nicht im Block (`SPEC-128` Nr. 3 m). Die neuen Dateien (`konfig_admin.go`, die Tests `admin_*_test.go`) gibt es noch nicht (die Namen sind ein Vorschlag des Architect; die Messung zählt dort 0, bis sie entstehen); ein anderer Name ersetzt den Vorschlag im selben Commit, und der Implementer nennt jede weitere Datei, die er ändert oder anlegt, im selben Commit im Block. `konfig_admin.go` führt Teil C ebenfalls im Block (nach diesem Teil, kein gleichzeitiger Zweig).

**Kollisionen** (Architect, 2026-11-23, gelesen an den Zweigen und den Plänen, vom Planner beim Schnitt übernommen; `SPEC-128` Nr. 10 g): (1) Teil A ([slice-bootstrap-m6-endpunkte](../next/slice-bootstrap-m6-endpunkte.md)) ändert `wurzel.go` und `export_test.go`: er liegt vor diesem Teil in `done/` (Start-Bedingung), seine Zeilen gehören zur Basis. (2) Teil C und Teil D laufen nach diesem Teil. (3) [slice-admin-workspaces-audit](../done/slice-admin-workspaces-audit.md) und [slice-test-security](../done/slice-test-security.md) liegen vor Teil A in `done/` und ändern keine Datei dieses Blocks. (4) Die Gate-Kette (Nr. 7 der Übergangsregeln): kein Pfad aus `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`; `harness/katalog-ausnahmen.txt` gehört zum Kataloge-Gate und nicht zu dieser Liste; der Planner prüft es beim Start.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); der Schnitt von B19 oder B17 ist vorab benannt (Absatz *Größenmarke*), aber nur vor dem Code zu nehmen.
- `in-progress` → `open` (blockiert — Carveout?): der Rundlauf lässt sich nicht ohne Umschalten des Modus zur Laufzeit bilden (Teil D, Auftrag 2), oder ein Slice der Welle liefert nicht den Handle, den der Test braucht (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23; die Zeilen dieses Teils; die Zeilen der Teile C und D stehen in deren Plänen; jede Zelle nennt ihren Ort in `SPEC-300`, den Verträgen davor oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Lesen von `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE` | Default: `127.0.0.1:8081` bzw. `<DATA_DIR>/admin.token` (`SPEC-300` Nr. 2) | wie fehlend: leer und nur Leerraum gelten als ungesetzt (`SPEC-110` Nr. 1, `SPEC-300` Nr. 2) | ungültige Form von `ADMIN_LISTEN`: Verstoß, Grund der Form von `LISTEN`, Port 0 ungültig (`SPEC-110` Nr. 7); `ADMIN_TOKEN_FILE` hat keine ungültige Form (`SPEC-300` Nr. 2, Nr. 4); jeder andere Name `ADMIN_*`: „nicht unterstützt“ (`SPEC-110` Nr. 2, `SPEC-300` Nr. 2) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | dasselbe Ergebnis (`SPEC-110` Nr. 13); das Laden liest und legt die Token-Datei nie an (`SPEC-110` Nr. 11) |
| Start der beiden Listener (Token-Datei zuerst, beide Bindungen vor dem ersten `Serve`) | `AdminListen` leer im Literal: kein Admin-Listener (`SPEC-300` Nr. 11); Token-Datei fehlt: wird erzeugt (`SPEC-239` Nr. 3, `SPEC-300` Nr. 4) | Token-Datei leer oder nur Leerraum: `weniger als 32 Zeichen` (`SPEC-239` Nr. 1 g, `SPEC-300` Nr. 4) | Token-Datei mit Verstoß: die Zeile `ODR-E2001 ADMIN_TOKEN_FILE: <Grund>` vor der Startwartung (`SPEC-300` Nr. 4); Bind belegt, auch `LISTEN` gleich `ADMIN_LISTEN`: `ODR-E2001 <Schlüssel>: Adresse nicht bindbar`, beide sortiert (`SPEC-300` Nr. 7) | `ctx` abgebrochen vor oder in einem Schritt: der Kontextfehler, nichts Weiteres geöffnet (`SPEC-139` Nr. 6, `SPEC-300` Nr. 7); Signal (`SPEC-138` Nr. 7, 8) | ein zweites `Hochfahren` ist `model.ErrUngueltig` (`SPEC-139` Nr. 6); ein zweiter Start auf demselben `DATA_DIR` lädt dasselbe Token (`SPEC-300` Nr. 4, 16) |
| Herunterfahren der beiden Listener | ohne `Hochfahren` schließt nur die Ablage (`SPEC-139` Nr. 7); ohne Admin-Listener wie bisher (`SPEC-300` Nr. 11) | gilt nicht: `SHUTDOWN_TIMEOUT` ist mindestens 1 ms (`SPEC-110` Nr. 4); der Rest der Frist für den Admin-Server ist nie unter 0 (`SPEC-300` Nr. 8) | ein Fehler eines Schritts hält die späteren nicht auf, die Fehler stehen im `errors.Join` in der Folge der Schritte (`SPEC-139` Nr. 7, `SPEC-300` Nr. 8) | `context.WithoutCancel` (`SPEC-138` Nr. 8 b, `SPEC-139` Nr. 7); ein zugelassener Request des Admin-Servers wird bis `rest2` bedient, ein späterer ist 503 `ODR-E5010` (`SPEC-238` Nr. 11, `SPEC-300` Nr. 8) | einmal, jeder weitere und jeder gleichzeitige Aufruf liefert dasselbe Ergebnis (`SPEC-139` Nr. 7) |

**Zusatztabelle** (`AGENTS.md` §3.16: die Operationen melden mehrere Verstöße, und Zusagen hängen vom Modus ab; je Paar von Verstößen am selben Schlüssel und je Modus eine Zeile; die Paare der TLS-Schlüssel stehen in Teil C):

| Paar oder Modus | Ergebnis | Fänger |
|---|---|---|
| `LISTEN` und `ADMIN_LISTEN` beide nicht bindbar | zwei Zeilen, `ADMIN_LISTEN` vor `LISTEN`, kein `Serve` | B08 |
| Verstoß der Token-Datei und Bind-Fehler | allein die Zeile der Token-Datei, der Bind läuft nicht (`SPEC-300` Nr. 4) | B05 |
| Verstoß der Token-Datei und Verstoß der Konfiguration | zwei Läufe, nie beide in einer Meldung (`SPEC-300` Nr. 4, *Folge im Betrieb*) | B18 (akzeptiertes Negativ) |
| jeder Modus (`record`, `replay`, `mock`, `hybrid`) | derselbe Weg, dieselben Verstöße; Routen der Imports nur in `record` (`SPEC-300` Nr. 3, 5 c) | alle Mengen mit der Achse Modus; A01 (Teil A) |

- Der Admin-Listener bindet in diesem Stand nur auf Loopback, obwohl `SPEC-110` Nr. 2 die drei Schlüssel von `SPEC-299` als geprüft führt (Zwischenstand, Architect, 2026-11-23): bis Teil C bleiben `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT` „nicht unterstützt“ — **Ausgang:** entschieden; der Plan nennt den Zwischenstand in §1, das Handbuch von Teil B nennt die drei Schlüssel nicht, Zeile B19 fängt den Listener an fremder Adresse; Teil C hebt den Zwischenstand auf.
- `SPEC-110` Nr. 7 schließt Port 0 aus; ein Test auf demselben Host braucht einen freien Port für `ADMIN_LISTEN` und `LISTEN` (Auftrag 1) — **Ausgang:** entschieden (Architect, 2026-11-23): `Build` mit einem Literal der Konfiguration und Port `0` für beide, die Adressen aus `Adresse()` und `AdminAdresse()`; der Prozess-Test wählt vorher einen Port (`SPEC-110` Nr. 7, `SPEC-300` Nr. 1).
- Die Reihenfolge von Token-Datei, Startwartung und Listenern und das Herunterfahren beider Listener (Auftrag 3): ein Fehler der Token-Datei öffnet keinen Listener — **Ausgang:** entschieden (Architect, 2026-11-23): Token-Datei zuerst, beide Bindungen vor dem ersten `Serve`, `Anhalten` des Admin-Servers vor `CloseAll`, `Schliessen` OData, dann Admin; `SPEC-300` Nr. 4, 7, 8, Zeilen B05, B08 bis B11.
- Der Rundlauf des M6-Tests unter dem Detektor (Auftrag 5) — **Ausgang:** entschieden (Architect, 2026-11-23): die Schritte laufen nacheinander auf der Goroutine des Tests, `close` wartet auf die Exchanges, kein Schlaf; Kanäle in den Zeilen B11 und B12 (`nebenläufig`, je Prozess unter dem Detektor und mit Wiederholung); `SPEC-300` Nr. 14; der Test selbst liegt in Teil D.
- Katalog: zwölf Ausnahmen, nicht eine (Architect, 2026-11-23; die Sendung von `slice-admin-listener-token` nannte allein `ODR-E2102`) — **Ausgang:** entschieden: zwölf der dreizehn Zeilen von `harness/katalog-ausnahmen.txt` entfallen mit diesem Teil, das Handbuch trägt die Codes (`SPEC-300` Nr. 15, Zeile B17); `ODR-E7012` bleibt.
- Handbuch, Katalog und Gate (Architect, 2026-11-23): die Texte schreibt der Teil, die Stellen sind `SPEC-200` (Beispiele als Datei gegen das Binary), `SPEC-201` Nr. 4 (Schlüssel), Nr. 9 e (der Satz der Klasse Admin-API unter den Ausnahmen) und Nr. 11 (Folge für Folgewellen), in `docs/user/benutzerhandbuch.md` die Tabellen `Einstellungen` und der Fehlercodes und ein Kapitel zum Admin-Listener (Token und Token-Datei, Endpunkte); in `src/bootstrap/konfig_werte.go` der Präfix `ADMIN_` in `nichtGelieferte` — **Ausgang:** entschieden; Träger: Zeile B17 (der Abschnitt im Kapitel `TLS` und die drei Schlüssel: Zeile C08, Teil C).
- Stufennummern im Treiber `tools/harness/test-integration.sh` und der Fall `ADMIN_LISTEN` von `TestE2EProzess` — **Ausgang:** entschieden (Architect, 2026-11-23): der Fall trägt `0.0.0.0:<Port>` und erwartet im Zwischenstand `ODR-E2001 ADMIN_LISTEN` (Exit 1); die Stufe 11 `TestE2EAdmin` schreibt Teil D.
- Größe: die Tabelle *Zusage · Fänger* liegt vor, die Schätzung des ungeschnittenen Plans (4366) liegt über der Schnittgrenze; dieser Teil liegt mit *S* = 2190 nur 33 unter ihr — **Ausgang:** entschieden (Planner, 2026-11-23, nach dem Vorschlag des Architect): Schnitt in vier Teile mit den Marken 1100, 3000 (dieser Teil), 1000 und 1000, nachgerechnet aus der Tabelle (keine Abweichung); der zweite Schnitt von B (B19 oder B17) ist vorab benannt (§4) und nur vor dem Code zu nehmen; `basis`, `schaetzung` und `marke` setzt der Planner beim Start.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-23; der Planner zieht Treffer und Zähler beim Start nach). Offen und für den Teil einschlägig: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (10×, B11 und B12 unter Last), `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×, die Kanäle der Zeilen B11 und B12) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (22×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (43×, `AGENTS.md` §3.9; der Nachzug von `SPEC-110`, das Handbuch und die Welle-Datei sind Stellen), `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16) und `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (`AGENTS.md` §3.18, für die Beispiele des Handbuchs).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
