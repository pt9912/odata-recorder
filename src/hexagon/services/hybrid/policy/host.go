package policy

import (
	"net/netip"
	"net/url"
	"strings"
)

// hostEintrag ist ein geparster Eintrag von Hosts: name in ASCII-Kleinschreibung oder die
// kanonische IPv6-Adresse ohne Klammern, port 0 ohne Angabe (SPEC-261 Nr. 5 a, b).
type hostEintrag struct {
	name string
	port int
}

// parseHost prüft einen Eintrag von Hosts und liefert ihn geparst oder die Ursache (SPEC-261
// Nr. 5 a); der Text des Eintrags verlässt die Funktion nie.
func parseHost(e string) (hostEintrag, string) {
	if e == "" {
		return hostEintrag{}, UrsacheLeer
	}
	var name, portText string
	var hatPort, ok bool
	if e[0] == '[' {
		name, portText, hatPort, ok = zerlegeKlammer(e)
	} else {
		name, portText, hatPort, ok = zerlegeName(e)
	}
	if !ok {
		return hostEintrag{}, UrsacheHost
	}
	h := hostEintrag{name: name}
	if hatPort {
		port, ok := parsePort(portText)
		if !ok {
			return hostEintrag{}, UrsacheHost
		}
		h.port = port
	}
	return h, ""
}

// zerlegeKlammer liest `[` IPv6-Adresse ohne Zone `]` und optional `:` Port; der Name ist die
// kanonische Form der Adresse.
func zerlegeKlammer(e string) (name, portText string, hatPort, ok bool) {
	ende := strings.IndexByte(e, ']')
	if ende < 0 {
		return "", "", false, false
	}
	adr, err := netip.ParseAddr(e[1:ende])
	if err != nil || !adr.Is6() || adr.Zone() != "" {
		return "", "", false, false
	}
	name = adr.String()
	rest := e[ende+1:]
	switch {
	case rest == "":
		return name, "", false, true
	case rest[0] == ':':
		return name, rest[1:], true, true
	}
	return "", "", false, false
}

// zerlegeName liest einen Namen und optional `:` Port; der Name wird in ASCII-Kleinschreibung
// geliefert.
func zerlegeName(e string) (name, portText string, hatPort, ok bool) {
	name = e
	if i := strings.IndexByte(e, ':'); i >= 0 {
		name, portText, hatPort = e[:i], e[i+1:], true
	}
	if name == "" {
		return "", "", false, false
	}
	klein := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		c, gueltig := nameByte(name[i])
		if !gueltig {
			return "", "", false, false
		}
		klein[i] = c
	}
	return string(klein), portText, hatPort, true
}

// nameByte liefert das Byte eines Namens in Kleinschreibung; erlaubt sind ASCII-Buchstaben,
// Ziffern, `.`, `-` und `_`.
func nameByte(c byte) (byte, bool) {
	switch {
	case c >= 'A' && c <= 'Z':
		return c + 'a' - 'A', true
	case c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '.', c == '-', c == '_':
		return c, true
	}
	return 0, false
}

// zielHost ist der gelesene Host eines Ziels: Name (ASCII-Kleinschreibung oder kanonische
// IPv6-Adresse), Port und der Standardport des Schemas.
type zielHost struct {
	name           string
	port, standard int
}

// hostFrei sagt, ob der Host von ziel durch einen Eintrag von Hosts freigegeben ist (SPEC-261
// Nr. 5 c, d): ein Eintrag ohne Port meint den Standardport des Schemas von ziel.
func (p *Policy) hostFrei(ziel string) bool {
	z, ok := leseZiel(ziel)
	if !ok {
		return false
	}
	for _, e := range p.hosts {
		port := e.port
		if port == 0 {
			port = z.standard
		}
		if e.name == z.name && port == z.port {
			return true
		}
	}
	return false
}

// leseZiel liest ein Ziel nach SPEC-261 Nr. 5 c; ein nicht lesbares Ziel liefert false. Wie im
// Eintrag steht eine IPv6-Adresse in eckigen Klammern und ein Name ohne sie.
func leseZiel(ziel string) (zielHost, bool) {
	u, err := url.Parse(ziel)
	if err != nil || u.User != nil {
		return zielHost{}, false
	}
	var z zielHost
	switch u.Scheme {
	case "http":
		z.standard = 80
	case "https":
		z.standard = 443
	default:
		return zielHost{}, false
	}
	z.port = z.standard
	if t := u.Port(); t != "" {
		n, ok := zielPort(t)
		if !ok {
			return zielHost{}, false
		}
		z.port = n
	}
	name, ok := zielName(u.Host, u.Hostname())
	z.name = name
	return z, ok
}

// zielName liest den Namen eines Ziels aus dem Host mit Port und Klammern (roh) und dem Hostnamen
// ohne sie: einen Namen in ASCII-Kleinschreibung oder die kanonische IPv6-Adresse ohne Zone.
func zielName(roh, host string) (string, bool) {
	if strings.HasPrefix(roh, "[") != strings.Contains(host, ":") {
		return "", false
	}
	if strings.Contains(host, ":") {
		adr, err := netip.ParseAddr(host)
		if err != nil || !adr.Is6() || adr.Zone() != "" {
			return "", false
		}
		return adr.String(), true
	}
	if host == "" {
		return "", false
	}
	klein := make([]byte, len(host))
	for i := 0; i < len(host); i++ {
		c, gueltig := nameByte(host[i])
		if !gueltig {
			return "", false
		}
		klein[i] = c
	}
	return string(klein), true
}

// zielPort liest den Port eines Ziels: eine Dezimalzahl von 1 bis 65535, auch mit führender Null.
func zielPort(s string) (int, bool) {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
		if n > 65535 {
			return 0, false
		}
	}
	return n, n >= 1
}

// parsePort liest Dezimalziffern ohne führende Null im Bereich 1 bis 65535.
func parsePort(s string) (int, bool) {
	if s == "" || len(s) > 5 || s[0] == '0' {
		return 0, false
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, n <= 65535
}
