#!/usr/bin/env bash
# gegenproben-eingabe — Prüfsumme der Eingaben von `make gegenproben` (SPEC-102): die Dateien
# der Eingabepfade (getrackt und ungetrackt, .gitignore respektiert) und die Pins der Images,
# die als Argumente folgen. Gibt eine Zeile aus.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# Die Eingabepfade; go.sum darf fehlen, alle übrigen müssen existieren.
EINGABEN=(src go.mod go.sum .a-check.yml .golangci.yml tools/harness/lint.sh tools/harness/migration-check.sh tools/gegenproben tools/pruefung Dockerfile .dockerignore)

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
