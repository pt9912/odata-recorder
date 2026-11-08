package policy_test

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// Das Orakel für einen Eintrag von Hosts ist unabhängig von parseHost gebaut: reguläre Ausdrücke
// und net.ParseIP statt netip.
var (
	hostName = regexp.MustCompile(`^([A-Za-z0-9._-]+)(?::([0-9]+))?$`)
	hostV6   = regexp.MustCompile(`^\[([0-9A-Fa-f:.]+)\](?::([0-9]+))?$`)
)

// orakelHost liefert Name (Kleinschreibung bzw. Text der Adresse), Port (0 ohne) und die Ursache.
func orakelHost(e string) (name string, port int, ursache string) {
	if e == "" {
		return "", 0, policy.UrsacheLeer
	}
	var portText string
	if m := hostName.FindStringSubmatch(e); m != nil {
		name, portText = strings.ToLower(m[1]), m[2]
	} else if m := hostV6.FindStringSubmatch(e); m != nil {
		if ip := net.ParseIP(m[1]); ip == nil || !strings.Contains(m[1], ":") {
			return "", 0, policy.UrsacheHost
		}
		name, portText = m[1], m[2]
	} else {
		return "", 0, policy.UrsacheHost
	}
	if portText != "" {
		n, err := strconv.Atoi(portText)
		if err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != portText {
			return "", 0, policy.UrsacheHost
		}
		port = n
	}
	return name, port, ""
}

type hostGut struct {
	eintrag, name string
	port          int
}

type eintragUrsache struct{ eintrag, ursache string }

func TestHostEintraege(t *testing.T) {
	plan := []hostGut{
		{"h", "h", 0}, {"h.example", "h.example", 0}, {"h-1_x", "h-1_x", 0},
		{"h:1", "h", 1}, {"h:65535", "h", 65535},
		{"[::1]", "::1", 0}, {"[::1]:8080", "::1", 8080}, {"[2001:db8::1]", "2001:db8::1", 0},
		{"H", "h", 0},
	}
	// Schreibweise: Name klein, IPv6 kanonisch (SPEC-261 Nr. 5 b)
	zusatz := []hostGut{
		{"Exa.MPLE:9", "exa.mple", 9}, {"[0:0:0:0:0:0:0:1]", "::1", 0}, {"[2001:DB8:0:0::1]:7", "2001:db8::1", 7},
	}
	if len(plan) != 9 {
		t.Fatalf("Gutfälle der Tabelle: %d, erwartet 9", len(plan))
	}
	for _, g := range append(plan, zusatz...) {
		name, port, ursache := policy.ParseHost(g.eintrag)
		if ursache != "" || name != g.name || port != g.port {
			t.Errorf("Gutfall %q: name %q port %d ursache %q, erwartet %q %d", g.eintrag, name, port, ursache, g.name, g.port)
		}
		if u := einzel(t, policy.SchluesselHosts, g.eintrag); u != "" {
			t.Errorf("Gutfall %q: New meldet %q", g.eintrag, u)
		}
	}
	planSchlecht := []eintragUrsache{
		{"", policy.UrsacheLeer},
		{"http://h", policy.UrsacheHost}, {"h/x", policy.UrsacheHost}, {"h/", policy.UrsacheHost}, {"u@h", policy.UrsacheHost},
		{"*.h", policy.UrsacheHost}, {"h*", policy.UrsacheHost}, {" h", policy.UrsacheHost}, {"h x", policy.UrsacheHost},
		{"h:0", policy.UrsacheHost}, {"h:65536", policy.UrsacheHost}, {"h:", policy.UrsacheHost}, {"h:+1", policy.UrsacheHost},
		{"h:01", policy.UrsacheHost}, {"h:1x", policy.UrsacheHost}, {"[::1", policy.UrsacheHost}, {"[]", policy.UrsacheHost},
		{"[1.2.3.4]", policy.UrsacheHost}, {"[fe80::1%25eth0]", policy.UrsacheHost}, {"h\xc3\xa4", policy.UrsacheHost},
	}
	zusatzSchlecht := []eintragUrsache{
		{"[::1]:", policy.UrsacheHost}, {"[::1]x", policy.UrsacheHost}, {"[::1]:0", policy.UrsacheHost},
		{"[fe80::1%eth0]", policy.UrsacheHost}, {":80", policy.UrsacheHost}, {"h:65535:1", policy.UrsacheHost},
		{"h:100000", policy.UrsacheHost}, {"h:18446744073709551617", policy.UrsacheHost},
		{"[::1]80", policy.UrsacheHost}, {"[::1]180", policy.UrsacheHost},
	}
	if len(planSchlecht) != 20 {
		t.Fatalf("Verstöße der Tabelle: %d, erwartet 20", len(planSchlecht))
	}
	for _, s := range append(planSchlecht, zusatzSchlecht...) {
		if _, _, u := policy.ParseHost(s.eintrag); u != s.ursache {
			t.Errorf("Verstoß %q: Ursache %q, erwartet %q", s.eintrag, u, s.ursache)
		}
		if u := einzel(t, policy.SchluesselHosts, s.eintrag); u != s.ursache {
			t.Errorf("Verstoß %q: New meldet %q, erwartet %q", s.eintrag, u, s.ursache)
		}
	}
}

// TestHostGenerator führt zu vier Grundformen jedes Byte an jeder Stelle (einfügen, ersetzen,
// löschen) und vergleicht Ursache, Name und Port mit dem Orakel.
func TestHostGenerator(t *testing.T) {
	basen := []string{"h.example:8080", "[2001:db8::1]:8080", "[::1]", "H-1_x"}
	erwartet, gezaehlt, gut, schlecht := 0, 0, 0, 0
	for _, b := range basen {
		erwartet += zahlVarianten(b)
		for _, e := range varianten(b) {
			gezaehlt++
			name, port, soll := orakelHost(e)
			hName, hPort, ursache := policy.ParseHost(e)
			if ursache != soll {
				t.Errorf("%q: Ursache %q, erwartet %q", e, ursache, soll)
				continue
			}
			if soll != "" {
				schlecht++
				if hName != "" || hPort != 0 {
					t.Errorf("%q: Verstoß mit Wert %q %d", e, hName, hPort)
				}
				continue
			}
			gut++
			if hPort != port {
				t.Errorf("%q: Port %d, erwartet %d", e, hPort, port)
			}
			if e[0] == '[' {
				if ipSoll, ipIst := net.ParseIP(name), net.ParseIP(hName); ipSoll == nil || ipIst == nil || !ipSoll.Equal(ipIst) {
					t.Errorf("%q: Adresse %q, erwartet %q", e, hName, name)
				}
			} else if hName != name {
				t.Errorf("%q: Name %q, erwartet %q", e, hName, name)
			}
		}
	}
	if gezaehlt != erwartet {
		t.Fatalf("Zahl der Elemente %d, erwartet %d", gezaehlt, erwartet)
	}
	if gut < 100 || schlecht < 1000 {
		t.Fatalf("Generator zu schmal: %d gültig, %d ungültig", gut, schlecht)
	}
}
