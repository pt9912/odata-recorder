package outbound

import (
	"context"
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Gegenstand ist der Gegenstand, auf den sich eine Operation bezog (SPEC-109 Nr. 4).
type Gegenstand int

// Die Gegenstände der Abbildung; GegenstandSonstiges ist der Nullwert.
const (
	GegenstandSonstiges Gegenstand = iota
	GegenstandSession
	GegenstandWorkspace
	GegenstandSnapshot
	GegenstandSnapshotname
	GegenstandImport
	GegenstandExchange
	GegenstandDienst
)

// benannt trägt den Namen der Variable, damit die Tests die Tabelle gegen den Quelltext halten.
type benannt struct {
	name string
	err  error
}

// zeile ist ein Eintrag der Tabelle in SPEC-109 Nr. 5: entweder ein Code für jeden Gegenstand
// oder ein Code je genanntem Gegenstand.
type zeile struct {
	fehler []benannt
	jeder  model.Code
	je     map[Gegenstand]model.Code
}

// tabelle ist die Abbildung aus SPEC-109 Nr. 5, in der Reihenfolge der Spezifikation.
func tabelle() []zeile {
	return []zeile{
		{fehler: []benannt{{"ErrBlobFehlt", ErrBlobFehlt}, {"ErrBlobBeschaedigt", ErrBlobBeschaedigt}, {"ErrBeschaedigt", ErrBeschaedigt}}, jeder: model.CodeE3001},
		{fehler: []benannt{{"ErrPfad", ErrPfad}}, jeder: model.CodeE2001},
		{fehler: []benannt{{"ErrZuGross", ErrZuGross}}, jeder: model.CodeE2401},
		{fehler: []benannt{{"ErrUpstreamNichtErreichbar", ErrUpstreamNichtErreichbar}, {"ErrUpstreamZeit", ErrUpstreamZeit}}, jeder: model.CodeE4001},
		{fehler: []benannt{{"ErrFremderService", ErrFremderService}}, jeder: model.CodeE5004},
		{fehler: []benannt{{"ErrDienstUnbekannt", ErrDienstUnbekannt}}, jeder: model.CodeE5005},
		{fehler: []benannt{{"ErrFalscheArt", ErrFalscheArt}}, jeder: model.CodeE7008},
		{fehler: []benannt{{"ErrReferenziert", ErrReferenziert}}, jeder: model.CodeE7004},
		{fehler: []benannt{{"ErrSessionGeschlossen", ErrSessionGeschlossen}}, jeder: model.CodeE5009},
		{fehler: []benannt{{"ErrUngueltigerUebergang", ErrUngueltigerUebergang}}, jeder: model.CodeE7011},
		{fehler: []benannt{{"ErrNichtGefunden", ErrNichtGefunden}}, je: map[Gegenstand]model.Code{
			GegenstandSession: model.CodeE7003, GegenstandWorkspace: model.CodeE7001, GegenstandSnapshot: model.CodeE7002,
			GegenstandImport: model.CodeE7005, GegenstandExchange: model.CodeE7012, GegenstandDienst: model.CodeE5005,
		}},
		{fehler: []benannt{{"ErrBereitsVorhanden", ErrBereitsVorhanden}}, je: map[Gegenstand]model.Code{
			GegenstandSession: model.CodeE7010, GegenstandWorkspace: model.CodeE7007,
			GegenstandSnapshotname: model.CodeE7006, GegenstandImport: model.CodeE6102,
		}},
		{fehler: []benannt{{"ErrUngueltig", model.ErrUngueltig}}, je: map[Gegenstand]model.Code{GegenstandImport: model.CodeE6101}},
		{fehler: []benannt{{"ErrMetadatenUngueltig", model.ErrMetadatenUngueltig}}, jeder: model.CodeE6004},
		{fehler: []benannt{{"ErrUngueltigeSequenz", ErrUngueltigeSequenz}, {"ErrAbgeschlossen", ErrAbgeschlossen}, {"ErrRevisionVeraltet", ErrRevisionVeraltet}}, jeder: model.CodeE9001},
	}
}

// EintraegeAbbildung liefert die Namen der Fehler, die die Tabelle nennt; die Tests halten sie
// gegen die Variablen `Err…` des Pakets (SPEC-109 Nr. 8).
func EintraegeAbbildung() []string {
	var namen []string
	for _, z := range tabelle() {
		for _, f := range z.fehler {
			namen = append(namen, f.name)
		}
	}
	return namen
}

// Abbilden ordnet einen Fehler der Outbound Ports einem Code zu (SPEC-109 Nr. 5): Der erste
// zutreffende Eintrag der Tabelle bestimmt den Code, ein Gegenstand, den die Zeile nicht nennt,
// und jeder unbekannte Fehler fallen auf ODR-E9001. Bei nil und bei einem abgebrochenen oder
// abgelaufenen Kontext liefert sie false: kein Code (Nr. 6).
func Abbilden(err error, g Gegenstand) (model.Code, bool) {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return 0, false
	}
	for _, z := range tabelle() {
		if !z.trifft(err) {
			continue
		}
		if z.je == nil {
			return z.jeder, true
		}
		if c, ok := z.je[g]; ok {
			return c, true
		}
		return model.CodeE9001, true
	}
	return model.CodeE9001, true
}

func (z zeile) trifft(err error) bool {
	for _, f := range z.fehler {
		if errors.Is(err, f.err) {
			return true
		}
	}
	return false
}
