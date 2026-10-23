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

**Bezug:** `LH-FA-02`, `LH-FA-03`, `LH-QA-03` (nebenläufige Tests), [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0007](../../adr/0007-race-detektor-im-gate.md) (das Gate fährt `make test-race`)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Status von Sessions) · `LH-FA-02.c` · `LH-FA-03.a` · `LH-RB-04.a` (Löschen, Bereinigung danach) · `SPEC-059` · `SPEC-097`, `SPEC-099`, `SPEC-100`, `SPEC-072` · `SPEC-105` · `SPEC-109` (Abbildung der Port-Fehler, Codes `ODR-E7003`, `ODR-E7004`, `ODR-E7009`, `ODR-E7010`, `ODR-E7011`), `SPEC-111` · `SPEC-114` (Nr. 2 `Clock`, Nr. 9 Form von Client-Kennungen) · `SPEC-121` (Nr. 13, *Ungeprüft*) · `SPEC-122` (Nr. 8) · `ARC-002`, `ARC-003`, `ARC-012`.
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

**Entstanden aus:** dem Schnitt von `slice-session-wartung` vor dem ersten Code-Commit (Schätzung × 1,25 über 3000, `AGENTS.md` §3.14); `slice-session-wartung` trägt `MaintenanceUseCase` samt Port der Bereinigung und liegt vor diesem Slice, dessen Start es in `done/` verlangt (§4). Die Übergaben aus `slice-session-wartung` §1 zu `SessionUseCase` gehören diesem Slice, sie stehen hier mit ihrer Herkunft.

**Übernimmt (Zeit und Kennungen):** Die Session-Kennung kommt aus `IDGenerator.SessionID()`, Zeit aus `Clock.Now()` (`SPEC-114`); die Form einer vom Aufrufer gelieferten `id` (leer, überlang, Zeichen) legt dieser Slice mit dem Architect **vor dem Code** fest (`SPEC-114` Nr. 9; Sendung von `slice-clock-idgenerator`, über `slice-session-wartung`, dessen §1 sie vor dem Schnitt führte).

**Übernimmt (Drain und Session-Regel):** `SessionUseCase` beendet eine Session mit `RecordUseCase.Drain(ctx, id, SHUTDOWN_TIMEOUT)` über den Inbound Port (`SPEC-121` Nr. 13). `Drain` liest und ändert die Session in der Ablage nicht: unbekannt (`ODR-E7003`) und nicht `open` (`ODR-E7011`) prüft dieser Slice vorher, den Status setzt er danach. `Drain` liefert den Bericht `Abgebrochen` und `NichtAngehaengt` und kehrt erst zurück, wenn kein Exchange der Session mehr läuft und keiner mehr beginnt; die Session wird `complete` genau dann, wenn jeder angehängte Exchange `capture_status=complete` trägt und `NichtAngehaengt` `0` ist, sonst `incomplete` (`LH-FA-13.a`). `CloseAll` darf `Drain` für mehrere Sessions zugleich rufen (Sendung von `slice-record-service`).

**Übernimmt (Bereinigung nach dem Löschen):** `SessionUseCase` löscht eine Session über `SessionRepository` und ruft danach die Blob-Bereinigung von `MaintenanceUseCase` über den Inbound Port, den `slice-session-wartung` liefert (`ARC-002`; `LH-RB-04.a`: unmittelbar nach dem Commit, best effort). Der Name der Operation für die erste Session je Service ist `OpenInitialSession` (`ARC-012`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `MaintenanceUseCase` (Startwartung, Integritätsprüfung, Blob-Abgleich, Aufbewahrung nach Alter, Operation der Bereinigung) — `slice-session-wartung`, in `done/` vor dem Start dieses Slice; hier wird der Port nur gerufen.
- Die Verdrahtung beider Use Cases im Composition Root (Reihenfolge Wartung, `OpenInitialSession`, `CloseAll` beim Herunterfahren) — `slice-bootstrap-m2`, das die Sendung in §1 mit der Kennung dieses Slice führt.
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

- [ ] `SessionUseCase`: anlegen (ID serverseitig oder vom Aufrufer, `ODR-E7010` bei Doppelung), `OpenInitialSession` (erste Session je Service beim Start im Modus `record`), lesen, auflisten.
- [ ] Beenden mit Warten auf laufende Exchanges bis `SHUTDOWN_TIMEOUT` über `RecordUseCase.Drain` und Status `complete` oder `incomplete` (`ODR-E7003`, `ODR-E7011`), `CloseAll` für mehrere Sessions zugleich.
- [ ] Löschen (`ODR-E7004`, `ODR-E7009`) und danach die Blob-Bereinigung über den Inbound Port von `MaintenanceUseCase` rufen (Sendung von `slice-session-wartung`).
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
| `src/hexagon/ports/inbound/` | update | `SessionUseCase` (der Port `MaintenanceUseCase` liegt aus `slice-session-wartung` vor) |
| `src/hexagon/services/session/` | neu | Session-Zustandsübergänge, `OpenInitialSession`, `CloseAll` |
| `spec/spezifikation.md` | update | Vertrag von `SessionUseCase` mit Randformen; Form einer Client-`id` |
| `spec/architecture.md` | update, falls der Vertrag sie ändert | Zeile `SessionUseCase` in `ARC-002`, `ARC-012`, §2.4 (Berichtigung an jeder Stelle, `AGENTS.md` §3.9; ohne Slices und ADR-Bezüge, `AGENTS.md` §3.4) |
| Tests des Pakets | neu | gegen die Repositories der SQLite-Datei und gegen Fakes (`RecordUseCase` mit `Drain`, `MaintenanceUseCase`); nebenläufige Fänger nach `AGENTS.md` §3.10 |

Der Vertrag, die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen vor dem ersten Code-Commit (`AGENTS.md` §3.12), nicht vor dem Start. Reihenfolge nach dem Start: Architect (Vertrag, Tabelle, Raster), Planner (Marke aus der Tabelle, §4), Implementer.

- Die Abhängigkeiten laufen über Inbound Ports, nie über das Paket (`architecture.md` §2): `SessionUseCase` → `RecordUseCase.Drain` und `SessionUseCase` → `MaintenanceUseCase` (Blob-Bereinigung nach dem Löschen).

**Auftrag an den Architect vor dem ersten Code-Commit** (Suchliste des Planner, kein Vertrag; die Verträge schreibt der Architect):

1. Vertrag von `SessionUseCase` mit Signaturen, Berichten (Löschbericht) und Fehlern, als eigene `SPEC-<NNN>`.
2. Form einer vom Aufrufer gelieferten `id` (leer, überlang, Zeichen; beide Seiten jeder Grenze), `SPEC-114` Nr. 9.
3. Randformen als Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), jede Zelle mit Ort oder „gilt nicht“ mit Grund; darin mindestens: Beenden einer unbekannten und einer nicht offenen Session und die Reihenfolge der Verstöße (`ODR-E7003`, `ODR-E7011`); Löschen einer offenen Session, die Quelle einer Baseline ist (`ODR-E7009` gegen `ODR-E7004`); Fehler von `Drain` und des Statuswechsels; zwei gleichzeitige Beenden derselben Session; `CloseAll` mit mehreren Sessions, mit einem Fehler, ohne offene Session und gleichzeitig mit Anlegen; `OpenInitialSession` in einem Modus ohne `record` und für einen unbekannten Service; Fehler der Bereinigung nach dem Löschen (best effort, ändert das Ergebnis des Löschens nicht?); `ctx` nil und beendet je Operation.
4. Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15, mit Fall je Element einer Menge, beiden Seiten jeder Grenze und dem Fall gleicher verglichener Werte) und, wo eine Zusage von der Nebenläufigkeit abhängt, je Fenster einer Abfolge aus Prüfen und Handeln (Session lesen und `open` prüfen, `Drain`, Status setzen) eine Zeile mit Haltepunkt an einem Argument des Tests, Wirkungs-Fänger oder benanntem Negativ (`AGENTS.md` §3.10, `SPEC-122` Nr. 8); akzeptierte Negative mit Grund und Folge im Betrieb (`AGENTS.md` §3.15).
5. Nachzug an jeder Stelle, die dieselbe Aussage trägt (`AGENTS.md` §3.9): zum Beenden `LH-FA-13.a`, `LH-FA-02.c`, `SPEC-121` Nr. 13, `ARC-012` und `architecture.md` §2.4; zum Löschen `LH-RB-04.a`, die DELETE-Zeile der Admin-Tabelle und `ARC-002`.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-session-wartung` und `slice-record-service` und `slice-fehlermodell-konfiguration` und `slice-clock-idgenerator` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, der Tabelle *Zusage · Fänger* (§3) und dem Raster (§6), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12, §3.15, §3.16); danach setzt der Planner die Marke (*Größenmarke* unten, `AGENTS.md` §3.14). Die Reihenfolge ist Architect, Planner, Implementer: der erste Code-Commit folgt erst auf die gesetzte Marke.

**Größenmarke** (`AGENTS.md` §3.14): **noch nicht gesetzt.** Der Planner setzt sie nach dem Vertrag des Architect auf die Schätzung aus der Tabelle × 1,25, aufgerundet auf volle Hundert. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; `<Basis>` ist der Commit, der Vertrag, Tabelle und Raster vor dem Code abschließt (Dokumente zählen nicht mit). Grobe Erwartung ohne Tabelle: rund 2000 bis 2100 (§6, Risiko *Größe*), unter der Grenze 3000 des Schnitts vor dem Code.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach dem Vertrag setzt (Metrik oben).
- `in-progress` → `open` (blockiert): der Port der Bereinigung aus `slice-session-wartung` ist nicht benutzbar oder `Drain` liefert nicht, was `SPEC-121` Nr. 13 zusagt.

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

- `complete` setzt voraus, dass alle Exchanges der Session `complete` sind und keiner unangehängt blieb (`SPEC-121` Nr. 13, Sendung in §1); ein Exchange, der nach `Drain` abgeschlossen wird, gegen den Statuswechsel (nach der Rückkehr von `Drain` läuft und beginnt keiner mehr, `SPEC-121` Nr. 13) — **Ausgang:** offen; das Gate fährt `-race` (`make test-race`, `SPEC-122`): ein Datenrennen macht es rot, je Prozess gezählt (`-test.count=1`); die Fenster zwischen Prüfen und Handeln (Session lesen und `open` prüfen, `Drain`, Status setzen; zwei Beenden derselben Session) meldet der Detektor nicht (`SPEC-122` Nr. 8b): je Fenster eine Zeile im Raster mit Wirkungs-Fänger oder Negativ (`AGENTS.md` §3.10); die nebenläufigen Fänger ordnen mit Kanälen, andere Wirkungen laufen vor der Übergabe mit `-count=20`.
- `CloseAll` ruft `Drain` für mehrere Sessions zugleich; ein Fehler einer Session und die gemeinsame Frist `SHUTDOWN_TIMEOUT` (`SPEC-059`) sind im Vertrag nicht entschieden — **Ausgang:** offen; Architect legt die Regel fest, das Raster führt die Zeilen.
- Die Bereinigung nach dem Löschen ist best effort (`LH-RB-04.a`); ob ein Fehler der Bereinigung das Ergebnis des Löschens ändert, steht nicht fest — **Ausgang:** offen; Architect, Zeile im Raster.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit (Auftrag in §3); jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.
- Größe (`BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`, 11×, verkörpert in `AGENTS.md` §3.14): grobe Erwartung rund 25 Zeilen in der Tabelle, davon rund 5 nebenläufig, 5 × 85 + 20 × 40 = rund 1200, dazu 2 Port-Fakes (`RecordUseCase` mit `Drain`, `MaintenanceUseCase`) zu je 200 = 1600, × 1,25 = rund 2000 bis 2100, unter 3000 (§4) — **Ausgang:** offen; der Planner setzt die Marke nach der Tabelle des Architect.
- Rate ohne Detektor (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 1×, offen): ein nebenläufiger Fänger, der nur mit Wahrscheinlichkeit rot wird (Wirkung ohne Datenrennen), steht in Plan oder Kommentar als Einzelzahl, obwohl die Rate mit der Last des Hosts schwankt — **Ausgang:** offen; Plan und Kommentare nennen für solche Fänger die Spanne mehrerer Läufe und den Messenden, die Rate unter dem Detektor (je Prozess) steht als feste Größe.
- Bericht des Implementers (`BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo`, 2×, offen): Mutationstabelle und Raten des Implementers liegen nicht im Repo, der Reviewer misst neu — **Ausgang:** offen; liegt der Bericht wieder nicht vor, ist das der dritte Beleg und der Lese-Schritt der Closure braucht einen Ausgang (der User entscheidet).

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
- `berichtigung-erreicht-nicht-jede-stelle` (5×, verkörpert in `AGENTS.md` §3.9): die Liste der Stellen steht in §3 Auftrag 5.
- `akzeptiertes-negativ-ohne-folge` (5×, verkörpert in `AGENTS.md` §3.15): was der Vertrag ungeprüft lässt, steht mit Grund und Folge im Betrieb; §3 Auftrag 4.
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (19×, verkörpert in `AGENTS.md` §3.15): der Vertrag nennt nur, was ein Fänger der Tabelle prüft.
- `randform-erst-im-review-entschieden` (16×, verkörpert in `AGENTS.md` §3.16) und `universalaussage-im-vertrag-ungeprueft` (6×, verkörpert in `AGENTS.md` §3.16): das Raster in §6 vor dem ersten Code-Commit; Allaussagen („nach `Drain` läuft kein Exchange“) nennen ihren Fänger über alle Fälle.
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (11×, verkörpert in `AGENTS.md` §3.14): Marke in §4, Erwartung unter 3000.
- `rate-ohne-detektor-schwankt-mit-der-last` (1×, offen) und `bericht-der-rolle-liegt-nicht-im-repo` (2×, offen): Risiken in §6; unter der Schwelle, kein Ausgang zu vergeben; ein dritter Beleg zum Bericht legt der Lese-Schritt der Closure dem User vor.
- `befund-an-rolle-ohne-artefakt` (2×, offen): jede offene Entscheidung steht in §6 mit dem Ausgang *offen* und der Adresse *Architect*, einer Rolle mit Artefakt (dieser Plan). Unter der Schwelle.
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×, offen): ohne Bezug zum Gegenstand; Fixtures mit Escape-Folgen (Zeichen der `id`) werden nach dem Schreiben gegengelesen.

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF.
