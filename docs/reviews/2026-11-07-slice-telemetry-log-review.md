# Review-Report: slice-telemetry-log — 2026-11-07

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers)

**Gegenstand:** Zweig slice-telemetry-log, HEAD 7e64f20, Basis 47b9716 (Handoff-Commit des Implementers); Messung `git diff --numstat` über `src`: 1372 von Marke 2900 (nachgerechnet: 1372; mit dem Plan 1408).

**Skill:** `.harness/skills/reviewer.md` @ 7e64f20 · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-07

**Eingangs-Kontext:**

- `slice-telemetry-log` (§3 Entscheidungen 10 bis 17, Tabelle L01 bis L20, Mengen M1 bis M10 und M16 (Anteil L), Bericht des Implementers, Raster, Fehlerformen der Senke, Fenster der Sperre, Zweig-Menge Z01 bis Z37 in §6)
- `SPEC-230` bis `SPEC-232` (`SPEC-231` Nr. 1 bis 3, `SPEC-232` Nr. 2, 3, 5)
- `LH-QA-05`, `LH-QA-04`, `LH-FA-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `AGENTS.md` (Hard Rules, besonders §3.9, §3.10, §3.11, §3.15, §3.17, §3.18)

**Vorgehen.** Eigene Mutationen mit `tools/harness/mutationslauf.sh` (Image `odata-recorder:race`, `--memory=2g`, unter `flock`, Last des Hosts 1,6 bis 2,8 von 4 CPUs): 26 Mutationen in zwei Listen, jede am Code, nie an der Fixture. Ergebnis: 14 rot, 5 `rot-paket`, 6 grün, 1 flatternd, 0 Übersetzungsfehler. Nicht neu gemessen: die Abdeckung 98,7 Prozent unter dem Detektor (ein Container-Lauf mit Kopie des Baums wurde von der Sicherheitsprüfung abgelehnt); die Zweig-Menge ist deshalb gegen den Bericht des Implementers, nicht gegen eine eigene Messung aufgelöst.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Der Fänger der Maskierung prüft Rohbytes nur für `<`, `>`, `&` und DEL; die Zusage sagt „jedes andere Zeichen steht unverändert“. Sechs Mutationen, die ein weiteres Zeichen maskieren (Leerzeichen als `<=`-Grenze, é, `/`, `a`, alle Zeichen von U+0080 bis U+FFFF), blieben in `TestMsgZeichen` und `TestMsgLaenge` grün, weil der Leser die Escape-Folge zum selben Wert liest. Szenario: eine Zeile, die Nicht-ASCII-Text oder Leerzeichen als Escape-Folgen trägt, verletzt `SPEC-231` Nr. 1 („Maskiert werden allein …“), und kein Test wird rot. | `AGENTS.md` §3.15 (Fänger führt jede Form der Eingabe), Reviewer-Skill „Fänger führt nur eine Form der Eingabe“ · `SPEC-231` Nr. 1 | `src/adapters/driven/telemetry/zeile_test.go` · `for _, c := range []byte{'<', '>', '&', 0x7F} {`; Code: `src/adapters/driven/telemetry/zeile.go` · `case r < ' ' \|\| r == zeilenTrenner \|\| r == absatzTrenner:` | ja: Mutationen `rv-space-maskiert`, `rv-e9-maskiert`, `rv-e9-maskiert-zeichen`, `rv-slash-maskiert`, `rv-a-maskiert`, `rv-nichtascii-maskiert` (alle `gruen`, 2 Prozesse) | Rohbyte-Fänger deckt nur vier der unmaskierten Zeichen |
| F-2 | INFO | Die Zusage „`Write` läuft im Aufruf“ trägt allein der Haltepunkt, wie der Plan sagt; das habe ich bestätigt. Haltepunkt `TestLogKehrtNachWriteZurueck`: 30 von 30 mit und 30 von 30 ohne Detektor für beide Formen (Goroutine je Zeile; Goroutine unter gehaltener Sperre, Reihenfolge je Goroutine erhalten). Die Masse `TestZeilenReihenfolgeMasse` ist bei der Form mit gehaltener Sperre flatternd (19 von 30 mit, 30 von 30 ohne Detektor), bei Goroutine je Zeile 30 von 30 und 30 von 30; die Raten des Plans (24 von 50 mit, 49 von 50 ohne) gelten für die Form des Implementers und sind damit ehrlich als Zusatzfänger geführt. | `AGENTS.md` §3.10 | `docs/plan/planning/done/slice-telemetry-log.md` · `Die Zusage „`Write` läuft im Aufruf“ trägt deshalb allein der Haltepunkt` | ja: Mutationen `rv-l18-goroutine-haltepunkt`, `rv-l18-goroutine-masse`, `rv-l18-async-sperre-haltepunkt`, `rv-l18-async-sperre-masse` | Rate mit Zahl der Prozesse bestätigt |
| F-3 | INFO | Der Ersatz `TestLogSchreibtZeile` (L13) prüft die Zeilen, aber nicht mehr, dass `Log` den Stand außer den Verwürfen nicht ändert (`TestKernSchreibtNichts` prüfte das früher nach den `Log`-Aufrufen). Eine Mutation, die `Log` einen Fehlerzähler erhöhen lässt, färbt weder `TestLogSchreibtZeile` noch `TestZeilenform`; sie wird vom Paket rot (`TestSnapshotZeitpunkt`), nicht von einem benannten Test. Kein Befund gegen die Zusagen der Tabelle, da L13 diese Wirkung nicht zusagt. | Maintainability | `src/adapters/driven/telemetry/zeile_test.go` · `func TestLogSchreibtZeile(t *testing.T)` | ja: Mutationen `rv2-log-zaehlt-fehler-*` (5 Namen, alle `rot-paket`, Zeile F-3 im Hinweis des Reviewer-Skills: rot, aber nicht aus dem benannten Test) | Stand-Wirkung von `Log` nur über Nachbartest gefangen |
| F-4 | INFO | Die Abdeckung 98,7 Prozent und „alle Funktionen von `zeile.go` 100,0 Prozent“ habe ich nicht neu gemessen. Die Seiten von `err != nil \|\| n < len(z)` trägt eine Mutation, die ich selbst fuhr: `n < len(z)-1` wird in `TestSenkeAusfall` rot. | `AGENTS.md` §3.15, Register `zweig-ohne-lauf-unter-dem-detektor` | `docs/plan/planning/done/slice-telemetry-log.md` · `Anweisungen des Pakets 98,7 %` | nein (Lauf nicht möglich in dieser Rolle); der Verifier kann `make test-race` mit Profil fahren | Abdeckung nicht unabhängig bestätigt |

## Mutationen (eigene Läufe)

Klasse `andere`, ein Prozess je Stufe:

| Zusage | Mutation | Name | Ergebnis |
|---|---|---|---|
| L05 | Kürzen mit `continue` statt `break` (längster Anfang entfällt) | `rv-msg-continue` | rot, `TestMsgLaenge` |
| L05 | Länge in Zeichen statt Bytes | `rv-msg-runen-gezaehlt` | rot, `TestMsgLaenge` |
| L06 | ungültiges Byte als `?` statt U+FFFD | `rv-msg-fragezeichen` | rot, `TestMsgZeichen` |
| L10 | `n < len(z)-1` | `rv-senke-len-minus-1` | rot, `TestSenkeAusfall` |
| L07 | `bitSize` 32 | `rv-float32` | rot, `TestZahlen` |
| L01 | `befunde` und `blob` getauscht (Ende der Folge) | `rv-folge-blob-befunde` | rot, `TestZeilenform` |
| L06 | sechs Maskierungen weiterer Zeichen | `rv-space-maskiert` u. a. | grün (F-1) |
| Log ändert Stand | `Log` erhöht einen Fehlerzähler | `rv2-log-zaehlt-fehler-*` | `rot-paket` (F-3) |

Klasse `nebenläufig`, je 30 Prozesse mit `-test.count=1` und `-race` und 30 ohne Detektor, vier Prozesse zugleich, Messende `mutationslauf.sh`:

| Zusage | Mutation | Name | mit Detektor | ohne Detektor |
|---|---|---|---|---|
| L11 | `Write` unter der Sperre des Zustands | `rv-l11-write-unter-zustand` | 30/30 | 30/30 |
| L17 | ohne Sperre, Haltepunkt | `rv-l17-ohne-sperre-haltepunkt` | 30/30 | 30/30 |
| L17 | ohne Sperre, Masse | `rv-l17-ohne-sperre-masse` | 30/30 | 30/30 |
| L18 | Goroutine je Zeile, Haltepunkt | `rv-l18-goroutine-haltepunkt` | 30/30 | 30/30 |
| L18 | Goroutine je Zeile, Masse | `rv-l18-goroutine-masse` | 30/30 | 30/30 |
| L18 | Goroutine unter gehaltener Sperre, Haltepunkt | `rv-l18-async-sperre-haltepunkt` | 30/30 | 30/30 |
| L18 | Goroutine unter gehaltener Sperre, Masse | `rv-l18-async-sperre-masse` | 19/30 (flatternd) | 30/30 |
| L19 | `Now()` unter der Sperre des Zustands | `rv-l19-now-unter-zustand` | 30/30 | 30/30 |
| L20 | `Now()` unter der Sperre der Zeile | `rv-l20-now-unter-zeile` | 30/30 | 30/30 |

Die Haltepunkt-Fänger sind in jedem gemessenen Lauf rot; die flatternde Masse ist im Plan als Zusatzfänger geführt, nicht als Träger der Zusage.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/telemetry/zeile.go` (Zeile, Stufen, `msg`, Zahlen, Senke, Sperre) | geprüft, ohne Befund am Code; Zusagen von `SPEC-231` Nr. 1 bis 3 und `SPEC-232` Nr. 3, 5 gegen die Mutationen oben aufgelöst |
| `src/adapters/driven/telemetry/telemetry.go` (Diff) | geprüft, ohne Befund; Kommentare tragen Zusage und Kopplung, keine Chronik |
| `src/adapters/driven/telemetry/hilfen_test.go` (`FakeSenke`, `RueckrufUhr`, Leser, Modelle) | geprüft, ohne Befund: die Hülle hält jedes Argument als Kopie und als Verweis, zählt jeden `Write`, hält den Höchstwert, kann Fehler, Kurzschreiben, Blockieren; der Leser nutzt `json.Decoder.Token`, das Modell ist eigenes Orakel |
| `src/adapters/driven/telemetry/zeile_test.go` (Mengen M1 bis M9, Canary M10) | geprüft; Zählprüfungen der Mengen gegen Quelle und Produkt stimmen (M1 aus `logSchluessel`, M2 469, M3 8909, M4 81, M6 14, M7 182, M8 18, M9 14, M10 120); einziger Befund F-1 |
| `src/adapters/driven/telemetry/sperre_test.go` (L11, L17 bis L20, M16) | geprüft, ohne Befund |
| `telemetry_test.go`, `regeln_test.go` (Anpassung der festen Zahlen 37, 7, 5, 12, 41 an die Quellen) | geprüft, ohne Befund: Orakel-Listen bleiben Textlisten, Länge gegen `model.AlleCodes()`, `support.Verwuerfe()`, `support.Spans()`, `support.Eimer()` gehalten, Meldung sagt „Orakel-Liste nachziehen“; die Sendung für `ODR-E6008` steht in `slice-mock-lesen` §1 |
| `docs/plan/planning/done/slice-telemetry-log.md` | geprüft, ohne Befund: Bericht, Raten mit Zahl der Prozesse und Messendem, Gegengelesen-Absatz, Zuschnitt des Zeilenschreibers (Zeile „Handle/Span-Zeile ohne zweite Fassung“ steht in §6 mit Ausgang offen; Adresse trägt `slice-telemetry-handler` §6 und `slice-telemetry-spans` §6) |
| `.golangci.yml`, `src/bootstrap/`, `spec/`, `harness/` | Diff leer; keine `//nolint` im Paket |
| Rohbytes (`AGENTS.md` §3.18) | `grep -P` auf Backslash-u mit vier Hex-Ziffern über `src/adapters/driven/telemetry/*.go` und den Plan: kein Treffer; Fixtures aus Bytewerten |
| Größenmarke | 1372 von 2900, nachgerechnet |

## Summary

Kein HIGH. Ein MEDIUM (F-1): die Klasse „unmaskierte Zeichen“ hat einen Rohbyte-Fänger nur für vier Zeichen; sechs Mutationen sind grün. Die nebenläufigen Zusagen sind mit Haltepunkt-Fängern in jedem gemessenen Lauf gefangen (Raten oben, 30 Prozesse je Arm); die Zusage „`Write` im Aufruf“ trägt, wie der Plan sagt, allein der Haltepunkt. Drei INFO (F-2 bis F-4). Die Laufzeit der Stufen und die Abdeckung hat der Verifier zu messen.
