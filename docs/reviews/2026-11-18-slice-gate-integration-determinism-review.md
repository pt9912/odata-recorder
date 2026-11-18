# Review-Report: slice-gate-integration-determinism — 2026-11-18

**Review-Art:** Code (gegen Plan, Vertraege `SPEC-211`, `SPEC-037`, `SPEC-038`, `SPEC-042` und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers)

**Gegenstand:** `slice-gate-integration-determinism`, Zweig `slice-gate-integration-determinism`, `b130016e..78417512`: Start-Move, Architect (`SPEC-211`, Nachzuege in `SPEC-037`, `SPEC-038`, `SPEC-042`, Index), Planner (Groessenmarke), Implementer (`harness/mk/test-integration.mk`, `harness/mk/test-determinism.mk`, `harness/mk/test-contract.mk`, `harness/README.md`, Plan §3 und §6).

**Skill:** `.harness/skills/reviewer.md` @ `78417512` <!-- d-check:ignore (Zitat des Standes im Lauf) -->
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

**Eingangs-Kontext** (die Vertraege, gegen die geprueft wurde):

- Slice-Plan `slice-gate-integration-determinism` (DoD §2, §3 Tabelle Zusage und Faenger Nr. 1 bis 12 mit Nachweis, Raster §6)
- `SPEC-211`, `SPEC-037` Nr. 2 und 10 f, `SPEC-038` Nr. 2 und 8, `SPEC-042` Nr. 2 und 9 f; ADR-0005
- `LH-QA-06`
- `AGENTS.md` §3.6, §3.7, §3.9 bis §3.18; Uebergangsregeln `docs/plan/planning/README.md` Nr. 1 bis 8

**Vorgehen und Belege.** Gelesen: der gesamte Diff. Gefahren unter `flock /home/builder/work/odata-recorder/.wt/heavy.lock`; der Zweig wurde nicht veraendert (`git status --short` leer vor und nach dem Lauf). Mutationen liefen in Kopien (`git archive HEAD` im Scratchpad), weil dieser Lauf keinen Editor fuer Repo-Dateien hat und `sed -i` auf Repo-Dateien verboten ist; die Kopien nutzen dieselben Image-Tags. Eine Mutation (`t.Fatal`, Nr. 2) hat die Tags `odata-recorder:integration` und `odata-recorder:contract` kurz mit Mutanten-Inhalt ueberschrieben; vor den folgenden Messungen wurden die Images aus dem Zweig neu gebaut (`make -j3 ... ` gruen). Messender: der Reviewer, Wanduhr mit `date`. Last des Hosts (20 Kerne, fremde Last anderer Sitzungen): 8 bis 15 bei den Raten von Nr. 3, 15 bis 27 bei den Handlaeufen von Nr. 4; alle Zahlen sind Groessenordnungen, kein Detektor.

Gegengelesen auf Rohbytes (`LC_ALL=C grep -P` auf Nicht-ASCII-Zeilen, `\\u[0-9A-Fa-f]{4}`, BOM, NBSP, Steuerzeichen): die drei Fragmente, `harness/README.md`, `SPEC-211`, `SPEC-037`, `SPEC-038`, `SPEC-042`, der Plan: Nicht-ASCII nur Umlaute und typografische Zeichen des Bestands, kein Treffer fuer die Escape-Folge, BOM, NBSP oder Steuerzeichen.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `SPEC-211` Nr. 2 c sagt, INT oder TERM "an `make` oder an die Prozessgruppe" erreiche jeden Treiber, jeder ende mit Exit 2. Gemessen: `kill -TERM <make>` beendet `make` mit Exit 143 nach 4 bis 5 ms, der Treiber (`sh tools/harness/test-<Ziel>.sh`) wird nicht erreicht und laeuft weiter (`test-integration` und `test-contract` fahren ihre Stufen zu Ende, `test-determinism` den Container bis zu seinem natuerlichen Ende nach rund 0,9 s). Der Faenger (Plan §3 Nr. 4) signalisiert nur den Treiber selbst; die Haelfte "an make" hat keinen Faenger und trifft nicht zu. | `AGENTS.md` §3.15, §3.11 | `harness/vertraege/SPEC-211-container-tests-in-der-kette.md` · "INT oder TERM an `make` oder an die Prozessgruppe erreicht jeden Treiber" | ja - Handlauf: `make test-determinism &`, `kill -TERM` an `make`, `pgrep -af tools/harness/test-` danach (in diesem Lauf fuer alle drei Ziele reproduziert) | Zusage ueber Signalweg reicht weiter als der Faenger |
| F-2 | MEDIUM | Die Container-Haelfte von `SPEC-211` Nr. 2 c ("`--rm` entfernt den Container") ist nur fuer `test-determinism` unterscheidbar (Container weg nach 259 bis 387 ms mit `trap`, 1079 bis 1498 ms ohne, je 4 von 4); fuer `test-integration` (83 bis 228 ms gegen 125 bis 191 ms) und `test-contract` (122 bis 132 ms gegen 74 bis 130 ms) bleibt die Mutation "`trap` entfernt" in dieser Haelfte unsichtbar. Das akzeptierte Negativ mit Grund und Folge steht im Plan §3, im Vertrag steht es nicht; §6 traegt die Rueckgabe mit Ausgang *offen* und Adresse Architect, wie §3.13 es verlangt. Der Vertragstext sagt weiter "jeder Treiber ... Container entfernt". | `AGENTS.md` §3.15, §3.11 | `harness/vertraege/SPEC-211-container-tests-in-der-kette.md` · "sein laufender `docker run` endet per Signalweitergabe, `--rm` entfernt den Container" | ja - Handlauf je Treiber, Exit-Haelfte gefangen: Exit 2 gegen 143 in 12 von 12 Laeufen (3 Treiber mal 4) | Negativ nur im Plan, nicht im Vertrag |
| F-3 | MEDIUM | `SPEC-211` Nr. 3 Zeilen 1 bis 3 (`docker` fehlt, Stufe scheitert beim Bau, Image fehlt nach dem Bau) tragen Verhalten (Exit 2, Meldung `Aufbau der Stufe ... gescheitert`), aber Nr. 5 fuehrt dafuer kein akzeptiertes Negativ mit Grund und Folge; es steht nur im Plan §3 Nr. 5 bis 7 als "kein automatischer Fall". Gemessen: Mutation `if false && fahre ... build` laesst den Exit 2 unveraendert (3 von 3 Zielen, `DOCKER=/bin/false`) und nimmt nur die Meldung; mit Syntaxfehler im Test plus derselben Mutation ist der Lauf Exit 0 und gruen in 3 von 3 Zielen (Kontrolle ohne Mutation: Exit 2 mit Meldung in 3 von 3). Nur der Handlauf faengt es. | `AGENTS.md` §3.15 | `harness/vertraege/SPEC-211-container-tests-in-der-kette.md` · "der Bau scheitert: die Meldung von `docker` auf stderr, danach `<Ziel>: Aufbau der Stufe <Stufe> gescheitert`, Exit 2" | ja - Handlauf `DOCKER=/bin/false make <Ziel>`; Gate-Lauf faengt es nicht | Zusage ohne Faenger im Vertrag, Negativ nur im Plan |
| F-4 | LOW | `SPEC-211` Nr. 2 c sagt, nach dem Scheitern eines Checks unter `-j` enden die Treiber "nach hoechstens einer Frist von 120 s je Stufe". Ein laufender Treiber setzt seine uebrigen Stufen fort (`test-integration` neun, `test-contract` zwei Binaries; beobachtet: der Treiber lief nach dem Ende von `make` alle Stufen zu Ende), die Obergrenze ist die Summe der Fristen, nicht eine. Nicht mit einer haengenden Stufe gemessen. | `AGENTS.md` §3.7 (Zusage) | `harness/vertraege/SPEC-211-container-tests-in-der-kette.md` · "die Treiber enden nach hoechstens einer Frist von 120 s je Stufe" | nein - ein Lauf mit haengender Stufe ist kein Fall der Kette | Zeitgrenze des Wartens schmaler als die Zahl der Stufen |
| F-5 | LOW | Die Zusage `SPEC-211` Nr. 2 a (zwei Bauten desselben Tags liefern denselben Inhalt, ein spaeteres Setzen des Tags aendert nichts an einem laufenden Container) hat in Plan §3 Nr. 3 nur die Mutation "fester `--name`" (Namenskonflikt); fuer die Tag-Haelfte gibt es keine Mutation und kein als Negativ gefuehrtes Gegenstueck (Grund: der Inhalt ist durch das `Dockerfile` gleich, nichts im Treiber kann ihn ungleich machen; Folge steht in Nr. 2 a). Der Gutfall ist 10 von 10 gruen. | `AGENTS.md` §3.15 | `docs/plan/planning/in-progress/slice-gate-integration-determinism.md` · "Die Laeufe zugleich ueberschreiben einander nicht und kollidieren nicht" | nein | Zusage mit zwei Haelften, ein Faenger |
| F-6 | INFO | Die Mutation "fester `--name`" nennt `SPEC-211` Nr. 2 e "in `laufe` von `test-integration.sh` und `test-determinism.sh`"; `test-determinism.sh` hat kein `laufe`, die Stelle ist der `docker run` im Skript. Der Plan §6 fuehrt das mit Ausgang *offen* und Adresse Architect (Wortlaut, keine Entscheidung). | `AGENTS.md` §3.13 | `harness/vertraege/SPEC-211-container-tests-in-der-kette.md` · "in `laufe` von `test-integration.sh` und `test-determinism.sh`" | nein | Wortlaut-Abweichung, bereits adressiert |
| F-7 | INFO | Nicht selbst gefahren: Plan §3 Nr. 11 (Zeile entfernt, Zelle mit 201 Zeichen). Eine Kopie des Baums ausserhalb des Repos ist fuer `make docs-check` nicht mountbar (`Scan-Wurzel ist leer`), und `sed -i` auf Repo-Dateien ist verboten. Gemessen statt dessen: `make docs-check` im Zweig meldet genau die fuenf bekannten Befunde der Roadmap; die Zellen der drei neuen Zeilen haben 93, 99 und 97 Zeichen (Spalte Vertrag), 25 und 22 (Target), 62 (Bindung), alle unter 200. | Maintainability | `harness/README.md` · "Binary im Container ohne Netz: Record, Upstream stoppen, Replay in neun Stufen" | ja - `make docs-check` | nicht gefahren: Nr. 11 |

---

## Mutationstabelle (Zeile des Plans §3, eigene Mutation, eigenes Ergebnis)

| Nr. | Eigene Mutation | Ergebnis |
|---|---|---|
| 1 | Zeile `GATE_CHECKS += <Ziel>` entfernt, je Ziel (Kopie) | `make -n -k gates`: Treffer des Treibers 3 ohne Mutation, 2 mit Mutation in 3 von 3 Zielen. Ohne Mutation steht je ein Rezept `sh tools/harness/test-contract.sh`, `test-determinism.sh`, `test-integration.sh`. Rot als Handprobe, kein Gate (`SPEC-211` Nr. 5 a). |
| 2 | `t.Fatal("mutation")` am Kopf von `TestE2EProzess`, `TestE2EDeterminismus`, `TestVertragFehlerklassen` zugleich (Kopie), `make -k record-gates GATE_CHECKS="test-integration test-determinism test-contract"` | alle drei rot aus dem benannten Grund (`stufe_prozess_test.go:14: mutation`, `determinismus_test.go:237: mutation`, `vertrag_fehlerklassen_test.go:309: mutation`); `record-gates` nicht gelaufen ("not remade because of errors"), kein Stempel. Geprueft ist die Kette der drei Ziele, nicht die volle `make gates`-Kette (Last). |
| 3 | fester `--name e2e-fest` in `laufe` von `test-integration.sh` und am `docker run` von `test-determinism.sh` (Kopie), `make -j3 test-integration test-determinism test-contract` | Kontrolle ohne Mutation: 0 von 10 rot (Last 11,2 bis 14,7, 1 Minute beim Start). Mit Mutation: **10 von 10 rot**, Exit 2, in jedem der zehn Logs genau einmal `is already in use` (Last 8,1 bis 10,4). Messender: der Reviewer; Prozesse: drei Treiber je Lauf, 10 Laeufe. Der Faenger ist in jedem gemessenen Lauf rot; zehn Laeufe sind eine Stichprobe, kein Beweis fuer jeden Lauf auf ruhigem Host. |
| 4 | `trap abbruch INT TERM` ersetzt durch einen Kommentar, je Treiber (Kopie); Treiber direkt gestartet, `kill -TERM` an den Treiber, sobald ein neuer Container laeuft | Exit-Haelfte: Kontrolle Exit 2 in 12 von 12, Mutation Exit 143 in 12 von 12 (je 4 pro Treiber). Container-Haelfte: nur `test-determinism` trennt (259 bis 387 ms gegen 1079 bis 1498 ms, 4 von 4 je Seite), `test-integration` und `test-contract` nicht (siehe F-2). Last 15 bis 27. Zusaetzlich `kill -TERM` an `make` allein: Treiber nicht erreicht (F-1). |
| 5 | `if ! fahre ... build` durch `if false && fahre ...` ersetzt (Kopie), `DOCKER=/bin/false make <Ziel>` | Kontrolle: Exit 2 und Meldung in 3 von 3. Mutation: Exit 2 in 3 von 3, Meldung fehlt in 3 von 3; der Faenger ist die Meldung, nicht der Exit (wie im Plan). |
| 6 | dieselbe Mutation plus Syntaxfehler (`x x x`) in `tests/e2e/stufe_prozess_test.go` bzw. `src/bootstrap/abnahme_etag_test.go` (Kopie) | Kontrolle: Exit 2 mit `Aufbau der Stufe ... gescheitert` in 3 von 3. Mutation: Exit 0, gruen in 3 von 3. |
| 7 | - | akzeptiertes Negativ, nur im Plan (F-3). |
| 8, 9 | `-test.timeout 120s` und `--memory 2g` aus allen drei Treibern entfernt (Kopie) | `make test-integration test-determinism test-contract` Exit 0: kein Faenger, wie `SPEC-211` Nr. 3 sagt; Negativ bestaetigt. |
| 10 | - | kein Pruefer, Soll mit Messvorschrift (`SPEC-211` Nr. 5 c); nicht neu gemessen. |
| 11 | siehe F-7 | nicht selbst gefahren; Ist-Stand `make docs-check`: fuenf Befunde (Roadmap), keiner weiter. |
| 12 | `grep -rn 'hängt nicht an .GATE_CHECKS\|KEIN GATE\|KEIN Gate' harness tools/harness docs/plan/planning/in-progress` | kein Treffer zu den drei Zielen; Treffer nur zu `traeger.mk`, `e2e-abdeckung.mk`, `selbstpruefung.mk`, `slice-mv.mk`, `archivierung.mk`, `doc-gate.mk` (andere Ziele, richtig). |

`make -n -k gates` nennt alle drei Treiber (je ein Rezept, Reihenfolge `test-contract`, `test-determinism`, `test-integration`); `make -n gates` allein bricht im Zweig am Rezept von `schema-check`, wie der Plan sagt. `make help` nennt die drei Ziele mit "Gate:" und ohne "KEIN Gate"; die Kopf-Kommentare der drei Fragmente nennen nur, was gilt (Zeile haengt an `GATE_CHECKS`, Stufe, Container je Stufe bzw. je Binary) und dass kein Fall die Zeile faengt (`SPEC-211` Nr. 5 a).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `harness/mk/test-integration.mk`, `harness/mk/test-determinism.mk`, `harness/mk/test-contract.mk` | geprueft, ohne Befund: Zeile `GATE_CHECKS += <Ziel>` je Datei, Kopf- und `make help`-Text stimmen mit dem Treiber (Stufe, Container je Stufe bzw. je Binary), kein Satz zu "KEIN GATE" |
| `harness/README.md` | geprueft, ohne Befund: drei Zeilen in der Gate-Tabelle mit der Bindung des Vertrags, Zellen unter 200 Zeichen, "Nicht behauptet" nennt nur Dinge, die `SPEC-211` Nr. 2 d, 3, 5 a bis c benennt; keine nicht vorhandene Target; keine Zeile mehr unter *Werkzeuge* |
| `harness/vertraege/SPEC-037`, `SPEC-038`, `SPEC-042`, `harness/vertraege/README.md` | geprueft, ohne Befund ausser F-1 bis F-3 (die Nachzuege tragen "haengt an `GATE_CHECKS`", die Image-Tags stimmen mit `IMAGE=` der Treiber ueberein: `integration` zweimal, `contract` einmal) |
| `SPEC-211` Nr. 1, 4, 5 | geprueft, ohne Befund: Entscheidung ohne ADR mit Alternativen und Folge (`AGENTS.md` §3.6: Verschaerfung, keine Senkung), Zeitbudget als Soll mit Messvorschrift und Trigger, Negative mit Grund und Folge |
| Raster Plan §6 | geprueft, ohne Befund: jede Zelle nennt einen Ort; die Zeilennummern von `SPEC-211` Nr. 3 (1 bis 10) stimmen mit der Tabelle; `SPEC-038` Nr. 8 c, d, e, `SPEC-042` Nr. 8 c, f und `SPEC-037` Nr. 10 g tragen die zitierte Aussage; die Zusatztabelle nennt je Modus einen Faenger oder das Negativ (die Haelften des Faengers fuer Abbruch siehe F-1, F-2) |
| Groessenmarke | geprueft, ohne Befund: Block in §4 vom Planner vor dem Code, `make groessenmarke`: "ok, Messung 35 von Marke 1500"; `git diff --numstat` ab Basis `2b885889` ueber die Pfade des Blocks: 35 (README 4+4, drei Fragmente je 5+4) |
| Plan Kopf, §1, §3, §6 nach der Berichtigung (`AGENTS.md` §3.9) | geprueft, ohne Befund: die beiden Rueckgaben (Container-Haelfte, Wortlaut `laufe`) stehen als Zeilen mit Ausgang *offen* und Adresse in §6 (§3.13), Roadmap-Befunde mit Adresse Planner |
| `make docs-check` | geprueft: genau die fuenf bekannten Befunde in `roadmap.md`, kein weiterer |
| Rohbytes (§3.18) | geprueft, ohne Befund: siehe Vorgehen |
| Arbeitsbaum nach dem Lauf | `git status --short` leer; kein Container `e2e-fest` uebrig |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 2 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Zusage ueber Signalweg reicht weiter als der Faenger · Negativ nur im Plan, nicht im Vertrag · Zusage ohne Faenger im Vertrag, Negativ nur im Plan · Zeitgrenze des Wartens schmaler als die Zahl der Stufen · Zusage mit zwei Haelften, ein Faenger

## Verdikt

**Merge-blockierend:** ja - F-1 bis F-3 sind MEDIUM; sie liegen im Vertrag `SPEC-211` (Architect), nicht im Code der drei Fragmente. Die Aenderung selbst (drei Zeilen `GATE_CHECKS`, Texte, `harness/README.md`) ist unauffaellig, und die Mutationen der Zeilen 1, 2, 3, 5, 6, 8 und 9 verhalten sich wie im Plan berichtet.

**Uebergabe:** F-1 bis F-5 an den Architect (Vertragstext `SPEC-211` Nr. 2 c, Nr. 3, Nr. 5; F-2 ist in Plan §6 schon als offen mit Adresse Architect gefuehrt, F-1 und F-3 sind neu und gehoeren im selben Commit wie die Antwort in §6 eingetragen); der Implementer zieht den Plan nach. Die Finding-Klassen gehen in die Closure §7. Dieser Report ersetzt keine Verifikation gegen die DoD.
