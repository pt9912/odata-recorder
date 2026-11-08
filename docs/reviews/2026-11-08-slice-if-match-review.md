# Review-Report: slice-if-match — 2026-11-08

**Review-Art:** Code — gegen Plan, Vertrag (`SPEC-283`) und Hard Rules.

**Gegenstand:** `slice-if-match`, Zweig `slice-if-match`, Stand `9745ce74` (Handoff `864c4d85`, Code `8423c802`), Basis `89f4747d`.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-08

**Eingangs-Kontext:**

- Plan `slice-if-match` (Tabelle F1 bis F28, Raster 6 × 6, Zusatztabelle, §3 Mutationen, Raten, Klassen)
- Vertrag `SPEC-283` Nr. 1 bis 10, dazu `SPEC-280`, `SPEC-281`, `SPEC-192` Nr. 2 und 6, `SPEC-109` Nr. 2 und 5, `SPEC-195` Nr. 10, `SPEC-201`
- [LH-FA-15](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `AGENTS.md` §3.2, §3.6, §3.10 bis §3.18

**Messende und Grenzen.** Eigene Läufe im Container (`odata-recorder:deps`, `--network none`, `--memory 3g --cpus 4`, Quellen per `docker cp`, unter `flock /tmp/odata-recorder-heavy.lock`; Last des Hosts 5,5 bis 9,7 bei 4 CPUs im Container). `go test -count=1 ./src/... ./tests/...` ist grün. Die Mutationen liefen über `model` und `mock` (nicht über `src/bootstrap`; dort lief nur der Prozess-Treiber unten). Nicht von mir wiederholt: `make gates`, `make katalog-check` (F23), `make test-race`.

---

## Findings

| kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|
| MEDIUM | `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe) | `src/hexagon/model/vorbedingung.go` · `liste.stern \|\| liste.enthaelt(tag)` | Mutation `liste.stern && version > 0` (`If-None-Match: *` bei Version 0 erfüllt) bleibt in `model` und `mock` grün: `If-None-Match: *` wird nur bei Versionen ≥ 1 geprüft, Version 0 (der Baseline-Eintrag) fehlt in F11 und im Wertebereich der Goldens. Szenario: `PUT` oder `PATCH` mit `If-None-Match: *` an einem unveränderten Baseline-Eintrag gelänge statt `412`. Nur mein Orakel wird rot. | ja (Mutation in `model`, `mock`) | Fänger lässt Version 0 aus (Grenze der Zähler) |
| MEDIUM | `AGENTS.md` §3.15, §3.16 (Zelle ohne Fänger); `SPEC-283` Nr. 9 d | `docs/plan/planning/in-progress/slice-if-match.md` · Raster, Spalte „Abbruch, Wiederholung“ (`Nr. 9 d`) | Die Zelle „Abbruch“ nennt für PATCH, PUT, DELETE und `If-None-Match` `Nr. 9 d`; weder Tabelle F1 bis F28 noch Zusatztabelle führen einen Fänger. Mutation „Auswertung vor der Fehlerbehandlung von `Entities.Get`“ (`z1`: bei Port-Fehler und nicht erfüllter Bedingung `ODR-E6010` statt Port- oder Kontextfehler) bleibt in `model` und `mock` grün; kein Test setzt einen Fehler an `Entities.Get` zusammen mit einem Header. | ja (Mutation `z1`) | Zusage Abbruch ohne Fänger |
| INFO | `AGENTS.md` §3.10 (Rate je Mutation) | Plan §3 · „Ende-zu-Ende über die echte SQLite-Transaktion … 20 von 20“ | Meine Mutation derselben Klasse („Version in einer eigenen Transaktion vor der Auswertung gelesen“) ist am SQLite-Test `TestIfMatchWettlaufSQLite` in 14 von 20 Prozessen rot (6 grün), am Mock-Test mit Haltepunkt in 20 von 20; die 20 von 20 des Plans gelten für dessen Form der Mutation. Der deterministische Fänger ist der Haltepunkt, der SQLite-Lauf ist der wahrscheinliche. Der Plan nennt „Größenordnung“, die Zusage ist damit gedeckt. | ja | Rate hängt an der Form der Mutation |
| INFO | `AGENTS.md` §3.11 | `src/hexagon/services/mock/if_match_test.go` · `etagUmg(t, true, v, v, v)` | Die Dienst-Fänger F1 bis F5, F13, F14 bauen alle drei Einträge mit derselben Version; eine Mutation, die die Version eines anderen Eintrags vergliche, bliebe dort grün. Der Code liest genau einen Eintrag (`Entities.Get`), eine solche Mutation ist hier nicht erreichbar. | nein | Fixture mit gleichen Werten |
| INFO | Maintainability | Plan §6 · „Zahl der Aufrufe von `Workspaces.Get`“ und „Rang von `ODR-E9001`“ | Beide beim Architect liegenden Punkte geprüft: Spec (`SPEC-283` Nr. 6, Stand `9745ce74`), Plan-Zeile F2 und `abgewiesen()` im Test nennen übereinstimmend `Workspaces.Get` 2; der Code wertet nach `entscheide` aus, ein Programmfehler in `legeAn` oder `aendere` geht vor, als akzeptiertes Negativ mit Grund und Folge geführt. Kein Befund. | ja | — |
| INFO | Maintainability | Zählstellen 40 auf 41 | `grep` nach Literalen (`40`, `41`, `E6009`, „15 Codes“) über `src`, `tests`, `tools`, `spec`, `docs/user`: alle acht genannten Dateien sind nachgezogen, `SPEC-230` Nr. 8 trägt 41; die Treffer `tools/katalog` (Fabrik von 40 Codes) und `snapshot_test.go` (40 Namen) sind andere Mengen. Die siebente Zählstelle (`policy/entscheide_test.go`, `total_test.go`) liegt auf diesem Zweig nicht vor (Geber noch nicht gemerged): bedingt, kein Befund. | ja | — |
| INFO | Maintainability | Mutation `v` (`!es.Deleted && …`) und `w` (`es.Version > 0 && …`) | `v` ist äquivalent (ein Tombstone endet vorher in `ODR-E6008`). `w` bleibt in `mock` grün (Version 0 mit nicht erfüllter Bedingung fehlt in F2, `TestIfMatchVeraltet` nimmt 1, 7, 10); `TestIfMatchHttpEndeZuEnde` in `src/bootstrap` hält den Fall (Version 0, Tag 7, `412`) fest. Dort habe ich nicht mutiert. | nein | Randform Version 0 nur im Ende-zu-Ende |

## Sicherheits- und Korrektheitsachse (Lost Update)

Eine Mutation, die ein nach `SPEC-283` abzuweisendes `If-Match` durchlässt, habe ich außer den zwei MEDIUM-Fällen oben nicht gefunden.

- **Grammatik.** Eigener Orakel-Probe aus dem Wortlaut von `SPEC-283` Nr. 3 bis 5 (Zustandsautomat über Bytes, unabhängig vom Code; `zz_orakel_test.go`, nicht im Repo): erschöpfend alle Werte bis Länge 5 über 16 Bytes (`W / w " 7 0 * , Leerzeichen Tab a 0x7F 0x80 0x00 1`) = 1 118 481 Werte × beide Header, dazu 400 000 zufällige Listen aus 23 Elementen mit 1 bis 3 Zeilen und 6 Trennern, jeweils gegen die Versionen 0, 1, 7, 10, 70, 99, 100, 2^40, 2^62+1, 2^63−1. Null Abweichungen von `VorbedingungErfuellt`.
- **Eigene Mutationen** (23, ohne die 58 des Implementers; Namen nur im Lauf, nicht im Repo): von den Tests des Implementers (ohne mein Orakel) rot waren 19 (`b`, `c`, `d`, `e`, `g`, `h`, `i`, `k`, `k2`, `l`, `m`, `m2`, `o`, `r`, `r2`, `u`, `x`, `y`, `z2`). Grün blieben `j` (nur mein Orakel rot, MEDIUM 1), `v` (äquivalent), `w` (siehe INFO), `z1` (MEDIUM 2).
- **Atomarität F15 bis F18.** Version lesen, vergleichen und `Put` liegen in `imWorkspace` in derselben `WithWorkspaceTransaction`, die der SQLite-Adapter mit `BEGIN IMMEDIATE` öffnet. Ohne Mutation: `TestIfMatchWettlaufSQLite` (4 Formen × 5 Aufbauten × 20 Anfragen; PATCH, PUT, Mischung, DELETE: ein `204`, 19 × `ODR-E6010`, bei DELETE 19 × `ODR-E6008`) in 20 von 20 Prozessen grün (`-test.count=1`, ohne Detektor, Last 5,5), die Mock-Tests `TestIfMatchWettlauf*` in 20 von 20 Prozessen grün. Mutation „Lesen in eigener Transaktion davor“: Mock 20 von 20 rot, SQLite 14 von 20 rot (siehe INFO).
- **Rangfolge.** `ODR-E6008`, `ODR-E6003`, `ODR-E6007`, `ODR-E7001` und die Verstöße vor der Transaktion gehen vor `ODR-E6010` (Mutationen `u`, `x` und die des Implementers rot); `PATCH` an `partial` ist `ODR-E6003`, `PUT` und `DELETE` an `partial` werden verglichen; `ETAG_ENABLED` aus und `POST` ohne Wirkung (Mutationen `l`, `m`, `m2` am Gate `d.Etag` und `Methode` rot).

## Raster- und Tabellenauflösung

Jede Zeile F1 bis F27 und jede der 36 Rasterzellen hat Fänger oder Negativ im Plan, außer der Spalte „Abbruch“ (MEDIUM 2); F28 und die Prüfgrenzen (Versionen, Länge, Zugriff bei aus) tragen Grund und Folge. Die Zeilen habe ich nicht alle mit eigener Mutation aufgelöst: durch eigene Mutationen sind F1 (`o`), F3 (`z2`), F5 (`x` für Zeilen, `h`), F6 (`h`), F7 (`i`, `g`), F8 (`b`, `e`, `g`), F9 (`x`), F10 (`k`, `k2`), F11 (`r`, `c`; Lücke `j`), F12 (`c`, `d`), F14 (`u`), F24 und F25 (`l`, `m`, `m2` am Gate) und F21 (`y`) gedeckt; F2, F4, F13, F15 bis F20, F26, F27 stützen sich auf die Mutationen des Implementers und meine Prozessläufe (F15 bis F17) ohne eigene Mutation je Zeile.

## Geprüft, ohne Befund

- `src/hexagon/model/` (`vorbedingung.go`, `fehler.go`, `fehler_test.go`): geprüft, außer MEDIUM 1 ohne Befund
- `src/hexagon/services/mock/` (`schreiben.go`, Tests): geprüft, außer MEDIUM 2 ohne Befund
- `src/adapters/driving/httpodata/` (`if_match_test.go`): geprüft, ohne Befund
- `src/bootstrap/` (`if_match_ende_zu_ende_test.go`): geprüft, ohne Befund
- `src/adapters/driven/audit/`, `telemetry/`, `src/hexagon/services/snapshot/` (Zählstellen): geprüft, ohne Befund
- `docs/user/benutzerhandbuch.md` (Zeile `ODR-E6010`, Satz zu `ETAG_ENABLED`): geprüft, ohne Befund; `make katalog-check` nicht wiederholt
- `AGENTS.md` §3.2 (keine `//nolint`), §3.7 (Kommentare tragen Zusage oder Kopplung), §3.18 (kein Escape-Wert im Produktcode): ohne Befund

## Zähler (Steering-Loop)

`Fänger lässt Version 0 aus` und `Zusage Abbruch ohne Fänger` sind je ein erstes Auftreten in diesem Report; ob eine der beiden Klassen früher schon vorkam, habe ich nicht verglichen.
