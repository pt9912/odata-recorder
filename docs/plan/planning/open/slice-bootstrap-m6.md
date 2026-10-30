# Slice slice-bootstrap-m6: Composition Root für den Admin-Listener und der M6-Test

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-FA-18`, `LH-FA-09`, `LH-FA-10`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-13.a` · `SPEC-110` Nr. 2, 7, 10, 11 (Schlüssel dieser Version, `ADMIN_LISTEN`, Zusammenhang, Dateien) · `SPEC-008`, `SPEC-056`, `SPEC-077`, `SPEC-089`, `SPEC-090` · `SPEC-005` · `ARC-007`, `ARC-012` · `architecture.md` §2.5, §4.2, §4.5, §4.9.
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

**Ziel:** Der Composition Root liest `ADMIN_LISTEN`, `ADMIN_TOKEN_FILE`, `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT`, verdrahtet den Admin-Adapter mit den Use Cases, `AuditLog` und `Telemetry` und startet den Admin-Listener in jedem Modus neben dem OData-Listener (die Token-Datei vor dem Öffnen eines Listeners, geordnetes Herunterfahren beider); der M6-Test aus §3 der Welle läuft grün in `make test`: Rundlauf ausschließlich über Admin-Listener und OData-Listener, ein Audit-Eintrag je Schreibendpunkt, `GET /healthz` ohne Token.

**Übernimmt:** das Lesen und Prüfen der `ADMIN_*`-Schlüssel samt Aufheben ihrer Ablehnung in `SPEC-110` Nr. 2 (Sendung von `slice-admin-listener-token` und `slice-admin-tls`, deren §1 *Ausdrücklich NICHT* diesen Slice nennen) und den Admin-Listener, den `slice-bootstrap-m2` der Folgewelle zuweist (`AGENTS.md` §3.13).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `HYBRID_*`, `FALLBACK_*` und der Modus `hybrid` — `slice-bootstrap-m4`; die Schlüssel bleiben `ODR-E2001`, solange dieser Slice zuerst liegt (`SPEC-110` Nr. 2).
- `LISTEN_TLS_*` (TLS des OData-Listeners) — `slice-listener-tls` (`welle-betrieb`).
- Neue Endpunkte oder Fachlogik — der Slice verdrahtet, was die Slices der Welle liefern.
- Das Umschalten des Modus zur Laufzeit — `LH-FA-13` schließt es aus; der Rundlauf wechselt den Modus durch Neustart (§6).

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

- [ ] Verdrahtung und Konfiguration: der Admin-Listener läuft in jedem Modus neben dem OData-Listener; die Werte der `ADMIN_*`-Schlüssel kommen aus dem Wert des Starts bis in Adapter und TLS (Fänger, dass der Root sie weiterreicht und den Nullwert nicht stehen lässt; Muster aus `slice-upstream-tls`); Reihenfolge Startwartung, Token-Datei, Listener, ein Fehler der Token-Datei öffnet keinen Listener; geordnetes Herunterfahren beider Listener.
- [ ] M6-Test aus §3 der Welle: Rundlauf (Session anlegen, über den OData-Listener aufzeichnen, beenden, importieren, Workspace anlegen, im Modus `mock` ändern, Snapshot erzeugen, zurücksetzen, wiederherstellen) ausschließlich über Admin- und OData-Listener; Anfragen ohne oder mit falschem Token ergeben `ODR-E2101` ohne Zustandsänderung; ein Audit-Eintrag je Schreibendpunkt; `GET /healthz` ohne Token.
- [ ] Nicht gelieferte Schlüssel (`HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`) bleiben `ODR-E2001`, soweit ihre Slices nicht in `done/` liegen, und die Spezifikation führt den Nachzug von `SPEC-110` Nr. 2.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand: Admin-Listener, Token und Token-Datei, Schlüssel `ADMIN_*`, Endpunkte, Fehlercodes, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Voraussetzung ist `slice-handbuch-meldungskatalog` in `done/` (Trigger von `welle-mock`).
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
| `src/bootstrap/` | update | `ADMIN_*` lesen, prüfen und verdrahten; zweiter Listener neben dem OData-Listener; Reihenfolge von Start und Herunterfahren |
| `tests/` | neu | M6-Test mit Referenzdienst, Recorder und Client; Rundlauf über Admin- und OData-Listener |
| `spec/spezifikation.md` | update (Architect) | Nachzug von `SPEC-110` Nr. 2, 7, 10 und 11 für die `ADMIN_*`-Schlüssel |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Nachzug in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle ist beim Anlegen des Plans leer:

| Zusage | Fänger |
|---|---|
| *(Architect: eine Zeile je Zusage, Randform und Fall des Vertrags)* | *(ein Test, ein Fall der Gegenprobe oder eine benannte Ausgabe des Gates)* |

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18). Das gilt auch für die Beispiele des Handbuch-Kapitels.
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Nachzug von `SPEC-110`:** Nr. 2 (die `ADMIN_*`-Schlüssel wandern aus der Ablehnung in die geprüften), Nr. 7 (Syntax von `ADMIN_LISTEN`; der Port-Bereich 1 bis 65535 schließt Port 0 aus, ein Test braucht einen freien Port: wie bindet er ihn), Nr. 10 (Zusammenhang aus `slice-admin-tls`), Nr. 11 (die Token-Datei im Ablauf der Dateien: wann gelesen oder erzeugt, relativ zu `DATA_DIR`, Reihenfolge zur Anlage von `DATA_DIR`) und der Default `ADMIN_TOKEN_FILE` (`SPEC-056`).
2. **Aufbau des Rundlaufs:** der Modus ist Startkonfiguration, der Rundlauf zeichnet aber auf (`record`) und ändert dann im Modus `mock`. Entscheiden: zwei Starts im Testprozess mit demselben `DATA_DIR`, was zwischen ihnen erhalten bleibt (Datenbank, Token-Datei), und dass der zweite Start das Token nicht neu erzeugt.
3. **Reihenfolge von Start und Herunterfahren beider Listener:** `CloseAll` vor dem Schließen, welcher Listener zuerst schließt, Frist (`SHUTDOWN_TIMEOUT`), Verhalten bei einem Fehler des zweiten Listeners nach dem Öffnen des ersten.
4. **Verdrahtung der Ports:** `AuditLog`, `Telemetry` und die Zusammenfassung der Konfiguration für die Diagnose (Allowlist aus `slice-inspection-use-case`) aus dem Wert des Starts; der Dienst sieht nie die Umgebung.
5. **M6-Test:** Quelle der Audit-Einträge (Port oder Ausgabe), Ordnung unter dem Detektor (Kanäle statt Schlaf, Muster aus `slice-bootstrap-m2`), Zusammenspiel mit dem OData-Listener derselben Instanz.
6. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* das Lesen der `ADMIN_*`-Schlüssel aus `slice-admin-listener-token` und `slice-admin-tls` und der Admin-Listener aus `slice-bootstrap-m2`: angenommen, §1 *Übernimmt* trägt sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* dort trifft sie (`slice-bootstrap-m2` §1 nennt den Admin-Listener als Folgewelle). *Von diesem Slice an:* keine.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-admin-listener-token`, `slice-admin-tls`, `slice-inspection-use-case`, `slice-admin-sessions`, `slice-admin-imports`, `slice-admin-workspaces`, `slice-test-security`, `slice-bootstrap-m2`, `slice-bootstrap-m3`, `slice-audit-log` und `slice-telemetry-port` liegen in `done/`, dazu `slice-handbuch-meldungskatalog` für das Handbuch-Kapitel. Die Spezifikation trägt den Nachzug von `SPEC-110`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** der Root verdrahtet alle Adapter-Slices der Welle (*Code-Abhängigkeit*), die Ports `AuditLog` und `Telemetry` (*Code-Abhängigkeit*), den Record-Pfad (`slice-bootstrap-m2`) und den Mock-Pfad (`slice-bootstrap-m3`); der Rundlauf braucht beide Modi. `slice-test-security` liegt zuletzt vor diesem Slice (`welle-admin-api` §5; *Reihenfolge*, keine Typ-Abhängigkeit).

**Nicht gebraucht:** `welle-hybrid` und `slice-bootstrap-m4` (keine Kante). Beide ändern `src/bootstrap/` und `SPEC-110` Nr. 2 und laufen nacheinander; welcher zuerst, ist Priorisierung, und der zweite zieht den Stand des ersten nach (`AGENTS.md` §3.9).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): der Rundlauf lässt sich nicht ohne Umschalten des Modus zur Laufzeit bilden (Auftrag 2), oder ein Slice der Welle liefert nicht den Handle, den der Test braucht (siehe §6).

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
| Lesen von `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | (Vertrag, offen) |
| Lesen von `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | (Vertrag, offen) |
| Start der beiden Listener (Reihenfolge, Token-Datei zuerst) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Herunterfahren der beiden Listener | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Rundlauf des M6-Tests (zwei Starts, ein `DATA_DIR`) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- Der Rundlauf zeichnet auf (`record`) und ändert im Modus `mock`; der Modus ist Startkonfiguration (`LH-FA-13`) — **Ausgang:** offen; Adresse: Architect (Auftrag 2): zwei Starts mit demselben `DATA_DIR`, ohne Laufzeit-Umschalten.
- `SPEC-110` Nr. 7 schließt Port 0 aus; ein Test auf demselben Host braucht einen freien Port für `ADMIN_LISTEN` und `LISTEN` (Auftrag 1) — **Ausgang:** offen; Adresse: Architect.
- Die Reihenfolge Startwartung, Token-Datei, Listener und das Herunterfahren beider Listener (Auftrag 3): ein Fehler der Token-Datei öffnet keinen Listener — **Ausgang:** offen; Adresse: Architect.
- Der M6-Test läuft mit Referenzdienst, Recorder, Admin-Client und OData-Client im selben Prozess unter dem Detektor (`make test-race`) — **Ausgang:** offen; Muster und Ordnung der Fänger wie in `slice-bootstrap-m2` (Kanäle statt Schlaf).
- `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×): der Admin-Pfad im Composition Root erweitert die Pakete, deren Laufzeit Fälle der Gegenproben mitfahren — **Ausgang:** weiter offen → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` im Register.
- Gleichzeitige Änderung von `src/bootstrap/` und `SPEC-110` Nr. 2 durch `slice-bootstrap-m4` — **Ausgang:** offen; Adresse: Planner (Reihenfolge beim Start, keine parallelen Arbeitsbäume auf diesen Dateien).
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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, Prozess, Referenzdienst und zwei Clients unter Last) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `AGENTS.md` §3.9; der Nachzug von `SPEC-110`, das Handbuch und die Welle-Datei sind Stellen), `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16) und `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (4×, `AGENTS.md` §3.18, für die Beispiele des Handbuchs). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
