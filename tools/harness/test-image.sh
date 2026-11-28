#!/bin/sh
# tools/harness/test-image.sh — Treiber von `make test-image` (SPEC-323): baut die Stufen image und integration des Dockerfile unter
# Tags dieses Laufs, prüft das Image, baut es ein zweites Mal ohne den Cache der Stufen build und image und vergleicht die beiden Binaries, startet
# das Image ohne Netz, fährt die Phase aufnehmen, stoppt und startet denselben Container und fährt die Phase pruefen. Jeder Name trägt die
# Kennung des Laufs; hinter jedem Ende entfernt der Treiber die Container, die Tags und das Verzeichnis seines Laufs und nichts sonst.
# Exit am Treiber: 0 grün, 1 rot (eine Prüfung, eine Phase, der Stopp, das Aufräumen), 2 Aufruf, Bau, Daemon nicht erreichbar oder Abbruch
# durch INT, HUP oder TERM.
set -u
LC_ALL=C
export LC_ALL

if [ $# -ne 0 ]; then
	echo "test-image: Aufruf: tools/harness/test-image.sh (ohne Argument)" >&2
	exit 2
fi

DOCKER=${DOCKER:-docker}
VERSION=${VERSION:-dev}

lauf=$$-$(date +%s)
IMAGE=odata-recorder:test-image-$lauf
CLIENT=odata-recorder:test-image-client-$lauf
ZWEIT=odata-recorder:test-image-zweit-$lauf
NAME=odr-test-image-$lauf
LABEL=odata-recorder.test-image=$lauf
SERVER=$NAME-server

tmp=$(mktemp -d) || exit 2
server=
rest=0

melde() {
	printf 'test-image: %s\n' "$1" >&2
}

# besteht meldet, ob die Abfrage $@ eine Kennung nennt; eine Abfrage, die scheitert, zählt als bestehendes Objekt.
besteht() {
	da=$("$@" 2>/dev/null) || return 0
	[ -n "$da" ]
}

gescheitert() {
	melde "Aufräumen gescheitert: $1"
	rest=1
}

# aufraeumen entfernt jeden Container, jedes Tag und das Verzeichnis dieses Laufs. Ein Objekt, das nie entstand, ist kein Fehler; die
# übrigen werden auch hinter einem Fehler entfernt (SPEC-323 Nr. 6 a, b).
aufraeumen() {
	for rolle in pruefung version zweit server aufnehmen pruefen; do
		$DOCKER rm -f -v "$NAME-$rolle" >/dev/null 2>&1 ||
			! besteht $DOCKER ps -a -q --filter "name=^/$NAME-$rolle\$" ||
			gescheitert "$NAME-$rolle"
	done
	for tag in "$IMAGE" "$CLIENT" "$ZWEIT"; do
		$DOCKER rmi "$tag" >/dev/null 2>&1 ||
			! besteht $DOCKER images -q "$tag" ||
			gescheitert "$tag"
	done
	rm -rf "$tmp" || gescheitert "$tmp"
}

ende() {
	aufraeumen
	if [ "$1" -eq 0 ]; then
		[ "$rest" -eq 0 ] || exit 1
		echo "test-image: grün"
	fi
	exit "$1"
}

# abbruch beendet den laufenden Aufruf, räumt auf und endet mit 2; ein weiteres Signal erreicht das Aufräumen nicht (SPEC-323 Nr. 6 d).
abbruch() {
	trap '' INT HUP TERM
	if [ -n "${!-}" ]; then
		kill -TERM "$!" 2>/dev/null
		wait "$!" 2>/dev/null
	fi
	melde abgebrochen
	aufraeumen
	exit 2
}
trap abbruch INT HUP TERM

# fahre führt $3… im Hintergrund aus, damit ein Signal des Treibers den Aufruf erreicht: Standardausgabe nach $1, Standardeingabe aus $2,
# Fehlerstrom nach $tmp/fehler.
fahre() {
	aus=$1
	ein=$2
	shift 2
	"$@" <"$ein" >"$aus" 2>"$tmp/fehler" &
	wait $!
}

# rot meldet die Prüfung $1, die Datei $2 und den Fehlerstrom des letzten Aufrufs, dazu das Log des Servers, und endet mit 1, bei nicht
# erreichbarem Daemon mit 2.
rot() {
	melde "rot: $1"
	[ -z "${2:-}" ] || cat "$2" >&2
	cat "$tmp/fehler" >&2
	if ! fahre "$tmp/daemon" /dev/null $DOCKER version; then
		melde "Daemon nicht erreichbar"
		ende 2
	fi
	[ -z "$server" ] || $DOCKER logs "$SERVER" >&2
	ende 1
}

# rufe fährt `docker $@`; scheitert der Aufruf, ist der Lauf rot. Die Standardausgabe steht in $tmp/aus.
rufe() {
	fahre "$tmp/aus" /dev/null $DOCKER "$@" || rot "docker $1" "$tmp/aus"
}

# bau baut mit den Argumenten $@; ein Fehler ist ein Baufehler.
bau() {
	fahre "$tmp/bau" /dev/null $DOCKER build --progress=plain --build-arg VERSION="$VERSION" "$@" . && return 0
	cat "$tmp/bau" "$tmp/fehler" >&2
	melde "Bau gescheitert: $*"
	ende 2
}

echo "test-image: Lauf $lauf"
bau --target image -t "$IMAGE"
bau --target integration -t "$CLIENT"

# Die Prüfungen am Image (SPEC-323 Nr. 3).
rufe image inspect --format 'user={{.Config.User}}{{println}}entrypoint={{json .Config.Entrypoint}}{{println}}cmd={{json .Config.Cmd}}{{println}}volumes={{json .Config.Volumes}}{{range .Config.Env}}{{println}}env={{.}}{{end}}' "$IMAGE"
cp "$tmp/aus" "$tmp/inspect"
grep -qxF 'user=65532:65532' "$tmp/inspect" || rot "Nutzer" "$tmp/inspect"
grep -qxF 'entrypoint=["/odata-recorder"]' "$tmp/inspect" || rot "Eintrittspunkt" "$tmp/inspect"
grep -qxF 'cmd=["start"]' "$tmp/inspect" || rot "Standardargument" "$tmp/inspect"
grep -qx 'volumes=null\|volumes={}' "$tmp/inspect" || rot "Volume" "$tmp/inspect"
grep -qxF 'env=DATA_DIR=/data' "$tmp/inspect" || rot "DATA_DIR" "$tmp/inspect"
for variable in LISTEN ADMIN_LISTEN MODE; do
	! grep -q "^env=$variable=" "$tmp/inspect" || rot "Variable $variable" "$tmp/inspect"
done

rufe create --name "$NAME-pruefung" --label "$LABEL" --network none "$IMAGE"

# eintrag listet den Tar-Strom von `docker cp` des Pfades $1 nach $tmp/liste, mit Eigentümer als Zahl.
eintrag() {
	fahre "$tmp/tar" /dev/null $DOCKER cp "$NAME-pruefung:$1" - || rot "docker cp $1"
	tar --numeric-owner -tvf "$tmp/tar" >"$tmp/liste" 2>"$tmp/fehler" || rot "Tar-Strom $1"
}

eintrag /data
[ "$(awk 'NR == 1 { print $1, $2, $NF }' "$tmp/liste")" = "drwx------ 65532/65532 data/" ] ||
	rot "/data Eigentümer und Rechte" "$tmp/liste"
eintrag /etc/ssl/certs/ca-certificates.crt
awk 'NR == 1 && $1 ~ /^-/ && $3 > 0 { ok = 1 } END { exit !ok }' "$tmp/liste" || rot "Vertrauensanker" "$tmp/liste"
eintrag /tmp
[ "$(awk 'NR == 1 { print $1, $NF }' "$tmp/liste")" = "drwxrwxrwt tmp/" ] || rot "/tmp Rechte" "$tmp/liste"

rufe run --rm --name "$NAME-version" --label "$LABEL" --network none --memory 512m --memory-swap 512m "$IMAGE" version
[ "$(cat "$tmp/aus")" = "odata-recorder $VERSION" ] || rot "Version" "$tmp/aus"
echo "test-image: Image geprüft"

# Gleiches Binary aus zwei Bauten desselben Standes (LH-RB-03). --no-cache-filter build übersetzt neu; erst --no-cache-filter image
# bringt das neue Binary in das zweite Image, sonst liefert der Cache der Stufe image das Binary eines früheren Baus.
bau --no-cache-filter build --no-cache-filter image --target image -t "$ZWEIT"
rufe create --name "$NAME-zweit" --label "$LABEL" --network none "$ZWEIT"
rufe cp "$NAME-pruefung:/odata-recorder" "$tmp/binary-1"
rufe cp "$NAME-zweit:/odata-recorder" "$tmp/binary-2"
cmp "$tmp/binary-1" "$tmp/binary-2" >"$tmp/aus" 2>"$tmp/fehler" || rot "Binary der zwei Bauten ungleich" "$tmp/aus"
echo "test-image: zwei Bauten, Binary byte-gleich"

# Der Ablauf (SPEC-323 Nr. 4).
rufe create --name "$SERVER" --label "$LABEL" --network none --memory 512m --memory-swap 512m \
	--env MODE=record --env UPSTREAM_URL=http://127.0.0.1:9000 "$IMAGE"
server=1
rufe start "$SERVER"

# Das Token geht allein von Datei zu Datei: in keine Ausgabe, kein Argument und keine Variable.
versuch=1
until fahre "$tmp/tar" /dev/null $DOCKER cp "$SERVER:/data/admin.token" - && tar -xOf "$tmp/tar" >"$tmp/token" 2>"$tmp/fehler"; do
	rufe inspect --format '{{.State.Running}}' "$SERVER"
	[ "$(cat "$tmp/aus")" = true ] || rot "Server beendet, kein Token"
	[ "$versuch" -lt 30 ] || rot "kein Token nach 30 Versuchen"
	versuch=$((versuch + 1))
	fahre /dev/null /dev/null sleep 1
done

# phase fährt den Test $2 als Phase $1 im Netz-Namensraum des Servers, mit dem Token auf der Standardeingabe und den Optionen $3…; grün
# bei Exit 0 mit `--- PASS: <Name>` und ohne `--- SKIP` und `--- FAIL` (SPEC-323 Nr. 4 e).
phase() {
	rolle=$1
	name=$2
	shift 2
	fahre "$tmp/phase" "$tmp/token" $DOCKER run --rm -i --name "$NAME-$rolle" --label "$LABEL" --network "container:$SERVER" \
		--memory 1g --memory-swap 1g "$@" "$CLIENT" /out/e2e.test -test.run "^$name\$" -test.v -test.count=1 -test.timeout 120s
	rc=$?
	grund=
	if [ "$rc" -ne 0 ]; then
		grund="Exit $rc"
	elif grep -q '^ *--- SKIP' "$tmp/phase"; then
		grund=SKIP
	elif grep -q '^ *--- FAIL' "$tmp/phase"; then
		grund=FAIL
	elif ! grep -q "^--- PASS: $name " "$tmp/phase"; then
		grund="kein PASS"
	fi
	[ -z "$grund" ] || rot "Phase $rolle ($grund)" "$tmp/phase"
	echo "test-image: Phase $rolle grün"
}

phase aufnehmen TestImageAufnehmen
werte=$(sed -n 's/^.*NEUSTART-WERTE=//p' "$tmp/phase")

rufe stop --time 30 "$SERVER"
rufe inspect --format '{{.State.ExitCode}}' "$SERVER"
[ "$(cat "$tmp/aus")" = 0 ] || rot "Stopp: Exit des Servers" "$tmp/aus"
rufe start "$SERVER"
echo "test-image: Server gestoppt (Exit 0) und gestartet"

phase pruefen TestImagePruefen --env "NEUSTART_WERTE=$werte"
ende 0
