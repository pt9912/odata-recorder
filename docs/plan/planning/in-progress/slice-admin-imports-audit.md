# Slice slice-admin-imports-audit: Admin-Endpunkte für Imports, Teil 2: Eintrag `import_veroeffentlicht`, Warten und Abbruch, Nebenläufigkeit (Teil 2 des Schnitts von `slice-admin-imports`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-18`, `LH-FA-13`, `LH-RB-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-297` Nr. 9 (Dauer, Kontext, Abbruch), Nr. 10 (Audit: der gelungene Import) und Nr. 11 (Nebenläufigkeit; den Zustand und die Wiederholung trägt Teil 1) und Nr. 12 (Geheimnisse, für die Felder des Eintrags; die Antworten trägt Teil 1) · `SPEC-234` Nr. 2, 4 d, 5 und 7 (die Art `import_veroeffentlicht`) · `SPEC-235` Nr. 3, 4 d · `SPEC-238` Nr. 9, 15 b und c (Rahmen, nicht geändert) · `SPEC-127` Nr. 5, 15 c und 18 (Fenster nach `Publish`, gleichzeitige Imports) · `SPEC-005` (Zeile `POST /__admin/imports`: Audit-Logging) · `LH-FA-13.a` (Änderungen sind auditierbar) · `ARC-007` · `architecture.md` §4.2.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-admin-imports-audit`, Worktree `/Development/odata-recorder-wt-admin-imports-audit`; der Zweig entsteht von `main` nach dem Start-Commit, der Implementer nimmt `main` vor dem ersten Code-Commit per `git merge main` auf)

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-admin-imports](../done/slice-admin-imports.md), Vorschlag des Architect vom 2026-11-22). **Datum:** 2026-11-22.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem** Abschnitt des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Admin-Adapter schreibt nach jedem gelungenen `POST /__admin/imports` (201 und 200) genau einen Eintrag `import_veroeffentlicht` (neue achte Art von `support.Arten()`; `service_id` und `import_id` aus dem Ergebnis des Dienstes, Korrelations-ID der Antwort), nie bei einem Fehler und nie bei `GET`; ein Fehler von `Append` ändert die Antwort nicht. Der Teil hält die Zusagen über Warten und Abbruch (die Antwort entsteht erst nach der Rückkehr von `Import` und `Get`, keine Frist, der `ctx` bleibt der des Requests, ein Kontextfehler bricht ab ohne Eintrag, ein Erfolg bei beendetem `ctx` ist Antwort und Eintrag) und über die Nebenläufigkeit (gleichzeitige Requests beeinflussen einander nicht, keine Sperre über einem Aufruf von `Import`, `Get` oder `Append`). Er hebt die Zählstelle der Arten von `support.Arten()` von 7 auf 8.

**Vertrag (Architect, 2026-11-22):** `SPEC-297` Nr. 9 bis 11; der Vertrag steht auf `main`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice war ungeschnitten über der Schnittgrenze von `SPEC-128` Nr. 1 (Schätzung 2490 gegen 2223); der Architect schlug in §3 von [slice-admin-imports](../done/slice-admin-imports.md) (*Größe und Schnitt*) zwei Teile vor, der Planner hat sie am 2026-11-22 vor dem Code vollzogen: [slice-admin-imports](../done/slice-admin-imports.md) trägt `ImportRouten`, Body, Antworten, Fehlerabbildung und Import-Autorisierung (Zeilen IM1 bis IM17), dieser Teil den Eintrag, das Warten und den Abbruch und die Nebenläufigkeit (IM18 bis IM24, IN1, IN2).

**Übernimmt (Sendung von `slice-admin-imports`, Teil 1; Geber: `slice-admin-imports`; Eintrag des Planners vom 2026-11-22 beim Schnitt, `AGENTS.md` §3.13; §1 *Ausdrücklich NICHT* und DoD dieses Teils vorher geschrieben und gelesen, kein Ausschluss trifft eine Sendung):** (a) die Mittel aus Teil 1, die dieser Teil erweitert: der Konstruktor `ImportRouten` (hält `audit` seit Teil 1, ruft es nicht; dieser Teil ruft `Append` im Handler von `POST`), die Handler von `POST` und `GET`, `importsFake` (Teil 1 liefert Zählen, Festhalten jedes Arguments und einstellbare Ausgänge; dieser Teil ergänzt die Haltepunkte `vor` und `nach` je Operation und den Zustand mit eigener Sperre für IN1), der zählende Reader des Bodys (Teil 1 liefert Zählen und den einstellbaren Fehler; dieser Teil ergänzt den Treffpunkt am Ende für IN1) und der Haltepunkt `auditNach` aus den Tests der Sessions; (b) die Mengen der ungültigen Bodys (IM7 und IM8: 86 Fälle; seit der Fix-Runde dazu `importRandUnicode` mit 24 Formen, die 44 ungültigen Elemente von `importZeichenMenge` und `importElementeUngueltig` mit 305 Formen, siehe (e)) und der ungültigen Reports (IM12: 21 Fälle; dazu `reportMitNutzlast` mit den 44 ungültigen Elementen), die IM21 mit demselben Orakel liest: Teil 1 legt sie als Funktionen im Testpaket an (nicht als Literal in einem Test), dieser Teil ruft sie; (c) die Zellen von IM3 (seit der Fix-Runde acht: Route × `Neu` × Zähler; `FakeAudit.Aufrufe()` gleich 0 in allen acht): dieser Teil hebt die vier Zellen von `POST` auf 1, im selben Commit wie der Code (IM20); (d) der Rest von IM16: `git diff --stat` über `src/bootstrap` ist in diesem Teil allein `audit_betrieb_test.go`; Teil 1 prüft ihn leer, gemessen gegen den `main`-Elter des letzten Merge (`git diff --stat <main-Elter> -- src/bootstrap`: gegen die Messbasis ist die Ausgabe nach dem Merge von `main` nicht leer, sie zeigt dann die Dateien fremder Slices), und diese Prüfung gilt auch hier; er steht als benannte Ausgabe in IM23; (e) was die Fix-Runde von Teil 1 an Mengen und Prüfungen ergänzt hat und dieser Teil liest oder berührt (Eintrag des Planners vom 2026-11-23 bei der Closure von Teil 1, `AGENTS.md` §3.13; §1 *Ausdrücklich NICHT* und DoD dieses Teils vorher gelesen, kein Ausschluss trifft die Sendung): `importRandUnicode` (24 Formen), die 44 ungültigen Elemente von `importZeichenMenge` (mal 3 Stellen im Wert, auch im Report über `reportMitNutzlast`), `importElementeUngueltig` (305 Formen) und die Zellen von M-Lesen (IM13, 208 Zellen) liegen als Funktionen im Testpaket, die IM21 mit demselben Orakel liest; die 208 Zellen führt `TestIM13Lesen` als Schleife im Test selbst, keine Funktion: dieser Teil extrahiert sie in eine Funktion (und nennt dann `imports_antwort_test.go` im Block, §4) oder liest sie nicht, und IM21 zählt nur, was als Funktion besteht; `TestIM15QuelleDerRouten` liest in `imports*.go` jede Datei, auch künftige (die Menge bildet der Test aus dem Verzeichnis), und verbietet `Err`, `Done` und `Deadline` auf einem Kontext, `context.Cause`, `sync` und `sync/atomic`; der Handler von `POST` mit `Append` auch bei beendetem `ctx` (`SPEC-297` Nr. 9 d) braucht keine dieser Stellen, und wo `Append` bei beendetem `ctx` eine braucht, berührt dieser Teil den Quelltest (`imports_quelltext_test.go`, im Block zu nennen) und trägt das in §6 ein, bevor er den Test ändert.

**Übernimmt (Sendung von `slice-admin-listener-token`, Pflichten der Endpunkte nach `SPEC-238` Nr. 15, Schnitt vom 2026-11-07, `AGENTS.md` §3.13; die Hälfte (c) dieses Teils, die übrigen Hälften trägt Teil 1):** (b) die Route ruft `Audit.Append` nie für `ODR-E2101` und schreibt kein Fehlerobjekt selbst: mit dem Aufruf von `Append` in diesem Teil bekommt der AST-Test IM15 aus Teil 1 sein Gegenstück, die Zählstelle `TestAppendEineAufrufstelle`; (c) die vier Arten von `SPEC-234` Nr. 2 schreibt der Dienst, nie die Route; die Art für den `POST` ist eine Zeile im Vertrag dieses Endpunkts (`import_veroeffentlicht`, `SPEC-297` Nr. 10): sie liefert dieser Teil (IM20 bis IM23). Der Reviewer löst jede Zeile von Nr. 15 mit einer Mutation auf.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `ImportRouten` (Konstruktor, Zuordnung), Body von `POST`, Antworten, Fehlerabbildung, Import-Autorisierung, Pfadwert und der Report als Body — [slice-admin-imports](../done/slice-admin-imports.md) (Teil 1); dieser Teil erweitert den Handler von `POST` um den Eintrag und ändert die Dateien von Teil 1 nur, wo §3 es nennt.
- Die Fachlogik des Imports — `ImportUseCase` (`done/`); der Adapter bildet nur HTTP ab.
- Der Audit-Eintrag eines fehlgeschlagenen Imports (Art `import_fehlgeschlagen`) — `ImportUseCase.Import` schreibt ihn (`slice-audit-dienste`); der Adapter schreibt die vier Arten von `SPEC-234` Nr. 2 nie, auch nicht für denselben `POST` (Entscheidung `SPEC-234` Nr. 7, Geber: `slice-audit-log` und `slice-audit-dienste`). Dieser Teil schreibt allein den Eintrag des gelungenen Imports; der Eintrag zu `ODR-E2101` gehört dem Rahmen (`slice-admin-abweisen-audit`).
- Die Zelle Modus × Art für `import_veroeffentlicht` im Test `src/bootstrap/audit_betrieb_test.go` — [slice-bootstrap-m6](../open/slice-bootstrap-m6.md): das Binary ruft `admin` erst mit diesem Slice; dieser Teil nimmt die Art wie `admin_abgewiesen` aus dem Test aus und ändert dort die Zählstelle (7 → 8).
- Der Port `AuditLog` — `welle-betrieb` (`slice-audit-log`); der Teil ruft ihn.
- Ein Audit-Eintrag je Anfrage oder je Versuch eines Imports, eine Art je Status (201 und 200 unterscheiden) und eine Drosselung — `SPEC-297` Nr. 10 (*Entscheidung*, Alternativen 2 und 4) und `SPEC-234` Nr. 2 legen es anders fest; eine Änderung ist eine Änderung des Vertrags.
- Ein Handbuch-Kapitel — das Binary bleibt unverändert (`SPEC-297` Nr. 13: kein Fehlercode, Schlüssel oder Kommando, kein Aufruf von `ImportRouten` aus `bootstrap`); das Kapitel zu `audit.jsonl` führt [slice-admin-sessions-beenden](../done/slice-admin-sessions-beenden.md) in §6 als offenen Punkt (Adresse User) und ist nicht Gegenstand dieses Teils.

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

- [ ] Audit: ein gelungener `POST` (201 oder 200) schreibt genau einen Eintrag `import_veroeffentlicht` nach der Rückkehr von `Import`, vor der Rückkehr des Handlers, mit `ServiceID` und `ImportID` aus dem Ergebnis des Dienstes und der Korrelations-ID der Antwort; kein Fehler, kein Verstoß vor dem Aufruf, kein Kontextfehler, kein ungültiger Report und kein `GET` schreibt einen; ein Fehler von `Append` ändert die Antwort nicht; die Art ist die achte Zeile von `support.Arten()` (`service_id` und `import_id` Pflicht, jedes andere Feld verboten) und die Zählstellen der Tests ziehen mit (7 → 8), die Schleifen von `TestKennungsform` und `TestLogBeiFehler` führen sie, `TestAppendEineAufrufstelle` zählt die Stelle in `imports.go` (IM20 bis IM23). Sendung (e) von `slice-admin-imports` (Geber: `slice-admin-imports`): IM21 liest die Mengen von Teil 1 als Funktionen mit demselben Orakel (86 Bodys, `importRandUnicode` 24, die 44 ungültigen Elemente von `importZeichenMenge`, `importElementeUngueltig` 305, 21 Reports und `reportMitNutzlast` mit den ungültigen Elementen; die 208 Zellen von `TestIM13Lesen` nur nach der Extraktion in eine Funktion), die vier `POST`-Zellen der acht Zellen von IM3 zählen 1 (IM20), und `TestIM15QuelleDerRouten` liest auch die Dateien dieses Teils und bleibt grün, ohne dass `Append` bei beendetem `ctx` `Err`, `Done`, `Deadline`, `Cause` oder `sync` braucht (braucht es eine, steht der Eingriff in den Quelltest vorher in §6).
- [ ] Warten und Abbruch: die Antwort von `POST` und `GET` entsteht erst nach der Rückkehr des Dienstes, ohne Frist und ohne Ablösung des Kontexts; ein Kontextfehler beim Aufruf des Dienstes bricht ab, ohne Fehlerobjekt, Zählen und Eintrag (der Kontextfehler beim Lesen des Bodys liegt in Teil 1, IM13); ein Erfolg bei beendetem `ctx` ist Antwort und Eintrag (IM18, IM19).
- [ ] Nebenläufigkeit: gleichzeitige Requests beeinflussen einander nicht, keine Sperre über einem Aufruf von `Import`, `Get` oder `Append`; zwei gleichzeitige `POST` mit derselben `import_id` und gleicher Anfrage: einmal 201, einmal 200, byte-gleicher Body, zwei Einträge (IN1, IN2). Pflichten (b) und (c) von `SPEC-238` Nr. 15 für den Eintrag (Geber `slice-admin-listener-token`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (`SPEC-297` Nr. 13: kein Code, Schlüssel oder Kommando; der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/adapters/driving/admin/imports.go` | update | der Handler von `POST` ruft `Append` nach dem Erfolg (kein neuer Dateiname); keine Änderung an `imports_body.go`, `admin.go`, `route.go`, `fehler.go`, `lebenszyklus.go`, `sessions*.go` |
| `src/hexagon/ports/support/audit.go`, `src/hexagon/ports/support/audit_test.go` | update | die Art `import_veroeffentlicht` (`ArtImportVeroeffentlicht`), achte Zeile von `Arten()` (`SPEC-297` Nr. 10); der Test zählt 7 Arten und wächst auf 8 |
| `src/adapters/driven/audit/audit_test.go`, `src/adapters/driven/audit/log_sperre_test.go`, `src/bootstrap/audit_betrieb_test.go` | update (nur Tests) | Zählstellen der Arten (7 → 8) und Zeile der Art; die Schleifen von `TestKennungsform` und `TestLogBeiFehler` laufen über `support.Arten()` und führen die achte Art mit; `audit_betrieb_test.go` nimmt die Art wie `admin_abgewiesen` aus (das Binary ruft `admin` nicht), sonst bleibt `bootstrap` unverändert |
| `src/adapters/driving/admin/abweisen_audit_test.go` | update (nur Test) | `TestAppendEineAufrufstelle` zählt `Append` je Datei; der Eintrag nach dem Import ist eine weitere Stelle (`imports.go` gleich 1) |
| `src/adapters/driving/admin/imports_fake_test.go`, `src/adapters/driving/admin/imports_routen_test.go` | update (nur Tests) | aus Teil 1: `importsFake` wächst um die Haltepunkte `vor` und `nach` je Operation und den Zustand mit Sperre, der Reader des Bodys um den Treffpunkt am Ende (Sendung (a)); die acht Zellen von IM3 heben die vier Zellen von `POST` auf 1 (Sendung (c)); `imports_antwort_test.go` (die Schleife von `TestIM13Lesen` als Funktion) und `imports_quelltext_test.go` (der Quelltest, wenn `Append` bei beendetem `ctx` eine der verbotenen Stellen braucht) nur, wenn Sendung (e) es verlangt; beide stehen seit dem Start im Block in §4 (ohne Änderung zählen sie 0) |
| `spec/spezifikation.md` | geliefert (Architect, auf `main`) | `SPEC-297`; Nachzug `SPEC-234` Nr. 2, 4 d und 7, `SPEC-238` Nr. 15 c; dieser Slice schreibt sie nicht |
| Tests | neu | Adapter gegen `importsFake` (aus Teil 1, hier erweitert) und `FakeAudit`: jedes Argument festgehalten, jeder Aufruf gezählt, auch die Operation, die die Route nie rufen darf (`AGENTS.md` §3.15); ein Fall je Aufrufstelle für den `ctx` des Requests und für `Append`. Neue Dateien: `imports_audit_test.go` (IM20 bis IM23: Eintrag, Ausgänge, Fehler von `Append`, Art), `imports_warten_test.go` (IM18, IM19), `imports_nebenlaeufig_test.go` (IN1, IN2) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem 2026-11-22 (Architect) mit dem Vertrag in der Spezifikation (`SPEC-297`) vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. `Nr.` ist die Kennung des ungeschnittenen Plans: **IM18 bis IM24** gehören zu diesem Teil, **IN1** und **IN2** sind seine nebenläufigen Zeilen; IM1 bis IM17 stehen in Teil 1 ([slice-admin-imports](../done/slice-admin-imports.md)). Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest; die Zeilen sind vom Planner beim Schnitt im Wortlaut übernommen, mit einer Ausnahme (IM23 trägt den Rest von IM16, Sendung (d)). Der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein (`AGENTS.md` §3.10).

**Entscheidungen des Architect für diesen Teil** (2026-11-22, vor dem ersten Code-Commit, `AGENTS.md` §3.12; Ort in der Spezifikation; die übrigen stehen in §3 von [slice-admin-imports](../done/slice-admin-imports.md)):

1. **Abbruch** (`SPEC-297` Nr. 9): der Kontext bleibt der des Requests (keine Ablösung, keine Frist); ein Kontextfehler bricht die Verbindung ab, ohne Eintrag; die Dauer ist unbegrenzt (kein `WriteTimeout`); im Fenster zwischen `Publish` und dem Ende von `SetMetadata` bleibt die Baseline stehen, die Wiederholung (200) setzt die Metadaten.
2. **Audit** (`SPEC-297` Nr. 10, `SPEC-234` Nr. 2 und 7): eine neue Art `import_veroeffentlicht` (achte Art; `service_id` und `import_id` Pflicht, sonst verboten), geschrieben von der Route nach jedem Erfolg (201 und 200), nie bei einem Fehler; der Dienst schreibt allein `import_fehlgeschlagen`, der Rahmen allein `admin_abgewiesen`; ein Fehler von `Append` ändert nichts.
3. **Nebenläufigkeit** (`SPEC-297` Nr. 11): keine Sperre über einem Aufruf von `imports` oder `audit`; zwei gleichzeitige `POST` mit gleicher `import_id` und gleicher Anfrage: 201 und 200, byte-gleich, zwei Einträge.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-22, Zeilen IM18 bis IM24, IN1 und IN2 des ungeschnittenen Plans; jede Zeile trägt ihre Menge):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| IM18 | (Nr. 9 a, b) Die Antwort von `POST` entsteht erst nach der Rückkehr von `Import`: bis dahin ist nichts an den Client geschrieben, und die Route setzt keine Frist; dasselbe für `GET` und `Get` | Wirkungs-Fänger (deterministisch, Kanäle): `Import` blockiert, der Test sieht, dass `ServeHTTP` nicht zurückgekehrt ist und der Recorder leer ist; nach der Freigabe 201 mit derselben Antwort wie ohne Wartezeit; `ctx.Deadline()` im Fake nicht gesetzt; ebenso `Get`; das fehlende `WriteTimeout` fängt `grenzen_test.go` des Rahmens; 2 Routen | `Import` in einer Goroutine und Antwort 202 sofort; `context.WithTimeout(ctx, 30*time.Second)`; `WriteHeader` vor dem Aufruf (im Rahmen: die Route schreibt nicht selbst, der Haltepunkt sieht den Kopf trotzdem) | andere |
| IM19 | (Nr. 9 c, d; `SPEC-238` Nr. 9) Endet der `ctx`, läuft der Aufruf mit dem `ctx` des Requests unverändert; ein Kontextfehler beim Aufruf des Dienstes bricht die Verbindung ab, ohne Fehlerobjekt, Zählen und Eintrag (der Kontextfehler beim Lesen des Bodys: IM13 in Teil 1); ein Erfolg bei beendetem `ctx` ist Antwort und Eintrag | M-Kontext: `POST` × Ende des `ctx` (`cancel` vor dem Aufruf, abgelaufene Frist vor dem Aufruf, im Aufruf, nach der Rückkehr von `Import` vor `Append`: 4) × Ausgang (Erfolg, `Canceled`, `DeadlineExceeded`, je allein und mit einem Wert: 5) = 20; `GET` × 3 Enden (vor dem Aufruf, im Aufruf, nach der Rückkehr) × 5 = 15; zusammen 35; Orakel `SPEC-297` Nr. 9 c, d: Abbruch oder Antwort, Eintrag nur beim Erfolg; der Fake sieht `ctx.Done()` nach `cancel()` (kein `WithoutCancel`) und ein Wert im `ctx` kommt an | `context.WithoutCancel(ctx)`; Eintrag auch bei Kontextfehler; kein Eintrag bei beendetem `ctx` nach dem Erfolg; Abbruch bei beendetem `ctx` nach dem Erfolg; `ctx.Err()` vor `Import` oder vor `Get` geprüft | andere |
| IM20 | (Nr. 10 a, b, d) Nach einem Erfolg von `Import` ruft `POST` `Append` genau einmal: Art `import_veroeffentlicht`, `ServiceID` und `ImportID` aus dem Ergebnis des Dienstes (nicht aus dem Body), `KorrelationsID` der Antwort, übrige Felder leer, vor der Rückkehr des Handlers, bei 201 und bei 200 | Der Fake liefert einen Import mit anderer `ID` und anderem `ServiceID` als der Body nennt; `FakeAudit.Aufrufe()` gleich 1; Eintrag Feld für Feld gegen Literale (alle sieben Felder); je 201 und 200; Haltepunkt `FakeAudit.Vor`: `Import` ist gelaufen, der Recorder hat nichts geschrieben (kein `WriteHeader`, kein `Write`, kein Header des Fehlerobjekts); ein Request ohne Korrelations-ID: Eintrag mit leerer ID | Eintrag aus dem Body; `ImportID` leer; `Append` vor `Import`; zweimal; nur bei 201; `SessionID` aus `SourceSessionIDs` gesetzt; `ErrorCode` gesetzt; `Append` nach `WriteHeader`; Eintrag nur bei nichtleerer Korrelations-ID | andere |
| IM21 | (Nr. 10 c) Kein Eintrag der Route bei jedem anderen Ausgang von `POST` und `GET`: Fehler des Dienstes, Verstoß vor dem Aufruf, Kontextfehler, ungültiger Report; `GET` schreibt nie einen; die Route schreibt nie `import_fehlgeschlagen` | M-Audit: Zellen Route × Ausgang: `POST` 2 Erfolge (201, 200) + 8 + 9 = 19, `GET` 1 + 8 + 3 = 12; zusammen 31; Orakel: `FakeAudit.Aufrufe()` gleich 1 in genau den 2 Erfolgszellen von `POST`, sonst 0; die 86 Bodys von IM7 und IM8, die 21 von IM12 und die 35 von IM19 zählen mit demselben Orakel; Zählprüfung 31 gleich (2 + 8 + 9) + (1 + 8 + 3) | Eintrag vor der Prüfung des Fehlers; Eintrag bei `ODR-E6102`; Eintrag bei `GET`; Eintrag bei ungültigem Report; Eintrag `import_fehlgeschlagen` bei einem Fehler | andere |
| IM22 | (Nr. 10 e) Ein Fehler von `Append` ändert Status, Body und Header der Antwort nicht | `FakeAudit.Fehler` gesetzt: Antwort Byte für Byte gleich der ohne Fehler, je 201 und 200; `Append` 1 Aufruf; 2 Zellen | Fehler von `Append` als 500; Fehler von `Append` als Status 202 | andere |
| IM23 | (`SPEC-234` Nr. 2, 4 d und 7, `SPEC-235` Nr. 3; `SPEC-297` Nr. 10) `support.Arten()` führt `import_veroeffentlicht` als achte Zeile mit Pflicht `service_id` und `import_id` und sonst verboten; der Adapter des Ports schreibt `{"timestamp":…,"correlation_id":…,"art":"import_veroeffentlicht","service_id":…,"import_id":…}`; in diesem Teil ändert sich in `src/bootstrap` allein `audit_betrieb_test.go` (Rest von IM16) | `audit_test.go` des Ports und des Adapters: Schleife über die 5 Felder je Art, Orakel Literale; `len(support.Arten())` gleich 8; die Zählstelle in `audit_betrieb_test.go` zieht mit (die Art ist wie `admin_abgewiesen` ausgenommen, das Binary ruft `admin` nicht); `TestAppendEineAufrufstelle` zählt die Stelle in `imports.go`; die Schleifen von `TestKennungsform` und `TestLogBeiFehler` laufen über `support.Arten()` und führen die achte Art mit (Paare aus Feld und Art: `import_id` kommt neu hinzu); benannte Ausgabe: `git diff --stat` über `src/bootstrap` ist allein `audit_betrieb_test.go` | Pflichtfeld `import_id` entfernt; `session_id` erlaubt; Text der Art mit Leerzeichen; Zeile entfernt oder mit der siebenten vertauscht; Zeile des Adapters ohne `import_id`; Prüfung der Form für die Art übersprungen; `Append` in einer Datei mehr oder einmal mehr; `ImportRouten` aus `bootstrap` gerufen | andere |
| IM24 | (Nr. 10 Folge) Akzeptiertes Negativ: bricht der Kontext zwischen dem Commit von `Publish` und dem Ende von `Import` ab, oder stirbt der Prozess zwischen `Import` und `Append`, steht die Baseline ohne Eintrag | kein Fänger. Grund: der Route ist nicht sichtbar, ob `Publish` schon geschrieben hat (der Dienst liefert den Kontextfehler ohne Ergebnis, und er schreibt dann keinen Eintrag, `SPEC-234` Nr. 4 d), und kein Test beendet den Prozess dazwischen. Folge: der Eintrag fehlt, bis die Wiederholung (200) ihn schreibt; die Spur ist allein die Baseline, die `GET` liefert. Maß: 2 von 2 Fenstern ohne Fänger (Kontextfehler in `SetMetadata`, Prozessende zwischen `Import` und `Append`); die dritte Lage (Fehler von `Append`) fängt IM22 | nicht anwendbar | andere |
| IN1 | (Nr. 11) Gleichzeitige `POST` und `GET` beeinflussen einander nicht: jede Antwort und jeder Eintrag gehört zum eigenen Request, die Routen halten keinen Zustand; zwei gleichzeitige `POST` mit derselben `import_id` und gleicher Anfrage: 201 und 200 mit byte-gleichem Body, zwei Einträge | 16 Goroutinen je Route zugleich, jede mit eigener `import_id`, die der Fake in `ServiceID`, `ID` und im Report zurückgibt; Orakel: Antwort und Eintrag tragen die eigene Kennung; Treffpunkt je Fenster: `POST` 4 (nach dem Lesen des Bodys, im Aufruf von `Import`, nach seiner Rückkehr, nach der Rückkehr von `Append`), `GET` 2 (im Aufruf von `Get`, nach seiner Rückkehr) = 6 Fenster, 150 Runden zu je 32 Requests je Route; ein Fake mit Zustand (Sperre im Fake) liefert dem zweiten `Import` mit derselben `import_id` `Neu` falsch: einmal 201, einmal 200, gleiche Bytes, 2 Einträge; Datenrennen: `make test-race`, je Prozess `-test.count=1`; Prüfgrenze 4800 Requests je Route in einem Prozess: Negativ (3); bei einem Prozessor fängt der Treffpunkt nicht jede Form des Fensters: Negativ (4) | Paketvariable hält den letzten Report, die Antwort liest sie; Zelle unter einer Sperre (Kanal, kein Datenrennen) in jedem der 6 Fenster (6 Mutationen); Kennung des Eintrags aus einer geteilten Variable; Handler teilt `Anfrage` zwischen Requests | nebenläufig |
| IN2 | (Nr. 11) Keine Sperre über einem Aufruf von `Import`, `Get` oder `Append`: zwei Requests derselben Route sind zugleich im Dienst | Wirkungs-Fänger: `importsFake.Import` kehrt erst zurück, wenn ein zweiter Aufruf eingetreten ist (Kanal), ebenso `Get` und `FakeAudit.Vor`; die Frist von 10 s ist nur der Abbruch im Fehlschlag; 3 Aufrufstellen | `sync.Mutex` um den Aufruf von `Import`; um `Get`; um `Append`; um `Import` und `Append` in einem Fenster | nebenläufig |

Zeilen: IM18 bis IM24, IN1 und IN2 gleich 9 (*a* = 2, *b* = 7).

**Mutationen und Raten** (Implementer; `AGENTS.md` §3.10, §3.15, §3.17): *wird beim Handoff gefüllt.* Je Zeile die Mutationen und der Test, der aus dem benannten Grund rot wurde; die Raten der Klasse `nebenläufig` (IN1 und IN2, auch IM18 und IM19) je Mutation mit Zahl der Prozesse, Messendem, Zahl der Prozessoren und Last; ein Fänger mit Wahrscheinlichkeit nennt seine Rate mit Zahl (`AGENTS.md` §3.10).

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, 2026-11-22; Generator oder Schleife, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15; das Orakel trägt Statuscodes, Codes, Namen und Bytes als Literale aus dem Vertrag, nie als Konstanten des Codes und nie als Aufruf des Lesers). Die Mengen von Teil 1 (M-Route bis M-Folge) stehen dort:

| Menge | Achsen und Elemente | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|
| M-Kontext (IM19) | `POST` 4 Enden × 5 Ausgänge = 20; `GET` 3 × 5 = 15; zusammen 35 (die Achse *Lesen* steht in Teil 1, M-Fehler, IM13) | Abbruch oder Antwort, Eintrag nur beim Erfolg | 35 gleich 20 + 15 |
| M-Audit (IM21) | `POST` 19 + `GET` 12 = 31 Zellen, genau 2 mit 1 Eintrag | `FakeAudit.Aufrufe()` und `Eintraege()` als Literale | 31 gleich (2 + 8 + 9) + (1 + 8 + 3); 2 Zellen mit Eintrag |
| M-Fenster (IN1) | `POST` 4 + `GET` 2 = 6 Treffpunkte; 150 Runden × 32 Requests | Antwort und Eintrag tragen die eigene Kennung | 6 gleich der Zahl der Haltepunkte des Fakes und des Audit (`vor`, `nach`, `auditNach`, Lesen) |

**Unbeschränkte Mengen und ihre Prüfgrenze (akzeptierte Negative; Grund, Folge, Maß):** (3) *Zahl der gleichzeitigen Requests* (IN1): Prüfgrenze 16 zugleich in 150 Runden, 4800 Requests je Route in einem Prozess; Folge: die Routen halten keinen Zustand, die Zahl ändert nichts; ein Zustand, der erst nach mehr als 4800 Requests je Route in einem Prozess wirkt, bleibt grün. (4) *Fenster ohne Fuge und enges Fenster ohne Detektor* (IN1): ein Zustand unter einer Sperre, der in einem Fenster ohne Fuge innerhalb der Route geschrieben und sofort zurückgelesen wird, ist bei einem Prozessor nicht gefangen (ein Treffpunkt erzwingt ein Interleaving erst ab zwei Prozessoren); ein Datenrennen ohne Detektor im engen Fenster nach dem Treffpunkt fängt allein `make test-race`; Maß aus dem Vorgänger: bei einem Prozessor 3 von 7 Fensterformen ungefangen; hier sind es 6 Fenster, der Implementer misst bei einem und bei vier Prozessoren und trägt n von 6 ein. Die Nummern (1) und (2) stehen in Teil 1; IM24 ist eine Zeile der Tabelle.

**Zweige als Menge** (Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat in `imports.go` null Blöcke mit Zähler 0; der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Z-3 Handler `POST`: der Eintrag (Erfolg bei 201 und 200, Fehler von `Append`), Abbruch bei beendetem `ctx` nach dem Erfolg | IM19 bis IM22 | ein Lauf |

**Aufrufstellen und zählende Fakes** („jeder Fall je Aufrufstelle“ gilt für diese drei Stellen; `importsFake` und der Reader des Bodys stammen aus Teil 1):

| Aufrufstelle | Port | Fall je Stelle in diesem Teil | Fake und Zählung |
|---|---|---|---|
| `Import` (`POST /__admin/imports`) | `ImportUseCase` | IM18, IM19, IM21, IN1, IN2 | `importsFake` (Teil 1; hier: Haltepunkt vor (im Aufruf) und nach (nach der Rückkehr), ein Zustand mit eigener Sperre für IN1 (zweiter `Import` mit derselben `import_id`: `Neu` falsch); der Haltepunkt blockiert bis zur Freigabe) |
| `Get` (`GET /__admin/imports/{id}`) | `ImportUseCase` | IM18, IM19, IN1, IN2 | wie oben |
| `Append` nach `Import` | `AuditLog` | IM20 bis IM23, IN2 | `FakeAudit` (besteht): `Vor`, `Fehler`, `Aufrufe`, `Eintraege`; der Haltepunkt `auditNach` (aus den Tests der Sessions) |

Dazu der Reader für den Body aus Teil 1, der hier am Ende einen Treffpunkt hält (IN1) und einen Fehler einstellbar liefert (Teil 1, auch der Kontextfehler beim Lesen, IM13); ein Test-Server mit echtem Client gehört nicht dazu: die Tests rufen `ServeHTTP` mit `httptest`; die Zusagen über das Trennen der Verbindung prüft der Kontext des Requests (IM19), nicht der Socket (*Ungeprüft*, `SPEC-297`). `FakeAudit`, `FakeTelemetry` und die Fakes der Kennungen bestehen. Dieser Teil legt keinen neuen Fake an (*f* = 0): `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` gelten für die Erweiterung von `importsFake`.

**Bindung an die Ports:** `bindung_test.go` bindet `importsFake` seit Teil 1; dieser Teil ändert sie nicht.

**Größe und Schnitt** (Planner, 2026-11-22; `SPEC-128` Nr. 1, `AGENTS.md` §3.14). Dieser Teil: *a* = 2 (IN1, IN2), *b* = 7 (IM18 bis IM24), *f* = 0 (`importsFake` aus Teil 1 wird erweitert), *p* = 25 (Eintrag im Handler 15, Art des Audit 10), also *S* = 85·2 + 80·7 + 200·0 + 25 = 170 + 560 + 0 + 25 = 755; 135·755 = 101925, *M* = 100 · ⌈10,19⌉ = 1100; keine Schnittfrage (*S* < 2223). Der Teil liegt 1468 unter der Schnittgrenze: jede weitere Zeile der Tabelle kostet 80, neunzehn weitere lösen die Schnittfrage neu aus (*S* = 2195 bei 18, 2275 bei 19 weiteren Zeilen der Klasse `andere`; 755 + 80·18 = 2195, 755 + 80·19 = 2275). Der ungeschnittene Plan hatte *S* = 2490 (*a* = 2, *b* = 24, *f* = 1, *p* = 200); Teil 1 trägt 1735, dieser Teil 755, die Summe gleicht 2490. Die Naht und der geteilte Mechanismus (`ImportRouten`, Handler, `importsFake`) stehen in §3 von [slice-admin-imports](../done/slice-admin-imports.md) (*Der Schnitt*); dieser Teil liegt nach Teil 1, nie neben ihm.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-11-22 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* die Mittel, die Mengen als Funktionen, die Zellen von IM3 und der Rest von IM16 aus Teil 1 ([slice-admin-imports](../done/slice-admin-imports.md)): angenommen, §1 *Übernimmt* trägt sie mit der Kennung des Gebers und die DoD nennt sie; kein *Ausdrücklich NICHT* dieses Teils trifft sie. *An diesen Slice, am 2026-11-23 bei der Closure von Teil 1 gelesen (§1 *Ausdrücklich NICHT*, DoD und §4 dieses Teils):* die Mengen der Fix-Runde, die acht Zellen von IM3, der Quelltest über alle `imports*.go` und die Messung von IM16 gegen den `main`-Elter (Sendung (e)): angenommen, §1 *Übernimmt* und DoD tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie (der Ausschluss „ändert die Dateien von Teil 1 nur, wo §3 es nennt“ lässt die Extraktion in `imports_antwort_test.go` und den Eingriff in den Quelltest zu, weil §3 sie bedingt nennt und der Block sie bei Bedarf trägt). *An diesen Slice* die Pflichten (b) und (c) von `SPEC-238` Nr. 15 (Geber `slice-admin-listener-token`): angenommen, §1 und DoD tragen sie. *Von diesem Slice an:* die Zelle Modus × Art für `import_veroeffentlicht` an [slice-bootstrap-m6](../open/slice-bootstrap-m6.md) (dessen §1 *Ausdrücklich NICHT* schließt `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`, neue Endpunkte und das Umschalten des Modus aus und trifft die Zelle nicht; Eintrag in §1, DoD und *Prüfung der Sendungen* dort im selben Commit wie dieser Plan, mit der Kennung dieses Teils als Geber).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-admin-imports](../done/slice-admin-imports.md) (Teil 1) liegt in `done/`; mit ihm liegen die Bedingungen von Teil 1 in `done/` (`slice-admin-listener-token` samt seinen Teilen `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit`, `slice-audit-log` und `slice-audit-dienste`). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): `SPEC-297` liegt vor (Architect, 2026-11-22), Tabelle, Raster und Mengen stehen in §3 und §6.

**Stand der Bedingungen beim Start** (Planner, 2026-11-23, `main` = `dfb26754` und die Moves nach `next/` und `in-progress/` bis `b199f9af`): [slice-admin-imports](../done/slice-admin-imports.md) (Teil 1) liegt in `done/` (Merge `dfb26754`), mit ihm `slice-admin-listener-token` samt `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit`, `slice-audit-log` und `slice-audit-dienste`; `slice-baseline-import` liegt unter `done/welle-record-replay/`, `ImportUseCase` mit `Import` und `Get` in `src/hexagon/ports/inbound/import.go` (geprüft an den Dateien). `SPEC-297` samt Nr. 9 bis 11, Tabelle IM18 bis IM24, IN1 und IN2, Raster und Mengen steht auf `main`. Die Welle `welle-admin-api` ist nicht gestartet (ihr Trigger M5 ist Priorisierung, keine Code-Abhängigkeit); der Start dieses Teils gilt als Ausnahme je Slice wie der von [slice-admin-imports](../done/slice-admin-imports.md) (Teil 1), `slice-admin-sessions` und `slice-admin-listener-token` ([welle-admin-api](../welle-admin-api.md) §2, Absatz *Schnitt vor dem Code* zu `slice-admin-imports`, „ebenso der Start von Teil 2“): die Freigabe des Users zum Start der Admin-Slices liegt vor (Reihenfolge `welle-admin-api` zuerst), und das Binary bleibt unverändert (IM23). Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`. In `in-progress/` laufen sonst [slice-admin-workspaces](slice-admin-workspaces.md) (Zweig `slice-admin-workspaces`; im Paket `admin` allein Dateien mit dem Präfix `workspaces` und eine Zeile in `bindung_test.go`, die dieser Teil nicht ändert; `git diff --name-only main...slice-admin-workspaces` nennt keine Datei dieses Blocks) und [slice-telemetry-einbau-replay](slice-telemetry-einbau-replay.md) (Zweig `slice-telemetry-einbau-replay`; `src/bootstrap/wurzel.go` und fünf Testdateien dort (`wurzel_test.go`, `m4_wurzel_test.go`, `telemetrie_wurzel_test.go`, `telemetrie_gateway_test.go`, `telemetrie_replay_test.go`), `src/hexagon/services/replay/`; `git diff --name-only main...slice-telemetry-einbau-replay` nennt weder `audit_betrieb_test.go` noch eine andere Datei dieses Blocks). [slice-admin-workspaces-audit](../open/slice-admin-workspaces-audit.md) schreibt die neunte bis zwölfte Art in dieselben vier Dateien und startet erst, wenn dieser Teil in `done/` liegt (sein Start-Trigger nennt es). Die Sendungen aus `done/` an diesen Slice (die Mittel, die Mengen als Funktionen, die acht Zellen von IM3, der Quelltest, die Messung von IM16, die Pflichten (b) und (c) von `SPEC-238` Nr. 15) prüfte der Planner bei der Closure von Teil 1 (2026-11-23, siehe *Prüfung der Sendungen* in §3); beim Start tritt keine hinzu.

**Abhängigkeiten — Grund je Eintrag:**

- `slice-admin-imports` (Teil 1): *Code-Abhängigkeit*. `ImportRouten`, die Handler, `importsFake`, der Reader des Bodys und die Mengen als Funktionen liegen dort; dieser Teil erweitert sie im selben Paket.
- `slice-admin-listener-token` und seine Teile: *Code-Abhängigkeit*. Der Rahmen (`Route`, Zuordnung, Fehlerabbildung, `Audit`) trägt die Routen.
- `slice-audit-log`: *Code-Abhängigkeit, Port*. `AuditLog` für den Eintrag; die achte Art liegt in `support.Arten()`.
- `slice-audit-dienste`: der Eintrag des fehlgeschlagenen Imports entsteht im Use Case, und die DoD verlangt „ohne Doppelung“ (nie `import_fehlgeschlagen` aus der Route).

**Nicht gebraucht:** `slice-inspection-use-case` (der Endpunkt liest keine Exchanges), `slice-admin-sessions` mit seinen Teilen, `welle-mock`, M2, `slice-http-adapter` und `slice-bootstrap-m6`.

**Reihenfolge und Konflikte** (Planner, 2026-11-22): Die zwei Teile schreiben im selben Paket `src/adapters/driving/admin/` und teilen den Konstruktor, den Handler von `POST` und `importsFake`; sie laufen nacheinander in derselben Linie, nie zugleich: erst Teil 1, dann dieser Teil. Beim Start dieses Teils nimmt der Zweig `main` nach dem Merge von Teil 1 auf. Die Dateien `src/hexagon/ports/support/audit*.go`, `src/adapters/driven/audit/audit_test.go`, `src/adapters/driven/audit/log_sperre_test.go` und `src/bootstrap/audit_betrieb_test.go` ändern allein Teile, die eine Art schreiben (hier die achte; [slice-admin-workspaces-audit](../open/slice-admin-workspaces-audit.md) schreibt die neunte bis zwölfte Art in dieselben Dateien und startet erst, wenn dieser Teil in `done/` liegt). Slices, die `src/bootstrap/` mit dem Verzeichnis als Pfad messen (`slice-telemetry-einbau-gateway`, Stand 2026-11-22, inzwischen in `done/`; das laufende `slice-telemetry-einbau-replay` nennt die Dateien von `src/bootstrap/` einzeln, ohne `audit_betrieb_test.go`), zählen die Zeilen von `audit_betrieb_test.go` als fremde mit (`SPEC-128` Nr. 10 g); der Block dieses Teils nennt die Dateien einzeln, damit weder dessen Zeilen noch die von `slice-admin-inspektion` (Präfix `inspektion` im Verzeichnis `src/adapters/driving/admin/`) und `slice-admin-workspaces` mitzählen. Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-22, Teil des Planner 2026-11-22): *S* = 755 gleich der Tabelle (9 Zeilen: *a* = 2, *b* = 7; *f* = 0; *p* = 25, Rechenweg in §3 *Größe und Schnitt*), **Marke 1100** gleich der Formel (135·755 = 101925, ⌈10,19⌉ = 11); keine Schnittfrage. **Messbasis** `b199f9af` (Planner, 2026-11-23, beim Start gesetzt: der Commit des Verweis-Nachzugs nach dem Move nach `in-progress/` auf `main`, der letzte Commit von `main` vor dem Code-Start und nach dem Merge von Teil 1 `dfb26754`; die Start-Commits ändern allein Pläne, Roadmap und Welle, keinen Pfad des Blocks; die Messung ist 0). Die Tabelle ist neu nachgerechnet: 9 Zeilen (IM18 bis IM24 *andere*, IN1 und IN2 *nebenläufig*), *a* = 2, *b* = 7, *f* = 0, *p* = 25, *S* = 85·2 + 80·7 + 200·0 + 25 = 755, *M* = 100 · ⌈135·755 / 10000⌉ = 100 · ⌈10,19⌉ = 1100; keine Schnittfrage (755 < 2223). Der Block trägt `basis: b199f9af`, `schaetzung: 755` und `marke: 1100`; `schaetzung` und `marke` setzte der Planner vor dem ersten Code-Commit (`AGENTS.md` §3.14; `BEO-REPO/marke-beim-start-nicht-gesetzt`). Geprüft gegen den Stand von `main` und der Zweige: `git diff --name-only main...slice-admin-workspaces` und `main...slice-telemetry-einbau-replay` nennen keinen Pfad des Blocks (`bindung_test.go` steht nicht im Block, dieser Teil ändert dort keine Zeile), die Basis trägt keine fremde Zeile, und ein Merge von `main` bringt voraussichtlich keine; der Implementer misst nach jedem Merge von `main` dennoch `git diff --numstat <basis>..<main-Elter des letzten Merge> -- <Pfade des Blocks>`, nennt fremde Zeilen in der Messpunkt-Zeile und setzt `ueberschreitung` nur bei Überschreitung (`SPEC-128` Nr. 4 g und Nr. 10 g; `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`). Der Block nennt `imports_antwort_test.go` und `imports_quelltext_test.go` schon beim Start (Sendung (e): die Extraktion der Schleife von `TestIM13Lesen`, der Eingriff in den Quelltest, wenn `Append` bei beendetem `ctx` eine verbotene Stelle braucht); ohne Änderung zählen sie 0. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Die Pfade nennt der Plan einzeln (Dateien, nicht das Verzeichnis `src/adapters/driving/admin/`), damit die Zeilen fremder Slices im selben Verzeichnis nicht mitzählen (`SPEC-128` Nr. 10 g); legt der Implementer eine weitere Datei an, nennt er sie im selben Commit im Block. `imports_fake_test.go` und `imports_routen_test.go` stammen aus Teil 1: ihre Zeilen vor der Basis zählen nicht, die Änderungen dieses Teils zählen.

```groessenmarke
pfad: src/adapters/driving/admin/imports.go
pfad: src/adapters/driving/admin/imports_fake_test.go
pfad: src/adapters/driving/admin/imports_routen_test.go
pfad: src/adapters/driving/admin/imports_audit_test.go
pfad: src/adapters/driving/admin/imports_warten_test.go
pfad: src/adapters/driving/admin/imports_nebenlaeufig_test.go
pfad: src/adapters/driving/admin/imports_antwort_test.go
pfad: src/adapters/driving/admin/imports_quelltext_test.go
pfad: src/adapters/driving/admin/abweisen_audit_test.go
pfad: src/hexagon/ports/support/audit.go
pfad: src/hexagon/ports/support/audit_test.go
pfad: src/adapters/driven/audit/audit_test.go
pfad: src/adapters/driven/audit/log_sperre_test.go
pfad: src/bootstrap/audit_betrieb_test.go
produktcode: 25
fakes: 0
basis: b199f9af
schaetzung: 755
marke: 1100
```

**Messpunkt beim Start** (Planner, 2026-11-23, `make groessenmarke` im Hauptbaum unter der Sperre `/tmp/odata-recorder-heavy.lock`, `main` = `b199f9af` ohne den Start-Commit; Exit 0, keine Befunde, kein Eintrag nötig): `groessenmarke slice-admin-imports-audit.md: ok, Messung 0 von Marke 1100` und `Messpunkt slice-admin-imports-audit.md: zeilen=9 nebenlaeufig=2 fakes=0 produktcode=25 schaetzung=755 marke=1100 messung=0` (Schätzung des Blocks gleich Tabelle, Marke gleich Formel). Beim Handoff und bei der Closure trägt der Planner die Zeile des Programms mit Stand und von Hand gemessenem Produktcode nach (`SPEC-128` Nr. 9). Teil 1 endete bei Auslastung 1,03 (2477 von 2400, Ende ÷ *S* = 1,43, die Zeilen kamen allein in den Tests hinzu, [slice-admin-imports](../done/slice-admin-imports.md) §7 *Größe*); die Formel gilt unverändert (`SPEC-128` Nr. 2), die Marke dieses Teils ist 1100, und die Mengen aus Teil 1 (Sendung (e)) kosten Test-Zeilen.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Ein weiterer Schnitt ist nicht vorgesehen; die Schätzung liegt weit unter der Schnittgrenze.
- `in-progress` → `open` (blockiert — Carveout?): eine Randform ohne Ort im Vertrag, den der Implementer nicht entscheidet (`AGENTS.md` §3.12); die Fragen des Umfangs des Audit, des Kontexts bei Abbruch und der Nebenläufigkeit sind entschieden (`SPEC-297` Nr. 9, 10, 11).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): jede Zelle nennt eine Nummer von `SPEC-297` (Architect, 2026-11-22) oder „gilt nicht“ mit Grund; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12). Dieser Teil trägt die Zeile des Audit-Eintrags und die Spalte *Abbruch* der übrigen; die Spalten *fehlender*, *leerer* und *unbekannter Wert* und die Spalte *Wiederholung* der übrigen Zeilen stehen im Raster von [slice-admin-imports](../done/slice-admin-imports.md), Teil 1 (dort mit ihrem Ort). Ohne Nummer steht in den Zellen `SPEC-297`.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Audit-Eintrag des `POST` (Art `import_veroeffentlicht`) | Nr. 10 b (jedes Feld je Art gesetzt nach `SPEC-234` Nr. 2; eine leere Korrelations-ID vergibt der Adapter des Ports, `SPEC-235` Nr. 4 d) | Nr. 10 e (eine leere Kennung weist `Append` ab, ohne Wirkung auf die Antwort) | Nr. 10 e und *Ungeprüft* (Kennung außerhalb der Form: für den Dienst nicht erreichbar) | Nr. 10 a, d und 9 d (`Append` ohne `ctx`: ein beendeter Kontext ändert den Eintrag nicht, `SPEC-234` Nr. 5 c; ein Kontextfehler des Dienstes: kein Eintrag) | Nr. 10 a und 11 (je 2xx genau ein Eintrag, keiner bei einem Fehler der Route) |
| `POST /__admin/imports`, Spalte Abbruch (Teil 1 trägt die übrigen) | Teil 1 | Teil 1 | Teil 1 | Nr. 9 (`ctx` unverändert, keine Frist; ein Kontextfehler bricht ab, ohne Eintrag; das Fenster nach `Publish` heilt die Wiederholung); Nr. 5 a (Lesen): Teil 1 | Teil 1 |
| `GET /__admin/imports/{id}`, Spalte Abbruch | Teil 1 | Teil 1 | Teil 1 | Nr. 9 (ein Kontextfehler bricht ab, ohne Antwort) | Teil 1 |
| Antwort aus `ReportJSON`, Spalte Abbruch | Teil 1 | Teil 1 | Teil 1 | Nr. 9 d (ein Erfolg bei beendetem `ctx` ist Antwort und Eintrag, die niemand liest) | Teil 1 |

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; der Adapter kennt den Modus nicht, `SPEC-297` *Ungeprüft*; die übrigen Paare stehen in Teil 1):

| Paar | Entscheidung | Ort | Fänger |
|---|---|---|---|
| Erfolg des Dienstes × Fehler von `Append` | der Erfolg gilt, die Antwort bleibt | Nr. 10 e | IM22 |
| Erfolg des Dienstes × ungültiger Report | `ODR-E9001`, kein Eintrag | Nr. 6, 10 c | IM12 (Teil 1), IM21 |
| Erfolg des Dienstes × beendeter `ctx` | Antwort und Eintrag | Nr. 9 d | IM19 |
| zwei gleichzeitige `POST`, gleiche `import_id`, gleiche Anfrage | einmal 201, einmal 200, byte-gleicher Body, zwei Einträge | Nr. 11 | IN1 (Fake mit Zustand) |

- Ein langer Import und der Abbruch des Clients: das Rollback ist atomar (`SPEC-047`), das Verhalten gegen Timeouts des Servers hatte keinen Ort — **Ausgang:** entschieden (Architect, 2026-11-22; IM18, IM19 fangen den Vertrag): `SPEC-297` Nr. 9 (der Kontext bleibt der des Requests, keine Frist, kein `WriteTimeout`; ein Kontextfehler bricht ab; das Fenster zwischen `Publish` und `SetMetadata` heilt die Wiederholung); die Folge für einen Client mit kurzem Timeout steht im Vertrag.
- Umfang des Audit (nur erfolgreiche oder jeder Import) — **Ausgang:** entschieden (Architect, 2026-11-22; IM20 bis IM23): `SPEC-297` Nr. 10 und `SPEC-234` Nr. 2, 4 d, 7: die neue Art `import_veroeffentlicht` nach jedem Erfolg (201 und 200), nie von der Route bei einem Fehler; der Dienst schreibt `import_fehlgeschlagen`, der Rahmen `admin_abgewiesen`.
- Zelle Modus × Art für `import_veroeffentlicht` (wie die Zellen der Arten `session_angelegt` und `session_beendet`, `slice-admin-sessions-beenden` Review F-10): der Test in `src/bootstrap/audit_betrieb_test.go` nimmt die Art wie `admin_abgewiesen` aus, weil das Binary `admin` nicht aufruft; kein Plan nannte eine Zelle für diese Art — **Ausgang:** eingetreten: [slice-bootstrap-m6](../open/slice-bootstrap-m6.md) trägt die Zelle (Eintrag des Planners vom 2026-11-22 beim Schnitt in §1, DoD und *Prüfung der Sendungen* dort, Geber: dieser Teil; `AGENTS.md` §3.13).
- Die Marke steht im Block erst beim Start, weil die Messbasis (der letzte Commit von `main` nach dem Merge von Teil 1) vorher nicht feststeht — **Ausgang:** entfallen (Planner, 2026-11-23, beim Start): der Block trägt `basis: b199f9af`, `schaetzung: 755` und `marke: 1100` vor dem ersten Code-Commit, die Tabelle ist nachgerechnet (§4, *Messbasis*); `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×) bekommt keine neue Evidenz.
- Fremde Zeilen unter den Pfaden des Blocks nach einem Merge von `main` (`BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`, 5 Dateien unter `evidence/`): die Zeilen fremder Slices im Verzeichnis `src/adapters/driving/admin/` zählen nicht mit, weil der Block die Dateien einzeln nennt; beim Start nennen die Zweige von `slice-admin-workspaces` und `slice-telemetry-einbau-replay` keinen Pfad des Blocks (§4, *Messbasis*), `slice-admin-workspaces-audit` startet erst nach diesem Teil — **Ausgang:** weiter offen; Adresse: Implementer (misst fremde Zeilen nach jedem Merge von `main` als `git diff --numstat <basis>..<main-Elter des letzten Merge> -- <Pfade des Blocks>`, nennt sie in der Messpunkt-Zeile, setzt `ueberschreitung` nur bei Überschreitung) und Planner bei der Closure (Evidenz im Register, wenn die Messung fremde Zeilen enthielt).
- Die Marke 1100 liegt bei *S* = 755 nach der Formel; Teil 1 endete bei Auslastung 1,03 (Ende ÷ *S* = 1,43), weil die Mengen der Fix-Runde als Schleifen mit Zählprüfung nur Test-Zeilen kosteten, und dieser Teil liest sie (Sendung (e), IM21 mit 31 Zellen und den Mengen aus Teil 1) — **Ausgang:** weiter offen; Adresse: Implementer (misst beim Handoff, trägt `ueberschreitung` mit Grund nur bei Überschreitung ein) und Planner (Messpunkt bei Handoff und Closure; über 1,00 geht die Zahl an den Architect, `SPEC-128` Nr. 9).
- Braucht `Append` bei beendetem `ctx` (Nr. 9 d) eine der Stellen, die `TestIM15QuelleDerRouten` verbietet (`Err`, `Done`, `Deadline`, `Cause`, `sync`; Sendung (e)), berührt dieser Teil `imports_quelltext_test.go` — **Ausgang:** weiter offen; Adresse: Implementer (trägt den Eingriff vor der Änderung des Tests hier ein, `AGENTS.md` §3.13).
- Die Fänger der Nebenläufigkeit (IN1, IN2, auch IM18 und IM19) hängen an der Last des Hosts: die Rate je Mutation steht mit Zahl, Messendem, Prozessoren und Last im Bericht (`AGENTS.md` §3.10) — **Ausgang:** weiter offen → `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (9×) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (19×; beide verkörpert in `AGENTS.md` §3.10) im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, verkörpert in `.claude/commands/implement-slice.md` Schritt 14 und im Reviewer-Skill) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (7×, verkörpert im Reviewer-Skill): die Erweiterung von `importsFake` (Haltepunkte, Zustand) trägt jede Schnittstelle des Originals, zählt jeden Aufruf und hält keine Panik — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Akzeptierte Negative (§3: (3) Zahl der gleichzeitigen Requests, (4) Fenster ohne Fuge und enges Fenster ohne Detektor, IM24 Eintrag fehlt in zwei Fenstern) — **Ausgang:** weiter offen; Adresse: Reviewer und Verifier lösen sie gegen die Klassen auf (`AGENTS.md` §3.15, §3.17); der Implementer trägt die Maße (n von m) der Messung ein.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-22, 84 Einträge) und beim Start erneut (2026-11-23, 84 Verzeichnisse unter `BEO-REPO/`; der Zähler ist die Zahl der Dateien unter `evidence/`, Treffer und Zähler beim Start nachgezogen). Einschlägig für die Erweiterung von `importsFake` und `FakeAudit`: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, verkörpert in `.claude/commands/implement-slice.md` Schritt 14 und im Reviewer-Skill) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (7×, verkörpert im Reviewer-Skill); für IN1 und IN2: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (9×) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (19×; beide verkörpert in `AGENTS.md` §3.10); für die Größenmarke (§4): `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (5 Dateien unter `evidence/`: `slice-if-match`, `slice-bootstrap-m4-konfig`, `slice-telemetry-einbau-gateway`, `slice-v2-bootstrap`, `slice-admin-imports`; die `state.md` nennt noch zwei Vorgänge, der Zähler folgt den Dateien, nicht der `state.md`) und `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×, offen; mit dem Block beim Start gesetzt, §6); außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (21×, verkörpert in den Übergangsregeln Nr. 8 und im Command-Guard). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (68×, `AGENTS.md` §3.17), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (73×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
