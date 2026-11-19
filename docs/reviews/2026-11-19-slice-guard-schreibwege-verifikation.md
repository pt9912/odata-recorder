# Verifikations-Report: slice-guard-schreibwege — 2026-11-19

**Rolle:** Verifier (Modul 11); Frage: Bauen wir es richtig, gegen Plan, DoD und Vertrag. Nicht die Frage des Reviewers (Diff gegen Plan und Hard Rules) und nicht die des Validators.

**Gegenstand:** `slice-guard-schreibwege`, Zweig `slice-guard-schreibwege`, HEAD `f89e63b9`, Basis `main` `ef8a0569`. Plan: [slice-guard-schreibwege](../plan/planning/done/slice-guard-schreibwege.md). Vertrag: [SPEC-214](../../harness/vertraege/SPEC-214-command-guard.md) Nr. 1 bis 12. Bezug: `LH-QA-06`, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (der Prüflauf ist keine Gegenprobe), `AGENTS.md` §3.1, §3.9, §3.10, §3.15 bis §3.18. Vorgänger: [Review](2026-11-19-slice-guard-schreibwege-review.md) (0 HIGH, 7 MEDIUM, 4 LOW, 5 INFO), danach die Entscheidungen des Architects (Rückgaben 8 bis 13) und die Fixes B1 bis B7 und F-2 bis F-6, F-10.

**Vorgehen.** Die Behauptungen des Implementer-Berichts habe ich nicht übernommen. Gefahren habe ich: den Prüflauf, `make gates`, die Zeitmessung und eigene Mutationen an Kopien des Hooks unter `<Scratchpad>/guard-schreibwege-verifier/` (nie am Arbeitsbaum; höchstens vier zugleich; die Mutanten sind nach dem Lauf entfernt). Der Zweig blieb unverändert: `git status --short` vor und nach den Läufen leer; der Prüflauf lässt kein `tmp.*` zurück (187 vorher, 187 nachher). Der Hook, der in meiner eigenen Sitzung lief, ist der alte von `main`; der Hook des Slice lief nur unter meinen Aufrufen.

## 1. Eigene Läufe (Belege)

| Lauf | Ergebnis | Last (1 min) |
|---|---|---|
| `bash tools/harness/guard-faelle.sh`, zwei Läufe am unveränderten Baum | `1842 Pruefungen, 0 Abweichungen in 24 s` und `in 26 s`; Median `git status` 12 und 10 ms, 32768 Zeichen 175 und 165 ms (Grenzen 60 s, 100 ms, 3 s) | 2,0 bis 2,2 |
| dasselbe in 67 Läufen der Mutanten, vier zugleich | 21 bis 26 s, Median 9 bis 12 ms, 32768 Zeichen 162 bis 288 ms | 2 bis 6 |
| Prüflauf mit `LC_CTYPE` und `LC_COLLATE` auf `de_DE.UTF-8` (Kopie, `LC_NUMERIC=C`) | `1842 Pruefungen, 0 Abweichungen in 33 s`, Median 14 ms, 32768 Zeichen 238 ms | 2,2 |
| `bash tools/harness/guard-faelle.sh x` und mit leerem Argument | Exit 2, `Aufruf` auf stderr (Fall 52) | 2 |
| `make gates` unter `flock /tmp/odata-recorder-heavy.lock` | **Exit 0**; darin `make a-check` `0 Befund(e)`, `make docs-check` `958 Datei(en) geprüft, 0 Befund(e)`, `make groessenmarke` `ok, Messung 1707 von Marke 2600`, `make zeichen-check` `ok, 1993 Datei(en)`, die Docker-Gates. Wanduhr 35 s: die Schichten kamen aus dem Cache, weil `src/` gegenüber `main` unverändert ist (der Slice berührt weder `src/` noch `Makefile`, `harness/mk/`, `Dockerfile`, `tools/gegenproben/`, den Stempel noch `harness/README.md`; `git diff --stat` leer) | 2,3 |
| Zeitmessung des Hooks von Hand (`EPOCHREALTIME`), `LC_ALL=C` und `de_DE.UTF-8` | siehe Abschnitt 6 | 1,7 bis 2,8 |
| Differenz zum Bestand: alle 325 Elemente der Fälle 1 bis 51 und 46 Befehle mit Host-Werkzeugen an beiden Hooks (alt `ef8a0569`, neu) | Sperre H nur dort gelockert, wo der Vertrag es sagt (Rumpf eines Heredoc, Nr. 1 e); 10 Formen sperren neu (`nohup`, `timeout`, `for … do`, `if … then`, `!`, `p\ip` und `\pip`, `sudo -u`, `xargs -0`, `flock`), keine Form, die der Bestand sperrte, geht durch; `command -v pip` geht in beiden durch | 2,3 |
| Rohbytes aller sieben geänderten Dateien (`LC_ALL=C grep -P`: BOM, NBSP, Nullbreite, Zeilentrenner, Steuerzeichen, CR, Backslash-u-vier-Hex) | 0 Treffer; Hook und Prüflauf sind reines ASCII; die Escape-Folge des Falls 40 entsteht aus Bytewerten (`printf` mit `$bs`) | — |
| Meldungstexte S, P, L, G gegen `SPEC-214` Nr. 1 d | wortgleich (awk-Auszug aus Vertrag und Hook, `cmp`); H unverändert gegenüber `ef8a0569` | — |

## 2. DoD, Punkt für Punkt (Plan §2)

| # | Punkt | Urteil | Beleg |
|---|---|---|---|
| 1 | Guard erfüllt `SPEC-214`, Namensregel in Übergangsregeln Nr. 8, jede Zusage hat Zeile mit rot gesehener Mutation | **erfüllt, mit Vorbehalt V-1 bis V-3** | alle 52 Fälle und 23 Mengen grün; Sondierung gegen den Wortlaut ohne abweichendes Verhalten; die Zeile 17 (Latenz) ist für andere Formen und die Locale des Hosts nicht erfüllt (V-1); drei Mengen tragen Elemente ohne Fänger (V-2, V-3) |
| 2 | Prüflauf belegt Sperre und Durchlass über 52 Fälle und die Schleifen; Zeit und Latenz gemessen **und in §7 berichtet** | **Prüflauf erfüllt; Ort offen** | 1842 Prüfungen, 0 Abweichungen, gemessen in §3 und §6, nicht in §7 (§7 steht auf „noch nicht begonnen“, Aufgabe des Planners) |
| 3 | Doku-Update Übergangsregeln Nr. 8, `harness/README.md` unverändert | **erfüllt** (Wortlaut: V-6) | Diff von `docs/plan/planning/README.md`; `harness/README.md` nicht im Diff |
| 4 | `make gates` grün | **erfüllt** | Abschnitt 1 |
| 5 | Review durchgeführt, Report liegt vor | **erfüllt** | `docs/reviews/2026-11-19-slice-guard-schreibwege-review.md` |
| 6 | Closure-Notiz mit Lerneintrag | **offen, Aufgabe des Planners** | §7 |
| 7 | Beobachtungs-Register fortgeschrieben | **offen, Aufgabe des Planners** | `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` |
| 8 | Jedes Risiko aus §6 trägt einen Ausgang | **erfüllt in der Form** | zehn Risiken, jede mit Ausgang; *weiter offen*: Commit-Wege, Quote-Blindheit, Architect ohne `Edit`, Prüflauf ohne Gate, Latenz, Wrapper und Löschwege |
| 9 | Die drei Paarungen | **offen, Closure** | — |

## 3. Plan gegen Code (Diff `ef8a0569..f89e63b9`)

Sieben Dateien: Hook (+554 −89), Prüflauf (+1064), `SPEC-214` und `harness/vertraege/README.md` (Architect), Übergangsregeln, Plan, Review. Die Pfadliste des Plans §3 stimmt: `src/`, `Makefile`, `harness/mk/`, `Dockerfile`, `tools/gegenproben/`, Stempel, `harness/README.md`, `.claude/agents/`, `.claude/commands/`, `AGENTS.md` sind unberührt (Ausschlüsse §1 gehalten). Der Vertrag hat nur drei Commits, alle des Architects (`d83a12e1`, `58ce6a9d`, `fc8264f3`); der Implementer hat ihn nicht geändert (`AGENTS.md` §3.12). Die Marke steht vor dem ersten Code-Commit (`b3134807` vor `0601382b`); die Messung 554 + 89 + 1064 = 1707 von 2600 stimmt (`git diff --numstat d83a12e1..HEAD`, `make groessenmarke` im Gate). Die Zahl 1842 Prüfungen stimmt mit 367 Elementen der 52 Fälle und 1475 Elementen der Mengen (Z5 bis Z28, nachgezählt). Die Zahl der Elemente je Fall im Prüflauf (`fall <Nr> <Klasse> <Zahl>`) gleicht der Zahl je Fall in `SPEC-214` Nr. 10 e; die Köpfe `fall` tragen 42 der 52 Fälle (die übrigen rufen `fall_roh` oder eigene Schleifen), und die drei Abweichungen (Fall 7, 11, 36) sind die Zusatzaufrufe, die der Vertrag mitzählt.

Nr. 8 der Übergangsregeln gegen `SPEC-214` Nr. 8 (Zeile 20): gleiche Sätze (Verzeichnis `<kurz>-<rolle>`, sechs Rollen, `-2`, Löschen nur unter dem eigenen Verzeichnis, `rm -rf <Scratchpad>/<kurz>-<rolle>`, keine generischen Namen auf oberster Ebene, der Guard prüft die Namen nicht). Abweichung im Satz über die Sperren: V-6.

## 4. Tabelle *Zusage · Fänger* (20 Zeilen) gegen den Vertrag, je Zeile eine eigene Mutation

Jede Mutation ist eine Änderung an einer Kopie des Hooks und ein Lauf des Prüflaufs der Kopie (deterministisch; jede in 1 von 1 Läufen). „rot“ nennt das erste Element, „grün“ heißt Exit 0. Alle Kennungen sind meine; keine stammt aus dem Bericht des Implementers. 68 Mutationen: 28 rot, 39 grün, eine am Hook gemessen (B14). Von den 39 grünen sind drei gleichwertig oder praktisch gleichwertig (V13, B05, V60), 36 haben eine trennende Eingabe; 22 davon habe ich an der Mutante mit einer Eingabe sondiert (Original gegen Mutante), die übrigen ändern den Ausdruck erkennbar. Die Vollständigkeit des Vertrags gegen die Tabelle: jede Nummer von `SPEC-214` Nr. 1 bis 11 hat eine Zeile (Nr. 3 a und Nr. 12 tragen Zeile 8, 16 und 19 bzw. keine: Nr. 12 ist eine Folge, kein Verhalten).

| Zeile | Zusage (Kurz) | Meine Mutation und Ergebnis |
|---|---|---|
| 1 | Umleitung nach *Repo* sperrt, jede Operatorform, jede Stelle | B02 Dup für jedes Wort hinter `>&`: rot Fall 1 El 10 und Z6 (10). **B01 nur die ersten zwei Umleitungen eines Segments prüfen: grün** (V-3) |
| 2 | Dup, Gerät, außen, kein Ziel, Zweifel gehen durch | B17 `2>&1` nicht als Dup lesen: rot Fall 3 (23 Abweichungen). **V12** Ziel beginnt mit `<` (`echo x > <(cat)`) als Operand lesen: grün; **V14** Dup nur einstellig (`>&10`): grün (V-4) |
| 3 | Zielklassen: Wurzel, Komponentengrenze, `..`, Tilde, `$PWD`, Zweifel | B03 `$PWD/…` nicht lesen: rot Fall 8 El 1 und Z6 (21). **V10/V11** `?` oder `[` nicht als Muster (Reihe 5): grün (V-2) |
| 4 | Verzeichnis-Zustand entscheidet relative Ziele | R7 `$` im Inneren des Operanden nicht lesen: rot Z23 (8). **V21** `cd ~Name` lässt den Zustand: grün; **V64/V66** Option vor dem Operanden (`cd -P docs`): V64 rot nur an den Formen ohne Operand, V66 grün (V-4) |
| 5 | Zitat und Backslash verbergen den Operator; offenes Zitat im Zielwort | B04 `\"` im Zitat nur ein Zeichen überspringen: rot Fall 12 El 7 |
| 6 | Heredoc und Here-String sind Eingabe; Commit-Wege gehen durch | B18 Zeilen mit `<<` nicht lesen: rot Fall 16, 17 u. a. (61). B05 Here-String-Zweig entfernen: grün, gleichwertig (der Bericht nennt es: die folgenden Zweige lesen `<<<` ohne Wort) |
| 7 | Rumpf: nicht geprüft, Rest der Zeile geprüft, nicht terminiert, `<<-` | R8 Tabulatoren immer entfernen: rot Z24 (4). **V73** Endewort mit angehängtem Leerzeichen als Ende lesen („genau WORT“): grün (V-4) |
| 8 | `tee` sperrt bei Operand *Repo* an jeder Stelle, nach `--`; Umleitungswörter keine Operanden | B06 Zweifel als *Repo* lesen: rot Fall 21 El 3. **V72 nur die ersten zwei Operanden prüfen: grün** (`tee /tmp/a /tmp/b docs/a.md` geht dann durch; V-3) |
| 9 | `sed -i` in jeder Form, auch im Scratchpad; Namensvettern durch | B07 Langform nur als `--in-place`: rot Fall 22 El 4, 5, Z9 (8). **V06** Bündel mit Zeichen außerhalb Buchstaben und Ziffern (`-n.i`), **V07** `-I`: grün (V-4) |
| 10 | `git apply` sperrt außer nur lesend oder `-C` außen | B08 `--stat` aus der lesenden Liste: rot Fall 25 El 2, Z12. **V18 `--check` nur als erstes Wort hinter `apply`: grün** (`git apply p --check` und `git apply --index --check p`; V-3) |
| 11 | Leer und ohne Text durch, Länge über 32768 G vor allem anderen | S1 `-ge` statt `-gt`: rot Fall 27, Fall 47, Z20 (7) |
| 12 | Segmente, Zeilenenden, Fortsetzung, Tiefe 3; Befehlsposition | R10 Tiefe ab 3 sperren: rot Fall 31 El 3, Fall 32 El 2. V65 `--` bei `cd` als Operand: rot Z23 (2). **V01** `-v` hinter anderem Präfix als `command` wie dort lesen (`sudo -v pip install x`), **V02** `sudo -- sed -i x f`, **V03** `sudo - sed -i x f`, **V04/V05** Zuweisungsname mit Ziffer oder Unterstrich am Anfang (`A1=x sed -i x f`, `_A=x …`): grün (V-4) |
| 13 | Rang und Meldungsklasse | B09 `Pattern-based` im Text S: rot Fall 37, 38 (636 Abweichungen mit allen Elementen der Klasse S). **T1 bis T5** Weg im Text jenseits des Merkmals (`git commit -F`, `docker rm -f`, `git rm`, `Write tool`) streichen: grün (V-5) |
| 14 | Bestand: Sperrliste, Extraktor-Zweifel, `python` | B19 `apt-get` aus der Liste: rot Fall 39 El 5. **B10, B11, B12** `cargo`, `conda`, `brew` aus der Liste: grün (V-4; die Liste führt 15 Wörter, Fall 39 vier) |
| 15 | `pkill`, `killall`, `pgrep -f` sperren; `pgrep` ohne `-f`, `kill` durch | B13 `--*` ausnehmen entfernen: rot Z11 `pgrep --list-full foo`. **V42 `-f` nur im ersten Wort hinter `pgrep`: grün** (`pgrep -u 1000 -f foo` geht dann durch; V-3) |
| 16 | Rekursives `rm` sperrt außer unter `…/scratchpad/<Name>` | R1 Komponente ohne Grenze davor: rot Z28 (`/xscratchpad/a`). **V08/V09** `?` oder `[` nicht als Muster (`rm -rf <Scratchpad>/a?`): grün (V-2). **V71 nur die ersten zwei Operanden, V45 Flag nur vor dem ersten Operanden (`rm docs/x -rf`), V19 Flag hinter `--`: grün** (V-3, V-4) |
| 17 | Latenz: Median 100 ms, 32768 Zeichen 3 s | B14 150 ms je Aufruf am Hook: Median 162 ms (5 Läufe), Fall 47 prüft `> 100000` µs; die Mutante habe ich am Hook gemessen, den vollen Prüflauf der Mutante (+0,15 s je Aufruf, rund 5 min) nicht abgewartet. **Die Zusage für 32768 Zeichen gilt nicht für jede Form und nicht in der Locale des Hosts (V-1)** |
| 18 | Prüflauf wird rot bei Abweichung, zustandslos, weist Argument ab | B15 Exit 1 nach jeder Sperre: rot Fall 1 und 1125 Abweichungen. Fall 52 am unveränderten Prüflauf (Abschnitt 1) |
| 19 | Akzeptierte Negative sind festgehalten (Nr. 11) | B16 `gofmt` sperren: rot Fall 50 El 7 |
| 20 | Negativ: Scratchpad-Namen sind Anweisung | keine Mutation möglich; Wortlaut gelesen (Abschnitt 3) |

**Tabelle gegen den Vertrag.** Jede Zeile hat Fall und Mutation; beide Seiten jeder Grenze führen Paare (Fall 27 und 28, Fall 12 und 13, Fall 15 und 16 und 17). Die Zeilen 8 und 16 tragen als Allaussage „an jeder Stelle“ bzw. „ein Operand genügt“ den Wortlaut des Vertrags; der Fänger führt zwei Operanden, nicht beliebig viele (V-3). Zeile 20 nennt Grund und Folge.

## 5. Die Fixes zu den 7 MEDIUM gegen die Klasse (`AGENTS.md` §3.17)

Je Befund bilde ich die Menge aus dem Wortlaut der Zusage in `SPEC-214`, nicht aus dem Befund, und mutiere an Elementen, die der Bericht nicht nannte.

| Befund | Menge (Wortlaut) | Meine Mutation an einem nicht genannten Element | Urteil |
|---|---|---|---|
| F-1 Befehlsposition (Nr. 2 h) | Schlüsselwörter 10, Präfixe 11, Optionen 29, Schreibweise 5 | R2 `sudo -D`, R3 `xargs -a`, R5 `flock -E`, R6 `time -o` je einzeln: rot (je 6); R4 `}`: rot Z27 (6); V51 Basisname nicht bilden: rot (10); V70 Backslash nur am Wortanfang entfernen: rot Z19 (6; die Form `/usr/bin/\w` trennt). Z18 enthält jede Option der Tabelle (Schleife über 11, 2, 1, 2, 8, 1, 2, 2 Formen) | **Klasse der Tabelle gedeckt.** Ungedeckt, weil nicht in der Tabelle, aber im Wortlaut „jedes Wort, das mit `-` beginnt“: `--`, `-` allein, `-v` hinter anderem Präfix, Zuweisungsnamen (V01 bis V05, V-4) |
| F-2 Löschgrenze (Nr. 7 c) | Operanden des rekursiven `rm`: Klasse, Stelle, Zahl, Komponente | R1: rot (1). V71 nur die ersten zwei Operanden: **grün**. V08 bis V11 Musterzeichen `?` und `[`: **grün** | **Klasse nur zum Teil gedeckt:** Z22 bildet Paare, kein Tripel, und Muster nur mit `*` (V-2, V-3) |
| F-3 Verzeichnis-Zustand (Nr. 3 c) | Operand von `cd`/`pushd`, `$` an jeder Stelle, ohne Operand | R7: rot (8). V21 `~Name`, V66 Option vor Operand: **grün** | Teil gedeckt (V-4) |
| F-4 Tabulator (Nr. 5 a) | Endewort × Operator × Einrückung | R8: rot (4). V73 Leerraum hinter dem Endewort: **grün** | Einrückung gedeckt, die Gegenseite „genau WORT“ am Zeilenende nicht (V-4) |
| F-5 Länge (Nr. 2 b, 9 c) | Länge 2 × Verstoß 5 | S1: rot (Fall 27, 47, Z20 mit 7) | gedeckt (die Mutation „G erst nach der Zerlegung“ habe ich nicht neu gefahren; Bericht Zeile 11) |
| F-6 Shell-Bündel (Nr. 2 g, 5 c) | Shell 5 × Bündel mit `c` 8 / ohne 3 | R9b Bündel höchstens zwei Buchstaben vor und hinter `c`: **grün** (Z21 führt Bündel bis zwei vor und einen hinter `c`; Nr. 2 g nennt keine Obergrenze). V15 `-c` nur im ersten Wort hinter der Shell (`bash -x -c "…"`, `bash --norc -c "…"`): **grün**. V16 Verzeichnis-Zustand beginnt im `-c`-String neu (`cd /tmp && bash -c "echo x > a.md"`): **grün**. V17 Shell hinter Präfix liest den Heredoc (`cat <<EOF \| sudo bash`): **grün** | Bündel gedeckt; die Stelle des `-c`, der Zustand in der Rekursion und die Shell hinter einem Präfix nicht (V-4) |
| F-7 Vertragsberichtigung | Rumpf mit Ersetzung bei unzitiertem Endewort | Fall 50 El 17, 18 am Hook sondiert: `cat <<EOF` mit `$(sed -i x y)` im Rumpf geht durch, mit `<<'EOF'` ist es Text; Nr. 5 b, 5 f, 1 e, 11 d tragen die Berichtigung, Grund und Folge stehen in Nr. 11 d | erfüllt |

## 6. Latenz und Zeitbudget (`SPEC-214` Nr. 10 c), selbst gemessen mit Last

Der Prüflauf (Fall 47) misst Median und eine Form mit `LC_ALL=C`; er setzt `LC_ALL=C` global (nötig für den Dezimalpunkt von `EPOCHREALTIME`). Grenzen und Belege:

| Messung | Ergebnis | Grenze |
|---|---|---|
| Median von 20 Läufen `git status` (Prüflauf, Last 2,2) | 10 bis 12 ms; von Hand 9 bis 10 ms in `C` und `de_DE.UTF-8` | 100 ms |
| `echo a > <SP>/x;` bis 32768 Zeichen (Fall 47), `C` / `de_DE.UTF-8` | 163 bis 175 ms / 204 bis 210 ms | 3 s |
| Skript | 24 bis 26 s (Last 2,2), 21 bis 26 s mit vier Prüfläufen zugleich | 60 s |
| 32768 Zeichen, gleiche Form mit Zeilenvorschub je Segment, `de_DE.UTF-8` | 412 ms | 3 s |
| **`'>` wiederholt bis 32768 Zeichen** (Zitat und Operator im Wechsel) | **`C` 4,3 bis 4,4 s, `de_DE.UTF-8` 17,4 s** (je zwei Läufe, Last 2,5 bis 2,8) | **3 s überschritten** |
| **`<<A ` wiederholt bis 32768 Zeichen** (8192 Heredoc-Operatoren) | **`C` 2,7 s, `de_DE.UTF-8` 18,3 s** | **3 s überschritten in der Locale des Hosts** |
| **`echo "{\"k\":\"v\"},…" > /tmp/x`, 32700 Zeichen** (JSON in Anführungszeichen mit Escapes) | **`C` 0,95 s, `de_DE.UTF-8` 7,4 bis 7,6 s** | **3 s überschritten in der Locale des Hosts** |
| `echo '{"k":"v"},…' > /tmp/x`, Wörter in `"a" ` und `'a' `, `a\ `-Folgen (je 32768) | 29 bis 57 ms | — |
| `$(` × 16384, `cd x;` × 6553, `bash -c ` × 4096, `if ` × 10900 mit `sed -i` am Ende | 127 bis 819 ms | — |

Ursache: `skip_quote`, `redir_check` und `find_heredocs` bilden je Zitat oder Operator eine Teilkopie `${t:i}` des Rests (quadratisch); in einer UTF-8-Locale kostet jede Kopie das Vier- bis Siebenfache. Der Hook trägt in `.claude/settings.json` kein `timeout` (Vorgabe des Hosts); eine Ausführung von 18 s führt also zu einer Wartezeit, nicht zu einem Durchlass. Die Locale meiner Sitzung ist `de_DE.UTF-8`; welche Locale der Host dem Hook gibt, habe ich nicht gesehen (ungeprüft). Dazu: `${#cmd}` zählt in `C` Bytes, in UTF-8 Zeichen (32763 `ä` nach `echo `: `C` sperrt G, `de_DE.UTF-8` lässt durch); `SPEC-214` Nr. 2 b sagt „in Zeichen“, der Prüflauf fährt nur `C` (V-1, V-8).

## 7. Rückgabe 14 (Adresse Architect): vollständig benannt, und was sie im Betrieb heißt

`command -p -v pip install x` sperrt (H): am Hook im Repo bestätigt, ebenso `command -pv pip install x`, `command -p pip install x` (H, richtig) und `command -p -V pip` (H); `command -v pip` und `command -V pip` gehen durch.

**Vollständig benannt:** Gegenstand, Wirkung (H, Sperre ohne Anlass), Ursache (Nr. 2 h nennt `-v`/`-V` nur unmittelbar hinter `command`), Fänger-Lücke (Z18 führt die Form nicht), die zwei Auflösungen, Adresse, Ausgang *offen*, Folge im Betrieb („keine, kein Anwendungsfall“); sie steht im selben Commit wie der Bericht und im Absatz *Ausgänge der Aufträge an den Architect*. Die Folge im Betrieb ist richtig: die Form sperrt zu viel (die Meldung H nennt „make/Docker-only“ für eine Abfrage und ist irreführend), sie öffnet keinen Weg; die Abhilfe `command -v pip` geht durch. Zu ergänzen für den Architect: (a) die Klasse hat vier Elemente (`-p -v`, `-p -V`, `-pv`, `-pV`), der Plan nennt eines; (b) die Zeile, die `-v` und `-V` ausnimmt (`[ "${pw##*/}" = command ]`), hat selbst keinen Fänger: V01 (die Prüfung auf `command` entfernen, `sudo -v pip install x` geht dann durch) bleibt grün; wer Rückgabe 14 löst, braucht dort einen Fall (V-4); (c) die Rückgabe steht nur in der Tabelle der Rückgaben, nicht unter den Risiken von §6 (wie F-12 des Reviews zur Rückgabe 8; die DoD „jedes Risiko trägt einen Ausgang“ berührt das nicht). Rückgaben 2 und 3 (Adresse User, `.claude/`) sind bekannt und von mir nicht bewertet.

## 8. Befunde

| ID | Schwere | Befund | Fundstelle | Zuständig |
|---|---|---|---|---|
| V-1 | MEDIUM | Die Zusage „ein Befehl von 32768 Zeichen höchstens 3 s“ (`SPEC-214` Nr. 10 c, Tabelle Zeile 17) hat einen Fänger für eine Form (`echo a > <SP>/x;`) in `LC_ALL=C`. Gemessen: `'>`-Folge 4,3 s (`C`) und 17,4 s (UTF-8), `<<A `-Folge 2,7 s und 18,3 s, JSON in Anführungszeichen mit Escapes (32700 Zeichen) 0,95 s und 7,5 s. Der Guard urteilt richtig, die Wartezeit liegt unter dem Standard-Timeout des Hosts; die Zusage ist für diese Formen und für die Locale gebrochen, und kein Lauf sieht es (Verifier-only-Klasse). Der Abschnitt über Prozesse je Wort (Nr. 11 j) nennt die Quadratik nicht | Hook `skip_quote`, `redir_check`, `find_heredocs`; `SPEC-214` Nr. 10 c, 11 j; Plan §3 Zeile 17 | Architect (Zusage auf Form und Locale begrenzen oder Negativ mit Zahlen oder Grenze G senken); Implementer (Fänger mit UTF-8 und den Formen) |
| V-2 | MEDIUM | Die Menge der Musterzeichen (`SPEC-214` Nr. 3 b Reihe 5, Nr. 7 c i: `*`, `?`, `[`) trägt einen Fänger für `*`. V08 und V09 (`?` oder `[` in der Vorprüfung von `rm`), V10 und V11 (in `classify`): grün; mit V08 geht `rm -rf <Scratchpad>/a?` durch statt L. Das ist der Kern der Gruppe 2 („`rm -rf` mit Muster“). Z22 führt `<SP>/*` und `$d/*`, Fall 44 `<SP>/*` und `<SP>/mut-*`, Fall 9 `a*.md` | Hook `rm_check`, `classify`; Prüflauf Z22, Fall 9, 44 | Implementer (Elemente nachziehen, kein Guard-Fix); Verhalten stimmt |
| V-3 | MEDIUM | Stelle und Zahl: der Wortlaut sagt „an jeder Stelle“ und „ein Operand genügt“ (Nr. 4 a, 6 a, 6 c, 7 a, 7 c); die Fänger führen Flag und Operand an der ersten oder zweiten Stelle. Grün: V71 (`rm`: nur zwei Operanden), V72 (`tee`: nur zwei Operanden), B01 (nur zwei Umleitungen je Segment), V42 (`pgrep -u 1000 -f foo`), V45 (`rm docs/x -rf`), V18 (`git apply --index --check p`). `sed` führt beide Stellungen (Z9) und zeigt, dass die Klasse bekannt war | Hook `rm_check`, `tee_check`, `redir_check`, `pgrep_check`, `git_check`; Prüflauf Z9 bis Z12, Z22, Fall 2, 20 | Implementer (Tripel, Flag hinter Operand, dritte Umleitung als Elemente); Verhalten stimmt |
| V-4 | LOW | Einzelne Elemente von Mengen des Vertrags ohne Fänger (Mutation grün, trennende Eingabe in Klammern): `-v` hinter anderem Präfix als `command` (V01 `sudo -v pip install x`; bei Rückgabe 14 mit lösen), `--` und `-` hinter einem Präfix (V02, V03), Zuweisungsnamen mit Ziffer oder Unterstrich am Anfang (V04, V05; Z8 führt `X=1`), `sed`-Bündel mit Zeichen außerhalb Buchstaben und Ziffern, `-I` (V06, V07), Ziel mit `<` am Anfang (V12), Dup mehrstellig `>&10` (V14), `-c` nicht am ersten Flag-Wort (V15), Verzeichnis-Zustand im `-c`-String (V16), Heredoc an `sudo bash` (V17), `rm`-Flag hinter `--` (V19), `cd ~Name` und Option vor dem Operanden (V21, V66), nicht absolutes `HOME` bei `rm` (V23), Endewort mit Leerraum dahinter (V73), Bündel länger als zwei Buchstaben vor und hinter `c` (R9b, Prüfgrenze), Wörter der Sperrliste `cargo`, `conda`, `brew` (B10 bis B12, Bestand), abschneidendes `)` am Ziel (V60, wirkt nur beim Ziel gleich der Wurzel). Kein Verhalten weicht vom Wortlaut ab | Hook, Prüflauf Z8, Z18, Z21 bis Z24, Fall 9, 39 | Implementer (Elemente) oder Architect (Negativ mit Grund und Folge) |
| V-5 | LOW | Die Zusage „die Meldung nennt den Weg“ (Nr. 1 d) hat keinen Fänger über das Merkmal hinaus: T1 bis T5 (den Weg in S, P, L, G streichen) bleiben grün. Die Texte sind wortgleich zum Vertrag (Abschnitt 1) | Fall 37, 38; Tabelle Zeile 13 | Implementer (Vergleich mit dem Text des Vertrags) |
| V-6 | LOW | Übergangsregeln Nr. 8: „sperrt `sed -i`, `git apply`, `tee` und die Umleitung … auf einen Pfad unter der Wurzel des Repos“ liest sich, als gelte die Einschränkung für alle vier; `sed -i` sperrt auch im Scratchpad (Nr. 6 b, Fall 22 El 10), `git apply` auch ohne Pfad (Nr. 6 c). Der Satz über die Folge (die Kollision fällt am überschriebenen Lauf auf) steht nur im Vertrag und im Plan | `docs/plan/planning/README.md` Nr. 8 | Planner |
| V-7 | INFO | DoD Punkt 2 nennt §7 als Ort der Zeit- und Latenzmessung; sie steht in §3 und §6, §7 ist der Closure vorbehalten | Plan §2, §7 | Planner |
| V-8 | INFO | `SPEC-214` Nr. 10 e nennt die Mengen Z21 bis Z28 nicht (Shell-Bündel, `rm`-Operanden, `cd`, Heredoc-Einrückung, Ziffernzahl, `sed`-Ziffern, Klammern, Komponente `scratchpad`); der Prüflauf führt sie. Nr. 2 b „in Zeichen“ gilt nur in einer UTF-8-Locale; der Prüflauf fährt `C` (Abschnitt 6) | `SPEC-214` Nr. 2 b, Nr. 10 e | Architect |
| V-9 | INFO | `AGENTS.md` §3.18: der Bericht des Implementers nennt die gegengelesene Datei nicht (der Plan nennt nur die Regel); ich habe sieben Dateien gegengelesen (Abschnitt 1) | Plan §3 | Implementer |

## 9. Was ich nicht prüfen konnte

- **Locale des Hooks im Betrieb:** gemessen mit `LC_ALL=C` und `de_DE.UTF-8` im Aufruf von Hand; welche Locale Claude Code dem Hook gibt, habe ich nicht gesehen.
- **Der volle Prüflauf der Mutante B14** (150 ms je Aufruf): nicht abgewartet; am Hook gemessen (Median 162 ms) und gegen die Prüfung in Fall 47 gelesen.
- **`make gates` aus kaltem Cache:** der Lauf kam fast ganz aus dem Cache der Docker-Schichten (`src/` ist gegenüber `main` unverändert); ein Kaltlauf ist nicht gefahren und nicht nötig für die Aussage über diesen Slice.
- **Mutationen der Zeilen 7, 8, 13 bis 16 an jedem Element:** die 29 Optionen und elf Präfixe habe ich nicht einzeln gefahren (vier Optionen und `}` je einzeln: rot; Z18 führt jede Option der Tabelle als Element, nachgelesen); die Mutationen „G erst nach der Zerlegung“ (F-5) und die Umleitungswörter Z13 bis Z16 habe ich nicht neu gefahren.
- **Die Wirkung von `git apply` ohne `Edit` im Rollen-Typ des Architects** (Rückgabe 2) und die Rollen-Anweisungen (Rückgabe 3): Adresse User, unbewertet.
- **Zufallsläufe:** keine eigenen; der Review hat 2400 Befehle ohne Abbruch gefahren.

## 10. Verdikt

**Bestanden, mit Auflagen vor der Closure.** Kein HIGH. Der Guard stimmt in allen meinen Sondierungen mit dem Wortlaut von `SPEC-214` überein (rund 100 Sonden, 22 trennende Eingaben an Mutanten, die Differenz zum Bestand, alle 52 Fälle und 23 Mengen grün, in `C` und in einer UTF-8-Ctype-Kopie); `make gates` Exit 0; die Tabelle trägt für jede der 20 Zeilen mindestens eine rot gesehene Mutation (meine und die des Berichts, Abschnitt 4); die Fixes der sieben MEDIUM des Reviews decken ihre Klassen, soweit der Befund sie nannte, und lassen Elemente derselben Mengen ungefangen (Abschnitt 5). Die Befunde sind Lücken der Fänger und eine Zusage der Zeit, nicht Fehler im Urteil des Guards.

**Auflagen** (merge-blockierend ist keine, wenn sie vor der Closure als Rückgaben in §6 mit Adresse und Ausgang *offen* stehen, `AGENTS.md` §3.13, oder behoben sind):

1. **Architect:** V-1 (Zusage der Zeit auf Form und Locale begrenzen, als Negativ mit Zahlen in Nr. 11 j aufnehmen oder die Grenze G senken), V-8 (Zeichen und Bytes, Mengen Z21 bis Z28 in Nr. 10 e); Rückgabe 14 mit den Ergänzungen aus Abschnitt 7 und einem Fänger für die Zeile, die `-v` und `-V` ausnimmt.
2. **Implementer** (nur Prüflauf, kein Guard-Fix; Elemente vor dem Wortlaut der Zusage gebildet, `AGENTS.md` §3.17): V-2 (Musterzeichen `?` und `[` in Z22 und Fall 9), V-3 (Tripel, Flag und `--check` hinter einem anderen Wort, dritte Umleitung), V-5; V-4 nach dem Urteil des Architects (Element oder Negativ). Ein zweiter Lauf des Prüflaufs in einer UTF-8-Locale, wenn V-1 so entschieden wird.
3. **Planner:** V-6 (Wortlaut der Übergangsregeln Nr. 8), V-7 (Ort der Messung), die Befunde V-1 bis V-5 und die Rückgabe 14 als Zeilen in §6, die Closure mit dem Lerneintrag zur Klasse „Zusage der Zeit an einer Form und einer Locale“ und zur Klasse „Allaussage ‚an jeder Stelle‘ mit Fänger an der ersten und zweiten Stelle“ (zweimal im selben Slice: V-2, V-3).

**Nicht behauptet:** ein Kaltlauf von `make gates`, die Locale des Hooks im Betrieb, ein Urteil über die Rückgaben 2 und 3.
