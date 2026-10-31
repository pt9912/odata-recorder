package replay_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
)

// erfolg legt einen replayfähigen Exchange unter dem Key A an.
func erfolg(u *umg) { u.lege(u.austausch(1, u.keyA(), true)) }

// dienstDefinition sind die vier Dienste von Z8 und Z9.
func vierDienste() []replay.Dienst {
	return []replay.Dienst{{ID: "svc-r", Root: "/"}, {ID: "svc-a", Root: "/a"}, {ID: "svc-ab", Root: "/a/b"}, {ID: "svc-x", Root: "/ab"}}
}

// fuenfDienste sind die vier Dienste und ein fünfter mit einem Root gleicher Länge wie /ab (Gleichstand der Längen,
// verschiedene Roots).
func fuenfDienste() []replay.Dienst {
	return append(vierDienste(), replay.Dienst{ID: "svc-y", Root: "/ac"})
}

// permutationen liefert jede Reihenfolge von ds.
func permutationen(ds []replay.Dienst) [][]replay.Dienst {
	if len(ds) <= 1 {
		return [][]replay.Dienst{append([]replay.Dienst{}, ds...)}
	}
	var r [][]replay.Dienst
	for i := range ds {
		rest := append(append([]replay.Dienst{}, ds[:i]...), ds[i+1:]...)
		for _, p := range permutationen(rest) {
			r = append(r, append([]replay.Dienst{ds[i]}, p...))
		}
	}
	return r
}

// TestDienstWahl hält Z8: Menge 2160 = 120 Reihenfolgen der fünf Dienste × 18 Pfade (die 14 des Plans, /A/b und
// /a/B für die Schreibweise hinter der Segmentgrenze, /ac und /ac/q für den Root gleicher Länge); das Orakel nennt
// Dienst und OData-Pfad und hängt nicht von der Reihenfolge der Konfig ab.
func TestDienstWahl(t *testing.T) {
	orakel := []struct{ pfad, dienst, odata string }{
		{"/", "svc-r", "/"}, {"/a", "svc-a", "/"}, {"/a/", "svc-a", "/"}, {"/a/b", "svc-ab", "/"}, {"/a/b/c", "svc-ab", "/c"},
		{"/ab", "svc-x", "/"}, {"/abc", "svc-r", "/abc"}, {"/A", "svc-r", "/A"}, {"/a%2Fb", "svc-r", "/a%2Fb"}, {"/x", "svc-r", "/x"},
		{"/a/b/", "svc-ab", "/"}, {"/a/bc", "svc-a", "/bc"}, {"", "", ""}, {"x", "", ""},
		{"/A/b", "svc-r", "/A/b"}, {"/a/B", "svc-a", "/B"}, {"/ac", "svc-y", "/"}, {"/ac/q", "svc-y", "/q"},
	}
	reihenfolgen := permutationen(fuenfDienste())
	if len(orakel) != 18 || len(reihenfolgen) != 120 {
		t.Fatalf("%d Pfade, %d Reihenfolgen", len(orakel), len(reihenfolgen))
	}
	n := 0
	for _, dienste := range reihenfolgen {
		for _, o := range orakel {
			n++
			u := neu(t)
			u.k.Dienste = dienste
			a := anf()
			a.Pfad = o.pfad
			if o.dienst == "" {
				_, err := u.replay(context.Background(), a)
				erwarteCode(t, "Pfad "+strconv.Quote(o.pfad), err, model.CodeE5005)
				continue
			}
			u.sess.setze(model.Session{ID: sesID, ServiceID: o.dienst, CreatedAt: zeit(), Status: model.SessionComplete})
			z := u.st.stand()
			z.ws[wsID] = model.Workspace{ID: wsID, ServiceID: o.dienst, BaselineSnapshotID: "snap-a", CreatedAt: zeit()}
			u.st.setze(z)
			u.lege(u.austausch(1, u.keyFuer(a, o.dienst, o.odata), true))
			resp, err := u.replay(context.Background(), a)
			if err != nil {
				t.Errorf("Pfad %q, Reihenfolge %v: %v", o.pfad, dienste, err)
				continue
			}
			_ = resp.Body.Close()
		}
	}
	if n != 2160 {
		t.Errorf("Menge: %d Elemente, erwartet 2160", n)
	}
}

func verbinde(root, rest string) string {
	if root == "/" {
		return rest
	}
	return root + rest
}

// TestKeyTraegtDienstUndPfad hält Z9: Menge 8 (4 Dienste × Root und Root+/x); die Aufzeichnung unter dem
// unabhängig gebildeten Key wird gefunden, die eines anderen Dienstes mit demselben OData-Pfad nicht.
func TestKeyTraegtDienstUndPfad(t *testing.T) {
	n := 0
	for _, d := range vierDienste() {
		for _, rest := range []string{"/", "/x"} {
			n++
			u := neu(t)
			u.k.Dienste = vierDienste()
			u.sess.setze(model.Session{ID: sesID, ServiceID: d.ID, CreatedAt: zeit(), Status: model.SessionComplete})
			z := u.st.stand()
			z.ws[wsID] = model.Workspace{ID: wsID, ServiceID: d.ID, BaselineSnapshotID: "snap-a", CreatedAt: zeit()}
			u.st.setze(z)
			a := anf()
			a.Pfad = verbinde(d.Root, rest)
			if rest == "/" {
				a.Pfad = d.Root
			}
			fremd := "svc-fremd"
			u.lege(u.austausch(1, u.keyFuer(a, fremd, rest), true), u.austausch(2, u.keyFuer(a, d.ID, rest), true))
			resp, err := u.replay(context.Background(), a)
			if err != nil {
				t.Errorf("%s %q: %v", d.ID, a.Pfad, err)
				continue
			}
			if got := seqVon(t, resp); got != 2 {
				t.Errorf("%s %q: SequenceNo %d, erwartet 2 (der eigene Key)", d.ID, a.Pfad, got)
			}
		}
	}
	if n != 8 {
		t.Errorf("Menge: %d, erwartet 8", n)
	}
}

// TestSteuerheader hält Z10: Menge 21 (3 Header × 7 Formen); 18 Verstöße ohne Port, 3 Gutfälle.
func TestSteuerheader(t *testing.T) {
	namen := []string{"X-Recorder-Workspace", "X-Recorder-Recording-Session", "X-Recorder-Replay-Session"}
	wert := map[string]string{namen[0]: wsID, namen[1]: sesID, namen[2]: repID}
	formen := []struct {
		name string
		bad  bool
		mod  func(h map[string][]string, n string)
	}{
		{"fehlt", true, func(h map[string][]string, n string) { delete(h, n) }},
		{"leere Wertliste", true, func(h map[string][]string, n string) { h[n] = []string{} }},
		{"ein leerer Wert", true, func(h map[string][]string, n string) { h[n] = []string{""} }},
		{"zwei verschiedene", true, func(h map[string][]string, n string) { h[n] = []string{wert[n], wert[n] + "x"} }},
		{"zwei gleiche", true, func(h map[string][]string, n string) { h[n] = []string{wert[n], wert[n]} }},
		{"zwei Schreibweisen", true, func(h map[string][]string, n string) { h[strings.ToLower(n)] = []string{wert[n]} }},
		{"kleingeschrieben", false, func(h map[string][]string, n string) { delete(h, n); h[strings.ToLower(n)] = []string{wert[n]} }},
	}
	var schlecht, gut int
	for _, n := range namen {
		for _, f := range formen {
			u := neu(t)
			erfolg(u)
			a := anf()
			f.mod(a.Header, n)
			resp, err := u.replay(context.Background(), a)
			if f.bad {
				schlecht++
				erwarteCode(t, n+" "+f.name, err, model.CodeE5003)
				if u.p.gesamt() != 0 {
					t.Errorf("%s %s: Port gerufen: %v", n, f.name, u.p.ops())
				}
				continue
			}
			gut++
			if err != nil {
				t.Errorf("%s %s: %v", n, f.name, err)
				continue
			}
			_ = resp.Body.Close()
		}
	}
	if schlecht != 18 || gut != 3 {
		t.Errorf("Menge: %d Verstöße, %d Gutfälle", schlecht, gut)
	}
}

// TestReplaySessionForm hält Z11: Menge 771 (Länge 1, 128, 129 und jedes Byte an drei Stellen einer Kennung der
// Länge 3); das Orakel ist das Alphabet 0-9A-Za-z._-, als Menge gebildet.
func TestReplaySessionForm(t *testing.T) {
	erlaubt := map[byte]bool{'.': true, '_': true, '-': true}
	for _, r := range "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" {
		erlaubt[byte(r)] = true
	}
	if len(erlaubt) != 65 {
		t.Fatalf("Alphabet hat %d Zeichen", len(erlaubt))
	}
	type fall struct {
		id  string
		gut bool
	}
	faelle := []fall{{"a", true}, {strings.Repeat("a", 128), true}, {strings.Repeat("a", 129), false}}
	for b := 0; b < 256; b++ {
		c := byte(b)
		for _, id := range [][]byte{{c, 'a', 'a'}, {'a', c, 'a'}, {'a', 'a', c}} {
			faelle = append(faelle, fall{string(id), erlaubt[c]})
		}
	}
	var gut, schlecht int
	for _, f := range faelle {
		u := neu(t)
		erfolg(u)
		a := anf()
		a.Header["X-Recorder-Replay-Session"] = []string{f.id}
		resp, err := u.replay(context.Background(), a)
		if f.gut {
			gut++
			if err != nil {
				t.Errorf("Kennung %q abgewiesen: %v", f.id, err)
				continue
			}
			_ = resp.Body.Close()
			continue
		}
		schlecht++
		erwarteCode(t, "Kennung "+strconv.Quote(f.id), err, model.CodeE5003)
		if u.p.gesamt() != 0 {
			t.Errorf("Kennung %q: Port gerufen", f.id)
		}
	}
	if len(faelle) != 771 || gut != 197 || schlecht != 574 {
		t.Errorf("Menge: %d Elemente, %d gut, %d Verstoß; erwartet 771, 197 und 574", len(faelle), gut, schlecht)
	}
}

// TestKennungenByteGleich hält Z12: Menge 6; Workspace- und Session-Kennung gehen unverändert an die Ports.
func TestKennungenByteGleich(t *testing.T) {
	for _, h := range []string{"X-Recorder-Workspace", "X-Recorder-Recording-Session"} {
		for _, w := range []string{" x", "x ", "X"} {
			u := neu(t)
			erfolg(u)
			a := anf()
			a.Header[h] = []string{w}
			op := "Sessions.Get"
			if h == "X-Recorder-Workspace" {
				op = "tx Workspaces.Get"
			}
			_, _ = u.replay(context.Background(), a)
			evs := u.p.von(op)
			if len(evs) != 1 || len(evs[0].args) != 1 || evs[0].args[0] != w {
				t.Errorf("%s=%q: %s erhielt %v", h, w, op, evs)
			}
		}
	}
}

// TestKeineImplizit hält Z13: ohne Header keine Wahl, auch bei genau einer Session und einem Workspace.
func TestKeineImplizit(t *testing.T) {
	for _, zustand := range []string{"genau einer", "mehrere", "keine"} {
		u := neu(t)
		erfolg(u)
		switch zustand {
		case "mehrere":
			u.sess.setze(model.Session{ID: "ses-b", ServiceID: svcID, CreatedAt: zeit(), Status: model.SessionComplete})
		case "keine":
			u.sess.weg(sesID)
		}
		a := anf()
		a.Header = map[string][]string{}
		_, err := u.replay(context.Background(), a)
		erwarteCode(t, zustand, err, model.CodeE5003)
		if u.p.gesamt() != 0 {
			t.Errorf("%s: Port gerufen: %v", zustand, u.p.ops())
		}
	}
}

// TestSessionZustaende hält Z14: Menge 5.
func TestSessionZustaende(t *testing.T) {
	faelle := []struct {
		name string
		set  func(u *umg)
		code model.Code
	}{
		{"unbekannt", func(u *umg) { u.sess.weg(sesID) }, model.CodeE7003},
		{"anderer Service", func(u *umg) {
			u.sess.setze(model.Session{ID: sesID, ServiceID: "svc-b", CreatedAt: zeit(), Status: model.SessionComplete})
		}, model.CodeE5004},
		{"open", func(u *umg) {
			u.sess.setze(model.Session{ID: sesID, ServiceID: svcID, CreatedAt: zeit(), Status: model.SessionOpen})
		}, model.CodeE7009},
		{"complete", func(u *umg) {}, 0},
		{"incomplete", func(u *umg) {
			u.sess.setze(model.Session{ID: sesID, ServiceID: svcID, CreatedAt: zeit(), Status: model.SessionIncomplete})
		}, 0},
	}
	for _, f := range faelle {
		u := neu(t)
		erfolg(u)
		f.set(u)
		vorher := u.st.stand()
		resp, err := u.replay(context.Background(), anf())
		if f.code == 0 {
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
				continue
			}
			_ = resp.Body.Close()
			continue
		}
		erwarteCode(t, f.name, err, f.code)
		for _, op := range []string{"Exchanges.ListBySession", "Transaktion.WithWorkspaceTransaction", "tx Cursors.Put", "Blobs.Open"} {
			if u.p.anzahl(op) != 0 {
				t.Errorf("%s: %s gerufen", f.name, op)
			}
		}
		if got := u.st.stand(); len(got.cur[wsID]) != len(vorher.cur[wsID]) {
			t.Errorf("%s: Cursor geändert", f.name)
		}
	}
}

// TestWorkspaceZustaende hält Z15: Menge 3.
func TestWorkspaceZustaende(t *testing.T) {
	faelle := []struct {
		name string
		set  func(u *umg)
		code model.Code
	}{
		{"unbekannt", func(u *umg) { z := u.st.stand(); delete(z.ws, wsID); u.st.setze(z) }, model.CodeE7001},
		{"anderer Service", func(u *umg) {
			z := u.st.stand()
			z.ws[wsID] = model.Workspace{ID: wsID, ServiceID: "svc-b", BaselineSnapshotID: "snap-a", CreatedAt: zeit()}
			u.st.setze(z)
		}, model.CodeE5004},
		{"gut", func(u *umg) {}, 0},
	}
	for _, f := range faelle {
		u := neu(t)
		erfolg(u)
		f.set(u)
		resp, err := u.replay(context.Background(), anf())
		if f.code == 0 {
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
				continue
			}
			_ = resp.Body.Close()
			continue
		}
		erwarteCode(t, f.name, err, f.code)
		if u.p.anzahl("tx Cursors.Get") != 0 || u.p.anzahl("tx Cursors.Put") != 0 {
			t.Errorf("%s: Cursor gelesen oder geschrieben", f.name)
		}
	}
}

// TestRangfolge hält Z16: Menge 17, je Paar ein Request mit beiden Verstößen; erwartet der Sieger.
func TestRangfolge(t *testing.T) {
	ohneWorkspaceHeader := func(a *inbound.Anfrage) { delete(a.Header, "X-Recorder-Workspace") }
	keinWorkspace := func(u *umg) { z := u.st.stand(); delete(z.ws, wsID); u.st.setze(z) }
	fremdWorkspace := func(u *umg) {
		z := u.st.stand()
		z.ws[wsID] = model.Workspace{ID: wsID, ServiceID: "svc-b", BaselineSnapshotID: "snap-a", CreatedAt: zeit()}
		u.st.setze(z)
	}
	gzipBody := func(a *inbound.Anfrage) {
		a.Body = &zaehlLeser{daten: []byte("{}")}
		a.Laenge = 2
		a.Header["Content-Encoding"] = []string{"gzip"}
	}
	reject := func(u *umg, a *inbound.Anfrage) {
		u.k.Regeln.InvalidEncoding = model.KodierungReject
		a.RawQuery = "a=%zz"
	}
	paare := []struct {
		name string
		code model.Code
		set  func(u *umg, a *inbound.Anfrage)
	}{
		{"P1", model.CodeE5005, func(u *umg, a *inbound.Anfrage) { a.Pfad = "/x"; ohneWorkspaceHeader(a) }},
		{"P2", model.CodeE5003, func(u *umg, a *inbound.Anfrage) { ohneWorkspaceHeader(a); u.sess.weg(sesID) }},
		{"P3", model.CodeE7003, func(u *umg, a *inbound.Anfrage) { u.sess.weg(sesID); keinWorkspace(u) }},
		{"P4", model.CodeE5004, func(u *umg, a *inbound.Anfrage) {
			u.sess.setze(model.Session{ID: sesID, ServiceID: "svc-b", CreatedAt: zeit(), Status: model.SessionOpen})
		}},
		{"P5", model.CodeE7009, func(u *umg, a *inbound.Anfrage) {
			u.sess.setze(model.Session{ID: sesID, ServiceID: svcID, CreatedAt: zeit(), Status: model.SessionOpen})
			gzipBody(a)
		}},
		{"P6", model.CodeE7003, func(u *umg, a *inbound.Anfrage) { u.sess.weg(sesID); gzipBody(a) }},
		{"P7", model.CodeE5001, func(u *umg, a *inbound.Anfrage) { gzipBody(a); reject(u, a) }},
		{"P8", model.CodeE5006, func(u *umg, a *inbound.Anfrage) { reject(u, a); keinWorkspace(u) }},
		{"P9", model.CodeE5001, func(u *umg, a *inbound.Anfrage) { gzipBody(a); keinWorkspace(u) }},
		{"P10", model.CodeE7001, func(u *umg, a *inbound.Anfrage) { keinWorkspace(u); u.lege() }},
		{"P11", model.CodeE5004, func(u *umg, a *inbound.Anfrage) { fremdWorkspace(u); u.cursor(model.StrategySequential, 1) }},
		{"P12", model.CodeE5001, func(u *umg, a *inbound.Anfrage) {
			u.lege(u.austausch(1, u.keyB(), true))
			u.cursor(model.StrategyStrict)
		}},
		{"P13", model.CodeE5001, func(u *umg, a *inbound.Anfrage) {
			u.lege(u.austausch(1, u.keyB(), true))
			u.cursor(model.StrategyStrict, 1)
		}},
		{"P14", model.CodeE5007, func(u *umg, a *inbound.Anfrage) {
			u.lege(u.austausch(1, u.keyB(), false), u.austausch(2, u.keyA(), true))
			u.cursor(model.StrategyStrict)
		}},
		{"P15", model.CodeE5008, func(u *umg, a *inbound.Anfrage) {
			u.lege(u.austausch(1, u.keyA(), true), u.austausch(2, u.keyA(), false))
			u.cursor(model.StrategySequential, 1)
		}},
		{"P16", model.CodeE3001, func(u *umg, a *inbound.Anfrage) {
			u.p.setzeFehler("Blobs.Open", outbound.ErrBlobBeschaedigt)
			u.p.setzeFehler("tx Cursors.Put", outbound.ErrFremderService)
		}},
		{"P17", model.CodeE5003, func(u *umg, a *inbound.Anfrage) {
			a.Header["X-Recorder-Replay-Session"] = []string{"a b"}
			u.sess.weg(sesID)
			keinWorkspace(u)
		}},
	}
	if len(paare) != 17 {
		t.Fatalf("%d Paare", len(paare))
	}
	for _, p := range paare {
		u := neu(t)
		erfolg(u)
		a := anf()
		p.set(u, &a)
		_, err := u.replay(context.Background(), a)
		erwarteCode(t, p.name, err, p.code)
	}
}
