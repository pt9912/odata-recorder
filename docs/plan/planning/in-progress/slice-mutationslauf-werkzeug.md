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

**Berührte Spec-Stellen:** ein **neuer Vertrag** unter `harness/vertraege/`: [SPEC-210](../../../../harness/vertraege/SPEC-210-mutationslauf.md) (Bereich der Linie C, `SPEC-200` bis `SPEC-229`; die Nummer hat der Architect vergeben, `SPEC-201` bis `SPEC-209` bleiben für die anderen Pläne der Linie frei) · `harness/vertraege/README.md` §Verträge (Zeile des Vertrags, Architect) · `SPEC-122` (`test-race`: Detektor und Rate je Prozess; der Vertrag des Werkzeugs knüpft daran an, ändert ihn nicht) · `SPEC-128` (Muster, [Vertrag](../../../../harness/vertraege/SPEC-128-groessenmarke.md)) · `harness/README.md` §Sensors/Werkzeuge, *Nicht behauptet* (Implementer, erst nach `slice-test-determinism`, §4).
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
| `tools/mutationslauf/` (Programm in Go, eigenes Modul ohne Abhängigkeiten, wie `groessenmarke`) und `tools/harness/mutationslauf.sh` (Treiber auf dem Host: `docker create`, `docker cp`, `docker start -a`, `docker rm`) | neu | Liste lesen, Wegwerfkopien, Test-Binary übersetzen, Prozesse starten, Ergebnis je Mutation ausgeben ([SPEC-210](../../../../harness/vertraege/SPEC-210-mutationslauf.md), [ADR-0001](../../adr/0001-implementierungssprache-go.md)) |
| Tests des Moduls `tools/mutationslauf/` | neu | Zusagen der Tabelle unten, Orakel aus dem Wortlaut des Vertrags; **keine Fälle der Gegenproben** ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)): sie liefen in der Stufe `gegenproben` und bräuchten eine Änderung des `Dockerfile` (`COPY`, Programm), also der Gate-Kette; die Tests laufen im Image `odata-recorder:race` (SPEC-210 Nr. 3 a, 12 d) |
| `harness/vertraege/SPEC-210-mutationslauf.md` | neu (Architect, geliefert) | Vertrag: Format der Liste, Ergebnis, Cache, Aufräumen, Randformen |
| `harness/vertraege/README.md` | update (Architect, geliefert) | Zeile in der Tabelle *Verträge* |
| `harness/README.md` | update (Implementer, erst nach `slice-test-determinism` in `done/`) | Werkzeug und *Nicht behauptet* (die Zusage „dieselbe Aussage wie ein Einzellauf“ nennt ihre Grenze: SPEC-210 Nr. 5 e, 6 d, 12) |
| `Makefile` / `harness/mk/` / `Dockerfile` | **nicht geändert** (Architect bestätigt, Auftrag 7) | das Werkzeug hat kein Ziel, keine Stufe und keinen Eintrag in `GATE_CHECKS`; es nutzt das Image, das `make test-race` baut |
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

**Entscheidungen des Architect** (2026-11-06; Wortlaut, Zahlen und Gründe in [SPEC-210](../../../../harness/vertraege/SPEC-210-mutationslauf.md), hier nur die Adresse):

1. **Gleichwertigkeit:** Das Werkzeug sagt den zuständigen Test zu **und** fährt das Paket als Stufe 2 nach, wenn Stufe 1 grün war; Ergebnisse `rot`, `rot-paket`, `gruen`. Vergleichsmaß ist der Exit-Status des Test-Prozesses; *rot oder grün* gleicht dem Einzellauf `go test -count=1 [-race] <Paket>`. Nie ein Grün, wo der Einzellauf rot wäre; Ausnahme als benanntes Negativ: ein Befund, den nur `go vet` meldet (Nr. 5 e, 6 d, 12 a). Folge für Reviewer und Verifier: sie lesen `rot-paket` als gefangen, aber nicht vom zuständigen Test. Nr. 5, 6.
2. **Test-Binary:** je Mutation und Art ein Binary in der Wegwerfkopie, alle Prozesse der Mutation starten daraus; Arbeitsverzeichnis ist das Paketverzeichnis der Kopie; eine **Nullprobe** am unveränderten Stand sichert, dass der Test genau einmal getroffen wird und der Ausgangsstand grün ist (kein Rot durch eine kaputte Umgebung). Nr. 5.
3. **Cache:** Go-Build-Cache im Container, warm aus dem Image `odata-recorder:race` (292 MB, gemessen), gemeinsam für alle Mutationen und Arme eines Aufrufs und für alle Rollen, die dasselbe Image nutzen; Größengrenze 1,5 GiB mit `go clean -cache`; kein Testcache (der Test läuft aus dem Binary). **Kein Mount, kein Volume**; die Ausnahme (c) bleibt die Entscheidung des Users (§6, Empfehlung dort). Nr. 8.
4. **Klasse und Rate:** `andere` ein Prozess je Stufe; `nebenläufig` zwei Arme (mit Detektor, ohne), je 50 Prozesse als Vorgabe, je Mutation 0 bis 200 einstellbar, höchstens vier zugleich, Ergebnis `rot` nur bei allen Prozessen rot, sonst `flatternd`. Nr. 7.
5. **Kopie und Aufräumen:** Wurzelmodul = `go.mod`, `go.sum`, `src`, `tests` (wie die Stufe `source`), anderes Modul = sein Verzeichnis; Ablage im beschreibbaren Dateisystem des Containers (kein tmpfs, es zählte gegen `--memory`), Kopie wird nach der Mutation entfernt; der Container wird in jedem Fall vom Treiber entfernt, Reste ohne Treiber trägt höchstens ein Container mit Label, den der nächste Aufruf räumt. Nr. 4, 10.
6. **Last:** `--memory=2g --memory-swap=2g --cpus=4 --pids-limit=1024`, `GOMEMLIMIT=1500MiB`, `-p 2`, ein Bau zugleich, Mutationen nacheinander; Zeitgrenzen 2 Minuten je Test-Prozess (30 Sekunden Gnade), 10 Minuten je Bau, 3600 Sekunden je Aufruf; Ergebniszeile mit `prozesse`, `last`, CPUs. Nr. 9. **Messen offen:** ob 2g für den Bau von `modernc.org/sqlite` mit `-race` reicht, ist nicht gemessen (§6).
7. **Gate-Kette: bestätigt, das Werkzeug kommt ohne Änderung aus.** Kein Ziel in `harness/mk/`, keine Stufe im `Dockerfile`, kein Eintrag in `GATE_CHECKS`, keine Änderung der Sensors-Tabelle. Das Image ist `odata-recorder:race` (bzw. `odata-recorder:test` für Listen ohne `nebenläufig`), das `make test-race` bzw. `make test` ohnehin baut; fehlt es, endet das Werkzeug mit Exit 2 und der Meldung. Eine ADR ist nicht nötig: keine neue Host-Abhängigkeit (`bash`, `tar`, `docker` wie `groessenmarke-docker.sh`), keine Bindung eines Gates, keine Lockerung. **Ein Folge-ADR wäre nötig, wenn der User (c) mit schreibbarem Mount freigibt** (Bereich 0030 bis 0039). Preis: das Modul `tools/mutationslauf` hat kein Gate (SPEC-210 Nr. 12 d, §6).
8. **Tabelle, Raster, Größe:** unten, in §6 und §4.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Spalten und Klassen nach [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 3 e). Fänger sind Tests im Modul `tools/mutationslauf/` (Namen als Vorschlag, der Implementer darf sie anders schneiden, solange die Zeile gefangen bleibt); „Fake“ ist ein Ersatz für den Prozess-Port (zählt Starts, Gleichzeitigkeit, Argumente, blockiert bis zur Freigabe) oder ein Ersatz-Skript für `docker`; „Fixture“ ist ein Wegwerfmodul mit bekannten Mutationen und deren Orakel (Einzellauf auf frischer Kopie). Der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein (`AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z1 | Jede Mutation läuft in einer eigenen Wegwerfkopie des Baums; keine Mutation beeinflusst eine andere (Reihenfolge, Zahl, Ausfall der vorigen) | `TestKopieJeMutation`: Fixture, Schleife über die Listen [A,B], [B,A], [A mit Übersetzungsfehler, B], [A,A unter zwei Namen]; das Ergebnis von B ist in allen gleich dem allein gefahrenen B | die Kopie einer Mutation wird für die nächste wiederverwendet (nicht frisch angelegt): B sieht die Datei von A | andere |
| Z2 | Das Ergebnis je Mutation gleicht dem Einzellauf (rot, grün) für jede Art von Mutation der Liste | `TestGleichwertigWieEinzellauf`: Fixture, Schleife über die Arten (zuständiger Test rot · anderer Test des Pakets rot · äquivalente Mutation grün · Syntaxfehler · Typfehler · Panic im Test · Endlosschleife bis zur Zeitgrenze · Datenrennen nur mit Detektor) mal beide Klassen; Orakel `go test -count=1 [-race]` auf frischer Kopie | Stufe 2 (Paket) entfällt: die Art „anderer Test rot“ gibt `gruen`, das Orakel rot | andere |
| Z3 | Ein Übersetzungsfehler der Mutation wird als solcher gemeldet und nicht als rot gezählt (`AGENTS.md` §3.10: ein Test zählt nur, wenn er aus dem benannten Grund rot wird) | `TestUebersetzungsfehlerIstKeinRot`: Fixture, Schleife über Syntaxfehler, Typfehler, fehlenden Import; erwartet `uebersetzungsfehler`, kein gestarteter Prozess (Fake zählt null), Exit 1 | der Exit-Code von `go test -c` wird als rot gewertet | andere |
| Z4 | Der Lauf trifft genau den zuständigen Test (`-test.run`); ein Name, der keinen oder mehrere Tests trifft, ein Paket ohne Tests und ein roter Ausgangsstand sind ein Fehler und kein grünes Ergebnis | `TestNullprobe`: Fixture, Schleife über Name ohne Treffer, Name als Präfix eines zweiten Tests (`TestA`, `TestAB`), Paket ohne Testdatei, roter Ausgangsstand; erwartet Exit 2, die Meldung aus SPEC-210 Nr. 5 a und keine Ergebniszeile | der Anker `$` im Muster fehlt (`TestA` trifft `TestAB`); null Treffer zählt als grün | andere |
| Z5 | Zahl der Prozesse und Last werden mit dem Ergebnis ausgegeben | `TestErgebniszeile`: Fake; Schleife über die sechs Ergebnisse mal beide Klassen, die Zeile gleicht dem Format von SPEC-210 Nr. 1 b, `prozesse` gleicht der Zahl der Starts des Fakes (Nullprobe nicht gezählt), `last` gleicht dem Ersatzwert von `/proc/loadavg` | die Nullprobe wird in `prozesse` mitgezählt; `last` zeigt den Wert am Ende statt am Beginn | andere |
| Z6 | Die Rate (viele Prozesse, mit und ohne Detektor) läuft nur für Mutationen der Klasse `nebenläufig`; eine Mutation der Klasse `andere` startet einen Lauf | `TestRateNurFuerNebenlaeufig`: Fake mit blockierenden Starts; `andere` gibt genau zwei Starts (Stufe 1, Stufe 2); `nebenläufig` mit 7 und 3 gibt 7 plus 3 plus einen Start der Stufe 2, nie mehr als vier zugleich, ein Arm mit 0 entfällt, `flatternd` bei gemischten Exit-Codes | die Rate läuft auch für `andere`; der Deckel von vier zugleich fehlt (fünf Starts zugleich); ein Arm mit 0 startet doch einen Prozess | nebenläufig |
| Z7 | Der gemeinsame Cache verändert die Aussage nicht (Treffer und Fehlen des Caches geben dasselbe Ergebnis) | `TestCacheAendertAussageNicht`: Fixture, Liste mit zwei Mutationen derselben Datei von gleicher Länge hintereinander; Lauf mit `MUTATIONSLAUF_KALT=1` und warm, dieselben Ergebniszeilen außer `dauer`; der Test läuft aus dem Binary, nicht über `go test` | die Ablage des Binary hängt am Dateinamen statt an der Mutation: die zweite Mutation erbt das Ergebnis der ersten | andere |
| Z8 | Abbruch (Signal, Zeitgrenze) räumt Kopien und Prozesse auf; nach dem Lauf bleibt nichts zurück | `TestAbbruchRaeumtAuf` (Programm, Fake mit Prozessgruppen): Schleife über SIGTERM, SIGINT, Gesamtgrenze, hängender Test-Prozess (30 Sekunden Gnade mit künstlicher Uhr), Abbruch beim Anlegen einer Kopie; erwartet Exit 3 (bzw. `abbruch` bei Hänger und Kopie), `/work/lauf` leer, alle Prozessgruppen beendet, jede nicht gelaufene Mutation `abbruch`. `TestTreiberRaeumtAuf` (Treiber, Ersatz-Skript für `docker`): `docker rm -f -v` läuft nach Erfolg, Exit 2 und je Signal HUP, INT, TERM; der Start entfernt nur Container mit dem Label, die nicht laufen | die Kopie wird nach dem Ergebnis nicht entfernt; der Treiber ohne `trap`; der Start entfernt auch einen laufenden Container | nebenläufig |
| Z9 | Das Werkzeug ändert den Baum des Aufrufers nicht (Mutationen nur in Kopien) | `TestBaumUnveraendert`: Fixture, Prüfsumme des Baums vor und nach jeder Art aus Z2; `TestTreiberSchreibtNichts`: Ersatz-Skript für `docker`, der Baum ist schreibgeschützt (Modus), der Treiber legt außer dem Ablauf des Skripts keine Datei an | die Mutation wird im Eingangsbaum statt in der Kopie angewandt | andere |
| Z10 | Das Werkzeug läuft netzlos; ein schreibbarer Pfad ist nur der, den die Entscheidung des Users nennt (bis dahin keiner) | `TestTreiberOptionen`: Ersatz-Skript für `docker` hält jeden Aufruf fest; Schleife über alle Aufrufe: `create` trägt `--network none`, keiner trägt `-v`, `--volume`, `--mount` oder `--tmpfs` | `--network none` fehlt; ein Aufruf trägt `-v` | andere |
| Z11 | Die Mutationen eines Aufrufs laufen in einem Lauf des Containers, nicht als getrennte Jobs (eine Wartezeit am Lock je Aufruf, nicht je Mutation) | `TestEinContainerJeAufruf`: Ersatz-Skript für `docker`, Liste mit fünf Mutationen; genau ein `create` und ein `start` | der Treiber legt je Mutation einen Container an | andere |
| Z12 | Speicher und Platz sind begrenzt: jeder Container trägt `--memory` (≤ 4g, Mutationsläufe 2g), kein zweiter Go-Build zugleich, Wegwerfkopien, Binaries, eigene Container und Images sind nach dem Lauf weg, der Cache hat Größengrenze oder Aufräumweg (Zusatz des Users 2026-11-06) | `TestSpeicherUndGleichzeitigkeit`: Ersatz-Skript für `docker`: `create` trägt `--memory` und `--memory-swap` gleich dem Wert, Schleife über die Werte 1g, 2g, 4g (gültig) und 0g, 5g, `abc`, leer (Exit 2); Fake mit blockierenden Bauen: nie zwei Baue zugleich über alle Mutationen und Arme; nach dem Lauf weder `/work/lauf` noch ein Binary; Cache über der Grenze (Fake meldet 1,5 GiB plus ein Byte) ruft `go clean -cache`, auf der Grenze nicht | der Deckel auf einen Bau fehlt (zwei zugleich); `--memory` fehlt; die Prüfung der Cache-Größe fehlt; Grenze mit `>=` statt `>` | nebenläufig |
| Z13 | Die Liste wird vollständig geprüft, bevor etwas startet: alle Befunde stehen in der festgelegten Reihenfolge, bei einem Befund startet nichts | `TestListeBefunde`: Fake zählt Starts und Kopien (null); Schleife über jedes Feld mal (fehlt, leer, unbekannt, fremd) und über die Paare der Zusatztabelle (SPEC-210 Nr. 11); erwartet Exit 2, jeder Befund einmal, Reihenfolge nach Nummer der Mutation und Feld | die Prüfung bricht beim ersten Befund ab; die Prüfung läuft erst nach dem Anlegen der ersten Kopie; ein Paar liefert nur einen Befund | andere |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): nicht produktiv, blockiert keine produktive Linie und wartet auf keine. Er startet erst, wenn **beobachtbar** die Schlange am `flock /tmp/odata-recorder-heavy.lock` kurz ist (kein wartender Lauf einer produktiven Linie, wenn der Slice einen schweren Lauf braucht) **oder auf ausdrückliche Anweisung des Users**; kein Datum. Außerdem liegt kein anderer Slice der Gate-Kette in `in-progress/` (Übergangsregeln Nr. 7), falls der Architect für das Werkzeug eine Änderung dort braucht (Auftrag 7).
**Start am 2026-11-06 auf ausdrückliche Anweisung des Users** (nicht wegen kurzer Schlange). **Bedingung für den Architect:** `slice-test-determinism` liegt in `in-progress/` und ändert `harness/mk/` und die Sensors-Tabelle (Gate-Kette). Aus dem Plan ist keine Änderung der Gate-Kette durch dieses Werkzeug erkennbar (§3: `Makefile` und `harness/mk/` nicht geändert); der Slice läuft nur weiter, wenn der Architect in Auftrag 7 bestätigt, dass das Werkzeug **ohne** Ziel in `harness/mk/`, ohne Stufe im `Dockerfile`, ohne Eintrag in `GATE_CHECKS` und ohne Änderung der Sensors-Tabelle auskommt. Braucht er eine, gibt er an den Planner zurück (§6), und der Slice geht nach `open/`, bis `slice-test-determinism` geschlossen ist. `harness/README.md` (*Nicht behauptet*, Werkzeug-Zeile) ist ebenfalls Gate-Kette (Übergangsregeln Nr. 7): der Implementer ändert sie erst, wenn `slice-test-determinism` in `done/` liegt, oder der Planner ordnet die Änderung beim Merge. Die Architect-Aufträge 1 bis 8 sind entschieden, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).
**Bestätigt vom Architect am 2026-11-06 (Auftrag 7):** Das Werkzeug kommt ohne Ziel in `harness/mk/`, ohne Stufe im `Dockerfile`, ohne Eintrag in `GATE_CHECKS` und ohne Änderung der Sensors-Tabelle aus; der Slice läuft weiter. Die Zeile in `harness/README.md` bleibt beim Merge oder nach `slice-test-determinism` (oben).

**Größenmarke** (`AGENTS.md` §3.14): **Marke 3000, Schätzung 2155, Basis `9c9b406`** (der Commit des Architect, der SPEC-210, Tabelle und Raster vor dem Code abschließt). Es gilt die durchgesetzte Formel des Gates ([SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 1). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat 9c9b406..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen, `harness/vertraege/` ausgenommen. Der Planner hat `basis`, `schaetzung` und `marke` vor dem ersten Code-Commit im eigenen Commit gesetzt; eine ältere Marke gab es nicht (kein Eintrag `ersetzt`). **Der Formelkonflikt bleibt beim User, der Planner löst ihn nicht:** nach `AGENTS.md` §3.14 wäre die Marke 1600. Aufwand nach Einschätzung des Planners: S bis M (Programm mit Prozessen und Cache, nebenläufig).
**Schätzung des Architect nach beiden Formeln, sie widersprechen sich** (Zeilen der Tabelle: 3 `nebenläufig` — Z6, Z8, Z12 —, 10 `andere`; drei Fakes: Prozess-Port, Ersatz-Skript für `docker`, Fixture-Modul):

- **Nach SPEC-128 Nr. 1** (das Gate, `make groessenmarke`): *S* = 85·3 + 80·10 + 200·3 + 500 = **2155**, Marke 100 · ⌈1,35·2155⌉ = **3000**; keine Schnittfrage (*S* < 2223). `produktcode: 500` ist das Programm (rund 420 Zeilen Go) und der Treiber (rund 80 Zeilen Shell).
- **Nach `AGENTS.md` §3.14**: 85·3 + 40·10 + 200·3 = **1255**, kein Leser fremder Bytes (die Liste ist ein Format dieses Werkzeugs, kein Fremdformat), also kein fester Anteil; × 1,25 = 1569, Marke **1600**.
- **Eigene Erwartung des Architect** für den Stand beim Handoff: rund 1500 bis 1800 Zeilen (Programm und Treiber rund 500, Tests mit den drei Fakes und der Fixture rund 1000 bis 1300); nach dem Review rund 20 % mehr. Sie liegt zwischen den Formeln, näher an der ersten.
- **Der Formelkonflikt** (80 gegen 40 je Zeile der Klasse `andere`, ×1,35 gegen ×1,25, `produktcode` nur in `SPEC-128`) ist eine Entscheidung des Users oder des Planners; der Architect setzt nichts um. Bis dahin gilt das Gate (`SPEC-128`): 3000.

```groessenmarke
pfad: tools/mutationslauf
pfad: tools/harness/mutationslauf.sh
produktcode: 500
fakes: 3
basis: 9c9b406
schaetzung: 2155
marke: 3000
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Schätzung des Architect mal 1,25 liegt über 3000 (`AGENTS.md` §3.14), oder die Messung nach dem Code liegt über der Marke (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Ein Schnitt ist vorab benannt: der Ansatzpunkt (c), der gemeinsame Cache über die Rollen, teilt keinen Mechanismus mit (a), (b) und (d) und lässt sich als eigener Teil zuletzt schreiben (er hängt an der Entscheidung des Users zum schreibbaren Mount). **Der Architect hat (c) ohne Mount entschieden** (Cache im Image, SPEC-210 Nr. 8): der Teil ist dann klein (Größengrenze, `MUTATIONSLAUF_KALT`) und der Schnitt entfällt, solange der User keinen Mount freigibt.
- `in-progress` → `open` (blockiert): der User gibt den schreibbaren Pfad für den Cache nicht frei **und** der Architect findet keinen Weg ohne ihn, der den Cache über die Rollen trägt (dann Slice ohne (c)); oder das Werkzeug braucht eine Änderung der Gate-Kette, während ein anderer Slice der Linie C in `in-progress/` liegt. **Beide Bedingungen sind nach dem Stand vom 2026-11-06 nicht eingetreten:** der Architect hat einen Weg ohne Mount (Cache im Image) und das Werkzeug braucht keine Änderung der Gate-Kette.

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

Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* steht im Vertrag ([SPEC-210](../../../../harness/vertraege/SPEC-210-mutationslauf.md) Nr. 11, mit Zusatztabelle); jede Zelle hier nennt die Nummer des Vertrags, die sie entscheidet, oder „gilt nicht“ mit Grund. Eine Zelle ohne Ort hält den Implementer an. Die Zusatztabelle (Paare von Verstößen am selben Schlüssel, Klasse, Cache) steht im Vertrag; ihr Fänger ist Z13 (Paare) und Z6, Z7 (Modi). **Ausgang:** entschieden vor dem Code (Architect, 2026-11-06).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Liste lesen | SPEC-210 Nr. 2 a (Pflichtfeld), 2 c (`mutationen`) | Nr. 2 b, 2 c (leere Liste, leerer Name, leeres `alt`; `neu` leer erlaubt) | Nr. 2 a, 2 b, 2 c (Klasse, Feld, Name doppelt, Datei außerhalb oder nicht `.go`, kein JSON) | gilt nicht: Lesen ändert nichts, der Lauf beginnt erst nach der Prüfung (Nr. 2 d) | Nr. 2 e (Z2) |
| Mutation anwenden | Nr. 2 b (Datei fehlt im Baum) | Nr. 2 b (`neu` leer = Löschung; `neu` gleich `alt` = Befund) | Nr. 2 b (`alt` trifft nicht oder mehrfach) | Nr. 10 e (Z8) | Nr. 4 b (dieselbe Mutation unter zwei Namen) |
| Test fahren (`-run`, Binary) | Nr. 2 a, 5 a (Z4) | Nr. 5 a (Paket ohne Testdatei, null Treffer) | Nr. 5 a (mehrere Treffer, Paket ohne Tests), 5 d (Übersetzungsfehler, Z3) | Nr. 6 b, 9 c, 10 a (Z8) | Nr. 6 d, 7 c (Z6) |
| Rate fahren (nebenläufig) | Nr. 2 b (Vorgabe 50) | Nr. 2 b, 7 b (Zahl 0: Arm entfällt; beide 0: Befund) | Nr. 2 b (über 200); Speicher: Nr. 9 b (höchstens vier zugleich) | Nr. 10 a (Z8) | gilt nicht: die Rate ist selbst die Wiederholung |
| Cache nutzen | Nr. 8 b, 8 c (Z7) | Nr. 8 c (Z7) | Nr. 3 a (veraltete Module); fremder Cache: gilt nicht, Go bildet den Schlüssel aus dem Inhalt (Nr. 8 b) | gilt nicht: der Cache stirbt mit dem Container (Nr. 8 a) | Nr. 8 a, 8 b (Z7) |
| Ergebnis ausgeben | Nr. 10 a (`abbruch`) | Nr. 6 b (leere Ausgabe, Exit-Code zählt) | Nr. 6 b, 7 d (nur der Exit-Code wird gewertet) | Nr. 10 a (Z8) | gilt nicht: die Ausgabe ändert nichts |
| Aufräumen | Nr. 10 c | Nr. 4 b | Nr. 10 e | Nr. 10 a | Nr. 10 c (Z8) |

Risiken und offene Punkte:

- **Gleichwertigkeit mit dem bisherigen Ergebnis je Mutation.** `-run` auf den zuständigen Test sieht ein Rot eines anderen Tests im Paket nicht, das der Lauf über das Paket zeigt; ein Test-Binary je Mutation und ein gemeinsamer Cache können dieselbe Aussage verschieben. Das Werkzeug darf kein Grün liefern, wo der Einzellauf rot wäre — **Ausgang:** entschieden (SPEC-210 Nr. 5, 6: zuständiger Test plus Paket als Stufe 2); Nachweis Z2, der Verifier löst die Klasse der Arten auf (`AGENTS.md` §3.17). **Akzeptiertes Negativ mit Grund und Folge:** `go vet` läuft nicht (SPEC-210 Nr. 5 e, 12 a) — Grund: die Mutationen der Fänger sind Verhaltens-Mutationen; Folge: eine Mutation, die nur Vet träfe, steht als `gruen`, die Rolle hält sie nie für gefangen, die sie nicht fängt.
- **Cache-Einfluss auf die Aussage.** Ein Treffer im Testcache von `go test` (ohne `-count=1`) wäre ein falsches Grün; der Cache der Übersetzung ändert die Aussage nicht, aber das Werkzeug muss das zusagen und belegen — **Ausgang:** entschieden (SPEC-210 Nr. 8 b: der Test läuft aus dem Binary, nie über `go test`, ein Testergebnis wird nicht zwischengespeichert); Nachweis Z7.
- **Die Rate der Fenster-Mutationen bleibt ein Wirkungs-Fänger-Thema.** `AGENTS.md` §3.10 verlangt für Nebenläufigkeit „in jedem Lauf rot“; das Werkzeug macht den Lauf billiger, ersetzt aber den Fänger nicht, der die Wirkung zählt (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`). Wer die Rate nur für Fenster-Mutationen fährt, verlässt sich darauf, dass die Rolle die Klasse richtig setzt (setzt sie `andere` für eine nebenläufige Zusage, fehlt die Rate, und das Werkzeug merkt es nicht) — **Ausgang:** weiter offen → Register; das akzeptierte Negativ steht im Vertrag (SPEC-210 Nr. 12 b: Grund, das Werkzeug glaubt der Liste; Folge, der Reviewer prüft die Klasse gegen die Zeile der Tabelle) (`BEO-REPO/akzeptiertes-negativ-ohne-folge`).
- **Docker-only und kein schreibbarer Mount (`AGENTS.md` §3.1) — Ansatzpunkt (c), Entscheidung des Users.** Ein Cache über die Rollen braucht einen Ort, der den Lauf überlebt; die Läufe sind netzlos und ohne schreibbaren Mount. **Entscheidung des Users**, nicht des Planners und nicht des Architect: ob ein schreibbarer Pfad für den Cache erlaubt ist. Die Optionen, mit Befund (gemessen am 2026-11-06: das Image `odata-recorder:race` trägt 292 MB Go-Cache aus seinem Bau, `odata-recorder:test` 442 MB; Platte des Docker-Hosts 90 % belegt, 30 GB frei):
  1. **Cache nur im Container, warm aus dem Image** (Vorgabe im Vertrag, SPEC-210 Nr. 8): kein Mount, keine Ausnahme von §3.1, alle Rollen teilen den Stand des Images; verloren geht, was ein Lauf neu übersetzt (das mutierte Paket, bei `-race` und ohne Detektor je einmal je Mutation; bei einer Liste mit `nebenläufig` auf dem Race-Image auch alle Pakete ohne Detektor einmal je Aufruf). Aufwand für den Nachzug: null.
  2. **Volume** (`-v <Name>:/root/.cache/go-build`): trägt den Cache über Läufe und Rollen, auch den ohne Detektor; ist die Ausnahme von §3.1 (schreibbarer Mount), verbindet die Rollen durch gemeinsamen Zustand (Unabhängigkeit der Rollen, `AGENTS.md` §3.10: der Cache hält nur Übersetzungen nach Inhalt, kein Ergebnis, ein Vergiftungspfad bleibt trotzdem) und wächst ohne Grenze, wenn kein Räumer läuft (der Platz des Docker-Hosts ist knapp). Braucht einen Folge-ADR (Bereich 0030 bis 0039) und eine Größengrenze im Vertrag.
  3. **`docker cp` in und aus dem Container** (wie bei `schema-*`): kein Mount, aber eine persistente Ablage auf dem Host (die Ausnahme verschiebt sich dorthin) und 300 bis 450 MB Kopie je Lauf in beide Richtungen; ob das billiger ist als das Neuübersetzen, ist ungemessen.
  
  **Empfehlung des Architect: Option 1 jetzt, Option 2 nur, wenn die Messung (§5) zeigt, dass der Verlust trägt** (Zeit des Neuübersetzens je Aufruf spürbar, gemessen mit `MUTATIONSLAUF_KALT=1` gegen warm). Grund: der teuerste Teil (Fremdpakete wie `modernc.org/sqlite`) steckt schon im Image, und Option 1 verlangt weder eine Ausnahme noch eine ADR noch einen Räumer. **Ausgang:** offen, Adresse User. Ohne Freigabe trägt alles (Option 1); mit Freigabe von Option 2 ist es ein Folge-ADR und ein kleiner Zusatz im Vertrag (Nr. 8 e), kein Neuschnitt.
- **Speicher und Platz sind Anforderung an das Werkzeug (Zusatz des Users, 2026-11-06; Host 31 GB RAM, 17 GB verfügbar, ein fremder Container läuft nebenher).** Das Werkzeug setzt `--memory` je Lauf selbst (Vorgabe 2g, bis 4g einstellbar), fährt nicht mehrere Go-Builds zugleich und räumt Wegwerfkopien, Test-Binaries und Container nach dem Lauf, auch nach Abbruch (Z8); es erzeugt kein Image; der Cache hat die Größengrenze 1,5 GiB — **Ausgang:** entschieden (SPEC-210 Nr. 9, 10, 8 d; Z8, Z12). **Offen:** ob 2g für den Bau von `modernc.org/sqlite` mit `-race` und die Bau-Spitze (`-p 2`, `GOMEMLIMIT`) reicht, ist nicht gemessen — **Ausgang:** offen, Adresse Implementer (misst die Spitze des Speichers der Nullprobe mit `docker stats` oder `memory.peak` unter `flock`; reicht 2g nicht, geht der Wert als Frage an den Architect zurück, der Code erhöht ihn nicht).
- **Ein Mutations-Protokoll mit Prüfsummen, das Ergebnisse zwischen Rollen wiederverwendet** (gleicher Baum, gleiche Mutation, gleiche Quellen: gleiches Ergebnis ohne Neulauf), würde Läufe sparen und die Unabhängigkeit der Rollen aufheben (`AGENTS.md` §3.10, §3.15: Reviewer und Verifier messen weiter selbst). Das ist eine Entscheidung des Users und des Architect über die Rollen, kein Werkzeug-Detail; der Slice baut es nicht (SPEC-210 Nr. 12 e) — **Ausgang:** offen, Adresse Architect und User; eine Entscheidung dafür wäre ein eigener Slice.
- **Das Werkzeug braucht eine Änderung der Gate-Kette** (ein Ziel in `harness/mk/`, eine Stufe im `Dockerfile`, einen Eintrag in `GATE_CHECKS`) und kollidiert mit einem anderen Slice der Linie C (Übergangsregeln Nr. 7) — **Ausgang:** entfallen: der Architect bestätigt in Auftrag 7, dass es ohne auskommt (§3, §4).
- **Das Modul `tools/mutationslauf` hat kein Gate** (SPEC-210 Nr. 12 d): `make gates` fährt weder `gofmt` noch `go vet` noch `golangci-lint` noch seine Tests, weil eine Stufe im `Dockerfile` und ein Glied von `GATE_CHECKS` dazu nötig wären. Der Implementer fährt sie selbst im Image `odata-recorder:race` unter `flock` und berichtet sie; eine Regression danach fängt erst ein späterer Lauf — **Ausgang:** weiter offen → Adresse Planner: nach dem Schluss von `slice-test-determinism` entscheidet er, ob das Modul eine Stufe und ein Gate bekommt (Änderung der Gate-Kette, dann mit ADR-Prüfung nach `AGENTS.md` §3.6).
- **Das Image kommt aus der Kette:** Das Werkzeug läuft nur, wenn `make test-race` (oder `make test`) das Image gebaut hat und `go.mod`/`go.sum` zum Baum passen (SPEC-210 Nr. 3 a) — **Ausgang:** entfallen als Risiko, akzeptiertes Negativ: Grund, ein eigenes Image verlangte eine Stufe (Gate-Kette); Folge, ein veralteter Stand endet mit Exit 2 und einer Meldung statt mit einem falschen Ergebnis.
- **Der Formelkonflikt der Größenmarke** (`SPEC-128` gegen `AGENTS.md` §3.14, siehe §4: 3000 gegen 1600) — **Ausgang:** offen, Adresse User oder Planner; der Architect setzt nichts um.
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
