#!/usr/bin/env bash
# tools/harness/zeichen.sh — Treiber von `make zeichen-check` (SPEC-233 Nr. 2 bis 4).
# Aufruf: zeichen.sh (ohne Argumente)
# Wählt die Dateien des Git-Arbeitsbaums (getrackt und ungetrackt, ohne ignorierte, nur was im
# Arbeitsbaum existiert), übergibt sie als Tar-Strom an einen Container der Stufe zeichen
# (`docker create`, `cp -`, `start -a`, `rm -f`; kein Mount, kein Netz) und gibt dessen Exit weiter:
# 0 und 1 bleiben, jeder andere Wert wird 2. Ein Container, der entstanden ist, wird nach normalem Ende, nach
# einem Fehler von cp, tar oder start und nach HUP, INT oder TERM entfernt (Fall 154). Ungeprüft bleibt SIGKILL
# (SPEC-233 Nr. 16 g); für ein Signal zwischen create und trap steht im Vertrag keine Zusage.
# Umgebung: ZEICHEN_IMAGE (Voreinstellung odata-recorder:zeichen), DOCKER (Voreinstellung docker).
set -euo pipefail
shopt -s lastpipe

[ $# -eq 0 ] || { echo "Aufruf: zeichen.sh" >&2; exit 2; }
fehler() { echo "zeichen.sh: $1" >&2; exit 2; }

command -v git > /dev/null 2>&1 || fehler "git nicht aufrufbar"
wurzel=$(git rev-parse --show-toplevel 2> /dev/null) || fehler "kein Git-Arbeitsbaum"
cd "$wurzel"

# Pfade mit jedem Byte außer NUL: die Liste ist NUL-getrennt. Was im Arbeitsbaum fehlt, gehört nicht
# dazu; ein hängender Symlink existiert (`-L`).
liste=()
git ls-files -z --cached --others --exclude-standard | while IFS= read -r -d '' f; do
  if [ -e "$f" ] || [ -L "$f" ]; then liste+=("$f"); fi
done || fehler "git ls-files scheitert"
[ ${#liste[@]} -gt 0 ] || fehler "Prüfmenge leer"

image=${ZEICHEN_IMAGE:-odata-recorder:zeichen}
docker=${DOCKER:-docker}

c=$("$docker" create --network none "$image" pruefen /work) || fehler "docker create scheitert"
trap '"$docker" rm -f "$c" > /dev/null 2>&1 || true' EXIT
trap 'exit 2' HUP INT TERM

printf '%s\0' "${liste[@]}" | tar -c --null --no-recursion -T - -f - | "$docker" cp - "$c:/work" || fehler "Übergabe scheitert"

rc=0
"$docker" start -a "$c" || rc=$?
case $rc in
  0 | 1) exit "$rc" ;;
  *) exit 2 ;;
esac
