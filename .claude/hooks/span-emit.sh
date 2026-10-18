#!/usr/bin/env bash
# span-emit — der Hook-Wrapper der Erfassungsschicht.
#
# WARUM EIN WRAPPER UND NICHT DER TRAEGER DIREKT. Der Traeger liegt im gitignorierten
# Zustands-Bereich: ein frischer Klon dieses Repos hat ihn nicht, und ein Aufraeum-Lauf
# kann ihn entfernen. Zeigte die Hook-Konfiguration direkt auf ihn, waere genau das ein
# Hook, der auf ein fehlendes Programm zeigt — nur zeitversetzt. Diese Datei
# ist committet, also auf jedem Checkout da. Fehlt der Traeger, wird nichts erfasst; er
# kommt zurueck, sobald das Werkzeug erneut laeuft.
#
# DIESER WRAPPER HAELT KEINEN TOOL-CALL AUF. Er endet in jedem Zweig mit 0 und schreibt
# selbst nichts auf stdout — dort liegt bei Hooks der ENTSCHEIDUNGS-Kanal, und wer dort
# schreibt, entscheidet ueber Berechtigungen mit, statt zu beobachten.
# Ein Fehlschlag des Traegers wird hier ein zweites Mal auf 0 geklemmt;
# dessen eigene Klemme bleibt davon unberuehrt.
#
# ZWEI NAMEN, EIN ORT: der Bootstrap legt das laufende Bild unter festem Namen ab und
# nimmt dessen Endung mit — auf Windows `.exe`, sonst keine. Beide Namen
# werden gesucht, damit dieselbe Datei auf jeder Plattform des Adopters greift.
#
# Verhalten belegt: test/span-emit-wrapper.bats (fehlender, vorhandener und nicht
# ausfuehrbarer Traeger, Wurzel ohne CLAUDE_PROJECT_DIR) und der Voll-E2E-Lauf
# harness/tools/full-smoke.sh (der Wrapper schreibt im gebootstrappten Ziel einen Span).
#
# Kein `set -e`: ein abbrechender Wrapper ist ein Hook mit Nicht-Null-Ausgang, und der
# blockt den Tool-Call, den er nur beobachten soll.
set -uo pipefail

root="${CLAUDE_PROJECT_DIR:-}"
if [ -z "$root" ]; then
  # Ohne die Projekt-Variable die Wurzel aus dem Ort DIESER Datei ableiten; sie liegt
  # als <wurzel>/.claude/hooks/span-emit.sh.
  root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." 2>/dev/null && pwd)" || exit 0
fi

bin_dir="$root/.harness/state/bin"
for carrier in "$bin_dir/ai-harness-init" "$bin_dir/ai-harness-init.exe"; do
  if [ -x "$carrier" ]; then
    "$carrier" span-emit || true
    break
  fi
done
exit 0
