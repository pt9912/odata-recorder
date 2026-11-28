# Review-Report: slice-batch-validator (Prüfung der Struktur eines `$batch`) — 2026-11-28

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf.

**Gegenstand:** `git diff e27f718e..f264a2b4` (`batch_pruefen.go`, `batch_pruefen_test.go`, Plan).

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`f264a2b4`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-28

**Eingangs-Kontext:** Plan `slice-batch-validator` (§1, §3 mit Zeilen 37 bis 43, Bericht und Prüfgrenzen, §6 mit Raster und Zusatztabelle); `SPEC-184`, `SPEC-182` Nr. 3 bis 5, `SPEC-183` Nr. 4 c, 7 und 8; `LH-FA-16`, `LH-FA-16.a`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `AGENTS.md` §3.

**Urteil:** Im Produktcode kein Befund: Methoden je Ebene, Eindeutigkeit, `Verweis`, Auflösbarkeit, Rang, `Ort` und Reinheit folgen dem Vertrag; die Differenzprobe fand keine Abweichung. Zwei Achsen der Zusagen haben weder Fänger noch benannte Prüfgrenze (F-1, F-2).

**Vorgehen.** Genau ein Container (`odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, ohne Mount, Baum per `docker cp`) unter `flock /tmp/odata-recorder-heavy.lock`; Last des Hosts 3,13 am Anfang und 3,61 am Ende, 14 GiB verfügbar. Messend: Reviewer.

- Differenzprobe gegen ein eigenes Orakel aus dem Wortlaut von `SPEC-184`, ohne Code des Modells: alle 157 464 Changesets aus drei Requests (6 Methoden × 3 `Content-ID` × 3 Ziele), 20 000 Zufalls-Batches (bis 39 Teile, bis 39 Requests je Changeset), 2 000 000 Zufallsziele für `Verweis`: keine Abweichung in Ursache, `Ort` und Verweis.
- Laufzeit der Prüfung: 1000 Durchläufe eines Changeset aus 999 Requests samt Orakel in 1,16 s.
- Prüflaufzeit der vierzehn neuen Tests: ohne Detektor 0,060, 0,071 und 0,062 s (drei Läufe), mit Detektor 1,38 s (ein Lauf), Last 3,1 bis 3,9. Das ganze Paket `model` (8,6 s, mit Detektor 128,5 s) wird davon nicht getragen.
- Abdeckung durch die neuen Tests: jede Funktion in `batch_pruefen.go` 100 %.
- Gleichwertigkeit `ende > 0` statt `ende >= 0` (Bericht des Implementers): durch Lesen bestätigt.
- Nicht gefahren: `make gates`, `make test`, `make test-race`, `make lint`, `make a-check`, `make gegenproben`; die 58 Mutationen des Implementers sind nicht wiederholt.

**Eigene Mutationen** (`mutationslauf` im selben Container, 16 Mutationen, Klasse `andere`, je ein Lauf):

| Mutation | Ergebnis |
|---|---|
| ein `DELETE` merkt seine `Content-ID` nicht | **grün** (F-1) |
| ein `DELETE` ist nie `content_id_doppelt` | **grün** (F-1) |
| Teilnummer im `Ort` modulo 20 | **grün** (F-2) |
| Requestnummer im `Ort` modulo 20 | **grün** (F-2) |
| höchstens 20 `Content-ID` werden gemerkt | **grün** (F-2) |
| nur die ersten 20 Teile werden geprüft | **grün** (F-2) |
| sechstes reserviertes Wort `root` | grün; von der Prüfgrenze zu Zeile 39 benannt, kein Befund |
| `PUT` bzw. `PATCH` merkt seine `Content-ID` nicht; `PATCH` nie doppelt; `DELETE` ohne Referenzprüfung; leere `Content-ID` gemerkt; `Content-ID` eines Requests mit Verweis nicht gemerkt; `Content-ID` nur im ersten Teil gemerkt; `GET` ohne Verweis; Schnitt auch an `#` | rot (9 von 9) |

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Methode des Trägers einer `Content-ID` und des zweiten Vorkommens ist als Achse weder geführt noch als ungeführt benannt: kein Fall hat ein `DELETE` als Träger, auf den verwiesen wird, oder als zweites Vorkommen. Versagen: eine Regression, die `DELETE` auslässt, weist einen gültigen Verweis mit `referenz` ab oder lässt eine doppelte `Content-ID` zur Ausführung durch, und beide Mutationen bleiben grün. | `AGENTS.md` §3.15 (Achsen einer Menge); `SPEC-184` Nr. 2, 4; Plan Zeilen 38, 40 | `src/hexagon/model/batch_pruefen_test.go` · `func batchPruefReihe(n int)` | ja, `make test` bliebe mit beiden Mutationen grün | Fänger schmaler als Zusage, Achse nicht benannt |
| F-2 | MEDIUM | Zahl und Stelle der Teile und Requests sind eine unbeschränkte Menge (der Leser lässt bis 1000 Teile zu), geführt bis etwa 14, und die Prüfgrenzen nennen sie nicht: sie führen allein Methodentext, Ziel, Kennungen und den Weg über den Leser. Versagen: eine Regel, die erst jenseits dieser Stellen greift (`Ort` falsch, doppelte `Content-ID` oder Teil ungeprüft), besteht die Gates; vier Mutationen mit Schwelle 20 bleiben grün. | `AGENTS.md` §3.15, §3.17; `SPEC-182` Nr. 4; `SPEC-184` Nr. 5 | Plan `slice-batch-validator` §3 · `**Prüfgrenzen** (akzeptierte Negative` | ja, `make test` bliebe mit den vier Mutationen grün | Prüfgrenze einer unbeschränkten Menge fehlt |
| F-3 | INFO | Das Negativ zu Zeile 43 (Leser und Prüfung nie als Paar geprüft) verweist auf den ersten Aufrufer; `slice-batch-changesets` §1 führt den Aufruf beider, nennt diesen Fang aber nicht. Zuständig wäre der Planner; keine Aktion erwartet. | `AGENTS.md` §3.13 | Plan `slice-batch-validator` §3 · `fällt erst beim ersten Aufrufer beider auf` | nein | Folge eines Negativs ohne Eintrag beim Nehmer |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/batch_pruefen.go`: Vertrag `SPEC-184` Nr. 1 bis 5, `Ort` und Fehlerwert nach `SPEC-182` Nr. 4, Reinheit, Laufzeit, Importe (`strconv`, `strings`), Aufruf von `ContentIDGueltig` ohne Nachbau der Form, Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft, ohne Befund |
| `src/hexagon/model/batch_pruefen_test.go`: Orakel ohne Code des Modells, Mengen als Schleifen mit Zählprüfung, Testkommentare gegen die Mengen, Namen der Tabelle gegen die Datei, Bytes aus Bytewerten, keine unsichtbaren Zeichen | geprüft, ohne Befund außer F-1 |
| Rang und Zusatztabelle (Zeile 41, drei Paare), Contract-Test gegen das Beispiel und die vier Zeilen aus `SPEC-183` Nr. 8 | geprüft, ohne Befund |
| Gates und Konfiguration (`.golangci.yml`, `.a-check.yml`, `harness/`, `tools/`, `Makefile`, `Dockerfile`, `//nolint`) | geprüft, ohne Befund: im Diff unverändert, keine Unterdrückung |
| Plan §1, §3 (Dateiliste, Messung 713 = 84 + 629, nachgerechnet), §4 (Block der Größenmarke), §6 | geprüft, ohne Befund außer F-2, F-3 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 0 |
| INFO | 1 |

**Finding-Klassen dieses Laufs:** Fänger schmaler als Zusage, Achse nicht benannt · Prüfgrenze einer unbeschränkten Menge fehlt · Folge eines Negativs ohne Eintrag beim Nehmer

## Verdikt

**Merge-blockierend:** ja, wegen F-1 und F-2 (Fänger oder benanntes Negativ mit Grund, Folge und Maß); der Produktcode ist davon nicht betroffen. F-3: ohne erwartete Aktion.

**Übergabe:** an den Implementer; die Finding-Klassen gehen in die Closure-Notiz. Dieser Report ersetzt keine Verifikation.
