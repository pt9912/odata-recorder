#!/bin/sh
# tools/harness/lasttest.sh — Treiber von `make lasttest` (SPEC-216 Nr. 2 und 3): prüft die Variablen, ruft `docker info`, baut die Stufe
# lasttest des Dockerfile und fährt /out/lasttest in einem Container mit 2 CPU, 4 GiB ohne Swap, ohne Netz und ohne Mount. Er ruft von
# docker allein info, build und run. Exit: der Exit des Programms (0 bis 3), 2 bei Aufruf, Variable, docker, Bau oder Abbruch, 3 bei jedem
# anderen Exit des Containers. Meldungen stehen auf stderr mit dem Präfix `lasttest: `.
set -u
LC_ALL=C
export LC_ALL

melde() {
	printf 'lasttest: %s\n' "$1" >&2
}

abbrechen() {
	melde "$1"
	exit 2
}

[ $# -eq 0 ] || abbrechen "Aufruf: tools/harness/lasttest.sh (ohne Argument)"

# Die Variablen in der Folge von SPEC-216 Nr. 2: nicht gesetzt gilt die Vorgabe, gesetzt und leer ist ein Fehler.
[ "${LASTTEST_BETRIEBSART+x}" ] || LASTTEST_BETRIEBSART=beide
[ "${LASTTEST_PROFIL+x}" ] || LASTTEST_PROFIL=abnahme
[ "${LASTTEST_ZIEL_MS+x}" ] || LASTTEST_ZIEL_MS=100
[ "${LASTTEST_STARTWERT+x}" ] || LASTTEST_STARTWERT=1

nicht_leer() {
	[ -n "$2" ] || abbrechen "$1 leer"
}

form() {
	abbrechen "$1 Form: '$2'"
}

nicht_leer LASTTEST_BETRIEBSART "$LASTTEST_BETRIEBSART"
case $LASTTEST_BETRIEBSART in
replay | mock | beide) ;;
*) form LASTTEST_BETRIEBSART "$LASTTEST_BETRIEBSART" ;;
esac

nicht_leer LASTTEST_PROFIL "$LASTTEST_PROFIL"
case $LASTTEST_PROFIL in
abnahme | rauch) ;;
*) form LASTTEST_PROFIL "$LASTTEST_PROFIL" ;;
esac

nicht_leer LASTTEST_ZIEL_MS "$LASTTEST_ZIEL_MS"
case $LASTTEST_ZIEL_MS in
*[!0]*) ;;
*) abbrechen "LASTTEST_ZIEL_MS null" ;;
esac
case $LASTTEST_ZIEL_MS in
[1-9] | [1-9][0-9] | [1-9][0-9][0-9] | [1-9][0-9][0-9][0-9]) ;;
*) form LASTTEST_ZIEL_MS "$LASTTEST_ZIEL_MS" ;;
esac

nicht_leer LASTTEST_STARTWERT "$LASTTEST_STARTWERT"
case $LASTTEST_STARTWERT in
0) ;;
0* | *[!0-9]*) form LASTTEST_STARTWERT "$LASTTEST_STARTWERT" ;;
*) [ ${#LASTTEST_STARTWERT} -le 18 ] || form LASTTEST_STARTWERT "$LASTTEST_STARTWERT" ;;
esac

DOCKER=${DOCKER:-docker}
IMAGE=odata-recorder:lasttest

tmp=$(mktemp -d) || exit 2
trap 'rm -rf "$tmp"' EXIT
signal=
pid=

# abbruch merkt sich INT, HUP oder TERM und reicht es als TERM an den laufenden Container weiter; der Treiber endet, sobald der Aufruf
# von docker zurückkehrt (SPEC-216 Nr. 3 e).
abbruch() {
	signal=1
	[ -z "$pid" ] || kill -TERM "$pid" 2>/dev/null
}
trap abbruch INT HUP TERM

pruefe_abbruch() {
	[ -z "$signal" ] || abbrechen abgebrochen
}

$DOCKER info --format '{{.NCPU}}' >"$tmp/info"
rc=$?
pruefe_abbruch
[ "$rc" -eq 0 ] || abbrechen "docker nicht erreichbar"
kerne=$(cat "$tmp/info")

$DOCKER build --progress=plain --target lasttest -t "$IMAGE" .
rc=$?
pruefe_abbruch
[ "$rc" -eq 0 ] || abbrechen "Bau gescheitert"

$DOCKER run --rm --network none --cpus 2 --memory 4g --memory-swap 4g \
	--env LASTTEST_BETRIEBSART="$LASTTEST_BETRIEBSART" --env LASTTEST_PROFIL="$LASTTEST_PROFIL" \
	--env LASTTEST_ZIEL_MS="$LASTTEST_ZIEL_MS" --env LASTTEST_STARTWERT="$LASTTEST_STARTWERT" \
	--env LASTTEST_KERNE="$kerne" "$IMAGE" /out/lasttest &
pid=$!
[ -z "$signal" ] || kill -TERM "$pid" 2>/dev/null
wait "$pid"
rc=$?
# Ein Signal unterbricht wait; der Treiber wartet danach, bis der Aufruf von docker zurückkehrt.
if [ -n "$signal" ]; then
	while kill -0 "$pid" 2>/dev/null; do
		wait "$pid"
	done
	pid=
	abbrechen abgebrochen
fi
pid=

case $rc in
0 | 1 | 2 | 3) exit "$rc" ;;
125 | 126 | 127) abbrechen "docker run gescheitert" ;;
*)
	melde "Container endete mit Exit $rc"
	exit 3
	;;
esac
