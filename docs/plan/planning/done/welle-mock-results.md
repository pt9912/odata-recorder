# Welle welle-mock — Mock-Modus mit Metadaten, zustandsbehaftetem CRUD, Abfrageoptionen und Snapshots — Closure-Notiz

> **Zitier-Form** *(bleibt stehen — Norm, kein Ausfüll-Hinweis).* Dieses
> Artefakt friert ein; was es zitiert, bewegt sich weiter. Deshalb: **Kennung,
> nicht Adresse** — `slice-<Kennung>` statt seines Lifecycle-Pfads, `make <target>`
> statt eines Links auf die Sensor-Datei, eine Baseline-Stelle als
> `v<X.Y.Z>` · `regelwerk/<datei>.md` §<Abschnitt> statt als Link
> (Baseline-Regelwerk `grundlagen-harness-dateien.md` §harness/README.md als
> Einstiegspunkt). Adoptierter Tag: `v6.18.1`.

**Welle:** welle-mock
**Abschluss:** 2026-11-08
**Verantwortlich:** pt9912

## Was wurde geliefert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — *was gelernt wurde*: geliefert · was
funktionierte · was anders lief. Mit ID-Bezug, wo es einen gibt.

14 Slices der Welle und vier wellenlose seit der letzten Closure (die Zahl aus dem Lauf von `archive-welle --vorschau`: 14 Mitglieder, 4 wellenlos). Der fünfzehnte Slice der ursprünglichen Liste, `slice-handbuch-mock`, gehört seit der Entscheidung des Users vom 2026-11-08 zu `welle-benutzerhandbuch` und ist kein Kriterium dieser Welle.

- **Der Meilenstein M3** (`LH-FA-06`, `LH-FA-07`, `LH-FA-08`, `LH-FA-09`, `LH-FA-10`, `LH-FA-11`, `LH-FA-12`): `TestM3` startet den Recorder im Modus `mock` mit einer Baseline aus einem Import, fragt `$metadata` und eine kombinierte Abfrage ab, legt eine Entity an, ändert und löscht sie, erzeugt einen Snapshot, stellt ihn wieder her und setzt zurück, ohne Verbindung zum beendeten Referenzdienst (`slice-bootstrap-m3`).
- **Parser, Ausdrücke, Interpreter:** Ressourcenpfad und Systemoptionen, `$filter`, die Typregeln der acht EDM-Typen mit dreiwertiger Null-Semantik, der Query-Interpreter mit stabiler Sortierung, Count vor Paging und Projektion (`LH-FA-07`, `LH-FA-08`, `LH-FA-11`; `slice-odata-url-pfad`, `slice-odata-url-filter`, `slice-edm-ausdruecke`, `slice-query-interpreter`).
- **Codecs:** Schreiber der Version 4 und Leser der Client-Bodies, `null` und fehlend unterscheidbar (`LH-FA-07`, `LH-RB-01`; `slice-odata-json-schreiben`, `slice-odata-json-eingabe`).
- **`MockUseCase` und `SnapshotUseCase`:** Zuordnung und `$metadata`, Lesen mit Coverage-Prüfung und Header `X-Recorder-Coverage`, Schreiben in der Workspace-Transaktion, Snapshots und `Restore` mit unverändertem Replay-Cursor (`LH-FA-06`, `LH-FA-07`, `LH-FA-09`, `LH-FA-10`, `LH-FA-11`; `slice-mock-zuordnung`, `slice-mock-lesen`, `slice-mock-schreiben`, `slice-snapshot-use-case`; Codes `ODR-E6008` und `ODR-E6009`).
- **Abnahme:** die Muss-Zeilen der Matrix `SPEC-049`, 100 identische Läufe nach Reset, 20 parallele Workspaces ohne fremden Zustand, Neustart mit erhaltenen Mutationen und Snapshots (`LH-QA-02`, `LH-QA-03`; `slice-mock-abnahme`).
- **Am gebauten Binary:** die Stufen 8 und 9 von `make test-integration` (`TestE2EMock`, `TestE2EMockNeustart`, `slice-bootstrap-m3-e2e-mock`) und `make test-contract` mit den Referenzanfragen `ODV4-*` und je Fehlerklasse von `LH-QA-05` einem Negativfall (`slice-test-contract`, `SPEC-042`).
- Verträge in `spec/spezifikation.md` und `harness/vertraege/` (darunter `SPEC-042`, `SPEC-049`, `SPEC-129` Nr. 16 bis 26, `SPEC-160` bis `SPEC-169`, `SPEC-180`, `SPEC-181`, `SPEC-190` bis `SPEC-199`); `model.AlleCodes()` zählt 40 Codes.

## Was hat funktioniert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3.

- Der gestaffelte Start: Teil 1 (sechs reine Funktionen und Codecs) lief als Linie B in einem eigenen Arbeitsbaum, bevor `welle-record-replay` schloss; die Handbuch-Pflicht ließ sich auf die Slices mit sichtbarem Verhalten begrenzen, die Sendungen gingen an `slice-bootstrap-m3`.
- Der Schnitt vor dem Code: `slice-mock-lesen` in `slice-mock-zuordnung` und `slice-mock-lesen`, `slice-bootstrap-m3` in Teil A und Teil E (`slice-bootstrap-m3-e2e-mock`); keine Rückführung nach dem Code, kein Eintrag `ueberschreitung`, keine Marke gerissen.
- Der Vertrag samt Tabelle *Zusage · Fänger* und Raster vor dem Code, durch den Architect geschrieben; Rückgaben an den Planner (Schnittfrage, Kennung, Abgrenzung der Codes) stehen in §6 der Pläne mit Ausgang (`AGENTS.md` §3.13).
- Die Abnahme als Tests in `make test` statt als Prosa: `TestM3` und die Gruppen `TestAbnahme*` laufen mit jedem Gate-Lauf.
- Das Register: 65 Beobachtungen mit 430 Belegen am Tag der Closure (gezählt mit `find … evidence`), gegen 43 und 278 bei der letzten.

## Was ging anders als geplant?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — jede Zeile möglichst mit der Konsequenz,
die daraus schon gezogen wurde (Folge-Slice, Spec-Version).

- **`slice-handbuch-mock` verließ die Welle** (Entscheidung des Users, 2026-11-08, Weg 1): Kapitel und gebundenes Beispiel zum Modus `mock` samt Erweiterung von `SPEC-200` sind ein Kettenslice der Handbuch-Gerüst-Linie und warteten auf die Gate-Kette. Konsequenz: der Closure-Trigger §3 der Welle-Datei nennt nur, was die Welle selbst liefert (Fehlercodes und Zeilen der Tabelle *Einstellungen* durch `slice-bootstrap-m3`); der Slice liegt in `open/` mit `Welle: welle-benutzerhandbuch`.
- **Die Linie B hat keine freie `SPEC`-Nummer mehr** (Bereich 160 bis 199 erschöpft): der Vertrag von `SnapshotUseCase` steht als `SPEC-129` Nr. 16 bis 26, ohne eigene Nummer. Die Kennungsfrage ist beim User offen (unten).
- **Auslastung der Marke:** neun von 14 Messpunkten liegen unter 0,60 (Ende gegen Marke), keiner über 1,00 (Tabelle unten); `slice-mock-abnahme` mit 0,24 ist der niedrigste Wert der Tabelle, ein Slice ohne Produktcode und ohne Port-Fakes außer zwei. Konsequenz: Evidenz in `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (28 Belege, Ausgang *geplant*, Auftrag an den Architect); der Architect legte bei `slice-mock-lesen` einen Vorschlag zur Neukalibrierung vor (Formel behalten und die Schwelle von Nr. 9 senken, oder neue Faktoren), die Entscheidung liegt beim User.
- **Die Betriebsregel „nur `Edit` und `Write`“ hielt in sieben Slices nicht** (17 Belege im Register, davon 13 seit der letzten Closure): `sed -i`, Heredocs auf Repo-Dateien, `git apply`-Patches der Architects ohne `Edit`, `pgrep`/`kill` und `rm -rf` mit Muster im gemeinsamen Scratchpad, generische Namen im Scratchpad. Konsequenz: `slice-guard-schreibwege` bleibt der Träger; die Formen ohne Repo-Pfad (Scratchpad, `git apply`, Prozesse) stehen in seinem Plan nicht.
- **Zwei Tests fremder Pakete flackerten unter Last** (`BEO-REPO/test-fremden-pakets-flackert-unter-last`, 1×) und die Nullprobe riss die Paketlaufzeit (`BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`, 2×); beide unter der Schwelle.
- **Die Ablage des Docker-Hosts wuchs** durch Einmal-Tags der Rollen: 833 Images (116 GB), 422 Volumes (127 GB) und 142 GB Build-Cache vor der Closure; neue Beobachtung `BEO-REPO/rollen-images-ohne-aufraeumen` (1×).
- **Die Archivierung ist gesperrt** (Feststellung unter *Verifikation*): neue Verweise mit Pfad auf Review-Reports entstanden nach der Auflösung der letzten Closure; neue Beobachtung `BEO-REPO/verweis-auf-review-report-sperrt-die-archivierung` (2×).

## Messpunkte der Größenmarke

Regeln dieser Sektion: `SPEC-128` Nr. 9 — Auslastung gleich Messung am Ende geteilt durch die Marke; über 1,00 oder unter 0,60 meldet der Planner dem Architect. Die Tabelle ist aus den Messpunkt-Zeilen und den Abschnitten *Größen* der Pläne unter `done/` nachgerechnet (*S* die Schätzung nach `SPEC-128` Nr. 1, *E* die Messung am Ende nach `git diff --numstat`); ein Strich heißt, der Plan trägt *S* nicht als Zahl.

| Slice | *S* | Marke | *E* | *E* ÷ *S* | *E* ÷ Marke |
|---|---|---|---|---|---|
| `slice-odata-url-pfad` | 6790 | 9200 | 6517 | 0,96 | 0,71 |
| `slice-odata-url-filter` | 3460 | 4700 | 2388 | 0,69 | 0,51 |
| `slice-edm-ausdruecke` | 3645 | 5000 | 3229 | 0,89 | 0,65 |
| `slice-query-interpreter` | 3275 | 4500 | 3201 | 0,98 | 0,71 |
| `slice-odata-json-schreiben` | 4570 | 6200 | 3367 | 0,74 | 0,54 |
| `slice-odata-json-eingabe` | 4360 | 5900 | 3692 | 0,85 | 0,63 |
| `slice-mock-zuordnung` | 4265 | 5800 | 2737 | 0,64 | 0,47 |
| `slice-mock-lesen` | 5370 | 7300 | 4371 | 0,81 | 0,60 |
| `slice-mock-schreiben` | 8220 | 11100 | 4544 | 0,55 | 0,41 |
| `slice-snapshot-use-case` | 7845 | 10600 | 5581 | 0,71 | 0,53 |
| `slice-mock-abnahme` | 7120 | 9700 | 2335 | 0,33 | 0,24 |
| `slice-bootstrap-m3` | 1745 | 2400 | 1367 | 0,78 | 0,57 |
| `slice-bootstrap-m3-e2e-mock` | 765 | 1100 | 570 | 0,75 | 0,52 |
| `slice-test-contract` | – | 2000 | 1311 | – | 0,66 |

Lesart: 14 Messpunkte. Median *E* ÷ Marke 0,56; Median *E* ÷ *S* 0,75 (13 Punkte mit *S*). Unter 0,60 liegen neun Punkte (`slice-odata-url-filter`, `slice-odata-json-schreiben`, `slice-mock-zuordnung`, `slice-mock-lesen` mit 0,599, `slice-mock-schreiben`, `slice-snapshot-use-case`, `slice-mock-abnahme`, `slice-bootstrap-m3`, `slice-bootstrap-m3-e2e-mock`), über 1,00 keiner. Die Zahlen gehen als Eingabe in den Auftrag an den Architect (Neukalibrierung von `SPEC-128` Nr. 1 und 2, Entscheidung des Users vom 2026-11-06); `slice-mock-lesen` zeigt 0,60 gerundet, der Plan führt dafür die Meldung an den Architect (Vorschlag V0 und V1).

## Steering-Loop-Einträge

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 (hier stehen **nur** Beobachtungen, die im
Register 3× erreicht haben; jeder Eintrag nennt seine `BEO-<KUERZEL>/<slug>`) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Lese-Schritt: 29 Beobachtungen haben 3× erreicht (gezählt: Verzeichnisse mit mindestens drei Dateien in `evidence/`). 27 standen schon in der Closure von welle-record-replay; zwei kamen dazu. **In dieser Closure ist keine Regel verkörpert worden**: Eine Regel in `AGENTS.md`, in einem Skill oder in einem Gate ändert ein Gate oder eine Hard Rule und ist Entscheidung des Users (`AGENTS.md` §3.5, §3.6); die beiden neuen Einträge stehen als Vorlage, ohne Feld `liegt in`.

- **Neu über der Schwelle, Vorlage, nicht verkörpert, beim User offen:** `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (16×, davon in dieser Welle `slice-mock-zuordnung`, `slice-mock-lesen`, `slice-mock-schreiben`, `slice-odata-json-schreiben`, `slice-odata-json-eingabe`). Ein Test sagt zu, `make test-race` melde jedes ungeordnete Paar, und führt nur die Zweige aus, die sein Autor nannte; jede Prüfrunde findet den nächsten Zweig. Das Mittel der Fix-Runden ist stets dasselbe: die Menge der Zweige aus dem Abdeckungsprofil der Detektor-Tests oder aus einer Menge im Modell mit Zählprüfung, nicht aus der Erinnerung. Vorlage 1 unten.
- **Neu über der Schwelle, Vorlage, nicht verkörpert, beim User offen:** `BEO-REPO/report-der-rolle-faellt-docs-check` (5×, in dieser Welle `slice-mock-zuordnung`). Ein Report von Reviewer oder Verifier trägt eine `ADR-*`-Kennung ohne Link und macht `make gates` rot, weil niemand nach dem Report `make docs-check` fährt. Vorlage 2 unten.
- **Spec-Lücke benannt, nicht aufgelöst:** der Workspace-Snapshot eines Workspace eines anderen Service ist heute `ODR-E7002` (`SPEC-129` Nr. 18), die Baseline eines anderen Service `ODR-E5004`; ob `ODR-E5004` auch für Workspace-Snapshots gilt, entscheidet der User (Rückgabe R3 in `slice-snapshot-use-case`, Negativ T78). Bezug `LH-FA-10`; Adresse User, weiter offen.
- **Ausgang bleibt, Zähler gewachsen** (kein neuer Eintrag, keine neue Regel; Retirement-Check: keine Regel wurde entfernt oder gelockert). Die Spalte *Ausgang* nennt den Stand aus `state.md`; die Anker dieser Regeln stehen in den Closure-Notizen davor.

| Beobachtung | Belege vorher → jetzt | Ausgang |
|---|---|---|
| `vertrag-sagt-mehr-zu-als-ein-gate-prueft` | 41 → 54 | verkörpert, `AGENTS.md` §3.15 und Reviewer-Skill (seit welle-record-replay) |
| `fix-deckt-die-genannte-mutation-nicht-die-klasse` | 21 → 44 | verkörpert, `AGENTS.md` §3.17, `implement-slice`, `verifier.md`; Werkzeug `tools/harness/mutationslauf.sh` geliefert (seit welle-record-replay) |
| `randform-erst-im-review-entschieden` | 30 → 35 | verkörpert, `AGENTS.md` §3.16 |
| `schaetzformel-ueberschaetzt-den-slice` | 8 → 28 | geplant, Auftrag an den Architect (`SPEC-128` Nr. 1 und 2), noch ohne Slice |
| `berichtigung-erreicht-nicht-jede-stelle` | 20 → 25 | verkörpert, `AGENTS.md` §3.9 |
| `betriebsregel-fuer-repo-dateien-verletzt` | 4 → 17 | geplant in `slice-guard-schreibwege` |
| `akzeptiertes-negativ-ohne-folge` | 13 → 15 | verkörpert, `AGENTS.md` §3.15 |
| `unbeschraenkte-menge-ohne-benannte-pruefgrenze` | 8 → 13 | verkörpert, `implement-slice` Schritt 14, Reviewer-Skill |
| `fixture-gleicht-dem-wert-der-mutation` | 3 → 7 | verkörpert, `implement-slice` Schritt 19, Reviewer-Skill |
| `sqlite-transaktionsvertrag-rot-unter-last` | 4 → 7 | behoben im Adapter `sqlite` (`slice-sqlite-interrupt-transaktion`) |
| `gegenprobe-haengt-an-der-laufzeit-des-produkts` | 4 → 7 | geplant in `slice-coverage-gate` |
| `rate-ohne-detektor-schwankt-mit-der-last` | 6 → 8 | verkörpert, `AGENTS.md` §3.10 |
| `zustand-des-arguments-als-dimension` | 4 → 6 | verkörpert, `implement-slice` Schritt 19, Reviewer-Skill |
| `universalaussage-im-vertrag-ungeprueft` | 14 → 15 | verkörpert, `AGENTS.md` §3.16 |
| `auslastung-am-handoff-als-endwert-gelesen` | 3 → 4 | geplant, Auftrag an den Architect (`SPEC-128` Nr. 9) |
| `kommentar-nennt-mehr-elemente-als-der-test-prueft` | 3 → 4 | verkörpert, Reviewer-Skill, `AGENTS.md` §3.11 |
| unverändert: `groessenmarke-ueberschritten-ohne-rueckfuehrung` (19), `nebenlaeufiger-lauf-nur-im-glueck-gruen` (13), `bericht-der-rolle-liegt-nicht-im-repo` (6), `suite-ohne-generator-fuer-klassen-der-spezifikation` (6), `werkzeug-eingabe-setzt-escape-in-zeichen-um` (5), `ziffern-klasse-ohne-schleife` (5), `befund-an-rolle-ohne-artefakt` (4), `hilfe-ersetzt-argument-verdeckt-eigenschaften` (4), `panik-eines-fakes-verdeckt-den-benannten-test` (4), `stempel-gehoert-zu-keinem-committeten-stand` (4), `menge-im-plan-weicht-vom-orakel-des-tests-ab` (3) | – | verkörpert wie in `welle-record-replay-results` |

**Lesart des Zuwachses.** Die Prosa-Regeln für Verträge, Fixes und Berichtigungen standen in dieser Welle von Anfang an im Auftrag und hielten die Zähler nicht: `fix-deckt-die-genannte-mutation-nicht-die-klasse` wuchs um 23, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` um 13, `berichtigung-erreicht-nicht-jede-stelle` um 5. Die Belege sind Befunde, die Review oder Verifikation fanden, bevor der Slice nach `done/` ging; die Regeln wirken auf den Plan, kein Sensor erzwingt sie. Das ist eine Beobachtung, kein Beleg einer Ursache. Der Sensor für die erste Beobachtung liegt als Werkzeug vor (`tools/harness/mutationslauf.sh`), ein Aufruf durch die Rollen ist nicht beschlossen (beim User offen, `SPEC-210`).

**Größengrenzen der Guide-Dateien als Rahmen jeder Verkörperung.** `.harness/skills/reviewer.md` steht auf 145 von 145 Zeilen und 8110 von 8110 Bytes, `.claude/commands/implement-slice.md` auf 17081 von 17081 Bytes, `.claude/agents/architect.md` auf 3759 von 3759 Bytes, `AGENTS.md` auf 329 von 329 Zeilen und 23648 von 23648 Bytes (Messung dieser Closure, Grenzen in `.d-check.yml`). Jede Vorlage unten verlängert eine solche Datei und macht `make docs-check` rot, solange sie nicht mit einer Kürzung im selben Diff oder einer Verschiebung der Grenze nach `SPEC-040` Nr. 8 b einhergeht (die Summe der Grenzen steigt nicht).

**Vorlage 1** (nicht verkörpert, beim User offen; Ziel: `.harness/skills/reviewer.md`, MEDIUM-Liste hinter *Fixture, Hilfe oder Kommentar verdeckt, was der Test prüft*; Kontextzeilen gekürzt, Größe: vier Zeilen, siehe oben):

```diff
--- a/.harness/skills/reviewer.md
+++ b/.harness/skills/reviewer.md
@@ MEDIUM, hinter dem Eintrag "Fixture, Hilfe oder Kommentar verdeckt, was der Test prüft"
+- **Zweig ohne Lauf unter dem Detektor** — ein Test sagt zu, `make test-race` melde jedes
+  ungeordnete Paar, und die Menge der Zweige stammt aus der Erinnerung des Autors, nicht aus dem
+  Abdeckungsprofil der Detektor-Tests oder aus einer Menge im Modell mit Zählprüfung; ein Zweig
+  ohne Lauf bleibt ungemeldet (`AGENTS.md` §3.10)  (seit welle-mock)
```

**Vorlage 2** (nicht verkörpert, beim User offen; Ziele: `.harness/skills/reviewer.md` §Output-Schema und `.claude/agents/verifier.md`; Kontextzeilen gekürzt, Größe: zwei bis drei Zeilen je Datei; `verifier.md` hat Platz bis 53 Zeilen und 3759 Bytes):

```diff
--- a/.harness/skills/reviewer.md
+++ b/.harness/skills/reviewer.md
@@ Output-Schema, hinter dem Absatz "Zusätzlich am Ende: …"
+Eine `ADR-*`-Kennung steht im Report als auflösender Link, nie nackt; nach dem Schreiben des
+Reports läuft `make docs-check`  (seit welle-mock).
--- a/.claude/agents/verifier.md
+++ b/.claude/agents/verifier.md
@@ hinter dem Absatz "Einen Fix löst du gegen die Klasse auf …"
+**Dein Report besteht `make docs-check`:** eine `ADR-*`-Kennung steht als auflösender Link, nie
+nackt, und der Lauf folgt dem Schreiben des Reports  (seit welle-mock).
```

Der Anker `seit welle-mock` in den Vorlagen ist der Anker, den die Verkörperung tragen würde; solange der User sie nicht übernimmt, steht er nirgends am Ziel, und die Einträge oben tragen deshalb kein `liegt in`.

- **Unter der Schwelle, neu in dieser Welle:** `BEO-REPO/rollen-images-ohne-aufraeumen` (1×), `BEO-REPO/verweis-auf-review-report-sperrt-die-archivierung` (2×). Unter der Schwelle sonst: 34 Beobachtungen, davon zehn mit 2× (darunter `nullprobe-grenze-unter-der-paketlaufzeit`, `pflichtlauf-neben-der-gate-kette-ohne-traeger`) und die übrigen mit 1× (darunter `test-fremden-pakets-flackert-unter-last`). Sie bleiben dem Sichtungs-Schritt der Planung.
- **Benannte Spec-Lücken** (ohne `liegt in`): neben der Lücke zu `ODR-E5004` oben die Kennungsfrage der Linie B (Bereich `SPEC-160` bis `SPEC-199` erschöpft, `SPEC-129` Nr. 21 bis 26 trägt den `SnapshotUseCase` ohne eigene Nummer; Adresse User, weiter offen).

## Beobachtungs-Register (Zeiger)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register — der Zähler wird **nicht** hier gepflegt; diese
Sektion ist ein Zeiger und trägt keine Daten.

Der Zähler steht in [`../observations/`](../observations/).
Was in dieser Welle **3×** erreicht hat, steht oben unter
*Steering-Loop-Einträge*.

## Folge-Slices

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — **derivativ**: Diese Liste zeigt nur,
das Original ist die Slice-Datei. Jeder genannte Folge-Slice muss als Datei im
Planning-Lifecycle existieren; genannt ohne angelegt ist dieselbe Klasse wie
ein halluziniertes Gate.

- `slice-handbuch-mock` (`open/`, `welle-benutzerhandbuch`, Linie C, Kettenslice): Kapitel und gebundenes Beispiel zum Modus `mock` samt Erweiterung von `SPEC-200`; aus dieser Welle verlegt.
- `slice-gate-integration-determinism` (`open/`, ohne Welle, Linie C, zurückgestellt): `make test-integration` und `make test-determinism` in `GATE_CHECKS`; bis dahin trägt der Pflichtlauf der Closure beide (`close-welle` Schritt 1). Er startet nach `slice-test-contract`, das in `done/` liegt.
- `slice-guard-schreibwege` (`open/`, Linie C): Wächter im Command-Guard für die Schreibwege auf Repo-Dateien; Träger von `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`.
- `slice-coverage-gate` (`open/`, `welle-sensoren`): Träger von `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`.
- Die Folger des Mock-Kerns, unverändert: `slice-hybrid-service` (`welle-hybrid`), `slice-telemetry-einbau-mock` (`welle-betrieb`), `slice-batch-changesets` (`welle-expand-batch`), `slice-bootstrap-m4` (`welle-hybrid`), `slice-admin-lebenszyklus` (`welle-admin-api`, braucht `SnapshotUseCase`), `slice-bootstrap-m6` (`welle-admin-api`) und `slice-lasttest` (`LH-QA-01` für den Mockserver, `welle-betrieb`).
- Kein Slice, ein Auftrag: die Neukalibrierung der Faktoren von `SPEC-128` durch den Architect (Eingabe: Tabelle oben).

## Verifikation

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 1 — keine Behauptung ohne nachprüfbaren
Anker (Hash, Lauf, Zahl).

Stand der Läufe: `main` auf `f3494bfd` (Arbeitsbaum ohne Änderung an Code, Spezifikation und Gates; die Änderungen dieser Closure betreffen Planungs-Dokumente). Alle schweren Läufe unter `flock /tmp/odata-recorder-heavy.lock`, nacheinander; Last des Hosts (Loadavg über eine Minute) 16,6 bis 22,0 durch fremde Läufe, keine Messung der Leistung.

- **Alle Slices der Welle in `done/`:** 14 von 14 Dateien `slice-<Kennung>.md` der Tabelle in §4 der Welle-Datei liegen unter `done/`; `slice-handbuch-mock` liegt in `open/` mit `Welle: welle-benutzerhandbuch` (Umplanung, Trigger §3 angepasst).
- **`make gates`:** Exit 0 (12:10:42 bis 12:11:05, 23 s). Aus dem Log: `a-check` 0 Befunde, `baseline-verify` v6.18.1 OK (54 Dateien), `docs-check` 914 Dateien ohne Befund, `groessenmarke` „kein Slice in Arbeit“, `migration-check` ok, `zeichen-check` ok (1835 Dateien), die Schema-Prüfung ohne Befund; die Stufen `test`, `race`, `lint`, `handbuch` und `katalog` kamen aus dem Docker-Zwischenspeicher (97 Schritte aus dem Zwischenspeicher, Quellen gleich dem letzten Bau), die Laufzeit der Tests ist damit nicht neu gemessen.
- **Der M3-Test und die Abnahmetests, frisch gefahren** (`go test ./tests/integration ./src/bootstrap -run '^(TestM3$|TestAbnahme)' -count=1 -v` im Image der Stufe `test`, `--memory 3g`, ohne Netz): Exit 0, 20 `--- PASS`, 0 `--- FAIL`; `TestM3` PASS in 1,15 s, `TestAbnahmeMatrix` 5,96 s, `TestAbnahmeDeterminismus` 4,04 s, `TestAbnahmeIsolation` 7,54 s, `TestAbnahmeNeustart` 0,34 s; Paket `src/bootstrap` 26,0 s. Beide Pakete sind Teil von `make test`.
- **`make test-integration`** (Pflichtlauf neben der Gate-Kette): Exit 0, neun Stufen grün (`TestE2EProzess`, `TestE2EWeiterleitung`, `TestE2EAufzeichnung`, `TestE2EPersistenz`, `TestE2EReplay`, `TestE2EMatching`, `TestE2ESchutz`, `TestE2EMock`, `TestE2EMockNeustart`), 33 s bei vorhandenen Images.
- **`make test-determinism`** (Pflichtlauf): Exit 0, `Läufe: 100, abweichend: 0` (`TestE2EDeterminismus`), 3 s.
- **`make test-contract`:** Exit 0, `/out/bootstrap.test` und `/out/integration.test` grün, 33 s. Das Ziel hängt nicht an `GATE_CHECKS`; die Aufnahme ist Entscheidung des Users und kein Kriterium der Welle.
- **`make gegenproben`:** der Stempel gehört zum Stand `f3494bfd` (Commit „Stempel der Gegenproben nach dem Merge von slice-test-contract“, 204 Fälle, 0 rot laut Auftraggeber, Lauf bei geänderter Eingabe vor der Closure); in dieser Closure meldete der Aufruf „Eingaben unverändert seit dem letzten grünen Lauf (8fa22172…)“ und lief nicht neu (1,3 s). Die Laufzeit des Laufs bei geänderter Eingabe ist in dieser Closure nicht erhoben (siehe [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) unten).
- **Kriterien §3 der Welle-Datei:** (1) alle Slices in `done/`: erfüllt. (2) `make gates` und `make gegenproben` grün: erfüllt, `make test-race` ist Teil der Kette (Stufe aus dem Zwischenspeicher). (3) M3-Test grün und Teil von `make test`: erfüllt, `TestM3`. (4) `make test-contract` grün: erfüllt. (5) Abnahmetests grün und Teil von `make test`: erfüllt, `TestAbnahme*`. (6) Handbuch deckt ab, was die Welle liefert: `make handbuch-check` und `make katalog-check` sind Teil des grünen `make gates`; Kapitel und Beispiel zum Modus `mock` sind kein Kriterium (Umplanung). (7) Ergebnisnotiz im Ruheort `done/`: diese Datei; die Einträge oben tragen kein `liegt in`, weil keine Regel verkörpert wurde.
- **Trigger-Audit, Carveouts:** `docs/plan/carveouts/` führt keine Datei; 0 offen, keine Welle schließt mit einem Carveout, kein Gate ist rot.
- **Trigger-Audit, bootstrap-aware Gates:** kein Gate trägt eine Reifestufe; die Schwellen sind feste Grenzen (Größengrenzen der Guide-Dateien nach [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md), Größenmarke nach `SPEC-128`). Die Größengrenzen stehen auf dem Ist (siehe oben), keine Verschiebung in dieser Closure. 0 offen.
- **Trigger-Audit, Entscheidungen** (Re-Evaluierungs-Trigger je ADR gegen den Stand der Closure):
  - [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): der Trigger „Lauf bei geänderter Eingabe über 30 s“ ist seit der Closure von welle-record-replay eingetreten (84,8 s Schritt der Fälle); der Auftraggeber führt das Budget von 120 s als Frage an den User. In dieser Closure lief kein Lauf bei geänderter Eingabe, die Messung bleibt aus. Die Gegenproben bleiben außerhalb der Kette; Entscheidung über die Folge beim User (offen), kein stilles Rot.
  - [ADR-0007](../../adr/0007-race-detektor-im-gate.md): `make test-race` ist Teil des grünen `make gates`, die Stufe kam aus dem Zwischenspeicher; kein Rennen gemeldet, kein Anlass für die Grenze von 10 Minuten. Nicht eingetreten.
  - [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md): `make groessenmarke` ohne Slice in Arbeit grün; die Messungen der 14 Slices lagen alle auf oder unter ihrer Marke. Status bleibt `Proposed`, die Annahme ist Entscheidung des Users (offen).
  - [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md): keine Verschiebung der Grenzen in dieser Closure; nicht eingetreten.
  - [ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md): der neue Schreiber von Blobs, der Snapshot-Dienst, schreibt über den Wächter und den Schutz in `ReferencedBlobs` (`SPEC-108` Nr. 16, Fänger T59 in `slice-snapshot-use-case`); kein Schreiber mit einem Aufruf zwischen `Commit` und Referenz über der Karenz. Nicht eingetreten.
  - [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md): der Lasttest zu `LH-QA-01` ist nicht gefahren (`slice-lasttest`, `welle-betrieb`); keine Beobachtung zu Pflege oder Lizenz. Nicht eingetreten, Geschwindigkeit ungeprüft.
  - [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Codec in Go geliefert), [ADR-0002](../../adr/0002-architektur-gate-a-check.md) und [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) (`make a-check` 0 Befunde), [ADR-0003](../../adr/0003-lint-profil-golangci.md) (`make lint` Teil des grünen Laufs), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Baseline unverändert `v6.18.1`), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) (diese Welle liefert Version 4, V2 bleibt Kann-Umfang von `welle-odata-v2`): Trigger nicht eingetreten.
- **Archivierung:** nicht ausgeführt. Die Vorprüfung `archive-welle --vorschau welle-mock` (14 Mitglieder, 4 wellenlos, 40 Review-Reports ohne Stub, 74 Präfix-Verweise in 14 Dateien) nannte vier Sperren: `[unsauber]` (eine ungetrackte, leere `.gitignore` im Wurzelverzeichnis, dieselbe Datei wie bei der Closure von welle-record-replay), `[ergebnisnotiz]` und `[kein-plan]` (Folge der Reihenfolge, mit dieser Notiz und dem Move der Welle-Datei nach `done/` aufgelöst) und `[haenger]` (36 Verweise mit Pfad auf Review-Reports in 34 Dateien, überwiegend Evidence-Dateien des Registers). Die Auflösung von `[haenger]` ist ein Handlauf über viele Dateien und gehört nicht in diese Closure (Auftrag: Feststellung statt Handlauf); `welle-record-replay` ist aus demselben Grund unarchiviert, `done/` führt für beide Wellen weiter flache Pläne. Die Archivierung führt der User nach Entfernen oder Festschreiben der `.gitignore` und nach dem Nachzug der Verweise; Sensoren: das Doku-Gate liest `.` rekursiv und sieht die Stubs unter `done/<welle-id>/`, das Modul `reviews` ist ausgeschaltet, `planning` liest die Roadmap; die Suche nach `planning/done` in Skripten, Make-Fragmenten und Go-Quellen ohne Tests fand nur `harness/mk/archivierung.mk` (Gegenproben-Fälle in Testdateien sind nicht durchsucht).
