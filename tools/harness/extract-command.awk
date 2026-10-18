# extract-command.awk — zieht tool_input.command aus der PreToolUse-Hook-JSON.
# POSIX-awk (busybox/gawk/BSD), kein gawk-Spezifikum. Stdout = dekodierter
# Befehl; Exit 0 = ok (ggf. leer), Exit 3 = Parse-Zweifel/fail-closed — der
# Guard blockt dann. \u-Escape im Befehl gilt als Zweifel (block): der Guard
# ist ein Stolperdraht, keine Sandbox. Vollstaendigkeit ist nicht das Ziel —
# bei Unsicherheit lieber blocken (fail-closed).
#
# Zeichenweiser Scanner mit Tiefen-/Key-Stack: er unterscheidet JSON-Keys von
# -Values und entschaerft so den "command-im-Value"-Fehlmatch. Nur der Pfad
# tool_input -> command (Objekt-Tiefe 2) zaehlt.

{ doc = (NR == 1) ? $0 : doc "\n" $0 }

END {
  n = length(doc)
  depth = 0       # Verschachtelungstiefe ({}/[])
  instr = 0       # in einem JSON-String?
  esc = 0         # letztes Zeichen war Backslash?
  buf = ""        # aktueller String-Inhalt (dekodiert)
  hadu = 0        # aktueller String enthielt \uXXXX
  sawobj = 0      # je ein Top-Level-Objekt gesehen?
  found = 0
  cmdval = ""

  for (i = 1; i <= n; i++) {
    c = substr(doc, i, 1)

    if (instr) {
      if (esc) {
        esc = 0
        if (c == "\"") buf = buf "\""
        else if (c == "\\") buf = buf "\\"
        else if (c == "/") buf = buf "/"
        else if (c == "n") buf = buf "\n"
        else if (c == "t") buf = buf "\t"
        else if (c == "r") buf = buf "\r"
        else if (c == "b") buf = buf sprintf("%c", 8)
        else if (c == "f") buf = buf sprintf("%c", 12)
        else if (c == "u") {
          # \u verlangt GENAU 4 Hex. Sonst malformed JSON -> fail-closed
          # (sonst desynct ein i+=4 ueber ein schliessendes " hinweg den
          # Scanner und der Guard koennte fail-OPEN gehen).
          if (substr(doc, i + 1, 4) !~ /^[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]$/) exit 3
          hadu = 1; i = i + 4
        }
        else buf = buf c                              # unbekannter Escape: Zeichen behalten
        continue
      }
      if (c == "\\") { esc = 1; continue }
      if (c == "\"") {
        # Stringende: Key oder Value?
        if (depth > 0 && ctype[depth] == "o" && wantkey[depth] == 1) {
          curkey[depth] = buf
        } else if (depth >= 2 && ctype[depth] == "o" && curkey[depth] == "command" &&
                   ctype[depth - 1] == "o" && curkey[depth - 1] == "tool_input") {
          if (hadu) exit 3
          found = 1
          cmdval = buf
        }
        instr = 0
        continue
      }
      buf = buf c
      continue
    }

    # ausserhalb eines Strings
    if (c == "\"") { instr = 1; buf = ""; hadu = 0; continue }
    if (c == "{") { depth++; sawobj = 1; ctype[depth] = "o"; wantkey[depth] = 1; curkey[depth] = ""; continue }
    if (c == "}") { if (depth > 0) depth--; continue }
    if (c == "[") { depth++; ctype[depth] = "a"; continue }
    if (c == "]") { if (depth > 0) depth--; continue }
    if (c == ":") { if (depth > 0 && ctype[depth] == "o") wantkey[depth] = 0; continue }
    if (c == ",") { if (depth > 0 && ctype[depth] == "o") wantkey[depth] = 1; continue }
  }

  if (!sawobj) exit 3                      # kein Objekt -> kein/kaputtes JSON -> block
  if (instr == 1 || depth != 0) exit 3     # abgeschnitten/unbalanciert -> block
  if (found) printf "%s", cmdval
  exit 0
}
