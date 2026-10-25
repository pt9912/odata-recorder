package baselineimport

import (
	"strconv"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// uri ist eine URI-Referenz, zerlegt nach RFC 3986 Anhang B.
type uri struct {
	schema     string
	hatSchema  bool
	autoritaet string
	hatAut     bool
	pfad       string
	query      string
	hatQuery   bool
}

// zerlege teilt s in die fünf Bestandteile; false heißt, s trägt ein Steuerzeichen oder ein Schema
// ungültiger Form.
func zerlege(s string) (uri, bool) {
	var u uri
	for i := 0; i < len(s); i++ {
		if s[i] <= 0x20 || s[i] == 0x7f {
			return uri{}, false
		}
	}
	rest, _, _ := strings.Cut(s, "#")
	if i := strings.IndexAny(rest, ":/?"); i > 0 && rest[i] == ':' {
		if !schemaGueltig(rest[:i]) {
			return uri{}, false
		}
		u.schema, u.hatSchema = rest[:i], true
		rest = rest[i+1:]
	}
	if r, ok := strings.CutPrefix(rest, "//"); ok {
		i := strings.IndexAny(r, "/?")
		if i < 0 {
			i = len(r)
		}
		u.autoritaet, u.hatAut = r[:i], true
		rest = r[i:]
	}
	u.pfad, u.query, u.hatQuery = strings.Cut(rest, "?")
	return u, true
}

func schemaGueltig(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		alpha := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		if !alpha && (i == 0 || (c < '0' || c > '9') && c != '+' && c != '-' && c != '.') {
			return false
		}
	}
	return s != ""
}

// entferneDotSegmente ist remove_dot_segments aus RFC 3986 §5.2.4.
func entferneDotSegmente(p string) string {
	var aus []string
	for p != "" {
		switch {
		case strings.HasPrefix(p, "../"):
			p = p[3:]
		case strings.HasPrefix(p, "./"):
			p = p[2:]
		case strings.HasPrefix(p, "/./"):
			p = p[2:]
		case p == "/.":
			p = "/"
		case strings.HasPrefix(p, "/../"):
			p = p[3:]
			aus = entferneLetztes(aus)
		case p == "/..":
			p = "/"
			aus = entferneLetztes(aus)
		case p == "." || p == "..":
			p = ""
		default:
			i := strings.IndexByte(p[1:], '/')
			if i < 0 {
				i = len(p) - 1
			}
			aus = append(aus, p[:i+1])
			p = p[i+1:]
		}
	}
	return strings.Join(aus, "")
}

func entferneLetztes(s []string) []string {
	if len(s) == 0 {
		return s
	}
	return s[:len(s)-1]
}

// loeseAuf wendet RFC 3986 §5.2.2 an: das Ergebnis trägt Schema und Autorität der Basis, wenn die
// Referenz keine nennt.
func loeseAuf(basis, ref uri) uri {
	var t uri
	switch {
	case ref.hatSchema:
		t = ref
		t.pfad = entferneDotSegmente(ref.pfad)
	case ref.hatAut:
		t = ref
		t.schema, t.hatSchema = basis.schema, true
		t.pfad = entferneDotSegmente(ref.pfad)
	default:
		t = uri{schema: basis.schema, hatSchema: true, autoritaet: basis.autoritaet, hatAut: basis.hatAut}
		if ref.pfad == "" {
			t.pfad, t.query, t.hatQuery = basis.pfad, basis.query, basis.hatQuery
			if ref.hatQuery {
				t.query, t.hatQuery = ref.query, true
			}
		} else {
			if strings.HasPrefix(ref.pfad, "/") {
				t.pfad = entferneDotSegmente(ref.pfad)
			} else {
				t.pfad = entferneDotSegmente(verbinde(basis, ref.pfad))
			}
			t.query, t.hatQuery = ref.query, ref.hatQuery
		}
	}
	return t
}

// verbinde ist merge aus RFC 3986 §5.2.3.
func verbinde(basis uri, pfad string) string {
	if basis.hatAut && basis.pfad == "" {
		return "/" + pfad
	}
	i := strings.LastIndexByte(basis.pfad, '/')
	return basis.pfad[:i+1] + pfad
}

// hostPort liefert Host (ASCII-klein) und Port einer Autorität; der Port ist der genannte oder der
// Standard des Schemas.
func hostPort(autoritaet, schema string) (string, int, bool) {
	a := autoritaet
	if i := strings.LastIndexByte(a, '@'); i >= 0 {
		a = a[i+1:]
	}
	host, port := a, ""
	if strings.HasPrefix(a, "[") {
		i := strings.IndexByte(a, ']')
		if i < 0 {
			return "", 0, false
		}
		host, port = a[:i+1], a[i+1:]
		if port != "" && port[0] != ':' {
			return "", 0, false
		}
		port = strings.TrimPrefix(port, ":")
	} else if i := strings.LastIndexByte(a, ':'); i >= 0 {
		host, port = a[:i], a[i+1:]
	}
	if host == "" {
		return "", 0, false
	}
	p := 80
	if asciiKlein(schema) == "https" {
		p = 443
	}
	if port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 0 || n > 65535 || port[0] == '+' || port[0] == '-' {
			return "", 0, false
		}
		p = n
	}
	return asciiKlein(host), p, true
}

// schluesselDesLinks löst den @odata.nextLink der Seite a auf und bildet den Match-Key der Seite, auf
// die er zeigt (SPEC-127 Nr. 12 e); false heißt link_nicht_aufloesbar.
func schluesselDesLinks(d Dienst, regeln model.SchluesselRegeln, a *antwort) (string, bool) {
	up, ok := zerlege(d.UpstreamURL)
	if !ok {
		return "", false
	}
	base := uri{schema: up.schema, hatSchema: true, autoritaet: up.autoritaet, hatAut: true, pfad: up.pfad + a.ex.Path}
	if a.ex.RawQuery != "" {
		base.query, base.hatQuery = a.ex.RawQuery, true
	}
	ref, ok := zerlege(a.seite.FolgeLink.Wert)
	if !ok {
		return "", false
	}
	t := loeseAuf(base, ref)
	if s := asciiKlein(t.schema); s != "http" && s != "https" {
		return "", false
	}
	h1, p1, ok1 := hostPort(t.autoritaet, t.schema)
	h2, p2, ok2 := hostPort(up.autoritaet, up.schema)
	if !ok1 || !ok2 || h1 != h2 || p1 != p2 {
		return "", false
	}
	rest := ""
	switch {
	case t.pfad == up.pfad:
		rest = "/"
	case strings.HasPrefix(t.pfad, up.pfad+"/"):
		rest = t.pfad[len(up.pfad):]
	default:
		return "", false
	}
	k, err := model.BildeSchluessel(model.SchluesselEingabe{
		Method: "GET", ServiceID: d.ID, Path: rest, RawQuery: t.query, Header: a.ex.RequestHeaders, RequestRedaction: model.RedactionNone,
	}, regeln)
	if err != nil {
		return "", false
	}
	return k.MatchKey, true
}
