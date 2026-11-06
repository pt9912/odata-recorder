// Command subst ersetzt in einer Datei genau eine Stelle (literal) und endet mit 3, wenn die Zahl der Treffer nicht eins ist.
// Aufruf: subst Datei alt neu. Hilfe des Nachbaus der Messung, kein Teil des Werkzeugs.
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "Aufruf: subst Datei alt neu")
		os.Exit(2)
	}
	roh, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if n := strings.Count(string(roh), os.Args[2]); n != 1 {
		fmt.Fprintf(os.Stderr, "subst: %d Treffer in %s\n", n, os.Args[1])
		os.Exit(3)
	}
	if err := os.WriteFile(os.Args[1], []byte(strings.Replace(string(roh), os.Args[2], os.Args[3], 1)), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
