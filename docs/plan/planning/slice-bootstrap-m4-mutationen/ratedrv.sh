#!/bin/sh
# ratedrv.sh <Liste> <Paket> <N>: Raten je Mutation, ein Container ohne Netz auf dem Race-Image (--memory=4g --cpus=4); der Aufrufer hält den flock.
set -u
S=/tmp/claude-1000/-home-builder-work-odata-recorder/52cdc2d3-3257-4d8f-945b-ac70d33f302c/scratchpad
W=/home/builder/work/odata-recorder/.wt/m4
liste=$1
paket=$2
n=$3
c=$(docker create --network none --memory=4g --memory-swap=4g --cpus=4 --pids-limit=1024 -e GOFLAGS=-mod=readonly -e CGO_ENABLED=1 -e GOMEMLIMIT=3000MiB \
	-w /work/baum odata-recorder:race sh -c "cd /work/treiber && go build -o /work/ratedrv . && cd /work/baum && /work/ratedrv /work/liste.json $paket $n") || exit 2
trap 'docker rm -f -v "$c" > /dev/null 2>&1' EXIT
tar -C "$W" --exclude=.git --transform 's,^,baum/,' -cf - go.mod go.sum src tests | docker cp - "$c:/work" || exit 2
tar -C "$S/ratedrv" --transform 's,^,treiber/,' -cf - main.go go.mod | docker cp - "$c:/work" || exit 2
tar -C "$(dirname "$liste")" --transform "s,^$(basename "$liste"),liste.json," -cf - "$(basename "$liste")" | docker cp - "$c:/work" || exit 2
docker start -a "$c"
