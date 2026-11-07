package inspection_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestKennungUnveraendert (X7): ListBySession wird genau einmal gerufen, mit der Kennung unverändert,
// sonst kein Port; eine unbekannte Session ist ODR-E7003, auch mit der leeren Kennung.
func TestKennungUnveraendert(t *testing.T) {
	ids := []string{"s1", "", " ", strings.Repeat("a", 64), "ses-" + string([]byte{0xc3, 0xa4})}
	for _, id := range ids {
		f := neuesFake()
		f.legeAn(id, 2)
		z := lies(t, neuerDienst(t, f), id, inbound.Seite{})
		if len(z.Eintraege) != 2 || !slices.Equal(f.kennungen(), []string{id}) || f.anzahl("ListBySession") != 1 || f.gesamt() != 1 {
			t.Errorf("%q: Kennungen %q, Aufrufe %d, Operationen %d", id, f.kennungen(), f.anzahl("ListBySession"), f.gesamt())
		}
		g := neuesFake()
		_, err := neuerDienst(t, g).Exchanges(context.Background(), id, inbound.Seite{})
		var fe *inbound.Fehler
		if !errors.As(err, &fe) || fe.Code != model.CodeE7003 || !slices.Equal(g.kennungen(), []string{id}) || g.gesamt() != 1 {
			t.Errorf("%q unbekannt: %v, Kennungen %q", id, err, g.kennungen())
		}
	}
}

// TestKontextfehlerUnveraendert (X8): ein Kontextfehler des Ports bleibt unverändert (gleicher Wert) und
// ohne Code, auch neben einem Fehler der Tabelle in jeder Reihenfolge.
func TestKontextfehlerUnveraendert(t *testing.T) {
	formen := map[string]error{
		"Canceled":         context.Canceled,
		"DeadlineExceeded": context.DeadlineExceeded,
		"verpackt":         fmt.Errorf("lesen: %w", context.Canceled),
		"Join vorn":        errors.Join(outbound.ErrNichtGefunden, context.Canceled),
		"Join hinten":      errors.Join(context.DeadlineExceeded, outbound.ErrNichtGefunden),
	}
	for name, e := range formen {
		f := neuesFake()
		f.legeAn(sid, 1)
		f.fehler = e
		_, err := neuerDienst(t, f).Exchanges(context.Background(), sid, inbound.Seite{})
		var fe *inbound.Fehler
		if err != e || errors.As(err, &fe) || f.anzahl("ListBySession") != 1 {
			t.Errorf("%s: Fehler %v (gleich: %v)", name, err, err == e)
		}
	}
}

// tabelleSession ist SPEC-109 Nr. 5 für den Gegenstand Session, abgeschrieben: Name wie in
// outbound.EintraegeAbbildung, der Fehler und der erwartete Code.
func tabelleSession() []struct {
	name string
	err  error
	code model.Code
} {
	return []struct {
		name string
		err  error
		code model.Code
	}{
		{"ErrBlobFehlt", outbound.ErrBlobFehlt, model.CodeE3001},
		{"ErrBlobBeschaedigt", outbound.ErrBlobBeschaedigt, model.CodeE3001},
		{"ErrBeschaedigt", outbound.ErrBeschaedigt, model.CodeE3001},
		{"ErrPfad", outbound.ErrPfad, model.CodeE2001},
		{"ErrZuGross", outbound.ErrZuGross, model.CodeE2401},
		{"ErrUpstreamNichtErreichbar", outbound.ErrUpstreamNichtErreichbar, model.CodeE4001},
		{"ErrUpstreamZeit", outbound.ErrUpstreamZeit, model.CodeE4001},
		{"ErrFremderService", outbound.ErrFremderService, model.CodeE5004},
		{"ErrDienstUnbekannt", outbound.ErrDienstUnbekannt, model.CodeE5005},
		{"ErrFalscheArt", outbound.ErrFalscheArt, model.CodeE7008},
		{"ErrReferenziert", outbound.ErrReferenziert, model.CodeE7004},
		{"ErrSessionGeschlossen", outbound.ErrSessionGeschlossen, model.CodeE5009},
		{"ErrUngueltigerUebergang", outbound.ErrUngueltigerUebergang, model.CodeE7011},
		{"ErrNichtGefunden", outbound.ErrNichtGefunden, model.CodeE7003},
		{"ErrBereitsVorhanden", outbound.ErrBereitsVorhanden, model.CodeE7010},
		{"ErrUngueltig", model.ErrUngueltig, model.CodeE9001},
		{"ErrMetadatenUngueltig", model.ErrMetadatenUngueltig, model.CodeE6004},
		{"ErrDatenUngueltig", model.ErrDatenUngueltig, model.CodeE6101},
		{"ErrURLUngueltig", model.ErrURLUngueltig, model.CodeE6005},
		{"ErrFilterNichtBeobachtet", model.ErrFilterNichtBeobachtet, model.CodeE6003},
		{"ErrURLNichtUnterstuetzt", model.ErrURLNichtUnterstuetzt, model.CodeE6001},
		{"ErrKompatibilitaet", model.ErrKompatibilitaet, model.CodeE6006},
		{"ErrEingabeUngueltig", model.ErrEingabeUngueltig, model.CodeE6007},
		{"ErrEingabeNichtUnterstuetzt", model.ErrEingabeNichtUnterstuetzt, model.CodeE6002},
		{"ErrUngueltigeSequenz", outbound.ErrUngueltigeSequenz, model.CodeE9001},
		{"ErrAbgeschlossen", outbound.ErrAbgeschlossen, model.CodeE9001},
		{"ErrRevisionVeraltet", outbound.ErrRevisionVeraltet, model.CodeE9001},
		{"anderer Fehler", errors.New("platte voll"), model.CodeE9001},
	}
}

// TestFehlerDesPortsMitCode (X9, Menge M-Fehler, Aufrufstelle ListBySession): jeder Fehler der Tabelle
// wird mit dem Gegenstand Session abgebildet: Code, der eine Status, Ursache erreichbar. Die Tabelle
// des Tests hat genau die Namen von outbound.EintraegeAbbildung() (und den anderen Fehler).
func TestFehlerDesPortsMitCode(t *testing.T) {
	var namen []string
	for _, z := range tabelleSession() {
		namen = append(namen, z.name)
	}
	quelle := append(outbound.EintraegeAbbildung(), "anderer Fehler")
	slices.Sort(namen)
	slices.Sort(quelle)
	if !slices.Equal(namen, quelle) {
		t.Fatalf("Namen der Tabelle des Tests %v, der Quelle %v", namen, quelle)
	}
	for _, z := range tabelleSession() {
		f := neuesFake()
		f.legeAn(sid, 1)
		f.fehler = fmt.Errorf("lesen: %w", z.err)
		_, err := neuerDienst(t, f).Exchanges(context.Background(), sid, inbound.Seite{})
		var fe *inbound.Fehler
		if !errors.As(err, &fe) {
			t.Errorf("%s: kein *Fehler: %v", z.name, err)
			continue
		}
		if fe.Code != z.code || len(z.code.Stati()) == 0 || fe.Status != z.code.Stati()[0] || !errors.Is(err, z.err) || !errors.Is(fe.Ursache, z.err) {
			t.Errorf("%s: Code %v Status %d Ursache %v, erwartet %v", z.name, fe.Code, fe.Status, fe.Ursache, z.code)
		}
	}
}

type ctxSchluessel struct{}

// TestKontextDesAufrufs (X10, Menge M-Fehler, Achse ctx-Zustand): der ctx geht unverändert an den Port;
// ein vor dem Aufruf beendeter Kontext endet mit seinem Fehler; einer, den der Port nach dem Lesen
// beendet, ändert das Ergebnis nicht.
func TestKontextDesAufrufs(t *testing.T) {
	f := neuesFake()
	f.legeAn(sid, 2)
	ctx := context.WithValue(context.Background(), ctxSchluessel{}, "mein ctx")
	if _, err := neuerDienst(t, f).Exchanges(ctx, sid, inbound.Seite{}); err != nil || len(f.kontexte()) != 1 || f.kontexte()[0] != ctx {
		t.Errorf("lebender ctx: Fehler %v, der Port erhielt %v", err, f.kontexte())
	}

	g := neuesFake()
	g.legeAn(sid, 2)
	abgebrochen, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := neuerDienst(t, g).Exchanges(abgebrochen, sid, inbound.Seite{})
	var fe *inbound.Fehler
	if err != context.Canceled || errors.As(err, &fe) || g.anzahl("ListBySession") != 1 {
		t.Errorf("vor dem Aufruf abgebrochen: %v", err)
	}

	h := neuesFake()
	h.legeAn(sid, 2)
	spaeter, cancel2 := context.WithCancel(context.Background())
	h.nachLesen = cancel2
	z, err := neuerDienst(t, h).Exchanges(spaeter, sid, inbound.Seite{})
	if err != nil || len(z.Eintraege) != 2 || spaeter.Err() == nil {
		t.Errorf("nach der Rückkehr beendet: Fehler %v, %d Einträge, Kontext beendet: %v", err, len(z.Eintraege), spaeter.Err() != nil)
	}
}

// TestLesenSchreibtNichts (X13): Exchanges ruft von RecordingRepository nur ListBySession, bei Erfolg,
// Fehler des Ports, Parameterfehler und ctx nil; der Bestand bleibt gleich.
func TestLesenSchreibtNichts(t *testing.T) {
	faelle := []struct {
		name     string
		ohneCtx  bool
		seite    inbound.Seite
		fehler   error
		listeGer int
	}{
		{"Erfolg", false, inbound.Seite{}, nil, 1},
		{"Fehler des Ports", false, inbound.Seite{}, errors.New("platte voll"), 1},
		{"Parameterfehler", false, inbound.Seite{Nach: -1}, nil, 0},
		{"ctx nil", true, inbound.Seite{}, nil, 0},
	}
	for _, c := range faelle {
		f := neuesFake()
		f.legeAn(sid, 3)
		f.fehler = c.fehler
		davor := f.stand()
		ctx := context.Background()
		if c.ohneCtx {
			ctx = nilCtx()
		}
		_, _ = neuerDienst(t, f).Exchanges(ctx, sid, c.seite)
		if f.anzahl("Append") != 0 || f.anzahl("Get") != 0 || f.anzahl("ListBySession") != c.listeGer || !reflect.DeepEqual(davor, f.stand()) {
			t.Errorf("%s: Append %d Get %d ListBySession %d", c.name, f.anzahl("Append"), f.anzahl("Get"), f.anzahl("ListBySession"))
		}
	}
}
