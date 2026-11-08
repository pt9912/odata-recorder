package bootstrap

import (
	"strconv"
	"strings"
)

// maxWurzel ist die größte Wurzel in Bytes, die der Schreiber der Version 4 annimmt (SPEC-110 Nr. 14 b).
const maxWurzel = 2048

// liesPublicBaseURL liest PUBLIC_BASE_URL (SPEC-110 Nr. 14): die Form in jedem Modus, die Pflicht, die Ableitung
// aus LISTEN und die Länge der Wurzel nur im Modus mock; ein Schlüssel meldet höchstens einen Grund.
func (p *pruefer) liesPublicBaseURL(k *Konfiguration) {
	w, gesetzt := p.wert("PUBLIC_BASE_URL")
	switch {
	case gesetzt && !istBasisURL(w):
		p.melde("PUBLIC_BASE_URL", "Form ungültig: http oder https, Host, optional Port, Segmente ohne abschließenden Schrägstrich")
		return
	case !gesetzt && k.Modus != modusMock:
		return
	case !gesetzt:
		abgeleitet, ok := basisAusListen(k.Listen)
		if !ok {
			p.melde("PUBLIC_BASE_URL", "im Modus mock bei diesem Host erforderlich")
			return
		}
		w = abgeleitet
	}
	k.PublicBaseURL = w
	if k.Modus == modusMock && laengsteWurzel(w, k.Dienste) > maxWurzel {
		p.melde("PUBLIC_BASE_URL", "mit dem Root eines Service länger als 2048 Bytes")
	}
}

// laengsteWurzel ist die Länge der längsten Wurzel eines Service: Basis plus Root, der Root `/` fügt nichts hinzu.
func laengsteWurzel(basis string, dienste []Dienst) int {
	n := len(basis)
	for _, d := range dienste {
		if d.Root != "/" {
			n = max(n, len(basis)+len(d.Root))
		}
	}
	return n
}

// basisAusListen bildet `http://<Host>:<Port>` für einen Loopback-Host (SPEC-110 Nr. 14 c); die Schreibweise des Hosts
// zählt, der Port steht als Dezimalzahl ohne führende Nullen.
func basisAusListen(listen string) (string, bool) {
	i := strings.LastIndexByte(listen, ':')
	if i < 0 {
		return "", false
	}
	host, port := listen[:i], listen[i+1:]
	if host != "127.0.0.1" && host != "[::1]" && host != "localhost" {
		return "", false
	}
	n, err := strconv.Atoi(port)
	if err != nil {
		return "", false
	}
	return "http://" + host + ":" + strconv.Itoa(n), true
}

// istBasisURL prüft die Form der Wurzel, die der Schreiber der Version 4 annimmt (SPEC-110 Nr. 14 b): Schema, Host,
// Port mit einer bis fünf Ziffern, Segmente.
func istBasisURL(s string) bool {
	if len(s) > maxWurzel {
		return false
	}
	rest, ok := strings.CutPrefix(s, "http://")
	if !ok {
		if rest, ok = strings.CutPrefix(s, "https://"); !ok {
			return false
		}
	}
	rest, ok = hostEnde(rest)
	if !ok {
		return false
	}
	if port, hat := strings.CutPrefix(rest, ":"); hat {
		n := laengeZiffern(port)
		if n == 0 || n > 5 {
			return false
		}
		rest = port[n:]
	}
	if rest == "" {
		return true
	}
	segmente, ok := strings.CutPrefix(rest, "/")
	if !ok {
		return false
	}
	for _, seg := range strings.Split(segmente, "/") {
		if !segmentGueltig(seg) {
			return false
		}
	}
	return true
}

func hexZiffer(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

// hostZeichen sind die Zeichen eines Hosts und eines Segments ohne Prozent-Kodierung.
func hostZeichen(c byte) bool { return buchstabeOderZiffer(c) || strings.IndexByte("-._~", c) >= 0 }

// hostEnde liest den Host am Anfang von s und liefert den Rest (leer, `:port` oder `/…`).
func hostEnde(s string) (string, bool) {
	if strings.HasPrefix(s, "[") {
		i := strings.IndexByte(s, ']')
		if i < 2 {
			return "", false
		}
		for j := 1; j < i; j++ {
			if !hexZiffer(s[j]) && s[j] != ':' && s[j] != '.' {
				return "", false
			}
		}
		return s[i+1:], true
	}
	i := 0
	for i < len(s) && hostZeichen(s[i]) {
		i++
	}
	return s[i:], i > 0
}

// segmentGueltig: nicht leer, Zeichen des Hosts oder `%` mit zwei Hex-Ziffern, dekodiert weder `.` noch `..`.
func segmentGueltig(seg string) bool {
	if seg == "" {
		return false
	}
	var dek strings.Builder
	for i := 0; i < len(seg); i++ {
		c := seg[i]
		switch {
		case hostZeichen(c):
			dek.WriteByte(c)
		case c == '%' && i+2 < len(seg) && hexZiffer(seg[i+1]) && hexZiffer(seg[i+2]):
			v, _ := strconv.ParseUint(seg[i+1:i+3], 16, 8)
			dek.WriteByte(byte(v))
			i += 2
		default:
			return false
		}
	}
	d := dek.String()
	return d != "." && d != ".."
}
