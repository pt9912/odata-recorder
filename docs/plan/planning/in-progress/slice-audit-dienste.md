# Slice slice-audit-dienste: Audit-Einträge der vier Aufrufstellen, `FakeAudit` und das Feld `Audit` der Dienste

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-RB-04`, `LH-FA-13`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)

**Berührte Spec-Stellen:** `LH-RB-04.a` („Löschungen werden im Audit-Log festgehalten“) · `LH-FA-13.a` (Audit-Eintrag ohne Tokenwert) · `SPEC-234` Nr. 4 bis 6 (wann ein Dienst schreibt, Reihenfolge, Fehler, Abbruch, Blobs ohne Eintrag), Nr. 8 (Konfig der Dienste), Nr. 9 (Fake), Nr. 10 (Wiederholung, Nebenläufigkeit) · `SPEC-123` Nr. 1, 6 (Aufbewahrung beim Start) · `SPEC-124` Nr. 1, 7 (Löschen einer Session) · `SPEC-127` Nr. 1, 4 (Import, Schritt m) · `SPEC-129` Nr. 1, 9, 15 (Workspace löschen) · `SPEC-004` Nr. 2 (ein fehlgeschlagener Import „wird im Audit-Log festgehalten“) · `SPEC-005` (Schlusszeile) · `SPEC-109` Nr. 6 (Kontextfehler) · `ARC-021` · `architecture.md` §Support-Ports.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-audit-dienste)

**Autor:** pt9912. **Datum:** 2026-11-06.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel (Teil E des Schnitts von `slice-audit-log` in drei Teile):** Die vier Dienste halten fest, was sie gelöscht haben: `SessionUseCase.Delete`, die Löschung der Aufbewahrung in `MaintenanceUseCase.Startup` (Stufe Aufbewahrung, `SPEC-123` Nr. 6) und `WorkspaceUseCase.Delete` schreiben je gelöschtem Gegenstand genau einen Eintrag über den Port `AuditLog` (Art, Gegenstand, ohne Geheimnisse, `LH-RB-04.a`), und `ImportUseCase.Import` schreibt bei einem fehlgeschlagenen Import den Eintrag `import_fehlgeschlagen`, damit seine `import_id` wiederverwendbar bleibt (`SPEC-004` Nr. 2). Jede der vier Konfigs führt dafür das Feld `Audit support.AuditLog` (`SPEC-234` Nr. 8); der Eintrag folgt der Löschung, sein Fehler und ein beendeter Kontext ändern das Ergebnis des Dienstes nicht (`SPEC-234` Nr. 5). Ein Fake `FakeAudit` im Paket `supporttest` hält jedes Argument fest und zählt jede Operation (`SPEC-234` Nr. 9). Port, `Eintrag`, `Art` und `Arten()` liefert `slice-audit-log`; dieser Slice verdrahtet nichts und ruft den Adapter nie.

**Übernimmt (Sendung von `slice-session-use-case`, §1 *Ausdrücklich NICHT*, Stand `done/`; Geber der Weitergabe: `slice-audit-log`):** der Audit-Eintrag der Löschung einer Session in `SessionUseCase.Delete`; die Löschung selbst liegt dort.

**Übernimmt (Sendung von `slice-session-wartung`, §1 *Ausdrücklich NICHT*, Stand `done/`; Geber der Weitergabe: `slice-audit-log`):** der Audit-Eintrag der Löschung durch die Aufbewahrung in `MaintenanceUseCase.Startup`; die Löschung selbst liegt dort. Beide Pläne liegen in `done/` und sind nicht änderbar; die Übernahme steht hier.

**Übernimmt (Zuweisung von `slice-audit-log`, ursprünglicher Umfang):** der Eintrag der Löschung eines Workspace in `WorkspaceUseCase.Delete` (`slice-workspace-use-case` liegt in `done/`; sein §1 schließt den Eintrag nicht aus) und der Eintrag eines fehlgeschlagenen Imports in `ImportUseCase` (`slice-baseline-import` in `done/`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Port `AuditLog`, `Eintrag`, `AuditArt`, `Arten()`, der Adapter `audit.Datei` und die Datei `audit.jsonl` — `slice-audit-log`. Grund: Naht des Schnitts; dieser Slice braucht den Port und die Typen, nie den Adapter, und prüft gegen `FakeAudit`.
- Die Verdrahtung im Composition Root, die Datei im Datenverzeichnis und der Canary-Test über das `Handle` — `slice-audit-verdrahtung`. Grund: `src/bootstrap/` ändert auch `slice-bootstrap-m2`; die Verdrahtung läuft nach den Diensten und nach ihm.
- Audit-Einträge der Admin-Schreibendpunkte (Anlegen, Beenden, Reset, Restore über HTTP, die Anfrage eines Import-`POST`) und der Eintrag zu `ODR-E2101` — `welle-admin-api` (`SPEC-005`, `LH-FA-13.a`); die Nehmer `slice-admin-sessions`, `slice-admin-workspaces`, `slice-admin-imports` und `slice-admin-listener-token` tragen die Entscheidung `SPEC-234` Nr. 7 (nur der Use Case schreibt die vier Arten, der Adapter nie dieselben). Dieser Slice schreibt die vier Arten und nur sie.
- Die Löschung der Blobs nach dem Commit (best effort, `SPEC-123` Nr. 8) — **entschieden: kein Eintrag**, weder eigener noch als Teil des Eintrags der Löschung (`SPEC-234` Nr. 6, Grund und Folge dort).
- Lesen, Abfragen und Exportieren des Audit-Logs — kein Endpunkt, kein Kommando im ersten Abnahmeumfang (`SPEC-235`, Entscheidung).
- Das Diagnose-Log und die Zähler (`Telemetry`) — `slice-telemetry-port`; ein Audit-Eintrag ist ein Nachweis, kein Diagnose-Log.

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

- [ ] Konfig und Fake: `session.Konfig`, `maintenance.Konfig`, `workspace.Konfig` und `baselineimport.Konfig` führen additiv das Feld `Audit`, `New` weist nil mit `model.ErrUngueltig` ab und ruft `Append` nie; `FakeAudit` hält jeden Eintrag mit allen sieben Feldern unverändert fest, zählt jeden Aufruf und ist für gleichzeitigen Aufruf sicher (`SPEC-234` Nr. 8, 9; Zeilen E01 bis E04, E17).
- [ ] Vier Aufrufstellen: `SessionUseCase.Delete`, die Löschung der Aufbewahrung beim Start und `WorkspaceUseCase.Delete` schreiben je gelöschtem Gegenstand genau einen Eintrag nach der Löschung und vor der Bereinigung; ein abgewiesenes Löschen (`ODR-E7004`, `ODR-E7009`, `ODR-E7003`, `ODR-E7001`) schreibt keinen Eintrag über eine Löschung, die nicht geschah (`LH-RB-04.a`); ein fehlgeschlagener Import schreibt seinen Eintrag mit dem Code des Fehlers; ein Fehler von `Append` und ein beendeter Kontext ändern Rückgabe, Bericht und Fehler nicht (`SPEC-234` Nr. 4, 5, 10; Zeilen E05 bis E16).
- [ ] Kein Geheimnis: kein Feld eines Eintrags trägt Bytes aus Fehlertext, `UpstreamURL` oder Header (`LH-QA-04`; Zeile E18); die Löschung der Blobs schreibt keinen Eintrag (E05).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: dieser Teil führt weder Konfigurationsschlüssel noch Fehlercode noch Datei ein und verdrahtet nichts; in §7 steht, dass das Binary unverändert blieb.
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
| `src/hexagon/ports/support/supporttest/` | update | `FakeAudit` (`SPEC-234` Nr. 9); Port und Typen kommen aus `slice-audit-log` |
| `src/hexagon/services/session/` | update | Feld `Audit` in der Konfig, Eintrag in `Delete` nach `Sessions.Delete` (`SPEC-124` Nr. 7 d1) |
| `src/hexagon/services/maintenance/` | update | Feld `Audit`, Eintrag je gelöschter Session der Aufbewahrung (`SPEC-123` Nr. 6); der Bericht bleibt unverändert |
| `src/hexagon/services/workspace/` | update | Feld `Audit`, Eintrag beim Löschen des Workspace (`SPEC-129` Nr. 9 b1) |
| `src/hexagon/services/baselineimport/` | update | Feld `Audit`, Eintrag bei fehlgeschlagenem Import (`SPEC-127` Nr. 4 m) |
| `src/bootstrap/`, `src/adapters/driven/audit/` | entfällt | Verdrahtung: `slice-audit-verdrahtung`; Adapter: `slice-audit-log` |
| `tools/schema/schema.yaml`, `src/adapters/driven/sqlite/migrations/` | entfällt | die Ablage ist eine Datei (`SPEC-235`): kein Schema, keine Migration |
| `spec/spezifikation.md` | update (Architect, getan) | `SPEC-234` und der Nachzug in `SPEC-123`, `SPEC-124`, `SPEC-127`, `SPEC-129` |
| Tests | update, neu | Fakes der Ports; ein Fall je Löschvorgang |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen hat der Architect entschieden, nicht der Implementer.

**Aufträge an den Architect** (alle entschieden, Ort je Auftrag): Menge der Vorgänge, Reihenfolge und Fehler, Aufrufstellen als Menge, Herkunft von Zeit und Korrelations-ID (`SPEC-234` Nr. 2, 4, 5, 7, 9); Ablage und Adapter liegen bei `slice-audit-log` (`SPEC-235`).

**Aufträge an den Architect aus Teil A** (Start, Planner, 2026-11-06; Quelle: `done/slice-audit-log.md` §7 *Weiter offen*, Punkt 1; Ausgang jeweils *offen* in §6):

1. **Fenster zwischen dem Commit der Löschung und `Append`** (E06 bis E13, E19). Zu entscheiden vor dem Code: ob die Zeilen E10 und E11 (Reihenfolge `Delete`, `Append`, `CleanupBlobs` bzw. vor dem nächsten `Delete`) das Fenster hinreichend fassen oder ob eine Zeile je Fenster fehlt (zwischen welchen Schritten, auf welcher Seite; `AGENTS.md` §3.10 gilt für Sperren, die Dienste halten keine, `SPEC-234` Nr. 10); und ob der Fänger über das gemeinsame Protokoll der Aufrufe (§4, Fake 2) die Lage *vor* `Append` (Löschung steht, Eintrag fehlt) beobachtbar macht, damit die Folge von E19 („eine Löschung ohne Eintrag“) im Test sichtbar wird, nicht nur im Negativ. Haltepunkt am Argument des Tests prüfen (`AGENTS.md` §3.10: Haltepunkt an einem Argument, das der Test übergibt, vor einem Negativ).
2. **Abbruch nach dem Commit / `Append` kehrt nie zurück** (E13, E19; `SPEC-234` *Ungeprüft*). Zu entscheiden: ob E19 bei einem Fake mit blockierendem `Append` (Haltepunkt, `ctx` des Dienstes beendet) einen Fänger bekommt, der zeigt, dass der Dienst wartet und nichts anderes tut, oder als benanntes Negativ mit Grund und Folge bleibt; dazu `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (2×, offen; dieser Slice wäre das dritte Auftreten, wenn der Zweig ohne Lauf bleibt).
3. **Hinweis aus der Closure von Teil A** (kein Auftrag an den Vertrag): die Tests von `slice-audit-log` legten Dateien in `/dev/shm` an; `Sync` dort sagt nichts über Dauer. Diese Fänger nutzen `FakeAudit`, keine Datei; der Verifier liest das mit.

**Bedingung Gate-Kette** (Übergangsregeln Nr. 7, Start, Planner): Dieser Slice berührt laut §3 keine Datei der Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel) und hat die Linie D; der Start hängt nicht an einem Slice der Linie C in `in-progress/`. Der Architect hält das als Bedingung: ändert sein Vertrag oder der Code eine dieser Dateien (auch `.a-check.yml` und `.golangci.yml`; `supporttest` liegt unter `ports/support` und ist den Diensten erlaubt, die Tests der Dienste importieren es schon), geht das als Rückgabe an den Planner (§6, Ausgang *offen*), nicht in den Zweig. Stand beim Start: kein Slice der Gate-Kette in `in-progress/` (dort liegen `slice-odata-json-schreiben` mit `src/hexagon/ports/outbound` und `src/adapters/driven/odata/jsoncodec` und `slice-query-interpreter` mit `src/hexagon/model`, beide ohne Pfad dieses Slice und ohne Datei der Gate-Kette; beide tragen dieselbe Bedingung).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Spalten nach `SPEC-128` Nr. 3 a; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Zeilen E01 bis E19 sind die des Architect aus dem ungeschnittenen Plan von `slice-audit-log`, unverändert.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| E01 | `FakeAudit` hält jeden Eintrag mit allen sieben Feldern unverändert fest, auch mit Werten, die der Adapter abwiese | Schleife über die sieben Felder, je ein Wert, Soll byte-gleich | Feld `ImportID` nicht festgehalten | andere |
| E02 | `FakeAudit` zählt jeden Aufruf von `Append` und `Eintraege`; `Eintraege` liefert eine Kopie | Zähler nach drei Aufrufen; eine Änderung am gelieferten Slice ändert nichts | Slice des Fakes statt einer Kopie geliefert | andere |
| E03 | `FakeAudit.Fehler` wird von `Append` geliefert, der Eintrag steht trotzdem | Fall mit gesetztem und mit ungesetztem Fehler | Eintrag bei Fehler verworfen | andere |
| E04 | `FakeAudit` ist für gleichzeitigen Aufruf sicher und verliert keinen Eintrag | 64 Goroutinen × 8 `Append`, Zähler 512; `make test-race` je Prozess | Sperre entfernt | nebenläufig |
| E05 | Bei Erfolg schreibt jede der vier Stellen genau einen Eintrag mit Art und Feldern nach `SPEC-234` Nr. 4 und leerer `KorrelationsID`, auch wenn `CleanupBlobs` Blobs entfernt; für die Blobs entsteht kein Eintrag | Schleife über die vier Stellen; `ServiceID` der Session aus `Sessions.Get`, nicht aus dem Argument; `Entfernt` 0 und 5 | Eintrag je entferntem Blob | andere |
| E06 | `SessionUseCase.Delete` schreibt bei jeder Ablehnung keinen Eintrag: `ctx` nil, `ODR-E7003` aus (b), `ODR-E7009`, `ODR-E7004`, `ODR-E7003` aus (d), Fehler der Ports | Schleife über die Ablehnungen, Zähler von `FakeAudit` 0 | Eintrag vor `Sessions.Delete` | andere |
| E07 | Die Aufbewahrung schreibt keinen Eintrag für eine Session mit `ErrReferenziert` oder `ErrNichtGefunden`, bei `MaxAge` 0 und ohne Session; ein anderer Fehler beendet `Startup` ohne Eintrag für diese Session | Schleife über die Fälle; Zahl der Einträge gleich Zahl der gelungenen `Delete` | Eintrag auch bei `ErrReferenziert` | andere |
| E08 | `WorkspaceUseCase.Delete` schreibt keinen Eintrag bei `ctx` nil, `ODR-E7001` und Fehler des Ports | Schleife über die Ablehnungen, Zähler 0 | Eintrag vor `Workspaces.Delete` | andere |
| E09 | `ImportUseCase.Import` schreibt keinen Eintrag bei einem Verstoß in (a) bis (c) (Anfrage, `ODR-E5005`, `ODR-E6102`, Fehler von `Imports.Get`), bei der Wiederholung mit gleichem Hash (auch scheiternd), bei einem Kontextfehler und bei Erfolg | Schleife über jeden Verstoß von `SPEC-127` Nr. 3 (a bis c) und die übrigen Fälle; Zähler 0 | Eintrag auch bei `ODR-E5005` | andere |
| E10 | Reihenfolge: `Sessions.Delete`, dann `Audit.Append`, dann `CleanupBlobs` (Session und Workspace) | gemeinsames Protokoll der Aufrufe über die Fakes, Soll-Folge | Eintrag nach `CleanupBlobs` | andere |
| E11 | Aufbewahrung: je Session `Delete`, dann `Append`, vor dem nächsten `Delete`; ein Fehler beendet `Startup`, die Einträge der vorigen Sessions bleiben | Soll-Folge bei drei Sessions, Fehler an der zweiten | alle Einträge am Ende der Stufe geschrieben | andere |
| E12 | Ein Fehler von `Append` ändert nichts: Rückgabe, Bericht und Fehler gleich, `CleanupBlobs` gerufen, `Startup` läuft weiter und schreibt den Eintrag der nächsten Session | Schleife über die vier Stellen mit gesetztem `FakeAudit.Fehler`, Vergleich mit dem Lauf ohne Fehler | Fehler von `Append` als Fehler des Dienstes geliefert | andere |
| E13 | Ein beendeter Kontext nach dem Commit ändert den Eintrag nicht: er steht auch dann | Schleife über die vier Stellen: der Fake des Ports beendet `ctx` und liefert `nil`, bei `Import` einen Fehler mit Code; Zähler 1 | Eintrag nur bei `ctx.Err() == nil` | andere |
| E14 | `ErrorCode` des Eintrags ist der Code des Fehlers jedes Schritts (d) bis (l); ein Fehler ohne Code ergibt `ODR-E9001` | Schleife über die Fehlerstellen von `SPEC-127` Nr. 17 × erreichbare Codes, Orakel aus der Tabelle, Zählprüfung | Code fest `ODR-E9001` | andere |
| E15 | Wiederholung: ein zweites `Delete` (Session, Workspace) und ein zweites `Startup` schreiben nichts, ein zweiter gescheiterter Import schreibt einen zweiten Eintrag, ein gelungener Import danach keinen | Schleife über die Stellen | Eintrag beim zweiten `Delete` (`ODR-E7003`) | andere |
| E16 | Zwei gleichzeitige `Delete` derselben Session und desselben Workspace schreiben genau einen Eintrag | 32 Paare gegen einen Fake des Ports, der atomar löscht; Zähler je Paar 1 | Eintrag vor dem Ergebnis des Ports | nebenläufig |
| E17 | `New` jedes der vier Dienste weist ein nil-`Audit` mit `model.ErrUngueltig` ab und ruft `Append` nie | Schleife über die vier Dienste, Zähler 0 | Prüfung von `Audit` fehlt | andere |
| E18 | Kein Feld eines Eintrags trägt Bytes eines Geheimnisses: Fehlertext der Ports, Userinfo der `UpstreamURL`, Wert eines Headers | Canary im Fehlertext jedes Ports und in der `UpstreamURL`; jedes der sieben Felder jedes Eintrags durchsucht | Fehlertext als Feld | andere |
| E19 | Akzeptierte Negative: der Prozess endet zwischen dem Commit der Löschung und `Append`; `Append` kehrt nie zurück | kein Fänger. Grund: kein Test im Prozess löst beides aus. Folge: eine Löschung ohne Eintrag, bzw. der Dienst wartet (`SPEC-234` *Ungeprüft*) | keine | andere |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede der vier Stellen, jeder Fall einer abgewiesenen Löschung, jeder Fehlerschritt des Imports, jedes der sieben Felder), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-audit-log` liegt in `done/` (Port, `Eintrag`, `AuditArt`, `Arten()`; dieser Slice importiert sie). `slice-workspace-use-case` liegt in `done/` (`WorkspaceUseCase` mit `Delete`). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-audit-log` (`welle-betrieb`): ohne den Port `AuditLog` und die Typen hat das Feld `Audit` keinen Typ (*Code-Abhängigkeit*). Der Adapter ist nicht Bedingung; dieser Slice prüft gegen `FakeAudit`.
- Code, der schon in `done/` liegt: `SessionUseCase.Delete` (`slice-session-use-case`), `MaintenanceUseCase.Startup` mit der Aufbewahrung (`slice-session-wartung`), `WorkspaceUseCase` (`slice-workspace-use-case`), `ImportUseCase` (`slice-baseline-import`).
- Folge-Slice: `slice-audit-verdrahtung` braucht diesen Slice (die vier Felder `Audit`).

**Unabhängig von `welle-mock` / M3 und von M2 (`slice-bootstrap-m2`), weil** alle Aufrufstellen Use Cases sind, die nicht über HTTP erreichbar sein müssen: die Tests rufen sie wie ihre Vorgänger über den Dienst. Dieser Slice ändert `src/bootstrap/` nicht. Berühren `slice-telemetry-einbau-gateway` oder ein anderer laufender Slice dieselben Dateien unter `src/hexagon/services/`, laufen beide nacheinander (der Planner nennt die Reihenfolge beim Start des zweiten).

**Start-Ausnahme je Slice** (Linie D, `welle-betrieb` §2): wie `slice-audit-log`; `welle-mock` und M2 sind nicht Bedingung. Kennungsbereich der Linie D (Übergangsregeln Nr. 6): `SPEC-230` bis `SPEC-259`, ADR `0040` bis `0049`, `CO-060` bis `CO-079`.

**Größenmarke** (`SPEC-128`; Faktoren allein dort): Der Block mit `basis`, `schaetzung` und `marke` entsteht beim Start, weil `basis` der Commit ist, der den Stand nach `slice-audit-log` abschließt; bis dahin steht hier die Rechnung als Prosa. Pfade: `src/hexagon/ports/support`, `src/hexagon/services/session`, `src/hexagon/services/maintenance`, `src/hexagon/services/workspace`, `src/hexagon/services/baselineimport`. Erwarteter Produktcode (Architect): 260 (vier Konfig-Felder mit Eintrag und `FakeAudit`). Fakes: 2 (`FakeAudit` und das gemeinsame Protokoll der Aufrufe über die Fakes der Dienste). Zeilen: 19, davon nebenläufig 2 (E04, E16), andere 17. Nach `SPEC-128` Nr. 1: *S* 2190, *M* 3000, die Schnittfrage stellt sich nicht (*S* unter 2223, knapp: Abstand 32 zu 2222). Der Slice liefert keinen Leser fremder Bytes. `basis` muss hinter dem letzten Commit von `slice-audit-log` liegen, damit dessen Zeilen unter `src/hexagon/ports/support` nicht in die Messung dieses Slice fallen (`SPEC-128` Nr. 10 g). Wächst der Slice über 3000, steht der Eintrag `ueberschreitung` mit Grund im Block, keine Rückführung (`SPEC-128` Nr. 5 d).
**Stand beim Start** (Planner, 2026-11-06): Rechnung nachgerechnet, *S* = 85·2 + 80·17 + 200·2 + 260 = 2190, *M* = 3000, keine Schnittfrage (2190 < 2223, Abstand 33); die Schätzung und die Marke bleiben, der Block steht auf `main` nicht (ein Block ohne Architect-Werte wäre ein Platzhalter). `basis` setzt der Architect im Zweig auf den Start-Commit des Zweigs (er liegt hinter `slice-audit-log`, `d0748f1`), `schaetzung` 2190, `marke` 3000, `produktcode` 260, `fakes` 2. Der Messwert von `slice-audit-log` (Ende ÷ Schätzung 0,87, Auslastung 0,63) liegt dem Architect für `SPEC-128` Nr. 9 vor; er ändert die Formel dieses Slice nicht.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke ohne gedeckten Eintrag; die Schnittfrage ist vor dem ersten Code-Commit entschieden (Naht: Dienste gegen Verdrahtung steht schon in `slice-audit-verdrahtung`).
- `in-progress` → `open` (blockiert): `slice-audit-log` liefert Port und Typen nicht in der Form von `SPEC-234`; oder eine Aufrufstelle verlangt eine Änderung der Gate-Kette (nicht zu erwarten: kein Gate ändert sich).

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

- Löschen und Eintrag sind zwei Schritte; scheitert einer, fehlt entweder die Löschung oder ihr Nachweis (Fenster zwischen Commit der Löschung und Schreiben des Eintrags) — **Ausgang:** offen (bei der Closure); entschieden in `SPEC-234` Nr. 5 (Eintrag nach der Tatsache, Fehler ohne Wirkung, Kontext ohne Wirkung); Fänger E06 bis E13, benanntes Negativ E19 (Prozessende im Fenster, Folge: eine Löschung ohne Eintrag).
- Der Bericht der Wartung nennt keine Kennungen der gelöschten Sessions; die Löschung der Aufbewahrung geht über `SessionRepository`, nicht über `SessionUseCase` (`SPEC-123`) — **Ausgang:** entfallen (Architect, 2026-11-06): der Bericht bleibt unverändert, die Wartung ruft den Port je Session (`SPEC-123` Nr. 6, `SPEC-234` Nr. 4 b); Fänger E07, E11.
- Doppelung der Einträge zwischen Use Case und Admin-Adapter — **Ausgang:** weiter offen bis `welle-admin-api`; Adresse: `slice-admin-sessions`, `slice-admin-workspaces`, `slice-admin-imports` (Geber: dieser Slice und `slice-audit-log`). Entschieden in `SPEC-234` Nr. 7 und `SPEC-005`: die vier Arten schreibt allein der Use Case, den dieser Slice liefert; die Nehmer tragen den Verweis seit der Anlage dieses Plans in §1 oder DoD.
- Diese Aufrufstellen sind die einzige Quelle der vier Arten: eine neue Löschstelle braucht eine Zeile in `SPEC-234` Nr. 4 und im Fänger (`SPEC-234` *Ungeprüft*) — **Ausgang:** weiter offen (bei der Closure); der Reviewer liest die Aufrufe von `Sessions.Delete`, `Workspaces.Delete` und `Imports` im Paket.
- Konflikt mit anderen Slices an `src/hexagon/services/` (Telemetrie-Einbau in `record` und `session`, Linie D) — **Ausgang:** weiter offen; Adresse: Planner beim Start des zweiten Slice (nacheinander, §4). Stand beim Start: `slice-telemetry-einbau-gateway` liegt in `open/` und startet nach diesem Slice; in `in-progress/` berührt kein anderer Slice `src/hexagon/services/session`, `maintenance`, `workspace`, `baselineimport` oder `ports/support`.
- Fenster zwischen Löschung und Eintrag, Abbruch nach dem Commit (Rückgabe aus Teil A, Aufträge 1 und 2 in §3) — **Ausgang:** offen; Adresse: Architect (vor dem ersten Code-Commit; bis dahin gilt E19 als benanntes Negativ mit Folge aus `SPEC-234` *Ungeprüft*).
- Gate-Kette (Bedingung in §3): ein Bedarf an einer Änderung der Gate-Kette — **Ausgang:** weiter offen (nicht zu erwarten); Adresse: Planner, falls der Architect oder der Implementer ihn meldet.
- Größenmarke knapp unter der Schwelle der Schnittfrage (*S* 2190 gegen 2223) und ein Vergleichswert über der Marke (`slice-telemetry-port`, Ende ÷ *S* 1,44) — **Ausgang:** weiter offen (bei der Closure); Folge ist ein Eintrag `ueberschreitung` mit Grund, keine Rückführung.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster unten steht vor dem ersten Code-Commit, jede Zelle mit Ort oder „gilt nicht“ und Grund; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen (bis zum Review).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Das Argument der Operationen ist die Kennung des Gegenstands (`id`) bzw. der Eintrag; „fehlende Kennung“ gibt es bei einem `string` nicht (er ist die leere), die Zelle zeigt auf die leere.

| Operation | fehlende Kennung | leere Kennung | unbekannte oder fremde Kennung (Löschung scheitert) | Abbruch (`ctx`) oder Schreibfehler der Ablage | Wiederholung (zweites Löschen) | Geheimnis im Gegenstand |
|---|---|---|---|---|---|---|
| Eintrag zu `SessionUseCase.Delete` | gilt nicht: ein `string` hat keinen fehlenden Wert, die leere Kennung steht in der nächsten Spalte | `SPEC-124` Nr. 7 b (`Sessions.Get` liefert `ErrNichtGefunden`, `ODR-E7003`), kein Eintrag: `SPEC-234` Nr. 4 a | `SPEC-124` Nr. 7 b, c, d (`ODR-E7003`, `ODR-E7009`, `ODR-E7004`), kein Eintrag: `SPEC-234` Nr. 4 a und 5 a | `SPEC-234` Nr. 5 b, c (Fehler und Abbruch ohne Wirkung), Schreibfehler: `SPEC-235` Nr. 4 und 5 | `SPEC-234` Nr. 10 (`ODR-E7003`, kein zweiter Eintrag) | `SPEC-234` Nr. 3 (Wertform, kein Freitext) und `SPEC-235` Nr. 6 |
| Eintrag zur Löschung der Aufbewahrung (Start) | gilt nicht: die Kennungen kommen aus `Sessions.List` der Ablage, die Operation hat kein Argument | gilt nicht: eine leere Kennung kann `Sessions.List` nicht liefern (`SPEC-124` Nr. 4: jede Kennung hat 1 bis 64 Bytes) | `SPEC-123` Nr. 6 (`ErrReferenziert`, `ErrNichtGefunden`), kein Eintrag: `SPEC-234` Nr. 4 b | `SPEC-234` Nr. 5 b, c (`Startup` bricht nicht ab, der Kontextfehler erscheint am nächsten Port), Schreibfehler: `SPEC-235` Nr. 4 und 5 | `SPEC-234` Nr. 10 (zweites `Startup` ohne neue Daten, `SPEC-123` Nr. 12) | `SPEC-234` Nr. 3 und `SPEC-235` Nr. 6 |
| Eintrag zum Löschen eines Workspace | gilt nicht: wie oben | `SPEC-129` Nr. 9 b (die `id` geht unverändert an `Workspaces.Delete`, `ODR-E7001`), kein Eintrag: `SPEC-234` Nr. 4 c | `SPEC-129` Nr. 9 b (`ODR-E7001`), kein Eintrag: `SPEC-234` Nr. 4 c | `SPEC-234` Nr. 5 b, c, Schreibfehler: `SPEC-235` Nr. 4 und 5 | `SPEC-129` Nr. 14 g und `SPEC-234` Nr. 10 (`ODR-E7001`, kein zweiter Eintrag) | `SPEC-234` Nr. 2 (keine `service_id`, kein Zähler) und Nr. 3 |
| Eintrag zu einem fehlgeschlagenen Import | `SPEC-127` Nr. 3 (Anfrage ungültig), kein Eintrag: `SPEC-234` Nr. 4 d | `SPEC-127` Nr. 3 a (leere `ImportID`) und b (leere `ServiceID`), kein Eintrag: `SPEC-234` Nr. 4 d | unbekannter Service `ODR-E5005`: kein Eintrag (`SPEC-234` Nr. 4 d); unbekannte Session `ODR-E7003`, fremder Service `ODR-E5004`: Eintrag mit diesem Code (`SPEC-234` Nr. 4 d) | Kontextfehler: kein Eintrag (`SPEC-234` Nr. 4 d); Fehler des Eintrags: `SPEC-234` Nr. 5 b | `SPEC-234` Nr. 10 (jeder Fehlschlag ein Eintrag; der Fall „bekannt“ keiner) | `SPEC-234` Nr. 2 und 4 d (die Quell-Session des Clients steht nicht im Eintrag), Nr. 3 |

**Zusatztabelle** (`AGENTS.md` §3.16: mehrere Verstöße und Abhängigkeit vom Modus), mit der Spalte *Fänger*:

| Fall | Ort | Fänger |
|---|---|---|
| Fenster zwischen dem Commit der Löschung und `Append` | `SPEC-234` Nr. 5 und *Ungeprüft* | E19 (benanntes Negativ), Reihenfolge E10, E11 |

Der Modus (`record` und `replay`, `SPEC-139` Nr. 4) und die Dateiebene gehören zu `slice-audit-verdrahtung`; die Zeilen der Sperre des Adapters zu `slice-audit-log`.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-06, aus dem Plan von `slice-audit-log` übernommen) und beim Start nachgezogen (Planner, 2026-11-06; Zähler = Zahl der Dateien in `evidence/`). Der Slice hat viele Fakes (vier Dienste, ein Port): `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, verkörpert in `implement-slice` Schritt 14 und im Reviewer-Skill) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×, verkörpert im Reviewer-Skill) gelten für die Fakes; offen ist `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (2×) für den Abbruch nach dem Commit: ein drittes Auftreten in diesem Slice löst die Verkörperung aus. `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (7×) ist geplant in `slice-guard-schreibwege` (Betriebsregel im Plan §3 gilt bis dahin). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (41×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (13×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (32×, `AGENTS.md` §3.16) und `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (24×, `AGENTS.md` §3.17; in `slice-audit-log` dreimal in Folge, die Achsen aus dem Wortlaut der Zusage ableiten, nicht aus dem Befund). Zu Schätzung und Marke: `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (10×).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
