# ADR-0013: Das Zeitbudget der Gegenproben ist ein Soll mit Messvorschrift; die Frist je Fall bleibt die einzige Schwelle

**Status:** Accepted

**Datum:** 2026-11-08

**Autor:** pt9912

**Bezug:** [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0009](0009-groessenmarke-als-gate-mit-messung-durch-den-host.md)

**Schärft:** [`SPEC-102`](../../../harness/vertraege/SPEC-102-gegenproben.md)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

`SPEC-102` nennt zwei Zeitgrenzen, und beide heißen 120 s.
Die **Frist je Fall** (`tools/gegenproben/main.go`, `fallFrist`) ist Code: ein Fall, der länger läuft, ist rot mit `Zeit überschritten`.
Das **Zeitbudget** von `time make gegenproben` bei geänderter Eingabe und kaltem Cache ist Prosa: „kein Gate und kein Hook prüft es“; liegt ein Lauf darüber, „misst der Implementer neu und trägt den Wert mit Grund ein“.

Das Budget wurde am 2026-10-22 auf einem Host mit 20 Kernen gesetzt (80,5 bis 93,8 s, Budget 28 % darüber).
Seither hält es nicht mehr: 143 s und 127 s (2026-10-30), 127,1 s und 126,0 s (2026-10-31), 137,3 s, 123,4 s und 132,5 s (2026-11-06), 169,3 s und 171,1 s (2026-11-08).
Jede dieser Messungen liegt bei einer Last von mehr als dem Zweieinhalbfachen der Kernzahl (Load Average 54 bis 68 am Ende), weil auf dem Host mehrere Slice-Zweige zugleich bauen; ein Lauf auf ruhigem Host steht seit 2026-10-30 aus.
Die Zusatzzeit der Fälle selbst ist klein (die jüngsten 16 Fälle brauchen Millisekunden), die Ursache des Sprungs von 137 auf 170 s ist offen, weil Last und Zusatzzeit nicht getrennt gemessen sind.
Die Frist je Fall ist ebenfalls unter Last gerissen (`test/basis` und `test/roter-test` bei 120,6 s, 2026-10-24), ohne dass ein Gate fehlte.

Folge heute: Jeder Implementer, der `make gegenproben` fährt, meldet das Budget als gerissen, trägt eine weitere Messung in `SPEC-102` ein, und die Welle-Closures führen die Frage als offen an den User (der Re-Evaluierungs-Trigger „über 30 s“ von [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md) ist eingetreten). Das Budget ist nicht wiederholbar messbar, weil seine Bedingung (Last) nicht Teil der Vorschrift ist.

## Entscheidung

Wir behalten das Gesamtbudget von 120 s als **Soll** und geben ihm eine **Messvorschrift**: Es gilt nur für einen Lauf unter definierter Last; ein Lauf außerhalb davon ist *ungültig*, kein Bruch des Soll, und wird wiederholt.
Die **Frist je Fall** bleibt unverändert die einzige maschinell geprüfte Schwelle; wir heben keine der beiden 120 s an.
Die Einzelheiten (Messgröße, Lastgrenze, Zahl der Wiederholungen, Ausgang bei Bruch) stehen in `SPEC-102`, nicht hier; sie werden erst nach `Accepted` dort eingetragen.
[ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md) bleibt gültig: Dies ist der Ausgang seines Re-Evaluierungs-Triggers, nicht seine Ablösung.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — Budget anheben (etwa auf 200 s) | endet die Meldung „gerissen“ sofort | Wert ohne Messbasis: er stammt aus Läufen unter fremder Last und würde die Last festschreiben, nicht den Aufwand der Fälle; das Budget wüchse mit jedem neuen Lauf mit; wo eine Zeitgrenze als Schwelle gilt, ist ihre Anhebung eine Lockerung (`AGENTS.md` §3.6) und nur per ADR zulässig |
| **B — Soll mit Messvorschrift unter definierter Last (gewählt)** | kein neuer Wert ohne Messung; trennt Last von Aufwand; kostet nur Prosa in `SPEC-102`; keine Schwelle sinkt, weil das Budget nie eine Schwelle war und die Frist je Fall bleibt | Auf einem Host mit Dauerlast gibt es keinen gültigen Lauf; die Vorschrift braucht dafür einen ruhigen Zeitpunkt (andere Läufe pausieren); das Soll bleibt ohne Prüfer |
| C — Fälle aufteilen, Stufen parallel bauen | könnte die Wanduhr senken, wenn die langen Fälle (1 bis 37) auf mehrere Stufen verteilt werden | Die Fälle laufen schon nebenläufig bis zur Kernzahl; auf einem ausgelasteten Host gewinnt eine Aufteilung nichts, die Export-Zeit des Images (17 bis 36 s) wächst mit jeder Stufe; eigener Slice, ohne Messung, dass die Last und nicht die Fälle die Zeit bestimmen |
| D — nur die Frist je Fall, Gesamtbudget streichen | eine Schwelle statt zwei; die geprüfte bleibt | Der Auftrag „der Lauf bleibt kurz“ (Re-Evaluierung von [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md)) hätte keinen Wert mehr; Streichen eines Budgets ohne Ersatz ist selbst eine Lockerung der Aussage |
| E — nichts tun | kein Aufwand | Jeder Lauf meldet weiter ein gerissenes Budget; die Frage bleibt in jeder Closure offen |

**Empfehlung: B.** B ist die einzige Option, die die Frage, ob das Budget gerissen ist, in eine beantwortbare verwandelt, ohne einen Wert zu erfinden: A rät, C und D verlassen die Frage. Die Eichung (eine Messung auf ruhigem Host) entscheidet, ob 120 s tragen; reicht sie nicht, folgt eine Anhebung als eigene ADR mit dieser Messung als Beleg, nicht mit Lastläufen.

## Konsequenzen

- Positiv: Ein Lauf bei Load Average über der Kernzahl ist ein Befund über den Host, nicht über die Gegenproben; die Closures führen die Frage nicht mehr als offen.
- Positiv: Die Frist je Fall, die einzige Schwelle, bleibt unberührt; kein Gate ist gelockert.
- Negativ: Das Soll hat weiter keinen Prüfer; ein Bruch unter gültiger Last bleibt eine Prosa-Meldung.
- Negativ: Die Frist je Fall kann unter Last weiter rot werden (`Zeit überschritten`); die Vorschrift behandelt das wie bisher mit einem einzeln wiederholten Lauf (`SPEC-102` Nr. 6).
- Folgepflicht: Nach `Accepted` übernimmt ein Slice den Entwurf in `SPEC-102` und führt die Eichung auf ruhigem Host aus; ein Planner legt ihn an. Ob das Programm die Fallzeiten ausgibt (ein Weg, Last und Aufwand zu trennen), ist Sache dieses Slice.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `tools/gegenproben` | Frist je Fall von 120 s, Text `Zeit überschritten` (besteht, unverändert) | `make gegenproben` |

Das Soll des Gesamtbudgets ist nicht maschinell geprüft und wird als solches in `SPEC-102` benannt.

## Re-Evaluierungs-Trigger

Drei gültige Läufe (nach der Messvorschrift) über 120 s: Anhebung des Soll als neue ADR mit den Messwerten.
Oder: die Frist je Fall reißt in einem gültigen Lauf (Last unter der Grenze): Befund an den Planner, kein Wiederholen bis grün.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-08 | Proposed | Architect, Auftrag zu den offenen Punkten aus [`welle-mock-results`](../planning/done/welle-mock-results.md) |
| 2026-11-08 | Accepted | Entscheidung des Users (Option B) |
