# ADR-0007: Race-Detektor im Gate

**Status:** Accepted

**Datum:** 2026-10-22

**Autor:** pt9912

**Bezug:** [`LH-QA-03`](../../../spec/lastenheft.md#lh-qa-03--isolation-und-konsistenz), [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [ADR-0001](0001-implementierungssprache-go.md), [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0006](0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)

**Schärft:** [`SPEC-122`](../../../harness/vertraege/SPEC-122-test-race.md)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Der Kern des Produkts ist nebenläufig: die Aufnahme der Bodies, der Record-Dienst, der Upstream-Adapter und die Uhr arbeiten in mehreren Goroutinen, und der HTTP-Adapter und weitere Dienste kommen dazu.
Der Race-Detektor von Go braucht cgo und einen C-Compiler; das `Dockerfile` baut mit `CGO_ENABLED=0` und ohne Compiler ([ADR-0006](0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)), und die Tests fuhren deshalb ohne Detektor.
Sie ordneten die Goroutinen mit Kanälen und wiederholten nebenläufige Tests; ein Datenrennen sahen sie nur, wenn es einen falschen Wert lieferte, und mehrere Fänger wurden nur mit einer Rate rot, nicht sicher.
Das Register der Beobachtungen führt diesen Befund über der Schwelle.
Annahmen: Ein Datenrennen ohne falschen Wert ist im Produkt ein Fehler (Panic nach gleichzeitigem Schreiben einer Map, ein Wert, der zur falschen Anfrage gehört), den nur ein Detektor findet; ein Go-Image mit Compiler lässt sich per Digest pinnen und baut ohne Netz.
Kippt eine der beiden, kippt die Entscheidung.

## Entscheidung

Wir fahren den Race-Detektor als **eigenes Gate `make test-race`**, in einer **eigenen Stufe des `Dockerfile`** auf einem per Digest gepinnten Go-Image, das denselben Go-Stand wie `deps` und einen C-Compiler mitbringt (`golang:1.27.1-trixie`), über alle Pakete, die auch `make test` fährt.
Das Ziel hängt an `GATE_CHECKS`; die Gegenprobe ist ein Fall von `make gegenproben` ([ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md) bleibt).
Das Binary und die Stufen bis `lint` bleiben bei `CGO_ENABLED=0` und ohne Compiler; die neue Stufe liefert ein Urteil, kein Artefakt.
Was das Ziel im Einzelnen prüft, wo es endet und wie es sich an den Randformen verhält, steht in `SPEC-122`.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — nichts ändern: Kanal-Ordnung und Wiederholung vor der Übergabe | kein cgo, kein neues Image | ein Datenrennen ohne falschen Wert bleibt unsichtbar; jeder Fänger wird nur mit einer Rate rot, die je Mutation zu messen ist; der Befund kehrt wieder |
| B — Wiederholungen im Gate (`-count=N` in `make test`) | kein cgo, kein neues Image | sieht dasselbe wie A, nur öfter; die Laufzeit wächst mit N bei jedem Lauf; ein Rennen ohne falschen Wert bleibt unsichtbar |
| C — Wirkungs-Fänger allein (Tests zählen die Wirkung, statt auf ein Rennen zu warten) | deterministisch, ohne cgo | fängt nur die Wirkung, die der Plan vorhersah; ein unbekanntes Rennen nicht; bleibt als Ergänzung, wo die Wirkung benennbar ist |
| D — `-race` außerhalb des Gates, nur auf Aufruf | `make gates` bleibt kurz | ein Gate, das niemand fährt, ist eine Behauptung ([ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md), Negativ); anders als bei den Gegenproben ändert sich die Eingabe des Detektors mit jeder Änderung am Quellstand, und ein unveränderter Stand kostet durch den Docker-Cache nichts |
| E — cgo und Compiler in allen Stufen, oder ein Compiler per Paketmanager in `deps` | ein Image | widerspricht [ADR-0006](0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md): das Binary entstünde mit cgo und wäre nicht mehr ohne Toolchain quer baubar; ein Paketmanager braucht Netz und liefert keinen per Digest gepinnten Stand |
| F — die Stufe auf dem Image von golangci-lint (Go und Compiler sind dort vorhanden) | kein neuer Pin | Go-Version und Compiler sind dort keine Zusage des Images; eine Anhebung des Linters verschöbe die Toolchain des Gates |
| **G — eigene Stufe auf einem Go-Image mit Compiler (gewählt)** | der Detektor meldet unabhängig vom Zufall der Reihenfolge; Binary und übrige Stufen bleiben unverändert; die Go-Version ist prüfbar gleich der von `deps` | ein weiteres gepinntes Image (rund 1,3 GB), längerer Lauf von `make gates`, mehr Speicher; ein Compiler in einer Stufe |

## Konsequenzen

- Positiv: Ein Datenrennen in einem Paket, dessen beide Zugriffe ein Test ausführt, macht das Gate rot, auch ohne falschen Wert. Ein Fänger, den der Detektor ersetzt, braucht keine Rate mehr.
- Negativ: Der Detektor sieht nur ausgeführte Zugriffe und kein Rennen ohne ungeordneten Zugriff (Prüfen und Handeln in getrennten Abschnitten); beides steht mit Grund und Folge in `SPEC-122`. Ein Rot, das aus Fremdcode kommt (etwa von `checkptr` in einer Bibliothek), ist ein Befund, keine Ausnahme im Gate. Die Fälle des Ziels verlängern `make gegenproben`.
- Folgepflicht: Vertrag `SPEC-122` und seine Fälle in `SPEC-102`; die Stufe im `Dockerfile` und das Ziel in `harness/mk/go-gates.mk`; die Zeile des Ziels in `harness/README.md`; `deps` und die neue Stufe werden gemeinsam angehoben, weil die Stufe die Go-Version gegen `deps` prüft. Der Satz in [ADR-0006](0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), das `Dockerfile` baue ohne Toolchain, beschreibt nach dieser ADR die Stufen bis `lint`; die Entscheidung von [ADR-0006](0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) (Bibliothek ohne cgo, Binary ohne Toolchain) bleibt unberührt. Die Hexagon-Regeln ([ADR-0002](0002-architektur-gate-a-check.md)) und das Lint-Profil ([ADR-0003](0003-lint-profil-golangci.md)) ändern sich nicht; das Ziel senkt keine Schwelle, es fügt eine Prüfung hinzu.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `go test -race` über `pruefung race` | kein Datenrennen in den Paketen, die `make test` fährt | `make test-race` |
| `tools/gegenproben` | ein Datenrennen im Wegwerfmodul macht das Ziel rot, ein sauberer Fall nicht | `make gegenproben` |

## Re-Evaluierungs-Trigger

Wenn ein Paket unter dem Detektor die Standardgrenze von `go test` (10 Minuten) erreicht, ohne dass ein Rennen vorliegt, wenn der Bau-Host eine Plattform ohne Detektor trägt oder wenn das Image keinen Compiler mehr mitbringt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-22 | Proposed | Entwurf des Architect; die Annahme steht beim Nutzer aus |
| 2026-10-22 | Accepted | Entscheidung des Owners im Gespräch |
