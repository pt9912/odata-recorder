# ADR-Index — OData Recorder

> **Derivativ:** Quelle der Wahrheit sind die ADR-Dateien; dieser Index ist
> eine Bequemlichkeits-Sicht — bei jedem neuen/akzeptierten ADR mitziehen.

| ID | Titel | Status | Bezug |
|---|---|---|---|
| [ADR-0001](0001-implementierungssprache-go.md) | Go als Implementierungssprache | Accepted | `LH-QA-06`, `LH-RB-03` |
| [ADR-0002](0002-architektur-gate-a-check.md) | Architektur-Gate mit a-check | Accepted | `LH-QA-06` |
| [ADR-0003](0003-lint-profil-golangci.md) | Lint-Profil mit golangci-lint | Accepted | — |
| [ADR-0004](0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) | Verträge der Harness-Werkzeuge im Ordner harness | Accepted | `LH-QA-06` |
| [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md) | Gegenproben außerhalb der Gate-Kette, nur bei geänderter Eingabe | Accepted | — |
| [ADR-0006](0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) | SQLite-Bibliothek ohne native Abhängigkeit | Accepted | — |
| [ADR-0007](0007-race-detektor-im-gate.md) | Race-Detektor im Gate | Accepted | `LH-QA-03`, `LH-QA-06` |
| [ADR-0008](0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md) | Blob-Bereinigung im laufenden Betrieb mit Wächter und Zurückstellen | Accepted | `LH-RB-04`, `LH-FA-03`, `LH-QA-03` |

## Konventionen

- ADRs sind nach `Accepted` **immutable** (siehe Baseline-Regelwerk `modul-04-adrs.md`).
- Schärfungen entstehen als neue ADR mit `Supersedes ADR-NNNN`.
- Bei `Accepted`: diesen Index aktualisieren (Status, Datum).
- Jede ADR deklariert im `**Schärft:**`-Feld *aufwärts*, welche Spec-Stelle
  sie verbindlich macht; Prozess-ADRs ohne Spec-Stratum tragen `—`.
