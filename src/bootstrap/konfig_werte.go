package bootstrap

import (
	"math"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

// nichtGelieferte sind die Schlüssel und Präfixe, die diese Version nicht liefert (SPEC-110 Nr. 2).
func nichtGelieferte() (praefixe, namen []string) {
	return []string{"ADMIN_", "HYBRID_", "FALLBACK_", "LISTEN_TLS_"}, []string{"STRICT_COVERAGE", "PUBLIC_BASE_URL"}
}

func (p *pruefer) nichtGeliefert() {
	praefixe, namen := nichtGelieferte()
	for k, v := range p.env {
		if strings.TrimSpace(v) == "" {
			continue
		}
		gesperrt := slices.Contains(namen, k)
		for _, pre := range praefixe {
			gesperrt = gesperrt || strings.HasPrefix(k, pre)
		}
		if gesperrt {
			p.melde(k, "wird von dieser Version nicht unterstützt")
		}
	}
}

func (p *pruefer) liesModus(k *Konfiguration) {
	vorher := len(p.verstoesse)
	k.Modus = p.aufzaehlung("MODE", "mock", "record", "replay", "mock", "hybrid")
	if len(p.verstoesse) > vorher {
		return
	}
	if k.Modus == "mock" || k.Modus == "hybrid" {
		p.melde("MODE", "Modus nicht unterstützt; lieferbar sind record und replay")
	}
}

// aufzaehlung liefert den Wert oder den Default; jede andere Schreibweise ist ein Verstoß.
func (p *pruefer) aufzaehlung(schluessel, standard string, erlaubt ...string) string {
	w, ok := p.wert(schluessel)
	switch {
	case !ok:
		return standard
	case !slices.Contains(erlaubt, w):
		p.melde(schluessel, "Wert nicht in der Aufzählung: "+strings.Join(erlaubt, ", "))
		return standard
	}
	return w
}

func (p *pruefer) wahrheitswert(schluessel string, standard bool) bool {
	w, gesetzt, gueltig := p.optionalerWahrheitswert(schluessel)
	if !gesetzt || !gueltig {
		return standard
	}
	return w
}

// optionalerWahrheitswert liest einen Wahrheitswert; gesetzt meldet einen vorhandenen Wert,
// gueltig seine Form. Ein ungültiger Wert ist gemeldet und liefert wert=false.
func (p *pruefer) optionalerWahrheitswert(schluessel string) (wert, gesetzt, gueltig bool) {
	w, ok := p.wert(schluessel)
	switch {
	case !ok:
		return false, false, true
	case w == "true":
		return true, true, true
	case w == "false":
		return false, true, true
	}
	p.melde(schluessel, "Wahrheitswert muss true oder false sein")
	return false, true, false
}

func (p *pruefer) liesEinfache(k *Konfiguration) {
	k.Listen = p.listen("LISTEN", "127.0.0.1:8080")
	k.DataDir = p.datenverzeichnis()
	k.ReplayStrategy = p.aufzaehlung("REPLAY_STRATEGY", "sequential", "sequential", "strict", "first", "last")
	k.ReplayAssignment = p.aufzaehlung("REPLAY_ASSIGNMENT", "explicit", "explicit", "ordered")
	k.ReplayRedactedBodies = p.wahrheitswert("REPLAY_REDACTED_BODIES", false)
	k.MatchInvalidEncoding = p.aufzaehlung("MATCH_INVALID_ENCODING", "literal", "literal", "reject")
	k.MatchBodyMode = p.aufzaehlung("MATCH_BODY_MODE", "bytes", "bytes", "json-canonical")
	k.LinkRewrite = p.aufzaehlung("LINK_REWRITE", "off", "off")
	k.MaxCaptureBody = p.groesse("MAX_CAPTURE_BODY", 8<<20)
	k.UpstreamTimeout = p.dauer("UPSTREAM_TIMEOUT", 30*time.Second, 0, false)
	k.ShutdownTimeout = p.dauer("SHUTDOWN_TIMEOUT", 10*time.Second, 0, false)
	k.SQLiteBusyTimeout = p.dauer("SQLITE_BUSY_TIMEOUT", 5000*time.Millisecond, 2147483647*time.Millisecond, false)
	k.RetentionMaxAge = p.dauer("RETENTION_MAX_AGE", 0, 0, true)
	if w, ok := p.wert("UPSTREAM_CA_FILE"); ok {
		k.UpstreamCAFile = w
	}
}

func (p *pruefer) datenverzeichnis() string {
	w, ok := p.wert("DATA_DIR")
	switch {
	case !ok:
		return "data"
	case strings.ContainsRune(w, 0):
		p.melde("DATA_DIR", "Pfad enthält ein NUL-Zeichen")
		return "data"
	}
	return w
}

// dauer liest eine Dauer nach SPEC-110 Nr. 4. hoechstens 0 heißt: nur die Grenze von 63 Bit.
// nullErlaubt lässt 0 mit oder ohne Einheit zu.
func (p *pruefer) dauer(schluessel string, standard, hoechstens time.Duration, nullErlaubt bool) time.Duration {
	w, ok := p.wert(schluessel)
	if !ok {
		return standard
	}
	d, gueltig := parseDauer(w, nullErlaubt)
	if !gueltig || (hoechstens > 0 && d > hoechstens) {
		p.melde(schluessel, "Dauer ungültig: Zahl mit Einheit ms, s, m, h oder d, im zulässigen Bereich")
		return standard
	}
	return d
}

func parseDauer(w string, nullErlaubt bool) (time.Duration, bool) {
	if nullErlaubt && w == "0" {
		return 0, true
	}
	ziffern := laengeZiffern(w)
	if ziffern == 0 {
		return 0, false
	}
	n, err := strconv.ParseUint(w[:ziffern], 10, 63)
	if err != nil {
		return 0, false
	}
	einheit := map[string]time.Duration{"ms": time.Millisecond, "s": time.Second, "m": time.Minute, "h": time.Hour, "d": 24 * time.Hour}[w[ziffern:]]
	if einheit == 0 || n > uint64(math.MaxInt64/einheit) {
		return 0, false
	}
	d := time.Duration(n) * einheit
	return d, d > 0 || nullErlaubt
}

func laengeZiffern(w string) int {
	n := 0
	for n < len(w) && w[n] >= '0' && w[n] <= '9' {
		n++
	}
	return n
}

// groesse liest eine Größe nach SPEC-110 Nr. 5, 1 Byte bis 1 GiB.
func (p *pruefer) groesse(schluessel string, standard int64) int64 {
	w, ok := p.wert(schluessel)
	if !ok {
		return standard
	}
	n, gueltig := parseGroesse(w)
	if !gueltig {
		p.melde(schluessel, "Größe ungültig: Zahl, optional mit B, KiB, MiB oder GiB, 1 Byte bis 1 GiB")
		return standard
	}
	return n
}

func parseGroesse(w string) (int64, bool) {
	ziffern := laengeZiffern(w)
	if ziffern == 0 {
		return 0, false
	}
	n, err := strconv.ParseUint(w[:ziffern], 10, 63)
	if err != nil {
		return 0, false
	}
	rest := strings.TrimPrefix(w[ziffern:], " ")
	if rest != w[ziffern:] && rest == "" {
		return 0, false
	}
	faktor, ok := map[string]uint64{"": 1, "B": 1, "KiB": 1 << 10, "MiB": 1 << 20, "GiB": 1 << 30}[rest]
	const grenze = 1 << 30
	if !ok || n == 0 || n > grenze/faktor {
		return 0, false
	}
	return int64(n * faktor), true
}

// liste liest eine Liste nach SPEC-110 Nr. 6; gesetzt ist false, wenn der Wert leer ist.
func (p *pruefer) liste(schluessel string) (eintraege []string, gesetzt bool) {
	w, ok := p.wert(schluessel)
	if !ok {
		return nil, false
	}
	for _, e := range strings.Split(w, ",") {
		e = strings.TrimSpace(e)
		if e == "" {
			p.melde(schluessel, "Liste enthält einen leeren Eintrag")
			return nil, true
		}
		eintraege = append(eintraege, e)
	}
	return p.pruefeNone(schluessel, eintraege), true
}

// pruefeNone lässt `none` nur als einzigen, klein geschriebenen Eintrag von MATCH_VARY_HEADERS und
// VOLATILE_HEADERS zu; jeder Eintrag, der ohne Beachtung der Schreibweise `none` lautet, ist sonst
// ein Verstoß (SPEC-110 Nr. 1). Bei einem Verstoß liefert sie keine Einträge.
func (p *pruefer) pruefeNone(schluessel string, eintraege []string) []string {
	if !slices.ContainsFunc(eintraege, func(e string) bool { return strings.EqualFold(e, "none") }) {
		return eintraege
	}
	if len(eintraege) == 1 && eintraege[0] == "none" && slices.Contains(ersetzendeListen(), schluessel) {
		return eintraege
	}
	p.melde(schluessel, "none ist nur als einziger, klein geschriebener Eintrag von MATCH_VARY_HEADERS und VOLATILE_HEADERS gültig")
	return nil
}

// ersetzendeListen sind die Listen, deren gesetzter Wert den Default ersetzt (SPEC-110 Nr. 1).
func ersetzendeListen() []string { return []string{"MATCH_VARY_HEADERS", "VOLATILE_HEADERS"} }

func menge(eintraege []string) []string {
	out := slices.Clone(eintraege)
	slices.Sort(out)
	return slices.Compact(out)
}

func headerMenge(eintraege []string) []string {
	out := make([]string, len(eintraege))
	for i, e := range eintraege {
		out[i] = http.CanonicalHeaderKey(e)
	}
	return menge(out)
}

func (p *pruefer) liesMengen(k *Konfiguration) {
	k.MatchVaryHeaders = p.ersetzendeHeaderListe("MATCH_VARY_HEADERS", []string{"Accept", "OData-Version", "Content-Type", "If-Match", "Prefer"})
	k.VolatileHeaders = p.ersetzendeHeaderListe("VOLATILE_HEADERS", []string{"Date", "Server", "Via", "Age", "X-Request-Id", "X-Correlation-Id", "Traceparent"})
	k.RedactHeaders = headerMenge(append(p.ergaenzendeListe("REDACT_HEADERS", istHeaderToken),
		"Authorization", "Proxy-Authorization", "Cookie", "Set-Cookie", "X-API-Key", "X-Auth-Token", "X-CSRF-Token"))
	k.RedactQueryParams = menge(append(p.ergaenzendeListe("REDACT_QUERY_PARAMS", istQueryName),
		"access_token", "token", "api_key", "apikey", "key", "sig", "signature", "code", "client_secret"))
	k.RedactJSONPaths = menge(append(p.ergaenzendeListe("REDACT_JSON_PATHS", istJSONPfad),
		"$..password", "$..passwd", "$..secret", "$..client_secret", "$..access_token", "$..refresh_token", "$..api_key"))
}

// ersetzendeHeaderListe: ein gesetzter Wert ersetzt den Default, `none` allein ist die leere Menge.
func (p *pruefer) ersetzendeHeaderListe(schluessel string, standard []string) []string {
	e, gesetzt := p.liste(schluessel)
	switch {
	case !gesetzt:
		return headerMenge(standard)
	case len(e) == 1 && e[0] == "none":
		return []string{}
	}
	p.pruefeEintraege(schluessel, e, istHeaderToken, "Header-Name ist kein HTTP-Token")
	return headerMenge(e)
}

// ergaenzendeListe liefert die gesetzten Einträge, die den Default ergänzen.
func (p *pruefer) ergaenzendeListe(schluessel string, gueltig func(string) bool) []string {
	e, gesetzt := p.liste(schluessel)
	if !gesetzt {
		return nil
	}
	p.pruefeEintraege(schluessel, e, gueltig, "Eintrag hat eine ungültige Form")
	return e
}

func (p *pruefer) pruefeEintraege(schluessel string, e []string, gueltig func(string) bool, grund string) {
	for _, x := range e {
		if !gueltig(x) {
			p.melde(schluessel, grund)
			return
		}
	}
}

func istHeaderToken(s string) bool {
	return s != "" && zeichenAus(s, "!#$%&'*+-.^_`|~")
}

func istQueryName(s string) bool {
	return s != "" && !strings.ContainsAny(s, " \t\r\n=&#,")
}

// istJSONPfad erkennt allein die Form `$..name` (SPEC-110 Nr. 6).
func istJSONPfad(s string) bool {
	name, ok := strings.CutPrefix(s, "$..")
	return ok && name != "" && zeichenAus(name, "_-")
}

// zeichenAus meldet, ob s nur aus Buchstaben, Ziffern und den Zeichen extra besteht.
func zeichenAus(s, extra string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !buchstabeOderZiffer(c) && strings.IndexByte(extra, c) < 0 {
			return false
		}
	}
	return true
}

func buchstabeOderZiffer(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// listen prüft `host:port` (SPEC-110 Nr. 7).
func (p *pruefer) listen(schluessel, standard string) string {
	w, ok := p.wert(schluessel)
	if !ok {
		return standard
	}
	if !istListenAdresse(w) {
		p.melde(schluessel, "Adresse muss host:port sein, Port 1 bis 65535")
		return standard
	}
	return w
}

func istListenAdresse(w string) bool {
	host, port, err := net.SplitHostPort(w)
	if err != nil || host == "" || laengeZiffern(port) != len(port) || port == "" {
		return false
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return false
	}
	if strings.HasPrefix(w, "[") {
		return net.ParseIP(host) != nil
	}
	return zeichenAus(host, ".-_")
}

func istDienstID(s string) bool {
	return len(s) >= 1 && len(s) <= 64 && zeichenAus(s, "._-")
}

// istRoot prüft `/` oder Segmente `/a/b` ohne abschließenden Schrägstrich (SPEC-110 Nr. 7).
func istRoot(s string) bool {
	if s == "/" {
		return true
	}
	rest, ok := strings.CutPrefix(s, "/")
	if !ok {
		return false
	}
	for _, seg := range strings.Split(rest, "/") {
		if seg == "" || seg == "." || seg == ".." || !zeichenAus(seg, ".~_-") {
			return false
		}
	}
	return true
}

// normalisiereURL prüft eine Upstream-URL (SPEC-110 Nr. 7) und entfernt einen abschließenden `/`.
func normalisiereURL(s string) (string, bool) {
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Opaque != "" || u.Hostname() == "" ||
		u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(s, "#") || !portGueltig(u) {
		return "", false
	}
	return strings.TrimSuffix(s, "/"), true
}

// portGueltig verlangt für einen genannten Port 1 bis 65535; `host:` ohne Ziffern ist ungültig.
func portGueltig(u *url.URL) bool {
	if strings.HasSuffix(u.Host, ":") {
		return false
	}
	port := u.Port()
	if port == "" {
		return true
	}
	n, err := strconv.Atoi(port)
	return err == nil && n >= 1 && n <= 65535
}
