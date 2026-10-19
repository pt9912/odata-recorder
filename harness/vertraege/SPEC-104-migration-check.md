# SPEC-104 — Migrations-Gate (`make migration-check`)

**Werkzeug:** `make migration-check`; Prüfung `tools/harness/migration-check.sh`.

**Festlegung:** Die eingebettete Migration `src/adapters/driven/sqlite/migrations/0001_init.sql` ist byte-gleich dem DDL, das `make schema-generate` aus `tools/schema/schema.yaml` erzeugt (`SPEC-004`).
Das Ziel erzeugt das DDL in einem temporären Verzeichnis und ruft `migration-check.sh <erzeugtes DDL> <Migrationsverzeichnis>`.
Die Kopfzeilen des erzeugten DDL (Version des Schemas) gehören zum Vergleich: eine neue Schemaversion verlangt eine neue Migration.

**Randformen von `migration-check.sh`:**
1. Gleich: Exit-Code 0, keine Ausgabe außer einer Zeile `migration-check: ok`.
2. Ungleich: Exit-Code 1, Meldung `migration-check: 0001_init.sql weicht vom erzeugten DDL ab` und die ersten Zeilen von `diff`.
3. Die Migration fehlt oder das erzeugte DDL fehlt: Exit-Code 2 mit Meldung, welche Datei fehlt.
4. Im Migrationsverzeichnis liegt eine weitere Datei `*.sql` neben `0001_init.sql`: Exit-Code 1 mit Meldung `migration-check: <Datei> ist ungeprüft`; die Kette über mehrere Versionen prüft das Ziel nicht.
5. Die Aufrufform stimmt nicht (weniger oder mehr als zwei Argumente): Exit-Code 2 mit Aufrufhilfe.

**Randformen von `make migration-check`:** Scheitert `schema-generate`, scheitert das Ziel mit dessen Fehler; das temporäre Verzeichnis wird in jedem Fall entfernt.

**Stand:** vorhanden für die Migration `0001`.

**Grenze:** Das Ziel belegt nicht, dass die Migration auf bestehenden Datenbanken läuft; die Migrationskette entsteht mit der zweiten Schemaversion.
