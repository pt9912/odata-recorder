# Welle welle-persistenz — Persistenz hinter den Outbound Ports — Closure-Notiz

> **Zitier-Form** *(bleibt stehen — Norm, kein Ausfüll-Hinweis).* Dieses
> Artefakt friert ein; was es zitiert, bewegt sich weiter. Deshalb: **Kennung,
> nicht Adresse** — `slice-<Kennung>` statt seines Lifecycle-Pfads, `make <target>`
> statt eines Links auf die Sensor-Datei, eine Baseline-Stelle als
> `v<X.Y.Z>` · `regelwerk/<datei>.md` §<Abschnitt> statt als Link
> (Baseline-Regelwerk `grundlagen-harness-dateien.md` §harness/README.md als
> Einstiegspunkt). Adopierter Tag: `v6.18.1`.

**Welle:** welle-persistenz
**Abschluss:** 2026-10-20
**Verantwortlich:** pt9912

## Was wurde geliefert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — *was gelernt wurde*: geliefert · was
funktionierte · was anders lief. Mit ID-Bezug, wo es einen gibt.

- `slice-sqlite-adapter-migration`: SQLite-Adapter mit eingebetteter Migration aus dem Schema, `make migration-check` ([ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), `LH-FA-03`).
- `slice-session-recording-persistenz`: Modell, Outbound Ports und SQLite-Repositories für Sessions und Exchanges, Vertragstests in `outboundtest` (`LH-FA-02`, `LH-FA-03`).
- `slice-blob-ablage`: content-adressierte Blob-Ablage mit Capture-Abschluss (`LH-FA-02`, `LH-FA-14`).
- `slice-baseline-persistenz`: Service-Metadaten, Baseline-Snapshots, Import und Coverage (`LH-FA-10`, `LH-FA-11`, `LH-FA-18`).
- `slice-workspace-state-persistenz`: Workspace, Entity, Cursor, Workspace-Snapshots, Wartungs-Port, `WithWorkspaceTransaction` und der Rundlauftest (`LH-FA-04`, `LH-FA-07`, `LH-FA-09`, `LH-FA-10`).
- Verträge `SPEC-105` bis `SPEC-108` in `spec/spezifikation.md`, [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md).

## Was hat funktioniert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3.

- Ein Vertragstest-Paket (`outboundtest`) gegen die Ports, an das der Adapter nur Hooks reicht: Zusagen sind am Port formuliert, nicht am SQL.
- Je Slice Implementer, Reviewer und Verifier in frischem Kontext mit eigener Mutationsrunde; zusammen mehrere hundert Mutationen je Slice, jede Zusage mit rotem Test belegt.
- Der Halt des Implementers nach `AGENTS.md` §3.12 (leere Kennung je Operation, `ReplaceAll` auf unbekanntem Workspace) hat zwei Randformen vor dem Code entscheiden lassen.
- Der Rundlauftest `TestRundlauf` beobachtet, was kein Slice-DoD beobachtet: alle Repositories in einer Datei, Schließen, Wiederöffnen, `foreign_key_check` und `integrity_check` leer.

## Was ging anders als geplant?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — jede Zeile möglichst mit der Konsequenz,
die daraus schon gezogen wurde (Folge-Slice, Spec-Version).

- Randformen blieben in jedem Slice teilweise offen, trotz `AGENTS.md` §3.12; das Review entschied sie nachträglich. Konsequenz: `AGENTS.md` §3.16 (Raster statt Liste).
- Die Marken der beiden letzten Slices wurden weit überschritten (1567 gegen 700, 3278 gegen 900 Zeilen, Summe aus `git diff --numstat`); beide stehen im Plan §4, die Slices blieben ungeteilt.
- Eine Fänger-Zeile des Plans (`ck_snapshot_scope` über den Port) war nicht realisierbar und wurde durch einen Vertragstest mit Hook ersetzt.
- Fachlogik, die die Ports voraussetzt, hat keinen Slice-Nehmer; sie steht in der Roadmap-Zeile von `welle-record-replay` als „noch ohne Slice".

## Steering-Loop-Einträge

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 (hier stehen **nur** Beobachtungen, die im
Register 3× erreicht haben; jeder Eintrag nennt seine `BEO-<KUERZEL>/<slug>`) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

- **Hard Rule** ergänzt: Randformen stehen im Plan als Raster Operation × Argumentart, jede Zelle mit Ort im Vertrag oder begründetem „gilt nicht".
  — liegt in `AGENTS.md §3.16`.
  Auslöser: `BEO-REPO/randform-erst-im-review-entschieden` (slice-go-skelett-build-test, slice-lint-gegenproben-modul, slice-sqlite-adapter-migration, slice-session-recording-persistenz, slice-blob-ablage, slice-baseline-persistenz, slice-workspace-state-persistenz — 7×).

Unter der Schwelle bleibt `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (2×).

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

- Keine Slice-Datei: Die Folgewelle `welle-record-replay` steht in der Roadmap und trägt die Übergaben (Abbildung von `ErrBeschaedigt` und `ErrBlob*` auf `ODR-E3001`, CRUD-Fachlogik, `MaintenanceUseCase` mit Blob-Abgleich, Aufbewahrung nach Alter, Durchsetzung von `LH-FA-11`).
- Offen für M1 (Entscheidung steht aus): `slice-gegenproben-build-test`.

## Verifikation

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 1 — keine Behauptung ohne nachprüfbaren
Anker (Hash, Lauf, Zahl).

- `make gates` grün auf `7024d7b` (Exit 0), darin `migration-check`.
- `make gegenproben` grün: 24 Fälle, 0 rot; Prüfsumme `9cfca5ea2483c77a`… im Stempel `harness/gegenproben.stempel`.
- Rundlauftest `TestRundlauf` Teil von `make test`, grün.
- Trigger-Audit: Carveouts 0 offen; Re-Evaluierungs-Trigger von [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) nicht eingetreten.
