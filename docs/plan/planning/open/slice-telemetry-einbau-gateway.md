# Slice slice-telemetry-einbau-gateway: Einbau von `Telemetry` in HTTP-Adapter (Gesamtspan), Record-Dienst (Capture-Span) und Replay-Dienst (Match-Span)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-01`, `LH-FA-01`, `LH-QA-05`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-026`, `SPEC-027`, `SPEC-028` (spezifikation.md §5) · `LH-QA-01.a` (Messpunkt: Gesamtspan `odata.gateway.request`) · `SPEC-121` Nr. 12 (Protokoll des Record-Dienstes) · `SPEC-121` Nr. 13 (Capture-Status) · `SPEC-109`, `SPEC-113` (Fehlercodes für den Zähler) · `ARC-002`, `ARC-005`, `ARC-012`, `ARC-021`.
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

**Ziel:** Der HTTP-Adapter öffnet für jeden Request den Gesamtspan `odata.gateway.request` (`SPEC-026`) vom Request-Eingang bis zum ersten Response-Byte, der Record-Dienst setzt `odata.record.capture` (`SPEC-027`) und der Replay-Dienst `odata.replay.match` (`SPEC-028`), jeweils mit den Pflicht-Attributen; der Record-Dienst schreibt seinen Protokolleintrag (`SPEC-121` Nr. 12) über den Port `Telemetry` statt über einen `*slog.Logger`, ohne dessen Wortlaut zu ändern; der Zähler je Fehlercode steigt für jeden Fehler, den der Adapter als Fehlerobjekt schreibt; der Composition Root verdrahtet den Adapter. Kein Query-Literal und kein Body wird Attribut.

**Übernimmt (Sendung von `slice-telemetry-port`):** der Einbau in die Pfade `record` und `replay`; der Port selbst und der Fake sind dort geliefert.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Port und Adapter `Telemetry` — `slice-telemetry-port`; dieser Slice ruft ihn nur.
- Spans von Mock-Dienst und Workspace-Mutation (`SPEC-029`, `SPEC-030`) — `slice-telemetry-einbau-mock`; der Gesamtspan im HTTP-Adapter ist modusunabhängig und trägt auch den Modus `mock`, sobald dieser existiert (der Adapter kennt den Modus nicht, `architecture.md` §2.5; wie das Attribut `recorder.mode` zu ihm kommt, entscheidet der Architect, Auftrag 2).
- Spans und Zähler des Hybrid-Fallbacks — kein Slice von `welle-betrieb` nimmt sie an, obwohl [welle-hybrid](../welle-hybrid.md) §6 sie `welle-betrieb` zuweist; keine Adresse (siehe §6, Ausgang *offen*, Adresse: Planner).
- Der Lasttest und seine Messung — `slice-lasttest`; dieser Slice liefert den Messpunkt.
- Das Handbuch-Kapitel über Logfelder, soweit der Slice keine neuen sichtbaren Felder einführt (siehe DoD).

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

- [ ] Gesamtspan im HTTP-Adapter: Messpunkt Request-Eingang bis erstes Response-Byte, Attribute `http.request.method`, `http.response.status_code`, `recorder.mode`; auch bei Fehlerobjekt, ohne Antwort und bei Abbruch des Clients wird der Span beendet (`SPEC-026`, `LH-QA-01.a`).
- [ ] Capture-Span im Record-Dienst und Match-Span im Replay-Dienst mit ihren Pflicht-Attributen; der Protokolleintrag des Record-Dienstes läuft über den Port, Attribute und Text unverändert (`SPEC-027`, `SPEC-028`, `SPEC-121` Nr. 12); der Zähler je Fehlercode steigt für jeden Code, den der Test auslöst.
- [ ] Verdrahtung im Composition Root und Test durch Record und Replay: die Spans `SPEC-026` bis `SPEC-028` mit ihren Pflicht-Attributen, ohne Query-Literal und Body (Canary), der Gesamtspan umfasst den Match-Span.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Macht der Slice Logfelder, einen Fehlercode oder Schlüssel am Binary sichtbar (die Logzeilen des laufenden Prozesses): Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb.
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
| `src/adapters/driving/httpodata/` | update | Gesamtspan, Zähler für geschriebene Fehlerobjekte |
| `src/hexagon/services/record/` | update | Capture-Span; Protokoll über den Port statt `*slog.Logger` (Konstruktor, Konfiguration) |
| `src/hexagon/services/replay/` | update | Match-Span |
| `src/bootstrap/` | update | Verdrahtung des Adapters `Telemetry` |
| `spec/spezifikation.md` | update | `SPEC-121` Nr. 12 (kein Port `Telemetry` mehr „noch nicht“), Aufrufstellen und Randformen (Architect) |
| Tests | update, neu | Fakes der Ports; der Test durch Record und Replay |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Aufträge an den Architect:**

1. **Messpunkt.** Wo der Gesamtspan beginnt (nach dem Lesen der Request-Zeile, vor der Abbildung auf `inbound.Anfrage`) und endet (erstes Response-Byte: `WriteHeader`, erster `Write`, oder Fehlerobjekt); Verhalten bei Antwort ohne Body, bei Abbruch vor dem ersten Byte, bei `Flush`; was „im Prozess“ gegen `LH-QA-01.a` heißt.
2. **Attribute.** Woher `recorder.mode` kommt (der Adapter kennt den Modus nicht); wann `recorder.capture_status`, `recorder.match_strategy` und `recorder.match_result` feststehen (nach dem Ende des Exchange, bei Abbruch, bei `Drain`); Wertemengen, auch für den Fehlerfall.
3. **Eltern-Kind.** Der Match-Span liegt im Gesamtspan (Weitergabe über `ctx`); die Reihenfolge, wenn der Capture-Span den Gesamtspan überlebt (der Exchange wird nach dem letzten Byte angehängt).
4. **Ersatz des Loggers.** Der feste Eintrag des Record-Dienstes (`SPEC-121` Nr. 12) geht über den Port, Text und Attribute bleiben; der Konstruktor verliert das Feld `Log` oder behält es nicht; wie die Fänger der bisherigen Einträge (Stufe, Attribute) unverändert bleiben.
5. **Fehlerzähler.** Welche Fehler den Zähler erhöhen (jeder Code, den der Adapter schreibt, auch `ODR-E9001`; Kontextfehler ohne Code nicht), und wo (Adapter oder Dienst), so dass kein Fehler doppelt zählt.
6. **Aufrufstellen als Menge** (`AGENTS.md` §3.15): jeder Port, jede Stelle, an der der Port gerufen wird, ein Fall je Stelle; der Fake hält jedes Argument fest.
7. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 7) | — | — |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jeder Fehlercode, jede Aufrufstelle des Ports, jeder Wert von `recorder.capture_status`), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-telemetry-port`, `slice-http-adapter`, `slice-replay-service` und `slice-bootstrap-m2` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-port`: der Port und sein Fake, die dieser Slice ruft.
- `slice-http-adapter` (`welle-record-replay`): der Ort des Gesamtspans, `src/adapters/driving/httpodata` besteht heute nicht.
- `slice-replay-service` (`welle-record-replay`): der Ort des Match-Spans, `src/hexagon/services/replay` besteht heute nicht. Der Record-Dienst (`slice-record-service`) liegt in `done/`.
- `slice-bootstrap-m2` (`welle-record-replay`): der Composition Root, der Adapter und Dienste je Modus verdrahtet, und der Handle, über den der Test Record und Replay durchläuft (kein CLI, keine Admin-API).
- Handbuch-Punkt der DoD (nur wenn er greift): `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` aus `welle-benutzerhandbuch` liegen dann in `done/` (das Gate liest Binary erst nach `slice-bootstrap-m2`).

**Unabhängig von `welle-mock` / M3, weil** der Slice nur die Pfade `record` und `replay` einbaut; `MockUseCase`, `WorkspaceUseCase`-Mutation und der Modus `mock` im Composition Root (`slice-bootstrap-m3`) sind nicht seine Orte (`slice-telemetry-einbau-mock`). **Nicht unabhängig von M2:** der Einbau braucht den HTTP-Adapter und den Replay-Dienst aus Linie A und den Composition Root von `slice-bootstrap-m2`; ohne sie gibt es keinen Ort für den Gesamtspan. Der letzte Slice von `welle-record-replay`, `slice-protokolltreue`, ist keine Bedingung.

**Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` (`SPEC-128`), sobald die Tabelle *Zusage · Fänger* vorliegt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht ist dann der HTTP-Adapter (Gesamtspan) gegen die Dienste (Capture, Match).
- `in-progress` → `open` (blockiert): der Gesamtspan lässt sich nicht ohne Typ der HTTP-Bibliothek im Port ausdrücken, oder der Wortlaut von `SPEC-121` Nr. 12 lässt sich über den Port nicht erhalten (siehe §6).

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

- Der Einbau ändert den Konstruktor des Record-Dienstes (Feld `Log`); Fänger und Tests der Vorgänger, die das Feld setzen, brechen — **Ausgang:** offen; Auftrag 4, die Fänger der bisherigen Einträge bleiben unverändert grün.
- Messpunkt „erstes Response-Byte“ im Adapter einer HTTP-Bibliothek, die `WriteHeader` und erstes `Write` trennt; ein Span, der den Abbruch des Clients nicht überlebt — **Ausgang:** offen; Adresse: Architect (Auftrag 1), nebenläufige Fänger ordnen mit Kanälen und laufen unter `make test-race`.
- Pfad ohne Lauf unter dem Detektor (Abbruch, Ausfall der Senke) — **Ausgang:** offen; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): Wirkungs-Fänger oder akzeptiertes Negativ mit Grund und Folge (`AGENTS.md` §3.10).
- Sendung ohne Nehmer: Spans und Zähler des Hybrid-Fallbacks. `welle-hybrid` §6 weist sie `welle-betrieb` zu, die Tabelle von [welle-betrieb](../welle-betrieb.md) §4 führt sie in keinem Slice — **Ausgang:** offen; Adresse: Planner (Frage an den User: eigener Slice in `welle-betrieb` oder Teil von `slice-telemetry-einbau-mock`, oder in `welle-hybrid`).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch (Client, `ctx`, `Drain`) | Wiederholung | Herunterfahren |
|---|---|---|---|---|---|---|
| Gesamtspan öffnen und beenden (Request-Eingang bis erstes Byte) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Capture-Span (Exchange anhängen, Capture-Status) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Match-Span (Strategie, Ergebnis) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Protokolleintrag des Record-Dienstes über den Port | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Zähler je Fehlercode (Fehlerobjekt des Adapters) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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
