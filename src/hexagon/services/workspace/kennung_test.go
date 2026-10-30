package workspace_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// formDerID ist das Orakel für Z13: die Form einer gelieferten Kennung nach SPEC-129 Nr. 3, unabhängig vom
// Dienst als regulärer Ausdruck über Bytes geschrieben.
var formDerID = regexp.MustCompile(`^[0-9a-z-]{1,64}$`)

// idFaelle erzeugt die Kennungen der Klasse: Längen mal Zeichen, jedes Byte 0 bis 255 an erster, mittlerer
// und letzter Stelle.
func idFaelle() []string {
	var r []string
	for _, n := range []int{1, 2, 63, 64, 65, 200} {
		r = append(r, strings.Repeat("a", n), strings.Repeat("-", n), strings.Repeat("7", n))
	}
	for b := 0; b < 256; b++ {
		for _, pos := range []int{0, 2, 4} {
			s := []byte("abcde")
			s[pos] = byte(b)
			r = append(r, string(s))
		}
	}
	return r
}

// TestFormDerID (Z13): eine id ist gültig genau mit 1 bis 64 Bytes aus 0-9, a-z, -; jede andere ist
// ErrUngueltig ohne Code, ohne Port und ohne den Wert im Text.
func TestFormDerID(t *testing.T) {
	faelle := idFaelle()
	if len(faelle) != 18+768 {
		t.Fatalf("Fälle %d", len(faelle))
	}
	gueltigN, ungueltigN := 0, 0
	for _, id := range faelle {
		w := neu(t)
		ws, err := w.v.Create(context.Background(), "bsl-1", id)
		if formDerID.MatchString(id) {
			gueltigN++
			if err != nil || ws.ID != id {
				t.Errorf("id %q: %v", id, err)
			}
			continue
		}
		ungueltigN++
		pruefeOhneCode(t, err, model.ErrUngueltig)
		if w.p.gesamt() != 0 {
			t.Errorf("id %q: Aufrufe %v", id, w.p.ops())
		}
		if err != nil && strings.Contains(err.Error(), id) {
			t.Errorf("id %q steht im Text", id)
		}
	}
	// Zählung der Klasse: gültig sind 18-4 Wiederholungen bis 64 und die Bytes 0-9, a-z, - an drei Stellen.
	if gueltigN != 4*3+3*(10+26+1) || ungueltigN != len(faelle)-gueltigN {
		t.Errorf("gültig %d, ungültig %d", gueltigN, ungueltigN)
	}
}

// TestLeereIDVergibt (Z14): eine leere id bezieht IDs.WorkspaceID genau einmal, eine gegebene verbraucht
// keine Kennung; die Server-Kennung erfüllt die Form (36 Bytes).
func TestLeereIDVergibt(t *testing.T) {
	w := neu(t)
	ws, err := w.v.Create(context.Background(), "bsl-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if ws.ID != "wsp-00000000000000000000000000000001" || len(ws.ID) != 36 || !formDerID.MatchString(ws.ID) {
		t.Errorf("Kennung %q", ws.ID)
	}
	zaehler(t, w, map[string]int{"IDs.WorkspaceID": 1})
	ws2, err := w.v.Create(context.Background(), "bsl-1", "")
	if err != nil || ws2.ID != "wsp-00000000000000000000000000000002" {
		t.Errorf("zweite Kennung %q: %v", ws2.ID, err)
	}
	ws3, err := w.v.Create(context.Background(), "bsl-1", "eigene-id")
	if err != nil || ws3.ID != "eigene-id" {
		t.Errorf("gegebene Kennung %q: %v", ws3.ID, err)
	}
	zaehler(t, w, map[string]int{"IDs.WorkspaceID": 2})
	if _, ok := w.st.stand().ws["wsp-00000000000000000000000000000001"]; !ok {
		t.Error("Workspace unter der vergebenen Kennung fehlt")
	}
	// Eine Client-Kennung in der Form der Server-Kennung trifft eine vergebene: ODR-E7007.
	_, err = w.v.Create(context.Background(), "bsl-1", "wsp-00000000000000000000000000000001")
	pruefeCode(t, err, model.CodeE7007, outbound.ErrBereitsVorhanden)
}

// TestBaselineIDLeerUndFormen (Z15): baselineID leer ist ErrUngueltig; jede nichtleere Form geht unverändert
// an Snapshots.Get.
func TestBaselineIDLeerUndFormen(t *testing.T) {
	w := neu(t)
	_, err := w.v.Create(context.Background(), "", "ws-neu")
	pruefeOhneCode(t, err, model.ErrUngueltig)
	if w.p.gesamt() != 0 {
		t.Errorf("Aufrufe %v", w.p.ops())
	}
	for _, id := range append(idFaelle(), "Bsl-1", "bsl 1", "snp-ws-1") {
		if id == "" {
			continue
		}
		w := neu(t)
		_, err := w.v.Create(context.Background(), id, "ws-neu")
		if id == "bsl-1" {
			continue
		}
		pruefeCode(t, err, model.CodeE7002, outbound.ErrNichtGefunden)
		if a := w.p.von("Snapshots.Get"); len(a) != 1 || a[0].args[0] != id {
			t.Errorf("baselineID %q: Get %v", id, a)
		}
	}
}
