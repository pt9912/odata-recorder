#!/usr/bin/env bash
# tools/harness/e2e-abdeckung.sh — Erzeuger der E2E-Abdeckungs-Sicht, emittiert von
# ai-harness-init. EIN KOMMANDO, KEIN GATE.
#
# WAS ER TUT. Er liest den TEXT des E2E-Skripts dieses Repos und schreibt daraus eine
# Tabelle: je Stufe eine Zeile mit der Anforderung, die sie traegt, einer
# Kurzbeschreibung, der Nummer der Stufe und dem Ort, an dem sie steht. Kein Docker,
# kein Netz, kein E2E-Lauf — die Sicht aendert sich mit den DEKLARATIONEN, nicht mit
# jedem Lauf. Darum ist das Kommando kein Gate: es urteilte ueber den Quelltext eines
# Skripts, nicht ueber den Zustand des Baums.
#
# DIE STUFEN-MENGE IST EIN KRITERIUM, KEINE AUFZAEHLUNG: eine Zeile der Form
#   echo "<Praefix>: … ..."
# eroeffnet eine Stufe, ihre Region reicht bis zur naechsten solchen Zeile — die letzte
# bis zum Dateiende. In der Region steht ein Aufruf
#   e2e_abdeckung "<Kennungen>" "<Kurzbeschreibung>" "<Anker>"
# und der Anker ist ein woertlicher Ausschnitt aus einer ANDEREN Zeile derselben Stufe.
# DREI ZEILEN, NICHT ZWEI: die Deklaration ist ein Funktionsaufruf, und ihre Funktion
# gehoert dazu — einmal, vor der ersten Stufe. Wer nur Kopfzeile und Aufruf uebernimmt,
# bekommt beim Lauf seines E2E ein `command not found`:
#   e2e_abdeckung() { echo "abdeckung: $1 — $2 (Anker: $3)"; }
# Was die Funktion ausgibt, ist frei; dieser Erzeuger liest den TEXT des Aufrufs.
#
# DIE ZWEI LUECKEN-RICHTUNGEN, beide laut:
#   (a) DEKLARATION OHNE STUFE — der Anker loest in der Region SEINER Stufe nicht
#       woertlich auf, oder der Aufruf steht vor der ersten Stufe.
#   (b) STUFE OHNE DEKLARATION — eine Region traegt keinen Aufruf.
# Beide enden mit Exit 1 und nennen Stufe, Region und bei (a) den Anker.
#
# VIER MARKER, JE MIT BELEGUNG. Sie sind Variablen, keine Platzhalter zum
# Suchen-und-Ersetzen:
#   E2E_ABDECKUNG_QUELLE   das E2E-Skript, dessen Stufen gelesen werden
#   E2E_ABDECKUNG_PRAEFIX  das Wort, mit dem eine Stufen-Kopfzeile dieses Skripts beginnt
#   E2E_ABDECKUNG_SPEC     die Spec-Datei, gegen die der LESER die Sicht haelt; die Sicht
#                          nennt sie in ihrem Kopf
#   E2E_ABDECKUNG_ZIEL     die zu schreibende Sicht
# Der Lauf nennt in seiner ersten Zeile die Werte, mit denen er faehrt.
#
# DIE KENNUNGSSPALTE TRAEGT CODE-SPANS, KEINE VERWEISE — und das ist eine Entscheidung,
# keine fehlende Bequemlichkeit. Ein Verweis braucht einen ANKER, und ein Anker entsteht
# aus dem GERENDERTEN Text einer Ueberschrift. Wer ihn aus der Rohzeile ableitet, bildet
# das Verhalten eines Markdown-Renderers nach: Satzzeichen, Inline-Syntax, Whitespace,
# Entities — jede Regel, die man dabei nicht trifft, ergibt einen Link, der AUSSIEHT wie
# ein Verweis, bei Exit 0 entsteht und erst im Doku-Gate des Adopters rot wird. Dieses
# Werkzeug kennt das Doku-Gate des Ziels nicht und bringt keines mit; es kann die
# Nachbildung also weder kalibrieren noch halten. Ein Code-Span behauptet nichts, wird
# nie tot und faerbt kein Gate rot. Was der Leser braucht, steht daneben: die Sicht nennt
# im Kopf die Spec-Datei, gegen die er sie haelt.
#
# UND EINE STUFE DARF NOCH KEINE KENNUNG HABEN. Steht im ersten Argument der Deklaration
# allein ein Gedankenstrich, heisst das: dieses Repo fuehrt fuer diese Stufe (noch) keine
# Anforderung. Die Zelle traegt dann denselben Gedankenstrich, und der Lauf nennt die
# Stufe. Das ist der ehrliche Zustand einer mitgelieferten Stufe: eine geratene Kennung
# loeste vielleicht auf und behauptete trotzdem eine Zuordnung, die niemand getroffen hat.
#
# DIESE DATEI IST KONVERGENT: jeder Lauf des Werkzeugs schreibt sie kanonisch neu, und
# ein Edit an ihr ist danach still weg — deshalb sind die vier Stellen oben Variablen.
# Gesetzt wird am Aufruf
#   make e2e-abdeckung E2E_ABDECKUNG_QUELLE=tools/harness/mein-e2e.sh
# oder dauerhaft in repo.mk an der Wurzel, der Datei des Repos:
#   E2E_ABDECKUNG_QUELLE = tools/harness/mein-e2e.sh
# Der Aggregator bindet sie nach den Fragmenten ein (`-include repo.mk`), und ein Lauf
# legt sie nur an, wo sie fehlt.
#
# SCHREIBEN NUR BEI ABWEICHUNG: gerendert wird in eine Temp-Datei, danach verglichen.
# Ueber identischem Inhalt meldet der Lauf "unveraendert" und laesst das Ziel stehen.
#
# ABHAENGIGKEIT. bash und coreutils (grep/sed/awk/cmp/mktemp). Kein git, kein Docker,
# kein Netz.
set -euo pipefail

E2E_ABDECKUNG_QUELLE="${E2E_ABDECKUNG_QUELLE:-tools/harness/selbstpruefung.sh}"
E2E_ABDECKUNG_PRAEFIX="${E2E_ABDECKUNG_PRAEFIX:-selbstpruefung}"
E2E_ABDECKUNG_SPEC="${E2E_ABDECKUNG_SPEC:-spec/lastenheft.md}"
E2E_ABDECKUNG_ZIEL="${E2E_ABDECKUNG_ZIEL:-docs/user/e2e-abdeckung.md}"

quelle="$E2E_ABDECKUNG_QUELLE"
ziel="$E2E_ABDECKUNG_ZIEL"
spec="$E2E_ABDECKUNG_SPEC"

STUFEN_MUSTER="^echo \"$E2E_ABDECKUNG_PRAEFIX: .* \\.\\.\\.\"\$"
# Das Muster gilt ueber den Zeilen-ANGABEN von `grep -n` (Praefix `<zeile>:`), damit
# dieselbe Zeichenkette die Form prueft, die spaeter zerlegt wird.
RUF_MUSTER='^[0-9]+:[[:space:]]*e2e_abdeckung "[^"]*" "[^"]*" "[^"]*"[[:space:]]*$'
# Dieselbe Zeile OHNE die Zeilen-Angabe: sie erkennt eine Deklarations-Zeile dort, wo
# ueber Text gelaufen wird.
RUF_TEIL='^[[:space:]]*e2e_abdeckung "'
# Der Backtick der Orts- und der Kennungsspalte steht als VARIABLE: woertlich im
# printf-Format liest shellcheck ihn als Kommando-Substitution (SC2016), und eine
# Inline-Suppression ist gesperrt.
BT='`'
# Das erste Argument einer Deklaration, das KEINE Kennung ist, sondern die Auskunft,
# dass dieses Repo fuer die Stufe (noch) keine Anforderung fuehrt.
OHNE_KENNUNG='—'

echo "e2e-abdeckung: Marker — Quelle=[$quelle] Praefix=[$E2E_ABDECKUNG_PRAEFIX] Spec=[$spec] Ziel=[$ziel]"

if [ ! -f "$quelle" ]; then
	echo "e2e-abdeckung: FEHLER — die Quelle liegt nicht: $quelle — der Marker E2E_ABDECKUNG_QUELLE nennt das E2E-Skript, dessen Stufen gelesen werden." >&2
	exit 1
fi

# DIE SPEC-DATEI IST KEINE VORBEDINGUNG, SONDERN DER MASSSTAB DES LESERS. Gelesen wird
# sie nicht — die Kennungsspalte traegt Code-Spans, keine Anker. Fehlt sie, entsteht die
# Sicht vollstaendig; was ihr fehlt, ist der Massstab. GESAGT WIRD ES TROTZDEM, an zwei
# Stellen: hier im Lauf und im Kopf der geschriebenen Sicht. Eine Sicht, die einen
# Massstab nennt, den es nicht gibt, behauptet Pruefbarkeit, die niemand einloesen kann.
# KEIN ABBRUCH: ein Abbruch machte aus diesem Kommando ein Gate ueber dem Zustand des
# Baums, und genau das ist es nicht.
spec_da=1
if [ ! -f "$spec" ]; then
	spec_da=0
	echo "e2e-abdeckung: Hinweis — die Spec-Datei liegt nicht: $spec. Die Sicht entsteht trotzdem und ist vollstaendig; was ihr fehlt, ist der Massstab, gegen den ihr Leser sie haelt — ihr Kopf sagt das. Der Marker E2E_ABDECKUNG_SPEC nennt die Datei."
fi

stufen="$(grep -nE "$STUFEN_MUSTER" "$quelle" || true)"
if [ -z "$stufen" ]; then
	echo "e2e-abdeckung: FEHLER — $quelle fuehrt keine Stufen-Kopfzeile, und eine Sicht ueber null Stufen sagt nichts: ihr Gruen belegte eine Abdeckung, die niemand deklariert hat." >&2
	echo "  Erwartete Form einer Kopfzeile (eine Zeile, woertlich so):" >&2
	echo "    echo \"$E2E_ABDECKUNG_PRAEFIX: <was die Stufe tut> ...\"" >&2
	echo "  Und in ihrer Region, auf einer eigenen Zeile, die Deklaration:" >&2
	echo "    e2e_abdeckung \"<Kennungen>\" \"<Kurzbeschreibung>\" \"<Anker aus einer anderen Zeile dieser Stufe>\"" >&2
	echo "  Die Deklaration ist ein Funktionsaufruf und braucht ihre Funktion — EINE Zeile, irgendwo" >&2
	echo "  vor der ersten Stufe; ohne sie endet der Lauf des E2E mit 'command not found':" >&2
	echo "    e2e_abdeckung() { echo \"abdeckung: \$1 — \$2 (Anker: \$3)\"; }" >&2
	echo "  Faehrt dieses Repo sein E2E woanders oder mit einem anderen Wort am Zeilenanfang, nennen es die Marker:" >&2
	echo "    make e2e-abdeckung E2E_ABDECKUNG_QUELLE=<pfad> E2E_ABDECKUNG_PRAEFIX=<wort>" >&2
	echo "  Dauerhaft in repo.mk an der Wurzel, der Datei des Repos, die ein Lauf nur anlegt, wo sie fehlt." >&2
	exit 1
fi

aufrufe="$(grep -nE '^[[:space:]]*e2e_abdeckung ' "$quelle" || true)"
unleserlich="$(grep -vE "$RUF_MUSTER" <<<"$aufrufe" || true)"
if [ -n "$unleserlich" ]; then
	echo "e2e-abdeckung: FEHLER — Deklarations-Zeile in unleserlicher Form; erwartet wird e2e_abdeckung \"<Kennungen>\" \"<Kurzbeschreibung>\" \"<Anker>\":" >&2
	printf '%s\n' "$unleserlich" >&2
	exit 1
fi

# Alles VOR der ersten Stufe ist eine Deklaration ohne Stufe; danach kacheln die
# Regionen der Stufen die Datei lueckenlos.
erste_stufe="$(sed -n '1p' <<<"$stufen" | cut -d: -f1)"
davor="$(awk -F: -v z="$erste_stufe" '$1 < z { print }' <<<"$aufrufe")"
if [ -n "$davor" ]; then
	echo "e2e-abdeckung: FEHLER — Deklaration ohne Stufe: der Aufruf steht VOR der ersten Stufen-Kopfzeile ($quelle:$erste_stufe):" >&2
	printf '%s\n' "$davor" >&2
	exit 1
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

{
	echo "# E2E-Abdeckung der Stufen des Voll-E2E"
	echo
	echo "Erzeugt von \`make e2e-abdeckung\` aus den Deklarationen, die jede Stufe von"
	echo "\`$quelle\` an sich selbst trägt. Diese Datei ist eine **stabile"
	echo "Abdeckungs-Deklaration, kein Lauf-Beleg**: der Erzeuger liest den Quelltext der Stufen,"
	echo "er führt den E2E nicht aus, und er schreibt sie nur bei inhaltlicher Abweichung. Sie"
	echo "ändert sich mit den Deklarationen und mit dem Ort ihrer Quellen — eine Einfügung"
	echo "oberhalb einer Stufe verschiebt deren Zeile in der Spalte \`Ort\`."
	echo
	echo "Die Spalte \`Spec-Kennung\` nennt die Anforderung, die diese Stufe trägt, **als"
	echo "Code-Span statt als Verweis**: ein Verweis brauchte einen Anker, und den müsste dieser"
	echo "Erzeuger aus der Rohzeile einer Überschrift nachbilden — ein Link, der aussieht wie"
	echo "einer und im Doku-Gate rot wird, wäre teurer als gar keiner. Gehalten wird die Sicht"
	if [ "$spec_da" -eq 1 ]; then
		echo "deshalb gegen \`$spec\`, und zwar vom Leser. Die Spalte"
	else
		# DER MASSSTAB FEHLT, UND DIE SICHT SAGT ES SELBST. Ohne diesen Zweig nennte sie
		# eine Datei, die es nicht gibt, und behauptete damit Pruefbarkeit.
		echo "deshalb vom Leser gegen \`$spec\` — **diese Datei liegt in diesem Repo derzeit"
		echo "nicht**; solange sie fehlt, hat die Sicht keinen Maßstab. Die Spalte"
	fi
	echo "\`Kurzbeschreibung\` trägt keine Kennung. Die Spalte \`Stufe\` zählt die"
	echo "Stufen-Kopfzeilen in der Reihenfolge des Skripts — eine Stufe eröffnet mit ihrer"
	echo "Ausgabe-Kopfzeile und reicht bis zur nächsten. Die Spalte \`Ort\` adressiert sie."
	echo "Ob eine hier fehlende Anforderung eine Lücke ist, urteilt ebenfalls der Leser; ein"
	echo "Waisen-Urteil fällt nicht hier."
	echo
	echo "| Spec-Kennung | Kurzbeschreibung | Stufe | Ort |"
	echo "| --- | --- | --- | --- |"
} >"$tmp"

zeilen_gesamt="$(wc -l <"$quelle")"
stufen_gesamt=0
deklarationen=0
ohne_kennung=0
while IFS=: read -r start _rest; do
	stufen_gesamt=$((stufen_gesamt + 1))
	naechste="$(awk -F: -v z="$start" '$1 > z { print $1; exit }' <<<"$stufen")"
	if [ -n "$naechste" ]; then
		ende=$((naechste - 1))
	else
		ende="$zeilen_gesamt"
	fi
	titel="$(sed -n "${start}p" "$quelle" | sed -e "s/^echo \"$E2E_ABDECKUNG_PRAEFIX: //" -e 's/ \.\.\."$//')"

	region_aufrufe="$(awk -F: -v s="$start" -v e="$ende" '$1 >= s && $1 <= e' <<<"$aufrufe")"
	if [ -z "$region_aufrufe" ]; then
		echo "e2e-abdeckung: FEHLER — Stufe ohne Deklaration: in der Region dieser Stufe steht kein Aufruf von e2e_abdeckung." >&2
		echo "  Stufe $stufen_gesamt: $titel" >&2
		echo "  Region: $quelle:$start-$ende" >&2
		echo "  Nachweis: sed -n '${start},${ende}p' $quelle" >&2
		exit 1
	fi

	while IFS=: read -r rufzeile ruftext; do
		kennungen="$(sed -n 's/^[[:space:]]*e2e_abdeckung "\([^"]*\)" "[^"]*" "[^"]*"[[:space:]]*$/\1/p' <<<"$ruftext")"
		kurz="$(sed -n 's/^[[:space:]]*e2e_abdeckung "[^"]*" "\([^"]*\)" "[^"]*"[[:space:]]*$/\1/p' <<<"$ruftext")"
		anker="$(sed -n 's/^[[:space:]]*e2e_abdeckung "[^"]*" "[^"]*" "\([^"]*\)"[[:space:]]*$/\1/p' <<<"$ruftext")"
		case "$kurz" in
		*"|"*)
			echo "e2e-abdeckung: FEHLER — die Kurzbeschreibung fuehrt ein '|', das die Tabellen-Zelle zerlegte ($quelle:$rufzeile)." >&2
			exit 1
			;;
		esac
		# Die Deklarations-Zeilen sind von der Anker-Suche AUSGENOMMEN: der Anker steht
		# wortwoertlich im dritten Argument des Aufrufs. Die Ausnahme ist der Grund, warum
		# der Ort auf eine Zeile der STUFE zeigt und nicht auf die Deklaration.
		ort="$(awk -v s="$start" -v e="$ende" -v a="$anker" -v rufmuster="$RUF_TEIL" \
			'NR >= s && NR <= e && $0 !~ rufmuster && index($0, a) { print NR; exit }' "$quelle")"
		if [ -z "$ort" ]; then
			echo "e2e-abdeckung: FEHLER — Deklaration ohne Stufe: der Anker loest in der Region der Stufe, die sie deklariert, nicht woertlich auf." >&2
			echo "  Deklaration: $quelle:$rufzeile (Stufe $stufen_gesamt: $titel)" >&2
			echo "  Anker: [$anker]" >&2
			echo "  Region: $quelle:$start-$ende" >&2
			echo "  Nachweis: sed -n '${start},${ende}p' $quelle" >&2
			exit 1
		fi
		if [ "$kennungen" = "$OHNE_KENNUNG" ]; then
			# KEINE ZUORDNUNG BEHAUPTEN, WO KEINE GETROFFEN IST: die Zelle traegt den
			# Gedankenstrich der Deklaration weiter, und die Stufe steht trotzdem in der
			# Sicht — sie laeuft ja.
			zelle="$OHNE_KENNUNG"
			ohne_kennung=$((ohne_kennung + 1))
			echo "e2e-abdeckung: Hinweis — Stufe $stufen_gesamt deklariert keine Kennung ($quelle:$rufzeile); die Zelle traegt $OHNE_KENNUNG. Eine Anforderung dieses Repos traegt sie, sobald die Deklaration sie nennt."
		else
			# JEDE KENNUNG ALS CODE-SPAN. Kein Nachschlagen, keine Anker-Ableitung, kein
			# Zustand, in dem ein Link entstuende, den die Zieldatei nicht traegt.
			zelle=""
			for k in $kennungen; do
				if [ -n "$zelle" ]; then
					zelle="$zelle, "
				fi
				zelle="${zelle}${BT}${k}${BT}"
			done
		fi
		printf '| %s | %s | Stufe %s | %s%s:%s%s |\n' "$zelle" "$kurz" "$stufen_gesamt" "$BT" "$quelle" "$ort" "$BT" >>"$tmp"
		deklarationen=$((deklarationen + 1))
	done <<<"$region_aufrufe"
done <<<"$stufen"

if [ "$deklarationen" -ne "$stufen_gesamt" ]; then
	echo "e2e-abdeckung: FEHLER — $stufen_gesamt Stufen, aber $deklarationen Deklarationen geschrieben; jede Stufe traegt genau die ihre." >&2
	exit 1
fi

if [ -f "$ziel" ] && cmp -s "$tmp" "$ziel"; then
	echo "e2e-abdeckung: unveraendert — $ziel ($stufen_gesamt Stufen, $deklarationen Deklarationen, $ohne_kennung ohne Kennung)."
else
	ziel_dir="$(dirname "$ziel")"
	if [ ! -d "$ziel_dir" ]; then
		mkdir -p "$ziel_dir"
	fi
	# Der Modus steht EXPLIZIT: `mktemp` legt die Temp-Datei mit 0600 an, und die Sicht
	# wird von einem Doku-Gate gelesen, das in einem Container als Nicht-Root laeuft.
	rm -f "$ziel"
	cat "$tmp" >"$ziel"
	chmod 0644 "$ziel"
	echo "e2e-abdeckung: geschrieben — $ziel ($stufen_gesamt Stufen, $deklarationen Deklarationen aus $quelle, $ohne_kennung ohne Kennung)."
fi
