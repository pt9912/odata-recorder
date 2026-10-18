// Command odata-recorder ist der Startpunkt des OData Recorder; die Verdrahtung
// liegt in src/bootstrap.
package main

import (
	"os"

	"github.com/pt9912/odata-recorder/src/bootstrap"
)

// version setzt der Build per -ldflags (Dockerfile, Stufe build).
var version = "dev"

func main() {
	os.Exit(bootstrap.Run(os.Args[1:], version, os.Stdout, os.Stderr))
}
