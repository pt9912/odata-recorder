// Command uebergang ist das Hilfsprogramm der Stufen von tests/e2e (SPEC-037 Nr. 6): Es legt Import und
// Workspace über denselben DATA_DIR an und liest den Stand einer Session. Es ist Teil des Tests, nicht des
// Produkts, und schreibt nur über die Schnittstellen des Composition Root.
package main

import (
	"context"
	"os"
)

func main() {
	os.Exit(lauf(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}
