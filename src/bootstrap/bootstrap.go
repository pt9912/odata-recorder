// Package bootstrap ist der Composition Root des OData Recorder: Es verdrahtet
// Kommando, Ein- und Ausgabe und liefert den Exit-Code des Prozesses.
package bootstrap

import (
	"fmt"
	"io"
)

// Exit-Codes von Run.
const (
	exitOK       = 0
	exitAufruf   = 2
	kommandoName = "odata-recorder"
)

// Run führt das Kommando aus args (ohne Programmnamen) aus und gibt den Exit-Code zurück.
// Ohne Argument und mit `version` schreibt es "odata-recorder <version>" nach stdout und
// liefert 0; jedes andere Kommando schreibt eine Fehlermeldung nach stderr und liefert 2.
func Run(args []string, version string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (len(args) == 1 && args[0] == "version") {
		_, _ = fmt.Fprintf(stdout, "%s %s\n", kommandoName, version)
		return exitOK
	}
	_, _ = fmt.Fprintf(stderr, "Unbekanntes Kommando: %s\n", args[0])
	return exitAufruf
}
