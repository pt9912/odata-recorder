#!/usr/bin/env bash
# record-gates — Nachweis schreiben, dass `make gates` den aktuellen
# Arbeitsbaum-Zustand abgedeckt hat. Laeuft als LETZTER gates-Prerequisite (nur
# bei gruenen Gates). Der Stop-Hook vergleicht denselben Hash und den HEAD-Stempel
# darunter — ein neuer HEAD ohne frischen Gate-Lauf ueber seinem Inhalt laesst ihn rot.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# HEAD-STEMPEL: neben dem Hash schreibt dieses Skript die
# aufgeloeste Commit-SHA von HEAD nach .harness/state/gates-passed.head — der Stop-Hook
# gibt ein Turn-Ende ohne neuen HEAD damit frei. Format von gates-passed.diffsha bleibt,
# wie es ist (weitere Leser). head_wert ermittelt den Wert genauso wie der Stop-Hook —
# Kopplung: beide Fassungen gleich halten. Ist HEAD nicht aufloesbar und das Repo nicht
# eng als commitlos erkannt, endet das Skript rot, bevor es einen Stempel schreibt.
head_wert() {
  local sha alle
  if sha="$(git rev-parse --verify -q HEAD)"; then
    printf '%s\n' "$sha"
    return 0
  fi
  git symbolic-ref -q HEAD >/dev/null || return 1
  alle="$(git rev-list -n1 --all)" || return 1
  [ -z "$alle" ] || return 1
  printf '%s\n' kein-commit
}

if ! head="$(head_wert)"; then
  echo "record-gates: FEHLER — HEAD ist nicht aufloesbar, und das Repo ist nicht als commitlos erkannt (ungeborener Zweig ohne einzigen Commit) — kein Stempel geschrieben." >&2
  exit 1
fi

mkdir -p .harness/state
bash tools/harness/working-tree-hash.sh > .harness/state/gates-passed.diffsha
printf '%s\n' "$head" > .harness/state/gates-passed.head
