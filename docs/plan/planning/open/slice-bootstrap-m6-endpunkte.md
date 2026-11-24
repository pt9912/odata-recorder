# Slice slice-bootstrap-m6-endpunkte: Endpunkte der Admin-Slices im Composition Root: Dienste in jedem Modus, Routen, Einstellungen der Diagnose und Ersatz der vier Wächter (Teil A des Schnitts von `slice-bootstrap-m6`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-FA-18`, `LH-FA-09`, `LH-FA-10`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen** (Teil A des Schnitts von [slice-bootstrap-m6](slice-bootstrap-m6.md)): `LH-FA-13.a` · `SPEC-300` Nr. 5, 6 und 13 a, b (Admin-Listener im Composition Root, Architect 2026-11-23: Dienste in jedem Modus, Routen, Einstellungen der Diagnose, Ersatz der Wächter) · `SPEC-139` Nr. 1 bis 3, 11, 12 (`Handle`) · `SPEC-005` (die 17 Routen) · `SPEC-236` Nr. 7 (Allowlist der Diagnose) · `SPEC-238` Nr. 14, `SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13, `SPEC-298` Nr. 14 (die Zusage „das Binary bleibt unverändert“ endet mit der Verdrahtung; der Ersatz der vier Wächter, §3 Zeilen A02 bis A07) · `ARC-007`, `ARC-012` · `architecture.md` §2.5, §4.2, §4.5, §4.9.
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

**Ziel:** Der Composition Root verdrahtet die Endpunkt-Slices der Welle (`SPEC-300` Nr. 5, 6 und 13 a, b): der `Handle` trägt `Sessions()` in jedem Modus (ohne Recorder mit einem Stand-in für `Drain`, `Record()` bleibt `nil`), den Dienst der Inspektion mit den 18 Einstellungen der Diagnose aus den typisierten Werten und die Routen der vier Konstruktoren `SessionRouten`, `InspektionRouten`, `ImportRouten` (nur im Modus `record`) und `WorkspaceRouten` in der Folge von `SPEC-300` Nr. 5 c (17 Routen im Modus `record`, 15 in den anderen), je Konstruktor genau einmal, mit den einen Instanzen von `AuditLog`, `Telemetry`, Generator und Uhr; die vier Wächter `TestIM16BinaryUnveraendert`, `TestI18BinaryUnveraendert`, `TestS17BinaryUnveraendert` und `TestWS15BinaryUnveraendert` im Paket `admin` sind durch Fänger der Verdrahtung im Paket `bootstrap` ersetzt. Das Binary öffnet mit diesem Teil noch keinen Admin-Listener, und die Schlüssel `ADMIN_*` bleiben abgelehnt: keine neue Zusage am Prozess, `make gates` grün. **Schnitt vor dem Code** (Vorschlag des Architect, Planner, 2026-11-23): die Schätzung des ungeschnittenen Plans (*S* = 4366, Marke 5900, 40 Zeilen) liegt über der Schnittgrenze von `SPEC-128` Nr. 1; vier Teile, die nacheinander laufen und je einzeln lieferbar sind: **A** (diese Datei, Zeilen A01 bis A07), **B** [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md) (Admin-Listener, B01 bis B19), **C** [slice-bootstrap-m6-transport](slice-bootstrap-m6-transport.md) (Transportschutz, C01 bis C08) und **D** [slice-bootstrap-m6](slice-bootstrap-m6.md) (M6-Test, D01 bis D06; der Teil trägt die Kennung des ungeteilten Slice, so dass jede Zeile „`slice-bootstrap-m6` in `done/`“ in den Plänen davor „alle vier Teile in `done/`“ bedeutet).

**Übernimmt** (Plan-Defekt aus `slice-admin-tls`, Eintrag des Planners vom 2026-11-23 nach der Entscheidung des Architect zum Teil *Betrieb* dort, `AGENTS.md` §3.13; beim Schnitt vom 2026-11-23 an diesen Teil gelegt; §1 *Ausdrücklich NICHT* und DoD dieses Slice vorher gelesen, kein Ausschluss trifft die Sendung): vier Tests im Paket `admin` (`src/adapters/driving/admin/`) lesen alle `src/bootstrap/*.go`, auch jede `_test.go`, und melden jeden Import von `…/driving/admin` und jede Nennung der Routen-Konstruktoren: `TestIM16BinaryUnveraendert` (`ImportRouten`), `TestI18BinaryUnveraendert` (`InspektionRouten`), `TestS17BinaryUnveraendert` (`SessionRouten`) und `TestWS15BinaryUnveraendert` (`WorkspaceRouten`). Sie sind Fänger der Zusage „das Binary bleibt unverändert“ der Verträge `SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13 und `SPEC-298` Nr. 14. Mit der Verdrahtung dieses Teils (der Root importiert `admin` und ruft die vier Konstruktoren) endet diese Zusage, und die Wächter werden rot. Der Teil **ersetzt** sie durch Fänger der Verdrahtung, er streicht sie nicht ersatzlos und lockert keinen Fänger (`AGENTS.md` §3.6): was Nr. 14, 12, 13 und 14 der vier Verträge für das Binary aussagen, gilt dann als Zusage der Verdrahtung; Wortlaut und Fänger je Vertrag hat der Architect am 2026-11-23 geliefert (Zeilen A02 bis A07, Ort des Ersatzes: das Paket `bootstrap`, die vier Funktionen entfallen). Folge-Slice: [slice-admin-tls-betrieb](slice-admin-tls-betrieb.md) startet erst, wenn alle vier Teile in `done/` liegen (die Wächter fallen schon mit diesem Teil, die Bedingung ist die Reihenfolge der Dateien in `src/bootstrap`). Geber: [slice-admin-tls](../done/slice-admin-tls.md).

**Übernimmt außerdem** (Befund des Architect, 2026-11-23, beim Nachzug von `SPEC-139`; beim Schnitt an diesen Teil gelegt): der `SessionUseCase` besteht in jedem Modus (`SPEC-300` Nr. 5 a, Zeile A06), weil die Routen der Sessions `Sessions()` verlangen, das außerhalb von `record` `nil` war.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Admin-Listener: `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE` lesen und prüfen, die Token-Datei, Bindung beider Listener, Herunterfahren, Fehlerkanal, `GET /healthz` am Prozess, Zelle `admin_abgewiesen`, Handbuch und Katalog zu den zwei Schlüsseln und den zwölf Codes — [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md) (Teil B, Zeilen B01 bis B19; §1 und DoD dort tragen die Sendungen). Dieser Teil bindet keinen Port; ohne Listener sind die Routen am `Handle` über die Naht der Tests erreichbar.
- Die Schlüssel `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT`, die vier Aufrufe von `SPEC-299` Nr. 8, die Warnung und die Umhüllung — [slice-bootstrap-m6-transport](slice-bootstrap-m6-transport.md) (Teil C, Zeilen C01 bis C08).
- Der M6-Test (Rundlauf über beide Listener, Abweisung ohne Zustandsänderung, Zellen Modus × Art, Stufe 11 `TestE2EAdmin` mit den Negativen von `LH-FA-14`) — [slice-bootstrap-m6](slice-bootstrap-m6.md) (Teil D, Zeilen D01 bis D06).
- `HYBRID_*`, `FALLBACK_*` und der Modus `hybrid` — geliefert von `slice-bootstrap-m4-konfig` und `slice-bootstrap-m4` (beide in `done/`); dieser Teil fährt `hybrid` als einen der vier Modi jeder Menge. `LISTEN_TLS_*` — geliefert von `slice-listener-tls-verdrahtung` (`done/`).
- Die Begrenzung der Fehlversuche (`LH-FA-13.a` nannte keine) — entschieden (User, 2026-11-24, Variante A, `SPEC-301`); Träger ist der Folge-Slice [slice-admin-abweisen-begrenzen](slice-admin-abweisen-begrenzen.md), Vorlage in §6 von [slice-bootstrap-m6](slice-bootstrap-m6.md).
- Der Handshake des Admin-Adapters über TLS (Zeilen AT11 bis AT17 von `slice-admin-tls`) — [slice-admin-tls-betrieb](slice-admin-tls-betrieb.md), startet nach allen vier Teilen.
- Neue Endpunkte oder Fachlogik — der Teil verdrahtet, was die Slices der Welle liefern. Das Umschalten des Modus zur Laufzeit — `LH-FA-13` schließt es aus.
- Ein Handbuch-Kapitel: das Binary ändert sich mit diesem Teil nicht (kein Listener, keine Schlüssel); Handbuch und Katalog zum Admin-Listener tragen Teil B und Teil C.

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

- [ ] Dienste und Routen im Root: `Sessions()` in jedem Modus (Stand-in für `Drain` ohne Recorder, `Record()` und `OpenInitialSession` nur in `record`), `Inspektion()` mit den 18 Einstellungen in der Kanonform und die Routen der vier Konstruktoren in der Folge Sessions, Inspektion, Imports, Workspaces (17 Routen in `record`, 15 sonst; `ImportRouten` nur in `record`), je Konstruktor genau einmal, mit den einen Instanzen von `AuditLog`, `Telemetry`, Generator und Uhr; die drei Zeilen, die `Sessions()` außerhalb von `record` als `nil` fordern (`wurzel_test.go` Zeile 210, `m3_wurzel_test.go` 154, `m4_wurzel_test.go` 221), fordern `nicht nil` (Zeilen A01, A03, A06, A07 der Tabelle in §3, vom Architect).
- [ ] Ersatz der vier Wächter `TestIM16BinaryUnveraendert`, `TestI18BinaryUnveraendert`, `TestS17BinaryUnveraendert` und `TestWS15BinaryUnveraendert` (Paket `admin`) durch Fänger der Verdrahtung im Paket `bootstrap`, nicht ersatzlos gestrichen: der Ersatz fängt, dass der Root `admin` und die vier Routen-Konstruktoren nennt, wo die Verträge `SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13 und `SPEC-298` Nr. 14 es zusagen, und das Verhalten am `Handle` (Zeilen A02 bis A05 und A07, Geber `slice-admin-tls`, §1 *Übernimmt*); `make test` ist grün.
- [ ] Zwischenstand benannt: das Binary öffnet keinen Admin-Listener, `ADMIN_*` bleiben „nicht unterstützt“, obwohl `SPEC-110` Nr. 2 sie als geprüft führt (die Spezifikation steht vor dem Code, wie bei jedem Vertrag); die Verdrahtung des Listeners liefert Teil B ([slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md)).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-300`, `SPEC-139`, `SPEC-292`, `SPEC-293`, `SPEC-297`, `SPEC-298` stehen vor dem Start auf `main`; dieser Teil schreibt keine Spezifikation).
- [ ] Handbuch: der Teil führt keinen Fehlercode, keinen Schlüssel und kein Kommando ein und ändert das Binary nicht (kein Listener); das Kataloge-Gate ist grün, und §7 sagt, dass kein Fehlercode, Schlüssel oder Kommando des Binary dazukam.
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
| `src/bootstrap/` (`wurzel.go`, neu `wurzel_admin.go`) | update | Dienste (`SessionUseCase` in jedem Modus mit Stand-in für `Drain`), Routen der vier Konstruktoren, Einstellungen der Diagnose aus den typisierten Werten (`SPEC-300` Nr. 5, 6) |
| `src/bootstrap/` (Tests: `export_test.go`, `wurzel_test.go`, `m3_wurzel_test.go`, `m4_wurzel_test.go`, neu `admin_routen_test.go`, `admin_sessions_test.go`, `admin_inspektion_test.go`, `admin_imports_test.go`, `admin_workspaces_test.go`, `admin_konstruktoren_test.go`) | update und neu | Naht (Zugriff auf Routen und Dienste des `Handle`); die drei Zeilen, die `Sessions()` außerhalb von `record` als `nil` fordern (`wurzel_test.go` Zeile 210, `m3_wurzel_test.go` 154, `m4_wurzel_test.go` 221), fordern jetzt `nicht nil` (`SPEC-300` Nr. 5 a, Zusage geändert vom Architect, kein Fänger gelockert: die Aussage ist eine andere); die Fänger der Zeilen A01 bis A07 |
| `src/adapters/driving/admin/imports_quelltext_test.go`, `inspektion_quelle_test.go`, `sessions_quelle_test.go`, `workspaces_quelltext_test.go` | update (die vier Funktionen entfallen) | die Wächter `…BinaryUnveraendert` entfallen im Commit ihres Ersatzes im Paket `bootstrap` (Zeilen A02 bis A07, Geber `slice-admin-tls`); die Pfade stehen einzeln im Block (`SPEC-128` Nr. 10 g) |
| `spec/spezifikation.md`, `harness/vertraege/SPEC-037-test-integration.md` | geliefert (Architect, 2026-11-23) | `SPEC-300` neu; Nachzug von `SPEC-139`, `SPEC-238` Nr. 14, `SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13, `SPEC-298` Nr. 14; dieser Teil schreibt sie nicht |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Nachzug in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Sie stehen unten (Architect, 2026-11-23, Zeilen A01 bis A07 der Tabelle des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; jede Zeile steht in genau einem der vier Teile); die Zusagen heißen mit der Nummer von `SPEC-300` und den Verträgen der Endpunkte.

**Entscheidungen des Architect** (2026-11-23; die Teile des ungeschnittenen Plans, die Endpunkte, Dienste und Wächter betreffen; Auftrag aus dem ungeschnittenen Plan → Ort in der Spezifikation):

| Auftrag | Entscheidung | Ort |
|---|---|---|
| 4 Verdrahtung der Ports | `AuditLog`, `Telemetry`, Generator und Uhr sind die einen Instanzen des Prozesses; die Einstellungen der Diagnose bildet der Root aus den typisierten Werten (18 Schlüssel, Kanonform); der Dienst sieht nie die Umgebung | `SPEC-300` Nr. 5 d, 5 e, 6 |
| 7 Ersatz der vier Wächter | die vier Funktionen entfallen; der Ersatz liegt im Paket `bootstrap` am `Handle` (ein Test im Paket `admin` kann das Binary nicht fahren und las nur Quelltext); je Vertrag eine Zeile (A02 bis A05) und die Zeile A07 für „genau einmal“. **Keine Lockerung nach `AGENTS.md` §3.6:** die Zusage „das Binary bleibt unverändert“ endet durch die Änderung der Verträge selbst (Spezifikation, Rang 2); jede Zusage, die bleibt (Routen, Ports, Einträge, Modus), hat einen Fänger mit Mutation, der mehr führt als der Wächter (Verhalten am `Handle` statt Namen im Quelltext) | `SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13, `SPEC-298` Nr. 14, `SPEC-238` Nr. 14, `SPEC-299` Nr. 14 |
| neu: `SessionUseCase` in jedem Modus | die Routen der Sessions stehen in allen Modi (`SPEC-005`, `LH-FA-13.a`; das Kriterium „keine Sessions, leere Liste“ ist im Modus `record` nie zu erreichen); ohne Recorder ersetzt ein Stand-in `Drain` | `SPEC-300` Nr. 5 a; `SPEC-139` Nr. 1, 3, 11, 12 |
| neu: Imports nur im Modus `record` | `Import()` bleibt außerhalb von `record` `nil`, die zwei Routen fehlen dort (404 `ODR-E2102` mit Token) | `SPEC-300` Nr. 5 c; `SPEC-139` *Ungeprüft* |
| Frage von `slice-test-security` (Ort des Servers) | **der Test bleibt im Paket `admin`** (echter Loopback-Listener, die vier Konstruktoren, die Fakes des Pakets): weder Wächter noch Kante des Gates berührt; was am Root zu belegen ist, tragen Zeile A01 dieses Teils und die Zeile D02 von [slice-bootstrap-m6](slice-bootstrap-m6.md). Der Trigger von `slice-test-security` wird **nicht** um diese Slices erweitert: das schlösse den Zyklus mit seiner Bedingung für den ersten Teil | §6 von `slice-test-security` |

**Zusage · Fänger** (`AGENTS.md` §3.15; Architect, 2026-11-23; `Nr.` ist die Kennung des ungeschnittenen Plans, `A` Endpunkte im Root, `B` Admin-Listener, `C` Transportschutz, `D` M6-Test; die Spalte *Zusage* nennt den Ort in `SPEC-300` und den Verträgen; die Tests laufen im Paket `bootstrap_test` ohne Netz außer dem Loopback des Prozesses, Pfad und Namen der Naht in `export_test.go` bestimmt der Implementer, die Mengen stehen mit Achsen unten; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10; **nicht in der Tabelle und nicht in der Marke:** die Begrenzung der Fehlversuche, Ausgang *entschieden* (Folge-Slice [slice-admin-abweisen-begrenzen](slice-admin-abweisen-begrenzen.md), §6 von [slice-bootstrap-m6](slice-bootstrap-m6.md))):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| A01 | (`SPEC-300` Nr. 5 c, `SPEC-005`) Der Root hängt die Routen der vier Konstruktoren in der Folge Sessions, Inspektion, Imports, Workspaces an `Konfig.Routen`: 17 Routen im Modus `record`, 15 in `replay`, `mock` und `hybrid` (ohne `POST /__admin/imports` und `GET /__admin/imports/{id}`) | `TestAdminRouten`: Menge M-Route × M-Modus, 17 × 4 = 68 Zellen; Orakel je Zelle als Literal (vorhanden oder nicht), über die Naht `AdminRouten(h)`; Zählprüfung: 17 Literale gleich den Zeilen von `SPEC-005` ohne `GET /healthz` (4 + 3 + 2 + 8) und die Folge der Konstruktoren | `WorkspaceRouten` nicht gehängt (rot: 32 Zellen); `InspektionRouten` nicht (12); `SessionRouten` nur im Modus `record` (12); `ImportRouten` in jedem Modus (rot: `Build` scheitert mit `model.ErrUngueltig`, 3 Modi); `ImportRouten` nie (2 Zellen); Folge vertauscht (rot) | andere |
| A02 | (`SPEC-292` Nr. 14, ersetzt `TestS17BinaryUnveraendert`) Der Root ruft `SessionRouten` genau einmal je `Build`, in jedem Modus, mit `Handle.Sessions()` und der einen Instanz von `AuditLog`; die vier Routen wirken auf die Ablage des Prozesses | `TestAdminSessionsVerdrahtet`: je Modus die vier Handler über die Naht (`POST` legt an, `GET` liest, `close` beendet, `DELETE` löscht), Orakel die Ablage (`db.Sessions()`) und die gelesene `audit.jsonl` (`session_angelegt` mit `service_id` und `session_id`, `session_beendet`, `session_geloescht`): 4 Routen × 4 Modi = 16 Zellen | `audit` nil (rot: `Build` scheitert, 4 Modi); `Sessions()` nur im Modus `record` übergeben (rot: 12 Zellen); der Handler von `close` ruft `Delete` (rot: 4); zweiter Aufruf des Konstruktors (rot in A07) | andere |
| A03 | (`SPEC-293` Nr. 12, ersetzt `TestI18BinaryUnveraendert`) `InspektionRouten` genau einmal je `Build`, mit dem Dienst aus `inspection.New` (Ports der Ablage, `Zaehler` der eine Adapter, Modus, Einstellungen nach `SPEC-300` Nr. 6); die Diagnose zeigt Modus, die 18 Einstellungen der Allowlist in der Kanonform und die Zähler des einen Adapters | `TestAdminInspektionVerdrahtet`: je Modus `diagnostics` (`Modus` gleich dem des `Handle`, `Einstellungen` gleich dem Literal von M-Einstellung, `Fehler` hat je Code der Tabelle einen Eintrag, der Zähler steigt nach `CountError` des Adapters aus der Naht) und im Modus `record` `exchanges` und `coverage` einer aufgezeichneten Session und Baseline; M-Einstellung: 18 Schlüssel × 2 Konfigurationen (Defaults; jeder Wert abweichend, Listen mit zwei Einträgen, Dauer `30d`, Größe `1GiB`) = 36 Zellen, Zählprüfung 18 gleich `len(inspection.Erlaubt())`; Canary im Wert von `ADMIN_TOKEN_FILE`, `LISTEN` und `UPSTREAM_URL` steht in keiner Antwort | `Zaehler` ein zweiter Adapter (rot: Zähler bleibt 0, 4 Modi); je Schlüssel eine Einstellung fehlt (18 Mutationen, je rot in 2 Zellen); `LISTEN` aufgenommen (rot); `UPSTREAM_TLS_INSECURE_SKIP_VERIFY` nicht abgeleitet (rot: 2 Zellen); Dauer ohne `ms` (4 Schlüssel, rot: 8 Zellen); leere Menge nicht `none` (rot); Modus fest `record` (rot: 3 Modi) | andere |
| A04 | (`SPEC-297` Nr. 13, ersetzt `TestIM16BinaryUnveraendert`) `ImportRouten` genau einmal je `Build` und nur im Modus `record`, mit `Handle.Import()` und der einen Instanz von `AuditLog`; in den drei anderen Modi wird der Konstruktor nicht gerufen, und `Build` gelingt mit `Import()` gleich `nil` | `TestAdminImportsVerdrahtet`: im Modus `record` der Handler von `POST` (Import einer beendeten Session: Baseline und Report in der Ablage, Zeile `import_veroeffentlicht` mit `import_id`) und von `GET {id}` (derselbe Body); in `replay`, `mock` und `hybrid` `Build` ohne Fehler und keine Route (A01): 2 + 3 = 5 Zellen | Konstruktor in jedem Modus (rot: `Build` scheitert, 3 Zellen); nie gerufen (rot: 2 Zellen); `audit` nil (rot); der Eintrag fehlt (rot) | andere |
| A05 | (`SPEC-298` Nr. 14, ersetzt `TestWS15BinaryUnveraendert`) `WorkspaceRouten` genau einmal je `Build`, in jedem Modus, mit `Workspaces()`, `Snapshots()` und der einen Instanz von `AuditLog`; die acht Routen wirken auf die Ablage | `TestAdminWorkspacesVerdrahtet`: je Modus die acht Handler über die Naht (Anlegen aus einer gesäten Baseline, Lesen, Auflisten, Snapshot erzeugen und auflisten, `reset` mit Revision n+1, `restore`, Löschen), Orakel die Ablage und die Zeilen der Arten `workspace_angelegt`, `snapshot_erzeugt`, `workspace_zurueckgesetzt`, `workspace_wiederhergestellt` und `workspace_geloescht` (vom Dienst): 8 Routen × 4 Modi = 32 Zellen | `Snapshots()` nil (rot: `Build` scheitert, 4); Konstruktor nur im Modus `record` (rot: 24 Zellen); `reset` ruft `Restore` (rot: 4) | andere |
| A06 | (`SPEC-300` Nr. 5 a) Der `SessionUseCase` besteht in jedem Modus; ohne Recorder (`replay`, `mock`, `hybrid`) ersetzt ein Stand-in `Drain` (leerer Bericht, `nil`), und `Record` ist `model.ErrUngueltig`; `Record()` bleibt `nil`, es entsteht weder Recorder noch Pipeline, und `OpenInitialSession` läuft nur im Modus `record` | `TestAdminSessionsAlleModi`: je Modus `Sessions()` ungleich `nil`, `Record()` nur im Modus `record` ungleich `nil`, `List` auf frischer Ablage leer (`LH-FA-13`, Boundary), `Create`, `Beenden` (`complete`), `Delete`; die Haken `NeuerRecorder` und `NeuePipeline` zählen 0 außerhalb von `record`: 4 Modi × 5 Beobachtungen = 20 Zellen | das Stand-in liefert einen Fehler aus `Drain` (rot: `Beenden`, 3 Modi); `Sessions()` außerhalb von `record` `nil` (rot: 3); Recorder in jedem Modus (rot: der Zähler, 3); `OpenInitialSession` in jedem Modus (rot: `List` nicht leer, 3) | andere |
| A07 | (`SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13, `SPEC-298` Nr. 14) Die vier Wächter entfallen im Commit ihres Ersatzes; „genau einmal“ gilt je Konstruktor im Quelltext von `src/bootstrap/` ohne `_test.go` | `TestAdminKonstruktorenEinmal`: Syntaxbaum von `src/bootstrap/*.go`, Menge der 4 Konstruktoren, je genau eine Aufrufstelle; Probe: ein Literal mit zwei Aufrufen zählt 2; kein Test im Paket `admin` liest `src/bootstrap/` (Zählprüfung 0 Treffer von `"bootstrap"` im Pfad) | ein zweiter Aufruf je Konstruktor (4 Mutationen, je rot); ein Wächter bleibt stehen (rot: `make test`, weil der Import von `admin` ihn rot macht) | andere |

Die Zeilen A01 bis A07 sind 7: *a* = 0, *b* = 7. **Mengen mit Achsen** (Orakel als Literal, nie aus dem Code und nie als Aufruf des Codes; die Zahl der Zellen steht im Test als Literal und wird gegen die gezählten verglichen, `AGENTS.md` §3.15): M-Modus (`record`, `replay`, `mock`, `hybrid`, 4); M-Route (17 Zeilen von `SPEC-005`, Achsen Methode, Pfad mit und ohne Parameter, Modus); M-Einstellung (18 Schlüssel, Achsen Art des Wertes {Aufzählung 5, Wahrheitswert 2 plus 1 abgeleitet, Größe 1, Dauer 4, Liste 5}, Konfiguration {Default, abweichend}). Die Mengen M-Flag, M-Weg (Teil C), M-Bind, M-Abweisung (Teil B) und M-Art (Teil D) stehen in den Plänen der Teile, die sie führen.

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

**Start** (`next` → `in-progress`): `slice-admin-listener-token` samt seinen Teilen `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` (Schnitt vor dem Code, 2026-11-07), [slice-admin-tls](../done/slice-admin-tls.md), `slice-inspection-use-case`, `slice-inspection-coverage-diagnose`, `slice-admin-sessions` samt seinen Teilen `slice-admin-sessions-beenden` und `slice-admin-inspektion` (Schnitt vor dem Code, 2026-11-21), `slice-admin-imports` samt seinem Teil `slice-admin-imports-audit` (Schnitt vor dem Code, 2026-11-22), `slice-admin-workspaces` samt seinem Teil [slice-admin-workspaces-audit](../next/slice-admin-workspaces-audit.md) (Schnitt vor dem Code, 2026-11-22), [slice-test-security](slice-test-security.md), `slice-bootstrap-m2`, `slice-m2-referenzdienst`, `slice-bootstrap-m3`, `slice-audit-log`, `slice-telemetry-port`, [slice-telemetry-einbau-record](../done/slice-telemetry-einbau-record.md), [slice-telemetry-einbau-hybrid-port](../in-progress/slice-telemetry-einbau-hybrid-port.md) und [slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md) (beide Teile des Schnitts vom 2026-11-23) liegen in `done/`. Die drei letzten ändern `src/bootstrap/wurzel.go`, ebenso `slice-admin-tls` (`export_test.go` und neue Dateien): die Reihenfolge ist `slice-telemetry-einbau-record`, `slice-telemetry-einbau-hybrid-port`, `slice-telemetry-einbau-hybrid`, dann dieser Teil (Architect, 2026-11-23: Teil 2 des Hybrid-Schnitts ändert eine Zeile in `wurzel.go`, dieser Teil viel; liefe er danach, zählte sein Block die Zeilen von diesem Teil mit, `SPEC-128` Nr. 10 g). Die Spezifikation trägt den Nachzug von `SPEC-110` und den Vertrag `SPEC-300` (geliefert am 2026-11-23) samt Tabelle und Raster, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Teil B startet erst, wenn dieser Teil in `done/` liegt, Teil C nach Teil B, Teil D ([slice-bootstrap-m6](slice-bootstrap-m6.md)) nach Teil C.

**Warum:** der Root verdrahtet alle Adapter-Slices der Welle (*Code-Abhängigkeit*), die Ports `AuditLog` und `Telemetry` (*Code-Abhängigkeit*) und den Record- und den Mock-Pfad (`slice-bootstrap-m2`, `slice-bootstrap-m3`: *Code-Abhängigkeit*). `slice-test-security` liegt vor diesem Teil (`welle-admin-api` §5; *Reihenfolge*, keine Typ-Abhängigkeit; der Trigger von `slice-test-security` nennt diese Teile nicht, weil das den Zyklus schlösse). Nicht gebraucht: `welle-hybrid`, `slice-bootstrap-m4-konfig` und `slice-bootstrap-m4` (keine Kante); sie ändern `src/bootstrap/` und `SPEC-110` Nr. 2 und liegen in `done/`.

**Größenmarke** (`SPEC-128` Nr. 1 und 10 g, `AGENTS.md` §3.14; Teil des Architect 2026-11-23 im Block unten; der Planner hat die Zahlen am 2026-11-23 aus der Tabelle nachgerechnet, **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 7 Zeilen der Tabelle (A01 bis A07), *a* = 0, *b* = 7, *f* = 0 (echte Dienste an der Naht, kein neuer Port-Fake und kein Test-Server), *p* = 200 (Dienste, Stand-in, Einstellungen, Routen, Zugänge): *S* = 85·0 + 80·7 + 200·0 + 200 = 760, 135·760 = 102600, **Marke 1100** = 100 · ⌈10,26⌉; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`; der Teil liegt 1463 unter der Schnittgrenze, jede weitere Zeile kostet 80. Summe der vier Teile: 760 + 2190 (B) + 730 (C) + 686 (D) = 4366, gleich dem ungeschnittenen Plan; keine Abweichung vom Vorschlag des Architect. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen.

```groessenmarke
pfad: src/bootstrap/wurzel.go
pfad: src/bootstrap/wurzel_admin.go
pfad: src/bootstrap/export_test.go
pfad: src/bootstrap/admin_routen_test.go
pfad: src/bootstrap/admin_sessions_test.go
pfad: src/bootstrap/admin_inspektion_test.go
pfad: src/bootstrap/admin_imports_test.go
pfad: src/bootstrap/admin_workspaces_test.go
pfad: src/bootstrap/admin_konstruktoren_test.go
pfad: src/bootstrap/wurzel_test.go
pfad: src/bootstrap/m3_wurzel_test.go
pfad: src/bootstrap/m4_wurzel_test.go
pfad: src/adapters/driving/admin/imports_quelltext_test.go
pfad: src/adapters/driving/admin/inspektion_quelle_test.go
pfad: src/adapters/driving/admin/sessions_quelle_test.go
pfad: src/adapters/driving/admin/workspaces_quelltext_test.go
produktcode: 200
fakes: 0
```

Der Block nennt unter `src/bootstrap/` und unter `src/adapters/driving/admin/` die Dateien einzeln (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`); die Dateien der Dokumente und der Spezifikation stehen nicht im Block (`SPEC-128` Nr. 3 m). Die Dateien der Naht und der neuen Tests gibt es noch nicht (die Namen sind ein Vorschlag des Architect; die Messung zählt dort 0, bis sie entstehen); ein anderer Name ersetzt den Vorschlag im selben Commit, und der Implementer nennt jede weitere Datei, die er ändert oder anlegt, im selben Commit im Block.

**Kollisionen** (Architect, 2026-11-23, gelesen an den Zweigen und den Plänen, vom Planner beim Schnitt übernommen; `SPEC-128` Nr. 10 g: die Messung zählt jede Änderung an den Pfaden zwischen `basis` und `HEAD`): (1) [slice-admin-tls](../done/slice-admin-tls.md) schreibt `src/bootstrap/export_test.go` (Naht), `admintls.go`, `admintls_test.go` und `.a-check.yml`; er liegt vor diesem Teil in `done/` (Start-Bedingung), dieser Teil erweitert dieselbe `export_test.go` auf `main` nach seinem Merge. (2) `slice-telemetry-einbau-record` und die zwei Teile des Hybrid-Schnitts ändern `wurzel.go`: sie laufen nacheinander vor diesem Teil (Start-Bedingung). (3) [slice-test-security](slice-test-security.md) schreibt neue Dateien (`src/bootstrap/admin_sicherheit_test.go`, ein Test im Paket `admin`), keine Datei dieses Teils; er liegt vor diesem Teil. (4) Die Gate-Kette (Nr. 7 der Übergangsregeln): kein Pfad aus `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`; der Planner prüft es beim Start.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein weiterer Schnitt ist nicht vorgesehen.
- `in-progress` → `open` (blockiert — Carveout?): ein Slice der Welle liefert nicht den Dienst, den die Routen brauchen (siehe §6), oder der Ersatz der Wächter lässt sich nicht ohne Lockerung eines Fängers schreiben (`AGENTS.md` §3.6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23; die Zeile dieses Teils; die Zeilen der Teile B, C und D stehen in deren Plänen; jede Zelle nennt ihren Ort in `SPEC-300`, den Verträgen davor oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Bilden der Einstellungen der Diagnose | jeder der 18 Schlüssel hat einen Eintrag, aus dem wirksamen Wert (`SPEC-300` Nr. 6) | die leere Menge einer Liste ist `none` (`SPEC-300` Nr. 6) | ein Schlüssel außerhalb der Allowlist nie (`SPEC-300` Nr. 6, `SPEC-236` Nr. 7) | gilt nicht: synchron in `Build`, ohne Kontext (`SPEC-139` Nr. 2) | der Dienst liefert je Aufruf eine Kopie (`SPEC-236` Nr. 7) |

**Zusatztabelle** (`AGENTS.md` §3.16: die Operationen hängen vom Modus ab; je Modus eine Zeile; die Paare von Verstößen stehen in den Plänen der Teile B und C):

| Paar oder Modus | Ergebnis | Fänger |
|---|---|---|
| jeder Modus (`record`, `replay`, `mock`, `hybrid`) | derselbe Weg, dieselben Verstöße, dieselbe Warnung; Routen der Imports nur in `record` (`SPEC-299` Nr. 3, `SPEC-300` Nr. 3, 5 c) | A01 und alle Mengen mit der Achse Modus; C07 (Teil C) |

- Die Wächter `TestIM16BinaryUnveraendert`, `TestI18BinaryUnveraendert`, `TestS17BinaryUnveraendert` und `TestWS15BinaryUnveraendert` (Paket `admin`) lesen alle `src/bootstrap/*.go`, auch `_test.go`, und färben mit der Verdrahtung `make test` rot (Geber `slice-admin-tls`, Plan-Defekt dort; `AGENTS.md` §3.13); der Plan ersetzt sie durch Fänger der Verdrahtung, seine Tabelle *Zusage · Fänger* trug die Zeilen nicht — **Ausgang:** entschieden (Architect, 2026-11-23, Auftrag 7): der Wortlaut der vier Verträge nach der Verdrahtung steht in `SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13 und `SPEC-298` Nr. 14, je Vertrag eine Zeile mit Fänger und Mutation (A02 bis A05) und die Zeile A07; Ort des Ersatzes ist das Paket `bootstrap`, die vier Funktionen entfallen im selben Commit. Keine Lockerung nach `AGENTS.md` §3.6 und keine Entscheidung des User: die Zusage endet durch die Änderung der Verträge, jede bleibende Zusage hat einen stärkeren Fänger (Verhalten am `Handle` statt Namen im Quelltext). Die Wirkung für `slice-admin-tls-betrieb`: er startet, wenn alle vier Teile in `done/` liegen (Planner, 2026-11-23); die Wächter fallen schon mit diesem Teil.
- `SessionUseCase` in jedem Modus (Architect, 2026-11-23; Befund beim Nachzug von `SPEC-139`: die Routen der Sessions verlangen `Sessions()`, das außerhalb von `record` `nil` war) — **Ausgang:** entschieden: der Dienst entsteht in jedem Modus, ohne Recorder mit einem Stand-in für `Drain` (`SPEC-300` Nr. 5 a, Zeile A06); die drei Tests, die `Sessions()` außerhalb von `record` als `nil` fordern, ändern ihre Aussage (§3, Tabelle der Pfade); die Folge im Betrieb steht als akzeptiertes Negativ in der Zeile B18 von [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md).
- Imports nur im Modus `record`: `Import()` bleibt außerhalb von `record` `nil`, die zwei Routen fehlen dort — **Ausgang:** entschieden (Architect, 2026-11-23, `SPEC-300` Nr. 5 c, `SPEC-139` *Ungeprüft*): 404 `ODR-E2102` mit Token; Zeilen A01 und A04.
- Verdrahtung der Ports und Einstellungen der Diagnose (Auftrag 4, Architect, 2026-11-23): `AuditLog`, `Telemetry`, Generator und Uhr sind die einen Instanzen des Prozesses, die 18 Einstellungen bildet der Root aus den typisierten Werten, der Dienst sieht nie die Umgebung — **Ausgang:** entschieden (`SPEC-300` Nr. 5 d, 5 e, 6; Zeilen A03 und A06).
- Gemeinsame Dateien in `src/bootstrap/` mit `slice-telemetry-einbau-record`, dem Hybrid-Schnitt und [slice-admin-tls](../done/slice-admin-tls.md) (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`) — **Ausgang:** entschieden (Planner, 2026-11-23): Reihenfolge `slice-telemetry-einbau-record`, `slice-telemetry-einbau-hybrid-port`, `slice-telemetry-einbau-hybrid`, dann dieser Teil (§4); `slice-admin-tls` liegt in `done/`; keine parallelen Arbeitsbäume auf `wurzel.go`.
- Frage von `slice-test-security` (Ort des Servers) — **Ausgang:** entschieden (Architect, 2026-11-23): der Test bleibt im Paket `admin`; was am Root zu belegen ist, trägt Zeile A01 (Routen) dieses Teils und Zeile D02 ([slice-bootstrap-m6](slice-bootstrap-m6.md)); der Planner hat in `slice-test-security` den Pfad im Block einzeln und die Zeile `tests/` nachgezogen.
- `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×): der Admin-Pfad im Composition Root erweitert die Pakete, deren Laufzeit Fälle der Gegenproben mitfahren — **Ausgang:** weiter offen → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` im Register.
- Größe: die Tabelle *Zusage · Fänger* liegt vor, die Schätzung des ungeschnittenen Plans (4366) liegt über der Schnittgrenze — **Ausgang:** entschieden (Planner, 2026-11-23, nach dem Vorschlag des Architect): Schnitt in vier Teile mit den Marken 1100 (dieser Teil), 3000, 1000 und 1000, nachgerechnet aus der Tabelle (keine Abweichung); `basis`, `schaetzung` und `marke` setzt der Planner je Teil beim Start (§4).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-23; der Planner zieht Treffer und Zähler beim Start nach). Offen und für den Teil einschlägig: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×), `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×, `src/bootstrap/`, §4) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (22×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (43×, `AGENTS.md` §3.9; der Nachzug der drei Zeilen zu `Sessions()` und die Tests, die die Zahl der Dienste festhalten, sind Stellen) und `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16). Keiner erreicht mit diesem Teil offen 3× ohne Träger.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
