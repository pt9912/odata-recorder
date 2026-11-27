# CO-001: Lasttest `LH-QA-01.a` — Replay ohne p95, das Binary endet an der Speichergrenze

**Status:** Aktiv.

**Datum angelegt:** 2026-11-26. **Letzte Prüfung:** 2026-11-26.

**Betroffenes Gate:** `make lasttest` (Werkzeug außerhalb der Gate-Kette, `SPEC-216` Nr. 1; den vorläufigen Zielwert p95 ≤ 100 ms setzt `LH-QA-01` im Lastenheft, nicht dieser Carveout).

**Geltungsbereich:** Betriebsart `replay` des Aufrufs im Profil `abnahme`, also der Nachweis von `LH-QA-01.a` für den Replay-Dienst. Nicht darin: `mock` (bestanden), das Profil `rauch`, jedes Gate der Kette.

**Folge-Slice:** [slice-replay-speicher-und-suche](../planning/next/slice-replay-speicher-und-suche.md).

**Frist:** vor der Closure von [welle-betrieb](../planning/welle-betrieb.md) (die Welle schließt nicht ohne p95 für Replay, §3 dort) und damit vor M5; geprüft bei jeder Wellen-Closure.

Regeln: Baseline-Regelwerk `modul-07-carveouts.md` §Ziel-Form: Carveout — ein
Carveout braucht immer einen Auflösungs-Trigger **und** einen Folge-Slice.

---

## Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-07-carveouts.md`
§Ziel-Form: Carveout — technische Begründung.

Der Lasttest ist gefahren und liefert für `replay` keinen Messwert, weil das gemessene Binary die Last nicht übersteht. Beleg: ein Aufruf `abnahme` mit den Vorgaben auf `e7bfebb4`, 2026-11-26, 22:54 bis 23:43 Uhr, Messender Verifier (Bericht `2026-11-26-slice-lasttest-verifikation`).

| Größe | `replay`, Lauf 1 | `replay`, Lauf 2 | `mock` (Läufe 1 und 2) |
|---|---|---|---|
| Urteil | `ungueltig`, `grund=binary` | `ungueltig`, `grund=binary` | `bestanden` |
| p95 des Gesamtspans | keines (`messphase=0`) | keines (`messphase=0`) | 29,032 ms und 28,841 ms |
| Fehler von 13200 Anfragen | 13099 | 13109 | 0 und 0 |
| Gesamtspans | 102 | 92 | 13200 und 13200 |
| gleichzeitig offene Anfragen, höchstens | 185 | 183 | 2 und 2 |
| Ende des Binary | Exit 137 | Exit 137 | ordentlich |

Die Zeile `Abnahme` ist `ziel_ms=100 abnahmewert_ms=100 abnahme=nein grund=ergebnis`. Das Kernel-Protokoll nennt zu beiden Läufen den OOM-Killer der Speicher-Gruppe des Containers an 4194304 kB (`anon-rss` des Binary 4147816 kB und 4147740 kB).

**Ursache: nicht belegt.** Verdacht ist, dass `SPEC-135` Nr. 8 je Anfrage die ganze Session liest (`ListBySession`); der Lauf trennt den Speicher je Anfrage nicht von der Zahl gleichzeitiger Anfragen. Bei welcher Rate oder Größe der Dienst hält, ist nicht gemessen.

Die Behebung liegt im Produkt (`src/`), nicht im Werkzeug; sie braucht zuerst eine Messung der Ursache und dann einen Vertrag, und beides schließt der Slice aus, der gemessen hat ([slice-lasttest](../planning/done/slice-lasttest.md) §1: Tuning ist ein eigener Slice).

**Folge im Betrieb:** unter den Bedingungen von `LH-QA-01.a` fällt der Replay-Dienst aus, bis der Folge-Slice geliefert ist. Zielwert und Lastenheft bleiben unverändert (Entscheidung des Users vom 2026-11-24, [welle-betrieb](../planning/welle-betrieb.md) §3); der Owner bestätigt den Zielwert nicht, solange dieser Carveout aktiv ist.

## Auflösungs-Trigger

Ein Aufruf `make lasttest` (Profil `abnahme`, Vorgaben, Zielwert 100) auf einem Commit von `main` schreibt `abnahme=ja`: beide Betriebsarten sind gültig und bestanden.

Schreibt er für `replay` `ungueltig` oder `durchgefallen`, bleibt der Carveout aktiv und nennt einen neuen Folge-Slice. Ein anderer Zielwert ist kein Weg zur Auflösung.

## Geltungs-Konfiguration

Keine Gate-Konfiguration ist geändert: das Ziel hängt nicht an `GATE_CHECKS`, nichts wird ausgeschlossen oder gesenkt. Die Ausgabe des Werkzeugs nennt den Carveout nicht.

| Datei | Zeile/Section | Wert |
|---|---|---|
| `harness/README.md` | Tabelle der Werkzeuge, Zeile `make lasttest`, Spalte *Bindung* | `CO-001` |

## Verifikation (nach Auflösung)

- [ ] Der Aufruf `abnahme` mit `abnahme=ja` ist mit Commit, Last und Messendem in der Closure-Notiz des Folge-Slice belegt.
- [ ] Die Bindung in `harness/README.md` nennt den Carveout nicht mehr.
- [ ] `make gates` grün.
- [ ] Datei wird nach `docs/plan/carveouts/done/` bewegt (reiner `git mv`). <!-- d-check:ignore (done/ entsteht erst bei der ersten Carveout-Auflösung) -->
- [ ] Der Folge-Slice liegt in `done/`.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-26 | Angelegt | [slice-lasttest](../planning/done/slice-lasttest.md) |
| 2026-11-27 | Der User nimmt die Änderung der DoD an, aus der dieser Carveout entstand (die Bestätigung stand offen); Status, Trigger und Frist bleiben | [slice-lasttest](../planning/done/slice-lasttest.md) §1, *Zustimmung des Users* |
