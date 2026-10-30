# ADR-0009: Die Größenmarke ist ein Gate, der Host misst, ein Programm im Container urteilt

**Status:** Proposed

**Datum:** 2026-10-30

**Autor:** pt9912

**Bezug:** [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [ADR-0001](0001-implementierungssprache-go.md), [ADR-0004](0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0007](0007-race-detektor-im-gate.md)

**Schärft:** [`SPEC-128`](../../../harness/vertraege/SPEC-128-groessenmarke.md)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Die Größenmarke eines Slice stand als Prosa-Regel im Briefing: eine Formel, die der Planner von Hand rechnete, und eine Zahl, die der Implementer von Hand maß.
Sechzehn Slices überschritten die Marke, keiner wurde zurückgeführt, und die Messung des dritten Punkts widerlegte einen Zuschlag der Formel.
Die Entscheidung des Nutzers lautet: ein Werkzeug, das rechnet und misst, statt einer Regel, die jemand beachtet.
Offen waren die Lage des Werkzeugs und der Zugang zur Historie, denn `git diff --numstat` braucht sie, und die Prüfläufe sind netzlos und ohne schreibbaren Mount.
Annahmen: Ob ein Slice in Arbeit ist, sagt das Verzeichnis `docs/plan/planning/in-progress/`; der Host hat `git` (die Host-Abhängigkeiten sind `bash`, `git`, `docker` und `make`); eine Überschreitung, die nach dem Review durch Korrekturen wächst, soll einen neuen Eintrag mit Grund brauchen.
Kippt die erste Annahme (die Pläne liegen nicht mehr dort), findet das Gate keinen Slice und prüft nichts.

## Entscheidung

Wir führen **`make groessenmarke` als Gate in `GATE_CHECKS`** ein.
Es prüft jeden Plan unter `in-progress/`, der einen Block mit den Eingaben der Marke trägt; ob vor dem ersten Code-Commit oder im Code geprüft wird, folgt aus der Messung, nicht aus einer Angabe im Plan.
**Der Host misst:** ein Treiber ruft `git` und gibt die Zahlen als Datei an ein Programm in Go, das in einem Container ohne Netz und ohne Mount läuft und allein urteilt.
Formel, Faktoren, Zustände, Meldungen und Randformen stehen in `SPEC-128`, nicht hier.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| **A — Gate in `GATE_CHECKS`, Zustand aus Plan und Messung, der Host misst (gewählt)** | die Regel gilt an jedem Stand und bei jedem `make gates`; eine Überschreitung ohne Grund ist rot, wo sie entsteht; keine Angabe im Plan, die jemand vergessen könnte; `git` bleibt auf dem Host, das Programm bleibt netzlos | das Gate liest Planungsdokumente und die Historie; es kann mitten in der Arbeit rot werden, sobald die Messung die Marke oder das letzte `N` eines Eintrags überschreitet |
| B — außerhalb der Kette, nur auf Aufruf, wie die Gegenproben ([ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md)) | `make gates` bleibt kurz, nie rot ohne Anlass | die Pflicht zum Fahren stünde in der Spezifikation und kein Gate erzwänge sie: dieselbe Lage wie die Prosa-Regel, gegen die sechzehn Überschreitungen standen; anders als bei den Gegenproben ändert sich die Eingabe mit jedem Commit, ein Stempel würde nichts sparen |
| C — geteilt: Marke gegen Formel in der Kette, Messung außerhalb | kein `git` im Gate | der Teil, der sechzehnmal fehlte, die Überschreitung, liefe außerhalb; die Prüfung Marke gegen Formel allein verhindert nichts, was gemessen wurde |
| D — Gate, `git` im Container (`.git` als Mount mit `:ro` oder per `docker cp`) | der Lauf ist ein Container | das Image der Stufe `deps` trägt kein `git`, ein Image mit `git` ist ein weiterer Pin; `git` verweigert ein Verzeichnis, das einem anderen Benutzer gehört (`safe.directory`); der Host braucht `git` ohnehin |
| E — Prosa-Regel bleibt, schärfer formuliert | keine Änderung am Gate | widerspricht der Entscheidung des Nutzers; Feedforward allein hielt die Regel sechzehnmal nicht |

## Konsequenzen

- Positiv: Die Formel steht an einer Stelle; Überschreitung, abweichende Marke und fehlender Grund sind rot, und der Eintrag mit Grund ist das einzige Mittel, sie zu quittieren.
- Negativ: Ein Gate liest Pläne und Historie. Rot wird es nur durch einen vollständigen Block (`SPEC-128` Nr. 8); ein Slice, dessen Planner keinen Block setzt, bleibt ungemessen, und Code vor der Marke bleibt ungesehen (`SPEC-128` Nr. 10 d). Ein Rebase, der die Basis aus dem Verlauf nimmt, macht das Gate rot, bis die Basis im Plan nachgezogen ist.
- Folgepflicht: Das Briefing zeigt auf das Werkzeug statt die Formel zu tragen, und die Rollen, die einen Slice übergeben oder prüfen, nennen den Lauf in ihrem Bericht; beides ändert allein der Nutzer.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `tools/groessenmarke` | die Marke jedes Plans mit Block ist die Formel, die Messung liegt darunter oder trägt einen Eintrag mit Grund | `make groessenmarke` |

## Re-Evaluierungs-Trigger

Wenn das Gate an einem Stand rot ist, an dem kein Fehler vorliegt (zweimal in einem Slice, etwa durch eine Änderung fremder Slices an denselben Pfaden), oder wenn der Lauf von `make groessenmarke` ohne Änderung der Eingaben länger als 10 s dauert.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-30 | Proposed | Entwurf des Architect; die Annahme steht beim Nutzer aus |
