#!/bin/sh
# tools/harness/test-security.sh — Treiber von `make test-security` (SPEC-041): baut die Stufe security des
# Dockerfile und fährt die Tests TestSicherheit… der drei Test-Binaries je in einem frischen Container ohne Netz,
# im Arbeitsverzeichnis ihres Pakets. Exit: 0 grün, 1 rot (alle drei Binaries laufen immer), 2 Aufruf, Aufbau,
# Name nicht gefunden oder Abbruch durch INT oder TERM.
set -u

if [ $# -ne 0 ]; then
	echo "Aufruf: tools/harness/test-security.sh (ohne Argument)" >&2
	exit 2
fi

DOCKER=${DOCKER:-docker}
VERSION=${VERSION:-dev}
IMAGE=odata-recorder:security

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

# laufe fährt das Binary $3 im Arbeitsverzeichnis $1 im Image: kein Netz, Speicherlimit, kein Schreibzugriff außer
# dem tmpfs unter /tmp und den tmpfs, die $2 trägt (für e2e.test /voll), kein Mount (SPEC-041 Nr. 1); $2 sind
# weitere Flags von docker run (ohne Leerzeichen in einem Wert), leer für keine. Ob der Container netzlos,
# schreibgeschützt und ohne schreibbaren Mount ist, beobachtet TestSicherheitUmgebung allein im Container von
# e2e.test (1 von 3); --tmpfs /tmp fangen bootstrap.test und e2e.test (2 von 3).
laufe() {
	wd=$1
	flags=$2
	shift 2
	# shellcheck disable=SC2086
	$DOCKER run --rm --network none --memory 2g --read-only --tmpfs /tmp $flags -w "$wd" "$IMAGE" "$@"
}

nicht_gefunden() {
	echo "test-security: Test $2 nicht gefunden ($1)" >&2
	rm -rf "$tmp"
	exit 2
}

# pruefe_namen: jeder Name $4… steht in -test.list des Binary $3 im Verzeichnis $1.
pruefe_namen() {
	wd=$1
	flags=$2
	bin=$3
	shift 3
	if ! fahre "$tmp/liste" laufe "$wd" "$flags" "$bin" -test.list '.*'; then
		cat "$tmp/liste" >&2
		echo "test-security: -test.list von $bin gescheitert" >&2
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

# binary fährt $3 im Verzeichnis $1 mit der Auswahl ^TestSicherheit; grün ist es bei Exit 0, jedem Namen $4… als
# `--- PASS: <Name> ` und ohne `--- SKIP` und `--- FAIL`.
binary() {
	wd=$1
	flags=$2
	bin=$3
	shift 3
	fahre "$tmp/aus" laufe "$wd" "$flags" "$bin" -test.run '^TestSicherheit' -test.v -test.count=1 -test.timeout 120s
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
		echo "test-security: $bin grün"
		return
	fi
	rot=1
	echo "test-security: rot ($bin: $grund)"
	cat "$tmp/aus"
}

if ! fahre "$tmp/bau" $DOCKER build --progress=plain --build-arg VERSION="$VERSION" --target security -t "$IMAGE" .; then
	cat "$tmp/bau" >&2
	echo "test-security: Aufbau der Stufe security gescheitert" >&2
	rm -rf "$tmp"
	exit 2
fi

adapter="TestSicherheitRouten TestSicherheitAbweisung TestSicherheitKeineAenderung TestSicherheitKontrolle TestSicherheitHealthz TestSicherheitToken"
root="TestSicherheitStartPruefung TestSicherheitErreichbar"
e2e="TestSicherheitListen TestSicherheitCanary TestSicherheitFormen TestSicherheitUmgebung"
voll="--tmpfs /voll:rw,size=1m -e E2E_BINARY=/out/odata-recorder"
# shellcheck disable=SC2086
pruefe_namen /src/src/adapters/driving/admin "" /out/admin.test $adapter
# shellcheck disable=SC2086
pruefe_namen /src/src/bootstrap "" /out/bootstrap.test $root
# shellcheck disable=SC2086
pruefe_namen /src/tests/e2e "$voll" /out/e2e.test $e2e

# shellcheck disable=SC2086
binary /src/src/adapters/driving/admin "" /out/admin.test $adapter
# shellcheck disable=SC2086
binary /src/src/bootstrap "" /out/bootstrap.test $root
# shellcheck disable=SC2086
binary /src/tests/e2e "$voll" /out/e2e.test $e2e

rm -rf "$tmp"
[ "$rot" -eq 0 ] && echo "test-security: grün"
exit "$rot"
