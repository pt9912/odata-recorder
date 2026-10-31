package replay_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// pfadPassword ist das Pfadmuster der Fixtures.
func pfadPassword(t *testing.T) []model.JSONPfad {
	t.Helper()
	p, err := model.ParseJSONPfad("$..password")
	if err != nil {
		t.Fatal(err)
	}
	return []model.JSONPfad{p}
}

// redigiert ist das Orakel für den Key: der Redaktor der Domain mit großer Grenze, unabhängig vom Dienst.
func redigiert(t *testing.T, body []byte, pfade []model.JSONPfad) ([]byte, model.RedactionStatus) {
	t.Helper()
	var aus bytes.Buffer
	st, err := model.RedigiereBody(&aus, bytes.NewReader(body), model.BodyEingang{ContentType: []string{"application/json"}},
		model.BodyRegeln{Pfade: pfade, MaxBytes: 1 << 20})
	if err != nil {
		t.Fatalf("Orakel: %v", err)
	}
	if aus.Len() == 0 {
		return nil, st
	}
	return aus.Bytes(), st
}

// keyMitBody ist der Key einer Aufzeichnung mit dem Body bytes und dem Status rr, unabhängig vom Dienst.
func (u *umg) keyMitBody(a inbound.Anfrage, bytes []byte, rr model.RedactionStatus) string {
	u.t.Helper()
	s, err := model.BildeSchluessel(model.SchluesselEingabe{
		Method: a.Methode, ServiceID: svcID, Path: "/Products", RawQuery: a.RawQuery, Header: a.Header, Body: bytes, RequestRedaction: rr,
	}, u.k.Regeln)
	if err != nil {
		u.t.Fatalf("BildeSchluessel: %v", err)
	}
	return s.MatchKey
}

// mitBodyAnfrage ist der Standard-Request mit Body und Content-Type ct (leer: keiner).
func mitBodyAnfrage(body []byte, ct ...string) inbound.Anfrage {
	a := anf()
	a.Methode = "POST"
	a.Body = &zaehlLeser{daten: bytes.Clone(body)}
	a.Laenge = int64(len(body))
	if len(ct) > 0 {
		a.Header["Content-Type"] = ct
	}
	return a
}

func wiederhole(s string, n int) string { return strings.Repeat(s, n) }

// TestRequestBodyFormen hält Z17: Menge 16; jede Form der Eingabe wird entschieden, der Key entsteht aus der Ausgabe.
func TestRequestBodyFormen(t *testing.T) {
	js := []string{"application/json"}
	passwort := `{"password":"x"}`
	ausgabe25 := `{"password":"` + model.Marker + `"}`
	if len(passwort) != 16 || len(ausgabe25) != 25 {
		t.Fatalf("Fixtures: %d und %d Bytes", len(passwort), len(ausgabe25))
	}
	type fall struct {
		name  string
		body  []byte
		ct    []string
		pfad  string
		max   int64
		laeng int64
		gut   bool
	}
	faelle := []fall{
		{"leer gelesen", nil, js, "", 64, -1, true},
		{"JSON mit Geheimnis", []byte(`{"a":1,"password":"s1"}`), js, "", 1 << 10, 22, true},
		{"JSON ohne Treffer", []byte(`{"a":1}`), js, "", 1 << 10, 7, true},
		{"Content-Type fehlt", []byte(`{}`), nil, "", 64, 2, false},
		{"text/plain", []byte(`{}`), []string{"text/plain"}, "", 64, 2, false},
		{"zwei Content-Type", []byte(`{}`), []string{"application/json", "application/json"}, "", 64, 2, false},
		{"application/xml bei $metadata", []byte(`<edmx/>`), []string{"application/xml"}, "/$metadata", 64, 7, false},
		{"Syntaxfehler", []byte(`{"a":`), js, "", 64, 5, false},
		{"ungültiges UTF-8", append(append([]byte(`{"a":"`), 0xFF), []byte(`"}`)...), js, "", 64, 10, false},
		{"Tiefe 513", []byte(wiederhole("[", 513) + wiederhole("]", 513)), js, "", 4096, 1026, false},
		{"Tiefe 512", []byte(wiederhole("[", 512) + wiederhole("]", 512)), js, "", 4096, 1024, true},
		{"Eingabe MaxBytes+1", []byte(`[` + wiederhole("1,", 15) + `1]`), js, "", 32, 33, false},
		{"Eingabe genau MaxBytes", []byte(`[` + wiederhole("1,", 14) + `1]`), js, "", 31, 31, true},
		{"Ausgabe MaxBytes+1", []byte(passwort), js, "", 24, 16, false},
		{"Ausgabe genau MaxBytes", []byte(passwort), js, "", 25, 16, true},
		{"Zahl unter treffendem Namen", []byte(`{"password":1}`), js, "", 64, 14, false},
	}
	if len(faelle) != 16 {
		t.Fatalf("%d Fälle", len(faelle))
	}
	for _, f := range faelle {
		u := neu(t)
		u.k.MaxBytes = f.max
		u.k.Pfade = pfadPassword(t)
		a := mitBodyAnfrage(f.body, f.ct...)
		a.Laenge = f.laeng
		if f.laeng == -1 {
			a.Body = &zaehlLeser{}
		}
		if f.pfad != "" {
			a.Pfad = root + f.pfad
		}
		if f.gut {
			rb, rr := redigiert(t, f.body, u.k.Pfade)
			if f.name == "Ausgabe genau MaxBytes" && string(rb) != ausgabe25 {
				t.Fatalf("Orakel: %q", rb)
			}
			u.lege(u.austausch(1, u.keyMitBody(a, rb, rr), true))
		}
		resp, err := u.replay(context.Background(), a)
		if f.gut {
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
				continue
			}
			_ = resp.Body.Close()
			continue
		}
		erwarteCode(t, f.name, err, model.CodeE5001)
		if u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 || u.p.anzahl("Exchanges.ListBySession") != 0 {
			t.Errorf("%s: Port nach dem Fehler gerufen", f.name)
		}
		if _, da := u.cursorStand(); da {
			t.Errorf("%s: Cursor", f.name)
		}
	}
}

// TestContentEncodingRequest hält Z18: Menge 22 (11 Werte × Body vorhanden oder nil); nur identity-Teile, sonst
// ODR-E5001 ohne ein gelesenes Byte; ohne Body hat der Header keine Wirkung.
func TestContentEncodingRequest(t *testing.T) {
	type wert struct {
		name  string
		setze func(h map[string][]string)
		bad   bool
	}
	setz := func(w ...string) func(map[string][]string) {
		return func(h map[string][]string) { h["Content-Encoding"] = w }
	}
	werte := []wert{
		{"fehlt", func(map[string][]string) {}, false}, {"leere Liste", setz(), false}, {"identity", setz("identity"), false},
		{"IDENTITY", setz("IDENTITY"), false}, {"identity, identity", setz("identity, identity"), false},
		{"mit Rand", setz(" identity "), false}, {"gzip", setz("gzip"), true}, {"deflate", setz("deflate"), true},
		{"br", setz("br"), true}, {"leerer Wert", setz(""), true}, {"identity,", setz("identity,"), true},
	}
	n := 0
	for _, w := range werte {
		for _, mitBody := range []bool{true, false} {
			n++
			u := neu(t)
			u.k.Pfade = pfadPassword(t)
			body := []byte(`{"a":1}`)
			var a inbound.Anfrage
			var rd *zaehlLeser
			if mitBody {
				a = mitBodyAnfrage(body, "application/json")
				rd = a.Body.(*zaehlLeser)
			} else {
				a = anf()
			}
			w.setze(a.Header)
			var key string
			if mitBody {
				rb, rr := redigiert(t, body, u.k.Pfade)
				key = u.keyMitBody(a, rb, rr)
			} else {
				key = u.keyMitBody(a, nil, model.RedactionNone)
			}
			u.lege(u.austausch(1, key, true))
			resp, err := u.replay(context.Background(), a)
			if w.bad && mitBody {
				erwarteCode(t, w.name+" mit Body", err, model.CodeE5001)
				if rd.gelesen != 0 {
					t.Errorf("%s: %d Bytes gelesen", w.name, rd.gelesen)
				}
				continue
			}
			if err != nil {
				t.Errorf("%s (Body %v): %v", w.name, mitBody, err)
				continue
			}
			_ = resp.Body.Close()
		}
	}
	if n != 22 {
		t.Errorf("Menge: %d, erwartet 22", n)
	}
}

// TestKeyAusAusgabe hält Z19: Menge 10 (2 Modi × 5 Paare); ein anderes Geheimnis trifft, eine andere Reihenfolge
// des Schlüssels oder des Leerraums nur bei json-canonical.
func TestKeyAusAusgabe(t *testing.T) {
	basis := `{"a":1,"password":"s1"}`
	paare := []struct {
		name   string
		replay string
		bytes  bool
		canon  bool
	}{
		{"gleich", basis, true, true},
		{"Reihenfolge", `{"password":"s1","a":1}`, false, true},
		{"Leerraum", `{"a": 1, "password": "s1"}`, false, true},
		{"Geheimnis anders", `{"a":1,"password":"s2"}`, true, true},
		{"anderer Inhalt", `{"a":2,"password":"s1"}`, false, false},
	}
	n := 0
	for _, modus := range []model.BodyModus{model.BodyBytes, model.BodyJSONCanonical} {
		for _, p := range paare {
			n++
			u := neu(t)
			u.k.Regeln.BodyMode = modus
			u.k.Pfade = pfadPassword(t)
			rb, rr := redigiert(t, []byte(basis), u.k.Pfade)
			u.lege(u.austausch(1, u.keyMitBody(mitBodyAnfrage([]byte(basis), "application/json"), rb, rr), true))
			resp, err := u.replay(context.Background(), mitBodyAnfrage([]byte(p.replay), "application/json"))
			want := p.bytes
			if modus == model.BodyJSONCanonical {
				want = p.canon
			}
			was := fmt.Sprintf("%s %s", modus, p.name)
			if want && err != nil {
				t.Errorf("%s: kein Treffer: %v", was, err)
			} else if !want {
				erwarteCode(t, was, err, model.CodeE5001)
			}
			if err == nil {
				_ = resp.Body.Close()
			}
		}
	}
	if n != 10 {
		t.Errorf("Menge: %d, erwartet 10", n)
	}
}

// fehlerLeser liefert daten und den Fehler zu einem Zeitpunkt.
type fehlerLeser struct {
	daten  []byte
	fehler error
	wann   string
	stufe  int
}

func (f *fehlerLeser) Read(p []byte) (int, error) {
	f.stufe++
	switch {
	case f.wann == "vor dem ersten Byte":
		return 0, f.fehler
	case f.wann == "mit Daten zusammen" && f.stufe == 1:
		return copy(p, f.daten), f.fehler
	case f.wann == "nach Daten" && f.stufe == 1:
		return copy(p, f.daten), nil
	}
	return 0, f.fehler
}

// TestLesefehlerBody hält Z23: Menge 15 (5 Fehler × 3 Zeitpunkte); 9 Kontextfehler ohne Code, 6 ODR-E5001.
func TestLesefehlerBody(t *testing.T) {
	fehler := []struct {
		name string
		err  error
		ctx  error
	}{
		{"Canceled", context.Canceled, context.Canceled},
		{"DeadlineExceeded", context.DeadlineExceeded, context.DeadlineExceeded},
		{"Join(Canceled, ErrUnexpectedEOF)", errors.Join(context.Canceled, io.ErrUnexpectedEOF), context.Canceled},
		{"ErrUnexpectedEOF", io.ErrUnexpectedEOF, nil},
		{"eigener", errors.New("eigener"), nil},
	}
	var ctxN, codeN int
	for _, f := range fehler {
		for _, wann := range []string{"vor dem ersten Byte", "nach Daten", "mit Daten zusammen"} {
			u := neu(t)
			a := mitBodyAnfrage(nil, "application/json")
			a.Body = &fehlerLeser{daten: []byte(`{"a":`), fehler: f.err, wann: wann}
			a.Laenge = -1
			_, err := u.replay(context.Background(), a)
			if f.ctx != nil {
				ctxN++
				erwarteOhneCode(t, f.name+" "+wann, err, f.ctx)
			} else {
				codeN++
				erwarteCode(t, f.name+" "+wann, err, model.CodeE5001)
			}
			if u.p.anzahl("Exchanges.ListBySession") != 0 || u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
				t.Errorf("%s %s: Port nach dem Lesefehler", f.name, wann)
			}
		}
	}
	if ctxN != 9 || codeN != 6 {
		t.Errorf("Menge: %d Kontextfehler, %d mit Code; erwartet 9 und 6", ctxN, codeN)
	}
}
