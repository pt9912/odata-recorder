# Slice slice-admin-workspaces-audit: Admin-Endpunkte für Workspaces und Snapshots, Teil 2: Einträge der vier Arten, Warten und Abbruch, Nebenläufigkeit (Teil 2 des Schnitts von `slice-admin-workspaces`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-09`, `LH-FA-10`, `LH-FA-13`, `LH-RB-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-298` Nr. 9 (Dauer, Kontext, Abbruch), Nr. 11 (Audit: Anlegen, Zurücksetzen, Erzeugen, Wiederherstellen) und Nr. 12 (Nebenläufigkeit; Zustand und Wiederholung trägt Teil 1) · `SPEC-234` Nr. 2, 5 und 7 (die neunte bis zwölfte Art) · `SPEC-235` Nr. 3, 4 d · `SPEC-238` Nr. 9, 15 b und c (Rahmen, nicht geändert) · `SPEC-129` Nr. 9, 11, 14, 19 und `SPEC-260` Nr. 25 (Ordnung und Abbruch bei den Diensten) · `SPEC-005` (Zeilen `POST /__admin/workspaces`, `reset`, `POST …/snapshots`, `restore`: Audit-Logging) · `LH-FA-13.a` (Änderungen sind auditierbar) · `ARC-007` · `architecture.md` §2.2, §4.5.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md), Vorschlag des Architect vom 2026-11-22). **Datum:** 2026-11-22.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem** Abschnitt des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Admin-Adapter schreibt nach jedem gelungenen `POST /__admin/workspaces` (201), `reset` (200), `POST …/snapshots` (201) und `restore` (200) genau einen Eintrag seiner Art (`workspace_angelegt`, `workspace_zurueckgesetzt`, `snapshot_erzeugt`, `workspace_wiederhergestellt`; neue neunte bis zwölfte Art von `support.Arten()`; `service_id` und `workspace_id` aus dem Ergebnis des Dienstes, Korrelations-ID der Antwort), nie bei einem Fehler und nie bei `GET` oder `DELETE`; ein Fehler von `Append` ändert die Antwort nicht. Der Teil hält die Zusagen über Warten und Abbruch (die Antwort entsteht erst nach der Rückkehr der Operation, keine Frist, der `ctx` bleibt der des Requests, ein Kontextfehler bricht ab ohne Eintrag, ein Erfolg bei beendetem `ctx` ist Antwort und Eintrag) und über die Nebenläufigkeit (gleichzeitige Requests beeinflussen einander nicht, keine Sperre über einem Aufruf von `workspaces`, `snapshots` oder `audit`). Er hebt die Zählstelle der Arten von `support.Arten()` von 8 auf 12 (die achte Art liefert `slice-admin-imports-audit`).

**Vertrag (Architect, 2026-11-22):** `SPEC-298` Nr. 9, 11 und 12; der Vertrag steht auf `main`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice war ungeschnitten über der Schnittgrenze von `SPEC-128` Nr. 1 (Schätzung 2780 gegen 2223); der Architect schlug in §3 von [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md) (*Größe und Schnitt*) zwei Teile vor, der Planner hat sie am 2026-11-22 vor dem Code vollzogen: [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md) trägt `WorkspaceRouten`, die drei Bodys, Antworten, JSON, Fehlerabbildung und Authentisierung (Zeilen WS1 bis WS16), dieser Teil die Einträge, das Warten und den Abbruch und die Nebenläufigkeit (WS17 bis WS23, WN1, WN2).

**Übernimmt (Sendung von `slice-admin-workspaces`, Teil 1; Geber: `slice-admin-workspaces`; Eintrag des Planners vom 2026-11-22 beim Schnitt, `AGENTS.md` §3.13; §1 *Ausdrücklich NICHT* und DoD dieses Teils vorher geschrieben und gelesen, kein Ausschluss trifft eine Sendung):** (a) die Mittel aus Teil 1, die dieser Teil erweitert: der Konstruktor `WorkspaceRouten` (hält `audit` seit Teil 1, ruft es nicht; dieser Teil ruft `Append` in den Handlern der vier Routen mit Eintrag), diese vier Handler (`POST /__admin/workspaces`, `reset`, `POST …/snapshots`, `restore`; die vier Handler ohne Eintrag ändert dieser Teil nicht, außer um einen Fund von WS17, WS18, WN1 oder WN2 an ihnen zu beheben: das akzeptierte Negativ (5) von Teil 1, Geber `slice-admin-workspaces`; §1 *Ausdrücklich NICHT* und DoD dieses Teils vorher gelesen, kein Ausschluss trifft es, WS17, WS18, WN1 und WN2 führen alle acht Routen), `workspaceFake` und `snapshotFake` (Teil 1 liefert Zählen, Festhalten jedes Arguments und einstellbare Ausgänge; dieser Teil ergänzt die Haltepunkte `vor` und `nach` je Operation und den Zustand mit eigener Sperre für WN1), der zählende Reader des Bodys (Teil 1 liefert Zählen und den einstellbaren Fehler; dieser Teil ergänzt den Treffpunkt am Ende für WN1) und der Haltepunkt `auditNach` aus den Tests der Sessions; (b) die Mengen der ungültigen Bodys (WS7 und WS8: 145 Fälle ohne den Rand) und der Fehler (M-Fehler, WS13: 125 Zellen), die WS20 mit demselben Orakel liest: Teil 1 legt sie als Funktionen im Testpaket an (nicht als Literal in einem Test), dieser Teil ruft sie; (c) die Zellen von WS3 (`FakeAudit.Aufrufe()` gleich 0, je Route eine Zelle): dieser Teil hebt die vier Zellen der Routen mit Eintrag auf 1, im selben Commit wie der Code (WS19); die 125 Zellen von WS13 sind Fehlerausgänge, ihr Orakel bleibt 0 (Teil 2 liest sie in WS20); (d) der Rest von WS15: `git diff --stat` über `src/bootstrap` ist in diesem Teil allein `audit_betrieb_test.go` (Teil 1 prüft ihn leer); er steht als benannte Ausgabe in WS22.

**Übernimmt (Sendung von `slice-admin-listener-token`, Pflichten der Endpunkte nach `SPEC-238` Nr. 15, Schnitt vom 2026-11-07, `AGENTS.md` §3.13; die Hälfte (c) dieses Teils, die übrigen Hälften trägt Teil 1):** (b) die Route ruft `Audit.Append` nie für `ODR-E2101` und schreibt kein Fehlerobjekt selbst: mit dem Aufruf von `Append` in diesem Teil bekommt der AST-Test WS15 aus Teil 1 sein Gegenstück, die Zählstelle `TestAppendEineAufrufstelle`; (c) die vier Arten von `SPEC-234` Nr. 2 schreibt der Dienst, nie die Route; die Arten für Anlegen, Reset, Snapshot und Restore sind Zeilen im Vertrag des Endpunkts (`SPEC-298` Nr. 11): sie liefert dieser Teil (WS19 bis WS22). Der Reviewer löst jede Zeile von Nr. 15 mit einer Mutation auf.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `WorkspaceRouten` (Konstruktor, Zuordnung), die drei Bodys, Pfadwert, Antworten, JSON, Fehlerabbildung und Authentisierung der Routen — [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md) (Teil 1); dieser Teil erweitert die vier Handler mit Eintrag um `Append` und ändert die Dateien von Teil 1 nur, wo §3 es nennt.
- Die Logik von Anlage, Reset, Löschen, Snapshot und Restore — `WorkspaceUseCase` und `SnapshotUseCase` (`done/`); der Adapter bildet nur HTTP ab und schreibt nie direkt in die Datenbank (Architektur §2).
- Der Audit-Eintrag der Löschung eines Workspace (Art `workspace_geloescht`) — `WorkspaceUseCase.Delete` schreibt ihn (`slice-audit-dienste`); der Adapter schreibt die vier Arten von `SPEC-234` Nr. 2 nie, auch nicht für denselben `DELETE` (Entscheidung `SPEC-234` Nr. 7, Geber: `slice-audit-log` und `slice-audit-dienste`). Dieser Teil schreibt allein die Einträge seiner vier Arten; der Eintrag zu `ODR-E2101` gehört dem Rahmen (`slice-admin-abweisen-audit`).
- Die Zellen Modus × Art für die vier Arten im Test `src/bootstrap/audit_betrieb_test.go` — [slice-bootstrap-m6](slice-bootstrap-m6.md): das Binary ruft `admin` erst mit diesem Slice; dieser Teil nimmt die Arten wie `admin_abgewiesen` aus dem Test aus und ändert dort die Zählstelle (8 → 12).
- Der Port `AuditLog` — `welle-betrieb` (`slice-audit-log`); der Teil ruft ihn.
- Ein Eintrag bei einem fehlgeschlagenen Aufruf, ein Feld `snapshot_id` oder `name` im Eintrag, ein Eintrag der Route für `DELETE` und eine Drosselung — `SPEC-298` Nr. 11 (*Entscheidung*, Alternativen 2, 3 und 5) und `SPEC-234` Nr. 2 legen es anders fest; eine Änderung ist eine Änderung des Vertrags.
- Ein Handbuch-Kapitel — das Binary bleibt unverändert (`SPEC-298` Nr. 14: kein Fehlercode, Schlüssel oder Kommando, kein Aufruf von `WorkspaceRouten` aus `bootstrap`); das Kapitel zu `audit.jsonl` führt [slice-admin-sessions-beenden](../done/slice-admin-sessions-beenden.md) in §6 als offenen Punkt (Adresse User) und ist nicht Gegenstand dieses Teils.

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

- [ ] Audit: `POST /__admin/workspaces`, `reset`, `POST …/snapshots` und `restore` schreiben nach einem Erfolg (201 und 200) genau einen Eintrag ihrer Art (`workspace_angelegt`, `workspace_zurueckgesetzt`, `snapshot_erzeugt`, `workspace_wiederhergestellt`) nach der Rückkehr der Operation, vor der Rückkehr des Handlers, mit `ServiceID` und `WorkspaceID` aus dem Ergebnis des Dienstes und der Korrelations-ID der Antwort; kein Fehler, kein Verstoß vor dem Aufruf, kein Kontextfehler und kein `GET` oder `DELETE` schreibt einen (`DELETE` schreibt der Dienst); ein Fehler von `Append` ändert die Antwort nicht; die vier Arten sind die neunte bis zwölfte Zeile von `support.Arten()` (`service_id` und `workspace_id` Pflicht, jedes andere Feld verboten) und die Zählstellen der Tests ziehen mit (8 → 12), die Schleifen von `TestKennungsform` und `TestLogBeiFehler` führen sie, `TestAppendEineAufrufstelle` zählt die Stellen in `workspaces*.go` (WS19 bis WS23).
- [ ] Warten und Abbruch: die Antwort jeder der acht Routen entsteht erst nach der Rückkehr der Operation, ohne Frist und ohne Ablösung des Kontexts; ein Kontextfehler beim Aufruf des Dienstes bricht ab, ohne Fehlerobjekt, Zählen und Eintrag (der Kontextfehler beim Lesen des Bodys liegt in Teil 1, WS13); ein Erfolg bei beendetem `ctx` ist Antwort und, bei den vier Routen mit Eintrag, Eintrag (WS17, WS18).
- [ ] Nebenläufigkeit: gleichzeitige Requests beeinflussen einander nicht, keine Sperre über einem Aufruf von `workspaces`, `snapshots` oder `audit`; zwei gleichzeitige `reset` desselben Workspace: beide 200 mit `n+1` und `n+2` und zwei Einträge, zwei gleichzeitige `POST …/snapshots` mit demselben Namen: 201 mit Eintrag und `ODR-E7006` ohne (WN1, WN2). Pflichten (b) und (c) von `SPEC-238` Nr. 15 für den Eintrag (Geber `slice-admin-listener-token`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (`SPEC-298` Nr. 14: kein Code, Schlüssel oder Kommando; der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/adapters/driving/admin/workspaces.go` | update | die Handler der vier Routen mit Eintrag (`POST /__admin/workspaces`, `reset`, `POST …/snapshots`, `restore`) rufen `Append` nach dem Erfolg (kein neuer Dateiname); keine Änderung an `workspaces_body.go`, `workspaces_json.go`, `admin.go`, `route.go`, `fehler.go`, `lebenszyklus.go`, `sessions*.go`, `inspektion*.go`, `imports*.go` |
| `src/hexagon/ports/support/audit.go`, `src/hexagon/ports/support/audit_test.go` | update | die vier Arten (`ArtWorkspaceAngelegt`, `ArtWorkspaceZurueckgesetzt`, `ArtSnapshotErzeugt`, `ArtWorkspaceWiederhergestellt`), neunte bis zwölfte Zeile von `Arten()` (`SPEC-298` Nr. 11); der Test zählt 8 Arten und wächst auf 12 |
| `src/adapters/driven/audit/audit_test.go`, `src/adapters/driven/audit/log_sperre_test.go`, `src/bootstrap/audit_betrieb_test.go` | update (nur Tests) | Zählstellen der Arten (8 → 12) und Zeilen der vier Arten; die Schleifen von `TestKennungsform` und `TestLogBeiFehler` laufen über `support.Arten()` und führen die vier Arten mit; `audit_betrieb_test.go` nimmt die Arten wie `admin_abgewiesen` aus (das Binary ruft `admin` nicht), sonst bleibt `bootstrap` unverändert |
| `src/adapters/driving/admin/abweisen_audit_test.go` | update (nur Test) | `TestAppendEineAufrufstelle` zählt `Append` je Datei; die Einträge nach dem Erfolg sind vier Stellen in `workspaces.go` |
| `src/adapters/driving/admin/workspaces_fake_test.go`, `src/adapters/driving/admin/workspaces_routen_test.go` | update (nur Tests) | aus Teil 1: `workspaceFake` und `snapshotFake` wachsen um die Haltepunkte `vor` und `nach` je Operation und den Zustand mit Sperre, der Reader des Bodys um den Treffpunkt am Ende (Sendung (a)); die Zellen von WS3 heben die vier Erfolgszellen der Routen mit Eintrag auf 1 (Sendung (c)); ändert ein weiterer Test von Teil 1 sein Orakel (etwa die Kontrolle mit gültigem Token in WS14), nennt der Implementer die Datei im selben Commit im Block (§4) |
| `spec/spezifikation.md` | geliefert (Architect, auf `main`) | `SPEC-298`; Nachzug `SPEC-005`, `SPEC-234` Nr. 2 und 7 und die Historie; dieser Slice schreibt sie nicht |
| Tests | neu | Adapter gegen `workspaceFake` und `snapshotFake` (aus Teil 1, hier erweitert) und `FakeAudit`: jedes Argument festgehalten, jeder Aufruf gezählt, auch die Operation, die die Route nie rufen darf (`AGENTS.md` §3.15); ein Fall je Aufrufstelle für den `ctx` des Requests und für `Append`. Neue Dateien: `workspaces_warten_test.go` (WS17, WS18), `workspaces_audit_test.go` (WS19 bis WS22: Eintrag, Ausgänge, Fehler von `Append`, Arten), `workspaces_nebenlaeufig_test.go` (WN1, WN2) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem 2026-11-22 (Architect) mit dem Vertrag in der Spezifikation (`SPEC-298`) vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. `Nr.` ist die Kennung des ungeschnittenen Plans: **WS17 bis WS23** gehören zu diesem Teil, **WN1** und **WN2** sind seine nebenläufigen Zeilen; WS1 bis WS16 stehen in Teil 1 ([slice-admin-workspaces](../in-progress/slice-admin-workspaces.md)). Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest; die Zeilen sind vom Planner beim Schnitt im Wortlaut übernommen, ohne Änderung (die Zeilen sind byte-gleich dem ungeschnittenen Plan des Architect). Der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein (`AGENTS.md` §3.10).

**Entscheidungen des Architect für diesen Teil** (2026-11-22, vor dem ersten Code-Commit, `AGENTS.md` §3.12; Ort in der Spezifikation; die Aufträge 1 bis 3 stehen in §3 von [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md), die Nummerierung bleibt die des ungeschnittenen Plans):

4. **Abbruch** (`SPEC-298` Nr. 9): der Kontext bleibt der des Requests (keine Ablösung, keine Frist); ein Kontextfehler bricht die Verbindung ab, ohne Eintrag; die Operationen der Dienste lassen bei einem Kontextfehler nie eine Änderung (kein Kontextfehler mit Wirkung); ein Erfolg bei beendetem `ctx` ist Antwort und Eintrag, die niemand liest; die Dauer ist unbegrenzt (kein `WriteTimeout`); die Wirkung bei gleichzeitiger Mutation im Modus `mock` gehört den Diensten.
5. **Audit** (`SPEC-298` Nr. 11, `SPEC-234` Nr. 2 und 7): vier neue Arten (`workspace_angelegt`, `workspace_zurueckgesetzt`, `snapshot_erzeugt`, `workspace_wiederhergestellt`; `service_id` und `workspace_id` Pflicht, sonst verboten), geschrieben von der Route nach jedem Erfolg der vier `POST`-Routen mit Eintrag, nie bei einem Fehler; `DELETE` schreibt die Route nie (der Dienst schreibt `workspace_geloescht`, kein doppelter Eintrag); ein Fehler von `Append` ändert nichts; die Menge der Felder bleibt geschlossen.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-22, Zeilen WS17 bis WS23, WN1 und WN2 des ungeschnittenen Plans; jede Zeile trägt ihre Menge):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| WS17 | (Nr. 9 a, b) Die Antwort entsteht erst nach der Rückkehr der Operation: bis dahin ist nichts an den Client geschrieben, und die Route setzt keine Frist | Wirkungs-Fänger (deterministisch, Kanäle): die Operation blockiert, der Test sieht, dass `ServeHTTP` nicht zurückgekehrt ist und der Recorder leer ist; nach der Freigabe der Erfolgsstatus der Route mit derselben Antwort wie ohne Wartezeit; `ctx.Deadline()` im Fake nicht gesetzt; 8 Routen = 8 Zellen; das fehlende `WriteTimeout` fängt `grenzen_test.go` des Rahmens | die Operation in einer Goroutine und Antwort 202 sofort; `context.WithTimeout(ctx, 30*time.Second)`; `WriteHeader` vor dem Aufruf (die Route schreibt nicht selbst, der Haltepunkt sieht den Kopf trotzdem) | andere |
| WS18 | (Nr. 9 c, d; `SPEC-238` Nr. 9) Endet der `ctx`, läuft der Aufruf mit dem `ctx` des Requests unverändert; ein Kontextfehler beim Aufruf des Dienstes bricht die Verbindung ab, ohne Fehlerobjekt, Zählen und Eintrag (beim Lesen des Bodys: WS13); ein Erfolg bei beendetem `ctx` ist Antwort und, bei den vier Routen mit Eintrag, Eintrag | M-Kontext: die vier Routen mit Eintrag (`POST /__admin/workspaces`, `reset`, `POST …/snapshots`, `restore`) × 4 Enden des `ctx` (`cancel` vor dem Aufruf, abgelaufene Frist vor dem Aufruf, im Aufruf, nach der Rückkehr der Operation vor `Append`) × 5 Ausgänge (Erfolg, `Canceled`, `DeadlineExceeded`, je allein und mit einem `*Fehler E7001` im `errors.Join`) = 80; die vier Routen ohne Eintrag (`GET` Liste, `GET {id}`, `DELETE`, `GET …/snapshots`) × 3 Enden (vor dem Aufruf, im Aufruf, nach der Rückkehr) × 5 Ausgänge = 60; zusammen 140; Orakel `SPEC-298` Nr. 9 c, d: Abbruch oder Antwort, Eintrag nur beim Erfolg der vier Routen; der Fake sieht `ctx.Done()` nach `cancel()` (kein `WithoutCancel`) und ein Wert im `ctx` kommt an; Zählprüfung 80 + 60 = 140 | `context.WithoutCancel(ctx)`; Eintrag auch bei Kontextfehler; kein Eintrag bei beendetem `ctx` nach dem Erfolg; Abbruch bei beendetem `ctx` nach dem Erfolg (je Route); `ctx.Err()` vor dem Aufruf geprüft | andere |
| WS19 | (Nr. 11 a, b, d) Nach einem Erfolg ruft jede der vier Routen `Append` genau einmal: die Art ihrer Route, `ServiceID` und `WorkspaceID` aus dem Ergebnis des Dienstes (nicht aus Pfad oder Body), `KorrelationsID` der Antwort, übrige Felder leer, vor der Rückkehr des Handlers, bei 201 und bei 200 | Der Fake liefert ein Ergebnis mit anderer `ID` und anderem `ServiceID` als Pfad und Body nennen; `FakeAudit.Aufrufe()` gleich 1; Eintrag Feld für Feld gegen Literale (alle sieben Felder) je Route: 4 × 7 = 28; Haltepunkt `FakeAudit.Vor`: die Operation ist gelaufen, der Recorder hat nichts geschrieben (kein `WriteHeader`, kein `Write`, kein Header des Fehlerobjekts), 4 Zellen; ein Request ohne Korrelations-ID: Eintrag mit leerer ID, 4 Zellen; Zählprüfung 4 Routen, 4 Arten | Eintrag aus Pfad oder Body; `WorkspaceID` leer; `Append` vor der Operation; zweimal; nur bei 201 (`reset` und `restore` ohne Eintrag); Arten von `reset` und `restore` vertauscht oder von Anlegen und Snapshot; `SessionID` oder `ImportID` gesetzt; `ErrorCode` gesetzt; `Append` nach `WriteHeader`; Eintrag nur bei nichtleerer Korrelations-ID | andere |
| WS20 | (Nr. 10, 11 c) Kein Eintrag der Route bei jedem anderen Ausgang der vier Routen und bei jedem Ausgang von `GET` und `DELETE`: Fehler des Dienstes, Verstoß vor dem Aufruf, Kontextfehler; die Route schreibt nie `workspace_geloescht` | M-Audit: Zellen Route × Ausgang: je Route 1 Erfolg + 11 allgemeine Ausgänge + die Codes der Tabelle von Nr. 8 (`POST /__admin/workspaces` 16, `reset` 15, `POST …/snapshots` 16, `restore` 18 = 65; `GET` Liste 14, `GET {id}` 15, `DELETE` 15, `GET …/snapshots` 15 = 59); zusammen 124; Orakel: `FakeAudit.Aufrufe()` gleich 1 in genau den 4 Erfolgszellen der Routen mit Eintrag, sonst 0; die 145 Bodys von WS7 und WS8, die 140 von WS18 und die 9 des Lesens zählen mit demselben Orakel; Zählprüfung 124 gleich 8 × 12 + 28 | Eintrag vor der Prüfung des Fehlers; Eintrag bei `ODR-E7006`, `ODR-E7007` oder `ODR-E7008`; Eintrag bei `GET`; Eintrag bei `DELETE` (doppelt mit dem Dienst); Eintrag `workspace_geloescht` | andere |
| WS21 | (Nr. 11 e) Ein Fehler von `Append` ändert Status, Body und Header der Antwort nicht | `FakeAudit.Fehler` gesetzt: Antwort Byte für Byte gleich der ohne Fehler, je Route mit Eintrag; `Append` 1 Aufruf; 4 Zellen | Fehler von `Append` als 500; Fehler von `Append` als Status 202; Antwort ohne Body | andere |
| WS22 | (`SPEC-234` Nr. 2 und 7, `SPEC-235` Nr. 3; `SPEC-298` Nr. 11, 14) `support.Arten()` führt `workspace_angelegt`, `workspace_zurueckgesetzt`, `snapshot_erzeugt` und `workspace_wiederhergestellt` als neunte bis zwölfte Zeile mit Pflicht `service_id` und `workspace_id` und sonst verboten; der Adapter des Ports schreibt `{"timestamp":…,"correlation_id":…,"art":…,"service_id":…,"workspace_id":…}`; in diesem Teil ändert sich in `src/bootstrap` allein `audit_betrieb_test.go` | `audit_test.go` des Ports und des Adapters: Schleife über die 5 Felder je Art (4 × 5 = 20 Zellen), Orakel Literale; der Text jeder Art Byte für Byte (4); `len(support.Arten())` gleich 12 (8 + 4); die Zählstelle in `audit_betrieb_test.go` zieht mit (die Arten sind wie `admin_abgewiesen` ausgenommen, das Binary ruft `admin` nicht); `TestAppendEineAufrufstelle` zählt die Stellen in `workspaces*.go`; die Schleifen von `TestKennungsform` und `TestLogBeiFehler` laufen über `support.Arten()` und führen die vier Arten mit (Paare aus Feld und Art: `workspace_id` kommt zu `service_id` hinzu); benannte Ausgabe: `git diff --stat` über `src/bootstrap` ist allein `audit_betrieb_test.go` | Pflichtfeld `workspace_id` oder `service_id` entfernt; `session_id` erlaubt; Text einer Art mit Leerzeichen; Zeile entfernt oder mit einer anderen vertauscht; Zeile des Adapters ohne `workspace_id`; Prüfung der Form für eine Art übersprungen; `Append` in einer Datei mehr oder einmal mehr; `WorkspaceRouten` aus `bootstrap` gerufen | andere |
| WS23 | (Nr. 11 Folge 3) Akzeptiertes Negativ: stirbt der Prozess zwischen der Operation und `Append`, ist der Workspace angelegt, zurückgesetzt, wiederhergestellt oder der Snapshot erzeugt, und der Eintrag fehlt | kein Fänger. Grund: kein Test beendet den Prozess zwischen zwei Aufrufen derselben Anfrage; der Haltepunkt `FakeAudit.Vor` zeigt die Lage (die Operation steht, der Eintrag fehlt), nicht das Ende des Prozesses. Folge: der Eintrag fehlt, und kein Lauf holt ihn nach; die Spur ist der Zustand, den `GET` liefert (die Revision, der Snapshot). Maß: 4 von 4 Routen mit Eintrag ohne Fänger für dieses Fenster; die Lage *Fehler von `Append`* fängt WS21 | nicht anwendbar | andere |
| WN1 | (Nr. 12) Gleichzeitige Requests beeinflussen einander nicht: jede Antwort und jeder Eintrag gehört zum eigenen Request, die Routen halten keinen Zustand; zwei gleichzeitige `reset` desselben Workspace: beide 200 mit `n+1` und `n+2`, zwei Einträge; zwei gleichzeitige `POST …/snapshots` mit demselben Namen: 201 mit einem Eintrag und `ODR-E7006` ohne | 16 Goroutinen je Route zugleich, jede mit eigener Kennung, die der Fake in `ID`, `ServiceID` und im Ergebnis zurückgibt; Orakel: Antwort und Eintrag tragen die eigene Kennung; Treffpunkt je Fenster: die drei `POST` mit Body 4 (nach dem Lesen des Bodys, im Aufruf, nach der Rückkehr der Operation, nach der Rückkehr von `Append`) = 12, `reset` 3 (ohne das Lesen) und die vier Routen ohne Eintrag 2 (im Aufruf, nach der Rückkehr) = 8: 23 Fenster, je 150 Runden zu 16 Requests; Fakes mit Zustand (eigene Sperre im Fake): `reset` hebt die Revision, `POST …/snapshots` vergibt den Namen einmal; Datenrennen: `make test-race`, je Prozess `-test.count=1`; Prüfgrenze 2400 Requests je Fenster in einem Prozess: Negativ (3); ein Zustand, der nur bei Überlappung wirkt und unter einer Sperre ohne Datenrennen liegt, und die Form des Fensters bei einem Prozessor: Negativ (4) | Paketvariable hält die letzte Antwort, die Antwort liest sie; Zelle unter einer Sperre (Kanal, kein Datenrennen) in jedem der 23 Fenster (23 Mutationen); Kennung des Eintrags aus einer geteilten Variable; Handler teilt `Anfrage` zwischen Requests | nebenläufig |
| WN2 | (Nr. 12) Keine Sperre über einem Aufruf von `workspaces`, `snapshots` oder `audit`: zwei Requests derselben Route sind zugleich im Dienst | Wirkungs-Fänger: die Operation des Fakes kehrt erst zurück, wenn ein zweiter Aufruf eingetreten ist (Kanal), ebenso `FakeAudit.Vor`; die Frist von 10 s ist nur der Abbruch im Fehlschlag; 9 Aufrufstellen (6 Operationen von `WorkspaceUseCase`, 2 von `SnapshotUseCase`, `Append`) | `sync.Mutex` um den Aufruf jeder der 8 Operationen (8 Mutationen); um `Append`; um Operation und `Append` in einem Fenster | nebenläufig |

Zeilen: WS17 bis WS23, WN1 und WN2 gleich 9 (*a* = 2, *b* = 7); WS1 bis WS16 (16 Zeilen, *a* = 0, *b* = 16) stehen in Teil 1; zusammen 25, keine Zeile in beiden Teilen.

**Mutationen und Raten** (Implementer; `AGENTS.md` §3.10, §3.15, §3.17): *wird beim Handoff gefüllt.* Je Zeile die Mutationen und der Test, der aus dem benannten Grund rot wurde; die Raten der Klasse `nebenläufig` (WN1 und WN2, auch WS17 und WS18) je Mutation mit Zahl der Prozesse, Messendem, Zahl der Prozessoren und Last; ein Fänger mit Wahrscheinlichkeit nennt seine Rate mit Zahl (`AGENTS.md` §3.10).

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, 2026-11-22; Generator oder Schleife, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15; das Orakel trägt Statuscodes, Codes, Namen und Bytes als Literale aus dem Vertrag, nie als Konstanten des Codes und nie als Aufruf des Lesers). Die Mengen von Teil 1 (M-Route bis M-Folge) stehen dort:

| Menge | Achsen und Elemente | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|
| M-Warten (WS17) | Routen 8 | Status, Recorder leer bis zur Freigabe | 8 gleich der Zahl der Routen |
| M-Kontext (WS18) | Routen mit Eintrag 4 × Enden 4 × Ausgänge 5 = 80; Routen ohne Eintrag 4 × Enden 3 × Ausgänge 5 = 60 | Abbruch oder Antwort, Eintrag nur beim Erfolg der vier Routen | 140 gleich 80 + 60; 8 Routen gleich 4 + 4 |
| M-Audit (WS19 bis WS21) | Eintrag Feld für Feld 4 × 7 = 28; Zellen Route × Ausgang 124 (65 + 59) | `FakeAudit.Aufrufe()` gleich 1 in genau 4 Zellen, sonst 0 | 124 gleich 8 × 12 + 28; 7 Felder gleich `NumField` von `support.Eintrag`; 4 Arten gleich den Zeilen 9 bis 12 von `Arten()` |
| M-Fenster (WN1) | Fenster 23: `POST` mit Body 3 × 4, `reset` 3, Routen ohne Eintrag 4 × 2 | Antwort und Eintrag tragen die eigene Kennung; Revisionen `n+1` und `n+2`; ein Eintrag bei einem Namen | 23 gleich 12 + 3 + 8 |
| M-Stellen (WN2) | Aufrufstellen 9 (6 + 2 + `Append`) | beide Aufrufe zugleich im Dienst | 9 gleich 6 + 2 + 1 gegen `NumMethod` beider Ports |

Die Mengen der ungültigen Bodys (M-Body ungültig, WS7, WS8) und der Fehler (M-Fehler, WS13) legt Teil 1 als Funktionen im Testpaket an; dieser Teil liest sie in WS20 (Sendung (b) in §1).

**Unbeschränkte Mengen und ihre Prüfgrenze (akzeptierte Negative; Grund, Folge, Maß):** (3) *Zahl der Requests je Fenster* (WN1): Prüfgrenze 2400 Requests je Fenster in einem Prozess; Grund: eine Wiederholung über die Grenze kostet Laufzeit ohne Gewinn; Folge: ein Zustand, der erst nach mehr Requests wirkt, bleibt grün; Maß: der Implementer trägt n von m ein. (4) *Zustand unter einer Sperre ohne Datenrennen und ein Prozessor* (WN1): ein Zustand, der allein bei Überlappung zweier Requests wirkt und unter einer Sperre liegt (Prüfen und Handeln in getrennten Abschnitten), meldet der Detektor nicht; WS16 (Teil 1) fängt jeden Zustand, der auch ohne Überlappung wirkt, der Treffpunkt nur, wenn das Fenster ihn umschließt; bei einem Prozessor trifft der Treffpunkt nicht jede Form des Fensters; Maß: der Implementer trägt je Mutation und Lauf n von m ein, mit Last des Hosts und Messendem (`AGENTS.md` §3.10). WS23 ist eine Zeile der Tabelle. Die Nummern (1), (2) und (5) stehen in Teil 1; (5) (Zustand nur bei Überlappung und Sperre ohne `sync` in den vier Routen ohne Eintrag) schließen WN1 und WN2 dieses Teils.

**Zweige als Menge** (Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat in `workspaces.go` null Blöcke mit Zähler 0, soweit dieser Teil sie führt; der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Z-4 Eintrag (vier Routen): Erfolg, Fehler von `Append`, Abbruch | WS18 bis WS21 | ein Lauf |

**Aufrufstellen und zählende Fakes** („jeder Fall je Aufrufstelle“ gilt für diese zwölf Stellen: die acht Operationen aus Teil 1 mit den Fällen dieses Teils und die vier Stellen von `Append`; `workspaceFake`, `snapshotFake` und der Reader des Bodys stammen aus Teil 1):

| Aufrufstelle | Port | Fall je Stelle in diesem Teil | Fake und Zählung |
|---|---|---|---|
| `List`, `Create`, `Get`, `Delete`, `Reset`, `Restore` (sechs Stellen, je eine Route) | `WorkspaceUseCase` | WS17, WS18, WN1, WN2 | `workspaceFake` (Teil 1; hier: Haltepunkt vor (im Aufruf) und nach (nach der Rückkehr) je Operation, der bis zur Freigabe blockiert, und ein Zustand mit eigener Sperre für WN1: `Reset` hebt die Revision) |
| `List`, `Create` (zwei Stellen, `GET` und `POST …/snapshots`) | `SnapshotUseCase` | WS17, WS18, WN1, WN2 | `snapshotFake` (Teil 1; hier: Haltepunkte und ein Zustand: ein Name wird einmal vergeben) |
| `Append` nach `POST /__admin/workspaces`, `reset`, `POST …/snapshots` und `restore` (vier Stellen) | `AuditLog` | WS19 bis WS22, WN2 (WS23: akzeptiertes Negativ) | `FakeAudit` (besteht): `Vor`, `Fehler`, `Aufrufe`, `Eintraege`; der Haltepunkt `auditNach` (aus den Tests der Sessions) |

Dazu der Reader für den Body aus Teil 1, der hier am Ende einen Treffpunkt hält (WN1) und einen Fehler einstellbar liefert (Teil 1, auch der Kontextfehler beim Lesen, WS13); ein Test-Server mit echtem Client gehört nicht dazu: die Tests rufen `ServeHTTP` mit `httptest`; die Zusagen über das Trennen der Verbindung prüft der Kontext des Requests (WS18), nicht der Socket (*Ungeprüft*, `SPEC-298`). `FakeAudit`, `FakeTelemetry` und die Fakes der Kennungen bestehen. Dieser Teil legt keinen neuen Fake an (*f* = 0): `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` gelten für die Erweiterung von `workspaceFake` und `snapshotFake`.

**Bindung an die Ports:** `bindung_test.go` bindet `workspaceFake` und `snapshotFake` seit Teil 1; dieser Teil ändert sie nicht.

**Größe und Schnitt** (Planner, 2026-11-22; `SPEC-128` Nr. 1, `AGENTS.md` §3.14). Dieser Teil: *a* = 2 (WN1, WN2), *b* = 7 (WS17 bis WS23), *f* = 0 (`workspaceFake` und `snapshotFake` aus Teil 1 werden erweitert), *p* = 45 (Eintrag im Handler 25, Arten des Audit 20), also *S* = 85·2 + 80·7 + 200·0 + 45 = 170 + 560 + 0 + 45 = 775; 135·775 = 104625, *M* = 100 · ⌈10,4625⌉ = 1100; keine Schnittfrage (*S* < 2223). Der Teil liegt 1448 unter der Schnittgrenze: jede weitere Zeile der Tabelle kostet 80, neunzehn weitere lösen die Schnittfrage neu aus (*S* = 2215 bei 18, 2295 bei 19 weiteren Zeilen der Klasse `andere`). Der ungeschnittene Plan hatte *S* = 2780 (*a* = 2, *b* = 23, *f* = 2, *p* = 370); Teil 1 trägt 2005, dieser Teil 775, die Summe gleicht 2780. Die Naht und der geteilte Mechanismus (`WorkspaceRouten`, die vier Handler mit Eintrag, beide Fakes) stehen in §3 von [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md) (*Größe und Schnitt*, *Warum diese Naht*); dieser Teil liegt nach Teil 1, nie neben ihm, und nach `slice-admin-imports-audit` (§4).

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-11-22 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* die Mittel, die Mengen als Funktionen, die Zellen von WS3 und der Rest von WS15 aus Teil 1 ([slice-admin-workspaces](../in-progress/slice-admin-workspaces.md)): angenommen, §1 *Übernimmt* trägt sie mit der Kennung des Gebers und die DoD nennt sie; kein *Ausdrücklich NICHT* dieses Teils trifft sie. *An diesen Slice* die Pflichten (b) und (c) von `SPEC-238` Nr. 15 (Geber `slice-admin-listener-token`): angenommen, §1 und DoD tragen sie. *Von diesem Slice an:* die Zellen Modus × Art für die vier Arten an [slice-bootstrap-m6](slice-bootstrap-m6.md) (dessen §1 *Ausdrücklich NICHT* schließt `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`, neue Endpunkte und das Umschalten des Modus aus und trifft die Zellen nicht; Eintrag in §1, DoD und *Prüfung der Sendungen* dort im selben Commit wie dieser Plan, mit der Kennung dieses Teils als Geber).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md) (Teil 1) **und** [slice-admin-imports-audit](../in-progress/slice-admin-imports-audit.md) liegen in `done/`; mit Teil 1 liegen die Bedingungen von Teil 1 in `done/` (`slice-admin-listener-token` samt seinen Teilen `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit`, `slice-audit-log`, `slice-audit-dienste`, `slice-workspace-use-case` und `slice-snapshot-use-case`). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): `SPEC-298` liegt vor (Architect, 2026-11-22), Tabelle, Raster und Mengen stehen in §3 und §6. Der Start gilt als Ausnahme je Slice wie der von Teil 1 ([welle-admin-api](../welle-admin-api.md) §2, Absatz *Schnitt vor dem Code* zu `slice-admin-workspaces`).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-admin-workspaces` (Teil 1): *Code-Abhängigkeit*. `WorkspaceRouten`, die Handler, `workspaceFake`, `snapshotFake`, der Reader des Bodys und die Mengen als Funktionen liegen dort; dieser Teil erweitert sie im selben Paket.
- `slice-admin-imports-audit`: *Reihenfolge, keine Typ-Abhängigkeit*. Beide Teile schreiben Arten in dieselben Dateien (`src/hexagon/ports/support/audit.go`, `audit_test.go`, `src/adapters/driven/audit/audit_test.go`, `log_sperre_test.go`, `src/bootstrap/audit_betrieb_test.go`, `abweisen_audit_test.go`): `import_veroeffentlicht` ist die achte Art, die vier Arten dieses Teils sind die neunte bis zwölfte, und die Zählstellen springen von 7 auf 8 (jener) und von 8 auf 12 (dieser); ohne die Reihenfolge stimmen Zeilen und Zählstellen nicht (`SPEC-297` Nr. 10, `SPEC-298` Nr. 11).
- `slice-admin-listener-token` und seine Teile: *Code-Abhängigkeit*. Der Rahmen (`Route`, Zuordnung, Fehlerabbildung, `Audit`) trägt die Routen.
- `slice-audit-log`: *Code-Abhängigkeit, Port*. `AuditLog` für den Eintrag; die Arten neun bis zwölf liegen in `support.Arten()`.
- `slice-audit-dienste`: der Eintrag der Löschung entsteht im Use Case, und die DoD verlangt „ohne Doppelung“ (nie `workspace_geloescht` aus der Route).

**Nicht gebraucht:** `slice-inspection-use-case` (der Endpunkt liest keine Exchanges), `slice-admin-sessions` mit seinen Teilen, `slice-admin-imports` (Teil 1; die Reihenfolge bindet `slice-admin-imports-audit`), `welle-mock` über die Bedingungen von Teil 1 hinaus, M2, `slice-http-adapter` und `slice-bootstrap-m6`.

**Reihenfolge und Konflikte** (Planner, 2026-11-22): Die zwei Teile schreiben im selben Paket `src/adapters/driving/admin/` und teilen den Konstruktor, die Handler und die beiden Fakes; sie laufen nacheinander in derselben Linie, nie zugleich: erst Teil 1, dann dieser Teil. Beim Start dieses Teils nimmt der Zweig `main` nach dem Merge von Teil 1 und von `slice-admin-imports-audit` auf. Die Dateien `src/hexagon/ports/support/audit*.go`, `src/adapters/driven/audit/audit_test.go`, `src/adapters/driven/audit/log_sperre_test.go`, `src/bootstrap/audit_betrieb_test.go` und `abweisen_audit_test.go` ändern allein Teile, die eine Art schreiben (hier die neunte bis zwölfte nach der achten). Slices, die `src/bootstrap/` mit dem Verzeichnis als Pfad messen, zählen die Zeilen von `audit_betrieb_test.go` als fremde mit (`SPEC-128` Nr. 10 g); der Block dieses Teils nennt die Dateien einzeln, damit weder dessen Zeilen noch die von `slice-admin-imports-audit` (Präfix `imports` im Verzeichnis `src/adapters/driving/admin/`) mitzählen. Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-22, Teil des Planner 2026-11-22): *S* = 775 gleich der Tabelle (9 Zeilen: *a* = 2, *b* = 7; *f* = 0; *p* = 45, Rechenweg in §3 *Größe und Schnitt*), **Marke 1100** gleich der Formel (135·775 = 104625, ⌈10,4625⌉ = 11); keine Schnittfrage. Die Messbasis ist der letzte Commit von `main` nach dem Merge von Teil 1 und von `slice-admin-imports-audit` (nach deren Closure und Move nach `done/`): sie ist erst dann bekannt. Der Block unten trägt deshalb den Teil des Architect (`pfad`, `produktcode`, `fakes`); **der Planner setzt `basis`, `schaetzung` (775) und `marke` (1100) beim Start dieses Teils vor dem ersten Code-Commit** (`AGENTS.md` §3.14; `BEO-REPO/marke-beim-start-nicht-gesetzt`) und prüft dabei die Tabelle neu. Bis zum Start meldet `make groessenmarke` den Plan nicht (er liegt nicht in `in-progress/`); `make groessenmarke GROESSENMARKE_PLAENE=<Verzeichnis mit einer Kopie>` meldet am 2026-11-22 `Hinweis: Marke steht aus: Schätzung 775, Marke 1100`, die Zahlen des Werkzeugs gleichen denen dieses Plans (Exit 0). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Die Pfade nennt der Plan einzeln (Dateien, nicht das Verzeichnis `src/adapters/driving/admin/`), damit die Zeilen fremder Slices im selben Verzeichnis nicht mitzählen (`SPEC-128` Nr. 10 g); legt der Implementer eine weitere Datei an oder ändert er eine weitere Datei von Teil 1, nennt er sie im selben Commit im Block. `workspaces_fake_test.go` und `workspaces_routen_test.go` stammen aus Teil 1: ihre Zeilen vor der Basis zählen nicht, die Änderungen dieses Teils zählen (die Datei `workspaces_routen_test.go` ist gegenüber dem Vorschlag des Architect ergänzt, weil dort die Zellen von WS3 stehen, Sendung (c)).

```groessenmarke
pfad: src/adapters/driving/admin/workspaces.go
pfad: src/adapters/driving/admin/workspaces_fake_test.go
pfad: src/adapters/driving/admin/workspaces_routen_test.go
pfad: src/adapters/driving/admin/workspaces_warten_test.go
pfad: src/adapters/driving/admin/workspaces_audit_test.go
pfad: src/adapters/driving/admin/workspaces_nebenlaeufig_test.go
pfad: src/adapters/driving/admin/abweisen_audit_test.go
pfad: src/hexagon/ports/support/audit.go
pfad: src/hexagon/ports/support/audit_test.go
pfad: src/adapters/driven/audit/audit_test.go
pfad: src/adapters/driven/audit/log_sperre_test.go
pfad: src/bootstrap/audit_betrieb_test.go
produktcode: 45
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Ein weiterer Schnitt ist nicht vorgesehen; die Schätzung liegt weit unter der Schnittgrenze.
- `in-progress` → `open` (blockiert — Carveout?): eine Randform ohne Ort im Vertrag, den der Implementer nicht entscheidet (`AGENTS.md` §3.12); die Fragen des Umfangs des Audit, des Kontexts bei Abbruch und der Nebenläufigkeit sind entschieden (`SPEC-298` Nr. 9, 11, 12).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): jede Zelle nennt eine Nummer von `SPEC-298` (Architect, 2026-11-22) oder „gilt nicht“ mit Grund; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12). Dieser Teil trägt die Zeile des Audit-Eintrags und die Spalte *Abbruch* der Zeilen 1 bis 5 des Rasters von Teil 1; die Spalten *fehlender*, *leerer* und *unbekannter Wert* und die Spalte *Wiederholung* dieser Zeilen und die Zeile der Authentisierung (alle fünf Spalten) stehen im Raster von [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md), Teil 1 (dort mit ihrem Ort). Ohne Nummer steht in den Zellen `SPEC-298`.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Audit-Eintrag der vier Routen (`workspace_angelegt`, `workspace_zurueckgesetzt`, `snapshot_erzeugt`, `workspace_wiederhergestellt`) | Nr. 11 b (`service_id` und `workspace_id` stammen aus dem Ergebnis des Dienstes; ein leeres Feld lehnt der Adapter des Ports ab, `SPEC-235` Nr. 4 a, *Ungeprüft*) | Nr. 11 b (leere `KorrelationsID`: der Adapter des Ports vergibt eine, `SPEC-234` Nr. 7; im Betrieb nicht erreichbar, der Rahmen vergibt immer) | gilt nicht: der Eintrag trägt keinen Wert des Clients (Nr. 11 b: nicht Pfad, nicht Body) | Nr. 9 d, 11 a (Eintrag auch bei beendetem `ctx` nach dem Erfolg; keiner bei einem Kontextfehler) | Nr. 11, 12 (je 2xx ein Eintrag; ein Fehler von `Append` ändert nichts, Nr. 11 e) |
| `GET` und `POST /__admin/workspaces`, Spalte Abbruch (Teil 1 trägt die übrigen) | Teil 1 | Teil 1 | Teil 1 | Nr. 9 (`ctx` unverändert, keine Frist, Kontextfehler beim Aufruf des Dienstes); Nr. 5 a (Lesen): Teil 1 | Teil 1 |
| `GET` und `DELETE /__admin/workspaces/{id}`, Spalte Abbruch | Teil 1 | Teil 1 | Teil 1 | Nr. 9 (ein Kontextfehler bricht ab, ohne Antwort; `DELETE` nach dem Commit: Antwort und Bereinigung laufen, Nr. 9 d, 10) | Teil 1 |
| `POST …/reset`, Spalte Abbruch | Teil 1 | Teil 1 | Teil 1 | Nr. 9 c (ein Kontextfehler heißt: nichts geändert, ohne Antwort und Eintrag; ein Erfolg bei beendetem `ctx` ist Antwort und Eintrag) | Teil 1 |
| `GET` und `POST …/snapshots`, Spalte Abbruch | Teil 1 | Teil 1 | Teil 1 | Nr. 9 (Kontextfehler, ein Blob ohne Referenz bleibt höchstens zurück); Nr. 5 a (Lesen): Teil 1 | Teil 1 |
| `POST …/restore`, Spalte Abbruch | Teil 1 | Teil 1 | Teil 1 | Nr. 9 c (nichts geändert, Cursor unverändert); Nr. 5 a (Lesen): Teil 1 | Teil 1 |

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; der Adapter kennt den Modus nicht, `SPEC-298` *Ungeprüft*; die Paare 1 bis 6, 9, 10 und 13 stehen in Teil 1, [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md); die Nummern sind die des ungeschnittenen Plans):

| Nr. | Paar | Entscheidung | Ort | Fänger |
|---|---|---|---|---|
| 7 | Erfolg × Fehler von `Append` (je Route mit Eintrag) | die Antwort bleibt, der Eintrag fehlt | Nr. 11 e | WS21 |
| 8 | Erfolg × beendeter `ctx` (je Route, auch `DELETE`) | Antwort und, bei den vier Routen, Eintrag, die niemand liest | Nr. 9 d, 11 a | WS18 |
| 11 | zwei gleichzeitige `reset` desselben Workspace; zwei gleichzeitige `POST …/snapshots` mit demselben Namen | beide `reset` gelingen mit `n+1` und `n+2` und je einem Eintrag; ein `POST` ist 201 mit Eintrag, das andere `ODR-E7006` ohne | Nr. 12, `SPEC-129` Nr. 14 c, `SPEC-260` Nr. 25 d | WN1 (Fakes mit Zustand) |
| 12 | `reset` oder `restore` × gleichzeitige Mutation im Modus `mock`, Replay oder Löschen | die Dienste ordnen; die Route fügt nichts hinzu und reicht den `ctx` | Nr. 12, `SPEC-129` Nr. 11, 19 | gilt nicht für den Adapter (Ordnung im Dienst); WS3 (Teil 1) prüft die Durchreichung des `ctx`, WN2 dass die Route keine Sperre hält |
| 14 | Modus (`record`, `replay`, `mock`, `hybrid`) | der Adapter kennt den Modus nicht; die Verdrahtung gehört dem Composition Root | Nr. 14, *Ungeprüft* | gilt nicht für den Adapter |

- Wer schreibt den Audit-Eintrag einer Löschung: der Adapter oder `WorkspaceUseCase` — **Ausgang:** entschieden (Architect, 2026-11-22): allein der Use Case (`SPEC-234` Nr. 7, `SPEC-129` Nr. 9 b1); die Route schreibt bei `DELETE` keinen Eintrag (`SPEC-298` Nr. 10, 11 c), vier Arten gehören der Route (`SPEC-298` Nr. 11; WS19, WS20).
- Die Wirkung von `reset` und `restore` bei gleichzeitiger Mutation im Modus `mock` gehört den Use Cases (`LH-FA-07.a`, Serialisierung je Workspace); der Adapter reicht den Kontext durch — **Ausgang:** entfallen: Zuständigkeit der Use Cases; der Adapter-Test prüft die Durchreichung des `ctx` (WS3 in Teil 1, WS18 hier).
- Der Audit-Eintrag nennt nicht, aus welcher Baseline ein Workspace angelegt und welcher Snapshot wiederhergestellt oder erzeugt wurde (die Menge der Felder ist geschlossen, `SPEC-298` Nr. 11 *Entscheidung* Alternative 5) — **Ausgang:** entschieden (Architect, 2026-11-22) als akzeptiertes Negativ im Vertrag (Folge im Betrieb (1)); ein neues Feld wäre eine Änderung von `SPEC-234`, `support.Eintrag` und `SPEC-235` Nr. 3; ob der Betreiber das verlangt, steht beim User.
- Zellen Modus × Art für die vier Arten (wie die Zellen der Arten `session_angelegt` und `session_beendet`, `slice-admin-sessions-beenden` Review F-10, und `import_veroeffentlicht`): der Test in `src/bootstrap/audit_betrieb_test.go` nimmt die Arten wie `admin_abgewiesen` aus, weil das Binary `admin` nicht aufruft; kein Plan nannte eine Zelle für diese Arten — **Ausgang:** eingetreten: [slice-bootstrap-m6](slice-bootstrap-m6.md) trägt die Zellen (Eintrag des Planners vom 2026-11-22 beim Schnitt in §1, DoD und *Prüfung der Sendungen* dort, Geber: dieser Teil; `AGENTS.md` §3.13).
- Die Reihenfolge gegenüber `slice-admin-imports-audit`: dieser Teil schreibt die Arten neun bis zwölf in dieselben Dateien wie jener die achte — **Ausgang:** entschieden (Planner, 2026-11-22): Start-Trigger in §4, dieser Teil startet erst, wenn `slice-admin-imports-audit` und Teil 1 in `done/` liegen; die Aussage steht in §3 von [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md) und in `welle-admin-api` §5.
- Die Marke steht im Block erst beim Start, weil die Messbasis (der letzte Commit von `main` nach dem Merge von Teil 1 und von `slice-admin-imports-audit`) vorher nicht feststeht — **Ausgang:** weiter offen; Adresse: Planner beim Start (§4: `basis`, `schaetzung` 775, `marke` 1100 vor dem ersten Code-Commit; `BEO-REPO/marke-beim-start-nicht-gesetzt`). Die Zeilen fremder Slices im Verzeichnis `src/adapters/driving/admin/` zählen nicht mit, weil der Block die Dateien einzeln nennt (`BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`).
- Die Fänger der Nebenläufigkeit (WN1, WN2, auch WS17 und WS18) hängen an der Last des Hosts: die Rate je Mutation steht mit Zahl, Messendem, Prozessoren und Last im Bericht (`AGENTS.md` §3.10) — **Ausgang:** weiter offen → `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`: die Erweiterung von `workspaceFake` und `snapshotFake` (Haltepunkte, Zustand) trägt jede Schnittstelle des Originals, zählt jeden Aufruf und hält keine Panik — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Akzeptierte Negative (§3: (3) Zahl der Requests je Fenster, (4) Zustand unter einer Sperre ohne Datenrennen und ein Prozessor, WS23 Eintrag fehlt bei Prozessende) — **Ausgang:** weiter offen; Adresse: Reviewer und Verifier lösen sie gegen die Klassen auf (`AGENTS.md` §3.15, §3.17); der Implementer trägt die Maße (n von m) der Messung ein.
- Die Fänger WS17, WS18, WN1 und WN2 prüfen auch den Code der vier Handler ohne Eintrag (`GET` Liste, `GET {id}`, `DELETE`, `GET …/snapshots`), den Teil 1 schreibt und dieser Teil nicht ändert (Rückgabe des Planners vom 2026-11-22, `AGENTS.md` §3.13; wie `GET` bei den Imports, IM18, IM19, IN1, IN2 in `slice-admin-imports-audit`) — **Ausgang:** entschieden (Architect, 2026-11-22; Gründe, Ort und Maß in §6 von [slice-admin-workspaces](../in-progress/slice-admin-workspaces.md)): die Zeilen WS17 bis WS23, WN1 und WN2 und die Zahlen dieses Teils bleiben (WS18 mit 140 Zellen, *S* = 775, Marke 1100). Teil 1 trägt die statischen Formen in WS15 (0 Aufrufe von `Err`, `Done`, `Deadline`, `Cause`, 0 Importe von `sync` und `sync/atomic` in `workspaces*.go`) und den Zustand ohne Überlappung in jeder Route in WS16; WS17, WS18, WN1 und WN2 sind für die vier Handler ohne Eintrag der dynamische Fänger und schließen das akzeptierte Negativ (5) von Teil 1 (Zustand nur bei Überlappung, Sperre ohne `sync`: 8 Zellen). Eine Mutation von WS18 (`ctx.Err()` vor dem Aufruf, Abbruch nach dem Erfolg) ist auch für WS15 rot; der Implementer weist die Rotfärbung von WS18 mit einem Lauf allein von WS18 nach (`-run`), der Bericht nennt beide Tests.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-22, 84 Einträge; der Zähler ist die Zahl der Dateien unter `evidence/`, der Planner zieht Treffer beim Start nach). Offen und für den Slice einschlägig: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (7×) für die Erweiterung von `workspaceFake`, `snapshotFake` und `FakeAudit`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (9×) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (19×) für WN1 und WN2, `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×) und `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (3×) für die Größenmarke (§4), außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (21×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (`AGENTS.md` §3.17), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (`AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (`AGENTS.md` §3.16).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
