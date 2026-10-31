package telemetry

import "github.com/pt9912/odata-recorder/src/hexagon/ports/support"

// Filtern macht filtern für die Tests erreichbar; op "" ist die Menge Log.
func Filtern(op string, attrs []support.Attr) ([]support.Attr, []support.Verwurf) {
	return filtern(operation(op), attrs)
}

// MaxSpans macht den gehaltenen Wert von MaxSpans für die Tests erreichbar.
func (a *Adapter) MaxSpans() int { return a.maxSpans }
