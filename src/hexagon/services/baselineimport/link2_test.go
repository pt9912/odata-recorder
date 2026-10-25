package baselineimport_test

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// linkOrakel löst ref gegen upstream, den OData-Pfad und die Query der Seite auf (SPEC-127 Nr. 12 e):
// net/url (Parse, ResolveReference) liefert die Auflösung nach RFC 3986 §5.3, eine zweite, hier
// geschriebene Regelfunktion die Regeln von Nr. 12 e darüber (Bytes unter 0x21 und 0x7F, Schema, Host
// und Port, Präfix des Pfades an der Grenze `/`, Fehler des Keys). kopf sind die Header der Vorgängerseite.
// ok ist falsch, wenn der Link nicht auflösbar ist.
func linkOrakel(upstream, seitenPfad, seitenQuery string, kopf map[string][]string, ref string) (key string, ok bool) {
	for i := 0; i < len(ref); i++ {
		if ref[i] <= 0x20 || ref[i] == 0x7f {
			return "", false
		}
	}
	up, err := url.Parse(upstream)
	if err != nil {
		return "", false
	}
	upPfad := strings.TrimSuffix(up.EscapedPath(), "/")
	base, err := url.Parse(up.Scheme + "://" + up.Host + upPfad + seitenPfad)
	if err != nil {
		return "", false
	}
	base.RawQuery = seitenQuery
	r, err := url.Parse(ref)
	if err != nil {
		return "", false
	}
	// net/url behandelt `//` mit leerer Autorität wie einen Pfad; RFC 3986 und Nr. 12 e kennen sie als Autorität ohne Host.
	if r.Scheme == "" && strings.HasPrefix(ref, "//") && r.Hostname() == "" {
		return "", false
	}
	t := base.ResolveReference(r)
	if s := strings.ToLower(t.Scheme); s != "http" && s != "https" {
		return "", false
	}
	hostPort := func(u *url.URL) (string, int, bool) {
		h := strings.ToLower(u.Hostname())
		if h == "" {
			return "", 0, false
		}
		p := 80
		if strings.ToLower(u.Scheme) == "https" {
			p = 443
		}
		if s := u.Port(); s != "" {
			n, err := strconv.Atoi(s)
			if err != nil || n < 0 || n > 65535 {
				return "", 0, false
			}
			p = n
		}
		return h, p, true
	}
	h1, p1, ok1 := hostPort(t)
	h2, p2, ok2 := hostPort(up)
	if !ok1 || !ok2 || h1 != h2 || p1 != p2 {
		return "", false
	}
	ep := t.EscapedPath()
	if strings.IndexFunc(t.Path, func(r rune) bool { return r >= 0x80 }) >= 0 {
		ep = t.Path // Bytes ab 0x80 bleiben, wie sie stehen (Nr. 12 e); net/url schreibt sie als %XX
	}
	rest := ""
	switch {
	case ep == upPfad:
		rest = "/"
	case strings.HasPrefix(ep, upPfad+"/"):
		rest = ep[len(upPfad):]
	default:
		return "", false
	}
	k, err := model.BildeSchluessel(model.SchluesselEingabe{
		Method: "GET", ServiceID: dienstID, Path: rest, RawQuery: t.RawQuery, Header: kopf, RequestRedaction: model.RedactionNone,
	}, regeln())
	if err != nil {
		return "", false
	}
	return k.MatchKey, true
}

// linkElem ist ein Element der Menge von H6: Upstream-URL, Link der ersten Seite und die Klasse.
type linkElem struct {
	klasse   string
	upstream string
	ref      string
}

// h6B erzeugt die Klasse (B) für die Upstream-URL up: Pfad, Query und Fragment bei gleicher Autorität (464).
func h6B(up string) []linkElem {
	var out []linkElem
	bpfade := []string{
		"/odata/v4/Products", "/odata/v4", "/odata/v4/", "/odata/v42/Products", "/odata/Products", "/ODATA/v4/Products", "/odata/v4/Prod%75cts",
		"/odata/v4/a%2Fb", "/odata/v4/./Products", "/odata/v4/x/../Products", "/odata/v4/Products/.", "/odata/v4/Products/..",
		"/odata/v4/../v4/Products", "/../odata/v4/Products", "/odata/./v4/./Products", "",
	}
	queries := []string{"", "?", "?$skiptoken=1", "?a=%zz"}
	fragmente := []string{"", "#f"}
	for _, pf := range bpfade {
		for _, q := range queries {
			for _, f := range fragmente {
				out = append(out, linkElem{"B absolut", up, "https://up.example" + pf + q + f}, linkElem{"B Netzpfad", up, "//up.example" + pf + q + f})
				if pf != "" {
					out = append(out, linkElem{"B pfadabsolut", up, pf + q + f})
				}
			}
		}
	}
	for _, pf := range []string{"Products", "./Products", "../Products", "../v4/Products", "../../odata/v4/Products", ".", "..", "Products/.", "a/../Products", "./../Products"} {
		for _, q := range queries {
			for _, f := range fragmente {
				out = append(out, linkElem{"B pfadrelativ", up, pf + q + f})
			}
		}
	}
	for _, q := range []string{"?", "?$skiptoken=1", "?a=%zz"} {
		for _, f := range fragmente {
			out = append(out, linkElem{"B nur Query", up, q + f})
		}
	}
	return append(out, linkElem{"B leer", up, ""}, linkElem{"B nur Fragment", up, "#f"})
}

// h6Elemente erzeugt das Produkt der Komponentenklassen (Plan §3, Zeile H6): (A) Autorität 3840, (B) Pfad,
// Query und Fragment 464, (C) Bytes 179, (D) Upstream-Formen 48; Referenz-Upstream `https://up.example/odata/v4`.
func h6Elemente() []linkElem {
	const ref = "https://up.example/odata/v4"
	var out []linkElem
	// (A) Autorität: absolut mit Schema (3200) und als Netzpfad (640).
	schemata := []string{"http", "https", "HTTP", "Https", "ftp"}
	zugang := []string{"", "u@", "u:p@", "@"}
	hosts := []string{"up.example", "UP.EXAMPLE", "other.example", "up.example.", "", "[::1]", "[::1", "[::1]x"}
	ports := []string{"", ":443", ":80", ":8443", ":", ":0", ":0443", ":+443", ":x", ":65536"}
	pfade := []string{"/odata/v4/Products", "/odata/v4"}
	for _, z := range zugang {
		for _, h := range hosts {
			for _, p := range ports {
				for _, pf := range pfade {
					for _, s := range schemata {
						out = append(out, linkElem{"A absolut", ref, s + "://" + z + h + p + pf})
					}
					out = append(out, linkElem{"A Netzpfad", ref, "//" + z + h + p + pf})
				}
			}
		}
	}
	out = append(out, h6B(ref)...)
	// (C) Bytes: 0x00 bis 0x20, 0x7F, 0x80, ä (zwei Bytes) und 0xFF an Pfad, Query und Fragment, dazu die 34 Bytes unter 0x21 und 0x7F an Zugangsdaten und Host.
	var klasse []string
	for c := 0; c <= 0x20; c++ {
		klasse = append(klasse, string([]byte{byte(c)}))
	}
	klasse = append(klasse, "\x7f", "\x80", "\xc3\xa4", "\xff")
	for _, b := range klasse {
		out = append(out,
			linkElem{"C Pfad", ref, "Products" + b},
			linkElem{"C Query", ref, "Products?$skiptoken=1" + b},
			linkElem{"C Fragment", ref, "Products?$skiptoken=1#f" + b})
		if len(b) == 1 && (b[0] <= 0x20 || b[0] == 0x7f) {
			out = append(out,
				linkElem{"C Zugangsdaten", ref, "https://u" + b + "@up.example/odata/v4/Products?$skiptoken=1"},
				linkElem{"C Host", ref, "https://up" + b + ".example/odata/v4/Products?$skiptoken=1"})
		}
	}
	// (D) Upstream-Formen mal sechs Links.
	for _, up := range []string{
		"https://up.example/odata/v4", "https://up.example", "https://up.example/odata/v4/", "https://up.example:8443/odata/v4",
		"http://up.example/odata/v4", "https://[::1]/odata/v4", "https://up.example/a%2Fb", "https://up.example/a%2Fb/",
	} {
		u, err := url.Parse(up)
		if err != nil {
			panic(err)
		}
		pfx := strings.TrimSuffix(u.EscapedPath(), "/")
		name := u.Hostname()
		if strings.Contains(name, ":") {
			name = "[" + name + "]"
		}
		standard := map[string]string{"http": "80", "https": "443"}[u.Scheme]
		const ziel = "/Products?$skiptoken=1"
		for _, l := range []string{
			u.Scheme + "://" + u.Host + pfx + ziel, u.Scheme + "://" + name + ":" + standard + pfx + ziel, pfx + ziel,
			strings.TrimPrefix(pfx, "/") + "/Products?$skiptoken=1", "//" + u.Host + pfx + ziel, u.Scheme + "://other.example" + pfx + ziel,
		} {
			out = append(out, linkElem{"D Upstream", up, l})
		}
	}
	return out
}

// pruefeLink importiert eine Session, deren erste Seite von Products den Link e.ref nennt, und vergleicht das Ende der
// Kette mit dem Orakel linkOrakel (aufgelöst oder nicht, und der Key der nächsten Seite); es liefert das erwartete Ende.
func pruefeLink(t *testing.T, e linkElem) string {
	t.Helper()
	w := neu(t)
	w.konfig.Dienste = []baselineimport.Dienst{{ID: dienstID, UpstreamURL: e.upstream}}
	b := w.standard()
	start := b.collection("Products", "", mitLink(seite(produkt(1, "P1", "1")), e.ref))
	key, ok := linkOrakel(e.upstream, "/Products", "", nil, e.ref)
	erwartet := "link_nicht_aufloesbar"
	switch {
	case ok && key == start.MatchKey:
		erwartet = "schleife"
	case ok:
		erwartet = "proven"
		b.collection("Products", "x", seite(produkt(2, "P2", "1")), func(x *model.Exchange) { x.MatchKey = key })
	}
	mustImport(t, w, anfrage())
	status, grund, ketten := ketteVon(t, w)
	if got := endeVon(ketten[0]); got != erwartet {
		t.Errorf("%s %q (Upstream %s): Ende %s, Orakel %s", e.klasse, e.ref, e.upstream, got, erwartet)
	}
	if (erwartet == "proven") != (status == "complete") || (erwartet == "proven") != (grund == "kette_belegt") {
		t.Errorf("%s %q: %s/%s, Orakel %s", e.klasse, e.ref, status, grund, erwartet)
	}
	return erwartet
}

// TestZeileH6OhnePfad: die Klasse (B) gegen Upstream-URLs ohne Pfad (`https://up.example` und mit abschließendem `/`),
// außerhalb der 4531 Elemente: bei einem leeren Pfad der Upstream-URL sind Links aufgelöst, deren Autorität von `?` oder
// `#` begrenzt wird und die keinen Pfad nennen (2 mal 464).
func TestZeileH6OhnePfad(t *testing.T) {
	enden := map[string]int{}
	for _, up := range []string{"https://up.example", "https://up.example/"} {
		elemente := h6B(up)
		if len(elemente) != 464 {
			t.Fatalf("Klasse B hat %d Elemente", len(elemente))
		}
		for _, e := range elemente {
			enden[pruefeLink(t, e)]++
		}
	}
	// Ohne Pfad der Upstream-URL liegt jeder Link auf demselben Host im Präfix: aufgelöst oder die Seite selbst.
	if enden["proven"] == 0 || enden["schleife"] == 0 || len(enden) != 2 {
		t.Errorf("Ausgänge %v", enden)
	}
	t.Logf("H6 ohne Pfad: %v", enden)
}

// TestZeileH6Linkformen: das Produkt der Komponentenklassen (4531 Elemente, der Test zählt sie) gegen das
// Orakel linkOrakel: aufgelöst oder nicht, und der Key der Seite, auf die der Link zeigt.
func TestZeileH6Linkformen(t *testing.T) {
	elemente := h6Elemente()
	zaehl := map[string]int{}
	for _, e := range elemente {
		zaehl[e.klasse[:1]]++
	}
	if zaehl["A"] != 3840 || zaehl["B"] != 464 || zaehl["C"] != 179 || zaehl["D"] != 48 || len(elemente) != 4531 {
		t.Fatalf("Menge %v, gesamt %d, erwartet A 3840, B 464, C 179, D 48, gesamt 4531", zaehl, len(elemente))
	}
	enden := map[string]int{}
	for _, e := range elemente {
		enden[fmt.Sprint(e.klasse[:1], " ", pruefeLink(t, e))]++
	}
	// Jede Klasse erzeugt aufgelöste und nicht aufgelöste Links (das Orakel ist nicht trivial).
	for _, k := range []string{"A", "B", "C", "D"} {
		if enden[k+" proven"] == 0 || enden[k+" link_nicht_aufloesbar"] == 0 {
			t.Errorf("Klasse %s: %v", k, enden)
		}
	}
	t.Logf("H6 Linkformen: %v", enden)
}

// TestZeileH6PortGrenze: beide Seiten der Grenze des Ports (SPEC-127 Nr. 12 e: Dezimalzahl von 0 bis 65535), außerhalb der
// 4531 Elemente: Upstream-Port {keiner, 443, 65535, 0} mal Port des Links {keiner, 443, 65535, 65536, 0, 00, 0065535} (28),
// Orakel linkOrakel.
func TestZeileH6PortGrenze(t *testing.T) {
	n, aufgeloest := 0, 0
	for _, upPort := range []string{"", ":443", ":65535", ":0"} {
		up := "https://up.example" + upPort + "/odata/v4"
		for _, linkPort := range []string{"", ":443", ":65535", ":65536", ":0", ":00", ":0065535"} {
			n++
			ref := "https://up.example" + linkPort + "/odata/v4/Products?$skiptoken=1"
			w := neu(t)
			w.konfig.Dienste = []baselineimport.Dienst{{ID: dienstID, UpstreamURL: up}}
			b := w.standard()
			b.collection("Products", "", mitLink(seite(produkt(1, "P1", "1")), ref))
			key, ok := linkOrakel(up, "/Products", "", nil, ref)
			erwartet := "link_nicht_aufloesbar"
			if ok {
				aufgeloest++
				erwartet = "proven"
				b.collection("Products", "x", seite(produkt(2, "P2", "1")), func(x *model.Exchange) { x.MatchKey = key })
			}
			mustImport(t, w, anfrage())
			if _, _, ketten := ketteVon(t, w); endeVon(ketten[0]) != erwartet {
				t.Errorf("Upstream %s, Link %s: Ende %s, Orakel %s", up, ref, endeVon(ketten[0]), erwartet)
			}
		}
	}
	if n != 28 || aufgeloest < 8 {
		t.Errorf("Menge %d, davon aufgelöst %d", n, aufgeloest)
	}
}
