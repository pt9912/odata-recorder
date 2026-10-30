# Slice slice-admin-workspaces: Admin-Endpunkte für Workspaces, Reset, Snapshots und Restore

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-09`, `LH-FA-10`, `LH-FA-13`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-005` (Zeilen `workspaces`, `reset`, `snapshots`, `restore`) · `LH-FA-09.a`, `LH-FA-09.b`, `LH-FA-10.a` · `SPEC-108` (Vertrag der Workspace- und Snapshot-Repositories) · `SPEC-025`, `SPEC-070`, `SPEC-094`, `SPEC-095`, `SPEC-096` · `SPEC-109` · `LH-RB-04.a` (Bereinigung nach dem Löschen) · `ARC-007` · `architecture.md` §2.2, §4.5.
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

**Ziel:** Der Admin-Adapter bildet die Endpunkte für Workspaces (auflisten, aus einer Baseline anlegen, lesen, löschen, `reset`) und für Snapshots (auflisten, benannt erzeugen, `restore`) aus `SPEC-005` auf `WorkspaceUseCase` und `SnapshotUseCase` ab; jede Antwort trägt Status, Fehlerobjekt und Korrelations-ID, jeder Schreibendpunkt einen Audit-Eintrag ohne Geheimnisse.

**Übernimmt:** den Admin-Endpunkt der Workspaces aus `slice-workspace-use-case` und die Endpunkte `GET` und `POST …/snapshots` und `POST …/restore` aus `slice-snapshot-use-case`, deren §1 *Ausdrücklich NICHT* sie `welle-admin-api` zuweisen (`AGENTS.md` §3.13).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Logik von Anlage, Reset, Löschen, Snapshot und Restore — `WorkspaceUseCase` und `SnapshotUseCase`; der Adapter bildet ab und schreibt nie direkt in die Datenbank (Architektur §2).
- Mutation des Entity-State und Lesen von Entities über die Admin-API — `SPEC-005` führt beides nicht; Mutationen laufen im Modus `mock` über den OData-Listener.
- Das Löschen einzelner Snapshots — `SPEC-108` kennt es nicht (ein Snapshot geht mit seinem Workspace).
- Der Rahmen (Listener, Token, Fehlerobjekt) — `slice-admin-listener-token`; der Port `AuditLog` — `welle-betrieb` (`slice-audit-log`).

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

- [ ] Workspaces: `GET` (Kennung, Service, Revision, Baseline, Erstellzeit), `POST` aus `baseline_snapshot_id` (optionale `id`; `ODR-E7007`, `ODR-E7008`), `GET {id}` (`ODR-E7001`), `DELETE` (die Bereinigung nach dem Commit ändert die Antwort nicht), `reset` (`ODR-E7001`).
- [ ] Snapshots: `GET …/snapshots`, `POST …/snapshots` mit `name` (`ODR-E7006`), `POST …/restore` mit `snapshot_id` (`ODR-E7002`, `ODR-E7008`; Cursor unverändert).
- [ ] Audit: jeder Schreibendpunkt schreibt über `AuditLog` einen Eintrag ohne Geheimnisse; formal ungültige Argumente (Name, `id`, Body) werden nach dem Rahmen aus `slice-admin-listener-token` abgebildet.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/adapters/driving/admin/` | update | Routen für Workspaces, Reset, Snapshots und Restore |
| `spec/spezifikation.md` | update (Architect) | Vertrag der Abbildung: Statuscodes der Erfolge, JSON-Form, Bodies, Reihenfolge der Prüfungen; Nachzug `SPEC-005` (Zeilen) |
| Tests | neu | Adapter gegen Fakes von `WorkspaceUseCase`, `SnapshotUseCase` und `AuditLog`: jedes Argument festgehalten, jede Operation gezählt, auch die, die der Adapter nie rufen darf, etwa `Restore` bei `reset` und umgekehrt (`AGENTS.md` §3.15); ein Fall je Aufrufstelle für den `ctx` des Requests |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle ist beim Anlegen des Plans leer:

| Zusage | Fänger |
|---|---|
| *(Architect: eine Zeile je Zusage, Randform und Fall des Vertrags)* | *(ein Test, ein Fall der Gegenprobe oder eine benannte Ausgabe des Gates)* |

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Bodies und Pfade:** `POST /__admin/workspaces` (`baseline_snapshot_id`, `id` optional; Form der `id` aus dem Vertrag von `slice-workspace-use-case`), `POST …/snapshots` (`name`: Form, Länge, leer), `POST …/restore` (`snapshot_id`); fehlende, leere und unbekannte Felder, falsche Typen, doppelte Schlüssel, Größe, `Content-Type`.
2. **Antworten:** Statuscodes der Erfolge (`200`, `201`, `204`), JSON-Form von Workspace und Snapshot, Fehlerabbildung nach `SPEC-109`.
3. **Reihenfolge der Fehler:** unbekannter Workspace (`ODR-E7001`) vor unbekanntem Snapshot (`ODR-E7002`), ein Snapshot eines anderen Workspace, Baseline-Snapshot beim Restore (`ODR-E7008`) — soweit der Vertrag der Use Cases sie nicht schon festlegt.
4. **Abbruch:** `DELETE`, `reset` und `restore` mit abgebrochenem `ctx`; die Wirkung bei gleichzeitiger Mutation im Modus `mock` gehört den Use Cases, der Adapter reicht den Kontext durch.
5. **Audit:** Umfang (jeder Schreibendpunkt, auch fehlgeschlagene?), Felder ohne Geheimnisse, und wer schreibt was, damit das Löschen eines Workspace keinen doppelten Eintrag erzeugt (`slice-audit-log` lässt `WorkspaceUseCase` seine Löschungen auditieren).
6. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* der Admin-Endpunkt aus `slice-workspace-use-case` und die Endpunkte aus `slice-snapshot-use-case` (beide `open/`): angenommen, §1 *Übernimmt* trägt sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* dort trifft sie (sie nennen `welle-admin-api` als Folgewelle). *Von diesem Slice an:* keine.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-admin-listener-token`, `slice-audit-log` (`welle-betrieb`), `slice-workspace-use-case` (`welle-record-replay`) und `slice-snapshot-use-case` (`welle-mock`) liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** Rahmen und Port `AuditLog` wie in `slice-admin-sessions`. `WorkspaceUseCase` (Anlegen, Lesen, Löschen, Reset) liegt heute nicht in `src/hexagon/ports/inbound/` und kommt mit `slice-workspace-use-case`; `SnapshotUseCase` und `WorkspaceUseCase.Restore` legt erst `slice-snapshot-use-case` an (alle *Code-Abhängigkeit, Port*).

**Nicht gebraucht:** `MockUseCase`, `slice-mock-lesen`, `slice-bootstrap-m3`, `slice-mock-abnahme`, M2 und M3. **Aber mittelbar am Mock-Kern:** `slice-snapshot-use-case` startet erst nach `slice-mock-schreiben` (dessen §4), so dass dieser Slice nicht vor dem Mock-Kern liegen kann. Der Teil Workspaces ohne Snapshots hängt nur an `slice-workspace-use-case`; ein Schnitt in zwei Slices wäre eine Umplanung durch den Planner.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): die Verträge von `WorkspaceUseCase` und `SnapshotUseCase` lassen eine Randform des Adapters (Name, `id`, Reihenfolge der Fehler) ohne Ort (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): „(Vertrag, offen)“ heißt, der Architect trägt den Ort in der Spezifikation vor dem Übergang `next` → `in-progress` ein; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `GET` und `POST /__admin/workspaces` (`baseline_snapshot_id`, `id`) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| `GET` und `DELETE /__admin/workspaces/{id}` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| `POST …/reset` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| `GET` und `POST …/snapshots` (`name`) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| `POST …/restore` (`snapshot_id`) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Audit-Eintrag der Schreibendpunkte | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- Form und Grenzen von `id` und `name` (leer, Länge, Zeichen) stehen im Vertrag der Use Cases, die noch nicht geschrieben sind (`slice-workspace-use-case`, `slice-snapshot-use-case` in `open/`) — **Ausgang:** offen; Adresse: Architect (Auftrag 1).
- Wer schreibt den Audit-Eintrag einer Löschung: der Adapter oder `WorkspaceUseCase` (`slice-audit-log`)? Beide ergäben einen doppelten Eintrag — **Ausgang:** offen; Adresse: Architect (Auftrag 5).
- Statuscodes der Erfolge und JSON-Form von Workspace und Snapshot stehen nicht in `SPEC-005` — **Ausgang:** offen; Adresse: Architect (Auftrag 2).
- Die Wirkung von `reset` und `restore` bei gleichzeitiger Mutation im Modus `mock` gehört den Use Cases (`LH-FA-07.a`, Serialisierung je Workspace); der Adapter reicht den Kontext durch — **Ausgang:** entfallen: Zuständigkeit der Use Cases; der Adapter-Test prüft die Durchreichung des `ctx`.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×): die Fakes von `WorkspaceUseCase`, `SnapshotUseCase` und `AuditLog` tragen jede Schnittstelle des Originals und zählen jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Marke steht nicht, bevor die Tabelle *Zusage · Fänger* vorliegt — **Ausgang:** offen; Adresse: Planner beim Start (§4).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes der Use Cases, außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `AGENTS.md` §3.9; die Zeilen von `SPEC-005` sind Stellen), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
