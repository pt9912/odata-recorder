# Slice slice-gegenproben-zeitbudget: Messvorschrift für das Zeitbudget der Gegenproben

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice und keine Closure-Bedingung, die von der DoD verschieden wäre (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln): berührt `tools/gegenproben/` und den Stempel, Nr. 7 (ein Besitzer der Gate-Kette zugleich).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Laufzeit der Gegenproben, der Bezug ist der nächste, den es gibt), [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (Entscheidung, `Accepted`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Re-Evaluierungs-Trigger „über 30 s“, dessen Ausgang [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) ist), [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Messung durch den Host).

**Berührte Spec-Stellen:** `SPEC-102` (Absatz *Zeitbudget* und Nr. 6) · gegebenenfalls ein weiterer Fall der Gegenproben in `SPEC-102`; die Nummer vergibt der Architect.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Planner, Auftrag des Users nach `Accepted` von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)). **Datum:** 2026-11-08.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Das Soll „höchstens 120 s“ für `time make gegenproben` (geänderte Eingabe, kalter Cache) hat in `SPEC-102` eine Messvorschrift: Es gilt nur für einen Lauf bei Load Average (1 Minute) höchstens der Kernzahl, zu Beginn und am Ende; ein Lauf darüber ist *ungültig*, kein Bruch, und wird wiederholt. Die Frist je Fall (`fallFrist`, 120 s) bleibt unverändert die einzige maschinell geprüfte Schwelle.

**Hintergrund:** [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) wählt Option B. Die Einzelheiten (Messgröße, Lastgrenze, Zahl der Wiederholungen, Ausgang bei Bruch) stehen laut ADR in `SPEC-102`. Ein Entwurf des Architect liegt vor (Messvorschrift Nr. 1 bis 4 und ein Zusatz am Ende von Nr. 6); der Architect prüft ihn gegen den Vertrag und übernimmt ihn nicht blind.

**Umfang:**

1. `SPEC-102`: Messvorschrift Nr. 1 bis 4 und Zusatz zu Nr. 6 (ein Lauf mit gerissener Frist bei Last über der Kernzahl ist ungültig; in einem gültigen Lauf ist das ein Befund an den Planner); der Stand-Satz „Das Budget liegt rund 28 % über …“ entfällt, die Messreihen bleiben als Belege.
2. `tools/gegenproben`: nur so weit, wie [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) es trägt. Die ADR überlässt dem Slice, **ob** das Programm die Fallzeiten ausgibt (ein Weg, Last und Aufwand zu trennen). Ob das Programm Load Average und Kernzahl misst und einen Lauf über der Grenze als *ungültig* meldet (statt nur die Frist zu reißen), steht in der ADR nicht; der Architect entscheidet es mit Begründung, oder die Messvorschrift bleibt reine Prosa für den Implementer.
3. Gegenproben-Fälle, falls der Architect Programmverhalten (Ausgabe der Fallzeiten, Meldung der Last) zusagt: ein Fall je Zusage, Stempel `harness/gegenproben.stempel` nachgezogen (Eingabe ändert sich).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Anhebung der Frist je Fall oder des Soll von 120 s — Lockerung (`AGENTS.md` §3.6), von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) ausgeschlossen; eine Anhebung ist eine eigene ADR mit gültigen Messwerten (Re-Evaluierungs-Trigger der ADR).
- Ein Prüfer, Gate oder Hook für das Gesamtbudget — [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) benennt das Soll ausdrücklich als nicht maschinell geprüft.
- Eine Aufteilung der Fälle auf Stufen (Option C der ADR) — eigener Slice, nur nach einer Messung, dass die Fälle und nicht die Last die Zeit bestimmen.
- Die Eichung auf ruhigem Host — Folgepflicht, kein Liefer-Punkt: sie braucht einen Zeitpunkt, an dem andere Läufe von `make` pausieren; erst messen, dann Zahl (§6, Ausgang *weiter offen*, Adresse: Planner).
- Änderungen am Format der Messreihen der Fälle außer dem Kopf der Vorschrift.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [ ] `SPEC-102` trägt die Messvorschrift (Gültigkeit eines Laufs, Lastgrenze, Wiederholung, Ausgang bei Bruch, Zusatz zu Nr. 6) und benennt das Soll als nicht maschinell geprüft; die Frist je Fall ist unverändert.
- [ ] Programmverhalten und Gegenproben-Fälle nur nach der Entscheidung des Architect; jede Zusage hat eine Zeile in der Tabelle *Zusage · Fänger* (§3) mit einer vom Implementer rot gesehenen Mutation, die Frist je Fall bleibt von einem Fall gefangen.
- [ ] Doku-Update: `harness/README.md` (Zeile `make gegenproben`, falls die Beschreibung sich ändert), `SPEC-102` *Stand*; `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

Der Planner schneidet; der Architect entscheidet (Aufträge unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/vertraege/SPEC-102-gegenproben.md` | update | Messvorschrift, Zusatz zu Nr. 6, Stand-Satz entfällt |
| `tools/gegenproben/` | update (nur nach Entscheidung des Architect) | Ausgabe der Fallzeiten, Meldung eines Laufs über der Lastgrenze; keine Änderung von `fallFrist` |
| `harness/gegenproben.stempel` | update (falls das Programm sich ändert) | Eingabe der Gegenproben ändert sich; Nr. 7 der Übergangsregeln |
| `harness/README.md` | update (falls nötig) | Beschreibung des Ziels |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Vertrag, Entscheidungen, Tabelle *Zusage · Fänger*, Raster, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Aufträge an den Architect:**

1. **Entwurf prüfen.** Die Messvorschrift Nr. 1 bis 4 und der Zusatz zu Nr. 6 aus dem Entwurf des Architect gegen den Vertrag und [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md): Messgröße (Load Average 1 Minute, Kernzahl des Hosts, Messzeitpunkte), Wiederholungsregel, Ausgang bei Bruch; keine Zahl ohne Messung, keine Anhebung einer Frist.
2. **Programmverhalten.** Ob das Programm die Fallzeiten ausgibt und ob es die Last misst und einen Lauf darüber als ungültig meldet; wenn ja: die Zusagen, der Ausgang (Exit-Code, Text), das Verhalten der Fälle 1 bis 37 und aller weiteren unter dem Programm, das Verhältnis zu `Zeit überschritten` (Nr. 6). Wenn nein: das als Entscheidung mit Grund in `SPEC-102`.
3. **Randformen** vor dem ersten Code-Commit im Vertrag (Raster in §6, `AGENTS.md` §3.16): Load Average nicht lesbar, Kernzahl nicht lesbar, Last genau gleich der Kernzahl (beide Seiten der Grenze), Last zu Beginn niedrig und am Ende hoch (und umgekehrt), Lauf bei unveränderter Eingabe (kein Lauf, 1 s), Abbruch des Laufs.
4. **Tabelle *Zusage · Fänger*** (§3), **erwartete Produktcode-Größe** für die Marke (Zeilen Go und Vertrag getrennt).
5. **Eichung.** Wie der erste gültige Lauf auf ruhigem Host seine Werte einträgt (Ort in `SPEC-102`, Form der Messung nach Nr. 1); der Slice führt sie nicht aus (§1).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect): folgt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln: ein Besitzer zugleich). Reihenfolge der Kette seit 2026-11-08: `slice-gate-integration-determinism`, `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, `slice-groessenformel-neu`, dann dieser Slice; der Start legt fest, ob er vor `slice-groessenformel-neu` rückt, solange dessen Beschluss des Users aussteht. Höchstens drei Slice-Worktrees zugleich. [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) ist `Accepted`.

**Größenmarke:** Prosa, kein Block, vorläufig. Die Marke setzt der Planner vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14; die Schätzung kommt aus der Tabelle des Architect. Erwartet ist ein kleiner Slice (Vertragstext, wenig Go, wenige Fälle); die Tabelle zur Formel fehlt noch, der Planner rechnet nach dem Architect neu und schneidet vor dem Code, wenn *S* die Schnittgrenze von `SPEC-128` Nr. 1 erreicht.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): der Architect entscheidet, dass die Last des Hosts im Programm nicht ohne eine ADR zu lesen ist (Fremdabhängigkeit, Mount), oder ein Slice der Gate-Kette hält Nr. 7.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag (Messvorschrift als geschärfte Regel in `SPEC-102` mit Anker, oder benannte Spec-Lücke).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Die Eichung auf ruhigem Host steht aus; ohne sie bleibt unbekannt, ob 120 s tragen — **Ausgang:** weiter offen; Adresse: Planner (Termin bei ruhigem Host, andere Läufe von `make` pausiert; liegt der gültige Lauf über 120 s, gilt der Re-Evaluierungs-Trigger von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)).
- Auf einem Host mit Dauerlast gibt es keinen gültigen Lauf (Negativ der ADR) — **Ausgang:** weiter offen; Folge: Meldungen nennen den Lauf als ungültig statt als Bruch, die Closures führen die Frage nicht mehr als offen.
- Das Programm wird um eine Last-Messung erweitert, die [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) nicht verlangt, und die Gegenproben bekommen eine Host-Abhängigkeit — **Ausgang:** weiter offen; Adresse: Architect (Auftrag 2).
- Raster der Randformen (Auftrag 3) ist noch leer — **Ausgang:** weiter offen: der Architect füllt es vor dem ersten Code-Commit (`AGENTS.md` §3.16).
- Der Slice berührt `tools/gegenproben/` und den Stempel, kollidiert mit einem anderen Kettenslice — **Ausgang:** weiter offen: Nr. 7, einer zugleich, Start laut §4.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus dem Beschluss des Users zu [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (2026-11-08); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** keine Beobachtung des Registers betrifft das Zeitbudget der Gegenproben unmittelbar; der Architect sichtet das Register beim Start.
