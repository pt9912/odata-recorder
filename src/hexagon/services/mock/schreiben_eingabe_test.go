package mock_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// dreiMethoden sind die Methoden, deren Body der Dienst liest (SPEC-192 Nr. 5).
func dreiMethoden() []string { return []string{"POST", "PATCH", "PUT"} }

// schluesselFuer ist die Kennung des Eintrags, den die Methode trifft: bei POST ein neuer, sonst der vorhandene.
func schluesselFuer(m string) string {
	if m == "POST" {
		return "7"
	}
	return "2"
}

// leseAufruf ist der eine Aufruf des Lesers, den ein Lauf festhielt: ctx und die sechs Argumente nach dem ctx.
func leseAufruf(t *testing.T, u *umg) ereignis {
	t.Helper()
	ev := u.p.von("Eingabe.LeseEntity")
	if len(ev) != 1 {
		t.Fatalf("%d Aufrufe des Lesers, erwartet 1", len(ev))
	}
	return ev[0]
}

// TestEingabeZusagen (W07 bis W12, M6): Zusage a bis f der Eingabe × Aufrufstelle {POST, PATCH, PUT}; der Fake hält jedes
// Argument fest.
func TestEingabeZusagen(t *testing.T) {
	n := 0
	for _, m := range dreiMethoden() {
		id := schluesselFuer(m)
		// (a) ctx des Aufrufs, (e) art, (d) Set, (f) grenze.
		for _, grenze := range []int64{1, 8388608, 1073741824} {
			for _, set := range []string{setProducts, "Kunden"} {
				u := schreibUmg(t)
				u.k.MaxBytes = grenze
				if m == "POST" {
					u.ressource(model.RessourceMenge, set, model.Optionen{})
				} else {
					u.ressource(model.RessourceEntity, set, model.Optionen{}, zahl("", id))
				}
				u.ei.ergebnis = eingabeIDName(id, "neu")
				_, _ = u.rufe(anfrage(m, dienstRoot+"/x"))
				ev := leseAufruf(t, u)
				if ev.ctx != u.ctx || ev.ctx.Value(schluesselKtx{}) != "ctx-des-aufrufs" {
					t.Errorf("%s: ctx nicht der des Aufrufs", m)
				}
				if ev.args[3] != set || ev.args[4] != model.EingabeArt(m) || ev.args[5] != grenze {
					t.Errorf("%s %s: Set %v art %v grenze %v, erwartet %s %s %d", m, set, ev.args[3], ev.args[4], ev.args[5], set, m, grenze)
				}
				if ev.args[3] == "N.Product" {
					t.Errorf("%s: Name des Typs statt des Sets", m)
				}
				n++
			}
		}
		// (b) src: der Body der Anfrage, bei nil ein Strom ohne Byte; der Dienst liest und schließt ihn nie.
		for _, body := range []string{"mit Bytes", "nil", "leer"} {
			u := schreibUmg(t)
			u.setzeMutation(m, id, eingabeIDName(id, "neu"))
			a := anfrage(m, dienstRoot+"/x")
			z := &zaehlStrom{inhalt: "{\"Id\":1}"}
			switch body {
			case "mit Bytes":
				a.Body, a.Laenge = z, int64(len(z.inhalt))
			case "leer":
				z.inhalt = ""
				a.Body, a.Laenge = z, 0
			}
			_, _ = u.rufe(a)
			ev := leseAufruf(t, u)
			src, ok := ev.args[0].(io.Reader)
			if !ok || src == nil {
				t.Fatalf("%s %s: src ist nil", m, body)
			}
			if body != "nil" && src != io.Reader(z) {
				t.Errorf("%s %s: src ist nicht der Body der Anfrage", m, body)
			}
			if body == "nil" {
				if k, err := src.Read(make([]byte, 4)); k != 0 || err != io.EOF {
					t.Errorf("%s nil: src liefert %d Bytes, %v", m, k, err)
				}
			}
			if z.gelesen != 0 || z.geschlossen != 0 {
				t.Errorf("%s %s: gelesen %d geschlossen %d, erwartet 0", m, body, z.gelesen, z.geschlossen)
			}
			n++
		}
		// (c) contentType: alle Kopfzeilen, Namen in der Bytefolge, dann in ihrer Reihenfolge; ohne Zeile nil.
		formen := []struct {
			name   string
			header map[string][]string
			soll   []string
		}{
			{"keine Zeile", nil, nil},
			{"eine", map[string][]string{"Content-Type": {"application/json"}}, []string{"application/json"}},
			{"zwei", map[string][]string{"Content-Type": {"a/1", "b/2"}}, []string{"a/1", "b/2"}},
			{"zwei Schreibweisen", map[string][]string{"content-type": {"c/3"}, "Content-Type": {"a/1", "b/2"}}, []string{"a/1", "b/2", "c/3"}},
			{"leerer Wert", map[string][]string{"Content-Type": {""}}, []string{""}},
			{"mit Parametern", map[string][]string{"CONTENT-TYPE": {"application/json; charset=utf-8; x=\"y\""}}, []string{"application/json; charset=utf-8; x=\"y\""}},
		}
		for _, f := range formen {
			u := schreibUmg(t)
			u.setzeMutation(m, id, eingabeIDName(id, "neu"))
			a := anfrage(m, dienstRoot+"/x")
			for k, v := range f.header {
				a.Header[k] = v
			}
			_, _ = u.rufe(a)
			ct := leseAufruf(t, u).args[1].([]string)
			if (f.soll == nil) != (ct == nil) || !slices.Equal(ct, f.soll) {
				t.Errorf("%s %s: contentType %#v, erwartet %#v", m, f.name, ct, f.soll)
			}
			n++
		}
	}
	if n != 3*(6+3+6) {
		t.Errorf("Menge %d, erwartet %d", n, 3*15)
	}
}

// ursacheCode ist das Orakel zu M7 aus SPEC-180 Nr. 3 und SPEC-109 Nr. 5: die ersten fünfzehn Ursachen sind ODR-E6007 mit 400,
// die letzten vier ODR-E6002 mit 501.
func ursacheCode(i int) (model.Code, int) {
	if i < 15 {
		return e6007, 400
	}
	return e6002, 501
}

// TestEingabeFehler (W13, W14, M7): Ursache 19 × Methode 3 = 57; der Fehler des Lesers wird nach SPEC-109 Nr. 5 abgebildet, die
// Ursache bleibt über Unwrap erreichbar, und bei jedem Fehler bleibt der Zustand unverändert (Transaktion 0, Put 0, Advance 0).
func TestEingabeFehler(t *testing.T) {
	alle := model.AlleEingabeUrsachen()
	if len(alle) != 19 {
		t.Fatalf("%d Ursachen, erwartet 19", len(alle))
	}
	n := 0
	for _, m := range dreiMethoden() {
		for i, ur := range alle {
			n++
			u := schreibUmg(t)
			vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
			quelle := &model.EingabeFehler{Ursache: ur, Ort: "/x"}
			u.p.setzeFehler("Eingabe.LeseEntity", quelle)
			_, err := u.mutiere(m, schluesselFuer(m), eingabeIDName(schluesselFuer(m), "neu"))
			code, status := ursacheCode(i)
			f := erwarteFehler(t, fmt.Sprintf("%s %s", m, ur), err, code)
			if f != nil {
				var ziel *model.EingabeFehler
				if f.Status != status || !errors.As(f, &ziel) || ziel != quelle || !errors.Is(f, quelle) {
					t.Errorf("%s %s: Status %d, Ursache nicht über Unwrap erreichbar", m, ur, f.Status)
				}
			}
			if u.p.anzahl("Eingabe.LeseEntity") != 1 || u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
				t.Errorf("%s %s: Leser %d Transaktion %d", m, ur, u.p.anzahl("Eingabe.LeseEntity"), u.p.anzahl("Transaktion.WithWorkspaceTransaction"))
			}
			u.ohneSchreiben(m + " " + string(ur))
			u.unveraendert(vorher, rev, m+" "+string(ur))
		}
		// model.ErrUngueltig ist ODR-E9001 ohne Code der Eingabe; ein Kontextfehler bleibt ohne Code, nackt, verpackt, im Join.
		for _, e := range []error{model.ErrUngueltig, fmt.Errorf("lesen: %w", model.ErrUngueltig)} {
			u := schreibUmg(t)
			u.p.setzeFehler("Eingabe.LeseEntity", e)
			_, err := u.mutiere(m, schluesselFuer(m), eingabeIDName(schluesselFuer(m), "neu"))
			pruefeFehler(t, m+" ErrUngueltig", err, e9001)
		}
		for _, e := range []error{context.Canceled, fmt.Errorf("lesen: %w", context.Canceled), errors.Join(&model.EingabeFehler{Ursache: model.EingabeSyntax}, context.Canceled), context.DeadlineExceeded} {
			u := schreibUmg(t)
			u.p.setzeFehler("Eingabe.LeseEntity", e)
			_, err := u.mutiere(m, schluesselFuer(m), eingabeIDName(schluesselFuer(m), "neu"))
			if err != e || fehlercode(err) != 0 || u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
				t.Errorf("%s: Kontextfehler %v kehrt als %v mit Code %v zurück", m, e, err, fehlercode(err))
			}
		}
	}
	if n != 57 {
		t.Errorf("Menge %d, erwartet 57", n)
	}
	// Die vier Ursachen der Klasse nicht_unterstuetzt (W49): ODR-E6002 mit 501.
	if k := len(alle[15:]); k != 4 {
		t.Errorf("%d Ursachen der Klasse nicht_unterstuetzt, erwartet 4", k)
	}
	for _, ur := range alle[15:] {
		if !ur.NichtUnterstuetzt() {
			t.Errorf("%s nicht in der Klasse nicht_unterstuetzt", ur)
		}
	}
}

// paar ist ein Eintrag von Pairs (zwei Schlüssel A und B vom Typ Edm.String) in der Kanonform.
func paar(a, b string) model.EntityState {
	return model.EntityState{
		WorkspaceID: wsID, EntitySet: "Pairs", EntityKey: "A='" + a + "',B='" + b + "'", Completeness: model.EntityComplete, Version: 1,
		Payload: `{"A":"` + a + `","B":"` + b + `","Name":null}`, KnownProperties: []string{"A", "B", "Name"},
	}
}

// paarEingabe ist der Body für Pairs mit den Schlüsseln a und b und dem Namen.
func paarEingabe(a, b, name string) model.EntityEingabe {
	return model.EntityEingabe{
		Typ: "N.Pair", Schluessel: "A='" + a + "',B='" + b + "'",
		Properties: []model.PropertyWert{pw("A", model.ArtString, a), pw("B", model.ArtString, b), pw("Name", model.ArtString, name)},
	}
}

// paarURL setzt die Ressource auf einen Eintrag von Pairs mit den Schlüsseltexten (Literale der URL).
func (u *umg) paarURL(a, b string) {
	str := func(n, s string) model.SchluesselTeil {
		return model.SchluesselTeil{Name: n, Literal: model.Literal{Art: model.LiteralZeichenkette, Text: s}}
	}
	u.ressource(model.RessourceEntity, "Pairs", model.Optionen{}, str("A", a), str("B", b))
}

// TestSchluesselWiderspruch (W19, M9): nennt der Body eine Schlüssel-Property, hat sie den Text des Teils der URL, sonst
// ODR-E6007 vor der Transaktion. Schlüssel 2 Formen (Products.Id; Pairs.A und Pairs.B) × Property × Art 2 × Abweichung 2 = 12
// Widersprüche, dazu die Sollfälle (gleich; PATCH ohne Schlüssel) und die Paare mit einer früheren Stufe.
func TestSchluesselWiderspruch(t *testing.T) {
	n := 0
	for _, m := range []string{"PATCH", "PUT"} {
		// Products.Id: URL 2, Body anderer Wert und derselbe Wert in anderer Schreibweise.
		for _, abw := range []string{"3", "02"} {
			n++
			u := schreibUmg(t)
			_, err := u.mutiere(m, "2", eingabeIDName(abw, "neu"))
			pruefeFehler(t, m+" Id "+abw, err, e6007)
			if u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
				t.Errorf("%s Id %s: Transaktion gerufen", m, abw)
			}
		}
		// Pairs: die erste und die zweite Schlüssel-Property widersprechen.
		for _, stelle := range []string{"A", "B"} {
			for _, abw := range []string{"z", "X"} {
				n++
				u := schreibUmg(t)
				u.en.setze("Pairs", paar("x", "y"))
				u.paarURL("'x'", "'y'")
				e := paarEingabe("x", "y", "neu")
				for i := range e.Properties {
					if e.Properties[i].Name == stelle {
						e.Properties[i].Wert.Text = abw
					}
				}
				u.ei.ergebnis = e
				_, err := u.rufe(anfrage(m, dienstRoot+"/x"))
				pruefeFehler(t, fmt.Sprintf("%s Pairs.%s %s", m, stelle, abw), err, e6007)
				if u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
					t.Errorf("%s Pairs.%s: Transaktion gerufen", m, stelle)
				}
			}
		}
		// Sollfälle: gleich.
		u := schreibUmg(t)
		a, err := u.mutiere(m, "2", eingabeIDName("2", "neu"))
		erwarteStatus(t, m+" gleich", a, err, 204)
		u = schreibUmg(t)
		u.en.setze("Pairs", paar("x", "y"))
		u.paarURL("'x'", "'y'")
		u.ei.ergebnis = paarEingabe("x", "y", "neu")
		a, err = u.rufe(anfrage(m, dienstRoot+"/x"))
		erwarteStatus(t, m+" Pairs gleich", a, err, 204)
		n += 2
	}
	// PATCH ohne Schlüssel im Body widerspricht nicht (einfacher und zusammengesetzter Schlüssel).
	u := schreibUmg(t)
	a, err := u.mutiere("PATCH", "2", model.EntityEingabe{Typ: typProduct, Properties: []model.PropertyWert{pw("Name", model.ArtString, "neu")}})
	erwarteStatus(t, "PATCH ohne Schlüssel", a, err, 204)
	u = schreibUmg(t)
	u.en.setze("Pairs", paar("x", "y"))
	u.paarURL("'x'", "'y'")
	u.ei.ergebnis = model.EntityEingabe{Typ: "N.Pair", Properties: []model.PropertyWert{pw("Name", model.ArtString, "neu")}}
	a, err = u.rufe(anfrage("PATCH", dienstRoot+"/x"))
	erwarteStatus(t, "PATCH Pairs ohne Schlüssel", a, err, 204)
	if n != 16 {
		t.Errorf("Menge %d, erwartet 16 (12 Widersprüche und 4 Sollfälle)", n)
	}
	// Paare: der Widerspruch gewinnt gegen die fehlende Entity, der Fehler des Lesers gegen den Widerspruch.
	u = schreibUmg(t)
	_, err = u.mutiere("PATCH", "99", eingabeIDName("98", "neu"))
	pruefeFehler(t, "Widerspruch und fehlende Entity", err, e6007)
	u = schreibUmg(t)
	u.p.setzeFehler("Eingabe.LeseEntity", &model.EingabeFehler{Ursache: model.EingabeNavigation})
	_, err = u.mutiere("PATCH", "2", eingabeIDName("3", "neu"))
	pruefeFehler(t, "Widerspruch und Fehler des Lesers", err, e6002)
}

// TestMarkerText (W30, M13): der Text `[REDACTED]` in einer Edm.String-Property ist ODR-E6007 vor der Transaktion; Methode 3 ×
// Text 4 × Property 4 = 48.
func TestMarkerText(t *testing.T) {
	texte := []struct {
		text    string
		marker  bool
		kontext string
	}{{"[REDACTED]", true, "Marker"}, {"[redacted]", false, "Kleinschreibung"}, {"[REDACTED] ", false, "Leerzeichen dahinter"}, {"", false, "leer"}}
	n := 0
	for _, m := range dreiMethoden() {
		for _, tx := range texte {
			for _, stelle := range []string{"A", "B", "Name", "Products.Name"} {
				n++
				u := schreibUmg(t)
				a, b, name := "x", "y", "w"
				switch stelle {
				case "A":
					a = tx.text
				case "B":
					b = tx.text
				case "Name":
					name = tx.text
				}
				var err error
				var ant inbound.Antwort
				if stelle == "Products.Name" {
					ant, err = u.mutiere(m, schluesselFuer(m), eingabeIDName(schluesselFuer(m), tx.text))
				} else {
					u.paarURL("'"+a+"'", "'"+b+"'")
					u.ei.ergebnis = paarEingabe(a, b, name)
					if m == "POST" {
						u.ressource(model.RessourceMenge, "Pairs", model.Optionen{})
					}
					ant, err = u.rufe(anfrage(m, dienstRoot+"/x"))
				}
				kontext := fmt.Sprintf("%s %s %s", m, tx.kontext, stelle)
				switch {
				case tx.marker:
					pruefeFehler(t, kontext, err, e6007)
					if u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
						t.Errorf("%s: Transaktion gerufen", kontext)
					}
				case m == "POST" || stelle == "Products.Name":
					erwarteStatus(t, kontext, ant, err, sollStatus(m))
				default:
					pruefeFehler(t, kontext, err, e6003)
				}
			}
		}
	}
	if n != 48 {
		t.Errorf("Menge %d, erwartet 48", n)
	}
}
