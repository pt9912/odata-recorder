# Verifikations-Report: slice-sqlite-adapter-migration — 2026-10-19

**Art:** Verifikation (Modul 11) — gegen DoD, Closure-Trigger, Spezifikation und Entscheidungen. Nicht Review, nicht Validierung.

**Gegenstand:** [slice-sqlite-adapter-migration](../plan/planning/done/slice-sqlite-adapter-migration.md), Diff `0355957` (Move nach `in-progress/`) bis `d015e82` (HEAD).
Bezug: [LH-FA-03](../../spec/lastenheft.md), [LH-QA-06](../../spec/lastenheft.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md).
Vorgänger: [Review-Report](2026-10-19-slice-sqlite-adapter-migration-review.md) (2 HIGH, 3 MEDIUM, 3 LOW).

**Rolle:** Verifier · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

## Urteil: NICHT BESTANDEN — zurück an den Implementer

Der Slice ist nicht closure-fähig.
Ein Befund des Reviews, H2 (gleichzeitiges Öffnen scheitert mit `SQLITE_BUSY`), ist **nicht behoben**: Der Plan §6 behauptet "beide behoben", gemessen flackert der Test weiter, und die Ursache ist jetzt lokalisiert (V-H1).
H1 (Byte-Gleichheit abgelehnter Dateien) ist behoben und mit zehn Fixtures belegt.
Alle Gates sind grün; die übrigen Befunde sind MEDIUM und LOW.

**Anzahl:** HIGH 1 · MEDIUM 3 · LOW 3 · INFO 3.

## Läufe (Verifier selbst)

| Lauf | Ergebnis |
|---|---|
| `make gates` (zweimal, HEAD) | beide Male Exit 0: a-check 0 Befunde, d-check 0 Befunde, Build, Test (Adapter-Paket `ok`), Lint 0 issues, `schema-check`, `migration-check: ok` |
| `make gegenproben` bei vorhandenem Stempel | "Eingaben unverändert", nichts läuft |
| `rm -f harness/gegenproben.stempel; make gegenproben` (erzwungen) | Exit 0, `gegenproben: 23 Fälle, 0 rot`; der neu geschriebene Stempel ist byte-gleich dem versionierten (`ee17d7bc…`) |
| Probe aus §5: eine Zeile an `0001_init.sql` angehängt, `make gates` | Exit 2, `migration-check: 0001_init.sql weicht vom erzeugten DDL ab`, `make: *** [harness/mk/schema.mk:40: migration-check] Fehler 1`; danach `git checkout` der Datei |
| `go test -count=40 -run Gleichzeit` im Container (`golang:1.27-alpine@sha256:8a5910f3…`, Netz für die Module), dreimal | 2 von 3 Läufen grün, 1 rot: `Öffnung 7: database is locked (5) (SQLITE_BUSY)` |
| Wiederholung: 10 Läufe à 40 ohne Begrenzung / 10 mit `--cpus=1` / 10 mit `--cpus=2` | 0 / 0 / 1 rot |
| Wiederholung mit Fehlerstellen-Markierung im Wegwerf-Klon, `--cpus=2`, 120 Läufe à 40 | 2 rot, **beide an `PRAGMA journal_mode = wal` nach `COMMIT`** |

Arbeitsbaum am Ende sauber (`git status --short` leer, `harness/gegenproben.stempel` per `git checkout` wiederhergestellt).
Alle Mutationen liefen in einem Wegwerf-Klon (`git archive HEAD`) außerhalb des Repos.

## Findings

### HIGH

#### V-H1 — H2 nicht behoben: das WAL-Umschalten scheitert unter Gleichzeitigkeit mit `SQLITE_BUSY`
- `stufe`: Review-Befund behoben · Plan §6 · DoD 3 · SPEC-103 Randformen 1 und 8
- `pfad`: `src/adapters/driven/sqlite/migrate.go` · `PRAGMA journal_mode = wal` nach `COMMIT`
- `befund`: Die Zusage "beide enden mit Erfolg" (SPEC-103 Rand 8) bricht weiter.
  Gemessen siehe oben: ein `-count=40`-Lauf von drei war rot, in der markierten Messung 2 von 120.
  Beide Fehler stammen von der Stelle, die der letzte Commit eingeführt hat: Der Opener, der migriert hat, schaltet nach dem `COMMIT` auf WAL, hält dabei keine Transaktion mehr, und eine zweite Öffnung hält eine Lesesperre; das Umschalten braucht eine exklusive Sperre und liefert `SQLITE_BUSY`, ohne dass der `busy_timeout` greift.
  Die Rate ist von etwa 20 bis 25 % je 40er-Lauf (Review) auf etwa 1 bis 2 % gefallen, nicht auf null.
- `Folge` (aus dem Code gelesen, nicht gemessen): Scheitert das Umschalten, liefert `Open` einen Fehler, obwohl die Datei migriert ist (`user_version=1`).
  Jedes spätere `Open` sieht `version == Schemaversion` und kehrt zurück; die Datei bleibt dauerhaft im Journal-Modus `delete`, ohne dass ein Test oder ein Fehler es meldet.
  Das bricht SPEC-103 Rand 1 ("eine neu migrierte Datei wechselt in WAL") still.
- `Plan`: §6 Risiko 1 trägt den Ausgang "beide behoben" und DoD 3 verlangt getestete Randformen; beides ist so nicht belegt.
- `Fänger fehlt`: Kein Test deckt den Fall "Umschalten scheitert".
  Der Gleichzeitigkeits-Test fängt ihn nur im Glück (`BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen`).
- `verifizierbar`: ja — `go test -count=40 -run Gleichzeit ./src/adapters/driven/sqlite/`, mehrfach.

### MEDIUM

#### V-M1 — SPEC-103 Randform 6, zweiter Satz, ohne Fänger
- `befund`: "Andere Fehler der Datei (etwa `SQLITE_CANTOPEN`, `SQLITE_IOERR`) gibt Öffnen unverändert weiter" hat keinen Test.
  Mutation `|| true` in `klassifiziere` (jeder Treiberfehler wird `ErrBeschaedigt`): `go test` bleibt grün (`ok`).
- `verifizierbar`: ja.

#### V-M2 — SPEC-103 Randform 6 sagt `ODR-E3001`, der Adapter liefert nur einen Sentinel
- `befund`: Die Spezifikation sagt "Öffnen scheitert mit `ODR-E3001` ([SPEC-018](../../spec/spezifikation.md))".
  Der Code liefert `ErrBeschaedigt`; die Zeichenkette `ODR-E3001` kommt nur im Kommentar vor, die Abbildung ist nirgends gebaut oder getestet.
  Der Plan §1 weist die Abbildung keinem Folge-Slice zu (`AGENTS.md` §3.13: eine Zusage braucht eine Adresse, die sie annimmt).
  Das Review führte das als INFO I1.
  Als Verifikation gegen die Spec ist es eine Lücke: Entweder schärft die Spec ("Sentinel, Abbildung beim Aufrufer") oder der Slice nennt den Nehmer.
- `verifizierbar`: ja — `grep -rn E3001 src` trifft nur den Kommentar in `sqlite.go`.

#### V-M3 — SPEC-104 Randformen 1 und 2 ohne Fänger
- `befund`: Zwei Mutationen von `migration-check.sh`, vollständig durch die Stufe `gegenproben` gefahren (Docker-Build, `--no-cache-filter gegenproben`), bleiben grün: `gegenproben: 23 Fälle, 0 rot`.
  - `diff … | head -20` ersetzt durch `:` — Zusage Rand 2: "die ersten Zeilen von `diff`".
  - `[ "$rc" -ne 0 ] || echo "migration-check: ok"` ersetzt durch `echo "migration-check: ok"` — Zusage Rand 1: "keine Ausgabe außer einer Zeile `migration-check: ok`"; das Gate meldet nun "ok" **neben** einer Abweichung.
  Der Fall `migration/abweichung` prüft mit `WantText` nur, dass ein Text vorkommt, und kann ein Fehlen nicht fordern.
- `Muster`: Das ist die Klasse von `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15); der Slice liefert einen weiteren Beleg, der in `evidence/` gehört.
- `verifizierbar`: ja.

### LOW

#### V-L1 — Der Plan folgt der Korrektur nicht überall (`AGENTS.md` §3.9)
- Kopf `Berührte Spec-Stellen` nennt `SPEC-004`, `SPEC-014`, `ARC-009`, nicht `SPEC-103`, `SPEC-104`, `SPEC-102` (Review L2, nicht behoben).
- §3, Zeile `tools/gegenproben/…`: "Fälle 17 bis 19"; der Vertrag führt 17 bis 23.
- §4 "Marke überschritten": 875 Zeilen. `git diff --numstat 0355957..HEAD` über die Pfade der §3-Tabelle (einschließlich `go.sum`, `SPEC-102`, `harness/README.md`, `spec/spezifikation.md`) ergibt 982 (953 Einfügungen, 29 Löschungen). Die Zahl ist nicht reproduzierbar; die Metrik ist benannt, die Pfadmenge nicht eindeutig.
- §6 Risiko 1 "beide behoben" (siehe V-H1); Risiko 2 trägt "weiter offen", obwohl der Treiber in der Stufe `deps` baut und alle Gates grün sind (Ausgang wäre *entfallen*, oder die Beobachtung ist anzulegen).

#### V-L2 — Gegenprobe-Fälle 21 und 22 unterscheiden die fehlende Datei nicht
- `WantText: "fehlt"` steht in beiden Fällen; ein Tausch der beiden Meldungstexte bliebe grün (Schluss aus dem Fall-Code, nicht als Mutation gefahren).

#### V-L3 — Kein Beobachtungs-Eintrag aus diesem Slice
- DoD und Closure-Notiz stehen aus (erwartet vor der Closure), aber V-H1 und V-M3 sind Belege für `nebenlaeufiger-lauf-nur-im-glueck-gruen` und `vertrag-sagt-mehr-zu-als-ein-gate-prueft`; die `evidence/`-Dateien führt der Slice noch nicht.

### INFO
- **I1** — Eine Datei mit intaktem Kopf und zerstörtem Rest (ab Byte 100 mit `0xAB` überschrieben) öffnet ohne Fehler. Das folgt aus SPEC-103 Rand 6 ("keine `integrity_check`-Prüfung beim Öffnen"), ist also spec-konform.
- **I2** — Die Mutation "ohne `BEGIN`" (nur `BEGIN IMMEDIATE` entfernt) wird von vielen Tests rot, aber aus dem falschen Grund (`cannot commit - no transaction is active`). Der Fänger der Atomarität (`TestMigrationLaeuftInEinerTransaktion`) wird erst mit der echten Mutation (`BEGIN`, `COMMIT`, `ROLLBACK` zu Nulloperationen) aus dem benannten Grund rot: `Teile der Migration bleiben: [a]`.
- **I3** — Die `tech`-Regel gilt nicht für die Composition Root (`composition_root` ist ausgenommen). Der Paketkommentar sagt das jetzt ehrlich; es ist eine Grenze, kein Befund.

## Stufe 1 — DoD-Punkte

| DoD | Urteil | Beleg |
|---|---|---|
| 1 Treiber gepinnt, `tech`-Regel, `gomodguard_v2`, `Open` setzt `foreign_keys`, `busy_timeout`, Journal-Modus, wendet Migration an | erfüllt, mit V-H1 | `go.mod`: `modernc.org/sqlite v1.60.1`, `libc v1.77.1`; `go.sum` +50 Zeilen; `.a-check.yml` `tech`-Eintrag; `.golangci.yml` `allowed: - module: modernc.org/sqlite`; `Open` setzt die Pragmas über die DSN, WAL nach der Migration |
| 2 Gate `migration-check` an `GATE_CHECKS`, byte-gleich, Unterschied färbt rot | erfüllt | `schema.mk` `GATE_CHECKS += schema-check migration-check`; Probe rot (Exit 2), Fälle `abweichung`, `weitere-datei` rot unter Mutation |
| 3 Tests: leere Datei, `foreign_key_check`/`integrity_check` leer, zweites Öffnen, Randformen | teilweise | Die drei ersten Punkte erfüllt (`TestNeueDateiWirdMigriert`, `TestZweitesOeffnenVeraendertNichts`). Randformen: V-H1 (Rand 1, 8), V-M1 und V-M2 (Rand 6) |
| 4 `make gates` grün | erfüllt, mit Vorbehalt | zweimal Exit 0; der Vorbehalt ist der flackernde Test aus V-H1, der das Gate zufällig rot färben kann |

Die übrigen DoD-Zeilen (Review-Report liegt vor: ja; Doku-Update: `SPEC-103`, `SPEC-104`, `SPEC-102`, `harness/README.md` nachgezogen; Closure-Notiz, Register, Risiko-Ausgänge, Paarungen) sind vor der Closure offen und nicht Gegenstand dieses Urteils, außer wie in V-L1 und V-L3 benannt.

## Stufe 2 — Closure-Trigger (§5)

| Kriterium | Urteil |
|---|---|
| DoD vollständig | nein (V-H1; Closure-Pflichten stehen aus) |
| `make gates` grün mit `migration-check` in der Kette | ja (zweimal) |
| geänderte Migration färbt `make gates` rot, vor der Rücknahme | ja (Probe oben) |
| `make gegenproben` grün | ja (erzwungen, 23 Fälle, 0 rot) |
| Closure-Notiz mit Lerneintrag | steht aus |

## Stufe 3 — SPEC-103 Randformen 1 bis 10

| Rand | Zusage | Urteil |
|---|---|---|
| 1 | Pragmas je Verbindung, WAL nach Migration, abgelehnte Datei byte-gleich | Pragmas und Byte-Gleichheit erfüllt; WAL-Umschalten unter Gleichzeitigkeit nicht (V-H1) |
| 2 | leere Datei migriert in einer Transaktion | erfüllt |
| 3 | zweites Öffnen ändert nichts | erfüllt |
| 4 | neuere Version verweigert, nennt beide Versionen | erfüllt |
| 5 | fremde Datenbank verweigert | erfüllt |
| 6 | NOTADB/CORRUPT gleich `ODR-E3001`; andere Fehler unverändert | Klassifizierung erfüllt; V-M1, V-M2 |
| 7 | leerer Pfad, Verzeichnis, fehlendes Verzeichnis; kein Anlegen | erfüllt |
| 8 | gleichzeitige Öffnung, genau eine migriert, beide Erfolg | **nicht erfüllt** (V-H1) |
| 9 | `Close` idempotent | erfüllt |
| 10 | `:memory:` nicht zugelassen | erfüllt |

### Byte-Gleichheit abgelehnter Dateien (H1) — eigene Fixtures

Wegwerf-Test im Klon: SHA-256 aller Dateien im Verzeichnis (Datei und Nebendateien) vor und nach `Open`.

| Fixture | `Open` | Verzeichnis unverändert |
|---|---|---|
| leere Datei mit `user_version=2` | `ErrSchemaNeuer` (Datei 2, bekannt 1) | ja |
| WAL-Datei mit Tabelle und `user_version=2` | `ErrSchemaNeuer` | ja |
| Tabellen, `user_version=3` | `ErrSchemaNeuer` | ja |
| Datei mit Tabelle `fremd` und einer Zeile | `ErrFremdeDatenbank` | ja |
| migrierte Datei auf 100 Byte gekürzt | `ErrBeschaedigt` (Code 11) | ja |
| auf 4096 Byte gekürzt | `ErrBeschaedigt` (Code 11) | ja |
| auf halbe Länge gekürzt | `ErrBeschaedigt` (Code 11) | ja |
| Text | `ErrBeschaedigt` (Code 26) | ja |
| 4096 Nullbytes | `ErrBeschaedigt` (Code 26) | ja |
| Kopf intakt, Rest mit `0xAB` | kein Fehler (I1) | ja |

H1 ist behoben.

## Stufe 4 — SPEC-104 und SPEC-102

- `SPEC-104` Randformen 2 bis 5: Meldungen, Exit-Codes und die Schleife stimmen mit `tools/harness/migration-check.sh` überein und sind durch die Fälle 17 bis 23 gefangen (Mutationen unten).
  Randformen 1 und 2 in ihrer vollen Zusage: V-M3.
  Die Randform von `make migration-check` (Fehler von `schema-generate`, Aufräumen des temporären Verzeichnisses): `trap … EXIT` und `&&`-Kette im Ziel vorhanden, kein Test.
- `SPEC-102` Fälle 17 bis 23: Namen, Gates, Eingriffe und Erwartungen im Vertrag stimmen mit `tools/gegenproben/faelle.go` überein; Anzahl 23, Namen eindeutig; erzwungener Lauf grün.
- `Stand` im Vertrag: "vorhanden für die Migration 0001" stimmt.

## Stufe 5 — ADR-0006 und ADR-0003

- [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md): Bibliothek `modernc.org/sqlite` v1.60.1 mit `libc`, `mathutil`, `memory` gepinnt; Folgepflichten (`go.mod`/`go.sum`, `tech`-Regel, `gomodguard_v2`) erfüllt; Fitness Function `tech-leak` durch Fall 20 belegt (Mutation: Regel entfernt, Fall rot).
- [ADR-0003](../plan/adr/0003-lint-profil-golangci.md): kein `//nolint` im Code (die Treffer in `faelle.go` sind per Konkatenation zusammengesetzte Probe-Texte); keine neue Ausnahme in `exclusions.rules`; `make lint` 0 issues; der Eintrag in `gomodguard_v2` trägt die ADR im Kommentar.

## Stufe 6 — Zusage · Fänger (§3, Zeile für Zeile)

Jede Mutation im Wegwerf-Klon; "rot" heißt: aus dem benannten Grund rot, sofern nicht anders vermerkt.

| # | Zusage | Mutationen | Rot in | Urteil |
|---|---|---|---|---|
| 1 | `foreign_keys=ON` | Pragma entfernt; `foreign_keys(0)` | `TestPragmasJederVerbindung` | ok |
| 2 | Migration genau einmal | `case version == Schemaversion` entfernt; `user_version` nicht gesetzt | `TestZweitesOeffnenVeraendertNichts` (+ weitere) | ok |
| 3 | höhere Version verweigert | Zweig entfernt; Schwelle `+5` | `TestNeuereSchemaversionWirdVerweigert` | ok |
| 4 | `busy_timeout`, `journal_mode` | Wert fest 5000; Pragma entfernt; WAL gegen `delete` | `TestBusyTimeoutAusDerOption`, `TestPragmasJederVerbindung` | ok (getestet an zwei gepoolten Verbindungen) |
| 5 | fremde Datenbank | Schwelle `> 1000`; Block entfernt | `TestFremdeDatenbankWirdVerweigert` | ok |
| 6 | keine SQLite-Datei gleich beschädigt | Klassifizierung aus; `corrupt` entfernt; `notADB` entfernt | `TestKeineSQLiteDateiIstBeschaedigt` (26), `TestAbgeschnittenerRumpfIstBeschaedigt` (11) | ok; die Gegenrichtung (`|| true`) bleibt grün: V-M1 |
| 7 | ungültige Pfade | leer; `:memory:`; Verzeichnis; Elternverzeichnis je entfernt | `TestUngueltigePfade` | ok |
| 8 | Gleichzeitigkeit | `BEGIN IMMEDIATE` gegen `BEGIN` (6 von 6 Läufen rot); `COMMIT` entfernt | `TestGleichzeitigesOeffnenMigrierteinmal` | die Mutationen fängt er; der Test selbst flackert: V-H1 |
| 9 | atomar | `BEGIN`/`COMMIT`/`ROLLBACK` zu Nulloperationen | `TestMigrationLaeuftInEinerTransaktion`: "Teile der Migration bleiben: [a]" | ok (siehe I2) |
| 10 | zweites `Close` | zweites `Close` liefert Fehler | `TestCloseIstIdempotent` | ok |
| 11 | abgelehnte Datei byte-gleich | WAL als DSN-Pragma (der alte Fehler H1); Schreiben vor der Versions-Ablehnung; Schreiben vor der Fremd-Ablehnung; Schreiben direkt nach dem Öffnen | `TestNeuereSchemaversionWirdVerweigert`, `TestFremdeDatenbankWirdVerweigert` | ok |
| 12 | Bibliothek bleibt im Adapter | `tech`-Zeile aus `.a-check.yml` entfernt | Fall `a-check/bibliothek-in-fremdem-adapter`: "Verstoß nicht gemeldet" | ok (Composition Root: I3) |
| 13 | fehlende Dateien, falscher Aufruf | Prüfzeile erzeugtes DDL entfernt; Prüfzeile Migration entfernt; Argumentprüfung entfernt | `migration/erzeugtes-ddl-fehlt` ("erwartet Exit 2, war 1"), `migration/migration-fehlt` (dito), `migration/falsche-aufrufform` ("Aufruf" fehlt) | ok (V-L2) |
| 14 | Abweichung, weitere Datei | `cmp`-Zweig entfernt; Dateischleife entfernt | `migration/abweichung`, `migration/weitere-datei`: "Verstoß nicht gemeldet" | ok |

Zusagen des Vertrags ohne Zeile in dieser Tabelle und ohne Fänger: V-M1, V-M3 (und die Randform von `make migration-check` aus Stufe 4).

## Stufe 7 — Plan-vs-Code-Diff (`0355957..HEAD`)

- Im Plan, im Code: Treiber, `Open`, `Migrate`, eingebettete Migration, `migration-check` mit Vertrag, Fälle 17 bis 23, `SPEC-103`, `SPEC-104`. Die §3-Dateien tragen alle Änderungen.
- Im Code, nicht im Plan: `export_test.go` (Testzugang `SQL()`/`MigrateWith`) — Testhilfe, unkritisch; Erweiterung der Gegenproben-Fallstruktur (`Kopiere`, `Argumente`, `OhneGoDateienInSrc`) im Plan nur als "Fälle" geführt; `docs/plan/planning/in-progress/roadmap.md` (Zeile "Nichts in Arbeit" entfernt) und die zwei Nachbar-Slices (Verweise nach dem Move) — Folge von `slice-mv`.
- Im Plan, nicht im Code: keiner.
- Plan hinkt dem Code: V-L1.
- Drift der Wirkung: Der Plan (§6) meldet einen Befund als behoben, der es nicht ist: V-H1.

## Für den Planner

1. Zurück an den Implementer: V-H1 (Umschalten auf WAL so führen, dass es unter Gleichzeitigkeit nicht scheitert, oder den Fehler entscheiden — SPEC-103 sagt nichts über ein gescheitertes Umschalten; das ist eine Randform für den Architect, `AGENTS.md` §3.12), V-M1, V-M3 mit Fängern.
2. An den Architect/Planner: V-M2 (Sentinel oder `ODR-E3001`; wenn Folge-Slice, dann dessen §1 oder DoD vor der Nennung, `AGENTS.md` §3.13).
3. Plan nachziehen: V-L1.
4. Nach der Behebung erneut verifizieren: `go test -count=40 -run Gleichzeit` über mindestens 50 Läufe à 40 muss durchgehend grün sein; ohne die Fehlerstelle zu kennen, ist ein einzelner grüner Lauf kein Beleg.
