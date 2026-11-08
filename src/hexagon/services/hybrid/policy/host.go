package policy

import (
	"net/netip"
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
