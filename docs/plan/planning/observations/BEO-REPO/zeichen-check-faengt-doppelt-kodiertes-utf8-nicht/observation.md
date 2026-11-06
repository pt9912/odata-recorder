# `make zeichen-check` fängt doppelt kodiertes UTF-8 nicht

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Schreibweg, der UTF-8-Bytes als Latin-1 liest und wieder als UTF-8 schreibt (`perl` ohne Kodierung), macht aus `ä` die Bytefolge `C3 83 C2 A4`. Das Ergebnis ist gültiges UTF-8 und trägt nur Zeichen außerhalb der 253 des Gates (`SPEC-233`); kein Gate meldet es, und der Diff zeigt eine unauffällige Zeile mit großem A, Tilde und Währungszeichen.
