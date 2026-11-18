# Verifikations-Report: slice-gegenproben-zeitbudget — 2026-11-18

**Rolle:** Verifier (Modul 11, `.harness/skills/closure-note-reviewer.md`): gegen DoD und Vertrag, nicht gegen den Diff (das ist das Review) und nicht gegen den Bedarf (das ist die Validierung).

**Gegenstand:** Zweig `slice-gegenproben-zeitbudget`, HEAD `ec4b6365`, Basis `a9287eb9`; Plan `docs/plan/planning/done/slice-gegenproben-zeitbudget.md`, [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) *Zeitbudget* Nr. 1 bis 7, [ADR-0013](../plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (`Accepted`), `AGENTS.md` §3.5 bis §3.18.

**Messender:** der Verifier. Host mit 20 Kernen (`nproc --all` = 20), Last = erstes Feld von `/proc/loadavg`. Ein Detektor fehlt: Zahlen sind Größenordnungen. Der Host war während der Läufe stark fremdbelastet (Last 10 bis 280 innerhalb einer Stunde); schwere Läufe liefen nur unter `flock …/.wt/heavy.lock` und, mit der einen genannten Ausnahme (Stufenbau, Beginn 17,84), bei Last unter 20.

## Urteil

**Der Liefergegenstand ist erfüllt; die DoD ist bis auf die Punkte, die dem Integrationsschritt, der Closure und der Eichung gehören, erfüllt.** Kein HIGH. Ein MEDIUM (Verhältnis zu [ADR-0013](../plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) ohne Adresse), zwei LOW, drei INFO. Ein gültiger Lauf der Gegenproben liegt nicht vor und war unter dieser Last nicht zu bekommen (siehe Stufenbau unten); das ist der Zweck des Slice, kein Befund gegen ihn.

## DoD Punkt für Punkt

| # | DoD-Punkt | Verdikt | Beleg |
|---|---|---|---|
| 1 | `SPEC-102` trägt die Messvorschrift (Gültigkeit, Lastgrenze, Wiederholung, Ausgang, Frist je Prozess), nennt das Soll als nicht maschinell geprüft; Frist unverändert, gehalten von drei Tests | erfüllt | `SPEC-102` *Zeitbudget* Kopf („ein Soll, kein Gate … Kein Programm, Gate oder Hook prüft die Gesamtzeit"), Nr. 1 bis 7, Randform 2a und 6. Der Diff `a9287eb9..HEAD` ändert `fallFrist` nicht (`main.go:31`: `120 * time.Second`). Mutation `fallFrist = 119 s` rot (unten). Die *Grenze* nennt vier Tests; die DoD nennt drei, die vierte ist `TestBewerteBau` (kein Widerspruch) |
| 2 | Jede Zeile der Tabelle *Zusage · Fänger* hat eine rot gesehene Mutation; `bewerteBau` und Tests liegen in `tools/gegenproben` | erfüllt | Tabelle Zeilen 1 bis 9 und *Nachweis* mit 17 Mutationszeilen (je 1 von 1 Prozess, Last 8,6 bis 14,6); ich habe 6 Zeilen neu gefahren, alle rot aus dem benannten Grund, der Lauf ohne Mutation grün |
| 3 | Doku-Update (*Stand*, *Grenze* geschrieben), `harness/README.md` unverändert, `make gates` grün | erfüllt | `harness/README.md` nicht im Diff; *Stand* trägt den Satz zur Messvorschrift; `make gates` im Zweig Exit 0 (unten) |
| 4 | Review durchgeführt, Report liegt vor | erfüllt | `docs/reviews/2026-11-18-slice-gegenproben-zeitbudget-review.md` (0 HIGH, 2 MEDIUM) und `…-review-2.md` (0 HIGH, 0 MEDIUM, 4 INFO) |
| 5 | Closure-Notiz mit Steering-Loop-Lerneintrag | erst nach Integrationsschritt / Closure | §7 trägt nur „noch nicht begonnen"; die Notiz schreibt der Planner bei der Closure (§5) |
| 6 | Beobachtungs-Register fortgeschrieben | erst bei der Closure | `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last` schreibt der Integrationsschritt (§6, Nr. 4 der Übergangsregeln); bis dahin nicht erfüllt |
| 7 | Jedes Risiko aus §6 trägt einen Ausgang | erfüllt | jede Zeile in §6 trägt *Ausgang* (entfallen, weiter offen oder offen mit Adresse) |
| 8 | Drei Paarungen (Anker · Folge-Slice · Register) | erst nach Integrationsschritt / Closure | Anker der geschärften Regel, Register-Eintrag und Folge-Eintrag entstehen mit der Closure |

## Prüfung (a): Liefergegenstand und §3.6

- **Messvorschrift in `SPEC-102`:** vorhanden, Nr. 1 bis 7, in der Form von `SPEC-211` Nr. 4. Der Stand-Satz „Das Budget liegt rund 28 % über …" ist entfallen, die Messreihen bleiben.
- **`bewerteBau`** (`tools/gegenproben/main.go:265`): `abgelaufen` vor `code != 0`; Exit 0 nicht rot; Aufruf `main.go:249`. Gegen Randform 2a und 6 gelesen: übereinstimmend.
- **Vier Tests der Frist** liegen vor und laufen: `TestFallFristIst120Sekunden`, `TestJederAufrufNenntDieFallFrist`, `TestFuehreFallReichtDenAblaufWeiter` (`frist_test.go`), `TestBewerteBau` (`main_test.go`). Sie liefen grün im Schritt `go test ./...` des Stufenbaus und im Lauf ohne Mutation.
- **§3.6:** keine Schwelle gelockert. `fallFrist` unverändert 120 s, kein Linter-, Coverage- oder Architekturgate berührt (der Diff trifft `SPEC-102`, `tools/gegenproben/{main,main_test,frist_test}.go`, Plan, Roadmap, Reports). Eine Beobachtung dazu unter I-1.

## Prüfung (b): „ein grüner ungültiger Lauf schreibt den Stempel" und ADR-0013

Der Entscheidungssatz der ADR lautet allgemein: ein Lauf außerhalb der Last sei ungültig und werde wiederholt. `SPEC-102` Nr. 4 und 5 d wiederholen nur rote ungültige Läufe; ein grüner ungültiger Lauf bleibt grün, wird nicht wiederholt und schreibt den Stempel.

Meine Wertung, ehrlich:

- **Kein Widerspruch, der eine Folge-ADR zwingt.** Die ADR delegiert „Zahl der Wiederholungen, Ausgang bei Bruch" ausdrücklich an `SPEC-102` und führt „wird wiederholt" im Zusammenhang mit dem Bruch des Soll. Ein grüner Lauf hat nichts zu entschuldigen; der Stempel hängt nach `SPEC-102` *Grenze* am grünen Lauf, nicht an der Last, und ein grüner Lauf unter Last ist der stärkere Beleg. Eine Wiederholung dort ändert kein Ergebnis. Die ADR bleibt unberührt (§3.5), die Schärfung steht dort, wo die ADR sie hinschickt (§3.8).
- **Aber es ist eine Auslegung des Wortlauts, die der User nicht bestätigt hat.** Wörtlich gelesen widerspricht „wird wiederholt" der Ausnahme. Eine Nebenwirkung ist real: Bleibt jeder Lauf unter Dauerlast grün-ungültig, bleibt das Soll unbelegt, und kein Befund meldet es (der Plan nimmt das hin: die Eichung liefert den ersten gültigen Lauf).
- **Was vor der Closure zu tun ist:** Der Architect hat die Lesart in §3 (Entscheidung 6 d) gesetzt, §6 führt sie aber nicht als offene Entscheidung mit Adresse (§3.13). Zu tun: eine Zeile in §6, Ausgang *offen*, Adresse User (Bestätigung, dass die Ausnahme eine zulässige Schärfung ist); bestätigt er sie nicht, ist die Folge eine neue ADR mit `Supersedes`, nie eine Änderung der alten. Bis dahin ist der Punkt eine Lesart, kein Beschluss (F-1).

## Prüfung (c): Sendung aus slice-gates-ohne-worktrees

- Der Geber `done/slice-gates-ohne-worktrees.md` §6 zeigt auf `done/slice-gegenproben-zeitbudget.md` §1 und §6; der Pfad ist mit diesem Slice nachgezogen (Diff der Datei: eine Zeile, `open/` zu `in-progress/`), der Link löst auf (`make docs-check` ohne Befund).
- Der Nehmer führt die Sendung in §1 *Ausdrücklich NICHT* (Eichung, Messung `time make gegenproben` mit den Fällen 205 bis 216, `build/ohne-trimpath`) und in §6 mit Ausgang *weiter offen*, Adresse Planner, mit der Kennung `slice-gates-ohne-worktrees`. Der Geber liegt in `done/`, der Nehmer ist der Planner und kein Slice; das entspricht §3.13 („liegt der Nehmer in `done/`, ist er keine Adresse; der Punkt geht an den Planner"). Geführt wie verlangt.
- **Slice-Zweck belegt (Lesart nach Vorschrift):** Mein Stufenbau hat die Beobachtung reproduziert (unten): dieselben neun Fälle `lint/*` mit `Zeit überschritten`, kein anderer Fall rot. Nach Nr. 3 ist der Lauf ungültig (Lesung am Ende 280,35 über 20), nach Nr. 5 e wird wiederholt. Das Rot des Programms bleibt unverändert (Randform 6: „der Fall ist in jedem Lauf rot"); der Slice liefert die Wertung, nicht eine Änderung des Falls. `build/ohne-trimpath` war bei mir grün; die Ursache des früheren Rot bleibt, wie der Plan sagt, offen.

## Prüfung (d): eigene Läufe

**Leichte Gates** (Last zu Beginn 21,2 bis 18,9, am Ende 18,9):

| Gate | Ergebnis |
|---|---|
| `make groessenmarke` | ok, „Messung 271 von Marke 1400", Hinweis „Tabelle ergibt 1135, Schätzung 975" (Hinweis, kein Rot, `SPEC-102` Fall 48) |
| `make docs-check` | 919 Dateien, 0 Befunde (vor Anlage dieses Reports) |
| `make zeichen-check` | ok, 1909 Dateien |
| `make lint` | Exit 0 |
| `make a-check` | 0 Befunde (Hinweis: 97 gescannte Dateien in keiner Schicht, vorbestehend, nicht Gegenstand) |

**`make gates` im Zweig** (Last zu Beginn 9,4, am Ende 10,9; unter `flock`): Exit 0, rund 41 s (warmer Cache); `.harness/state/gates-passed.head` trägt danach `ec4b6365`. Das Ergebnis gilt für `ec4b6365`, nicht nur für `d1e6c27c`. Der Nachweis liegt unter `.harness/state` und ist nicht versioniert (`git status` sauber).

**Stufenbau `gegenproben` ohne Speichergrenze** (`docker build … --target gegenproben`, Build-Args wie `harness/mk/gegenproben.mk`, Tag `odata-recorder:gegenproben-ver`; `make gegenproben` nicht gefahren, Stempel unberührt; unter `flock`):

- Lesung zu Beginn 17,84 (höchstens 20), Lesung am Ende 280,35 (unmittelbar nach dem Aufruf), Kernzahl 20; Wanduhr des Aufrufs 147 s (nur der Build der Stufe, nicht `time make gegenproben`).
- Schritt 10 (`gofmt`, `go vet`, `go test ./...` des Programms) grün: die Tests des Slice laufen.
- `gegenproben: 216 Fälle, 9 rot`. Rot sind genau neun Fälle `lint/*` (`basis`, `direktive-nolint`, `ausnahme-ohne-why`, `ausnahme-mit-why`, `direktive-mit-leerzeichen`, `why-falsch-eingerueckt`, `zweites-modul-sauber`, `befund-im-zweiten-modul`, `modul-ohne-go-datei`), jeder mit `Zeit überschritten`. Die Zahl der Fälle (216) stimmt mit Plan und `SPEC-102` *Stand* überein.
- **Lesung vor der Vorschrift:** neun rote Fälle der Gegenprobe; Befund gegen `lint` oder die Fälle.
- **Lesung nach der Vorschrift:** Lesung am Ende 280,35 über der Kernzahl, also *ungültig*; jeder rote Fall trägt `Zeit überschritten` (Nr. 5 e): kein Befund, **wiederholen**. Der Lauf ist kein gültiger Lauf und belegt weder das Soll noch seinen Bruch. Ich habe ihn nicht wiederholt: die Last blieb danach über 20 bis 200 (fremde Läufe), ein zweiter schwerer Lauf wäre gegen die Betriebsregel gewesen. Ob der Eigenanteil des Laufs (neun parallele `golangci-lint`-Läufe) an der Last von 280 beteiligt war, ist mit dieser Messung nicht trennbar (Plan §6, Eigenanteil der Lesung am Ende).
- Der Lauf des Reviewers mit `--memory=4g` (Compiler mit `signal: killed`) war nicht gültig; meiner lief ohne Speichergrenze und zeigt kein `killed`, nur `Zeit überschritten`.

**Mutationen** (Kopien in `scratchpad/ver-zb/m/`, nicht im Zweig; Container aus `odata-recorder:race`, ohne Netz, `--memory=4g --memory-swap=4g --cpus=4`, `go vet .` und `go test -count=1 .`, ein Prozess je Mutation, Klasse *andere*, rot in 1 von 1):

| Zeile | Mutation | Last zu Beginn | Ergebnis |
|---|---|---|---|
| ohne | keine | 17,4 | grün (`ok`, 0,86 s) |
| 1 | `fallFrist = 119 * time.Second` | 16,2 | rot: `TestFallFristIst120Sekunden`: `fallFrist = 1m59s, erwartet 120s` (die Gegenseite 121 s fing der Implementer) |
| 2 | `faelle.go:82` mit `3600000000000` statt `fallFrist` (eine andere Stelle als `zeichen_treiber.go` aus Plan und Review) | 10,0 | rot: `TestJederAufrufNenntDieFallFrist`: `faelle.go:82:18: Aufruf von starte nennt fallFrist nicht als erstes Argument` |
| 3 | im Zweig `abgelaufen` von `bewerteBau` der Text `Bau des Binary scheitert` statt `Zeit überschritten` | 14,4 | rot: `TestBewerteBau/Ablauf`, `…/Ablauf_mit_Exit_0` und `TestBewerte/Frist`: `nennt "Zeit überschritten" nicht` |
| 4 und 5 | `case code != 0` wird `case code > 1` (Exit 1 nicht mehr rot; eine Form, die der Plan nicht nannte) | 12,7 | rot: `TestBewerteBau/Exit_1_mit_langer_Ausgabe`, `…/nicht_startbar`, `TestFuehreFallBauScheitertRuftKeinGate` (`Rot = ""`, `Befehl 1-mal … gerufen, erwartet 0`) |
| 7 | Text von `bewerte` bei `abgelaufen`: `Zeit abgelaufen` | 12,0 | rot: `TestBewerte/Frist`: `bewerte = "Zeit abgelaufen", erwartet "Zeit überschritten"` |
| 8 | `bewerteBau(code, out, abgelaufen \|\| false)` (der Plan nennt `&& false`) | 10,8 | rot: `TestFuehreFallReichtDenAblaufWeiter/Bau`: `das letzte Argument von bewerteBau ist nicht der Bezeichner abgelaufen` |

Eine erste Fassung der Mutation zu Zeile 2 baute nicht (`undefined: time`, mein Fehler, `go vet` rot, kein gültiger Fänger); ich habe sie mit einer Konstanten neu gefahren, die Tabelle führt die zweite. In allen übrigen Kopien war `go vet` grün, die Fänger wurden aus dem benannten Grund rot.

## Prüfung (e): Stempel, Rohbytes, Marke, Roadmap

- **`harness/gegenproben.stempel`** unverändert gegenüber `a9287eb9` (`git diff --stat` leer). Die Eingabe `tools/gegenproben/` hat sich geändert, der Stempel ist damit veraltet: der Integrationsschritt muss `make gegenproben` fahren (Übergangsregel Nr. 2) und den Stempel allein committen, **bei Last unter der Kernzahl**, sonst entstehen wieder Rote an `lint/*` (siehe Stufenbau).
- **Rohbytes (§3.18):** `LC_ALL=C grep -cP '\\u[0-9A-Fa-f]{4}|\xef\xbb\xbf|\xc2\xa0|\xe2\x80[\x8b-\x8f\xa8\xa9]'` über `tools/gegenproben/main.go`, `main_test.go`, `frist_test.go`, `SPEC-102-gegenproben.md` und den Plan: je 0 Treffer. `make zeichen-check` bestätigt es über alle Dateien (1910 mit diesem Report). Das Gegenlesen des Implementers (`ü` als zwei Bytes) habe ich nicht wiederholt.
- **Marke 1400:** `make groessenmarke` meldet „ok, Messung 271 von Marke 1400" und den Hinweis „Tabelle ergibt 1135, Schätzung 975". §4 und §6 sagen es ehrlich: Tabelle von 7 auf 9 Zeilen, die Formel ergäbe 1135 und Marke 1600, die Marke wird nicht neu gerechnet, die Entscheidung steht beim Planner; Auslastung rund 0,19. Die Marke blieb vor dem Code unverändert und ist keine Rückführung nach dem Code (§3.14). Die Nicht-Neuberechnung ist als offene Entscheidung des Planners benannt, nicht als erledigt.
- **Roadmap Zeile 30:** der Link zeigt auf `in-progress/`, die Klammer trägt weiter `open/`. §6 führt das als Review F-4 mit Ausgang *offen*, Adresse Planner (Closure); die Roadmap schreibt der Architect nicht. Geführt wie verlangt (§3.13), aber ein Zustandsfeld (§3.7) bleibt falsch, bis der Planner es berichtigt (F-2).

## Findings

**MEDIUM**

- **F-1: Das Verhältnis von `SPEC-102` Nr. 4 und 5 d zum Entscheidungssatz von [ADR-0013](../plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) hat keine Adresse.** `docs/plan/planning/done/slice-gegenproben-zeitbudget.md:100` (Entscheidung 6 d des Architect) setzt die Lesart; §6 (`:198` bis `:208`) führt sie nicht als offene Entscheidung mit Ausgang und Adresse (`AGENTS.md` §3.13: „ein Bericht allein ist keine Adresse"). Der Review-Report 2 nennt sie als INFO G-2; der Plan hat sie nicht übernommen. Vor der Closure: Zeile in §6, Ausgang *offen*, Adresse User; bestätigt er die Lesart, schließt sie mit der Closure, sonst folgt eine neue ADR mit `Supersedes` (§3.5). Meine Wertung: zulässige Schärfung, kein Zwang zur Folge-ADR (siehe (b)).

**LOW**

- **F-2: Zustandsfeld der Roadmap trägt `open/` für einen Slice in `in-progress/`.** `docs/plan/planning/in-progress/roadmap.md:30`. Ausgang offen, Adresse Planner (§6), Berichtigung bei der Closure; keine DoD-Verletzung, aber ein falsches Zustandsfeld bis dahin (§3.7).
- **F-3: Die Marke ist im Plan nicht neu gerechnet, obwohl der Gate-Hinweis „Tabelle ergibt 1135, Schätzung 975" läuft.** `docs/plan/planning/done/slice-gegenproben-zeitbudget.md:166` bis `:182` (Block: `schaetzung: 975`, `marke: 1400`). Der Plan benennt es in §6 ehrlich und lässt die Entscheidung beim Planner; die Auslastung (0,19) liegt unter 0,60 und löst nach §6 eine Meldung an den Architect aus. Offene Entscheidung des Planners, kein Mangel an Wahrheit.

**INFO**

- **I-1: Das Soll ist von „kaltem Docker-Build-Cache" auf „lokal vorhandene Images" umformuliert** (`SPEC-102` *Zeitbudget* Kopf und Nr. 1; Review F-3). Alle Messreihen stammen aus dem Zustand „Images lokal, Stufe `gegenproben` ohne Cache", Nr. 1 nennt ihn genau; die Messgröße ist enger beschrieben, nicht lockerer. Der ADR-Kontext sagt weiter „kalter Cache"; keine Schwelle ist berührt (§3.6).
- **I-2: Der Lauf der Stufe ist unter der Last dieses Hosts nicht gültig zu bekommen.** Mein Lauf: Beginn 17,84 gültig, Ende 280,35 ungültig; die Last stieg binnen 2,5 Minuten um das Fünfzehnfache durch fremde Läufe. Die Vorschrift behandelt das (Nr. 4: drei ungültige rote Läufe in Folge gehen an den Planner). Ein roter Lauf schreibt den Stempel nie; der Integrationsschritt braucht einen ruhigen Zeitpunkt.
- **I-3: `SPEC-102` *Stand* sagt, die Eichung stehe aus.** Das ist richtig und ist die Folgepflicht in §6 (Adresse Planner); sie ist kein Liefer-Punkt.

## Nicht geprüft

- **Ein gültiger Lauf von `time make gegenproben`** (Last zu Beginn und am Ende höchstens 20): nicht zu bekommen, der Host war fremd belastet (Last 10 bis 280). Damit ist **nicht belegt**, ob das Soll von 120 s hält und ob `build/ohne-trimpath` und die Fälle 205 bis 216 gültig grün sind. Das ist die Eichung (Nr. 7), Adresse Planner.
- **Wiederholung des roten Stufenlaufs:** nicht gefahren (Last blieb über 20).
- **Die Mutationen der Zeilen 6 und 9, die Reihenfolge-Mutation der Zeile 3, die Zwischenvariable und die Bindungs-Mutation der Zeile 8:** nicht neu gefahren; nur aus der Tabelle und dem Review-Report 2 (31 Mutationen) gelesen. Das Verschieben von `fuehreFall` in eine andere Datei (G-1) ist weder von mir noch vom Reviewer als Mutation gefahren.
- **Die einzelnen Gates in `make gates`:** Exit 0, Stempel `gates-passed.head` und die Gates, die ich einzeln fuhr; die Ausgabe der übrigen Gates habe ich nicht Zeile für Zeile gelesen.
- **Die Zahlen der Messreihen in `SPEC-102`** (2026-10 und 2026-11): nicht neu gemessen, nicht nachgerechnet.
- **Closure, Register, Anker und Folge-Slice:** nicht Gegenstand dieser Stufe, §7 ist leer.

## Übergabe an den Planner

1. Vor der Closure: F-1 (Zeile in §6, Adresse User) und die Berichtigung der Roadmap-Klammer (F-2).
2. Integrationsschritt: `make gegenproben` im Hauptbaum bei Last unter 20, Stempel allein committen, Messung nach *Zeitbudget* Nr. 6 mit beiden Lesungen eintragen; ist der Lauf ungültig und rot, wiederholen (höchstens dreimal, Nr. 4).
3. Eichung nach Nr. 7 auf ruhigem Host; sie schließt die Zeilen von §6 zu Eichung und `build/ohne-trimpath`.
4. Entscheidung über die Marke (F-3) beim Planner.
