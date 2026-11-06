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

**Ziel (Teil E des Schnitts von `slice-audit-log` in drei Teile):** Die vier Dienste halten fest, was sie gelöscht haben: `SessionUseCase.Delete`, die Löschung der Aufbewahrung in `MaintenanceUseCase.Startup` (Stufe Aufbewahrung, `SPEC-123` Nr. 6) und `WorkspaceUseCase.Delete` schreiben je gelöschtem Gegenstand genau einen Eintrag über den Port `AuditLog` (Art, Gegenstand, ohne Geheimnisse, `LH-RB-04.a`), und `ImportUseCase.Import` schreibt bei einem fehlgeschlagenen Import den Eintrag `import_fehlgeschlagen`, damit seine `import_id` wiederverwendbar bleibt (`SPEC-004` Nr. 2). Jede der vier Konfigs führt dafür das Feld `Audit support.AuditLog` (`SPEC-234` Nr. 8); der Eintrag folgt der Löschung, sein Fehler und ein beendeter Kontext ändern das Ergebnis des Dienstes nicht, und der Dienst wartet ohne Frist auf `Append` (`SPEC-234` Nr. 5). Ein Fake `FakeAudit` im Paket `supporttest` hält jedes Argument fest, zählt jede Operation und bietet den Haltepunkt `Vor`, in dem ein Test die Lage vor dem Eintrag sieht (`SPEC-234` Nr. 9). Port, `Eintrag`, `Art` und `Arten()` liefert `slice-audit-log`; die Dienste rufen den Adapter nie und ihre Tests prüfen gegen `FakeAudit`; die Verdrahtung im Composition Root trägt der Zweig als Brücke (zwei Commits »Brücke:«), die die eine Instanz `audit.Datei` an die vier Konfigs gibt, damit `Build` und `make gates` laufen (§6, Ausgang: die Brücke bleibt; `slice-audit-verdrahtung` ist auf die Fänger der Datei im Betrieb verkleinert); die Tests der Brücke sind `TestAuditEineInstanz` und `TestAuditPfadUndZeile` in `src/bootstrap/`.

**Übernimmt (Sendung von `slice-session-use-case`, §1 *Ausdrücklich NICHT*, Stand `done/`; Geber der Weitergabe: `slice-audit-log`):** der Audit-Eintrag der Löschung einer Session in `SessionUseCase.Delete`; die Löschung selbst liegt dort.

**Übernimmt (Sendung von `slice-session-wartung`, §1 *Ausdrücklich NICHT*, Stand `done/`; Geber der Weitergabe: `slice-audit-log`):** der Audit-Eintrag der Löschung durch die Aufbewahrung in `MaintenanceUseCase.Startup`; die Löschung selbst liegt dort. Beide Pläne liegen in `done/` und sind nicht änderbar; die Übernahme steht hier.

**Übernimmt (Zuweisung von `slice-audit-log`, ursprünglicher Umfang):** der Eintrag der Löschung eines Workspace in `WorkspaceUseCase.Delete` (`slice-workspace-use-case` liegt in `done/`; sein §1 schließt den Eintrag nicht aus) und der Eintrag eines fehlgeschlagenen Imports in `ImportUseCase` (`slice-baseline-import` in `done/`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Port `AuditLog`, `Eintrag`, `AuditArt`, `Arten()`, der Adapter `audit.Datei` und die Datei `audit.jsonl` — `slice-audit-log`. Grund: Naht des Schnitts; dieser Slice braucht den Port und die Typen, nie den Adapter, und prüft gegen `FakeAudit`.
- Die Fänger der Datei im Betrieb über das `Handle` (Zeilen, Neustart, Canary, keine Datei ohne Löschung) — `slice-audit-verdrahtung`, verkleinert auf V01, V02, V03 und den Rest von V05. Grund: sie laufen gegen das gebaute `Handle` und brauchen die Dienste fertig. Die Verdrahtung selbst (Instanz, Pfad, Übergabe an die vier Konfigs) liegt in diesem Slice als Brücke (zwei Commits »Brücke:«), weil `New` ein nil-`Audit` abweist und `Build` sonst scheitert (§6, Ausgang: die Brücke bleibt).
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

- [x] Konfig und Fake: `session.Konfig`, `maintenance.Konfig`, `workspace.Konfig` und `baselineimport.Konfig` führen additiv das Feld `Audit`, `New` weist nil mit `model.ErrUngueltig` ab und ruft `Append` nie; `FakeAudit` hält jeden Eintrag mit allen sieben Feldern unverändert fest, zählt jeden Aufruf, bietet den Haltepunkt `Vor` und ist für gleichzeitigen Aufruf sicher (`SPEC-234` Nr. 8, 9; Zeilen E01 bis E04, E17).
- [x] Vier Aufrufstellen: `SessionUseCase.Delete`, die Löschung der Aufbewahrung beim Start und `WorkspaceUseCase.Delete` schreiben je gelöschtem Gegenstand genau einen Eintrag nach der Löschung und vor der Bereinigung; ein abgewiesenes Löschen (`ODR-E7004`, `ODR-E7009`, `ODR-E7003`, `ODR-E7001`) schreibt keinen Eintrag über eine Löschung, die nicht geschah (`LH-RB-04.a`); ein fehlgeschlagener Import schreibt seinen Eintrag mit dem Code des Fehlers; ein Fehler von `Append` und ein beendeter Kontext ändern Rückgabe, Bericht und Fehler nicht, und der Dienst wartet auf `Append`, ohne Frist und ohne Goroutine (`SPEC-234` Nr. 4, 5, 10; Zeilen E05 bis E16 und E19).
- [x] Kein Geheimnis: kein Feld eines Eintrags trägt Bytes aus Fehlertext, `UpstreamURL` oder Header (`LH-QA-04`; Zeile E18); die Löschung der Blobs schreibt keinen Eintrag (E05).
- [x] `make gates` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [x] Handbuch: dieser Teil führt weder Konfigurationsschlüssel noch Fehlercode noch Datei ein; die Brücke im Composition Root bleibt (§6), das Binary schreibt bei einer Löschung durch die Aufbewahrung die Datei `audit.jsonl` im Datenverzeichnis (`SPEC-235`), und §7 sagt das; das Kapitel dazu trägt `slice-admin-sessions`.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/support/supporttest/` | update | `FakeAudit` mit dem Haltepunkt `Vor` (`SPEC-234` Nr. 9); Port und Typen kommen aus `slice-audit-log` |
| `src/hexagon/services/session/` | update | Feld `Audit` in der Konfig, Eintrag in `Delete` nach `Sessions.Delete` (`SPEC-124` Nr. 7 d1) |
| `src/hexagon/services/maintenance/` | update | Feld `Audit`, Eintrag je gelöschter Session der Aufbewahrung (`SPEC-123` Nr. 6); der Bericht bleibt unverändert |
| `src/hexagon/services/workspace/` | update | Feld `Audit`, Eintrag beim Löschen des Workspace (`SPEC-129` Nr. 9 b1) |
| `src/hexagon/services/baselineimport/` | update | Feld `Audit`, Eintrag bei fehlgeschlagenem Import (`SPEC-127` Nr. 4 m) |
| `src/bootstrap/` | update (Brücke) | `wurzel.go` übergibt `Audit` an die vier Konfigs, sonst scheitert `Build` an dem nil-`Audit`; die Brücke bleibt, `slice-audit-verdrahtung` fängt die Datei im Betrieb (§6) |
| `src/adapters/driven/audit/` | entfällt | Adapter: `slice-audit-log`; dieser Slice ruft ihn nur über die Brücke |
| `tools/schema/schema.yaml`, `src/adapters/driven/sqlite/migrations/` | entfällt | die Ablage ist eine Datei (`SPEC-235`): kein Schema, keine Migration |
| `spec/spezifikation.md` | update (Architect, getan) | `SPEC-234` und der Nachzug in `SPEC-123`, `SPEC-124`, `SPEC-127`, `SPEC-129` |
| Tests | update, neu | Fakes der Ports; ein Fall je Löschvorgang; je Paket `audit_test.go` (Zeilen E05 bis E19 der Stelle) und `supporttest/audit_test.go` (E01 bis E04); Haken in den vorhandenen Fakes: `nachGet` und `nachDelete` (Session), `nachDelete` (Aufbewahrung), `beiCleanup` (Workspace), `nach` für `Write` und `Commit` und `CodeVon` in `export_test.go` (Import) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen hat der Architect entschieden, nicht der Implementer.

**Aufträge an den Architect** (alle entschieden, Ort je Auftrag): Menge der Vorgänge, Reihenfolge und Fehler, Aufrufstellen als Menge, Herkunft von Zeit und Korrelations-ID (`SPEC-234` Nr. 2, 4, 5, 7, 9); Ablage und Adapter liegen bei `slice-audit-log` (`SPEC-235`).

**Aufträge an den Architect aus Teil A** (Start, Planner, 2026-11-06; Quelle: `done/slice-audit-log.md` §7 *Weiter offen*, Punkt 1) — **entschieden** (Architect, 2026-11-06; Ausgang in §6):

1. **Fenster zwischen dem Commit der Löschung und `Append`.** Es sind drei Fenster, nicht eines (Zusatztabelle in §6): F1 vor `Append` (die Löschung steht, der Eintrag fehlt; Session, Aufbewahrung, Workspace; bei `Import` ist der Fehler entstanden, der Eintrag fehlt), F2 nach `Append` vor der Bereinigung (Session, Workspace), F3 nach `Append` vor dem nächsten Aufruf eines Ports (Aufbewahrung). E10 und E11 fassen F2 und F3, wenn ihr Fänger den Stand im folgenden Aufruf prüft (der Eintrag steht schon), nicht nur die Folge; F1 hatte keine Zeile. Die Lage von F1 wird im Test sichtbar: `FakeAudit.Vor` hält den Aufruf von `Append`, und das gemeinsame Protokoll zeigt die Löschung bei fehlendem Eintrag (E19; `SPEC-234` Nr. 5 d, Nr. 9). Haltepunkt-Check (`AGENTS.md` §3.10): das Argument, das der Test übergibt, ist `Konfig.Audit`, der Fake selbst; E19 braucht deshalb kein Negativ für die Lage, nur für das Ende des Prozesses. Eine Sperre gibt es nicht (`SPEC-234` Nr. 10): die Fenster liegen zwischen zwei Ports.
2. **Abbruch nach dem Commit, `Append` kehrt nie zurück.** E19 bekommt einen Fänger (Haltepunkt `Vor`, `ctx` im Halt beendet, Frist 100 ms wie in `bereinigung_test.go`) und die Zusage `SPEC-234` Nr. 5 d (synchron, ohne Frist, ohne Abbruch durch den Kontext). Negativ bleiben das Ende des Prozesses und das nie endende `Append`, je mit Grund und Folge in E19 und in `SPEC-234` *Ungeprüft*. `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`: E16 und E19 laufen unter `make test-race`; kein Kopf eines Tests sagt zu, der Detektor melde jedes Paar (`SPEC-234` *Ungeprüft*, Datenrennen).
3. **Hinweis aus der Closure von Teil A:** diese Fänger nutzen `FakeAudit` ohne Datei und nie `/dev/shm`; ein Test, der eine Datei bräuchte, ist eine Abweichung vom Plan (§6, Ausgang *offen*, Adresse Architect). Der Verifier liest das mit.

**Bedingung Gate-Kette** (Übergangsregeln Nr. 7, Start, Planner): Dieser Slice berührt laut §3 keine Datei der Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel) und hat die Linie D; der Start hängt nicht an einem Slice der Linie C in `in-progress/`. Der Architect hält das als Bedingung: ändert sein Vertrag oder der Code eine dieser Dateien (auch `.a-check.yml` und `.golangci.yml`; `supporttest` liegt unter `ports/support` und ist den Diensten erlaubt, die Tests der Dienste importieren es schon), geht das als Rückgabe an den Planner (§6, Ausgang *offen*), nicht in den Zweig. Stand beim Start: kein Slice der Gate-Kette in `in-progress/` (dort liegen `slice-odata-json-schreiben` mit `src/hexagon/ports/outbound` und `src/adapters/driven/odata/jsoncodec` und `slice-query-interpreter` mit `src/hexagon/model`, beide ohne Pfad dieses Slice und ohne Datei der Gate-Kette; beide tragen dieselbe Bedingung).

**Prüfung der Bedingung durch den Architect** (2026-11-06): `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel, `.a-check.yml` und `.golangci.yml` bleiben unberührt. `.a-check.yml` erlaubt `services` nach `ports-support` (`from: services, to: ports-support`) und trägt `supporttest` mit dem Glob `src/hexagon/ports/support/**`; die Dienste importieren `ports/support` schon. `.golangci.yml` nennt in `forbidigo` nur Ausgabe, nicht `time.After` oder Goroutinen im Test. Es entsteht keine Zeile für den Planner.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Spalten nach `SPEC-128` Nr. 3 a; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Zeilen E01 bis E19 stammen vom Architect aus dem ungeschnittenen Plan von `slice-audit-log`; E02, E03, E05, E09 bis E11, E13, E16 und E19 tragen die Schärfung vor dem Code (Haltepunkt `Vor`, Mengen, Barriere), die Zahl der Zeilen ist gleich.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| E01 | `FakeAudit` hält jeden Eintrag mit allen sieben Feldern unverändert fest, auch mit Werten, die der Adapter abwiese | Schleife über die sieben Felder, je ein Wert, Soll byte-gleich | Feld `ImportID` nicht festgehalten | andere |
| E02 | `FakeAudit` zählt jeden Aufruf von `Append` und `Eintraege`; `Aufrufe()` ist die Summe, die Länge von `Eintraege()` die Zahl der Aufrufe von `Append`; `Eintraege` liefert eine Kopie | Zähler nach drei Aufrufen von `Append` und zwei von `Eintraege` (Summe 5, Länge 3); eine Änderung am gelieferten Slice ändert nichts | Slice des Fakes statt einer Kopie geliefert; `Eintraege` nicht gezählt | andere |
| E03 | `FakeAudit.Fehler` wird von `Append` geliefert, der Eintrag steht trotzdem; `Vor` läuft nach dem Zählen und vor dem Festhalten, mit dem Eintrag, ohne Sperre | Fall mit gesetztem und mit ungesetztem Fehler; in `Vor`: `Aufrufe()` um 1 gewachsen, `Eintraege()` ohne den Eintrag, Argument gleich dem Eintrag, `Eintraege()` und `Aufrufe()` blockieren nicht | Eintrag bei Fehler verworfen; `Vor` nach dem Festhalten gerufen | andere |
| E04 | `FakeAudit` ist für gleichzeitigen Aufruf sicher und verliert keinen Eintrag | 64 Goroutinen × 8 `Append`, Zähler 512; `make test-race` je Prozess | Sperre entfernt | nebenläufig |
| E05 | Bei Erfolg schreibt jede der vier Stellen genau einen Eintrag mit Art und Feldern nach `SPEC-234` Nr. 4 und leerer `KorrelationsID`, auch wenn `CleanupBlobs` Blobs entfernt; für die Blobs entsteht kein Eintrag; die Kennungen stehen unverändert im Eintrag | Schleife über die vier Stellen; `ServiceID` der Session aus `Sessions.Get`, nicht aus dem Argument; `Entfernt` 0 und 5; Kennungen nach M5 (Länge 1 und 64, jedes Zeichen der Form) | Eintrag je entferntem Blob; `ServiceID` klein geschrieben | andere |
| E06 | `SessionUseCase.Delete` schreibt bei jeder Ablehnung keinen Eintrag: `ctx` nil, `ODR-E7003` aus (b), `ODR-E7009`, `ODR-E7004`, `ODR-E7003` aus (d), Fehler der Ports | Schleife über die Ablehnungen, Zähler von `FakeAudit` 0 | Eintrag vor `Sessions.Delete` | andere |
| E07 | Die Aufbewahrung schreibt keinen Eintrag für eine Session mit `ErrReferenziert` oder `ErrNichtGefunden`, bei `MaxAge` 0 und ohne Session; ein anderer Fehler beendet `Startup` ohne Eintrag für diese Session | Schleife über die Fälle; Zahl der Einträge gleich Zahl der gelungenen `Delete` | Eintrag auch bei `ErrReferenziert` | andere |
| E08 | `WorkspaceUseCase.Delete` schreibt keinen Eintrag bei `ctx` nil, `ODR-E7001` und Fehler des Ports | Schleife über die Ablehnungen, Zähler 0 | Eintrag vor `Workspaces.Delete` | andere |
| E09 | `ImportUseCase.Import` schreibt keinen Eintrag bei einem Verstoß in (a) bis (c) (Anfrage, `ODR-E5005`, `ODR-E6102`, Fehler von `Imports.Get`), bei der Wiederholung mit gleichem Hash (auch scheiternd), bei einem Kontextfehler jeder Form und bei Erfolg | Schleife über jeden Verstoß von `SPEC-127` Nr. 3 (a bis c), ihre Paare und alle zugleich, und die übrigen Fälle; Kontextfehler nach M4 (vier Formen an jeder Stelle); Zähler 0 | Eintrag auch bei `ODR-E5005`; Kontextfehler nur mit `==` erkannt | andere |
| E10 | Reihenfolge: `Sessions.Delete`, dann `Audit.Append`, dann `CleanupBlobs` (Session und Workspace); im Aufruf von `CleanupBlobs` steht der Eintrag schon (Fenster F2) | gemeinsames Protokoll der Aufrufe über die Fakes, Soll-Folge; Stand von `FakeAudit.Eintraege()` im Aufruf des Fakes der Wartung | Eintrag nach `CleanupBlobs` | andere |
| E11 | Aufbewahrung: je Session `Delete`, dann `Append`, vor dem nächsten `Delete` (Fenster F3); ein Fehler beendet `Startup`, die Einträge der vorigen Sessions bleiben | Soll-Folge bei drei Sessions, Fehler an der zweiten; im `Delete` der zweiten Session steht der Eintrag der ersten schon; Position und Ergebnis nach M7 | alle Einträge am Ende der Stufe geschrieben | andere |
| E12 | Ein Fehler von `Append` ändert nichts: Rückgabe, Bericht und Fehler gleich, `CleanupBlobs` gerufen, `Startup` läuft weiter und schreibt den Eintrag der nächsten Session | Schleife über die vier Stellen mit gesetztem `FakeAudit.Fehler`, Vergleich mit dem Lauf ohne Fehler | Fehler von `Append` als Fehler des Dienstes geliefert | andere |
| E13 | Ein beendeter Kontext nach dem Commit ändert den Eintrag nicht: er steht auch dann, mit und ohne Fehler von `Append` | Schleife über die vier Stellen × `FakeAudit.Fehler` gesetzt und ungesetzt: der Fake des Ports beendet `ctx` und liefert `nil`, bei `Import` einen Fehler mit Code; Zähler 1; mit gesetztem Fehler dieselbe Rückgabe wie ohne | Eintrag nur bei `ctx.Err() == nil` | andere |
| E14 | `ErrorCode` des Eintrags ist der Code des Fehlers jedes Schritts (d) bis (l); ein Fehler ohne Code ergibt `ODR-E9001` | Schleife über die Fehlerstellen von `SPEC-127` Nr. 17 × erreichbare Codes, Orakel aus der Tabelle, Zählprüfung | Code fest `ODR-E9001` | andere |
| E15 | Wiederholung: ein zweites `Delete` (Session, Workspace) und ein zweites `Startup` schreiben nichts, ein zweiter gescheiterter Import schreibt einen zweiten Eintrag, ein gelungener Import danach keinen | Schleife über die Stellen | Eintrag beim zweiten `Delete` (`ODR-E7003`) | andere |
| E16 | Zwei gleichzeitige `Delete` derselben Session und desselben Workspace schreiben genau einen Eintrag | 32 Paare gegen einen Fake des Ports, der atomar löscht; bei der Session hält ein Fake von `Sessions.Get` beide Aufrufe an einer Barriere (mit `zeitgrenze`), bis beide gelesen haben, und erst danach beginnt ein `Delete`, sonst hinge die Mutation vom Ablauf ab; Zähler je Paar 1 | Eintrag vor dem Ergebnis des Ports (nach `Sessions.Get`, vor `Sessions.Delete`) | nebenläufig |
| E17 | `New` jedes der vier Dienste weist ein nil-`Audit` mit `model.ErrUngueltig` ab und ruft `Append` nie | Schleife über die vier Dienste, Zähler 0 | Prüfung von `Audit` fehlt | andere |
| E18 | Kein Feld eines Eintrags trägt Bytes eines Geheimnisses: Fehlertext der Ports, Userinfo der `UpstreamURL`, Wert eines Headers | Canary im Fehlertext jedes Ports und in der `UpstreamURL`; jedes der sieben Felder jedes Eintrags durchsucht | Fehlertext als Feld | andere |
| E19 | `Append` läuft synchron auf der Goroutine des Dienstes, und der Dienst setzt erst nach seiner Rückkehr fort, auch bei beendetem `ctx` (`SPEC-234` Nr. 5 d); im Halt steht die Löschung, fehlt der Eintrag, und der Dienst tut nichts anderes | Schleife über die vier Stellen (Aufbewahrung: die zweite von drei Sessions) × `ctx` (aktiv, im Halt beendet): `FakeAudit.Vor` meldet das Erreichen und hält; im Halt zeigt das gemeinsame Protokoll die Löschung (bei `Import`: den Fehlerschritt), der Eintrag dieses Aufrufs fehlt in `Eintraege()`, kein `CleanupBlobs`, kein nächster Aufruf eines Ports, und der Aufruf des Dienstes ist nach der Frist von 100 ms nicht zurückgekehrt; nach der Freigabe steht der Eintrag, das Ergebnis ist das des Laufs mit demselben `ctx` ohne Halt. Akzeptierte Negative: das Ende des Prozesses zwischen Commit und `Append` (kein Test beendet den Prozess; Folge: eine Löschung ohne Eintrag, `SPEC-234` *Ungeprüft*) und ein `Append`, das nie zurückkehrt (der Fänger beobachtet das Warten, nicht dessen Dauer; Folge: der Dienst wartet) | `Append` in einer Goroutine gestartet, der Dienst setzt fort; das Warten endet bei `ctx.Done()`; `CleanupBlobs` vor `Append` | nebenläufig |

**Mengen** der Fänger (`implement-slice` Schritt 14 und 19, `AGENTS.md` §3.17): Der Architect nennt vor dem Code Menge, Achsen und Zahl der Elemente je Achse; die Fänger laufen als Schleifen über alle Elemente, mit Orakel und Zählprüfung. Der Bericht des Implementers nennt je Menge die Zahl, die der Test zählt, und je Achse eine Mutation an einem Element, das die Zeile nicht nennt, rot gesehen; die Spalte *Mutation vorab* nennt eine solche.

| Nr. | Menge | Achsen und Elemente | Fänger | Mutation vorab |
|---|---|---|---|---|
| M1 | Aufrufstellen von `Append` | Stelle 4 (`SessionUseCase.Delete`, Aufbewahrung, `WorkspaceUseCase.Delete`, `ImportUseCase.Import`) × Zustand von `ctx` 2 (aktiv, im Halt oder im Port beendet) | E05, E10 bis E13, E15, E17, E19 | Aufbewahrung: `SessionID` der vorigen Session der Schleife im Eintrag |
| M2 | Ablehnungen vor der Tatsache | Stelle 3 × Grund: Session 7 (`ctx` nil · `Get` mit `ErrNichtGefunden` · `Get` mit anderem Fehler · offen `ODR-E7009` · referenziert `ODR-E7004` · `Delete` mit `ErrNichtGefunden` · `Delete` mit anderem Fehler), Aufbewahrung 8 (`ErrReferenziert` · `ErrNichtGefunden` · anderer Fehler beendet `Startup` · `MaxAge` 0 · ohne Session · Session innerhalb der Frist · Fehler von `Dienste.List` · Fehler von `Sessions.List`), Workspace 3 (`ctx` nil · `ODR-E7001` · anderer Fehler) = 18 | E06, E07, E08 | Workspace: Eintrag auch bei einem Fehler von `Workspaces.Delete` ohne Code |
| M3 | Fehlerschritte des Imports | Stelle 9 (`Sessions.Get`, `Exchanges.ListBySession`, `Blobs.Open`, Strom des Codecs, `Blobs.Begin`, `Write`, `Commit`, `Imports.Publish`, `Metadaten.SetMetadata`) × Code 7 mit Eintrag (`ODR-E5004`, `E6004`, `E6101`, `E7003`, `E7009`, `E3001`, `E9001`) und ein Fehler ohne Code × Verpackung 2 (nackt, mit `%w`); nur die Paare, die `SPEC-127` Nr. 17 erreichbar macht, Zählprüfung gegen die Tabelle | E14 | ein mit `%w` verpackter `*inbound.Fehler` ergibt `ODR-E9001` |
| M4 | Kontextfehler beim Import | Form 4 (`Canceled`, `DeadlineExceeded`, mit `%w` verpackt, in `errors.Join` neben einem Fehler mit Code) × Stelle 10 (die 9 von M3 und die scheiternde Wiederholung mit gleichem Hash) = 40 | E09 | `DeadlineExceeded` in `errors.Join` neben einem Code: Eintrag steht |
| M5 | Kennung im Eintrag | Länge 2 (1 und 64) × Zeichen 37 (`0-9`, `a-z`, `-`) für `SessionID`, `WorkspaceID`, `ImportID` und 65 (`A-Z`, `a-z`, `0-9`, `.`, `_`, `-`) für `ServiceID` × Feld-Stelle 7 (`ServiceID` an drei Stellen, `SessionID` an zwei, `WorkspaceID`, `ImportID`) | E05, E01 | `SessionID` der Aufbewahrung auf 63 Bytes gekürzt |
| M6 | Geheimnis im Eintrag | Feld 7 × Quelle 2 (Fehlertext eines Ports, Userinfo der `UpstreamURL`) × Stelle 4 | E18 | `KorrelationsID` mit dem Fehlertext besetzt |
| M7 | Position in der Aufbewahrung | Position 3 (erste, mittlere, letzte von drei Sessions) × Ergebnis von `Delete` 4 (`nil`, `ErrReferenziert`, `ErrNichtGefunden`, anderer Fehler; die anderen Sessions `nil`) = 12 Läufe | E07, E11 | die Schleife endet bei `ErrNichtGefunden` an der letzten Session |

**Mutationen und Raten** (Implementer, `AGENTS.md` §3.10; gemessen mit `tools/harness/mutationslauf.sh`, `SPEC-210`, Image `odata-recorder:race`, ein Container je Liste, `--cpus=4`, `--memory=2g`, Listen unter dem `flock` auf `/tmp/odata-recorder-heavy.lock`; die Last ist der erste Wert von `/proc/loadavg` des Hosts zu Beginn der Mutation, 7 bis 15 bei 4 CPUs des Containers). Je Mutation ein Prozess und der Test der Zeile rot, bei der Klasse *nebenläufig* 20 Prozesse mit Detektor (`-race`, `-test.count=1`) und 20 ohne. Alle 104 Mutationen der Tabelle sind rot im Test der Zeile (zwei weitere mit `rot-paket` zählen nicht, siehe unten).

| Zeile | Mutationen (Test der Zeile rot) |
|---|---|
| E01 | Feld `ImportID`, `KorrelationsID`, `ErrorCode`, `Art` nicht festgehalten (4 von 4) |
| E02 | Slice statt Kopie; `Eintraege` nicht gezählt; `Append` nicht gezählt; `Aufrufe` ohne `Eintraege` (4 von 4) |
| E03 | `Fehler` nicht geliefert; Eintrag bei Fehler verworfen; `Vor` nach dem Festhalten; `Vor` vor dem Zählen; `Vor` unter der Sperre (Test rot nach der Frist von 5 s); `Vor` ohne den Eintrag (6 von 6) |
| E04 | `Append` ohne Sperre: Detektor 20 von 20, ohne Detektor 16 von 20; Zähler ohne Sperre: Detektor 20 von 20, ohne Detektor 12 von 20. Die Zusage „verliert keinen Eintrag“ gilt über den Detektor (jedes Rennen in jedem Prozess); die Wirkung allein ist kein Fänger in jedem Lauf. |
| E05 | je Stelle `ServiceID`/`SessionID`/`WorkspaceID`/`ImportID` auf 63 Bytes gekürzt, ein Zeichen (`-`, `.`) ersetzt, `KorrelationsID` oder ein verbotenes Feld besetzt, falsche `Art`, `ServiceID` aus dem Argument, Eintrag je entferntem Blob, Quell-Session im Eintrag, `ServiceID` und `SessionID` vertauscht: Session 8, Aufbewahrung 7, Workspace 6, Import 7 (28 von 28) |
| E06 | Eintrag vor `Sessions.Get`, vor der Prüfung auf offen, vor `Sessions.Delete`, bei einem Fehler von `Sessions.Delete` (4 von 4) |
| E07 | Eintrag bei `ErrReferenziert`, bei `ErrNichtGefunden`, bei einem anderen Fehler; die Schleife endet bei `ErrNichtGefunden`; `MaxAge` 0 nicht ausgeschaltet (5 von 5) |
| E08 | Eintrag vor der Prüfung von `ctx`, vor `Workspaces.Delete`, bei einem Fehler von `Workspaces.Delete` (3 von 3) |
| E09 | Eintrag bei `ODR-E5005`, bei einem Verstoß der Anfrage, bei einem Fehler der Wiederholung, bei Erfolg; Kontextfehler nur mit `==` erkannt; `DeadlineExceeded` nicht erkannt (6 von 6) |
| E10 | Eintrag nach `CleanupBlobs` (Session, Workspace; 2 von 2) |
| E11 | Eintrag vor `Sessions.Delete`; Eintrag in einer Goroutine (Detektor 20 von 20, ohne 20 von 20); ein Eintrag am Ende der Stufe zusätzlich (3 von 3) |
| E12 | Fehler von `Append` als Fehler des Dienstes; Fehler von `Append` ohne Bereinigung (Session 2, Aufbewahrung 1, Workspace 2, Import 1; 6 von 6) |
| E13 | Eintrag nur bei `ctx.Err() == nil` (vier Stellen; 4 von 4) |
| E14 | Code fest `ODR-E9001`; ein mit `%w` verpackter `*inbound.Fehler` ergibt `ODR-E9001`; Rückfall ohne Code `ODR-E3001` (3 von 3) |
| E15 | Eintrag bei `ODR-E7003` des Lesens (Session), bei einem Fehler von `Workspaces.Delete`, Eintrag auch bei Erfolg (Import); der Abschlusseintrag der Aufbewahrung unter E11 ist hier der Fänger des zweiten `Startup` (3 von 3) |
| E16 | Eintrag vor dem Ergebnis des Ports: Session Detektor 20 von 20, ohne 20 von 20; Workspace Detektor 20 von 20, ohne 20 von 20 |
| E17 | die Prüfung von `Audit` entfernt (vier Dienste; 4 von 4) |
| E18 | Fehlertext des Ports als Feld (Session, Aufbewahrung, Workspace), Ursache des `*inbound.Fehler` als Feld, Userinfo der `UpstreamURL` als Feld (Import; 5 von 5) |
| E19 | `Append` in einer Goroutine; das Warten endet bei `ctx.Done()`; `CleanupBlobs` vor `Append` (Session, Workspace); nur die ersten beiden bei Aufbewahrung und Import: je Mutation Detektor 20 von 20, ohne Detektor 20 von 20 (10 Mutationen) |

Zwei Mutationen lieferten `rot-paket` (der Test der Zeile blieb grün) und stehen nicht in der Zählung. Eintrag bei einem Fehler von `Sessions.Delete` unter E15 (Session): das zweite `Delete` endet schon in `Sessions.Get`, die Mutation gehört zu E06 und wurde unter E15 durch den Eintrag bei `ODR-E7003` des Lesens ersetzt (rot). `KorrelationsID` mit `err.Error()` im Import: äquivalent, denn `inbound.Fehler.Error()` trägt den Code und seinen Text, nicht die Ursache; der Fänger von E18 fängt statt dessen die Ursache (rot). Akzeptierte Negative bleiben, mit Grund und Folge, das Ende des Prozesses zwischen Commit und `Append` und ein nie endendes `Append` (E19, `SPEC-234` *Ungeprüft*).

**Mengen, gezählt** (der Test zählt, `AGENTS.md` §3.17): M1: 4 Stellen × 2 Zustände von `ctx` (E13 und E19 je Stelle, E05 mit aktivem ctx). M2: Session 7, Aufbewahrung 8 (3 Ergebnisse von `Delete` an 3 Positionen, `MaxAge` 0, ohne Session, eine Session innerhalb der Frist, Fehler von `Dienste.List`, Fehler von `Sessions.List`), Workspace 3 (ctx nil, `ODR-E7001`, jeder der Fehler von `tabellenFehlerListe`). M3: 12 Stellen des Falls „neu“ × 10 Fehler der Ports = 120, dazu `codeVon` an 16 Werten (7 Codes × nackt und mit `%w`, ein Fehler ohne Code × 2). M4: 15 Stellen × 6 Formen (Art `Canceled`, `DeadlineExceeded` × unverpackt, mit `%w`, in `errors.Join`) = 90, dazu die vier Formen von `kontextFehler()` = 60; eine gelungene Wiederholung schreibt nichts (`TestAuditGelungeneWiederholung`). M5: 37 Zeichen × 2 Längen für `SessionID` (Session, Aufbewahrung), `WorkspaceID` und `ImportID`, 65 × 2 für `ServiceID` (Session, Aufbewahrung, Import). M6: 7 Felder × 4 Stellen und × 12 Stellen des Imports mit zwei Quellen. M7: 12 Läufe.

**Fix-Runde nach dem Review** (Implementer; gleiches Werkzeug und gleiche Last wie oben, 25 Mutationen, je Befund eine Menge und mindestens eine im Review nicht genannte Mutation rot):

| Befund | Menge · Elemente | Fänger | Mutationen rot |
|---|---|---|---|
| M-1 | Dienst × Folge der Konfiguration: Aufbewahrung 3 Dienste × 2 Sessions × 3 Rotationen = 18 Einträge; Import 3 Dienste × 3 Rotationen = 9 Läufe, Kennungen verschieden von den Werten der Mutationen | `TestAuditMehrereDienste` (Aufbewahrung, Import) | erster Dienst, letzter Dienst (je Paket; der letzte war im Review nicht genannt) |
| M-2 | Status 3 (`complete`, `incomplete`, ein Status außerhalb der Menge) und `open` als Gegenprobe | `TestAuditStatusAchse` | nur `complete`; nicht `incomplete`; nur die zwei bekannten Status (der unbekannte Status war nicht genannt) |
| M-3 | Zweige 2 (ohne und mit `Vor`) × 64 Schreiber × 8 Aufrufe und 8 Leser × 50 Runden; Zähler 512 + 400 (+ 512 aus `Vor`) | `TestFakeAuditGleichzeitig` | `Eintraege`, `Aufrufe` ohne Sperre, Schreibzugriff im Zweig `Vor`, Lesen von `eintraege` außerhalb der Sperre (nicht genannt): Detektor je 20 von 20; ohne Detektor 8, 0, 0, 0 von 20 (die Zusage gilt über den Detektor, `AGENTS.md` §3.10) |
| M-4 | Instanz je Träger: record 4, replay 2; Pfad in beiden Modi | `TestAuditEineInstanz`, `TestAuditPfadUndZeile` | eigene Instanz für Session, Import, Workspace; Pfad im Blob-Verzeichnis, anderer Name, ohne `DATA_DIR` (die drei letzten waren nicht genannt) |
| L-1 | Fehler 2 × `Vor` 2 = 4 Paare | `TestFakeAuditFehlerMalZaehlerUndVor` | Zähler nur ohne oder nur mit Fehler, `Vor` nur ohne oder nur mit Fehler (4 von 4) |
| L-2 | Position 3 × ctx 2 = 6 Läufe | `TestAuditSynchronOhneFrist` (Aufbewahrung) | `Append` in einer Goroutine an der ersten, zweiten, dritten Session, Warten bei `ctx.Done()` an der dritten: Detektor 20 von 20, ohne Detektor 20 von 20 |

Akzeptierte Negative der Brücke, mit Grund und Folge: eine zweite Uhr oder ein zweiter Generator für den Adapter (`SPEC-139` Nr. 3) bleibt ungefangen, weil `clock.NewSystemuhr` und `clock.NewZufallsIDs` ohne Zustand sind und die Instanz die Uhr privat hält; Folge: keine im Verhalten, eine Zeile trägt dieselbe Zeit und eine zufällige ID. Die Menge der Sessions der Aufbewahrung (I-1) ist unbeschränkt; geprüft sind drei Sessions an den Positionen erste, mittlere, letzte, weil der Code eine Schleife ohne Zähler führt; Folge einer künftigen Schleife mit Obergrenze: Sessions über der Grenze ohne Eintrag.

**Nachzug nach der Verifikation** (Implementer; 16 Mutationen, alle im Test rot, je Klasse mindestens eine im Verifikationsbericht nicht genannte):

| Befund | Menge · Elemente | Fänger | Mutationen rot |
|---|---|---|---|
| V-1 | gelungene Wiederholung mit gleichem Hash: 1 Weg × 2 Aufrufe | `TestAuditGelungeneWiederholung` | Eintrag im Erfolgszweig von `wiederhole`; Eintrag vor dem Hash-Vergleich (nicht genannt) |
| V-2 | Art 2 × Verpackung 3 = 6 Formen × 15 Stellen = 90 Läufe | `TestAuditKontextfehlerSechsFormen` | verpackter `DeadlineExceeded`; nur `Canceled`; nur `DeadlineExceeded`; `errors.Join` nicht erkannt (die letzten drei nicht genannt) |
| V-3 | Fehler × ctx × 12 Stellen (Text des Rückgabefehlers) | `TestAuditFehlerAendertNichts` (Import) | Fehler von `Append` per `errors.Join` hinten und vorn |
| V-4 | M2 der Aufbewahrung von 5 auf 8 Elemente | `TestAuditPositionUndErgebnis` | junge Session, Fehler von `Dienste.List`, von `Sessions.List`, Eintrag vor der Liste (nicht genannt) |
| V-5 | Raster »leere Kennung«: Session, Workspace | `TestAuditLeereKennung` (je Paket) | Eintrag nur bei leerer `id` |
| V-6 | Logger der Brücke, beide Modi | `TestAuditLoggerDesProzesses` | `slog.Default()`; verworfener Logger (nicht genannt) |

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

**Unabhängig von `welle-mock` / M3 und von M2 (`slice-bootstrap-m2`), weil** alle Aufrufstellen Use Cases sind, die nicht über HTTP erreichbar sein müssen: die Tests rufen sie wie ihre Vorgänger über den Dienst. Dieser Slice ändert `src/bootstrap/` nur durch die Brücke (§6, Ausgang: sie bleibt; `slice-bootstrap-m2` liegt in `done/`, der Zweig trägt seinen Stand von `wurzel.go`). Berühren `slice-telemetry-einbau-gateway` oder ein anderer laufender Slice dieselben Dateien unter `src/hexagon/services/`, laufen beide nacheinander (der Planner nennt die Reihenfolge beim Start des zweiten).

**Start-Ausnahme je Slice** (Linie D, `welle-betrieb` §2): wie `slice-audit-log`; `welle-mock` und M2 sind nicht Bedingung. Kennungsbereich der Linie D (Übergangsregeln Nr. 6): `SPEC-230` bis `SPEC-259`, ADR `0040` bis `0049`, `CO-060` bis `CO-079`.

**Größenmarke** (`SPEC-128`; Faktoren allein dort): Der Block gilt für diesen Slice (19 Zeilen). `pfad`, `produktcode` und `fakes` sind der Teil des Architect, `basis`, `schaetzung` und `marke` der des Planner; `basis` ist der Commit, der den Stand nach `slice-audit-log` abschließt (`1543b02`, kein Code danach), sodass Port und Typen aus Teil A nicht in die Messung fallen (`SPEC-128` Nr. 10 g).

```groessenmarke
pfad: src/hexagon/ports/support
pfad: src/hexagon/services/session
pfad: src/hexagon/services/maintenance
pfad: src/hexagon/services/workspace
pfad: src/hexagon/services/baselineimport
produktcode: 260
fakes: 2
basis: 1543b02dd1fa577f3dc3f268bfb87a6e8a880c2e
schaetzung: 2195
marke: 3000
```

**Schätzung und Schnittentscheidung.** Zählung nach `SPEC-128` Nr. 1: 19 Zeilen, davon nebenläufig 3 (E04, E16, E19), andere 16; Fakes 2 (`FakeAudit` mit `Vor` und das gemeinsame Protokoll der Aufrufe über die Fakes der Dienste); Produktcode 260 (vier Konfig-Felder mit Eintrag und `FakeAudit`). *S* = 85·3 + 80·16 + 200·2 + 260 = 2195, *M* = 3000; die Schnittfrage stellt sich nicht (*S* unter 2223, Abstand 28). E19 war als Negativ mit der Klasse `andere` gezählt und ist ein nebenläufiger Fänger (*S* 2190 → 2195); die Fenster F2 und F3 liegen in E10 und E11, die schon die Reihenfolge zusagten, und es kam keine Zeile hinzu und keine ist gebündelt (`SPEC-128` Nr. 10 b). Jede weitere Zeile (80 oder 85) löst die Schnittfrage aus; sie geht vor dem Code als Rückgabe an den Planner (`schnitt-grund` oder Schnitt), nie als Bündelung in eine bestehende Zeile. Der Slice liefert keinen Leser fremder Bytes. Vergleichswert: `slice-audit-log` Teil A, Ende ÷ *S* 0,87 (Messung 1895 bei *S* 2185); Gegenbild `slice-telemetry-port` 1,44. Wächst der Slice über 3000, steht der Eintrag `ueberschreitung` mit Grund im Block, keine Rückführung (`SPEC-128` Nr. 5 d).

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

- Löschen und Eintrag sind zwei Schritte; scheitert einer, fehlt entweder die Löschung oder ihr Nachweis (Fenster zwischen Commit der Löschung und Schreiben des Eintrags) — **Ausgang:** entfallen (Closure): entschieden in `SPEC-234` Nr. 5 (Eintrag nach der Tatsache, Fehler ohne Wirkung, Kontext ohne Wirkung, Warten ohne Frist); Fänger E06 bis E13 und E19 (Haltepunkt), von Review und Verifikation mit eigenen Mutationen bestätigt; benanntes Negativ nur das Ende des Prozesses im Fenster (E19, Folge: eine Löschung ohne Eintrag).
- Der Bericht der Wartung nennt keine Kennungen der gelöschten Sessions; die Löschung der Aufbewahrung geht über `SessionRepository`, nicht über `SessionUseCase` (`SPEC-123`) — **Ausgang:** entfallen (Architect, 2026-11-06): der Bericht bleibt unverändert, die Wartung ruft den Port je Session (`SPEC-123` Nr. 6, `SPEC-234` Nr. 4 b); Fänger E07, E11.
- Doppelung der Einträge zwischen Use Case und Admin-Adapter — **Ausgang:** weiter offen bis `welle-admin-api`; Adresse: `slice-admin-sessions`, `slice-admin-workspaces`, `slice-admin-imports` (Geber: dieser Slice und `slice-audit-log`). Entschieden in `SPEC-234` Nr. 7 und `SPEC-005`: die vier Arten schreibt allein der Use Case, den dieser Slice liefert; die Nehmer tragen den Verweis seit der Anlage dieses Plans in §1 oder DoD.
- Diese Aufrufstellen sind die einzige Quelle der vier Arten: eine neue Löschstelle braucht eine Zeile in `SPEC-234` Nr. 4 und im Fänger (`SPEC-234` *Ungeprüft*) — **Ausgang:** entfallen (Closure): Reviewer und Verifier haben die Aufrufe von `Sessions.Delete`, `Workspaces.Delete` und `Imports` im Paket gelesen (je Art genau eine Stelle, `SPEC-234` Nr. 4); das Risiko einer künftigen neuen Löschstelle steht in `SPEC-234` *Ungeprüft*.
- Konflikt mit anderen Slices an `src/hexagon/services/` (Telemetrie-Einbau in `record` und `session`, Linie D) — **Ausgang:** entfallen (Closure): kein Konflikt eingetreten, der Diff liegt in den fünf Pfaden der Marke und in `src/bootstrap/` (Brücke); die Regel »nacheinander« bleibt in §4 für den Start der Telemetrie-Einbauten. Stand beim Start: `slice-telemetry-einbau-gateway` liegt in `open/` und startet nach diesem Slice; in `in-progress/` berührt kein anderer Slice `src/hexagon/services/session`, `maintenance`, `workspace`, `baselineimport` oder `ports/support`.
- Fenster zwischen Löschung und Eintrag, Abbruch nach dem Commit (Rückgabe aus Teil A, Aufträge 1 und 2 in §3) — **Ausgang:** entfallen (Architect, 2026-11-06): entschieden in §3 (drei Fenster F1 bis F3, Zusatztabelle unten), `SPEC-234` Nr. 5 d und Nr. 9 (`FakeAudit.Vor`) und E19; als Negativ bleiben das Ende des Prozesses und das nie endende `Append`, mit Grund und Folge in E19.
- Gate-Kette (Bedingung in §3): ein Bedarf an einer Änderung der Gate-Kette — **Ausgang:** entfallen (Closure): der Architect hat Vertrag und Plan geprüft (§3), der Verifier den Diff `1543b02..303e09b`: keine dieser Dateien ist berührt.
- Größenmarke knapp unter der Schwelle der Schnittfrage (*S* 2195 gegen 2223, Abstand 28) und ein Vergleichswert über der Marke (`slice-telemetry-port`, Ende ÷ *S* 1,44) — **Ausgang:** entfallen (Closure): Messung 2184 von Marke 3000 (`make groessenmarke`), Ende ÷ *S* 0,995, kein Eintrag `ueberschreitung`; die Tabelle blieb bei 19 Zeilen.
- Der Composition Root übergibt `Audit` nicht: `src/bootstrap/wurzel.go` baut die Konfigs von `maintenance`, `workspace`, `session` und `baselineimport` ohne das Feld, `New` weist nil ab (`SPEC-234` Nr. 8), `Build` scheitert mit `ODR-E9001`, und jeder Test, der über `Build` baut, ist rot: 26 Tests unter `src/bootstrap/` (darunter `TestStartReihenfolge`, `TestAbbruchInJederPhase`, `TestZugaengeJeModus`, `TestImportErhaeltWerte`, `TestWorkspaceErhaeltWerte`) und 19 unter `tests/integration/` (darunter `TestDurchstichM2`), also `make test`, `make test-race` und `make gates`. Der Test `workspace_sqlite_test.go` in `src/bootstrap/` baut `maintenance.New` und `workspace.New` selbst und trägt seit diesem Slice ein `FakeAudit` (Test-Anpassung, kein Bootstrap-Code). Der Plan schließt `src/bootstrap/` aus (§1) und weist die Verdrahtung `slice-audit-verdrahtung` zu, das in `open/` liegt; zwischen den beiden Slices gibt es keinen grünen Stand. Ohne Eingriff in `src/bootstrap/` gibt es keinen grünen Stand; damit `make gates` läuft, trägt der Zweig eine **Brücke** aus zwei Commits mit dem Betreff »Brücke:«: `wurzel.go` erzeugt `audit.New` über `<DATA_DIR>/audit.jsonl` mit der einen Uhr, dem einen Generator und dem einen Logger und gibt die eine Instanz allen vier Konfigs, und die zwei Tabellen der Konfig-Felder in `import_workspace_test.go` führen `Audit`. Das ist der Kern von `slice-audit-verdrahtung` (Instanz, Pfad, Konfigs) ohne dessen Fänger V01 bis V05 (Datei im Betrieb, Neustart, Canary, ohne Löschung keine Datei); beide Commits sind getrennt und mit `git revert` rückgängig zu machen, ohne den Rest zu berühren. — **Ausgang:** entfallen (Planner, Closure, 2026-11-06): **die Brücke bleibt.** Gründe: (1) Ohne sie gibt es zwischen den beiden Slices keinen grünen Stand, denn `SPEC-139` Nr. 3 und `SPEC-234` Nr. 8 verlangen das Feld, und `New` weist nil ab. (2) Ein Revert nähme `make gates` die Grundlage dieses Slice. (3) Ein gemeinsamer Neuschnitt hätte die Schätzung 2635 (2195 plus 440) und die Marke 3600 und verlangte einen neuen Plan nach Review und Verifikation. (4) Die Brücke ist durch `TestAuditEineInstanz`, `TestAuditPfadUndZeile` und `TestAuditLoggerDesProzesses` gefangen, ihre Größe liegt außerhalb der Pfade der Marke (117 Zeilen). `slice-audit-verdrahtung` ist verkleinert auf V01, V02, V03 und den Rest von V05 (Marke 500 statt 600), mit dem Eintrag der Brücke in §1, DoD, §3, §4 und §6 (`main`, Commit `d3ea7d3`); V04 ist geliefert. Das Handbuch-Kapitel zur Datei bleibt bei `slice-admin-sessions` (DoD seit der Anlage). Register: `BEO-REPO/schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices`.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster unten steht vor dem ersten Code-Commit, jede Zelle mit Ort oder „gilt nicht“ und Grund (der Architect hat es gegen die Spezifikation gelesen: keine leere Zelle); eine Zelle ohne Ort geht an den Architect — **Ausgang:** entfallen: Review und Verifikation haben das Raster gegen die Spezifikation gelesen und keine Zelle ohne Ort gefunden; die Zellen »leere Kennung« (Session, Workspace) tragen jetzt einen Fänger (`TestAuditLeereKennung`).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Das Argument der Operationen ist die Kennung des Gegenstands (`id`) bzw. der Eintrag; „fehlende Kennung“ gibt es bei einem `string` nicht (er ist die leere), die Zelle zeigt auf die leere.

| Operation | fehlende Kennung | leere Kennung | unbekannte oder fremde Kennung (Löschung scheitert) | Abbruch (`ctx`) oder Schreibfehler der Ablage | Wiederholung (zweites Löschen) | Geheimnis im Gegenstand |
|---|---|---|---|---|---|---|
| Eintrag zu `SessionUseCase.Delete` | gilt nicht: ein `string` hat keinen fehlenden Wert, die leere Kennung steht in der nächsten Spalte | `SPEC-124` Nr. 7 b (`Sessions.Get` liefert `ErrNichtGefunden`, `ODR-E7003`), kein Eintrag: `SPEC-234` Nr. 4 a; Fänger `TestAuditLeereKennung` | `SPEC-124` Nr. 7 b, c, d (`ODR-E7003`, `ODR-E7009`, `ODR-E7004`), kein Eintrag: `SPEC-234` Nr. 4 a und 5 a | `SPEC-234` Nr. 5 b, c, d (Fehler und Abbruch ohne Wirkung, Warten ohne Frist), Schreibfehler: `SPEC-235` Nr. 4 und 5 | `SPEC-234` Nr. 10 (`ODR-E7003`, kein zweiter Eintrag) | `SPEC-234` Nr. 3 (Wertform, kein Freitext) und `SPEC-235` Nr. 6 |
| Eintrag zur Löschung der Aufbewahrung (Start) | gilt nicht: die Kennungen kommen aus `Sessions.List` der Ablage, die Operation hat kein Argument | gilt nicht: eine leere Kennung kann `Sessions.List` nicht liefern (`SPEC-124` Nr. 4: jede Kennung hat 1 bis 64 Bytes) | `SPEC-123` Nr. 6 (`ErrReferenziert`, `ErrNichtGefunden`), kein Eintrag: `SPEC-234` Nr. 4 b | `SPEC-234` Nr. 5 b, c, d (`Startup` bricht nicht ab, der Kontextfehler erscheint am nächsten Port, das Warten auf `Append` hat keine Frist), Schreibfehler: `SPEC-235` Nr. 4 und 5 | `SPEC-234` Nr. 10 (zweites `Startup` ohne neue Daten, `SPEC-123` Nr. 12) | `SPEC-234` Nr. 3 und `SPEC-235` Nr. 6 |
| Eintrag zum Löschen eines Workspace | gilt nicht: wie oben | `SPEC-129` Nr. 9 b (die `id` geht unverändert an `Workspaces.Delete`, `ODR-E7001`), kein Eintrag: `SPEC-234` Nr. 4 c; Fänger `TestAuditLeereKennung` | `SPEC-129` Nr. 9 b (`ODR-E7001`), kein Eintrag: `SPEC-234` Nr. 4 c | `SPEC-234` Nr. 5 b, c, d, Schreibfehler: `SPEC-235` Nr. 4 und 5 | `SPEC-129` Nr. 14 g und `SPEC-234` Nr. 10 (`ODR-E7001`, kein zweiter Eintrag) | `SPEC-234` Nr. 2 (keine `service_id`, kein Zähler) und Nr. 3 |
| Eintrag zu einem fehlgeschlagenen Import | `SPEC-127` Nr. 3 (Anfrage ungültig), kein Eintrag: `SPEC-234` Nr. 4 d | `SPEC-127` Nr. 3 a (leere `ImportID`) und b (leere `ServiceID`), kein Eintrag: `SPEC-234` Nr. 4 d | unbekannter Service `ODR-E5005`: kein Eintrag (`SPEC-234` Nr. 4 d); unbekannte Session `ODR-E7003`, fremder Service `ODR-E5004`: Eintrag mit diesem Code (`SPEC-234` Nr. 4 d) | Kontextfehler: kein Eintrag (`SPEC-234` Nr. 4 d); Fehler des Eintrags: `SPEC-234` Nr. 5 b; Warten auf `Append`: `SPEC-234` Nr. 5 d | `SPEC-234` Nr. 10 (jeder Fehlschlag ein Eintrag; der Fall „bekannt“ keiner) | `SPEC-234` Nr. 2 und 4 d (die Quell-Session des Clients steht nicht im Eintrag), Nr. 3 |

**Zusatztabelle** (`AGENTS.md` §3.16: mehrere Verstöße und Abhängigkeit vom Modus), mit der Spalte *Fänger*:

| Fall | Ort | Fänger |
|---|---|---|
| Fenster F1 zwischen dem Commit der Löschung und `Append` (Session, Aufbewahrung, Workspace; bei `Import` zwischen dem Fehler und `Append`) | `SPEC-234` Nr. 5 d und *Ungeprüft* | E19: `Vor` hält, das Protokoll zeigt die Löschung ohne Eintrag; das Ende des Prozesses ist das Negativ (Folge: eine Löschung ohne Eintrag; bei `Import` fehlt nur der Eintrag, die `import_id` ist frei, weil keine Zeile besteht, `SPEC-004` Nr. 2) |
| Fenster F2 zwischen `Append` und der Bereinigung (Session, Workspace) | `SPEC-234` Nr. 5 a, `SPEC-124` Nr. 7 e, `SPEC-129` Nr. 9 | E10: im Aufruf von `CleanupBlobs` steht der Eintrag; das Ende des Prozesses ist das Negativ (Folge: die Blobs bleiben bis zum nächsten Start, `SPEC-234` Nr. 6; im Audit fehlt nichts) |
| Fenster F3 zwischen `Append` und dem nächsten Aufruf eines Ports (Aufbewahrung) | `SPEC-234` Nr. 4 b | E11: im `Delete` der nächsten Session stehen die Einträge der vorigen; das Ende des Prozesses ist das Negativ (Folge: die übrigen Sessions folgen beim nächsten Start, `SPEC-123` Nr. 6) |
| Fehler von `Append` und beendeter Kontext zugleich | `SPEC-234` Nr. 5 b, c | E13 (Schleife × `Fehler` gesetzt und ungesetzt) |
| Kontextfehler neben einem Fehler mit Code (Import) | `SPEC-234` Nr. 4 d, `SPEC-109` Nr. 6 | E09 (M4, Form `errors.Join`) |
| mehrere Verstöße der Anfrage zugleich (Import) | `SPEC-127` Nr. 3, `SPEC-234` Nr. 4 d | E09 (alle Paare aus (a) bis (c) und alle zugleich) |

Der Modus (`record` und `replay`, `SPEC-139` Nr. 4) und die Dateiebene gehören zu `slice-audit-verdrahtung`; die Dienste kennen keinen Modus (welche der vier der Root im Modus baut, regelt `SPEC-139` Nr. 3); die Zeilen der Sperre des Adapters gehören zu `slice-audit-log`.

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

**Stand der Closure Phase 1** (Planner, Integrationsschritt, 2026-11-06, im Zweig bei HEAD `412405d` plus Merge von `main`, Gates grün; das reine `git mv` nach `done/` folgt als eigener Commit, danach der Merge nach `main`, kein Push). Belege: Vertrag des Architect `0ae0d54` (`SPEC-234` Nr. 5 d, Nr. 9, Fenster F1 bis F3, Mengen M1 bis M7), Implementer bis `afe88ad` und Brücke `5b7636e`, `d210747`; Review-Report `docs/reviews/2026-11-06-slice-audit-dienste-review.md` (`d57a4d7`: 0 HIGH, 5 MEDIUM M-1 bis M-5, 2 LOW, 3 INFO; Fix-Runde `13d9885`, `303e09b`); Verifikation `docs/reviews/2026-11-06-slice-audit-dienste-verifikation.md` (`761083e`: Closure ja nach Nachzug, 0 HIGH, 3 MEDIUM V-1, V-2, V-4, 3 LOW V-3, V-5, V-6, V-7 für den Planner; `make gates` Exit 0, 57 eigene Probe-Mutationen, 14 Zähler); Nachzug des Implementers `fb9d219`, `412405d` (V-1 bis V-6, 16 Mutationen rot). **Kein weiterer Verifier-Lauf** (Durchstich vor Härtung): **der Nachzug nach der Verifikation ist nicht nachverifiziert**; seine Zahlen (90 Läufe zu V-2, M2 auf 8 Elemente, 16 Mutationen) sind die des Implementers.

**Gegenstand:** geliefert (Teil E des Schnitts). `FakeAudit` mit Haltepunkt `Vor` im Paket `supporttest`, das Feld `Audit` in den vier Konfigs (`New` weist nil ab, ruft `Append` nie), die vier Aufrufstellen: `SessionUseCase.Delete`, die Löschung der Aufbewahrung je Session beim Start, `WorkspaceUseCase.Delete` (je ein Eintrag nach der Löschung und vor der Bereinigung) und `ImportUseCase.Import` (Eintrag `import_fehlgeschlagen` mit dem Code des Fehlers; kein Eintrag bei Kontextfehler, bei Wiederholung mit gleichem Hash, bei Erfolg). Der Fehler von `Append` und ein beendeter Kontext ändern das Ergebnis nicht, der Dienst wartet synchron und ohne Frist (`SPEC-234` Nr. 4, 5, 8, 9, 10). Zeilen E01 bis E19, Mengen M1 bis M7 und Raster in §3 und §6. **Das Binary hat sich geändert, durch die Brücke:** `src/bootstrap/wurzel.go` baut `audit.New` über `<DATA_DIR>/audit.jsonl` und gibt die eine Instanz allen Diensten, die der Modus baut (`record` vier, `replay` zwei). `Build` legt die Datei nicht an; sie entsteht mit dem ersten Eintrag. Im Binary löst heute nur die Aufbewahrung beim Start (`RETENTION_MAX_AGE` größer 0, eine Session älter als die Frist) einen Eintrag aus und schreibt damit **`audit.jsonl` im Datenverzeichnis**; eine HTTP-Löschung gibt es noch nicht. **Handbuch:** `make handbuch-check` grün, das Handbuch nennt weder Aufbewahrung, `DATA_DIR` noch die Datei (Verifier, grep), keine Aussage wird falsch. Das Kapitel zu `audit.jsonl` trägt `slice-admin-sessions` (DoD seit der Anlage, mit dem Hinweis auf `slice-bootstrap-m6`); bestätigt, nicht verlegt. **Gate-Kette unverändert:** weder `Makefile`, `harness/`, `Dockerfile`, `tools/gegenproben/`, `.a-check.yml` noch `.golangci.yml` im Diff, kein `//nolint`, kein neuer Fehlercode, Schlüssel oder Kommando.

**Größen:** Messung 2184 Zeilen gegen Marke 3000 und Schätzung 2195 nach `SPEC-128` (Ende ÷ *S* 0,995), Auslastung 0,73; `make groessenmarke` am Stand `412405d`: `ok, Messung 2184 von Marke 3000`. Kein Eintrag `ueberschreitung`, keine Rückführung. Die Brücke (117 Zeilen unter `src/bootstrap/`) liegt außerhalb der fünf Pfade der Marke und ist nicht gezählt; mit ihr wären es rund 2300. Wachstum durch die Runden: 2058 nach der Fix-Runde (Verifier), 2184 nach dem Nachzug.

**Mutationen:** 104 Mutationen des Implementers beim Handoff (Tabelle in §3, zwei weitere `rot-paket` zählen nicht), 25 in der Fix-Runde nach dem Review, 16 im Nachzug nach der Verifikation, je Klasse mindestens eine, die der Befund nicht nannte; Review 35 eigene (13 grün, Befunde M-1 bis M-4), Verifier 57 eigene an nicht genannten Elementen (neun grün, Befunde V-1 bis V-6) und 14 Zähler. Raten der Nebenläufigkeit: Zähler und Wirkungs-Mutationen je 20 Prozesse mit Detektor (`-race`, `-test.count=1`) und 20 ohne, Messender die jeweilige Rolle im Container `odata-recorder:race`, `--cpus=4`, `--memory=2g`, Last des Hosts 3 bis 24 bei 4 CPUs des Containers; mit Detektor überall 20 von 20, ohne Detektor schwankend (0 bis 20 von 20, §3 *Mutationen und Raten*).

**Akzeptierte Negative** (Grund und Folge in §3, §6 und in `SPEC-234` *Ungeprüft*; keine weitere Fix-Runde):

| Punkt | Zusage ohne Fänger | Folge, wenn der Code regrediert |
|---|---|---|
| E19: Ende des Prozesses zwischen Commit und `Append`; ein nie endendes `Append` | kein Test beendet den Prozess; der Fänger beobachtet das Warten, nicht dessen Dauer | eine Löschung ohne Eintrag; der Dienst wartet |
| Brücke: zweite Uhr oder zweiter Generator für den Adapter (Review R-B4, Verifier vK3) | beide Konstruktoren ohne Zustand, die Instanz hält die Uhr privat | keine im Verhalten: eine Zeile trägt dieselbe Zeit und eine zufällige ID |
| Menge der Sessions der Aufbewahrung (Review I-1) | unbeschränkt; geprüft sind drei Sessions an den Positionen erste, mittlere, letzte | eine künftige Schleife mit Obergrenze ließe Sessions über der Grenze ohne Eintrag |
| Zähler in Zweigen ohne gleichzeitigen Aufruf (Verifier zC1 bis zC3, zC8, zC9) | die Dienste halten keinen Zustand, kein Test ruft Wartung oder Import gleichzeitig | ein Datenrennen in einem solchen Zweig bliebe ungesehen (`SPEC-234` *Ungeprüft*) |
| Rate ohne Detektor (E04, M-3) | die Wirkung allein ist kein Fänger in jedem Lauf; die Zusage gilt über den Detektor | ohne `make test-race` bliebe ein fehlendes Locking oft grün |

Äquivalent: `KorrelationsID` mit `err.Error()` im Import (`inbound.Fehler.Error()` trägt keine Ursache); `TestAuditKeinGeheimnis` des Imports bleibt bei Fehlertext im Feld `ImportID` grün, ein anderer Test des Pakets wird rot (Verifier I-1, doppelter Schutz).

**Risiko-Ausgänge (§6):** je Punkt im Plan selbst. Entfallen: Fenster zwischen Löschung und Eintrag und Abbruch nach dem Commit, Bericht der Wartung, Aufrufstellen als einzige Quelle, Konflikt mit anderen Slices, Gate-Kette, Größenmarke, Randformen, **Brücke (Entscheidung des Planner: sie bleibt, Gründe in §6)**. **Weiter offen, bei Adresse:** (1) Doppelung der Einträge zwischen Use Case und Admin-Adapter: `slice-admin-sessions`, `slice-admin-workspaces`, `slice-admin-imports` (Entscheidung `SPEC-234` Nr. 7, die Nehmer tragen den Verweis); (2) das Handbuch-Kapitel zu `audit.jsonl`: `slice-admin-sessions`.

**Nehmer der Brücke:** `slice-audit-verdrahtung` (`open/`, auf `main` nachgezogen, Commit `d3ea7d3`): Ziel, DoD, Tabelle (V04 geliefert, V05 Rest), Marke (320 und 500), Trigger und Ausgang; `welle-betrieb` und die Roadmap tragen die Verkleinerung.

**Steering-Loop-Einträge:**

1. **Ein Schnitt, der das Pflichtfeld eines Dienstes und seine Übergabe im Composition Root trennt, hinterlässt keinen grünen Stand.** Der Plan von `slice-audit-log` wies die Verdrahtung dem Nehmer zu, ohne zu prüfen, ob `Build` nach dem Geber übersetzt und die Tests über den Root grün sind: 26 Tests unter `src/bootstrap/` und 19 unter `tests/integration/` waren rot, bis der Implementer eine Brücke aus zwei Commits einzog; erst Review (M-5) und Verifikation (V-7) trugen den Widerspruch in den Plänen nach. Neuer Eintrag `BEO-REPO/schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices` (1×, `offen`). Vorschlag für die Sichtung in `/plan-welle`: beim Schnitt je Pflichtfeld, das ein Geber einführt, den Ort der Übergabe dem Geber zuweisen, wenn ohne sie `Build` scheitert; die Schnittfrage von `SPEC-128` bleibt davon unberührt.
2. **Der Fix deckte die Achse des Befunds, nicht die Klasse, zum 27. Mal.** Dienst-Achse (erster und letzter Dienst, der mittlere fehlte, M-1), Art × Verpackung der Kontextfehler (vier statt sechs Formen, V-2), die Allaussage »jeder Fehler eines Ports« gegen fünf von acht Elementen (V-4). Evidenz zu `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (verkörpert in `AGENTS.md` §3.17; gezählt, nicht geschärft: der Fehler liegt in der Anwendung, nicht in der Regel).
3. **Zustand des Arguments** (Status der Session `complete`, `incomplete`, außerhalb der Menge) fehlte in E05 (M-2). Evidenz zu `BEO-REPO/zustand-des-arguments-als-dimension`.
4. **Zweig ohne Lauf unter dem Detektor, vierter Vorgang:** die Zweige `Eintraege`, `Aufrufe` und `Vor` des Fakes liefen in keinem nebenläufigen Test (M-3), die Zähler in Zweigen der Dienste ohne gleichzeitigen Aufruf blieben als Negativ stehen. Evidenz zu `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (4 Dateien, über der Schwelle; der Lese-Schritt gehört in die Welle-Closure, der Eintrag trägt die Verkörperung noch nicht, Stand der Evidenz dort).
5. **Rate ohne Detektor schwankt mit der Last** (12 bis 16 von 20, 14, 2, 0, 0 von 20 für Mutationen, die der Detektor je 20 von 20 meldet). Evidenz zu `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (verkörpert in `AGENTS.md` §3.10); der Plan führte die Zahlen mit Messendem und Last, wie die Regel es sagt.
6. **Der Logger der Brücke war weder Fänger noch Negativ** (V-6), Fälle des Vertrags ohne Zeile (V-1, V-3, V-5). Evidenz zu `BEO-REPO/akzeptiertes-negativ-ohne-folge` und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (beide verkörpert).
7. **Der Nachzug nach der Verifikation ist nicht nachverifiziert** (Durchstich vor Härtung, Entscheidung des Users). Keine eigene Beobachtung.
8. **Hinweis aus Teil A, `/dev/shm`:** die Fänger E01 bis E19 nutzen `FakeAudit` ohne Datei, kein Test der Dienste schreibt unter `/dev/shm` (grep über `src/hexagon/services`, `src/bootstrap` und `supporttest`: kein Treffer). Keine Abweichung gemeldet.

**Spec-Lücken:** keine. `SPEC-234` trug jede Nummer im Code (Verifier, Abschnitt *Konformität je Nummer*); was fehlte, waren Fänger an Randformen (V-1 bis V-6) und der Schnitt (Eintrag 1), nicht der Vertrag.

**Kein Feld `liegt in` in diesem Slice:** nichts wurde in einer Regel, einem Sensor oder einer Spec-Stelle verkörpert, was ein Steering-Loop-Eintrag benennt; `SPEC-234` ist der Vertrag der Dienste und des Fakes, keine Verkörperung einer gezählten Regel.

**Register (Phase 1):** Evidenz `slice-audit-dienste` zu `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`, `BEO-REPO/zustand-des-arguments-als-dimension`, `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, `BEO-REPO/akzeptiertes-negativ-ohne-folge` und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`; ein neuer Eintrag `BEO-REPO/schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices`; kein Zähler gesetzt.

**Matrix:** `docs/plan/planning/rueckverfolgung.md` führt `slice-audit-dienste` nach dem Umzug in der Spalte *gebaut (`done/`)* statt *geplant* bei `LH-FA-13`, `LH-QA-04`, `LH-RB-04` und den `SPEC`-Zeilen `SPEC-004`, `SPEC-005`, `SPEC-109`, `SPEC-123`, `SPEC-124`, `SPEC-127`, `SPEC-129` (Zuordnung aus dem Kopf des Plans, kein Beleg für Verhalten); bei `SPEC-005`, `SPEC-123`, `SPEC-124` und `SPEC-127` rückt auch `slice-audit-log` von *geplant* nach *gebaut*, weil deren Stellen mit diesem Slice geliefert sind. `SPEC-234` steht nicht in der Matrix.

**Drei Paarungen:** Anker (kein Feld `liegt in`, nichts verkörpert) · Folge-Slice (`slice-audit-verdrahtung` in `open/`, auf `main` geprüft: Ziel, DoD und Tabelle tragen die Brücke mit der Kennung dieses Slice als Geber; `slice-admin-sessions`, `slice-admin-workspaces`, `slice-admin-imports` für die Doppelung und das Handbuch-Kapitel, in deren §1 oder DoD) · Register (sechs Evidenz-Dateien, ein neuer Eintrag).

**Offener Rest:** das reine `git mv` nach `done/` und der Merge nach `main` sind nicht Teil dieser Phase; kein Push. `make gates` im Hauptbaum nach dem Merge ist der Closure-Trigger aus §5 (Ergebnis im Bericht des Integrationsschritts). Der Nachzug nach der Verifikation ist nicht nachverifiziert.

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

**Architect (2026-11-06), zu `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`:** Die Dienste halten für den Eintrag keinen Zustand; die Fänger mit Goroutinen (E04, E16, E19) laufen unter `make test-race`, und kein Kopf eines Tests sagt zu, der Detektor melde jedes Paar. Wird in diesem Slice dennoch ein Zweig ohne Lauf gefunden, ist es das dritte Auftreten des Eintrags.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
