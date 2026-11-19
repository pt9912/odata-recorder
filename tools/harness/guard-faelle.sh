#!/usr/bin/env bash
# guard-faelle — Pruefplan des Command-Guards (SPEC-214 Nr. 10): 52 Faelle ueber die Hook-Eingabe.
# Ruft das Skript des Arbeitsbaums (.claude/hooks/pretooluse-command-guard.sh) je Element mit der
# Eingabe auf stdin und vergleicht: durch = leere stdout und Exit 0; S, P, L, G, H = JSON-Objekt
# decision block, Exit 0, dessen reason das Merkmal der Klasse traegt. Jede Abweichung steht als
# `FAIL <Fall> <Befehl>` auf stdout, die Summe auf stderr; Exit 0 ohne, 1 mit Abweichung.
# Kein Gate, kein make-Ziel (SPEC-214 Nr. 10 d). Das Skript schreibt nichts in den Arbeitsbaum;
# nur der Fall 49 legt ein Wegwerfverzeichnis ausserhalb der Wurzel an und entfernt es.
# Notation der Elemente: <W> Wurzel, <SP> Scratchpad-Pfad, \n \r \t Steuerzeichen, \L Backslash
# vor Zeilenvorschub. Die Escape-Folge des Falls 40 und die Fuellung der Faelle 27, 28 und 47
# entstehen zur Laufzeit aus Bytewerten.
set -euo pipefail
export LC_ALL=C

if [ "$#" -ne 0 ]; then echo "Aufruf: bash tools/harness/guard-faelle.sh (ohne Argument)" >&2; exit 2; fi

self=${BASH_SOURCE[0]}
W="$(cd "$(dirname "$self")/../.." && pwd -P)"
GUARD="$W/.claude/hooks/pretooluse-command-guard.sh"
SP=/tmp/claude-0/p/s/scratchpad
HOME_STD=/home/guard-test
bs='\'
bsnl=$'\\\n'

TMP="$(mktemp -d)"
trap 'rm -rf -- "$TMP"' EXIT

total=0
fails=0
OUT=""
RC=0
KL=""

expand() {  # Notation -> Befehlstext in EXP
  local s=$1
  s=${s//'<W>'/$W}
  s=${s//'<SP>'/$SP}
  s=${s//'\L'/$bsnl}
  s=${s//'\r'/$'\r'}
  s=${s//'\n'/$'\n'}
  s=${s//'\t'/$'\t'}
  EXP=$s
}

json_command() {  # json_command <befehlstext> -> JSON der Hook-Eingabe in JSON
  local s=$1
  s=${s//"$bs"/"$bs$bs"}
  s=${s//'"'/"$bs\""}
  s=${s//$'\n'/"${bs}n"}
  s=${s//$'\r'/"${bs}r"}
  s=${s//$'\t'/"${bs}t"}
  JSON=$(printf '{"tool_input":{"command":"%s"}}' "$s")
}

run_guard() {  # run_guard <eingabe> <home> [guard] -> OUT, RC, ERR
  local g=${3:-$GUARD}
  OUT=$(printf '%s' "$1" | HOME="$2" bash "$g" 2>"$TMP/err") && RC=0 || RC=$?
  ERR=$(cat "$TMP/err")
}

klasse() {  # OUT -> KL: durch, S, P, L, G, H oder ungueltig (kein oder mehr als ein Merkmal)
  local n=0 k=""
  if [ -z "$OUT" ]; then KL=durch; return 0; fi
  if [[ $OUT == *'Edit and Write'* ]]; then k=S; n=$((n+1)); fi
  if [[ $OUT == *'Pattern-based'* ]]; then k=P; n=$((n+1)); fi
  if [[ $OUT == *'scratchpad directory'* ]]; then k=L; n=$((n+1)); fi
  if [[ $OUT == *'32768'* ]]; then k=G; n=$((n+1)); fi
  if [[ $OUT == *'This repository'* ]]; then k=H; n=$((n+1)); fi
  if [ "$n" -eq 1 ]; then KL=$k; else KL=ungueltig; fi
}

verdict() {  # verdict <fall> <erwartung> <label> <eingabe-json> <home>
  local fall=$1 erw=$2 label=$3 ok=1
  total=$((total+1))
  run_guard "$4" "$5"
  klasse
  if [ "$RC" -ne 0 ]; then ok=0; fi
  if [ "$KL" != "$erw" ]; then ok=0; fi
  if [ "$ok" -eq 0 ]; then
    fails=$((fails+1))
    printf 'FAIL %s %s\n' "$fall" "$label"
  fi
}

fall() {  # fall <nr> <erwartung> <zahl> mit den Elementen auf stdin, eines je Zeile
  local nr=$1 erw=$2 count=$3 line n=0
  while IFS= read -r line; do
    n=$((n+1))
    expand "$line"
    json_command "$EXP"
    verdict "$nr" "$erw" "$line" "$JSON" "$HOME_STD"
  done
  if [ "$n" -ne "$count" ]; then
    total=$((total+1)); fails=$((fails+1))
    printf 'FAIL %s Zahl der Elemente %s statt %s\n' "$nr" "$n" "$count"
  fi
}

fall_home() {  # fall_home <nr> <erwartung> <home> <befehl>
  expand "$4"
  json_command "$EXP"
  verdict "$1" "$2" "$4 (HOME=$3)" "$JSON" "$3"
}

fall_roh() {  # fall_roh <nr> <erwartung> <label> <eingabe>
  verdict "$1" "$2" "$3" "$4" "$HOME_STD"
}

# 1 S, Operatoren
fall 1 S 11 <<'ENDE'
echo x > docs/a.md
echo x >> docs/a.md
echo x >docs/a.md
echo x>docs/a.md
echo x 2> docs/a.md
echo x 2>>docs/a.md
echo x &> docs/a.md
echo x &>> docs/a.md
echo x >| docs/a.md
echo x >& docs/a.md
echo x 10> docs/a.md
ENDE

# 2 S, zusammengesetzte Befehle, mehrere Umleitungen
fall 2 S 7 <<'ENDE'
{ echo a; echo b; } > docs/a.md
for f in a b; do echo $f; done > docs/a.md
(echo a; echo b) > docs/a.md
exec > docs/a.md
echo x | cat > docs/a.md
cat docs/a.md > docs/b.md
echo x > <SP>/a > docs/b.md
ENDE

# 3 durch, Dup und aussen
fall 3 durch 11 <<'ENDE'
echo x > /dev/null
echo x 2>/dev/null
echo x > /dev/stderr
make gates > /tmp/g.log 2>&1
echo x 2>&1
echo x >&2
echo x 2>&-
echo x &>/dev/null
echo x > <SP>/n.txt
echo x >> <SP>/n.txt
echo x > /dev/null 2>&1; echo y 1>&2
ENDE

# 4 S, Wurzel und Normalisierung
fall 4 S 5 <<'ENDE'
echo x > <W>
echo x > <W>/
echo x > <W>/docs/a.md
echo x > /tmp/..<W>/f
echo x > <W>/./docs//a.md
ENDE

# 5 durch, Komponentengrenze und .. aus der Wurzel
fall 5 durch 3 <<'ENDE'
echo x > <W>-nachbar/f
echo x > <W>x
echo x > <W>/../f
ENDE

# 6 S, relative Ziele
fall 6 S 3 <<'ENDE'
echo x > ./docs/a.md
echo x > ../a.md
echo x > a.md
ENDE

# 7 Tilde
fall 7 durch 3 <<'ENDE'
echo x > ~/a.md
echo x > ~guard/a.md
echo x > ~+/a.md
ENDE
fall_home 7 durch "" 'echo x > ~/a.md'
fall_home 7 S "$W" 'echo x > ~/a.md'

# 8 S, Wurzel-Variablen
fall 8 S 4 <<'ENDE'
echo x > $PWD/a.md
echo x > "${PWD}/a.md"
echo x > $CLAUDE_PROJECT_DIR/a.md
echo x > "${CLAUDE_PROJECT_DIR}"/a.md
ENDE

# 9 durch, Zweifel am Ziel
fall 9 durch 8 <<'ENDE'
echo x > $OUT
echo x > "$OUT/a.md"
echo x > $PWDX/a.md
echo x > "$(pwd)/a.md"
echo x > a*.md
echo x >
echo x > >(cat)
echo x > ""
ENDE

# 10 durch, Verzeichnis-Zustand aussen
fall 10 durch 8 <<'ENDE'
cd <SP> && echo x > a.md
cd $D && echo x > a.md
cd <SP> && cd sub && echo x > a.md
cd ~ && echo x > a.md
pushd <SP> && echo x > a.md
(cd <SP> && echo x > a.md)
cd - && echo x > a.md
popd && echo x > a.md
ENDE

# 11 S, Verzeichnis-Zustand innen
fall 11 S 5 <<'ENDE'
cd docs && echo x > a.md
cd <W>/docs && echo x > a.md
cd <SP>; cd <W>; echo x > a.md
cd /tmp && echo x > <W>/a.md
cd <SP> && cd <W> && echo x > a.md
ENDE
fall_home 11 S "$W" 'cd ~ && echo x > a.md'

# 12 durch, Zitat und Backslash
fall 12 durch 7 <<'ENDE'
echo "a > b"
echo 'a > docs/a.md'
git commit -m "fix: a > docs/a.md"
awk '$1 > 5' f
echo \> docs/a.md
echo "unclosed > docs/a.md
echo "a\" > docs/a.md"
ENDE

# 13 S, Zitat vor der Umleitung
fall 13 S 4 <<'ENDE'
echo "a" > docs/a.md
echo 'a' >docs/a.md
echo "a" "b" > docs/a.md
echo '"' > docs/a.md
ENDE

# 14 S, Heredoc mit Schreibweg
fall 14 S 3 <<'ENDE'
cat > docs/a.md <<'EOF'\ntext\nEOF
cat <<'EOF' > docs/a.md\ntext\nEOF
cat <<< "x" > docs/a.md
ENDE

# 15 durch, Commit-Wege und Scratchpad-Heredoc
fall 15 durch 7 <<'ENDE'
git commit -F - <<'EOF'\nmsg\nEOF
cat <<'EOF' | git commit -F -\nmsg\nEOF
printf msg | git commit -F -
git commit -F - <<< "msg"
git commit -F <SP>/m.txt
git commit -F - <<-'EOF'\n\tmsg\n\tEOF
cat > <SP>/a.txt <<'EOF'\ntext\nEOF
ENDE

# 16 durch, Rumpf nicht geprueft
fall 16 durch 6 <<'ENDE'
git commit -F - <<'EOF'\nsed -i x y\n> z\npip install q\ntee docs/a\nEOF
cat <<"EOF"\n> z\nEOF
cat <<\EOF\n> z\nEOF
cat << EOF\n> z\nEOF
cat <<A <<B\nsed -i x\nA\n> z\nB
cat <<EOF\n  EOF\nsed -i x y\nEOF
ENDE

# 17 S, Rest der Zeile, Zeilen danach, nicht terminiert, mehrere Operatoren, <<-, Zitat
fall 17 S 6 <<'ENDE'
cat <<EOF; sed -i x y\nbody\nEOF
cat <<EOF\nbody\nEOF\nsed -i x y
cat <<EOF\nbody\nsed -i x y
cat > docs/a <<A <<B\n1\nA\n2\nB
cat <<-EOF\n\tEOF\nsed -i x y
echo "<<EOF"; sed -i x y
ENDE

# 18 S, Rumpf an eine Shell
fall 18 S 2 <<'ENDE'
bash <<'EOF'\nsed -i x y\nEOF
cat <<EOF | bash\nsed -i x y\nEOF
ENDE

# 19 H, Rumpf an eine Shell
fall 19 H 1 <<'ENDE'
bash <<'EOF'\npip install q\nEOF
ENDE

# 20 S, tee
fall 20 S 7 <<'ENDE'
echo x | tee docs/a.md
echo x | tee -a docs/a.md
echo x | tee --append docs/a.md
echo x | tee <SP>/a.txt docs/a.md
echo x | tee -- docs/a.md
echo x | sudo tee docs/a.md
echo x | /usr/bin/tee docs/a.md
ENDE

# 21 durch, tee
fall 21 durch 4 <<'ENDE'
echo x | tee <SP>/a.txt
echo x | tee
echo x | tee -a "$F"
echo tee docs/a.md
ENDE

# 22 S, sed -i
fall 22 S 17 <<'ENDE'
sed -i s/a/b/ docs/a.md
sed -i.bak s/a/b/ docs/a.md
sed --in-place s/a/b/ docs/a.md
sed --in-place=.bak s/a/b/ docs/a.md
sed --in s/a/b/ docs/a.md
sed -ni s/a/b/p docs/a.md
sed -Ei s/a/b/ docs/a.md
sed -n -i s/a/b/p docs/a.md
sed s/a/b/ docs/a.md -i
sed -i s/a/b/ <SP>/f.txt
xargs sed -i s/a/b/
env X=1 sed -i s/a/b/ f
/usr/bin/sed -i s/a/b/ f
"sed" -i s/a/b/ f
bash -c "sed -i s/a/b/ f"
sh -lc 'sed -i s/a/b/ f'
sed \L-i s/a/b/ f
ENDE

# 23 durch, sed ohne -i und Namensvettern
fall 23 durch 11 <<'ENDE'
sed -n s/a/b/p docs/a.md
sed -E s/a/b/ docs/a.md
sed -e s/-i/x/ docs/a.md
sed -s -n p docs/a.md
sed --expression=p --quiet docs/a.md
echo sed -i x
git commit -m "fix sed -i handling"
docker run --rm img sed -i s/a/b/ f
grep -i x docs/a.md
docker exec c sed -i s/a/b/ f
sed -n -e p docs/a.md
ENDE

# 24 S, git apply
fall 24 S 9 <<'ENDE'
git apply p.patch
git apply - <<'EOF'\ndiff\nEOF
git -C docs apply p.patch
git -c core.x=1 apply p.patch
git --no-pager apply p.patch
git apply --index p.patch
git apply --apply --stat p.patch
git diff | git apply -R
git -C <W> apply p.patch
ENDE

# 25 durch, git apply nur lesend, aussen, Namensvettern
fall 25 durch 11 <<'ENDE'
git apply --check p.patch
git apply --stat p.patch
git apply --numstat p.patch
git apply --summary p.patch
git -C <SP>/r apply p.patch
git -C "$R" apply p.patch
git status
git log --grep apply
git commit -m "apply"
git am p.mbox
git apply --check --index p.patch
ENDE

# 26 durch, leer
total_26_vorher=$total
json_command ""; fall_roh 26 durch "der leere Befehl" "$JSON"
expand '   \t  '; json_command "$EXP"; fall_roh 26 durch "Leerraum" "$JSON"
expand '\n\n'; json_command "$EXP"; fall_roh 26 durch "Zeilenenden" "$JSON"
fall_roh 26 durch 'Eingabe {"tool_input":{}}' '{"tool_input":{}}'
fall_roh 26 durch 'Eingabe {"tool_input":{"command":""}}' '{"tool_input":{"command":""}}'
if [ $((total-total_26_vorher)) -ne 5 ]; then total=$((total+1)); fails=$((fails+1)); echo "FAIL 26 Zahl der Elemente"; fi

# 27 durch, Laenge an der Grenze; 28 G, Laenge darueber
printf -v pad '%*s' 32763 ''
cmd27="echo ${pad// /a}"
json_command "$cmd27"; fall_roh 27 durch "echo mit 32768 Zeichen (${#cmd27})" "$JSON"
if [ "${#cmd27}" -ne 32768 ]; then total=$((total+1)); fails=$((fails+1)); echo "FAIL 27 Laenge ${#cmd27}"; fi
cmd28="${cmd27}a"
json_command "$cmd28"; fall_roh 28 G "echo mit 32769 Zeichen (${#cmd28})" "$JSON"

# 29 S, Segmente, Zeilenenden, Fortsetzung
fall 29 S 8 <<'ENDE'
echo a; echo b > docs/a.md
echo a && sed -i x f
true || echo x > docs/a.md
echo a\necho b > docs/a.md
echo a\r\necho b > docs/a.md
echo x \L> docs/a.md
echo a & echo b > docs/a.md
echo a |& cat > docs/a.md
ENDE

# 30 durch, Segmente
fall 30 durch 2 <<'ENDE'
echo a; echo b > <SP>/n; echo c
echo a && echo b | cat
ENDE

# 31 S, Verschachtelung
fall 31 S 4 <<'ENDE'
bash -c "echo x > docs/a.md"
sh -lc 'echo x > docs/a.md'
bash -c bash -c bash -c sed -i x f
bash -c "bash -c 'echo x > docs/a.md'"
ENDE

# 32 durch, Verschachtelung
fall 32 durch 2 <<'ENDE'
bash -c 'echo x > <SP>/a'
bash -c bash -c bash -c echo x
ENDE

# 33 H, Tiefe 4
fall 33 H 1 <<'ENDE'
bash -c bash -c bash -c bash -c echo x
ENDE

# 34 H, Rang: der erste Verstoss ist die Sperrliste
fall 34 H 3 <<'ENDE'
pip install x; echo y > docs/a.md
pip install x > docs/a.md
pip install x; pkill y
ENDE

# 35 S, Rang: der erste Verstoss ist S
fall 35 S 6 <<'ENDE'
echo y > docs/a.md; pip install x
sed -i x f > docs/a.md
tee docs/a > docs/b
bash -c "pip install x" > docs/a.md
bash -c "echo x > docs/a.md; pip install y"
echo y > docs/a.md; pkill y
ENDE

# 36 P und L, Rang im Segment
fall 36 P 1 <<'ENDE'
pkill x > docs/a.md
ENDE
fall 36 L 1 <<'ENDE'
rm -rf docs > docs/a.md
ENDE

# 37 Meldungstext je Klasse: eigenes Merkmal, keines der vier anderen
merkmal() {  # merkmal <klasse> -> M
  case $1 in
    S) M='Edit and Write';; P) M='Pattern-based';; L) M='scratchpad directory';;
    G) M='32768';; H) M='This repository';;
  esac
}
meldung_von() {  # meldung_von <klasse> -> OUT, RC, ERR
  local c
  case $1 in
    S) c='echo x > docs/a.md';; P) c='pkill x';; L) c='rm -rf docs';;
    G) c=$cmd28;; H) c='pip install x';;
  esac
  json_command "$c"
  run_guard "$JSON" "$HOME_STD"
}
for k in S P L G H; do
  meldung_von "$k"
  merkmal "$k"; eigen=$M
  total=$((total+1)); ok=1
  [[ $OUT == *"$eigen"* ]] || ok=0
  for k2 in S P L G H; do
    [ "$k2" = "$k" ] && continue
    merkmal "$k2"
    if [[ $OUT == *"$M"* ]]; then ok=0; fi
  done
  if [ "$ok" -eq 0 ]; then fails=$((fails+1)); echo "FAIL 37 Meldung $k"; fi
done

# 38 Meldung ist gueltig: genau decision (block) und reason, ASCII ohne " und Backslash, Exit 0, stderr leer
re='^\{[[:space:]]*"decision":[[:space:]]*"block",[[:space:]]*"reason":[[:space:]]*"[^"\\]*"[[:space:]]*\}$'
for k in S P L G H; do
  meldung_von "$k"
  total=$((total+1)); ok=1
  [[ $OUT =~ $re ]] || ok=0
  [[ $OUT == *[^[:print:][:space:]]* ]] && ok=0
  [ "$RC" -eq 0 ] || ok=0
  [ -z "$ERR" ] || ok=0
  if [ "$ok" -eq 0 ]; then fails=$((fails+1)); echo "FAIL 38 Meldung $k"; fi
done

# 39 H, Bestand
fall 39 H 6 <<'ENDE'
pip install x
/usr/bin/pip install x
sudo pip install x
python3 x.py
apt-get install x
git commit -m "x & pip y"
ENDE

# 40 H, Zweifel des Extraktors; die Escape-Folge aus Bytewerten
fall_roh 40 H 'Eingabe abgeschnitten' '{"tool_input":{"command":"echo'
fall_roh 40 H 'Eingabe mit Escape-Folge aus Backslash, u und vier Hex-Ziffern' "$(printf '{"tool_input":{"command":"echo %su0041"}}' "$bs")"
fall_roh 40 H 'Eingabe garbage' 'garbage'

# 41 P, pkill und killall
fall 41 P 5 <<'ENDE'
pkill -f foo
pkill foo
killall foo
sudo pkill foo
/usr/bin/pkill foo
ENDE

# 42 P, pgrep -f
fall 42 P 6 <<'ENDE'
pgrep -f foo
pgrep -fl foo
pgrep -af foo
pgrep --full foo
kill $(pgrep -f foo)
pgrep -f foo | xargs kill
ENDE

# 43 durch, Prozess
fall 43 durch 9 <<'ENDE'
pgrep foo
pgrep -u 1000 foo
pgrep -l foo
pgrep -F pidfile
kill 1234
kill -9 $pid
kill %1
echo pkill
git commit -m "pkill"
ENDE

# 44 L, rekursives rm
fall 44 L 22 <<'ENDE'
rm -rf docs/x
rm -r docs
rm -R docs
rm --recursive docs
rm -fr docs
rm -vrf docs
rm -rf <W>/x
rm -rf <W>/scratchpad/x
rm -rf .
rm -rf ..
rm -rf x
rm -rf -- docs/x
rm -rf /
rm -rf ~
rm -rf ~/x
rm -rf /tmp/x
rm -rf /tmp
rm -rf <SP>
rm -rf <SP>/*
rm -rf <SP>/mut-*
rm -rf <SP>/../x
rm -rf <SP>/a docs/x
ENDE

# 45 durch, rm
fall 45 durch 11 <<'ENDE'
rm -rf <SP>/mut-guard
rm -rf <SP>/a/b
rm -rf <SP>/a <SP>/b
rm -rf "$d"
rm -rf ~guard/x
rm -rf
cd <SP> && rm -rf mut-x
rm -f docs/x
rm docs/x
git rm -r docs/x
docker rm -f c
ENDE

# 46 L, Muster vor Zweifel
fall 46 L 1 <<'ENDE'
rm -rf $d/*
ENDE

# 47 Latenz: Median von 20 Laeufen git status; 32768 Zeichen mit vielen Segmenten
now_us() { local t=$EPOCHREALTIME; NOW=${t//./}; }
json_command "git status"
lat=()
for _ in $(seq 1 20); do
  now_us; t0=$NOW
  run_guard "$JSON" "$HOME_STD"
  now_us; lat+=($((NOW - t0)))
done
sorted=($(printf '%s\n' "${lat[@]}" | sort -n))
median=$(( (sorted[9] + sorted[10]) / 2 ))
total=$((total+1))
if [ "$median" -gt 100000 ]; then fails=$((fails+1)); echo "FAIL 47 Median von 20 Laeufen git status: $((median/1000)) ms"; fi
expand 'echo a > <SP>/x;'
unit=$EXP
lang=""
for ((k=0; k<32768/${#unit}; k++)); do lang+=$unit; done
printf -v fuell '%*s' $((32768 - ${#lang})) ''
lang+=$fuell
json_command "$lang"
now_us; t0=$NOW
run_guard "$JSON" "$HOME_STD"
now_us; lang_us=$((NOW - t0))
klasse
total=$((total+1))
if [ "${#lang}" -ne 32768 ] || [ "$KL" != durch ] || [ "$lang_us" -gt 3000000 ]; then
  fails=$((fails+1)); echo "FAIL 47 Befehl aus ${#lang} Zeichen: Klasse $KL, $((lang_us/1000)) ms"
fi
LATENZ="Median git status $((median/1000)) ms, 32768 Zeichen $((lang_us/1000)) ms"

# 48 Wiederholung: je ein Befehl der Klassen S, P, L, G, H zweimal, dieselbe Ausgabe
for k in S P L G H; do
  meldung_von "$k"; o1=$OUT
  meldung_von "$k"; o2=$OUT
  total=$((total+1))
  if [ -z "$o1" ] || [ "$o1" != "$o2" ]; then fails=$((fails+1)); echo "FAIL 48 Wiederholung $k"; fi
done

# 49 S, Wurzel ueber einen Symlink
ln -s "$W" "$TMP/link"
LGUARD="$TMP/link/.claude/hooks/pretooluse-command-guard.sh"
for c in "echo x > $TMP/link/docs/a.md" "echo x > $W/docs/a.md"; do
  json_command "$c"
  total=$((total+1))
  run_guard "$JSON" "$HOME_STD" "$LGUARD"
  klasse
  if [ "$RC" -ne 0 ] || [ "$KL" != S ]; then fails=$((fails+1)); echo "FAIL 49 $c"; fi
done

# 50 durch, die Negative von Nr. 11
fall 50 durch 14 <<'ENDE'
cp <SP>/a docs/a.md
mv <SP>/a docs/a.md
dd if=<SP>/a of=docs/a.md
install <SP>/a docs/a.md
perl -i -pe 1 docs/a.md
patch -p1 < p.patch
gofmt -w docs/a.go
bash <<< "sed -i x f"
find . -name x -exec sed -i s/a/b/ {} +
awk -i inplace 1 docs/a.md
echo x <> docs/a.md
echo x > <W>-nachbar/docs/a.md
ps aux
rm docs/a.md
ENDE

# 51 S, die bekannten Falschmeldungen
fall 51 S 2 <<'ENDE'
git commit -m "a; b > docs/a.md"
echo $((1 > 2))
ENDE

# 52 Aufruf mit einem Argument, auch einem leeren: Exit 2, Aufruf auf stderr
for arg in "" x; do
  total=$((total+1))
  rc52=0
  err52=$(bash "$self" "$arg" 2>&1 >/dev/null) || rc52=$?
  if [ "$rc52" -ne 2 ] || [[ $err52 != *Aufruf* ]]; then fails=$((fails+1)); echo "FAIL 52 Argument '$arg'"; fi
done

# Zusatz zu den 52 Faellen: Klassen, die der Vertrag als Menge zusagt, als Schleifen ueber alle
# Elemente (Generator je Klasse, das Orakel steht in den Tabellen der Schleife, nicht im Guard),
# und Einzelfaelle, die die 52 Faelle mit ihren Elementen nicht trennen.
zusatz() {  # zusatz <label> <erwartung> <befehl>
  expand "$3"
  json_command "$EXP"
  verdict "$1" "$2" "$3" "$JSON" "$HOME_STD"
}
zahl() {  # zahl <label> <ist> <soll>
  if [ "$2" -ne "$3" ]; then total=$((total+1)); fails=$((fails+1)); echo "FAIL $1 Zahl der Elemente $2 statt $3"; fi
}

# Z1 bis Z5: Einzelfaelle
zusatz Z1 durch 'cat <<-EOF\n\tsed -i x y\n\tEOF'
zusatz Z2 S 'echo x | tee -- -a'
zusatz Z3 L 'rm -rf -- -x'
zusatz Z4 S 'echo a\rsed -i x y'
mkdir "$TMP/bin"
ln -s "$(command -v cat)" "$TMP/bin/cat"
total=$((total+1))
run_guard_ohne_awk() { OUT=$(printf '%s' '{"tool_input":{"command":"echo x"}}' | PATH="$TMP/bin" "$BASH" "$GUARD" 2>/dev/null) && RC=0 || RC=$?; }
run_guard_ohne_awk; klasse
if [ "$RC" -ne 0 ] || [ "$KL" != H ]; then fails=$((fails+1)); echo "FAIL Z5 ohne awk"; fi

# Z6 Operatoren x Abstand x Ziel (10 x 2 x 8 = 160): Ziele unter der Wurzel sperren, die uebrigen gehen durch
n=0
for o in '>' '>>' '>|' '&>' '&>>' '>&' '2>' '2>>' '10>' '1>|'; do
  for s in ' ' ''; do
    for t in 'docs/a.md' '<W>/docs/a.md' './a.md' '$PWD/a.md'; do zusatz Z6 S "echo x $o$s$t"; n=$((n+1)); done
    for t in '/dev/null' '<SP>/n.txt' '<W>-nachbar/f' '$OUT'; do zusatz Z6 durch "echo x $o$s$t"; n=$((n+1)); done
  done
done
zahl Z6 "$n" 160

# Z7 Trenner x Kopf (10 x 6 = 60): hinter jedem Trenner gilt der Kopf
n=0
for sep in ';' '&&' '||' '|' '&' '\n' '$(' '`' '(' '|&'; do
  for h in 'sed -i x f|S' 'tee docs/a|S' 'git apply p|S' 'pkill x|P' 'rm -rf docs|L' 'pip install x|H'; do
    zusatz Z7 "${h##*|}" "echo a $sep ${h%|*}"; n=$((n+1))
  done
done
zahl Z7 "$n" 60

# Z8 Praefix x Kopf (10 x 6 = 60)
n=0
for p in sudo env command exec nice time xargs eval 'X=1' '{'; do
  for h in 'sed -i x f|S' 'tee docs/a|S' 'git apply p|S' 'pkill x|P' 'rm -rf docs|L' 'pip install x|H'; do
    zusatz Z8 "${h##*|}" "$p ${h%|*}"; n=$((n+1))
  done
done
zahl Z8 "$n" 60

# Z9 sed: Flag x Stellung (10 x 2 = 20 sperren) und Flags ohne -i (8 gehen durch)
n=0
for f in -i -i.bak -ni -in -Ei -nEi --in-place --in-place=.b --in --i; do
  zusatz Z9 S "sed $f s/a/b/ docs/a.md"
  zusatz Z9 S "sed s/a/b/ docs/a.md $f"
  n=$((n+2))
done
for f in -n -E -s -nE --quiet --expression=p --null-data --regexp-extended; do zusatz Z9 durch "sed $f s/a/b/ docs/a.md"; n=$((n+1)); done
zahl Z9 "$n" 28

# Z10 rm: rekursive Flags x Operand (10 x 5 = 50) und Flags ohne r (8)
n=0
for f in -r -R -rf -fr -Rf -vrf -rv --recursive --rec --r; do
  zusatz Z10 L "rm $f docs/x"
  zusatz Z10 durch "rm $f <SP>/a/b"
  zusatz Z10 L "rm $f <SP>/*"
  zusatz Z10 L "rm $f <SP>"
  zusatz Z10 L "rm $f <SP>/../x"
  n=$((n+5))
done
for f in -f -v -fv --force --verbose -i --interactive --no-preserve-root; do zusatz Z10 durch "rm $f docs/x"; n=$((n+1)); done
zahl Z10 "$n" 58

# Z11 pgrep: Flags mit f (7 sperren) und ohne f (8 gehen durch)
n=0
for f in -f -fl -lf -af -fa --full --f; do zusatz Z11 P "pgrep $f foo"; n=$((n+1)); done
for f in -l -a -x -c '-F pidfile' '-u 1000' --list-full --count; do zusatz Z11 durch "pgrep $f foo"; n=$((n+1)); done
zahl Z11 "$n" 15

# Z12 git apply: Optionen vor dem Unterbefehl (9 sperren), nur lesend mit und ohne --apply (8), -C (5)
n=0
for o in '-C docs' '-c a=b' --no-pager --bare '--git-dir x' '--work-tree x' '--namespace x' '--config-env a=B' -p; do zusatz Z12 S "git $o apply p.patch"; n=$((n+1)); done
for r in --check --stat --numstat --summary; do
  zusatz Z12 durch "git apply $r p.patch"
  zusatz Z12 S "git apply --apply $r p.patch"
  n=$((n+2))
done
for c in '<SP>/r|durch' '$R|durch' '<W>/r|S' 'docs|S' '.|S'; do zusatz Z12 "${c##*|}" "git -C ${c%|*} apply p.patch"; n=$((n+1)); done
zahl Z12 "$n" 22

echo "guard-faelle: $total Pruefungen, $fails Abweichungen in ${SECONDS} s ($LATENZ)" >&2
if [ "$fails" -ne 0 ]; then exit 1; fi
exit 0
