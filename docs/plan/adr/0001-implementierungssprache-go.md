# ADR-0001: Go als Implementierungssprache

**Status:** Accepted

**Datum:** 2026-10-18

**Autor:** pt9912

**Bezug:** [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [`LH-RB-03`](../../../spec/lastenheft.md#lh-rb-03--lokale-entwicklungsumgebung)

**Schärft:** —

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Der OData Recorder läuft als einzelner Prozess in einem Container mit einem Datenvolumen (`LH-QA-06`).
Er streamt HTTP-Verkehr, schreibt Recordings und hält Workspace-State in SQLite.
Das Schwesterprojekt `pgwire-recorder` verwendet dieselbe Architekturform (Hexagon, Driving/Driven Adapter) und dieselben Harness-Werkzeuge.
Die Spec-Straten sind sprachfrei; diese ADR legt die Sprache für den Code fest.
Wenn die Annahme kippt, dass Streaming, Container-Größe und gemeinsame Werkzeuge den Ausschlag geben, kippt die Entscheidung.

## Entscheidung

Wir wählen **Go**.
Der Quellbaum liegt unter `src/` (Wurzel-`go.mod`), die Paketstruktur folgt der Architektur-Sicht.
Gebaut, getestet und geprüft wird ausschließlich über `make` in Containern (`AGENTS.md` §3.1).

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — Kotlin/JVM | reiche Bibliotheken, `a-check` unterstützt es | größeres Container-Image, langsamerer Start, andere Toolchain als `pgwire-recorder` |
| B — Python | schnelle Entwicklung | schwächere Streaming- und Nebenläufigkeitsgarantien für den Proxy-Pfad, Laufzeit im Image |
| **C — Go (gewählt)** | statisches Binary, kleines Image, Streaming und Nebenläufigkeit in der Standardbibliothek, gleiche Werkzeuge wie `pgwire-recorder`, `a-check` unterstützt es | weniger Ausdrucksmittel für Domänenmodelle, OData-Bibliotheken sind rar |
| D — nichts festlegen | keine Bindung | `a-check` verlangt `languages`; ohne Sprache kein Build, kein Dockerfile |

## Konsequenzen

- Positiv: Ein Toolchain-Image, ein Binary, gemeinsame Make- und Docker-Muster mit `pgwire-recorder`.
- Negativ: Der OData-Parser ist Eigenbau hinter dem Port `ODataCodec`; eine Fremdbibliothek muss ihre eigene ADR bekommen.
- Folgepflicht: Dockerfile und Build-Gates im ersten Code-Slice; die Wahl der SQLite-Bibliothek ist eine eigene ADR.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `a-check` | `languages: go` in `.a-check.yml` | `make a-check` |

## Re-Evaluierungs-Trigger

Wenn kein tragfähiger OData-Codec in Go entsteht oder `pgwire-recorder` die Sprache wechselt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-18 | Accepted | Entscheidung des Owners im Gespräch zur Architektur-Gate-Einführung |
