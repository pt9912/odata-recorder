# Slice slice-audit-log: Port `AuditLog`, Eintragstypen und Adapter (Teil A des Schnitts)

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

**Berührte Spec-Stellen (Teil A):** `LH-RB-04.a` („Löschungen werden im Audit-Log festgehalten“) · `LH-FA-13.a` (Audit-Eintrag ohne Tokenwert) · `SPEC-234` Nr. 1 bis 3 (Operation, Arten und Felder, Wertformen), Nr. 7 (Herkunft von Zeit und Korrelations-ID, Seite des Adapters) · `SPEC-235` (Vertrag des Adapters und der Datei) · `SPEC-109` (Orakel der Codes für `error_code`) · `SPEC-114` (`Clock`, `IDGenerator`) · `ARC-020`, `ARC-021` · `architecture.md` §Support-Ports. Die Stellen der Dienste (`SPEC-123` Nr. 6, `SPEC-124` Nr. 7, `SPEC-127` Nr. 4, `SPEC-129` Nr. 9, `SPEC-234` Nr. 4 bis 6, 8 bis 10, `SPEC-004` Nr. 2, `SPEC-005`) gehören `slice-audit-dienste`, die der Verdrahtung (`SPEC-139` Nr. 2 bis 4, `SPEC-077`) `slice-audit-verdrahtung` (§4, Schnitt).
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

**Ziel (Teil A des Schnitts in drei Teile, §4):** Der Querschnitts-Port `AuditLog` mit `Eintrag`, `Art` und `Arten()` (Paket `src/hexagon/ports/support`, `ARC-021`) und sein Driven Adapter `audit.Datei` (`src/adapters/driven/audit`, `ARC-020`) halten einen Audit-Eintrag fest: Art, Gegenstand (Kennung, Service), Zeitpunkt aus `Clock` und Korrelations-ID, ohne Geheimnisse (`LH-RB-04.a`). Die Ablage ist die Datei `audit.jsonl` im Datenverzeichnis, eine JSON-Zeile je Eintrag, anhängend, mit Sperre, `Sync` und Fehlerpfad (`SPEC-235`). Der Slice liefert Typen und Adapter, die ein Test gegen den Adapter allein belegt; kein Dienst ruft den Port, und der Composition Root kennt den Adapter nicht.

**Umfang und Schnitt (Planner, 2026-11-06):** Die Tabelle des Architect hatte 45 Zeilen (*S* 4815) und lag über der Schwelle der Schnittfrage (`SPEC-128` Nr. 1: *S* ab 2223). Der Planner schneidet vor dem Code in drei Teile (Gründe und Zahlen in §4): **A** ist dieser Slice (Zeilen A01 bis A21), **E** ist `slice-audit-dienste` (E01 bis E19) und **V** ist `slice-audit-verdrahtung` (V01 bis V05). Die Zeilen E und V stehen in deren Plänen, nicht hier.

**Gibt weiter (Sendung an `slice-audit-dienste`):** die Übernahme der beiden Sendungen, die dieser Slice ursprünglich trug, liegt bei `slice-audit-dienste` (dessen §1 *Übernimmt* trägt sie mit der Kennung dieses Slice als Geber):

- von `slice-session-use-case` (§1 *Ausdrücklich NICHT*, Stand `done/`): der Audit-Eintrag der Löschung einer Session in `SessionUseCase.Delete`; die Löschung selbst liegt dort.
- von `slice-session-wartung` (§1 *Ausdrücklich NICHT*, Stand `done/`): der Audit-Eintrag der Löschung durch die Aufbewahrung in `MaintenanceUseCase.Startup`; die Löschung selbst liegt dort.

Beide Pläne liegen in `done/` und sind nicht änderbar. Dazu geht der Eintrag der Löschung eines Workspace (`WorkspaceUseCase.Delete`) und des fehlgeschlagenen Imports (`ImportUseCase`) an `slice-audit-dienste`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die vier Aufrufstellen (`SessionUseCase.Delete`, `MaintenanceUseCase.Startup`, `WorkspaceUseCase.Delete`, `ImportUseCase.Import`), das Feld `Audit` der vier Konfigs und `FakeAudit` — `slice-audit-dienste` (Zeilen E01 bis E19). Grund: der Schnitt liegt an der Naht, die dieser Plan vorab nannte; die Typen und der Adapter brauchen keinen Dienst, die Dienste brauchen nur den Port.
- Die Verdrahtung im Composition Root (`src/bootstrap/`), die Datei im Datenverzeichnis als Ganzes und der Canary-Test über das `Handle` — `slice-audit-verdrahtung` (Zeilen V01 bis V05). Grund: `slice-bootstrap-m2` ändert dieselben Pakete; die Verdrahtung läuft nach den Diensten und nach ihm (siehe §4).
- Audit-Einträge der Admin-Schreibendpunkte (Anlegen, Beenden, Reset, Restore über HTTP, der `POST` eines Imports als Anfrage) und der Eintrag zu `ODR-E2101` (Token falsch) — `welle-admin-api` (`SPEC-005`, `LH-FA-13.a`); dieser Slice liefert Port und Adapter, jene Slices rufen ihn mit eigenen Arten (`SPEC-234` Nr. 7). Die Einträge einer Löschung und eines fehlgeschlagenen Imports schreibt der Use Case, nie der Adapter (`SPEC-234` Nr. 7); die Nehmer `slice-admin-sessions`, `slice-admin-workspaces`, `slice-admin-imports` und `slice-admin-listener-token` tragen diese Entscheidung mit der Kennung dieses Slice als Geber in §1 oder DoD (§6).
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

- [ ] Port `AuditLog` mit `Eintrag`, `Art` und `Arten()`: die Menge der vier Arten ist geschlossen, je Art ist jedes Feld Pflicht oder verboten, jede Wertform ist byteweise geprüft (`SPEC-234` Nr. 1 bis 3; Zeilen A01, A06 bis A09), ein Verstoß ist `model.ErrUngueltig` und schreibt nichts (`LH-RB-04`, `LH-QA-04`).
- [ ] Adapter `audit.Datei`: ein Eintrag hat die feste Zeile (Art, Gegenstand, Zeitpunkt aus `Clock`, Korrelations-ID), wird anhängend und unter einer Sperre abgelegt, übersteht einen Neustart, enthält nie ein Geheimnis, und ein Fehler der Ablage kommt als Rückgabe und genau einmal im Log (`SPEC-235`; Zeilen A02 bis A05, A10 bis A21).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Führt der Slice einen Konfigurationsschlüssel, einen Fehlercode oder eine für den Anwender sichtbare Datei ein: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Dieser Teil verdrahtet nichts (`src/bootstrap/` bleibt unberührt, Verdrahtung: `slice-audit-verdrahtung`); in §7 steht, dass das Binary unverändert blieb.
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
| `src/hexagon/ports/support/` | update | Port `AuditLog`, `Eintrag`, `Art` und `Arten()` (`SPEC-234` Nr. 1 bis 3); `supporttest/` bleibt unberührt (`FakeAudit`: `slice-audit-dienste`) |
| `src/adapters/driven/audit/` | neu | Adapter `audit.Datei` und die Datei `audit.jsonl` (`SPEC-235`) |
| `src/hexagon/services/session/`, `maintenance/`, `workspace/`, `baselineimport/` | entfällt | Feld `Audit` und die vier Aufrufstellen: `slice-audit-dienste` (Zeilen E01 bis E19) |
| `src/bootstrap/` | entfällt | Verdrahtung einer Instanz in allen vier Diensten: `slice-audit-verdrahtung` (Zeilen V01 bis V05) |
| `tools/schema/schema.yaml`, `src/adapters/driven/sqlite/migrations/` | entfällt | die Ablage ist eine Datei (`SPEC-235`): kein Schema, keine Migration, `make schema-check` und `make migration-check` bleiben unberührt |
| `spec/spezifikation.md` | update (Architect, getan) | `SPEC-234`, `SPEC-235` und der Nachzug in `SPEC-123`, `SPEC-124`, `SPEC-127`, `SPEC-129`, `SPEC-139`, `SPEC-077`, `SPEC-005`, `SPEC-004` |
| Tests | neu | Tests zu `Arten()`, zu den Wertformen und zum Adapter; die blockierende Test-Uhr (A18) als einziger Fake dieses Teils |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen hat der Architect entschieden, nicht der Implementer.

**Aufträge an den Architect** (Stand 2026-11-06: alle entschieden, Ort je Auftrag):

1. **Menge der Vorgänge.** Vier Arten, geschlossen, mit Pflichtfeldern (`SPEC-234` Nr. 2); die Aufbewahrung schreibt einen Eintrag je Session, der Bericht der Wartung bleibt unverändert und die Wartung ruft den Port (`SPEC-234` Nr. 4 b, `SPEC-123` Nr. 6); das Löschen eines Workspace ist ein Eintrag samt Snapshots; die Bereinigung der Blobs schreibt keinen (`SPEC-234` Nr. 6).
2. **Ablage.** Datei `audit.jsonl` im Datenverzeichnis, anhängend, Rechte `0600`, Zeilenende nach Absturz, ohne Rotation (`SPEC-235` Nr. 2 bis 4 und Entscheidung); kein Schema.
3. **Reihenfolge und Fehler.** Der Eintrag folgt der Löschung, ein Fehler des Eintrags ändert das Ergebnis nicht, ein beendeter Kontext ändert den Eintrag nicht, `Startup` bricht nicht ab (`SPEC-234` Nr. 5 mit Entscheidung und Alternativen).
4. **Inhalt.** Geschlossene Felder, Wertformen ohne Freitext, kein Tokenwert möglich (`SPEC-234` Nr. 2, 3, 7).
5. **Handle für Wartung und Start.** Zeit und Korrelations-ID kommen aus `Clock` und `IDGenerator` des Adapters, je Eintrag, auch bei der Wartung (`SPEC-234` Nr. 7, `SPEC-235` Nr. 4 c, d).
6. **Aufrufstellen als Menge.** Vier Stellen (`SPEC-234` Nr. 4); die Zeilen E05 bis E18 der Tabelle sind Schleifen über sie, der Fake hält jedes Argument fest und zählt jede Operation (`SPEC-234` Nr. 9). Gehört zu `slice-audit-dienste`; in diesem Teil ohne Zeile.
7. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe.** Unten und in §4; Gegenproben gibt es keine (kein Gate ändert sich).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Spalten nach `SPEC-128` Nr. 3 a; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Tabelle trägt nur die 21 Zeilen des Teils A; die Zeilen E01 bis E19 und V01 bis V05 des Architect stehen unverändert in `slice-audit-dienste` bzw. `slice-audit-verdrahtung`, die Nr. behalten ihr Präfix (§4).

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
| A21 | Akzeptiertes Negativ: ein Fehler von `Sync`, ein kurzes Schreiben und ein Fehler des Schließens sind nicht auslösbar; die Lücke zwischen Schreiben und `Sync` (ein zweiter Eintrag dazwischen) ist folgenlos | kein Fänger. Grund: der Adapter ruft `os.File` ohne Naht. Folge: der Fehler käme als Rückgabe und im Log (A16), kein Test sieht ihn; der Dienst ändert sein Ergebnis nicht (E12 in `slice-audit-dienste`) | keine | andere |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Art, jedes Feld, jede Zelle der Feldtabelle, jedes Byte der Kennungsform, jeder Code), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-workspace-use-case` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- Teil A hat keine Slice-Abhängigkeit außer dem, was in `done/` liegt: `Clock` und `IDGenerator` (`slice-clock-idgenerator`), die Codes (`SPEC-109`) und `model.ErrUngueltig`. Er ruft keinen Dienst und liegt vor den beiden anderen Teilen.
- `slice-workspace-use-case` (`welle-record-replay`, Stand `done/`) war die Bedingung der dritten Aufrufstelle; sie gehört jetzt `slice-audit-dienste`.
- Folge-Teile, die diesen Slice brauchen: `slice-audit-dienste` startet, wenn dieser Slice in `done/` liegt (Port, `Eintrag`, `Art`); `slice-audit-verdrahtung` startet, wenn `slice-audit-dienste` und `slice-bootstrap-m2` in `done/` liegen (beide ändern `src/bootstrap/`, sie laufen nacheinander).
- Handbuch-Punkt der DoD: siehe §6, Adresse Planner; die Datei `audit.jsonl` ist kein Konfigurationsschlüssel und kein Fehlercode, und kein gebundenes Beispiel gegen das gebaute Binary kann sie erzeugen.

**Unabhängig von `welle-mock` / M3 und von M2 (`slice-bootstrap-m2`, `slice-http-adapter`, `slice-replay-service`), weil** Port und Adapter keinen Dienst und keinen Listener berühren: die Tests rufen den Adapter allein. Dieser Teil ändert `src/bootstrap/` nicht; die Reihenfolge-Vorsicht gegenüber `slice-bootstrap-m2` liegt bei `slice-audit-verdrahtung`, der erst nach ihm startet.

**Start am 2026-11-06 als Ausnahme je Slice** (Linie D, `welle-betrieb` §2): `slice-workspace-use-case` liegt in `done/`, der Trigger ist eingetreten; `welle-mock` und M2 sind nicht Bedingung (Welle-Datei §2). Der Kennungsbereich der Linie D (Übergangsregeln Nr. 6): `SPEC-230` bis `SPEC-259` (vergeben bis `SPEC-235`; frei ab `SPEC-236`), `ARC-040` bis `ARC-049` (am Startstand keine vergeben, dieser Slice vergibt keine), ADR `0040` bis `0049` (keine vergeben, siehe unten), `CO-060` bis `CO-079`.

**Bedingung für den Architect (Gate-Kette, Übergangsregeln Nr. 7) — geprüft, erfüllt:** `slice-test-determinism` hat `harness/mk/` und die Sensors-Tabelle berührt, und `slice-mutationslauf-werkzeug` (Linie C) läuft. Dieser Slice (Linie D) ändert keine Datei der Gate-Kette (§3: weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` noch der Stempel). Der Vertrag kommt ohne ein neues Ziel, ohne Stufe im `Dockerfile`, ohne Eintrag in `GATE_CHECKS` und ohne Änderung der Sensors-Tabelle aus: die Ablage ist eine Datei, `tools/schema/schema.yaml` und die Migration bleiben unberührt, und `.a-check.yml` deckt `src/adapters/driven/audit` mit der Regel `adapters-driven` schon ab (Import von `model` und `ports-support`). Eine Rückführung wegen der Gate-Kette entfällt.

**Entscheidung ohne ADR (Architect):** Die Wahl der Ablage (Datei statt Tabelle) und die Reihenfolge (Eintrag nach der Löschung, Fehler ohne Wirkung) stehen mit Alternativen, Gründen und Folgen im Vertrag (`SPEC-234` Nr. 5, `SPEC-235` *Entscheidung*), wie `SPEC-129` es für die Service-Kennung hält. Grund: die Entscheidung bindet nur dieses Paket, ändert kein Gate, keine Bibliothek und kein Stratum, und ein ADR trüge dieselben Zeilen ein zweites Mal. Wünscht der Owner ein ADR, trägt es die nächste freie Nummer aus dem Bereich der Linie D und schärft `SPEC-235`.

**Größenmarke** (`SPEC-128`; die Faktoren stehen allein dort, Nr. 1 und 2): Der Block gilt für Teil A (21 Zeilen). `pfad`, `produktcode` und `fakes` sind der Teil des Architect, anteilig auf A geschnitten; `basis` ist der Commit des Architect, der den Vertrag abschließt (`1c99ede`, kein Code danach), `schaetzung` und `marke` setzt der Planner vor dem ersten Code-Commit.

```groessenmarke
pfad: src/hexagon/ports/support
pfad: src/adapters/driven/audit
produktcode: 290
fakes: 1
basis: 1c99ede6e9f9b24d2fb6a154513260b88aab306b
schaetzung: 2185
marke: 3000
```

**Schätzung und Schnittentscheidung (Architect, Planner).** `produktcode` des Teils A ist die erwartete Größe des Produktcodes ohne Tests nach dem Architect: Port, `Eintrag`, `Arten()` und Adapter 290 (die Konfig-Felder und `FakeAudit` mit 260 gehören `slice-audit-dienste`, die Verdrahtung mit 40 `slice-audit-verdrahtung`). `fakes` zählt einen: den Fake der blockierenden Test-Uhr (A18); `FakeAudit` und das gemeinsame Protokoll der Aufrufe über die Fakes der Dienste zählen bei `slice-audit-dienste`. Der Slice liefert keinen Leser fremder Bytes (die Datei schreibt er, liest sie nicht). Nach `SPEC-128` Nr. 1 ergibt die Zählung:

| Teil | Zeilen | davon nebenläufig (*a*) | andere (*b*) | *f* | *p* | *S* | *M* |
|---|---|---|---|---|---|---|---|
| ganze Tabelle des Architect | 45 | 5 | 40 | 3 | 590 | 4815 | 6600 |
| `A` Port und Adapter (A01 bis A21), dieser Slice | 21 | 3 | 18 | 1 | 290 | 2185 | 3000 |
| `E` Fake und vier Aufrufstellen (E01 bis E19), `slice-audit-dienste` | 19 | 2 | 17 | 2 | 260 | 2190 | 3000 |
| `V` Verdrahtung (V01 bis V05), `slice-audit-verdrahtung` | 5 | 0 | 5 | 0 | 40 | 440 | 600 |
| `E` und `V` zusammen | 24 | 2 | 22 | 2 | 300 | 2630 | 3600 |

Nebenläufig sind A18, A19, A20, E04 und E16. **Entscheidung des Planner (2026-11-06): geschnitten in drei Teile, vor dem Code.** Gründe: (1) Die ganze Tabelle liegt bei *S* 4815 und damit weit über der Schwelle der Schnittfrage (*S* ab 2223, `SPEC-128` Nr. 1); zwei Teile genügten nicht, denn jeder dürfte höchstens 2222 tragen, zusammen also 4444, die Tabelle trägt 4815. Auch die Zusammenlegung von `E` und `V` hätte *S* 2630 und die Schnittfrage ausgelöst; der User hat am 2026-11-06 entschieden, dass `SPEC-128` gilt, es gibt keine zweite Formel. (2) Die Naht zwischen `A` und `E` ist die, die dieser Plan vorab nannte: `A` liefert Typen und Adapter, `E` braucht nur den Port und die Typen, nie den Adapter; `A` ist allein mit dem Adapter prüfbar. (3) Die Naht zwischen `E` und `V` teilt keinen Mechanismus, den `E` nicht schon geliefert hat: `V` braucht die vier Konfig-Felder von `E` und den Adapter von `A`, ändert aber nur `src/bootstrap/` und läuft nach `slice-bootstrap-m2`, der dieselben Dateien ändert; ein Teil `V` in `E` zu halten hieße, `E` an `slice-bootstrap-m2` zu binden. (4) Ein vierter Schnitt, etwa der Nebenläufigkeit (A18 bis A20) vom Rest des Adapters, teilt dagegen den Mechanismus (Sperre, Datei, Uhr), den der Rest braucht, und ist nicht abtrennbar. Die Slice-Dateien `slice-audit-dienste` und `slice-audit-verdrahtung` legt der Planner an (Entwurf liegt vor; auf `main`, sobald der Planner dort schreibt, §6 Zeile *Schnitt*).

**Naht und Rest.** Die Teile liegen knapp unter der Schwelle (`A` 2185, `E` 2190, Abstand 37 und 32 zu 2222): eine Zeile mehr nach dem Handoff wird eine Überschreitung der Marke, nicht eine Schnittfrage. **Vergleich:** `slice-telemetry-port` maß 3093 bei Marke 3000 und Schätzung 2150 (Ende ÷ *S* 1,44 gegenüber 1,35 der Marke). Für `A` ist nach diesem Vergleich ein Ende bei etwa 3150 möglich, 5 % über der Marke. Das ist keine Rückführung: die Folge ist ein Eintrag `ueberschreitung` mit Grund (`SPEC-128` Nr. 5 d, 10 a), die Faktoren rechnet der Architect aus dem Messpunkt neu (`SPEC-128` Nr. 9). Der gemessene Wert des ersten Slice mit Zeilen, die keine reinen Funktionen sind, steht in dessen §4 nach der Messung. Die Marken von `E` (3000) und `V` (600) setzt der Planner mit deren Block beim Start; hier stehen sie als Prosa, weil beide Basen erst mit dem Start entstehen (`SPEC-128` Nr. 10 d).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke des Blocks ohne Eintrag `ueberschreitung`; die Schnittfrage (`SPEC-128` Nr. 1) ist vor dem ersten Code-Commit entschieden, die Naht steht oben.
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

- Löschen und Eintrag sind zwei Schritte; scheitert einer, fehlt entweder die Löschung oder ihr Nachweis (Fenster zwischen Commit der Löschung und Schreiben des Eintrags) — **Ausgang:** weiter offen; Adresse: `slice-audit-dienste` (Geber: dieser Slice; der Nehmer führt Fänger E06 bis E13 und das Negativ E19 in Tabelle und DoD). Entschieden in `SPEC-234` Nr. 5 (Eintrag nach der Tatsache, Fehler ohne Wirkung, Kontext ohne Wirkung); die Aufrufstellen liegen nicht mehr in diesem Slice.
- Der Bericht der Wartung nennt keine Kennungen der gelöschten Sessions; die Löschung der Aufbewahrung geht über `SessionRepository`, nicht über `SessionUseCase` (`SPEC-123`) — **Ausgang:** entfallen (Architect, 2026-11-06): der Bericht bleibt unverändert, die Wartung ruft den Port je Session (`SPEC-123` Nr. 6, `SPEC-234` Nr. 4 b); Fänger E07, E11 in `slice-audit-dienste`.
- Sendung ohne Nehmer mit Datei: die Audit-Einträge der Admin-Schreibendpunkte und des Zugriffs mit falschem Token (`SPEC-005`, `LH-FA-13.a`) gehören `welle-admin-api`; die Slice-Dateien bestehen seit dem Anlegen in `open/` und führen den Punkt in ihrer DoD: `slice-admin-listener-token` (Eintrag zu `ODR-E2101` ohne Tokenwert), `slice-admin-sessions` und `slice-admin-workspaces` (je Schreibendpunkt), `slice-admin-imports` (jeder `POST`) — **Ausgang:** entfallen (Stand beim Start 2026-11-06, Planner; geprüft nach `AGENTS.md` §3.13: die Nehmer liegen in `open/`, ihr §1 schließt den Punkt nicht aus).
- Doppelung der Einträge zwischen Use Case und Admin-Adapter (`slice-admin-sessions` DoD „ohne Doppelung mit dem Eintrag, den die Löschung im Use Case schreibt“, `slice-admin-workspaces` §6, `slice-admin-imports` „jeder `POST` auditiert“) — **Ausgang:** weiter offen; Adresse: Planner (trägt auf `main` die Entscheidung mit der Kennung dieses Slice als Geber in §1 oder DoD der vier Nehmer ein; danach entfallen). Entschieden in `SPEC-234` Nr. 7 und `SPEC-005` (Schlusszeile): die vier Arten einer Löschung und eines fehlgeschlagenen Imports schreibt allein der Use Case, der Adapter schreibt sie nie; was der Adapter selbst auditiert (Abweisen, Anlegen, Beenden, Reset, Restore, die Anfrage eines Import-`POST`), sind eigene Arten, die der Vertrag des Adapters in die Tabelle von `SPEC-234` Nr. 2 einträgt, mit der Korrelations-ID seiner Antwort in `KorrelationsID`. Der Architect hat in den `open/`-Plänen der Nehmer nichts geändert (Zweig-Regel): der Planner trägt den Verweis auf `SPEC-234` Nr. 7 in §1 oder DoD der vier Nehmer ein, bevor dieser Slice ihn nennt (`AGENTS.md` §3.13).
- Die Tabelle von [welle-betrieb](../welle-betrieb.md) §4 nennt als Löschvorgänge Session, Aufbewahrung und Workspace; der fehlgeschlagene Import (Schema-Abschnitt der Spezifikation) steht dort nicht — **Ausgang:** entfallen (Stand beim Start 2026-11-06: Zeile `slice-audit-log` in §4 und Closure-Trigger in §3 der Welle-Datei nennen den fehlgeschlagenen Import).
- Ablage als SQLite-Tabelle kollidiert mit der Kette aus `make schema-check` und `make migration-check` (zweite Schemaversion, „Migrationskette fehlt“ in `harness/README.md`) — **Ausgang:** entfallen (Architect, 2026-11-06): die Ablage ist eine Datei (`SPEC-235` *Entscheidung*), `schema.yaml` und die Migration bleiben unberührt.
- Pfad ohne Lauf unter dem Detektor (Schreibfehler der Ablage, Abbruch nach dem Commit) — **Ausgang:** offen (bei der Closure); `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): Schreibfehler hat den Fänger A14, A15 (Pfad `/dev/full`), die Sperre die Wirkungs-Fänger A18 bis A20 (kein Zugriff liegt außerhalb der Sperre, der Detektor meldet nichts), Sync-Fehler A21. Der Abbruch nach dem Commit (E13, Kontext) und das benannte Negativ E19 (Prozessende) liegen bei `slice-audit-dienste`.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster unten steht vor dem ersten Code-Commit, jede Zelle mit Ort oder „gilt nicht“ und Grund; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen (bis zum Review).
- **Schnitt** in drei Teile (§4, entschieden vom Planner am 2026-11-06): `slice-audit-log` behält `A`; `slice-audit-dienste` (E01 bis E19, Marke 3000, Sendungen von `slice-session-use-case` und `slice-session-wartung`) und `slice-audit-verdrahtung` (V01 bis V05, Marke 600) tragen den Rest; die Kennungen sind vergeben, die Dateien legt der Planner auf `main` in `open/` an (Welle `welle-betrieb`, Linie D) — **Ausgang:** weiter offen bis dahin; Adresse: Planner. Der Formelkonflikt `AGENTS.md` §3.14 gegen `SPEC-128` Nr. 1 ist vom User am 2026-11-06 zugunsten von `SPEC-128` entschieden und kein Risiko dieses Slice mehr.
- **Handbuch** (DoD, Handbuch-Punkt): Die Datei `audit.jsonl` im Datenverzeichnis ist eine für den Anwender sichtbare Datei, entsteht aber erst bei einer Löschung, und kein gebundenes Beispiel gegen das gebaute Binary kann sie erzeugen (das Binary kennt `version` und `start`; eine Löschung ohne Admin-Endpunkt löst nur die Aufbewahrung bei vorhandenen Daten aus). Eine Aussage im Handbuch ohne Beispiel verstieße gegen `SPEC-200` Nr. 1 (K1) — **Ausgang:** offen; Adresse: Planner. Vorschlag: das Kapitel folgt mit dem ersten Slice, dessen Binary eine Löschung auslöst (`slice-admin-sessions`), und steht in dessen DoD; in diesem Teil steht in §7, dass das Binary unverändert blieb (nichts ist verdrahtet); das Binary legt die Datei erst nach der Verdrahtung an, und dass es sie nicht ohne Löschung anlegt, fängt V05 in `slice-audit-verdrahtung`.
- **Telemetrie-Abbildung:** `SPEC-231` Nr. 4 (Abbildung der Dienste) und Nr. 7 (Verdrahtung) nennen die Einträge von `SPEC-121`, `SPEC-123`, `SPEC-124` und die Dienste `record`, `session`, `maintenance`; der Eintrag `audit_fehlgeschlagen` (`SPEC-235` Nr. 5, Schlüssel nur aus der Menge `Log`) und der Adapter `audit` sind dort nachzutragen, wenn der Einbau der Telemetrie sie verdrahtet. Der Architect hat `SPEC-231` nicht geändert (Nummern der Linie D, anderer Slice) — **Ausgang:** offen; Adresse: Planner (`slice-telemetry-einbau-*`).
- Gate-Kette (Bedingung aus §4) — **Ausgang:** entfallen: der Vertrag braucht kein neues Ziel, keine Stufe, keinen Eintrag in `GATE_CHECKS`, keine Änderung der Sensors-Tabelle und keine Änderung von `.a-check.yml`.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Das Raster des Architect führte vier Zeilen der Dienste (`SessionUseCase.Delete`, Aufbewahrung, `WorkspaceUseCase.Delete`, `ImportUseCase.Import`); sie stehen unverändert in `slice-audit-dienste`. Dieser Teil führt die zwei Operationen, die er liefert: `Append` und `New`. Die Zellen dieser zwei Zeilen hat der Planner aus dem Wortlaut von `SPEC-234` und `SPEC-235` gelegt (jede mit der Nummer, die die Entscheidung trägt); fehlt einer Zelle der Ort, geht sie an den Architect.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch (`ctx`) oder Schreibfehler der Ablage | Wiederholung | Geheimnis im Gegenstand |
|---|---|---|---|---|---|---|
| `Append(Eintrag)` des Ports und des Adapters | Pflichtfeld der Art nicht gesetzt: `model.ErrUngueltig`, `SPEC-234` Nr. 1 und 2, `SPEC-235` Nr. 4 a (Fänger A07) | leere Kennung (Länge 0) und leeres Feld, das verboten ist, bleibt leer: `SPEC-234` Nr. 3 (Länge 1 bis 64) und Nr. 2, `SPEC-235` Nr. 4 a (A07, A08) | Art außerhalb der geschlossenen Menge, Byte außerhalb der Klasse, Code außerhalb von `AlleCodes()`: `SPEC-234` Nr. 2 und 3, `SPEC-235` Nr. 4 a (A06, A08, A09); Zeit der Uhr außerhalb: `SPEC-235` Nr. 4 c (A10) | gilt nicht für den Abbruch: der Port hat keinen `ctx` (`SPEC-234` Nr. 1); Schreibfehler: `SPEC-235` Nr. 4 e bis h und Nr. 5 (A14, A15, A16) | zweiter `Append` mit demselben Eintrag hängt eine zweite Zeile an (kein Abgleich, keine Eindeutigkeit): `SPEC-235` Nr. 4 und 7 (A12) | `SPEC-234` Nr. 3 (Wertform, kein Freitext) und `SPEC-235` Nr. 6 (A03, A17) |
| `audit.New(Config)` | nil in `Uhr`, `IDs` oder `Log`: `model.ErrUngueltig`, `SPEC-235` Nr. 1 (A02) | leerer `Pfad`: `SPEC-235` Nr. 1 (A02) | gilt nicht: `Pfad` hat keine weitere Form; ein Pfad ohne Verzeichnis ist bei `New` kein Fehler, das Verzeichnis prüft erst `Append` (`SPEC-235` Nr. 1, 4 e; A02, A14) | gilt nicht: `New` hat keinen `ctx` und berührt die Datei nicht (`SPEC-235` Nr. 1) | gilt nicht: ein zweiter Adapter auf derselben Datei ist erlaubt (`SPEC-235` Nr. 7, *Ungeprüft* zu zwei Adaptern) | `SPEC-235` Nr. 1 (der Wert steht nicht im Fehlertext, A02) |
| Ablage nach Neustart lesen (Nachweis der Dauer) | gilt nicht: das Lesen ist keine Operation des Ports (`SPEC-235`, Entscheidung: kein Lesepfad), es gibt kein Argument | gilt nicht: wie links | gilt nicht: wie links | halbe letzte Zeile nach Absturz: `SPEC-235` Nr. 4 f; Schreibfehler: `SPEC-235` Nr. 4 und 5 | neuer Adapter auf derselben Datei: `SPEC-235` Nr. 7 | `SPEC-235` Nr. 6 |

**Zusatztabelle** (`AGENTS.md` §3.16: mehrere Verstöße und Abhängigkeit vom Modus), mit der Spalte *Fänger*:

| Fall | Ort | Fänger |
|---|---|---|
| Mehrere Verstöße im selben Eintrag (Art und Feld, Feld und Form, Form und Zeit) | `SPEC-234` Nr. 1 (ein `ErrUngueltig`, welcher Verstoß ist nicht zugesagt) | A06, A07, A08, A10: jeder Verstoß einzeln; die Paare sind nicht unterscheidbar und nicht zugesagt |
| Fenster der Sperre: von der Sperre bis `Uhr.Now()` | `SPEC-235` Nr. 4 b, c | A18 (Wirkungs-Fänger, die Uhr blockiert) |
| Fenster der Sperre: von `Uhr.Now()` bis zum Schreiben | `SPEC-235` Nr. 4 b bis g | A19 (Zeitpunkte nehmen nicht ab) |
| Fenster der Sperre: vom Schreiben bis `Sync` | `SPEC-235` Nr. 4 g, h | A21 (benanntes Negativ: ein Eintrag dazwischen ist folgenlos, das `Sync` deckt beide) |
| Fenster der Sperre: von der Freigabe auf jedem Pfad | `SPEC-235` Nr. 4 | A20 |
| Modus `record` und `replay`, Fenster zwischen dem Commit der Löschung und `Append` | gilt nicht in diesem Teil: kein Dienst, keine Verdrahtung | `slice-audit-dienste` (E10, E11, E19) und `slice-audit-verdrahtung` (V01) tragen die Zeilen des Architect |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Dieser Teil hat einen Fake (die blockierende Test-Uhr); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) sind offen und gelten für sie und, stärker, für die Fakes von `slice-audit-dienste`; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für Schreibfehler und Abbruch; offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
