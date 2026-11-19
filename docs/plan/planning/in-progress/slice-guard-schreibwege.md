# Slice slice-guard-schreibwege: Command-Guard sperrt die Schreibwege auf Repo-Dateien

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice und keine Closure-Bedingung, die von der DoD verschieden wäre (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) gilt nicht mehr: der Architect wählt keine Gegenprobe (Auftrag 4), und der Slice berührt keine der Dateien, die Nr. 7 der Gate-Kette zuweist (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel); das schreibt der Vertrag fest ([SPEC-214](../../../../harness/vertraege/SPEC-214-command-guard.md) Nr. 10 d).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Werkzeuge der Rollen, der Bezug ist der nächste, den es gibt), `AGENTS.md` §3.1 (Docker-only, der Guard ist sein Träger) und §3.9 (die Betriebsregel gegen `sed -i` und Heredoc steht in den Übergangsregeln Nr. 8), Register-Eintrag `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (17 Belege am 2026-11-08, Ausgang `geplant` in diesem Slice; Belege u. a. `slice-baseline-import`, `slice-unsichtbare-zeichen`, `slice-v2-metadaten`, `slice-workspace-use-case`, `slice-edm-ausdruecke`, `slice-bootstrap-m3-e2e-mock`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (der Prüflauf ist keine Gegenprobe, Grund in `SPEC-214` Nr. 10 d).

**Übernimmt:** die Sendung aus der Closure von `welle-mock` (Geber `welle-mock`, Punkt „Betriebsregel nur `Edit` und `Write`“ in der Closure-Notiz, Entscheidung des Users vom 2026-11-08, Punkte lösen): die Formen `git apply` (Patches der Architects ohne `Edit`), `pgrep -f`, `kill` und `rm -rf` mit Muster oder auf Verzeichnisse, die der Lauf nicht angelegt hat, und die generischen Namen im gemeinsamen Scratchpad (`mut/`, `gt.sh`, `fresh.sh`), die der Plan bis zum 2026-11-08 nicht nannte (Evidenz `slice-bootstrap-m3-e2e-mock`, `slice-bootstrap-m3`, `slice-edm-ausdruecke`, `slice-snapshot-use-case`).

**Berührte Spec-Stellen:** [SPEC-214](../../../../harness/vertraege/SPEC-214-command-guard.md) (neuer Vertrag des Guards, Kennung aus dem Bereich der Linie C, vom Architect vergeben und geschrieben; der Code folgt ihm) · [harness/vertraege/README.md](../../../../harness/vertraege/README.md) §Verträge (Zeile steht). `SPEC-102` bleibt unverändert (keine Gegenprobe).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Planner der Closure von `welle-record-replay`). **Datum:** 2026-11-06.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Der PreToolUse-Guard `.claude/hooks/pretooluse-command-guard.sh` sperrt `sed -i` und das Schreiben einer Repo-Datei per Heredoc oder Umleitung (`>`, `>>`, `tee`) im Befehl einer Rolle, mit einer Meldung, die auf `Edit` und `Write` zeigt; die Betriebsregel „auf Repo-Dateien nur `Edit` und `Write`“ (Übergangsregeln Nr. 8) hat damit einen Wächter statt der Disziplin der Rollen.

**Hintergrund:** 17 Vorgänge haben die Regel gebrochen (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`, Stand 2026-11-08); kein Gate und kein Wächter erzwang sie, der Guard sperrt heute Host-Toolchains und `python`, nicht `sed`. Die Folge zeigte sich erst am Diff. Neben `sed -i` und Heredoc zeigt das Register weitere Formen: `git apply` und `perl`, `head`/`tail`/`cat` als Schreibweg der Architects, deren Rollen-Typ kein `Edit` hat (die Ursache liegt in der Werkzeugliste des Typs, nicht in einer gebrochenen Regel), `gofmt -w` auf Repo-Dateien, `pgrep -f`/`kill` und `rm -rf` auf Prozesse und Verzeichnisse fremder Läufe und generische Namen im gemeinsamen Scratchpad, die ein anderer Lauf überschrieb (`mut/`, `gt.sh`, `fresh.sh`).

**Erweiterung des Umfangs (Planner, 2026-11-08, Entscheidung des Users „Punkte lösen“):** Der Schnitt gilt dem Guard als Stolperdraht in drei Gruppen: (1) Schreibwege auf Repo-Dateien (`sed -i`, Heredoc, `>`, `>>`, `tee`, **`git apply`**), (2) Prozess- und Löschwege auf fremde Gegenstände (**`pgrep -f` mit Muster, `kill`/`pkill` und `rm -rf` mit Muster oder auf ein Verzeichnis außerhalb des eigenen Laufs**), (3) **generische Namen im gemeinsamen Scratchpad** (`mut/`, `gt.sh`, `fresh.sh`; der Name trägt die Kennung des Laufs oder der Rolle).
**Entschieden (Architect, 2026-11-19, [SPEC-214](../../../../harness/vertraege/SPEC-214-command-guard.md)):** (1) sperrt der Guard (Nr. 4, 6). (2) sperrt `pkill`, `killall`, `pgrep -f` und rekursives `rm` außer unterhalb eines Scratchpad-Verzeichnisses; `kill` bleibt durchgelassen (Nr. 7, 11 b). (3) ist eine Anweisung der Rollen, kein Prüfpunkt des Guards (Nr. 8, 11 i). Der Umfang bleibt ein Slice: Schätzung *S* = 1900 liegt unter der Schnittgrenze (§3, §4).

**Warnung an den Architect (Pflicht vor dem Entwurf):** `git commit -F -` mit Heredoc oder Pipe auf der Standardeingabe sowie ein Heredoc, das eine Commit-Message nach `git commit -F` liefert, **bleiben erlaubt**. Der Guard ist quote-blind und trennt Segmente an `;`, `&`, `|`, `(` (Kopf des Skripts); eine Regel, die jedes Heredoc oder jedes `>` sperrt, bricht die Commit-Wege, die `AGENTS.md` und `implement-slice` vorschreiben, und Schreibwege in das Scratchpad (`/tmp/claude-…`) und in `/dev/null`, die erlaubt sind. Der Architect entscheidet, **welche Ziele** als Repo-Datei gelten (Pfad unter dem Wurzelverzeichnis des Repos) und wie der Guard sie erkennt, ohne das Ziel einer Umleitung falsch zu lesen. *(Erledigt: `SPEC-214` Nr. 3 bis 5; ein Heredoc ist Eingabe, nur eine Umleitung oder `tee` schreibt.)*

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Sandbox oder eine vollständige Erkennung aller Schreibwege (`perl -i`, `python` über eine andere Tür, `cp`, `mv`, `dd`, `install`) — der Guard ist ein Stolperdraht, keine Sandbox (Kopf des Skripts); die ungefangenen Wege stehen als akzeptiertes Negativ mit Grund und Folge in `SPEC-214` Nr. 11. `git apply`, `pgrep -f`/`kill`/`rm -rf` und die Scratchpad-Namen gehören seit dem 2026-11-08 nicht mehr zu diesem Ausschluss, sondern zum Umfang (Aufträge 7 und 8).
- Eine Änderung der Werkzeugliste der Rollen-Typen (`.claude/agents/architect.md` hat kein `Edit`; die Datei steht auf ihrer Größengrenze) — Entscheidung des Users nach einem Diff-Vorschlag des Architect (§6, Rückgabe 2); ein Wächter, der `git apply` sperrt, ohne dass der Architect ein `Edit` hat, sperrt seine einzige Schreibform (Auftrag 7, entschieden: er schreibt ganze Dateien mit `Write`, `SPEC-214` Nr. 12).
- Eine Sperre im Dateisystem oder im Container — andere Mechanik, andere Schicht; Sache einer eigenen Entscheidung.
- Die Änderung der Regel selbst (Übergangsregeln Nr. 8) — sie gilt unverändert; dieser Slice gibt ihr einen Wächter.
- Die Stop-Hook-Kette (`stop-require-gates.sh`) und `span-emit.sh` — anderer Zweck.
- Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills — nur als Diff-Vorschlag in der Closure, der User ändert.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [ ] Der Guard erfüllt `SPEC-214`: er sperrt `sed -i` (auch `-i.bak`, `--in-place`, in Flag-Bündeln), `git apply` und das Schreiben per Umleitung (`>`, `>>`, `&>` und Verwandte, auch mit Heredoc) und `tee` auf einen Pfad unter dem Wurzelverzeichnis des Repos, sperrt `pkill`, `killall`, `pgrep -f` und rekursives `rm` außer unterhalb eines Scratchpad-Verzeichnisses, lässt `kill` durch, und lässt `git commit -F -` samt Heredoc für die Message, Umleitungen in `/dev/null` und außerhalb der Wurzel durch; die Regel für Scratchpad-Namen mit Lauf-Kennung steht als Anweisung in den Übergangsregeln Nr. 8 (`SPEC-214` Nr. 8); jede Zusage des Vertrags hat eine Zeile in der Tabelle *Zusage · Fänger* (§3) mit einer vom Implementer rot gesehenen Mutation.
- [ ] Der Prüflauf `bash tools/harness/guard-faelle.sh` (kein Gate, `SPEC-214` Nr. 10) belegt Sperre und Durchlass über die 52 Fälle des Vertrags; die Zeit des Skripts und die Latenz des Guards sind gemessen und in §7 berichtet.
- [ ] Doku-Update: Übergangsregeln Nr. 8 (`docs/plan/planning/README.md`) nennen den Wächter statt der Disziplin und tragen die Namensregel; `harness/README.md` bleibt unverändert (kein Ziel).
- [ ] `make gates` grün.
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
| `.claude/hooks/pretooluse-command-guard.sh` | update (Implementer) | Erkennung nach `SPEC-214` Nr. 1 bis 9: Umleitung, `tee`, `sed -i`, `git apply`, `pkill`/`killall`/`pgrep -f`, rekursives `rm`; Zitat-, Heredoc- und Verzeichnis-Zustand; Länge; vier neue Meldungen; Kopfkommentar nach `AGENTS.md` §3.7 und §3.11 nachgezogen |
| `tools/harness/guard-faelle.sh` | neu (Implementer) | Prüflauf nach `SPEC-214` Nr. 10: die 52 Fälle, rund 300 Elemente, Latenz; `#!/usr/bin/env bash`, `set -euo pipefail`; kein `make`-Ziel |
| `harness/vertraege/SPEC-214-command-guard.md`, `harness/vertraege/README.md` | neu und update (Architect, erledigt) | Zusagen, Randformen, Fälle, *Ungeprüft* |
| `docs/plan/planning/README.md` | update (Implementer) | Übergangsregeln Nr. 8: der Wächter statt der Disziplin; die Namensregel aus `SPEC-214` Nr. 8 im Wortlaut des Vertrags; Nr. 7 bleibt |
| `tools/gegenproben/`, `harness/mk/gegenproben.mk`, `tools/harness/gegenproben-eingabe.sh`, `harness/README.md` | entfällt | Auftrag 4: kein Gate, keine Gegenprobe, kein Ziel (`SPEC-214` Nr. 10 d) |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Vertrag, Entscheidungen, Tabelle *Zusage · Fänger*, Raster, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18); Dateien und Verzeichnisse im Scratchpad nur unter `<kurz>-<rolle>/` (`SPEC-214` Nr. 8), in diesem Slice `guard-schreibwege-<rolle>/`. Der Implementer fährt `bash tools/harness/guard-faelle.sh` vor der Übergabe, Reviewer und Verifier fahren es selbst; das Skript ruft den Guard des eigenen Arbeitsbaums, also den, den der Implementer schreibt.

**Aufträge an den Architect** (alle entschieden, Fundstellen in §6):

1. **Zielerkennung.** Wann ein Ziel als Repo-Datei gilt (absoluter Pfad unter dem Wurzelverzeichnis, relativer Pfad, `~`, Variablen, Umleitung mit Dateideskriptor, `&>`); wie der Guard das Ziel einer Umleitung findet, obwohl er quote-blind ist; das Verhalten bei Parse-Zweifel (fail-closed wie heute für den Extraktor oder durchlassen).
2. **Durchlass.** `git commit -F -` mit Heredoc oder Pipe, Heredoc für eine Datei im Scratchpad, `/dev/null`, `2>&1`; die Menge der erlaubten Ziele als Liste mit Grund. Die Erlaubnis für Commit-Messages ist Pflicht (Warnung in §1).
3. **Menge der Schreibwege.** `sed -i`, `sed --in-place`, Flag-Bündel (`-ni`, `-i.bak`), `tee`, `>`, `>>`, `<<`, `<<<`; welche übrigen (`cp`, `mv`, `dd`, `install`, `perl -i`) als akzeptiertes Negativ mit Grund und Folge stehen (`AGENTS.md` §3.15).
4. **Prüflauf.** Gegenprobe (berührt die Gate-Kette, Nr. 7) oder ein Test des Skripts ohne Gate; wenn Gegenprobe, die Fälle und das Zeitbudget (`SPEC-102`).
5. **Randformen** vor dem ersten Code-Commit im Vertrag (Raster in §6): leerer Befehl, nur Whitespace, sehr langer Befehl, Verschachtelung (`bash -c "sed -i …"`, Tiefe wie heute), mehrere Segmente, Zeilenumbruch im Befehl.
6. **Tabelle *Zusage · Fänger*** (§3), **Raster** (§6), **erwartete Produktcode-Größe** für die Marke (Skript, kein Go; ob die Messung zählt, entscheidet `SPEC-128`).
7. **`git apply` und die Prozess- und Löschwege** (Sendung aus `welle-mock`). `git apply` mit Patch aus Datei oder Standardeingabe als Schreibweg; die Architects ohne `Edit` schrieben so (Evidenz `slice-edm-ausdruecke`, `slice-snapshot-use-case`), eine Sperre ohne `Edit` im Rollen-Typ nimmt ihnen den Weg (siehe Ausschluss). `pgrep -f` mit Muster, `kill`, `pkill`, `rm -rf` mit Muster oder Verzeichnis außerhalb des Scratchpad-Unterverzeichnisses des eigenen Laufs: sperren, melden oder Negativ mit Grund und Folge; die Erkennung des Ziels ist quote-blind (Auftrag 1).
8. **Scratchpad-Namen.** Eine Regel, dass jede Datei oder jedes Verzeichnis, das ein Lauf im gemeinsamen Scratchpad anlegt, die Kennung des Laufs oder der Rolle im Namen trägt (`mut/`, `gt.sh`, `fresh.sh` sind die Fälle), und ob der Guard sie erkennt oder die Anweisung der Rollen sie trägt; das Gegenstück (Prozess-Muster beschränkt auf eigene Namen) gehört zu Auftrag 7.

**Entscheidungen des Architect zu den Aufträgen 1 bis 8** (Ort im Vertrag `SPEC-214`; keine neue ADR: kein Accepted-ADR wird berührt, der Guard ist kein Gate und senkt keines, `SPEC-214` Nr. 1 e; die Bindung folgt [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)):

| Auftrag | Entscheidung | Ort |
|---|---|---|
| 1 Zielerkennung | Ein Ziel ist *Repo* (relativ, absolut unter der Wurzel auf der Komponentengrenze, `~`/`$HOME` unter der Wurzel, `$PWD` und `$CLAUDE_PROJECT_DIR`), *außen* oder *Zweifel*; der Guard liest Operatoren zeichenweise mit Zitat- und Backslash-Zustand je Segment und führt einen Verzeichnis-Zustand für `cd`. Im Zweifel am Ziel geht der Befehl durch; fail-closed bleiben nur Extraktor, Tiefe und die neue Längengrenze 32768. | Nr. 2 e, 3, 4, 9 |
| 2 Durchlass | Ein Heredoc ist Eingabe, nur Umleitung und `tee` schreiben. Durch gehen die Commit-Wege, jedes Ziel außerhalb der Wurzel (Scratchpad, `/tmp`, `/dev/null`), `2>&1` und Verwandte; der Rumpf eines Heredoc wird nicht als Befehlsfolge gelesen, außer eine Shell liest ihn oder er ist nicht terminiert. | Nr. 5, 4 b |
| 3 Menge | Gesperrt: Umleitungen (`>`, `>>`, `>` mit Strich, `&>`, `&>>`, `>&Wort`, mit Deskriptor-Zahl), `tee`, `sed -i` in allen Formen auch auf dem Scratchpad, `git apply`. Negativ mit Grund und Folge: `cp`, `mv`, `dd`, `install`, `perl -i`, `patch`, `ed`, `gofmt -w`, `find -exec`, `git am`, Schreibwege hinter einer zweiten Tür, `<>`. | Nr. 6, 11 a, d, e |
| 4 Prüflauf | Skript `tools/harness/guard-faelle.sh` ohne Gate und ohne `make`-Ziel (kein Lauf der Gegenproben-Stufe je Änderung am Hook, kein Besitzer der Gate-Kette); 52 Fälle, rund 300 Elemente; höchstens 60 s, Median der Latenz höchstens 100 ms. Eine Regression fängt nur ein Lauf; Auslöser für ein Gate ist ein ungefangener Bruch. | Nr. 10 |
| 5 Randformen | Raster in §6; leer, Leerraum, Länge (32768 durch, 32769 gesperrt G), Verschachtelung (Tiefe 3 geprüft, 4 fail-closed), mehrere Segmente, Zeilenumbruch, CR und Fortsetzung. | Nr. 2, 9, §6 |
| 6 Tabelle, Raster, Größe | Tabelle unten (20 Zeilen), Raster und Zusatztabelle in §6, Größe: *a* = 0, *b* = 20, *f* = 0, *p* = 300. | §3, §4, §6 |
| 7 `git apply`, Prozess, Löschen | `git apply` gesperrt (nur lesend und `-C` nach außen durch), `pkill`, `killall`, `pgrep -f` gesperrt (P), rekursives `rm` gesperrt außer unter `…/scratchpad/<Name>` (L), `kill` durch (Negativ). Der Architect-Typ ohne `Edit` schreibt ganze Dateien mit `Write`; Diff-Vorschlag an den User in §6, Rückgabe 2. | Nr. 6 c, 7, 11 b, 12 |
| 8 Scratchpad-Namen | Anweisung `<kurz>-<rolle>/` für jeden Lauf, in den Übergangsregeln Nr. 8 und in den Betriebsregeln des Plans; der Guard prüft keine Namen (ein „generischer“ Name ist kein Wortmuster). Die Rollen-Anweisungen schreibt der User (Diff-Vorschlag, §6, Rückgabe 3). | Nr. 8, 11 i |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect). Ein Fall ist eine Nummer aus `SPEC-214` Nr. 10 e, ein Element ein Befehl darin; Mengen sind Schleifen über alle ihre Elemente. Der Implementer trägt die roten Läufe nach (Zusage · Mutation · roter Test, `AGENTS.md` §3.10) und liest die Mutation je Zeile als Klasse, nicht als genannte Form (§3.17). Der Guard ist zustandslos und einfädig; die Zusagen machen keine Aussage über Nebenläufigkeit, keine Zeile trägt eine Rate, jede Zeile ist deterministisch (Zeile 17 misst Laufzeit, sie ist mit Reserve gesetzt).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Eine Umleitung mit Ziel *Repo* sperrt (S), in jeder Operatorform, an jedem Befehl und an jeder Stelle der Umleitungen eines Segments (Nr. 4 a, d) | Fälle 1 und 2 | `>>` oder `&>` aus der Erkennung nehmen: Fall 1 Element 2 bzw. 7 rot; nur die erste Umleitung prüfen: Fall 2 Element 7 rot; den Strich oder das `&` hinter `>` als Trenner lesen: Fall 1 Elemente 9 und 10 rot | andere |
| 2 | Eine Umleitung ohne Ziel *Repo* geht durch: Dup, Gerät, außen, kein Ziel, Zweifel (Nr. 4 b, c, 5 f) | Fälle 3 und 9 | `2>&1` als Ziel `1` lesen: Fall 3 Element 5 rot; ein Operator ohne Ziel sperrt: Fall 9 Elemente 6 und 7 rot; Zweifel sperrt: Fall 9 rot | andere |
| 3 | Die Zielklassen: Wurzel, Komponentengrenze, `..`, relativ, Tilde mit `$HOME`, `$PWD` und `$CLAUDE_PROJECT_DIR`, Zweifel (Nr. 3 b) | Fälle 4 bis 9 | Präfixvergleich ohne `/`: Fall 5 Elemente 1 und 2 rot; `..` nicht auflösen: Fall 4 Element 4 und Fall 5 Element 3 rot; Reihe 2 entfernen: Fall 8 rot; `$HOME` nicht einsetzen: Fall 7 Element 5 rot | andere |
| 4 | Der Verzeichnis-Zustand entscheidet relative Ziele (Nr. 3 c) | Fälle 10 und 11 | `cd` ignorieren: Fall 10 rot; jedes `cd` als *außen*: Fall 11 Element 1 rot; relatives `cd` setzt zurück: Fall 10 Element 3 rot; `popd` nicht lesen: Fall 10 Element 8 rot | andere |
| 5 | Zitat und Backslash verbergen einen Operator; ein Zitat davor verbirgt ihn nicht (Nr. 2 e) | Fälle 12 und 13 | Zitat-Zustand entfernen: Fall 12 rot; Zitat nie schließen: Fall 13 rot; Backslash ignorieren: Fall 12 Element 5 rot | andere |
| 6 | Ein Heredoc und ein Here-String sind Eingabe; die Commit-Wege und der Heredoc ins Scratchpad gehen durch; ein Heredoc schreibt nur mit Umleitung (Nr. 5 e, f) | Fälle 14 und 15 | `<<` oder `<<<` als Schreibweg sperren: Fall 15 rot; Heredoc mit Umleitung nicht erkennen: Fall 14 rot | andere |
| 7 | Der Rumpf eines Heredoc: nicht geprüft, Rest der Zeile und Zeilen danach geprüft, nicht terminiert geprüft, mehrere Operatoren, `<<-`, eine Shell liest ihn (Nr. 5 a bis c) | Fälle 16 bis 19 | Rumpf bis zum Ende überspringen: Fall 17 Elemente 2 und 3 rot; Rumpf nie überspringen: Fall 16 rot; Shell-Ausnahme entfernen: Fall 18 rot; Tabulatoren bei `<<-` nicht entfernen: Fall 15 Element 6 und Fall 17 Element 5 rot | andere |
| 8 | `tee` sperrt bei einem Operanden *Repo* an jeder Stelle und hinter `--` (Nr. 6 a) | Fälle 20 und 21 | nur den ersten Operanden prüfen: Fall 20 Element 4 rot; `--` nicht lesen: Fall 20 Element 5 rot; Kopf nur ohne Präfix: Fall 20 Element 6 rot | andere |
| 9 | `sed -i` sperrt in jeder Form, auch im Scratchpad; Namensvettern gehen durch (Nr. 6 b) | Fälle 22 und 23 | Langform entfernen: Fall 22 Elemente 3 bis 5 rot; nur das Wort `-i` lesen: Elemente 2, 6 bis 8 rot; Scratchpad-Ausnahme: Element 10 rot; jedes Wort mit `-i` darin lesen: Fall 23 Elemente 3 und 5 rot | andere |
| 10 | `git apply` sperrt, außer nur lesend oder `-C` nach außen (Nr. 6 c) | Fälle 24 und 25 | Optionen vor dem Unterbefehl nicht überspringen: Fall 24 Elemente 3 bis 5 rot; die lesende Ausnahme entfernen: Fall 25 Elemente 1 bis 4 rot; `--apply` nicht beachten: Fall 24 Element 7 rot; `-C` außen sperren: Fall 25 Element 5 rot | andere |
| 11 | Leere Befehle gehen durch; die Länge sperrt über 32768 (G), genau 32768 wird geprüft (Nr. 2 a, b) | Fälle 26 bis 28 | `-ge` statt `-gt`: Fall 27 rot; Grenze um eins höher: Fall 28 rot; leeren Befehl sperren: Fall 26 rot | andere |
| 12 | Segmente, Zeilenenden, Fortsetzung und Verschachtelung bis Tiefe 3; Tiefe 4 ist fail-closed (Nr. 2 c, d, g, 1 b) | Fälle 29 bis 33 | Fortsetzung nicht verbinden: Fall 29 Element 6 und Fall 22 Element 17 rot; `&` hinter `>` trennen: Fall 1 Element 10 rot; Rekursion ohne Umleitungsprüfung: Fall 31 rot; Tiefe 4 durchlassen: Fall 33 rot | andere |
| 13 | Rang und Meldungsklasse: der erste Verstoß, im Segment Sperrliste vor Kopfprüfung vor Umleitung; jeder Text trägt sein Merkmal und ist gültiges JSON (Nr. 1 d, 9 c) | Fälle 34 bis 38 | Umleitung vor die Kopfprüfung: Fall 36 rot; Sperrliste hinter S: Fall 34 rot; fremdes Merkmal im Text: Fall 37 rot; `"` im Text: Fall 38 rot | andere |
| 14 | Der Bestand gilt unverändert: Sperrliste, Extraktor-Zweifel, `python` (Nr. 1 b, 9 a) | Fälle 39 und 40 | Extraktor-Zweifel durchlassen: Fall 40 rot; Sperrliste leeren: Fall 39 rot | andere |
| 15 | `pkill`, `killall`, `pgrep -f` sperren (P); `pgrep` ohne `-f` und `kill` gehen durch (Nr. 7 a, b) | Fälle 41 bis 43 | `--full` nicht lesen: Fall 42 Element 4 rot; `-F` als `-f` lesen: Fall 43 Element 4 rot; `kill` sperren: Fall 43 Elemente 5 bis 7 rot | andere |
| 16 | Rekursives `rm` sperrt (L) außer unter `…/scratchpad/<Name>`: Muster, Repo, flache Pfade; Zweifel und ohne Operand gehen durch (Nr. 7 c) | Fälle 44 bis 46 | Scratchpad-Ausnahme für Muster: Fall 44 Elemente 19 und 20 rot; Komponente dahinter nicht verlangen: Element 18 rot; Repo nach Scratchpad prüfen: Element 8 rot; Zweifel sperren: Fall 45 Elemente 4 und 5 rot; Muster nach Zweifel: Fall 46 rot | andere |
| 17 | Die Latenz bleibt bei Median 100 ms und 3 s für 32768 Zeichen (Nr. 10 c) | Fall 47 | die Segmentliste je Segment neu aufbauen (quadratisch): Fall 47 Element 2 rot; je Aufruf einen Prozess für jedes Wort: Element 1 rot | andere |
| 18 | Der Prüflauf wird rot, wenn ein Element abweicht; er ist zustandslos, kennt die Symlink-Form der Wurzel und weist ein Argument ab (Nr. 10 a, 1 a) | Fälle 48, 49 und 52 | jede Mutation der Zeilen 1 bis 16 am Guard: Skript Exit 1; Argument ignorieren: Fall 52 rot; Symlink-Form der Wurzel streichen: Fall 49 rot; Ergebnis zwischen Aufrufen halten: Fall 48 rot | andere |
| 19 | Die akzeptierten Negative sind festgehalten (Nr. 11); wer eines behebt, ändert den Vertrag (Nr. 11 a bis h) | Fälle 50 und 51 | `cp` oder `patch` sperren: Fall 50 rot; `<>` als Operator lesen: Fall 50 Element 11 rot; die Quote-Blindheit beheben: Fall 51 rot (Nachzug im Vertrag) | andere |
| 20 | Akzeptiertes Negativ: die Scratchpad-Namen sind Anweisung, der Guard prüft sie nicht (Nr. 8, 11 i). Grund: ein „generischer“ Name ist kein Wortmuster. Folge: eine Kollision bei vergessener Regel fällt am überschriebenen Lauf auf | keiner; der Reviewer liest den Wortlaut in den Übergangsregeln Nr. 8 gegen `SPEC-214` Nr. 8 | keine Mutation möglich | andere |

**Erwartete Produktcode-Größe** (Auftrag 6; Skript, kein Go): *p* = 300 Zeilen für den Guard (heute 133, danach rund 400: Länge und Fortsetzung 10, Zitat-, Heredoc- und Verzeichnis-Zustand 90, Ziel-Klassen und Normalisierung 45, Umleitung 40, `tee`, `sed -i`, `git apply` 55, `pkill`, `pgrep`, `rm` 40, Meldungen und Rang 20). *a* = 0 (der Guard ist einfädig und zustandslos), *b* = 20 Zeilen der Tabelle, *f* = 0 (kein Port-Fake). *S* = 85·0 + 80·20 + 200·0 + 300 = 1900, Marke 2600 nach `SPEC-128` Nr. 1; die Schnittfrage (*S* ≥ 2223) stellt sich nicht, Abstand 323. Zur Einordnung: Gruppe 1 allein (Zeilen 1 bis 14, 17 bis 19, *p* = 210) wäre *S* = 1570, Marke 2200; Gruppen 2 und 3 (Zeilen 15, 16, 20, *p* = 90) *S* = 330. Die Formel ist an Go-Tests geeicht; die Messung dieses Slice (Guard rund 300 Zeilen, Skript rund 400, Doku 20) liegt erwartet bei 700 bis 800 Zeilen und damit weit unter der Marke. Ob die Marke gesenkt wird, entscheidet der Planner nach `SPEC-128` Nr. 1 und mit Grund in §4.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` liegt in `done/` (Closure-Notiz `welle-record-replay-results`); berührt der Architect-Entwurf `tools/gegenproben/` oder `harness/mk/`, liegt kein anderer Slice der Gate-Kette in `in-progress/` (Nr. 7 der Übergangsregeln: ein Besitzer zugleich; `slice-gate-integration-determinism` läuft davor, nie zugleich; Reihenfolge der Kette seit 2026-11-08: `slice-gate-integration-determinism`, `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, `slice-groessenformel-neu`; höchstens drei Slice-Worktrees zugleich). Der Slice hat keinen Code-Bezug zu einem anderen Slice. *(Der Architect berührt weder `tools/gegenproben/` noch `harness/mk/`: die Bedingung entfällt für diesen Slice, `SPEC-214` Nr. 10 d; die Reihenfolge der Kette bleibt eine Setzung des Planners.)*

**Start belegt (2026-11-19, Planner):** `welle-record-replay-results` liegt in `done/`; `in-progress/` enthielt nur `roadmap.md`, kein anderer Slice der Kette lief (`slice-gate-integration-determinism` und `slice-gates-ohne-worktrees` liegen in `done/`); Worktree neben dem Repo (Verzeichnis `odata-recorder-wt-guard`), Zweig `slice-guard-schreibwege`. Der Slice führt die Sendung aus `welle-mock` in §1 *Übernimmt*; ein Nehmer für Gruppen 2 und 3 existiert nicht, ein Slice dafür entsteht erst bei einem Schnitt (§3.13 trägt dann der Planner nach).

**Größenmarke:** Der Architect-Teil des Blocks steht (Pfade, `produktcode`, `fakes`, Schätzung in §3: *S* = 1900, Marke 2600, keine Schnittfrage); die Marke setzt der Planner in eigenem Commit vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14 (`basis`, `schaetzung`, `marke`).

```groessenmarke
pfad: .claude/hooks/pretooluse-command-guard.sh
pfad: tools/harness/guard-faelle.sh
produktcode: 300
fakes: 0
```

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): der Architect entscheidet, dass die Erkennung des Ziels ohne eine ADR nicht möglich ist. *(Tritt nicht ein: `SPEC-214` Nr. 3 und 4 entscheiden sie ohne ADR.)*

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag (Sperre mit `liegt in` und Anker, oder benannte Lücke der ungefangenen Schreibwege).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Der Guard sperrt eine Commit-Message oder einen Schreibweg ins Scratchpad und bricht die Arbeit der Rollen — **Ausgang:** weiter offen: der Durchlass ist entschieden (`SPEC-214` Nr. 5 f) und führt je Form einen Fall (Fälle 3, 14 bis 16); der Ausgang steht erst, wenn der Implementer die Mutationen rot gesehen hat.
- Die Erkennung des Ziels einer Umleitung ist wegen der Quote-Blindheit unvollständig — **Ausgang:** weiter offen: die ungefangenen und die falsch gemeldeten Formen stehen als akzeptiertes Negativ mit Grund und Folge in `SPEC-214` Nr. 11 a bis h (Fälle 50 und 51 halten sie fest); die Falschmeldungen sperren eine Rolle ohne Anlass, die Abhilfe `git commit -F <Datei>` steht in der Meldung S.
- Raster der Randformen (Auftrag 5) ist noch leer — **Ausgang:** entfallen: das Raster steht unten, jede Zelle trägt eine Nummer von `SPEC-214` oder „gilt nicht“ mit Grund; die Zeilen für `git apply`, `pgrep`/`kill`/`rm -rf` und Scratchpad-Namen sind darin.
- Eine Sperre von `git apply` nimmt den Architects ohne `Edit` ihre einzige Schreibform — **Ausgang:** weiter offen; Adresse: User (Rückgabe 2), der Architect schreibt bis dahin ganze Dateien mit `Write` (`SPEC-214` Nr. 12).
- Der Umfang mit Gruppen 2 und 3 überschreitet die Schnittgrenze — **Ausgang:** weiter offen bis zur Marke des Planners: *S* = 1900 liegt unter der Schnittgrenze 2223 (§3, Abschnitt *Erwartete Produktcode-Größe*).
- Der Prüflauf ist kein Gate; eine Regression am Guard fängt nur ein Lauf des Skripts — **Ausgang:** weiter offen; Grund, Folge und Auslöser für ein Gate: `SPEC-214` Nr. 10 d; Reviewer und Verifier fahren das Skript selbst.
- Die Latenz des Guards wächst mit der Zerlegung (der Guard läuft vor jedem Bash-Aufruf) — **Ausgang:** weiter offen; Messung vor der Änderung 12 ms je Lauf (`SPEC-214` Nr. 10 c), Grenzen Median 100 ms und 3 s für 32768 Zeichen; der Implementer misst und berichtet in §7.

**Rückgaben und offene Entscheidungen (`AGENTS.md` §3.13, Ausgang *offen*):**

| Nr. | Gegenstand | Adresse | Ausgang |
|---|---|---|---|
| 1 | Die Marke nach `SPEC-128` Nr. 1 aus Tabelle (*b* = 20) und Block (*p* = 300); der Planner entscheidet, ob er sie wegen der erwarteten Messung (700 bis 800 Zeilen) niedriger ansetzt | Planner | offen |
| 2 | Werkzeugliste des Typs `architect`: Diff-Vorschlag für `.claude/agents/architect.md`, Zeile `tools: Read, Write, Bash` wird `tools: Read, Write, Edit, Bash` (gleiche Zeilenzahl, die Datei steht auf ihrer Größengrenze); ohne die Änderung schreibt der Architect Dateien ganz mit `Write` | User (Closure legt den Diff bei) | offen |
| 3 | Rollen-Anweisung zu den Scratchpad-Namen (`SPEC-214` Nr. 8) in `.claude/commands/` und `.claude/agents/`: Diff-Vorschlag in der Closure; bis dahin trägt sie nur die Übergangsregel Nr. 8 und der Plan jedes Slice | User (Closure legt den Diff bei) | offen |
| 4 | Wortlaut der Übergangsregeln Nr. 8 (Wächter statt Disziplin, Namensregel aus `SPEC-214` Nr. 8) | Implementer (§3, Zeile `docs/plan/planning/README.md`) | offen |

**Ausgänge der Aufträge an den Architect (Stand 2026-11-19):** alle acht *entschieden*, der Vertrag liegt als [SPEC-214](../../../../harness/vertraege/SPEC-214-command-guard.md) vor dem ersten Code-Commit (`AGENTS.md` §3.12): (1) Zielerkennung, Parse-Zweifel — *entschieden*, Nr. 2 e, 3, 4, 9 · (2) Durchlass — *entschieden*, Nr. 5 f, 4 b · (3) Menge der Schreibwege und Negative — *entschieden*, Nr. 4, 6, 11 · (4) Prüflauf — *entschieden*, Skript ohne Gate, Nr. 10 · (5) Randformen — *entschieden*, Raster unten · (6) Tabelle, Raster, Größe — *entschieden*, §3 und §6 · (7) `git apply`, Prozess- und Löschwege, Architects ohne `Edit` — *entschieden*, Nr. 6 c, 7, 11 b, 12 · (8) Scratchpad-Namen — *entschieden*, Nr. 8, 11 i.

**Raster der Randformen** (`AGENTS.md` §3.16): Operation × Argumentart; jede Zelle nennt die Nummer von `SPEC-214` oder „gilt nicht“ mit Grund. Ein Abbruch trifft einen zustandslosen Hook: stirbt der Prozess, urteilt der Guard nicht, und das Werkzeug des Hosts behandelt den Aufruf als Hook-Fehler; das gilt für jede Operation und steht nur einmal in den Zellen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Hook-Eingabe lesen | kein `command`: Nr. 2 a; kein Objekt: Nr. 9 a | leerer Befehl, nur Leerraum und Zeilenenden: Nr. 2 a | kaputtes JSON, Escape-Folge im Befehl: Nr. 9 a; über 32768 Zeichen: Nr. 2 b | gilt nicht: der Hook hält keinen Zustand, ein gestorbener Prozess hat kein Urteil gefällt | derselbe Befehl zweimal: Nr. 10 e Fall 48 |
| Befehl zerlegen | kein Trenner: ein Segment (Nr. 1 b) | leeres Segment: übersprungen (Nr. 1 b) | CR, Zeilenumbruch, Fortsetzung: Nr. 2 c, d; Tiefe über 3: Nr. 1 b, 9 a | gilt nicht: wie oben | gilt nicht: reine Funktion des Befehls, Fall 48 |
| Ziel klassifizieren | kein Ziel: Nr. 4 c | leeres Wort: Nr. 3 b Reihe 1 | Variable, Tilde, Muster, absolut außen, fremder Arbeitsbaum: Nr. 3 b, 11 g | gilt nicht: wie oben | mehrere `cd` im Befehl: Nr. 3 c |
| Umleitung prüfen | Operator ohne Ziel: Nr. 4 c | `> ""`: Nr. 4 c | Dup, `<>`, `{fd}>`: Nr. 4 b, e, 11 e | gilt nicht: wie oben | mehrere Umleitungen in einem Segment: Nr. 4 a |
| Heredoc, Here-String lesen | Endezeile fehlt: Nr. 5 c | Rumpf aus null Zeilen: Nr. 5 a | Zitatformen des Wortes, `<<-`, Shell als Leser, Here-String: Nr. 5 a, c, d | gilt nicht: wie oben | mehrere Operatoren in einer Zeile: Nr. 5 a |
| `tee` prüfen | kein Operand: Nr. 6 a | Operand `""`: Nr. 3 b Reihe 1 | Operand außen, Zweifel, nach `--`: Nr. 6 a, 3 b | gilt nicht: wie oben | mehrere Operanden: Nr. 6 a |
| `sed -i` prüfen | kein `-i`: Nr. 6 b | `sed` ohne Argumente: Nr. 6 b (kein Flag) | Bündel, Langform, Abkürzung, Operand im Scratchpad: Nr. 6 b | gilt nicht: wie oben | mehrere Flag-Wörter: Nr. 6 b („ein Wort genügt“) |
| `git apply` prüfen | `git` ohne Unterbefehl: Nr. 6 c | `git apply` ohne Patch: Nr. 6 c (gesperrt) | Optionen vor dem Unterbefehl, nur lesend, `-C` nach außen: Nr. 6 c | gilt nicht: wie oben | `git apply` in zwei Segmenten: Nr. 9 c (der erste Verstoß) |
| `pkill`, `killall`, `pgrep` prüfen | ohne Muster: Nr. 7 a (der Kopf genügt) | `pgrep` ohne `-f`: Nr. 7 a | Kurzbündel, `--full`, `-F`, `kill` mit Variable: Nr. 7 a, b | gilt nicht: wie oben | `kill $(pgrep -f x)`, `pgrep -f x` mit `xargs`: Nr. 7 b |
| `rm` rekursiv prüfen | ohne Operand: Nr. 7 c | Operand `""`: Nr. 3 b Reihe 1 (Zweifel, durch) | Muster, Repo, flach, `~`, `..`, Scratchpad, Variable: Nr. 7 c | gilt nicht: wie oben | mehrere Operanden: Nr. 7 c („ein Operand genügt“) |
| Meldung ausgeben | gilt nicht: die Klassen sind fest, jede trägt ihren Text (Nr. 1 d) | gilt nicht: wie links | Sonderzeichen im Text: Nr. 1 d, Fall 38 | gilt nicht: wie oben | Fall 48 |
| Scratchpad-Namen (Anweisung) | kein Verzeichnis `<kurz>-<rolle>`: Nr. 8 (nicht geprüft, Nr. 11 i) | gilt nicht: eine Anweisung, kein Eingabewert | generischer Name: Nr. 8 | gilt nicht: eine Anweisung | zwei Läufe derselben Rolle: Nr. 8 (`-2`) |
| Prüflauf aufrufen | kein Argument: Lauf, Nr. 10 a | Argument `""`: Exit 2, Nr. 10 a | ein Argument: Exit 2, Nr. 10 a, Fall 52 | gilt nicht: das Wegwerfverzeichnis liegt außerhalb des Arbeitsbaums, ein Rest ist ein leeres Verzeichnis unter `/tmp` | zweiter Lauf: Nr. 10 a (zustandslos) |

**Zusatztabelle** (mehrere Verstöße am selben Schlüssel, Modi; `AGENTS.md` §3.16):

| Fall | Entscheidung | Fänger |
|---|---|---|
| Sperrliste und S im selben Befehl | der erste in der Reihenfolge der Segmente bestimmt die Meldung (Nr. 9 c) | Fälle 34 und 35 |
| Sperrliste, P oder L und Umleitung im selben Segment | Sperrliste vor Kopfprüfung vor Umleitung (Nr. 9 c) | Fälle 34 (Element 2) und 36 |
| `sed -i` und Umleitung im selben Segment | S (beide gleich, Nr. 6 b, 4 a) | Fall 35 Element 2 |
| Zwei Umleitungen, die erste *außen*, die zweite *Repo* | S (Nr. 4 a) | Fall 2 Element 7 |
| `tee` mit zwei Operanden, einer *außen* | S (Nr. 6 a) | Fall 20 Element 4 |
| Umleitung nach `cd` nach außen und absolutes Ziel unter der Wurzel | S: der Zustand gilt allein für relative Ziele (Nr. 3 c) | Fall 11 Element 4 |
| `rm -rf` mit einem Operanden im Scratchpad und einem im Repo | L (Nr. 7 c) | Fall 44 Element 22 |
| Heredoc an eine Shell und Rumpf mit Sperrlisten-Wort | H (Nr. 5 c, 9 c) | Fall 19 |
| Modus: `HOME` innerhalb der Wurzel, `~`-Ziel | S (Nr. 3 b Reihe 4) | Fall 7 Element 5, Fall 11 Element 6 |
| Modus: Guard über Symlink aufgerufen | beide Schreibweisen der Wurzel (Nr. 1 a) | Fall 49 |
| Modus: Befehl in `-c`-String, Tiefe 1 bis 4 | geprüft bis 3, Tiefe 4 fail-closed (Nr. 1 b) | Fälle 31 bis 33 |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `welle-record-replay`; noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (4×, Ausgang `geplant` in diesem Slice); `BEO-REPO/report-der-rolle-faellt-docs-check` (2×) und `BEO-REPO/regel-in-zwei-adaptern-ohne-sensor` (1×) berühren den Guard nicht.
