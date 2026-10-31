// Command odata-recorder ist der Startpunkt des OData Recorder; die Verdrahtung
// liegt in src/bootstrap.
package main

import (
	"os"

	"github.com/pt9912/odata-recorder/src/adapters/driving/cli"
)

// version setzt der Build per -ldflags (Dockerfile, Stufe build).
var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], cli.Konfig{Version: version, Stdout: os.Stdout, Stderr: os.Stderr}))
}
