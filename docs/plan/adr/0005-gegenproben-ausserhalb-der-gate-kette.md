# ADR-0005: Die Gegenproben verlassen die Gate-Kette und laufen nur bei geänderter Eingabe

**Status:** Accepted

**Datum:** 2026-10-19

**Autor:** pt9912

**Bezug:** [ADR-0002](0002-architektur-gate-a-check.md), [ADR-0003](0003-lint-profil-golangci.md)

**Schärft:** [`SPEC-102`](../../../harness/vertraege/SPEC-102-gegenproben.md)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Eine Gegenprobe weist an einem Gate nach, dass es einen Verstoß rot meldet; sie prüft das Gate, nicht das Produkt.
In Schwesterprojekten dauerte sie so lange, dass sie `make gates` bestimmte, obwohl sich an den Gates nichts geändert hatte.
Hier braucht der Lauf bei geänderter Eingabe 13,5 s und bei unveränderter 0,04 s.

## Entscheidung

Wir nehmen die Gegenproben aus `GATE_CHECKS` und fahren sie mit `make gegenproben`, und zwar nur, wenn sich ihre Eingaben geändert haben.
Eine Prüfsumme der Eingaben, versioniert als `harness/gegenproben.stempel`, entscheidet.
Wann sie gefahren wird, und die Einzelheiten, stehen in `SPEC-102`.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — in `GATE_CHECKS` | jeder Lauf von `make gates` belegt die Gegenproben | bei geänderter Eingabe kommen 13,5 s zu jedem Lauf von `make gates` hinzu, die Kette wächst mit jeder Gegenprobe |
| **B — außerhalb, mit Prüfsumme (gewählt)** | `make gates` bleibt kurz; ein unveränderter Stand kostet 0,04 s; der Stempel reist mit dem Commit | die Pflicht zum Fahren steht in der Spezifikation, kein Gate erzwingt sie; der Stempel belegt nur, was ein Lauf geschrieben hat |
| C — außerhalb, nur von Hand | einfachste Mechanik | ohne Stempel läuft jede Gegenprobe jedes Mal oder nie |

## Konsequenzen

- Positiv: Schnelle Kette; die Gegenprobe läuft, wenn und nur wenn ihre Eingabe sich änderte.
- Negativ: Ein Gate, das niemand fährt, ist eine Behauptung (`AGENTS.md` §3.6 bleibt unberührt, weil keine Schwelle sinkt, aber der Beleg entfällt aus der Kette).
- Folgepflicht: Der Arbeitscontainer für Entwickler fährt `make gegenproben`; bis dahin gilt die Pflicht zum Fahren aus `SPEC-102`.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `tools/gegenproben` | die Fälle für `a-check` und `lint.sh` | `make gegenproben` |

## Re-Evaluierungs-Trigger

Wenn der Lauf bei geänderter Eingabe 30 s überschreitet oder ein Fall hinzukommt, dessen Eingaben jede Änderung berühren.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-19 | Accepted | Entscheidung des Owners im Gespräch |
