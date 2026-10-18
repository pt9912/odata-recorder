# ADR-0004: Die Verträge der Harness-Werkzeuge stehen im Ordner harness

**Status:** Accepted

**Datum:** 2026-10-18

**Autor:** pt9912

**Bezug:** [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [`MR-001`](../../../harness/conventions.md#mr-001)

**Schärft:** —

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md`
§Ziel-Form: ADR (MADR).

> **IDs als Markdown-Link** (klickbar zur Quelle, Baseline-Regelwerk `grundlagen-referenz-richtung.md` §Referenz-Richtung (SDP)).
> Der `<anker>` ist der GitHub-Heading-Slug der Ziel-Überschrift. Der
> Der Referenz-Richtungs-Gate prüft heute nur Token-Richtung, **nicht** die
> Anker-Auflösung — ein umbenannter Abschnitt rottet den Link still. Die
> anker-validierende Reifestufe löst Links auf und prüft die Anker-Existenz
> am Zielknoten.

---

## Kontext

Gates, Prüfer und Hooks legen technisch fest, was sie prüfen und wie sie an ihren Randformen entscheiden.
Die Baseline nennt dafür die Spezifikation als Ort.
Die Produktspezifikation wächst so mit jedem Harness-Slice, und wer das Produkt liest, liest Werkzeugverträge mit.
Annahme: Der Änderungs-Prozess der Werkzeugverträge gleicht dem der Produktspezifikation (fortschreibbar ohne Change Request, das Lastenheft ist die Decke).

## Entscheidung

Wir legen die Verträge der Harness-Werkzeuge **in den Ordner `harness/vertraege/`**, eine Datei je Kennung, und halten die Abweichung von der Baseline als [`MR-001`](../../../harness/conventions.md#mr-001) fest.
`spec/spezifikation.md` legt nur das Produkt fest.
Rang 2 der Source Precedence trägt beide Orte.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — Werkzeugverträge in `spec/spezifikation.md` (Baseline) | ein Ort, Referenz-Richtung des Doku-Gates gilt | Produktspezifikation wächst mit jedem Harness-Slice; Leser des Produkts lesen Werkzeugverträge mit |
| B — Verträge in den Sensor-Dateien (`harness/sensors/`) | nah am Werkzeug | Lesart und Festlegung vermischen sich; keine eigene Kennung-Reihe |
| **C — `harness/vertraege/` (gewählt)** | Produkt und Werkzeug getrennt, Kennungen bleiben fortlaufend | Abweichung von der Baseline, Verweise der Spezifikation sind Text statt Link |

## Konsequenzen

- Positiv: Die Produktspezifikation führt nur das Produkt; ein Werkzeugvertrag ändert sich, ohne dass ein Spec-Review nötig ist.
- Negativ: Die Referenz-Richtung des Doku-Gates schützt den Ordner nicht; die Disziplin „Verträge erweitern das Lastenheft nie" bleibt Review-Sache.
- Folgepflicht: Jeder neue Werkzeugvertrag entsteht in `harness/vertraege/` und steht in dessen `README.md`.

## Fitness Function (falls maschinell prüfbar)

Keine; die Ordnung des Ordners ist nicht maschinell prüfbar.

## Re-Evaluierungs-Trigger

Wenn die Baseline einen Ort für Werkzeugverträge außerhalb der Produktspezifikation nennt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-18 | Accepted | Entscheidung des Owners |

Nach `Accepted` wird diese Datei **nicht mehr inhaltlich überschrieben**.
Spätere Korrekturen oder Schärfungen entstehen als neue ADR mit `Supersedes ADR-0004`.
