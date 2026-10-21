package record

import (
	"slices"
	"strings"
)

// asciiKlein setzt allein ASCII-Großbuchstaben herab: Header-Namen vergleicht der Dienst ohne
// Beachtung der ASCII-Schreibweise (SPEC-121, Kopf).
func asciiKlein(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

// werteVon nimmt die Werte aller Schreibweisen eines Namens zusammen, in der Bytefolge der
// Namen, dann in ihrer Reihenfolge (SPEC-121, Kopf); klein ist der Name in Kleinbuchstaben.
func werteVon(h map[string][]string, klein string) []string {
	var namen []string
	for n := range h {
		if asciiKlein(n) == klein {
			namen = append(namen, n)
		}
	}
	slices.Sort(namen)
	var werte []string
	for _, n := range namen {
		werte = append(werte, h[n]...)
	}
	return werte
}

// hopByHop meldet die feste Liste der Hop-by-Hop-Header (SPEC-121 Nr. 4); klein ist der Name in
// Kleinbuchstaben.
func hopByHop(klein string) bool {
	switch klein {
	case "connection", "proxy-connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
		"te", "trailer", "transfer-encoding", "upgrade":
		return true
	}
	return false
}

// teile trennt einen Wert an Kommas und entfernt Leerzeichen und Tabulatoren am Rand jedes Teils.
func teile(w string) []string {
	var out []string
	for _, t := range strings.Split(w, ",") {
		out = append(out, strings.Trim(t, " \t"))
	}
	return out
}

// ohneHopByHop liefert eine Kopie von h ohne die feste Hop-by-Hop-Liste und ohne die Namen, die
// ein Wert von Connection nennt (SPEC-121 Nr. 4, 7); leere Teile haben keine Wirkung.
func ohneHopByHop(h map[string][]string) map[string][]string {
	genannt := map[string]bool{}
	for _, w := range werteVon(h, "connection") {
		for _, t := range teile(w) {
			if t != "" {
				genannt[asciiKlein(t)] = true
			}
		}
	}
	out := make(map[string][]string, len(h))
	for n, werte := range h {
		k := asciiKlein(n)
		if !hopByHop(k) && !genannt[k] {
			out[n] = slices.Clone(werte)
		}
	}
	return out
}

// weiterzuleiten liefert die Header zum Upstream (SPEC-121 Nr. 4): ohne Hop-by-Hop, ohne jeden
// Namen mit dem Präfix X-Recorder-, ohne Expect, Host und Content-Length; Accept-Encoding nur mit
// den Gliedern, die die Pipeline entpackt. Alle übrigen gehen mit allen Werten.
func weiterzuleiten(h map[string][]string) map[string][]string {
	out := ohneHopByHop(h)
	ae, mitAE := acceptEncoding(out)
	for n := range out {
		switch k := asciiKlein(n); {
		case strings.HasPrefix(k, "x-recorder-"), k == "expect", k == "host", k == "content-length", k == "accept-encoding":
			delete(out, n)
		}
	}
	if mitAE {
		out["Accept-Encoding"] = []string{ae}
	}
	return out
}

// acceptEncoding liefert den Wert von Accept-Encoding zum Upstream und ob der Header mit
// mindestens einem Wert steht (SPEC-121 Nr. 4): ein Glied bleibt mit Text und Parametern, wenn
// sein Coding gzip, deflate oder identity ist; bleibt keines, ist der Wert identity.
func acceptEncoding(h map[string][]string) (string, bool) {
	werte := werteVon(h, "accept-encoding")
	if len(werte) == 0 {
		return "", false
	}
	var glieder []string
	for _, w := range werte {
		for _, g := range teile(w) {
			coding, _, _ := strings.Cut(g, ";")
			switch asciiKlein(strings.Trim(coding, " \t")) {
			case "gzip", "deflate", "identity":
				glieder = append(glieder, g)
			}
		}
	}
	if len(glieder) == 0 {
		return "identity", true
	}
	return strings.Join(glieder, ", "), true
}
