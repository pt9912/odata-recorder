# Slice slice-handbuch-telemetrie: Kapitel zu den Logzeilen des Prozesses mit gebundenen Beispielen und die Maske unbestimmter Felder im Beispiel-Format

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md) (Zuordnung des Planners, 2026-11-21, nach dem Muster von `slice-handbuch-mock`: ein Kettenslice der Gate-Kette würde die Closure von [welle-betrieb](../welle-betrieb.md) an die Kette hängen, siehe §6).

**Bezug:** `LH-QA-05` (diagnostischer Kontext der Meldungen), `LH-QA-04` (keine Geheimnisse in regulären Logs), `LH-RB-04` (dokumentierte Schutzmaßnahmen), `LH-RB-01` (Ist-Zustand des Handbuchs), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-200` Nr. 1, 6 d, 7 a, 7 b, 7 d und die zugehörigen Nummern des Beispiel-Formats und des Laufs (`harness/vertraege/SPEC-200-handbuch.md`; der Architect erweitert sie vor dem Code) · `SPEC-231` Nr. 1, 2, 6, 7 (Zeile, Zeit, Span-Zeile, Verdrahtung) · `SPEC-232` Nr. 2 (sensible Inhalte) · `SPEC-026` bis `SPEC-030` und `SPEC-290` (Spans) · `SPEC-139` Nr. 8 und 13 d (Zeilen des Starts, sichtbar am Binary) · `SPEC-129` Nr. 13 (Einträge des Dienstes `workspace`) · `SPEC-201` (Kataloge: kein Eintrag erwartet).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner). **Datum:** 2026-11-21.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Handbuch beschreibt in einem Kapitel *Logzeilen*, was das Binary während `start` auf die Fehlerausgabe schreibt: die Form der JSON-Zeile (Felder und ihre Reihenfolge), die Zeile eines Spans je Anfrage, die Zeilen des Starts, die Senke (Fehlerausgabe, die Standardausgabe bleibt leer) und was nie in eine Zeile gelangt (Query-Literale, Bodies, Header-Werte, Geheimnisse). Jede dieser Aussagen steht als gebundenes Beispiel (`SPEC-200` Nr. 1). Dafür erweitert der Slice den Vertrag `SPEC-200` und das Werkzeug `pruefung handbuch` um die Maske der Felder, die kein Lauf gleich schreibt (Zeitstempel, Dauer, gegebenenfalls Kennung des Spans), und, soweit `slice-handbuch-mock` sie nicht liefert, um Anfragen an den Listener im Beispiel-Format.

**Übernimmt (Entscheidung des Users vom 2026-11-21 und Sendung von `slice-telemetry-einbau-mock`, `slice-telemetry-einbau-mock-mutation` und `slice-telemetry-einbau-gateway` (nach dem Schnitt vom 2026-11-22 dessen Teile 1 und 3, `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-record`; Teil 2, `slice-telemetry-einbau-replay`, schließt es ebenfalls aus), Geber der Reihe nach; `AGENTS.md` §3.13):** das Handbuch-Kapitel zu den Logzeilen des Prozesses (Zeile des Spans je Anfrage, JSON-Zeilen der Einträge der Dienste). Die Geber haben das Kapitel ausgeschlossen, weil ein Verhalten im Handbuch nur als gebundenes Beispiel steht (`SPEC-200` Nr. 1) und das Beispiel-Format keine Anfrage an den Listener kennt (`SPEC-200` Nr. 6 d, 7 b); die Frage stand bei `welle-betrieb` §5, Punkt c, und ist entschieden: eigener Slice, dieser. Gelesen am 2026-11-21, `AGENTS.md` §3.13: kein Ausschluss eines Nachbarn trifft die Sendung. `slice-handbuch-mock` §1 schließt „Telemetrie“ unter den Kapiteln der Folgewellen aus und trägt dort jetzt diesen Slice als Nehmer; `slice-handbuch-abnahme` §1 schließt Kapitel über Hybrid und Admin-API aus, nicht Logzeilen, und sein Probelauf prüft, was das Handbuch bei der Closure sagt; `welle-benutzerhandbuch` §6 führte Telemetrie unter den Folgewellen und nimmt sie seit demselben Commit aus (§4 und §6 dort).

**Übernimmt (bedingt; Geber `slice-handbuch-mock`, Planner 2026-11-21):** Anfragen an den Listener im Beispiel-Format (`SPEC-200` Nr. 6 d, 7 b). Der Architect von `slice-handbuch-mock` entscheidet vor dem Code, ob sie in dessen Erweiterung gehören (Beispiel zu `ETAG_ENABLED`); trägt er sie, setzt dieser Slice sie voraus und ändert sie nicht, gibt er sie zurück, führt dieser Slice sie als Teil der Erweiterung von `SPEC-200` (§6, Ausgang *offen*, Adresse Architect).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Umgebung im Beispiel-Format und Lauf, der den Dauerlauf mit einem Signal beendet — `slice-handbuch-mock` (Sendung von `slice-bootstrap-m3`); beide sind Bedingung dieses Slice (§4), er ändert sie nicht. Beide Erweiterungen am selben Vertrag in zwei Slices zugleich zu führen, hieße zwei Besitzer der Gate-Kette (Nr. 7 der Übergangsregeln).
- Gebundene Beispiele der Zeile von `odata.mock.query` und `odata.workspace.mutation` und der Einträge des Dienstes `workspace` — nicht belegbar, solange ein Workspace von außen nicht entsteht: Das gebaute Binary kennt die Kommandos `start` und `version`; ein Workspace entsteht über die Admin-API (`welle-admin-api`), ein Span des Mock-Dienstes trägt nur eine Anfrage mit gefundenem Workspace (`SPEC-290` Nr. 2 und 3 a), die Einträge des Dienstes `workspace` nur ein Fehler der Bereinigung nach dem Löschen (`SPEC-129` Nr. 13). Das Kapitel sagt diese Zeilen nicht (`SPEC-200` Nr. 1); die Frage, wie sie später belegt werden, steht in §6 mit Adresse User.
- Spans `odata.record.capture` und `odata.replay.match` und die Einträge `capture`, `upstream` und `ablage` des Record-Dienstes als Beispiel — sie brauchen einen Upstream, eine offene Session oder importierte Austausche; das Format kennt weder Netz noch Upstream (`SPEC-200` Nr. 7 b; `slice-handbuch-mock` §1 schließt beides ebenfalls aus). Der Gesamtspan `odata.gateway.request` trägt jeden Modus und braucht keines von beiden.
- Span und Zähler des Hybrid-Fallbacks — das Binary schreibt sie erst mit [slice-telemetry-einbau-hybrid](../next/slice-telemetry-einbau-hybrid.md) (`welle-betrieb` §5, Punkt b; Entscheidung des Users vom 2026-11-23), der nach [slice-telemetry-einbau-record](../done/slice-telemetry-einbau-record.md) läuft und keine Bedingung dieses Slice ist; die Zeile braucht einen Upstream oder importierte Austausche und ist als gebundenes Beispiel nicht erreichbar (`SPEC-200` Nr. 7 b), das Kapitel sagt sie nicht (`SPEC-200` Nr. 1).
- Metriken, ihr Abruf und der Export von Spans — das Binary hat keinen Weg dafür (`spezifikation.md` §5, letzter Absatz); was es nicht tut, steht nicht im Handbuch.
- Die Auswertung der Span-Zeile durch den Lasttest — `slice-lasttest`; ein Werkzeug außerhalb des Binary.
- Der Probelauf durch eine Rolle ohne Vorwissen — `slice-handbuch-abnahme`, in der Reihenfolge nach diesem Slice (`welle-benutzerhandbuch` §5).
- Eine Einstellung für Stufe, Format oder Senke der Zeile — es gibt keine (`SPEC-231` Nr. 7); das Kapitel sagt, was die Beispiele zeigen, und führt keinen Schlüssel ein.

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

- [ ] `SPEC-200` und `pruefung handbuch` kennen eine Maske für die Felder der Zeile, die kein Lauf gleich schreibt, und vergleichen die Fehlerausgabe damit; sind Anfragen an den Listener im Beispiel-Format bis dahin nicht geliefert (`slice-handbuch-mock`), kennen sie auch diese (Vertrag vom Architect vor dem Code; Tabelle *Zusage · Fänger* in §3, Raster in §6); die Gegenproben zu `pruefung` fangen je Zusage eine Mutation (`LH-QA-05`).
- [ ] Kapitel *Logzeilen* im Handbuch mit gebundenen Beispielen (Ist-Zustand, je Beispiel eine Datei gegen das gebaute Binary in der Stufe `handbuch`): Form der Zeile, die Zeile des Gesamtspans einer Anfrage, die Zeilen des Starts, die Senke (Standardausgabe leer) und ein Beispiel, dessen Anfrage ein Query-Literal und einen Header-Wert trägt, die in der Zeile fehlen; Aussagen ohne Beispiel stehen nicht im Kapitel (`LH-QA-05`, `LH-QA-04`, `LH-RB-04`). Das Kapitel führt keinen Fehlercode, Schlüssel oder Kommando ein, `make katalog-check` bleibt grün.
- [ ] `make gates` grün, `make gegenproben` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Handbuch ist das Update; Kopf (Stand, Software-Version) nachgezogen; es entsteht kein Änderungsverlauf (`welle-benutzerhandbuch` §1).
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
| `harness/vertraege/SPEC-200-handbuch.md` | update | Maske der unbestimmten Felder im Beispiel-Format und im Vergleich (Nr. 6, 7 d), gegebenenfalls Anfragen (Nr. 6 d, 7 b); der Architect schreibt sie vor dem Code |
| `tools/pruefung/` (Programm `pruefung`, Stufe `handbuch`) | update | Maske im Vergleich der Fehlerausgabe, gegebenenfalls Anfragen im Lauf; Pfade nennt der Architect |
| `tools/gegenproben/` und `harness/gegenproben.stempel` | update | je Zusage der Erweiterung ein Fall der Gegenprobe (Nr. 7 der Übergangsregeln: Kettenpfade) |
| `docs/user/benutzerhandbuch.md`, `docs/user/beispiele/` | update, neu | Kapitel *Logzeilen*, Beispiel-Dateien (Gesamtspan, Start, Query-Literal und Header-Wert fehlen in der Zeile) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit der Erweiterung von `SPEC-200` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Aufträge an den Architect:**

1. **Maske.** Welche Felder die Prüfung maskiert und wie das Beispiel sie schreibt (Menge aus dem Wortlaut von `SPEC-231` Nr. 1, 2, 3 und 6: `timestamp` und `duration_ms` der JSON-Zeile, `span_id` und `parent_id` der Span-Zeile, `time` der Textzeile des Starts, der Port in `listen`); welche Felder bleiben bytegleich (Name des Spans, `level`, `msg`, `http.request.method`, `http.response.status_code`, `recorder.mode`, `error_code`), damit die Maske das Beispiel nicht wertlos macht; Verhalten bei fehlendem, doppeltem, leerem und fremdem Feld und bei einer Zeile, die kein JSON-Objekt ist. Die Zeilen eines Aufrufers stehen in der Reihenfolge der Aufrufe, die gleichzeitiger Aufrufe nicht (`SPEC-232` Nr. 5 d): das Beispiel stellt seine Anfragen nacheinander, und der Vertrag sagt, was ein Beispiel mit gleichzeitigen Anfragen ist.
2. **Anfragen an den Listener.** Ob `slice-handbuch-mock` sie liefert (Entscheidung des Architect dort, §6); wenn nicht: Form der Anfrage im Beispiel (Methode, Pfad mit Query, Header, Body), Zeitpunkt (nach der Zeile `start: bereit`, nie davor), Vergleich der Antwort (gehört die Antwort zum Beispiel oder nur die Fehlerausgabe), Loopback in der Stufe `handbuch` (kein Netz nach außen, kein Mehrprozess-Lauf; die Grenze des letzten Ausschlusses in `slice-handbuch-mock` §1).
3. **Erreichbarkeit.** Welche Zeilen ein leeres `DATA_DIR` ohne Upstream erzeugt: der Gesamtspan einer Anfrage ohne Workspace-Header (`ODR-E5003`) und einer mit unbekanntem Workspace (`ODR-E7001`) in `mock` und `hybrid`, die Zeilen des Starts; ob `slice-telemetry-einbau-gateway` den Gesamtspan auch für abgewiesene Anfragen schreibt (sein Ziel nennt „jeden Request“); welche Zeile in `replay` und `record` ohne Session oder Upstream entsteht. Nur die Zeilen, die ein Beispiel erreicht, stehen im Kapitel.
4. **Aussagen ohne Beispiel.** Welche Aussagen des Kapitels der Klasse K1 angehören (Beispiel nötig) und welche K6 (Aufbau, kein Lauf); für „keine Einstellung für Stufe, Format und Senke“ und „unterhalb von INFO nichts“ (`SPEC-231` Nr. 2, 7) die Entscheidung, ob sie als akzeptiertes Negativ in `SPEC-200` Nr. 14 b stehen oder im Kapitel entfallen.
5. **Zeilen des Starts.** Ob die Textzeilen `start: bereit` und die Warnung zur abgeschalteten Prüfung des Zertifikats (`SPEC-139` Nr. 8) zum Kapitel gehören (sie stehen auf demselben Strom in einer anderen Form als die JSON-Zeile) und ob das Beispiel des Modus `hybrid` den Satz zur Warnung im Abschnitt *Modus hybrid* belegt oder dieser Satz unberührt bleibt.
6. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** für die Marke des Planners (`SPEC-128` Nr. 1; Handbuch und Vertrag zählen nicht, `SPEC-128` Nr. 3 m).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 6) | — | — |

**Betriebsregeln für die Rollen dieses Slice:** Dateien des Repos nur mit `Edit` und `Write`, kein `sed -i`, kein Heredoc auf Repo-Dateien (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`). Werte mit Backslash-u-Folge, BOM, NBSP oder Zeichen jenseits von ASCII entstehen aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18); das gilt für die Beispiel-Dateien, deren Zeilen Escape-Folgen tragen können. Docker-Läufe mit `--memory` und wenige parallel.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): `slice-handbuch-mock` liegt in `done/` (Umgebung im Beispiel-Format, Lauf mit Signal, Entscheidung zu Anfragen), `slice-telemetry-einbau-gateway` (Gesamtspan jeder Anfrage) und `slice-telemetry-einbau-record` (Einträge der Dienste `record`, `session` und `maintenance` als JSON-Zeile) liegen in `done/` (Teil 1 und Teil 3 des Schnitts vom 2026-11-22; `slice-telemetry-einbau-replay`, Teil 2, ist keine Bedingung), kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln), und der Architect hat `SPEC-200` erweitert und die Tabelle in §3 sowie das Raster in §6 gefüllt, bevor der erste Code-Commit entsteht. Beim Start liest der Planner §1 *Ausdrücklich NICHT* und die DoD der Nachbarn neu (`AGENTS.md` §3.13) und gleicht die Zeilen der Tabelle unten gegen `main` ab.

**Entscheidung (Planner, 2026-11-21): das Kapitel kommt einmal, nach `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-record`, nicht gestaffelt nach jedem Einbau-Slice.** Welche Zeile des Binary mit welchem Slice sichtbar wird, und ob ein Beispiel sie ohne Workspace und ohne Upstream erreicht:

| Zeile auf der Fehlerausgabe | sichtbar mit | als gebundenes Beispiel erreichbar |
|---|---|---|
| Textzeilen des Starts: `start: bereit`, Warnung bei abgeschalteter Prüfung des Zertifikats | dem Binary heute | ja (Maske für `time`) |
| Span `odata.mock.query` (`mock`, `hybrid`), Einträge des Dienstes `workspace` | `slice-telemetry-einbau-mock` (`done/`) | nein: nur mit gefundenem Workspace, den kein Kommando des Binary anlegt |
| Span `odata.workspace.mutation` | `slice-telemetry-einbau-mock-mutation` | nein, derselbe Grund |
| Span `odata.gateway.request` (jeder Modus) | `slice-telemetry-einbau-gateway` | ja: eine Anfrage ohne Workspace-Header genügt (Architect bestätigt, Auftrag 3) |
| Spans `odata.record.capture`, `odata.replay.match`, Einträge `capture`, `upstream`, `ablage` und die der Dienste `session` und `maintenance` | `slice-telemetry-einbau-record` (Capture-Span und Einträge), `slice-telemetry-einbau-replay` (Match-Span) | nein: Upstream, offene Session oder importierte Austausche |
| Span und Zähler des Hybrid-Fallbacks | `slice-telemetry-einbau-hybrid` (nach Teil 3, keine Bedingung dieses Slice) | nein: Upstream oder importierte Austausche |

Gründe für „einmal, danach“: (a) Der Gesamtspan ist die einzige Span-Zeile, die ein Beispiel ohne Workspace erreicht; vor `slice-telemetry-einbau-gateway` zeigte das Kapitel keine Zeile eines Spans. (b) `slice-telemetry-einbau-record` stellt die Einträge der Dienste `record`, `session` und `maintenance` von der Textzeile auf die JSON-Zeile um (`SPEC-231` Nr. 7); ein Kapitel davor beschriebe dieselbe Ausgabe zweimal. (c) Der Slice ändert `tools/pruefung` und die Gegenproben und belegt die Gate-Kette einmal, nicht je Einbau-Slice. `slice-telemetry-einbau-mock-mutation` ist keine Bedingung: Er führt keine neue Form der Zeile ein, und seine Zeile ist nur mit Workspace sichtbar; liegt er beim Start in `done/` oder nicht, ändert das das Kapitel nicht.

**Gate-Kette (Nr. 7 der Übergangsregeln):** Linie C, Kettenslice. Er ändert `tools/` (Programm `pruefung`, Stufe `handbuch`), `tools/gegenproben/` und den Stempel `harness/gegenproben.stempel`, also Pfade aus Nr. 7; er läuft nie zugleich mit einem anderen Kettenslice und liegt in der Reihenfolge von `welle-benutzerhandbuch` §5 nach `slice-handbuch-mock` und vor `slice-handbuch-abnahme`. Ändert der Architect das `Dockerfile` oder `Makefile` zusätzlich, steht das in §3.

**Größenmarke** (`AGENTS.md` §3.14, vorläufig als Prosa, keine Rechnung des Architect, keine Marke): Der Plan ist beim Anlegen so geschnitten, dass er unter der Schnittgrenze von `SPEC-128` Nr. 1 (*S* < 2223) bleibt: Umgebung und Signal liegen bei `slice-handbuch-mock` (dort vorläufig *S* rund 900), die Mock-Zeilen, Upstream-Zeilen und der Hybrid-Fallback sind ausgeschlossen, das Kapitel und der Vertrag zählen nicht (`SPEC-128` Nr. 3 m: `docs/` und `harness/vertraege/` zählen nicht); es bleibt die Maske samt Gegenproben und, falls `slice-handbuch-mock` sie nicht liefert, die Anfragen. Das ist eine Größenordnung wie die Erweiterung dort, keine Messung. Der Block `groessenmarke` fehlt bewusst; der Planner setzt `basis`, `schaetzung` und `marke` nach `SPEC-128` Nr. 1 aus der Tabelle des Architect in eigenem Commit vor dem ersten Code-Commit (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Plans, Tests eingeschlossen). Nebenläufig zählt der Lauf mit Signal und Anfragen, weil das Werkzeug ein Binary startet, anspricht und beendet.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nur vor dem Code über der Schnittgrenze nach `SPEC-128` Nr. 1; die Naht, falls der Slice die Anfragen trägt: Anfragen im Beispiel-Format und Maske sind getrennte Zusagen ohne gemeinsamen Code außer dem Lauf; vorab benannter Teil nach dem Code: das Kapitel samt Beispiel-Dateien als eigener Commit nach dem Werkzeug (zählt nicht zur Marke).
- `in-progress` → `open` (blockiert): der Architect findet, dass `slice-telemetry-einbau-gateway` den Gesamtspan für eine abgewiesene Anfrage nicht schreibt (dann erreicht kein Beispiel eine Span-Zeile, Rückgabe an den Planner und den User), oder die Erweiterung lockert ein Gate oder ändert das `Dockerfile`, während ein anderer Kettenslice die Kette hält.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, `make handbuch-check` fährt die Beispiele des Kapitels gegen das gebaute Binary, und je Zusage der Erweiterung ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor. Der Slice gehört zu `welle-benutzerhandbuch` (Kapitel zu den Logzeilen, Closure-Trigger dort §3) und ist kein Kriterium von `welle-betrieb`.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Erweiterung von `SPEC-200` (Maske, gegebenenfalls Anfragen) braucht den Architect vor dem Code; sie ändert `pruefung handbuch`, seine Gegenproben und das Format der Beispiel-Datei und ist keine Entscheidung des Implementers (`AGENTS.md` §3.12) — **Ausgang:** offen; Adresse: Architect (vor dem ersten Code-Commit).
- Anfragen an den Listener im Beispiel-Format: `slice-handbuch-mock` entscheidet (Beispiel zu `ETAG_ENABLED`); gibt der Architect dort sie zurück, trägt dieser Slice sie (§1, *Übernimmt bedingt*) und die Prosa-Einordnung der Größe in §4 gilt neu — **Ausgang:** offen; Adresse: Architect von `slice-handbuch-mock` (Entscheidung), Planner (Start, Marke).
- Die Zeilen von `odata.mock.query`, `odata.workspace.mutation` und der Einträge des Dienstes `workspace` bleiben ohne Beispiel, solange kein Weg einen Workspace von außen anlegt; das Kapitel sagt sie dann nicht, obwohl das Binary sie schreibt (`welle-betrieb` §5, Punkt c nannte sie ausdrücklich) — **Ausgang:** weiter offen; Adresse: User (Entscheidung: (a) das Kapitel zu diesen Zeilen folgt dem Kapitel zur Admin-API aus `welle-admin-api`, die Workspaces anlegt; (b) das Beispiel-Format erhält einen vorbefüllten Zustand von `DATA_DIR`, ein größerer Eingriff in `SPEC-200`; (c) die Zeilen bleiben ohne Kapitel) und Planner (legt einen Folge-Slice erst nach der Entscheidung an; kein Register-Eintrag, weil keine Beobachtung eines Vorgangs, sondern eine Entscheidung).
- Der Gesamtspan entsteht nicht für eine abgewiesene Anfrage (ohne Workspace-Header oder mit unbekanntem Workspace), oder `slice-telemetry-einbau-gateway` schreibt ihn anders, als §4 annimmt — **Ausgang:** offen; Adresse: Architect (Auftrag 3, vor dem Code); der Plan von `slice-telemetry-einbau-gateway` (Auftrag 1 dort) legt den Messpunkt fest.
- Die Maske macht das Beispiel wertlos (ein maskiertes Feld, dessen Wert die Aussage trägt, oder eine Maske, die jede Zeile gleich macht) — `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, `verkörpert` in `AGENTS.md` §3.15 — **Ausgang:** offen; Fänger: Tabelle *Zusage · Fänger* mit je einer Mutation je maskiertem Feld und je Feld, das bytegleich bleibt.
- Zeilen gleichzeitiger Aufrufe stehen in unbestimmter Reihenfolge (`SPEC-232` Nr. 5 d); ein Beispiel mit mehreren Anfragen zugleich wäre nicht reproduzierbar (`LH-QA-06`) — **Ausgang:** offen; Adresse: Architect (Auftrag 1: Beispiele stellen ihre Anfragen nacheinander; der Vertrag nennt die Randform).
- Das Handbuch sagt mehr zu, als ein Test prüft: Aussagen über nicht geschriebene Zeilen (Stufen unter INFO, Bodies) lassen sich als Beispiel nicht belegen — **Ausgang:** offen; Adresse: Architect (Auftrag 4: akzeptiertes Negativ in `SPEC-200` Nr. 14 b oder keine Aussage).
- Zuordnung der Welle: `welle-benutzerhandbuch` statt `welle-betrieb`; `welle-betrieb` §3 verlangt, dass das Handbuch alles abdeckt, was das Binary dieser Welle kann — **Ausgang:** entschieden (Planner, 2026-11-21, Muster `slice-handbuch-mock` und `welle-mock`): der Slice ist ein Kettenslice der Linie C und hinge `welle-betrieb` an die Gate-Kette; `welle-betrieb` §3 nennt das Kapitel zu den Logzeilen ausdrücklich nicht als Kriterium (Eintrag im selben Commit); die Zuordnung ändert der User, wenn er sie anders will.
- Die Gate-Kette ist belegt, wenn der Slice startet (`slice-handbuch-mock` davor, `slice-gegenproben-eichung` und weitere Kettenslices in `open/`) — **Ausgang:** offen; Adresse: Planner (Reihenfolge beim Start, Nr. 7 der Übergangsregeln).
- Das Kapitel unterliegt den Mengen A bis C von `SPEC-200` (Nr. 10 bis 12; die Menge C gilt auch in Zitaten, Nr. 13 c); eine Zeile des Binary mit einer Kennung der Planung wäre ein Befund am Binary, nicht am Handbuch — **Ausgang:** offen; Adresse: Architect (Auftrag 3, die Menge der Zeilen gegen die Muster prüfen).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster unten als Gerüst; die Zellen füllt der Architect vor dem ersten Code-Commit, eine leere Zelle ist ein Fehler des Plans — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Maske im Beispiel-Format lesen | (Architect, Auftrag 1) | (Architect, Auftrag 1) | (Architect, Auftrag 1) | (Architect, Auftrag 1) | (Architect, Auftrag 1) |
| Zeile der Fehlerausgabe gegen Beispiel und Maske vergleichen (Feld fehlt, doppelt, leer, fremd; Zeile kein JSON-Objekt) | (Architect, Auftrag 1) | (Architect, Auftrag 1) | (Architect, Auftrag 1) | (Architect, Auftrag 1) | (Architect, Auftrag 1) |
| Anfrage im Lauf stellen (falls hier; sonst `slice-handbuch-mock`) | (Architect, Auftrag 2) | (Architect, Auftrag 2) | (Architect, Auftrag 2) | (Architect, Auftrag 2) | (Architect, Auftrag 2) |
| Beispiel des Kapitels binden (Handbuch gegen Datei) | (Architect, Auftrag 3) | (Architect, Auftrag 3) | (Architect, Auftrag 3) | gilt nicht: reiner Lesevorgang (Architect bestätigt) | (Architect, Auftrag 3) |

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

**Vorgelagert — offene Beobachtungen sichten:** Der Planner liest das Register (`../observations/`) beim Start neu, denn bis dahin schreiben andere Slices es fort; zu erwarten sind `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle`, `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (für den Lauf mit Signal und Anfragen).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
