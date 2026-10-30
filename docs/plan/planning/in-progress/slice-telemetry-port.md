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

**Verantwortlich:** pt9912 (Zweig slice-telemetry-port)

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

- [ ] Port `Telemetry` und Fake: Log-Eintrag mit den Feldern des Katalogs, Span mit Namen und Pflicht-Attributen nach `SPEC-026` bis `SPEC-030`, Zähler je Fehlercode und die weiteren Metriken, die der Architect in der Spezifikation festlegt (Auftrag 3); der Fake hält jedes Argument fest und zählt jede Operation (`LH-QA-05`, `LH-FA-01`).
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
| `.a-check.yml` | update, falls nötig | der neue Adapter-Pfad unter den Regeln des Hexagons (`ARC-020`); die Ebene `adapters-driven` hat die Kante zu `ports-support` bereits, ein Eintrag ist nur nötig, wenn der Architect die Lage des Adapters anders schneidet |
| `spec/architecture.md` | update, falls nötig | Telemetry-Zeilen in `ARC-020`, `ARC-021` und der Komponenten-Tabelle, nur wenn der Abrufweg oder die Form des Ports sie ändert; frei von Slices, Wellen und ADR-Bezügen (`AGENTS.md` §3.4) |
| `docs/plan/adr/`, `go.mod`, `go.sum` | nur falls der Architect eine Bibliothek wählt | eine ADR (Nummer aus dem Bereich der Linie D, siehe §4) `Accepted` vor dem ersten Code-Commit; sonst bleiben `go.mod` und `go.sum` unverändert |
| Handbuch (Ort nach `slice-handbuch-geruest`) | nur falls der Slice etwas am Binary sichtbar macht | Kapitel im selben Slice, siehe DoD und §4 |
| Tests | neu | je Zusage ein Fall; Canary-Test; nebenläufige Tests laufen unter `make test-race` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor dem ersten Code-Commit** (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Der Slice liegt seit dem Start in `in-progress/`; der Architect arbeitet im selben Zweig (Übergangsregel 3 in [`README.md`](../README.md) §Parallelbetrieb), der Planner setzt die Marke danach (§4).

**Kennungsbereiche der Linie D** (Setzung des Auftrags vom 2026-10-30; der Integrationsschritt nimmt sie in die Übergangsregeln, `README.md` §Parallelbetrieb Nr. 6, auf): `SPEC-230` bis `SPEC-259`, `ARC-040` bis `ARC-049`, ADR `0040` bis `0049`, `CO-060` bis `CO-079`. Jede Kennung dieses Slice kommt aus diesem Bereich; ein erschöpfter Bereich geht als Frage an den Planner. `MR-*` vergibt dieser Zweig nicht, `LH-*` ändert er nicht. Die bestehenden Kennungen `SPEC-026` bis `SPEC-030`, `SPEC-005`, `SPEC-121` und `SPEC-110` werden in ihren Zeilen geschärft, nicht neu vergeben.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft):

Der Planner hat die Aufträge beim Start gegen den Ist-Zustand des Repos geschärft (gelesen: `src/hexagon/ports/support` mit `Clock` und `IDGenerator`, `supporttest`, die Tabelle der Fehlercodes in `src/hexagon/model/fehler.go`, `go.mod`, `.a-check.yml`, `spec/architecture.md`, die `*slog.Logger` in `src/hexagon/services/{record,session,maintenance}`). Vorbefund: kein Teil des Produkts erzeugt heute einen `*slog.Logger` — `src/bootstrap` baut keinen; die Dienste `record`, `session` und `maintenance` nehmen ihn als Feld. Die Entscheidungen stehen vor dem ersten Code-Commit in `spec/spezifikation.md` (Produkt, Kennungen aus dem Bereich der Linie D) und nicht im Plan:

1. **Abrufweg und Senke.** Die Spezifikation nennt keinen Abrufweg für Metriken und Spans ([welle-betrieb](../welle-betrieb.md) §6). Zu entscheiden, mit Grund und Folge, und zwar gegen drei Leser, die der Plan benennt: (a) Tests dieses Slice und der beiden Einbau-Slices (Lesen im Prozess); (b) `slice-lasttest`, dessen DoD das **gebaute Binary** als eigenen Prozess misst und den Gesamtspan `odata.gateway.request` liest — ein Weg, der nur im Prozess lesbar ist, erreicht ihn nicht, es sei denn, der Architect benennt, wie der Lasttest im Prozess läuft; (c) die spätere Diagnose (`/__admin/diagnostics`, `SPEC-005`): `InspectionUseCase` ist ein Dienst und erreicht den Zähler nur über einen **Lese-Teil des Ports** (Dienste dürfen `ports-support` nutzen, nie den Adapter), nicht über eine Funktion des Adapters. Zu entscheiden: ob Spans und Zähler nur im Prozess lesbar sind oder zusätzlich ausgegeben werden (eine Log-Zeile je beendetem Span mit Dauer ist der Weg, der keine neue Schnittstelle des Produkts braucht); ob der Lese-Teil im Port `Telemetry` oder in einem zweiten Port liegt. Ein Endpunkt, ein Netz-Port oder ein Format, das ein Anwender außerhalb des Prozesses sieht, ist eine neue Schnittstelle des Produkts und braucht eine Anforderung im Lastenheft (Sache des Owners; Lastenheft §5: kein verpflichteter externer Dienst): verlangt der gewählte Weg eine solche, meldet der Architect das als Frage an den Planner (Adresse und Ausgang in §6) und nimmt für diesen Slice den Weg, der sie nicht braucht. Dazu die Bibliotheksfrage: genügen `log/slog`, `sync`, `sync/atomic` und `time` der Standardbibliothek? Eine Bibliothek für Spans und Metriken ist eine neue Abhängigkeit (die Abhängigkeiten in `go.mod` sind heute `modernc.org/sqlite` und seine Last); sie braucht eine ADR `Accepted` vor dem Code mit `go mod verify` und netzlosen Läufen (`AGENTS.md` §3.1). Der Planner erwartet die Standardbibliothek; wählt der Architect anders, nennt er den Grund und was die Standardbibliothek nicht kann.
2. **Form des Ports.** Operationen (Log, Span beginnen und beenden mit Attributen, Zähler erhöhen), Weitergabe über `ctx` (der Span-Kontext reist im `ctx` des Aufrufs; welche Operation gibt einen neuen `ctx` zurück), Eltern-Kind-Beziehung der Spans (`odata.replay.match` und `odata.mock.query` „zerlegen“ den Gesamtspan, `LH-QA-01.a`: der Messpunkt ist der Gesamtspan, die Kinder weisen Anteile aus), und die Aufrufstellen als Menge für den Fake (`AGENTS.md` §3.15). Die Span-Namen und ihre Pflicht-Attribute sind Daten des Ports: sind sie eine **geschlossene** Menge (ein Name außerhalb von `SPEC-026` bis `SPEC-030` ist eine Randform) oder offen? Der Hybrid-Fallback (Spans und Zähler, [welle-hybrid](../welle-hybrid.md) §6) hat keinen Nehmer in `welle-betrieb`; der Port darf ihn nicht ausschließen und nicht vorwegnehmen. Die Werte von `recorder.mutation_type` und `recorder.match_strategy` legen die Einbau-Slices fest (`slice-telemetry-einbau-mock` Auftrag 1): der Port prüft das Vorhandensein und die Art des Pflicht-Attributs, nicht eine Wertemenge, die ihm ein anderer Slice erst liefert — oder der Architect legt die Menge hier fest, und die Einbau-Slices erben sie.
3. **Metriken.** §5 der Spezifikation nennt nur Spans; Namen, Art (Zähler, Verteilung), Einheit und Schlüssel der Metriken, darunter der Zähler je Fehlercode (`SPEC-005`) und die Dauer, stehen vor dem Code in der Spezifikation. Die Menge der Fehlercodes ist geschlossen (Tabelle in `src/hexagon/model/fehler.go`, `CodeE9001` als letzter), die Kardinalität damit beschränkt; ein unbekannter Code (Wert außerhalb der Tabelle, `0`) ist eine Randform, und die Zahl der Elemente der Menge liefert die Tabelle, nicht der Plan (Zählprüfung des Generators, Auftrag 8). Die Verteilung (Dauer) braucht eine Obergrenze ihrer Eimer oder ihres Speichers.
4. **Logfelder.** Welche Felder je Ereignis Pflicht sind, die Stufen, das Format (JSON-Zeilen, Feldnamen fest, Reihenfolge), und die Abbildung des festen Eintrags des Record-Dienstes (`SPEC-121` Nr. 12, Attribute `mode`, `ereignis`, `error_code`, `service_id`, `session_id`, `exchange_id`, `correlation_id`) auf den Port, so dass der Einbau den Wortlaut nicht ändern muss. Die Dienste `session` und `maintenance` loggen ebenfalls über einen `*slog.Logger` (`src/hexagon/services/session/protokoll.go`, `src/hexagon/services/maintenance/protokoll.go`) und kein Einbau-Slice stellt sie um: der Architect entscheidet, ob der Adapter einen `slog.Handler` bereitstellt, mit dem alle drei Dienste im Format des Katalogs loggen, ohne dass ihr Code sich ändert, und nennt, wer die Verdrahtung trägt (kein Slice nimmt sie heute an, siehe §6).
5. **Ausschluss sensibler Inhalte.** Wer erzwingt „keine Credentials, keine Bodies, keine Query-Literale“ (`LH-QA-04`, `SPEC-026` ff. Schlusssatz): der Typ der Attribute (kein beliebiger `any`), eine Liste erlaubter Schlüssel, eine Prüfung am Wert; und der Fänger. Die Allaussage „kein Credential und kein Body erreicht ein Log“ hat eine **Menge**: jede Logstelle, jedes Attribut, jeder Ausgabeweg (Log-Zeile, Span-Attribut, Metrik-Schlüssel, Fehlertext der Telemetrie selbst). Der Fänger ist eine Schleife über alle drei, mit einem Canary je Einbringungsweg (Header, Body, Query-Literal, Wert einer Konfiguration), der Zählprüfung der Wege und dem Orakel „das Canary-Secret steht in keiner Ausgabe“ — kein Satz handgewählter Fälle (`AGENTS.md` §3.15, §3.17).
6. **Fehlerabbildung und Nebenläufigkeit.** Das Ergebnis des Aufrufers hängt nie vom Gelingen der Telemetrie ab; was bei Ausfall der Senke geschieht (Schreibfehler des Log-Writers, voller Puffer), was bei einem beendeten `ctx` (der Span wird trotzdem beendet?), bei einem Span, der nie beendet wird, bei doppeltem Beenden und beim Beenden eines Kindes nach seinem Eltern-Span; Obergrenze des Speichers (Zahl der gehaltenen Spans, Länge eines Attributs). Die Zähler und die Sperre sind nebenläufig (Formel: Faktor 85): welche Zusagen über Nebenläufigkeit der Vertrag macht (Zähler verliert keinen Zuwachs, Lesen ist ein Stand zu einem Zeitpunkt, Reihenfolge der Log-Zeilen), und für jede den Wirkungs-Fänger oder das Negativ nach `AGENTS.md` §3.10: das Detektor-Negativ der Register-Einträge `zweig-ohne-lauf-unter-dem-detektor` (Ausfall der Senke und Abbruch führt kein Lauf aus, wenn nicht ein Fake sie erzwingt) und `rate-ohne-detektor-schwankt-mit-der-last` (jede Wirkung, die der Detektor nicht meldet, trägt ihre Rate mit Zahl der Läufe).
7. **Katalog des Binary und Handbuch.** Führt der Slice einen Schlüssel (Log-Stufe, Format) oder Code ein, ist es ein Nachzug von `SPEC-110` und ein Eintrag im Handbuch; sonst bleibt das Binary unverändert. Der Planner erwartet Letzteres: der Port wird erst vom Einbau gerufen, kein Bootstrap verdrahtet ihn (siehe §4). Entscheidet der Architect anders (ein Konfigurationsschlüssel für Stufe oder Format ist der naheliegende Fall), wartet der Slice nach §4 auf die beiden Handbuch-Slices — das ist eine Rückführung, vorab benannt, und der Architect sagt es vor dem Code.
8. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** für die Marke des Planners (Metrik und Formel: `AGENTS.md` §3.14, `SPEC-128` Nr. 1). Die Tabelle hat die fünf Köpfe `Nr.`, `Zusage`, `Fänger`, `Mutation`, `Klasse` (`SPEC-128` Nr. 3 a; die Klasse ist `nebenläufig` oder `andere`, ein akzeptiertes Negativ zählt `andere`); der Architect schreibt den Block `groessenmarke` mit den Teilen, die ihm gehören: `pfad` (mindestens `src/hexagon/ports/support`, `src/adapters/driven/telemetry`; weitere nur, wenn der Slice sie ändert; weder `spec/` noch `docs/`), `produktcode` (erwartete Zeilen des Produktcodes in der Metrik des Plans, Tests ausgenommen) und `fakes` (Port-Fakes und Test-Server, Zahl der Zeilen; hier mindestens der Fake in `supporttest` und ein Fake der Senke). `basis`, `schaetzung` und `marke` setzt der Planner danach (§4). Jede Zeile nennt Menge und Orakel, wo eine Klasse generierbar ist.
9. **Mengen für die Fänger** (`AGENTS.md` §3.15, §3.17), vom Architect aus dem Wortlaut des Vertrags gebildet, mit Zahl der Elemente: alle Felder des Logkatalogs (zehn), alle Span-Namen mit allen Pflicht-Attributen (fünf Spans, elf Pflicht-Attribute — gezählt aus `SPEC-026` bis `SPEC-030`), alle Fehlercodes der Tabelle, alle Ausgabewege, alle Operationen des Ports (jede Operation des Fakes zählt, auch eine, die der Dienst nie rufen darf), beide Seiten jeder Grenze (Obergrenze des Speichers, Länge, Gleichheit zweier Werte).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| — | (Architect, Auftrag 8; Fänger ist ein Test, ein Fall der Gegenprobe oder eine Ausgabe des Gates) | — | — | — |

**Betriebsregeln für alle Rollen:**

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts (Entscheidung des Users vom 2026-10-21).
- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4` (höchstens zwei Container je Linie zugleich) unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Die Image-Falle: das Image `odata-recorder:race` trägt `/src/src` eingebacken; ein Lauf, der Quellen hineinkopiert, leert den Pfad zuvor oder nutzt die Stufe `deps`. Ein Rot unter Last wird auf ruhigem Host wiederholt, bevor es als Befund gilt.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python (Register-Eintrag `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` steht bei 1× in der Basis dieses Zweigs, laut Auftrag des Users im Hauptbaum bei 3×: eine weitere Verletzung ist ein Befund des Reviewers); keine Hilfsdateien im Wurzelverzeichnis. Commit-Nachrichten gehen als Datei (`git commit -F`), nicht als Heredoc.
- Eine Hülle um ein Argument (ein Fake, der `ctx`, einen Strom oder einen Writer umhüllt, ein Test-Writer der Senke) darf keine Eigenschaft des Originals verdecken (Abbruch, Fehler, Blockieren, Zahl der Aufrufe; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`): der Fake hält jedes Argument fest, zählt jede Operation, und eine Panik des Fakes färbt den benannten Test, nicht einen anderen.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jedes Feld des Katalogs, jeder Span-Name und jedes Pflicht-Attribut von `SPEC-026` bis `SPEC-030`, jeder Fehlercode), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt, nicht im Slice-Commit.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13); der Bericht allein ist keine Adresse.
- Der Implementer hält an einer Randform ohne Ort in der Spezifikation an und gibt sie an den Architect zurück (`AGENTS.md` §3.12); er entscheidet sie nicht im Code.
- Wert mit Escape-Folge oder Rohbytes: der Plan und die Berichte schreiben eine solche Folge nicht als Zeichen, sondern beschrieben (Werkzeug-Falle Backslash-u, `AGENTS.md` §3.18).

**Rollenfolge** (`.claude/agents/`, kein Self-Review): Architect (Vertrag, Tabelle, Raster, Block `groessenmarke`, ADR falls Bibliothek) → Planner (Marke, §4) → Implementer (Code, Mutationen selbst rot, Absatz zu Zusagen und Raten in §3, ein Handoff-Commit) → Reviewer (Report unter `docs/reviews/`) → Fix-Runden (Menge je Befund, `AGENTS.md` §3.17) → Verifier → Planner (Closure mit Lerneintrag; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorlage an den User) → Integrationsschritt (Merge, Register, Roadmap, Stempel, `make gates` im Hauptbaum).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`, 2026-10-30, Linie D): kein anderer Slice muss in `done/` liegen. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Entscheidung zum Trigger der Welle (Planner, Auftrag des Users vom 2026-10-30, produktive Slices laufen parallel):** Der Start-Trigger von [welle-betrieb](../welle-betrieb.md) §2 (`welle-mock` in `done/`, M3 erreicht) wird für **diesen** Slice gelockert, die Welle bleibt unverändert. Grund: Code-Unabhängigkeit — der Slice braucht nur `src/hexagon/ports/support` (`slice-clock-idgenerator`) und die Fehlercode-Tabelle (`slice-fehlermodell-konfiguration`), beide in `done/`, und ruft weder `MockUseCase` noch `WorkspaceUseCase`; die Mock-Spans sind Namen als Daten (siehe unten). Der Trigger der Welle bindet den Einbau in Mock und Workspace und den Lasttest, nicht den Port. Die Lockerung gilt nur, solange der Slice **ohne eine Wirkung am Binary** bleibt: Die Handbuch-Pflicht der Welle ([welle-betrieb](../welle-betrieb.md) §4) greift genau dann, wenn der Slice einen neuen Code, Schlüssel oder ein Kommando in den Katalog des Binary bringt oder Logfelder am Binary sichtbar macht; dann wartet er auf `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` in `done/` (Rückführung `in-progress` → `open`, unten), der Architect sagt das vor dem Code (Auftrag 7). Der Plan ändert `src/bootstrap` nicht und berührt damit `slice-bootstrap-m2` nicht. Die Closure-Trigger der Welle (§3 dort) bleiben; die Roadmap und `welle-betrieb.md` schreibt der Integrationsschritt nach (§6).

**Abhängigkeiten — Grund je Eintrag:**

- Code, den der Slice braucht und der schon in `done/` liegt: `src/hexagon/ports/support` und `supporttest` (`slice-clock-idgenerator`), die Fehlercode-Tabelle in `src/hexagon/model` (`slice-fehlermodell-konfiguration`), der Katalog und die Tabelle der Spans in der Spezifikation. Mehr braucht ein Port mit Adapter nicht.
- Handbuch-Punkt der DoD (nur wenn er greift, siehe DoD): `slice-handbuch-geruest` (Ort der Kapitel) und `slice-handbuch-meldungskatalog` (das Kataloge-Gate) aus `welle-benutzerhandbuch` liegen dann in `done/`; ohne sichtbare Änderung am Binary entfällt die Bedingung.

**Unabhängig von `welle-mock` / M3, von M2 und von Linie A, weil** der Port die Namen der Spans als Daten kennt und nichts ruft: `MockUseCase`, `WorkspaceUseCase`-Mutation, HTTP-Adapter und Replay-Dienst sind Orte des Einbaus (`slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-mock`), nicht des Ports. Der Start-Trigger der Welle (`welle-mock` in `done/`, [welle-betrieb](../welle-betrieb.md) §2) gilt für den Abschluss mit Mock-Spans und Lasttest; dieser Slice hängt an keinem Code daraus. Reihenfolge-Vorsicht, kein Blocker: der Slice ändert `src/bootstrap` nicht (die Verdrahtung ist Einbau), so dass er die Dateien der Slices von M2 (`slice-bootstrap-m2`) nicht berührt.

**Größenmarke:** noch nicht gesetzt, kein Code. Der Planner setzt sie vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` (`SPEC-128`: Schätzung *S* = 85·*a* + 80·*b* + 200·*f* + *p*, Marke *M* = 100·⌈135·*S*/10000⌉), sobald die Tabelle *Zusage · Fänger* (fünf Köpfe) und der Block `groessenmarke` mit `pfad`, `produktcode` und `fakes` des Architect vorliegen (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Der Port hat nebenläufige Zeilen (Zähler, Sperre, Rate): die Klasse `nebenläufig` setzt der Architect je Zeile (`SPEC-128` Nr. 3 e; Nr. 10 b: eine zu niedrige Klasse senkt die Schätzung, das Programm sieht es nicht, der Reviewer löst die Tabelle gegen den Vertrag auf). Weder die Schätzung noch die Marke stehen im Plan, bevor der Architect fertig ist; der Planner schreibt sie dann mit `basis`, `schaetzung` und `marke` in den Block.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach dem Architect setzt (siehe oben); ergibt die Formel von `SPEC-128` die Schnittfrage, wird vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht ist dann Logs und Spans einerseits, Metriken und Zähler andererseits.
- `in-progress` → `open` (blockiert): der Abrufweg (Auftrag 1) ist nicht entschieden, die Wahl einer Bibliothek braucht eine ADR, die nicht `Accepted` ist, oder der Slice bringt einen Code, Schlüssel oder ein Kommando in den Katalog des Binary und `slice-handbuch-geruest` oder `slice-handbuch-meldungskatalog` liegen nicht in `done/` (Lockerung des Triggers, siehe oben).

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

- Abrufweg der Metriken und Spans ist nicht spezifiziert; ohne Entscheidung ist der Lasttest (`slice-lasttest`, DoD: das gebaute Binary als eigener Prozess) nicht lesbar, und `InspectionUseCase` erreicht den Zähler nur über einen Lese-Teil des Ports — **Ausgang:** offen; Adresse: Architect (Auftrag 1), vor dem ersten Code-Commit (Rückführung `in-progress` → `open`, §4).
- Eine Bibliothek für Spans und Metriken ist eine neue Abhängigkeit (Supply-Chain, netzlose Läufe) — **Ausgang:** offen; Adresse: Architect, ADR (Bereich 0040 bis 0049) vor dem Code, falls sie gewählt wird.
- Spezifikationslücke: §5 nennt Spans, keine Metriken; das Lastenheft führt keine Anforderung an Metriken oder Spans außer der Messmethode von `LH-QA-01` — **Ausgang:** offen; Adresse: Architect (Auftrag 3), Meldung an den Planner, falls eine Anforderung im Lastenheft fehlt.
- Sendung an einen Nehmer ohne Datei: der Zähler je Fehlercode ist der Gegenstand von `/__admin/diagnostics` in `welle-admin-api` (Slice `slice-inspection-use-case` der Tabelle dort); dessen Datei besteht nicht (`open/` trägt keinen Plan von `welle-admin-api`, geprüft beim Start) — **Ausgang:** offen; Adresse: Planner beim Anlegen der Slices von `welle-admin-api` (`AGENTS.md` §3.13: erst der Nehmer führt den Punkt, dann nennt der Geber ihn); der Lese-Teil des Ports (Auftrag 1) ist die Zusage, die der Nehmer erbt.
- Sendung ohne Nehmer: die Dienste `session` und `maintenance` loggen über `*slog.Logger` (Felder `ereignis`, `error_code`, `service_id`, `session_id`), und kein Einbau-Slice stellt sie auf den Port oder den Handler des Adapters um (`slice-telemetry-einbau-gateway` nimmt nur den Record-Dienst an); ohne Umstellung gilt der Logkatalog (Logfelder am Ende von spezifikation.md §4) für sie nicht einheitlich — **Ausgang:** offen; Adresse: Planner (Nehmer anlegen oder `slice-telemetry-einbau-gateway` ergänzen, sobald der Architect Auftrag 4 entschieden hat; in `open/` schreibt dieser Zweig nichts).
- Sendung ohne Nehmer, vom Geber `welle-hybrid` ([welle-hybrid](../welle-hybrid.md) §6): Spans und Zähler des Hybrid-Fallbacks sind kein Gegenstand dieses Slice und keines Einbau-Slice (beide schließen sie aus); der Port soll sie nicht verbauen (Auftrag 2) — **Ausgang:** offen; Adresse: Planner (derselbe Punkt steht in §6 von `slice-telemetry-einbau-gateway`).
- Der Stand von `AGENTS.md` §3.14 (Schätzung × 1,25, Faktoren 85 und 40) weicht vom Vertrag `SPEC-128` (Faktoren 85, 80, 200, Zuschlag 135 vom Hundert) ab; die Marke dieses Slice folgt `SPEC-128` — **Ausgang:** offen; Adresse: User (Diff-Vorlage für `AGENTS.md` im Closure des Planners; Änderungen an `AGENTS.md` schreibt dieser Zweig nicht).
- Nachzüge für den Integrationsschritt (Roadmap, `welle-betrieb.md`, Übergangsregeln, Register): Linie D mit Kennungsbereichen, Start dieses Slice vor dem Welle-Trigger, Register-Zähler — **Ausgang:** offen; Adresse: Planner im Integrationsschritt.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Prüfung der Nehmer durch den Planner** (`AGENTS.md` §3.13, beim Start gelesen: §1 *Ausdrücklich NICHT* und DoD): (1) `slice-telemetry-einbau-gateway` nimmt den Einbau in `record` und `replay` mit der Kennung dieses Slice an (*Übernimmt*, DoD: Gesamtspan, Capture- und Match-Span, Protokolleintrag über den Port mit unverändertem Wortlaut); kein *Ausdrücklich NICHT* trifft ihn. (2) `slice-telemetry-einbau-mock` nimmt den Einbau in `mock` und Workspace-Mutation an; kein Ausschluss trifft ihn, die Werte von `recorder.mutation_type` liegen bei ihm (Auftrag 2). (3) `slice-audit-log` schließt das Diagnose-Log und die Zähler aus und nennt diesen Slice — Gegenstück stimmt. (4) `slice-lasttest` liest den Gesamtspan und nennt den Abrufweg als offenes Risiko mit Adresse Architect — dieser Slice liefert die Entscheidung (Auftrag 1), die Datei des Nehmers bleibt unverändert. (5) `slice-inspection-use-case` und die Slices `slice-admin-*` bestehen als Dateien nicht; sie sind keine Adresse (siehe oben). (6) Die Spans und Zähler des Hybrid-Fallbacks und der Umstieg der Dienste `session` und `maintenance` haben keinen Nehmer (siehe oben).

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

**Vorgelagert — offene Beobachtungen sichten:** Register beim Start gesichtet (2026-10-30, Zähler aus den Dateien unter `evidence/` der Basis `f4d9ac3`; neunzehn Einträge in `BEO-REPO`). Der Adapter ist nebenläufig und der Fake umhüllt `ctx`: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) gilt für Pfade, die kein Lauf ausführt (Ausfall der Senke, Abbruch), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für den Fake des Ports, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) für jede Wirkung, die der Detektor nicht meldet (Zähler unter Last: Rate mit Zahl der Läufe, `AGENTS.md` §3.10); offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1× in der Basis dieses Zweigs; laut Auftrag des Users im Hauptbaum und den anderen Zweigen 3×, dann ist die Schwelle erreicht und die Verkörperung Sache des nächsten Lese-Schritts des Integrationsschritts — dieser Zweig setzt keinen Zähler). Zusätzlich für die Fänger: `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, verkörpert in `implement-slice` Schritt 14), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (7×, `AGENTS.md` §3.10) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, `AGENTS.md` §3.16) für die Allaussagen „kein Credential in keinem Log“ und „der Zähler verliert nichts“. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16) und `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (16×, `AGENTS.md` §3.14). Von den offenen erreicht keiner mit diesem Slice 3×, außer dem Eintrag zu den Betriebsregeln nach Stand des Hauptbaums.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
