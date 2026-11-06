#!/bin/sh
# nachbau.sh — Messung des Gewinns, Seite "bisheriges Vorgehen": führt nachbau-innen.sh in einem Container aus
# (odata-recorder:race, --memory=2g, ohne Netz, ohne Mount; der Baum geht per docker cp hinein). Der Aufrufer hält den
# flock. Aufruf: nachbau.sh   (Umgebung REP: Wiederholungen des Satzes, Vorgabe 1)
# Die Werkzeug-Seite: sh tools/harness/mutationslauf.sh tools/harness/messung-mutationslauf/messsatz.json
set -u
hier=$(cd "$(dirname "$0")" && pwd)
wurzel=$(cd "$hier/../../.." && pwd)
c=$(docker create --network none --memory=2g --memory-swap=2g --cpus=4 --pids-limit=1024 -e GOMEMLIMIT=1500MiB \
	-e REP="${REP:-1}" odata-recorder:race sh /tmp/x/nachbau-innen.sh) || exit 2
trap 'docker rm -f -v "$c" > /dev/null 2>&1' EXIT
tar -C "$wurzel" --exclude=./.git --transform 's,^\.,baum,' -cf - . | docker cp - "$c:/tmp" || exit 2
tar -C "$hier" --transform 's,^,x/,' -cf - subst nachbau-innen.sh | docker cp - "$c:/tmp" || exit 2
docker start -a "$c"
