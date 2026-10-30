# Slice slice-lasttest: Lasttest nach `LH-QA-01.a` mit Messpunkt am Gesamtspan für Replay und Mock

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `LH-QA-01.a` (Messaufbau: 2 vCPU, 4 GiB, 10.000 Exchanges, Bodies bis 64 KiB, 20 Requests/s, 10 Minuten nach Warm-up; Mockserver mit 10.000 Entities) · `SPEC-026` (Gesamtspan, Messpunkt) · `SPEC-028` (Match-Span) · `SPEC-029` (Query-Span) · `SPEC-004` (Match-Key-Index `(session_id, match_key)` ohne Tabellenscan) · spezifikation.md §6 (Abnahmeparameter: „der Messaufbau muss reproduzierbar im Test-Harness hinterlegt sein“) · `ARC-012`.
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

**Ziel:** Ein reproduzierbarer Lasttest nach `LH-QA-01.a` fährt das gebaute Binary mit 2 vCPU und 4 GiB gegen 10.000 aufgezeichnete Exchanges (Replay) und gegen einen Workspace aus 10.000 Entities (Mock), mit 20 Requests/s über 10 Minuten nach Warm-up, und gibt p95 des Gesamtspans `odata.gateway.request` für Replay und Mock aus, mit dem Anteil von `odata.replay.match` bzw. `odata.mock.query`; die Zahlen stehen in der Closure-Notiz samt Last und Messendem des Laufs. Er prüft gegen den Zielwert des Lastenhefts (vorläufig p95 ≤ 100 ms) und berichtet; ein Durchfallen ist ein Carveout mit Folge-Slice, kein stilles Rot.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Bestätigung des Zielwerts von `LH-QA-01` — Entscheidung des fachlichen Owners, nicht des Repos (Teil von M7, [welle-betrieb](../welle-betrieb.md) §3 und §6); der Slice misst und berichtet gegen den Wert, den der User vor dem Start nennt (siehe §4), sonst gegen den vorläufigen Wert des Lastenhefts.
- Tuning bei verfehltem Zielwert (Indizes, Puffer) — eigener Slice nach der Messung; fällt der Wert durch, ist es ein Carveout mit Folge-Slice (`harness/conventions.md`, Carveout-Disziplin), kein stilles Rot.
- Latenz im Record- und im Pass-through-Betrieb — `LH-QA-01` sagt dort keine Latenz zu (der Upstream bestimmt sie); ebenso der Hybrid-Fallback auf den Upstream.
- Der Einbau der Spans — `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock`; dieser Slice liest sie.
- Eine Admin-Schnittstelle zum Aufbau der Last (Import, Workspace) — `welle-admin-api`; die Fixtures entstehen im Treiber über die Persistenz-Adapter.

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

- [ ] Messaufbau und Fixtures: 10.000 Exchanges mit Bodies bis 64 KiB für Replay, ein Workspace aus 10.000 Entities über mehrere EntitySets für Mock (Mischung aus Einzel-Lookups und Collection-Abfragen mit `$filter` und `$top`), erzeugt deterministisch aus einem Startwert; der Match-Key-Index wird ohne Tabellenscan genutzt (`LH-QA-01.a`, `SPEC-004`).
- [ ] Lauf: das gebaute Binary mit 2 vCPU und 4 GiB, 20 Requests/s, 10 Minuten nach Warm-up; Ausgabe von p95 des Gesamtspans für Replay und Mock samt Anteil des Match- bzw. Query-Spans, Zahl der Requests, Fehlerrate, Last des Hosts und Messendem; zwei Läufe in Folge liegen in der Schwankungsbreite, die der Architect benennt (`LH-QA-01`, Messmethode).
- [ ] Bericht und Entscheidung: die Zahlen stehen in der Closure-Notiz; Vergleich gegen den Zielwert; bei Überschreitung ein Carveout nach `docs/plan/carveouts/` mit Frist, Auflösungs-Trigger und Folge-Slice (`harness/conventions.md`), nie ein stilles Rot.
- [ ] `make gates` grün (der Lasttest selbst in der Gate-Kette oder als Werkzeug außerhalb nach Entscheidung des Architect, Auftrag 1).
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`harness/README.md` §Sensors).
- [ ] Führt der Slice ein sichtbares Verhalten des Binary ein (Schlüssel, Code, Kommando): Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Lasttest ist ein Werkzeug außerhalb des Binary).
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
| Werkzeug des Lasttests (Ort: Auftrag 1, `tools/…`) | neu | Erzeugung der Fixtures, Lastgenerator, Auswertung des Gesamtspans |
| `Dockerfile`, `harness/mk/` (neues Fragment), `Makefile` | update, neu | Stufe und Ziel, 2 vCPU und 4 GiB als Grenzen des Laufs, netzlos und ohne Mount (`AGENTS.md` §3.1) |
| `harness/vertraege/` | neu | Vertrag des Lasttests (Zusagen, Randformen, Schwankungsbreite) ([ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)) |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` | update | Fälle der Gegenprobe zum Werkzeug ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); der Stempel erst im Integrationsschritt |
| `harness/README.md` | update | Zeile des Ziels in §Sensors, *Nicht behauptet* (der Lasttest zu `LH-QA-01`) |
| `docs/plan/carveouts/` | neu, falls der Wert durchfällt | Carveout mit Frist und Folge-Slice |
| Tests | neu | Fälle der Zusagen des Werkzeugs, Mengen als Schleifen |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Aufträge an den Architect:**

1. **Gate oder Werkzeug.** Ein Lauf von 10 Minuten plus Warm-up gehört nicht in die Gate-Kette; zu entscheiden, mit Grund und Folge: Werkzeug außerhalb der Kette nach dem Muster von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, 1×), ein Gate mit kurzem Rauchlauf und der volle Lauf als Werkzeug, oder ein anderer Schnitt; und was die Gate-Kette dennoch prüft (der Treiber ist gebaut, die Auswertung rechnet richtig: p95 aus bekannten Werten, Fänger der Rechnung).
2. **Messpunkt und Auswertung.** Der Messpunkt ist der Gesamtspan `odata.gateway.request` (`SPEC-026`); wie das Werkzeug ihn liest (Abruf im Prozess nach Entscheidung von `slice-telemetry-port`, Auftrag 1 dort, oder Log), wie p95 berechnet wird (Verfahren, Auflösung, Behandlung von Fehlerantworten und Zeitüberschreitungen), und was „nach Warm-up“ ist (Zahl der Requests oder Dauer), und dass die Messung die Last nicht verändert.
3. **Reproduzierbarkeit.** 2 vCPU und 4 GiB als Grenzen des Containers (`--cpus=2 --memory=4g`), Startwert der Fixtures, feste Reihenfolge, Zahl der Läufe; die Schwankung mit Last und Messendem (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 2×): jede Zahl trägt die Last des Hosts während des Laufs, und ein Lauf unter fremder Last gilt nicht als Messung. Ob der Lauf im Arbeitscontainer von `welle-dev-container` liegt, wenn er bestanden hat (Reihenfolge-Vorsicht, siehe §4).
4. **Last und Anfragen.** Verteilung der Requests (Mischung aus Einzel-Lookups und Collections mit `$filter` und `$top` für Mock; welche Requests für Replay, welche Strategie), offene oder geschlossene Last, Ströme (Bodies bis 64 KiB), Verbindungen.
5. **Fixture-Aufbau.** Wie 10.000 Exchanges und 10.000 Entities entstehen, ohne einen Weg über HTTP, den es nicht gibt: Aufbau über die Adapter der Persistenz im Treiber, als Daten außerhalb des Binary; Größe des Datenverzeichnisses, Löschen danach.
6. **Zielwert.** Wie der Slice den Zielwert führt: Entscheidung des Users liegt vor (siehe §4) oder der vorläufige Wert; das Werkzeug trägt den Wert als Parameter und gibt ihn mit aus, und ein Lauf gegen einen anderen Wert wird nicht als Abnahme gelesen.
7. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** für die Marke des Planners (Treiber und Auswertung sind Code, der nicht nur aus reinen Funktionen besteht, `AGENTS.md` §3.14).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 7) | — | — |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Der Lauf des Lasttests selbst trägt `--cpus=2 --memory=4g` nach `LH-QA-01.a` und liegt ebenfalls unter `flock /tmp/odata-recorder-heavy.lock`, damit kein anderer schwerer Lauf die Messung stört; die Rolle meldet die Last des Hosts mit der Zahl.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: die Mischung der Anfragen, die Fixtures der Menge), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Eine Zahl aus einem Lauf nennt Zahl der Prozesse, Last und Messenden; „stabil“ oder „schwankt wenig“ steht nur mit Zahl (`AGENTS.md` §3.10).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock` liegen in `done/`; die Entscheidung des Users zum Zielwert von `LH-QA-01` liegt vor (siehe unten). Der Vertrag des Werkzeugs samt Randformen liegt in `harness/vertraege/`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-einbau-gateway`: der Gesamtspan `odata.gateway.request` im HTTP-Adapter ist der Messpunkt nach `LH-QA-01.a`, `odata.replay.match` weist den Anteil der Auswahl aus; mit ihm kommen mittelbar `slice-http-adapter`, `slice-replay-service` und `slice-bootstrap-m2` (M2), ohne die es das Binary im Modus `replay` nicht gibt.
- `slice-telemetry-einbau-mock`: `odata.mock.query` weist den Anteil der Query-Ausführung aus; mit ihm kommen mittelbar `welle-mock` und `slice-bootstrap-m3` (M3), ohne die es den Modus `mock` nicht gibt.
- **Entscheidung des Users zum Zielwert** (Owner von `LH-QA-01`; [Bericht](../../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §10 Nr. 8, Adresse: User, Frist vor dem Start dieses Slice): sie nennt p95 ≤ 100 ms als bestätigt oder einen anderen Wert, oder sie beschließt ausdrücklich, gegen den vorläufigen Wert zu messen. Beobachtbar ist sie als Ausgang an dieser Stelle des Berichts. Ohne sie startet der Slice nicht, weil das Werkzeug den Wert als Parameter trägt (Auftrag 6) und ein späterer anderer Wert die Zahlen nicht entwertet, aber die Abnahme-Aussage ändert.

**Abhängig von `welle-mock` / M3 und von M2, weil** der Lasttest beide Pfade (Replay und Mock) am Gesamtspan misst und der Lasttest des Mockservers den Modus `mock` braucht; er ist der Grund des Start-Triggers der Welle ([welle-betrieb](../welle-betrieb.md) §2). Ein Teilstart nur für Replay (nach `slice-telemetry-einbau-gateway`, ohne Mock) ist möglich, ergibt aber nicht die DoD und wird hier nicht geplant; schneidet der User so, geht der Mock-Teil an einen Folge-Slice und der Planner passt diese Datei an. **Nicht Bedingung:** `slice-container-image`, `slice-listener-tls` und `slice-audit-log` (der Audit-Eintrag der Löschung liegt nicht im Messpfad). Reihenfolge-Vorsicht, kein Blocker: ein Lauf von 10 Minuten bei 20 Requests/s schwankt mit der Last des Hosts (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 2×); liegt `welle-dev-container` in `done/`, ist der Lauf dort vor fremder Last geschützt, sonst gilt der `flock` der Betriebsregeln und der Bericht nennt die Last.

**Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` (`SPEC-128`), sobald die Tabelle *Zusage · Fänger* vorliegt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht ist dann Fixture-Aufbau gegen Lastgenerator und Auswertung.
- `in-progress` → `open` (blockiert): die Spans sind außerhalb des Prozesses nicht lesbar (Abrufweg von `slice-telemetry-port` ungeklärt), oder die Entscheidung zum Zielwert fehlt (siehe §6).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor und enthält die Zahlen des Laufs (p95 für Replay und Mock, Last, Messender).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Zielwert von `LH-QA-01` ist vorläufig und nicht freigegeben — **Ausgang:** offen; Adresse: User (Owner), Frist vor dem Start dieses Slice, sicher vor M7 ([Bericht](../../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §10 Nr. 8).
- Jede Zahl p95 hängt an Last und Messendem; zwei Läufe unter verschiedener Last sind keine Wiederholung — **Ausgang:** offen; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×): jede Zahl trägt Last, Zahl der Läufe und Messenden (Auftrag 3).
- Ein Lauf von mehr als 10 Minuten reißt die Frist einer Gate-Kette oder Gegenprobe — **Ausgang:** offen; `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×); Adresse: Architect (Auftrag 1).
- Der Abrufweg der Spans ist nicht spezifiziert ([welle-betrieb](../welle-betrieb.md) §6); ohne lesbaren Gesamtspan misst das Werkzeug an der Gegenseite (Client) und nicht im Prozess — **Ausgang:** offen; Adresse: Architect von `slice-telemetry-port` (Auftrag 1 dort), vor dem Start dieses Slice.
- Das Durchfallen des Zielwerts — **Ausgang:** offen; Carveout mit Frist, Auflösungs-Trigger und Folge-Slice (Tuning), nie ein stilles Rot ([welle-betrieb](../welle-betrieb.md) §6).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in `harness/vertraege/` oder `gilt nicht` mit Grund):

| Operation | fehlender Wert (Zielwert, Startwert, Binary) | leerer Wert (leere Fixture, keine Antwort) | unbekannter oder fremder Wert (Fehlerantwort, Zeitüberschreitung, Zielwert 0) | Abbruch (Lauf vorzeitig beendet) | Wiederholung (zweiter Lauf, andere Last) | Betriebsart (Replay, Mock) |
|---|---|---|---|---|---|---|
| Fixtures erzeugen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Last fahren (20 Requests/s, Warm-up, 10 Minuten) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| p95 aus dem Gesamtspan berechnen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Gegen den Zielwert vergleichen und ausgeben | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Treffer: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, jede Zahl p95 hängt an Last und Messendem und trägt beides) und `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, lange Läufe in der Gate-Kette; Entscheidung des Architect: Gate oder Werkzeug, Muster von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für Testhilfen, die `ctx` und Ströme des Lastgenerators umhüllen; offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
