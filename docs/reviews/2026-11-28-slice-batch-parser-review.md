# Review-Report: slice-batch-parser (Leser für `$batch` `multipart/mixed`) — 2026-11-28

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf.

**Gegenstand:** eigener Anteil von `git diff 7fb0f523..866d90ec` über die Pfade des Plans (Commit `866d90ec`), ohne Merges.

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`866d90ec`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-28

**Eingangs-Kontext:** Plan `slice-batch-parser` (§1, §3 mit Tabelle und Handoff, §6 mit Raster und Zusatztabelle); `SPEC-182`, `SPEC-183`, `SPEC-184` (allein der Leser), `SPEC-126` Nr. 2; `LH-FA-16`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `AGENTS.md` §3.

**Urteil:** Im Produktcode kein Befund: kein HIGH, keine Panik, keine Quadratlaufzeit, keine Abweichung vom Vertrag gefunden. Drei Zusagen haben keinen Fänger (F-1 bis F-3, je eine grüne Mutation).

**Vorgehen.** Wegwerfkopie (`git archive HEAD` über `go.mod`, `go.sum`, `src`), nur lesend in genau einen Container des Go-Images eingehängt und darin kopiert (`--network none --memory=4g --memory-swap=4g --cpus=4`), unter `flock /tmp/odata-recorder-heavy.lock`; Last 2,95 vor und 6,60 nach dem Lauf, 15 GiB verfügbar. Messend: Reviewer.

- Nullprobe der Pakete `batchcodec`, `model`, `ports/outbound`: grün.
- Differenzprobe gegen ein eigenes Orakel aus dem Wortlaut von `SPEC-183`, ohne Code des Lesers: 1 500 000 Bodies (zur Hälfte verformt; Ergebnis, Ursache und Ort verglichen), 1 000 000 Hüllen-`Content-Type`, 300 Verteilungen der Teilzahl um 1000: keine Abweichung.
- Fuzz (`go test -fuzz`, 90 s, 4 Prozesse, 8,4 Mio. Ausführungen, dasselbe Orakel): keine Panik, keine Abweichung.
- Laufzeit bei 1, 4 und 16 MiB, 12 Familien (nur `CR`, nur `LF`, viele Grenzzeilen auf beiden Ebenen, Fast-Treffer mit 70 Zeichen, lange Füllung, viele Parameter, viele Changesets): linear, höchstens 128 ms bei 16 MiB.
- Nicht gefahren: `make gates`, `make test`, `make test-race`, `make lint`, `make gegenproben`; die Raten zu Zeile 32 und die Zuteilung zu Zeile 33 sind nicht nachgemessen.

**Eigene Mutationen** (nur in der Wegwerfkopie, je `go test` des Pakets `batchcodec`):

| Nr. | Mutation | Ergebnis |
|---|---|---|
| M1 | `lies`: `io.EOF` wird vor der Größe geprüft | **grün** (F-1) |
| M2a | `lies`: ein beendeter Kontext vor einem Read meldet immer `context.Canceled` | **grün** (F-2) |
| M2b | `LeseBatch`: nach dem letzten Read immer `context.Canceled` | **grün** (F-2) |
| M3 | Changeset-Teil: `boundary` geht `Content-Transfer-Encoding` vor | **grün** (F-3) |
| M4 | `lies`: ein `context.DeadlineExceeded` von `src` beendet den Aufruf nicht | grün; gekünstelt, aus unbeschränkter Menge der Fehlerwerte, kein Befund |

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Kein Fall liefert das Byte `grenze + 1` im selben Read wie `io.EOF`; mit M1 bleibt die Suite grün. Versagen: ein Strom, der sein letztes Stück zusammen mit `io.EOF` gibt, bringt einen Body von `grenze + 1` Bytes ohne `zu_gross` durch. | `SPEC-182` Nr. 2; `SPEC-126` Nr. 2; `AGENTS.md` §3.15 (Plan Zeile 8) | `src/adapters/driven/odata/batchcodec/strom_test.go` · `q.hoechstens = stueck` in `groesseFall` | ja, `make test` bliebe mit M1 grün | Fänger führt nur eine Form der Eingabe |
| F-2 | MEDIUM | Der Kontextfehler „unverändert (`==`)“ ist an der ersten Prüfung mit `Canceled`, Ursache und Frist geführt, an den Prüfungen vor jedem weiteren Read und nach dem letzten allein mit `Canceled` (M2a, M2b). Versagen: eine während des Lesens abgelaufene Frist erreicht den Aufrufer als Abbruch. | `SPEC-182` Nr. 1; `AGENTS.md` §3.15 (je Aufrufstelle; Plan Zeile 2) | `src/adapters/driven/odata/batchcodec/argumente_test.go` · `ctx, cancel := context.WithCancel(context.Background())` in der Schleife von `TestBatchAbbruch` | ja, `make test` bliebe mit M2a, M2b grün | Fänger führt nur eine Form der Eingabe |
| F-3 | MEDIUM | Die Paare am selben Teil führen allein Verstöße eines Teils mit `application/http`; am Changeset-Teil hat kein Fall `Content-Transfer-Encoding` zusammen mit `boundary` (M3 grün) oder, nach Lesen der Fixtures, mit `rahmen`, `zu_viele` oder einem Verstoß seiner Teile. Versagen: die Folge von `SPEC-182` Nr. 5 g lässt sich am Changeset umkehren, der Client erhält `400` statt `501`. | `SPEC-182` Nr. 5 g; `AGENTS.md` §3.16 (Plan Zeile 30) | `src/adapters/driven/odata/batchcodec/rang_test.go` · `func changesetVerstoesse()` | ja, `make test` bliebe mit M3 grün | Paar von Ursachen ohne Fänger |
| F-4 | LOW | `*BatchFehler` ist der 13. Typ mit `Unwrap` auf einem Zeiger und greift darin auf den Empfänger zu; nachgezogen ist allein die Zahl 13, nicht das Maß des Negativs „nil-Zeiger anderer Fehlertypen“ (Typen mit Zugriff, davon ungefangen), das der Testkommentar im Plan nachzuziehen verlangt. | `AGENTS.md` §3.9, §3.15 | `src/adapters/driving/admin/suche_test.go` · `ist das Maß im Plan nachzuziehen` | nein | Maß eines Negativs wandert nicht mit |
| F-5 | INFO | `Content-Type: ;x=1` am Teil als `teil_typ` folgt dem Wortlaut („fehlt oder ist leer“ meint die Kopfzeile, „jeder andere“ den Medientyp) und ist von `TestBatchTeilTyp` gehalten; die Hülle nennt dieselbe Form `content_type`. Zuständig für eine Klarstellung wäre der Architect; keine Aktion erwartet. | `SPEC-183` Nr. 1, Nr. 4 a | Plan `slice-batch-parser` §3 · `Eine Lesart nennt der Bericht dem Review` | nein | Randform am Rand des Wortlauts |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/odata/batchcodec/` Produktcode: Panik, Laufzeit, Speicher, Grenzen (Größe, 1000 Teile, 100 Zeilen, Tiefe), Zustand, Importe (kein `mime/multipart`, `net/http`, `net/textproto`), Nachbau der `Content-ID`-Form (keiner), Kommentare (`AGENTS.md` §3.7) | geprüft, ohne Befund |
| Konformität je Nummer von `SPEC-182` Nr. 1 bis 6 und `SPEC-183` Nr. 1 bis 7 (Lesen und Differenzprobe) | geprüft, ohne Befund außer F-5 |
| `src/adapters/driven/odata/batchcodec/` Tests, übrige Zeilen: Mengen als Schleifen mit Zählprüfung, Fixtures aus Bytewerten (`AGENTS.md` §3.18; kein Steuerzeichen außer Tab und `LF` in den Quellen) | geprüft, ohne Befund außer F-1 bis F-3 |
| `src/hexagon/model/batch*.go`, `src/hexagon/ports/outbound/` (Port, zwei Zeilen der Abbildung, Tests) | geprüft, ohne Befund |
| Bestehende zählende Tests (`abbildung_test.go`, `services/workspace/`, `services/session/`, `services/inspection/`, `driving/admin/`) | geprüft, ohne Befund außer F-4: allein Listen und Zahlen, keine Prüfung entfällt, keine Zählung wird Untergrenze |
| Gates und Konfiguration (`.golangci.yml`, `.a-check.yml`, `harness/`, `tools/`, `Dockerfile`, `//nolint`) | geprüft, ohne Befund: im eigenen Anteil unverändert, keine Unterdrückung |
| Plan §1, §3 (Dateiliste, Messpunkt: 3049 = 2408 Tests + 641 Produktcode, nachgerechnet), §6 | geprüft, ohne Befund außer F-5 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 1 |
| INFO | 1 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe (2×) · Paar von Ursachen ohne Fänger · Maß eines Negativs wandert nicht mit · Randform am Rand des Wortlauts

## Verdikt

**Merge-blockierend:** ja, wegen F-1 bis F-3 (Fänger oder benanntes Negativ mit Grund, Folge und Maß); der Produktcode ist davon nicht betroffen. F-4: Annahme oder Begründung. F-5: ohne erwartete Aktion.

**Übergabe:** an den Implementer; die Finding-Klassen gehen in die Closure-Notiz. Dieser Report ersetzt keine Verifikation.
