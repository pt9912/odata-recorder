package policy_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// portTexte erzeugt die Ports von TestZielPort: 2^k und 2^k - 1 für acht Stellen k (16, 32, 48, 63,
// 64, 65, 96, 128), je mit den Zuschlägen 0, 80, 443 und 8443 (so liegt 2^64 + 443 hinter der
// Wrap-Grenze eines 64-Bit-Zählers auf dem Port 443), dazu Randformen: gültige Grenzen, führende
// Nullen, sehr lange Ziffernfolgen.
func portTexte() []string {
	var aus []string
	for _, k := range []uint{16, 32, 48, 63, 64, 65, 96, 128} {
		zwei := new(big.Int).Lsh(big.NewInt(1), k)
		for _, grund := range []*big.Int{zwei, new(big.Int).Sub(zwei, big.NewInt(1))} {
			for _, zu := range []int64{0, 80, 443, 8443} {
				aus = append(aus, new(big.Int).Add(grund, big.NewInt(zu)).String())
			}
		}
	}
	return append(aus,
		"1", "65535", "65536", "99999", "0", "00000", "0443", "00443", "080",
		strings.Repeat("0", 1000)+"443", strings.Repeat("0", 1000)+"65536", strings.Repeat("0", 40)+"65535",
		strings.Repeat("1", 1000), strings.Repeat("9", 30),
		strings.Repeat("0", 40)+"18446744073709552059", strings.Repeat("0", 1000)+"18446744073709552059",
	)
}

// TestZielPort (P12, Nr. 5 c): ein Port des Ziels gilt genau dann, wenn er eine Dezimalzahl von 1
// bis 65535 ist, auch mit führenden Nullen und gleich lang wie jede Ziffernfolge; sonst gibt das
// Ziel nichts frei, auch wenn der Wert eines 64-Bit-Zählers nach einem Überlauf ein Port wäre. Die
// Menge ist Port x Schema {http, https} x Eintrag {h, h:80, h:443, h:8443, h:65535, h:1}; das
// Orakel liest den Port mit math/big.
func TestZielPort(t *testing.T) {
	texte := portTexte()
	eintraege := []string{"h", "h:80", "h:443", "h:8443", "h:65535", "h:1"}
	faelle, frei := 0, 0
	for _, text := range texte {
		wert, ok := new(big.Int).SetString(text, 10)
		if !ok {
			t.Fatalf("Port %q", text)
		}
		gueltig := wert.Sign() > 0 && wert.Cmp(big.NewInt(65535)) <= 0
		for _, schema := range []string{"http", "https"} {
			standard := int64(80)
			if schema == "https" {
				standard = 443
			}
			for _, e := range eintraege {
				faelle++
				eport := standard
				if i := strings.IndexByte(e, ':'); i >= 0 {
					eport = map[string]int64{"80": 80, "443": 443, "8443": 8443, "65535": 65535, "1": 1}[e[i+1:]]
				}
				want := policy.Bleiben
				if gueltig && wert.Cmp(big.NewInt(eport)) == 0 {
					want = policy.Upstream
					frei++
				}
				ziel := schema + "://h:" + text
				if got := ent(hostPolicy(t, e), "GET", "/", ziel, e6001); got != want {
					t.Errorf("Ziel mit Port %.40q (%d Zeichen), Eintrag %q: %s, erwartet %s", ziel, len(text), e, name(got), name(want))
				}
			}
		}
	}
	if len(texte) != 8*2*4+16 || faelle != len(texte)*2*6 || frei == 0 {
		t.Fatalf("%d Ports, %d Fälle, %d frei", len(texte), faelle, frei)
	}
}

// TestZielName (P12, Nr. 5 c): zielName liest Host mit Klammern (roh) und Hostnamen ohne sie
// unabhängig von url.Parse. Eine IPv6-Adresse steht nur in Klammern und ein Name nie; jede der
// beiden Richtungen hat Fälle.
func TestZielName(t *testing.T) {
	nicht := []struct{ roh, host string }{
		{"[1.2.3.4]", "1.2.3.4"}, {"[h]", "h"}, {"[]", ""}, {"::1", "::1"}, {"h:80:80", "h:80"},
		{"[1.2.3.4]:80", "1.2.3.4"}, {"::ffff:1.2.3.4", "::ffff:1.2.3.4"},
		{"", ""}, {"[::1%eth0]", "::1%eth0"}, {"[1::2::3]", "1::2::3"}, {"h", "h" + string([]byte{0x80})},
	}
	gut := []struct{ roh, host, name string }{
		{"[::1]", "::1", "::1"}, {"h", "h", "h"}, {"H", "H", "h"}, {"[::FFFF:1.2.3.4]", "::FFFF:1.2.3.4", "::ffff:1.2.3.4"},
		{"[0:0:0:0:0:0:0:1]:8443", "0:0:0:0:0:0:0:1", "::1"}, {"h:443", "h", "h"}, {"1.2.3.4", "1.2.3.4", "1.2.3.4"},
	}
	for _, f := range nicht {
		if n, ok := policy.ZielName(f.roh, f.host); ok || n != "" {
			t.Errorf("(%q, %q): %q, %v; erwartet nicht lesbar", f.roh, f.host, n, ok)
		}
	}
	for _, f := range gut {
		if n, ok := policy.ZielName(f.roh, f.host); !ok || n != f.name {
			t.Errorf("(%q, %q): %q, %v; erwartet %q", f.roh, f.host, n, ok, f.name)
		}
	}
	if len(nicht) != 11 || len(gut) != 7 {
		t.Fatalf("%d und %d Fälle", len(nicht), len(gut))
	}
}
