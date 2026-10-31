// Command zeichen sucht in den Dateien eines Arbeitsbaums nach Zeichen, die dort nicht hingehören:
// BOM, geschütztes Leerzeichen, Zeilen- und Absatztrenner, Nullbreiten- und übrige Formatzeichen,
// rohe Steuerzeichen (Menge Z) und ungültiges UTF-8 (SPEC-233): `zeichen pruefen <Wurzel>`. Es liest
// nur unter der Wurzel und schreibt nur nach stdout und stderr.
package main

import (
	"fmt"
	"io"
	"os"
)

const aufrufText = "Aufruf: zeichen pruefen <Wurzel>\n"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run prüft den Aufruf (SPEC-233 Nr. 13) und fährt die Prüfung (Nr. 11 a: Exit 0, 1 oder 2).
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "pruefen" || args[1] == "" {
		_, _ = fmt.Fprint(stderr, aufrufText)
		return 2
	}
	return neuerPruefer(stdout, stderr).pruefen(args[1])
}
