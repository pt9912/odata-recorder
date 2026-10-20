#!/usr/bin/env bash
# gegenproben-eingabe — Prüfsumme der Eingaben von `make gegenproben` (SPEC-102): die Dateien
# der Eingabepfade (getrackt und ungetrackt, .gitignore respektiert) und die Pins der Images,
# die als Argumente folgen. Gibt eine Zeile aus.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

for f in src go.mod .a-check.yml .golangci.yml tools/harness/lint.sh tools/harness/migration-check.sh tools/gegenproben tools/pruefung Dockerfile .dockerignore; do
  [ -e "$f" ] || { echo "gegenproben-eingabe: Eingabe $f fehlt" >&2; exit 1; }
done

{
  printf 'PIN %s\n' "$@"
  git ls-files -z --cached --others --exclude-standard -- \
      src go.mod go.sum .a-check.yml .golangci.yml tools/harness/lint.sh tools/harness/migration-check.sh tools/gegenproben tools/pruefung Dockerfile .dockerignore \
    | sort -zu \
    | while IFS= read -r -d '' f; do
        if [ -f "$f" ]; then sha256sum -- "$f"; fi
      done
} | sha256sum | awk '{print $1}'
