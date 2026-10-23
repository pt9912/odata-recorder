# Slice slice-session-use-case: SessionUseCase: Anlegen, Beenden, CloseAll, Lesen, Löschen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-02`, `LH-FA-03`, `LH-FA-13`, `LH-RB-04`, `LH-QA-03` (nebenläufige Tests), [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0007](../../adr/0007-race-detektor-im-gate.md) (das Gate fährt `make test-race`), [ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md) (Status `Accepted` seit 2026-10-23: das Zurückstellen der Bereinigung im Betrieb beruht darauf, dass dieser Slice den Status einer Session erst nach `Drain` setzt)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Status von Sessions: Anlegen, Beenden, Wiederherstellung) · `LH-FA-02.c` (Beenden des Prozesses) · `LH-FA-03.a` · `LH-RB-04.a` (Löschen, Bereinigung danach) · `SPEC-005` (Zeilen `GET` und `POST /__admin/sessions`, `close`, `DELETE`) · `SPEC-059` · `SPEC-097`, `SPEC-099`, `SPEC-100`, `SPEC-072` · `SPEC-105` (Nr. 1, 3, 6) · `SPEC-109` (Nr. 4 und 5 mit dem Gegenstand `Session`, Nr. 6; Codes `ODR-E7003`, `ODR-E7004`, `ODR-E7009`, `ODR-E7010`, `ODR-E7011`; die Ausnahme der Wartung in Nr. 5 nennt nur `Startup` und `CleanupBlobs`), `SPEC-111` · `SPEC-114` (Nr. 2 `Clock`, Nr. 6 Form serverseitiger Kennungen, Nr. 9 Form von Client-Kennungen) · `SPEC-121` (Nr. 2d, 13, 14, *Ungeprüft*) · `SPEC-122` (Nr. 8) · `SPEC-123` (Nr. 2 und *Ungeprüft*: Warten auf die Sperre nicht abbrechbar; Nr. 8 und 8a: Pflicht des Aufrufers und Zurückstellen; Nr. 10: Fehler der Wartung) · `ARC-002`, `ARC-003`, `ARC-012`, `architecture.md` §2.2, §2.3 (Port-Design), §2.4, §4.9.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-23.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `SessionUseCase` legt Recording-Sessions an (auch die erste je Service beim Start, `OpenInitialSession`), beendet sie (auch alle beim Herunterfahren, `CloseAll` mit `SHUTDOWN_TIMEOUT`), liest, listet und löscht sie und ruft nach dem Löschen die Blob-Bereinigung von `MaintenanceUseCase`.

**Entstanden aus:** dem Schnitt von `slice-session-wartung` vor dem ersten Code-Commit (Schätzung × 1,25 über 3000, `AGENTS.md` §3.14); [slice-session-wartung](../done/slice-session-wartung.md) trägt `MaintenanceUseCase` samt Port der Bereinigung (`src/hexagon/ports/inbound/maintenance.go`: `Startup`, `CleanupBlobs`, `Wartungsbericht`, `Loeschbericht`) und liegt in `done/`. Die Übergaben aus `slice-session-wartung` §1 zu `SessionUseCase` gehören diesem Slice, sie stehen hier mit ihrer Herkunft.

**Übernimmt (Zeit und Kennungen):** Die Session-Kennung kommt aus `IDGenerator.SessionID()`, Zeit aus `Clock.Now()` (`SPEC-114`); die Form einer vom Aufrufer gelieferten `id` (leer, überlang, Zeichen) legt dieser Slice mit dem Architect **vor dem Code** fest (`SPEC-114` Nr. 9; Sendung von `slice-clock-idgenerator`, über `slice-session-wartung`, dessen §1 sie vor dem Schnitt führte).

**Übernimmt (Drain und Session-Regel):** `SessionUseCase` beendet eine Session mit `RecordUseCase.Drain(ctx, id, SHUTDOWN_TIMEOUT)` über den Inbound Port (`SPEC-121` Nr. 13). `Drain` liest und ändert die Session in der Ablage nicht: unbekannt (`ODR-E7003`) und nicht `open` (`ODR-E7011`) prüft dieser Slice vorher, den Status setzt er danach. `Drain` liefert den Bericht `Abgebrochen` und `NichtAngehaengt` und kehrt erst zurück, wenn kein Exchange der Session mehr läuft und keiner mehr beginnt; die Session wird `complete` genau dann, wenn jeder angehängte Exchange `capture_status=complete` trägt und `NichtAngehaengt` `0` ist, sonst `incomplete` (`LH-FA-13.a`). `CloseAll` darf `Drain` für mehrere Sessions zugleich rufen (Sendung von `slice-record-service`). Der Status wird erst nach der Rückkehr von `Drain` gesetzt, die Session bleibt bis dahin `open`: das Zurückstellen der Blob-Bereinigung im Betrieb (`SPEC-123` Nr. 8a, [ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md)) beruht darauf, dass kein Exchange einen Blob hält, solange keine Session `open` ist; ändert sich das, fällt seine Begründung (Sendung von `slice-session-wartung`, Fänger: der Status ist beim Aufruf von `Drain` noch `open` und wird erst danach gesetzt, und zwar auf jedem Rückkehrpfad von `Drain`, §3 Auftrag 4). Dieser Slice ist der Träger dieser Annahme (§6). `Drain` lebt im Zustand des Dienstes (`SPEC-121` Nr. 13: die Session nimmt „in diesem Dienst“ nichts mehr an): `SessionUseCase` erhält denselben `RecordUseCase` wie der HTTP-Adapter, den der Composition Root verdrahtet (Sendung an `slice-bootstrap-m2`, §6).

**Übernimmt (Bereinigung nach dem Löschen):** `SessionUseCase` löscht eine Session über `SessionRepository` und ruft danach die Blob-Bereinigung von `MaintenanceUseCase` über den Inbound Port, den `slice-session-wartung` liefert (`ARC-002`; `LH-RB-04.a`: unmittelbar nach dem Commit, best effort). Die Operation heißt `CleanupBlobs(ctx) (Loeschbericht, error)` (`SPEC-123`); ein Fehler von ihr und `Zurueckgestellt` ändern das Ergebnis des Löschens nicht — das Löschen ist committed, der Aufrufer protokolliert den Fehler und meldet das Löschen als gelungen (`SPEC-123` Nr. 8, Pflicht des Aufrufers; Sendung von `slice-session-wartung`). Der Fehler der Operation ist ein `*inbound.Fehler` mit `ODR-E3001` oder `ODR-E9001` (`SPEC-123` Nr. 10, Ausnahme von `SPEC-109` Nr. 5); wie das Protokoll des Aufrufers aussieht (Ereignis, Attribute, kein Fehlertext, `LH-QA-04`), steht in `SPEC-123` nicht und ist Auftrag 1. Die Operation wartet auf die Sperre des Wächters und ist dabei nicht durch `ctx` abbrechbar (`SPEC-123` Nr. 2, *Ungeprüft*): wartet der Aufrufer unter einer Frist, überschreitet er sie höchstens um die Dauer eines Abgleichs, und wer nicht warten will, ruft sie in einer eigenen Goroutine (Sendung von `slice-session-wartung`). Der Name der Operation für die erste Session je Service ist `OpenInitialSession` (`ARC-012`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `MaintenanceUseCase` (Startwartung, Integritätsprüfung, Blob-Abgleich, Aufbewahrung nach Alter, Operation der Bereinigung) — [slice-session-wartung](../done/slice-session-wartung.md), in `done/`; hier wird allein `CleanupBlobs` gerufen, `Startup` nie.
- Die Verdrahtung beider Use Cases im Composition Root (Reihenfolge Wartung, `OpenInitialSession`, `CloseAll` beim Herunterfahren; derselbe `RecordUseCase` für `SessionUseCase` und HTTP-Adapter) — `slice-bootstrap-m2`, das die Sendung in §1 und DoD mit der Kennung dieses Slice führt.
- Das Lesen der Exchanges einer Session für die Ansicht (seitenweise, `InspectionUseCase`, `SPEC-105` *Ungeprüft*) — Folgewelle; dieser Slice liest die Exchanges allein, um die Statusregel von `SPEC-121` Nr. 13 anzuwenden.
- Admin-Endpunkte für Sessions (`POST /__admin/sessions`, `close`, `DELETE`) — Folgewelle (Welle §6); die Use Cases sind der Vertrag, den der Admin-Adapter später ruft.
- Audit-Einträge der Löschvorgänge — der Port `AuditLog` entsteht mit der Folgewelle; die Löschung selbst gehört hierher, ihr Audit-Eintrag nicht.
- Das Beenden einer laufenden Aufzeichnung ist `Drain` des `RecordUseCase` (`slice-record-service`); dieser Slice ruft es über den Inbound Port.

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

- [ ] `SessionUseCase`: anlegen (ID serverseitig oder vom Aufrufer in der Form, die der Vertrag festlegt, `ODR-E7010` bei Doppelung), `OpenInitialSession` (erste Session je Service beim Start im Modus `record`), lesen, auflisten.
- [ ] Beenden mit Warten auf laufende Exchanges bis `SHUTDOWN_TIMEOUT` über `RecordUseCase.Drain` und Status `complete` oder `incomplete` (`ODR-E7003`, `ODR-E7011`), der Status erst nach der Rückkehr von `Drain` (Fänger: beim Aufruf von `Drain` ist die Session `open`, `SPEC-123` Nr. 8a), `CloseAll` für mehrere Sessions zugleich.
- [ ] Löschen (`ODR-E7004`, `ODR-E7009`) und danach die Blob-Bereinigung über den Inbound Port von `MaintenanceUseCase` rufen; ihr Fehler und `Zurueckgestellt` ändern das Ergebnis des Löschens nicht, der Fehler wird protokolliert (Sendung von `slice-session-wartung`, `SPEC-123` Nr. 8).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/inbound/` | update | neue Datei `session.go` mit `SessionUseCase` (je Port eine Datei wie `record.go` und `maintenance.go`); `fehler.go` (`Fehler`) bleibt; `MaintenanceUseCase` liegt aus `slice-session-wartung` vor und wird nur gerufen |
| `src/hexagon/services/session/` | neu | Session-Zustandsübergänge, `OpenInitialSession`, `CloseAll`, Löschen mit dem Aufruf der Bereinigung; ruft `RecordUseCase` und `MaintenanceUseCase` über die Inbound Ports, kein anderes Paket unter `services` |
| `spec/spezifikation.md` | update | Vertrag von `SessionUseCase` mit Randformen; Form einer Client-`id`; Nachzüge nach Auftrag 5 |
| `spec/architecture.md` | update, falls der Vertrag sie ändert | Zeile `SessionUseCase` in §2.2, `ARC-002`, `ARC-012`, §2.3 (Port-Design), §2.4, §4.9 (Berichtigung an jeder Stelle, `AGENTS.md` §3.9; ohne Slices und ADR-Bezüge, `AGENTS.md` §3.4) |
| Tests des Pakets | neu | gegen Fakes und, wo die Tabelle es festlegt, gegen die Repositories der SQLite-Datei; Fakes nach der Zählung des Planner (§4): `RecordUseCase` mit `Drain` (hält und scheitert auf Wunsch, liefert den Bericht), `MaintenanceUseCase` (der Port selbst: zwei Operationen, ein Bericht, ein `Fehler`; zählt Aufrufe, hält und scheitert auf Wunsch), `SessionRepository` mit Ereignisprotokoll und Haltepunkten, `RecordingRepository` (`ListBySession`), Fang-Handler für `slog`; `FestUhr` und `FestIDs` liegen in `supporttest`; nebenläufige Fänger nach `AGENTS.md` §3.10 |

Der Vertrag, die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen vor dem ersten Code-Commit (`AGENTS.md` §3.12), nicht vor dem Start. Reihenfolge nach dem Start: Architect (Vertrag, Tabelle, Raster), Planner (Entscheidung über den Schnitt und Marke aus der Tabelle, §4), Implementer. Der Implementer trägt Tabelle *Zusage · Mutation · roter Test* und die Raten (Zahl der Prozesse, Messender) in einen Absatz von §3 ein (`AGENTS.md` §3.10).

- Die Abhängigkeiten laufen über Inbound Ports, nie über das Paket (`architecture.md` §2): `SessionUseCase` → `RecordUseCase.Drain` und `SessionUseCase` → `MaintenanceUseCase.CleanupBlobs` (Blob-Bereinigung nach dem Löschen); `MaintenanceUseCase` ruft keinen Use Case, der Aufrufweg ist zyklenfrei.
- Importe: `session` nur Domain, Ports (`inbound`, `outbound`, `support`) und die Standardbibliothek ohne `net/http`, `net`, `os`, `crypto/tls`, kein anderes Paket unter `services`; `.a-check.yml` bleibt (die Kanten `services → ports-inbound`, `services → ports-outbound`, `services → ports-support` stehen darin, Glob `src/hexagon/services/**`).

**Auftrag an den Architect vor dem ersten Code-Commit** (Suchliste des Planner, kein Vertrag; die Verträge schreibt der Architect):

1. Vertrag von `SessionUseCase` als eigene `SPEC-<NNN>`: Signaturen (Anlegen, `OpenInitialSession`, Lesen, Auflisten, Beenden, `CloseAll`, Löschen), Berichte (was liefert das Löschen: ein eigener Bericht oder keiner, `architecture.md` §2.2 nennt „Löschbericht“), Konfig und Konstruktor samt Prüfung (nil-Ports, Quelle von `SHUTDOWN_TIMEOUT`: Feld der Konfig oder Argument von `CloseAll`), Fehler als `*inbound.Fehler` und ihre Abbildung (`SPEC-109` Nr. 5 mit dem Gegenstand `Session`; `ODR-E7009` hat dort keine Zeile, den Fehler bildet der Use Case aus dem Status; Reihenfolge der Verstöße), `ctx` nil und beendet je Operation (`SPEC-109` Nr. 6), das Protokoll des Aufrufers für den Fehler und `Zurueckgestellt` von `CleanupBlobs` und für Fehler in `CloseAll` (Ereignis, Attribute, kein Fehlertext, `LH-QA-04`; ob dasselbe Ereignis auch für `slice-workspace-use-case` gilt, entscheidet der Architect; ohne seine Entscheidung geht dorthin keine Sendung dieses Slice) und der Aufruf der Bereinigung (synchron oder in eigener Goroutine, `ctx` des Löschens oder ein eigener: das Warten auf die Sperre ist nicht abbrechbar, `SPEC-123` Nr. 2 *Ungeprüft*).
2. Form einer vom Aufrufer gelieferten `id` (leer, überlang, Zeichen; beide Seiten jeder Grenze), `SPEC-114` Nr. 9: im Verhältnis zur serverseitigen Form (`SPEC-114` Nr. 6: `ses-` und 32 Zeichen), zum Pfad `{id}` in `SPEC-005` und zum Header `X-Recorder-Recording-Session`, in dem dieselbe Kennung steht und der nur mit genau einem nichtleeren Wert wählt (`SPEC-121` Nr. 2d).
3. Randformen als Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), jede Zelle mit Ort oder „gilt nicht“ mit Grund; darin mindestens: Beenden einer unbekannten und einer nicht offenen Session und die Reihenfolge der Verstöße (`ODR-E7003`, `ODR-E7011`; eine Kennung, die `Drain` nie sah, ist dort kein Fehler, `SPEC-121` Nr. 13); Beenden zum zweiten Mal nacheinander (`ODR-E7011`, nichts geändert, `SPEC-111`); Löschen einer offenen Session, die Quelle einer Baseline ist (`ODR-E7009` gegen `ODR-E7004`), und einer unbekannten (`ODR-E7003`); Fehler von `Drain`, von `ListBySession` und des Statuswechsels; zwei gleichzeitige Beenden derselben Session (der zweite `SetStatus` ist `ErrUngueltigerUebergang`: welchen Code und welchen Status trägt der Verlierer?); `CloseAll` mit mehreren Sessions, mit einem Fehler (an der ersten, einer mittleren und der letzten), ohne offene Session und gleichzeitig mit Anlegen und mit Beenden einer einzelnen Session; Anlegen in den Modi (`LH-FA-13.a`: der Endpunkt steht in allen Modi; `OpenInitialSession` nur im Modus `record`: wer prüft den Modus, der Use Case oder der Composition Root) und für einen unbekannten Service (`ErrDienstUnbekannt`, `SPEC-109` Nr. 5); Fehler der Bereinigung nach dem Löschen und `Zurueckgestellt` (`SPEC-123` Nr. 8: ändert das Ergebnis nicht); `ctx` nil und beendet je Operation und je Stelle der Abfolge.
4. Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) **mit einer Zeile je Pfad, Stelle und Operation, wo der Vertrag „jeder“, „einmal“, „unverändert“, „erst nach“ oder „nach dem Commit“ sagt** (Lerneintrag von `slice-session-wartung`: dort hatte die Tabelle für solche Wörter eine Zeile für den Gutfall, und erst das Review fand W28 bis W30). Für diesen Slice sind das mindestens: (i) der Status wird erst nach der Rückkehr von `Drain` gesetzt, und zwar auf **jedem Rückkehrpfad** von `Drain` (gelungen, Fehler, `ctx` beendet) mit Haltepunkt im Fake von `Drain`: beim Aufruf ist die Session `open`; (ii) `CloseAll` beendet **jede** offene Session (ein Fall je Element, mindestens drei Sessions, der Fehler an erster, mittlerer und letzter Stelle) und ruft `Drain` für mehrere zugleich; (iii) **einmal**: `Drain`, `CleanupBlobs`, `Clock.Now()` und `IDGenerator.SessionID()` je Aufruf genau einmal, `CleanupBlobs` nur **nach** dem Commit des Löschens und nie nach `ODR-E7003`, `ODR-E7009`, `ODR-E7004`; (iv) **unverändert**: Bericht und Fehler von `Drain` und `CleanupBlobs`, Kontextfehler ohne Code, ein Fall je Operation und je gerufenem Port; (v) hält der Vertrag eine Sperre im Use Case, eine Zeile je Rückkehrpfad (wie W28 in `slice-session-wartung`). Je Fenster einer Abfolge aus Prüfen und Handeln eine Zeile mit Haltepunkt an einem Argument des Tests (`ctx`, Fake), Wirkungs-Fänger oder benanntem Negativ (`AGENTS.md` §3.10, `SPEC-122` Nr. 8b): Session lesen und `open` prüfen → `Drain`; `Drain` → Exchanges lesen → Status setzen; zwei Beenden derselben Session; `CloseAll` gegen Beenden; `CloseAll` gegen Anlegen; Beenden gegen Löschen; zwei Löschen; zwei Anlegen derselben `id`. Jede Allaussage des Vertrags („nach `Drain` läuft kein Exchange“, „jede offene Session“) nennt ihren Fänger über alle Fälle; die Statusregel mit einem Fall je Element der Menge (`complete`, `incomplete`, `failed`; `NichtAngehaengt` `0` und `1`; keine Exchanges); akzeptierte Negative mit Grund und Folge im Betrieb (`AGENTS.md` §3.15), darunter das nicht abbrechbare Warten und ein Exchange, der nach dem Statuswechsel noch angenommen wird (`SPEC-121` *Ungeprüft*).
5. Entscheidungen zum Beenden, die kein Plan trifft und die nicht erst im Code fallen dürfen (`AGENTS.md` §3.12): (a) Quelle der Statusregel — `RecordingRepository.ListBySession` liefert alle Exchanges einer Session auf einmal (Speicher bei großen Sessions; `SPEC-105` *Ungeprüft*), entweder Grenze mit Grund und Folge im Betrieb oder eine andere Quelle; (b) Zustand nach Fehler von `Drain` oder des Statuswechsels, auch bei beendetem `ctx` (der Bericht von `Drain` kommt dann mit `nil`, `SPEC-121` Nr. 13): die Session bleibt `open` in der Ablage, nimmt im Dienst aber keinen Exchange mehr an, und `CleanupBlobs` bleibt bis zum nächsten Start zurückgestellt — gilt das, oder setzt der Statuswechsel unter einem Kontext, den das Herunterfahren nicht beendet (Vorbild: das Anhängen in `SPEC-121` Nr. 9)?; (c) die Frist: `SHUTDOWN_TIMEOUT` als `frist` je `Drain` oder gemeinsam für `CloseAll`, und dass `Drain` nach Ablauf der Frist ohne Frist auf die abgebrochenen wartet (`SPEC-121` Nr. 13), das Herunterfahren also länger als `SHUTDOWN_TIMEOUT` dauern kann (`LH-FA-02.c`); (d) zwei gleichzeitige Beenden derselben Session: Sperre je Session im Use Case oder der Verlierer meldet den Fehler des Statuswechsels.
6. Nachzug an jeder Stelle, die dieselbe Aussage trägt (`AGENTS.md` §3.9; der Planner suchte beim Start `SessionUseCase`, `OpenInitialSession`, `CloseAll` in `spec/`, `harness/`, `docs/` und `slice-session-use-case` in allen Plänen): zum Beenden `LH-FA-13.a`, `LH-FA-02.c`, `SPEC-121` Nr. 13, die Zeilen `close` und `POST` in `SPEC-005`, `ARC-012`, `architecture.md` §2.2 (Zeile `SessionUseCase`), §2.4 und §4.9 (Sequenz: Lesen der Exchanges, `Drain`, danach der Status); zum Löschen `LH-RB-04.a`, die `DELETE`-Zeile in `SPEC-005`, `ARC-002` und `architecture.md` §2.3 (Port-Design nennt für das Löschen einer Session die Prüfung der Baseline-Quellen über `ImportRepository`, `SPEC-105` Nr. 6 macht sie in `Delete` selbst: `ErrReferenziert`; eine der beiden Stellen ist zu berichtigen); zur Annahme von `SPEC-123` Nr. 8a, falls der Vertrag sie ändert. Die Zeilen der Code-Tabelle (`SPEC-072`, `SPEC-097`, `SPEC-099`, `SPEC-100`, `SPEC-111`) stehen im Kopf dieses Plans und sind mit dem Vertrag gegenzulesen.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-session-wartung](../done/slice-session-wartung.md) und [slice-record-service](../done/slice-record-service.md) und [slice-fehlermodell-konfiguration](../done/slice-fehlermodell-konfiguration.md) und [slice-clock-idgenerator](../done/slice-clock-idgenerator.md) liegen in `done/` (beim Start geprüft); `in-progress/` trug keinen anderen Slice, das WIP-Limit von 1 war frei. Die Spezifikation trägt den Vertrag samt Randformen, der Tabelle *Zusage · Fänger* (§3) und dem Raster (§6), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12, §3.15, §3.16); danach setzt der Planner die Marke (*Größenmarke* unten, `AGENTS.md` §3.14). Die Reihenfolge ist Architect, Planner, Implementer: der erste Code-Commit folgt erst auf die gesetzte Marke.

**Größenmarke** (`AGENTS.md` §3.14): **noch nicht gesetzt.** Der Planner setzt sie nach dem Vertrag des Architect auf die Schätzung aus der Tabelle × 1,25, aufgerundet auf volle Hundert. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; `<Basis>` ist der Commit, der Vertrag, Tabelle und Raster vor dem Code abschließt (Dokumente zählen nicht mit).

**Erwartung des Planner beim Start (vorläufig, vor der Tabelle des Architect; `AGENTS.md` §3.14):** Die ältere Erwartung „rund 2000 bis 2100“ (25 Zeilen, davon 5 nebenläufig, 2 Port-Fakes: 5 × 85 + 20 × 40 + 2 × 200 = 1625, × 1,25 = rund 2030) gilt nicht mehr. Sie stammt aus dem Schnitt von `slice-session-wartung` und lag unter drei Werten, die dessen Slice danach gemessen hat: (1) Sie zählte zwei Fakes; die Fenster der Abfolge aus Prüfen und Handeln (§3 Auftrag 4) brauchen Haltepunkte im `SessionRepository` und im Lesen der Exchanges, und der Fang-Handler für `slog` kommt hinzu, zusammen fünf (`slice-session-wartung` zählte nach dem Review ebenfalls fünf). (2) Sie kannte nicht die Zeilen je Pfad, Stelle und Operation, die „jeder“, „einmal“ und „unverändert“ verlangen (§3 Auftrag 4); dort wuchs die Tabelle von 27 auf 30 Zeilen, drei erst durch das Review. (3) Die Formel unterschätzte den Handoff von `slice-session-wartung` um 26 % (2150 gegen gemessen 2712 bei 27 Zeilen und 4 Fakes), weil die Tests die Fakes mit Haltepunkten tragen; der Faktor 1,25 der Marke deckt den Abstand zwischen Handoff und Stand nach dem Review (`AGENTS.md` §3.14), nicht diese Unterschätzung.

Zählung des Planner jetzt: 30 Zeilen (Spanne 25 bis 35), davon 9 nebenläufig (Spanne 7 bis 11: Status erst nach `Drain` mit Haltepunkt, zwei Beenden derselben Session, `CloseAll` mit mehreren Sessions zugleich, `CloseAll` mit einem Fehler und gemeinsame Frist, `CloseAll` gegen Anlegen, Beenden gegen Löschen, zwei Löschen, Zustand unter `-race`), 21 andere, und 5 Port-Fakes (`RecordUseCase`, `MaintenanceUseCase`, `SessionRepository`, `RecordingRepository`, Fang-Handler; `FestUhr` und `FestIDs` zählen nicht): 9 × 85 + 21 × 40 + 5 × 200 = 765 + 840 + 1000 = **2605**, × 1,25 = 3256, aufgerundet **3300**; Spanne 2700 (25 Zeilen, 7 nebenläufig, 4 Fakes) bis 3700 (35 Zeilen, 11 nebenläufig, 5 Fakes). Die Mitte liegt **über 3000**, die Grenze des Schnitts vor dem ersten Code-Commit; die Spanne reicht von unter bis über diese Grenze, die Tabelle entscheidet. Eine Marke setzt der Planner damit noch nicht (§3 Auftrag, Reihenfolge Architect, Planner, Implementer).

**Vorschlag des Planner für den Schnitt (nicht vollzogen, `AGENTS.md` §3.14):** Zeigt die Tabelle des Architect eine Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit entlang der Operationen geschnitten, in zwei Slices desselben Pakets `session`: (A) **Verwaltung** — Anlegen, `OpenInitialSession`, Lesen, Auflisten, Löschen mit dem Aufruf von `CleanupBlobs` (rund 14 Zeilen, 2 nebenläufig, Fakes für `MaintenanceUseCase`, `SessionRepository` und `slog`: 2 × 85 + 12 × 40 + 3 × 200 = 1250, × 1,25 = rund 1600); (B) **Beenden** — `Drain`, Statusregel, `CloseAll` und die Fenster samt der Annahme aus `SPEC-123` Nr. 8a (rund 15 Zeilen, 7 nebenläufig, Fakes für `RecordUseCase` und `RecordingRepository` zusätzlich zu denen aus (A): 7 × 85 + 8 × 40 + 2 × 200 = 1315, × 1,25 = rund 1700). (B) braucht (A) in `done/`, weil beide Konstruktor, Konfig, Fehlerabbildung und die Fakes in `helfer_test.go` teilen; geteilt wird nur, was (A) liefert, nichts wird zweimal geschrieben. Gegen den Schnitt spricht, dass (A) und (B) denselben Vertrag und dasselbe Paket tragen und die Annahme aus `SPEC-123` Nr. 8a erst mit (B) geliefert wird; ohne Schnitt gilt `AGENTS.md` §3.14: lässt sich kein Teil ohne geteilten Mechanismus abtrennen, steht der Grund hier. Folgen eines Schnitts, die der Planner dann nachzieht (`AGENTS.md` §3.9, §3.13): Welle §4 und §5, Roadmap-Drift-Log, die Starts von `slice-baseline-import` und `slice-bootstrap-m2` (sie verlangen `slice-session-use-case` in `done/`; `slice-baseline-import` nennt in §1 nicht, welche Operation es braucht) und die Sendungen von `slice-session-wartung` in §1, die je nach Operation auf (A) und (B) zu verteilen wären.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach dem Vertrag setzt (Metrik oben).
- `in-progress` → `open` (blockiert): `MaintenanceUseCase.CleanupBlobs` verhält sich nicht, wie `SPEC-123` Nr. 8 zusagt, oder `Drain` liefert nicht, was `SPEC-121` Nr. 13 zusagt.

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

- `complete` setzt voraus, dass alle Exchanges der Session `complete` sind und keiner unangehängt blieb (`SPEC-121` Nr. 13, Sendung in §1); ein Exchange, der nach `Drain` abgeschlossen wird, gegen den Statuswechsel (nach der Rückkehr von `Drain` läuft und beginnt keiner mehr, `SPEC-121` Nr. 13) — **Ausgang:** offen; das Gate fährt `-race` (`make test-race`, `SPEC-122`): ein Datenrennen macht es rot, je Prozess gezählt (`-test.count=1`); die Fenster zwischen Prüfen und Handeln (Session lesen und `open` prüfen, `Drain`, Exchanges lesen, Status setzen; zwei Beenden derselben Session) meldet der Detektor nicht (`SPEC-122` Nr. 8b): je Fenster eine Zeile im Raster mit Wirkungs-Fänger oder Negativ (`AGENTS.md` §3.10, §3 Auftrag 4); die nebenläufigen Fänger ordnen mit Kanälen, andere Wirkungen laufen vor der Übergabe mit `-count=20`.
- Eine offene Annahme von `SPEC-123` Nr. 8a: die Session bleibt `open`, bis `Drain` zurückkehrt, und ein Exchange läuft nur in einer offenen Session — **Ausgang:** offen; Träger ist dieser Slice (Sendung von `slice-session-wartung`, dort §6 *weiter offen*; die zweite Hälfte trägt `slice-record-service` in `done/`, `SPEC-121` Nr. 2d). Fänger: der Status ist beim Aufruf von `Drain` noch `open`, auf jedem Rückkehrpfad von `Drain` (§3 Auftrag 4 (i)). Ändert sich eine der beiden Aussagen, fällt die Begründung des Zurückstellens ([ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md)); bei der Closure ist der Ausgang *entfallen* (Fänger rot gesehen) oder *eingetreten*.
- `CloseAll` ruft `Drain` für mehrere Sessions zugleich; ein Fehler einer Session und die gemeinsame Frist `SHUTDOWN_TIMEOUT` (`SPEC-059`) sind im Vertrag nicht entschieden, und `Drain` wartet nach Ablauf der Frist ohne Frist auf die abgebrochenen Exchanges (`SPEC-121` Nr. 13): das Herunterfahren kann länger als `SHUTDOWN_TIMEOUT` dauern (`LH-FA-02.c`) — **Ausgang:** offen; Architect legt die Regel fest (§3 Auftrag 5 (c)), das Raster führt die Zeilen.
- Zustand nach einem Fehler von `Drain` oder des Statuswechsels: `Drain` nimmt die Session dauerhaft aus der Annahme des Dienstes (`SPEC-121` Nr. 13) und liefert seinen Bericht mit `nil` auch bei beendetem `ctx`; scheitert danach der Statuswechsel (etwa am Kontext des Herunterfahrens), bleibt die Session `open` in der Ablage und nimmt keinen Exchange mehr an, `CleanupBlobs` bleibt zurückgestellt, und erst der nächste Start setzt sie auf `incomplete` (`SPEC-123` Nr. 5), auch wenn jeder Exchange `complete` war — **Ausgang:** offen; Architect entscheidet (§3 Auftrag 5 (b)) und nennt Grund und Folge im Betrieb.
- Die Statusregel liest alle Exchanges einer Session auf einmal (`RecordingRepository.ListBySession`); große Sessions kosten Speicher (`SPEC-105` *Ungeprüft*) — **Ausgang:** offen; Architect entscheidet (§3 Auftrag 5 (a)): Grenze mit Grund und Folge im Betrieb oder eine andere Quelle.
- Die Bereinigung nach dem Löschen ist best effort (`LH-RB-04.a`); ob ein Fehler der Bereinigung das Ergebnis des Löschens ändert — **Ausgang:** entfallen; entschieden in `SPEC-123` Nr. 8 (Pflicht des Aufrufers): weder ein Fehler noch `Zurueckgestellt` ändert es, der Aufrufer protokolliert den Fehler. Die Zeile im Raster und ihr Fänger (der Fake von `MaintenanceUseCase` scheitert, das Löschen bleibt gelungen) schreibt der Architect dieses Slice vor dem Code.
- Das Warten von `CleanupBlobs` auf die Sperre des Wächters ist nicht durch `ctx` abbrechbar (`SPEC-123` Nr. 2 *Ungeprüft*): ein Löschen, das die Bereinigung synchron ruft, kehrt erst nach der Freigabe zurück, höchstens die Dauer eines Abgleichs über einer Frist des Aufrufers — **Ausgang:** offen; Architect legt fest, ob der Use Case sie synchron oder in einer eigenen Goroutine ruft und mit welchem `ctx` (§3 Auftrag 1), und führt das Negativ mit Grund und Folge im Betrieb.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit (Auftrag in §3); jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.
- `Drain` wirkt im Zustand des Dienstes, der die Aufnahme annimmt (`SPEC-121` Nr. 13): ein `SessionUseCase`, den der Composition Root mit einem zweiten `RecordUseCase` baut, wartet auf keinen laufenden Exchange und stoppt die Annahme nicht, setzt den Status aber trotzdem; ein laufender Exchange wird dann nicht angehängt (`SPEC-105` Nr. 7) — **Ausgang:** weiter offen; Träger ist `slice-bootstrap-m2` (Sendung in §1 und DoD dort, Kennung dieses Slice).
- Größe (`BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`, 12×, verkörpert in `AGENTS.md` §3.14): Zählung des Planner 30 Zeilen, 9 nebenläufig, 5 Port-Fakes: 2605, × 1,25 = 3300, Spanne 2700 bis 3700, die Mitte über 3000 (§4); die ältere Erwartung „rund 2000 bis 2100“ ist ersetzt (§4) — **Ausgang:** offen; der Planner entscheidet nach der Tabelle des Architect über den Schnitt (Vorschlag in §4, nicht vollzogen) und setzt die Marke.
- Rate ohne Detektor (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 2×, offen): ein nebenläufiger Fänger, der nur mit Wahrscheinlichkeit rot wird (Wirkung ohne Datenrennen), steht in Plan oder Kommentar als Einzelzahl, obwohl die Rate mit der Last des Hosts schwankt (`slice-session-wartung`: W27 ohne Detektor bei `--cpus=0.05` 83 von 88 Prozessen statt „16 von 16“) — **Ausgang:** offen; Plan und Kommentare nennen für solche Fänger die Spanne mehrerer Läufe und den Messenden, die Rate unter dem Detektor (je Prozess) steht als feste Größe; ein weiterer Beleg in diesem Slice wäre der dritte, und der Lese-Schritt der Closure legt den Eintrag dem User vor.
- Bericht des Implementers (`BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo`, 3×, verkörpert in `AGENTS.md` §3.10): der Implementer trägt Zusage, Mutation, roten Test und die Raten (Zahl der Prozesse, Messender) in einen Absatz von §3 ein, damit Reviewer und Verifier ihn haben — **Ausgang:** offen; Review und Verifikation prüfen, ob der Absatz in §3 vorliegt; fehlt er, ist das ein weiterer Beleg gegen die verkörperte Regel und geht an den User.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Register unter `../observations/BEO-REPO/` trägt elf Einträge (Sub-Area `*`); dieser Plan antwortet auf jeden (Zahlen aus den Dateien unter `evidence/`, Ausgänge aus den `state.md`). Der Slice ist nebenläufig (Statuswechsel gegen laufende Exchanges, `CloseAll`) und liefert einen neuen Vertrag:
- `nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, verkörpert in `AGENTS.md` §3.10, Sensor `make test-race`): das Gate fährt `-race` (`SPEC-122`); ein Datenrennen zählt je Prozess, jede andere Wirkung eine Wiederholung mit `-count=20`; je Fenster der Abfolge aus Prüfen und Handeln eine Zeile mit Haltepunkt, Wirkungs-Fänger oder Negativ (`SPEC-122` Nr. 8b); §3 Auftrag 4, §6.
- `berichtigung-erreicht-nicht-jede-stelle` (6×, verkörpert in `AGENTS.md` §3.9): die Liste der Stellen steht in §3 Auftrag 6; die Sendungen dieses Slice an `slice-bootstrap-m2` stehen dort in §1 und DoD mit der Kennung dieses Slice (`AGENTS.md` §3.13), geprüft gegen *Ausdrücklich NICHT* und DoD des Nehmers; `slice-workspace-use-case`, `slice-baseline-import`, `slice-replay-service`, `slice-http-adapter` bekommen keine Sendung.
- `akzeptiertes-negativ-ohne-folge` (6×, verkörpert in `AGENTS.md` §3.15): was der Vertrag ungeprüft lässt, steht mit Grund und Folge im Betrieb; §3 Auftrag 4.
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (20×, verkörpert in `AGENTS.md` §3.15): der Vertrag nennt nur, was ein Fänger der Tabelle prüft; Wörter wie „jeder“, „einmal“, „unverändert“ bekommen eine Zeile je Pfad, Stelle und Operation (§3 Auftrag 4, Lerneintrag von `slice-session-wartung`).
- `randform-erst-im-review-entschieden` (17×, verkörpert in `AGENTS.md` §3.16) und `universalaussage-im-vertrag-ungeprueft` (6×, verkörpert in `AGENTS.md` §3.16): das Raster in §6 vor dem ersten Code-Commit; Allaussagen („nach `Drain` läuft kein Exchange“) nennen ihren Fänger über alle Fälle; die Zelle *Abbruch beim Warten* hat ihren Ort in `SPEC-123` Nr. 2 und *Ungeprüft*.
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (12×, verkörpert in `AGENTS.md` §3.14): Marke in §4, Erwartung über 3000 in der Mitte, Vorschlag des Schnitts vor dem Code, nicht vollzogen.
- `rate-ohne-detektor-schwankt-mit-der-last` (2×, offen) und `befund-an-rolle-ohne-artefakt` (2×, offen): unter der Schwelle, kein Ausgang zu vergeben; ein weiterer Beleg wäre der dritte, und der Lese-Schritt der Closure legt den Eintrag dem User vor. Zur Rate: Risiko in §6, weil Fänger der Fenster nur mit Wahrscheinlichkeit rot werden können. Zum Befund: jede offene Entscheidung steht in §6 mit dem Ausgang *offen* und der Adresse *Architect*, einer Rolle mit Artefakt (dieser Plan).
- `bericht-der-rolle-liegt-nicht-im-repo` (3×, verkörpert in `AGENTS.md` §3.10): Risiko in §6; der Implementer trägt Tabelle und Raten in §3 ein.
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×, offen): Bezug über die Zeichen einer Client-`id` und ihre Fixtures (Escape-Folgen, `\u`); sie werden nach dem Schreiben gegengelesen. Unter der Schwelle.

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF.
