# Verifikations-Report: slice-gegenproben — 2026-10-19

**Rolle:** Verifier (Modul 11) · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

**Gegenstand:** `docs/plan/planning/in-progress/slice-gegenproben.md`, DoD-Punkte 1 bis 4 und Closure-Trigger aus §5; Plan-vs-Code-Diff `3117b17` (Move nach `in-progress/`) bis `646a369` (HEAD).

**Quellen:** [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md), [SPEC-101](../../harness/vertraege/SPEC-101-lint-profil.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Proposed), [LH-QA-06](../../spec/lastenheft.md), Review `2026-10-19-slice-gegenproben-review.md`.

## Urteil

**Nicht bestanden, solange V-1 offen ist.** DoD 1 bis 3 sind im Aufbau erfüllt (Fälle, Mutationen, Stempel, Budget, Sensors-Zeile), aber `make gegenproben` ist auf einem unveränderten, sauberen Baum **nicht reproduzierbar grün**: 5 von 8 erzwungenen Läufen endeten rot. Damit trägt der Closure-Trigger „`make gegenproben` ist grün" (§5) nicht. DoD 4 (`make gates`) ist erfüllt. Der Rest sind Mittel- und Niedrig-Befunde und die erwarteten offenen Closure-Punkte.

## Selbst gefahrene Sensoren

| Sensor | Ergebnis |
|---|---|
| `make gates` (Exit-Code von make selbst erfasst) | Exit 0, rund 8 s; `git status` danach sauber |
| `make docs-check` | 62 Dateien, 0 Befunde |
| `make gegenproben`, Stand unverändert | Exit 0, 0,044 s (Budget 1 s gehalten), „nichts läuft neu" |
| Änderung nur an der Roadmap (Doku) | Prüfsumme unverändert, kein Lauf |
| Änderung an `src/bootstrap/bootstrap.go` (Eingabe) | Lauf erzwungen, 14,8 s (Budget 30 s gehalten); Lauf rot (siehe V-1), Stempel unverändert, Exit 2 |
| Nach `git checkout` von Eingabe und Stempel | Prüfsumme stimmt wieder, „unverändert" |
| 8 erzwungene Läufe (Stempel auf Müll gesetzt) auf unverändertem Baum | **5 rot, 3 grün** (V-1) |
| Kalter Docker-Cache (`builder prune`) | **nicht gemessen** (hätte den Cache des Hosts verworfen); gemessen ist nur `--no-cache-filter` der Stufe bei warmem Cache |

Stempel und Arbeitsbaum sind am Ende wiederhergestellt (`git status` sauber).

## Mutationen (je Gegenprobe selbst gefahren, danach `git checkout`)

| Mutation | Erwartung | Ergebnis |
|---|---|---|
| `lint.sh`: `# Why:`-Prüfung aus | `lint/ausnahme-ohne-why`, `lint/why-falsch-eingerueckt` rot | beide rot, aus dem benannten Grund (zusätzlich `lint/basis` durch V-1) |
| `lint.sh`: Direktiven-Regex ohne Leerzeichen-Klasse | `lint/direktive-mit-leerzeichen` rot | rot („Ausgabe nennt "Direktive" nicht"), sonst nichts |
| `lint.sh`: Profil fehlt endet mit Exit 1 statt 2 | `lint/profil-fehlt` rot | rot („erwartet Exit 2, war 1") |
| `lint.sh`: Einrückung des `# Why:`-Blocks beliebig | `lint/why-falsch-eingerueckt` rot | rot („Verstoß nicht gemeldet") |
| `lint.sh`: immer Exit 0 | vier Verstoß-Fälle rot | vier rot |
| `lint.sh`: `sleep 250` am Anfang (Frist, Randform 6) | alle Lint-Fälle „Zeit überschritten" nach 120 s | alle sieben Lint-Fälle rot, Gesamtlauf 2 m 9 s (F-2 des Reviews behoben) |
| `.a-check.yml`: Rolle `domain` des Modells auf `app` | `a-check/kern-auf-driven-adapter` rot | rot („nennt "core-impurity" nicht") |
| `.a-check.yml`: Kante Driven Adapter zu Outbound Port entfernt | `a-check/erlaubter-weg-ueber-port` rot | rot („erlaubter Fall abgelehnt") |

Damit sind F-4, F-5 und F-6 von `slice-lint-stufe` (Review-F-1) nun durch Fälle gefangen. Zusage `lint/basis` und `a-check/basis` wurden nicht gesondert mutiert.

## Findings

| ID | Stufe | Kategorie | Befund | Beleg |
|---|---|---|---|---|
| V-1 | DoD 1 bis 3, §5 | HIGH | Die Lint-Fälle laufen nebenläufig und teilen den Cache von golangci-lint; golangci-lint bricht dann mit „parallel golangci-lint is running" ab (Exit 3, von `lint.sh` zu 1 gemacht). Auf unverändertem Baum wurden `lint/basis` und `lint/ausnahme-mit-why` (erlaubte Fälle) in 5 von 8 Läufen rot; der Stempel bleibt dann zu Recht stehen, aber der Lauf ist nicht grün. Der Review (13,4 s, grün) und die Angabe im Plan („gemessen 13,5 s") haben nur Glücksläufe gesehen. Folge für die Verstoß-Fälle: sie bleiben auch dann „ok", wenn golangci-lint gar nicht lief, weil die eigenen Prüfungen von `lint.sh` den erwarteten Text liefern; kein Fall belegt den golangci-lint-Teil von `lint.sh` auf der Verstoßseite. | `make gegenproben` achtmal mit erzwungenem Lauf; Logs mit „parallel golangci-lint is running" in `lint/basis`; `tools/gegenproben/main.go` · `sem := make(chan struct{}, runtime.NumCPU())` |
| V-2 | DoD 3, SPEC-102 Randform 10 | MEDIUM | `gegenproben-eingabe.sh` meldet eine fehlende Eingabedatei nicht: `tools/harness/lint.sh` verschoben, Skript endet mit Exit 0 und liefert eine andere Prüfsumme (`if [ -f "$f" ]` überspringt die Datei still). SPEC-102 sagt zu: „meldet den Fehler und endet ungleich 0". `make gegenproben` bricht erst im Docker-Build ab (Kontext ohne die Datei). Review-F-5 ist damit nur teilweise behoben. | `bash tools/harness/gegenproben-eingabe.sh a b` → Exit 0 |
| V-3 | Plan §4 | MEDIUM | Rückführungs-Marke „über 600 Zeilen" (Summe aus `git diff --numstat 3117b17..HEAD` über die Pfade aus §3, Tests eingeschlossen) liegt bei 723 (Review nannte 576 vor den Fixes). Der Plan nennt weder die Überschreitung noch eine Entscheidung dazu (`AGENTS.md` §3.14, §3.9). | `git diff --numstat` über `tools/gegenproben`, `gegenproben-eingabe.sh`, Stempel, `Dockerfile`, `.dockerignore`, SPEC-102, `gegenproben.mk`, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) samt Index, `harness/README.md` |
| V-4 | §5 | OFFEN (erwartet) | [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) steht auf `Proposed` (Index gleich); die Owner-Annahme steht aus, §5 „ADR `Accepted`" ist damit nicht erfüllt. Kein Fehler des Codes. Ebenso offen: alle DoD-Häkchen, die Closure-Notiz (§7 „Wird bei der Closure geschrieben") samt Lerneintrag, Register-Eintrag, Ausgänge aus §6 bei Closure. Der Verifikations-Report selbst ist der noch fehlende Beleg zu „Review durchgeführt" nur für das Review; ein Review nach den Fixes ist nicht dokumentiert. | `docs/plan/adr/README.md` · Zeile 0005 |
| V-5 | §6, Plan | LOW | §6 setzt „Das Zeitbudget hält nicht, weil `golangci-lint` je Fall seinen Cache neu füllt" auf *entfallen* mit „die Fälle teilen den Cache des Prozesses". Dieselbe Teilung ist die Ursache von V-1; der Ausgang trägt nicht. Der kalte Lauf ohne Docker-Cache ist nicht gemessen. | `slice-gegenproben.md` §6 |
| V-6 | Review-Reste | INFO | F-8 (`tools/gegenproben` außerhalb von `make lint`, vier Linter-Befunde, `//nolint`-Text im String von `faelle.go`) und F-10 (Prüfsumme folgt `.gitignore`, der Docker-Kontext nicht) sind bewusst offen und unverändert; SPEC-102 nennt F-8 unter „Grenze", F-10 steht nirgends. | `tools/harness/gegenproben-eingabe.sh` · `--exclude-standard` |

## Konformität je Stufe

- **DoD 1 (`a-check`-Gegenprobe):** Fälle 1 bis 4 laut SPEC-102 im Code (`faelle.go`), beide Mutationen rot aus dem richtigen Grund. Erfüllt, vorbehaltlich V-1 (a-check-Fälle selbst waren in keinem Lauf rot).
- **DoD 2 (Lint-Gegenprobe):** Fälle 5 bis 11, fünf Mutationen rot. Inhaltlich erfüllt; Zuverlässigkeit durch V-1 verletzt, Wirkung des golangci-lint-Anteils ungeprüft.
- **DoD 3 (Stempel, Stufe, ADR, Sensors-Zeile, Budget):** Stempel-Mechanismus verhält sich wie zugesagt (Doku ohne Lauf, Eingabe mit Lauf, roter Lauf ohne Stempelschreiben, `git checkout` stellt wieder her); `make gegenproben` steht nicht in `GATE_CHECKS` (`grep` über `harness/mk/*.mk`), Sensors-Zeile in `harness/README.md` mit Bindung an [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), ADR-Index gefüllt (einmal, F-4 behoben), Budget warm gehalten. Offen: V-2, kalter Lauf nicht gemessen, ADR Proposed.
- **DoD 4:** `make gates` grün, `make docs-check` ohne Befund.
- **SPEC-101 und [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md)/0003:** `lint.sh` und Profilkommentar sind mit SPEC-101 (Einrückung) im selben Commit nachgezogen; keine Gate-Lockerung (`AGENTS.md` §3.6), keine Direktive im Code, keine neue Ausnahme in `.golangci.yml`.
- **Plan-vs-Code:** Der Plan wurde mit den Korrekturen nachgezogen (§1, §2, §3, §4, §6; `slice-lint-stufe` und Roadmap angepasst); Abweichung nur V-3 und V-5. `go.sum`-Wiederholung (Review F-9) ist unverändert als INFO zu lesen.

## Pfad

`/Development/odata-recorder/docs/reviews/2026-10-19-slice-gegenproben-verifikation.md`
