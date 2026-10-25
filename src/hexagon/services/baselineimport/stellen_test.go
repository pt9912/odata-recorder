package baselineimport_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// stelle ist eine der 16 Stellen des Imports, an denen ein Port oder ein Codec gerufen wird
// (SPEC-127 Nr. 17): die Operation, ihr n-ter Aufruf im Lauf, der Weg (neu, Wiederholung, Get),
// der Gegenstand der Abbildung und was nach einem Fehler noch geschehen darf.
type stelle struct {
	name       string
	op         string
	n          int
	weg        string
	gegenstand string
	nachher    []string
}

func stellen() []stelle {
	return []stelle{
		{"Imports.Get", "Imports.Get", 1, "neu", "Import", nil},
		{"Snapshots.Get", "Snapshots.Get", 1, "wdh", "Sonstiges", nil},
		{"SetMetadata (Wiederholung)", "Metadaten.SetMetadata", 1, "wdh", "Dienst", nil},
		{"Sessions.Get", "Sessions.Get", 1, "neu", "Session", nil},
		{"ListBySession", "Exchanges.ListBySession", 1, "neu", "Session", nil},
		{"Open ($metadata)", "Blobs.Open", 1, "neu", "Sonstiges", nil},
		{"ParseMetadata", "ParseMetadata", 1, "neu", "Sonstiges", []string{"Close"}},
		{"Open (Antwort)", "Blobs.Open", 2, "neu", "Sonstiges", nil},
		{"ParseCollection", "ParseCollection", 1, "neu", "Sonstiges", []string{"Close"}},
		{"ParseEntity", "ParseEntity", 1, "neu", "Sonstiges", []string{"Close"}},
		{"Begin", "Blobs.Begin", 1, "neu", "Sonstiges", nil},
		{"Write", "Write", 1, "neu", "Sonstiges", []string{"Abort"}},
		{"Commit", "Commit", 1, "neu", "Sonstiges", []string{"Abort"}},
		{"Publish", "Imports.Publish", 1, "neu", "Import", nil},
		{"SetMetadata (nach Publish)", "Metadaten.SetMetadata", 1, "neu", "Dienst", nil},
		{"Imports.Get (Get)", "Imports.Get", 1, "get", "Import", nil},
	}
}

// lauf führt den Weg der Stelle aus, lässt den Aufruf der Stelle mit fehler scheitern und liefert
// den Fehler des Aufrufs und die Ereignisse nach dem gescheiterten Aufruf.
func (s stelle) lauf(t *testing.T, fehler error) (*welt, []string, error) {
	t.Helper()
	w := neu(t)
	w.sessionR()
	w.start()
	ctx := context.Background()
	vor := 0
	if s.weg != "neu" {
		if _, err := w.imp.Import(ctx, anfrage()); err != nil {
			t.Fatalf("erster Aufruf: %v", err)
		}
		vor = len(w.p.folge())
	}
	n := s.n
	if s.weg != "neu" {
		n = w.p.anzahl(s.op) + 1
	}
	w.p.scheitere(s.op, n, fehler)
	var err error
	if s.weg == "get" {
		_, err = w.imp.Get(ctx, "imp-1")
	} else {
		_, err = w.imp.Import(ctx, anfrage())
	}
	folge := w.p.folge()[vor:]
	idx, gesehen := -1, 0
	for i, e := range folge {
		if e == s.op {
			gesehen++
			if s.weg == "neu" && gesehen == s.n || s.weg != "neu" && gesehen == 1 {
				idx = i
				break
			}
		}
	}
	if idx < 0 {
		t.Fatalf("%s: der Aufruf %s wurde nicht erreicht: %v", s.name, s.op, folge)
	}
	return w, folge[idx+1:], err
}

type ctxSchluessel struct{}

// TestZeileF1KontextUnveraendert: jede Operation mit Kontext erhält denselben Kontext (Menge 14 der
// Stellen), auf dem Weg neu, bei der Wiederholung und bei Get.
func TestZeileF1KontextUnveraendert(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxSchluessel{}, "dieser")
	w := neu(t)
	w.sessionR().entity("Orders", "OrderID=1,LineNo=2", produkt(1, "A", "1"))
	w.start()
	if _, err := w.imp.Import(ctx, anfrage()); err != nil {
		t.Fatal(err)
	}
	if _, err := w.imp.Import(ctx, anfrage()); err != nil {
		t.Fatal(err)
	}
	if _, err := w.imp.Get(ctx, "imp-1"); err != nil {
		t.Fatal(err)
	}
	ops := []string{"Imports.Get", "Snapshots.Get", "Metadaten.SetMetadata", "Sessions.Get", "Exchanges.ListBySession", "Blobs.Open", "ParseMetadata",
		"ParseCollection", "ParseEntity", "Blobs.Begin", "Imports.Publish"}
	for _, op := range ops {
		cs := w.p.ctxVon(op)
		if len(cs) == 0 {
			t.Errorf("%s: nie gerufen", op)
		}
		for i, c := range cs {
			if c != ctx {
				t.Errorf("%s Aufruf %d: anderer Kontext", op, i+1)
			}
		}
	}
	// Die Zählung der Stellen: Imports.Get dreimal (neu, Wiederholung, Get), SetMetadata zweimal, Open je Exchange.
	if w.p.anzahl("Imports.Get") != 3 || w.p.anzahl("Metadaten.SetMetadata") != 2 || w.p.anzahl("Blobs.Open") != 4 {
		t.Errorf("Zählung: Get %d, SetMetadata %d, Open %d", w.p.anzahl("Imports.Get"), w.p.anzahl("Metadaten.SetMetadata"), w.p.anzahl("Blobs.Open"))
	}
}

func kontextFehler() []error {
	c := context.Canceled
	return []error{c, context.DeadlineExceeded, errors.Join(c, outbound.ErrBlobFehlt), fmt.Errorf("%w", c)}
}

// TestZeileF2Kontextfehler: an jeder der 16 Stellen und für jede der vier Formen (Menge 64) kehrt
// derselbe Fehlerwert ohne Code zurück, danach geschieht nichts außer Close und Abort.
func TestZeileF2Kontextfehler(t *testing.T) {
	n := 0
	for _, s := range stellen() {
		for _, k := range kontextFehler() {
			n++
			_, nachher, err := s.lauf(t, k)
			if err != k {
				t.Errorf("%s mit %v: Fehler %v ist nicht derselbe Wert", s.name, k, err)
			}
			if !ohneCode(err) {
				t.Errorf("%s mit %v: Fehler mit Code", s.name, k)
			}
			if !slices.Equal(nachher, s.nachher) {
				t.Errorf("%s mit %v: danach %v, erwartet %v", s.name, k, nachher, s.nachher)
			}
		}
	}
	if n != 64 {
		t.Errorf("Menge %d, erwartet 64", n)
	}
}

// code, status und Abbildung nach SPEC-109 Nr. 5, von Hand aus der Tabelle der Spezifikation geschrieben.
func erwarteterCode(gegenstand, fehler string) (model.Code, int) {
	e9001 := func() (model.Code, int) { return model.CodeE9001, 500 }
	je := func(g map[string]model.Code, st map[model.Code]int) (model.Code, int) {
		c, ok := g[gegenstand]
		if !ok {
			return e9001()
		}
		return c, st[c]
	}
	switch fehler {
	case "ErrBlobFehlt", "ErrBlobBeschaedigt", "ErrBeschaedigt":
		return model.CodeE3001, 500
	case "ErrZuGross":
		return model.CodeE2401, 0
	case "ErrFremderService":
		return model.CodeE5004, 409
	case "ErrDienstUnbekannt":
		return model.CodeE5005, 404
	case "ErrNichtGefunden":
		return je(map[string]model.Code{"Session": model.CodeE7003, "Import": model.CodeE7005, "Dienst": model.CodeE5005},
			map[model.Code]int{model.CodeE7003: 404, model.CodeE7005: 404, model.CodeE5005: 404})
	case "ErrBereitsVorhanden":
		return je(map[string]model.Code{"Session": model.CodeE7010, "Import": model.CodeE6102},
			map[model.Code]int{model.CodeE7010: 409, model.CodeE6102: 409})
	case "ErrUngueltig":
		return je(map[string]model.Code{"Import": model.CodeE6101}, map[model.Code]int{model.CodeE6101: 422})
	}
	return e9001()
}

func portFehler() []struct {
	name string
	err  error
} {
	return []struct {
		name string
		err  error
	}{
		{"ErrNichtGefunden", outbound.ErrNichtGefunden}, {"ErrBereitsVorhanden", outbound.ErrBereitsVorhanden},
		{"ErrDienstUnbekannt", outbound.ErrDienstUnbekannt}, {"ErrBlobFehlt", outbound.ErrBlobFehlt},
		{"ErrBlobBeschaedigt", outbound.ErrBlobBeschaedigt}, {"ErrBeschaedigt", outbound.ErrBeschaedigt},
		{"ErrZuGross", outbound.ErrZuGross}, {"ErrUngueltig", model.ErrUngueltig},
		{"ErrFremderService", outbound.ErrFremderService}, {"fremd", errPlatte},
	}
}

// TestZeileF3FehlerEinesPorts: Stelle mal Fehler (Menge 160): der Code nach der Tabelle mit dem
// Gegenstand der Stelle, der Status des Codes, die Ursache erreichbar.
func TestZeileF3FehlerEinesPorts(t *testing.T) {
	n := 0
	for _, s := range stellen() {
		for _, f := range portFehler() {
			n++
			g := s.gegenstand
			if s.name == "Publish" && f.name == "ErrNichtGefunden" {
				g = "Session"
			}
			_, _, err := s.lauf(t, f.err)
			if s.name == "Imports.Get" && f.name == "ErrNichtGefunden" {
				if err != nil {
					t.Errorf("Imports.Get mal ErrNichtGefunden: %v, erwartet weiter", err)
				}
				continue
			}
			code, status := erwarteterCode(g, f.name)
			fe := fehlerMit(t, s.name+" mit "+f.name, err, code, status)
			if fe != nil && !errors.Is(err, f.err) {
				t.Errorf("%s mit %s: Ursache nicht erreichbar", s.name, f.name)
			}
		}
	}
	if n != 160 {
		t.Errorf("Menge %d, erwartet 160", n)
	}
}
