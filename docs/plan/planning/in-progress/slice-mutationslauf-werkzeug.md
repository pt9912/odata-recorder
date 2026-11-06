# Slice slice-mutationslauf-werkzeug: Mutationsläufe der Rollen in einem Container-Lauf mit gemeinsamem Cache

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — **kein Produkt-Slice**: Er liefert keine Produktfunktion, liegt auf keinem Pfad eines Meilensteins und blockiert keinen produktiven Slice und keine produktive Welle. Er läuft in einem eigenen Worktree (Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)); ob er in eine Welle gehört (etwa [welle-sensoren](../welle-sensoren.md)), entscheidet der Planner im Integrationsschritt, dieser Plan ordnet ihn nicht ein.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit: gleiche Eingabe gleiches Testverhalten, Läufe im Container; das Lastenheft führt keine Anforderung an Werkzeuge der Rollen, der Bezug ist der nächste, den es gibt), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Rahmen der Gegenproben; Muster für Werkzeuge mit Treiber auf dem Host, dieser Slice entscheidet ihn nicht), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Verträge der Werkzeuge unter `harness/vertraege/`), [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** ein **neuer Vertrag** unter `harness/vertraege/` (Kennung aus dem Bereich der Linie C, `SPEC-200` bis `SPEC-229`; **die Nummer vergibt der Architect**, dieser Plan nennt keine) · `harness/vertraege/README.md` §Verträge (Zeile des neuen Vertrags, Architect) · `SPEC-122` (`test-race`: Detektor und Rate je Prozess; der Vertrag des Werkzeugs knüpft daran an, ändert ihn nicht) · `SPEC-128` (Muster, [Vertrag](../../../../harness/vertraege/SPEC-128-groessenmarke.md)) · `harness/README.md` §Sensors/Werkzeuge, *Nicht behauptet* (Implementer, falls das Werkzeug ein Ziel bekommt).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** pt9912 (Zweig slice-mutationslauf-werkzeug, Worktree /Development/odata-recorder-wt-mutationslauf)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Anlass (vom User bestätigt):** Die Mutationsläufe der Rollen Implementer, Reviewer und Verifier (`AGENTS.md` §3.10, §3.15, §3.17) sind der größte Teil der schweren Läufe in der Schlange am `flock /tmp/odata-recorder-heavy.lock` (Übergangsregeln Nr. 1). Befunde aus dem **Lesen** der Skripte der Läufe (`mb2.sh`, `mb3.sh`, `probe2.sh`, `runmut.sh`, `runconc.sh`, `inner*.sh` in den Scratchpads der Läufe `rev2-workspace`, `rev-workspace`, `rev2-geruest`, `url-pfad`, `telport`; die Skripte liegen **nicht im Repo**, und es wurde **nichts gemessen**):

1. Je Mutation wird der ganze Quellbaum kopiert (`cp -r /work /tmp/m`) und `go test -count=1` über das ganze Paket gefahren, nicht über den zuständigen Test; nur das Gerüst-Skript `batch.sh` nutzt `-run`.
2. Bei nebenläufigen Zusagen läuft je Mutation `go test -race -c`, dann 50 Prozesse mit und 50 ohne Detektor, also rund 100 Testläufe je Mutation; `AGENTS.md` §3.10 verlangt die Rate nur für den Fänger der Fenster-Mutation.
3. Jeder Reviewer legt einen eigenen kalten `GOCACHE` an; zwischen Implementer, Reviewer, Verifier und Runden gibt es keinen gemeinsamen.
4. Die Mutationen eines Reviews laufen als mehrere getrennte Jobs, jeder wartet einzeln am Lock.

Die Größe der Wirkung ist unbekannt; **Messwerte gibt es noch nicht**. Der Slice misst sie vor und nach der Änderung (Closure-Trigger in §5), der Plan setzt keine Zahl voraus.

**Ziel:** Ein Werkzeug (Programm oder Skript, das der Architect wählt) arbeitet eine **Mutationsliste** (je Eintrag: Name, Datei, Ersetzung, zuständiger Test, Klasse `nebenläufig` oder `andere`) in **einem** Container-Lauf mit **gemeinsamem Cache** ab und gibt das Ergebnis je Mutation aus. Es begrenzt dabei Speicher (`--memory` je Lauf) und räumt Kopien, Binaries und eigene Container auf (Z12, §6). Es setzt vier Ansatzpunkte des Users um — und nur diese vier: (a) `-run` auf den zuständigen Test statt auf das Paket; (b) das Paket einmal mit `-c` zu einem Test-Binary übersetzen und die Prozesse daraus starten statt je Lauf neu zu übersetzen; (c) ein Cache, den die Rollen gemeinsam nutzen; (d) die Rate (viele Prozesse) nur für die Mutationen der Klasse `nebenläufig`, nicht für jede. Das Werkzeug **entscheidet nicht, welche Mutationen nötig sind**: Die Liste schreibt die Rolle nach `AGENTS.md` §3.10, §3.15 und §3.17.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- **Ein Mutations-Protokoll mit Prüfsummen, das Ergebnisse zwischen Rollen wiederverwendet.** Die Unabhängigkeit der Rollen (`AGENTS.md` §3.10: der Implementer sieht seine Mutation selbst rot, Reviewer und Verifier messen weiter selbst) ist eine Entscheidung des Users und des Architect, kein Nebenprodukt eines Werkzeugs. Der Slice baut es nicht; die Frage steht in §6 mit Ausgang *offen* und Adresse Architect/User.
- **Eine Änderung am `Dockerfile` und an der Kette `make gates`.** Das Werkzeug ist kein Gate: Es prüft nichts, was die Kette zusagt, und eine Stufe im `Dockerfile` oder ein Glied von `GATE_CHECKS` wäre eine Änderung der Gate-Kette (Linie C, Übergangsregeln Nr. 7: ein Besitzer zugleich). Braucht das Werkzeug ein Image, entscheidet der Architect, wie es ohne diese Änderung entsteht, oder gibt den Punkt an den Planner zurück (§6).
- **Ein Test-Cache der Kette.** Ein Versuch dazu liegt im Zweig `adr-gate-laufzeit` (pausiert); er betrifft `make test` und nicht die Mutationsläufe der Rollen. Dieser Slice übernimmt ihn nicht und greift ihm nicht vor.
- **Eine Entscheidung darüber, welche Mutationen nötig sind, und über den Inhalt der Fänger.** Das ist Sache der Rollen und der Tabelle *Zusage · Fänger* des jeweiligen Slice (`AGENTS.md` §3.15); das Werkzeug führt aus.
- **Eine Lockerung von `AGENTS.md` §3.10.** Die Zusage „Rate in jedem Lauf rot“ für Nebenläufigkeit und die Zahl der Prozesse bleiben Sache des Fängers (Wirkungs-Fänger, siehe §6); das Werkzeug macht sie billiger, nicht kleiner (`AGENTS.md` §3.6).
- **Änderungen an `AGENTS.md`, `.claude/commands/`, `.claude/agents/` und `.harness/skills/`** (etwa damit die Rollen das Werkzeug rufen): Sie schreibt keine Rolle dieses Slice; sie gehen als Diff-Vorschlag an den User (Übergangsregeln Nr. 4).
- **Produktcode (`src/`).** Der Slice ist Werkzeug; sein Nutzen ist die Laufzeit der Prüfung, nicht ein Verhalten des Recorders.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Das Werkzeug arbeitet eine Mutationsliste in einem Container-Lauf ab und gibt je Mutation Name, Ergebnis (rot, grün, Übersetzungsfehler, Abbruch), Zahl der Prozesse und Laufzeit aus; jede Mutation läuft in einer eigenen Wegwerfkopie des Baums, und die Ergebnisse gleichen denen eines Einzellaufs; belegt durch Gegenprobe oder Test (Kopie mit bekannten Mutationen, Orakel im Test) und eine benannte Ausgabe.
- [ ] Die Ansatzpunkte (a) bis (d) aus §1 sind umgesetzt oder je Punkt mit Messung und Grund als nicht tragend benannt; belegt durch eine Messung vor und nach der Änderung an einem benannten, wiederholbaren Satz von Mutationen (Host-Zustand und Last genannt, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`).
- [ ] Der Vertrag des Werkzeugs liegt unter `harness/vertraege/` (Architect) samt Tabelle *Zusage · Fänger* und Raster; `harness/README.md` nennt das Werkzeug und seine Grenzen (*Nicht behauptet*).
- [ ] `make gates` grün; `make gegenproben` grün, falls der Slice Fälle der Gegenproben oder deren Eingabe ändert (Eingabe geändert, Lauf ausgeführt, Stempel erst nach dem letzten Eingabe-Commit; im Zweig nicht committet, Übergangsregeln Nr. 2).
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure, falls der Slice einer Welle zugeordnet wird, sonst hier.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

Der Planner schneidet; der Architect entscheidet (Aufträge unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten; die Pfade im Block in §4 nennt der Architect.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/<werkzeug>/` oder `tools/harness/<skript>` (Architect: Programm in Go wie `groessenmarke` oder Skript; Treiber auf dem Host, Lauf im Container) | neu | Mutationsliste lesen, Wegwerfkopien, Test-Binary übersetzen, Prozesse starten, Ergebnis je Mutation ausgeben ([ADR-0001](../../adr/0001-implementierungssprache-go.md)) |
| Test des Werkzeugs und Fälle der Gegenprobe | neu | Zusagen der Tabelle unten, Orakel aus dem Wortlaut des Vertrags; ob ein Fall der Gegenproben ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)) oder ein Test des Werkzeugs trägt, entscheidet der Architect |
| `harness/vertraege/SPEC-2NN-<kurztitel>.md` | neu (Architect) | Vertrag: Format der Liste, Ergebnis, Cache, Aufräumen, Randformen; Kennung aus `SPEC-200` bis `SPEC-229` |
| `harness/vertraege/README.md` | update (Architect) | Zeile in der Tabelle *Verträge* |
| `harness/README.md` | update | Werkzeug und *Nicht behauptet* (die Zusage „dieselbe Aussage wie ein Einzellauf“ nennt ihre Grenze) |
| `Makefile` / `harness/mk/` | **nicht geändert**, solange das Werkzeug kein Ziel braucht | ein Ziel in `harness/mk/` ist eine Änderung der Gate-Kette (Übergangsregeln Nr. 7); der Architect nennt, ob das Werkzeug ohne sie auskommt, sonst Rückgabe an den Planner (§6) |
| der Plan dieses Slice (`slice-mutationslauf-werkzeug.md` im Verzeichnis seines Zustands) | update | §3 bis §7 nach `AGENTS.md` §3.9 |

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Aufträge 1 bis 8, Tabelle *Zusage · Fänger*, Raster, Nachzüge, erwartete Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Closure (Planner). Worktree: `/Development/odata-recorder-wt-mutationslauf`, Zweig `slice-mutationslauf-werkzeug` (Linie C).

**Betriebsregeln für alle Rollen:** Wortlaut und Gründe in den [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln). Gate-Läufe und die Messungen dieses Slice laufen unter `flock /tmp/odata-recorder-heavy.lock`; jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container je Linie zugleich, vor einem großen Lauf `free -g`. Auf Repo-Dateien nur `Edit` und `Write`; Fixtures mit Escape-Folge oder Rohbytes aus Bytewerten und gegengelesen (`AGENTS.md` §3.18). `make gates` nie parallel zu Mutationsläufen. Der Implementer fährt für die Mutationen seiner Fänger selbst, was das Werkzeug ersetzen soll, bis das Werkzeug fertig und verifiziert ist.

**Aufträge an den Architect** (Entscheidungen, die kein Plan trifft; der Vertrag steht in `harness/vertraege/`, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), `AGENTS.md` §3.8; eine Entscheidung über eine Host-Abhängigkeit oder die Bindung eines Gates ist eine ADR aus dem Bereich 0030 bis 0039):

1. **Gleichwertigkeit.** Was „dieselbe rot/grün-Aussage wie ein Einzellauf“ heißt: Vergleichsmaß (Exit-Status des Tests, benannter fehlgeschlagener Test, Ausgabe), und welche Abweichungen das Werkzeug zulässt (Reihenfolge, Laufzeit) und welche nie (ein Grün, wo der Einzellauf rot wäre, und umgekehrt). `-run` auf den zuständigen Test ändert die Aussage: Wird ein anderer Test des Pakets durch die Mutation rot, sieht es der Lauf über das Paket, `-run` nicht. Der Architect entscheidet, ob das Werkzeug den zuständigen Test zusagt und das Paket als eigene Stufe nachfährt, oder ob es die Aussage als „der zuständige Test wird rot“ bestimmt, mit Folge für Reviewer und Verifier.
2. **Das Test-Binary.** `-c`: Lebensdauer, Ablage, Aufruf (`-test.run`, `-test.count=1`, Arbeitsverzeichnis des Pakets, Testdaten), Verhalten bei Paketen mit Testdaten unter dem Quellpfad, mit cgo (`-race`) und ohne. Die Mutation ändert Quellen, also muss je Mutation neu übersetzt werden; was dann „einmal übersetzen“ trägt (etwa je Mutation einmal übersetzen und alle Prozesse daraus starten), steht im Vertrag, nicht im Plan.
3. **Der Cache.** Was der gemeinsame Cache enthält (Go-Build-Cache, Modul-Cache), wo er liegt, wer ihn schreibt, ob ein Treffer das Ergebnis einer Mutation verändern kann (Cache-Einfluss auf die Aussage, §6) und wie das Werkzeug ihn aus der Aussage heraushält; Verhältnis zu `AGENTS.md` §3.1 (netzlos, ohne schreibbaren Mount): **die Ausnahme eines schreibbaren Mounts für den Cache ist eine offene Entscheidung des Users** (§6); der Architect legt die Alternativen vor (Volume, `docker cp` in und aus dem Container wie bei den `schema-*`-Zielen, Cache nur im Container je Lauf) und entscheidet nicht für den User.
4. **Die Klasse und die Rate.** Wie `nebenläufig` und `andere` in der Liste stehen; was für `nebenläufig` läuft (mit und ohne Detektor, Zahl der Prozesse, Zeitgrenze) und was für `andere` (ein Lauf); Abgrenzung zu `AGENTS.md` §3.10 (ein Datenrennen zählt der Detektor je Prozess mit `-test.count=1`, jede andere Wirkung eine Wiederholung). Das Werkzeug fährt die Rate nur auf Anforderung der Klasse; es entscheidet nicht, welche Mutation eine Fenster-Mutation ist (das tut die Rolle in der Liste).
5. **Wegwerfkopie und Aufräumen.** Umfang der Kopie (ganzer Baum oder nur das Paket und seine Abhängigkeiten; Module unter `tools/` und `src/`), Ablage, Abbruch (Signal, Zeitgrenze, Container beendet): was zurückbleibt und was nicht.
6. **Last und Grenzen.** Zahl gleichzeitiger Prozesse und Mutationen, Speicher und CPU (`--memory`, `--cpus` der Betriebsregeln), Verhalten bei Zeitgrenze je Mutation; Ausgabe von Zahl der Prozesse und Last zum Ergebnis, damit ein Rot unter Last erkennbar ist (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`).
7. **Ob das Werkzeug ohne Änderung der Gate-Kette auskommt** (kein Ziel in `harness/mk/`, kein `Dockerfile`-Eintrag): wie das Image entsteht, in dem es läuft, und ob eine ADR zur Bindung nötig ist. Kommt es nicht aus, geht der Punkt an den Planner (§6); die Linie C läuft nacheinander.
8. **Tabelle *Zusage · Fänger*** (unten), **Raster** (§6), **Nachzüge** nach `AGENTS.md` §3.9 und **erwartete Größe** (Zeilen Programm, Produktcode `p`, Port-Fakes; Programm mit Prozessen und Cache ist nebenläufig und zählt × 85) für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Spalten und Klassen nach [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 3 e). Die **Zusagen** sind Kandidaten des Planners aus dem Anlass; **Fänger, Mutation und Klasse** (`nebenläufig` oder `andere`) füllt der Architect, bis dahin steht „(Architect)“. Der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein (`AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z1 | Jede Mutation läuft in einer eigenen Wegwerfkopie des Baums; keine Mutation beeinflusst eine andere (Reihenfolge, Zahl, Ausfall der vorigen) | (Architect) | (Architect) | andere |
| Z2 | Das Ergebnis je Mutation gleicht dem Einzellauf (rot, grün) für jede Art von Mutation der Liste | (Architect: Schleife über die Arten; Orakel: Einzellauf auf frischer Kopie) | (Architect) | andere |
| Z3 | Ein Übersetzungsfehler der Mutation wird als solcher gemeldet und nicht als rot gezählt (`AGENTS.md` §3.10: ein Test zählt nur, wenn er aus dem benannten Grund rot wird) | (Architect) | (Architect) | andere |
| Z4 | Der Lauf trifft den zuständigen Test (`-run`); ein Name, der keinen Test trifft, ist ein Fehler und kein grünes Ergebnis | (Architect) | (Architect) | andere |
| Z5 | Zahl der Prozesse und Last werden mit dem Ergebnis ausgegeben | (Architect) | (Architect) | andere |
| Z6 | Die Rate (viele Prozesse, mit und ohne Detektor) läuft nur für Mutationen der Klasse `nebenläufig`; eine Mutation der Klasse `andere` startet einen Lauf | (Architect) | (Architect) | nebenläufig |
| Z7 | Der gemeinsame Cache verändert die Aussage nicht (Treffer und Fehlen des Caches geben dasselbe Ergebnis) | (Architect) | (Architect) | andere |
| Z8 | Abbruch (Signal, Zeitgrenze) räumt Kopien und Prozesse auf; nach dem Lauf bleibt nichts zurück | (Architect) | (Architect) | nebenläufig |
| Z9 | Das Werkzeug ändert den Baum des Aufrufers nicht (Mutationen nur in Kopien) | (Architect) | (Architect) | andere |
| Z10 | Das Werkzeug läuft netzlos; ein schreibbarer Pfad ist nur der, den die Entscheidung des Users nennt | (Architect) | (Architect) | andere |
| Z11 | Die Mutationen eines Aufrufs laufen in einem Lauf des Containers, nicht als getrennte Jobs (eine Wartezeit am Lock je Aufruf, nicht je Mutation) | (Architect) | (Architect) | andere |
| Z12 | Speicher und Platz sind begrenzt: jeder Container trägt `--memory` (≤ 4g, Mutationsläufe 2g), kein zweiter Go-Build zugleich, Wegwerfkopien, Binaries, eigene Container und Images sind nach dem Lauf weg, der Cache hat Größengrenze oder Aufräumweg (Zusatz des Users 2026-11-06) | (Architect) | (Architect) | nebenläufig |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): nicht produktiv, blockiert keine produktive Linie und wartet auf keine. Er startet erst, wenn **beobachtbar** die Schlange am `flock /tmp/odata-recorder-heavy.lock` kurz ist (kein wartender Lauf einer produktiven Linie, wenn der Slice einen schweren Lauf braucht) **oder auf ausdrückliche Anweisung des Users**; kein Datum. Außerdem liegt kein anderer Slice der Gate-Kette in `in-progress/` (Übergangsregeln Nr. 7), falls der Architect für das Werkzeug eine Änderung dort braucht (Auftrag 7).
**Start am 2026-11-06 auf ausdrückliche Anweisung des Users** (nicht wegen kurzer Schlange). **Bedingung für den Architect:** `slice-test-determinism` liegt in `in-progress/` und ändert `harness/mk/` und die Sensors-Tabelle (Gate-Kette). Aus dem Plan ist keine Änderung der Gate-Kette durch dieses Werkzeug erkennbar (§3: `Makefile` und `harness/mk/` nicht geändert); der Slice läuft nur weiter, wenn der Architect in Auftrag 7 bestätigt, dass das Werkzeug **ohne** Ziel in `harness/mk/`, ohne Stufe im `Dockerfile`, ohne Eintrag in `GATE_CHECKS` und ohne Änderung der Sensors-Tabelle auskommt. Braucht er eine, gibt er an den Planner zurück (§6), und der Slice geht nach `open/`, bis `slice-test-determinism` geschlossen ist. `harness/README.md` (*Nicht behauptet*, Werkzeug-Zeile) ist ebenfalls Gate-Kette (Übergangsregeln Nr. 7): der Implementer ändert sie erst, wenn `slice-test-determinism` in `done/` liegt, oder der Planner ordnet die Änderung beim Merge. Die Architect-Aufträge 1 bis 8 sind entschieden, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Größenmarke** (`AGENTS.md` §3.14): **noch nicht gesetzt.** Der Architect füllt den Teil `pfad`, `produktcode`, `fakes` des Blocks ([SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 3); der Planner schreibt danach `basis`, `schaetzung` und `marke` und setzt die Marke mit `make groessenmarke` im eigenen Commit vor dem ersten Code-Commit. Die Werte unten sind **Platzhalter**, keine Schätzung; der Plan liegt in `open/` und wird dort nicht gemessen. Aufwand nach Einschätzung des Planners: S bis M (Programm mit Prozessen und Cache, nebenläufig).

```groessenmarke
pfad: tools/mutationslauf
pfad: tools/harness/mutationslauf.sh
produktcode: 500
fakes: 3
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Schätzung des Architect mal 1,25 liegt über 3000 (`AGENTS.md` §3.14), oder die Messung nach dem Code liegt über der Marke (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Ein Schnitt ist vorab benannt: der Ansatzpunkt (c), der gemeinsame Cache über die Rollen, teilt keinen Mechanismus mit (a), (b) und (d) und lässt sich als eigener Teil zuletzt schreiben (er hängt an der Entscheidung des Users zum schreibbaren Mount).
- `in-progress` → `open` (blockiert): der User gibt den schreibbaren Pfad für den Cache nicht frei **und** der Architect findet keinen Weg ohne ihn, der den Cache über die Rollen trägt (dann Slice ohne (c)); oder das Werkzeug braucht eine Änderung der Gate-Kette, während ein anderer Slice der Linie C in `in-progress/` liegt.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

- Die DoD aus §2 ist vollständig, `make gates` ist grün (im Hauptbaum, nach dem Merge, auf ruhigem Host), der Verifier hat bestanden.
- Die Messung aus §2 liegt vor: derselbe Satz von Mutationen mit dem bisherigen Vorgehen und mit dem Werkzeug, Laufzeit und Zahl der Prozesse, Host-Zustand genannt, jedes Ergebnis je Mutation gleich. Misst das Werkzeug keinen Gewinn für einen Ansatzpunkt, steht das mit Zahl in §7 (kein Erfolg ohne Messung).
- Der Lerneintrag nennt, welcher der vier Ansatzpunkte die Laufzeit der Mutationsläufe trug und welcher nicht, und ob die Rollen das Werkzeug rufen (Diff-Vorschlag an den User für `AGENTS.md` und die Commands, nicht Teil des Slice).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* füllt der Architect vor dem ersten Code-Commit in den Vertrag; die Zellen unten nennen als Ort **Architect** (offen) oder „gilt nicht“ mit Grund. Eine Zelle ohne Ort hält den Implementer an. Zusatztabelle mit der Spalte *Fänger* (je Paar von Verstößen am selben Schlüssel und je Modus): der Architect legt sie an, wenn eine Operation mehrere Verstöße meldet; Kandidat: eine Mutation mit unbekannter Datei **und** leerer Ersetzung. **Ausgang:** offen, Adresse Architect.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Liste lesen | Feld einer Mutation fehlt (Name, Datei, Ersetzung, Test, Klasse): Architect | leere Liste; leerer Name; leere Ersetzung (Löschung?): Architect | Klasse außerhalb der zwei, Datei außerhalb des Baums oder nicht vorhanden, Name doppelt, nicht lesbares Format: Architect | gilt nicht (Lesen ändert nichts, es endet mit dem Aufruf) | gleiche Liste zweimal, Ergebnis gleich (Z2): Architect |
| Mutation anwenden (Ersetzung in der Kopie) | Datei fehlt in der Kopie: Architect | Ersetzung leer oder gleich dem Original (die Mutation ändert nichts, Meldung?): Architect | Ersetzung trifft nicht oder mehrfach (Stelle nicht eindeutig): Architect | Kopie halb geschrieben: Architect (Z8) | gleiche Mutation zweimal in einem Lauf: Architect |
| Test fahren (`-run`, Binary) | zuständiger Test fehlt in der Liste: Architect (Z4) | Muster trifft keinen Test (null Tests): Architect (Z4) | Name, der mehrere Tests trifft; Paket ohne Tests; Übersetzungsfehler: Architect (Z3) | Zeitgrenze, Signal, Prozess beendet sich selbst: Architect (Z8) | Wiederholung bei flatterigem Lauf: Architect (Z6) |
| Rate fahren (nebenläufig) | Zahl der Prozesse fehlt: Architect | Zahl 0: Architect | Zahl über der Grenze von Speicher und CPU: Architect (Auftrag 6) | Abbruch mitten in den Prozessen: Architect (Z8) | gilt nicht (die Rate ist selbst die Wiederholung) |
| Cache nutzen | Cache fehlt (kalt), gleiche Aussage: Architect (Z7) | leerer Cache, wie kalt: Architect (Z7) | beschädigter oder fremder Cache (andere Go-Version, anderer Baum): Architect | Abbruch beim Schreiben des Caches: Architect | zweiter Lauf trifft den Cache: Architect (Z7) |
| Ergebnis ausgeben | Mutation ohne Ergebnis (Abbruch vor dem Test): Architect | leere Ausgabe des Tests: Architect | Ausgabe des Tests, die das Werkzeug nicht deutet: Architect | Abbruch vor der Ausgabe: Architect (Z8) | gilt nicht (die Ausgabe ändert nichts) |
| Aufräumen | nichts zu räumen: Architect | leere Kopie: Architect | Ablage nicht schreibbar oder nicht löschbar: Architect | Abbruch während des Räumens: Architect | zweiter Aufruf nach Abbruch findet Reste: Architect (Z8) |

Risiken und offene Punkte:

- **Gleichwertigkeit mit dem bisherigen Ergebnis je Mutation.** `-run` auf den zuständigen Test sieht ein Rot eines anderen Tests im Paket nicht, das der Lauf über das Paket zeigt; ein Test-Binary je Mutation und ein gemeinsamer Cache können dieselbe Aussage verschieben. Das Werkzeug darf kein Grün liefern, wo der Einzellauf rot wäre — **Ausgang:** offen, Adresse Architect (Auftrag 1, Z2, Z4); Nachweis: Schleife über die Arten der Mutation gegen den Einzellauf (`AGENTS.md` §3.17).
- **Cache-Einfluss auf die Aussage.** Ein Treffer im Testcache von `go test` (ohne `-count=1`) wäre ein falsches Grün; der Cache der Übersetzung ändert die Aussage nicht, aber das Werkzeug muss das zusagen und belegen — **Ausgang:** offen, Adresse Architect (Auftrag 3, Z7).
- **Die Rate der Fenster-Mutationen bleibt ein Wirkungs-Fänger-Thema.** `AGENTS.md` §3.10 verlangt für Nebenläufigkeit „in jedem Lauf rot“; das Werkzeug macht den Lauf billiger, ersetzt aber den Fänger nicht, der die Wirkung zählt (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`). Wer die Rate nur für Fenster-Mutationen fährt, verlässt sich darauf, dass die Rolle die Klasse richtig setzt (setzt sie `andere` für eine nebenläufige Zusage, fehlt die Rate, und das Werkzeug merkt es nicht) — **Ausgang:** weiter offen → Register; der Architect nennt die Grenze im Vertrag als akzeptiertes Negativ mit Grund und Folge (`BEO-REPO/akzeptiertes-negativ-ohne-folge`).
- **Docker-only und kein schreibbarer Mount (`AGENTS.md` §3.1).** Ein Cache über die Rollen braucht einen Ort, der den Lauf überlebt; die Läufe sind netzlos und ohne schreibbaren Mount. Der User erwägt eine Ausnahme für den Cache. **Offene Entscheidung des Users**, nicht des Planners und nicht des Architect: ob ein schreibbarer Pfad (Volume oder Verzeichnis) für den Cache erlaubt ist, wer ihn schreibt und ob er die Läufe der Rollen trennen darf — **Ausgang:** offen, Adresse User (der Architect legt die Alternativen vor, Auftrag 3). Ohne Freigabe entfällt der Ansatzpunkt (c); (a), (b) und (d) laufen mit einem Cache im Container je Lauf.
- **Speicher und Platz sind Anforderung an das Werkzeug (Zusatz des Users, 2026-11-06; Host 31 GB RAM, 17 GB verfügbar, ein fremder Container läuft nebenher).** Das Werkzeug setzt `--memory` je Lauf selbst (Obergrenze 4g, Mutationsläufe 2g, Wert vom Architect im Vertrag), fährt nicht mehrere Go-Builds zugleich und räumt Wegwerfkopien, Test-Binaries und eigene Container/Images nach dem Lauf, auch nach Abbruch (Z8); der Cache hat eine Größengrenze oder einen Aufräumweg (Auftrag 3, 6). Die Zusage steht als Z12 in §3; **Ausgang:** offen, Adresse Architect.
- **Ein Mutations-Protokoll mit Prüfsummen, das Ergebnisse zwischen Rollen wiederverwendet** (gleicher Baum, gleiche Mutation, gleiche Quellen: gleiches Ergebnis ohne Neulauf), würde Läufe sparen und die Unabhängigkeit der Rollen aufheben (`AGENTS.md` §3.10, §3.15: Reviewer und Verifier messen weiter selbst). Das ist eine Entscheidung des Users und des Architect über die Rollen, kein Werkzeug-Detail; der Slice baut es nicht — **Ausgang:** offen, Adresse Architect und User; eine Entscheidung dafür wäre ein eigener Slice.
- **Das Werkzeug braucht eine Änderung der Gate-Kette** (ein Ziel in `harness/mk/`, eine Stufe im `Dockerfile`, einen Eintrag in `GATE_CHECKS`) und kollidiert mit einem anderen Slice der Linie C (Übergangsregeln Nr. 7) — **Ausgang:** offen, Adresse Planner (Rückgabe des Architect nach Auftrag 7), dann Reihenfolge der Linie C oder Schnitt.
- **Das Werkzeug wird von keiner Rolle gerufen.** Die Anweisungen der Rollen (`.claude/commands/`, `.claude/agents/`, `.harness/skills/`) schreibt dieser Slice nicht; ein Werkzeug ohne Aufrufer spart nichts — **Ausgang:** offen, Adresse Planner (Diff-Vorschlag an den User bei der Closure, Übergangsregeln Nr. 4).
- **Die Messung des Gewinns ist selbst ein schwerer Lauf unter der Last des geteilten Hosts** (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`); eine Laufzeit ohne Host-Zustand ist keine Messung — **Ausgang:** weiter offen → Register; der Slice nennt Host-Zustand und Last zu jeder Zahl und wiederholt ein Rot unter Last auf ruhigem Host.
- **Die Befunde (1) bis (4) in §1 stammen aus dem Lesen von Skripten, ohne Messung**; welcher Ansatzpunkt wie viel trägt, ist ungemessen — **Ausgang:** weiter offen → Messung in §5; bis dahin ist keine Zahl im Plan ein Befund.

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

*(Wird bei der Closure gefüllt.)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt wird die Sub-Area `*` (Kürzel `REPO`, Greenfield, `harness/conventions.md` §Modus-Deklaration); `harness/` und `tools/` gehören zu ihr, eine feinere Sub-Area ist nicht deklariert und nicht nötig.

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Treffer: `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (verkörpert in `AGENTS.md` §3.17; die Mutationsläufe der Fix-Runden sind Gegenstand, den das Werkzeug billiger macht) · `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (beide offen; Laufzeit und Last des geteilten Hosts betreffen die Messung und die Rate) · `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15; „dieselbe Aussage wie ein Einzellauf“ muss ihre Grenze nennen) · `BEO-REPO/akzeptiertes-negativ-ohne-folge` (für jeden benannten Rest, etwa die Klasse der Rolle) · `BEO-REPO/randform-erst-im-review-entschieden` (für das Raster in §6) · `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (für die Betriebsregel „nur `Edit` und `Write`“). Die Zähler stehen im Register und werden hier nicht kopiert. Keiner erreicht mit diesem Slice 3× neu.

**Modus-Begründungsblock:** alle berührten Sub-Areas GF.
