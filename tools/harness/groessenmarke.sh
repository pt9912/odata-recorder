#!/bin/sh
# tools/harness/groessenmarke.sh — Treiber von `make groessenmarke` (SPEC-128 Nr. 7).
# Aufruf: groessenmarke.sh <Programm> <Planverzeichnis>
# Prüft jede Datei slice-*.md unmittelbar im Planverzeichnis, nach Namen sortiert: fragt das
# Programm (`<Programm> abfrage <Plan>`), ruft bei `messen` an drei Stellen `git` (rev-parse,
# merge-base, diff) und übergibt die Messung als Datei an `<Programm> pruefen <Plan> <Messung>`.
# Exit 2, wenn ein Plan 2 ergab, sonst 1, wenn einer 1 ergab, sonst 0; alle Pläne laufen zu Ende.
set -u
LC_ALL=C
export LC_ALL

if [ $# -ne 2 ] || [ -z "$1" ] || [ -z "$2" ]; then
	echo "Aufruf: groessenmarke.sh <Programm> <Planverzeichnis>" >&2
	exit 2
fi
programm=$1
verzeichnis=$2
[ -d "$verzeichnis" ] || { echo "groessenmarke: Planverzeichnis $verzeichnis fehlt" >&2; exit 2; }

tmp=$(mktemp -d) || { echo "groessenmarke: mktemp scheitert" >&2; exit 2; }
trap 'rm -rf "$tmp"' EXIT
trap 'exit 2' HUP INT TERM

# je_plan <Plan>: Ergebnis 0, 1 oder 2 als Exit-Code der Funktion.
je_plan() {
	plan=$1
	rm -f "$tmp/abfrage" "$tmp/pfade" "$tmp/numstat" "$tmp/messung" "$tmp/git"
	"$programm" abfrage "$plan" > "$tmp/abfrage" || return 2

	modus=
	basis=
	set --
	: > "$tmp/pfade"
	while IFS= read -r z; do
		case $z in
			keine-messung) modus=keine ;;
			messen) modus=messen ;;
			"basis "*) basis=${z#basis } ;;
			"pfad "*) printf '%s\n' "$z" >> "$tmp/pfade"; set -- "$@" ":(literal)${z#pfad }" ;;
		esac
	done < "$tmp/abfrage"
	[ "$modus" != keine ] || return 0
	if [ "$modus" != messen ]; then
		echo "groessenmarke $plan: Eingabe: abfrage ohne Ergebnis" >&2
		return 2
	fi

	if ! command -v git > /dev/null 2>&1; then
		echo "groessenmarke $plan: Eingabe: git nicht aufrufbar" >&2
		return 2
	fi
	spitze=$(git rev-parse --verify HEAD 2> /dev/null) || {
		echo "groessenmarke $plan: Eingabe: HEAD nicht lesbar" >&2
		return 2
	}
	git merge-base --is-ancestor "$basis" HEAD > "$tmp/git" 2>&1
	rc=$?
	if [ "$rc" -eq 1 ]; then
		echo "groessenmarke $plan: Eingabe: Basis $basis ist kein Vorfahr von HEAD" >&2
		return 2
	elif [ "$rc" -ne 0 ]; then
		echo "groessenmarke $plan: Eingabe: Basis $basis unbekannt oder git scheitert: $(cat "$tmp/git")" >&2
		return 2
	fi
	git diff --numstat -z --no-renames "$basis..HEAD" -- "$@" ":(exclude,literal)harness/vertraege" > "$tmp/numstat" 2> "$tmp/git" || {
		echo "groessenmarke $plan: Eingabe: git diff scheitert: $(cat "$tmp/git")" >&2
		return 2
	}
	{
		printf 'basis %s\nspitze %s\n' "$basis" "$spitze"
		cat "$tmp/pfade"
		printf 'numstat\n'
		cat "$tmp/numstat"
	} > "$tmp/messung"
	"$programm" pruefen "$plan" "$tmp/messung"
	rc=$?
	[ "$rc" -le 2 ] || rc=2
	return "$rc"
}

gefunden=0
ergebnis=0
for plan in "$verzeichnis"/slice-*.md; do
	[ -f "$plan" ] || continue
	gefunden=1
	je_plan "$plan"
	rc=$?
	if [ "$rc" -eq 2 ]; then
		ergebnis=2
	elif [ "$rc" -eq 1 ] && [ "$ergebnis" -eq 0 ]; then
		ergebnis=1
	fi
done
[ "$gefunden" -eq 1 ] || echo "groessenmarke: kein Slice in Arbeit"
exit "$ergebnis"
