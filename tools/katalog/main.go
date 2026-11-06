// Command katalog hält den Katalog der Fehlercodes gegen das Handbuch (SPEC-201): `katalog <Wurzel>`.
// Es liest Dateien unter der Wurzel, vergleicht Mengen von Namen und schreibt nur nach stdout und stderr.
package main

import (
	"fmt"
	"io"
	"os"
)

const aufrufText = "Aufruf: katalog <Wurzel>\n"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run prüft den Aufruf (SPEC-201 Nr. 2 a), liest die Quellen und gibt den Bericht aus (Nr. 2 b).
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] == "" {
		_, _ = fmt.Fprint(stderr, aufrufText)
		return 2
	}
	return pruefeFehlercodes(args[0]).ausgeben(stdout, stderr)
}
