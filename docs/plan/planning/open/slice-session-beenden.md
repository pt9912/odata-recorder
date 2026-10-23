# Slice slice-session-beenden: SessionUseCase, Teil Beenden: Drain, Statusregel, CloseAll

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-02`, `LH-FA-03`, `LH-FA-13`, `LH-QA-03` (nebenläufige Tests), [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0007](../../adr/0007-race-detektor-im-gate.md) (das Gate fährt `make test-race`), [ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md) (Status `Accepted` seit 2026-10-23: das Zurückstellen der Bereinigung im Betrieb beruht darauf, dass dieser Slice den Status einer Session erst nach `Drain` setzt)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Status von Sessions, Beenden) · `LH-FA-02.c` (Beenden des Prozesses) · `LH-FA-03.a` · `SPEC-005` (Zeile `close`) · `SPEC-059` · `SPEC-105` (Nr. 3, `SetStatus`; Nr. 7, 11, `ListBySession`) · `SPEC-109` (Nr. 4 bis 6; Codes `ODR-E7003`, `ODR-E7011`), `SPEC-111` · `SPEC-114` (Nr. 2 `Clock`) · `SPEC-121` (Nr. 2d, 9, 13, 14, *Ungeprüft*) · `SPEC-122` (Nr. 8) · `SPEC-123` (Nr. 5 und 8a: Zurückstellen setzt voraus, dass die Session `open` bleibt, bis `Drain` zurückkehrt) · `ARC-002`, `ARC-003`, `ARC-012`, `architecture.md` §2.2, §2.4, §4.9.
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

**Ziel:** `SessionUseCase` beendet Recording-Sessions: eine Session mit `RecordUseCase.Drain(ctx, id, SHUTDOWN_TIMEOUT)` und dem Status `complete` oder `incomplete`, alle offenen Sessions beim Herunterfahren mit `CloseAll`.

**Entstanden aus:** dem Schnitt von `slice-session-use-case` vor dem ersten Code-Commit (Schätzung × 1,25 rund 3300 über 3000, `AGENTS.md` §3.14). [slice-session-use-case](../done/slice-session-use-case.md) trägt Port `SessionUseCase`, Paket `session`, Konstruktor, Konfig, Fehlerabbildung, Protokoll und die Fakes und liegt vor diesem Slice, dessen Start es in `done/` verlangt (§4); dieser Slice ergänzt Beenden und `CloseAll` im selben Paket. Die Übergaben aus `slice-session-wartung` und `slice-record-service` zum Beenden gehören diesem Slice, sie stehen hier mit ihrer Herkunft.

**Übernimmt (Drain und Session-Regel):** `SessionUseCase` beendet eine Session mit `RecordUseCase.Drain(ctx, id, SHUTDOWN_TIMEOUT)` über den Inbound Port (`SPEC-121` Nr. 13; Sendung von `slice-record-service`, über `slice-session-wartung`). `Drain` liest und ändert die Session in der Ablage nicht: unbekannt (`ODR-E7003`) und nicht `open` (`ODR-E7011`) prüft dieser Slice vorher, den Status setzt er danach. `Drain` liefert den Bericht `Abgebrochen` und `NichtAngehaengt` und kehrt erst zurück, wenn kein Exchange der Session mehr läuft und keiner mehr beginnt; die Session wird `complete` genau dann, wenn jeder angehängte Exchange `capture_status=complete` trägt und `NichtAngehaengt` `0` ist, sonst `incomplete` (`LH-FA-13.a`). `CloseAll` darf `Drain` für mehrere Sessions zugleich rufen (Sendung von `slice-record-service`).

**Übernimmt (Annahme der Wartung):** Der Status wird erst nach der Rückkehr von `Drain` gesetzt, die Session bleibt bis dahin `open`: das Zurückstellen der Blob-Bereinigung im Betrieb (`SPEC-123` Nr. 8a, [ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md)) beruht darauf, dass kein Exchange einen Blob hält, solange keine Session `open` ist; ändert sich das, fällt seine Begründung (Sendung von `slice-session-wartung`, Fänger: der Status ist beim Aufruf von `Drain` noch `open` und wird erst danach gesetzt, und zwar auf jedem Rückkehrpfad von `Drain`, §3 Auftrag 3). Dieser Slice ist der Träger dieser Annahme (§6). `Drain` lebt im Zustand des Dienstes (`SPEC-121` Nr. 13: die Session nimmt „in diesem Dienst“ nichts mehr an): `SessionUseCase` erhält denselben `RecordUseCase` wie der HTTP-Adapter, den der Composition Root verdrahtet (Sendung an `slice-bootstrap-m2`, §6).

**Übernimmt (Zeit):** Die Zeit kommt aus `Clock.Now()` (`SPEC-114` Nr. 2), soweit das Beenden eine braucht.

**Übernimmt (Konfig):** `session.Konfig` und `session.New` aus [slice-session-use-case](../done/slice-session-use-case.md) (`SPEC-124` Nr. 1) tragen die Ports des Beenden **nicht**; dieser Slice ergänzt die Felder für `RecordUseCase`, `RecordingRepository` und, wenn der Vertrag es so bestimmt, die Frist (Namen und Quelle der Frist legt der Vertrag des Beenden fest), prüft sie in `New` (`model.ErrUngueltig` bei nil) und füllt sie im Helfer der Tests. Kein Feld der Konfig und keine Signatur aus `slice-session-use-case` wird umbenannt oder entfernt, `New` behält seine Signatur, und die Tests der Verwaltung laufen bis auf das Füllen der neuen Felder im Helfer unverändert (Sendung von `slice-session-use-case`, `AGENTS.md` §3.13).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Anlegen, `OpenInitialSession`, Lesen, Auflisten, Löschen mit Bereinigung, Form einer Client-`id`, Konstruktor, Konfig, Fehlerabbildung und Protokoll des Pakets — [slice-session-use-case](../done/slice-session-use-case.md), in `done/` vor dem Start dieses Slice; hier werden sie benutzt, nicht geändert. Ausgenommen davon ist die Konfig: sie trägt die Ports des Beenden nicht, dieser Slice ergänzt sie additiv (siehe *Übernimmt (Konfig)*).
- `MaintenanceUseCase` — `slice-session-wartung`, in `done/`; dieser Slice ruft `Startup` und `CleanupBlobs` nie.
- Die Verdrahtung im Composition Root (Reihenfolge Wartung, erste Session, `CloseAll` beim Herunterfahren; derselbe `RecordUseCase` für `SessionUseCase` und HTTP-Adapter) — `slice-bootstrap-m2`, das die Sendung in §1 und DoD mit der Kennung dieses Slice führt.
- Der Endpunkt `POST /__admin/sessions/{id}/close` — Folgewelle (Welle §6); der Use Case ist der Vertrag, den der Admin-Adapter später ruft.
- Das Lesen der Exchanges für die Ansicht (seitenweise, `InspectionUseCase`, `SPEC-105` *Ungeprüft*) — Folgewelle; dieser Slice liest die Exchanges allein, um die Statusregel von `SPEC-121` Nr. 13 anzuwenden.
- Das Anhalten einer laufenden Aufzeichnung selbst ist `Drain` des `RecordUseCase` (`slice-record-service`); dieser Slice ruft es über den Inbound Port.

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

- [ ] Beenden: prüfen (`ODR-E7003`, `ODR-E7011`), `RecordUseCase.Drain` mit `SHUTDOWN_TIMEOUT` bei noch offener Session, danach erst der Status `complete` oder `incomplete` nach der Regel von `SPEC-121` Nr. 13, auf jedem Rückkehrpfad von `Drain` (Fänger: beim Aufruf von `Drain` ist die Session `open`, `SPEC-123` Nr. 8a).
- [ ] `CloseAll` beendet jede offene Session, mehrere zugleich, mit der Regel für Fehler einer Session und für die Frist, die der Vertrag festlegt.
- [ ] Die Fenster der Abfolge aus Prüfen und Handeln (Session lesen und `open` prüfen, `Drain`, Exchanges lesen, Status setzen; zwei Beenden derselben Session) tragen je eine Zeile in der Tabelle *Zusage · Fänger* mit Haltepunkt, Wirkungs-Fänger oder benanntem Negativ (`AGENTS.md` §3.10). Eine weitere Zeile trägt die Sendung von [slice-session-use-case](../done/slice-session-use-case.md) (`AGENTS.md` §3.13): die Konfig wächst additiv (kein Feld wird umbenannt oder entfernt, `New` behält die Signatur), und die Tests der Verwaltung laufen bis auf das Füllen der neuen Felder im Helfer unverändert (Prüfung: die Testdateien von `slice-session-use-case` bleiben bis auf `helfer_test.go` im Diff dieses Slice unberührt).
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
| `src/hexagon/ports/inbound/session.go` | update | Operationen `Beenden` und `CloseAll` am Port `SessionUseCase` (der Port liegt aus `slice-session-use-case` vor) |
| `src/hexagon/services/session/` | update | Beenden, Statusregel, `CloseAll`; ruft `RecordUseCase` (`Drain`) und `RecordingRepository` (`ListBySession`) über die Ports, kein anderes Paket unter `services` |
| `spec/spezifikation.md` | update | Vertrag von Beenden und `CloseAll` samt Randformen, als Abschnitt der SPEC aus `slice-session-use-case` oder eigene SPEC (Architect); Nachzüge nach Auftrag 5 |
| `spec/architecture.md` | update, falls der Vertrag sie ändert | Zeile `SessionUseCase` in §2.2, `ARC-012`, §2.4, §4.9 (Berichtigung an jeder Stelle, `AGENTS.md` §3.9; ohne Slices und ADR-Bezüge, `AGENTS.md` §3.4) |
| Tests des Pakets | neu | gegen Fakes und, wo die Tabelle es festlegt, gegen die Repositories der SQLite-Datei; Fakes nach der Zählung des Planner (§4): `RecordUseCase` mit `Drain` (hält und scheitert auf Wunsch, liefert den Bericht) und `RecordingRepository` (`ListBySession`) neu, `SessionRepository` mit Ereignisprotokoll und Haltepunkten und Fang-Handler für `slog` aus `slice-session-use-case`; `FestUhr` liegt in `supporttest`; nebenläufige Fänger nach `AGENTS.md` §3.10 |

Der Vertrag, die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen vor dem ersten Code-Commit (`AGENTS.md` §3.12), nicht vor dem Start. Reihenfolge nach dem Start: Architect (Vertrag, Tabelle, Raster, Entscheidungen), Planner (Marke aus der Tabelle, §4), Implementer. Der Implementer trägt Tabelle *Zusage · Mutation · roter Test* und die Raten (Zahl der Prozesse, Messender) in einen Absatz von §3 ein (`AGENTS.md` §3.10).

- Die Abhängigkeiten laufen über Inbound Ports, nie über das Paket (`architecture.md` §2): `SessionUseCase` → `RecordUseCase.Drain`.
- Importe: wie `slice-session-use-case`; `.a-check.yml` bleibt.

**Auftrag an den Architect vor dem ersten Code-Commit** (Suchliste des Planner, kein Vertrag; die Verträge schreibt der Architect; der Vertrag der Verwaltung `SPEC-124`, Fehlerabbildung (Nr. 8) und Protokoll (Nr. 9) des Pakets liegen aus `slice-session-use-case` vor, die Konfig wächst hier additiv):

1. Vertrag von Beenden und `CloseAll`: Signaturen (Rückgabe, Fehler `*inbound.Fehler`, `ctx` nil und beendet), Quelle von `SHUTDOWN_TIMEOUT` (Feld der Konfig oder Argument von `CloseAll`), Fehler `ODR-E7003`, `ODR-E7011` und ihre Reihenfolge, Protokoll bei Fehlern in `CloseAll`.
2. Randformen als Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), jede Zelle mit Ort oder „gilt nicht“ mit Grund; darin mindestens: Beenden einer unbekannten und einer nicht offenen Session und die Reihenfolge der Verstöße (eine Kennung, die `Drain` nie sah, ist dort kein Fehler, `SPEC-121` Nr. 13); Beenden zum zweiten Mal nacheinander (`ODR-E7011`, nichts geändert, `SPEC-111`); Fehler von `Drain`, von `ListBySession` und des Statuswechsels; zwei gleichzeitige Beenden derselben Session (der zweite `SetStatus` ist `ErrUngueltigerUebergang`: welchen Code und welchen Status trägt der Verlierer?); `CloseAll` mit mehreren Sessions, mit einem Fehler (an der ersten, einer mittleren und der letzten), ohne offene Session und gleichzeitig mit Anlegen und mit Beenden einer einzelnen Session; `ctx` nil und beendet je Operation und je Stelle der Abfolge.
3. Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) **mit einer Zeile je Pfad, Stelle und Operation, wo der Vertrag „jeder“, „einmal“, „unverändert“, „erst nach“ sagt** (Lerneintrag von `slice-session-wartung`: dort hatte die Tabelle für solche Wörter eine Zeile für den Gutfall, und erst das Review fand W28 bis W30). Mindestens: (i) der Status wird erst nach der Rückkehr von `Drain` gesetzt, auf **jedem Rückkehrpfad** von `Drain` (gelungen, Fehler, `ctx` beendet), mit Haltepunkt im Fake von `Drain`: beim Aufruf ist die Session `open`; (ii) `CloseAll` beendet **jede** offene Session (ein Fall je Element, mindestens drei Sessions, der Fehler an erster, mittlerer und letzter Stelle) und ruft `Drain` für mehrere zugleich; (iii) **einmal**: `Drain` je Session, `Clock.Now()` falls bezogen; (iv) **unverändert**: Bericht und Fehler von `Drain`, Kontextfehler ohne Code, ein Fall je Operation und je gerufenem Port; (v) hält der Vertrag eine Sperre im Use Case, eine Zeile je Rückkehrpfad (wie W28 in `slice-session-wartung`). Je Fenster einer Abfolge aus Prüfen und Handeln eine Zeile mit Haltepunkt an einem Argument des Tests (`ctx`, Fake), Wirkungs-Fänger oder benanntem Negativ (`AGENTS.md` §3.10, `SPEC-122` Nr. 8b): Session lesen und `open` prüfen → `Drain`; `Drain` → Exchanges lesen → Status setzen; zwei Beenden derselben Session; `CloseAll` gegen Beenden; `CloseAll` gegen Anlegen; Beenden gegen Löschen. Jede Allaussage („nach `Drain` läuft kein Exchange“, „jede offene Session“) nennt ihren Fänger über alle Fälle; die Statusregel mit einem Fall je Element der Menge (`complete`, `incomplete`, `failed`; `NichtAngehaengt` `0` und `1`; keine Exchanges); akzeptierte Negative mit Grund und Folge im Betrieb (`AGENTS.md` §3.15), darunter ein Exchange, der nach dem Statuswechsel noch angenommen wird (`SPEC-121` *Ungeprüft*).
4. Entscheidungen, die kein Plan trifft und die nicht erst im Code fallen dürfen (`AGENTS.md` §3.12; sie gehören dem Architect): (a) Quelle der Statusregel — `RecordingRepository.ListBySession` liefert alle Exchanges einer Session auf einmal (Speicher bei großen Sessions; `SPEC-105` *Ungeprüft*), entweder Grenze mit Grund und Folge im Betrieb oder eine andere Quelle; (b) Zustand nach Fehler von `Drain` oder des Statuswechsels, auch bei beendetem `ctx` (der Bericht von `Drain` kommt dann mit `nil`, `SPEC-121` Nr. 13): die Session bleibt `open` in der Ablage, nimmt im Dienst aber keinen Exchange mehr an, und `CleanupBlobs` bleibt bis zum nächsten Start zurückgestellt — gilt das, oder setzt der Statuswechsel unter einem Kontext, den das Herunterfahren nicht beendet (Vorbild: das Anhängen in `SPEC-121` Nr. 9)? Dazu die Randform *Statuswechsel scheitert nach `Drain` am Kontext des Herunterfahrens*; (c) die Frist: `SHUTDOWN_TIMEOUT` als `frist` je `Drain` oder gemeinsam für `CloseAll`, und dass `Drain` nach Ablauf der Frist ohne Frist auf die abgebrochenen wartet (`SPEC-121` Nr. 13), das Herunterfahren also länger als `SHUTDOWN_TIMEOUT` dauern kann (`LH-FA-02.c`); (d) zwei gleichzeitige Beenden derselben Session: Sperre je Session im Use Case oder der Verlierer meldet den Fehler des Statuswechsels.
5. Nachzug an jeder Stelle, die dieselbe Aussage trägt (`AGENTS.md` §3.9): `LH-FA-13.a`, `LH-FA-02.c`, `SPEC-121` Nr. 13, die Zeile `close` in `SPEC-005`, `ARC-012`, `architecture.md` §2.2 (Zeile `SessionUseCase`), §2.4 und §4.9 (Sequenz: Lesen der Exchanges, `Drain`, danach der Status); `SPEC-123` Nr. 8a, falls der Vertrag die Annahme ändert.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-session-use-case](../done/slice-session-use-case.md) und [slice-session-wartung](../done/slice-session-wartung.md) und [slice-record-service](../done/slice-record-service.md) und [slice-fehlermodell-konfiguration](../done/slice-fehlermodell-konfiguration.md) und [slice-clock-idgenerator](../done/slice-clock-idgenerator.md) liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, der Tabelle *Zusage · Fänger* (§3) und dem Raster (§6), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12, §3.15, §3.16); danach setzt der Planner die Marke (*Größenmarke* unten, `AGENTS.md` §3.14). Die Reihenfolge ist Architect, Planner, Implementer: der erste Code-Commit folgt erst auf die gesetzte Marke.

**Größenmarke** (`AGENTS.md` §3.14): **noch nicht gesetzt.** Der Planner setzt sie nach dem Vertrag des Architect auf die Schätzung aus der Tabelle × 1,25, aufgerundet auf volle Hundert. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; `<Basis>` ist der Commit, der Vertrag, Tabelle und Raster vor dem Code abschließt (Dokumente zählen nicht mit).

**Erwartung des Planner beim Schnitt (vorläufig, vor der Tabelle des Architect; `AGENTS.md` §3.14):** Die Gesamtzählung vor dem Schnitt war 30 Zeilen, 9 nebenläufig, 5 Port-Fakes: 2605, × 1,25 = 3300 (Spanne 2700 bis 3700). Auf diesen Slice entfallen rund 15 Zeilen, davon 7 nebenläufig (Status erst nach `Drain`, zwei Beenden derselben Session, `CloseAll` mit mehreren Sessions zugleich, `CloseAll` mit einem Fehler und gemeinsame Frist, `CloseAll` gegen Anlegen, Beenden gegen Löschen, Zustand unter `-race`), 8 andere (Statusregel je Element, Prüfreihenfolge, Fehler je Port, unverändert, einmal), und 2 neue Port-Fakes (`RecordUseCase`, `RecordingRepository`; die Fakes von `SessionRepository` und `slog` liegen aus `slice-session-use-case` vor und bekommen Haltepunkte): 7 × 85 + 8 × 40 + 2 × 200 = 595 + 320 + 400 = **1315**, × 1,25 = 1644, aufgerundet **1700**. Die Marke bleibt offen; sie steht fest, wenn der Architect die Tabelle schreibt, und ersetzt diese Zahl.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach dem Vertrag setzt (Metrik oben).
- `in-progress` → `open` (blockiert): `Drain` liefert nicht, was `SPEC-121` Nr. 13 zusagt, oder der Port `SessionUseCase` aus `slice-session-use-case` trägt Beenden nicht ohne Änderung seiner Signaturen.

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

Der Vertrag, das Raster und die Tabelle entstehen vor dem ersten Code-Commit (§3 Aufträge 1 bis 3); eine Zelle ohne Ort geht an den Architect und wird nicht im Code entschieden (`AGENTS.md` §3.12).

- `complete` setzt voraus, dass alle Exchanges der Session `complete` sind und keiner unangehängt blieb (`SPEC-121` Nr. 13, Sendung in §1); ein Exchange, der nach `Drain` abgeschlossen wird, gegen den Statuswechsel — **Ausgang:** offen; das Gate fährt `-race` (`make test-race`, `SPEC-122`): ein Datenrennen macht es rot, je Prozess gezählt (`-test.count=1`); die Fenster zwischen Prüfen und Handeln (Session lesen und `open` prüfen, `Drain`, Exchanges lesen, Status setzen; zwei Beenden derselben Session) meldet der Detektor nicht (`SPEC-122` Nr. 8b): je Fenster eine Zeile im Raster mit Wirkungs-Fänger oder Negativ (`AGENTS.md` §3.10, §3 Auftrag 3); die nebenläufigen Fänger ordnen mit Kanälen, andere Wirkungen laufen vor der Übergabe mit `-count=20`.
- Eine offene Annahme von `SPEC-123` Nr. 8a: die Session bleibt `open`, bis `Drain` zurückkehrt, und ein Exchange läuft nur in einer offenen Session — **Ausgang:** offen; Träger ist dieser Slice (Sendung von `slice-session-wartung`, dort §6 *weiter offen*, über `slice-session-use-case`; die zweite Hälfte trägt `slice-record-service` in `done/`, `SPEC-121` Nr. 2d). Fänger: der Status ist beim Aufruf von `Drain` noch `open`, auf jedem Rückkehrpfad von `Drain` (§3 Auftrag 3 (i)). Ändert sich eine der beiden Aussagen, fällt die Begründung des Zurückstellens ([ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md)); bei der Closure ist der Ausgang *entfallen* (Fänger rot gesehen) oder *eingetreten*.
- `CloseAll` ruft `Drain` für mehrere Sessions zugleich; ein Fehler einer Session und die gemeinsame Frist `SHUTDOWN_TIMEOUT` (`SPEC-059`) sind im Vertrag nicht entschieden, und `Drain` wartet nach Ablauf der Frist ohne Frist auf die abgebrochenen Exchanges (`SPEC-121` Nr. 13): das Herunterfahren kann länger als `SHUTDOWN_TIMEOUT` dauern (`LH-FA-02.c`) — **Ausgang:** offen; Architect legt die Regel fest (§3 Auftrag 4 (c)), das Raster führt die Zeilen.
- Zustand nach einem Fehler von `Drain` oder des Statuswechsels: `Drain` nimmt die Session dauerhaft aus der Annahme des Dienstes (`SPEC-121` Nr. 13) und liefert seinen Bericht mit `nil` auch bei beendetem `ctx`; scheitert danach der Statuswechsel (etwa am Kontext des Herunterfahrens), bleibt die Session `open` in der Ablage und nimmt keinen Exchange mehr an, `CleanupBlobs` bleibt zurückgestellt, und erst der nächste Start setzt sie auf `incomplete` (`SPEC-123` Nr. 5), auch wenn jeder Exchange `complete` war — **Ausgang:** offen; Architect entscheidet (§3 Auftrag 4 (b)) und nennt Grund und Folge im Betrieb.
- Die Statusregel liest alle Exchanges einer Session auf einmal (`RecordingRepository.ListBySession`); große Sessions kosten Speicher (`SPEC-105` *Ungeprüft*) — **Ausgang:** offen; Architect entscheidet (§3 Auftrag 4 (a)): Grenze mit Grund und Folge im Betrieb oder eine andere Quelle.
- Zwei gleichzeitige Beenden derselben Session: beide bestehen die Prüfung auf `open`, beide rufen `Drain` (gleichzeitige `Drain` derselben Session sind erlaubt, `SPEC-121` Nr. 13), der zweite Statuswechsel ist `ErrUngueltigerUebergang` — **Ausgang:** offen; Architect entscheidet (§3 Auftrag 4 (d)).
- `Drain` wirkt im Zustand des Dienstes, der die Aufnahme annimmt (`SPEC-121` Nr. 13): ein `SessionUseCase`, den der Composition Root mit einem zweiten `RecordUseCase` baut, wartet auf keinen laufenden Exchange und stoppt die Annahme nicht, setzt den Status aber trotzdem; ein laufender Exchange wird dann nicht angehängt (`SPEC-105` Nr. 7) — **Ausgang:** weiter offen; Träger ist `slice-bootstrap-m2` (Sendung in §1 und DoD dort, Kennung dieses Slice).
- Größe (`BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`, 12×, verkörpert in `AGENTS.md` §3.14): Erwartung 15 Zeilen, 7 nebenläufig, 2 neue Fakes, 1315, × 1,25 = 1700 (§4) — **Ausgang:** offen; der Planner setzt die Marke nach der Tabelle des Architect.
- Rate ohne Detektor (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 2×, offen): ein nebenläufiger Fänger, der nur mit Wahrscheinlichkeit rot wird (Wirkung ohne Datenrennen), steht in Plan oder Kommentar als Einzelzahl, obwohl die Rate mit der Last des Hosts schwankt (`slice-session-wartung`: W27 ohne Detektor bei `--cpus=0.05` 83 von 88 Prozessen statt „16 von 16“) — **Ausgang:** offen; Plan und Kommentare nennen für solche Fänger die Spanne mehrerer Läufe und den Messenden, die Rate unter dem Detektor (je Prozess) steht als feste Größe; ein weiterer Beleg wäre der dritte, und der Lese-Schritt der Closure legt den Eintrag dem User vor. Dieser Slice trägt die meisten Fenster und damit die größte Wahrscheinlichkeit dafür.
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
- `nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, verkörpert in `AGENTS.md` §3.10, Sensor `make test-race`): das Gate fährt `-race` (`SPEC-122`); ein Datenrennen zählt je Prozess, jede andere Wirkung eine Wiederholung mit `-count=20`; je Fenster der Abfolge aus Prüfen und Handeln eine Zeile mit Haltepunkt, Wirkungs-Fänger oder Negativ (`SPEC-122` Nr. 8b); §3 Auftrag 3, §6.
- `berichtigung-erreicht-nicht-jede-stelle` (6×, verkörpert in `AGENTS.md` §3.9): die Liste der Stellen steht in §3 Auftrag 5; die Sendung an `slice-bootstrap-m2` steht dort in §1 und DoD mit der Kennung dieses Slice (`AGENTS.md` §3.13), geprüft gegen *Ausdrücklich NICHT* und DoD des Nehmers.
- `akzeptiertes-negativ-ohne-folge` (6×, verkörpert in `AGENTS.md` §3.15): was der Vertrag ungeprüft lässt, steht mit Grund und Folge im Betrieb; §3 Auftrag 3.
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (20×, verkörpert in `AGENTS.md` §3.15): der Vertrag nennt nur, was ein Fänger der Tabelle prüft; Wörter wie „jeder“, „einmal“, „unverändert“ bekommen eine Zeile je Pfad, Stelle und Operation (Lerneintrag von `slice-session-wartung`).
- `randform-erst-im-review-entschieden` (17×, verkörpert in `AGENTS.md` §3.16) und `universalaussage-im-vertrag-ungeprueft` (6×, verkörpert in `AGENTS.md` §3.16): das Raster in §6 vor dem ersten Code-Commit; Allaussagen („nach `Drain` läuft kein Exchange“) nennen ihren Fänger über alle Fälle.
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (12×, verkörpert in `AGENTS.md` §3.14): Marke in §4; Erwartung 1700, unter 3000.
- `rate-ohne-detektor-schwankt-mit-der-last` (2×, offen) und `befund-an-rolle-ohne-artefakt` (2×, offen): unter der Schwelle, kein Ausgang zu vergeben; ein weiterer Beleg wäre der dritte, und der Lese-Schritt der Closure legt den Eintrag dem User vor. Zur Rate: Risiko in §6. Zum Befund: jede offene Entscheidung steht in §6 mit dem Ausgang *offen* und der Adresse *Architect*, einer Rolle mit Artefakt (dieser Plan).
- `bericht-der-rolle-liegt-nicht-im-repo` (3×, verkörpert in `AGENTS.md` §3.10): Risiko in §6; der Implementer trägt Tabelle und Raten in §3 ein.
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×, offen): ohne Bezug zum Gegenstand.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
