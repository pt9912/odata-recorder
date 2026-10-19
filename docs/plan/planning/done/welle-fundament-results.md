# Welle welle-fundament — Code-Gates auf einem Go-Skelett — Closure-Notiz

> **Zitier-Form** *(bleibt stehen — Norm, kein Ausfüll-Hinweis).* Dieses
> Artefakt friert ein; was es zitiert, bewegt sich weiter. Deshalb: **Kennung,
> nicht Adresse** — `slice-<Kennung>` statt seines Lifecycle-Pfads, `make <target>`
> statt eines Links auf die Sensor-Datei, eine Baseline-Stelle als
> `v<X.Y.Z>` · `regelwerk/<datei>.md` §<Abschnitt> statt als Link
> (Baseline-Regelwerk `grundlagen-harness-dateien.md` §harness/README.md als
> Einstiegspunkt). Adopierter Tag: `v6.18.1`.

**Welle:** welle-fundament
**Abschluss:** 2026-10-19
**Verantwortlich:** pt9912

## Was wurde geliefert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — *was gelernt wurde*: geliefert · was
funktionierte · was anders lief. Mit ID-Bezug, wo es einen gibt.

- `slice-go-skelett-build-test`: Go-Modul, Composition Root mit Version, `make build` und `make test` über das `Dockerfile` ([ADR-0001](../../adr/0001-implementierungssprache-go.md), `LH-QA-06`).
- `slice-lint-stufe`: `make lint` mit golangci-lint und den eigenen Prüfungen von `lint.sh` ([ADR-0003](../../adr/0003-lint-profil-golangci.md)).
- `slice-gegenproben`: `make gegenproben` fährt die Gegenproben zu `a-check` und `lint.sh` in einem Go-Programm, nur bei geänderter Eingabe ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)).
- `slice-lint-gegenproben-modul`: `make lint` prüft auch das Modul `tools/gegenproben`.
- Verträge `SPEC-036`, `SPEC-101` und `SPEC-102` unter `harness/vertraege/`, [ADR-0001](../../adr/0001-implementierungssprache-go.md) bis [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md).

## Was hat funktioniert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3.

- Randformen vor dem Code in den Verträgen (ab dem zweiten Slice) und je Zusage eine selbst gefahrene Mutation; Review und Verifikation in frischem Kontext fanden danach nur Lücken am Rand, im dritten Slice ein HIGH (nebenläufige Sperre), das erzwungene Wiederholungen aufdeckten.
- Ein Go-Programm statt Shell-Skripten für die Gegenproben: 13,5 s bei geänderter, 0,04 s bei unveränderter Eingabe, gesteuert von einer Prüfsumme.
- Der Nutzer-Hinweis, die Gegenproben müssten schnell sein und nur bei Änderung laufen, hat die Form des Slice bestimmt, bevor Code entstand.

## Was ging anders als geplant?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — jede Zeile möglichst mit der Konsequenz,
die daraus schon gezogen wurde.

- Die Randformen des ersten Slice standen erst nach dem Code im Vertrag; seitdem verlangt `AGENTS.md` §3.12 sie vorher.
- Zusagen ohne fangende Prüfung traten in allen vier Slices auf (Version und `-trimpath`, leere Regelliste, drei Lint-Fälle, Modul-Schleife); die Konsequenz ist die Regel unter *Steering-Loop-Einträge*.
- Ein vierter Slice (`slice-lint-gegenproben-modul`) kam hinzu, weil das Programm der Gegenproben außerhalb des Lint-Gates lag.
- Die Rückführungs-Marke von `slice-gegenproben` wurde überschritten und im Plan benannt.

## Steering-Loop-Einträge

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 (hier stehen **nur** Beobachtungen, die im
Register 3× erreicht haben; jeder Eintrag nennt seine `BEO-<KUERZEL>/<slug>`) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

- **Hard Rule** ergänzt: Zusage und Fänger stehen als Tabelle im Plan, der Reviewer löst jede Zeile durch eine Mutation auf.
  — liegt in `AGENTS.md §3.15`.
  Auslöser: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (slice-go-skelett-build-test, slice-lint-stufe, slice-gegenproben, slice-lint-gegenproben-modul — 4×).

Unter der Schwelle bleiben `BEO-REPO/randform-erst-im-review-entschieden` (2×) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (1×).

## Beobachtungs-Register (Zeiger)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register — der Zähler wird **nicht** hier gepflegt; diese
Sektion ist ein Zeiger und trägt keine Daten.

Der Zähler steht in [`../observations/`](../observations/).
Was in dieser Welle **3×** erreicht hat, steht oben unter
*Steering-Loop-Einträge*.

## Folge-Slices

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — **derivativ**.

- Keine Slice-Datei: Die Folgewellen stehen in der Roadmap (`welle-dev-container`, `welle-persistenz`, `welle-record-replay`).
- Offen für M1 (Entscheidung steht aus): Gegenproben zu `make build` und `make test` fehlen; die Mutationen der Slices belegen sie nur einmalig.

## Verifikation

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 1 — keine Behauptung ohne nachprüfbaren
Anker (Hash, Lauf, Zahl).

- `make gates` grün auf dem Stand `9a2660d` (Exit 0): `baseline-verify`, `docs-check` (73 Dateien, 0 Befunde), `schema-check`, `a-check` (0 Befunde), `lint-profil`, `lint` (beide Module, 0 issues), `build`, `test`.
- `make gegenproben` grün: 16 Fälle, 0 rot; Prüfsumme der Eingaben `466a62674a09e2f7af0e9d65`… im Stempel `harness/gegenproben.stempel`.
- Zeitbudget der Gegenproben: 11,4 bis 11,7 s bei erzwungenem Lauf, 0,04 s bei unveränderter Eingabe (Budget 30 s und 1 s).
- Trigger-Audit: Carveouts 0 offen (das Verzeichnis `docs/plan/carveouts/` trägt keinen); bootstrap-aware Gates 0 (keine Reifestufe); Re-Evaluierungs-Trigger von [ADR-0001](../../adr/0001-implementierungssprache-go.md) bis [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) bestätigt, keiner eingetreten.
- Archivierung: `make archive-welle WELLE=welle-fundament` lief; `done/welle-fundament/archiv.zip` trägt die vier Slices, den Welle-Plan und neun Review-Reports, Stubs liegen daneben.
