#!/bin/sh
# tools/harness/lint.sh — Prüfungen des Gates `make lint` (SPEC-101), im Modul-Wurzelverzeichnis.
# Eigene Prüfungen, dann golangci-lint; beide laufen, der Exit-Code ist 1 bei einem Befund
# in einem von ihnen und 2, wenn das Profil fehlt.
# Gelesene Form des Profils: unter `linters:` der Block `  exclusions:` mit
# `    warn-unused: true` und `    rules:`, jede Regel als `      - ` (sechs Leerzeichen), ihr `# Why:`-Block mit derselben Einrückung.
set -u

profil=.golangci.yml
[ -r "$profil" ] || { echo "lint.sh: Profil $profil fehlt oder ist nicht lesbar" >&2; exit 2; }

rc=0

# Keine Unterdrückungs-Direktive im Code (AGENTS.md §3.2), auch nicht in Testdateien.
d=$(grep -rnE --include='*.go' '//[[:space:]]*(nolint|lint:ignore)' . | sed 's/^\([^:]*:[0-9]*\):.*/lint.sh: \1: Direktive nolint oder lint:ignore/')
[ -z "$d" ] || { echo "$d"; rc=1; }

# Das Modul trägt mindestens eine Go-Datei.
[ -n "$(find . -name '*.go' -print -quit)" ] || { echo "lint.sh: ./: keine Datei *.go im Modul"; rc=1; }

# Jede Regel unter exclusions.rules steht unter einem Kommentarblock (sechs Leerzeichen eingerückt),
# dessen erste Zeile `# Why:` ist;
# warn-unused ist gesetzt.
awk -v p="$profil" '
function befund(n, t) { printf "lint.sh: %s:%d: %s\n", p, n, t; bad = 1 }
/^[^ #]/ || /^  [^ #]/ { inx = 0; inr = 0 }
/^  exclusions:/ { inx = 1; first = ""; prev = 0; next }
inx && /^    warn-unused: true[ ]*(#.*)?$/ { wu = 1 }
inx && /^    rules:[ ]*\[\][ ]*(#.*)?$/ { inr = 0; next }
inx && /^    rules:[ ]*[^ #]/ { befund(NR, "rules nicht als Block-Liste"); next }
inx && /^    rules:/ { inr = 1; first = ""; prev = 0; next }
inx && /^    [^ #-]/ { inr = 0 }
inr && /^ *#/ {
    if (/^      #/) { if (!prev) first = $0; prev = 1 } else { first = ""; prev = 0 }
    next
}
inr && /^[ ]*$/ { first = ""; prev = 0; next }
inr && /^ ? ? ? ? ?- / { befund(NR, "Regel nicht in der gelesenen Form (sechs Leerzeichen)") }
inr && /^      - / {
    if (first !~ /^ *# Why: *[^ ]/) befund(NR, "Regel ohne Kommentarblock, dessen erste Zeile mit # Why: beginnt")
}
inr { first = ""; prev = 0 }
END {
    if (!wu) befund(0, "linters.exclusions.warn-unused: true fehlt")
    exit bad
}' "$profil" || rc=1

golangci-lint run -c "$profil" ./... || rc=1

exit "$rc"
