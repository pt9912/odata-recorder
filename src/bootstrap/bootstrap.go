// Package bootstrap ist der Composition Root des OData Recorder: Es verdrahtet
// Kommando, Ein- und Ausgabe und liefert den Exit-Code des Prozesses.
package bootstrap

import (
	"fmt"
	"io"
)

// Exit-Codes von Run.
const (
	exitOK     = 0
	exitAufruf = 2
)

// kommandoName ist der Name, den die Versionsausgabe voranstellt.
const kommandoName = "odata-recorder"

// Run führt das Kommando aus args (ohne Programmnamen) aus und gibt den Exit-Code zurück.
// Ohne Argument und mit genau dem Argument `version` schreibt es "odata-recorder <version>"
// nach stdout und liefert 0. Jedes andere Aufrufmuster (unbekanntes Kommando, leeres Argument,
// Flag wie `--help`, `version` mit Zusatzargument) schreibt "Unbekanntes Kommando: <erstes
// Argument>" nach stderr und liefert 2. Ein nil-Writer verwirft die Ausgabe.
func Run(args []string, version string, stdout, stderr io.Writer) int {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if len(args) == 0 || (len(args) == 1 && args[0] == "version") {
		_, _ = fmt.Fprintf(stdout, "%s %s\n", kommandoName, version)
		return exitOK
	}
	_, _ = fmt.Fprintf(stderr, "Unbekanntes Kommando: %s\n", args[0])
	return exitAufruf
}
