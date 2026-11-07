#!/usr/bin/env bash
# gegenproben-eingabe — Prüfsumme der Eingaben von `make gegenproben` (SPEC-102): die Dateien
# der Eingabepfade (getrackt und ungetrackt, .gitignore respektiert) und die Pins der Images,
# die als Argumente folgen. Gibt eine Zeile aus.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# Drei Pins (Lint-Image, a-check, d-check), jeder nicht leer: ein fehlendes oder leeres Argument
# im Aufruf von make gegenproben ließe die Prüfsumme bei einem Pin-Wechsel gleich.
[ "$#" -eq 3 ] || { echo "gegenproben-eingabe: 3 Pins erwartet, $# übergeben" >&2; exit 1; }
for p in "$@"; do [ -n "$p" ] || { echo "gegenproben-eingabe: leerer Pin" >&2; exit 1; }; done

# Die Eingabepfade; go.sum darf fehlen, alle übrigen müssen existieren.
EINGABEN=(src go.mod go.sum .a-check.yml .d-check.yml .golangci.yml tools/harness/lint.sh tools/harness/migration-check.sh tools/harness/groessenmarke.sh tools/harness/zeichen.sh tools/gegenproben tools/pruefung tools/groessenmarke tools/zeichen tools/katalog Dockerfile .dockerignore)

for f in "${EINGABEN[@]}"; do
  [ "$f" = go.sum ] && continue
  [ -e "$f" ] || { echo "gegenproben-eingabe: Eingabe $f fehlt" >&2; exit 1; }
done

{
  printf 'PIN %s\n' "$@"
  git ls-files -z --cached --others --exclude-standard -- "${EINGABEN[@]}" \
    | sort -zu \
    | while IFS= read -r -d '' f; do
        if [ -f "$f" ]; then sha256sum -- "$f"; fi
      done
} | sha256sum | awk '{print $1}'
