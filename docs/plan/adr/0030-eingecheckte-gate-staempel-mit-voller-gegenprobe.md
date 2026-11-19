# ADR-0030: Je Gate ein eingecheckter Stempel über seine Eingaben, `make gates-voll` ist die Gegenprobe dazu

**Status:** Proposed

**Datum:** 2026-11-17

**Autor:** pt9912

**Bezug:** [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [ADR-0003](0003-lint-profil-golangci.md), [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0007](0007-race-detektor-im-gate.md), [ADR-0009](0009-groessenmarke-als-gate-mit-messung-durch-den-host.md), [ADR-0010](0010-groesse-der-guide-dateien-als-gate.md)

**Schärft:** `SPEC-213` (künftiger Vertrag der Gate-Stempel unter harness/vertraege, noch nicht angelegt; die Nummer ist die nächste freie im Bereich der Linie C nach `SPEC-212`)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

`make gates` fährt jedes Gate bei jedem frischen Lauf vollständig; der erste Lauf in einem Klon oder Container ohne warmen Docker-Cache dauert lange (die Dauer ist nicht gemessen).
Bei unveränderter Eingabe spart nur der Docker-Cache des eigenen Hosts Arbeit, und der reist nicht mit.
Zwei Mechanismen für den Nachweis liegen vor: der Stempel von `make gegenproben` ([ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md)), eingecheckt, nur für dieses eine Ziel; und der Nachweis von `record-gates` (`.harness/state/`), nur lokal, vom Stop-Hook gelesen, kein Beleg für andere Klone.
Die Beobachtung `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` zeigt die Schwäche eines Stempels, der am Arbeitsbaum entsteht: er gehört zu keinem Commit, wenn danach ein Eingabe-Commit folgt oder der Lauf ungetrackte Dateien sah.
Eine CI gibt es nicht; sie ist ein eigener Folge-Slice.
Annahmen: Ein Stempel je Gate belegt „dieses Gate war an dieser Eingabe grün", nicht „der Stand ist in Ordnung"; die Eingabemenge eines Gates ist vollständig benennbar; ein Nachtlauf ohne Stempel (voller Lauf auf `main`) ist verfügbar, sobald die CI steht.
Kippt die zweite Annahme, ist ein Stempel ein Gate-Lockerung durch die Hintertür (`AGENTS.md` §3.6).

## Entscheidung

Wir führen **je Gate einen eingecheckten Stempel** unter der Ordner harness/stempel ein, eine Datei je Gate, benannt nach dem Target; sie hält den Hash der deklarierten Eingaben des Gates.
`make gates` rechnet den Hash je Gate nach und lässt ein Gate bei gleichem Hash aus; läuft es grün, erneuert es **seinen** Stempel, auch wenn ein anderes Gate rot ist; ein rotes Gate lässt seinen Stempel unverändert.
**`make gates-voll`** fährt jedes Gate ohne Stempel und schreibt keinen Stempel; es ist die Gegenprobe zu den Stempeln: Ein Gate, das rot ist, obwohl sein Stempel zum Stand passt, ist ein **falsch gesetzter Stempel** und wird als solcher gemeldet.
Der Nachtlauf (`make gates-voll` auf `main`) und die billige Hash-Prüfung je Push oder PR sind **Slice (b)**; bis dahin ersetzt der volle Lauf vor Handoff und Closure den Nachtlauf.
Entscheidungen zu den Randformen:

| Randform | Entscheidung | Ort der Regel |
|---|---|---|
| Stempel als Repo-Dateien | Der Ordner harness/stempel steht in keiner Eingabemenge und nicht im Hash von `tools/harness/working-tree-hash.sh`, damit der Stempel weder `docs-check`, `zeichen-check`, `groessenmarke` noch den Nachweis des Stop-Hooks berührt; die Dateien bestehen `zeichen-check` | `SPEC-213` |
| Erneuerung je Gate, nicht je Lauf | ein grünes Gate schreibt seinen Stempel unabhängig von den anderen; Fänger je Gate | `SPEC-213` |
| Ungetrackte, nicht ignorierte Dateien | sie gehen in den Hash ein, weil die Gates sie sehen; der Stempel gilt dem Inhalt, nicht dem Commit, und ein Commit mit gleichem Inhalt behält ihn | `SPEC-213` |
| Reihenfolge mit `record-gates` | der Ordner harness/stempel ist aus `working-tree-hash.sh` ausgenommen; die Reihenfolge der Schreiber ist damit gleichgültig | `SPEC-213`, `tools/harness/working-tree-hash.sh` |
| Vollständigkeit der Eingaben | eine Deklaration an einer Stelle, eine Zeile je Gate; die Zeile selbst geht in den Hash ein; je Eingabe ein Fänger | `SPEC-213` |
| Gate mit Wahrscheinlichkeit | `make test-race` bekommt keinen Stempel | `SPEC-213`, `SPEC-122` |
| Merge-Konflikt am Stempel | nie von Hand lösen: die Datei verwerfen und mit `make gates` neu erzeugen; ein Gate, das nicht erzeugen kann, ist rot | `SPEC-213` |
| Gegenproben-Stempel | `harness/gegenproben.stempel` **bleibt außen vor** | diese ADR |

Welche Gates einen Stempel tragen, die Eingaben je Gate, das Format der Datei, die Meldungen und das Verhalten von `gates-voll` im Einzelnen stehen in `SPEC-213`, nicht hier.

**Einordnung nach `AGENTS.md` §3.6.** Ein Stempel, der einen Lauf bei unveränderter Eingabe überspringt, senkt keine Schwelle: das Gate prüft denselben Stand mit derselben Strenge, nur nicht ein zweites Mal.
Eine zu kleine Eingabemenge dagegen lässt ein Gate bei geänderter Eingabe aus und ist faktisch eine Lockerung; deshalb ist die Eingabemenge je Gate Teil des Vertrags: Erweitern ist frei, Kürzen braucht eine ADR.
Was die Kontrolle abfängt:

| Fehler | Fänger | Fenster |
|---|---|---|
| Stempel veraltet (Hash passt nicht zum Stand) | Hash-Prüfung je Push (Slice b); lokal `make gates-voll` | bis zum nächsten Push |
| Stempel gefälscht oder bei rotem Gate gesetzt (Hash passt, Gate rot) | nur der volle Lauf: `make gates-voll`, nachts auf `main` | bis zu rund 24 h auf `main` (ab Slice b), davor bis zum nächsten vollen Lauf vor Handoff |
| Eingabe fehlt in der Deklaration, Gate bleibt grün | der Nachtlauf sieht sie nur, solange das Gate rot ist; der Fänger je Eingabe in den Gegenproben | nicht anders als durch den Fänger geschlossen |

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A0 — nichts tun, Docker-Cache genügt | keine Mechanik, kein Risiko eines falschen Stempels | der Cache reist nicht mit: jeder frische Klon und jeder Container zahlt den vollen Lauf; der Nachweis `record-gates` ist lokal |
| A — nur `make gates-voll` schreibt Stempel | jeder Stempel stammt aus einem echten vollen Lauf; kein falscher Stempel durch einen Fehler der Eingabemenge | nach jeder Änderung ist der Stempel veraltet, bis jemand den vollen Lauf fährt; der normale Lauf spart nichts, der Nutzen entfällt |
| C — nur die CI schreibt Stempel | der Stempel stammt aus einer Umgebung, die kein Entwickler fälschen kann | die CI fehlt; Bot-Commits auf `main`; lokal spart nichts, weil der Stempel dem Commit hinterherläuft |
| **B — `make gates` erneuert je Gate nach echtem Lauf, `gates-voll` und Nachtlauf kontrollieren (gewählt)** | ein Lauf spart, sobald ein Stempel passt, auch im frischen Klon; Erneuerung dort, wo der Lauf stattfindet; die Gegenprobe bleibt vorhanden | ein falscher Stempel ist möglich (Fehler in der Eingabemenge, Handarbeit) und wird erst vom vollen Lauf gesehen; Stempel-Dateien im Verlauf; Merge-Konflikte an ihnen |
| Eingaben zentral (eine Deklaration) | ein Ort für Hash-Prüfung und Fänger; die Prüfung je Push rechnet ohne die Fragmente nach | Eingabe und Rezept driften auseinander, wenn jemand nur eines ändert |
| Eingaben verteilt (je `harness/mk`-Fragment) | die Eingabe steht beim Rezept | die Prüfung je Push muss jedes Fragment lesen; kein Überblick, ob eine Eingabe fehlt; kein Fänger über die Menge |

Der Gegenproben-Stempel zieht **nicht** um; eine Alternative der Pfad harness/stempel/gegenproben.stempel wäre:

| Option | Pro | Contra |
|---|---|---|
| Umzug nach der Ordner harness/stempel | ein Ordner, eine Regel für Merge und Ausnahme aus dem Hash | [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md) nennt den Pfad in seinem Entscheidungssatz; der Umzug machte eine Accepted-ADR falsch, und sie nur dafür abzulösen ist unverhältnismäßig; rund zwei Dutzend Verweise in offenen Plänen (`AGENTS.md` §3.9) |
| **bleibt außen (gewählt)** | keine Änderung an [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md), `SPEC-102` und den Plänen; die Eingabe (Pins als Argumente) und der Charakter (kein Gate) sind ohnehin andere | zwei Orte für Stempel; die Merge-Regel gilt für beide, steht aber nur für die neuen in `SPEC-213` |

## Konsequenzen

- Positiv: Ein Klon mit passenden Stempeln fährt die Gates mit Stempel nicht neu; ein Fehler der Eingabemenge ist an einer Stelle zu finden; die Gegenprobe zu den Stempeln hat einen Namen und einen Lauf.
- Negativ: Ein Stempel ist eine Behauptung, die ein Mensch fälschen kann; bis die CI steht, fängt das nur der volle Lauf vor Handoff. Der Nutzen gegenüber dem Docker-Cache ist nicht gemessen und liegt in frischen Klonen und Containern; ohne Messung vor Slice (a) ist er eine Annahme. `make test-race` bleibt ohne Stempel und damit der teuerste Teil jedes Laufs: ein Rennen, das nur zufällig auftritt, ließe einen Stempel ein grünes Zufallsergebnis festhalten, und der Detektor ist genau für dieses Gate da ([ADR-0007](0007-race-detektor-im-gate.md), `AGENTS.md` §3.10); der Docker-Cache hält ein solches Ergebnis heute schon lokal, ein Stempel reichte es in jeden Klon.
- Folgepflicht: Slice (a): `make gates-voll`, Stempel je Gate, Vertrag `SPEC-213` samt Fängern und Gegenproben, Änderung an `tools/harness/working-tree-hash.sh` (Träger des Stop-Hooks; Zustimmung des Nutzers). Slice (b): CI mit Nachtlauf `make gates-voll` und Hash-Prüfung je Push oder PR. Die Übergangsregel Nr. 2 in `docs/plan/planning/README.md` (Stempel allein im Integrationsschritt) und `harness/README.md` §Sensors ziehen mit.

## Offene Fragen vor `Accepted`

Diese Entscheidungen stehen beim Nutzer; der Entwurf nimmt keine davon vorweg.

1. **Messung zuerst.** [ADR-0007](0007-race-detektor-im-gate.md) sagt, ein unveränderter Stand koste durch den Docker-Cache nichts; der Nutzen der Stempel liegt in frischen Klonen und Containern. Soll Slice (a) mit einer Messung des Kaltlaufs je Gate beginnen? Welche Gates einen Stempel tragen, sollte auf dieser Messung beruhen.
2. **`gates-voll` bei grünem Gate mit veraltetem Stempel:** rot (Vorschlag: „Stempel veraltet, `make gates` erneuert"; Reihenfolge dann `make gates`, danach `gates-voll`), oder lokal nur eine Warnung und rot erst im Nachtlauf.
3. **Gate-Auswahl.** `test-race` ist entschieden (ohne Stempel). Offen sind `groessenmarke` (liest Git-Historie und Pläne, kein reiner Dateihash) und `baseline-verify` (billig); Vorschlag: beide ohne Stempel.
4. **Fährt `make gegenproben` in `gates-voll` mit?** Vorschlag: nein, es bleibt der eigene Mechanismus aus [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md) und gehört in die Pflichtliste der vollen Probe vor Handoff. Der Betrieb stützt das: Auf dem geteilten Host ist `make gegenproben` zurzeit auch bei ruhigem Host rot, weil seine neun `lint`-Fälle die Last selbst erzeugen (`BEO-REPO/gegenproben-erzeugen-die-last-selbst`).
5. **Zustimmung zur Änderung von `tools/harness/working-tree-hash.sh`.** Es ist die gemeinsame Quelle von `record-gates` und dem Stop-Hook; die Ausnahme des Ordners macht handgeschriebene Stempel für den Nachweis unsichtbar, und das Risiko deckt nur der Nachtlauf.
6. **Ablauf mit der Übergangsregel Nr. 2.** Sollen die Gate-Stempel dem Stempel der Gegenproben folgen (allein im Integrationsschritt nach dem letzten Eingabe-Commit) oder im Slice-Zweig mitlaufen? Letzteres erzeugt mehr Merge-Konflikte (dafür gilt „neu erzeugen"). Vorschlag: dieselbe Regel.
7. **Namen und Zählung.** Die Namen der beiden Folge-Slices (Gate-Stempel mit `gates-voll`; CI mit Nachtlauf) und ob `SPEC-213` unter harness/vertraege liegt, bestätigt der Planner.

**Aus dem Betrieb, bevor Slice (a) beginnt:**

- Jede Maschine der Gruppe läuft in einem eigenen Container mit eigenem Dateisystem und eigenem Klon; ausgetauscht wird nur über `origin`. Ein eingecheckter Stempel reist deshalb mit dem Commit in jeden Klon, eine Sperrdatei im Dateisystem verbindet keine Maschinen. Last und Docker-Daemon des Hosts sind dagegen geteilt.
- Slice (a) ändert die Gate-Kette (`Makefile`, `harness/mk/`, `harness/README.md`); nach Nr. 7 der Übergangsregeln läuft nie zugleich ein anderer Kettenslice. Er beginnt erst, wenn die Entscheidung zu den Gegenproben (Parallelität der `lint`-Fälle, Messvorschrift, Soll) gefallen ist und kein anderer Kettenslice läuft.
- Das Sitzungsende der Gruppe hat diesen Entwurf aus dem Scratchpad in das Repo gerettet; der Planner und der Nutzer führen ihn fort.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `tools/harness` (Stempel-Skript, Name offen) und `tools/gegenproben` | je Eingabe eines Gates: Datei ändern, der Stempel gilt nicht mehr und das Gate läuft; ein grünes Gate erneuert nur seinen Stempel; ein rotes lässt ihn | `make gegenproben` |
| Stempel-Skript im Lauf aller Gates | ein Gate, rot trotz passendem Stempel, wird als falsch gesetzter Stempel gemeldet | `make gates-voll` |

## Re-Evaluierungs-Trigger

Wenn die Messung vor Slice (a) keinen Nutzen gegenüber dem Docker-Cache zeigt (ein frischer Klon spart weniger als eine Minute), wenn ein falscher Stempel auf `main` auftritt (zweimal), oder wenn die CI steht und einen Cache der Gate-Läufe trägt, der die Stempel ersetzt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-17 | Proposed | Entwurf des Architect; die Annahme steht beim Nutzer aus |
| 2026-11-18 | Nummer 0030 aus dem Bereich der Linie C, offene Fragen aufgenommen | Ablage im Repo, kein Slice gestartet |

Nach `Accepted` wird diese Datei **nicht mehr inhaltlich überschrieben**.
Spätere Korrekturen oder Schärfungen entstehen als neue ADR mit
`Supersedes ADR-NNNN` (Baseline-Regelwerk `modul-04-adrs.md`
§Hard Rule für Accepted-ADRs).
