package mock_test

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// zaehlLeser ist ein Body, der jeden Lesezugriff zählt: der Dienst liest den Body nie (SPEC-195 Nr. 2).
type zaehlLeser struct{ reads int }

func (l *zaehlLeser) Read([]byte) (int, error) { l.reads++; return 0, io.EOF }

// mitPraefix liefert die Operationen des Protokolls, die mit dem Präfix beginnen.
func mitPraefix(p *prot, praefix string) []string {
	var r []string
	for _, o := range p.ops() {
		if strings.HasPrefix(o, praefix) {
			r = append(r, o)
		}
	}
	return r
}

// TestArgumente hält Z03: vier Verstöße sind ErrUngueltig ohne Code und ohne Port, und die Grenzen Laenge -1 und 0 sind gültig;
// der Body wird nie gelesen.
func TestArgumente(t *testing.T) {
	l := &zaehlLeser{}
	faelle := []struct {
		name      string
		aendere   func(a *inbound.Anfrage)
		keinCtx   bool
		ungueltig bool
	}{
		{"ctx nil", func(*inbound.Anfrage) {}, true, true},
		{"Methode leer", func(a *inbound.Anfrage) { a.Methode = "" }, false, true},
		{"Laenge -2", func(a *inbound.Anfrage) { a.Laenge, a.Body = -2, l }, false, true},
		{"Body nil bei Laenge 1", func(a *inbound.Anfrage) { a.Laenge, a.Body = 1, nil }, false, true},
		{"Laenge -1", func(a *inbound.Anfrage) { a.Laenge, a.Body = -1, nil }, false, false},
		{"Laenge 0", func(a *inbound.Anfrage) { a.Laenge, a.Body = 0, nil }, false, false},
		{"Laenge 3 mit Body", func(a *inbound.Anfrage) { a.Laenge, a.Body = 3, l }, false, false},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neueUmg(t)
			a := metadatenAnfrage()
			a.Pfad += "?" + kanarienwert
			a.RawQuery = "c=" + kanarienwert
			f.aendere(&a)
			m := u.mocker()
			ctx := u.ctx
			if f.keinCtx {
				ctx = keinKontext()
			}
			antwort, err := m.Mock(ctx, a)
			if l.reads != 0 {
				t.Errorf("Body gelesen: %d", l.reads)
			}
			if !f.ungueltig {
				if err != nil {
					t.Fatalf("Fehler: %v", err)
				}
				lies(t, antwort)
				return
			}
			var fe *inbound.Fehler
			if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &fe) {
				t.Errorf("Fehler = %v, erwartet ErrUngueltig ohne Code", err)
			}
			if err != nil && strings.Contains(err.Error(), kanarienwert) {
				t.Errorf("Text nennt einen Wert: %q", err.Error())
			}
			if ops := u.p.ops(); len(ops) != 0 {
				t.Errorf("Ports gerufen: %v", ops)
			}
			if antwort.Body != nil || antwort.Status != 0 {
				t.Errorf("Antwort nicht der Nullwert: %+v", antwort)
			}
		})
	}
}

func wurzelnM2() []string { return []string{"/", "/a", "/a/b", "/ab", "/a%2Fb"} }

// formen sind die zwölf Pfade zu einem Root (M2): gleich, mit /, mit /x, mit x, in anderer Schreibweise, leer, ohne
// führenden /, nur /, mit %2F, mit //, kürzer als der Root, Root als Präfix eines Segments.
func formenM2() []func(r string) string {
	return []func(r string) string{
		func(r string) string { return r },
		func(r string) string { return r + "/" },
		func(r string) string { return r + "/x" },
		func(r string) string { return r + "x" },
		strings.ToUpper,
		func(string) string { return "" },
		func(r string) string { return strings.TrimPrefix(r, "/") + "x" },
		func(string) string { return "/" },
		func(r string) string { return r + "/x%2Fy" },
		func(r string) string { return r + "//x" },
		func(r string) string { return r[:len(r)-1] },
		func(r string) string { return r + "xy/z" },
	}
}

// keiner steht in der Tabelle für „kein Dienst“.
const keiner = "-"

// erwartetePfade ist das Orakel von Z04 und Z05 (M2): der OData-Pfad je Root und Form, von Hand gesetzt.
func erwartetePfade() [5][12]string {
	return [5][12]string{
		{"/", "//", "//x", "/x", "/", keiner, keiner, "/", "//x%2Fy", "///x", keiner, "/xy/z"},
		{"/", "/", "/x", keiner, keiner, keiner, keiner, keiner, "/x%2Fy", "//x", keiner, keiner},
		{"/", "/", "/x", keiner, keiner, keiner, keiner, keiner, "/x%2Fy", "//x", keiner, keiner},
		{"/", "/", "/x", keiner, keiner, keiner, keiner, keiner, "/x%2Fy", "//x", keiner, keiner},
		{"/", "/", "/x", keiner, keiner, keiner, keiner, keiner, "/x%2Fy", "//x", keiner, keiner},
	}
}

// passt ist das Orakel für beliebige Paare (Root, Pfad) der Mehrdienstläufe: gleich oder an einer Segmentgrenze; der
// Root / geht jedem Pfad voraus, der mit / beginnt.
func passt(root, pfad string) (string, bool) {
	if root == "/" {
		return pfad, strings.HasPrefix(pfad, "/")
	}
	rest, ok := strings.CutPrefix(pfad, root)
	if !ok || rest != "" && !strings.HasPrefix(rest, "/") {
		return "", false
	}
	if rest == "" {
		rest = "/"
	}
	return rest, true
}

const rohquery = "a=%2F&$top=1&b=%20x"

// TestDienstWahl hält Z04 und Z05 (M2): 5 Roots × 12 Pfade gegen die Tabelle von Hand, dann alle fünf Roots in einer Konfig
// gegen das Orakel (der längste Root gewinnt); ParseURL erhält genau einmal den OData-Pfad und RawQuery unverändert.
func TestDienstWahl(t *testing.T) {
	wurzeln, formen, tabelle := wurzelnM2(), formenM2(), erwartetePfade()
	if len(wurzeln) != 5 || len(formen) != 12 {
		t.Fatalf("%d Roots und %d Formen, M2 nennt 5 und 12", len(wurzeln), len(formen))
	}
	zelle := 0
	for i, root := range wurzeln {
		for j, form := range formen {
			zelle++
			pfad := form(root)
			soll := tabelle[i][j]
			if rest, ok := passt(root, pfad); (soll != keiner) != ok || ok && rest != soll {
				t.Fatalf("Orakel passt nicht zur Tabelle bei Root %q Pfad %q: %q %v, Tabelle %q", root, pfad, rest, ok, soll)
			}
			u := neueUmg(t)
			u.k.Dienste = []mock.Dienst{{ID: dienstID, Root: root}}
			a := anfrage("GET", pfad)
			a.RawQuery = rohquery
			antwort, err := u.rufe(a)
			if soll == keiner {
				_ = erwarteFehler(t, root+" "+pfad, err, model.CodeE5005)
				if ops := u.p.ops(); len(ops) != 0 {
					t.Errorf("%q %q: Ports gerufen: %v", root, pfad, ops)
				}
				continue
			}
			if err != nil {
				t.Errorf("%q %q: %v", root, pfad, err)
				continue
			}
			lies(t, antwort)
			urls := u.p.von("URL.ParseURL")
			if len(urls) != 1 || !reflect.DeepEqual(urls[0].args, []any{soll, rohquery}) {
				t.Errorf("%q %q: ParseURL %v, erwartet (%q, %q)", root, pfad, urls, soll, rohquery)
			}
		}
	}
	if zelle != 60 {
		t.Errorf("%d Zellen, M2 nennt 60", zelle)
	}

	// Alle fünf Roots zugleich: der längste passende gewinnt. Die Reihenfolge von Dienste ist eine Achse: wie angegeben,
	// umgekehrt, nach Länge auf- und absteigend (4 Ordnungen).
	for _, ordnung := range [][]int{{0, 1, 2, 3, 4}, {4, 3, 2, 1, 0}, {0, 1, 3, 2, 4}, {4, 2, 3, 1, 0}} {
		pfadWahl(t, wurzeln, formen, ordnung)
	}
}

// pfadWahl fährt alle Pfade gegen alle Roots, die Dienste in der Reihenfolge ordnung.
func pfadWahl(t *testing.T, wurzeln []string, formen []func(string) string, ordnung []int) {
	t.Helper()
	for i, root := range wurzeln {
		for _, form := range formen {
			pfad := form(root)
			u := neueUmg(t)
			var gewaehlt, rest string
			u.k.Dienste = nil
			for _, k := range ordnung {
				r := wurzeln[k]
				id := "d" + string(rune('0'+k))
				u.k.Dienste = append(u.k.Dienste, mock.Dienst{ID: id, Root: r})
				u.md.setze(model.Service{ID: id, MetadataBlob: metaRef})
				if x, ok := passt(r, pfad); ok && (gewaehlt == "" || len(r) > len(wurzeln[int(gewaehlt[1]-'0')])) {
					gewaehlt, rest = id, x
				}
			}
			u.ws.ersatz = func(id string) (model.Workspace, error) {
				return model.Workspace{ID: id, ServiceID: gewaehlt, BaselineSnapshotID: "b"}, nil
			}
			a := anfrage("GET", pfad)
			a.RawQuery = rohquery
			antwort, err := u.rufe(a)
			if gewaehlt == "" {
				_ = erwarteFehler(t, "mehrere "+pfad, err, model.CodeE5005)
				continue
			}
			if err != nil {
				t.Errorf("Root %d, Pfad %q: %v", i, pfad, err)
				continue
			}
			lies(t, antwort)
			urls := u.p.von("URL.ParseURL")
			meta := u.p.von("Metadaten.Get")
			if len(urls) != 1 || !reflect.DeepEqual(urls[0].args, []any{rest, rohquery}) ||
				len(meta) != 1 || meta[0].args[0] != gewaehlt {
				t.Errorf("Pfad %q: ParseURL %v, Metadaten %v, erwartet %s (%q)", pfad, urls, meta, gewaehlt, rest)
			}
		}
	}
}

const headerName = "X-Recorder-Workspace"

// TestZuordnung hält Z06 (M3): 16 Formen des Headers × 6 Werte; genau ein nichtleerer Wert, byte-gleich benutzt, ohne
// Beachtung der Schreibweise des Namens; jede andere Form ist ODR-E5003 ohne Port.
func TestZuordnung(t *testing.T) {
	werte := []string{"a", strings.Repeat("w", 64), strings.Repeat("W", 200), "WsMitGross", "a/b", "a\x00b"}
	formenH := []struct {
		name string
		ok   bool
		bau  func(v string) map[string][]string
	}{
		{"fehlt", false, func(string) map[string][]string { return map[string][]string{"X-Anderes": {"x"}} }},
		{"leere Liste", false, func(string) map[string][]string { return map[string][]string{headerName: nil} }},
		{"leerer Wert", false, func(string) map[string][]string { return map[string][]string{headerName: {""}} }},
		{"ein Wert", true, func(v string) map[string][]string { return map[string][]string{headerName: {v}} }},
		{"zwei gleiche", false, func(v string) map[string][]string { return map[string][]string{headerName: {v, v}} }},
		{"zwei verschiedene", false, func(v string) map[string][]string { return map[string][]string{headerName: {v, v + "2"}} }},
		{"zwei Schreibweisen", false, func(v string) map[string][]string {
			return map[string][]string{headerName: {v}, "x-recorder-workspace": {v}}
		}},
		{"Name klein", true, func(v string) map[string][]string { return map[string][]string{"x-recorder-workspace": {v}} }},
		{"Name gemischt", true, func(v string) map[string][]string { return map[string][]string{"x-RECORDER-workspace": {v}} }},
		{"Rand-Leerzeichen", true, func(v string) map[string][]string { return map[string][]string{headerName: {" " + v + " "}} }},
		{"drei Werte", false, func(v string) map[string][]string {
			return map[string][]string{headerName: {v, v + "2", v + "3"}}
		}},
		{"zwei Werte, der erste leer", false, func(v string) map[string][]string { return map[string][]string{headerName: {"", v}} }},
		{"zwei Werte, der zweite leer", false, func(v string) map[string][]string { return map[string][]string{headerName: {v, ""}} }},
		// Zeichen jenseits von ASCII aus Bytewerten (AGENTS.md §3.18): U+212A KELVIN SIGN faltet zu k, U+017F zu s.
		{"Name mit Kelvin-Zeichen", false, func(v string) map[string][]string {
			return map[string][]string{"x-recorder-wor" + string([]byte{0xE2, 0x84, 0xAA}) + "space": {v}}
		}},
		{"Name mit langem s", false, func(v string) map[string][]string {
			return map[string][]string{"x-recorder-work" + string([]byte{0xC5, 0xBF}) + "pace": {v}}
		}},
		{"Wert mit Kelvin-Zeichen", true, func(v string) map[string][]string {
			return map[string][]string{headerName: {v + string([]byte{0xE2, 0x84, 0xAA})}}
		}},
	}
	n := 0
	for _, f := range formenH {
		for _, v := range werte {
			n++
			u := neueUmg(t)
			u.ws.ersatz = func(id string) (model.Workspace, error) {
				return model.Workspace{ID: id, ServiceID: dienstID, BaselineSnapshotID: "b"}, nil
			}
			a := metadatenAnfrage()
			a.Header = f.bau(v)
			antwort, err := u.rufe(a)
			if !f.ok {
				_ = erwarteFehler(t, f.name, err, model.CodeE5003)
				if ops := u.p.ops(); len(ops) != 0 {
					t.Errorf("%s: Ports gerufen: %v", f.name, ops)
				}
				continue
			}
			if err != nil {
				t.Errorf("%s %q: %v", f.name, v, err)
				continue
			}
			lies(t, antwort)
			var soll string
			for _, w := range a.Header {
				soll = w[0]
			}
			gets := u.p.von("Workspaces.Get")
			if len(gets) != 1 || gets[0].args[0] != soll {
				t.Errorf("%s: Get %v, erwartet %q", f.name, gets, soll)
			}
		}
	}
	if n != 96 {
		t.Errorf("%d Zellen, 16 Formen × 6 Werte = 96", n)
	}
}

// TestFremdeKennungen hält Z12 für die Menge der Kennungsformen des fremden Service: gleiche Länge, Präfix, Erweiterung,
// andere Schreibweise, Rand-Leerzeichen, anderer Anfang und leer sind je ODR-E5004; nur die gleiche Kennung ist eigen.
func TestFremdeKennungen(t *testing.T) {
	fremd := []string{"svx", "xvc", "sXc", "sv", "s", "svc2", "svcx", "SVC", "svc ", " svc", "xsvc", ""}
	for _, k := range append([]string{dienstID}, fremd...) {
		u := neueUmg(t)
		u.ws.setze(model.Workspace{ID: wsID, ServiceID: k, BaselineSnapshotID: "b"})
		antwort, err := u.rufe(metadatenAnfrage())
		if k == dienstID {
			if err != nil {
				t.Errorf("eigene Kennung: %v", err)
				continue
			}
			lies(t, antwort)
			continue
		}
		_ = erwarteFehler(t, "fremd "+k, err, model.CodeE5004)
	}
}

// TestSessionHeaderOhneWirkung hält Z07: die Header der Recording- und Replay-Session ändern weder Antwort noch Aufrufe
// (fehlend, vorhanden, mehrfach, leer).
func TestSessionHeaderOhneWirkung(t *testing.T) {
	lauf := func(h map[string][]string) (int, map[string][]string, []byte, []string) {
		u := neueUmg(t)
		a := metadatenAnfrage()
		for k, v := range h {
			a.Header[k] = v
		}
		antwort, err := u.rufe(a)
		if err != nil {
			t.Fatalf("Header %v: %v", h, err)
		}
		return antwort.Status, antwort.Header, lies(t, antwort), u.p.ops()
	}
	s0, h0, b0, o0 := lauf(nil)
	namen := []string{"X-Recorder-Recording-Session", "X-Recorder-Replay-Session"}
	werte := [][]string{nil, {"s1"}, {"s1", "s2"}, {""}, {"ungültig/ü"}}
	n := 0
	for _, name := range namen {
		for _, w := range werte {
			n++
			s, h, b, o := lauf(map[string][]string{name: w})
			if s != s0 || !reflect.DeepEqual(h, h0) || !bytes.Equal(b, b0) || !reflect.DeepEqual(o, o0) {
				t.Errorf("%s %v ändert Antwort oder Aufrufe: %d %v %v", name, w, s, h, o)
			}
		}
	}
	if n != 10 {
		t.Errorf("%d Fälle", n)
	}
}

// TestMethode hält Z10: GET und HEAD laufen weiter, jede andere Methode ist ODR-E6002, bei einem nur lesenden Satz
// ODR-E6006; nach dem Lesen der URL, vor dem Workspace, ohne den Body zu lesen.
func TestMethode(t *testing.T) {
	methoden := []string{"POST", "PATCH", "PUT", "DELETE", "OPTIONS", "get", "Get", "FOO", "head", "Head", "hEAD"}
	if len(methoden) != 11 {
		t.Fatal("Z10: elf Methoden")
	}
	modi := []struct {
		name   string
		satz   func(u *umg) *outbound.CodecSatz
		kompat bool
	}{
		{"ohne Satz", func(*umg) *outbound.CodecSatz { return nil }, false},
		{"Satz schreibend", func(u *umg) *outbound.CodecSatz { return u.c.satz(false, "") }, false},
		{"Satz nur lesend", func(u *umg) *outbound.CodecSatz { return u.c.satz(true, "") }, true},
	}
	for _, mo := range modi {
		for _, me := range methoden {
			u := neueUmg(t)
			u.k.Dienste[0].Codecs = mo.satz(u)
			l := &zaehlLeser{}
			a := anfrage(me, dienstRoot+"/$metadata")
			a.Body, a.Laenge = l, 3
			_, err := u.rufe(a)
			code := model.CodeE6002
			if mo.kompat {
				code = model.CodeE6006
			}
			f := erwarteFehler(t, mo.name+" "+me, err, code)
			if f != nil {
				var kf *model.KompatFehler
				switch {
				case mo.kompat && (!errors.As(f, &kf) || kf.Ursache != model.KompatMethode):
					t.Errorf("%s %s: Ursache %v, erwartet KompatFehler methode", mo.name, me, f.Ursache)
				case !mo.kompat && f.Ursache != nil:
					t.Errorf("%s %s: Ursache %v, erwartet keine", mo.name, me, f.Ursache)
				}
			}
			if ops := u.p.ops(); !reflect.DeepEqual(ops, []string{"URL.ParseURL"}) {
				t.Errorf("%s %s: Aufrufe %v, erwartet nur ParseURL", mo.name, me, ops)
			}
			if l.reads != 0 {
				t.Errorf("%s %s: Body gelesen", mo.name, me)
			}
		}
		for _, me := range []string{"GET", "HEAD"} {
			u := neueUmg(t)
			u.k.Dienste[0].Codecs = mo.satz(u)
			antwort, err := u.rufe(anfrage(me, dienstRoot+"/$metadata"))
			if err != nil || antwort.Status != 200 {
				t.Errorf("%s %s: %d %v", mo.name, me, antwort.Status, err)
				continue
			}
			lies(t, antwort)
		}
	}
}

// TestWorkspace hält Z12: Workspaces.Get genau einmal mit dem Wert unverändert, sonst keine Operation; die Fehler des Ports
// nach SPEC-109 Nr. 5 mit dem Gegenstand Workspace, ein fremder Service ist ODR-E5004 ohne Ursache.
func TestWorkspace(t *testing.T) {
	const wert = "  Ws-Gross 1  "
	faelle := []struct {
		name    string
		antwort func(id string) (model.Workspace, error)
		code    model.Code
		ursache bool
	}{
		{"eigen", func(id string) (model.Workspace, error) { return model.Workspace{ID: id, ServiceID: dienstID}, nil }, 0, false},
		{"fremd", func(id string) (model.Workspace, error) { return model.Workspace{ID: id, ServiceID: "anderer"}, nil }, model.CodeE5004, false},
		{"nicht gefunden", func(string) (model.Workspace, error) { return model.Workspace{}, outbound.ErrNichtGefunden }, model.CodeE7001, true},
		{"beschädigt", func(string) (model.Workspace, error) { return model.Workspace{}, outbound.ErrBeschaedigt }, model.CodeE3001, true},
		{"anderer", func(string) (model.Workspace, error) { return model.Workspace{}, errors.New("roh " + kanarienwert) }, model.CodeE9001, true},
	}
	for _, f := range faelle {
		u := neueUmg(t)
		u.ws.ersatz = f.antwort
		a := metadatenAnfrage()
		a.Header[headerName] = []string{wert}
		antwort, err := u.rufe(a)
		gets := u.p.von("Workspaces.Get")
		if len(gets) != 1 || gets[0].args[0] != wert || gets[0].ctx != u.ctx {
			t.Errorf("%s: Get %v", f.name, gets)
		}
		if w := mitPraefix(u.p, "Workspaces."); len(w) != 1 {
			t.Errorf("%s: weitere Operationen am Port: %v", f.name, w)
		}
		if f.code == 0 {
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
				continue
			}
			lies(t, antwort)
			continue
		}
		fe := erwarteFehler(t, f.name, err, f.code)
		if fe == nil {
			continue
		}
		if f.ursache && fe.Ursache == nil || !f.ursache && fe.Ursache != nil {
			t.Errorf("%s: Ursache %v", f.name, fe.Ursache)
		}
		if strings.Contains(fe.Error(), kanarienwert) {
			t.Errorf("%s: Text nennt einen Wert", f.name)
		}
		if got := mitPraefix(u.p, "Metadaten."); len(got) != 0 {
			t.Errorf("%s: Metadaten gerufen: %v", f.name, got)
		}
	}
}

// TestService hält Z13: Metadaten.Get genau einmal mit der Kennung des Dienstes, sonst keine Operation; ein fehlender
// Eintrag ist ODR-E5005, ein leerer Blob ODR-E6004 ohne Ursache und ohne Blobs.Open.
func TestService(t *testing.T) {
	faelle := []struct {
		name    string
		bau     func(u *umg)
		code    model.Code
		ursache bool
	}{
		{"Eintrag", func(*umg) {}, 0, false},
		{"leerer Blob", func(u *umg) { u.md.setze(model.Service{ID: dienstID}) }, model.CodeE6004, false},
		{"nicht gefunden", func(u *umg) { u.md.entferne(dienstID) }, model.CodeE5005, true},
		{"anderer Fehler", func(u *umg) { u.p.setzeFehler("Metadaten.Get", errors.New("roh "+kanarienwert)) }, model.CodeE9001, true},
		{"beschädigt", func(u *umg) { u.p.setzeFehler("Metadaten.Get", outbound.ErrBeschaedigt) }, model.CodeE3001, true},
	}
	for _, f := range faelle {
		u := neueUmg(t)
		u.k.Dienste = []mock.Dienst{{ID: "anderer", Root: "/x"}, {ID: dienstID, Root: dienstRoot}}
		f.bau(u)
		antwort, err := u.rufe(metadatenAnfrage())
		gets := u.p.von("Metadaten.Get")
		if len(gets) != 1 || gets[0].args[0] != dienstID || gets[0].ctx != u.ctx {
			t.Errorf("%s: Get %v", f.name, gets)
		}
		if w := mitPraefix(u.p, "Metadaten."); len(w) != 1 {
			t.Errorf("%s: weitere Operationen am Port: %v", f.name, w)
		}
		if f.code == 0 {
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
				continue
			}
			lies(t, antwort)
			continue
		}
		fe := erwarteFehler(t, f.name, err, f.code)
		if fe == nil {
			continue
		}
		if f.ursache && fe.Ursache == nil || !f.ursache && fe.Ursache != nil {
			t.Errorf("%s: Ursache %v", f.name, fe.Ursache)
		}
		if got := mitPraefix(u.p, "Blobs."); len(got) != 0 {
			t.Errorf("%s: Blobs gerufen: %v", f.name, got)
		}
	}
}
