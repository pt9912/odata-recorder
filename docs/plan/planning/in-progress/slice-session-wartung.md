# Slice slice-session-wartung: SessionUseCase und MaintenanceUseCase samt Aufbewahrung

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-02`, `LH-FA-03`, `LH-RB-04`, `LH-QA-03` (nebenläufige Tests), [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0007](../../adr/0007-race-detektor-im-gate.md) (das Gate fährt `make test-race`)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Status von Sessions) · `LH-FA-02.c` · `LH-FA-03.a` · `LH-RB-04.a` (Aufbewahrung, Bereinigung) · `SPEC-048` · `SPEC-058`, `SPEC-059` · `SPEC-097`, `SPEC-099`, `SPEC-100`, `SPEC-072` · `SPEC-105`, `SPEC-106` (Nr. 8 `Remove`, *Unentschieden*), `SPEC-108` (`StorageMaintenance`) · `SPEC-109` (Abbildung der Port-Fehler, Codes `ODR-E7003`, `ODR-E7004`, `ODR-E7009`, `ODR-E7010`, `ODR-E7011`, `ODR-E3001`), `SPEC-111` · `SPEC-114` (Nr. 2 `Clock`, Nr. 9 Form von Client-Kennungen) · `SPEC-121` (Nr. 10, 13, *Ungeprüft*) · `SPEC-122` (Nr. 8) · `ARC-002`, `ARC-003`, `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `SessionUseCase` legt Recording-Sessions an (auch die erste je Service beim Start), beendet sie (auch alle beim Herunterfahren), liest und löscht sie, und `MaintenanceUseCase` setzt beim Start offene Sessions auf `incomplete`, prüft die Datenbank, gleicht Blobs gegen die Referenzmenge ab und wendet die Aufbewahrung nach Alter an.

**Übernimmt:** Die Übergaben aus `welle-persistenz`: `MaintenanceUseCase` samt Blob-Abgleich und die Aufbewahrung nach Alter (beide standen als „noch ohne Slice“ in der Roadmap).

**Übernimmt (Zeit und Kennungen):** Die Session-Kennung kommt aus `IDGenerator.SessionID()`, Zeit und Aufbewahrungsgrenze aus `Clock.Now()` (`SPEC-114`); die Form einer vom Aufrufer gelieferten `id` (leer, überlang, Zeichen) legt dieser Slice mit dem Architect **vor dem Code** fest (`SPEC-114` Nr. 9; Sendung von `slice-clock-idgenerator`).

**Übernimmt (Drain und Session-Regel):** `SessionUseCase` beendet eine Session mit `RecordUseCase.Drain(ctx, id, SHUTDOWN_TIMEOUT)` über den Inbound Port (`SPEC-121` Nr. 13). `Drain` liest und ändert die Session in der Ablage nicht: unbekannt (`ODR-E7003`) und nicht `open` (`ODR-E7011`) prüft dieser Slice vorher, den Status setzt er danach. `Drain` liefert den Bericht `Abgebrochen` und `NichtAngehaengt` und kehrt erst zurück, wenn kein Exchange der Session mehr läuft und keiner mehr beginnt; die Session wird `complete` genau dann, wenn jeder angehängte Exchange `capture_status=complete` trägt und `NichtAngehaengt` `0` ist, sonst `incomplete` (`LH-FA-13.a`). `CloseAll` darf `Drain` für mehrere Sessions zugleich rufen (Sendung von `slice-record-service`).

**Übernimmt (Bereinigung und Aufruf-Richtung):** `SessionUseCase` löscht eine Session über `SessionRepository` und ruft danach die Blob-Bereinigung von `MaintenanceUseCase` über den Inbound Port (`ARC-002`: löschende Use Cases rufen sie nach dem Commit; `LH-RB-04.a`: unmittelbar nach dem Commit, best effort); `MaintenanceUseCase` ruft keinen Use Case und löscht Sessions der Aufbewahrung über `SessionRepository`, nicht über `SessionUseCase`. Der Name der Operation für die erste Session je Service ist `OpenInitialSession` (`ARC-012`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Aufruf der Blob-Bereinigung nach dem Löschen eines Workspace — `slice-workspace-use-case`: `WorkspaceUseCase` ist der zweite löschende Use Case (`ARC-002`); der Port `MaintenanceUseCase` entsteht hier, der Aufruf gehört zum Nehmer, der in §1 und DoD die Sendung mit der Kennung dieses Slice führt (`AGENTS.md` §3.13).
- Die Verdrahtung beider Use Cases im Composition Root (Reihenfolge Wartung, erste Session, Listener; `CloseAll` beim Herunterfahren) — `slice-bootstrap-m2`, das die Sendung in §1 mit der Kennung dieses Slice führt.
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

- [ ] `SessionUseCase`: anlegen (ID serverseitig oder vom Aufrufer, `ODR-E7010` bei Doppelung), erste Session je Service beim Start im Modus `record`, beenden mit Warten auf laufende Exchanges bis `SHUTDOWN_TIMEOUT` und Status `complete` oder `incomplete`, `CloseAll`, lesen, auflisten, löschen (`ODR-E7004`, `ODR-E7009`) und danach die Blob-Bereinigung über den Inbound Port von `MaintenanceUseCase` rufen; `OpenInitialSession` und `CloseAll` (`ARC-012`).
- [ ] `MaintenanceUseCase` beim Start vor der Annahme von Requests: offene Sessions auf `incomplete`, `foreign_key_check` und `integrity_check`, Abgleich der Blobs gegen die Referenzmenge mit erneuter Referenzprüfung vor dem Löschen, und die Operation der Bereinigung nach dem Löschen einer Session (im laufenden Betrieb, best effort), die `SessionUseCase` und später `WorkspaceUseCase` rufen.
- [ ] Aufbewahrung nach Alter (`RETENTION_MAX_AGE`): ältere Sessions löschen, Sessions als Quelle einer Baseline überspringen mit Warnung im Log, ohne den Start abzubrechen.
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
| `src/hexagon/ports/inbound/` | update | `SessionUseCase`, `MaintenanceUseCase` |
| `src/hexagon/services/session/` | neu | Session-Zustandsübergänge |
| `src/hexagon/services/maintenance/` | neu | Startwartung, Blob-Abgleich, Aufbewahrung; ruft keinen anderen Use Case |
| `spec/spezifikation.md` | update | Vertrag beider Use Cases mit Randformen; Bezugszeitpunkt des Alters |
| `spec/architecture.md` | update, falls der Vertrag sie ändert | Zeilen `SessionUseCase` und `MaintenanceUseCase` in `ARC-002`, `ARC-012`, §2.4 (Berichtigung an jeder Stelle, `AGENTS.md` §3.9; ohne Slices und ADR-Bezüge, `AGENTS.md` §3.4) |
| Tests der Pakete | neu | gegen die Repositories der SQLite-Datei und gegen Fakes; nebenläufige Fänger nach `AGENTS.md` §3.10 |

Der Vertrag, die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen vor dem ersten Code-Commit (`AGENTS.md` §3.12), nicht vor dem Start. Reihenfolge nach dem Start: Architect (Vertrag, Tabelle, Raster), Planner (Marke aus der Tabelle, §4), Implementer.

- Die Abhängigkeiten laufen über Inbound Ports, nie über das Paket (`architecture.md` §2): `SessionUseCase` → `RecordUseCase.Drain` und `SessionUseCase` → `MaintenanceUseCase` (Blob-Bereinigung nach dem Löschen); `MaintenanceUseCase` ruft selbst keinen Use Case (zyklenfreier Aufrufweg).

**Auftrag an den Architect vor dem ersten Code-Commit** (Suchliste des Planner, kein Vertrag; die Verträge schreibt der Architect):

1. Verträge beider Use Cases mit Signaturen, Berichten (Wartungsbericht, Löschbericht) und Fehlern; je Use Case eine eigene `SPEC-<NNN>`, damit ein Schnitt entlang der Use Cases möglich bleibt (§4).
2. Die Entscheidungen aus §6, die dort *Architect* nennen: Bezugszeitpunkt des Alters; Regel der Bereinigung im laufenden Betrieb gegen `Commit` und gegen einen Exchange zwischen `Commit` und `Append`; Folge einer beschädigten Datenbank oder eines beschädigten Blobs beim Start.
3. Form einer vom Aufrufer gelieferten `id` (leer, überlang, Zeichen; beide Seiten jeder Grenze), `SPEC-114` Nr. 9.
4. Randformen als Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), jede Zelle mit Ort oder „gilt nicht“ mit Grund; darin mindestens: Reihenfolge und Abbruchverhalten der Startwartung (welche Stufe vor welcher; bricht ein Fehler die folgenden ab); Beenden einer unbekannten, einer nicht offenen Session und die Reihenfolge der Verstöße (`ODR-E7003`, `ODR-E7011`; Löschen einer offenen Session, die Quelle einer Baseline ist: `ODR-E7009` gegen `ODR-E7004`); Fehler von `Drain` und des Statuswechsels; zwei gleichzeitige Beenden derselben Session; `CloseAll` mit mehreren Sessions, mit einem Fehler, ohne offene Session, gleichzeitig mit Anlegen; erste Session je Service in einem Modus ohne `record`; Alter genau an der Grenze, `RETENTION_MAX_AGE=0`; `ctx` nil und beendet je Operation.
5. Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15, mit Fall je Element einer Menge, beiden Seiten jeder Grenze und dem Fall gleicher verglichener Werte) und, wo eine Zusage von der Nebenläufigkeit abhängt, je Fenster einer Sperre oder einer Abfolge aus Prüfen und Handeln eine Zeile mit Haltepunkt an einem Argument des Tests, Wirkungs-Fänger oder benanntem Negativ (`AGENTS.md` §3.10, `SPEC-122` Nr. 8); akzeptierte Negative mit Grund und Folge im Betrieb (`AGENTS.md` §3.15).
6. Nachzug an jeder Stelle, die dieselbe Aussage trägt (`AGENTS.md` §3.9): zum Alter `LH-RB-04.a`, `SPEC-058`, die Zeile zu `DELETE /__admin/sessions/{id}` und `ARC-002`; zum Beenden `LH-FA-13.a`, `LH-FA-02.c`, `SPEC-121` Nr. 13, `ARC-012` und `architecture.md` §2.4.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-record-service` und `slice-fehlermodell-konfiguration` und `slice-clock-idgenerator` liegen in `done/` (beim Start geprüft). Die Spezifikation trägt den Vertrag samt Randformen, der Tabelle *Zusage · Fänger* (§3) und dem Raster (§6), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12, §3.15, §3.16); danach setzt der Planner die Marke (*Größenmarke* unten, `AGENTS.md` §3.14). Die Reihenfolge ist Architect, Planner, Implementer: der erste Code-Commit folgt erst auf die gesetzte Marke.

**Größenmarke** (`AGENTS.md` §3.14): **noch nicht gesetzt.** Die ältere Zahl 1400 aus der freien Schätzung von rund 1000 Zeilen gilt nicht mehr (sie entstand vor der Formel und vor dem Faktor 1,25); der Planner setzt die Marke nach dem Vertrag des Architect auf die Schätzung aus der Tabelle × 1,25, aufgerundet auf volle Hundert, und ersetzt 1400 dabei mit Grund. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; `<Basis>` ist der Commit, der Vertrag, Tabelle und Raster vor dem Code abschließt (Dokumente zählen nicht mit).

**Vorläufige Erwartung des Planner (grob, ohne Tabelle, Stand beim Start):** rund 40 Zeilen in der Tabelle, davon rund 9 nebenläufig (Beenden gegen Beenden, Beenden gegen laufenden Exchange, `CloseAll`, doppelte `id` gleichzeitig, Bereinigung gegen `Commit` und gegen einen Exchange im Flug): 9 × 85 + 31 × 40 = rund 2000, dazu 3 bis 5 Port-Fakes zu je 200 (`RecordUseCase` mit `Drain`, `MaintenanceUseCase`, Repositories mit Fehlern, Log) = 600 bis 1000, zusammen 2600 bis 3000, × 1,25 = **rund 3250 bis 3750**. Das liegt über 3000: nach `AGENTS.md` §3.14 wird dann vor dem ersten Code-Commit geschnitten, wenn die Tabelle es bestätigt. Vorschlag des Planner für den Schnitt (nicht vollzogen; die Tabelle des Architect entscheidet): entlang der Use Cases, `MaintenanceUseCase` (Startwartung, Integrität, Abgleich, Aufbewahrung, Operation der Bereinigung; rund 16 Zeilen, 3 nebenläufig, 2 Fakes: rund 1200 × 1,25 = 1500) zuerst, dann `SessionUseCase` (rund 25 Zeilen, 5 nebenläufig, 2 Fakes: rund 1600 × 1,25 = 2100). Gemeinsam ist nur der Inbound Port der Bereinigung, den der erste Slice liefert und der zweite ruft; die Aufrufrichtung (`SessionUseCase` → `MaintenanceUseCase`) lässt das Maintenance-Paket zuerst fertig werden, ohne Fake des anderen.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach dem Vertrag setzt (Metrik oben); ein Schnitt vor dem Code ist nur bei Schätzung × 1,25 über 3000 vorgesehen.
- `in-progress` → `open` (blockiert): der Bezugszeitpunkt der Aufbewahrung ist nicht entschieden (siehe §6).

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

- Bezugszeitpunkt des Alters: Beginn der Session, Ende oder letzter Exchange — `LH-RB-04.a` sagt „Sessions, die älter sind“ — **Ausgang:** offen; Architect entscheidet vor dem Code.
- Blob-Bereinigung gegen gleichzeitiges `Commit` desselben Inhalts (offene Randform aus `slice-blob-ablage`: `Remove` gegen `Commit` gleicher Bytes, `SPEC-106` *Unentschieden*) — **Ausgang:** offen; der Abgleich läuft beim Start vor der Annahme von Requests, nach dem Löschen einer Session aber im laufenden Betrieb; Architect legt die Regel fest.
- Blob-Bereinigung im laufenden Betrieb gegen einen Exchange zwischen `Commit` und `Append`: `SPEC-121` Nr. 10 committet jeden Blob vor dem Anhängen, die Referenzmenge (`SPEC-108`, `ReferencedBlobs`) kennt ihn bis zum `Append` nicht, die Bereinigung könnte ihn nach der Referenzprüfung entfernen und der angehängte Exchange verwiese auf eine fehlende Datei (`ODR-E3001`, `SPEC-048`: „keine referenzierte fehlende Datei nach erfolgreichem Commit“) — **Ausgang:** offen; Architect legt die Regel fest (beim Start läuft kein Exchange; im laufenden Betrieb nach dem Löschen einer Session schon). Der Fänger ist kein Datenrennen, sondern eine Abfolge aus Prüfen und Handeln (`SPEC-122` Nr. 8b): er braucht einen Wirkungs-Fänger (`AGENTS.md` §3.10).
- `complete` setzt voraus, dass alle Exchanges der Session `complete` sind und keiner unangehängt blieb (`SPEC-121` Nr. 13, Sendung in §1); ein Exchange, der nach `Drain` abgeschlossen wird, gegen den Statuswechsel (nach der Rückkehr von `Drain` läuft und beginnt keiner mehr, `SPEC-121` Nr. 13) — **Ausgang:** offen; das Gate fährt `-race` (`make test-race`, `SPEC-122` aus `slice-race-gate`): ein Datenrennen macht es rot, je Prozess gezählt (`-test.count=1`); die Fenster zwischen Prüfen und Handeln (Session lesen und `open` prüfen, `Drain`, Status setzen; Prüfung auf Baseline-Quelle und Löschen) meldet der Detektor nicht (`SPEC-122` Nr. 8b): je Fenster eine Zeile im Raster mit Wirkungs-Fänger oder Negativ (`AGENTS.md` §3.10); die nebenläufigen Fänger ordnen mit Kanälen, andere Wirkungen laufen vor der Übergabe mit `-count=20`.
- Startwartung findet Datenbank oder Blob beschädigt (`ODR-E3001`): Start scheitert oder läuft weiter — `SPEC-048` nennt die Prüfung, nicht die Folge — **Ausgang:** offen; Architect.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit (Auftrag in §3); jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.
- Größe (`BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`, 11×, verkörpert in `AGENTS.md` §3.14): die vorläufige Erwartung liegt mit Schätzung × 1,25 bei rund 3250 bis 3750 (§4), über der Grenze 3000 des Schnitts vor dem Code — **Ausgang:** offen; der Planner entscheidet nach der Tabelle des Architect, ob er schneidet (Vorschlag in §4) oder die Marke setzt.
- Rate ohne Detektor (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 1×, offen): ein nebenläufiger Fänger, der nur mit Wahrscheinlichkeit rot wird (Wirkung ohne Datenrennen), steht in Plan oder Kommentar als Einzelzahl, obwohl die Rate mit der Last des Hosts schwankt — **Ausgang:** offen; Plan und Kommentare nennen für solche Fänger die Spanne mehrerer Läufe und den Messenden, die Rate unter dem Detektor (je Prozess) steht als feste Größe; ein weiterer Beleg macht den Eintrag zu 2×, kein Ausgang ist vor 3× zu vergeben.
- Bericht des Implementers (`BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo`, 2×, offen): Mutationstabelle und Raten des Implementers liegen nicht im Repo, der Reviewer misst neu und der Plan nennt keine Herkunft, die auflöst — **Ausgang:** offen; liegt der Bericht wieder nicht vor, ist das der dritte Beleg und der Lese-Schritt der Closure braucht einen Ausgang (der User entscheidet); bis dahin nennt der Plan die Zahlen des Implementers mit dem Vermerk, woher sie stammen.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Register unter `../observations/BEO-REPO/` trägt elf Einträge (Sub-Area `*`); dieser Plan antwortet auf jeden (Zahlen aus den Dateien unter `evidence/`, Ausgänge aus den `state.md`). Der Slice ist nebenläufig (Statuswechsel gegen laufende Exchanges, Bereinigung gegen `Commit` und gegen einen Exchange im Flug) und liefert zwei neue Verträge:
- `nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, verkörpert in `AGENTS.md` §3.10, Sensor `make test-race`): das Gate fährt `-race` (`SPEC-122`); ein Datenrennen zählt je Prozess, jede andere Wirkung eine Wiederholung mit `-count=20`; je Fenster einer Sperre oder einer Abfolge aus Prüfen und Handeln eine Zeile mit Haltepunkt, Wirkungs-Fänger oder Negativ, weil der Detektor ein Rennen ohne ungeordneten Zugriff nicht meldet (`SPEC-122` Nr. 8b); §3 Auftrag 5, §6.
- `berichtigung-erreicht-nicht-jede-stelle` (5×, verkörpert in `AGENTS.md` §3.9, geschärft seit welle-record-replay): der Vertrag berichtigt Aussagen an mehreren Stellen (Alter, Beenden, Bereinigung); die Liste der Stellen steht in §3 Auftrag 6; Kopf, §1, §3, §4, §6 und §8 dieses Plans nennen dieselbe Regel zur Nebenläufigkeit (`AGENTS.md` §3.10) und dieselbe Reihenfolge Architect, Planner, Implementer.
- `akzeptiertes-negativ-ohne-folge` (5×, verkörpert in `AGENTS.md` §3.15, geschärft seit welle-record-replay): was der Vertrag ungeprüft lässt (ein Rennen, das kein Lauf auslöst; `SPEC-106` *Unentschieden*) steht mit Grund und Folge im Betrieb, als Grund gilt Gleichwertigkeit oder die Folge, nicht die Lücke der Fänger; §3 Auftrag 5.
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (19×, verkörpert in `AGENTS.md` §3.15): beide Verträge nennen nur, was ein Fänger der Tabelle prüft.
- `randform-erst-im-review-entschieden` (16×, verkörpert in `AGENTS.md` §3.16) und `universalaussage-im-vertrag-ungeprueft` (6×, verkörpert in `AGENTS.md` §3.16): das Raster in §6 vor dem ersten Code-Commit; Allaussagen („kein Exchange der Session läuft nach `Drain`“, „Reihenfolge der Verstöße“) nennen ihren Fänger über alle Fälle.
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (11×, verkörpert in `AGENTS.md` §3.14, geschärft seit welle-record-replay): Marke und Schnitt in §4; die vorläufige Erwartung liegt über der Grenze 3000.
- `rate-ohne-detektor-schwankt-mit-der-last` (1×, offen): Risiko in §6, weil der Slice Fänger mit Wahrscheinlichkeit haben kann; unter der Schwelle, kein Ausgang zu vergeben.
- `bericht-der-rolle-liegt-nicht-im-repo` (2×, offen): Risiko in §6; ein dritter Beleg in diesem Slice erreicht die Schwelle ohne Ausgang, der Lese-Schritt der Closure legt ihn dem User vor.
- `befund-an-rolle-ohne-artefakt` (2×, offen): Bezug über die Aufträge an den Architect: jede offene Entscheidung steht in §6 mit dem Ausgang *offen* und der Adresse *Architect*, einer Rolle mit Artefakt (dieser Plan); ein Befund des Reviews an den Validator oder Owner wird bei der Closure als Risiko eingetragen. Unter der Schwelle.
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×, offen): ohne Bezug zum Gegenstand; Fixtures mit Escape-Folgen werden nach dem Schreiben gegengelesen, falls der Vertrag sie braucht (Zeichen der `id`).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
