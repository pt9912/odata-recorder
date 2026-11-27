# CO-001: Lasttest `LH-QA-01.a` — Replay ohne p95, das Binary endet an der Speichergrenze

**Status:** Aufgelöst am 2026-11-27 (Auflösung unten; die Bedingung an den Merge steht dort).

**Datum angelegt:** 2026-11-26. **Letzte Prüfung:** 2026-11-27.

**Betroffenes Gate:** `make lasttest` (Werkzeug außerhalb der Gate-Kette, `SPEC-216` Nr. 1; den vorläufigen Zielwert p95 ≤ 100 ms setzt `LH-QA-01` im Lastenheft, nicht dieser Carveout).

**Geltungsbereich:** Betriebsart `replay` des Aufrufs im Profil `abnahme`, also der Nachweis von `LH-QA-01.a` für den Replay-Dienst. Nicht darin: `mock` (bestanden), das Profil `rauch`, jedes Gate der Kette.

**Folge-Slice:** [slice-replay-speicher-und-suche](../planning/in-progress/slice-replay-speicher-und-suche.md).

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

**Ursache: belegt** (berichtigt am 2026-11-27; beim Anlegen war sie ein Verdacht): der Replay-Dienst las je Anfrage die ganze Session (`ListBySession`). Gemessen am unveränderten Code, 10 000 Exchanges: 15,6 MiB gehaltener Speicher und 106 bis 108 ms CPU je Anfrage, 24 bis 29 MiB je gleichzeitiger Anfrage, nach der Behebung 0,19 bis 0,50 MiB und unter 1 ms ([slice-replay-speicher-und-suche](../planning/in-progress/slice-replay-speicher-und-suche.md) §3, Messender Implementer, Stichprobe des Verifiers).

Die Behebung liegt im Produkt (`src/`), nicht im Werkzeug; sie brauchte zuerst eine Messung der Ursache und dann einen Vertrag, und beides schloss der Slice aus, der gemessen hat ([slice-lasttest](../planning/done/slice-lasttest.md) §1: Tuning ist ein eigener Slice).

**Folge im Betrieb während des Carveouts:** unter den Bedingungen von `LH-QA-01.a` fiel der Replay-Dienst aus. Zielwert und Lastenheft blieben unverändert (Entscheidung des Users vom 2026-11-24, [welle-betrieb](../planning/welle-betrieb.md) §3).

## Auflösung

Ein Aufruf `make lasttest` (Profil `abnahme`, Vorgaben, Zielwert 100) schrieb `abnahme=ja grund=ok` (Exit 0).

- **Commit:** `9c3f5f10`, Zweig `slice-replay-speicher-und-suche`. Er trägt die Behebung und enthält `main` bei `a4b79211`.
- **Messender und Zeit:** Verifier, 2026-11-27, 12:33:03 bis 13:21:50; Host mit 20 Kernen und 31 GiB, Last 3,39 vor dem Start, Proben 1,33 bis 5,37 gegen die Schwelle 10,00; fremde Container liefen mit. Beleg: Verifikation `2026-11-27-slice-replay-speicher-und-suche-verifikation`, Schritt 2.
- **Werte:** `replay` `bestanden`, p95 10,613 und 10,399 ms, 0 Fehler; `mock` `bestanden`, p95 31,509 und 32,648 ms, 0 Fehler.

**Zum Wortlaut des Triggers („auf einem Commit von `main`“):** beim Lauf war `9c3f5f10` kein Commit von `main`. Der Auftraggeber hat vor dem Aufruf entschieden: der Zweig wird vor [slice-bootstrap-m6-transport](../planning/in-progress/slice-bootstrap-m6-transport.md) nach `main` gemergt, ohne dass sich dazwischen `src/` oder `tests/` auf `main` ändern. Diese Datei erreicht `main` allein mit diesem Merge, und mit ihm ist `9c3f5f10` ein Commit von `main`, dessen `src/` und `tests/` der Merge-Commit trägt.

- **Geprüft bei der Closure:** `git diff 9c3f5f10 ae367c2f -- src tests` ist leer; die Commits der Closure ändern allein Dateien außerhalb von `src/` und `tests/`; `main` steht bei `a4b79211`, einem Vorfahren von `9c3f5f10`.
- **Offen für den Integrationsschritt:** `git diff 9c3f5f10 <Merge-Commit> -- src tests` ist leer. Ist es das nicht, trägt die Messung den Stand von `main` nicht: der Carveout wird wieder aktiv, oder ein neuer Aufruf `abnahme` auf dem Merge-Commit schreibt `abnahme=ja`.

**Nicht Teil der Auflösung:** die Bestätigung des Zielwerts von `LH-QA-01` durch den fachlichen Owner. `abnahme=ja` heißt nicht, dass er bestätigt ist (`SPEC-216` Nr. 12 b).

## Auflösungs-Trigger

Ein Aufruf `make lasttest` (Profil `abnahme`, Vorgaben, Zielwert 100) auf einem Commit von `main` schreibt `abnahme=ja`: beide Betriebsarten sind gültig und bestanden.

Schreibt er für `replay` `ungueltig` oder `durchgefallen`, bleibt der Carveout aktiv und nennt einen neuen Folge-Slice. Ein anderer Zielwert ist kein Weg zur Auflösung.

## Geltungs-Konfiguration

Keine Gate-Konfiguration war geändert: das Ziel hängt nicht an `GATE_CHECKS`, nichts wurde ausgeschlossen oder gesenkt. Die Ausgabe des Werkzeugs nannte den Carveout nicht.

| Datei | Zeile/Section | Wert |
|---|---|---|
| `harness/README.md` | Tabelle der Werkzeuge, Zeile `make lasttest`, Spalte *Bindung* | wieder `kein Gate` (bis zur Auflösung zusätzlich `CO-001`) |

## Verifikation (nach Auflösung)

- [x] Der Aufruf `abnahme` mit `abnahme=ja` ist mit Commit, Last und Messendem in der Closure-Notiz des Folge-Slice belegt.
- [x] Die Bindung in `harness/README.md` nennt den Carveout nicht mehr.
- [ ] `make gates` grün. Gelaufen auf `9c3f5f10` (Verifier, Exit 0, mehrere Stufen aus dem Schicht-Cache); die Commits danach deckt kein Lauf. Offen für den Integrationsschritt.
- [x] Datei wird nach `docs/plan/carveouts/done/` bewegt (reiner `git mv`). <!-- d-check:ignore (done/ entsteht erst bei der ersten Carveout-Auflösung) -->
- [x] Der Folge-Slice liegt in `done/` (Move in derselben Closure).

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-26 | Angelegt | [slice-lasttest](../planning/done/slice-lasttest.md) |
| 2026-11-27 | Der User nimmt die Änderung der DoD an, aus der dieser Carveout entstand (die Bestätigung stand offen); Status, Trigger und Frist bleiben | [slice-lasttest](../planning/done/slice-lasttest.md) §1, *Zustimmung des Users* |
| 2026-11-27 | Aufgelöst: Aufruf `abnahme` auf `9c3f5f10` mit `abnahme=ja grund=ok`; Ursache berichtigt auf „belegt“; Bedingung an den Merge unter *Auflösung* | [slice-replay-speicher-und-suche](../planning/in-progress/slice-replay-speicher-und-suche.md) §7 |
