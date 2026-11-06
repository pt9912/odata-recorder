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

**Berührte Spec-Stellen:** `LH-RB-04.a` (Aufbewahrung: „Löschungen werden im Audit-Log festgehalten“) · `LH-FA-13.a` („Änderungen sind auditierbar“, Audit-Eintrag ohne Tokenwert) · `SPEC-005` (Schlusszeile: Audit-Logging der Schreibendpunkte) · `SPEC-123` Nr. 6 (Aufbewahrung beim Start), Nr. 8 · `SPEC-124` Nr. 7 (Löschen einer Session) · spezifikation.md §2 (Schema: ein fehlgeschlagener Import „wird im Audit-Log festgehalten“) · `SPEC-127` (Import) · `ARC-020`, `ARC-021` · `architecture.md` §Support-Ports.
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

**Ziel:** Der Querschnitts-Port `AuditLog` (Paket `src/hexagon/ports/support`, `ARC-021`) und sein Driven Adapter (`src/adapters/driven/audit`, `ARC-020`) halten fest, was gelöscht wurde: `SessionUseCase.Delete`, die Löschung der Aufbewahrung in `MaintenanceUseCase.Startup` (Stufe Aufbewahrung, `SPEC-123` Nr. 6) und die Löschung eines Workspace durch `WorkspaceUseCase` schreiben je Vorgang einen Audit-Eintrag mit Art des Vorgangs, Gegenstand (Kennung, Service), Zeitpunkt aus `Clock` und Korrelations-ID, ohne Geheimnisse (`LH-RB-04.a`); ein fehlgeschlagener Import (`ImportUseCase`) hinterlässt ebenfalls einen Eintrag, damit seine `import_id` wiederverwendbar bleibt (Schema-Abschnitt der Spezifikation). Ein Fake im Paket `supporttest` hält jedes Argument fest und zählt jede Operation.

**Übernimmt (Sendung von `slice-session-use-case`, §1 *Ausdrücklich NICHT*, Stand `done/`):** der Audit-Eintrag der Löschung einer Session in `SessionUseCase.Delete`; die Löschung selbst liegt dort.

**Übernimmt (Sendung von `slice-session-wartung`, §1 *Ausdrücklich NICHT*, Stand `done/`):** der Audit-Eintrag der Löschung durch die Aufbewahrung in `MaintenanceUseCase.Startup`; die Löschung selbst liegt dort. Beide Pläne liegen in `done/` und sind nicht änderbar; die Übernahme steht hier.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Audit-Einträge der Admin-Schreibendpunkte (Anlegen, Beenden, Import, Reset, Restore über HTTP) und der Eintrag zu `ODR-E2101` (Token falsch) — `welle-admin-api` (`SPEC-005`, `LH-FA-13.a`); dieser Slice liefert Port und Adapter, jene Slices rufen ihn. Die Sendung hat noch keinen Nehmer mit Datei (siehe §6).
- Lesen, Abfragen und Exportieren des Audit-Logs — kein Endpunkt, kein Kommando im ersten Abnahmeumfang (`LH-FA-13` verlangt keine Auskunft über das Audit); der Eintrag liegt in der Ablage, die der Architect wählt.
- Das Diagnose-Log und die Zähler (`Telemetry`) — `slice-telemetry-port`; ein Audit-Eintrag ist ein Nachweis, kein Diagnose-Log.
- Die Löschung der Blobs nach dem Commit (best effort, `SPEC-123` Nr. 8) — ob sie einen eigenen Eintrag erhält, entscheidet der Architect (Auftrag 1); ohne Entscheidung ist sie nicht Gegenstand.

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
| `src/hexagon/ports/support/`, `supporttest/` | update | Port `AuditLog` und Fake |
| `src/adapters/driven/audit/` | neu | Adapter und Ablage |
| `src/hexagon/services/session/` | update | Eintrag in `Delete` (Konstruktor, Konfiguration) |
| `src/hexagon/services/maintenance/` | update | Eintrag je gelöschter Session der Aufbewahrung (der Bericht zählt heute nur: `SessionsGeloescht`) |
| `src/hexagon/services/workspace/` | update | Eintrag beim Löschen des Workspace |
| `src/hexagon/services/baselineimport/` | update | Eintrag bei fehlgeschlagenem Import |
| `src/bootstrap/` | update | Verdrahtung |
| `tools/schema/schema.yaml`, `src/adapters/driven/sqlite/migrations/` | update, falls die Ablage eine Tabelle ist | `make schema-check` und `make migration-check` ([ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)) halten Schema und Migration byte-gleich |
| `spec/spezifikation.md` | update | Vertrag mit Randformen: Port, Eintragsform, Ablage (Architect) |
| Tests | update, neu | Fakes der Ports; ein Fall je Löschvorgang |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Aufträge an den Architect:**

1. **Menge der Vorgänge.** Welche Vorgänge einen Eintrag tragen: Löschen einer Session, Löschung der Aufbewahrung (ein Eintrag je Session oder ein Sammeleintrag), Löschen eines Workspace (mit Snapshots), fehlgeschlagener Import, Bereinigung der Blobs; die Menge ist geschlossen, jede Art hat ein Pflichtfeld-Schema. Der Bericht der Wartung trägt heute nur Zahlen, nicht die Kennungen der gelöschten Sessions; ob der Bericht sie trägt oder die Wartung den Port ruft, ist eine Entscheidung mit Folge für `SPEC-123`.
2. **Ablage.** Datei im Datenverzeichnis (anhängend) oder Tabelle in SQLite; bei der Tabelle ist es eine Schemaänderung mit Migration (`schema.yaml`, [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md): keine ungeprüfte weitere Migration) und die Frage, ob der Eintrag in der Transaktion der Löschung steht; bei der Datei die Frage der Atomarität, der Rechte und der Rotation. Neustart-Festigkeit ist Zusage.
3. **Reihenfolge und Fehler.** Wird der Eintrag vor, mit oder nach der Löschung geschrieben; scheitert die Ablage: schlägt die Löschung fehl, wird sie ohne Eintrag ausgeführt oder gemeldet? Verhalten bei beendetem `ctx` nach dem Commit der Löschung (die Löschung ist geschehen, der Eintrag fehlt?) und bei der Aufbewahrung beim Start (vor jeder Annahme von Requests, `SPEC-123` Nr. 3).
4. **Inhalt.** Welche Felder (Art, Gegenstand, `service_id`, Zeitpunkt, `correlation_id`, Anzahl gelöschter Exchanges/Blobs), und der Ausschluss von Geheimnissen: ein Eintrag zu `ODR-E2101` trägt nie den Tokenwert (`LH-FA-13.a`); die Menge der Felder ist geschlossen, kein Freitext aus dem Request.
5. **Handle für Wartung und Start.** Die Wartung läuft vor den Listenern und ohne Request: woher `correlation_id` und Zeitpunkt kommen (`Clock`, `IDGenerator`).
6. **Aufrufstellen als Menge** (`AGENTS.md` §3.15): jede Art von Vorgang, jede Aufrufstelle des Ports, ein Fall je Stelle; der Fake hält jedes Argument fest.
7. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 7) | — | — |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Art von Vorgang, jedes Pflichtfeld, jeder Fall einer abgewiesenen Löschung), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
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
- Handbuch-Punkt der DoD (nur wenn er greift): `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` aus `welle-benutzerhandbuch` liegen dann in `done/`; bei einer Ablage als Datei oder Tabelle ohne Schlüssel und Code entfällt die Bedingung.

**Unabhängig von `welle-mock` / M3 und von M2 (`slice-bootstrap-m2`, `slice-http-adapter`, `slice-replay-service`), weil** alle Aufrufstellen Use Cases sind, die nicht über HTTP erreichbar sein müssen: die Tests rufen sie wie ihre Vorgänger über den Dienst. Der Composition-Root-Test (DoD Punkt 3) ruft die Löschung über den Handle des Roots, nicht über einen Listener. Der Mock löscht nichts, was nicht schon `WorkspaceUseCase.Delete` löscht. Reihenfolge-Vorsicht, kein Blocker: `slice-bootstrap-m2` ändert `src/bootstrap/` für die Verdrahtung der Dienste; läuft dieser Slice davor, legt er die Verdrahtung des Adapters `AuditLog` an und `slice-bootstrap-m2` übernimmt sie, läuft er danach, ergänzt er sie (der Planner nennt die Reihenfolge beim Start des zweiten). `slice-workspace-use-case` selbst ist der erste Slice von Linie A und liegt vor `slice-replay-service`; dieser Slice verlangt nicht, dass Linie A bis M2 fertig ist.

**Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` (`SPEC-128`), sobald die Tabelle *Zusage · Fänger* vorliegt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht ist dann Port und Adapter gegen die vier Aufrufstellen.
- `in-progress` → `open` (blockiert): die Ablage verlangt eine Schemaänderung, für die die Migration nach [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) nicht entschieden ist, oder der Bericht der Wartung trägt die Kennungen nicht (siehe §6).

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

- Löschen und Eintrag sind zwei Schritte; scheitert einer, fehlt entweder die Löschung oder ihr Nachweis (Fenster zwischen Commit der Löschung und Schreiben des Eintrags) — **Ausgang:** offen; Adresse: Architect (Auftrag 3), je Fenster eine Zeile im Raster mit Fänger oder benanntem Negativ (`AGENTS.md` §3.10).
- Der Bericht der Wartung nennt keine Kennungen der gelöschten Sessions; die Löschung der Aufbewahrung geht über `SessionRepository`, nicht über `SessionUseCase` (`SPEC-123`) — **Ausgang:** offen; Adresse: Architect (Auftrag 1), Folge für `SPEC-123`.
- Sendung ohne Nehmer mit Datei: die Audit-Einträge der Admin-Schreibendpunkte und des Zugriffs mit falschem Token (`SPEC-005`, `LH-FA-13.a`) gehören `welle-admin-api`, deren Slice-Dateien noch nicht bestehen — **Ausgang:** offen; Adresse: Planner beim Anlegen der Slices von `welle-admin-api` (`AGENTS.md` §3.13: erst der Nehmer führt den Punkt in §1 oder DoD).
- Die Tabelle von [welle-betrieb](../welle-betrieb.md) §4 nennt als Löschvorgänge Session, Aufbewahrung und Workspace; der fehlgeschlagene Import (Schema-Abschnitt der Spezifikation) steht dort nicht — **Ausgang:** offen; Adresse: Planner (Welle-Datei nachziehen, siehe Bericht des Anlegens).
- Ablage als SQLite-Tabelle kollidiert mit der Kette aus `make schema-check` und `make migration-check` (zweite Schemaversion, „Migrationskette fehlt“ in `harness/README.md`) — **Ausgang:** offen; Adresse: Architect (Auftrag 2).
- Pfad ohne Lauf unter dem Detektor (Schreibfehler der Ablage, Abbruch nach dem Commit) — **Ausgang:** offen; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): Wirkungs-Fänger oder akzeptiertes Negativ mit Grund und Folge.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund):

| Operation | fehlende Kennung | leere Kennung | unbekannte oder fremde Kennung (Löschung scheitert) | Abbruch (`ctx`) oder Schreibfehler der Ablage | Wiederholung (zweites Löschen) | Geheimnis im Gegenstand |
|---|---|---|---|---|---|---|
| Eintrag zu `SessionUseCase.Delete` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Eintrag zur Löschung der Aufbewahrung (Start) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Eintrag zum Löschen eines Workspace | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Eintrag zu einem fehlgeschlagenen Import | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Ablage nach Neustart lesen (Nachweis der Dauer) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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
