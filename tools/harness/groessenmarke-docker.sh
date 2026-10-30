#!/bin/sh
# tools/harness/groessenmarke-docker.sh — führt das Programm groessenmarke aus der Stufe
# `groessenmarke` des Dockerfile im Container aus (SPEC-128 Nr. 7 g): ohne Netz, ohne Mount; die
# Dateien gehen mit `docker cp` hinein, Ausgabe und Exit-Code des Programms kommen heraus, der
# Container wird in jedem Fall entfernt.
# Aufruf: groessenmarke-docker.sh abfrage <Plan>
#         groessenmarke-docker.sh pruefen <Plan> <Messung>
set -u
image=${GROESSENMARKE_IMAGE:-odata-recorder:groessenmarke}
docker=${DOCKER:-docker}

aufruf() {
	echo "Aufruf: groessenmarke-docker.sh abfrage <Plan> | pruefen <Plan> <Messung>" >&2
	exit 2
}
case ${1:-} in
	abfrage) [ $# -eq 2 ] || aufruf ;;
	pruefen) [ $# -eq 3 ] || aufruf ;;
	*) aufruf ;;
esac

kommando=$1
plan=$2
name=${plan##*/}
if [ "$kommando" = pruefen ]; then
	c=$("$docker" create --network none "$image" pruefen "$name" messung) || exit 2
else
	c=$("$docker" create --network none "$image" abfrage "$name") || exit 2
fi
trap '"$docker" rm -f "$c" > /dev/null 2>&1' EXIT
trap 'exit 2' HUP INT TERM
"$docker" cp "$plan" "$c:/work/$name" || exit 2
if [ "$kommando" = pruefen ]; then
	"$docker" cp "$3" "$c:/work/messung" || exit 2
fi
"$docker" start -a "$c"
rc=$?
exit "$rc"
