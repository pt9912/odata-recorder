# Verifikations-Report: slice-gate-integration-determinism - 2026-11-18

**Art:** Verifikation (Modul 11) gegen DoD, Plan, Vertrag `SPEC-211` samt Nachzuegen und Hard Rules; nicht das Review (Diff gegen Plan) und nicht die Validierung.

**Gegenstand:** Zweig `slice-gate-integration-determinism`, HEAD `5904c6db`, Basis `main` = `b130016e`; `git diff b130016e..HEAD`. Arbeitsbaum vor und nach dem Lauf: `git status --short` leer. Nichts im Zweig geaendert; der einzige neue Pfad ist dieser Report (nicht committet).

**Eingang:** Plan `docs/plan/planning/in-progress/slice-gate-integration-determinism.md` (DoD §2, §3 Tabelle und Nachweis, §4, §6), `harness/vertraege/SPEC-211-container-tests-in-der-kette.md`, Nachzuege `SPEC-037`, `SPEC-038`, `SPEC-042`, `harness/vertraege/README.md`, `harness/README.md`, Review-Report `docs/reviews/2026-11-18-slice-gate-integration-determinism-review.md`, `AGENTS.md` §3.7 bis §3.18.

**Messbedingungen.** Host mit 20 Kernen, Last fremder Sitzungen: 6 bis 8 (Load Average 1 Minute) bei Beginn; waehrend der zweiten Haelfte der Rate-Messung stieg sie auf 30 bis 205 (nicht von diesem Lauf erzeugt, die Container sind 2 GiB, drei zugleich). Messender: der Verifier, ohne Detektor, Zahlen sind Groessenordnungen. Schwere Laeufe unter `flock /home/builder/work/odata-recorder/.wt/heavy.lock`, eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`. Mutationen liefen in einer Kopie (`git archive HEAD` nach `/home/builder/work/odata-recorder/.wt/verif-copy`, danach geloescht); die Kopie nutzt dieselben Image-Tags aus demselben Quellstand.

---

## 1. Verdikt je DoD-Punkt (§2, in Reihenfolge)

| # | DoD-Punkt | Verdikt | Beleg |
|---|---|---|---|
| 1 | `GATE_CHECKS += <Ziel>` je Datei, Kopf und `make help` ohne "KEIN GATE", `make gates` faehrt alle drei und ist gruen | **Teil 1 erfuellt; "und ist gruen" nicht erfuellt im Zweig, erst nach Integrationsschritt erfuellbar** | `harness/mk/test-{integration,determinism,contract}.mk` tragen die Zeile; `make help` nennt "Gate:"; `grep -rn 'KEIN Gate\|KEIN GATE\|haengt nicht an .GATE_CHECKS' harness tools/harness` trifft nur andere Ziele (`traeger`, `archivierung`, `ai-harness-init.md` u. a.). `make -n -k gates`: 3 Treffer des Treibers (Zeilen 36 bis 38), mit entfernter Zeile je Ziel 2 (3 von 3). `make -k gates` (siehe §3): ein einziges `***`, `docs-check`; die drei Ziele laufen und sind gruen. `make gates` ist im Zweig rot, nicht wegen dieses Slice |
| 2 | Gleichzeitige Laeufe entschieden und an der Kette belegt | **erfuellt** (Nachweis als akzeptiertes Negativ mit Rate nur fuer Teile, siehe F-3) | `SPEC-211` Nr. 2 a bis e; eigene Rate: Kontrolle 0 von 10 rot, Mutation fester `--name` an allen drei Treibern 10 von 10 rot (§3, Nr. 3) |
| 3 | Zeitbudget gemessen und im Vertrag; `harness/README.md` fuehrt alle drei in der Gate-Tabelle, "Nicht behauptet" nur noch Ungeprueftes | **erfuellt mit Einschraenkungen** | `SPEC-211` Nr. 4 (Soll mit Messvorschrift, Trigger 120 s, Messung 2026-11-17 und Plan §3 *Zeitbudget* 2026-11-18); von mir nicht neu gemessen. Gate-Tabelle: drei Zeilen (Zeilen 70 bis 72 von `harness/README.md`), Zellen unter 200 Zeichen (`make docs-check` ohne Befund dort, Nr. 11). Luecke in *Nicht behauptet*: F-4 |
| 4 | `make gates` gruen | **nicht erfuellt im Zweig; erst nach dem Integrationsschritt erfuellbar, dann nicht belegt** | `make docs-check` im Zweig: 5 Befunde, alle in `docs/plan/planning/in-progress/roadmap.md` (Zeile 11 `planning-drift`; Zeilen 27, 159, 168, 170 `target-missing` auf `../open/slice-gate-integration-determinism.md`). Kein weiterer Befund (auch nicht in meinen Mutationskopien: Basis 5, mit Mutation 6). Die Roadmap schreibt nur der Integrationsschritt (Uebergangsregeln Nr. 4). Ob nach deren Nachzug `make gates` gruen ist und `record-gates` stempelt, ist **nicht gefahren, nur geschlossen**: alle uebrigen Checks liefen im `-k`-Lauf ohne Fehler |
| 5 | Review durchgefuehrt, Report liegt vor | **erfuellt, mit offenem Nachlauf** | `docs/reviews/2026-11-18-slice-gate-integration-determinism-review.md` (0 HIGH, 3 MEDIUM, 2 LOW, 2 INFO). Die Fix-Runde (`b8b0eeec`, `5904c6db`) ist von keinem Review gegengelesen; ich loese sie hier gegen die Klasse auf (§2) |
| 6 | Doku-Update `SPEC-037`, `SPEC-038`, `SPEC-042`, `harness/README.md`, `close-welle.md` als Diff-Vorschlag | **Teil 1 erfuellt; Diff-Vorschlag zu `close-welle.md` offen** | Nachzuege im Diff vorhanden (Stand, Treiber, *Ungeprueft* je Vertrag). `.claude/commands/close-welle.md` Zeile 47 traegt den Pflichtlauf noch; der Diff-Vorschlag an den User gehoert in die Closure und liegt nicht vor (Plan §6 fuehrt es als *offen*, Adresse Planner) |
| 7 | Closure-Notiz mit Lerneintrag | **offen** (Planner) | Plan §7 sagt "noch nicht begonnen"; Alle DoD-Haken in §2 stehen auf `[ ]` |
| 8 | Beobachtungs-Register fortgeschrieben | **offen** (Planner, Closure) | nicht im Diff; `BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger` Ausgang `geplant` unveraendert |
| 9 | Jedes Risiko aus §6 mit Ausgang | **erfuellt** | alle Zeilen von §6 tragen *entfallen* oder *offen/weiter offen* mit Adresse (Zeilen 189 bis 199) |
| 10 | Drei Paarungen (Anker, Folge-Slice, Register) | **offen** (Planner, Closure) | kein Anker, kein Steering-Loop-Eintrag, keine Register-Fortschreibung im Diff |

Zusammenfassung: Der Liefergegenstand (drei Zeilen, Texte, Gate-Tabelle, Vertrag) ist erfuellt und von mir an vier Zeilen der Tabelle mit eigener Mutation rot gesehen. Nicht erfuellt im Zweig ist die DoD-Zeile "`make gates` gruen" (nur Roadmap, extern verursacht). Offen sind die Closure-Punkte 6 (Diff-Vorschlag), 7, 8, 10, die dem Planner gehoeren; sie sind vor der Closure keine Verfehlung, aber noch nicht erfuellt.

## 2. Findings

| ID | Schwere | Befund | Ort | Klasse |
|---|---|---|---|---|
| V-1 | MEDIUM | `SPEC-211` Nr. 3 sagt "Der Exit eines Treibers ist der Exit von `make`". Falsch: `make` meldet bei jedem fehlschlagenden Rezept Exit 2. Gemessen: `test-determinism.sh` mit der Mutation von Nr. 5 (Bau-Pruefung entfernt) endet als Treiber mit Exit 1 (`test-determinism: rot (Exit 1)`), `make test-determinism` endet mit `Error 1`, Prozess-Exit 2. Die Zeilen der Tabelle, die "Exit 1" des Treibers (Zeile *Frist*, Zeile *Image fehlt*: "Exit 1") oder "Exit 2" vom Treiber unterscheiden, sind am Ziel `make` nicht unterscheidbar; die "Exit 2"-Aussagen der Fänger (Plan §3 Nr. 5, "Exit weiterhin 2 (3 von 3)") gelten fuer `make` trivial und fuer den Treiber nur fuer zwei von drei (siehe V-2). | `harness/vertraege/SPEC-211-container-tests-in-der-kette.md:22` (Intro Nr. 3) und `:28`, `:29`; Plan Nachweis Nr. 5 | Zusage ueber den Exit reicht weiter als Messung (`AGENTS.md` §3.11) |
| V-2 | MEDIUM | Plan-Nachweis Nr. 5 und `SPEC-211` Nr. 5 f ("ohne die Pruefung des Baus bleibt Exit 2, nur die Meldung fehlt (3 von 3)") gelten nicht fuer den Treiber `test-determinism`: mit der Mutation `if false && fahre ... build` laeuft der Treiber ohne Image-Bau weiter (`DOCKER=/bin/false`), `docker run` scheitert, der Treiber endet mit **Exit 1** (`rot (Exit 1)`), nicht Exit 2; die Meldung `Aufbau der Stufe ... gescheitert` fehlt. Bei `test-integration` und `test-contract` bleibt es bei Exit 2 (`-test.list ... gescheitert`). Der Bericht "3 von 3 Exit 2" ist eine Messung an `make`, nicht am Treiber (V-1). Die Aussage "der Faenger ist die Meldung, nicht der Exit" bleibt richtig; die Zahl "Exit weiterhin 2 (3 von 3)" ist es fuer den Treiber nicht. | `harness/vertraege/SPEC-211-container-tests-in-der-kette.md:53`; Plan §3 Nachweis Nr. 5 | wie V-1 (Messgroesse Treiber gegen `make`) |
| V-3 | LOW | Nr. 5 e/2 c (2) und (3) und die Kette "Handlauf 3 von 3" sind in der Reihe Zusage/Negativ konsistent, aber die Zusage Nr. 2 c (2) ("kein Container bleibt, ... Exit von `make` ist 143") steht im Text der Festlegung als Aussage, deren einzige Stuetze ein Handlauf ohne Mutation ist; Nr. 5 e sagt es selbst. Kein Mangel an Grund und Folge; ich fuehre es als Hinweis, weil die Gruppen-Aussage in der README-Zeile *Nicht behauptet* nicht steht (V-4). | `SPEC-211` `:17`, `:49` | Negativ vorhanden, nicht im Index |
| V-4 | LOW | `harness/README.md` *Nicht behauptet* nennt von den Negativen aus `SPEC-211` Nr. 5 a, b, c, Nr. 3 (Speicher und Frist), Nr. 2 d, Nr. 5 e (nur TERM allein an `make`), Nr. 5 f. Es fehlen Nr. 5 g (Gleichheit des Inhalts desselben Tags), Nr. 5 h (hoechstens drei Container), Nr. 5 i (Kopf-Kommentare/Text, Argument-Check, haengender Daemon) und aus Nr. 5 e der Container-Teil bei `test-integration` und `test-contract`, die Gruppe und die Obergrenze der Wartezeit. DoD Punkt 3 verlangt "nennt nur noch, was weiter ungeprueft ist", nicht "alles"; die Zeile ist damit unvollstaendig, nicht falsch. | `harness/README.md:94` (Absatz *Nicht behauptet*, Satz "Von `make test-integration`, `make test-determinism` und `make test-contract` ist nicht behauptet ...") | Index nennt nicht jedes Negativ des Vertrags |
| V-5 | INFO | Wortlaut der Spalte *Bindung*: Vertrag verlangt `Reproduzierbarkeit (LH-QA-06), Zeitbudget SPEC-211 Nr. 4`, die Tabelle traegt `LH-QA-06` und `SPEC-211` in Backticks. Kein Fänger (Nr. 11 prueft Existenz und Zellgroesse), keine inhaltliche Abweichung. | `harness/README.md:70` bis `:72` | Form |
| V-6 | INFO | Fänger Nr. 3: die Mutation muss an mindestens zwei Treibern stehen, damit sie rot wird (ein fester Name nur in einem Treiber kollidiert nirgends, weil jeder Treiber seine Container nacheinander startet). Gemessen sind die Mengen {alle drei}, {integration, contract}, {integration, determinism} (Review); die Menge {determinism, contract} ist nicht gefahren, die Zusage `SPEC-211` Nr. 2 e ("an jedem `docker run` der drei Treiber") ist damit im Wortlaut gedeckt, eine Teilmenge nicht. Folge: keine, die Mutation ist als Gesamtmutation definiert. | `SPEC-211` `:21`, Plan Nachweis Nr. 3 | akzeptiert |

Keine HIGH. V-1 und V-2 sind Vertragsfehler im Wortlaut (Architect), nicht im Code der drei Fragmente.

### Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Klasse laut Plan: jede Zusage von `SPEC-211`, 24 Elemente (Nr. 1 bis 5, zehn Zeilen der Tabelle Nr. 3, Nr. 2 a bis e, Nr. 5 a bis d). Der Plan zaehlt Nr. 5 a bis d; der Vertrag hat nach der Fix-Runde Nr. 5 a bis i (neun), darunter e bis i neu. Ich habe jedes Element gegen "Fänger oder Negativ mit Grund und Folge im Vertrag selbst" geprueft:

| Element | Fänger oder Negativ im Vertrag | Verdikt |
|---|---|---|
| Nr. 1 Bindung (Zeile, Kommentare, `make help`, Bindungswortlaut) | Negative 5 a, 5 b, 5 i; Handprobe (`grep`) | gedeckt; die Zusage "Kopf-Kommentare ... kein KEIN GATE mehr" ist nur per Hand (5 i), im Index fehlt es (V-4) |
| Nr. 1 Entscheidung (Alternativen, Verworfenes, Folge) | Text; keine Zusage ueber Verhalten | gedeckt |
| Nr. 2 a Tag (gleicher Inhalt; Tag-Setzen aendert laufenden Container nicht) | 5 g (Negativ, Grund, Folge) | gedeckt (F-5 des Reviews) |
| Nr. 2 b Container (`--rm`, kein fester Name; drei Container, 6 GiB) | Fänger Nr. 2 e; 5 h (Negativ drei Container) | gedeckt; Namens-Teil per Rate 10/10 |
| Nr. 2 c (1) INT/TERM am Treiber | Exit-Haelfte Handlauf mit Mutation; Container-Haelfte 5 e (nur `test-determinism`) | gedeckt (F-2); Zahlen nicht von mir nachgefahren, siehe unten |
| Nr. 2 c (2) Prozessgruppe | 5 e (Handlauf ohne Mutation, Grund, Folge) | gedeckt als Negativ |
| Nr. 2 c (3) TERM allein an `make` | 5 e (Negativ, Adresse Planner, Trigger) | gedeckt (F-1) |
| Nr. 2 c Wartezeit nach rotem Check | 5 e (Summe der Fristen, ungemessen) | gedeckt (F-4) |
| Nr. 2 d SIGKILL | Negativ im Punkt selbst (Folge: bis 120 s verwaister Container) | gedeckt |
| Nr. 2 e Nachweislauf | Fänger selbst, Rate und akzeptiertes Negativ definiert | gedeckt; von mir 10 von 10 gefahren |
| Nr. 3 Zeile 1 bis 3 (Docker, Bau, Image fehlt) | 5 f (Handlauf Meldung; Exit ungefangen; Zeile 3 ohne Handlauf) | gedeckt als Negativ; Exit-Aussagen siehe V-1, V-2 |
| Nr. 3 Zeile 4, 5 (Frist, Speicher) | Absatz unter der Tabelle (kein Fänger, Folge) | gedeckt; von mir nicht gefahren, Review: beide Mutationen gruen |
| Nr. 3 Zeile 6 (INT/TERM) | wie 2 c (1) | gedeckt |
| Nr. 3 Zeile 7 (Argument), 9 (Daemon haengt) | 5 i | gedeckt |
| Nr. 3 Zeile 8 (zwei Laeufe zugleich) | 2 a, b, 2 e | gedeckt |
| Nr. 3 Zeile 10 (Host unter Last) | Nr. 4 (kein Prüfer) | gedeckt |
| Nr. 4 Zeitbudget, Trigger | 5 c | gedeckt |
| Nr. 5 a bis i | selbst Negative mit Grund und Folge | gedeckt |

Mutation an einem Element, das der Befund nicht nannte, von mir rot gesehen: Befund F-6 nannte `laufe` in `test-determinism.sh`; ich habe die Mutation an `test-contract.sh` (Treiber, den F-6 nicht nannte, `laufe` dort Zeile 48) zusammen mit den zwei anderen gefahren: 10 von 10 rot (§3, Nr. 3). Fuer die Menge "Bau-Pruefung je Treiber" (Befund F-3 nannte `DOCKER=/bin/false` allgemein) habe ich alle drei einzeln gefahren und bei `test-determinism` ein Element gefunden, bei dem die Zahl der Fix-Runde nicht stimmt (V-2): die Klasse ist an dieser Stelle **nicht** vollstaendig gegen das Element aufgeloest.

## 3. Eigene Laeufe (Tabelle Zusage - Fänger, Stichprobe: Nr. 1, 3, 5, 11)

| Nr. | Mutation (Kopie) | Ergebnis | Urteil |
|---|---|---|---|
| 1 | `GATE_CHECKS += <Ziel>` je Ziel entfernt, `make -n -k gates`, Treffer des Treibers zaehlen | ohne Mutation 3; ohne `test-integration` 2, ohne `test-determinism` 2, ohne `test-contract` 2 | rot als Handprobe in 3 von 3 (kein Gate, `SPEC-211` Nr. 5 a) |
| 3 | fester `--name e2e-fest` an `laufe` von `test-integration.sh` und `test-contract.sh` und am `docker run` von `test-determinism.sh`; `make -j3 test-integration test-determinism test-contract` zehnmal je Zustand; drei Treiber-Prozesse je Lauf, bis zu drei Container zugleich; Messender: Verifier, Wanduhr, kein Detektor | Kontrolle **0 von 10 rot**, je Exit 0, kein `already in use` (Last 5,5 bis 8,2 beim Start). Mutation **10 von 10 rot**, je Exit 2, je Lauf genau 2 Zeilen `already in use`; Last beim Start 8,3 bis 10,5 in den Laeufen 1 bis 4, danach 30 bis 204 (Fremdlast) | gefangen in 10 von 10, auch unter stark erhoehter Last; `SPEC-211` Nr. 2 a, b gilt als gefangen |
| 5 | `if ! fahre ... build` durch `if false && fahre ...` je Treiber, `DOCKER=/bin/false make <Ziel>` und direkt am Treiber | Kontrolle: Meldung `Aufbau der Stufe ... gescheitert`, Exit 2, 3 von 3 (`make` und Treiber). Mutation: Meldung fehlt in 3 von 3; `make` Exit 2 in 3 von 3; **Treiber-Exit 2 bei `test-integration` und `test-contract`, Exit 1 bei `test-determinism`** | Meldung ist der Fänger (wie im Plan); die Exit-Aussage im Plan stimmt am Treiber nur 2 von 3 (V-2) |
| 11 | in der Kopie: Zeile `make test-contract` aus der Gate-Tabelle entfernt; Zelle der Spalte *Vertrag* mit 200 und mit 201 Zeichen (nach dem Zuschnitt auf den Inhalt ohne umgebende Leerzeichen) | ohne Mutation 5 Befunde (Roadmap); Zeile entfernt: 6, zusaetzlich `harness/mk/test-contract.mk:9 gate-undocumented`; 200 Zeichen: 5; 201 Zeichen: 6, zusaetzlich `harness/README.md:72 section-cell-oversized` ("201 Zeichen, erlaubt sind 200") | beide Seiten der Grenze rot/gruen aus dem benannten Grund |

Nicht von mir gefahren (Review und Plan tragen die Zahlen): Nr. 2 (`t.Fatal` in den drei Tests, `record-gates` laeuft nicht), Nr. 4 (`trap` entfernt, Exit 2 gegen 143), Nr. 6 (Syntaxfehler im Test, Exit 0 gruen), Nr. 8/9 (Frist und Limit entfernt, gruen), Nr. 10 (Zeitbudget). Nr. 4: ein Versuch, den Treiber im Hintergrund zu starten und mit TERM zu beenden, wurde von der Berechtigungsschicht des Laufs verweigert; ich habe es nicht umgangen. Nr. 12 (grep) gefahren: kein Treffer zu den drei Zielen in `harness`, `tools/harness`; `make help` nennt "Gate:" fuer alle drei.

**`make groessenmarke`:** `slice-gate-integration-determinism.md: ok, Messung 35 von Marke 1500` (Pfade des Blocks: `harness/README.md` 4+4, drei Fragmente je 5+4). Dockerfile und Treiber unveraendert (nicht im Diff). Metrik nach `AGENTS.md` §3.14 steht im Block (`SPEC-128`).

**`make zeichen-check`:** `zeichen: ok, 1836 Datei(en) geprueft`.

**`make docs-check`:** rot, 864 Dateien, 5 Befunde, alle in der Roadmap (siehe DoD Punkt 4). Kein Befund zu `harness/`, `SPEC-211` oder dem Plan.

**`make -k gates`** (unter Sperre, Last 7,3 bei Beginn): genau ein `***` (`docs-check`, `Error 1`), `make: Target 'gates' not remade because of errors`, Exit 2; `record-gates` lief nicht, `.harness/state/` existiert im Zweig nicht. Gruen in derselben Ausgabe: `groessenmarke` (35 von 1500), `migration-check: ok`, `test-contract: grün` (beide Binaries), `test-determinism: grün (TestE2EDeterminismus)`, `test-integration` Stufen 1 bis 9 grün, `zeichen: ok`, `a-check` (`gesamt: 0 Befund(e)`), die Stufen `handbuch` und `katalog` ohne Fehler; kein weiteres Ziel hat einen Fehler gemeldet (ein `***` in 715 Zeilen). Ich nenne das **nicht** "`make gates` gruen".

Gegengelesen auf Rohbytes: dieser Report; Nicht-ASCII nur das Paragraphenzeichen und Umlaute, kein Backslash-u-Muster, kein BOM, kein NBSP (LC_ALL=C grep -P).

## 4. Nicht geprueft

- Nr. 2, 4, 6, 8, 9, 10 der Tabelle (Zahlen des Reviews und des Plans uebernommen; V-1 wirft eine Frage an Nr. 4: die Exit-Haelfte 2 gegen 143 ist am Treiber gemessen und davon nicht beruehrt).
- Zeitbudget (`SPEC-211` Nr. 4) nicht neu gemessen; Host mit Fremdlast, "ein Host ohne Fremdlast ist nicht gemessen" gilt auch fuer mich.
- `make gates` ohne Roadmap-Befunde: nicht gefahren (die Roadmap darf ich nicht aendern); `record-gates`/Stempel daher unbelegt.
- Eine Probe des Roadmap-Nachzugs auf die Annahme "danach gruen": nur aus der Befundliste geschlossen.
- Rate fuer die Mutationsmenge {determinism, contract} (V-6).
- Der Pfad `.claude/commands/close-welle.md` Diff-Vorschlag (Closure).

## 5. Uebergabe

- An den **Architect**: V-1, V-2 (Wortlaut Exit von Treiber und `make`; Zahl "Exit 2" fuer `test-determinism` in 5 f und im Plan-Nachweis), V-4 (Index).
- An den **Planner** (Integrationsschritt und Closure): Roadmap (5 Befunde, ohne ihn ist `make gates` im Zweig nicht gruen), Closure-Notiz §7, Register, Paarungen, Diff-Vorschlag `close-welle.md`; `make gates` im Hauptbaum nach dem Nachzug und `make -j3` des Nachweises (Zehnerlauf) vor der Closure.
- Merge-Empfehlung des Verifiers: **nicht merge-blockierend durch Code**, aber DoD nicht vollstaendig: V-1 und V-2 sind MEDIUM im Vertrag; Closure-Punkte offen.
