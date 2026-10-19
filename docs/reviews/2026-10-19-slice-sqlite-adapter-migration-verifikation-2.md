# Verifikations-Report 2: slice-sqlite-adapter-migration — 2026-10-19

**Art:** Zweite Verifikation (Modul 11, frischer Kontext) — gegen DoD, Closure-Trigger, Spezifikation und Entscheidungen. Nicht Review, nicht Validierung.

**Gegenstand:** [slice-sqlite-adapter-migration](../plan/planning/in-progress/slice-sqlite-adapter-migration.md), HEAD `4caee05` (davor `234ec86`, Vertrag zuerst).
Bezug: [LH-FA-03](../../spec/lastenheft.md), [LH-QA-06](../../spec/lastenheft.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md).
Vorgänger: [Verifikation 1](2026-10-19-slice-sqlite-adapter-migration-verifikation.md) (NICHT BESTANDEN: V-H1, V-M1 bis V-M3, V-L1 bis V-L3).

**Rolle:** Verifier · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

## Urteil: BESTANDEN MIT AUFLAGEN

Alle Vorbefunde sind behoben oder benannt: V-H1 ist gemessen behoben, V-M1, V-M3, V-L2 sind gefangen, V-M2 ist als Ausschluss im Plan geführt (mit Auflage, siehe V2-M2).
Es gibt kein HIGH.
Offen sind ein MEDIUM mit Fänger-Lücke (V2-M1), ein MEDIUM im Plan (V2-M2, ungecommittet) und LOW-Befunde.
Closure-Notiz, Register und DoD-Häkchen stehen aus (erwartet, nicht Gegenstand des Urteils).

**Anzahl:** HIGH 0 · MEDIUM 2 · LOW 5 · INFO 2.

## Läufe (Verifier selbst)

| Lauf | Ergebnis |
|---|---|
| `make gates` (HEAD, Arbeitsbaum mit der ungecommitteten Plan-Änderung) | Exit 0: a-check 0 Befunde, d-check 76 Dateien 0 Befunde, Build, Test, Lint, `schema-check` (alle `ok`, `foreign_key_check`, `integrity_check`), `migration-check: ok`. Die Stufen Build/Test/Lint kamen großenteils aus dem Layer-Cache; die Tests des Adapters habe ich deshalb unten selbst gefahren |
| `make gegenproben` bei vorhandenem Stempel | "Eingaben unverändert", nichts läuft |
| `rm harness/gegenproben.stempel; make gegenproben` (erzwungen) | Exit 0, `gegenproben: 23 Fälle, 0 rot`; neuer Stempel byte-gleich dem versionierten (`7351d513…`) |
| `go test -count=2000 -run Gleichzeitig ./src/adapters/driven/sqlite/` im Container (`golang:1.27-alpine@sha256:8a5910f3…`, Netz für Module, ohne Parallellast) | `ok`, 319,7 s, 2000 von 2000 grün (50×40 = 2000 erfüllt). Ein früherer, ungewollt überlappender Lauf gleicher Art war ebenfalls grün (361 s), zählt nicht als Beleg |
| dasselbe mit `--cpus=2` (die Konfiguration mit der höchsten Fehlerrate in Verifikation 1), `-count=600` | `ok`, 60,9 s |
| ganzes Adapter-Paket, `--cpus=1`, `-count=30` (Zeitgrenzen der WAL-Tests) | `ok`, 56,8 s |

Alle Mutationen liefen in einem Wegwerf-Klon (`git archive HEAD`) außerhalb des Repos; Quellen danach mit dem Repo verglichen (gleich).
Arbeitsbaum am Ende: nur die ungecommittete Plan-Änderung (nicht von mir), Stempel unverändert.
Grundlauf des Klons: Adapter-Paket `ok`, Gegenproben 23 Fälle grün.

## Findings

### HIGH

Keine.

### MEDIUM

#### V2-M1 — SPEC-103 Rand 1: `SQLITE_LOCKED` als Wiederholungsgrund ohne Fänger, obwohl der Plan "falsche Codes" als Mutation führt
- `stufe`: Zusage · Fänger (§3) · `AGENTS.md` §3.10
- `befund`: Die Spezifikation sagt "wiederholt sich bei `SQLITE_BUSY` und `SQLITE_LOCKED`".
  Mutation `return c == busy || c == locked` zu `return c == busy`: Adapter-Paket bleibt `ok`.
  Der Test `belegtFehler` erzeugt nur BUSY (Code 5); ein LOCKED-Fehler kommt in keinem Test vor.
  Ebenso grün: Maske `se.Code() & 0xff` entfernt (erweiterte Codes wie `SQLITE_BUSY_RECOVERY` bleiben ungetestet).
  Der Plan §3 behauptet die Mutation "falsche Codes"; sie ist nur für die Gegenrichtung (`return c == locked`, `return false`, `return true`) rot.
- `verifizierbar`: ja — `src/adapters/driven/sqlite/sqlite.go`, Funktion `istBelegt`.

#### V2-M2 — Die Abbildung `ErrBeschaedigt` auf `ODR-E3001`: Plan trägt sie nur im Arbeitsbaum, und §3 widerspricht
- `stufe`: Plan · `AGENTS.md` §3.9, §3.13 · V-M2 aus Verifikation 1
- `befund`:
  1. §1 *Ausdrücklich NICHT* nennt die Abbildung jetzt ausdrücklich (Zeile 40: gehört der meldenden Schicht, kein Slice dafür, Punkt geht an den Planner, Vergabe beim Planen von `welle-record-replay`).
     Das erfüllt §3.13 in der Form "Nehmer fehlt: Punkt geht an den Planner", und `welle-record-replay` steht in der Roadmap.
     Aber: Die Zeile steht **nur im Arbeitsbaum** (`git diff HEAD` zeigt +1 Zeile; zu Beginn meines Laufs war der Baum sauber, die Änderung kam während des Laufs hinzu).
     Der Commit `4caee05` trägt sie nicht.
  2. §3 (Zeile 91) sagt weiter "`ODR-E3001` (Recording beschädigt) gilt für eine beschädigte Datei", und die Zusage-Zeile "Keine SQLite-Datei ist `ODR-E3001`" nennt den Code als Zusage des Adapters.
     Geliefert wird `ErrBeschaedigt`; die Abbildung ist nirgends gebaut oder getestet. §3 ist damit nicht nachgezogen (§3.9) und widerspricht §1.
  3. Der Punkt steht an keiner Stelle, die der Planner beim Planen von `welle-record-replay` liest (`welle-persistenz.md`, Roadmap, Register): Wandert der Slice nach `done/`, bleibt die Zusage ohne Adresse.
- `verifizierbar`: ja — `git diff HEAD -- docs/plan/planning/in-progress/slice-sqlite-adapter-migration.md`; `grep -rn E3001 src docs/plan`.

### LOW

#### V2-L1 — SPEC-103 Rand 6, zweiter Satz: die genannten Codes sind nicht gefangen
- Mutation `c == notADB || c == corrupt || true` ist jetzt rot (`TestAndereFehlerBleibenUnklassifiziert`, V-M1 im Kern behoben).
- Mutation `… || c == 14` (`SQLITE_CANTOPEN` wird beschädigt) bleibt grün; `SQLITE_IOERR` analog.
  Der Test nutzt als "anderen Fehler" einen Fremdschlüssel-Fehler (787), nicht die in der Spezifikation genannten Codes.

#### V2-L2 — SPEC-104 Rand 5: "mehr als zwei Argumente" ohne Fänger
- Mutation `[ "$#" -eq 2 ]` zu `-ge 2` in `migration-check.sh`: `gegenproben: 23 Fälle, 0 rot`. Fall 23 fährt nur ein Argument.

#### V2-L3 — Plan §3 ist unvollständig und die Marke stimmt nicht (`AGENTS.md` §3.9, §3.14)
- Der Test `TestWalUmstellenScheitertNachDerFristUndWirdNachgeholt` (SPEC-103 Rand 1: Fehler nach Ablauf, Migration bleibt, späteres Öffnen stellt WAL erneut sicher) steht in keiner Zeile der Tabelle *Zusage · Fänger*.
- §4 nennt 987 Zeilen. `git diff --numstat 0355957..4caee05` über die Pfade aus §3 (go.sum eingeschlossen) ergibt **1198** (1165 Einfügungen, 33 Löschungen; 1148 ohne `go.sum`); bei `d015e82` waren es 982, bei `234ec86` 983. Die Zahl ist nicht reproduzierbar (V-L1 teilweise offen).
- Behoben aus V-L1: Kopf `Berührte Spec-Stellen` nennt `SPEC-103/104/102`; "Fälle 17 bis 23"; Risiko 2 ist *entfallen*.

#### V2-L4 — Kleinere Lücken ohne Spec-Zusage
- Mutation "Abbruch durch Kontext liefert `nil`" (`case <-ctx.Done(): return nil`): grün; kein Test der Kontext-Abbruchs-Zweiges von `stelleWalSicher`. SPEC-103 sagt dazu nichts; Randform des Abbruchs gehört nach `AGENTS.md` §3.12 vor den Code in die Spezifikation.
- Mutation "jeder Nicht-SQLite-Fehler gilt als belegt": grün (`TestWalUmstellenWiederholtNurBeiBelegt` nutzt einen SQLite-Fehler 787). Der Plan sagt "nur dann"; belegt ist es für SQLite-Fehler.

#### V2-L5 — Beobachtungs-Register noch ohne Beleg aus diesem Slice (erwartet vor der Closure)
- Verifikation 1 und diese hier liefern Belege für `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (V2-M1, V2-L1, V2-L2 sind erneut dieselbe Klasse: Vertrag nennt mehr Fälle, als der Fänger trennt). Die `evidence/`-Dateien fehlen noch.

### INFO
- **I1** — Die eigene Last dieser Prüfung: ein ungewollt überlappender zweiter 2000er-Lauf verfälschte den ersten Messlauf (Last höher, nicht niedriger); der gewertete Lauf lief allein.
- **I2** — Fortbestand aus Verifikation 1: Composition Root ist von der `tech`-Regel ausgenommen; Datei mit intaktem Kopf und zerstörtem Rest öffnet (Spezifikation: keine `integrity_check`-Prüfung beim Öffnen).

## Stufe 1 — DoD-Punkte

| DoD | Urteil | Beleg |
|---|---|---|
| 1 Treiber, `tech`-Regel, `gomodguard_v2`, `Open` setzt Pragmas und Migration | erfüllt | Gates grün; Mutationen `foreign_keys` und `busy_timeout` rot; WAL bei jedem Öffnen (Mutation: Aufruf entfernt, rot in `TestOeffnenStelltWalWiederHer`: "Modus nach dem Öffnen: delete") |
| 2 Gate `migration-check` | erfüllt | in `GATE_CHECKS`; Gegenproben-Mutationen unten rot |
| 3 Tests am Adapter und Randformen aus §6 | erfüllt, mit V2-M1, V2-L1 | alle zehn Randformen von SPEC-103 haben einen Test; Lücken bei den genannten Codes |
| 4 `make gates` grün | erfüllt | Exit 0; Flackern nicht mehr beobachtet (2000 + 600 + 30 Läufe) |

Closure-Pflichten (Review-Report liegt vor, Doku-Update, Closure-Notiz, Register, Risiko-Ausgänge, Paarungen): Review-Report und Doku-Update (SPEC-102/103/104, `harness/README.md`) liegen vor; Risiko-Ausgänge in §6 stehen alle (entfallen); Closure-Notiz und Register stehen aus.

## Stufe 2 — Closure-Trigger (§5)

| Kriterium | Urteil |
|---|---|
| DoD vollständig | nein (Closure-Notiz, Register, Häkchen stehen aus; V2-M1/V2-M2 zu entscheiden) |
| `make gates` grün mit `migration-check` | ja |
| geänderte Migration färbt Gates rot | ja (in Verifikation 1 gefahren; `migration-check.sh` und Gate unverändert seither, Gegenproben 18 und 19 rot unter Mutation) |
| `make gegenproben` grün | ja, erzwungen, 23 Fälle |
| Closure-Notiz mit Lerneintrag | steht aus |

## Stufe 3 — SPEC-103 Randformen

| Rand | Urteil |
|---|---|
| 1 Pragmas je Verbindung; WAL bei jedem Öffnen mit Wiederholung bei BUSY/LOCKED bis zur Frist; danach Fehler der Datenbank, Migration bleibt, späteres Öffnen stellt erneut sicher; abgelehnte Datei byte-gleich | erfüllt; Wiederholung bei BUSY gefangen (keine Wiederholung, nie belegt, nur LOCKED, Frist ignoriert: rot bzw. Zeitüberschreitung); LOCKED nur als Lücke V2-M1; Nachholen und Frist durch `TestWalUmstellenScheitertNachDerFristUndWirdNachgeholt`; WAL vor der Versionsprüfung: `TestNeuereSchemaversionWirdVerweigert`, `TestFremdeDatenbankWirdVerweigert` rot |
| 2 leere Datei, eine Transaktion | erfüllt (`BEGIN IMMEDIATE` auf Nulloperation: viele Tests rot; `ROLLBACK` entfernt: rot) |
| 3 zweites Öffnen | erfüllt (Zweig entfernt: rot) |
| 4 neuere Version | erfüllt (Zweig entfernt: rot) |
| 5 fremde Datenbank | erfüllt (Zweig entfernt: rot) |
| 6 NOTADB/CORRUPT gleich `ErrBeschaedigt`, andere Fehler unverändert | erfüllt für NOTADB, CORRUPT (je einzeln entfernt: rot; Wrap entfernt: rot; Klassifizierung aus: rot; "jeder Fehler": rot); CANTOPEN/IOERR: V2-L1 |
| 7 Pfade | erfüllt (Prüfungen der ersten Verifikation unverändert, Code unverändert) |
| 8 gleichzeitiges Öffnen | erfüllt: 2000/2000 und 600/600 (`--cpus=2`); `BEGIN IMMEDIATE` zu `BEGIN`: rot |
| 9 `Close` idempotent | erfüllt |
| 10 `:memory:` | erfüllt |

V-H1 ist behoben: Die Ursache der ersten Verifikation (Umschalten ohne Wiederholung nach dem `COMMIT`) ist durch die Wiederholung bis zum Ablauf des `busy_timeout` ersetzt, und der Versuch ist injizierbar.

## Stufe 4 — SPEC-104 und SPEC-102

Mutationen von `migration-check.sh`, vollständig durch die Stufe `gegenproben` (Docker-Build, `--no-cache-filter gegenproben`):

| Mutation | Ergebnis |
|---|---|
| `diff … \| head -20` zu `:` (V-M3 Rand 2) | **rot**: `migration/abweichung: Ausgabe nennt "< -- zusätzliche Zeile" nicht` |
| `ok` auch bei Abweichung (V-M3 Rand 1) | **rot**: `migration/abweichung: Ausgabe nennt "migration-check: ok"` |
| `cmp`-Zweig entfernt | rot (`abweichung`: Verstoß nicht gemeldet) |
| Dateischleife entfernt | rot (`weitere-datei`) |
| Meldungen "erzeugtes DDL fehlt" und "Migration fehlt" getauscht (V-L2) | **rot**: `erzeugtes-ddl-fehlt: Ausgabe nennt "erzeugtes DDL" nicht` |
| Exit bei Abweichung 1 zu 2 | rot |
| Text `migration-check: ok` geändert | rot (`migration/gleich`) |
| Argumentprüfung `-eq 2` zu `-ge 2` | **grün**: V2-L2 |

`SPEC-102` Fälle 17 bis 23: Namen und Erwartungen im Vertrag stimmen mit den Fällen überein (Verboten-Texte in 18, 21, 22; Feld `Verboten` mit Test `verbotener Text` in `main_test.go`); Anzahl 23.
Die Randform von `make migration-check` (Fehler von `schema-generate`, Aufräumen) bleibt ohne Test, wie in Verifikation 1 (kein neuer Befund; im Vertrag als Ziel-Randform benannt).

## Stufe 5 — ADR-0006 und ADR-0003

- [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md): Treiber `modernc.org/sqlite` v1.60.1 gepinnt, `tech-leak` durch Fall 20 belegt; im Fix-Commit unverändert.
- [ADR-0003](../plan/adr/0003-lint-profil-golangci.md): Lint grün, keine neue Ausnahme in `exclusions.rules`, keine `//nolint`-Direktive im Adapter (`grep` leer); `d-check` 0 Befunde.

## Stufe 6 — Zusage · Fänger (§3, Zeile für Zeile)

| # | Zusage | Mutationen (dieser Lauf) | Urteil |
|---|---|---|---|
| 1 | `foreign_keys=ON` | Pragma entfernt | rot (`TestPragmasJederVerbindung`) |
| 2 | Migration genau einmal | Zweig `version == Schemaversion` weg | rot (`TestZweitesOeffnenVeraendertNichts`) |
| 3 | höhere Version verweigert | Zweig weg | rot (`TestNeuereSchemaversionWirdVerweigert`: Fehler `<nil>`) |
| 4 | `busy_timeout`, `journal_mode` je Verbindung | Wert fest 5000 | rot (`TestBusyTimeoutAusDerOption`) |
| 5 | fremde Datenbank | Zweig weg | rot (`TestFremdeDatenbankWirdVerweigert`) |
| 6 | keine SQLite-Datei gleich beschädigt | Klassifizierung aus; nur NOTADB; nur CORRUPT; Wrap entfernt | rot, aus dem benannten Grund; CANTOPEN-Erweiterung grün (V2-L1) |
| 7 | ungültige Pfade | Code unverändert seit Verifikation 1 | nicht erneut gefahren |
| 8 | gleichzeitiges Öffnen | `BEGIN` statt `BEGIN IMMEDIATE`; `ROLLBACK` entfernt | rot (`TestGleichzeitigesOeffnenMigrierteinmal`) |
| 9 | atomar | `BEGIN IMMEDIATE` zu `SELECT 1` | viele Tests rot, auch `TestMigrationLaeuftInEinerTransaktion` bei `ROLLBACK`-Mutation (Grund wie I2 in Verifikation 1: die echte Mutation bleibt "alle drei zu Nulloperationen") |
| 10 | zweites `Close` | Code unverändert | nicht erneut gefahren |
| 11 | WAL bei jedem Öffnen | Aufruf entfernt | rot (`TestOeffnenStelltWalWiederHer`) |
| 12 | Wiederholung bei BUSY bis zur Frist, nur dann | keine Wiederholung; nie belegt; alle SQLite-Fehler belegt; Frist ignoriert; nur LOCKED | rot (Frist: Zeitüberschreitung des Tests); `return c == busy` (nur BUSY) und Maske entfernt **grün**: V2-M1 |
| 13 | gleichzeitiges Öffnen gelingt dauerhaft | statistisch | 2000/2000, 600/600 (`--cpus=2`) grün |
| 14 | andere Fehler unklassifiziert | "jeder Fehler beschädigt" | rot (`TestAndereFehlerBleibenUnklassifiziert`) |
| 15 | abgelehnte Datei byte-gleich | WAL vor Versionsprüfung | rot |
| 16 | Bibliothek bleibt im Adapter | in Verifikation 1 gefahren, unverändert | nicht erneut gefahren |
| 17/18 | `migration-check.sh` Dateien, Aufruf, Abweichung, weitere Datei | siehe Stufe 4 | rot bis auf V2-L2 |

## Stufe 7 — Plan-vs-Code-Diff (`0355957..4caee05`)

- Im Plan und im Code: Treiber, `Open` mit `stelleWalSicher`, `Migrate`, Migration, `migration-check` mit Vertrag, Fälle 17 bis 23, SPEC-102/103/104.
- Im Code, nicht im Plan (§3): `tools/gegenproben/main.go` (Feld `Verboten`, `letzte` ohne Parameter) fällt unter "tools/gegenproben"; `export_test.go` Testhilfe; `harness/mk/gegenproben.mk`, `harness/vertraege/README.md`, Reviews, Nachbar-Slices, Roadmap als Folgen.
- Im Plan, nicht im Code: nichts. Plan hinkt dem Code: V2-L3 (Testzeile fehlt, Marke), V2-M2 (§3 gegen §1).
- Drift der Wirkung: keiner mehr; "beide behoben" in §6 Risiko 1 ist jetzt belegt.

## Für den Planner

1. Vor `done/`: §1-Ausschluss der Abbildung `ErrBeschaedigt` auf `ODR-E3001` committen, §3 (Zeile 91 und die Zusage-Zeile) auf "Sentinel `ErrBeschaedigt`" nachziehen und den Punkt dort ablegen, wo `welle-record-replay` geplant wird (V2-M2).
2. Zurück an den Implementer: Fänger für `SQLITE_LOCKED` in der Wiederholung (V2-M1), optional für CANTOPEN/IOERR und für ein drittes Argument (V2-L1, V2-L2); §3 um den Nachhol-Test und die Marke (1198 bei `4caee05`, Metrik und Pfadmenge nennen) ergänzen (V2-L3).
3. Randform des Kontext-Abbruchs beim Umstellen entscheiden (V2-L4, `AGENTS.md` §3.12), oder ausdrücklich als nicht zugesagt führen.
4. Closure: Register-Belege (V2-L5), Closure-Notiz, DoD-Häkchen.
