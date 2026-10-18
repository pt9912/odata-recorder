# ADR-0002: Architektur-Gate mit a-check

**Status:** Accepted

**Datum:** 2026-10-18

**Autor:** pt9912

**Bezug:** [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [ADR-0001](0001-implementierungssprache-go.md)

**Schärft:** [`SPEC-039`](../../../spec/spezifikation.md#7-festlegungen-der-harness-werkzeuge)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Sieben Reviews der Spec-Straten haben die Befundklasse „Port-Katalog und Importkanten inkonsistent" sechsmal gefunden.
Die Importregeln der Architektur-Sicht müssen maschinell am Code gehalten werden, sobald es Code gibt.
`pgwire-recorder` prüft dieselbe Hexagon-Form mit `a-check` (netzlos, schreibgeschützt, per Digest gepinnt).

## Entscheidung

Wir wählen **`a-check`**, in der Konfiguration `.a-check.yml` und über `make a-check`.
Die Schichten und Kanten der Datei spiegeln den Importgraph aus `spec/architecture.md` §2; das Target hängt an `GATE_CHECKS`.
Das Image ist per Digest gepinnt; die Hebung des Pins ist ein bewusster Commit.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — nichts tun | kein Aufwand | die Befundklasse bleibt Handarbeit des Reviews |
| B — Go-eigene Linter (z. B. depguard) | ohne eigenes Image | zweite Regelquelle neben der Sicht, an die Sprache gebunden |
| **C — `a-check` (gewählt)** | eine deklarative Datei, derselbe Weg wie `pgwire-recorder`, netzlos und schreibgeschützt | Werkzeug eines Schwesterprojekts, das mitgepflegt werden muss |

## Konsequenzen

- Positiv: Importrichtungen sind ein Gate, kein Review-Thema.
- Negativ: Das Gate prüft, solange kein Code existiert, einen leeren Baum und ist dann nur ein Beleg der Konfiguration.
- Folgepflicht: Mit dem ersten Code-Slice eine Gegenprobe wie `a-check-negativ` in `pgwire-recorder` anlegen; Bibliotheks-Grenzen (`tech`) kommen mit der ADR der jeweiligen Bibliothek.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `a-check` | Kanten der Schichten laut `.a-check.yml`, keine Adapter-Imports im Kern, keine lateralen Adapter-Imports | `make a-check` |

## Re-Evaluierungs-Trigger

Wenn `a-check` die Sprache oder eine Regelform nicht mehr trägt, die die Architektur-Sicht verlangt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-18 | Accepted | Entscheidung des Owners im Gespräch |
