# ADR-0006: SQLite-Bibliothek ohne native Abhängigkeit

**Status:** Accepted

**Datum:** 2026-10-19

**Autor:** pt9912

**Bezug:** [ADR-0001](0001-implementierungssprache-go.md), [ADR-0002](0002-architektur-gate-a-check.md), [ADR-0003](0003-lint-profil-golangci.md)

**Schärft:** —

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Die Persistenz des Recorders liegt in SQLite (`ARC-009`, `ARC-014` der Architektur).
Das `Dockerfile` baut mit `CGO_ENABLED=0` und ohne C-Toolchain, und alle Läufe sind Docker-only ([ADR-0001](0001-implementierungssprache-go.md)).
`pgwire-recorder` hat dieselbe Frage entschieden (dort `modernc.org/sqlite`).
Das Profil lässt nur Bibliotheken mit Eintrag in `gomodguard_v2` zu und verlangt dafür eine ADR ([ADR-0003](0003-lint-profil-golangci.md)).

## Entscheidung

Wir wählen **`modernc.org/sqlite`**, eine SQLite-Bibliothek in reinem Go ohne cgo, in der Version `v1.60.1` mit den von ihr gepinnten Abhängigkeiten `modernc.org/libc`, `modernc.org/mathutil` und `modernc.org/memory`.
Sie darf nur im Driven Adapter `src/adapters/driven/sqlite` stehen; `.a-check.yml` bindet sie mit einer `tech`-Regel dorthin, und `gomodguard_v2` führt den Eintrag.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — `github.com/mattn/go-sqlite3` (cgo) | verbreitet, schnell | verlangt einen C-Compiler im Build-Image; widerspricht `CGO_ENABLED=0` und dem Cross-Build ohne Toolchain |
| B — `github.com/ncruces/go-sqlite3` (WebAssembly) | ohne cgo | zusätzliche Laufzeit für WebAssembly, weniger verbreitet |
| C — keine Datenbank, Dateien | keine Fremdbibliothek | die Transaktionen, Fremdschlüssel und Constraints des Schemas gingen verloren |
| **D — `modernc.org/sqlite` (gewählt)** | ein Build ohne C-Toolchain, passt zu Docker-only und zu einem Binary für mehrere Plattformen | nach eigener Angabe etwa 1,3- bis 2-mal langsamer als native SQLite; größeres Binary; die Version von `modernc.org/libc` ist exakt zu pinnen |

## Konsequenzen

- Positiv: Der Build bleibt netzlos außer der Stufe `deps` und braucht keine C-Laufzeit.
- Negativ: Die Geschwindigkeit bei großen Recordings ist zu messen; das Binary wächst.
- Folgepflicht: `go.mod` und `go.sum` pinnen die Version; die `tech`-Regel in `.a-check.yml` und der Eintrag in `gomodguard_v2` im Profil folgen im Slice, der den Treiber einbindet.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| a-check | `modernc.org/sqlite` nur im Driven Adapter `src/adapters/driven/sqlite` | `make a-check` |
| golangci-lint | `gomodguard_v2` lässt die Bibliothek zu | `make lint` |

## Re-Evaluierungs-Trigger

Wenn die Geschwindigkeit bei großen Recordings nicht genügt, die Bibliothek nicht mehr gepflegt wird oder ihre Lizenz sich ändert.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-19 | Accepted | Entscheidung des Owners im Gespräch: wie `pgwire-recorder` |
