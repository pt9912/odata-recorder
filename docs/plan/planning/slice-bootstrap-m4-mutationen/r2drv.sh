#!/bin/sh
# r2drv.sh <Liste> [Paket]: ein Container ohne Netz (--memory=4g --cpus=4), Baum und Treiber per docker cp, Lauf je Mutation.
# Der Aufrufer hält den flock. Der Treiber ersetzt die Nullprobe von tools/harness/mutationslauf.sh (Zeitgrenze 2 Minuten je Prozess
# reicht für das Paket src/bootstrap nicht).
set -u
S=$(cd "$(dirname "$0")" && pwd)
W=$(cd "$S/../../../.." && pwd)
liste=$1
paket=${2:-./src/bootstrap/}
c=$(docker create --network none --memory=4g --memory-swap=4g --cpus=4 --pids-limit=1024 -e GOFLAGS=-mod=readonly -e CGO_ENABLED=0 -e GOMEMLIMIT=3000MiB \
	-w /work/baum odata-recorder:test sh -c "cd /work/treiber && go build -o /work/r2drv . && cd /work/baum && /work/r2drv /work/liste.json $paket") || exit 2
trap 'docker rm -f -v "$c" > /dev/null 2>&1' EXIT
tar -C "$W" --exclude=.git --transform 's,^,baum/,' -cf - go.mod go.sum src tests | docker cp - "$c:/work" || exit 2
tar -C "$S/r2drv" --transform 's,^,treiber/,' -cf - main.go go.mod | docker cp - "$c:/work" || exit 2
tar -C "$(dirname "$liste")" --transform "s,^$(basename "$liste"),liste.json," -cf - "$(basename "$liste")" | docker cp - "$c:/work" || exit 2
docker start -a "$c"
