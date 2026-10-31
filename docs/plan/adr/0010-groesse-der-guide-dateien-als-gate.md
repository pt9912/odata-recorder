# ADR-0010: Die Größe der Guide-Dateien ist ein Gate, die Grenze steht auf dem Ist-Stand und wandert nur nach unten

**Status:** Accepted

**Datum:** 2026-10-31

**Autor:** pt9912

**Bezug:** [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [ADR-0003](0003-lint-profil-golangci.md), [ADR-0004](0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0009](0009-groessenmarke-als-gate-mit-messung-durch-den-host.md)

**Schärft:** [`SPEC-040`](../../../harness/vertraege/SPEC-040-docs-check.md)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Das Briefing `AGENTS.md` ist auf 24648 Bytes und 337 Zeilen gewachsen, davon über die Hälfte Hard Rules; der Retirement-Check vom 2026-10-25 findet kein Gate, das die Größe einer Guide-Datei misst, und die Prosa-Regeln haben die Beobachtungen, aus denen sie entstanden, nicht verhindert.
Das Modul `file` von d-check (Zeilen- und Byte-Obergrenze einer ganzen Datei) und die Bedingung `max-lines` von `structure` liegen am Pin des Doku-Gates vor und sind nicht aktiviert.
Offen waren der Träger, die Höhe der Grenze und die Regel, nach der sie wandert, denn `AGENTS.md` §3.6 verlangt für jede Lockerung eine ADR, und die Verschlankung der Guide-Dateien verschiebt Text von einer Datei in eine andere.
Annahmen: Die Guide-Dateien sind die Dateien, die ein Lauf als Anweisung lädt; Bytes genügen als Einheit (`AGENTS.md`: 24648 Bytes, 24193 Zeichen); die Verschlankung ist ein Vorgang des Nutzers.
Kippt die zweite Annahme (Zeichen oder Token würden gebraucht), misst das Modul das Falsche.

## Entscheidung

Wir führen die **Größe der Guide-Dateien als Teil von `make docs-check`** ein: die Module `file` und `structure` in `.d-check.yml`, kein neues Ziel, kein Skript.
**Die Grenze steht auf dem Ist-Stand** (Zeilen und Bytes), nicht auf einer Richtgröße, und sinkt mit der Verschlankung.
**Senken ist frei; Anheben nur als Verschiebung** (derselbe Commit senkt die Summe der Grenzen desselben Maßes nicht höher) **oder durch eine eigene ADR**, die diese nennt.
Die Fälle, Meldungen und Randformen stehen in `SPEC-040`, nicht hier.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| **A: Module `file` und `structure` in `docs-check`, Grenze gleich Ist, Anheben nur als Verschiebung oder ADR (gewählt)** | kein neuer Träger; erreicht die Dateien trotz `scan.ignore` (belegt); die Verschlankung bleibt möglich, ohne dass jede Verschiebung eine ADR ist; Gegenprobe ohne Docker auf dem Host | Handarbeit bei der Messung; Zeilen sind ein grobes Maß für die Hard Rules; wer `AGENTS.md` verlängern will, muss Text verschieben |
| B: eigenes Programm und Ziel nach dem Muster von `zeichen-check` | könnte Zeichen oder Bytes je Abschnitt zählen | Stufe, Treiber, Vertrag und Gegenproben für dieselbe Zählung; kein Bedarf an der anderen Einheit |
| C: Prosa und Review wie bisher | keine Änderung | der Anlass: kein Gate misst die Größe, die Regel wirkt nicht |
| D: Grenze mit Spielraum (Ist plus Anteil) | weniger Rot bei kleinen Erweiterungen | der Spielraum ist der Wildwuchs, den das Gate begrenzen soll; jeder Anteil ist eine unbelegte Zahl |
| E: Grenze je Datei statt je Gruppe | friert jede Datei exakt ein | die Verschlankung verschiebt Text in kleine Dateien der Gruppe und brauchte dafür je eine Anhebung |
| F: jede Anhebung eine ADR | strikt nach `AGENTS.md` §3.6 | die Verschlankung wird zu einem Stapel von ADRs; ein Gate, das niemand anhebt, wird umgangen |
| G: jede Anhebung ein sichtbarer Commit mit Begründung | wenig Aufwand | eine Lockerung ohne ADR; widerspricht `AGENTS.md` §3.6 |

## Konsequenzen

- Positiv: Eine Erweiterung einer Guide-Datei ist rot, wo sie entsteht; die Grenze sinkt nur, nie unbemerkt nach oben; die Regel zum Anheben steht einmal.
- Negativ: Jede Erweiterung von `AGENTS.md` (auch ein Diff-Vorschlag des Nutzers) bricht das Gate, bis eine Verschiebung dasteht; das ist gewollt. `harness/README.md` steht außerhalb der Menge und wächst weiter (`SPEC-040` Nr. 10 d). Die Höhe der Grenze prüft kein Gate (`SPEC-040` Nr. 10 a).
- Folgepflicht: Die Verschlankung des Briefings senkt die Grenzen und verschiebt Text nach Nr. 8 b; das ändert allein der Nutzer.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| d-check, Module `file` und `structure` | keine Datei der Prüfmenge über ihrer Grenze, kein Abschnitt `### 3.<n>` über seiner | `make docs-check` |

## Re-Evaluierungs-Trigger

Wenn die Verschlankung beendet ist und die Grenzen nicht mehr sinken (zwei Verschiebungen in Folge ohne Senkung), oder wenn ein Fall eintritt, in dem Zeichen oder Token statt Bytes gebraucht werden.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-31 | Proposed | Entwurf des Architect; die Annahme steht beim Nutzer aus |
| 2026-10-31 | Accepted | Entscheidung des Owners im Gespräch: angenommen, Anweisung des Users im Hauptgespräch am 2026-10-31 |
