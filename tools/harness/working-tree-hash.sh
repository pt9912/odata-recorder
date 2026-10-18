#!/usr/bin/env bash
# working-tree-hash — inhaltsbasierter Hash ueber alle getrackten UND untracked
# Dateien (exclude-standard, respektiert .gitignore). Gemeinsame Quelle der
# Wahrheit fuer record-gates und den Stop-Hook — keine Logik-Dopplung.
#
# INHALTSBASIERT statt diff-basiert: ein Commit ohne Inhaltsaenderung laesst den
# Hash unveraendert — der Gate-Nachweis bleibt gueltig. Ein Commit OHNE
# vorherigen Gate-Lauf erzeugt dagegen keinen passenden Nachweis und wird vom
# Stop-Hook nicht mehr durchgewunken.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

git ls-files -z --cached --others --exclude-standard \
  | sort -zu \
  | while IFS= read -r -d '' f; do
      if [ -L "$f" ]; then
        printf 'LINK %s -> %s\n' "$f" "$(readlink -- "$f")"
      elif [ -f "$f" ]; then
        sha256sum -- "$f"
      else
        printf 'GONE %s\n' "$f"
      fi
    done | sha256sum | awk '{print $1}'
