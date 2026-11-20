# Verifikationsbericht: slice-regeln-verkoerpern — 2026-11-20

**Rolle:** Verifier (Frage: bauen wir es richtig, gegen Plan und DoD) · **Gegenstand:** Zweig `slice-regeln-verkoerpern`, HEAD `47e1f32d`, Basis des Architect `8433db48` · **Datum:** 2026-11-20

**Eingangs-Kontext:** Plan `slice-regeln-verkoerpern` (§1 Entscheidung des Users, §2 DoD, §3 A bis I, §4, §6); [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md); `SPEC-040` Nr. 8 b, Nr. 9; `SPEC-201` Nr. 11; `LH-QA-06` (Bezug); `AGENTS.md` §3.6, §3.9, §3.13, §3.15, §3.17, §3.18; Review `2026-11-20-slice-regeln-verkoerpern-review` (Fixes in `47e1f32d`).

Alle Mutationen liefen auf Kopien (`git archive HEAD`) im Scratchpad, nie im Zweig. Kein Push, kein Merge, keine Änderung am Quellbaum.

## Ergebnis

**Freigegeben, mit zwei Auflagen an den Planner für die Closure** (V-1, V-3; beide betreffen Plantext, keinen Code und keine Guide-Datei). Die acht Regeln stehen an Ort und im Wortlaut des Plans, die Kürzungen haben ihren Träger, die Grenzen gleichen dem Ist, die Summe sinkt je Maß, kein Gate ist gelockert. Findings: 1 MEDIUM, 1 LOW, 2 INFO.

## Sensor-Belege (selbst gefahren, nicht aus der Behauptung übernommen)

| Sensor | Ergebnis |
|---|---|
| `make docs-check` (Worktree, unter Sperre) | `968 Datei(en) geprüft, 0 Befund(e)` |
| `make zeichen-check` | `ok, 2003 Datei(en) geprüft` |
| `make gates` (Worktree, HEAD `47e1f32d`, sauberer Baum, unter Sperre) | Exit 0; im Plan nennt der Bericht des Implementers keinen Lauf (V-3) |
| `make groessenmarke` | `kein Block `groessenmarke`, der Plan wird nicht gemessen` (Hinweis, Exit 0; `SPEC-128` Nr. 10 d) |
| Rohbytes (`LC_ALL=C grep -P`: Backslash-u-vier-Hex, BOM, NBSP, Zeilentrenner, Nullbreite, CR) | über `AGENTS.md`, `implement-slice.md`, sechs Typen, Skill, `.d-check.yml`, Plan: ein Treffer, das Beispiel in `AGENTS.md` §3.18 (unverändert, nicht berührt) |
| Messung `wc -l`/`wc -c` | `AGENTS.md` 323/23905 · `implement-slice.md` 212/16180 · `architect.md` 50/3614 · Skill 151/8668 · `conventions.md` 228/16065 · `implementer` 49/3209 · `verifier` 48/3046 · `reviewer` 42/2682 · `validator` 43/2633 · `planner` 38/2306 · `close-welle` 146/11731 · `plan-welle` 99/6554 · Skill `closure-note-reviewer` 90/4144 |
| Diff nach §4 (`git diff --numstat 8433db48..HEAD`, fünf Pfade) | 81 Einfügungen, 75 Löschungen, Summe 156 (Erwartung 148, Rückführungsschwelle 400) |

## DoD (§2) Punkt für Punkt

| Punkt | Stand | Beleg |
|---|---|---|
| Acht Regeln am Ort und im Wortlaut, Anker, Zeiger | erfüllt im Code; `liegt in` in §7 steht aus (Closure) | Diff gegen Abschnitt B zeichengleich (Skill 3 Blöcke, Schritt 17 und 25, K3, K4, K1, K2, §3.10, §3.15 mit beiden Überschriften, `tools:`, Satz in sechs Typen, Absatz im `verifier`); Anker-Zählung `welle-mock`/`welle-hybrid`: `AGENTS.md` 0/2, `implement-slice.md` 1/1, Skill 3/1, `verifier.md` 1/1, alle anderen 0/0 (Soll aus Abschnitt F); Zeiger `SPEC-201` Nr. 11 in Schritt 17 und Prüfzeile; Zeiger *Übergangsregeln Nr. 8* löst auf (`docs/plan/planning/README.md` Nr. 8) |
| Grenzen gehalten, keine Hebung ohne Senkung, `make docs-check` grün | erfüllt | Grenzen gleich Ist (alle fünf Regeln); Zeilen 973 auf 964 (323+212+50+151+228), Bytes 68648 auf 68432 (23905+16180+3614+8668+16065); Hebung AGENTS +272 Bytes, Skill +6 Zeilen/+558 Bytes unter Senkung Commands −9/−901 und Agenten −3/−145: je Maß unter einer Senkung (`SPEC-040` Nr. 8 b), alles in einem Commit `c3741274`, Messung und Grund in Kommentar und Commit-Nachricht; `47e1f32d` ändert an `.d-check.yml` nur Kommentarzeilen; Messung vor/nach in §7 steht aus (V-3) |
| `make gates` grün | erfüllt (selbst gefahren); Beleg im Plan fehlt | siehe Tabelle oben, V-3 |
| Review, Report unter `docs/reviews/` | Review durchgeführt; Report liegt noch im Scratchpad | siehe Auflagen |
| Doku-Update: `SPEC-201` Nr. 11 unverändert; Register | `SPEC-201` unverändert (`git diff` leer); Register steht aus | Register: sechs `state.md` tragen noch *offen*/*geplant* |
| Closure-Notiz, Register-Fortschreibung, Ausgänge in §6, drei Paarungen | stehen aus (Planner) | siehe Auflagen |

Plan-Überschriften: `## 1.` bis `## 8.` ohne Lücke (`grep -n '^## '`: 27, 63, 81, 376, 396, 400, 436, 443); F-2 des Reviews behoben. F-6 behoben (Tabelle und Abschnitt D nennen 48/3046 und 38/2306).

## Entscheidungs-Konformität

- **User-Entscheidung §1** (acht Regeln: drei aus `welle-mock`, E1, E2, E4 mit zwei Beobachtungen, Rückgaben 2 und 3): alle acht im Code; nicht gewählte Punkte (E3, E5, E7 bis E9, Nr. 3 und 4 aus §7 von `slice-guard-schreibwege`) nicht berührt (`git diff --stat`: keine weitere Datei).
- **`AGENTS.md` §3.6** (keine Lockerung): erfüllt, Summe sinkt je Maß; `.d-check.yml` Blöcke `scan` und `structure` unverändert (Abschnittsgrenze 29).
- **§3.9** (Plan folgt der Korrektur): Plan §3, Kopf, §6 und I nachgezogen; Rest siehe V-2.
- **§3.13** (Adresse nimmt an): offene Punkte in §6 (Stempel, E4-Einspruch, Zeiger Nr. 8) tragen Adresse und Ausgang; Nehmer `slice-groessenformel-neu` unverändert.
- **§3.18**: gegengelesen (Tabelle oben).
- **[ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md)**: gilt unverändert, keine neue ADR nötig.
- **Kürzungen K1 bis K6**: Träger selbst nachgelesen. K1 `harness/conventions.md` §Adoptierte Konventions-Quellen (Adresse des ZIP, Pfad); K2 `harness/README.md` §Guides (Zeile der Vorlagen, *Referenz-Form … Ziel-Form*); K3 Kopf von `.claude/hooks/stop-require-gates.sh` (DEFAULT, STRENG mit Datei und Variable, fehlender Stempel, GRENZEN); K4 `.githooks/commit-msg` (Aktivierung, `--no-verify`), `harness/mk/hooks-install.mk` (skip-if-present, eigener Hook bleibt, Prüfung wird kanonisch neu geschrieben), `tools/harness/commit-msg-traceability.sh` (`patterns=`, Grenze Anwesenheit/Wahrheit); K5 Vorlage `reviewer.template.md` der Baseline führt den Block, HIGH-Liste trägt Spec-Straten, Geheimnis, Zustandsmodell, Zustandsfeld, Kommentar; K6 Absatz *Warum dieser Typ existiert* im Typ `architect`. Eingehende Verweise auf die gekürzten Überschriften: keine außerhalb des Plans (`grep` über `*.md`, `*.yml`, `*.sh`, `*.mk`). Keine Zusage verloren.
- **Anker auflösbar:** `done/welle-mock-results.md` und `done/welle-hybrid-results.md` liegen; die Ziele der künftigen `liegt in`-Felder tragen den Anker (`AGENTS.md` §3.10 und §3.15 in der Überschrift, die drei anderen Dateien am Satz).

## Tabelle Zusage · Fänger gegen die Zusagen, eigene Mutationen (Kopien)

| Nr. | Meine Mutation (nicht die des Plans oder des Reviews) | Ergebnis |
|---|---|---|
| 2 | Zeiger `SPEC-201` aus der Prüfzeile genommen; Wortlaut der Pflicht in die Prüfzeile gesetzt | Zeiger: rot (`grep -c SPEC-201` 0 statt 1); Wortlaut: **grün** (Zählung bleibt 1; durch N7 benannt) |
| 6 | je Datei die Zeile `ohne Pfad und` gelöscht (3 Läufe); `ohne` durch `mit` in `verifier.md` | 3 von 3 rot; Inhaltswechsel rot |
| 10 | je Typ die Zeile mit dem Token gelöscht (6 Läufe); Token ohne Schrägstrich; Satz als Ganzes je Typ (Zeilen zusammengefügt) | 6 von 6 rot; Variante rot; alle sechs tragen den vollen Satz |
| 11 | Zeilen oder Bytes an Dateien angehängt, die **nicht** das Maximum ihrer Regel sind: `implementer.md` +2 Zeilen (51), `AGENTS.md` +113 Bytes ohne neue Zeile, `closure-note-reviewer.md` auf 152 Zeilen, `close-welle.md` auf 213 Zeilen | 4 von 4 rot, je mit `file-lines-exceeded` oder `file-bytes-exceeded` und der richtigen Datei |
| 12 | §3.10 +9 und §3.15 +14 Zeilen in einem Lauf | genau ein `section-lines-exceeded` (§3.10, 30), §3.15 mit 29 grün: Ist 21 und 15 bestätigt |
| 13 | `max-bytes` der Agenten 3613; `max-lines` von `AGENTS.md` 330 | 3613: rot (`file-bytes-exceeded`, 3614 gegen 3613); 330 über Ist: **grün** (N4); Summe und Gleichheit von Hand gerechnet, stimmen |
| 14 | `katalog-ausnahmen.txs` in Schritt 17 | `codepath-missing` an Zeile 118 |
| 5 | E1-Satz hinter den Fixture-Satz verschoben | Reihenfolge (Zeile 240 Fänger-Satz, 242 E1, 241 Fixture): rot |
| 7 | je Fragment des Satzes gestrichen (sechs Fragmente) | `ungefangene Klasse`, `mit ihrem Maß`: rot; `(n von m)`, `nicht eine Aufzählung`, `Fenster als Zahl`, `die nächste Messung überholt`: **grün** (4 von 6) |
| 1, 3, 8 | je ein Halbsatz der Zusage gestrichen (Schritt 17: Satz mit `harness/katalog-ausnahmen.txt`; Vorlage 1: `oder aus einer Menge im Modell mit Zählprüfung`; Laufzeit-Satz: `und nur, wenn der Unterschied die Streuung übersteigt`) | alle drei **grün**: Anker-Zählung, Zeilenfolge und Zeilenabstand bleiben (siehe V-1) |

Die Zeilen 4, 9, 15 und 16 habe ich nicht neu gemutet: 4 und 16 folgen demselben Zählmuster wie 6 und 10 (Reviewer und Implementer rot gesehen); 9 ist eine exakte Zeile; 15 läuft über `make zeichen-check` (grün).

## Akzeptierte Negative (Grund, Folge, Maß)

N1 bis N9 nennen alle Grund, Folge und Klasse mit Maß (n von m). Sachlich stimmen N1 (10 von 10, durch Löschen je eines Satzes in Kopien bestätigt), N3, N4 (330 über Ist grün), N6, N8 (`SPEC-999` grün laut Review), N9 (3 von 3: Zeilen 3, 5, 8 sind Handproben). **N7 stimmt im Maß nicht** (V-1).

## Findings

| ID | Kategorie | Befund | Fundstelle | Verifizierbar |
|---|---|---|---|---|
| V-1 | MEDIUM | N7 nennt als Maß *7 von 16 Zeilen der Tabelle (2, 4, 5, 6, 7, 10, 16)* und *neun Mutationen des Reviews*. Die Klasse *Teilmutation eines zugesagten Satzes* trifft mehr: die Zeilen **1, 3 und 8** lassen eine Teilmutation grün (Zeile 1: der Halbsatz mit `harness/katalog-ausnahmen.txt` gestrichen, Anker und `SPEC-201. Nr. 11` bleiben 1; Zeile 3: `oder aus einer Menge im Modell mit Zählprüfung` gestrichen, Zählung 3 und Zeilenfolge unverändert; Zeile 8: Halbsatz gestrichen, Zeile bleibt Rate plus 1). Das Maß ist 10 von 16 (1, 2, 3, 4, 5, 6, 7, 8, 10, 16). Der Reviewer-Table nennt acht grüne Mutationen, nicht neun. Das ist die Form, die die in diesem Slice geschriebene Regel (E4: Maß, nicht eine Aufzählung) trifft. Folge im Betrieb unverändert (Reviewer am Diff, Retirement-Check), nur die Zahl ist zu klein. | Plan §3 F, N7, Satz *Klasse und Maß* | ja (Mutationen oben) |
| V-2 | LOW | Die Regel E4 (Negativ mit Klasse und Maß) steht nur in `AGENTS.md` §3.15. Der Prüfer *Akzeptiertes Negativ ohne Grund und Folge* im Skill (`.harness/skills/reviewer.md` Zeile 76) und Schritt 14 von `implement-slice.md` (`Grund und Folge`) nennen weiter nur Grund und Folge; kein Konsument fragt nach Klasse und Maß. Nicht im Umfang (Auftrag des Users und Abschnitt B nennen allein §3.15 und §3.10), also kein Verstoß gegen den Plan; aber dieselbe Aussage an mehreren Stellen (`AGENTS.md` §3.9). | `.harness/skills/reviewer.md:76`, `.claude/commands/implement-slice.md:104` | ja (`grep -rn 'Grund und Folge'`) |
| V-3 | INFO | Der Bericht des Implementers (Abschnitt I) nennt keinen Lauf von `make gates`, der Plan (§3 H Nr. 4) verlangt ihn, der Reviewer fuhr ihn nicht. Beleg jetzt: mein Lauf, Exit 0. Die DoD verlangt Messung vor und nach in §7; sie steht noch in §3 I. | Plan §3 I, §2 DoD | ja |
| V-4 | INFO | `harness/gegenproben.stempel` ist am Zweig bewusst veraltet (`.d-check.yml` steht in den Eingaben); `make gegenproben` fuhr ich nicht (Adresse Integrationsschritt, Nr. 2 der Übergangsregeln, §6 Ausgang *offen*). Die Fälle 156 bis 170 und 213 lesen die Grenzen der echten Datei; die neuen Grenzen ändern ihren Ausgang laut Plan nicht (nicht von mir gefahren). | Plan §6, Abschnitt D | nein |

Negativbefunde ohne Befund: Wortlaut der acht Regeln (Zeichen für Zeichen gegen Abschnitt B); alle sechs Kürzungen; `.d-check.yml` (Werte, Summe, Verschiebung, Kommentar nach dem Fix F-1: nennt Messung vorher und nachher je Datei, Summe und Grund, wie `SPEC-040` Nr. 8 b verlangt); Raster in §6; Größenmarke (kein Block, Hinweis gefahren); kein Produktcode, keine ADR, kein Geheimnis berührt.

## Was der Planner in der Closure noch tun muss

1. **N7 berichtigen (Auflage, V-1):** Maß auf *10 von 16 (Zeilen 1, 2, 3, 4, 5, 6, 7, 8, 10, 16)* und die Zahl der Mutationen des Reviews auf die gezählten acht; dasselbe in §7. Das ist Plantext im Zweig (Closure-Commit), kein Code.
2. **§7 Closure-Notiz schreiben** (Zweig, vor dem `git mv`): je Regel ein Eintrag mit `liegt in <Zielort>` nach Abschnitt B (Pflicht: `.claude/commands/implement-slice.md §Schritt 17`, Skill; Vorlage 1: Skill; Vorlage 2: Skill und `verifier.md`; E1: `AGENTS.md §3.15`; E2: die drei Dateien; E4: `AGENTS.md §3.15` und `AGENTS.md §3.10`); Rückgaben 2 und 3 ohne Anker (Werkzeugliste, Zeiger *Übergangsregeln Nr. 8*), Form des Zielorts `<Pfad> §<Nr>` mit Anker in der Überschrift. Enthält: Messung vor und nach (Tabelle von Abschnitt D), Diff 156 Zeilen (Schwelle 400), `make groessenmarke`: Hinweis *kein Block* (`SPEC-128` Nr. 10 d), den `make gates`-Lauf (V-3), Verweis auf Review und Verifikation **mit Kennung ohne Pfad und ohne Endung** (Regel E2 gilt ab jetzt; sonst sperrt `make archive-welle` mit `[haenger]`).
3. **DoD-Haken in §2 setzen** und **jedes Risiko in §6 mit Ausgang** versehen: 1 (Kürzung) und 2 (Pflicht an einem Ort: Wortlaut steht nur in Schritt 17) und 3 (Lockerung: Summe sinkt) lassen sich jetzt auf *entfallen* setzen, der Rückführungsschwelle (156 gemessen, unter 400) ebenso; offen bleiben mit Adresse: Stempel (Integrationsschritt), E4-Wortlaut (Einspruch beim User), Zeiger Nr. 8 (N6, Planner beim Umbau der Übergangsregeln).
4. **Reports ablegen:** Review und diese Verifikation unter `docs/reviews/` als `2026-11-20-slice-regeln-verkoerpern-review.md` und `2026-11-20-slice-regeln-verkoerpern-verifikation.md` (beide liegen bisher im Scratchpad; DoD verlangt den Report im Repo). Vor dem Commit `make docs-check` (ADR-Kennungen als Link).
5. **Register (Integrationsschritt, Nr. 4 der Übergangsregeln: nur dort, nicht im Slice-Zweig):** die sechs `state.md` auf *verkörpert* mit Zielort setzen (`zweig-ohne-lauf-unter-dem-detektor`, `report-der-rolle-faellt-docs-check`, `menge-ohne-ihre-positionen`, `verweis-auf-review-report-sperrt-die-archivierung`, `ein-negativ-benennt-die-ungefangene-klasse-zu-eng`, `laufzeitmessung-widerspricht-sich-unter-last`); in `report-der-rolle-faellt-docs-check` steht noch *Anker `seit welle-betrieb`*, verkörpert ist `seit welle-mock` (angleichen). Ein neuer Beleg `evidence/slice-regeln-verkoerpern.md` für `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (V-1: das Maß eines Negativs zu klein, fünftes Auftreten); V-2 als offen mit Adresse (Architect oder User) oder als Beleg einer eigenen Beobachtung führen. Roadmap-Zeile aus der Arbeit auf den Beleg ziehen. *Keine Beobachtung angefallen* trifft nicht zu.
6. **Move und Integration:** erst der Inhalt (DoD-Häkchen, Closure-Notiz), dann der reine `git mv` nach `done/` (`AGENTS.md` §3.3 Nr. 2, `make slice-mv`); Merge per `git merge --no-ff slice-regeln-verkoerpern` in `main`; danach `make gates` im Hauptbaum unter der Sperre (§5 Closure-Trigger), `make gegenproben` und den Stempel allein committen (Nr. 2); in `README.md` der Planungs-Dokumente nichts ändern (Nr. 4 ist für diesen Slice bestätigt, §6).
7. **Folge an den User/Architect (V-2):** ob der Prüfer im Skill (Zeile *Akzeptiertes Negativ ohne Grund und Folge*) die Klasse und das Maß mitprüfen soll; ohne diese Zeile hat E4 keinen Konsumenten (Grundsatz *Jedes Artefakt hat einen Konsumenten*, Baseline-Regelwerk `grundlagen-harness-dateien.md`), und die Aussage steht an drei Stellen (§3.9, §3.17).

## Evidenz und Pfade

Probe-Kopien und Skripte: `/tmp/claude-1000/-Development-odata-recorder/efb0f904-f3aa-4d9d-a147-ca0b12313d6a/scratchpad/regeln-verifier/` (`base/` = `git archive HEAD`, `dc.sh` = `d-check` auf einer Kopie, `gates.log` = Lauf von `make gates`).
