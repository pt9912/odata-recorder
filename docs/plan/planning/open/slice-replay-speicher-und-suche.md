# Slice slice-replay-speicher-und-suche: Replay unter `LH-QA-01.a` messbar machen — Ursache des Speicherverbrauchs messen, Suche nach dem Match-Key, erneuter Abnahmelauf

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md). Die Welle schließt nicht, solange kein p95 für Replay in der Größe `abnahme` vorliegt (§3 der Welle).

**Bezug:** `LH-QA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md); Carveout [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md).

**Berührte Spec-Stellen** (Stand bei der Anlage; der Architect ergänzt sie mit dem Vertrag): `SPEC-135` Nr. 8 (Kandidaten: `Exchanges.ListBySession` je Anfrage; verworfene Alternative `ListByMatchKey`), Nr. 9 (Auswahl je Strategie) und *Ungeprüft* (Aufwand je Request) · `SPEC-105` (Vertrag der Repositories) · `SPEC-216` Nr. 1, 12 d und 16 (der Lasttest, unverändert) · `LH-QA-01.a`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner, Closure von [slice-lasttest](../done/slice-lasttest.md)). **Datum:** 2026-11-26.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Unter dem Messaufbau von `LH-QA-01.a` (2 vCPU, 4 GiB, 10.000 Exchanges, 20 Requests/s) läuft der Replay-Dienst die Messung zu Ende und liefert ein p95; die Ursache des heutigen Speicherverbrauchs ist vorher gemessen und nicht vermutet.

**Ausgangslage** (Aufruf `abnahme` auf `e7bfebb4`, 2026-11-26, Verifier; Bericht `2026-11-26-slice-lasttest-verifikation`): `replay` ist in beiden Läufen `ungueltig` mit `grund=binary`; das Binary wurde je Lauf einmal vom OOM-Killer an der Grenze von 4 GiB beendet (Exit 137, `anon-rss` 4147816 kB und 4147740 kB); 13099 und 13109 von 13200 Anfragen sind Fehler, 102 und 92 Gesamtspans, bis zu 185 und 183 gleichzeitig offene Anfragen. Es gibt kein p95. **Die Ursache ist nicht belegt.** Verdacht, nicht Befund: `SPEC-135` Nr. 8 liest je Anfrage die ganze Session (`ListBySession`).

**Übernimmt:** `slice-lasttest` (liegt in `done/`): (a) den Produktbefund V-8 seiner Verifikation, (b) den aus seiner DoD genommenen Punkt *p95 des Gesamtspans für Replay in der Größe `abnahme` samt Anteil des Match-Spans*, (c) die Auflösung oder Verlängerung von [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Ein anderer Zielwert oder eine Änderung am Lastenheft — Entscheidung des Users vom 2026-11-24 ([welle-betrieb](../welle-betrieb.md) §3): eine Lockerung ist eine Änderung am Lastenheft und kein Mittel dieses Slice.
- Änderungen am Lasttest-Werkzeug (`tests/lasttest/`, `tools/harness/lasttest.sh`, `SPEC-216`) — es verhält sich wie zugesagt (Verifikation); der erneute Lauf misst mit demselben Werkzeug, damit die Läufe vergleichbar sind. Braucht die Ursachenmessung etwas daran, geht das an den Planner zurück.
- Die übrigen Aufrufer von `ListBySession` (Record, Session beenden, Inspektion, Import) — sie liegen nicht im Messpfad von `LH-QA-01`; ihr Speicherbedarf ist eigens benannt (`SPEC-124` *Ungeprüft*).
- Tuning der Latenz, wenn der erneute Lauf ein p95 über 100 ms liefert — eigener Slice nach diesem Ergebnis, unter [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md).
- Eine Begrenzung gleichzeitiger Anfragen oder eine Abweisung unter Last — sie änderte das sichtbare Verhalten des Binary. Zeigt die Messung, dass die Grenze ohne sie nicht zu halten ist, geht der Slice an den Planner zurück (§4).

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] **Ursache gemessen:** der Speicherbedarf einer Replay-Anfrage bei einer Session mit 10.000 Exchanges und sein Verlauf bei gleichzeitigen Anfragen stehen als Zahlen mit Methode, Last und Messendem im Plan; die Zahlen sagen, ob das Lesen der ganzen Session die Grenze von 4 GiB erklärt, und welcher Anteil auf anderes fällt. Vor dieser Messung entsteht keine Behebung.
- [ ] **Behebung:** die Replay-Suche liest nach dem Vertrag des Architect allein, was die Auswahl über den Match-Key und die Strategie braucht; ein Fänger zählt, dass Zahl der gelesenen Exchanges oder Speicher je Anfrage nicht mit der Größe der Session wächst; das Verhalten der vier Strategien und die Fehlercodes bleiben (`SPEC-135` Nr. 9), je Zusage eine Mutation (`AGENTS.md` §3.10).
- [ ] **Erneuter Abnahmelauf:** der Verifier fährt `make lasttest` im Profil `abnahme` mit den Vorgaben einmal auf einem Commit, der die Behebung trägt; p95 für Replay und Mock, Last, Messender und Commit stehen in der Closure-Notiz. Bei `abnahme=ja` wird [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) in das Verzeichnis `done/` der Carveouts bewegt und die Bindung in `harness/README.md` zurückgesetzt; sonst bleibt der Carveout aktiv und nennt den nächsten Folge-Slice.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-105`, `SPEC-135`).
- [ ] Führt der Slice ein sichtbares Verhalten des Binary ein: Handbuch-Kapitel im selben Slice nachgezogen; sonst sagt §7, dass kein Schlüssel, Code oder Kommando hinzukam.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen (im Wellen-Betrieb von der Closure der Welle).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort.

Pfad-Kandidaten, nicht entschieden: `src/hexagon/services/replay/`, `src/hexagon/ports/outbound/`, `src/adapters/driven/sqlite/`, die Fakes des Ports, `spec/spezifikation.md` (Architect). Den Index `(session_id, match_key)` führt das Schema schon (Fänger `make schema-check`, [slice-lasttest](../done/slice-lasttest.md) §3).

**Auftrag an den Architect** (vor dem Übergang `next` → `in-progress`; der Plan trägt bis dahin keine Tabelle, kein Raster, keinen Block der Größenmarke und keine Marke):

1. **Messung der Ursache:** womit, wo und an welchem Stand gemessen wird (netzlos, im Container, `AGENTS.md` §3.1), welche Größen die Ursache von den Alternativen trennen (Speicher je Anfrage gegen Zahl gleichzeitiger Anfragen) und ob die Messung Code braucht. Ob Messung und Behebung ein Slice bleiben oder zwei werden, gehört zur Antwort (§6).
2. **Vertrag der Suche:** Operationen des Ports (`SPEC-105`), die Neufassung von `SPEC-135` Nr. 8 samt der dort verworfenen Alternative, Reihenfolge, Fehlerabbildung je Aufrufstelle, `ctx`; je Strategie, was gelesen wird (`strict` braucht den Exchange mit `sequence_no` *p* + 1 unabhängig vom Match-Key, und `ODR-E5001` geht `ODR-E5002` voraus, `SPEC-135` Nr. 9).
3. **Zusage zum Speicher:** ob die Spezifikation für den Replay-Dienst eine Aussage zum Speicher je Anfrage und unter gleichzeitigen Anfragen führt; `SPEC-135` *Ungeprüft* nennt heute allein den Aufwand.
4. **Tabelle *Zusage · Fänger*, Raster *Operation × Argumentart*, Erwartung für den Produktcode und Zahl der Fakes** (`AGENTS.md` §3.12, §3.15, §3.16; `SPEC-128` Nr. 1); ob eine ADR nötig ist.

Danach setzt der Planner Block, Schätzung und Marke; über der Schnittgrenze wird vor dem Code geschnitten (`AGENTS.md` §3.14).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-lasttest` liegt in `done/` und `make lasttest` ist auf `main`; die Antworten des Architect aus §3 liegen vor; der Planner hat die Marke gesetzt.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `open` (blockiert): die Messung trägt den Verdacht nicht (das Lesen der Session erklärt den Verbrauch nicht), oder die Grenze ist nur mit einem anderen sichtbaren Verhalten des Binary zu halten. Dann schneidet der Planner mit dem Architect neu; [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) bleibt aktiv.
- `in-progress` → `next` (zu groß): der Umfang liegt vor dem Code über der Schnittgrenze von `SPEC-128` Nr. 1.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag.

Die DoD ist vollständig, `make gates` ist grün, der Aufruf `abnahme` liefert für Replay einen gültigen Lauf mit p95, und die Closure-Notiz mit Lerneintrag nennt die Zahlen und den Stand von [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang.

- Die Ursache ist nicht die Suche: der Lauf trennt Speicher je Anfrage nicht von der Zahl gleichzeitiger Anfragen, und das Kernel-Protokoll nennt allein den Prozess — **Ausgang:** offen; Adresse: DoD 1.
- Messung und Vertrag der Behebung hängen aneinander: ein Vertrag vor der Messung wäre geraten, Randformen müssen aber vor dem ersten Code-Commit entschieden sein — **Ausgang:** offen; Adresse: Architect (§3, Auftrag 1).
- Auch nach der Behebung liegt p95 über 100 ms, oder das Binary erreicht die Grenze aus einem zweiten Grund — **Ausgang:** offen; Folge: [CO-001](../../carveouts/CO-001-lasttest-replay-ohne-p95.md) bleibt aktiv mit neuem Folge-Slice.
- Der Aufruf `abnahme` hält `flock /tmp/odata-recorder-heavy.lock` rund 50 Minuten (gemessen: 49 Minuten, Aufbau 275,5 s, 69 MiB); die Wurzel von Docker (`/docker`) war beim letzten Lauf zu 96 % belegt (23 GiB frei), und fremde Container liefen — **Ausgang:** offen; Adresse: Verifier (Bedingungen im Bericht), Planner (Zeitfenster).
- Seit dem letzten Aufruf `abnahme` änderte sich das Binary (Admin-Listener auf `main`) — der erneute Lauf gilt für den neuen Stand und fährt auch `mock`, dessen Ergebnis `bestanden` er nicht voraussetzt — **Ausgang:** offen; Adresse: Verifier.
- Die Setzungen des Lasttests sind ungemessen; ihre Prüfung nach drei gültigen Läufen je Betriebsart (`SPEC-216` Nr. 16 g) hat bisher zwei Läufe `mock` und keinen `replay` — **Ausgang:** offen; Adresse: Architect nach dem dritten gültigen Lauf.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für
Steering-Loop-Regeln.

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`.

**Vorgelagert — offene Beobachtungen sichten:** bei der Anlage (2026-11-26): `BEO-REPO/kleines-profil-gruen-abnahmegroesse-scheitert` (1×; das Profil `rauch` war grün, die Größe `abnahme` beendet das Binary) und `BEO-REPO/fanger-endet-an-der-speichergrenze-des-containers` (1×; ein Fänger zum Speicher deckt die Größen nur bis zur Grenze seines Containers). Der Planner sichtet beim Start erneut.

Alle berührten Sub-Areas GF.
