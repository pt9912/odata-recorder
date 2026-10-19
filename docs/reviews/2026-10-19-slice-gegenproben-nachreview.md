# Review-Report: slice-gegenproben (Nachreview) — 2026-10-19

**Review-Art:** Code (gegen Plan + Konventionen), Nachreview der Fixes

**Gegenstand:** `slice-gegenproben`, Diff `646a369^..HEAD` (Fixes aus dem ersten Review und der Verifikation, darunter V-1).

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

**Eingangs-Kontext:**

- Slice-Plan `slice-gegenproben` (§1, §3, §4, §6)
- [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Proposed; kein Befund)
- `harness/vertraege/SPEC-102-gegenproben.md` (Randformen 1 bis 11), `SPEC-101-lint-profil.md`
- [LH-QA-06](../../spec/lastenheft.md)
- `AGENTS.md` (§3.7, §3.9 bis §3.14); vorheriger Report `2026-10-19-slice-gegenproben-review.md` und die Verifikation

## Läufe und Mutationen

- `make gegenproben` fünfmal erzwungen (`rm -f harness/gegenproben.stempel` vor jedem Lauf): 5 von 5 grün, 11 Fälle, 0 rot, 10,7 bis 12,0 s. Keine Flakiness beobachtet. Danach `git checkout harness/gegenproben.stempel`, Arbeitsbaum sauber.
- Mutation Randform 11 (TMPDIR-Umgebung je Fall entfernt), dreimal gefahren: jedes Mal rot über `make gegenproben` (`parallel golangci-lint is running`, `lint/basis`, `lint/ausnahme-mit-why`, `lint/direktive-nolint` rot). Das Unit-Test-Paket ist dabei grün geblieben (Befund N-1).
- Randform 10 von Hand: `.a-check.yml` entfernt, dann `make gegenproben`: Meldung `Eingabe .a-check.yml fehlt`, Make bricht ab, Stempel unberührt. Außerhalb eines Git-Baums endet das Skript ungleich 0.
- Mutationen der übrigen neuen Zusagen (Prozessgruppen-Kill, mehrere `WantText`, Einrückung des `# Why:`-Blocks, `Entferne`) sind durch benannte Tests oder Fälle rot zu sehen; die Zuordnung ist am Code gelesen, nicht je einzeln gefahren.

Vollständigkeit der Fixes: V-1 (eigenes TMPDIR), Want-Texte als Liste, die drei neuen Lint-Fälle, Prozessgruppen-Kill, Prüfsumme über Eingabepfade und `go.sum` sind im Diff umgesetzt; Randformen 1 bis 9 tragen Tests (`TestRun*`, `TestBewerte`, `TestStarte...`).

## Findings

| kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|
| MEDIUM | [LH-QA-06](../../spec/lastenheft.md), `AGENTS.md` §3.10 | `tools/gegenproben/main.go` · `[]string{"TMPDIR=" + tmps[i]}` | Randform 11 (eigenes TMPDIR je Fall, die Korrektur von V-1) hat keinen Unit-Test; entfernt man die Umgebung, bleibt `go test` grün und nur der Docker-Lauf wird rot. Auf einem Host mit einem Prozessorkern (`runtime.NumCPU()` = 1, Semaphor 1) laufen die Fälle nicht nebenläufig, dort bleibt die Mutation auch im Docker-Lauf grün. | ja | Zusage eines neuen Vertrags ohne Mutationsfänger im Test |
| MEDIUM | [LH-QA-06](../../spec/lastenheft.md), `AGENTS.md` §3.10 | `harness/vertraege/SPEC-102-gegenproben.md` · Randform 10; `tools/gegenproben/main.go` · `p == "go.sum"` | Randform 10 (Skript meldet fehlende Eingabe und bricht ab; `go.sum` ausgenommen) und die `go.sum`-Ausnahme in `kopiere` und `hashEingaben` haben keinen Test; nur der Handlauf oben belegt sie. Entfernt man die Ausnahme, schlägt der Lauf ohne `go.sum` (heutiger Stand) erst an der Stelle fehl, an der es heute auffiele, aber kein Test benennt den Grund. | ja | Zusage eines neuen Vertrags ohne Mutationsfänger im Test |
| LOW | `AGENTS.md` §3.9 | `docs/plan/planning/done/slice-gegenproben.md` · `lässt .a-check.yml und tools/gegenproben in den Kontext` | §3 nennt für `.dockerignore` weiter nur zwei Pfade (jetzt auch `go.sum`) und führt die geänderten Dateien `tools/harness/lint.sh`, `.golangci.yml` (Kommentar) und `SPEC-101-lint-profil.md` nicht; die Korrektur hat §3 nicht mitgezogen. | nein | Slice-Plan §3 folgt der Korrektur nicht |
| INFO | `AGENTS.md` §3.14 | `docs/plan/planning/done/slice-gegenproben.md` · `Marke überschritten` | Der Slice trägt die Marke 600 Zeilen und vermerkt ihre Überschreitung selbst; die Entscheidung, ungeteilt zu bleiben, trifft hier der Implementer. Ein grober `git diff --numstat 3117b17..HEAD` (alle Pfade, auch Reports) liefert 913. Zuständig: Planner. | ja | Rückführungs-Marke vom Implementer selbst freigegeben |
| INFO | Maintainability | `tools/gegenproben/main_test.go` · `mitBefehl` | Tests tauschen die Paketvariablen `befehl` und `programme` aus; solange keine Tests parallel laufen (`t.Parallel`), ist das sicher. Kein Aufruf nötig, nur Kopplung für spätere Änderer. | nein | Paketvariable als Test-Naht |

## Kommentare gegen Tests (§3.7, §3.11)

Kommentare an `starte` (Prozessgruppe), `eingaben` (`go.sum`) und `lint.sh` (sechs Leerzeichen) sagen nur zu, was ein Test oder Fall fängt (`TestStarteBeendetDieProzessgruppeBeiFristablauf`, Fall `lint/why-falsch-eingerueckt`); die `go.sum`-Aussage fängt kein Test (siehe zweites Finding). Der Kommentar an der TMPDIR-Zeile ist indikativ; sein Fänger liegt nur im Docker-Lauf (erstes Finding).

## Negativbefund

- `tools/gegenproben/`: geprüft, Befunde oben, sonst ohne Befund (Aufräumen der `.tmp`-Verzeichnisse liegt unter `basis`, Frist-Auswertung über `ctx.Err()` korrekt).
- `tools/harness/` (`lint.sh`, `gegenproben-eingabe.sh`): geprüft, ohne Befund über die genannten hinaus.
- `harness/mk/`, `harness/README.md`, `harness/vertraege/`: geprüft, ohne Befund (elf Fälle konsistent in SPEC-102, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), README).
- `docs/plan/adr/` (Index-Doppelzeile entfernt): geprüft, ohne Befund.
- `Dockerfile`, `.dockerignore`, `.golangci.yml`: geprüft, ohne Befund.
