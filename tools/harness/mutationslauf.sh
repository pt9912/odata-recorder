#!/bin/sh
# tools/harness/mutationslauf.sh — Treiber des Mutationslaufs (SPEC-210); kein Gate, kein make-Ziel.
# Aufruf: mutationslauf.sh <Liste>
# Legt einen Container aus dem Image der Kette an (ohne Netz, ohne Mount, mit --memory), schickt den
# Arbeitsbaum, die Liste und die Quelle von tools/mutationslauf mit `docker cp` hinein, startet den
# Lauf und gibt dessen Ausgabe und Exit-Code weiter. Der Container wird nach Erfolg, bei Exit 2 und bei
# HUP, INT und TERM entfernt; bei SIGKILL des Treibers bleibt höchstens einer mit dem Label, den der
# nächste Aufruf räumt.
# Umgebung: MUTATIONSLAUF_IMAGE (odata-recorder:race), MUTATIONSLAUF_SPEICHER (2g, 1g bis 4g),
# MUTATIONSLAUF_GRENZE (3600 Sekunden, 1 bis 14400), MUTATIONSLAUF_KALT (1: leerer Build-Cache),
# DOCKER (docker). Der Aufrufer hält den flock; der Treiber schreibt nichts auf den Host.
set -u
LC_ALL=C
export LC_ALL

docker=${DOCKER:-docker}
wurzel=$(cd "$(dirname "$0")/../.." && pwd)

if [ $# -ne 1 ] || [ ! -f "$1" ] || [ ! -r "$1" ]; then
	echo "Aufruf: mutationslauf.sh <Liste>" >&2
	exit 2
fi
liste=$1

# Ein gesetzter, leerer Wert ist ungültig (außer bei MUTATIONSLAUF_KALT); unbelegt gilt die Vorgabe.
image=${MUTATIONSLAUF_IMAGE:-odata-recorder:race}
speicher=2g
[ "${MUTATIONSLAUF_SPEICHER+gesetzt}" = gesetzt ] && speicher=$MUTATIONSLAUF_SPEICHER
grenze=3600
[ "${MUTATIONSLAUF_GRENZE+gesetzt}" = gesetzt ] && grenze=$MUTATIONSLAUF_GRENZE
kalt=${MUTATIONSLAUF_KALT:-}
ungueltig() {
	echo "mutationslauf: Umgebung ungültig: $1" >&2
	exit 2
}
case $speicher in [1-4]g) ;; *) ungueltig MUTATIONSLAUF_SPEICHER ;; esac
case $grenze in '' | *[!0-9]* | 0*) ungueltig MUTATIONSLAUF_GRENZE ;; esac
[ "$grenze" -le 14400 ] || ungueltig MUTATIONSLAUF_GRENZE
case $kalt in '' | 1) ;; *) ungueltig MUTATIONSLAUF_KALT ;; esac
case $speicher in
	2g) gomemlimit=1500MiB ;;
	*) gomemlimit=$((${speicher%g} * 768))MiB ;;
esac

kb=$(du -sk --apparent-size --exclude=.git "$wurzel" | cut -f1)
if [ "$kb" -gt 262144 ]; then
	echo "mutationslauf: Eingabe: Baum zu groß" >&2
	exit 2
fi
"$docker" image inspect "$image" > /dev/null 2>&1 || {
	echo "Image fehlt: $image (make test-race bzw. make test)" >&2
	exit 2
}

# Reste eines früheren Aufrufs: nur Container mit dem Label, die nicht laufen.
reste=$("$docker" ps -aq --filter label=odata-recorder.mutationslauf=1 --filter status=exited --filter status=created --filter status=dead)
# shellcheck disable=SC2086
[ -z "$reste" ] || "$docker" rm -f -v $reste > /dev/null 2>&1

befehl='cd /tmp/mutationslauf && CGO_ENABLED=0 GOFLAGS=-mod=readonly GOTOOLCHAIN=local GOPROXY=off go build -o /usr/local/bin/mutationslauf . || exit 2; exec mutationslauf /tmp/liste.json /tmp/baum'
c=$("$docker" create --network none --memory="$speicher" --memory-swap="$speicher" --cpus=4 --pids-limit=1024 \
	--label odata-recorder.mutationslauf=1 -e GOMEMLIMIT="$gomemlimit" -e MUTATIONSLAUF_GRENZE="$grenze" \
	-e MUTATIONSLAUF_KALT="$kalt" "$image" sh -c "$befehl") || exit 2
trap '"$docker" rm -f -v "$c" > /dev/null 2>&1' EXIT
trap 'exit 2' HUP INT TERM

tar -C "$wurzel" --exclude=./.git --transform 's,^\.,baum,' -cf - . | "$docker" cp - "$c:/tmp" || exit 2
tar -C "$wurzel/tools" -cf - mutationslauf | "$docker" cp - "$c:/tmp" || exit 2
"$docker" cp "$liste" "$c:/tmp/liste.json" || exit 2

"$docker" start -a "$c" &
pid=$!
trap '"$docker" kill --signal TERM "$c" > /dev/null 2>&1' HUP INT TERM
# Frist des Treibers: die Gesamtgrenze des Programms und 60 Sekunden, danach wird der Container getötet.
(
	s=0
	while kill -0 "$pid" 2> /dev/null && [ "$s" -lt $((grenze + 60)) ]; do
		sleep 1
		s=$((s + 1))
	done
	if kill -0 "$pid" 2> /dev/null; then "$docker" kill "$c" > /dev/null 2>&1; fi
) &
wd=$!
wait "$pid"
rc=$?
while kill -0 "$pid" 2> /dev/null; do
	wait "$pid"
	rc=$?
done
kill "$wd" 2> /dev/null
exit "$rc"
