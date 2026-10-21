package httpupstream_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// --- A2: Argumente -------------------------------------------------------------------------------

// argumentFall baut einen Fall, dessen Anfrage aendere ungültig macht: model.ErrUngueltig, der
// Server nahm keine Verbindung an, der Body wurde nicht gelesen (SPEC-120 Nr. 2).
func argumentFall(ctx context.Context, name string, aendere func(a *outbound.UpstreamAnfrage, s *roherServer)) fehlerFall {
	return fehlerFall{name: name, art: model.ErrUngueltig, lauf: func(t *testing.T) error {
		s := starteServer(t, beantworte(nil, antwortOK))
		c := neuerClient(t, zeitgrenze)
		body := &zaehlBody{r: strings.NewReader(geheim + "-b")}
		a := mitGeheim(anfrage(s.ziel(), "POST", ""))
		a.Body, a.Laenge = body, int64(len(geheim+"-b"))
		aendere(&a, s)
		_, err := sende(ctx, t, c, a)
		s.keineVerbindung(t)
		if n := body.n.Load(); n != 0 {
			t.Errorf("Body %d-mal gelesen", n)
		}
		return err
	}}
}

func setzeZiel(z string) func(*outbound.UpstreamAnfrage, *roherServer) {
	return func(a *outbound.UpstreamAnfrage, s *roherServer) {
		a.Ziel = strings.ReplaceAll(z, "{p}", fmt.Sprint(s.port()))
	}
}

func setzeMethode(m string) func(*outbound.UpstreamAnfrage, *roherServer) {
	return func(a *outbound.UpstreamAnfrage, _ *roherServer) { a.Methode = m }
}

func setzePfad(p string) func(*outbound.UpstreamAnfrage, *roherServer) {
	return func(a *outbound.UpstreamAnfrage, _ *roherServer) { a.Pfad = p }
}

func setzeQuery(q string) func(*outbound.UpstreamAnfrage, *roherServer) {
	return func(a *outbound.UpstreamAnfrage, _ *roherServer) { a.RawQuery = q }
}

func setzeHeader(name string, werte ...string) func(*outbound.UpstreamAnfrage, *roherServer) {
	return func(a *outbound.UpstreamAnfrage, _ *roherServer) { a.Header[name] = werte }
}

func argumentFaelle() []fehlerFall {
	bg := context.Background()
	var ohne context.Context
	beendet, cancel := context.WithCancel(bg)
	cancel()
	f := []fehlerFall{
		argumentFall(ohne, "ctx nil", func(*outbound.UpstreamAnfrage, *roherServer) {}),
		argumentFall(beendet, "beendeter ctx mit ungültigem Ziel", setzeZiel("/x")),
		argumentFall(bg, "Laenge -2", func(a *outbound.UpstreamAnfrage, _ *roherServer) { a.Laenge = -2 }),
		argumentFall(bg, "Body nil bei Laenge 5", func(a *outbound.UpstreamAnfrage, _ *roherServer) { a.Body, a.Laenge = nil, 5 }),
		argumentFall(bg, "Body nil bei Laenge -1", func(a *outbound.UpstreamAnfrage, _ *roherServer) { a.Body, a.Laenge = nil, -1 }),
		argumentFall(bg, "Header-Name mit Leerzeichen", setzeHeader("Bad "+geheim, "v")),
		argumentFall(bg, "Header-Name leer", setzeHeader("", "v")),
		argumentFall(bg, "Header-Name mit Klammer", setzeHeader("Bad("+geheim, "v")),
	}
	for _, z := range []string{"", "/x", "ftp://127.0.0.1:{p}", "http://", "http://u:" + geheim + "-z@127.0.0.1:{p}",
		"http://127.0.0.1:{p}?q", "http://127.0.0.1:{p}#f", "http://127.0.0.1:{p}?", "http://127.0.0.1:{p}#",
		"http://@127.0.0.1:{p}", "http://127.0.0.1:", "http://127.0.0.1:0",
		"http://127.0.0.1:65536", "http://127.0.0.1:{p}/a b", "http://127.0.0.1:{p}/%zz",
		`http://a"b:1`, "http://a<b:1", "http://a>b:1", "http://a]b:1", "http://[fe80::1%25%3C]:1"} {
		f = append(f, argumentFall(bg, "Ziel "+z, setzeZiel(z)))
	}
	for _, m := range []string{"", "GE T", "GET\n", "GE(T", "CONNECT"} {
		f = append(f, argumentFall(bg, fmt.Sprintf("Methode %q", m), setzeMethode(m)))
	}
	for _, p := range []string{"", geheim + "-x", "/" + geheim + "?b", "/" + geheim + "#b", "/" + geheim + " b", "/" + geheim + "\x7f"} {
		f = append(f, argumentFall(bg, fmt.Sprintf("Pfad %q", p), setzePfad(p)))
	}
	for _, q := range []string{geheim + " b", geheim + "#b", geheim + "\x00"} {
		f = append(f, argumentFall(bg, fmt.Sprintf("RawQuery %q", q), setzeQuery(q)))
	}
	for _, b := range []byte{0x00, 0x01, 0x08, 0x0A, 0x0B, 0x0D, 0x1F, 0x7F} {
		f = append(f, argumentFall(bg, fmt.Sprintf("Header-Wert mit 0x%02X", b), setzeHeader("X-Wert", geheim+"-w"+string([]byte{b})+"x")))
	}
	return f
}

// tokenZeichen enthält jedes Sonderzeichen eines HTTP-Tokens (RFC 9110 §5.6.2), Ziffern und
// Buchstaben an beiden Enden ihrer Bereiche.
const tokenZeichen = "!#$%&'*+-.^_`|~09azAZ"

// argumentGrenzFaelle sind Ziele an der gültigen Seite einer Grenze: die Ports 1 und 65535; im Namen
// des Hosts das `]` einer IP-Adresse, eine Zone, jedes Sonderzeichen von reg-name mit Ziffern und
// Buchstaben an den Enden ihrer Bereiche, eine Prozent-Kodierung und Bytes ab 0x80 samt 0x80. Dort
// lauscht niemand, und ohne Netz scheitern Auflösung oder Verbindung sofort.
func argumentGrenzFaelle() []fehlerFall {
	var f []fehlerFall
	for _, z := range []string{"http://127.0.0.1:1", "http://127.0.0.1:65535", "http://[::1]:1", "http://[fe80::1%25x]:1",
		"http://09azAZ-._~!$&'()*+,;=%C3%A4:1", "http://a\x80\xc3\xa4b:1"} {
		f = append(f, fehlerFall{name: fmt.Sprintf("gültiges Ziel %q", z), art: outbound.ErrUpstreamNichtErreichbar, lauf: func(t *testing.T) error {
			c := neuerClient(t, zeitgrenze)
			_, err := sende(context.Background(), t, c, mitGeheim(anfrage(z, "GET", "")))
			return err
		}})
	}
	return f
}

// TestSendenArgumente: ungültige Argumente sind model.ErrUngueltig ohne Verbindung und ohne Lesen
// aus Body; Argument vor Kontext; die Ziele an der Grenze sind gültig (A2, A19, SPEC-120 Nr. 2).
func TestSendenArgumente(t *testing.T) {
	fahre(t, slices.Concat(argumentFaelle(), argumentGrenzFaelle()))
}

// TestSendenArgumenteGegenseite: die andere Seite jeder Grenze geht ohne model.ErrUngueltig zum
// Upstream (A2).
func TestSendenArgumenteGegenseite(t *testing.T) {
	faelle := []struct {
		name    string
		aendere func(a *outbound.UpstreamAnfrage)
		pruefe  func(t *testing.T, e empfangen)
	}{
		{"Methode connect", func(a *outbound.UpstreamAnfrage) { a.Methode = "connect" }, func(t *testing.T, e empfangen) {
			if !strings.HasPrefix(e.zeile, "connect / ") {
				t.Errorf("Zeile %q", e.zeile)
			}
		}},
		{"Body bei Laenge 0", func(a *outbound.UpstreamAnfrage) { a.Body = strings.NewReader("x") }, nil},
		{"leerer Header-Wert", func(a *outbound.UpstreamAnfrage) { a.Header = map[string][]string{"X-Leer": {""}} }, func(t *testing.T, e empfangen) {
			if w := e.werte("X-Leer"); !slices.Equal(w, []string{""}) {
				t.Errorf("X-Leer %q", w)
			}
		}},
		{"Pfad /", func(a *outbound.UpstreamAnfrage) { a.Pfad = "/" }, nil},
		{"Methode aus allen Token-Zeichen", func(a *outbound.UpstreamAnfrage) { a.Methode = tokenZeichen }, func(t *testing.T, e empfangen) {
			if !strings.HasPrefix(e.zeile, tokenZeichen+" / ") {
				t.Errorf("Zeile %q", e.zeile)
			}
		}},
		{"Header-Name aus allen Token-Zeichen", func(a *outbound.UpstreamAnfrage) {
			a.Header = map[string][]string{tokenZeichen: {"v"}}
		}, func(t *testing.T, e empfangen) {
			if w := e.werte(tokenZeichen); !slices.Equal(w, []string{"v"}) {
				t.Errorf("%s %q", tokenZeichen, w)
			}
		}},
	}
	for _, b := range []byte{0x09, 0x20, 0x7E, 0x80} {
		wert := "a" + string([]byte{b}) + "b"
		faelle = append(faelle, struct {
			name    string
			aendere func(a *outbound.UpstreamAnfrage)
			pruefe  func(t *testing.T, e empfangen)
		}{fmt.Sprintf("Header-Wert mit 0x%02X", b), func(a *outbound.UpstreamAnfrage) {
			a.Header = map[string][]string{"X-Wert": {wert}}
		}, func(t *testing.T, e empfangen) {
			if w := e.werte("X-Wert"); !slices.Equal(w, []string{wert}) {
				t.Errorf("X-Wert %q", w)
			}
		}})
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			an := make(chan empfangen, 1)
			s := starteServer(t, beantworte(an, antwortOK))
			c := neuerClient(t, zeitgrenze)
			a := anfrage(s.ziel(), "GET", "/")
			f.aendere(&a)
			ant, err := sende(context.Background(), t, c, a)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = liesAlles(t, ant.Body)
			if e := warte(t, an); f.pruefe != nil {
				f.pruefe(t, e)
			}
		})
	}
}

// --- A3: Request-Zeile ---------------------------------------------------------------------------

// TestRequestZeile: Methode, Pfad aus Ziel, Pfad und Query byte-gleich, `?` nur bei Query, absolute
// Form bei `//`; die erwartete Zeile steht je Fall ausgeschrieben, {p} ist der Port (A3, SPEC-120
// Nr. 3).
func TestRequestZeile(t *testing.T) {
	faelle := []struct{ zielPfad, methode, pfad, query, soll string }{
		{"", "GET", "/", "", "GET / HTTP/1.1"},
		{"/svc/v1", "GET", "/Products(1)", "", "GET /svc/v1/Products(1) HTTP/1.1"},
		{"/a%2fb", "GET", "/x", "", "GET /a%2fb/x HTTP/1.1"},
		{"/a{b", "GET", "/x", "", "GET /a{b/x HTTP/1.1"},
		{"", "GET", "/a%2Fb", "", "GET /a%2Fb HTTP/1.1"},
		{"", "GET", "/a%2fb", "", "GET /a%2fb HTTP/1.1"},
		{"", "GET", "/%24metadata", "", "GET /%24metadata HTTP/1.1"},
		{"", "GET", "/P('a%27b')", "", "GET /P('a%27b') HTTP/1.1"},
		{"", "GET", "/a{b}", "", "GET /a{b} HTTP/1.1"},
		{"", "GET", "/\xc3\xa4", "", "GET /\xc3\xa4 HTTP/1.1"},
		{"", "GET", "/x", "$top=1&$filter=a%20eq%20'x'", "GET /x?$top=1&$filter=a%20eq%20'x' HTTP/1.1"},
		{"", "GET", "/x", "a+b", "GET /x?a+b HTTP/1.1"},
		{"", "GET", "/x", "%zz", "GET /x?%zz HTTP/1.1"},
		{"", "GET", "/x", "a=%2F", "GET /x?a=%2F HTTP/1.1"},
		{"", "GET", "/x", "a=\xc3\xa4", "GET /x?a=\xc3\xa4 HTTP/1.1"},
		{"", "get", "/x", "", "get /x HTTP/1.1"},
		{"", "PATCH", "/x", "", "PATCH /x HTTP/1.1"},
		{"", "MERGE", "/x", "", "MERGE /x HTTP/1.1"},
		{"", "connect", "/x", "", "connect /x HTTP/1.1"},
		{"", "GET", "//x", "", "GET http://127.0.0.1:{p}//x HTTP/1.1"},
		{"", "GET", "//x", "q=1", "GET http://127.0.0.1:{p}//x?q=1 HTTP/1.1"},
		{"/", "GET", "/x", "", "GET http://127.0.0.1:{p}//x HTTP/1.1"},
		{"/", "GET", "/x", "q=1", "GET http://127.0.0.1:{p}//x?q=1 HTTP/1.1"},
	}
	an := make(chan empfangen, 1)
	s := starteServer(t, beantworte(an, antwortOK))
	c := neuerClient(t, zeitgrenze)
	for _, f := range faelle {
		t.Run(fmt.Sprintf("%q %q %q %q", f.zielPfad, f.methode, f.pfad, f.query), func(t *testing.T) {
			a := anfrage(s.ziel()+f.zielPfad, f.methode, f.pfad)
			a.RawQuery = f.query
			ant, err := sende(context.Background(), t, c, a)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = liesAlles(t, ant.Body)
			soll := strings.ReplaceAll(f.soll, "{p}", fmt.Sprint(s.port()))
			if e := warte(t, an); e.zeile != soll {
				t.Errorf("Request-Zeile %q, erwartet %q", e.zeile, soll)
			}
		})
	}
}

// --- A4: Header ----------------------------------------------------------------------------------

// jeName liefert die Kopfzeilen je Name in ihrer Schreibweise, die Werte in ihrer Reihenfolge.
func jeName(e empfangen) map[string][]string {
	m := map[string][]string{}
	for _, k := range e.kopf {
		m[k.name] = append(m[k.name], k.wert)
	}
	return m
}

// headerFall sendet GET ohne Body mit h und liefert die empfangenen Kopfzeilen je Name.
func headerFall(t *testing.T, h map[string][]string) (map[string][]string, string) {
	t.Helper()
	an := make(chan empfangen, 1)
	s := starteServer(t, beantworte(an, antwortOK))
	c := neuerClient(t, zeitgrenze)
	a := anfrage(s.ziel(), "GET", "/")
	a.Header = h
	ant, err := sende(context.Background(), t, c, a)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = liesAlles(t, ant.Body)
	return jeName(warte(t, an)), s.ln.Addr().String()
}

// TestHeaderGenau: genau die Header der Map in ihrer Schreibweise; Ausnahmen in jeder Schreibweise;
// User-Agent als eine Zeile mit dem ersten Wert; kein Zusatz (A4, SPEC-120 Nr. 3).
func TestHeaderGenau(t *testing.T) {
	t.Run("a", func(t *testing.T) {
		got, host := headerFall(t, map[string][]string{
			"Accept": {"a", "b"}, "X-Leer": {""}, "X-Ohne": {}, "x-klein": {"k"}, "X-Rand": {" r\t"}, "X-Raum": {" \t"},
			"Prefer": {"odata.maxpagesize=1"}, "Connection": {"X-Hop"}, "Host": {"fremd"}, "Content-Length": {"99"},
			"Transfer-Encoding": {"gzip"}, "Trailer": {"X-T"},
		})
		soll := map[string][]string{
			"Host": {host}, "Accept": {"a", "b"}, "X-Leer": {""}, "x-klein": {"k"}, "X-Rand": {"r"}, "X-Raum": {""},
			"Prefer": {"odata.maxpagesize=1"}, "Connection": {"X-Hop"},
		}
		if !reflect.DeepEqual(got, soll) {
			t.Errorf("Header %q, erwartet %q", got, soll)
		}
	})
	for _, kv := range [][2]string{{"host", "fremd"}, {"content-length", "99"}, {"TRANSFER-ENCODING", "gzip"}, {"trailer", "X-T"}} {
		t.Run("b "+kv[0], func(t *testing.T) {
			got, host := headerFall(t, map[string][]string{kv[0]: {kv[1]}})
			soll := map[string][]string{"Host": {host}}
			if !reflect.DeepEqual(got, soll) {
				t.Errorf("Header %q, erwartet %q", got, soll)
			}
		})
	}
	faelle := []struct {
		name string
		h    map[string][]string
		ua   []string
	}{
		{"zwei Werte", map[string][]string{"User-Agent": {"u1", "u2"}}, []string{"u1"}},
		{"klein", map[string][]string{"user-agent": {"u3"}}, []string{"u3"}},
		{"Leerraum am Rand", map[string][]string{"User-Agent": {" u4 "}}, []string{"u4"}},
		{"leer", map[string][]string{"User-Agent": {""}}, nil},
		{"nur Leerzeichen", map[string][]string{"User-Agent": {" "}}, nil},
		{"leere Liste", map[string][]string{"User-Agent": {}}, nil},
		{"leere Liste neben klein", map[string][]string{"User-Agent": {}, "user-agent": {"u5"}}, []string{"u5"}},
		{"drei Schreibweisen", map[string][]string{"USER-AGENT": {"a"}, "User-Agent": {"b"}, "user-agent": {"c"}}, []string{"a"}},
	}
	for _, f := range faelle {
		t.Run("c "+f.name, func(t *testing.T) {
			got, _ := headerFall(t, f.h)
			var ua []string
			for _, n := range slices.Sorted(maps.Keys(got)) {
				if strings.EqualFold(n, "User-Agent") {
					if n != "User-Agent" {
						t.Errorf("User-Agent unter dem Namen %q", n)
					}
					ua = append(ua, got[n]...)
				}
			}
			if !slices.Equal(ua, f.ua) {
				t.Errorf("User-Agent %q, erwartet %q", ua, f.ua)
			}
		})
	}
}

// --- A5: Rahmung ---------------------------------------------------------------------------------

// protokollBody liefert r und merkt sich je Read die Zahl der Bytes und den Fehler.
type protokollBody struct {
	r      io.Reader
	bytes  []int
	fehler []error
}

func (p *protokollBody) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.bytes = append(p.bytes, n)
	p.fehler = append(p.fehler, err)
	return n, err
}

// rahmenNamen liefert die Kopfzeilen der Rahmung in jeder Schreibweise.
func rahmenNamen(e empfangen) []string {
	var z []string
	for _, k := range e.kopf {
		if strings.EqualFold(k.name, "Content-Length") || strings.EqualFold(k.name, "Transfer-Encoding") {
			z = append(z, k.name+": "+k.wert)
		}
	}
	return z
}

// TestRahmung: Rahmung nach Laenge je Methode; Body bis io.EOF gelesen; bei 0 nicht gelesen; eine
// abweichende Länge ist ErrUpstreamNichtErreichbar (A5, SPEC-120 Nr. 4).
func TestRahmung(t *testing.T) {
	fahre(t, rahmungFehlerFaelle())
	faelle := []struct {
		name    string
		methode string
		laenge  int64
		body    string
		rahmen  []string
		daten   string
		stuecke bool
	}{
		{"POST -1", "POST", -1, "abc", []string{"Transfer-Encoding: chunked"}, "abc", true},
		{"GET -1 leer", "GET", -1, "", []string{"Transfer-Encoding: chunked"}, "", true},
		{"Laenge 1", "POST", 1, "a", []string{"Content-Length: 1"}, "a", false},
		{"Laenge 3", "POST", 3, "abc", []string{"Content-Length: 3"}, "abc", false},
		{"Laenge 0 GET", "GET", 0, "x", nil, "", false},
		{"Laenge 0 DELETE", "DELETE", 0, "x", nil, "", false},
		{"Laenge 0 post", "post", 0, "x", nil, "", false},
		{"Laenge 0 POST", "POST", 0, "x", []string{"Content-Length: 0"}, "", false},
		{"Laenge 0 PUT", "PUT", 0, "x", []string{"Content-Length: 0"}, "", false},
		{"Laenge 0 PATCH", "PATCH", 0, "x", []string{"Content-Length: 0"}, "", false},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			an := make(chan empfangen, 1)
			s := starteServer(t, beantworte(an, antwortOK))
			c := neuerClient(t, zeitgrenze)
			body := &protokollBody{r: strings.NewReader(f.body)}
			a := anfrage(s.ziel(), f.methode, "/")
			a.Body, a.Laenge = body, f.laenge
			ant, err := sende(context.Background(), t, c, a)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = liesAlles(t, ant.Body)
			e := warte(t, an)
			if r := rahmenNamen(e); !slices.Equal(r, f.rahmen) {
				t.Errorf("Rahmung %q, erwartet %q", r, f.rahmen)
			}
			if string(e.body) != f.daten {
				t.Errorf("Body %s", text(e.body))
			}
			if f.stuecke && (len(e.stuecke) == 0 || e.stuecke[len(e.stuecke)-1] != 0) {
				t.Errorf("Stücke %v", e.stuecke)
			}
			if f.laenge == 0 && len(body.fehler) != 0 {
				t.Errorf("Body bei Laenge 0 %d-mal gelesen", len(body.fehler))
			}
			if f.laenge == 3 {
				pruefeBisEOF(t, body)
			}
		})
	}
}

// pruefeBisEOF verlangt nach den drei Bytes einen weiteren Aufruf, der io.EOF lieferte.
func pruefeBisEOF(t *testing.T, body *protokollBody) {
	t.Helper()
	summe := 0
	for i, n := range body.bytes {
		if summe == 3 {
			if !errors.Is(body.fehler[i], io.EOF) || n != 0 {
				t.Errorf("Aufruf nach den 3 Bytes: %d, %v", n, body.fehler[i])
			}
			return
		}
		summe += n
	}
	t.Errorf("kein Aufruf nach den 3 Bytes: %v %v", body.bytes, body.fehler)
}

// rahmungFehlerFaelle: weicht die Länge von Laenge ab, ist das ErrUpstreamNichtErreichbar; der
// Server liest Kopf und drei Bytes und antwortet nie.
func rahmungFehlerFaelle() []fehlerFall {
	var f []fehlerFall
	for _, body := range []string{"ab", "abcd"} {
		f = append(f, fehlerFall{name: "Laenge 3 mit " + body, art: outbound.ErrUpstreamNichtErreichbar, lauf: func(t *testing.T) error {
			s := starteServer(t, func(_ net.Conn, r *bufio.Reader) {
				if _, err := leseKopf(r); err == nil {
					_, _ = io.ReadFull(r, make([]byte, 3))
					_, _ = io.Copy(io.Discard, r)
				}
			})
			c := neuerClient(t, zeitgrenze)
			a := mitGeheim(anfrage(s.ziel(), "PUT", ""))
			a.Body, a.Laenge = strings.NewReader(body), 3
			_, err := sende(context.Background(), t, c, a)
			return err
		}})
	}
	return f
}

// --- A6: Strom -----------------------------------------------------------------------------------

// leseStueck liest ein Stück von chunked; das letzte ist leer.
func leseStueck(r *bufio.Reader) (string, error) {
	z, err := leseZeile(r)
	if err != nil {
		return "", err
	}
	var n int
	if _, err := fmt.Sscanf(z, "%x", &n); err != nil {
		return "", err
	}
	b := make([]byte, n+2)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b[:n]), nil
}

// TestStreaming: der Upstream erhält ein Stück des Request-Bodys, bevor das nächste geschrieben
// ist, und Senden kehrt zurück, bevor die Antwort ganz da ist (A6, SPEC-120 Nr. 4, 5).
func TestStreaming(t *testing.T) {
	teil1Da := make(chan struct{})
	weiter, frei := freigabe()
	fehler := make(chan error, 1)
	s := starteServer(t, func(c net.Conn, r *bufio.Reader) {
		fehler <- func() error {
			if _, err := leseKopf(r); err != nil {
				return err
			}
			if st, err := leseStueck(r); err != nil || st != "teil1" {
				return fmt.Errorf("erstes Stück %q %v", st, err)
			}
			close(teil1Da)
			for _, soll := range []string{"teil2", ""} {
				if st, err := leseStueck(r); err != nil || st != soll {
					return fmt.Errorf("Stück %q %v", st, err)
				}
			}
			_, _ = io.WriteString(c, "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nteil1\r\n")
			<-weiter
			_, err := io.WriteString(c, "5\r\nteil2\r\n0\r\n\r\n")
			return err
		}()
	})
	t.Cleanup(frei)
	c := neuerClient(t, zeitgrenze)
	pr, pw := io.Pipe()
	a := anfrage(s.ziel(), "POST", "/")
	a.Body, a.Laenge = pr, -1
	ch := sendeNebenbei(context.Background(), c, a)
	if _, err := pw.Write([]byte("teil1")); err != nil {
		t.Fatal(err)
	}
	warte(t, teil1Da)
	_, _ = pw.Write([]byte("teil2"))
	_ = pw.Close()
	e := warte(t, ch)
	if e.err != nil {
		t.Fatal(e.err, warte(t, fehler))
	}
	p := make([]byte, 16)
	if n, err := e.ant.Body.Read(p); err != nil || string(p[:n]) != "teil1" {
		t.Fatalf("erstes Read %q %v", p[:n], err)
	}
	frei()
	rest, err := liesAlles(t, e.ant.Body)
	if err != nil || string(rest) != "teil2" {
		t.Errorf("Rest %q %v", rest, err)
	}
	if err := warte(t, fehler); err != nil {
		t.Error(err)
	}
}

// TestKeinVorrat: der Upstream erhält ein Stück, bevor das nächste gelesen ist. In der Blase liest
// der Server nur den Kopf und lässt die Pipe offen; das erste Stück bleibt im Schreiben der
// Bibliothek hängen, und Body wurde genau einmal gelesen (A6, SPEC-120 Nr. 4).
func TestKeinVorrat(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c, err := httpupstream.New(httpupstream.Konfig{Frist: time.Hour, Pruefen: true})
		if err != nil {
			t.Fatal(err)
		}
		cli, srv := net.Pipe()
		tr := httpupstream.Transport(c)
		tr.DialContext = func(context.Context, string, string) (net.Conn, error) { return cli, nil }
		kopf := make(chan error, 1)
		go func() {
			_, err := http.ReadRequest(bufio.NewReader(srv))
			kopf <- err
		}()
		ctx, cancel := context.WithCancel(context.Background())
		body := &stueckBody{stuecke: []stueckWert{{daten: "teil1"}, {daten: "teil2"}}, ende: io.EOF}
		a := anfrage(zielBlase, "POST", "/")
		a.Body, a.Laenge = body, -1
		ch := sendeNebenbei(ctx, c, a)
		if err := <-kopf; err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		if n := body.aufrufe.Load(); n != 1 {
			t.Errorf("%d Aufrufe von Read, bevor der Upstream das erste Stück las", n)
		}
		cancel()
		<-ch
		tr.CloseIdleConnections()
		_ = cli.Close()
		_ = srv.Close()
	})
}

// --- A17: nacheinander ---------------------------------------------------------------------------

// TestNacheinander: zwei Aufrufe nacheinander auf einem Client teilen keinen Wert; der Server
// nimmt mehrere Requests je Verbindung und weitere Verbindungen, die Antworten zählen über alle
// (A17, SPEC-120 Nr. 9).
func TestNacheinander(t *testing.T) {
	var nr atomic.Int32
	an := make(chan empfangen, 2)
	s := starteServer(t, func(c net.Conn, r *bufio.Reader) {
		for {
			e, err := leseAnfrage(r)
			if err != nil {
				return
			}
			an <- e
			body := fmt.Sprintf("antwort-%d", nr.Add(1))
			if _, err := fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\n%s", len(body), body); err != nil {
				return
			}
		}
	})
	c := neuerClient(t, zeitgrenze)
	a := anfrage(s.ziel(), "POST", "/eins")
	a.Header = map[string][]string{"X-Eins": {"1"}}
	a.Body, a.Laenge = strings.NewReader("abc"), 3
	ant, err := sende(context.Background(), t, c, a)
	if err != nil {
		t.Fatal(err)
	}
	if b, err := liesAlles(t, ant.Body); err != nil || string(b) != "antwort-1" {
		t.Fatalf("erste Antwort %q %v", b, err)
	}
	warte(t, an)
	ant, err = sende(context.Background(), t, c, anfrage(s.ziel(), "GET", "/zwei"))
	if err != nil {
		t.Fatal(err)
	}
	if b, err := liesAlles(t, ant.Body); err != nil || string(b) != "antwort-2" {
		t.Errorf("zweite Antwort %q %v", b, err)
	}
	if e := warte(t, an); !slices.Equal(e.zeilen(), []string{"Host: " + s.ln.Addr().String()}) || len(e.body) != 0 {
		t.Errorf("zweiter Request %q, Body %s", e.zeilen(), text(e.body))
	}
}
