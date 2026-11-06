#!/bin/sh
# nachbau-innen.sh — das bisherige Vorgehen der Rollen, nachgebaut nach dem Plan (Befund 1 und 2): je Mutation die
# ganze Kopie des Baums (/tmp/baum), `go test -count=1` über das Paket, bei nebenläufig zusätzlich ein Bau mit und ohne
# Detektor und 50 Prozesse je Art. Läuft im Container, gestartet von nachbau.sh; Umgebung REP: Wiederholungen des Satzes.
cd /tmp/x/subst && go build -o /tmp/subst . || exit 1
PKG=./src/hexagon/services/record
A=src/hexagon/services/record/anhaengen.go
T=$(printf '\t')
NL='
'
echo "host: $(cut -d' ' -f1-3 /proc/loadavg) cpus=$(nproc)"
mut() { # name klasse test alt neu
	t0=$(date +%s.%N)
	rm -rf /tmp/m
	cp -r /tmp/baum /tmp/m
	/tmp/subst "/tmp/m/$A" "$4" "$5" || { echo "nachbau $1 KEINE-AENDERUNG"; return; }
	cd /tmp/m || return
	n=0
	go test -count=1 $PKG > /dev/null 2>&1
	rc=$?
	n=$((n + 1))
	erg=gruen
	[ $rc -ne 0 ] && erg=rot
	if [ "$2" = nebenlaeufig ]; then
		CGO_ENABLED=1 go test -race -c -o /tmp/a.race $PKG || echo "race-Bau scheitert"
		go test -c -o /tmp/a $PKG
		cd src/hexagon/services/record || return
		mit=0
		ohne=0
		for i in $(seq 1 50); do /tmp/a.race -test.run "^$3\$" -test.count=1 > /dev/null 2>&1 || mit=$((mit + 1)); n=$((n + 1)); done
		for i in $(seq 1 50); do /tmp/a -test.run "^$3\$" -test.count=1 > /dev/null 2>&1 || ohne=$((ohne + 1)); n=$((n + 1)); done
		erg="$erg mit=$mit/50 ohne=$ohne/50"
	fi
	cd /tmp || return
	echo "nachbau $1 $erg prozesse=$n dauer=$(awk -v a="$(date +%s.%N)" -v b="$t0" 'BEGIN{printf "%.1f", a-b}')"
}
T0=$(date +%s.%N)
for rep in $(seq 1 "${REP:-1}"); do
	mut m1 nebenlaeufig TestSequenzGleichzeitig "${T}sz.anhaengen.Lock()${NL}${T}defer sz.anhaengen.Unlock()${NL}" ""
	mut m2 andere TestSequenz "${T}sz.zahl++${NL}${T}return nil" "${T}return nil"
	mut m3 andere TestSequenz "e.SequenceNo = sz.zahl + 1" "e.SequenceNo = sz.zahl"
	mut m4 andere TestSequenz "sz.zahl, sz.zahlBekannt = len(liste), true" "sz.zahl, sz.zahlBekannt = len(liste)+1, true"
	mut m5 andere TestSequenz "if !sz.zahlBekannt {" "if !sz.zahlBekannt || false {"
done
echo "nachbau gesamt dauer=$(awk -v a="$(date +%s.%N)" -v b="$T0" 'BEGIN{printf "%.1f", a-b}') host-ende: $(cut -d' ' -f1-3 /proc/loadavg)"
