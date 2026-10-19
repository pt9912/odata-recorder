# Review-Report: slice-sqlite-adapter-migration — 2026-10-19

**Review-Art:** Code — gegen Plan, Entscheidungen und Hard Rules.

**Gegenstand:** `slice-sqlite-adapter-migration`, Diff `0355957..ab2671d` (Verträge `SPEC-103` und `SPEC-104` in `4e75712`, Umsetzung `bb6b8ad`, Plan-Nachzug `ab2671d`).
Plan: [slice-sqlite-adapter-migration](../plan/planning/in-progress/slice-sqlite-adapter-migration.md).
Bezug: [LH-FA-03](../../spec/lastenheft.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md).

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

**Läufe (Reviewer selbst):** `make test`, `make lint`, `make a-check`, `make migration-check` grün auf HEAD.
`make gegenproben` meldete "Eingaben unverändert". Mutation von `migration-check.sh` (`cmp`-Zweig und Dateischleife entfernt) färbt `migration/abweichung` und `migration/weitere-datei` rot (19 Fälle, 2 rot).
Zusätzlich Wegwerf-Tests im Adapter-Paket (nach dem Lauf gelöscht) und Wegwerf-Dateien für `a-check`/`lint`; alles per `git checkout` bzw. Löschen zurückgenommen, `harness/gegenproben.stempel` wiederhergestellt, Arbeitsbaum sauber.

**Anzahl:** HIGH 2 · MEDIUM 3 · LOW 3 · INFO 3

---

## HIGH

### H1 — Open verändert die Datei vor jeder Prüfung (WAL-Umschaltung als Verbindungs-Pragma)
- `kategorie`: HIGH
- `quelle`: SPEC-103 Randformen 4 und 5 ("die Datei bleibt unverändert"); Plan §3 Zusage "Höhere Schemaversion / Fremde Datenbank wird verweigert"
- `pfad`: `src/adapters/driven/sqlite/sqlite.go` · `q.Add("_pragma", "journal_mode(wal)")`
- `befund`: `journal_mode(wal)` läuft beim Verbindungsaufbau, also vor der Versions- und Fremd-Prüfung in `migrate`. Gemessen (SHA-256 vor/nach `Open`): eine Datei mit `user_version=2` und eine fremde Datenbank (`CREATE TABLE fremd`) sind nach dem verweigerten Öffnen verändert (Header auf WAL umgestellt). Eine fremde Datenbank wird so durch den Versuch geschrieben. Eine Nicht-SQLite-Datei (Text, 4096 Nullbytes, 100-Byte-Rumpf) bleibt dagegen unverändert.
- `verifizierbar`: ja — `TestNeuereSchemaversionWirdVerweigert` und `TestFremdeDatenbankWirdVerweigert` prüfen nur `user_version`/Tabellen, nicht die Bytes; mit einem Hash-Vergleich werden beide rot.
- `klasse`: Zusage "Datei bleibt unverändert" durch Pragma vor der Prüfung gebrochen

### H2 — Gleichzeitiges Öffnen einer neuen Datei scheitert nichtdeterministisch mit SQLITE_BUSY
- `kategorie`: HIGH
- `quelle`: SPEC-103 Randform 8 ("beide enden mit Erfolg"); Plan §3 Zusage "Gleichzeitiges Öffnen migriert einmal"; Beobachtung `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (Plan §8)
- `pfad`: `src/adapters/driven/sqlite/sqlite_test.go` · `TestGleichzeitigesOeffnenMigrierteinmal`
- `befund`: Im Container (`go test -count=40 -run Gleichzeitig`, zwei Läufe) scheiterten 7 und 11 von 40 Läufen mit `database is locked (5) (SQLITE_BUSY)` in einer der acht Öffnungen. Das Verhalten bricht SPEC-103 Rand 8 für den Aufrufer und macht `make test` zufällig rot. Ursache nicht isoliert; naheliegend ist die WAL-Umschaltung beim Verbindungsaufbau (siehe H1), die den `busy_timeout` nicht zuverlässig durchläuft.
- `verifizierbar`: ja — `make test` wiederholt bzw. der genannte Lauf.
- `klasse`: nebenläufiger Vertrag nur im Glück grün

## MEDIUM

### M1 — Fehlerklassifizierung: Zweig `corrupt` (11) und weitere Codes ohne Fänger
- `kategorie`: MEDIUM
- `quelle`: SPEC-103 Randform 6; Plan §3 Zusage "Keine SQLite-Datei ist `ODR-E3001`"; AGENTS.md §3.15
- `pfad`: `src/adapters/driven/sqlite/sqlite.go` · `const notADB, corrupt = 26, 11`
- `befund`: Der einzige Test schreibt Text und trifft damit nur Code 26; ein abgeschnittener SQLite-Rumpf liefert Code 11 (gemessen), der Zweig ist ungeprüft — die Mutation "`corrupt` entfernt" bleibt grün. Andere Fehler beim Öffnen (z. B. nicht lesbare Datei, `CANTOPEN`, `IOERR`) werden unklassifiziert durchgereicht. Ein `integrity_check` läuft beim Öffnen nicht; eine Datei mit intakter Seite 1 und beschädigten Folgeseiten öffnet ohne Fehler, obwohl Rand 6 "Integrität nicht lesbar" nennt.
- `verifizierbar`: ja — Test mit abgeschnittener Datei.
- `klasse`: Klassifizierung nur zum Teil durch Test gefangen

### M2 — `tech`-Regel greift nicht für die Composition Root; keine Gegenprobe zur Regel
- `kategorie`: MEDIUM
- `quelle`: [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md); AGENTS.md §3.11 und §3.15; SPEC-102 *Grenze*
- `pfad`: `.a-check.yml` · `tech:` und `src/adapters/driven/sqlite/sqlite.go` · `Nur dieses Paket bindet die SQLite-Bibliothek ein.`
- `befund`: Mutation "Import von `modernc.org/sqlite` in `src/bootstrap`": `make a-check` bleibt grün (0 Befunde), weil `composition_root` ausgenommen ist; `make lint` wird nur durch `blank-imports` rot, ein benannter Import wäre grün. Dieselbe Regel meldet `tech-leak` in `src/adapters/driven/andere` und `core-impurity` im Kern (gemessen). Die Zusage im Paketkommentar ist für die Composition Root von keinem Gate gedeckt; `gomodguard_v2` erlaubt das Modul repo-weit. Weder die Tabelle *Zusage · Fänger* noch `SPEC-102` führen einen Fall für `tech-leak`.
- `verifizierbar`: ja — Mutation wie beschrieben; fehlender Gegenprobe-Fall.
- `klasse`: Kopfzusage über das hinaus, was ein Gate fängt

### M3 — `migration-check.sh`: Randformen 3 und 5 sowie die Diff-Kürzung ohne Fall
- `kategorie`: MEDIUM
- `quelle`: AGENTS.md §3.10 (jede Zusage eine Mutation); SPEC-104 Randformen 2, 3, 5
- `pfad`: `tools/gegenproben/faelle.go` · `migration/weitere-datei` (letzter Fall; Fälle 17 bis 19 sind die einzigen)
- `befund`: Exit 2 bei fehlendem erzeugtem DDL, fehlender Migration und falscher Aufrufform sowie `head -20` haben keinen Fall; die Zeilen `[ -r ... ] || exit 2` und die Argumentprüfung können entfallen, ohne dass etwas rot wird. Plan §3 führt sie in der Tabelle nicht.
- `verifizierbar`: ja — Mutation "Prüfzeilen entfernt" ist aus der Fallliste ersichtlich grün.
- `klasse`: Vertragsrandform ohne Fänger

## LOW

### L1 — Kopf und Hilfetext von `gegenproben` nennen `migration-check` nicht
- `kategorie`: LOW · `quelle`: AGENTS.md §3.7, §3.11
- `pfad`: `harness/mk/gegenproben.mk` · `gegenproben: ## Gegenproben zu a-check und lint, nur bei geänderter Eingabe`
- `befund`: Kopfzeile und `##`-Hilfe (`make help`) sagen "a-check und lint"; `harness/README.md` und `SPEC-102` nennen inzwischen auch `migration-check.sh`.
- `verifizierbar`: nein — kein Gate.
- `klasse`: Kommentar hinkt hinter der Zusage

### L2 — §6 des Plans nennt die Randformen des Slice nicht
- `kategorie`: LOW · `quelle`: AGENTS.md §3.12, §3.9
- `pfad`: Plan · `Randformen der Repositories (`AGENTS.md` §3.12)`
- `befund`: Der Punkt behandelt Repository-Randformen (doppelter Schlüssel, Transaktionsabbruch), die dieser Slice nicht liefert; die zehn Randformen von `SPEC-103` und die fünf von `SPEC-104` stehen in §6 nicht je Randform mit Entscheidungsort. `Berührte Spec-Stellen` im Kopf nennen `SPEC-103`/`SPEC-104` nicht.
- `verifizierbar`: nein
- `klasse`: Plan §6 folgt der Korrektur nicht

### L3 — Marke nicht in der benannten Metrik reproduzierbar; Rückführung "über 400" überschritten
- `kategorie`: LOW · `quelle`: AGENTS.md §3.14
- `pfad`: Plan · `liegt bei 783 Zeilen`
- `befund`: Die Summe aus `git diff --numstat 0355957..HEAD` über die Pfade der §3-Tabelle ergibt bei mir 868, nicht 783. Die Rückführungsmarke "über 400" bleibt neben der Notiz "Marke überschritten, bleibt ungeteilt" bestehen.
- `verifizierbar`: ja — numstat-Lauf.
- `klasse`: Marken-Metrik nicht nachrechenbar

## INFO

- **I1** — `ErrBeschaedigt` ist ein Sentinel; die Abbildung auf den Code `ODR-E3001` ([SPEC-018](../../spec/spezifikation.md)) liegt beim Aufrufer und ist noch nicht gebaut. Zuständig: Planner/Folge-Slice `slice-session-recording-persistenz`, nicht an diesem Diff zu messen.
- **I2** — `make a-check` meldet weiter, dass `tools/gegenproben/*.go` in keiner Schicht liegen (nicht aus diesem Diff).
- **I3** — Ressourcen: `Open` schließt die `sql.DB` bei jedem Migrationsfehler, `migrate` gibt `Conn` per `defer` zurück und rollt vor dem Zurückgeben zurück (`context.WithoutCancel`); abgebrochener Kontext legt keine Datei an. Pfade mit `?`, `#`, `%20`, `&_pragma=…` im Namen werden korrekt als Dateiname behandelt (gemessen). Kein Befund.

## Geprüft, ohne Befund
- `src/adapters/driven/sqlite/migrate.go` (BEGIN IMMEDIATE, ROLLBACK-Pfade, `done`-Flag, Conn-Rückgabe): ohne Befund außer H1/H2 im Zusammenspiel mit dem Pragma.
- `src/adapters/driven/sqlite/migrations/`: gleich dem erzeugten DDL (`make migration-check` grün).
- `go.mod`, `go.sum`, `.golangci.yml` (`gomodguard_v2`-Eintrag, kein `//nolint`, `# Why:`-Regeln unberührt): ohne Befund.
- `Dockerfile` (Stufe `gegenproben`: Modulcache aus `deps`, `--network=none`), `.dockerignore`, `tools/harness/gegenproben-eingabe.sh`: ohne Befund; die neue Eingabe geht in die Prüfsumme ein (Stempel rot nach Mutation).
- `harness/mk/schema.mk`, `tools/harness/migration-check.sh`: ohne Befund außer M3.
- `tools/gegenproben` (`gateMigration`, `Kopiere`, `OhneGoDateienInSrc`): ohne Befund; Fälle 14 bis 16 bleiben durch die neue Löschform erfüllt.
- `spec/spezifikation.md` SPEC-103, `harness/vertraege/`: widerspruchsfrei zu Lastenheft, Architektur ohne Slice- oder ADR-Bezug.
