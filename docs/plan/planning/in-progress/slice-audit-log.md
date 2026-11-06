# Slice slice-audit-log: Port `AuditLog` und Adapter, Audit-Einträge der Löschvorgänge

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

**Berührte Spec-Stellen:** `LH-RB-04.a` (Aufbewahrung: „Löschungen werden im Audit-Log festgehalten“) · `LH-FA-13.a` („Änderungen sind auditierbar“, Audit-Eintrag ohne Tokenwert) · `SPEC-005` (Schlusszeile: Audit-Logging der Schreibendpunkte) · `SPEC-123` Nr. 1, 6 (Aufbewahrung beim Start), Nr. 8 · `SPEC-124` Nr. 1, 7 (Löschen einer Session) · `SPEC-127` Nr. 1, 4 (Import) · `SPEC-129` Nr. 1, 9, 15 (Workspace löschen) · `SPEC-139` Nr. 2 bis 4 (Verdrahtung) · `SPEC-077` (`DATA_DIR`) · `SPEC-004` Nr. 2 (ein fehlgeschlagener Import „wird im Audit-Log festgehalten“) · `SPEC-234` (Vertrag des Ports und der Einträge, neu) · `SPEC-235` (Vertrag des Adapters und der Datei, neu) · `ARC-020`, `ARC-021` · `architecture.md` §Support-Ports.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-audit-log, Worktree /Development/odata-recorder-wt-audit-log, Linie D)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Querschnitts-Port `AuditLog` (Paket `src/hexagon/ports/support`, `ARC-021`) und sein Driven Adapter (`src/adapters/driven/audit`, `ARC-020`) halten fest, was gelöscht wurde: `SessionUseCase.Delete`, die Löschung der Aufbewahrung in `MaintenanceUseCase.Startup` (Stufe Aufbewahrung, `SPEC-123` Nr. 6) und die Löschung eines Workspace durch `WorkspaceUseCase` schreiben je Vorgang einen Audit-Eintrag mit Art des Vorgangs, Gegenstand (Kennung, Service), Zeitpunkt aus `Clock` und Korrelations-ID, ohne Geheimnisse (`LH-RB-04.a`); ein fehlgeschlagener Import (`ImportUseCase`) hinterlässt ebenfalls einen Eintrag, damit seine `import_id` wiederverwendbar bleibt (Schema-Abschnitt der Spezifikation). Die Ablage ist die Datei `audit.jsonl` im Datenverzeichnis, eine JSON-Zeile je Eintrag (`SPEC-235`). Ein Fake im Paket `supporttest` hält jedes Argument fest und zählt jede Operation.

**Umfang und Schnitt (Architect, 2026-11-06):** Die Tabelle in §3 hat 45 Zeilen und liegt über der Schwelle der Schnittfrage; der Vorschlag zum Schnitt in drei Teile steht in §4, die Entscheidung liegt beim Planner.

**Übernimmt (Sendung von `slice-session-use-case`, §1 *Ausdrücklich NICHT*, Stand `done/`):** der Audit-Eintrag der Löschung einer Session in `SessionUseCase.Delete`; die Löschung selbst liegt dort.

**Übernimmt (Sendung von `slice-session-wartung`, §1 *Ausdrücklich NICHT*, Stand `done/`):** der Audit-Eintrag der Löschung durch die Aufbewahrung in `MaintenanceUseCase.Startup`; die Löschung selbst liegt dort. Beide Pläne liegen in `done/` und sind nicht änderbar; die Übernahme steht hier.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Audit-Einträge der Admin-Schreibendpunkte (Anlegen, Beenden, Reset, Restore über HTTP, der `POST` eines Imports als Anfrage) und der Eintrag zu `ODR-E2101` (Token falsch) — `welle-admin-api` (`SPEC-005`, `LH-FA-13.a`); dieser Slice liefert Port und Adapter, jene Slices rufen ihn mit eigenen Arten (`SPEC-234` Nr. 7). Die Einträge einer Löschung und eines fehlgeschlagenen Imports schreibt der Use Case, nie der Adapter (`SPEC-234` Nr. 7). Die Sendung hat noch keinen Nehmer mit Datei (siehe §6).
- Lesen, Abfragen und Exportieren des Audit-Logs — kein Endpunkt, kein Kommando im ersten Abnahmeumfang (`LH-FA-13` verlangt keine Auskunft über das Audit); der Eintrag liegt in der Datei (`SPEC-235`), der Betreiber liest sie mit eigenen Werkzeugen.
- Das Diagnose-Log und die Zähler (`Telemetry`) — `slice-telemetry-port`; ein Audit-Eintrag ist ein Nachweis, kein Diagnose-Log.
- Die Löschung der Blobs nach dem Commit (best effort, `SPEC-123` Nr. 8) — **entschieden: kein Eintrag**, weder eigener noch als Teil des Eintrags der Löschung (`SPEC-234` Nr. 6, Grund und Folge dort).

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

- [ ] Port `AuditLog`, Fake und Adapter: ein Eintrag hat eine feste Form (Art, Gegenstand, Zeitpunkt, Korrelations-ID), wird dauerhaft abgelegt (Ablage und Verhalten bei Schreibfehler nach Entscheidung des Architect), enthält nie ein Geheimnis und übersteht einen Neustart (`LH-RB-04`, `LH-QA-04`).
- [ ] Löschvorgänge: `SessionUseCase.Delete`, die Löschung der Aufbewahrung beim Start und die Löschung eines Workspace schreiben je gelöschtem Gegenstand genau einen Eintrag; ein abgewiesenes Löschen (`ODR-E7004`, `ODR-E7009`, `ODR-E7003`) schreibt keinen Eintrag über eine Löschung, die nicht geschah (`LH-RB-04.a`); ein fehlgeschlagener Import schreibt seinen Eintrag.
- [ ] Verdrahtung im Composition Root und Test: nach einer Löschung steht der Eintrag in der Ablage, ein Canary-Secret (Token, Header) erscheint in keinem Eintrag.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Führt der Slice einen Konfigurationsschlüssel (Ablage des Audit-Logs), einen Fehlercode oder eine für den Anwender sichtbare Datei im Datenverzeichnis ein: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb.
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
| `src/hexagon/ports/support/`, `supporttest/` | update | Port `AuditLog`, `Eintrag`, `Art`, `Arten()` und `FakeAudit` (`SPEC-234`) |
| `src/adapters/driven/audit/` | neu | Adapter `audit.Datei` und die Datei `audit.jsonl` (`SPEC-235`) |
| `src/hexagon/services/session/` | update | Feld `Audit` in der Konfig, Eintrag in `Delete` nach `Sessions.Delete` (`SPEC-124` Nr. 7 d1) |
| `src/hexagon/services/maintenance/` | update | Feld `Audit`, Eintrag je gelöschter Session der Aufbewahrung (`SPEC-123` Nr. 6); der Bericht bleibt unverändert |
| `src/hexagon/services/workspace/` | update | Feld `Audit`, Eintrag beim Löschen des Workspace (`SPEC-129` Nr. 9 b1) |
| `src/hexagon/services/baselineimport/` | update | Feld `Audit`, Eintrag bei fehlgeschlagenem Import (`SPEC-127` Nr. 4 m) |
| `src/bootstrap/` | update | Verdrahtung: eine Instanz des Adapters in allen vier Diensten (`SPEC-139` Nr. 3, 4) |
| `tools/schema/schema.yaml`, `src/adapters/driven/sqlite/migrations/` | entfällt | die Ablage ist eine Datei (`SPEC-235`): kein Schema, keine Migration, `make schema-check` und `make migration-check` bleiben unberührt |
| `spec/spezifikation.md` | update (Architect, getan) | `SPEC-234`, `SPEC-235` und der Nachzug in `SPEC-123`, `SPEC-124`, `SPEC-127`, `SPEC-129`, `SPEC-139`, `SPEC-077`, `SPEC-005`, `SPEC-004` |
| Tests | update, neu | Fakes der Ports; ein Fall je Löschvorgang |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen hat der Architect entschieden, nicht der Implementer.

**Aufträge an den Architect** (Stand 2026-11-06: alle entschieden, Ort je Auftrag):

1. **Menge der Vorgänge.** Vier Arten, geschlossen, mit Pflichtfeldern (`SPEC-234` Nr. 2); die Aufbewahrung schreibt einen Eintrag je Session, der Bericht der Wartung bleibt unverändert und die Wartung ruft den Port (`SPEC-234` Nr. 4 b, `SPEC-123` Nr. 6); das Löschen eines Workspace ist ein Eintrag samt Snapshots; die Bereinigung der Blobs schreibt keinen (`SPEC-234` Nr. 6).
2. **Ablage.** Datei `audit.jsonl` im Datenverzeichnis, anhängend, Rechte `0600`, Zeilenende nach Absturz, ohne Rotation (`SPEC-235` Nr. 2 bis 4 und Entscheidung); kein Schema.
3. **Reihenfolge und Fehler.** Der Eintrag folgt der Löschung, ein Fehler des Eintrags ändert das Ergebnis nicht, ein beendeter Kontext ändert den Eintrag nicht, `Startup` bricht nicht ab (`SPEC-234` Nr. 5 mit Entscheidung und Alternativen).
4. **Inhalt.** Geschlossene Felder, Wertformen ohne Freitext, kein Tokenwert möglich (`SPEC-234` Nr. 2, 3, 7).
5. **Handle für Wartung und Start.** Zeit und Korrelations-ID kommen aus `Clock` und `IDGenerator` des Adapters, je Eintrag, auch bei der Wartung (`SPEC-234` Nr. 7, `SPEC-235` Nr. 4 c, d).
6. **Aufrufstellen als Menge.** Vier Stellen (`SPEC-234` Nr. 4); die Zeilen E05 bis E18 der Tabelle sind Schleifen über sie, der Fake hält jedes Argument fest und zählt jede Operation (`SPEC-234` Nr. 9).
7. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe.** Unten und in §4; Gegenproben gibt es keine (kein Gate ändert sich).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Spalten nach `SPEC-128` Nr. 3 a; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Nr. tragen das Präfix des Teils (`A` Port und Adapter, `E` Fake und die vier Aufrufstellen, `V` Verdrahtung), so dass die Zeilen ohne Umbau auf die Pläne verteilt werden können (§4).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| A01 | Die Menge der Arten ist geschlossen und gleich der Tabelle von `SPEC-234` Nr. 2: vier Arten, je Art fünf Felder Pflicht oder verboten (20 Zellen) | Test von `support.Arten()` gegen ein Orakel aus dem Wortlaut der Tabelle, Schleife über 4 Arten × 5 Felder, Zählprüfung 4 und 20 | Pflichtfeld `session_id` von `session_aufbewahrung` entfernt | andere |
| A02 | `New` weist leeren `Pfad` und nil `Uhr`, `IDs`, `Log` mit `model.ErrUngueltig` ab, ohne den Wert im Text, und berührt die Datei nicht | Schleife über die vier Felder; ein Pfad in einem nicht vorhandenen Verzeichnis: `New` gelingt und legt nichts an | Prüfung von `Log` entfernt | andere |
| A03 | Die Zeile hat die Feldfolge `timestamp`, `correlation_id`, `art`, dann die gesetzten Felder in der Folge der Tabelle, endet mit LF, hat keinen Leerraum und keine Maskierung | Test je Art gegen eine Soll-Zeile aus Bytewerten (vier Zeilen, `AGENTS.md` §3.18), Datei auf Rohbytes gegengelesen | `session_id` vor `service_id` geschrieben | andere |
| A04 | `timestamp` ist `Uhr.Now()` in UTC mit genau neun Nachkommastellen | Fälle 0 ns, 1 ns, 999999999 ns, Jahr 1 und Jahr 9999 | Nachkommastellen auf drei gekürzt | andere |
| A05 | Eine leere `KorrelationsID` ergibt genau eine Kennung aus `IDs.KorrelationsID()`; eine gesetzte gilt unverändert, `IDs` wird nicht gerufen | Zähler von `FestIDs` in beiden Fällen | ID immer vergeben | andere |
| A06 | Eine Art außerhalb der Menge ist `ErrUngueltig`: leer, Großschreibung, Leerraum davor und dahinter, Präfix, um ein Zeichen verlängert; nichts geschrieben, `Uhr` und `IDs` nicht gerufen | Schleife über die Formen; Datei unverändert, beide Zähler 0 | Prüfung der Art erst nach `Uhr.Now()` | andere |
| A07 | Je Art und Feld (20 Zellen) sind ein fehlendes Pflichtfeld und ein gesetztes verbotenes Feld `ErrUngueltig`, die erlaubte Belegung schreibt | Schleife über die 20 Zellen mit Orakel aus A01, je Zelle ein Fall je Richtung | verbotenes Feld wird geschrieben statt abgewiesen | andere |
| A08 | Kennungsform: Länge 1 bis 64, Bytes aus `A-Z`, `a-z`, `0-9`, `.`, `_`, `-`; jedes andere Byte, Länge 0 und Länge 65 sind `ErrUngueltig` | Schleife über 5 Felder × 256 Bytewerte an einer Stelle (65 gültig) und die Längen 0, 1, 64, 65, dazu NUL, LF, `"`, Backslash und Mehrbyte-UTF-8 | Klasse um `/` erweitert | andere |
| A09 | `ErrorCode`: jeder Code aus `model.AlleCodes()` (36) wird als sein Text geschrieben; `0`, ein negativer Wert und der Wert nach dem letzten sind `ErrUngueltig` | Schleife über die 36 Codes und drei Werte, Orakel aus der Tabelle von `SPEC-109` | Prüfung nach oben entfernt | andere |
| A10 | Eine Zeit der Uhr außerhalb der Jahre 1 bis 9999 und die Nullzeit sind `ErrUngueltig`; nichts geschrieben, `IDs` nicht gerufen | Test-Uhr, die Jahr 0, Jahr 10000 und die Nullzeit liefert; Zähler von `FestIDs` 0, Datei unverändert | Bereichsprüfung entfernt | andere |
| A11 | Die Datei entsteht mit dem ersten Eintrag mit den Rechten `0600`; eine vorhandene Datei behält ihre Rechte | Umask 0 im Test (danach wiederhergestellt); vorhandene Datei mit `0640` bleibt `0640` | Anlegen mit `0666` | andere |
| A12 | Ein zweiter Eintrag steht hinter dem ersten; ein neuer Adapter auf derselben Datei hängt an; Vorhandenes bleibt byte-gleich | zwei Adapter nacheinander, Datei vor und nach dem Anhängen verglichen | Öffnen mit Abschneiden | andere |
| A13 | Eine abgebrochene letzte Zeile wird mit einem LF abgeschlossen: leere Datei (kein führendes LF), Datei endet mit LF (kein zweites), Datei endet ohne LF (genau ein LF, der Rest unverändert) | Schleife über die drei Zustände | LF immer vorangestellt | andere |
| A14 | Scheitert das Öffnen (Pfad ist ein Verzeichnis, Elternverzeichnis fehlt), liefert `Append` den Fehler des Betriebssystems (`errors.Is`) und schreibt nichts | beide Fälle; danach wird das Elternverzeichnis angelegt und das nächste `Append` gelingt | Fehler verschluckt (`nil`) | andere |
| A15 | Scheitert das Schreiben (Pfad `/dev/full`), liefert `Append` den Fehler und kehrt zurück | Pfad `/dev/full`: Fehler `ENOSPC` (`errors.Is`), `Append` kehrt innerhalb der Frist zurück | Fehler des Schreibens verworfen | andere |
| A16 | Scheitert `Append`, steht genau ein Eintrag im Log (Stufe, Text, `ereignis`, `error_code`, `grund`, die gesetzten Kennungen), ohne Fehlertext und ohne Pfad | Fehler aus A14, der Pfad trägt ein Canary; Soll-Eintrag exakt, Canary nirgends | Text des Fehlers als Attribut | andere |
| A17 | Bei einem ungültigen Eintrag trägt der Log-Eintrag nur `ereignis` und `error_code` und keinen Wert des Eintrags | Eintrag mit Canary in `service_id` und ungültiger Art; Log durchsucht | Wert des Eintrags im Log | andere |
| A18 | Die Sperre umfasst `Uhr.Now()` bis zur Rückkehr des Schreibens: ein zweiter `Append` ruft `Uhr.Now()` erst nach dem Ende des ersten (Wirkungs-Fänger; der Detektor sieht nichts, jeder Zugriff liegt unter der Sperre) | Test-Uhr, die im ersten Aufruf blockiert; der zweite `Append` in einer Goroutine; der Zähler der Uhr bleibt 1 bis zur Freigabe; Rate je Mutation mit Zahl der Prozesse im Absatz darunter | `Uhr.Now()` vor die Sperre gezogen | nebenläufig |
| A19 | Gleichzeitige `Append`: N vollständige Zeilen, jede gültiges JSON, Zeitpunkte nicht abnehmend, N verschiedene IDs | 64 Goroutinen mit einer Test-Uhr, die bei jedem Aufruf vorrückt; Zeilen gelesen und geprüft | Sperre entfernt (die Zeitpunkte nehmen ab) | nebenläufig |
| A20 | Jeder Rückkehrpfad gibt die Sperre frei: Verstoß der Form, Zeit, Öffnen, Schreiben | Schleife über die vier Pfade; danach gelingt ein gültiges `Append` innerhalb der Frist | Freigabe im Fehlerpfad der Zeit fehlt | nebenläufig |
| A21 | Akzeptiertes Negativ: ein Fehler von `Sync`, ein kurzes Schreiben und ein Fehler des Schließens sind nicht auslösbar; die Lücke zwischen Schreiben und `Sync` (ein zweiter Eintrag dazwischen) ist folgenlos | kein Fänger. Grund: der Adapter ruft `os.File` ohne Naht. Folge: der Fehler käme als Rückgabe und im Log (A16), kein Test sieht ihn; der Dienst ändert sein Ergebnis nicht (E12) | keine | andere |
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
| V01 | Über das `Handle` stehen die Einträge in `<DATA_DIR>/audit.jsonl`: im Modus `record` alle vier Arten, im Modus `replay` Aufbewahrung und Workspace | Schleife über beide Modi und ihre Arten: Löschung über `Sessions()`, `Workspaces()`, fehlgeschlagener Import über `Import()`, Aufbewahrung mit `RetentionMaxAge` beim zweiten `Build`; die Datei gelesen und geparst | Wartung ohne `Audit` verdrahtet | andere |
| V02 | Neustart: ein zweites `Build` auf demselben `DATA_DIR` findet die Zeilen unverändert und hängt an | zwei `Build` nacheinander, Datei vor und nach | Datei beim `Build` abgeschnitten | andere |
| V03 | Ein Canary (Token, Userinfo der `UpstreamURL`, Fehlertext) steht in keinem Byte der Datei | Datei nach allen Canary durchsucht, nach allen vier Arten | Wert aus der Konfiguration als Feld | andere |
| V04 | Alle vier Dienste erhalten dieselbe Instanz des Adapters | Haken (`NeueWartung`, `NeuerImport`, `NeueWorkspaces`, Session-Dienst): `Audit` ist je Konfig dieselbe Instanz | zweite Instanz für den Workspace-Dienst | andere |
| V05 | Ohne Löschung entsteht die Datei nicht, auch nicht beim Hochfahren und Herunterfahren | `DATA_DIR` nach `Hochfahren` und `Herunterfahren` ohne Löschung | Datei in `Build` angelegt | andere |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Art, jedes Feld, jede Zelle der Feldtabelle, jedes Byte der Kennungsform, jeder Code, jeder Fall einer abgewiesenen Löschung), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-workspace-use-case` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-workspace-use-case` (`welle-record-replay`): `WorkspaceUseCase` mit `Delete` besteht heute nicht (`src/hexagon/services/` führt `baselineimport`, `maintenance`, `record`, `session`); ohne ihn fehlt die dritte Aufrufstelle der Löschvorgänge. Das ist die einzige Slice-Abhängigkeit.
- Code, der schon in `done/` liegt: `SessionUseCase.Delete` (`slice-session-use-case`), `MaintenanceUseCase.Startup` mit der Aufbewahrung (`slice-session-wartung`), `ImportUseCase` (`slice-baseline-import`), `Clock` und `IDGenerator` (`slice-clock-idgenerator`).
- Handbuch-Punkt der DoD: siehe §6, Adresse Planner; die Datei `audit.jsonl` ist kein Konfigurationsschlüssel und kein Fehlercode, und kein gebundenes Beispiel gegen das gebaute Binary kann sie erzeugen.

**Unabhängig von `welle-mock` / M3 und von M2 (`slice-bootstrap-m2`, `slice-http-adapter`, `slice-replay-service`), weil** alle Aufrufstellen Use Cases sind, die nicht über HTTP erreichbar sein müssen: die Tests rufen sie wie ihre Vorgänger über den Dienst. Der Composition-Root-Test (DoD Punkt 3) ruft die Löschung über den Handle des Roots, nicht über einen Listener. Der Mock löscht nichts, was nicht schon `WorkspaceUseCase.Delete` löscht. Reihenfolge-Vorsicht, kein Blocker: `slice-bootstrap-m2` ändert `src/bootstrap/` für die Verdrahtung der Dienste; läuft dieser Slice davor, legt er die Verdrahtung des Adapters `AuditLog` an und `slice-bootstrap-m2` übernimmt sie, läuft er danach, ergänzt er sie (der Planner nennt die Reihenfolge beim Start des zweiten). `slice-workspace-use-case` selbst ist der erste Slice von Linie A und liegt vor `slice-replay-service`; dieser Slice verlangt nicht, dass Linie A bis M2 fertig ist.

**Start am 2026-11-06 als Ausnahme je Slice** (Linie D, `welle-betrieb` §2): `slice-workspace-use-case` liegt in `done/`, der Trigger ist eingetreten; `welle-mock` und M2 sind nicht Bedingung (Welle-Datei §2). Der Kennungsbereich der Linie D (Übergangsregeln Nr. 6): `SPEC-230` bis `SPEC-259` (vergeben bis `SPEC-235`; frei ab `SPEC-236`), `ARC-040` bis `ARC-049` (am Startstand keine vergeben, dieser Slice vergibt keine), ADR `0040` bis `0049` (keine vergeben, siehe unten), `CO-060` bis `CO-079`.

**Bedingung für den Architect (Gate-Kette, Übergangsregeln Nr. 7) — geprüft, erfüllt:** `slice-test-determinism` hat `harness/mk/` und die Sensors-Tabelle berührt, und `slice-mutationslauf-werkzeug` (Linie C) läuft. Dieser Slice (Linie D) ändert keine Datei der Gate-Kette (§3: weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` noch der Stempel). Der Vertrag kommt ohne ein neues Ziel, ohne Stufe im `Dockerfile`, ohne Eintrag in `GATE_CHECKS` und ohne Änderung der Sensors-Tabelle aus: die Ablage ist eine Datei, `tools/schema/schema.yaml` und die Migration bleiben unberührt, und `.a-check.yml` deckt `src/adapters/driven/audit` mit der Regel `adapters-driven` schon ab (Import von `model` und `ports-support`). Eine Rückführung wegen der Gate-Kette entfällt.

**Entscheidung ohne ADR (Architect):** Die Wahl der Ablage (Datei statt Tabelle) und die Reihenfolge (Eintrag nach der Löschung, Fehler ohne Wirkung) stehen mit Alternativen, Gründen und Folgen im Vertrag (`SPEC-234` Nr. 5, `SPEC-235` *Entscheidung*), wie `SPEC-129` es für die Service-Kennung hält. Grund: die Entscheidung bindet nur dieses Paket, ändert kein Gate, keine Bibliothek und kein Stratum, und ein ADR trüge dieselben Zeilen ein zweites Mal. Wünscht der Owner ein ADR, trägt es die nächste freie Nummer aus dem Bereich der Linie D und schärft `SPEC-235`.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`): Der Block steht im Zweig mit dem Teil des Architect (`pfad`, `produktcode`, `fakes`); `basis`, `schaetzung` und `marke` setzt der Planner mit `make groessenmarke` in eigenem Commit vor dem ersten Code-Commit. Der Block steht nicht auf `main` (ein Block mit Platzhaltern machte `make groessenmarke` rot).

```groessenmarke
pfad: src/hexagon/ports/support
pfad: src/adapters/driven/audit
pfad: src/hexagon/services/session
pfad: src/hexagon/services/maintenance
pfad: src/hexagon/services/workspace
pfad: src/hexagon/services/baselineimport
pfad: src/bootstrap
produktcode: 590
fakes: 3
```

**Schätzung und Schnittfrage (Architect).** Der Block gilt für die ungeschnittene Tabelle (45 Zeilen). `produktcode` ist die erwartete Größe des Produktcodes ohne Tests: Port, `Eintrag`, `Arten()` und Adapter 290, `FakeAudit` und vier Konfig-Felder mit Eintrag 260, Verdrahtung 40. `fakes` zählt drei: der Fake einer blockierenden Test-Uhr (`A`), `FakeAudit` und das gemeinsame Protokoll der Aufrufe über die Fakes der Dienste (`E`). Der Slice liefert keinen Leser fremder Bytes. Nach `SPEC-128` Nr. 1 (*S* = 85·*a* + 80·*b* + 200·*f* + *p*, *M* = 100·⌈135·*S*/10000⌉) und nach `AGENTS.md` §3.14 (85 je nebenläufige, 40 je andere Zeile, 200 je Fake, mal 1,25, aufgerundet auf volle Hundert; ein Produktcode-Anteil nur beim Leser fremder Bytes) ergibt die Zählung:

| Teil | Zeilen | davon nebenläufig (*a*) | andere (*b*) | *f* | *p* | *S* nach `SPEC-128` | *M* nach `SPEC-128` | Marke nach `AGENTS.md` §3.14 |
|---|---|---|---|---|---|---|---|---|
| ganze Tabelle | 45 | 5 | 40 | 3 | 590 | 4815 | 6600 | 3300 |
| `A` Port und Adapter (A01 bis A21) | 21 | 3 | 18 | 1 | 290 | 2185 | 3000 | 1500 |
| `E` Fake und vier Aufrufstellen (E01 bis E19) | 19 | 2 | 17 | 2 | 260 | 2190 | 3000 | 1600 |
| `V` Verdrahtung (V01 bis V05) | 5 | 0 | 5 | 0 | 40 | 440 | 600 | 300 |
| `E` und `V` zusammen | 24 | 2 | 22 | 2 | 300 | 2630 | 3600 | 1900 |

Nebenläufig sind A18, A19, A20, E04 und E16. Die Schnittfrage stellt sich nach beiden Formeln für die ganze Tabelle (*S* 4815 ≥ 2223; 3300 über 3000). **Entscheidung des Architect: geschnitten, vor dem Code, an der Naht, die dieser Plan vorab nannte** (Port und Adapter gegen die vier Aufrufstellen), und an einer zweiten, kleineren (die Verdrahtung gegen die Dienste); jeder Teil liegt nach `SPEC-128` unter der Schwelle (*S* ≤ 2222) und teilt mit dem nächsten keinen Mechanismus, den er nicht schon geliefert hat: `A` liefert Typen und Adapter, `E` braucht den Port und die Typen, `V` braucht beide. Der Zuschnitt: `A` bleibt `slice-audit-log` (Titel „Port `AuditLog` und Adapter“); `E` ist ein neuer Slice (Vorschlag `slice-audit-dienste`) mit `FakeAudit`, den Feldern `Audit` und den vier Aufrufstellen samt den beiden Sendungen von `slice-session-use-case` und `slice-session-wartung` (§1 *Übernimmt*); `V` ist ein dritter Slice (Vorschlag `slice-audit-verdrahtung`) mit dem Composition Root, der Datei im Datenverzeichnis und dem Canary der DoD. Die neuen Slice-Dateien, die Verteilung der DoD-Punkte und der Sendungen nach `AGENTS.md` §3.13 und die Marken setzt der Planner (§6, Zeile *Schnitt*). **Formelkonflikt, nichts umgesetzt:** nach `AGENTS.md` §3.14 reichte ein Schnitt in zwei Teile (`A` Marke 1500, `E` und `V` Marke 1900); nach `SPEC-128` liegt `E` und `V` zusammen bei *S* 2630 und *M* 3600 und braucht den dritten Schnitt. Die Teile `E` und `V` sind so geschrieben, dass der dritte Schnitt entfällt, wenn der User die Formel von `AGENTS.md` §3.14 gelten lässt (`V` wird der zuletzt in eigenem Commit geschriebene, vorab benannte Teil von `E`, `AGENTS.md` §3.14). Vergleich: `slice-telemetry-port` maß 3093 bei Marke 3000 und der Schätzung 2150; für `A` (*S* 2185, *M* 3000) ist ein Ende über der Marke nach diesem Vergleich möglich, die Folge ist ein Eintrag `ueberschreitung` mit Grund, keine Rückführung. Der gemessene Wert des ersten Slice mit Zeilen, die keine reinen Funktionen sind, steht in dessen §4 nach der Messung.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; liegt die Schätzung × 1,25 über 3000, wurde vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht steht oben.
- `in-progress` → `open` (blockiert): ein Teil verlangt eine Änderung der Gate-Kette (die Bedingung oben ist erfüllt, solange die Ablage eine Datei bleibt). Der Bericht der Wartung trägt die Kennungen nicht mehr als Bedingung: die Wartung ruft den Port je Session (`SPEC-234` Nr. 4 b).

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
- Sendung ohne Nehmer mit Datei: die Audit-Einträge der Admin-Schreibendpunkte und des Zugriffs mit falschem Token (`SPEC-005`, `LH-FA-13.a`) gehören `welle-admin-api`; die Slice-Dateien bestehen seit dem Anlegen in `open/` und führen den Punkt in ihrer DoD: `slice-admin-listener-token` (Eintrag zu `ODR-E2101` ohne Tokenwert), `slice-admin-sessions` und `slice-admin-workspaces` (je Schreibendpunkt), `slice-admin-imports` (jeder `POST`) — **Ausgang:** entfallen (Stand beim Start 2026-11-06, Planner; geprüft nach `AGENTS.md` §3.13: die Nehmer liegen in `open/`, ihr §1 schließt den Punkt nicht aus).
- Doppelung der Einträge zwischen Use Case und Admin-Adapter (`slice-admin-sessions` DoD „ohne Doppelung mit dem Eintrag, den die Löschung im Use Case schreibt“, `slice-admin-workspaces` §6, `slice-admin-imports` „jeder `POST` auditiert“) — **Ausgang:** offen; Adresse: Planner. Entschieden in `SPEC-234` Nr. 7 und `SPEC-005` (Schlusszeile): die vier Arten einer Löschung und eines fehlgeschlagenen Imports schreibt allein der Use Case, der Adapter schreibt sie nie; was der Adapter selbst auditiert (Abweisen, Anlegen, Beenden, Reset, Restore, die Anfrage eines Import-`POST`), sind eigene Arten, die der Vertrag des Adapters in die Tabelle von `SPEC-234` Nr. 2 einträgt, mit der Korrelations-ID seiner Antwort in `KorrelationsID`. Der Architect hat in den `open/`-Plänen der Nehmer nichts geändert (Zweig-Regel): der Planner trägt den Verweis auf `SPEC-234` Nr. 7 in §1 oder DoD der vier Nehmer ein, bevor dieser Slice ihn nennt (`AGENTS.md` §3.13).
- Die Tabelle von [welle-betrieb](../welle-betrieb.md) §4 nennt als Löschvorgänge Session, Aufbewahrung und Workspace; der fehlgeschlagene Import (Schema-Abschnitt der Spezifikation) steht dort nicht — **Ausgang:** entfallen (Stand beim Start 2026-11-06: Zeile `slice-audit-log` in §4 und Closure-Trigger in §3 der Welle-Datei nennen den fehlgeschlagenen Import).
- Ablage als SQLite-Tabelle kollidiert mit der Kette aus `make schema-check` und `make migration-check` (zweite Schemaversion, „Migrationskette fehlt“ in `harness/README.md`) — **Ausgang:** entfallen (Architect, 2026-11-06): die Ablage ist eine Datei (`SPEC-235` *Entscheidung*), `schema.yaml` und die Migration bleiben unberührt.
- Pfad ohne Lauf unter dem Detektor (Schreibfehler der Ablage, Abbruch nach dem Commit) — **Ausgang:** offen (bei der Closure); `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): Schreibfehler hat den Fänger A14, A15 (Pfad `/dev/full`), die Sperre die Wirkungs-Fänger A18 bis A20 (kein Zugriff liegt außerhalb der Sperre, der Detektor meldet nichts), der Abbruch nach dem Commit E13 (Kontext) und das benannte Negativ E19 (Prozessende); Sync-Fehler A21.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster unten steht vor dem ersten Code-Commit, jede Zelle mit Ort oder „gilt nicht“ und Grund; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen (bis zum Review).
- **Schnitt** in drei Teile (§4): `slice-audit-log` behält `A`, zwei neue Slice-Dateien (Vorschlag `slice-audit-dienste` und `slice-audit-verdrahtung`) tragen `E` und `V` samt den DoD-Punkten 2 und 3, den Sendungen von `slice-session-use-case` und `slice-session-wartung` und den Marken — **Ausgang:** offen; Adresse: Planner (neue Dateien, §3.13-Prüfung der Nehmer, Marken). Formelkonflikt `AGENTS.md` §3.14 gegen `SPEC-128` Nr. 1: Adresse User.
- **Handbuch** (DoD Punkt 7): Die Datei `audit.jsonl` im Datenverzeichnis ist eine für den Anwender sichtbare Datei, entsteht aber erst bei einer Löschung, und kein gebundenes Beispiel gegen das gebaute Binary kann sie erzeugen (das Binary kennt `version` und `start`; eine Löschung ohne Admin-Endpunkt löst nur die Aufbewahrung bei vorhandenen Daten aus). Eine Aussage im Handbuch ohne Beispiel verstieße gegen `SPEC-200` Nr. 1 (K1) — **Ausgang:** offen; Adresse: Planner. Vorschlag: das Kapitel folgt mit dem ersten Slice, dessen Binary eine Löschung auslöst (`slice-admin-sessions`), und steht in dessen DoD; hier steht in §7, dass das Binary keine Datei anlegt, solange nichts gelöscht wird (V05).
- **Telemetrie-Abbildung:** `SPEC-231` Nr. 4 (Abbildung der Dienste) und Nr. 7 (Verdrahtung) nennen die Einträge von `SPEC-121`, `SPEC-123`, `SPEC-124` und die Dienste `record`, `session`, `maintenance`; der Eintrag `audit_fehlgeschlagen` (`SPEC-235` Nr. 5, Schlüssel nur aus der Menge `Log`) und der Adapter `audit` sind dort nachzutragen, wenn der Einbau der Telemetrie sie verdrahtet. Der Architect hat `SPEC-231` nicht geändert (Nummern der Linie D, anderer Slice) — **Ausgang:** offen; Adresse: Planner (`slice-telemetry-einbau-*`).
- Gate-Kette (Bedingung aus §4) — **Ausgang:** entfallen: der Vertrag braucht kein neues Ziel, keine Stufe, keinen Eintrag in `GATE_CHECKS`, keine Änderung der Sensors-Tabelle und keine Änderung von `.a-check.yml`.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Das Argument der Operationen ist die Kennung des Gegenstands (`id`) bzw. der Eintrag; „fehlende Kennung“ gibt es bei einem `string` nicht (er ist die leere), die Zelle zeigt auf die leere.

| Operation | fehlende Kennung | leere Kennung | unbekannte oder fremde Kennung (Löschung scheitert) | Abbruch (`ctx`) oder Schreibfehler der Ablage | Wiederholung (zweites Löschen) | Geheimnis im Gegenstand |
|---|---|---|---|---|---|---|
| Eintrag zu `SessionUseCase.Delete` | gilt nicht: ein `string` hat keinen fehlenden Wert, die leere Kennung steht in der nächsten Spalte | `SPEC-124` Nr. 7 b (`Sessions.Get` liefert `ErrNichtGefunden`, `ODR-E7003`), kein Eintrag: `SPEC-234` Nr. 4 a | `SPEC-124` Nr. 7 b, c, d (`ODR-E7003`, `ODR-E7009`, `ODR-E7004`), kein Eintrag: `SPEC-234` Nr. 4 a und 5 a | `SPEC-234` Nr. 5 b, c (Fehler und Abbruch ohne Wirkung), Schreibfehler: `SPEC-235` Nr. 4 und 5 | `SPEC-234` Nr. 10 (`ODR-E7003`, kein zweiter Eintrag) | `SPEC-234` Nr. 3 (Wertform, kein Freitext) und `SPEC-235` Nr. 6 |
| Eintrag zur Löschung der Aufbewahrung (Start) | gilt nicht: die Kennungen kommen aus `Sessions.List` der Ablage, die Operation hat kein Argument | gilt nicht: eine leere Kennung kann `Sessions.List` nicht liefern (`SPEC-124` Nr. 4: jede Kennung hat 1 bis 64 Bytes) | `SPEC-123` Nr. 6 (`ErrReferenziert`, `ErrNichtGefunden`), kein Eintrag: `SPEC-234` Nr. 4 b | `SPEC-234` Nr. 5 b, c (`Startup` bricht nicht ab, der Kontextfehler erscheint am nächsten Port), Schreibfehler: `SPEC-235` Nr. 4 und 5 | `SPEC-234` Nr. 10 (zweites `Startup` ohne neue Daten, `SPEC-123` Nr. 12) | `SPEC-234` Nr. 3 und `SPEC-235` Nr. 6 |
| Eintrag zum Löschen eines Workspace | gilt nicht: wie oben | `SPEC-129` Nr. 9 b (die `id` geht unverändert an `Workspaces.Delete`, `ODR-E7001`), kein Eintrag: `SPEC-234` Nr. 4 c | `SPEC-129` Nr. 9 b (`ODR-E7001`), kein Eintrag: `SPEC-234` Nr. 4 c | `SPEC-234` Nr. 5 b, c, Schreibfehler: `SPEC-235` Nr. 4 und 5 | `SPEC-129` Nr. 14 g und `SPEC-234` Nr. 10 (`ODR-E7001`, kein zweiter Eintrag) | `SPEC-234` Nr. 2 (keine `service_id`, kein Zähler) und Nr. 3 |
| Eintrag zu einem fehlgeschlagenen Import | `SPEC-127` Nr. 3 (Anfrage ungültig), kein Eintrag: `SPEC-234` Nr. 4 d | `SPEC-127` Nr. 3 a (leere `ImportID`) und b (leere `ServiceID`), kein Eintrag: `SPEC-234` Nr. 4 d | unbekannter Service `ODR-E5005`: kein Eintrag (`SPEC-234` Nr. 4 d); unbekannte Session `ODR-E7003`, fremder Service `ODR-E5004`: Eintrag mit diesem Code (`SPEC-234` Nr. 4 d) | Kontextfehler: kein Eintrag (`SPEC-234` Nr. 4 d); Fehler des Eintrags: `SPEC-234` Nr. 5 b | `SPEC-234` Nr. 10 (jeder Fehlschlag ein Eintrag; der Fall „bekannt“ keiner) | `SPEC-234` Nr. 2 und 4 d (die Quell-Session des Clients steht nicht im Eintrag), Nr. 3 |
| Ablage nach Neustart lesen (Nachweis der Dauer) | gilt nicht: das Lesen ist keine Operation des Ports (`SPEC-235`, Entscheidung: kein Lesepfad), es gibt kein Argument | gilt nicht: wie links | gilt nicht: wie links | halbe letzte Zeile nach Absturz: `SPEC-235` Nr. 4 f; Schreibfehler: `SPEC-235` Nr. 4 und 5 | neuer Adapter auf derselben Datei: `SPEC-235` Nr. 7 | `SPEC-235` Nr. 6 |

**Zusatztabelle** (`AGENTS.md` §3.16: mehrere Verstöße und Abhängigkeit vom Modus), mit der Spalte *Fänger*:

| Fall | Ort | Fänger |
|---|---|---|
| Mehrere Verstöße im selben Eintrag (Art und Feld, Feld und Form, Form und Zeit) | `SPEC-234` Nr. 1 (ein `ErrUngueltig`, welcher Verstoß ist nicht zugesagt) | A06, A07, A08, A10: jeder Verstoß einzeln; die Paare sind nicht unterscheidbar und nicht zugesagt |
| Modus `record`: alle vier Arten | `SPEC-139` Nr. 4 (Session-Dienst und Import nur im Modus `record`) | V01, Modus `record` |
| Modus `replay`: Aufbewahrung und Workspace, kein Session-Dienst und kein Import | `SPEC-139` Nr. 4 | V01, Modus `replay` |
| Fenster der Sperre: von der Sperre bis `Uhr.Now()` | `SPEC-235` Nr. 4 b, c | A18 (Wirkungs-Fänger, die Uhr blockiert) |
| Fenster der Sperre: von `Uhr.Now()` bis zum Schreiben | `SPEC-235` Nr. 4 b bis g | A19 (Zeitpunkte nehmen nicht ab) |
| Fenster der Sperre: vom Schreiben bis `Sync` | `SPEC-235` Nr. 4 g, h | A21 (benanntes Negativ: ein Eintrag dazwischen ist folgenlos, das `Sync` deckt beide) |
| Fenster der Sperre: von der Freigabe auf jedem Pfad | `SPEC-235` Nr. 4 | A20 |
| Fenster zwischen dem Commit der Löschung und `Append` | `SPEC-234` Nr. 5 und *Ungeprüft* | E19 (benanntes Negativ), Reihenfolge E10, E11 |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Slice hat viele Fakes (vier Dienste, ein Port): `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) sind offen und gelten für die Fakes; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für Schreibfehler und Abbruch; offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
