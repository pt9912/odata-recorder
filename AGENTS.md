# AGENTS.md — Briefing für AI-Coding-Agenten

---

## 1. Was diese Datei ist

Onboarding-Briefing für jede AI-Session, die in diesem Repo Code oder
Dokumentation ändert. Sie verweist auf die kanonischen Quellen und
formuliert die Hard Rules, die der Implementer-Agent immer
einhalten muss.

Regeln dieser Datei: Baseline-Regelwerk `modul-09-implementierung.md`
§Ziel-Form: AGENTS.md — sie trägt Hard Rules und Pointer auf kanonische
Quellen, sie dupliziert deren Inhalt nicht; sonst entsteht Drift.

**Bei Konflikt zwischen dieser Datei und einer kanonischen Quelle gilt
die kanonische Quelle** (Source Precedence — siehe
`harness/README.md`).

Strukturregeln (ID-Schemata, Verzeichniskonvention, Adaptionen ggü.
Baseline, Modus-Deklarationen pro Sub-Area, Zusatzklassen für
Sensors-Bindung) leben in
[`harness/conventions.md`](harness/conventions.md).

Das **Regelwerk der adoptierten Baseline** ist die **präsente,
nachschlagbare Vertiefung** zu diesem Briefing: ein self-navigierbares
**Modul-Bundle** (`README.md` = Index). Beim Bootstrap wird das
self-contained Release-ZIP
(<https://github.com/pt9912/ai-harness-course/releases/download/v6.18.1/lab-regelwerk.zip>)
**committet vendored** unter `.harness/baseline/<tag>/{regelwerk,templates}/`
(Regelwerk *und* Templates parallel, netzlos materialisiert samt `SHA256SUMS`
— Vorgehen siehe
Baseline-Regelwerk `modul-02-harness-bootstrap.md`;
Quelle/Stand in [`harness/conventions.md`](harness/conventions.md) §Baseline).

Die verkörperte Form (dieses Briefing, die Konventionen, deine
ausgefüllten Artefakte) **führt**; das Regelwerk wird **pro Entscheidung
nachgeschlagen, deren
operative Detailtiefe das Briefing nicht trägt** — Trigger-Klassen,
Sub-Area-Qualifikation, Carveout-vs-Reconciliation, Modus-Diagnose. Dabei
**nur den benötigten Abschnitt** laden (README ist der Index), **nicht das
ganze Regelwerk im Kontext halten**. Breiterer Pflicht-Blick bleibt bei:
Bootstrap, Änderung an [`harness/conventions.md`](harness/conventions.md)
(Adaptionen `MR-<NNN>`, Source-Precedence, ID-Schema), Drift-Audit gegen die
Baseline (Baseline-Regelwerk `modul-02-harness-bootstrap.md`
§Freshness-Audit der vendored Baseline — darunter die Stichprobe gegen
den Bestand, die auch bei aktuellem Pin läuft). Derivativ: bei Konflikt gelten die kanonischen Quellen.

Die **Skelett-Vorlagen** der Baseline liegen **vendored** unter
`.harness/baseline/<tag>/templates/` (aus demselben Baseline-Bundle) und
tragen zwei Rollen: als **Referenz-Form**, auf die das Regelwerk mit
`../templates/…` als „Ziel-Form" verweist (netzlos, weil parallel zu
`regelwerk/` vendored), und als Vorlage, die beim Anlegen neuer Artefakte
(ADR, Slice, Welle, Carveout, Review-Report) **kopiert und ausgefüllt** wird
statt frei zu formulieren.

## 2. Kanonische Quellen (Source Precedence)

In dieser Reihenfolge:

1. [`spec/lastenheft.md`](spec/lastenheft.md) — vertraglich abnahmebindend.
2. [`spec/spezifikation.md`](spec/spezifikation.md) für das Produkt und [`harness/vertraege/`](harness/vertraege/README.md) für die Werkzeuge — technisch verbindlich, fortschreibbar.
3. [`spec/architecture.md`](spec/architecture.md) — Komponenten- und Sequenzsicht.
4. [`docs/plan/adr/`](docs/plan/adr/) — ADR-Verzeichnis und -Index.
5. [`docs/plan/planning/in-progress/roadmap.md`](docs/plan/planning/in-progress/roadmap.md) — Wellen-Sequenz.
6. `docs/user/*` *(falls vorhanden)* — Operations, Quality, Releasing. <!-- d-check:ignore (Verzeichnis optional; entlinkt, da im frischen Repo selten vorhanden) -->
7. [`README.md`](README.md) — Projekt-Überblick.
8. **AGENTS.md (diese Datei).**
9. [`harness/README.md`](harness/README.md) — Harness-Einstieg.

## 3. Harte Regeln



### 3.1 Docker-only



Kein lokaler Toolchain-Install (Compiler, Python, SQLite, Java). Alles läuft über
`make`, das Docker nutzt; Images sind per Digest gepinnt, die Prüfläufe netzlos
(`--network none`) und ohne schreibbaren Mount. Die `schema-*`-Targets arbeiten ganz
ohne Mount (Eingaben per `docker cp`, Ergebnisse als Tar-Strom); die Doku-Gates
mounten das Repo schreibgeschützt (`:ro`). Host braucht nur Docker und GNU `make`.

**Falsch:** `sqlite3 :memory: < schema.sql`, `python tools/schema/check.py`
**Richtig:** `make schema-check`

**Begründung:** Toolchain-Reproduzierbarkeit + Supply-Chain-Defense.

### 3.2 Suppression-Verbot

Eine Prüfung wird nicht durch Unterdrücken im geprüften Text umgangen.
Im Code ist eine Direktive `//nolint` verboten; Ausnahmen leben in `.golangci.yml` unter
`exclusions.rules`, jede mit einem Kommentarblock, dessen erste Zeile mit `# Why:` beginnt
([ADR-0003](docs/plan/adr/0003-lint-profil-golangci.md)).
`nolintlint` meldet eine Direktive ohne Begründung, ohne benannten Linter oder ohne Wirkung.
In der Doku gibt es die Ausnahme `<!-- d-check:ignore (Grund) -->`; sie trägt ihre Begründung in der Klammer, und eine ohne Begründung ist ein Befund.

### 3.3 git mv + Inhaltsänderung = zwei Commits

Wenn eine Datei verschoben **und** der Inhalt umgeschrieben wird, sind das
zwei Commits — der Move-Commit bleibt rein (Git erkennt R-Rename). Welcher
zuerst kommt, sagt der Vorgang:

1. Regelfall: `git mv source target` → eigener Commit, dann Inhalt umschreiben.
2. Lifecycle-Übergang nach `done/`: erst der Inhalt (DoD-Häkchen,
   Closure-Notiz), dann der reine `git mv` — die Notiz ist die Bedingung für
   `done/`, nicht ihre Folge.

**Begründung:** Sonst fällt die Rename-Detection unter die 50%-
Similarity-Schwelle und `git log --follow` wird unzuverlässig.

### 3.4 Architektur ist sprach- und meilensteinfrei

`spec/architecture.md` darf Pfade zu **Code-Modulen** referenzieren
(`src/service/`), aber **keine** Wellen, Slices, Commit-Hashes oder
Closure-Daten. Die zeitliche Schicht lebt in
`docs/plan/planning/` und den späteren Closure-Notizen. Auch **keine
ADR-Bezüge**: Die Sicht steht im Stabilitäts-Rang über der ADR; welche ADR
eine Aussage verbindlich macht, deklariert die ADR in ihrem `Schärft:`-Feld.

Diese Regel ist *verkörpert*, nicht hier entschieden — sie folgt aus dem
Sicht-Stratum (Baseline-Regelwerk `modul-03-spec.md`
§Ziel-Form: Architektur-Sicht).

### 3.5 ADRs sind nach `Accepted` immutable

Eine ADR mit Status `Accepted` wird nicht inhaltlich überschrieben.
Korrekturen entstehen als neue ADR mit `Supersedes ADR-NNNN`.

### 3.6 Gates dürfen nicht ohne ADR gelockert werden

Jede Schwellen-Senkung (Coverage, Linter-Strenge, Architekturregel)
ist ein ADR, kein PR-Kommentar. Eine befristete Ausnahme für einen Teil (einen
Layer, einen Pfad) ist keine Senkung, sondern ein Carveout mit Trigger und
Folge-Slice; die Schwelle selbst bleibt.

### 3.7 Ein Kommentar beschreibt, was da ist

Gilt für Code, Konfiguration und Skripte — und für Zustandsfelder (unten).
Ein Kommentar trägt eine dieser Klassen — **Zusage · Kopplung · Abgrenzung ·
Rang-Zeiger · Grenze** — und schreibt an den, der die Stelle *ändert*, nicht an den, der die
Entscheidung *trifft*. Regeln dieser Sektion: Baseline-Regelwerk
`grundlagen-harness-dateien.md` §Was ein Kommentar trägt.

**Falsch:** „Ohne diese Prüfung hätte Replay einen abgebrochenen Exchange
ausgeliefert" — Konjunktiv über die verworfene Alternative.
**Richtig:** „Nur `capture_status=complete` ist replayfähig" — Indikativ über den
Zustand.

**Falsch:** „die frühere Fassung zählte `position` statt `consumed`" — beschreibt
abwesenden Text.
**Richtig:** die geltende Zusage nennen; die vorige hält `git`.

**Zustandsfelder ebenso:** Eine `Stand`-/`Status`-Zelle in Roadmap,
Beobachtungs-Register oder Meilenstein-Tabelle nennt den Zustand und den Beleg
als auflösbaren Anker, nicht die Chronik; das Drift-Log der Roadmap trägt nur
Umplanungen, keine Schließungen und keine erreichten Meilensteine.

**Begründung:** Die Abwägung gehört in die ADR, die Historie in `git`, die
Herkunft in **ein** auflösbares Feld (`LH-*`, `ADR-*`, `· seit welle-<Kennung>`).
Was daneben steht, liest jeder Lauf mit und bezahlt es mit Kontext.





### 3.8 Eine ADR trägt Entscheidung und Gründe, keine Regeldetails

Eine ADR hält fest, was entschieden wurde und warum.
Die Regeln im Einzelnen (Fälle, Ausnahmen, Codes, Rangfolgen) stehen in der Spezifikation oder in `harness/vertraege/`; die ADR zeigt über `Schärft:` auf sie und wiederholt sie nicht.

**Falsch:** die Fallunterscheidungen einer Spezifikationsstelle in der ADR nacherzählen.
**Richtig:** die Entscheidung, die Alternativen und die Gründe nennen und für die Einzelregeln auf die Stelle zeigen.

### 3.9 Der Slice-Plan folgt jeder Korrektur   (geschärft seit welle-record-replay)

Wer im Slice Code, Gates, Tests, die Spezifikation oder Harness-Artefakte (`.claude/commands/`, `.harness/skills/`, Hooks) ändert, zieht im selben Commit §1 (Ziel und Abgrenzung), §3 (Plan) und §6 (Risiken) des Slice-Plans nach, dazu den Kopf (`Bezug`, `Berührte Spec-Stellen`) und betroffene Folge-Slices.
Wer eine Aussage berichtigt, sucht alle Stellen, die sie tragen, und zieht jede im selben Commit nach.

**Falsch:** Ein Gate bekommt eine zweite Prüfung, und die Sensors-Tabelle und §3 nennen weiter die erste.
**Richtig:** Dieselbe Änderung trägt die Tabellen- und Plan-Zeile mit.

### 3.10 Jede Zusage eines neuen Vertrags fängt eine Mutation   (geschärft seit welle-record-replay)

Wer einen neuen Vertrag liefert (Format, Fehlercode, Leser, Gate, Option), schreibt zu jeder seiner Zusagen einen Test, der rot wird, wenn der Code die Zusage bricht, und fährt diese Mutation vor der Übergabe selbst.
Der Bericht nennt je Zusage: Zusage · Mutation · roter Test.
Ein Test zählt nur, wenn er aus dem benannten Grund rot wird, nicht weil der Aufbau bricht.
Wird ein Fänger nur mit Wahrscheinlichkeit rot, nennt der Bericht seine Rate je Mutation und je Lauf von `go test`; „hoch“ steht nur mit Zahl.
Wo ein Vertrag eine Zusage über Nebenläufigkeit macht (Sperre, Zähler, Reihenfolge), gilt sie nur als gefangen, wenn die Mutation in jedem Lauf rot wird.
Ein Datenrennen zählt der Detektor von `make test-race` je Prozess (`-test.count=1`; er meldet dasselbe Rennen in einem Prozess nur einmal, ein `-count=20` zeigt es nicht), jede andere Wirkung eine Wiederholung (`-count=20`).
Sonst bekommt die Zusage einen Wirkungs-Fänger, der die Wirkung zählt statt auf ein Datenrennen zu warten, oder steht als akzeptiertes Negativ mit Grund im Plan; das trifft, was der Detektor nicht meldet: einen Zugriff, den kein Lauf ausführt, Prüfen und Handeln in getrennten Abschnitten bei lauter geordneten Zugriffen und ein Modul unter `tools/`.
Eine Zusage der Form „unter einer Sperre, ohne sie dazwischen freizugeben“ bekommt im Raster je Fenster der Sperre (zwischen welchen Schritten, auf welcher Seite) eine Zeile mit Fänger oder benanntem Negativ.
Vor einem Test-Haken oder einem Negativ sucht der Plan einen Haltepunkt an einem Argument, das der Test übergibt (`ctx`, Strom, Fake).
Der Plan nennt keinen Lauf, den das Image nicht fährt.

**Falsch:** Die Tests decken den Gutfall; dass eine Mutation grün bleibt, findet erst das Review.
**Richtig:** Je Zusage eine Mutation, vom Implementer selbst rot gesehen.

**Falsch:** Die Zeile für das Fenster zwischen Prüfen und Handeln nennt `make test-race` als Fänger; jeder Zugriff liegt unter der Sperre, der Detektor sieht nichts.
**Richtig:** Die Zeile nennt einen Wirkungs-Fänger, der zählt, was das Fenster zuließe, oder ein Negativ mit Grund; das Fehlen der Sperre fängt der Detektor, rot in <N> von <N> Prozessläufen.

### 3.11 Ein Kommentar sagt nur zu, was ein Test prüft

Schärft §3.7 für Zusagen: Ein Kopf- oder Hilfetext, der ein Verhalten zusagt („beide Adapter", „jede Randform"), nennt nur, was die Prüfung tatsächlich fängt.

**Falsch:** Der Kopf der Gegenprobe sagt „prüft beide Adapter", sie prüft einen.
**Richtig:** Der Kopf nennt den einen und sagt, was offen bleibt.

### 3.12 Randformen eines neuen Vertrags sind vor dem Code entschieden

Ein Slice, der einen neuen Vertrag liefert (§3.10), nennt in §6 des Plans dessen Randformen — fehlender, leerer oder unbekannter Wert, Abbruch, Herunterfahren, Verhalten je Upstream — und je Randform, wo sie entschieden ist.
Entschieden ist jede vor dem ersten Code-Commit, in der Spezifikation (Produkt) oder in `harness/vertraege/` (Werkzeug).
Bemerkt der Implementer eine Randform, die dort nicht steht, hält er an und gibt sie an den Architect zurück; er entscheidet sie nicht im Code.

**Falsch:** Der Leser entscheidet `null` und den leeren Wert still, das Review findet es.
**Richtig:** Beide Formen stehen vor dem Code in der Spezifikation.

### 3.13 Eine Adresse nimmt an, bevor sie genannt wird

Wer einem anderen Slice etwas zuweist (Folge-Slice, Risiko-Ausgang *eingetreten*, Abgrenzung), liest vorher dessen §1 *Ausdrücklich NICHT* und dessen DoD.
Im selben Commit trägt er die Sendung in §1 oder die DoD des Nehmers ein, mit der Kennung des Gebers.
Trifft ein Ausschluss des Nehmers die Sendung oder liegt der Nehmer in `done/`, ist er keine Adresse; der Punkt geht an den Planner.

**Falsch:** §6 weist einen Punkt `slice-x` zu, dessen §1 ihn ausschließt.
**Richtig:** Der Nehmer führt den Punkt in §1 oder seiner DoD, bevor der Geber ihn nennt.

### 3.14 Die Größenmarke schneidet vor dem Code oder am vorab benannten Teil und nennt ihre Metrik   (geschärft seit welle-record-replay)

Jede Marke in §4 eines Plans („zu groß, über N Zeilen") nennt ihre Metrik: die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade, die der Plan nennt, Tests eingeschlossen.
Schätzungen werden in dieser Metrik angegeben: Zeilen der Tabelle *Zusage · Fänger* (§3.15) × 85 (nebenläufig) bzw. × 40 (jede andere Zeile), plus 200 je Port-Fake oder Test-Server.
Nebenläufig zählt eine Zeile, wenn der Code oder sein Fänger nebenläufig ist.
Der erste Slice mit Zeilen, die keine reinen Funktionen sind, nennt den gemessenen Wert in §4.
Die Formel schätzt den Stand beim Handoff; die Marke misst den Stand nach dem Review, und dafür wird die ganze Schätzung, feste Anteile eingeschlossen, × 1,25 gehalten, denn das Ende lag nachgemessen 17 bis 25 % über dem Handoff.
Die Marke setzt der Planner vor dem ersten Code-Commit auf die Schätzung × 1,25, aufgerundet auf volle Hundert; eine ältere Marke ersetzt er dabei mit Grund in §4.
Liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten; lässt sich kein Teil ohne geteilten Mechanismus abtrennen, steht der Grund in §4.
Daneben ist ein vorab benannter Schnitt beim Handoff erlaubt: der Teil ist vor dem Code in §4 benannt, zuletzt in eigenem Commit geschrieben und davor gemessen, und er teilt keinen Mechanismus mit dem Rest.
Sonst wird nach dem Code gemessen und mit Grund berichtet.

**Falsch:** „über 700 Zeilen", ohne Metrik; eine Marke unter der Schätzung × 1,25, die der Planner vor dem ersten Code-Commit nicht neu gesetzt hat; eine Marke, die nach dem fertigen Code eine Rückführung verspricht; ein Schnitt beim Handoff an einem Teil, den erst die Messung benennt.
**Richtig:** „über 700 Zeilen (Summe aus `git diff --numstat`, Tests eingeschlossen)", vor dem ersten Code-Commit vom Planner auf die Schätzung × 1,25, aufgerundet auf volle Hundert, gesetzt; vor dem Code wird nur über 3000 geschnitten, beim Handoff nur der Teil, den §4 vor dem Code benannt hat; sonst steht nach dem Code die Messung mit Grund im Plan.

### 3.15 Zusage und Fänger stehen als Paar im Plan   (seit welle-fundament, geschärft seit welle-record-replay)

Ein Slice, der einen Vertrag liefert, führt in §3 seines Plans die Tabelle *Zusage · Fänger*, bevor er an den Reviewer übergeht.
Jede Zusage, Randform und jeder Fall des Vertrags hat eine Zeile; der Fänger ist ein Test, ein Fall der Gegenprobe oder eine benannte Ausgabe des Gates.
Ein Fänger führt jede Form der Eingabe, die der Code verschieden behandelt: ein Fall je Element einer Menge, beide Seiten jeder Grenze und der Fall, in dem zwei verglichene Werte gleich sind.
Eine Fixture gleicht nie dem Wert der Mutation.
Der Reviewer löst jede Zeile mit einer Mutation auf, der Verifier prüft die Tabelle gegen den Vertrag.
Jedes akzeptierte Negativ nennt Grund und Folge, sonst ist es eine Zusage ohne Fänger.
Eine Zusage ohne Fänger ist ein MEDIUM.

**Falsch:** Der Vertrag nennt eine Prüfung, und sie lässt sich entfernen, ohne dass etwas rot wird.
**Richtig:** Die Tabelle nennt zur Prüfung den Fall, der ohne sie rot wird.

### 3.16 Randformen stehen als Raster im Plan, nicht als Aufzählung   (seit welle-persistenz, geschärft seit slice-fehlermodell-konfiguration und seit welle-record-replay)

Schärft §3.12: §6 des Plans führt die Randformen als Raster *Operation × Argumentart* (fehlender, leerer, unbekannter oder fremder Wert; Abbruch; Wiederholung), nicht als Liste der Fälle, die dem Planner einfielen.
Jede Zelle nennt ihren Ort in der Spezifikation oder in `harness/vertraege/`, oder sie nennt ausdrücklich „gilt nicht" mit Grund; eine leere Zelle ist ein Fehler des Plans.
Eine Zelle nennt nur einen Ort, dessen Wortlaut die Entscheidung trägt.
Meldet die Operation mehrere Verstöße oder hängt eine Zusage vom Modus ab, führt das Raster zusätzlich eine Zusatztabelle mit der Spalte *Fänger*, darin je Paar von Verstößen am selben Schlüssel und je Modus eine Zeile; eine Allaussage des Vertrags („kein Fall …“) nennt ihren Fänger über alle Fälle, sonst ist sie eine Zusage ohne Fänger (§3.15).
Der Implementer hält an einer Zelle ohne Ort an (§3.12); der Reviewer löst das Raster mit je einer Mutation auf, und eine im Code entschiedene Zelle ist ein MEDIUM.

**Falsch:** Der Vertrag nennt die leere ID für `Get`; für `Delete` und `Advance` entscheidet der Code.
**Richtig:** Das Raster hat für jede Operation eine Zeile, und jede Zelle verweist auf eine Nummer des Vertrags.

## 4. Quality Gates

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-harness-dateien.md`
§harness/README.md als Einstiegspunkt. Der Gate-Index steht **einmal**, in
[`harness/README.md`](harness/README.md) §Sensors — dort steht auch die
*Bindung* jedes Targets; Targets aus Werkzeug-Fragmenten stehen in dem Teil des
Werkzeugs, den §Sensors verlinkt. Diese Datei führt die Liste nicht.

Kein Target nennen, das im Makefile nicht existiert — auch nicht in Prosa.

## 5. Dokumentations-Regeln



| # | Regel | Datei |
|---|---|---|
| 1 | **Anforderungs-IDs und ADR-Nummern** müssen in PRs/Commits referenziert sein — sie sagen, welche Zusage oder Entscheidung berührt ist. Struktur-IDs (`SPEC-<NNN>`, `ARC-<NNN>`) adressieren *innerhalb* der Spec und gehören nicht in die Commit-Message. | — |
| 2 | Vergeben werden IDs beim Spec-/ADR-Schreiben nach dem in `harness/conventions.md` deklarierten ID-Schema (Default: `<PREFIX>-FA-<NN>` / `<PREFIX>-QA-<NN>` / `<PREFIX>-RB-<NN>` aus dem Lastenheft, `SPEC-<NNN>` in der Spezifikation, `ARC-<NNN>` in der Sicht, ADR-Nummern über den ADR-Index) — nie ad hoc im PR. | — |
| 3 | Neue ADRs müssen den ADR-Index aktualisieren. | — |
| 4 | Roadmap/Status-Geschichte lebt in `docs/plan/planning/`, nicht in `spec/architecture.md`. | — |

## 6. Minimal Agent Workflow

Pro Slice:

1. `harness/README.md` lesen.
2. Relevante kanonische Quelle lesen (Source Precedence beachten).
3. Betroffene Requirement-/ADR-IDs identifizieren.
4. Kleinste sinnvolle Änderung planen.
5. Engsten nützlichen Sensor laufen lassen.
6. Repo-weiten Gate-Lauf vor Handoff (`make gates`).
7. Doku/Indizes aktualisieren, falls ein öffentlicher Vertrag berührt.
8. Ausgeführte Sensors und verbleibende Risiken berichten — keine Erfolgsmeldung ohne Gate-Ausführung.

Dieser Workflow deckt ausschließlich die Implementer-Rolle ab. Schritt 8
ist der Rollenwechsel, kein Abschluss: Bericht → Handoff an Reviewer
(`.harness/skills/reviewer.md`, siehe `harness/README.md` §Guides) →
Verifier. Kein Self-Review — anderer Kontext findet andere Findings,
derselbe Kontext dieselben blinden Flecken (Baseline-Regelwerk
`modul-08-agentenrollen.md`).
