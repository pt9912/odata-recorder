package admin_test

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// allowAmPfad ist der Wert von Allow am Pfad /r von einRoute: die Methoden aller Routen des Pfades, aufsteigend.
const allowAmPfad = "DELETE, GET, PUT"

// einRoute baut einen Server mit der Route GET /r und liefert ihn mit dem Fake dieser Route; am Pfad /r liegen auch PUT
// und DELETE (zwei Routen mit eigenen Fakes), damit Allow mehr nennt als die Methode der Route.
func einRoute(t *testing.T) (*umgebung, *routeFake) {
	t.Helper()
	route, put, del := neueRoute(), neueRoute(), neueRoute()
	return neuMit(t,
		admin.Route{Methode: "GET", Muster: "/r", Handler: route.handle},
		admin.Route{Methode: "PUT", Muster: "/r", Handler: put.handle},
		admin.Route{Methode: "DELETE", Muster: "/r", Handler: del.handle}), route
}

// pruefeFehlerDerRoute lässt die Route err liefern und hält die Antwort gegen Status, Body (Literal), die Header des
// Status (bei 405 Allow gleich allowAmPfad) und die eine Zählung des Codes; die Route lief genau einmal, und keine Panik
// verließ ServeHTTP.
func pruefeFehlerDerRoute(t *testing.T, name string, err error, status int, body, code string) {
	t.Helper()
	u, route := einRoute(t)
	route.err = err
	rec := u.rufeSicher(t, name, anfrageMit("GET", "/r", nil, nil))
	allow := ""
	if status == 405 {
		allow = allowAmPfad
	}
	pruefeFehlerAntwort(t, name, rec, status, body, allow)
	u.nurZaehlung(t, name, code)
	if route.aufrufe() != 1 {
		t.Errorf("%s: die Route lief %d-mal, erwartet 1", name, route.aufrufe())
	}
}

// verpackung ist eine Form, einen Fehler zu verpacken (M-Verpackung).
type verpackung struct {
	name  string
	packe func(inner error) error
}

// verpackungen sind die fünf Formen von M-Verpackung: direkt, %w, %w dreifach, Join vorn, Join hinten.
func verpackungen() []verpackung {
	anderer := errors.New("anderer Fehler")
	return []verpackung{
		{"direkt", func(e error) error { return e }},
		{"%w", func(e error) error { return fmt.Errorf("verpackt: %w", e) }},
		{"%w dreifach", func(e error) error {
			return fmt.Errorf("a: %w", fmt.Errorf("b: %w", fmt.Errorf("c: %w", e)))
		}},
		{"Join vorn", func(e error) error { return errors.Join(e, anderer) }},
		{"Join hinten", func(e error) error { return errors.Join(anderer, e) }},
	}
}

// TestVerpackungenZahl hält M-Verpackung gegen die Quelle des Tests: fünf Formen, jede liefert den Fehler über errors.Is.
func TestVerpackungenZahl(t *testing.T) {
	gleichAuf(t, "Zahl der Verpackungen", len(verpackungen()), 5)
	for _, v := range verpackungen() {
		if !errors.Is(v.packe(model.ErrUngueltig), model.ErrUngueltig) {
			t.Errorf("%s: errors.Is findet den Fehler nicht", v.name)
		}
	}
}

// TestFehlerobjektFormat hält E1 (SPEC-238 Nr. 5, SPEC-109 Nr. 2): jeder Code aus der Tabelle mit jedem Status, den er
// trägt, ergibt das Fehlerobjekt mit genau drei Feldern in der Reihenfolge code, message, target, ohne Leerraum und ohne
// Zeilenende, mit Content-Type und Content-Length gleich der Länge; Schleife über die Literal-Tabelle × Status.
func TestFehlerobjektFormat(t *testing.T) {
	faelle, erwartet := 0, 0
	for i, z := range specCodes() {
		erwartet += len(z.stati)
		for _, s := range z.stati {
			faelle++
			body := fehlerBody(z, modellCode(i))
			if !strings.HasPrefix(body, `{"error":{"code":"`+z.code+`","message":"`) || !strings.HasSuffix(body, `","target":"`+z.target+`"}}`) {
				t.Fatalf("%s: Orakel %q", z.code, body)
			}
			pruefeFehlerDerRoute(t, z.code+" "+strconv.Itoa(s), &inbound.Fehler{Code: modellCode(i), Status: s}, s, body, z.code)
		}
	}
	if faelle != erwartet || faelle == 0 {
		t.Errorf("%d Paare aus Code und Status geprüft, erwartet %d (Summe der Status der Tabelle)", faelle, erwartet)
	}
}

// TestGueltigerFehlerVerpackt hält B1 (Nr. 6 a 2): ein gültiger *inbound.Fehler bestimmt Status und Fehlerobjekt, direkt,
// verpackt und im Join; 5 Verpackungen × 4 Paare (ODR-E7003 404, ODR-E4001 502 und 504, ODR-E6102 409) = 20 Fälle.
func TestGueltigerFehlerVerpackt(t *testing.T) {
	paare := []struct {
		code   string
		status int
	}{{"ODR-E7003", 404}, {"ODR-E4001", 502}, {"ODR-E4001", 504}, {"ODR-E6102", 409}}
	faelle := 0
	for _, p := range paare {
		c, _ := model.ParseCode(p.code)
		for i, z := range specCodes() {
			if z.code != p.code {
				continue
			}
			gleichAuf(t, "Code der Tabelle "+p.code, modellCode(i), c)
			for _, v := range verpackungen() {
				faelle++
				pruefeFehlerDerRoute(t, p.code+" "+strconv.Itoa(p.status)+" "+v.name, v.packe(&inbound.Fehler{Code: c, Status: p.status}),
					p.status, fehlerBody(z, c), p.code)
			}
		}
	}
	gleichAuf(t, "Zahl der Fälle", faelle, 20)
}

// TestBruchDesVertrags hält B2 (Nr. 6 a 2): ein Code außerhalb der Tabelle (0, -1, letzter + 1), ein Status außerhalb von
// 400 bis 599 (0, 100, 200, 399, 600, 999) und ein Status, den der Code nicht trägt, sind ODR-E9001 mit 500, und zwar
// einmal gezählt. Schleife über 41 Codes der Tabelle und 3 ungültige × die Menge der Status; was der Code trägt, ist E1.
func TestBruchDesVertrags(t *testing.T) {
	stati := slices.Concat([]int{0, 100, 200, 399, 600, 999}, alleStati())
	codes := make([]model.Code, 0, 44)
	traegt := map[model.Code][]int{}
	for i, z := range specCodes() {
		codes = append(codes, modellCode(i))
		traegt[modellCode(i)] = z.stati
	}
	codes = append(codes, 0, -1, model.Code(len(specCodes())+1))
	faelle, getragen := 0, 0
	for _, c := range codes {
		for _, s := range stati {
			if slices.Contains(traegt[c], s) {
				getragen++
				continue
			}
			faelle++
			pruefeFehlerDerRoute(t, fmt.Sprintf("Code %d Status %d", int(c), s), &inbound.Fehler{Code: c, Status: s}, 500, objekt500, "ODR-E9001")
		}
	}
	gleichAuf(t, "Zahl der Codes", len(codes), len(model.AlleCodes())+3)
	gleichAuf(t, "Zahl der Status", len(stati), 6+len(alleStati()))
	gleichAuf(t, "Zahl der Fälle und getragenen Paare", faelle+getragen, len(codes)*len(stati))
	if getragen == 0 || faelle <= getragen {
		t.Errorf("%d getragene Paare, %d Brüche", getragen, faelle)
	}
}

// TestNilZeiger hält B3 (Nr. 6 a 2): ein nil-Zeiger *inbound.Fehler zählt nicht als dieser Fehler, nackt, verpackt und im
// Join mit ErrUngueltig oder einem gültigen *inbound.Fehler, in beiden Reihenfolgen; die Route erhält keine Panik.
func TestNilZeiger(t *testing.T) {
	var nil1 *inbound.Fehler
	gueltig := &inbound.Fehler{Code: model.CodeE7003, Status: 404}
	faelle := []struct {
		name   string
		err    error
		status int
		body   string
		code   string
	}{
		{"nackt", nil1, 500, objekt500, "ODR-E9001"},
		{"verpackt", fmt.Errorf("verpackt: %w", nil1), 500, objekt500, "ODR-E9001"},
		{"Join vorn mit ErrUngueltig", errors.Join(nil1, model.ErrUngueltig), 400, objekt400, "ODR-E2102"},
		{"Join hinten mit ErrUngueltig", errors.Join(model.ErrUngueltig, nil1), 400, objekt400, "ODR-E2102"},
		{"Join vorn mit gültigem Fehler", errors.Join(nil1, gueltig), 404, objektE7003, "ODR-E7003"},
		{"Join hinten mit gültigem Fehler", errors.Join(gueltig, nil1), 404, objektE7003, "ODR-E7003"},
	}
	for _, f := range faelle {
		pruefeFehlerDerRoute(t, f.name, f.err, f.status, f.body, f.code)
	}
	gleichAuf(t, "Zahl der Fälle", len(faelle), 6)
}

// TestErrUngueltig hält B4 (Nr. 6 a 3): model.ErrUngueltig ohne *inbound.Fehler ist ODR-E2102 mit 400 in jeder
// Verpackung (5), der Text des Fehlers nennt den Canary, die Antwort nicht.
func TestErrUngueltig(t *testing.T) {
	faelle := 0
	for _, v := range verpackungen() {
		faelle++
		pruefeFehlerDerRoute(t, v.name, v.packe(fmt.Errorf("%w: %s", model.ErrUngueltig, canary())), 400, objekt400, "ODR-E2102")
	}
	gleichAuf(t, "Zahl der Fälle", faelle, 5)
}

// TestFehlerVorErrUngueltig hält B5 (Nr. 6 a 2 vor 3): neben einem *inbound.Fehler gilt dieser, mit der Ursache
// ErrUngueltig und im Join in beiden Reihenfolgen; ein ungültiger *inbound.Fehler neben ErrUngueltig ist ein Bruch des
// Vertrags (ODR-E9001), nicht ODR-E2102.
func TestFehlerVorErrUngueltig(t *testing.T) {
	ursache := &inbound.Fehler{Code: model.CodeE6101, Status: 422, Ursache: model.ErrUngueltig}
	e7003 := &inbound.Fehler{Code: model.CodeE7003, Status: 404}
	ungueltig := &inbound.Fehler{Code: 0, Status: 400, Ursache: model.ErrUngueltig}
	const objektE6101 = `{"error":{"code":"ODR-E6101","message":"Baseline import invalid or inconsistent","target":"import"}}`
	pruefeFehlerDerRoute(t, "Ursache ErrUngueltig", ursache, 422, objektE6101, "ODR-E6101")
	pruefeFehlerDerRoute(t, "Join, Fehler hinten", errors.Join(model.ErrUngueltig, e7003), 404, objektE7003, "ODR-E7003")
	pruefeFehlerDerRoute(t, "Join, Fehler vorn", errors.Join(e7003, model.ErrUngueltig), 404, objektE7003, "ODR-E7003")
	pruefeFehlerDerRoute(t, "ungültiger Fehler mit Ursache ErrUngueltig", ungueltig, 500, objekt500, "ODR-E9001")
	pruefeFehlerDerRoute(t, "ungültiger Fehler neben ErrUngueltig", errors.Join(ungueltig, model.ErrUngueltig), 500, objekt500, "ODR-E9001")
}

// errQuelle ist eine exportierte Variable `Err…` im Quelltext: ihr Name `paket.Name` und der Text von `errors.New`.
type errQuelle struct {
	name, text string
}

// errVariablen liefert die exportierten Variablen `Err…` der Pakete model und ports/outbound aus dem Quelltext (ohne
// _test.go), nach Namen geordnet; das Orakel von M-Err. Eine Variable, die nicht mit `errors.New("…")` gesetzt wird,
// bricht den Test ab.
func errVariablen(t *testing.T) []errQuelle {
	t.Helper()
	var out []errQuelle
	for _, p := range []string{"model", "ports/outbound"} {
		dir := filepath.Join("..", "..", "..", "hexagon", filepath.FromSlash(p))
		eintraege, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range eintraege {
			if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, e.Name()), nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range f.Decls {
				g, ok := d.(*ast.GenDecl)
				if !ok || g.Tok != token.VAR {
					continue
				}
				for _, s := range g.Specs {
					vs := s.(*ast.ValueSpec)
					for i, n := range vs.Names {
						if strings.HasPrefix(n.Name, "Err") && n.IsExported() {
							out = append(out, errQuelle{f.Name.Name + "." + n.Name, errorsNewText(t, n.Name, vs, i)})
						}
					}
				}
			}
		}
	}
	slices.SortFunc(out, func(a, b errQuelle) int { return strings.Compare(a.name, b.name) })
	return out
}

// errorsNewText liefert den Text des Literals in `errors.New("…")`, mit dem die Variable i von vs gesetzt wird.
func errorsNewText(t *testing.T, name string, vs *ast.ValueSpec, i int) string {
	t.Helper()
	if i < len(vs.Values) {
		if c, ok := vs.Values[i].(*ast.CallExpr); ok && len(c.Args) == 1 {
			if p, fn := aufrufName(c); p == "errors" && fn == "New" {
				if lit, ok := c.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if s, err := strconv.Unquote(lit.Value); err == nil {
						return s
					}
				}
			}
		}
	}
	t.Fatalf("%s wird nicht mit errors.New und einem Literal gesetzt", name)
	return ""
}

// errWert ist eine Variable `Err…` des Pakets model mit ihrem Namen `model.Name`.
type errWert struct {
	name string
	err  error
}

// modellErrWerte sind die Variablen `Err…` des Pakets model ohne model.ErrUngueltig, von Hand gesammelt; die
// Zählprüfung von TestAndereFehler hält sie gegen den Quelltext. Die Variablen von ports/outbound darf kein Adapter
// importieren (`make a-check`): dort baut der Test den Fehler mit dem Text der Variablen aus dem Quelltext.
func modellErrWerte() []errWert {
	return []errWert{
		{"model.ErrKompatibilitaet", model.ErrKompatibilitaet},
		{"model.ErrKonflikt", model.ErrKonflikt},
		{"model.ErrUnsicher", model.ErrUnsicher},
		{"model.ErrDatenUngueltig", model.ErrDatenUngueltig},
		{"model.ErrEingabeUngueltig", model.ErrEingabeUngueltig},
		{"model.ErrEingabeNichtUnterstuetzt", model.ErrEingabeNichtUnterstuetzt},
		{"model.ErrBatchUngueltig", model.ErrBatchUngueltig},
		{"model.ErrBatchNichtUnterstuetzt", model.ErrBatchNichtUnterstuetzt},
		{"model.ErrFilterNichtBeobachtet", model.ErrFilterNichtBeobachtet},
		{"model.ErrMetadatenUngueltig", model.ErrMetadatenUngueltig},
		{"model.ErrURLUngueltig", model.ErrURLUngueltig},
		{"model.ErrURLNichtUnterstuetzt", model.ErrURLNichtUnterstuetzt},
		{"model.ErrKodierung", model.ErrKodierung},
	}
}

// TestAndereFehler hält B6 (Nr. 6 a 4): jeder andere Fehler, auch jeder Fehler eines Ports ohne Code, ist ODR-E9001 mit
// 500. Menge M-Err: jede exportierte Variable `Err…` von model und ports/outbound außer model.ErrUngueltig, gefunden im
// Quelltext (Zählprüfung gegen die Quelle), dazu ein Fehler ohne Variable; jede in jeder Verpackung (5).
func TestAndereFehler(t *testing.T) {
	var quelle []errQuelle
	for _, q := range errVariablen(t) {
		if q.name != "model.ErrUngueltig" {
			quelle = append(quelle, q)
		}
	}
	modell := map[string]error{}
	for _, w := range modellErrWerte() {
		modell[w.name] = w.err
	}
	gezaehlt := 0
	faelle := 0
	for _, q := range quelle {
		err, ok := modell[q.name]
		switch {
		case ok:
			gezaehlt++
		case strings.HasPrefix(q.name, "model."):
			t.Fatalf("%s steht im Quelltext, aber nicht in modellErrWerte", q.name)
		default:
			err = errors.New(q.text)
		}
		for _, v := range verpackungen() {
			faelle++
			pruefeFehlerDerRoute(t, q.name+" "+v.name, v.packe(err), 500, objekt500, "ODR-E9001")
		}
	}
	gleichAuf(t, "Zahl der Variablen von model im Quelltext und im Test", gezaehlt, len(modell))
	pruefeFehlerDerRoute(t, "errors.New", errors.New("x"), 500, objekt500, "ODR-E9001")
	gleichAuf(t, "Zahl der Fälle", faelle, len(quelle)*5)
}

// TestFehlerVorAntwort hält B7 (Nr. 6 b): liefert die Route Fehler und Antwort, gilt der Fehler: *inbound.Fehler,
// ErrUngueltig und ein anderer Fehler (3 Formen), Antwort mit Status 200 und Body.
func TestFehlerVorAntwort(t *testing.T) {
	faelle := []struct {
		name   string
		err    error
		status int
		body   string
		code   string
	}{
		{"*inbound.Fehler", &inbound.Fehler{Code: model.CodeE7003, Status: 404}, 404, objektE7003, "ODR-E7003"},
		{"ErrUngueltig", model.ErrUngueltig, 400, objekt400, "ODR-E2102"},
		{"anderer Fehler", errors.New("x"), 500, objekt500, "ODR-E9001"},
	}
	for _, f := range faelle {
		u, route := einRoute(t)
		route.err, route.antwort = f.err, admin.Antwort{Status: 200, Body: []byte(`{"a":1}`)}
		rec := u.rufeSicher(t, f.name, anfrageMit("GET", "/r", nil, nil))
		pruefeFehlerAntwort(t, f.name, rec, f.status, f.body, "")
		u.nurZaehlung(t, f.name, f.code)
	}
	gleichAuf(t, "Zahl der Fälle", len(faelle), 3)
}

// TestStatusDerAntwort hält B9 (Nr. 6 d): ein Status der Antwort außerhalb von 200 bis 299 ist ODR-E9001 mit 500 und
// zählt einmal; 200 und 299 gehen durch und zählen nicht. M-Status: 0, 199, 200, 299, 300, 399, 400, 500, 600.
func TestStatusDerAntwort(t *testing.T) {
	stati := []int{0, 199, 200, 299, 300, 399, 400, 500, 600}
	for _, s := range stati {
		name := "Status " + strconv.Itoa(s)
		u, route := einRoute(t)
		route.antwort = admin.Antwort{Status: s, Body: []byte(`{"a":1}`)}
		rec := u.rufeSicher(t, name, anfrageMit("GET", "/r", nil, nil))
		if s >= 200 && s <= 299 {
			if rec.Code != s || rec.Body.String() != `{"a":1}` {
				t.Errorf("%s: Antwort %d %q, erwartet %d und den Body der Route", name, rec.Code, rec.Body.String(), s)
			}
			u.nurZaehlung(t, name)
			continue
		}
		pruefeFehlerAntwort(t, name, rec, 500, objekt500, "")
		u.nurZaehlung(t, name, "ODR-E9001")
	}
	gleichAuf(t, "Zahl der Status", len(stati), 9)
}

// panikWert ist ein Wert, mit dem eine Route in Panik gerät (M-Panik).
type panikWert struct {
	name string
	wert any
}

// panikWerte sind die sieben Werte von M-Panik: string, error, *inbound.Fehler, nil, http.ErrAbortHandler, int, Canary.
func panikWerte() []panikWert {
	return []panikWert{
		{"string", "boom"}, {"error", errors.New("boom")},
		{"*inbound.Fehler", &inbound.Fehler{Code: model.CodeE7003, Status: 404}},
		{"nil", nil}, {"http.ErrAbortHandler", http.ErrAbortHandler}, {"int", 42}, {"Canary", canary()},
	}
}

// TestPanikDerRoute hält B8 (Nr. 6 c) auf einem laufenden Server: eine Panik der Route mit jedem der sieben Werte (string,
// error, *inbound.Fehler, nil, http.ErrAbortHandler, int, Canary) ergibt ODR-E9001 mit 500, einmal gezählt; der Server
// bedient weiter (zweiter Request 200), und der Wert erreicht weder Antwort noch Telemetrie noch Audit.
func TestPanikDerRoute(t *testing.T) {
	werte := panikWerte()
	gleichAuf(t, "Zahl der Panikwerte", len(werte), 7)
	for _, w := range werte {
		route := neueRoute()
		route.aendere = func(a admin.Anfrage) {
			if a.Pfad == "/haengt" {
				panic(w.wert)
			}
		}
		u := neuLive(t, route)
		x := erwarte(t, u.async("GET", "/haengt"), w.name)
		if x.err != nil || x.status != 500 || string(x.body) != objekt500 {
			t.Errorf("%s: Antwort %+v, erwartet 500 %s", w.name, x, objekt500)
		}
		if x := erwarte(t, u.async("GET", "/p"), w.name+": zweiter Request"); x.err != nil || x.status != 200 {
			t.Errorf("%s: zweiter Request %+v, erwartet 200", w.name, x)
		}
		codes, andere := zaehlungenVon(u.tele)
		if !slices.Equal(codes, []string{"ODR-E9001"}) || andere != 0 || u.audit.Aufrufe() != 0 {
			t.Errorf("%s: CountError %q, andere Aufrufe %d, Audit %d, erwartet einmal ODR-E9001 und sonst 0", w.name, codes, andere, u.audit.Aufrufe())
		}
		if text := fmt.Sprintf("%+v", u.tele.Aufrufe()); strings.Contains(text, canary()) || strings.Contains(string(x.body), canary()) {
			t.Errorf("%s: der Canary steht in Telemetrie oder Antwort", w.name)
		}
		gleichAuf(t, w.name+": Aufrufe der Route", route.aufrufe(), 2)
	}
}
