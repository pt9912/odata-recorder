# Slice slice-telemetry-einbau-mock: Einbau von `Telemetry` in Mock-Dienst (Query-Span) und Workspace-Mutation (Mutations-Span)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-01`, `LH-FA-07`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-029`, `SPEC-030` (spezifikation.md §5) · `LH-QA-01.a` (Mockserver: Zerlegung des Gesamtspans durch `odata.mock.query`) · `LH-FA-07.b` (Vertragsmatrix der Mutationen) · `LH-FA-11.a` (Coverage-Status) · `SPEC-044` · `ARC-002`, `ARC-003`, `ARC-021` · `architecture.md` §4.4, §4.5.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `MockUseCase` setzt den Span `odata.mock.query` (`SPEC-029`, Attribute `recorder.entity_set`, `recorder.coverage_status`) um die Query-Ausführung (Parser, Coverage-Prüfung, Interpreter, Codec) und die Mutationen des Workspace den Span `odata.workspace.mutation` (`SPEC-030`, Attribute `recorder.workspace_id`, `recorder.mutation_type`); der Zähler je Fehlercode steigt für die Fehler des Mock-Pfads; der Composition Root verdrahtet den Adapter im Modus `mock`. Kein Query-Literal (`$filter`-Werte), kein Body und kein Key-Wert wird Attribut.

**Übernimmt (Sendung von `slice-telemetry-port`):** der Einbau in die Pfade `mock` und Workspace-Mutation.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Port und Adapter `Telemetry` — `slice-telemetry-port`.
- Gesamtspan, Capture-Span und Match-Span (`SPEC-026` bis `SPEC-028`) — `slice-telemetry-einbau-gateway`; der Gesamtspan liegt im HTTP-Adapter und ist modusunabhängig.
- Spans und Zähler des Hybrid-Fallbacks — kein Slice von `welle-betrieb` nimmt sie an (Sendung ohne Nehmer, siehe §6 von `slice-telemetry-einbau-gateway`, Adresse: Planner).
- Mutations-Spans der Admin-Endpunkte (Reset, Restore, Delete über HTTP) — `welle-admin-api`; welche Operationen des `WorkspaceUseCase` und `SnapshotUseCase` ein `recorder.mutation_type` tragen, entscheidet der Architect (Auftrag 1), die Endpunkte kommen später.
- Die Messung des Mockservers (`LH-QA-01.a`) — `slice-lasttest`; dieser Slice liefert den Span, der die Query-Ausführung ausweist.

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

- [ ] Query-Span: jede Abfrage des Mock-Dienstes (Collection, Key-Lookup, `/$count`, `$metadata` nach Entscheidung des Architect) öffnet und beendet `odata.mock.query` mit Entity-Set und Coverage-Status, auch im Fehlerfall (`ODR-E6003`, `ODR-E6001`, Abbruch); der Span umfasst die Query-Ausführung, nicht die Serialisierung außerhalb (`LH-QA-01.a`).
- [ ] Mutations-Span: jede Mutation des Workspace (nach `LH-FA-07.b`) setzt `odata.workspace.mutation` mit Workspace-ID und Art der Mutation; eine abgewiesene Mutation (Constraint, Coverage) trägt den Span mit ihrem Ergebnis; der Zähler je Fehlercode steigt für jeden Code des Mock-Pfads, den der Test auslöst.
- [ ] Verdrahtung im Composition Root im Modus `mock` und Test durch Mock-Lesen und Mutation: die Spans `SPEC-029` und `SPEC-030` mit ihren Pflicht-Attributen, ohne Query-Literal und Body (Canary).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Macht der Slice einen Fehlercode, Schlüssel oder neue Logfelder am Binary sichtbar: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb.
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
| `src/hexagon/services/mock/` | update | Query-Span und Mutations-Span (Konstruktor, Konfiguration) |
| `src/hexagon/services/workspace/` | update, falls der Architect Operationen des `WorkspaceUseCase` einbezieht (Auftrag 1) | Mutations-Span an der Stelle, an der der Zustand sich ändert |
| `src/bootstrap/` | update | Verdrahtung des Adapters im Modus `mock` |
| `spec/spezifikation.md` | update | Aufrufstellen, Werte von `recorder.mutation_type`, Randformen (Architect) |
| Tests | update, neu | Fakes der Ports; der Test durch Lesen und Mutation |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Aufträge an den Architect:**

1. **Mutation und ihre Art.** Welche Operationen `odata.workspace.mutation` tragen (POST, PATCH, PUT, DELETE nach `LH-FA-07.b`; Reset, Restore, Löschen des Workspace, Anlegen) und die Wertemenge von `recorder.mutation_type`, geschlossen und ohne Nutzdaten.
2. **Grenzen der Spans.** Wo `odata.mock.query` beginnt und endet (nach der Zuordnung von Service und Workspace? Einschluss der Coverage-Prüfung?), wo `odata.workspace.mutation` liegt (im Dienst oder im Repository, innerhalb der Transaktion oder darum), und die Eltern-Kind-Beziehung zum Gesamtspan (Weitergabe über `ctx`, nach Vertrag von `slice-telemetry-port`).
3. **Attribute.** Wertemenge von `recorder.coverage_status` (auch `unknown`, ohne Eintrag), `recorder.entity_set` bei unbekanntem Set (ein Wert aus dem Request ist kein Attribut ohne Prüfung gegen das Modell); `recorder.workspace_id` bei unbekanntem Workspace (`ODR-E7001`).
4. **Fehlerzähler.** Welche Codes des Mock-Pfads zählen und wo, so dass kein Fehler doppelt zählt (der Adapter des Gesamtspans zählt, was er schreibt, `slice-telemetry-einbau-gateway`).
5. **Aufrufstellen als Menge** (`AGENTS.md` §3.15): jede Operation des Mock-Dienstes, jede Mutationsart, ein Fall je Stelle.
6. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 6) | — | — |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Mutationsart, jeder Coverage-Status, jeder Fehlercode des Mock-Pfads), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-telemetry-port`, `slice-mock-lesen`, `slice-mock-schreiben`, `slice-snapshot-use-case` und `slice-bootstrap-m3` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-port`: der Port und sein Fake, die dieser Slice ruft.
- `slice-mock-lesen` und `slice-mock-schreiben` (`welle-mock`): die Orte der beiden Spans (`src/hexagon/services/mock`, heute nicht vorhanden); ohne Lesen und Mutation gibt es weder Query-Ausführung noch Zustandsänderung zu umschließen.
- `slice-snapshot-use-case` (`welle-mock`): Reset, Restore und Snapshot sind Zustandsänderungen des Workspace; ob sie `odata.workspace.mutation` tragen, entscheidet Auftrag 1, und der Dienst muss dann bestehen.
- `slice-bootstrap-m3` (`welle-mock`): der Modus `mock` im Composition Root, über den der Test die Spans erreicht (kein HTTP-Adapter für Admin, kein CLI).
- Handbuch-Punkt der DoD (nur wenn er greift): `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` aus `welle-benutzerhandbuch` liegen dann in `done/`; beide sind Bedingung von `welle-mock` und damit hier erfüllt.

**Abhängig von `welle-mock` / M3, weil** die Orte beider Spans in `welle-mock` entstehen; das ist der Grund des Start-Triggers der Welle ([welle-betrieb](../welle-betrieb.md) §2). **Nicht abhängig** von `slice-telemetry-einbau-gateway`, `slice-container-image` und `slice-lasttest`. Reihenfolge-Vorsicht, kein Blocker: beide Einbau-Slices ändern die Verdrahtung des Adapters in `src/bootstrap/`; läuft dieser Slice nach `slice-telemetry-einbau-gateway`, übernimmt er dessen Verdrahtung, läuft er davor, legt er sie an und der Gateway-Slice zieht sie nach (der Planner nennt die Reihenfolge beim Start des zweiten).

**Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` (`SPEC-128`), sobald die Tabelle *Zusage · Fänger* vorliegt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht ist dann Query-Span gegen Mutations-Span.
- `in-progress` → `open` (blockiert): die Grenze des Mutations-Spans (innerhalb der Transaktion des Repositorys) lässt sich nicht ohne Typ des Adapters im Dienst ausdrücken (siehe §6).

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

- Ein Span, der den Zustandswechsel des Workspace umschließt, kann den Abbruch (`ctx`) mitten in der Transaktion sehen; der Span muss trotzdem enden und darf das Ergebnis der Mutation nicht ändern — **Ausgang:** offen; Adresse: Architect (Auftrag 2), nebenläufiger Wirkungs-Fänger und `make test-race`.
- `recorder.entity_set` und `recorder.workspace_id` stammen aus dem Request; ein Wert aus dem Request darf nur als Attribut stehen, wenn er gegen das Modell geprüft ist — **Ausgang:** offen; Adresse: Architect (Auftrag 3), Canary mit einem Entity-Set-Namen und einer Workspace-ID mit Sonderzeichen.
- Der Span `odata.mock.query` zerlegt den Gesamtspan nur, wenn beide verbunden sind (`LH-QA-01.a`); wird die Verbindung über `ctx` nicht gehalten, weist der Lasttest den Anteil nicht aus — **Ausgang:** offen; Adresse: Architect (Auftrag 2), Fänger: der Span ist Kind des Gesamtspans im Test durch den Composition Root.
- Pfad ohne Lauf unter dem Detektor (Abbruch, Ausfall der Senke) — **Ausgang:** offen; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): Wirkungs-Fänger oder akzeptiertes Negativ mit Grund und Folge (`AGENTS.md` §3.10).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund):

| Operation | fehlender Wert (Workspace, Entity-Set) | leerer Wert | unbekannter oder fremder Wert | Abbruch (`ctx`) | Wiederholung | Coverage `partial`/`unknown` |
|---|---|---|---|---|---|---|
| Query-Span: Collection | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Query-Span: Key-Lookup | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Query-Span: `/$count` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Mutations-Span: POST, PATCH, PUT, DELETE | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Mutations-Span: Reset, Restore (falls einbezogen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Zähler je Fehlercode des Mock-Pfads | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Einbau trifft nebenläufige Pfade und Fakes, die `ctx` umhüllen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) sind offen; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) gilt für jede Wirkung, die der Detektor nicht meldet; offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
