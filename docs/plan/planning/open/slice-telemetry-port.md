# Slice slice-telemetry-port: Port `Telemetry` und Adapter für strukturierte Logs, Metriken, Spans und Zähler je Fehlercode

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-05`, `LH-QA-04`, `LH-FA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-026` bis `SPEC-030` (spezifikation.md §5) · die Logfelder am Ende von spezifikation.md §4 · `SPEC-121` Nr. 12 (Protokoll des Record-Dienstes, heute über `*slog.Logger`) · `SPEC-005` (Diagnose: Zähler je Fehlercode) · `SPEC-109`, `SPEC-113` (Fehlerabbildung) · `ARC-020`, `ARC-021` · `architecture.md` §Support-Ports.
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

**Ziel:** Der Querschnitts-Port `Telemetry` (Paket `src/hexagon/ports/support`, `ARC-021`) und sein Driven Adapter (`src/adapters/driven/telemetry`, `ARC-020`) liefern, was die Spezifikation an Beobachtbarkeit zusagt: Log-Einträge mit den Feldern des Katalogs (`timestamp`, `level`, `mode`, `service_id`, `session_id`, `exchange_id`, `workspace_id`, `correlation_id`, `error_code`, `duration_ms`), Spans mit den Namen und Pflicht-Attributen von `SPEC-026` bis `SPEC-030`, Metriken und einen Zähler je Fehlercode; kein Credential, kein Body und kein Query-Literal erreicht Log oder Attribut. Ein Fake im Paket `supporttest` hält jedes Argument fest und zählt jede Operation. Der Slice liefert Port und Adapter, nicht den Einbau: kein Dienst und kein Adapter des Produkts ruft den Port, bevor die Einbau-Slices laufen.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Einbau in den HTTP-Adapter (Gesamtspan), den Record-Dienst (Capture-Span, Ersatz des `*slog.Logger`) und den Replay-Dienst (Match-Span) — `slice-telemetry-einbau-gateway`; der Port ist vom Einbau trennbar, und der Einbau braucht Code, den dieser Slice nicht kennt (HTTP-Adapter, Replay-Dienst).
- Einbau in Mock-Dienst und Workspace-Mutation (`SPEC-029`, `SPEC-030`) — `slice-telemetry-einbau-mock`; der Port kennt die fünf Span-Namen als Daten und braucht dafür weder `MockUseCase` noch `WorkspaceUseCase`.
- Abruf der Metriken und Export der Spans an einen Collector — die Spezifikation nennt keinen Abrufweg ([welle-betrieb](../welle-betrieb.md) §6); ein verpflichteter externer Dienst scheidet nach Lastenheft §5 aus. Was der Slice an Lesbarkeit im Prozess braucht (für Tests und Lasttest), entscheidet der Architect (Auftrag 1).
- Der Admin-Endpunkt `/__admin/diagnostics`, der die Zähler je Fehlercode ausgibt — `welle-admin-api` (`SPEC-005`); dieser Slice liefert den Zähler, nicht den Endpunkt. Eine Sendung an einen Nehmer, dessen Datei noch nicht angelegt ist (siehe §6).
- Audit-Einträge — `slice-audit-log`; ein Audit-Eintrag ist kein Log-Eintrag (anderer Port, anderer Zweck: Nachweis einer Löschung, nicht Diagnose).

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

- [ ] Port `Telemetry` und Fake: Log-Eintrag mit den Feldern des Katalogs, Span mit Namen und Pflicht-Attributen nach `SPEC-026` bis `SPEC-030`, Zähler je Fehlercode; der Fake hält jedes Argument fest und zählt jede Operation (`LH-QA-05`, `LH-FA-01`).
- [ ] Adapter: schreibt die Logs mit festen Feldnamen, hält Spans und Zähler im Prozess so, wie der Architect den Abruf entscheidet (Auftrag 1), nebenläufig sicher und unter `make test-race` ohne Befund; ein Ausfall der Telemetrie ändert das Ergebnis des Aufrufers nicht.
- [ ] Keine Geheimnisse: ein Canary-Secret in Header, Body und Query-Literal erreicht weder Log noch Span-Attribut noch Metrik (`LH-QA-04`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary, oder macht er Logfelder am Binary sichtbar: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Port wird erst vom Einbau gerufen).
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
| `src/hexagon/ports/support/` | update | Port `Telemetry` neben `Clock` und `IDGenerator` |
| `src/hexagon/ports/support/supporttest/` | update | Fake des Ports (hält jedes Argument fest, zählt jede Operation) |
| `src/adapters/driven/telemetry/` | neu | Adapter: Logs, Spans, Zähler |
| `spec/spezifikation.md` | update | Vertrag mit Randformen: Port, Metriken (§5 nennt nur Spans), Logfelder je Ereignis (Architect) |
| `.a-check.yml` | update, falls nötig | der neue Adapter-Pfad unter den Regeln des Hexagons (`ARC-020`) |
| Tests | neu | je Zusage ein Fall; Canary-Test; nebenläufige Tests laufen unter `make test-race` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft):

1. **Abrufweg und Senke.** Die Spezifikation nennt keinen Abrufweg für Metriken und Spans ([welle-betrieb](../welle-betrieb.md) §6). Zu entscheiden, mit Grund und Folge: ob Spans und Zähler nur im Prozess lesbar sind (Schnittstelle für Tests, den Lasttest und die spätere Diagnose) oder zusätzlich ausgegeben werden (Log-Zeile je Span), und ob eine Bibliothek für Spans und Metriken genutzt wird (die Abhängigkeiten in `go.mod` sind heute `modernc.org/sqlite` und seine Last; eine weitere ist eine Entscheidung mit ADR vor dem Code, mit `go mod verify` und netzlosen Läufen, `AGENTS.md` §3.1) oder die Standardbibliothek genügt.
2. **Form des Ports.** Operationen (Log, Span beginnen und beenden mit Attributen, Zähler erhöhen), Weitergabe über `ctx`, Eltern-Kind-Beziehung der Spans (`odata.replay.match` und `odata.mock.query` „zerlegen“ den Gesamtspan, `LH-QA-01.a`) und die Aufrufstellen als Menge für den Fake (`AGENTS.md` §3.15).
3. **Metriken.** §5 der Spezifikation nennt nur Spans; Namen, Art (Zähler, Verteilung), Einheit und Schlüssel der Metriken, darunter der Zähler je Fehlercode (`SPEC-005`) und die Dauer, stehen vor dem Code in der Spezifikation. Die Menge der Fehlercodes ist geschlossen (Tabelle in `src/hexagon/model`), die Kardinalität damit beschränkt; ein unbekannter Code ist eine Randform.
4. **Logfelder.** Welche Felder je Ereignis Pflicht sind, die Stufen, das Format, und die Abbildung des festen Eintrags des Record-Dienstes (`SPEC-121` Nr. 12, Attribute `mode`, `ereignis`, `error_code`, `service_id`, `session_id`, `exchange_id`, `correlation_id`) auf den Port, so dass der Einbau den Wortlaut nicht ändern muss.
5. **Ausschluss sensibler Inhalte.** Wer erzwingt „keine Credentials, keine Bodies, keine Query-Literale“ (`LH-QA-04`, `SPEC-026` ff. Schlusssatz): der Typ der Attribute, eine Liste erlaubter Schlüssel, eine Prüfung am Wert; und der Fänger (Canary-Test über alle Ausgabewege).
6. **Fehlerabbildung und Nebenläufigkeit.** Das Ergebnis des Aufrufers hängt nie vom Gelingen der Telemetrie ab; was bei Ausfall der Senke geschieht, was bei einem beendeten `ctx` (der Span wird trotzdem beendet?), bei einem Span, der nie beendet wird, und bei doppeltem Beenden; Obergrenze des Speichers.
7. **Katalog des Binary.** Führt der Slice einen Schlüssel (Log-Stufe, Format) oder Code ein, ist es ein Nachzug von `SPEC-110` und ein Eintrag im Handbuch; sonst bleibt das Binary unverändert.
8. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** für die Marke des Planners (Metrik und Formel: `AGENTS.md` §3.14).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 8) | — | — |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jedes Feld des Katalogs, jeder Span-Name und jedes Pflicht-Attribut von `SPEC-026` bis `SPEC-030`, jeder Fehlercode), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt, nicht im Slice-Commit.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein anderer Slice muss in `done/` liegen. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- Code, den der Slice braucht und der schon in `done/` liegt: `src/hexagon/ports/support` und `supporttest` (`slice-clock-idgenerator`), die Fehlercode-Tabelle in `src/hexagon/model` (`slice-fehlermodell-konfiguration`), der Katalog und die Tabelle der Spans in der Spezifikation. Mehr braucht ein Port mit Adapter nicht.
- Handbuch-Punkt der DoD (nur wenn er greift, siehe DoD): `slice-handbuch-geruest` (Ort der Kapitel) und `slice-handbuch-meldungskatalog` (das Kataloge-Gate) aus `welle-benutzerhandbuch` liegen dann in `done/`; ohne sichtbare Änderung am Binary entfällt die Bedingung.

**Unabhängig von `welle-mock` / M3, von M2 und von Linie A, weil** der Port die Namen der Spans als Daten kennt und nichts ruft: `MockUseCase`, `WorkspaceUseCase`-Mutation, HTTP-Adapter und Replay-Dienst sind Orte des Einbaus (`slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-mock`), nicht des Ports. Der Start-Trigger der Welle (`welle-mock` in `done/`, [welle-betrieb](../welle-betrieb.md) §2) gilt für den Abschluss mit Mock-Spans und Lasttest; dieser Slice hängt an keinem Code daraus. Reihenfolge-Vorsicht, kein Blocker: der Slice ändert `src/bootstrap` nicht (die Verdrahtung ist Einbau), so dass er die Dateien der Slices von M2 (`slice-bootstrap-m2`) nicht berührt.

**Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` (`SPEC-128`), sobald die Tabelle *Zusage · Fänger* vorliegt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt (siehe oben); liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht ist dann Logs und Spans einerseits, Metriken und Zähler andererseits.
- `in-progress` → `open` (blockiert): der Abrufweg (Auftrag 1) ist nicht entschieden, oder die Wahl einer Bibliothek braucht eine ADR, die nicht `Accepted` ist.

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

- Abrufweg der Metriken und Spans ist nicht spezifiziert; ohne Entscheidung ist der Lasttest (`slice-lasttest`) nicht lesbar — **Ausgang:** offen; Adresse: Architect (Auftrag 1), vor dem Start.
- Eine Bibliothek für Spans und Metriken ist eine neue Abhängigkeit (Supply-Chain, netzlose Läufe) — **Ausgang:** offen; Adresse: Architect, ADR vor dem Code, falls sie gewählt wird.
- Spezifikationslücke: §5 nennt Spans, keine Metriken; das Lastenheft führt keine Anforderung an Metriken oder Spans außer der Messmethode von `LH-QA-01` — **Ausgang:** offen; Adresse: Architect (Auftrag 3), Meldung an den Planner, falls eine Anforderung im Lastenheft fehlt.
- Sendung an einen Nehmer ohne Datei: der Zähler je Fehlercode ist der Gegenstand von `/__admin/diagnostics` in `welle-admin-api` (Slice `slice-inspection-use-case` der Tabelle dort); dessen Datei besteht nicht — **Ausgang:** offen; Adresse: Planner beim Anlegen der Slices von `welle-admin-api` (`AGENTS.md` §3.13: erst der Nehmer führt den Punkt, dann nennt der Geber ihn).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund; der Architect füllt die Zellen vor dem Code):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | sensibler Wert | Abbruch (`ctx`) | Wiederholung |
|---|---|---|---|---|---|---|
| Log schreiben | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Span beginnen und beenden | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Zähler je Fehlercode erhöhen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Metrik und Span lesen (Abruf im Prozess) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Adapter ist nebenläufig und der Fake umhüllt `ctx`: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) gilt für Pfade, die kein Lauf ausführt (Ausfall der Senke, Abbruch), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für den Fake des Ports, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) für jede Wirkung, die der Detektor nicht meldet (Zähler unter Last: Rate mit Zahl der Läufe, `AGENTS.md` §3.10); offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
