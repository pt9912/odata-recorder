#!/usr/bin/env bash
# stop-require-gates — das Handoff-Gate am Stop von Claude Code. Nutzt
# dieselbe inhaltsbasierte Hash-Funktion wie record-gates (keine Logik-Dopplung).
#
# DEFAULT — BINDUNG AN DEN COMMIT. Der Hook blockiert nur, wenn BEIDES gilt: HEAD ist
# ein anderer als der, den der letzte gruene `make gates`-Lauf gestempelt hat
# (.harness/state/gates-passed.head), UND der Inhalts-Hash des Arbeitsbaums weicht vom
# Nachweis ab (.harness/state/gates-passed.diffsha). Ein Turn-Ende ohne neuen HEAD geht
# frei, auch mit ungedeckter Aenderung; ein Commit, dessen Inhalt der letzte gruene Lauf
# deckt, ebenso.
#
# STRENG — jedes Turn-Ende mit ungedecktem Inhalt blockiert —, wenn die Datei
# .harness/stop-gate-streng existiert (versioniert, Inhalt bedeutungslos; dieses
# Werkzeug schreibt und loescht sie nie) ODER STOP_GATE_STRENG genau `1` ist. Jeder
# andere Wert der Variablen sagt nichts und hebt die Datei nicht auf. Ohne HEAD-Stempel
# gilt ebenfalls der strenge Zweig, bis ein gruener Lauf ihn schreibt.
#
# FAIL-CLOSED. Jeder unerwartete Fehler endet mit Exit 2 (blockierend), nie mit einem
# anderen Code (Trap unten). Den HEAD-Wert ermittelt head_wert genauso wie
# record-gates.sh — Kopplung: beide Fassungen gleich halten.
#
# GRENZEN. Zugesagt ist "HEAD gleich geblieben", nicht "kein Commit": wer nach einem
# Commit auf die gestempelte SHA zurueckgeht, kommt frei durch. Eine "fertig"-Meldung
# ohne neuen HEAD geht ohne Gate-Lauf durch; das Netz dort ist CI auf dem Push. Ein
# frischer Klon ohne Nachweis und mit sauberem Baum geht frei (kein Nachweis pruefbar).
set -euo pipefail
rc=0
trap 'rc=$?; [ "$rc" -eq 0 ] || exit 2' EXIT
cd "$(git rev-parse --show-toplevel)"

state_file=".harness/state/gates-passed.diffsha"
head_file=".harness/state/gates-passed.head"

approve() {
  printf '%s\n' '{"decision":"approve"}'
  exit 0
}

block() {
  printf '{\n  "decision": "block",\n  "reason": "%s"\n}\n' "$1"
  exit 0
}

# head_wert — die aufgeloeste SHA von HEAD; `kein-commit` nur auf einem ungeborenen
# Zweig in einem Repo ohne einzigen Commit; jede andere Lage ist ein Git-Fehler (Exit 1).
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

# Schleifen-Schutz: Hat dieser Hook den Stop bereits einmal blockiert
# (stop_hook_active), nicht erneut blockieren — sonst Endlosschleife bei
# dauerhaft rotem Gate.
input="$(cat || true)"
if grep -q '"stop_hook_active"[[:space:]]*:[[:space:]]*true' <<<"$input"; then
  approve
fi

current_head="$(head_wert)"

streng=0
if [ -e .harness/stop-gate-streng ]; then streng=1; fi
if [ "${STOP_GATE_STRENG:-}" = 1 ]; then streng=1; fi

if [ "$streng" -eq 0 ] && [ -e "$head_file" ]; then
  recorded_head="$(cat "$head_file")"
  if [ "$current_head" = "$recorded_head" ]; then approve; fi
fi

if [ ! -f "$state_file" ]; then
  status="$(git status --porcelain=v1)"
  # Frischer Klon ohne lokale Aenderungen: kein Nachweis pruefbar.
  if [ -z "$status" ]; then approve; fi
  block "There are working tree changes, but no recorded successful make gates run. Run \`make gates\`."
fi

current="$(bash tools/harness/working-tree-hash.sh)"
recorded="$(cat "$state_file")"

if [ "$current" != "$recorded" ]; then
  block "Repo content is not covered by the last recorded gates run (a new HEAD without a gates run, or strict mode via .harness/stop-gate-streng / STOP_GATE_STRENG=1). Run \`make gates\` again."
fi

approve
