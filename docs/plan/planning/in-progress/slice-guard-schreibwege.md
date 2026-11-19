# Slice slice-guard-schreibwege: Command-Guard sperrt die Schreibwege auf Repo-Dateien

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice und keine Closure-Bedingung, die von der DoD verschieden wäre (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln), sobald eine Gegenprobe `tools/gegenproben/` berührt (Nr. 7: ein Besitzer der Gate-Kette zugleich).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Werkzeuge der Rollen, der Bezug ist der nächste, den es gibt), `AGENTS.md` §3.1 (Docker-only, der Guard ist sein Träger) und §3.9 (die Betriebsregel gegen `sed -i` und Heredoc steht in den Übergangsregeln Nr. 8), Register-Eintrag `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (17 Belege am 2026-11-08, Ausgang `geplant` in diesem Slice; Belege u. a. `slice-baseline-import`, `slice-unsichtbare-zeichen`, `slice-v2-metadaten`, `slice-workspace-use-case`, `slice-edm-ausdruecke`, `slice-bootstrap-m3-e2e-mock`).

**Übernimmt:** die Sendung aus der Closure von `welle-mock` (Geber `welle-mock`, Punkt „Betriebsregel nur `Edit` und `Write`“ in der Closure-Notiz, Entscheidung des Users vom 2026-11-08, Punkte lösen): die Formen `git apply` (Patches der Architects ohne `Edit`), `pgrep -f`, `kill` und `rm -rf` mit Muster oder auf Verzeichnisse, die der Lauf nicht angelegt hat, und die generischen Namen im gemeinsamen Scratchpad (`mut/`, `gt.sh`, `fresh.sh`), die der Plan bis zum 2026-11-08 nicht nannte (Evidenz `slice-bootstrap-m3-e2e-mock`, `slice-bootstrap-m3`, `slice-edm-ausdruecke`, `slice-snapshot-use-case`).

**Berührte Spec-Stellen:** ein Abschnitt oder Vertrag zum Guard unter `harness/vertraege/` (Kennung aus dem Bereich der Linie C, `SPEC-200` bis `SPEC-229`; die Nummer vergibt der Architect) · `harness/vertraege/README.md` §Verträge (Zeile, Architect) · `SPEC-102` (nur falls der Architect eine Gegenprobe wählt).
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

**Erweiterung des Umfangs (Planner, 2026-11-08, Entscheidung des Users „Punkte lösen“):** Der Schnitt gilt dem Guard als Stolperdraht in drei Gruppen: (1) Schreibwege auf Repo-Dateien (`sed -i`, Heredoc, `>`, `>>`, `tee`, **`git apply`**), (2) Prozess- und Löschwege auf fremde Gegenstände (**`pgrep -f` mit Muster, `kill`/`pkill` und `rm -rf` mit Muster oder auf ein Verzeichnis außerhalb des eigenen Laufs**), (3) **generische Namen im gemeinsamen Scratchpad** (`mut/`, `gt.sh`, `fresh.sh`; der Name trägt die Kennung des Laufs oder der Rolle). Ob (2) und (3) der Guard sperrt, in der Rollenanweisung stehen oder als akzeptiertes Negativ bleiben, entscheidet der Architect (Aufträge 7 und 8); ist der Schnitt der Marke damit über der Schnittgrenze, schneidet der Planner vor dem Code in `slice-guard-schreibwege` (Gruppe 1) und einen neuen Slice (Gruppen 2 und 3), siehe §4.

**Warnung an den Architect (Pflicht vor dem Entwurf):** `git commit -F -` mit Heredoc oder Pipe auf der Standardeingabe sowie ein Heredoc, das eine Commit-Message nach `git commit -F` liefert, **bleiben erlaubt**. Der Guard ist quote-blind und trennt Segmente an `;`, `&`, `|`, `(` (Kopf des Skripts); eine Regel, die jedes Heredoc oder jedes `>` sperrt, bricht die Commit-Wege, die `AGENTS.md` und `implement-slice` vorschreiben, und Schreibwege in das Scratchpad (`/tmp/claude-…`) und in `/dev/null`, die erlaubt sind. Der Architect entscheidet, **welche Ziele** als Repo-Datei gelten (Pfad unter dem Wurzelverzeichnis des Repos) und wie der Guard sie erkennt, ohne das Ziel einer Umleitung falsch zu lesen.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Sandbox oder eine vollständige Erkennung aller Schreibwege (`perl -i`, `python` über eine andere Tür, `cp`, `mv`, `dd`, `install`) — der Guard ist ein Stolperdraht, keine Sandbox (Kopf des Skripts); der Architect benennt die ungefangenen Wege als akzeptiertes Negativ mit Grund und Folge. `git apply`, `pgrep -f`/`kill`/`rm -rf` und die Scratchpad-Namen gehören seit dem 2026-11-08 nicht mehr zu diesem Ausschluss, sondern zum Umfang (Aufträge 7 und 8).
- Eine Änderung der Werkzeugliste der Rollen-Typen (`.claude/agents/architect.md` hat kein `Edit`; die Datei steht auf ihrer Größengrenze) — Entscheidung des Users nach einem Diff-Vorschlag des Architect; ein Wächter, der `git apply` sperrt, ohne dass der Architect ein `Edit` hat, sperrt seine einzige Schreibform (Auftrag 7).
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

- [ ] Der Guard sperrt `sed -i` (auch `-i.bak`, `--in-place`, in Flag-Bündeln), `git apply` und das Schreiben per Heredoc, `>`, `>>` und `tee` auf einen Pfad unter dem Wurzelverzeichnis des Repos, sperrt oder meldet `pgrep -f`/`kill`/`pkill` und `rm -rf` mit Muster oder auf ein fremdes Verzeichnis nach der Entscheidung des Architect (Auftrag 7), führt die Regel für Scratchpad-Namen mit Lauf-Kennung (Auftrag 8), und lässt `git commit -F -` samt Heredoc für die Message, Umleitungen in `/dev/null` und in das Scratchpad durch; jede Zusage des Vertrags hat eine Zeile in der Tabelle *Zusage · Fänger* (§3) mit einer vom Implementer rot gesehenen Mutation.
- [ ] Ein Prüflauf der Sperre (Gegenprobe oder Test des Skripts, Wahl des Architect) belegt Sperre und Durchlass über die Menge der Schreibwege und die Menge der Ziele; die Zusatzzeit ist gemessen und in §7 berichtet.
- [ ] Doku-Update: der Vertrag zum Guard, `harness/README.md` (falls ein Ziel entsteht), Übergangsregeln Nr. 8 nennen den Wächter statt der Disziplin.
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
| `.claude/hooks/pretooluse-command-guard.sh` | update | Erkennung von `sed -i`, Heredoc, `>`, `>>`, `tee` auf Repo-Pfade; Meldung mit Verweis auf `Edit` und `Write`; Durchlass für Commit-Message, `/dev/null`, Scratchpad |
| `harness/vertraege/` (neuer Vertrag oder Abschnitt) | neu (Architect) | Zusagen, Randformen, *Ungeprüft* |
| `tools/gegenproben/`, `harness/mk/gegenproben.mk`, `tools/harness/gegenproben-eingabe.sh` | update (falls der Architect eine Gegenprobe wählt) | Fälle der Sperre und des Durchlasses; berührt die Gate-Kette, damit Linie C (Nr. 7) |
| `docs/plan/planning/README.md` | update | Übergangsregeln Nr. 8 nennen den Wächter |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Vertrag, Entscheidungen, Tabelle *Zusage · Fänger*, Raster, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Aufträge an den Architect:**

1. **Zielerkennung.** Wann ein Ziel als Repo-Datei gilt (absoluter Pfad unter dem Wurzelverzeichnis, relativer Pfad, `~`, Variablen, Umleitung mit Dateideskriptor, `&>`); wie der Guard das Ziel einer Umleitung findet, obwohl er quote-blind ist; das Verhalten bei Parse-Zweifel (fail-closed wie heute für den Extraktor oder durchlassen).
2. **Durchlass.** `git commit -F -` mit Heredoc oder Pipe, Heredoc für eine Datei im Scratchpad, `/dev/null`, `2>&1`; die Menge der erlaubten Ziele als Liste mit Grund. Die Erlaubnis für Commit-Messages ist Pflicht (Warnung in §1).
3. **Menge der Schreibwege.** `sed -i`, `sed --in-place`, Flag-Bündel (`-ni`, `-i.bak`), `tee`, `>`, `>>`, `<<`, `<<<`; welche übrigen (`cp`, `mv`, `dd`, `install`, `perl -i`) als akzeptiertes Negativ mit Grund und Folge stehen (`AGENTS.md` §3.15).
4. **Prüflauf.** Gegenprobe (berührt die Gate-Kette, Nr. 7) oder ein Test des Skripts ohne Gate; wenn Gegenprobe, die Fälle und das Zeitbudget (`SPEC-102`).
5. **Randformen** vor dem ersten Code-Commit im Vertrag (Raster in §6): leerer Befehl, nur Whitespace, sehr langer Befehl, Verschachtelung (`bash -c "sed -i …"`, Tiefe wie heute), mehrere Segmente, Zeilenumbruch im Befehl.
6. **Tabelle *Zusage · Fänger*** (§3), **Raster** (§6), **erwartete Produktcode-Größe** für die Marke (Skript, kein Go; ob die Messung zählt, entscheidet `SPEC-128`).
7. **`git apply` und die Prozess- und Löschwege** (Sendung aus `welle-mock`). `git apply` mit Patch aus Datei oder Standardeingabe als Schreibweg; die Architects ohne `Edit` schrieben so (Evidenz `slice-edm-ausdruecke`, `slice-snapshot-use-case`), eine Sperre ohne `Edit` im Rollen-Typ nimmt ihnen den Weg (siehe Ausschluss; der Architect legt dem User den Diff vor oder lässt `git apply` auf Dateien außerhalb von `spec/`, `docs/plan/adr/` und `harness/vertraege/` bestehen). `pgrep -f` mit Muster, `kill`, `pkill`, `rm -rf` mit Muster oder Verzeichnis außerhalb des Scratchpad-Unterverzeichnisses des eigenen Laufs: sperren, melden oder Negativ mit Grund und Folge; die Erkennung des Ziels ist quote-blind (Auftrag 1).
8. **Scratchpad-Namen.** Eine Regel, dass jede Datei oder jedes Verzeichnis, das ein Lauf im gemeinsamen Scratchpad anlegt, die Kennung des Laufs oder der Rolle im Namen trägt (`mut/`, `gt.sh`, `fresh.sh` sind die Fälle), und ob der Guard sie erkennt (Befehle mit festem Namen im Scratchpad ohne Kennung) oder die Anweisung der Rollen sie trägt; das Gegenstück (Prozess-Muster beschränkt auf eigene Namen) gehört zu Auftrag 7. Ob der Umfang (Gruppen 2 und 3) mit Gruppe 1 in einem Slice bleibt, sagt die Marke (§4).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect): folgt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` liegt in `done/` (Closure-Notiz `welle-record-replay-results`); berührt der Architect-Entwurf `tools/gegenproben/` oder `harness/mk/`, liegt kein anderer Slice der Gate-Kette in `in-progress/` (Nr. 7 der Übergangsregeln: ein Besitzer zugleich; `slice-gate-integration-determinism` läuft davor, nie zugleich; Reihenfolge der Kette seit 2026-11-08: `slice-gate-integration-determinism`, `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, `slice-groessenformel-neu`; höchstens drei Slice-Worktrees zugleich). Der Slice hat keinen Code-Bezug zu einem anderen Slice.

**Größenmarke:** Prosa, kein Block. Die Marke setzt der Planner vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14; die Schätzung kommt aus der Tabelle des Architect. Erwartet war ein kleiner Slice (ein Skript, eine Tabelle, ein Prüflauf); mit den Gruppen 2 und 3 (Umfang vom 2026-11-08) rechnet der Planner nach der Tabelle des Architect neu und schneidet vor dem Code, wenn *S* die Schnittgrenze von `SPEC-128` Nr. 1 erreicht (Gruppe 1 bleibt dieser Slice, die Gruppen 2 und 3 werden ein neuer Slice nach ihm; der Nehmer trägt die Sendung dann vor dem Geber ein, `AGENTS.md` §3.13).

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): der Architect entscheidet, dass die Erkennung des Ziels ohne eine ADR nicht möglich ist.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag (Sperre mit `liegt in` und Anker, oder benannte Lücke der ungefangenen Schreibwege).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Der Guard sperrt eine Commit-Message oder einen Schreibweg ins Scratchpad und bricht die Arbeit der Rollen — **Ausgang:** weiter offen: der Architect fixiert den Durchlass (Auftrag 2) und die Tabelle *Zusage · Fänger* führt einen Fall je Durchlass.
- Die Erkennung des Ziels einer Umleitung ist wegen der Quote-Blindheit unvollständig — **Ausgang:** weiter offen: der Architect nennt die ungefangenen Formen als akzeptiertes Negativ mit Grund und Folge (Auftrag 3).
- Raster der Randformen (Auftrag 5) ist noch leer — **Ausgang:** weiter offen: der Architect füllt es vor dem ersten Code-Commit (`AGENTS.md` §3.16); die Zeilen für `git apply`, `pgrep`/`kill`/`rm -rf` und Scratchpad-Namen (Aufträge 7 und 8) gehören dazu.
- Eine Sperre von `git apply` nimmt den Architects ohne `Edit` ihre einzige Schreibform — **Ausgang:** weiter offen; Adresse: Architect (Auftrag 7), Entscheidung über die Werkzeugliste des Rollen-Typs beim User.
- Der Umfang mit Gruppen 2 und 3 überschreitet die Schnittgrenze — **Ausgang:** weiter offen; Adresse: Planner (Marke nach der Tabelle, Schnitt vor dem Code, §4).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `welle-record-replay`; noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (4×, Ausgang `geplant` in diesem Slice); `BEO-REPO/report-der-rolle-faellt-docs-check` (2×) und `BEO-REPO/regel-in-zwei-adaptern-ohne-sensor` (1×) berühren den Guard nicht.
