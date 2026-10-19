#!/bin/sh
# tools/harness/migration-check.sh <erzeugtes DDL> <Migrationsverzeichnis> — prüft, dass
# 0001_init.sql im Verzeichnis byte-gleich dem erzeugten DDL ist und keine weitere
# Migration daneben liegt (SPEC-104). Exit 0 gleich, 1 Abweichung, 2 Datei fehlt oder
# falscher Aufruf.
set -u

[ "$#" -eq 2 ] || { echo "Aufruf: migration-check.sh <erzeugtes DDL> <Migrationsverzeichnis>" >&2; exit 2; }
erzeugt=$1
dir=$2
migration=$dir/0001_init.sql

[ -r "$erzeugt" ] || { echo "migration-check: erzeugtes DDL $erzeugt fehlt" >&2; exit 2; }
[ -r "$migration" ] || { echo "migration-check: Migration $migration fehlt" >&2; exit 2; }

rc=0
for f in "$dir"/*.sql; do
  [ "$f" = "$migration" ] && continue
  echo "migration-check: $f ist ungeprüft"
  rc=1
done

if ! cmp -s "$erzeugt" "$migration"; then
  echo "migration-check: 0001_init.sql weicht vom erzeugten DDL ab"
  diff "$erzeugt" "$migration" | head -20
  rc=1
fi

[ "$rc" -ne 0 ] || echo "migration-check: ok"
exit "$rc"
