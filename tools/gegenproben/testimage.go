package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Fälle 226 bis 236 (SPEC-102, SPEC-323 Nr. 8): tools/harness/test-image.sh im Verzeichnis des Falls mit einem Ersatz für docker und für
// sleep am Anfang von PATH. Der Ersatz hält jeden Aufruf mit allen Argumenten fest, führt Buch über die Container und Tags, die er
// angelegt hat, liefert die Tar-Ströme mit dem tar des Images und antwortet nach den Variablen TI_*. Weder ein Daemon noch ein Image noch
// Netz; weder go noch make. Dass ein Image die Werte trägt, die der Ersatz liefert, prüft allein `make test-image`.
const tiErsatzDocker = `#!/bin/sh
l=${TI_LOG:?}
s=$l/stand
mkdir -p "$s/c" "$s/i"
{
	echo AUFRUF
	for a in "$@"; do printf 'ARG %s\n' "$a"; done
} >>"$l/docker.log"

zaehle() {
	n=$(cat "$s/z-$1" 2>/dev/null || echo 0)
	n=$((n + 1))
	echo "$n" >"$s/z-$1"
	echo "$n"
}
letzte() {
	for a in "$@"; do :; done
	printf '%s' "$a"
}
wert() {
	o=$1
	shift
	while [ $# -gt 1 ]; do
		[ "$1" != "$o" ] || { printf '%s' "$2"; return; }
		shift
	done
}
art() {
	case $1 in
	*test-image-client-*) echo client ;;
	*test-image-zweit-*) echo zweit ;;
	*) echo image ;;
	esac
}
tarraus() {
	w=$l/tarbau
	rm -rf "$w"
	mkdir -p "$w"
	if [ "$2" = d ]; then mkdir "$w/$1"; else printf '%s' "$6" >"$w/$1"; fi
	tar -C "$w" --numeric-owner --owner="$3" --group="$4" --mode="$5" -cf - "$1"
	rm -rf "$w"
}
signal() {
	[ "${TI_SIG_BEI:-}" = "$1" ] && [ ! -e "$l/signal" ] || return 0
	: >"$l/signal"
	trap 'echo TERM >>"$l/empfangen"; exit 143' TERM
	exec >/dev/null 2>&1
	kill -"$TI_SIG" "$PPID"
	i=0
	while [ "$i" -lt 20 ]; do /bin/sleep 0.1; i=$((i + 1)); done
	exit 0
}

name=
tag=
case ${1:-} in
build) tag=$(wert -t "$@"); ort=bau-$(art "$tag") ;;
image) ort=image-inspect ;;
create) name=$(wert --name "$@"); ort=create-${name##*-} ;;
run) name=$(wert --name "$@"); ort=run-${name##*-} ;;
export) name=$2; ort=export ;;
cp)
	q=${2%%:*}
	case $2 in
	*:/data) ort=cp-data ;;
	*:/etc/ssl/certs/ca-certificates.crt) ort=cp-certs ;;
	*:/tmp) ort=cp-tmp ;;
	*:/data/admin.token) ort=cp-token ;;
	*:/odata-recorder) ort=cp-binary-${q##*-} ;;
	*) ort=cp-fremd ;;
	esac ;;
start) if [ "$(zaehle start)" -eq 1 ]; then ort=start; else ort=neustart; fi ;;
stop) ort=stop ;;
inspect) case $3 in *Running*) ort=inspect-running ;; *) ort=inspect-exit ;; esac ;;
logs) ort=logs ;;
rm) name=$(letzte "$@"); ort=rm-${name##*-} ;;
ps) name=$(letzte "$@"); name=${name#name=^/}; name=${name%\$}; ort=ps-${name##*-} ;;
rmi) tag=$2; ort=rmi-$(art "$tag") ;;
images) tag=$3; ort=images-$(art "$tag") ;;
version) ort=version ;;
*) ort=fremd ;;
esac
signal "$ort"
# Das zweite Signal: der Aufruf läuft danach zu Ende.
if [ "${TI_SIG2_BEI:-}" = "$ort" ] && [ ! -e "$l/signal2" ]; then
	: >"$l/signal2"
	kill -"$TI_SIG2" "$PPID"
fi
case " ${TI_FEHLER:-} " in
*" $ort "*)
	echo "ersatz: $ort scheitert" >&2
	exit 1 ;;
esac
merke=$s/i/$(printf '%s' "$tag" | tr ':/' '__')

case $ort in
bau-*) : >"$merke" ;;
image-inspect)
	eintritt=${TI_ENTRYPOINT-'["/odata-recorder"]'}
	standard=${TI_CMD-'["start"]'}
	printf 'user=%s\nentrypoint=%s\ncmd=%s\nvolumes=%s\n' "${TI_USER-65532:65532}" "$eintritt" "$standard" "${TI_VOLUMES-null}"
	printf 'exposed=%s\nhealthcheck=%s\n' "${TI_EXPOSED-null}" "${TI_HEALTHCHECK-null}"
	printf '%s\n' "${TI_ENV-DATA_DIR=/data}" | sed 's/^/env=/' ;;
export)
	[ -e "$s/c/$name" ] || { echo "No such container: $name" >&2; exit 1; }
	w=$l/tarbau
	rm -rf "$w"
	mkdir -p "$w"
	vorn=${TI_VORN-}
	namen='odata-recorder data/ tmp/ etc/ etc/ssl/ etc/ssl/certs/ etc/ssl/certs/ca-certificates.crt .dockerenv dev/ dev/console
		dev/pts/ dev/shm/ etc/hostname etc/hosts etc/mtab etc/resolv.conf proc/ sys/'
	ohne=${TI_OHNE-}
	extra=${TI_EXTRA-}
	set --
	for n in $vorn $namen $extra; do
		[ "$n" != "$ohne" ] || continue
		case $n in
		*/) mkdir -p "$w/$n" ;;
		*) mkdir -p "$w/$(dirname "$n")"; : >"$w/$n" ;;
		esac
		set -- "$@" "$n"
	done
	tar -C "$w" --no-recursion -cf - "$@"
	rm -rf "$w" ;;
create-*)
	: >"$s/c/$name"
	echo kennung ;;
run-version) printf '%s\n' "${TI_VERSION_AUS-odata-recorder dev}" ;;
run-aufnehmen | run-pruefen)
	cat >"$l/stdin-${ort#run-}"
	if [ "$ort" = run-aufnehmen ]; then
		t=TestImageAufnehmen
		form=${TI_AUFNEHMEN-pass}
		echo "    image_test.go:1: NEUSTART-WERTE=d2VydGU="
	else
		t=TestImagePruefen
		form=${TI_PRUEFEN-pass}
	fi
	echo "=== RUN   $t"
	case $form in
	skip) echo "--- SKIP: $t (0.00s)" ;;
	skip-unter) echo "--- PASS: $t (0.00s)"; echo "    --- SKIP: $t/x (0.00s)" ;;
	fail) echo "--- PASS: $t (0.00s)"; echo "--- FAIL: TestAnderer (0.00s)" ;;
	fail-unter) echo "--- PASS: $t (0.00s)"; echo "    --- FAIL: $t/x (0.00s)" ;;
	kein-pass) : ;;
	anderer-name) echo "--- PASS: ${t}X (0.00s)" ;;
	*) echo "--- PASS: $t (0.00s)" ;;
	esac
	echo PASS
	[ "$form" != exit ] || exit 1 ;;
cp-data) tarraus data d "${TI_DATA_OWNER-65532}" "${TI_DATA_GROUP-65532}" "${TI_DATA_MODE-0700}" ;;
cp-certs) tarraus ca-certificates.crt f 0 0 0644 "${TI_CERT-PEM}" ;;
cp-tmp) tarraus tmp d 0 0 "${TI_TMP_MODE-1777}" ;;
cp-binary-pruefung)
	echo A >"$3"
	touch -t 202601010000.00 "$3" ;;
cp-binary-zweit)
	echo "${TI_BINARY_ZWEIT-A}" >"$3"
	touch -t "${TI_ZEIT_ZWEIT-202601010000.01}" "$3" ;;
cp-token)
	n=$(zaehle token)
	[ "${TI_TOKEN_NACH-0}" != nie ] && [ "$n" -gt "${TI_TOKEN_NACH-0}" ] || exit 1
	tarraus admin.token f 65532 65532 0600 "$TI_TOKEN" ;;
start | neustart | stop) echo "${2:-}" ;;
inspect-running) echo "${TI_RUNNING-true}" ;;
inspect-exit) echo "${TI_EXIT-0}" ;;
logs) echo "ersatz: Log des Servers" ;;
rm-*)
	[ -e "$s/c/$name" ] || { echo "No such container: $name" >&2; exit 1; }
	rm "$s/c/$name" ;;
ps-*) [ ! -e "$s/c/$name" ] || echo kennung ;;
rmi-*)
	[ -e "$merke" ] || { echo "No such image: $tag" >&2; exit 1; }
	rm "$merke" ;;
images-*) [ ! -e "$merke" ] || echo kennung ;;
version) echo 29 ;;
*) exit 99 ;;
esac
`

const tiErsatzSleep = `#!/bin/sh
l=${TI_LOG:?}
printf '%s\n' "$*" >>"$l/sleep.log"
if [ "${TI_SIG_BEI:-}" = sleep ] && [ ! -e "$l/signal" ]; then
	: >"$l/signal"
	trap 'echo TERM >>"$l/empfangen"; exit 143' TERM
	kill -"$TI_SIG" "$PPID"
	i=0
	while [ "$i" -lt 20 ]; do /bin/sleep 0.1; i=$((i + 1)); done
fi
`

const (
	// tiToken ist das Muster, das der Ersatz als Token liefert.
	tiToken = "GEHEIM-TOKEN-MUSTER-5e1f"
	tiWerte = "d2VydGU="
	tiLabel = "odata-recorder.test-image="
)

// tiRollen sind die Rollen der Container eines Laufs in der Folge des Aufräumens, tiArten die drei Tags.
func tiRollen() []string {
	return []string{"pruefung", "version", "zweit", "server", "aufnehmen", "pruefen"}
}

func tiArten() []string { return []string{"image", "client", "zweit"} }

// tiNamen sind die Tags und Container des Laufs mit der Kennung l.
func tiTag(art, l string) string {
	if art == "image" {
		return "odata-recorder:test-image-" + l
	}
	return "odata-recorder:test-image-" + art + "-" + l
}

func tiContainer(rolle, l string) string { return "odr-test-image-" + l + "-" + rolle }

// tiVerbotene sind die 15 Optionen, die kein create und kein run trägt (SPEC-323 Nr. 2 e).
func tiVerbotene() []string {
	return []string{"-v", "--volume", "--mount", "--volumes-from", "-p", "--publish", "-P", "--privileged", "--device", "--pid", "--ipc",
		"--uts", "--userns", "--user", "--entrypoint"}
}

// tiPruefeOptionen liest die Optionen eines create oder run vor dem Image als Liste und nennt jede verbotene Form: eine Option aus
// tiVerbotene und ein --network außer none und container:<server>.
func tiPruefeOptionen(args []string, server string) []string {
	mitWert := []string{"--name", "--label", "--network", "--memory", "--memory-swap", "--env"}
	var befunde []string
	for i := 1; i < len(args) && strings.HasPrefix(args[i], "-"); i++ {
		name, wert, gleich := strings.Cut(args[i], "=")
		if slices.Contains(tiVerbotene(), name) {
			befunde = append(befunde, "verboten: "+name)
		}
		if slices.Contains(mitWert, name) && !gleich && i+1 < len(args) {
			i++
			wert = args[i]
		}
		if name == "--network" && wert != "none" && wert != "container:"+server {
			befunde = append(befunde, "verboten: --network "+wert)
		}
	}
	return befunde
}

// tiSoll sind die Aufrufe eines grünen Laufs mit der Kennung l in ihrer Folge (SPEC-323 Nr. 2 bis 4 und 6): jede Option und jedes
// Argument; `*binary-1` steht für die Datei im Verzeichnis des Laufs.
func tiSoll(l string) [][]string {
	image, client, zweit := tiTag("image", l), tiTag("client", l), tiTag("zweit", l)
	c := func(rolle string) string { return tiContainer(rolle, l) }
	server := c("server")
	bau := func(rest ...string) []string {
		return append(append([]string{"build", "--progress=plain", "--build-arg", "VERSION=dev"}, rest...), ".")
	}
	kopf := func(befehl []string, rolle, netz string) []string {
		return append(befehl, "--name", c(rolle), "--label", tiLabel+l, "--network", netz)
	}
	phase := func(rolle, test string, env ...string) []string {
		a := append(kopf([]string{"run", "--rm", "-i"}, rolle, "container:"+server), "--memory", "1g", "--memory-swap", "1g")
		a = append(a, env...)
		return append(a, client, "/out/e2e.test", "-test.run", "^"+test+"$", "-test.v", "-test.count=1", "-test.timeout", "120s")
	}
	format := "user={{.Config.User}}{{println}}entrypoint={{json .Config.Entrypoint}}{{println}}cmd={{json .Config.Cmd}}" +
		"{{println}}volumes={{json .Config.Volumes}}{{println}}exposed={{json .Config.ExposedPorts}}" +
		"{{println}}healthcheck={{json .Config.Healthcheck}}{{range .Config.Env}}{{println}}env={{.}}{{end}}"
	soll := [][]string{
		bau("--target", "image", "-t", image),
		bau("--target", "integration", "-t", client),
		{"image", "inspect", "--format", format, image},
		append(kopf([]string{"create"}, "pruefung", "none"), image),
		{"cp", c("pruefung") + ":/data", "-"},
		{"cp", c("pruefung") + ":/etc/ssl/certs/ca-certificates.crt", "-"},
		{"cp", c("pruefung") + ":/tmp", "-"},
		{"export", c("pruefung")},
		append(kopf([]string{"run", "--rm"}, "version", "none"), "--memory", "512m", "--memory-swap", "512m", image, "version"),
		bau("--no-cache-filter", "build", "--no-cache-filter", "image", "--target", "image", "-t", zweit),
		append(kopf([]string{"create"}, "zweit", "none"), zweit),
		{"cp", c("pruefung") + ":/odata-recorder", "*binary-1"},
		{"cp", c("zweit") + ":/odata-recorder", "*binary-2"},
		append(kopf([]string{"create"}, "server", "none"), "--memory", "512m", "--memory-swap", "512m",
			"--env", "MODE=record", "--env", "UPSTREAM_URL=http://127.0.0.1:9000", image),
		{"start", server},
		{"cp", server + ":/data/admin.token", "-"},
		phase("aufnehmen", "TestImageAufnehmen"),
		{"stop", "-t", "30", server},
		{"inspect", "--format", "{{.State.ExitCode}}", server},
		{"start", server},
		phase("pruefen", "TestImagePruefen", "--env", "NEUSTART_WERTE="+tiWerte),
	}
	// Aufräumen: rm je Rolle; ein Container, den --rm entfernt hat, scheitert und wird nachgefragt.
	for _, rolle := range tiRollen() {
		soll = append(soll, []string{"rm", "-f", "-v", c(rolle)})
		if rolle == "version" || rolle == "aufnehmen" || rolle == "pruefen" {
			soll = append(soll, []string{"ps", "-a", "-q", "--filter", "name=^/" + c(rolle) + "$"})
		}
	}
	for _, art := range tiArten() {
		soll = append(soll, []string{"rmi", tiTag(art, l)})
	}
	return soll
}

// tiLauf ist ein Aufruf des Treibers mit eigenem Protokoll und eigenem TMPDIR.
type tiLauf struct {
	code        int
	out, errOut string
	aufrufe     [][]string
	log         string
	kennung     string
}

// tiFahre ruft den Treiber unter dem Protokoll name mit env; das TMPDIR des Laufs ist danach leer.
func tiFahre(l Lauf, name string, env []string, args ...string) (tiLauf, string) {
	return tiStarte(l, name, env, append([]string{l.U.TestImageTreiber}, args...)...)
}

// tiStarte ruft `sh` mit args unter dem Protokoll name mit env; das TMPDIR des Laufs ist danach leer. Das Ergebnis trägt Exit, Ströme
// und Aufrufe auch dann, wenn der Grund nicht leer ist.
func tiStarte(l Lauf, name string, env []string, args ...string) (tiLauf, string) {
	var r tiLauf
	basis := filepath.Join(l.Tmp, name)
	r.log = filepath.Join(basis, "log")
	tmp := filepath.Join(basis, "tmp")
	for _, d := range []string{r.log, tmp} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return r, err.Error()
		}
	}
	if err := os.WriteFile(filepath.Join(r.log, "docker.log"), nil, 0o644); err != nil {
		return r, err.Error()
	}
	umgebung := append(slices.Clone(l.Env), "TI_LOG="+r.log, "TMPDIR="+tmp, "TI_TOKEN="+tiToken, "VERSION=", "DOCKER=docker")
	umgebung = append(umgebung, env...)
	var abgelaufen bool
	r.code, r.out, r.errOut, abgelaufen = starteStroeme(fallFrist, l.Dir, umgebung, true, "sh", args...)
	if abgelaufen {
		return r, name + ": Zeit überschritten"
	}
	aufrufe, err := gmProtokoll(basis, "docker.log")
	if err != nil {
		return r, err.Error()
	}
	r.aufrufe = aufrufe
	if len(aufrufe) > 0 {
		if i := slices.Index(aufrufe[0], "-t"); i >= 0 && i+1 < len(aufrufe[0]) {
			r.kennung = strings.TrimPrefix(aufrufe[0][i+1], "odata-recorder:test-image-")
		}
	}
	if g := gmLeer(tmp); g != "" {
		return r, name + ": " + g
	}
	return r, ""
}

// zahl zählt die Aufrufe, deren Anfang die Argumente a sind.
func (r tiLauf) zahl(a ...string) int {
	n := 0
	for _, x := range r.aufrufe {
		if len(x) >= len(a) && slices.Equal(x[:len(a)], a) {
			n++
		}
	}
	return n
}

// aufgeraeumt prüft, dass hinter jedem Ende jeder Container und jedes Tag des Laufs entfernt wurde und der Ersatz keines mehr führt.
func (r tiLauf) aufgeraeumt() string {
	for _, rolle := range tiRollen() {
		if r.zahl("rm", "-f", "-v", tiContainer(rolle, r.kennung)) < 1 {
			return "kein rm für " + rolle
		}
	}
	for _, art := range tiArten() {
		if r.zahl("rmi", tiTag(art, r.kennung)) < 1 {
			return "kein rmi für " + art
		}
	}
	for _, d := range []string{"c", "i"} {
		es, err := os.ReadDir(filepath.Join(r.log, "stand", d))
		if err != nil {
			return err.Error()
		}
		if len(es) > 0 {
			return fmt.Sprintf("%s bleibt: %s", d, es[0].Name())
		}
	}
	return ""
}

// meldungen sind die Zeilen des Treibers auf stderr.
func (r tiLauf) meldungen() []string {
	var m []string
	for _, z := range strings.Split(r.errOut, "\n") {
		if s, ok := strings.CutPrefix(z, "test-image: "); ok {
			m = append(m, s)
		}
	}
	return m
}

// erwarte prüft Exit, die Meldungen des Treibers und das Aufräumen eines Laufs hinter dem ersten Bau; die grüne Zeile steht genau bei
// Exit 0 auf stdout.
func (r tiLauf) erwarte(name string, code int, meldungen ...string) string {
	switch {
	case r.code != code:
		return fmt.Sprintf("%s: Exit %d statt %d\n%s", name, r.code, code, letzte(r.errOut))
	case !slices.Equal(r.meldungen(), meldungen):
		return fmt.Sprintf("%s: Meldungen %q statt %q", name, r.meldungen(), meldungen)
	case strings.Contains(r.out, "test-image: grün\n") != (code == 0):
		return fmt.Sprintf("%s: stdout %q bei Exit %d", name, r.out, r.code)
	}
	if g := r.aufgeraeumt(); g != "" {
		return name + ": " + g
	}
	return ""
}

// tiFall baut einen Fall mit dem Treiber; sein erster Aufruf ist grün, die weiteren fährt pruefe.
func tiFall(name string, pruefe func(Lauf) string) Fall {
	f := gmFall(name, gateTestImage, map[string]string{"bin/docker": tiErsatzDocker, "bin/sleep": tiErsatzSleep, "log/docker.log": ""},
		nil, 0, "1>test-image: grün")
	f.Name = "testimage/" + name
	f.Ausfuehrbar = []string{"bin/docker", "bin/sleep"}
	f.Env = []string{"PATH={dir}/bin:{PATH}", "TI_LOG={dir}/log", "TI_TOKEN=" + tiToken, "VERSION=", "DOCKER=docker"}
	f.Nachher = pruefe
	return f
}

// tiErste fährt die Läufe nacheinander und nennt den ersten Grund.
func tiErste(laeufe []func() string) string {
	for _, f := range laeufe {
		if g := f(); g != "" {
			return g
		}
	}
	return ""
}

// befehlTestImage ruft für gateTestImage den Treiber unter dem Pfad treiber im Verzeichnis des Falls, sonst naechster.
func befehlTestImage(naechster func(string, Fall) (string, []string, string), treiber string) func(string, Fall) (string, []string, string) {
	return func(dir string, f Fall) (string, []string, string) {
		if f.Gate == gateTestImage {
			return "sh", append([]string{treiber}, f.ersetzeOrte(dir)...), dir
		}
		return naechster(dir, f)
	}
}

// tiGut ist Fall 226: die Aufrufe eines grünen Laufs Argument für Argument, keine verbotene Form, und das Token in keiner Ausgabe und
// keinem Argument, aber auf der Standardeingabe beider Phasen.
func tiGut() Fall {
	return tiFall("gut", func(l Lauf) string {
		r, g := tiErwartet(l, "gut", nil, 0)
		if g == "" {
			g = tiVergleiche(tiOhneVerzeichnis(r.aufrufe), tiSoll(r.kennung))
		}
		if g == "" {
			g = tiKeineVerbotene(r)
		}
		if g == "" {
			g = tiOhneToken(r)
		}
		return g
	})
}

// tiErwartet fährt den Lauf name mit env und prüft Exit, Meldungen und Aufräumen.
func tiErwartet(l Lauf, name string, env []string, code int, meldungen ...string) (tiLauf, string) {
	r, g := tiFahre(l, name, env)
	if g == "" {
		g = r.erwarte(name, code, meldungen...)
	}
	return r, g
}

// tiOhneVerzeichnis ersetzt die Datei eines Binary im Verzeichnis des Laufs durch `*<Name>`.
func tiOhneVerzeichnis(aufrufe [][]string) [][]string {
	out := make([][]string, len(aufrufe))
	for i, a := range aufrufe {
		out[i] = slices.Clone(a)
		for j, x := range out[i] {
			if b := filepath.Base(x); strings.HasPrefix(b, "binary-") && filepath.IsAbs(x) {
				out[i][j] = "*" + b
			}
		}
	}
	return out
}

// tiAufrufNr ist der Aufruf i der Liste, hinter ihrem Ende keiner.
func tiAufrufNr(aufrufe [][]string, i int) []string {
	if i < len(aufrufe) {
		return aufrufe[i]
	}
	return nil
}

// tiVergleiche nennt den ersten Aufruf, der abweicht, fehlt oder zusätzlich ist.
func tiVergleiche(got, soll [][]string) string {
	for i := range max(len(got), len(soll)) {
		if a, b := tiAufrufNr(got, i), tiAufrufNr(soll, i); !slices.Equal(a, b) {
			return fmt.Sprintf("Aufruf %d: %q statt %q", i+1, a, b)
		}
	}
	return ""
}

// tiKeineVerbotene prüft jedes der sechs create und run eines grünen Laufs auf die verbotenen Formen.
func tiKeineVerbotene(r tiLauf) string {
	n := 0
	for _, a := range r.aufrufe {
		if a[0] != "create" && a[0] != "run" {
			continue
		}
		n++
		if b := tiPruefeOptionen(a, tiContainer("server", r.kennung)); len(b) > 0 {
			return strings.Join(b, "; ")
		}
	}
	if n != 6 {
		return fmt.Sprintf("%d create und run statt 6", n)
	}
	return ""
}

// tiOhneToken prüft SPEC-323 Nr. 4 a: das Muster kommt auf der Standardeingabe beider Phasen an und steht sonst nirgends.
func tiOhneToken(r tiLauf) string {
	for _, rolle := range []string{"aufnehmen", "pruefen"} {
		b, err := os.ReadFile(filepath.Join(r.log, "stdin-"+rolle))
		if err != nil || string(b) != tiToken {
			return fmt.Sprintf("Standardeingabe von %s: %q, %v", rolle, b, err)
		}
	}
	if strings.Contains(r.out, tiToken) || strings.Contains(r.errOut, tiToken) {
		return "Token in der Ausgabe"
	}
	for _, a := range r.aufrufe {
		for _, x := range a {
			if strings.Contains(x, tiToken) {
				return fmt.Sprintf("Token als Argument: %q", a)
			}
		}
	}
	return ""
}

// tiZweiLaeufe ist Fall 227: zwei Läufe haben getrennte Namen, jeder Name trägt die Kennung seines Laufs, jeder Container das Label.
func tiZweiLaeufe() Fall {
	return tiFall("zwei-laeufe", func(l Lauf) string {
		var kennungen []string
		for i := range 2 {
			name := "lauf-" + strconv.Itoa(i)
			r, g := tiErwartet(l, name, nil, 0)
			if g == "" {
				g = tiNamenDesLaufs(name, r)
			}
			if g != "" {
				return g
			}
			kennungen = append(kennungen, r.kennung)
		}
		if kennungen[0] == kennungen[1] {
			return "zwei Läufe mit der Kennung " + kennungen[0]
		}
		return ""
	})
}

// tiOptionswerte sind die Werte der Option in allen Aufrufen.
func tiOptionswerte(aufrufe [][]string, option string) []string {
	var werte []string
	for _, a := range aufrufe {
		for j := 0; j+1 < len(a); j++ {
			if a[j] == option {
				werte = append(werte, a[j+1])
			}
		}
	}
	return werte
}

// tiNamenDesLaufs prüft, dass jeder Name die Kennung des Laufs trägt und die sechs Label sie nennen.
func tiNamenDesLaufs(name string, r tiLauf) string {
	if r.kennung == "" || strings.ContainsAny(r.kennung, ":/ ") {
		return fmt.Sprintf("%s: Kennung %q", name, r.kennung)
	}
	labels := tiOptionswerte(r.aufrufe, "--label")
	if len(labels) != 6 {
		return fmt.Sprintf("%s: %d Label statt 6", name, len(labels))
	}
	for _, w := range labels {
		if w != tiLabel+r.kennung {
			return fmt.Sprintf("%s: Label %q", name, w)
		}
	}
	for _, w := range tiOptionswerte(r.aufrufe, "--name") {
		if !strings.HasPrefix(w, "odr-test-image-"+r.kennung+"-") {
			return fmt.Sprintf("%s: Name %q", name, w)
		}
	}
	return ""
}

// tiAufruf ist Fall 228: ein Argument, auch ein leeres, ist Exit 2 ohne einen Aufruf von docker; VERSION leer und ungesetzt ist dev (Fall
// 226 fährt sie leer), eine gesetzte VERSION erreicht Bau und Vergleich.
func tiAufruf() Fall {
	return tiFall("aufruf", func(l Lauf) string {
		for i, a := range []string{"", "x"} {
			name := "argument-" + strconv.Itoa(i)
			r, g := tiFahre(l, name, nil, a)
			if g != "" {
				return g
			}
			if r.code != 2 || !slices.Equal(r.meldungen(), []string{"Aufruf: tools/harness/test-image.sh (ohne Argument)"}) || len(r.aufrufe) != 0 {
				return fmt.Sprintf("%s: Exit %d, %q, %d Aufrufe", name, r.code, r.meldungen(), len(r.aufrufe))
			}
		}
		r, g := tiFahre(l, "version", []string{"VERSION=1.2.3", "TI_VERSION_AUS=odata-recorder 1.2.3"})
		if g == "" {
			g = r.erwarte("version", 0)
		}
		if g == "" && r.zahl("build", "--progress=plain", "--build-arg", "VERSION=1.2.3") != 3 {
			g = "VERSION erreicht nicht jeden der drei Bauten"
		}
		return g
	})
}

// tiImageBild ist eine Abweichung des Image und der Name der Prüfung, die sie rot meldet; zeile steht, wenn gesetzt, als eigene Zeile
// in der Fehlerausgabe.
type tiImageBild struct{ name, env, pruefung, zeile string }

// tiAngabenBilder führt je Angabe aus `image inspect` ihre Abweichungen; eine leere Liste ist bei den Ports erlaubt, beim Healthcheck
// nicht.
func tiAngabenBilder() []tiImageBild {
	return []tiImageBild{
		{name: "nutzer-root", env: "TI_USER=0", pruefung: "Nutzer"},
		{name: "nutzer-leer", env: "TI_USER=", pruefung: "Nutzer"},
		{name: "nutzer-gruppe", env: "TI_USER=65532:0", pruefung: "Nutzer"},
		{name: "eintritt", env: `TI_ENTRYPOINT=["/odata-recorder","start"]`, pruefung: "Eintrittspunkt"},
		{name: "eintritt-fehlt", env: "TI_ENTRYPOINT=null", pruefung: "Eintrittspunkt"},
		{name: "cmd-fehlt", env: "TI_CMD=null", pruefung: "Standardargument"},
		{name: "cmd-anders", env: `TI_CMD=["version"]`, pruefung: "Standardargument"},
		{name: "volume", env: `TI_VOLUMES={"/data":{}}`, pruefung: "Volume"},
		{name: "expose", env: `TI_EXPOSED={"8080/tcp":{}}`, pruefung: "EXPOSE"},
		{name: "healthcheck", env: `TI_HEALTHCHECK={"Test":["CMD","/odata-recorder","version"]}`, pruefung: "HEALTHCHECK"},
		{name: "healthcheck-none", env: `TI_HEALTHCHECK={"Test":["NONE"]}`, pruefung: "HEALTHCHECK"},
		{name: "healthcheck-leer", env: `TI_HEALTHCHECK={}`, pruefung: "HEALTHCHECK"},
		{name: "data-dir-fehlt", env: "TI_ENV=PATH=/bin", pruefung: "DATA_DIR"},
		{name: "data-dir-anders", env: "TI_ENV=DATA_DIR=/data2", pruefung: "DATA_DIR"},
	}
}

// tiFremdeVariablen sind sechs Namen und je erlaubtem Namen (DATA_DIR, PATH) sechs, die ihm gleichen: länger vorn, hinten und mit `_`,
// kürzer vorn und hinten, klein geschrieben. Keiner ist DATA_DIR oder PATH.
func tiFremdeVariablen() []string {
	namen := []string{"LISTEN", "ADMIN_LISTEN", "MODE", "UPSTREAM_URL", "MODEL", "LISTEN_TLS_CERT"}
	for _, e := range []string{"DATA_DIR", "PATH"} {
		namen = append(namen, e+"X", "X"+e, e+"_", e[:len(e)-1], e[1:], strings.ToLower(e))
	}
	return namen
}

// tiVariablenBilder führt je fremdem Namen eine Abweichung mit leerem Wert vor und eine mit Wert hinter den erlaubten Variablen, dazu
// eine Variable, deren Wert wie eine erlaubte aussieht.
func tiVariablenBilder() []tiImageBild {
	var bilder []tiImageBild
	for _, v := range tiFremdeVariablen() {
		bilder = append(bilder,
			tiImageBild{name: "variable-vorn-" + v, env: "TI_ENV=" + v + "=\nDATA_DIR=/data\nPATH=/bin", pruefung: "Variable " + v},
			tiImageBild{name: "variable-hinten-" + v, env: "TI_ENV=DATA_DIR=/data\nPATH=/bin\n" + v + "=x", pruefung: "Variable " + v})
	}
	return append(bilder, tiImageBild{name: "variable-wert", env: "TI_ENV=DATA_DIR=/data\nFOO=PATH=x", pruefung: "Variable FOO"})
}

// tiErlaubteNamen sind die 18 Namen von SPEC-323 Nr. 3 *Inhalt*: sieben aus dem Image, elf von der Laufzeit.
func tiErlaubteNamen() []string {
	return []string{"odata-recorder", "data/", "tmp/", "etc/", "etc/ssl/", "etc/ssl/certs/", "etc/ssl/certs/ca-certificates.crt",
		".dockerenv", "dev/", "dev/console", "dev/pts/", "dev/shm/", "etc/hostname", "etc/hosts", "etc/mtab", "etc/resolv.conf",
		"proc/", "sys/"}
}

// tiFremdeNamen sind je Ort ein zusätzlicher Name, je erlaubtem Namen einer, der ihn als Anfang, und einer, der ihn als Ende trägt, und
// je erlaubtem Namen mit Punkt einer mit `x` an dessen Stelle.
func tiFremdeNamen() []string {
	namen := []string{"bin/sh", "data/seed", "tmp/rest", "etc/passwd", "etc/ssl/certs/anderes.pem", "dev/fuse", "var/"}
	for _, e := range tiErlaubteNamen() {
		namen = append(namen, e+"x", "x"+e)
		if strings.Contains(e, ".") {
			namen = append(namen, strings.ReplaceAll(e, ".", "x"))
		}
	}
	return namen
}

// tiInhaltBilder führt je fremdem Namen eine Abweichung am Ende des Exports, dazu eine am Anfang.
func tiInhaltBilder() []tiImageBild {
	var bilder []tiImageBild
	for i, n := range tiFremdeNamen() {
		bilder = append(bilder, tiImageBild{name: "inhalt-" + strconv.Itoa(i), env: "TI_EXTRA=" + n, pruefung: "Inhalt", zeile: n})
	}
	return append(bilder, tiImageBild{name: "inhalt-vorn", env: "TI_VORN=bin/sh", pruefung: "Inhalt", zeile: "bin/sh"})
}

// tiOrtsBilder führt die Abweichungen an den Dateien, an der Version und an den zwei Binaries. Das erste Binary des Ersatzes ist von
// 202601010000.00, das zweite eine Sekunde jünger; die Formen des Alters setzen das zweite gleich alt oder eine Sekunde älter.
func tiOrtsBilder() []tiImageBild {
	return []tiImageBild{
		{name: "data-eigentuemer", env: "TI_DATA_OWNER=0", pruefung: "/data Eigentümer und Rechte"},
		{name: "data-gruppe", env: "TI_DATA_GROUP=0", pruefung: "/data Eigentümer und Rechte"},
		{name: "data-rechte", env: "TI_DATA_MODE=0755", pruefung: "/data Eigentümer und Rechte"},
		{name: "data-fehlt", env: "TI_FEHLER=cp-data", pruefung: "docker cp /data"},
		{name: "anker-leer", env: "TI_CERT=", pruefung: "Vertrauensanker"},
		{name: "anker-fehlt", env: "TI_FEHLER=cp-certs", pruefung: "docker cp /etc/ssl/certs/ca-certificates.crt"},
		{name: "tmp-ohne-sticky", env: "TI_TMP_MODE=0777", pruefung: "/tmp Rechte"},
		{name: "tmp-rechte", env: "TI_TMP_MODE=1755", pruefung: "/tmp Rechte"},
		{name: "tmp-fehlt", env: "TI_FEHLER=cp-tmp", pruefung: "docker cp /tmp"},
		{name: "version", env: "TI_VERSION_AUS=odata-recorder 0.0.0", pruefung: "Version"},
		{name: "binary-ungleich", env: "TI_BINARY_ZWEIT=B", pruefung: "Binary der zwei Bauten ungleich"},
		{name: "binary-gleich-alt", env: "TI_ZEIT_ZWEIT=202601010000.00", pruefung: "zweites Binary nicht neu übersetzt"},
		{name: "binary-aelter", env: "TI_ZEIT_ZWEIT=202512312359.59", pruefung: "zweites Binary nicht neu übersetzt"},
	}
}

// tiImageBilder sind alle Abweichungen des Image (SPEC-323 Nr. 3).
func tiImageBilder() []tiImageBild {
	return slices.Concat(tiAngabenBilder(), tiVariablenBilder(), tiOrtsBilder(), tiInhaltBilder())
}

// tiImage ist Fall 229: jede Abweichung ist Exit 1 mit dem Namen ihrer Prüfung, vor dem Server und ohne Log; die erlaubten Formen sind
// grün.
func tiImage() Fall {
	return tiFall("image", func(l Lauf) string {
		bilder := tiImageBilder()
		if len(bilder) != 111 {
			return fmt.Sprintf("%d Abweichungen statt 111", len(bilder))
		}
		for _, b := range bilder {
			if g := tiImageRot(l, b); g != "" {
				return g
			}
		}
		return tiImageErlaubt(l)
	})
}

// tiImageRot ist die Abweichung b: rot mit ihrem Namen, ohne Server und ohne Log.
func tiImageRot(l Lauf, b tiImageBild) string {
	r, g := tiErwartet(l, b.name, []string{b.env}, 1, "rot: "+b.pruefung)
	if g == "" && (r.zahl("create", "--name", tiContainer("server", r.kennung)) != 0 || r.zahl("logs") != 0) {
		g = b.name + ": Server oder Log trotz roter Prüfung am Image"
	}
	if g == "" && b.zeile != "" && !strings.Contains(r.errOut, "\n"+b.zeile+"\n") {
		g = fmt.Sprintf("%s: %q nicht in der Ausgabe: %q", b.name, b.zeile, r.errOut)
	}
	return g
}

// tiErlaubteFormen sind PATH neben DATA_DIR mit einem `=` im Wert, je eine leere Liste der Volumes und der Ports und ein Export ohne
// einen der Namen.
func tiErlaubteFormen() []string {
	return []string{"TI_ENV=PATH=/a=b\nDATA_DIR=/data", "TI_VOLUMES={}", "TI_EXPOSED={}", "TI_OHNE=proc/"}
}

// tiImageErlaubt fährt jede erlaubte Form grün.
func tiImageErlaubt(l Lauf) string {
	for i, env := range tiErlaubteFormen() {
		if _, g := tiErwartet(l, "erlaubt-"+strconv.Itoa(i), []string{env}, 0); g != "" {
			return g
		}
	}
	return ""
}

// tiForm ist die Ausgabe einer Phase und der Grund, mit dem der Treiber sie rot wertet (SPEC-323 Nr. 4 e).
type tiForm struct{ form, grund string }

func tiPhasenFormen() []tiForm {
	return []tiForm{{"exit", "Exit 1"}, {"skip", "SKIP"}, {"skip-unter", "SKIP"}, {"fail", "FAIL"}, {"fail-unter", "FAIL"},
		{"kein-pass", "kein PASS"}, {"anderer-name", "kein PASS"}}
}

// tiPhasen ist Fall 230: jede Form ist in jeder der zwei Phasen Exit 1 mit Log des Servers, und die spätere Phase läuft nicht.
func tiPhasen() Fall {
	return tiFall("phasen", func(l Lauf) string {
		formen := tiPhasenFormen()
		if len(formen) != 7 {
			return fmt.Sprintf("%d Formen statt 7", len(formen))
		}
		for _, phase := range []string{"aufnehmen", "pruefen"} {
			for _, f := range formen {
				if g := tiPhaseRot(l, phase, f); g != "" {
					return g
				}
			}
		}
		return ""
	})
}

// tiPhaseRot ist die Form f in der Phase: rot mit Grund, Ausgabe der Phase und Log des Servers; hinter der roten Phase aufnehmen kein
// Stopp.
func tiPhaseRot(l Lauf, phase string, f tiForm) string {
	name := phase + "-" + f.form
	r, g := tiErwartet(l, name, []string{"TI_" + strings.ToUpper(phase) + "=" + f.form}, 1, "rot: Phase "+phase+" ("+f.grund+")")
	switch {
	case g != "":
		return g
	case r.zahl("logs", tiContainer("server", r.kennung)) != 1:
		return name + ": kein Log des Servers"
	case phase == "aufnehmen" && r.zahl("stop") != 0:
		return name + ": Stopp hinter der roten Phase"
	case !strings.Contains(r.errOut, "=== RUN") || !strings.Contains(r.errOut, "ersatz: Log des Servers"):
		return name + ": Ausgabe der Phase oder des Servers fehlt"
	}
	return ""
}

// tiStopp ist Fall 231: ein Exit des Servers ungleich 0 ist rot, ohne zweiten Start und ohne Phase pruefen (SPEC-323 Nr. 4 c).
func tiStopp() Fall {
	return tiFall("stopp", func(l Lauf) string {
		for _, exit := range []string{"137", "143", "1", "2", ""} {
			name := "exit-" + exit
			r, g := tiFahre(l, name, []string{"TI_EXIT=" + exit})
			if g == "" {
				g = r.erwarte(name, 1, "rot: Stopp: Exit des Servers")
			}
			if g == "" && (r.zahl("start") != 1 || r.zahl("logs") != 1 || r.zahl("run", "--rm", "-i") != 1) {
				g = fmt.Sprintf("%s: %d start, %d logs, %d Phasen", name, r.zahl("start"), r.zahl("logs"), r.zahl("run", "--rm", "-i"))
			}
			if g != "" {
				return g
			}
		}
		return ""
	})
}

// tiTokenLauf ist ein Lauf der Token-Abholung: Versuche, Pausen, Exit und Meldungen.
type tiTokenLauf struct {
	name           string
	env            []string
	code           int
	versuche, schl int
	meldungen      []string
}

func tiTokenLaeufe() []tiTokenLauf {
	nach30 := []string{"rot: kein Token nach 30 Versuchen"}
	return []tiTokenLauf{
		{"spaeter", []string{"TI_TOKEN_NACH=2"}, 0, 3, 2, nil},
		{"letzter", []string{"TI_TOKEN_NACH=29"}, 0, 30, 29, nil},
		{"zu-spaet", []string{"TI_TOKEN_NACH=30"}, 1, 30, 29, nach30},
		{"nie", []string{"TI_TOKEN_NACH=nie"}, 1, 30, 29, nach30},
		{"beendet", []string{"TI_TOKEN_NACH=nie", "TI_RUNNING=false"}, 1, 1, 0, []string{"rot: Server beendet, kein Token"}},
	}
}

// tiTokenFall ist Fall 232: höchstens 30 Versuche im Abstand von 1 s, solange der Server läuft (SPEC-323 Nr. 4 a).
func tiTokenFall() Fall {
	return tiFall("token", func(l Lauf) string {
		for _, t := range tiTokenLaeufe() {
			if g := tiTokenAbholung(l, t); g != "" {
				return g
			}
		}
		return ""
	})
}

func tiTokenAbholung(l Lauf, t tiTokenLauf) string {
	r, g := tiErwartet(l, t.name, t.env, t.code, t.meldungen...)
	if g != "" {
		return g
	}
	b, err := os.ReadFile(filepath.Join(r.log, "sleep.log"))
	if err != nil && !os.IsNotExist(err) {
		return err.Error()
	}
	cp := r.zahl("cp", tiContainer("server", r.kennung)+":/data/admin.token", "-")
	switch {
	case cp != t.versuche || string(b) != strings.Repeat("1\n", t.schl):
		return fmt.Sprintf("%s: %d Versuche, Pausen %q", t.name, cp, b)
	case t.code == 0:
		if g := tiOhneToken(r); g != "" {
			return t.name + ": " + g
		}
	}
	return ""
}

// tiOrte sind die Aufrufe von docker ausserhalb des Baus und der Phasen, mit dem Namen, unter dem ihr Scheitern rot ist.
func tiOrte() [][2]string {
	return [][2]string{
		{"image-inspect", "docker image"}, {"create-pruefung", "docker create"}, {"cp-data", "docker cp /data"},
		{"cp-certs", "docker cp /etc/ssl/certs/ca-certificates.crt"}, {"cp-tmp", "docker cp /tmp"}, {"export", "docker export"},
		{"run-version", "docker run"},
		{"create-zweit", "docker create"}, {"cp-binary-pruefung", "docker cp"}, {"cp-binary-zweit", "docker cp"},
		{"create-server", "docker create"}, {"start", "docker start"}, {"stop", "docker stop"}, {"inspect-exit", "docker inspect"},
		{"neustart", "docker start"},
	}
}

// tiExitCodes ist Fall 233: ein Fehler eines Baus ist Exit 2; ein anderer Aufruf, der scheitert, ist Exit 1, bei nicht erreichbarem
// Daemon Exit 2 (SPEC-323 Nr. 6 c).
func tiExitCodes() Fall {
	return tiFall("exit-codes", func(l Lauf) string {
		var laeufe []func() string
		lauf := func(name string, env []string, code int, meldungen ...string) {
			laeufe = append(laeufe, func() string {
				r, g := tiFahre(l, name, env)
				if g == "" {
					g = r.erwarte(name, code, meldungen...)
				}
				return g
			})
		}
		for _, art := range tiArten() {
			laeufe = append(laeufe, func() string {
				r, g := tiFahre(l, "bau-"+art, []string{"TI_FEHLER=bau-" + art})
				if g != "" {
					return g
				}
				m := r.meldungen()
				if len(m) != 1 || !strings.HasPrefix(m[0], "Bau gescheitert: ") || !strings.HasSuffix(m[0], tiTag(art, r.kennung)) {
					return fmt.Sprintf("bau-%s: Meldungen %q", art, m)
				}
				return r.erwarte("bau-"+art, 2, m...)
			})
		}
		for _, o := range tiOrte() {
			lauf(o[0], []string{"TI_FEHLER=" + o[0]}, 1, "rot: "+o[1])
			lauf(o[0]+"-daemon", []string{"TI_FEHLER=" + o[0] + " version"}, 2, "rot: "+o[1], "Daemon nicht erreichbar")
		}
		lauf("phase-daemon", []string{"TI_AUFNEHMEN=exit", "TI_FEHLER=version"}, 2, "rot: Phase aufnehmen (Exit 1)", "Daemon nicht erreichbar")
		if len(laeufe) != 3+2*15+1 {
			return fmt.Sprintf("%d Läufe statt 34", len(laeufe))
		}
		return tiErste(laeufe)
	})
}

// tiRest ist ein Lauf mit einem Objekt, das bleibt, und seinen Meldungen; {L} steht für die Kennung.
type tiRest struct {
	name           string
	env, meldungen []string
}

// tiReste sind 14 Läufe: je Container und Tag ein Entfernen oder eine Nachfrage, die scheitert, und zwei auf rotem Lauf. Ein Tag zu
// einem Container, der bleibt, entfernt der Ersatz dennoch.
func tiReste() []tiRest {
	const g = "Aufräumen gescheitert: "
	var reste []tiRest
	for _, rolle := range []string{"pruefung", "zweit", "server"} {
		reste = append(reste, tiRest{"rm-" + rolle, []string{"TI_FEHLER=rm-" + rolle}, []string{g + "odr-test-image-{L}-" + rolle}})
	}
	for _, rolle := range []string{"version", "aufnehmen", "pruefen"} {
		env := []string{"TI_FEHLER=rm-" + rolle + " ps-" + rolle}
		reste = append(reste, tiRest{"ps-" + rolle, env, []string{g + "odr-test-image-{L}-" + rolle}})
	}
	for _, art := range tiArten() {
		meldung := []string{g + tiTag(art, "{L}")}
		reste = append(reste, tiRest{"rmi-" + art, []string{"TI_FEHLER=rmi-" + art}, meldung},
			tiRest{"images-" + art, []string{"TI_FEHLER=rmi-" + art + " images-" + art}, meldung})
	}
	return append(reste,
		tiRest{"rot-und-rest", []string{"TI_PRUEFEN=fail", "TI_FEHLER=rm-server"},
			[]string{"rot: Phase pruefen (FAIL)", g + "odr-test-image-{L}-server"}},
		tiRest{"rot-und-zwei", []string{"TI_EXIT=1", "TI_FEHLER=rm-pruefung rmi-zweit"},
			[]string{"rot: Stopp: Exit des Servers", g + "odr-test-image-{L}-pruefung", g + tiTag("zweit", "{L}")}})
}

// tiBleibt ist der Lauf x: Exit 1, die Meldungen, jedes Objekt wird dennoch entfernt und nichts ist grün.
func tiBleibt(l Lauf, x tiRest) string {
	r, g := tiFahre(l, x.name, x.env)
	if g != "" {
		return g
	}
	var soll []string
	for _, m := range x.meldungen {
		soll = append(soll, strings.ReplaceAll(m, "{L}", r.kennung))
	}
	switch {
	case r.code != 1 || !slices.Equal(r.meldungen(), soll):
		return fmt.Sprintf("%s: Exit %d, Meldungen %q statt 1 %q", x.name, r.code, r.meldungen(), soll)
	case r.zahl("rm", "-f", "-v")+r.zahl("rmi") != 9:
		return fmt.Sprintf("%s: %d Aufrufe rm und rmi statt 9", x.name, r.zahl("rm", "-f", "-v")+r.zahl("rmi"))
	case strings.Contains(r.out, "test-image: grün"):
		return x.name + ": grün gemeldet"
	}
	return ""
}

// tiGeschlossen ruft `sh "$@"` mit dem Strom $0 (1 oder 2) auf dem Schreibende einer Pipe, deren Leser vorher geschlossen ist. Beendet
// ein Schreiben darauf eine Sub-Shell nicht mit SIGPIPE (141), endet der Aufruf mit 97, ohne den Treiber zu rufen.
const tiGeschlossen = `p=$TI_LOG/pipe
mkfifo "$p" || exit 97
exec 3<>"$p" 4>"$p" 3<&-
( echo >&4 ) 2>/dev/null
[ $? -eq 141 ] || exit 97
if [ "$0" = 1 ]; then exec sh "$@" 1>&4 4>&-; fi
exec sh "$@" 2>&4 4>&-`

// tiWeg ist ein Weg des Treibers: grün, rot oder Baufehler.
type tiWeg struct {
	name      string
	env       []string
	code      int
	meldungen []string
}

func tiWege() []tiWeg {
	return []tiWeg{
		{"gruen", nil, 0, nil},
		{"rot", []string{"TI_USER=0"}, 1, []string{"rot: Nutzer"}},
		{"baufehler", []string{"TI_FEHLER=bau-image"}, 2, nil},
	}
}

// ohneKennung ersetzt die Kennung des Laufs in text durch `L`.
func (r tiLauf) ohneKennung(text string) string {
	if r.kennung == "" {
		return text
	}
	return strings.ReplaceAll(text, r.kennung, "L")
}

// gestalt sind die Aufrufe eines Laufs ohne seine Kennung und sein Verzeichnis.
func (r tiLauf) gestalt() string {
	return r.ohneKennung(fmt.Sprintf("%q", tiOhneVerzeichnis(r.aufrufe)))
}

// tiOhnePipe fährt den Weg w mit offenen Strömen und prüft ihn.
func tiOhnePipe(l Lauf, w tiWeg) (tiLauf, string) {
	r, g := tiFahre(l, w.name, w.env)
	if g != "" {
		return r, g
	}
	meldungen := w.meldungen
	if w.code == 2 {
		meldungen = r.meldungen()
		if len(meldungen) != 1 || !strings.HasPrefix(meldungen[0], "Bau gescheitert: ") {
			return r, fmt.Sprintf("%s: Meldungen %q", w.name, meldungen)
		}
	}
	return r, r.erwarte(w.name, w.code, meldungen...)
}

// tiMitPipe fährt den Weg w mit dem geschlossenen Strom und hält ihn gegen den Lauf offen mit offenen Strömen: derselbe Exit, dieselben
// Aufrufe, derselbe Inhalt des offenen Stroms (Meldungen des Treibers bzw. Standardausgabe), nichts bleibt (SPEC-323 Nr. 6 e).
func tiMitPipe(l Lauf, w tiWeg, strom string, offen tiLauf) string {
	name := w.name + "-pipe-" + strom
	r, g := tiStarte(l, name, w.env, "-c", tiGeschlossen, strom, l.U.TestImageTreiber)
	switch {
	case r.code != w.code:
		return fmt.Sprintf("%s: Exit %d statt %d (%s)\n%s", name, r.code, w.code, g, letzte(r.errOut))
	case g != "":
		return g
	case r.gestalt() != offen.gestalt():
		return fmt.Sprintf("%s: Aufrufe %s statt %s", name, r.gestalt(), offen.gestalt())
	case strom == "1" && (r.out != "" || !slices.Equal(tiOhneKennungen(r), tiOhneKennungen(offen))):
		return fmt.Sprintf("%s: stdout %q, Meldungen %q", name, r.out, r.meldungen())
	case strom == "2" && (r.errOut != "" || r.ohneKennung(r.out) != offen.ohneKennung(offen.out)):
		return fmt.Sprintf("%s: stderr %q, stdout %q", name, r.errOut, r.out)
	}
	if g := r.aufgeraeumt(); g != "" {
		return name + ": " + g
	}
	return ""
}

// tiOhneKennungen sind die Meldungen des Treibers ohne seine Kennung.
func tiOhneKennungen(r tiLauf) []string {
	m := r.meldungen()
	for i := range m {
		m[i] = r.ohneKennung(m[i])
	}
	return m
}

// tiPipes fährt jeden Weg mit geschlossenem Leser je Strom.
func tiPipes(l Lauf) string {
	wege := tiWege()
	if len(wege) != 3 {
		return fmt.Sprintf("%d Wege statt 3", len(wege))
	}
	for _, w := range wege {
		offen, g := tiOhnePipe(l, w)
		if g != "" {
			return g
		}
		for _, strom := range []string{"1", "2"} {
			if g := tiMitPipe(l, w, strom, offen); g != "" {
				return g
			}
		}
	}
	return ""
}

// tiAufraeumen ist Fall 234: ein vorhandenes Objekt, das sich nicht entfernen lässt, ist eine Meldung und aus grün Exit 1; die übrigen
// werden dennoch entfernt. Ein Objekt, das nie entstand, ist kein Fehler (Fall 226), eine Nachfrage, die scheitert, zählt als vorhanden.
// Ein geschlossener Leser der Standardausgabe oder des Fehlerstroms ändert auf keinem Weg Aufrufe, Exit oder Aufräumen.
func tiAufraeumen() Fall {
	return tiFall("aufraeumen", func(l Lauf) string {
		reste := tiReste()
		if len(reste) != 14 {
			return fmt.Sprintf("%d Läufe statt 14", len(reste))
		}
		for _, x := range reste {
			if g := tiBleibt(l, x); g != "" {
				return g
			}
		}
		return tiPipes(l)
	})
}

// tiSignalOrte sind die Stellen, an denen der Fall ein Signal sendet; an `rm-pruefung` läuft schon das Aufräumen.
func tiSignalOrte() []string {
	return []string{"bau-image", "bau-zweit", "create-server", "cp-token", "sleep", "run-aufnehmen", "stop", "run-pruefen", "rm-pruefung"}
}

// tiZweiteSignale sind die Signale, die während des Aufräumens hinter einem Abbruch eintreffen.
func tiZweiteSignale() []string { return []string{"INT", "HUP", "TERM"} }

// tiZweites sendet TERM im ersten Bau und sig während des ersten `rm` des Aufräumens, das danach zu Ende läuft: eine Meldung
// `abgebrochen`, Exit 2, je Container und Tag ein Entfernen, nichts bleibt.
func tiZweites(l Lauf, sig string) string {
	name := "zweites-" + sig
	env := []string{"TI_SIG=TERM", "TI_SIG_BEI=bau-image", "TI_SIG2=" + sig, "TI_SIG2_BEI=rm-pruefung"}
	r, g := tiErwartet(l, name, env, 2, "abgebrochen")
	if g != "" {
		return g
	}
	if _, err := os.Stat(filepath.Join(r.log, "signal2")); err != nil {
		return name + ": kein zweites Signal gesendet: " + err.Error()
	}
	if n := r.zahl("rm", "-f", "-v") + r.zahl("rmi"); n != 9 {
		return fmt.Sprintf("%s: %d Aufrufe rm und rmi statt 9", name, n)
	}
	return ""
}

// tiAbbruch ist Fall 235: INT, HUP und TERM an jeder Stelle enden mit 2 und `abgebrochen`, räumen auf und starten keinen weiteren
// Schritt; ausserhalb des Aufräumens erreicht TERM den laufenden Aufruf; ein zweites Signal während des Aufräumens ändert nichts
// (SPEC-323 Nr. 6 d).
func tiAbbruch() Fall {
	return tiFall("abbruch", func(l Lauf) string {
		orte := tiSignalOrte()
		if len(orte) != 9 {
			return fmt.Sprintf("%d Orte statt 9", len(orte))
		}
		for _, sig := range []string{"INT", "HUP", "TERM"} {
			for _, ort := range orte {
				if g := tiSignal(l, sig, ort); g != "" {
					return g
				}
			}
		}
		if len(tiZweiteSignale()) != 3 {
			return fmt.Sprintf("%d zweite Signale statt 3", len(tiZweiteSignale()))
		}
		for _, sig := range tiZweiteSignale() {
			if g := tiZweites(l, sig); g != "" {
				return g
			}
		}
		return ""
	})
}

// tiSignal ist das Paar aus Signal und Ort.
func tiSignal(l Lauf, sig, ort string) string {
	name := sig + "-" + ort
	env := []string{"TI_SIG=" + sig, "TI_SIG_BEI=" + ort}
	if ort == "sleep" {
		env = append(env, "TI_TOKEN_NACH=nie")
	}
	r, g := tiErwartet(l, name, env, 2, "abgebrochen")
	if g != "" || ort == "rm-pruefung" {
		return g
	}
	b, err := os.ReadFile(filepath.Join(r.log, "empfangen"))
	if err != nil || string(b) != "TERM\n" {
		return fmt.Sprintf("%s: der laufende Aufruf empfing %q, %v", name, b, err)
	}
	if g := tiNachSignal(r, ort); g != "" {
		return name + ": " + g
	}
	return ""
}

// tiNachSignal prüft, dass hinter dem unterbrochenen Aufruf nur das Aufräumen folgt.
func tiNachSignal(r tiLauf, ort string) string {
	erstes := slices.IndexFunc(r.aufrufe, func(a []string) bool { return a[0] == "rm" })
	if erstes < 1 {
		return "kein Aufräumen"
	}
	vor := r.aufrufe[erstes-1]
	soll := map[string]string{"bau-image": "build", "bau-zweit": "build", "create-server": "create", "cp-token": "cp", "sleep": "inspect",
		"run-aufnehmen": "run", "stop": "stop", "run-pruefen": "run"}[ort]
	if vor[0] != soll {
		return fmt.Sprintf("letzter Aufruf vor dem Aufräumen %q, erwartet %s", vor, soll)
	}
	for _, a := range r.aufrufe[erstes:] {
		if !slices.Contains([]string{"rm", "ps", "rmi", "images"}, a[0]) {
			return fmt.Sprintf("Aufruf hinter dem Signal: %q", a)
		}
	}
	return ""
}

// tiDocker ist Fall 236: die Variable DOCKER wählt das Programm.
func tiDocker() Fall {
	return tiFall("docker-variable", func(l Lauf) string {
		if err := os.Rename(filepath.Join(l.Dir, "bin", "docker"), filepath.Join(l.Dir, "bin", "anderes-docker")); err != nil {
			return err.Error()
		}
		r, g := tiFahre(l, "variable", []string{"DOCKER=anderes-docker"})
		if g == "" {
			g = r.erwarte("variable", 0)
		}
		return g
	})
}

// testImageFaelle sind die Fälle 226 bis 236.
func testImageFaelle() []Fall {
	return []Fall{tiGut(), tiZweiLaeufe(), tiAufruf(), tiImage(), tiPhasen(), tiStopp(), tiTokenFall(), tiExitCodes(), tiAufraeumen(),
		tiAbbruch(), tiDocker()}
}
