#!/usr/bin/env bash
# pretooluse-command-guard — Stolperdraht fuer den Befehl, den eine Rolle dem Bash-Werkzeug
# uebergibt (SPEC-214). Reines bash + awk, KEIN node/jq/OCI. Drei Gruppen werden gesperrt:
#   H  Host-Paketmanager und Host-Toolchains (AGENTS.md Hard Rule 3.1), Extraktor-Zweifel, Tiefe
#   S  Schreibwege auf Repo-Dateien: Umleitung (> >> &> >| >&) mit Ziel unter der Wurzel,
#      tee, sed -i, git apply
#   P  pkill, killall, pgrep -f        L  rm mit rekursivem Flag ausser unter <scratchpad>/<Name>
#   G  Befehl laenger als 32768 Zeichen
# Alles andere geht ohne Ausgabe durch (kein "approve": die normale Permission-Pruefung laeuft).
# Im Zweifel am Ziel (Variable, Ersetzung, Muster, ~Name, kein Ziel) geht der Befehl durch.
#
# Der awk-Extraktor (tools/harness/extract-command.awk) zieht das Feld tool_input.command aus der
# Hook-JSON; bei Parse-Zweifel, fehlendem awk, Tiefe > 3 und Laenge > 32768 blockt der Guard.
#
# Zerlegung: quote-BLIND an ; & && || | $( ` ( und Zeilenenden (ein Trenner in einem Argument
# startet ein neues Segment: `git commit -m "a; b > f"` meldet S; Abhilfe `git commit -F <datei>`;
# die Gegenseite `echo "a; b" > f` meldet nichts: das Zitat bleibt im zweiten Segment offen und
# verbirgt die Umleitung, SPEC-214 Nr. 11 f, Fall 50).
# Ausnahme: &> >& >| gehoeren zur Umleitung. Je Segment gilt Zitat- und Backslash-Zustand fuer
# Umleitungen; ein Heredoc-Rumpf wird nur geprueft, wenn eine Shell ohne -c ihn lesen kann oder
# die Endezeile fehlt. Sub-Shell-Strings (`bash -c "..."`, auch -lc/-ec/-cx) rekursiv bis Tiefe 3.
# tee und rm lesen ein Umleitungswort und das Wort hinter einem reinen Operator nicht als Operand
# oder Flag (SPEC-214 Nr. 4 f).
# Rang: erster Verstoss in der Reihenfolge der Segmente; im Segment Sperrliste, Kopfpruefung,
# Umleitung, Rekursion.
#
# Der Guard ist KEINE Sandbox: er sieht den Befehlstext, nicht die Wirkung. Ungeprueft bleiben
# cp, mv, dd, install, perl -i, patch, awk -i inplace, find -exec, <>, bash <<< "..." u. a.
# (SPEC-214 Nr. 11). Prueflauf: `bash tools/harness/guard-faelle.sh` — kein Gate, kein make-Ziel;
# wer Guard, Extraktor oder tools/harness/blocked/ aendert, faehrt ihn vor der Uebergabe.
set -uo pipefail

here_src=${BASH_SOURCE[0]}
case $here_src in */*) here=${here_src%/*};; *) here=.;; esac
extractor="$here/../../tools/harness/extract-command.awk"
blocked_dir="$here/../../tools/harness/blocked"

# Wurzel: Elternverzeichnis von .claude/ in der Schreibweise des Aufrufs und als pwd -P.
roots="$(cd "$here/../.." 2>/dev/null && pwd && pwd -P)" || roots=""
ROOT1=${roots%%$'\n'*}
ROOT2=${roots##*$'\n'}
[ -n "$ROOT1" ] || ROOT1=/nonexistent-guard-root
[ -n "$ROOT2" ] || ROOT2=$ROOT1

emit() {  # emit <klasse>: Sperre als JSON auf stdout
  local reason
  case $1 in
    S) reason='Writing to repo files by redirect, tee, sed -i or git apply is blocked in this repository. Use the Edit and Write tools for repo files; send output to the scratchpad (outside the repo root); give commit messages with git commit -F <file> or git commit -F - and a heredoc.';;
    P) reason='Pattern-based process control (pkill, killall, pgrep -f) can hit processes of other runs and is blocked. Use kill with the PID of your own process, or docker rm -f with your own container name.';;
    L) reason='Recursive rm is only allowed on a directory below a scratchpad directory (a path with a scratchpad component followed by a name), never on repo paths, patterns or shallow paths. Remove your own directory by exact name; use git rm for repo files.';;
    G) reason='The command is longer than 32768 characters and is not checked. Write the content to a file in the scratchpad with the Write tool and run that file.';;
    *) reason='This repository is make/Docker-only (AGENTS.md Hard Rule 3.1). Use make targets; do not install or run host package managers or host toolchains (apt/brew/pip/npm/cargo/go/...). On parse doubt the guard fails closed.';;
  esac
  printf '{\n  "decision": "block",\n  "reason": "%s"\n}\n' "$reason"
}

# BLOCKED = gebackener universeller Boden (fail-safe, NIE fail-open) + optionale
# Sprach-Fragmente aus tools/harness/blocked/* (Wortlisten, whitespace-getrennt); ein
# geleertes oder fehlendes blocked/ laesst den Boden unberuehrt.
BLOCKED="apt apt-get brew pip pip3 pipx npm pnpm yarn npx corepack cargo rustup gem conda"
if [ -d "$blocked_dir" ]; then
  for bf in "$blocked_dir"/*; do
    [ -f "$bf" ] && BLOCKED="$BLOCKED $(cat "$bf")"
  done
fi
PREFIXES="sudo env command exec nice time xargs eval"
SHELLS="bash sh zsh dash ksh"
# Umleitungswort und reiner Operator (SPEC-214 Nr. 4 f); als Variablen, weil < > & | im Ausdruck
# nicht im [[ =~ ]] stehen duerfen.
RE_RWORT='^([0-9]*[<>]|&>)'
RE_ROP='^([0-9]*(>|>>|>[|]|>&|<|<<|<<-|<<<|<&)|&>|&>>)$'

set -f   # Wortzerlegung ohne Pfadnamen-Expansion (nach der Glob-Schleife oben)

VERDICT=""        # Klasse der Sperre: H S P L G
DIR=in            # Verzeichnis-Zustand fuer relative Ziele: in | out (SPEC-214 Nr. 3 c)
CLS=""            # Ergebnis von classify: repo | out | doubt
NPATH=""          # normalisierter Pfad der Reihe 6 (nur bei CLS=repo oder out)
NORM=""
STRIPPED=""
QI=0
WORD=""
WEND=0
SEGTXT=""
CMDTEXT=""
FULLTEXT=""
HDBODY=0
UNTERM=0
PH=0
RW=0              # Ergebnis von redir_word: 0 Operand oder Flag, 1 Umleitungswort, 2 reiner Operator
HW=()
HD=()
HEAD=""
HI=-1

# Ergebnis in der globalen STRIPPED; der Guard laeuft vor jedem Bash-Call, Latenz zaehlt
# (SPEC-214 Nr. 10 c).
strip_quotes() {  # fuehrende/abschliessende " und ' entfernen
  local s=$1
  while [ -n "$s" ]; do case $s in \"*|\'*) s=${s#?};; *) break;; esac; done
  while [ -n "$s" ]; do case $s in *\"|*\') s=${s%?};; *) break;; esac; done
  STRIPPED=$s
}

normalize() {  # normalize <absoluter Pfad> -> NORM: lexikalisch, . und leere Komponenten weg, .. aufgeloest
  local IFS=/ comp n=0
  local -a st=()
  for comp in $1; do
    case $comp in
      ''|.) ;;
      ..) if [ "$n" -gt 0 ]; then n=$((n-1)); unset "st[$n]"; fi;;
      *) st[n]=$comp; n=$((n+1));;
    esac
  done
  NORM="/${st[*]}"
}

classify() {  # classify <wort> -> CLS (repo|out|doubt), NPATH; Reihen 1 bis 7 von SPEC-214 Nr. 3 b
  local w=$1
  CLS=doubt
  NPATH=""
  [ -n "$w" ] || return 0
  case $w in
    '$PWD'|'$PWD/'*) w=$ROOT1${w#'$PWD'};;
    '${PWD}'|'${PWD}/'*) w=$ROOT1${w#'${PWD}'};;
    '$CLAUDE_PROJECT_DIR'|'$CLAUDE_PROJECT_DIR/'*) w=$ROOT1${w#'$CLAUDE_PROJECT_DIR'};;
    '${CLAUDE_PROJECT_DIR}'|'${CLAUDE_PROJECT_DIR}/'*) w=$ROOT1${w#'${CLAUDE_PROJECT_DIR}'};;
    *'$'*|*'`'*) return 0;;
    '~'|'~/'*)
      case ${HOME:-} in /*) ;; *) return 0;; esac
      w=$HOME${w#'~'};;
    '~'*) return 0;;
    *'*'*|*'?'*|*'['*) return 0;;
    /*) ;;
    *) if [ "$DIR" = out ]; then CLS=doubt; else CLS=repo; fi; return 0;;
  esac
  if [[ $w == *'/.'* || $w == *'//'* ]]; then normalize "$w"; w=$NORM; fi
  NPATH=$w
  if [[ $w == "$ROOT1" || $w == "$ROOT1"/* || $w == "$ROOT2" || $w == "$ROOT2"/* ]]; then
    CLS=repo
  else
    CLS=out
  fi
}

skip_quote() {  # skip_quote <text> <i>: i zeigt auf ' oder "; QI = Index hinter dem schliessenden Zitat (oder Laenge)
  local t=$1 i=$2 n=${#1} q pre
  q=${t:i:1}
  i=$((i+1))
  if [ "$q" = "'" ]; then
    pre=${t:i}; pre=${pre%%\'*}
    if [ $((i+${#pre})) -ge "$n" ]; then QI=$n; else QI=$((i+${#pre}+1)); fi
    return 0
  fi
  while [ "$i" -lt "$n" ]; do
    pre=${t:i}; pre=${pre%%[\\\"]*}
    i=$((i+${#pre}))
    [ "$i" -lt "$n" ] || break
    if [ "${t:i:1}" = '"' ]; then QI=$((i+1)); return 0; fi
    i=$((i+2))
  done
  QI=$n
}

read_word() {  # read_word <text> <i>: Wort ab i bis Leerraum ausserhalb eines Zitats, ohne Anfuehrungszeichen und Backslashes -> WORD, WEND
  local t=$1 i=$2 n=${#1} c pre inner
  WORD=""
  while [ "$i" -lt "$n" ]; do
    c=${t:i:1}
    case $c in
      ' '|$'\t') break;;
      '\') WORD+=${t:i+1:1}; i=$((i+2));;
      "'"|'"')
        skip_quote "$t" "$i"
        if [ "$QI" -gt "$((i+1))" ] && [ "${t:QI-1:1}" = "$c" ]; then
          inner=${t:i+1:QI-i-2}
        else
          inner=${t:i+1}
        fi
        inner=${inner//\\/}
        WORD+=$inner
        i=$QI;;
      *)
        pre=${t:i}; pre=${pre%%[[:space:]\\\'\"]*}
        WORD+=$pre
        i=$((i+${#pre}));;
    esac
  done
  [ "$i" -le "$n" ] || i=$n
  WEND=$i
}

redir_check() {  # redir_check <segment>: return 0 = eine Umleitung hat das Ziel Repo (SPEC-214 Nr. 4)
  local t=$1 n=${#1} i=0 j c pre r f amp w
  while [ "$i" -lt "$n" ]; do
    pre=${t:i}; pre=${pre%%[\\\'\"\>]*}
    i=$((i+${#pre}))
    [ "$i" -lt "$n" ] || return 1
    c=${t:i:1}
    case $c in
      '\') i=$((i+2));;
      "'"|'"') skip_quote "$t" "$i"; i=$QI;;
      '>')
        if [ "$i" -gt 0 ] && [ "${t:i-1:1}" = '<' ]; then i=$((i+1)); continue; fi
        j=$((i+1)); amp=0
        case ${t:j:1} in
          '>'|'|') j=$((j+1));;
          '&') j=$((j+1)); amp=1;;
        esac
        r=${t:j}
        r=${r#"${r%%[![:space:]]*}"}
        j=$((n-${#r}))
        if [ -z "$r" ]; then return 1; fi
        f=${r:0:1}
        case $f in '>'|'<'|'&'|'|'|'(') i=$j; continue;; esac
        w=${r%%[[:space:]]*}
        if [[ $w == *[\'\"\\]* ]]; then
          read_word "$t" "$j"; w=$WORD; i=$WEND
        else
          i=$((j+${#w}))
        fi
        if [ "$amp" -eq 1 ] && [[ $w == - || $w =~ ^[0-9]+$ ]]; then continue; fi
        w=${w%)}
        classify "$w"
        [ "$CLS" = repo ] && return 0;;
    esac
  done
  return 1
}

get_head() {  # get_head: aus toks (Aufrufer) -> HEAD (Basisname) und HI; HI=-1 ohne Kopf
  local i=0 x
  HEAD=""; HI=-1
  while [ "$i" -lt "$nt" ]; do
    strip_quotes "${toks[i]}"; x=$STRIPPED
    if [[ $x =~ ^[A-Za-z_][A-Za-z0-9_]*= ]]; then i=$((i+1)); continue; fi
    if [[ " $PREFIXES " == *" $x "* ]]; then i=$((i+1)); continue; fi
    # fuehrende Brace-Group-Delimiter ueberspringen: `{ go build; }`
    case $x in "{"|"}") i=$((i+1)); continue;; esac
    HEAD=${x##*/}; HI=$i
    return 0
  done
}

pgrep_check() {  # return 0 = ein Wort -...f... oder --f...
  local i
  for ((i=HI+1; i<nt; i++)); do
    strip_quotes "${toks[i]}"
    case $STRIPPED in
      --f*) return 0;;
      --*) ;;
      -*) [[ ${STRIPPED:1} == *f* ]] && return 0;;
    esac
  done
  return 1
}

redir_word() {  # redir_word <Wort wie geschrieben> -> RW (SPEC-214 Nr. 4 f)
  RW=0
  if [[ $1 =~ $RE_RWORT ]]; then
    RW=1
    if [[ $1 =~ $RE_ROP ]]; then RW=2; fi
  fi
}

rm_check() {  # return 0 = L (SPEC-214 Nr. 7 c); Umleitungswoerter sind weder Operand noch Flag (Nr. 4 f)
  local i x w flag=0 dd=0 skip=0
  local -a ops=()
  for ((i=HI+1; i<nt; i++)); do
    if [ "$skip" -eq 1 ]; then skip=0; continue; fi
    redir_word "${toks[i]}"
    case $RW in 1) continue;; 2) skip=1; continue;; esac
    strip_quotes "${toks[i]}"; x=$STRIPPED
    if [ "$dd" -eq 0 ]; then
      if [ "$x" = -- ]; then dd=1; continue; fi
      case $x in
        --r*) flag=1; continue;;
        --*) continue;;
        -*) [[ ${x:1} == *[rR]* ]] && flag=1; continue;;
      esac
    fi
    w=${toks[i]//[\"\'\\]/}
    ops+=("$w")
  done
  [ "$flag" -eq 1 ] || return 1
  for w in "${ops[@]}"; do
    case $w in *'*'*|*'?'*|*'['*) return 0;; esac
    classify "$w"
    case $CLS in
      doubt) continue;;
      repo) return 0;;
    esac
    [[ $NPATH == */scratchpad/?* ]] || return 0
  done
  return 1
}

tee_check() {  # return 0 = ein Operand hat das Ziel Repo (SPEC-214 Nr. 6 a); Umleitungswoerter sind keine Operanden (Nr. 4 f)
  local i x w dd=0 skip=0
  for ((i=HI+1; i<nt; i++)); do
    if [ "$skip" -eq 1 ]; then skip=0; continue; fi
    x=${toks[i]}
    redir_word "$x"
    case $RW in 1) continue;; 2) skip=1; continue;; esac
    if [ "$dd" -eq 0 ]; then
      if [ "$x" = -- ]; then dd=1; continue; fi
      strip_quotes "$x"
      case $STRIPPED in -*) continue;; esac
    fi
    w=${x//[\"\'\\]/}
    classify "$w"
    [ "$CLS" = repo ] && return 0
  done
  return 1
}

sed_check() {  # return 0 = ein Wort -[A-Za-z0-9]*i... oder --i... (SPEC-214 Nr. 6 b)
  local i
  for ((i=HI+1; i<nt; i++)); do
    strip_quotes "${toks[i]}"
    case $STRIPPED in
      --i*) return 0;;
      --*) ;;
      -*) [[ $STRIPPED =~ ^-[A-Za-z0-9]*i ]] && return 0;;
    esac
  done
  return 1
}

git_check() {  # return 0 = git apply, das schreibt (SPEC-214 Nr. 6 c)
  local i=$((HI+1)) x sub="" si=0 capath="" hasc=0 ro=0 ap=0
  while [ "$i" -lt "$nt" ]; do
    strip_quotes "${toks[i]}"; x=$STRIPPED
    case $x in
      -C) i=$((i+1)); capath=${toks[i]:-}; capath=${capath//[\"\'\\]/}; hasc=1; i=$((i+1));;
      -c|--git-dir|--work-tree|--namespace|--config-env) i=$((i+2));;
      -*) i=$((i+1));;
      *) sub=$x; si=$i; break;;
    esac
  done
  [ "$sub" = apply ] || return 1
  for ((i=si+1; i<nt; i++)); do
    strip_quotes "${toks[i]}"
    case $STRIPPED in
      --check|--stat|--numstat|--summary) ro=1;;
      --apply) ap=1;;
    esac
  done
  if [ "$ro" -eq 1 ] && [ "$ap" -eq 0 ]; then return 1; fi
  if [ "$hasc" -eq 1 ]; then
    classify "$capath"
    [ "$CLS" = repo ] || return 1
  fi
  return 0
}

update_dir() {  # Kopf cd, pushd, popd: DIR fuer die folgenden Segmente (SPEC-214 Nr. 3 c)
  local i x w="" found=0
  if [ "$HEAD" = popd ]; then DIR=out; return 0; fi
  for ((i=HI+1; i<nt; i++)); do
    strip_quotes "${toks[i]}"; x=$STRIPPED
    case $x in
      --) continue;;
      -) w=-; found=1; break;;
      -*) continue;;
      *) w=${toks[i]//[\"\'\\]/}; found=1; break;;
    esac
  done
  [ "$found" -eq 1 ] || w='~'
  case $w in
    -) DIR=out;;
    /*|'~'*|'$'*|*'`'*)
      classify "$w"
      if [ "$CLS" = repo ]; then DIR=in; else DIR=out; fi;;
    *'$'*) DIR=out;;
    *) ;;
  esac
}

find_heredocs() {  # find_heredocs <zeile>: << ausserhalb von Zitaten -> HW (Wort), HD (1 bei <<-)
  local l=$1 n=${#1} i=0 j c w dash pre inner
  while [ "$i" -lt "$n" ]; do
    pre=${l:i}; pre=${pre%%[\\\'\"<]*}
    i=$((i+${#pre}))
    [ "$i" -lt "$n" ] || return 0
    c=${l:i:1}
    case $c in
      '\') i=$((i+2));;
      "'"|'"') skip_quote "$l" "$i"; i=$QI;;
      '<')
        if [ "${l:i:3}" = '<<<' ]; then i=$((i+3)); continue; fi
        if [ "${l:i:2}" != '<<' ]; then i=$((i+1)); continue; fi
        j=$((i+2)); dash=0
        if [ "${l:j:1}" = '-' ]; then dash=1; j=$((j+1)); fi
        while [ "$j" -lt "$n" ]; do
          case ${l:j:1} in ' '|$'\t') j=$((j+1));; *) break;; esac
        done
        w=""
        while [ "$j" -lt "$n" ]; do
          c=${l:j:1}
          case $c in
            ' '|$'\t'|';'|'&'|'|'|'('|')'|'<'|'>') break;;
            '\') w+=${l:j+1:1}; j=$((j+2));;
            "'"|'"')
              skip_quote "$l" "$j"
              if [ "$QI" -gt "$((j+1))" ] && [ "${l:QI-1:1}" = "$c" ]; then
                inner=${l:j+1:QI-j-2}
              else
                inner=${l:j+1}
              fi
              w+=${inner//\\/}
              j=$QI;;
            *) w+=$c; j=$((j+1));;
          esac
        done
        if [ -n "$w" ]; then HW+=("$w"); HD+=("$dash"); fi
        i=$j;;
    esac
  done
}

heredoc_split() {  # heredoc_split <befehl> -> CMDTEXT (ohne Rumpfzeilen), FULLTEXT (mit), HDBODY (Zahl der Rumpfzeilen), UNTERM
  local line t tab=$'\t'
  CMDTEXT=""; FULLTEXT=""; HDBODY=0; UNTERM=0; HW=(); HD=(); PH=0
  while IFS= read -r line; do
    if [ "${#HW[@]}" -gt "$PH" ]; then
      t=$line
      if [ "${HD[PH]}" = 1 ]; then t=${t#"${t%%[!$tab]*}"}; fi
      if [ "$t" = "${HW[PH]}" ]; then
        PH=$((PH+1))
      else
        FULLTEXT+=$line$'\n'
        HDBODY=$((HDBODY+1))
      fi
      continue
    fi
    CMDTEXT+=$line$'\n'
    FULLTEXT+=$line$'\n'
    if [[ $line == *'<<'* ]]; then find_heredocs "$line"; fi
  done <<< "$1"
  if [ "${#HW[@]}" -gt "$PH" ]; then UNTERM=1; fi
}

segment_text() {  # segment_text <text> -> SEGTXT: je Zeile ein Segment
  local s=$1 p1=$'\x01' p2=$'\x02'
  s=${s//'&>'/"${p1}>"}
  s=${s//'>&'/">${p1}"}
  s=${s//'>|'/">${p2}"}
  s=${s//'&&'/$'\n'}; s=${s//'&'/$'\n'}; s=${s//'||'/$'\n'}; s=${s//'|'/$'\n'}
  s=${s//';'/$'\n'};  s=${s//\$\(/$'\n'};  s=${s//'`'/$'\n'}
  s=${s//'('/$'\n'}
  s=${s//"$p1"/'&'}; s=${s//"$p2"/'|'}
  SEGTXT=$s
}

text_has_shell_noc() {  # return 0 = ein Segment ruft eine Shell ohne -c (SPEC-214 Nr. 5 c)
  local s seg i has_c
  local -a toks
  local nt
  segment_text "$1"
  s=$SEGTXT
  while IFS= read -r seg; do
    toks=($seg); nt=${#toks[@]}
    [ "$nt" -gt 0 ] || continue
    get_head
    [ "$HI" -ge 0 ] && [ -n "$HEAD" ] || continue
    if [[ " $SHELLS " == *" $HEAD "* ]]; then
      has_c=0
      for ((i=HI+1; i<nt; i++)); do
        strip_quotes "${toks[i]}"
        if [[ $STRIPPED =~ ^-[a-z]*c[a-z]*$ ]]; then has_c=1; break; fi
      done
      [ "$has_c" -eq 1 ] || return 0
    fi
  done <<< "$s"
  return 1
}

shell_recurse() {  # shell_recurse <tiefe>: -c-String der Shell pruefen; return 0 = Sperre
  local d=$1 j k rest=""
  for ((j=HI+1; j<nt; j++)); do
    strip_quotes "${toks[j]}"
    # -c auch in Flag-Buendeln (-lc, -ec, -cx, ...): bei sh/bash ist c das einzige
    # Single-Letter-Flag mit Kommando-String-Semantik.
    if [[ $STRIPPED =~ ^-[a-z]*c[a-z]*$ ]]; then
      for ((k=j+1; k<nt; k++)); do
        strip_quotes "${toks[k]}"
        rest+="$STRIPPED "
      done
      scan "$rest" "$((d+1))" && return 0
      return 1
    fi
  done
  return 1
}

check_seg() {  # check_seg <segment> <tiefe>; return 0 = Sperre (VERDICT)
  local seg=$1 depth=$2
  local -a toks
  local nt
  toks=($seg); nt=${#toks[@]}
  [ "$nt" -gt 0 ] || return 1
  get_head
  if [ "$HI" -ge 0 ] && [ -n "$HEAD" ]; then
    if [[ " $BLOCKED " == *" $HEAD "* ]]; then VERDICT=H; return 0; fi
    case $HEAD in
      pkill|killall) VERDICT=P; return 0;;
      pgrep) pgrep_check && { VERDICT=P; return 0; };;
      rm) rm_check && { VERDICT=L; return 0; };;
      tee) tee_check && { VERDICT=S; return 0; };;
      sed) sed_check && { VERDICT=S; return 0; };;
      git) git_check && { VERDICT=S; return 0; };;
    esac
  fi
  if [[ $seg == *'>'* ]]; then
    redir_check "$seg" && { VERDICT=S; return 0; }
  fi
  if [ "$HI" -ge 0 ] && [ -n "$HEAD" ]; then
    case $HEAD in cd|pushd|popd) update_dir;; esac
    if [[ " $SHELLS " == *" $HEAD "* ]]; then
      shell_recurse "$depth" && return 0
    fi
  fi
  return 1
}

scan() {  # scan <befehl> <tiefe>; return 0 = Sperre (VERDICT gesetzt), 1 = durch
  local cmd=$1 depth=$2 bsnl=$'\\\n' text s seg
  if [ "$depth" -gt 3 ]; then VERDICT=H; return 0; fi   # zu tief verschachtelt -> fail-closed
  cmd=${cmd//"$bsnl"/ }
  cmd=${cmd//$'\r'/$'\n'}
  text=$cmd
  if [[ $cmd == *'<<'* ]]; then
    heredoc_split "$cmd"
    text=$CMDTEXT
    if [ "$UNTERM" -eq 1 ]; then
      text=$FULLTEXT
    elif [ "$HDBODY" -gt 0 ] && text_has_shell_noc "$CMDTEXT"; then
      text=$FULLTEXT
    fi
  fi
  segment_text "$text"
  s=$SEGTXT
  while IFS= read -r seg; do
    check_seg "$seg" "$depth" && return 0
  done <<< "$s"
  return 1
}

input="$(cat)"

# Ohne awk keine Pruefung -> fail-closed. (awk ist POSIX-Basis.)
command -v awk >/dev/null 2>&1 || { emit H; exit 0; }

cmd="$(printf '%s' "$input" | awk -f "$extractor")"
rc=$?
[ "$rc" -ne 0 ] && { emit H; exit 0; }   # Parse-Zweifel -> fail-closed

if [ "${#cmd}" -gt 32768 ]; then emit G; exit 0; fi

if scan "$cmd" 0; then emit "$VERDICT"; fi
# Pass-Fall: keine Ausgabe — normale Permission-Pruefung uebernimmt.
exit 0
