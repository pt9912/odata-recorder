# SPEC-212 — Arbeitsbäume unter `.wt/` außerhalb der Prüfmenge der Gates (`make a-check`, `make docs-check`)

**Werkzeug:** die Konfigurationen `.a-check.yml` (`make a-check`, `SPEC-039`) und `.d-check.yml` (`make docs-check`, `SPEC-040`); die übrigen Gates der Kette sind gemessen oder begründet unberührt (Nr. 1).

**Festlegung:** Der Pfad `.wt/` an der Wurzel des Hauptbaums trägt die Arbeitsbäume (`git worktree`), je Arbeitsbaum eine volle Kopie des Repos.
Er gehört nicht zum geprüften Stand: Er ist nie getrackt, und jeder Arbeitsbaum wird an seinem eigenen Ort in seinem Zweig von denselben Gates geprüft.
`make a-check` und `make docs-check` nehmen ihn aus; Bindung ohne ADR (Nr. 2).
Alles Übrige bleibt in der Prüfmenge: Schwellen, Regeln und Module sind unverändert, ausgenommen wird ein Pfad.

1. **Bestand: welche Gates `.wt/` lesen.** Gemessen am 2026-11-18 im Hauptbaum (`make <Ziel>` an `main`) mit einem Arbeitsbaum `.wt/gates-ohne-wt/` (1930 Dateien) und der Last des Hosts im Lauf (Load 150 bis 170 am Anfang, 25 am Ende); die Läufe der Gates unter *nein* mit Mechanismus statt Lauf sind wegen der Last nicht gefahren, ihr Mechanismus ist gemessen oder gelesen, wie je Zeile genannt.

   | Gate | liest `.wt/` | Messung oder Grund |
   |---|---|---|
   | `make a-check` | **ja** | Der Hauptbaum ist als Wurzel `/src` schreibgeschützt gemountet; 1351 Befunde `wrong-direction`, alle unter `.wt/gates-ohne-wt/`, dazu der Hinweis „962 Dateien liegen in keiner Schicht“; Exit 1, 2,3 s. |
   | `make docs-check` | **ja** | `scan.roots` ist `.`, `scan.ignore` ist an der Wurzel verankert (`.harness/**` trifft `.wt/…/.harness/` nicht); 39 Befunde (34 `codepath-missing`, 5 `id-unlinked`), alle unter `.wt/gates-ohne-wt/`, 19 davon in der Kopie der vendorten Baseline, 1807 Dateien, Exit 1, 4,7 s. Die Module `file`, `structure`, `targets` und `planning` lesen `.wt/` nicht: ihre Pfade und Globs sind an der Wurzel verankert (Messprobe: `.wt/w/AGENTS.md` um 63 Zeilen und einen Abschnitt `### 3.99` verlängert, kein Befund von `file` oder `structure`). |
   | `make zeichen-check` | nein | Die Dateien wählt `git ls-files --cached --others --exclude-standard`; ein Arbeitsbaum trägt eine Datei `.git` und erscheint als **ein** Eintrag `.wt/gates-ohne-wt/` ohne Inhalt (`git ls-files --cached --others` ohne den Ausschluss des Hauptbaums: 1938 Einträge, davon `.wt/gates-ohne-wt/` und `.wt/heavy.lock`; mit ihm 1934). Lauf: `zeichen: ok, 1869 Datei(en) geprüft`, Exit 0, 4,9 s. |
   | `make lint`, `make lint-profil` | nein, Mechanismus | Der Build-Kontext ist die Allowlist `.dockerignore` (`*`, danach benannte Pfade); `.wt/` steht nicht darin, und `lint.sh` läuft im Container auf `src`, `tests` und `tools/*`. Gemessen mit einem Kontext derselben Form (`*`, `!src/`, `!tools/zeichen`) und einer Datei `.wt/w/src/a.go`: der Kontext trägt nur `src/b.go` und `tools/zeichen/z.go`. `lint-profil` kopiert allein `.golangci.yml`. |
   | `make build`, `make test`, `make test-race`, `make handbuch-check`, `make katalog-check`, `make groessenmarke` (Stufe) | nein, Mechanismus | derselbe Kontext; die Stufen kopieren benannte Pfade (`COPY src`, `COPY tests`, `COPY tools/…`, `COPY docs/user`). |
   | `make groessenmarke` (Host) | nein | `git diff --numstat` über die Pfade des Plans und `docs/plan/planning/in-progress`; `.wt/` ist kein Pfad davon. |
   | `make schema-check`, `make migration-check` | nein, Mechanismus | `docker cp` benannter Dateien (`tools/schema/…`, Migrationsverzeichnis). |
   | `make baseline-verify` | nein | `find` im Verzeichnis der vendorten Baseline. |
   | `make test-integration`, `make test-determinism`, `make test-contract` | nein, Mechanismus | Stufen aus demselben Kontext, kein Host-Scan. |
   | `make gegenproben` | nein, Mechanismus | Die Prüfsumme (`gegenproben-eingabe.sh`) liest `git ls-files` über benannte Pfade (`src`, `tools/…`, `.a-check.yml` und so weiter); die Stufe kopiert benannte Pfade; die Fälle laufen in Wegwerfverzeichnissen. |
   | Gate-Nachweis (`record-gates`, `working-tree-hash.sh`) | nein | `git ls-files --exclude-standard`; ein Arbeitsbaum ist ein Eintrag ohne Datei (Zeile `GONE`), kein Inhalt. |
   | `tools/harness/mutationslauf.sh` (kein Gate) | ja | `tar -C <Wurzel> --exclude=./.git … .` kopiert `.wt/` mit in den Container; Folge nur Größe der Kopie, kein Befund (Go überspringt Verzeichnisse mit Punkt-Präfix); Nr. 6 c. |
2. **Bindung ohne ADR.** *Entscheidung:* Die Ausnahme ist kein Fall von `AGENTS.md` §3.6. *Alternativen:* (a) eine ADR, die die Reichweite von `make a-check` und `make docs-check` neu fasst; (b) die Arbeitsbäume außerhalb des Repos (anderer Pfad); (c) die Ausnahme je Gate im Vertrag ohne ADR.
   *Gewählt:* (c). *Gründe:* §3.6 bindet Senkungen von Schwellen, Linter-Strenge und Architekturregeln; hier sinkt keine davon, `.a-check.yml` behält Schichten und Kanten, `.d-check.yml` Module, Klassen, Regeln und Grenzen, und die Prüfmenge eines Standes bleibt gleich: `git ls-files .wt` ist leer (gemessen, 0 Einträge), also schließt die Ausnahme keine Datei eines Commits aus. Die Arbeitsbäume sind Kopien eines Standes, den das Gate an seinem eigenen Ort liest; sie im Hauptbaum zu lesen meldet denselben Befund ein zweites Mal an einem Pfad, den keine Regel meint (die Muster `.tmp/**` und `.harness/**` in `scan.ignore` sind dieselbe Art Ausnahme). Gemeldet wurden 1351 und 39 Befunde, die kein Stand des Hauptbaums trägt. *Verworfen:* (a) schafft ein Dokument für eine Zeile Konfiguration je Gate und ändert keine Regel; (b) ist keine Entscheidung dieses Vertrags (der Pfad der Arbeitsbäume ist nicht sein Gegenstand) und entlastet nur die Gates, die ihn nicht lesen. *Bedingung der Entscheidung:* die Ausnahme trifft nur Pfade, die nie Teil eines Commits sind (Nr. 4, Zeile 11). Hält sie jemand für eine Senkung, gilt der Weg des Konflikts: eine Folge-Entscheidung, die diese Nr. ablöst; die Konfiguration bleibt bis dahin.
3. **Die Ausnahme je Gate.** Genau ein Muster, an der Wurzel verankert, nicht `**/.wt/**`.
   - **a)** `.a-check.yml`: der Schlüssel `exclude: [".wt/**"]` auf oberster Ebene. Gemessen am Pin (`ghcr.io/pt9912/a-check@sha256:9738…`): `.wt/**` nimmt `.wt/` und jede tiefere Verschachtelung aus (`.wt/a/.wt/b/…`); `.wt/` und `.wt` nehmen nichts aus; `.wt/*` nimmt nichts aus (ein Segment ist keine Datei); `**/.wt/**` nimmt zusätzlich ein Verzeichnis `.wt` im Quellbaum aus (Nr. 4, Zeile 3).
   - **b)** `.d-check.yml`: der Eintrag `".wt/**"` in `scan.ignore`. Gemessen am Pin (`ghcr.io/pt9912/d-check@sha256:d8ce…`, Modul `links`): `.wt/**` nimmt `.wt/` samt `.wt/a/.wt/b/` aus und lässt `x/.wt/b.md` (ein Verzeichnis `.wt` unter einem Quellverzeichnis `x`) geprüft; `.wt/` nimmt nichts aus; `.wt/*` verhält sich wie `.wt/**`; `**/.wt/**` nimmt `x/.wt/b.md` mit aus. Das Muster gilt für die Module, die `scan.ignore` lesen (`links`, `anchors`, `ids`, `codepaths`, `matrix`); die Module `file`, `structure`, `targets` und `planning` brauchen es nicht (Nr. 1).
   - **c)** Die übrigen Gates ändern sich nicht (Nr. 1). Eine Konfiguration oder ein Skript, das ein weiteres Gate `.wt/` lesen ließe, trägt dieselbe Ausnahme und diese Nr. ihre Zeile.
4. **Randformen.** Raster Konfiguration × Form des Pfades; jede Zelle nennt Verhalten und Ort. *Gemessen* heißt am 2026-11-18 an den Pins aus Nr. 3, an einem Wegwerfverzeichnis mit den Dateien der Form, `.a-check.yml` aus dem Repo mit der Zeile von Nr. 3 a bzw. einer `.d-check.yml` mit `scan`-Block, `ignore: [".wt/**"]` und `modules: [links]`.

   | Nr. | Form | `make a-check` | `make docs-check` | Ort |
   |---|---|---|---|---|
   | 1 | Arbeitsbaum mit Verletzung unter `.wt/<Name>/…` | ungemeldet, Exit 0 (gemessen) | ungemeldet, Exit 0 (gemessen) | Nr. 3; Fänger Nr. 5 Zeilen 1 und 5 |
   | 2 | Arbeitsbaum tiefer verschachtelt, `.wt/a/.wt/b/…` | ungemeldet (gemessen) | ungemeldet (gemessen) | Nr. 3; Fänger Nr. 5 Zeilen 4 und 8 |
   | 3 | Verzeichnis `.wt` im Quellbaum (`src/…/.wt/`, `x/.wt/`) | gemeldet (`core-impurity`, gemessen) | gemeldet (`target-missing`, gemessen) | Nr. 3: das Muster ist an der Wurzel verankert; Fänger Nr. 5 Zeilen 3 und 7 |
   | 4 | `.wt` ist ein Symlink | nicht gefolgt, kein Befund (gemessen) | nicht gefolgt (gemessen: die Datei hinter dem Link zählt nicht) | Verhalten der Pins; kein Fall, Nr. 6 d |
   | 5 | ein Symlink im Baum zeigt in `.wt/` | nicht gefolgt (gemessen) | nicht gefolgt (gemessen) | wie Zeile 4 |
   | 6 | `.wt/` fehlt (kein Arbeitsbaum) | Exit 0, kein Konfigurationsfehler (gemessen) | kein Konfigurationsfehler, übrige Befunde unverändert (gemessen) | ein Muster ohne Treffer ist kein Fehler beider Pins; kein Fall, Nr. 6 d |
   | 7 | Arbeitsbaum ohne Verletzung | kein Befund | kein Befund | trivial, kein Fall |
   | 8 | dieselbe Verletzung im Hauptbaum neben `.wt/` | gemeldet (gemessen) | gemeldet (gemessen) | Nr. 3; Fänger Nr. 5 Zeilen 2 und 6 |
   | 9 | Arbeitsbaum außerhalb des Repos (anderer Pfad) | nicht gelesen: nicht im Mount | nicht gelesen: nicht im Mount | der Mount ist die Wurzel des Hauptbaums (`a-check.mk`, `d-check.mk`) |
   | 10 | `AGENTS.md` oder Datei einer Regel von `file:` in einer Kopie unter `.wt/` über der Grenze | nicht betroffen | nicht gemeldet (gemessen: die Globs sind an der Wurzel verankert) | Nr. 1; Fänger Nr. 5 Zeile 9 |
   | 11 | eine Datei unter `.wt/` ist getrackt (committet) | ungelesen | ungelesen | akzeptiertes Negativ, Nr. 6 a |
   | 12 | Kopie unter `.wt/` ohne `.git` (kein `git worktree`) | ungelesen | ungelesen; `make zeichen-check` liest sie, wo der Ausschluss des Hauptbaums fehlt | Nr. 6 b |
5. **Zusage und Fänger.** Je Zusage ein Fall von `SPEC-102` (die Fälle laufen im Wegwerfverzeichnis mit den echten Konfigurationen; die Kennungen vergibt der Implementer in der Folge des Vertrags). Die Verletzung ist beidseitig dieselbe Datei; die Mutation setzt der Implementer selbst und meldet sie rot.

   | Nr. | Zusage | Fall | Mutation, die ihn rot macht |
   |---|---|---|---|
   | 1 | `a-check`: eine Verletzung unter `.wt/w/…` bleibt ungemeldet | Datei außerhalb jeder Schicht, die ein Paket des Modells importiert, unter `.wt/w/x/x.go`: Exit 0 | `exclude` aus `.a-check.yml` entfernt; Muster `.wt/*` |
   | 2 | `a-check`: dieselbe Datei im Hauptbaum wird gemeldet | dieselbe Datei unter `x/x.go`: Exit 1, `wrong-direction` | Muster `**` |
   | 3 | `a-check`: ein Verzeichnis `.wt` im Quellbaum bleibt geprüft | `src/hexagon/model/.wt/x.go` importiert einen Dienst: Exit 1, `core-impurity` | Muster `**/.wt/**` |
   | 4 | `a-check`: ein Arbeitsbaum in einem Arbeitsbaum bleibt ungemeldet | dieselbe Datei unter `.wt/a/.wt/b/x/x.go`: Exit 0 | Muster `.wt/*/x/**` oder ein anderes, das nur die erste Ebene trifft (Muster und Messung im Bericht) |
   | 5 | `docs-check`: ein Verweis auf eine fehlende Datei unter `.wt/w/` bleibt ungemeldet | Exit 0, `0 Befund(e)` | `".wt/**"` aus `scan.ignore` entfernt |
   | 6 | `docs-check`: derselbe Verweis in `x/a.md` wird gemeldet | Exit 1, `target-missing` | Muster `**` |
   | 7 | `docs-check`: `x/.wt/b.md` bleibt geprüft | Exit 1, `target-missing` | Muster `**/.wt/**` |
   | 8 | `docs-check`: `.wt/a/.wt/b/c.md` bleibt ungemeldet | Exit 0 | ein Muster, das nur die erste Ebene trifft |
   | 9 | `docs-check`: `file` und `structure` lesen `.wt/` nicht | die echten Blöcke `scan`, `structure`, `file`: `.wt/w/AGENTS.md` mit 50 000 Zeilen, Exit 0 | Glob `AGENTS.md` in `file:` zu `**/AGENTS.md` |

   Jede Zusage führt beide Seiten der Grenze: der Fall unter `.wt/` und der Fall am Pfad daneben stehen in den Zeilen 1 und 2 sowie 5 und 6.
6. **Ungeprüft und akzeptierte Negative.** Je Punkt Grund und Folge im Betrieb.
   - **a) Ein getrackter Pfad unter `.wt/`** (Nr. 4 Zeile 11). Grund: Ein Arbeitsbaum trägt eine Datei `.git`, `git add` nimmt ihn als eingebettetes Repository und legt nie seinen Inhalt ab; kein Gate fragt nach getrackten Pfaden unter `.wt/`. Folge: Eine eingecheckte Datei dort bliebe von `make a-check` und `make docs-check` ungelesen; `make zeichen-check` liest sie (`git ls-files`), und der Reviewer sieht sie im Diff.
   - **b) Eine Kopie unter `.wt/` ohne `.git`** (Nr. 4 Zeile 12). Grund: Kein Arbeitsablauf legt sie an; die Arbeitsbäume entstehen mit `git worktree add`. Folge: `make zeichen-check` und der Gate-Nachweis lesen den Inhalt, wo der lokale Ausschluss `/.wt/` in `.git/info/exclude` des Hauptbaums fehlt; der Fehler wäre sichtbar (Befund oder verändertes Nachweis-Hash), nicht still. Der Ausschluss steht nicht im Repo.
   - **c) `tools/harness/mutationslauf.sh`** (Nr. 1, letzte Zeile): kopiert `.wt/` in den Container. Grund: kein Gate, ändert kein Urteil (`go test ./...` überspringt Verzeichnisse mit Punkt-Präfix), und `SPEC-210` führt das Skript als Werkzeug ohne Zusage zur Laufzeit. Folge: ein größerer Tar-Strom je Lauf, proportional zur Zahl der Arbeitsbäume.
   - **d) Verhalten der Pins ohne Fall:** Symlinks (Nr. 4 Zeilen 4 und 5) und ein Muster ohne Treffer (Zeile 6). Grund: Eigenschaft der Werkzeuge, nicht Zusage dieses Repos, gemessen am Pin. Folge: Ein Pin-Wechsel kann sie ändern; er ändert die Prüfsumme der Gegenproben (Pin in den Eingaben), und die Fälle der Nr. 5 laufen mit dem neuen Pin.
   - **e) Der Kontext von `docker build`** (Nr. 1, Zeilen `lint` bis `groessenmarke`). Grund: Die Allowlist `*` in `.dockerignore` schließt `.wt/` aus, und kein Fall fährt einen Bau mit geänderter Datei. Folge: Fehlte das `*`, ginge `.wt/` an den Daemon (Zeit und Platz je Bau), aber nie in ein Ergebnis, weil die Stufen benannte Pfade kopieren.
   - **f) Die Gates, die nicht gefahren wurden** (Nr. 1): `make lint`, `make build`, `make test`, `make test-race`, `make handbuch-check`, `make katalog-check`, `make migration-check`, `make schema-check`, `make gegenproben`. Grund: Last des Hosts über der Grenze für schwere Läufe; der Mechanismus ist gelesen (Skripte, `Dockerfile`, `.dockerignore`) und für den Kontext gemessen. Folge: eine Abweichung fände erst der Lauf der Gates in einem Hauptbaum mit Arbeitsbaum; `make gates` im Hauptbaum mit einem Arbeitsbaum unter `.wt/` ist der Beleg, den der Verifier fährt.
   - **g) Verdrahtung.** Dass die Zeile in `.a-check.yml` und der Eintrag in `.d-check.yml` im Repo stehen, fängt nur Fall 1 und 5 der Nr. 5 (Mutation: Zeile weg). Dass `.a-check.yml` und `.d-check.yml` in den Eingaben der Prüfsumme der Gegenproben stehen, ist Bestand (`SPEC-040` Nr. 10 h).

**Stand:** vorgesehen; Vertrag vor dem Code. Geliefert wird mit den Zeilen von Nr. 3 und den Fällen von Nr. 5.

**Grenze:** Nimmt einen Pfad aus, der kein Stand ist; prüft nicht, ob ein anderer Pfad künftig zu einem Arbeitsbaum wird, und nicht, ob jemand etwas unter `.wt/` committet (Nr. 6 a).
