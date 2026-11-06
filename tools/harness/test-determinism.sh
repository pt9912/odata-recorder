#!/bin/sh
# tools/harness/test-determinism.sh — Treiber von `make test-determinism` (SPEC-038): baut die Stufe
# integration des Dockerfile und fährt TestE2EDeterminismus (100 Läufe des Replay nach Reset) in einem
# frischen Container ohne Netz. Exit: 0 grün, 1 rot, 2 Aufruf, Aufbau oder Abbruch durch INT oder TERM.
set -u

if [ $# -ne 0 ]; then
	echo "Aufruf: tools/harness/test-determinism.sh (ohne Argument)" >&2
	exit 2
fi

DOCKER=${DOCKER:-docker}
VERSION=${VERSION:-dev}
IMAGE=odata-recorder:integration
NAME=TestE2EDeterminismus

tmp=$(mktemp -d) || exit 2
pid=

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

if ! fahre "$tmp/bau" $DOCKER build --progress=plain --build-arg VERSION="$VERSION" --target integration -t "$IMAGE" .; then
	cat "$tmp/bau" >&2
	echo "test-determinism: Aufbau der Stufe integration gescheitert" >&2
	rm -rf "$tmp"
	exit 2
fi

# Ein Container, kein Netz, kein Mount; Schreibzugriff nur auf das tmpfs unter /tmp (SPEC-038 Nr. 1).
fahre "$tmp/aus" $DOCKER run --rm --network none --memory 2g --read-only --tmpfs /tmp \
	-e E2E_BINARY=/out/odata-recorder -e E2E_UEBERGANG=/out/uebergang "$IMAGE" \
	/out/e2e.test -test.run "^$NAME\$" -test.v -test.count=1 -test.timeout 120s
rc=$?
grund=
if [ "$rc" -ne 0 ]; then
	grund="Exit $rc"
elif grep -q '^ *--- SKIP' "$tmp/aus"; then
	grund=SKIP
elif grep -q '^ *--- FAIL' "$tmp/aus"; then
	grund=FAIL
elif ! grep -q "^--- PASS: $NAME " "$tmp/aus"; then
	grund="kein PASS"
fi
if [ -n "$grund" ]; then
	echo "test-determinism: rot ($grund)"
	cat "$tmp/aus"
	rm -rf "$tmp"
	exit 1
fi
grep 'Läufe: ' "$tmp/aus"
echo "test-determinism: grün ($NAME)"
rm -rf "$tmp"
exit 0
