package httpodata_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Die Notation einer Form des Fehlerbaums (SPEC-136 Nr. 7 a). Blätter: G gültiger *inbound.Fehler, U *inbound.Fehler mit Status 0 (ein Bruch
// des Vertrags), R Fehler ohne Code, C context.Canceled, N typisierter nil-Zeiger *inbound.Fehler, I Fehler mit der Methode Is auf
// context.DeadlineExceeded, A Fehler, dessen Methode As einen gültigen *inbound.Fehler setzt. Einstellig: W(x) fmt.Errorf mit %w, V(x)
// *inbound.Fehler mit x als Ursache, i(x) und a(x) wie I und A mit Unwrap auf x. Zwei- und dreistellig: J(x,y) und J(x,y,z) errors.Join.
type baumKnoten struct {
	art    byte
	kinder []*baumKnoten
}

func (b *baumKnoten) String() string {
	if len(b.kinder) == 0 {
		return string(b.art)
	}
	var s []string
	for _, k := range b.kinder {
		s = append(s, k.String())
	}
	return string(b.art) + "(" + strings.Join(s, ",") + ")"
}

// isBaum ist der Fehler der Notation I und i: nach seiner Methode Is ist er context.DeadlineExceeded; Unwrap liefert kind.
type isBaum struct{ kind error }

func (*isBaum) Error() string { return "Is" }

func (*isBaum) Is(ziel error) bool { return ziel == context.DeadlineExceeded }

func (b *isBaum) Unwrap() error { return b.kind }

// asBaum ist der Fehler der Notation A und a: seine Methode As setzt f; Unwrap liefert kind.
type asBaum struct {
	f    *inbound.Fehler
	kind error
}

func (*asBaum) Error() string { return "As" }

func (a *asBaum) As(ziel any) bool {
	p, ok := ziel.(**inbound.Fehler)
	if !ok {
		return false
	}
	*p = a.f
	return true
}

func (a *asBaum) Unwrap() error { return a.kind }

// fehlerPaar ist Code und Status eines gültigen *inbound.Fehler.
type fehlerPaar struct {
	code   model.Code
	status int
}

// baumPaare liefert fünf verschiedene gültige Paare (ODR-E9001 nicht darunter): der Knoten mit der Nummer k der Form im Vorlauf trägt das
// Paar k, so dass der Code des Treffers den Knoten nennt.
func baumPaare(t *testing.T) []fehlerPaar {
	t.Helper()
	var out []fehlerPaar
	for _, c := range model.AlleCodes() {
		if c == model.CodeE9001 {
			continue
		}
		for _, s := range c.Stati() {
			if s >= 400 && len(out) < 5 {
				out = append(out, fehlerPaar{c, s})
				break
			}
		}
	}
	if len(out) != 5 {
		t.Fatalf("%d Paare, erwartet 5", len(out))
	}
	return out
}

// bau baut den Fehler der Form; die Knoten sind im Vorlauf (Wurzel, dann Kinder von vorn nach hinten) von 0 an nummeriert.
func (b *baumKnoten) bau(nr *int, paare []fehlerPaar) error {
	k := *nr
	*nr++
	p := paare[k%len(paare)]
	var kinder []error
	for _, kind := range b.kinder {
		kinder = append(kinder, kind.bau(nr, paare))
	}
	var kind error
	if len(kinder) == 1 {
		kind = kinder[0]
	}
	switch b.art {
	case 'G':
		return &inbound.Fehler{Code: p.code, Status: p.status}
	case 'U':
		return &inbound.Fehler{Code: p.code, Status: 0}
	case 'R':
		return errors.New("roh")
	case 'C':
		return context.Canceled
	case 'N':
		return (*inbound.Fehler)(nil)
	case 'I', 'i':
		return &isBaum{kind: kind}
	case 'A', 'a':
		return &asBaum{f: &inbound.Fehler{Code: p.code, Status: p.status}, kind: kind}
	case 'W':
		return fmt.Errorf("w: %w", kind)
	case 'V':
		return &inbound.Fehler{Code: p.code, Status: p.status, Ursache: kind}
	case 'J':
		return errors.Join(kinder...)
	}
	panic("unbekannte Art " + string(b.art))
}

// urteil ist das Orakel der Form: eine eigene Vorlauf-Suche über die Beschreibung (nicht über den Fehler und nicht über Code des Adapters).
// Ein nil-Zeiger ist nichts und trägt keine Kinder; ein Kontextfehler (C, I, i) bricht ab; der erste Knoten G, V, A, a oder U im Vorlauf ist
// der Treffer, und U (Status 0) ist ein Bruch.
type urteil struct {
	abbruch bool
	treffer int
	art     byte
	nr      int
}

func (u *urteil) besuche(b *baumKnoten) {
	k := u.nr
	u.nr++
	switch b.art {
	case 'N':
		return
	case 'C', 'I', 'i':
		u.abbruch = true
	case 'G', 'V', 'A', 'a', 'U':
		if u.treffer < 0 {
			u.treffer, u.art = k, b.art
		}
	}
	for _, kind := range b.kinder {
		u.besuche(kind)
	}
}

// baumFormen liefert alle Formen mit genau n Knoten: Blätter (n = 1), einstellige Knoten über n-1 Knoten, Join mit zwei und mit drei Elementen
// über zusammen n-1 Knoten, jede Verteilung der Knoten auf die Elemente.
func baumFormen(n int, ged map[int][]*baumKnoten) []*baumKnoten {
	if f, da := ged[n]; da {
		return f
	}
	var out []*baumKnoten
	if n == 1 {
		for _, a := range []byte("GURCNIA") {
			out = append(out, &baumKnoten{art: a})
		}
	}
	if n >= 2 {
		for _, kind := range baumFormen(n-1, ged) {
			for _, a := range []byte{'W', 'V', 'i', 'a'} {
				out = append(out, &baumKnoten{art: a, kinder: []*baumKnoten{kind}})
			}
		}
	}
	for a := 1; a <= n-2; a++ {
		for _, l := range baumFormen(a, ged) {
			for _, r := range baumFormen(n-1-a, ged) {
				out = append(out, &baumKnoten{art: 'J', kinder: []*baumKnoten{l, r}})
			}
		}
	}
	for a := 1; a <= n-3; a++ {
		for b := 1; a+b <= n-2; b++ {
			for _, x := range baumFormen(a, ged) {
				for _, y := range baumFormen(b, ged) {
					for _, z := range baumFormen(n-1-a-b, ged) {
						out = append(out, &baumKnoten{art: 'J', kinder: []*baumKnoten{x, y, z}})
					}
				}
			}
		}
	}
	ged[n] = out
	return out
}

// TestFehlerbaumJedeForm hält SPEC-136 Nr. 7 a (Suche im Fehlerbaum: Wurzel zuerst, Unwrap() error und Unwrap() []error von vorn nach hinten,
// der erste Treffer gilt, ein nil-Zeiger ist nichts und wird nicht entpackt, die Methoden Is und As folgen errors.Is und errors.As): jede Form
// bis zu fünf Knoten aus 7 Blättern, 4 einstelligen Knoten und dem Join mit zwei und mit drei Elementen (die Grammatik oben; 7, 28, 161, 1379
// und 12670 Formen je Knotenzahl, 14245), mit jeder der 14 Methoden von M-Methoden (199430 Fälle). Das Orakel ist die Vorlauf-Suche über die
// Beschreibung (urteil): Kontextfehler bricht ab, sonst gilt der erste gültige *inbound.Fehler (Code und Status), ein Bruch oder kein Treffer
// ist ODR-E9001 mit 500; gezählt wird genau dieser Code. 16 Formen, die ein Review nannte oder die eine Achse einzeln führen (Wurzel vor
// Kind, zwei gültige Fehler, nur das letzte Element, Join im Join, Wrap im Join, Join im Wrap, Is und As an der Wurzel, im Join und unter
// einem Wrap), müssen in der Grammatik stehen.
func TestFehlerbaumJedeForm(t *testing.T) {
	paare := baumPaare(t)
	ged := map[int][]*baumKnoten{}
	alle := map[string]*baumKnoten{}
	var geordnet []*baumKnoten
	for n, soll := range []int{7, 28, 161, 1379, 12670} {
		f := baumFormen(n+1, ged)
		if len(f) != soll {
			t.Fatalf("%d Formen mit %d Knoten, erwartet %d", len(f), n+1, soll)
		}
		for _, b := range f {
			if alle[b.String()] != nil {
				t.Fatalf("die Form %s kommt zweimal vor", b)
			}
			alle[b.String()] = b
			geordnet = append(geordnet, b)
		}
	}
	if len(geordnet) != 14245 {
		t.Fatalf("%d Formen, erwartet 14245", len(geordnet))
	}
	benannt := []string{
		"V(G)", "V(V(G))", "J(G,G)", "J(R,R,G)", "J(R,G,G)", "W(W(W(G)))", "J(W(G),R)", "W(J(R,G))", "J(J(G,R),G)", "J(J(R,R),G)",
		"a(G)", "J(R,a(R))", "J(R,I)", "W(i(R))", "V(I)", "J(U,G)",
	}
	if len(benannt) != 16 {
		t.Fatalf("%d benannte Formen, erwartet 16", len(benannt))
	}
	for _, s := range benannt {
		if alle[s] == nil {
			t.Fatalf("die benannte Form %s ist nicht in der Grammatik", s)
		}
	}
	type fall struct {
		b *baumKnoten
		m string
	}
	var faelle []fall
	ms := methodenSatz()
	for _, b := range geordnet {
		for _, m := range ms {
			faelle = append(faelle, fall{b, m})
		}
	}
	if len(faelle) != 14245*14 || len(ms) != 14 {
		t.Fatalf("%d Fälle, erwartet %d", len(faelle), 14245*14)
	}
	rot := 0
	for _, f := range faelle {
		nr := 0
		err := f.b.bau(&nr, paare)
		u := &urteil{treffer: -1}
		u.besuche(f.b)
		soll := struct {
			abbruch bool
			code    model.Code
			status  int
		}{abbruch: u.abbruch}
		switch {
		case u.abbruch:
		case u.treffer >= 0 && u.art != 'U':
			soll.code, soll.status = paare[u.treffer%len(paare)].code, paare[u.treffer%len(paare)].status
		default:
			soll.code, soll.status = model.CodeE9001, 500
		}
		p := neuProbe(t, "record", fehlerMit(err))
		jw := neuJWriter(p.j)
		abbruch := p.ruf(jw, jRequest(context.Background(), f.m, "/p"))
		var codes []model.Code
		if !soll.abbruch {
			codes = []model.Code{soll.code}
		}
		got := zaehlerFolge(p.tel.FakeTelemetry)
		if abbruch != soll.abbruch || jw.status != soll.status || fmt.Sprint(got) != fmt.Sprint(codes) {
			rot++
			t.Errorf("%s %s: Abbruch %v, Status %d, Zähler %v; erwartet %v, %d, %v", f.b, f.m, abbruch, jw.status, got, soll.abbruch, soll.status, codes)
			if rot >= 20 {
				t.Fatalf("20 Fälle sind rot, die übrigen laufen nicht")
			}
		}
	}
}
