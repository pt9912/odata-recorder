#!/bin/sh
# tools/harness/test-integration.sh — Treiber von `make test-integration` (SPEC-037): baut die Stufe
# integration des Dockerfile und fährt jede Stufe als Test von tests/e2e in einem frischen Container
# ohne Netz. Exit: 0 alle Stufen grün, 1 mindestens eine rot (alle laufen), 2 Aufruf, Aufbau oder
# Abbruch durch INT oder TERM. Das Skript ist die Quelle von `make e2e-abdeckung` (SPEC-037 Nr. 3).
set -u

if [ $# -ne 0 ]; then
	echo "Aufruf: tools/harness/test-integration.sh (ohne Argument)" >&2
	exit 2
fi

DOCKER=${DOCKER:-docker}
VERSION=${VERSION:-dev}
IMAGE=odata-recorder:integration

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

# fahre führt das Kommando im Hintergrund aus, damit ein Signal des Treibers es erreicht, und
# schreibt Ausgabe und Fehlerstrom in die Datei $1.
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

# laufe fährt ein Kommando im Image ohne Netz, mit Speicherlimit, ohne Schreibzugriff außer den beiden
# tmpfs und ohne Mount; /voll ist 1 MiB groß (SPEC-037 Nr. 8).
laufe() {
	$DOCKER run --rm --network none --memory 2g --read-only --tmpfs /tmp --tmpfs /voll:rw,size=1m \
		-e E2E_BINARY=/out/odata-recorder -e E2E_UEBERGANG=/out/uebergang "$IMAGE" "$@"
}

e2e_abdeckung() { echo "abdeckung: $1 — $2 (Anker: $3)"; }

# stufe fährt den Test $2 als Stufe $1: ein Name, den das Test-Binary nicht listet, ist Exit 2; grün ist
# die Stufe bei Exit 0 mit der Zeile `--- PASS: <Name>` und ohne `--- SKIP` und `--- FAIL`.
stufe() {
	n=$1
	name=$2
	case $name in
	'' | *[!A-Za-z0-9_]*) nicht_gefunden ;;
	esac
	grep -qx "$name" "$tmp/liste" || nicht_gefunden
	fahre "$tmp/aus" laufe /out/e2e.test -test.run "^$name\$" -test.v -test.count=1 -test.timeout 120s
	rc=$?
	grund=
	if [ "$rc" -ne 0 ]; then
		grund="Exit $rc"
	elif grep -q '^ *--- SKIP' "$tmp/aus"; then
		grund=SKIP
	elif grep -q '^ *--- FAIL' "$tmp/aus"; then
		grund=FAIL
	elif ! grep -q "^--- PASS: $name " "$tmp/aus"; then
		grund="kein PASS"
	fi
	if [ -z "$grund" ]; then
		echo "Stufe $n: $name grün"
		return
	fi
	rot=1
	echo "Stufe $n: $name rot ($grund)"
	cat "$tmp/aus"
}

nicht_gefunden() {
	echo "Stufe $n: Test $name nicht gefunden" >&2
	rm -rf "$tmp"
	exit 2
}

if ! fahre "$tmp/bau" $DOCKER build --progress=plain --build-arg VERSION="$VERSION" --target integration -t "$IMAGE" .; then
	cat "$tmp/bau" >&2
	echo "test-integration: Aufbau der Stufe integration gescheitert" >&2
	rm -rf "$tmp"
	exit 2
fi
if ! fahre "$tmp/liste" laufe /out/e2e.test -test.list '.*'; then
	cat "$tmp/liste" >&2
	echo "test-integration: -test.list gescheitert" >&2
	rm -rf "$tmp"
	exit 2
fi

echo "integration: Prozess, Konfiguration und Signal des Binary ..."
e2e_abdeckung "LH-RB-03" "Kein Außennetz, ungültige Konfiguration, SIGTERM (Happy Path, Negative)" "TestE2EProzess"
stufe 1 TestE2EProzess

echo "integration: Weiterleitung jeder Antwortart, leere Antwort, Upstream gestoppt ..."
e2e_abdeckung "LH-FA-01" "Antwortarten unverändert, leere Antwort, Upstream gestoppt (Happy Path, Boundary, Negative)" "TestE2EWeiterleitung"
stufe 2 TestE2EWeiterleitung

echo "integration: Aufzeichnung, Upstream stoppen, Replay, Speicher voll ..."
e2e_abdeckung "LH-FA-02" "Rundlauf mit GET und zwei POST, leere Aufnahme, Speicher voll (Happy Path, Boundary, Negative)" "TestE2EAufzeichnung"
stufe 3 TestE2EAufzeichnung

echo "integration: Persistenz über Neustart, leere Session, beschädigte Daten ..."
e2e_abdeckung "LH-FA-03" "Neustart auf demselben DATA_DIR, Session ohne Request, beschädigte Datei (Happy Path, Boundary, Negative)" "TestE2EPersistenz"
stufe 4 TestE2EPersistenz

echo "integration: Replay bei gestopptem Referenzdienst ..."
e2e_abdeckung "LH-FA-04" "Replay gleich Aufnahme, Folge 7, 11, 13, Fehlerfälle und Übergang (Happy Path, Boundary, Negative)" "TestE2EReplay"
stufe 5 TestE2EReplay

echo "integration: Matching von Query und Prozent-Kodierung ..."
e2e_abdeckung "LH-FA-05" "Reihenfolge und Hexziffern, Filterwert, ungültige Kodierung (Happy Path, Boundary, Negative)" "TestE2EMatching"
stufe 6 TestE2EMatching

echo "integration: Schutz der Geheimnisse in Dateien und Strömen ..."
e2e_abdeckung "LH-FA-14" "Geheimnisse, Nutzlast ohne sensible Felder, nicht redigierbarer Body, Speicher voll (Happy Path, Boundary, Negative)" "TestE2ESchutz"
stufe 7 TestE2ESchutz

echo "integration: Modus mock, CRUD, Reset und Zuordnung am gebauten Binary ..."
e2e_abdeckung "LH-FA-07 LH-FA-11 LH-FA-12" "Mock CRUD und Reset ohne Upstream, Header der Zuordnung, nicht unterstützter Filter (Happy Path, Boundary, Negative)" "TestE2EMock"
stufe 8 TestE2EMock

echo "integration: Modus mock, Neustart auf demselben DATA_DIR ..."
e2e_abdeckung "LH-FA-03 LH-QA-03" "Mutationen nach SIGTERM und Neustart lesbar, zweiter Neustart, fehlendes DATA_DIR wird angelegt (Happy Path, Boundary)" "TestE2EMockNeustart"
stufe 9 TestE2EMockNeustart

rm -rf "$tmp"
exit "$rot"
