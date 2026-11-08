#!/bin/sh
# tools/harness/test-contract.sh — Treiber von `make test-contract` (SPEC-042): baut die Stufe contract des
# Dockerfile und fährt die Tests TestVertrag… der beiden Test-Binaries je in einem frischen Container ohne
# Netz, im Arbeitsverzeichnis ihres Pakets. Exit: 0 grün, 1 rot (beide Binaries laufen immer), 2 Aufruf,
# Aufbau, Name nicht gefunden oder Abbruch durch INT oder TERM.
set -u

if [ $# -ne 0 ]; then
	echo "Aufruf: tools/harness/test-contract.sh (ohne Argument)" >&2
	exit 2
fi

DOCKER=${DOCKER:-docker}
VERSION=${VERSION:-dev}
IMAGE=odata-recorder:contract

tmp=$(mktemp -d) || exit 2
pid=
rot=0

# abbruch beendet den laufenden docker-Aufruf: docker leitet das Signal an den Container weiter,
# --rm entfernt ihn.
abbruch() {
	[ -z "$pid" ] || { kill -TERM "$pid" 2>/dev/null; wait "$pid" 2>/dev/null; }
	rm -rf "$tmp"
	exit 2
}
trap abbruch INT TERM

# fahre führt das Kommando im Hintergrund aus, damit ein Signal des Treibers es erreicht, und schreibt
# Ausgabe und Fehlerstrom in die Datei $1.
fahre() {
	aus=$1
	shift
	"$@" >"$aus" 2>&1 &
	pid=$!
	wait "$pid"
	rc=$?
	pid=
	return "$rc"
}

# laufe fährt $2 (Binary) im Arbeitsverzeichnis $1 im Image: kein Netz, Speicherlimit, kein Schreibzugriff außer
# dem tmpfs unter /tmp, kein Mount (SPEC-042 Nr. 1).
laufe() {
	wd=$1
	shift
	$DOCKER run --rm --network none --memory 2g --read-only --tmpfs /tmp -w "$wd" "$IMAGE" "$@"
}

nicht_gefunden() {
	echo "test-contract: Test $2 nicht gefunden ($1)" >&2
	rm -rf "$tmp"
	exit 2
}

# pruefe_namen: jeder Name $3… steht in -test.list des Binary $2 im Verzeichnis $1.
pruefe_namen() {
	wd=$1
	bin=$2
	shift 2
	if ! fahre "$tmp/liste" laufe "$wd" "$bin" -test.list '.*'; then
		cat "$tmp/liste" >&2
		echo "test-contract: -test.list von $bin gescheitert" >&2
		rm -rf "$tmp"
		exit 2
	fi
	for name in "$@"; do
		case $name in
		'' | *[!A-Za-z0-9_]*) nicht_gefunden "$bin" "$name" ;;
		esac
		grep -qx "$name" "$tmp/liste" || nicht_gefunden "$bin" "$name"
	done
}

# binary fährt $2 im Verzeichnis $1 mit der Auswahl ^TestVertrag; grün ist es bei Exit 0, jedem Namen $3… als
# `--- PASS: <Name> ` und ohne `--- SKIP` und `--- FAIL`.
binary() {
	wd=$1
	bin=$2
	shift 2
	fahre "$tmp/aus" laufe "$wd" "$bin" -test.run '^TestVertrag' -test.v -test.count=1 -test.timeout 120s
	rc=$?
	grund=
	if [ "$rc" -ne 0 ]; then
		grund="Exit $rc"
	elif grep -q '^ *--- SKIP' "$tmp/aus"; then
		grund=SKIP
	elif grep -q '^ *--- FAIL' "$tmp/aus"; then
		grund=FAIL
	else
		for name in "$@"; do
			grep -q "^--- PASS: $name " "$tmp/aus" || grund="kein PASS ($name)"
		done
	fi
	if [ -z "$grund" ]; then
		echo "test-contract: $bin grün"
		return
	fi
	rot=1
	echo "test-contract: rot ($bin: $grund)"
	cat "$tmp/aus"
}

if ! fahre "$tmp/bau" $DOCKER build --progress=plain --build-arg VERSION="$VERSION" --target contract -t "$IMAGE" .; then
	cat "$tmp/bau" >&2
	echo "test-contract: Aufbau der Stufe contract gescheitert" >&2
	rm -rf "$tmp"
	exit 2
fi

boot="TestVertragMatrix TestVertragMatrixZaehlung TestVertragVergleich TestVertragDrahtQuery"
integ="TestVertragFehlerklassen TestVertragFehlerobjekt"
# shellcheck disable=SC2086
pruefe_namen /src/src/bootstrap /out/bootstrap.test $boot
# shellcheck disable=SC2086
pruefe_namen /src/tests/integration /out/integration.test $integ

# shellcheck disable=SC2086
binary /src/src/bootstrap /out/bootstrap.test $boot
# shellcheck disable=SC2086
binary /src/tests/integration /out/integration.test $integ

rm -rf "$tmp"
[ "$rot" -eq 0 ] && echo "test-contract: grün"
exit "$rot"
